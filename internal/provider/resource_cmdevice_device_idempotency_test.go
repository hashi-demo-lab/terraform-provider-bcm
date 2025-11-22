// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCMDeviceDevice_Idempotency(t *testing.T) {
	deviceName := generateUniqueTestName("citest-device-idempotent")
	categoryName := generateUniqueTestName("citest-category-idempotent")
	imageName := generateUniqueTestName("citest-image-idempotent")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device
			{
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("mac"),
						knownvalue.StringExact("00:11:22:33:44:BB"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("category"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Verify idempotency - no changes expected
			{
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Another refresh to ensure stability
			{
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCMDeviceDevice_IdempotencyWithOptionalFields(t *testing.T) {
	deviceName := generateUniqueTestName("citest-device-optional")
	categoryName := generateUniqueTestName("citest-category-optional")
	imageName := generateUniqueTestName("citest-image-optional")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with optional fields
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyOptional(deviceName, categoryName, imageName, imagePath),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Idempotency test device"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kernel_parameters"),
						knownvalue.StringExact("console=ttyS0"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Verify idempotency with optional fields
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyOptional(deviceName, categoryName, imageName, imagePath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCMDeviceDevice_IdempotencyAfterUpdate(t *testing.T) {
	deviceName := generateUniqueTestName("citest-device-update")
	categoryName := generateUniqueTestName("citest-category-update")
	imageName := generateUniqueTestName("citest-image-update")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with initial notes
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Initial notes"),
					),
				},
			},
			// Step 2: Verify idempotency after creation
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Update device notes
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Updated notes"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Updated notes"),
					),
				},
			},
			// Step 4: Verify idempotency after update
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Updated notes"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[7]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[5]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[6]q
  mac                = "00:11:22:33:44:BB"
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  depends_on = [bcm_cmdevice_category.test]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		categoryName,
		deviceName,
		imagePath,
	)
}

func testAccCMDeviceDeviceConfigIdempotencyOptional(deviceName, categoryName, imageName, imagePath string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[7]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[5]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[6]q
  mac                = "00:11:22:33:44:CC"
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id
  notes              = "Idempotency test device"
  kernel_parameters  = "console=ttyS0"

  depends_on = [bcm_cmdevice_category.test]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		categoryName,
		deviceName,
		imagePath,
	)
}

func testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[8]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[5]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[6]q
  mac                = "00:11:22:33:44:DD"
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id
  notes              = %[7]q

  depends_on = [bcm_cmdevice_category.test]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		categoryName,
		deviceName,
		notes,
		imagePath,
	)
}
