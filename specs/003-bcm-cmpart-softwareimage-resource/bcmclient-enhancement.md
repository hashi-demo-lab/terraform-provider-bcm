# BCMClient Enhancement: Args Parameter Support

**Date**: 2025-11-20
**Status**: Recommended for Phase 1 (Setup) of Feature 003
**Priority**: HIGH (blocks efficient Read implementation)

---

## Problem Statement

The current `BCMClient.CallJSONRPC` method does NOT support passing arguments to API methods:

```go
// Current signature (bcm_client.go:147)
func (c *BCMClient) CallJSONRPC(ctx context.Context, service, call string) ([]byte, error)

// JSONRPCRequest struct explicitly excludes args (bcm_client.go:30-34)
type JSONRPCRequest struct {
    Service string `json:"service"`
    Call    string `json:"call"`
    // arg field intentionally omitted - POV limitation
}
```

This limitation prevents efficient API calls like:
- `getSoftwareImage(name)` - Direct lookup by name
- `updateSoftwareImage(entity, force)` - Update with force parameter
- `removeSoftwareImage(uuid, removeData, removeAll, force)` - Delete with options

---

## Verification Results

**Test Script**: `/workspace/test_bcmclient_args.go`

**Test Results** (2025-11-20):
```
✓ Test 1: Login (no args) - PASS
✓ Test 2: getSoftwareImages() (no args) - PASS (returned 2 images)
✓ Test 3: getSoftwareImage(name) WITH args - PASS

Request:  {"service":"cmpart","call":"getSoftwareImage","args":["default-image"]}
Response: HTTP 200, returned single SoftwareImage entity

📋 VERDICT: ✅ Args parameter IS supported by BCM API
```

**Conclusion**: The BCM API DOES support the `args` parameter. The limitation is CLIENT-SIDE ONLY.

---

## Recommended Solution

### Option 1: Extend CallJSONRPC (Recommended)

**Implementation**:

```go
// File: internal/provider/bcm_client.go

// Update JSONRPCRequest struct to include optional args
type JSONRPCRequest struct {
    Service string        `json:"service"`
    Call    string        `json:"call"`
    Args    []interface{} `json:"args,omitempty"` // Add args parameter
}

// Update CallJSONRPC signature to accept variadic args
func (c *BCMClient) CallJSONRPC(ctx context.Context, service, call string, args ...interface{}) ([]byte, error) {
    reqBody := JSONRPCRequest{
        Service: service,
        Call:    call,
    }

    // Only add args if provided
    if len(args) > 0 {
        reqBody.Args = args
    }

    jsonBody, err := json.Marshal(reqBody)
    if err != nil {
        return nil, fmt.Errorf("failed to marshal JSONRPC request: %w", err)
    }

    // ... rest of implementation unchanged
}
```

**Benefits**:
- ✅ Backward compatible (existing calls with no args continue to work)
- ✅ Consistent API across all provider code
- ✅ Type-safe with Go's variadic parameters
- ✅ Minimal code changes

**Migration Path**:
```go
// Before (still works - backward compatible)
client.CallJSONRPC(ctx, "cmdevice", "getNodes")

// After (new feature)
client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName)
client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, force)
client.CallJSONRPC(ctx, "CMPart", "removeSoftwareImage", uuid, removeData, removeAll, force)
```

---

### Option 2: New Method (Alternative)

**Implementation**:

```go
// Add new method alongside existing CallJSONRPC
func (c *BCMClient) CallJSONRPCWithArgs(ctx context.Context, service, call string, args ...interface{}) ([]byte, error) {
    reqBody := map[string]interface{}{
        "service": service,
        "call":    call,
        "args":    args,
    }
    // ... rest of implementation
}
```

**Benefits**:
- ✅ No risk to existing code
- ✅ Clear separation of concerns

**Drawbacks**:
- ❌ Two similar methods with confusing naming
- ❌ Requires updating all future code to use new method
- ❌ Less elegant API

---

### Option 3: Direct HTTP POST (Temporary Workaround)

**Use Case**: If BCMClient cannot be modified immediately

**Implementation Example**:

```go
func (r *CMPartSoftwareImageResource) callAPIWithArgs(ctx context.Context, service, call string, args ...interface{}) ([]byte, error) {
    reqBody := map[string]interface{}{
        "service": service,
        "call":    call,
        "args":    args,
    }

    jsonBody, _ := json.Marshal(reqBody)
    req, _ := http.NewRequestWithContext(ctx, "POST", r.client.Endpoint+"/json", bytes.NewReader(jsonBody))
    req.Header.Set("Content-Type", "application/json")

    resp, err := r.client.HTTPClient.Do(req)
    // ... handle response
}
```

**Drawbacks**:
- ❌ Bypasses BCMClient error handling and logging
- ❌ Duplicate code across resources
- ❌ Not maintainable long-term

---

## Recommendation

**Use Option 1: Extend CallJSONRPC**

**Rationale**:
1. Backward compatible with all existing code
2. Minimal implementation effort (~10 lines of code)
3. Benefits all future provider development
4. Type-safe and elegant API
5. Consistent with Go idioms (variadic parameters)

**Implementation Timeline**:
- **Phase 1 (Setup)**: Extend BCMClient.CallJSONRPC
- **Phase 2 (RED)**: Use new signature in acceptance tests
- **Phase 3 (GREEN)**: Hardcoded implementation (no API calls yet)
- **Phase 4 (REFACTOR)**: Full API integration with args parameter

---

## Testing Strategy

**Unit Tests** (internal/provider/bcm_client_test.go):

```go
func TestCallJSONRPC_NoArgs(t *testing.T) {
    // Test backward compatibility
    client := setupTestClient()
    _, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    // Assert success
}

func TestCallJSONRPC_WithArgs(t *testing.T) {
    // Test new args parameter
    client := setupTestClient()
    _, err := client.CallJSONRPC(ctx, "cmpart", "getSoftwareImage", "test-image")
    // Assert success
}

func TestCallJSONRPC_MultipleArgs(t *testing.T) {
    // Test multiple arguments
    client := setupTestClient()
    entity := map[string]interface{}{"name": "test"}
    _, err := client.CallJSONRPC(ctx, "cmpart", "updateSoftwareImage", entity, true)
    // Assert success
}
```

**Integration Tests**:
- Acceptance tests for Feature 003 will validate args support end-to-end
- Use test script `/workspace/test_bcmclient_args.go` for manual verification

---

## Impact Assessment

**Affected Code**:
- `internal/provider/bcm_client.go` (BCMClient implementation)
- `internal/provider/data_source_cmdevice_nodes.go` (no changes needed - backward compatible)
- `internal/provider/resource_cmpart_softwareimage.go` (NEW - will use args parameter)

**Breaking Changes**: NONE (backward compatible)

**Performance Impact**: Negligible (just JSON marshaling of args array)

---

## Action Items

- [ ] **HIGH**: Update `internal/provider/bcm_client.go` with variadic args support
- [ ] **HIGH**: Add unit tests for args parameter (no args, single arg, multiple args)
- [ ] **MEDIUM**: Update CLAUDE.md to document new CallJSONRPC signature
- [ ] **LOW**: Update existing comments referencing "POV limitation"

---

## References

- **Test Script**: `/workspace/test_bcmclient_args.go`
- **Research**: `/workspace/specs/003-bcm-cmpart-softwareimage-resource/research.md` (section: API Quirks)
- **Current Implementation**: `/workspace/internal/provider/bcm_client.go` (lines 27-34, 147-194)
- **BCM API Documentation**: `/workspace/sampleRest/CMDevice_Complete_Documentation.md`

---

## Appendix: Example Usage

### Before (Current - Limited)

```go
// Data source: Works fine (no args needed)
body, err := d.client.CallJSONRPC(ctx, "cmdevice", "getNodes")
```

### After (Enhanced - Supports Args)

```go
// Data source: Still works (backward compatible)
body, err := d.client.CallJSONRPC(ctx, "cmdevice", "getNodes")

// Resource Read: Efficient direct lookup
body, err := r.client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName)

// Resource Create: With force parameter
entity := buildSoftwareImageEntity(data)
force := data.Force.ValueBool()
body, err := r.client.CallJSONRPC(ctx, "CMPart", "addSoftwareImage", entity, force)

// Resource Update: With validation
entity := buildSoftwareImageEntity(data)
// Validate first
_, err := r.client.CallJSONRPC(ctx, "CMPart", "validateSoftwareImage", entity)
if err == nil {
    // Then update
    _, err = r.client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, false)
}

// Resource Delete: With options
uuid := data.UUID.ValueString()
body, err := r.client.CallJSONRPC(ctx, "CMPart", "removeSoftwareImage", uuid, false, false, false)
```

---

**End of Enhancement Recommendation**
