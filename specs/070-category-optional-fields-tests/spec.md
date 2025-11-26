# Feature Specification: CMDevice Category Optional Fields Test Coverage

**Feature Branch**: `070-category-optional-fields-tests`
**Created**: 2025-11-26
**Status**: Draft
**Input**: User description: "Add test coverage for 28 optional fields in cmdevice_category"
**GitHub Issue**: #70

## Overview

The `bcm_cmdevice_category` resource has 28 optional fields that currently lack comprehensive test coverage. This feature adds acceptance tests for these fields to verify they persist correctly through Create/Update/Import cycles using modern `statecheck.ExpectKnownValue()` patterns.

**Current State**:
- Overall test grade: A (98/100)
- Core CRUD, import, drift detection: All covered
- Optional field coverage: 71/99 fields tested (from gap analysis)
- These 28 untested fields are advanced BCM configuration options

## Optional Fields Inventory

### Already Tested Fields (Reference)

| Field | Type | Tested In |
|-------|------|-----------|
| `name` | string | Basic, Import, all tests |
| `management_network` | string (UUID) | Basic, Import, all tests |
| `notes` | string | Basic, DriftNotes |
| `kernel_parameters` | string | Basic |
| `software_image_proxy` | object | All tests |
| `force` | bool | ForceParameter |
| `boot_loader` | string (enum) | ValidationInvalidBootLoader |
| `boot_loader_file` | string | BootLoaderFields |
| `boot_loader_protocol` | string | BootLoaderFields |
| `kernel_output_console` | string | BootLoaderFields |
| `default_gateway` | string | NetworkConfiguration |
| `default_gateway_metric` | int64 | NetworkConfiguration |
| `allow_networking_restart` | bool | NetworkConfiguration |
| `disksetup` | string (XML) | PartitionConfiguration, DiskSetupAdvanced |
| `raidconf` | string | PartitionConfiguration |
| `install_boot_record` | bool | DiskSetupAdvanced |
| `fips` | string (enum) | ValidationInvalidFIPS |
| `initialize` | string | ProvisioningScripts |
| `finalize` | string | ProvisioningScripts |
| `install_mode` | string (enum) | InstallationModes |
| `new_node_install_mode` | string (enum) | InstallationModes |

### Untested Optional Fields (28 Fields)

#### Simple String Fields (7 fields)

| Field | Type | BCM API Field | Description |
|-------|------|---------------|-------------|
| `io_scheduler` | string | ioScheduler | I/O scheduler (noop, deadline, cfq) |
| `kernel_version` | string | kernelVersion | Kernel version string |
| `authentication_service` | string (enum) | authenticationService | AUTH_AUTO, LDAP, SSSD, LOCAL |
| `interactive_user` | string | interactiveUser | Interactive user mode (ALWAYS, NEVER) |
| `use_exclusively_for` | string | useExclusivelyFor | Exclusive resource allocation |
| `exclude_list_manipulate_script` | string | excludeListManipulateScript | Exclude list manipulation script |

#### Boolean Fields (3 fields)

| Field | Type | BCM API Field | Description |
|-------|------|---------------|-------------|
| `node_installer_disk` | bool | nodeInstallerDisk | Node installer disk flag |
| `version_config_files` | bool | versionConfigFiles | Version config files flag |
| `data_node` | bool | dataNode | Data node classification |

#### Large Text Fields - Exclude Lists (5 fields)

| Field | Type | BCM API Field | Max Size | Description |
|-------|------|---------------|----------|-------------|
| `exclude_list_full` | string | excludeListFull | 50KB | Exclude list for full operations |
| `exclude_list_grab` | string | excludeListGrab | 50KB | Exclude list for grab operations |
| `exclude_list_grabnew` | string | excludeListGrabnew | 50KB | Exclude list for grabnew operations |
| `exclude_list_sync` | string | excludeListSync | 50KB | Exclude list for sync operations |
| `exclude_list_update` | string | excludeListUpdate | 50KB | Exclude list for update operations |

#### List Fields (6 fields)

| Field | Type | Element Type | BCM API Field | Description |
|-------|------|--------------|---------------|-------------|
| `modules` | list | KernelModuleModel | modules | Kernel modules to load |
| `name_servers` | list | string | nameServers | DNS name servers |
| `search_domains` | list | string | searchDomains | DNS search domains |
| `time_servers` | list | string | timeServers | NTP time servers |
| `static_routes` | list | StaticRouteModel | staticRoutes | Static network routes |
| `gpu_settings` | list | GPUSettingModel | gpuSettings | GPU hardware configuration |

#### Nested Object Fields (5 fields)

| Field | Type | BCM API Field | Description |
|-------|------|---------------|-------------|
| `fsmounts` | list | FSMountModel | Filesystem mount configurations |
| `fsexports` | list | FSExportModel | NFS filesystem exports |
| `roles` | list | CategoryRoleModel | Service role assignments |
| `bmc_settings` | object | bmcSettings | BMC hardware management configuration |
| `services` | list | ServiceModel | Service configurations (TODO schema) |

#### Empty Schema Objects (Not Recommended for Testing)

These fields have empty schemas and are marked as POST-MVP:

- `bios_setup` - BIOS setup configuration (empty schema)
- `dpu_settings` - DPU settings (empty schema)
- `access_settings` - Access settings (empty schema)
- `selinux_settings` - SELinux settings (empty schema)
- `proxy_settings` - Proxy settings (empty schema)
- `timezone_settings` - Timezone settings (empty schema)
- `ztp_settings` - ZTP settings (empty schema)

## User Scenarios & Testing

### User Story 1 - Simple String and Boolean Field Tests (Priority: P1)

As a Terraform provider developer, I want to verify that simple string and boolean optional fields persist correctly through CRUD operations, so that users can configure category attributes reliably.

**Why this priority**: These are the most commonly used configuration fields and the simplest to implement tests for. They provide immediate value with low implementation effort.

**Independent Test**: Can be fully tested by creating a category with all simple fields set, verifying persistence, and testing update operations.

**Acceptance Scenarios**:

1. **Given** a new category configuration with `io_scheduler="noop"`, `authentication_service="LDAP"`, `node_installer_disk=true`, `version_config_files=true`, `data_node=true`, **When** the category is created and read back, **Then** all field values match the configuration.

2. **Given** an existing category, **When** fields are updated to `io_scheduler="deadline"`, `authentication_service="SSSD"`, `node_installer_disk=false`, **Then** the updated values persist correctly.

3. **Given** an existing category with optional fields set, **When** the category is imported, **Then** all field values are preserved after import.

4. **Given** a category configuration that has not changed, **When** Terraform plan is run, **Then** no changes are detected (idempotency).

---

### User Story 2 - Exclude List Field Tests (Priority: P2)

As a Terraform provider developer, I want to verify that large text exclude list fields handle content correctly, so that users can configure node provisioning exclusions.

**Why this priority**: Exclude lists are important for production deployments but less commonly used than simple fields. They have specific size constraints (50KB max) that need verification.

**Independent Test**: Can be fully tested by creating a category with exclude list content, verifying multi-line text preservation, and testing updates.

**Acceptance Scenarios**:

1. **Given** a new category with `exclude_list_full` containing multi-line patterns, **When** the category is created and read back, **Then** the exclude list content is preserved exactly including newlines.

2. **Given** an existing category with exclude lists, **When** lists are updated with new patterns, **Then** the updated content persists correctly.

3. **Given** a category with all five exclude list fields populated, **When** Terraform apply completes, **Then** all exclude list fields are idempotent on subsequent plans.

---

### User Story 3 - Network List Field Tests (Priority: P2)

As a Terraform provider developer, I want to verify that network-related list fields (name_servers, search_domains, time_servers) persist correctly, so that users can configure DNS and NTP settings.

**Why this priority**: Network configuration is essential for cluster operation. List fields require different testing patterns than simple strings.

**Independent Test**: Can be fully tested by creating a category with network lists, verifying list contents, and testing list modifications.

**Acceptance Scenarios**:

1. **Given** a new category with `name_servers=["8.8.8.8", "8.8.4.4"]`, `search_domains=["example.com", "local"]`, `time_servers=["pool.ntp.org"]`, **When** the category is created, **Then** all list values are preserved.

2. **Given** an existing category, **When** a new DNS server is added to the list, **Then** the list update persists correctly.

3. **Given** a category with network lists, **When** the category is imported, **Then** list contents match the original configuration.

---

### User Story 4 - Static Routes Field Tests (Priority: P3)

As a Terraform provider developer, I want to verify that static route nested objects persist correctly, so that users can configure custom routing for category nodes.

**Why this priority**: Static routes are an advanced networking feature with nested object structure requiring more complex validation.

**Independent Test**: Can be fully tested by creating a category with static routes, verifying nested object attributes, and testing route modifications.

**Acceptance Scenarios**:

1. **Given** a new category with `static_routes` containing destination, gateway, and metric, **When** the category is created, **Then** all static route attributes are preserved.

2. **Given** an existing category with static routes, **When** a route is added or modified, **Then** the changes persist correctly.

3. **Given** a category with static routes, **When** the category is imported, **Then** route definitions match the original configuration.

---

### User Story 5 - Kernel Modules Field Tests (Priority: P3)

As a Terraform provider developer, I want to verify that kernel modules list persists correctly, so that users can configure kernel module loading for category nodes.

**Why this priority**: Kernel modules are an advanced configuration option with nested object structure (name, parameters).

**Independent Test**: Can be fully tested by creating a category with kernel modules, verifying module attributes, and testing modifications.

**Acceptance Scenarios**:

1. **Given** a new category with `modules` containing module name and parameters, **When** the category is created, **Then** all module attributes are preserved.

2. **Given** an existing category with modules, **When** a module is added or parameters updated, **Then** the changes persist correctly.

---

### User Story 6 - BMC Settings Field Tests (Priority: P3)

As a Terraform provider developer, I want to verify that BMC settings nested object persists correctly, so that users can configure hardware management for category nodes.

**Why this priority**: BMC settings are hardware-specific and less commonly used. The nested object has sensitive fields (password) requiring special handling.

**Independent Test**: Can be fully tested by creating a category with BMC settings, verifying non-sensitive attributes, and testing updates.

**Acceptance Scenarios**:

1. **Given** a new category with `bmc_settings` containing user_name, privilege, and firmware_manage_mode, **When** the category is created, **Then** non-sensitive BMC settings are preserved.

2. **Given** an existing category with BMC settings, **When** settings are updated, **Then** the changes persist correctly.

---

### User Story 7 - Filesystem Configuration Tests (Priority: P4)

As a Terraform provider developer, I want to verify that filesystem mount and export configurations persist correctly, so that users can configure storage for category nodes.

**Why this priority**: Filesystem configurations (fsmounts, fsexports) are advanced features with complex nested object structures.

**Independent Test**: Can be fully tested by creating a category with filesystem configurations, verifying nested attributes, and testing modifications.

**Acceptance Scenarios**:

1. **Given** a new category with `fsmounts` containing device, mountpoint, filesystem, and mount options, **When** the category is created, **Then** all mount attributes are preserved.

2. **Given** a new category with `fsexports` containing path, network UUID, and export options, **When** the category is created, **Then** all export attributes are preserved.

---

### User Story 8 - GPU Settings and Roles Tests (Priority: P4)

As a Terraform provider developer, I want to verify that GPU settings and role assignments persist correctly, so that users can configure specialized hardware and service roles.

**Why this priority**: GPU settings and roles are specialized configurations used in HPC and AI workloads. Lower priority as they require specific cluster configurations.

**Independent Test**: Can be fully tested by creating a category with GPU settings and roles, verifying attributes, and testing modifications.

**Acceptance Scenarios**:

1. **Given** a new category with `gpu_settings` containing device_id, model, and compute_mode, **When** the category is created, **Then** GPU settings are preserved.

2. **Given** a new category with `roles` containing name, child_type, and add_services, **When** the category is created, **Then** role assignments are preserved.

---

### Edge Cases

- What happens when exclude list content approaches 50KB limit?
- How does system handle empty lists vs null lists for network configuration?
- What happens when static route destination/gateway are invalid CIDR/IP formats?
- How does import handle sensitive fields in BMC settings (password)?
- What happens when kernel module with invalid name is specified?

## Requirements

### Functional Requirements

- **FR-001**: Tests MUST verify field persistence through Create/Read cycle using `statecheck.ExpectKnownValue()`
- **FR-002**: Tests MUST verify field persistence through Update/Read cycle
- **FR-003**: Tests MUST verify idempotency using `plancheck.ExpectEmptyPlan()` after Create and Update
- **FR-004**: Tests MUST verify ImportState preserves field values (using `ImportStateVerify: true`)
- **FR-005**: Tests MUST use unique test names via `generateUniqueTestName()` to avoid conflicts
- **FR-006**: Tests MUST include proper cleanup using `testAccCMDeviceCategoryPreCheck()`
- **FR-007**: Tests MUST track ID consistency using `statecheck.CompareValue(compare.ValuesSame())`
- **FR-008**: Tests MUST use CheckDestroy to verify resource cleanup

### Test Configuration Requirements

- **FR-009**: All test configurations MUST use environment variables for credentials (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- **FR-010**: All test configurations MUST use data sources to lookup management_network and software_image UUIDs
- **FR-011**: All test configurations MUST set `insecure_skip_verify = true` for self-signed certificates

### Key Entities

- **CMDeviceCategoryResourceModel**: Resource data model containing all optional fields
- **KernelModuleCategoryModel**: Nested object for kernel module configuration (name, parameters)
- **StaticRouteModel**: Nested object for static routes (destination, gateway, metric)
- **BMCSettingsModel**: Nested object for BMC configuration (user_name, password, privilege, etc.)
- **FSMountModel**: Nested object for filesystem mounts (device, mountpoint, filesystem, etc.)
- **FSExportModel**: Nested object for NFS exports (path, network, allow_write, etc.)
- **CategoryRoleModel**: Nested object for role assignments (name, child_type, uuid, add_services)
- **GPUSettingModel**: Nested object for GPU configuration (device_id, model, compute_mode)

## Success Criteria

### Measurable Outcomes

- **SC-001**: All 28 untested optional fields have at least one acceptance test verifying Create/Read persistence
- **SC-002**: All tests pass with 100% success rate on the BCM test cluster
- **SC-003**: All tests are idempotent (empty plan on second apply)
- **SC-004**: Import tests verify field persistence for all testable fields
- **SC-005**: Test coverage for optional fields increases from 71/99 to 99/99 (100%)
- **SC-006**: No test failures due to hardcoded values or environment-specific assumptions
- **SC-007**: All tests complete within 120 minutes total timeout
- **SC-008**: Tests follow modern patterns from terraform-plugin-testing v1.13.3+

## Test Implementation Strategy

### Pattern Reference

All tests should follow the modern testing patterns documented in CLAUDE.md:

```go
// Required imports
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/compare"
)

// Modern state check pattern
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_cmdevice_category.test",
        tfjsonpath.New("field_name"),
        knownvalue.StringExact("expected-value"),
    ),
}

// Idempotency check pattern
ConfigPlanChecks: resource.ConfigPlanChecks{
    PreApply: []plancheck.PlanCheck{
        plancheck.ExpectEmptyPlan(),
    },
}

// ID consistency tracking pattern
compareID := statecheck.CompareValue(compare.ValuesSame())
// Add in each step: compareID.AddStateValue("resource.test", tfjsonpath.New("id"))
```

### Type-Specific Matchers

| Field Type | knownvalue Matcher |
|------------|-------------------|
| string | `knownvalue.StringExact("value")` |
| bool | `knownvalue.Bool(true)` |
| int64 | `knownvalue.Int64Exact(100)` |
| computed | `knownvalue.NotNull()` |
| list size | `knownvalue.ListSizeExact(2)` |

### Test Grouping Strategy

Tests should be grouped by field type to minimize test resource creation:

1. **TestAccCMDeviceCategoryResource_SimpleOptionalFields** - io_scheduler, authentication_service, interactive_user, use_exclusively_for
2. **TestAccCMDeviceCategoryResource_BooleanOptionalFields** - node_installer_disk, version_config_files, data_node
3. **TestAccCMDeviceCategoryResource_ExcludeLists** - All exclude_list_* fields
4. **TestAccCMDeviceCategoryResource_NetworkLists** - name_servers, search_domains, time_servers
5. **TestAccCMDeviceCategoryResource_StaticRoutes** - static_routes nested object
6. **TestAccCMDeviceCategoryResource_KernelModules** - modules nested object
7. **TestAccCMDeviceCategoryResource_BMCSettings** - bmc_settings nested object
8. **TestAccCMDeviceCategoryResource_FilesystemMounts** - fsmounts nested object
9. **TestAccCMDeviceCategoryResource_FilesystemExports** - fsexports nested object
10. **TestAccCMDeviceCategoryResource_GPUSettings** - gpu_settings nested object
11. **TestAccCMDeviceCategoryResource_Roles** - roles nested object

## Dependencies and Assumptions

### Dependencies

- BCM test cluster must be accessible at BCM_ENDPOINT
- Valid BCM credentials (BCM_USERNAME, BCM_PASSWORD)
- At least one existing category with management_network for UUID lookup
- At least one existing software image for parent_software_image reference

### Assumptions

- Optional fields with empty/TODO schemas (bios_setup, dpu_settings, etc.) are excluded from testing until schemas are defined
- The `services` field is excluded until its schema is properly documented (marked POST-MVP)
- BCM API accepts the field values specified in the resource schema
- Sensitive fields (BMC password) can be verified for non-null state but exact values cannot be imported

## Out of Scope

- Testing empty schema objects (bios_setup, dpu_settings, access_settings, selinux_settings, proxy_settings, timezone_settings, ztp_settings)
- Testing services field (schema is TODO)
- Drift detection tests for optional fields (covered by existing DriftNotes test pattern)
- Performance testing for large exclude list content
- Validation error testing for invalid field values (except where schema validators exist)
