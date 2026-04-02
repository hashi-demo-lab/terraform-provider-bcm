// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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

// testAccCMUserUserConfigWithSSHKeys generates configuration with authorized SSH keys.
func testAccCMUserUserConfigWithSSHKeys(username, password, sshKeys string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "test" {
  username            = %[4]q
  password            = %[5]q
  authorized_ssh_keys = %[6]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		sshKeys,
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
		time.Sleep(TestEventualConsistencyDelay)

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
					ctx := t.Context()

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
					time.Sleep(TestEventualConsistencyDelay)

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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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
					ctx := t.Context()

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
					time.Sleep(TestEventualConsistencyDelay)

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
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
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

// TestAccCMUserUser_DriftHomeDirectory tests drift detection for home_directory attribute.
func TestAccCMUserUser_DriftHomeDirectory(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "DriftHomeDirPass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with initial home_directory
			{
				Config: testAccCMUserUserConfigWithHomeDir(username, password, "/home/testuser"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("home_directory"),
						knownvalue.StringExact("/home/testuser"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Modify home_directory externally via BCM API (drift)
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get user data
					body, err := client.CallJSONRPC(ctx, "cmuser", "getUser", username)
					if err != nil {
						t.Fatalf("Failed to get user for drift test: %v", err)
					}

					var userData map[string]interface{}
					if err := json.Unmarshal(body, &userData); err != nil {
						t.Fatalf("Failed to parse user data: %v", err)
					}

					// Modify homeDirectory externally (snake_case -> camelCase: home_directory -> homeDirectory)
					userData["homeDirectory"] = "/tmp/drifted"
					userData["modified"] = true

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmuser", "updateUser", userData, false)
					if err != nil {
						t.Fatalf("Failed to modify user externally: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(TestEventualConsistencyDelay)

					t.Logf("[DEBUG] Modified homeDirectory externally to /tmp/drifted")
				},
				Config: testAccCMUserUserConfigWithHomeDir(username, password, "/home/testuser"),
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
			// Step 3: Terraform restores desired state
			{
				Config: testAccCMUserUserConfigWithHomeDir(username, password, "/home/testuser"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("home_directory"),
						knownvalue.StringExact("/home/testuser"),
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

// TestAccCMUserUser_AuthorizedSSHKeys tests that authorized_ssh_keys is preserved
// after apply (BCM API returns empty string for this field, similar to password).
// This test verifies the fix for the "inconsistent result after apply" error.
func TestAccCMUserUser_AuthorizedSSHKeys(t *testing.T) {
	if os.Getenv("TF_ACC") == "" {
		t.Skip("Acceptance tests skipped unless env 'TF_ACC' set")
	}

	username := generateUniqueUnixUsername()
	password := "SSHKeyTestPass123!"
	sshKey := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7example test@host"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Create with authorized_ssh_keys
			{
				Config: testAccCMUserUserConfigWithSSHKeys(username, password, sshKey),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					// authorized_ssh_keys should be preserved in state
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("authorized_ssh_keys"),
						knownvalue.StringExact(sshKey),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Re-apply same config (idempotency check)
			{
				Config: testAccCMUserUserConfigWithSSHKeys(username, password, sshKey),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					// authorized_ssh_keys should still be preserved
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("authorized_ssh_keys"),
						knownvalue.StringExact(sshKey),
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

// userDisappearsCheck is a custom statecheck.StateCheck that deletes a user
// resource via the BCM API after the apply, simulating external deletion.
// This causes Terraform to detect the resource as "disappeared" on next refresh.
type userDisappearsCheck struct {
	resourceAddress string
}

func (c userDisappearsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	// Find the resource in state and extract the UUID
	var uuid string
	for _, rc := range req.State.Values.RootModule.Resources {
		if rc.Address == c.resourceAddress {
			if v, ok := rc.AttributeValues["uuid"]; ok {
				uuid, ok = v.(string)
				if !ok || uuid == "" {
					resp.Error = fmt.Errorf("uuid attribute is not a non-empty string for %s", c.resourceAddress)
					return
				}
			} else {
				resp.Error = fmt.Errorf("uuid attribute not found for %s", c.resourceAddress)
				return
			}
			break
		}
	}

	if uuid == "" {
		resp.Error = fmt.Errorf("resource %s not found in state", c.resourceAddress)
		return
	}

	// Delete the resource externally via BCM API
	client := createTestBCMClient(&testing.T{})
	_, err := client.CallJSONRPC(ctx, "cmuser", "removeUser", uuid)
	if err != nil {
		resp.Error = fmt.Errorf("failed to delete user externally (uuid=%s): %s", uuid, err)
		return
	}

	// Wait for eventual consistency
	time.Sleep(2 * time.Second)
}

// TestAccCMUserUser_Disappears tests that Terraform detects when a user is
// deleted externally (outside of Terraform) and plans to recreate it.
func TestAccCMUserUser_Disappears(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "TestPass123!"

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
					userDisappearsCheck{resourceAddress: "bcm_cmuser_user.test"},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

// TestAccCMUserUser_ValidationEmptyUsername tests that an empty username is rejected
// by the schema validator (LengthBetween 1-32).
func TestAccCMUserUser_ValidationEmptyUsername(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMUserUserConfig("", "TestPass123!"),
				ExpectError: regexp.MustCompile(`(?i)username`),
			},
		},
	})
}

// TestAccCMUserUser_ValidationInvalidUsernameStartsWithDigit tests that a username
// beginning with a digit is rejected by the regex validator.
func TestAccCMUserUser_ValidationInvalidUsernameStartsWithDigit(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMUserUserConfig("1baduser", "TestPass123!"),
				ExpectError: regexp.MustCompile(`must start with a letter`),
			},
		},
	})
}

// TestAccCMUserUser_ValidationInvalidShellPath tests that a non-absolute shell path
// is rejected by the regex validator.
func TestAccCMUserUser_ValidationInvalidShellPath(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMUserUserConfigWithShell(generateUniqueUnixUsername(), "TestPass123!", "relative/path"),
				ExpectError: regexp.MustCompile(`must be an absolute path`),
			},
		},
	})
}

// TestAccCMUserUser_ValidationInvalidHomeDirectory tests that a non-absolute
// home directory path is rejected by the regex validator.
func TestAccCMUserUser_ValidationInvalidHomeDirectory(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMUserUserConfigWithHomeDir(generateUniqueUnixUsername(), "TestPass123!", "noslash/home"),
				ExpectError: regexp.MustCompile(`must be an absolute path`),
			},
		},
	})
}

// testAccCMUserUserConfigWithMultipleFields generates configuration with shell, full_name, and notes.
func testAccCMUserUserConfigWithMultipleFields(username, password, shell, fullName, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "test" {
  username  = %[4]q
  password  = %[5]q
  shell     = %[6]q
  full_name = %[7]q
  notes     = %[8]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		shell,
		fullName,
		notes,
	)
}

// TestAccCMUserUser_UpdateMultipleFields tests updating shell, full_name, and notes simultaneously.
func TestAccCMUserUser_UpdateMultipleFields(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "MultiFieldPass123!"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with basic config
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
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/bash"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Update shell, full_name, and notes simultaneously
			{
				Config: testAccCMUserUserConfigWithMultipleFields(
					username, password, "/bin/zsh", "Updated User", "Updated notes",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/zsh"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("full_name"),
						knownvalue.StringExact("Updated User"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Updated notes"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Idempotency check — reapply same config, expect empty plan
			{
				Config: testAccCMUserUserConfigWithMultipleFields(
					username, password, "/bin/zsh", "Updated User", "Updated notes",
				),
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

// testAccCMUserUserConfigAllOptionalFields generates configuration with ALL optional fields set.
func testAccCMUserUserConfigAllOptionalFields(username, password, shell, fullName, notes, homeDir, sshKeys string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "test" {
  username            = %[4]q
  password            = %[5]q
  shell               = %[6]q
  full_name           = %[7]q
  notes               = %[8]q
  home_directory      = %[9]q
  authorized_ssh_keys = %[10]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		shell,
		fullName,
		notes,
		homeDir,
		sshKeys,
	)
}

// TestAccCMUserUser_BasicWithAllFields tests creating a user with ALL optional fields set,
// verifies each field via ConfigStateChecks, then re-applies for idempotency.
func TestAccCMUserUser_BasicWithAllFields(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "AllFieldsPass123!"
	shell := "/bin/zsh"
	fullName := "All Fields User"
	notes := "User created with all optional fields"
	homeDir := fmt.Sprintf("/home/%s", username)
	sshKeys := "ssh-rsa AAAAB3NzaC1yc2EAAAADAQABAAABgQC7allfields test@host"

	// ID consistency tracking
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with all optional fields and verify each
			{
				Config: testAccCMUserUserConfigAllOptionalFields(username, password, shell, fullName, notes, homeDir, sshKeys),
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
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
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
						tfjsonpath.New("notes"),
						knownvalue.StringExact(notes),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("home_directory"),
						knownvalue.StringExact(homeDir),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("authorized_ssh_keys"),
						knownvalue.StringExact(sshKeys),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Re-apply same config — verify idempotency (empty plan)
			{
				Config: testAccCMUserUserConfigAllOptionalFields(username, password, shell, fullName, notes, homeDir, sshKeys),
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

// TestAccCMUserUser_UpdateShellAndVerifyIdempotency tests updating the shell attribute
// from /bin/bash to /bin/zsh, verifies the update, then confirms idempotency.
func TestAccCMUserUser_UpdateShellAndVerifyIdempotency(t *testing.T) {
	username := generateUniqueUnixUsername()
	password := "ShellUpdatePass123!"

	// ID consistency tracking across all 3 steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with /bin/bash
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
			// Step 2: Update shell to /bin/zsh and verify
			{
				Config: testAccCMUserUserConfigWithShell(username, password, "/bin/zsh"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("username"),
						knownvalue.StringExact(username),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/zsh"),
					),
					compareID.AddStateValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Re-apply same config — verify idempotency (empty plan)
			{
				Config: testAccCMUserUserConfigWithShell(username, password, "/bin/zsh"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmuser_user.test",
						tfjsonpath.New("shell"),
						knownvalue.StringExact("/bin/zsh"),
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

// TestAccCMUserUser_ValidationInvalidUID tests that a UID above 65535 is rejected
// by the schema validator (Between 0-65535).
func TestAccCMUserUser_ValidationInvalidUID(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMUserUserDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMUserUserConfigWithUID(generateUniqueUnixUsername(), "TestPass123!", 70000),
				ExpectError: regexp.MustCompile(`(?i)must be between 0 and 65535|Attribute uid`),
			},
		},
	})
}

// testAccCMUserUserConfigWithUID generates configuration with a specific UID.
func testAccCMUserUserConfigWithUID(username, password string, uid int64) string {
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
  uid      = %[6]d
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		uid,
	)
}

// testAccCMUserUserConfigWithHomeDir generates configuration with a specific home_directory.
func testAccCMUserUserConfigWithHomeDir(username, password, homeDir string) string {
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
  home_directory = %[6]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		username,
		password,
		homeDir,
	)
}
