# Quick Start: Implementing bcm_cmkube_clusters Data Source

**Target Audience**: Developers implementing the data source using TDD workflow
**Time to Complete**: 2-3 hours (RED: 30min, GREEN: 45min, REFACTOR: 60min, DOCS: 30min)

---

## Prerequisites

### Environment Setup

1. **Go Environment**:
   ```bash
   go version  # Verify Go 1.24+
   ```

2. **BCM Test Environment**:
   ```bash
   export BCM_ENDPOINT="https://172.21.15.254:8081"
   export BCM_USERNAME="root"
   export BCM_PASSWORD="Hashicorp123!"
   export TF_ACC=1  # Enable acceptance tests
   ```

3. **Verify BCM API Access**:
   ```bash
   cd /workspace/sampleRest
   python3 cmkube-get-clusters.py  # Should list clusters
   ```

### Read These First

- **Spec**: `/workspace/specs/001-cmkube-clusters-datasource/spec.md` (requirements)
- **Research**: `/workspace/specs/001-cmkube-clusters-datasource/research.md` (API details)
- **Data Model**: `/workspace/specs/001-cmkube-clusters-datasource/data-model.md` (schema design)
- **Pattern Example**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go` (reference implementation)

---

## Phase 1: RED - Write Failing Tests (30 minutes)

### Step 1.1: Create Test File

**File**: `/workspace/internal/provider/data_source_cmkube_clusters_test.go`

**Template**:
```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
    "fmt"
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCMKubeClustersDataSource_Basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMKubeClustersDataSourceConfig_basic(),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmkube_clusters.test",
                        tfjsonpath.New("id"),
                        knownvalue.NotNull(),
                    ),
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmkube_clusters.test",
                        tfjsonpath.New("clusters").AtSliceIndex(0).AtMapKey("uuid"),
                        knownvalue.NotNull(),
                    ),
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmkube_clusters.test",
                        tfjsonpath.New("clusters").AtSliceIndex(0).AtMapKey("name"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}

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

### Step 1.2: Add Remaining Tests

**Add these test functions** (7 total):
- `TestAccCMKubeClustersDataSource_Basic` ✅ (above)
- `TestAccCMKubeClustersDataSource_FilterByName`
- `TestAccCMKubeClustersDataSource_FilterByVersion`
- `TestAccCMKubeClustersDataSource_FilterByMasterNode`
- `TestAccCMKubeClustersDataSource_MultipleFilters`
- `TestAccCMKubeClustersDataSource_EmptyResults`
- `TestAccCMKubeClustersDataSource_NullFields`

**Reference**: `/workspace/internal/provider/data_source_cmpart_softwareimages_test.go` for filter test patterns

### Step 1.3: Run Tests (Verify RED)

```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource
```

**Expected**: All 7 tests FAIL with "data source not found" errors

---

## Phase 2: GREEN - Minimal Implementation (45 minutes)

### Step 2.1: Create Data Source File

**File**: `/workspace/internal/provider/data_source_cmkube_clusters.go`

**Minimal Implementation**:
```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure implementation satisfies interfaces
var (
    _ datasource.DataSource              = &CMKubeClustersDataSource{}
    _ datasource.DataSourceWithConfigure = &CMKubeClustersDataSource{}
)

func NewCMKubeClustersDataSource() datasource.DataSource {
    return &CMKubeClustersDataSource{}
}

type CMKubeClustersDataSource struct {
    client *BCMClient
}

type CMKubeClustersDataSourceModel struct {
    ID       types.String         `tfsdk:"id"`
    Filter   *ClusterFilterModel  `tfsdk:"filter"`
    Clusters []KubeClusterModel   `tfsdk:"clusters"`
}

type ClusterFilterModel struct {
    NamePattern   types.String `tfsdk:"name_pattern"`
    Version       types.String `tfsdk:"version"`
    MasterNodeID  types.String `tfsdk:"master_node_id"`
}

type KubeClusterModel struct {
    // Copy from data-model.md (all fields)
    ID                types.String `tfsdk:"id"`
    UUID              types.String `tfsdk:"uuid"`
    Name              types.String `tfsdk:"name"`
    MasterNodes       types.List   `tfsdk:"master_nodes"`
    WorkerNodes       types.List   `tfsdk:"worker_nodes"`
    ManagementNetwork types.String `tfsdk:"management_network"`
    OverlayNetwork    types.String `tfsdk:"overlay_network"`
    DNSServers        types.List   `tfsdk:"dns_servers"`
    Version           types.String `tfsdk:"version"`
    CNIPlugin         types.String `tfsdk:"cni_plugin"`
    StorageClasses    types.String `tfsdk:"storage_classes"`
    LoadBalancerMode  types.String `tfsdk:"load_balancer_mode"`
    Addons            types.String `tfsdk:"addons"`
    IngressController types.String `tfsdk:"ingress_controller"`
    CreationTime      types.Int64  `tfsdk:"creation_time"`
    RevisionID        types.Int64  `tfsdk:"revision_id"`
}

func (d *CMKubeClustersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cmkube_clusters"
}

func (d *CMKubeClustersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    // Copy from data-model.md (complete schema)
    resp.Schema = schema.Schema{
        MarkdownDescription: "Data source to discover and filter BCM Kubernetes clusters.",
        // ... complete schema definition
    }
}

func (d *CMKubeClustersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*BCMClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Data Source Configure Type",
            fmt.Sprintf("Expected *BCMClient, got: %T", req.ProviderData),
        )
        return
    }

    d.client = client
}

func (d *CMKubeClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data CMKubeClustersDataSourceModel

    // Minimal GREEN implementation: Return empty clusters list
    data.ID = types.StringValue("cmkube-clusters")
    data.Clusters = []KubeClusterModel{}

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Step 2.2: Register Data Source

**File**: `/workspace/internal/provider/provider.go`

**Add to `DataSources()` method**:
```go
func (p *BCMProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewCMDeviceCategoriesDataSource,
        NewCMDeviceNodesDataSource,
        NewCMNetNetworksDataSource,
        NewCMPartPartitionsDataSource,
        NewCMPartSoftwareImagesDataSource,
        NewCMKubeClustersDataSource,  // <-- ADD THIS LINE
    }
}
```

### Step 2.3: Run Tests (Verify GREEN)

```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource_Basic
```

**Expected**: Basic test PASSES (empty clusters list is valid)

---

## Phase 3: REFACTOR - Full Implementation (60 minutes)

### Step 3.1: Add BCM API Call

**Update `Read()` method**:
```go
func (d *CMKubeClustersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data CMKubeClustersDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call BCM API
    body, err := d.client.CallJSONRPC(ctx, "cmkube", "getKubeClusters")
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading Clusters",
            "Could not read Kubernetes clusters: "+err.Error(),
        )
        return
    }

    // Parse response
    var apiResponse []map[string]interface{}
    if err := json.Unmarshal(body, &apiResponse); err != nil {
        resp.Diagnostics.AddError(
            "Error Parsing Clusters",
            "Could not parse response: "+err.Error(),
        )
        return
    }

    // Apply filters and map to model
    var filteredClusters []KubeClusterModel
    for _, clusterData := range apiResponse {
        if matchesClusterFilter(data.Filter, clusterData) {
            model := mapClusterDataToModel(clusterData)
            filteredClusters = append(filteredClusters, model)
        }
    }

    data.Clusters = filteredClusters
    data.ID = types.StringValue("cmkube-clusters")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Step 3.2: Implement Filter Logic

**Add helper function**:
```go
func matchesClusterFilter(filter *ClusterFilterModel, clusterData map[string]interface{}) bool {
    if filter == nil {
        return true
    }

    // Filter 1: name_pattern (case-insensitive substring)
    if !filter.NamePattern.IsNull() {
        pattern := strings.ToLower(filter.NamePattern.ValueString())
        name := strings.ToLower(getStringValue(clusterData, "name").ValueString())
        if !strings.Contains(name, pattern) {
            return false
        }
    }

    // Filter 2: version (exact match)
    if !filter.Version.IsNull() {
        version := getStringValue(clusterData, "version").ValueString()
        if version != filter.Version.ValueString() {
            return false
        }
    }

    // Filter 3: master_node_id (ANY match)
    if !filter.MasterNodeID.IsNull() {
        targetNodeID := filter.MasterNodeID.ValueString()
        found := false

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
            return false
        }
    }

    return true  // All filters passed
}
```

### Step 3.3: Implement Data Mapping

**Add helper functions**:
```go
func mapClusterDataToModel(apiData map[string]interface{}) KubeClusterModel {
    model := KubeClusterModel{}

    uuid := getStringValue(apiData, "uuid")
    model.ID = uuid
    model.UUID = uuid
    model.Name = getStringValue(apiData, "name")

    model.MasterNodes = getListValue(apiData, "masterNodes")
    model.WorkerNodes = getListValue(apiData, "workerNodes")

    model.ManagementNetwork = getStringValue(apiData, "managementNetwork")
    model.OverlayNetwork = getStringValue(apiData, "overlayNetwork")
    model.DNSServers = getListValue(apiData, "dnsServers")

    model.Version = getStringValue(apiData, "version")
    model.CNIPlugin = getStringValue(apiData, "cniPlugin")

    model.StorageClasses = getStringValue(apiData, "storageClasses")
    model.LoadBalancerMode = getStringValue(apiData, "loadBalancerMode")

    model.Addons = getStringValue(apiData, "addons")
    model.IngressController = getStringValue(apiData, "ingressController")

    model.CreationTime = getInt64Value(apiData, "creationTime")
    model.RevisionID = getInt64Value(apiData, "revisionID")

    return model
}

func getListValue(data map[string]interface{}, key string) types.List {
    if val, ok := data[key]; ok && val != nil {
        if slice, ok := val.([]interface{}); ok && len(slice) > 0 {
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
    return types.ListNull(types.StringType)
}
```

**Reuse existing helpers**: `getStringValue()`, `getInt64Value()` from `data_source_cmpart_softwareimages.go`

### Step 3.4: Run All Tests

```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeClustersDataSource
```

**Expected**: All 7 tests PASS

---

## Phase 4: DOCUMENTATION - Examples & Docs (30 minutes)

### Step 4.1: Create Example File

**File**: `/workspace/examples/data-sources/bcm_cmkube_clusters/data-source.tf`

```hcl
# List all Kubernetes clusters
data "bcm_cmkube_clusters" "all" {}

# Filter clusters by name pattern (case-insensitive)
data "bcm_cmkube_clusters" "production" {
  filter {
    name_pattern = "prod"
  }
}

# Filter clusters by Kubernetes version
data "bcm_cmkube_clusters" "k8s_v128" {
  filter {
    version = "1.28.0"
  }
}

# Find clusters containing a specific master node
data "bcm_cmkube_clusters" "by_master_node" {
  filter {
    master_node_id = "node-abc-123"
  }
}

# Combine multiple filters (AND logic)
data "bcm_cmkube_clusters" "prod_v128" {
  filter {
    name_pattern = "prod"
    version      = "1.28.0"
  }
}

# Use for terraform import
output "cluster_uuids" {
  value = [for cluster in data.bcm_cmkube_clusters.all.clusters : cluster.uuid]
}

# Example import command (run in shell):
# terraform import bcm_cmkube_cluster.imported <uuid>
```

### Step 4.2: Generate Documentation

```bash
cd /workspace
make generate
```

**Verify**:
- Documentation generated at: `/workspace/docs/data-sources/cmkube_clusters.md`
- No errors in output
- Examples included in documentation

### Step 4.3: Run Linting

```bash
cd /workspace
make lint
```

**Fix any issues** reported by golangci-lint

---

## Verification Checklist

Before marking as complete, verify:

- [ ] All 7 acceptance tests pass (`TF_ACC=1 go test -run TestAccCMKubeClustersDataSource`)
- [ ] Data source registered in `provider.go` DataSources() method
- [ ] Examples created in `examples/data-sources/bcm_cmkube_clusters/data-source.tf`
- [ ] Documentation generated (`make generate` succeeds)
- [ ] Linting passes (`make lint` succeeds)
- [ ] Schema matches `resource_cmkube_cluster.go` (cross-reference)
- [ ] All 12 functional requirements from spec.md verified by tests
- [ ] Filter logic uses AND (not OR) for multiple filters
- [ ] Null-safe extraction for all fields

---

## Troubleshooting

### Test Failures

**Problem**: "unknown data source bcm_cmkube_clusters"
**Solution**: Verify data source registered in `provider.go` DataSources() method

**Problem**: "clusters.0.master_nodes: attribute not found"
**Solution**: Check schema definition - master_nodes must be `ListAttribute` with `ElementType: types.StringType`

### API Errors

**Problem**: "Error Reading Clusters: authentication failed"
**Solution**: Verify BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD environment variables

**Problem**: "Error Parsing Clusters: invalid JSON"
**Solution**: Check BCM API response format - should be array, not single object

### Filter Issues

**Problem**: Filter returns no results (expected matches)
**Solution**: Verify case-insensitive matching - use `strings.ToLower()` for both pattern and cluster name

---

## Next Steps

After completing implementation:

1. Run full test suite: `make test`
2. Run acceptance tests: `make testacc`
3. Commit changes to feature branch
4. Create pull request for review
5. Update GitHub issue #27 with progress

---

## References

- **TDD Guide**: `/workspace/AGENTS.md` (Terraform Provider TDD patterns)
- **Pattern Examples**: `/workspace/internal/provider/data_source_*.go`
- **Test Helpers**: `/workspace/internal/provider/test_helpers.go`
- **BCM API**: `/workspace/sampleRest/cmkube-get-clusters.py`
