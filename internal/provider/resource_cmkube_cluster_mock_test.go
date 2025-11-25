// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/compare"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// ========================================
// Mock BCM Server for CMKube Cluster Tests
// ========================================

// mockClusterState tracks cluster state for CRUD operations
type mockClusterState struct {
	mu       sync.RWMutex
	clusters map[string]map[string]interface{}
}

func newMockClusterState() *mockClusterState {
	return &mockClusterState{
		clusters: make(map[string]map[string]interface{}),
	}
}

// createMockBCMServerForKubeCluster creates a stateful mock BCM server for cluster testing.
// Supports full CRUD lifecycle including etcd_nodes handling.
func createMockBCMServerForKubeCluster() (*httptest.Server, *mockClusterState) {
	state := newMockClusterState()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Service  string        `json:"service"`
			Call     string        `json:"call"`
			Args     []interface{} `json:"args,omitempty"`
			Username string        `json:"username,omitempty"`
			Password string        `json:"password,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Handle login
		if req.Service == "login" {
			w.Header().Set("Set-Cookie", "cm-login-token=mock-token; Path=/")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(true)
			return
		}

		// Route cmkube API calls
		if req.Service == "cmkube" {
			handleCMKubeCall(w, req, state)
			return
		}

		// Default success for other services
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))

	return server, state
}

// handleCMKubeCall routes cmkube service API calls
func handleCMKubeCall(w http.ResponseWriter, req struct {
	Service  string        `json:"service"`
	Call     string        `json:"call"`
	Args     []interface{} `json:"args,omitempty"`
	Username string        `json:"username,omitempty"`
	Password string        `json:"password,omitempty"`
}, state *mockClusterState) {
	w.Header().Set("Content-Type", "application/json")

	switch req.Call {
	case "validateKubeCluster":
		// Return empty array (no validation errors)
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode([]interface{}{})

	case "addKubeCluster":
		handleAddKubeCluster(w, req.Args, state)

	case "getKubeCluster":
		handleGetKubeCluster(w, req.Args, state)

	case "updateKubeCluster":
		handleUpdateKubeCluster(w, req.Args, state)

	case "removeKubeCluster":
		handleRemoveKubeCluster(w, req.Args, state)

	case "getKubeClusters":
		handleGetKubeClusters(w, state)

	default:
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}
}

// handleAddKubeCluster handles cluster creation
func handleAddKubeCluster(w http.ResponseWriter, args []interface{}, state *mockClusterState) {
	if len(args) < 1 {
		http.Error(w, "Missing cluster entity", http.StatusBadRequest)
		return
	}

	entity, ok := args[0].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid cluster entity", http.StatusBadRequest)
		return
	}

	uuid, _ := entity["uuid"].(string)
	if uuid == "" {
		uuid = generateUUID()
	}

	// Store cluster (simulating BCM behavior: etcdNodes is accepted but not returned)
	state.mu.Lock()
	state.clusters[uuid] = map[string]interface{}{
		"uuid":              uuid,
		"name":              entity["name"],
		"version":           entity["version"],
		"managementNetwork": entity["managementNetwork"],
		"creationTime":      1700000000,
		"revisionId":        1,
		// Note: masterNodes, workerNodes, etcdNodes are NOT stored/returned
		// This simulates BCM's write-only behavior for node lists
	}
	state.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"uuid":    uuid,
	})
}

// handleGetKubeCluster handles cluster read (simulates write-only fields)
func handleGetKubeCluster(w http.ResponseWriter, args []interface{}, state *mockClusterState) {
	if len(args) < 1 {
		http.Error(w, "Missing cluster UUID", http.StatusBadRequest)
		return
	}

	uuid, ok := args[0].(string)
	if !ok {
		http.Error(w, "Invalid cluster UUID", http.StatusBadRequest)
		return
	}

	state.mu.RLock()
	cluster, exists := state.clusters[uuid]
	state.mu.RUnlock()

	if !exists {
		http.Error(w, "Cluster not found", http.StatusNotFound)
		return
	}

	// Return cluster WITHOUT masterNodes, workerNodes, etcdNodes
	// This simulates BCM's actual behavior where these are write-only fields
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(cluster)
}

// handleUpdateKubeCluster handles cluster update
func handleUpdateKubeCluster(w http.ResponseWriter, args []interface{}, state *mockClusterState) {
	if len(args) < 1 {
		http.Error(w, "Missing cluster entity", http.StatusBadRequest)
		return
	}

	entity, ok := args[0].(map[string]interface{})
	if !ok {
		http.Error(w, "Invalid cluster entity", http.StatusBadRequest)
		return
	}

	uuid, _ := entity["uuid"].(string)

	state.mu.Lock()
	if _, exists := state.clusters[uuid]; exists {
		// Update stored fields (but NOT node lists - they're write-only)
		state.clusters[uuid]["name"] = entity["name"]
		state.clusters[uuid]["version"] = entity["version"]
		state.clusters[uuid]["managementNetwork"] = entity["managementNetwork"]
		state.clusters[uuid]["revisionId"] = 2
	}
	state.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleRemoveKubeCluster handles cluster deletion
func handleRemoveKubeCluster(w http.ResponseWriter, args []interface{}, state *mockClusterState) {
	if len(args) < 1 {
		http.Error(w, "Missing cluster UUID", http.StatusBadRequest)
		return
	}

	uuid, ok := args[0].(string)
	if !ok {
		http.Error(w, "Invalid cluster UUID", http.StatusBadRequest)
		return
	}

	state.mu.Lock()
	delete(state.clusters, uuid)
	state.mu.Unlock()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleGetKubeClusters handles listing all clusters
func handleGetKubeClusters(w http.ResponseWriter, state *mockClusterState) {
	state.mu.RLock()
	clusters := make([]map[string]interface{}, 0, len(state.clusters))
	for _, cluster := range state.clusters {
		clusters = append(clusters, cluster)
	}
	state.mu.RUnlock()

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(clusters)
}

// testAccCheckMockClusterDestroy creates a CheckDestroy function for mock cluster tests.
// It verifies that all clusters have been removed from the mock state.
func testAccCheckMockClusterDestroy(state *mockClusterState) resource.TestCheckFunc {
	return func(s *terraform.State) error {
		for _, rs := range s.RootModule().Resources {
			if rs.Type != "bcm_cmkube_cluster" {
				continue
			}

			id := rs.Primary.ID

			state.mu.RLock()
			_, exists := state.clusters[id]
			state.mu.RUnlock()

			if exists {
				return fmt.Errorf("bcm_cmkube_cluster %s still exists in mock state after destroy", id)
			}
		}

		return nil
	}
}

// ========================================
// Mock Tests for etcd_nodes
// ========================================

// TestAccCMKubeClusterResource_MockEtcdNodes tests etcd_nodes with mock server.
// This test validates the write-only field behavior without requiring physical nodes.
func TestAccCMKubeClusterResource_MockEtcdNodes(t *testing.T) {
	cleanup := clearBCMEnvVars(t)
	defer cleanup()

	mockServer, mockState := createMockBCMServerForKubeCluster()
	defer mockServer.Close()

	clusterName := generateUniqueTestName("mock-etcd-cluster")

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMockClusterDestroy(mockState),
		Steps: []resource.TestStep{
			// Create cluster with etcd_nodes
			{
				Config: testAccCMKubeClusterMockConfigWithEtcdNodes(mockServer.URL, clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("etcd_nodes"),
						knownvalue.ListSizeExact(3),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check - etcd_nodes should be preserved
			{
				Config: testAccCMKubeClusterMockConfigWithEtcdNodes(mockServer.URL, clusterName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_MockEtcdNodesUpdate tests updating etcd_nodes.
func TestAccCMKubeClusterResource_MockEtcdNodesUpdate(t *testing.T) {
	cleanup := clearBCMEnvVars(t)
	defer cleanup()

	mockServer, mockState := createMockBCMServerForKubeCluster()
	defer mockServer.Close()

	clusterName := generateUniqueTestName("mock-etcd-update")

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMockClusterDestroy(mockState),
		Steps: []resource.TestStep{
			// Create with 3 etcd nodes
			{
				Config: testAccCMKubeClusterMockConfigWithEtcdNodes(mockServer.URL, clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("etcd_nodes"),
						knownvalue.ListSizeExact(3),
					),
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update to 5 etcd nodes
			{
				Config: testAccCMKubeClusterMockConfigWithEtcdNodesCount(mockServer.URL, clusterName, 5),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("etcd_nodes"),
						knownvalue.ListSizeExact(5),
					),
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency after update
			{
				Config: testAccCMKubeClusterMockConfigWithEtcdNodesCount(mockServer.URL, clusterName, 5),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// TestAccCMKubeClusterResource_MockEtcdNodesValidationError tests validation error for etcd_nodes.
func TestAccCMKubeClusterResource_MockEtcdNodesValidationError(t *testing.T) {
	cleanup := clearBCMEnvVars(t)
	defer cleanup()

	// Create mock server that returns validation ERROR for invalid etcd nodes count
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			Service string        `json:"service"`
			Call    string        `json:"call"`
			Args    []interface{} `json:"args,omitempty"`
		}
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.Service == "login" {
			w.Header().Set("Set-Cookie", "cm-login-token=mock-token; Path=/")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(true)
			return
		}

		if req.Service == "cmkube" && req.Call == "validateKubeCluster" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			// Use ERROR severity to halt execution (WARNING would allow continuation)
			_ = json.NewEncoder(w).Encode([]map[string]interface{}{
				{
					"Field":    "etcdNodes",
					"Message":  "etcd cluster requires odd number of nodes (1, 3, or 5) for quorum",
					"Severity": "ERROR",
				},
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
	}))
	defer mockServer.Close()

	clusterName := generateUniqueTestName("mock-etcd-val")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMKubeClusterMockConfigWithEtcdNodesCount(mockServer.URL, clusterName, 2),
				ExpectError: regexp.MustCompile(`(?is)Validation Error.*etcdNodes.*quorum`),
			},
		},
	})
}

// TestAccCMKubeClusterResource_MockEtcdNodesNullToValue tests adding etcd_nodes to existing cluster.
func TestAccCMKubeClusterResource_MockEtcdNodesNullToValue(t *testing.T) {
	cleanup := clearBCMEnvVars(t)
	defer cleanup()

	mockServer, mockState := createMockBCMServerForKubeCluster()
	defer mockServer.Close()

	clusterName := generateUniqueTestName("mock-etcd-null")

	// ID consistency tracking across test steps
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckMockClusterDestroy(mockState),
		Steps: []resource.TestStep{
			// Create without etcd_nodes
			{
				Config: testAccCMKubeClusterMockConfigBasic(mockServer.URL, clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Add etcd_nodes
			{
				Config: testAccCMKubeClusterMockConfigWithEtcdNodes(mockServer.URL, clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("etcd_nodes"),
						knownvalue.ListSizeExact(3),
					),
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency
			{
				Config: testAccCMKubeClusterMockConfigWithEtcdNodes(mockServer.URL, clusterName),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmkube_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// ========================================
// Mock Test Configuration Helpers
// ========================================

func testAccCMKubeClusterMockConfigBasic(mockEndpoint, name string) string {
	return `
provider "bcm" {
  endpoint             = "` + mockEndpoint + `"
  username             = "mock-user"
  password             = "mock-pass"
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = "` + name + `"
  master_nodes = ["master-uuid-1"]
  version      = "1.28.0"
}
`
}

func testAccCMKubeClusterMockConfigWithEtcdNodes(mockEndpoint, name string) string {
	return `
provider "bcm" {
  endpoint             = "` + mockEndpoint + `"
  username             = "mock-user"
  password             = "mock-pass"
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = "` + name + `"
  master_nodes = ["master-uuid-1", "master-uuid-2", "master-uuid-3"]
  etcd_nodes   = ["etcd-uuid-1", "etcd-uuid-2", "etcd-uuid-3"]
  version      = "1.28.0"
}
`
}

func testAccCMKubeClusterMockConfigWithEtcdNodesCount(mockEndpoint, name string, count int) string {
	etcdNodes := ""
	for i := 1; i <= count; i++ {
		if i > 1 {
			etcdNodes += ", "
		}
		etcdNodes += fmt.Sprintf(`"etcd-uuid-%d"`, i)
	}

	return `
provider "bcm" {
  endpoint             = "` + mockEndpoint + `"
  username             = "mock-user"
  password             = "mock-pass"
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "test" {
  name         = "` + name + `"
  master_nodes = ["master-uuid-1"]
  etcd_nodes   = [` + etcdNodes + `]
  version      = "1.28.0"
}
`
}
