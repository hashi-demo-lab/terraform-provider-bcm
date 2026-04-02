// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

func TestAccCMDeviceNodesDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmdevice-nodes"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("hostname"),
						knownvalue.StringExact("bcm11-headnode"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("child_type"),
						knownvalue.StringExact("HeadNode"),
					),
				},
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmdevice-nodes"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("hostname"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("child_type"),
						knownvalue.StringExact("PhysicalNode"),
					),
				},
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
    child_type = "PhysicalNode"
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
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmdevice-nodes"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("hostname"),
						knownvalue.StringExact("node001"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("child_type"),
						knownvalue.StringExact("PhysicalNode"),
					),
				},
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
    hostname_pattern = "node001"
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

func TestAccCMDeviceNodesDataSource_FilterMultiple(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceNodesDataSourceConfig_filterMultiple(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmdevice-nodes"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("child_type"),
						knownvalue.StringExact("PhysicalNode"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_nodes.test",
						tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("hostname"),
						knownvalue.StringExact("node001"),
					),
				},
			},
		},
	})
}

func testAccCMDeviceNodesDataSourceConfig_filterMultiple() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_nodes" "test" {
  filter {
    child_type        = "PhysicalNode"
    hostname_pattern  = "node001"
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}
