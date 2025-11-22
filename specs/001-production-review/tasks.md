# Tasks: Production-Ready Codebase Review

**Input**: Design documents from `/workspace/specs/001-production-review/`
**Prerequisites**: plan.md (complete), spec.md (complete)

**Tests**: Not applicable - this is an analysis and reporting feature that examines existing code.

**Organization**: Tasks are grouped by analysis type (user story) with dependencies clearly marked. The research phase documents methodologies, then individual analyses can run in parallel, and the remediation plan synthesizes all findings.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1=Test Coverage, US2=API Gap, US3=Documentation, US4=Consistency, US5=Remediation)
- Include exact file paths in descriptions

---

## Phase 1: Research & Discovery (Phase 0 per plan.md)

**Purpose**: Document analysis methodologies before executing analyses

- [X] T001 Create research.md documenting test coverage analysis methodology in /workspace/specs/001-production-review/research.md
- [X] T002 Document BCM API discovery approach using live JSON-RPC queries in /workspace/specs/001-production-review/research.md
- [X] T003 Document documentation validation strategy using scripts/test-examples.sh in /workspace/specs/001-production-review/research.md
- [X] T004 Document code consistency analysis patterns and HashiCorp best practices queries in /workspace/specs/001-production-review/research.md
- [X] T005 Document remediation planning framework with prioritization criteria in /workspace/specs/001-production-review/research.md

**Checkpoint**: Methodologies documented - ready to execute analyses

---

## Phase 2: Test Coverage Audit (User Story 1 - P1) 🎯

**Goal**: Analyze all resources and data sources for complete CRUD + Import + Drift + Idempotency test coverage

**Independent Test**: Run analysis against all resource/data source files and verify report identifies specific gaps with file paths and line numbers

### Analysis for User Story 1

- [X] T006 [P] [US1] Use Glob to find all resource implementation files matching internal/provider/resource_*.go pattern
- [X] T007 [P] [US1] Use Glob to find all resource test files matching internal/provider/resource_*_test.go pattern
- [X] T008 [P] [US1] Use Glob to find all data source implementation files matching internal/provider/data_source_*.go pattern
- [X] T009 [P] [US1] Use Glob to find all data source test files matching internal/provider/data_source_*_test.go pattern
- [X] T010 [US1] For each resource test file, use Grep to search for CRUD test coverage patterns (TestStep with Config)
- [X] T011 [US1] For each resource test file, use Grep to search for Import test patterns (ImportState: true)
- [X] T012 [US1] For each resource test file, use Grep to search for Drift test patterns (PreConfig + plancheck.ExpectNonEmptyPlan)
- [X] T013 [US1] For each resource test file, use Grep to search for Idempotency test patterns (plancheck.ExpectEmptyPlan)
- [X] T014 [US1] For each test file, use Grep to identify modern testing patterns (statecheck.ExpectKnownValue, compareValue) vs legacy patterns (resource.TestCheckResourceAttr)
- [X] T015 [US1] For each test file, use Grep to identify environment portability issues (hardcoded counts, names, UUIDs)
- [X] T016 [US1] For each resource test file, analyze CheckDestroy implementation for enhanced pattern usage (verifyResourceDeleted)
- [X] T017 [US1] Generate comprehensive test coverage matrix showing CRUD, Import, Drift, Idempotency status per resource in /workspace/specs/001-production-review/test-coverage-report.md
- [X] T018 [US1] Identify specific gaps with file paths and line numbers for remediation in /workspace/specs/001-production-review/test-coverage-report.md
- [X] T019 [US1] Document testing pattern modernization recommendations in /workspace/specs/001-production-review/test-coverage-report.md

**Checkpoint**: Test coverage report complete with 100% resource analysis and specific gap identification

---

## Phase 3: BCM API Gap Analysis (User Story 2 - P1) 🎯

**Goal**: Query live BCM API to discover all services/methods and compare against provider implementation

**Independent Test**: Validate that analysis queries real BCM endpoint, enumerates all services, and produces prioritized gap list

### Analysis for User Story 2

- [X] T020 [US2] Query live BCM API endpoint (https://172.21.15.254:8081/json) to enumerate all available services using authenticated bcm_client
- [X] T021 [US2] For each discovered BCM service, enumerate available methods through API introspection or systematic exploration
- [X] T022 [US2] Read internal/provider/provider.go Resources() method to list all implemented resources
- [X] T023 [US2] Read internal/provider/provider.go DataSources() method to list all implemented data sources
- [X] T024 [US2] Cross-reference discovered BCM API methods against implemented resources to identify gaps
- [X] T025 [US2] Cross-reference discovered BCM API methods against implemented data sources to identify gaps
- [X] T026 [US2] Categorize API gaps by service type (CMDevice, CMPart, CMNet, CMProv, CMJob, CMServ, CMMon)
- [X] T027 [US2] Assess business value for each gap (high-value: core device/network/software, medium-value: monitoring/jobs, low-value: GUI settings)
- [X] T028 [US2] Identify partial implementations where only data source exists but no resource (or vice versa)
- [X] T029 [US2] Cross-reference sampleRest/ documentation as supplementary material (live API is authoritative)
- [X] T030 [US2] Generate prioritized list of at least 10 high-value missing resources in /workspace/specs/001-production-review/api-gap-analysis.md
- [X] T031 [US2] Document implementation effort estimates for identified gaps in /workspace/specs/001-production-review/api-gap-analysis.md
- [X] T032 [US2] Create recommended implementation order by priority in /workspace/specs/001-production-review/api-gap-analysis.md

**Checkpoint**: API gap analysis complete with service-by-service coverage and prioritized roadmap

---

## Phase 4: Documentation & Examples Validation (User Story 3 - P2)

**Goal**: Validate all examples work correctly and perform root cause analysis on failures

**Independent Test**: Execute scripts/test-examples.sh and verify all examples pass or have documented root cause with fix steps

### Analysis for User Story 3

- [X] T033 [P] [US3] Use Glob to find all resource examples in examples/resources/**/resource.tf
- [X] T034 [P] [US3] Use Glob to find all data source examples in examples/data-sources/**/data-source.tf
- [X] T035 [US3] Execute scripts/test-examples.sh with full output capture to validate all examples
- [X] T036 [US3] For each failing example, extract exact error message and failure point from test output
- [X] T037 [US3] For each failing example, determine root cause (configuration issue vs provider bug)
- [X] T038 [US3] For each failing example, document specific code changes needed to fix
- [X] T039 [US3] For each failing example, define validation approach to confirm fix works
- [X] T040 [US3] Verify all resource examples use unique naming patterns (citest- prefix with timestamps)
- [X] T041 [US3] Verify all examples use environment variables for authentication (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- [X] T042 [US3] Cross-reference examples/ directory structure against registered resources/data sources to identify missing examples
- [X] T043 [US3] Verify generated documentation in docs/ directory matches example configurations
- [X] T044 [US3] Check git timestamps to determine if docs/ needs regeneration via make generate
- [X] T045 [US3] Generate examples validation report with pass/fail status per example in /workspace/specs/001-production-review/documentation-review.md
- [X] T046 [US3] Document root cause analysis for each failing example in /workspace/specs/001-production-review/documentation-review.md
- [X] T047 [US3] List missing examples requiring creation in /workspace/specs/001-production-review/documentation-review.md

**Checkpoint**: Documentation review complete with 100% example coverage assessment and failure remediation plans

---

## Phase 5: Code Consistency Review (User Story 4 - P2)

**Goal**: Analyze code patterns against HashiCorp best practices and identify deviations

**Independent Test**: Validate that analysis uses terraform-provider-design skill and flags all deviations from HashiCorp standards

### Analysis for User Story 4

- [X] T048 [P] [US4] Query terraform-provider-design skill for HashiCorp error handling best practices
- [X] T049 [P] [US4] Query terraform-provider-design skill for HashiCorp schema definition best practices
- [X] T050 [P] [US4] Query terraform-provider-design skill for HashiCorp API client usage best practices
- [X] T051 [P] [US4] Query terraform-provider-design skill for HashiCorp async operation handling best practices
- [X] T052 [P] [US4] Query terraform-provider-design skill for HashiCorp state management best practices
- [X] T053 [US4] Analyze all resource files for consistent parseErrorResponse() usage in error handling
- [X] T054 [US4] Analyze all resource/data source files for schema attribute MarkdownDescription completeness
- [X] T055 [US4] Analyze all resource Read implementations for direct lookup with args parameter usage
- [X] T056 [US4] Analyze all data source Read implementations for list methods with client-side filtering
- [X] T057 [US4] Identify resources with async operations (cloning, provisioning) and verify exponential backoff polling
- [X] T058 [US4] Search for Unknown value propagation in state management across all resources
- [X] T059 [US4] Search for plan value preservation patterns (e.g., original_image after cloning)
- [X] T060 [US4] Analyze schema validators for consistent validation approaches across similar attribute types
- [X] T061 [US4] Flag all deviations from HashiCorp standards regardless of current functionality
- [X] T062 [US4] Generate code consistency report with specific file paths and line numbers in /workspace/specs/001-production-review/code-consistency-report.md
- [X] T063 [US4] Categorize consistency issues by severity (Critical, Medium, Low) in /workspace/specs/001-production-review/code-consistency-report.md
- [X] T064 [US4] Document recommended fixes per HashiCorp best practices in /workspace/specs/001-production-review/code-consistency-report.md

**Checkpoint**: Code consistency analysis complete with categorized issues and actionable recommendations

---

## Phase 6: Phased Remediation Plan (User Story 5 - P3)

**Goal**: Synthesize all analysis findings into prioritized phases with success criteria

**Independent Test**: Validate that remediation plan groups all identified issues with measurable outcomes and regression testing requirements

### Synthesis for User Story 5

- [X] T065 [US5] Aggregate all identified issues from test coverage report (T019)
- [X] T066 [US5] Aggregate all identified issues from API gap analysis (T032)
- [X] T067 [US5] Aggregate all identified issues from documentation review (T047)
- [X] T068 [US5] Aggregate all identified issues from code consistency report (T064)
- [X] T069 [US5] Categorize all issues by production-readiness impact (Critical, High, Medium, Low)
- [X] T070 [US5] Group issues into Phase 0: Critical Blockers (must fix for production)
- [X] T071 [US5] Group issues into Phase 1: High-Priority Gaps (production readiness)
- [X] T072 [US5] Group issues into Phase 2: Medium-Priority Improvements (quality and consistency)
- [X] T073 [US5] Group issues into Phase 3: Polish & Nice-to-Have (future enhancements)
- [X] T074 [US5] Define measurable success criteria for each remediation phase
- [X] T075 [US5] Specify regression testing requirements for each phase (go test, test-examples.sh, BCM API validation)
- [X] T076 [US5] Estimate effort and resource requirements for each remediation phase
- [X] T077 [US5] Document phase dependencies and validation checkpoints
- [X] T078 [US5] Generate complete remediation plan with timeline estimates in /workspace/specs/001-production-review/remediation-plan.md

**Checkpoint**: Remediation plan complete with prioritized phases and actionable roadmap

---

## Phase 7: Polish & Deliverables

**Purpose**: Finalize documentation and validate cross-report consistency

- [X] T079 [P] Create developer quickstart guide in /workspace/specs/001-production-review/quickstart.md
- [X] T080 Validate cross-report consistency (remediation plan references specific findings from other reports)
- [X] T081 Verify all reports include specific file paths, line numbers, and actionable recommendations
- [X] T082 Validate remediation plan groups issues into 3-5 distinct phases per success criteria SC-006
- [X] T083 Verify each remediation phase has quantifiable success criteria per success criteria SC-007
- [X] T084 Confirm test coverage report shows 100% resource analysis per success criteria SC-002
- [X] T085 Confirm API gap analysis identifies at least 10 high-value gaps per success criteria SC-003
- [X] T086 Confirm documentation review addresses all examples with root cause analysis per success criteria SC-004
- [X] T087 Confirm code consistency review covers at least 4 categories per success criteria SC-005

---

## Dependencies & Execution Order

### Phase Dependencies

- **Research (Phase 1)**: No dependencies - start immediately
- **Test Coverage (Phase 2)**: Depends on Research complete - can run in parallel with Phase 3
- **API Gap (Phase 3)**: Depends on Research complete - can run in parallel with Phase 2
- **Documentation (Phase 4)**: Depends on Research complete - can run in parallel with Phase 5
- **Code Consistency (Phase 5)**: Depends on Research complete - can run in parallel with Phase 4
- **Remediation Planning (Phase 6)**: Depends on Phases 2, 3, 4, 5 all complete - synthesizes all findings
- **Polish (Phase 7)**: Depends on Phase 6 complete - final validation

### User Story Dependencies

- **US1 (Test Coverage Audit - P1)**: Can start after Research - No dependencies on other stories
- **US2 (API Gap Analysis - P1)**: Can start after Research - No dependencies on other stories (parallel with US1)
- **US3 (Documentation Validation - P2)**: Can start after Research - No dependencies on other stories (parallel with US1/US2)
- **US4 (Code Consistency Review - P2)**: Can start after Research - No dependencies on other stories (parallel with US1/US2/US3)
- **US5 (Remediation Plan - P3)**: Depends on US1, US2, US3, US4 complete - synthesizes all findings

### Within Each User Story

**US1 (Test Coverage)**:
- T006-T009 (Glob file discovery) can run in parallel [P]
- T010-T016 (Grep analysis) must run sequentially after file discovery
- T017-T019 (Report generation) must run after analysis complete

**US2 (API Gap)**:
- T020-T021 (API discovery) must run sequentially (authenticated session)
- T022-T023 (Provider inventory) can run in parallel with API discovery
- T024-T032 (Gap analysis and report) must run after both complete

**US3 (Documentation)**:
- T033-T034 (Glob examples discovery) can run in parallel [P]
- T035 (Execute test-examples.sh) must run after file discovery
- T036-T047 (Root cause analysis and report) must run after test execution

**US4 (Code Consistency)**:
- T048-T052 (Skill queries) can run in parallel [P]
- T053-T061 (Code analysis) must run after skill queries complete
- T062-T064 (Report generation) must run after analysis complete

**US5 (Remediation Planning)**:
- T065-T068 (Aggregate findings) can run in parallel [P] after prior phases
- T069-T078 (Categorize, group, generate plan) must run sequentially

### Parallel Opportunities

**Research Phase (T001-T005)**:
- All research tasks can be documented in parallel sections of research.md

**Analysis Phase (After Research)**:
- US1 (T006-T019): Test Coverage Analysis
- US2 (T020-T032): API Gap Analysis
- US3 (T033-T047): Documentation Validation
- US4 (T048-T064): Code Consistency Review

All four analyses can run in parallel by different team members or concurrent subagents.

**Within US1**:
```bash
# Launch file discovery in parallel:
Task T006: "Find all resource_*.go files"
Task T007: "Find all resource_*_test.go files"
Task T008: "Find all data_source_*.go files"
Task T009: "Find all data_source_*_test.go files"
```

**Within US4**:
```bash
# Launch skill queries in parallel:
Task T048: "Query error handling best practices"
Task T049: "Query schema definition best practices"
Task T050: "Query API client usage best practices"
Task T051: "Query async operation best practices"
Task T052: "Query state management best practices"
```

**Within US5**:
```bash
# Launch aggregation in parallel:
Task T065: "Aggregate test coverage issues"
Task T066: "Aggregate API gap issues"
Task T067: "Aggregate documentation issues"
Task T068: "Aggregate code consistency issues"
```

---

## Implementation Strategy

### Sequential (Single Analyst)

1. Complete Phase 1: Research (T001-T005) → Methodologies documented
2. Complete Phase 2: User Story 1 (T006-T019) → Test coverage report
3. Complete Phase 3: User Story 2 (T020-T032) → API gap report
4. Complete Phase 4: User Story 3 (T033-T047) → Documentation report
5. Complete Phase 5: User Story 4 (T048-T064) → Consistency report
6. Complete Phase 6: User Story 5 (T065-T078) → Remediation plan
7. Complete Phase 7: Polish (T079-T087) → Final validation and quickstart

**Timeline**: ~15-20 hours of analysis work (per plan.md estimate)

### Parallel (Multiple Analysts or Subagents)

1. **Week 1 Day 1**: Complete Research (T001-T005) together
2. **Week 1 Day 2-3**: Launch parallel analyses:
   - Analyst A: User Story 1 (Test Coverage)
   - Analyst B: User Story 2 (API Gap)
   - Analyst C: User Story 3 (Documentation)
   - Analyst D: User Story 4 (Code Consistency)
3. **Week 1 Day 4**: Synthesize findings (User Story 5 - Remediation Plan)
4. **Week 1 Day 5**: Polish and validation (Phase 7)

**Timeline**: ~5 days with 4 parallel analysts

### MVP Approach (Minimum Viable Analysis)

If time/resources are constrained:

1. **Phase 1**: Research (required for methodology)
2. **Phase 2**: User Story 1 - Test Coverage (P1 - highest priority)
3. **Phase 3**: User Story 2 - API Gap (P1 - highest priority)
4. **Phase 6**: User Story 5 - Remediation Plan (synthesize just US1+US2)
5. **STOP**: Deliver minimal production-readiness assessment

Then incrementally add:
- US3 (Documentation) → Update remediation plan
- US4 (Code Consistency) → Update remediation plan
- Phase 7 (Polish) → Finalize all deliverables

---

## Validation Checkpoints

### After Research (Phase 1)
- [ ] research.md exists and documents all 5 analysis methodologies
- [ ] Each methodology includes tools, commands, and validation approaches
- [ ] Research addresses all research questions from plan.md

### After Test Coverage Analysis (US1)
- [ ] test-coverage-report.md exists with complete resource matrix
- [ ] Report identifies exactly which CRUD/Import/Drift/Idempotency tests are missing
- [ ] Report includes specific file paths and line numbers for gaps
- [ ] 100% of resources and data sources analyzed per SC-002

### After API Gap Analysis (US2)
- [ ] api-gap-analysis.md exists with service-by-service coverage
- [ ] Analysis queried live BCM API endpoint (not just sampleRest/ docs)
- [ ] At least 10 high-value gaps identified per SC-003
- [ ] Gaps prioritized by business value with implementation estimates

### After Documentation Review (US3)
- [ ] documentation-review.md exists with all examples validated
- [ ] scripts/test-examples.sh executed with captured output
- [ ] Each failing example has root cause analysis per SC-004
- [ ] Missing examples identified across all registered resources

### After Code Consistency Review (US4)
- [ ] code-consistency-report.md exists with categorized issues
- [ ] terraform-provider-design skill queried for best practices
- [ ] At least 4 consistency categories analyzed per SC-005
- [ ] All deviations from HashiCorp standards flagged

### After Remediation Planning (US5)
- [ ] remediation-plan.md exists with prioritized phases
- [ ] Issues grouped into 3-5 distinct phases per SC-006
- [ ] Each phase has quantifiable success criteria per SC-007
- [ ] Regression testing requirements specified for each phase
- [ ] Effort estimates and timeline included

### After Polish (Phase 7)
- [ ] quickstart.md provides clear navigation of all reports
- [ ] Cross-report consistency validated (remediation references findings)
- [ ] All success criteria SC-001 through SC-008 verified
- [ ] Production readiness assessment complete and actionable

---

## Notes

- This is an analysis feature - no production code changes are made
- All outputs are markdown reports in /workspace/specs/001-production-review/
- Live BCM cluster access required for US2 (API Gap Analysis)
- Acceptance test environment required for US1 (Test Coverage Analysis)
- terraform-provider-design skill access required for US4 (Code Consistency Review)
- Thoroughness prioritized over speed per clarifications
- Each user story produces an independent, actionable report
- Remediation plan (US5) synthesizes all findings into implementation roadmap
