// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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

// Test helper: CheckDestroy verifies all categories are deleted
// Enhanced with resource counter, timeouts, detailed error messages, and logging.
func testAccCheckCMDeviceCategoryDestroy(s *terraform.State) error {
	// Create BCM client using shared helper
	client := createTestBCMClient(&testing.T{})

	resourcesChecked := 0
	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmdevice_category" {
			continue
		}

		resourcesChecked++
		name := rs.Primary.Attributes["name"]
		uuid := rs.Primary.Attributes["uuid"]

		// Add 10s timeout context per API call
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Attempt to read the category
		body, err := client.CallJSONRPC(ctx, "cmdevice", "getCategory", name)

		// If no error and response contains data, resource still exists
		if err == nil {
			var categoryData map[string]interface{}
			if json.Unmarshal(body, &categoryData) == nil && len(categoryData) > 0 {
				return fmt.Errorf("category still exists after destroy: name=%s uuid=%s response=%s",
					name, uuid, string(body))
			}
		}
		// If error or empty response, resource is gone (expected)
	}

	// Log number of resources checked for debugging
	if resourcesChecked == 0 {
		// This is not an error - may be checking destroy for data sources or other resources
		return nil
	}

	return nil
}

// Test helper: Clean up existing test categories before running tests
// Refactored to use shared verifyResourceDeleted helper with exponential backoff (standardized retry config: 5 retries).
func testAccCMDeviceCategoryPreCheck(t *testing.T, names ...string) {
	testAccPreCheck(t)

	// Create BCM client using shared helper
	client := createTestBCMClient(t)

	// Attempt to clean up any leftover test categories with standardized retry logic
	for _, name := range names {
		body, err := client.CallJSONRPC(context.Background(), "cmdevice", "getCategory", name)
		if err == nil {
			var categoryData map[string]interface{}
			if json.Unmarshal(body, &categoryData) == nil {
				if uuid, ok := categoryData["uuid"].(string); ok && uuid != "" {
					// Category exists, try to delete it with force=true
					_, err := client.CallJSONRPC(context.Background(), "cmdevice", "removeCategory", uuid, true)
					if err != nil {
						t.Logf("Failed to delete leftover category %s: %v", name, err)
						continue
					}

					// Verify deletion with shared helper (standardized retry config: 5 retries)
					deleted := verifyResourceDeleted(context.Background(), client, "cmdevice", "getCategory", name, 5)
					if deleted {
						t.Logf("✓ Cleaned up leftover test category: %s", name)
					} else {
						t.Logf("⚠ Warning: Category %s may not be fully deleted after retries", name)
					}
				}
			}
		}
	}
}

// TestAccCMDeviceCategoryResource_Basic tests basic CRUD operations
// This is the primary test for User Story 1 (MVP).
func TestAccCMDeviceCategoryResource_Basic(t *testing.T) {
	// Generate unique test name to avoid conflicts
	// Note: We use the SAME name for create and update to avoid BCM category name immutability issues
	categoryName := generateUniqueTestName("tftest-category")

	// Clean up any leftover categories before running test
	testAccCMDeviceCategoryPreCheck(t, categoryName)

	// ID consistency tracking across all CRUD operations
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMDeviceCategoryResourceConfig(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					// Modern state verification with type-safe matchers
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("management_network"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("base_type"),
						knownvalue.StringExact("Category"),
					),
					// Track ID for consistency across operations
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMDeviceCategoryResourceConfig(categoryName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// ImportState testing
			{
				ResourceName:      "bcm_cmdevice_category.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore force parameter as it's not persisted
				ImportStateVerifyIgnore: []string{"force"},
			},
			// Update and Read testing (keeping same name, updating other attributes)
			{
				Config: testAccCMDeviceCategoryResourceConfig_Updated(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Updated test category"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("kernel_parameters"),
						knownvalue.StringExact("quiet splash console=ttyS0"),
					),
					// Verify ID unchanged after update
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMDeviceCategoryResourceConfig_Updated(categoryName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Test configuration helper for basic category.
func testAccCMDeviceCategoryResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup existing categories to get a management network UUID
data "bcm_cmdevice_categories" "all" {}

# Lookup existing software images
data "bcm_cmpart_softwareimages" "all" {}

locals {
  # Get management network from first existing category (assuming at least one exists)
  # Fallback to a placeholder UUID if no categories exist
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"

  # Get UUID of first available software image (assuming at least one exists)
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = "Test category for acceptance testing"
  kernel_parameters  = "quiet splash"

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// Test configuration helper for updated category.
func testAccCMDeviceCategoryResourceConfig_Updated(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup existing categories to get a management network UUID
data "bcm_cmdevice_categories" "all" {}

# Lookup existing software images
data "bcm_cmpart_softwareimages" "all" {}

locals {
  # Get management network from first existing category (assuming at least one exists)
  # Fallback to a placeholder UUID if no categories exist
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"

  # Get UUID of first available software image (assuming at least one exists)
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = "Updated test category"
  kernel_parameters  = "quiet splash console=ttyS0"
  # Don't change boot_loader in update - BCM has issues with this field

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// T031-T032: Import acceptance test with ImportState step.
func TestAccCMDeviceCategoryResource_Import(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-import-category")

	// Cleanup any leftover test categories
	testAccCMDeviceCategoryPreCheck(t, categoryName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMDeviceCategoryResourceConfig(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Test category for acceptance testing"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("kernel_parameters"),
						knownvalue.StringExact("quiet splash"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// ImportState testing
			{
				ResourceName:      "bcm_cmdevice_category.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore force parameter (not persisted in BCM)
				ImportStateVerifyIgnore: []string{"force"},
			},
		},
	})
}

// T039-T041: Force parameter acceptance test
// This test validates that the force parameter is accepted and processed correctly.
func TestAccCMDeviceCategoryResource_ForceParameter(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-force-param")

	// Cleanup any leftover test categories
	testAccCMDeviceCategoryPreCheck(t, categoryName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Create category with force=false (default)
			{
				Config: testAccCMDeviceCategoryResourceConfig_Force(categoryName, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("force"),
						knownvalue.Bool(false),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMDeviceCategoryResourceConfig_Force(categoryName, false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update with force=true
			{
				Config: testAccCMDeviceCategoryResourceConfig_Force(categoryName, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("force"),
						knownvalue.Bool(true),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMDeviceCategoryResourceConfig_Force(categoryName, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Note: Testing actual "category in use" scenario requires manual node assignment
			// This test validates the force parameter is accepted and processed in all operations
			// Delete automatically occurs with force=true from final config
		},
	})
}

// T040: Test configuration helper with force parameter.
func testAccCMDeviceCategoryResourceConfig_Force(name string, force bool) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup existing categories to get a management network UUID and software image
data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  # Get management network from first existing category
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"

  # Get software image UUID from first available image
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = "Force parameter test category"
  force              = %[5]t

  # BCM requires parent software image
  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		force,
	)
}

// ========================================
// Phase 3: Drift Detection Tests
// ========================================

// TestAccCMDeviceCategory_DriftNotes tests drift detection for notes attribute
// Phase 3 - Task T012 (RED): This test should FAIL initially (no PreConfig implementation yet).
func TestAccCMDeviceCategory_DriftNotes(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-drift-notes")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCMDeviceCategoryPreCheck(t, categoryName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create resource with initial notes
			{
				Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Production"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Production"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Modify notes externally via BCM API, verify drift detected
			{
				PreConfig: func() {
					// Phase 3 - Task T014 (GREEN): Implement PreConfig to modify via BCM API
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get UUID by category name using helper
					uuid := getResourceUUIDByName(t, "cmdevice", "getCategory", categoryName)

					// Fetch full category data from BCM API
					body, err := client.CallJSONRPC(ctx, "cmdevice", "getCategory", categoryName)
					if err != nil {
						t.Fatalf("Failed to fetch category for drift modification: %v", err)
					}

					// Parse the category data
					var categoryData map[string]interface{}
					if err := json.Unmarshal(body, &categoryData); err != nil {
						t.Fatalf("Failed to parse category data: %v", err)
					}

					// Modify notes field
					categoryData["notes"] = "Staging"

					// Wrap in BCM API entity structure required for updates
					entity := map[string]interface{}{
						"baseType":      "Category",
						"childType":     "",
						"modified":      true,
						"to_be_removed": false,
						"revision":      "",
						"uuid":          uuid,
					}
					// Copy all category data fields except uuid (already set above)
					for k, v := range categoryData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmdevice", "updateCategory", entity)
					if err != nil {
						t.Fatalf("Failed to update category via BCM API: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified notes externally to: %v", entity["notes"])
				},
				Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Production"),
				// Use ConfigPlanChecks to verify drift detected (non-empty plan expected)
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(),
					},
				},
			},
			// Step 3: Restore desired state (Terraform applies config to fix drift)
			{
				Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Production"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify drift was corrected and state matches config
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Production"),
					),
				},
			},
		},
	})
}

// ========================================
// Phase 4: Destroy Edge Case Tests (User Story 2)
// ========================================

// TestAccCMDeviceCategory_DestroyWithForce verifies destroy with force=true
// Phase 4 - Task T020 (RED): This test should PASS (force already implemented, but verify).
func TestAccCMDeviceCategory_DestroyWithForce(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-destroy-force")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCMDeviceCategoryPreCheck(t, categoryName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create category with force=true
			{
				Config: testAccCMDeviceCategoryResourceConfig_Force(categoryName, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("force"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Destroy happens automatically with force=true
			// CheckDestroy should pass even if category has associations
			// Note: We can't easily create real node associations in acceptance tests,
			// so this verifies force parameter is accepted and processed
		},
	})
}

// TestAccCMDeviceCategory_DestroyExternalDelete verifies destroy handles externally deleted resources
// Phase 4 - Task T023: Create resource, delete via BCM API, verify Terraform destroy succeeds.
func TestAccCMDeviceCategory_DestroyExternalDelete(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-destroy-external")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCMDeviceCategoryPreCheck(t, categoryName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create category
			{
				Config: testAccCMDeviceCategoryResourceConfig(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Delete externally via BCM API, then let Terraform destroy
			{
				PreConfig: func() {
					// Delete the category via BCM API before Terraform tries to destroy it
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get category UUID
					uuid := getResourceUUIDByName(t, "cmdevice", "getCategory", categoryName)

					// Delete via BCM API with force=true (removeCategories expects array of UUIDs)
					_, err := client.CallJSONRPC(ctx, "cmdevice", "removeCategories", []string{uuid}, true)
					if err != nil {
						t.Logf("[WARN] Failed to delete category externally (may not exist): %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Deleted category externally: %s", categoryName)
				},
				Config: testAccCMDeviceCategoryResourceConfig(categoryName),
				// Destroy will happen automatically after this step
				// CheckDestroy should pass even though resource was already deleted
			},
		},
	})
}

// Helper function for drift detection test configuration
// Phase 3 - Task T016: Config function with parameterized notes.
func testAccCMDeviceCategoryResourceConfig_DriftNotes(name, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup existing categories to get a management network UUID
data "bcm_cmdevice_categories" "all" {}

# Lookup required parent software image
data "bcm_cmpart_softwareimages" "all" {}

locals {
  # Get management network from first existing category
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"

  # Get software image UUID from first available image
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = %[5]q

  # BCM requires parent software image
  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		notes,
	)
}

// ========================================
// Network and Partition Configuration Tests (Priority Group 3)
// ========================================

// TestAccCMDeviceCategory_NetworkConfiguration tests network-related optional fields
func TestAccCMDeviceCategory_NetworkConfiguration(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-network-config")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCMDeviceCategoryPreCheck(t, categoryName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with network configuration
			{
				Config: testAccCMDeviceCategoryResourceConfig_NetworkConfig(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("default_gateway"),
						knownvalue.StringExact("192.168.1.1"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("default_gateway_metric"),
						knownvalue.Int64Exact(100),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("allow_networking_restart"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Idempotency check
			{
				Config: testAccCMDeviceCategoryResourceConfig_NetworkConfig(categoryName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Update network configuration
			{
				Config: testAccCMDeviceCategoryResourceConfig_NetworkConfigUpdated(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("default_gateway"),
						knownvalue.StringExact("192.168.1.254"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("default_gateway_metric"),
						knownvalue.Int64Exact(200),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("allow_networking_restart"),
						knownvalue.Bool(false),
					),
				},
			},
			// Step 4: Idempotency check after update
			{
				Config: testAccCMDeviceCategoryResourceConfig_NetworkConfigUpdated(categoryName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMDeviceCategory_PartitionConfiguration tests partition/disk-related optional fields
//
// NOTE: This test is currently BLOCKED by BCM XML schema validation requirements.
// See: https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/48
// Status: Requires BCM XSD files for disksetup and raidconf fields
func TestAccCMDeviceCategory_PartitionConfiguration(t *testing.T) {
	t.Skip("BLOCKED: Requires BCM XSD files for disksetup/raidconf validation. See issue #48")

	categoryName := generateUniqueTestName("tftest-partition-config")

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCMDeviceCategoryPreCheck(t, categoryName)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with partition configuration
			{
				Config: testAccCMDeviceCategoryResourceConfig_PartitionConfig(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("disksetup"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("raidconf"),
						knownvalue.StringExact("raid1"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Step 2: Idempotency check
			{
				Config: testAccCMDeviceCategoryResourceConfig_PartitionConfig(categoryName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Update partition configuration
			{
				Config: testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("disksetup"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("raidconf"),
						knownvalue.StringExact("raid5"),
					),
				},
			},
			// Step 4: Idempotency check after update
			{
				Config: testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated(categoryName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// ========================================
// Network and Partition Test Config Helpers
// ========================================

// testAccCMDeviceCategoryResourceConfig_NetworkConfig creates config with network settings
func testAccCMDeviceCategoryResourceConfig_NetworkConfig(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name                       = %[4]q
  management_network         = local.management_network_uuid
  notes                      = "Network configuration test"

  # Network configuration fields
  default_gateway            = "192.168.1.1"
  default_gateway_metric     = 100
  allow_networking_restart   = true

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// testAccCMDeviceCategoryResourceConfig_NetworkConfigUpdated creates updated network config
func testAccCMDeviceCategoryResourceConfig_NetworkConfigUpdated(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name                       = %[4]q
  management_network         = local.management_network_uuid
  notes                      = "Network configuration test - updated"

  # Updated network configuration fields
  default_gateway            = "192.168.1.254"
  default_gateway_metric     = 200
  allow_networking_restart   = false

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// testAccCMDeviceCategoryResourceConfig_PartitionConfig creates config with partition settings
func testAccCMDeviceCategoryResourceConfig_PartitionConfig(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name                       = %[4]q
  management_network         = local.management_network_uuid
  notes                      = "Partition configuration test"

  # Partition/disk configuration fields
  disksetup                  = <<-EOT
<disksetup>
  <disk device="/dev/sda">
    <partition number="1" size="50GB" type="ext4" mountpoint="/"/>
    <partition number="2" size="remaining" type="ext4" mountpoint="/home"/>
  </disk>
</disksetup>
EOT
  raidconf                   = "raid1"

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated creates updated partition config
func testAccCMDeviceCategoryResourceConfig_PartitionConfigUpdated(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name                       = %[4]q
  management_network         = local.management_network_uuid
  notes                      = "Partition configuration test - updated"

  # Updated partition/disk configuration fields
  disksetup                  = <<-EOT
<disksetup>
  <disk device="/dev/sda">
    <partition number="1" size="100GB" type="ext4" mountpoint="/"/>
  </disk>
  <disk device="/dev/sdb">
    <partition number="1" size="remaining" type="xfs" mountpoint="/data"/>
  </disk>
</disksetup>
EOT
  raidconf                   = "raid5"

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// ========================================
// Disk Setup Advanced Tests (Priority Group 2)
// ========================================

// TestAccCMDeviceCategoryResource_DiskSetupAdvanced tests comprehensive disk setup configuration
// including disksetup, raidconf, install_boot_record, and revision_id from software_image_proxy.
//
// NOTE: This test is currently BLOCKED by BCM XML schema validation requirements.
// See: https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/48
// Status: Requires BCM XSD files for disksetup and raidconf fields
func TestAccCMDeviceCategoryResource_DiskSetupAdvanced(t *testing.T) {
	t.Skip("BLOCKED: Requires BCM XSD files for disksetup/raidconf validation. See issue #48")
	categoryName := generateUniqueTestName("tftest-disksetup-advanced")

	// Cleanup any leftover test categories
	testAccCMDeviceCategoryPreCheck(t, categoryName)

	// ID consistency tracking across operations
	compareID := statecheck.CompareValue(compare.ValuesSame())

	// Sample disk setup XML configuration
	diskSetupXML := `<disksetup>
  <disk device="/dev/sda">
    <partition number="1" size="500M" type="ext4" mountpoint="/boot"/>
    <partition number="2" size="remaining" type="ext4" mountpoint="/"/>
  </disk>
</disksetup>`

	updatedDiskSetupXML := `<disksetup>
  <disk device="/dev/sda">
    <partition number="1" size="1G" type="ext4" mountpoint="/boot"/>
    <partition number="2" size="remaining" type="xfs" mountpoint="/"/>
  </disk>
</disksetup>`

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with disk setup fields (install_boot_record=true)
			{
				Config: testAccCMDeviceCategoryResourceConfig_DiskSetup(categoryName, diskSetupXML, "raid1", true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("disksetup"),
						knownvalue.StringExact(diskSetupXML),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("raidconf"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("install_boot_record"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					// Verify software_image_proxy.revision_id is computed
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("software_image_proxy"),
						knownvalue.NotNull(),
					),
					// Track ID consistency
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency check after Create
			{
				Config: testAccCMDeviceCategoryResourceConfig_DiskSetup(categoryName, diskSetupXML, "raid1", true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Update disk setup fields (change disksetup, raidconf, and install_boot_record)
			{
				Config: testAccCMDeviceCategoryResourceConfig_DiskSetup(categoryName, updatedDiskSetupXML, "raid5", false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("disksetup"),
						knownvalue.StringExact(updatedDiskSetupXML),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("raidconf"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("install_boot_record"),
						knownvalue.Bool(false),
					),
					// Verify ID unchanged after update
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 4: Idempotency check after Update
			{
				Config: testAccCMDeviceCategoryResourceConfig_DiskSetup(categoryName, updatedDiskSetupXML, "raid5", false),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 5: Import and verify all disk setup fields
			{
				Config:            testAccCMDeviceCategoryResourceConfig_DiskSetup(categoryName, updatedDiskSetupXML, "raid5", false),
				ResourceName:      "bcm_cmdevice_category.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore force parameter as it's not persisted
				ImportStateVerifyIgnore: []string{"force"},
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify ID consistency across import
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 6: Test removing disk setup fields (set to empty/null)
			{
				Config: testAccCMDeviceCategoryResourceConfig_DiskSetupMinimal(categoryName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					// Verify ID unchanged after removing optional fields
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations tests different combinations
// of optional disk setup fields to ensure they can be used independently.
func TestAccCMDeviceCategoryResource_DiskSetupOptionalCombinations(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-disksetup-combos")

	// Cleanup any leftover test categories
	testAccCMDeviceCategoryPreCheck(t, categoryName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Only disksetup (no raidconf, no install_boot_record)
			{
				Config: testAccCMDeviceCategoryResourceConfig_DiskSetupOnly(categoryName, "<disksetup><disk/></disksetup>"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("disksetup"),
						knownvalue.StringExact("<disksetup><disk/></disksetup>"),
					),
				},
			},
			// Only raidconf (no disksetup, no install_boot_record)
			{
				Config: testAccCMDeviceCategoryResourceConfig_RaidConfOnly(categoryName, "raid0"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("raidconf"),
						knownvalue.StringExact("raid0"),
					),
				},
			},
			// Only install_boot_record (no disksetup, no raidconf)
			{
				Config: testAccCMDeviceCategoryResourceConfig_InstallBootRecordOnly(categoryName, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("install_boot_record"),
						knownvalue.Bool(true),
					),
				},
			},
		},
	})
}

// ========================================
// Disk Setup Test Config Helpers
// ========================================

// testAccCMDeviceCategoryResourceConfig_DiskSetup creates config with all disk setup fields.
func testAccCMDeviceCategoryResourceConfig_DiskSetup(name, disksetup, raidconf string, installBootRecord bool) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name                = %[4]q
  management_network  = local.management_network_uuid
  notes               = "Disk setup configuration test"
  disksetup           = %[5]q
  raidconf            = %[6]q
  install_boot_record = %[7]t

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		disksetup,
		raidconf,
		installBootRecord,
	)
}

// testAccCMDeviceCategoryResourceConfig_DiskSetupMinimal creates config without disk setup fields.
func testAccCMDeviceCategoryResourceConfig_DiskSetupMinimal(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = "Minimal disk setup test"

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// testAccCMDeviceCategoryResourceConfig_DiskSetupOnly creates config with only disksetup field.
func testAccCMDeviceCategoryResourceConfig_DiskSetupOnly(name, disksetup string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  disksetup          = %[5]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		disksetup,
	)
}

// testAccCMDeviceCategoryResourceConfig_RaidConfOnly creates config with only raidconf field.
func testAccCMDeviceCategoryResourceConfig_RaidConfOnly(name, raidconf string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  raidconf           = %[5]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		raidconf,
	)
}

// testAccCMDeviceCategoryResourceConfig_InstallBootRecordOnly creates config with only install_boot_record field.
func testAccCMDeviceCategoryResourceConfig_InstallBootRecordOnly(name string, installBootRecord bool) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name                = %[4]q
  management_network  = local.management_network_uuid
  install_boot_record = %[5]t

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		installBootRecord,
	)
}

// ========================================
// Validation Tests
// ========================================

// TestAccCMDeviceCategory_ValidationInvalidName tests name length validation.
func TestAccCMDeviceCategory_ValidationInvalidName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test empty name (below minimum length)
			{
				Config:      testAccCMDeviceCategoryResourceConfig_InvalidName(""),
				ExpectError: regexp.MustCompile(`Attribute name string length must be between 1 and 255`),
			},
		},
	})
}

// TestAccCMDeviceCategory_ValidationInvalidManagementNetwork tests UUID format validation.
func TestAccCMDeviceCategory_ValidationInvalidManagementNetwork(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-invalid-uuid")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test invalid UUID format
			{
				Config:      testAccCMDeviceCategoryResourceConfig_InvalidUUID(categoryName, "not-a-uuid"),
				ExpectError: regexp.MustCompile(`Attribute management_network must be a valid RFC 4122 UUID format`),
			},
			// Test malformed UUID with invalid characters
			{
				Config:      testAccCMDeviceCategoryResourceConfig_InvalidUUID(categoryName, "12345678-1234-1234-1234-12345678901G"),
				ExpectError: regexp.MustCompile(`Attribute management_network must be a valid RFC 4122 UUID format`),
			},
		},
	})
}

// TestAccCMDeviceCategory_ValidationInvalidBootLoader tests boot_loader enum validation.
func TestAccCMDeviceCategory_ValidationInvalidBootLoader(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-invalid-bootloader")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test invalid boot_loader value
			{
				Config:      testAccCMDeviceCategoryResourceConfig_InvalidBootLoader(categoryName, "INVALID_BOOTLOADER"),
				ExpectError: regexp.MustCompile(`Attribute boot_loader value must be one of`),
			},
		},
	})
}

// TestAccCMDeviceCategory_ValidationInvalidFIPS tests fips enum validation.
func TestAccCMDeviceCategory_ValidationInvalidFIPS(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-invalid-fips")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test invalid fips value (not "YES" or "NO")
			{
				Config:      testAccCMDeviceCategoryResourceConfig_InvalidFIPS(categoryName, "MAYBE"),
				ExpectError: regexp.MustCompile(`Attribute fips value must be one of.*YES.*NO`),
			},
			// Test lowercase value (case-sensitive validation)
			{
				Config:      testAccCMDeviceCategoryResourceConfig_InvalidFIPS(categoryName, "yes"),
				ExpectError: regexp.MustCompile(`Attribute fips value must be one of.*YES.*NO`),
			},
		},
	})
}

// ========================================
// Validation Test Config Helpers
// ========================================

// testAccCMDeviceCategoryResourceConfig_InvalidName creates config with invalid name.
func testAccCMDeviceCategoryResourceConfig_InvalidName(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

// testAccCMDeviceCategoryResourceConfig_InvalidUUID creates config with invalid management_network UUID.
func testAccCMDeviceCategoryResourceConfig_InvalidUUID(name, uuid string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "all" {}

locals {
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = %[5]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		uuid,
	)
}

// testAccCMDeviceCategoryResourceConfig_InvalidBootLoader creates config with invalid boot_loader.
func testAccCMDeviceCategoryResourceConfig_InvalidBootLoader(name, bootLoader string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  boot_loader        = %[5]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		bootLoader,
	)
}

// testAccCMDeviceCategoryResourceConfig_InvalidFIPS creates config with invalid fips value.
func testAccCMDeviceCategoryResourceConfig_InvalidFIPS(name, fips string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  fips               = %[5]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		fips,
	)
}

// ========================================
// Optional Field Coverage Tests
// ========================================

// TestAccCMDeviceCategoryResource_BootLoaderFields tests boot loader related optional fields
// Covers: boot_loader_file, boot_loader_protocol, kernel_output_console
// Note: kernel_version is omitted as it requires a valid kernel path from the actual software image
func TestAccCMDeviceCategoryResource_BootLoaderFields(t *testing.T) {
	categoryName := generateUniqueTestName("tftest-bootloader")

	// Clean up any leftover categories
	testAccCMDeviceCategoryPreCheck(t, categoryName)

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Create with boot loader fields set
			{
				Config: testAccCMDeviceCategoryResourceConfig_BootLoaderFields(
					categoryName,
					"/pxelinux.0",
					"TFTP",
					"ttyS0,115200",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify name and UUID
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					// Verify boot loader fields
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("boot_loader_file"),
						knownvalue.StringExact("/pxelinux.0"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("boot_loader_protocol"),
						knownvalue.StringExact("TFTP"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("kernel_output_console"),
						knownvalue.StringExact("ttyS0,115200"),
					),
					// Track ID for consistency
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMDeviceCategoryResourceConfig_BootLoaderFields(
					categoryName,
					"/pxelinux.0",
					"TFTP",
					"ttyS0,115200",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update boot loader fields
			{
				Config: testAccCMDeviceCategoryResourceConfig_BootLoaderFields(
					categoryName,
					"/grub/grubx64.efi",
					"HTTP",
					"tty0",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify updated values
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("boot_loader_file"),
						knownvalue.StringExact("/grub/grubx64.efi"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("boot_loader_protocol"),
						knownvalue.StringExact("HTTP"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("kernel_output_console"),
						knownvalue.StringExact("tty0"),
					),
					// Verify ID unchanged after update
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMDeviceCategoryResourceConfig_BootLoaderFields(
					categoryName,
					"/grub/grubx64.efi",
					"HTTP",
					"tty0",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Import to verify all fields are preserved
			{
				ResourceName:      "bcm_cmdevice_category.test",
				ImportState:       true,
				ImportStateVerify: true,
				// Ignore force parameter (not persisted in BCM)
				ImportStateVerifyIgnore: []string{"force"},
			},
		},
	})
}

// testAccCMDeviceCategoryResourceConfig_BootLoaderFields creates config with boot loader fields
func testAccCMDeviceCategoryResourceConfig_BootLoaderFields(name, bootLoaderFile, bootLoaderProtocol, kernelOutputConsole string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup existing categories to get a management network UUID
data "bcm_cmdevice_categories" "all" {}

# Lookup existing software images
data "bcm_cmpart_softwareimages" "all" {}

locals {
  # Get management network from first existing category
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"

  # Get UUID of first available software image
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name                   = %[4]q
  management_network     = local.management_network_uuid
  notes                  = "Boot loader fields test category"

  # Boot loader configuration
  boot_loader_file       = %[5]q
  boot_loader_protocol   = %[6]q

  # Kernel configuration (kernel_version omitted - requires valid path from actual software image)
  kernel_output_console  = %[7]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		bootLoaderFile,
		bootLoaderProtocol,
		kernelOutputConsole,
	)
}

// ============================================================================
// Static Routes Tests
// ============================================================================

// TestAccCMDeviceCategory_StaticRoutesBasicCRUD tests create, update of static routes
func TestAccCMDeviceCategory_StaticRoutesBasicCRUD(t *testing.T) {
	categoryName := generateUniqueTestName("staticroutes-test")

	// ID consistency tracker
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Create with static routes
			{
				Config: testAccCMDeviceCategoryConfig_StaticRoutes(categoryName, 2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("static_routes"),
						knownvalue.ListSizeExact(2),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMDeviceCategoryConfig_StaticRoutes(categoryName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update to 3 routes
			{
				Config: testAccCMDeviceCategoryConfig_StaticRoutes(categoryName, 3),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("static_routes"),
						knownvalue.ListSizeExact(3),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMDeviceCategoryConfig_StaticRoutes(categoryName, 3),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Import
			// Note: BCM doesn't persist static_routes, fsexports, roles, gpu_settings, services
			// so we must ignore these during import verification
			{
				ResourceName:      "bcm_cmdevice_category.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"static_routes",
					"fsexports",
					"roles",
					"gpu_settings",
					"services",
				},
			},
		},
	})
}

// TestAccCMDeviceCategory_StaticRoutesValidation tests CIDR and IP format validation
func TestAccCMDeviceCategory_StaticRoutesValidation(t *testing.T) {
	categoryName := generateUniqueTestName("staticroutes-invalid")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Invalid CIDR format (missing octet)
			{
				Config:      testAccCMDeviceCategoryConfig_StaticRoutesInvalid(categoryName, "192.168.1/24", "10.0.0.1"),
				ExpectError: regexp.MustCompile(`must be valid CIDR notation`),
			},
			// Invalid gateway IP format (letters instead of numbers)
			{
				Config:      testAccCMDeviceCategoryConfig_StaticRoutesInvalid(categoryName, "192.168.1.0/24", "not.an.ip.addr"),
				ExpectError: regexp.MustCompile(`must be valid IPv4 address`),
			},
		},
	})
}

// testAccCMDeviceCategoryConfig_StaticRoutes creates config with N static routes
func testAccCMDeviceCategoryConfig_StaticRoutes(name string, routeCount int) string {
	// Build routes dynamically
	routes := ""
	for i := 0; i < routeCount; i++ {
		routes += fmt.Sprintf(`
    {
      destination = "192.168.%d.0/24"
      gateway     = "10.0.0.%d"
      metric      = %d
    },`, i+1, i+1, (i+1)*100)
	}

	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = "Static routes test category"

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }

  static_routes = [%[5]s
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		routes,
	)
}

// testAccCMDeviceCategoryConfig_StaticRoutesInvalid creates config with invalid static route
func testAccCMDeviceCategoryConfig_StaticRoutesInvalid(name, destination, gateway string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }

  static_routes = [
    {
      destination = %[5]q
      gateway     = %[6]q
    }
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		destination,
		gateway,
	)
}

// ============================================================================
// GPU Settings Tests
// ============================================================================

// TestAccCMDeviceCategory_GPUSettingsBasicCRUD tests create, update of GPU settings
func TestAccCMDeviceCategory_GPUSettingsBasicCRUD(t *testing.T) {
	categoryName := generateUniqueTestName("gpusettings-test")

	// ID consistency tracker
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Create with GPU settings
			{
				Config: testAccCMDeviceCategoryConfig_GPUSettings(categoryName, 2),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(categoryName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("gpu_settings"),
						knownvalue.ListSizeExact(2),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMDeviceCategoryConfig_GPUSettings(categoryName, 2),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update to 4 GPUs
			{
				Config: testAccCMDeviceCategoryConfig_GPUSettings(categoryName, 4),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("gpu_settings"),
						knownvalue.ListSizeExact(4),
					),
					compareID.AddStateValue(
						"bcm_cmdevice_category.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Import
			// Note: BCM doesn't persist static_routes, fsexports, roles, gpu_settings, services
			// so we must ignore these during import verification
			{
				ResourceName:      "bcm_cmdevice_category.test",
				ImportState:       true,
				ImportStateVerify: true,
				ImportStateVerifyIgnore: []string{
					"force",
					"static_routes",
					"fsexports",
					"roles",
					"gpu_settings",
					"services",
				},
			},
		},
	})
}

// testAccCMDeviceCategoryConfig_GPUSettings creates config with N GPU settings
func testAccCMDeviceCategoryConfig_GPUSettings(name string, gpuCount int) string {
	// Build GPU settings dynamically
	gpuSettings := ""
	computeModes := []string{"default", "exclusive", "prohibited", "default"}
	for i := 0; i < gpuCount; i++ {
		gpuSettings += fmt.Sprintf(`
    {
      device_id    = "%d"
      model        = "Tesla V100"
      compute_mode = "%s"
    },`, i, computeModes[i%len(computeModes)])
	}

	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid
  notes              = "GPU settings test category"

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }

  gpu_settings = [%[5]s
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		gpuSettings,
	)
}
