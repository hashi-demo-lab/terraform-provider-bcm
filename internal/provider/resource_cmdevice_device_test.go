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

// testAccCMDeviceDevicePreCheck performs pre-test cleanup of leftover test devices
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
		deleted, _ := verifyResourceDeleted(ctx, client, "cmdevice", "getDevice", uuid, 5)
		if !deleted {
			t.Logf("Warning: Device %s (UUID: %s) still exists after cleanup attempt", name, uuid)
		}
	}
}

// testAccCheckCMDeviceDeviceDestroy verifies all devices are deleted after test
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
		deleted, err := verifyResourceDeleted(
			ctx,
			client,
			"cmdevice",
			"getDevice",
			id,
			4, // retry count
		)

		if err != nil {
			errors = append(errors, fmt.Sprintf(
				"Resource type: %s, ID: %s, Error: %v",
				rs.Type,
				id,
				err,
			))
		}

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

// testAccCMDeviceDeviceResourceConfig_Basic returns basic device configuration
func testAccCMDeviceDeviceResourceConfig_Basic(hostname string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Get default category
data "bcm_cmdevice_categories" "all" {}

# Use first available category
locals {
  category_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].uuid : ""
}

# Get management network
data "bcm_cmnet_networks" "all" {}

# Use first available network
locals {
  network_uuid = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].uuid : ""
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = "00:11:22:33:44:55"
  category           = local.category_uuid
  management_network = local.network_uuid
  partition          = "ddd19eb5-f04a-48dc-9cc6-f160b704a7dd"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
	)
}

// testAccCMDeviceDeviceResourceConfig_Updated returns updated device configuration
func testAccCMDeviceDeviceResourceConfig_Updated(hostname string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Get default category
data "bcm_cmdevice_categories" "all" {}

locals {
  category_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].uuid : ""
}

# Get management network
data "bcm_cmnet_networks" "all" {}

locals {
  network_uuid = length(data.bcm_cmnet_networks.all.networks) > 0 ? data.bcm_cmnet_networks.all.networks[0].uuid : ""
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[4]q
  mac                = "00:11:22:33:44:55"
  category           = local.category_uuid
  management_network = local.network_uuid
  partition          = "ddd19eb5-f04a-48dc-9cc6-f160b704a7dd"
  notes              = "Updated device notes"
  kernel_parameters  = "quiet splash"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
	)
}

// TestAccCMDeviceDeviceResource_Basic tests full CRUD lifecycle
func TestAccCMDeviceDeviceResource_Basic(t *testing.T) {
	deviceName := generateUniqueTestName("test-device")

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
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName),
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
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName),
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
					"force",                // Write-only field
					"management_network",   // BCM resets to nil UUID
					"boot_loader",          // BCM returns "CATEGORY" when inheriting from category
					"boot_loader_protocol", // BCM returns "CATEGORY" when inheriting from category
					"partition",            // BCM may populate from category default
				},
			},
			// Verify ID consistency after Import
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName),
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update and Read testing
			{
				Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName),
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
				Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// testAccCMDeviceDeviceResourceConfig_Drift returns config for drift detection tests
func testAccCMDeviceDeviceResourceConfig_Drift(hostname, notesValue string) string {
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
  mac                = "00:11:22:33:44:66"
  category           = local.category_uuid
  management_network = local.network_uuid
  partition          = "ddd19eb5-f04a-48dc-9cc6-f160b704a7dd"
  notes              = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		hostname,
		notesValue,
	)
}

// TestAccCMDeviceDevice_DriftNotes tests drift detection for notes field
func TestAccCMDeviceDevice_DriftNotes(t *testing.T) {
	deviceName := generateUniqueTestName("test-device-drift")

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
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, "initial-notes"),
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
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, "initial-notes"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Terraform restores desired state
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, "initial-notes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "notes", "initial-notes"),
				),
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationInvalidHostname tests hostname validation
func TestAccCMDeviceDevice_ValidationInvalidHostname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("UPPERCASE"),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("-leadinghyphen"),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("trailing-hyphen-"),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
		},
	})
}

// testAccCMDeviceDeviceResourceConfig_InvalidMAC returns config with invalid MAC address
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

// TestAccCMDeviceDevice_ValidationInvalidMAC tests MAC address validation
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
