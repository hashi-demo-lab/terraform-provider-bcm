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

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

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

	// Attempt to clean up any leftover test images
	for _, name := range names {
		body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", name)
		if err == nil {
			var imageData map[string]interface{}
			if json.Unmarshal(body, &imageData) == nil {
				if uuid, ok := imageData["uuid"].(string); ok && uuid != "" {
					// Image exists, try to delete it
					_, _ = client.CallJSONRPC(context.Background(), "CMPart", "removeSoftwareImage", uuid, false, false, false)
					t.Logf("Cleaned up leftover test image: %s", name)
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

  kernel_version = local.default_image.kernel_version
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

  kernel_version        = local.default_image.kernel_version  # Use kernel from default-image
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

  kernel_version = local.default_image.kernel_version  # Use kernel from default-image
  original_image = local.default_image.uuid            # Clone from default-image

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "e1000e"
      parameters = ""
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

  kernel_version = local.default_image.kernel_version  # Use kernel from default-image
  original_image = local.default_image.uuid            # Clone from default-image

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

  kernel_version    = local.default_image.kernel_version
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

  kernel_version = local.default_image.kernel_version
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
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, "test-basic-image") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic("test-basic-image", "/cm/images/test-basic-image"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-basic-image"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "path", "/cm/images/test-basic-image"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_version", "6.8.0-51-generic"),
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
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, "test-full-config") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with basic config (works around BCM API validation timing)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic("test-full-config", "/cm/images/test-full-config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-full-config"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
				),
			},
			// Step 2: Update with full config (kernel params + SOL settings)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Full("test-full-config", "/cm/images/test-full-config"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-full-config"),
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

// T038: US1 - Create Software Image (With Modules)
// Fixed: Use two-step pattern to work around BCM API kernel validation timing
func TestAccCMPartSoftwareImageResource_WithModules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, "test-with-modules") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with basic config (works around BCM API validation timing)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic("test-with-modules", "/cm/images/test-with-modules"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-with-modules"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
				),
			},
			// Step 2: Update with modules
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Modules("test-with-modules", "/cm/images/test-with-modules"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-with-modules"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.#", "2"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.0.name", "nvidia-drm"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.0.parameters", "modeset=1"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.1.name", "e1000e"),
				),
			},
		},
	})
}

// T041: US3 - Update Kernel Configuration
func TestAccCMPartSoftwareImageResource_UpdateKernelConfig(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, "test-update-kernel") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Create initial resource
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic("test-update-kernel", "/cm/images/test-update-kernel"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "kernel_version"),
				),
			},
			// Update kernel configuration (parameters)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_UpdateKernel("test-update-kernel", "/cm/images/test-update-kernel"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "kernel_version"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash nomodeset"),
				),
			},
		},
	})
}

// T042: US3 - Update Modules List
// Fixed: Use two-step pattern to work around BCM API kernel validation timing
func TestAccCMPartSoftwareImageResource_UpdateModules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, "test-update-modules") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with basic config (works around BCM API validation timing)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic("test-update-modules", "/cm/images/test-update-modules"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-update-modules"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
				),
			},
			// Step 2: Add initial modules
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Modules("test-update-modules", "/cm/images/test-update-modules"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.#", "2"),
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "modules.0.name", "nvidia-drm"),
				),
			},
			// Step 3: Update modules list (remove e1000e, add mlx5_core)
			{
				Config: testAccCMPartSoftwareImageResourceConfig_ModulesUpdated("test-update-modules", "/cm/images/test-update-modules"),
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
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccCMPartSoftwareImagePreCheck(t, "test-update-sol") },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
		Steps: []resource.TestStep{
			// Create with default SOL
			{
				Config: testAccCMPartSoftwareImageResourceConfig_Basic("test-update-sol", "/cm/images/test-update-sol"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", "test-update-sol"),
				),
			},
			// Update SOL configuration
			{
				Config: testAccCMPartSoftwareImageResourceConfig_UpdateSOL("test-update-sol", "/cm/images/test-update-sol"),
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

// Phase 2 RED tests complete (including negative tests)
// All tests should fail with "resource type not found" or validation errors
