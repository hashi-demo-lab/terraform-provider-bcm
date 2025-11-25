# Tasks: Category Dynamic Fields Schema Implementation

**Input**: Design documents from `/workspace/specs/001-category-dynamic-fields/`
**Prerequisites**: plan.md (required), spec.md (required)

**Tests**: All test tasks are REQUIRED for this TDD implementation. Each field requires 7 test scenarios.

**Organization**: Tasks are grouped by priority and field to enable incremental implementation following RED-GREEN-REFACTOR pattern.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1-US5)
- Include exact file paths in descriptions

## Path Conventions

- Resource implementation: `/workspace/internal/provider/resource_cmdevice_category.go`
- Resource tests: `/workspace/internal/provider/resource_cmdevice_category_test.go`
- Examples: `/workspace/examples/resources/bcm_cmdevice_category/*.tf`
- Documentation: Auto-generated in `/workspace/docs/` via `make generate`

---

## Phase 0: Research & API Discovery

**Purpose**: Document BCM API structure for all 5 dynamic fields

**⚠️ CRITICAL**: This phase provides the API contracts needed for all subsequent implementation phases

### BCM API Exploration Tasks

- [ ] T001 [P] Research static_routes BCM API structure: Query BCM API for categories with static routes configured, document JSON response format including field names (destination, gateway, metric), data types, and validate CIDR/IPv4 patterns. Output to sampleRest/category-dynamic-fields/static-routes.json
- [ ] T002 [P] Research fsexports BCM API structure: Query BCM API for categories with NFS exports, document JSON response format including field names (path, network, allowWrite, rootSquash, async, baseType), validate UUID format for network field. Output to sampleRest/category-dynamic-fields/fsexports.json
- [ ] T003 [P] Research roles BCM API structure: Query BCM API for categories with roles assigned, document JSON response format including field names (name, childType, uuid, addServices, baseType), list known role types (HeadNodeRole, StorageRole, BackupRole, etc.), confirm uuid is computed by BCM. Output to sampleRest/category-dynamic-fields/roles.json
- [ ] T004 [P] Research gpu_settings BCM API structure: Query BCM API for categories with GPU configurations, document JSON response format including field names (deviceId, model, computeMode, baseType), list known compute modes (default, exclusive, prohibited). Output to sampleRest/category-dynamic-fields/gpu-settings.json
- [ ] T005 [P] Research services BCM API structure: Query BCM API for categories with services configured, document JSON response format and service object structure. If structure is unclear, document findings and mark services field as POST-MVP candidate. Output to sampleRest/category-dynamic-fields/services.json

**Checkpoint**: All API structures documented - schema definition can proceed

---

## Phase 1: Foundational (Blocking Prerequisites)

**Purpose**: Define schemas and model structs that ALL field implementations depend on

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [ ] T006 Define StaticRouteModel struct in /workspace/internal/provider/resource_cmdevice_category.go with destination (types.String, required), gateway (types.String, required), and metric (types.Int64, optional) fields
- [ ] T007 Define FSExportModel struct in /workspace/internal/provider/resource_cmdevice_category.go with path (types.String, required), network (types.String, required), allow_write (types.Bool, optional), root_squash (types.Bool, optional), and async (types.Bool, optional) fields
- [ ] T008 Define RoleModel struct in /workspace/internal/provider/resource_cmdevice_category.go with name (types.String, required), child_type (types.String, required), uuid (types.String, computed), and add_services (types.Bool, optional) fields
- [ ] T009 Define GPUSettingModel struct in /workspace/internal/provider/resource_cmdevice_category.go with device_id (types.String, required), model (types.String, optional), and compute_mode (types.String, optional) fields
- [ ] T010 Define ServiceModel struct in /workspace/internal/provider/resource_cmdevice_category.go based on Phase 0 research findings (or mark as POST-MVP if structure unclear)

**Checkpoint**: All model structs defined - field implementations can now proceed in parallel

---

## Phase 2: User Story 1 - Static Route Configuration (Priority: P1) 🎯 MVP

**Goal**: Enable static route configuration at category level for multi-network cluster architectures

**Independent Test**: Create category with 2 static routes, verify routes persist through CRUD operations, update routes, import resource, detect external drift

### Tests for User Story 1 (RED Phase) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T011 [P] [US1] Test 1: Basic CRUD test for static_routes - Create category with 2 routes (storage network 192.168.1.0/24 via 10.0.0.1, backup network 192.168.2.0/24 via 10.0.0.2), verify routes persist on read, update by adding 3rd route, verify changes applied. Test in /workspace/internal/provider/resource_cmdevice_category_test.go as TestAccCMDeviceCategory_StaticRoutesBasicCRUD
- [ ] T012 [P] [US1] Test 2: Idempotency test after Create - Create category with static routes, run terraform apply again with same config, verify plan is empty (no changes detected). Test as TestAccCMDeviceCategory_StaticRoutesIdempotencyCreate
- [ ] T013 [P] [US1] Test 3: Idempotency test after Update - Update category routes, run terraform apply again with same updated config, verify plan is empty. Test as TestAccCMDeviceCategory_StaticRoutesIdempotencyUpdate
- [ ] T014 [P] [US1] Test 4: Import test - Create category with static routes, import using UUID, verify all route fields (destination, gateway, metric) are accurately imported into state. Test as TestAccCMDeviceCategory_StaticRoutesImport
- [ ] T015 [P] [US1] Test 5: Drift detection test - Create category with routes, modify routes externally via BCM API (change gateway or add/remove route), verify Terraform detects drift (non-empty plan) and can restore desired state. Test as TestAccCMDeviceCategory_StaticRoutesDrift
- [ ] T016 [P] [US1] Test 6: Empty list test - Create category with empty static_routes array [], verify empty list is preserved in state (not converted to null), test read/update operations with empty list. Test as TestAccCMDeviceCategory_StaticRoutesEmpty
- [ ] T017 [P] [US1] Test 7: Validation test - Test invalid CIDR notation (e.g., "192.168.1/24"), invalid gateway IP (e.g., "999.999.999.999"), verify clear validation error messages before API submission. Test as TestAccCMDeviceCategory_StaticRoutesValidation

**Verify Tests Fail**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_StaticRoutes` and confirm all 7 tests fail

### Implementation for User Story 1 (GREEN Phase)

- [ ] T018 [US1] Replace types.Dynamic with ListNestedAttribute schema for static_routes field at line 80 in /workspace/internal/provider/resource_cmdevice_category.go: Define NestedObject with destination (StringAttribute with CIDR validator regex `^([0-9]{1,3}\.){3}[0-9]{1,3}/[0-9]{1,2}$`), gateway (StringAttribute with IPv4 validator regex `^([0-9]{1,3}\.){3}[0-9]{1,3}$`), and metric (Int64Attribute optional). Mark as Optional at top level
- [ ] T019 [US1] Implement static_routes serialization in buildAPIEntity method (line ~1207): Extract routes list from model.StaticRoutes, convert each route to map[string]interface{} with snake_case to camelCase mapping (destination→destination, gateway→gateway, metric→metric), handle empty list with empty array (not null), add to entity["staticRoutes"]
- [ ] T020 [US1] Implement static_routes deserialization in readCategory method (line ~1414): Extract staticRoutes from categoryData, convert BCM API array to Terraform list using types.ListValueMust, map each route object to StaticRouteModel with null-safe helpers, handle null/empty array cases, set model.StaticRoutes
- [ ] T021 [US1] Handle empty list preservation: In readCategory, use types.ListValueMust(elementType, []attr.Value{}) for empty staticRoutes array from API instead of types.ListNull, ensure buildAPIEntity sends [] for empty lists not null

**Verify Tests Pass**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_StaticRoutes` and confirm all 7 tests pass

### Refactor for User Story 1 (REFACTOR Phase)

- [ ] T022 [US1] Add detailed error messages: Enhance validation errors for static routes to include helpful messages like "destination must be valid CIDR notation (e.g., 192.168.1.0/24)" and "gateway must be valid IPv4 address (e.g., 10.0.0.1)"
- [ ] T023 [US1] Add debug logging: Add tflog.Debug statements in buildAPIEntity and readCategory for static_routes serialization/deserialization with route count and sample values for troubleshooting
- [ ] T024 [US1] Optimize allocations: Review static_routes serialization code for unnecessary allocations, consider pre-allocating slices when route count is known

### Documentation for User Story 1

- [ ] T025 [US1] Create static routes example: Create /workspace/examples/resources/bcm_cmdevice_category/static-routes.tf with example showing 2-3 static routes with different metrics demonstrating storage and backup network routing scenarios
- [ ] T026 [US1] Generate provider documentation: Run `make generate` to regenerate /workspace/docs/resources/bcm_cmdevice_category.md with static_routes field documentation, verify markdown formatting and examples are correct

**Checkpoint**: At this point, User Story 1 (static_routes) should be fully functional and testable independently

---

## Phase 3: User Story 2 - NFS Export Management (Priority: P1)

**Goal**: Enable NFS export configuration at category level for cluster-wide storage access

**Independent Test**: Create category with NFS export for /home with read-write access, verify export persists through CRUD operations, test different permission combinations

### Tests for User Story 2 (RED Phase) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T027 [P] [US2] Test 1: Basic CRUD test for fsexports - Create category with NFS export for /home (read-write, root squash enabled), verify export persists on read, update permissions from read-only to read-write, verify changes applied. Test in /workspace/internal/provider/resource_cmdevice_category_test.go as TestAccCMDeviceCategory_FSExportsBasicCRUD
- [ ] T028 [P] [US2] Test 2: Idempotency test after Create - Create category with fsexports, run terraform apply again with same config, verify plan is empty. Test as TestAccCMDeviceCategory_FSExportsIdempotencyCreate
- [ ] T029 [P] [US2] Test 3: Idempotency test after Update - Update category fsexports, run terraform apply again with same updated config, verify plan is empty. Test as TestAccCMDeviceCategory_FSExportsIdempotencyUpdate
- [ ] T030 [P] [US2] Test 4: Import test - Create category with fsexports, import using UUID, verify all export fields (path, network UUID, allow_write, root_squash, async) are accurately imported. Test as TestAccCMDeviceCategory_FSExportsImport
- [ ] T031 [P] [US2] Test 5: Drift detection test - Create category with exports, modify export permissions externally via BCM API (toggle allow_write or async), verify Terraform detects drift and can remediate. Test as TestAccCMDeviceCategory_FSExportsDrift
- [ ] T032 [P] [US2] Test 6: Empty list test - Create category with empty fsexports array [], verify empty list preserved in state, test operations with empty list. Test as TestAccCMDeviceCategory_FSExportsEmpty
- [ ] T033 [P] [US2] Test 7: Validation test - Test invalid network UUID format (e.g., "not-a-uuid"), verify UUID validator catches error with clear message. Test as TestAccCMDeviceCategory_FSExportsValidation

**Verify Tests Fail**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_FSExports` and confirm all 7 tests fail

### Implementation for User Story 2 (GREEN Phase)

- [ ] T034 [US2] Replace types.Dynamic with ListNestedAttribute schema for fsexports field at line 85 in /workspace/internal/provider/resource_cmdevice_category.go: Define NestedObject with path (StringAttribute required), network (StringAttribute required with UUID validator regex `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`), allow_write (BoolAttribute optional), root_squash (BoolAttribute optional), async (BoolAttribute optional). Mark as Optional at top level
- [ ] T035 [US2] Implement fsexports serialization in buildAPIEntity method: Extract exports from model.FSExports, convert to BCM API format with baseType="FSExport", map snake_case to camelCase (allow_write→allowWrite, root_squash→rootSquash), handle empty list with empty array, add to entity["fsexports"]
- [ ] T036 [US2] Implement fsexports deserialization in readCategory method (line ~1434): Extract fsexports from categoryData, convert BCM API array to Terraform list, map each export object to FSExportModel with camelCase to snake_case conversion, handle null/empty array, set model.FSExports
- [ ] T037 [US2] Handle baseType exclusion: Ensure baseType field from BCM API is NOT included in Terraform schema (internal BCM implementation detail), but include baseType="FSExport" when serializing to BCM API in buildAPIEntity

**Verify Tests Pass**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_FSExports` and confirm all 7 tests pass

### Refactor for User Story 2 (REFACTOR Phase)

- [ ] T038 [US2] Add detailed error messages: Enhance validation errors for fsexports to include helpful messages for UUID format and export path requirements
- [ ] T039 [US2] Add debug logging: Add tflog.Debug statements for fsexports serialization/deserialization with export count and paths
- [ ] T040 [US2] Optimize allocations: Review fsexports serialization code for unnecessary allocations

### Documentation for User Story 2

- [ ] T041 [US2] Create fsexports example: Create /workspace/examples/resources/bcm_cmdevice_category/fsexports.tf with example showing /home and /shared exports with different permission combinations (read-only, read-write, root squash variations)
- [ ] T042 [US2] Generate provider documentation: Run `make generate` to update docs with fsexports field documentation

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently

---

## Phase 4: User Story 3 - Role Assignment (Priority: P2)

**Goal**: Enable role assignment at category level for role-based cluster organization

**Independent Test**: Create category with HeadNodeRole, verify role persists with computed UUID, test multiple role assignments

### Tests for User Story 3 (RED Phase) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T043 [P] [US3] Test 1: Basic CRUD test for roles - Create category with HeadNodeRole (add_services=true), verify role persisted with computed UUID, add StorageRole and BackupRole, verify all roles have unique UUIDs. Test in /workspace/internal/provider/resource_cmdevice_category_test.go as TestAccCMDeviceCategory_RolesBasicCRUD
- [ ] T044 [P] [US3] Test 2: Idempotency test after Create - Create category with roles, run terraform apply again, verify plan is empty. Test as TestAccCMDeviceCategory_RolesIdempotencyCreate
- [ ] T045 [P] [US3] Test 3: Idempotency test after Update - Update category roles (add/remove), run terraform apply again with updated config, verify plan is empty. Test as TestAccCMDeviceCategory_RolesIdempotencyUpdate
- [ ] T046 [P] [US3] Test 4: Import test - Create category with roles, import using UUID, verify role names, child_type values, UUIDs, and add_services flags are all preserved. Test as TestAccCMDeviceCategory_RolesImport
- [ ] T047 [P] [US3] Test 5: Drift detection test - Create category with roles, modify roles externally via BCM API (add/remove/reorder roles), verify Terraform detects drift. Test as TestAccCMDeviceCategory_RolesDrift
- [ ] T048 [P] [US3] Test 6: Empty list test - Create category with empty roles array [], verify empty list preserved in state. Test as TestAccCMDeviceCategory_RolesEmpty
- [ ] T049 [P] [US3] Test 7: Role types test - Create category with multiple role types (HeadNodeRole, StorageRole, ComputeRole, BackupRole, MonitoringRole), verify all are accepted and child_type validator allows flexibility for future role types. Test as TestAccCMDeviceCategory_RolesMultipleTypes

**Verify Tests Fail**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_Roles` and confirm all 7 tests fail

### Implementation for User Story 3 (GREEN Phase)

- [ ] T050 [US3] Replace types.Dynamic with ListNestedAttribute schema for roles field at line 88 in /workspace/internal/provider/resource_cmdevice_category.go: Define NestedObject with name (StringAttribute required), child_type (StringAttribute required), uuid (StringAttribute computed), add_services (BoolAttribute optional). Mark as Optional at top level
- [ ] T051 [US3] Implement roles serialization in buildAPIEntity method: Extract roles from model.Roles, convert to BCM API format with baseType="Role", childType from child_type field, map add_services→addServices, include uuid if present (update), generate if new (create), add to entity["roles"]
- [ ] T052 [US3] Implement roles deserialization in readCategory method (line ~1445): Extract roles from categoryData, convert BCM API array to Terraform list, map each role object to RoleModel with camelCase to snake_case (childType→child_type, addServices→add_services), extract computed uuid, set model.Roles
- [ ] T053 [US3] Handle role UUID computation: Ensure uuid field is marked Computed in schema, populated from BCM API response after create/update, never sent to BCM API for new roles (BCM assigns), but preserved for existing roles during updates

**Verify Tests Pass**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_Roles` and confirm all 7 tests pass

### Refactor for User Story 3 (REFACTOR Phase)

- [ ] T054 [US3] Add detailed error messages: Enhance validation to list known role types in error messages while still allowing flexibility for future types
- [ ] T055 [US3] Add debug logging: Add tflog.Debug statements for roles serialization/deserialization with role count and types
- [ ] T056 [US3] Optimize allocations: Review roles serialization code for unnecessary allocations

### Documentation for User Story 3

- [ ] T057 [US3] Create roles example: Create /workspace/examples/resources/bcm_cmdevice_category/roles.tf with example showing HeadNodeRole, StorageRole, and BackupRole assignments with add_services flags
- [ ] T058 [US3] Generate provider documentation: Run `make generate` to update docs with roles field documentation

**Checkpoint**: At this point, User Stories 1, 2, AND 3 should all work independently

---

## Phase 5: User Story 4 - GPU Configuration (Priority: P3)

**Goal**: Enable GPU settings at category level for AI/ML cluster GPU workloads

**Independent Test**: Create category with 4 Tesla V100 GPUs, verify device IDs and compute modes persist, test compute mode updates

### Tests for User Story 4 (RED Phase) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T059 [P] [US4] Test 1: Basic CRUD test for gpu_settings - Create category with 4 GPUs (device IDs 0-3, Tesla V100 model, default compute mode), verify settings persist, update compute mode for GPU 0 to exclusive, verify only that GPU changed. Test in /workspace/internal/provider/resource_cmdevice_category_test.go as TestAccCMDeviceCategory_GPUSettingsBasicCRUD
- [ ] T060 [P] [US4] Test 2: Idempotency test after Create - Create category with gpu_settings, run terraform apply again, verify plan is empty. Test as TestAccCMDeviceCategory_GPUSettingsIdempotencyCreate
- [ ] T061 [P] [US4] Test 3: Idempotency test after Update - Update gpu_settings, run terraform apply again with updated config, verify plan is empty. Test as TestAccCMDeviceCategory_GPUSettingsIdempotencyUpdate
- [ ] T062 [P] [US4] Test 4: Import test - Create category with gpu_settings, import using UUID, verify device IDs, models, and compute modes are all preserved. Test as TestAccCMDeviceCategory_GPUSettingsImport
- [ ] T063 [P] [US4] Test 5: Drift detection test - Create category with GPU settings, modify compute mode externally via BCM API, verify Terraform detects changes. Test as TestAccCMDeviceCategory_GPUSettingsDrift
- [ ] T064 [P] [US4] Test 6: Empty list test - Create category with empty gpu_settings array [], verify empty list preserved in state. Test as TestAccCMDeviceCategory_GPUSettingsEmpty
- [ ] T065 [P] [US4] Test 7: Multiple GPUs test - Create category with GPUs having different models (Tesla V100, A100, H100), different compute modes (default, exclusive, prohibited), verify all combinations work. Test as TestAccCMDeviceCategory_GPUSettingsMultiple

**Verify Tests Fail**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_GPUSettings` and confirm all 7 tests fail

### Implementation for User Story 4 (GREEN Phase)

- [ ] T066 [US4] Replace types.Dynamic with ListNestedAttribute schema for gpu_settings field at line 95 in /workspace/internal/provider/resource_cmdevice_category.go: Define NestedObject with device_id (StringAttribute required), model (StringAttribute optional), compute_mode (StringAttribute optional). Mark as Optional at top level
- [ ] T067 [US4] Implement gpu_settings serialization in buildAPIEntity method: Extract GPU settings from model.GPUSettings, convert to BCM API format with baseType="GPUSetting", map snake_case to camelCase (device_id→deviceId, compute_mode→computeMode), handle empty list with empty array, add to entity["gpuSettings"]
- [ ] T068 [US4] Implement gpu_settings deserialization in readCategory method (line ~1451): Extract gpuSettings from categoryData, convert BCM API array to Terraform list, map each GPU object to GPUSettingModel with camelCase to snake_case conversion, handle null/empty array, set model.GPUSettings
- [ ] T069 [US4] Handle baseType exclusion: Ensure baseType field from BCM API is NOT included in Terraform schema, but include baseType="GPUSetting" when serializing to BCM API

**Verify Tests Pass**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_GPUSettings` and confirm all 7 tests pass

### Refactor for User Story 4 (REFACTOR Phase)

- [ ] T070 [US4] Add detailed error messages: Enhance validation to list known compute modes (default, exclusive, prohibited) in error messages while allowing flexibility
- [ ] T071 [US4] Add debug logging: Add tflog.Debug statements for gpu_settings serialization/deserialization with GPU count and device IDs
- [ ] T072 [US4] Optimize allocations: Review gpu_settings serialization code for unnecessary allocations

### Documentation for User Story 4

- [ ] T073 [US4] Create gpu_settings example: Create /workspace/examples/resources/bcm_cmdevice_category/gpu-settings.tf with example showing 4 GPUs with different models and compute modes for AI/ML workloads
- [ ] T074 [US4] Generate provider documentation: Run `make generate` to update docs with gpu_settings field documentation

**Checkpoint**: At this point, User Stories 1-4 should all work independently

---

## Phase 6: User Story 5 - Service Configuration (Priority: P3)

**Goal**: Enable service configuration at category level for consistent system services

**Independent Test**: Create category with services (structure determined from Phase 0 research), verify services persist through CRUD operations

**⚠️ NOTE**: This phase may be marked POST-MVP if Phase 0 research determines the BCM API service structure is unclear

### Tests for User Story 5 (RED Phase) ⚠️

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation - OR skip if marked POST-MVP**

- [ ] T075 [P] [US5] Test 1: Basic CRUD test for services - Based on Phase 0 research findings, create category with service configurations, verify services persist, update service settings, verify changes applied. Test in /workspace/internal/provider/resource_cmdevice_category_test.go as TestAccCMDeviceCategory_ServicesBasicCRUD (OR skip if POST-MVP)
- [ ] T076 [P] [US5] Test 2: Idempotency test after Create - Create category with services, run terraform apply again, verify plan is empty. Test as TestAccCMDeviceCategory_ServicesIdempotencyCreate (OR skip if POST-MVP)
- [ ] T077 [P] [US5] Test 3: Idempotency test after Update - Update services, run terraform apply again with updated config, verify plan is empty. Test as TestAccCMDeviceCategory_ServicesIdempotencyUpdate (OR skip if POST-MVP)
- [ ] T078 [P] [US5] Test 4: Import test - Create category with services, import using UUID, verify all service configurations are preserved. Test as TestAccCMDeviceCategory_ServicesImport (OR skip if POST-MVP)
- [ ] T079 [P] [US5] Test 5: Drift detection test - Create category with services, modify services externally via BCM API, verify Terraform detects changes. Test as TestAccCMDeviceCategory_ServicesDrift (OR skip if POST-MVP)
- [ ] T080 [P] [US5] Test 6: Empty list test - Create category with empty services array [], verify empty list preserved in state. Test as TestAccCMDeviceCategory_ServicesEmpty (OR skip if POST-MVP)
- [ ] T081 [P] [US5] Test 7: Multiple services test - Based on research, test multiple service configurations. Test as TestAccCMDeviceCategory_ServicesMultiple (OR skip if POST-MVP)

**Verify Tests Fail**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_Services` and confirm tests fail (OR skip if POST-MVP)

### Implementation for User Story 5 (GREEN Phase)

- [ ] T082 [US5] Replace types.Dynamic with ListNestedAttribute schema for services field at line 89 in /workspace/internal/provider/resource_cmdevice_category.go based on Phase 0 research findings. Define appropriate NestedObject attributes. Mark as Optional. (OR keep as types.Dynamic and document as POST-MVP if structure unclear)
- [ ] T083 [US5] Implement services serialization in buildAPIEntity method based on research findings, include appropriate baseType if needed, map field names. (OR skip if POST-MVP)
- [ ] T084 [US5] Implement services deserialization in readCategory method based on research findings, convert BCM API format to Terraform list. (OR skip if POST-MVP)
- [ ] T085 [US5] Handle services field null/empty array cases appropriately. (OR skip if POST-MVP)

**Verify Tests Pass**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_Services` and confirm tests pass (OR skip if POST-MVP)

### Refactor for User Story 5 (REFACTOR Phase)

- [ ] T086 [US5] Add detailed error messages for services validation. (OR skip if POST-MVP)
- [ ] T087 [US5] Add debug logging for services serialization/deserialization. (OR skip if POST-MVP)
- [ ] T088 [US5] Optimize services serialization code. (OR skip if POST-MVP)

### Documentation for User Story 5

- [ ] T089 [US5] Create services example: Create /workspace/examples/resources/bcm_cmdevice_category/services.tf with service configuration examples based on research. (OR skip if POST-MVP)
- [ ] T090 [US5] Generate provider documentation: Run `make generate` to update docs with services field documentation. (OR add POST-MVP note to docs)

**Checkpoint**: All user stories should now be independently functional (OR 4/5 complete if services marked POST-MVP)

---

## Phase 7: Polish & Cross-Cutting Concerns

**Purpose**: Improvements and validation that affect the overall implementation

- [ ] T091 [P] Extract common field mapping helpers: If repetitive snake_case ↔ camelCase conversion patterns emerge across multiple fields (3+), extract to helper functions in /workspace/internal/provider/helpers.go (e.g., toCamelCase, toSnakeCase, serializeList, deserializeList) - ONLY if DRY principle clearly applies
- [ ] T092 Code review and cleanup: Review all 5 field implementations for consistency in error handling, naming conventions, and patterns. Ensure all field code follows same structure for maintainability
- [ ] T093 Run full test suite: Execute `make testacc` to run all acceptance tests for category resource including new dynamic field tests. Verify all existing tests still pass and new tests (35+ scenarios for 5 fields) complete successfully
- [ ] T094 Update base resource example: Update /workspace/examples/resources/bcm_cmdevice_category/resource.tf to include brief comments indicating where to find specific field examples (static-routes.tf, fsexports.tf, etc.)
- [ ] T095 Final documentation generation: Run `make generate` one final time to ensure all field documentation is up-to-date in /workspace/docs/resources/bcm_cmdevice_category.md, verify markdown formatting, attribute tables, and examples are all correct
- [ ] T096 Verify CheckDestroy enhancement: Review testAccCheckCMDeviceCategoryDestroy function in test file to ensure it properly verifies deletion of categories with newly implemented dynamic fields
- [ ] T097 [P] Security review: Verify no sensitive data (network UUIDs, role names, file paths) is logged at INFO level, ensure debug logging uses tflog.Debug not tflog.Info for field transformation details

---

## Dependencies & Execution Order

### Phase Dependencies

- **Phase 0: Research (T001-T005)**: No dependencies - all research tasks can run in parallel. This is the starting point.
- **Phase 1: Foundational (T006-T010)**: Depends on Phase 0 completion for API structure knowledge. BLOCKS all user story phases.
- **Phase 2: User Story 1 (T011-T026)**: Depends on Phase 1 completion - static_routes schema must be defined
- **Phase 3: User Story 2 (T027-T042)**: Depends on Phase 1 completion - fsexports schema must be defined. Can run in parallel with Phase 2 if desired
- **Phase 4: User Story 3 (T043-T058)**: Depends on Phase 1 completion - roles schema must be defined. Can run in parallel with Phases 2-3 if desired
- **Phase 5: User Story 4 (T059-T074)**: Depends on Phase 1 completion - gpu_settings schema must be defined. Can run in parallel with Phases 2-4 if desired
- **Phase 6: User Story 5 (T075-T090)**: Depends on Phase 1 completion - services schema must be defined. Can run in parallel with Phases 2-5 if desired. May be skipped if marked POST-MVP
- **Phase 7: Polish (T091-T097)**: Depends on completion of desired user story phases (minimally Phase 2 for MVP, ideally all phases)

### User Story Dependencies

- **User Story 1 (P1) - static_routes**: Can start after Foundational (Phase 1) - No dependencies on other stories
- **User Story 2 (P1) - fsexports**: Can start after Foundational (Phase 1) - No dependencies on other stories, can run parallel to US1
- **User Story 3 (P2) - roles**: Can start after Foundational (Phase 1) - No dependencies on other stories, can run parallel to US1-2
- **User Story 4 (P3) - gpu_settings**: Can start after Foundational (Phase 1) - No dependencies on other stories, can run parallel to US1-3
- **User Story 5 (P3) - services**: Can start after Foundational (Phase 1) - No dependencies on other stories, can run parallel to US1-4, or skip if POST-MVP

### Within Each User Story (TDD RED-GREEN-REFACTOR Pattern)

**RED Phase (Tests)**:
- All 7 test tasks marked [P] can be written in parallel (different test functions, no dependencies)
- Tests MUST be written first and verified to FAIL before implementation
- Verification step: Run test suite and confirm failures

**GREEN Phase (Implementation)**:
- Schema definition task (e.g., T018) MUST complete before serialization/deserialization tasks
- Serialization (buildAPIEntity) and deserialization (readCategory) can be done in parallel after schema complete
- All implementation tasks must complete before running tests
- Verification step: Run test suite and confirm all 7 tests PASS

**REFACTOR Phase**:
- Error messages, logging, and optimization tasks can run in parallel (different code sections)
- Should maintain all tests passing
- Verification step: Run tests after each refactor, confirm still passing

**Documentation Phase**:
- Example creation and doc generation can run in parallel
- Should be done after implementation is stable
- Verification step: Review generated docs for accuracy

### Parallel Opportunities

**Phase 0 (Research)**:
- ALL 5 research tasks (T001-T005) can run in parallel - different API queries, different output files

**Phase 1 (Foundational)**:
- ALL 5 model struct definitions (T006-T010) can run in parallel - different structs, different code sections in same file

**Within Each User Story - RED Phase (Tests)**:
```bash
# Example: User Story 1 - All 7 tests can be written in parallel
Task T011: TestAccCMDeviceCategory_StaticRoutesBasicCRUD
Task T012: TestAccCMDeviceCategory_StaticRoutesIdempotencyCreate
Task T013: TestAccCMDeviceCategory_StaticRoutesIdempotencyUpdate
Task T014: TestAccCMDeviceCategory_StaticRoutesImport
Task T015: TestAccCMDeviceCategory_StaticRoutesDrift
Task T016: TestAccCMDeviceCategory_StaticRoutesEmpty
Task T017: TestAccCMDeviceCategory_StaticRoutesValidation
```

**Across User Stories - After Phase 1 Complete**:
- User Stories 1-5 (Phases 2-6) can ALL run in parallel by different developers or in sequence by priority (P1 → P2 → P3)
- Example parallel team approach:
  - Developer A: User Story 1 (static_routes) - P1
  - Developer B: User Story 2 (fsexports) - P1
  - Developer C: User Story 3 (roles) - P2
  - Developer D: User Story 4 (gpu_settings) - P3
  - Developer E: User Story 5 (services) - P3

**Phase 7 (Polish)**:
- Tasks T091, T092, T097 can run in parallel (different code review aspects)
- T093, T094, T095, T096 should run sequentially (test suite → cleanup → doc generation → verify)

---

## Implementation Strategy

### MVP First (User Stories 1-2 Only - Both P1 Fields)

**Estimated Time**: 8-12 hours

1. Complete Phase 0: Research (T001-T005) - **~2-3 hours**
   - Run BCM API queries for all 5 fields in parallel
   - Document findings in sampleRest/category-dynamic-fields/*.json

2. Complete Phase 1: Foundational (T006-T010) - **~1 hour**
   - Define all 5 model structs in parallel

3. Complete Phase 2: User Story 1 - static_routes (T011-T026) - **~3-4 hours**
   - RED: Write 7 tests, verify failures (~1 hour)
   - GREEN: Implement schema + CRUD, verify tests pass (~1.5 hours)
   - REFACTOR: Error messages, logging, optimization (~30 min)
   - DOCS: Example + generate docs (~30 min)
   - **VALIDATE**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_StaticRoutes`

4. Complete Phase 3: User Story 2 - fsexports (T027-T042) - **~3-4 hours**
   - RED: Write 7 tests, verify failures (~1 hour)
   - GREEN: Implement schema + CRUD, verify tests pass (~1.5 hours)
   - REFACTOR: Error messages, logging, optimization (~30 min)
   - DOCS: Example + generate docs (~30 min)
   - **VALIDATE**: Run `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_FSExports`

5. **STOP and VALIDATE MVP** - **~30 min**
   - Run full test suite: `make testacc`
   - Test both P1 fields independently
   - Verify no regressions in existing category tests
   - Review generated documentation

6. Optional: Complete Phase 7: Polish (T091-T097) for MVP quality - **~1-2 hours**

**MVP Deliverables**:
- 2/5 dynamic fields replaced (static_routes, fsexports)
- 14+ passing acceptance tests (7 per field)
- Updated documentation with examples
- Production-ready P1 functionality

### Full Implementation (All 5 Fields)

**Estimated Time**: 20-31 hours total

Follow MVP approach, then add:

7. Complete Phase 4: User Story 3 - roles (T043-T058) - **~3-4 hours** (P2 priority)
8. Complete Phase 5: User Story 4 - gpu_settings (T059-T074) - **~3-4 hours** (P3 priority)
9. Complete Phase 6: User Story 5 - services (T075-T090) - **~3-4 hours** (P3 priority, OR skip if POST-MVP)
10. Complete Phase 7: Polish (T091-T097) - **~1-2 hours** (comprehensive cleanup)

**Full Deliverables**:
- 5/5 dynamic fields replaced (OR 4/5 if services POST-MVP)
- 35+ passing acceptance tests (7 per field × 5 fields)
- 5 example files demonstrating each field
- Complete generated documentation
- All success criteria met (SC-001 through SC-010 from spec.md)

### Incremental Delivery Strategy

**Phase 0-1 (Foundation)**: ~3-4 hours
- Research + model definitions
- **Deliverable**: API contracts documented, schemas defined

**Phase 2 (MVP - First P1 Field)**: +3-4 hours (cumulative: ~6-8 hours)
- Implement static_routes
- **Deliverable**: One dynamic field replaced, 7 tests passing, can deploy/demo

**Phase 3 (MVP Complete - Both P1 Fields)**: +3-4 hours (cumulative: ~9-12 hours)
- Implement fsexports
- **Deliverable**: Both P1 fields complete, 14 tests passing, MVP ready for production

**Phase 4 (P2 Field)**: +3-4 hours (cumulative: ~12-16 hours)
- Implement roles
- **Deliverable**: 3/5 fields complete, role-based organization enabled

**Phase 5 (P3 Field #1)**: +3-4 hours (cumulative: ~15-20 hours)
- Implement gpu_settings
- **Deliverable**: 4/5 fields complete, GPU cluster support enabled

**Phase 6 (P3 Field #2)**: +3-4 hours (cumulative: ~18-24 hours)
- Implement services (OR skip if POST-MVP)
- **Deliverable**: 5/5 fields complete (or 4/5 with services deferred)

**Phase 7 (Polish)**: +1-2 hours (cumulative: ~19-26 hours)
- Final cleanup and documentation
- **Deliverable**: Production-quality implementation with comprehensive test coverage

Each phase adds value incrementally without breaking previous functionality.

---

## Validation Checkpoints

### After Phase 0 (Research)
- [ ] All 5 JSON files exist in sampleRest/category-dynamic-fields/
- [ ] Each file contains at least one complete API response example
- [ ] Field names and types are documented
- [ ] Decision made on services field (implement or POST-MVP)

### After Phase 1 (Foundational)
- [ ] All 5 model structs defined in resource_cmdevice_category.go
- [ ] Structs follow terraform-plugin-framework conventions (types.String, types.Bool, types.Int64)
- [ ] Code compiles without errors: `go build ./internal/provider/`

### After Each User Story Implementation
- [ ] All 7 acceptance tests for the field PASS
- [ ] Example file created in examples/resources/bcm_cmdevice_category/
- [ ] Documentation generated with `make generate`
- [ ] No regressions in existing tests: `make testacc` still passes
- [ ] Empty list handling verified ([] remains [] in state, not null)
- [ ] Drift detection test confirms external BCM changes are detected

### After Phase 7 (Polish)
- [ ] Full test suite passes: `make testacc` completes successfully
- [ ] Test execution time under 120 minutes (acceptance test timeout)
- [ ] All 5 field examples exist and demonstrate proper usage
- [ ] Generated docs in docs/resources/bcm_cmdevice_category.md are accurate
- [ ] No dynamic types remain for completed fields (grep for types.Dynamic)
- [ ] Code review complete: consistent patterns across all fields
- [ ] Security review complete: no sensitive data in INFO logs

### Success Criteria Validation (from spec.md)

Before marking implementation complete, verify ALL success criteria:

- [ ] **SC-001**: Grep for `types.Dynamic` in resource file, confirm only placeholders for POST-MVP fields remain
- [ ] **SC-002**: Count test functions: 7 × (number of implemented fields) test functions exist and pass
- [ ] **SC-003**: BCM API service structure documented in sampleRest/category-dynamic-fields/services.json (or POST-MVP note)
- [ ] **SC-004**: Run drift tests for all fields, verify CRUD operations don't lose data
- [ ] **SC-005**: Create category with empty lists [], verify state shows empty arrays not null
- [ ] **SC-006**: Drift detection tests complete in <5 seconds (time the external BCM modification detection)
- [ ] **SC-007**: Review buildAPIEntity and readCategory, verify bidirectional snake_case ↔ camelCase mapping
- [ ] **SC-008**: Check docs/resources/bcm_cmdevice_category.md for all field documentation with examples
- [ ] **SC-009**: Test null values in nested objects (e.g., metric in static routes), verify no runtime errors
- [ ] **SC-010**: Test invalid CIDR and IP formats, verify validators catch errors pre-API with clear messages

---

## Notes

- **[P] marker**: Task can run in parallel with other [P] tasks in same phase (different files or independent code sections)
- **[Story] label**: Maps task to specific user story for traceability (US1-US5)
- **TDD discipline**: RED-GREEN-REFACTOR cycle is mandatory - verify tests FAIL before implementation, PASS after implementation
- **Independent stories**: Each user story should be fully functional and testable on its own after completion
- **Commit strategy**: Commit after completing each phase or logical group (e.g., after GREEN phase for each story)
- **Test timing**: Acceptance tests can be slow (~2-5 minutes per field), run field-specific tests during development: `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceCategory_StaticRoutes`
- **BCM eventual consistency**: Read operations may need retry logic (already implemented in resource), be patient with test execution
- **Field name mapping**: Critical for correctness - verify camelCase in buildAPIEntity, snake_case in readCategory, test with drift detection
- **Empty list semantics**: Use `types.ListValueMust(elementType, []attr.Value{})` for empty arrays, never `types.ListNull()` when BCM API returns `[]`
- **Acceptance test environment**: Requires BCM cluster at 172.21.15.254:8081 with valid credentials (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- **POST-MVP decisions**: If services field structure is unclear from Phase 0 research, document findings and mark T075-T090 as POST-MVP
- **Parallel execution**: P1 fields (static_routes, fsexports) can be implemented in parallel by different developers after Phase 1

**Avoid These Anti-Patterns**:
- ❌ Starting implementation before tests are written and verified to fail
- ❌ Marking tests as passing when they actually skip or have incomplete assertions
- ❌ Converting empty BCM arrays `[]` to null in state (breaks empty list preservation - SC-005)
- ❌ Propagating Unknown values to state (causes "invalid result object" errors)
- ❌ Forgetting baseType field in BCM API serialization (causes API validation errors)
- ❌ Exposing baseType in Terraform schema (internal BCM detail, not user-relevant)
- ❌ Skipping drift detection tests (these catch critical field mapping bugs)
- ❌ Hardcoding test values that may not exist in all BCM environments
- ❌ Running tests without TF_ACC=1 environment variable (tests will be skipped)
- ❌ Implementing multiple fields before completing test coverage for previous field
