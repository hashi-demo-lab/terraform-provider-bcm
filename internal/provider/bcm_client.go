// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Retry configuration constants
const (
	// DefaultMaxRetries is the default number of retries for transient errors
	DefaultMaxRetries = 3
	// DefaultBaseDelay is the initial delay between retries
	DefaultBaseDelay = 1 * time.Second
	// DefaultMaxDelay is the maximum delay between retries
	DefaultMaxDelay = 30 * time.Second
	// DefaultJitterFactor adds randomness to prevent thundering herd (0.0-1.0)
	DefaultJitterFactor = 0.2
)

// BCMClient handles JSON-RPC API calls to BCM with cookie-based authentication.
type BCMClient struct {
	HTTPClient   *http.Client  // Includes cookie jar for automatic cm-login-token management
	Endpoint     string        // Base URL (e.g., https://172.21.15.254:8081)
	MaxRetries   int           // Maximum number of retries for transient errors
	BaseDelay    time.Duration // Initial delay between retries
	MaxDelay     time.Duration // Maximum delay between retries
	JitterFactor float64       // Randomness factor to prevent thundering herd (0.0-1.0)
}

// ValidationError represents a structured validation error from BCM API.
// BCM validation responses return an array of validation objects with these fields.
type ValidationError struct {
	Field      string // Attribute that failed validation (e.g., "SOLSpeed", "hostname", "name")
	Message    string // Human-readable error description from BCM
	ErrorCode  string // BCM error code (e.g., "BAD_VALUE", "DUPLICATE_FIELD", "NOT_NULL")
	Severity   string // Error severity level - either "ERROR" or "WARNING"
	EntityUUID string // Reference to related entity if applicable (may be empty)
}

// IsError returns true if the validation error has ERROR severity or unknown severity.
// Unknown severity is treated as ERROR for safety (halt operation).
func (v ValidationError) IsError() bool {
	return v.Severity == "ERROR" || (v.Severity != "WARNING" && v.Severity != "")
}

// IsWarning returns true if the validation error has WARNING severity.
func (v ValidationError) IsWarning() bool {
	return v.Severity == "WARNING"
}

// JSONRPCRequest represents BCM JSON-RPC request body
// ENHANCEMENT: Args parameter now supported for parameterized API calls
// This enables efficient direct lookups like getSoftwareImage(name) instead of
// client-side filtering of getSoftwareImages() results.
type JSONRPCRequest struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"` // Optional arguments array
}

// JSONRPCRequestArg represents BCM JSON-RPC request with single "arg" field.
// Some BCM methods (like reboot) expect "arg" (single value) instead of "args" (array).
type JSONRPCRequestArg struct {
	Service string      `json:"service"`
	Call    string      `json:"call"`
	Arg     interface{} `json:"arg,omitempty"` // Optional single argument
}

// LoginRequest represents login API call request body.
type LoginRequest struct {
	Service  string `json:"service"` // Always "login"
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewBCMClient creates authenticated client with cookie jar and performs login.
func NewBCMClient(ctx context.Context, endpoint, username, password string, insecureSkipVerify bool, timeout int) (*BCMClient, error) {
	// Create cookie jar for automatic cookie management
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create cookie jar: %w", err)
	}

	// Create HTTP client with TLS config and cookie jar
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: insecureSkipVerify,
		},
	}

	client := &http.Client{
		Jar:       jar,
		Transport: transport,
		Timeout:   time.Duration(timeout) * time.Second,
	}

	// Perform login to obtain authentication token
	loginReq := LoginRequest{
		Service:  "login",
		Username: username,
		Password: password,
	}

	jsonBody, err := json.Marshal(loginReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal login request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", endpoint+"/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create login request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	tflog.Debug(ctx, "Attempting BCM login", map[string]interface{}{
		"endpoint": endpoint + "/json",
		"username": username,
	})

	// Retry login with exponential backoff for transient errors
	var resp *http.Response
	var body []byte
	var lastErr error
	var loginCookie *http.Cookie

	for attempt := 0; attempt <= DefaultMaxRetries; attempt++ {
		// Recreate request body for each attempt
		if attempt > 0 {
			req.Body = io.NopCloser(bytes.NewReader(jsonBody))
			// Exponential backoff
			backoff := DefaultBaseDelay * time.Duration(1<<uint(attempt-1))
			if backoff > DefaultMaxDelay {
				backoff = DefaultMaxDelay
			}
			// Add jitter
			jitter := time.Duration(float64(backoff) * DefaultJitterFactor * rand.Float64())
			backoff += jitter

			tflog.Warn(ctx, "BCM login failed, retrying", map[string]interface{}{
				"attempt":     attempt + 1,
				"max_retries": DefaultMaxRetries,
				"error":       lastErr.Error(),
				"backoff_ms":  backoff.Milliseconds(),
			})

			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, fmt.Errorf("login cancelled: %w", ctx.Err())
			}
		}

		resp, lastErr = client.Do(req)
		if lastErr != nil {
			if isRetryableError(lastErr) && attempt < DefaultMaxRetries {
				continue
			}
			return nil, fmt.Errorf("login API call failed after %d attempts: %w", attempt+1, lastErr)
		}

		body, lastErr = io.ReadAll(resp.Body)
		resp.Body.Close()
		if lastErr != nil {
			if isRetryableError(lastErr) && attempt < DefaultMaxRetries {
				continue
			}
			return nil, fmt.Errorf("failed to read login response after %d attempts: %w", attempt+1, lastErr)
		}

		// Verify HTTP status - retry on 5xx errors
		if resp.StatusCode >= 500 && attempt < DefaultMaxRetries {
			lastErr = fmt.Errorf("login failed with HTTP %d: %s", resp.StatusCode, string(body))
			continue
		}
		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("login failed with HTTP %d: %s", resp.StatusCode, string(body))
		}

		// Verify response is boolean true
		var loginSuccess bool
		if err := json.Unmarshal(body, &loginSuccess); err != nil || !loginSuccess {
			return nil, fmt.Errorf("login failed: expected boolean true, got: %s", string(body))
		}

		// Verify Set-Cookie header contains cm-login-token
		cookies := resp.Cookies()
		hasLoginToken := false
		for _, cookie := range cookies {
			if cookie.Name == "cm-login-token" {
				hasLoginToken = true
				loginCookie = cookie
				break
			}
		}
		if !hasLoginToken {
			return nil, fmt.Errorf("login response missing cm-login-token cookie")
		}

		// Login successful
		break
	}

	// Log warning if security attributes are missing (production guidance)
	if loginCookie != nil {
		if !loginCookie.Secure {
			tflog.Warn(ctx, "cm-login-token cookie missing Secure attribute - ensure HTTPS is used")
		}
		if !loginCookie.HttpOnly {
			tflog.Warn(ctx, "cm-login-token cookie missing HttpOnly attribute - potential security risk")
		}
	}

	tflog.Debug(ctx, "BCM login successful", map[string]interface{}{
		"endpoint": endpoint,
	})

	// Return authenticated client (cookie jar now contains cm-login-token)
	return &BCMClient{
		HTTPClient:   client,
		Endpoint:     endpoint,
		MaxRetries:   DefaultMaxRetries,
		BaseDelay:    DefaultBaseDelay,
		MaxDelay:     DefaultMaxDelay,
		JitterFactor: DefaultJitterFactor,
	}, nil
}

// isRetryableError determines if an error is transient and should be retried.
// This includes connection refused, EOF, connection reset, timeout, and temporary network errors.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := err.Error()

	// Check for common transient error patterns
	if strings.Contains(errStr, "connection refused") ||
		strings.Contains(errStr, "connection reset") ||
		strings.Contains(errStr, "EOF") ||
		strings.Contains(errStr, "i/o timeout") ||
		strings.Contains(errStr, "no such host") ||
		strings.Contains(errStr, "connection timed out") ||
		strings.Contains(errStr, "TLS handshake timeout") ||
		strings.Contains(errStr, "server closed idle connection") {
		return true
	}

	// Check for net.Error interface (includes timeout and temporary flags)
	var netErr net.Error
	if errors.As(err, &netErr) {
		if netErr.Timeout() {
			return true
		}
	}

	// Check for specific network operation errors
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		return true
	}

	return false
}

// calculateBackoff calculates the delay for the next retry attempt with exponential backoff and jitter.
// delay = min(maxDelay, baseDelay * 2^attempt) + random jitter
func (c *BCMClient) calculateBackoff(attempt int) time.Duration {
	// Exponential backoff: baseDelay * 2^attempt
	delay := c.BaseDelay * time.Duration(1<<uint(attempt))

	// Cap at max delay
	if delay > c.MaxDelay {
		delay = c.MaxDelay
	}

	// Add jitter to prevent thundering herd
	if c.JitterFactor > 0 {
		jitter := time.Duration(float64(delay) * c.JitterFactor * rand.Float64())
		delay += jitter
	}

	return delay
}

// doHTTPRequest performs an HTTP request with retry logic for transient errors.
// Returns the response body and any error encountered after all retries.
func (c *BCMClient) doHTTPRequest(ctx context.Context, req *http.Request, jsonBody []byte, logPrefix string) ([]byte, error) {
	var lastErr error

	for attempt := 0; attempt <= c.MaxRetries; attempt++ {
		// Recreate request body for each attempt (body is consumed after reading)
		if attempt > 0 {
			req.Body = io.NopCloser(bytes.NewReader(jsonBody))
		}

		resp, err := c.HTTPClient.Do(req)
		if err != nil {
			lastErr = err

			// Check if error is retryable
			if isRetryableError(err) && attempt < c.MaxRetries {
				backoff := c.calculateBackoff(attempt)
				tflog.Warn(ctx, fmt.Sprintf("%s request failed, retrying", logPrefix), map[string]interface{}{
					"attempt":     attempt + 1,
					"max_retries": c.MaxRetries,
					"error":       err.Error(),
					"backoff_ms":  backoff.Milliseconds(),
				})

				// Wait before retry (with context cancellation support)
				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return nil, fmt.Errorf("%s call cancelled: %w", logPrefix, ctx.Err())
				}
			}

			return nil, fmt.Errorf("%s call failed after %d attempts: %w", logPrefix, attempt+1, err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			lastErr = err

			// Reading response body failed - may be retryable
			if isRetryableError(err) && attempt < c.MaxRetries {
				backoff := c.calculateBackoff(attempt)
				tflog.Warn(ctx, fmt.Sprintf("%s response read failed, retrying", logPrefix), map[string]interface{}{
					"attempt":     attempt + 1,
					"max_retries": c.MaxRetries,
					"error":       err.Error(),
					"backoff_ms":  backoff.Milliseconds(),
				})

				select {
				case <-time.After(backoff):
					continue
				case <-ctx.Done():
					return nil, fmt.Errorf("%s call cancelled: %w", logPrefix, ctx.Err())
				}
			}

			return nil, fmt.Errorf("failed to read response body after %d attempts: %w", attempt+1, err)
		}

		tflog.Trace(ctx, fmt.Sprintf("%s response", logPrefix), map[string]interface{}{
			"status": resp.StatusCode,
			"body":   string(body),
		})

		// Check for HTTP 5xx errors which may be retryable
		if resp.StatusCode >= 500 && attempt < c.MaxRetries {
			backoff := c.calculateBackoff(attempt)
			tflog.Warn(ctx, fmt.Sprintf("%s server error, retrying", logPrefix), map[string]interface{}{
				"attempt":     attempt + 1,
				"max_retries": c.MaxRetries,
				"status":      resp.StatusCode,
				"backoff_ms":  backoff.Milliseconds(),
			})

			select {
			case <-time.After(backoff):
				continue
			case <-ctx.Done():
				return nil, fmt.Errorf("%s call cancelled: %w", logPrefix, ctx.Err())
			}
		}

		// Defensive error parsing (see section 3 of research.md)
		if err := parseErrorResponse(resp.StatusCode, body); err != nil {
			return nil, err
		}

		return body, nil
	}

	return nil, fmt.Errorf("%s call failed after %d attempts: %w", logPrefix, c.MaxRetries+1, lastErr)
}

// CallJSONRPC executes JSON-RPC call using authenticated cookie jar
// Supports optional variadic args parameter for parameterized API calls
// Includes automatic retry with exponential backoff for transient errors
// Examples:
//
//	CallJSONRPC(ctx, "CMPart", "getSoftwareImages") // No args
//	CallJSONRPC(ctx, "CMPart", "getSoftwareImage", "image-name") // Single arg
//	CallJSONRPC(ctx, "CMPart", "addSoftwareImage", entity, false) // Multiple args
func (c *BCMClient) CallJSONRPC(ctx context.Context, service, call string, args ...interface{}) ([]byte, error) {
	reqBody := JSONRPCRequest{
		Service: service,
		Call:    call,
	}

	// Only include args field if arguments provided (backward compatibility)
	if len(args) > 0 {
		reqBody.Args = args
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONRPC request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint+"/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Cookie header with cm-login-token automatically added by cookie jar

	logFields := map[string]interface{}{
		"service":  service,
		"call":     call,
		"endpoint": c.Endpoint + "/json",
	}
	if len(args) > 0 {
		logFields["args_count"] = len(args)
	}

	tflog.Trace(ctx, "JSONRPC request", logFields)

	// Use retry-enabled HTTP request
	logPrefix := fmt.Sprintf("JSONRPC %s.%s", service, call)
	return c.doHTTPRequest(ctx, req, jsonBody, logPrefix)
}

// CallJSONRPCArg executes JSON-RPC call with single "arg" parameter
// Some BCM methods (like reboot) expect "arg" (single value) instead of "args" (array).
// Use this method for power operations and other methods requiring single argument format.
//
// Examples:
//
//	CallJSONRPCArg(ctx, "cmdevice", "reboot", "node001") // Reboot specific node
//	CallJSONRPCArg(ctx, "cmdevice", "reboot", hostname)  // Reboot by hostname
func (c *BCMClient) CallJSONRPCArg(ctx context.Context, service, call string, arg interface{}) ([]byte, error) {
	reqBody := JSONRPCRequestArg{
		Service: service,
		Call:    call,
	}

	// Only include arg field if argument provided
	if arg != nil {
		reqBody.Arg = arg
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSONRPC request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", c.Endpoint+"/json", bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	// Cookie header with cm-login-token automatically added by cookie jar

	logFields := map[string]interface{}{
		"service":  service,
		"call":     call,
		"endpoint": c.Endpoint + "/json",
	}
	if arg != nil {
		logFields["arg"] = arg
	}

	tflog.Trace(ctx, "JSONRPC request (arg format)", logFields)

	// Use retry-enabled HTTP request
	logPrefix := fmt.Sprintf("JSONRPC %s.%s (arg)", service, call)
	return c.doHTTPRequest(ctx, req, jsonBody, logPrefix)
}

// parseErrorResponse performs multi-layer error detection
// Layer 1: HTTP status code
// Layer 2: JSON object with error field
// Layer 3: Empty array success
// Layer 4: Parse errors.
func parseErrorResponse(statusCode int, body []byte) error {
	// Layer 1: HTTP status
	if statusCode < 200 || statusCode >= 300 {
		return fmt.Errorf("HTTP %d: %s", statusCode, limitString(string(body), 500))
	}

	// Try to unmarshal as JSON
	var jsonData interface{}
	if err := json.Unmarshal(body, &jsonData); err != nil {
		// Layer 4: Parse error
		return fmt.Errorf("failed to parse JSON response: %w, body: %s", err, limitString(string(body), 500))
	}

	// Layer 2: JSON object with error field OR validation array
	if objMap, ok := jsonData.(map[string]interface{}); ok {
		// Check for validation error format: {"success": false, "validation": [...]}
		if success, hasSuccess := objMap["success"].(bool); hasSuccess && !success {
			if validationList, hasValidation := objMap["validation"]; hasValidation {
				return formatValidationErrors(validationList)
			}
			// success=false but no validation array, generic error
			return fmt.Errorf("API operation failed: %s", limitString(string(body), 500))
		}

		// Check for standard error field format: {"error": "message", "code": ...}
		if errMsg, exists := objMap["error"]; exists {
			errCode := objMap["code"] // May be nil
			return fmt.Errorf("API error (code: %v): %v", errCode, errMsg)
		}

		// Object without error or success field - could be valid response object
		// Don't error on this - let caller handle object response
		return nil
	}

	// Layer 3: JSON array - success (may be empty)
	if _, ok := jsonData.([]interface{}); ok {
		return nil // Success
	}

	// Also handle boolean responses (e.g., from login)
	if _, ok := jsonData.(bool); ok {
		return nil // Success
	}

	// Unknown JSON type
	return fmt.Errorf("unexpected JSON type in response: %s", limitString(string(body), 500))
}

// formatValidationErrors parses BCM validation error array and formats as error message
// BCM validation format: {"success": false, "validation": [{"message": "...", "field": "..."}]}.
func formatValidationErrors(validationData interface{}) error {
	validationArray, ok := validationData.([]interface{})
	if !ok {
		return fmt.Errorf("validation error (invalid format): %v", validationData)
	}

	if len(validationArray) == 0 {
		return fmt.Errorf("validation failed (no details provided)")
	}

	// Build error message from validation array
	var errorMessages []string
	for _, item := range validationArray {
		if validationMap, ok := item.(map[string]interface{}); ok {
			message := "unknown error"
			field := ""

			if msg, hasMsg := validationMap["message"].(string); hasMsg {
				message = msg
			}
			if fld, hasFld := validationMap["field"].(string); hasFld {
				field = fld
			}

			if field != "" {
				errorMessages = append(errorMessages, fmt.Sprintf("%s: %s", field, message))
			} else {
				errorMessages = append(errorMessages, message)
			}
		}
	}

	if len(errorMessages) == 0 {
		return fmt.Errorf("validation failed: %v", validationArray)
	}

	if len(errorMessages) == 1 {
		return fmt.Errorf("validation error: %s", errorMessages[0])
	}

	// Multiple validation errors
	return fmt.Errorf("validation errors: %s", limitString(fmt.Sprintf("%v", errorMessages), 500))
}

// limitString truncates string to maxLen with ellipsis.
func limitString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... (truncated)"
}

// ValidateEntity calls BCM's validate* API method to validate an entity before CREATE or UPDATE.
// This provides pre-flight validation with field-specific error messages.
//
// Parameters:
//   - service: BCM service name (e.g., "CMPart", "CMDevice", "CMNet", "cmkube")
//   - validateMethod: BCM validation method (e.g., "validateSoftwareImage", "validateCategory")
//   - entity: Entity data to validate (same structure as used for add*/update* calls)
//   - isCreate: true for CREATE operations, false for UPDATE operations
//
// Returns:
//   - []ValidationError: Array of validation errors/warnings (empty if validation passes)
//   - error: API communication error (nil if API call succeeded)
//
// Zero UUID Filtering:
// During CREATE operations (isCreate=true), BCM returns expected "Zero UUID" errors since
// the entity doesn't have a UUID yet. These are automatically filtered out.
// During UPDATE operations (isCreate=false), all validation errors are preserved.
//
// Severity Handling:
//   - ERROR severity: Caller should halt operation and display error
//   - WARNING severity: Caller should display advisory but allow operation to proceed
//   - Unknown severity: Treated as ERROR for safety
func (c *BCMClient) ValidateEntity(ctx context.Context, service, validateMethod string, entity map[string]interface{}, isCreate bool) ([]ValidationError, error) {
	tflog.Debug(ctx, "Validating entity", map[string]interface{}{
		"service":        service,
		"validateMethod": validateMethod,
		"isCreate":       isCreate,
	})

	// Call BCM validation API
	body, err := c.CallJSONRPC(ctx, service, validateMethod, entity, false)
	if err != nil {
		tflog.Error(ctx, "Validation API call failed", map[string]interface{}{
			"service": service,
			"method":  validateMethod,
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("validation API call failed: %w", err)
	}

	// Parse validation response (should be array of validation objects or empty array)
	var validationArray []map[string]interface{}
	if err := json.Unmarshal(body, &validationArray); err != nil {
		tflog.Error(ctx, "Failed to parse validation response", map[string]interface{}{
			"service": service,
			"method":  validateMethod,
			"body":    string(body),
			"error":   err.Error(),
		})
		return nil, fmt.Errorf("validation response expected array, got: %s", limitString(string(body), 200))
	}

	// Convert to ValidationError structs
	var validationErrors []ValidationError
	for _, item := range validationArray {
		// Extract fields using null-safe helpers
		// BCM API returns lowercase field names: field, message, error_code, severity
		field := getString(item, "field")
		message := getString(item, "message")
		errorCode := getString(item, "error_code")
		severity := getString(item, "severity")
		entityUUID := getString(item, "ref_entity_uuid")

		// Create ValidationError struct
		valErr := ValidationError{
			Field:      field,
			Message:    message,
			ErrorCode:  errorCode,
			Severity:   severity,
			EntityUUID: entityUUID,
		}

		// Filter Zero UUID errors for CREATE operations
		// During CREATE, entities don't have UUIDs yet, so BCM returns expected "Zero UUID" errors
		if isCreate && field == "uuid" && (errorCode == "NOT_NULL" || contains(message, "Zero UUID")) {
			tflog.Debug(ctx, "Filtering expected Zero UUID error for CREATE operation", map[string]interface{}{
				"field":   field,
				"message": message,
			})
			continue
		}

		validationErrors = append(validationErrors, valErr)
	}

	tflog.Debug(ctx, "Validation complete", map[string]interface{}{
		"service":            service,
		"method":             validateMethod,
		"total_errors":       len(validationArray),
		"filtered_errors":    len(validationErrors),
		"zero_uuid_filtered": len(validationArray) - len(validationErrors),
	})

	return validationErrors, nil
}

// getString safely extracts string value from map, returning empty string if not found or wrong type.
func getString(data map[string]interface{}, key string) string {
	if val, ok := data[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// contains checks if a string contains a substring (case-sensitive).
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && indexOfSubstring(s, substr) >= 0))
}

// indexOfSubstring returns the index of substr in s, or -1 if not found.
func indexOfSubstring(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}
