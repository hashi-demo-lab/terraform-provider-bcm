# Implementation Plan: CMDevice Category Optional Fields Test Coverage

**Branch**: `070-category-optional-fields-tests` | **Date**: 2025-11-26 | **Spec**: `/workspace/specs/070-category-optional-fields-tests/spec.md`
**Input**: Feature specification from `/specs/070-category-optional-fields-tests/spec.md`
**GitHub Issue**: #70

## Summary

Add comprehensive acceptance test coverage for 28 untested optional fields in the `bcm_cmdevice_category` resource. The implementation follows TDD principles with modern `terraform-plugin-testing` v1.13.3+ patterns including `statecheck.ExpectKnownValue()`, `plancheck.ExpectEmptyPlan()`, and ID consistency tracking via `statecheck.CompareValue()`.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: BCM JSON-RPC API (cookie-based auth)
**Testing**: Acceptance tests with TF_ACC=1
**Target Platform**: BCM Cluster at 172.21.15.254:8081
**Project Type**: Terraform Provider (single Go project)
**Performance Goals**: Tests complete within 120m timeout
**Constraints**: BCM API eventual consistency (2s sleep after updates)
**Scale/Scope**: 28 optional fields requiring test coverage

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Based on `.specify/memory/constitution.md`:

| Gate | Status | Notes |
|------|--------|-------|
| TDD-FIRST | PASS | Tests are written before implementation (tests exist, verifying additional coverage) |
| ACCEPTANCE-TESTS | PASS | All new tests are acceptance tests with TF_ACC=1 |
| MODERN-PATTERNS | PASS | Using statecheck, plancheck, knownvalue patterns |
| IDEMPOTENCY | PASS | All tests include plancheck.ExpectEmptyPlan() verification |
| ID-CONSISTENCY | PASS | Using statecheck.CompareValue(compare.ValuesSame()) |

## Project Structure

### Documentation (this feature)

```text
specs/070-category-optional-fields-tests/
├── plan.md              # This file (/speckit.plan command output)
├── spec.md              # Feature specification
├── checklists/          # Test checklists
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
# Terraform Provider structure
internal/provider/
├── resource_cmdevice_category.go       # Resource implementation (existing)
├── resource_cmdevice_category_test.go  # Test file (MODIFIED - add new tests)
└── test_helpers.go                     # Shared test utilities (existing)

# Tests follow acceptance test patterns with:
# - testAccCMDeviceCategoryResourceConfig_* helper functions
# - TestAccCMDeviceCategoryResource_* test functions
```

**Structure Decision**: Modifications are contained within the existing test file (`resource_cmdevice_category_test.go`). No new files required.

---

## Phase 0: Research

### Current Test Coverage Analysis

The test file has 21 existing test functions covering 42 field verifications. Analysis of untested fields:

| Field Category | Fields | BCM Persistence | Import Support |
|---------------|--------|-----------------|----------------|
| Simple Strings | io_scheduler, kernel_version, use_exclusively_for, exclude_list_manipulate_script | YES | YES |
| Booleans | node_installer_disk, version_config_files, data_node | YES (as defaults) | YES |
| Auth Enums | authentication_service, interactive_user | YES | YES |
| Exclude Lists | exclude_list_full/grab/grabnew/sync/update | YES | YES |
| Nested: modules | KernelModuleCategoryModel list | UNKNOWN | UNKNOWN |
| Nested: bmc_settings | BMCSettingsModel object | YES | PARTIAL (password excluded) |
| Nested: fsmounts | FSMountModel list | YES | YES |
| Nested: fsexports | FSExportModel list | NO (like static_routes) | NO |
| Nested: roles | CategoryRoleModel list | NO (like gpu_settings) | NO |

### BCM API Behavior Research

Based on existing test patterns:
1. **BCM Default Values**: BCM returns default values for Optional fields even when not specified
   - Workaround: Explicitly set BCM defaults in test configs (see `testAccCMDeviceCategoryResourceConfig_InstallationModes`)
2. **Empty Array Handling**: BCM converts empty arrays to null
   - Workaround: Don't test empty list transitions
3. **Non-Persisted Fields**: Some nested lists (static_routes, fsexports, roles, gpu_settings) don't persist after category creation
   - Workaround: Add to ImportStateVerifyIgnore list

### Decision: Test Approach

- **Decision**: Group tests by field type to minimize BCM resource creation
- **Rationale**: Each test creates/destroys a BCM category (~10-15s), grouping reduces total test time
- **Alternatives Rejected**:
  - Individual tests per field (too slow, ~30min total)
  - Single monolithic test (harder to debug failures)

---

## Phase 1: Design

### Data Model

No new data models required. Using existing models from `resource_cmdevice_category.go`:

```go
// Existing models used in tests
type CMDeviceCategoryResourceModel struct {
    // Simple string fields
    IOScheduler                 types.String `tfsdk:"io_scheduler"`
    KernelVersion               types.String `tfsdk:"kernel_version"`
    UseExclusivelyFor           types.String `tfsdk:"use_exclusively_for"`
    ExcludeListManipulateScript types.String `tfsdk:"exclude_list_manipulate_script"`

    // Boolean fields
    NodeInstallerDisk    types.Bool `tfsdk:"node_installer_disk"`
    VersionConfigFiles   types.Bool `tfsdk:"version_config_files"`
    DataNode             types.Bool `tfsdk:"data_node"`

    // Auth enum fields
    AuthenticationService types.String `tfsdk:"authentication_service"`
    InteractiveUser       types.String `tfsdk:"interactive_user"`

    // Exclude lists
    ExcludeListFull    types.String `tfsdk:"exclude_list_full"`
    ExcludeListGrab    types.String `tfsdk:"exclude_list_grab"`
    ExcludeListGrabnew types.String `tfsdk:"exclude_list_grabnew"`
    ExcludeListSync    types.String `tfsdk:"exclude_list_sync"`
    ExcludeListUpdate  types.String `tfsdk:"exclude_list_update"`

    // Nested objects
    Modules     types.List   `tfsdk:"modules"`     // []KernelModuleCategoryModel
    BMCSettings types.Object `tfsdk:"bmc_settings"` // BMCSettingsModel
    FSMounts    types.List   `tfsdk:"fsmounts"`    // []FSMountModel
    FSExports   types.List   `tfsdk:"fsexports"`   // []FSExportModel
    Roles       types.List   `tfsdk:"roles"`       // []CategoryRoleModel
}
```

### API Contracts

No API changes required. Tests verify existing BCM API contracts through Terraform acceptance tests.

### Test Function Design

```text
New Test Functions (9 total):
1. TestAccCMDeviceCategoryResource_SimpleStringFields
   - Tests: io_scheduler, use_exclusively_for, exclude_list_manipulate_script
   - Note: kernel_version skipped (requires valid kernel path)

2. TestAccCMDeviceCategoryResource_BooleanFieldsNonDefault
   - Tests: node_installer_disk=true, version_config_files=true, data_node=true

3. TestAccCMDeviceCategoryResource_AuthEnumFields
   - Tests: authentication_service="LDAP", interactive_user="NEVER"

4. TestAccCMDeviceCategoryResource_ExcludeLists
   - Tests: All 5 exclude_list_* fields with multi-line content

5. TestAccCMDeviceCategoryResource_KernelModules
   - Tests: modules list with name and parameters

6. TestAccCMDeviceCategoryResource_BMCSettings
   - Tests: bmc_settings object (non-sensitive fields)

7. TestAccCMDeviceCategoryResource_FilesystemMounts
   - Tests: fsmounts list with NFS mount configuration

8. TestAccCMDeviceCategoryResource_FilesystemExports
   - Tests: fsexports list (add to ImportStateVerifyIgnore)

9. TestAccCMDeviceCategoryResource_RolesConfiguration
   - Tests: roles list (add to ImportStateVerifyIgnore)
```

### Quickstart

```bash
# Run all category optional field tests
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMDeviceCategoryResource_(SimpleString|Boolean|AuthEnum|Exclude|Kernel|BMC|Filesystem|Roles)"

# Run specific test
TF_ACC=1 go test -v -timeout 30m ./internal/provider/ \
  -run "TestAccCMDeviceCategoryResource_SimpleStringFields"
```

---

## Task Summary

| ID | Task | Priority | Est. Time | Dependencies |
|----|------|----------|-----------|--------------|
| T001 | SimpleStringFields test | P1 | 1h | None |
| T002 | BooleanFieldsNonDefault test | P1 | 1h | None |
| T003 | AuthEnumFields test | P1 | 1h | None |
| T004 | ExcludeLists test | P2 | 1.5h | None |
| T005 | KernelModules test | P3 | 1h | None |
| T006 | BMCSettings test | P3 | 1h | None |
| T007 | FilesystemMounts test | P4 | 1h | None |
| T008 | FilesystemExports test | P4 | 1h | None |
| T009 | RolesConfiguration test | P4 | 1h | None |
| T010 | Integration testing | P1 | 2h | T001-T009 |

**Total Estimated Time**: 11-12 hours

---

## Detailed Task Specifications

### T001: SimpleStringFields Test

**Objective**: Test io_scheduler, use_exclusively_for, exclude_list_manipulate_script fields

**Test Steps**:
1. Create category with io_scheduler="noop", use_exclusively_for="GPU", exclude_list_manipulate_script="/path/script.sh"
2. Verify all values with statecheck.ExpectKnownValue()
3. Verify idempotency with plancheck.ExpectEmptyPlan()
4. Update to io_scheduler="deadline", use_exclusively_for="CPU"
5. Verify update with state checks
6. Verify idempotency after update
7. Import and verify

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_SimpleStringFields(name, ioScheduler, useExclusivelyFor, excludeScript string) string
```

### T002: BooleanFieldsNonDefault Test

**Objective**: Test boolean fields with non-default (true) values

**Test Steps**:
1. Create category with node_installer_disk=true, version_config_files=true, data_node=true
2. Verify all values with knownvalue.Bool(true)
3. Verify idempotency
4. Update to node_installer_disk=false
5. Verify update
6. Verify idempotency after update
7. Import and verify

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_BooleanFields(name string, nodeInstallerDisk, versionConfigFiles, dataNode bool) string
```

### T003: AuthEnumFields Test

**Objective**: Test authentication_service and interactive_user enum values

**Test Steps**:
1. Create category with authentication_service="LDAP", interactive_user="NEVER"
2. Verify enum values persist
3. Verify idempotency
4. Update to authentication_service="SSSD", interactive_user="ALWAYS"
5. Verify update
6. Import and verify

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_AuthEnumFields(name, authService, interactiveUser string) string
```

### T004: ExcludeLists Test

**Objective**: Test all 5 exclude_list_* fields with multi-line content

**Test Steps**:
1. Create category with all exclude lists populated (rsync patterns)
2. Verify content preservation with StringExact matcher
3. Verify idempotency
4. Update exclude_list_full with additional patterns
5. Verify update preserves newlines
6. Import and verify

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_ExcludeLists(name, excludeFull, excludeGrab, excludeGrabnew, excludeSync, excludeUpdate string) string
```

### T005: KernelModules Test

**Objective**: Test modules list field

**Test Steps**:
1. Create category with modules = [{ name = "nvidia", parameters = "..." }]
2. Verify list size with knownvalue.ListSizeExact(1)
3. Verify idempotency
4. Add second module to list
5. Verify list size = 2
6. Import (add to ImportStateVerifyIgnore if not persisted)

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_KernelModules(name string, moduleCount int) string
```

### T006: BMCSettings Test

**Objective**: Test bmc_settings nested object

**Test Steps**:
1. Create category with bmc_settings = { user_name = "admin", privilege = "ADMINISTRATOR", firmware_manage_mode = "AUTO" }
2. Verify nested attributes with tfjsonpath.New("bmc_settings").AtMapKey("user_name")
3. Verify idempotency
4. Update privilege to "OPERATOR"
5. Import (add password to ImportStateVerifyIgnore)

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_BMCSettings(name, userName, privilege, firmwareMode string) string
```

### T007: FilesystemMounts Test

**Objective**: Test fsmounts list field

**Test Steps**:
1. Create category with fsmounts = [{ device = "head:/home", mountpoint = "/home", filesystem = "nfs", mountoptions = "defaults" }]
2. Verify list size
3. Verify idempotency
4. Add second mount
5. Import and verify

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_FilesystemMounts(name string, mountCount int) string
```

### T008: FilesystemExports Test

**Objective**: Test fsexports list field

**Test Steps**:
1. Get valid network UUID from data source
2. Create category with fsexports = [{ path = "/shared", network = UUID, allow_write = true }]
3. Verify list size
4. Verify idempotency
5. Import (add fsexports to ImportStateVerifyIgnore - not persisted by BCM)

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_FilesystemExports(name string, exportCount int) string
```

### T009: RolesConfiguration Test

**Objective**: Test roles list field

**Test Steps**:
1. Create category with roles = [{ name = "compute", child_type = "ComputeRole", add_services = true }]
2. Verify list size
3. Verify idempotency
4. Import (add roles to ImportStateVerifyIgnore - not persisted by BCM)

**Config Helper**:
```go
func testAccCMDeviceCategoryResourceConfig_Roles(name string, roleCount int) string
```

### T010: Integration Testing

**Objective**: Run full test suite and document results

**Steps**:
1. Run all new tests
2. Document pass/fail for each test
3. Document any BCM API limitations discovered
4. Update spec with test results
5. Calculate final coverage percentage

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| BCM API doesn't persist some fields | Medium | Add to ImportStateVerifyIgnore, document limitation |
| Test timeout | Low | Group tests efficiently, 120m total timeout |
| BCM default value conflicts | Medium | Explicitly set all related fields to BCM defaults |
| kernel_version requires valid path | Low | Skip field or use computed value from software image |

## Success Criteria

1. All 9 new test functions pass
2. All tests are idempotent (empty plan on second apply)
3. Import verification passes for BCM-persistable fields
4. Test coverage increases to 95%+ for optional fields
5. No test timeout issues
6. Documentation of any BCM API limitations

---

## Appendix: Required Imports

```go
import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "regexp"
    "strings"
    "testing"
    "time"

    "github.com/hashicorp/terraform-plugin-testing/compare"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/terraform"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)
```

## Appendix: Base Test Config Template

```go
func testAccCMDeviceCategoryResourceConfig_BaseTemplate(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }

  # Additional fields added per test...
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}
```
