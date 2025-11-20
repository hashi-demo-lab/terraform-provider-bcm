# Tasks: BCM CMDevice Nodes Data Source

**Input**: Design documents from `/workspace/specs/002-bcm-cmdevice-nodes-datasource/`
**Prerequisites**: spec.md, plan.md, data-model.md, quickstart.md, research.md

**Feature**: Implement `bcm_cmdevice_nodes` Terraform data source for querying BCM cluster nodes
**TDD Approach**: RED-GREEN-REFACTOR cycle with acceptance tests
**Tests**: Required (TDD workflow with acceptance tests)

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story mapping (US1 = Node Discovery, US2 = Filtering, US3 = Network/Roles)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Project Prerequisites)

**Purpose**: Verify project structure and environment setup

- [X] T001 Verify Go environment (1.24+) and Terraform CLI (1.5+) installed
- [X] T002 Verify access to BCM cluster at https://172.21.15.254:8081 with credentials
- [X] T003 [P] Run existing provider tests to confirm environment: `make test`
- [X] T004 [P] Set environment variables for acceptance testing (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD, BCM_INSECURE, TF_ACC)

**Checkpoint**: Environment ready for TDD implementation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core test infrastructure and provider registration - MUST complete before user stories

**⚠️ CRITICAL**: No user story implementation can begin until this phase is complete

- [X] T005 Review existing data source pattern in /workspace/internal/provider/data_source_cmpart_softwareimages.go
- [X] T006 Review existing BCMClient implementation in /workspace/internal/provider/bcm_client.go
- [X] T007 Verify existing helper functions (getStringValue, getBoolValue, getInt64Value) in provider code

**Checkpoint**: Foundation ready - TDD cycle can now begin

---

## Phase 3: User Story 1 - Node Discovery (Priority: P1) 🎯 MVP

**Goal**: Query all BCM cluster nodes with basic attributes (hostname, UUID, MAC, types)

**Independent Test**: Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Basic` - should retrieve all nodes successfully

### RED Phase: Write Failing Acceptance Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T008 [US1] Create test file /workspace/internal/provider/data_source_cmdevice_nodes_test.go with package declaration and imports
- [X] T009 [US1] Write TestAccCMDeviceNodesDataSource_Basic test function - query all nodes without filters, verify id and nodes.# attributes exist
- [X] T010 [US1] Add test configuration constant testAccCMDeviceNodesDataSourceConfig_basic with minimal data source block
- [X] T011 [US1] Run acceptance test to verify it fails with "data source not registered" error: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Basic`

**Checkpoint RED Phase**: Tests fail as expected - data source doesn't exist yet

### GREEN Phase: Minimal Implementation for User Story 1

- [X] T012 [US1] Create /workspace/internal/provider/data_source_cmdevice_nodes.go with copyright header and package declaration
- [X] T013 [US1] Define CMDeviceNodesDataSource struct with client *BCMClient field
- [X] T014 [US1] Define CMDeviceNodesDataSourceModel struct with ID, Filter (optional), and Nodes fields
- [X] T015 [P] [US1] Define NodeModel struct with basic identity fields (ID, UUID, Hostname, BaseType, ChildType, MAC, CreationTime)
- [X] T016 [P] [US1] Define FilterModel struct with NodeType, CategoryUUID, and HostnamePattern fields
- [X] T017 [US1] Implement NewCMDeviceNodesDataSource() constructor function
- [X] T018 [US1] Implement Metadata() method returning "bcm_cmdevice_nodes" type name
- [X] T019 [US1] Implement Configure() method to receive BCMClient from provider
- [X] T020 [US1] Implement minimal Schema() method with id and nodes attributes (basic node fields only - no nested attributes yet)
- [X] T021 [US1] Implement minimal Read() method with hardcoded test data (2 nodes with ID, UUID, Hostname)
- [X] T022 [US1] Register data source in /workspace/internal/provider/provider.go DataSources() method
- [X] T023 [US1] Run acceptance test to verify it passes with hardcoded data: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Basic`

**Checkpoint GREEN Phase**: Tests pass with minimal hardcoded implementation

### REFACTOR Phase: Production-Ready Implementation for User Story 1

- [X] T024 [US1] Add remaining NodeModel fields (Category, Partition, PowerControl, AuthenticationService, ProvisioningTransport, Modified, ToBeRemoved)
- [X] T025 [US1] Update Schema() to include all node attributes with MarkdownDescription
- [X] T026 [US1] Replace hardcoded Read() implementation with real API call to d.client.CallJSONRPC(ctx, "cmdevice", "getNodes")
- [X] T027 [US1] Add JSON response parsing with error handling for unmarshal failures
- [X] T028 [US1] Implement mapAPIToNode() helper function for API-to-model conversion using existing getStringValue/getBoolValue/getInt64Value helpers
- [X] T029 [US1] Add comprehensive error messages for API failures (auth, network, parse errors)
- [X] T030 [US1] Add tflog.Debug logging for successful node retrieval with count
- [X] T031 [US1] Run acceptance test with real API to verify full implementation: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Basic`
- [X] T032 [US1] Run code quality checks: `make fmt && golangci-lint run internal/provider/data_source_cmdevice_nodes.go`

**Checkpoint US1 Complete**: Basic node discovery working with all node attributes from real API

---

## Phase 4: User Story 2 - Node Filtering (Priority: P2)

**Goal**: Filter nodes by type, category UUID, and hostname pattern

**Independent Test**: Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Filter` - should filter nodes correctly

### RED Phase: Write Failing Acceptance Tests for User Story 2

- [ ] T033 [US2] Write TestAccCMDeviceNodesDataSource_FilterByType test function - filter by node_type = "PhysicalNode"
- [ ] T034 [US2] Add test configuration constant testAccCMDeviceNodesDataSourceConfig_filterType with filter block
- [ ] T035 [P] [US2] Write TestAccCMDeviceNodesDataSource_FilterByCategory test function - filter by category_uuid
- [ ] T036 [P] [US2] Add test configuration constant testAccCMDeviceNodesDataSourceConfig_filterCategory with category filter
- [ ] T037 [P] [US2] Write TestAccCMDeviceNodesDataSource_FilterByHostname test function - filter by hostname_pattern = "node"
- [ ] T038 [P] [US2] Add test configuration constant testAccCMDeviceNodesDataSourceConfig_filterHostname with hostname filter
- [ ] T039 [US2] Run all filter tests to verify they fail: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Filter`

**Checkpoint RED Phase**: Filter tests fail - filtering not implemented yet

### GREEN Phase: Minimal Filtering Implementation for User Story 2

- [ ] T040 [US2] Add filter block to Schema() in data_source_cmdevice_nodes.go with SingleNestedBlock containing node_type, category_uuid, hostname_pattern attributes
- [ ] T041 [US2] Update Read() method to parse filter configuration from req.Config.Get()
- [ ] T042 [US2] Implement matchesFilter() helper function with exact match for node_type and category_uuid
- [ ] T043 [US2] Add substring matching logic for hostname_pattern (case-insensitive using strings.ToLower and strings.Contains)
- [ ] T044 [US2] Apply filtering in Read() method before appending nodes to state
- [ ] T045 [US2] Run filter tests to verify they pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Filter`

**Checkpoint GREEN Phase**: Filtering tests pass with real API data

### REFACTOR Phase: Optimize Filtering Logic for User Story 2

- [ ] T046 [US2] Extract filter logic into separate filterNodes() helper function for better testability
- [ ] T047 [US2] Add null checking for filter fields (handle IsNull() and IsUnknown())
- [ ] T048 [US2] Add logging for filter application: `tflog.Debug(ctx, "Filtering nodes", map[string]interface{}{"total": total, "filtered": filtered})`
- [ ] T049 [US2] Run all User Story 1 and 2 tests together to verify no regression: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource`

**Checkpoint US2 Complete**: Node filtering working for all three filter types

---

## Phase 5: User Story 3 - Network Interfaces and Roles (Priority: P3)

**Goal**: Expose nested network interfaces and roles for each node

**Independent Test**: Verify nodes[0].interfaces and nodes[0].roles arrays are populated and accessible in Terraform

### RED Phase: Write Failing Acceptance Tests for User Story 3

- [ ] T050 [US3] Write TestAccCMDeviceNodesDataSource_NestedAttributes test function - verify interfaces and roles arrays exist
- [ ] T051 [US3] Add test checks for nodes.0.interfaces.# and nodes.0.roles.# attributes
- [ ] T052 [US3] Add test checks for specific interface fields (name, mac, ip) and role fields (name, uuid)
- [ ] T053 [US3] Run nested attributes test to verify it fails: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Nested`
- [ ] T053b [US-Error] Write TestAccCMDeviceNodesDataSource_ErrorHandling test function with error scenarios
  - Test authentication failure (401) by using invalid provider credentials
  - Test API service error (400) by mocking invalid service call if possible
  - Test network timeout scenario
  - Use `ExpectError` in test configuration to verify error messages are clear
  - Verify error messages include actionable information (e.g., "authentication failed", "connection timeout")

**Checkpoint RED Phase**: Nested attribute tests fail - schema doesn't include interfaces/roles yet

### GREEN Phase: Nested Attributes Implementation for User Story 3

- [ ] T054 [P] [US3] Define NetworkInterfaceModel struct with all interface fields (Name, MAC, IP, IPv6IP, DHCP, Network, BaseType, ChildType, CardType, Bootable, StartIf)
- [ ] T055 [P] [US3] Define RoleModel struct with all role fields (UUID, Name, BaseType, ChildType, AddServices)
- [ ] T056 [US3] Add Interfaces []NetworkInterfaceModel and Roles []RoleModel fields to NodeModel struct
- [ ] T057 [US3] Add interfaces ListNestedAttribute to node schema with all interface attributes
- [ ] T058 [US3] Add roles ListNestedAttribute to node schema with all role attributes
- [ ] T059 [US3] Implement mapInterfaces() helper function to convert API interfaces array to []NetworkInterfaceModel
- [ ] T060 [US3] Implement mapRoles() helper function to convert API roles array to []RoleModel
- [ ] T061 [US3] Update mapAPIToNode() to call mapInterfaces() and mapRoles() for nested arrays
- [ ] T062 [US3] Handle empty arrays gracefully (return []NetworkInterfaceModel{} and []RoleModel{} for null/missing data)
- [ ] T063 [US3] Run nested attributes test to verify it passes: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Nested`

**Checkpoint GREEN Phase**: Nested attributes tests pass with real API data

### REFACTOR Phase: Polish Nested Attributes for User Story 3

- [ ] T064 [US3] Add comprehensive MarkdownDescription to all interface and role schema attributes
- [ ] T065 [US3] Add null safety checks in mapInterfaces() and mapRoles() for type assertions
- [ ] T066 [US3] Add logging for nested attribute counts: `tflog.Debug(ctx, "Mapped node", map[string]interface{}{"hostname": hostname, "interfaces": len(interfaces), "roles": len(roles)})`
- [ ] T067 [US3] Run full test suite to verify all user stories work together: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource`

**Checkpoint US3 Complete**: Full node data including nested interfaces and roles exposed in data source

---

## Phase 6: Documentation and Examples

**Purpose**: Generate provider documentation and create example configurations

- [X] T068 [P] Create /workspace/examples/data-sources/bcm_cmdevice_nodes/ directory
- [X] T069 [P] Create /workspace/examples/data-sources/bcm_cmdevice_nodes/data-source.tf with basic query all nodes example
- [X] T070 [P] Create /workspace/examples/data-sources/bcm_cmdevice_nodes/filter_by_type.tf with node type filter example
- [X] T071 [P] Create /workspace/examples/data-sources/bcm_cmdevice_nodes/filter_by_category.tf with category filter example
- [X] T072 [P] Create /workspace/examples/data-sources/bcm_cmdevice_nodes/filter_by_hostname.tf with hostname pattern filter example
- [X] T073 [P] Create /workspace/examples/data-sources/bcm_cmdevice_nodes/dynamic_inventory.tf with complex usage example
- [X] T074 Generate provider documentation: `make generate`
- [X] T075 Review generated documentation in /workspace/docs/data-sources/cmdevice_nodes.md
- [X] T076 Verify documentation includes all attributes and examples correctly

**Checkpoint Documentation**: Provider documentation generated and examples created

---

## Phase 7: Manual Testing and Validation

**Purpose**: Manual end-to-end testing with Terraform CLI

- [ ] T077 Build and install provider locally: `make install`
- [ ] T078 Create test directory /tmp/tf-test-cmdevice-nodes/ and test configuration file
- [ ] T079 Write test.tf with provider configuration and data source query
- [ ] T080 Run `terraform init` in test directory to initialize provider
- [ ] T081 Run `terraform plan` and verify data source plan output
- [ ] T082 Run `terraform apply` and verify node data retrieval
- [ ] T083 Verify outputs: run `terraform output` to check node hostnames and IPs
- [ ] T084 Test with filter configurations (node_type, category_uuid, hostname_pattern)
- [ ] T085 Test error scenarios (invalid credentials, network failure) and verify error messages

**Checkpoint Manual Testing**: Data source works correctly with Terraform CLI

---

## Phase 8: Polish & Cross-Cutting Concerns

**Purpose**: Final code quality, testing, and validation

- [X] T086 [P] Run full acceptance test suite: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/`
- [X] T087 [P] Run unit tests if any: `make test`
- [X] T088 [P] Run golangci-lint and fix any issues: `golangci-lint run --fix ./internal/provider/`
- [X] T089 [P] Run go fmt to format code: `make fmt`
- [X] T090 [P] Verify pre-commit hooks pass: `pre-commit run --all-files`
- [X] T091 Review all error messages for clarity and actionability
- [X] T092 Review all schema MarkdownDescription fields for completeness
- [X] T093 Validate quickstart.md instructions by following them from scratch
- [X] T094 Run final acceptance test sweep: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource`
- [X] T095 Update feature spec.md with "Implemented" status and completion date

**Checkpoint Final**: Feature complete, tested, documented, and ready for commit

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase completion - Core node discovery
- **User Story 2 (Phase 4)**: Depends on User Story 1 completion - Adds filtering to node discovery
- **User Story 3 (Phase 5)**: Depends on User Story 1 completion - Adds nested attributes to nodes
- **Documentation (Phase 6)**: Depends on all user stories being complete
- **Manual Testing (Phase 7)**: Depends on Documentation phase completion
- **Polish (Phase 8)**: Depends on all previous phases being complete

### User Story Dependencies

- **User Story 1 (P1)**: Can start after Foundational (Phase 2) - No dependencies on other stories - **THIS IS THE MVP**
- **User Story 2 (P2)**: Depends on User Story 1 (needs basic node retrieval working first) - Adds filtering capability
- **User Story 3 (P3)**: Depends on User Story 1 (needs basic node structure) - Adds nested attributes - Can work in parallel with US2 after US1 completes

### Within Each User Story

**TDD Cycle Order** (MUST follow strictly):
1. **RED Phase**: Write failing tests FIRST
2. **GREEN Phase**: Write minimal code to pass tests
3. **REFACTOR Phase**: Improve code quality while keeping tests green

**Implementation Order Within GREEN Phase**:
- Models/structs before schema
- Schema before Read() implementation
- Helper functions as needed
- Registration in provider.go last

### Parallel Opportunities

**Within Setup (Phase 1)**:
- T003 and T004 can run in parallel (independent verification tasks)

**Within US1 GREEN Phase**:
- T015 (NodeModel) and T016 (FilterModel) can run in parallel (different structs)

**Within US2 RED Phase**:
- T035, T036, T037, T038 can all run in parallel (independent test functions and configs)

**Within US3 GREEN Phase**:
- T054 (NetworkInterfaceModel) and T055 (RoleModel) can run in parallel (different structs)

**Within Documentation (Phase 6)**:
- T068-T073 can all run in parallel (creating example files)

**Within Polish (Phase 8)**:
- T086, T087, T088, T089, T090 can all run in parallel (independent quality checks)

---

## Parallel Example: User Story 1 GREEN Phase

```bash
# Can be done in parallel:
Task T015: Define NodeModel struct with basic identity fields
Task T016: Define FilterModel struct

# Then sequentially:
Task T017: Implement NewCMDeviceNodesDataSource() constructor
Task T018: Implement Metadata() method
Task T019: Implement Configure() method
Task T020: Implement minimal Schema()
Task T021: Implement minimal Read() with hardcoded data
Task T022: Register data source in provider.go
Task T023: Run acceptance test
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (verify environment)
2. Complete Phase 2: Foundational (review existing patterns)
3. Complete Phase 3: User Story 1 (basic node discovery)
   - RED: Write failing tests
   - GREEN: Minimal implementation with hardcoded data
   - REFACTOR: Real API integration
4. **STOP and VALIDATE**: Run acceptance tests, verify basic node query works
5. **MVP COMPLETE**: Can query all BCM nodes with basic attributes

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 (Node Discovery) → Test independently → **MVP DEPLOYED**
3. Add User Story 2 (Filtering) → Test independently → Enhanced capability
4. Add User Story 3 (Nested Attributes) → Test independently → Full feature set
5. Add Documentation + Testing → Complete feature
6. Each story adds value without breaking previous stories

### TDD Cycle Strategy

For each user story, follow strict TDD:

1. **RED Phase**: Write ALL acceptance tests for the story, verify they FAIL
2. **GREEN Phase**: Write minimal code to make tests pass (hardcode if needed)
3. **REFACTOR Phase**: Replace with production code, keep tests green
4. **Validate**: Run all tests including previous stories to ensure no regression

---

## Testing Strategy

### Acceptance Test Coverage

**User Story 1 Tests**:
- `TestAccCMDeviceNodesDataSource_Basic` - Query all nodes, verify basic attributes

**User Story 2 Tests**:
- `TestAccCMDeviceNodesDataSource_FilterByType` - Filter by node_type
- `TestAccCMDeviceNodesDataSource_FilterByCategory` - Filter by category_uuid
- `TestAccCMDeviceNodesDataSource_FilterByHostname` - Filter by hostname_pattern

**User Story 3 Tests**:
- `TestAccCMDeviceNodesDataSource_NestedAttributes` - Verify interfaces and roles arrays

**Test Execution**:
```bash
# Run all tests for the data source
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource

# Run specific test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Basic

# Run with verbose logging
TF_ACC=1 TF_LOG=DEBUG go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource
```

### Quality Gates

Before completing each phase:
- All acceptance tests must pass (100% pass rate)
- golangci-lint must pass with no errors
- make fmt must show no changes needed
- Pre-commit hooks must pass

---

## Success Metrics

### Functionality
- Data source retrieves all nodes successfully from BCM API
- All three filter types work correctly (node_type, category_uuid, hostname_pattern)
- Nested interfaces array properly populated with all fields
- Nested roles array properly populated with all fields
- Error messages are clear and actionable

### Quality
- 100% acceptance test pass rate (all tests green)
- golangci-lint score: 0 issues
- All schema attributes have MarkdownDescription
- Code follows existing provider patterns (matches bcm_cmpart_softwareimages style)

### Performance
- API call completes in <5 seconds for 100 nodes
- Client-side filtering executes in <100ms
- Terraform plan/apply completes in <10 seconds

### Documentation
- tfplugindocs generates complete documentation
- All examples work without modification
- Quickstart guide is accurate and complete

---

## Notes

- **TDD Discipline**: Never write implementation before tests fail
- **[P] tasks**: Different files, no dependencies, can run in parallel
- **[Story] label**: Maps task to specific user story for traceability
- **File paths**: Always use absolute paths starting with /workspace/
- **Commit strategy**: Commit after each phase checkpoint
- **Stop points**: Each user story checkpoint allows independent validation
- **Reference implementation**: Follow patterns from /workspace/internal/provider/data_source_cmpart_softwareimages.go

---

## Total Task Count: 95 tasks

**Breakdown by Phase**:
- Setup: 4 tasks
- Foundational: 3 tasks
- User Story 1 (Node Discovery): 25 tasks
- User Story 2 (Filtering): 17 tasks
- User Story 3 (Network/Roles): 18 tasks
- Documentation: 9 tasks
- Manual Testing: 9 tasks
- Polish: 10 tasks

**Estimated Effort**: 4-6 hours for experienced Terraform provider developer following TDD workflow

**MVP Scope** (minimum viable product): Complete through Phase 3 (User Story 1) = 32 tasks = ~2 hours
