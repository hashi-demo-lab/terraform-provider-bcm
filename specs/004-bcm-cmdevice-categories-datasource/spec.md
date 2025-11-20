# Feature Specification: BCM CMDevice Categories Data Source

## Overview

Implement a Terraform data source to query and retrieve node categories from the NVIDIA Bright Cluster Manager (BCM) CMDevice API. This data source enables users to reference existing categories for node provisioning workflows.

## User Story

**As a** Terraform user managing BCM infrastructure
**I want** to query existing node categories
**So that** I can reference category UUIDs in node provisioning and configuration workflows

## Background

The BCM CMDevice service provides category management functionality. Categories define configuration templates for nodes, including:
- Software image assignments via `softwareImageProxy`
- Disk partitioning schemes via `disksetup` XML
- Boot loader configuration
- Network settings
- Kernel parameters and modules
- Filesystem mounts and exports
- GPU settings
- Roles

This data source is essential for the k8s control plane workflow where users need to:
1. Clone a software image with kernel modules (already supported by `bcm_cmpart_softwareimage`)
2. Create/configure a category referencing the software image (requires category lookup)

## Requirements

### Functional Requirements

#### FR1: List All Categories
**Priority:** HIGH
**Description:** Data source must retrieve all categories from the BCM API
**API Call:** `cmdevice.getCategories()`
**Acceptance Criteria:**
- Returns array of category objects
- Includes all category attributes
- Handles empty results gracefully

#### FR2: Filter by Category Name
**Priority:** HIGH
**Description:** Users can filter categories by exact name match
**Rationale:** Most common use case is finding a category by name to reference its UUID
**Acceptance Criteria:**
- Optional `name` filter attribute
- Returns only matching categories
- Empty results when no match found

#### FR3: Essential Attributes
**Priority:** HIGH
**Description:** Data source must expose these core attributes per category:

**Identity:**
- `uuid` (string) - Unique identifier
- `name` (string) - Category name
- `base_type` (string) - Always "Category"
- `child_type` (string) - Category subtype (usually empty)

**Configuration:**
- `software_image_id` (string) - UUID of associated software image from `softwareImageProxy.parentSoftwareImage`
- `management_network_id` (string) - UUID of management network
- `disksetup` (string) - Disk partitioning XML configuration
- `boot_loader` (string) - Boot loader type (SYSLINUX, GRUB, etc.)
- `boot_loader_protocol` (string) - Boot protocol (HTTP, TFTP, etc.)
- `install_mode` (string) - Installation mode (AUTO, FULL, etc.)
- `kernel_version` (string) - Kernel version
- `kernel_parameters` (string) - Kernel boot parameters
- `kernel_output_console` (string) - Console output device

**Network:**
- `default_gateway` (string) - Default gateway IP
- `default_gateway_metric` (number) - Gateway metric
- `name_servers` (list of strings) - DNS servers
- `search_domains` (list of strings) - DNS search domains
- `time_servers` (list of strings) - NTP servers

**System:**
- `authentication_service` (string) - Authentication service type
- `io_scheduler` (string) - I/O scheduler
- `fips` (string) - FIPS mode setting
- `interactive_user` (string) - Interactive user mode

**Flags:**
- `allow_networking_restart` (bool) - Allow network restart
- `install_boot_record` (bool) - Install boot record
- `data_node` (bool) - Is data node
- `version_config_files` (bool) - Version config files
- `modified` (bool) - Has unsaved changes
- `to_be_removed` (bool) - Scheduled for removal

**Metadata:**
- `notes` (string) - Administrative notes
- `parent_uuid` (string) - Parent category UUID (for cloned categories)

#### FR4: Complex Nested Attributes
**Priority:** MEDIUM
**Description:** Support for nested configuration objects

**Modules (list of objects):**
- `name` (string) - Kernel module name
- `parameters` (string) - Module parameters

**Filesystem Mounts (list of objects):**
- `device` (string) - Device or NFS path
- `mountpoint` (string) - Mount point path
- `filesystem` (string) - Filesystem type
- `mountoptions` (string) - Mount options
- `rdma` (bool) - RDMA enabled

**Filesystem Exports (list of objects):**
- `path` (string) - Export path
- `network` (string) - Network UUID
- `allow_write` (bool) - Allow write access
- `root_squash` (bool) - Root squash enabled
- `async` (bool) - Async mode

**Roles (list of objects):**
- `uuid` (string) - Role UUID
- `name` (string) - Role name
- `child_type` (string) - Role type
- `add_services` (bool) - Add services flag

**Services (list of objects):**
- `uuid` (string) - Service UUID
- `name` (string) - Service name
- `enabled` (bool) - Service enabled

**GPU Settings (list of objects):**
- `device_id` (string) - GPU device ID
- `model` (string) - GPU model
- `compute_mode` (string) - Compute mode

**Static Routes (list of objects):**
- `destination` (string) - Destination network
- `gateway` (string) - Gateway IP
- `metric` (number) - Route metric

#### FR5: BMC Settings
**Priority:** LOW (Optional for datasource, more relevant for resource)
**Description:** Baseboard Management Controller settings

**Attributes:**
- `bmc_username` (string) - BMC username
- `bmc_password` (string, sensitive) - BMC password
- `bmc_privilege` (string) - BMC privilege level
- `firmware_manage_mode` (string) - Firmware management mode

#### FR6: Exclude Lists
**Priority:** LOW (Optional)
**Description:** Rsync exclude patterns for node provisioning

**Attributes:**
- `exclude_list_full` (string) - Full install exclude patterns
- `exclude_list_grab` (string) - Grab image exclude patterns
- `exclude_list_grabnew` (string) - Grab new image exclude patterns
- `exclude_list_sync` (string) - Sync exclude patterns
- `exclude_list_update` (string) - Update exclude patterns
- `exclude_list_manipulate_script` (string) - Manipulate script

### Non-Functional Requirements

#### NFR1: TDD Implementation
**Priority:** CRITICAL
**Description:** Must follow Test-Driven Development pattern
**Requirements:**
- RED: Write failing acceptance test first
- GREEN: Minimal implementation to pass tests
- REFACTOR: Improve code quality while keeping tests green

#### NFR2: API Efficiency
**Priority:** HIGH
**Description:** Use efficient API calls
**Implementation:**
- Use `cmdevice.getCategories()` to list all categories
- Client-side filtering by name (no API-level filter available)
- Single API call per data source invocation

#### NFR3: Documentation
**Priority:** HIGH
**Description:** Complete documentation generation
**Requirements:**
- Example configurations in `examples/data-sources/bcm_cmdevice_categories/`
- Auto-generated docs via `make generate`
- Clear descriptions for all attributes

#### NFR4: Error Handling
**Priority:** HIGH
**Description:** Graceful error handling
**Scenarios:**
- API connection failures
- Authentication failures
- Empty result sets (valid, not an error)
- Malformed API responses

## API Contract

### Endpoint
```
POST https://<bcm-endpoint>/json
Content-Type: application/json
```

### Authentication
Session-based cookie authentication (`cm-login-token`)

### List Categories Request
```json
{
  "service": "cmdevice",
  "call": "getCategories"
}
```

### Response Structure
```json
[
  {
    "baseType": "Category",
    "childType": "",
    "uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
    "name": "default",
    "modified": false,
    "to_be_removed": false,
    "parent_uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
    "notes": "",

    "softwareImageProxy": {
      "baseType": "SoftwareImageProxy",
      "uuid": "7abe08d4-4c18-4d66-9eff-fa1a1b87e84c",
      "parentSoftwareImage": "8482c4e9-383c-43de-873f-8c54ee77ee74",
      "modified": false,
      "to_be_removed": false
    },

    "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",

    "disksetup": "<?xml version=\"1.0\" encoding=\"UTF-8\"?>...",

    "bootLoader": "SYSLINUX",
    "bootLoaderProtocol": "HTTP",
    "bootLoaderFile": "",
    "installMode": "AUTO",

    "kernelVersion": "",
    "kernelParameters": "",
    "kernelOutputConsole": "",

    "modules": [],
    "roles": [],
    "fsexports": [],
    "fsmounts": [...],
    "gpuSettings": [],
    "services": [],
    "staticRoutes": [],

    "defaultGateway": "0.0.0.0",
    "defaultGatewayMetric": 0,
    "nameServers": [],
    "searchDomains": [],
    "timeServers": [],

    "authenticationService": "AUTO",
    "ioScheduler": "",
    "fips": "NO",
    "interactiveUser": "ALWAYS",

    "allowNetworkingRestart": false,
    "installBootRecord": false,
    "dataNode": false,
    "versionConfigFiles": false,

    "bmcSettings": {...},

    "excludeListFull": "...",
    "excludeListGrab": "...",
    "excludeListGrabnew": "...",
    "excludeListSync": "...",
    "excludeListUpdate": "..."
  }
]
```

## Terraform Schema

### Data Source Declaration
```hcl
data "bcm_cmdevice_categories" "all" {
  # Optional filter
  name = "default"
}
```

### Computed Attributes
```hcl
output "categories" {
  value = data.bcm_cmdevice_categories.all.categories
  # Returns list of category objects with all attributes
}

output "default_category_id" {
  value = data.bcm_cmdevice_categories.all.categories[0].uuid
  # Example: "0ae6d733-3015-4479-bfab-ce2d237a2809"
}

output "software_image_id" {
  value = data.bcm_cmdevice_categories.all.categories[0].software_image_id
  # Example: "8482c4e9-383c-43de-873f-8c54ee77ee74"
}
```

## Implementation Strategy

### Phase 1: TDD - RED (Failing Test)
1. Create `data_source_cmdevice_categories_test.go`
2. Write acceptance test for basic category retrieval
3. Write test for name filtering
4. Run tests - verify failures
5. Commit RED phase

### Phase 2: TDD - GREEN (Minimal Implementation)
1. Create `data_source_cmdevice_categories.go`
2. Define schema with essential attributes (FR3)
3. Implement `Read()` method:
   - Call `client.CallJSONRPC(ctx, "cmdevice", "getCategories")`
   - Parse JSON response
   - Apply client-side name filter if specified
   - Map to Terraform state using helper functions
4. Register data source in `provider.go`
5. Run tests - verify all pass
6. Commit GREEN phase

### Phase 3: TDD - REFACTOR (Enhancement)
1. Add nested attributes (modules, fsmounts, etc.) - FR4
2. Add error handling and validation
3. Add comprehensive logging
4. Optimize data mapping
5. Run tests - verify still pass
6. Commit REFACTOR phase

### Phase 4: Documentation
1. Create `examples/data-sources/bcm_cmdevice_categories/data-source.tf`
2. Create advanced example with filtering
3. Run `make generate` to create docs
4. Verify documentation accuracy

## Usage Examples

### Example 1: List All Categories
```hcl
data "bcm_cmdevice_categories" "all" {}

output "all_categories" {
  value = {
    for cat in data.bcm_cmdevice_categories.all.categories :
    cat.name => cat.uuid
  }
}
```

### Example 2: Find Specific Category
```hcl
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

output "default_category_uuid" {
  value = data.bcm_cmdevice_categories.default.categories[0].uuid
}

output "default_software_image" {
  value = data.bcm_cmdevice_categories.default.categories[0].software_image_id
}
```

### Example 3: Integration with k8s Workflow
```hcl
# Step 1: Create software image with kernel modules
resource "bcm_cmpart_softwareimage" "k8s_control_plane" {
  name = "k8s-control-plane-image"
  original_image = data.bcm_cmpart_softwareimages.all.images[
    index(data.bcm_cmpart_softwareimages.all.images.*.name, "default-image")
  ].id

  path = "/cm/images/k8s-control-plane-image"

  modules = [
    {
      name       = "mlx5_core"
      parameters = ""
    },
    {
      name       = "bonding"
      parameters = ""
    }
  ]
}

# Step 2: Find default category for cloning
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

# Step 3: Create category (future resource implementation)
# resource "bcm_cmdevice_category" "k8s_control_plane" {
#   name                = "k8s-control-plane"
#   original_category   = data.bcm_cmdevice_categories.default.categories[0].uuid
#   software_image_id   = bcm_cmpart_softwareimage.k8s_control_plane.id
#   disksetup           = "/cm/local/apps/cmd/etc/htdocs/disk-setup/x86_64-slave-one-big-partition-ext4.xml"
# }
```

### Example 4: Category Analysis
```hcl
data "bcm_cmdevice_categories" "all" {}

output "category_analysis" {
  value = {
    for cat in data.bcm_cmdevice_categories.all.categories : cat.name => {
      uuid                = cat.uuid
      software_image_id   = cat.software_image_id
      management_network  = cat.management_network_id
      boot_loader         = cat.boot_loader
      install_mode        = cat.install_mode
      kernel_modules      = length(cat.modules)
      filesystem_mounts   = length(cat.fsmounts)
      roles               = length(cat.roles)
    }
  }
}
```

## Testing Strategy

### Acceptance Test Cases

#### Test 1: Basic Category Retrieval
```go
func TestAccCMDeviceCategoriesDataSource_Basic(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:                 func() { testAccPreCheck(t) },
    ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
    Steps: []resource.TestStep{
      {
        Config: testAccCMDeviceCategoriesDataSourceConfig(),
        Check: resource.ComposeAggregateTestCheckFunc(
          // Verify at least one category exists
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.#"),
          // Verify first category has required attributes
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.uuid"),
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.name"),
          resource.TestCheckResourceAttr("data.bcm_cmdevice_categories.test", "categories.0.base_type", "Category"),
        ),
      },
    },
  })
}
```

#### Test 2: Filter by Name
```go
func TestAccCMDeviceCategoriesDataSource_FilterByName(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:                 func() { testAccPreCheck(t) },
    ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
    Steps: []resource.TestStep{
      {
        Config: testAccCMDeviceCategoriesDataSourceConfigFilterByName("default"),
        Check: resource.ComposeAggregateTestCheckFunc(
          resource.TestCheckResourceAttr("data.bcm_cmdevice_categories.test", "categories.#", "1"),
          resource.TestCheckResourceAttr("data.bcm_cmdevice_categories.test", "categories.0.name", "default"),
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.uuid"),
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.software_image_id"),
        ),
      },
    },
  })
}
```

#### Test 3: Nested Attributes
```go
func TestAccCMDeviceCategoriesDataSource_NestedAttributes(t *testing.T) {
  resource.Test(t, resource.TestCase{
    PreCheck:                 func() { testAccPreCheck(t) },
    ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
    Steps: []resource.TestStep{
      {
        Config: testAccCMDeviceCategoriesDataSourceConfig(),
        Check: resource.ComposeAggregateTestCheckFunc(
          // Verify nested attributes exist
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.modules"),
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.fsmounts"),
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.roles"),
          resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.services"),
        ),
      },
    },
  })
}
```

## Success Criteria

- [ ] All acceptance tests pass
- [ ] Data source retrieves categories from BCM API
- [ ] Name filtering works correctly
- [ ] All essential attributes (FR3) are exposed
- [ ] Nested attributes (FR4) are properly structured
- [ ] Error handling works for API failures
- [ ] Documentation generated successfully
- [ ] Examples are clear and functional
- [ ] Code follows HashiCorp Terraform Provider best practices
- [ ] No hardcoded values or test pollution

## Dependencies

- BCM API endpoint accessible
- BCM credentials configured
- `internal/provider/bcm_client.go` - API client
- Helper functions: `getStringValue()`, `getBoolValue()`, `getInt64Value()`
- Terraform Plugin Framework v1.16.1
- Terraform Plugin Testing v1.13.3

## References

- BCM API Investigation: `/workspace/sampleRest/category_schema_documentation_20251121_070629.md`
- API Full Schema: `/workspace/sampleRest/category_full_schema_20251121_070627.json`
- Workflow Analysis: `/workspace/RESOURCE_ANALYSIS.md`
- CMDevice API Docs: `/workspace/sampleRest/CMDevice_Complete_Documentation.md`
- Existing Pattern: `data_source_cmpart_softwareimages.go`
- Existing Pattern: `data_source_cmdevice_nodes.go`

## Out of Scope

- Category resource implementation (separate feature)
- Category creation/modification (data source is read-only)
- Disk setup validation
- BMC credential management
- Exclude list parsing/validation

## Future Enhancements

1. **Category Resource** - Create/update/delete categories
2. **Disk Setup Data Source** - Query available disk setup XMLs
3. **Category Validation** - Validate category configuration consistency
4. **Category Diff** - Compare categories for differences
5. **Category Templates** - Pre-defined category templates for common use cases
