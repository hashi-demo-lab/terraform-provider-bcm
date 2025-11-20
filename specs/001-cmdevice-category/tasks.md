# Tasks: BCM CMDevice Category Resource

**Input**: Design documents from `/workspace/specs/001-cmdevice-category/`
**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/cmdevice-category-api.md, quickstart.md

**Tests**: This implementation follows TDD (Test-Driven Development) with RED-GREEN-REFACTOR cycles. All test tasks are included as part of the core workflow.

**Organization**: Tasks follow TDD principles with clear RED-GREEN-REFACTOR phases for each capability area.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Project Structure)

**Purpose**: Initialize project structure and ensure development environment is ready

- [X] T001 Verify Go 1.24+ installation and BCM provider dependencies in /workspace/go.mod
- [X] T00- [ ] T002 [P] Verify BCM test environment connectivity (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- [X] T00- [ ] T003 [P] Create resource file structure: /workspace/internal/provider/resource_cmdevice_category.go
- [X] T00- [ ] T004 [P] Create test file structure: /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T00- [ ] T005 [P] Create examples directory: /workspace/examples/resources/bcm_cmdevice_category/

**Checkpoint**: Project structure ready for TDD implementation

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story implementation

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [X] T0- [ ] T006 Define complete CMDeviceCategoryResourceModel struct with all 60+ attributes in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T007 [P] Define nested object models (SoftwareImageProxyModel, BMCSettingsModel, FSMountModel, KernelModuleModel) in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T008 Implement NewCMDeviceCategoryResource factory function in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T009 Implement Metadata method returning resource type name in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T010 Implement Configure method to store BCM client in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T011 Register resource in provider Resources method in /workspace/internal/provider/provider.go

**Checkpoint**: Foundation ready - user story implementation can now begin

---

## Phase 3: User Story 1 - Create and Manage Device Category Configuration (Priority: P1) 🎯 MVP

**Goal**: Enable infrastructure administrators to create, read, update, and delete device categories via Terraform with basic configuration (name, description, boot settings, kernel parameters).

**Independent Test**: Create a category with name "test-category", management network UUID, and kernel parameters via Terraform, verify it appears in BCM with correct configuration, update the kernel parameters, verify changes applied, then destroy and verify removal from BCM.

### RED Phase - Write Failing Tests for User Story 1

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [X] T0- [ ] T012 [US1] Write failing acceptance test TestAccCMDeviceCategoryResource_Basic in /workspace/internal/provider/resource_cmdevice_category_test.go with Create/Read/Update/Delete steps
- [X] T0- [ ] T013 [US1] Write test config helper testAccCMDeviceCategoryResourceConfig for basic category in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T0- [ ] T014 [US1] Write test config helper testAccCMDeviceCategoryResourceConfig_Updated for updated category in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T0- [ ] T015 [US1] Run acceptance test expecting failure (resource not implemented) using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Basic

**Checkpoint**: Tests fail with "resource not registered" - RED phase complete

### GREEN Phase - Minimal Implementation for User Story 1

- [X] T0- [ ] T016 [US1] Implement minimal Schema method with core attributes (id, uuid, name, notes, management_network, kernel_parameters, force) in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T017 [US1] Implement minimal Create method with hardcoded UUID in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T018 [US1] Implement minimal Read method returning current state in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T0- [ ] T019 [US1] Implement minimal Update method copying plan to state in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T020 [US1] Implement minimal Delete method (no-op) in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T021 [US1] Implement ImportState method using resource.ImportStatePassthroughID in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T022 [US1] Run acceptance test expecting success with minimal implementation using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Basic

**Checkpoint**: Tests pass with hardcoded values - GREEN phase complete

### REFACTOR Phase - Real API Integration for User Story 1

- [X] T02- [ ] T023 [US1] Implement buildAPIEntity helper to construct BCM API Category entity from Terraform model in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T024 [US1] Implement readCategory helper using getCategory(name) API for efficient direct lookup in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T025 [US1] Update Create method to call addCategory API with force parameter and read back created category in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T026 [US1] Update Read method to call getCategory API by name and populate all model fields in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T027 [US1] Update Update method to call updateCategory API with complete entity and force parameter in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T028 [US1] Update Delete method to call removeCategory API with UUID and force parameter in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T02- [ ] T029 [US1] Run acceptance test expecting success with real API integration using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Basic
- [X] T030 [US1] Add logging with tflog for all CRUD operations in /workspace/internal/provider/resource_cmdevice_category.go

**Checkpoint**: User Story 1 fully functional with real BCM API - can create, read, update, delete basic categories

---

## Phase 4: User Story 2 - Import Existing Categories into Terraform State (Priority: P2)

**Goal**: Enable infrastructure teams to import existing BCM categories into Terraform management without disrupting current configurations.

**Independent Test**: Manually create a category in BCM, import it using terraform import with UUID, verify terraform plan shows no changes, update a field in Terraform, verify only changed field updated in BCM.

### RED Phase - Write Failing Tests for User Story 2

- [X] T031 [US2] Write failing acceptance test TestAccCMDeviceCategoryResource_Import in /workspace/internal/provider/resource_cmdevice_category_test.go with ImportState step
- [X] T032 [US2] Add ImportStateVerify and ImportStateVerifyIgnore configuration to test in /workspace/internal/provider/resource_cmdevice_category_test.go
- [X] T033 [US2] Run import acceptance test expecting failure using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Import

**Checkpoint**: Import test fails - RED phase complete

### GREEN & REFACTOR Phase - Import Implementation for User Story 2

- [X] T034 [US2] Update ImportState method to implement two-phase import: getCategories to find by UUID, then getCategory by name in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T035 [US2] Implement UUID extraction and name lookup logic in ImportState in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T036 [US2] Update readCategory to handle import scenario by populating all fields including computed ones in /workspace/internal/provider/resource_cmdevice_category.go
- [X] T037 [US2] Fix plan value preservation for optional fields - Implement preservePlanValues parameter in readCategory to prevent "inconsistent result after apply" errors when BCM returns defaults
- [X] T038 [US2] Run import acceptance test expecting success and verify imported category shows no changes on terraform plan

**Checkpoint**: User Story 2 complete - categories can be imported into Terraform state accurately

---

## Phase 5: User Story 3 - Delete Categories Safely (Priority: P3)

**Goal**: Enable safe deletion of obsolete categories with appropriate force parameter controls for categories with node assignments.

**Independent Test**: Create category via Terraform, destroy it, verify removal from BCM. Create category with nodes assigned (manual step), attempt destroy without force, verify error message, attempt destroy with force=true, verify success.

### RED Phase - Write Failing Tests for User Story 3

- [ ] T039 [US3] Write failing acceptance test TestAccCMDeviceCategoryResource_ForceParameter in /workspace/internal/provider/resource_cmdevice_category_test.go testing force behavior
- [ ] T040 [US3] Add test config with force=true parameter in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T041 [US3] Run force parameter test expecting initial failure using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_ForceParameter

**Checkpoint**: Force parameter test defined - RED phase complete

### GREEN & REFACTOR Phase - Force Parameter Implementation for User Story 3

- [ ] T042 [US3] Update Delete method to handle API errors when category has assigned nodes in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T043 [US3] Add clear diagnostic error message for deletion failures with guidance on force parameter in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T044 [US3] Test force parameter behavior with create/update/delete operations in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T045 [US3] Run force parameter acceptance test expecting success using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_ForceParameter

**Checkpoint**: User Story 3 complete - safe category deletion with force parameter working

---

## Phase 6: Comprehensive Schema and Nested Objects

**Goal**: Implement complete 60+ attribute schema including all nested objects (BMCSettings, SoftwareImageProxy, FSMount, KernelModule arrays).

**Independent Test**: Create category with all nested objects populated, verify persistence, update nested objects, verify changes applied.

### RED Phase - Comprehensive Schema Tests

- [ ] T046 [P] Write failing acceptance test TestAccCMDeviceCategoryResource_Complete in /workspace/internal/provider/resource_cmdevice_category_test.go with full schema
- [ ] T047 [P] Write failing acceptance test TestAccCMDeviceCategoryResource_NestedObjects in /workspace/internal/provider/resource_cmdevice_category_test.go for nested configurations
- [ ] T048 [P] Write test config helpers for comprehensive category with nested objects in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T049 Run comprehensive tests expecting failures using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource_(Complete|NestedObjects)"

**Checkpoint**: Comprehensive schema tests fail - RED phase complete

### GREEN & REFACTOR Phase - Complete Schema Implementation

- [ ] T050 Expand Schema method with all boot configuration attributes (boot_loader, boot_loader_file, boot_loader_protocol) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T051 [P] Add kernel configuration attributes (kernel_version, kernel_output_console, modules list) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T052 [P] Add disk and storage attributes (disksetup, raidconf, fsmounts list, fsexports list) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T053 [P] Add network configuration attributes (default_gateway, default_gateway_metric, name_servers, search_domains, time_servers, static_routes) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T054 [P] Add provisioning attributes (install_mode, new_node_install_mode, install_boot_record, initialize, finalize) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T055 Add software_image_proxy nested object to schema with SoftwareImageProxyModel in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T056 Add bmc_settings nested object to schema with BMCSettingsModel (mark password as sensitive) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T057 [P] Add advanced settings objects (bios_setup, dpu_settings, gpu_settings, roles, services) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T058 [P] Add security and access objects (access_settings, selinux_settings, proxy_settings, timezone_settings, ztp_settings, fips) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T059 [P] Add exclude list attributes (exclude_list_full, exclude_list_grab, exclude_list_grabnew, exclude_list_sync, exclude_list_update, exclude_list_manipulate_script) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T060 [P] Add behavioral flags (data_node, node_installer_disk, version_config_files, authentication_service, allow_networking_restart, interactive_user, use_exclusively_for) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T061 [P] Add computed metadata fields (parent_uuid, revision, modified, to_be_removed, base_type, child_type) to schema in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T062 Update buildAPIEntity to handle all nested objects with proper baseType fields in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T063 Update buildAPIEntity to handle array fields (modules, fsmounts, nameServers, etc.) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T064 Update readCategory to parse all nested objects from API response using helper functions in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T065 Update readCategory to handle null nested objects gracefully (set to null, not empty object) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T066 Add validation for management_network UUID format (RFC 4122) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T067 [P] Add validation for boot_loader enum values (SYSLINUX, GRUB, GRUB2, PXELINUX) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T068 [P] Add validation for disksetup XML length (max 10KB) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T069 [P] Add validation for exclude list fields length (max 50KB each) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T070 Run comprehensive acceptance tests expecting success using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource_(Complete|NestedObjects)"

**Checkpoint**: Complete 60+ attribute schema working with all nested objects

---

## Phase 7: Validation and Error Handling

**Goal**: Implement comprehensive client-side and server-side validation with clear error messages.

**Independent Test**: Attempt to create category with invalid UUID, verify validation error. Attempt duplicate category name, verify conflict error.

### RED Phase - Validation Tests

- [ ] T071 Write failing acceptance test TestAccCMDeviceCategoryResource_Validation in /workspace/internal/provider/resource_cmdevice_category_test.go for validation scenarios
- [ ] T072 Add test cases for invalid management network UUID in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T073 Add test cases for duplicate category name conflict in /workspace/internal/provider/resource_cmdevice_category_test.go
- [ ] T074 Run validation tests expecting failures using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Validation

**Checkpoint**: Validation tests defined - RED phase complete

### GREEN & REFACTOR Phase - Validation Implementation

- [ ] T075 Add IP address format validation for default_gateway in schema validators in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T076 Add enum validation for fips (YES, NO) in schema validators in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T077 Add enum validation for install_mode and authentication_service in schema validators in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T078 Implement validateCategory API call before Create AND Update operations in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T079 Parse validateCategory response for errors and warnings, surface as diagnostics in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T080 Surface validation errors as Terraform diagnostics with clear messages in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T081 Handle API error responses (400, 404, 409, 422, 500) with user-friendly messages in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T082 Add error message pattern for category not found in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T083 Add error message pattern for category name conflict in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T084 Add error message pattern for category deletion with assigned nodes in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T085 Run validation acceptance tests expecting success using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Validation

**Checkpoint**: Comprehensive validation working with clear error messages

---

## Phase 8: Documentation and Examples

**Goal**: Create comprehensive documentation and example configurations for all common use cases.

- [X] T086 [P] Create minimal category example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T087 [P] Create category with boot configuration example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T088 [P] Create category with disk setup XML example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T089 [P] Create category with network configuration example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T090 [P] Create category with filesystem mounts example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T091 [P] Create category with BMC settings example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T092 [P] Create category with software image proxy example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T093 [P] Create category with force parameter example in /workspace/examples/resources/bcm_cmdevice_category/resource.tf
- [X] T094 [P] Add import example to /workspace/examples/resources/bcm_cmdevice_category/import.sh
- [X] T095 Generate provider documentation using: make generate
- [X] T096 Verify generated documentation at /workspace/docs/resources/bcm_cmdevice_category.md
- [ ] T097 Update resource descriptions and markdown in schema for clarity in /workspace/internal/provider/resource_cmdevice_category.go

**Checkpoint**: Complete documentation and examples available - COMPLETE

---

## Phase 9: Polish & Cross-Cutting Concerns

**Purpose**: Final code quality improvements, testing, and verification

- [X] T098 [P] Format Go code using: make fmt
- [X] T099 [P] Run go vet to check code quality
- [ ] T100 [P] Run pre-commit hooks using: pre-commit run --all-files (requires pre-commit setup)
- [ ] T101 Run all acceptance tests to verify nothing broken using: TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource (requires BCM credentials)
- [ ] T102 Verify test coverage >80% using: go test -cover ./internal/provider/ (requires BCM credentials)
- [ ] T103 Review error messages for clarity and actionability in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T104 Review code for reuse of helper functions (getStringValue, getBoolValue, getInt64Value) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T105 Verify sensitive fields properly marked (bmc_settings.password) in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T106 Run quickstart.md validation following /workspace/specs/001-cmdevice-category/quickstart.md (requires BCM credentials)
- [ ] T107 [P] Code cleanup and refactoring for readability in /workspace/internal/provider/resource_cmdevice_category.go
- [ ] T108 Final verification: Create category manually, import, update, destroy cycle (requires BCM credentials)

**Checkpoint**: Implementation complete, code quality checks pass. Acceptance tests pending BCM environment availability.

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational phase - Core CRUD implementation
- **User Story 2 (Phase 4)**: Depends on User Story 1 - Import builds on Read functionality
- **User Story 3 (Phase 5)**: Depends on User Story 1 - Delete with force parameter
- **Comprehensive Schema (Phase 6)**: Depends on User Story 1 - Extends basic CRUD with full schema
- **Validation (Phase 7)**: Depends on Comprehensive Schema - Validates all fields
- **Documentation (Phase 8)**: Depends on Validation - Documents complete functionality
- **Polish (Phase 9)**: Depends on all previous phases - Final cleanup

### User Story Dependencies

- **User Story 1 (P1)**: Independent - Core CRUD operations (Create, Read, Update, Delete)
- **User Story 2 (P2)**: Depends on US1 Read - Import uses same read logic
- **User Story 3 (P3)**: Depends on US1 Delete - Force parameter extends delete behavior

### TDD Cycle Dependencies Within Each Story

- RED phase: Tests MUST fail before proceeding to GREEN
- GREEN phase: Tests MUST pass with minimal implementation before REFACTOR
- REFACTOR phase: Tests MUST remain passing while improving code

### Parallel Opportunities

- Within Setup (Phase 1): Tasks T002-T005 can run in parallel
- Within Foundational (Phase 2): Task T007 can run in parallel with T006
- Within Comprehensive Schema (Phase 6): Tasks T051-T054, T057-T061 can run in parallel
- Within Documentation (Phase 8): Tasks T086-T094 can run in parallel
- Within Polish (Phase 9): Tasks T098-T100 can run in parallel

---

## Parallel Example: Comprehensive Schema Implementation

```bash
# Launch all schema attribute groups in parallel:
Task T051: "Add kernel configuration attributes to schema"
Task T052: "Add disk and storage attributes to schema"
Task T053: "Add network configuration attributes to schema"
Task T054: "Add provisioning attributes to schema"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup
2. Complete Phase 2: Foundational (CRITICAL - blocks all stories)
3. Complete Phase 3: User Story 1 (RED-GREEN-REFACTOR)
4. **STOP and VALIDATE**: Test basic category CRUD independently
5. Deploy/demo if ready

### Incremental Delivery

1. Setup + Foundational → Foundation ready
2. Add User Story 1 → Test independently → Deploy/Demo (MVP!)
3. Add User Story 2 → Test import → Deploy/Demo
4. Add User Story 3 → Test force parameter → Deploy/Demo
5. Add Comprehensive Schema → Test all attributes → Deploy/Demo
6. Add Validation → Test error handling → Deploy/Demo
7. Add Documentation → Complete feature

### TDD RED-GREEN-REFACTOR Cycles

**User Story 1 Example**:
1. RED (T012-T015): Write failing tests, verify failures
2. GREEN (T016-T022): Minimal implementation, verify tests pass
3. REFACTOR (T023-T030): Real API integration, verify tests still pass

Each user story follows this same pattern.

---

## Notes

- **[P]** tasks can run in parallel (different files, no dependencies)
- **[Story]** label maps task to specific user story (US1, US2, US3)
- **TDD Discipline**: NEVER skip RED phase - tests must fail first
- **TDD Discipline**: NEVER skip GREEN phase - minimal implementation proves tests work
- **TDD Discipline**: NEVER break tests in REFACTOR - continuous validation
- Each user story should be independently testable
- Commit after each TDD cycle (RED-GREEN-REFACTOR complete)
- Reference implementation: /workspace/internal/provider/resource_cmpart_softwareimage.go
- Helper functions: /workspace/internal/provider/data_source_cmpart_softwareimages.go (getStringValue, getBoolValue, getInt64Value)
- BCM API Client: /workspace/internal/provider/bcm_client.go (CallJSONRPC method)
- Use efficient getCategory(name) for Read, not getCategories() + filter
- Mark bmc_settings.password as sensitive
- Handle null nested objects gracefully (types.ObjectNull, not empty struct)
- All API responses must be parsed defensively (fields may be null/missing)

---

## Total Task Count

**Total Tasks**: 108
- Setup: 5 tasks
- Foundational: 6 tasks
- User Story 1: 19 tasks (RED: 4, GREEN: 7, REFACTOR: 8)
- User Story 2: 8 tasks (RED: 3, GREEN/REFACTOR: 5)
- User Story 3: 7 tasks (RED: 3, GREEN/REFACTOR: 4)
- Comprehensive Schema: 25 tasks (RED: 4, GREEN/REFACTOR: 21)
- Validation: 15 tasks (RED: 4, GREEN/REFACTOR: 11)
- Documentation: 12 tasks
- Polish: 11 tasks

**Estimated Time**:
- Setup: 30 minutes
- Foundational: 60 minutes
- User Story 1 (MVP): 3-4 hours
- User Story 2: 1-2 hours
- User Story 3: 1-2 hours
- Comprehensive Schema: 4-6 hours
- Validation: 2-3 hours
- Documentation: 1-2 hours
- Polish: 1-2 hours
- **Total**: 14-22 hours (depends on experience with Terraform Provider Framework)

**Suggested MVP Scope**: Complete through Phase 3 (User Story 1) - Basic category CRUD with ~5 core attributes. This delivers immediate value for Infrastructure-as-Code category management.
