---
description: "Task list for bcm_cmpart_softwareimage Terraform Resource implementation"
---

# Tasks: bcm_cmpart_softwareimage Resource

**Feature**: Terraform resource for managing BCM software images (OS images for DPU node provisioning)

**Input**: Design documents from `/workspace/specs/003-bcm-cmpart-softwareimage-resource/`
- `spec.md` - Feature requirements and API contracts
- `plan.md` - Implementation design
- `quickstart.md` - Developer quick start guide

**Architecture**: Terraform Plugin Framework v1.16.1, TDD with RED-GREEN-REFACTOR cycles

**API Integration**: BCM JSON-RPC API (CMPart service) with cookie authentication

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1=Create, US2=Read, US3=Update, US4=Delete, US5=Import)
- Include exact file paths in descriptions

## User Story Reference

From spec.md:
- **US1**: Create Software Image
- **US2**: Read Software Image (detect drift)
- **US3**: Update Software Image
- **US4**: Delete Software Image
- **US5**: Import Existing Images

---

## Phase 0: API Research & Discovery

**Purpose**: Discover Update and Delete API methods before implementation begins

**⚠️ CRITICAL**: This phase MUST be completed before Phase 1. The Update and Delete API methods are currently unknown and must be discovered.

**Research Goals**:
1. Find Update API method name and signature (likely `updateSoftwareImage` or `modifySoftwareImage`)
2. Find Delete API method name and signature (likely `removeSoftwareImage`)
3. Understand CRUD operation semantics and error responses
4. Test unique constraint enforcement (duplicate name/path)
5. Verify nested modules management (inline with entity)

### API Method Discovery

- [ ] T001 Search BCM API documentation in `/workspace/sampleRest/` for update/delete methods
- [ ] T002 Check existing CMPart data source implementation at `/workspace/internal/provider/data_source_cmpart_softwareimages.go` for patterns
- [ ] T003 Review BCM API complete documentation at `/workspace/sampleRest/BCM_API_Complete_Documentation.md` for CMPart methods
- [ ] T004 Create Python test script in `/workspace/specs/003-bcm-cmpart-softwareimage-resource/test_api_methods.py` to test candidate API methods
- [ ] T005 Test update API candidates: `updateSoftwareImage`, `modifySoftwareImage`, `setSoftwareImage` against BCM endpoint
- [ ] T006 Test delete API candidates: `removeSoftwareImage`, `deleteSoftwareImage`, `removeSoftwareImageByUUID` against BCM endpoint
- [ ] T007 Document confirmed API method names and signatures in research notes

### CRUD Lifecycle Testing

- [ ] T008 Create test software image via API (`addSoftwareImage`) with basic config (name, path, kernel version)
- [ ] T009 Read created image via API (`getSoftwareImages`) and verify all fields returned
- [ ] T010 Update test image using discovered update method - modify kernel parameters and modules list
- [ ] T011 Read updated image and verify changes persisted
- [ ] T012 Delete test image using discovered delete method
- [ ] T013 Verify deletion by attempting to read deleted image (should not be found)
- [ ] T014 Document complete CRUD flow with example requests/responses

### Constraint Validation Testing

- [ ] T015 Create test image with unique name `test-constraint-001`
- [ ] T016 Attempt to create second image with same name `test-constraint-001` and capture error response
- [ ] T017 Attempt to create image with same path and capture error response
- [ ] T018 Document error response format for unique constraint violations
- [ ] T019 Clean up test images created during constraint testing

### Nested Modules Management

- [ ] T020 Create test image with modules list containing 2 kernel modules
- [ ] T021 Read image and verify modules list structure (UUID generation, parameters handling)
- [ ] T022 Update image to add 1 module and remove 1 module
- [ ] T023 Verify module updates work inline (no separate API calls required)
- [ ] T024 Test empty modules list behavior
- [ ] T025 Document module management pattern for implementation

### Research Documentation

- [ ] T026 Create `/workspace/specs/003-bcm-cmpart-softwareimage-resource/research.md` with sections:
  - API Method Reference (Create, Read, Update, Delete with confirmed method names)
  - Entity Lifecycle (CRUD flow with example requests/responses)
  - Constraint Enforcement (unique name/path validation and error formats)
  - Nested Resource Patterns (modules inline management)
  - Design Decisions (force parameter default, read strategy, import approach)
- [ ] T027 Add example API request/response pairs for all CRUD operations to research.md
- [ ] T028 Document any API quirks, edge cases, or limitations discovered during testing

**Checkpoint**: Research complete - Update and Delete API methods confirmed and documented. Ready to begin TDD implementation.

---

## Phase 1: Setup & Foundation

### Validation Tasks (Added from Analysis)

- [ ] T029 [P] Add path regex validator to resource schema in `/workspace/internal/provider/resource_cmpart_softwareimage.go`
  - Use `stringvalidator.RegexMatches()` with pattern `^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$`
  - Add helpful error message: "path must match format: /cm/images/name"
- [ ] T030 [P] Add SOL speed enum validator to resource schema
  - Use `stringvalidator.OneOf()` with valid speeds: "115200", "57600", "38400", "19200", "9600", "4800", "2400", "1200"
  - Add error message listing valid options
- [ ] T031 [P] Add module name non-empty validator to nested module schema
  - Use `stringvalidator.LengthAtLeast(1)` for module name attribute
  - Add error message: "module name cannot be empty"
- [ ] T032 [P] Add name length validator to resource schema
  - Use `stringvalidator.LengthBetween(1, 255)` for name attribute
  - Add error message: "name must be between 1 and 255 characters"

---

## Phase 1: Setup & Foundation (Original)

**Purpose**: Terraform provider infrastructure setup (shared across all user stories)

**Dependencies**: Phase 0 must be complete

### Test Framework Setup

- [ ] T029 Create acceptance test file `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` with package declaration and imports
- [ ] T030 [P] Define test helper functions: `testAccCMPartSoftwareImageResourceConfig_Basic()`, `testAccCMPartSoftwareImageResourceConfig_Full()`, `testAccCMPartSoftwareImageResourceConfig_Modules()`
- [ ] T031 [P] Define test configuration generator functions that include provider config with BCM credentials

### Example Configuration Setup

- [ ] T032 Create examples directory `/workspace/examples/resources/bcm_cmpart_softwareimage/`
- [ ] T033 [P] Create basic example in `/workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf`
- [ ] T034 [P] Create advanced example with full config in `/workspace/examples/resources/bcm_cmpart_softwareimage/resource-advanced.tf`
- [ ] T035 [P] Create import example in `/workspace/examples/resources/bcm_cmpart_softwareimage/import.sh`

**Checkpoint**: Test infrastructure and examples ready for TDD RED phase.

---

## Phase 2: RED Phase - Failing Acceptance Tests

**Purpose**: Write comprehensive failing tests BEFORE implementation (TDD First Law)

**Dependencies**: Phase 1 complete

**⚠️ CRITICAL TDD RULE**: All tests in this phase MUST FAIL when first run. This proves tests are valid.

### US1: Create Software Image - Tests

- [ ] T036 [P] [US1] Write `TestAccCMPartSoftwareImageResource_Basic` in `resource_cmpart_softwareimage_test.go` testing minimal create (name, path only)
- [ ] T037 [P] [US1] Write `TestAccCMPartSoftwareImageResource_FullConfig` testing create with all optional attributes (kernel config, SOL settings, modules, notes)
- [ ] T038 [P] [US1] Write `TestAccCMPartSoftwareImageResource_WithModules` testing create with nested kernel modules list

### US2: Read Software Image - Tests

- [ ] T039 [P] [US2] Add state drift detection check to `TestAccCMPartSoftwareImageResource_Basic` (verify Read operation fetches current state)
- [ ] T040 [P] [US2] Add computed field checks to all tests (uuid, creation_time, revision_id, file_operation_in_progress)

### US3: Update Software Image - Tests

- [ ] T041 [P] [US3] Write `TestAccCMPartSoftwareImageResource_UpdateKernelConfig` testing kernel_version and kernel_parameters updates
- [ ] T042 [P] [US3] Write `TestAccCMPartSoftwareImageResource_UpdateModules` testing modules list add/remove/modify operations
- [ ] T043 [P] [US3] Write `TestAccCMPartSoftwareImageResource_UpdateSOL` testing SOL settings updates (enable_sol, sol_speed, etc)

### US4: Delete Software Image - Tests

- [ ] T044 [US4] Add delete verification to all test cases (automatic in Terraform test framework - verify no errors)

### US5: Import Existing Images - Tests

- [ ] T045 [P] [US5] Add ImportState test step to `TestAccCMPartSoftwareImageResource_Basic` with ImportStateVerify
- [ ] T046 [P] [US5] Add ImportState test step to `TestAccCMPartSoftwareImageResource_FullConfig` with ImportStateVerifyIgnore for force parameter

### Negative Test Coverage (Added from Analysis)

- [ ] T046a [P] Write `TestAccCMPartSoftwareImageResource_MissingRequired` testing validation when name or path is missing
  - Expect: Plan-time error "attribute is required"
- [ ] T046b [P] Write `TestAccCMPartSoftwareImageResource_InvalidSOLSpeed` testing invalid SOL speed value
  - Use value "9999" or "invalid" to trigger enum validation
  - Expect: Plan-time error listing valid SOL speed options
- [ ] T046c [P] Write `TestAccCMPartSoftwareImageResource_InvalidPath` testing path regex validation
  - Use invalid path "invalid path with spaces" to trigger regex validation
  - Expect: Plan-time error "path must match format: /cm/images/name"

### RED Phase Verification

- [ ] T047 Run acceptance tests with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource` and verify ALL tests FAIL with "resource type not found" error
- [ ] T048 Document test failure output confirming RED phase success

**Checkpoint**: All tests written and failing as expected. Ready for GREEN phase minimal implementation.

---

## Phase 3: GREEN Phase - Minimal Hardcoded Implementation

**Purpose**: Make tests PASS with minimal hardcoded implementation (TDD Second Law)

**Dependencies**: Phase 2 complete (RED tests exist and fail)

**⚠️ CRITICAL TDD RULE**: Implement ONLY enough code to make tests pass. No real API calls yet.

### Resource Scaffold

- [ ] T049 Create resource file `/workspace/internal/provider/resource_cmpart_softwareimage.go` with package and imports
- [ ] T050 Define `CMPartSoftwareImageResource` struct with BCMClient field
- [ ] T051 Define `CMPartSoftwareImageResourceModel` struct with all schema attributes using `types.String`, `types.Bool`, `types.Int64`, etc
- [ ] T052 Define `KernelModuleResourceModel` struct for nested modules

### US1: Create - Minimal Implementation

- [ ] T053 [US1] Implement `Metadata()` method returning resource type name `bcm_cmpart_softwareimage`
- [ ] T054 [US1] Implement `Schema()` method with complete attribute definitions (all required, optional, computed fields)
  - Include validators added in T029-T032 (path regex, SOL speed enum, name length, module name non-empty)
  - Add PlanModifiers: `stringplanmodifier.UseStateForUnknown()` for id and uuid attributes
- [ ] T055 [US1] Implement `Configure()` method to receive BCMClient from provider
- [ ] T056 [US1] Implement `Create()` method with HARDCODED success response:
  - Set ID to "hardcoded-uuid-12345"
  - Set UUID to "hardcoded-uuid-12345"
  - Set creation_time to 1700000000
  - Set revision_id to 0
  - Accept all plan values as-is
  - NO API calls in this phase

### US2: Read - Minimal Implementation

- [ ] T057 [US2] Implement `Read()` method as NO-OP (state already has data from Create, just return it unchanged)

### US3: Update - Minimal Implementation

- [ ] T058 [US3] Implement `Update()` method as NO-OP (accept plan values and set to state without API calls)

### US4: Delete - Minimal Implementation

- [ ] T059 [US4] Implement `Delete()` method as NO-OP (no API call, just return success)

### US5: Import - Minimal Implementation

- [ ] T060 [US5] Implement `ImportState()` method using `resource.ImportStatePassthroughID` for ID passthrough

### Resource Registration

- [ ] T061 Register `NewCMPartSoftwareImageResource` in provider's `Resources()` method in `/workspace/internal/provider/provider.go`

### GREEN Phase Verification

- [ ] T062 Run acceptance tests with `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic` and verify tests PASS
- [ ] T063 Run full test suite `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource` and verify ALL tests PASS
- [ ] T064 Document test success output confirming GREEN phase complete

**Checkpoint**: All tests passing with minimal hardcoded implementation. Ready for REFACTOR phase with real API integration.

---

## Phase 4: REFACTOR Phase - Full API Integration

**Purpose**: Replace hardcoded implementation with real BCM API calls (TDD Third Law)

**Dependencies**: Phase 3 complete (GREEN tests passing with stubs)

**⚠️ CRITICAL TDD RULE**: Tests must REMAIN PASSING throughout refactoring. Run tests after each task.

### US1: Create - API Integration

- [ ] T065 [US1] Refactor `Create()` to build API entity from plan:
  - Map Terraform attributes to BCM entity fields (name, path, kernelVersion, kernelParameters, etc)
  - Build modules list with baseType, childType, name, parameters, modified, to_be_removed fields
  - Handle optional fields (fspart, bootfspart, notes)
- [ ] T066 [US1] Refactor `Create()` to call BCM API:
  - **STEP 1**: Call `validateSoftwareImage(entity)` for pre-flight validation (per user request in research.md)
    - Marshal validation request: service="CMPart", call="validateSoftwareImage", args=[entity]
    - If validation fails, return error diagnostic with validation message
  - **STEP 2**: Call `addSoftwareImage(entity, force)` to create the resource
    - Marshal request body with service="CMPart", call="addSoftwareImage", args=[entity, force]
    - Execute HTTP POST to `/json` endpoint using BCMClient (requires args parameter support)
    - Parse response to extract UUID (handle both string response and object response)
  - Add comprehensive error handling with Diagnostics API for both validation and creation failures
- [ ] T067 [US1] Refactor `Create()` to read back created resource:
  - Set ID and UUID from API response
  - Call internal `readSoftwareImage()` helper to populate all computed fields
  - Verify creation_time, revision_id, and other computed fields populated
- [ ] T068 [US1] Run `TestAccCMPartSoftwareImageResource_Basic` and verify PASSING with real API
- [ ] T069 [US1] Run `TestAccCMPartSoftwareImageResource_FullConfig` and verify PASSING with real API

### US2: Read - API Integration

- [ ] T070 [US2] Create helper method `readSoftwareImage(ctx, model, diags)` in resource file:
  - Call `getSoftwareImage(name)` BCM API method using `client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", name)`
  - NOTE: Verify args parameter support in BCMClient (fallback to direct HTTP POST if needed)
  - Parse response as single software image entity (not array)
  - Check for null or empty response (indicates image not found)
  - Return error if image not found (triggers recreation in Terraform)
- [ ] T071 [US2] Implement field mapping in `readSoftwareImage()`:
  - Map all API fields to model using helper functions: `getStringValue()`, `getBoolValue()`, `getInt64Value()`
  - Map identity fields: uuid, name, path
  - Map kernel config: kernelVersion, kernelParameters, kernelOutputConsole
  - Map SOL settings: enableSOL, SOLPort, SOLSpeed, SOLFlowControl
  - Map partitions: fspart, bootfspart
  - Map metadata: notes, creationTime, revisionID, fileOperationInProgress, originalImage, parentSoftwareImage
- [ ] T072 [US2] Implement modules list parsing in `readSoftwareImage()`:
  - Parse modules array from API response
  - Create KernelModuleResourceModel for each module
  - Map name and parameters fields using getStringValue()
  - Handle empty modules list
- [ ] T073 [US2] Refactor `Read()` method to call `readSoftwareImage()` helper
- [ ] T074 [US2] Run `TestAccCMPartSoftwareImageResource_Basic` and verify state refresh works correctly

### US3: Update - API Integration

- [ ] T075 [US3] Refactor `Update()` to build updated entity from plan:
  - Include UUID in entity (required for update)
  - Map all updatable fields from plan to entity
  - Build updated modules list from plan
  - Set modified=true flag on entity
- [ ] T076 [US3] Refactor `Update()` to call BCM API using discovered update method from Phase 0 research:
  - **STEP 1**: Call `validateSoftwareImage(entity)` for pre-flight validation (per user request in research.md)
    - Marshal validation request: service="CMPart", call="validateSoftwareImage", args=[entity]
    - If validation fails, return error diagnostic with validation message
  - **STEP 2**: Call `updateSoftwareImage(entity, force)` to update the resource (confirmed in Phase 0)
    - Marshal request body with service="CMPart", call="updateSoftwareImage", args=[entity, force]
    - Execute HTTP POST to `/json` endpoint using BCMClient (requires args parameter support)
    - Parse response (may be boolean success or updated entity)
  - Handle API errors with clear diagnostic messages for both validation and update failures
- [ ] T077 [US3] Refactor `Update()` to read back updated resource:
  - Call `readSoftwareImage()` to fetch updated state from API
  - Verify changes persisted
  - Set state with updated values
- [ ] T078 [US3] Run `TestAccCMPartSoftwareImageResource_UpdateKernelConfig` and verify PASSING
- [ ] T079 [US3] Run `TestAccCMPartSoftwareImageResource_UpdateModules` and verify PASSING
- [ ] T080 [US3] Run `TestAccCMPartSoftwareImageResource_UpdateSOL` and verify PASSING

### US4: Delete - API Integration

- [ ] T081 [US4] Refactor `Delete()` to call BCM API using discovered delete method from Phase 0 research:
  - Use confirmed delete method name (e.g., `removeSoftwareImage`)
  - Marshal request body with service="CMPart", call=[confirmed_method], args=[uuid] or args=[name]
  - Execute HTTP POST to `/json` endpoint
  - Handle 404 errors gracefully (already deleted = success)
  - Add error handling for delete failures
- [ ] T082 [US4] Run full test suite to verify delete works in all test cases
- [ ] T083 [US4] Verify no state remains after delete operations

### US5: Import - API Integration

- [ ] T084 [US5] Verify `ImportState()` works with `ImportStatePassthroughID`:
  - Import flow: terraform import bcm_cmpart_softwareimage.test <uuid>
  - UUID passed to Read() method via ID
  - Read() calls readSoftwareImage() to fetch full state
- [ ] T085 [US5] Run `TestAccCMPartSoftwareImageResource_Basic` ImportState step and verify PASSING
- [ ] T086 [US5] Run `TestAccCMPartSoftwareImageResource_FullConfig` ImportState step and verify PASSING

### Error Handling Enhancement

- [ ] T087 [P] Add error handling for duplicate name constraint violation in Create() with clear diagnostic message
- [ ] T088 [P] Add error handling for duplicate path constraint violation in Create() with clear diagnostic message
- [ ] T089 [P] Add error handling for image not found in Read() (removes from state, triggers recreation)
- [ ] T090 [P] Add error handling for API unavailable scenarios with connection error diagnostics
- [ ] T091 [P] Add error handling for authentication failures with credential error diagnostics

### REFACTOR Phase Verification

- [ ] T092 Run complete acceptance test suite `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource` and verify 100% PASS rate
- [ ] T093 Run acceptance tests multiple times to verify idempotency (no spurious diffs)
- [ ] T094 Verify test resource cleanup (no leftover test images in BCM)
- [ ] T095 Document refactor completion with test results summary

**Checkpoint**: Full API integration complete. All CRUD operations working with real BCM API. Tests remain passing.

---

## Phase 5: Documentation Generation

**Purpose**: Generate Terraform provider documentation and validate examples

**Dependencies**: Phase 4 complete (implementation and tests passing)

### Documentation Generation

- [ ] T096 Set Go environment variables for documentation generation:
  ```bash
  export GOMODCACHE=/workspace/.go/pkg/mod
  export GOCACHE=/workspace/.go/cache
  export GOPATH=/workspace/.go
  ```
- [ ] T097 Run Terraform docs generator with `/workspace/.go/bin/tfplugindocs generate --provider-name bcm --tf-version 1.13.5`
- [ ] T098 Verify generated documentation file exists at `/workspace/docs/resources/cmpart_softwareimage.md`
- [ ] T099 Review generated documentation for completeness:
  - Resource description present
  - All attributes documented with descriptions
  - Required vs optional vs computed clearly marked
  - Default values documented
  - Example configurations included
  - Import example shown

### Example Validation

- [ ] T100 [P] Verify basic example in `/workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf` is syntactically valid HCL
- [ ] T101 [P] Verify advanced example in `/workspace/examples/resources/bcm_cmpart_softwareimage/resource-advanced.tf` is syntactically valid HCL
- [ ] T102 [P] Run `terraform fmt` on all example files to ensure proper formatting
- [ ] T103 [P] Add comments to examples explaining each configuration option

### Schema Documentation Enhancement

- [ ] T104 Review schema attribute descriptions in `resource_cmpart_softwareimage.go` Schema() method for clarity
- [ ] T105 Add usage guidance to MarkdownDescription for complex attributes (modules, path format, sol_speed enum values)
- [ ] T106 Verify default values are documented in MarkdownDescription strings

### Documentation Verification

- [ ] T107 Run documentation generation again with `make generate` to ensure examples are included
- [ ] T108 Verify docs contain correct provider name, resource name, and attribute types
- [ ] T109 Check that import command format is correct: `terraform import bcm_cmpart_softwareimage.example <uuid>`

**Checkpoint**: Documentation complete and verified. Resource fully documented for users.

---

## Phase 6: Code Quality & Polish

**Purpose**: Ensure code quality, formatting, and compliance with standards

**Dependencies**: Phase 5 complete

### Code Formatting

- [ ] T110 Run `make fmt` or `gofmt -s -w -e .` to format all Go files
- [ ] T111 Run `terraform fmt -recursive examples/` to format all example HCL files
- [ ] T112 Verify no formatting changes required (code should already be formatted from development)

### Linting

- [ ] T113 Run `make lint` or `golangci-lint run` to check for code issues
- [ ] T114 Fix any linting errors reported in resource implementation
- [ ] T115 Fix any linting errors reported in test files
- [ ] T116 Verify 0 linting errors remain

### Unit Tests (if applicable)

- [ ] T117 [P] Consider adding unit tests for helper functions in `/workspace/internal/provider/resource_cmpart_softwareimage_test.go`:
  - Test module list transformation
  - Test entity building logic
  - Test error handling paths
- [ ] T118 [P] Run unit tests with `go test -v -cover ./internal/provider/` and verify coverage

### Pre-commit Hooks

- [ ] T119 Run `pre-commit run --all-files` to execute all pre-commit hooks
- [ ] T120 Fix any issues reported by pre-commit hooks (formatting, trailing whitespace, etc)
- [ ] T121 Verify all pre-commit hooks pass

### Integration Testing

- [ ] T122 Run full provider test suite with `make test` (unit tests, 120s timeout)
- [ ] T123 Run full acceptance test suite with `make testacc` (TF_ACC=1, 120m timeout)
- [ ] T124 Verify 100% test pass rate for both unit and acceptance tests

### Manual Smoke Test

- [ ] T125 Build provider locally with `go install`
- [ ] T126 Create manual test configuration in `/tmp/test-softwareimage.tf` using basic example
- [ ] T127 Run `terraform init` in test directory
- [ ] T128 Run `terraform plan` and verify plan output is correct
- [ ] T129 Run `terraform apply` and verify resource created in BCM
- [ ] T130 Run `terraform import` on existing image and verify import works
- [ ] T131 Modify configuration and run `terraform apply` to test update
- [ ] T132 Run `terraform destroy` and verify resource deleted from BCM
- [ ] T133 Document manual test results

**Checkpoint**: Code quality verified. All tests passing. Manual validation complete.

---

## Phase 7: Research Validation & Cleanup

**Purpose**: Validate research findings were correctly implemented and clean up test resources

**Dependencies**: Phase 6 complete

### Research Validation

- [ ] T134 Compare implementation in `resource_cmpart_softwareimage.go` against research.md findings:
  - Verify Update API method matches research discovery
  - Verify Delete API method matches research discovery
  - Verify constraint error handling matches research findings
  - Verify module management matches research patterns
- [ ] T135 Update research.md with any implementation insights or corrections discovered during development
- [ ] T136 Add implementation notes section to research.md documenting final design decisions

### Test Resource Cleanup

- [ ] T137 Run cleanup script to remove any orphaned test software images from BCM:
  ```bash
  # Find all test images with names starting with "test-"
  # Delete using discovered delete API method
  ```
- [ ] T138 Verify BCM cluster is clean of test resources
- [ ] T139 Document cleanup procedure in research.md for future reference

### Final Verification

- [ ] T140 Re-run full acceptance test suite one final time: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource`
- [ ] T141 Verify 100% pass rate with no flaky tests
- [ ] T142 Check test execution time is under 120 minutes
- [ ] T143 Verify no test images remain after test completion (cleanup working)

**Checkpoint**: Implementation validated against research. All tests passing. No test artifacts remain.

---

## Dependencies & Execution Order

### Phase Dependencies (Sequential)

1. **Phase 0: API Research** → MUST complete first (blocks all other work)
2. **Phase 1: Setup** → Depends on Phase 0 (needs confirmed API methods for test design)
3. **Phase 2: RED** → Depends on Phase 1 (needs test infrastructure)
4. **Phase 3: GREEN** → Depends on Phase 2 (needs failing tests to make pass)
5. **Phase 4: REFACTOR** → Depends on Phase 3 (needs passing tests to maintain)
6. **Phase 5: Documentation** → Depends on Phase 4 (needs working implementation)
7. **Phase 6: Quality** → Depends on Phase 5 (needs complete implementation)
8. **Phase 7: Validation** → Depends on Phase 6 (final verification)

### Within Phase Parallelization

**Phase 0 (Research)**:
- T001-T007 can run in parallel (different API methods)
- T015-T019 constraint tests can run in parallel with T020-T025 module tests

**Phase 1 (Setup)**:
- T030-T031 test helpers can run in parallel
- T033-T035 examples can run in parallel (different files)

**Phase 2 (RED)**:
- T036-T038 US1 tests can run in parallel (different test functions)
- T039-T040 US2 tests can run in parallel
- T041-T043 US3 tests can run in parallel
- T045-T046 US5 tests can run in parallel

**Phase 4 (REFACTOR)**:
- T087-T091 error handling can run in parallel (different error types)

**Phase 6 (Quality)**:
- T110-T112 formatting can run in parallel
- T117-T118 unit tests can run in parallel with linting

### User Story Dependencies

- **US1 (Create)**: No dependencies - foundation for all other stories
- **US2 (Read)**: Depends on US1 (needs created resources to read)
- **US3 (Update)**: Depends on US1 and US2 (needs create and read working)
- **US4 (Delete)**: Depends on US1 (needs created resources to delete)
- **US5 (Import)**: Depends on US2 (uses Read operation to fetch state)

**Implementation Flow**:
US1 → US2 → US3/US4/US5 (last three can be done in any order after US1+US2)

---

## Parallel Execution Examples

### Phase 0: API Discovery (Parallel Research)

```bash
# Launch parallel research threads:
Thread 1: "Search BCM API documentation" (T001-T003)
Thread 2: "Test update API methods" (T005)
Thread 3: "Test delete API methods" (T006)
Thread 4: "Test constraint validation" (T015-T019)
Thread 5: "Test module management" (T020-T025)
```

### Phase 2: RED Phase (Parallel Test Writing)

```bash
# Launch all US1 tests together:
Task: "Write TestAccCMPartSoftwareImageResource_Basic" (T036)
Task: "Write TestAccCMPartSoftwareImageResource_FullConfig" (T037)
Task: "Write TestAccCMPartSoftwareImageResource_WithModules" (T038)

# Launch all US3 tests together:
Task: "Write TestAccCMPartSoftwareImageResource_UpdateKernelConfig" (T041)
Task: "Write TestAccCMPartSoftwareImageResource_UpdateModules" (T042)
Task: "Write TestAccCMPartSoftwareImageResource_UpdateSOL" (T043)
```

### Phase 4: Error Handling (Parallel Implementation)

```bash
# Launch all error handlers together:
Task: "Add duplicate name error handling" (T087)
Task: "Add duplicate path error handling" (T088)
Task: "Add not found error handling" (T089)
Task: "Add API unavailable error handling" (T090)
Task: "Add authentication error handling" (T091)
```

---

## Implementation Strategy

### Recommended Approach: Sequential TDD Cycles

1. **Week 1: Research** (Phase 0)
   - Complete API discovery and testing
   - Document findings in research.md
   - Validate all CRUD operations work manually

2. **Week 2: TDD Implementation** (Phases 1-4)
   - Day 1: Setup (Phase 1) → RED (Phase 2)
   - Day 2: GREEN (Phase 3) → Verify all tests pass with stubs
   - Day 3-4: REFACTOR (Phase 4) → Integrate real API calls
   - Day 5: Testing and debugging

3. **Week 3: Polish** (Phases 5-7)
   - Day 1: Documentation generation and validation
   - Day 2: Code quality, linting, formatting
   - Day 3: Manual testing and validation
   - Day 4: Research validation and cleanup
   - Day 5: Buffer for issues and final verification

### Alternative: Fast-Track MVP

Focus on basic CRUD only (minimal attributes):

1. Phase 0: Research (required)
2. Phase 1: Setup (required)
3. Phase 2 RED: T036 only (basic test)
4. Phase 3 GREEN: T049-T061 (basic implementation)
5. Phase 4 REFACTOR: T065-T083 (CRUD only, skip error handling)
6. Phase 5: Documentation
7. Phase 6: Basic quality (fmt, lint only)

**MVP Delivery**: Basic software image resource in 3-4 days

Then iterate to add:
- Full attribute support (kernel config, SOL, modules)
- Comprehensive error handling
- Advanced test scenarios

---

## Success Criteria

### Functional Requirements ✅

- [ ] Resource creates software images via `addSoftwareImage` API
- [ ] Resource reads software images via `getSoftwareImages` with UUID filtering
- [ ] Resource updates software images via discovered update method
- [ ] Resource deletes software images via discovered delete method
- [ ] ImportState functionality works with UUID identifier
- [ ] Nested modules list properly managed in state
- [ ] All computed fields populated from API responses (UUID, creation_time, revision_id, etc)
- [ ] Unique constraints (name, path) validated with clear error messages

### Testing Requirements ✅

- [ ] All acceptance tests pass with `TF_ACC=1`
- [ ] Test coverage includes:
  - Basic resource creation (minimal config)
  - Full configuration with all attributes
  - Module list updates (add/remove/modify)
  - ImportState verification
  - Update operations (kernel, modules, SOL)
  - Delete operations
- [ ] Tests use unique resource names to avoid conflicts
- [ ] Tests include provider configuration in HCL
- [ ] Tests complete within 120 minute timeout
- [ ] No flaky tests (100% pass rate across multiple runs)

### Code Quality Requirements ✅

- [ ] Follows HashiCorp provider development best practices
- [ ] Uses terraform-plugin-framework patterns consistently
- [ ] Implements null-safe helper functions for field extraction
- [ ] Includes comprehensive error handling with Diagnostics API
- [ ] Passes golangci-lint with no errors
- [ ] Code formatted with gofmt (no changes needed)
- [ ] Pre-commit hooks pass
- [ ] TDD principles followed (RED → GREEN → REFACTOR cycle documented)

### Documentation Requirements ✅

- [ ] Resource documentation auto-generated in `docs/resources/cmpart_softwareimage.md`
- [ ] Example configurations in `examples/resources/bcm_cmpart_softwareimage/`
- [ ] Schema descriptions clear and accurate
- [ ] Attribute defaults documented
- [ ] Computed vs required attributes clearly marked
- [ ] Import command documented with example

### Research Requirements ✅

- [ ] Update API method discovered and documented
- [ ] Delete API method discovered and documented
- [ ] CRUD lifecycle tested and documented
- [ ] Constraint enforcement understood and documented
- [ ] Module management pattern validated and documented

---

## Verification Commands

```bash
# Phase 0: Manual API testing
cd /workspace/specs/003-bcm-cmpart-softwareimage-resource
python test_api_methods.py

# Phase 2: Verify RED (tests should FAIL)
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic

# Phase 3: Verify GREEN (tests should PASS with stubs)
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource

# Phase 4: Verify REFACTOR (tests should PASS with real API)
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource

# Phase 5: Generate documentation
cd /workspace
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
/workspace/.go/bin/tfplugindocs generate --provider-name bcm --tf-version 1.13.5

# Phase 6: Code quality checks
cd /workspace
make fmt
make lint
pre-commit run --all-files
make test
make testacc

# Manual smoke test
cd /workspace
go install
cd /tmp
# Create test-softwareimage.tf with example config
terraform init
terraform plan
terraform apply -auto-approve
terraform import bcm_cmpart_softwareimage.imported <uuid>
terraform destroy -auto-approve
```

---

## Risk Mitigation

| Risk | Impact | Mitigation | Task Reference |
|------|--------|------------|----------------|
| Update/Delete API methods unknown | High | ✅ RESOLVED: Confirmed in Phase 0 (updateSoftwareImage, removeSoftwareImage) | T001-T007 |
| Nested modules complicate state | Medium | Follow existing data source pattern | T020-T025, T072 |
| Unique constraints not API-enforced | Medium | Test constraint violations, document behavior | T015-T019, T087-T088 |
| Force parameter semantics unclear | Low | Research default behavior, expose as optional | T008-T014 |
| BCMClient args parameter support | Medium | Test in Phase 1, fallback to direct HTTP POST if needed | T070-T074 |
| API response differs from docs | Medium | Comprehensive error handling + response logging | T087-T091 |

---

## Notes

- **[P] markers**: Tasks that can run in parallel (different files, no dependencies)
- **[Story] labels**: Map tasks to user stories for traceability (US1, US2, US3, US4, US5)
- **TDD Discipline**: Must maintain RED → GREEN → REFACTOR discipline for code quality
- **Test-First**: All tests must be written and fail before implementation begins
- **Idempotency**: All operations must be idempotent (multiple applies cause no changes)
- **Cleanup**: Test resources must be cleaned up after tests complete
- **Commit Frequency**: Commit after each phase completion or logical task group
- **Phase Checkpoints**: Stop at each checkpoint to validate before proceeding

---

## File Reference

| File Path | Purpose | Created In |
|-----------|---------|------------|
| `/workspace/internal/provider/resource_cmpart_softwareimage.go` | Resource implementation | T049 |
| `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` | Acceptance tests | T029 |
| `/workspace/internal/provider/provider.go` | Resource registration | T061 |
| `/workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf` | Basic example | T033 |
| `/workspace/examples/resources/bcm_cmpart_softwareimage/resource-advanced.tf` | Advanced example | T034 |
| `/workspace/examples/resources/bcm_cmpart_softwareimage/import.sh` | Import example | T035 |
| `/workspace/docs/resources/cmpart_softwareimage.md` | Generated documentation | T098 |
| `/workspace/specs/003-bcm-cmpart-softwareimage-resource/research.md` | Research findings | T026 |
| `/workspace/specs/003-bcm-cmpart-softwareimage-resource/test_api_methods.py` | API test script | T004 |

---

## Total Task Count

- **Phase 0 (Research)**: 28 tasks
- **Phase 1 (Setup)**: 11 tasks (7 original + 4 validation tasks from analysis)
- **Phase 2 (RED)**: 16 tasks (13 original + 3 negative tests from analysis)
- **Phase 3 (GREEN)**: 16 tasks
- **Phase 4 (REFACTOR)**: 31 tasks
- **Phase 5 (Documentation)**: 14 tasks
- **Phase 6 (Quality)**: 24 tasks
- **Phase 7 (Validation)**: 10 tasks

**Total**: 150 tasks (143 original + 7 added from speckit.analyze recommendations)

**Estimated Effort**:
- Phase 0: 8-12 hours (research and API discovery)
- Phase 1: 2-3 hours (setup)
- Phase 2: 4-6 hours (write failing tests)
- Phase 3: 4-6 hours (minimal implementation)
- Phase 4: 12-16 hours (full API integration)
- Phase 5: 2-3 hours (documentation)
- Phase 6: 4-6 hours (quality and testing)
- Phase 7: 2-3 hours (validation)

**Total Estimated Time**: 38-55 hours (5-7 working days)

---

## Next Steps After Completion

1. **Integration Testing**: Test resource with actual BCM cluster production workflows
2. **Performance Testing**: Benchmark large module lists and bulk operations
3. **Error Handling Enhancement**: Add retry logic for transient API failures
4. **Related Resources**: Implement filesystem partition resources (fspart, bootfspart)
5. **Data Source Alignment**: Ensure resource and data source use same model structs
6. **User Feedback**: Gather feedback from BCM users on attribute naming and defaults
7. **Advanced Features**: Add support for software image cloning, versioning, rollback

---

## References

- **Feature Specification**: `/workspace/specs/003-bcm-cmpart-softwareimage-resource/spec.md`
- **Implementation Plan**: `/workspace/specs/003-bcm-cmpart-softwareimage-resource/plan.md`
- **Quick Start Guide**: `/workspace/specs/003-bcm-cmpart-softwareimage-resource/quickstart.md`
- **BCM API Documentation**: `/workspace/sampleRest/BCM_API_Complete_Documentation.md`
- **Existing Data Source**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
- **TDD Constitution**: `/workspace/.specify/memory/constitution.md`
- **Provider TDD Guide**: `/workspace/AGENTS.md`
- **Terraform Plugin Framework**: https://developer.hashicorp.com/terraform/plugin/framework
- **Terraform Plugin Testing**: https://developer.hashicorp.com/terraform/plugin/testing
