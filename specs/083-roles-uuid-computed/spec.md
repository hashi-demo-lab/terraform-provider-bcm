# Feature Specification: Fix roles[].uuid Computed Value Population

**Feature Branch**: `083-roles-uuid-computed`
**Created**: 2025-11-26
**Status**: Draft
**Issue**: [#83](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/83)
**Input**: bcm_cmdevice_category: Fix roles[].uuid computed value not populated from BCM API

## Problem Statement

The `roles[].uuid` computed attribute in `bcm_cmdevice_category` resource is never populated with BCM-assigned UUIDs. Users configure roles with `name` and `child_type`, expecting BCM to assign a UUID during creation that is then available in Terraform state for reference in other resources.

### Root Cause Analysis

In `resource_cmdevice_category.go`, the Read operation preserves roles from state instead of reading from BCM API:

1. **Line 1075**: `originalRoles := state.Roles` - Original roles captured from state before API read
2. **Line 1193**: `state.Roles = originalRoles` - After `readCategory()`, original roles are restored, **discarding BCM API values**
3. **Lines 2249-2276**: The code correctly parses role UUIDs from BCM API response, but this parsed data is **overwritten** by the preservation logic

### Current Behavior

```hcl
resource "bcm_cmdevice_category" "example" {
  name = "test-category"
  roles {
    name       = "head"
    child_type = "HeadNode"
  }
}

# After apply:
# roles[0].uuid = null (or unknown)  <-- BUG: should contain BCM-assigned UUID
```

### Expected Behavior

```hcl
# After apply:
# roles[0].uuid = "550e8400-e29b-41d4-a716-446655440000"  <-- BCM-assigned UUID
```

## User Scenarios & Testing

### User Story 1 - Role UUID Available After Create (Priority: P1)

As a Terraform practitioner managing BCM infrastructure, I want the BCM-assigned role UUIDs to be populated in Terraform state after creating a category with roles, so that I can reference these UUIDs in other resources or outputs.

**Why this priority**: This is the core bug fix - without role UUIDs in state, users cannot build dependent resources or track role identifiers.

**Independent Test**: Create a category with roles, verify the `roles[0].uuid` attribute is populated with a valid UUID string after apply.

**Acceptance Scenarios**:

1. **Given** a category configuration with one role specifying `name` and `child_type`, **When** terraform apply completes successfully, **Then** the role's `uuid` attribute contains a non-empty BCM-assigned UUID string
2. **Given** a category configuration with multiple roles, **When** terraform apply completes, **Then** each role's `uuid` attribute contains a unique BCM-assigned UUID

---

### User Story 2 - Role UUID Preserved on Refresh (Priority: P1)

As a Terraform practitioner, I want role UUIDs to remain populated after a terraform refresh or plan operation, so that my state accurately reflects the BCM-assigned identifiers.

**Why this priority**: Equal priority to Story 1 - if UUIDs are populated on create but lost on refresh, the bug is not fully resolved.

**Independent Test**: After initial create, run terraform refresh and verify role UUIDs remain populated.

**Acceptance Scenarios**:

1. **Given** a category with roles that have populated UUIDs in state, **When** terraform refresh is executed, **Then** the role UUIDs remain populated with the same values
2. **Given** a category with roles, **When** no configuration changes are made and terraform plan is run, **Then** the plan shows no changes (no drift detected)

---

### User Story 3 - Role UUID Available After Import (Priority: P2)

As a Terraform practitioner importing existing BCM categories, I want the imported state to include role UUIDs, so that I have complete visibility into the BCM-assigned identifiers.

**Why this priority**: Import is a secondary workflow, but important for adopting existing infrastructure.

**Independent Test**: Import an existing category with roles and verify role UUIDs are populated.

**Acceptance Scenarios**:

1. **Given** an existing BCM category with roles, **When** terraform import is executed, **Then** the imported state includes populated role UUIDs

---

### User Story 4 - Merge User Config with API Values (Priority: P1)

As a Terraform practitioner, I want the provider to preserve my configured role attributes (name, child_type, add_services) while populating computed attributes (uuid) from BCM API, so that I get a complete and accurate state without false drift.

**Why this priority**: This addresses the core fix approach - merging rather than wholesale replacement.

**Independent Test**: Create a category with roles, modify the role externally in BCM (if possible), run terraform plan, and verify only true drift is detected.

**Acceptance Scenarios**:

1. **Given** a category with roles where user specified `name="head"` and `child_type="HeadNode"`, **When** terraform read completes, **Then** state contains user's `name` and `child_type` values plus BCM's computed `uuid`
2. **Given** a category where user specified `add_services=true` for a role, **When** terraform read completes, **Then** state preserves the user's `add_services` value alongside the computed `uuid`

---

### Edge Cases

- What happens when BCM returns an empty roles array but user configured roles? (Role was deleted externally - should detect as drift)
- What happens when BCM returns roles not in user config? (Additional roles added externally - should detect as drift)
- How does the system handle a role that exists in config but BCM returns no UUID for it? (Unexpected API response - should log warning, preserve what we can)

## Requirements

### Functional Requirements

- **FR-001**: Provider MUST populate `roles[].uuid` computed attribute with BCM-assigned UUID after category create operation
- **FR-002**: Provider MUST preserve `roles[].uuid` values in state during refresh operations by reading from BCM API
- **FR-003**: Provider MUST match roles by `name` attribute when merging user configuration with BCM API response
- **FR-004**: Provider MUST preserve user-specified role attributes (`name`, `child_type`, `add_services`) while populating computed attributes (`uuid`)
- **FR-005**: Provider MUST populate `roles[].uuid` during import operations
- **FR-006**: Provider MUST detect drift when roles are added or removed externally in BCM
- **FR-007**: Provider MUST NOT cause false drift detection due to role UUID population logic

### Key Entities

- **CategoryRoleModel**: Represents a role within a category
  - `name` (Required): Role identifier used for matching between config and API
  - `child_type` (Required): Role type classification
  - `uuid` (Computed): BCM-assigned unique identifier - **this is the field to populate**
  - `add_services` (Optional): Whether to add role services

## Success Criteria

### Measurable Outcomes

- **SC-001**: After terraform apply on a category with roles, 100% of roles have non-empty `uuid` attributes in state
- **SC-002**: Terraform plan shows no changes immediately after terraform apply (idempotency verified)
- **SC-003**: Terraform refresh preserves role UUIDs without causing drift on unchanged configurations
- **SC-004**: All existing acceptance tests for bcm_cmdevice_category continue to pass
- **SC-005**: New drift detection tests for roles[].uuid pass

## Technical Context

### Files Requiring Modification

**Primary file**: `internal/provider/resource_cmdevice_category.go`

- **Lines 1068-1077**: Preservation logic captures original roles - needs modification to support merging
- **Lines 1189-1195**: Unconditional restoration of original roles - needs replacement with merge logic
- **Lines 2249-2276**: Correct role UUID parsing from BCM API - this code is correct, needs to be used

### Proposed Fix Approach

Replace the unconditional preservation logic with a merge strategy:

1. Parse roles from BCM API response (existing code at lines 2249-2276)
2. Match API roles to config roles by `name` attribute
3. For each matched role:
   - Preserve user-specified values: `name`, `child_type`, `add_services`
   - Populate computed values from API: `uuid`
4. Handle unmatched roles appropriately (detect as drift)

### Test Verification

The fix should be verified with:

1. **Unit tests**: Verify role merge logic correctly combines config and API values
2. **Acceptance tests**:
   - `TestAccCMDeviceCategory_RolesWithUUID` - verify UUID population on create
   - `TestAccCMDeviceCategory_RolesIdempotency` - verify no drift after apply
   - `TestAccCMDeviceCategory_RolesDrift` - verify external changes detected

## Assumptions

- BCM API consistently returns role UUIDs in the `roles` array when querying a category
- Role `name` is unique within a category and can be used for matching
- The existing role parsing logic (lines 2249-2276) correctly extracts UUIDs from BCM API responses
