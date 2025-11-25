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

// TestAccCMUserUsersDataSource_Basic tests basic user retrieval without filters.
func TestAccCMUserUsersDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUsersDataSourceConfigBasic(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify computed id field
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					// Verify users list exists (cannot verify exact count - environment portable)
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMUserUsersDataSource_FilterUsername tests username pattern filtering.
func TestAccCMUserUsersDataSource_FilterUsername(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUsersDataSourceConfigFilterUsername("cms*"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify computed id field
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					// Verify users list exists (filtered results may be empty or populated)
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMUserUsersDataSource_FilterGroupID tests group_id filtering.
func TestAccCMUserUsersDataSource_FilterGroupID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUsersDataSourceConfigFilterGroupID("1000"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify computed id field
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					// Verify users list exists
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMUserUsersDataSource_FilterUserID tests user_id filtering.
func TestAccCMUserUsersDataSource_FilterUserID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUsersDataSourceConfigFilterUserID("1000"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify computed id field
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					// Verify users list exists
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMUserUsersDataSource_NestedAttributes tests Unix attribute population.
func TestAccCMUserUsersDataSource_NestedAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUsersDataSourceConfigBasic(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify first user has expected Unix attributes
					// Note: Cannot verify exact values due to environment portability
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("username"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("user_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("group_id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("home_directory"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("login_shell"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMUserUsersDataSource_AccountActive tests account_active computation.
func TestAccCMUserUsersDataSource_AccountActive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUsersDataSourceConfigBasic(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify account_active is computed (bool, not null)
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("account_active"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// Test helper functions

func testAccCMUserUsersDataSourceConfigBasic() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

func testAccCMUserUsersDataSourceConfigFilterUsername(pattern string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
  username_pattern = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		pattern,
	)
}

func testAccCMUserUsersDataSourceConfigFilterGroupID(groupID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
  group_id = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		groupID,
	)
}

func testAccCMUserUsersDataSourceConfigFilterUserID(userID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
  user_id = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		userID,
	)
}
