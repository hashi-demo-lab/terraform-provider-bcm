# MAC Address Conflict Fix - TestAccCMDeviceDevice_DriftNotes

## Issue Summary

**Problem**: Test `TestAccCMDeviceDevice_DriftNotes` was failing with error:
```
A device with MAC 00:11:22:33:44:66 already exists
```

**Root Cause**: Hardcoded MAC addresses in test configurations caused conflicts when:
1. Previous test runs didn't clean up properly
2. Tests ran in parallel
3. Multiple test runs occurred in succession

## Solution Implemented

### 1. Utilized Existing Helper Function

The codebase already had a `generateUniqueMAC()` helper function in `/workspace/internal/provider/test_helpers.go` (lines 442-474) that generates unique MAC addresses with format `02:00:00:XX:YY:ZZ` using time-based components.

### 2. Modified Test Configuration Functions

Updated three configuration functions to accept a `mac` parameter:

- `testAccCMDeviceDeviceResourceConfig_Basic(hostname, categoryName, imageName, imagePath, mac)`
- `testAccCMDeviceDeviceResourceConfig_Updated(hostname, categoryName, imageName, imagePath, mac)`
- `testAccCMDeviceDeviceResourceConfig_Drift(hostname, categoryName, imageName, imagePath, mac)`
- `testAccCMDeviceDeviceResourceConfig_WithMockEndpoint(endpoint, hostname, categoryName, imageName, imagePath, mac)`

Changed hardcoded MAC addresses:
- `"00:11:22:33:44:55"` → `%[8]q` (parameter placeholder)
- `"00:11:22:33:44:66"` → `%[8]q` (parameter placeholder)
- `"00:11:22:33:44:77"` → `%[6]q` (for mock endpoint config)

### 3. Updated Test Functions

Added `mac := generateUniqueMAC()` and updated all configuration calls in:

**Test Functions Modified**:
1. `TestAccCMDeviceDeviceResource_Basic` - Full CRUD lifecycle test
2. `TestAccCMDeviceDevice_DriftNotes` - Drift detection for notes field
3. `TestAccCMDeviceDevice_Drift` - Drift detection for hostname attribute
4. `TestAccCMDeviceDevice_ValidationInvalidHostname` - Hostname validation tests
5. `TestAccCMDeviceDevice_PartitionCommitTimeout` - Timeout scenario test

**Example Change**:
```go
// Before
Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath),

// After  
mac := generateUniqueMAC()
Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath, mac),
```

### 4. Updated State Assertions

Changed hardcoded MAC assertions to use the generated variable:

```go
// Before
statecheck.ExpectKnownValue(
    "bcm_cmdevice_device.test",
    tfjsonpath.New("mac"),
    knownvalue.StringExact("00:11:22:33:44:55"),
),

// After
statecheck.ExpectKnownValue(
    "bcm_cmdevice_device.test",
    tfjsonpath.New("mac"),
    knownvalue.StringExact(mac),
),
```

## Benefits

1. **Eliminates MAC Address Conflicts**: Each test run generates unique MAC addresses
2. **Parallel Test Safety**: Tests can run in parallel without MAC collisions
3. **Cleanup Independence**: Tests don't fail due to incomplete cleanup from previous runs
4. **Consistency**: Follows pattern already established in `resource_cmdevice_device_idempotency_test.go`

## Files Modified

- `/workspace/internal/provider/resource_cmdevice_device_test.go`

## Verification

Test now compiles successfully:
```bash
export GOPATH=/workspace/.go && \
export GOCACHE=/workspace/.go/cache && \
export GOMODCACHE=/workspace/.go/pkg/mod && \
go test -c -o /tmp/test_binary ./internal/provider/
```

## Related Pattern

This fix follows the same pattern as the existing idempotency tests in:
- `/workspace/internal/provider/resource_cmdevice_device_idempotency_test.go`

Which already used `generateUniqueMAC()` successfully.

## Cleanup Enhancement

The existing `testAccCMDeviceDevicePreCheck()` function already handles device cleanup by name, so the unique MAC addresses work in conjunction with the existing cleanup infrastructure.

