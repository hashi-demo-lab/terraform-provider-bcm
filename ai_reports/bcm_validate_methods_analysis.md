# BCM API Validate Methods Analysis

**Date**: 2024-11-24
**Scope**: Investigation of BCM API validation capabilities across all implemented provider resources

---

## Executive Summary

The BCM API provides **validate methods for all major resource types**. These methods perform server-side validation before CRUD operations and can significantly improve error messaging and user experience.

### Discovered Validate Methods

| Service | Method | Resource Type | Status |
|---------|--------|---------------|--------|
| `CMPart` | `validateSoftwareImage` | Software Images | ✅ **EXISTS** |
| `CMDevice` | `validateCategory` | Device Categories | ✅ **EXISTS** |
| `CMDevice` | `validateDevice` | Devices/Nodes | ✅ **EXISTS** |
| `CMNet` | `validateNetwork` | Networks | ✅ **EXISTS** |

---

## Validation Behavior

### Request Format

```json
{
  "service": "<service-name>",
  "call": "validate<ResourceType>",
  "args": [<entity-object>]
}
```

### Response Format

**Success (no validation errors):**
```json
[]
```

**Validation Errors:**
```json
[
  {
    "baseType": "Validation",
    "childType": "",
    "error_code": "BAD_VALUE",
    "field": "SOLSpeed",
    "message": "Illegal value for: SOLSpeed",
    "severity": "ERROR",
    "ref_entity_uuid": "8482c4e9-383c-43de-873f-8c54ee77ee74",
    "uuid": "421301aa-aa4e-5f69-84c8-d61f6a449ed5"
  }
]
```

**Severity Levels:**
- `ERROR` - Must be fixed before operation can succeed
- `WARNING` - Advisory messages (operation may still succeed)

---

## Common Validation Error Codes

| Error Code | Description | Example |
|------------|-------------|---------|
| `BAD_VALUE` | Invalid field value | Invalid SOL speed, zero UUID |
| `DUPLICATE_FIELD` | Unique constraint violation | Image name already exists |
| `NOT_SET` | Missing required field | Kernel version not set |
| `BAD_REFERENCE` | Invalid foreign key | Invalid partition UUID |

---

## Test Results

### Test 1: validateSoftwareImage (Valid Entity)

**Request:**
```json
{
  "service": "CMPart",
  "call": "validateSoftwareImage",
  "args": [
    {
      "uuid": "8482c4e9-383c-43de-873f-8c54ee77ee74",
      "name": "default-image",
      "path": "/cm/images/default-image.iso",
      ...
    }
  ]
}
```

**Response:** `[]` (validation passed)

---

### Test 2: validateSoftwareImage (Invalid SOL Speed)

**Modified Entity:**
```json
{
  "SOLSpeed": "999999"  // Invalid baud rate
}
```

**Response:**
```json
[
  {
    "error_code": "BAD_VALUE",
    "field": "SOLSpeed",
    "message": "Illegal value for: SOLSpeed",
    "severity": "ERROR"
  }
]
```

---

### Test 3: validateSoftwareImage (Duplicate Name)

**New Entity with Existing Name:**
```json
{
  "name": "default-image"  // Already exists
}
```

**Response:**
```json
[
  {
    "error_code": "DUPLICATE_FIELD",
    "field": "name",
    "message": "A softwareimage with that name already exists",
    "severity": "ERROR"
  }
]
```

---

## Known Limitations

### 1. New Entity Validation (No UUID)

When validating a **new entity** without a UUID:

**Error:**
```json
{
  "error_code": "BAD_VALUE",
  "field": "uuid",
  "message": "Zero UUID SoftwareImage:name",
  "severity": "ERROR"
}
```

**Implication:** `validateSoftwareImage` is **more useful for UPDATE operations** than CREATE operations, as new entities don't yet have UUIDs.

### 2. Path Validation

**Warning:**
```json
{
  "error_code": "BAD_VALUE",
  "field": "path",
  "message": "The software image path does not exist",
  "severity": "WARNING"
}
```

**Note:** This is often a WARNING, not ERROR, as BCM may allow creation even if path doesn't exist yet.

---

## Implementation Recommendations

### Option 1: Resource-Specific Validation (Current Approach)

Each resource implements its own validation logic:

```go
// In resource_cmpart_softwareimage.go
func (r *CMPartSoftwareImageResource) validateEntity(ctx, entity) error {
    body, err := r.client.CallJSONRPC(ctx, "CMPart", "validateSoftwareImage", entity)
    // Parse validation errors
}

// In resource_cmdevice_category.go
func (r *CMDeviceCategoryResource) validateEntity(ctx, entity) error {
    body, err := r.client.CallJSONRPC(ctx, "CMDevice", "validateCategory", entity)
    // Parse validation errors
}
```

**Pros:**
- Simple and straightforward
- No abstraction overhead
- Clear ownership per resource

**Cons:**
- Code duplication
- Inconsistent error formatting
- Harder to maintain

---

### Option 2: Generic Helper Function (Recommended)

Create a shared validation helper in `bcm_client.go`:

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
// Returns empty slice if validation passes, or slice of ValidationErrors
func (c *BCMClient) ValidateEntity(ctx context.Context, service, validateMethod string, entity map[string]interface{}) ([]ValidationError, error) {
    body, err := c.CallJSONRPC(ctx, service, validateMethod, entity)
    if err != nil {
        return nil, err
    }

    // Empty array means validation passed
    var validationResults []map[string]interface{}
    if err := json.Unmarshal(body, &validationResults); err != nil {
        return nil, fmt.Errorf("failed to parse validation response: %w", err)
    }

    if len(validationResults) == 0 {
        return nil, nil // Validation passed
    }

    // Parse validation errors
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

**Usage in Resources:**

```go
// In Update() method
func (r *CMPartSoftwareImageResource) Update(ctx, req, resp) {
    entity := r.buildAPIEntity(&plan, plan.UUID.ValueString())

    // Pre-flight validation
    validationErrors, err := r.client.ValidateEntity(
        ctx,
        "CMPart",
        "validateSoftwareImage",
        entity,
    )

    if err != nil {
        resp.Diagnostics.AddError("Validation Failed", err.Error())
        return
    }

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
        return
    }

    // Proceed with update
    _, err = r.client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, false)
    // ...
}
```

**Pros:**
- ✅ Single source of truth for validation logic
- ✅ Consistent error handling across all resources
- ✅ Easy to maintain and extend
- ✅ Proper handling of ERROR vs WARNING severity
- ✅ Reusable across all resources (Software Image, Category, Network)

**Cons:**
- Slightly more upfront implementation effort

---

## Implementation Strategy

### Phase 1: Create Generic Helper (Recommended)

1. Add `ValidationError` struct to `bcm_client.go`
2. Add `ValidateEntity()` method to `BCMClient`
3. Add helper function `getString()` for null-safe field extraction

### Phase 2: Integrate into UPDATE Operations

Update these resources to use pre-flight validation:

1. ✅ `resource_cmpart_softwareimage.go` - **UPDATE operation**
2. ✅ `resource_cmdevice_category.go` - **UPDATE operation**
3. ✅ `resource_cmnet_network.go` - **UPDATE operation** (if implemented)

**Skip CREATE operations** due to "zero UUID" validation errors for new entities.

### Phase 3: Enhance Error Messages

Convert validation errors to user-friendly Terraform diagnostics:

- `ERROR` severity → `resp.Diagnostics.AddError()`
- `WARNING` severity → `resp.Diagnostics.AddWarning()`
- Include field name and validation message
- Log full validation response for debugging

---

## Test Coverage Recommendations

Add acceptance tests for validation behavior:

### Test: Update with Invalid Field Value

```go
func TestAccCMPartSoftwareImage_ValidationInvalidSOLSpeed(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: testAccSoftwareImageConfig("test", "999999"), // Invalid SOL speed
                ExpectError: regexp.MustCompile("Validation Error.*SOLSpeed"),
            },
        },
    })
}
```

### Test: Update with Duplicate Name

```go
func TestAccCMPartSoftwareImage_ValidationDuplicateName(t *testing.T) {
    // Create image1
    // Try to update image2 to have same name as image1
    // Expect validation error
}
```

---

## Performance Considerations

**Validation API Call Overhead:**
- Adds 1 extra API call per UPDATE operation
- Network latency: ~50-200ms
- Validation processing: negligible

**Mitigation:**
- Only validate on UPDATE (not CREATE or READ)
- Validation is lightweight compared to actual update operation
- Better UX (catch errors before applying changes)

**Trade-off Analysis:**
- ✅ **Benefit:** Better error messages, faster failure detection
- ⚠️ **Cost:** 1 additional API round-trip per update
- ✅ **Verdict:** Worth it for production use

---

## Current Status in Codebase

### Software Image Resource
- **Location:** `internal/provider/resource_cmpart_softwareimage.go`
- **Current State:** Validation **NOT implemented**
- **Comment (line 472-473):**
  ```go
  // NOTE: Pre-flight validation skipped - will rely on server-side validation
  // Optional enhancement: Add validateSoftwareImage call here for better error messages
  ```

### Category Resource
- **Location:** `internal/provider/resource_cmdevice_category.go`
- **Current State:** Validation **NOT implemented**

### Device Resource
- **Location:** `internal/provider/resource_cmdevice_device.go`
- **Current State:** Validation **NOT implemented**
- **Update method:** Line 783 - `updateDevice` call (no pre-flight validation)
- **Validation method available:** `CMDevice.validateDevice`

### Network Resource
- **Status:** May not be implemented yet (need to verify)

---

## Device Validation Test Results

### Test: validateDevice (Valid Entity)

**Entity:** Existing device with hostname "unknown"

**Response:** `[]` (validation passed)

### Test: validateDevice (Invalid Hostname)

**Modified Entity:**
```json
{
  "hostname": "",  // Empty hostname
  "modified": true
}
```

**Response:**
```json
[
  {
    "baseType": "Validation",
    "error_code": "BAD_VALUE",
    "field": "hostname",
    "message": "The hostname can only contain a-z, A-Z, 0-9 and dashes (-). It can not start or end with a dash, or contain only numbers.",
    "severity": "ERROR",
    "ref_entity_uuid": "9f885869-a146-4cd6-af1f-f9b6c674a84c"
  }
]
```

**Key Findings:**
- ✅ `CMDevice.validateDevice` exists and works
- ✅ Catches invalid hostname formats
- ✅ Returns structured ERROR validation messages
- ✅ Can be used for pre-flight validation in Update() operation

---

## Conclusion

**Key Findings:**
1. ✅ BCM API supports validation for all major resource types (4 methods discovered)
2. ✅ Validation returns structured error/warning messages
3. ✅ `validateDevice` successfully validates device/node entities
4. ✅ Can significantly improve error messages before CRUD operations
5. ⚠️ More useful for UPDATE than CREATE (due to UUID requirement)
6. ✅ Generic helper function recommended for code reuse

**Recommendation:**
Implement **Option 2 (Generic Helper Function)** in `bcm_client.go` and integrate into all UPDATE operations across Software Image, Category, Device, and Network resources.

**Next Steps:**
1. Create `ValidateEntity()` helper in `bcm_client.go`
2. Update `resource_cmpart_softwareimage.go` Update() method
3. Update `resource_cmdevice_category.go` Update() method
4. Add validation acceptance tests
5. Document validation behavior in CLAUDE.md

---

## References

- **Test Scripts:**
  - `/workspace/sampleRest/test_validate_softwareimage.py`
  - `/workspace/sampleRest/test_validate_comprehensive.py`
  - `/workspace/sampleRest/validate_methods_with_entities.py`

- **API Documentation:**
  - BCM JSON-RPC API: `https://172.21.15.254:8081/json`

- **Related Code:**
  - BCM Client: `/workspace/internal/provider/bcm_client.go`
  - Software Image Resource: `/workspace/internal/provider/resource_cmpart_softwareimage.go`
  - Category Resource: `/workspace/internal/provider/resource_cmdevice_category.go`
