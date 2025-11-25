# Feature Specification: BCM XML Schema for disksetup and raidconf Tests

## Overview

Fix the blocked acceptance tests for `bcm_cmdevice_category` resource by implementing proper XML validation using the discovered BCM XML schema format for `disksetup` and `raidconf` fields.

## Issue Reference

- **GitHub Issue**: [#48 - BCM XML Schema Required for disksetup and raidconf Optional Field Tests](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/48)
- **Priority**: Medium
- **Current Status**: 2 tests blocked by XML validation errors

## Problem Statement

Two acceptance tests for the `bcm_cmdevice_category` resource fail due to incorrect XML format:

1. `TestAccCMDeviceCategory_PartitionConfiguration` (lines 831-909)
2. `TestAccCMDeviceCategoryResource_DiskSetupAdvanced` (lines 1108-1313)

The tests were using incorrect XML structure that doesn't match BCM's XSD schema validation.

## Root Cause Analysis

### Discovery

Analysis of the BCM API response for the default category revealed the actual XML schema:

```xml
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <blockdev>/dev/hda</blockdev>
    <blockdev>/dev/vda</blockdev>
    <blockdev>/dev/xvda</blockdev>
    <blockdev>/dev/nvme0n1</blockdev>
    <blockdev mode="cloud">/dev/sdb</blockdev>
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
    <!-- Additional partitions... -->
  </device>
</diskSetup>
```

### Key Schema Differences from Failed Attempts

| Failed Attempt | Correct Schema |
|----------------|----------------|
| `<disksetup>` (lowercase) | `<diskSetup>` (camelCase) |
| `<disk device="...">` | `<device>` with `<blockdev>` children |
| `number="1"` attribute | `id="a0"` attribute |
| `type="ext4"` attribute | `<filesystem>ext4</filesystem>` child |
| `mountpoint="/"` attribute | `<mountPoint>/</mountPoint>` child |

### raidconf Field

The `raidconf` field in the default category is an empty string. Tests need to either:
1. Use empty string (valid)
2. Discover the correct XML format through API exploration
3. Skip raidconf testing until XSD documentation is available

## Solution Design

### Approach 1: Use Discovered Schema (Recommended)

Update tests to use the exact XML format discovered from the BCM API:

```hcl
resource "bcm_cmdevice_category" "test" {
  name               = "test-category"
  management_network = local.management_network_uuid

  disksetup = <<-EOT
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
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
  </device>
</diskSetup>
EOT

  raidconf = ""  # Empty string is valid

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
```

### Approach 2: Remove raidconf from Current Tests

If `raidconf` requires additional XSD research:
1. Keep `disksetup` tests with correct schema
2. Remove `raidconf` from tests temporarily
3. Create separate follow-up issue for raidconf XSD research

## Implementation Tasks

### Task 1: Update Test XML Format
- Update `testAccCMDeviceCategoryResourceConfig_PartitionConfig()`
- Update `testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated()`
- Update `testAccCMDeviceCategoryResourceConfig_DiskSetupAdvanced()`
- Use discovered XML schema format

### Task 2: Remove t.Skip() Statements
- Remove skip from `TestAccCMDeviceCategory_PartitionConfiguration`
- Remove skip from `TestAccCMDeviceCategoryResource_DiskSetupAdvanced`

### Task 3: Handle raidconf Field
- Determine if empty string is acceptable for tests
- If not, explore BCM API for valid raidconf examples
- Update tests accordingly

### Task 4: Run and Validate Tests
- Execute acceptance tests
- Verify all CRUD operations work
- Ensure idempotency checks pass

### Task 5: Documentation
- Update CLAUDE.md with discovered XML schema format
- Add helper function documentation for XML generation
- Update test coverage metrics

## Acceptance Criteria

1. **Tests Pass**: Both blocked tests pass against live BCM API
2. **Schema Documented**: XML schema format documented in CLAUDE.md
3. **Coverage Improved**: Optional field coverage increases from 55% toward 75%+
4. **No Regressions**: All existing tests continue to pass

## API Contract

### disksetup Field Format

```xml
<?xml version="1.0" encoding="UTF-8"?>
<diskSetup>
  <device>
    <blockdev>[device path]</blockdev>
    <blockdev mode="cloud">[cloud device path]</blockdev>
    <partition id="[unique-id]" [partitiontype="esp"]>
      <size>[size: 100M, 20G, max]</size>
      <type>[linux|linux swap]</type>
      <filesystem>[fat|xfs|ext4]</filesystem>
      <mountPoint>[mount path]</mountPoint>
      <mountOptions>[comma-separated options]</mountOptions>
    </partition>
  </device>
</diskSetup>
```

### Element Reference

| Element | Required | Attributes | Description |
|---------|----------|------------|-------------|
| `diskSetup` | Yes | None | Root element (camelCase) |
| `device` | Yes | None | Device container |
| `blockdev` | Yes (1+) | `mode` (optional) | Block device paths |
| `partition` | Yes (1+) | `id` (required), `partitiontype` (optional) | Partition definition |
| `size` | Yes | None | Size (e.g., "100M", "20G", "max") |
| `type` | Yes | None | Partition type (e.g., "linux", "linux swap") |
| `filesystem` | No | None | Filesystem type (e.g., "fat", "xfs", "ext4") |
| `mountPoint` | No | None | Mount path (e.g., "/", "/boot/efi") |
| `mountOptions` | No | None | Mount options |

### raidconf Field Format

Currently unknown. The default category uses an empty string. Options:
- Empty string `""` (known valid)
- XML format (requires XSD documentation)

## Test Matrix

| Test | Status | disksetup | raidconf | install_boot_record |
|------|--------|-----------|----------|---------------------|
| PartitionConfiguration | 🔴 Blocked | XML | "raid1" | Not set |
| DiskSetupAdvanced | 🔴 Blocked | XML | "raid1" | true |

After fix:

| Test | Status | disksetup | raidconf | install_boot_record |
|------|--------|-----------|----------|---------------------|
| PartitionConfiguration | 🟢 Pass | Valid XML | "" | Not set |
| DiskSetupAdvanced | 🟢 Pass | Valid XML | "" | true |

## Risk Assessment

| Risk | Impact | Mitigation |
|------|--------|------------|
| raidconf requires XML | Medium | Use empty string, create follow-up issue |
| BCM cluster unavailable | High | Tests already have PreCheck for connectivity |
| Schema varies by BCM version | Medium | Document tested BCM version |

## Dependencies

- BCM cluster at `https://172.21.15.254:8081`
- Existing test infrastructure and helpers
- terraform-plugin-testing v1.13.3+

## Timeline

This is a fix-only change with no new features:
- Schema discovery: Complete (found in API response)
- Test updates: ~30 minutes
- Validation: ~20 minutes
- Documentation: ~10 minutes

## References

- `/workspace/sampleRest/category_schema_documentation_20251121_070629.md` - Source of discovered schema
- `/workspace/internal/provider/resource_cmdevice_category_test.go` - Test file
- `/workspace/internal/provider/resource_cmdevice_category.go` - Resource implementation
