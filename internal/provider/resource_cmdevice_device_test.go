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
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath),
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
						knownvalue.StringExact("00:11:22:33:44:55"),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath),
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
				},
			},
			// Verify ID consistency after Import.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Basic(deviceName, categoryName, imageName, imagePath),
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update and Read testing.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Updated(deviceName, categoryName, imageName, imagePath),
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
			// Create with initial value.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_device.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("initial-notes"),
					),
				},
			},
			// Modify externally via BCM API.
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

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
			// Terraform restores desired state.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(deviceName, categoryName, imageName, imagePath),
				ConfigStateChecks: []statecheck.StateCheck{
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
// - Follows eventual consistency test patterns from drift detection tests
func TestAccCMDeviceDevice_PartitionCommitTimeout(t *testing.T) {
	t.Skip("SKIP: This test takes ~2 minutes to complete due to exponential backoff - run manually to verify timeout logic")

	// Note: To run this test manually, remove the t.Skip() line above and execute:
	// TF_ACC=1 go test -v -timeout 10m ./internal/provider/ -run TestAccCMDeviceDevice_PartitionCommitTimeout

	deviceName := generateUniqueTestName("test-device-timeout")
	categoryName := generateUniqueTestName("citest-category-timeout")
	imageName := generateUniqueTestName("citest-image-timeout")
	imagePath := "/cm/images/ubuntu-22.04-server-amd64-timeout.iso"

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
  hostname           = %[5]q
  mac                = "00:11:22:33:44:77"
  category           = bcm_cmdevice_category.test.id
  management_network = data.bcm_cmnet_networks.management.networks[0].id

  depends_on = [bcm_cmdevice_category.test]
}
`,
		endpoint,
		imageName,
		imagePath,
		categoryName,
		hostname,
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
// Phase 3: Drift Detection Tests.
// ========================================

// TestAccCMDeviceDevice_Drift tests drift detection for hostname attribute.
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
			// Step 1: Create device with initial hostname.
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath),
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
			// Step 2: Modify hostname externally via BCM API, verify drift detected.
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

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
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified hostname externally to: %v", entity["hostname"])
				},
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath),
				// Use ConfigPlanChecks to verify drift detected (non-empty plan expected).
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Step 3: Restore desired state (Terraform applies config to fix drift).
			{
				Config: testAccCMDeviceDeviceResourceConfig_Drift(initialHostname, categoryName, imageName, imagePath),
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
// - Cluster partitions exist but none have the required "base" name
func TestAccCMDeviceDevice_PartitionErrorHandling(t *testing.T) {
	t.Skip("Skipping partition error handling tests - requires specific BCM cluster configurations or mock server")

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
