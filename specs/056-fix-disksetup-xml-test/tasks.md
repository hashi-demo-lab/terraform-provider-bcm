# Tasks: Fix Disk Setup XML Validation Test

**Input**: Design documents from `/workspace/specs/056-fix-disksetup-xml-test/`
**Prerequisites**: plan.md (required), spec.md (required)

**Tests**: This is a test fix feature - we're correcting test data, not adding new tests.

**Organization**: Tasks are organized by user story to enable independent implementation and testing.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (e.g., US1, US2, US3)
- Include exact file paths in descriptions

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Extract valid XML and verify current test failure

- [ ] T001 Read BCM category schema documentation at `/workspace/sampleRest/category_schema_documentation_20251121_070629.md` (line 113) to extract valid disksetup XML example
- [ ] T002 Verify current test failure by running `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations` to confirm BCM validation error

**Checkpoint**: Valid XML structure documented, test confirmed failing with BCM validation error

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Research and document the valid XML structure

**CRITICAL**: This phase provides the foundation for the fix

- [ ] T003 Extract minimal valid disk setup XML from BCM documentation with proper structure (XML declaration, diskSetup root, device element, blockdev, partition definitions)
- [ ] T004 Document XML schema requirements: element names (case-sensitive), required children, attribute requirements based on BCM's XSD validation

**Checkpoint**: Valid minimal XML structure identified and ready to use in test

---

## Phase 3: User Story 1 - Valid XML Test (Priority: P1) 🎯 MVP

**Goal**: Fix the test by replacing invalid XML with valid BCM disk setup XML that conforms to the BCM XSD schema

**Independent Test**: Run `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations` and verify test passes without validation errors

### Implementation for User Story 1

- [ ] T005 [US1] RED - Confirm test is currently failing by running acceptance test and capturing BCM validation error message
- [ ] T006 [US1] GREEN - Replace invalid XML in `/workspace/internal/provider/resource_cmdevice_category_test.go` line 1277 with valid minimal disk setup XML structure
- [ ] T007 [US1] GREEN - Update state check expectation in `/workspace/internal/provider/resource_cmdevice_category_test.go` line 1282 to match the new valid XML
- [ ] T008 [US1] GREEN - Run acceptance test to verify test passes: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations`
- [ ] T009 [US1] REFACTOR - Add inline documentation comments in `/workspace/internal/provider/resource_cmdevice_category_test.go` above the testAccCMDeviceCategoryResourceConfig_DiskSetupOnly function explaining XML structure requirements
- [ ] T010 [US1] REFACTOR - Add XML structure comments: XML declaration required, diskSetup (capital S) root element, device/blockdev/partition hierarchy, required child elements
- [ ] T011 [US1] Verify test passes consistently by running 3 times to ensure no flakiness

**Valid XML Structure to Use** (from plan.md):

```xml
<?xml version="1.0" encoding="UTF-8"?>

<diskSetup>
  <device>
    <blockdev>/dev/sda</blockdev>
    <partition id="a0" partitiontype="esp">
      <size>100M</size>
      <type>linux</type>
      <filesystem>fat</filesystem>
      <mountPoint>/boot/efi</mountPoint>
      <mountOptions>defaults,noatime</mountOptions>
    </partition>
    <partition id="a1">
      <size>max</size>
      <type>linux</type>
      <filesystem>xfs</filesystem>
      <mountPoint>/</mountPoint>
      <mountOptions>defaults,noatime</mountOptions>
    </partition>
  </device>
</diskSetup>
```

**Checkpoint**: Test passes reliably, BCM accepts XML without validation errors, test verifies disk setup field works independently

---

## Phase 4: User Story 2 - Negative Validation Test (Priority: P2)

**Goal**: Add test to verify provider correctly handles BCM's XML validation errors when invalid disk setup XML is provided

**Independent Test**: Run new test and verify it correctly expects and catches BCM validation error

### Implementation for User Story 2

- [ ] T012 [US2] RED - Create new test function `TestAccCMDeviceCategoryResource_DiskSetupInvalidXML` in `/workspace/internal/provider/resource_cmdevice_category_test.go`
- [ ] T013 [US2] RED - Add test step with invalid XML (e.g., `<disksetup><disk/></disksetup>`) using ExpectError to catch validation message
- [ ] T014 [US2] GREEN - Run new test to verify it passes: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_DiskSetupInvalidXML`
- [ ] T015 [US2] REFACTOR - Add comments explaining that this test verifies BCM's server-side XML validation error handling

**Checkpoint**: Negative validation test ensures provider properly reports BCM validation errors to users

---

## Phase 5: User Story 3 - Documentation and Examples (Priority: P3)

**Goal**: Document valid disk setup XML formats to prevent future validation issues

**Independent Test**: Review test file and verify documentation is clear and complete

### Implementation for User Story 3

- [ ] T016 [US3] Add detailed comment block above `testAccCMDeviceCategoryResourceConfig_DiskSetupOnly` function in `/workspace/internal/provider/resource_cmdevice_category_test.go` documenting XML structure
- [ ] T017 [US3] Document required XML elements: XML declaration, diskSetup root (capital S), device, blockdev (device path), partition elements
- [ ] T018 [US3] Document required partition child elements: size, type, filesystem, mountPoint, and optional mountOptions
- [ ] T019 [US3] Add reference comment pointing to BCM category schema documentation location: `/workspace/sampleRest/category_schema_documentation_20251121_070629.md`

**Checkpoint**: Clear documentation helps developers understand XML requirements and prevents future validation errors

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Final verification and regression testing

- [ ] T020 Run all category acceptance tests to ensure no regressions: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategory`
- [ ] T021 Verify CI/CD pipeline would pass by running full test suite: `make testacc` (or subset if time constrained)
- [ ] T022 Review final changes and ensure code quality (formatting, comments, test clarity)

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories
- **User Story 1 (Phase 3)**: Depends on Foundational completion - MVP fix
- **User Story 2 (Phase 4)**: Independent of US1 - can start after Foundational
- **User Story 3 (Phase 5)**: Can start anytime, enhances US1
- **Polish (Phase 6)**: Depends on desired user stories being complete

### User Story Dependencies

- **User Story 1 (P1)**: MVP - Can start after Foundational (Phase 2) - No dependencies on other stories
- **User Story 2 (P2)**: Can start after Foundational (Phase 2) - Independent test, can be done in parallel with US1
- **User Story 3 (P3)**: Documentation - Can be done anytime, enhances US1

### Within Each User Story

**User Story 1 (TDD Workflow)**:
1. T005: RED - Confirm failure
2. T006-T007: GREEN - Fix test data and expectations
3. T008: GREEN - Verify test passes
4. T009-T010: REFACTOR - Add documentation
5. T011: Verify - Ensure reliability

**User Story 2**:
1. RED: Create new test expecting error
2. GREEN: Verify test passes
3. REFACTOR: Add comments

**User Story 3**:
1. Add comprehensive documentation

### Parallel Opportunities

- **Phase 2**: T003 and T004 can be done together (research and document in single pass)
- **User Stories**: US2 can be worked on in parallel with US1 after Foundational complete
- **User Story 3**: Can be done in parallel with US1 or US2

---

## Parallel Example: After Foundational Phase

```bash
# User Story 1 (P1) - Primary fix
Task: "Replace invalid XML with valid structure"
Task: "Run test to verify pass"

# User Story 2 (P2) - Can run in parallel
Task: "Create negative validation test"

# User Story 3 (P3) - Can run in parallel
Task: "Add documentation comments"
```

---

## Implementation Strategy

### MVP First (User Story 1 Only)

1. Complete Phase 1: Setup (extract valid XML, verify failure)
2. Complete Phase 2: Foundational (document XML requirements)
3. Complete Phase 3: User Story 1 (fix test with valid XML)
4. **STOP and VALIDATE**: Run test 3 times to ensure reliability
5. Commit fix, closes issue #56

### Incremental Delivery

1. Complete Setup + Foundational → XML structure documented
2. Add User Story 1 → Test passes → MVP Complete! ✅
3. Add User Story 2 → Negative validation covered → Enhanced error testing
4. Add User Story 3 → Documentation complete → Future-proof
5. Each story adds value without breaking previous work

### Single Developer Strategy

Sequential priority order (recommended):

1. Phases 1-2: Setup + Foundational (15 minutes)
2. Phase 3: User Story 1 (30 minutes) → MVP DONE
3. Phase 4: User Story 2 (20 minutes) → Optional enhancement
4. Phase 5: User Story 3 (15 minutes) → Documentation
5. Phase 6: Polish (10 minutes) → Regression check

**Total estimated time**: ~90 minutes for complete implementation

---

## Notes

- Invalid XML causing test failure: `<disksetup><disk/></disksetup>` (wrong element names, missing required children)
- Valid XML must have: XML declaration, `diskSetup` (capital S), `device`, `blockdev`, `partition` elements with children
- BCM validates XML server-side against XSD schema - provider cannot pre-validate
- Test must remain environment-portable (uses existing software images and networks via data sources)
- Element names are case-sensitive: `diskSetup` not `disksetup`, `blockdev` not `blockDev`
- XML must include all required partition child elements: size, type, filesystem, mountPoint

**Critical File**: `/workspace/internal/provider/resource_cmdevice_category_test.go`
- Line 1277: Invalid XML passed to test config helper
- Line 1282: State check expectation must match new valid XML
- Lines 1392-1426: `testAccCMDeviceCategoryResourceConfig_DiskSetupOnly` function to modify

**Reference Documentation**: `/workspace/sampleRest/category_schema_documentation_20251121_070629.md` (line 113)

**Test Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations
```

**Success Criteria**:
- SC-001: Test passes 100% of the time ✅
- SC-002: Test completes in under 2 minutes ✅
- SC-003: Zero BCM validation errors ✅
- SC-004: CI/CD shows green status ✅
