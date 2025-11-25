# TDD Compliance Analysis: disksetup/raidconf XML Schema Fix

## Analysis Summary

| Criteria | Status | Notes |
|----------|--------|-------|
| TDD Compliance | N/A | Test fix (not new feature) |
| Test Coverage | Improving | 2 blocked tests → 2 passing |
| Modern Patterns | Yes | Uses terraform-plugin-testing v1.13.3+ |
| API Contract | Discovered | XML schema documented from API |

## Current Test State

### Blocked Tests Analysis

#### Test 1: TestAccCMDeviceCategory_PartitionConfiguration
- **Location:** Lines 831-915
- **Status:** BLOCKED (t.Skip)
- **Issue:** Invalid XML format for `disksetup` and `raidconf`
- **Modern Patterns:** ✅ Uses statecheck.ExpectKnownValue, plancheck.ExpectEmptyPlan

#### Test 2: TestAccCMDeviceCategoryResource_DiskSetupAdvanced
- **Location:** Lines 1105-1260
- **Status:** BLOCKED (t.Skip)
- **Issue:** Invalid XML format
- **Modern Patterns:** ✅ Uses compareID for ID consistency tracking

### Working Test Reference

#### TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations
- **Location:** Lines 1264-1333
- **Status:** Uses CORRECT XML format (already fixed!)
- **Evidence:** Line 1282-1302 shows proper `<diskSetup>` XML structure

```go
// Already correct in codebase:
Config: testAccCMDeviceCategoryResourceConfig_DiskSetupOnly(categoryName, `<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <partition id="a0" partitiontype="esp">
      ...
```

## Root Cause Analysis

### XML Format Errors

| Element | Incorrect (Current) | Correct (Required) |
|---------|--------------------|--------------------|
| Root | `<disksetup>` | `<diskSetup>` |
| Device | `<disk device="...">` | `<device><blockdev>...</blockdev>` |
| Partition ID | `number="1"` | `id="a0"` |
| Size | attribute | `<size>` child element |
| Type | attribute | `<type>` child element |
| Filesystem | attribute | `<filesystem>` child element |
| Mount Point | attribute | `<mountPoint>` child element |

### raidconf Field

- **Current:** `raidconf = "raid1"` or `raidconf = "raid5"`
- **Required:** `raidconf = ""` (empty string) OR valid XML (unknown)
- **Solution:** Use empty string, defer RAID XML investigation

## Test Coverage Impact

### Before Fix
- 2 tests blocked by XML validation
- Optional field coverage: ~55%
- Fields untested: `disksetup`, `raidconf`, `install_boot_record`

### After Fix
- 2 tests passing
- Optional field coverage: ~65%+
- Fields tested: `disksetup` (with valid XML), `install_boot_record`
- Still deferred: `raidconf` XML format (empty string only)

## TDD Compliance Checklist

| Item | Status | Evidence |
|------|--------|----------|
| Tests written before implementation | N/A | Tests exist, format wrong |
| Tests use modern assertions | ✅ | statecheck, plancheck, compare |
| Tests cover CRUD operations | ✅ | Create, Update, Idempotency, Import |
| Tests verify idempotency | ✅ | plancheck.ExpectEmptyPlan() |
| Tests track ID consistency | ✅ | compareID.AddStateValue() |
| Tests have CheckDestroy | ✅ | testAccCheckCMDeviceCategoryDestroy |
| Tests use unique names | ✅ | generateUniqueTestName() |

## Implementation Recommendations

### Priority 1: Fix testAccCMDeviceCategoryResourceConfig_PartitionConfig
- Replace incorrect XML (lines 1025-1032)
- Change `raidconf = "raid1"` to `raidconf = ""`
- Use format from working test (DiskSetupOptionalCombinations)

### Priority 2: Fix testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated
- Replace incorrect XML (lines 1071-1080)
- Change `raidconf = "raid5"` to `raidconf = ""`
- Use different partition IDs for update (b0, b1, etc.)

### Priority 3: Fix TestAccCMDeviceCategoryResource_DiskSetupAdvanced
- Replace `diskSetupXML` variable (lines 1116-1121)
- Replace `updatedDiskSetupXML` variable (lines 1123-1128)
- Update assertions to expect empty raidconf

### Priority 4: Remove t.Skip() Statements
- Line 837: TestAccCMDeviceCategory_PartitionConfiguration
- Line 1106: TestAccCMDeviceCategoryResource_DiskSetupAdvanced

### Priority 5: Update Test Assertions
- Remove `knownvalue.StringExact("raid1")` checks
- Remove `knownvalue.StringExact("raid5")` checks
- Keep `knownvalue.NotNull()` for optional fields

## Risk Assessment

| Risk | Impact | Likelihood | Mitigation |
|------|--------|------------|------------|
| XML still invalid | Medium | Low | Working example exists in codebase |
| raidconf needs XML | Low | Medium | Use empty string, create follow-up |
| BCM cluster unavailable | High | Low | Test only when cluster available |

## Files to Modify

1. **internal/provider/resource_cmdevice_category_test.go**
   - `testAccCMDeviceCategoryResourceConfig_PartitionConfig()` - lines 1002-1045
   - `testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated()` - lines 1048-1093
   - `TestAccCMDeviceCategory_PartitionConfiguration()` - remove t.Skip at line 837
   - `TestAccCMDeviceCategoryResource_DiskSetupAdvanced()` - fix XML and remove t.Skip

2. **CLAUDE.md**
   - Add disksetup XML schema documentation section

## Validation Commands

```bash
# Build verification
go build ./internal/provider/

# Single test
TF_ACC=1 go test -v -timeout 30m ./internal/provider/ \
  -run "TestAccCMDeviceCategory_PartitionConfiguration"

# Both blocked tests
TF_ACC=1 go test -v -timeout 60m ./internal/provider/ \
  -run "TestAccCMDeviceCategory_Partition|TestAccCMDeviceCategoryResource_DiskSetupAdvanced"

# Full category suite
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMDeviceCategory"
```

## Conclusion

The fix is straightforward:
1. Copy working XML format from `TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations`
2. Apply to blocked test helper functions
3. Set `raidconf = ""` (empty string)
4. Remove `t.Skip()` statements
5. Update assertions

**Confidence Level:** HIGH - Working example already exists in the same test file.
