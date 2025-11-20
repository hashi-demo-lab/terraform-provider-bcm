# Data Model: BCM CMDevice Category Resource

**Date**: 2025-11-21
**Feature**: BCM CMDevice Category Resource
**Phase**: 1 - Data Model Design

## Overview

This document defines the data entities for the BCM CMDevice Category resource, including Terraform resource models, BCM API entities, and their mappings. The design follows terraform-plugin-framework patterns and reuses proven approaches from the `bcm_cmpart_softwareimage` resource.

---

## Entity: Category (Primary Resource)

### Terraform Resource Model

```go
// CMDeviceCategoryResourceModel describes the resource data model
type CMDeviceCategoryResourceModel struct {
    // Identity fields
    ID   types.String `tfsdk:"id"`   // Computed, same as UUID
    UUID types.String `tfsdk:"uuid"` // Computed, BCM-assigned
    Name types.String `tfsdk:"name"` // Required, unique

    // Core configuration
    Notes             types.String `tfsdk:"notes"`              // Optional
    ManagementNetwork types.String `tfsdk:"management_network"` // Required, UUID reference

    // Software image reference
    SoftwareImageProxy types.Object `tfsdk:"software_image_proxy"` // Optional, nested SoftwareImageProxyModel

    // Boot configuration
    BootLoader         types.String `tfsdk:"boot_loader"`          // Optional, default: "SYSLINUX"
    BootLoaderFile     types.String `tfsdk:"boot_loader_file"`     // Optional
    BootLoaderProtocol types.String `tfsdk:"boot_loader_protocol"` // Optional, default: "HTTP"

    // Kernel configuration
    KernelVersion       types.String `tfsdk:"kernel_version"`        // Optional
    KernelParameters    types.String `tfsdk:"kernel_parameters"`     // Optional
    KernelOutputConsole types.String `tfsdk:"kernel_output_console"` // Optional
    Modules             types.List   `tfsdk:"modules"`               // Optional, list of KernelModuleModel

    // Disk and storage
    Disksetup types.String `tfsdk:"disksetup"` // Optional, XML string up to 10KB
    Raidconf  types.String `tfsdk:"raidconf"`  // Optional

    // Installation settings
    InstallMode           types.String `tfsdk:"install_mode"`            // Optional, default: "AUTO"
    NewNodeInstallMode    types.String `tfsdk:"new_node_install_mode"`   // Optional, default: "FULL"
    InstallBootRecord     types.Bool   `tfsdk:"install_boot_record"`     // Optional, default: false
    IOScheduler           types.String `tfsdk:"io_scheduler"`            // Optional
    NodeInstallerDisk     types.Bool   `tfsdk:"node_installer_disk"`     // Optional
    VersionConfigFiles    types.Bool   `tfsdk:"version_config_files"`    // Optional
    AuthenticationService types.String `tfsdk:"authentication_service"`  // Optional

    // Network configuration
    DefaultGateway       types.String `tfsdk:"default_gateway"`        // Optional, IP address
    DefaultGatewayMetric types.Int64  `tfsdk:"default_gateway_metric"` // Optional
    NameServers          types.List   `tfsdk:"name_servers"`           // Optional, list of strings
    SearchDomains        types.List   `tfsdk:"search_domains"`         // Optional, list of strings
    TimeServers          types.List   `tfsdk:"time_servers"`           // Optional, list of strings
    StaticRoutes         types.List   `tfsdk:"static_routes"`          // Optional, list of objects
    AllowNetworkingRestart types.Bool `tfsdk:"allow_networking_restart"` // Optional

    // Filesystem configuration
    FSMounts  types.List `tfsdk:"fsmounts"`  // Optional, list of FSMountModel
    FSExports types.List `tfsdk:"fsexports"` // Optional, list of objects

    // Role and service assignments
    Roles    types.List `tfsdk:"roles"`    // Optional, list of objects
    Services types.List `tfsdk:"services"` // Optional, list of objects

    // Hardware-specific settings
    BMCSettings  types.Object `tfsdk:"bmc_settings"`  // Optional, nested BMCSettingsModel
    BiosSetup    types.Object `tfsdk:"bios_setup"`    // Optional, nested object
    DPUSettings  types.Object `tfsdk:"dpu_settings"`  // Optional, nested object
    GPUSettings  types.List   `tfsdk:"gpu_settings"`  // Optional, list of objects

    // Security and access
    AccessSettings   types.Object `tfsdk:"access_settings"`   // Optional, nested object
    SELinuxSettings  types.Object `tfsdk:"selinux_settings"`  // Optional, nested object
    ProxySettings    types.Object `tfsdk:"proxy_settings"`    // Optional, nested object
    TimeZoneSettings types.Object `tfsdk:"timezone_settings"` // Optional, nested object
    ZTPSettings      types.Object `tfsdk:"ztp_settings"`      // Optional, nested object
    FIPS             types.String `tfsdk:"fips"`              // Optional, "YES" or "NO"

    // Provisioning scripts
    Initialize types.String `tfsdk:"initialize"` // Optional, initialization script
    Finalize   types.String `tfsdk:"finalize"`   // Optional, finalization script

    // Exclude lists (large text fields)
    ExcludeListFull             types.String `tfsdk:"exclude_list_full"`              // Optional, up to 50KB
    ExcludeListGrab             types.String `tfsdk:"exclude_list_grab"`              // Optional, up to 50KB
    ExcludeListGrabnew          types.String `tfsdk:"exclude_list_grabnew"`           // Optional, up to 50KB
    ExcludeListSync             types.String `tfsdk:"exclude_list_sync"`              // Optional, up to 50KB
    ExcludeListUpdate           types.String `tfsdk:"exclude_list_update"`            // Optional, up to 50KB
    ExcludeListManipulateScript types.String `tfsdk:"exclude_list_manipulate_script"` // Optional

    // Behavioral flags
    DataNode         types.Bool   `tfsdk:"data_node"`          // Optional
    InteractiveUser  types.String `tfsdk:"interactive_user"`   // Optional
    UseExclusivelyFor types.String `tfsdk:"use_exclusively_for"` // Optional

    // Force parameter
    Force types.Bool `tfsdk:"force"` // Optional, default: false

    // Computed metadata fields
    ParentUUID   types.String `tfsdk:"parent_uuid"`   // Computed
    Revision     types.String `tfsdk:"revision"`      // Computed
    Modified     types.Bool   `tfsdk:"modified"`      // Computed
    ToBeRemoved  types.Bool   `tfsdk:"to_be_removed"` // Computed
    BaseType     types.String `tfsdk:"base_type"`     // Computed, always "Category"
    ChildType    types.String `tfsdk:"child_type"`    // Computed, empty for base type
}
```

### Field Mappings (Terraform ↔ BCM API)

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `id` | `uuid` | string | Computed, resource identifier |
| `uuid` | `uuid` | string | Computed, BCM-assigned UUID |
| `name` | `name` | string | Required, unique within cluster |
| `notes` | `notes` | string | Optional |
| `management_network` | `managementNetwork` | string | Required, UUID of network |
| `software_image_proxy` | `softwareImageProxy` | object | Optional, see SoftwareImageProxy entity |
| `boot_loader` | `bootLoader` | string | Optional, enum values |
| `boot_loader_file` | `bootLoaderFile` | string | Optional |
| `boot_loader_protocol` | `bootLoaderProtocol` | string | Optional, enum values |
| `kernel_version` | `kernelVersion` | string | Optional |
| `kernel_parameters` | `kernelParameters` | string | Optional |
| `kernel_output_console` | `kernelOutputConsole` | string | Optional |
| `modules` | `modules` | array | Optional, see KernelModule entity |
| `disksetup` | `disksetup` | string | Optional, XML content |
| `raidconf` | `raidconf` | string | Optional |
| `install_mode` | `installMode` | string | Optional, enum values |
| `new_node_install_mode` | `newNodeInstallMode` | string | Optional, enum values |
| `install_boot_record` | `installBootRecord` | boolean | Optional |
| `io_scheduler` | `ioScheduler` | string | Optional |
| `default_gateway` | `defaultGateway` | string | Optional, IP address |
| `default_gateway_metric` | `defaultGatewayMetric` | integer | Optional |
| `name_servers` | `nameServers` | array[string] | Optional |
| `search_domains` | `searchDomains` | array[string] | Optional |
| `time_servers` | `timeServers` | array[string] | Optional |
| `static_routes` | `staticRoutes` | array[object] | Optional |
| `fsmounts` | `fsmounts` | array[object] | Optional, see FSMount entity |
| `fsexports` | `fsexports` | array[object] | Optional |
| `roles` | `roles` | array[object] | Optional |
| `services` | `services` | array[object] | Optional |
| `bmc_settings` | `bmcSettings` | object | Optional, see BMCSettings entity |
| `bios_setup` | `biosSetup` | object | Optional |
| `dpu_settings` | `dpuSettings` | object | Optional |
| `gpu_settings` | `gpuSettings` | array[object] | Optional |
| `access_settings` | `accessSettings` | object | Optional |
| `selinux_settings` | `seLinuxSettings` | object | Optional |
| `proxy_settings` | `proxySettings` | object | Optional |
| `timezone_settings` | `timeZoneSettings` | object | Optional |
| `ztp_settings` | `ztpSettings` | object | Optional |
| `fips` | `fips` | string | Optional, enum: YES/NO |
| `initialize` | `initialize` | string | Optional, script content |
| `finalize` | `finalize` | string | Optional, script content |
| `exclude_list_full` | `excludeListFull` | string | Optional, up to 50KB |
| `exclude_list_grab` | `excludeListGrab` | string | Optional, up to 50KB |
| `exclude_list_grabnew` | `excludeListGrabnew` | string | Optional, up to 50KB |
| `exclude_list_sync` | `excludeListSync` | string | Optional, up to 50KB |
| `exclude_list_update` | `excludeListUpdate` | string | Optional, up to 50KB |
| `exclude_list_manipulate_script` | `excludeListManipulateScript` | string | Optional |
| `data_node` | `dataNode` | boolean | Optional |
| `node_installer_disk` | `nodeInstallerDisk` | boolean | Optional |
| `version_config_files` | `versionConfigFiles` | boolean | Optional |
| `authentication_service` | `authenticationService` | string | Optional, enum values |
| `allow_networking_restart` | `allowNetworkingRestart` | boolean | Optional |
| `interactive_user` | `interactiveUser` | string | Optional |
| `use_exclusively_for` | `useExclusivelyFor` | string | Optional |
| `force` | (method parameter) | boolean | Not persisted, controls API behavior |
| `parent_uuid` | `parent_uuid` | string | Computed |
| `revision` | `revision` | string | Computed |
| `modified` | `modified` | boolean | Computed |
| `to_be_removed` | `to_be_removed` | boolean | Computed |
| `base_type` | `baseType` | string | Computed, always "Category" |
| `child_type` | `childType` | string | Computed, empty string |

### Validation Rules

| Field | Rule | Error Message |
|-------|------|---------------|
| `name` | Length 1-255 characters | "Category name must be between 1 and 255 characters" |
| `management_network` | RFC 4122 UUID format | "Management network must be a valid UUID" |
| `boot_loader` | OneOf: SYSLINUX, GRUB, GRUB2, PXELINUX | "Invalid boot loader type" |
| `boot_loader_protocol` | OneOf: HTTP, TFTP, NFS | "Invalid boot loader protocol" |
| `default_gateway` | Valid IP address format | "Default gateway must be a valid IP address" |
| `name_servers` | Each element valid IP address | "Name servers must be valid IP addresses" |
| `disksetup` | Length ≤ 10,240 bytes (10KB) | "Disksetup XML must be less than 10KB" |
| `exclude_list_*` | Length ≤ 51,200 bytes (50KB) | "Exclude list must be less than 50KB" |
| `fips` | OneOf: YES, NO | "FIPS must be YES or NO" |
| `install_mode` | OneOf: AUTO, FULL, etc. | "Invalid install mode" |

### State Transitions

```
┌─────────┐
│ Planned │  terraform plan
└────┬────┘
     │
     ▼
┌─────────┐
│ Created │  terraform apply (addCategory API call)
└────┬────┘
     │
     ▼
┌─────────┐
│  Active │  ◄────┐ terraform apply (updateCategory API call)
└────┬────┘       │
     │            │
     ├────────────┘
     │
     ▼
┌─────────┐
│ Deleted │  terraform destroy (removeCategory API call)
└─────────┘
```

---

## Entity: SoftwareImageProxy (Nested Object)

### Terraform Model

```go
type SoftwareImageProxyModel struct {
    UUID                 types.String `tfsdk:"uuid"`                   // Computed
    ParentSoftwareImage  types.String `tfsdk:"parent_software_image"`  // Required, UUID reference
    RevisionID           types.Int64  `tfsdk:"revision_id"`            // Computed
}
```

### BCM API Entity

```json
{
  "uuid": "string",
  "baseType": "SoftwareImageProxy",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "parentSoftwareImage": "uuid-string",
  "revisionID": 0
}
```

### Field Mappings

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `uuid` | `uuid` | string | Computed, auto-generated |
| `parent_software_image` | `parentSoftwareImage` | string | Required, UUID of software image |
| `revision_id` | `revisionID` | integer | Computed |

---

## Entity: BMCSettings (Nested Object)

### Terraform Model

```go
type BMCSettingsModel struct {
    UUID               types.String  `tfsdk:"uuid"`                 // Computed
    UserName           types.String  `tfsdk:"user_name"`            // Optional
    Password           types.String  `tfsdk:"password"`             // Optional, sensitive
    Privilege          types.String  `tfsdk:"privilege"`            // Optional
    UserID             types.Int64   `tfsdk:"user_id"`              // Optional
    FirmwareManageMode types.String  `tfsdk:"firmware_manage_mode"` // Optional
    LeakPolicy         types.String  `tfsdk:"leak_policy"`          // Optional
    LeakReactionDelay  types.Float64 `tfsdk:"leak_reaction_delay"`  // Optional, seconds
    PowerResetDelay    types.Int64   `tfsdk:"power_reset_delay"`    // Optional, seconds
}
```

### BCM API Entity

```json
{
  "uuid": "string",
  "baseType": "BMCSettings",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "userName": "string",
  "password": "string",
  "privilege": "ADMINISTRATOR",
  "userID": 0,
  "firmwareManageMode": "AUTO",
  "leakPolicy": "NONE",
  "leakReactionDelay": 0.0,
  "powerResetDelay": 5
}
```

### Field Mappings

| Terraform Attribute | BCM API Field | Type | Sensitive | Notes |
|---------------------|---------------|------|-----------|-------|
| `uuid` | `uuid` | string | No | Computed |
| `user_name` | `userName` | string | No | Optional |
| `password` | `password` | string | **Yes** | Optional, marked sensitive |
| `privilege` | `privilege` | string | No | Optional, enum values |
| `user_id` | `userID` | integer | No | Optional |
| `firmware_manage_mode` | `firmwareManageMode` | string | No | Optional, enum values |
| `leak_policy` | `leakPolicy` | string | No | Optional, enum values |
| `leak_reaction_delay` | `leakReactionDelay` | float | No | Optional, seconds |
| `power_reset_delay` | `powerResetDelay` | integer | No | Optional, seconds |

---

## Entity: FSMount (Nested Object)

### Terraform Model

```go
type FSMountModel struct {
    UUID         types.String `tfsdk:"uuid"`          // Computed
    Device       types.String `tfsdk:"device"`        // Required
    Mountpoint   types.String `tfsdk:"mountpoint"`    // Required
    Filesystem   types.String `tfsdk:"filesystem"`    // Required
    MountOptions types.String `tfsdk:"mountoptions"`  // Optional
    Fsck         types.String `tfsdk:"fsck"`          // Optional
    Dump         types.Bool   `tfsdk:"dump"`          // Optional
    RDMA         types.Bool   `tfsdk:"rdma"`          // Optional
}
```

### BCM API Entity

```json
{
  "uuid": "string",
  "baseType": "FSMount",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "device": "nfs-server:/export/home",
  "mountpoint": "/home",
  "filesystem": "nfs",
  "mountoptions": "rsize=32768,wsize=32768",
  "fsck": "NONE",
  "dump": false,
  "rdma": false
}
```

### Field Mappings

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `uuid` | `uuid` | string | Computed |
| `device` | `device` | string | Required, device path or NFS export |
| `mountpoint` | `mountpoint` | string | Required, mount point path |
| `filesystem` | `filesystem` | string | Required, filesystem type (nfs, xfs, etc.) |
| `mountoptions` | `mountoptions` | string | Optional, mount options |
| `fsck` | `fsck` | string | Optional, filesystem check mode |
| `dump` | `dump` | boolean | Optional, dump backup flag |
| `rdma` | `rdma` | boolean | Optional, use RDMA for NFS |

---

## Entity: KernelModule (Nested Object)

### Terraform Model

```go
type KernelModuleModel struct {
    Name       types.String `tfsdk:"name"`       // Required
    Parameters types.String `tfsdk:"parameters"` // Optional
}
```

### BCM API Entity

```json
{
  "baseType": "KernelModule",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "name": "nvidia-drm",
  "parameters": "modeset=1"
}
```

### Field Mappings

| Terraform Attribute | BCM API Field | Type | Notes |
|---------------------|---------------|------|-------|
| `name` | `name` | string | Required, module name |
| `parameters` | `parameters` | string | Optional, module parameters |

---

## Entity Relationships

```
Category (1) ──────────────── (0..1) SoftwareImageProxy
    │                                      │
    │                                      └─► SoftwareImage (reference)
    │
    ├────────────────────── (0..1) BMCSettings
    │
    ├────────────────────── (0..*) FSMount
    │
    ├────────────────────── (0..*) KernelModule
    │
    ├────────────────────── (1) ManagementNetwork (reference)
    │
    └────────────────────── (0..*) Role (reference)
```

---

## Data Transformation Examples

### Example 1: Minimal Category (Terraform → BCM API)

**Terraform Configuration:**
```hcl
resource "bcm_cmdevice_category" "minimal" {
  name               = "minimal-category"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
}
```

**BCM API Entity (addCategory call):**
```json
{
  "baseType": "Category",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "name": "minimal-category",
  "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
}
```

### Example 2: Category with Nested Objects (Terraform → BCM API)

**Terraform Configuration:**
```hcl
resource "bcm_cmdevice_category" "with_nested" {
  name               = "gpu-nodes"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"

  bmc_settings = {
    user_name = "admin"
    password  = "secret123"
    privilege = "ADMINISTRATOR"
  }

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    }
  ]

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

**BCM API Entity (addCategory call):**
```json
{
  "baseType": "Category",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "name": "gpu-nodes",
  "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
  "bmcSettings": {
    "baseType": "BMCSettings",
    "childType": "",
    "modified": true,
    "to_be_removed": false,
    "revision": "",
    "userName": "admin",
    "password": "secret123",
    "privilege": "ADMINISTRATOR"
  },
  "modules": [
    {
      "baseType": "KernelModule",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "name": "nvidia-drm",
      "parameters": "modeset=1"
    }
  ],
  "fsmounts": [
    {
      "baseType": "FSMount",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "device": "nfs-server:/export/home",
      "mountpoint": "/home",
      "filesystem": "nfs",
      "mountoptions": "rsize=32768,wsize=32768"
    }
  ]
}
```

### Example 3: BCM API Response → Terraform State

**BCM API Response (getCategory call):**
```json
{
  "uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
  "baseType": "Category",
  "childType": "",
  "modified": false,
  "to_be_removed": false,
  "revision": "v1",
  "parent_uuid": null,
  "name": "gpu-nodes",
  "notes": "GPU compute nodes",
  "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
  "bootLoader": "GRUB2",
  "kernelVersion": "5.15.0-58-generic",
  "kernelParameters": "quiet splash",
  "modules": [
    {
      "name": "nvidia-drm",
      "parameters": "modeset=1"
    }
  ]
}
```

**Terraform State:**
```go
CMDeviceCategoryResourceModel{
    ID:                 types.StringValue("0ae6d733-3015-4479-bfab-ce2d237a2809"),
    UUID:               types.StringValue("0ae6d733-3015-4479-bfab-ce2d237a2809"),
    Name:               types.StringValue("gpu-nodes"),
    Notes:              types.StringValue("GPU compute nodes"),
    ManagementNetwork:  types.StringValue("84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"),
    BootLoader:         types.StringValue("GRUB2"),
    KernelVersion:      types.StringValue("5.15.0-58-generic"),
    KernelParameters:   types.StringValue("quiet splash"),
    Modules:            types.ListValueMust(...), // List of KernelModuleModel
    ParentUUID:         types.StringNull(),
    Revision:           types.StringValue("v1"),
    Modified:           types.BoolValue(false),
    ToBeRemoved:        types.BoolValue(false),
    BaseType:           types.StringValue("Category"),
    ChildType:          types.StringValue(""),
}
```

---

## Implementation Notes

### Null Value Handling

- **Optional Strings**: Use `types.StringNull()` when API returns null or empty
- **Optional Objects**: Use `types.ObjectNull()` when nested object not provided
- **Optional Lists**: Use `types.ListValueMust(elementType, []attr.Value{})` for empty arrays
- **Never propagate Unknown values to state** (causes "invalid result object" errors)

### Helper Function Usage

```go
// Reuse from data_source_cmpart_softwareimages.go
func getStringValue(data map[string]interface{}, key string) types.String
func getBoolValue(data map[string]interface{}, key string) types.Bool
func getInt64Value(data map[string]interface{}, key string) types.Int64

// Example usage in readCategory:
model.Name = getStringValue(categoryData, "name")
model.DataNode = getBoolValue(categoryData, "dataNode")
model.DefaultGatewayMetric = getInt64Value(categoryData, "defaultGatewayMetric")
```

### Sensitive Data Handling

```go
// Mark password as sensitive in schema
"password": schema.StringAttribute{
    Optional:  true,
    Sensitive: true,
    MarkdownDescription: "BMC password (sensitive)",
}
```

---

## Summary

This data model provides:
- Complete type-safe Terraform resource model with 60+ attributes
- Nested object models for complex configurations
- Clear field mappings between Terraform and BCM API
- Validation rules for all user-provided fields
- Transformation examples for common scenarios
- Implementation guidance for null handling and sensitive data

All design decisions follow terraform-plugin-framework best practices and proven patterns from the existing `bcm_cmpart_softwareimage` resource.
