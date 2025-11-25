// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
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

func TestAccCMNetNetwork_Basic(t *testing.T) {
	networkName := generateUniqueTestName("tftest-network")

	// Initialize ID tracker for consistency verification across operations
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMNetNetworkDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMNetNetworkConfigBasic(networkName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(networkName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("domain_name"),
						knownvalue.StringExact("cluster.local"),
					),
					compareID.AddStateValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMNetNetworkConfigBasic(networkName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// ImportState testing
			{
				Config:            testAccCMNetNetworkConfigBasic(networkName),
				ResourceName:      "bcm_cmnet_network.test",
				ImportState:       true,
				ImportStateVerify: true,
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

func TestAccCMNetNetwork_Complete(t *testing.T) {
	networkName := generateUniqueTestName("tftest-network")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMNetNetworkDestroy,
		Steps: []resource.TestStep{
			// Create with all attributes
			{
				Config: testAccCMNetNetworkConfigComplete(networkName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(networkName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("subnet"),
						knownvalue.StringExact("192.168.100.0/24"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("gateway"),
						knownvalue.StringExact("192.168.100.1"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("mtu"),
						knownvalue.Int64Exact(9000),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("domain_name"),
						knownvalue.StringExact("test.local"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("dhcp_range_start"),
						knownvalue.StringExact("192.168.100.100"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("dhcp_range_end"),
						knownvalue.StringExact("192.168.100.200"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("dhcp_enabled"),
						knownvalue.Bool(true),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Test network"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMNetNetworkConfigComplete(networkName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func TestAccCMNetNetwork_Update(t *testing.T) {
	networkName := generateUniqueTestName("tftest-network")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMNetNetworkDestroy,
		Steps: []resource.TestStep{
			// Create with initial config
			{
				Config: testAccCMNetNetworkConfigBasic(networkName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(networkName),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMNetNetworkConfigBasic(networkName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update MTU and notes
			{
				Config: testAccCMNetNetworkConfigUpdate(networkName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(networkName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("mtu"),
						knownvalue.Int64Exact(9000),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Updated notes"),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMNetNetworkConfigUpdate(networkName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMNetNetwork_DriftDetection verifies Terraform detects external modifications.
func TestAccCMNetNetwork_DriftDetection(t *testing.T) {
	networkName := generateUniqueTestName("tftest-network-drift")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMNetNetworkDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create network with initial notes
			{
				Config: testAccCMNetNetworkConfigWithNotes(networkName, "Initial notes"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Initial notes"),
					),
				},
			},
			// Step 2: Modify network externally via BCM API (drift simulation)
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get network UUID by name
					uuid := getResourceUUIDByName(t, "cmnet", "getNetwork", networkName)

					// Fetch full network data
					body, err := client.CallJSONRPC(ctx, "cmnet", "getNetwork", uuid)
					if err != nil {
						t.Fatalf("Failed to fetch network: %v", err)
					}

					var networkData map[string]interface{}
					if err := json.Unmarshal(body, &networkData); err != nil {
						t.Fatalf("Failed to unmarshal network data: %v", err)
					}

					// Modify notes field externally (this is the drift we're introducing)
					networkData["notes"] = "Modified externally"

					// Build BCM entity structure (required for updateNetwork)
					entity := map[string]interface{}{
						"baseType":      "Network",
						"childType":     "",
						"modified":      true,
						"to_be_removed": false,
						"revision":      networkData["revision"],
						"uuid":          uuid,
					}

					// Copy all network fields
					for k, v := range networkData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmnet", "updateNetwork", entity, false)
					if err != nil {
						t.Fatalf("Failed to update network externally: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified notes externally to: %v", entity["notes"])
				},
				Config: testAccCMNetNetworkConfigWithNotes(networkName, "Initial notes"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Drift should be detected!
					},
				},
			},
			// Step 3: Terraform restores desired state
			{
				Config: testAccCMNetNetworkConfigWithNotes(networkName, "Initial notes"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmnet_network.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Initial notes"),
					),
				},
			},
		},
	})
}

func testAccCheckCMNetNetworkDestroy(s *terraform.State) error {
	// Create BCM client using shared helper
	client := createTestBCMClient(&testing.T{})
	ctx := context.Background()

	resourcesChecked := 0
	var errors []string

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmnet_network" {
			continue
		}

		resourcesChecked++
		name := rs.Primary.Attributes["name"]
		uuid := rs.Primary.Attributes["uuid"]

		// Verify network deleted with exponential backoff (4 retries)
		deleted := verifyResourceDeleted(
			ctx,
			client,
			"cmnet",
			"getNetwork",
			name,
			4, // retry count with exponential backoff
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"Network still exists after destroy. Name: %s, UUID: %s, Retries: 4",
				name, uuid))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("CheckDestroy failures:\n  - %s\n  - Verified: %d networks",
			strings.Join(errors, "\n  - "),
			resourcesChecked)
	}

	// Log number of resources checked for debugging
	if resourcesChecked > 0 {
		fmt.Printf("[DEBUG] CheckDestroy verified %d network resources were deleted\n", resourcesChecked)
	}

	return nil
}

func testAccCMNetNetworkConfigBasic(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name        = %[4]q
  domain_name = "cluster.local"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

func testAccCMNetNetworkConfigComplete(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name             = %[4]q
  subnet           = "192.168.100.0/24"
  gateway          = "192.168.100.1"
  mtu              = 9000
  domain_name      = "test.local"
  dhcp_range_start = "192.168.100.100"
  dhcp_range_end   = "192.168.100.200"
  notes            = "Test network"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

func testAccCMNetNetworkConfigUpdate(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name        = %[4]q
  domain_name = "cluster.local"
  mtu         = 9000
  notes       = "Updated notes"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

func testAccCMNetNetworkConfigWithNotes(name, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name        = %[4]q
  domain_name = "cluster.local"
  notes       = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		notes,
	)
}
