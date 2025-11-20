// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCMDeviceNodesDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "id"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.#"),
				),
			},
		},
	})
}

func testAccCMDeviceNodesDataSourceConfig_basic() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_nodes" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

func TestAccCMDeviceNodesDataSource_FilterByType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_filterType(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "id"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.#"),
				),
			},
		},
	})
}

func testAccCMDeviceNodesDataSourceConfig_filterType() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_nodes" "test" {
  filter {
    node_type = "PhysicalNode"
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

func TestAccCMDeviceNodesDataSource_FilterByHostname(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_filterHostname(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "id"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.#"),
				),
			},
		},
	})
}

func testAccCMDeviceNodesDataSourceConfig_filterHostname() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_nodes" "test" {
  filter {
    hostname_pattern = "node"
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

func TestAccCMDeviceNodesDataSource_NestedAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "id"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.#"),
					// Verify nested attributes exist (will validate structure in REFACTOR phase)
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.0.interfaces.#"),
					resource.TestCheckResourceAttrSet("data.bcm_cmdevice_nodes.test", "nodes.0.roles.#"),
				),
			},
		},
	})
}
