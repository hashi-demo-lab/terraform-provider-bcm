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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

// generateUniqueTestName creates a unique test resource name with timestamp suffix
// This prevents collisions from parallel test runs or incomplete cleanup
func generateUniqueTestName(prefix string) string {
	timestamp := time.Now().Unix()
	return fmt.Sprintf("%s-%d", prefix, timestamp)
}

// Test helper: CheckDestroy verifies all software images are deleted
func testAccCheckCMPartSoftwareImageDestroy(s *terraform.State) error {
	// Create BCM client using environment variables
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		return fmt.Errorf("BCM credentials not set")
	}

	client, err := NewBCMClient(context.Background(), endpoint, username, password, true, 30)
	if err != nil {
		return fmt.Errorf("failed to create BCM client: %w", err)
	}

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmpart_softwareimage" {
			continue
		}

		// Attempt to read the software image
		name := rs.Primary.Attributes["name"]
		body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", name)

		// If no error and response contains data, resource still exists
		if err == nil {
			var imageData map[string]interface{}
			if json.Unmarshal(body, &imageData) == nil && len(imageData) > 0 {
				return fmt.Errorf("software image %s still exists", name)
			}
		}
		// If error or empty response, resource is gone (expected)
	}

	return nil
}

// Test helper: Clean up existing test images before running tests
// Enhanced with retry logic and deletion verification
func testAccCMPartSoftwareImagePreCheck(t *testing.T, names ...string) {
	testAccPreCheck(t)

	// Create BCM client using environment variables
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	client, err := NewBCMClient(context.Background(), endpoint, username, password, true, 30)
	if err != nil {
		t.Logf("Failed to create BCM client for cleanup: %v", err)
		return
	}

	// Attempt to clean up any leftover test images with retry logic
	for _, name := range names {
		body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", name)
		if err == nil {
			var imageData map[string]interface{}
			if json.Unmarshal(body, &imageData) == nil {
				if uuid, ok := imageData["uuid"].(string); ok && uuid != "" {
					// Image exists, try to delete it
					_, err := client.CallJSONRPC(context.Background(), "CMPart", "removeSoftwareImage", uuid, false, false, false)
					if err != nil {
						t.Logf("Failed to delete leftover image %s: %v", name, err)
						continue
					}

					// Wait for deletion to complete with exponential backoff
					maxRetries := 5
					waitTime := 1 * time.Second
					deleted := false

					for retry := 0; retry < maxRetries; retry++ {
						time.Sleep(waitTime)

						// Verify image is gone
						body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", name)
						if err != nil || len(body) == 0 {
							// Image not found - deletion successful
							deleted = true
							t.Logf("✓ Cleaned up leftover test image: %s (verified after %v)", name, waitTime*(1<<retry))
							break
						}

						// Check if response is empty object
						var checkData map[string]interface{}
						if json.Unmarshal(body, &checkData) == nil && len(checkData) == 0 {
							deleted = true
							t.Logf("✓ Cleaned up leftover test image: %s (verified after %v)", name, waitTime*(1<<retry))
							break
						}

						// Image still exists, wait longer
						waitTime *= 2
					}

					if !deleted {
						t.Logf("⚠ Warning: Image %s may not be fully deleted after %d retries", name, maxRetries)
					}
				}
			}
		}
	}
}

// Test helper functions for provider configuration
func testAccCMPartSoftwareImageResourceConfig_Basic(name, path string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup default-image properties dynamically
data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q

  # When cloning, kernel_version is inherited from original_image
  original_image = local.default_image.uuid  # Clone from default-image
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		path,
	)
}

func testAccCMPartSoftwareImageResourceConfig_Full(name, path string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup default-image properties dynamically
data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q

  # When cloning, kernel_version is inherited from original_image
  kernel_parameters     = "quiet splash"
  kernel_output_console = "tty0"

  enable_sol       = true
  sol_port         = "ttyS1"
  sol_speed        = "115200"
  sol_flow_control = true

  notes          = "Test image with full configuration"
  original_image = local.default_image.uuid  # Clone from default-image
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		path,
	)
}

func testAccCMPartSoftwareImageResourceConfig_Modules(name, path string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup default-image properties dynamically
data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q

  # When cloning, kernel_version is inherited from original_image
  original_image = local.default_image.uuid  # Clone from default-image

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "e1000e"
      parameters = "debug=1"
    }
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		path,
	)
}

func testAccCMPartSoftwareImageResourceConfig_ModulesUpdated(name, path string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup default-image properties dynamically
data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q

  # When cloning, kernel_version is inherited from original_image
  original_image = local.default_image.uuid  # Clone from default-image

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "mlx5_core"
      parameters = "enable_roce=1"
    }
  ]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		path,
	)
}

func testAccCMPartSoftwareImageResourceConfig_UpdateKernel(name, path string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup default-image properties dynamically
data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q

  # When cloning, kernel_version is inherited from original_image
  kernel_parameters = "quiet splash nomodeset"
  original_image = local.default_image.uuid  # Clone from default-image
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		path,
	)
}

func testAccCMPartSoftwareImageResourceConfig_UpdateSOL(name, path string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup default-image properties dynamically
data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

resource "bcm_cmpart_softwareimage" "test" {
  name = %[4]q
  path = %[5]q

  # When cloning, kernel_version is inherited from original_image
  original_image = local.default_image.uuid  # Clone from default-image

  enable_sol       = true
  sol_port         = "ttyS0"
  sol_speed        = "57600"
  sol_flow_control = false
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		path,
	)
}

// Phase 2 RED: Failing Acceptance Tests
// These tests MUST fail initially with "resource type not found" error

// T036: US1 - Create Software Image (Basic)
func TestAccCMPartSoftwareImageResource_Basic(t *testing.T) {
	imageName := generateUniqueTestName("test-basic-image")
	imagePath := fmt.Sprintf("/cm/images/%s", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, imageName) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "path", imagePath),
					// kernel_version is inherited from original_image during clone
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "id"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "creation_time"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "bcm_cmpart_softwareimage.test",
				ImportState:       true,
				ImportStateVerify: true, // REFACTOR phase: Real API integration enabled
				// BCM resets original_image to all zeros after cloning completes
				ImportStateVerifyIgnore: []string{"original_image"},
			},
		},
	})
}

// T037: US1 - Create Software Image (Full Config)
// Fixed: Use two-step pattern to work around BCM API kernel validation timing
func TestAccCMPartSoftwareImageResource_FullConfig(t *testing.T) {
	imageName := generateUniqueTestName("test-full-config")
	imagePath := fmt.Sprintf("/cm/images/%s", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, imageName) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with basic config (works around BCM API validation timing)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
				),
			},
			// Step 2: Update with full config (kernel params + SOL settings)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Full(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "enable_sol", "true"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "sol_speed", "115200"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "sol_flow_control", "true"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
				),
			},
			// ImportState testing
			{
				ResourceName:      "bcm_cmpart_softwareimage.test",
				ImportState:       true,
				ImportStateVerify: true, // REFACTOR phase: Real API integration enabled
				// BCM resets original_image to all zeros after cloning completes
				ImportStateVerifyIgnore: []string{"original_image"},
			},
		},
	})
}

// T041: US3 - Update Kernel Configuration
func TestAccCMPartSoftwareImageResource_UpdateKernelConfig(t *testing.T) {
	imageName := generateUniqueTestName("test-update-kernel")
	imagePath := fmt.Sprintf("/cm/images/%s", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, imageName) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Create initial resource
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
				),
			},
			// Update kernel configuration (parameters)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_UpdateKernel(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash nomodeset"),
				),
			},
		},
	})
}

// T042: US3 - Update Modules List
// Fixed: Use two-step pattern to work around BCM API kernel validation timing
func TestAccCMPartSoftwareImageResource_UpdateModules(t *testing.T) {
	imageName := generateUniqueTestName("test-update-modules")
	imagePath := fmt.Sprintf("/cm/images/%s", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, imageName) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with basic config (works around BCM API validation timing)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
				),
			},
			// Step 2: Add initial modules
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Modules(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.#", "2"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.0.name", "nvidia-drm"),
				),
			},
			// Step 3: Update modules list (remove e1000e, add mlx5_core)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_ModulesUpdated(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.#", "2"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.1.name", "mlx5_core"),
				),
			},
		},
	})
}

// T043: US3 - Update SOL Settings
func TestAccCMPartSoftwareImageResource_UpdateSOL(t *testing.T) {
	imageName := generateUniqueTestName("test-update-sol")
	imagePath := fmt.Sprintf("/cm/images/%s", imageName)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, imageName) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Create with default SOL
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
				),
			},
			// Update SOL configuration
			{
				Config: testAccCMPartSoftwareImageResourceConfig_UpdateSOL(imageName, imagePath),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "enable_sol", "true"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "sol_speed", "57600"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "sol_flow_control", "false"),
				),
			},
		},
	})
}

// T046a: Negative Test - Missing Required Fields
func TestAccCMPartSoftwareImageResource_MissingRequired(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  # Missing required 'name' field
  path = "/cm/images/test-missing"
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`argument "name" is required`),
			},
		},
	})
}

// T046b: Negative Test - Invalid SOL Speed
func TestAccCMPartSoftwareImageResource_InvalidSOLSpeed(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name = "test-invalid-sol"
  path = "/cm/images/test-invalid-sol"

  kernel_version = "6.8.0-51-generic"
  sol_speed      = "9999" # Invalid SOL speed
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`Invalid Attribute Value Match`),
			},
		},
	})
}

// T046c: Negative Test - Invalid Path Format
func TestAccCMPartSoftwareImageResource_InvalidPath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name = "test-invalid-path"
  path = "invalid path with spaces" # Invalid path format

  kernel_version = "6.8.0-51-generic"
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
				),
				ExpectError: regexp.MustCompile(`path must match format`),
			},
		},
	})
}

// T037: Edge Case - Unknown Value with Data Source Reference
// Tests that Unknown values during plan phase are correctly resolved to known values in state
func TestAccCMPartSoftwareImageResource_UnknownValue(t *testing.T) {
	imageName1 := generateUniqueTestName("test-base-image")
	imagePath1 := fmt.Sprintf("/cm/images/%s", imageName1)
	imageName2 := generateUniqueTestName("test-cloned-image")
	imagePath2 := fmt.Sprintf("/cm/images/%s", imageName2)

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			testAccCMPartSoftwareImagePreCheck(t, imageName1, imageName2)
		},
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			{
				// Step 1: Create base image and clone it using resource reference
				// This introduces Unknown value during plan phase for original_image
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

# Lookup default-image to use as base for first clone
data "bcm_cmpart_softwareimages" "default" {}

locals {
  default_image = [for img in data.bcm_cmpart_softwareimages.default.images : img if img.name == "default-image"][0]
}

# Create base image by cloning default-image
resource "bcm_cmpart_softwareimage" "base" {
  name = %[4]q
  path = %[5]q

  # Clone from default-image (kernel_version inherited)
  original_image = local.default_image.uuid
}

# Clone using resource reference (original_image is Unknown during plan)
# This is the key test: bcm_cmpart_softwareimage.base.uuid is Unknown during plan
resource "bcm_cmpart_softwareimage" "cloned" {
  name           = %[6]q
  path           = %[7]q

  # This creates Unknown value during plan phase
  original_image = bcm_cmpart_softwareimage.base.uuid
}
`,
					os.Getenv("BCM_ENDPOINT"),
					os.Getenv("BCM_USERNAME"),
					os.Getenv("BCM_PASSWORD"),
					imageName1,
					imagePath1,
					imageName2,
					imagePath2,
				),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify base image created
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.base", "name", imageName1),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.base", "uuid"),

					// Verify cloned image - original_image should be concrete UUID, never Unknown
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.cloned", "name", imageName2),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.cloned", "uuid"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.cloned", "original_image"),

					// Verify state doesn't contain Unknown values
					func(s *terraform.State) error {
						rs, ok := s.RootModule().Resources["bcm_cmpart_softwareimage.cloned"]
						if !ok {
							return fmt.Errorf("resource not found in state")
						}

						// Check that original_image is a concrete value, not Unknown
						originalImage := rs.Primary.Attributes["original_image"]
						if originalImage == "" {
							return fmt.Errorf("original_image is empty (should be concrete UUID)")
						}

						// Verify modules is a known list (not Unknown)
						if _, ok := rs.Primary.Attributes["modules.#"]; !ok {
							return fmt.Errorf("modules is Unknown or missing (should be known list)")
						}

						return nil
					},
				),
			},
		},
	})
}

// NOTE: Invalid/non-existent original_image UUID test removed
// BCM Behavior: BCM treats original_image as optional and creates image without cloning
// if the UUID doesn't exist or is invalid. This is permissive API behavior, not an error.
// The UnknownValue test above covers the important scenario of Unknown resolution.

// Phase 2 RED tests complete (including negative tests and edge cases)
// All tests should fail with "resource type not found", validation errors, or API errors
