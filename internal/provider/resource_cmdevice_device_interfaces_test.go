// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
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
func TestAccCMDeviceDevice_InterfaceSingle(t *testing.T) {
	hostname := generateShortTestName("tftest-dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with single interface
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("eth0"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("type"),
						knownvalue.StringExact("physical"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("mac"),
						knownvalue.StringExact(mac),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("bootable"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency check - no changes expected
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Import test
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
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceMultiple tests creating a device with multiple physical interfaces.
func TestAccCMDeviceDevice_InterfaceMultiple(t *testing.T) {
	hostname := generateShortTestName("tftest-dev")
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

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with multiple interfaces
			{
				Config: testAccCMDeviceDeviceConfigInterfaceMultiple(hostname, mac1, mac2, categoryUUID, networkUUID1, networkUUID2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("eth0"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("type"),
						knownvalue.StringExact("physical"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("bootable"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(1).AtMapKey("name"),
						knownvalue.StringExact("eth1"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(1).AtMapKey("type"),
						knownvalue.StringExact("physical"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(1).AtMapKey("bootable"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigInterfaceMultiple(hostname, mac1, mac2, categoryUUID, networkUUID1, networkUUID2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceBond tests creating a device with a bond interface.
func TestAccCMDeviceDevice_InterfaceBond(t *testing.T) {
	hostname := generateShortTestName("tftest-dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with bond interface
			{
				Config: testAccCMDeviceDeviceConfigInterfaceBond(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("bond0"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("type"),
						knownvalue.StringExact("bond"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("bond_mode"),
						knownvalue.StringExact("802.3ad"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("members"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigInterfaceBond(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceBMC tests creating a device with a BMC interface.
// SKIPPED: BCM may not support BMC interfaces via API or requires specific hardware.
// This test can be enabled when testing against BCM clusters with BMC support.
func TestAccCMDeviceDevice_InterfaceBMC(t *testing.T) {
	t.Skip("SKIPPED: BCM cluster may not support BMC interfaces via API")
	hostname := generateShortTestName("tftest-dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with BMC interface
			{
				Config: testAccCMDeviceDeviceConfigInterfaceBMC(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("eth0"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("type"),
						knownvalue.StringExact("physical"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(1).AtMapKey("name"),
						knownvalue.StringExact("ipmi0"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(1).AtMapKey("type"),
						knownvalue.StringExact("bmc"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(1).AtMapKey("dhcp"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigInterfaceBMC(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceImport tests importing a device with interfaces.
func TestAccCMDeviceDevice_InterfaceImport(t *testing.T) {
	hostname := generateShortTestName("tftest-dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Import and verify interfaces are populated
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
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Idempotency after import
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_InterfaceUpdate tests updating interface configurations.
func TestAccCMDeviceDevice_InterfaceUpdate(t *testing.T) {
	hostname := generateShortTestName("tftest-dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with bootable=true
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("bootable"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency after creation
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Update to bootable=false
			{
				Config: testAccCMDeviceDeviceConfigInterfaceUpdateBootable(hostname, mac, categoryUUID, networkUUID, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("bootable"),
						knownvalue.Bool(false),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 4: Idempotency after update
			{
				Config: testAccCMDeviceDeviceConfigInterfaceUpdateBootable(hostname, mac, categoryUUID, networkUUID, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
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
	hostname := generateShortTestName("tftest-dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with legacy mac-only configuration (no interfaces block)
			{
				Config: testAccCMDeviceDeviceConfigLegacy(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("mac"),
						knownvalue.StringExact(mac),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigLegacy(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
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

// TestAccCMDeviceDevice_InterfaceDrift tests drift detection for interface configurations.
// This test verifies that Terraform detects external changes to interface bootable flag.
func TestAccCMDeviceDevice_InterfaceDrift(t *testing.T) {
	hostname := generateShortTestName("tftest-dev")
	mac := generateUniqueMAC()

	// Get test data from environment
	categoryUUID := os.Getenv("BCM_TEST_CATEGORY_UUID")
	networkUUID := os.Getenv("BCM_TEST_NETWORK_UUID")

	if categoryUUID == "" || networkUUID == "" {
		t.Skip("BCM_TEST_CATEGORY_UUID and BCM_TEST_NETWORK_UUID environment variables must be set")
	}

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with bootable=true
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(hostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("bootable"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Modify interface externally via BCM API (drift)
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get device by hostname
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", hostname)
					if err != nil {
						t.Fatalf("Failed to get device for drift test: %v", err)
					}

					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					uuid, ok := deviceData["uuid"].(string)
					if !ok {
						t.Fatalf("Failed to extract device UUID")
					}

					// Modify the interface bootable flag externally
					// BCM stores interfaces as "interfaces" array
					interfaces, ok := deviceData["interfaces"].([]interface{})
					if !ok || len(interfaces) == 0 {
						t.Fatalf("No interfaces found on device")
					}

					// Modify the first interface's bootable flag
					firstInterface, ok := interfaces[0].(map[string]interface{})
					if !ok {
						t.Fatalf("Failed to parse interface data")
					}
					firstInterface["bootable"] = false

					// Wrap in BCM API entity structure
					entity := map[string]interface{}{
						"baseType":      "Node",
						"childType":     "",
						"modified":      true,
						"to_be_removed": false,
						"revision":      "",
						"uuid":          uuid,
					}

					// Copy all fields except uuid (already set)
					for k, v := range deviceData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", entity, false)
					if err != nil {
						t.Fatalf("Failed to update device: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(TestEventualConsistencyDelay)

					t.Logf("[DEBUG] Modified interface bootable externally to: false")
				},
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Terraform restores desired state
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("interfaces").AtSliceIndex(0).AtMapKey("bootable"),
						knownvalue.Bool(true),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 4: Verify idempotency after drift correction
			{
				Config: testAccCMDeviceDeviceConfigInterfaceSingle(hostname, mac, categoryUUID, networkUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}
