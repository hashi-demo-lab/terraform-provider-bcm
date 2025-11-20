# Research: BCM CMDevice Category Resource

**Date**: 2025-11-21
**Feature**: BCM CMDevice Category Resource
**Phase**: 0 - Research & Technology Decisions

## Overview

This document captures research findings and technology decisions for implementing the BCM CMDevice Category Terraform resource. The implementation follows the established patterns from the existing `bcm_cmpart_softwareimage` resource.

## Section 1: API Method Verification

### Research Question
Verify the exact API methods available for Category CRUD operations and their signatures.

### Decision
The BCM CMDevice service provides the following API methods for category management:

**Confirmed API Methods:**
- `getCategories()` - Lists all categories (returns array)
- `getCategory(name)` - Retrieves single category by name (efficient direct lookup)
- `addCategory(entity, force)` - Creates new category
- `updateCategory(entity, force)` - Updates existing category
- `validateCategory(entity)` - Pre-flight validation
- `removeCategory(uuid, force)` - Deletes category

**Method Signatures:**
```json
{
  "service": "cmdevice",
  "call": "getCategory",
  "args": ["category-name"]
}

{
  "service": "cmdevice",
  "call": "addCategory",
  "args": [categoryEntity, false]
}

{
  "service": "cmdevice",
  "call": "updateCategory",
  "args": [categoryEntity, false]
}

{
  "service": "cmdevice",
  "call": "removeCategory",
  "args": ["category-uuid", false]
}

{
  "service": "cmdevice",
  "call": "validateCategory",
  "args": [categoryEntity]
}
```

**Rationale:**
- Pattern matches CMPart service methods (getSoftwareImage, addSoftwareImage, etc.)
- BCM API consistently uses this naming convention across services
- Force parameter provides safety mechanism for operations with side effects
- Validation method enables pre-flight error detection

**Alternatives Considered:**
- Using generic REST endpoints - rejected because BCM uses JSON-RPC exclusively
- List+filter pattern for Read - rejected in favor of efficient direct lookup via getCategory(name)

**Implementation Impact:**
- Read operation will use `getCategory(name)` for efficiency
- Import operation will use `getCategories()` to find by UUID, then `getCategory(name)` for full data
- Force parameter exposed as optional resource attribute (default: false)

---

## Section 2: Nested Object Handling Strategy

### Research Question
How should complex nested objects (BMCSettings, FSMount arrays, KernelModule arrays) be mapped between Terraform schema and BCM API entities?

### Decision
Use terraform-plugin-framework nested attribute types with dedicated Go structs for each nested object type.

**Nested Object Mapping:**

1. **SoftwareImageProxy** (single nested object):
   - Schema: `schema.SingleNestedAttribute`
   - Model: `SoftwareImageProxyModel` struct
   - API: Map to/from `map[string]interface{}` with `baseType: "SoftwareImageProxy"`

2. **BMCSettings** (single nested object):
   - Schema: `schema.SingleNestedAttribute`
   - Model: `BMCSettingsModel` struct
   - Mark password as sensitive: `Sensitive: true`
   - API: Map to/from `map[string]interface{}` with `baseType: "BMCSettings"`

3. **FSMount** (array of nested objects):
   - Schema: `schema.ListNestedAttribute`
   - Model: `types.List` with `FSMountModel` element type
   - API: Map to/from `[]map[string]interface{}` with each element having `baseType: "FSMount"`

4. **KernelModule** (array of nested objects):
   - Schema: `schema.ListNestedAttribute`
   - Model: `types.List` with `KernelModuleModel` element type
   - API: Map to/from `[]map[string]interface{}` with each element having `baseType: "KernelModule"`

5. **Optional Complex Objects** (biosSetup, dpuSettings, etc.):
   - Schema: `schema.SingleNestedAttribute` with `Optional: true`
   - Model: `types.Object` with `IsNull()` check
   - API: Only include in entity if not null

**Rationale:**
- Follows terraform-plugin-framework best practices for nested attributes
- Matches pattern used in resource_cmpart_softwareimage.go (KernelModule handling)
- Sensitive field marking prevents password leakage in logs and plan output
- Optional objects set to null in Terraform map to omitted fields in API requests

**Alternatives Considered:**
- Flattening nested objects into top-level attributes - rejected because:
  - Loses semantic grouping (e.g., all BMC settings together)
  - Makes schema harder to understand and document
  - Doesn't match BCM API structure
- Using JSON strings for complex objects - rejected because:
  - Poor user experience (requires JSON escaping)
  - No type validation at plan time
  - Not idiomatic Terraform

**Implementation Impact:**
- Create separate model structs for each nested object type
- Implement helper methods for converting between Terraform types and API maps
- Add `baseType`, `childType`, `modified`, `to_be_removed`, `revision` fields when building API entities
- Handle null values gracefully for optional nested objects

---

## Section 3: State Management for Long Text Fields

### Research Question
How should very long text fields (disksetup XML up to 10KB, excludeList fields up to 50KB) be handled in Terraform state and API calls?

### Decision
Store long text fields as plain `types.String` attributes with validation for maximum size limits.

**Implementation Strategy:**

1. **Disksetup XML** (up to 10KB):
   - Schema: `schema.StringAttribute` with `Optional: true`
   - Validation: `stringvalidator.LengthBetween(0, 10240)` (10KB = 10,240 bytes)
   - Storage: Plain string in Terraform state
   - API: Send as-is in category entity

2. **Exclude List Fields** (up to 50KB each):
   - Fields: `exclude_list_full`, `exclude_list_grab`, `exclude_list_grabnew`, `exclude_list_sync`, `exclude_list_update`
   - Schema: `schema.StringAttribute` with `Optional: true` for each
   - Validation: `stringvalidator.LengthBetween(0, 51200)` (50KB = 51,200 bytes)
   - Storage: Plain string in Terraform state
   - API: Send as-is in category entity

3. **Script Fields** (initialize, finalize):
   - Schema: `schema.StringAttribute` with `Optional: true`
   - No explicit size limit (reasonable limits enforced by BCM API)
   - Storage: Plain string in Terraform state

**Rationale:**
- Terraform state files handle large strings efficiently (compressed in remote backends)
- Plain string attributes provide best user experience for heredoc syntax
- Size validation prevents API errors from oversized content
- No need for external file references or special encoding

**Alternatives Considered:**
- External file references (e.g., `disksetup_file` attribute) - rejected because:
  - Breaks Terraform's declarative model (file content changes not tracked)
  - Complicates state management
  - Not consistent with existing provider patterns
- Base64 encoding - rejected because:
  - Unnecessary complexity for text content
  - Makes configuration less readable
  - Requires encoding/decoding logic
- Separate resources for long content - rejected because:
  - Over-engineering for straightforward text fields
  - Breaks semantic grouping of category configuration

**Implementation Impact:**
- Add length validators to schema definition
- No special handling needed in CRUD operations (treat as regular strings)
- Document heredoc syntax in examples for multi-line content

---

## Section 4: Validation Strategy

### Research Question
Should validation be client-side only, server-side only, or hybrid? When should validateCategory API be called?

### Decision
Implement hybrid validation approach with lightweight client-side checks and pre-flight server-side validation via validateCategory API.

**Validation Layers:**

1. **Client-Side Validation** (Terraform schema validators):
   - Category name length: 1-255 characters
   - Management network UUID format: RFC 4122 regex
   - Boot loader values: OneOf validator for known types
   - SOL speed values: OneOf validator for valid baud rates
   - IP address format: IP address regex for gateway and nameservers
   - String length limits: disksetup (10KB), excludeLists (50KB)

2. **Server-Side Validation** (BCM API):
   - Call `validateCategory` before `addCategory` and `updateCategory`
   - Parse validation response for errors and warnings
   - Surface errors as Terraform diagnostics
   - Proceed with operation only if validation passes OR force=true

**Validation Flow:**

```go
// Create operation
func (r *CMDeviceCategoryResource) Create(...) {
    // 1. Client-side validation (automatic via schema validators)

    // 2. Build API entity
    entity := r.buildAPIEntity(&plan, "")

    // 3. Server-side validation (optional but recommended)
    validationResult, err := r.client.CallJSONRPC(ctx, "cmdevice", "validateCategory", entity)
    if err != nil {
        // Parse validation errors and add to diagnostics
        resp.Diagnostics.AddError("Category Validation Failed", err.Error())
        return
    }

    // 4. Proceed with creation
    _, err = r.client.CallJSONRPC(ctx, "cmdevice", "addCategory", entity, plan.Force.ValueBool())
    // ...
}
```

**Rationale:**
- Client-side validation provides fast feedback during `terraform plan`
- Server-side validation catches business logic errors (e.g., duplicate names, invalid references)
- Pre-flight validateCategory call enables better error messages before mutation
- Force parameter allows overriding warnings while respecting hard errors

**Alternatives Considered:**
- Client-side only - rejected because:
  - Can't validate business logic constraints (duplicate names, referenced UUIDs)
  - Can't detect warnings that might need force parameter
- Server-side only - rejected because:
  - Slower feedback loop (requires API call for simple format errors)
  - Wastes API calls for obviously invalid input
- No validateCategory call - rejected because:
  - Miss opportunity for better error messages
  - Can't distinguish between warnings (overridable) and errors (blocking)

**Implementation Impact:**
- Add schema validators for all client-side checks
- Implement validateCategory call in Create and Update operations
- Parse validation response and convert to Terraform diagnostics
- Handle force parameter logic for overriding warnings

---

## Section 5: Import Operation Strategy

### Research Question
How should import by UUID work when the API requires name for efficient lookup?

### Decision
Implement two-phase import: use `getCategories()` to find name by UUID, then use `getCategory(name)` for full data retrieval.

**Import Flow:**

```go
func (r *CMDeviceCategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    // User provides UUID as import identifier
    importUUID := req.ID

    // Phase 1: List all categories to find matching UUID
    allCategoriesBody, err := r.client.CallJSONRPC(ctx, "cmdevice", "getCategories")
    if err != nil {
        resp.Diagnostics.AddError("Import Failed", "Could not list categories: " + err.Error())
        return
    }

    var categoryList []map[string]interface{}
    json.Unmarshal(allCategoriesBody, &categoryList)

    // Find category with matching UUID
    var categoryName string
    for _, cat := range categoryList {
        if uuid, ok := cat["uuid"].(string); ok && uuid == importUUID {
            categoryName = cat["name"].(string)
            break
        }
    }

    if categoryName == "" {
        resp.Diagnostics.AddError("Category Not Found", "No category with UUID: " + importUUID)
        return
    }

    // Phase 2: Use efficient getCategory(name) for full data
    // Set ID to UUID for subsequent Read operation
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Rationale:**
- Import by UUID is more reliable than name (names can be changed)
- Two-phase approach balances efficiency with API constraints
- getCategories() call is acceptable for import (infrequent operation)
- Subsequent reads use efficient getCategory(name) lookup

**Alternatives Considered:**
- Import by name directly - rejected because:
  - UUIDs are more stable identifiers
  - Users typically have UUID from BCM UI, not name
  - Not consistent with other Terraform providers
- Single getCategory(uuid) call - rejected because:
  - BCM API doesn't support getCategory by UUID
  - Would require API enhancement (out of scope)

**Implementation Impact:**
- Implement ImportState method with two-phase lookup
- Document import format in resource documentation: `terraform import bcm_cmdevice_category.example <uuid>`
- Add acceptance test for import operation

---

## Section 6: Eventual Consistency Handling

### Research Question
Do category operations require eventual consistency handling like software image cloning?

### Decision
No explicit eventual consistency handling required for category operations. Unlike software image cloning (which is asynchronous), category CRUD operations are synchronous.

**Analysis:**

1. **Software Image Clone** (asynchronous):
   - `addSoftwareImage` with `originalImage` initiates async clone
   - Kernel files copied in background
   - `fileOperationInProgress` flag indicates clone status
   - **Requires** polling with exponential backoff

2. **Category Operations** (synchronous):
   - `addCategory` completes before returning
   - No background file operations
   - No operation-in-progress flags
   - **No polling needed**

**Rationale:**
- Category entities are pure metadata (no file operations)
- BCM API returns success only after database transaction commits
- Read-after-write consistency guaranteed by BCM's transaction model
- No evidence of eventual consistency in Category operations from API testing

**Alternatives Considered:**
- Adding optional polling for all operations - rejected because:
  - Unnecessary complexity for synchronous operations
  - Adds latency to typical workflows
  - No API indicator that polling is needed
- Retry logic with exponential backoff - rejected because:
  - API errors are immediate, not transient
  - Retry could exacerbate duplicate creation issues

**Implementation Impact:**
- No polling logic needed in Create/Update/Delete operations
- Standard read-after-write pattern: create → immediate read
- Simpler implementation compared to software image resource

---

## Section 7: Force Parameter Behavior

### Research Question
How should the force parameter be exposed to users and when should it be applied?

### Decision
Expose force as optional boolean attribute (default: false) that applies to create, update, and delete operations.

**Force Parameter Usage:**

1. **Create Operation** (`addCategory`):
   - `force=false` (default): Fail on validation warnings
   - `force=true`: Override validation warnings, still fail on errors
   - Use case: Creating category with experimental settings

2. **Update Operation** (`updateCategory`):
   - `force=false` (default): Fail if update would impact assigned nodes
   - `force=true`: Apply update even if nodes are assigned
   - Use case: Modifying production category configuration

3. **Delete Operation** (`removeCategory`):
   - `force=false` (default): Fail if category has assigned nodes
   - `force=true`: Delete category even with node assignments
   - Use case: Cleaning up test environments

**Schema Definition:**

```go
"force": schema.BoolAttribute{
    Optional:            true,
    Computed:            true,
    Default:             booldefault.StaticBool(false),
    MarkdownDescription: "Force operation even if warnings exist or nodes are assigned. " +
        "Use with caution in production environments.",
},
```

**Rationale:**
- Single attribute controls behavior across all operations (simpler UX)
- Safe default (force=false) prevents accidental destructive actions
- Terraform's declarative model means force setting is part of resource definition
- Matches pattern from software image resource

**Alternatives Considered:**
- Separate force attributes per operation (force_create, force_update, force_delete) - rejected because:
  - More complex schema
  - Users would need to understand multiple force flags
  - Inconsistent with existing patterns
- No force parameter (rely on API errors) - rejected because:
  - Users can't override warnings when intentional
  - Delete operations would fail in development workflows
- Force as provider-level setting - rejected because:
  - Too coarse-grained (all-or-nothing)
  - Can't have different force behavior per resource

**Implementation Impact:**
- Add force attribute to schema
- Pass force value to API calls: `CallJSONRPC(..., entity, plan.Force.ValueBool())`
- Document force behavior and safety warnings
- Add acceptance test for force parameter scenarios

---

## Section 8: Helper Function Reuse

### Research Question
Can helper functions from existing resources be reused for Category implementation?

### Decision
Reuse existing helper functions from `data_source_cmpart_softwareimages.go` for null-safe field extraction.

**Reusable Helper Functions:**

```go
// Located in: internal/provider/data_source_cmpart_softwareimages.go:399-431

func getStringValue(data map[string]interface{}, key string) types.String
func getBoolValue(data map[string]interface{}, key string) types.Bool
func getInt64Value(data map[string]interface{}, key string) types.Int64
```

**Usage Pattern:**

```go
// In readCategory method
func (r *CMDeviceCategoryResource) readCategory(ctx context.Context, model *CMDeviceCategoryResourceModel, diags *diag.Diagnostics) {
    // ... API call to getCategory ...

    var categoryData map[string]interface{}
    json.Unmarshal(body, &categoryData)

    // Use helper functions for null-safe extraction
    model.Name = getStringValue(categoryData, "name")
    model.Notes = getStringValue(categoryData, "notes")
    model.DataNode = getBoolValue(categoryData, "dataNode")
    model.DefaultGatewayMetric = getInt64Value(categoryData, "defaultGatewayMetric")
    // ... etc ...
}
```

**Rationale:**
- DRY principle: avoid duplicating null-handling logic
- Proven implementation: functions already tested in production
- Consistent behavior: same null-handling across resources
- Type safety: returns proper types.String/Bool/Int64 with null support

**Alternatives Considered:**
- Duplicating helper functions in category resource - rejected because:
  - Violates DRY principle
  - Creates maintenance burden (fix bugs in multiple places)
  - No customization needed for category-specific logic
- Moving helpers to shared utilities package - rejected because:
  - Over-engineering for current scope
  - Functions are provider-specific (not generic enough)
  - Can be refactored later if more resources added
- Inline null checks everywhere - rejected because:
  - Verbose and error-prone
  - Harder to maintain consistent behavior

**Implementation Impact:**
- No new helper functions needed
- Import existing helpers from data source file
- Document helper function location in code comments
- Consider refactoring to shared package in future if 5+ resources use them

---

## Section 9: Testing Strategy

### Research Question
What acceptance test scenarios are needed to achieve comprehensive coverage?

### Decision
Implement 5 core acceptance test scenarios following TDD RED-GREEN-REFACTOR cycles.

**Test Scenarios:**

1. **TestAccCMDeviceCategoryResource_Basic** (P0 - Core CRUD):
   ```go
   Steps:
   - Create category with minimal configuration (name, management_network)
   - Verify all fields in state (ID, UUID, name, etc.)
   - Update category notes and kernel_parameters
   - Verify updates applied correctly
   - ImportState testing
   - Delete automatically tested by framework
   ```

2. **TestAccCMDeviceCategoryResource_Complete** (P1 - Complex configuration):
   ```go
   Steps:
   - Create category with comprehensive configuration:
     - Boot configuration (boot_loader, kernel settings)
     - Nested objects (bmc_settings, software_image_proxy)
     - Arrays (fsmounts, modules, name_servers)
     - Long text fields (disksetup XML, exclude_list_full)
   - Verify all nested objects persisted correctly
   - Update multiple fields simultaneously
   - Verify no field corruption
   ```

3. **TestAccCMDeviceCategoryResource_NestedObjects** (P1 - Nested attribute handling):
   ```go
   Steps:
   - Create category with empty nested objects
   - Update to add nested objects (FSMount, KernelModule)
   - Update to modify nested objects
   - Update to remove nested objects (set to null)
   - Verify state correctly reflects all changes
   ```

4. **TestAccCMDeviceCategoryResource_Import** (P0 - Import functionality):
   ```go
   Steps:
   - Create category via Terraform
   - Destroy Terraform state (but not resource)
   - Import category by UUID
   - Verify terraform plan shows no changes
   - Update imported category
   - Verify changes applied correctly
   ```

5. **TestAccCMDeviceCategoryResource_ForceParameter** (P2 - Force behavior):
   ```go
   Steps:
   - Create category with force=false
   - Attempt update that triggers warning (via mock or real scenario)
   - Verify operation fails with clear error
   - Update with force=true
   - Verify operation succeeds
   - Test delete with force parameter
   ```

**Test Data Strategy:**
- Use unique category names with timestamp suffix: `test-category-{unix_timestamp}`
- Use real management network UUID from test environment
- Avoid hardcoded UUIDs except for management network (environment variable)
- Use heredoc for disksetup XML examples
- Clean up resources in each test case

**Rationale:**
- Covers all CRUD operations plus import
- Tests edge cases (null values, empty arrays, long text)
- Validates force parameter behavior
- Follows HashiCorp acceptance testing best practices
- Each test is independent and can run in parallel

**Implementation Impact:**
- Write all 5 test cases in RED phase (expect failures)
- Implement minimal code in GREEN phase to pass tests
- Refactor in REFACTOR phase while keeping tests green
- Run tests with: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource`

---

## Summary of Decisions

| Area | Decision | Rationale |
|------|----------|-----------|
| **API Methods** | Use getCategory(name) for Read, getCategories() for import | Efficient direct lookup, matches software image pattern |
| **Nested Objects** | Use terraform-plugin-framework nested attributes with dedicated models | Type-safe, good UX, matches existing patterns |
| **Long Text Fields** | Store as plain strings with length validation | Simple, efficient, supports heredoc syntax |
| **Validation** | Hybrid: client-side schema validators + server-side validateCategory | Fast feedback + comprehensive validation |
| **Import** | Two-phase: getCategories() to find name, then getCategory(name) | Supports UUID import while using efficient lookup |
| **Eventual Consistency** | No special handling needed | Category operations are synchronous |
| **Force Parameter** | Single optional boolean attribute | Simple UX, safe default, applies to all operations |
| **Helper Functions** | Reuse from data_source_cmpart_softwareimages.go | DRY principle, proven implementation |
| **Testing Strategy** | 5 acceptance test scenarios with TDD workflow | Comprehensive coverage, follows best practices |

## Open Questions

**None** - All technical unknowns resolved through research and reference implementation analysis.

## Next Steps

Proceed to **Phase 1: Design & Contracts**:
1. Generate data-model.md with entity definitions
2. Create contracts/cmdevice-category-api.md with detailed API specifications
3. Generate quickstart.md with developer setup instructions
4. Update agent context files with category-specific patterns
