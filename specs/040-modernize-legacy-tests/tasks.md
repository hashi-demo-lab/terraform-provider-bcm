---
description: "Task list for modernizing legacy testing patterns (Issue #40)"
---

# Tasks: Modernize Legacy Testing Patterns

**Input**: Design documents from `/workspace/specs/040-modernize-legacy-tests/`
**Prerequisites**: plan.md, spec.md

**Organization**: Tasks are organized by priority phases (P1 → P2 → P3) for sequential file-by-file transformation with validation gates between phases.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **No Story label**: This is a refactoring effort, not new user stories

## Phase 1: Setup & Research

**Purpose**: Prepare for transformation with comprehensive analysis

- [ ] T001 Create research.md with legacy pattern inventory using grep analysis on 3 target test files
- [ ] T002 Document modern pattern mappings in research.md from CLAUDE.md Modern Testing Patterns section
- [ ] T003 [P] Identify import requirements per file and document in research.md
- [ ] T004 [P] Analyze missing idempotency checks and document in research.md
- [ ] T005 [P] Analyze missing ID tracking patterns and document in research.md
- [ ] T006 Verify research.md shows 66 total legacy assertions cataloged across all files
- [ ] T007 Run baseline acceptance tests to establish performance benchmark: `TF_ACC=1 time go test -v -timeout 120m ./internal/provider/ -run "TestAccCMNetNetwork|TestAccCMKubeCluster|TestAccCMPartPartitions"`

**Checkpoint**: Research phase complete - 66 legacy assertions identified, transformation rules documented

---

## Phase 2: Design & Contracts

**Purpose**: Create detailed transformation specifications and developer guides

**⚠️ CRITICAL**: This phase defines the exact before/after patterns for all transformations

- [ ] T008 Create data-model.md with test function structure mapping for resource_cmnet_network_test.go (3 functions)
- [ ] T009 [P] Extend data-model.md with test function structure mapping for resource_cmkube_cluster_test.go (10 functions)
- [ ] T010 [P] Extend data-model.md with test function structure mapping for data_source_cmpart_partitions_test.go (8 functions)
- [ ] T011 Create contracts/01-string-attribute-transformation.md with before/after examples
- [ ] T012 [P] Create contracts/02-numeric-attribute-transformation.md with Int64Exact pattern
- [ ] T013 [P] Create contracts/03-boolean-attribute-transformation.md with Bool type safety pattern
- [ ] T014 [P] Create contracts/04-computed-attribute-transformation.md with NotNull pattern
- [ ] T015 [P] Create contracts/05-list-attribute-transformation.md with ListSizeExact vs NotNull guidance
- [ ] T016 Create contracts/06-idempotency-addition.md with ExpectEmptyPlan step pattern
- [ ] T017 Create contracts/07-id-tracking-addition.md with CompareValue pattern and initialization
- [ ] T018 Verify CheckDestroy functions in resource_cmnet_network_test.go and resource_cmkube_cluster_test.go match CLAUDE.md enhanced pattern
- [ ] T019 Document migration sequence dependencies in data-model.md (P1→P2→P3 with validation gates)
- [ ] T020 Create quickstart.md with prerequisites, transformation checklist, pattern quick reference, and troubleshooting guide
- [ ] T021 Run update-agent-context.sh to add testing modernization patterns to agent context

**Checkpoint**: Design artifacts complete - 7 transformation contracts created, developer guide ready

---

## Phase 3: P1 Implementation - resource_cmnet_network_test.go

**Goal**: Modernize simplest file first to establish transformation workflow (3 functions, 18 legacy assertions)

**Independent Test**: `TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run TestAccCMNetNetwork`

### Add Required Imports

- [ ] T022 Add required imports to internal/provider/resource_cmnet_network_test.go: statecheck, knownvalue, tfjsonpath, compare

### Transform TestAccCMNetNetwork_Basic

- [ ] T023 Replace Check block with ConfigStateChecks in TestAccCMNetNetwork_Basic (4 legacy assertions → 4 modern)
- [ ] T024 Add `compareID := statecheck.CompareValue(compare.ValuesSame())` initialization to TestAccCMNetNetwork_Basic
- [ ] T025 Add ID tracking to Create step in TestAccCMNetNetwork_Basic using compareID.AddStateValue
- [ ] T026 Add ID tracking to Import step in TestAccCMNetNetwork_Basic using compareID.AddStateValue
- [ ] T027 Run test validation: `TF_ACC=1 go test -v -run TestAccCMNetNetwork_Basic ./internal/provider/`

### Transform TestAccCMNetNetwork_Complete

- [ ] T028 Replace Check block with ConfigStateChecks in TestAccCMNetNetwork_Complete (10 legacy assertions → 10 modern, including mtu as Int64Exact(9000), dhcp_enabled as Bool(true))
- [ ] T029 Add idempotency step after Create in TestAccCMNetNetwork_Complete with ExpectEmptyPlan
- [ ] T030 Run test validation: `TF_ACC=1 go test -v -run TestAccCMNetNetwork_Complete ./internal/provider/`

### Transform TestAccCMNetNetwork_Update

- [ ] T031 Replace Create step Check block with ConfigStateChecks in TestAccCMNetNetwork_Update
- [ ] T032 Add idempotency step after Create in TestAccCMNetNetwork_Update with ExpectEmptyPlan
- [ ] T033 Replace Update step Check block with ConfigStateChecks in TestAccCMNetNetwork_Update
- [ ] T034 Add idempotency step after Update in TestAccCMNetNetwork_Update with ExpectEmptyPlan
- [ ] T035 Run test validation: `TF_ACC=1 go test -v -run TestAccCMNetNetwork_Update ./internal/provider/`

### P1 File Validation

- [ ] T036 Verify zero legacy patterns remain in resource_cmnet_network_test.go: `grep -c "TestCheckResourceAttr" internal/provider/resource_cmnet_network_test.go` must return 0
- [ ] T037 Run full P1 acceptance tests: `TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run TestAccCMNetNetwork`
- [ ] T038 Verify all 3 test functions pass with zero regressions
- [ ] T039 Compare test execution time against baseline (must be within 10% threshold)

**Checkpoint**: P1 complete - resource_cmnet_network_test.go 100% modern (18 assertions converted), all tests passing

---

## Phase 4: P2 Implementation - resource_cmkube_cluster_test.go

**Goal**: Modernize largest file with mixed patterns (10 functions, 38 legacy assertions to remove)

**Independent Test**: `TF_ACC=1 go test -v -timeout 60m ./internal/provider/ -run TestAccCMKubeCluster`

**Note**: 6 functions already fully modern (P3 tests), 2 validation tests with no assertions - focus on 4 functions with legacy Check blocks

### Add Required Imports

- [ ] T040 Add required imports to internal/provider/resource_cmkube_cluster_test.go: statecheck, knownvalue, tfjsonpath, compare (if not already present)

### Transform TestAccCMKubeClusterResource_Basic

- [ ] T041 Remove legacy Check block from Create step in TestAccCMKubeClusterResource_Basic (lines 178-183) - keep ConfigStateChecks
- [ ] T042 Remove legacy Check block from Update step in TestAccCMKubeClusterResource_Basic (line 228) - convert to ConfigStateChecks
- [ ] T043 Run test validation: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_Basic ./internal/provider/`

### Transform TestAccCMKubeClusterResource_DriftDetection

- [ ] T044 Remove legacy Check block at line 270 in TestAccCMKubeClusterResource_DriftDetection - keep ConfigStateChecks
- [ ] T045 Remove legacy Check block at line 340 in TestAccCMKubeClusterResource_DriftDetection - keep ConfigStateChecks
- [ ] T046 Run test validation: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_DriftDetection ./internal/provider/`

### Transform TestAccCMKubeClusterResource_WorkerNodes

- [ ] T047 Remove legacy Check block at line 374 in TestAccCMKubeClusterResource_WorkerNodes - keep ConfigStateChecks
- [ ] T048 Remove legacy Check block at line 392 in TestAccCMKubeClusterResource_WorkerNodes - keep ConfigStateChecks
- [ ] T049 Remove legacy Check block at line 410 in TestAccCMKubeClusterResource_WorkerNodes - keep ConfigStateChecks
- [ ] T050 Run test validation: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_WorkerNodes ./internal/provider/`

### Transform TestAccCMKubeClusterResource_CompleteConfiguration

- [ ] T051 Remove legacy Check block at lines 486-492 in TestAccCMKubeClusterResource_CompleteConfiguration - keep ConfigStateChecks
- [ ] T052 Remove legacy Check block at lines 545-548 in TestAccCMKubeClusterResource_CompleteConfiguration - keep ConfigStateChecks
- [ ] T053 Run test validation: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_CompleteConfiguration ./internal/provider/`

### P2 File Validation

- [ ] T054 Verify zero legacy patterns remain in modernized functions: `grep -n "TestCheckResourceAttr\|TestCheckResourceAttrSet" internal/provider/resource_cmkube_cluster_test.go` shows only P3 test comments if any
- [ ] T055 Run full P2 acceptance tests: `TF_ACC=1 go test -v -timeout 60m ./internal/provider/ -run TestAccCMKubeCluster`
- [ ] T056 Verify all 10 test functions pass with zero regressions
- [ ] T057 Compare test execution time against baseline (must be within 10% threshold)

**Checkpoint**: P2 complete - resource_cmkube_cluster_test.go 100% modern (38 legacy assertions removed), all tests passing

---

## Phase 5: P3 Implementation - data_source_cmpart_partitions_test.go

**Goal**: Modernize data source patterns (8 functions, 10 legacy assertions to remove)

**Independent Test**: `TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run TestAccCMPartPartitions`

**Note**: 4 functions already fully modern - focus on 5 functions with legacy Check blocks

### Add Required Imports

- [ ] T058 Add required imports to internal/provider/data_source_cmpart_partitions_test.go: statecheck, knownvalue, tfjsonpath (if not already present)

### Transform TestAccCMPartPartitionsDataSource_ComputedFields

- [ ] T059 Remove legacy Check block at lines 93-96 in TestAccCMPartPartitionsDataSource_ComputedFields - keep ConfigStateChecks
- [ ] T060 Run test validation: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_ComputedFields ./internal/provider/`

### Transform TestAccCMPartPartitionsDataSource_AttributeTypes

- [ ] T061 Remove legacy Check block at lines 166-177 in TestAccCMPartPartitionsDataSource_AttributeTypes - keep ConfigStateChecks
- [ ] T062 Run test validation: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_AttributeTypes ./internal/provider/`

### Transform TestAccCMPartPartitionsDataSource_ListAttributes

- [ ] T063 Remove legacy Check block at lines 226-229 in TestAccCMPartPartitionsDataSource_ListAttributes - keep ConfigStateChecks
- [ ] T064 Run test validation: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_ListAttributes ./internal/provider/`

### Transform TestAccCMPartPartitionsDataSource_FilterEmptyString

- [ ] T065 Remove legacy Check block at lines 333-336 in TestAccCMPartPartitionsDataSource_FilterEmptyString - keep ConfigStateChecks
- [ ] T066 Run test validation: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_FilterEmptyString ./internal/provider/`

### Transform TestAccCMPartPartitionsDataSource_FilterSubsetProperty

- [ ] T067 Remove legacy Check block at lines 366-370 in TestAccCMPartPartitionsDataSource_FilterSubsetProperty - keep ConfigStateChecks
- [ ] T068 Remove legacy Check block at lines 386-390 in TestAccCMPartPartitionsDataSource_FilterSubsetProperty - keep ConfigStateChecks
- [ ] T069 Run test validation: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_FilterSubsetProperty ./internal/provider/`

### P3 File Validation

- [ ] T070 Verify zero legacy patterns remain in modernized functions: `grep -c "TestCheckResourceAttr\|TestCheckResourceAttrSet" internal/provider/data_source_cmpart_partitions_test.go` shows 0 matches
- [ ] T071 Run full P3 acceptance tests: `TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run TestAccCMPartPartitions`
- [ ] T072 Verify all 8 test functions pass with zero regressions
- [ ] T073 Compare test execution time against baseline (must be within 10% threshold)

**Checkpoint**: P3 complete - data_source_cmpart_partitions_test.go 100% modern (10 legacy assertions removed), all tests passing

---

## Phase 6: Final Validation & Documentation

**Purpose**: Comprehensive verification across all modernized files

- [ ] T074 Run full acceptance test suite: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/`
- [ ] T075 Verify success criteria SC-001: Zero `resource.TestCheckResourceAttr()` calls remain in 3 target files
- [ ] T076 Verify success criteria SC-002: Zero `resource.TestCheckResourceAttrSet()` calls remain in 3 target files
- [ ] T077 Verify success criteria SC-003: All numeric attributes use `Int64Exact()` not string "9000"
- [ ] T078 Verify success criteria SC-004: All boolean attributes use `Bool(true/false)` not string "true"
- [ ] T079 Verify success criteria SC-005: All resource tests include idempotency verification after Create and Update
- [ ] T080 Verify success criteria SC-006: All resource tests include ID consistency tracking across operations
- [ ] T081 Verify success criteria SC-007: 100% modern pattern adoption in all 3 files
- [ ] T082 Verify success criteria SC-008: All 21 test functions pass without regressions
- [ ] T083 Verify success criteria SC-009: Test execution time within 10% of baseline
- [ ] T084 Run code quality checks: `make fmt && make lint`
- [ ] T085 Update CLAUDE.md if new testing patterns discovered during implementation
- [ ] T086 Create final validation report documenting: total assertions converted (66), test pass rate (100%), performance comparison

**Checkpoint**: All success criteria met - 100% modern pattern adoption achieved

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 1 (Setup & Research)**: No dependencies - start immediately
- **Phase 2 (Design & Contracts)**: Depends on Phase 1 complete
- **Phase 3 (P1 Implementation)**: Depends on Phase 2 complete - MUST validate before P2
- **Phase 4 (P2 Implementation)**: Depends on Phase 3 validated - MUST validate before P3
- **Phase 5 (P3 Implementation)**: Depends on Phase 4 validated - MUST validate before final
- **Phase 6 (Final Validation)**: Depends on all implementation phases complete

### Critical Validation Gates

**IMPORTANT**: Each priority phase MUST be validated before proceeding to next phase

1. **After Phase 1**: Research complete, 66 assertions cataloged, baseline established
2. **After Phase 2**: 7 transformation contracts created, migration sequence documented
3. **After Phase 3**: resource_cmnet_network_test.go 100% modern, all tests passing
4. **After Phase 4**: resource_cmkube_cluster_test.go 100% modern, all tests passing
5. **After Phase 5**: data_source_cmpart_partitions_test.go 100% modern, all tests passing
6. **After Phase 6**: All 9 success criteria verified

### Within Each Implementation Phase

- Imports before any transformations
- Test function transformations can proceed sequentially (one function at a time)
- Function validation immediately after transformation (fail fast)
- File validation after all functions in file transformed
- NO proceeding to next priority file until current file fully validated

### Parallel Opportunities

- **Phase 1**: Tasks T003, T004, T005 can run in parallel (different analysis types)
- **Phase 2**: Contract creation tasks T012-T015 can run in parallel (different files)
- **Within each file**: Transform individual test functions sequentially (safer, easier debugging)

---

## Parallel Example: Phase 2 Contract Creation

```bash
# Launch contract file creation tasks together:
Task: "Create contracts/02-numeric-attribute-transformation.md"
Task: "Create contracts/03-boolean-attribute-transformation.md"
Task: "Create contracts/04-computed-attribute-transformation.md"
Task: "Create contracts/05-list-attribute-transformation.md"
```

---

## Implementation Strategy

### Sequential File Transformation (Recommended)

1. Complete Phase 1: Research & Setup → Establish baseline
2. Complete Phase 2: Design & Contracts → Define transformation rules
3. **Phase 3: P1 File (Simplest)**
   - Transform all 3 functions
   - Validate with acceptance tests
   - **GATE**: Must pass before P2
4. **Phase 4: P2 File (Most Complex)**
   - Transform 4 legacy functions
   - Validate with acceptance tests
   - **GATE**: Must pass before P3
5. **Phase 5: P3 File (Data Source)**
   - Transform 5 legacy functions
   - Validate with acceptance tests
   - **GATE**: Must pass before final validation
6. **Phase 6: Final Validation**
   - Run full suite
   - Verify all success criteria
   - Document results

### Risk Mitigation Strategy

- **Start simple**: P1 file establishes workflow with minimal risk
- **Validate early**: Each file validated independently before next file
- **Fail fast**: Run test after each function transformation
- **Baseline comparison**: Track performance at each phase gate
- **Rollback ready**: Git commit after each validated phase

### Performance Monitoring

```bash
# Track execution time at each phase gate:
# After P1: TF_ACC=1 time go test -v -run TestAccCMNetNetwork
# After P2: TF_ACC=1 time go test -v -run TestAccCMKubeCluster
# After P3: TF_ACC=1 time go test -v -run TestAccCMPartPartitions
# After P6: TF_ACC=1 time go test -v -timeout 120m ./internal/provider/
```

---

## Notes

### Task Format Compliance

- All tasks follow `- [ ] [ID] [P?] Description with file path` format
- Task IDs sequential (T001-T086)
- [P] marker only for truly parallelizable tasks
- No [Story] labels (this is refactoring, not user stories)
- File paths included where applicable

### Type Matcher Strategy

- **Empirical approach**: Start with `Int64Exact()` for numeric attributes
- **Fallback**: If type errors occur, switch to `Float64Exact()`
- **Document decisions**: Track numeric type choices in validation tasks
- **Computed fields**: Always use `NotNull()`, never hardcoded values

### Environment Requirements

```bash
# Required for all acceptance test validations:
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

### Success Metrics Summary

- **Total Tasks**: 86 tasks across 6 phases
- **Estimated Duration**: 14-19 hours (per plan.md timeline)
- **Legacy Assertions**: 66 to convert (18 P1 + 38 P2 + 10 P3)
- **Test Functions**: 21 total (3 P1 + 10 P2 + 8 P3)
- **Files Modified**: 3 test files
- **Validation Gates**: 5 critical checkpoints
- **Success Criteria**: 9 criteria (SC-001 through SC-009)

### Commit Strategy

- Commit after each validated phase (5 commits minimum)
- Commit messages format: "feat(tests): Modernize [file] to terraform-plugin-testing v1.13.3 patterns"
- Include success criteria verification in final commit message
