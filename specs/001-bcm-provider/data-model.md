# Data Model - BCM Terraform Provider POV

**Date**: 2025-11-20
**Feature**: BCM Terraform Provider POV
**Branch**: 001-bcm-provider

This document defines the data entities exposed by the BCM Terraform provider POV and their mapping from the BCM JSON-RPC API to Terraform schema.

---

## Overview

The POV exposes software image data from the BCM CMPart service through a single data source. The data model includes:

1. **SoftwareImage** (Primary Entity) - BCM operating system images with kernel configuration
2. **KernelModule** (Nested Entity) - Linux kernel modules configured within software images

---

## Entity: SoftwareImage

**Source**: BCM JSON-RPC API - CMPart service, getSoftwareImages call

**Terraform Resource**: `data.bcm_cmpart_softwareimages` (data source, read-only)

**Description**: Represents a BCM software image (operating system image) with kernel configuration, Serial Over LAN settings, and associated kernel modules. Software images are used to provision DPU nodes in BCM-managed clusters.

### Fields

#### Identity Fields

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `id` | `uuid` | string (computed) | Terraform resource identifier (mapped from uuid) |
| `uuid` | `uuid` | string (computed) | Software image UUID (primary identifier in BCM) |
| `name` | `name` | string (computed) | Human-readable software image name |
| `path` | `path` | string (computed) | File system path to image on BCM server |

#### Kernel Configuration

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `kernel_version` | `kernelVersion` | string (computed) | Linux kernel version (e.g., "5.15.0-58-generic") |
| `kernel_parameters` | `kernelParameters` | string (computed) | Kernel boot parameters (e.g., "console=ttyS0") |
| `kernel_output_console` | `kernelOutputConsole` | string (computed) | Kernel output console configuration |

#### Partition Configuration

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `bootfs_part` | `bootfspart` | string (computed) | Boot filesystem partition (e.g., "/dev/mmcblk0p1") |
| `fs_part` | `fspart` | string (computed) | Root filesystem partition (e.g., "/dev/mmcblk0p2") |

#### Serial Over LAN (SOL) Settings

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `enable_sol` | `enableSOL` | bool (computed) | Serial Over LAN enabled flag |
| `sol_port` | `SOLPort` | string (computed) | SOL serial port (e.g., "ttyS0") |
| `sol_speed` | `SOLSpeed` | string (computed) | SOL baud rate (e.g., "115200") |
| `sol_flow_control` | `SOLFlowControl` | bool (computed) | SOL hardware flow control enabled |

#### Type Classification

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `base_type` | `baseType` | string (computed) | Base operating system type (e.g., "Linux") |
| `child_type` | `childType` | string (computed) | OS distribution type (e.g., "Ubuntu") |

#### Timestamps and Versioning

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `creation_time` | `creationTime` | int64 (computed) | Unix timestamp (milliseconds) when image was created |
| `revision` | `revision` | string (computed) | Image revision string (e.g., "1.0.0") |
| `revision_id` | `revisionID` | int64 (computed) | Numeric revision identifier |

#### Relationship Fields

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `original_image` | `originalImage` | string (computed) | Name of original base image |
| `parent_software_image` | `parentSoftwareImage` | string (computed) | Name of parent image (for derived images) |
| `parent_uuid` | `parent_uuid` | string (computed) | UUID of parent image |

#### State Flags

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `file_operation_in_progress` | `fileOperationInProgress` | bool (computed) | File operation currently executing on image |
| `modified` | `modified` | bool (computed) | Image has been modified from original |
| `to_be_removed` | `to_be_removed` | bool (computed) | Image scheduled for deletion |

#### Metadata

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `notes` | `notes` | string (computed) | Free-form notes about the image |

#### Nested Relationships

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `modules` | `modules` | list(KernelModule) (computed) | Kernel modules configured for this image |

### Validation Rules

- **UUID Format**: Must be valid UUID v4 format (validated by API)
- **Required Fields**: uuid and name are always present in API responses
- **Optional Fields**: All other fields may be null/missing (handled with types.StringNull(), etc.)
- **Modules Array**: May be empty array if no modules configured

### Example Terraform Output

```hcl
data "bcm_cmpart_softwareimages" "all" {}

output "example_image" {
  value = {
    id                  = data.bcm_cmpart_softwareimages.all.images[0].id
    uuid                = data.bcm_cmpart_softwareimages.all.images[0].uuid
    name                = data.bcm_cmpart_softwareimages.all.images[0].name
    kernel_version      = data.bcm_cmpart_softwareimages.all.images[0].kernel_version
    enable_sol          = data.bcm_cmpart_softwareimages.all.images[0].enable_sol
    modules_count       = length(data.bcm_cmpart_softwareimages.all.images[0].modules)
  }
}
```

---

## Entity: KernelModule

**Source**: Nested within SoftwareImage.modules array

**Description**: Represents a Linux kernel module (loadable kernel driver) configured to load with the software image. Modules extend kernel functionality for hardware drivers, filesystems, network protocols, etc.

### Fields

| Terraform Attribute | API Field | Type | Description |
|-------------------|-----------|------|-------------|
| `uuid` | `uuid` | string (computed) | Kernel module UUID |
| `name` | `name` | string (computed) | Module name (e.g., "nvidia-drm") |
| `parameters` | `parameters` | string (computed) | Module load parameters (e.g., "modeset=1") |
| `base_type` | `baseType` | string (computed) | Module category (e.g., "kernel-module") |
| `child_type` | `childType` | string (computed) | Module subcategory (e.g., "graphics") |
| `revision` | `revision` | string (computed) | Module version (e.g., "525.60.13") |
| `modified` | `modified` | bool (computed) | Module configuration has been modified |
| `to_be_removed` | `to_be_removed` | bool (computed) | Module scheduled for removal |

### Relationship

- **Parent Entity**: SoftwareImage
- **Cardinality**: Many KernelModule per SoftwareImage (one-to-many)
- **Access Pattern**: `data.bcm_cmpart_softwareimages.all.images[0].modules[0]`

### Example Terraform Output

```hcl
output "first_module" {
  value = {
    name       = data.bcm_cmpart_softwareimages.all.images[0].modules[0].name
    parameters = data.bcm_cmpart_softwareimages.all.images[0].modules[0].parameters
    revision   = data.bcm_cmpart_softwareimages.all.images[0].modules[0].revision
  }
}

output "gpu_modules" {
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images : [
      for mod in img.modules : mod.name
      if mod.child_type == "graphics"
    ]
  ]
}
```

---

## Field Mapping Rules

### camelCase to snake_case Conversion

API responses use camelCase, Terraform schema uses snake_case:

| API Field (camelCase) | Terraform Attribute (snake_case) |
|----------------------|-----------------------------------|
| `kernelVersion` | `kernel_version` |
| `kernelParameters` | `kernel_parameters` |
| `kernelOutputConsole` | `kernel_output_console` |
| `bootfspart` | `bootfs_part` |
| `fspart` | `fs_part` |
| `enableSOL` | `enable_sol` |
| `SOLPort` | `sol_port` |
| `SOLSpeed` | `sol_speed` |
| `SOLFlowControl` | `sol_flow_control` |
| `baseType` | `base_type` |
| `childType` | `child_type` |
| `creationTime` | `creation_time` |
| `fileOperationInProgress` | `file_operation_in_progress` |
| `originalImage` | `original_image` |
| `parentSoftwareImage` | `parent_software_image` |
| `parent_uuid` | `parent_uuid` (already snake_case) |
| `revisionID` | `revision_id` |
| `to_be_removed` | `to_be_removed` (already snake_case) |

### Type Conversions

| API Type | Terraform Type | Null Handling |
|----------|---------------|---------------|
| `string` | `types.String` | `types.StringNull()` if missing/null |
| `bool` | `types.Bool` | `types.BoolNull()` if missing/null |
| `number` (int) | `types.Int64` | `types.Int64Null()` if missing/null |
| `number` (float) | `types.Int64` (cast) | `types.Int64Null()` if missing/null |
| `[]object` (modules) | `[]KernelModuleModel` | Empty slice `[]KernelModuleModel{}` if missing/null |

### Special Mappings

- **ID field**: Mapped from `uuid` for Terraform resource identity
- **UUID preservation**: Keep both `id` and `uuid` attributes (id references uuid value)
- **Empty arrays**: `modules: null` or `modules: []` → empty slice in Terraform
- **Missing fields**: Any missing field → null value in Terraform state

---

## Data Source State Structure

The `bcm_cmpart_softwareimages` data source state contains:

```go
type CMPartSoftwareImagesDataSourceModel struct {
    ID     types.String         `tfsdk:"id"`      // Placeholder: "placeholder"
    Images []SoftwareImageModel `tfsdk:"images"`  // Array of software images
}
```

**ID field**: Data sources require an ID for Terraform state management. We use a placeholder value "placeholder" since the data source represents a list query, not a single resource.

---

## API to Terraform Mapping Example

**API Response** (JSON):
```json
[
  {
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "name": "ubuntu-22.04-dpu",
    "kernelVersion": "5.15.0-58-generic",
    "enableSOL": true,
    "modules": [
      {
        "uuid": "650e8400-e29b-41d4-a716-446655440000",
        "name": "nvidia-drm",
        "parameters": "modeset=1"
      }
    ]
  }
]
```

**Terraform State** (HCL):
```hcl
data "bcm_cmpart_softwareimages" "all" {
  id = "placeholder"

  images = [
    {
      id             = "550e8400-e29b-41d4-a716-446655440000"
      uuid           = "550e8400-e29b-41d4-a716-446655440000"
      name           = "ubuntu-22.04-dpu"
      kernel_version = "5.15.0-58-generic"
      enable_sol     = true

      modules = [
        {
          uuid       = "650e8400-e29b-41d4-a716-446655440000"
          name       = "nvidia-drm"
          parameters = "modeset=1"
        }
      ]
    }
  ]
}
```

---

## Out of Scope (POV)

The following entities are **not** included in the POV:

- **Other CMPart Entities**: Devices, Hosts, Networks, etc. (deferred to post-POV)
- **Other BCM Services**: CMJob, CMEvent, etc. (deferred to post-POV)
- **Write Operations**: No CREATE/UPDATE/DELETE resources in POV (read-only data sources only)
- **Filtering**: No input attributes to filter by name/uuid (returns all images in POV)

---

## Summary

**POV Data Model**:
- **1 Data Source**: `bcm_cmpart_softwareimages`
- **2 Entities**: SoftwareImage (30+ fields), KernelModule (8 fields)
- **Nested Relationship**: SoftwareImage has many KernelModule
- **Field Count**: 30+ SoftwareImage attributes + 8 KernelModule attributes
- **Mapping Strategy**: camelCase→snake_case with null-safe helpers

**Next Phase**: TDD Implementation - Use this data model to generate schemas, models, and acceptance tests.
