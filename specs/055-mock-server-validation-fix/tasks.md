# Tasks: Fix Mock Server Validation Response Format

## Task Checklist

- [ ] Task 1: Verify failing tests (RED phase)
- [ ] Task 2: Update handleDeviceCreateError handler
- [ ] Task 3: Update handleDeviceCreateInvalidJSON handler
- [ ] Task 4: Update handleDeviceValidationFailure handler
- [ ] Task 5: Update handleDeviceReadError handler
- [ ] Task 6: Update handleDeviceReadInvalidJSON handler
- [ ] Task 7: Update TestAccCMDeviceDeviceResource_ErrorDeviceValidationFailed expectations
- [ ] Task 8: Run affected tests (GREEN phase)
- [ ] Task 9: Run full test suite
- [ ] Task 10: Lint and format code
- [ ] Task 11: Commit changes

---

## Task 1: Verify failing tests (RED phase)

**Description**: Run the affected tests to confirm they fail with the expected error pattern.

**Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "^TestAccCMDeviceDeviceResource_Error(DeviceCreateFailed|DeviceCreateInvalidJSON|DeviceValidationFailed|DeviceReadAfterCreateFailed|DeviceReadInvalidJSON)$" 2>&1 | head -100
```

**Expected Output**: All 5 tests fail with "validation response expected array" error.

**Exit Criteria**: Tests fail with expected error message.

---

## Task 2: Update handleDeviceCreateError handler

**File**: `internal/provider/resource_cmdevice_device_mock_test.go`
**Lines**: ~421-451

**Change**: Add validation call handling before the default response.

**Code to Add** (after partition handling, before addDevice check):
```go
// Handle validation calls - return empty array for success
if req.Call == "validateDevice" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("[]"))
    return
}
```

**Exit Criteria**: Handler returns `[]` for validateDevice calls.

---

## Task 3: Update handleDeviceCreateInvalidJSON handler

**File**: `internal/provider/resource_cmdevice_device_mock_test.go`
**Lines**: ~453-487

**Change**: Add validation call handling before the addDevice check.

**Code to Add** (after partition handling, before addDevice check):
```go
// Handle validation calls - return empty array for success
if req.Call == "validateDevice" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("[]"))
    return
}
```

**Exit Criteria**: Handler returns `[]` for validateDevice calls.

---

## Task 4: Update handleDeviceValidationFailure handler

**File**: `internal/provider/resource_cmdevice_device_mock_test.go`
**Lines**: ~489-531

**Change**: Replace the addDevice validation response with proper validateDevice response.

**Code to Replace**: The `addDevice` handler that returns validation errors in object format.

**New Code**:
```go
// Handle validation calls - return validation errors in array format
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
```

**Also**: Update or simplify the addDevice handler since validation will fail before reaching it.

**Exit Criteria**: Handler returns validation errors in array format for validateDevice calls.

---

## Task 5: Update handleDeviceReadError handler

**File**: `internal/provider/resource_cmdevice_device_mock_test.go`
**Lines**: ~533-575

**Change**: Add validation call handling before the addDevice check.

**Code to Add** (after partition handling, before addDevice check):
```go
// Handle validation calls - return empty array for success
if req.Call == "validateDevice" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("[]"))
    return
}
```

**Exit Criteria**: Handler returns `[]` for validateDevice calls.

---

## Task 6: Update handleDeviceReadInvalidJSON handler

**File**: `internal/provider/resource_cmdevice_device_mock_test.go`
**Lines**: ~577-619

**Change**: Add validation call handling before the addDevice check.

**Code to Add** (after partition handling, before addDevice check):
```go
// Handle validation calls - return empty array for success
if req.Call == "validateDevice" {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    _, _ = w.Write([]byte("[]"))
    return
}
```

**Exit Criteria**: Handler returns `[]` for validateDevice calls.

---

## Task 7: Update TestAccCMDeviceDeviceResource_ErrorDeviceValidationFailed expectations

**File**: `internal/provider/resource_cmdevice_device_mock_test.go`
**Lines**: ~999-1027

**Change**: Update the ExpectError regex to match the new validation error format.

**Current**:
```go
ExpectError: regexp.MustCompile(`(?is)Error Creating Device.*validation.*errors.*hostname.*already.*exists`),
```

**New**:
```go
ExpectError: regexp.MustCompile(`(?is)Validation Error.*hostname.*already exists`),
```

**Exit Criteria**: Test expects the correct error message format from validation.

---

## Task 8: Run affected tests (GREEN phase)

**Description**: Run the affected tests to confirm they pass.

**Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "^TestAccCMDeviceDeviceResource_Error(DeviceCreateFailed|DeviceCreateInvalidJSON|DeviceValidationFailed|DeviceReadAfterCreateFailed|DeviceReadInvalidJSON)$"
```

**Expected Output**: All 5 tests PASS.

**Exit Criteria**: 5/5 tests pass.

---

## Task 9: Run full test suite

**Description**: Run all mock tests to ensure no regressions.

**Command**:
```bash
make test
```

**Expected Output**: All tests PASS.

**Exit Criteria**: Full test suite passes.

---

## Task 10: Lint and format code

**Description**: Ensure code meets style guidelines.

**Commands**:
```bash
make fmt
make lint
```

**Expected Output**: No lint errors.

**Exit Criteria**: Code passes all linting checks.

---

## Task 11: Commit changes

**Description**: Commit the fix with appropriate message.

**Command**:
```bash
git add internal/provider/resource_cmdevice_device_mock_test.go
git commit -m "fix(tests): correct mock server validation response format

Updates mock server handlers to return array format ([]) for validation
API calls instead of object format ({\"success\": true}).

The BCM API ValidateEntity() function expects array responses for
validation calls. The mock server was returning incorrect format,
causing 5 device error tests to fail.

Changes:
- handleDeviceCreateError: Add validateDevice handler returning []
- handleDeviceCreateInvalidJSON: Add validateDevice handler returning []
- handleDeviceValidationFailure: Return validation errors in array format
- handleDeviceReadError: Add validateDevice handler returning []
- handleDeviceReadInvalidJSON: Add validateDevice handler returning []
- Update test expectations for validation error format

Fixes #55"
```

**Exit Criteria**: Changes committed with descriptive message.

---

## Summary

| Task | Description | Dependencies |
|------|-------------|--------------|
| 1 | Verify failing tests | None |
| 2 | Update handleDeviceCreateError | Task 1 |
| 3 | Update handleDeviceCreateInvalidJSON | Task 1 |
| 4 | Update handleDeviceValidationFailure | Task 1 |
| 5 | Update handleDeviceReadError | Task 1 |
| 6 | Update handleDeviceReadInvalidJSON | Task 1 |
| 7 | Update test expectations | Task 4 |
| 8 | Run affected tests | Tasks 2-7 |
| 9 | Run full test suite | Task 8 |
| 10 | Lint and format | Task 9 |
| 11 | Commit changes | Task 10 |

**Parallelizable**: Tasks 2-6 can be done in parallel after Task 1.
