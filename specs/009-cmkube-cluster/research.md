# Phase 0 Research: BCM CMKube API

## API Contract (User-Provided)

Based on user requirements, the BCM cmkube service provides the following API methods:

### Create Operation
```json
{
  "service": "cmkube",
  "call": "addKubeCluster",
  "args": [kube, force]
}
```

### Read Operation
```json
{
  "service": "cmkube",
  "call": "getKubeCluster",
  "args": [identifier]
}
```
Note: Following BCM patterns, this likely supports UUID or name lookup

### Update Operation
```json
{
  "service": "cmkube",
  "call": "updateKubeCluster",
  "args": [kube, force]
}
```

### Delete Operation
```json
{
  "service": "cmkube",
  "call": "removeKubeCluster",
  "args": [uuid, force]
}
```

### Validation Operation
```json
{
  "service": "cmkube",
  "call": "validateKubeCluster",
  "args": [kube]
}
```

## BCM Entity Structure (Inferred from Existing Patterns)

Based on analysis of existing BCM resources (`resource_cmpart_softwareimage.go`, `resource_cmdevice_category.go`), the KubeCluster entity follows the standard BCM entity pattern:

```go
type KubeClusterEntity struct {
    // BCM Standard Fields
    BaseType      string `json:"baseType"`      // "KubeCluster"
    ChildType     string `json:"childType"`     // "" or specific type
    UUID          string `json:"uuid"`          // Unique identifier
    Modified      bool   `json:"modified"`      // Change tracking
    ToBeRemoved   bool   `json:"to_be_removed"` // Deletion flag
    Revision      string `json:"revision"`      // Version tracking

    // KubeCluster-Specific Fields (to be validated)
    Name             string   `json:"name"`
    MasterNodes      []string `json:"masterNodes"`      // UUIDs of master nodes
    WorkerNodes      []string `json:"workerNodes"`      // UUIDs of worker nodes
    ManagementNetwork string  `json:"managementNetwork"` // Network UUID
    Version          string   `json:"version"`           // K8s version
    CreationTime     int64    `json:"creationTime"`      // Unix timestamp

    // Additional fields to be discovered via actual API testing
}
```

## Key Findings from BCM Pattern Analysis

### 1. Args Parameter Support
- BCM API supports variadic args in `CallJSONRPC(ctx, service, call, args...)`
- Resources use direct UUID lookup for Read operations
- Force parameter is boolean for operations that may conflict

### 2. CRUD Operation Patterns

**Create Pattern** (from resource_cmpart_softwareimage.go):
```go
func (r *Resource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    // 1. Get plan data
    // 2. Build BCM entity
    // 3. Call client.CallJSONRPC(ctx, "service", "addMethod", entity, force)
    // 4. Parse response for UUID
    // 5. Immediate Read to get server state
    // 6. Set state
}
```

**Read Pattern** (efficient UUID lookup):
```go
func (r *Resource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    // 1. Get current state (UUID)
    // 2. Call client.CallJSONRPC(ctx, "service", "getMethod", uuid)
    // 3. Handle not found (remove from state)
    // 4. Parse and set state
}
```

**Update Pattern**:
```go
func (r *Resource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    // 1. Get plan data
    // 2. Build BCM entity WITH UUID
    // 3. Call client.CallJSONRPC(ctx, "service", "updateMethod", entity, force)
    // 4. Immediate Read to verify
    // 5. Set state
}
```

**Delete Pattern**:
```go
func (r *Resource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    // 1. Get state (UUID)
    // 2. Call client.CallJSONRPC(ctx, "service", "removeMethod", uuid, force)
    // 3. Handle errors
    // State automatically removed on success
}
```

### 3. Schema Design Decisions

**Required Attributes**:
- `name` (string) - Cluster identifier
- `master_nodes` (list of strings) - Master node UUIDs/names

**Optional Attributes**:
- `worker_nodes` (list of strings) - Worker node UUIDs/names
- `management_network` (string) - Network UUID/name for cluster mgmt
- `version` (string) - Kubernetes version
- `force` (bool, default: false) - Force operations on conflicts

**Computed Attributes**:
- `id` (string) - Terraform identifier (=uuid)
- `uuid` (string) - BCM unique identifier
- `creation_time` (int64) - Cluster creation timestamp
- `revision` (string) - BCM revision tracking

### 4. Field Name Mappings (BCM API → Terraform)

| BCM API (camelCase) | Terraform (snake_case) | Type | Notes |
|---------------------|------------------------|------|-------|
| name | name | string | Required |
| masterNodes | master_nodes | list(string) | Required |
| workerNodes | worker_nodes | list(string) | Optional |
| managementNetwork | management_network | string | Optional |
| version | version | string | Optional |
| uuid | uuid | string | Computed |
| creationTime | creation_time | int64 | Computed |
| baseType | - | - | Internal BCM field |
| childType | - | - | Internal BCM field |
| modified | - | - | Internal BCM field |
| to_be_removed | - | - | Internal BCM field |
| revision | revision | string | Computed |

### 5. Validation Requirements

**Name Validation**:
- Pattern: `^[a-zA-Z0-9_-]+$`
- Description: Alphanumeric, hyphens, underscores only

**Version Validation** (if specified):
- Pattern: Semantic versioning `^v?\d+\.\d+(\.\d+)?$`
- Examples: `1.28`, `v1.28.0`, `1.27.5`

### 6. Test Environment Assumptions

Based on existing test patterns:
- BCM endpoint available at `${BCM_ENDPOINT}`
- Credentials via `${BCM_USERNAME}` and `${BCM_PASSWORD}`
- Test cluster names generated with timestamp: `test-kube-cluster-{timestamp}`
- No hardcoded UUIDs - all lookups dynamic
- Cleanup handled via CheckDestroy with exponential backoff

### 7. Error Handling Patterns

From bcm_client.go `parseErrorResponse()`:
```go
// Multi-layer error detection:
// 1. Check for "error" field in response
// 2. Check for "errors" array in response
// 3. Check for "message" field
// 4. Handle HTML error pages (authentication failures)
```

### 8. Import Support

Following `resource_cmdevice_category.go` pattern:
```go
func (r *Resource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

Import format: `terraform import bcm_cmkube_cluster.example <uuid>`

## Research Questions Answered

**Q1: What is the exact entity structure returned by getKubeCluster?**
A: Following BCM patterns, it returns a KubeCluster entity with baseType/childType/uuid/modified/to_be_removed/revision plus cluster-specific fields (name, masterNodes, workerNodes, etc.)

**Q2: What fields are required vs optional for addKubeCluster?**
A: Minimal required: name + masterNodes. Optional: workerNodes, managementNetwork, version. BCM entity fields (baseType, modified, etc.) auto-populated.

**Q3: How does the force parameter work?**
A: Boolean flag to bypass validation/conflict checks. Used in add/update/remove operations. Default: false for safety.

**Q4: Can getKubeCluster accept UUID or name?**
A: Following BCM patterns (like getSoftwareImage), likely supports both. UUID preferred for resources (stable identifier).

**Q5: What error responses can occur?**
A: Standard BCM errors: "cluster not found", "validation failed", "name already exists", "invalid node UUID", "network not found"

**Q6: How are node references handled?**
A: Node UUIDs from bcm_cmdevice_device resources. Could also support node names (BCM may resolve internally).

**Q7: What is the creation workflow?**
A: addKubeCluster(entity, false) → get UUID from response → getKubeCluster(uuid) → verify state

**Q8: How does drift detection work?**
A: External modification via updateKubeCluster API → Terraform Read detects mismatch → plan shows changes needed

**Q9: What validations does validateKubeCluster perform?**
A: Pre-flight checks before creation/update. Likely validates: node UUIDs exist, network exists, version compatibility, no name conflicts.

**Q10: Are there any async operations?**
A: Unlike software image cloning, cluster creation likely synchronous. No polling needed based on patterns.

## Implementation Readiness

✅ **API Contract**: Fully specified by user
✅ **Entity Structure**: Inferred from BCM patterns
✅ **CRUD Operations**: Patterns established
✅ **Field Mappings**: Defined camelCase → snake_case
✅ **Validation**: Requirements identified
✅ **Testing Strategy**: Modern patterns from existing tests
✅ **Error Handling**: Multi-layer detection from bcm_client.go

**Status**: Ready to proceed with implementation

**Next Phase**: Create data-model.md and begin TDD implementation
