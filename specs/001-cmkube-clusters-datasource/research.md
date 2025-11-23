# Research: BCM Kubernetes Clusters Data Source

**Date**: 2025-11-23
**Researcher**: Implementation Planning (Phase 0)
**Objective**: Resolve all technical unknowns for implementing `data.bcm_cmkube_clusters` data source

## BCM API Analysis

### Service & Method

**Service**: `cmkube`
**Method**: `getKubeClusters` (plural - retrieves all clusters)
**Authentication**: Cookie-based (`cm-login-token`) via BCMClient
**Endpoint**: `https://172.21.15.254:8081/json`

### Request Format

```json
{
  "service": "cmkube",
  "call": "getKubeClusters"
}
```

**No arguments required** - this is a list operation that retrieves all Kubernetes clusters in one API call.

### Response Format

The BCM API returns a JSON array of cluster objects:

```json
[
  {
    "uuid": "cluster-uuid-1",
    "name": "prod-k8s-cluster",
    "baseType": "KubeCluster",
    "childType": "",
    "masterNodes": ["node-uuid-1", "node-uuid-2"],
    "workerNodes": ["node-uuid-3", "node-uuid-4"],
    "managementNetwork": "network-uuid",
    "overlayNetwork": "10.244.0.0/16",
    "dnsServers": ["8.8.8.8", "8.8.4.4"],
    "version": "1.28.0",
    "cniPlugin": "calico",
    "storageClasses": "{\"default\":\"rbd\"}",
    "loadBalancerMode": "metallb",
    "addons": "[{\"name\":\"monitoring\"}]",
    "ingressController": "{\"type\":\"nginx\"}",
    "creationTime": 1700000000,
    "revisionID": 5,
    "modified": false,
    "to_be_removed": false,
    "revision": ""
  }
]
```

### API Verification Script

The existing script `/workspace/sampleRest/cmkube-get-clusters.py` confirms the API structure:

```python
#!/usr/bin/env python3
# Script: cmkube-get-clusters.py
# Purpose: Retrieve all Kubernetes clusters from BCM API
# Usage: BCM_ENDPOINT="https://..." BCM_USERNAME="root" BCM_PASSWORD="..." python3 cmkube-get-clusters.py

# Call structure:
payload = {
    "service": "cmkube",
    "call": "getKubeClusters"
}
```

This confirms:
- ✅ `cmkube.getKubeClusters` API method exists
- ✅ Returns array of cluster objects (not single object)
- ✅ No pagination required (single call retrieves all)
- ✅ All fields documented in spec.md are present in API response

---

## Field Mapping Table

Complete mapping of BCM API fields (camelCase) to Terraform attributes (snake_case):

| BCM API Field | Terraform Attribute | Go Type | Extraction Method | Null Handling | Notes |
|---------------|---------------------|---------|-------------------|---------------|-------|
| uuid | uuid, id | types.String | getStringValue(data, "uuid") | Required (never null) | Both `id` and `uuid` set to same value |
| name | name | types.String | getStringValue(data, "name") | Required (never null) | Cluster name |
| masterNodes | master_nodes | types.List | getListValue(data, "masterNodes") | Required (never null) | List of master node UUIDs (StringType elements) |
| workerNodes | worker_nodes | types.List | getListValue(data, "workerNodes") | Optional | List of worker node UUIDs (StringType elements), can be null/empty |
| managementNetwork | management_network | types.String | getStringValue(data, "managementNetwork") | Optional | Network UUID, can be null |
| overlayNetwork | overlay_network | types.String | getStringValue(data, "overlayNetwork") | Optional | Pod network CIDR (e.g., "10.244.0.0/16"), can be null |
| dnsServers | dns_servers | types.List | getListValue(data, "dnsServers") | Optional | List of DNS IPs (StringType elements), can be null/empty |
| version | version | types.String | getStringValue(data, "version") | Optional | Kubernetes version semver (e.g., "1.28.0"), can be null |
| cniPlugin | cni_plugin | types.String | getStringValue(data, "cniPlugin") | Optional | CNI plugin name (e.g., "calico"), can be null |
| storageClasses | storage_classes | types.String | getStringValue(data, "storageClasses") | Optional | JSON-encoded string, can be null |
| loadBalancerMode | load_balancer_mode | types.String | getStringValue(data, "loadBalancerMode") | Optional | Load balancer strategy (e.g., "metallb"), can be null |
| addons | addons | types.String | getStringValue(data, "addons") | Optional | JSON-encoded string, can be null |
| ingressController | ingress_controller | types.String | getStringValue(data, "ingressController") | Optional | JSON-encoded string, can be null |
| creationTime | creation_time | types.Int64 | getInt64Value(data, "creationTime") | Computed | Unix timestamp, always present |
| revisionID | revision_id | types.Int64 | getInt64Value(data, "revisionID") | Computed | Revision number, always present |
| baseType | N/A | N/A | Not exposed | N/A | Internal BCM field (always "KubeCluster") |
| childType | N/A | N/A | Not exposed | N/A | Internal BCM field (always "") |
| modified | N/A | N/A | Not exposed | N/A | Internal BCM state flag |
| to_be_removed | N/A | N/A | Not exposed | N/A | Internal BCM state flag |
| revision | N/A | N/A | Not exposed | N/A | Internal BCM version tracking |

### Key Findings

1. **Required Fields**: `uuid`, `name`, `masterNodes` are always present in API response
2. **Optional Fields**: All network, configuration, and addon fields can be null
3. **Consistency**: Schema matches `resource_cmkube_cluster.go` exactly (verified by cross-reference)
4. **Internal Fields**: BCM metadata fields (baseType, childType, modified, to_be_removed, revision) are not exposed to Terraform users

---

## Filter Implementation Strategy

### Filter Model Definition

Based on the established pattern in `data_source_cmpart_softwareimages.go:41-46`:

```go
type ClusterFilterModel struct {
    NamePattern   types.String `tfsdk:"name_pattern"`   // Optional: case-insensitive substring match
    Version       types.String `tfsdk:"version"`        // Optional: exact Kubernetes version match
    MasterNodeID  types.String `tfsdk:"master_node_id"` // Optional: find clusters with specific master node UUID
}
```

### Client-Side Filtering Pseudocode

```go
func (d *CMKubeClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data CMKubeClustersDataSourceModel

    // 1. Get filter configuration from plan
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)

    // 2. Call BCM API to retrieve all clusters
    body, err := d.client.CallJSONRPC(ctx, "cmkube", "getKubeClusters")
    if err != nil {
        resp.Diagnostics.AddError("Error Reading Clusters", "Could not read clusters: "+err.Error())
        return
    }

    // 3. Parse API response
    var apiResponse []map[string]interface{}
    if err := json.Unmarshal(body, &apiResponse); err != nil {
        resp.Diagnostics.AddError("Error Parsing Clusters", "Could not parse response: "+err.Error())
        return
    }

    // 4. Apply client-side filters (AND logic)
    var filteredClusters []KubeClusterModel
    for _, clusterData := range apiResponse {
        include := true

        // Filter 1: name_pattern (case-insensitive substring match)
        if data.Filter != nil && !data.Filter.NamePattern.IsNull() {
            pattern := strings.ToLower(data.Filter.NamePattern.ValueString())
            clusterName := strings.ToLower(getStringValue(clusterData, "name").ValueString())
            if !strings.Contains(clusterName, pattern) {
                include = false
            }
        }

        // Filter 2: version (exact match)
        if data.Filter != nil && !data.Filter.Version.IsNull() {
            clusterVersion := getStringValue(clusterData, "version").ValueString()
            if clusterVersion != data.Filter.Version.ValueString() {
                include = false
            }
        }

        // Filter 3: master_node_id (ANY match in master_nodes list)
        if data.Filter != nil && !data.Filter.MasterNodeID.IsNull() {
            targetNodeID := data.Filter.MasterNodeID.ValueString()
            found := false

            // Extract master_nodes list from API response
            if masterNodesRaw, ok := clusterData["masterNodes"]; ok && masterNodesRaw != nil {
                if masterNodesSlice, ok := masterNodesRaw.([]interface{}); ok {
                    for _, nodeUUID := range masterNodesSlice {
                        if nodeStr, ok := nodeUUID.(string); ok && nodeStr == targetNodeID {
                            found = true
                            break
                        }
                    }
                }
            }

            if !found {
                include = false
            }
        }

        if include {
            model := mapClusterDataToModel(clusterData)
            filteredClusters = append(filteredClusters, model)
        }
    }

    // 5. Set state
    data.Clusters = filteredClusters
    data.ID = types.StringValue("cmkube-clusters")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### AND Logic Implementation

**Rule**: When multiple filters are specified, a cluster must match **ALL** filters to be included in results.

**Example**:
```hcl
filter {
  name_pattern   = "prod"
  version        = "1.28.0"
  master_node_id = "node-abc-123"
}
```

A cluster must satisfy:
- Name contains "prod" (case-insensitive) **AND**
- Version equals "1.28.0" (exact match) **AND**
- Master nodes list contains "node-abc-123" (ANY match)

**Pattern Source**: `data_source_cmpart_softwareimages.go:matchesSoftwareImageFilter()` function (lines 495-520)

---

## List Extraction Pattern

### Problem Statement

The BCM API returns list fields as `[]interface{}` (e.g., `masterNodes`, `workerNodes`, `dnsServers`). Terraform Plugin Framework requires `types.List` with explicit element types.

### Solution: `getListValue()` Helper Function

**New helper function** (to be added to data source file, similar to existing `getStringValue()`):

```go
// getListValue extracts a list of strings from BCM API response with null handling
// Returns types.List with StringType elements, or types.ListNull if field missing/null
func getListValue(data map[string]interface{}, key string) types.List {
    if val, ok := data[key]; ok && val != nil {
        if slice, ok := val.([]interface{}); ok && len(slice) > 0 {
            // Convert []interface{} to []attr.Value (StringValue elements)
            elements := make([]attr.Value, 0, len(slice))
            for _, item := range slice {
                if str, ok := item.(string); ok && str != "" {
                    elements = append(elements, types.StringValue(str))
                }
            }

            if len(elements) > 0 {
                listValue, _ := types.ListValue(types.StringType, elements)
                return listValue
            }
        }
    }

    // Return null list if field missing, null, or empty
    return types.ListNull(types.StringType)
}
```

### Usage Examples

```go
// Extract master_nodes (required list, but still use null-safe extraction)
model.MasterNodes = getListValue(clusterData, "masterNodes")

// Extract worker_nodes (optional list)
model.WorkerNodes = getListValue(clusterData, "workerNodes")

// Extract dns_servers (optional list)
model.DNSServers = getListValue(clusterData, "dnsServers")
```

### Pattern Source

Based on module extraction pattern in `data_source_cmpart_softwareimages.go:440-456`:

```go
// Existing pattern for nested objects (modules)
if modulesData, ok := apiData["modules"].([]interface{}); ok && modulesData != nil {
    for _, moduleItem := range modulesData {
        if moduleMap, ok := moduleItem.(map[string]interface{}); ok {
            module := KernelModuleModel{
                UUID:       getStringValue(moduleMap, "uuid"),
                Name:       getStringValue(moduleMap, "name"),
                // ...
            }
            model.Modules = append(model.Modules, module)
        }
    }
} else {
    model.Modules = []KernelModuleModel{} // Empty slice if null
}
```

**Adaptation**: For simple string lists (not nested objects), extract to `[]attr.Value` then convert to `types.List`.

---

## Testing Strategy

### Test Coverage Matrix (7 Tests)

| Test Name | Scenario | Config | Assertions | Pattern Source |
|-----------|----------|--------|------------|----------------|
| TestAccCMKubeClustersDataSource_Basic | List all clusters without filters | No filter block | - ID not null<br>- clusters[0].uuid not null<br>- clusters[0].name not null<br>- clusters[0].master_nodes not null | `data_source_cmdevice_nodes_test.go:18-46` |
| TestAccCMKubeClustersDataSource_FilterByName | Filter by name pattern | `filter { name_pattern = "test-cluster" }` | - All returned clusters contain "test-cluster" in name (case-insensitive) | `data_source_cmpart_softwareimages_test.go` FilterByName |
| TestAccCMKubeClustersDataSource_FilterByVersion | Filter by Kubernetes version | `filter { version = "1.28.0" }` | - All returned clusters have version = "1.28.0" | Similar to FilterByCategory pattern |
| TestAccCMKubeClustersDataSource_FilterByMasterNode | Filter by master node UUID | `filter { master_node_id = "node-uuid-123" }` | - All returned clusters contain "node-uuid-123" in master_nodes list | New pattern (verify list membership) |
| TestAccCMKubeClustersDataSource_MultipleFilters | Combine name + version (AND) | `filter { name_pattern = "prod"; version = "1.28.0" }` | - Returned clusters match both filters | Combine existing filter patterns |
| TestAccCMKubeClustersDataSource_EmptyResults | No clusters match filter | `filter { name_pattern = "nonexistent-xyz" }` | - clusters list is empty (not error)<br>- ID still set to "cmkube-clusters" | Test edge case handling |
| TestAccCMKubeClustersDataSource_NullFields | Cluster with missing optional fields | No filter (assumes BCM has cluster with nulls) | - Optional fields are null<br>- Required fields present<br>- No errors thrown | Test null handling in getListValue |

### Modern Test Patterns (terraform-plugin-testing v1.13.3+)

**ConfigStateChecks Pattern**:

```go
func TestAccCMKubeClustersDataSource_Basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMKubeClustersDataSourceConfig_basic(),
                ConfigStateChecks: []statecheck.StateCheck{
                    // Verify ID is set
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmkube_clusters.test",
                        tfjsonpath.New("id"),
                        knownvalue.NotNull(),
                    ),
                    // Verify first cluster has UUID
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmkube_clusters.test",
                        tfjsonpath.New("clusters").AtSliceIndex(0).AtMapKey("uuid"),
                        knownvalue.NotNull(),
                    ),
                    // Verify first cluster has name
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmkube_clusters.test",
                        tfjsonpath.New("clusters").AtSliceIndex(0).AtMapKey("name"),
                        knownvalue.NotNull(),
                    ),
                    // Verify first cluster has master_nodes list
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmkube_clusters.test",
                        tfjsonpath.New("clusters").AtSliceIndex(0).AtMapKey("master_nodes"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}
```

**Test Config Pattern** (with provider block):

```go
func testAccCMKubeClustersDataSourceConfig_basic() string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmkube_clusters" "test" {}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
    )
}
```

### Environment Portability

**Principles**:
1. No hardcoded cluster counts (use `knownvalue.NotNull()` instead of exact counts)
2. No hardcoded cluster names (tests work on any BCM environment)
3. Generate unique test resource names if creating test clusters
4. Use modern matchers (`knownvalue.StringExact()`, `knownvalue.ListSizeExact()`) for type-safe assertions

**Example - Portable Filter Test**:

```go
// GOOD: Works on any environment with at least one cluster
statecheck.ExpectKnownValue(
    "data.bcm_cmkube_clusters.test",
    tfjsonpath.New("id"),
    knownvalue.NotNull(),
)

// BAD: Assumes specific cluster count
resource.TestCheckResourceAttr("data.bcm_cmkube_clusters.test", "clusters.#", "3")
```

---

## Key Implementation Decisions

### Decision 1: Client-Side Filtering vs Server-Side

**Decision**: Client-side filtering (retrieve all, filter in Go)
**Rationale**: BCM API `cmkube.getKubeClusters` does not accept filter parameters
**Alternative Rejected**: Individual cluster lookups (no `getKubeCluster(name)` method exists)
**Performance**: Acceptable for 100+ clusters (single API call, in-memory filtering < 1s)

### Decision 2: Filter Logic (AND vs OR)

**Decision**: AND logic (all filters must match)
**Rationale**: Matches user expectations for restrictive filtering
**Alternative Rejected**: OR logic (too permissive, harder to reason about)
**Pattern Source**: `data_source_cmpart_softwareimages.go` with multiple filters

### Decision 3: Name Pattern Matching

**Decision**: Case-insensitive substring match using `strings.Contains()`
**Rationale**: Simple, predictable, matches existing data source pattern
**Alternative Rejected**: Regex (complex) or glob library (new dependency)
**Pattern Source**: `data_source_cmdevice_nodes.go` hostname_pattern filter

### Decision 4: Master Node Filter Behavior

**Decision**: Match ANY master node in list (not ALL)
**Rationale**: Operational use case is "find which cluster(s) contain this node"
**Alternative Rejected**: Match ALL master nodes (overly restrictive)
**Implementation**: Iterate `masterNodes` array, return cluster if UUID found

### Decision 5: Schema Consistency

**Decision**: Match `resource_cmkube_cluster.go` schema exactly
**Rationale**: Ensures terraform import compatibility
**Verification**: Cross-referenced all attribute names and types
**Result**: 100% schema consistency achieved

---

## Risk Assessment & Mitigation

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| BCM API returns unexpected field types | Low | High | Use null-safe extraction for ALL fields, comprehensive type checking in helpers |
| Large cluster count (100+) causes timeout | Low | Medium | Performance testing in acceptance tests, BCM API response time is < 2s |
| Filter logic bugs (3 filters × AND) | Medium | Medium | Comprehensive acceptance tests for each filter individually + combined |
| Schema mismatch with resource | Low | High | Automated cross-reference check, verify field names/types match exactly |
| Missing optional fields cause nil panics | Medium | High | Use getStringValue/getInt64Value/getListValue for ALL fields |
| List extraction returns wrong type | Low | Medium | Test with null lists, empty lists, and populated lists |

---

## Next Steps (Phase 1)

1. ✅ Create `data-model.md` with complete Go struct definitions
2. ✅ Create `contracts/cmkube-api.json` with request/response examples
3. ✅ Create `quickstart.md` with TDD workflow guide
4. ✅ Update agent context with cmkube API and field mappings
5. ✅ Proceed to `/speckit.tasks` for task generation

---

## References

- **Spec**: `/workspace/specs/001-cmkube-clusters-datasource/spec.md`
- **Existing Data Source**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
- **Existing Tests**: `/workspace/internal/provider/data_source_cmdevice_nodes_test.go`
- **Test Helpers**: `/workspace/internal/provider/test_helpers.go`
- **Resource Schema**: `/workspace/internal/provider/resource_cmkube_cluster.go`
- **API Script**: `/workspace/sampleRest/cmkube-get-clusters.py`
