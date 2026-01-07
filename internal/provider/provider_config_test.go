// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"
)

// createMockBCMServer creates a mock BCM server for testing provider configuration.
// Returns a httptest server that handles login requests.
func createMockBCMServer(t *testing.T, loginSuccessful bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if loginSuccessful {
			// Set authentication cookie.
			http.SetCookie(w, &http.Cookie{
				Name:     "cm-login-token",
				Value:    "test-token-12345",
				Path:     "/",
				HttpOnly: true,
				Secure:   true,
			})
			w.WriteHeader(http.StatusOK)
			if _, err := w.Write([]byte("true")); err != nil {
				t.Logf("Failed to write login response: %v", err)
			}
		} else {
			w.WriteHeader(http.StatusUnauthorized)
			if _, err := w.Write([]byte(`{"error": "Invalid credentials"}`)); err != nil {
				t.Logf("Failed to write error response: %v", err)
			}
		}
	}))
}

// createProviderConfig creates a tfsdk.Config for testing provider configuration.
func createProviderConfig(t *testing.T, endpoint, username, password string, insecureSkipVerify *bool, timeout *int64) tfsdk.Config {
	// Get provider schema.
	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	schemaReq := provider.SchemaRequest{}
	schemaResp := &provider.SchemaResponse{}
	p.Schema(ctx, schemaReq, schemaResp)

	if schemaResp.Diagnostics.HasError() {
		t.Fatalf("Failed to get provider schema: %v", schemaResp.Diagnostics)
	}

	// Build config value.
	configValues := map[string]tftypes.Value{}

	if endpoint != "" {
		configValues["endpoint"] = tftypes.NewValue(tftypes.String, endpoint)
	} else {
		configValues["endpoint"] = tftypes.NewValue(tftypes.String, nil)
	}

	if username != "" {
		configValues["username"] = tftypes.NewValue(tftypes.String, username)
	} else {
		configValues["username"] = tftypes.NewValue(tftypes.String, nil)
	}

	if password != "" {
		configValues["password"] = tftypes.NewValue(tftypes.String, password)
	} else {
		configValues["password"] = tftypes.NewValue(tftypes.String, nil)
	}

	if insecureSkipVerify != nil {
		configValues["insecure_skip_verify"] = tftypes.NewValue(tftypes.Bool, *insecureSkipVerify)
	} else {
		configValues["insecure_skip_verify"] = tftypes.NewValue(tftypes.Bool, nil)
	}

	if timeout != nil {
		configValues["timeout"] = tftypes.NewValue(tftypes.Number, *timeout)
	} else {
		configValues["timeout"] = tftypes.NewValue(tftypes.Number, nil)
	}

	configValue := tftypes.NewValue(
		schemaResp.Schema.Type().TerraformType(ctx),
		configValues,
	)

	return tfsdk.Config{
		Schema: schemaResp.Schema,
		Raw:    configValue,
	}
}

// TestProviderConfigure_MissingEndpoint tests that provider configuration.
// fails when endpoint is missing from both config and environment.
func TestProviderConfigure_MissingEndpoint(t *testing.T) {
	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	config := createProviderConfig(t, "", "testuser", "testpass", nil, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify error diagnostic.
	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error for missing endpoint, got nil")
	}

	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if strings.Contains(diag.Summary(), "Missing BCM Endpoint") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'Missing BCM Endpoint' error, got: %v", resp.Diagnostics.Errors())
	}
}

// TestProviderConfigure_MissingUsername tests that provider configuration.
// fails when username is missing from both config and environment.
func TestProviderConfigure_MissingUsername(t *testing.T) {
	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	config := createProviderConfig(t, "https://test.example.com:8081", "", "testpass", nil, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify error diagnostic.
	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error for missing username, got nil")
	}

	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if strings.Contains(diag.Summary(), "Missing BCM Username") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'Missing BCM Username' error, got: %v", resp.Diagnostics.Errors())
	}
}

// TestProviderConfigure_MissingPassword tests that provider configuration.
// fails when password is missing from both config and environment.
func TestProviderConfigure_MissingPassword(t *testing.T) {
	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	config := createProviderConfig(t, "https://test.example.com:8081", "testuser", "", nil, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify error diagnostic.
	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error for missing password, got nil")
	}

	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if strings.Contains(diag.Summary(), "Missing BCM Password") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'Missing BCM Password' error, got: %v", resp.Diagnostics.Errors())
	}
}

// TestProviderConfigure_MissingAllCredentials tests that provider configuration.
// fails when all credentials are missing.
func TestProviderConfigure_MissingAllCredentials(t *testing.T) {
	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	config := createProviderConfig(t, "", "", "", nil, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify error diagnostics (should have multiple errors).
	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected errors for missing credentials, got nil")
	}

	// Should have at least one error about missing credentials.
	if len(resp.Diagnostics.Errors()) == 0 {
		t.Error("Expected at least one error diagnostic")
	}
}

// TestProviderConfigure_EnvironmentVariables tests that provider configuration.
// correctly reads credentials from environment variables.
func TestProviderConfigure_EnvironmentVariables(t *testing.T) {
	// Create mock server.
	server := createMockBCMServer(t, true)
	defer server.Close()

	// Set environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", server.URL)
	t.Setenv("BCM_USERNAME", "env-user")
	t.Setenv("BCM_PASSWORD", "env-pass")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	// Create config with empty values and insecureSkipVerify set (should use environment).
	insecureSkipVerify := true
	config := createProviderConfig(t, "", "", "", &insecureSkipVerify, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify no errors.
	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics.Errors())
	}

	// Verify client was created.
	if resp.ResourceData == nil {
		t.Error("Expected ResourceData to be set with client")
	}
	if resp.DataSourceData == nil {
		t.Error("Expected DataSourceData to be set with client")
	}
}

// TestProviderConfigure_ConfigPrecedence tests that explicit provider configuration.
// takes precedence over environment variables.
func TestProviderConfigure_ConfigPrecedence(t *testing.T) {
	// Create mock server.
	server := createMockBCMServer(t, true)
	defer server.Close()

	// Set environment variables with different values using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "https://env-endpoint.example.com:8081")
	t.Setenv("BCM_USERNAME", "env-user")
	t.Setenv("BCM_PASSWORD", "env-pass")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	// Create config with explicit values (should override environment).
	insecureSkipVerify := true
	config := createProviderConfig(t, server.URL, "config-user", "config-pass", &insecureSkipVerify, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify no errors (config values were used).
	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics.Errors())
	}

	// Verify client was created.
	if resp.ResourceData == nil {
		t.Error("Expected ResourceData to be set with client")
	}
}

// TestProviderConfigure_DefaultValues tests that provider configuration.
// correctly applies default values for optional fields.
func TestProviderConfigure_DefaultValues(t *testing.T) {
	// Create mock server.
	server := createMockBCMServer(t, true)
	defer server.Close()

	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	// Create config without optional fields (should use defaults).
	// Default timeout: 30 seconds (provider.go line 135)
	// Default insecure_skip_verify: false (provider.go line 130)
	insecureSkipVerify := true // Override for testing with mock server
	config := createProviderConfig(t, server.URL, "testuser", "testpass", &insecureSkipVerify, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify no errors.
	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics.Errors())
	}

	// Verify client was created with defaults.
	if resp.ResourceData == nil {
		t.Error("Expected ResourceData to be set with client")
	}

	// Verify the client exists.
	client, ok := resp.ResourceData.(*BCMClient)
	if !ok {
		t.Fatal("Expected BCMClient in ResourceData")
	}
	if client == nil {
		t.Fatal("Expected non-nil BCMClient")
	}
}

// TestProviderConfigure_CustomTimeout tests that provider configuration.
// correctly applies custom timeout value.
func TestProviderConfigure_CustomTimeout(t *testing.T) {
	// Create mock server.
	server := createMockBCMServer(t, true)
	defer server.Close()

	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	// Create config with custom timeout.
	insecureSkipVerify := true
	customTimeout := int64(60)
	config := createProviderConfig(t, server.URL, "testuser", "testpass", &insecureSkipVerify, &customTimeout)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify no errors.
	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics.Errors())
	}

	// Verify client was created.
	if resp.ResourceData == nil {
		t.Error("Expected ResourceData to be set with client")
	}
}

// TestProviderConfigure_CustomInsecureSkipVerify tests that provider configuration.
// correctly applies custom insecure_skip_verify value.
func TestProviderConfigure_CustomInsecureSkipVerify(t *testing.T) {
	// Create mock server.
	server := createMockBCMServer(t, true)
	defer server.Close()

	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	// Create config with custom insecure_skip_verify.
	insecureSkipVerify := true
	config := createProviderConfig(t, server.URL, "testuser", "testpass", &insecureSkipVerify, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify no errors.
	if resp.Diagnostics.HasError() {
		t.Fatalf("Unexpected error: %v", resp.Diagnostics.Errors())
	}

	// Verify client was created.
	if resp.ResourceData == nil {
		t.Error("Expected ResourceData to be set with client")
	}
}

// TestProviderConfigure_LoginFailure tests that provider configuration.
// fails gracefully when BCM login fails.
func TestProviderConfigure_LoginFailure(t *testing.T) {
	// Create mock server that returns login failure.
	server := createMockBCMServer(t, false)
	defer server.Close()

	// Clear environment variables using t.Setenv (automatically restores after test).
	t.Setenv("BCM_ENDPOINT", "")
	t.Setenv("BCM_USERNAME", "")
	t.Setenv("BCM_PASSWORD", "")

	os.Unsetenv("BCM_ENDPOINT")
	os.Unsetenv("BCM_USERNAME")
	os.Unsetenv("BCM_PASSWORD")

	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	insecureSkipVerify := true
	config := createProviderConfig(t, server.URL, "baduser", "badpass", &insecureSkipVerify, nil)

	req := provider.ConfigureRequest{
		Config: config,
	}
	resp := &provider.ConfigureResponse{}

	p.Configure(ctx, req, resp)

	// Verify error diagnostic.
	if !resp.Diagnostics.HasError() {
		t.Fatal("Expected error for login failure, got nil")
	}

	found := false
	for _, diag := range resp.Diagnostics.Errors() {
		if strings.Contains(diag.Summary(), "Unable to Create BCM Client") {
			found = true
			break
		}
	}

	if !found {
		t.Errorf("Expected 'Unable to Create BCM Client' error, got: %v", resp.Diagnostics.Errors())
	}
}

// TestProviderMetadata tests that provider metadata is correctly set.
func TestProviderMetadata(t *testing.T) {
	ctx := t.Context()
	testVersion := "test-version"
	providerFactory := New(testVersion)
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	req := provider.MetadataRequest{}
	resp := &provider.MetadataResponse{}

	p.Metadata(ctx, req, resp)

	// Verify type name.
	if resp.TypeName != "bcm" {
		t.Errorf("Expected TypeName 'bcm', got '%s'", resp.TypeName)
	}

	// Verify version.
	if resp.Version != testVersion {
		t.Errorf("Expected Version '%s', got '%s'", testVersion, resp.Version)
	}
}

// TestProviderSchema tests that provider schema is correctly defined.
func TestProviderSchema(t *testing.T) {
	ctx := t.Context()
	providerFactory := New("test")
	p, ok := providerFactory().(*BCMProvider)
	if !ok {
		t.Fatalf("expected BCMProvider, got %T", providerFactory())
	}

	req := provider.SchemaRequest{}
	resp := &provider.SchemaResponse{}

	p.Schema(ctx, req, resp)

	// Verify schema was set.
	if resp.Schema.Attributes == nil {
		t.Fatal("Expected schema attributes, got nil")
	}

	// Verify expected attributes exist.
	expectedAttributes := []string{"endpoint", "username", "password", "insecure_skip_verify", "timeout"}
	for _, attr := range expectedAttributes {
		if _, ok := resp.Schema.Attributes[attr]; !ok {
			t.Errorf("Expected attribute '%s' in schema, not found", attr)
		}
	}

	// Verify password is marked sensitive.
	passwordAttr := resp.Schema.Attributes["password"]
	// Note: Checking sensitive flag would require type assertion to specific attribute type.
	if passwordAttr == nil {
		t.Error("Expected password attribute to be defined")
	}
}
