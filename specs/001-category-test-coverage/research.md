# Research: Category Test Coverage Enhancement

**Feature Branch**: `001-category-test-coverage`
**Date**: 2025-11-25
**Status**: Complete

## Research Summary

This document captures technical research findings for implementing comprehensive test coverage for `bcm_cmdevice_category` optional fields.

---

## 1. Current Test Coverage Analysis

### Existing Tests in `resource_cmdevice_category_test.go`

| Test Function | Fields Covered | Test Type |
|---------------|----------------|-----------|
| `TestAccCMDeviceCategoryResource_Basic` | name, uuid, management_network, notes, kernel_parameters, base_type | CRUD + Idempotency |
| `TestAccCMDeviceCategoryResource_Import` | name, notes, kernel_parameters, uuid, id | Import |
| `TestAccCMDeviceCategoryResource_ForceParameter` | name, force, id | CRUD + Idempotency |
| `TestAccCMDeviceCategory_DriftNotes` | name, notes, uuid | Drift Detection |
| `TestAccCMDeviceCategory_DestroyWithForce` | name, force, uuid | Destroy |
| `TestAccCMDeviceCategory_DestroyExternalDelete` | name, uuid | Destroy Edge Case |
| `TestAccCMDeviceCategory_NetworkConfiguration` | default_gateway, default_gateway_metric, allow_networking_restart | CRUD + Idempotency |
| `TestAccCMDeviceCategory_PartitionConfiguration` | disksetup | CRUD + Idempotency |
| `TestAccCMDeviceCategoryResource_DiskSetupAdvanced` | disksetup, install_boot_record, software_image_proxy | CRUD + Import |
| `TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations` | disksetup only | Combination |
| `TestAccCMDeviceCategory_ValidationInvalidName` | name | Validation |
| `TestAccCMDeviceCategory_ValidationInvalidManagementNetwork` | management_network | Validation |
| `TestAccCMDeviceCategory_ValidationInvalidBootLoader` | boot_loader | Validation |
| `TestAccCMDeviceCategory_ValidationInvalidFIPS` | fips | Validation |
| `TestAccCMDeviceCategoryResource_BootLoaderFields` | boot_loader_file, boot_loader_protocol, kernel_output_console | CRUD + Import |

### Fields Already Tested (~18 fields)

- name, uuid, id, management_network (Required/Computed)
- notes, kernel_parameters, force (Optional)
- default_gateway, default_gateway_metric, allow_networking_restart (Network)
- disksetup, install_boot_record (Disk)
- boot_loader_file, boot_loader_protocol, kernel_output_console (Boot)
- software_image_proxy (Nested Object)
- base_type (Computed metadata)

### Fields NOT Yet Tested (~28 fields from spec)

**Priority 1 - Installation Mode Fields:**
- install_mode
- new_node_install_mode

**Priority 1 - Network Settings Lists:**
- name_servers (list of strings)
- search_domains (list of strings)
- time_servers (list of strings)

**Priority 2 - I/O and Kernel Fields:**
- io_scheduler
- kernel_version

**Priority 2 - Provisioning Scripts:**
- initialize
- finalize

**Priority 3 - BMC Settings (Nested Object):**
- bmc_settings.user_name
- bmc_settings.privilege
- bmc_settings.user_id
- bmc_settings.firmware_manage_mode
- bmc_settings.leak_policy
- bmc_settings.leak_reaction_delay
- bmc_settings.power_reset_delay

**Priority 3 - Kernel Modules (Nested List):**
- modules[].name
- modules[].parameters

**Priority 3 - Exclude Lists:**
- exclude_list_full
- exclude_list_grab
- exclude_list_sync
- exclude_list_update
- exclude_list_grabnew
- exclude_list_manipulate_script

**Priority 3 - Miscellaneous:**
- fips (valid values, CRUD not just validation)
- data_node
- interactive_user
- use_exclusively_for
- node_installer_disk
- version_config_files
- authentication_service
- raidconf

---

## 2. Resource Implementation Analysis

### buildAPIEntity Support

From `resource_cmdevice_category.go` lines 1445-1559:

**Currently Mapped to API:**
- name, managementNetwork, notes (Required/Optional)
- bootLoader, bootLoaderFile, bootLoaderProtocol (Boot)
- kernelVersion, kernelParameters, kernelOutputConsole (Kernel)
- installMode, newNodeInstallMode, installBootRecord (Installation)
- defaultGateway, defaultGatewayMetric (Network)
- disksetup, raidconf (Disk)
- softwareImageProxy (Nested Object)

**NOT Mapped in buildAPIEntity (Missing Implementation):**
- ioScheduler
- nameServers, searchDomains, timeServers (lists)
- initialize, finalize (scripts)
- modules (nested list)
- bmcSettings (nested object)
- fips, dataNode, interactiveUser, useExclusivelyFor
- nodeInstallerDisk, versionConfigFiles, authenticationService
- All exclude list fields

### readCategory Support

From `resource_cmdevice_category.go` lines 1619-1755:

**Currently Read from API:**
- All identity and required fields
- Boot, kernel, installation, network config fields
- disksetup, raidconf
- softwareImageProxy (parsed as nested object)

**Set to Null (Not Parsed):**
- nameServers, searchDomains, timeServers (lines 1648-1652)
- modules (line 1679)
- bmcSettings (line 1704)
- All fsmounts, fsexports, roles, services, gpuSettings (dynamic types)

---

## 3. BCM API Field Support Research

### Decision: Test Only Fields with Implementation Support

**Rationale:**
1. Testing fields not implemented in buildAPIEntity/readCategory will fail
2. The spec focuses on test coverage, not resource implementation
3. We should prioritize testing fields that ARE implemented

### Fields to Test (Implementation Exists)

| Field | Type | buildAPIEntity | readCategory | Test Priority |
|-------|------|----------------|--------------|---------------|
| install_mode | String | Yes | Yes | P1 |
| new_node_install_mode | String | Yes | Yes | P1 |
| default_gateway | String | Yes | Yes | Already Tested |
| default_gateway_metric | Int64 | Yes | Yes | Already Tested |
| allow_networking_restart | Bool | No | No | Skip (not in buildAPIEntity) |
| kernel_version | String | Yes | Yes | P2 |
| kernel_parameters | String | Yes | Yes | Already Tested |
| kernel_output_console | String | Yes | Yes | Already Tested |
| disksetup | String | Yes | Yes | Already Tested |
| raidconf | String | Yes | Yes | P3 (needs valid XML) |
| install_boot_record | Bool | Yes | Yes | Already Tested |

### Fields Requiring Implementation First (Out of Scope)

These fields are in the schema but not implemented in CRUD operations:

- name_servers, search_domains, time_servers (lists - set to null in readCategory)
- io_scheduler (not in buildAPIEntity)
- initialize, finalize (not in buildAPIEntity)
- modules (not in buildAPIEntity, set to null in readCategory)
- bmc_settings (not in buildAPIEntity, set to null in readCategory)
- All exclude_list_* fields (not in buildAPIEntity)
- fips, data_node, interactive_user, use_exclusively_for (not in buildAPIEntity)
- node_installer_disk, version_config_files, authentication_service (not in buildAPIEntity)

---

## 4. Test Strategy Decision

### Decision: Focus on Fields with Existing Implementation

Given the analysis, we will:

1. **Create tests for fields already implemented** in buildAPIEntity/readCategory
2. **Document coverage gaps** for fields needing implementation work
3. **Follow TDD** - tests should pass with current implementation

### Test Grouping Strategy

**Group 1: Installation Modes (P1)**
- Test Function: `TestAccCMDeviceCategoryResource_InstallationModes`
- Fields: install_mode, new_node_install_mode
- Pattern: Create with values, idempotency, update, idempotency, import

**Group 2: Boot Configuration (Enhancement)**
- Test Function: `TestAccCMDeviceCategoryResource_BootConfiguration` (enhance existing)
- Fields: boot_loader, boot_loader_file, boot_loader_protocol
- Note: Already has `TestAccCMDeviceCategoryResource_BootLoaderFields`

**Group 3: Kernel Configuration (P2)**
- Test Function: `TestAccCMDeviceCategoryResource_KernelConfiguration`
- Fields: kernel_version, kernel_parameters, kernel_output_console
- Note: kernel_parameters already tested in basic, enhance coverage

**Group 4: Disk Configuration (Enhancement)**
- Test Function: Enhance `TestAccCMDeviceCategoryResource_DiskSetupAdvanced`
- Fields: disksetup, raidconf, install_boot_record
- Note: raidconf needs valid BCM XML format

---

## 5. Modern Testing Patterns Research

### Required Imports (terraform-plugin-testing v1.13.3+)

```go
import (
    "github.com/hashicorp/terraform-plugin-testing/compare"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)
```

### State Check Patterns

**String Field:**
```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("install_mode"),
    knownvalue.StringExact("AUTO"),
)
```

**Boolean Field:**
```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("install_boot_record"),
    knownvalue.Bool(true),
)
```

**Computed Field (exists but value varies):**
```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("uuid"),
    knownvalue.NotNull(),
)
```

### Idempotency Verification

```go
{
    Config: testAccConfig(name, "value"),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

### ID Consistency Tracking

```go
compareID := statecheck.CompareValue(compare.ValuesSame())

// In each step:
ConfigStateChecks: []statecheck.StateCheck{
    compareID.AddStateValue(
        "bcm_cmdevice_category.test",
        tfjsonpath.New("id"),
    ),
},
```

---

## 6. Implementation Recommendations

### Recommendation 1: Add Tests for install_mode and new_node_install_mode

These fields are fully implemented and should be tested immediately.

### Recommendation 2: Enhance kernel_version Test

kernel_version is implemented but not tested. However, it requires a valid kernel path from the actual software image, which may be environment-dependent.

### Recommendation 3: Document Implementation Gaps

Create a tracking issue for fields needing implementation before they can be tested:
- Network lists (name_servers, search_domains, time_servers)
- Script fields (initialize, finalize)
- BMC settings (nested object)
- Kernel modules (nested list)
- Exclude lists
- Boolean/string misc fields

### Recommendation 4: Use Environment-Portable Tests

All tests should:
- Generate unique names with `generateUniqueTestName()`
- Look up required UUIDs from existing categories/images
- Not assume specific values exist in the BCM cluster

---

## 7. Alternatives Considered

### Alternative A: Test All Schema Fields Regardless of Implementation

**Rejected because:**
- Tests would fail due to missing CRUD implementation
- Would require implementing all field mappings (out of scope)
- Spec explicitly focuses on test coverage, not implementation

### Alternative B: Create Stub Implementations for Missing Fields

**Rejected because:**
- Changes resource implementation (out of spec scope)
- Risk of introducing bugs
- Spec focuses on acceptance tests only

### Alternative C: Focus Only on Fields Explicitly Listed in Spec

**Adopted:**
- Tests fields mentioned in spec user stories
- Prioritizes fields with existing implementation
- Documents gaps for future work

---

## 8. Conclusion

The test coverage enhancement will focus on:

1. **Installation Modes** (install_mode, new_node_install_mode) - New test
2. **Kernel Configuration** (kernel_version enhancement) - Requires investigation
3. **Documentation** of implementation gaps for ~20 fields

The remaining fields from the spec (network lists, scripts, BMC, modules, exclude lists, misc fields) require implementation work in buildAPIEntity/readCategory before tests can be written.
