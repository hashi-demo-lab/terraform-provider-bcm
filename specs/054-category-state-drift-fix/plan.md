# Implementation Plan: Category Resource State Drift Fix

## Issue Reference
- **GitHub Issue**: [#54](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/54)
- **Specification**: `specs/054-category-state-drift-fix/spec.md`

## Executive Summary

This plan addresses the category resource state drift issue where computed fields show unexpected `(known after apply)` changes after resource creation. The fix involves:

1. Adding `UseStateForUnknown` plan modifiers to computed fields
2. Preserving `software_image_proxy` values from state in the Read function
3. Fixing test configurations to use deterministic UUID lookups

## Prerequisites

- Working BCM cluster connection (existing)
- Understanding of Terraform Plugin Framework plan modifiers
- Reference implementation: `resource_cmpart_softwareimage.go` (lines 104-112)

## Implementation Phases

### Phase 1: Add Plan Modifiers to Schema

**File**: `internal/provider/resource_cmdevice_category.go`

**Imports to Add** (lines 14-22):
```go
import (
    // ... existing imports ...
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)
```

**Fields to Update with `UseStateForUnknown`**:

| Field | Type | Line | Plan Modifier |
|-------|------|------|---------------|
| `id` | String | 190-193 | `stringplanmodifier.UseStateForUnknown()` |
| `uuid` | String | 194-197 | `stringplanmodifier.UseStateForUnknown()` |
| `base_type` | String | 560-563 | `stringplanmodifier.UseStateForUnknown()` |
| `child_type` | String | 564-567 | `stringplanmodifier.UseStateForUnknown()` |
| `parent_uuid` | String | 544-547 | `stringplanmodifier.UseStateForUnknown()` |
| `revision` | String | 548-551 | `stringplanmodifier.UseStateForUnknown()` |
| `modified` | Bool | 552-555 | `boolplanmodifier.UseStateForUnknown()` |
| `to_be_removed` | Bool | 556-559 | `boolplanmodifier.UseStateForUnknown()` |

**Nested Object Fields** (`software_image_proxy`):

| Field | Type | Line | Plan Modifier |
|-------|------|------|---------------|
| `software_image_proxy.uuid` | String | 223-226 | `stringplanmodifier.UseStateForUnknown()` |
| `software_image_proxy.revision_id` | Int64 | 230-233 | `int64planmodifier.UseStateForUnknown()` |

### Phase 2: Preserve `software_image_proxy` in Read

**Pattern Reference**: See `resource_cmpart_softwareimage.go` lines 466-482 for `original_image` preservation pattern.

**Implementation in `readCategory()` function** (after line 1548):

```go
// Preserve software_image_proxy from prior state if configured
// BCM may return different UUID references on subsequent reads
if !model.SoftwareImageProxy.IsNull() && !model.SoftwareImageProxy.IsUnknown() {
    // Get the original parent_software_image from state
    var originalProxy SoftwareImageProxyModel
    model.SoftwareImageProxy.As(ctx, &originalProxy, basetypes.ObjectAsOptions{})

    // After reading from BCM, restore the user-configured parent_software_image
    // if the proxy UUID matches (same object, just different parent reference)
    if !originalProxy.ParentSoftwareImage.IsNull() && !originalProxy.ParentSoftwareImage.IsUnknown() {
        // Preserve user's configured value
        // BCM returns the proxy, but parent_software_image may differ
    }
}
```

**Alternative**: Modify the Read function to preserve `software_image_proxy` from state entirely, similar to `force` parameter handling.

### Phase 3: Fix Test Configuration

**File**: `internal/provider/resource_cmdevice_category_test.go`

**Problem**: Current config uses `data.bcm_cmpart_softwareimages.all.images[0].uuid` which can change between test steps.

**Solution Options**:

#### Option A: Use a Named Filter (Preferred)
```hcl
# Filter for a specific software image by name pattern
data "bcm_cmpart_softwareimages" "default" {
  # Use first available or filter by name
}

locals {
  # Store the UUID once during initial evaluation
  software_image_uuid = data.bcm_cmpart_softwareimages.default.images[0].uuid
}
```

#### Option B: Self-Reference in Update Config
Use the created resource's own state for subsequent operations (avoids data source re-evaluation).

#### Option C: Accept BCM's UUID
If BCM legitimately changes the parent_software_image reference, the test should accept this behavior. This would mean the test expectation is wrong.

**Recommended**: Option A with additional investigation into BCM API behavior.

## Implementation Order

1. **Phase 1**: Add plan modifiers to schema (prevents `(known after apply)` for computed fields)
2. **Phase 2**: Preserve `software_image_proxy` in Read (prevents parent_software_image drift)
3. **Phase 3**: Fix test configuration (ensures deterministic test behavior)
4. **Validation**: Run acceptance tests to verify fix

## Code Changes Summary

### `resource_cmdevice_category.go`

```go
// 1. Add imports (after line 22)
import (
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/boolplanmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
)

// 2. Add plan modifiers to schema attributes
"id": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Resource identifier (same as UUID)",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
"uuid": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Unique identifier assigned by BCM",
    PlanModifiers: []planmodifier.String{
        stringplanmodifier.UseStateForUnknown(),
    },
},
// ... similar for other computed fields
```

### `resource_cmdevice_category_test.go`

Investigate whether test config issue is causing the parent_software_image UUID change, or if it's BCM API behavior.

## Testing Strategy

### Unit Tests
- Verify schema has correct plan modifiers
- Verify Read preserves configured values

### Acceptance Tests
Run existing tests after changes:
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 30m ./internal/provider/ -run "^TestAccCMDeviceCategoryResource_Basic$"
```

### Full Test Suite
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"
```

## Success Criteria

1. ✅ `TestAccCMDeviceCategoryResource_Basic` passes all 5 steps
2. ✅ Idempotency check (Step 2) shows empty plan
3. ✅ Idempotency check after Update (Step 5) shows empty plan
4. ✅ Import preserves all field values
5. ✅ All 16 existing category tests pass
6. ✅ `make lint` passes
7. ✅ `make test` passes

## Rollback Plan

If the fix causes regressions:
1. Revert plan modifier additions
2. Keep original `readCategory()` behavior
3. Investigate BCM API behavior further

## Dependencies

- No external dependencies
- No breaking changes to existing configurations
- Backward compatible with existing state files

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| Plan modifiers cause other test failures | Low | Medium | Run full test suite before merging |
| BCM API behavior changes | Low | High | Document expected API behavior |
| State migration issues | Very Low | Low | Plan modifiers don't affect stored state |

## Timeline Estimate

- Phase 1 (Plan Modifiers): 30 minutes
- Phase 2 (Read Preservation): 30 minutes
- Phase 3 (Test Fix): 30 minutes
- Testing & Validation: 60 minutes
- **Total**: ~2.5 hours
