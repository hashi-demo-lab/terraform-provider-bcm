# Data Model: BCM Software Image Resource

**Feature**: Complete TDD-Based Review and Refactoring of resource_cmpart_softwareimage
**Date**: 2025-11-21
**Purpose**: Define the complete mapping between BCM API entities and Terraform schema

## Overview

The BCM Software Image resource manages operating system images used for DPU node provisioning. Each image consists of:
- **Kernel**: Linux kernel binary with version identifier
- **Filesystem**: Root filesystem with installed packages and configurations
- **Modules**: Kernel modules to load at boot with optional parameters
- **Boot Configuration**: Kernel parameters, console settings, Serial-over-LAN (SOL) settings
- **Metadata**: Creation time, revision tracking, filesystem partition references

---

## Entity Mapping Table

| BCM API Field | Terraform Attribute | BCM Type | Terraform Type | Classification | Null Handling |
|--------------|-------------------|----------|----------------|----------------|---------------|
| `uuid` | `id`, `uuid` | string | types.String | Computed | Never null (BCM assigns) |
| `name` | `name` | string | types.String | Required | Cannot be null |
| `path` | `path` | string | types.String | Required | Cannot be null |
| `kernelVersion` | `kernel_version` | string | types.String | Optional+Computed | Set to null if missing |
| `kernelParameters` | `kernel_parameters` | string | types.String | Optional+Computed | Set to "" if missing |
| `kernelOutputConsole` | `kernel_output_console` | string | types.String | Optional+Computed | Set to "" if missing |
| `enableSOL` | `enable_sol` | boolean | types.Bool | Optional+Computed | Default: false |
| `SOLPort` | `sol_port` | string | types.String | Optional+Computed | Default: "0" |
| `SOLSpeed` | `sol_speed` | string | types.String | Optional+Computed | Default: "115200" |
| `SOLFlowControl` | `sol_flow_control` | boolean | types.Bool | Optional+Computed | Default: true |
| `modules` | `modules` | KernelModule[] | types.List | Optional+Computed | Set to [] if missing |
| `fspart` | `fspart` | string (UUID) | types.String | Computed | Set to null if missing |
| `bootfspart` | `bootfspart` | string (UUID) | types.String | Computed | Set to null if missing |
| `notes` | `notes` | string | types.String | Optional | Set to "" if missing |
| `creationTime` | `creation_time` | int64 | types.Int64 | Computed | Never null (BCM assigns) |
| `revisionID` | `revision_id` | int64 | types.Int64 | Computed | Never null (BCM assigns) |
| `fileOperationInProgress` | `file_operation_in_progress` | boolean | types.Bool | Computed | Never null (BCM assigns) |
| `originalImage` | `original_image` | string (UUID) | types.String | Optional+Computed | **Special: Preserve plan value** |
| `parentSoftwareImage` | `parent_software_image` | string (UUID) | types.String | Computed | Set to null if missing |

### Special Attributes

#### original_image (Optional+Computed with Plan Preservation)
**BCM Behavior**: After clone completes, BCM API resets this field to zero UUID or empty string
**Terraform Behavior**: Preserve the plan's original_image value in state for audit trail purposes
**Rationale**: Users need to know which image was cloned from, even after BCM resets it

**State Resolution Logic**:
```
if plan.original_image is Known (not Unknown):
    state.original_image = plan.original_image  # Preserve user's intent
else if api.originalImage is non-zero UUID:
    state.original_image = api.originalImage  # Use API value
else:
    state.original_image = null  # Neither plan nor API has value
```

**Plan Modifier**: `stringplanmodifier.UseStateForUnknown()` prevents unnecessary diffs

#### modules (Optional+Computed with Known List Guarantee)
**BCM Behavior**: Returns array of module objects, or omits field if no modules
**Terraform Behavior**: Always set to known list (never Unknown, never null)
**Rationale**: Terraform framework requires lists to be known values after apply completes

**State Resolution Logic**:
```
if api.modules exists and has length > 0:
    state.modules = convert_api_modules_to_terraform_list(api.modules)
else:
    state.modules = []  # Empty list, not null or Unknown
```

**Framework Requirement**: Unknown list values cause "invalid result object" errors

---

## Nested Object: KernelModule

| BCM API Field | Terraform Attribute | BCM Type | Terraform Type | Classification | Null Handling |
|--------------|-------------------|----------|----------------|----------------|---------------|
| `name` | `name` | string | types.String | Required | Cannot be null |
| `parameters` | `parameters` | string | types.String | Optional | **Set to "" if null** |

### Special Handling: Module Parameters

**BCM API Requirement**: The `parameters` field must be a string type. If no parameters are provided, BCM expects empty string `""`, not `null`.

**Terraform Mapping**:
```go
// Building API entity
if moduleParameters.IsNull() || moduleParameters.IsUnknown() {
    module["parameters"] = ""  // Empty string, not null
} else {
    module["parameters"] = moduleParameters.ValueString()
}
```

**Example**:
```hcl
modules = [
  {
    name       = "nvidia-drm"
    parameters = "modeset=1"  # Has parameters
  },
  {
    name       = "nvme"
    # parameters omitted → API receives ""
  }
]
```

---

## State Transitions

### Create Operation

**Input**: Terraform plan with user-provided attributes
**Process**:
1. Extract required fields: name, path
2. Extract optional fields: kernel config, SOL config, modules, original_image
3. Build BCM API entity with baseType="SoftwareImage"
4. Call `CMPart.addSoftwareImage(entity, cloneData=true)`
5. **If original_image set**: Poll fileOperationInProgress until clone completes
6. Read back final state from BCM API
7. **Critical**: Preserve plan.original_image in state (even if BCM reset it)
8. Set computed fields: uuid, creation_time, revision_id, fspart, bootfspart

**Output**: Terraform state with all attributes populated

**State Diagram**:
```
PLAN (user config)
  ↓
BUILD ENTITY (add baseType, childType, modified, revision)
  ↓
API CALL (addSoftwareImage)
  ↓ (if original_image set)
POLL CLONE STATUS (exponential backoff: 1s, 2s, 4s, 8s, 16s, 16s)
  ↓
READ BACK STATE (getSoftwareImage)
  ↓
PRESERVE original_image (from plan, not API)
  ↓
SET STATE (all attributes known)
```

### Read Operation

**Input**: Terraform state with UUID
**Process**:
1. Extract UUID/name from state
2. Call `CMPart.getSoftwareImage(name)` (direct lookup)
3. Map BCM API response to Terraform attributes
4. **Critical**: Check if plan has known original_image value
   - If plan exists AND original_image is Known: Preserve plan value
   - Otherwise: Use API value (may be zero UUID or null)
5. Set modules to known list (empty list if API returns null)
6. Update computed fields from API response

**Output**: Updated Terraform state reflecting BCM's current state

**State Diagram**:
```
STATE (with UUID)
  ↓
API CALL (getSoftwareImage by name)
  ↓
MAP API FIELDS (to Terraform types)
  ↓
CHECK PLAN (for known original_image value)
  ↓
PRESERVE original_image (if plan has known value)
  ↓
SET modules (to known list, never Unknown/null)
  ↓
UPDATE STATE (all attributes refreshed)
```

### Update Operation

**Input**: Terraform plan with changed attributes
**Process**:
1. Extract all attributes from plan (including unchanged ones)
2. Build BCM API entity with **UUID included** (update requirement)
3. **Exclude original_image** from update entity (create-only field)
4. Call `CMPart.updateSoftwareImage(entity)`
5. Call Read operation to refresh state
6. Read operation handles original_image preservation

**Output**: Updated Terraform state with new values

**State Diagram**:
```
PLAN (with changes)
  ↓
BUILD ENTITY (include UUID, exclude original_image)
  ↓
API CALL (updateSoftwareImage)
  ↓
CALL READ (to refresh state)
  ↓
UPDATE STATE (via Read operation)
```

### Delete Operation

**Input**: Terraform state with UUID
**Process**:
1. Extract UUID from state
2. Call `CMPart.removeSoftwareImage(uuid, false, false, false)`
   - removeData=false: Keep filesystem data
   - removeAll=false: Don't remove related entities
   - force=false: Check dependencies before deleting
3. Log deletion (tflog.Info)
4. Remove resource from state (automatic)

**Output**: Resource removed from Terraform state

**State Diagram**:
```
STATE (with UUID)
  ↓
API CALL (removeSoftwareImage with safe flags)
  ↓
LOG DELETION
  ↓
REMOVE FROM STATE (automatic)
```

### Import Operation

**Input**: UUID provided by user
**Process**:
1. User runs: `terraform import bcm_cmpart_softwareimage.test <uuid>`
2. ImportStatePassthroughID sets id attribute to UUID
3. Read operation populates all attributes from BCM API
4. **Note**: original_image will be null or zero UUID (BCM doesn't preserve clone source)

**Output**: Imported resource in Terraform state

**State Diagram**:
```
USER INPUT (UUID)
  ↓
SET id = UUID
  ↓
CALL READ (to populate all attributes)
  ↓
IMPORT COMPLETE (state populated from BCM API)
```

---

## BCM API Entity Structure

### SoftwareImage Entity (for Create/Update)

**Required Fields** (always present):
```json
{
  "baseType": "SoftwareImage",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "name": "<user-provided>",
  "path": "<user-provided>"
}
```

**Optional Fields for Create**:
```json
{
  "originalImage": "<uuid>",  // Only for clone operations
  "kernelVersion": "<version>",
  "kernelParameters": "<params>",
  "kernelOutputConsole": "<console>",
  "enableSOL": true/false,
  "SOLPort": "<port>",
  "SOLSpeed": "<speed>",
  "SOLFlowControl": true/false,
  "notes": "<text>",
  "modules": [
    {
      "baseType": "KernelModule",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "name": "<module-name>",
      "parameters": "<module-params>"
    }
  ]
}
```

**Additional Field for Update**:
```json
{
  "uuid": "<resource-uuid>"  // Required for updates, not for creates
}
```

### API Response (from getSoftwareImage)

**Complete Response**:
```json
{
  "baseType": "SoftwareImage",
  "childType": "",
  "modified": false,
  "to_be_removed": false,
  "revision": "",
  "uuid": "<uuid>",
  "name": "<name>",
  "path": "<path>",
  "kernelVersion": "<version>",
  "kernelParameters": "<params>",
  "kernelOutputConsole": "<console>",
  "enableSOL": true/false,
  "SOLPort": "<port>",
  "SOLSpeed": "<speed>",
  "SOLFlowControl": true/false,
  "notes": "<text>",
  "modules": [/* KernelModule objects */],
  "fspart": "<uuid>",
  "bootfspart": "<uuid>",
  "creationTime": 1234567890,
  "revisionID": 1,
  "fileOperationInProgress": false,
  "originalImage": "",  // Reset to empty after clone
  "parentSoftwareImage": "<uuid>"
}
```

---

## Validation Rules

### Name Validation
**Rule**: Length between 1-255 characters
**Validator**: `stringvalidator.LengthBetween(1, 255)`
**Rationale**: BCM database field constraint

**Invalid Examples**:
```hcl
name = ""  # Error: length must be at least 1
name = string(repeat("a", 256))  # Error: length must be at most 255
```

### Path Validation
**Rule**: Regex `^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$`
**Validator**: `stringvalidator.RegexMatches(...)`
**Rationale**: BCM filesystem path requirements

**Valid Examples**:
```hcl
path = "/cm/images/my-image"
path = "/cm/images/my-image-v2"
path = "/cm/images/my-image@123"  # With revision
path = "/cm/images/nested/path/image"
```

**Invalid Examples**:
```hcl
path = "relative/path"  # Error: must be absolute (start with /)
path = "/cm/images/my image"  # Error: spaces not allowed
path = "/cm/images/my@image@123"  # Error: @ only allowed at end with digits
```

### SOL Speed Validation
**Rule**: OneOf(115200, 57600, 38400, 19200, 9600, 4800, 2400, 1200)
**Validator**: `stringvalidator.OneOf(...)`
**Rationale**: BCM serial port hardware limitations

**Valid Examples**:
```hcl
sol_speed = "115200"  # Default
sol_speed = "9600"
```

**Invalid Examples**:
```hcl
sol_speed = "9999"  # Error: must be one of valid speeds
sol_speed = "115200bps"  # Error: exact match required
```

### Module Name Validation
**Rule**: Length at least 1 character
**Validator**: `stringvalidator.LengthAtLeast(1)`
**Rationale**: Empty module names are invalid

**Valid Examples**:
```hcl
modules = [
  {
    name = "nvidia-drm"
    parameters = "modeset=1"
  }
]
```

**Invalid Examples**:
```hcl
modules = [
  {
    name = ""  # Error: name must have at least 1 character
  }
]
```

---

## Two-Step Create Pattern (BCM API Limitation)

### Problem
BCM API validates kernel file paths BEFORE clone operation completes. If `kernel_version` is set during initial create with `original_image`, the API may return validation error because kernel files don't exist yet (still being cloned).

### Solution: Two-Step Pattern

**Step 1: Create with Basic Config**
```hcl
resource "bcm_cmpart_softwareimage" "test" {
  name           = "my-image"
  path           = "/cm/images/my-image"
  original_image = data.bcm_cmpart_softwareimages.all.images[0].uuid
  # Do NOT set kernel_version, kernel_parameters, or modules here
}
```

**Step 2: Update with Kernel Config**
```hcl
resource "bcm_cmpart_softwareimage" "test" {
  name           = "my-image"
  path           = "/cm/images/my-image"
  original_image = data.bcm_cmpart_softwareimages.all.images[0].uuid

  # Now add kernel configuration
  kernel_version        = "5.15.0-custom"
  kernel_parameters     = "quiet splash nomodeset"
  kernel_output_console = "ttyS0,115200"

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    }
  ]
}
```

### Implementation in Tests
Tests use two separate `resource.TestStep` blocks:
1. First step: Create with basic config, verify creation
2. Second step: Update with kernel config, verify updates

This pattern is **not a Terraform bug** - it's a workaround for BCM API's validation timing.

---

## Helper Functions

### buildAPIEntity()
**Purpose**: Construct BCM API entity from Terraform state
**Location**: `resource_cmpart_softwareimage.go` lines 738-828
**Signature**: `buildAPIEntity(ctx, model, diagnostics, isUpdate) map[string]interface{}`

**Logic**:
- Always include: baseType, childType, modified, to_be_removed, revision
- Include UUID only if isUpdate=true
- Include original_image only if isUpdate=false AND value is set
- Convert modules list to BCM array format
- Set module parameters to "" if null (BCM requirement)
- Handle all null/Unknown checks for optional fields

### readSoftwareImage()
**Purpose**: Map BCM API response to Terraform state
**Location**: `resource_cmpart_softwareimage.go` lines 639-736
**Signature**: `readSoftwareImage(ctx, model, plan, diagnostics) error`

**Logic**:
- Call getSoftwareImage API with name parameter
- Map all API fields using helper functions (getStringValue, getBoolValue, etc.)
- **Critical**: Preserve plan.original_image if plan is provided and value is Known
- Set modules to known list (empty list if API returns null)
- Update all computed fields from API response

### Helper Functions (from data_source_cmpart_softwareimages.go)
**Purpose**: Null-safe extraction from map[string]interface{}

- `getStringValue(data, key)`: Returns types.String with null handling
- `getBoolValue(data, key)`: Returns types.Bool with null handling
- `getInt64Value(data, key)`: Returns types.Int64 with null handling (handles float64, int64, int)

**Usage**:
```go
model.Name = getStringValue(imageData, "name")
model.EnableSOL = getBoolValue(imageData, "enableSOL")
model.CreationTime = getInt64Value(imageData, "creationTime")
```

---

## Summary

The BCM Software Image resource implements a sophisticated state management pattern that:

1. **Preserves user intent** - original_image maintained in state for audit trail
2. **Handles async operations** - Clone polling with exponential backoff
3. **Resolves Unknown values** - All attributes known after apply completes
4. **Validates inputs** - Schema validators catch errors during plan phase
5. **Supports all CRUD operations** - Create, Read, Update, Delete, Import
6. **Works around API limitations** - Two-step pattern for kernel configuration

The data model is production-ready and follows HashiCorp Terraform Plugin Framework best practices.
