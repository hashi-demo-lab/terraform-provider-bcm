# Implementation Plan: Fix disksetup/raidconf XML Schema Tests

## Summary

Update the blocked `bcm_cmdevice_category` acceptance tests to use the correct BCM XML schema format discovered from API analysis.

## Current State

### Blocked Tests
1. `TestAccCMDeviceCategory_PartitionConfiguration` - Lines 831-915
2. `TestAccCMDeviceCategoryResource_DiskSetupAdvanced` - Lines 1108-1313

### Error Messages
```
disksetup/xml: The disk setup contains invalid XML
disksetup/xsd: Element 'disksetup': No matching global declaration available for the validation root.
```

## Implementation Strategy

### Phase 1: Update Test Configuration Functions

#### 1.1 Create Valid disksetup XML Generator

Create a helper function that generates valid BCM disksetup XML:

```go
// generateValidDiskSetupXML creates a valid BCM disksetup XML string
func generateValidDiskSetupXML(partitions []testPartition) string {
    // Use discovered schema format
}

type testPartition struct {
    id           string
    partitionType string // "esp" for EFI partitions, empty otherwise
    size         string // "100M", "20G", "max"
    fsType       string // "linux", "linux swap"
    filesystem   string // "fat", "xfs", "ext4"
    mountPoint   string
    mountOptions string
}
```

#### 1.2 Update testAccCMDeviceCategoryResourceConfig_PartitionConfig

**Before (incorrect):**
```hcl
disksetup = <<-EOT
<disksetup>
  <disk device="/dev/sda">
    <partition number="1" size="50GB" type="ext4" mountpoint="/"/>
  </disk>
</disksetup>
EOT
```

**After (correct):**
```hcl
disksetup = <<-EOT
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <blockdev>/dev/vda</blockdev>
    <partition id="a0" partitiontype="esp">
      <size>100M</size>
      <type>linux</type>
      <filesystem>fat</filesystem>
      <mountPoint>/boot/efi</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
    <partition id="a1">
      <size>20G</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
    <partition id="a2">
      <size>max</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/local</mountPoint>
      <mountOptions>defaults,noatime,nodiratime</mountOptions>
    </partition>
  </device>
</diskSetup>
EOT
```

#### 1.3 Handle raidconf Field

Based on BCM API analysis, the default category uses an empty string for `raidconf`. Update tests to:

1. Use empty string `raidconf = ""` for initial tests
2. Create separate follow-up issue if RAID XML format needed

### Phase 2: Remove Test Blocks

#### 2.1 Remove t.Skip() Statements

```go
// BEFORE:
func TestAccCMDeviceCategory_PartitionConfiguration(t *testing.T) {
    t.Skip("BLOCKED: Requires BCM XSD files for disksetup/raidconf validation. See issue #48")
    // ...
}

// AFTER:
func TestAccCMDeviceCategory_PartitionConfiguration(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-partition-config")
    // ...
}
```

### Phase 3: Test Execution Matrix

| Test Case | Create | Read | Update | Import | Idempotency |
|-----------|--------|------|--------|--------|-------------|
| PartitionConfiguration | ✓ | ✓ | ✓ | ✓ | ✓ |
| DiskSetupAdvanced | ✓ | ✓ | ✓ | ✓ | ✓ |

### Phase 4: Documentation Updates

#### 4.1 CLAUDE.md Updates

Add to "BCM-Specific Notes" section:

```markdown
### disksetup XML Schema

BCM validates `disksetup` against an internal XSD schema. The correct format is:

- Root element: `<diskSetup>` (camelCase, not lowercase)
- Device wrapper: `<device>` containing `<blockdev>` and `<partition>` elements
- Partition attributes: `id` (required), `partitiontype` (optional, "esp" for EFI)
- Partition children: `<size>`, `<type>`, `<filesystem>`, `<mountPoint>`, `<mountOptions>`

Example:
```xml
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <partition id="a0" partitiontype="esp">
      <size>100M</size>
      <type>linux</type>
      <filesystem>fat</filesystem>
      <mountPoint>/boot/efi</mountPoint>
    </partition>
  </device>
</diskSetup>
```
```

## File Changes

### Modified Files

| File | Changes |
|------|---------|
| `internal/provider/resource_cmdevice_category_test.go` | Update XML format, remove t.Skip() |
| `CLAUDE.md` | Add disksetup XML schema documentation |

### New Files

None required - this is a test fix only.

## Validation Steps

1. **Unit Compilation**: `go build ./...`
2. **Single Test Run**:
   ```bash
   TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run "TestAccCMDeviceCategory_PartitionConfiguration"
   ```
3. **Both Tests**:
   ```bash
   TF_ACC=1 go test -v -timeout 60m ./internal/provider/ -run "TestAccCMDeviceCategory_Partition|TestAccCMDeviceCategoryResource_DiskSetup"
   ```
4. **Full Test Suite** (optional):
   ```bash
   TF_ACC=1 go test -v -timeout 120m ./internal/provider/
   ```

## Rollback Plan

If tests still fail:
1. Restore t.Skip() statements
2. Create detailed follow-up issue with new error messages
3. Consider reaching out to BCM team for XSD files

## Success Metrics

| Metric | Before | After |
|--------|--------|-------|
| Blocked tests | 2 | 0 |
| Optional field coverage | 55% | ~65%+ |
| Test compile errors | 0 | 0 |

## Timeline Estimate

| Phase | Duration |
|-------|----------|
| XML format updates | 15 min |
| Remove t.Skip() | 5 min |
| Test execution | 20-30 min |
| Documentation | 10 min |
| **Total** | ~60 min |

## Dependencies

- BCM cluster availability at `https://172.21.15.254:8081`
- Valid BCM credentials in environment variables
- Go 1.24+ and terraform-plugin-testing v1.13.3+
