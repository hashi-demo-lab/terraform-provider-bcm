// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCMDeviceDevice_Idempotency(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-idempotent")
	categoryName := generateUniqueTestName("tftest-category-idempotent")
	imageName := generateUniqueTestName("tftest-image-idempotent")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device
			{
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("mac"),
						knownvalue.StringExact(mac),
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
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Another refresh to ensure stability
			{
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCMDeviceDevice_IdempotencyWithImport(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-import")
	categoryName := generateUniqueTestName("tftest-category-import")
	imageName := generateUniqueTestName("tftest-image-import")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device
			{
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Import the device
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"software_image_proxy",
					"boot_loader",          // BCM returns "CATEGORY" when not explicitly set
					"boot_loader_protocol", // BCM returns "CATEGORY" when not explicitly set
				},
			},
			// Step 3: Verify idempotency after import
			{
				Config: testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath, mac),
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
	deviceName := generateUniqueTestName("tftest-device-optional")
	categoryName := generateUniqueTestName("tftest-category-optional")
	imageName := generateUniqueTestName("tftest-image-optional")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with optional fields
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyOptional(deviceName, categoryName, imageName, imagePath, mac),
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
				Config: testAccCMDeviceDeviceConfigIdempotencyOptional(deviceName, categoryName, imageName, imagePath, mac),
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
	deviceName := generateUniqueTestName("tftest-device-update")
	categoryName := generateUniqueTestName("tftest-category-update")
	imageName := generateUniqueTestName("tftest-image-update")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with initial notes
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes", mac),
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
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes", mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Update device notes
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Updated notes", mac),
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
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Updated notes", mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCMDeviceDevice_IdempotencyDrift(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-drift")
	categoryName := generateUniqueTestName("tftest-category-drift")
	imageName := generateUniqueTestName("tftest-image-drift")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with initial notes
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes", mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Initial notes"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Modify notes externally to trigger drift detection
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get device UUID by name
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceName)
					if err != nil {
						t.Fatalf("Failed to get device: %v", err)
					}

					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to unmarshal device data: %v", err)
					}

					uuid, ok := deviceData["uuid"].(string)
					if !ok {
						t.Fatalf("Failed to extract device UUID")
					}

					// Modify notes field externally (snake_case → camelCase)
					deviceData["notes"] = "Modified externally"

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
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified notes externally to: %v", entity["notes"])
				},
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes", mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Step 3: Terraform restores desired state
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes", mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Initial notes"),
					),
				},
			},
			// Step 4: Verify idempotency after drift correction
			{
				Config: testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, "Initial notes", mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCMDeviceDeviceConfigIdempotency(deviceName, categoryName, imageName, imagePath, mac string) string {
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
  mac                = %[8]q
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
		mac,
	)
}

func testAccCMDeviceDeviceConfigIdempotencyOptional(deviceName, categoryName, imageName, imagePath, mac string) string {
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
  mac                = %[8]q
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
		mac,
	)
}

func testAccCMDeviceDeviceConfigIdempotencyUpdate(deviceName, categoryName, imageName, imagePath, notes, mac string) string {
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
  mac                = %[9]q
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
		mac,
	)
}
