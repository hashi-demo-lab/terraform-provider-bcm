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

// TestAccCMDeviceRolesDataSource_All tests querying all roles without filters.
// This is User Story 1: Query All Available Roles (P1 - MVP).
func TestAccCMDeviceRolesDataSource_All(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMDeviceRolesDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("roles"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("roles"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("roles").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("roles").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceRolesDataSource_FilterByChildType tests filtering by role type.
// This is User Story 2: Filter Roles by Type (P2).
func TestAccCMDeviceRolesDataSource_FilterByChildType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test filtering by a common role type
				// Using HeadNodeRole as it's typically present on BCM clusters
				Config: testAccCMDeviceRolesDataSourceConfigFilterByChildType("HeadNodeRole"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("roles"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("roles").AtSliceIndex(0).AtMapKey("child_type"),
						knownvalue.StringExact("HeadNodeRole"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceRolesDataSource_FilterByNamePattern tests filtering by glob pattern.
// This is User Story 3: Filter Roles by Name Pattern (P3).
func TestAccCMDeviceRolesDataSource_FilterByNamePattern(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test filtering with wildcard pattern that matches all roles
				Config: testAccCMDeviceRolesDataSourceConfigFilterByNamePattern("*"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("roles"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("roles").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceRolesDataSource_CombinedFilters tests both filters together with AND logic.
// This tests User Story 3 combined filter functionality.
func TestAccCMDeviceRolesDataSource_CombinedFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test combined filters - both must match (AND logic)
				Config: testAccCMDeviceRolesDataSourceConfigCombinedFilters("*", "HeadNodeRole"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("roles"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("roles").AtSliceIndex(0).AtMapKey("child_type"),
						knownvalue.StringExact("HeadNodeRole"),
					),
				},
			},
		},
	})
}

// TestAccCMDeviceRolesDataSource_EmptyResults tests filter that returns no matches.
// This validates edge case handling for empty result sets.
func TestAccCMDeviceRolesDataSource_EmptyResults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Test filtering with non-existent type
				Config: testAccCMDeviceRolesDataSourceConfigFilterByChildType("NonExistentRoleType12345"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("roles"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmdevice_roles.test",
						tfjsonpath.New("roles"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

// testAccCMDeviceRolesDataSourceConfig returns a basic configuration for querying all roles.
func testAccCMDeviceRolesDataSourceConfig() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccCMDeviceRolesDataSourceConfigFilterByChildType returns configuration with child_type filter.
func testAccCMDeviceRolesDataSourceConfigFilterByChildType(childType string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "test" {
  child_type = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		childType,
	)
}

// testAccCMDeviceRolesDataSourceConfigFilterByNamePattern returns configuration with name_pattern filter.
func testAccCMDeviceRolesDataSourceConfigFilterByNamePattern(pattern string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "test" {
  name_pattern = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		pattern,
	)
}

// testAccCMDeviceRolesDataSourceConfigCombinedFilters returns configuration with both filters.
func testAccCMDeviceRolesDataSourceConfigCombinedFilters(namePattern, childType string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_roles" "test" {
  name_pattern = %[4]q
  child_type   = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		namePattern,
		childType,
	)
}
