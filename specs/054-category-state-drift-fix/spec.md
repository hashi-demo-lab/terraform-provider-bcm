# Specification: Category Resource State Drift Fix

## Issue Reference
- **GitHub Issue**: [#54](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/54)
- **Title**: Category resource state drift: Computed fields showing unexpected changes after create
- **Priority**: High

## Problem Statement

The `TestAccCMDeviceCategoryResource_Basic` test fails with a non-empty plan after resource creation (Step 2 - idempotency check). Multiple computed fields show drift when they should remain stable:

### Error Output
```
Error: After applying this test step, the non-refresh plan was not empty.

# bcm_cmdevice_category.test will be updated in-place
~ resource "bcm_cmdevice_category" "test" {
    ~ base_type              = "Category" -> (known after apply)
    ~ boot_loader            = "SYSLINUX" -> (known after apply)
    + boot_loader_file       = (known after apply)
    ~ boot_loader_protocol   = "HTTP" -> (known after apply)
    + child_type             = (known after apply)
    ~ default_gateway        = "0.0.0.0" -> (known after apply)
    ~ default_gateway_metric = 0 -> (known after apply)
    ~ id                     = "aaf6f307..." -> (known after apply)
    ~ install_boot_record    = false -> (known after apply)
    ~ install_mode           = "AUTO" -> (known after apply)
    ~ modified               = false -> (known after apply)
      name                   = "tftest-category-..."
    ~ new_node_install_mode  = "FULL" -> (known after apply)
    ~ parent_uuid            = "aaf6f307..." -> (known after apply)
    + revision               = (known after apply)
    ~ software_image_proxy   = {
        ~ parent_software_image = "8482c4e9..." -> "49ef004e..."
        ~ revision_id           = -1 -> (known after apply)
        ~ uuid                  = "2bccf924..." -> (known after apply)
      }
    ~ to_be_removed          = false -> (known after apply)
    ~ uuid                   = "aaf6f307..." -> (known after apply)
}
```

## Root Cause Analysis

### Primary Issue: `software_image_proxy.parent_software_image` UUID Instability

The test configuration uses a data source reference that may return different values:
```hcl
local.software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0
  ? data.bcm_cmpart_softwareimages.all.images[0].uuid
  : "00000000-0000-0000-0000-000000000000"
```

**Problem**: If the list order of software images changes between Create and subsequent plan operations, `images[0]` returns a different UUID, causing:
1. Config says `parent_software_image = "new-uuid"`
2. State has `parent_software_image = "original-uuid"`
3. Terraform sees a difference and plans an update

### Secondary Issue: Computed Fields Showing "(known after apply)"

All computed fields show `(known after apply)` because the schema marking is inconsistent with the Read behavior:

1. **Schema Definition**: Fields like `base_type`, `boot_loader`, `id`, `uuid` are marked as `Computed: true`
2. **Read Behavior**: `readCategory()` correctly populates these fields
3. **Issue**: When Terraform compares plan vs state for Optional+Computed fields, if the plan doesn't explicitly set a value but the schema marks it as Computed, Terraform may mark it as "known after apply" for subsequent plans

### Test Configuration Issue

The test uses dynamic data source lookups that can produce non-deterministic values:
- `data.bcm_cmdevice_categories.all.categories[0].management_network_id`
- `data.bcm_cmpart_softwareimages.all.images[0].uuid`

## Affected Files

1. **Primary**: `internal/provider/resource_cmdevice_category.go`
   - `readCategory()` function
   - Schema definition for `software_image_proxy`

2. **Test**: `internal/provider/resource_cmdevice_category_test.go`
   - `testAccCMDeviceCategoryResourceConfig()` function
   - Dynamic UUID lookup pattern

## Requirements

### R1: Stable `software_image_proxy.parent_software_image` After Create
The `parent_software_image` field should remain stable in state after creation. The Read operation should preserve the user's configured value rather than reading potentially different values from BCM.

### R2: Computed Fields Should Not Show Drift
All purely computed fields (`uuid`, `id`, `base_type`, `child_type`, `modified`, `to_be_removed`, `revision`, `parent_uuid`) should remain stable after initial population.

### R3: Optional+Computed Fields Should Not Show False Drift
Fields marked as Optional+Computed (`boot_loader`, `boot_loader_file`, `boot_loader_protocol`, `install_mode`, `new_node_install_mode`, `install_boot_record`, `default_gateway`, `default_gateway_metric`) should:
- Accept user configuration if provided
- Use BCM defaults if not provided
- Not show drift when user doesn't specify a value

### R4: Test Configuration Should Be Deterministic
Test configurations should produce consistent results across test runs by avoiding non-deterministic data source lookups.

## Technical Solution

### Solution 1: Preserve `software_image_proxy` from State in Read
In `readCategory()`, preserve the `software_image_proxy.parent_software_image` from prior state if:
1. The field was set in configuration
2. BCM API returns the same proxy UUID (indicating same proxy object)

**Pattern Reference**: See `resource_cmpart_softwareimage.go` lines 466-482 for `original_image` preservation pattern.

### Solution 2: Use UseStateForUnknown Plan Modifier
For purely computed fields that should never change after create, add `UseStateForUnknown` plan modifier:
```go
"uuid": schema.StringAttribute{
    Computed:            true,
    PlanModifiers:       []planmodifier.String{stringplanmodifier.UseStateForUnknown()},
}
```

### Solution 3: Fix Test Configuration
Capture the software image UUID during resource creation step and reuse it:
```hcl
# Instead of: data.bcm_cmpart_softwareimages.all.images[0].uuid
# Use: bcm_cmdevice_category.test.software_image_proxy.parent_software_image
```

Or use a specific known software image instead of `images[0]`.

## Success Criteria

1. **SC1**: `TestAccCMDeviceCategoryResource_Basic` passes all steps including idempotency check
2. **SC2**: No spurious drift for computed fields after Create
3. **SC3**: No spurious drift for computed fields after Update
4. **SC4**: Import operation preserves all field values correctly
5. **SC5**: All 16 existing category tests pass

## Out of Scope

- Changing the schema structure significantly
- BCM API behavior changes
- Other resource implementations

## API Contract

### BCM API: `getCategory(name)`
- **Input**: Category name string
- **Output**: Category entity with all fields populated
- **Behavior**: Returns current values from BCM, which may differ from original configuration if BCM applies defaults or transforms values

### Key Field Mappings
| Terraform Field | BCM API Field | Notes |
|----------------|---------------|-------|
| `software_image_proxy.parent_software_image` | `softwareImageProxy.parentSoftwareImage` | UUID reference to software image |
| `software_image_proxy.uuid` | `softwareImageProxy.uuid` | BCM-assigned proxy object UUID |
| `base_type` | `baseType` | Always "Category" |
| `management_network` | `managementNetwork` | UUID reference to network |

## Testing Strategy

1. **Unit Test**: Verify `readCategory()` preserves configured values
2. **Acceptance Test**: Verify idempotency after Create, Update, and Import
3. **Drift Test**: Verify external modifications are detected but don't cause false positives
