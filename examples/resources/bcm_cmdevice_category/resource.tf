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
