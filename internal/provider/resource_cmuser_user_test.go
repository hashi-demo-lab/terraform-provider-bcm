// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// testAccCMUserUserConfig generates provider configuration for tests.
func testAccCMUserUserConfig(username, password string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "test" {
  username = %[4]q
  password = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
	)
}

// testAccCMUserUserConfigComplete generates configuration with all attributes.
func testAccCMUserUserConfigComplete(username, password, shell, fullName, email, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "test" {
  username       = %[4]q
  password       = %[5]q
  shell          = %[6]q
  full_name      = %[7]q
  email          = %[8]q
  notes          = %[9]q
  shadow_max     = 90
  shadow_warning = 14
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		shell,
		fullName,
		email,
		notes,
	)
}

// testAccCMUserUserConfigWithShell generates configuration with specific shell.
func testAccCMUserUserConfigWithShell(username, password, shell string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "test" {
  username = %[4]q
  password = %[5]q
  shell    = %[6]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		shell,
	)
}

// testAccCheckCMUserUserDestroy verifies user is deleted.
func testAccCheckCMUserUserDestroy(s *terraform.State) error {
	client := createTestBCMClient(&testing.T{})
	ctx := context.Background()

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmuser_user" {
			continue
		}

		username := rs.Primary.Attributes["username"]

		// Wait for eventual consistency
		time.Sleep(2 * time.Second)

		// Check if user still exists using verifyResourceDeleted with retries
		deleted := verifyResourceDeleted(ctx, client, "cmuser", "getUser", username, 4)
		if !deleted {
			return fmt.Errorf("user %s still exists after destroy", username)
		}
	}

	return nil
}

// generateUniqueUnixUsername creates a unique Unix-compliant username.
// Unix usernames: 1-32 chars, must start with letter, alphanumeric + underscore only.
func generateUniqueUnixUsername() string {
	now := time.Now()
	// Use compact format: tfu + timestamp suffix (max 32 chars total)
	// Format: tfu_YYMMDDHHmmss (14 chars for timestamp)
	timestamp := now.Format("060102150405")
	return fmt.Sprintf("tfu_%s", timestamp)
}

// TestAccCMUserUser_Basic tests basic user creation and deletion.
func TestAccCMUserUser_Basic(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "TestPass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create and Read
			{
				Config: testAccCMUserUserConfig(username, password),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/bash"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMUserUser_Complete tests user with all attributes.
func TestAccCMUserUser_Complete(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "CompletePass123!"
	shell := "/bin/zsh"
	fullName := "Test User Complete"
	email := "testuser@example.com"
	notes := "Complete test user"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create with all attributes
			{
				Config: testAccCMUserUserConfigComplete(username, password, shell, fullName, email, notes),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact(shell),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("full_name"),
						knownvalue.StringExact(fullName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("email"),
						knownvalue.StringExact(email),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact(notes),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shadow_max"),
						knownvalue.Int64Exact(90),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shadow_warning"),
						knownvalue.Int64Exact(14),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMUserUser_Update tests user attribute updates.
func TestAccCMUserUser_Update(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "UpdatePass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create with initial shell
			{
				Config: testAccCMUserUserConfigWithShell(username, password, "/bin/bash"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/bash"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update shell
			{
				Config: testAccCMUserUserConfigWithShell(username, password, "/bin/sh"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/sh"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMUserUser_Idempotent tests that reapplying the same config produces no changes.
func TestAccCMUserUser_Idempotent(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "IdempotentPass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccCMUserUserConfig(username, password),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Verify idempotent - no plan changes
			{
				Config: testAccCMUserUserConfig(username, password),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMUserUser_Import tests importing an existing user.
func TestAccCMUserUser_Import(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "ImportPass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create user first
			{
				Config: testAccCMUserUserConfig(username, password),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Import by username
			{
				ResourceName:            "bcm_cmuser_user.test",
				ImportState:             true,
				ImportStateId:           username,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"password", "force"}, // Password cannot be recovered
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMUserUser_DriftShell tests drift detection for shell attribute.
func TestAccCMUserUser_DriftShell(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "DriftPass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create with initial shell
			{
				Config: testAccCMUserUserConfigWithShell(username, password, "/bin/bash"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/bash"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Modify shell externally via BCM API (drift)
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get user data
					body, err := client.CallJSONRPC(ctx, "cmuser", "getUser", username)
					if err != nil {
						t.Fatalf("Failed to get user for drift test: %v", err)
					}

					var userData map[string]interface{}
					if err := json.Unmarshal(body, &userData); err != nil {
						t.Fatalf("Failed to parse user data: %v", err)
					}

					// Modify shell externally (snake_case -> camelCase: shell -> loginShell)
					userData["loginShell"] = "/bin/sh"
					userData["modified"] = true

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmuser", "updateUser", userData, false)
					if err != nil {
						t.Fatalf("Failed to modify user externally: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified shell externally to /bin/sh")
				},
				Config: testAccCMUserUserConfigWithShell(username, password, "/bin/bash"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Drift detected!
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Terraform restores desired state
			{
				Config: testAccCMUserUserConfigWithShell(username, password, "/bin/bash"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/bash"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMUserUser_DriftNotes tests drift detection for notes attribute.
func TestAccCMUserUser_DriftNotes(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "DriftNotesPass123!"
	initialNotes := "Initial notes"
	driftNotes := "Externally modified notes"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create with initial notes
			{
				Config: testAccCMUserUserConfigComplete(username, password, "/bin/bash", "Test User", "test@example.com", initialNotes),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact(initialNotes),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Modify notes externally via BCM API (drift)
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get user data
					body, err := client.CallJSONRPC(ctx, "cmuser", "getUser", username)
					if err != nil {
						t.Fatalf("Failed to get user for drift test: %v", err)
					}

					var userData map[string]interface{}
					if err := json.Unmarshal(body, &userData); err != nil {
						t.Fatalf("Failed to parse user data: %v", err)
					}

					// Modify notes externally
					userData["notes"] = driftNotes
					userData["modified"] = true

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmuser", "updateUser", userData, false)
					if err != nil {
						t.Fatalf("Failed to modify user externally: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified notes externally to: %s", driftNotes)
				},
				Config: testAccCMUserUserConfigComplete(username, password, "/bin/bash", "Test User", "test@example.com", initialNotes),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Drift detected!
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Terraform restores desired state
			{
				Config: testAccCMUserUserConfigComplete(username, password, "/bin/bash", "Test User", "test@example.com", initialNotes),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact(initialNotes),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMUserUser_PasswordSensitive tests that password is not logged.
func TestAccCMUserUser_PasswordSensitive(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "SensitivePass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCMUserUserConfig(username, password),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					// Password should be set but marked sensitive
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("password"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}
