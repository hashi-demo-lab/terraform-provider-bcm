# Quick Start: BCM CMDevice Category Resource Implementation

**Feature**: BCM CMDevice Category Resource
**Branch**: 001-cmdevice-category
**Date**: 2025-11-21

## Overview

This guide provides step-by-step instructions for implementing the BCM CMDevice Category Terraform resource following TDD (Test-Driven Development) principles with RED-GREEN-REFACTOR cycles.

---

## Prerequisites

### Development Environment

```bash
# Verify Go installation
go version  # Expected: go1.24 or higher

# Verify working directory
pwd  # Expected: /workspace (terraform-provider-bcm root)

# Verify existing resources
ls internal/provider/resource_cmpart_softwareimage.go  # Reference implementation

# Install pre-commit hooks (if not already done)
pre-commit install
```

### Test Environment

```bash
# Set BCM credentials for acceptance tests
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Verify BCM connectivity (optional)
cd sampleRest
python3 01-list-images.py  # Should list software images successfully
cd ..
```

---

## Implementation Workflow

### Phase 0: RED - Write Failing Tests (45-60 minutes)

#### Step 1: Create Test File

```bash
# Create acceptance test file
touch internal/provider/resource_cmdevice_category_test.go
```

#### Step 2: Write Basic CRUD Test

Create `internal/provider/resource_cmdevice_category_test.go`:

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
    "fmt"
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCMDeviceCategoryResource_Basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccCMDeviceCategoryResourceConfig("test-category"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", "test-category"),
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "id"),
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "uuid"),
                    resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "management_network"),
                ),
            },
            // ImportState testing
            {
                ResourceName:      "bcm_cmdevice_category.test",
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{"force"},
            },
            // Update and Read testing
            {
                Config: testAccCMDeviceCategoryResourceConfig_Updated("test-category-updated"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", "test-category-updated"),
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Updated notes"),
                ),
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}

func testAccCMDeviceCategoryResourceConfig(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Test category"
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}

func testAccCMDeviceCategoryResourceConfig_Updated(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Updated notes"
  kernel_parameters  = "quiet splash"
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}
```

#### Step 3: Run Tests (Expect Failures)

```bash
# Run acceptance tests - should FAIL (resource not registered)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Basic

# Expected output:
# --- FAIL: TestAccCMDeviceCategoryResource_Basic
# Error: provider does not support resource type "bcm_cmdevice_category"
```

**Checkpoint**: Tests fail with "resource not registered" error. This is expected (RED phase).

---

### Phase 1: GREEN - Minimal Implementation (60-90 minutes)

#### Step 4: Create Resource File

```bash
# Create resource implementation file
touch internal/provider/resource_cmdevice_category.go
```

#### Step 5: Implement Minimal Resource Structure

Create `internal/provider/resource_cmdevice_category.go`:

```go
// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
    "context"
    "fmt"

    "github.com/hashicorp/terraform-plugin-framework/path"
    "github.com/hashicorp/terraform-plugin-framework/resource"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema"
    "github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces
var (
    _ resource.Resource                = &CMDeviceCategoryResource{}
    _ resource.ResourceWithImportState = &CMDeviceCategoryResource{}
)

// CMDeviceCategoryResource defines the resource implementation
type CMDeviceCategoryResource struct {
    client *BCMClient
}

// CMDeviceCategoryResourceModel describes the resource data model
type CMDeviceCategoryResourceModel struct {
    ID                types.String `tfsdk:"id"`
    UUID              types.String `tfsdk:"uuid"`
    Name              types.String `tfsdk:"name"`
    Notes             types.String `tfsdk:"notes"`
    ManagementNetwork types.String `tfsdk:"management_network"`
    KernelParameters  types.String `tfsdk:"kernel_parameters"`
    Force             types.Bool   `tfsdk:"force"`
}

// NewCMDeviceCategoryResource creates a new resource instance
func NewCMDeviceCategoryResource() resource.Resource {
    return &CMDeviceCategoryResource{}
}

// Metadata returns the resource type name
func (r *CMDeviceCategoryResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cmdevice_category"
}

// Schema defines the resource schema (minimal for GREEN phase)
func (r *CMDeviceCategoryResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages a BCM device category.",

        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Resource identifier (same as UUID)",
            },
            "uuid": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Unique identifier assigned by BCM",
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Category name (must be unique)",
            },
            "notes": schema.StringAttribute{
                Optional:            true,
                MarkdownDescription: "Administrative notes",
            },
            "management_network": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "UUID of management network",
            },
            "kernel_parameters": schema.StringAttribute{
                Optional:            true,
                MarkdownDescription: "Kernel command-line parameters",
            },
            "force": schema.BoolAttribute{
                Optional:            true,
                MarkdownDescription: "Force operation even if warnings exist",
            },
        },
    }
}

// Configure stores the BCM client
func (r *CMDeviceCategoryResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
    if req.ProviderData == nil {
        return
    }

    client, ok := req.ProviderData.(*BCMClient)
    if !ok {
        resp.Diagnostics.AddError(
            "Unexpected Resource Configure Type",
            fmt.Sprintf("Expected *BCMClient, got: %T", req.ProviderData),
        )
        return
    }

    r.client = client
}

// Create implements the resource Create operation (MINIMAL for GREEN phase)
func (r *CMDeviceCategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan CMDeviceCategoryResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL: Hardcode UUID for now
    plan.ID = types.StringValue("test-uuid-12345")
    plan.UUID = types.StringValue("test-uuid-12345")

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read implements the resource Read operation (MINIMAL)
func (r *CMDeviceCategoryResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state CMDeviceCategoryResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL: Just return current state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update implements the resource Update operation (MINIMAL)
func (r *CMDeviceCategoryResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan CMDeviceCategoryResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL: Just save plan to state
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete implements the resource Delete operation (MINIMAL)
func (r *CMDeviceCategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state CMDeviceCategoryResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL: Nothing to do
}

// ImportState implements resource import
func (r *CMDeviceCategoryResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

#### Step 6: Register Resource in Provider

Edit `internal/provider/provider.go` and add to `Resources()` method:

```go
func (p *BCMProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewCMPartSoftwareImageResource,
        NewCMDeviceCategoryResource,  // ADD THIS LINE
    }
}
```

#### Step 7: Run Tests (Expect Pass)

```bash
# Run acceptance tests - should PASS with minimal implementation
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Basic

# Expected output:
# --- PASS: TestAccCMDeviceCategoryResource_Basic (X.XXs)
# PASS
```

**Checkpoint**: Tests pass with hardcoded values. This is GREEN phase success.

---

### Phase 2: REFACTOR - Real API Integration (90-120 minutes)

#### Step 8: Implement Full Schema

Update `resource_cmdevice_category.go` schema with all 60+ attributes (refer to data-model.md for complete schema).

Key additions:
- Boot configuration fields (boot_loader, kernel_version, etc.)
- Network configuration (default_gateway, name_servers, etc.)
- Nested objects (software_image_proxy, bmc_settings, fsmounts, modules)
- Long text fields (disksetup, exclude_list_*)
- All computed fields (parent_uuid, revision, etc.)

#### Step 9: Implement buildAPIEntity Helper

```go
// buildAPIEntity constructs BCM API entity from Terraform model
func (r *CMDeviceCategoryResource) buildAPIEntity(model *CMDeviceCategoryResourceModel, uuid string) map[string]interface{} {
    entity := map[string]interface{}{
        "baseType":      "Category",
        "childType":     "",
        "modified":      true,
        "to_be_removed": false,
        "revision":      "",
    }

    if uuid != "" {
        entity["uuid"] = uuid
    }

    if !model.Name.IsNull() {
        entity["name"] = model.Name.ValueString()
    }

    if !model.ManagementNetwork.IsNull() {
        entity["managementNetwork"] = model.ManagementNetwork.ValueString()
    }

    if !model.Notes.IsNull() {
        entity["notes"] = model.Notes.ValueString()
    }

    if !model.KernelParameters.IsNull() {
        entity["kernelParameters"] = model.KernelParameters.ValueString()
    }

    // Add all other fields following same pattern

    return entity
}
```

#### Step 10: Implement Real CRUD Operations

Replace minimal implementations with real API calls:

```go
func (r *CMDeviceCategoryResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan CMDeviceCategoryResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build API entity
    entity := r.buildAPIEntity(&plan, "")

    // Call addCategory API
    createBody, err := r.client.CallJSONRPC(ctx, "cmdevice", "addCategory", entity, plan.Force.ValueBool())
    if err != nil {
        resp.Diagnostics.AddError("Category Creation Failed", err.Error())
        return
    }

    // Parse UUID from response
    var createdUUID string
    json.Unmarshal(createBody, &createdUUID)

    plan.ID = types.StringValue(createdUUID)
    plan.UUID = types.StringValue(createdUUID)

    // Read back to populate computed fields
    r.readCategory(ctx, &plan, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

Similarly implement:
- `Read()` using `getCategory(name)` API
- `Update()` using `updateCategory` API
- `Delete()` using `removeCategory` API
- `readCategory()` helper method

#### Step 11: Run Tests (Expect Pass with Real API)

```bash
# Run acceptance tests with real BCM API
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource_Basic

# Expected output:
# --- PASS: TestAccCMDeviceCategoryResource_Basic (X.XXs)
# PASS
```

**Checkpoint**: Tests pass with real API integration. REFACTOR phase complete.

---

### Phase 3: Additional Tests (60-90 minutes)

#### Step 12: Add Complete Configuration Test

Add test for comprehensive category configuration with nested objects:

```go
func TestAccCMDeviceCategoryResource_Complete(t *testing.T) {
    // Test with nested objects, arrays, long text fields
    // Refer to spec.md for complete test scenario
}
```

#### Step 13: Add Nested Objects Test

```go
func TestAccCMDeviceCategoryResource_NestedObjects(t *testing.T) {
    // Test adding/updating/removing nested objects
}
```

#### Step 14: Run All Tests

```bash
# Run all category resource tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceCategoryResource

# Expected: All tests pass
```

---

### Phase 4: Documentation & Examples (30-45 minutes)

#### Step 15: Create Example Configuration

Create `examples/resources/bcm_cmdevice_category/resource.tf`:

```hcl
# Minimal category
resource "bcm_cmdevice_category" "minimal" {
  name               = "minimal-category"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
}

# Category with boot configuration
resource "bcm_cmdevice_category" "compute" {
  name               = "compute-nodes"
  management_network = var.management_network_id

  boot_loader          = "GRUB2"
  kernel_parameters    = "quiet splash intel_iommu=on"

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    }
  ]
}

# Category with nested objects
resource "bcm_cmdevice_category" "with_nested" {
  name               = "gpu-nodes"
  management_network = var.management_network_id

  bmc_settings = {
    user_name = "admin"
    password  = var.bmc_password
    privilege = "ADMINISTRATOR"
  }

  fsmounts = [
    {
      device       = "nfs-server:/export/home"
      mountpoint   = "/home"
      filesystem   = "nfs"
      mountoptions = "rsize=32768,wsize=32768"
    }
  ]
}
```

#### Step 16: Generate Documentation

```bash
# Generate provider documentation
make generate

# Verify documentation created
ls docs/resources/bcm_cmdevice_category.md
```

#### Step 17: Format and Lint

```bash
# Format Go code
make fmt

# Run linter
make lint

# Run pre-commit hooks
pre-commit run --all-files
```

---

## Verification Checklist

Before considering implementation complete:

- [ ] All 5 acceptance test scenarios pass
- [ ] Code coverage > 80% for resource implementation
- [ ] Documentation generated and accurate
- [ ] Examples directory populated with common use cases
- [ ] golangci-lint passes with no errors
- [ ] Pre-commit hooks pass
- [ ] Resource registered in provider.go
- [ ] Import functionality tested
- [ ] Force parameter behavior validated
- [ ] Sensitive fields marked appropriately (bmc_settings.password)
- [ ] Helper functions reused from existing resources
- [ ] Error messages clear and actionable

---

## Common Issues & Solutions

### Issue 1: Provider Not Registered

**Error**: `provider does not support resource type "bcm_cmdevice_category"`

**Solution**: Add `NewCMDeviceCategoryResource` to `provider.go` `Resources()` method

### Issue 2: API Authentication Failure

**Error**: `401 Unauthorized`

**Solution**: Verify BCM credentials and endpoint:
```bash
echo $BCM_ENDPOINT
echo $BCM_USERNAME
# Test with sampleRest scripts
```

### Issue 3: Management Network UUID Invalid

**Error**: `Management network UUID not found`

**Solution**: Get valid management network UUID from BCM:
```bash
cd sampleRest
python3 02-list-networks.py  # Find management network UUID
```

### Issue 4: Test Timeout

**Error**: `test timed out after 120m`

**Solution**:
- Check BCM cluster reachability
- Verify no firewall blocking
- Increase timeout: `-timeout 180m`

### Issue 5: Import State Mismatch

**Error**: `Terraform will perform the following actions after import`

**Solution**:
- Verify all computed fields set correctly in `readCategory`
- Check force parameter in `ImportStateVerifyIgnore`

---

## Performance Benchmarks

Expected performance for typical category operations:

| Operation | Expected Time | Notes |
|-----------|---------------|-------|
| Create (minimal) | 2-5 seconds | Includes API call + readback |
| Create (complete) | 3-8 seconds | More fields = longer processing |
| Read | < 1 second | Direct getCategory(name) lookup |
| Update | 2-5 seconds | Similar to create |
| Delete | 1-3 seconds | Quick if no nodes assigned |
| Import | 2-4 seconds | Two-phase: list + read |

---

## Next Steps After Implementation

1. **Create Pull Request**:
   ```bash
   git checkout -b 001-cmdevice-category
   git add internal/provider/resource_cmdevice_category*
   git add internal/provider/provider.go
   git add examples/resources/bcm_cmdevice_category/
   git commit -m "feat: Add bcm_cmdevice_category resource"
   git push origin 001-cmdevice-category
   ```

2. **Run CI Pipeline**: Verify all tests pass in CI environment

3. **Update CHANGELOG.md**: Document new resource

4. **Update README.md**: Add category resource to feature list

5. **Consider Data Source**: Implement `data_source_cmdevice_categories.go` for querying

---

## Reference Files

Key files to reference during implementation:

- **Pattern Reference**: `/workspace/internal/provider/resource_cmpart_softwareimage.go`
- **API Client**: `/workspace/internal/provider/bcm_client.go`
- **Helper Functions**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
- **Data Model**: `/workspace/specs/001-cmdevice-category/data-model.md`
- **API Contract**: `/workspace/specs/001-cmdevice-category/contracts/cmdevice-category-api.md`
- **Research**: `/workspace/specs/001-cmdevice-category/research.md`

---

## Support & Resources

- **BCM API Documentation**: `sampleRest/CMDevice_Complete_Documentation.md`
- **Terraform Plugin Framework**: https://developer.hashicorp.com/terraform/plugin/framework
- **Provider Best Practices**: `AGENTS.md` and `.specify/memory/constitution.md`
- **Testing Guide**: Terraform Plugin Testing documentation

---

**Happy Coding! Follow TDD and make tests your documentation.**
