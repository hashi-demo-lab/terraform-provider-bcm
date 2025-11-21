# Tasks: BCM Device Resource Management

**Feature**: `bcm_cmdevice_device` resource for managing individual devices (compute nodes) in BCM clusters
**Branch**: `008-cmdevice-device-resource`
**Input**: Design documents from `/workspace/specs/008-cmdevice-device-resource/`

**Implementation Approach**: TDD RED-GREEN-REFACTOR with modern terraform-plugin-testing patterns

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **Checkbox**: `- [ ]` (markdown task checkbox)
- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 0: Research & Discovery

**Goal**: Resolve unknowns from Technical Context, validate BCM device API, document validation approaches

**Prerequisites**: BCM test cluster access, existing bcm_client.go and test_helpers.go

### Research Tasks

- [ ] T001 Verify BCM device API methods and args parameter support using test script in /workspace/sampleRest/
- [ ] T002 [P] Document childType assignment logic by examining existing devices via BCM API
- [ ] T003 [P] Identify force parameter scenarios for create/update/delete operations
- [ ] T004 [P] Research network interface configuration patterns and ordering requirements
- [ ] T005 [P] Document validation implementations for hostname (RFC 1123), MAC address, UUID (RFC 4122)
- [ ] T006 Create /workspace/specs/008-cmdevice-device-resource/research.md with all findings
- [ ] T007 Verify async operation behavior (if device creation requires polling like software images)

**Completion Criteria**: research.md exists with confirmed API signatures, childType logic, force scenarios, validation patterns

---

## Phase 1: Design & Contracts

**Goal**: Generate data model, API contracts, quickstart guide

**Dependencies**: Phase 0 (research.md) must be complete

### Design Artifacts

- [ ] T008 Create /workspace/specs/008-cmdevice-device-resource/data-model.md with complete Device entity schema
- [ ] T009 [P] Create /workspace/specs/008-cmdevice-device-resource/contracts/device-api.json with JSON-RPC contract
- [ ] T010 [P] Create /workspace/specs/008-cmdevice-device-resource/quickstart.md with TDD workflow guide
- [ ] T011 Run /workspace/.specify/scripts/bash/update-agent-context.sh copilot to update agent context

**Completion Criteria**: data-model.md, contracts/device-api.json, quickstart.md exist; agent context updated

---

## Phase 2: User Story 1 - Create and Manage Individual Cluster Nodes (Priority: P1) 🎯 MVP

**Goal**: Allow cluster administrators to create and manage individual devices using Terraform with full CRUD operations

**Independent Test**: Create a single device with minimal required fields (hostname, MAC, category, management network), verify it appears in BCM, update it, import it, and delete it

**Why This Story**: Core value proposition - enables Infrastructure as Code for cluster topology

### RED Phase: Write Failing Tests FIRST

> **CRITICAL**: These tests MUST be written first and MUST fail before any implementation

- [ ] T012 [P] [US1] Create /workspace/internal/provider/resource_cmdevice_device_test.go with required imports (terraform-plugin-testing v1.13.3+)
- [ ] T013 [P] [US1] Write testAccCMDeviceDevicePreCheck function with cleanup using verifyResourceDeleted helper (5 retries)
- [ ] T014 [P] [US1] Write testAccCheckCMDeviceDeviceDestroy function using createTestBCMClient and verifyResourceDeleted
- [ ] T015 [US1] Write testAccCMDeviceDeviceResourceConfig_Basic helper function with provider config from environment variables
- [ ] T016 [US1] Write testAccCMDeviceDeviceResourceConfig_Updated helper function for update scenarios
- [ ] T017 [US1] Write TestAccCMDeviceDeviceResource_Basic with full CRUD lifecycle (Create, Idempotency, Import, Update, Idempotency, Delete)
- [ ] T018 [US1] Add statecheck.ExpectKnownValue assertions for hostname, uuid, category, management_network
- [ ] T019 [US1] Add plancheck.ExpectEmptyPlan checks after Create and Update for idempotency verification
- [ ] T020 [US1] Add compare.ValuesSame() ID consistency tracking across all test steps
- [ ] T021 [US1] Run RED phase verification: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDeviceResource_Basic

**RED Phase Checkpoint**: ❌ All tests must FAIL with "resource not implemented" error

### GREEN Phase: Minimal Implementation to Pass Tests

> **CRITICAL**: Write ONLY enough code to make tests pass - no extra features

- [ ] T022 [US1] Create /workspace/internal/provider/resource_cmdevice_device.go with basic structure and imports
- [ ] T023 [US1] Define CMDeviceDeviceResourceModel struct with required fields only (ID, UUID, Hostname, MAC, Category, ManagementNetwork)
- [ ] T024 [US1] Implement Metadata() method for resource type bcm_cmdevice_device
- [ ] T025 [US1] Implement Schema() method with required attributes and validators (hostname RFC 1123, MAC address format, UUID format)
- [ ] T026 [US1] Implement Create() method with minimal BCM API call using client.CallJSONRPC("cmdevice", "addDevice", deviceEntity, false)
- [ ] T027 [US1] Implement Read() method with efficient direct lookup using client.CallJSONRPC("cmdevice", "getDevice", identifier)
- [ ] T028 [US1] Handle external deletion detection in Read() - remove from state if device not found
- [ ] T029 [US1] Implement Update() method with client.CallJSONRPC("cmdevice", "updateDevice", deviceEntity, false)
- [ ] T030 [US1] Implement Delete() method with client.CallJSONRPC("cmdevice", "removeDevice", uuid, false)
- [ ] T031 [US1] Implement ImportState() method using resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
- [ ] T032 [US1] Implement buildDeviceAPIEntity helper function to construct BCM entity JSON (baseType, childType, modified, to_be_removed)
- [ ] T033 [US1] Implement parseDeviceFromAPI helper function with null-safe field extraction
- [ ] T034 [US1] Register resource in /workspace/internal/provider/provider.go Resources() method
- [ ] T035 [US1] Run GREEN phase verification: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDeviceResource_Basic

**GREEN Phase Checkpoint**: ✅ All tests must PASS - minimal implementation working

### REFACTOR Phase: Add Full Implementation

> **CRITICAL**: Improve code quality while keeping tests GREEN

- [ ] T036 [US1] Add optional fields to schema (notes, kernel_parameters, boot_loader, boot_loader_protocol, force)
- [ ] T037 [US1] Add computed fields to schema (creation_time, base_type, child_type)
- [ ] T038 [US1] Enhance buildDeviceAPIEntity to handle all optional fields with null safety
- [ ] T039 [US1] Enhance parseDeviceFromAPI to extract all fields (required, optional, computed)
- [ ] T040 [US1] Add error handling for duplicate hostname with clear diagnostic message
- [ ] T041 [US1] Add error handling for invalid category/network references with clear diagnostic messages
- [ ] T042 [US1] Add error handling for force parameter scenarios with helpful suggestions
- [ ] T043 [US1] Add terraform-plugin-log (tflog) statements for debugging Create/Read/Update/Delete operations
- [ ] T044 [US1] Run REFACTOR phase verification: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDeviceResource_Basic

**REFACTOR Phase Checkpoint**: ✅ Tests still PASS with full implementation

### Story Completion

- [ ] T045 [US1] Create /workspace/examples/resources/bcm_cmdevice_device/resource.tf with basic example
- [ ] T046 [US1] Run make generate to create documentation in /workspace/docs/resources/bcm_cmdevice_device.md
- [ ] T047 [US1] Manually test: Create device → Import → Update → Delete cycle with real BCM cluster
- [ ] T048 [US1] Verify idempotency: Run terraform apply twice with no changes between applies

**User Story 1 Complete**: ✅ Basic device resource fully functional with CRUD, import, validation

---

## Phase 3: User Story 2 - Import Existing Devices into Terraform Management (Priority: P2)

**Goal**: Allow cluster administrators to import existing BCM devices into Terraform state for brownfield infrastructure management

**Independent Test**: Manually create a device in BCM UI, run terraform import, verify all fields populated correctly, run terraform plan with no changes

**Why This Story**: Essential for adoption in existing clusters - most users have pre-existing devices

**Dependencies**: User Story 1 must be complete (basic CRUD working)

### RED Phase: Import-Specific Tests

- [ ] T049 [US2] Write testAccCMDeviceDeviceResourceConfig_ImportPrep helper to create device directly via BCM API
- [ ] T050 [US2] Write TestAccCMDeviceDevice_ImportExistingDevice with manual BCM device creation in PreConfig
- [ ] T051 [US2] Add state checks to verify all fields populated correctly after import (hostname, MAC, category, UUID, creation_time)
- [ ] T052 [US2] Add plancheck.ExpectEmptyPlan after import to verify idempotency
- [ ] T053 [US2] Run RED phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_Import

**RED Phase Checkpoint**: ❌ Test should fail if import doesn't populate all fields correctly

### GREEN Phase: Import Enhancement

- [ ] T054 [US2] Verify ImportState implementation handles UUID correctly
- [ ] T055 [US2] Verify Read() populates all fields from BCM API response (required, optional, computed)
- [ ] T056 [US2] Add null safety to parseDeviceFromAPI for fields that may not exist on imported devices
- [ ] T057 [US2] Run GREEN phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_Import

**GREEN Phase Checkpoint**: ✅ Import test passes

### REFACTOR Phase: Import Documentation

- [ ] T058 [US2] Add import example to /workspace/examples/resources/bcm_cmdevice_device/import.tf
- [ ] T059 [US2] Run make generate to update documentation with import instructions
- [ ] T060 [US2] Manually test: Import 3 different existing devices from BCM, verify all scenarios work

**User Story 2 Complete**: ✅ Import functionality working for brownfield devices

---

## Phase 4: User Story 3 - Drift Detection for External Modifications (Priority: P2)

**Goal**: Detect when devices are modified externally in BCM (outside Terraform) and restore desired state

**Independent Test**: Create device via Terraform, modify it directly in BCM API, run terraform plan, verify drift detected, apply to restore

**Why This Story**: Critical for multi-user environments where devices may be modified outside Terraform

**Dependencies**: User Story 1 must be complete (Read operation working)

### RED Phase: Drift Detection Tests

- [ ] T061 [US3] Write testAccCMDeviceDeviceResourceConfig_Drift helper with configurable attribute values
- [ ] T062 [US3] Write TestAccCMDeviceDevice_DriftHostname with PreConfig external modification via BCM API
- [ ] T063 [US3] Add PreConfig logic to modify hostname externally: getDevice → modify → updateDevice → sleep 2s
- [ ] T064 [US3] Add plancheck.ExpectNonEmptyPlan to verify drift detected
- [ ] T065 [US3] Add final step to verify Terraform restores original hostname value
- [ ] T066 [US3] Write TestAccCMDeviceDevice_DriftKernelParameters with PreConfig external modification
- [ ] T067 [US3] Write TestAccCMDeviceDevice_DriftNotes with PreConfig external modification
- [ ] T068 [US3] Run RED phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_Drift

**RED Phase Checkpoint**: ❌ Tests may fail if Read() doesn't properly detect external changes

### GREEN Phase: Drift Detection Implementation

- [ ] T069 [US3] Verify Read() implementation compares BCM API response with current state
- [ ] T070 [US3] Verify Read() sets all attributes from API (not just computed fields)
- [ ] T071 [US3] Test Read() handles field name mapping correctly (camelCase → snake_case)
- [ ] T072 [US3] Run GREEN phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_Drift

**GREEN Phase Checkpoint**: ✅ All drift detection tests pass

### REFACTOR Phase: Drift Documentation

- [ ] T073 [US3] Add drift detection example to /workspace/examples/resources/bcm_cmdevice_device/drift-example.md
- [ ] T074 [US3] Document external modification scenarios in resource documentation
- [ ] T075 [US3] Manually test: Create device, modify in BCM UI, verify terraform plan shows drift

**User Story 3 Complete**: ✅ Drift detection working for all modified fields

---

## Phase 5: User Story 4 - Schema Validation (Priority: P2)

**Goal**: Catch invalid configurations (malformed hostnames, MAC addresses, UUIDs) before making API calls

**Independent Test**: Attempt to create devices with invalid values, verify Terraform returns validation errors before API call

**Why This Story**: Improves user experience by catching errors early with clear messages

**Dependencies**: User Story 1 (schema defined)

### RED Phase: Validation Tests

- [ ] T076 [US4] Write TestAccCMDeviceDevice_ValidationInvalidHostname with multiple invalid cases (uppercase, special chars, too long, leading hyphen)
- [ ] T077 [US4] Add ExpectError checks with regexp.MustCompile for each validation message
- [ ] T078 [US4] Write TestAccCMDeviceDevice_ValidationInvalidMAC with invalid formats (dashes, missing octets, non-hex)
- [ ] T079 [US4] Write TestAccCMDeviceDevice_ValidationInvalidUUID with invalid category and management_network UUIDs
- [ ] T080 [US4] Run RED phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_Validation

**RED Phase Checkpoint**: ❌ Tests fail if validators not implemented correctly

### GREEN Phase: Validation Implementation

- [ ] T081 [US4] Add stringvalidator.RegexMatches for hostname (RFC 1123: ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$)
- [ ] T082 [US4] Add stringvalidator.RegexMatches for MAC address (^([0-9A-Fa-f]{2}:){5}[0-9A-Fa-f]{2}$)
- [ ] T083 [US4] Add stringvalidator.RegexMatches for UUID (RFC 4122: ^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$)
- [ ] T084 [US4] Verify all validators have clear, helpful error messages
- [ ] T085 [US4] Run GREEN phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_Validation

**GREEN Phase Checkpoint**: ✅ All validation tests pass

### REFACTOR Phase: Validation Documentation

- [ ] T086 [US4] Document validation rules in /workspace/examples/resources/bcm_cmdevice_device/validation-examples.tf
- [ ] T087 [US4] Run make generate to include validation rules in resource documentation
- [ ] T088 [US4] Manually test: Try each invalid case, verify clear error messages appear

**User Story 4 Complete**: ✅ Schema validation prevents invalid configurations

---

## Phase 6: User Story 5 - Configure Network Interfaces (Priority: P3)

**Goal**: Allow configuration of multiple network interfaces with IP addresses and network references

**Independent Test**: Create device with 2 network interfaces (management + data), verify both interfaces configured correctly in BCM

**Why This Story**: Advanced configuration for production clusters requiring multiple networks

**Dependencies**: User Story 1 (basic CRUD working)

### RED Phase: Network Interface Tests

- [ ] T089 [US5] Define NetworkInterfaceModel nested struct in resource_cmdevice_device.go
- [ ] T090 [US5] Write testAccCMDeviceDeviceResourceConfig_WithInterfaces helper with multiple interface configurations
- [ ] T091 [US5] Write TestAccCMDeviceDevice_NetworkInterfaces with 2 interfaces (management + data network)
- [ ] T092 [US5] Add state checks for interfaces list size using knownvalue.ListSizeExact(2)
- [ ] T093 [US5] Add state checks for individual interface attributes (name, MAC, IP, network UUID)
- [ ] T094 [US5] Run RED phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_NetworkInterfaces

**RED Phase Checkpoint**: ❌ Test fails because interfaces attribute not implemented

### GREEN Phase: Network Interface Implementation

- [ ] T095 [US5] Add interfaces attribute to schema as types.List with NetworkInterfaceModel nested object
- [ ] T096 [US5] Add MAC address validator to interface MAC field
- [ ] T097 [US5] Add UUID validator to interface network field
- [ ] T098 [US5] Implement buildInterfacesList helper to construct interface JSON for BCM API
- [ ] T099 [US5] Implement parseInterfacesFromAPI helper to extract interfaces from BCM response
- [ ] T100 [US5] Update buildDeviceAPIEntity to include interfaces if configured
- [ ] T101 [US5] Update parseDeviceFromAPI to extract interfaces from BCM response
- [ ] T102 [US5] Run GREEN phase: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice_NetworkInterfaces

**GREEN Phase Checkpoint**: ✅ Network interfaces test passes

### REFACTOR Phase: Interface Documentation

- [ ] T103 [US5] Create /workspace/examples/resources/bcm_cmdevice_device/network-interfaces.tf with multi-interface example
- [ ] T104 [US5] Run make generate to update documentation with interface configuration
- [ ] T105 [US5] Manually test: Create device with 3 interfaces, verify all configured correctly

**User Story 5 Complete**: ✅ Multiple network interfaces supported

---

## Phase 7: Polish & Cross-Cutting Concerns

**Goal**: Final improvements, documentation, validation

**Dependencies**: All desired user stories complete

### Documentation & Examples

- [ ] T106 [P] Create /workspace/examples/resources/bcm_cmdevice_device/compute-node.tf with realistic compute node configuration
- [ ] T107 [P] Create /workspace/examples/resources/bcm_cmdevice_device/head-node.tf with head node configuration
- [ ] T108 [P] Create /workspace/examples/resources/bcm_cmdevice_device/storage-node.tf with storage node configuration
- [ ] T109 Run make generate to regenerate all documentation with complete examples
- [ ] T110 Review /workspace/docs/resources/bcm_cmdevice_device.md for accuracy and completeness

### Code Quality

- [ ] T111 [P] Run make fmt to format all Go code
- [ ] T112 [P] Run make lint to check for Go linting issues
- [ ] T113 [P] Run pre-commit run --all-files to verify all pre-commit hooks pass
- [ ] T114 Add code comments to complex helper functions (buildDeviceAPIEntity, parseDeviceFromAPI)
- [ ] T115 Review error messages for clarity and actionability

### Testing & Validation

- [ ] T116 Run full acceptance test suite: TF_ACC=1 go test -v -timeout 120m /workspace/internal/provider/ -run TestAccCMDeviceDevice
- [ ] T117 Verify 100% test pass rate across all test cases
- [ ] T118 Run /workspace/scripts/test-examples.sh --resources-only to test all example configurations
- [ ] T119 Validate quickstart.md instructions by following them end-to-end
- [ ] T120 Manual testing: Create → Import → Update → Delete cycle with all optional fields populated

### Documentation Updates

- [ ] T121 Update /workspace/CLAUDE.md with any new patterns discovered (if applicable)
- [ ] T122 Update /workspace/specs/008-cmdevice-device-resource/quickstart.md with final implementation notes
- [ ] T123 Document any force parameter scenarios discovered during testing in research.md

---

## Dependencies & Execution Order

### Phase Dependencies

```
Phase 0 (Research) → Phase 1 (Design) → Phase 2 (US1: Basic CRUD) ┐
                                                                    ├→ Phase 7 (Polish)
                                        Phase 3 (US2: Import)     ─┤
                                        Phase 4 (US3: Drift)      ─┤
                                        Phase 5 (US4: Validation)─┤
                                        Phase 6 (US5: Interfaces)─┘
```

- **Phase 0 & 1**: Sequential prerequisites (research → design)
- **Phase 2**: BLOCKS all other user stories (foundation must work)
- **Phase 3-6**: Can proceed in parallel after Phase 2 complete OR sequentially in priority order
- **Phase 7**: Depends on all implemented user stories

### User Story Dependencies

- **User Story 1 (P1)**: No dependencies after Phase 1 - MUST be implemented first
- **User Story 2 (P2)**: Depends on US1 (needs Read working)
- **User Story 3 (P2)**: Depends on US1 (needs Read working)
- **User Story 4 (P2)**: Depends on US1 (needs Schema defined)
- **User Story 5 (P3)**: Depends on US1 (needs CRUD working)

### TDD Workflow Within Each Story

```
RED (Write failing tests) → GREEN (Minimal implementation) → REFACTOR (Full implementation)
```

**CRITICAL**: Never proceed to GREEN until tests are written and failing. Never proceed to REFACTOR until tests are passing.

### Parallel Opportunities

- **Phase 0**: T002, T003, T004, T005 (research tasks on different topics)
- **Phase 1**: T009, T010 (contracts and quickstart independent)
- **Phase 2 RED**: T012, T013, T014 (test helper functions independent)
- **Phase 7 Docs**: T106, T107, T108 (example files independent)
- **Phase 7 Quality**: T111, T112, T113 (fmt, lint, pre-commit independent)

---

## Parallel Example: Phase 2 RED Phase (User Story 1)

```bash
# Launch test helper functions in parallel (different concerns):
Task T013: "Write testAccCMDeviceDevicePreCheck function"
Task T014: "Write testAccCheckCMDeviceDeviceDestroy function"
Task T015: "Write testAccCMDeviceDeviceResourceConfig_Basic helper"
```

---

## Implementation Strategy

### MVP First (Phase 2 Only)

1. Complete Phase 0: Research (resolve unknowns)
2. Complete Phase 1: Design (generate artifacts)
3. Complete Phase 2: User Story 1 - Basic CRUD
4. **STOP and VALIDATE**: Run all tests, manually test with real cluster
5. **DEPLOY/DEMO**: MVP is now usable for basic device management

**MVP Delivers**: Users can create, read, update, delete, and import devices with minimal configuration

### Incremental Delivery

1. **Foundation** (Phase 0-1): Research + Design → 7 tasks
2. **MVP** (Phase 2): User Story 1 → 37 tasks → **Deploy/Demo** ✅
3. **Import** (Phase 3): User Story 2 → 12 tasks → **Deploy/Demo** ✅
4. **Drift** (Phase 4): User Story 3 → 15 tasks → **Deploy/Demo** ✅
5. **Validation** (Phase 5): User Story 4 → 13 tasks → **Deploy/Demo** ✅
6. **Interfaces** (Phase 6): User Story 5 → 17 tasks → **Deploy/Demo** ✅
7. **Polish** (Phase 7): Final improvements → 18 tasks → **Release** 🚀

**Each story adds value without breaking previous stories**

### Parallel Team Strategy

With multiple developers after Phase 2 complete:

- **Developer A**: User Story 3 (Drift Detection)
- **Developer B**: User Story 4 (Validation)
- **Developer C**: User Story 5 (Network Interfaces)

All stories integrate independently since they build on the foundation from Phase 2.

---

## Verification Checklist

**Before considering feature complete, verify:**

- [ ] All acceptance tests pass at 100% (run full suite)
- [ ] All example configurations in /workspace/examples/ validate and plan successfully
- [ ] Documentation auto-generated and reviewed for accuracy
- [ ] Import functionality tested manually with existing BCM devices
- [ ] Drift detection tested manually (create → modify in BCM → plan shows drift)
- [ ] Idempotency verified (apply twice with no changes → empty plan both times)
- [ ] Validation tested (try invalid hostname/MAC/UUID → clear error before API call)
- [ ] quickstart.md instructions followed end-to-end successfully
- [ ] All Go code formatted (make fmt) and linted (make lint)
- [ ] All pre-commit hooks passing

---

## Notes

- **[P] marker**: Tasks can run in parallel (different files, no blocking dependencies)
- **[Story] marker**: Maps task to specific user story for traceability
- **TDD discipline**: Always RED → GREEN → REFACTOR, never skip steps
- **Test-first mindset**: Write tests, watch them fail, then implement
- **Incremental delivery**: Each phase checkpoint is a potential stopping point for validation
- **Modern testing patterns**: Use statecheck.ExpectKnownValue, plancheck.ExpectEmptyPlan, compare.ValuesSame
- **File paths**: All paths are absolute from repository root (/workspace/)
- **BCM API**: Use args parameter for efficient direct lookup (getDevice) vs list+filter (getNodes)

**Total Tasks**: 123 tasks across 7 phases
**Estimated MVP**: 44 tasks (Phase 0 + Phase 1 + Phase 2)
**Priority Recommendation**: Implement sequentially by phase (0→1→2→3→4→5→6→7) OR in parallel after Phase 2 (3, 4, 5, 6 can proceed simultaneously)
