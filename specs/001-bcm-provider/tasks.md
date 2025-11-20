# Tasks: BCM Terraform Provider POV

**Input**: Design documents from `/workspace/specs/001-bcm-provider/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/

**Tests**: This feature includes acceptance tests following Terraform Provider TDD methodology (RED-GREEN-REFACTOR cycles).

**Organization**: Tasks are grouped by user story to enable independent implementation and testing of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2)
- Include exact file paths in descriptions

## Path Conventions

- **Terraform Provider Structure**: `internal/provider/`, `examples/`, `docs/`
- All paths are absolute from repository root `/workspace/`

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure for BCM provider

- [X] T001 Verify Go module dependencies in /workspace/go.mod (terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3)
- [X] T002 [P] Create provider test setup in /workspace/internal/provider/provider_test.go with testAccProtoV6ProviderFactories and testAccPreCheck
- [X] T003 [P] Configure environment variables for acceptance tests (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core BCM client infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T004 Create BCM client types structure in /workspace/internal/provider/bcm_client.go (BCMClient struct, JSONRPCRequest struct)
- [X] T005 Implement NewBCMClient constructor with TLS configuration in /workspace/internal/provider/bcm_client.go
- [X] T006 [P] Implement base64EncodeCredentials helper function in /workspace/internal/provider/bcm_client.go
- [X] T007 [P] Create unit test file /workspace/internal/provider/bcm_client_test.go with test setup
- [X] T007a [P] Document JSONRPCRequest struct excludes arg field (POV limitation - parameter passing out of scope) in /workspace/internal/provider/bcm_client.go code comments

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Provider Authentication with Cookie-Based Auth (Priority: P0) 🎯 MVP

**Goal**: Configure BCM provider with token-based cookie authentication credentials and handle self-signed certificates so that JSON-RPC API calls can be executed

**Independent Test**: Configure provider block with credentials (root:Hashicorp123!) and `insecure_skip_verify = true`, execute a simple JSON-RPC call to verify authentication and connectivity

### Tests for User Story 1 - RED Phase (Write Failing Tests First) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T008 [P] [US1] Unit test: Provider schema validation in /workspace/internal/provider/provider_test.go (verify endpoint, username, password, insecure_skip_verify, timeout attributes exist)
- [X] T009 [P] [US1] Unit test: Provider Configure success with valid credentials in /workspace/internal/provider/provider_test.go
- [X] T010 [P] [US1] Unit test: Provider Configure failure with invalid endpoint in /workspace/internal/provider/provider_test.go
- [X] T011 [P] [US1] Unit test: TLS configuration with insecure_skip_verify in /workspace/internal/provider/provider_test.go
- [X] T012 [P] [US1] Unit test: Login API call request construction in /workspace/internal/provider/bcm_client_test.go (verify POST body {"service":"login","username":"...","password":"..."})
- [X] T013 [P] [US1] Unit test: Cookie jar creation and Set-Cookie header parsing in /workspace/internal/provider/bcm_client_test.go
- [X] T014 [P] [US1] Unit test: Authentication token storage in HTTP client cookie jar in /workspace/internal/provider/bcm_client_test.go

**Run Tests**: `go test ./internal/provider -v -run "TestProvider|TestBCMClient"` - All tests should FAIL

### Implementation for User Story 1 - GREEN Phase (Minimal Implementation)

- [X] T015 [US1] Update provider schema in /workspace/internal/provider/provider.go (add endpoint, username, password, insecure_skip_verify, timeout attributes)
- [X] T016 [US1] Update ScaffoldingProviderModel struct in /workspace/internal/provider/provider.go with new fields
- [X] T017 [US1] Change provider TypeName from "scaffolding" to "bcm" in /workspace/internal/provider/provider.go
- [X] T018 [US1] Implement provider Configure method in /workspace/internal/provider/provider.go (extract config, validate required fields, create BCMClient with cookie jar, perform login API call, store client in resp.DataSourceData)
- [X] T019 [US1] Implement login API call method in /workspace/internal/provider/bcm_client.go (POST {"service":"login","username":"...","password":"..."} to /json endpoint, parse Set-Cookie header with cm-login-token)
- [X] T020 [US1] Implement cookie jar initialization in NewBCMClient in /workspace/internal/provider/bcm_client.go (create http.Client with cookiejar.New())
- [X] T021 [US1] Add HTTP client timeout configuration in NewBCMClient in /workspace/internal/provider/bcm_client.go
- [X] T021a [P] [US1] Unit test: Login API call failure handling (HTTP 401/403 returns clear authentication error) in /workspace/internal/provider/bcm_client_test.go

**Run Tests**: `go test ./internal/provider -v -run "TestProvider|TestBCMClient"` - All tests should PASS

### Refactor for User Story 1 - REFACTOR Phase (Production-Ready Code)

- [X] T022 [US1] Add comprehensive error handling for missing required fields in provider Configure in /workspace/internal/provider/provider.go
- [X] T023 [US1] Add TLS configuration error guidance messages in /workspace/internal/provider/provider.go
- [X] T024 [US1] Add login failure error handling with clear "authentication failed" message in /workspace/internal/provider/bcm_client.go
- [X] T025 [US1] Add tflog.Debug for client initialization in /workspace/internal/provider/bcm_client.go
- [X] T026 [US1] Add tflog.Trace for login API call request/response (excluding password) in /workspace/internal/provider/bcm_client.go
- [X] T027 [US1] Add context cancellation handling for login timeout in /workspace/internal/provider/bcm_client.go
- [X] T028 [US1] Validate Set-Cookie header presence after login and return error if missing in /workspace/internal/provider/bcm_client.go
- [X] T028a [US1] Validate Set-Cookie header format (cm-login-token cookie name, check Secure/httponly/path attributes, log warning if attributes missing) in /workspace/internal/provider/bcm_client.go

**Run Tests**: `go test ./internal/provider -v -run "TestProvider|TestBCMClient"` - All tests should still PASS with improvements

**Checkpoint**: At this point, User Story 1 (Provider Authentication) should be fully functional and testable independently

---

## Phase 4: User Story 2 - CMPart SoftwareImages Data Source (Priority: P0)

**Goal**: Query BCM software images through a Terraform data source so that available OS images and their kernel module configurations can be discovered

**Independent Test**: Define `data "bcm_cmpart_softwareimages" "all"` data source, run terraform plan/apply, verify all software images returned with complete schema including nested modules array

### Tests for User Story 2 - RED Phase (Write Failing Tests First) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T029 [P] [US2] Acceptance test: Basic data source read in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go (verify data source exists, images attribute accessible)
- [X] T030 [P] [US2] Acceptance test: Empty response handling in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go (verify empty array returns empty list, not error)
- [X] T031 [P] [US2] Acceptance test: Nested modules array parsing in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go (verify modules list exists and module fields accessible)
- [X] T032 [P] [US2] Acceptance test: All 30+ fields validation in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go (check snake_case fields: kernel_version, enable_sol, etc.)
- [X] T033 [P] [US2] Acceptance test: Authentication failure handling in /workspace/internal/provider/data_source_cmpart_softwareimages_test.go (invalid credentials return clear error)
- [X] T034 [P] [US2] Unit test: JSON-RPC request construction for getSoftwareImages in /workspace/internal/provider/bcm_client_test.go
- [X] T035 [P] [US2] Unit test: HTTP headers for API call (Content-Type, Cookie with cm-login-token) in /workspace/internal/provider/bcm_client_test.go
- [X] T036 [P] [US2] Unit test: Cookie jar automatic inclusion of authentication token in /workspace/internal/provider/bcm_client_test.go

**Run Tests**: `TF_ACC=1 go test ./internal/provider -v -run "TestAccCMPartSoftwareImages" -timeout 120m` - All tests should FAIL

### Implementation for User Story 2 - GREEN Phase (Minimal Implementation)

- [X] T037 [US2] Create data source file /workspace/internal/provider/data_source_cmpart_softwareimages.go with CMPartSoftwareImagesDataSource type
- [X] T038 [US2] Create model structs in /workspace/internal/provider/data_source_cmpart_softwareimages.go (CMPartSoftwareImagesDataSourceModel, SoftwareImageModel with 30+ fields, KernelModuleModel with 8 fields)
- [X] T039 [US2] Implement complete schema with ListNestedAttribute for images in /workspace/internal/provider/data_source_cmpart_softwareimages.go (all attributes Computed: true)
- [X] T040 [US2] Implement nested modules schema within SoftwareImage as ListNestedAttribute in /workspace/internal/provider/data_source_cmpart_softwareimages.go
- [X] T041 [US2] Implement CallJSONRPC method in /workspace/internal/provider/bcm_client.go (construct POST request to endpoint/json with {"service":"CMPart","call":"getSoftwareImages"}, set Content-Type header, use cookie jar for automatic Cookie header)
- [X] T042 [US2] Implement defensive error parsing in /workspace/internal/provider/bcm_client.go (Layer 1: HTTP status check, Layer 2: JSON error object detection, Layer 3: empty array success, Layer 4: parse errors)
- [X] T043 [US2] Implement Read method in /workspace/internal/provider/data_source_cmpart_softwareimages.go (get client, call getSoftwareImages API, parse JSON array response)
- [X] T044 [US2] Implement field mapping from API response to models in /workspace/internal/provider/data_source_cmpart_softwareimages.go (camelCase → snake_case conversion, handle null fields with types.StringNull())
- [X] T045 [US2] Implement nested modules array mapping in /workspace/internal/provider/data_source_cmpart_softwareimages.go
- [X] T046 [US2] Register data source in provider DataSources method in /workspace/internal/provider/provider.go

**Run Tests**: `TF_ACC=1 go test ./internal/provider -v -run "TestAccCMPartSoftwareImages" -timeout 120m` - All tests should PASS

### Refactor for User Story 2 - REFACTOR Phase (Production-Ready Code)

- [X] T047 [US2] Extract camelCase→snake_case field mapping to helper functions in /workspace/internal/provider/data_source_cmpart_softwareimages.go (getStringValue, getBoolValue, getInt64Value, mapAPIResponseToModel)
- [X] T048 [US2] Add comprehensive null handling for all optional fields in /workspace/internal/provider/data_source_cmpart_softwareimages.go
- [X] T049 [US2] Add error messages with API call details (service, call) in /workspace/internal/provider/data_source_cmpart_softwareimages.go
- [X] T050 [US2] Add tflog.Debug for successful API calls with image count in /workspace/internal/provider/data_source_cmpart_softwareimages.go
- [X] T051 [US2] Add tflog.Trace for full JSON-RPC request/response logging in /workspace/internal/provider/bcm_client.go
- [X] T052 [US2] Add response structure validation before unmarshaling in /workspace/internal/provider/data_source_cmpart_softwareimages.go
- [X] T053 [US2] Add error messages with response snippet (first 500 chars) for unexpected formats in /workspace/internal/provider/bcm_client.go
- [X] T054 [US2] Add connection error detection (connection refused, timeout, DNS) with specific guidance in /workspace/internal/provider/bcm_client.go

**Run Tests**: `TF_ACC=1 go test ./internal/provider -v -timeout 120m` - All tests should still PASS with improvements

**Checkpoint**: At this point, User Story 2 (Data Source) should be fully functional and testable independently. Both user stories (authentication + data source) work together.

---

## Phase 5: Documentation & Examples

**Purpose**: Generate provider documentation and working examples for users

- [X] T055 [P] Create provider configuration example in /workspace/examples/provider/provider.tf (with endpoint, username, password, insecure_skip_verify)
- [X] T056 [P] Create data source example in /workspace/examples/data-sources/bcm_cmpart_softwareimages/data-source.tf (with output examples for filtering)
- [X] T057 Generate provider documentation with tfplugindocs: `make generate` (creates /workspace/docs/index.md and /workspace/docs/data-sources/cmpart_softwareimages.md)
- [X] T058 Verify generated documentation in /workspace/docs/ includes all schema fields and working examples
- [X] T059 [P] Update /workspace/specs/001-bcm-provider/quickstart.md with provider installation instructions and troubleshooting guide
- [X] T060 [P] Validate examples with `terraform validate` in /workspace/examples/

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final improvements that affect both user stories

- [X] T061 [P] Run golangci-lint on all provider code: `make lint`
- [X] T062 [P] Run go fmt on all provider code: `make fmt`
- [X] T063 Verify acceptance test coverage for all scenarios (success, empty response, auth failure, nested attributes, all fields)
- [X] T064 [P] Add code comments for exported types and functions in /workspace/internal/provider/
- [X] T065 Run full acceptance test suite: `TF_ACC=1 make testacc`
- [X] T066 Verify unit test coverage >= 80%: `go test -cover ./internal/provider`
- [X] T067 [P] Create README section documenting POV scope and post-POV roadmap
- [X] T068 Run quickstart.md validation (manual testing of provider setup and data source usage)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User Story 1 (Authentication) must complete before User Story 2 (Data Source) can be tested end-to-end
  - However, User Story 2 tests can be written in parallel with User Story 1 implementation
- **Documentation (Phase 5)**: Depends on User Story 1 and User Story 2 completion
- **Polish (Phase 6)**: Depends on all implementation and documentation being complete

### User Story Dependencies

- **User Story 1 (Authentication - P0)**: Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (Data Source - P0)**: Functionally depends on User Story 1 for authentication, but tests can be written in parallel

### Within Each User Story (TDD Cycles)

**RED Phase**:
- Tests MUST be written and FAIL before implementation
- All tests for a story can be written in parallel (marked [P])

**GREEN Phase**:
- Minimal implementation to make tests pass
- Schema before Read implementation
- Models before field mapping
- Client methods before data source integration

**REFACTOR Phase**:
- Error handling improvements
- Logging enhancements
- Code quality improvements
- All refactors can run in parallel (marked [P]) if affecting different files

### Parallel Opportunities

**Phase 1 (Setup)**:
- T002 and T003 can run in parallel

**Phase 2 (Foundational)**:
- T006 and T007 can run in parallel after T004-T005 complete

**User Story 1 - RED Phase**:
- T008, T009, T010, T011 (provider tests) can all run in parallel
- T012, T013, T014 (client tests) can all run in parallel

**User Story 1 - REFACTOR Phase**:
- T022-T028 can run in parallel (different functions/concerns)

**User Story 2 - RED Phase**:
- T029, T030, T031, T032, T033 (acceptance tests) can all run in parallel
- T034, T035, T036 (client tests) can all run in parallel

**User Story 2 - REFACTOR Phase**:
- T047-T054 can run in parallel (different functions/concerns)

**Phase 5 (Documentation)**:
- T055, T056, T059, T060 can all run in parallel

**Phase 6 (Polish)**:
- T061, T062, T064, T067 can all run in parallel

---

## Parallel Example: User Story 1 - RED Phase

```bash
# Launch all provider schema tests together:
Task T008: "Unit test: Provider schema validation"
Task T009: "Unit test: Provider Configure success"
Task T010: "Unit test: Provider Configure failure"
Task T011: "Unit test: TLS configuration"

# In parallel, launch all client auth tests:
Task T012: "Unit test: Login API call request construction"
Task T013: "Unit test: Cookie jar creation"
Task T014: "Unit test: Authentication token storage"
```

## Parallel Example: User Story 2 - RED Phase

```bash
# Launch all acceptance tests together:
Task T029: "Acceptance test: Basic data source read"
Task T030: "Acceptance test: Empty response handling"
Task T031: "Acceptance test: Nested modules array parsing"
Task T032: "Acceptance test: All 30+ fields validation"
Task T033: "Acceptance test: Authentication failure handling"

# In parallel, launch all client tests:
Task T034: "Unit test: JSON-RPC request construction"
Task T035: "Unit test: HTTP headers"
Task T036: "Unit test: Cookie jar automatic inclusion"
```

---

## Implementation Strategy

### MVP First (Both User Stories - POV Scope)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (Authentication)
4. **STOP and VALIDATE**: Test authentication independently with manual curl commands
5. Complete Phase 4: User Story 2 (Data Source)
6. **STOP and VALIDATE**: Test data source independently with terraform plan/apply
7. Complete Phase 5: Documentation
8. Complete Phase 6: Polish
9. **POV COMPLETE**: Full provider authentication + data source validated

### TDD Workflow Per User Story

**For User Story 1 (Authentication)**:
1. RED: Write all failing tests (T008-T014) - verify they FAIL
2. GREEN: Minimal implementation (T015-T021) - verify tests PASS
3. REFACTOR: Improve quality (T022-T028) - verify tests still PASS

**For User Story 2 (Data Source)**:
1. RED: Write all failing tests (T029-T036) - verify they FAIL
2. GREEN: Minimal implementation (T037-T046) - verify tests PASS
3. REFACTOR: Improve quality (T047-T054) - verify tests still PASS

### Incremental Delivery

1. Complete Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Authentication validated
3. Add User Story 2 → Test independently → Data source validated
4. Add Documentation → POV ready for demo
5. Add Polish → POV ready for production consideration

### Parallel Team Strategy

With multiple developers after Foundational phase completes:

**Option 1: Sequential TDD (Safer)**:
- Developer A: Complete User Story 1 (RED → GREEN → REFACTOR)
- Developer B: Starts User Story 2 tests (RED phase) while A is in GREEN/REFACTOR
- Developer B: Completes User Story 2 (GREEN → REFACTOR) after A finishes

**Option 2: Test-First Parallel (Advanced)**:
- Developer A: User Story 1 tests (RED) + Implementation (GREEN + REFACTOR)
- Developer B: User Story 2 tests (RED) in parallel, waits for A to complete for GREEN phase
- Requires coordination but maximizes parallel work

---

## Acceptance Test Execution

### Environment Setup

```bash
# Set environment variables for acceptance tests
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

### Running Tests

```bash
# Run all acceptance tests (after User Story 2 complete)
TF_ACC=1 go test -v -timeout 120m ./internal/provider

# Run specific test
TF_ACC=1 go test -v -run TestAccCMPartSoftwareImagesDataSource_Basic -timeout 120m ./internal/provider

# Run unit tests only (no TF_ACC)
go test -v ./internal/provider

# Run with coverage
go test -v -cover ./internal/provider
```

---

## Notes

- [P] tasks = different files, no dependencies - can run in parallel
- [Story] label maps task to specific user story (US1 = Authentication, US2 = Data Source)
- Each user story follows strict TDD: RED (failing tests) → GREEN (minimal implementation) → REFACTOR (production quality)
- Verify tests FAIL before implementing (RED phase critical for TDD)
- Verify tests PASS after minimal implementation (GREEN phase validates approach)
- Verify tests still PASS after refactoring (REFACTOR phase ensures no regression)
- Commit after each TDD phase (RED commit, GREEN commit, REFACTOR commit)
- Stop at any checkpoint to validate story independently
- Both user stories required for POV completion (authentication + data source)
- All file paths are absolute from repository root `/workspace/`
- POV focuses on cookie-based authentication with login API call (not HTTP Basic Auth)
- Self-signed certificate handling via `insecure_skip_verify` provider attribute
- Defensive error parsing for unknown API error formats
- Complete schema with 30+ SoftwareImage fields + nested modules array (8 fields)

---

## Task Count Summary

- **Total Tasks**: 71
- **Phase 1 (Setup)**: 3 tasks
- **Phase 2 (Foundational)**: 5 tasks (includes T007a for arg field documentation)
- **Phase 3 (User Story 1 - Authentication)**: 23 tasks (7 RED, 8 GREEN with T021a, 8 REFACTOR with T028a)
- **Phase 4 (User Story 2 - Data Source)**: 26 tasks (8 RED, 10 GREEN, 8 REFACTOR)
- **Phase 5 (Documentation)**: 6 tasks
- **Phase 6 (Polish)**: 8 tasks

**Parallel Opportunities**: 44 tasks marked [P] can run in parallel when their phase is active (includes T007a and T021a)

**Independent Test Criteria**:
- User Story 1: Provider configures successfully with credentials and performs login API call → authentication token stored in cookie jar
- User Story 2: Data source queries getSoftwareImages → complete schema with all fields and nested modules returned

**Suggested MVP Scope**: Complete both User Stories (POV requires authentication + data source to validate JSON-RPC pattern)
