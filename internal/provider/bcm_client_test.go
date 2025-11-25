// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TestNewBCMClient_Success tests successful client creation and login.
func TestNewBCMClient_Success(t *testing.T) {
	// Create mock server that returns successful login response
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify login request structure
		if r.URL.Path != "/json" {
			t.Errorf("Expected path /json, got %s", r.URL.Path)
		}
		if r.Method != "POST" {
			t.Errorf("Expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse login request body
		var loginReq LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
			t.Fatalf("Failed to decode login request: %v", err)
		}

		// Verify login credentials
		if loginReq.Service != "login" {
			t.Errorf("Expected service 'login', got '%s'", loginReq.Service)
		}
		if loginReq.Username != "testuser" {
			t.Errorf("Expected username 'testuser', got '%s'", loginReq.Username)
		}
		if loginReq.Password != "testpass" {
			t.Errorf("Expected password 'testpass', got '%s'", loginReq.Password)
		}

		// Set authentication cookie
		http.SetCookie(w, &http.Cookie{
			Name:     "cm-login-token",
			Value:    "test-token-12345",
			Path:     "/",
			HttpOnly: true,
			Secure:   true,
		})

		// Return successful login response
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("true")); err != nil {
			t.Logf("Failed to write login response: %v", err)
		}
	}))
	defer server.Close()

	// Create client with mock server
	ctx := context.Background()
	client, err := NewBCMClient(ctx, server.URL, "testuser", "testpass", true, 30)

	// Verify client creation succeeded
	if err != nil {
		t.Fatalf("Expected successful client creation, got error: %v", err)
	}
	if client == nil {
		t.Fatal("Expected non-nil client")
	}
	if client.HTTPClient == nil {
		t.Fatal("Expected non-nil HTTP client")
	}
	if client.HTTPClient.Jar == nil {
		t.Fatal("Expected non-nil cookie jar")
	}
	if client.Endpoint != server.URL {
		t.Errorf("Expected endpoint %s, got %s", server.URL, client.Endpoint)
	}
}

// TestNewBCMClient_LoginFailure_InvalidCredentials tests login failure with 401.
func TestNewBCMClient_LoginFailure_InvalidCredentials(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(`{"error": "Invalid credentials"}`)); err != nil {
			t.Logf("Failed to write error response: %v", err)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := NewBCMClient(ctx, server.URL, "baduser", "badpass", true, 30)

	// Verify client creation failed
	if err == nil {
		t.Fatal("Expected error for invalid credentials, got nil")
	}
	if client != nil {
		t.Errorf("Expected nil client on failure, got %v", client)
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("Expected error to mention HTTP 401, got: %v", err)
	}
}

// TestNewBCMClient_LoginFailure_MissingCookie tests login failure when cookie is missing.
func TestNewBCMClient_LoginFailure_MissingCookie(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return success but no cookie
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("true")); err != nil {
			t.Logf("Failed to write response: %v", err)
		}
		// Intentionally not setting cm-login-token cookie
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := NewBCMClient(ctx, server.URL, "testuser", "testpass", true, 30)

	// Verify client creation failed
	if err == nil {
		t.Fatal("Expected error for missing cookie, got nil")
	}
	if client != nil {
		t.Errorf("Expected nil client on failure, got %v", client)
	}
	if !strings.Contains(err.Error(), "cm-login-token") {
		t.Errorf("Expected error to mention cm-login-token, got: %v", err)
	}
}

// TestNewBCMClient_LoginFailure_FalseResponse tests login failure with false response.
func TestNewBCMClient_LoginFailure_FalseResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.SetCookie(w, &http.Cookie{
			Name:  "cm-login-token",
			Value: "test-token",
		})
		w.WriteHeader(http.StatusOK)
		if _, err := w.Write([]byte("false")); err != nil { // Login failed
			t.Logf("Failed to write login response: %v", err)
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := NewBCMClient(ctx, server.URL, "testuser", "testpass", true, 30)

	// Verify client creation failed
	if err == nil {
		t.Fatal("Expected error for false login response, got nil")
	}
	if client != nil {
		t.Errorf("Expected nil client on failure, got %v", client)
	}
	if !strings.Contains(err.Error(), "login failed") {
		t.Errorf("Expected error to mention login failed, got: %v", err)
	}
}

// TestCallJSONRPC_Success tests successful JSON-RPC call without args.
func TestCallJSONRPC_Success(t *testing.T) {
	// Track login and API calls
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// First call: login
			http.SetCookie(w, &http.Cookie{
				Name:  "cm-login-token",
				Value: "test-token",
			})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Second call: actual API call
			var req JSONRPCRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Fatalf("Failed to decode JSONRPC request: %v", err)
			}

			// Verify request structure
			if req.Service != "CMDevice" {
				t.Errorf("Expected service 'CMDevice', got '%s'", req.Service)
			}
			if req.Call != "getNodes" {
				t.Errorf("Expected call 'getNodes', got '%s'", req.Call)
			}
			if len(req.Args) != 0 {
				t.Errorf("Expected no args, got %d args", len(req.Args))
			}

			// Return successful response
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`[{"hostname": "node1"}, {"hostname": "node2"}]`)); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, err := NewBCMClient(ctx, server.URL, "testuser", "testpass", true, 30)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	// Make JSON-RPC call
	response, err := client.CallJSONRPC(ctx, "CMDevice", "getNodes")
	if err != nil {
		t.Fatalf("Expected successful call, got error: %v", err)
	}

	// Verify response
	var nodes []map[string]interface{}
	if err := json.Unmarshal(response, &nodes); err != nil {
		t.Fatalf("Failed to parse response: %v", err)
	}
	if len(nodes) != 2 {
		t.Errorf("Expected 2 nodes, got %d", len(nodes))
	}
}

// TestCallJSONRPC_WithArgs tests JSON-RPC call with arguments.
func TestCallJSONRPC_WithArgs(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// API call with args
			var req JSONRPCRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("Failed to decode request: %v", err)
			}

			// Verify args parameter is present
			if len(req.Args) != 1 {
				t.Errorf("Expected 1 arg, got %d args", len(req.Args))
			}
			if req.Args[0] != "test-image" {
				t.Errorf("Expected arg 'test-image', got '%v'", req.Args[0])
			}

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"name": "test-image"}`)); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	// Make call with single argument
	response, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", "test-image")
	if err != nil {
		t.Fatalf("Expected successful call, got error: %v", err)
	}

	// Verify response contains expected data
	if !strings.Contains(string(response), "test-image") {
		t.Errorf("Expected response to contain 'test-image', got: %s", string(response))
	}
}

// TestCallJSONRPC_WithMultipleArgs tests JSON-RPC call with multiple arguments.
func TestCallJSONRPC_WithMultipleArgs(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// API call with multiple args
			var req JSONRPCRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				t.Logf("Failed to decode request: %v", err)
			}

			// Verify multiple args
			if len(req.Args) != 2 {
				t.Errorf("Expected 2 args, got %d", len(req.Args))
			}

			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"success": true}`)); err != nil {
				t.Logf("Failed to write response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	// Make call with multiple arguments
	entity := map[string]interface{}{"name": "test"}
	_, err := client.CallJSONRPC(ctx, "CMPart", "addSoftwareImage", entity, false)
	if err != nil {
		t.Fatalf("Expected successful call, got error: %v", err)
	}
}

// TestParseErrorResponse_Success tests successful response parsing.
func TestParseErrorResponse_Success(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		body       string
		wantError  bool
	}{
		{
			name:       "Empty array success",
			statusCode: 200,
			body:       "[]",
			wantError:  false,
		},
		{
			name:       "Array with data",
			statusCode: 200,
			body:       `[{"name": "test"}]`,
			wantError:  false,
		},
		{
			name:       "Boolean true",
			statusCode: 200,
			body:       "true",
			wantError:  false,
		},
		{
			name:       "Object response",
			statusCode: 200,
			body:       `{"name": "test", "value": 123}`,
			wantError:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := parseErrorResponse(tc.statusCode, []byte(tc.body))
			if tc.wantError && err == nil {
				t.Errorf("Expected error, got nil")
			}
			if !tc.wantError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

// TestParseErrorResponse_Errors tests error response parsing.
func TestParseErrorResponse_Errors(t *testing.T) {
	testCases := []struct {
		name       string
		statusCode int
		body       string
		wantError  string
	}{
		{
			name:       "HTTP 401",
			statusCode: 401,
			body:       "Unauthorized",
			wantError:  "HTTP 401",
		},
		{
			name:       "HTTP 500",
			statusCode: 500,
			body:       "Internal Server Error",
			wantError:  "HTTP 500",
		},
		{
			name:       "JSON error field",
			statusCode: 200,
			body:       `{"error": "Something went wrong", "code": 1234}`,
			wantError:  "API error",
		},
		{
			name:       "Invalid JSON",
			statusCode: 200,
			body:       "not json",
			wantError:  "failed to parse JSON",
		},
		{
			name:       "Validation error format",
			statusCode: 200,
			body:       `{"success": false, "validation": [{"field": "name", "message": "Name is required"}]}`,
			wantError:  "validation error",
		},
		{
			name:       "Success false without validation",
			statusCode: 200,
			body:       `{"success": false}`,
			wantError:  "API operation failed",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := parseErrorResponse(tc.statusCode, []byte(tc.body))
			if err == nil {
				t.Fatalf("Expected error containing '%s', got nil", tc.wantError)
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.wantError, err)
			}
		})
	}
}

// TestFormatValidationErrors tests validation error formatting.
func TestFormatValidationErrors(t *testing.T) {
	testCases := []struct {
		name       string
		input      interface{}
		wantError  string
		wantSubstr string
	}{
		{
			name: "Single validation error",
			input: []interface{}{
				map[string]interface{}{
					"field":   "name",
					"message": "Name is required",
				},
			},
			wantError:  "validation error",
			wantSubstr: "name: Name is required",
		},
		{
			name: "Multiple validation errors",
			input: []interface{}{
				map[string]interface{}{"field": "name", "message": "Name is required"},
				map[string]interface{}{"field": "path", "message": "Path is invalid"},
			},
			wantError:  "validation errors",
			wantSubstr: "name: Name is required",
		},
		{
			name:       "Empty validation array",
			input:      []interface{}{},
			wantError:  "validation failed",
			wantSubstr: "no details",
		},
		{
			name:       "Invalid format",
			input:      "not an array",
			wantError:  "validation error",
			wantSubstr: "invalid format",
		},
		{
			name: "Validation without field",
			input: []interface{}{
				map[string]interface{}{
					"message": "General error",
				},
			},
			wantError:  "validation error",
			wantSubstr: "General error",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := formatValidationErrors(tc.input)
			if err == nil {
				t.Fatalf("Expected error, got nil")
			}
			if !strings.Contains(err.Error(), tc.wantError) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.wantError, err)
			}
			if !strings.Contains(err.Error(), tc.wantSubstr) {
				t.Errorf("Expected error to contain '%s', got: %v", tc.wantSubstr, err)
			}
		})
	}
}

// TestLimitString tests string truncation.
func TestLimitString(t *testing.T) {
	testCases := []struct {
		name      string
		input     string
		maxLen    int
		wantLen   int
		wantTrunc bool
	}{
		{
			name:      "Short string not truncated",
			input:     "hello",
			maxLen:    10,
			wantLen:   5,
			wantTrunc: false,
		},
		{
			name:      "Exact length not truncated",
			input:     "hello",
			maxLen:    5,
			wantLen:   5,
			wantTrunc: false,
		},
		{
			name:      "Long string truncated",
			input:     "this is a very long string that needs truncation",
			maxLen:    10,
			wantLen:   10 + len("... (truncated)"),
			wantTrunc: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := limitString(tc.input, tc.maxLen)
			if len(result) != tc.wantLen {
				t.Errorf("Expected length %d, got %d (result: %s)", tc.wantLen, len(result), result)
			}
			if tc.wantTrunc && !strings.Contains(result, "truncated") {
				t.Errorf("Expected truncated string to contain 'truncated', got: %s", result)
			}
			if !tc.wantTrunc && strings.Contains(result, "truncated") {
				t.Errorf("Expected non-truncated string, got: %s", result)
			}
		})
	}
}

// TestCallJSONRPC_HTTPError tests error handling for failed HTTP requests.
func TestCallJSONRPC_HTTPError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Return HTTP error
			w.WriteHeader(http.StatusInternalServerError)
			if _, err := w.Write([]byte(`{"error": "Internal server error"}`)); err != nil {
				t.Logf("Failed to write error response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	// Make call that should fail
	_, err := client.CallJSONRPC(ctx, "CMDevice", "getNodes")
	if err == nil {
		t.Fatal("Expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "500") {
		t.Errorf("Expected error to mention HTTP 500, got: %v", err)
	}
}

// TestCallJSONRPC_ValidationError tests handling of BCM validation errors.
func TestCallJSONRPC_ValidationError(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Return validation error
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{
				"success": false,
				"validation": [
					{"field": "path", "message": "The software image path does not exist"},
					{"field": "kernelVersion", "message": "Specified kernel does not exist"}
				]
			}`)); err != nil {
				t.Logf("Failed to write validation error response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	// Make call that returns validation errors
	_, err := client.CallJSONRPC(ctx, "CMPart", "addSoftwareImage", map[string]interface{}{})
	if err == nil {
		t.Fatal("Expected validation error, got nil")
	}
	if !strings.Contains(err.Error(), "validation") {
		t.Errorf("Expected error to mention validation, got: %v", err)
	}
	if !strings.Contains(err.Error(), "path") {
		t.Errorf("Expected error to mention path field, got: %v", err)
	}
}

// TestValidateEntity_Success tests successful validation with empty array response.
func TestValidateEntity_Success(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Validation success - empty array
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("[]")); err != nil {
				t.Logf("Failed to write validation response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	// Call ValidateEntity with valid entity
	entity := map[string]interface{}{"name": "test-image", "path": "/valid/path"}
	validationErrors, err := client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity, true)

	// Verify no errors
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("Expected 0 validation errors, got %d", len(validationErrors))
	}
}

// TestValidateEntity_ErrorResponse tests validation with ERROR severity response.
func TestValidateEntity_ErrorResponse(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Validation error with ERROR severity
			w.WriteHeader(http.StatusOK)
			validationResponse := `[{
				"baseType": "Validation",
				"Field": "SOLSpeed",
				"Message": "SOL speed must be one of: 9600, 19200, 38400, 57600, 115200",
				"ErrorCode": "BAD_VALUE",
				"Severity": "ERROR",
				"EntityUUID": ""
			}]`
			if _, err := w.Write([]byte(validationResponse)); err != nil {
				t.Logf("Failed to write validation response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	entity := map[string]interface{}{"name": "test-image", "SOLSpeed": 999999}
	validationErrors, err := client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity, true)

	// Verify validation errors returned
	if err != nil {
		t.Fatalf("Expected no API error, got: %v", err)
	}
	if len(validationErrors) != 1 {
		t.Fatalf("Expected 1 validation error, got %d", len(validationErrors))
	}

	// Verify error details
	valErr := validationErrors[0]
	if valErr.Field != "SOLSpeed" {
		t.Errorf("Expected Field 'SOLSpeed', got '%s'", valErr.Field)
	}
	if valErr.Severity != "ERROR" {
		t.Errorf("Expected Severity 'ERROR', got '%s'", valErr.Severity)
	}
	if !valErr.IsError() {
		t.Error("Expected IsError() to return true for ERROR severity")
	}
	if valErr.IsWarning() {
		t.Error("Expected IsWarning() to return false for ERROR severity")
	}
}

// TestValidateEntity_WarningResponse tests validation with WARNING severity response.
func TestValidateEntity_WarningResponse(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Validation warning with WARNING severity
			w.WriteHeader(http.StatusOK)
			validationResponse := `[{
				"baseType": "Validation",
				"Field": "path",
				"Message": "The software image path does not exist on the server",
				"ErrorCode": "WARNING",
				"Severity": "WARNING",
				"EntityUUID": ""
			}]`
			if _, err := w.Write([]byte(validationResponse)); err != nil {
				t.Logf("Failed to write validation response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	entity := map[string]interface{}{"name": "test-image", "path": "/nonexistent/path"}
	validationErrors, err := client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity, true)

	// Verify validation warnings returned
	if err != nil {
		t.Fatalf("Expected no API error, got: %v", err)
	}
	if len(validationErrors) != 1 {
		t.Fatalf("Expected 1 validation warning, got %d", len(validationErrors))
	}

	// Verify warning details
	valErr := validationErrors[0]
	if valErr.Field != "path" {
		t.Errorf("Expected Field 'path', got '%s'", valErr.Field)
	}
	if valErr.Severity != "WARNING" {
		t.Errorf("Expected Severity 'WARNING', got '%s'", valErr.Severity)
	}
	if valErr.IsError() {
		t.Error("Expected IsError() to return false for WARNING severity")
	}
	if !valErr.IsWarning() {
		t.Error("Expected IsWarning() to return true for WARNING severity")
	}
}

// TestValidateEntity_ZeroUUIDFiltering tests Zero UUID filtering for CREATE operations.
func TestValidateEntity_ZeroUUIDFiltering(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Validation response with Zero UUID error (expected during CREATE)
			w.WriteHeader(http.StatusOK)
			validationResponse := `[
				{
					"baseType": "Validation",
					"Field": "uuid",
					"Message": "Zero UUID not allowed",
					"ErrorCode": "NOT_NULL",
					"Severity": "ERROR",
					"EntityUUID": ""
				},
				{
					"baseType": "Validation",
					"Field": "path",
					"Message": "Path is required",
					"ErrorCode": "NOT_NULL",
					"Severity": "ERROR",
					"EntityUUID": ""
				}
			]`
			if _, err := w.Write([]byte(validationResponse)); err != nil {
				t.Logf("Failed to write validation response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	entity := map[string]interface{}{"name": "test-image"}

	// Test CREATE (isCreate=true) - should filter Zero UUID error
	validationErrors, err := client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity, true)
	if err != nil {
		t.Fatalf("Expected no API error, got: %v", err)
	}
	if len(validationErrors) != 1 {
		t.Fatalf("Expected 1 validation error after filtering, got %d", len(validationErrors))
	}
	if validationErrors[0].Field == "uuid" {
		t.Error("Expected Zero UUID error to be filtered for CREATE operation")
	}
	if validationErrors[0].Field != "path" {
		t.Errorf("Expected remaining error to be for 'path', got '%s'", validationErrors[0].Field)
	}
}

// TestValidateEntity_MalformedResponse tests handling of malformed validation response.
func TestValidateEntity_MalformedResponse(t *testing.T) {
	callCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if callCount == 1 {
			// Login
			http.SetCookie(w, &http.Cookie{Name: "cm-login-token", Value: "token"})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			// Malformed validation response (not an array)
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte(`{"error": "unexpected"}`)); err != nil {
				t.Logf("Failed to write malformed response: %v", err)
			}
		}
	}))
	defer server.Close()

	ctx := context.Background()
	client, _ := NewBCMClient(ctx, server.URL, "user", "pass", true, 30)

	entity := map[string]interface{}{"name": "test-image"}
	validationErrors, err := client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity, true)

	// Verify error returned for malformed response
	if err == nil {
		t.Fatal("Expected error for malformed response, got nil")
	}
	if !strings.Contains(err.Error(), "expected array") && !strings.Contains(err.Error(), "unexpected") {
		t.Errorf("Expected error to mention unexpected format, got: %v", err)
	}
	if len(validationErrors) != 0 {
		t.Errorf("Expected 0 validation errors on API error, got %d", len(validationErrors))
	}
}
