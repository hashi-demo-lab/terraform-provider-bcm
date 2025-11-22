// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// testAccCMDeviceDevicePreCheck performs pre-test cleanup of leftover test devices.
func testAccCMDeviceDevicePreCheck(t *testing.T, deviceNames ...string) {
	client := createTestBCMClient(t)
	ctx := context.Background()

	for _, name := range deviceNames {
		// Try to get device UUID by name
		body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", name)
		if err != nil || len(body) == 0 {
			// Device doesn't exist, nothing to clean up
			continue
		}

		// Parse response to get UUID
		var deviceData map[string]interface{}
		if err := json.Unmarshal(body, &deviceData); err != nil {
			t.Logf("Warning: Could not parse device response for %s: %v", name, err)
			continue
		}

		uuid, ok := deviceData["uuid"].(string)
		if !ok {
			t.Logf("Warning: Could not extract UUID for device %s", name)
			continue
		}

		// Try to delete the device
		_, err = client.CallJSONRPC(ctx, "cmdevice", "removeDevice", uuid, true) // force=true
		if err != nil {
			t.Logf("Warning: Could not delete leftover device %s (UUID: %s): %v", name, uuid, err)
		}

		// Verify deletion with retries
		deleted := verifyResourceDeleted(ctx, client, "cmdevice", "getDevice", uuid, 5)
		if !deleted {
			t.Logf("Warning: Device %s (UUID: %s) still exists after cleanup attempt", name, uuid)
		}
	}
}

// testAccCheckCMDeviceDeviceDestroy verifies all devices are deleted after test.
func testAccCheckCMDeviceDeviceDestroy(s *terraform.State) error {
	client := createTestBCMClient(&testing.T{})
	ctx := context.Background()

	var errors []string
	resourceCount := 0

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmdevice_device" {
			continue
		}

		resourceCount++
		id := rs.Primary.ID

		// Verify device deleted with exponential backoff
		deleted := verifyResourceDeleted(
			ctx,
			client,
			"cmdevice",
			"getDevice",
			id,
			4, // retry count
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"Device still exists after destroy. Type: %s, ID: %s, Retries: 4",
				rs.Type,
				id,
			))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("CheckDestroy failures:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// testAccCMDeviceDeviceResourceConfig_Basic returns basic device configuration.
func testAccCMDeviceDeviceResourceConfig_Basic(hostname string, categoryName string, imageName string, imagePath string) string {
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
  path = %[5]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[7]q
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  depends_on = [bcm_cmdevice_category.test]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		hostname,
	)
}

// testAccCMDeviceDeviceResourceConfig_Updated returns updated device configuration.
func testAccCMDeviceDeviceResourceConfig_Updated(hostname string, categoryName string, imageName string, imagePath string) string {
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
  path = %[5]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[7]q
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id
  notes              = "Updated device notes"
  kernel_parameters  = "quiet splash"

  depends_on = [bcm_cmdevice_category.test]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		hostname,
	)
}

// TestAccCMDeviceDeviceResource_Basic tests full CRUD lifecycle.
func TestAccCMDeviceDeviceResource_Basic(t *testing.T) {
	deviceName := generateUniqueTestName("test-device")
	categoryName := generateUniqueTestName("citest-category-basic")
	imageName := generateUniqueTestName("citest-image-basic")
	imagePath := "/cm/images/ubuntu-22.04-server-amd64-basic.iso"

	// ID consistency tracking across all CRUD operations
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", deviceName),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_device.test", "uuid"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_device.test", "id"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "mac", "00:11:22:33:44:55"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("mac"),
						knownvalue.StringExact("00:11:22:33:44:55"),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Import testing
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",                  // Write-only field
					"management_network",     // BCM resets to nil UUID
					"boot_loader",            // BCM returns "CATEGORY" when inheriting from category
					"boot_loader_protocol",   // BCM returns "CATEGORY" when inheriting from category
					"partition",              // BCM may populate from category default
					"power_control",          // BCM returns default "none" when not set
					"default_gateway",        // BCM returns default "0.0.0.0" when not set
					"default_gateway_metric", // BCM returns default 0 when not set
					"serial_number",          // BCM may populate from hardware discovery
					"part_number",            // BCM may populate from hardware discovery
				},
			},
			// Verify ID consistency after Import
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath),
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName, categoryName, imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", deviceName),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "Updated device notes"),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "kernel_parameters", "quiet splash"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Updated device notes"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kernel_parameters"),
						knownvalue.StringExact("quiet splash"),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName, categoryName, imageName, imagePath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// testAccCMDeviceDeviceResourceConfig_Drift returns config for drift detection tests.
func testAccCMDeviceDeviceResourceConfig_Drift(hostname string, categoryName string, imageName string, imagePath string) string {
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
  path = %[5]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[7]q
  mac                = "00:11:22:33:44:66"
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id
  notes              = "initial-notes"

  depends_on = [bcm_cmdevice_category.test]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		hostname,
	)
}

// TestAccCMDeviceDevice_DriftNotes tests drift detection for notes field.
func TestAccCMDeviceDevice_DriftNotes(t *testing.T) {
	deviceName := generateUniqueTestName("test-device-drift")
	categoryName := generateUniqueTestName("citest-category-drift")
	imageName := generateUniqueTestName("citest-image-drift")
	imagePath := "/cm/images/ubuntu-22.04-server-amd64-drift-notes.iso"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create with initial value
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "initial-notes"),
				),
			},
			// Modify externally via BCM API
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get device by hostname
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceName)
					if err != nil {
						t.Fatalf("Failed to get device for drift test: %v", err)
					}

					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					uuid, _ := deviceData["uuid"].(string)

					// Modify notes field externally
					deviceData["notes"] = "externally-modified"

					// Build BCM entity structure
					entity := map[string]interface{}{
						"baseType":      "Device",
						"childType":     deviceData["childType"],
						"modified":      true,
						"to_be_removed": false,
						"uuid":          uuid,
					}
					for k, v := range deviceData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", entity, false)
					if err != nil {
						t.Fatalf("Failed to update device externally: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified notes externally to: externally-modified")
				},
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Terraform restores desired state
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "initial-notes"),
				),
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationInvalidHostname tests hostname validation.
func TestAccCMDeviceDevice_ValidationInvalidHostname(t *testing.T) {
	categoryName := generateUniqueTestName("citest-category-validation")
	imageName := generateUniqueTestName("citest-image-validation")
	imagePath := "/cm/images/ubuntu-22.04-server-amd64-validation.iso"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("UPPERCASE", categoryName, imageName, imagePath),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("-leadinghyphen", categoryName, imageName, imagePath),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("trailing-hyphen-", categoryName, imageName, imagePath),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
		},
	})
}

// testAccCMDeviceDeviceResourceConfig_InvalidMAC returns config with invalid MAC address.
func testAccCMDeviceDeviceResourceConfig_InvalidMAC(hostname, mac string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
locals {
  category_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].uuid : ""
}

data "bcm_cmnet_networks" "all" {}
locals {
  network_uuid = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].uuid : ""
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = %[5]q
  category           = local.category_uuid
  management_network = local.network_uuid
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		mac,
	)
}

// TestAccCMDeviceDevice_ValidationInvalidMAC tests MAC address validation.
func TestAccCMDeviceDevice_ValidationInvalidMAC(t *testing.T) {
	deviceName := generateUniqueTestName("test-device")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMDeviceDeviceResourceConfig_InvalidMAC(deviceName, "00-11-22-33-44-55"),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_InvalidMAC(deviceName, "00:11:22:33:44"),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_InvalidMAC(deviceName, "ZZ:11:22:33:44:55"),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

// ========================================
// Phase 3: Drift Detection Tests
// ========================================

// TestAccCMDeviceDevice_Drift tests drift detection for hostname attribute
// This test verifies the provider correctly detects when a device's hostname is modified externally via BCM API.
func TestAccCMDeviceDevice_Drift(t *testing.T) {
	initialHostname := generateUniqueTestName("test-device-drift")
	driftedHostname := generateUniqueTestName("drifted-device")
	categoryName := generateUniqueTestName("citest-category-drift-hostname")
	imageName := generateUniqueTestName("citest-image-drift-hostname")
	imagePath := "/cm/images/ubuntu-22.04-server-amd64-drift-hostname.iso"

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, initialHostname, driftedHostname)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with initial hostname
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", initialHostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "initial-notes"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_device.test", "uuid"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(initialHostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("initial-notes"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Modify hostname externally via BCM API, verify drift detected
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get UUID by device hostname using helper
					uuid := getResourceUUIDByName(t, "cmdevice", "getDevice", initialHostname)

					// Fetch full device data from BCM API
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", uuid)
					if err != nil {
						t.Fatalf("Failed to fetch device for drift modification: %v", err)
					}

					// Parse the device data
					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					// Modify hostname field (Terraform snake_case -> BCM API camelCase)
					deviceData["hostname"] = driftedHostname

					// Wrap in BCM API entity structure required for updates
					entity := map[string]interface{}{
						"baseType":      "Device",
						"childType":     deviceData["childType"],
						"modified":      true,
						"to_be_removed": false,
						"revision":      "",
						"uuid":          uuid,
					}
					// Copy all device data fields except uuid (already set above)
					for k, v := range deviceData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", entity, false)
					if err != nil {
						t.Fatalf("Failed to update device via BCM API: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified hostname externally to: %v", entity["hostname"])
				},
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath),
				// Use ConfigPlanChecks to verify drift detected (non-empty plan expected)
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Step 3: Restore desired state (Terraform applies config to fix drift)
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify drift was corrected and state matches config
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", initialHostname),
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "initial-notes"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(initialHostname),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("initial-notes"),
					),
				},
			},
		},
	})
}
