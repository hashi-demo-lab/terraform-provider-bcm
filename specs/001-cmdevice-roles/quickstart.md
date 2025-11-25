# Quick Start: BCM Device Roles Data Source

**Feature**: bcm_cmdevice_roles data source
**Date**: 2025-11-25
**Estimated Time**: 2-3 hours (following TDD cycle)

## Prerequisites

- Go 1.24+ installed
- Terraform 1.0+ installed
- BCM cluster accessible with credentials
- Repository cloned: `terraform-provider-bcm`

## Environment Setup

```bash
# Set BCM credentials for acceptance tests
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Enable acceptance tests
export TF_ACC=1

# Optional: Enable detailed logging
export TF_LOG=DEBUG
```

## TDD Workflow: RED-GREEN-REFACTOR

### Phase 1: RED (Write Failing Tests) - 30 minutes

**Goal**: Write acceptance tests that fail because implementation doesn't exist yet.

**Step 1**: Create test file

```bash
cd /workspace
touch internal/provider/data_source_cmdevice_roles_test.go
```

**Step 2**: Write failing acceptance test

```go
// internal/provider/data_source_cmdevice_roles_test.go
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

func TestAccCMDeviceRolesDataSource_All(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceRolesDataSourceConfig(),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmdevice_roles.test",
                        tfjsonpath.New("id"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}

func testAccCMDeviceRolesDataSourceConfig() string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "test" {}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
    )
}
```

**Step 3**: Run test and verify it fails

```bash
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceRolesDataSource_All
# Expected: FAIL - data source not found
```

**Output**:
```
Error: data source "bcm_cmdevice_roles" not found in provider "bcm"
```

---

### Phase 2: GREEN (Minimal Implementation) - 60 minutes

**Goal**: Write minimal code to make tests pass.

**Step 1**: Create data source file

```bash
touch internal/provider/data_source_cmdevice_roles.go
```

**Step 2**: Implement minimal data source

```go
// internal/provider/data_source_cmdevice_roles.go
package provider

import (
    "context"
    "encoding/json"
    "fmt"
    "path/filepath"

    "github.com/hashicorp/terraform-plugin-framework/datasource"
    "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &CMDeviceRolesDataSource{}

func NewCMDeviceRolesDataSource() datasource.DataSource {
    return &CMDeviceRolesDataSource{}
}

type CMDeviceRolesDataSource struct {
    client *BCMClient
}

type CMDeviceRolesDataSourceModel struct {
    ID          types.String   `tfsdk:"id"`
    NamePattern types.String   `tfsdk:"name_pattern"`
    ChildType   types.String   `tfsdk:"child_type"`
    Roles       []RoleModel    `tfsdk:"roles"`
}

type RoleModel struct {
    ID          types.String `tfsdk:"id"`
    UUID        types.String `tfsdk:"uuid"`
    Name        types.String `tfsdk:"name"`
    ChildType   types.String `tfsdk:"child_type"`
    BaseType    types.String `tfsdk:"base_type"`
    AddServices types.Bool   `tfsdk:"add_services"`
}

func (d *CMDeviceRolesDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cmdevice_roles"
}

func (d *CMDeviceRolesDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Fetches available role types from BCM for device role assignment.",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
                MarkdownDescription: "Data source identifier",
            },
            "name_pattern": schema.StringAttribute{
                Optional: true,
                MarkdownDescription: "Glob pattern to filter roles by name (e.g., 'kube-*')",
            },
            "child_type": schema.StringAttribute{
                Optional: true,
                MarkdownDescription: "Exact match filter for role type (e.g., 'ComputeRole')",
            },
            "roles": schema.ListNestedAttribute{
                Computed: true,
                MarkdownDescription: "List of roles matching filter criteria",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "id": schema.StringAttribute{
                            Computed: true,
                            MarkdownDescription: "Role identifier",
                        },
                        "uuid": schema.StringAttribute{
                            Computed: true,
                            MarkdownDescription: "Unique role identifier",
                        },
                        "name": schema.StringAttribute{
                            Computed: true,
                            MarkdownDescription: "Role name",
                        },
                        "child_type": schema.StringAttribute{
                            Computed: true,
                            MarkdownDescription: "Role type",
                        },
                        "base_type": schema.StringAttribute{
                            Computed: true,
                            MarkdownDescription: "Base type (always 'Role')",
                        },
                        "add_services": schema.BoolAttribute{
                            Computed: true,
                            MarkdownDescription: "Whether role adds services",
                        },
                    },
                },
            },
        },
    }
}

func (d *CMDeviceRolesDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *CMDeviceRolesDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    var data CMDeviceRolesDataSourceModel

    resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call BCM API to get nodes
    tflog.Debug(ctx, "Calling cmdevice.getNodes to extract roles")
    result, err := d.client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    if err != nil {
        resp.Diagnostics.AddError("Error Reading Roles", fmt.Sprintf("Could not query nodes: %s", err.Error()))
        return
    }

    // Parse nodes
    var nodes []map[string]interface{}
    if err := json.Unmarshal(result, &nodes); err != nil {
        resp.Diagnostics.AddError("Error Parsing Nodes", fmt.Sprintf("Could not parse nodes: %s", err.Error()))
        return
    }

    // Extract and deduplicate roles
    roleMap := make(map[string]map[string]interface{})
    for _, node := range nodes {
        if rolesData, ok := node["roles"].([]interface{}); ok {
            for _, roleData := range rolesData {
                if role, ok := roleData.(map[string]interface{}); ok {
                    uuid := getStringValue(role, "uuid")
                    if !uuid.IsNull() {
                        roleMap[uuid.ValueString()] = role
                    }
                }
            }
        }
    }

    tflog.Debug(ctx, fmt.Sprintf("Found %d unique roles across %d nodes", len(roleMap), len(nodes)))

    // Convert to slice and apply filters
    data.Roles = make([]RoleModel, 0, len(roleMap))
    for _, roleData := range roleMap {
        role := RoleModel{
            UUID:        getStringValue(roleData, "uuid"),
            Name:        getStringValue(roleData, "name"),
            ChildType:   getStringValue(roleData, "childType"),
            BaseType:    getStringValue(roleData, "baseType"),
            AddServices: getBoolValue(roleData, "addServices"),
        }
        role.ID = role.UUID

        // Apply filters
        if matchesRoleFilter(role, data.NamePattern, data.ChildType) {
            data.Roles = append(data.Roles, role)
        }
    }

    data.ID = types.StringValue("roles")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

// matchesRoleFilter applies AND logic for multiple filters
func matchesRoleFilter(role RoleModel, namePattern, childType types.String) bool {
    // name_pattern check (glob matching)
    if !namePattern.IsNull() && !namePattern.IsUnknown() {
        pattern := namePattern.ValueString()
        matched, err := filepath.Match(pattern, role.Name.ValueString())
        if err != nil || !matched {
            return false
        }
    }

    // child_type check (exact match)
    if !childType.IsNull() && !childType.IsUnknown() {
        if role.ChildType.ValueString() != childType.ValueString() {
            return false
        }
    }

    return true
}
```

**Step 3**: Register data source in provider

```go
// internal/provider/provider.go
func (p *BCMProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        NewCMDeviceNodesDataSource,
        NewCMPartSoftwareImagesDataSource,
        NewCMNetNetworksDataSource,
        NewCMDeviceCategoriesDataSource,
        NewCMKubeClustersDataSource,
        NewCMDeviceRolesDataSource,  // ADD THIS LINE
    }
}
```

**Step 4**: Run test and verify it passes

```bash
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceRolesDataSource_All
# Expected: PASS
```

---

### Phase 3: REFACTOR (Add Remaining Tests) - 60 minutes

**Goal**: Add comprehensive test coverage and refine implementation.

**Step 1**: Add filter tests

```go
// internal/provider/data_source_cmdevice_roles_test.go

func TestAccCMDeviceRolesDataSource_FilterByChildType(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceRolesDataSourceConfigFilterByChildType("ComputeRole"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmdevice_roles.test",
                        tfjsonpath.New("id"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}

func testAccCMDeviceRolesDataSourceConfigFilterByChildType(childType string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "test" {
  child_type = %[4]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        childType,
    )
}

func TestAccCMDeviceRolesDataSource_FilterByNamePattern(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMDeviceRolesDataSourceConfigFilterByNamePattern("*"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "data.bcm_cmdevice_roles.test",
                        tfjsonpath.New("id"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}

func testAccCMDeviceRolesDataSourceConfigFilterByNamePattern(pattern string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "test" {
  name_pattern = %[4]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        pattern,
    )
}
```

**Step 2**: Run all tests

```bash
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMDeviceRoles
# Expected: All tests PASS
```

---

### Phase 4: Examples & Documentation - 30 minutes

**Step 1**: Create example configurations

```bash
mkdir -p examples/data-sources/bcm_cmdevice_roles
```

**Example 1**: Query all roles
```hcl
# examples/data-sources/bcm_cmdevice_roles/data-source.tf
data "bcm_cmdevice_roles" "all" {}

output "available_roles" {
  value = [for role in data.bcm_cmdevice_roles.all.roles : role.name]
}
```

**Example 2**: Filter by type
```hcl
# examples/data-sources/bcm_cmdevice_roles/filter-by-type.tf
data "bcm_cmdevice_roles" "compute" {
  child_type = "ComputeRole"
}

output "compute_roles" {
  value = data.bcm_cmdevice_roles.compute.roles
}
```

**Example 3**: Filter by pattern
```hcl
# examples/data-sources/bcm_cmdevice_roles/filter-by-pattern.tf
data "bcm_cmdevice_roles" "kube_roles" {
  name_pattern = "kube-*"
}

output "kube_roles" {
  value = data.bcm_cmdevice_roles.kube_roles.roles
}
```

**Step 2**: Generate documentation

```bash
make generate
# Runs tfplugindocs to generate docs/data-sources/bcm_cmdevice_roles.md
```

**Step 3**: Verify documentation

```bash
cat docs/data-sources/bcm_cmdevice_roles.md
```

---

## Verification Checklist

- [ ] All acceptance tests pass
- [ ] Examples created in `examples/data-sources/bcm_cmdevice_roles/`
- [ ] Documentation generated in `docs/data-sources/bcm_cmdevice_roles.md`
- [ ] Data source registered in `provider.go`
- [ ] Code follows existing patterns (null-safe helpers, filter logic)
- [ ] Manual test on live BCM cluster successful

## Manual Testing

```bash
# Create test configuration
cat > test-roles.tf <<EOF
terraform {
  required_providers {
    bcm = {
      source = "hashicorp/bcm"
    }
  }
}

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "all" {}

output "roles" {
  value = data.bcm_cmdevice_roles.all.roles
}
EOF

# Initialize and apply
terraform init
terraform plan
terraform apply -auto-approve

# Verify output shows roles
terraform output -json roles
```

## Troubleshooting

**Problem**: Data source not found
- **Solution**: Verify data source registered in `provider.go` DataSources() method

**Problem**: Test fails with API error
- **Solution**: Check BCM credentials in environment variables

**Problem**: Empty roles list
- **Solution**: Verify BCM cluster has nodes with roles assigned

**Problem**: Filter not working
- **Solution**: Check glob pattern syntax, verify childType exact match case

## Next Steps

After successful implementation:

1. Create pull request with changes
2. Run CI/CD pipeline for full test suite
3. Update CHANGELOG.md with new data source
4. Consider adding data source to provider documentation website

## Time Estimate Breakdown

- RED phase (failing tests): 30 minutes
- GREEN phase (minimal implementation): 60 minutes
- REFACTOR phase (complete tests): 60 minutes
- Examples & documentation: 30 minutes
- **Total**: 3 hours

## References

- **Spec**: `/workspace/specs/001-cmdevice-roles/spec.md`
- **Research**: `/workspace/specs/001-cmdevice-roles/research.md`
- **Data Model**: `/workspace/specs/001-cmdevice-roles/data-model.md`
- **Contracts**: `/workspace/specs/001-cmdevice-roles/contracts/`
- **TDD Guide**: `/workspace/AGENTS.md`
- **Project Guide**: `/workspace/CLAUDE.md`
