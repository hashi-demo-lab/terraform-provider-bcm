# Feature Specification: Category Test Coverage Enhancement

**Feature Branch**: `001-category-test-coverage`
**Created**: 2025-11-25
**Status**: Draft
**Input**: User description: "GitHub issue #66: Test coverage: Add tests for untested optional fields in bcm_cmdevice_category"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Installation Mode Field Testing (Priority: P1)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests that verify the `install_mode` and `new_node_install_mode` fields function correctly across Create, Read, Update, and Delete operations so that users can reliably configure node installation behavior.

**Why this priority**: Installation modes are the most commonly used fields for node provisioning workflows. Users frequently configure AUTO, FULL, GRAB modes to control how nodes are installed. These are high-priority because incorrect behavior could result in failed node deployments.

**Independent Test**: Can be fully tested by creating a category with install_mode="AUTO" and new_node_install_mode="FULL", updating to install_mode="FULL", and verifying state is preserved across terraform plan cycles.

**Acceptance Scenarios**:

1. **Given** a category configuration with install_mode="AUTO" and new_node_install_mode="FULL", **When** terraform apply is executed, **Then** the category is created with the specified installation modes and subsequent plans show no changes.

2. **Given** an existing category with install_mode="AUTO", **When** the configuration is updated to install_mode="FULL", **Then** the update succeeds and the new value is reflected in state.

3. **Given** a category with installation modes configured, **When** terraform import is executed, **Then** the installation mode values are correctly imported into state.

---

### User Story 2 - Network Settings Field Testing (Priority: P1)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests for `name_servers`, `search_domains`, and `time_servers` list fields to ensure DNS and NTP configuration works correctly for infrastructure provisioning.

**Why this priority**: Network settings are critical infrastructure configuration. DNS and NTP settings must work correctly for nodes to communicate and synchronize properly. These fields are commonly configured in production environments.

**Independent Test**: Can be fully tested by creating a category with name_servers=["8.8.8.8", "8.8.4.4"], search_domains=["example.com"], time_servers=["ntp.example.com"], updating the lists, and verifying list handling behavior.

**Acceptance Scenarios**:

1. **Given** a category configuration with name_servers, search_domains, and time_servers lists, **When** terraform apply is executed, **Then** all list values are correctly stored and returned on subsequent reads.

2. **Given** an existing category with network settings, **When** a server is added to or removed from the name_servers list, **Then** the update succeeds and the modified list is reflected in state.

3. **Given** a category with empty network settings lists, **When** terraform apply is executed, **Then** the category is created without errors and null/empty list handling works correctly.

---

### User Story 3 - I/O Scheduler and Kernel Fields Testing (Priority: P2)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests for `io_scheduler` and `kernel_version` fields to ensure kernel-level configuration is properly managed.

**Why this priority**: I/O scheduler and kernel version are important for performance tuning in HPC environments. While not as commonly configured as installation modes, incorrect handling could impact node performance.

**Independent Test**: Can be fully tested by creating a category with io_scheduler="mq-deadline", verifying the value persists, and testing update to io_scheduler="none".

**Acceptance Scenarios**:

1. **Given** a category configuration with io_scheduler="mq-deadline", **When** terraform apply is executed, **Then** the I/O scheduler value is correctly stored and persisted.

2. **Given** an existing category with io_scheduler configured, **When** the configuration is updated to a different scheduler, **Then** the update succeeds and idempotency is maintained.

---

### User Story 4 - Provisioning Scripts Field Testing (Priority: P2)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests for `initialize` and `finalize` provisioning script fields to ensure custom provisioning workflows are supported.

**Why this priority**: Initialize and finalize scripts are commonly used for custom node setup. Users need confidence that multi-line scripts are handled correctly without corruption or truncation.

**Independent Test**: Can be fully tested by creating a category with initialize and finalize scripts containing multi-line bash content, updating the scripts, and verifying exact string preservation.

**Acceptance Scenarios**:

1. **Given** a category configuration with initialize="#!/bin/bash\necho 'setup'" and finalize="#!/bin/bash\necho 'done'", **When** terraform apply is executed, **Then** both scripts are stored exactly as provided including newlines and special characters.

2. **Given** an existing category with scripts, **When** the initialize script is updated, **Then** the update succeeds and the finalize script remains unchanged.

---

### User Story 5 - BMC Settings Nested Object Testing (Priority: P3)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests for the `bmc_settings` nested object to ensure BMC/IPMI configuration fields work correctly.

**Why this priority**: BMC configuration is important for hardware management but requires specific hardware support. Testing this field group ensures the nested object handling works correctly for users with BMC-capable hardware.

**Independent Test**: Can be fully tested by creating a category with bmc_settings containing user_name, privilege, and user_id fields, updating individual nested fields, and verifying object state management.

**Acceptance Scenarios**:

1. **Given** a category configuration with bmc_settings containing user_name="admin", privilege="ADMINISTRATOR", **When** terraform apply is executed, **Then** the nested object is correctly stored with all fields.

2. **Given** an existing category with bmc_settings, **When** an individual field like user_id is updated, **Then** only that field changes while others remain unchanged.

---

### User Story 6 - Kernel Modules Nested List Testing (Priority: P3)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests for the `modules` (kernel_modules) nested list attribute to verify kernel module configuration with parameters works correctly.

**Why this priority**: Kernel modules are important for hardware drivers and specialized workloads. The nested list structure with name and parameters fields requires specific handling that needs validation.

**Independent Test**: Can be fully tested by creating a category with modules=[{name="mlx5_core", parameters="debug=1"}], adding/removing modules, and verifying list ordering is preserved.

**Acceptance Scenarios**:

1. **Given** a category configuration with modules containing kernel module entries with name and parameters, **When** terraform apply is executed, **Then** the module list is correctly stored with all nested fields.

2. **Given** an existing category with modules, **When** a new module is added to the list, **Then** the update succeeds and existing modules remain unchanged.

---

### User Story 7 - Exclude Lists Field Testing (Priority: P3)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests for exclude list fields (`exclude_list_full`, `exclude_list_grab`, `exclude_list_sync`, etc.) to verify large text fields are handled correctly.

**Why this priority**: Exclude lists can contain large amounts of data (up to 50KB) with file paths and patterns. Testing ensures these fields handle multiline content without truncation.

**Independent Test**: Can be fully tested by creating a category with exclude_list_full containing multiple file patterns, updating the list, and verifying exact content preservation.

**Acceptance Scenarios**:

1. **Given** a category configuration with exclude_list_full containing multiple lines of file patterns, **When** terraform apply is executed, **Then** all patterns are stored exactly as provided.

2. **Given** an existing category with exclude lists, **When** exclude_list_sync is updated, **Then** only that field changes while other exclude lists remain unchanged.

---

### User Story 8 - Miscellaneous Boolean and String Fields Testing (Priority: P3)

As a developer maintaining the bcm_cmdevice_category resource, I need acceptance tests for remaining untested fields: `fips`, `data_node`, `interactive_user`, `use_exclusively_for`, `node_installer_disk`, and `version_config_files` to achieve comprehensive test coverage.

**Why this priority**: These fields are less commonly used but still need coverage for completeness. Testing ensures all schema attributes are properly handled.

**Independent Test**: Can be fully tested by creating a category with these optional fields set, updating values, and verifying boolean/string handling.

**Acceptance Scenarios**:

1. **Given** a category configuration with fips="YES", data_node=true, **When** terraform apply is executed, **Then** both values are correctly stored.

2. **Given** an existing category with miscellaneous fields, **When** fips is updated from "YES" to "NO", **Then** the update succeeds and idempotency is maintained.

---

### Edge Cases

- What happens when name_servers list is set to an empty list after having values? (List removal behavior)
- How does the system handle kernel_modules with duplicate module names?
- What happens when exclude_list fields exceed 50KB? (Size limit validation)
- How does the system handle special characters in script fields (initialize, finalize)?
- What happens when bmc_settings is set to null after being configured? (Nested object removal)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Test suite MUST verify `install_mode` and `new_node_install_mode` fields across CRUD operations
- **FR-002**: Test suite MUST verify `name_servers`, `search_domains`, and `time_servers` list fields handle list operations correctly
- **FR-003**: Test suite MUST verify `io_scheduler` field accepts and persists valid scheduler values
- **FR-004**: Test suite MUST verify `initialize` and `finalize` script fields preserve multi-line content exactly
- **FR-005**: Test suite MUST verify `bmc_settings` nested object handles all nested fields correctly
- **FR-006**: Test suite MUST verify `modules` (kernel_modules) nested list handles add/remove/update of module entries
- **FR-007**: Test suite MUST verify exclude list fields (`exclude_list_full`, `exclude_list_grab`, `exclude_list_sync`, `exclude_list_update`, `exclude_list_grabnew`, `exclude_list_manipulate_script`) handle large text content
- **FR-008**: Test suite MUST verify remaining fields (`fips`, `data_node`, `interactive_user`, `use_exclusively_for`, `node_installer_disk`, `version_config_files`) work correctly
- **FR-009**: All tests MUST include idempotency verification using `plancheck.ExpectEmptyPlan()`
- **FR-010**: All tests MUST use modern testing patterns from terraform-plugin-testing v1.13.3+ (statecheck, knownvalue matchers)

### Key Entities

- **Test Function**: A Go acceptance test that validates one or more related optional fields through CRUD operations
- **Test Configuration Helper**: A Go function that generates Terraform HCL configuration for test steps
- **Field Coverage Matrix**: Tracks which optional fields have test coverage and what operations are tested

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 8 user stories have passing acceptance tests demonstrating CRUD + idempotency
- **SC-002**: Test coverage for optional fields increases from ~15 to 40+ fields (based on issue #66 analysis)
- **SC-003**: All new tests use modern statecheck/knownvalue patterns (no legacy TestCheckResourceAttr for new tests)
- **SC-004**: All tests complete within 15 minutes when run as part of the full acceptance test suite
- **SC-005**: Zero test flakiness - all tests pass consistently on 3 consecutive runs
- **SC-006**: Gap analysis report shows improved coverage for bcm_cmdevice_category optional fields

## Assumptions

- BCM cluster is available at 172.21.15.254:8081 for acceptance tests
- Existing test infrastructure (createTestBCMClient, generateUniqueTestName, etc.) remains unchanged
- The buildAPIEntity and readCategory helper functions in the resource already support all optional fields
- Some fields may require specific BCM cluster configuration to test (e.g., BMC settings require BMC-enabled hardware)
- Fields marked with "TODO Phase 6" in the resource implementation may have limited API support

## Out of Scope

- Adding new optional fields to the resource schema
- Modifying the BCM API client
- Performance optimization of existing tests
- Documentation generation updates
- Unit tests (this feature focuses on acceptance tests only)
