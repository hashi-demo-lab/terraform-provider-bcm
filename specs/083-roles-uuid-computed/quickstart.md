# Quickstart: Fix roles[].uuid Computed Value Population

**Issue**: #83 - bcm_cmdevice_category: roles[].uuid computed value never populated from BCM API
**Branch**: `083-roles-uuid-computed`

## Problem Summary

The `roles[].uuid` attribute in `bcm_cmdevice_category` is always null after apply because the Read operation overwrites BCM API response data with the original state values.

## Root Cause

In `resource_cmdevice_category.go`:

```go
// Line 1075: Captures roles before API read
originalRoles := state.Roles

// Lines 2249-2276: BCM API response is correctly parsed with UUIDs
// ... (this works correctly)

// Line 1193: BUG - overwrites API data with original state
state.Roles = originalRoles  // Discards BCM-assigned UUIDs!
```

## Fix Overview

Replace unconditional state preservation with a merge strategy:
1. Match roles by `name` attribute
2. Preserve user-specified values: `name`, `child_type`, `add_services`
3. Populate computed values: `uuid` from BCM API

## Implementation Steps

### Step 1: Run Existing Tests (Baseline)

```bash
# Verify existing tests pass before changes
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"
```

### Step 2: Add Failing Tests (RED Phase)

Add to `internal/provider/resource_cmdevice_category_test.go`:

```go
// TestAccCMDeviceCategory_RolesUUIDPopulated verifies role UUID is populated after create
func TestAccCMDeviceCategory_RolesUUIDPopulated(t *testing.T) {
    // See plan.md for full test implementation
}
```

Run and verify tests fail:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "RolesUUID"
# Expected: FAIL - roles[].uuid is null
```

### Step 3: Implement Fix (GREEN Phase)

1. Add helper function `mergeRolesWithAPIResponse()` to `resource_cmdevice_category.go`
2. Replace line 1193:
   ```go
   // Before:
   state.Roles = originalRoles

   // After:
   state.Roles = mergeRolesWithAPIResponse(ctx, originalRoles, state.Roles)
   ```

### Step 4: Verify Tests Pass

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "RolesUUID"
# Expected: PASS
```

### Step 5: Run Full Test Suite

```bash
# All category tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"

# Lint and generate docs
make lint
make generate
```

## Test Config Helper

Add this helper function for tests with roles:

```go
func testAccCMDeviceCategoryResourceConfig_WithRole(name, roleName, childType string) string {
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
  management_network_uuid = data.bcm_cmdevice_categories.all.categories[0].management_network_id
  software_image_uuid = data.bcm_cmpart_softwareimages.all.images[0].uuid
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }

  roles {
    name       = %[5]q
    child_type = %[6]q
  }
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        roleName,
        childType,
    )
}
```

## Verification Checklist

- [ ] `TestAccCMDeviceCategory_RolesUUIDPopulated` passes
- [ ] `TestAccCMDeviceCategory_RolesIdempotency` passes
- [ ] All existing `CMDeviceCategory` tests pass
- [ ] `make lint` passes
- [ ] `make generate` completes without changes to docs

## Environment Setup

```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

## Key Files

| File | Purpose |
|------|---------|
| `internal/provider/resource_cmdevice_category.go` | Resource implementation (fix location) |
| `internal/provider/resource_cmdevice_category_test.go` | Test file (add new tests) |
| `specs/083-roles-uuid-computed/plan.md` | Full implementation plan |
| `specs/083-roles-uuid-computed/spec.md` | Feature specification |
