# Implementation Plan: Complete TDD-Based Review and Refactoring of resource_cmpart_softwareimage

**Branch**: `005-tdd-softwareimage-refactor` | **Date**: 2025-11-21 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/005-tdd-softwareimage-refactor/spec.md`

## Summary

This plan delivers a comprehensive TDD-based review and refactoring of the existing `bcm_cmpart_softwareimage` Terraform resource. The resource manages BCM software images (OS kernels + filesystems) for DPU node provisioning, supporting cloning, kernel configuration, module management, and Serial-over-LAN (SOL) settings.

**Current State Analysis**: The existing implementation at `/workspace/internal/provider/resource_cmpart_softwareimage.go` (836 lines) and test suite at `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` (640 lines) have achieved significant functionality:

- Full CRUD operations implemented with BCM API integration
- Async clone operation polling with exponential backoff
- Unknown value handling for original_image and modules
- Comprehensive test coverage (9 acceptance tests covering all user stories)
- Schema validation for required fields, SOL speed, and path format
- Import functionality using UUID passthrough
- Helper functions for null-safe attribute extraction

**Refactoring Approach**: Following Option A (Full TDD Implementation) from the specification, we will execute complete RED-GREEN-REFACTOR cycles for all 7 user stories, ensuring 100% test coverage and TDD discipline. The refactoring focuses on:

1. Systematic review of all CRUD operations against HashiCorp best practices
2. Enhancement of test coverage with additional edge cases and error scenarios
3. Validation of async operation handling and state management patterns
4. Documentation of TDD cycles and architectural decisions
5. Elimination of any remaining technical debt or anti-patterns

**Technical Approach**: Leverage the existing BCM test environment (172.21.15.254:8081) with real API calls (no mocking). Use parallel test execution where possible, following the terraform-provider-design skill patterns for optimal TDD workflows.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (BCM cluster manages image storage)
**Testing**: terraform-plugin-testing with TF_ACC=1 for acceptance tests, real BCM API at 172.21.15.254:8081
**Target Platform**: Linux (BCM cluster management server)
**Project Type**: Terraform Provider (single Go module with internal/provider structure)
**Performance Goals**:
  - Clone operations complete in <60s (includes exponential backoff polling)
  - CRUD operations return in <5s each
  - Full acceptance test suite runs in <15 minutes
  - Individual test cases complete in <2 minutes

**Constraints**:
  - BCM API has eventual consistency for clone operations (fileOperationInProgress polling required)
  - BCM API resets original_image to zero UUID after cloning (state preservation needed)
  - Kernel configuration updates require existing filesystem (two-step create+update pattern)
  - Schema validators must match BCM API expectations (SOL speed OneOf, path regex)
  - Unknown values MUST be resolved before state persistence (Terraform framework requirement)

**Scale/Scope**:
  - Single resource implementation (~800 lines Go code)
  - 11+ acceptance tests covering all user stories and edge cases
  - 25+ schema attributes (required, optional, computed)
  - 3 nested object types (KernelModule)
  - Target: 100% CRUD operation coverage with full error handling

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

Based on `/workspace/.specify/memory/constitution.md` analysis:

### Gate 1: Test-First Development (TDD)
**Status**: PASS with enhancement opportunity

**Current State**:
- Existing implementation has comprehensive test coverage (9 acceptance tests)
- Tests follow HashiCorp acceptance test patterns
- Test-driven discipline evident in test structure

**Enhancement Required**:
- Document RED-GREEN-REFACTOR cycle for each refactoring iteration
- Add missing edge case tests (async timeout, API error handling)
- Validate that all tests were written before implementation

**Remediation**: Phase 0 will identify gaps in test coverage, Phase 2 will add missing tests before any implementation changes.

### Gate 2: Simplicity Over Patterns
**Status**: PASS

**Current State**:
- Direct BCM API integration without abstraction layers
- Helper functions (buildAPIEntity, readSoftwareImage) serve clear purposes
- No repository patterns, factories, or unnecessary abstractions

**Validation**: Code follows "three strikes" rule - helper functions extracted only after repeated code patterns emerged.

### Gate 3: Clear Business Value
**Status**: PASS

**Current State**:
- Each user story maps to specific cluster management workflows
- Features directly enable node provisioning and kernel configuration
- No speculative features or premature optimization

**Validation**: All 7 user stories have clear acceptance criteria tied to BCM cluster operations.

### Gate 4: Documentation & Communication
**Status**: PASS with enhancement opportunity

**Current State**:
- Code includes comprehensive comments explaining BCM API quirks
- tflog traces at appropriate levels (Debug for operations, Info for lifecycle events)
- Schema includes MarkdownDescription for all attributes

**Enhancement Required**:
- Generate up-to-date documentation via `make generate`
- Add examples/ directory content for all use cases
- Document TDD cycle decisions in this plan

**Remediation**: Phase 1 will update examples/, Phase 2 will run documentation generation.

### Summary: All Gates PASS - Proceed to Phase 0

## Project Structure

### Documentation (this feature)

```text
specs/005-tdd-softwareimage-refactor/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output: TDD analysis, test gap identification
├── data-model.md        # Phase 1 output: BCM entity mapping, state schema
├── quickstart.md        # Phase 1 output: Developer guide for TDD cycles
├── contracts/           # Phase 1 output: BCM API contracts (JSON-RPC specs)
│   ├── cmpart_api.json     # CMPart service API methods
│   └── entity_schemas.json # SoftwareImage entity structure
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/provider/
│   ├── resource_cmpart_softwareimage.go       # REFACTOR TARGET (837 lines)
│   ├── resource_cmpart_softwareimage_test.go  # REFACTOR TARGET (640 lines)
│   ├── bcm_client.go                          # API client (supports args parameter)
│   ├── provider.go                            # Provider registration
│   └── data_source_cmpart_softwareimages.go   # Helper functions (getStringValue, etc.)
│
├── examples/resources/bcm_cmpart_softwareimage/
│   ├── resource.tf                            # Basic usage example
│   ├── import.sh                              # Import example
│   └── full-config.tf                         # Advanced configuration
│
├── docs/resources/cmpart_softwareimage.md     # Auto-generated documentation
│
├── specs/005-tdd-softwareimage-refactor/      # This feature's artifacts
│
└── sampleRest/                                # BCM API exploration scripts
    ├── CMDevice_Complete_Documentation.md
    └── BCM_API_Complete_Documentation.md
```

**Structure Decision**: Single project structure (Option 1 from template). The Terraform provider follows standard HashiCorp provider layout with `internal/provider/` for implementation, `examples/` for documentation source, and `docs/` for auto-generated provider documentation. The refactoring targets two files while leveraging existing helper functions and BCM client infrastructure.

## Complexity Tracking

> This section is intentionally left blank as no Constitution Check violations were found that require justification.

---

## Phase 0: Research & TDD Analysis

**Objective**: Analyze existing implementation against TDD best practices, identify test coverage gaps, and research BCM API behavior for edge cases.

### Research Tasks

#### R001: TDD Cycle Audit
**Goal**: Document the TDD discipline used in original implementation and identify deviations

**Method**:
1. Review git history for test commits vs implementation commits
2. Identify which tests were written first vs retrofitted
3. Document any implementation code without corresponding tests
4. Analyze test failure patterns in RED phase (were they true RED tests?)

**Expected Findings**:
- Test commit timeline relative to implementation
- Areas where implementation preceded tests (technical debt)
- Test coverage percentage by CRUD operation
- List of untested code paths (error handling, edge cases)

**Deliverable**: TDD audit report in `research.md` with coverage gaps highlighted

#### R002: BCM API Edge Case Research
**Goal**: Identify BCM API behaviors that need test coverage

**Method**:
1. Test clone operation timeout scenarios (what happens after 60s?)
2. Test concurrent clone operations (does fileOperationInProgress handle multiple clones?)
3. Test original_image with invalid UUID (what error format?)
4. Test kernel_version update on non-cloned image (is validation different?)
5. Test removeSoftwareImage with force=true (what's the difference from force=false?)
6. Test updateSoftwareImage with non-existent UUID (404 vs 500 error?)

**Expected Findings**:
- BCM API error response formats (JSON structure, error codes)
- Timeout behavior for clone operations >60s
- Validation timing for kernel configuration updates
- Error handling patterns for not-found vs validation errors

**Deliverable**: BCM API behavior matrix in `research.md` with test scenarios

#### R003: HashiCorp Provider Best Practices Review
**Goal**: Compare current implementation against terraform-provider-design skill patterns

**Method**:
1. Review schema validators (are all BCM constraints covered?)
2. Review plan modifiers (UseStateForUnknown usage correct?)
3. Review diagnostic messages (are they actionable?)
4. Review import functionality (does it follow ImportStatePassthroughID pattern?)
5. Review test patterns (are all 4 test steps present: Create, Import, Update, Delete?)
6. Review helper functions (are they reusable vs single-purpose?)

**Expected Findings**:
- Schema validator gaps (length checks, regex patterns)
- Missing plan modifiers for computed attributes
- Diagnostic message improvements (context, user actions)
- Test pattern gaps (missing CheckDestroy, missing negative tests)

**Deliverable**: Best practices comparison matrix in `research.md`

#### R004: Unknown Value Handling Analysis
**Goal**: Validate that all Unknown values are correctly resolved to known values

**Method**:
1. Trace all types.String, types.Bool, types.List attributes
2. Identify which attributes can be Unknown during plan phase
3. Verify each Unknown attribute is resolved in Create/Read/Update
4. Test plan/apply cycles with data source references (introduces Unknown values)
5. Verify modules list is never Unknown after apply

**Expected Findings**:
- List of attributes that can be Unknown during plan
- Code paths that resolve Unknown to known values
- Potential "invalid result object" error scenarios
- Test cases to validate Unknown handling

**Deliverable**: Unknown value flow diagram in `research.md`

#### R005: Async Operation Polling Strategy Review
**Goal**: Validate clone polling implementation against production workloads

**Method**:
1. Measure actual clone times for different image sizes (default-image vs larger images)
2. Test exponential backoff timing (1s, 2s, 4s, 8s, 16s = 31s total)
3. Identify maximum clone time observed in BCM environment
4. Research BCM fileOperationInProgress reliability (does it ever stick at true?)
5. Test clone failure scenarios (insufficient disk, invalid source)

**Expected Findings**:
- Actual clone time distribution (p50, p95, p99)
- Whether 31s max wait is sufficient (may need adjustment)
- Error handling for clone failures
- Polling frequency optimization opportunities

**Deliverable**: Polling strategy analysis in `research.md` with recommendations

### Phase 0 Output: research.md

**Contents**:
1. TDD Audit Summary
   - Original test-first percentage
   - Coverage gaps by user story
   - Retrofitted tests vs true RED tests

2. BCM API Behavior Matrix
   - Error response formats
   - Edge case scenarios
   - Validation timing quirks

3. Best Practices Comparison
   - Schema improvements needed
   - Diagnostic enhancements
   - Test pattern gaps

4. Unknown Value Flow
   - Attributes with Unknown during plan
   - Resolution code paths
   - Test scenarios

5. Async Polling Analysis
   - Clone time measurements
   - Backoff tuning recommendations
   - Error handling gaps

---

## Phase 1: Design & Contracts

**Objective**: Define the refined data model, API contracts, and TDD execution strategy based on Phase 0 findings.

### Design Artifacts

#### D001: data-model.md
**Purpose**: Document the complete BCM SoftwareImage entity mapping to Terraform schema

**Contents**:
1. **Entity Mapping Table**
   - BCM API field name → Terraform attribute name
   - Data types (BCM JSON → terraform-plugin-framework types)
   - Null handling strategy
   - Computed vs user-provided classification

2. **State Transitions**
   - Create: plan → BCM addSoftwareImage → state (with clone polling)
   - Read: state UUID → BCM getSoftwareImage → updated state
   - Update: state + plan → BCM updateSoftwareImage → updated state
   - Delete: state UUID → BCM removeSoftwareImage → removed from state
   - Import: UUID input → BCM getSoftwareImage → initial state

3. **Nested Object: KernelModule**
   - name: types.String (required)
   - parameters: types.String (optional, defaults to "")
   - Terraform list → BCM array transformation

4. **Special Handling Rules**
   - original_image: Plan value preserved even when BCM resets to zero UUID
   - modules: Always set to known list (empty list if API returns null)
   - kernel_version: Inherited from original_image during clone
   - fspart/bootfspart: Auto-generated by BCM during clone

5. **Validation Rules**
   - name: Length 1-255 characters
   - path: Regex `^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$`
   - sol_speed: OneOf(115200, 57600, 38400, 19200, 9600, 4800, 2400, 1200)
   - modules[].name: Length >= 1

#### D002: contracts/
**Purpose**: Define BCM API contracts for CMPart service methods

**File: contracts/cmpart_api.json**
```json
{
  "service": "CMPart",
  "methods": {
    "addSoftwareImage": {
      "args": ["entity:SoftwareImageEntity", "cloneData:boolean"],
      "returns": "string (UUID) or {updated_entity: {...}}",
      "notes": "Returns UUID immediately but clone is async (poll fileOperationInProgress)"
    },
    "getSoftwareImage": {
      "args": ["name:string"],
      "returns": "SoftwareImageEntity or null",
      "notes": "Direct lookup by name (efficient)"
    },
    "getSoftwareImages": {
      "args": [],
      "returns": "SoftwareImageEntity[]",
      "notes": "Returns all images (use for list operations)"
    },
    "updateSoftwareImage": {
      "args": ["entity:SoftwareImageEntity"],
      "returns": "void or error",
      "notes": "Entity must include uuid field"
    },
    "removeSoftwareImage": {
      "args": ["uuid:string", "removeData:boolean", "removeAll:boolean", "force:boolean"],
      "returns": "void or error",
      "notes": "Standard deletion uses false, false, false"
    }
  }
}
```

**File: contracts/entity_schemas.json**
```json
{
  "SoftwareImageEntity": {
    "required_fields": {
      "baseType": "SoftwareImage",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "name": "string",
      "path": "string"
    },
    "optional_fields": {
      "uuid": "string (only for updates)",
      "kernelVersion": "string",
      "kernelParameters": "string",
      "kernelOutputConsole": "string",
      "enableSOL": "boolean",
      "SOLPort": "string",
      "SOLSpeed": "string",
      "SOLFlowControl": "boolean",
      "notes": "string",
      "originalImage": "string (UUID, only for create)",
      "modules": "KernelModule[]"
    },
    "computed_fields": {
      "uuid": "string",
      "creationTime": "int64",
      "revisionID": "int64",
      "fileOperationInProgress": "boolean",
      "fspart": "string (UUID)",
      "bootfspart": "string (UUID)",
      "parentSoftwareImage": "string (UUID)"
    }
  },
  "KernelModule": {
    "required_fields": {
      "baseType": "KernelModule",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "name": "string"
    },
    "optional_fields": {
      "parameters": "string (defaults to empty string)"
    }
  }
}
```

#### D003: quickstart.md
**Purpose**: Developer guide for executing TDD cycles during refactoring

**Contents**:
1. **Environment Setup**
   ```bash
   export TF_ACC=1
   export BCM_ENDPOINT="https://172.21.15.254:8081"
   export BCM_USERNAME="root"
   export BCM_PASSWORD="Hashicorp123!"
   ```

2. **TDD Workflow for Each User Story**
   ```bash
   # RED Phase: Write failing test
   vim internal/provider/resource_cmpart_softwareimage_test.go
   TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAcc<UserStory>
   # Verify test fails

   # GREEN Phase: Minimal implementation
   vim internal/provider/resource_cmpart_softwareimage.go
   TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAcc<UserStory>
   # Verify test passes

   # REFACTOR Phase: Improve code quality
   vim internal/provider/resource_cmpart_softwareimage.go
   TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAcc<UserStory>
   # Verify test still passes
   ```

3. **Running Individual Tests**
   - Test US1 (Create): `go test -run TestAccCMPartSoftwareImageResource_Basic`
   - Test US2 (Read): `go test -run TestAccCMPartSoftwareImageResource_Import`
   - Test US3 (Update): `go test -run TestAccCMPartSoftwareImageResource_UpdateKernelConfig`
   - Test US4 (Delete): `go test -run TestAccCMPartSoftwareImageResource_Basic` (includes delete)
   - Test US5 (Async): `go test -run TestAccCMPartSoftwareImageResource_Basic` (includes clone)
   - Test US6 (Validation): `go test -run TestAccCMPartSoftwareImageResource_MissingRequired`
   - Test US7 (Unknown): `go test -run TestAccCMPartSoftwareImageResource_FullConfig`

4. **Parallel Test Execution**
   ```bash
   # Run all software image tests in parallel
   TF_ACC=1 go test -v -parallel=4 -timeout 120m ./internal/provider/ \
     -run "TestAccCMPartSoftwareImageResource"
   ```

5. **Documentation Generation**
   ```bash
   make generate
   # Verify docs/resources/cmpart_softwareimage.md updated
   ```

6. **Code Quality Checks**
   ```bash
   make fmt
   make lint
   pre-commit run --all-files
   ```

### Agent Context Update

After Phase 1 design completion, run:
```bash
.specify/scripts/bash/update-agent-context.sh copilot
```

This updates `.github/agents/copilot-instructions.md` with:
- BCM SoftwareImage entity patterns
- Async clone polling strategy
- Unknown value resolution patterns
- Test data strategy (unique timestamp-based names)

---

## Phase 2: TDD Execution Plan (Tasks Generation Preview)

**Note**: This section provides a preview of the task breakdown. The actual `tasks.md` file will be generated by the `/speckit.tasks` command.

### TDD Cycle Structure

Each user story follows this pattern:

1. **RED Phase**: Write failing acceptance test
   - Define test configuration function
   - Add resource.TestStep with expected assertions
   - Run test to verify failure
   - Commit failing test

2. **GREEN Phase**: Minimal implementation to pass test
   - Add only code needed to make test pass
   - No optimization, no extra features
   - Run test to verify success
   - Commit passing implementation

3. **REFACTOR Phase**: Improve code quality
   - Extract helper functions if needed
   - Add error handling
   - Improve diagnostics
   - Add logging
   - Run test to verify still passing
   - Commit refactored code

### Task Categories

#### Category 1: User Story 1 - Create Software Image (3 tasks)
- Task 1.1: RED - Write failing test for basic image creation
- Task 1.2: GREEN - Implement Create() with API integration
- Task 1.3: REFACTOR - Add clone polling and error handling

#### Category 2: User Story 2 - Read and Verify State (3 tasks)
- Task 2.1: RED - Write failing test for import
- Task 2.2: GREEN - Implement Read() with getSoftwareImage
- Task 2.3: REFACTOR - Add original_image preservation logic

#### Category 3: User Story 3 - Update Kernel Configuration (3 tasks)
- Task 3.1: RED - Write failing tests for kernel/SOL/module updates
- Task 3.2: GREEN - Implement Update() with buildAPIEntity
- Task 3.3: REFACTOR - Add update validation and state verification

#### Category 4: User Story 4 - Delete Software Image (2 tasks)
- Task 4.1: RED - Write failing test with CheckDestroy
- Task 4.2: GREEN+REFACTOR - Implement Delete() with error handling

#### Category 5: User Story 5 - Async Clone Operations (2 tasks)
- Task 5.1: RED - Write failing test for clone completion
- Task 5.2: GREEN+REFACTOR - Implement exponential backoff polling

#### Category 6: User Story 6 - Input Validation (3 tasks)
- Task 6.1: RED - Write failing tests for missing required fields
- Task 6.2: RED - Write failing tests for invalid SOL speed
- Task 6.3: RED - Write failing tests for invalid path format
- (No GREEN phase - schema validators handle this)

#### Category 7: User Story 7 - Unknown Value Handling (2 tasks)
- Task 7.1: RED - Write failing test with data source references
- Task 7.2: REFACTOR - Add Unknown resolution in Create/Read/Update

#### Category 8: Documentation and Quality (3 tasks)
- Task 8.1: Update examples/ directory with all use cases
- Task 8.2: Run make generate and verify documentation
- Task 8.3: Final test suite run with all 11+ tests

### Task Dependencies

```
Phase 0 Research (R001-R005)
  ↓
Phase 1 Design (D001-D003)
  ↓
┌─────────────────────────────────────┐
│ US1: Create (Tasks 1.1-1.3)         │ ← Must complete first
└─────────────────────────────────────┘
  ↓
┌─────────────────────────────────────┐
│ US2: Read (Tasks 2.1-2.3)           │ ← Depends on Create
└─────────────────────────────────────┘
  ↓
┌──────────────────────────────────────┬──────────────────────────────────┐
│ US3: Update (Tasks 3.1-3.3)          │ US5: Async (Tasks 5.1-5.2)       │ ← Parallel
└──────────────────────────────────────┴──────────────────────────────────┘
  ↓
┌──────────────────────────────────────┬──────────────────────────────────┐
│ US4: Delete (Tasks 4.1-4.2)          │ US6: Validation (Tasks 6.1-6.3)  │ ← Parallel
└──────────────────────────────────────┴──────────────────────────────────┘
  ↓
┌─────────────────────────────────────┐
│ US7: Unknown (Tasks 7.1-7.2)        │
└─────────────────────────────────────┘
  ↓
┌─────────────────────────────────────┐
│ Documentation (Tasks 8.1-8.3)       │ ← Final step
└─────────────────────────────────────┘
```

### Parallel Execution Opportunities

**Parallel Group 1** (after US2 complete):
- US3 Update tests (kernel, SOL, modules)
- US5 Async polling tests

**Parallel Group 2** (after US3 complete):
- US4 Delete tests
- US6 Validation tests (schema-level, no code changes)

**Final Validation** (all tests together):
```bash
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource"
```

---

## Success Criteria Validation Strategy

### Automated Validation

#### SC-001: CRUD Operation Timing
```bash
# Measure test execution time
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic \
  | grep -E "PASS|FAIL" | awk '{print $NF}'
# Target: <2 minutes per test
```

#### SC-002: Test Coverage Percentage
```bash
# Run all 11+ acceptance tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource" \
  | tee test-results.log

# Count passing tests
grep "PASS:" test-results.log | wc -l
# Target: 11+ tests passing
```

#### SC-003: Zero Test Failures
```bash
# Run full suite multiple times
for i in {1..3}; do
  TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
    -run "TestAccCMPartSoftwareImageResource" \
    || echo "FAILURE in run $i"
done
# Target: 0 failures across all runs
```

#### SC-004: Import Functionality
```bash
# Verify import in every CRUD test
grep "ImportState:" internal/provider/resource_cmpart_softwareimage_test.go \
  | grep "true" | wc -l
# Target: All resource tests have ImportState step
```

#### SC-007: Schema Validation Coverage
```bash
# Count negative validation tests
grep -E "TestAcc.*Invalid|Missing" \
  internal/provider/resource_cmpart_softwareimage_test.go \
  | wc -l
# Target: 3 validation tests (missing required, invalid SOL, invalid path)
```

#### SC-012: Documentation Auto-Generation
```bash
# Generate docs and verify no git diff
make generate
git diff --exit-code docs/resources/cmpart_softwareimage.md
# Target: 0 exit code (no changes after generation)
```

### Manual Validation

#### SC-005: State Drift Detection
**Test Procedure**:
1. Create software image via Terraform
2. Manually modify kernel_parameters via BCM GUI
3. Run `terraform plan`
4. Verify plan shows drift (kernel_parameters change detected)

**Expected Result**: Terraform detects drift and proposes update

#### SC-006: Clone Operation Reliability
**Test Procedure**:
1. Run TestAccCMPartSoftwareImageResource_Basic 10 times
2. Measure clone completion time for each run
3. Verify fileOperationInProgress polling succeeds every time

**Expected Result**: 100% success rate, all clones complete within 31s

#### SC-008: Unknown Value Compliance
**Test Procedure**:
1. Create configuration with original_image referencing data source
2. Run `terraform plan` (original_image is Unknown)
3. Run `terraform apply`
4. Verify state contains concrete UUID, not Unknown

**Expected Result**: No "invalid result object" errors

#### SC-009: Parallel Test Execution
**Test Procedure**:
```bash
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource"
```

**Expected Result**: All tests pass without name collisions (timestamp-based names prevent conflicts)

#### SC-010: TDD Discipline Verification
**Test Procedure**:
1. Review git history for each user story
2. Verify test commit precedes implementation commit
3. Document any deviations from RED-GREEN-REFACTOR

**Expected Result**: 100% of features have test-first commits

### Quality Metrics Dashboard

After implementation, generate quality report:

```bash
# Test coverage
TF_ACC=1 go test -v -cover ./internal/provider/ -run TestAccCMPartSoftwareImage

# Test execution time
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource 2>&1 | \
  grep -E "PASS|--- PASS" | awk '{print $2, $3}'

# Code quality
golangci-lint run ./internal/provider/resource_cmpart_softwareimage.go

# Documentation freshness
make generate && git diff --stat docs/
```

**Target Metrics**:
- Test coverage: 100% of user stories covered
- Test pass rate: 100% (0 failures)
- Average test time: <2 minutes per test
- Total suite time: <15 minutes
- Lint issues: 0
- Documentation drift: 0 files changed

---

## Architecture Decisions

### AD-001: Direct API Lookup vs List-and-Filter

**Decision**: Use `getSoftwareImage(name)` for Read operations instead of `getSoftwareImages()` + client-side filtering

**Rationale**:
- BCM API supports args parameter for direct lookup
- More efficient than listing all images and filtering
- Reduces API response payload size
- Matches existing data source pattern (data source uses list, resource uses direct lookup)

**Trade-offs**:
- Requires name parameter for lookup (UUID-only import needs two-step: list all, find UUID, extract name)
- Import operation is slightly more complex (but acceptable cost)

**Implementation**: Existing code already uses this pattern correctly

### AD-002: Async Clone Polling Strategy

**Decision**: Exponential backoff with 6 retries (1s, 2s, 4s, 8s, 16s = 31s total) + soft timeout

**Rationale**:
- BCM clone operation is asynchronous (addSoftwareImage returns immediately)
- fileOperationInProgress field indicates clone status
- Most clones complete in <10s (exponential backoff efficient)
- Hard timeout would fail legitimate slow clones (large images, slow storage)
- Soft timeout logs warning but proceeds (allows manual verification)

**Trade-offs**:
- 31s max wait may be insufficient for very large images (user must verify manually)
- No retry on transient errors (accepts single-attempt limitation)

**Implementation**: Existing code implements this correctly, Phase 0 research will validate adequacy

### AD-003: Unknown Value Resolution Strategy

**Decision**: Always resolve Unknown values to known values before state persistence

**Rationale**:
- Terraform framework requirement (Unknown in state causes "invalid result object" error)
- original_image: Preserve plan value if known, otherwise use API value (or null)
- modules: Always set to known list (empty list if API returns null)
- Explicit null is valid, Unknown is not

**Trade-offs**:
- Must track plan values through CRUD operations (additional state management)
- Cannot rely solely on API responses (BCM resets original_image after clone)

**Implementation**: Existing code handles this correctly in Create/Read/Update

### AD-004: Two-Step Create Pattern for Kernel Configuration

**Decision**: Create with basic config, then Update with kernel parameters

**Rationale**:
- BCM API validates kernel files exist before accepting kernel_version updates
- Clone operation creates kernel files asynchronously
- Single-step create with kernel_version fails due to timing (files not ready)
- Two-step pattern ensures kernel files exist before update

**Trade-offs**:
- Two API calls instead of one (slight performance cost)
- More complex test configurations
- Better user experience (avoids cryptic validation errors)

**Implementation**: Tests use two-step pattern, implementation supports both

### AD-005: State Preservation for original_image

**Decision**: Preserve plan's original_image value in state even when BCM API resets it

**Rationale**:
- BCM API resets original_image to zero UUID after clone completes
- Audit trail benefit: users can see which image was cloned from
- No functional impact (original_image only used during create)
- Terraform state should reflect user's intent, not just API response

**Trade-offs**:
- State diverges from API representation (intentional)
- ImportStateVerifyIgnore needed for original_image (expected)

**Implementation**: Existing code preserves plan value correctly

---

## Risk Assessment

### Risk 1: Clone Polling Timeout Inadequate
**Probability**: Medium
**Impact**: High (test failures, user frustration)
**Mitigation**:
- Phase 0 research measures actual clone times
- Adjust retry count if needed (current: 31s max)
- Soft timeout prevents hard failures
- Document expected clone times in provider docs

### Risk 2: BCM API Behavior Changes
**Probability**: Low
**Impact**: High (breaking changes)
**Mitigation**:
- Comprehensive acceptance tests detect API changes
- Tests run against real BCM instance (not mocked)
- Document API quirks in code comments
- Version pin BCM cluster for testing

### Risk 3: Parallel Test Conflicts
**Probability**: Low
**Impact**: Medium (flaky tests)
**Mitigation**:
- Unique timestamp-based resource names
- PreCheck cleanup function with deletion verification
- CheckDestroy verification
- Retry logic in cleanup functions

### Risk 4: Unknown Value Edge Cases
**Probability**: Medium
**Impact**: Medium (plan/apply errors)
**Mitigation**:
- Comprehensive test with data source references
- Explicit Unknown checks in all CRUD operations
- Always set to known values (null or concrete value)
- Phase 0 research validates all code paths

### Risk 5: Test Environment Instability
**Probability**: Low
**Impact**: High (cannot run tests)
**Mitigation**:
- Document BCM environment setup in quickstart.md
- Environment variables for endpoint configuration
- Credentials management via environment (not hardcoded)
- Fallback to manual testing if environment unavailable

---

## Monitoring & Observability

### Test Execution Monitoring

**Logging Strategy**:
- tflog.Trace: State mappings, attribute conversions
- tflog.Debug: API calls, request/response bodies
- tflog.Info: CRUD operation completion
- tflog.Warn: Clone polling timeouts, non-fatal errors

**Example Logging**:
```go
tflog.Debug(ctx, "Creating software image via BCM API", map[string]interface{}{
  "name": plan.Name.ValueString(),
  "path": plan.Path.ValueString(),
  "has_original_image": !plan.OriginalImage.IsNull(),
})
```

### Acceptance Test Metrics

**Capture per test**:
- Execution time (via go test -v output)
- API call count (via tflog.Debug grep)
- Clone polling iterations (via tflog.Debug grep)
- Error rate (via grep FAIL)

**Aggregate metrics**:
```bash
# Test execution time distribution
TF_ACC=1 go test -v -json ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource 2>&1 | \
  jq -r 'select(.Action=="pass") | .Elapsed' | \
  sort -n
```

### Performance Baselines

**Established Targets**:
- Basic create: <30s (includes clone polling)
- Read operation: <2s
- Update operation: <5s
- Delete operation: <3s
- Import operation: <5s
- Full test suite: <15 minutes

**Monitoring Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource 2>&1 | \
  grep -E "--- PASS|FAIL" | \
  awk '{print $3, $4}'
```

---

## Next Steps

After plan approval:

1. **Execute Phase 0 Research** (Estimated: 4 hours)
   - Run R001-R005 research tasks
   - Generate `research.md` with findings
   - Present findings for review

2. **Execute Phase 1 Design** (Estimated: 3 hours)
   - Create `data-model.md`
   - Create `contracts/*.json`
   - Create `quickstart.md`
   - Update agent context

3. **Generate Tasks** (Estimated: 1 hour)
   - Run `/speckit.tasks` command
   - Review task breakdown
   - Adjust task ordering if needed

4. **Execute Implementation** (Estimated: 12 hours)
   - Follow RED-GREEN-REFACTOR cycles
   - Run tests after each phase
   - Update examples/ and docs/

5. **Final Validation** (Estimated: 2 hours)
   - Run full test suite 3 times
   - Verify all success criteria
   - Generate quality metrics report

**Total Estimated Time**: 22 hours

**Acceptance Criteria**: All 15 success criteria (SC-001 through SC-015) validated and passing.
