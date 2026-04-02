// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCMPartSoftwareImagesDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Modern state verification - verify id computed field
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-softwareimages"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_EmptyResponse(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-softwareimages"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_NestedModules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfigFilterByName("default-image"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-softwareimages"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("default-image"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("modules"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("modules").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_AllFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-softwareimages"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("path"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("bootfs_part"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("fs_part"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("enable_sol"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("sol_port"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("sol_speed"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("sol_flow_control"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("creation_time"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("revision_id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_InvalidCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "bcm" {
  endpoint             = "` + os.Getenv("BCM_ENDPOINT") + `"
  username             = "invalid_user"
  password             = "invalid_password"
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {}
`,
				ExpectError: regexp.MustCompile(`(login failed|authentication|401|unauthorized)`),
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_FilterByCategory(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfigFilterByCategory("Ubuntu"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-softwareimages"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_FilterByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfigFilterByName("image"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-softwareimages"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("images").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringRegexp(regexp.MustCompile(`(?i)image`)),
					),
				},
			},
		},
	})
}

// testAccCMPartSoftwareImagesDataSourceConfig returns a basic test configuration
// for the bcm_cmpart_softwareimages data source.
func testAccCMPartSoftwareImagesDataSourceConfig() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccCMPartSoftwareImagesDataSourceConfigFilterByCategory returns a configuration with category filter.
func testAccCMPartSoftwareImagesDataSourceConfigFilterByCategory(category string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {
  filter {
    category = %[4]q
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		category,
	)
}

// testAccCMPartSoftwareImagesDataSourceConfigFilterByName returns a configuration with name pattern filter.
func testAccCMPartSoftwareImagesDataSourceConfigFilterByName(namePattern string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {
  filter {
    name_pattern = %[4]q
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		namePattern,
	)
}
