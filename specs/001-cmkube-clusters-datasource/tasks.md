# Tasks: BCM Kubernetes Clusters Data Source

**Input**: Design documents from `/workspace/specs/001-cmkube-clusters-datasource/`
**Prerequisites**: spec.md, plan.md, research.md, data-model.md
**GitHub Issue**: #27 - Implement data.bcm_cmkube_clusters data source

**Organization**: Tasks organized by TDD workflow phases (RED-GREEN-REFACTOR-DOCS) following terraform-provider-design patterns

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 1: Setup & Infrastructure

**Purpose**: Verify BCM API and establish test infrastructure

- [X] T001 Verify cmkube.getKubeClusters API method via /workspace/sampleRest/cmkube-get-clusters.py script
- [X] T002 Document API response structure in /workspace/specs/001-cmkube-clusters-datasource/contracts/cmkube-api.json
- [X] T003 [P] Review existing data source patterns in /workspace/internal/provider/data_source_cmpart_softwareimages.go for filter implementation
- [X] T004 [P] Review modern testing patterns in /workspace/internal/provider/data_source_cmdevice_nodes_test.go for statecheck usage

**Checkpoint**: API verified, patterns identified, ready to write tests ✅

---

## Phase 2: RED - Write Failing Acceptance Tests

**Purpose**: Test-driven development - write tests FIRST, verify they FAIL

**⚠️ CRITICAL**: All tests in this phase MUST fail initially (no implementation exists yet)

### User Story 1 Tests - Cluster Discovery (P1)

**Goal**: Engineers can discover all BCM Kubernetes clusters without filters
**Independent Test**: Query data source without filters, verify all clusters returned with UUIDs and names

- [X] T005 [US1] Write TestAccCMKubeClustersDataSource_Basic in /workspace/internal/provider/data_source_cmkube_clusters_test.go
  - Assert: ID is not null (placeholder "cmkube-clusters")
  - Assert: clusters[0].uuid is not null
  - Assert: clusters[0].name is not null
  - Assert: clusters[0].master_nodes is not null
  - Use statecheck.ExpectKnownValue() with knownvalue.NotNull()
  - Pattern: data_source_cmdevice_nodes_test.go:18-46

### User Story 2 Tests - Filter by Name Pattern (P2)

**Goal**: Engineers can filter clusters by name pattern for environment organization
**Independent Test**: Create test cluster, filter by name pattern, verify only matching clusters returned

- [X] T006 [US2] Write TestAccCMKubeClustersDataSource_FilterByName in /workspace/internal/provider/data_source_cmkube_clusters_test.go
  - Config: filter { name_pattern = "test-cluster" }
  - Assert: All returned clusters contain pattern in name (case-insensitive)
  - Use statecheck.ExpectKnownValue() for type-safe assertions
  - Pattern: data_source_cmpart_softwareimages_test.go FilterByName

### User Story 3 Tests - Filter by Version (P3)

**Goal**: Engineers can identify clusters by Kubernetes version for upgrade planning
**Independent Test**: Filter by specific version, verify only matching clusters returned

- [X] T007 [US3] Write TestAccCMKubeClustersDataSource_FilterByVersion in /workspace/internal/provider/data_source_cmkube_clusters_test.go
  - Config: filter { version = "1.28.0" }
  - Assert: All returned clusters have version = "1.28.0"
  - Use knownvalue.StringExact() for version matching
  - Pattern: Similar to FilterByCategory in softwareimages test

### User Story 4 Tests - Filter by Master Node (P3)

**Goal**: Engineers can find which cluster(s) contain a specific master node
**Independent Test**: Filter by master node UUID, verify clusters containing that node returned

- [X] T008 [US4] Write TestAccCMKubeClustersDataSource_FilterByMasterNode in /workspace/internal/provider/data_source_cmkube_clusters_test.go
  - Config: filter { master_node_id = "node-uuid-123" }
  - Assert: All returned clusters contain specified UUID in master_nodes list
  - Use tfjsonpath to navigate nested list attributes

### Edge Case Tests (All User Stories)

- [X] T009 [P] Write TestAccCMKubeClustersDataSource_MultipleFilters in /workspace/internal/provider/data_source_cmkube_clusters_test.go
  - Config: filter { name_pattern = "prod"; version = "1.28.0" }
  - Assert: Returned clusters match ALL filters (AND logic)
  - Verify filter combination behavior

- [X] T010 [P] Write TestAccCMKubeClustersDataSource_EmptyResults in /workspace/internal/provider/data_source_cmkube_clusters_test.go
  - Config: filter { name_pattern = "nonexistent-cluster-xyz" }
  - Assert: clusters list is empty (not error)
  - Assert: ID still set to "cmkube-clusters"
  - Verify graceful handling of no matches

- [X] T011 [P] Write TestAccCMKubeClustersDataSource_NullFields in /workspace/internal/provider/data_source_cmkube_clusters_test.go
  - Assumes BCM has cluster with null optional fields
  - Assert: Optional fields (worker_nodes, dns_servers) can be null
  - Assert: Required fields (uuid, name, master_nodes) present
  - Verify null-safe field extraction

- [X] T012 Run acceptance tests to verify ALL tests FAIL with "resource not found" errors
  - Command: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource
  - Expected: 0 passed, 7 failed (no data source implementation exists)
  - Verify: Clear error messages about missing data source

**Checkpoint**: 7 failing tests written, RED phase complete ✅ DONE

---

## Phase 3: GREEN - Minimal Implementation

**Purpose**: Write minimal code to make tests PASS (hardcoded responses acceptable)

### Data Source Skeleton

- [X] T013 Create /workspace/internal/provider/data_source_cmkube_clusters.go with minimal structure
  - Define CMKubeClustersDataSource struct
  - Implement Metadata() method (returns "cmkube_clusters")
  - Define placeholder Schema() method (empty schema)
  - Define placeholder Read() method (returns empty state)
  - Pattern: data_source_cmpart_softwareimages.go:23-40

### Schema Definition

- [X] T014 Implement Schema() method in /workspace/internal/provider/data_source_cmkube_clusters.go
  - Define "id" attribute (computed string, placeholder "cmkube-clusters")
  - Define "clusters" attribute (computed list of nested objects)
  - Define cluster nested attributes: id, uuid, name, master_nodes, worker_nodes, etc.
  - Define "filter" block (optional) with name_pattern, version, master_node_id
  - Add MarkdownDescription for each attribute
  - Pattern: data_source_cmpart_softwareimages.go:69-180

### Helper Functions

- [X] T015 [P] Add getListValue() helper function in /workspace/internal/provider/data_source_cmkube_clusters.go
  - Extract []interface{} from BCM API response
  - Convert to types.List with StringType elements
  - Handle null/empty lists gracefully
  - Return types.ListNull(types.StringType) if missing
  - Pattern: research.md lines 246-269

- [X] T016 [P] Add mapClusterDataToModel() helper function in /workspace/internal/provider/data_source_cmkube_clusters.go
  - Map BCM API fields (camelCase) to Terraform attributes (snake_case)
  - Use getStringValue() for string fields
  - Use getInt64Value() for numeric fields
  - Use getListValue() for list fields
  - Pattern: data_model.md lines 268-307

### Configure Method

- [X] T017 Implement Configure() method in /workspace/internal/provider/data_source_cmkube_clusters.go
  - Accept BCMClient from provider configuration
  - Store client reference for Read() method
  - Handle nil client gracefully
  - Pattern: data_source_cmpart_softwareimages.go:48-66

### Read Method (Minimal - Hardcoded Response)

- [X] T018 Implement Read() method with HARDCODED response in /workspace/internal/provider/data_source_cmkube_clusters.go
  - Create hardcoded cluster data (static JSON)
  - Map to KubeClusterModel structs
  - Set data.Clusters = hardcoded list
  - Set data.ID = types.StringValue("cmkube-clusters")
  - NO API call yet (GREEN phase - minimal implementation)
  - NO filtering yet (just return all hardcoded clusters)

### Provider Registration

- [X] T019 Register CMKubeClustersDataSource in /workspace/internal/provider/provider.go DataSources() method
  - Add NewCMKubeClustersDataSource to DataSources() return slice
  - Ensure alphabetical ordering with other data sources
  - Pattern: provider.go DataSources() method

### Verify GREEN Phase

- [X] T020 Run acceptance tests to verify ALL tests PASS with minimal implementation
  - Command: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource
  - Expected: 7 passed, 0 failed (hardcoded data satisfies test assertions)
  - Fix any failing tests by adjusting hardcoded data

**Checkpoint**: All tests passing with minimal implementation, GREEN phase complete ✅ DONE

---

## Phase 4: REFACTOR - Full Implementation with BCM API

**Purpose**: Replace hardcoded responses with real BCM API integration and filtering

### API Integration

- [X] T021 Replace hardcoded response with BCM API call in Read() method
  - Call d.client.CallJSONRPC(ctx, "cmkube", "getKubeClusters")
  - Handle API errors with resp.Diagnostics.AddError()
  - Parse JSON response into []map[string]interface{}
  - Use mapClusterDataToModel() for each cluster
  - Pattern: data_source_cmpart_softwareimages.go:120-155

### Filter Implementation - User Story 2 (Name Pattern)

- [X] T022 [US2] Implement name_pattern filter in Read() method
  - Check if data.Filter != nil && !data.Filter.NamePattern.IsNull()
  - Extract pattern: strings.ToLower(data.Filter.NamePattern.ValueString())
  - Extract cluster name: strings.ToLower(getStringValue(clusterData, "name").ValueString())
  - Apply filter: !strings.Contains(clusterName, pattern) → exclude = true
  - Pattern: data_source_cmpart_softwareimages.go:165-173

### Filter Implementation - User Story 3 (Version)

- [X] T023 [US3] Implement version filter in Read() method
  - Check if data.Filter != nil && !data.Filter.Version.IsNull()
  - Extract cluster version: getStringValue(clusterData, "version").ValueString()
  - Apply filter: clusterVersion != data.Filter.Version.ValueString() → exclude = true
  - Exact match (not substring)
  - Pattern: Similar to category filter in softwareimages

### Filter Implementation - User Story 4 (Master Node)

- [X] T024 [US4] Implement master_node_id filter in Read() method
  - Check if data.Filter != nil && !data.Filter.MasterNodeID.IsNull()
  - Extract masterNodes array from clusterData["masterNodes"].([]interface{})
  - Iterate list, check if any element matches target UUID
  - Apply filter: UUID not found in list → exclude = true
  - Pattern: research.md lines 178-198

### Error Handling Enhancement

- [X] T025 Add comprehensive error handling in Read() method
  - BCM API unreachable → clear error message
  - Authentication failure → clear error message
  - Invalid JSON response → clear error message
  - Empty cluster list → success with empty array (not error)
  - Log warnings for invalid cluster data, skip cluster, continue
  - Pattern: Error Handling Strategy from spec.md lines 271-279

### Verify REFACTOR Phase

- [X] T026 Run acceptance tests to verify ALL tests PASS with real BCM API integration
  - Command: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource
  - Expected: 7 passed, 0 failed (real API data satisfies test assertions)
  - Verify filters work correctly with live BCM data
  - Fix any issues with null field handling or filter logic

**Checkpoint**: Full implementation complete with BCM API, tests passing, REFACTOR phase complete ✅ DONE

---

## Phase 5: Documentation & Validation

**Purpose**: Generate documentation and validate feature completeness

### Example Configurations

- [X] T027 [P] Create /workspace/examples/data-sources/bcm_cmkube_clusters/data-source.tf with basic example
  - Example: List all clusters (no filter)
  - Include provider block with environment variables
  - Pattern: examples/data-sources/bcm_cmdevice_nodes/data-source.tf

- [X] T028 [P] Add filter examples to /workspace/examples/data-sources/bcm_cmkube_clusters/data-source.tf
  - Example: Filter by name pattern (filter { name_pattern = "prod-*" })
  - Example: Filter by version (filter { version = "1.28.0" })
  - Example: Filter by master node (filter { master_node_id = "node-uuid" })
  - Example: Multiple filters combined (AND logic)
  - Example: Use cluster UUID for terraform import

### Documentation Generation

- [X] T029 Generate provider documentation via make generate
  - Command: make generate
  - Expected: /workspace/docs/data-sources/cmkube_clusters.md created
  - Verify: All attributes documented with descriptions
  - Verify: Filter block documented with examples
  - Verify: Examples formatted correctly

### Code Quality

- [X] T030 [P] Run code formatting and linting
  - Command: make fmt && make lint
  - Expected: No formatting issues, no lint errors
  - Fix any issues reported by golangci-lint

### Final Validation

- [X] T031 Run full acceptance test suite for data source
  - Command: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource
  - Expected: All 7 tests pass (Basic, FilterByName, FilterByVersion, FilterByMasterNode, MultipleFilters, EmptyResults, NullFields)
  - Verify: No flaky tests, consistent results across runs

- [X] T032 Validate example configurations work with BCM
  - Navigate to /workspace/examples/data-sources/bcm_cmkube_clusters/
  - Run: terraform init && terraform validate
  - Run: terraform plan (with BCM credentials)
  - Verify: Examples execute successfully without errors

- [X] T033 Verify schema consistency with resource_cmkube_cluster.go
  - Cross-reference attribute names and types
  - Verify: All resource attributes present in data source
  - Verify: Type consistency (types.String, types.List, types.Int64)
  - Verify: Import compatibility (cluster UUID can be used for terraform import)

**Checkpoint**: Documentation complete, all validation passing, feature ready for review ✅ DONE

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup)**: No dependencies - start immediately
- **Phase 2 (RED)**: Depends on Phase 1 (API verified, patterns reviewed)
- **Phase 3 (GREEN)**: Depends on Phase 2 (tests written and failing)
- **Phase 4 (REFACTOR)**: Depends on Phase 3 (minimal implementation passing tests)
- **Phase 5 (DOCS)**: Depends on Phase 4 (full implementation complete)

### Task Dependencies Within Phases

**Phase 2 (RED)**:
- T005-T011 (all test files): Can run in parallel [P] - writing to same file but different test functions
- T012 (run tests): Depends on T005-T011 completion

**Phase 3 (GREEN)**:
- T013 (skeleton): Must complete before all other tasks
- T014 (schema): Depends on T013
- T015-T016 (helpers): Can run in parallel [P] - different functions
- T017 (Configure): Depends on T013
- T018 (Read): Depends on T014, T015, T016
- T019 (registration): Can run in parallel with T018
- T020 (verify): Depends on T013-T019 completion

**Phase 4 (REFACTOR)**:
- T021 (API integration): Depends on Phase 3 completion
- T022-T024 (filters): Depend on T021, can run sequentially (modify same Read() method)
- T025 (error handling): Depends on T021
- T026 (verify): Depends on T021-T025 completion

**Phase 5 (DOCS)**:
- T027-T028 (examples): Can run in parallel [P] - writing to same file incrementally
- T029 (generate): Depends on T027-T028
- T030 (quality): Can run in parallel with T029
- T031-T033 (validation): Depend on T029-T030 completion

### Parallel Opportunities

**Phase 1**: T003 and T004 can run in parallel (different files)

**Phase 2**: All test writing (T005-T011) can proceed in parallel

**Phase 3**: T015 and T016 (helpers) can run in parallel

**Phase 5**: T027-T028 (examples), T030 (quality) can run in parallel

---

## Parallel Example: Phase 2 (RED)

```bash
# Launch all test writing tasks together (different test functions):
Task T005: "Write TestAccCMKubeClustersDataSource_Basic"
Task T006: "Write TestAccCMKubeClustersDataSource_FilterByName"
Task T007: "Write TestAccCMKubeClustersDataSource_FilterByVersion"
Task T008: "Write TestAccCMKubeClustersDataSource_FilterByMasterNode"
Task T009: "Write TestAccCMKubeClustersDataSource_MultipleFilters"
Task T010: "Write TestAccCMKubeClustersDataSource_EmptyResults"
Task T011: "Write TestAccCMKubeClustersDataSource_NullFields"

# Then verify all fail together:
Task T012: "Run tests to verify all fail"
```

---

## Implementation Strategy

### TDD Workflow (Recommended)

1. **Complete Phase 1**: Verify API and patterns (foundation)
2. **Complete Phase 2**: Write all failing tests (RED phase) ✅
3. **Complete Phase 3**: Minimal implementation to pass tests (GREEN phase) ✅
4. **Complete Phase 4**: Full BCM API integration (REFACTOR phase) ✅
5. **Complete Phase 5**: Documentation and validation

### User Story Delivery Order

Since this is a single data source implementation:

1. **US1 (P1)** - Cluster Discovery: Basic list all functionality → Tests T005, Implementation T013-T020
2. **US2 (P2)** - Filter by Name: Name pattern filtering → Test T006, Implementation T022
3. **US3 (P3)** - Filter by Version: Version filtering → Test T007, Implementation T023
4. **US4 (P3)** - Filter by Master Node: Master node filtering → Test T008, Implementation T024

Each user story builds incrementally on the previous foundation.

### MVP Scope

**Minimum Viable Product** = User Story 1 (Cluster Discovery):
- Basic data source implementation
- List all clusters without filters
- No filter functionality yet
- Sufficient for cluster discovery and import use cases

Stop after Phase 3 (GREEN) + Phase 5 (DOCS) for MVP.

### Success Criteria

- ✅ All 7 acceptance tests pass (100% test coverage)
- ✅ Data source registered in provider.go
- ✅ Examples created and validated
- ✅ Documentation auto-generated via make generate
- ✅ Schema matches resource_cmkube_cluster.go (import compatibility)
- ✅ All 12 functional requirements from spec.md verified
- ✅ Code passes make lint (HashiCorp style guide)
- ✅ Zero new complexity violations (reuses existing patterns)

---

## Notes

- **[P] tasks**: Different functions/files, can run in parallel
- **[Story] labels**: Track which user story each task supports
- **TDD discipline**: Tests MUST fail before implementation (verify RED phase)
- **Pattern reuse**: All tasks reference existing implementations as patterns
- **Autonomous execution**: Each task has clear acceptance criteria for /speckit.implement
- **Schema consistency**: Cross-reference resource_cmkube_cluster.go throughout
- **Modern testing**: Use statecheck, knownvalue, tfjsonpath (terraform-plugin-testing v1.13.3+)
- **Environment portability**: No hardcoded cluster counts or names in tests

---

## References

- **Specification**: `/workspace/specs/001-cmkube-clusters-datasource/spec.md`
- **Implementation Plan**: `/workspace/specs/001-cmkube-clusters-datasource/plan.md`
- **Research**: `/workspace/specs/001-cmkube-clusters-datasource/research.md`
- **Data Model**: `/workspace/specs/001-cmkube-clusters-datasource/data-model.md`
- **Pattern Sources**:
  - `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
  - `/workspace/internal/provider/data_source_cmdevice_nodes.go`
  - `/workspace/internal/provider/data_source_cmdevice_nodes_test.go`
  - `/workspace/internal/provider/resource_cmkube_cluster.go`
- **Project Guidelines**: `/workspace/CLAUDE.md`, `/workspace/AGENTS.md`
