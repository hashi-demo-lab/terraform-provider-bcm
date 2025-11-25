# Feature Specification: Fix Mock Server Validation Response Format

## Overview

**Issue**: [#55](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/55)
**Type**: Bug Fix (Test Infrastructure)
**Priority**: Low
**Impact**: 5 tests failing (27.8% of device mock tests)

## Problem Statement

Five device resource error tests fail because the mock HTTP server returns `{"success":true}` instead of the expected empty array `[]` for validation responses.

### Error Pattern
```
validation response expected array, got: {"success":true}
```

### Affected Tests
All tests in `resource_cmdevice_device_mock_test.go`:
1. `TestAccCMDeviceDeviceResource_ErrorDeviceCreateFailed`
2. `TestAccCMDeviceDeviceResource_ErrorDeviceCreateInvalidJSON`
3. `TestAccCMDeviceDeviceResource_ErrorDeviceValidationFailed`
4. `TestAccCMDeviceDeviceResource_ErrorDeviceReadAfterCreateFailed`
5. `TestAccCMDeviceDeviceResource_ErrorDeviceReadInvalidJSON`

## Root Cause Analysis

### Current Behavior

The mock server handlers return a generic success response for unhandled API calls:

```go
// Default success response for unknown calls
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
```

### Expected Behavior

The real BCM API validation endpoints return:
- **Success**: `[]` (empty array)
- **Errors**: Array of validation error objects

```go
[]  // Empty array on success

// Or with validation errors:
[
  {
    "baseType": "Validation",
    "error_code": "BAD_VALUE",
    "field": "hostname",
    "message": "Invalid hostname format",
    "severity": "ERROR"
  }
]
```

### Code Path Analysis

**File**: `internal/provider/bcm_client.go:363-391`

```go
func (c *BCMClient) ValidateEntity(ctx context.Context, service, validateMethod string, entity map[string]interface{}, isCreate bool) ([]ValidationError, error) {
    // ...
    body, err := c.CallJSONRPC(ctx, service, validateMethod, entity, false)
    // ...

    // Parse validation response (should be array of validation objects or empty array)
    var validationArray []map[string]interface{}
    if err := json.Unmarshal(body, &validationArray); err != nil {
        return nil, fmt.Errorf("validation response expected array, got: %s", limitString(string(body), 200))
    }
    // ...
}
```

The `ValidateEntity()` function expects an array response. When the mock server returns `{"success": true}`, `json.Unmarshal` fails because it cannot unmarshal an object into `[]map[string]interface{}`.

## Solution Design

### Approach

Update mock server handlers to detect validation API calls and return the correct response format (empty array `[]` for successful validation).

### Implementation Strategy

1. **Identify validation calls**: Check if `req.Call` starts with "validate"
2. **Return correct format**: Return `[]` (empty array) for validation calls
3. **Preserve existing behavior**: Keep `{"success": true}` for non-validation calls

### Code Change

Update each affected handler in `resource_cmdevice_device_mock_test.go` to detect validation calls:

```go
// Handle validation calls - must return array format
if strings.HasPrefix(req.Call, "validate") {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode([]interface{}{}) // Empty array
    return
}
```

### Affected Handlers

The following handler functions need to be updated:
1. `handleDeviceCreateError` (line 421-451)
2. `handleDeviceCreateInvalidJSON` (line 453-487)
3. `handleDeviceValidationFailure` (line 489-531) - needs special handling for validation failure scenario
4. `handleDeviceReadError` (line 533-575)
5. `handleDeviceReadInvalidJSON` (line 577-619)

### Special Case: handleDeviceValidationFailure

The `handleDeviceValidationFailure` handler is specifically testing validation failure scenarios. It should return validation errors in the correct array format:

```go
// For validation failure scenario, return errors in array format
if req.Service == "cmdevice" && req.Call == "validateDevice" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _ = json.NewEncoder(w).Encode([]map[string]interface{}{
        {
            "Field":    "hostname",
            "Message":  "Hostname already exists in cluster",
            "Severity": "ERROR",
        },
        {
            "Field":    "mac",
            "Message":  "MAC address is already in use",
            "Severity": "ERROR",
        },
    })
    return
}
```

## Acceptance Criteria

1. All 5 affected tests pass
2. Existing passing tests remain passing
3. Mock server correctly simulates BCM API validation response format
4. No changes to production code required

## Verification

After implementation, all 5 tests should pass:

```bash
TF_ACC=1 go test -v ./internal/provider/ -run "^TestAccCMDeviceDeviceResource_Error"
```

Expected output: All tests PASS

## Non-Goals

- No changes to production validation code (already working correctly)
- No changes to real BCM API integration tests
- No additional test scenarios beyond fixing existing failures

## Risk Assessment

- **Risk Level**: Very Low
- **Reason**: Changes are isolated to mock test infrastructure
- **Production Impact**: None - test-only changes
