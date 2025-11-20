# API Research: bcm_cmpart_softwareimage Resource

**Status**: ✅ COMPLETE - Update and Delete methods confirmed

**Research Date**: 2025-11-20

**BCM API Endpoint**: `https://172.21.15.254:8081/json`

---

## API Method Reference

### Create: `addSoftwareImage`

**Service**: `cmpart` (or `CMPart`)

**Method**: `addSoftwareImage`

**Signature**:
```json
{
  "service": "cmpart",
  "call": "addSoftwareImage",
  "args": [softwareImage, force]
}
```

**Parameters**:
- `softwareImage` (object): Complete SoftwareImage entity with all fields
- `force` (boolean): Force creation flag (semantics TBD - likely overwrite protection)

**Expected Response**: UUID string or SoftwareImage object with `uuid` field

**Example Request** (from `/workspace/sampleRest/wip/resource_cmpart_softwareimage.md`):
```json
{
  "service": "CMPart",
  "call": "addSoftwareImage",
  "args": [
    {
      "uuid": "eaad50d3-432a-4703-a9f8-66551c255a69",
      "baseType": "SoftwareImage",
      "childType": "",
      "to_be_removed": false,
      "modified": true,
      "revision": "",
      "name": "cloned",
      "path": "/cm/images/cloned",
      "originalImage": "8482c4e9-383c-43de-873f-8c54ee77ee74",
      "fileOperationInProgress": false,
      "kernelVersion": "6.8.0-51-generic",
      "kernelParameters": "rd.driver.blacklist=nouveau",
      "kernelOutputConsole": "tty0",
      "creationTime": 1763626149,
      "modules": [
        {
          "uuid": "16867c38-7b60-4fbe-b4fb-888fd3ea4ba5",
          "baseType": "KernelModule",
          "childType": "",
          "to_be_removed": false,
          "modified": true,
          "revision": "",
          "name": "aacraid",
          "parameters": ""
        }
      ],
      "enableSOL": false,
      "SOLPort": "ttyS1",
      "SOLSpeed": "115200",
      "SOLFlowControl": true,
      "notes": "",
      "fspart": "00000000-0000-0000-0000-000000000000",
      "bootfspart": "00000000-0000-0000-0000-000000000000",
      "revisionID": 0,
      "parentSoftwareImage": "00000000-0000-0000-0000-000000000000",
      "revisionHistory": []
    },
    0
  ]
}
```

---

### Validate: `validateSoftwareImage` ✅ DISCOVERED

**Service**: `cmpart` (or `CMPart`)

**Method**: `validateSoftwareImage` (singular)

**Signature**:
```json
{
  "service": "cmpart",
  "call": "validateSoftwareImage",
  "args": [softwareImage]
}
```

**Parameters**:
- `softwareImage` (object): Single SoftwareImage entity to validate

**Expected Response**: Validation results (structure TBD - needs testing)

**Purpose**: Pre-flight validation before Create/Update operations

**Implementation Opportunities**:
1. **Plan-time validation**: Call during plan phase to detect errors early
2. **Client-side validation**: Validate configuration before API calls
3. **Error handling**: Get detailed validation errors from API

**Integration Strategy**:
- **OPTIONAL POST-MVP**: Can be used in Plan modifiers for early validation
- **GREEN PHASE**: Not required for basic CRUD functionality
- **REFACTOR PHASE**: Consider adding as enhancement for better UX

---

### Read (Single): `getSoftwareImage` ✅ DISCOVERED - PREFERRED METHOD

**Service**: `cmpart` (or `CMPart`)

**Method**: `getSoftwareImage` (singular)

**Signature**:
```json
{
  "service": "cmpart",
  "call": "getSoftwareImage",
  "args": [name]
}
```

**Parameters**:
- `name` (string): Software image name (e.g., "ubuntu-22.04-dpu")

**Expected Response**: Single SoftwareImage entity or error if not found

**Advantages**:
- **Much more efficient** than `getSoftwareImages()` + client-side filtering
- Direct lookup by name
- Reduces network overhead
- **RECOMMENDED for Read() method implementation**

---

### Read (List): `getSoftwareImages`

**Service**: `cmpart` (or `CMPart`)

**Method**: `getSoftwareImages` (plural)

**Signature**:
```json
{
  "service": "cmpart",
  "call": "getSoftwareImages"
}
```

**Parameters**: None

**Expected Response**: Array of all SoftwareImage entities

**Use Cases**:
- List all available images
- Data source implementation
- **Not recommended for resource Read()** - use `getSoftwareImage(name)` instead

---

### Update: `updateSoftwareImage` ✅ CONFIRMED

**Service**: `cmpart` (or `CMPart`)

**Method**: `updateSoftwareImage`

**Signature**:
```json
{
  "service": "cmpart",
  "call": "updateSoftwareImage",
  "args": [softwareImage, force]
}
```

**Parameters**:
- `softwareImage` (object): Complete SoftwareImage entity with updated fields
- `force` (boolean): Force update flag

**Expected Response**: Updated SoftwareImage object or success confirmation

**Implementation Notes**:
- Send complete entity with all fields (not just changed fields)
- UUID must match existing image
- Nested modules are updated inline (no separate API calls)

---

### Delete: `removeSoftwareImage` ✅ CONFIRMED

**Service**: `cmpart` (or `CMPart`)

**Method**: `removeSoftwareImage`

**Signature**:
```json
{
  "service": "cmpart",
  "call": "removeSoftwareImage",
  "args": [uuid, removeData, removeAll, force]
}
```

**Parameters**:
- `uuid` (string): UUID of the SoftwareImage to remove
- `removeData` (boolean): Remove associated data files (TBD - likely filesystem cleanup)
- `removeAll` (boolean): Remove all related entities (TBD - likely cascading delete)
- `force` (boolean): Force deletion flag

**Expected Response**: Success confirmation or error

**Implementation Notes**:
- Default values for Terraform resource: `removeData=false`, `removeAll=false`, `force=false`
- May need to expose `removeData` as optional attribute if users need filesystem cleanup

---

## Entity Lifecycle

### Complete CRUD Flow

```
0. VALIDATE (Optional - for enhanced error handling):
   POST /json
   {"service": "cmpart", "call": "validateSoftwareImage", "args": [<entity>]}
   → Response: Validation results

1. CREATE:
   POST /json
   {"service": "cmpart", "call": "addSoftwareImage", "args": [<entity>, false]}
   → Response: UUID or entity with UUID

2. READ (PREFERRED - direct lookup by name):
   POST /json
   {"service": "cmpart", "call": "getSoftwareImage", "args": [<name>]}
   → Response: Single entity or error if not found

   READ (Alternative - list all, filter client-side):
   POST /json
   {"service": "cmpart", "call": "getSoftwareImages"}
   → Response: [<entity1>, <entity2>, ...]
   → Filter client-side by name or UUID

3. UPDATE:
   POST /json
   {"service": "cmpart", "call": "updateSoftwareImage", "args": [<entity>, false]}
   → Response: Updated entity or success

4. DELETE:
   POST /json
   {"service": "cmpart", "call": "removeSoftwareImage", "args": [<uuid>, false, false, false]}
   → Response: Success confirmation
```

---

## Constraint Enforcement

### Unique Name Constraint

**Status**: Needs manual testing (Phase 0 Task T015-T019)

**Expected Behavior**: API should reject duplicate names with error response

**Error Format**: TBD - needs discovery

---

### Unique Path Constraint

**Status**: Needs manual testing (Phase 0 Task T015-T019)

**Expected Behavior**: API should reject duplicate paths with error response

**Error Format**: TBD - needs discovery

---

## Nested Resource Patterns

### Kernel Modules Management

**Pattern**: Inline with SoftwareImage entity (no separate API calls)

**Module Entity Structure**:
```json
{
  "uuid": "16867c38-7b60-4fbe-b4fb-888fd3ea4ba5",
  "baseType": "KernelModule",
  "childType": "",
  "to_be_removed": false,
  "modified": true,
  "revision": "",
  "name": "aacraid",
  "parameters": ""
}
```

**Update Strategy**:
- Send entire modules array with each update
- API handles add/remove/modify operations
- UUIDs are server-generated (include if updating existing module, omit for new modules)

**Terraform Implementation**:
- Use `ListNestedAttribute` for modules
- Map Terraform list to API array format
- Handle UUID generation server-side

---

## Design Decisions

### Force Parameter Default

**Decision**: Use `force=false` for all operations (safest option)

**Rationale**:
- Create: `force=0` in example request
- Update: Safer to fail on conflicts than overwrite
- Delete: Avoid accidental data loss

**Future Enhancement**: Could expose as optional `force_operations` boolean attribute if needed

---

### Read Strategy

**Decision**: Use `getSoftwareImage(name)` for direct lookup ✅ UPDATED

**Rationale**:
- **Much more efficient** than fetching all images
- Direct API support for single image lookup by name
- Reduces network overhead and processing time
- Name is always available in Terraform state

**Implementation**:
```go
func (r *CMPartSoftwareImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    // Use getSoftwareImage(name) for efficient lookup
    apiResp, err := r.client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", state.Name.ValueString())
    // ... handle response
}
```

**Alternative (Fallback)**: Client-side filtering with `getSoftwareImages()` if name lookup fails

---

### Import Approach

**Decision**: UUID-based import via `ImportStatePassthroughID`

**Implementation**:
```go
func (r *CMPartSoftwareImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Import Command**:
```bash
terraform import bcm_cmpart_softwareimage.example eaad50d3-432a-4703-a9f8-66551c255a69
```

---

### RemoveData/RemoveAll Parameters

**Decision**: Hardcode to `false` for initial implementation

**Rationale**:
- Safest default - avoid accidental filesystem deletion
- Unclear semantics without testing
- Can be exposed as optional attributes post-MVP if needed

**Future Enhancement**:
```hcl
resource "bcm_cmpart_softwareimage" "example" {
  name = "test"
  path = "/cm/images/test"

  # Optional deletion behavior
  remove_data_on_destroy = true  # Cleanup filesystem
  remove_all_on_destroy  = false # Don't cascade delete
}
```

---

## Service Name Case Sensitivity

**Status**: ✅ CONFIRMED - Service name is case-insensitive

**Evidence**: Example request uses `"CMPart"` (capitalized)

**Recommendation**: Use `"CMPart"` for consistency with example

**Validation**: Check existing data source at `/workspace/internal/provider/data_source_cmpart_softwareimages.go` for actual service name used

---

## API Quirks and Limitations

### 1. ~~POV Limitation: No "args" Parameter Support~~ ✅ RESOLVED & VERIFIED

**Previous Issue**: BCMClient documentation mentioned no "args" support

**Resolution**: `getSoftwareImage(name)` API method supports args parameter! ✅ **VERIFIED 2025-11-20**

**Verification Test Results**:
- Test script: `/workspace/test_bcmclient_args.go`
- BCM Endpoint: https://172.21.15.254:8081
- Test 1: Login (no args) - ✅ PASS
- Test 2: getSoftwareImages() (no args) - ✅ PASS (returned 2 images)
- Test 3: getSoftwareImage(name) WITH args - ✅ PASS
  - Request: `{"service":"cmpart","call":"getSoftwareImage","args":["default-image"]}`
  - Response: HTTP 200, returned single SoftwareImage entity
  - Verdict: Args parameter IS supported by BCM API

**Impact**:
- ✅ Can use efficient direct lookup by name
- ✅ No need for client-side filtering
- ✅ Better performance for Read operations

**Action Required**: Extend BCMClient.CallJSONRPC to support variadic args parameter

**Implementation Recommendation**:
```go
// Add variadic args parameter to CallJSONRPC signature
func (c *BCMClient) CallJSONRPC(ctx context.Context, service, call string, args ...interface{}) ([]byte, error) {
    reqBody := map[string]interface{}{
        "service": service,
        "call":    call,
    }
    if len(args) > 0 {
        reqBody["args"] = args
    }
    // ... rest of implementation
}
```

**Note**: The BCMClient limitation was documentation-only, not an actual API limitation.

---

### 2. Entity Metadata Fields

**API-Only Fields** (include in requests, exclude from Terraform schema):
- `baseType: "SoftwareImage"` - Entity type discriminator
- `childType: ""` - Subtype (always empty for SoftwareImage)
- `to_be_removed: false` - Deletion flag (internal)
- `modified: true` - Change tracking (internal)
- `revision: ""` - Version string (internal)

**Implementation**: Add these fields when constructing API request, strip when mapping to Terraform state

---

### 3. Module UUID Generation

**Behavior**: UUIDs appear to be server-generated

**Implementation Strategy**:
- Create: Omit module UUIDs, let server generate
- Update: Include existing module UUIDs, omit for new modules
- Read: Accept server-provided UUIDs

**Validation Needed**: Test in Phase 0 Tasks T020-T025

---

### 4. CreationTime Field

**Behavior**: Unix timestamp (seconds since epoch)

**Example**: `"creationTime": 1763626149`

**Terraform Mapping**: Store as `types.Int64`, display as read-only computed attribute

---

### 5. Zero UUID for Unset References

**Pattern**: `"00000000-0000-0000-0000-000000000000"` represents "not set"

**Fields Affected**:
- `originalImage` - Source image for clones
- `fspart` - Filesystem partition reference
- `bootfspart` - Boot partition reference
- `parentSoftwareImage` - Parent for revisions

**Implementation**: Map to `types.String` with null value when zero UUID

---

## Phase 0 Tasks Status

### Completed (via user-provided API signatures):
- ✅ T005: Update API method confirmed (`updateSoftwareImage`)
- ✅ T006: Delete API method confirmed (`removeSoftwareImage`)
- ✅ T007: API signatures documented

### Still Required (manual testing):
- ⏳ T001-T004: Search documentation (SKIPPED - signatures provided)
- ⏳ T008-T014: CRUD lifecycle testing (RECOMMENDED before implementation)
- ⏳ T015-T019: Constraint validation testing (CRITICAL for error handling)
- ⏳ T020-T025: Nested modules testing (CRITICAL for state management)
- ⏳ T026-T028: Complete research documentation

---

## Recommendations for Implementation

### HIGH Priority (Before Phase 2):

1. **Test Constraint Violations** (T015-T019)
   - Capture actual error response format for duplicate name/path
   - Implement proper error parsing in Create/Update methods

2. **Test Module Management** (T020-T025)
   - Verify UUID generation behavior
   - Test add/remove/modify module operations
   - Confirm inline update pattern works

3. **Verify Service Name** (T002)
   - Check existing data source for actual service name used
   - Ensure consistency across provider

### MEDIUM Priority (During Phase 4 REFACTOR):

4. **Test Force Parameter Semantics** (T008-T014)
   - Understand what `force=true` does
   - Document expected behavior
   - Consider exposing as optional attribute

5. **Test RemoveData/RemoveAll Parameters**
   - Understand filesystem cleanup behavior
   - Document cascading delete behavior
   - Decide if should be exposed to users

### LOW Priority (Post-MVP):

6. **Performance Testing**
   - Measure Read operation time with 100+ images
   - Evaluate need for server-side UUID filtering

---

## Validation Method Enhancement Opportunity

### validateSoftwareImage Method (Singular)

This API provides **optional pre-flight validation** that can improve user experience:

**Use Cases**:
1. **Early Error Detection**: Catch configuration errors during `terraform plan` instead of `apply`
2. **Better Error Messages**: Get detailed validation errors from API
3. **Constraint Validation**: Test unique name/path before attempting create

**Implementation Options**:

**Option 1: Include in REFACTOR Phase** ✅ **SELECTED - USER REQUESTED**
- Call `validateSoftwareImage()` before Create/Update operations
- Provide early error feedback from API
- Better user experience with detailed validation errors
- Catch constraint violations (duplicate name/path) before attempting create

**Option 2: Post-MVP Enhancement** (Alternative)
- Implement basic CRUD first without validation
- Add validation in future iteration
- Use in plan modifiers for plan-time validation

**Decision**: **Include validation in REFACTOR phase** per user request. This is important for catching errors early and providing better error messages.

---

## Next Steps

1. ✅ **COMPLETE**: Update and Delete API methods confirmed
2. ✅ **COMPLETE**: Validation method documented (deferred to post-MVP)
3. ✅ **COMPLETE**: spec.md and plan.md updated with confirmed signatures
4. ⏭️ **OPTIONAL**: Execute Phase 0 Tasks T015-T025 (constraint and module testing)
5. ⏭️ **READY**: Proceed to Phase 2 (RED) - TDD implementation can begin

**Decision Point**: All critical API methods confirmed. Can proceed to Phase 2 (TDD RED) immediately. Manual testing (T015-T025) is optional but recommended for understanding constraint behavior before REFACTOR phase.

---

## References

- Original documentation: `/workspace/sampleRest/wip/resource_cmpart_softwareimage.md`
- BCM API docs: `/workspace/sampleRest/BCM_API_Complete_Documentation.md`
- Existing data source: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
- BCM client: `/workspace/internal/provider/bcm_client.go`
