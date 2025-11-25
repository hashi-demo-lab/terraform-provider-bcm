# Implementation Plan: Category Test Coverage Enhancement

**Branch**: `001-category-test-coverage` | **Date**: 2025-11-25 | **Spec**: [spec.md](/workspace/specs/001-category-test-coverage/spec.md)
**Input**: Feature specification from `/specs/001-category-test-coverage/spec.md`
**GitHub Issue**: #66

## Summary

Add comprehensive acceptance test coverage for 25+ untested optional fields in the `bcm_cmdevice_category` resource. Tests will use modern terraform-plugin-testing v1.13.3+ patterns including `statecheck.ExpectKnownValue`, `plancheck.ExpectEmptyPlan`, and `knownvalue` matchers. Tests are grouped by functional area to balance coverage and execution time.

**Key Finding from Research**: Many optional fields listed in the spec are not yet implemented in the resource's `buildAPIEntity` and `readCategory` functions. This plan addresses two tracks:
1. **Track A**: Test fields that ARE implemented (install_mode, new_node_install_mode, etc.)
2. **Track B**: Document and defer fields that need implementation work first

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (test infrastructure only)
**Testing**: Go acceptance tests with `TF_ACC=1`
**Target Platform**: BCM cluster at 172.21.15.254:8081
**Project Type**: Single (Terraform provider test enhancement)
**Performance Goals**: All new tests complete within 15 minutes
**Constraints**: Tests must be environment-portable (no hardcoded values)
**Scale/Scope**: 8-10 new test functions covering ~25 optional fields

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| TDD First | PASS | Tests are the deliverable; implementation exists |
| No New Dependencies | PASS | Using existing terraform-plugin-testing |
| Environment Portable | PASS | Tests use `generateUniqueTestName()` and data source lookups |
| Modern Patterns | PASS | Using statecheck, plancheck, knownvalue matchers |
| Idempotency Required | PASS | All tests include `ExpectEmptyPlan()` checks |

## Project Structure

### Documentation (this feature)

```text
specs/001-category-test-coverage/
├── spec.md              # Feature requirements (input)
├── plan.md              # This file (Phase 1 output)
├── research.md          # Phase 0 output - field analysis
├── data-model.md        # Phase 1 output - test entities
├── quickstart.md        # Phase 1 output - developer guide
└── tasks.md             # Phase 2 output (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmdevice_category.go       # Existing resource implementation
├── resource_cmdevice_category_test.go  # Test file to enhance
└── test_helpers.go                     # Shared test utilities
```

**Structure Decision**: All new tests will be added to the existing test file `resource_cmdevice_category_test.go` following the established patterns.

---

## Phase 0: Research Summary

Completed. See [research.md](/workspace/specs/001-category-test-coverage/research.md) for full analysis.

### Key Findings

1. **Current Coverage**: ~18 fields tested across 15 existing test functions
2. **Fields with Implementation**: install_mode, new_node_install_mode, kernel_version (can test now)
3. **Fields Needing Implementation**: ~20 fields not mapped in buildAPIEntity/readCategory

### Implementation Status Matrix

| Category | Field | buildAPIEntity | readCategory | Test Status |
|----------|-------|----------------|--------------|-------------|
| Installation | install_mode | YES | YES | **TO TEST** |
| Installation | new_node_install_mode | YES | YES | **TO TEST** |
| Network | name_servers | NO | NO (null) | NEEDS IMPL |
| Network | search_domains | NO | NO (null) | NEEDS IMPL |
| Network | time_servers | NO | NO (null) | NEEDS IMPL |
| Kernel | io_scheduler | NO | NO | NEEDS IMPL |
| Kernel | kernel_version | YES | YES | **TO TEST** |
| Scripts | initialize | NO | NO | NEEDS IMPL |
| Scripts | finalize | NO | NO | NEEDS IMPL |
| BMC | bmc_settings.* | NO | NO (null) | NEEDS IMPL |
| Modules | modules[].* | NO | NO (null) | NEEDS IMPL |
| Exclude | exclude_list_* | NO | NO | NEEDS IMPL |
| Misc | fips | NO | NO | NEEDS IMPL |
| Misc | data_node | NO | NO | NEEDS IMPL |
| Misc | authentication_service | NO | NO | NEEDS IMPL |

---

## Phase 1: Design

### Test Grouping Strategy

Based on research findings, tests are grouped into two tracks:

#### Track A: Tests for Implemented Fields (Priority)

| Test Function | Fields | User Story | Priority |
|---------------|--------|------------|----------|
| `TestAccCMDeviceCategoryResource_InstallationModes` | install_mode, new_node_install_mode | US1 | P1 |
| `TestAccCMDeviceCategoryResource_KernelVersion` | kernel_version | US3 | P2 |

#### Track B: Tests Pending Implementation (Deferred)

These tests require resource implementation work before they can be written:

| Test Function | Fields | User Story | Blocker |
|---------------|--------|------------|---------|
| `TestAccCMDeviceCategoryResource_NetworkListFields` | name_servers, search_domains, time_servers | US2 | Not in buildAPIEntity |
| `TestAccCMDeviceCategoryResource_IOScheduler` | io_scheduler | US3 | Not in buildAPIEntity |
| `TestAccCMDeviceCategoryResource_ProvisioningScripts` | initialize, finalize | US4 | Not in buildAPIEntity |
| `TestAccCMDeviceCategoryResource_BMCSettings` | bmc_settings.* | US5 | Not in buildAPIEntity |
| `TestAccCMDeviceCategoryResource_KernelModules` | modules[].* | US6 | Not in buildAPIEntity |
| `TestAccCMDeviceCategoryResource_ExcludeLists` | exclude_list_* | US7 | Not in buildAPIEntity |
| `TestAccCMDeviceCategoryResource_MiscellaneousFields` | fips, data_node, etc. | US8 | Not in buildAPIEntity |

### Test Function Designs

#### 1. TestAccCMDeviceCategoryResource_InstallationModes

**Purpose**: Test install_mode and new_node_install_mode fields (User Story 1)

**Test Steps**:
1. Create with install_mode="AUTO", new_node_install_mode="FULL"
2. Idempotency check (ExpectEmptyPlan)
3. Update to install_mode="FULL", new_node_install_mode="MINIMAL"
4. Idempotency check
5. Import and verify

**State Checks**:
```go
statecheck.ExpectKnownValue("bcm_cmdevice_category.test",
    tfjsonpath.New("install_mode"), knownvalue.StringExact("AUTO"))
statecheck.ExpectKnownValue("bcm_cmdevice_category.test",
    tfjsonpath.New("new_node_install_mode"), knownvalue.StringExact("FULL"))
```

**Config Helper**: `testAccCMDeviceCategoryResourceConfig_InstallationModes(name, installMode, newNodeInstallMode string)`

#### 2. TestAccCMDeviceCategoryResource_KernelVersion

**Purpose**: Test kernel_version field (User Story 3, partial)

**Consideration**: kernel_version requires a valid kernel path from the actual software image. This may be environment-dependent.

**Test Steps**:
1. Create with kernel_version set
2. Idempotency check
3. Update kernel_version
4. Idempotency check
5. Import and verify

**Config Helper**: `testAccCMDeviceCategoryResourceConfig_KernelVersion(name, kernelVersion string)`

### Config Helper Template

All new config helpers follow this pattern:

```go
func testAccCMDeviceCategoryResourceConfig_FEATURE(name string, params...) string {
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

  # FEATURE-SPECIFIC FIELDS
  feature_field = %[5]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        param,
    )
}
```

### Modern Testing Pattern Examples

#### State Checks (statecheck.ExpectKnownValue)

```go
ConfigStateChecks: []statecheck.StateCheck{
    // String field
    statecheck.ExpectKnownValue("bcm_cmdevice_category.test",
        tfjsonpath.New("install_mode"),
        knownvalue.StringExact("AUTO")),
    // Computed field
    statecheck.ExpectKnownValue("bcm_cmdevice_category.test",
        tfjsonpath.New("uuid"),
        knownvalue.NotNull()),
    // ID tracking
    compareID.AddStateValue("bcm_cmdevice_category.test",
        tfjsonpath.New("id")),
}
```

#### Idempotency Check (plancheck.ExpectEmptyPlan)

```go
{
    Config: testAccConfig(name, "value"),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
}
```

#### ID Consistency (statecheck.CompareValue)

```go
compareID := statecheck.CompareValue(compare.ValuesSame())
// In each step:
compareID.AddStateValue("bcm_cmdevice_category.test", tfjsonpath.New("id"))
```

---

## Implementation Phases

### Phase 1: Track A Tests (Implemented Fields)

1. **Task T001**: Create `TestAccCMDeviceCategoryResource_InstallationModes` test
2. **Task T002**: Create `testAccCMDeviceCategoryResourceConfig_InstallationModes` helper
3. **Task T003**: Create `testAccCMDeviceCategoryResourceConfig_InstallationModesUpdated` helper
4. **Task T004**: Run and verify test passes
5. **Task T005**: Create `TestAccCMDeviceCategoryResource_KernelVersion` test (if feasible)

### Phase 2: Resource Implementation (Prerequisite for Track B)

**Note**: These tasks are OUT OF SCOPE for this spec but documented for future work.

6. **Task T006**: Implement name_servers, search_domains, time_servers in buildAPIEntity
7. **Task T007**: Implement list parsing in readCategory for network lists
8. **Task T008**: Implement io_scheduler in buildAPIEntity/readCategory
9. **Task T009**: Implement initialize, finalize in buildAPIEntity/readCategory
10. **Task T010**: Implement bmc_settings nested object handling
11. **Task T011**: Implement modules nested list handling
12. **Task T012**: Implement exclude_list_* fields
13. **Task T013**: Implement fips, data_node, misc fields

### Phase 3: Track B Tests (After Implementation)

14. **Task T014**: Create `TestAccCMDeviceCategoryResource_NetworkListFields` test
15. **Task T015**: Create `TestAccCMDeviceCategoryResource_IOScheduler` test
16. **Task T016**: Create `TestAccCMDeviceCategoryResource_ProvisioningScripts` test
17. **Task T017**: Create `TestAccCMDeviceCategoryResource_BMCSettings` test
18. **Task T018**: Create `TestAccCMDeviceCategoryResource_KernelModules` test
19. **Task T019**: Create `TestAccCMDeviceCategoryResource_ExcludeLists` test
20. **Task T020**: Create `TestAccCMDeviceCategoryResource_MiscellaneousFields` test

---

## Complexity Tracking

| Consideration | Decision | Rationale |
|---------------|----------|-----------|
| Many fields not implemented | Focus on Track A first | Tests must pass with current implementation |
| Network lists require implementation | Defer to Phase 2 | Out of scope for test-only spec |
| Nested objects (BMC, modules) complex | Defer to Phase 2 | Requires significant implementation work |
| kernel_version environment-dependent | Test with investigation | May need dynamic lookup |

---

## Success Criteria Mapping

| Success Criteria | Plan Coverage |
|------------------|---------------|
| SC-001: All 8 user stories have tests | Partial: US1 (Track A), others need implementation |
| SC-002: Coverage increases from ~15 to 40+ fields | Partial: +2 fields (Track A), +20 after implementation |
| SC-003: All new tests use modern patterns | Yes: statecheck, plancheck, knownvalue |
| SC-004: Tests complete in 15 minutes | Yes: 2 new tests estimated <5 minutes |
| SC-005: Zero test flakiness | Design includes idempotency, cleanup |
| SC-006: Gap analysis shows improved coverage | Yes: research.md documents gaps |

---

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| BCM API doesn't return installation modes | Low | High | Verify with manual API call first |
| kernel_version requires specific path | Medium | Medium | Use null/empty test if needed |
| Implementation work delays Track B | High | Medium | Document as future work |

---

## Appendix: Field Implementation Checklist

For each field requiring implementation before testing:

- [ ] Add to `buildAPIEntity()` mapping
- [ ] Add to `readCategory()` parsing
- [ ] Handle null/empty values appropriately
- [ ] Test with BCM API manually
- [ ] Write acceptance test

---

## Generated Artifacts

| Artifact | Path | Status |
|----------|------|--------|
| Research | `/workspace/specs/001-category-test-coverage/research.md` | Complete |
| Data Model | `/workspace/specs/001-category-test-coverage/data-model.md` | Complete |
| Quickstart | `/workspace/specs/001-category-test-coverage/quickstart.md` | Complete |
| Plan | `/workspace/specs/001-category-test-coverage/plan.md` | Complete |
