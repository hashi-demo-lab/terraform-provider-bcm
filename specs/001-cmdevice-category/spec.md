# Feature Specification: BCM CMDevice Category Resource

**Feature Branch**: `001-cmdevice-category`
**Created**: 2025-11-21
**Status**: Draft
**Input**: User description: "Implement a Terraform resource for managing BCM device categories following the same patterns as the existing bcm_cmpart_softwareimage resource."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create and Manage Device Category Configuration (Priority: P1)

Infrastructure administrators need to define and manage device categories that establish default configuration templates for cluster nodes. Categories define how nodes are provisioned, including disk layouts, boot configurations, kernel parameters, network settings, and filesystem mounts.

**Why this priority**: Categories are foundational configuration templates that must exist before nodes can be assigned to them. This is the core MVP functionality - without the ability to create and manage categories, users cannot leverage Terraform for BCM category lifecycle management.

**Independent Test**: Can be fully tested by creating a new category with basic configuration (name, description, disk setup) via Terraform, verifying it appears in BCM, and validating that updates are reflected correctly. Delivers immediate value by enabling Infrastructure-as-Code for category management.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with no custom categories, **When** I apply a Terraform configuration defining a new category with name "gpu-nodes" and disk setup XML, **Then** the category is created in BCM with the specified configuration
2. **Given** an existing category managed by Terraform, **When** I update the notes field and apply the changes, **Then** the category notes are updated in BCM without recreating the resource
3. **Given** an existing category managed by Terraform, **When** I modify the kernel parameters and apply changes, **Then** the category's kernel configuration is updated and all computed fields remain intact

---

### User Story 2 - Import Existing Categories into Terraform State (Priority: P2)

Infrastructure teams need to bring existing BCM categories under Terraform management without disrupting current node assignments or configurations.

**Why this priority**: Many BCM deployments already have categories configured manually. Import functionality enables teams to adopt Terraform gradually without requiring cluster reconfiguration or downtime.

**Independent Test**: Can be tested by creating a category manually in BCM, then importing it into Terraform state using `terraform import`, and verifying that subsequent applies detect no changes (demonstrating accurate state representation).

**Acceptance Scenarios**:

1. **Given** a category already exists in BCM with UUID "abc-123", **When** I run `terraform import bcm_cmdevice_category.test abc-123`, **Then** the category's full configuration is imported into Terraform state
2. **Given** an imported category, **When** I run `terraform plan`, **Then** no changes are detected (state matches actual BCM configuration)
3. **Given** an imported category, **When** I modify a field in Terraform and apply, **Then** only the changed field is updated in BCM

---

### User Story 3 - Delete Categories Safely (Priority: P3)

Administrators need to remove obsolete category configurations from BCM when they are no longer needed, with appropriate safety controls.

**Why this priority**: Category deletion is important for cleanup and governance, but is less critical than creation/update. Categories that are actively assigned to nodes should be protected from accidental deletion.

**Independent Test**: Can be tested by creating a category via Terraform, then destroying it, and verifying it's removed from BCM. Also test that force parameter behavior works correctly for categories with node assignments.

**Acceptance Scenarios**:

1. **Given** a category with no nodes assigned, **When** I run `terraform destroy`, **Then** the category is successfully removed from BCM
2. **Given** a category with nodes assigned, **When** I attempt to destroy without force flag, **Then** BCM API returns an error and Terraform reports the failure clearly
3. **Given** a category with force parameter set to true, **When** I destroy the resource, **Then** BCM removes the category even if nodes are assigned

---

### Edge Cases

- What happens when updating a category that has active node assignments (provisioning in progress)?
- How does system handle conflicting category names during creation?
- What validation occurs for complex nested objects (bmcSettings, softwareImageProxy)?
- How are very long text fields (disksetup XML, excludeLists) handled in state and API calls?
- What happens when the management network UUID referenced by a category no longer exists?
- How does validation handle optional nested objects that are set to null vs. empty objects?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST allow users to create categories with required fields (name, UUID, management network reference)
- **FR-002**: System MUST allow users to update category configuration including boot loader settings, kernel parameters, disk setup XML, and all other mutable fields
- **FR-003**: Users MUST be able to delete categories, with force parameter controlling deletion behavior for categories with node assignments
- **FR-004**: System MUST support importing existing categories into Terraform state using UUID
- **FR-005**: System MUST validate category names are unique within the cluster
- **FR-006**: System MUST persist all category configuration including complex nested objects (bmcSettings, softwareImageProxy, fsmounts, roles, modules)
- **FR-007**: System MUST handle eventual consistency for category operations that may require server-side processing time
- **FR-008**: System MUST preserve computed fields (UUID, parent_uuid, revision) during CRUD operations
- **FR-009**: System MUST call validateCategory API before create/update operations to catch configuration errors early
- **FR-010**: System MUST support optional force parameter for addCategory and updateCategory operations
- **FR-011**: System MUST correctly handle null values for optional nested objects (accessSettings, biosSetup, dpuSettings, proxySettings, seLinuxSettings, timeZoneSettings, ztpSettings)
- **FR-012**: System MUST use efficient getCategory(name) API for Read operations instead of list+filter pattern
- **FR-013**: System MUST handle long text fields (disksetup XML up to 10KB, excludeList fields up to 50KB) without truncation

### Key Entities

- **Category**: Represents a node configuration template with provisioning settings, boot configuration, kernel parameters, network settings, filesystem mounts, and role assignments. Categories are assigned to nodes to define their default configuration during provisioning.

- **SoftwareImageProxy**: Nested object referencing the parent software image used for node provisioning, containing UUID and revision information.

- **BMCSettings**: Nested object defining BIOS Management Controller configuration for remote hardware management, including credentials, power control settings, and firmware management mode.

- **FSMount**: Nested object within fsmounts array defining filesystem mount configurations, including device path, mount point, filesystem type, and mount options.

- **KernelModule**: Nested object within modules array defining kernel modules to load at boot with optional parameters.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Infrastructure administrators can create a new category configuration in under 5 minutes using Terraform
- **SC-002**: Category updates apply changes without node reprovisioning unless explicitly required
- **SC-003**: 100% of category fields returned by BCM API are represented in Terraform resource schema
- **SC-004**: Import operation successfully captures 100% of category configuration on first attempt
- **SC-005**: Category validation errors are detected before API calls 95% of the time, providing clear error messages
- **SC-006**: Terraform plan accurately detects configuration drift for categories modified outside Terraform
- **SC-007**: Category resource supports full lifecycle management (create, read, update, delete, import) with test coverage greater than 80 percent

## API Contract *(mandatory)*

### BCM CMDevice Service

**Service**: `cmdevice`
**Endpoint**: `POST https://{bcm-host}:8081/json`
**Authentication**: Session cookie (`cm-login-token`)

### Category Entity Structure

The Category entity represents a comprehensive node configuration template. Below is the complete schema:

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `uuid` | string | Computed | Unique identifier (RFC 4122 format) |
| `baseType` | string | Computed | Always "Category" |
| `childType` | string | Computed | Empty string for base category type |
| `name` | string | Required | Category name (must be unique) |
| `notes` | string | Optional | Administrative notes/description |
| `modified` | boolean | Computed | Indicates unsaved changes |
| `to_be_removed` | boolean | Computed | Scheduled for deletion flag |
| `revision` | string | Computed | Revision identifier |
| `parent_uuid` | string | Computed | Parent category UUID for inheritance |
| `managementNetwork` | string | Required | UUID of management network |
| `softwareImageProxy` | object | Optional | Software image reference (SoftwareImageProxy) |
| `bootLoader` | string | Optional | Boot loader type: SYSLINUX, GRUB, GRUB2, etc. |
| `bootLoaderFile` | string | Optional | Custom boot loader file path |
| `bootLoaderProtocol` | string | Optional | Boot protocol: HTTP, TFTP, etc. |
| `kernelVersion` | string | Optional | Kernel version string |
| `kernelParameters` | string | Optional | Kernel command line parameters |
| `kernelOutputConsole` | string | Optional | Kernel output console device |
| `modules` | array | Optional | Array of KernelModule objects |
| `disksetup` | string | Optional | XML configuration for disk partitioning |
| `raidconf` | string | Optional | RAID configuration |
| `installMode` | string | Optional | Installation mode: AUTO, FULL, etc. |
| `newNodeInstallMode` | string | Optional | New node installation mode |
| `installBootRecord` | boolean | Optional | Install boot record flag |
| `ioScheduler` | string | Optional | I/O scheduler type |
| `defaultGateway` | string | Optional | Default gateway IP address |
| `defaultGatewayMetric` | integer | Optional | Gateway metric value |
| `nameServers` | array | Optional | Array of DNS server IPs |
| `searchDomains` | array | Optional | Array of DNS search domains |
| `timeServers` | array | Optional | Array of NTP server addresses |
| `staticRoutes` | array | Optional | Array of static route definitions |
| `fsmounts` | array | Optional | Array of FSMount objects |
| `fsexports` | array | Optional | Array of NFS export configurations |
| `roles` | array | Optional | Array of Role assignments |
| `services` | array | Optional | Array of service configurations |
| `gpuSettings` | array | Optional | Array of GPU configurations |
| `bmcSettings` | object | Optional | BMC configuration (BMCSettings) |
| `biosSetup` | object | Optional | BIOS setup configuration |
| `dpuSettings` | object | Optional | DPU-specific settings |
| `proxySettings` | object | Optional | Proxy configuration |
| `accessSettings` | object | Optional | Access control settings |
| `seLinuxSettings` | object | Optional | SELinux configuration |
| `timeZoneSettings` | object | Optional | Timezone configuration |
| `ztpSettings` | object | Optional | Zero Touch Provisioning settings |
| `initialize` | string | Optional | Initialization script |
| `finalize` | string | Optional | Finalization script |
| `excludeListFull` | string | Optional | Full installation exclude patterns |
| `excludeListGrab` | string | Optional | Image grab exclude patterns |
| `excludeListGrabnew` | string | Optional | New image grab exclude patterns |
| `excludeListSync` | string | Optional | Sync operation exclude patterns |
| `excludeListUpdate` | string | Optional | Update operation exclude patterns |
| `excludeListManipulateScript` | string | Optional | Exclude list manipulation script |
| `authenticationService` | string | Optional | Authentication service: AUTO, LDAP, etc. |
| `allowNetworkingRestart` | boolean | Optional | Allow networking service restart |
| `dataNode` | boolean | Optional | Mark as data node |
| `nodeInstallerDisk` | boolean | Optional | Node installer disk flag |
| `versionConfigFiles` | boolean | Optional | Version configuration files |
| `interactiveUser` | string | Optional | Interactive user mode |
| `useExclusivelyFor` | string | Optional | Exclusive use designation |
| `fips` | string | Optional | FIPS mode: YES, NO |

### Nested Object: SoftwareImageProxy

| Field | Type | Description |
|-------|------|-------------|
| `uuid` | string | Proxy UUID |
| `baseType` | string | Always "SoftwareImageProxy" |
| `parentSoftwareImage` | string | UUID of parent software image |
| `revisionID` | integer | Revision number |

### Nested Object: BMCSettings

| Field | Type | Description |
|-------|------|-------------|
| `uuid` | string | BMC settings UUID |
| `baseType` | string | Always "BMCSettings" |
| `userName` | string | BMC username |
| `password` | string | BMC password (sensitive) |
| `privilege` | string | User privilege level |
| `userID` | integer | User ID |
| `firmwareManageMode` | string | Firmware management mode |
| `leakPolicy` | string | Leak detection policy |
| `leakReactionDelay` | float | Leak reaction delay seconds |
| `powerResetDelay` | integer | Power reset delay seconds |

### Nested Object: FSMount

| Field | Type | Description |
|-------|------|-------------|
| `uuid` | string | Mount UUID |
| `baseType` | string | Always "FSMount" |
| `device` | string | Device path or NFS export |
| `mountpoint` | string | Mount point path |
| `filesystem` | string | Filesystem type (nfs, xfs, ext4, etc.) |
| `mountoptions` | string | Mount options |
| `fsck` | string | Filesystem check mode |
| `dump` | boolean | Dump backup flag |
| `rdma` | boolean | Use RDMA for NFS |

### Nested Object: KernelModule

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Module name |
| `parameters` | string | Module parameters |

### API Methods

#### getCategories

Lists all categories in the cluster.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "getCategories"
}
```

**Response**: Array of Category objects

**Usage**: Data source pattern - list all, then filter client-side

---

#### getCategory

Retrieves a single category by name or UUID.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "getCategory",
  "args": ["category-name"]
}
```

**Arguments**:
- `name` (string): Category name
- OR `uuid` (string): Category UUID

**Response**: Single Category object

**Usage**: Resource Read operation - efficient direct lookup

---

#### addCategory

Creates a new category.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "addCategory",
  "args": [category_entity, force]
}
```

**Arguments**:
- `category` (object): Complete Category entity with `baseType: "Category"`, required fields populated
- `force` (boolean): Force creation even if validation warnings exist

**Response**: Category UUID (string)

**Error Handling**:
- Returns error if category name already exists
- Returns validation errors if required fields missing
- May return warnings that can be overridden with force=true

---

#### updateCategory

Updates an existing category.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "updateCategory",
  "args": [category_entity, force]
}
```

**Arguments**:
- `category` (object): Complete Category entity including UUID
- `force` (boolean): Force update even if validation warnings exist

**Response**: Success indicator or updated entity

**Error Handling**:
- Returns error if category UUID not found
- Returns validation errors if configuration invalid
- May return warnings about node impacts that can be overridden with force=true

---

#### validateCategory

Validates category configuration without persisting changes.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "validateCategory",
  "args": [category_entity]
}
```

**Arguments**:
- `category` (object): Category entity to validate

**Response**: Validation results with errors/warnings

**Usage**: Called before addCategory/updateCategory to provide early validation feedback

---

#### removeCategory

Deletes a category (method name inferred from BCM API patterns).

**Request** (inferred):
```json
{
  "service": "cmdevice",
  "call": "removeCategory",
  "args": [category_uuid, force]
}
```

**Arguments**:
- `uuid` (string): Category UUID to delete
- `force` (boolean): Force deletion even if nodes are assigned

**Response**: Success indicator

**Error Handling**:
- Returns error if category has nodes assigned (unless force=true)
- Returns error if category UUID not found

## Implementation Considerations *(mandatory)*

### Resource Schema Design

The Terraform resource schema should follow these principles:

1. **Identity Fields**:
   - `id` (computed): Resource identifier (same as UUID)
   - `uuid` (computed): BCM-assigned unique identifier
   - `name` (required): Category name (must be unique)

2. **Core Configuration** (required/optional with sensible defaults):
   - `management_network` (required): UUID of management network
   - `notes` (optional): Administrative notes
   - `boot_loader` (optional, default: "SYSLINUX")
   - `boot_loader_protocol` (optional, default: "HTTP")

3. **Kernel Configuration** (all optional):
   - `kernel_version`
   - `kernel_parameters`
   - `kernel_output_console`
   - `modules` (list of nested objects with name/parameters)

4. **Disk and Storage** (all optional):
   - `disksetup` (string - XML content up to 10KB)
   - `raidconf` (string)
   - `fsmounts` (list of nested FSMount objects)
   - `fsexports` (list of NFS export objects)

5. **Network Configuration** (all optional):
   - `default_gateway`
   - `default_gateway_metric`
   - `name_servers` (list of strings)
   - `search_domains` (list of strings)
   - `time_servers` (list of strings)
   - `static_routes` (list of objects)

6. **Provisioning** (optional with defaults):
   - `install_mode` (default: "AUTO")
   - `new_node_install_mode` (default: "FULL")
   - `install_boot_record` (boolean, default: false)
   - `initialize` (string - script)
   - `finalize` (string - script)

7. **Advanced Settings** (all optional, complex nested objects):
   - `software_image_proxy` (nested SoftwareImageProxy)
   - `bmc_settings` (nested BMCSettings - mark password as sensitive)
   - `bios_setup` (nested object)
   - `gpu_settings` (list of GPU config objects)
   - `roles` (list of Role objects)

8. **Exclude Lists** (optional long strings):
   - `exclude_list_full`
   - `exclude_list_grab`
   - `exclude_list_grabnew`
   - `exclude_list_sync`
   - `exclude_list_update`

9. **Force Parameter**:
   - `force` (optional, default: false): Control force behavior for create/update/delete operations

### CRUD Implementation Patterns

**Create**:
1. Build Category entity from plan with `baseType: "Category"`
2. Call validateCategory for pre-flight validation
3. Call addCategory with entity and force parameter
4. Parse UUID from response (may be string or object)
5. Use getCategory(name) to read back full entity
6. Map to Terraform state

**Read**:
1. Use efficient getCategory(name) API for direct lookup (NOT list+filter)
2. During import, use getCategories() to find by UUID, extract name, then use getCategory(name)
3. Parse response and map all fields to model
4. Handle null values appropriately for optional nested objects

**Update**:
1. Build Category entity from plan including UUID
2. Call validateCategory for pre-flight validation
3. Call updateCategory with entity and force parameter
4. Use getCategory(name) to read back updated entity
5. Verify changes applied correctly

**Delete**:
1. Call removeCategory with UUID and force parameter
2. Handle "category in use" errors gracefully when force=false
3. Return clear error message if deletion fails

**Import**:
1. Accept UUID as import identifier
2. Call getCategories() to list all categories
3. Find category with matching UUID, extract name
4. Call getCategory(name) to fetch full entity
5. Map to Terraform state with all fields populated

### Validation Approach

**Client-Side Validation**:
- Category name length (1-255 characters)
- Management network UUID format (RFC 4122)
- Disksetup XML well-formedness (optional)
- Boot loader values against known types
- IP address format for gateways and name servers

**Server-Side Validation**:
- Call validateCategory before create/update operations
- Parse validation response for errors and warnings
- Surface validation errors as Terraform diagnostics
- Only proceed with operation if validation passes (or force=true)

### State Management

**Computed Fields** (never in plan, always from API):
- `uuid`, `id`, `parent_uuid`, `revision`
- `modified`, `to_be_removed`
- `baseType`, `childType`

**Optional Nested Objects** (handle null gracefully):
- `accessSettings`, `biosSetup`, `dpuSettings`, `proxySettings`
- `seLinuxSettings`, `timeZoneSettings`, `ztpSettings`
- When null in API response, set to null in state (not empty object)

**Sensitive Fields**:
- Mark `bmc_settings.password` as sensitive
- Ensure it's not logged or displayed in plan output

**Long Text Fields**:
- Validate disksetup XML length (less than 10KB)
- Validate exclude list fields (less than 50KB each)
- Store as plain strings in state

### Error Handling

**Common API Errors**:
- 400 Bad Request: Invalid category configuration, name conflict
- 404 Not Found: Category UUID not found during read/update/delete
- 409 Conflict: Category name already exists
- 422 Validation Error: validateCategory found configuration errors

**Error Message Patterns**:
- "Category Creation Failed: {detailed error from API}"
- "Category Not Found: Category '{name}' does not exist in BCM"
- "Category Validation Failed: {validation errors}"
- "Category Deletion Failed: Category has {N} nodes assigned (use force=true to override)"

### Testing Requirements

**Acceptance Tests** (following TDD RED-GREEN-REFACTOR):

1. **TestAccCMDeviceCategoryResource_Basic**:
   - Create category with minimal configuration
   - Verify all fields in state
   - Update category name and notes
   - Verify updates applied
   - Delete category

2. **TestAccCMDeviceCategoryResource_Complete**:
   - Create category with comprehensive configuration
   - Include nested objects (bmcSettings, softwareImageProxy)
   - Include arrays (fsmounts, modules, nameServers)
   - Include long text fields (disksetup XML, excludeLists)
   - Verify all fields persisted correctly

3. **TestAccCMDeviceCategoryResource_Import**:
   - Create category manually in BCM
   - Import using UUID
   - Verify terraform plan shows no changes
   - Update imported category
   - Verify changes applied

4. **TestAccCMDeviceCategoryResource_Validation**:
   - Attempt to create category with invalid management network UUID
   - Verify validation error surfaced
   - Attempt to create category with duplicate name
   - Verify error handled gracefully

5. **TestAccCMDeviceCategoryResource_ForceParameter**:
   - Create category with force=false
   - Update with validation warning, verify failure
   - Update with force=true, verify success
   - Test deletion with/without force when nodes assigned

## Assumptions *(mandatory)*

1. **API Method Naming**: removeCategory method exists following BCM API conventions (validated through patterns in CMPart service)
2. **Management Network**: Management network UUID must exist before category creation (users handle network creation separately)
3. **Software Image**: softwareImageProxy can reference null or existing software image UUID
4. **Default Values**: BCM API applies reasonable defaults for optional fields not specified in request
5. **Validation Performance**: validateCategory API call completes within 2 seconds for typical configurations
6. **Character Encoding**: All text fields (notes, scripts, disksetup XML) support UTF-8 encoding
7. **Concurrent Modifications**: BCM API handles concurrent category updates with optimistic locking via revision field
8. **Large Configurations**: Category entities with all fields populated remain under 500KB JSON size
9. **Network Latency**: API calls to BCM complete within 5 seconds in typical network conditions
10. **Authentication**: BCM session cookies remain valid for duration of Terraform operations (typically 30+ minutes)

## Out of Scope *(mandatory)*

1. **Node Assignment**: This resource does NOT handle assigning categories to nodes (separate resource: bcm_cmdevice_node)
2. **Network Creation**: Management network must exist prior to category creation (separate resource: bcm_cmnet_network)
3. **Software Image Management**: Software images referenced by softwareImageProxy managed separately (existing resource: bcm_cmpart_softwareimage)
4. **Disksetup XML Builder**: Users must provide valid disksetup XML (no Terraform-native disk layout builder)
5. **Category Cloning**: No support for cloning categories (users can use Terraform resource duplication)
6. **Bulk Operations**: No support for batch category creation/deletion
7. **Category Templates**: No built-in category templates (users define via Terraform modules)
8. **Validation Rule Customization**: Uses BCM's built-in validation rules (no custom validation rule engine)
9. **Migration Tools**: No automated migration from other cluster management tools
10. **Category Dependencies**: No automatic dependency resolution between categories (users manage dependencies via Terraform)

## Dependencies *(mandatory)*

### Terraform Provider Dependencies

- **terraform-plugin-framework** (v1.16.1+): Core framework for resource implementation
- **terraform-plugin-log**: Structured logging for debugging
- **terraform-plugin-testing**: Acceptance test framework

### BCM API Dependencies

- **BCM Endpoint**: Accessible BCM API at `https://{host}:8081/json`
- **CMDevice Service**: cmdevice service with category methods (getCategories, getCategory, addCategory, updateCategory, validateCategory, removeCategory)
- **Session Authentication**: Valid BCM session cookie for API calls

### Related BCM Resources

- **bcm_cmnet_network**: Categories reference management network UUID (must exist)
- **bcm_cmpart_softwareimage**: Categories optionally reference software images via softwareImageProxy
- **bcm_cmdevice_node**: Nodes are assigned to categories (future resource)

### External Dependencies

- **Go 1.24+**: Provider compilation
- **BCM 9.x+**: Compatible BCM cluster version
- **Network Access**: HTTPS access to BCM API endpoint

## Example Terraform Configurations *(mandatory)*

### Minimal Category

```hcl
resource "bcm_cmdevice_category" "minimal" {
  name               = "minimal-category"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Minimal category configuration for testing"
}
```

### Category with Boot Configuration

```hcl
resource "bcm_cmdevice_category" "compute" {
  name               = "compute-nodes"
  management_network = var.management_network_id
  notes              = "Compute node category with custom kernel"

  boot_loader          = "GRUB2"
  boot_loader_protocol = "HTTP"

  kernel_version    = "5.15.0-58-generic"
  kernel_parameters = "quiet splash intel_iommu=on"
  kernel_output_console = "tty0"

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "mlx5_core"
      parameters = ""
    }
  ]

  install_mode           = "AUTO"
  new_node_install_mode = "FULL"
}
```

### Category with Disk Setup

```hcl
resource "bcm_cmdevice_category" "storage" {
  name               = "storage-nodes"
  management_network = var.management_network_id
  notes              = "Storage nodes with custom disk layout"

  disksetup = <<-EOX
    <?xml version="1.0" encoding="UTF-8"?>
    <diskSetup>
      <device>
        <blockdev>/dev/sda</blockdev>
        <partition id="a0" partitiontype="esp">
          <size>100M</size>
          <type>linux</type>
          <filesystem>fat</filesystem>
          <mountPoint>/boot/efi</mountPoint>
          <mountOptions>defaults,noatime,nodiratime</mountOptions>
        </partition>
        <partition id="a1">
          <size>50G</size>
          <type>linux</type>
          <filesystem>xfs</filesystem>
          <mountPoint>/</mountPoint>
          <mountOptions>defaults,noatime,nodiratime</mountOptions>
        </partition>
        <partition id="a2">
          <size>max</size>
          <type>linux</type>
          <filesystem>xfs</filesystem>
          <mountPoint>/data</mountPoint>
          <mountOptions>defaults,noatime,nodiratime</mountOptions>
        </partition>
      </device>
    </diskSetup>
  EOX
}
```

### Category with Network Configuration

```hcl
resource "bcm_cmdevice_category" "gpu_cluster" {
  name               = "gpu-cluster"
  management_network = var.management_network_id

  default_gateway       = "192.168.1.1"
  default_gateway_metric = 100

  name_servers = [
    "8.8.8.8",
    "8.8.4.4"
  ]

  search_domains = [
    "example.com",
    "cluster.local"
  ]

  time_servers = [
    "time.nist.gov"
  ]
}
```

### Category with Filesystem Mounts

```hcl
resource "bcm_cmdevice_category" "with_mounts" {
  name               = "nfs-client-nodes"
  management_network = var.management_network_id

  fsmounts = [
    {
      device       = "nfs-server.example.com:/export/home"
      mountpoint   = "/home"
      filesystem   = "nfs"
      mountoptions = "rsize=32768,wsize=32768,hard,async"
      fsck         = "NONE"
      dump         = false
      rdma         = false
    },
    {
      device       = "nfs-server.example.com:/export/scratch"
      mountpoint   = "/scratch"
      filesystem   = "nfs"
      mountoptions = "rsize=65536,wsize=65536,hard,async"
      fsck         = "NONE"
      dump         = false
      rdma         = true
    }
  ]
}
```

### Category with BMC Settings

```hcl
resource "bcm_cmdevice_category" "ipmi_managed" {
  name               = "ipmi-nodes"
  management_network = var.management_network_id

  bmc_settings = {
    user_name            = "admin"
    password             = var.bmc_password  # Marked as sensitive
    privilege            = "ADMINISTRATOR"
    firmware_manage_mode = "AUTO"
    leak_policy          = "NONE"
    power_reset_delay    = 5
  }
}
```

### Category with Software Image Proxy

```hcl
resource "bcm_cmdevice_category" "ubuntu_nodes" {
  name               = "ubuntu-22-04"
  management_network = var.management_network_id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.ubuntu_2204.uuid
  }

  install_mode = "AUTO"
}
```

### Category with Force Parameter

```hcl
resource "bcm_cmdevice_category" "force_example" {
  name               = "test-category"
  management_network = var.management_network_id
  force              = true  # Override validation warnings

  # Configuration that may trigger warnings
  kernel_parameters = "experimental_option=1"
}
```

### Import Existing Category

```bash
# Import by UUID
terraform import bcm_cmdevice_category.existing 0ae6d733-3015-4479-bfab-ce2d237a2809
```

```hcl
# After import, define matching configuration
resource "bcm_cmdevice_category" "existing" {
  name               = "default"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  # ... rest of configuration to match imported state
}
```

## Success Metrics *(mandatory)*

### Implementation Quality

- **Test Coverage**: Greater than 80 percent code coverage with comprehensive acceptance tests
- **API Compatibility**: All 6 API methods (getCategories, getCategory, addCategory, updateCategory, validateCategory, removeCategory) correctly implemented
- **Schema Completeness**: 100 percent of Category entity fields represented in resource schema (60+ attributes)
- **Validation Coverage**: Client-side and server-side validation for all user-provided fields

### User Experience

- **Documentation Quality**: Complete examples for 8+ common use cases
- **Error Clarity**: All API errors mapped to user-friendly Terraform diagnostics
- **Import Success Rate**: 100 percent of existing categories importable on first attempt
- **Plan Accuracy**: Zero false-positive drift detection for unmodified categories

### Performance

- **Create Operation**: Complete within 10 seconds for typical category configuration
- **Read Operation**: Complete within 3 seconds using efficient getCategory(name) API
- **Update Operation**: Complete within 10 seconds for configuration changes
- **Import Operation**: Complete within 5 seconds for category with full configuration

### Reliability

- **Test Success Rate**: 100 percent of acceptance tests pass consistently
- **Error Handling**: Graceful handling of all API error scenarios
- **State Consistency**: Terraform state always reflects actual BCM configuration after successful apply
- **Idempotency**: Running `terraform apply` twice with no changes results in zero modifications
