# Tasks: Complete TDD-Based Review and Refactoring of resource_cmpart_softwareimage

**Input**: Design documents from `/workspace/specs/005-tdd-softwareimage-refactor/`
**Prerequisites**: plan.md (required), spec.md (required for user stories)

**Feature Branch**: `005-tdd-softwareimage-refactor`
**Implementation Files**:
- `/workspace/internal/provider/resource_cmpart_softwareimage.go` (836 lines)
- `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` (639 lines)

**Tests**: This is a TDD-focused refactoring with 100% test coverage requirement. All test tasks MUST be completed before corresponding implementation tasks (strict RED-GREEN-REFACTOR discipline).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story, following the terraform-provider-design skill patterns.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US7)
- Include exact file paths in descriptions

---

## Phase 0: Research & TDD Analysis

**Purpose**: Analyze existing implementation against TDD best practices and identify test coverage gaps

**Duration Estimate**: 4 hours

### Research Tasks

- [ ] R001 Conduct TDD cycle audit by reviewing git history for test-first vs implementation-first commits in `/workspace/internal/provider/resource_cmpart_softwareimage*.go`, document coverage gaps, and mark any missing test-first evidence as **BLOCKING technical debt** requiring remediation before refactoring can proceed - output results in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

- [ ] R002 [P] Execute BCM API edge case research by testing clone timeout scenarios (>60s), concurrent clones, invalid original_image UUID, kernel_version updates, and removeSoftwareImage force flags, documenting error response formats and behaviors in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

- [ ] R003 [P] Review HashiCorp provider best practices by comparing current schema validators, plan modifiers, diagnostics, import functionality, and helper functions against terraform-provider-design skill patterns, documenting gaps in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

- [ ] R004 [P] Analyze Unknown value handling by tracing all types.String, types.Bool, types.List attributes through Create/Read/Update operations and testing plan/apply cycles with data source references, documenting resolution paths in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

- [ ] R005 [P] Review async operation polling strategy by measuring actual clone times for different image sizes, testing exponential backoff timing (1s, 2s, 4s, 8s, 16s), and identifying maximum clone time, documenting recommendations in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

**Checkpoint**: Research complete - review findings before proceeding to Phase 1

---

## Phase 1: Design & Contracts

**Purpose**: Define refined data model, API contracts, and TDD execution strategy

**Duration Estimate**: 3 hours

### Design Artifacts

- [ ] D001 Create data-model.md in `/workspace/specs/005-tdd-softwareimage-refactor/data-model.md` documenting BCM SoftwareImage entity mapping, state transitions (Create/Read/Update/Delete/Import), KernelModule nested object structure, special handling rules for original_image/modules/kernel_version, and validation rules for name/path/sol_speed

- [ ] D002 [P] Create contracts directory and CMPart API specification at `/workspace/specs/005-tdd-softwareimage-refactor/contracts/cmpart_api.json` documenting addSoftwareImage, getSoftwareImage, getSoftwareImages, updateSoftwareImage, and removeSoftwareImage methods with args, returns, and notes

- [ ] D003 [P] Create entity schemas at `/workspace/specs/005-tdd-softwareimage-refactor/contracts/entity_schemas.json` documenting SoftwareImageEntity and KernelModule with required_fields, optional_fields, and computed_fields

- [ ] D004 Create quickstart.md at `/workspace/specs/005-tdd-softwareimage-refactor/quickstart.md` with environment setup (TF_ACC, BCM credentials), TDD workflow for RED-GREEN-REFACTOR cycles, individual test run commands, parallel test execution examples, documentation generation commands, and code quality checks

- [ ] D005 Update agent context by running `/workspace/.specify/scripts/bash/update-agent-context.sh copilot` to refresh `.github/agents/copilot-instructions.md` with BCM SoftwareImage patterns, async clone polling strategy, Unknown value resolution, and test data strategy

**Checkpoint**: Design artifacts complete - foundation ready for TDD execution

---

## Phase 2: User Story 1 - Create Software Image by Cloning (Priority: P1) 🎯 MVP

**Goal**: Implement create operation that clones existing BCM images with async polling for clone completion

**Independent Test**: Create Terraform config cloning "default-image" to new name/path, apply it, verify new image exists in BCM with expected attributes

**Success Criteria**: FR-001 (Create API call), FR-016 (clone polling), FR-019-FR-021 (buildAPIEntity for create)

### RED Phase: Write Failing Tests

- [ ] T001 [US1] Write failing acceptance test TestAccCMPartSoftwareImageResource_Basic in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates image with name, path, original_image and verifies uuid, creation_time, revision_id are set, run test to confirm RED state (expected: "resource not found" or "not implemented" error)

- [ ] T001a [US1] Create testAccProviderConfig helper function in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that returns provider configuration block with BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD from environment variables and insecure_skip_verify=true, ensuring all test configs include proper provider authentication

- [ ] T002 [US1] Write failing test for clone completion detection in TestAccCMPartSoftwareImageResource_Basic that verifies fileOperationInProgress becomes false after create, run test to confirm RED state (expected: test fails because polling not implemented)

- [ ] T002a [US1] **CHECKPOINT**: Verify all US1 RED tests fail with expected error messages (T001: "resource not found" or "not implemented", T002: "polling not implemented" or "fileOperationInProgress not checked") before proceeding to GREEN phase - document verification in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

### GREEN Phase: Minimal Implementation

- [ ] T003 [US1] Implement basic Create() method in `/workspace/internal/provider/resource_cmpart_softwareimage.go` that extracts plan values, builds minimal BCM entity with baseType/childType/name/path, calls CallJSONRPC("CMPart", "addSoftwareImage", entity, true), extracts UUID from response, sets UUID and ID in state, run test to verify GREEN state

- [ ] T004 [US1] Add clone polling logic to Create() method that checks fileOperationInProgress field with loop (max 6 iterations: 1s, 2s, 4s, 8s, 16s wait times) after addSoftwareImage call, proceed when false or max retries reached, run test to verify GREEN state (all assertions pass)

### REFACTOR Phase: Improve Code Quality

- [ ] T005 [US1] Extract buildAPIEntity helper function in `/workspace/internal/provider/resource_cmpart_softwareimage.go` that constructs BCM entity with baseType="SoftwareImage", childType="", modified=true, to_be_removed=false, revision="", includes UUID only for updates, includes original_image only for creates, run test to verify still GREEN

- [ ] T006 [US1] Add comprehensive error handling to Create() method for API errors, response validation (null/empty checks), UUID extraction (handle multiple formats: string, object with uuid field, object with updated_entity.uuid), add tflog.Debug for API calls and tflog.Trace for polling iterations, run test to verify still GREEN

- [ ] T007 [US1] Add tflog messages at appropriate levels: Debug for "Creating software image via BCM API" with name/path/has_original_image, Trace for each polling iteration with attempt number/wait duration/fileOperationInProgress status, Warn for max retries exceeded, Info for "created software image resource", run test to verify still GREEN

**Checkpoint**: User Story 1 complete - basic create with clone polling works

---

## Phase 3: User Story 2 - Read and Verify Image State (Priority: P1)

**Goal**: Implement read operation with direct name lookup and state preservation for original_image

**Independent Test**: Create image via BCM API, use Terraform import to read it back, verify all attributes match

**Success Criteria**: FR-002 (Read API call), FR-005 (ImportState), FR-013 (preserve original_image), FR-014 (modules as known list)

### RED Phase: Write Failing Tests

- [ ] T008 [US2] Write failing acceptance test TestAccCMPartSoftwareImageResource_Import in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that includes ImportState test step with ImportStateVerify=true, run test to confirm RED state (expected: "ImportState not implemented" error)

- [ ] T009 [US2] Write failing test for original_image preservation that creates image with original_image, reads it back, verifies state contains plan's original_image value (not BCM's zero UUID), run test to confirm RED state (expected: original_image mismatch in state)

- [ ] T010 [US2] Write failing test for modules as known list that creates image without modules, reads it back, verifies state has modules=[] (empty list, not null or Unknown), run test to confirm RED state (expected: modules is null or Unknown in state)

- [ ] T010a [US2] **CHECKPOINT**: Verify all US2 RED tests fail with expected error messages (T008: "ImportState not implemented", T009: "original_image mismatch", T010: "modules is null or Unknown") before proceeding to GREEN phase - document verification in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

### GREEN Phase: Minimal Implementation

- [ ] T011 [US2] Implement Read() method in `/workspace/internal/provider/resource_cmpart_softwareimage.go` that extracts UUID/ID from state, calls CallJSONRPC("CMPart", "getSoftwareImage", name) with name lookup, unmarshals response into map[string]interface{}, maps all BCM fields to Terraform state using helper functions (getStringValue, getBoolValue, getInt64Value), run test to verify GREEN state

- [ ] T012 [US2] Implement ImportState() method in `/workspace/internal/provider/resource_cmpart_softwareimage.go` using resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp) pattern, run test to verify GREEN state (import test passes)

### REFACTOR Phase: Improve Code Quality

- [ ] T013 [US2] Add original_image preservation logic to Read() method that checks if plan has known original_image value, preserves it in state even when BCM returns zero UUID, only overwrites if plan value was Unknown, run test to verify still GREEN

- [ ] T014 [US2] Add modules list handling to Read() method that always sets modules to known list (extract from API response if present, empty list if null/missing), ensures modules is never Unknown or null after Read(), run test to verify still GREEN

- [ ] T015 [US2] Add error handling to Read() for "not found" errors (return diagnostic without crashing), add tflog.Debug for "Reading software image from BCM API" with uuid/name, add tflog.Trace for attribute mappings, add null response validation before unmarshaling, run test to verify still GREEN

**Checkpoint**: User Story 2 complete - read and import work with state preservation

---

## Phase 4: User Story 3 - Update Kernel Configuration (Priority: P2)

**Goal**: Implement update operation for kernel parameters, SOL settings, and module lists

**Independent Test**: Create image, update kernel_parameters from empty to "quiet splash", apply, verify BCM reflects change

**Success Criteria**: FR-003 (Update API call), FR-022-FR-023 (modules array construction)

### RED Phase: Write Failing Tests

- [ ] T016 [US3] Write failing acceptance test TestAccCMPartSoftwareImageResource_UpdateKernelConfig in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates image with basic config, then updates kernel_parameters="quiet splash nomodeset", kernel_output_console="ttyS0,115200", verifies changes applied, run test to confirm RED state

- [ ] T017 [US3] Write failing acceptance test TestAccCMPartSoftwareImageResource_UpdateSOLConfig in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates image, then updates enable_sol=true, sol_speed="115200", sol_port="0", sol_flow_control=true, verifies changes applied, run test to confirm RED state

- [ ] T018 [US3] Write failing acceptance test TestAccCMPartSoftwareImageResource_UpdateModules in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates image with 2 modules, updates to 3 modules (add one), then updates to 1 module (remove two), verifies module list changes, run test to confirm RED state

- [ ] T018a [US3] **CHECKPOINT**: Verify all US3 RED tests fail with expected error messages (T016: "Update not implemented" or "kernel config unchanged", T017: "SOL config unchanged", T018: "modules unchanged") before proceeding to GREEN phase - document verification in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

### GREEN Phase: Minimal Implementation

- [ ] T019 [US3] Implement Update() method in `/workspace/internal/provider/resource_cmpart_softwareimage.go` that extracts plan values, calls buildAPIEntity with UUID included, calls CallJSONRPC("CMPart", "updateSoftwareImage", entity), calls Read() to refresh state, run test to verify GREEN state for kernel config test

- [ ] T020 [US3] Add modules array construction to buildAPIEntity in `/workspace/internal/provider/resource_cmpart_softwareimage.go` that iterates plan.Modules list, creates array of objects with baseType="KernelModule", childType="", modified=true, name and parameters fields, sets parameters="" when null, run test to verify GREEN state for modules test

### REFACTOR Phase: Improve Code Quality

- [ ] T021 [US3] Add validation to Update() that entity includes UUID field, add error handling for updateSoftwareImage API errors, add tflog.Debug for "Updating software image via BCM API" with uuid and changed attributes, add tflog.Info for "updated software image resource", run test to verify still GREEN

- [ ] T022 [US3] Optimize buildAPIEntity to accept isUpdate boolean parameter, include UUID only when isUpdate=true, include original_image only when isUpdate=false and value is set, add comprehensive null/Unknown checks for all optional fields, run test to verify still GREEN for all three update tests

**Checkpoint**: User Story 3 complete - kernel, SOL, and module updates work

---

## Phase 5: User Story 4 - Delete Software Image (Priority: P2)

**Goal**: Implement delete operation with proper cleanup verification

**Independent Test**: Create image via Terraform, run terraform destroy, verify image no longer exists in BCM

**Success Criteria**: FR-004 (Delete API call), FR-031-FR-032 (CheckDestroy function)

### RED Phase: Write Failing Tests

- [ ] T023 [US4] Write failing test for delete operation by adding CheckDestroy function to existing TestAccCMPartSoftwareImageResource_Basic test in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that verifies all test images are deleted after test completes, run test to confirm RED state (expected: "Delete not implemented" error)

- [ ] T024 [US4] Write CheckDestroy helper function testAccCheckCMPartSoftwareImageDestroy in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that iterates all resources in state, calls BCM getSoftwareImage for each, verifies response is null/empty (image deleted), run test to confirm RED state

- [ ] T024a [US4] **CHECKPOINT**: Verify all US4 RED tests fail with expected error messages (T023: "Delete not implemented", T024: "CheckDestroy not called" or "images not deleted") before proceeding to GREEN phase - document verification in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

### GREEN Phase: Minimal Implementation

- [ ] T025 [US4] Implement Delete() method in `/workspace/internal/provider/resource_cmpart_softwareimage.go` that extracts UUID from state, calls CallJSONRPC("CMPart", "removeSoftwareImage", uuid, false, false, false) with standard flags (removeData=false, removeAll=false, force=false), run test to verify GREEN state (CheckDestroy passes)

### REFACTOR Phase: Improve Code Quality

- [ ] T026 [US4] Add error handling to Delete() for API errors and "not found" cases (log warning if already deleted, don't fail), add tflog.Debug for "Deleting software image via BCM API" with uuid, add tflog.Info for "deleted software image resource", run test to verify still GREEN

**Checkpoint**: User Story 4 complete - delete with verification works

---

## Phase 6: User Story 5 - Handle Async Clone Operations (Priority: P2)

**Goal**: Validate and enhance async clone polling implementation

**Independent Test**: Create image with original_image, monitor fileOperationInProgress, verify Terraform waits until false

**Success Criteria**: FR-016-FR-018 (exponential backoff polling with logging)

### RED Phase: Write Failing Tests (Enhancement Tests)

- [ ] T027 [US5] Write failing acceptance test TestAccCMPartSoftwareImageResource_CloneWithPolling in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates image with original_image, verifies clone completes (fileOperationInProgress=false in final state), measures completion time (should be <31s for normal clones), run test to confirm current state (may pass if existing implementation works)

- [ ] T028 [US5] Write test scenario documentation in `/workspace/specs/005-tdd-softwareimage-refactor/research.md` for clone polling edge cases: fast clones (<5s), slow clones (10-30s), timeout scenario (>31s), verifying soft timeout behavior (logs warning, proceeds rather than hard-failing)

- [ ] T028a [US5] **CHECKPOINT**: Verify US5 RED test results (T027: may already pass if polling works, verify timing <31s; T028: edge cases documented) before proceeding to validation phase - document findings in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

### GREEN Phase: Validate/Fix Implementation

- [ ] T029 [US5] Review existing clone polling implementation in Create() method at `/workspace/internal/provider/resource_cmpart_softwareimage.go` lines for exponential backoff (1s, 2s, 4s, 8s, 16s), verify timing matches spec, add test run to confirm polling works correctly, document any needed fixes in research.md

### REFACTOR Phase: Improve Code Quality

- [ ] T030 [US5] Enhance polling logging in Create() method to include detailed tflog.Trace messages: "Clone operation polling attempt {N} of 6, waiting {duration}, fileOperationInProgress={status}", add tflog.Warn when max retries exceeded: "Clone operation polling timeout after 31s, proceeding with state read (manual verification recommended)", run test to verify still GREEN

**Checkpoint**: User Story 5 complete - async clone handling validated and enhanced

---

## Phase 7: User Story 6 - Validate Input Attributes (Priority: P3)

**Goal**: Implement schema validators for required fields, SOL speed, and path format

**Independent Test**: Create Terraform configs with invalid values, verify plan phase fails with appropriate errors

**Success Criteria**: FR-006-FR-009 (schema validators), FR-033 (negative tests)

### RED Phase: Write Failing Tests (Validation Tests)

- [ ] T031 [US6] Write failing acceptance test TestAccCMPartSoftwareImageResource_MissingRequired in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates config missing required name attribute, expects plan error "argument \"name\" is required", run test to confirm RED state (may pass if validator exists)

- [ ] T032 [US6] Write failing acceptance test TestAccCMPartSoftwareImageResource_InvalidSOLSpeed in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates config with sol_speed="9999", expects plan error indicating value must be one of (115200, 57600, 38400, 19200, 9600, 4800, 2400, 1200), run test to confirm RED state

- [ ] T033 [US6] Write failing acceptance test TestAccCMPartSoftwareImageResource_InvalidPath in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates config with path="invalid path with spaces", expects plan error indicating path must match regex ^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$, run test to confirm RED state

- [ ] T033a [US6] **CHECKPOINT**: Verify US6 RED tests results (T031: may pass if Required=true exists, T032: "invalid SOL speed", T033: "invalid path format") before proceeding to validator implementation - document current validation coverage in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

### GREEN Phase: Implement Validators (if needed)

- [ ] T034 [US6] Review schema definition in Schema() method at `/workspace/internal/provider/resource_cmpart_softwareimage.go` for name and path Required=true attributes, add stringvalidator.LengthBetween(1, 255) for name if missing, add stringvalidator.RegexMatches for path if missing, run test to verify GREEN state for T031 and T033

- [ ] T035 [US6] Review sol_speed schema attribute for stringvalidator.OneOf validator with values ["115200", "57600", "38400", "19200", "9600", "4800", "2400", "1200"], add if missing, run test to verify GREEN state for T032

### REFACTOR Phase: Improve Validators (if needed)

- [ ] T036 [US6] Add MarkdownDescription to all schema attributes documenting validation rules, add custom error messages to validators explaining expected format/values, run all validation tests to verify still GREEN

**Checkpoint**: User Story 6 complete - schema validation catches invalid inputs during plan

---

## Phase 8: User Story 7 - Handle Unknown Values Correctly (Priority: P3)

**Goal**: Ensure Unknown values during plan phase are never propagated to state

**Independent Test**: Create config with original_image referencing data source (Unknown during plan), run apply, verify state contains concrete UUID or null (never Unknown)

**Success Criteria**: FR-012 (no Unknown in state), FR-013 (preserve original_image), FR-014 (modules as known list)

### RED Phase: Write Failing Tests

- [ ] T037 [US7] Write failing acceptance test TestAccCMPartSoftwareImageResource_UnknownOriginalImage in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates config with original_image referencing data source lookup (introduces Unknown during plan), runs apply, verifies state contains concrete UUID (not Unknown), run test to confirm RED state (expected: "invalid result object" error or Unknown in state)

- [ ] T038 [US7] Write failing acceptance test TestAccCMPartSoftwareImageResource_UnknownModules in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` that creates image, verifies modules attribute is never Unknown (always known list, even if empty), run test to confirm RED state if Unknown handling has gaps

- [ ] T038a [US7] **CHECKPOINT**: Verify US7 RED tests fail with expected error messages (T037: "invalid result object" or "Unknown in state", T038: "modules is Unknown") before proceeding to Unknown resolution fixes - document Unknown propagation paths in `/workspace/specs/005-tdd-softwareimage-refactor/research.md`

### GREEN Phase: Fix Unknown Handling (if needed)

- [ ] T039 [US7] Review Create() method in `/workspace/internal/provider/resource_cmpart_softwareimage.go` for Unknown resolution: verify original_image uses plan value if known (even if different from API response), verify modules always set to known list (extract from API or empty list []), add explicit checks for types.IsUnknown before state setting, run test to verify GREEN state

- [ ] T040 [US7] Review Read() method for Unknown handling: verify modules field always returns known list (never null/Unknown), verify original_image preserves plan value when known, add explicit Unknown resolution before resp.State.Set(), run test to verify GREEN state

- [ ] T041 [US7] Review Update() method for Unknown handling: verify all attributes resolve to known values before state setting, verify modules list construction never produces Unknown elements, run test to verify GREEN state for both T037 and T038

### REFACTOR Phase: Add Unknown Validation

- [ ] T042 [US7] Add UseStateForUnknown plan modifier to original_image attribute in Schema() method if not present, ensuring plan value is preserved when possible, run test to verify still GREEN

- [ ] T043 [US7] Add comprehensive comments in Create/Read/Update methods explaining Unknown resolution strategy: "BCM API resets original_image to zero UUID after clone, but we preserve plan value for audit trail", "modules must always be known list, never null or Unknown per Terraform framework requirements", run test to verify still GREEN

**Checkpoint**: User Story 7 complete - Unknown values correctly resolved in all code paths

---

## Phase 9: Documentation & Quality

**Purpose**: Generate documentation, update examples, and run final validation

**Duration Estimate**: 2 hours

### Documentation Tasks

- [ ] T044 [P] Create basic usage example at `/workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf` showing minimal image creation with name, path, original_image

- [ ] T045 [P] Create advanced usage example at `/workspace/examples/resources/bcm_cmpart_softwareimage/full-config.tf` showing image with all optional attributes: kernel_version, kernel_parameters, kernel_output_console, enable_sol, sol_speed, sol_port, sol_flow_control, modules list, notes

- [ ] T046 [P] Create import example script at `/workspace/examples/resources/bcm_cmpart_softwareimage/import.sh` demonstrating terraform import bcm_cmpart_softwareimage.example <uuid>

- [ ] T046a [P] Create edge case example at `/workspace/examples/resources/bcm_cmpart_softwareimage/edge-case-two-step-create.tf` demonstrating two-step pattern: create without kernel_version first, then update with kernel_version after clone completes (documents BCM API limitation where kernel_version must be set after initial creation)

- [ ] T046b [P] Create edge case example at `/workspace/examples/resources/bcm_cmpart_softwareimage/edge-case-empty-modules.tf` demonstrating module with empty parameters field (parameters="" not null) to document BCM API requirement for empty string when no parameters provided

- [ ] T046c [P] Create edge case example at `/workspace/examples/resources/bcm_cmpart_softwareimage/edge-case-path-revision.tf` demonstrating valid path with @revision syntax (e.g., path="/cm/images/custom@123") to document advanced path format support

- [ ] T047 Run make generate from `/workspace` to auto-generate provider documentation at `/workspace/docs/resources/cmpart_softwareimage.md`, verify output is correct and matches schema

- [ ] T048 [P] Update examples/resources/bcm_cmpart_softwareimage/resource.tf with provider configuration block (endpoint, username, password, insecure_skip_verify) for copy-paste usability

- [ ] T048a [P] Update `/workspace/CLAUDE.md` and `/workspace/AGENTS.md` with refactoring insights discovered during implementation: async polling patterns with exponential backoff, Unknown value resolution strategies (UseStateForUnknown, explicit checks), two-step create pattern for kernel validation, BCM API quirks (original_image reset, modules empty string requirement), edge cases and their solutions

### Quality Assurance Tasks

- [ ] T049 Run make fmt from `/workspace` to format all Go code with gofmt standards

- [ ] T050 Run make lint from `/workspace` to execute golangci-lint and verify 0 issues for resource_cmpart_softwareimage.go and test file

- [ ] T051 Run pre-commit run --all-files from `/workspace` to execute all pre-commit hooks and verify passing

### Final Validation Tasks

- [ ] T052 Run full acceptance test suite with TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMPartSoftwareImageResource from `/workspace` and verify all 11+ tests pass (target: 0 failures)

- [ ] T053 Run parallel acceptance tests with TF_ACC=1 go test -v -parallel=4 -timeout 120m /workspace/internal/provider/ -run TestAccCMPartSoftwareImageResource from `/workspace` to verify no test name collisions and consistent pass rate

- [ ] T054 Execute three consecutive full test suite runs from `/workspace` to verify test stability and 100% pass rate across multiple runs (no flaky tests)

- [ ] T055 Generate test coverage report by running TF_ACC=1 go test -v -cover /workspace/internal/provider/ -run TestAccCMPartSoftwareImageResource from `/workspace` and verify 100% of user stories covered (all acceptance scenarios have passing tests)

### Success Criteria Validation

- [ ] T056 Validate SC-001 through SC-015 by running validation scripts from plan.md section "Success Criteria Validation Strategy", documenting results in `/workspace/specs/005-tdd-softwareimage-refactor/validation-report.md`

**Checkpoint**: All documentation generated, all quality checks pass, all success criteria validated

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 0: Research (R001-R005)
  ↓ (BLOCKS all subsequent phases)
Phase 1: Design (D001-D005)
  ↓ (BLOCKS all implementation)
┌─────────────────────────────────────────────────────────────┐
│ Phase 2: US1 - Create (T001-T007) ← MUST complete first    │
└─────────────────────────────────────────────────────────────┘
  ↓ (Create is prerequisite for all other CRUD operations)
┌─────────────────────────────────────────────────────────────┐
│ Phase 3: US2 - Read (T008-T015) ← Depends on Create        │
└─────────────────────────────────────────────────────────────┘
  ↓ (Read needed for state verification in subsequent phases)
┌──────────────────────────────┬──────────────────────────────┐
│ Phase 4: US3 - Update        │ Phase 6: US5 - Async Clone   │
│         (T016-T022)          │         (T027-T030)          │ ← Parallel
└──────────────────────────────┴──────────────────────────────┘
  ↓ (Update and Async can proceed independently)
┌──────────────────────────────┬──────────────────────────────┐
│ Phase 5: US4 - Delete        │ Phase 7: US6 - Validation    │
│         (T023-T026)          │         (T031-T036)          │ ← Parallel
└──────────────────────────────┴──────────────────────────────┘
  ↓ (Delete and Validation can proceed independently)
┌─────────────────────────────────────────────────────────────┐
│ Phase 8: US7 - Unknown (T037-T043)                          │
└─────────────────────────────────────────────────────────────┘
  ↓ (All CRUD operations complete, now validate edge cases)
┌─────────────────────────────────────────────────────────────┐
│ Phase 9: Documentation & Quality (T044-T056)                │
└─────────────────────────────────────────────────────────────┘
```

### User Story Dependencies

- **US1 (Create)**: No dependencies - can start immediately after Phase 1
- **US2 (Read)**: Depends on US1 complete - cannot read without ability to create
- **US3 (Update)**: Depends on US1+US2 complete - needs create and read for update testing
- **US4 (Delete)**: Depends on US1+US2 complete - needs create and read for delete testing
- **US5 (Async)**: Depends on US1+US2 complete - validates create's clone polling
- **US6 (Validation)**: Depends on US2 complete - schema validators only, can run after read works
- **US7 (Unknown)**: Depends on US1+US2+US3 complete - tests all CRUD paths for Unknown handling
- **Documentation**: Depends on ALL user stories complete

### RED-GREEN-REFACTOR Discipline

Within each user story phase:

1. **RED Phase Tasks**: Write ALL failing tests first, run to verify RED state
2. **GREEN Phase Tasks**: Implement minimal code to pass tests, verify GREEN state
3. **REFACTOR Phase Tasks**: Improve code quality while maintaining GREEN state

**CRITICAL**: Never proceed to GREEN phase until RED tests are written and failing. Never proceed to REFACTOR until tests are GREEN.

### Parallel Execution Opportunities

**Phase 0 Research** - All research tasks R002-R005 can run in parallel:
- R002: BCM API edge cases
- R003: Best practices review
- R004: Unknown value analysis
- R005: Async polling review

**Phase 1 Design** - All design tasks D002-D005 can run in parallel after D001 (data-model):
- D002: CMPart API contracts
- D003: Entity schemas
- D004: Quickstart guide
- D005: Agent context update

**Parallel Group 1** (after US2 complete):
- US3 Update implementation (T016-T022)
- US5 Async validation (T027-T030)

**Parallel Group 2** (after US3/US5 complete):
- US4 Delete implementation (T023-T026)
- US6 Validation tests (T031-T036)

**Phase 9 Documentation** - Most documentation tasks can run in parallel:
- T044: Basic example
- T045: Full config example
- T046: Import example
- T048: Provider config in examples

---

## Parallel Example: Phase 0 Research

```bash
# Launch all parallel research tasks together (R002-R005):
# Terminal 1: BCM API edge case testing
BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go run /workspace/sampleRest/test_clone_timeout.py

# Terminal 2: Best practices review
cd /workspace && grep -r "stringvalidator" internal/provider/resource_cmpart_softwareimage.go

# Terminal 3: Unknown value analysis
cd /workspace && grep -r "IsUnknown\|Unknown" internal/provider/resource_cmpart_softwareimage.go

# Terminal 4: Async polling measurements
TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic
```

---

## Parallel Example: User Story 3 (Update)

```bash
# Launch all RED phase tests together (T016-T018):
# All three test functions can be written in parallel since they test different scenarios

# Then run all tests together to verify RED state:
TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource_Update"

# Expected: All three update tests fail (RED state confirmed)
```

---

## Implementation Strategy

### Recommended Approach: Sequential User Stories with Parallel Tasks

Given this is a refactoring of existing code (not greenfield development), follow this strategy:

1. **Phase 0: Research** (4 hours)
   - Run all R001-R005 research tasks in parallel
   - Document findings in research.md
   - Review findings before proceeding

2. **Phase 1: Design** (3 hours)
   - Create data-model.md first (D001)
   - Run D002-D005 in parallel after D001
   - Update agent context (D005)

3. **Phase 2-3: Foundation** (4 hours)
   - Complete US1 (Create) fully: RED → GREEN → REFACTOR
   - Complete US2 (Read) fully: RED → GREEN → REFACTOR
   - Validate both stories work together

4. **Phase 4-6: Parallel User Stories** (6 hours)
   - Parallel track 1: US3 (Update) + US5 (Async validation)
   - Parallel track 2: US4 (Delete) + US6 (Validation)
   - Each track completes independently

5. **Phase 7-8: Edge Cases** (2 hours)
   - Complete US7 (Unknown handling)
   - Validate all edge cases documented in spec.md

6. **Phase 9: Documentation & Quality** (2 hours)
   - Run all documentation tasks in parallel
   - Execute quality checks sequentially
   - Run final validation

**Total Estimated Time**: 21 hours

### MVP Definition

For this refactoring project, MVP = Phase 2-3 complete:
- US1 (Create) fully functional with clone polling
- US2 (Read) fully functional with import
- Basic CRUD cycle works end-to-end
- All tests passing for these two stories

This delivers a working resource suitable for testing before adding update/delete/validation.

---

## Task Summary

**Total Tasks**: 68 tasks across 10 phases (12 tasks added during analysis review)

### Task Distribution by Phase

- **Phase 0 (Research)**: 5 tasks (R001-R005) - 4 parallel
- **Phase 1 (Design)**: 5 tasks (D001-D005) - 4 parallel after D001
- **Phase 2 (US1 - Create)**: 9 tasks (T001-T007 + T001a, T002a) - Sequential RED-GREEN-REFACTOR with checkpoint
- **Phase 3 (US2 - Read)**: 9 tasks (T008-T015 + T010a) - Sequential RED-GREEN-REFACTOR with checkpoint
- **Phase 4 (US3 - Update)**: 8 tasks (T016-T022 + T018a) - Sequential RED-GREEN-REFACTOR with checkpoint
- **Phase 5 (US4 - Delete)**: 5 tasks (T023-T026 + T024a) - Sequential RED-GREEN-REFACTOR with checkpoint
- **Phase 6 (US5 - Async)**: 5 tasks (T027-T030 + T028a) - Sequential validation with checkpoint
- **Phase 7 (US6 - Validation)**: 7 tasks (T031-T036 + T033a) - Sequential RED-GREEN with checkpoint
- **Phase 8 (US7 - Unknown)**: 8 tasks (T037-T043 + T038a) - Sequential RED-GREEN-REFACTOR with checkpoint
- **Phase 9 (Documentation)**: 17 tasks (T044-T056 + T046a-c, T048a) - 11 parallel possible

### Task Distribution by User Story

- **US1 (Create)**: 9 tasks - Foundation for all others (includes provider config helper + checkpoint)
- **US2 (Read)**: 9 tasks - Required for import and state verification (includes checkpoint)
- **US3 (Update)**: 8 tasks - Kernel/SOL/module updates (includes checkpoint)
- **US4 (Delete)**: 5 tasks - Resource cleanup (includes checkpoint)
- **US5 (Async)**: 5 tasks - Clone polling validation (includes checkpoint)
- **US6 (Validation)**: 7 tasks - Schema validators (includes checkpoint)
- **US7 (Unknown)**: 8 tasks - Edge case handling (includes checkpoint)
- **Infrastructure**: 22 tasks (Research + Design + Documentation + edge case examples + project docs update)

### Parallel Opportunities

- **Phase 0**: 4 tasks can run in parallel (R002-R005)
- **Phase 1**: 4 tasks can run in parallel (D002-D005 after D001)
- **Phase 4+6**: 2 user stories can run in parallel (US3 + US5)
- **Phase 5+7**: 2 user stories can run in parallel (US4 + US6)
- **Phase 9**: 11 documentation tasks can run in parallel (T044-T046c, T048, T048a)

**Maximum Parallelism**: Up to 4 tasks simultaneously during research/design phases

---

## Notes

- **[P] marker**: Indicates tasks that can run in parallel with other [P] tasks in the same phase
- **[Story] label**: Maps task to specific user story for traceability and progress tracking
- **File paths**: All paths are absolute (starting with /workspace) for clarity
- **TDD discipline**: RED-GREEN-REFACTOR cycle is mandatory - tests must fail before implementation, pass after minimal code, and remain green during refactoring
- **Test environment**: All acceptance tests require TF_ACC=1 and BCM credentials (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- **Unique naming**: All test resources use timestamp suffixes to prevent collisions during parallel test runs
- **Validation checkpoints**: Each user story phase ends with a checkpoint to validate independent functionality before proceeding

---

## Success Validation Checklist

After completing all tasks, verify these success criteria from spec.md:

- [ ] SC-001: All CRUD operations complete in <2 minutes per acceptance test
- [ ] SC-002: 100% test coverage - all 7 user stories have passing acceptance tests
- [ ] SC-003: Zero test failures across 3 consecutive full test suite runs
- [ ] SC-004: 100% import functionality - all resources can be imported via UUID
- [ ] SC-005: State drift detection works - manual BCM changes detected by terraform refresh
- [ ] SC-006: Clone operations 100% reliable - fileOperationInProgress polling succeeds every time
- [ ] SC-007: Schema validation catches 100% of invalid inputs during plan phase
- [ ] SC-008: Unknown value handling 100% compliant - zero "invalid result object" errors
- [ ] SC-009: Parallel test execution completes without name collisions
- [ ] SC-010: TDD discipline maintained - all implementation has tests written first
- [ ] SC-011: Code follows HashiCorp Terraform Plugin Framework best practices
- [ ] SC-012: Documentation auto-generated correctly via make generate
- [ ] SC-013: Error messages are actionable with sufficient context
- [ ] SC-014: Comprehensive logging at appropriate levels (Trace/Debug/Info/Warn)
- [ ] SC-015: Code is maintainable with well-structured helper functions

**Final Deliverable**: Fully refactored resource_cmpart_softwareimage with 100% test coverage, TDD discipline documented, and all success criteria validated.
