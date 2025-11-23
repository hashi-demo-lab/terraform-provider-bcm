// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccCMPartPartition_BasicMinimal tests Import and Update of the existing base partition
// NOTE: BCM requires exactly one partition named "base" - we import and update it rather than create
func TestAccCMPartPartition_BasicMinimal(t *testing.T) {
	// Get existing partition UUID for import
	_ = createTestBCMClient(t)
	uuid := getResourceUUIDByName(t, "cmpart", "getPartitions", "base")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		// Skip CheckDestroy - base partition cannot be deleted
		Steps: []resource.TestStep{
			// Step 1: Import existing partition
			{
				Config:            testAccPartitionConfigMinimal("Test Import"),
				ResourceName:      "bcm_cmpart_partition.test",
				ImportState:       true,
				ImportStateId:     uuid,
				ImportStateVerify: true,
			},
			// Step 2: Update cluster_name
			{
				Config: testAccPartitionConfigMinimal("Updated Cluster"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact("base"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("Updated Cluster"),
					),
				},
			},
			// Step 3: Idempotency check
			{
				Config: testAccPartitionConfigMinimal("Updated Cluster"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMPartPartition_UpdateMinimal tests updating partition fields
func TestAccCMPartPartition_UpdateMinimal(t *testing.T) {
	// Get existing partition UUID for import
	client := createTestBCMClient(t)
	ctx := context.Background()
	uuid := getResourceUUIDByName(t, "cmpart", "getPartitions", "base")

	// Store original cluster name to restore after test
	body, _ := client.CallJSONRPC(ctx, "cmpart", "getPartitions")
	var partitions []map[string]interface{}
	_ = json.Unmarshal(body, &partitions)
	originalClusterName := partitions[0]["clusterName"].(string)

	// Cleanup: restore original cluster name after test
	defer func() {
		body, _ := client.CallJSONRPC(ctx, "cmpart", "getPartitions")
		var partitions []map[string]interface{}
		_ = json.Unmarshal(body, &partitions)

		partitionData := partitions[0]
		partitionData["clusterName"] = originalClusterName

		entity := map[string]interface{}{
			"baseType":      "Partition",
			"childType":     "",
			"modified":      true,
			"to_be_removed": false,
			"revision":      partitionData["revision"],
			"uuid":          uuid,
		}
		for k, v := range partitionData {
			if k != "uuid" {
				entity[k] = v
			}
		}
		client.CallJSONRPC(ctx, "cmpart", "updatePartition", entity, false)
	}()

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			// Step 1: Import and update to Cluster A
			{
				Config: testAccPartitionConfigMinimal("Cluster A"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("Cluster A"),
					),
				},
			},
			// Step 2: Update to Cluster B
			{
				Config: testAccPartitionConfigMinimal("Cluster B"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmpart_partition.test",
						tfjsonpath.New("cluster_name"),
						knownvalue.StringExact("Cluster B"),
					),
				},
			},
		},
	})
}

// testAccPartitionConfigMinimal generates minimal partition configuration
// This works with the existing "base" partition
func testAccPartitionConfigMinimal(clusterName string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "test" {
  name         = "base"
  cluster_name = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		clusterName,
	)
}
