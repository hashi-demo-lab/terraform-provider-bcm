# Implementation Plan: Fix BMC Settings Password Perpetual Drift

**Branch**: `082-bmc-password-drift` | **Date**: 2025-11-26 | **Spec**: [spec.md](spec.md)
**Input**: Feature specification from `/specs/082-bmc-password-drift/spec.md`

## Summary

Fix perpetual drift of `bmc_settings.password` in `bcm_cmdevice_category` resource by preserving the password from prior state during Read operations. This follows the established State Preservation Pattern already used for `roles`, `fsexports`, `static_routes`, `gpu_settings`, and `services` fields at lines 1068-1077, 1189-1195.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: Terraform state (password as sensitive attribute)
**Testing**: TF_ACC=1 acceptance tests with BCM API
**Target Platform**: Linux (BCM cluster at 172.21.15.254:8081)
**Project Type**: Terraform Provider (single Go module)
**Performance Goals**: N/A (state management fix)
**Constraints**: Backward compatible, no breaking changes
**Scale/Scope**: Single resource fix (`bcm_cmdevice_category`)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Rule | Status | Notes |
|------|--------|-------|
| TDD: Tests first | PASS | Write acceptance tests before implementation |
| Single responsibility | PASS | Fix affects only Read operation password handling |
| Minimal changes | PASS | Follows existing preservation pattern |
| No breaking changes | PASS | Existing configurations continue to work |
| Sensitive data handling | PASS | Password remains marked as Sensitive |

## Project Structure

### Documentation (this feature)

```text
specs/082-bmc-password-drift/
├── spec.md              # Feature specification (complete)
├── plan.md              # This file
├── research.md          # Phase 0 output (minimal - pattern already established)
├── data-model.md        # Phase 1 output (N/A - no new entities)
├── quickstart.md        # Phase 1 output (N/A - fix only)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmdevice_category.go       # Main resource - Read() modification
└── resource_cmdevice_category_test.go  # New acceptance tests
```

## Complexity Tracking

No constitution violations. This is a straightforward fix following an established pattern.

---

## Phase 0: Research Summary

### Research Task: Existing Preservation Pattern Analysis

**Decision**: Use State Preservation Pattern from lines 1068-1077, 1189-1195

**Rationale**: The codebase already has a proven, tested pattern for preserving fields that BCM API does not return or returns differently:

1. **Lines 1068-1077 (Read function)**: Original values captured before `readCategory()` call
   ```go
   originalStaticRoutes := state.StaticRoutes
   originalFSExports := state.FSExports
   originalRoles := state.Roles
   originalGPUSettings := state.GPUSettings
   originalServices := state.Services
   ```

2. **Lines 1189-1195 (Read function)**: Values restored after `readCategory()` call
   ```go
   state.StaticRoutes = originalStaticRoutes
   state.FSExports = originalFSExports
   state.Roles = originalRoles
   state.GPUSettings = originalGPUSettings
   state.Services = originalServices
   ```

3. **Lines 1197-1232**: Complex preservation for `software_image_proxy` with merge logic

**Alternatives Considered**:

| Alternative | Why Rejected |
|-------------|--------------|
| Write-Only Arguments (Terraform 1.11+) | Breaking change, version requirement, more complex UX |
| API modification | Not possible - BCM is external system |
| Schema UseStateForUnknown | Does not apply to nested object attributes |

### Research Task: BMC Settings Object Structure

**Decision**: BMC password requires nested object attribute handling

**Findings**:
- `bmc_settings` is `types.Object` containing `BMCSettingsModel`
- Password is `types.String` with `Sensitive: true`
- Pattern from `software_image_proxy` (lines 1197-1232) shows how to handle nested objects with selective preservation

---

## Phase 1: Design

### Implementation Approach

The fix requires three changes to `resource_cmdevice_category.go`:

#### Change 1: Capture Original BMC Settings (Line ~1077)

Add preservation capture alongside existing fields:

```go
// In Read(), after other original* captures
originalBMCSettings := state.BMCSettings
```

#### Change 2: Restore Password After readCategory (Line ~1196)

Add password restoration logic after existing preservation block:

```go
// CRITICAL FIX: Preserve bmc_settings.password from state
// BCM API does not return password (sensitive) - preserve user's configured value
if !originalBMCSettings.IsNull() && !originalBMCSettings.IsUnknown() {
    var originalBMC BMCSettingsModel
    originalBMCSettings.As(ctx, &originalBMC, basetypes.ObjectAsOptions{})

    // Only restore if password was configured in state
    if !originalBMC.Password.IsNull() && !originalBMC.Password.IsUnknown() {
        // Get API-returned values from current state
        var apiBMC BMCSettingsModel
        if !state.BMCSettings.IsNull() {
            state.BMCSettings.As(ctx, &apiBMC, basetypes.ObjectAsOptions{})
        }

        // Build merged model: API values + preserved password
        mergedBMC := BMCSettingsModel{
            UUID:               apiBMC.UUID,
            UserName:           apiBMC.UserName,
            Password:           originalBMC.Password, // Preserve from state
            Privilege:          apiBMC.Privilege,
            UserID:             apiBMC.UserID,
            FirmwareManageMode: apiBMC.FirmwareManageMode,
            LeakPolicy:         apiBMC.LeakPolicy,
            LeakReactionDelay:  apiBMC.LeakReactionDelay,
            PowerResetDelay:    apiBMC.PowerResetDelay,
        }

        // Convert back to types.Object
        bmcSettingsObjectType := map[string]attr.Type{
            "uuid":                 types.StringType,
            "user_name":            types.StringType,
            "password":             types.StringType,
            "privilege":            types.StringType,
            "user_id":              types.Int64Type,
            "firmware_manage_mode": types.StringType,
            "leak_policy":          types.StringType,
            "leak_reaction_delay":  types.Float64Type,
            "power_reset_delay":    types.Int64Type,
        }

        bmcObj, diagsBMC := types.ObjectValueFrom(ctx, bmcSettingsObjectType, mergedBMC)
        if !diagsBMC.HasError() {
            state.BMCSettings = bmcObj
            tflog.Debug(ctx, "Preserved bmc_settings.password from state", map[string]interface{}{
                "has_password": !originalBMC.Password.IsNull(),
            })
        }
    }
}
```

### Edge Cases Handled

| Edge Case | Handling |
|-----------|----------|
| `bmc_settings` is null in state | Skip preservation - nothing to preserve |
| `bmc_settings` added first time | Password comes from plan in Create, preserved in subsequent Reads |
| Password is empty string `""` | Preserved as empty string (not converted to null) |
| Only non-password BMC fields updated | Password preserved while other fields update |
| Import without password | `originalBMCSettings` will be null after import, no preservation occurs |

### Test Implementation (TDD - RED Phase First)

#### Test 1: `TestAccCMDeviceCategory_BMCPasswordNoDrift`

**Purpose**: Verify idempotency - no drift detected on subsequent plans

```go
func TestAccCMDeviceCategory_BMCPasswordNoDrift(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-bmc-nodrift")
    testAccCMDeviceCategoryPreCheck(t, categoryName)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create category with BMC password
            {
                Config: testAccCMDeviceCategoryResourceConfig_BMCPassword(categoryName, "secret123"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_category.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(categoryName),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_category.test",
                        tfjsonpath.New("bmc_settings").AtMapKey("user_name"),
                        knownvalue.StringExact("admin"),
                    ),
                },
            },
            // Step 2: Idempotency check - no drift expected
            {
                Config: testAccCMDeviceCategoryResourceConfig_BMCPassword(categoryName, "secret123"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
            // Step 3: Import and verify (password not imported)
            {
                Config:            testAccCMDeviceCategoryResourceConfig_BMCPassword(categoryName, "secret123"),
                ResourceName:      "bcm_cmdevice_category.test",
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{
                    "force",
                    "bmc_settings", // Password cannot be imported from BCM
                },
            },
        },
    })
}
```

#### Test 2: `TestAccCMDeviceCategory_BMCPasswordUpdate`

**Purpose**: Verify password changes are detected and applied

```go
func TestAccCMDeviceCategory_BMCPasswordUpdate(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-bmc-update")
    testAccCMDeviceCategoryPreCheck(t, categoryName)

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create with initial password
            {
                Config: testAccCMDeviceCategoryResourceConfig_BMCPassword(categoryName, "oldpass123"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_category.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(categoryName),
                    ),
                },
            },
            // Step 2: Update password - should detect change
            {
                Config: testAccCMDeviceCategoryResourceConfig_BMCPassword(categoryName, "newpass456"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmdevice_category.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(categoryName),
                    ),
                },
            },
            // Step 3: Idempotency after update
            {
                Config: testAccCMDeviceCategoryResourceConfig_BMCPassword(categoryName, "newpass456"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
        },
    })
}
```

#### Test Config Helper

```go
func testAccCMDeviceCategoryResourceConfig_BMCPassword(name, password string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = [for c in data.bcm_cmdevice_categories.all.categories : c.management_network if c.name == "default"][0]
  software_image_uuid     = [for i in data.bcm_cmpart_softwareimages.all.software_images : i.uuid if i.name == "default-image"][0]
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = "BMC password drift test"

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }

  bmc_settings = {
    user_name = "admin"
    password  = %[5]q
    privilege = "admin"
  }
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        password,
    )
}
```

---

## Risk Assessment

### Risk 1: Nested Object State Handling

**Risk Level**: Low
**Description**: Terraform's nested object handling can be tricky with `types.ObjectValueFrom`
**Mitigation**: Pattern already proven with `software_image_proxy` at lines 1197-1232

### Risk 2: Import Behavior

**Risk Level**: Low
**Description**: After import, password will be null since BCM doesn't return it
**Mitigation**:
- Test explicitly ignores `bmc_settings` during import verification
- First apply after import will set password from configuration

### Risk 3: Empty Password vs Null Password

**Risk Level**: Medium
**Description**: Need to distinguish between `password = ""` and no password configured
**Mitigation**:
- `!originalBMC.Password.IsNull()` check only preserves when password was set
- Empty string `""` is treated as a valid configured value

### Risk 4: Concurrent Modification

**Risk Level**: Low
**Description**: If BMC password is changed outside Terraform, Terraform won't detect it
**Mitigation**:
- This is expected behavior - BCM doesn't expose passwords
- Same limitation exists for all sensitive fields

---

## Success Criteria

| Criteria | Verification Method |
|----------|---------------------|
| SC-001: No drift on `terraform plan` after apply | `TestAccCMDeviceCategory_BMCPasswordNoDrift` passes |
| SC-002: Password changes detected | `TestAccCMDeviceCategory_BMCPasswordUpdate` passes |
| SC-003: No regression | All existing acceptance tests pass |
| SC-004: Idempotency | `plancheck.ExpectEmptyPlan()` succeeds |

---

## Implementation Order (TDD)

1. **RED Phase**: Write failing tests
   - Add `TestAccCMDeviceCategory_BMCPasswordNoDrift`
   - Add `TestAccCMDeviceCategory_BMCPasswordUpdate`
   - Add `testAccCMDeviceCategoryResourceConfig_BMCPassword` helper
   - Run tests - expect failures

2. **GREEN Phase**: Implement fix
   - Add `originalBMCSettings` capture at line ~1077
   - Add password preservation logic at line ~1196
   - Run tests - expect passes

3. **REFACTOR Phase**: Clean up
   - Add debug logging
   - Update/remove skip comment on existing `TestAccCMDeviceCategoryResource_BMCSettings`
   - Run `make generate` to update documentation

---

## Files Modified

| File | Changes |
|------|---------|
| `internal/provider/resource_cmdevice_category.go` | Add BMC password preservation in Read() |
| `internal/provider/resource_cmdevice_category_test.go` | Add 2 new acceptance tests + helper |

---

## Generated Artifacts

- `/workspace/specs/082-bmc-password-drift/plan.md` - This implementation plan
- No `research.md` needed - pattern already documented in codebase
- No `data-model.md` needed - no new entities
- No `quickstart.md` needed - fix only, no new developer setup
- No `contracts/` needed - no API changes
