# Feature Specification: Fix Device Role Association Bug

**Feature Branch**: `086-fix-device-role-association`
**Created**: 2025-11-26
**Status**: Draft
**Input**: User description: "Fix device role association bug - roles should be passed as names, not UUIDs, with client-side validation"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Device Role Assignment by Name (Priority: P1)

As a Terraform user, I want to assign roles to devices using role names (like "backup", "provisioning") instead of UUIDs, so that my Terraform configurations are human-readable, maintainable, and less error-prone.

**Why this priority**: This is the core bug fix. Currently the example code shows users looking up role UUIDs via the data source, but the provider should accept role names directly since BCM uses role names internally for association. Using UUIDs is confusing and unnecessary.

**Independent Test**: Can be fully tested by creating a device with `roles = ["backup", "provisioning"]` using role names directly, and verifying the roles are correctly associated with the device in BCM.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with roles named "backup" and "provisioning", **When** I create a device with `roles = ["backup", "provisioning"]`, **Then** the device is created with both roles associated correctly.

2. **Given** a device with roles assigned by name, **When** I read the device state, **Then** the roles attribute shows the same role names that were configured.

3. **Given** an existing device created with roles by UUID (legacy), **When** I import the device, **Then** the roles attribute correctly shows the role names (not UUIDs).

---

### User Story 2 - Client-Side Role Validation (Priority: P1)

As a Terraform user, I want the provider to validate that role names exist in the BCM cluster before attempting to create or update a device, so that I receive a clear error message if I specify an invalid role name.

**Why this priority**: The BCM API performs zero validation on role associations. Invalid role names are silently ignored or cause cryptic failures. Client-side validation is essential for a good user experience.

**Independent Test**: Can be tested by attempting to create a device with `roles = ["nonexistent-role"]` and verifying that a clear validation error is returned before any API call is made.

**Acceptance Scenarios**:

1. **Given** a BCM cluster that does not have a role named "nonexistent-role", **When** I create a device with `roles = ["nonexistent-role"]`, **Then** I receive an error message stating "Role 'nonexistent-role' does not exist in the BCM cluster".

2. **Given** a BCM cluster with roles "backup" and "provisioning", **When** I create a device with `roles = ["backup", "invalid-role"]`, **Then** I receive an error message identifying "invalid-role" as not existing.

3. **Given** an existing device, **When** I update it with `roles = ["nonexistent-role"]`, **Then** I receive the same validation error as during creation.

---

### User Story 3 - Updated Example Documentation (Priority: P2)

As a new Terraform user learning to use the BCM provider, I want the example code to show the correct, simplified approach for assigning roles, so that I can follow best practices from the start.

**Why this priority**: Correct documentation prevents users from learning the wrong pattern and reduces support burden.

**Independent Test**: Can be validated by reviewing the updated example file and ensuring it demonstrates role assignment by name without UUID lookups.

**Acceptance Scenarios**:

1. **Given** the example file `examples/resources/bcm_cmdevice_device/with_roles.tf`, **When** I review it, **Then** it shows `roles = ["backup", "provisioning"]` instead of UUID lookups.

2. **Given** the updated example, **When** terraform validate is run, **Then** it passes validation.

3. **Given** the example, **When** I run terraform plan against a real BCM cluster, **Then** it produces a valid plan without errors.

---

### User Story 4 - Backward Compatibility with UUID Input (Priority: P3)

As a user with existing Terraform configurations that use role UUIDs, I want the provider to continue accepting UUIDs as a fallback, so that my existing configurations do not break immediately.

**Why this priority**: Backward compatibility ensures existing users are not disrupted. However, this is lower priority than fixing the primary issue.

**Independent Test**: Can be tested by creating a device with both role names and role UUIDs to verify both formats are accepted.

**Acceptance Scenarios**:

1. **Given** an existing Terraform configuration using role UUIDs, **When** I apply it with the updated provider, **Then** it continues to work without modification.

2. **Given** a role UUID that exists in the cluster, **When** I specify `roles = ["<uuid>"]`, **Then** the role is correctly associated (though a deprecation warning may be shown).

---

### Edge Cases

- What happens when a role name contains special characters (spaces, unicode)?
  - The provider should validate the role name exists exactly as specified; BCM role names are typically alphanumeric with hyphens.

- How does the system handle case sensitivity in role names?
  - Role names should be matched case-sensitively, matching BCM's behavior.

- What happens when the same role is specified multiple times?
  - Since roles is a Set, duplicates should be automatically deduplicated.

- What happens when roles list is empty `roles = []`?
  - Empty list should explicitly remove all role associations from the device.

- What happens when roles attribute is omitted entirely?
  - Device should keep existing roles (or have no roles if new device).

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: The `roles` attribute on `bcm_cmdevice_device` resource MUST accept role names (strings) as input.

- **FR-002**: The provider MUST validate that each specified role name exists in the BCM cluster before sending any create/update API request.

- **FR-003**: The provider MUST return a clear, user-friendly error message when a role name is not found, including the invalid role name and a hint to use the `bcm_cmdevice_roles` data source to discover available roles.

- **FR-004**: The provider MUST query the BCM cluster to resolve role names to their corresponding role objects for the API request (BCM API requires full role objects, not names).

- **FR-005**: The provider MUST continue to accept role UUIDs as input for backward compatibility.

- **FR-006**: When reading device state, the provider MUST return role names (not UUIDs) in the `roles` attribute.

- **FR-007**: The provider MUST handle the scenario where a role is deleted from BCM after being assigned to a device (drift detection).

- **FR-008**: The example file `with_roles.tf` MUST be updated to demonstrate role assignment by name.

- **FR-009**: The schema documentation for the `roles` attribute MUST be updated to reflect that role names are the expected input format.

### Key Entities

- **Role**: Represents a device role in BCM. Key attributes: `name` (string, the role identifier used in Terraform configs), `uuid` (string, internal BCM identifier), `child_type` (string, role type like "HeadNodeRole").

- **Device**: Represents a compute node in BCM. The `roles` attribute associates zero or more roles to the device.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Users can assign roles to devices using role names directly (e.g., `roles = ["backup"]`) without needing to look up UUIDs.

- **SC-002**: Invalid role names produce clear validation errors within 2 seconds of plan/apply (client-side validation).

- **SC-003**: 100% of existing Terraform configurations using role UUIDs continue to work without modification (backward compatibility).

- **SC-004**: Example code demonstrates the simplified, correct approach for role assignment.

- **SC-005**: Role state is consistently represented by role names in Terraform state, regardless of how they were specified in configuration.

## Assumptions

- Role names in BCM are unique within a cluster (no two roles have the same name).
- Role names are case-sensitive and should be matched exactly.
- The BCM API will continue to require full role objects when assigning roles to devices.
- The `bcm_cmdevice_roles` data source can be used to query all available roles in the cluster.
- BCM API `validateDevice` does not validate role associations, requiring provider-side validation.
