// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
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
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmuser-users"),
					),
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
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmuser-users"),
					),
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
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmuser-users"),
					),
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
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmuser-users"),
					),
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

// TestAccCMUserUsersDataSource_FilterManagedUser verifies exact datasource filter behavior
// against a Terraform-managed user instead of relying on ambient BCM users.
func TestAccCMUserUsersDataSource_FilterManagedUser(t *testing.T) {
	username := generateUniqueUnixUsername()
	unixID := getAvailableTestUnixID(t)
	unixIDString := strconv.FormatInt(unixID, 10)
	groupIDString := "1000"
	homeDirectory := fmt.Sprintf("/home/%s-ds", username)
	shell := "/bin/zsh"

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUsersDataSourceConfigManagedUser(
					username,
					unixID,
					1000,
					homeDirectory,
					shell,
					fmt.Sprintf("%s*", username),
					"",
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmuser-users"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("user_id"),
						knownvalue.StringExact(unixIDString),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("group_id"),
						knownvalue.StringExact(groupIDString),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("home_directory"),
						knownvalue.StringExact(homeDirectory),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("login_shell"),
						knownvalue.StringExact(shell),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("account_active"),
						knownvalue.Bool(true),
					),
				},
			},
			{
				Config: testAccCMUserUsersDataSourceConfigManagedUser(
					username,
					unixID,
					1000,
					homeDirectory,
					shell,
					fmt.Sprintf("%s*", username),
					groupIDString,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("group_id"),
						knownvalue.StringExact(groupIDString),
					),
				},
			},
			{
				Config: testAccCMUserUsersDataSourceConfigManagedUser(
					username,
					unixID,
					1000,
					homeDirectory,
					shell,
					fmt.Sprintf("%s*", username),
					unixIDString,
					"",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
			{
				Config: testAccCMUserUsersDataSourceConfigManagedUser(
					username,
					unixID,
					1000,
					homeDirectory,
					shell,
					"",
					"",
					unixIDString,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("user_id"),
						knownvalue.StringExact(unixIDString),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmuser_users.test",
						tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// Test helper functions

func getAvailableTestUnixID(t *testing.T) int64 {
	t.Helper()

	client := createTestBCMClient(t)
	body, err := client.CallJSONRPC(context.Background(), "cmuser", "getUsers")
	if err != nil {
		t.Fatalf("failed to query BCM users for available Unix ID: %v", err)
	}

	var users []map[string]interface{}
	if err := json.Unmarshal(body, &users); err != nil {
		t.Fatalf("failed to parse BCM users response: %v", err)
	}

	usedIDs := make(map[int64]struct{})
	for _, user := range users {
		for _, field := range []string{"ID", "groupID"} {
			if id, ok := parseTestUnixID(user[field]); ok {
				usedIDs[id] = struct{}{}
			}
		}
	}

	for candidate := int64(61000); candidate <= 65534; candidate++ {
		if _, exists := usedIDs[candidate]; !exists {
			return candidate
		}
	}

	t.Fatal("no available Unix IDs found in range 61000-65534")
	return 0
}

func parseTestUnixID(value interface{}) (int64, bool) {
	if value == nil {
		return 0, false
	}

	id, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
	if err != nil {
		return 0, false
	}

	return id, true
}

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

func testAccCMUserUsersDataSourceConfigManagedUser(username string, unixID, groupID int64, homeDirectory, shell, usernamePattern, groupIDFilter, userID string) string {
	password := "TestPass123!"
	filterConfig := ""
	if usernamePattern != "" {
		filterConfig += fmt.Sprintf("  username_pattern = %q\n", usernamePattern)
	}
	if groupIDFilter != "" {
		filterConfig += fmt.Sprintf("  group_id         = %q\n", groupIDFilter)
	}
	if userID != "" {
		filterConfig += fmt.Sprintf("  user_id          = %q\n", userID)
	}

	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "fixture" {
  username       = %[4]q
  password       = %[5]q
  uid            = %[6]d
  gid            = %[7]d
  home_directory = %[8]q
  shell          = %[9]q
}

data "bcm_cmuser_users" "test" {
  depends_on = [bcm_cmuser_user.fixture]
%[10]s}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		unixID,
		groupID,
		homeDirectory,
		shell,
		filterConfig,
	)
}
