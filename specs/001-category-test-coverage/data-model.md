# Data Model: Category Test Coverage Enhancement

**Feature Branch**: `001-category-test-coverage`
**Date**: 2025-11-25

## Test Entities Overview

This document defines the test entity models and their relationships for acceptance tests.

---

## 1. Test Function Entity

Each test function represents a logical grouping of related field tests.

### Schema

```
TestFunction {
    name: String              // Go function name (e.g., "TestAccCMDeviceCategoryResource_InstallationModes")
    priority: P1|P2|P3        // Priority level from spec
    fields_covered: []String  // List of optional fields tested
    test_steps: []TestStep    // Ordered list of test steps
    dependencies: []String    // Required data sources/resources
}
```

### Test Functions to Create

| Name | Priority | Fields Covered | User Story |
|------|----------|----------------|------------|
| `TestAccCMDeviceCategoryResource_InstallationModes` | P1 | install_mode, new_node_install_mode | US1 |
| `TestAccCMDeviceCategoryResource_NetworkListFields` | P1 | name_servers, search_domains, time_servers | US2 |
| `TestAccCMDeviceCategoryResource_IOSchedulerKernel` | P2 | io_scheduler, kernel_version | US3 |
| `TestAccCMDeviceCategoryResource_ProvisioningScripts` | P2 | initialize, finalize | US4 |
| `TestAccCMDeviceCategoryResource_BMCSettings` | P3 | bmc_settings.* | US5 |
| `TestAccCMDeviceCategoryResource_KernelModules` | P3 | modules[].* | US6 |
| `TestAccCMDeviceCategoryResource_ExcludeLists` | P3 | exclude_list_* | US7 |
| `TestAccCMDeviceCategoryResource_MiscellaneousFields` | P3 | fips, data_node, interactive_user, etc. | US8 |

---

## 2. Test Step Entity

Each step within a test function follows this pattern.

### Schema

```
TestStep {
    step_number: Int           // 1-based index
    step_type: String          // "create", "idempotency", "update", "import", "drift"
    config_function: String    // Name of config helper function
    config_params: Map         // Parameters passed to config function
    state_checks: []StateCheck // Verifications to perform
    plan_checks: []PlanCheck   // Plan verifications (for idempotency)
    pre_config: Function       // Optional pre-step hook (for drift)
}
```

### Standard Step Sequence

For each test function, use this step sequence:

1. **Create** - Create resource with initial values
2. **Idempotency (Create)** - Verify no changes on re-apply
3. **Update** - Modify field values
4. **Idempotency (Update)** - Verify no changes on re-apply
5. **Import** - Verify import preserves values (optional)

---

## 3. Config Helper Entity

Each test requires a config helper function.

### Schema

```
ConfigHelper {
    name: String              // Function name (e.g., "testAccCMDeviceCategoryResourceConfig_InstallationModes")
    parameters: []Param       // Function parameters
    returns: String           // Always returns string (HCL config)
    base_config: String       // Provider + data source boilerplate
    resource_config: String   // Resource-specific configuration
}
```

### Base Config Template

All config helpers should use this base template:

```go
func testAccCMDeviceCategoryResourceConfig_FEATURE(name string, params...) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %%[1]q
  username             = %%[2]q
  password             = %%[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %%[4]q
  management_network = local.management_network_uuid

  # FEATURE-SPECIFIC FIELDS HERE

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        // Additional params...
    )
}
```

---

## 4. Field Coverage Matrix

### Priority 1 Fields (User Stories 1-2)

| Field | Type | Default | Valid Values | Test Config |
|-------|------|---------|--------------|-------------|
| install_mode | String | "AUTO" | AUTO, FULL, MINIMAL, CUSTOM | `install_mode = "AUTO"` |
| new_node_install_mode | String | "FULL" | FULL, MINIMAL, SKIP | `new_node_install_mode = "FULL"` |
| name_servers | List(String) | null | List of IP addresses | `name_servers = ["8.8.8.8", "8.8.4.4"]` |
| search_domains | List(String) | null | List of domains | `search_domains = ["example.com"]` |
| time_servers | List(String) | null | List of NTP servers | `time_servers = ["ntp.example.com"]` |

### Priority 2 Fields (User Stories 3-4)

| Field | Type | Default | Valid Values | Test Config |
|-------|------|---------|--------------|-------------|
| io_scheduler | String | null | mq-deadline, none, bfq, kyber | `io_scheduler = "mq-deadline"` |
| kernel_version | String | null | Kernel path string | `kernel_version = "/boot/vmlinuz"` |
| initialize | String | null | Shell script | `initialize = "#!/bin/bash\necho 'init'"` |
| finalize | String | null | Shell script | `finalize = "#!/bin/bash\necho 'done'"` |

### Priority 3 Fields (User Stories 5-8)

| Field | Type | Default | Valid Values | Test Config |
|-------|------|---------|--------------|-------------|
| bmc_settings.user_name | String | null | Username | `bmc_settings = { user_name = "admin" }` |
| bmc_settings.privilege | String | null | USER, OPERATOR, ADMINISTRATOR | `bmc_settings = { privilege = "ADMINISTRATOR" }` |
| bmc_settings.user_id | Int64 | null | Integer | `bmc_settings = { user_id = 2 }` |
| modules[].name | String | Required | Module name | `modules = [{ name = "mlx5_core" }]` |
| modules[].parameters | String | null | Parameters | `modules = [{ name = "mlx5_core", parameters = "debug=1" }]` |
| exclude_list_full | String | null | Multiline patterns | `exclude_list_full = "/var/log/*\n/tmp/*"` |
| exclude_list_grab | String | null | Multiline patterns | Similar |
| exclude_list_sync | String | null | Multiline patterns | Similar |
| exclude_list_update | String | null | Multiline patterns | Similar |
| exclude_list_grabnew | String | null | Multiline patterns | Similar |
| exclude_list_manipulate_script | String | null | Script path | `exclude_list_manipulate_script = "/opt/script.sh"` |
| fips | String | null | "YES", "NO" | `fips = "YES"` |
| data_node | Bool | null | true, false | `data_node = true` |
| interactive_user | String | null | Username | `interactive_user = "testuser"` |
| use_exclusively_for | String | null | Purpose string | `use_exclusively_for = "compute"` |
| node_installer_disk | Bool | null | true, false | `node_installer_disk = true` |
| version_config_files | Bool | null | true, false | `version_config_files = true` |
| authentication_service | String | null | AUTO, LDAP, SSSD, LOCAL | `authentication_service = "AUTO"` |

---

## 5. State Check Patterns by Type

### String Field State Check

```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("install_mode"),
    knownvalue.StringExact("AUTO"),
)
```

### Boolean Field State Check

```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("data_node"),
    knownvalue.Bool(true),
)
```

### Integer Field State Check

```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("bmc_settings").AtMapKey("user_id"),
    knownvalue.Int64Exact(2),
)
```

### List Field State Check (Size)

```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("name_servers"),
    knownvalue.ListSizeExact(2),
)
```

### Nested Object Field State Check

```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("bmc_settings").AtMapKey("user_name"),
    knownvalue.StringExact("admin"),
)
```

### Nested List Element State Check

```go
statecheck.ExpectKnownValue(
    "bcm_cmdevice_category.test",
    tfjsonpath.New("modules").AtSliceIndex(0).AtMapKey("name"),
    knownvalue.StringExact("mlx5_core"),
)
```

---

## 6. Implementation Status

### Implemented in buildAPIEntity (Can Test Now)

- install_mode, new_node_install_mode
- kernel_version, kernel_parameters, kernel_output_console
- boot_loader, boot_loader_file, boot_loader_protocol
- default_gateway, default_gateway_metric
- disksetup, raidconf, install_boot_record

### NOT Implemented in buildAPIEntity (Requires Implementation First)

- name_servers, search_domains, time_servers (lists)
- io_scheduler
- initialize, finalize (scripts)
- modules (nested list)
- bmc_settings (nested object)
- All exclude_list_* fields
- fips, data_node, interactive_user, use_exclusively_for
- node_installer_disk, version_config_files, authentication_service

---

## 7. Test Data Requirements

### Unique Names

All tests must use unique names to avoid conflicts:

```go
categoryName := generateUniqueTestName("tftest-FEATURE")
```

### Required Data Sources

All tests require these data sources for UUID lookups:
- `bcm_cmdevice_categories` - For management_network UUID
- `bcm_cmpart_softwareimages` - For software_image UUID

### Cleanup

Tests should clean up leftover resources:

```go
testAccCMDeviceCategoryPreCheck(t, categoryName)
```
