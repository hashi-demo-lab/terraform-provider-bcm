


# Lookup existing category and network
data "bcm_cmdevice_categories" "default" {
  name = "default"
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "DefaultEthernet"
  }
}

# =============================================================================
# Example 1: Import with known values
# =============================================================================
# When you know the device's MAC and network, specify them directly.
#
# Step 1: terraform import bcm_cmdevice_device.known <device-uuid>
# Step 2: terraform plan (verify no unexpected changes)

resource "bcm_cmdevice_device" "known" {
  hostname = "citest-import-known"
  category = one(data.bcm_cmdevice_categories.default.categories[*].id)

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "00:11:22:33:44:CC"
    network  = one(data.bcm_cmnet_networks.management.networks[*].id)
    bootable = true
    dhcp     = true
  }
}

# =============================================================================
# Example 2: Import using data source lookup
# =============================================================================
# When importing an existing device, use the nodes data source to discover
# the device's current MAC and interface configuration. This avoids
# hardcoding values that may differ from what BCM has.
#
# Step 1: terraform apply -target=data.bcm_cmdevice_nodes.existing
# Step 2: terraform import bcm_cmdevice_device.discovered <device-uuid>
# Step 3: terraform plan (verify no unexpected changes)

data "bcm_cmdevice_nodes" "existing" {
  filter {
    hostname_pattern = "existing-server-01"
  }
}

resource "bcm_cmdevice_device" "discovered" {
  hostname = data.bcm_cmdevice_nodes.existing.nodes[0].hostname
  category = one(data.bcm_cmdevice_categories.default.categories[*].id)

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = data.bcm_cmdevice_nodes.existing.nodes[0].mac
    network  = one(data.bcm_cmnet_networks.management.networks[*].id)
    bootable = true
    dhcp     = true
  }
}
