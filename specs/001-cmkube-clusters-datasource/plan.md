# Implementation Plan: BCM Kubernetes Clusters Data Source

**Branch**: `001-cmkube-clusters-datasource` | **Date**: 2025-11-23 | **Spec**: [spec.md](/workspace/specs/001-cmkube-clusters-datasource/spec.md)

**Input**: Feature specification from `/workspace/specs/001-cmkube-clusters-datasource/spec.md`

**GitHub Issue**: #27 - Implement data.bcm_cmkube_clusters data source for listing and filtering Kubernetes clusters managed by BCM

## Summary

Implement a read-only Terraform data source (`data.bcm_cmkube_clusters`) that enables platform engineers to discover and filter BCM-managed Kubernetes clusters. The data source will call the `cmkube.getKubeClusters` API, perform client-side filtering by name pattern, version, and master node ID, and expose all cluster attributes (uuid, name, nodes, network config, Kubernetes version) for import and dependency workflows. This follows the established TDD pattern used in `data_source_cmdevice_nodes.go` and `data_source_cmpart_softwareimages.go` with comprehensive acceptance test coverage for list all, filter, and edge case scenarios.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (read-only data source)
**Testing**: Acceptance tests (TF_ACC=1), terraform-plugin-testing modern patterns (statecheck, knownvalue)
**Target Platform**: Linux server, BCM API endpoint (https://172.21.15.254:8081)
**Project Type**: Terraform Provider (single project)
**Performance Goals**: Handle 100+ clusters without timeout, client-side filtering < 1s
**Constraints**: BCM API eventual consistency (~2s), single API call retrieves all clusters (no pagination)
**Scale/Scope**: 1 data source, 7 acceptance tests, ~500 LOC implementation + tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### TDD Compliance Gates

✅ **PASS** - TDD workflow required: Implementation follows RED-GREEN-REFACTOR with acceptance tests written first
✅ **PASS** - Parallel execution: Data source can be implemented in parallel with other data sources
✅ **PASS** - Test coverage: All CRUD operations covered (Read only for data sources)
✅ **PASS** - Documentation: Auto-generated via `make generate` using tfplugindocs

### Complexity Gates

✅ **PASS** - Pattern reuse: Uses established data source pattern from `data_source_cmdevice_nodes.go` and `data_source_cmpart_softwareimages.go`
✅ **PASS** - Client-side filtering: Follows existing pattern of retrieving all entities then filtering in Go
✅ **PASS** - Null-safe extraction: Uses existing helper functions `getStringValue()`, `getInt64Value()`, custom list extraction
✅ **PASS** - No new abstractions: Reuses `BCMClient`, `FilterModel` pattern, existing test helpers

**Result**: ✅ ALL GATES PASSED - Proceed to Phase 0 research

## Project Structure

### Documentation (this feature)

```text
specs/001-cmkube-clusters-datasource/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (API research, pattern analysis)
├── data-model.md        # Phase 1 output (schema design, field mappings)
├── quickstart.md        # Phase 1 output (developer quick start)
├── contracts/           # Phase 1 output (API request/response examples)
│   └── cmkube-api.json  # BCM cmkube.getKubeClusters contract
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/provider/
│   ├── data_source_cmkube_clusters.go         # New: Data source implementation (~400 LOC)
│   ├── data_source_cmkube_clusters_test.go    # New: Acceptance tests (~350 LOC)
│   ├── test_helpers.go                        # Existing: Reuse createTestBCMClient, generateUniqueTestName
│   ├── bcm_client.go                          # Existing: BCMClient.CallJSONRPC
│   └── provider.go                            # Modified: Register NewCMKubeClustersDataSource
│
├── examples/data-sources/bcm_cmkube_clusters/  # New: HCL examples for documentation
│   └── data-source.tf                          # Examples: list all, filters, import use case
│
└── docs/data-sources/cmkube_clusters.md       # Generated: Auto-generated documentation
```

**Structure Decision**: This is a single-project Terraform provider using the standard HashiCorp Plugin Framework structure. The data source implementation follows the established pattern in `internal/provider/data_source_*.go` files with parallel test files. All existing test helpers and BCM client infrastructure are reused without modification.

## Complexity Tracking

> **No violations - table intentionally left empty**

This implementation introduces zero new patterns or abstractions. All complexity is handled by existing, proven patterns from `data_source_cmdevice_nodes.go` and `data_source_cmpart_softwareimages.go`.

---

## Phase 0: Outline & Research

**Objective**: Resolve all NEEDS CLARIFICATION items from Technical Context and research BCM API patterns, filter implementation strategies, and test patterns for Kubernetes cluster data sources.

### Research Tasks

1. **BCM API Investigation** (`research.md` - API Section)
   - **Task**: Verify `cmkube.getKubeClusters` API method exists and returns expected structure
   - **Method**: Run `sampleRest/cmkube_*.py` scripts or create test script to call API
   - **Success Criteria**: Document exact request/response JSON structure with all fields
   - **Deliverable**: API contract with field name mappings (camelCase → snake_case)

2. **Field Mapping Research** (`research.md` - Field Mappings Section)
   - **Task**: Cross-reference `resource_cmkube_cluster.go` schema with BCM API response fields
   - **Method**: Compare resource schema attributes to API response JSON keys
   - **Success Criteria**: Complete mapping table of all BCM API fields to Terraform attributes
   - **Deliverable**: Field mapping table with types and null-handling strategy

3. **Filter Pattern Analysis** (`research.md` - Filter Implementation Section)
   - **Task**: Study filter implementation in `data_source_cmpart_softwareimages.go` and `data_source_cmdevice_nodes.go`
   - **Method**: Extract common patterns for FilterModel, client-side filtering logic
   - **Success Criteria**: Document filter pattern with AND logic for multiple filters
   - **Deliverable**: Pseudocode for name_pattern, version, master_node_id filters

4. **List Extraction Pattern** (`research.md` - Helper Functions Section)
   - **Task**: Research how to extract lists from BCM API (master_nodes, worker_nodes, dns_servers)
   - **Method**: Examine existing list extraction in `data_source_cmdevice_nodes.go` (interfaces, roles)
   - **Success Criteria**: Document pattern for converting []interface{} to types.List
   - **Deliverable**: Code pattern for null-safe list extraction with element type

5. **Test Pattern Research** (`research.md` - Testing Strategy Section)
   - **Task**: Analyze modern test patterns using statecheck, knownvalue, tfjsonpath from existing tests
   - **Method**: Review `data_source_cmdevice_nodes_test.go`, `data_source_cmpart_softwareimages_test.go`
   - **Success Criteria**: Document test structure for all 7 test scenarios from spec
   - **Deliverable**: Test template with ConfigStateChecks pattern for each scenario

### Research Deliverable: `research.md`

**Structure**:
```markdown
# Research: BCM Kubernetes Clusters Data Source

## BCM API Analysis
- Service: cmkube
- Method: getKubeClusters
- Request/Response Examples
- Field Name Mappings (camelCase → snake_case)

## Field Mapping Table
| BCM API Field | Terraform Attribute | Type | Null Handling |
|---------------|---------------------|------|---------------|
| masterNodes   | master_nodes        | List | Required      |
| ...           | ...                 | ...  | ...           |

## Filter Implementation Strategy
- FilterModel definition
- Client-side filtering pseudocode
- AND logic for multiple filters

## List Extraction Pattern
- Helper function for []interface{} → types.List
- Element type: types.StringType
- Null/empty list handling

## Testing Strategy
- Test coverage matrix (7 tests)
- ConfigStateChecks pattern with statecheck/knownvalue
- Environment portability considerations
```

**Output**: All NEEDS CLARIFICATION items resolved with concrete implementation details.

---

## Phase 1: Design & Contracts

**Objective**: Generate data model, API contracts, and update agent context based on Phase 0 research findings.

### Design Artifacts

#### 1. Data Model Design (`data-model.md`)

**Entity**: CMKubeClustersDataSource

**Schema Structure**:
```go
type CMKubeClustersDataSourceModel struct {
    ID       types.String              `tfsdk:"id"`       // Placeholder: "cmkube-clusters"
    Filter   *ClusterFilterModel       `tfsdk:"filter"`   // Optional filter block
    Clusters []KubeClusterModel        `tfsdk:"clusters"` // List of clusters
}

type ClusterFilterModel struct {
    NamePattern   types.String `tfsdk:"name_pattern"`   // Optional: case-insensitive substring match
    Version       types.String `tfsdk:"version"`        // Optional: exact semver match
    MasterNodeID  types.String `tfsdk:"master_node_id"` // Optional: find clusters with master node UUID
}

type KubeClusterModel struct {
    // Identity
    ID   types.String `tfsdk:"id"`   // Computed: same as UUID
    UUID types.String `tfsdk:"uuid"` // Computed: BCM cluster UUID
    Name types.String `tfsdk:"name"` // Computed: cluster name

    // Node configuration
    MasterNodes types.List `tfsdk:"master_nodes"` // Computed: list of master node UUIDs (StringType elements)
    WorkerNodes types.List `tfsdk:"worker_nodes"` // Computed: list of worker node UUIDs (StringType elements)

    // Network configuration
    ManagementNetwork types.String `tfsdk:"management_network"` // Computed: management network UUID
    OverlayNetwork    types.String `tfsdk:"overlay_network"`    // Computed: pod network overlay config
    DNSServers        types.List   `tfsdk:"dns_servers"`        // Computed: list of DNS IPs (StringType elements)

    // Kubernetes configuration
    Version   types.String `tfsdk:"version"`    // Computed: semver string
    CNIPlugin types.String `tfsdk:"cni_plugin"` // Computed: CNI plugin name

    // Storage configuration
    StorageClasses types.String `tfsdk:"storage_classes"` // Computed: JSON-encoded string

    // Load balancer configuration
    LoadBalancerMode types.String `tfsdk:"load_balancer_mode"` // Computed: load balancer strategy

    // Cluster addons
    Addons            types.String `tfsdk:"addons"`             // Computed: JSON-encoded string
    IngressController types.String `tfsdk:"ingress_controller"` // Computed: JSON-encoded string

    // Computed metadata
    CreationTime types.Int64 `tfsdk:"creation_time"` // Computed: Unix timestamp
    RevisionID   types.Int64 `tfsdk:"revision_id"`   // Computed: revision number
}
```

**Field Validation Rules**:
- `name_pattern`: No schema validation (invalid patterns return empty list)
- `version`: No schema validation (exact match filtering in Go)
- `master_node_id`: No schema validation (UUID format not enforced)

**Relationships**:
- Data source ↔ Resource: Schema matches `resource_cmkube_cluster.go` for import compatibility
- Filter → Clusters: One-to-many (filter applies to all clusters, results in subset)

#### 2. API Contract Documentation (`contracts/cmkube-api.json`)

**Request Format**:
```json
{
  "service": "cmkube",
  "call": "getKubeClusters"
}
```

**Response Format**:
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
    "revisionID": 5
  }
]
```

**Field Mappings**:
| BCM API Field | Terraform Attribute | Type | Extraction Method |
|---------------|---------------------|------|-------------------|
| uuid | uuid, id | types.String | getStringValue(data, "uuid") |
| name | name | types.String | getStringValue(data, "name") |
| masterNodes | master_nodes | types.List | getListValue(data, "masterNodes") |
| workerNodes | worker_nodes | types.List | getListValue(data, "workerNodes") |
| managementNetwork | management_network | types.String | getStringValue(data, "managementNetwork") |
| overlayNetwork | overlay_network | types.String | getStringValue(data, "overlayNetwork") |
| dnsServers | dns_servers | types.List | getListValue(data, "dnsServers") |
| version | version | types.String | getStringValue(data, "version") |
| cniPlugin | cni_plugin | types.String | getStringValue(data, "cniPlugin") |
| storageClasses | storage_classes | types.String | getStringValue(data, "storageClasses") |
| loadBalancerMode | load_balancer_mode | types.String | getStringValue(data, "loadBalancerMode") |
| addons | addons | types.String | getStringValue(data, "addons") |
| ingressController | ingress_controller | types.String | getStringValue(data, "ingressController") |
| creationTime | creation_time | types.Int64 | getInt64Value(data, "creationTime") |
| revisionID | revision_id | types.Int64 | getInt64Value(data, "revisionID") |

#### 3. Quick Start Guide (`quickstart.md`)

**Target Audience**: Developers implementing the data source following TDD workflow

**Contents**:
- Prerequisites (Go 1.24, TF_ACC=1 environment setup)
- RED Phase: Write failing acceptance test template
- GREEN Phase: Minimal implementation skeleton
- REFACTOR Phase: Full CRUD implementation with API integration
- Run tests: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource`
- Generate docs: `make generate`

#### 4. Agent Context Update

**Command**: `.specify/scripts/bash/update-agent-context.sh copilot`

**Updates**:
- Add `cmkube.getKubeClusters` API method to BCM API section
- Add cluster field mappings to test_helpers.go mapping table
- Preserve manual additions between markers

### Design Deliverables

- `data-model.md` - Complete schema design with Go struct definitions
- `contracts/cmkube-api.json` - BCM API request/response contract
- `quickstart.md` - Developer quick start guide for TDD workflow
- Agent context files updated with new technology

---

## Phase 2: Implementation Planning (Task Generation)

**Objective**: Generate actionable, dependency-ordered tasks for `/speckit.implement` execution.

**Note**: This phase is executed by the `/speckit.tasks` command, NOT by `/speckit.plan`. The tasks will be generated after Phase 1 design artifacts are complete.

**Expected Task Breakdown** (Preview):

1. **RED Phase Tasks**:
   - Write `TestAccCMKubeClustersDataSource_Basic` (list all without filters)
   - Write `TestAccCMKubeClustersDataSource_FilterByName` (name_pattern filter)
   - Write `TestAccCMKubeClustersDataSource_FilterByVersion` (version filter)
   - Write `TestAccCMKubeClustersDataSource_FilterByMasterNode` (master_node_id filter)
   - Write `TestAccCMKubeClustersDataSource_MultipleFilters` (AND logic)
   - Write `TestAccCMKubeClustersDataSource_EmptyResults` (no matches)
   - Write `TestAccCMKubeClustersDataSource_NullFields` (missing optional fields)
   - Run tests: verify all 7 tests fail

2. **GREEN Phase Tasks**:
   - Implement `data_source_cmkube_clusters.go` minimal skeleton
   - Implement `Schema()` method with all attributes
   - Implement `Read()` method with hardcoded response
   - Implement `Configure()` method
   - Register data source in `provider.go`
   - Run tests: verify all 7 tests pass with minimal implementation

3. **REFACTOR Phase Tasks**:
   - Add BCM API call to `Read()` method (CallJSONRPC)
   - Implement client-side filtering (name_pattern, version, master_node_id)
   - Implement null-safe field extraction (getStringValue, getInt64Value, list extraction)
   - Add comprehensive error handling
   - Run tests: verify all 7 tests pass with real API

4. **DOCUMENTATION Phase Tasks**:
   - Create `examples/data-sources/bcm_cmkube_clusters/data-source.tf`
   - Add example: list all clusters
   - Add example: filter by name pattern
   - Add example: filter by version
   - Add example: combine multiple filters
   - Add example: use for terraform import
   - Run `make generate` to generate documentation

**Dependencies**:
- Task 2 (GREEN) depends on Task 1 (RED) - tests must fail first
- Task 3 (REFACTOR) depends on Task 2 (GREEN) - minimal implementation must pass tests
- Task 4 (DOCUMENTATION) depends on Task 3 (REFACTOR) - implementation must be complete

---

## Implementation Notes

### Key Design Decisions

1. **Client-Side Filtering Strategy**
   - **Decision**: Retrieve all clusters via `cmkube.getKubeClusters`, then filter in Go
   - **Rationale**: BCM API does not support server-side filtering parameters
   - **Alternative Rejected**: Individual cluster queries (inefficient, no batch API)
   - **Pattern Source**: `data_source_cmpart_softwareimages.go:Read()` lines 120-180

2. **Filter AND Logic**
   - **Decision**: When multiple filters specified, cluster must match ALL filters
   - **Rationale**: Most restrictive filtering matches user expectations
   - **Alternative Rejected**: OR logic (too permissive, harder to reason about)
   - **Pattern Source**: `data_source_cmpart_softwareimages.go` FilterModel with category + name_pattern

3. **Name Pattern Matching**
   - **Decision**: Case-insensitive substring match using `strings.Contains()`
   - **Rationale**: Simple, predictable, matches existing data source pattern
   - **Alternative Rejected**: Regex (complex, error-prone) or glob library (new dependency)
   - **Pattern Source**: `data_source_cmdevice_nodes.go:Read()` hostname_pattern filter

4. **List Field Extraction**
   - **Decision**: Custom helper function `getListValue()` for BCM array → types.List
   - **Rationale**: No existing helper for list extraction, need consistent null handling
   - **Alternative Rejected**: Inline extraction (duplicated logic across fields)
   - **Implementation**: Extract []interface{} → []string → types.ListValueMust()

5. **Master Node Filter Logic**
   - **Decision**: Match ANY master node in cluster's master_nodes list
   - **Rationale**: Operational use case is "find which cluster(s) contain this node"
   - **Alternative Rejected**: Match ALL master nodes (overly restrictive, unusual use case)
   - **Implementation**: Iterate master_nodes, return cluster if UUID found

### Risk Mitigation

| Risk | Impact | Mitigation Strategy |
|------|--------|---------------------|
| BCM API returns unexpected field types | Test failures, runtime errors | Use null-safe extraction, comprehensive type checking in helpers |
| Large cluster count (100+) causes timeout | User-facing error | Performance testing in acceptance tests, consider caching strategy |
| Filter logic complexity (3 filters × AND) | Bugs in filtering | Unit tests for filter logic, comprehensive acceptance tests |
| Schema mismatch with resource | Import failures | Cross-reference `resource_cmkube_cluster.go` schema, verify field names/types |
| Missing optional fields in API response | Null pointer errors | Use getStringValue/getInt64Value for ALL fields, handle null lists |

### Testing Strategy

**Acceptance Test Coverage** (7 tests):

1. **TestAccCMKubeClustersDataSource_Basic**: List all clusters without filters
   - **Assertions**: ID not null, clusters[0].uuid not null, clusters[0].name not null, clusters[0].master_nodes has elements
   - **Pattern**: `data_source_cmdevice_nodes_test.go:TestAccCMDeviceNodesDataSource_Basic`

2. **TestAccCMKubeClustersDataSource_FilterByName**: Filter by name_pattern
   - **Config**: `filter { name_pattern = "test-cluster" }`
   - **Assertions**: All returned clusters contain pattern in name (case-insensitive)
   - **Pattern**: `data_source_cmpart_softwareimages_test.go:TestAccCMPartSoftwareImagesDataSource_FilterByName`

3. **TestAccCMKubeClustersDataSource_FilterByVersion**: Filter by Kubernetes version
   - **Config**: `filter { version = "1.28.0" }`
   - **Assertions**: All returned clusters have version = "1.28.0"

4. **TestAccCMKubeClustersDataSource_FilterByMasterNode**: Filter by master node UUID
   - **Config**: `filter { master_node_id = "node-uuid-123" }`
   - **Assertions**: All returned clusters contain specified UUID in master_nodes list

5. **TestAccCMKubeClustersDataSource_MultipleFilters**: Combine name + version filters (AND logic)
   - **Config**: `filter { name_pattern = "prod-*"; version = "1.28.0" }`
   - **Assertions**: Returned clusters match both filters

6. **TestAccCMKubeClustersDataSource_EmptyResults**: No clusters match filter
   - **Config**: `filter { name_pattern = "nonexistent-cluster-xyz" }`
   - **Assertions**: clusters list is empty (not error), ID still set

7. **TestAccCMKubeClustersDataSource_NullFields**: Cluster with missing optional fields
   - **Setup**: Assumes BCM has cluster with null worker_nodes, null dns_servers
   - **Assertions**: Optional fields are null, required fields present, no errors

**Modern Test Patterns** (terraform-plugin-testing v1.13.3+):
- Use `statecheck.ExpectKnownValue()` for type-safe state assertions
- Use `knownvalue.StringExact()`, `knownvalue.NotNull()`, `knownvalue.ListSizeExact()` matchers
- Use `tfjsonpath.New("clusters").AtSliceIndex(0).AtMapKey("name")` for nested attributes
- Environment-portable: no hardcoded counts, generate unique test names

### Dependencies

**Existing Code Reuse**:
- `bcm_client.go:BCMClient.CallJSONRPC()` - BCM API calls
- `test_helpers.go:createTestBCMClient()` - Test BCM client creation
- `test_helpers.go:generateUniqueTestName()` - Unique test resource names
- `data_source_cmpart_softwareimages.go:getStringValue()` - Null-safe string extraction (lines 450-460)
- `data_source_cmpart_softwareimages.go:getInt64Value()` - Null-safe int64 extraction (lines 470-490)

**New Code Required**:
- `getListValue()` helper function for list extraction (similar to getStringValue but for []string)
- `ClusterFilterModel` struct (similar to `SoftwareImageFilterModel`)
- `KubeClusterModel` struct (similar to `SoftwareImageModel` but for clusters)

### Success Criteria

**Definition of Done**:
1. ✅ All 7 acceptance tests pass with `TF_ACC=1`
2. ✅ Data source registered in `provider.go` DataSources() method
3. ✅ Examples created in `examples/data-sources/bcm_cmkube_clusters/data-source.tf`
4. ✅ Documentation auto-generated via `make generate` without errors
5. ✅ Schema matches `resource_cmkube_cluster.go` attribute names and types
6. ✅ All 12 functional requirements from spec.md verified by tests
7. ✅ Code follows HashiCorp style guide (passes `make lint`)
8. ✅ Zero new complexity violations (reuses existing patterns)

---

## Next Steps

1. **Execute Phase 0**: Run research tasks to resolve all NEEDS CLARIFICATION items
2. **Generate research.md**: Document BCM API analysis, field mappings, filter patterns, test strategy
3. **Execute Phase 1**: Generate data-model.md, contracts/, quickstart.md
4. **Update Agent Context**: Run `.specify/scripts/bash/update-agent-context.sh copilot`
5. **Re-evaluate Constitution Check**: Verify no new complexity introduced
6. **STOP**: Report completion of `/speckit.plan` command

**Command to proceed**: `/speckit.tasks` (generates tasks.md for implementation)

---

## Appendix: Reference Patterns

### Data Source Read Pattern (from `data_source_cmpart_softwareimages.go`)

```go
func (d *CMPartSoftwareImagesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data CMPartSoftwareImagesDataSourceModel

    // 1. Get filter configuration from plan
    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // 2. Call BCM API to retrieve all entities
    body, err := d.client.CallJSONRPC(ctx, "cmpart", "getSoftwareImages")
    if err != nil {
        resp.Diagnostics.AddError("Error Reading Software Images", "Could not read software images: "+err.Error())
        return
    }

    // 3. Parse API response
    var apiResponse []map[string]interface{}
    if err := json.Unmarshal(body, &apiResponse); err != nil {
        resp.Diagnostics.AddError("Error Parsing Software Images", "Could not parse response: "+err.Error())
        return
    }

    // 4. Apply client-side filters
    var filteredImages []SoftwareImageModel
    for _, imageData := range apiResponse {
        include := true

        // Apply filter 1: name_pattern
        if data.Filter != nil && !data.Filter.NamePattern.IsNull() {
            pattern := strings.ToLower(data.Filter.NamePattern.ValueString())
            imageName := strings.ToLower(getStringValue(imageData, "name").ValueString())
            if !strings.Contains(imageName, pattern) {
                include = false
            }
        }

        // Apply filter 2: category
        if data.Filter != nil && !data.Filter.Category.IsNull() {
            if getStringValue(imageData, "childType").ValueString() != data.Filter.Category.ValueString() {
                include = false
            }
        }

        if include {
            model := mapImageDataToModel(imageData)
            filteredImages = append(filteredImages, model)
        }
    }

    // 5. Set state
    data.Images = filteredImages
    data.ID = types.StringValue("cmpart-softwareimages")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Null-Safe Helper Pattern (from `data_source_cmpart_softwareimages.go`)

```go
// getStringValue extracts a string field with null handling
func getStringValue(data map[string]interface{}, key string) types.String {
    if val, ok := data[key]; ok && val != nil {
        if strVal, ok := val.(string); ok && strVal != "" {
            return types.StringValue(strVal)
        }
    }
    return types.StringNull()
}

// getInt64Value extracts an int64 field with null handling (supports float64, int, int64)
func getInt64Value(data map[string]interface{}, key string) types.Int64 {
    if val, ok := data[key]; ok && val != nil {
        switch v := val.(type) {
        case float64:
            return types.Int64Value(int64(v))
        case int64:
            return types.Int64Value(v)
        case int:
            return types.Int64Value(int64(v))
        }
    }
    return types.Int64Null()
}
```

### Filter Model Pattern (from `data_source_cmpart_softwareimages.go`)

```go
type SoftwareImageFilterModel struct {
    NamePattern types.String `tfsdk:"name_pattern"` // Case-insensitive substring match
    Category    types.String `tfsdk:"category"`     // Exact match for child_type
}

// Schema definition
"filter": schema.SingleNestedBlock{
    MarkdownDescription: "Filter results using client-side filtering. Multiple filters use AND logic (all filters must match).",
    Attributes: map[string]schema.Attribute{
        "name_pattern": schema.StringAttribute{
            Optional:            true,
            MarkdownDescription: "Case-insensitive substring match for image name (e.g., 'ubuntu' matches 'Ubuntu-20.04').",
        },
        "category": schema.StringAttribute{
            Optional:            true,
            MarkdownDescription: "Exact match for image category (childType field, e.g., 'regular', 'golden').",
        },
    },
}
```

### Modern Test Pattern (from `data_source_cmdevice_nodes_test.go`)

```go
func TestAccCMDeviceNodesDataSource_Basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceNodesDataSourceConfig_basic(),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmdevice_nodes.test",
                        tfjsonpath.New("id"),
                        knownvalue.NotNull(),
                    ),
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmdevice_nodes.test",
                        tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("uuid"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}

func testAccCMDeviceNodesDataSourceConfig_basic() string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_nodes" "test" {}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
    )
}
```
