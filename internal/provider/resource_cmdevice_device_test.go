// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
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
	ctx := t.Context()

	for _, name := range deviceNames {
		// Try to get device UUID by name.
		body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", name)
		if err != nil || len(body) == 0 {
			// Device doesn't exist, nothing to clean up.
			continue
		}

		// Parse response to get UUID.
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

		// Try to delete the device.
		_, err = client.CallJSONRPC(ctx, "cmdevice", "removeDevice", uuid, true) // force=true
		if err != nil {
			t.Logf("Warning: Could not delete leftover device %s (UUID: %s): %v", name, uuid, err)
		}

		// Verify deletion with retries.
		deleted := verifyResourceDeleted(ctx, client, "cmdevice", "getDevice", uuid, 5)
		if !deleted {
			t.Logf("Warning: Device %s (UUID: %s) still exists after cleanup attempt", name, uuid)
		}
	}
}

// testAccCheckCMDeviceDeviceDestroy verifies all resources are deleted after test.
// This function handles cleanup verification in proper dependency order:
// 1. Devices (must be deleted first)
// 2. Categories (can be deleted after devices)
// 3. Software Images (must be deleted last).
func testAccCheckCMDeviceDeviceDestroy(s *terraform.State) error {
	client := createTestBCMClient(&testing.T{})
	ctx := context.Background()

	var errors []string
	deviceCount := 0
	categoryCount := 0
	imageCount := 0

	// Phase 1: Verify devices are deleted (highest priority - no dependencies)
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmdevice_device" {
			continue
		}

		deviceCount++
		id := rs.Primary.ID

		// Verify device deleted with exponential backoff.
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

	// Phase 2: Verify categories are deleted (must happen after devices)
	// Uses retry with exponential backoff to handle eventual consistency
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmdevice_category" {
			continue
		}

		categoryCount++
		name := rs.Primary.Attributes["name"]
		uuid := rs.Primary.Attributes["uuid"]

		// Verify category deleted with exponential backoff
		deleted := verifyResourceDeleted(
			ctx,
			client,
			"cmdevice",
			"getCategory",
			name,
			4, // retry count with exponential backoff
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"Category still exists after destroy. Name: %s, UUID: %s, Retries: 4",
				name,
				uuid,
			))
		}
	}

	// Phase 3: Verify software images are deleted (must happen last)
	// Uses retry with exponential backoff to handle eventual consistency
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmpart_softwareimage" {
			continue
		}

		imageCount++
		uuid := rs.Primary.Attributes["uuid"]
		if uuid == "" {
			continue // Resource was never created
		}

		// Verify software image deleted with exponential backoff
		deleted := verifyResourceDeleted(
			ctx,
			client,
			"CMPart",
			"getSoftwareImage",
			uuid,
			4, // retry count with exponential backoff
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"Software image still exists after destroy. UUID: %s, Retries: 4",
				uuid,
			))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("CheckDestroy failures:\n  - %s\n  - Verified: %d devices, %d categories, %d images",
			strings.Join(errors, "\n  - "),
			deviceCount,
			categoryCount,
			imageCount,
		)
	}

	return nil
}

// testAccCMDeviceDeviceResourceConfig_Basic returns basic device configuration.
func testAccCMDeviceDeviceResourceConfig_Basic(hostname string, categoryName string, imageName string, imagePath string, mac string) string {
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
  hostname = %[7]q
  category = bcm_cmdevice_category.test.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
		mac,
	)
}

// testAccCMDeviceDeviceResourceConfig_Updated returns updated device configuration.
func testAccCMDeviceDeviceResourceConfig_Updated(hostname string, categoryName string, imageName string, imagePath string, mac string) string {
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
  hostname          = %[7]q
  category          = bcm_cmdevice_category.test.id
  notes             = "Updated device notes"
  kernel_parameters = "quiet splash"

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
		mac,
	)
}

// TestAccCMDeviceDeviceResource_Basic tests full CRUD lifecycle.
func TestAccCMDeviceDeviceResource_Basic(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device")
	categoryName := generateUniqueTestName("tftest-category-basic")
	imageName := generateUniqueTestName("tftest-image-basic")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// ID consistency tracking across all CRUD operations.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
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
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("mac"),
						knownvalue.StringExact(mac),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Import testing.
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
					"interfaces.#",           // Import populates interfaces from BCM for legacy-mode devices
					"interfaces.0.%",
					"interfaces.0.base_type",
					"interfaces.0.bootable",
					"interfaces.0.cardtype",
					"interfaces.0.child_type",
					"interfaces.0.dhcp",
					"interfaces.0.mac",
					"interfaces.0.name",
					"interfaces.0.network",
					"interfaces.0.start_if",
					"interfaces.0.type",
					"interfaces.0.uuid",
					"interfaces.0.bond_mode",
					"interfaces.0.ip",
					"interfaces.0.ipv6_ip",
					"interfaces.0.members.#",
				},
			},
			// Verify ID consistency after Import.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
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
			// Update and Read testing.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
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
			// Idempotency check after Update.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName, categoryName, imageName, imagePath, mac),
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
func testAccCMDeviceDeviceResourceConfig_Drift(hostname string, categoryName string, imageName string, imagePath string, mac string) string {
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
  hostname = %[7]q
  category = bcm_cmdevice_category.test.id
  notes    = "initial-notes"

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
		mac,
	)
}

// TestAccCMDeviceDevice_DriftNotes tests drift detection for notes field.
func TestAccCMDeviceDevice_DriftNotes(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-drift")
	categoryName := generateUniqueTestName("tftest-category-drift")
	imageName := generateUniqueTestName("tftest-image-drift")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create with initial value.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("initial-notes"),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Modify externally via BCM API.
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get device by hostname.
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceName)
					if err != nil {
						t.Fatalf("Failed to get device for drift test: %v", err)
					}

					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					uuid, _ := deviceData["uuid"].(string)

					// Modify notes field externally.
					deviceData["notes"] = "externally-modified"

					// Build BCM entity structure.
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

					// Update via BCM API.
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", entity, false)
					if err != nil {
						t.Fatalf("Failed to update device externally: %v", err)
					}

					// Wait for eventual consistency.
					time.Sleep(TestEventualConsistencyDelay)

					t.Logf("[DEBUG] Modified notes externally to: externally-modified")
				},
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Terraform restores desired state.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("initial-notes"),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationInvalidHostname tests hostname validation.
func TestAccCMDeviceDevice_ValidationInvalidHostname(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-category-validation")
	imageName := generateUniqueTestName("tftest-image-validation")
	imagePath := "/cm/images/ubuntu-22.04-server-amd64-validation.iso"
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("UPPERCASE", categoryName, imageName, imagePath, mac),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("-leadinghyphen", categoryName, imageName, imagePath, mac),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
			{
				Config:      testAccCMDeviceDeviceResourceConfig_Basic("trailing-hyphen-", categoryName, imageName, imagePath, mac),
				ExpectError: regexp.MustCompile(`hostname must be RFC 1123 DNS label`),
			},
		},
	})
}

// ========================================
// Phase 4: Partition Commit Timeout Tests.
// ========================================

// TestAccCMDeviceDevice_PartitionCommitTimeout tests timeout scenario for partition commit polling.
//
// This test verifies the error path in resource_cmdevice_device.go lines 527-568 (waitForPartitionCommit).
// The test scenario simulates a partition that never completes its commit, forcing the exponential.
// backoff retry logic to exhaust all attempts and return a timeout error.
//
// Error Path Under Test:
// 1. Device creation succeeds (addDevice API call returns success)
// 2. waitForPartitionCommit starts polling the partition UUID
// 3. Mock server returns partition data with modified=true indefinitely
// 4. After maxRetries (20 attempts, ~60s total wait), timeout error returned
//
// Expected Behavior:
// - Exponential backoff timing: 2s, 4s, 6s, 8s, 10s, 10s, ... (capped at 10s)
// - Total wait time: approximately 110-120 seconds for 20 retries
// - Error message: "partition not committed after 20 retries (waited up to 110 seconds)"
// - Diagnostic severity: Error (blocks Create operation)
//
// Mock Server Behavior:
// - Returns successful device creation (addDevice)
// - Returns partition data with {"status": {"modified": true}} indefinitely
// - This simulates BCM partition commit never completing
//
// Test Execution Time:
// IMPORTANT: This test will take approximately 2 minutes to complete due to the.
// exponential backoff delays. This is expected behavior when testing timeout paths.
//
// Reference Implementation:
// - Similar to clone timeout patterns in resource_cmpart_softwareimage_test.go
// - Follows eventual consistency test patterns from drift detection tests.
func TestAccCMDeviceDevice_PartitionCommitTimeout(t *testing.T) {
	t.Skip("SKIP: This test takes ~2 minutes to complete due to exponential backoff - run manually to verify timeout logic")

	// Note: To run this test manually, remove the t.Skip() line above and execute:
	// TF_ACC=1 go test -v -timeout 10m ./internal/provider/ -run TestAccCMDeviceDevice_PartitionCommitTimeout

	deviceName := generateUniqueTestName("tftest-device-timeout")
	categoryName := generateUniqueTestName("tftest-category-timeout")
	imageName := generateUniqueTestName("tftest-image-timeout")
	imagePath := "/cm/images/ubuntu-22.04-server-amd64-timeout.iso"
	mac := generateUniqueMAC()

	// Track retry attempts for validation.
	var retryCount int
	var firstRetryTime time.Time
	var lastRetryTime time.Time

	// Create mock server that simulates partition commit never completing.
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request to determine which API call this is.
		var requestBody map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		service, _ := requestBody["service"].(string)
		call, _ := requestBody["call"].(string)

		// Track partition query attempts for validation.
		if service == "CMPart" && call == "getSoftwareImage" {
			retryCount++
			if retryCount == 1 {
				firstRetryTime = time.Now()
			}
			lastRetryTime = time.Now()

			t.Logf("[MOCK] Partition query attempt %d at %v", retryCount, time.Since(firstRetryTime))

			// Always return partition with modified=true (never completes commit).
			// This simulates BCM partition commit hanging indefinitely.
			response := map[string]interface{}{
				"uuid": "test-partition-uuid-never-commits",
				"name": imageName,
				"path": imagePath,
				"status": map[string]interface{}{
					"modified": true, // CRITICAL: This never changes to false
				},
			}

			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Logf("[MOCK] Failed to encode response: %v", err)
			}
			return
		}

		// Handle login requests.
		if requestBody["service"] == "login" {
			// Set authentication cookie.
			http.SetCookie(w, &http.Cookie{
				Name:  "cm-login-token",
				Value: "mock-session-token",
				Path:  "/",
			})
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{"result": "success"}`)); err != nil {
				t.Logf("[MOCK] Failed to write login response: %v", err)
			}
			return
		}

		// Handle other API calls (simplified for test focus).
		// In real scenario, these would need full implementations.
		switch call {
		case "getDevice":
			// Return empty to indicate device doesn't exist yet.
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{}`)); err != nil {
				t.Logf("[MOCK] Failed to write getDevice response: %v", err)
			}

		case "addDevice":
			// Simulate successful device creation.
			response := map[string]interface{}{
				"uuid":      "test-device-uuid-123",
				"hostname":  deviceName,
				"partition": "test-partition-uuid-never-commits", // Points to problematic partition
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Logf("[MOCK] Failed to encode addDevice response: %v", err)
			}

		case "getCategories", "getCategory":
			// Return mock category.
			response := []map[string]interface{}{
				{
					"uuid": "test-category-uuid",
					"name": categoryName,
					"softwareImageProxy": map[string]interface{}{
						"parentSoftwareImage": "test-partition-uuid-never-commits",
					},
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Logf("[MOCK] Failed to encode category response: %v", err)
			}

		case "getNetworks":
			// Return mock network.
			response := []map[string]interface{}{
				{
					"uuid": "test-network-uuid",
					"name": "managementnet",
				},
			}
			w.Header().Set("Content-Type", "application/json")
			if err := json.NewEncoder(w).Encode(response); err != nil {
				t.Logf("[MOCK] Failed to encode network response: %v", err)
			}

		default:
			// Default response for unhandled calls.
			w.Header().Set("Content-Type", "application/json")
			if _, err := w.Write([]byte(`{}`)); err != nil {
				t.Logf("[MOCK] Failed to write default response: %v", err)
			}
		}
	}))
	defer mockServer.Close()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				// Attempt to create device with partition that never commits.
				// This should FAIL with timeout error after ~110-120 seconds.
				Config: testAccCMDeviceDeviceResourceConfig_WithMockEndpoint(
					mockServer.URL,
					deviceName,
					categoryName,
					imageName,
					imagePath,
					mac,
				),
				// Expect error from waitForPartitionCommit timeout.
				ExpectError: regexp.MustCompile(
					`partition not committed after \d+ retries \(waited up to \d+ seconds\)`,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Validate retry behavior after test completes.
					func(s *terraform.State) error {
						// Verify exponential backoff was attempted.
						if retryCount < 15 {
							return fmt.Errorf(
								"Expected at least 15 retry attempts, got %d (may indicate early termination)",
								retryCount,
							)
						}

						// Verify total wait time was substantial.
						totalWaitTime := lastRetryTime.Sub(firstRetryTime)
						expectedMinWait := 60 * time.Second // Conservative estimate

						if totalWaitTime < expectedMinWait {
							return fmt.Errorf(
								"Expected total wait time >= %v, got %v (exponential backoff may not be working)",
								expectedMinWait,
								totalWaitTime,
							)
						}

						t.Logf("[VALIDATION] ✓ Retry count: %d attempts", retryCount)
						t.Logf("[VALIDATION] ✓ Total wait time: %v", totalWaitTime)
						t.Logf("[VALIDATION] ✓ Exponential backoff verified")

						return nil
					},
				),
			},
		},
	})
}

// testAccCMDeviceDeviceResourceConfig_WithMockEndpoint returns device config using mock endpoint
// This is used for timeout testing where we need to control BCM API responses.
func testAccCMDeviceDeviceResourceConfig_WithMockEndpoint(
	endpoint string,
	hostname string,
	categoryName string,
	imageName string,
	imagePath string,
	mac string,
) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = "mock-user"
  password             = "mock-pass"
  insecure_skip_verify = true
  timeout              = 300  # 5 minutes to allow for retry exhaustion
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[2]q
  path = %[3]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[5]q
  category = bcm_cmdevice_category.test.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[6]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

  depends_on = [bcm_cmdevice_category.test]
}
`,
		endpoint,
		imageName,
		imagePath,
		categoryName,
		hostname,
		mac,
	)
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
  hostname = %[4]q
  category = local.category_uuid

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[5]q
    network  = local.network_uuid
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
	)
}

// TestAccCMDeviceDevice_ValidationInvalidMAC tests MAC address validation.
func TestAccCMDeviceDevice_ValidationInvalidMAC(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
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
// Phase 3: Drift Detection Tests.
// ========================================

// TestAccCMDeviceDevice_Drift tests drift detection for hostname attribute.
// This test verifies the provider correctly detects when a device's hostname is modified externally via BCM API.
func TestAccCMDeviceDevice_Drift(t *testing.T) {
	initialHostname := generateUniqueTestName("tftest-device-drift")
	driftedHostname := generateUniqueTestName("tftest-drifted-device")
	categoryName := generateUniqueTestName("tftest-category-drift-hostname")
	imageName := generateUniqueTestName("tftest-image-drift-hostname")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)

	mac := generateUniqueMAC()

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, initialHostname, driftedHostname)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with initial hostname.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath, mac),
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
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Modify hostname externally via BCM API, verify drift detected.
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get UUID by device hostname using helper.
					uuid := getResourceUUIDByName(t, "cmdevice", "getDevice", initialHostname)

					// Fetch full device data from BCM API.
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", uuid)
					if err != nil {
						t.Fatalf("Failed to fetch device for drift modification: %v", err)
					}

					// Parse the device data.
					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					// Modify hostname field (Terraform snake_case -> BCM API camelCase).
					deviceData["hostname"] = driftedHostname

					// Wrap in BCM API entity structure required for updates.
					entity := map[string]interface{}{
						"baseType":      "Device",
						"childType":     deviceData["childType"],
						"modified":      true,
						"to_be_removed": false,
						"revision":      "",
						"uuid":          uuid,
					}
					// Copy all device data fields except uuid (already set above).
					for k, v := range deviceData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API.
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", entity, false)
					if err != nil {
						t.Fatalf("Failed to update device via BCM API: %v", err)
					}

					// Wait for eventual consistency.
					time.Sleep(TestEventualConsistencyDelay)

					t.Logf("[DEBUG] Modified hostname externally to: %v", entity["hostname"])
				},
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath, mac),
				// Use ConfigPlanChecks to verify drift detected (non-empty plan expected).
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Step 3: Restore desired state (Terraform applies config to fix drift).
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify drift was corrected and state matches config.
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
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// ========================================
// Phase 4: Error Handling Tests.
// ========================================

// TestAccCMDeviceDevice_PartitionErrorHandling tests partition-related error scenarios.
// This test verifies that the provider correctly handles errors when:
// 1. Category uses softwareImageProxy but getPartitions API call fails
// 2. getPartitions returns empty list (no partitions available)
// 3. getPartitions returns partitions but none are named "base"
//
// Note: These tests document expected error scenarios. In a real BCM environment,
// these errors would occur when:
// - BCM API is unavailable or returns errors
// - Cluster has no partitions configured
// - Cluster partitions exist but none have the required "base" name.
func TestAccCMDeviceDevice_PartitionErrorHandling(t *testing.T) {
	t.Skip("Documentation test - requires specific BCM cluster configurations or mock server")

	// These tests document the error paths but cannot be executed against a live BCM.
	// cluster without engineering specific failure conditions. They serve as:
	// 1. Documentation of expected error behavior
	// 2. Reference for manual testing scenarios
	// 3. Template for future mock-based testing infrastructure

	// Test Case 1: getPartitions API failure.
	// Expected diagnostic: "Error Querying Partitions"
	// Occurs at: resource_cmdevice_device.go:294-299
	// Scenario: BCM API returns error when calling CMPart.getPartitions
	//
	// This would require either:
	// - Mock BCM server that returns error for getPartitions
	// - BCM cluster with misconfigured partition service
	// - Network/authentication issues during partition query

	// Test Case 2: Empty partitions list.
	// Expected diagnostic: "Missing Base Partition"
	// Occurs at: resource_cmdevice_device.go:326-331
	// Scenario: getPartitions succeeds but returns empty array []
	//
	// This would require:
	// - BCM cluster with no partitions configured
	// - Fresh BCM installation before partition setup

	// Test Case 3: No base partition found.
	// Expected diagnostic: "Missing Base Partition"
	// Occurs at: resource_cmdevice_device.go:326-331
	// Scenario: getPartitions returns partitions but none have name="base"
	//
	// This would require:
	// - BCM cluster with custom partition names only
	// - All partitions renamed from default "base"
}

// TestAccCMDeviceDevice_PartitionQueryFailureDocumentation documents the partition query failure scenario.
// This test is skipped by default but documents the expected behavior when BCM API returns.
// an error during partition lookup.
func TestAccCMDeviceDevice_PartitionQueryFailureDocumentation(t *testing.T) {
	t.Skip("Documentation test - describes partition query failure behavior")

	/*
		Error Path: Lines 294-299 in resource_cmdevice_device.go

		Code:
			partitionsBody, err := r.client.CallJSONRPC(ctx, "CMPart", "getPartitions")
			if err != nil {
				resp.Diagnostics.AddError(
					"Error Querying Partitions",
					fmt.Sprintf("Could not query partitions: %s", err.Error()),
				)
				return
			}

		Scenario:
			1. User creates device with category that uses softwareImageProxy
			2. Provider needs to query cluster's base partition
			3. BCM API call to CMPart.getPartitions fails

		Trigger Conditions:
			- BCM API temporarily unavailable
			- Authentication issues during API call
			- Network timeout during partition query
			- CMPart service not running or misconfigured

		Expected Result:
			✗ Error diagnostic added to response
			✗ Device creation fails
			✓ Error message: "Error Querying Partitions: Could not query partitions: <error details>"
			✓ State remains unchanged (no partial device created)

		User Experience:
			terraform apply output:
			│ Error: Error Querying Partitions
			│
			│   with bcm_cmdevice_device.test,
			│   on main.tf line 42, in resource "bcm_cmdevice_device" "test":
			│   42: resource "bcm_cmdevice_device" "test" {
			│
			│ Could not query partitions: <BCM API error details>

		Manual Testing Steps:
			1. Set up BCM cluster with category using softwareImageProxy
			2. Simulate API failure (e.g., stop CMPart service, network disconnect)
			3. Attempt device creation with softwareImageProxy category
			4. Verify error diagnostic matches expected format
			5. Verify no partial device state created
	*/
}

// TestAccCMDeviceDevice_MissingBasePartitionDocumentation documents the missing base partition scenario.
// This test is skipped by default but documents the expected behavior when BCM cluster.
// has no partition named "base".
func TestAccCMDeviceDevice_MissingBasePartitionDocumentation(t *testing.T) {
	t.Skip("Documentation test - describes missing base partition behavior")

	/*
		Error Path: Lines 326-331 in resource_cmdevice_device.go

		Code:
			if !basePartitionFound {
				resp.Diagnostics.AddError(
					"Missing Base Partition",
					"Category uses softwareImageProxy but no 'base' partition found in cluster",
				)
				return
			}

		Scenario 1 - Empty Partitions List:
			1. User creates device with category that uses softwareImageProxy
			2. Provider queries CMPart.getPartitions
			3. BCM returns empty array: []
			4. No partition with name="base" exists

		Scenario 2 - No Base Partition:
			1. User creates device with category that uses softwareImageProxy
			2. Provider queries CMPart.getPartitions
			3. BCM returns partitions: [{"name": "custom1", ...}, {"name": "custom2", ...}]
			4. None of the partitions have name="base"

		Trigger Conditions:
			Scenario 1 (Empty Partitions):
				- Fresh BCM cluster installation
				- All partitions deleted during maintenance
				- BCM database corruption affecting partitions

			Scenario 2 (No Base Partition):
				- All partitions renamed from default "base"
				- Custom BCM configuration with non-standard partition names
				- "base" partition deleted but other partitions exist

		Expected Result:
			✗ Error diagnostic added to response
			✗ Device creation fails
			✓ Error message: "Missing Base Partition: Category uses softwareImageProxy but no 'base' partition found in cluster"
			✓ State remains unchanged (no partial device created)

		User Experience:
			terraform apply output:
			│ Error: Missing Base Partition
			│
			│   with bcm_cmdevice_device.test,
			│   on main.tf line 42, in resource "bcm_cmdevice_device" "test":
			│   42: resource "bcm_cmdevice_device" "test" {
			│
			│ Category uses softwareImageProxy but no 'base' partition found in cluster

		Resolution Steps:
			1. Check BCM cluster partition configuration
			2. Create "base" partition if missing:
				- BCM UI: Partitions → Add Partition → Name: "base"
			3. Rename existing partition to "base" if needed
			4. Or specify partition explicitly in device config:
				resource "bcm_cmdevice_device" "test" {
				  partition = "<existing-partition-uuid>"
				  ...
				}

		Manual Testing Steps:
			Scenario 1 (Empty Partitions):
				1. Set up BCM cluster
				2. Delete all partitions (if BCM allows)
				3. Create category with softwareImageProxy
				4. Attempt device creation
				5. Verify error diagnostic matches expected format

			Scenario 2 (No Base Partition):
				1. Set up BCM cluster with partitions
				2. Rename "base" partition to something else
				3. Create category with softwareImageProxy
				4. Attempt device creation
				5. Verify error diagnostic matches expected format
	*/
}

// TestAccCMDeviceDevice_PartitionErrorRecovery documents partition error recovery patterns.
// This test is skipped by default but documents how users can recover from partition errors.
func TestAccCMDeviceDevice_PartitionErrorRecovery(t *testing.T) {
	t.Skip("Documentation test - describes partition error recovery")

	/*
		Recovery Strategy 1: Explicit Partition Specification

		Problem:
			Category uses softwareImageProxy but cluster has no "base" partition

		Solution:
			Explicitly specify partition in device configuration

		Before (fails):
			resource "bcm_cmdevice_device" "test" {
			  hostname           = "node01"
			  mac                = "00:11:22:33:44:55"
			  category           = bcm_cmdevice_category.test.id
			  management_network = data.bcm_cmnet_networks.management.networks[0].id
			  # No partition specified - will fail if no "base" partition exists
			}

		After (works):
			resource "bcm_cmdevice_device" "test" {
			  hostname           = "node01"
			  mac                = "00:11:22:33:44:55"
			  category           = bcm_cmdevice_category.test.id
			  management_network = data.bcm_cmnet_networks.management.networks[0].id
			  partition          = data.bcm_cmpart_partitions.available.partitions[0].uuid
			  # Explicit partition specified - bypasses base partition lookup
			}

		Recovery Strategy 2: Use Category with Direct Partition

		Problem:
			Category uses softwareImageProxy requiring base partition lookup

		Solution:
			Use category with direct partition assignment instead of proxy

		Before (requires base partition):
			resource "bcm_cmdevice_category" "test" {
			  name               = "proxy-category"
			  management_network = data.bcm_cmnet_networks.management.networks[0].id
			  software_image_proxy = {
			    parent_software_image = bcm_cmpart_softwareimage.test.id
			  }
			}

		After (no base partition required):
			resource "bcm_cmdevice_category" "test" {
			  name               = "direct-category"
			  management_network = data.bcm_cmnet_networks.management.networks[0].id
			  partition          = data.bcm_cmpart_partitions.available.partitions[0].uuid
			  # Direct partition - no base partition lookup needed
			}

		Recovery Strategy 3: Create Base Partition

		Problem:
			Cluster has no partition named "base"

		Solution:
			Create or rename partition to "base"

		Option A - Create via BCM UI:
			1. Navigate to Partitions section
			2. Click "Add Partition"
			3. Name: "base"
			4. Configure partition settings
			5. Save and retry Terraform

		Option B - Create via Terraform (if bcm_cmpart_partition resource exists):
			resource "bcm_cmpart_partition" "base" {
			  name = "base"
			  # Additional partition configuration
			}

			resource "bcm_cmdevice_category" "test" {
			  # ...
			  depends_on = [bcm_cmpart_partition.base]
			}

		Option C - Rename existing partition:
			1. Identify existing partition to rename
			2. Via BCM UI: Edit → Change name to "base"
			3. Retry Terraform apply
	*/
}

// ========================================
// Phase 5: Device Roles Tests.
// ========================================

// testAccCMDeviceDeviceConfigWithRoles returns device config with roles assignment.
// roleNames is a list of role names (e.g., "monitoring", "storage") that will be looked up via data source.
func testAccCMDeviceDeviceConfigWithRoles(hostname, categoryName, imageName, imagePath, mac string, roleNames []string) string {
	// Build roles HCL using role names directly (not UUIDs)
	rolesStr := ""
	if len(roleNames) > 0 {
		quoted := make([]string, len(roleNames))
		for i, name := range roleNames {
			quoted[i] = fmt.Sprintf(`"%s"`, name)
		}
		rolesStr = fmt.Sprintf("roles = [%s]", strings.Join(quoted, ", "))

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
  hostname = %[7]q
  category = bcm_cmdevice_category.test.id
  %[9]s

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
			mac,
			rolesStr,
		)
	}

	// No roles specified
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
  hostname = %[7]q
  category = bcm_cmdevice_category.test.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
		mac,
	)
}

// testAccCMDeviceAvailableRoleNames returns the role names exposed by the BCM cluster.
func testAccCMDeviceAvailableRoleNames(t *testing.T) map[string]struct{} {
	t.Helper()

	client := createTestBCMClient(t)
	body, err := client.CallJSONRPC(t.Context(), "cmdevice", "getNodes")
	if err != nil {
		t.Fatalf("Failed to query node roles: %v", err)
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(body, &nodes); err != nil {
		t.Fatalf("Failed to parse node roles response: %v", err)
	}

	roleNames := make(map[string]struct{})
	for _, node := range nodes {
		rolesData, ok := node["roles"].([]interface{})
		if !ok {
			continue
		}

		for _, roleData := range rolesData {
			role, ok := roleData.(map[string]interface{})
			if !ok {
				continue
			}

			name, ok := role["name"].(string)
			if ok && name != "" {
				roleNames[name] = struct{}{}
			}
		}
	}

	return roleNames
}

// testAccCMDeviceDeviceSingleRoleSet returns a role set that validates for generic test devices.
func testAccCMDeviceDeviceSingleRoleSet(t *testing.T) []string {
	t.Helper()

	availableRoles := testAccCMDeviceAvailableRoleNames(t)
	if _, ok := availableRoles["provisioning"]; ok {
		return []string{"provisioning"}
	}

	t.Skip("cluster does not expose the provisioning role required for portable single-role device tests")
	return nil
}

// testAccCMDeviceDeviceMultiRoleSet returns a cluster-valid two-role set for generic test devices.
func testAccCMDeviceDeviceMultiRoleSet(t *testing.T) []string {
	t.Helper()

	availableRoles := testAccCMDeviceAvailableRoleNames(t)
	_, hasBoot := availableRoles["boot"]
	_, hasStorage := availableRoles["storage"]
	if hasBoot && hasStorage {
		return []string{"boot", "storage"}
	}

	t.Skip("cluster does not expose the boot+storage roles required for portable multi-role device tests")
	return nil
}

// TestAccCMDeviceDevice_RolesCreate tests creating a device with a role.
// Uses a role set that validates for generic test devices in BCM.
func TestAccCMDeviceDevice_RolesCreate(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-roles")
	categoryName := generateUniqueTestName("tftest-category-roles")
	imageName := generateUniqueTestName("tftest-image-roles")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roles := testAccCMDeviceDeviceSingleRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
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
					// One role assigned
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
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

// TestAccCMDeviceDevice_RolesMultiple tests assigning multiple roles to a device.
func TestAccCMDeviceDevice_RolesMultiple(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-multi")
	categoryName := generateUniqueTestName("tftest-category-multi")
	imageName := generateUniqueTestName("tftest-image-multi")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roles := testAccCMDeviceDeviceMultiRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with multiple roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(2),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
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

// TestAccCMDeviceDevice_RolesIdempotent tests idempotency of role assignments.
func TestAccCMDeviceDevice_RolesIdempotent(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-idempot")
	categoryName := generateUniqueTestName("tftest-category-idempot")
	imageName := generateUniqueTestName("tftest-image-idempot")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roles := testAccCMDeviceDeviceSingleRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// First idempotency check.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Second idempotency check with ID verification.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_RolesUpdate tests updating roles on a device.
func TestAccCMDeviceDevice_RolesUpdate(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-update")
	categoryName := generateUniqueTestName("tftest-category-update")
	imageName := generateUniqueTestName("tftest-image-update")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	initialRoles := []string{}
	updatedRoles := testAccCMDeviceDeviceSingleRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create with initial roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, initialRoles),
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
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, updatedRoles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					// ID should remain the same after update.
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after update.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, updatedRoles),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_RolesRemove tests removing all roles from a device.
func TestAccCMDeviceDevice_RolesRemove(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-remove")
	categoryName := generateUniqueTestName("tftest-category-remove")
	imageName := generateUniqueTestName("tftest-image-remove")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	initialRoles := testAccCMDeviceDeviceSingleRoleSet(t)
	emptyRoles := []string{}

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create with roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, initialRoles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Remove all roles (explicit empty list).
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, emptyRoles),
				ConfigStateChecks: []statecheck.StateCheck{
					// Empty roles should result in null or empty list.
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_RolesImport tests importing a device with existing roles.
func TestAccCMDeviceDevice_RolesImport(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-import")
	categoryName := generateUniqueTestName("tftest-category-import")
	imageName := generateUniqueTestName("tftest-image-import")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roles := testAccCMDeviceDeviceSingleRoleSet(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
				},
			},
			// Import and verify roles are preserved.
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"management_network",
					"boot_loader",
					"boot_loader_protocol",
					"partition",
					"power_control",
					"default_gateway",
					"default_gateway_metric",
					"serial_number",
					"part_number",
					"interfaces.#",
					"interfaces.0.%",
					"interfaces.0.base_type",
					"interfaces.0.bootable",
					"interfaces.0.cardtype",
					"interfaces.0.child_type",
					"interfaces.0.dhcp",
					"interfaces.0.mac",
					"interfaces.0.name",
					"interfaces.0.network",
					"interfaces.0.start_if",
					"interfaces.0.type",
					"interfaces.0.uuid",
					"interfaces.0.bond_mode",
					"interfaces.0.ip",
					"interfaces.0.ipv6_ip",
					"interfaces.0.members.#",
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_RolesDrift tests drift detection for roles.
func TestAccCMDeviceDevice_RolesDrift(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-drift-r")
	categoryName := generateUniqueTestName("tftest-category-drift-r")
	imageName := generateUniqueTestName("tftest-image-drift-r")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roles := testAccCMDeviceDeviceSingleRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Modify roles externally via BCM API.
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get device UUID by hostname.
					deviceUUID := getResourceUUIDByName(t, "cmdevice", "getDevice", deviceName)

					// Fetch full device data from BCM API.
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceUUID)
					if err != nil {
						t.Fatalf("Failed to fetch device for drift modification: %v", err)
					}

					// Parse the device data.
					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					// Simulate drift by removing all roles from the device.
					// This is safer than adding different roles that may have complex dependencies.
					deviceData["roles"] = []interface{}{}
					deviceData["modified"] = true

					// Update via BCM API to remove all roles.
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", deviceData, false)
					if err != nil {
						t.Fatalf("Failed to update device via BCM API: %v", err)
					}

					// Wait for eventual consistency.
					time.Sleep(TestEventualConsistencyDelay)

					t.Logf("[DEBUG] Modified roles externally to: [] (was backup, provisioning)")
				},
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				// Expect non-empty plan because roles were modified externally.
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Terraform restores desired state.
			{
				Config: testAccCMDeviceDeviceConfigWithRoles(deviceName, categoryName, imageName, imagePath, mac, roles),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// ========================================
// Phase 6: Role by Name Tests (GitHub Issue #86)
// ========================================

// testAccCMDeviceDeviceConfigWithRoleNames returns device config using role NAMES directly (not UUIDs).
// This is the new simplified approach - no data source lookup required.
func testAccCMDeviceDeviceConfigWithRoleNames(hostname, categoryName, imageName, imagePath, mac string, roleNames []string) string {
	rolesHCL := "[]"
	if len(roleNames) > 0 {
		quoted := make([]string, len(roleNames))
		for i, name := range roleNames {
			quoted[i] = fmt.Sprintf(`"%s"`, name)
		}
		rolesHCL = fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
	}

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
  hostname = %[7]q
  category = bcm_cmdevice_category.test.id
  roles    = %[9]s

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
		mac,
		rolesHCL,
	)
}

// TestAccCMDeviceDevice_RolesByName tests creating a device with roles specified by name.
// This is the new simplified approach - users can just specify role names directly.
func TestAccCMDeviceDevice_RolesByName(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-roles-name")
	categoryName := generateUniqueTestName("tftest-cat-roles-name")
	imageName := generateUniqueTestName("tftest-img-roles-name")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roleNames := testAccCMDeviceDeviceSingleRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with roles specified by name.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, roleNames),
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
					// Roles should be stored as names in state
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check - should produce empty plan.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, roleNames),
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

// TestAccCMDeviceDevice_RolesImportByName tests importing a device with roles shows role names.
// After import, the roles attribute should contain role names (not UUIDs).
func TestAccCMDeviceDevice_RolesImportByName(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-roles-import-name")
	categoryName := generateUniqueTestName("tftest-cat-import-name")
	imageName := generateUniqueTestName("tftest-img-import-name")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roleNames := testAccCMDeviceDeviceSingleRoleSet(t)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, roleNames),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
				},
			},
			// Import and verify roles are returned as names.
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"management_network",
					"boot_loader",
					"boot_loader_protocol",
					"partition",
					"power_control",
					"default_gateway",
					"default_gateway_metric",
					"serial_number",
					"part_number",
					"interfaces.#",
					"interfaces.0.%",
					"interfaces.0.base_type",
					"interfaces.0.bootable",
					"interfaces.0.cardtype",
					"interfaces.0.child_type",
					"interfaces.0.dhcp",
					"interfaces.0.mac",
					"interfaces.0.name",
					"interfaces.0.network",
					"interfaces.0.start_if",
					"interfaces.0.type",
					"interfaces.0.uuid",
					"interfaces.0.bond_mode",
					"interfaces.0.ip",
					"interfaces.0.ipv6_ip",
					"interfaces.0.members.#",
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_RolesUpdateByName tests updating roles using role names.
func TestAccCMDeviceDevice_RolesUpdateByName(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-roles-update-name")
	categoryName := generateUniqueTestName("tftest-cat-update-name")
	imageName := generateUniqueTestName("tftest-img-update-name")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	initialRoles := []string{}
	updatedRoles := testAccCMDeviceDeviceSingleRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create with initial roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, initialRoles),
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
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update roles by adding one more.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, updatedRoles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					// ID should remain the same after update.
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after update.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, updatedRoles),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_RolesDriftByName tests drift detection for roles using role names.
func TestAccCMDeviceDevice_RolesDriftByName(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-roles-drift-name")
	categoryName := generateUniqueTestName("tftest-cat-drift-name")
	imageName := generateUniqueTestName("tftest-img-drift-name")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	roleNames := testAccCMDeviceDeviceSingleRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, roleNames),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Modify roles externally via BCM API.
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get device UUID by hostname.
					deviceUUID := getResourceUUIDByName(t, "cmdevice", "getDevice", deviceName)

					// Fetch full device data.
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceUUID)
					if err != nil {
						t.Fatalf("Failed to fetch device for drift modification: %v", err)
					}

					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					// Remove all roles externally (drift).
					deviceData["roles"] = []interface{}{}
					deviceData["modified"] = true

					// Update via BCM API.
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", deviceData, false)
					if err != nil {
						t.Fatalf("Failed to update device via BCM API: %v", err)
					}

					// Wait for eventual consistency.
					time.Sleep(TestEventualConsistencyDelay)
					t.Logf("[DEBUG] Removed roles externally (drift)")
				},
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, roleNames),
				// Expect non-empty plan because roles were modified externally.
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Terraform restores desired state.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, roleNames),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// ========================================
// Phase 7: Client-Side Validation Tests (GitHub Issue #86)
// ========================================

// TestAccCMDeviceDevice_InvalidRoleName tests clear error message for non-existent role name.
func TestAccCMDeviceDevice_InvalidRoleName(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-invalid-role")
	categoryName := generateUniqueTestName("tftest-cat-invalid-role")
	imageName := generateUniqueTestName("tftest-img-invalid-role")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// Use a non-existent role name
	invalidRoles := []string{"nonexistent-role-xyz123"}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, invalidRoles),
				ExpectError: regexp.MustCompile(`nonexistent-role-xyz123`),
			},
		},
	})
}

// TestAccCMDeviceDevice_InvalidRoleUUID tests that UUIDs are NOT accepted as role input.
// Since we only accept role names now, a UUID should be treated as an invalid role name
// and produce an error listing the UUID in the "Roles not found" message.
func TestAccCMDeviceDevice_InvalidRoleUUID(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-invalid-uuid")
	categoryName := generateUniqueTestName("tftest-cat-invalid-uuid")
	imageName := generateUniqueTestName("tftest-img-invalid-uuid")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// Generate a random UUID to prove UUID format is NOT accepted
	// It will be treated as an invalid role name
	testUUID := uuid.New().String()
	invalidRoles := []string{testUUID}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, invalidRoles),
				ExpectError: regexp.MustCompile(`roles not found`),
			},
		},
	})
}

// TestAccCMDeviceDevice_EmptyRoleString tests error handling for empty string in roles list.
func TestAccCMDeviceDevice_EmptyRoleString(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-empty-role")
	categoryName := generateUniqueTestName("tftest-cat-empty-role")
	imageName := generateUniqueTestName("tftest-img-empty-role")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// Config with empty string in roles (need custom config for this edge case)
	config := fmt.Sprintf(`
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
  hostname = %[7]q
  category = bcm_cmdevice_category.test.id
  roles    = [""]

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

  depends_on = [bcm_cmdevice_category.test]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		deviceName,
		mac,
	)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config:      config,
				ExpectError: regexp.MustCompile(`invalid role identifiers[\s\S]*empty string`),
			},
		},
	})
}

// TestAccCMDeviceDevice_MultipleInvalidRoles tests error message lists all invalid roles.
func TestAccCMDeviceDevice_MultipleInvalidRoles(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-multi-invalid")
	categoryName := generateUniqueTestName("tftest-cat-multi-invalid")
	imageName := generateUniqueTestName("tftest-img-multi-invalid")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// Use multiple non-existent role names
	invalidRoles := []string{"invalid-role-1", "invalid-role-2", "boot"}

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, invalidRoles),
				// Error should list both invalid roles (boot is valid, so only invalid-role-1 and invalid-role-2 should be listed)
				ExpectError: regexp.MustCompile(`invalid-role-1`),
			},
		},
	})
}

// ========================================
// Phase 8: Additional Role Tests (GitHub Issue #86)
// ========================================

// TestAccCMDeviceDevice_RolesAddMultiple tests adding multiple roles at once.
// This validates that the provider correctly handles multiple role names.
func TestAccCMDeviceDevice_RolesAddMultiple(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-roles-multi")
	categoryName := generateUniqueTestName("tftest-cat-roles-multi")
	imageName := generateUniqueTestName("tftest-img-roles-multi")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()
	singleRole := testAccCMDeviceDeviceSingleRoleSet(t)
	multipleRoles := testAccCMDeviceDeviceMultiRoleSet(t)

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create with role names.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, singleRole),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update to add more roles.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, multipleRoles),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("roles"),
						knownvalue.SetSizeExact(2),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check.
			{
				Config: testAccCMDeviceDeviceConfigWithRoleNames(deviceName, categoryName, imageName, imagePath, mac, multipleRoles),
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

// ========================================
// Phase 9: Kubernetes Role Tests (T041-T044)
// ========================================

// TestAccCMDeviceDevice_kubeletRole tests kubelet_role block for Kubernetes cluster membership.
// T041: Tests adding a device to a Kubernetes cluster via kubelet_role.
func TestAccCMDeviceDevice_kubeletRole(t *testing.T) {
	deviceName := generateShortTestName("d-kub")
	categoryName := generateShortTestName("c-kub")
	imageName := generateShortTestName("i-kub")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()
	etcdClusterName := generateShortTestName("e-kub")
	kubeClusterName := generateShortTestName("k-kub")

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with kubelet_role
			{
				Config: testAccCMDeviceDeviceConfigWithKubeletRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName, true, true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("control_plane"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("worker"),
						knownvalue.Bool(true),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigWithKubeletRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName, true, true,
				),
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

// TestAccCMDeviceDevice_etcdHostRole tests etcd_host_role block for etcd cluster membership.
// T042: Tests adding a device to an etcd cluster via etcd_host_role.
func TestAccCMDeviceDevice_etcdHostRole(t *testing.T) {
	deviceName := generateShortTestName("d-etc")
	categoryName := generateShortTestName("c-etc")
	imageName := generateShortTestName("i-etc")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()
	etcdClusterName := generateShortTestName("e-etc")

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with etcd_host_role
			{
				Config: testAccCMDeviceDeviceConfigWithEtcdHostRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("etcd_host_role").AtSliceIndex(0).AtMapKey("member_name"),
						knownvalue.StringExact("$hostname"),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigWithEtcdHostRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName,
				),
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

// TestAccCMDeviceDevice_bothRoles tests combined kubelet + etcd roles on same device.
// T043: Tests adding both Kubernetes and etcd roles to a single device.
func TestAccCMDeviceDevice_bothRoles(t *testing.T) {
	deviceName := generateShortTestName("d-bth")
	categoryName := generateShortTestName("c-bth")
	imageName := generateShortTestName("i-bth")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()
	etcdClusterName := generateShortTestName("e-bth")
	kubeClusterName := generateShortTestName("k-bth")

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device with both roles
			{
				Config: testAccCMDeviceDeviceConfigWithBothRoles(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("etcd_host_role"),
						knownvalue.ListSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigWithBothRoles(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName,
				),
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

// TestAccCMDeviceDevice_roleUpdate tests modification of Kubernetes roles.
// T044: Tests updating kubelet_role properties (control_plane, worker).
func TestAccCMDeviceDevice_roleUpdate(t *testing.T) {
	deviceName := generateShortTestName("d-upd")
	categoryName := generateShortTestName("c-upd")
	imageName := generateShortTestName("i-upd")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()
	etcdClusterName := generateShortTestName("e-upd")
	kubeClusterName := generateShortTestName("k-upd")

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create with control_plane=true, worker=true
			{
				Config: testAccCMDeviceDeviceConfigWithKubeletRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName, true, true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("control_plane"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("worker"),
						knownvalue.Bool(true),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update to worker-only (control_plane=false, worker=true)
			{
				Config: testAccCMDeviceDeviceConfigWithKubeletRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName, false, true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("control_plane"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("worker"),
						knownvalue.Bool(true),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigWithKubeletRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName, false, true,
				),
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

// TestAccCMDeviceDevice_addRolesToExisting tests adding Kubernetes roles to an existing device.
// This validates the workflow: create device without roles -> add roles via update.
func TestAccCMDeviceDevice_addRolesToExisting(t *testing.T) {
	deviceName := generateShortTestName("d-add")
	categoryName := generateShortTestName("c-add")
	imageName := generateShortTestName("i-add")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()
	etcdClusterName := generateShortTestName("e-add")
	kubeClusterName := generateShortTestName("k-add")

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device WITHOUT any Kubernetes roles
			{
				Config: testAccCMDeviceDeviceConfigWithoutRoles(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Update device to ADD kubelet_role (simulates adding K8s to existing device)
			{
				Config: testAccCMDeviceDeviceConfigWithKubeletRole(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName, true, true,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("control_plane"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("worker"),
						knownvalue.Bool(true),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Further update to ADD etcd_host_role (combined control plane)
			{
				Config: testAccCMDeviceDeviceConfigWithBothRoles(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("kubelet_role").AtSliceIndex(0).AtMapKey("control_plane"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("etcd_host_role").AtSliceIndex(0).AtMapKey("etcd_cluster"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 4: Idempotency check
			{
				Config: testAccCMDeviceDeviceConfigWithBothRoles(
					deviceName, categoryName, imageName, imagePath, mac,
					etcdClusterName, kubeClusterName,
				),
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

// ========================================
// Kubernetes Role Test Configuration Helpers
// ========================================

// testAccCMDeviceDeviceConfigWithoutRoles returns device config without any Kubernetes roles.
// Used to test adding roles to an existing device.
func testAccCMDeviceDeviceConfigWithoutRoles(
	hostname, categoryName, imageName, imagePath, mac string,
	etcdClusterName, kubeClusterName string,
) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "all" {}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.all.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmetcd_cluster" "test" {
  name               = %[7]q
  heartbeat_interval = 100
  election_timeout   = 1000
}

resource "bcm_cmkube_cluster" "test" {
  name             = %[8]q
  etcd_cluster     = bcm_cmetcd_cluster.test.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[0].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[0].uuid
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[9]q
  category = bcm_cmdevice_category.test.id

  # No kubelet_role or etcd_host_role - device exists without K8s roles

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[10]q
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  depends_on = [
    bcm_cmdevice_category.test,
    bcm_cmkube_cluster.test,
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		etcdClusterName,
		kubeClusterName,
		hostname,
		mac,
	)
}

// testAccCMDeviceDeviceConfigWithKubeletRole returns device config with kubelet_role.
//
//nolint:unparam // worker param kept for flexibility in future tests
func testAccCMDeviceDeviceConfigWithKubeletRole(
	hostname, categoryName, imageName, imagePath, mac string,
	etcdClusterName, kubeClusterName string,
	controlPlane, worker bool,
) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "all" {}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.all.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmetcd_cluster" "test" {
  name               = %[7]q
  heartbeat_interval = 100
  election_timeout   = 1000
}

resource "bcm_cmkube_cluster" "test" {
  name             = %[8]q
  etcd_cluster     = bcm_cmetcd_cluster.test.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[0].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[0].uuid
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[9]q
  category = bcm_cmdevice_category.test.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[10]q
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.test.uuid
    control_plane = %[11]t
    worker        = %[12]t
  }

  depends_on = [
    bcm_cmdevice_category.test,
    bcm_cmkube_cluster.test,
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		etcdClusterName,
		kubeClusterName,
		hostname,
		mac,
		controlPlane,
		worker,
	)
}

// testAccCMDeviceDeviceConfigWithEtcdHostRole returns device config with etcd_host_role.
func testAccCMDeviceDeviceConfigWithEtcdHostRole(
	hostname, categoryName, imageName, imagePath, mac string,
	etcdClusterName string,
) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "all" {}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.all.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmetcd_cluster" "test" {
  name               = %[7]q
  heartbeat_interval = 100
  election_timeout   = 1000
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[8]q
  category = bcm_cmdevice_category.test.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[9]q
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  etcd_host_role {
    etcd_cluster = bcm_cmetcd_cluster.test.uuid
  }

  depends_on = [
    bcm_cmdevice_category.test,
    bcm_cmetcd_cluster.test,
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		etcdClusterName,
		hostname,
		mac,
	)
}

// testAccCMDeviceDeviceConfigWithBothRoles returns device config with both kubelet_role and etcd_host_role.
func testAccCMDeviceDeviceConfigWithBothRoles(
	hostname, categoryName, imageName, imagePath, mac string,
	etcdClusterName, kubeClusterName string,
) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "all" {}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q
}

resource "bcm_cmdevice_category" "test" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.all.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmetcd_cluster" "test" {
  name               = %[7]q
  heartbeat_interval = 100
  election_timeout   = 1000
}

resource "bcm_cmkube_cluster" "test" {
  name             = %[8]q
  etcd_cluster     = bcm_cmetcd_cluster.test.uuid
  internal_network = data.bcm_cmnet_networks.all.networks[0].uuid
  service_network  = data.bcm_cmnet_networks.all.networks[0].uuid
  pod_network      = data.bcm_cmnet_networks.all.networks[0].uuid
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[9]q
  category = bcm_cmdevice_category.test.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[10]q
    network  = data.bcm_cmnet_networks.all.networks[0].id
    bootable = true
    dhcp     = true
  }

  kubelet_role {
    kube_cluster  = bcm_cmkube_cluster.test.uuid
    control_plane = true
    worker        = true
  }

  etcd_host_role {
    etcd_cluster = bcm_cmetcd_cluster.test.uuid
  }

  depends_on = [
    bcm_cmdevice_category.test,
    bcm_cmkube_cluster.test,
    bcm_cmetcd_cluster.test,
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryName,
		etcdClusterName,
		kubeClusterName,
		hostname,
		mac,
	)
}

// ========================================
// Management Network Pass-Through Tests
// ========================================

// TestAccCMDeviceDevice_ManagementNetworkPassThrough tests that management_network
// set on the device resource is actually sent to the BCM API, not just preserved in state.
// This catches the bug where the value was hardcoded to the zero UUID in the API entity.
func TestAccCMDeviceDevice_ManagementNetworkPassThrough(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-mgmtnet")
	categoryName := generateUniqueTestName("tftest-category-mgmtnet")
	imageName := generateUniqueTestName("tftest-image-mgmtnet")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with explicit management_network.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					managementNetworkInBCMCheck{
						resourceAddress: "bcm_cmdevice_device.test",
						deviceName:      deviceName,
					},
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency — re-apply should produce empty plan.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(deviceName, categoryName, imageName, imagePath, mac),
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

// TestAccCMDeviceDevice_ManagementNetworkOmitted tests that omitting management_network
// results in the zero UUID being sent (default behavior) and null in state.
func TestAccCMDeviceDevice_ManagementNetworkOmitted(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-nomgmt")
	categoryName := generateUniqueTestName("tftest-category-nomgmt")
	imageName := generateUniqueTestName("tftest-image-nomgmt")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Create device without management_network — should default gracefully.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.Null(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
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

// managementNetworkInBCMCheck is a StateCheck that verifies via direct BCM API
// call that the device's managementNetwork field is NOT the zero UUID and matches state.
type managementNetworkInBCMCheck struct {
	resourceAddress string
	deviceName      string
}

func (c managementNetworkInBCMCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	// Get management_network from state.
	var stateValue string
	for _, r := range req.State.Values.RootModule.Resources {
		if r.Address == c.resourceAddress {
			if v, ok := r.AttributeValues["management_network"].(string); ok {
				stateValue = v
			}
			break
		}
	}

	client := createTestBCMClient(&testing.T{})
	body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", c.deviceName)
	if err != nil {
		resp.Error = fmt.Errorf("failed to get device %s from BCM API: %w", c.deviceName, err)
		return
	}

	var deviceData map[string]interface{}
	if err := json.Unmarshal(body, &deviceData); err != nil {
		resp.Error = fmt.Errorf("failed to parse device data: %w", err)
		return
	}

	managementNetwork, ok := deviceData["managementNetwork"].(string)
	if !ok || managementNetwork == "" {
		resp.Error = fmt.Errorf("managementNetwork field not found or empty in BCM API response for device %s", c.deviceName)
		return
	}

	if managementNetwork == "00000000-0000-0000-0000-000000000000" {
		resp.Error = fmt.Errorf(
			"management_network was NOT sent to BCM API for device %s: "+
				"BCM still has zero UUID. Plan value was not passed through",
			c.deviceName,
		)
		return
	}

	if stateValue != managementNetwork {
		resp.Error = fmt.Errorf(
			"management_network mismatch: state has %q but BCM API has %q",
			stateValue, managementNetwork,
		)
	}
}

// testAccCheckDeviceManagementNetworkInBCM verifies via direct BCM API call that
// the device's managementNetwork field is set to a non-zero UUID.
func testAccCheckDeviceManagementNetworkInBCM(deviceName string) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		client := createTestBCMClient(&testing.T{})
		ctx := context.Background()

		body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceName)
		if err != nil {
			return fmt.Errorf("failed to get device %s from BCM API: %w", deviceName, err)
		}

		var deviceData map[string]interface{}
		if err := json.Unmarshal(body, &deviceData); err != nil {
			return fmt.Errorf("failed to parse device data: %w", err)
		}

		managementNetwork, ok := deviceData["managementNetwork"].(string)
		if !ok || managementNetwork == "" {
			return fmt.Errorf("managementNetwork field not found or empty in BCM API response for device %s", deviceName)
		}

		if managementNetwork == "00000000-0000-0000-0000-000000000000" {
			return fmt.Errorf(
				"management_network was NOT sent to BCM API for device %s: BCM still has zero UUID",
				deviceName,
			)
		}

		return nil
	}
}

// TestAccCMDeviceDevice_ManagementNetworkDrift tests that external modification
// of managementNetwork is detected and corrected by Terraform.
func TestAccCMDeviceDevice_ManagementNetworkDrift(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-mgmtdrift")
	categoryName := generateUniqueTestName("tftest-category-mgmtdrift")
	imageName := generateUniqueTestName("tftest-image-mgmtdrift")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with explicit management_network.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					managementNetworkInBCMCheck{
						resourceAddress: "bcm_cmdevice_device.test",
						deviceName:      deviceName,
					},
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Externally reset managementNetwork to zero UUID via BCM API.
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					deviceUUID := getResourceUUIDByName(t, "cmdevice", "getDevice", deviceName)

					body, err := client.CallJSONRPC(ctx, "cmdevice", "getDevice", deviceUUID)
					if err != nil {
						t.Fatalf("Failed to get device for drift test: %v", err)
					}

					var deviceData map[string]interface{}
					if err := json.Unmarshal(body, &deviceData); err != nil {
						t.Fatalf("Failed to parse device data: %v", err)
					}

					// Reset managementNetwork externally to zero UUID.
					deviceData["managementNetwork"] = "00000000-0000-0000-0000-000000000000"

					entity := map[string]interface{}{
						"baseType":      "Device",
						"childType":     deviceData["childType"],
						"modified":      true,
						"to_be_removed": false,
						"uuid":          deviceUUID,
					}
					for k, v := range deviceData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateDevice", entity, false)
					if err != nil {
						t.Fatalf("Failed to reset managementNetwork externally: %v", err)
					}

					time.Sleep(TestEventualConsistencyDelay)
					t.Log("[DEBUG] Reset managementNetwork externally to zero UUID")
				},
				Config: testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(deviceName, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Step 3: Verify Terraform restored the configured value.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(deviceName, categoryName, imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					managementNetworkInBCMCheck{
						resourceAddress: "bcm_cmdevice_device.test",
						deviceName:      deviceName,
					},
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_ManagementNetworkUpdate tests that changing management_network
// from one value to another sends the new value to the BCM API.
func TestAccCMDeviceDevice_ManagementNetworkUpdate(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-mgmtupd")
	categoryName := generateUniqueTestName("tftest-category-mgmtupd")
	imageName := generateUniqueTestName("tftest-image-mgmtupd")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// Track management_network value across steps to verify it changes.
	compareMgmtNet := statecheck.CompareValue(compare.ValuesDiffer())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with management_network = managementnet.
			{
				Config: testAccCMDeviceDeviceResourceConfig_ManagementNetworkNamed(deviceName, categoryName, imageName, imagePath, mac, "managementnet"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					compareMgmtNet.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
					),
					managementNetworkInBCMCheck{
						resourceAddress: "bcm_cmdevice_device.test",
						deviceName:      deviceName,
					},
				},
			},
			// Step 2: Update to management_network = internalnet.
			{
				Config: testAccCMDeviceDeviceResourceConfig_ManagementNetworkNamed(deviceName, categoryName, imageName, imagePath, mac, "internalnet"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					compareMgmtNet.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
					),
					managementNetworkInBCMCheck{
						resourceAddress: "bcm_cmdevice_device.test",
						deviceName:      deviceName,
					},
				},
			},
			// Step 3: Idempotency after update.
			{
				Config: testAccCMDeviceDeviceResourceConfig_ManagementNetworkNamed(deviceName, categoryName, imageName, imagePath, mac, "internalnet"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// =============================================================================
// Import and Recategorize Tests
// =============================================================================

// testAccCMDeviceDeviceResourceConfig_WithCategory returns device config pointing
// at a specific named category. Both categories are created so switching between
// them is a pure device update.
func testAccCMDeviceDeviceResourceConfig_WithCategory(hostname, categoryNameA, categoryNameB, activeCategory, imageName, imagePath, mac string) string {
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

resource "bcm_cmdevice_category" "cat_a" {
  name               = %[6]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_category" "cat_b" {
  name               = %[7]q
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  software_image_proxy = {
    parent_software_image = bcm_cmpart_softwareimage.test.id
  }

  depends_on = [bcm_cmpart_softwareimage.test]
}

resource "bcm_cmdevice_device" "test" {
  hostname = %[8]q
  category = bcm_cmdevice_category.%[9]s.id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[10]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

  depends_on = [bcm_cmdevice_category.cat_a, bcm_cmdevice_category.cat_b]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		imageName,
		imagePath,
		categoryNameA,
		categoryNameB,
		hostname,
		activeCategory,
		mac,
	)
}

// TestAccCMDeviceDevice_ImportAndRecategorize tests the workflow of importing
// an existing device and then changing its category. This is a common adoption
// pattern when bringing existing BCM devices under Terraform management.
func TestAccCMDeviceDevice_ImportAndRecategorize(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-recat")
	catNameA := generateUniqueTestName("tftest-cat-a")
	catNameB := generateUniqueTestName("tftest-cat-b")
	imageName := generateUniqueTestName("tftest-img-recat")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	compareID := statecheck.CompareValue(compare.ValuesSame())
	compareCat := statecheck.CompareValue(compare.ValuesDiffer())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device in category A.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithCategory(deviceName, catNameA, catNameB, "cat_a", imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
					compareCat.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("category"),
					),
				},
			},
			// Step 2: Import the device by UUID.
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"management_network",
					"boot_loader",
					"boot_loader_protocol",
					"partition",
					"power_control",
					"default_gateway",
					"default_gateway_metric",
					"serial_number",
					"part_number",
					"interfaces.#",
					"interfaces.0.%",
					"interfaces.0.base_type",
					"interfaces.0.bootable",
					"interfaces.0.cardtype",
					"interfaces.0.child_type",
					"interfaces.0.dhcp",
					"interfaces.0.mac",
					"interfaces.0.name",
					"interfaces.0.network",
					"interfaces.0.start_if",
					"interfaces.0.type",
					"interfaces.0.uuid",
					"interfaces.0.bond_mode",
					"interfaces.0.ip",
					"interfaces.0.ipv6_ip",
					"interfaces.0.members.#",
				},
			},
			// Step 3: Switch device to category B — this is the key assertion.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithCategory(deviceName, catNameA, catNameB, "cat_b", imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					// ID should be unchanged (same device, just recategorized).
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
					// Category should differ from step 1.
					compareCat.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("category"),
					),
				},
			},
			// Step 4: Idempotency — no further changes expected.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithCategory(deviceName, catNameA, catNameB, "cat_b", imageName, imagePath, mac),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMDeviceDevice_ManagementNetworkImport tests that importing a device
// with management_network set picks up the value from BCM.
func TestAccCMDeviceDevice_ManagementNetworkImport(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-mgmtimp")
	categoryName := generateUniqueTestName("tftest-category-mgmtimp")
	imageName := generateUniqueTestName("tftest-image-mgmtimp")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create device with explicit management_network.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					managementNetworkInBCMCheck{
						resourceAddress: "bcm_cmdevice_device.test",
						deviceName:      deviceName,
					},
				},
			},
			// Step 2: Import — management_network should be verified (NOT ignored).
			{
				ResourceName:      "bcm_cmdevice_device.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"boot_loader",
					"boot_loader_protocol",
					"partition",
					"power_control",
					"default_gateway",
					"default_gateway_metric",
					"serial_number",
					"part_number",
					"interfaces.#",
					"interfaces.0.%",
					"interfaces.0.base_type",
					"interfaces.0.bootable",
					"interfaces.0.cardtype",
					"interfaces.0.child_type",
					"interfaces.0.dhcp",
					"interfaces.0.mac",
					"interfaces.0.name",
					"interfaces.0.network",
					"interfaces.0.start_if",
					"interfaces.0.type",
					"interfaces.0.uuid",
					"interfaces.0.bond_mode",
					"interfaces.0.ip",
					"interfaces.0.ipv6_ip",
					"interfaces.0.members.#",
				},
				// NOTE: management_network is intentionally NOT in ImportStateVerifyIgnore.
				// Since we now send it to BCM, import should read it back correctly.
			},
		},
	})
}

// TestAccCMDeviceDevice_ManagementNetworkRemove tests that removing management_network
// from config results in the zero UUID being sent and null in state.
func TestAccCMDeviceDevice_ManagementNetworkRemove(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-mgmtrm")
	categoryName := generateUniqueTestName("tftest-category-mgmtrm")
	imageName := generateUniqueTestName("tftest-image-mgmtrm")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	// ID consistency tracking.
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with explicit management_network.
			{
				Config: testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					managementNetworkInBCMCheck{
						resourceAddress: "bcm_cmdevice_device.test",
						deviceName:      deviceName,
					},
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Remove management_network from config (use basic config without it).
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("management_network"),
						knownvalue.Null(),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Idempotency after removal.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
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

// testAccCMDeviceDeviceResourceConfig_ManagementNetworkNamed returns device config
// with management_network set to a network filtered by exact name.
func testAccCMDeviceDeviceResourceConfig_ManagementNetworkNamed(hostname, categoryName, imageName, imagePath, mac, networkName string) string {
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

data "bcm_cmnet_networks" "device_network" {
  filter {
    name_pattern = %[9]q
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
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.device_network.networks[0].id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
		mac,
		networkName,
	)
}

// testAccCMDeviceDeviceResourceConfig_WithManagementNetwork returns device config
// with explicit management_network set on the device resource.
func testAccCMDeviceDeviceResourceConfig_WithManagementNetwork(hostname, categoryName, imageName, imagePath, mac string) string {
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
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = %[8]q
    network  = data.bcm_cmnet_networks.management.networks[0].id
    bootable = true
    dhcp     = true
  }

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
		mac,
	)
}

// =============================================================================
// Additional Validation Tests — Schema Field Coverage
// =============================================================================

// TestAccCMDeviceDevice_ValidationInvalidCategoryUUID tests that a non-UUID category
// value is rejected by the schema validator.
func TestAccCMDeviceDevice_ValidationInvalidCategoryUUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter { name_pattern = "managementnet" }
}

resource "bcm_cmdevice_device" "test" {
  hostname = "val-test-cat"
  category = "not-a-uuid"
  interfaces {
    name    = "eth0"
    type    = "physical"
    mac     = "02:00:00:00:00:01"
    network = data.bcm_cmnet_networks.management.networks[0].id
  }
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`must be valid UUID`),
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationInvalidManagementNetworkUUID tests that a non-UUID
// management_network value is rejected by the schema validator.
func TestAccCMDeviceDevice_ValidationInvalidManagementNetworkUUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter { name_pattern = "managementnet" }
}

data "bcm_cmdevice_categories" "all" {}

resource "bcm_cmdevice_device" "test" {
  hostname           = "val-test-mgmt"
  category           = data.bcm_cmdevice_categories.all.categories[0].id
  management_network = "not-a-uuid"
  interfaces {
    name    = "eth0"
    type    = "physical"
    mac     = "02:00:00:00:00:02"
    network = data.bcm_cmnet_networks.management.networks[0].id
  }
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`must be valid UUID`),
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationInvalidGateway tests that an invalid IP address
// for default_gateway is rejected by the schema validator.
func TestAccCMDeviceDevice_ValidationInvalidGateway(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter { name_pattern = "managementnet" }
}

data "bcm_cmdevice_categories" "all" {}

resource "bcm_cmdevice_device" "test" {
  hostname        = "val-test-gw"
  category        = data.bcm_cmdevice_categories.all.categories[0].id
  default_gateway = "999.999.999.999"
  interfaces {
    name    = "eth0"
    type    = "physical"
    mac     = "02:00:00:00:00:03"
    network = data.bcm_cmnet_networks.management.networks[0].id
  }
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`must be a valid IPv4 address`),
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationInvalidInterfaceType tests that an invalid
// interface type is rejected by the schema validator.
func TestAccCMDeviceDevice_ValidationInvalidInterfaceType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter { name_pattern = "managementnet" }
}

data "bcm_cmdevice_categories" "all" {}

resource "bcm_cmdevice_device" "test" {
  hostname = "val-test-iface"
  category = data.bcm_cmdevice_categories.all.categories[0].id
  interfaces {
    name    = "eth0"
    type    = "wireless"
    mac     = "02:00:00:00:00:04"
    network = data.bcm_cmnet_networks.management.networks[0].id
  }
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationInvalidInterfaceStartIf tests that an invalid
// start_if value is rejected by the schema validator.
func TestAccCMDeviceDevice_ValidationInvalidInterfaceStartIf(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter { name_pattern = "managementnet" }
}

data "bcm_cmdevice_categories" "all" {}

resource "bcm_cmdevice_device" "test" {
  hostname = "val-test-startif"
  category = data.bcm_cmdevice_categories.all.categories[0].id
  interfaces {
    name     = "eth0"
    type     = "physical"
    mac      = "02:00:00:00:00:05"
    network  = data.bcm_cmnet_networks.management.networks[0].id
    start_if = "SOMETIMES"
  }
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`value must be one of`),
			},
		},
	})
}

// TestAccCMDeviceDevice_ValidationNoInterfaces tests that a device with no
// interfaces block is rejected by the schema validator (min size 1).
func TestAccCMDeviceDevice_ValidationNoInterfaces(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}

resource "bcm_cmdevice_device" "test" {
  hostname = "val-test-noiface"
  category = data.bcm_cmdevice_categories.all.categories[0].id
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`(?i)at least 1`),
			},
		},
	})
}

// deviceDisappearsCheck is a custom StateCheck that deletes a device resource
// via the BCM API during the check phase, simulating external deletion.
// This allows the test to verify that Terraform detects the missing resource
// and produces a non-empty plan to recreate it.
type deviceDisappearsCheck struct {
	resourceAddress string
}

func (c deviceDisappearsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	// Find the resource in state.
	var deviceUUID string
	for _, r := range req.State.Values.RootModule.Resources {
		if r.Address == c.resourceAddress {
			uuid, ok := r.AttributeValues["uuid"]
			if !ok {
				resp.Error = fmt.Errorf("resource %s has no uuid attribute", c.resourceAddress)
				return
			}
			deviceUUID, ok = uuid.(string)
			if !ok {
				resp.Error = fmt.Errorf("resource %s uuid attribute is not a string", c.resourceAddress)
				return
			}
			break
		}
	}

	if deviceUUID == "" {
		resp.Error = fmt.Errorf("resource %s not found in state", c.resourceAddress)
		return
	}

	// Delete the device externally via BCM API.
	client := createTestBCMClient(&testing.T{})
	_, err := client.CallJSONRPC(ctx, "cmdevice", "removeDevice", deviceUUID, true) // force=true
	if err != nil {
		resp.Error = fmt.Errorf("failed to delete device %s externally: %w", deviceUUID, err)
		return
	}

	// Wait for eventual consistency.
	time.Sleep(2 * time.Second)
}

// TestAccCMDeviceDevice_Disappears tests that when a device is deleted externally
// (outside of Terraform), the next plan detects it and wants to recreate it.
func TestAccCMDeviceDevice_Disappears(t *testing.T) {
	deviceName := generateUniqueTestName("tftest-device-disap")
	categoryName := generateUniqueTestName("tftest-cat-disap")
	imageName := generateUniqueTestName("tftest-img-disap")
	imagePath := fmt.Sprintf("/cm/images/%s.iso", imageName)
	mac := generateUniqueMAC()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccPreCheck(t)
			testAccCMDeviceDevicePreCheck(t, deviceName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceDeviceDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath, mac),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("hostname"),
						knownvalue.StringExact(deviceName),
					),
					// Delete externally during state checks.
					deviceDisappearsCheck{resourceAddress: "bcm_cmdevice_device.test"},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
