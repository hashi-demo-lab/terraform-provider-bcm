# Developer Quickstart: BCM CMDevice Nodes Data Source

**Feature**: bcm_cmdevice_nodes
**Phase**: 1 - Design & Contracts
**Audience**: Developers implementing the data source

## Overview

This guide provides step-by-step instructions for implementing the `bcm_cmdevice_nodes` data source using Test-Driven Development (TDD). Follow the RED-GREEN-REFACTOR cycle with acceptance tests.

**Estimated Time**: 4-6 hours (including testing and documentation)

---

## Prerequisites

### Required Tools

- **Go**: 1.24 or later
- **Terraform CLI**: 1.5+ (for manual testing)
- **Make**: For build automation
- **golangci-lint**: For code quality checks

### Access Requirements

- **BCM Cluster**: https://172.21.15.254:8081
- **Credentials**: root / Hashicorp123!
- **Network Access**: HTTPS access to BCM cluster

### Verification

```bash
# Check Go version
go version  # Should be 1.24+

# Check Terraform
terraform version  # Should be 1.5+

# Check Make
make --version

# Check golangci-lint
golangci-lint --version

# Test BCM API access (Python client)
cd /workspace/sampleRest
python3 capture_device_api.py  # Should connect successfully
```

---

## Project Setup

### 1. Clone and Branch

```bash
cd /workspace
git checkout 002-bcm-cmdevice-nodes-datasource
git status  # Verify on correct branch
```

### 2. Review Existing Code

**Study these files first**:

```bash
# Provider infrastructure
cat internal/provider/provider.go
cat internal/provider/bcm_client.go

# Reference implementation
cat internal/provider/data_source_cmpart_softwareimages.go
cat internal/provider/data_source_cmpart_softwareimages_test.go

# Helper functions
grep -A 5 "func getStringValue" internal/provider/data_source_cmpart_softwareimages.go
```

### 3. Environment Variables

Set up for acceptance tests:

```bash
export TF_ACC=1
export BCM_ENDPOINT=https://172.21.15.254:8081
export BCM_USERNAME=root
export BCM_PASSWORD='Hashicorp123!'
export BCM_INSECURE=true
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log
```

Add to `.bashrc` or `.zshrc` for persistence:

```bash
cat >> ~/.bashrc <<'EOF'
# BCM Terraform Provider Test Environment
export BCM_ENDPOINT=https://172.21.15.254:8081
export BCM_USERNAME=root
export BCM_PASSWORD='Hashicorp123!'
export BCM_INSECURE=true
EOF
```

### 4. Verify Existing Tests

```bash
# Run existing tests to ensure environment is working
make test

# Run existing acceptance test (without TF_ACC=1, should skip)
go test -v ./internal/provider/ -run TestAccCMPartSoftwareImagesDataSource

# Run with TF_ACC=1 (should pass)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImagesDataSource
```

---

## TDD Cycle 1: RED Phase

**Goal**: Write failing acceptance tests that define expected behavior

### Step 1: Create Test File

```bash
vim internal/provider/data_source_cmdevice_nodes_test.go
```

**Initial Test Structure**:

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"testing"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// Basic test - query all nodes
func TestAccCMDeviceNodesDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_basic,
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source exists
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "id"),
					// Verify nodes array exists
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.#"),
					// Verify at least one node (assuming test cluster has nodes)
					resource.TestCheckResourceAttr("data.bcm_cmdevice_nodes.test", "nodes.#", "2"),
				),
			},
		},
	})
}

const testAccCMDeviceNodesDataSourceConfig_basic = `
data "bcm_cmdevice_nodes" "test" {}
`

// Filter by node type
func TestAccCMDeviceNodesDataSource_FilterByType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_filterType,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.#"),
					// Note: Exact count depends on test cluster configuration
				),
			},
		},
	})
}

const testAccCMDeviceNodesDataSourceConfig_filterType = `
data "bcm_cmdevice_nodes" "test" {
  filter {
    node_type = "PhysicalNode"
  }
}
`

// Filter by hostname pattern
func TestAccCMDeviceNodesDataSource_FilterByHostname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_filterHostname,
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.#"),
				),
			},
		},
	})
}

const testAccCMDeviceNodesDataSourceConfig_filterHostname = `
data "bcm_cmdevice_nodes" "test" {
  filter {
    hostname_pattern = "node"
  }
}
`
```

### Step 2: Run Tests (Should Fail)

```bash
# Run test - expected to fail (data source doesn't exist yet)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource

# Expected output:
# Error: data source "bcm_cmdevice_nodes" is not registered
```

**RED Phase Complete**: Tests fail as expected

---

## TDD Cycle 2: GREEN Phase

**Goal**: Write minimal code to make tests pass

### Step 1: Create Data Source File

```bash
vim internal/provider/data_source_cmdevice_nodes.go
```

**Minimal Implementation**:

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &CMDeviceNodesDataSource{}
	_ datasource.DataSourceWithConfigure = &CMDeviceNodesDataSource{}
)

func NewCMDeviceNodesDataSource() datasource.DataSource {
	return &CMDeviceNodesDataSource{}
}

type CMDeviceNodesDataSource struct {
	client *BCMClient
}

type CMDeviceNodesDataSourceModel struct {
	ID     types.String `tfsdk:"id"`
	Filter *FilterModel `tfsdk:"filter"`
	Nodes  []NodeModel  `tfsdk:"nodes"`
}

type FilterModel struct {
	NodeType        types.String `tfsdk:"node_type"`
	CategoryUUID    types.String `tfsdk:"category_uuid"`
	HostnamePattern types.String `tfsdk:"hostname_pattern"`
}

type NodeModel struct {
	ID                    types.String `tfsdk:"id"`
	UUID                  types.String `tfsdk:"uuid"`
	Hostname              types.String `tfsdk:"hostname"`
	BaseType              types.String `tfsdk:"base_type"`
	ChildType             types.String `tfsdk:"child_type"`
	MAC                   types.String `tfsdk:"mac"`
	CreationTime          types.Int64  `tfsdk:"creation_time"`
	Interfaces            []NetworkInterfaceModel `tfsdk:"interfaces"`
	Roles                 []RoleModel `tfsdk:"roles"`
	Category              types.String `tfsdk:"category"`
	Partition             types.String `tfsdk:"partition"`
	PowerControl          types.String `tfsdk:"power_control"`
	AuthenticationService types.String `tfsdk:"authentication_service"`
	ProvisioningTransport types.String `tfsdk:"provisioning_transport"`
	Modified              types.Bool `tfsdk:"modified"`
	ToBeRemoved           types.Bool `tfsdk:"to_be_removed"`
}

type NetworkInterfaceModel struct {
	Name      types.String `tfsdk:"name"`
	MAC       types.String `tfsdk:"mac"`
	IP        types.String `tfsdk:"ip"`
	IPv6IP    types.String `tfsdk:"ipv6_ip"`
	DHCP      types.Bool   `tfsdk:"dhcp"`
	Network   types.String `tfsdk:"network"`
	BaseType  types.String `tfsdk:"base_type"`
	ChildType types.String `tfsdk:"child_type"`
	CardType  types.String `tfsdk:"cardtype"`
	Bootable  types.Bool   `tfsdk:"bootable"`
	StartIf   types.String `tfsdk:"start_if"`
}

type RoleModel struct {
	UUID        types.String `tfsdk:"uuid"`
	Name        types.String `tfsdk:"name"`
	BaseType    types.String `tfsdk:"base_type"`
	ChildType   types.String `tfsdk:"child_type"`
	AddServices types.Bool   `tfsdk:"add_services"`
}

func (d *CMDeviceNodesDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cmdevice_nodes"
}

func (d *CMDeviceNodesDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches cluster nodes from BCM CMDevice service",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				MarkdownDescription: "Placeholder identifier",
			},
			"nodes": schema.ListNestedAttribute{
				Computed: true,
				MarkdownDescription: "List of cluster nodes",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed: true,
						},
						"uuid": schema.StringAttribute{
							Computed: true,
						},
						"hostname": schema.StringAttribute{
							Computed: true,
						},
						"base_type": schema.StringAttribute{
							Computed: true,
						},
						"child_type": schema.StringAttribute{
							Computed: true,
						},
						"mac": schema.StringAttribute{
							Computed: true,
						},
						"creation_time": schema.Int64Attribute{
							Computed: true,
						},
						"category": schema.StringAttribute{
							Computed: true,
						},
						"partition": schema.StringAttribute{
							Computed: true,
						},
						"power_control": schema.StringAttribute{
							Computed: true,
						},
						"authentication_service": schema.StringAttribute{
							Computed: true,
						},
						"provisioning_transport": schema.StringAttribute{
							Computed: true,
						},
						"modified": schema.BoolAttribute{
							Computed: true,
						},
						"to_be_removed": schema.BoolAttribute{
							Computed: true,
						},
				// Placeholder for interfaces (will be completed in REFACTOR phase with full nested attributes)
				"interfaces": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed: true,
							},
							// Additional interface attributes added in REFACTOR phase
						},
					},
				},
				// Placeholder for roles (will be completed in REFACTOR phase with full nested attributes)
				"roles": schema.ListNestedAttribute{
					Computed: true,
					NestedObject: schema.NestedAttributeObject{
						Attributes: map[string]schema.Attribute{
							"name": schema.StringAttribute{
								Computed: true,
							},
							// Additional role attributes added in REFACTOR phase
						},
					},
				},
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				MarkdownDescription: "Filter nodes",
				Attributes: map[string]schema.Attribute{
					"node_type": schema.StringAttribute{
						Optional: true,
					},
					"category_uuid": schema.StringAttribute{
						Optional: true,
					},
					"hostname_pattern": schema.StringAttribute{
						Optional: true,
					},
				},
			},
		},
	}
}

func (d *CMDeviceNodesDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*BCMClient)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			"Expected *BCMClient",
		)
		return
	}

	d.client = client
}

func (d *CMDeviceNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state CMDeviceNodesDataSourceModel

	// MINIMAL IMPLEMENTATION - Hardcoded test data
	state.ID = types.StringValue("placeholder")
	state.Nodes = []NodeModel{
		{
			ID:       types.StringValue("test-uuid-1"),
			UUID:     types.StringValue("test-uuid-1"),
			Hostname: types.StringValue("test-node-1"),
		},
		{
			ID:       types.StringValue("test-uuid-2"),
			UUID:     types.StringValue("test-uuid-2"),
			Hostname: types.StringValue("test-node-2"),
		},
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

### Step 2: Register Data Source

```bash
vim internal/provider/provider.go
```

Find the `DataSources` method and add:

```go
func (p *bcmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewCMPartSoftwareImagesDataSource,
		NewCMDeviceNodesDataSource, // ADD THIS LINE
	}
}
```

### Step 3: Run Tests (Should Pass with Minimal Data)

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Basic

# Expected: PASS (with hardcoded data)
```

**GREEN Phase Complete**: Tests pass with minimal implementation

---

## TDD Cycle 3: REFACTOR Phase

**Goal**: Replace hardcoded data with real API calls, add full schema

### Step 1: Complete Schema (Nested Attributes)

Add full nested schemas for interfaces and roles (see data-model.md for complete definition).

### Step 2: Implement Real API Call

Replace Read() method:

```go
func (d *CMDeviceNodesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config CMDeviceNodesDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

	// Call BCM API
	body, err := d.client.CallJSONRPC(ctx, "cmdevice", "getNodes")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BCM Nodes",
			"Error: "+err.Error(),
		)
		return
	}

	// Parse response
	var apiResponse []map[string]interface{}
	if err := json.Unmarshal(body, &apiResponse); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse API Response",
			"Error: "+err.Error(),
		)
		return
	}

	// Map to models
	state := CMDeviceNodesDataSourceModel{
		ID:     types.StringValue("placeholder"),
		Filter: config.Filter,
		Nodes:  make([]NodeModel, 0, len(apiResponse)),
	}

	for _, nodeData := range apiResponse {
		node := mapAPIToNode(nodeData)
		if matchesFilter(node, config.Filter) {
			state.Nodes = append(state.Nodes, node)
		}
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

### Step 3: Add Helper Functions

Implement mapAPIToNode(), filterNodes(), and nested mappers (see data-model.md).

### Step 4: Run Full Test Suite

```bash
# Run all acceptance tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource

# Expected: All tests pass with real API data
```

### Step 5: Code Quality

```bash
# Format code
make fmt

# Run linter
golangci-lint run internal/provider/data_source_cmdevice_nodes.go

# Fix any issues
```

**REFACTOR Phase Complete**: Production-ready implementation

---

## Documentation Generation

### Generate Provider Docs

```bash
# Generate documentation
make generate

# Review generated docs
cat docs/data-sources/cmdevice_nodes.md

# Check for errors
git diff docs/
```

### Create Examples

```bash
mkdir -p examples/data-sources/bcm_cmdevice_nodes

# Create example files
cat > examples/data-sources/bcm_cmdevice_nodes/data-source.tf <<'EOF'
data "bcm_cmdevice_nodes" "all" {}

output "all_nodes" {
  value = data.bcm_cmdevice_nodes.all.nodes
}
EOF

cat > examples/data-sources/bcm_cmdevice_nodes/filter_by_type.tf <<'EOF'
data "bcm_cmdevice_nodes" "compute" {
  filter {
    node_type = "PhysicalNode"
  }
}

output "compute_nodes" {
  value = data.bcm_cmdevice_nodes.compute.nodes
}
EOF
```

---

## Manual Testing

### Build Provider

```bash
make install
```

### Create Test Configuration

```bash
mkdir -p /tmp/tf-test-cmdevice
cd /tmp/tf-test-cmdevice

cat > main.tf <<'EOF'
terraform {
  required_providers {
    bcm = {
      source = "hashicorp.com/edu/bcm"
    }
  }
}

provider "bcm" {
  endpoint = "https://172.21.15.254:8081"
  username = "root"
  password = "Hashicorp123!"
  insecure = true
}

data "bcm_cmdevice_nodes" "all" {}

output "nodes" {
  value = data.bcm_cmdevice_nodes.all.nodes
}

output "node_count" {
  value = length(data.bcm_cmdevice_nodes.all.nodes)
}

output "hostnames" {
  value = [for node in data.bcm_cmdevice_nodes.all.nodes : node.hostname]
}
EOF
```

### Test with Terraform

```bash
terraform init
terraform plan
terraform apply -auto-approve

# Verify output
terraform output nodes
terraform output hostnames
```

---

## Quality Checklist

Before submitting:

- [ ] All acceptance tests pass (TF_ACC=1)
- [ ] golangci-lint passes with no errors
- [ ] make fmt applied
- [ ] Documentation generated and reviewed
- [ ] Examples created and tested
- [ ] Manual Terraform test successful
- [ ] Error messages are clear
- [ ] Null handling tested
- [ ] Filter logic tested
- [ ] Nested attributes verified

---

## Troubleshooting

### Test Failures

**Problem**: Tests fail with "connection refused"
**Solution**: Verify BCM_ENDPOINT and network access

**Problem**: Tests fail with "unauthorized"
**Solution**: Check BCM_USERNAME and BCM_PASSWORD

**Problem**: Parse errors
**Solution**: Check API response format with Python client

### Build Issues

**Problem**: golangci-lint errors
**Solution**: Run `golangci-lint run --fix`

**Problem**: Import errors
**Solution**: Run `go mod tidy`

---

## Next Steps

1. **Review**: Code review with team
2. **Commit**: Create git commit with descriptive message
3. **Documentation**: Update project README if needed
4. **Integration**: Merge to main branch

---

## References

- Data Model: `data-model.md`
- API Contract: `contracts/cmdevice_getNodes_contract.json`
- Research: `research.md`
- Feature Spec: `spec.md`
- Existing Data Source: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
