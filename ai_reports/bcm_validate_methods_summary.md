# BCM Validate Methods - Complete Discovery

**Date**: 2024-11-24
**Author**: Claude Code
**Scope**: Complete investigation of all BCM API validation methods

---

## Summary

✅ **5 validation methods discovered across 4 BCM services**

| Service | Method | Resource Type | Provider Resource | Status |
|---------|--------|---------------|-------------------|--------|
| `CMPart` | `validateSoftwareImage` | Software Images | `resource_cmpart_softwareimage.go` | ✅ TESTED |
| `CMDevice` | `validateCategory` | Device Categories | `resource_cmdevice_category.go` | ✅ TESTED |
| `CMDevice` | `validateDevice` | Devices/Nodes | `resource_cmdevice_device.go` | ✅ TESTED |
| `CMNet` | `validateNetwork` | Networks | `resource_cmnet_network.go` (if exists) | ✅ TESTED |
| `cmkube` | `validateKubeCluster` | Kubernetes Clusters | `resource_cmkube_cluster.go` | ✅ TESTED |

---

## Test Results

### 1. validateSoftwareImage (CMPart)
- **Service**: `CMPart` (case-sensitive)
- **Status**: ✅ Works
- **Success Response**: `[]`
- **Error Example**: Invalid SOL speed, duplicate names
- **Use Case**: Software image validation before create/update

### 2. validateCategory (CMDevice)
- **Service**: `CMDevice` (case-sensitive)
- **Status**: ✅ Works
- **Success Response**: `[]`
- **Use Case**: Category validation before create/update

### 3. validateDevice (CMDevice)
- **Service**: `CMDevice` (case-sensitive)
- **Status**: ✅ Works
- **Success Response**: `[]`
- **Error Example**: Invalid hostname format
- **Error Message**: *"The hostname can only contain a-z, A-Z, 0-9 and dashes (-). It can not start or end with a dash, or contain only numbers."*
- **Use Case**: Device/node validation before create/update

### 4. validateNetwork (CMNet)
- **Service**: `CMNet` (case-sensitive)
- **Status**: ✅ Works
- **Success Response**: `[]`
- **Use Case**: Network validation before create/update

### 5. validateKubeCluster (cmkube)
- **Service**: `cmkube` (lowercase!)
- **Method**: `validateKubeCluster` (NOT `validateCluster`)
- **Status**: ✅ Works
- **Success Response**: `[]`
- **Error Example**: Zero UUID, invalid cluster configuration
- **Use Case**: Kubernetes cluster validation before create/update
- **Important**: Service name is **lowercase** `cmkube`, not `CMKube`

---

## Key Findings

### Service Name Case Sensitivity

⚠️ **IMPORTANT**: BCM API services use different casing conventions:

| Service | Correct Case | Common Mistake |
|---------|--------------|----------------|
| Software Images | `CMPart` | `cmpart` |
| Categories | `CMDevice` | `cmdevice` |
| Devices | `CMDevice` | `cmdevice` |
| Networks | `CMNet` | `cmnet` |
| Kubernetes | `cmkube` | `CMKube` ❌ |

The Kubernetes service uses **lowercase** `cmkube`, which is inconsistent with other services!

### Validation Response Format

**All validation methods follow the same pattern:**

**Success (no errors):**
```json
[]
```

**Errors/Warnings:**
```json
[
  {
    "baseType": "Validation",
    "error_code": "BAD_VALUE",
    "field": "hostname",
    "message": "The hostname can only contain...",
    "severity": "ERROR",
    "ref_entity_uuid": "9f885869-a146-4cd6-af1f-f9b6c674a84c"
  }
]
```

**Severity Levels:**
- `ERROR` - Must be fixed (operation will fail)
- `WARNING` - Advisory (operation may succeed)

---

## Implementation Recommendation

### Generic Helper Function

Create a **single reusable helper** in `bcm_client.go`:

```go
// ValidationError represents a BCM validation error
type ValidationError struct {
    Field      string
    Message    string
    ErrorCode  string
    Severity   string
    EntityUUID string
}

// ValidateEntity performs pre-flight validation for any BCM entity
// Returns nil error and empty slice if validation passes
// Returns validation errors if validation fails
func (c *BCMClient) ValidateEntity(
    ctx context.Context,
    service string,        // "CMPart", "CMDevice", "cmkube", etc.
    validateMethod string, // "validateSoftwareImage", "validateDevice", etc.
    entity map[string]interface{},
) ([]ValidationError, error) {
    body, err := c.CallJSONRPC(ctx, service, validateMethod, entity)
    if err != nil {
        return nil, err
    }

    var validationResults []map[string]interface{}
    if err := json.Unmarshal(body, &validationResults); err != nil {
        return nil, fmt.Errorf("failed to parse validation response: %w", err)
    }

    if len(validationResults) == 0 {
        return nil, nil // Validation passed
    }

    // Convert to structured errors
    var errors []ValidationError
    for _, v := range validationResults {
        errors = append(errors, ValidationError{
            Field:      getString(v, "field"),
            Message:    getString(v, "message"),
            ErrorCode:  getString(v, "error_code"),
            Severity:   getString(v, "severity"),
            EntityUUID: getString(v, "ref_entity_uuid"),
        })
    }

    return errors, nil
}
```

### Usage Across All Resources

**Software Image:**
```go
validationErrors, err := r.client.ValidateEntity(ctx, "CMPart", "validateSoftwareImage", entity)
```

**Category:**
```go
validationErrors, err := r.client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity)
```

**Device:**
```go
validationErrors, err := r.client.ValidateEntity(ctx, "CMDevice", "validateDevice", entity)
```

**Network:**
```go
validationErrors, err := r.client.ValidateEntity(ctx, "CMNet", "validateNetwork", entity)
```

**Kubernetes Cluster:**
```go
validationErrors, err := r.client.ValidateEntity(ctx, "cmkube", "validateKubeCluster", entity)
```

### Error Handling Pattern

```go
if len(validationErrors) > 0 {
    for _, ve := range validationErrors {
        if ve.Severity == "ERROR" {
            resp.Diagnostics.AddError(
                fmt.Sprintf("Validation Error: %s", ve.Field),
                ve.Message,
            )
        } else if ve.Severity == "WARNING" {
            resp.Diagnostics.AddWarning(
                fmt.Sprintf("Validation Warning: %s", ve.Field),
                ve.Message,
            )
        }
    }
    return // Stop execution on validation errors
}
```

---

## Resources Requiring Updates

### Phase 1: Add Generic Helper
- [ ] `internal/provider/bcm_client.go`
  - [ ] Add `ValidationError` struct
  - [ ] Add `ValidateEntity()` method
  - [ ] Add helper `getString()` if not exists

### Phase 2: Integrate into Resources
- [ ] `internal/provider/resource_cmpart_softwareimage.go`
  - Line 483: Add validation before `updateSoftwareImage`

- [ ] `internal/provider/resource_cmdevice_category.go`
  - Add validation before `updateCategory`

- [ ] `internal/provider/resource_cmdevice_device.go`
  - Line 783: Add validation before `updateDevice`

- [ ] `internal/provider/resource_cmkube_cluster.go`
  - Line 569: Add validation before `updateKubeCluster`
  - **Note**: Use `"cmkube"` (lowercase) as service name!

- [ ] `internal/provider/resource_cmnet_network.go` (if exists)
  - Add validation before update operation

### Phase 3: Testing
- [ ] Unit tests for `ValidateEntity()` helper
- [ ] Integration tests for validation errors
- [ ] Test severity handling (ERROR vs WARNING)
- [ ] Test all 5 resources with validation

---

## Benefits

1. ✅ **Consistent validation** across all resource types
2. ✅ **Better error messages** before operations fail
3. ✅ **Fail fast** - catch errors before API calls
4. ✅ **Single source of truth** - one helper function
5. ✅ **Proper severity handling** - distinguish ERROR from WARNING
6. ✅ **Maintainable** - easy to extend to new resources

---

## Known Limitations

1. **CREATE Operations**: Validation often returns "Zero UUID" errors for new entities
   - **Recommendation**: Only use validation for UPDATE operations

2. **Performance**: Adds 1 extra API call per UPDATE (~50-200ms)
   - **Trade-off**: Worth it for better UX

3. **Service Name Inconsistency**: `cmkube` uses lowercase while others use CamelCase
   - **Mitigation**: Document clearly in code comments

---

## Test Scripts Created

All test scripts are in `/workspace/sampleRest/`:

1. `test_validate_softwareimage.py` - Software Image validation
2. `test_validate_comprehensive.py` - Comprehensive software image tests
3. `validate_methods_with_entities.py` - Multi-resource validation test
4. `test_validate_device.py` - Device validation
5. `test_validate_cluster.py` - Cluster validation discovery
6. `test_validate_kubecluster_detailed.py` - Detailed cluster validation

---

## Next Steps

1. Review and approve generic helper function design
2. Implement `ValidateEntity()` in `bcm_client.go`
3. Update all 5 resources to use validation
4. Add acceptance tests for validation behavior
5. Update CLAUDE.md with validation patterns
6. Update GitHub issue #51 with cluster validation info

---

## Related Files

- **Analysis Report**: `/workspace/ai_reports/bcm_validate_methods_analysis.md`
- **GitHub Issue**: https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/51
- **Provider Resources**:
  - `/workspace/internal/provider/resource_cmpart_softwareimage.go`
  - `/workspace/internal/provider/resource_cmdevice_category.go`
  - `/workspace/internal/provider/resource_cmdevice_device.go`
  - `/workspace/internal/provider/resource_cmkube_cluster.go`
  - `/workspace/internal/provider/resource_cmnet_network.go` (if exists)
