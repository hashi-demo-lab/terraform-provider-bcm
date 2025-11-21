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
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// Test helper: CheckDestroy verifies all categories are deleted
// Enhanced with resource counter, timeouts, detailed error messages, and logging
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
// Refactored to use shared verifyResourceDeleted helper with exponential backoff (standardized retry config: 5 retries)
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
					deleted, err := verifyResourceDeleted(context.Background(), client, "cmdevice", "getCategory", name, 5)
					if err != nil {
						t.Logf("Error verifying deletion of %s: %v", name, err)
					} else if deleted {
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
// This is the primary test for User Story 1 (MVP)
func TestAccCMDeviceCategoryResource_Basic(t *testing.T) {
	// Generate unique test name to avoid conflicts
	// Note: We use the SAME name for create and update to avoid BCM category name immutability issues
	categoryName := generateUniqueTestName("test-category")

	// Clean up any leftover categories before running test
	testAccCMDeviceCategoryPreCheck(t, categoryName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMDeviceCategoryResourceConfig(categoryName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "id"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "uuid"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "management_network"),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "base_type", "Category"),
				),
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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "uuid"),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Updated test category"),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "kernel_parameters", "quiet splash console=ttyS0"),
				),
			},
			// Delete testing automatically occurs in TestCase
		},
	})
}

// Test configuration helper for basic category
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

// Test configuration helper for updated category
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

// T031-T032: Import acceptance test with ImportState step
func TestAccCMDeviceCategoryResource_Import(t *testing.T) {
	categoryName := fmt.Sprintf("test-import-category-%d", time.Now().Unix())

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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "id"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "uuid"),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Test category for acceptance testing"),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "kernel_parameters", "quiet splash"),
				),
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
// This test validates that the force parameter is accepted and processed correctly
func TestAccCMDeviceCategoryResource_ForceParameter(t *testing.T) {
	categoryName := fmt.Sprintf("test-force-param-%d", time.Now().Unix())

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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "force", "false"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "id"),
				),
			},
			// Update with force=true
			{
				Config: testAccCMDeviceCategoryResourceConfig_Force(categoryName, true),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "force", "true"),
				),
			},
			// Note: Testing actual "category in use" scenario requires manual node assignment
			// This test validates the force parameter is accepted and processed in all operations
			// Delete automatically occurs with force=true from final config
		},
	})
}

// T040: Test configuration helper with force parameter
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
// Phase 3 - Task T012 (RED): This test should FAIL initially (no PreConfig implementation yet)
func TestAccCMDeviceCategory_DriftNotes(t *testing.T) {
	categoryName := generateUniqueTestName("test-drift-notes")

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
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Production"),
					resource.TestCheckResourceAttrSet("bcm_cmdevice_category.test", "uuid"),
				),
			},
			// Step 2: Modify notes externally via BCM API, verify drift detected
			{
				PreConfig: func() {
					// TODO T014: Implement PreConfig to modify via BCM API
					// This will:
					// 1. Create BCM client with createTestBCMClient(t)
					// 2. Query BCM API to get UUID by category name
					// 3. Call updateCategory API to change notes to "Staging"
					// For now, this will fail with "no PreConfig logic implemented"
				},
				Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// After external modification, Terraform should detect drift
					// State should reflect the BCM API value (modified value)
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Staging"),
				),
				// CRITICAL: This tells Terraform we expect the plan to be non-empty
				// because drift was detected
				ExpectNonEmptyPlan: true,
			},
			// Step 3: Restore desired state (Terraform applies config to fix drift)
			{
				Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Production"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify drift was corrected and state matches config
					resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Production"),
				),
			},
		},
	})
}

// Helper function for drift detection test configuration
// Phase 3 - Task T016: Config function with parameterized notes
func testAccCMDeviceCategoryResourceConfig_DriftNotes(name, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup required parent software image
data "bcm_cmpart_softwareimages" "default" {}

locals {
  software_image_uuid = [for img in data.bcm_cmpart_softwareimages.default.images : img.uuid if img.name == "default-image"][0]
}

# Lookup management network
data "bcm_cmnet_networks" "all" {}

locals {
  management_network_uuid = data.bcm_cmnet_networks.all.networks[0].uuid
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
