# Quickstart Guide - BCM Terraform Provider POV

**Version**: 0.1.0 (POV)
**Last Updated**: 2025-11-20

This guide helps you get started with the BCM Terraform Provider POV, which integrates with Nvidia BCM (BlueField Configuration Manager) to query software image inventory.

---

## What's Included in the POV

The POV (Proof of Value) validates the BCM JSON-RPC API integration with minimal scope:

- **Provider Configuration**: HTTP Basic Auth with self-signed certificate support
- **Data Source**: `bcm_cmpart_softwareimages` - query all software images from BCM
- **Full Schema**: 30+ fields per software image including nested kernel modules

**Not Included** (deferred to post-POV):
- Additional data sources (devices, hosts, networks)
- Write operations (resources)
- Certificate-based authentication (mTLS)
- Filtering by name/UUID

---

## Prerequisites

1. **BCM Instance**: Access to BCM server with JSON-RPC API enabled
   - Example: `https://172.21.15.254:8081`
2. **Credentials**: BCM username and password with CMPart service access
   - Example: `root` / `Hashicorp123!`
3. **Network Access**: Connectivity to BCM server on port 8081 (HTTPS)
4. **Terraform**: Terraform >= 1.5.0 installed

---

## Installation

### Step 1: Configure Terraform Provider

Create a new directory for your Terraform configuration:

```bash
mkdir bcm-provider-test
cd bcm-provider-test
```

Create `main.tf`:

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true    # Required for self-signed certificates
  timeout              = 30       # Optional: API timeout in seconds (default: 30)
}
```

**Security Note**: Store credentials in environment variables for production:

```hcl
provider "bcm" {
  endpoint             = var.bcm_endpoint
  username             = var.bcm_username
  password             = var.bcm_password
  insecure_skip_verify = true
}
```

Create `variables.tf`:

```hcl
variable "bcm_endpoint" {
  description = "BCM JSON-RPC API endpoint"
  type        = string
  default     = "https://172.21.15.254:8081"
}

variable "bcm_username" {
  description = "BCM username"
  type        = string
  sensitive   = true
}

variable "bcm_password" {
  description = "BCM password"
  type        = string
  sensitive   = true
}
```

Create `terraform.tfvars` (add to .gitignore):

```hcl
bcm_username = "root"
bcm_password = "Hashicorp123!"
```

### Step 2: Initialize Terraform

```bash
terraform init
```

Expected output:
```
Initializing provider plugins...
- Finding hashicorp/bcm versions matching "~> 0.1"...
- Installing hashicorp/bcm v0.1.0...
- Installed hashicorp/bcm v0.1.0 (signed by HashiCorp)

Terraform has been successfully initialized!
```

---

## Using the bcm_cmpart_softwareimages Data Source

### Basic Query - Fetch All Software Images

Add to `main.tf`:

```hcl
data "bcm_cmpart_softwareimages" "all" {}

output "all_images" {
  description = "All software images from BCM"
  value       = data.bcm_cmpart_softwareimages.all.images
}
```

Run Terraform:

```bash
terraform plan
terraform apply
```

Expected output:
```
Changes to Outputs:
  + all_images = [
      + {
          + id                          = "550e8400-e29b-41d4-a716-446655440000"
          + uuid                        = "550e8400-e29b-41d4-a716-446655440000"
          + name                        = "ubuntu-22.04-dpu"
          + path                        = "/var/bcm/images/ubuntu-22.04"
          + kernel_version              = "5.15.0-58-generic"
          + enable_sol                  = true
          + modules                     = [
              + {
                  + name       = "nvidia-drm"
                  + parameters = "modeset=1"
                  + revision   = "525.60.13"
                  ...
                }
            ]
          ...
        }
    ]
```

### Filter Results with HCL

**Count images**:
```hcl
output "image_count" {
  value = length(data.bcm_cmpart_softwareimages.all.images)
}
```

**Find images by name pattern**:
```hcl
output "ubuntu_images" {
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img if can(regex("ubuntu", lower(img.name)))
  ]
}
```

**List image names only**:
```hcl
output "image_names" {
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img.name
  ]
}
```

**Find images with specific kernel version**:
```hcl
output "kernel_5_15_images" {
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img.name if can(regex("^5\\.15", img.kernel_version))
  ]
}
```

**Images with kernel modules**:
```hcl
output "images_with_modules" {
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    {
      name          = img.name
      module_count  = length(img.modules)
      module_names  = [for mod in img.modules : mod.name]
    }
    if length(img.modules) > 0
  ]
}
```

**Images with SOL enabled**:
```hcl
output "sol_enabled_images" {
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    {
      name     = img.name
      sol_port = img.sol_port
      sol_speed = img.sol_speed
    }
    if img.enable_sol
  ]
}
```

### Access Nested Modules

```hcl
output "nvidia_modules" {
  value = flatten([
    for img in data.bcm_cmpart_softwareimages.all.images : [
      for mod in img.modules :
      {
        image_name     = img.name
        module_name    = mod.name
        module_version = mod.revision
      }
      if can(regex("nvidia", lower(mod.name)))
    ]
  ])
}
```

### Use in Other Resources

```hcl
# Example: Use image UUIDs to provision VMs (conceptual - requires additional provider)
locals {
  ubuntu_image_uuid = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img.uuid if img.name == "ubuntu-22.04-dpu"
  ][0]
}

# Could be used with another provider:
# resource "some_provider_vm" "example" {
#   image_id = local.ubuntu_image_uuid
# }
```

---

## Complete Schema Reference

The `bcm_cmpart_softwareimages` data source exposes these attributes:

### Root Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Data source identifier (placeholder) |
| `images` | list(object) | List of software image objects |

### SoftwareImage Object Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Image identifier (same as uuid) |
| `uuid` | string | Image UUID |
| `name` | string | Image name |
| `path` | string | Image file path on BCM server |
| `kernel_version` | string | Linux kernel version |
| `kernel_parameters` | string | Kernel boot parameters |
| `kernel_output_console` | string | Kernel console configuration |
| `bootfs_part` | string | Boot filesystem partition |
| `fs_part` | string | Root filesystem partition |
| `enable_sol` | bool | Serial Over LAN enabled |
| `sol_port` | string | SOL serial port |
| `sol_speed` | string | SOL baud rate |
| `sol_flow_control` | bool | SOL flow control enabled |
| `base_type` | string | Base OS type (e.g., "Linux") |
| `child_type` | string | OS distribution (e.g., "Ubuntu") |
| `creation_time` | number | Unix timestamp (milliseconds) |
| `file_operation_in_progress` | bool | File operation in progress |
| `modified` | bool | Image has been modified |
| `notes` | string | Free-form notes |
| `original_image` | string | Original base image name |
| `parent_software_image` | string | Parent image name |
| `parent_uuid` | string | Parent image UUID |
| `revision` | string | Image revision string |
| `revision_id` | number | Numeric revision ID |
| `to_be_removed` | bool | Image scheduled for deletion |
| `modules` | list(object) | Kernel modules (see below) |

### KernelModule Object Attributes (within modules)

| Attribute | Type | Description |
|-----------|------|-------------|
| `uuid` | string | Module UUID |
| `name` | string | Module name |
| `parameters` | string | Module load parameters |
| `base_type` | string | Module category |
| `child_type` | string | Module subcategory |
| `revision` | string | Module version |
| `modified` | bool | Module modified flag |
| `to_be_removed` | bool | Module scheduled for removal |

---

## Troubleshooting

### Error: TLS certificate verification failed

**Problem**: Provider cannot verify BCM's self-signed certificate

**Solution**: Set `insecure_skip_verify = true` in provider configuration

```hcl
provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true  # Add this line
}
```

**Warning**: This disables TLS security checks. Only use for testing. For production, configure certificate-based authentication.

---

### Error: Authentication failed (HTTP 401)

**Problem**: Invalid username or password

**Solution**: Verify credentials in provider configuration

```bash
# Test credentials with curl:
curl -k -u "root:Hashicorp123!" -X POST \
  -H "Content-Type: application/json" \
  -d '{"service":"CMPart","call":"getSoftwareImages"}' \
  https://172.21.15.254:8081/json
```

If curl succeeds but provider fails, check for:
- Special characters in password (ensure proper escaping)
- Whitespace in username/password
- Environment variable values (print with `terraform console`)

---

### Error: Connection timeout

**Problem**: Cannot reach BCM server

**Checklist**:
1. Verify BCM server is running: `ping 172.21.15.254`
2. Check port 8081 is open: `telnet 172.21.15.254 8081`
3. Verify network routing to BCM server
4. Check firewall rules
5. Increase timeout: `timeout = 60` in provider config

---

### Error: Failed to parse JSON response

**Problem**: Unexpected API response format

**Debug Steps**:
1. Enable Terraform trace logging:
   ```bash
   export TF_LOG=TRACE
   export TF_LOG_PATH=./terraform.log
   terraform apply
   ```

2. Check `terraform.log` for full API response

3. Verify API endpoint is correct (should be `/json` path)

4. Test API directly with curl to see raw response

---

### No images returned (empty list)

**Problem**: Data source returns empty array

**Check**:
1. Verify software images exist in BCM:
   ```bash
   curl -k -u "root:Hashicorp123!" -X POST \
     -H "Content-Type: application/json" \
     -d '{"service":"CMPart","call":"getSoftwareImages"}' \
     https://172.21.15.254:8081/json
   ```

2. If API returns `[]`, this is expected behavior (no images in BCM)

3. If API returns images but Terraform shows empty, check Terraform logs for parsing errors

---

## Best Practices

### 1. Store Credentials Securely

Never commit credentials to version control:

```bash
# .gitignore
terraform.tfvars
*.tfstate
*.tfstate.backup
```

Use environment variables or secrets management:

```bash
export TF_VAR_bcm_username="root"
export TF_VAR_bcm_password="Hashicorp123!"
terraform apply
```

### 2. Use Data Source for Read-Only Queries

The POV data source is read-only. Use it for:
- Discovering available software images
- Validating image configurations
- Feeding image UUIDs to other resources
- Reporting and documentation

### 3. Refresh Data on Every Apply

Data sources re-query on every `terraform refresh` or `terraform apply`:

```bash
# Get latest software image data
terraform refresh
```

### 4. Filter in HCL, Not API

POV does not support API-side filtering. Filter in Terraform:

```hcl
locals {
  production_images = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    img if can(regex("prod", lower(img.name)))
  ]
}
```

### 5. Handle Null Fields

Some fields may be null. Use `coalesce` or null checks:

```hcl
output "image_notes" {
  value = [
    for img in data.bcm_cmpart_softwareimages.all.images :
    coalesce(img.notes, "No notes")
  ]
}
```

---

## Next Steps

After POV validation:

1. **Expand Data Sources**: Add getDevices, getHosts, etc.
2. **Add Filtering**: Test API parameter support for name/UUID filters
3. **Certificate Auth**: Configure mTLS for production security
4. **Write Operations**: Explore resource creation if API supports it
5. **Terraform Registry**: Publish provider for team consumption

---

## Support and Feedback

For POV feedback and issues:

1. Check Terraform logs with `TF_LOG=TRACE`
2. Test API directly with curl
3. Review BCM API documentation
4. Report issues with full logs and API responses

**POV Scope**: This quickstart covers POV functionality only. Additional features will be documented in future releases.
