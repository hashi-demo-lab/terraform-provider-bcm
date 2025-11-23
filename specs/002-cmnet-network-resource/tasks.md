# Tasks: BCM Network Resource Management

**Feature**: `bcm_cmnet_network` resource for full CRUD lifecycle management of BCM network configurations

**Input**: Design documents from `/workspace/specs/002-cmnet-network-resource/`
**Prerequisites**: plan.md, spec.md, contracts/, data-model.md

**Organization**: Tasks organized by user story to enable independent implementation and testing per story

**TDD Workflow**: RED (write failing tests) → GREEN (minimal implementation) → REFACTOR (production quality)

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story label (US1, US2, US3, US4, US5, US6)
- All file paths are absolute

---

## Phase 0: API Research & Exploration

**Purpose**: Validate BCM CMNet API methods and resolve design unknowns before TDD implementation

- [ ] T001 [P] Create API exploration script at `/workspace/sampleRest/explore_network_crud.py`
- [ ] T002 Run API exploration to verify `getNetwork(uuid)` supports args parameter for direct lookup
- [ ] T003 [P] Test CIDR parsing with Go net.ParseCIDR to confirm baseAddress/netmaskBits conversion works
- [ ] T004 Verify BCM CMNet API methods: addNetwork, getNetwork, updateNetwork, removeNetwork with sample calls
- [ ] T005 Test force parameter behavior for removeNetwork (with/without active node assignments)
- [ ] T006 Inspect BCM API response for VLAN field mapping (confirm if vlanId or vlan exists)
- [ ] T007 Document research findings in `/workspace/specs/002-cmnet-network-resource/research.md`

**Checkpoint**: API contracts confirmed, CIDR parsing validated, VLAN support determined

---

## Phase 1: Design Artifacts & Data Modeling

**Purpose**: Document data model, API contracts, and developer quick start before implementation

**Dependencies**: Phase 0 complete (research.md exists)

- [ ] T008 Create data model documentation at `/workspace/specs/002-cmnet-network-resource/data-model.md`
- [ ] T009 [P] Create API contracts documentation at `/workspace/specs/002-cmnet-network-resource/contracts/bcm-cmnet-api.md`
- [ ] T010 [P] Create developer quick start guide at `/workspace/specs/002-cmnet-network-resource/quickstart.md`
- [ ] T011 Update agent context with `update-agent-context.sh copilot` for BCM CMNet patterns

**Checkpoint**: Design artifacts complete, ready for TDD implementation

---

## Phase 2: TDD RED - Write Failing Acceptance Tests

**Purpose**: Write all acceptance tests FIRST, ensure they FAIL before any implementation exists

**Dependencies**: Phase 1 complete

### Test Infrastructure Setup

- [ ] T012 Create test file at `/workspace/internal/provider/resource_cmnet_network_test.go`
- [ ] T013 Implement `testAccCheckCMNetNetworkDestroy` with enhanced error messages using `verifyResourceDeleted`
- [ ] T014 [P] Create test config helper `testAccCMNetNetworkConfigBasic(name)` with provider block
- [ ] T015 [P] Create test config helper `testAccCMNetNetworkConfigSubnet(name, subnet, gateway, mtu)`
- [ ] T016 [P] Create test config helper `testAccCMNetNetworkConfigDHCP(name, subnet, rangeStart, rangeEnd)`
- [ ] T017 [P] Create test config helper `testAccCMNetNetworkConfigComplete(name)` with all optional attributes
- [ ] T018 [P] Create test config helper `testAccCMNetNetworkConfigUpdate(name)` for update scenarios

### User Story 1: Basic Network Creation (P1) - Tests

- [ ] T019 [US1] Write `TestAccCMNetNetwork_Basic` - create with name only, verify BCM defaults
- [ ] T020 [US1] Add statecheck.ExpectKnownValue for name, uuid, id (uses knownvalue.StringExact and NotNull)
- [ ] T021 [US1] Add ID consistency tracking with `compareID.AddStateValue` for create step

### User Story 2: Network Update Management (P2) - Tests

- [ ] T022 [US2] Write `TestAccCMNetNetwork_Update` - modify MTU from default to 9000, add gateway
- [ ] T023 [US2] Add statecheck.ExpectKnownValue for mtu (knownvalue.Int64Exact) and gateway after update
- [ ] T024 [US2] Add ID consistency check to verify UUID unchanged across update

### User Story 3: DHCP Configuration (P2) - Tests

- [ ] T025 [US3] Write `TestAccCMNetNetwork_DHCP` - enable DHCP with range, verify dhcp_enabled computed
- [ ] T026 [US3] Add statecheck.ExpectKnownValue for dhcp_enabled (knownvalue.Bool), dhcp_range_start, dhcp_range_end
- [ ] T027 [US3] Write `TestAccCMNetNetwork_UpdateDHCP` - toggle DHCP on/off across three steps

### User Story 4: VLAN Segmentation (P3) - Tests (conditional on research)

- [ ] T028 [US4] Write `TestAccCMNetNetwork_VLAN` - configure vlan_id if BCM API supports it
- [ ] T029 [US4] Add statecheck.ExpectKnownValue for vlan_id (knownvalue.Int64Exact) if field exists

### User Story 5: Import Existing Networks (P2) - Tests

- [ ] T030 [US5] Write `TestAccCMNetNetwork_Import` - import by UUID, verify ImportStateVerify true
- [ ] T031 [US5] Add ID consistency tracking across import step with compareID.AddStateValue

### User Story 6: Drift Detection (P3) - Tests

- [ ] T032 [US6] Write `TestAccCMNetNetwork_DriftDetection` - modify MTU externally via API
- [ ] T033 [US6] Implement PreConfig func to modify network externally using createTestBCMClient
- [ ] T034 [US6] Add plancheck.ExpectNonEmptyPlan to detect drift in Step 2
- [ ] T035 [US6] Verify Terraform restores config in Step 3

### Additional Test Coverage

- [ ] T036 [P] Write `TestAccCMNetNetwork_Subnet` - CIDR parsing to baseAddress/netmaskBits
- [ ] T037 [P] Write `TestAccCMNetNetwork_CompleteConfig` - all optional attributes configured
- [ ] T038 [P] Write `TestAccCMNetNetwork_IdempotencyAfterCreate` - empty plan after create (plancheck.ExpectEmptyPlan)
- [ ] T039 [P] Write `TestAccCMNetNetwork_IdempotencyAfterUpdate` - empty plan after update

### Verify RED Phase Complete

- [ ] T040 Run all tests with `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMNetNetwork` - expect 11 failures

**Checkpoint**: All 11 tests written and FAILING (no implementation exists yet)

---

## Phase 3: TDD GREEN - Minimal CRUD Implementation

**Purpose**: Write minimal code to make tests PASS (hardcoded values initially)

**Dependencies**: Phase 2 complete (all tests failing)

### Resource Skeleton

- [ ] T041 Create resource file at `/workspace/internal/provider/resource_cmnet_network.go`
- [ ] T042 Define `CMNetNetworkResource` struct with client field
- [ ] T043 Define `CMNetNetworkResourceModel` struct with all attributes (ID, UUID, Name, Subnet, DHCP fields, etc.)
- [ ] T044 Implement `NewCMNetNetworkResource()` factory function
- [ ] T045 Implement `Metadata()` method returning "bcm_cmnet_network" type name

### Schema Definition (Minimal)

- [ ] T046 Implement `Schema()` method with required attribute: name (stringvalidator for uniqueness)
- [ ] T047 [P] Add optional attributes: subnet (with CIDR regex validator), gateway, mtu, domain_name
- [ ] T048 [P] Add optional DHCP attributes: dhcp_range_start, dhcp_range_end
- [ ] T049 [P] Add computed attributes: id, uuid, dhcp_enabled, base_address, netmask_bits
- [ ] T050 [P] Add BCM entity computed attributes: base_type, child_type, revision, modified, to_be_removed
- [ ] T051 Add conditional VLAN attribute: vlan_id (int64) if research confirmed support

### Minimal CRUD Methods (Hardcoded)

- [ ] T052 Implement minimal `Create()` - set hardcoded UUID, dhcp_enabled=false, base_type="Network"
- [ ] T053 Implement minimal `Read()` - return existing state unchanged (no API call)
- [ ] T054 Implement minimal `Update()` - accept plan values, set to state (no API call)
- [ ] T055 Implement minimal `Delete()` - no-op implementation (no API call)
- [ ] T056 Implement `ImportState()` using `resource.ImportStatePassthroughID`

### Register Resource

- [ ] T057 Add `NewCMNetNetworkResource` to provider.go Resources() method

### Verify GREEN Phase Partial

- [ ] T058 Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMNetNetwork_Basic` - expect different failures (not "resource not found")

**Checkpoint**: Resource registered, tests run but fail validation (wrong values)

---

## Phase 4: TDD REFACTOR - Full CRUD Implementation

**Purpose**: Implement production-quality CRUD with real BCM API integration

**Dependencies**: Phase 3 complete (minimal implementation exists)

### Helper Functions

- [ ] T059 [P] Implement `parseCIDR(cidr string)` helper - returns baseAddress, netmaskBits using net.ParseCIDR
- [ ] T060 [P] Implement `formatCIDR(baseAddress, netmaskBits)` helper - reconstructs CIDR notation
- [ ] T061 [P] Implement `isDHCPEnabled(rangeStart, rangeEnd string)` helper - derives dhcp_enabled logic
- [ ] T062 Implement `buildNetworkAPIEntity(ctx, data)` - converts Terraform model to BCM API entity with baseType/childType/modified/to_be_removed
- [ ] T063 Implement `mapNetworkAPIResponseToState(ctx, apiData, data)` - maps BCM response to Terraform state using getStringValue/getInt64Value/getBoolValue helpers

### Full Create Implementation

- [ ] T064 Refactor `Create()` - call buildNetworkAPIEntity to construct entity
- [ ] T065 Add BCM API call: `client.CallJSONRPC(ctx, "cmnet", "addNetwork", entity, false)`
- [ ] T066 Parse JSON response and unmarshal to map[string]interface{}
- [ ] T067 Call mapNetworkAPIResponseToState to populate data model
- [ ] T068 Add error handling for API failures with actionable messages
- [ ] T069 Add tflog.Trace for successful create with UUID and name

### Full Read Implementation

- [ ] T070 Refactor `Read()` - call `client.CallJSONRPC(ctx, "cmnet", "getNetwork", data.UUID.ValueString())`
- [ ] T071 Handle 404 errors by calling `resp.State.RemoveResource(ctx)` for drift detection
- [ ] T072 Parse response and call mapNetworkAPIResponseToState
- [ ] T073 Reconstruct subnet attribute from baseAddress and netmaskBits using formatCIDR

### Full Update Implementation

- [ ] T074 Refactor `Update()` - get plan data and build entity with UUID and revision
- [ ] T075 Call `client.CallJSONRPC(ctx, "cmnet", "updateNetwork", entity, false)`
- [ ] T076 Parse response and map to state
- [ ] T077 Add error handling for revision conflicts

### Full Delete Implementation

- [ ] T078 Refactor `Delete()` - call `client.CallJSONRPC(ctx, "cmnet", "removeNetwork", data.UUID.ValueString(), false)`
- [ ] T079 Add error handling with message about force parameter if dependencies exist
- [ ] T080 Add tflog.Trace for successful deletion

### CIDR and DHCP Logic Integration

- [ ] T081 In buildNetworkAPIEntity: parse subnet attribute if present, set baseAddress/netmaskBits
- [ ] T082 In buildNetworkAPIEntity: map dhcp_range_start/end to dynamicRangeStart/End
- [ ] T083 In mapNetworkAPIResponseToState: derive dhcp_enabled using isDHCPEnabled helper
- [ ] T084 In mapNetworkAPIResponseToState: reconstruct subnet from baseAddress/netmaskBits

### Verify REFACTOR Phase Complete

- [ ] T085 Run all tests: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetwork`
- [ ] T086 Verify all 11 tests PASS with 100% success rate

**Checkpoint**: All acceptance tests passing, full CRUD implementation complete

---

## Phase 5: Examples & Documentation

**Purpose**: Create example configurations and auto-generate documentation

**Dependencies**: Phase 4 complete (all tests passing)

### Example Configurations

- [ ] T087 Create directory `/workspace/examples/resources/bcm_cmnet_network/`
- [ ] T088 [P] Create `basic.tf` - network with name only
- [ ] T089 [P] Create `subnet.tf` - network with CIDR, gateway, MTU 9000, notes
- [ ] T090 [P] Create `dhcp.tf` - network with DHCP enabled and IP range
- [ ] T091 [P] Create `complete.tf` - all optional attributes configured
- [ ] T092 Create `vlan.tf` - network with VLAN ID (if supported from research)

### Documentation Generation

- [ ] T093 Run `make generate` to auto-generate docs/resources/cmnet_network.md
- [ ] T094 Verify examples formatted with `terraform fmt -check examples/resources/bcm_cmnet_network/`
- [ ] T095 Verify copyright headers added to all new files

### Validation

- [ ] T096 Manually review `/workspace/docs/resources/cmnet_network.md` for accuracy
- [ ] T097 Validate example configs: `cd examples/resources/bcm_cmnet_network && terraform init && terraform validate`

**Checkpoint**: Documentation complete, examples validated

---

## Phase 6: Integration Testing & Quality Checks

**Purpose**: Real-world testing with BCM cluster and code quality validation

**Dependencies**: Phase 5 complete

### Real-World Testing Scenarios

- [ ] T098 Test basic network creation: apply basic.tf, verify in BCM UI, destroy
- [ ] T099 Test subnet configuration: apply subnet.tf, verify CIDR parsed correctly, destroy
- [ ] T100 Test DHCP toggle: create without DHCP, update to enable, update to disable, destroy
- [ ] T101 Test import: manually create network via BCM UI, import with terraform import, verify state match
- [ ] T102 Test drift detection: create via Terraform, modify MTU via BCM UI, run plan (should detect drift), apply (should restore)

### Code Quality

- [ ] T103 Run `make fmt` - ensure all code formatted correctly
- [ ] T104 Run `make lint` - verify no golangci-lint warnings
- [ ] T105 Run `pre-commit run --all-files` - ensure all hooks pass
- [ ] T106 Verify test coverage: all CRUD operations, import, drift detection, idempotency

### Final Validation Checklist

- [ ] T107 Verify all 11 acceptance tests pass consistently (run 3 times)
- [ ] T108 Verify network creation completes in <30 seconds (SC-001)
- [ ] T109 Verify import works for 100% of existing networks (SC-003)
- [ ] T110 Verify drift detection catches external changes (SC-004)
- [ ] T111 Verify idempotency - repeated apply shows empty plan (SC-005)
- [ ] T112 Verify DHCP changes apply within 60 seconds (SC-007)

**Checkpoint**: Production ready, all quality checks passed

---

## Dependencies & Execution Order

### Phase Dependencies

1. **Phase 0 (Research)**: No dependencies - start immediately
2. **Phase 1 (Design)**: Depends on Phase 0 (research.md must exist)
3. **Phase 2 (RED)**: Depends on Phase 1 (data-model.md, contracts/ must exist)
4. **Phase 3 (GREEN)**: Depends on Phase 2 (all tests written and failing)
5. **Phase 4 (REFACTOR)**: Depends on Phase 3 (minimal implementation exists)
6. **Phase 5 (Examples)**: Depends on Phase 4 (all tests passing)
7. **Phase 6 (Integration)**: Depends on Phase 5 (docs generated)

### User Story Coverage by Phase

- **Phase 2-4** cover all 6 user stories through tests and implementation:
  - **US1 (P1)**: Basic network creation - T019-T021 (tests), T041-T069 (implementation)
  - **US2 (P2)**: Update management - T022-T024 (tests), T074-T077 (implementation)
  - **US3 (P2)**: DHCP configuration - T025-T027 (tests), T081-T084 (implementation)
  - **US4 (P3)**: VLAN segmentation - T028-T029 (tests), T051 (implementation)
  - **US5 (P2)**: Import networks - T030-T031 (tests), T056 (implementation)
  - **US6 (P3)**: Drift detection - T032-T035 (tests), T071 (implementation)

### Critical Path (Sequential Tasks)

1. T001-T007 (Research) → T008-T011 (Design)
2. T012-T018 (Test infrastructure) → T019-T039 (Write all tests)
3. T040 (Verify RED) → T041-T058 (Minimal implementation)
4. T058 (Verify GREEN partial) → T059-T086 (Full implementation)
5. T086 (Verify tests pass) → T087-T097 (Examples & docs)
6. T097 (Validate examples) → T098-T112 (Integration testing)

### Parallel Opportunities

#### Phase 0 (Research)
- T001, T003 can run in parallel (different exploration tasks)

#### Phase 1 (Design)
- T009, T010 can run in parallel (contracts vs quickstart docs)

#### Phase 2 (RED - Test Helpers)
- T014-T018 can all run in parallel (different test config helpers)

#### Phase 2 (RED - Test Scenarios)
- T019-T039 can run in parallel by user story:
  - Developer A: T019-T021 (US1 tests)
  - Developer B: T022-T024 (US2 tests)
  - Developer C: T025-T027 (US3 tests)
  - Developer D: T028-T029, T030-T031, T032-T035 (US4, US5, US6 tests)
  - Developer E: T036-T039 (additional coverage)

#### Phase 3 (GREEN - Schema)
- T047-T051 can run in parallel (different attribute groups)

#### Phase 4 (REFACTOR - Helpers)
- T059-T063 can run in parallel (independent helper functions)

#### Phase 5 (Examples)
- T088-T092 can run in parallel (different example files)

#### Phase 6 (Integration)
- T098-T102 can run in parallel (independent test scenarios)

---

## Parallel Execution Example

### Phase 2 (RED) - Write Tests Concurrently

```bash
# Launch all test helpers in parallel:
Task T014: testAccCMNetNetworkConfigBasic
Task T015: testAccCMNetNetworkConfigSubnet
Task T016: testAccCMNetNetworkConfigDHCP
Task T017: testAccCMNetNetworkConfigComplete
Task T018: testAccCMNetNetworkConfigUpdate

# Launch all user story tests in parallel:
Task T019-T021: US1 Basic creation tests
Task T022-T024: US2 Update tests
Task T025-T027: US3 DHCP tests
Task T028-T029: US4 VLAN tests (if applicable)
Task T030-T031: US5 Import tests
Task T032-T035: US6 Drift detection tests
```

### Phase 4 (REFACTOR) - Implement Helpers Concurrently

```bash
# Launch all helper functions in parallel:
Task T059: parseCIDR
Task T060: formatCIDR
Task T061: isDHCPEnabled
Task T062: buildNetworkAPIEntity
Task T063: mapNetworkAPIResponseToState
```

### Phase 5 (Examples) - Create Examples Concurrently

```bash
# Launch all example files in parallel:
Task T088: basic.tf
Task T089: subnet.tf
Task T090: dhcp.tf
Task T091: complete.tf
Task T092: vlan.tf
```

---

## Implementation Strategy

### MVP First (Phase 0-4, User Story 1 Only)

For fastest time-to-value, focus on User Story 1 (P1) basic network creation:

1. Complete Phase 0: Research (T001-T007)
2. Complete Phase 1: Design (T008-T011)
3. Complete Phase 2 (RED): Write US1 tests only (T012-T021)
4. Complete Phase 3 (GREEN): Minimal implementation (T041-T058)
5. Complete Phase 4 (REFACTOR): US1 implementation (T059-T073, partial T081-T084)
6. **STOP and VALIDATE**: Run TestAccCMNetNetwork_Basic, verify it passes
7. Complete Phase 5: Create basic.tf example only
8. Deploy/demo basic network creation capability

### Incremental Delivery (All User Stories)

1. **Foundation**: Phase 0-1 (Research + Design)
2. **MVP (US1)**: Phase 2-4 for US1 → Test → Deploy
3. **Iteration 1 (US2, US3, US5)**: Add P2 stories → Test → Deploy
4. **Iteration 2 (US4, US6)**: Add P3 stories → Test → Deploy
5. **Polish**: Phase 5-6 → Final documentation and validation

### Full TDD Cycle (Recommended)

1. **Phase 0**: Research (resolve unknowns)
2. **Phase 1**: Design (document contracts)
3. **Phase 2 (RED)**: Write ALL tests, ensure they FAIL
4. **Phase 3 (GREEN)**: Minimal implementation (hardcoded)
5. **Phase 4 (REFACTOR)**: Full implementation (BCM API integration)
6. **Phase 5**: Examples and documentation
7. **Phase 6**: Integration testing and validation

This approach ensures:
- Tests written before implementation (true TDD)
- All user stories validated independently
- Production-quality code with 100% acceptance test coverage
- Complete documentation auto-generated

---

## Task Summary

**Total Tasks**: 112

**By Phase**:
- Phase 0 (Research): 7 tasks
- Phase 1 (Design): 4 tasks
- Phase 2 (RED): 29 tasks (including 21 test scenarios)
- Phase 3 (GREEN): 18 tasks
- Phase 4 (REFACTOR): 28 tasks
- Phase 5 (Examples): 11 tasks
- Phase 6 (Integration): 15 tasks

**By User Story**:
- Setup/Infrastructure: 11 tasks
- US1 (Basic Creation - P1): 18 tasks
- US2 (Update Management - P2): 8 tasks
- US3 (DHCP Configuration - P2): 10 tasks
- US4 (VLAN Segmentation - P3): 2 tasks (conditional)
- US5 (Import Networks - P2): 3 tasks
- US6 (Drift Detection - P3): 5 tasks
- Cross-cutting (Examples, Docs, QA): 26 tasks

**Parallel Opportunities**: 35 tasks marked [P] can run concurrently

**MVP Scope (US1 only)**: 48 tasks (Phase 0-4 US1 + basic example)

**Full Implementation**: 112 tasks (all phases, all user stories)

---

## Notes

- Tasks follow strict TDD RED-GREEN-REFACTOR cycle
- All test tasks (T012-T040) MUST complete before implementation tasks (T041+)
- CIDR parsing validated in Phase 0 before schema design
- VLAN support conditional on Phase 0 research findings
- Modern testing patterns required: statecheck, plancheck, knownvalue
- Import functionality uses resource.ImportStatePassthroughID pattern
- Drift detection uses PreConfig to modify resources externally
- All examples must validate with terraform init/validate
- Documentation auto-generated via make generate (never edit manually)
- Each user story independently testable and deliverable
