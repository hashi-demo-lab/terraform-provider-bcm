# Feature Specification: Fix Disk Setup XML Validation Test

**Feature Branch**: `056-fix-disksetup-xml-test`
**Created**: 2025-11-25
**Status**: Draft
**Input**: User description: "Fix disk setup XML validation test issue #56"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Valid XML Test (Priority: P1)

Developers running acceptance tests for the `bcm_cmdevice_category` resource need the test `TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations` to pass when using valid BCM disk setup XML that conforms to the BCM XSD schema.

**Why this priority**: This is the MVP - the test currently fails with invalid XML, blocking CI/CD and provider development. Fixing this enables developers to validate disk setup functionality.

**Independent Test**: Can be fully tested by running `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations` and verifying it passes with valid XML.

**Acceptance Scenarios**:

1. **Given** a test configuration with valid BCM disk setup XML, **When** the acceptance test runs, **Then** the test passes without validation errors
2. **Given** the test creates a category with disk setup XML, **When** BCM validates the XML, **Then** BCM accepts the XML without schema errors
3. **Given** valid disk setup XML from BCM documentation, **When** used in the test, **Then** the category is created successfully

---

### User Story 2 - Negative Validation Test (Priority: P2)

Developers need to verify that the provider correctly handles BCM's XML validation errors when invalid disk setup XML is provided, ensuring proper error reporting to users.

**Why this priority**: While not blocking test passes, this ensures the provider properly handles and reports BCM validation errors to users, improving user experience when they provide invalid XML.

**Independent Test**: Can be tested independently by creating a separate test with invalid XML and verifying ExpectError catches the validation message.

**Acceptance Scenarios**:

1. **Given** a test with invalid disk setup XML (like `<disksetup><disk/></disksetup>`), **When** BCM validates the XML, **Then** the test expects and catches the validation error
2. **Given** malformed XML missing required elements, **When** the provider attempts to create the category, **Then** the error message clearly indicates the XML validation failure
3. **Given** XML that doesn't conform to BCM's XSD schema, **When** validation occurs, **Then** BCM's detailed validation error is surfaced to the user

---

### User Story 3 - Documentation and Examples (Priority: P3)

Developers and users need clear documentation and examples showing valid disk setup XML formats that work with BCM, preventing future XML validation issues.

**Why this priority**: Long-term value - prevents users from encountering validation errors by providing working examples upfront.

**Independent Test**: Can be verified by reviewing test helper functions and ensuring they include comments with valid XML structure examples.

**Acceptance Scenarios**:

1. **Given** the test file contains disk setup XML, **When** developers review the code, **Then** comments explain the XML structure and requirements
2. **Given** valid XML examples from BCM documentation, **When** included in test helpers, **Then** developers can reference these for their own configurations
3. **Given** the category schema documentation, **When** developers look up disk setup, **Then** they find the complete valid XML example structure

---

### Edge Cases

- What happens when disk setup XML is empty string versus null?
- How does BCM handle XML with valid structure but invalid device paths?
- What validation errors occur when partition sizes don't match BCM's expectations?
- How does the test behave when BCM's XSD schema requirements change in future versions?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The test MUST use valid BCM disk setup XML that conforms to BCM's XSD schema
- **FR-002**: The test MUST pass when run in the acceptance test suite without validation errors
- **FR-003**: The disk setup XML MUST include proper XML declaration (`<?xml version="1.0" encoding="UTF-8"?>`)
- **FR-004**: The disk setup XML MUST use the correct root element structure (`<diskSetup>` with proper casing)
- **FR-005**: The disk setup XML MUST include valid partition definitions with required attributes (size, type, filesystem, mountPoint)
- **FR-006**: The test MUST verify that categories can be created with disk setup XML independently of raidconf and install_boot_record
- **FR-007**: Invalid XML test cases MUST properly use ExpectError to verify validation error handling
- **FR-008**: The test helper functions MUST use realistic disk setup XML matching BCM's default category structure

### Key Entities *(include if feature involves data)*

- **Disk Setup XML**: A string containing valid XML conforming to BCM's diskSetup XSD schema, defining partition layout and filesystem configuration
  - Root element: `<diskSetup>` (note: capital S)
  - Device element: `<device>` containing block device specifications
  - Blockdev elements: `<blockdev>` specifying device paths (e.g., `/dev/sda`)
  - Partition elements: `<partition>` with attributes (id, partitiontype) and child elements (size, type, filesystem, mountPoint, mountOptions)

- **BCM Category**: The resource being tested, which can optionally include disk setup configuration
  - Attributes: name, management_network, disksetup (optional), raidconf (optional), install_boot_record (optional)
  - Validation: BCM validates disksetup XML against XSD schema during category creation/update

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: The test `TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations` passes 100% of the time when using valid XML
- **SC-002**: The test completes in under 2 minutes (standard acceptance test timeout)
- **SC-003**: BCM validation errors are eliminated - zero "disksetup/xml" or "disksetup/xsd" errors occur
- **SC-004**: CI/CD pipeline shows green status for disk setup optional combinations test

## Scope *(mandatory)*

### In Scope

- Updating the test with valid BCM disk setup XML from documented examples
- Ensuring XML conforms to BCM's XSD schema requirements
- Verifying the test passes with proper partition configuration
- Adding inline comments documenting the XML structure
- Option to convert to negative test if valid XML examples are insufficient

### Out of Scope

- Implementing BCM XSD schema validation in the provider itself (BCM handles this)
- Creating comprehensive documentation of all possible disk setup configurations
- Testing every possible disk setup XML variation
- Modifying BCM's XSD schema or validation behavior
- Testing the comprehensive `TestAccCMDeviceCategoryResource_DiskSetupAdvanced` test (already skipped pending XSD documentation)

## Dependencies & Constraints *(mandatory)*

### Dependencies

- BCM API endpoint must be available for acceptance testing
- BCM must have consistent XSD schema validation for disk setup XML
- Existing category schema documentation (`sampleRest/category_schema_documentation_20251121_070629.md`) provides valid XML example
- Test environment must have at least one existing software image for test configuration

### Constraints

- XML must match BCM's exact schema expectations (case-sensitive element names)
- BCM validation occurs server-side - provider cannot pre-validate XML structure
- Test must remain portable across different BCM cluster configurations
- XML structure may vary between BCM versions (current example is from BCM 10.1)

## Assumptions *(mandatory)*

- The valid XML example from BCM's default category is representative and will work in tests
- BCM's XSD schema for disk setup is stable and won't reject the documented XML structure
- The test failure is solely due to invalid XML, not other test configuration issues
- BCM requires the XML declaration header for proper validation
- Element names are case-sensitive (diskSetup not disksetup, blockdev not blockDev)

## Non-Functional Requirements *(if applicable)*

### Performance
- Test execution time should not increase by more than 10% when using valid XML

### Reliability
- Test should have 100% pass rate when BCM API is available
- Test should fail gracefully with clear error messages if BCM is unavailable

### Maintainability
- XML examples should be extracted to constants or helper functions for reusability
- Comments should document XML structure requirements for future developers

## Open Questions *(if any)*

None - the valid XML structure is documented in BCM's default category configuration.

## References

- Issue: https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/56 (implied)
- BCM Category Schema: `/workspace/sampleRest/category_schema_documentation_20251121_070629.md` (line 113)
- Test File: `/workspace/internal/provider/resource_cmdevice_category_test.go` (lines 1264-1310)
- Valid XML Example: Default category's disksetup field shows proper structure with `<diskSetup>`, `<device>`, `<blockdev>`, and `<partition>` elements

## Validation Strategy

The specification will be validated through:

1. **Test Execution**: Run the updated test and verify it passes
2. **BCM Validation**: Confirm BCM accepts the XML without schema errors
3. **Code Review**: Ensure XML structure matches BCM documentation
4. **CI/CD**: Verify test passes in automated pipeline
