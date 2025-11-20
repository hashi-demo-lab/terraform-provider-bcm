# Implementation Plan: bcm_cmpart_softwareimage Resource

## Executive Summary

This plan implements a Terraform resource for managing Nvidia BCM (Bright Cluster Manager) software images through the CMPart service. The implementation follows Test-Driven Development (TDD) principles with RED-GREEN-REFACTOR cycles, consistent with existing provider patterns.

## Feature Context

**Resource Name:** `bcm_cmpart_softwareimage`
**API Service:** `cmpart` (or `CMPart`)
**API Methods:**
- **Create:** `addSoftwareImage(softwareImage, force)`
- **Read:** `getSoftwareImage(name)` (direct lookup by name - efficient single-fetch)
- **Update:** `updateSoftwareImage(entity, force)` ✅ Confirmed in Phase 0
- **Delete:** `removeSoftwareImage(uuid, removeData, removeAll, force)` ✅ Confirmed in Phase 0

**Resource Type:** Manages BCM software images (OS images for provisioning DPU nodes)

## Technical Context

### Known Information

1. **Existing Infrastructure:**
   - BCM API client with JSON-RPC support (`internal/provider/bcm_client.go`)
   - Cookie-based authentication (cm-login-token)
   - Helper functions for null-safe field extraction (getStringValue, getBoolValue, getInt64Value)
   - Existing data source implementation: `data_source_cmpart_softwareimages.go`

2. **SoftwareImage Entity Schema:**
   - **Identity:** uuid (computed), name (required, unique), path (required, unique)
   - **Kernel Config:** kernelVersion, kernelParameters, kernelOutputConsole (default "tty0")
   - **Serial Over LAN:** enableSOL (default false), SOLPort (default "ttyS1"), SOLSpeed (default "115200"), SOLFlowControl (default true)
   - **Partitions:** fspart (UUID ref), bootfspart (UUID ref)
   - **Nested Resources:** modules (List of KernelModule with name and parameters)
   - **Metadata:** creationTime (computed), revisionID (computed), baseType (computed)
   - **Relationships:** originalImage (UUID), parentSoftwareImage (UUID)
   - **State Flags:** fileOperationInProgress (computed), modified (computed), to_be_removed (computed)
   - **Notes:** notes (optional string)

3. **API Constraints (POV Scope):**
   - Args parameter support needs verification for getSoftwareImage(name)
   - Fallback to direct HTTP POST if CallJSONRPC doesn't support args
   - All CRUD operations use cookie-authenticated JSON-RPC

4. **Test Requirements:**
   - Environment variables: TF_ACC=1, BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD
   - Full CRUD coverage with acceptance tests
   - ImportState functionality required
   - Unique resource names per test (avoid conflicts)

### Unknowns (NEEDS CLARIFICATION)

1. **Update API Method:**
   - Method name: `updateSoftwareImage(uuid, entity)` or entity modification pattern?
   - Does update require full entity or partial fields?
   - Are certain fields immutable (path, name uniqueness)?

2. **Delete API Method:**
   - Method name: `removeSoftwareImage(uuid)` or `removeSoftwareImage(name)`?
   - Parameter format: UUID or name?
   - Force delete option required?

3. **Read Operation:** ✅ RESOLVED
   - **YES:** `getSoftwareImage(name)` single-fetch method discovered in Phase 0
   - Direct lookup by name (not UUID) - much more efficient than list+filter
   - Resource deletion detected via 404 error or null response

4. **Nested Modules Management:**
   - Are modules managed inline with software image CRUD?
   - Do modules require separate API calls?
   - Module UUID generation: client-side or server-side?

5. **Unique Constraints:**
   - How does API enforce name uniqueness?
   - How does API enforce path uniqueness?
   - What error response format for constraint violations?

6. **Computed Fields:**
   - Which fields are read-only vs configurable?
   - Is `uuid` server-generated on create?
   - Is `creationTime` server-set?

7. **Force Parameter in addSoftwareImage:**
   - What does `force` parameter do?
   - When should it be true vs false?
   - Default value recommendation?

## Constitution Check

### TDD Principles Compliance

✅ **Test-First Development:** All acceptance tests written before implementation
✅ **RED-GREEN-REFACTOR Cycles:** Three-phase implementation per resource
✅ **Minimal Implementation:** GREEN phase uses hardcoded values
✅ **Comprehensive Coverage:** CRUD operations + ImportState tested
✅ **Parallel Execution:** Independent operations run concurrently

### Architecture Principles

✅ **Terraform Plugin Framework:** Using framework v1.16.1 with schema.Schema
✅ **Resource Interfaces:** Implements resource.Resource and resource.ResourceWithImportState
✅ **State Management:** Uses terraform-plugin-framework types (types.String, types.Int64, etc.)
✅ **Error Handling:** Uses Diagnostics API for user-friendly errors
✅ **API Client Pattern:** Reuses existing BCMClient with CallJSONRPC

### Known Risks

⚠️ **API Method Unknowns:** Update and Delete methods need research/discovery
⚠️ **Nested Resource Complexity:** KernelModule list may complicate state management
⚠️ **Unique Constraint Testing:** Must verify API error responses for duplicate name/path
⚠️ **Force Parameter Semantics:** Unclear when to use force=true in addSoftwareImage

**Risk Mitigation:**
- Phase 0 research resolves all API method unknowns
- Use existing data source as reference for nested modules pattern
- Acceptance tests include negative cases for constraint violations

## Phase 0: Research & Planning

### Goals
1. Discover Update and Delete API methods
2. Understand CRUD operation semantics
3. Validate entity structure and constraints
4. Document API error response patterns

### Research Tasks

#### Task 0.1: API Method Discovery
**Objective:** Find Update and Delete method names and signatures

**Approach:**
1. Search BCM API documentation in `/workspace/sampleRest/`
2. Check JavaScript bundle analysis for CMPart methods
3. Test API calls against live BCM endpoint:
   ```json
   {"service": "CMPart", "call": "updateSoftwareImage", "args": [...]}
   {"service": "CMPart", "call": "removeSoftwareImage", "args": [...]}
   {"service": "CMPart", "call": "modifySoftwareImage", "args": [...]}
   ```
4. Document actual method names and parameter formats

**Expected Output:** Confirmed API method names and call patterns

#### Task 0.2: CRUD Operation Semantics
**Objective:** Understand full lifecycle of software image management

**Approach:**
1. Create test software image via API
2. Read it back using getSoftwareImages
3. Modify entity fields and update
4. Delete software image
5. Document success/error responses at each step

**Expected Output:** research.md with CRUD flow documentation

#### Task 0.3: Constraint Validation Testing
**Objective:** Understand API enforcement of unique constraints

**Approach:**
1. Create software image with name "test-image-001"
2. Attempt to create another with same name → expect error
3. Attempt to create with same path → expect error
4. Document error response structure

**Expected Output:** Error handling patterns documented

#### Task 0.4: Nested Modules Management
**Objective:** Clarify KernelModule lifecycle

**Approach:**
1. Create software image with modules list
2. Update modules (add/remove/modify)
3. Verify modules are managed inline (not separate API calls)
4. Test module UUID behavior (client vs server generated)

**Expected Output:** Module management patterns for implementation

### Research Deliverable: research.md

**Location:** `/workspace/specs/003-bcm-cmpart-softwareimage-resource/research.md`

**Required Sections:**
1. **API Method Reference**
   - Create: `addSoftwareImage(entity, force)` confirmed
   - Read: `getSoftwareImages()` or `getSoftwareImage(uuid)`
   - Update: [Method name TBD]
   - Delete: [Method name TBD]

2. **Entity Lifecycle**
   - Create flow with example request/response
   - Read flow with filtering strategy
   - Update flow with entity modification
   - Delete flow with cleanup verification

3. **Constraint Enforcement**
   - Unique name validation
   - Unique path validation
   - Error response formats

4. **Nested Resource Patterns**
   - KernelModule inline management
   - UUID generation strategy
   - Empty modules list handling

5. **Design Decisions**
   - Force parameter: default false, expose as optional attribute
   - Read strategy: use getSoftwareImage(name) for efficient direct lookup
   - ImportState: use UUID as import identifier
   - Modules: manage as nested list_nested_attribute
   - Validation: use validateSoftwareImage() in REFACTOR phase for Create/Update

## Phase 1: Design & Contracts

### Goals
1. Define Terraform resource schema
2. Create example configurations
3. Document resource attributes
4. Plan test scenarios

### Task 1.1: Resource Schema Definition

**File:** `/workspace/specs/003-bcm-cmpart-softwareimage-resource/schema.md`

**Schema Structure:**

```hcl
resource "bcm_cmpart_softwareimage" "example" {
  # Required attributes
  name = "ubuntu-22.04-dpu"
  path = "/cm/images/ubuntu-22.04-dpu"

  # Kernel configuration
  kernel_version         = "5.15.0-58-generic"
  kernel_parameters      = "rd.driver.blacklist=nouveau"
  kernel_output_console  = "tty0"  # Default

  # Serial Over LAN (optional)
  enable_sol       = false  # Default
  sol_port         = "ttyS1"  # Default
  sol_speed        = "115200"  # Default: "115200"|"57600"|"38400"|"19200"|"9600"|"4800"|"2400"|"1200"
  sol_flow_control = true  # Default

  # Kernel modules (optional)
  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "e1000e"
      parameters = ""
    }
  ]

  # Partitions (optional, typically managed separately)
  fspart      = "uuid-of-filesystem-partition"
  bootfspart  = "uuid-of-boot-partition"

  # Metadata (optional)
  notes = "Ubuntu 22.04 LTS for DPU nodes with NVIDIA drivers"

  # Force flag (optional)
  force = false  # Default

  # Computed attributes (read-only)
  # id, uuid, creation_time, revision_id, file_operation_in_progress
}
```

**Attribute Mapping:**

| Terraform Attribute | API Field | Type | Required | Default | Notes |
|---------------------|-----------|------|----------|---------|-------|
| `id` | `uuid` | string | computed | - | Unique identifier |
| `uuid` | `uuid` | string | computed | - | Server-generated UUID |
| `name` | `name` | string | required | - | Unique name constraint |
| `path` | `path` | string | required | - | Unique path constraint, format: `^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$` |
| `kernel_version` | `kernelVersion` | string | optional | "" | Kernel version string |
| `kernel_parameters` | `kernelParameters` | string | optional | "" | Boot parameters |
| `kernel_output_console` | `kernelOutputConsole` | string | optional | "tty0" | Output console |
| `enable_sol` | `enableSOL` | bool | optional | false | Serial Over LAN |
| `sol_port` | `SOLPort` | string | optional | "ttyS1" | Serial port |
| `sol_speed` | `SOLSpeed` | string | optional | "115200" | Baud rate enum |
| `sol_flow_control` | `SOLFlowControl` | bool | optional | true | Flow control |
| `notes` | `notes` | string | optional | "" | Free-form notes |
| `fspart` | `fspart` | string | optional | null | Filesystem partition UUID |
| `bootfspart` | `bootfspart` | string | optional | null | Boot partition UUID |
| `modules` | `modules` | list(object) | optional | [] | Nested kernel modules |
| `force` | force param | bool | optional | false | Force create flag |
| `creation_time` | `creationTime` | int64 | computed | - | Unix timestamp |
| `revision_id` | `revisionID` | int64 | computed | - | Revision number |
| `file_operation_in_progress` | `fileOperationInProgress` | bool | computed | - | Operation status |
| `original_image` | `originalImage` | string | computed | - | Source image UUID |
| `parent_software_image` | `parentSoftwareImage` | string | computed | - | Parent UUID |

### Task 1.2: Example Configurations

**File:** `/workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf`

**Basic Example:**
```hcl
resource "bcm_cmpart_softwareimage" "ubuntu" {
  name = "ubuntu-22.04-basic"
  path = "/cm/images/ubuntu-22.04-basic"

  kernel_version    = "5.15.0-58-generic"
  kernel_parameters = "quiet splash"
}
```

**Advanced Example:**
```hcl
resource "bcm_cmpart_softwareimage" "dpu_optimized" {
  name = "dpu-ubuntu-nvidia"
  path = "/cm/images/dpu-ubuntu-nvidia"

  kernel_version    = "5.15.0-58-generic"
  kernel_parameters = "rd.driver.blacklist=nouveau console=tty0 console=ttyS1,115200"

  enable_sol       = true
  sol_port         = "ttyS1"
  sol_speed        = "115200"
  sol_flow_control = true

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "nvidia-uvm"
      parameters = ""
    },
    {
      name       = "mlx5_core"
      parameters = ""
    }
  ]

  notes = "Optimized image for NVIDIA BlueField DPUs with NVIDIA drivers and Mellanox networking"
}
```

**Import Example:**
```hcl
# Import existing software image by UUID
terraform import bcm_cmpart_softwareimage.existing <uuid>
```

### Task 1.3: Test Scenario Planning

**Test Coverage Matrix:**

| Test Scenario | CRUD Operation | Attributes Tested | Expected Behavior |
|---------------|----------------|-------------------|-------------------|
| Basic Create | Create, Read | name, path | Resource created with minimal config |
| Full Create | Create, Read | All optional attrs | Resource created with full config |
| Update Name | Update, Read | name | Name update applied (if mutable) |
| Update Kernel | Update, Read | kernel_version, kernel_parameters | Kernel config updated |
| Update Modules | Update, Read | modules list | Modules added/removed |
| Delete | Delete | - | Resource removed from BCM |
| Import | ImportState | id | Existing resource imported by UUID |
| Duplicate Name | Create | name | Error: name already exists |
| Invalid Path | Create | path | Error: path format invalid |
| Missing Required | Create | name, path | Error: required field missing |

## Phase 2: Implementation (TDD Cycles)

### RED Phase: Failing Acceptance Tests

**File:** `/workspace/internal/provider/resource_cmpart_softwareimage_test.go`

**Test Structure:**

```go
package provider

import (
    "fmt"
    "os"
    "testing"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// TestAccCMPartSoftwareImageResource_Basic tests minimal resource creation
func TestAccCMPartSoftwareImageResource_Basic(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccCMPartSoftwareImageResourceConfig_Basic("test-image-basic"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-image-basic"),
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "path", "/cm/images/test-image-basic"),
                    resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "id"),
                    resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
                    resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "creation_time"),
                ),
            },
            // ImportState testing
            {
                ResourceName:      "bcm_cmpart_softwareimage.test",
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{"force"}, // Force is create-time only
            },
            // Update and Read testing
            {
                Config: testAccCMPartSoftwareImageResourceConfig_BasicUpdated("test-image-basic"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_version", "5.15.0-60-generic"),
                ),
            },
            // Delete testing automatically occurs in TestCase
        },
    })
}

// TestAccCMPartSoftwareImageResource_FullConfig tests resource with all attributes
func TestAccCMPartSoftwareImageResource_FullConfig(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccCMPartSoftwareImageResourceConfig_Full("test-image-full"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-image-full"),
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "enable_sol", "true"),
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "sol_speed", "115200"),
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.#", "2"),
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.0.name", "nvidia-drm"),
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "notes", "Test image with full config"),
                ),
            },
        },
    })
}

// TestAccCMPartSoftwareImageResource_UpdateModules tests module list updates
func TestAccCMPartSoftwareImageResource_UpdateModules(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create with 1 module
            {
                Config: testAccCMPartSoftwareImageResourceConfig_Modules("test-image-modules", 1),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.#", "1"),
                ),
            },
            // Update to 3 modules
            {
                Config: testAccCMPartSoftwareImageResourceConfig_Modules("test-image-modules", 3),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.#", "3"),
                ),
            },
        },
    })
}

// Helper functions for test configurations

func testAccCMPartSoftwareImageResourceConfig_Basic(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = "/cm/images/%[4]s"
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}

func testAccCMPartSoftwareImageResourceConfig_BasicUpdated(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name           = %[4]q
  path           = "/cm/images/%[4]s"
  kernel_version = "5.15.0-60-generic"
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}

func testAccCMPartSoftwareImageResourceConfig_Full(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name                  = %[4]q
  path                  = "/cm/images/%[4]s"
  kernel_version        = "5.15.0-58-generic"
  kernel_parameters     = "quiet splash"
  kernel_output_console = "tty0"

  enable_sol       = true
  sol_port         = "ttyS1"
  sol_speed        = "115200"
  sol_flow_control = true

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "e1000e"
      parameters = ""
    }
  ]

  notes = "Test image with full config"
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}

func testAccCMPartSoftwareImageResourceConfig_Modules(name string, moduleCount int) string {
    modulesHCL := ""
    if moduleCount > 0 {
        modulesHCL = "modules = [\n"
        moduleNames := []string{"nvidia-drm", "e1000e", "mlx5_core"}
        for i := 0; i < moduleCount && i < len(moduleNames); i++ {
            modulesHCL += fmt.Sprintf(`    {
      name       = "%s"
      parameters = ""
    },
`, moduleNames[i])
        }
        modulesHCL += "  ]"
    }

    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = "/cm/images/%[4]s"

  %[5]s
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        modulesHCL,
    )
}
```

**RED Phase Verification:**
```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource
# Expected: All tests FAIL (resource not implemented)
```

### GREEN Phase: Minimal Implementation

**File:** `/workspace/internal/provider/resource_cmpart_softwareimage.go`

**Minimal Implementation (Hardcoded Success):**

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
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
    "github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
    "github.com/hashicorp/terraform-plugin-framework/types"
    "github.com/hashicorp/terraform-plugin-log/tflog"
)

// Ensure implementation satisfies expected interfaces
var (
    _ resource.Resource                = &CMPartSoftwareImageResource{}
    _ resource.ResourceWithImportState = &CMPartSoftwareImageResource{}
)

// NewCMPartSoftwareImageResource creates the resource
func NewCMPartSoftwareImageResource() resource.Resource {
    return &CMPartSoftwareImageResource{}
}

// CMPartSoftwareImageResource manages software images
type CMPartSoftwareImageResource struct {
    client *BCMClient
}

// CMPartSoftwareImageResourceModel maps schema to Go struct
type CMPartSoftwareImageResourceModel struct {
    // Identity
    ID   types.String `tfsdk:"id"`
    UUID types.String `tfsdk:"uuid"`
    Name types.String `tfsdk:"name"`
    Path types.String `tfsdk:"path"`

    // Kernel configuration
    KernelVersion       types.String `tfsdk:"kernel_version"`
    KernelParameters    types.String `tfsdk:"kernel_parameters"`
    KernelOutputConsole types.String `tfsdk:"kernel_output_console"`

    // Serial Over LAN
    EnableSOL      types.Bool   `tfsdk:"enable_sol"`
    SOLPort        types.String `tfsdk:"sol_port"`
    SOLSpeed       types.String `tfsdk:"sol_speed"`
    SOLFlowControl types.Bool   `tfsdk:"sol_flow_control"`

    // Partitions
    FSPart     types.String `tfsdk:"fspart"`
    BootfsPart types.String `tfsdk:"bootfspart"`

    // Metadata
    Notes        types.String `tfsdk:"notes"`
    Force        types.Bool   `tfsdk:"force"`
    CreationTime types.Int64  `tfsdk:"creation_time"`
    RevisionID   types.Int64  `tfsdk:"revision_id"`

    // State flags (computed)
    FileOperationInProgress types.Bool   `tfsdk:"file_operation_in_progress"`
    OriginalImage           types.String `tfsdk:"original_image"`
    ParentSoftwareImage     types.String `tfsdk:"parent_software_image"`

    // Nested modules
    Modules []KernelModuleResourceModel `tfsdk:"modules"`
}

// KernelModuleResourceModel represents a kernel module
type KernelModuleResourceModel struct {
    Name       types.String `tfsdk:"name"`
    Parameters types.String `tfsdk:"parameters"`
}

// Metadata returns resource type name
func (r *CMPartSoftwareImageResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
    resp.TypeName = req.ProviderTypeName + "_cmpart_softwareimage"
}

// Schema defines the resource schema
func (r *CMPartSoftwareImageResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages BCM software images for node provisioning. Software images are OS images with kernel configuration and modules.",

        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Software image identifier (same as uuid)",
            },
            "uuid": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Software image UUID",
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Software image name (must be unique)",
            },
            "path": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "File system path to image (must be unique, format: /cm/images/name)",
            },
            "kernel_version": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                MarkdownDescription: "Linux kernel version (e.g., 5.15.0-58-generic)",
            },
            "kernel_parameters": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                MarkdownDescription: "Kernel boot parameters",
            },
            "kernel_output_console": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                Default:             stringdefault.StaticString("tty0"),
                MarkdownDescription: "Kernel output console. Default: tty0",
            },
            "enable_sol": schema.BoolAttribute{
                Optional:            true,
                Computed:            true,
                Default:             booldefault.StaticBool(false),
                MarkdownDescription: "Enable Serial Over LAN. Default: false",
            },
            "sol_port": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                Default:             stringdefault.StaticString("ttyS1"),
                MarkdownDescription: "SOL serial port. Default: ttyS1",
            },
            "sol_speed": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                Default:             stringdefault.StaticString("115200"),
                MarkdownDescription: "SOL baud rate (115200|57600|38400|19200|9600|4800|2400|1200). Default: 115200",
            },
            "sol_flow_control": schema.BoolAttribute{
                Optional:            true,
                Computed:            true,
                Default:             booldefault.StaticBool(true),
                MarkdownDescription: "SOL hardware flow control. Default: true",
            },
            "fspart": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                MarkdownDescription: "Root filesystem partition UUID",
            },
            "bootfspart": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                MarkdownDescription: "Boot filesystem partition UUID",
            },
            "notes": schema.StringAttribute{
                Optional:            true,
                Computed:            true,
                MarkdownDescription: "Free-form notes about the image",
            },
            "force": schema.BoolAttribute{
                Optional:            true,
                Computed:            true,
                Default:             booldefault.StaticBool(false),
                MarkdownDescription: "Force image creation. Default: false",
            },
            "creation_time": schema.Int64Attribute{
                Computed:            true,
                MarkdownDescription: "Unix timestamp when image was created",
            },
            "revision_id": schema.Int64Attribute{
                Computed:            true,
                MarkdownDescription: "Revision number",
            },
            "file_operation_in_progress": schema.BoolAttribute{
                Computed:            true,
                MarkdownDescription: "File operation currently executing",
            },
            "original_image": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Source image UUID (for clones)",
            },
            "parent_software_image": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Parent image UUID (for derived images)",
            },
            "modules": schema.ListNestedAttribute{
                Optional:            true,
                Computed:            true,
                MarkdownDescription: "Kernel modules to load",
                NestedObject: schema.NestedAttributeObject{
                    Attributes: map[string]schema.Attribute{
                        "name": schema.StringAttribute{
                            Required:            true,
                            MarkdownDescription: "Module name (e.g., nvidia-drm)",
                        },
                        "parameters": schema.StringAttribute{
                            Optional:            true,
                            Computed:            true,
                            MarkdownDescription: "Module load parameters",
                        },
                    },
                },
            },
        },
    }
}

// Configure sets the BCM client
func (r *CMPartSoftwareImageResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create provisions a new software image
func (r *CMPartSoftwareImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL IMPLEMENTATION: Hardcoded success response
    // GREEN phase goal: make tests pass quickly
    plan.ID = types.StringValue("hardcoded-uuid-12345")
    plan.UUID = types.StringValue("hardcoded-uuid-12345")
    plan.CreationTime = types.Int64Value(1700000000)
    plan.RevisionID = types.Int64Value(0)
    plan.FileOperationInProgress = types.BoolValue(false)

    tflog.Trace(ctx, "Created software image (minimal implementation)")
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Read fetches software image state
func (r *CMPartSoftwareImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL IMPLEMENTATION: State already has data, no-op
    tflog.Trace(ctx, "Read software image (minimal implementation)")
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// Update modifies software image
func (r *CMPartSoftwareImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL IMPLEMENTATION: Accept plan as-is
    tflog.Trace(ctx, "Updated software image (minimal implementation)")
    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}

// Delete removes software image
func (r *CMPartSoftwareImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // MINIMAL IMPLEMENTATION: No-op delete
    tflog.Trace(ctx, "Deleted software image (minimal implementation)")
}

// ImportState imports resource by UUID
func (r *CMPartSoftwareImageResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

**Register Resource in Provider:**

Edit `/workspace/internal/provider/provider.go`:

```go
// In Resources() method, add:
func (p *BCMProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewCMPartSoftwareImageResource, // Add this line
    }
}
```

**GREEN Phase Verification:**
```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic
# Expected: Tests PASS (minimal hardcoded implementation)
```

### REFACTOR Phase: Full API Integration

**Refactor Create Method:**

```go
// Create provisions a new software image
func (r *CMPartSoftwareImageResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var plan CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build API entity from plan
    entity := map[string]interface{}{
        "baseType":            "SoftwareImage",
        "childType":           "",
        "name":                plan.Name.ValueString(),
        "path":                plan.Path.ValueString(),
        "kernelVersion":       plan.KernelVersion.ValueString(),
        "kernelParameters":    plan.KernelParameters.ValueString(),
        "kernelOutputConsole": plan.KernelOutputConsole.ValueString(),
        "enableSOL":           plan.EnableSOL.ValueBool(),
        "SOLPort":             plan.SOLPort.ValueString(),
        "SOLSpeed":            plan.SOLSpeed.ValueString(),
        "SOLFlowControl":      plan.SOLFlowControl.ValueBool(),
        "notes":               plan.Notes.ValueString(),
    }

    // Add optional partition UUIDs if set
    if !plan.FSPart.IsNull() {
        entity["fspart"] = plan.FSPart.ValueString()
    }
    if !plan.BootfsPart.IsNull() {
        entity["bootfspart"] = plan.BootfsPart.ValueString()
    }

    // Build modules list
    modules := []map[string]interface{}{}
    for _, mod := range plan.Modules {
        modules = append(modules, map[string]interface{}{
            "baseType":     "KernelModule",
            "childType":    "",
            "name":         mod.Name.ValueString(),
            "parameters":   mod.Parameters.ValueString(),
            "modified":     true,
            "to_be_removed": false,
        })
    }
    entity["modules"] = modules

    // STEP 1: Validate entity before creating ✅ ADDED
    validateReqBody := map[string]interface{}{
        "service": "CMPart",
        "call":    "validateSoftwareImage",
        "args":    []interface{}{entity},
    }

    validateJSON, err := json.Marshal(validateReqBody)
    if err != nil {
        resp.Diagnostics.AddError("Validation Request Marshal Error", err.Error())
        return
    }

    validateHTTPReq, err := http.NewRequestWithContext(ctx, "POST", r.client.Endpoint+"/json", bytes.NewReader(validateJSON))
    if err != nil {
        resp.Diagnostics.AddError("Validation HTTP Request Error", err.Error())
        return
    }
    validateHTTPReq.Header.Set("Content-Type", "application/json")

    validateHTTPResp, err := r.client.HTTPClient.Do(validateHTTPReq)
    if err != nil {
        resp.Diagnostics.AddError("Validation API Call Failed", err.Error())
        return
    }
    defer validateHTTPResp.Body.Close()

    validateBody, _ := io.ReadAll(validateHTTPResp.Body)

    // Check validation response
    if validateHTTPResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Validation Failed",
            fmt.Sprintf("Software image configuration is invalid: %s", string(validateBody)))
        return
    }

    tflog.Debug(ctx, "Validation passed", map[string]interface{}{
        "name": plan.Name.ValueString(),
    })

    // STEP 2: Call API: addSoftwareImage(entity, force)
    force := plan.Force.ValueBool()
    reqBody := map[string]interface{}{
        "service": "CMPart",
        "call":    "addSoftwareImage",
        "args":    []interface{}{entity, force},
    }

    jsonBody, err := json.Marshal(reqBody)
    if err != nil {
        resp.Diagnostics.AddError("Request Marshal Error", err.Error())
        return
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", r.client.Endpoint+"/json", bytes.NewReader(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("HTTP Request Error", err.Error())
        return
    }
    httpReq.Header.Set("Content-Type", "application/json")

    httpResp, err := r.client.HTTPClient.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("API Call Failed", err.Error())
        return
    }
    defer httpResp.Body.Close()

    body, _ := io.ReadAll(httpResp.Body)

    // Parse response (expected: UUID string or entity object)
    var apiResponse interface{}
    if err := json.Unmarshal(body, &apiResponse); err != nil {
        resp.Diagnostics.AddError("Response Parse Error", err.Error())
        return
    }

    // Extract UUID from response
    var createdUUID string
    switch v := apiResponse.(type) {
    case string:
        createdUUID = v
    case map[string]interface{}:
        if uuid, ok := v["uuid"].(string); ok {
            createdUUID = uuid
        }
    }

    if createdUUID == "" {
        resp.Diagnostics.AddError("Create Failed", "No UUID returned from API")
        return
    }

    // Read back the created resource to populate computed fields
    plan.ID = types.StringValue(createdUUID)
    plan.UUID = types.StringValue(createdUUID)

    // Call Read to populate all fields
    r.readSoftwareImage(ctx, &plan, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    tflog.Debug(ctx, "Created software image", map[string]interface{}{
        "uuid": createdUUID,
        "name": plan.Name.ValueString(),
    })

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

**Refactor Read Method:**

```go
// Read fetches software image state
func (r *CMPartSoftwareImageResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
    var state CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    r.readSoftwareImage(ctx, &state, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}

// readSoftwareImage fetches image from API and updates model
func (r *CMPartSoftwareImageResource) readSoftwareImage(ctx context.Context, model *CMPartSoftwareImageResourceModel, diags *diag.Diagnostics) {
    // Call getSoftwareImage(name) API for efficient direct lookup
    // NOTE: Requires args parameter support in BCMClient (verify in Phase 1)
    imageName := model.Name.ValueString()
    body, err := r.client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName)
    if err != nil {
        diags.AddError("Unable to Read Software Image",
            fmt.Sprintf("Failed to read image '%s': %s", imageName, err.Error()))
        return
    }

    // Parse response as single entity (not array)
    var imageData map[string]interface{}
    if err := json.Unmarshal(body, &imageData); err != nil {
        diags.AddError("Parse Error", err.Error())
        return
    }

    // Check if image was found (null response or empty object indicates not found)
    if imageData == nil || len(imageData) == 0 {
        diags.AddError("Software Image Not Found",
            fmt.Sprintf("Image '%s' not found in BCM", imageName))
        return
    }

    // Map API response to model (reuse helper from data source)
    model.UUID = getStringValue(imageData, "uuid")
    model.ID = model.UUID
    model.Name = getStringValue(imageData, "name")
    model.Path = getStringValue(imageData, "path")
    model.KernelVersion = getStringValue(imageData, "kernelVersion")
    model.KernelParameters = getStringValue(imageData, "kernelParameters")
    model.KernelOutputConsole = getStringValue(imageData, "kernelOutputConsole")
    model.EnableSOL = getBoolValue(imageData, "enableSOL")
    model.SOLPort = getStringValue(imageData, "SOLPort")
    model.SOLSpeed = getStringValue(imageData, "SOLSpeed")
    model.SOLFlowControl = getBoolValue(imageData, "SOLFlowControl")
    model.FSPart = getStringValue(imageData, "fspart")
    model.BootfsPart = getStringValue(imageData, "bootfspart")
    model.Notes = getStringValue(imageData, "notes")
    model.CreationTime = getInt64Value(imageData, "creationTime")
    model.RevisionID = getInt64Value(imageData, "revisionID")
    model.FileOperationInProgress = getBoolValue(imageData, "fileOperationInProgress")
    model.OriginalImage = getStringValue(imageData, "originalImage")
    model.ParentSoftwareImage = getStringValue(imageData, "parentSoftwareImage")

    // Parse modules
    if modulesData, ok := imageData["modules"].([]interface{}); ok {
        model.Modules = make([]KernelModuleResourceModel, 0, len(modulesData))
        for _, modData := range modulesData {
            if modMap, ok := modData.(map[string]interface{}); ok {
                module := KernelModuleResourceModel{
                    Name:       getStringValue(modMap, "name"),
                    Parameters: getStringValue(modMap, "parameters"),
                }
                model.Modules = append(model.Modules, module)
            }
        }
    }
}
```

**Refactor Update Method:**

```go
// Update modifies software image
func (r *CMPartSoftwareImageResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
    var plan CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Build updated entity (based on research findings for update method)
    // NOTE: Update API method determined in Phase 0 research
    entity := map[string]interface{}{
        "uuid":                plan.UUID.ValueString(),
        "baseType":            "SoftwareImage",
        "childType":           "",
        "name":                plan.Name.ValueString(),
        "path":                plan.Path.ValueString(),
        "kernelVersion":       plan.KernelVersion.ValueString(),
        "kernelParameters":    plan.KernelParameters.ValueString(),
        "kernelOutputConsole": plan.KernelOutputConsole.ValueString(),
        "enableSOL":           plan.EnableSOL.ValueBool(),
        "SOLPort":             plan.SOLPort.ValueString(),
        "SOLSpeed":            plan.SOLSpeed.ValueString(),
        "SOLFlowControl":      plan.SOLFlowControl.ValueBool(),
        "notes":               plan.Notes.ValueString(),
        "modified":            true,
    }

    // Add partitions if set
    if !plan.FSPart.IsNull() {
        entity["fspart"] = plan.FSPart.ValueString()
    }
    if !plan.BootfsPart.IsNull() {
        entity["bootfspart"] = plan.BootfsPart.ValueString()
    }

    // Build modules list
    modules := []map[string]interface{}{}
    for _, mod := range plan.Modules {
        modules = append(modules, map[string]interface{}{
            "baseType":      "KernelModule",
            "childType":     "",
            "name":          mod.Name.ValueString(),
            "parameters":    mod.Parameters.ValueString(),
            "modified":      true,
            "to_be_removed": false,
        })
    }
    entity["modules"] = modules

    // STEP 1: Validate entity before updating ✅ ADDED
    validateReqBody := map[string]interface{}{
        "service": "CMPart",
        "call":    "validateSoftwareImage",
        "args":    []interface{}{entity},
    }

    validateJSON, err := json.Marshal(validateReqBody)
    if err != nil {
        resp.Diagnostics.AddError("Validation Request Marshal Error", err.Error())
        return
    }

    validateHTTPReq, err := http.NewRequestWithContext(ctx, "POST", r.client.Endpoint+"/json", bytes.NewReader(validateJSON))
    if err != nil {
        resp.Diagnostics.AddError("Validation HTTP Request Error", err.Error())
        return
    }
    validateHTTPReq.Header.Set("Content-Type", "application/json")

    validateHTTPResp, err := r.client.HTTPClient.Do(validateHTTPReq)
    if err != nil {
        resp.Diagnostics.AddError("Validation API Call Failed", err.Error())
        return
    }
    defer validateHTTPResp.Body.Close()

    validateBody, _ := io.ReadAll(validateHTTPResp.Body)

    // Check validation response
    if validateHTTPResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Validation Failed",
            fmt.Sprintf("Software image update configuration is invalid: %s", string(validateBody)))
        return
    }

    tflog.Debug(ctx, "Validation passed for update", map[string]interface{}{
        "uuid": plan.UUID.ValueString(),
    })

    // STEP 2: Call update API ✅ CONFIRMED: updateSoftwareImage(entity, force)
    reqBody := map[string]interface{}{
        "service": "CMPart",
        "call":    "updateSoftwareImage",
        "args":    []interface{}{entity, false}, // force=false (safest default)
    }

    jsonBody, err := json.Marshal(reqBody)
    if err != nil {
        resp.Diagnostics.AddError("Request Marshal Error", err.Error())
        return
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", r.client.Endpoint+"/json", bytes.NewReader(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("HTTP Request Error", err.Error())
        return
    }
    httpReq.Header.Set("Content-Type", "application/json")

    httpResp, err := r.client.HTTPClient.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("API Call Failed", err.Error())
        return
    }
    defer httpResp.Body.Close()

    // Read back to verify update
    r.readSoftwareImage(ctx, &plan, &resp.Diagnostics)
    if resp.Diagnostics.HasError() {
        return
    }

    tflog.Debug(ctx, "Updated software image", map[string]interface{}{
        "uuid": plan.UUID.ValueString(),
    })

    resp.Diagnostics.Append(resp.State.Set(ctx, &plan)...)
}
```

**Refactor Delete Method:**

```go
// Delete removes software image
func (r *CMPartSoftwareImageResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state CMPartSoftwareImageResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Call delete API ✅ CONFIRMED: removeSoftwareImage(uuid, removeData, removeAll, force)
    reqBody := map[string]interface{}{
        "service": "CMPart",
        "call":    "removeSoftwareImage",
        "args":    []interface{}{
            state.UUID.ValueString(),
            false, // removeData: don't delete filesystem data
            false, // removeAll: don't cascade delete
            false, // force: safest default
        },
    }

    jsonBody, err := json.Marshal(reqBody)
    if err != nil {
        resp.Diagnostics.AddError("Request Marshal Error", err.Error())
        return
    }

    httpReq, err := http.NewRequestWithContext(ctx, "POST", r.client.Endpoint+"/json", bytes.NewReader(jsonBody))
    if err != nil {
        resp.Diagnostics.AddError("HTTP Request Error", err.Error())
        return
    }
    httpReq.Header.Set("Content-Type", "application/json")

    httpResp, err := r.client.HTTPClient.Do(httpReq)
    if err != nil {
        resp.Diagnostics.AddError("API Call Failed", err.Error())
        return
    }
    defer httpResp.Body.Close()

    body, _ := io.ReadAll(httpResp.Body)

    // Verify success response
    if httpResp.StatusCode != http.StatusOK {
        resp.Diagnostics.AddError("Delete Failed",
            fmt.Sprintf("HTTP %d: %s", httpResp.StatusCode, string(body)))
        return
    }

    tflog.Debug(ctx, "Deleted software image", map[string]interface{}{
        "uuid": state.UUID.ValueString(),
    })
}
```

**REFACTOR Phase Verification:**
```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource
# Expected: All tests PASS with real API calls
```

## Phase 3: Documentation Generation

### Task 3.1: Generate Provider Documentation

```bash
cd /workspace
make generate
# OR
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
/workspace/.go/bin/tfplugindocs generate --provider-name bcm --tf-version 1.13.5
```

**Expected Output:**
- `/workspace/docs/resources/cmpart_softwareimage.md` - Generated resource documentation

### Task 3.2: Verify Examples

Ensure examples in `/workspace/examples/resources/bcm_cmpart_softwareimage/` are:
- Syntactically valid HCL
- Reference actual BCM endpoints/values
- Include comments explaining configuration

### Task 3.3: Update Provider README

If applicable, update `/workspace/README.md` with:
- New resource in feature matrix
- Example usage snippet
- Link to generated docs

## Task Breakdown (Dependency-Ordered)

### Phase 0: Research (2-4 hours)

1. **[P0-T1] API Method Discovery** - Search BCM docs and test API for update/delete methods
2. **[P0-T2] CRUD Semantics Testing** - Perform manual CRUD lifecycle against BCM API
3. **[P0-T3] Constraint Validation** - Test unique name/path enforcement
4. **[P0-T4] Nested Modules Testing** - Verify inline module management
5. **[P0-T5] Write research.md** - Document all findings with examples

### Phase 1: Design (1-2 hours)

6. **[P1-T1] Write schema.md** - Document Terraform resource schema
7. **[P1-T2] Create example configurations** - Write basic and advanced HCL examples
8. **[P1-T3] Write test plan** - Define test coverage matrix

### Phase 2: Implementation (4-6 hours)

#### RED Phase (1 hour)
9. **[P2-T1] Write acceptance test suite** - Create all test functions in `*_test.go`
10. **[P2-T2] Run RED verification** - Confirm all tests fail (TF_ACC=1)

#### GREEN Phase (1-2 hours)
11. **[P2-T3] Implement minimal resource** - Hardcoded CRUD operations
12. **[P2-T4] Register resource in provider** - Add to Resources() method
13. **[P2-T5] Run GREEN verification** - Confirm tests pass with minimal impl

#### REFACTOR Phase (2-3 hours)
14. **[P2-T6] Refactor Create with API** - Replace hardcoded with addSoftwareImage call + validation
15. **[P2-T7] Refactor Read with API** - Implement getSoftwareImage(name) direct lookup
16. **[P2-T8] Refactor Update with API** - Implement updateSoftwareImage + validation
17. **[P2-T9] Refactor Delete with API** - Implement removeSoftwareImage method
18. **[P2-T10] Refactor ImportState** - Test UUID-based import
19. **[P2-T11] Run REFACTOR verification** - Confirm all tests pass with real API

### Phase 3: Documentation (30 minutes)

20. **[P3-T1] Generate docs** - Run `make generate` or tfplugindocs
21. **[P3-T2] Review generated docs** - Verify accuracy and completeness
22. **[P3-T3] Update examples** - Ensure HCL examples are production-ready

### Phase 4: Quality Assurance (1 hour)

23. **[QA-T1] Run full test suite** - `make test` and `make testacc`
24. **[QA-T2] Run linter** - `make lint` or `golangci-lint run`
25. **[QA-T3] Format code** - `make fmt` or `gofmt -s -w -e .`
26. **[QA-T4] Pre-commit hooks** - `pre-commit run --all-files`
27. **[QA-T5] Manual smoke test** - Apply example configuration against BCM

## Success Criteria

### Functional Requirements

✅ Resource creates software images via addSoftwareImage API with validateSoftwareImage pre-flight check
✅ Resource reads software images via getSoftwareImage(name) direct lookup (efficient single-fetch)
✅ Resource updates software images via updateSoftwareImage with validateSoftwareImage pre-flight check
✅ Resource deletes software images via removeSoftwareImage(uuid, removeData, removeAll, force)
✅ ImportState functionality works with UUID identifier
✅ Nested modules list properly managed in state
✅ All computed fields populated from API responses
✅ Unique constraints (name, path) validated with clear error messages

### Testing Requirements

✅ All acceptance tests pass with TF_ACC=1
✅ Test coverage includes:
  - Basic resource creation
  - Full configuration with all attributes
  - Module list updates (add/remove)
  - ImportState verification
  - Update operations
  - Delete operations
✅ Tests use unique resource names to avoid conflicts
✅ Tests include provider configuration in HCL

### Code Quality Requirements

✅ Follows HashiCorp provider development best practices
✅ Uses terraform-plugin-framework patterns consistently
✅ Implements null-safe helper functions for field extraction
✅ Includes comprehensive error handling with Diagnostics API
✅ Passes golangci-lint with no errors
✅ Code formatted with gofmt
✅ Pre-commit hooks pass

### Documentation Requirements

✅ Resource documentation auto-generated in docs/
✅ Example configurations in examples/resources/
✅ Schema descriptions clear and accurate
✅ Attribute defaults documented
✅ Computed vs required attributes clearly marked

## Risks and Mitigations

| Risk | Impact | Probability | Mitigation |
|------|--------|-------------|------------|
| Update/Delete API methods unknown | High | Medium | Phase 0 research discovers methods before implementation |
| Nested modules complicate state management | Medium | Low | Reuse data source pattern for modules list |
| Unique constraints not enforced by API | Medium | Low | Document expected behavior, test negative cases |
| Force parameter semantics unclear | Low | Medium | Research default behavior, expose as optional attribute |
| BCMClient args parameter support unclear | Medium | High | Test args support in Phase 1, fallback to direct HTTP if needed |
| API response format differs from docs | Medium | Low | Comprehensive error handling with response logging |

## Next Steps

After completing this implementation:

1. **Integration Testing:** Test resource with actual BCM cluster
2. **Performance Optimization:** Benchmark large module lists
3. **Error Handling Enhancement:** Add retry logic for transient API failures
4. **Related Resources:** Implement partition management resources (fspart, bootfspart)
5. **Data Source Alignment:** Ensure resource and data source use same model structs
6. **User Feedback:** Gather feedback from BCM users on attribute naming/defaults

## References

- **Terraform Plugin Framework Docs:** https://developer.hashicorp.com/terraform/plugin/framework
- **BCM API Documentation:** `/workspace/sampleRest/BCM_API_Complete_Documentation.md`
- **SoftwareImage Entity Spec:** `/workspace/sampleRest/wip/resource_cmpart_softwareimage.md`
- **Existing Data Source:** `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
- **TDD Constitution:** `/workspace/.specify/memory/constitution.md`
- **Provider Development Guide:** `/workspace/AGENTS.md`
