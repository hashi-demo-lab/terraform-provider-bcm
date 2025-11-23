// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
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

// TestAccCMPartPartition_Basic tests Create, Read, Import, and basic CRUD workflow
func TestAccCMPartPartition_Basic(t *testing.T) {
	partitionName := generateUniqueTestName("test-partition")
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create and Read
			{
				Config: testAccPartitionConfigBasic(partitionName, "HPC Cluster"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "name", partitionName),
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "cluster_name", "HPC Cluster"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_partition.test", "uuid"),
					resource.TestCheckResourceAttrSet("bcm_cmpart_partition.test", "id"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(partitionName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("HPC Cluster"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Idempotency check - verify no changes after creation
			{
				Config: testAccPartitionConfigBasic(partitionName, "HPC Cluster"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Step 3: Import by UUID
			{
				Config:            testAccPartitionConfigBasic(partitionName, "HPC Cluster"),
				ResourceName:      "bcm_cmpart_partition.test",
				ImportState:       true,
				ImportStateVerify: true,
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartition_Update tests in-place updates of partition configuration
func TestAccCMPartPartition_Update(t *testing.T) {
	partitionName := generateUniqueTestName("test-partition")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy,
		Steps: []resource.TestStep{
			// Create with initial values
			{
				Config: testAccPartitionConfigComplete(partitionName, "Cluster A", "node", 3, "Initial notes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "cluster_name", "Cluster A"),
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "slave_name", "node"),
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "notes", "Initial notes"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("Cluster A"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_name"),
						knownvalue.StringExact("node"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_digits"),
						knownvalue.Int64Exact(3),
					),
				},
			},
			// Update cluster_name
			{
				Config: testAccPartitionConfigComplete(partitionName, "Cluster B", "node", 3, "Initial notes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "cluster_name", "Cluster B"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("Cluster B"),
					),
				},
			},
			// Idempotency after cluster_name update
			{
				Config: testAccPartitionConfigComplete(partitionName, "Cluster B", "node", 3, "Initial notes"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update slave_name and slave_digits
			{
				Config: testAccPartitionConfigComplete(partitionName, "Cluster B", "compute", 4, "Initial notes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "slave_name", "compute"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_name"),
						knownvalue.StringExact("compute"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_digits"),
						knownvalue.Int64Exact(4),
					),
				},
			},
			// Idempotency after slave_name update
			{
				Config: testAccPartitionConfigComplete(partitionName, "Cluster B", "compute", 4, "Initial notes"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update notes field
			{
				Config: testAccPartitionConfigComplete(partitionName, "Cluster B", "compute", 4, "Updated notes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "notes", "Updated notes"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Updated notes"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartition_NetworkSettings tests list attribute configuration
func TestAccCMPartPartition_NetworkSettings(t *testing.T) {
	partitionName := generateUniqueTestName("test-partition")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy,
		Steps: []resource.TestStep{
			// Create with network configuration lists
			{
				Config: testAccPartitionConfigNetworkSettings(
					partitionName,
					[]string{"admin@example.com", "ops@example.com"},
					[]string{"ntp1.example.com", "ntp2.example.com"},
					[]string{"8.8.8.8", "8.8.4.4"},
					[]string{"example.com"},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("admin_email"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("time_servers"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("name_servers"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("search_domains"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
			// Update network lists (add/remove entries)
			{
				Config: testAccPartitionConfigNetworkSettings(
					partitionName,
					[]string{"admin@example.com"},
					[]string{"ntp1.example.com"},
					[]string{"8.8.8.8"},
					[]string{"example.com", "corp.example.com"},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("admin_email"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("search_domains"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartition_DriftDetection verifies Terraform detects external modifications
func TestAccCMPartPartition_DriftDetection(t *testing.T) {
	partitionName := generateUniqueTestName("test-partition")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create partition with initial notes
			{
				Config: testAccPartitionConfigWithNotes(partitionName, "Initial notes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "notes", "Initial notes"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Initial notes"),
					),
				},
			},
			// Step 2: Modify partition externally via BCM API (drift simulation)
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := context.Background()

					// Get partition UUID by name
					uuid := getResourceUUIDByName(t, "cmpart", "getPartition", partitionName)

					// Fetch full partition data
					body, err := client.CallJSONRPC(ctx, "cmpart", "getPartition", uuid)
					if err != nil {
						t.Fatalf("Failed to fetch partition: %v", err)
					}

					var partitionData map[string]interface{}
					if err := json.Unmarshal(body, &partitionData); err != nil {
						t.Fatalf("Failed to unmarshal partition data: %v", err)
					}

					// Modify notes field (this is the drift we're introducing)
					partitionData["notes"] = "Modified externally"

					// Build BCM entity structure (required for updatePartition)
					entity := map[string]interface{}{
						"baseType":      "Partition",
						"childType":     "",
						"modified":      true,
						"to_be_removed": false,
						"revision":      partitionData["revision"],
						"uuid":          uuid,
					}

					// Copy all partition fields
					for k, v := range partitionData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmpart", "updatePartition", entity, false)
					if err != nil {
						t.Fatalf("Failed to update partition externally: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(2 * time.Second)

					t.Logf("[DEBUG] Modified notes externally to: %v", entity["notes"])
				},
				Config: testAccPartitionConfigWithNotes(partitionName, "Initial notes"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Drift should be detected!
					},
				},
			},
			// Step 3: Terraform restores desired state
			{
				Config: testAccPartitionConfigWithNotes(partitionName, "Initial notes"),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmpart_partition.test", "notes", "Initial notes"),
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("notes"),
						knownvalue.StringExact("Initial notes"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartition_SlaveNaming tests node naming configuration
func TestAccCMPartPartition_SlaveNaming(t *testing.T) {
	partitionName := generateUniqueTestName("test-partition")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy,
		Steps: []resource.TestStep{
			// Create with custom slave naming
			{
				Config: testAccPartitionConfigComplete(partitionName, "Test Cluster", "compute", 4, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_name"),
						knownvalue.StringExact("compute"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_digits"),
						knownvalue.Int64Exact(4),
					),
				},
			},
			// Update slave_digits
			{
				Config: testAccPartitionConfigComplete(partitionName, "Test Cluster", "compute", 3, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_digits"),
						knownvalue.Int64Exact(3),
					),
				},
			},
			// Update slave_name
			{
				Config: testAccPartitionConfigComplete(partitionName, "Test Cluster", "gpu", 3, ""),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("slave_name"),
						knownvalue.StringExact("gpu"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartition_IDConsistency verifies ID remains stable across operations
func TestAccCMPartPartition_IDConsistency(t *testing.T) {
	partitionName := generateUniqueTestName("test-partition")
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckPartitionDestroy,
		Steps: []resource.TestStep{
			// Create - capture ID
			{
				Config: testAccPartitionConfigBasic(partitionName, "Test Cluster"),
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Import - verify ID unchanged
			{
				Config:            testAccPartitionConfigBasic(partitionName, "Test Cluster"),
				ResourceName:      "bcm_cmpart_partition.test",
				ImportState:       true,
				ImportStateVerify: true,
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update - verify ID still unchanged
			{
				Config: testAccPartitionConfigBasic(partitionName, "Updated Cluster"),
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartition_ValidationErrors tests schema validation
func TestAccCMPartPartition_ValidationErrors(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Test empty partition name
			{
				Config:      testAccPartitionConfigBasic("", "Test Cluster"),
				ExpectError: regexp.MustCompile(`Attribute name string length must be between 1 and`),
			},
		},
	})
}

// testAccCheckPartitionDestroy verifies partition deletion with exponential backoff
func testAccCheckPartitionDestroy(s *terraform.State) error {
	client := createTestBCMClient(&testing.T{})
	ctx := context.Background()

	var errors []string
	resourceCount := 0

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmpart_partition" {
			continue
		}

		resourceCount++
		uuid := rs.Primary.ID

		// Verify partition deleted with exponential backoff
		deleted := verifyResourceDeleted(
			ctx,
			client,
			"cmpart",
			"getPartition",
			uuid,
			4, // retry count (15s total: 1+2+4+8)
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"Partition still exists after destroy. UUID: %s, Retries: 4",
				uuid,
			))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("CheckDestroy failures:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// testAccPartitionConfigBasic generates basic partition configuration
func testAccPartitionConfigBasic(name, clusterName string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "test" {
  name         = %[4]q
  cluster_name = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		clusterName,
	)
}

// testAccPartitionConfigWithNotes generates config with notes field
func testAccPartitionConfigWithNotes(name, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "test" {
  name         = %[4]q
  cluster_name = "Test Cluster"
  notes        = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		notes,
	)
}

// testAccPartitionConfigNetworkSettings generates config with network list attributes
func testAccPartitionConfigNetworkSettings(name string, adminEmails, timeServers, nameServers, searchDomains []string) string {
	adminEmailsHCL := formatStringListForHCL(adminEmails)
	timeServersHCL := formatStringListForHCL(timeServers)
	nameServersHCL := formatStringListForHCL(nameServers)
	searchDomainsHCL := formatStringListForHCL(searchDomains)

	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "test" {
  name         = %[4]q
  cluster_name = "Test Cluster"

  admin_email    = %[5]s
  time_servers   = %[6]s
  name_servers   = %[7]s
  search_domains = %[8]s
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		adminEmailsHCL,
		timeServersHCL,
		nameServersHCL,
		searchDomainsHCL,
	)
}

// testAccPartitionConfigComplete generates config with all optional fields
func testAccPartitionConfigComplete(name, clusterName, slaveName string, slaveDigits int64, notes string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "test" {
  name         = %[4]q
  cluster_name = %[5]q
  slave_name   = %[6]q
  slave_digits = %[7]d
  notes        = %[8]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		clusterName,
		slaveName,
		slaveDigits,
		notes,
	)
}

// formatStringListForHCL formats a string slice for HCL list syntax
func formatStringListForHCL(strs []string) string {
	if len(strs) == 0 {
		return "[]"
	}

	quoted := make([]string, len(strs))
	for i, s := range strs {
		quoted[i] = fmt.Sprintf("%q", s)
	}

	return fmt.Sprintf("[%s]", strings.Join(quoted, ", "))
}
