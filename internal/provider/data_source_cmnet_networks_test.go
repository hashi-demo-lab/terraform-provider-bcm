// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCMNetNetworksDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all networks without filters
			{
				Config: testAccCMNetNetworksDataSourceConfig_Basic(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source ID is set
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "id"),
					// Verify networks list is populated
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.#"),
					// Verify at least one network exists (based on research findings)
					resource.TestCheckResourceAttr("data.bcm_cmnet_networks.test", "networks.#", "3"),
					// Verify first network has required attributes
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.0.id"),
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.0.uuid"),
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.0.name"),
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.0.base_address"),
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.0.base_type"),
				),
			},
		},
	})
}

func TestAccCMNetNetworksDataSource_NameFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Filter by name pattern "management" (should match "managementnet")
			{
				Config: testAccCMNetNetworksDataSourceConfig_NameFilter("management"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source ID is set
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.filtered", "id"),
					// Verify at least one network matches (managementnet exists)
					resource.TestCheckResourceAttr("data.bcm_cmnet_networks.filtered", "networks.#", "1"),
					// Verify the matched network contains "management" in name
					resource.TestCheckResourceAttr("data.bcm_cmnet_networks.filtered", "networks.0.name", "managementnet"),
				),
			},
		},
	})
}

func TestAccCMNetNetworksDataSource_DHCPFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Filter by DHCP enabled (should match networks with DHCP ranges)
			{
				Config: testAccCMNetNetworksDataSourceConfig_DHCPFilter(true),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source ID is set
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.dhcp", "id"),
					// Verify networks with DHCP are returned (managementnet and internalnet)
					resource.TestCheckResourceAttr("data.bcm_cmnet_networks.dhcp", "networks.#", "2"),
					// Verify first network has DHCP enabled
					resource.TestCheckResourceAttr("data.bcm_cmnet_networks.dhcp", "networks.0.dhcp_enabled", "true"),
				),
			},
		},
	})
}

func TestAccCMNetNetworksDataSource_NoMatch(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Filter with pattern that doesn't match any network
			{
				Config: testAccCMNetNetworksDataSourceConfig_NameFilter("nonexistent-network-pattern-xyz"),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source ID is set
					resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.filtered", "id"),
					// Verify no networks match (empty list, not an error)
					resource.TestCheckResourceAttr("data.bcm_cmnet_networks.filtered", "networks.#", "0"),
				),
			},
		},
	})
}

// Test configuration helpers

func testAccCMNetNetworksDataSourceConfig_Basic() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

func testAccCMNetNetworksDataSourceConfig_NameFilter(pattern string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "filtered" {
  filter {
    name_pattern = %[4]q
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		pattern,
	)
}

func testAccCMNetNetworksDataSourceConfig_DHCPFilter(enabled bool) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "dhcp" {
  filter {
    dhcp_enabled = %[4]t
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		enabled,
	)
}
