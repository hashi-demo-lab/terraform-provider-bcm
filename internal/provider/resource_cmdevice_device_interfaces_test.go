// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// testAccCMDeviceDeviceConfigInterfaceSingle generates a config for a device with a single interface.
func testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[4]q
  mac      = %[5]q
  category = %[6]q

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[5]q
    network  = %[7]q
    bootable = true
    dhcp     = true
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		mac,
		categoryUUID,
		networkUUID,
	)
}

// testAccCMDeviceDeviceConfigInterfaceMultiple generates a config for a device with multiple interfaces.
func testAccCMDeviceDeviceConfigInterfaceMultiple(hostname, mac1, mac2, categoryUUID, networkUUID1, networkUUID2 string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[4]q
  mac      = %[5]q
  category = %[7]q

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[5]q
    network  = %[8]q
    bootable = true
    dhcp     = true
  }

  interfaces {
    name     = "eth1"
    type     = "physical"
    mac      = %[6]q
    network  = %[9]q
    bootable = false
    dhcp     = true
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		mac1,
		mac2,
		categoryUUID,
		networkUUID1,
		networkUUID2,
	)
}

// testAccCMDeviceDeviceConfigInterfaceBond generates a config for a device with a bond interface.
func testAccCMDeviceDeviceConfigInterfaceBond(hostname, mac, categoryUUID, networkUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[4]q
  mac      = %[5]q
  category = %[6]q

  interfaces {
    name      = "bond0"
    type      = "bond"
    members   = ["eth0", "eth1"]
    bond_mode = "802.3ad"
    network   = %[7]q
    bootable  = true
    dhcp      = true
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		mac,
		categoryUUID,
		networkUUID,
	)
}

// testAccCMDeviceDeviceConfigInterfaceBMC generates a config for a device with a BMC interface.
// Note: BMC interface names must follow the pattern ipmiX, iloX, cimcX, dracX, or rfX where X is a number.
func testAccCMDeviceDeviceConfigInterfaceBMC(hostname, mac, categoryUUID, networkUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[4]q
  mac      = %[5]q
  category = %[6]q

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[5]q
    network  = %[7]q
    bootable = true
    dhcp     = true
  }

  interfaces {
    name    = "ipmi0"
    type    = "bmc"
    network = %[7]q
    dhcp    = true
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		mac,
		categoryUUID,
		networkUUID,
	)
}

// TestAccCMDeviceDevice_InterfaceSingle tests creating a device with a single physical interface.
// RED Phase: This test is expected to fail until implementation is complete.
func TestAccCMDeviceDevice_InterfaceSingle(t *testing.T) {
	hostname := generateShortTestName("dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with single interface
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", hostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.name", "eth0"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.type", "physical"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.mac", mac),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.bootable", "true"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_device.test", "interfaces.0.uuid"),
				),
			},
			// Import test
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"management_network", // BCM may not return this
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceMultiple tests creating a device with multiple physical interfaces.
// RED Phase: This test is expected to fail until implementation is complete.
func TestAccCMDeviceDevice_InterfaceMultiple(t *testing.T) {
	hostname := generateShortTestName("dev")
	mac1 := generateUniqueMAC()
	mac2 := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID1 := os.Getenv("BCM_TEST_NETWORK_UUID")
	networkUUID2 := os.Getenv("BCM_TEST_NETWORK_UUID_2")

	if categoryUUID == "" || networkUUID1 == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// Use same network if second not provided
	if networkUUID2 == "" {
		networkUUID2 = networkUUID1
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with multiple interfaces
			{
				Config: testAccCMDeviceDeviceConfigInterfaceMultiple(hostname, mac1, mac2, categoryUUID, networkUUID1, networkUUID2),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", hostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.name", "eth0"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.type", "physical"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.bootable", "true"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.1.name", "eth1"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.1.type", "physical"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.1.bootable", "false"),
				),
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceBond tests creating a device with a bond interface.
// RED Phase: This test is expected to fail until implementation is complete.
func TestAccCMDeviceDevice_InterfaceBond(t *testing.T) {
	hostname := generateShortTestName("dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with bond interface
			{
				Config: testAccCMDeviceDeviceConfigInterfaceBond(hostname, mac, categoryUUID, networkUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", hostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.#", "1"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.name", "bond0"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.type", "bond"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.bond_mode", "802.3ad"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.members.#", "2"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.members.0", "eth0"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.members.1", "eth1"),
				),
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceBMC tests creating a device with a BMC interface.
// RED Phase: This test is expected to fail until implementation is complete.
func TestAccCMDeviceDevice_InterfaceBMC(t *testing.T) {
	hostname := generateShortTestName("dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with BMC interface
			{
				Config: testAccCMDeviceDeviceConfigInterfaceBMC(hostname, mac, categoryUUID, networkUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", hostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.#", "2"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.name", "eth0"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.type", "physical"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.1.name", "ipmi0"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.1.type", "bmc"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.1.dhcp", "true"),
				),
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceImport tests importing a device with interfaces.
// RED Phase: This test is expected to fail until implementation is complete.
func TestAccCMDeviceDevice_InterfaceImport(t *testing.T) {
	hostname := generateShortTestName("dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create device
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", hostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.#", "1"),
				),
			},
			// Import and verify interfaces are populated
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"management_network", // BCM may not return this
					"default_gateway",    // BCM returns "0.0.0.0" default
					"power_control",      // BCM returns "none" default
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceUpdate tests updating interface configurations.
// RED Phase: This test is expected to fail until implementation is complete.
func TestAccCMDeviceDevice_InterfaceUpdate(t *testing.T) {
	hostname := generateShortTestName("dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with bootable=true
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.bootable", "true"),
				),
			},
			// Update to bootable=false (using modified config)
			{
				Config: testAccCMDeviceDeviceConfigInterfaceUpdateBootable(hostname, mac, categoryUUID, networkUUID, false),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "interfaces.0.bootable", "false"),
				),
			},
		},
	})
}

// testAccCMDeviceDeviceConfigInterfaceUpdateBootable generates a config with configurable bootable flag.
func testAccCMDeviceDeviceConfigInterfaceUpdateBootable(hostname, mac, categoryUUID, networkUUID string, bootable bool) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[4]q
  mac      = %[5]q
  category = %[6]q

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[5]q
    network  = %[7]q
    bootable = %[8]t
    dhcp     = true
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		mac,
		categoryUUID,
		networkUUID,
		bootable,
	)
}

// TestAccCMDeviceDevice_LegacyMACOnly tests backward compatibility with existing configurations.
// This test ensures that configurations without the interfaces block still work.
func TestAccCMDeviceDevice_LegacyMACOnly(t *testing.T) {
	hostname := generateShortTestName("dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Create with legacy mac-only configuration (no interfaces block)
			{
				Config: testAccCMDeviceDeviceConfigLegacy(hostname, mac, categoryUUID, networkUUID),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", hostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "mac", mac),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_device.test", "uuid"),
				),
			},
		},
	})
}

// testAccCMDeviceDeviceConfigLegacy generates a config without interfaces block (legacy mode).
func testAccCMDeviceDeviceConfigLegacy(hostname, mac, categoryUUID, networkUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = %[5]q
  category           = %[6]q
  management_network = %[7]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		mac,
		categoryUUID,
		networkUUID,
	)
}
