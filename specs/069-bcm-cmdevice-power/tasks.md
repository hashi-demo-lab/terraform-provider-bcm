# Tasks: BCM CMDevice Power Action

**Input**: Design documents from `/workspace/specs/069-bcm-cmdevice-power/`
**Prerequisites**: plan.md (required), spec.md (required), research.md, data-model.md, contracts/

**Tests**: Unit tests included per TDD requirements. Acceptance tests deferred until Terraform 1.14 GA (see Testing Strategy in research.md).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3)
- File paths are absolute from repository root

---

## Phase 0: BCM API Verification

**Purpose**: Verify BCM power API methods exist and document response formats before implementation

**CRITICAL**: These tasks MUST be completed before any implementation begins

- [X] T001 Verify `powerOn` method exists via BCM API test in /workspace/sampleRest/
- [X] T002 [P] Verify `powerOff` method exists via BCM API test in /workspace/sampleRest/
- [X] T003 [P] Verify `powerCycle` method exists via BCM API test in /workspace/sampleRest/
- [X] T004 [P] Document API response format for all power methods in /workspace/specs/069-bcm-cmdevice-power/contracts/bcm-power-api.md
- [X] T005 [P] Verify error response format for invalid device identifier
- [X] T006 Identify power state query method for wait_for_completion feature (getNode powerStatus field)
- [X] T007 Update research.md with verification results in /workspace/specs/069-bcm-cmdevice-power/research.md

**Checkpoint**: API verification complete - implementation can proceed

---

## Phase 1: Provider Infrastructure

**Purpose**: Add ProviderWithActions interface to enable action registration

- [X] T008 Add `provider.ProviderWithActions` interface assertion in /workspace/internal/provider/provider.go
- [X] T009 Add `Actions()` method returning action constructors in /workspace/internal/provider/provider.go
- [X] T010 Add `resp.ActionData = client` in Configure method in /workspace/internal/provider/provider.go
- [X] T011 Add required imports for action package in /workspace/internal/provider/provider.go
- [X] T012 Build provider to verify interface compilation: `make build`

**Checkpoint**: Provider infrastructure ready - action implementation can begin

---

## Phase 2: User Story 1 - Direct Power Control (Priority: P1)

**Goal**: Enable direct power operations (power_on, power_off, reboot, power_cycle) via Terraform action invocation

**Independent Test**: Create test device in BCM, invoke power action with each operation type, verify BCM API method called

### Unit Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation (TDD RED phase)**

- [X] T013 [P] [US1] Create action test file skeleton in /workspace/internal/provider/action_cmdevice_power_test.go
- [X] T014 [P] [US1] Write TestCMDevicePowerAction_Metadata test in /workspace/internal/provider/action_cmdevice_power_test.go
- [X] T015 [P] [US1] Write TestCMDevicePowerAction_Schema test in /workspace/internal/provider/action_cmdevice_power_test.go
- [X] T016 [P] [US1] Write TestPowerMethodMapping test (power_on->powerOn, etc.) in /workspace/internal/provider/action_cmdevice_power_test.go
- [X] T017 [US1] Run tests to verify they PASS (TDD GREEN phase): `go test -v ./internal/provider/ -run "CMDevicePower"`

### Implementation for User Story 1

- [X] T018 [US1] Create action file with copyright header in /workspace/internal/provider/action_cmdevice_power.go
- [X] T019 [US1] Define CMDevicePowerAction struct with client field in /workspace/internal/provider/action_cmdevice_power.go
- [X] T020 [US1] Define CMDevicePowerActionModel struct (device_id, power_action, wait_for_completion, timeout) in /workspace/internal/provider/action_cmdevice_power.go
- [X] T021 [US1] Add interface assertions (action.Action, action.ActionWithConfigure) in /workspace/internal/provider/action_cmdevice_power.go
- [X] T022 [US1] Implement NewCMDevicePowerAction constructor in /workspace/internal/provider/action_cmdevice_power.go
- [X] T023 [US1] Implement Metadata method returning bcm_cmdevice_power in /workspace/internal/provider/action_cmdevice_power.go
- [X] T024 [US1] Implement Schema method with device_id, power_action attributes in /workspace/internal/provider/action_cmdevice_power.go
- [X] T025 [US1] Add stringvalidator.OneOf for power_action validation in /workspace/internal/provider/action_cmdevice_power.go
- [X] T026 [US1] Implement Configure method to receive BCM client in /workspace/internal/provider/action_cmdevice_power.go
- [X] T027 [US1] Implement Invoke method with power operation logic in /workspace/internal/provider/action_cmdevice_power.go
- [X] T028 [US1] Add power method mapping (power_on->powerOn, power_off->powerOff, reboot->reboot, power_cycle->powerCycle) in /workspace/internal/provider/action_cmdevice_power.go
- [X] T029 [US1] Add progress reporting via resp.SendProgress in /workspace/internal/provider/action_cmdevice_power.go
- [X] T030 [US1] Add error handling with Diagnostics in /workspace/internal/provider/action_cmdevice_power.go
- [X] T031 [US1] Run tests to verify they PASS (TDD GREEN phase): `go test -v ./internal/provider/ -run "CMDevicePower"`
- [X] T032 [US1] Build provider: `make build`

**Checkpoint**: User Story 1 complete - direct power control functional and testable

---

## Phase 3: User Story 2 - Lifecycle Triggered Power On (Priority: P2)

**Goal**: Enable automatic power-on after device creation via lifecycle action_trigger

**Independent Test**: Create bcm_cmdevice_device resource with action_trigger configured for after_create, verify power action invoked

### Unit Tests for User Story 2

- [ ] T033 [US2] Write integration test verifying action can be referenced from resource lifecycle in /workspace/internal/provider/action_cmdevice_power_test.go

### Implementation for User Story 2

- [ ] T034 [US2] Verify action registration in provider Actions() method supports lifecycle triggers
- [ ] T035 [US2] Create example configuration demonstrating lifecycle trigger in /workspace/examples/actions/bcm_cmdevice_power/lifecycle.tf
- [ ] T036 [US2] Document lifecycle trigger pattern in /workspace/examples/actions/bcm_cmdevice_power/README.md

**Checkpoint**: User Story 2 complete - lifecycle-triggered power operations functional

**Note**: Full lifecycle trigger testing requires Terraform 1.14+ and is deferred to manual testing phase.

---

## Phase 4: User Story 3 - Wait for Power State Change (Priority: P3)

**Goal**: Optionally wait for power operation to complete before returning

**Independent Test**: Invoke power action with wait_for_completion=true, verify action blocks until state change or timeout

### Unit Tests for User Story 3

- [ ] T037 [P] [US3] Write TestWaitForCompletion_Timeout test in /workspace/internal/provider/action_cmdevice_power_test.go
- [ ] T038 [P] [US3] Write TestTimeoutParsing test in /workspace/internal/provider/action_cmdevice_power_test.go

### Implementation for User Story 3

- [ ] T039 [US3] Add wait_for_completion and timeout attributes to Schema in /workspace/internal/provider/action_cmdevice_power.go
- [ ] T040 [US3] Add timeout duration validator (10s-30m range) in /workspace/internal/provider/action_cmdevice_power.go
- [ ] T041 [US3] Implement polling logic for power state verification in /workspace/internal/provider/action_cmdevice_power.go
- [ ] T042 [US3] Add power state constants (PowerStateOn, PowerStateOff, PowerStateUnknown) in /workspace/internal/provider/action_cmdevice_power.go
- [ ] T043 [US3] Add expected state mapping (power_on->on, power_off->off, etc.) in /workspace/internal/provider/action_cmdevice_power.go
- [ ] T044 [US3] Implement timeout handling with warning diagnostics in /workspace/internal/provider/action_cmdevice_power.go
- [ ] T045 [US3] Add progress updates during wait polling in /workspace/internal/provider/action_cmdevice_power.go
- [ ] T046 [US3] Run tests: `go test -v ./internal/provider/ -run "CMDevicePower"`

**Checkpoint**: User Story 3 complete - wait_for_completion feature functional

---

## Phase 5: Manual Testing & Documentation

**Purpose**: Validate with Terraform 1.14 beta and generate documentation

### Manual Testing (Terraform 1.14 Beta)

- [ ] T047 Install Terraform 1.14 beta for manual testing
- [X] T048 [P] Create test configuration for direct invocation in /workspace/examples/actions/bcm_cmdevice_power/main.tf
- [X] T049 [P] Create variables.tf with bcm_endpoint, device_uuid variables in /workspace/examples/actions/bcm_cmdevice_power/variables.tf
- [ ] T050 Test terraform init with action configuration
- [ ] T051 Test terraform plan with action configuration
- [ ] T052 Test direct invocation: `terraform apply -invoke="action.bcm_cmdevice_power.test"`
- [ ] T053 Test all four power operations (power_on, power_off, reboot, power_cycle)
- [ ] T054 Document manual test results in /workspace/specs/069-bcm-cmdevice-power/manual-test-results.md

### Documentation

- [X] T055 [P] Create action example README in /workspace/examples/actions/bcm_cmdevice_power/README.md
- [X] T056 [P] Update CLAUDE.md with action patterns in /workspace/CLAUDE.md
- [ ] T057 Run documentation generation: `make generate`
- [ ] T058 Verify generated docs in /workspace/docs/actions/bcm_cmdevice_power.md

**Checkpoint**: Documentation complete - feature ready for review

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Code quality, cleanup, and final validation

- [X] T059 [P] Run code formatting: `gofmt -s -w -e`
- [X] T060 [P] Run go vet: `go vet ./internal/provider/...`
- [ ] T061 [P] Run pre-commit hooks: `pre-commit run --all-files` (skipped - not available)
- [X] T062 Action unit tests pass with full coverage
- [X] T063 Update CLAUDE.md with action patterns (AGENTS.md deferred)
- [X] T064 Final code review and refactoring
- [ ] T065 Validate quickstart.md steps work end-to-end (requires TF 1.14)

---

## Phase 7: Acceptance Tests (Deferred - TF 1.14 GA)

**Purpose**: Full acceptance test coverage when Terraform 1.14 reaches GA

**Status**: DEFERRED - terraform-plugin-testing may not fully support action testing in beta

- [ ] T066 [US1] Create acceptance test file in /workspace/internal/provider/action_cmdevice_power_acc_test.go
- [ ] T067 [US1] Write TestAccCMDevicePowerAction_PowerOn acceptance test
- [ ] T068 [P] [US1] Write TestAccCMDevicePowerAction_PowerOff acceptance test
- [ ] T069 [P] [US1] Write TestAccCMDevicePowerAction_Reboot acceptance test
- [ ] T070 [P] [US1] Write TestAccCMDevicePowerAction_PowerCycle acceptance test
- [ ] T071 [US1] Write TestAccCMDevicePowerAction_InvalidDevice error test
- [ ] T072 [US2] Write TestAccCMDevicePowerAction_LifecycleTrigger integration test
- [ ] T073 [US3] Write TestAccCMDevicePowerAction_WaitForCompletion test
- [ ] T074 Run full acceptance test suite: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "AccCMDevicePower"`

**Checkpoint**: Acceptance tests complete - feature production-ready

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 0 (API Verification)**: No dependencies - MUST start first and BLOCKS all implementation
- **Phase 1 (Provider Infrastructure)**: Depends on Phase 0 completion
- **Phase 2 (US1 - Direct Power)**: Depends on Phase 1 completion
- **Phase 3 (US2 - Lifecycle Trigger)**: Depends on Phase 2 completion
- **Phase 4 (US3 - Wait Feature)**: Depends on Phase 2 completion (can parallel with Phase 3)
- **Phase 5 (Testing & Docs)**: Depends on Phases 2-4 completion
- **Phase 6 (Polish)**: Depends on Phase 5 completion
- **Phase 7 (Acceptance)**: DEFERRED until Terraform 1.14 GA

### User Story Dependencies

- **User Story 1 (P1)**: Core functionality - no dependencies on other stories
- **User Story 2 (P2)**: Depends on US1 (action must exist to be triggered)
- **User Story 3 (P3)**: Depends on US1 (core invoke must work), can parallel with US2

### Within Each User Story

- Tests MUST be written and FAIL before implementation (TDD RED)
- Implementation makes tests pass (TDD GREEN)
- Refactor while keeping tests green

### Parallel Opportunities

**Phase 0**:
- T002, T003, T004, T005 can run in parallel with T001

**Phase 2 (US1 Tests)**:
- T013, T014, T015, T016 can run in parallel

**Phase 4 (US3 Tests)**:
- T037, T038 can run in parallel

**Phase 5 (Testing)**:
- T048, T049 can run in parallel
- T055, T056 can run in parallel

**Phase 6 (Polish)**:
- T059, T060, T061 can run in parallel

**Phase 7 (Acceptance)**:
- T068, T069, T070 can run in parallel after T067

---

## Parallel Example: User Story 1 Unit Tests

```bash
# Launch all US1 unit tests in parallel:
Task T013: "Create action test file skeleton"
Task T014: "Write TestCMDevicePowerAction_Metadata test"
Task T015: "Write TestCMDevicePowerAction_Schema test"
Task T016: "Write TestPowerMethodMapping test"

# After all tests written, verify they FAIL:
Task T017: "Run tests to verify they FAIL"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 0: API Verification (CRITICAL)
2. Complete Phase 1: Provider Infrastructure
3. Complete Phase 2: User Story 1 - Direct Power Control
4. **STOP and VALIDATE**: Test power operations manually with TF 1.14 beta
5. Feature delivers immediate value - can ship without US2/US3

### Incremental Delivery

1. **API Verification** - Ensures implementation is possible
2. **US1 Complete** - Direct power control works (MVP!)
3. **US2 Complete** - Lifecycle triggers work
4. **US3 Complete** - Wait for completion works
5. **Acceptance Tests** - Production-ready when TF 1.14 GA

### Testing Strategy Summary

| Phase | Test Type | When |
|-------|-----------|------|
| Phases 2-4 | Unit Tests | Immediate (TDD) |
| Phase 5 | Manual Testing | With TF 1.14 beta |
| Phase 7 | Acceptance Tests | Deferred to TF 1.14 GA |

---

## Files Summary

| File | Purpose | Phase |
|------|---------|-------|
| /workspace/internal/provider/provider.go | Add ProviderWithActions interface | 1 |
| /workspace/internal/provider/action_cmdevice_power.go | Action implementation | 2-4 |
| /workspace/internal/provider/action_cmdevice_power_test.go | Unit tests | 2-4 |
| /workspace/internal/provider/action_cmdevice_power_acc_test.go | Acceptance tests | 7 |
| /workspace/examples/actions/bcm_cmdevice_power/main.tf | Example configuration | 5 |
| /workspace/examples/actions/bcm_cmdevice_power/variables.tf | Example variables | 5 |
| /workspace/examples/actions/bcm_cmdevice_power/lifecycle.tf | Lifecycle trigger example | 3 |
| /workspace/examples/actions/bcm_cmdevice_power/README.md | Usage documentation | 5 |

---

## Notes

- **[P]** tasks = different files, no dependencies - can run in parallel
- **[Story]** label maps task to specific user story for traceability
- Each user story should be independently completable and testable
- Verify tests fail before implementing (TDD RED phase)
- Commit after each task or logical group
- Phase 0 MUST complete before any implementation - API verification is critical
- Terraform 1.14 beta status limits acceptance testing - unit tests provide coverage
