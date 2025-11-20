# Phase 0: Research Findings - BCM Terraform Provider POV

**Date**: 2025-11-20
**Feature**: BCM Terraform Provider POV
**Branch**: 001-bcm-provider

This document captures technical research findings that resolve unknowns from the Technical Context and establish implementation patterns for the POV.

---

## 1. JSON-RPC Client Design

**Decision**: Custom HTTP client with JSON-RPC request builder and defensive response parser

**Rationale**:
- BCM uses non-standard JSON-RPC pattern (POST /json with {"service":"...", "call":"..."} body)
- Standard JSON-RPC libraries assume id/method/params structure, not service/call
- Custom client gives full control over request format, headers, and error handling
- Simplifies integration with terraform-plugin-framework provider lifecycle

**Alternatives Considered**:
1. **Standard JSON-RPC library (github.com/ybbus/jsonrpc)**: Rejected - assumes standard JSON-RPC 2.0 format with method/params, incompatible with BCM's service/call pattern
2. **Generic HTTP client wrapper**: Rejected - adds abstraction overhead without BCM-specific value
3. **Third-party BCM SDK**: Rejected - no official Go SDK available, POV validates API directly

**Implementation Notes**:

**Client Structure**:
```go
// BCMClient handles JSON-RPC API calls to BCM with cookie-based authentication
type BCMClient struct {
    HTTPClient *http.Client  // Includes cookie jar for automatic cm-login-token management
    Endpoint   string        // Base URL (e.g., https://172.21.15.254:8081)
}

// JSONRPCRequest represents BCM JSON-RPC request body
type JSONRPCRequest struct {
    Service  string `json:"service"`
    Call     string `json:"call"`
    // NOTE: "arg" field excluded from POV scope (parameter passing deferred to post-POV)
}

// LoginRequest represents login API call request body
type LoginRequest struct {
    Service  string `json:"service"`   // Always "login"
    Username string `json:"username"`
    Password string `json:"password"`
}

// NewBCMClient creates authenticated client with cookie jar and performs login
func NewBCMClient(ctx context.Context, endpoint, username, password string, insecureSkipVerify bool, timeout int) (*BCMClient, error)

// CallJSONRPC executes JSON-RPC call using authenticated cookie jar
func (c *BCMClient) CallJSONRPC(ctx context.Context, service, call string) ([]byte, error)
```

**Authentication Flow** (executed once during NewBCMClient):
1. Create HTTP client with cookie jar (cookiejar.New())
2. Construct LoginRequest with service="login", username, password
3. POST to endpoint + "/json" with login credentials
4. Verify response is JSON boolean `true`
5. Verify Set-Cookie header contains cm-login-token
6. Cookie jar automatically stores cm-login-token for subsequent requests
7. Return authenticated BCMClient

**Request Flow** (for CallJSONRPC):
1. Construct JSONRPCRequest with service and call fields
2. Marshal to JSON body
3. Create POST request to endpoint + "/json"
4. Set headers: Content-Type: application/json (Cookie header automatically added by cookie jar)
5. Execute with http.Client.Do(req.WithContext(ctx))
6. Defensive error parsing (see section 3)
7. Return raw response bytes for caller to unmarshal

**Code Example - NewBCMClient with Login**:
```go
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
    for _, cookie := range cookies {
        if cookie.Name == "cm-login-token" {
            hasLoginToken = true
            break
        }
    }
    if !hasLoginToken {
        return nil, fmt.Errorf("login response missing cm-login-token cookie")
    }

    // Return authenticated client (cookie jar now contains cm-login-token)
    return &BCMClient{
        HTTPClient: client,
        Endpoint:   endpoint,
    }, nil
}

// CallJSONRPC executes JSON-RPC call using authenticated cookie jar
func (c *BCMClient) CallJSONRPC(ctx context.Context, service, call string) ([]byte, error) {
    reqBody := JSONRPCRequest{
        Service: service,
        Call:    call,
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

    resp, err := c.HTTPClient.Do(req)
    if err != nil {
        return nil, fmt.Errorf("JSONRPC call failed: %w", err)
    }
    defer resp.Body.Close()

    body, err := io.ReadAll(resp.Body)
    if err != nil {
        return nil, fmt.Errorf("failed to read response body: %w", err)
    }

    // Defensive error parsing (see section 3)
    if err := parseErrorResponse(resp.StatusCode, body); err != nil {
        return nil, err
    }

    return body, nil
}
```

---

## 2. TLS Configuration for Self-Signed Certificates

**Decision**: Configurable http.Transport with InsecureSkipVerify option controlled by provider attribute

**Rationale**:
- POV BCM instance uses self-signed certificate (https://172.21.15.254:8081)
- Default Go TLS client rejects self-signed certs (x509: certificate signed by unknown authority)
- insecure_skip_verify provider attribute gives users explicit control
- Aligns with Terraform provider patterns (e.g., terraform-provider-vsphere)
- Security risk clearly communicated via attribute name and documentation

**Alternatives Considered**:
1. **Always skip TLS verification**: Rejected - insecure by default, no path to production hardening
2. **CA certificate bundle attribute**: Rejected - adds complexity for POV, deferred to post-POV with certificate-based auth
3. **Auto-detect and skip**: Rejected - implicit behavior confusing, prefer explicit user choice

**Implementation Notes**:

**Provider Schema**:
```go
"insecure_skip_verify": schema.BoolAttribute{
    MarkdownDescription: "Skip TLS certificate verification. WARNING: This makes connections susceptible to man-in-the-middle attacks. Only use for testing with self-signed certificates.",
    Optional:            true,
    // Default: false (verify certificates)
}
```

**Client Initialization**:
```go
func NewBCMClient(endpoint, username, password string, insecureSkipVerify bool, timeout int) (*BCMClient, error) {
    transport := &http.Transport{
        TLSClientConfig: &tls.Config{
            InsecureSkipVerify: insecureSkipVerify,
        },
    }

    client := &http.Client{
        Transport: transport,
        Timeout:   time.Duration(timeout) * time.Second,
    }

    return &BCMClient{
        HTTPClient: client,
        Endpoint:   endpoint,
        Username:   username,
        Password:   password,
    }, nil
}
```

**Error Guidance**:
When TLS verification fails (insecure_skip_verify=false with self-signed cert):
```
Error: TLS certificate verification failed

The BCM endpoint uses a certificate that cannot be verified. This is common with
self-signed certificates in test environments.

To proceed with testing, set insecure_skip_verify = true in the provider configuration:

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "your-password"
  insecure_skip_verify = true
}

WARNING: This disables TLS security checks. For production use, configure certificate-based
authentication (see documentation for mTLS setup).
```

**Security Documentation**:
- Quickstart guide includes prominent warning about insecure_skip_verify
- Provider documentation explains security implications
- Post-POV roadmap includes certificate-based auth (mTLS) for production

---

## 3. Defensive Error Parsing Strategy

**Decision**: Multi-layer error detection with HTTP status, JSON structure analysis, and error field checking

**Rationale**:
- BCM JSON-RPC API error format unknown (spec clarification: "NEEDS CLARIFICATION")
- Must handle multiple error scenarios: HTTP status, JSON error objects, connection failures
- Defensive approach catches errors regardless of API response format
- Logging at TRACE level captures full response for debugging unknown formats

**Alternatives Considered**:
1. **HTTP status only**: Rejected - may miss errors returned with HTTP 200 + error JSON
2. **Assume error object format**: Rejected - brittle if API uses different format
3. **Try/catch all as errors**: Rejected - valid empty arrays [] would be treated as errors

**Implementation Notes**:

**Error Detection Layers** (in order):

**Layer 1: HTTP Status Code**
- Check: resp.StatusCode < 200 || resp.StatusCode >= 300
- Action: Return HTTP error immediately (don't parse body)
- Example: 401 Unauthorized, 500 Internal Server Error

**Layer 2: JSON Structure Detection**
- Check: Response body is JSON object (not array)
- Action: Check for "error" field in object
- Example: {"error": "Authentication failed", "code": 401}

**Layer 3: Empty Array Success**
- Check: Response body is empty JSON array []
- Action: Return success (not error)
- Rationale: getSoftwareImages returns [] when no images exist (valid state)

**Layer 4: Parse Errors**
- Check: Response body is not valid JSON
- Action: Return parse error with response snippet
- Example: HTML error page instead of JSON

**Code Example**:
```go
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

    // Layer 2: JSON object with error field
    if objMap, ok := jsonData.(map[string]interface{}); ok {
        if errMsg, exists := objMap["error"]; exists {
            errCode := objMap["code"] // May be nil
            return fmt.Errorf("API error (code: %v): %v", errCode, errMsg)
        }
        // Object without error field - unexpected format
        return fmt.Errorf("unexpected JSON object response (expected array): %s", limitString(string(body), 500))
    }

    // Layer 3: JSON array - success (may be empty)
    if _, ok := jsonData.([]interface{}); ok {
        return nil // Success
    }

    // Unknown JSON type
    return fmt.Errorf("unexpected JSON type in response: %s", limitString(string(body), 500))
}

func limitString(s string, maxLen int) string {
    if len(s) <= maxLen {
        return s
    }
    return s[:maxLen] + "... (truncated)"
}
```

**Logging Strategy**:
```go
// In CallJSONRPC, before parseErrorResponse:
tflog.Trace(ctx, "JSONRPC request", map[string]interface{}{
    "service":  service,
    "call":     call,
    "endpoint": c.Endpoint + "/json",
})

// After receiving response:
tflog.Trace(ctx, "JSONRPC response", map[string]interface{}{
    "status": resp.StatusCode,
    "body":   string(body), // Full response for debugging
})
```

**Error Message Format**:
```
Error: BCM API call failed

Failed to execute JSON-RPC call to BCM endpoint.

Endpoint: https://172.21.15.254:8081/json
Service: CMPart
Call: getSoftwareImages
HTTP Status: 401
Response: {"error": "Authentication failed", "code": 401}

Check your username and password in the provider configuration.
```

---

## 4. Nested Schema Design for terraform-plugin-framework

**Decision**: ListNestedAttribute for SoftwareImage array, nested ListNestedAttribute for modules array, snake_case conversion via helper function

**Rationale**:
- terraform-plugin-framework provides ListNestedAttribute for complex nested structures
- 30+ SoftwareImage fields require structured schema definition
- Nested modules array requires nested ListNestedAttribute within SoftwareImage schema
- camelCase→snake_case conversion maintains Terraform conventions
- All fields Computed: true (read-only data source)

**Alternatives Considered**:
1. **Flatten modules to separate data source**: Rejected - loses parent-child relationship, requires multiple API calls
2. **JSON string for modules**: Rejected - forces users to parse JSON in HCL, poor UX
3. **Map[string]interface{} for flexibility**: Rejected - loses type safety, no schema validation

**Implementation Notes**:

**Schema Structure**:
```go
func (d *CMPartSoftwareImagesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Fetches software images from BCM CMPart service",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Placeholder identifier for data source",
            },
            "images": schema.ListNestedAttribute{
                Computed:            true,
                MarkdownDescription: "List of software images",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        // Identity fields
                        "id": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Software image identifier (same as uuid)",
                        },
                        "uuid": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Software image UUID",
                        },
                        "name": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Software image name",
                        },

                        // ... (25+ more fields - see full schema below)

                        // Nested modules array
                        "modules": schema.ListNestedAttribute{
                            Computed:            true,
                            MarkdownDescription: "Kernel modules configured for this image",
                            NestedObject: schema.NestedAttributeObject{
                                Attributes: map[string]schema.Attribute{
                                    "uuid": schema.StringAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module UUID",
                                    },
                                    "name": schema.StringAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module name",
                                    },
                                    "parameters": schema.StringAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module parameters",
                                    },
                                    "base_type": schema.StringAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module base type",
                                    },
                                    "child_type": schema.StringAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module child type",
                                    },
                                    "revision": schema.StringAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module revision",
                                    },
                                    "modified": schema.BoolAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module modified flag",
                                    },
                                    "to_be_removed": schema.BoolAttribute{
                                        Computed:            true,
                                        MarkdownDescription: "Module scheduled for removal",
                                    },
                                },
                            },
                        },
                    },
                },
            },
        },
    }
}
```

**Model Structures**:
```go
type CMPartSoftwareImagesDataSourceModel struct {
    ID     types.String         `tfsdk:"id"`
    Images []SoftwareImageModel `tfsdk:"images"`
}

type SoftwareImageModel struct {
    // Identity
    ID   types.String `tfsdk:"id"`   // Mapped from uuid
    UUID types.String `tfsdk:"uuid"`
    Name types.String `tfsdk:"name"`
    Path types.String `tfsdk:"path"`

    // Kernel configuration
    KernelVersion       types.String `tfsdk:"kernel_version"`        // camelCase → snake_case
    KernelParameters    types.String `tfsdk:"kernel_parameters"`
    KernelOutputConsole types.String `tfsdk:"kernel_output_console"`

    // Partitions
    BootfsPart types.String `tfsdk:"bootfs_part"`
    FsPart     types.String `tfsdk:"fs_part"`

    // Serial Over LAN
    EnableSOL       types.Bool   `tfsdk:"enable_sol"`
    SOLPort         types.String `tfsdk:"sol_port"`
    SOLSpeed        types.String `tfsdk:"sol_speed"`
    SOLFlowControl  types.Bool   `tfsdk:"sol_flow_control"`

    // Metadata
    BaseType     types.String `tfsdk:"base_type"`
    ChildType    types.String `tfsdk:"child_type"`
    CreationTime types.Int64  `tfsdk:"creation_time"`
    Revision     types.String `tfsdk:"revision"`
    RevisionID   types.Int64  `tfsdk:"revision_id"`

    // Relationships
    OriginalImage        types.String `tfsdk:"original_image"`
    ParentSoftwareImage  types.String `tfsdk:"parent_software_image"`
    ParentUUID           types.String `tfsdk:"parent_uuid"`

    // State flags
    FileOperationInProgress types.Bool `tfsdk:"file_operation_in_progress"`
    Modified                types.Bool `tfsdk:"modified"`
    ToBeRemoved             types.Bool `tfsdk:"to_be_removed"`

    // Notes
    Notes types.String `tfsdk:"notes"`

    // Nested modules
    Modules []KernelModuleModel `tfsdk:"modules"`
}

type KernelModuleModel struct {
    UUID        types.String `tfsdk:"uuid"`
    Name        types.String `tfsdk:"name"`
    Parameters  types.String `tfsdk:"parameters"`
    BaseType    types.String `tfsdk:"base_type"`
    ChildType   types.String `tfsdk:"child_type"`
    Revision    types.String `tfsdk:"revision"`
    Modified    types.Bool   `tfsdk:"modified"`
    ToBeRemoved types.Bool   `tfsdk:"to_be_removed"`
}
```

**Field Mapping Helper**:
```go
// mapAPIResponseToModel converts API JSON to SoftwareImageModel
func mapAPIResponseToModel(apiData map[string]interface{}) SoftwareImageModel {
    model := SoftwareImageModel{}

    // String fields with camelCase conversion
    model.UUID = getStringValue(apiData, "uuid")
    model.ID = model.UUID // Use UUID as ID
    model.Name = getStringValue(apiData, "name")
    model.Path = getStringValue(apiData, "path")
    model.KernelVersion = getStringValue(apiData, "kernelVersion")
    model.KernelParameters = getStringValue(apiData, "kernelParameters")
    model.KernelOutputConsole = getStringValue(apiData, "kernelOutputConsole")
    model.BootfsPart = getStringValue(apiData, "bootfspart")
    model.FsPart = getStringValue(apiData, "fspart")
    model.SOLPort = getStringValue(apiData, "SOLPort")
    model.SOLSpeed = getStringValue(apiData, "SOLSpeed")
    model.BaseType = getStringValue(apiData, "baseType")
    model.ChildType = getStringValue(apiData, "childType")
    model.Revision = getStringValue(apiData, "revision")
    model.OriginalImage = getStringValue(apiData, "originalImage")
    model.ParentSoftwareImage = getStringValue(apiData, "parentSoftwareImage")
    model.ParentUUID = getStringValue(apiData, "parent_uuid")
    model.Notes = getStringValue(apiData, "notes")

    // Bool fields
    model.EnableSOL = getBoolValue(apiData, "enableSOL")
    model.SOLFlowControl = getBoolValue(apiData, "SOLFlowControl")
    model.FileOperationInProgress = getBoolValue(apiData, "fileOperationInProgress")
    model.Modified = getBoolValue(apiData, "modified")
    model.ToBeRemoved = getBoolValue(apiData, "to_be_removed")

    // Int64 fields
    model.CreationTime = getInt64Value(apiData, "creationTime")
    model.RevisionID = getInt64Value(apiData, "revisionID")

    // Nested modules array
    if modulesData, ok := apiData["modules"].([]interface{}); ok {
        model.Modules = make([]KernelModuleModel, 0, len(modulesData))
        for _, modData := range modulesData {
            if modMap, ok := modData.(map[string]interface{}); ok {
                module := KernelModuleModel{
                    UUID:        getStringValue(modMap, "uuid"),
                    Name:        getStringValue(modMap, "name"),
                    Parameters:  getStringValue(modMap, "parameters"),
                    BaseType:    getStringValue(modMap, "baseType"),
                    ChildType:   getStringValue(modMap, "childType"),
                    Revision:    getStringValue(modMap, "revision"),
                    Modified:    getBoolValue(modMap, "modified"),
                    ToBeRemoved: getBoolValue(modMap, "to_be_removed"),
                }
                model.Modules = append(model.Modules, module)
            }
        }
    } else {
        model.Modules = []KernelModuleModel{} // Empty slice if modules null/missing
    }

    return model
}

// Helper functions for null-safe field extraction
func getStringValue(data map[string]interface{}, key string) types.String {
    if val, ok := data[key]; ok && val != nil {
        if str, ok := val.(string); ok && str != "" {
            return types.StringValue(str)
        }
    }
    return types.StringNull()
}

func getBoolValue(data map[string]interface{}, key string) types.Bool {
    if val, ok := data[key]; ok && val != nil {
        if b, ok := val.(bool); ok {
            return types.BoolValue(b)
        }
    }
    return types.BoolNull()
}

func getInt64Value(data map[string]interface{}, key string) types.Int64 {
    if val, ok := data[key]; ok && val != nil {
        switch v := val.(type) {
        case float64:
            return types.Int64Value(int64(v))
        case int64:
            return types.Int64Value(v)
        case int:
            return types.Int64Value(int64(v))
        }
    }
    return types.Int64Null()
}
```

**Usage in Read Method**:
```go
func (d *CMPartSoftwareImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var state CMPartSoftwareImagesDataSourceModel

    // Get client from provider
    client, ok := req.ProviderData.(*BCMClient)
    if !ok {
        resp.Diagnostics.AddError("Provider Data Error", "Expected *BCMClient")
        return
    }

    // Call API
    body, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImages")
    if err != nil {
        resp.Diagnostics.AddError("API Call Failed", err.Error())
        return
    }

    // Parse response array
    var apiResponse []map[string]interface{}
    if err := json.Unmarshal(body, &apiResponse); err != nil {
        resp.Diagnostics.AddError("Parse Error", fmt.Sprintf("Failed to parse response: %v", err))
        return
    }

    // Map to models
    state.Images = make([]SoftwareImageModel, 0, len(apiResponse))
    for _, imgData := range apiResponse {
        state.Images = append(state.Images, mapAPIResponseToModel(imgData))
    }

    // Set placeholder ID
    state.ID = types.StringValue("placeholder")

    // Save state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)

    tflog.Debug(ctx, "Successfully read software images", map[string]interface{}{
        "count": len(state.Images),
    })
}
```

---

## 5. Data Source Acceptance Test Patterns

**Decision**: terraform-plugin-testing with testAccProtoV6ProviderFactories, focusing on Read operation and schema validation (no ImportState, no Update)

**Rationale**:
- Data sources are read-only (no Create/Update/Delete operations to test)
- No ImportState for data sources (only resources)
- Focus on: Read success, empty responses, nested attributes, error handling
- Environment variables for BCM credentials (TF_ACC=1 gates real API calls)

**Alternatives Considered**:
1. **Mock HTTP server**: Rejected for POV - want to validate against real BCM API
2. **Fixture-based tests**: Rejected - doesn't validate actual API integration
3. **Integration tests separate from acceptance**: Rejected - Terraform provider pattern is acceptance tests with real API

**Implementation Notes**:

**Test Setup** (internal/provider/provider_test.go):
```go
package provider

import (
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "bcm": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
    // Check for required environment variables
    if v := os.Getenv("BCM_ENDPOINT"); v == "" {
        t.Fatal("BCM_ENDPOINT must be set for acceptance tests")
    }
    if v := os.Getenv("BCM_USERNAME"); v == "" {
        t.Fatal("BCM_USERNAME must be set for acceptance tests")
    }
    if v := os.Getenv("BCM_PASSWORD"); v == "" {
        t.Fatal("BCM_PASSWORD must be set for acceptance tests")
    }
}
```

**Test Pattern 1: Basic Read**:
```go
func TestAccCMPartSoftwareImagesDataSource_Basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMPartSoftwareImagesDataSourceConfig,
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify data source exists
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "id"),
                    // Verify images attribute exists (may be empty list)
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.#"),
                ),
            },
        },
    })
}

const testAccCMPartSoftwareImagesDataSourceConfig = `
provider "bcm" {
  endpoint             = "` + os.Getenv("BCM_ENDPOINT") + `"
  username             = "` + os.Getenv("BCM_USERNAME") + `"
  password             = "` + os.Getenv("BCM_PASSWORD") + `"
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {}
`
```

**Test Pattern 2: Nested Attributes**:
```go
func TestAccCMPartSoftwareImagesDataSource_NestedModules(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMPartSoftwareImagesDataSourceConfig,
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Check first image exists
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.id"),
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.name"),
                    // Check modules nested attribute exists
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.modules.#"),
                ),
            },
        },
    })
}
```

**Test Pattern 3: All Fields Validation**:
```go
func TestAccCMPartSoftwareImagesDataSource_AllFields(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMPartSoftwareImagesDataSourceConfig,
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify all 30+ fields accessible (using first image)
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.id"),
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.uuid"),
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.name"),
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.path"),
                    resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.0.kernel_version"),
                    // ... (check all fields - see full test file)
                ),
            },
        },
    })
}
```

**Test Pattern 4: Error Handling**:
```go
func TestAccCMPartSoftwareImagesDataSource_InvalidCredentials(t *testing.T) {
    resource.Test(t, resource.TestCase{
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: `
provider "bcm" {
  endpoint             = "` + os.Getenv("BCM_ENDPOINT") + `"
  username             = "invalid"
  password             = "invalid"
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {}
`,
                ExpectError: regexp.MustCompile(`(authentication|401|unauthorized)`),
            },
        },
    })
}
```

**Running Acceptance Tests**:
```bash
# Set environment variables
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Run all acceptance tests
go test -v -timeout 120m ./internal/provider

# Run specific test
go test -v -run TestAccCMPartSoftwareImagesDataSource_Basic -timeout 120m ./internal/provider
```

---

## Research Summary

All unknowns from Technical Context resolved:

1. **JSON-RPC Client Design**: Custom HTTP client with service/call request pattern
2. **Self-Signed Certificates**: Configurable InsecureSkipVerify with user control via provider attribute
3. **Error Parsing**: Multi-layer defensive parsing (HTTP status, JSON structure, error fields)
4. **Nested Schema**: ListNestedAttribute for images and modules with camelCase→snake_case helpers
5. **Acceptance Tests**: terraform-plugin-testing patterns for read-only data sources

**Next Phase**: Phase 1 Design - Generate data-model.md, contracts/, and quickstart.md based on research findings.
