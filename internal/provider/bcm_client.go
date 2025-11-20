// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/cookiejar"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// BCMClient handles JSON-RPC API calls to BCM with cookie-based authentication
type BCMClient struct {
	HTTPClient *http.Client // Includes cookie jar for automatic cm-login-token management
	Endpoint   string       // Base URL (e.g., https://172.21.15.254:8081)
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

// LoginRequest represents login API call request body
type LoginRequest struct {
	Service  string `json:"service"` // Always "login"
	Username string `json:"username"`
	Password string `json:"password"`
}

// NewBCMClient creates authenticated client with cookie jar and performs login
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

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("login API call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read login response: %w", err)
	}

	// Verify HTTP status
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
	var loginCookie *http.Cookie
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
		HTTPClient: client,
		Endpoint:   endpoint,
	}, nil
}

// CallJSONRPC executes JSON-RPC call using authenticated cookie jar
// Supports optional variadic args parameter for parameterized API calls
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

	resp, err := c.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("JSONRPC call failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	tflog.Trace(ctx, "JSONRPC response", map[string]interface{}{
		"status": resp.StatusCode,
		"body":   string(body),
	})

	// Defensive error parsing (see section 3 of research.md)
	if err := parseErrorResponse(resp.StatusCode, body); err != nil {
		return nil, err
	}

	return body, nil
}

// parseErrorResponse performs multi-layer error detection
// Layer 1: HTTP status code
// Layer 2: JSON object with error field
// Layer 3: Empty array success
// Layer 4: Parse errors
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
// BCM validation format: {"success": false, "validation": [{"message": "...", "field": "..."}]}
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

// limitString truncates string to maxLen with ellipsis
func limitString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "... (truncated)"
}
