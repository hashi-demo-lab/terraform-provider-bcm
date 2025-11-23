# Feature Specification: BCM Kubernetes Clusters Data Source

**Feature Branch**: `001-cmkube-clusters-datasource`
**Created**: 2025-11-23
**Status**: Draft
**Input**: User description: "Create a feature specification for implementing the data.bcm_cmkube_clusters data source to list and filter Kubernetes clusters managed by BCM."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Cluster Discovery for Import (Priority: P1)

Platform engineers need to discover existing Kubernetes clusters in BCM to import them into Terraform state for infrastructure management. This enables transitioning from manually-managed clusters to infrastructure-as-code without recreating clusters.

**Why this priority**: This is the foundational use case that enables Terraform adoption. Without the ability to discover and import existing clusters, users cannot begin managing BCM Kubernetes infrastructure with Terraform. This is a prerequisite for all other cluster management workflows.

**Independent Test**: Can be fully tested by querying the data source without filters and verifying all existing clusters are returned with their UUIDs and names, then using those UUIDs for terraform import commands.

**Acceptance Scenarios**:

1. **Given** BCM contains 3 Kubernetes clusters, **When** engineer queries `data.bcm_cmkube_clusters.all` without filters, **Then** data source returns all 3 clusters with uuid, name, master_nodes, worker_nodes, and version attributes populated
2. **Given** engineer has cluster UUID from data source, **When** they run `terraform import bcm_cmkube_cluster.prod <uuid>`, **Then** cluster is successfully imported into Terraform state
3. **Given** BCM contains no clusters, **When** engineer queries data source, **Then** data source returns empty clusters list without error

---

### User Story 2 - Filter Clusters by Name Pattern (Priority: P2)

DevOps engineers managing multiple environments need to filter clusters by naming convention (e.g., "prod-*", "dev-*", "staging-*") to organize cluster references and build environment-specific dependencies.

**Why this priority**: This enables organized multi-environment management and is commonly needed for production Terraform configurations. While not absolutely required for basic import, it significantly improves usability for teams with multiple clusters.

**Independent Test**: Can be tested independently by creating test clusters with specific name patterns, then verifying the filter returns only matching clusters while excluding others.

**Acceptance Scenarios**:

1. **Given** BCM has clusters "prod-us-east", "prod-eu-west", "dev-us-east", **When** engineer applies filter `name_pattern = "prod-*"`, **Then** data source returns only the 2 production clusters
2. **Given** BCM has cluster "k8s-monitoring-001", **When** engineer applies filter `name_pattern = "*monitoring*"`, **Then** data source returns the monitoring cluster
3. **Given** BCM has 10 clusters but none match "nonexistent-*", **When** engineer applies this filter, **Then** data source returns empty list without error

---

### User Story 3 - Filter Clusters by Version (Priority: P3)

Platform teams performing Kubernetes version upgrades need to identify clusters running specific Kubernetes versions to plan upgrade sequences and validate version consistency across environments.

**Why this priority**: This is valuable for operational workflows but not critical for basic Terraform adoption. Most teams can discover versions through the unfiltered data source, making this a convenience feature rather than a requirement.

**Independent Test**: Can be tested by creating clusters with different Kubernetes versions and verifying the filter returns only clusters matching the specified version.

**Acceptance Scenarios**:

1. **Given** BCM has clusters running v1.27.3, v1.28.1, v1.28.1, **When** engineer applies filter `version = "v1.28.1"`, **Then** data source returns the 2 clusters running v1.28.1
2. **Given** all clusters have been upgraded to v1.29.0, **When** engineer checks for legacy versions with `version = "v1.27.3"`, **Then** data source returns empty list confirming upgrade completion
3. **Given** version filter is not specified, **When** engineer queries clusters, **Then** all clusters are returned regardless of version

---

### User Story 4 - Filter Clusters by Master Node (Priority: P3)

Infrastructure operators troubleshooting node assignments or planning node replacements need to find which cluster(s) a specific node belongs to based on master node UUID.

**Why this priority**: This is a specialized troubleshooting scenario. While helpful for operations, it's not frequently needed and can be accomplished through other means (examining cluster details manually). It's a quality-of-life improvement rather than a core requirement.

**Independent Test**: Can be tested by creating a cluster with a known master node UUID, then verifying the filter returns that cluster when queried by the master node ID.

**Acceptance Scenarios**:

1. **Given** cluster "prod-main" has master node with UUID "node-abc-123", **When** engineer applies filter `master_node_id = "node-abc-123"`, **Then** data source returns only the "prod-main" cluster
2. **Given** a node UUID exists but is not a master node in any cluster, **When** engineer filters by this UUID, **Then** data source returns empty list
3. **Given** multiple clusters share a master node in HA configuration, **When** engineer filters by shared master node UUID, **Then** data source returns all clusters using that master node

---

### Edge Cases

- What happens when BCM API is unreachable or returns authentication errors? (System should return clear Terraform error message with API failure details)
- How does the data source handle clusters with missing or null attributes (e.g., no worker nodes, no version specified)? (Use null-safe field extraction to return types.String/types.List null values for missing fields)
- What happens when name_pattern filter contains invalid glob syntax? (Terraform validates pattern syntax and returns validation error before API call)
- How are multiple filters combined (name_pattern + version)? (Use AND logic - clusters must match all specified filters)
- What happens when BCM returns clusters with unexpected field types or structures? (Log warning and skip invalid clusters, continue processing valid ones)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Data source MUST retrieve all Kubernetes clusters from BCM via cmkube.getKubeClusters API
- **FR-002**: Data source MUST support optional name_pattern filter using glob pattern matching (case-sensitive, supports * wildcard)
- **FR-003**: Data source MUST support optional version filter for exact Kubernetes version match (semver string)
- **FR-004**: Data source MUST support optional master_node_id filter to find clusters containing specific master node UUID
- **FR-005**: Data source MUST apply multiple filters with AND logic (cluster must match all specified filters)
- **FR-006**: Data source MUST populate all cluster attributes including uuid, name, master_nodes, worker_nodes, etcd_nodes, management_network, version, cni_plugin, creation_time
- **FR-007**: Data source MUST use null-safe field extraction for optional API fields (overlayNetwork, dnsServers, storageClasses, etc.)
- **FR-008**: Data source MUST set ID to placeholder value (e.g., "cmkube-clusters") since this is a list data source
- **FR-009**: Data source MUST handle BCM API errors gracefully with clear error messages for authentication failures, network errors, and invalid API responses
- **FR-010**: Data source MUST return empty clusters list (not error) when no clusters match filters or BCM has no clusters
- **FR-011**: Data source MUST map BCM API camelCase field names to Terraform snake_case attribute names (masterNodes → master_nodes, workerNodes → worker_nodes, managementNetwork → management_network)
- **FR-012**: Data source MUST perform client-side filtering after retrieving all clusters from BCM (not API-side filtering)

### Key Entities

- **Cluster**: Represents a Kubernetes cluster managed by BCM with attributes:
  - Identity: uuid (unique identifier), name (user-friendly name)
  - Node Configuration: master_nodes (list of master node UUIDs), worker_nodes (list of worker node UUIDs), etcd_nodes (list of etcd node UUIDs)
  - Network Configuration: management_network (UUID), overlay_network (pod network config), dns_servers (list of IPs)
  - Kubernetes Configuration: version (semver string), cni_plugin (CNI plugin name)
  - Storage: storage_classes (JSON config)
  - Load Balancer: load_balancer_mode (strategy)
  - Addons: addons (JSON config), ingress_controller (JSON config)
  - Metadata: creation_time (Unix timestamp), revision_id (version number), base_type, child_type

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Engineers can discover all BCM Kubernetes clusters by querying the data source without filters, retrieving uuid, name, and node configuration for each cluster in under 5 seconds
- **SC-002**: Engineers can filter clusters using name_pattern glob syntax with 100% accuracy (only matching clusters returned)
- **SC-003**: Data source handles environments with 0-100+ clusters without performance degradation or timeout errors
- **SC-004**: Data source successfully processes clusters with missing optional attributes (null worker_nodes, null version) without errors, returning appropriate null values
- **SC-005**: Engineers can use discovered cluster UUIDs for terraform import commands with 100% success rate
- **SC-006**: All 12 functional requirements are verified by passing acceptance tests covering list all, filter by name, filter by version, filter by master node, and edge cases
- **SC-007**: Data source documentation auto-generates via `make generate` with complete attribute descriptions and filter usage examples

---

## Clarifications

### Session 2025-11-23

- Q: What API method should be used to retrieve clusters - getKubeClusters (plural) or getKubeCluster (singular)? → A: Use `cmkube.getKubeClusters` (plural) to retrieve all clusters in a single API call, following the data source list+filter pattern established in `data_source_cmdevice_nodes.go` and `data_source_cmpart_softwareimages.go`.

- Q: Should filters be applied client-side (in Terraform provider code) or server-side (via BCM API parameters)? → A: Client-side filtering after retrieving all clusters from the API. This matches the established pattern in existing data sources (`CMDeviceNodesDataSource`, `CMPartSoftwareImagesDataSource`) where all entities are retrieved then filtered in Go code.

- Q: What field naming convention does the BCM API use for cluster attributes - camelCase, snake_case, or mixed? → A: BCM API uses camelCase (e.g., `masterNodes`, `workerNodes`, `managementNetwork`, `overlayNetwork`). The Terraform provider must map these to snake_case for HCL schema (e.g., `master_nodes`, `worker_nodes`, `management_network`) using `getStringValue()` and `getListValue()` helper functions.

- Q: How should the data source handle the `etcd_nodes` field - is it required, optional, or always present in the API response? → A: Optional field. Based on the resource implementation in `resource_cmkube_cluster.go`, `etcd_nodes` is not present in the schema or API calls. The current cluster model only includes `master_nodes` and `worker_nodes`. This field should be omitted from the data source schema for consistency.

- Q: What schema validation should be applied to the `name_pattern` filter - regex validation, glob pattern validation, or no validation? → A: No schema-level validation required. The `name_pattern` filter uses Go's `strings.Contains()` for substring matching (case-insensitive pattern matching as implemented in existing data sources). Invalid patterns will simply match nothing and return an empty list.

- Q: Should the `master_node_id` filter match ANY master node in the cluster's list or require ALL master nodes to match? → A: Match ANY master node. When filtering by `master_node_id`, return clusters where the specified UUID appears in the `master_nodes` list. This enables the operational use case of finding which cluster(s) contain a specific master node.

- Q: How should multiple filters be combined - AND logic (all must match), OR logic (any must match), or configurable? → A: AND logic (all filters must match). This follows the established pattern in `CMPartSoftwareImagesDataSource` where multiple filter fields use AND logic. A cluster must satisfy all specified filters to be included in results.

- Q: What should the placeholder `id` value be for this list data source - "cmkube-clusters", timestamp-based, or hash of cluster UUIDs? → A: Use constant string `"cmkube-clusters"`. This follows the established pattern in existing data sources (`data_source_cmdevice_nodes.go` sets `"cmdevice-nodes"`, `data_source_cmpart_softwareimages.go` sets `"cmpart-softwareimages"`).

- Q: Should the data source expose complex nested fields like `storage_classes`, `addons`, and `ingress_controller` as JSON strings or structured objects? → A: JSON strings (types.String). Based on `resource_cmkube_cluster.go` (lines 60, 64, 67), these fields are stored as JSON-encoded strings. The data source should match the resource schema for consistency and compatibility with import workflows.

- Q: What timeout should be applied to the BCM API call for retrieving clusters? → A: No explicit timeout in data source code. Use the client's default timeout (configured via `http.Client` in `bcm_client.go`). The BCM client handles timeouts at the transport layer, consistent with all other data sources in the provider.

---

## Implementation Specifications

### BCM API Contract

**Service**: `cmkube`
**Method**: `getKubeClusters` (list all clusters)
**Request Pattern**:
```json
{
  "service": "cmkube",
  "call": "getKubeClusters"
}
```

**Response Structure** (array of cluster objects):
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

**Field Mappings** (BCM API → Terraform):
| BCM API Field | Terraform Attribute | Type | Notes |
|---------------|---------------------|------|-------|
| uuid | uuid, id | types.String | Both set to same value |
| name | name | types.String | Required |
| masterNodes | master_nodes | types.List (of String) | Required list |
| workerNodes | worker_nodes | types.List (of String) | Optional list |
| managementNetwork | management_network | types.String | Optional UUID |
| overlayNetwork | overlay_network | types.String | Optional config |
| dnsServers | dns_servers | types.List (of String) | Optional list |
| version | version | types.String | Optional semver |
| cniPlugin | cni_plugin | types.String | Optional |
| storageClasses | storage_classes | types.String | Optional JSON string |
| loadBalancerMode | load_balancer_mode | types.String | Optional |
| addons | addons | types.String | Optional JSON string |
| ingressController | ingress_controller | types.String | Optional JSON string |
| creationTime | creation_time | types.Int64 | Computed timestamp |
| revisionID | revision_id | types.Int64 | Computed version |

### Filter Implementation Details

**Filter Block Schema**:
```hcl
filter {
  name_pattern    = "prod-*"           # Optional: substring match (case-insensitive)
  version         = "1.28.0"           # Optional: exact version match
  master_node_id  = "node-uuid-abc"    # Optional: find clusters containing this master node
}
```

**Filter Logic** (client-side, applied sequentially):
```go
// Pseudocode for filter application
for each cluster in apiResponse {
  include := true

  // Filter: name_pattern (case-insensitive substring match)
  if filter.NamePattern != nil && !filter.NamePattern.IsNull() {
    pattern := strings.ToLower(filter.NamePattern.ValueString())
    clusterName := strings.ToLower(cluster.Name)
    if !strings.Contains(clusterName, pattern) {
      include = false
    }
  }

  // Filter: version (exact match)
  if filter.Version != nil && !filter.Version.IsNull() {
    if cluster.Version != filter.Version.ValueString() {
      include = false
    }
  }

  // Filter: master_node_id (ANY match in master_nodes list)
  if filter.MasterNodeID != nil && !filter.MasterNodeID.IsNull() {
    found := false
    for _, nodeUUID := range cluster.MasterNodes {
      if nodeUUID == filter.MasterNodeID.ValueString() {
        found = true
        break
      }
    }
    if !found {
      include = false
    }
  }

  if include {
    append cluster to results
  }
}
```

### Null-Safe Field Extraction

Use existing helper functions from `data_source_cmpart_softwareimages.go:399-431`:
- `getStringValue(data, key)` - Handles missing/null string fields
- `getInt64Value(data, key)` - Handles missing/null int64 fields (supports float64/int/int64)
- Custom list extraction for `master_nodes`, `worker_nodes`, `dns_servers`

### Error Handling Strategy

| Error Condition | Handling Strategy | HTTP Status | Response to User |
|----------------|-------------------|-------------|------------------|
| BCM API unreachable | Return error diagnostic | N/A (network error) | "Error Reading Clusters: Could not connect to BCM API: {error}" |
| Authentication failure | Return error diagnostic | 401/403 | "Error Reading Clusters: Authentication failed: {error}" |
| Invalid JSON response | Return error diagnostic | 200 | "Error Parsing Clusters: Could not parse response: {error}" |
| Empty cluster list | Success with empty clusters list | 200 | clusters = [] (no error) |
| Missing optional field | Use null-safe extraction | 200 | Set field to types.StringNull() or types.ListNull() |
| Invalid cluster data | Log warning, skip cluster | 200 | Log "Skipping cluster due to invalid data", continue processing |

### Testing Requirements

**Test Coverage Matrix**:

| Test Name | Scenario | Assertions | Filters Used |
|-----------|----------|------------|--------------|
| TestAccCMKubeClustersDataSource_Basic | List all clusters without filters | - ID is not null<br>- clusters[0].uuid is not null<br>- clusters[0].name is not null<br>- clusters[0].master_nodes has elements | None |
| TestAccCMKubeClustersDataSource_FilterByName | Filter by name pattern | - Only matching clusters returned<br>- All returned clusters contain pattern in name | name_pattern |
| TestAccCMKubeClustersDataSource_FilterByVersion | Filter by Kubernetes version | - Only clusters with matching version returned | version |
| TestAccCMKubeClustersDataSource_FilterByMasterNode | Filter by master node UUID | - Only clusters containing specified master node returned | master_node_id |
| TestAccCMKubeClustersDataSource_MultipleFilters | Combine multiple filters (AND logic) | - Only clusters matching ALL filters returned | name_pattern + version |
| TestAccCMKubeClustersDataSource_EmptyResults | Filter returns no matches | - clusters list is empty (not error)<br>- ID still set | name_pattern with no matches |
| TestAccCMKubeClustersDataSource_NullFields | Cluster with missing optional fields | - Optional fields are null<br>- Required fields present<br>- No errors | None |

**Test Naming Convention**: Follow existing pattern `TestAccCMKubeClustersDataSource_<Scenario>`

**Test Config Pattern**: Include provider block with environment variables
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

### Documentation Requirements

**Auto-Generated via `make generate`**:
- Data source description explaining cluster discovery use case
- Complete attribute reference table with types and descriptions
- Filter block documentation with examples
- Example HCL configurations for:
  - List all clusters
  - Filter by name pattern
  - Filter by version
  - Filter by master node
  - Combine multiple filters
  - Use cluster UUIDs for import

**Example Location**: `examples/data-sources/bcm_cmkube_clusters/data-source.tf`

---

## Dependencies & Constraints

### Technical Dependencies
- BCM API service: `cmkube` (Kubernetes cluster management)
- BCM API method: `getKubeClusters` (must return list of all clusters)
- Existing helper functions: `getStringValue()`, `getInt64Value()` (from `data_source_cmpart_softwareimages.go`)
- Terraform Plugin Framework: v1.16.1+
- Go standard library: `strings` (for pattern matching), `encoding/json` (for parsing)

### Compatibility Constraints
- Data source schema must match `resource_cmkube_cluster.go` attribute names and types for import compatibility
- Field naming must follow snake_case convention (Terraform standard)
- BCM API field mapping must handle camelCase → snake_case conversion
- Must work with BCM clusters that have 0 clusters (empty list scenario)

### Performance Constraints
- Client-side filtering acceptable for up to 100+ clusters (based on SC-003)
- Single API call to retrieve all clusters (no pagination required)
- Filtering operations performed in-memory (no API-side filtering support)

### Known Limitations
- No server-side filtering: All clusters retrieved then filtered in Go code (performance consideration for large deployments)
- Pattern matching is case-insensitive substring match (not full glob/regex support)
- `etcd_nodes` field intentionally omitted (not present in current resource schema)
- JSON fields (`storage_classes`, `addons`, `ingress_controller`) exposed as strings, not structured objects

---

## Out of Scope

- **Server-side filtering via BCM API**: All filtering is client-side
- **Regex pattern matching**: Only substring matching supported for `name_pattern`
- **Pagination**: Single API call retrieves all clusters
- **Cluster health/status fields**: Only metadata attributes exposed
- **Nested cluster resources**: No sub-resources like nodes, pods, or services
- **Cluster CRUD operations**: This is read-only data source (use `bcm_cmkube_cluster` resource for management)
- **Custom cluster queries**: Only predefined filters (name, version, master_node_id) supported
- **Real-time cluster state**: Data reflects BCM's last-known state, not live Kubernetes API state
