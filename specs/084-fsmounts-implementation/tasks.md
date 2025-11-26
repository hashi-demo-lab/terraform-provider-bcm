# Tasks: fsmounts Field Implementation

**Input**: Design documents from `/workspace/specs/084-fsmounts-implementation/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/fsmounts-api.yaml
**TDD Workflow**: RED-GREEN-REFACTOR - Write failing tests first, implement minimal code to pass, then refactor

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies on incomplete tasks)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

## User Story Mapping

| Story | Title | Priority | Description |
|-------|-------|----------|-------------|
| US1 | Configure Filesystem Mounts | P1 | Core functionality - serialize/parse fsmounts |
| US2 | Update Filesystem Mounts | P1 | Add/modify/remove mount configurations |
| US3 | Populate Computed UUID | P2 | BCM-assigned UUID population in state |
| US4 | Import Category with Mounts | P2 | Import support for existing categories |

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Verify existing schema and model structures are in place

- [ ] T001 Verify FSMountModel struct exists in internal/provider/resource_cmdevice_category.go:158-168
- [ ] T002 Verify fsmounts schema definition exists in internal/provider/resource_cmdevice_category.go:478-517
- [ ] T003 Identify insertion points for new code in internal/provider/resource_cmdevice_category.go

**Acceptance Criteria**:
- FSMountModel struct has all required fields: UUID, Device, Mountpoint, Filesystem, MountOptions, Fsck, Dump, RDMA
- Schema defines uuid (Computed), device (Required), mountpoint (Required), filesystem (Required), and optional fields
- Insertion points documented: buildAPIEntity (~line 2023), readCategory (~lines 2184-2196), CRUD functions

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Create merge function that all CRUD operations depend on

**CRITICAL**: This function MUST be complete before implementing Create/Read/Update modifications

- [ ] T004 Implement mergeFSMountsWithAPIResponse function after mergeRolesWithAPIResponse (~line 2560) in internal/provider/resource_cmdevice_category.go

**Acceptance Criteria for T004**:
- Function signature: `func mergeFSMountsWithAPIResponse(ctx context.Context, originalMounts types.List, apiMounts types.List) types.List`
- Handles null originalMounts by returning null
- Handles unknown originalMounts by using API response
- Handles null/unknown apiMounts by preserving original
- Builds lookup map by device+mountpoint combination
- Merges: preserves user config (device, mountpoint, filesystem, mountoptions, fsck, dump, rdma) + populates UUID from API
- Generates UUID if mount not in API response (BCM didn't persist)
- Converts merged results back to types.List

**Checkpoint**: Foundational ready - User Story implementation can now begin

---

## Phase 3: User Story 1 - Configure Filesystem Mounts (Priority: P1)

**Goal**: Enable users to define fsmount configurations that are serialized to BCM API and parsed from responses

**Independent Test**: Create a category with fsmounts, verify mounts appear in Terraform state after apply

### Tests for User Story 1 (RED Phase)

> **TDD: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T005 [US1] Write TestAccCMDeviceCategory_FSMountsBasic acceptance test in internal/provider/resource_cmdevice_category_test.go

**Test T005 Acceptance Criteria**:
- Creates category with single fsmount (device=/dev/sdb1, mountpoint=/data, filesystem=xfs)
- Uses testAccConfigWithManagementNetwork pattern for provider setup
- Verifies mount appears in state using resource.TestCheckResourceAttr
- Verifies no drift on immediate re-plan (idempotency)
- Uses generateUniqueTestName for category name
- Includes CheckDestroy for cleanup

- [ ] T006 [US1] Write TestAccCMDeviceCategory_FSMountsMultiple acceptance test in internal/provider/resource_cmdevice_category_test.go

**Test T006 Acceptance Criteria**:
- Creates category with two fsmounts: local device (/dev/sdb1) and NFS mount (nfs-server:/export)
- Tests mountoptions and rdma optional fields
- Verifies both mounts appear in state
- Verifies idempotency with plancheck.ExpectEmptyPlan

**Run tests to verify they FAIL**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_FSMountsBasic|TestAccCMDeviceCategory_FSMountsMultiple"
```

### Implementation for User Story 1 (GREEN Phase)

- [ ] T007 [US1] Add fsmounts serialization in buildAPIEntity function (~line 2023) in internal/provider/resource_cmdevice_category.go

**Task T007 Acceptance Criteria**:
- Check `!model.FSMounts.IsNull() && !model.FSMounts.IsUnknown()` before processing
- Use `model.FSMounts.ElementsAs(ctx, &mounts, false)` to extract typed FSMountModel
- Build `[]map[string]interface{}` with baseType "FSMount"
- Map fields: device->device, mountpoint->path, filesystem->type, mountoptions->options
- Include UUID if present (for updates)
- Include optional fields (fsck, dump, rdma) only if not null
- Set `entity["fsmounts"] = mountsList`

- [ ] T008 [US1] Replace fsmounts parsing in readCategory function (~lines 2184-2196) in internal/provider/resource_cmdevice_category.go

**Task T008 Acceptance Criteria**:
- Define fsMountObjectType with all 8 attributes (uuid, device, mountpoint, filesystem, mountoptions, fsck, dump, rdma)
- Check for `categoryData["fsmounts"].([]interface{})` with length > 0
- Use helper functions (getStringValue, getBoolValue) for null-safe extraction
- Map BCM fields to Terraform: path->mountpoint, type->filesystem, options->mountoptions
- Set model.FSMounts with types.ListValue or types.ListNull

- [ ] T009 [US1] Add fsmounts preservation in Create function (~lines 828 and 1052) in internal/provider/resource_cmdevice_category.go

**Task T009 Acceptance Criteria**:
- Before readCategory call: `planFSMounts := plan.FSMounts`
- After readCategory call: `plan.FSMounts = mergeFSMountsWithAPIResponse(ctx, planFSMounts, plan.FSMounts)`

- [ ] T010 [US1] Add fsmounts preservation in Read function (~lines 1077 and 1195) in internal/provider/resource_cmdevice_category.go

**Task T010 Acceptance Criteria**:
- Before readCategory call: `originalFSMounts := state.FSMounts`
- After readCategory call: `state.FSMounts = mergeFSMountsWithAPIResponse(ctx, originalFSMounts, state.FSMounts)`

**Run tests to verify they PASS**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_FSMountsBasic|TestAccCMDeviceCategory_FSMountsMultiple"
```

**Checkpoint**: User Story 1 complete - Basic fsmounts configuration works

---

## Phase 4: User Story 2 - Update Filesystem Mounts (Priority: P1)

**Goal**: Enable users to add, modify, or remove fsmount configurations on existing categories

**Independent Test**: Update an existing category to add/modify/remove mounts and verify changes are applied

### Tests for User Story 2 (RED Phase)

- [ ] T011 [US2] Write TestAccCMDeviceCategory_FSMountsUpdate acceptance test in internal/provider/resource_cmdevice_category_test.go

**Test T011 Acceptance Criteria**:
- Step 1: Create category with one fsmount
- Step 2: Add second fsmount, verify both exist
- Step 3: Modify first fsmount's mountoptions, verify change applied
- Step 4: Remove second fsmount, verify only first remains
- Each step verifies state and idempotency

- [ ] T012 [US2] Write TestAccCMDeviceCategory_FSMountsIdempotency acceptance test in internal/provider/resource_cmdevice_category_test.go

**Test T012 Acceptance Criteria**:
- Create category with fsmounts
- Immediate re-apply with same config
- Use ConfigPlanChecks with plancheck.ExpectEmptyPlan()
- Verifies no unintended drift detection

**Run tests to verify they FAIL**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_FSMountsUpdate|TestAccCMDeviceCategory_FSMountsIdempotency"
```

### Implementation for User Story 2 (GREEN Phase)

- [ ] T013 [US2] Add fsmounts preservation in Update function (~lines 1321 and 1390) in internal/provider/resource_cmdevice_category.go

**Task T013 Acceptance Criteria**:
- Before readCategory call: `planFSMounts := plan.FSMounts`
- After readCategory call: `plan.FSMounts = mergeFSMountsWithAPIResponse(ctx, planFSMounts, plan.FSMounts)`

**Run tests to verify they PASS**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_FSMountsUpdate|TestAccCMDeviceCategory_FSMountsIdempotency"
```

**Checkpoint**: User Story 2 complete - CRUD operations on fsmounts work

---

## Phase 5: User Story 3 - Populate Computed UUID (Priority: P2)

**Goal**: BCM-assigned UUIDs for fsmounts are populated in Terraform state

**Independent Test**: Create a category with fsmounts and verify each mount's uuid attribute is populated

### Tests for User Story 3 (RED Phase)

- [ ] T014 [US3] Write TestAccCMDeviceCategory_FSMountsDrift acceptance test in internal/provider/resource_cmdevice_category_test.go

**Test T014 Acceptance Criteria**:
- Step 1: Create category with fsmount
- Step 2: PreConfig modifies fsmount externally via BCM API using createTestBCMClient
- Step 2: Use ConfigPlanChecks with plancheck.ExpectNonEmptyPlan() to verify drift detected
- Step 3: Terraform restores desired state
- Uses getResourceUUIDByName helper for UUID lookup

### Implementation for User Story 3 (GREEN Phase)

> **Note**: UUID population is already implemented in mergeFSMountsWithAPIResponse (T004). This phase validates the behavior.

- [ ] T015 [US3] Add debug logging for UUID merge operations in mergeFSMountsWithAPIResponse in internal/provider/resource_cmdevice_category.go

**Task T015 Acceptance Criteria**:
- Add tflog.Debug for each merge case: null original, unknown original, null/unknown API, match found, match not found
- Log matched keys for debugging
- Log generated UUIDs when API doesn't return data

**Run tests to verify they PASS**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_FSMountsDrift"
```

**Checkpoint**: User Story 3 complete - UUID population and drift detection work

---

## Phase 6: User Story 4 - Import Category with Mounts (Priority: P2)

**Goal**: Import existing BCM categories with fsmount configurations

**Independent Test**: Import an existing category that has fsmounts configured and verify mounts appear in state

### Tests for User Story 4 (RED Phase)

- [ ] T016 [US4] Write TestAccCMDeviceCategory_FSMountsImport acceptance test in internal/provider/resource_cmdevice_category_test.go

**Test T016 Acceptance Criteria**:
- Step 1: Create category with fsmounts via Terraform
- Step 2: Import using ResourceName and ImportState: true
- Step 3: ImportStateVerify: true to verify state matches
- May need ImportStateVerifyIgnore for computed fields with timing differences

### Implementation for User Story 4 (GREEN Phase)

> **Note**: Import is already supported via readCategory parsing (T008). This phase validates the behavior.

- [ ] T017 [US4] Verify import handles null original state correctly in mergeFSMountsWithAPIResponse in internal/provider/resource_cmdevice_category.go

**Task T017 Acceptance Criteria**:
- When originalMounts is null (import case), function returns null appropriately
- When originalMounts is unknown (fresh import), function uses API response
- Add test case to validate import specifically populates from API

**Run tests to verify they PASS**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory_FSMountsImport"
```

**Checkpoint**: User Story 4 complete - Import with fsmounts works

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Quality assurance and documentation

- [ ] T018 [P] Run all existing CMDeviceCategory tests to verify no regressions in internal/provider/resource_cmdevice_category_test.go
- [ ] T019 [P] Run make fmt to format code
- [ ] T020 [P] Run make lint to check for linting issues
- [ ] T021 Run make generate to update documentation
- [ ] T022 Validate quickstart.md test scenarios work as documented

**Regression Test Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory"
```

**Code Quality Commands**:
```bash
make fmt
make lint
make generate
```

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 1 (Setup)
    |
    v
Phase 2 (Foundational) - mergeFSMountsWithAPIResponse
    |
    v
Phase 3 (US1) - Basic fsmounts (P1)
    |
    v
Phase 4 (US2) - Update fsmounts (P1)
    |
    +---> Phase 5 (US3) - UUID population (P2) [can start after Phase 3]
    |
    +---> Phase 6 (US4) - Import support (P2) [can start after Phase 3]
    |
    v
Phase 7 (Polish) - After all stories complete
```

### Within Each Phase (TDD Order)

1. Write failing tests (RED)
2. Implement minimal code to pass (GREEN)
3. Refactor while keeping tests green
4. Verify all tests pass before next phase

### Parallel Opportunities

**After Phase 2 completes**:
- T005 and T006 can run in parallel (different test functions)
- T007 and T008 can run in parallel (different functions in same file)

**After Phase 3 completes**:
- Phase 5 (US3) and Phase 6 (US4) can proceed in parallel
- T018, T019, T020 can run in parallel (different tools)

---

## Implementation Strategy

### MVP First (User Stories 1 + 2)

1. Complete Phase 1: Setup (verify existing structures)
2. Complete Phase 2: Foundational (mergeFSMountsWithAPIResponse)
3. Complete Phase 3: User Story 1 (basic fsmounts)
4. Complete Phase 4: User Story 2 (update fsmounts)
5. **STOP and VALIDATE**: All basic CRUD operations work
6. Deploy/demo if ready

### Full Implementation

1. MVP (Phases 1-4)
2. Phase 5: User Story 3 (UUID population + drift detection)
3. Phase 6: User Story 4 (import support)
4. Phase 7: Polish (regression tests, code quality, docs)

---

## Success Criteria Verification

| Criterion | Test | Status |
|-----------|------|--------|
| SC-001: Mounts in state | TestAccCMDeviceCategory_FSMountsBasic | Pending |
| SC-002: Idempotency | TestAccCMDeviceCategory_FSMountsIdempotency | Pending |
| SC-003: No regressions | All existing CMDeviceCategory tests | Pending |
| SC-004: New tests pass | All FSMounts tests | Pending |
| SC-005: NFS + local mounts | TestAccCMDeviceCategory_FSMountsMultiple | Pending |

---

## File Modification Summary

**Primary File**: `internal/provider/resource_cmdevice_category.go`

| Location | Modification | Task |
|----------|--------------|------|
| ~line 828 | Add `planFSMounts := plan.FSMounts` | T009 |
| ~line 1052 | Add merge call after readCategory | T009 |
| ~line 1077 | Add `originalFSMounts := state.FSMounts` | T010 |
| ~line 1195 | Add merge call after readCategory | T010 |
| ~line 1321 | Add `planFSMounts := plan.FSMounts` | T013 |
| ~line 1390 | Add merge call after readCategory | T013 |
| ~line 2023 | Add fsmounts serialization | T007 |
| ~lines 2184-2196 | Replace null with parsing | T008 |
| ~line 2560 | Add mergeFSMountsWithAPIResponse function | T004 |

**Test File**: `internal/provider/resource_cmdevice_category_test.go`

| Test Function | Task |
|---------------|------|
| TestAccCMDeviceCategory_FSMountsBasic | T005 |
| TestAccCMDeviceCategory_FSMountsMultiple | T006 |
| TestAccCMDeviceCategory_FSMountsUpdate | T011 |
| TestAccCMDeviceCategory_FSMountsIdempotency | T012 |
| TestAccCMDeviceCategory_FSMountsDrift | T014 |
| TestAccCMDeviceCategory_FSMountsImport | T016 |

---

## Notes

- All tasks include exact line numbers from plan.md for quick navigation
- TDD workflow: RED (failing test) -> GREEN (minimal implementation) -> REFACTOR
- Use `TF_ACC=1` environment variable for acceptance tests
- BCM cluster at 172.21.15.254 required for acceptance tests
- Follow existing patterns: fsexports for serialization/parsing, roles for merge strategy
- Field mapping: mountpoint<->path, filesystem<->type, mountoptions<->options
