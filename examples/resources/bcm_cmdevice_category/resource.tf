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
    <disksetup>
      <disk device="/dev/sda">
        <partition number="1" size="512M" type="ef00" label="EFI"/>
        <partition number="2" size="100G" type="8300" label="root"/>
        <partition number="3" size="remaining" type="8300" label="data"/>
      </disk>
    </disksetup>
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
