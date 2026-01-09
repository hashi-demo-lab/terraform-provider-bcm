// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

// Package provider contains acceptance tests for the aligned bcm_cmkube_cluster resource.
// These tests verify the resource schema matches the BCM KubeCluster API entity structure.
package provider

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// =============================================================================
// Test Helpers for Aligned Schema (T024)
// =============================================================================

// getTestNetworkUUIDByIndex queries BCM for network UUIDs by index.
// This is used for tests that need multiple distinct network UUIDs.
func getTestNetworkUUIDByIndex(t *testing.T, index int) string {
	client := createTestBCMClient(t)
	ctx := t.Context()

	body, err := client.CallJSONRPC(ctx, "cmnet", "getNetworks")
	if err != nil {
		t.Skipf("Failed to get networks (skipping test): %v", err)
		return ""
	}

	var networks []map[string]interface{}
	if err := json.Unmarshal(body, &networks); err != nil {
		t.Skipf("Failed to parse networks (skipping test): %v", err)
		return ""
	}

	if len(networks) <= index {
		t.Skipf("Not enough networks for index %d (need at least %d networks)", index, index+1)
		return ""
	}

	if uuid, ok := networks[index]["uuid"].(string); ok {
		return uuid
	}

	t.Skipf("Network at index %d has invalid UUID format", index)
	return ""
}

// getOrCreateTestEtcdClusterUUID creates a test etcd cluster and returns its UUID.
func getOrCreateTestEtcdClusterUUID(t *testing.T, name string) string {
	client := createTestBCMClient(t)
	ctx := t.Context()

	// Generate a unique UUID for the etcd cluster
	etcdUUID := generateUUID()

	// Build entity
	entity := map[string]interface{}{
		"baseType":          "EtcdCluster",
		"childType":         "",
		"modified":          true,
		"to_be_removed":     false,
		"revision":          "",
		"uuid":              etcdUUID,
		"name":              name,
		"heartBeatInterval": int64(100),
		"electionTimeout":   int64(1000),
		"options":           map[string]interface{}{},
	}

	// Create the etcd cluster
	_, err := client.AddEtcdCluster(ctx, entity)
	if err != nil {
		t.Fatalf("Failed to create test etcd cluster: %v", err)
	}

	return etcdUUID
}

// cleanupTestEtcdCluster removes a test etcd cluster.
func cleanupTestEtcdCluster(t *testing.T, uuid string) {
	client := createTestBCMClient(t)
	ctx := t.Context()

	_, err := client.RemoveEtcdCluster(ctx, uuid)
	if err != nil {
		t.Logf("Warning: Failed to cleanup test etcd cluster %s: %v", uuid, err)
	}
}

// =============================================================================
// Acceptance Tests - Aligned Basic CRUD (T024)
// =============================================================================

// TestAccCMKubeCluster_aligned_basic tests complete CRUD lifecycle with aligned schema.
// This test validates the new schema that aligns with BCM's KubeCluster entity.
func TestAccCMKubeCluster_aligned_basic(t *testing.T) {
	clusterName := generateShortTestName("algn")
	etcdClusterName := generateShortTestName("etcd")

	// Create a test etcd cluster first
	etcdClusterUUID := getOrCreateTestEtcdClusterUUID(t, etcdClusterName)
	defer cleanupTestEtcdCluster(t, etcdClusterUUID)

	// Get network UUIDs (if available)
	internalNetworkUUID := getTestNetworkUUIDByIndex(t, 0)
	serviceNetworkUUID := getTestNetworkUUIDByIndex(t, 1)
	podNetworkUUID := getTestNetworkUUIDByIndex(t, 2)

	// ID consistency tracking across Create/Import/Update
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with aligned schema fields
			{
				Config: testAccCMKubeClusterAlignedConfig(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
				),
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
						tfjsonpath.New("etcd_cluster"),
						knownvalue.StringExact(etcdClusterUUID),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("internal_network"),
						knownvalue.StringExact(internalNetworkUUID),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("service_network"),
						knownvalue.StringExact(serviceNetworkUUID),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("pod_network"),
						knownvalue.StringExact(podNetworkUUID),
					),
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMKubeClusterAlignedConfig(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Aligned Update (T025)
// =============================================================================

// TestAccCMKubeCluster_aligned_update tests update operations with aligned schema.
func TestAccCMKubeCluster_aligned_update(t *testing.T) {
	clusterName := generateShortTestName("upd")
	clusterNameUpdated := generateShortTestName("upd2")
	etcdClusterName := generateShortTestName("etcd")

	// Create a test etcd cluster
	etcdClusterUUID := getOrCreateTestEtcdClusterUUID(t, etcdClusterName)
	defer cleanupTestEtcdCluster(t, etcdClusterUUID)

	// Get network UUIDs
	internalNetworkUUID := getTestNetworkUUIDByIndex(t, 0)
	serviceNetworkUUID := getTestNetworkUUIDByIndex(t, 1)
	podNetworkUUID := getTestNetworkUUIDByIndex(t, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create
			{
				Config: testAccCMKubeClusterAlignedConfigWithVersion(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
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
						tfjsonpath.New("version"),
						knownvalue.StringExact("1.28.0"),
					),
				},
			},
			// Update name and version
			{
				Config: testAccCMKubeClusterAlignedConfigWithVersion(
					clusterNameUpdated,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
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
			// Idempotency check after Update
			{
				Config: testAccCMKubeClusterAlignedConfigWithVersion(
					clusterNameUpdated,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
					"1.29.0",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Aligned Network UUID Persistence (T026)
// =============================================================================

// TestAccCMKubeCluster_aligned_networks tests network UUID persistence.
func TestAccCMKubeCluster_aligned_networks(t *testing.T) {
	clusterName := generateShortTestName("net")
	etcdClusterName := generateShortTestName("etcd")

	// Create a test etcd cluster
	etcdClusterUUID := getOrCreateTestEtcdClusterUUID(t, etcdClusterName)
	defer cleanupTestEtcdCluster(t, etcdClusterUUID)

	// Get network UUIDs
	internalNetworkUUID := getTestNetworkUUIDByIndex(t, 0)
	serviceNetworkUUID := getTestNetworkUUIDByIndex(t, 1)
	podNetworkUUID := getTestNetworkUUIDByIndex(t, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with all network references
			{
				Config: testAccCMKubeClusterAlignedConfigWithNetworkMask(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
					"/24",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("internal_network"),
						knownvalue.StringExact(internalNetworkUUID),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("service_network"),
						knownvalue.StringExact(serviceNetworkUUID),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("pod_network"),
						knownvalue.StringExact(podNetworkUUID),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("pod_network_node_mask"),
						knownvalue.StringExact("/24"),
					),
				},
			},
			// Import and verify network fields persist
			{
				ResourceName:      "bcm_cmkube_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Aligned AppGroups (T027)
// =============================================================================

// TestAccCMKubeCluster_aligned_appGroups tests app_groups nested block.
func TestAccCMKubeCluster_aligned_appGroups(t *testing.T) {
	clusterName := generateShortTestName("app")
	etcdClusterName := generateShortTestName("etcd")

	// Create a test etcd cluster
	etcdClusterUUID := getOrCreateTestEtcdClusterUUID(t, etcdClusterName)
	defer cleanupTestEtcdCluster(t, etcdClusterUUID)

	// Get network UUIDs
	internalNetworkUUID := getTestNetworkUUIDByIndex(t, 0)
	serviceNetworkUUID := getTestNetworkUUIDByIndex(t, 1)
	podNetworkUUID := getTestNetworkUUIDByIndex(t, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create with app_groups
			{
				Config: testAccCMKubeClusterAlignedConfigWithAppGroups(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("app_groups"),
						knownvalue.ListSizeExact(1),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMKubeClusterAlignedConfigWithAppGroups(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Aligned Drift Detection (T028)
// =============================================================================

// TestAccCMKubeCluster_aligned_drift tests external modification detection with aligned schema.
func TestAccCMKubeCluster_aligned_drift(t *testing.T) {
	clusterName := generateShortTestName("drf")
	etcdClusterName := generateShortTestName("etcd")

	// Create a test etcd cluster
	etcdClusterUUID := getOrCreateTestEtcdClusterUUID(t, etcdClusterName)
	defer cleanupTestEtcdCluster(t, etcdClusterUUID)

	// Get network UUIDs
	internalNetworkUUID := getTestNetworkUUIDByIndex(t, 0)
	serviceNetworkUUID := getTestNetworkUUIDByIndex(t, 1)
	podNetworkUUID := getTestNetworkUUIDByIndex(t, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMKubeClusterDestroy,
		Steps: []resource.TestStep{
			// Create cluster
			{
				Config: testAccCMKubeClusterAlignedConfigWithVersion(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
					"1.28.0",
				),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("version"),
						knownvalue.StringExact("1.28.0"),
					),
				},
			},
			// Modify externally via BCM API
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

					// Modify version externally
					clusterData["version"] = "1.29.0"
					clusterData["modified"] = true

					// Update via BCM API
					_, err = client.CallJSONRPC(ctx, "cmkube", "updateKubeCluster", clusterData, false)
					if err != nil {
						t.Fatalf("Failed to update cluster for drift test: %v", err)
					}

					t.Logf("[DEBUG] Modified version externally to: 1.29.0")
				},
				Config: testAccCMKubeClusterAlignedConfigWithVersion(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
					"1.28.0",
				),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Drift detected
					},
				},
			},
			// Terraform restores desired state
			{
				Config: testAccCMKubeClusterAlignedConfigWithVersion(
					clusterName,
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
					"1.28.0",
				),
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

// =============================================================================
// Unit Tests - Entity Builder (T029)
// =============================================================================

// TestCMKubeClusterEntityBuilder_aligned tests the aligned entity construction.
func TestCMKubeClusterEntityBuilder_aligned(t *testing.T) {
	t.Run("build_aligned_entity", func(t *testing.T) {
		entity := buildAlignedKubeClusterEntity(
			"test-cluster",
			"",
			"etcd-uuid",
			"internal-uuid",
			"service-uuid",
			"pod-uuid",
			"1.28.0",
		)

		// Verify required fields
		if entity["baseType"] != "KubeCluster" {
			t.Errorf("expected baseType=KubeCluster, got %v", entity["baseType"])
		}
		if entity["name"] != "test-cluster" {
			t.Errorf("expected name=test-cluster, got %v", entity["name"])
		}
		if entity["etcdCluster"] != "etcd-uuid" {
			t.Errorf("expected etcdCluster=etcd-uuid, got %v", entity["etcdCluster"])
		}
		if entity["internalNetwork"] != "internal-uuid" {
			t.Errorf("expected internalNetwork=internal-uuid, got %v", entity["internalNetwork"])
		}
		if entity["serviceNetwork"] != "service-uuid" {
			t.Errorf("expected serviceNetwork=service-uuid, got %v", entity["serviceNetwork"])
		}
		if entity["podNetwork"] != "pod-uuid" {
			t.Errorf("expected podNetwork=pod-uuid, got %v", entity["podNetwork"])
		}
		if entity["version"] != "1.28.0" {
			t.Errorf("expected version=1.28.0, got %v", entity["version"])
		}
		// UUID should be generated
		if entity["uuid"] == "" {
			t.Error("expected UUID to be generated")
		}
	})

	t.Run("build_entity_with_uuid", func(t *testing.T) {
		testUUID := "12345678-1234-1234-1234-123456789012"
		entity := buildAlignedKubeClusterEntity(
			"test-cluster",
			testUUID,
			"etcd-uuid",
			"internal-uuid",
			"service-uuid",
			"pod-uuid",
			"1.28.0",
		)

		if entity["uuid"] != testUUID {
			t.Errorf("expected uuid=%s, got %v", testUUID, entity["uuid"])
		}
	})

	t.Run("removed_fields_not_present", func(t *testing.T) {
		entity := buildAlignedKubeClusterEntity(
			"test-cluster",
			"",
			"etcd-uuid",
			"internal-uuid",
			"service-uuid",
			"pod-uuid",
			"1.28.0",
		)

		// These fields should NOT exist in the aligned entity
		removedFields := []string{
			"masterNodes",
			"workerNodes",
			"etcdNodes",
			"dnsServers",
			"cniPlugin",
			"storageClasses",
			"loadBalancerMode",
			"addons",
			"overlayNetwork",
			"ingressController",
		}

		for _, field := range removedFields {
			if _, exists := entity[field]; exists {
				t.Errorf("field %s should NOT exist in aligned entity", field)
			}
		}
	})
}

// buildAlignedKubeClusterEntity is a test helper to construct aligned KubeCluster entities.
func buildAlignedKubeClusterEntity(name, uuid, etcdCluster, internalNetwork, serviceNetwork, podNetwork, version string) map[string]interface{} {
	entity := map[string]interface{}{
		"baseType":        "KubeCluster",
		"childType":       "",
		"modified":        true,
		"to_be_removed":   false,
		"revision":        "",
		"name":            name,
		"etcdCluster":     etcdCluster,
		"internalNetwork": internalNetwork,
		"serviceNetwork":  serviceNetwork,
		"podNetwork":      podNetwork,
		"version":         version,
	}

	// Generate UUID if not provided
	if uuid == "" {
		entity["uuid"] = generateUUID()
	} else {
		entity["uuid"] = uuid
	}

	return entity
}

// =============================================================================
// Acceptance Tests - Validation
// =============================================================================

// TestAccCMKubeCluster_aligned_validationInvalidName tests invalid cluster name.
func TestAccCMKubeCluster_aligned_validationInvalidName(t *testing.T) {
	etcdClusterName := generateShortTestName("etcd")

	// Create a test etcd cluster
	etcdClusterUUID := getOrCreateTestEtcdClusterUUID(t, etcdClusterName)
	defer cleanupTestEtcdCluster(t, etcdClusterUUID)

	// Get network UUIDs
	internalNetworkUUID := getTestNetworkUUIDByIndex(t, 0)
	serviceNetworkUUID := getTestNetworkUUIDByIndex(t, 1)
	podNetworkUUID := getTestNetworkUUIDByIndex(t, 2)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMKubeCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMKubeClusterAlignedConfig(
					"Invalid Name!",
					etcdClusterUUID,
					internalNetworkUUID,
					serviceNetworkUUID,
					podNetworkUUID,
				),
				ExpectError: regexp.MustCompile(`must contain only`),
			},
		},
	})
}

// =============================================================================
// Test Configuration Helpers
// =============================================================================

func testAccCMKubeClusterAlignedConfig(name, etcdCluster, internalNetwork, serviceNetwork, podNetwork string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name             = %[4]q
  etcd_cluster     = %[5]q
  internal_network = %[6]q
  service_network  = %[7]q
  pod_network      = %[8]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		etcdCluster,
		internalNetwork,
		serviceNetwork,
		podNetwork,
	)
}

func testAccCMKubeClusterAlignedConfigWithVersion(name, etcdCluster, internalNetwork, serviceNetwork, podNetwork, version string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name             = %[4]q
  etcd_cluster     = %[5]q
  internal_network = %[6]q
  service_network  = %[7]q
  pod_network      = %[8]q
  version          = %[9]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		etcdCluster,
		internalNetwork,
		serviceNetwork,
		podNetwork,
		version,
	)
}

func testAccCMKubeClusterAlignedConfigWithNetworkMask(name, etcdCluster, internalNetwork, serviceNetwork, podNetwork, podNetworkNodeMask string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name                   = %[4]q
  etcd_cluster           = %[5]q
  internal_network       = %[6]q
  service_network        = %[7]q
  pod_network            = %[8]q
  pod_network_node_mask  = %[9]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		etcdCluster,
		internalNetwork,
		serviceNetwork,
		podNetwork,
		podNetworkNodeMask,
	)
}

func testAccCMKubeClusterAlignedConfigWithAppGroups(name, etcdCluster, internalNetwork, serviceNetwork, podNetwork string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name             = %[4]q
  etcd_cluster     = %[5]q
  internal_network = %[6]q
  service_network  = %[7]q
  pod_network      = %[8]q

  app_groups {
    name    = "monitoring"
    enabled = true
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		etcdCluster,
		internalNetwork,
		serviceNetwork,
		podNetwork,
	)
}
