# Data Model: BCM Kubernetes Clusters Data Source

**Entity**: CMKubeClustersDataSource
**Purpose**: Read-only data source for discovering and filtering Kubernetes clusters managed by BCM
**API Service**: `cmkube.getKubeClusters`

---

## Go Struct Definitions

### Main Data Source Model

```go
// CMKubeClustersDataSourceModel describes the data source data model.
type CMKubeClustersDataSourceModel struct {
    ID       types.String         `tfsdk:"id"`       // Computed: placeholder "cmkube-clusters"
    Filter   *ClusterFilterModel  `tfsdk:"filter"`   // Optional: filter criteria
    Clusters []KubeClusterModel   `tfsdk:"clusters"` // Computed: list of clusters
}
```

**Attributes**:
- `ID`: Placeholder identifier for the data source (always "cmkube-clusters")
- `Filter`: Optional nested block for client-side filtering
- `Clusters`: Computed list of Kubernetes clusters matching filter criteria

---

### Filter Model

```go
// ClusterFilterModel describes the filter block for client-side filtering.
// Multiple filters use AND logic (all filters must match for a cluster to be included).
type ClusterFilterModel struct {
    NamePattern   types.String `tfsdk:"name_pattern"`   // Optional: case-insensitive substring match for cluster name
    Version       types.String `tfsdk:"version"`        // Optional: exact match for Kubernetes version (semver)
    MasterNodeID  types.String `tfsdk:"master_node_id"` // Optional: find clusters containing this master node UUID
}
```

**Filter Behavior**:
- `name_pattern`: Case-insensitive substring matching (e.g., "prod" matches "Prod-Cluster-01")
- `version`: Exact semver match (e.g., "1.28.0" matches only "1.28.0", not "1.28.1")
- `master_node_id`: ANY match in master_nodes list (cluster includes this UUID in master nodes)

**AND Logic**: When multiple filters are specified, a cluster must satisfy **all** filters to be included.

---

### Cluster Model

```go
// KubeClusterModel represents a BCM Kubernetes cluster with all fields.
type KubeClusterModel struct {
    // Identity fields
    ID   types.String `tfsdk:"id"`   // Computed: same as UUID
    UUID types.String `tfsdk:"uuid"` // Computed: BCM-assigned cluster UUID
    Name types.String `tfsdk:"name"` // Computed: cluster name

    // Node configuration
    MasterNodes types.List `tfsdk:"master_nodes"` // Computed: list of master node UUIDs (StringType elements)
    WorkerNodes types.List `tfsdk:"worker_nodes"` // Computed: list of worker node UUIDs (StringType elements)

    // Network configuration
    ManagementNetwork types.String `tfsdk:"management_network"` // Computed: management network UUID
    OverlayNetwork    types.String `tfsdk:"overlay_network"`    // Computed: pod network overlay config (CIDR)
    DNSServers        types.List   `tfsdk:"dns_servers"`        // Computed: list of DNS server IPs (StringType elements)

    // Kubernetes configuration
    Version   types.String `tfsdk:"version"`    // Computed: Kubernetes version (semver string)
    CNIPlugin types.String `tfsdk:"cni_plugin"` // Computed: CNI plugin name (calico, flannel, etc.)

    // Storage configuration
    StorageClasses types.String `tfsdk:"storage_classes"` // Computed: JSON-encoded storage class definitions

    // Load balancer configuration
    LoadBalancerMode types.String `tfsdk:"load_balancer_mode"` // Computed: load balancer strategy (metallb, etc.)

    // Cluster addons
    Addons            types.String `tfsdk:"addons"`             // Computed: JSON-encoded addon configurations
    IngressController types.String `tfsdk:"ingress_controller"` // Computed: JSON-encoded ingress controller config

    // Computed metadata
    CreationTime types.Int64 `tfsdk:"creation_time"` // Computed: Unix timestamp of cluster creation
    RevisionID   types.Int64 `tfsdk:"revision_id"`   // Computed: BCM revision number
}
```

**Field Categories**:

1. **Identity** (Required):
   - `id`, `uuid`: Cluster identifier (both set to same value for consistency with resource)
   - `name`: Human-readable cluster name

2. **Node Configuration** (Required/Optional):
   - `master_nodes`: Required list (minimum 1 master node)
   - `worker_nodes`: Optional list (can be null/empty for minimal clusters)

3. **Network Configuration** (Optional):
   - `management_network`: Network UUID for cluster management traffic
   - `overlay_network`: Pod network CIDR (e.g., "10.244.0.0/16")
   - `dns_servers`: List of DNS server IPs for cluster DNS resolution

4. **Kubernetes Configuration** (Optional):
   - `version`: Kubernetes version (e.g., "1.28.0")
   - `cni_plugin`: Container Network Interface plugin (e.g., "calico")

5. **Storage Configuration** (Optional):
   - `storage_classes`: JSON-encoded storage class definitions

6. **Load Balancer Configuration** (Optional):
   - `load_balancer_mode`: Load balancer implementation strategy

7. **Addons** (Optional):
   - `addons`: JSON-encoded cluster addons configuration
   - `ingress_controller`: JSON-encoded ingress controller configuration

8. **Metadata** (Computed):
   - `creation_time`: Unix timestamp of cluster creation
   - `revision_id`: BCM internal revision number for change tracking

---

## Terraform Schema Definition

### Data Source Schema

```go
func (d *CMKubeClustersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Data source to discover and filter BCM Kubernetes clusters.\n\n" +
            "Use this data source to list all Kubernetes clusters managed by BCM or filter by name pattern, " +
            "Kubernetes version, or master node UUID. This is useful for discovering clusters for import, " +
            "building dependencies, or operational queries.",

        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Data source identifier (always 'cmkube-clusters').",
            },
            "clusters": schema.ListNestedAttribute{
                Computed:            true,
                MarkdownDescription: "List of Kubernetes clusters matching the filter criteria.",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "id": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Cluster identifier (same as uuid).",
                        },
                        "uuid": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "BCM-assigned cluster UUID.",
                        },
                        "name": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Cluster name.",
                        },
                        "master_nodes": schema.ListAttribute{
                            ElementType:         types.StringType,
                            Computed:            true,
                            MarkdownDescription: "List of master node UUIDs.",
                        },
                        "worker_nodes": schema.ListAttribute{
                            ElementType:         types.StringType,
                            Computed:            true,
                            MarkdownDescription: "List of worker node UUIDs.",
                        },
                        "management_network": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Management network UUID for cluster management traffic.",
                        },
                        "overlay_network": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Pod network overlay configuration (CIDR).",
                        },
                        "dns_servers": schema.ListAttribute{
                            ElementType:         types.StringType,
                            Computed:            true,
                            MarkdownDescription: "List of DNS server IP addresses.",
                        },
                        "version": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Kubernetes version (semver format).",
                        },
                        "cni_plugin": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Container Network Interface (CNI) plugin name.",
                        },
                        "storage_classes": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Storage class definitions (JSON-encoded).",
                        },
                        "load_balancer_mode": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Load balancer mode/strategy.",
                        },
                        "addons": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Cluster addons configuration (JSON-encoded).",
                        },
                        "ingress_controller": schema.StringAttribute{
                            Computed:            true,
                            MarkdownDescription: "Ingress controller configuration (JSON-encoded).",
                        },
                        "creation_time": schema.Int64Attribute{
                            Computed:            true,
                            MarkdownDescription: "Cluster creation timestamp (Unix epoch).",
                        },
                        "revision_id": schema.Int64Attribute{
                            Computed:            true,
                            MarkdownDescription: "BCM revision number for change tracking.",
                        },
                    },
                },
            },
        },

        Blocks: map[string]schema.Block{
            "filter": schema.SingleNestedBlock{
                MarkdownDescription: "Optional filter criteria to limit returned clusters. " +
                    "Multiple filters use AND logic (all filters must match).",
                Attributes: map[string]schema.Attribute{
                    "name_pattern": schema.StringAttribute{
                        Optional:            true,
                        MarkdownDescription: "Case-insensitive substring match for cluster name (e.g., 'prod' matches 'Prod-Cluster-01').",
                    },
                    "version": schema.StringAttribute{
                        Optional:            true,
                        MarkdownDescription: "Exact match for Kubernetes version (semver format, e.g., '1.28.0').",
                    },
                    "master_node_id": schema.StringAttribute{
                        Optional:            true,
                        MarkdownDescription: "Find clusters containing this master node UUID in their master_nodes list.",
                    },
                },
            },
        },
    }
}
```

---

## Field Validation Rules

### Schema-Level Validation

**No validators required** for filter attributes:
- `name_pattern`: No regex/glob validation (invalid patterns simply return no matches)
- `version`: No semver validation (exact string comparison in filter logic)
- `master_node_id`: No UUID format validation (allows any string)

**Rationale**: Filters are optional and permissive. Invalid filter values result in empty results, not errors.

### Runtime Validation

**In Read() method**:
- Verify BCM API response is valid JSON array
- Verify each cluster object has required fields (uuid, name)
- Handle missing optional fields gracefully (return null values)

---

## Data Mapping Logic

### API Response → Terraform State

```go
func mapClusterDataToModel(apiData map[string]interface{}) KubeClusterModel {
    model := KubeClusterModel{}

    // Identity fields
    uuid := getStringValue(apiData, "uuid")
    model.ID = uuid
    model.UUID = uuid
    model.Name = getStringValue(apiData, "name")

    // Node configuration
    model.MasterNodes = getListValue(apiData, "masterNodes")
    model.WorkerNodes = getListValue(apiData, "workerNodes")

    // Network configuration
    model.ManagementNetwork = getStringValue(apiData, "managementNetwork")
    model.OverlayNetwork = getStringValue(apiData, "overlayNetwork")
    model.DNSServers = getListValue(apiData, "dnsServers")

    // Kubernetes configuration
    model.Version = getStringValue(apiData, "version")
    model.CNIPlugin = getStringValue(apiData, "cniPlugin")

    // Storage configuration
    model.StorageClasses = getStringValue(apiData, "storageClasses")

    // Load balancer configuration
    model.LoadBalancerMode = getStringValue(apiData, "loadBalancerMode")

    // Cluster addons
    model.Addons = getStringValue(apiData, "addons")
    model.IngressController = getStringValue(apiData, "ingressController")

    // Computed metadata
    model.CreationTime = getInt64Value(apiData, "creationTime")
    model.RevisionID = getInt64Value(apiData, "revisionID")

    return model
}
```

---

## Relationships

### Data Source ↔ Resource Compatibility

The data source schema **exactly matches** `resource_cmkube_cluster.go` schema to ensure:

1. **Terraform Import Compatibility**: Cluster UUID from data source can be used for `terraform import`
2. **Reference Consistency**: Resource and data source expose identical attributes
3. **Type Safety**: Field types match exactly (types.String, types.List, types.Int64)

**Verification**:
```hcl
# Discover cluster UUID from data source
data "bcm_cmkube_clusters" "existing" {
  filter {
    name_pattern = "prod-cluster"
  }
}

# Use UUID for import or reference
resource "bcm_cmkube_cluster" "imported" {
  # terraform import bcm_cmkube_cluster.imported <uuid>
  # UUID from: data.bcm_cmkube_clusters.existing.clusters[0].uuid
}
```

### Filter → Clusters Relationship

**Cardinality**: One-to-Many (One filter configuration → Zero or more matching clusters)

**Filter Application**:
1. Retrieve all clusters from BCM API (no API-side filtering)
2. For each cluster, apply all filter conditions
3. Include cluster in results only if ALL filters match (AND logic)

---

## Null Handling Strategy

### Required Fields (Never Null)

- `uuid`: Always present (BCM cluster identifier)
- `name`: Always present (cluster name)
- `master_nodes`: Always present (minimum 1 master node required)

### Optional Fields (Can Be Null)

- `worker_nodes`: Can be null/empty (clusters can run without workers)
- `management_network`: Can be null (uses default network)
- `overlay_network`: Can be null (uses default overlay)
- `dns_servers`: Can be null/empty (uses default DNS)
- `version`: Can be null (version not specified)
- `cni_plugin`: Can be null (default CNI)
- `storage_classes`: Can be null (no custom storage classes)
- `load_balancer_mode`: Can be null (no load balancer)
- `addons`: Can be null (no addons installed)
- `ingress_controller`: Can be null (no ingress controller)

### Null-Safe Extraction

**Pattern**: Use helper functions for ALL field extraction (even required fields):

```go
// String fields: Returns types.StringNull() if missing/null
model.Name = getStringValue(apiData, "name")

// Int64 fields: Returns types.Int64Null() if missing/null
model.CreationTime = getInt64Value(apiData, "creationTime")

// List fields: Returns types.ListNull(types.StringType) if missing/null/empty
model.WorkerNodes = getListValue(apiData, "workerNodes")
```

**Rationale**: Defensive programming - handles BCM API inconsistencies gracefully.

---

## Performance Considerations

### API Call Strategy

**Single API Call**: `cmkube.getKubeClusters` retrieves all clusters in one request
- **Pros**: Fast (1 API call), simple, no pagination logic
- **Cons**: Inefficient for large deployments (100+ clusters)
- **Mitigation**: Acceptable for current BCM scale (typically < 50 clusters)

### Filtering Performance

**Client-Side Filtering**: O(n) where n = total cluster count
- **Typical Case**: 10-50 clusters → < 100ms filtering
- **Worst Case**: 100+ clusters → < 1s filtering
- **Optimization**: Early exit on failed filter conditions (AND logic short-circuits)

### Memory Usage

**Cluster Data**: ~2KB per cluster (JSON representation)
- **Typical Case**: 50 clusters × 2KB = 100KB memory
- **Worst Case**: 100 clusters × 2KB = 200KB memory
- **Impact**: Negligible memory footprint

---

## Schema Consistency Verification

**Cross-Reference**: `resource_cmkube_cluster.go` lines 40-75

| Resource Attribute | Data Source Attribute | Type Match | Notes |
|-------------------|----------------------|------------|-------|
| id | id | ✅ types.String | Both computed |
| uuid | uuid | ✅ types.String | Both computed |
| name | name | ✅ types.String | Resource: required, Data source: computed |
| master_nodes | master_nodes | ✅ types.List (StringType) | Both list of UUIDs |
| worker_nodes | worker_nodes | ✅ types.List (StringType) | Both optional |
| management_network | management_network | ✅ types.String | Both optional |
| overlay_network | overlay_network | ✅ types.String | Both optional |
| dns_servers | dns_servers | ✅ types.List (StringType) | Both optional |
| version | version | ✅ types.String | Both optional |
| cni_plugin | cni_plugin | ✅ types.String | Both optional |
| storage_classes | storage_classes | ✅ types.String | Both optional (JSON) |
| load_balancer_mode | load_balancer_mode | ✅ types.String | Both optional |
| addons | addons | ✅ types.String | Both optional (JSON) |
| ingress_controller | ingress_controller | ✅ types.String | Both optional (JSON) |
| creation_time | creation_time | ✅ types.Int64 | Both computed |
| revision_id | revision_id | ✅ types.Int64 | Both computed |

**Result**: 100% schema consistency achieved. No mismatches detected.

---

## References

- **Specification**: `/workspace/specs/001-cmkube-clusters-datasource/spec.md`
- **Research**: `/workspace/specs/001-cmkube-clusters-datasource/research.md`
- **Resource Schema**: `/workspace/internal/provider/resource_cmkube_cluster.go:106-200`
- **Pattern Source**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go:34-100`
