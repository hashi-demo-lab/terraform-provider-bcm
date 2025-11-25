# Implementation Plan: Fix Mock Server Validation Response Format

## Overview

This plan addresses the bug in mock server test infrastructure where validation API responses return incorrect format.

## Design Decisions

### Approach: Centralized Validation Handler

Rather than adding validation handling to each individual handler function, we'll add validation call detection at the beginning of each affected handler. This approach:
- Maintains the existing handler structure
- Minimizes code changes
- Makes the validation handling explicit and clear

### Response Format

**For successful validation (no errors)**:
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_, _ = w.Write([]byte("[]"))
```

**For validation failure (with errors)**:
```go
w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
_ = json.NewEncoder(w).Encode([]map[string]interface{}{
    {"Field": "hostname", "Message": "...", "Severity": "ERROR"},
})
```

## Implementation Details

### File: `internal/provider/resource_cmdevice_device_mock_test.go`

### Change 1: Update handleDeviceCreateError (lines ~421-451)

Add validation handling before the default response:

```go
func handleDeviceCreateError(w http.ResponseWriter, req struct {...}) {
    // Return valid category and partition
    if req.Service == "cmdevice" && req.Call == "getCategory" {
        // ... existing code ...
    }
    if req.Service == "CMPart" && req.Call == "getSoftwareImage" {
        // ... existing code ...
    }

    // ADD: Handle validation calls - return empty array for success
    if req.Call == "validateDevice" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("[]"))
        return
    }

    // Fail on addDevice
    if req.Service == "cmdevice" && req.Call == "addDevice" {
        // ... existing code ...
    }
    // ... rest of function ...
}
```

### Change 2: Update handleDeviceCreateInvalidJSON (lines ~453-487)

Similar pattern - add validation handling:

```go
func handleDeviceCreateInvalidJSON(w http.ResponseWriter, req struct {...}) {
    // ... existing handlers ...

    // ADD: Handle validation calls
    if req.Call == "validateDevice" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("[]"))
        return
    }

    // Return invalid JSON only for addDevice
    // ... existing code ...
}
```

### Change 3: Update handleDeviceValidationFailure (lines ~489-531)

This handler tests validation failure, so it needs to return validation errors in the correct format AND remove the incorrect validation from addDevice response:

```go
func handleDeviceValidationFailure(w http.ResponseWriter, req struct {...}) {
    // ... existing category and partition handlers ...

    // CHANGE: Return validation errors in correct array format
    if req.Call == "validateDevice" {
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

    // REMOVE/SIMPLIFY: addDevice handling (won't be reached if validation fails)
    // But keep for backwards compatibility in case validation is bypassed
    if req.Service == "cmdevice" && req.Call == "addDevice" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
        return
    }

    // ... default response ...
}
```

### Change 4: Update handleDeviceReadError (lines ~533-575)

Add validation handling:

```go
func handleDeviceReadError(w http.ResponseWriter, req struct {...}) {
    // ... existing handlers ...

    // ADD: Handle validation calls
    if req.Call == "validateDevice" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("[]"))
        return
    }

    // Allow device creation to succeed
    // ... existing code ...
}
```

### Change 5: Update handleDeviceReadInvalidJSON (lines ~577-619)

Add validation handling:

```go
func handleDeviceReadInvalidJSON(w http.ResponseWriter, req struct {...}) {
    // ... existing handlers ...

    // ADD: Handle validation calls
    if req.Call == "validateDevice" {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("[]"))
        return
    }

    // Return success for device creation
    // ... existing code ...
}
```

### Change 6: Update Test Expectations

For `TestAccCMDeviceDeviceResource_ErrorDeviceValidationFailed`, update the expected error message to match the validation response format:

```go
{
    Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
    ExpectError: regexp.MustCompile(`(?s)Validation Error.*hostname.*already exists`),
},
```

## Test Verification Plan

### Step 1: Run affected tests before changes
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "^TestAccCMDeviceDeviceResource_Error(DeviceCreateFailed|DeviceCreateInvalidJSON|DeviceValidationFailed|DeviceReadAfterCreateFailed|DeviceReadInvalidJSON)$"
```
Expected: All 5 tests FAIL

### Step 2: Apply changes

### Step 3: Run affected tests after changes
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "^TestAccCMDeviceDeviceResource_Error"
```
Expected: All tests PASS

### Step 4: Run full mock test suite
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Mock|mock"
```
Expected: All mock tests PASS

### Step 5: Run full test suite
```bash
make test
```
Expected: All tests PASS

## Rollback Plan

If issues are found:
1. Revert changes to `resource_cmdevice_device_mock_test.go`
2. Re-run tests to confirm original state

## Dependencies

None - this is a self-contained test infrastructure fix.

## Timeline

Estimated completion: Immediate (< 30 minutes of implementation + testing)
