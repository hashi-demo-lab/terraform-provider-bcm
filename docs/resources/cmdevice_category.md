---
page_title: "bcm_cmdevice_category Resource - bcm"
subcategory: ""
description: |-
  Manages a BCM device category.
  Device categories define node configuration templates including boot configuration, kernel parameters, disk layouts, network settings, and filesystem mounts used to provision compute nodes in the BCM cluster.
---

# bcm_cmdevice_category (Resource)

Manages a BCM device category.

Device categories define node configuration templates including boot configuration, kernel parameters, disk layouts, network settings, and filesystem mounts used to provision compute nodes in the BCM cluster.

## Known Limitations

~> **Important**: The BCM API does not persist certain list fields on categories. The following fields are stored in Terraform state only and are not persisted to BCM:

| Attribute | Description |
|-----------|-------------|
| `static_routes` | Static network routes - BCM accepts values but returns empty arrays on read |
| `fsexports` | NFS filesystem exports - BCM accepts values but returns empty arrays on read |
| `roles` | Service role assignments - BCM accepts values but returns empty arrays on read; UUIDs are generated locally by the provider |
| `gpu_settings` | GPU hardware configuration - BCM accepts values but returns empty arrays on read |
| `services` | Service configurations - BCM accepts values but returns empty arrays on read |

**Impact:**
- After `terraform import`, these fields will be empty in state. Run `terraform apply` to restore configured values.
- The provider preserves plan values in state to prevent false drift detection.
- These fields work correctly for create, update, and destroy operations.

**Evidence:** See [GitHub Issue #73](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/73) for investigation details.

## Example Usage

```terraform
# Example 1: Minimal category configuration
# This is the simplest possible category with only required fields
resource "bcm_cmdevice_category" "minimal" {
  name               = "minimal-category"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe" # Replace with your management network UUID
}

# Example 2: Category with boot configuration
# Configure boot loader and kernel parameters for the category
resource "bcm_cmdevice_category" "with_boot" {
  name               = "gpu-compute-nodes"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "GPU compute nodes with CUDA drivers"

  # Boot loader configuration
  boot_loader          = "GRUB2"
  boot_loader_protocol = "HTTP"
  boot_loader_file     = "/boot/grub2/grubx64.efi"

  # Kernel configuration
  kernel_version        = "5.15.0-58-generic"
  kernel_parameters     = "quiet splash intel_iommu=on"
  kernel_output_console = "ttyS0,115200n8"

  # Kernel modules
  modules = [
    {
      name       = "nvidia"
      parameters = ""
    },
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    }
  ]
}

# Example 3: Category with disk setup XML
# Configure disk partitioning using disksetup XML
resource "bcm_cmdevice_category" "with_disk_setup" {
  name               = "storage-nodes"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Storage nodes with custom disk layout"

  # Disk setup with XML configuration
  disksetup = <<-EOT
    <?xml version="1.0" encoding="UTF-8"?>
    <diskSetup>
      <disk device="/dev/sda">
        <partition number="1" size="512M" type="ef00" label="EFI"/>
        <partition number="2" size="100G" type="8300" label="root"/>
        <partition number="3" size="remaining" type="8300" label="data"/>
      </disk>
    </diskSetup>
  EOT

  raidconf = <<-EOT
    DEVICE /dev/sd[a-d]
    ARRAY /dev/md0 level=raid10 num-devices=4
  EOT

  # Installation settings
  install_mode          = "FULL"
  new_node_install_mode = "FULL"
  install_boot_record   = true
}

# Example 4: Category with network configuration
# Configure networking with gateway, DNS, and NTP servers
resource "bcm_cmdevice_category" "with_network" {
  name               = "web-servers"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Web server nodes with custom network configuration"

  # Network settings
  default_gateway        = "192.168.1.1"
  default_gateway_metric = 100

  name_servers = [
    "8.8.8.8",
    "8.8.4.4",
    "1.1.1.1"
  ]

  search_domains = [
    "example.com",
    "internal.local"
  ]

  time_servers = [
    "time.google.com",
    "time.cloudflare.com"
  ]

  allow_networking_restart = true
}

# Example 5: Category with filesystem mounts
# Configure NFS and local filesystem mounts
resource "bcm_cmdevice_category" "with_mounts" {
  name               = "compute-with-shared-home"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Compute nodes with shared home directory"

  fsmounts = [
    {
      device       = "nfs-server:/export/home"
      mountpoint   = "/home"
      filesystem   = "nfs"
      mountoptions = "rsize=32768,wsize=32768,vers=4.2"
      fsck         = "NONE"
      dump         = false
      rdma         = false
    },
    {
      device       = "nfs-server:/export/scratch"
      mountpoint   = "/scratch"
      filesystem   = "nfs"
      mountoptions = "rsize=131072,wsize=131072,vers=4.2"
      fsck         = "NONE"
      dump         = false
      rdma         = true
    }
  ]
}

# Example 6: Category with BMC settings
# Configure BMC (Baseboard Management Controller) credentials
resource "bcm_cmdevice_category" "with_bmc" {
  name               = "managed-nodes"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Nodes with BMC remote management enabled"

  bmc_settings = {
    user_name            = "admin"
    password             = "changeme123" # Sensitive - will be hidden in logs
    privilege            = "ADMINISTRATOR"
    user_id              = 2
    firmware_manage_mode = "AUTO"
    leak_policy          = "NONE"
    leak_reaction_delay  = 300.0
    power_reset_delay    = 5
  }
}

# Example 7: Category with software image proxy
# Link category to a specific software image
resource "bcm_cmdevice_category" "with_software_image" {
  name               = "ubuntu-22-04-nodes"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Ubuntu 22.04 LTS compute nodes"

  software_image_proxy = {
    parent_software_image = "f8e9a7b6-4c3d-2e1f-0a9b-8c7d6e5f4a3b" # Replace with your software image UUID
  }

  kernel_parameters = "quiet splash"
}

# Example 8: Category with force parameter
# Use force to override warnings or delete categories with assigned nodes
resource "bcm_cmdevice_category" "with_force" {
  name               = "test-category"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Test category that can be forcefully modified"

  kernel_parameters = "debug verbose"

  # Force parameter allows operations even if there are warnings or assigned nodes
  # Use with caution in production environments
  force = true
}

# Example 9: Category with roles configuration
# Configure roles for the category - each role gets a BCM-assigned UUID after creation
resource "bcm_cmdevice_category" "with_roles" {
  name               = "roles-category"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Category with role definitions"

  software_image_proxy = {
    parent_software_image = "f8e9a7b6-4c3d-2e1f-0a9b-8c7d6e5f4a3b"
  }

  # Role definitions - BCM will assign UUIDs to each role after creation
  roles = [
    {
      name       = "head"
      child_type = "HeadNode"
    },
    {
      name       = "compute"
      child_type = "ComputeNode"
    }
  ]
}

# Output role UUIDs after creation (populated from BCM API)
output "head_role_uuid" {
  value       = bcm_cmdevice_category.with_roles.roles[0].uuid
  description = "BCM-assigned UUID for the head role"
}

output "compute_role_uuid" {
  value       = bcm_cmdevice_category.with_roles.roles[1].uuid
  description = "BCM-assigned UUID for the compute role"
}

# Example 10: Comprehensive category with multiple features
# This example shows many features combined
resource "bcm_cmdevice_category" "comprehensive" {
  name               = "production-gpu-cluster"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Production GPU cluster with comprehensive configuration"

  # Boot configuration
  boot_loader          = "GRUB2"
  boot_loader_protocol = "HTTP"

  # Kernel configuration
  kernel_version    = "5.15.0-58-generic"
  kernel_parameters = "quiet splash intel_iommu=on iommu=pt"

  modules = [
    {
      name       = "nvidia"
      parameters = ""
    },
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    }
  ]

  # Network configuration
  default_gateway = "10.0.0.1"
  name_servers    = ["8.8.8.8", "8.8.4.4"]
  search_domains  = ["prod.example.com"]
  time_servers    = ["time.google.com"]

  # Filesystem mounts
  fsmounts = [
    {
      device       = "nfs-prod:/home"
      mountpoint   = "/home"
      filesystem   = "nfs"
      mountoptions = "rsize=32768,wsize=32768,vers=4.2"
      fsck         = "NONE"
      dump         = false
      rdma         = false
    }
  ]

  # BMC settings for remote management
  bmc_settings = {
    user_name            = "admin"
    password             = "secure_password_here"
    privilege            = "ADMINISTRATOR"
    firmware_manage_mode = "AUTO"
  }

  # Software image reference
  software_image_proxy = {
    parent_software_image = "f8e9a7b6-4c3d-2e1f-0a9b-8c7d6e5f4a3b"
  }

  # Installation settings
  install_mode          = "FULL"
  new_node_install_mode = "FULL"
  install_boot_record   = true

  # Security settings
  fips = "NO"

  # Provisioning scripts
  initialize = <<-EOT
    #!/bin/bash
    # Initialization script runs after node provisioning
    echo "Initializing GPU node..."
    nvidia-smi
  EOT

  finalize = <<-EOT
    #!/bin/bash
    # Finalization script runs at the end of provisioning
    echo "GPU node provisioning complete"
    systemctl enable nvidia-persistenced
  EOT
}

# Example 11: Category with static routes
# Configure custom network routing for the category
resource "bcm_cmdevice_category" "with_static_routes" {
  name               = "multi-network-nodes"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Nodes with custom routing for multiple networks"

  # Static routes for network segmentation
  # BCM assigns UUIDs to each route after creation
  # NOTE: name and network (UUID) are required fields
  static_routes = [
    {
      name         = "route-datacenter-a"
      ip           = "10.100.0.0"
      netmask_bits = 16
      gateway      = "192.168.1.254"
      metric       = 100
      network      = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
    },
    {
      name         = "route-datacenter-b"
      ip           = "10.200.0.0"
      netmask_bits = 16
      gateway      = "192.168.1.253"
      metric       = 200
      network      = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
    },
    {
      name         = "route-private-net"
      ip           = "172.16.0.0"
      netmask_bits = 12
      gateway      = "192.168.1.252"
      metric       = 0
      network      = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
    }
  ]

  # Network configuration
  default_gateway = "192.168.1.1"
  name_servers    = ["8.8.8.8"]
}

# Example 12: Category with Nvidia GPU settings
# Configure Nvidia GPU power and compute settings
resource "bcm_cmdevice_category" "nvidia_gpu_nodes" {
  name               = "nvidia-a100-cluster"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Nvidia A100 GPU nodes with optimized settings"

  # GPU settings for Nvidia cards
  # Each entry configures a range of GPUs (e.g., "0" for GPU 0, "0-3" for GPUs 0-3)
  gpu_settings = [
    {
      name         = "0-3"     # Configure GPUs 0, 1, 2, 3
      child_type   = "nvidia"  # Required: "nvidia" or "amd"
      power_limit  = 400       # Power limit in Watts
      ecc_mode     = "ENABLED" # ECC memory: ENABLED, DISABLED, NONE
      compute_mode = "DEFAULT" # DEFAULT, EXCLUSIVE_PROCESS, EXCLUSIVE_THREAD, PROHIBITED
    },
    {
      name                       = "4-7"
      child_type                 = "nvidia"
      power_limit                = 350
      ecc_mode                   = "ENABLED"
      compute_mode               = "EXCLUSIVE_PROCESS"
      clock_sync_boost_mode      = "NONE"
      multiprocessor_clock_speed = 1800000000 # 1.8 GHz in Hz
      memory_clock_speed         = 1500000000 # 1.5 GHz in Hz
    }
  ]

  # Kernel modules for Nvidia
  modules = [
    {
      name       = "nvidia"
      parameters = ""
    },
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    }
  ]

  kernel_parameters = "intel_iommu=on iommu=pt"
}

# Example 13: Category with Nvidia MIG (Multi-Instance GPU) profiles
# Configure MIG partitioning for Nvidia GPUs
resource "bcm_cmdevice_category" "nvidia_mig_nodes" {
  name               = "nvidia-mig-cluster"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Nvidia GPUs with MIG partitioning enabled"

  gpu_settings = [
    {
      name         = "0"
      child_type   = "nvidia"
      ecc_mode     = "ENABLED"
      compute_mode = "DEFAULT"
      # MIG profiles define GPU partitioning
      # Format: "Xg.Ygb" where X=compute instances, Y=memory in GB
      mig_profiles = ["1g.5gb", "1g.5gb", "1g.5gb", "1g.5gb"]
    },
    {
      name         = "1"
      child_type   = "nvidia"
      ecc_mode     = "ENABLED"
      compute_mode = "DEFAULT"
      # Different MIG configuration for second GPU
      mig_profiles = ["2g.10gb", "2g.10gb"]
    }
  ]
}

# Example 14: Category with AMD GPU settings
# Configure AMD GPU clock and power settings
resource "bcm_cmdevice_category" "amd_gpu_nodes" {
  name               = "amd-mi250-cluster"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "AMD MI250 GPU nodes with optimized settings"

  # GPU settings for AMD cards
  gpu_settings = [
    {
      name               = "0-1"     # Configure GPUs 0 and 1
      child_type         = "amd"     # Required: "nvidia" or "amd"
      gpu_clock_level    = 5         # GPU clock frequency level (0-7)
      memory_clock_level = 2         # Memory clock frequency level (0-3)
      power_play         = "DEFAULT" # Power play mode
      fan_speed          = 128       # Fan speed value (0-255)
    },
    {
      name                 = "2-3"
      child_type           = "amd"
      gpu_clock_level      = 7
      memory_clock_level   = 3
      gpu_overdrive        = 0.1       # GPU overdrive percentage (0-0.2)
      minimal_gpu_clock    = 800000000 # Minimum GPU clock in Hz
      minimal_memory_clock = 400000000 # Minimum memory clock in Hz
      activity_threshold   = 0.5       # Workload threshold before clock change
      hysteresis_up        = 1.0       # Delay before clock increase (seconds)
      hysteresis_down      = 2.0       # Delay before clock decrease (seconds)
    }
  ]

  # Kernel modules for AMD
  modules = [
    {
      name       = "amdgpu"
      parameters = ""
    }
  ]
}

# Example 15: Mixed GPU category with static routes
# Comprehensive example combining GPU settings and network routing
resource "bcm_cmdevice_category" "mixed_gpu_with_routing" {
  name               = "mixed-gpu-cluster"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Mixed GPU nodes with custom routing"

  # Static routes for GPU data networks
  static_routes = [
    {
      name         = "gpu-data-network"
      ip           = "10.50.0.0"
      netmask_bits = 16
      gateway      = "192.168.1.100"
      metric       = 50
      network      = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
    }
  ]

  # Nvidia GPU settings
  gpu_settings = [
    {
      name         = "0"
      child_type   = "nvidia"
      power_limit  = 300
      compute_mode = "DEFAULT"
    }
  ]

  # Network configuration
  default_gateway = "192.168.1.1"
  name_servers    = ["8.8.8.8", "8.8.4.4"]
}

# Example 16: Category with OS Services
# Configure monitored system services with health checking
resource "bcm_cmdevice_category" "with_services" {
  name               = "service-monitored-nodes"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Nodes with monitored system services"

  services = [
    {
      name                          = "sshd"
      monitored                     = true
      autostart                     = true
      managed                       = false
      run_if                        = "ALWAYS"
      sickness_check_script_timeout = 10
      sickness_check_interval       = 60
      script_timeout                = 30
    },
    {
      name      = "nginx"
      monitored = true
      autostart = true
      managed   = true
      run_if    = "ALWAYS"
    },
    {
      name                    = "custom-daemon"
      monitored               = true
      autostart               = true
      managed                 = false
      sickness_check_script   = "/usr/local/bin/health_check.sh"
      sickness_check_interval = 120
    }
  ]
}

# Example 17: Category with services and health monitoring
# Configure services with custom health check scripts
resource "bcm_cmdevice_category" "services_with_health_check" {
  name               = "health-monitored-cluster"
  management_network = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  notes              = "Cluster with comprehensive health monitoring"

  services = [
    {
      name                          = "slurmd"
      monitored                     = true
      autostart                     = true
      managed                       = true
      run_if                        = "ALWAYS"
      sickness_check_script         = "/usr/local/bin/check_slurmd.sh"
      sickness_check_script_timeout = 30
      sickness_check_interval       = 60
      script_timeout                = 120
    },
    {
      name                          = "node_exporter"
      monitored                     = true
      autostart                     = true
      managed                       = false
      sickness_check_script         = "/usr/local/bin/check_exporter.sh"
      sickness_check_script_timeout = 10
      sickness_check_interval       = 30
    }
  ]
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `management_network` (String) Management network UUID reference (must be valid RFC 4122 UUID)
- `name` (String) Category name (must be unique, 1-255 characters)

### Optional

- `access_settings` (Attributes) Access settings. Note: Nested attributes not yet implemented - use raw API calls if needed. (see [below for nested schema](#nestedatt--access_settings))
- `allow_networking_restart` (Boolean) Allow networking restart flag. If not specified, BCM assigns a default.
- `authentication_service` (String) Authentication service (AUTO, LDAP, SSSD, LOCAL). If not specified, BCM assigns AUTO.
- `bios_setup` (Attributes) BIOS setup configuration. Note: Nested attributes not yet implemented - use raw API calls if needed. (see [below for nested schema](#nestedatt--bios_setup))
- `bmc_settings` (Attributes) BMC configuration settings (see [below for nested schema](#nestedatt--bmc_settings))
- `boot_loader` (String) Boot loader type (SYSLINUX, GRUB, GRUB2, PXELINUX). If not specified, BCM assigns a default.
- `boot_loader_file` (String) Boot loader file path. If not specified, BCM uses defaults based on boot_loader.
- `boot_loader_protocol` (String) Boot loader protocol (HTTP, TFTP, NFS). If not specified, BCM assigns a default.
- `data_node` (Boolean) Data node flag. If not specified, BCM assigns a default.
- `default_gateway` (String) Default gateway IP address. If not specified, BCM may assign a default.
- `default_gateway_metric` (Number) Default gateway metric. If not specified, BCM may assign a default.
- `disksetup` (String) Disk setup XML configuration (max 10KB)
- `dpu_settings` (Attributes) DPU settings. Note: Nested attributes not yet implemented - use raw API calls if needed. (see [below for nested schema](#nestedatt--dpu_settings))
- `exclude_list_full` (String) Exclude list for full operations (max 50KB)
- `exclude_list_grab` (String) Exclude list for grab operations (max 50KB)
- `exclude_list_grabnew` (String) Exclude list for grabnew operations (max 50KB)
- `exclude_list_manipulate_script` (String) Exclude list manipulate script
- `exclude_list_sync` (String) Exclude list for sync operations (max 50KB)
- `exclude_list_update` (String) Exclude list for update operations (max 50KB)
- `finalize` (String) Finalization script
- `fips` (String) FIPS mode (YES or NO). If not specified, BCM assigns NO.
- `force` (Boolean) Force parameter to override warnings and constraints
- `fsexports` (Attributes List) NFS filesystem exports for nodes in this category. Configures entries for /etc/exports. (see [below for nested schema](#nestedatt--fsexports))
- `fsmounts` (Attributes List) Filesystem mounts (see [below for nested schema](#nestedatt--fsmounts))
- `gpu_settings` (Attributes List) GPU hardware configuration for nodes in this category. Supports Nvidia and AMD GPUs with vendor-specific settings. **Known Limitation**: BCM API does not persist this field - values are stored in Terraform state only. After import, re-apply configuration to restore values. (see [below for nested schema](#nestedatt--gpu_settings))
- `initialize` (String) Initialization script
- `install_boot_record` (Boolean) Install boot record flag. If not specified, BCM assigns a default.
- `install_mode` (String) Installation mode (AUTO, FULL, MINIMAL, CUSTOM). If not specified, BCM assigns a default.
- `interactive_user` (String) Interactive user. If not specified, BCM assigns ALWAYS.
- `io_scheduler` (String) I/O scheduler
- `kernel_output_console` (String) Kernel output console device
- `kernel_parameters` (String) Kernel command-line parameters
- `kernel_version` (String) Kernel version string
- `modules` (Attributes List) Kernel modules to load (see [below for nested schema](#nestedatt--modules))
- `name_servers` (List of String) DNS name servers
- `new_node_install_mode` (String) New node installation mode (FULL, MINIMAL, SKIP). If not specified, BCM assigns a default.
- `node_installer_disk` (Boolean) Node installer disk flag. If not specified, BCM assigns a default.
- `notes` (String) User notes or description for the category
- `proxy_settings` (Attributes) Proxy settings. Note: Nested attributes not yet implemented - use raw API calls if needed. (see [below for nested schema](#nestedatt--proxy_settings))
- `raidconf` (String) RAID configuration
- `roles` (Attributes List) Service role assignments for nodes in this category. **Known Limitation**: BCM API does not persist this field - values are stored in Terraform state only. Role UUIDs are generated locally by the provider. After import, re-apply configuration to restore values. (see [below for nested schema](#nestedatt--roles))
- `search_domains` (List of String) DNS search domains
- `selinux_settings` (Attributes) SELinux settings. Note: Nested attributes not yet implemented - use raw API calls if needed. (see [below for nested schema](#nestedatt--selinux_settings))
- `services` (Attributes List) OS service configurations for nodes in this category. Defines which services should be monitored and managed by CMDaemon. **Known Limitation**: BCM API may not persist this field consistently - values are stored in Terraform state. After import, re-apply configuration to restore values. (see [below for nested schema](#nestedatt--services))
- `software_image_proxy` (Attributes) Software image proxy configuration (see [below for nested schema](#nestedatt--software_image_proxy))
- `static_routes` (Attributes List) Static network routes for nodes in this category. (see [below for nested schema](#nestedatt--static_routes))
- `time_servers` (List of String) NTP time servers
- `timezone_settings` (Attributes) Timezone settings. Note: Nested attributes not yet implemented - use raw API calls if needed. (see [below for nested schema](#nestedatt--timezone_settings))
- `use_exclusively_for` (String) Use exclusively for
- `version_config_files` (Boolean) Version config files flag. If not specified, BCM assigns a default.
- `ztp_settings` (Attributes) ZTP settings. Note: Nested attributes not yet implemented - use raw API calls if needed. (see [below for nested schema](#nestedatt--ztp_settings))

### Read-Only

- `base_type` (String) Base type (always 'Category')
- `child_type` (String) Child type
- `id` (String) Resource identifier (same as UUID)
- `modified` (Boolean) Modified flag
- `parent_uuid` (String) Parent UUID
- `revision` (String) Revision
- `to_be_removed` (Boolean) To be removed flag
- `uuid` (String) Unique identifier assigned by BCM

<a id="nestedatt--access_settings"></a>
### Nested Schema for `access_settings`


<a id="nestedatt--bios_setup"></a>
### Nested Schema for `bios_setup`


<a id="nestedatt--bmc_settings"></a>
### Nested Schema for `bmc_settings`

Optional:

- `firmware_manage_mode` (String) Firmware management mode (AUTO, MANUAL, DISABLED)
- `leak_policy` (String) Leak policy
- `leak_reaction_delay` (Number) Leak reaction delay in seconds
- `password` (String, Sensitive) BMC password (sensitive)
- `power_reset_delay` (Number) Power reset delay in seconds
- `privilege` (String) BMC privilege level (USER, OPERATOR, ADMINISTRATOR)
- `user_id` (Number) BMC user ID
- `user_name` (String) BMC username

Read-Only:

- `uuid` (String) Unique identifier


<a id="nestedatt--dpu_settings"></a>
### Nested Schema for `dpu_settings`


<a id="nestedatt--fsexports"></a>
### Nested Schema for `fsexports`

Required:

- `network` (String) Network UUID reference for export access
- `path` (String) Path to export (e.g., /home, /shared)

Optional:

- `all_squash` (Boolean) Map all uids and gids to the anonymous user (default: false)
- `allow_write` (Boolean) Allow write access (default: false)
- `anon_gid` (Number) Anonymous account group id number (default: 65534)
- `anon_uid` (Number) Anonymous account user id number (default: 65534)
- `async` (Boolean) Allow async NFS operations (default: true). When true, the NFS server can reply to requests before changes are committed to stable storage.
- `check_tree` (Boolean) Check tree (default: false)
- `disabled` (Boolean) Disable the export (default: false)
- `extra_options` (String) Extra NFS options to be added to this export
- `fsid` (Number) File system id for exports used in failover setup. Make sure these are identical for failover pairs.
- `hosts` (String) Extra hosts-range allowed access to this export (space separated)
- `name` (String) Export name (unique). If not specified, defaults to the path value.
- `rdma` (Boolean) Enable NFS over RDMA (default: false)
- `root_squash` (Boolean) Map requests from uid/gid 0 to the anonymous uid/gid (default: false)


<a id="nestedatt--fsmounts"></a>
### Nested Schema for `fsmounts`

Required:

- `device` (String) Device path or NFS export
- `filesystem` (String) Filesystem type
- `mountpoint` (String) Mount point path

Optional:

- `dump` (Boolean) Dump backup flag
- `fsck` (String) Filesystem check mode
- `mountoptions` (String) Mount options
- `rdma` (Boolean) Use RDMA for NFS

Read-Only:

- `uuid` (String) Unique identifier


<a id="nestedatt--gpu_settings"></a>
### Nested Schema for `gpu_settings`

Required:

- `child_type` (String) GPU vendor type: 'nvidia' or 'amd'.
- `name` (String) GPU range for which these settings apply (e.g., '0', '0-3', '0,1,2').

Optional:

- `activity_threshold` (Number) Activity threshold percentage 0-1 (AMD only).
- `clock_sync_boost_mode` (String) Clock sync boost mode among GPUs in group (Nvidia only).
- `compute_mode` (String) Compute mode (Nvidia only).
- `ecc_mode` (String) ECC mode: DISABLED, ENABLED, NONE (Nvidia only).
- `fan_speed` (Number) Fan speed value 0-255 (AMD only).
- `gpu_clock_level` (Number) GPU clock frequency level 0-7 (AMD only).
- `gpu_overdrive` (Number) GPU overdrive percentage 0-0.2 (AMD only).
- `hysteresis_down` (Number) Delay in seconds before clock level is decreased (AMD only).
- `hysteresis_up` (Number) Delay in seconds before clock level is increased (AMD only).
- `memory_clock_level` (Number) Memory clock frequency level 0-3 (AMD only).
- `memory_clock_speed` (Number) Memory clock speed in Hz (Nvidia only).
- `memory_overdrive` (Number) Memory overdrive percentage 0-0.2 (AMD only).
- `mig_profiles` (List of String) MIG profiles that will be applied to the GPU (Nvidia only).
- `minimal_gpu_clock` (Number) Minimum GPU clock speed in Hz (AMD only).
- `minimal_memory_clock` (Number) Minimum memory clock speed in Hz (AMD only).
- `multiprocessor_clock_speed` (Number) Streaming multiprocessor clock speed in Hz (Nvidia only).
- `power_limit` (Number) Power limit in Watts (Nvidia only).
- `power_play` (String) Power play mode (AMD only).
- `secondary_workload_power_profile` (String) Secondary workload power profile (Nvidia only).
- `workload_power_profile` (String) Workload power profile (Nvidia only).

Read-Only:

- `uuid` (String) BCM-assigned UUID for this GPU settings entry.


<a id="nestedatt--modules"></a>
### Nested Schema for `modules`

Required:

- `name` (String) Module name

Optional:

- `parameters` (String) Module parameters


<a id="nestedatt--proxy_settings"></a>
### Nested Schema for `proxy_settings`


<a id="nestedatt--roles"></a>
### Nested Schema for `roles`

Required:

- `child_type` (String) Role type (e.g., HeadNodeRole, StorageRole, BackupRole)
- `name` (String) Role name (e.g., headnode, storage, compute)

Optional:

- `add_services` (Boolean) Automatically add role services (default: false)

Read-Only:

- `uuid` (String) Role UUID (assigned by BCM)


<a id="nestedatt--selinux_settings"></a>
### Nested Schema for `selinux_settings`


<a id="nestedatt--services"></a>
### Nested Schema for `services`

Required:

- `name` (String) Service name (max 64 characters). Must be unique within the category.

Optional:

- `autostart` (Boolean) If true, CMDaemon will automatically restart a failed service.
- `managed` (Boolean) If true, manage config files from cmd (if any).
- `monitored` (Boolean) If true, CMDaemon will periodically check if the service is running.
- `run_if` (String) Condition for running the service. Common values: ALWAYS, NEVER. BCM validates additional states.
- `script_timeout` (Number) Service operation timeout in seconds. Use -1 for no timeout (default).
- `sickness_check_interval` (Number) Interval in seconds between sickness checks. Rounded up to 30-second monitoring intervals. Default: 60 seconds.
- `sickness_check_script` (String) Script path for sickness checking. The script is executed periodically to determine service health.
- `sickness_check_script_timeout` (Number) Timeout in seconds after which the sickness check script is killed. Default: 10 seconds.


<a id="nestedatt--software_image_proxy"></a>
### Nested Schema for `software_image_proxy`

Required:

- `parent_software_image` (String) Parent software image UUID reference

Read-Only:

- `revision_id` (Number) Revision identifier
- `uuid` (String) Unique identifier


<a id="nestedatt--static_routes"></a>
### Nested Schema for `static_routes`

Required:

- `gateway` (String) Gateway IPv4 address (e.g., '10.0.0.1')
- `ip` (String) Destination IP address (e.g., '0.0.0.0' for default route, '192.168.1.0' for specific network)
- `name` (String) Route name identifier (e.g., 'default-route', 'internal-net')
- `netmask_bits` (Number) Network mask in CIDR notation bits (0-32, e.g., 0 for default route, 24 for /24 subnet)
- `network` (String) Network UUID reference for this route (must be valid RFC 4122 UUID)

Optional:

- `metric` (Number) Route metric (priority, lower is preferred). Defaults to 0 if not specified.
- `network_device_name` (String) Specific network device name for this route (optional, leave empty for auto-selection)
- `notes` (String) User notes for this route

Read-Only:

- `uuid` (String) Unique identifier assigned by BCM


<a id="nestedatt--timezone_settings"></a>
### Nested Schema for `timezone_settings`


<a id="nestedatt--ztp_settings"></a>
### Nested Schema for `ztp_settings`

## Import

Import is supported using the following syntax:

```shell
#!/bin/bash
# Import an existing BCM category into Terraform state

# Import by UUID (recommended - UUIDs are stable identifiers)
terraform import bcm_cmdevice_category.example "0ae6d733-3015-4479-bfab-ce2d237a2809"

# After import, run terraform plan to verify no changes are detected
terraform plan

# Example workflow:
# 1. Create a category resource configuration in your .tf file
# 2. Run the import command with the category's UUID
# 3. Verify terraform plan shows no changes
# 4. Manage the category through Terraform going forward

# To find the UUID of a category:
# - Use BCM web UI: Navigate to CMDevice > Categories > Category Details
# - Or use BCM API: curl -k -X POST https://bcm-server:8081/json \
#   -H "Cookie: cm-login-token=<token>" \
#   -d '{"service":"cmdevice","call":"getCategories"}'
```

### Import with Data Source Lookup

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Import an existing BCM category into Terraform
# =============================================================================
# Use the categories data source to discover the UUID, then write a matching
# resource configuration before running terraform import.
#
# Step 1: terraform apply -target=data.bcm_cmdevice_categories.existing
# Step 2: terraform import bcm_cmdevice_category.existing <category-uuid>
# Step 3: terraform plan (verify no unexpected changes)

# Lookup the existing category to discover its current configuration
data "bcm_cmdevice_categories" "existing" {
  name = "gpu-compute-nodes"
}

# Lookup the management network for the category
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

# Write a resource configuration that matches the existing category
resource "bcm_cmdevice_category" "existing" {
  name               = data.bcm_cmdevice_categories.existing.categories[0].name
  management_network = data.bcm_cmnet_networks.management.networks[0].id
}
```

### Generate Configuration (Terraform 1.5+)

```terraform
# Copyright (c) HashiCorp, Inc.
# SPDX-License-Identifier: MPL-2.0

# =============================================================================
# Example: Generate Terraform configuration from existing BCM category
# =============================================================================
# Terraform 1.5+ supports generating HCL configuration from existing resources.
# This is useful for adopting existing infrastructure into Terraform management.
#
# Step 1: Add the import block below to your configuration
# Step 2: Run: terraform plan -generate-config-out=generated_category.tf
# Step 3: Review and adjust the generated configuration
# Step 4: Run: terraform plan (verify no unexpected changes)
# Step 5: Remove the import block and move the resource to your main config
#
# Equivalent CLI command:
#   terraform import bcm_cmdevice_category.example "0ae6d733-3015-4479-bfab-ce2d237a2809"

import {
  to = bcm_cmdevice_category.example
  id = "0ae6d733-3015-4479-bfab-ce2d237a2809"
}
```
