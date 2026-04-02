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

// testAccPreCheckCMKubeCluster verifies environment variables for cluster tests.
func testAccPreCheckCMKubeCluster(t *testing.T) {
	if v := os.Getenv("BCM_ENDPOINT"); v == "" {
		t.Fatal("BCM_ENDPOINT must be set for acceptance tests")
	}
	if v := os.Getenv("BCM_USERNAME"); v == "" {
		t.Fatal("BCM_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("BCM_PASSWORD"); v == "" {
		t.Fatal("BCM_PASSWORD must be set for acceptance tests")
	}
}

// testAccCheckCMKubeClusterDestroy verifies cluster deletion with enhanced error messages.
func testAccCheckCMKubeClusterDestroy(s *terraform.State) error {
	client := createTestBCMClient(&testing.T{})

	var errors []string
	resourceCount := 0

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmkube_cluster" {
			continue
		}

		resourceCount++
		id := rs.Primary.ID

		// Verify cluster deleted with exponential backoff
		deleted := verifyResourceDeleted(
			context.Background(),
			client,
			"cmkube",
			"getKubeCluster",
			id,
			4, // retry count
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"Cluster still exists after destroy. Type: %s, ID: %s, Retries: 4",
				rs.Type,
				id,
			))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("CheckDestroy failures:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// getTestMasterNodeUUID queries BCM for available master node.
func getTestMasterNodeUUID(t *testing.T) string {
	client := createTestBCMClient(t)
	ctx := t.Context()

	// Query available nodes
	body, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
	if err != nil {
		t.Fatalf("Failed to get nodes: %v", err)
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(body, &nodes); err != nil {
		t.Fatalf("Failed to parse nodes: %v", err)
	}

	// Return first available node UUID
	if len(nodes) > 0 {
		if uuid, ok := nodes[0]["uuid"].(string); ok {
			return uuid
		}
	}

	t.Fatal("No available nodes for test cluster")
	return ""
}

// getTestWorkerNodeUUID queries BCM for available worker node.
func getTestWorkerNodeUUID(t *testing.T, index int) string {
	client := createTestBCMClient(t)
	ctx := t.Context()

	body, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
	if err != nil {
		t.Fatalf("Failed to get nodes: %v", err)
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(body, &nodes); err != nil {
		t.Fatalf("Failed to parse nodes: %v", err)
	}

	if len(nodes) <= index+1 { // +1 because first node is master
		t.Skipf("Not enough nodes for worker index %d (need at least %d nodes)", index, index+2)
	}

	if uuid, ok := nodes[index+1]["uuid"].(string); ok {
		return uuid
	}

	t.Fatalf("Node at index %d has invalid UUID format", index+1)
	return ""
}

// getTestEtcdNodeUUID queries BCM for available etcd node (uses a different node than master).
func getTestEtcdNodeUUID(t *testing.T, index int) string {
	client := createTestBCMClient(t)
	ctx := t.Context()

	body, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
	if err != nil {
		t.Fatalf("Failed to get nodes: %v", err)
	}

	var nodes []map[string]interface{}
	if err := json.Unmarshal(body, &nodes); err != nil {
		t.Fatalf("Failed to parse nodes: %v", err)
	}

	// For etcd nodes, we use nodes starting from index 2+ (master is 0, workers start at 1)
	// This allows etcd to be on different nodes than master for HA testing
	etcdIndex := index + 2
	if len(nodes) <= etcdIndex {
		t.Skipf("Not enough nodes for etcd index %d (need at least %d nodes)", index, etcdIndex+1)
	}

	if uuid, ok := nodes[etcdIndex]["uuid"].(string); ok {
		return uuid
	}

	t.Fatalf("Node at index %d has invalid UUID format", etcdIndex)
	return ""
}

// getTestManagementNetworkUUID queries BCM for available management network.
func getTestManagementNetworkUUID(t *testing.T) string {
	client := createTestBCMClient(t)
	ctx := t.Context()

	body, err := client.CallJSONRPC(ctx, "cmnet", "getNetworks")
	if err != nil {
		t.Skipf("Failed to get networks (skipping test): %v", err)
	}

	var networks []map[string]interface{}
	if err := json.Unmarshal(body, &networks); err != nil {
		t.Skipf("Failed to parse networks (skipping test): %v", err)
	}

	// Return first available network UUID
	if len(networks) > 0 {
		if uuid, ok := networks[0]["uuid"].(string); ok {
			return uuid
		}
	}

	t.Skip("No available networks for test cluster (skipping test)")
	return ""
}

// TestAccCMKubeClusterResource_Basic tests complete CRUD lifecycle.
// DEPRECATED: Uses old schema with master_nodes. See TestAccCMKubeCluster_aligned_* for new API.
func TestAccCMKubeClusterResource_Basic(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with master_nodes. Use TestAccCMKubeCluster_aligned_basic instead.")
	clusterName := generateUniqueTestName("tftest-cluster")
	clusterNameUpdated := generateUniqueTestName("tftest-cluster-updated")

	// Get available node UUIDs from test environment
	masterNodeUUID := getTestMasterNodeUUID(t)

	// ID consistency tracking across Create/Import/Update
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with minimal config
			{
				Config: testAccCMKubeClusterResourceConfig(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("master_nodes"),
						knownvalue.ListSizeExact(1),
					),
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMKubeClusterResourceConfig(clusterName, masterNodeUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Import
			{
				ResourceName:      "bcm_cmkube_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
				// BCM cmkube API Limitation: getKubeCluster does NOT return master_nodes/worker_nodes/etcd_nodes
				// These fields are write-only (used during create/update but not returned in read)
				// This is a known BCM API limitation documented in resource_cmkube_cluster.go:296-301
				ImportStateVerifyIgnore: []string{"master_nodes", "worker_nodes", "etcd_nodes"},
			},
			// Update name
			{
				Config: testAccCMKubeClusterResourceConfig(clusterNameUpdated, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterNameUpdated),
					),
					// Verify ID remains unchanged after update
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMKubeClusterResourceConfig(clusterNameUpdated, masterNodeUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_DriftDetection tests external modification detection.
// DEPRECATED: Uses old schema with master_nodes. See TestAccCMKubeCluster_aligned_drift for new API.
func TestAccCMKubeClusterResource_DriftDetection(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with master_nodes. Use TestAccCMKubeCluster_aligned_drift instead.")
	clusterName := generateUniqueTestName("tftest-cluster-drift")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create cluster with version
			{
				Config: testAccCMKubeClusterResourceConfigWithVersion(clusterName, masterNodeUUID, "1.28.0"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("version"),
						knownvalue.StringExact("1.28.0"),
					),
				},
			},
			// Modify cluster externally via BCM API
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get cluster UUID by name
					uuid := getResourceUUIDByName(t, "cmkube", "getKubeCluster", clusterName)

					// Fetch full cluster entity
					body, err := client.CallJSONRPC(ctx, "cmkube", "getKubeCluster", uuid)
					if err != nil {
						t.Fatalf("Failed to fetch cluster for drift test: %v", err)
					}

					var clusterData map[string]interface{}
					if err := json.Unmarshal(body, &clusterData); err != nil {
						t.Fatalf("Failed to parse cluster data: %v", err)
					}

					// Modify version externally (snake_case → camelCase mapping)
					clusterData["version"] = "1.29.0"

					// Build full entity structure
					entity := map[string]interface{}{
						"baseType":      "KubeCluster",
						"childType":     "",
						"modified":      true,
						"to_be_removed": false,
						"revision":      "",
						"uuid":          uuid,
					}
					for k, v := range clusterData {
						if k != "uuid" {
							entity[k] = v
						}
					}

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmkube", "updateKubeCluster", entity, false)
					if err != nil {
						t.Fatalf("Failed to update cluster for drift test: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(TestEventualConsistencyDelay)

					t.Logf("[DEBUG] Modified version externally to: 1.29.0")
				},
				Config: testAccCMKubeClusterResourceConfigWithVersion(clusterName, masterNodeUUID, "1.28.0"),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Drift detected
					},
				},
			},
			// Terraform restores desired state
			{
				Config: testAccCMKubeClusterResourceConfigWithVersion(clusterName, masterNodeUUID, "1.28.0"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("version"),
						knownvalue.StringExact("1.28.0"),
					),
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_WorkerNodes tests worker node scaling.
// DEPRECATED: Uses old schema with worker_nodes. Use device roles instead.
func TestAccCMKubeClusterResource_WorkerNodes(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with worker_nodes. Use bcm_cmdevice_device kubelet_role instead.")
	clusterName := generateUniqueTestName("tftest-cluster-workers")
	masterNodeUUID := getTestMasterNodeUUID(t)
	workerNodeUUID1 := getTestWorkerNodeUUID(t, 0)
	workerNodeUUID2 := getTestWorkerNodeUUID(t, 1)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with 1 worker
			{
				Config: testAccCMKubeClusterResourceConfigWithWorkers(
					clusterName,
					masterNodeUUID,
					[]string{workerNodeUUID1},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("worker_nodes"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
			// Scale to 2 workers
			{
				Config: testAccCMKubeClusterResourceConfigWithWorkers(
					clusterName,
					masterNodeUUID,
					[]string{workerNodeUUID1, workerNodeUUID2},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("worker_nodes"),
						knownvalue.ListSizeExact(2),
					),
				},
			},
			// Scale down to 0 workers
			{
				Config: testAccCMKubeClusterResourceConfigWithWorkers(
					clusterName,
					masterNodeUUID,
					[]string{},
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("worker_nodes"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_ValidationInvalidName tests invalid cluster name.
// DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_validationInvalidName for new API.
func TestAccCMKubeClusterResource_ValidationInvalidName(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with master_nodes. See TestAccCMKubeCluster_aligned_validationInvalidName instead.")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMKubeClusterResourceConfig("invalid name!", masterNodeUUID),
				ExpectError: regexp.MustCompile(`Attribute.*name.*must contain only alphanumeric`),
			},
		},
	})
}

// TestAccCMKubeClusterResource_ValidationInvalidVersion tests invalid version format.
// DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_* for new API.
func TestAccCMKubeClusterResource_ValidationInvalidVersion(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with master_nodes. Use aligned tests instead.")
	clusterName := generateUniqueTestName("tftest-cluster")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMKubeClusterResourceConfigWithVersion(clusterName, masterNodeUUID, "invalid"),
				ExpectError: regexp.MustCompile(`Attribute.*version.*must be valid semver`),
			},
		},
	})
}

// TestAccCMKubeClusterResource_ComputedFields is disabled because BCM API doesn't return creation_time or revision_id fields.
// The fields remain in the schema for potential future BCM API support.
// func TestAccCMKubeClusterResource_ComputedFields(t *testing.T) { ... }

// TestAccCMKubeClusterResource_CompleteConfiguration tests all optional fields.
// DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_* for new API.
func TestAccCMKubeClusterResource_CompleteConfiguration(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with deprecated fields. See aligned tests for new API.")
	clusterName := generateUniqueTestName("tftest-cluster-complete")
	clusterNameUpdated := generateUniqueTestName("tftest-cluster-complete-updated")
	masterNodeUUID := getTestMasterNodeUUID(t)
	workerNodeUUID := getTestWorkerNodeUUID(t, 0)

	// Get management network UUID (if available)
	// This test will be skipped if no management network is available
	managementNetworkUUID := getTestManagementNetworkUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with all optional fields
			{
				Config: testAccCMKubeClusterResourceConfigComplete(
					clusterName,
					masterNodeUUID,
					workerNodeUUID,
					managementNetworkUUID,
					"1.28.0",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("master_nodes"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("worker_nodes"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("management_network"),
						knownvalue.StringExact(managementNetworkUUID),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("version"),
						knownvalue.StringExact("1.28.0"),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMKubeClusterResourceConfigComplete(
					clusterName,
					masterNodeUUID,
					workerNodeUUID,
					managementNetworkUUID,
					"1.28.0",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update all fields
			{
				Config: testAccCMKubeClusterResourceConfigComplete(
					clusterNameUpdated,
					masterNodeUUID,
					workerNodeUUID,
					managementNetworkUUID,
					"1.29.0",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("version"),
						knownvalue.StringExact("1.29.0"),
					),
				},
			},
		},
	})
}

// Test configuration helper functions

func testAccCMKubeClusterResourceConfig(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

func testAccCMKubeClusterResourceConfigWithVersion(name, masterNodeUUID, version string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]
  version      = %[6]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
		version,
	)
}

func testAccCMKubeClusterResourceConfigWithWorkers(name, masterNodeUUID string, workerNodeUUIDs []string) string {
	var workersStr string
	if len(workerNodeUUIDs) > 0 {
		workers := make([]string, len(workerNodeUUIDs))
		for i, uuid := range workerNodeUUIDs {
			workers[i] = fmt.Sprintf("%q", uuid)
		}
		workersStr = fmt.Sprintf("\n  worker_nodes = [%s]", strings.Join(workers, ", "))
	} else {
		// Explicitly set empty list when scaling down to 0 workers
		workersStr = "\n  worker_nodes = []"
	}

	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]%[6]s
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
		workersStr,
	)
}

func testAccCMKubeClusterResourceConfigComplete(name, masterNodeUUID, workerNodeUUID, managementNetworkUUID, version string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name               = %[4]q
  master_nodes       = [%[5]q]
  worker_nodes       = [%[6]q]
  management_network = %[7]q
  version            = %[8]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
		workerNodeUUID,
		managementNetworkUUID,
		version,
	)
}

// TestAccCMKubeClusterResource_P3AdvancedNetworking tests P3 networking features.
// DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_* for new API.
func TestAccCMKubeClusterResource_P3AdvancedNetworking(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with deprecated fields. See aligned tests for new API.")
	clusterName := generateUniqueTestName("tftest-cluster-p3-network")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with advanced networking features
			{
				Config: testAccCMKubeClusterResourceConfigP3Networking(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("cni_plugin"),
						knownvalue.StringExact("calico"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("dns_servers"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("overlay_network"),
						knownvalue.StringExact("overlay-test-uuid"),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMKubeClusterResourceConfigP3Networking(clusterName, masterNodeUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update networking configuration
			{
				Config: testAccCMKubeClusterResourceConfigP3NetworkingUpdated(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("cni_plugin"),
						knownvalue.StringExact("flannel"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("dns_servers"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_P3StorageAndLoadBalancer tests P3 storage and LB features.
// DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_* for new API.
func TestAccCMKubeClusterResource_P3StorageAndLoadBalancer(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with deprecated fields. See aligned tests for new API.")
	clusterName := generateUniqueTestName("tftest-cluster-p3-storage")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with storage classes and load balancer
			{
				Config: testAccCMKubeClusterResourceConfigP3Storage(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("load_balancer_mode"),
						knownvalue.StringExact("metallb"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("storage_classes"),
						knownvalue.NotNull(),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMKubeClusterResourceConfigP3Storage(clusterName, masterNodeUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update storage configuration
			{
				Config: testAccCMKubeClusterResourceConfigP3StorageUpdated(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("load_balancer_mode"),
						knownvalue.StringExact("haproxy"),
					),
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_P3Addons tests P3 cluster addons.
// DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_appGroups for new API.
func TestAccCMKubeClusterResource_P3Addons(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_appGroups for new API.")
	clusterName := generateUniqueTestName("tftest-cluster-p3-addons")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with addons and ingress controller
			{
				Config: testAccCMKubeClusterResourceConfigP3Addons(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("addons"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("ingress_controller"),
						knownvalue.NotNull(),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMKubeClusterResourceConfigP3Addons(clusterName, masterNodeUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Update addons configuration
			{
				Config: testAccCMKubeClusterResourceConfigP3AddonsUpdated(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					// Both should still be populated after update
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("addons"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("ingress_controller"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_P3FullStack tests all P3 fields together.
// DEPRECATED: Uses old schema with deprecated fields. See TestAccCMKubeCluster_aligned_* for new API.
func TestAccCMKubeClusterResource_P3FullStack(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with deprecated fields. See aligned tests for new API.")
	clusterName := generateUniqueTestName("tftest-cluster-p3-full")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with all P3 fields
			{
				Config: testAccCMKubeClusterResourceConfigP3Full(clusterName, masterNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify all P3 fields are set
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("cni_plugin"),
						knownvalue.StringExact("calico"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("dns_servers"),
						knownvalue.ListSizeExact(2),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("load_balancer_mode"),
						knownvalue.StringExact("metallb"),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("storage_classes"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("addons"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("ingress_controller"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("overlay_network"),
						knownvalue.StringExact("overlay-full-uuid"),
					),
				},
			},
			// Idempotency check with all P3 fields
			{
				Config: testAccCMKubeClusterResourceConfigP3Full(clusterName, masterNodeUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// P3 Test Configuration Helpers

func testAccCMKubeClusterResourceConfigP3Networking(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name            = %[4]q
  master_nodes    = [%[5]q]
  version         = "1.28.0"
  cni_plugin      = "calico"
  overlay_network = "overlay-test-uuid"
  dns_servers     = ["8.8.8.8", "8.8.4.4"]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

func testAccCMKubeClusterResourceConfigP3NetworkingUpdated(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name            = %[4]q
  master_nodes    = [%[5]q]
  version         = "1.29.0"
  cni_plugin      = "flannel"
  overlay_network = "overlay-updated-uuid"
  dns_servers     = ["1.1.1.1"]
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

func testAccCMKubeClusterResourceConfigP3Storage(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name               = %[4]q
  master_nodes       = [%[5]q]
  version            = "1.28.0"
  load_balancer_mode = "metallb"

  storage_classes = jsonencode([
    {
      name        = "fast-ssd"
      provisioner = "kubernetes.io/csi-driver"
      parameters  = {
        type = "ssd"
        iops = "3000"
      }
    }
  ])
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

func testAccCMKubeClusterResourceConfigP3StorageUpdated(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name               = %[4]q
  master_nodes       = [%[5]q]
  version            = "1.29.0"
  load_balancer_mode = "haproxy"

  storage_classes = jsonencode([
    {
      name        = "standard"
      provisioner = "kubernetes.io/csi-driver"
      parameters  = {
        type = "standard"
      }
    },
    {
      name        = "fast"
      provisioner = "kubernetes.io/csi-driver"
      parameters  = {
        type = "nvme"
      }
    }
  ])
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

func testAccCMKubeClusterResourceConfigP3Addons(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]
  version      = "1.28.0"

  addons = jsonencode([
    {
      name    = "prometheus"
      enabled = true
      version = "2.45.0"
      config  = {
        retention = "30d"
        storage   = "100Gi"
      }
    }
  ])

  ingress_controller = jsonencode({
    type    = "nginx"
    enabled = true
    version = "1.8.0"
    config  = {
      replicaCount = 2
    }
  })
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

func testAccCMKubeClusterResourceConfigP3AddonsUpdated(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]
  version      = "1.29.0"

  addons = jsonencode([
    {
      name    = "prometheus"
      enabled = true
      version = "2.46.0"
    },
    {
      name    = "grafana"
      enabled = true
      version = "10.0.0"
    }
  ])

  ingress_controller = jsonencode({
    type    = "traefik"
    enabled = true
    version = "2.10.0"
    config  = {
      replicaCount = 3
    }
  })
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

func testAccCMKubeClusterResourceConfigP3Full(name, masterNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name               = %[4]q
  master_nodes       = [%[5]q]
  version            = "1.29.0"
  cni_plugin         = "calico"
  overlay_network    = "overlay-full-uuid"
  dns_servers        = ["8.8.8.8", "8.8.4.4"]
  load_balancer_mode = "metallb"

  storage_classes = jsonencode([
    {
      name        = "fast-ssd"
      provisioner = "kubernetes.io/csi-driver"
      parameters  = { type = "ssd" }
    }
  ])

  addons = jsonencode([
    {
      name    = "prometheus"
      enabled = true
    },
    {
      name    = "grafana"
      enabled = true
    }
  ])

  ingress_controller = jsonencode({
    type    = "nginx"
    enabled = true
    config  = { replicaCount = 3 }
  })
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
	)
}

// TestAccCMKubeClusterResource_ForceOption tests the force parameter.
// DEPRECATED: Uses old schema with deprecated force field. See TestAccCMKubeCluster_aligned_* for new API.
func TestAccCMKubeClusterResource_ForceOption(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with deprecated force field. See aligned tests for new API.")
	clusterName := generateUniqueTestName("tftest-cluster-force")
	masterNodeUUID := getTestMasterNodeUUID(t)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with force = false (default)
			{
				Config: testAccCMKubeClusterResourceConfigWithForce(clusterName, masterNodeUUID, false),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("force"),
						knownvalue.Bool(false),
					),
				},
			},
			// Update force to true
			{
				Config: testAccCMKubeClusterResourceConfigWithForce(clusterName, masterNodeUUID, true),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("force"),
						knownvalue.Bool(true),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMKubeClusterResourceConfigWithForce(clusterName, masterNodeUUID, true),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

func testAccCMKubeClusterResourceConfigWithForce(name, masterNodeUUID string, force bool) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]
  force        = %[6]t
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
		force,
	)
}

// TestAccCMKubeClusterResource_EtcdNodes tests etcd_nodes attribute for HA clusters.
// NVIDIA DGX BasePOD deployments require dedicated etcd nodes for production HA.
// DEPRECATED: Uses old schema with deprecated etcd_nodes field. Use bcm_cmetcd_cluster + device roles instead.
func TestAccCMKubeClusterResource_EtcdNodes(t *testing.T) {
	t.Skip("DEPRECATED: Uses old schema with etcd_nodes. Use bcm_cmetcd_cluster + bcm_cmdevice_device etcd_host_role instead.")
	clusterName := generateUniqueTestName("tftest-cluster-etcd")
	masterNodeUUID := getTestMasterNodeUUID(t)
	etcdNodeUUID := getTestEtcdNodeUUID(t, 0)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with etcd_nodes
			{
				Config: testAccCMKubeClusterResourceConfigWithEtcdNodes(clusterName, masterNodeUUID, etcdNodeUUID),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("etcd_nodes"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
			// Idempotency check - etcd_nodes preserved across plan
			{
				Config: testAccCMKubeClusterResourceConfigWithEtcdNodes(clusterName, masterNodeUUID, etcdNodeUUID),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
			// Import - etcd_nodes should be ignored (write-only field)
			{
				ResourceName:            "bcm_cmkube_cluster.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"master_nodes", "worker_nodes", "etcd_nodes"},
			},
		},
	})
}

func testAccCMKubeClusterResourceConfigWithEtcdNodes(name, masterNodeUUID, etcdNodeUUID string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = %[4]q
  master_nodes = [%[5]q]
  etcd_nodes   = [%[6]q]
  version      = "1.28.0"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		masterNodeUUID,
		etcdNodeUUID,
	)
}
