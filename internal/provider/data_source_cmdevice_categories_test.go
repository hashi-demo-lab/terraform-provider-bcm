// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCMDeviceCategoriesDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceCategoriesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify at least one category exists
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.#"),
					// Verify first category has required attributes
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.uuid"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.name"),
					resource.TestCheckResourceAttr("data.bcm_cmdevice_categories.test", "categories.0.base_type", "Category"),
					// Verify software image reference
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.software_image_id"),
					// Verify management network
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.management_network_id"),
				),
			},
		},
	})
}

func TestAccCMDeviceCategoriesDataSource_FilterByName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceCategoriesDataSourceConfigFilterByName("default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify we got exactly one category
					resource.TestCheckResourceAttr("data.bcm_cmdevice_categories.test", "categories.#", "1"),
					// Verify it's the default category
					resource.TestCheckResourceAttr("data.bcm_cmdevice_categories.test", "categories.0.name", "default"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.uuid"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.software_image_id"),
					// Verify boot configuration
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.boot_loader"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.install_mode"),
				),
			},
		},
	})
}

func TestAccCMDeviceCategoriesDataSource_NestedAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceCategoriesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify nested list attributes exist by checking their count
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.modules.#"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.fsmounts.#"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.roles.#"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.services.#"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.name_servers.#"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.search_domains.#"),
				),
			},
		},
	})
}

func TestAccCMDeviceCategoriesDataSource_DiskSetup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceCategoriesDataSourceConfigFilterByName("default"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify disksetup exists and contains XML
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_categories.test", "categories.0.disksetup"),
				),
			},
		},
	})
}

// testAccCMDeviceCategoriesDataSourceConfig returns a basic configuration
func testAccCMDeviceCategoriesDataSourceConfig() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccCMDeviceCategoriesDataSourceConfigFilterByName returns a configuration with name filter
func testAccCMDeviceCategoriesDataSourceConfigFilterByName(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "test" {
  name = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}
