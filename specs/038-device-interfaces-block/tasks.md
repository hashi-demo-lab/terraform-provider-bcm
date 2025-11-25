# Tasks: Add Interfaces Block to bcm_cmdevice_device

**Input**: Design documents from `/workspace/specs/038-device-interfaces-block/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/interfaces.json

**Tests**: Acceptance tests are REQUIRED per TDD workflow and CLAUDE.md constitution.

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

## Path Conventions

- **Provider source**: `internal/provider/`
- **Examples**: `examples/resources/bcm_cmdevice_device/`
- **Docs**: `docs/resources/` (auto-generated)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Create new files and add interface model struct

- [ ] T001 Create interface helpers file at `/workspace/internal/provider/resource_cmdevice_device_interfaces.go` with package declaration and imports
- [ ] T002 [P] Create test file at `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go` with package declaration and imports
- [ ] T003 [P] Add `DeviceInterfaceModel` struct to `/workspace/internal/provider/resource_cmdevice_device_interfaces.go` per data-model.md definition

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core schema and helper functions that ALL user stories depend on

**CRITICAL**: No user story work can begin until this phase is complete

### Schema and Model Changes

- [ ] T004 Add `Interfaces []DeviceInterfaceModel` field to `CMDeviceDeviceResourceModel` struct in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T005 Add `interfaces` ListNestedBlock to Schema() function in `/workspace/internal/provider/resource_cmdevice_device.go` with all attributes per data-model.md

### Type Mapping Functions

- [ ] T006 [P] Implement `interfaceTypeToBCMChildType()` function in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`
- [ ] T007 [P] Implement `bcmChildTypeToInterfaceType()` function in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`

### API Entity Functions

- [ ] T008 Implement `buildInterfaceAPIEntity()` function in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go` per data-model.md
- [ ] T009 Implement `parseInterfaceFromAPI()` function in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go` per data-model.md

### Legacy Mode Detection

- [ ] T010 Implement `isLegacyMode()` helper function in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go` for backward compatibility

### Schema Validators

- [ ] T011 [P] Add unique interface name validator (per device) to schema in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T012 [P] Add bond members required validator to schema in `/workspace/internal/provider/resource_cmdevice_device.go`

### Test Infrastructure

- [ ] T013 Add `generateUniqueMAC()` test helper to `/workspace/internal/provider/test_helpers.go` if not already present
- [ ] T014 [P] Add `testAccCMDeviceDeviceConfigInterfaceSingle()` config helper to `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Configure Multiple Physical Interfaces (Priority: P1)

**Goal**: Enable administrators to configure multiple physical network interfaces on a device for network segmentation (management, data, storage)

**Independent Test**: Create a device with two physical interfaces on different networks and verify both interfaces are created with correct network assignments

### Tests for User Story 1 (RED Phase - Write Failing Tests First)

- [ ] T015 [US1] Write `TestAccCMDeviceDevice_InterfaceSingle` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T016 [P] [US1] Write `TestAccCMDeviceDevice_InterfaceMultiple` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Implementation for User Story 1 (GREEN Phase)

- [ ] T017 [US1] Update `buildDeviceAPIEntity()` in `/workspace/internal/provider/resource_cmdevice_device.go` to build interfaces array from plan.Interfaces
- [ ] T018 [US1] Update `parseDeviceFromAPI()` in `/workspace/internal/provider/resource_cmdevice_device.go` to extract interfaces array into model
- [ ] T019 [US1] Update Create() method in `/workspace/internal/provider/resource_cmdevice_device.go` to handle interfaces block with UUID generation
- [ ] T020 [US1] Update Read() method in `/workspace/internal/provider/resource_cmdevice_device.go` to populate interfaces from BCM response
- [ ] T021 [US1] Implement provisioning interface selection (first bootable interface) in `/workspace/internal/provider/resource_cmdevice_device.go`

### Verification

- [ ] T022 [US1] Run single interface test and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceSingle`
- [ ] T023 [US1] Run multiple interface test and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceMultiple`

**Checkpoint**: Physical interfaces create/read working - MVP scope complete

---

## Phase 4: User Story 2 - Configure Bonded Interfaces (Priority: P1)

**Goal**: Enable administrators to create bonded network interfaces for network redundancy and increased bandwidth

**Independent Test**: Create a device with a bond interface specifying member interfaces and verify the bond is created with correct member assignment

### Tests for User Story 2 (RED Phase)

- [ ] T024 [US2] Write `TestAccCMDeviceDevice_InterfaceBond` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T025 [P] [US2] Write `testAccCMDeviceDeviceConfigInterfaceBond()` config helper in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Implementation for User Story 2 (GREEN Phase)

- [ ] T026 [US2] Extend `buildInterfaceAPIEntity()` to handle bond-specific fields (members, bondMode) in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`
- [ ] T027 [US2] Extend `parseInterfaceFromAPI()` to parse bond members array in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`
- [ ] T028 [US2] Add bond members to types.List conversion in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`

### Verification

- [ ] T029 [US2] Run bond interface test and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceBond`

**Checkpoint**: Bond interfaces working with member assignment

---

## Phase 5: User Story 3 - Configure BMC/IPMI Interface (Priority: P2)

**Goal**: Enable administrators to configure BMC interfaces for out-of-band management

**Independent Test**: Create a device with a BMC interface block and verify the interface is created with correct type and network assignment

### Tests for User Story 3 (RED Phase)

- [ ] T030 [US3] Write `TestAccCMDeviceDevice_InterfaceBMC` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T031 [P] [US3] Write `testAccCMDeviceDeviceConfigInterfaceBMC()` config helper in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Implementation for User Story 3 (GREEN Phase)

- [ ] T032 [US3] Extend `buildInterfaceAPIEntity()` to handle BMC-specific cardtype mapping in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`
- [ ] T033 [US3] Verify BMC childType mapping (NetworkBMCInterface) in type mapping functions

### Verification

- [ ] T034 [US3] Run BMC interface test and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceBMC`

**Checkpoint**: All interface types (physical, bond, BMC) working

---

## Phase 6: User Story 4 - Import Device with Interfaces (Priority: P2)

**Goal**: Enable administrators to import existing devices with interfaces into Terraform management

**Independent Test**: Import an existing device with multiple interfaces and verify all interface configurations are populated in state

### Tests for User Story 4 (RED Phase)

- [ ] T035 [US4] Write `TestAccCMDeviceDevice_InterfaceImport` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T036 [P] [US4] Write `TestAccCMDeviceDevice_InterfaceImportIdempotency` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Implementation for User Story 4 (GREEN Phase)

- [ ] T037 [US4] Ensure ImportState handler populates interfaces in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T038 [US4] Verify interface UUID preservation on import in Read() method
- [ ] T039 [US4] Add interface ordering normalization (match BCM order to Terraform order) in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`

### Verification

- [ ] T040 [US4] Run import test and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceImport`
- [ ] T041 [US4] Run import idempotency test and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceImportIdempotency`

**Checkpoint**: Import functionality working with full interface state population

---

## Phase 7: User Story 5 - Remove and Replace Interfaces (Priority: P3)

**Goal**: Enable administrators to modify interface configurations without recreating devices

**Independent Test**: Create a device with interfaces, remove one interface block, and verify the interface is removed from the device

### Tests for User Story 5 (RED Phase)

- [ ] T042 [US5] Write `TestAccCMDeviceDevice_InterfaceUpdate` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T043 [P] [US5] Write `TestAccCMDeviceDevice_InterfaceAdd` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T044 [P] [US5] Write `TestAccCMDeviceDevice_InterfaceRemove` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Implementation for User Story 5 (GREEN Phase)

- [ ] T045 [US5] Update Update() method in `/workspace/internal/provider/resource_cmdevice_device.go` to handle interface modifications
- [ ] T046 [US5] Implement interface UUID matching by name for updates in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`
- [ ] T047 [US5] Handle interface removal via `to_be_removed` flag in `/workspace/internal/provider/resource_cmdevice_device_interfaces.go`

### Verification

- [ ] T048 [US5] Run update test and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceDevice_InterfaceUpdate`
- [ ] T049 [US5] Run add/remove tests and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Interface(Add|Remove)"`

**Checkpoint**: Full CRUD lifecycle for interfaces working

---

## Phase 8: Drift Detection and Validation

**Purpose**: Ensure provider detects external changes and validates configurations

### Drift Detection Tests

- [ ] T050 Write `TestAccCMDeviceDevice_InterfaceDrift` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T051 [P] Write `TestAccCMDeviceDevice_InterfaceDriftCorrection` acceptance test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Validation Tests

- [ ] T052 [P] Write `TestAccCMDeviceDevice_InterfaceValidationDuplicateName` test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T053 [P] Write `TestAccCMDeviceDevice_InterfaceValidationBondMembers` test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T054 [P] Write `TestAccCMDeviceDevice_InterfaceValidationInvalidNetwork` test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Implementation

- [ ] T055 Verify drift detection in Read() compares interfaces correctly in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T056 Integrate BCM validateDevice for interface validation errors in Create/Update methods

### Verification

- [ ] T057 Run drift tests and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_InterfaceDrift"`
- [ ] T058 Run validation tests and verify pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_InterfaceValidation"`

**Checkpoint**: Drift detection and validation working

---

## Phase 9: Backward Compatibility

**Purpose**: Ensure existing configurations without interfaces block continue to work

### Backward Compatibility Tests

- [ ] T059 Write `TestAccCMDeviceDevice_LegacyMACOnly` test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`
- [ ] T060 [P] Write `TestAccCMDeviceDevice_MixedLegacyAndInterfaces` test in `/workspace/internal/provider/resource_cmdevice_device_interfaces_test.go`

### Implementation

- [ ] T061 Ensure Create() handles legacy mode (mac field without interfaces block) in `/workspace/internal/provider/resource_cmdevice_device.go`
- [ ] T062 Ensure Read() populates both legacy and interfaces fields on import in `/workspace/internal/provider/resource_cmdevice_device.go`

### Verification

- [ ] T063 Run backward compatibility tests: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Legacy|TestAccCMDeviceDevice_Mixed"`

**Checkpoint**: Backward compatibility verified

---

## Phase 10: Polish and Cross-Cutting Concerns

**Purpose**: Documentation, examples, and code quality improvements

### Terraform Examples

- [ ] T064 [P] Create `/workspace/examples/resources/bcm_cmdevice_device/interfaces_physical.tf` example
- [ ] T065 [P] Create `/workspace/examples/resources/bcm_cmdevice_device/interfaces_bond.tf` example
- [ ] T066 [P] Create `/workspace/examples/resources/bcm_cmdevice_device/interfaces_bmc.tf` example
- [ ] T067 [P] Create `/workspace/examples/resources/bcm_cmdevice_device/interfaces_multi.tf` DGX example

### Documentation

- [ ] T068 Run `make generate` to auto-generate documentation in `/workspace/docs/resources/bcm_cmdevice_device.md`
- [ ] T069 Verify interfaces block appears in generated docs with all attributes documented

### Code Quality

- [ ] T070 [P] Run `make fmt` to format all Go code
- [ ] T071 [P] Run `make lint` to check for linting issues
- [ ] T072 Review and clean up any TODO comments in implementation files

### Final Verification

- [ ] T073 Run all interface tests: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Interface"`
- [ ] T074 Run existing device tests to verify no regressions: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceDevice_Basic"`
- [ ] T075 Validate all examples: `./scripts/test-examples.sh --resources-only`

---

## Dependencies and Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - can start immediately
- **Phase 2 (Foundational)**: Depends on Phase 1 - BLOCKS all user stories
- **Phases 3-7 (User Stories)**: All depend on Phase 2 completion
  - US1 and US2 (P1) can proceed in parallel after Phase 2
  - US3 and US4 (P2) depend on US1 basic implementation
  - US5 (P3) depends on US1 basic implementation
- **Phase 8 (Drift/Validation)**: Depends on all user story implementations
- **Phase 9 (Backward Compat)**: Can run in parallel with Phase 8
- **Phase 10 (Polish)**: Depends on all implementation phases

### User Story Dependencies

| Story | Priority | Depends On | Can Parallelize With |
|-------|----------|------------|---------------------|
| US1 | P1 | Phase 2 (Foundational) | US2 |
| US2 | P1 | Phase 2 (Foundational) | US1 |
| US3 | P2 | US1 (basic interface CRUD) | US4 |
| US4 | P2 | US1 (basic interface CRUD) | US3 |
| US5 | P3 | US1 (basic interface CRUD) | None |

### Parallel Opportunities

**Within Phase 2 (Foundational):**
```
Parallel batch: T006, T007 (type mapping functions)
Parallel batch: T011, T012 (validators)
Parallel batch: T013, T014 (test infrastructure)
```

**Within Phase 3 (US1):**
```
Parallel batch: T015, T016 (tests)
Sequential: T017 -> T018 -> T019 -> T020 -> T021 (implementation)
```

**Within Phase 10 (Polish):**
```
Parallel batch: T064, T065, T066, T067 (examples)
Parallel batch: T070, T071 (code quality)
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Single and Multiple Physical Interfaces)
4. **STOP and VALIDATE**: Test interfaces independently
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational -> Foundation ready
2. Add US1 (Physical Interfaces) -> Test -> Deploy (MVP!)
3. Add US2 (Bond Interfaces) -> Test -> Deploy
4. Add US3 (BMC Interfaces) -> Test -> Deploy
5. Add US4 (Import) -> Test -> Deploy
6. Add US5 (Update/Remove) -> Test -> Deploy
7. Drift + Validation + Backward Compat -> Test -> Deploy
8. Polish -> Final Release

### TDD Cycle Per User Story

1. **RED**: Write acceptance tests (T0XX tests)
2. **GREEN**: Implement minimum code to pass tests
3. **REFACTOR**: Improve code quality, add validators
4. **VERIFY**: Run tests, ensure pass
5. **DOCUMENT**: Update examples if needed

---

## Task Summary

| Phase | Task Count | Parallel Tasks | Story Coverage |
|-------|------------|----------------|----------------|
| Phase 1: Setup | 3 | 2 | N/A |
| Phase 2: Foundational | 11 | 6 | N/A |
| Phase 3: US1 | 9 | 1 | Physical Interfaces |
| Phase 4: US2 | 6 | 1 | Bond Interfaces |
| Phase 5: US3 | 5 | 1 | BMC Interfaces |
| Phase 6: US4 | 7 | 1 | Import |
| Phase 7: US5 | 8 | 2 | Update/Remove |
| Phase 8: Drift/Validation | 9 | 4 | Cross-cutting |
| Phase 9: Backward Compat | 5 | 1 | Legacy Mode |
| Phase 10: Polish | 12 | 6 | Cross-cutting |
| **Total** | **75** | **25** | 5 User Stories |

---

## Notes

- [P] tasks = different files, no dependencies within batch
- [Story] label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests FAIL before implementing (RED phase)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Run `make fmt && make lint` before commits
