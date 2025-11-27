# BCM Category List Fields - Developer Quick Start

**Feature Branch**: `073-category-list-fields`
**GitHub Issue**: #73
**Last Updated**: 2025-11-27

## Background

BCM API accepts certain list fields on category create/update operations but returns empty arrays when reading the category back. The Terraform provider implements a workaround to preserve user-configured values in state.

## Affected Fields

| Field | Terraform Attribute | BCM API Field | Data Type |
|-------|---------------------|---------------|-----------|
| Static Routes | `static_routes` | `staticRoutes` | `[]StaticRoute` |
| FS Exports | `fsexports` | `fsexports` | `[]FSExport` |
| Roles | `roles` | `roles` | `[]Role` |
| GPU Settings | `gpu_settings` | `gpuSettings` | `[]GPUSetting` |
| Services | `services` | `services` | `[]Service` |

## Provider Behavior

### Create/Update Flow

```
User Config --> Terraform Plan --> Provider sends to BCM API
                                          |
                                   (BCM accepts but doesn't persist)
                                          |
                              Provider reads category back
                                          |
                              BCM returns empty arrays []
                                          |
                         Provider preserves plan values in state
                                          |
                              State matches user config (no drift)
```

### Read Flow

1. Provider calls `getCategory(name)` from BCM API
2. BCM returns category with empty arrays for affected fields
3. Provider preserves existing state values (not API values)
4. No false drift detection

### Import Flow

1. User runs `terraform import bcm_cmdevice_category.name "uuid"`
2. Provider reads category from BCM API
3. Affected fields are empty in imported state
4. User must re-apply config to restore values in state

## Code References

### Workaround Implementation

**File**: `internal/provider/resource_cmdevice_category.go`

```go
// Lines 1051-1054 - Restore plan values after Create
plan.StaticRoutes = planStaticRoutes
plan.FSExports = planFSExports
plan.Roles = planRoles
plan.GPUSettings = planGPUSettings
plan.Services = planServices
```

### Role UUID Generation

**File**: `internal/provider/resource_cmdevice_category.go`

```go
// Lines 2736-2748 - Generate UUIDs for roles since BCM doesn't persist them
if origRole.UUID.IsNull() || origRole.UUID.IsUnknown() || origRole.UUID.ValueString() == "" {
    newUUID := generateUUID()
    mergedRole.UUID = types.StringValue(newUUID)
    tflog.Debug(ctx, "Generated UUID for role (BCM doesn't persist category roles)", ...)
}
```

### Import State Verification

**File**: `internal/provider/resource_cmdevice_category_test.go`

```go
// Lines 2596-2603 - Ignore non-persisted fields during import verification
ImportStateVerifyIgnore: []string{
    "force",
    "static_routes",
    "fsexports",
    "roles",
    "gpu_settings",
    "services",
},
```

## Testing These Fields

### Run Specific Tests

```bash
# Static routes test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMDeviceCategoryResource_StaticRoutes"

# Filesystem exports test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMDeviceCategoryResource_FilesystemExports"

# Roles test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMDeviceCategoryResource_Roles"

# All list fields combined
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMDeviceCategoryResource_AllListFields"
```

### Verify Idempotency

Tests must pass idempotency checks - no plan changes after apply:

```go
ConfigPlanChecks: resource.ConfigPlanChecks{
    PreApply: []plancheck.PlanCheck{
        plancheck.ExpectEmptyPlan(),
    },
},
```

### Run Investigation Script

```bash
cd /workspace/sampleRest
python3 investigate_category_list_fields.py
```

## Debugging

### Enable Debug Logging

```bash
export TF_LOG=DEBUG
terraform apply
```

Look for log messages like:
- "Restoring plan values for list fields BCM doesn't persist"
- "Generated UUID for role (BCM doesn't persist category roles)"

### Verify State Values

```bash
terraform show -json | jq '.values.root_module.resources[] | select(.type=="bcm_cmdevice_category") | .values.static_routes'
```

## Common Issues

### Issue: Import shows empty values for list fields

**Cause**: BCM doesn't persist these fields, so import reads empty arrays.

**Solution**: After import, run `terraform apply` to restore configured values in state.

### Issue: Plan shows changes after import

**Cause**: User config has list field values, but imported state is empty.

**Solution**: This is expected. Apply the changes to sync state with config.

### Issue: Role UUIDs change on each apply

**Cause**: Provider generates UUIDs locally when BCM returns none.

**Solution**: This is expected behavior. UUIDs are for state tracking only.

## Related Files

| File | Purpose |
|------|---------|
| `internal/provider/resource_cmdevice_category.go` | Resource implementation with workarounds |
| `internal/provider/resource_cmdevice_category_test.go` | Acceptance tests |
| `docs/resources/cmdevice_category.md` | User documentation |
| `sampleRest/investigate_category_list_fields.py` | API investigation script |
| `specs/073-category-list-fields/research.md` | Investigation findings |

## References

- GitHub Issue: #73
- Spec: `/workspace/specs/073-category-list-fields/spec.md`
- Plan: `/workspace/specs/073-category-list-fields/plan.md`
- Research: `/workspace/specs/073-category-list-fields/research.md`
