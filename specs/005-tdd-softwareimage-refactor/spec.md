# Feature Specification: Complete TDD-Based Review and Refactoring of resource_cmpart_softwareimage

**Feature Branch**: `005-tdd-softwareimage-refactor`
**Created**: 2025-11-21
**Status**: Draft
**Input**: User description: "Complete TDD-based review and refactoring of resource_cmpart_softwareimage following Option A: Full TDD Implementation (Recommended by analysis) with complete RED-GREEN-REFACTOR cycles for all user stories ensuring 100% test coverage and TDD discipline"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create Software Image by Cloning (Priority: P1)

As a cluster administrator, I need to create a new software image by cloning an existing base image so that I can deploy standardized OS configurations to compute nodes with custom kernel settings and modules.

**Why this priority**: This is the foundational operation that enables all software image management. Without the ability to clone and create images, no provisioning workflows are possible. This represents the core value proposition of the resource.

**Independent Test**: Can be fully tested by creating a Terraform configuration that clones "default-image" to a new image name/path, applies it, and verifies the new image exists in BCM with the expected attributes. Delivers immediate value by enabling image creation workflows.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with a "default-image" base image, **When** I create a Terraform resource with `name`, `path`, and `original_image` set to the default image UUID, **Then** BCM creates a new software image with filesystem cloned from the base image, assigns a unique UUID, and returns all computed fields (creation_time, uuid, revision_id)

2. **Given** a newly created software image from cloning, **When** the clone operation completes (fileOperationInProgress = false), **Then** the image inherits kernel_version and modules from the original image and has filesystem partitions (fspart, bootfspart) populated

3. **Given** a created software image, **When** I run `terraform import bcm_cmpart_softwareimage.test <uuid>`, **Then** Terraform successfully imports the resource and populates all attributes from BCM API

---

### User Story 2 - Read and Verify Image State (Priority: P1)

As a cluster administrator, I need to read the current state of a software image from BCM so that Terraform can detect drift, verify successful provisioning, and maintain accurate state.

**Why this priority**: Read operations are critical for Terraform's core functionality. Without accurate read operations, plan/apply/refresh cycles cannot function correctly, drift detection fails, and import doesn't work. This is equally important as Create for basic TDD cycles.

**Independent Test**: Can be fully tested by creating an image directly via BCM API, then using Terraform data source or resource import to read it back and verify all attributes match. Delivers value by enabling state verification and drift detection.

**Acceptance Scenarios**:

1. **Given** a software image exists in BCM with UUID "abc-123", **When** I call the Read operation with that UUID, **Then** Terraform retrieves all attributes (name, path, kernel settings, SOL settings, modules, metadata) and maps them to the Terraform state

2. **Given** an image was created with `original_image` set, **When** the clone completes and I read the image, **Then** the original_image attribute is preserved in state (BCM resets it to zero UUID after cloning, but Terraform should maintain the original value for audit purposes)

3. **Given** an image with 3 kernel modules, **When** I read the image state, **Then** the modules list contains all 3 modules with correct name and parameters attributes, and the list is never Unknown or null

4. **Given** a software image that doesn't exist, **When** I attempt to read it, **Then** Terraform returns an appropriate error diagnostic without crashing

---

### User Story 3 - Update Kernel Configuration (Priority: P2)

As a cluster administrator, I need to update kernel parameters, console settings, and module lists on existing software images so that I can tune boot behavior and hardware support without recreating images.

**Why this priority**: After image creation, administrators need to iteratively refine kernel configuration based on hardware requirements and testing. This enables in-place updates without destroying/recreating images, which is critical for production workflows where image UUIDs may be referenced elsewhere.

**Independent Test**: Can be fully tested by creating an image, then updating kernel_parameters from empty to "quiet splash", applying the change, and verifying BCM API reflects the update. Delivers value by enabling non-destructive configuration changes.

**Acceptance Scenarios**:

1. **Given** an existing software image with `kernel_parameters = ""`, **When** I update the Terraform config to set `kernel_parameters = "quiet splash nomodeset"` and apply, **Then** BCM updates the image and subsequent reads return the new kernel parameters

2. **Given** an image with 2 kernel modules, **When** I add a third module to the Terraform config and apply, **Then** BCM API updates the modules list to include all 3 modules

3. **Given** an image with 3 kernel modules, **When** I remove one module from the Terraform config and apply, **Then** BCM API updates the modules list to only contain the remaining 2 modules

4. **Given** an image with default SOL settings, **When** I update enable_sol, sol_speed, and sol_port attributes and apply, **Then** BCM API updates the SOL configuration and subsequent reads return the new values

---

### User Story 4 - Delete Software Image (Priority: P2)

As a cluster administrator, I need to delete software images that are no longer needed so that I can clean up unused resources and free storage on the BCM cluster.

**Why this priority**: Resource cleanup is essential for resource lifecycle management and cost control. While less critical than Create/Read/Update, Delete completes the CRUD cycle and is required for Terraform's destroy operations and test cleanup.

**Independent Test**: Can be fully tested by creating an image via Terraform, then running `terraform destroy` and verifying the image no longer exists in BCM API. Delivers value by enabling proper resource cleanup in both production and testing scenarios.

**Acceptance Scenarios**:

1. **Given** a software image managed by Terraform, **When** I run `terraform destroy`, **Then** Terraform calls BCM's removeSoftwareImage API with the UUID and appropriate flags (removeData=false, removeAll=false, force=false)

2. **Given** a successfully deleted image, **When** I attempt to read that image via BCM API, **Then** the API returns empty response or error indicating the image doesn't exist

3. **Given** an image that is in use by provisioned nodes, **When** I attempt to delete it via Terraform, **Then** BCM API returns an error and Terraform propagates the diagnostic to the user (this validates proper error handling)

---

### User Story 5 - Handle Async Clone Operations (Priority: P2)

As a cluster administrator, when I create a software image by cloning, I need Terraform to wait for the clone operation to complete before marking the resource as successfully created so that subsequent operations (like kernel parameter updates) don't fail due to incomplete filesystem setup.

**Why this priority**: BCM's clone operation is asynchronous - the addSoftwareImage API returns immediately but filesystem copying continues in the background. Without proper polling, Terraform may try to update kernel files before they exist, causing failures. This is critical for reliability but can be handled in the Create operation.

**Independent Test**: Can be fully tested by creating an image with original_image set, monitoring BCM's fileOperationInProgress field, and verifying Terraform waits until it becomes false before returning success. Delivers value by ensuring reliable clone operations without race conditions.

**Acceptance Scenarios**:

1. **Given** I create a new image with original_image set, **When** Terraform calls BCM's addSoftwareImage, **Then** Terraform polls the fileOperationInProgress field with exponential backoff (1s, 2s, 4s, 8s, 16s) until it becomes false

2. **Given** a clone operation that completes in 5 seconds, **When** Terraform's polling detects fileOperationInProgress=false, **Then** Terraform proceeds to read back the final state and completes the create operation successfully

3. **Given** a clone operation that exceeds maximum polling time, **When** Terraform exhausts retries, **Then** Terraform logs a warning but proceeds (allowing manual verification), rather than hard-failing the create

---

### User Story 6 - Validate Input Attributes (Priority: P3)

As a cluster administrator, when I provide invalid input values (missing required fields, invalid SOL speed, malformed path), I need Terraform to reject the configuration during plan phase so that I can correct errors before attempting to apply changes.

**Why this priority**: Client-side validation provides fast feedback and better user experience compared to waiting for API errors. However, it's lower priority because BCM API also validates inputs, so this is an enhancement rather than a requirement.

**Independent Test**: Can be fully tested by creating Terraform configs with invalid values (e.g., sol_speed="9999") and verifying plan phase fails with appropriate error messages. Delivers value by improving user experience and reducing wasted API calls.

**Acceptance Scenarios**:

1. **Given** a Terraform config missing the required `name` attribute, **When** I run `terraform plan`, **Then** Terraform returns an error diagnostic indicating `argument "name" is required` before making any API calls

2. **Given** a Terraform config with `sol_speed = "9999"` (invalid value), **When** I run `terraform plan`, **Then** Terraform returns an error diagnostic indicating the value must match the OneOf validator (115200, 57600, 38400, 19200, 9600, 4800, 2400, 1200)

3. **Given** a Terraform config with `path = "invalid path with spaces"`, **When** I run `terraform plan`, **Then** Terraform returns an error diagnostic indicating the path must match the regex pattern `^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$`

---

### User Story 7 - Handle Unknown Values Correctly (Priority: P3)

As a Terraform provider developer, I need to ensure that Unknown values during plan phase are never propagated to state after apply, so that users don't encounter "invalid result object" errors when working with computed attributes like modules and original_image.

**Why this priority**: This is a technical correctness requirement that prevents a specific class of Terraform errors. While critical for quality, it's lower priority because it manifests as a bug fix rather than a user-facing feature. Most users won't encounter this unless they hit specific edge cases.

**Independent Test**: Can be fully tested by creating a configuration where modules or original_image are computed, running plan/apply, and verifying state never contains Unknown values. Delivers value by ensuring Terraform compliance and preventing cryptic errors.

**Acceptance Scenarios**:

1. **Given** a Terraform config with original_image referencing a data source (Unknown during plan), **When** I run `terraform apply`, **Then** the final state contains either a concrete UUID value or null, never Unknown

2. **Given** a software image with no modules in BCM API, **When** I read the image, **Then** the modules attribute is set to an empty list (known value) rather than Unknown or null

3. **Given** a cloned image where BCM resets original_image to zero UUID, **When** I read the image during Create operation, **Then** Terraform preserves the plan's original_image value in state (as long as it was a known value in the plan), rather than overwriting it with the zero UUID or Unknown

---

### Edge Cases

- **What happens when a user tries to clone from a non-existent original_image UUID?** BCM API should return an error during addSoftwareImage, which Terraform should catch and return as a diagnostic (test: T046d - negative test for invalid original_image)

- **What happens when two Terraform applies try to create images with the same name simultaneously?** BCM API enforces name uniqueness, so the second create should fail with a conflict error (test: verify error handling with duplicate names)

- **What happens when a user updates an image that was deleted outside of Terraform (drift)?** The Update operation should fail with "not found" error, and user should run `terraform refresh` to detect the drift

- **What happens when kernel_version is set during initial create with original_image?** BCM API may reject this because kernel files are cloned from original image. The current tests use a two-step approach: create with basic config, then update with kernel params (this is an API limitation, not a Terraform bug)

- **What happens when fileOperationInProgress remains true beyond max polling time?** Terraform logs a warning but proceeds, allowing users to verify manually rather than hard-failing (this handles slow storage or large images)

- **What happens when modules list contains a module with empty parameters?** The parameters field should be set to empty string "" rather than null, as BCM API expects string type

- **What happens when path contains special characters like spaces or @revision?** The regex validator `^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$` should reject invalid characters during plan phase

## Requirements *(mandatory)*

### Functional Requirements

#### Core CRUD Operations

- **FR-001**: Resource MUST support Create operation that calls BCM's `addSoftwareImage` API with a properly constructed entity including baseType, childType, name, path, and all optional fields

- **FR-002**: Resource MUST support Read operation that calls BCM's `getSoftwareImage(name)` API with direct name lookup (not list+filter pattern) and maps all API fields to Terraform schema attributes

- **FR-003**: Resource MUST support Update operation that calls BCM's `updateSoftwareImage` API with UUID included in the entity and properly handles all updatable fields

- **FR-004**: Resource MUST support Delete operation that calls BCM's `removeSoftwareImage(uuid, false, false, false)` API with appropriate flags for standard deletion

- **FR-005**: Resource MUST support ImportState operation using UUID as the import identifier and successfully populate all attributes from BCM API

#### Schema and Validation

- **FR-006**: Resource schema MUST define `name` and `path` as required attributes with appropriate validators (length check for name, regex check for path)

- **FR-007**: Resource schema MUST define optional attributes for kernel configuration (kernel_version, kernel_parameters, kernel_output_console) with appropriate defaults

- **FR-008**: Resource schema MUST define optional SOL configuration attributes (enable_sol, sol_port, sol_speed, sol_flow_control) with appropriate defaults and validators (OneOf for sol_speed)

- **FR-009**: Resource schema MUST define modules as a list of nested objects with name (required) and parameters (optional) attributes

- **FR-010**: Resource schema MUST define computed-only attributes (uuid, id, creation_time, revision_id, file_operation_in_progress, fspart, bootfspart, parent_software_image)

- **FR-011**: Resource schema MUST define original_image as optional + computed with UseStateForUnknown plan modifier to support cloning workflows

#### State Management

- **FR-012**: Resource MUST never propagate Unknown values to state - all attributes must resolve to known values (concrete value or null) after apply completes

- **FR-013**: Resource MUST preserve original_image value from plan in state even when BCM API resets it to zero UUID after cloning (this maintains audit trail of clone source)

- **FR-014**: Resource MUST set modules to an empty list (not null, not Unknown) when BCM API returns no modules or missing modules field

- **FR-015**: Resource ID attribute MUST be set equal to UUID for import compatibility

#### Async Operation Handling

- **FR-016**: During Create with original_image set, resource MUST poll BCM's fileOperationInProgress field with exponential backoff (1s, 2s, 4s, 8s, 16s) until clone completes

- **FR-017**: Resource MUST log detailed trace messages during clone polling including attempt number, wait duration, and fileOperationInProgress status

- **FR-018**: If clone polling exceeds maximum retries, resource MUST log a warning but proceed with reading final state rather than hard-failing

#### API Entity Construction

- **FR-019**: buildAPIEntity helper MUST construct BCM entities with required fields: baseType="SoftwareImage", childType="", modified=true, to_be_removed=false, revision=""

- **FR-020**: buildAPIEntity MUST include UUID field only for update operations, not for create operations

- **FR-021**: buildAPIEntity MUST include original_image field only during create operations when set, not during update operations

- **FR-022**: buildAPIEntity MUST construct modules as array of objects with baseType="KernelModule", childType="", modified=true, and module-specific fields

- **FR-023**: buildAPIEntity MUST set module parameters to empty string "" when null/not provided (BCM API expects string type)

#### Error Handling

- **FR-024**: Resource MUST return appropriate error diagnostics when BCM API returns errors, including error message from API response

- **FR-025**: Resource MUST handle "not found" errors during Read by returning appropriate diagnostic without crashing

- **FR-026**: Resource MUST validate API responses (check for null/empty response) before attempting to unmarshal

- **FR-027**: Resource MUST handle UUID extraction from multiple response formats: direct string, object with uuid field, object with updated_entity.uuid field

#### Test Requirements

- **FR-028**: Acceptance tests MUST include provider configuration block with endpoint, username, password, and insecure_skip_verify=true

- **FR-029**: Acceptance tests MUST use unique resource names with timestamp suffix to prevent collisions from parallel test runs

- **FR-030**: Acceptance tests MUST include PreCheck function that cleans up leftover test images from previous runs

- **FR-031**: Acceptance tests MUST include CheckDestroy function that verifies all test resources are deleted after test completes

- **FR-032**: Acceptance tests MUST test full CRUD cycle: Create+Read, Import, Update+Read, Delete for each user story

- **FR-033**: Acceptance tests MUST include negative tests for schema validation (missing required fields, invalid SOL speed, invalid path format)

- **FR-034**: Each acceptance test MUST use data source lookup for default-image UUID rather than hardcoded values (portability)

#### TDD Requirements

- **FR-035**: Implementation MUST follow RED-GREEN-REFACTOR cycle: write failing test first, implement minimal code to pass, then refactor while keeping tests green

- **FR-036**: All tests MUST be written before implementation code for that feature (strict TDD discipline)

- **FR-037**: Test coverage MUST include all CRUD operations, all schema attributes, all validation rules, and all error paths

- **FR-038**: Each refactoring step MUST maintain all tests in passing state (no broken tests during refactor phase)

### Key Entities *(this feature involves Terraform resource state and BCM API entities)*

- **Software Image Resource**: Represents a BCM software image (OS kernel + filesystem) managed by Terraform. Key attributes include name (unique identifier), path (filesystem location), uuid (BCM-assigned ID), kernel configuration (version, parameters, console, modules), SOL configuration (serial console settings), metadata (creation time, revision, file operation status), and clone source (original_image for audit trail).

- **Kernel Module**: Nested object within Software Image representing a kernel module to load at boot. Attributes include name (module name like "nvidia-drm") and parameters (module options like "modeset=1").

- **BCM API Entity**: The JSON object structure expected by BCM's addSoftwareImage/updateSoftwareImage APIs. Required fields: baseType="SoftwareImage", childType="", modified=true, to_be_removed=false, revision="". Optional fields include all resource attributes plus UUID for updates.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All CRUD operations (Create, Read, Update, Delete) complete successfully in acceptance tests with real BCM API endpoint in under 2 minutes per test

- **SC-002**: Test coverage reaches 100% for all user stories - each acceptance test corresponds to one or more acceptance scenarios and all scenarios have passing tests

- **SC-003**: Zero test failures in final test suite - all 11+ acceptance tests pass consistently across multiple runs

- **SC-004**: Import functionality works for 100% of created resources - every resource created in tests can be successfully imported using its UUID

- **SC-005**: State drift detection works correctly - manual changes to images in BCM API are correctly detected by Terraform refresh operations

- **SC-006**: Clone operations complete successfully with 100% reliability - fileOperationInProgress polling handles both fast (<5s) and slow (>10s) clone operations

- **SC-007**: Schema validation catches 100% of invalid inputs during plan phase - missing required fields, invalid SOL speeds, and malformed paths all produce appropriate error diagnostics before API calls

- **SC-008**: Unknown value handling is 100% compliant - zero "invalid result object" errors occur during plan/apply cycles with computed attributes

- **SC-009**: Parallel test execution completes without conflicts - unique timestamp-based naming prevents collisions when running tests concurrently

- **SC-010**: TDD discipline maintained throughout - all implementation code has corresponding test written first, RED-GREEN-REFACTOR cycle documented for each feature

### Implementation Quality Metrics

- **SC-011**: Code follows HashiCorp Terraform Plugin Framework best practices - proper use of schema validators, plan modifiers, diagnostics, and framework types

- **SC-012**: Documentation is auto-generated and accurate - `make generate` produces correct docs that match resource schema and examples

- **SC-013**: Error messages are actionable - all error diagnostics include sufficient context for users to understand and resolve issues

- **SC-014**: Logging is comprehensive - all CRUD operations log key events at appropriate levels (Trace for details, Debug for operations, Info for lifecycle events, Warn for issues)

- **SC-015**: Code is maintainable - helper functions (buildAPIEntity, readSoftwareImage) are well-structured, single-purpose, and reusable
