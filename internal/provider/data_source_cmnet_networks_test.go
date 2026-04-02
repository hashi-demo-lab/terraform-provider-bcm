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

func TestAccCMNetNetworksDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Read all networks without filters
			{
				Config: testAccCMNetNetworksDataSourceConfig_Basic(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source ID is set
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmnet-networks"),
					),
					// Verify first network has required attributes
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.test",
						tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.test",
						tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.test",
						tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.test",
						tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("base_address"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.test",
						tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("base_type"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

func TestAccCMNetNetworksDataSource_NameFilter(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Filter by name pattern "management" (should match networks containing "management")
			{
				Config: testAccCMNetNetworksDataSourceConfig_NameFilter("management"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source ID is set
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.filtered",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmnet-networks"),
					),
					// Verify filtered networks match the name pattern
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.filtered",
						tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringRegexp(regexp.MustCompile("management")),
					),
				},
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
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source ID is set
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.dhcp",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmnet-networks"),
					),
					// Verify filtered networks have DHCP enabled matching filter value
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.dhcp",
						tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("dhcp_enabled"),
						knownvalue.Bool(true),
					),
				},
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
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source ID is set
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.filtered",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmnet-networks"),
					),
					// Verify empty result list
					statecheck.ExpectKnownValue(
						"data.bcm_cmnet_networks.filtered",
						tfjsonpath.New("networks"),
						knownvalue.ListSizeExact(0),
					),
				},
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
