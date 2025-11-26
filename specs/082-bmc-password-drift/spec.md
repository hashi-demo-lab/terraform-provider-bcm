# Feature Specification: Fix BMC Settings Password Perpetual Drift

**Feature Branch**: `082-bmc-password-drift`
**Created**: 2025-11-26
**Status**: Draft
**GitHub Issue**: #82
**Input**: Fix bmc_settings.password perpetual drift issue in bcm_cmdevice_category resource

## Problem Statement

The `bmc_settings.password` field in the `bcm_cmdevice_category` resource causes perpetual drift because the password is not preserved from state during Read operations.

### Root Cause Analysis

In `resource_cmdevice_category.go` (around line 2334), during the `readCategory()` function, the password is explicitly set to null:

```go
bmcModel := BMCSettingsModel{
    // ...
    Password: types.StringNull(), // Don't read back password (sensitive)
    // ...
}
```

This is **correct security behavior** because:
1. BCM API does not return passwords in responses (standard security practice)
2. Sensitive data should not be exposed through API reads

However, the provider does not preserve the password from the previous state/plan, unlike other fields such as `roles`, `fsexports`, `static_routes`, `gpu_settings`, and `services` which are preserved at lines 1189-1195.

### Current Behavior (Bug)

1. User configures `bmc_settings.password` in Terraform configuration
2. `terraform apply` sends password to BCM during Create/Update (lines 1948-1968)
3. During Read (refresh), password is set to `null` (line 2334)
4. Next `terraform plan` compares config (with password) to state (null password)
5. Result: Perpetual drift - password always shows as changed

### Expected Behavior (After Fix)

1. User configures `bmc_settings.password` in Terraform configuration
2. `terraform apply` sends password to BCM during Create/Update
3. During Read (refresh), password is preserved from prior state
4. Next `terraform plan` shows no changes if password is unchanged in configuration
5. Result: Stable state - no drift when password has not changed

## User Scenarios & Testing

### User Story 1 - Password Stability on Refresh (Priority: P1)

As a Terraform user managing BMC credentials, I want my `bmc_settings.password` to remain stable across plan/apply cycles so that I do not see false drift on every `terraform plan`.

**Why this priority**: This is the core bug being fixed. Without this, users cannot reliably use BMC password management through Terraform.

**Independent Test**: Can be fully tested by creating a category with BMC password, running `terraform plan` after apply, and verifying no changes detected.

**Acceptance Scenarios**:

1. **Given** a category resource with `bmc_settings.password = "secret123"` has been applied, **When** user runs `terraform plan`, **Then** the plan shows "No changes" for the password field

2. **Given** a category resource with `bmc_settings.password = "secret123"` exists in state, **When** user runs `terraform apply` without config changes, **Then** no update is performed

3. **Given** a category resource with `bmc_settings.password = "secret123"` exists, **When** user changes password to `"newsecret456"` and runs `terraform plan`, **Then** the plan shows password will be updated

---

### User Story 2 - Password Update Detection (Priority: P1)

As a Terraform user, I want password changes in my configuration to be detected and applied so that I can rotate BMC credentials through Terraform.

**Why this priority**: Password rotation is a critical security operation that must work correctly.

**Independent Test**: Can be tested by modifying the password value in configuration and verifying the change is detected and applied.

**Acceptance Scenarios**:

1. **Given** a category with `bmc_settings.password = "oldpass"`, **When** user changes config to `password = "newpass"` and runs `terraform plan`, **Then** plan shows password change

2. **Given** a planned password change, **When** user runs `terraform apply`, **Then** the new password is sent to BCM API

---

### User Story 3 - Import with Password (Priority: P2)

As a Terraform user importing existing categories, I want to be able to set the BMC password after import without unexpected drift.

**Why this priority**: Import is a secondary workflow but important for adopting Terraform on existing infrastructure.

**Independent Test**: Can be tested by importing a category, adding BMC password to config, and verifying stable state after apply.

**Acceptance Scenarios**:

1. **Given** a category imported without BMC settings configured, **When** user adds `bmc_settings.password` to config and applies, **Then** password is sent to BCM and state is stable

2. **Given** an imported category with BMC settings added, **When** user runs subsequent `terraform plan`, **Then** no drift is detected

---

### Edge Cases

- What happens when `bmc_settings` block is removed entirely from configuration?
  - BMC settings should be set to null, and password should not cause drift
- What happens when `bmc_settings` is added for the first time to an existing category?
  - Password should be sent to BCM and preserved in state
- What happens when password is set to empty string `""`?
  - Empty string should be treated as a valid value and preserved (not converted to null)
- What happens when only non-password BMC fields are updated?
  - Password should be preserved from state while other fields are updated

## Requirements

### Functional Requirements

- **FR-001**: System MUST preserve `bmc_settings.password` from prior state during Read operations
- **FR-002**: System MUST detect changes to `bmc_settings.password` in configuration and plan updates accordingly
- **FR-003**: System MUST send password to BCM API during Create and Update operations when configured
- **FR-004**: System MUST NOT expose password in state file beyond what user configured (maintain sensitivity)
- **FR-005**: System MUST handle the case where `bmc_settings` is null or not configured without errors

### Key Entities

- **CMDeviceCategoryResourceModel**: Contains `BMCSettings` field as `types.Object`
- **BMCSettingsModel**: Contains `Password` field as `types.String` (Sensitive)
- **State Management**: Password preserved across Read operations using pattern from lines 1068-1077

## Success Criteria

### Measurable Outcomes

- **SC-001**: Running `terraform plan` after a successful `terraform apply` with BMC password configured shows "No changes"
- **SC-002**: Changing password value in configuration correctly shows planned change in `terraform plan`
- **SC-003**: All existing acceptance tests continue to pass (no regression)
- **SC-004**: New acceptance test specifically validates password drift fix passes

## API Contract

### Existing Behavior (BCM API)

**Create/Update Request** - Password is sent:
```json
{
  "service": "CMDevice",
  "call": "updateCategory",
  "args": [{
    "bmcSettings": {
      "userName": "admin",
      "password": "secret123",
      "privilege": "admin"
    }
  }]
}
```

**Read Response** - Password is NOT returned:
```json
{
  "bmcSettings": {
    "uuid": "...",
    "userName": "admin",
    "privilege": "admin"
  }
}
```

### Provider Contract

**Before Fix (Broken)**:
- Read: `Password = types.StringNull()`
- Result: Config ("secret123") != State (null) = Drift detected

**After Fix (Correct)**:
- Read: `Password = originalBMCSettings.Password` (preserved from state)
- Result: Config ("secret123") == State ("secret123") = No drift

## Design Options

### Option A: State Preservation Pattern (Recommended)

Follow the established pattern used for other fields (lines 1068-1077, 1189-1195):

```go
// In Read(), before readCategory()
originalBMCSettings := state.BMCSettings

// After readCategory(), restore password from original state
// (Implementation details to be defined in plan.md)
```

**Pros**:
- Consistent with existing codebase patterns (roles, fsexports, static_routes, etc.)
- No Terraform version requirements beyond current
- No schema changes required
- Password remains in state (encrypted by backend)

**Cons**:
- Password stored in state (sensitive but present)
- Relies on Terraform's sensitive attribute handling

### Option B: Write-Only Arguments (Terraform 1.11+)

Use Terraform Plugin Framework's `WriteOnly: true` attribute flag:

```go
"password_wo": schema.StringAttribute{
    Optional:    true,
    WriteOnly:   true,
    Sensitive:   true,
    Description: "BMC password (write-only, not stored in state)",
},
"password_wo_version": schema.Int64Attribute{
    Optional:    true,
    Description: "Version number to trigger password updates",
},
```

**Pros**:
- Password NEVER stored in state or plan files
- Aligns with HashiCorp best practices for sensitive data
- Future-proof approach

**Cons**:
- Requires Terraform 1.11+ (released early 2025)
- Breaking change: requires new attribute names (`password_wo` + `password_wo_version`)
- More complex user experience (version tracking)
- Would need migration strategy for existing users

### Recommendation

**Option A (State Preservation)** is recommended for this fix because:
1. Maintains backward compatibility with existing configurations
2. Consistent with established patterns in this codebase
3. No Terraform version requirements
4. Simpler user experience

Option B (Write-Only) should be considered for a future enhancement (separate issue) once Terraform 1.11+ adoption increases.

## Implementation Constraints

### Testing Requirements (TDD)

Per project TDD constitution, tests must be written first:

1. **Acceptance Test**: `TestAccCMDeviceCategory_BMCPasswordNoDrift`
   - Create category with BMC password
   - Verify idempotency (second plan shows no changes)
   - Use `plancheck.ExpectEmptyPlan()` pattern

2. **Acceptance Test**: `TestAccCMDeviceCategory_BMCPasswordUpdate`
   - Create category with BMC password
   - Update password value
   - Verify change is detected and applied

### Files to Modify

- `internal/provider/resource_cmdevice_category.go` - Read function to preserve password
- `internal/provider/resource_cmdevice_category_test.go` - New acceptance tests

### Constraints

- **No breaking changes**: Existing configurations must continue to work
- **Backward compatibility**: State files without BMC password should not error
- **Security**: Password must remain marked as Sensitive in schema
- **No API changes**: Fix is entirely within provider logic

## Assumptions

- BCM API will continue to not return passwords in responses (standard security practice)
- The `bmc_settings` nested object structure remains unchanged
- Terraform Plugin Framework v1.16+ behavior for nested objects is stable
- The established pattern for preserving fields from state (used for roles, fsexports, etc.) is the correct approach for this fix
