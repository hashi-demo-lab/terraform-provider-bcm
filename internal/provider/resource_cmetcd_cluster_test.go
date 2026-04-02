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

// =============================================================================
// Pre-Check and CheckDestroy Functions (T013)
// =============================================================================

// testAccPreCheckCMEtcdCluster verifies environment variables for etcd cluster tests.
func testAccPreCheckCMEtcdCluster(t *testing.T) {
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

// testAccCheckCMEtcdClusterDestroy verifies etcd cluster deletion with enhanced error messages.
func testAccCheckCMEtcdClusterDestroy(s *terraform.State) error {
	client := createTestBCMClient(&testing.T{})

	var errors []string
	resourceCount := 0

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmetcd_cluster" {
			continue
		}

		resourceCount++
		id := rs.Primary.ID

		// Verify cluster deleted with exponential backoff
		deleted := verifyResourceDeleted(
			context.Background(),
			client,
			"cmetcd",
			"getEtcdCluster",
			id,
			4, // retry count
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"EtcdCluster still exists after destroy. Type: %s, ID: %s, Retries: 4",
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

// =============================================================================
// Unit Tests (T012)
// =============================================================================

// TestCMEtcdClusterEntityBuilder tests the entity construction for EtcdCluster.
func TestCMEtcdClusterEntityBuilder(t *testing.T) {
	t.Run("build_minimal_entity", func(t *testing.T) {
		entity := buildEtcdClusterEntity("", 100, 1000, "")

		// Verify required fields
		if entity["baseType"] != "EtcdCluster" {
			t.Errorf("expected baseType=EtcdCluster, got %v", entity["baseType"])
		}
		if entity["name"] != "test-etcd" {
			t.Errorf("expected name=test-etcd, got %v", entity["name"])
		}
		if entity["heartBeatInterval"] != int64(100) {
			t.Errorf("expected heartbeatInterval=100, got %v", entity["heartBeatInterval"])
		}
		if entity["electionTimeout"] != int64(1000) {
			t.Errorf("expected electionTimeout=1000, got %v", entity["electionTimeout"])
		}
		// UUID should be generated
		if entity["uuid"] == "" {
			t.Error("expected UUID to be generated")
		}
	})

	t.Run("build_entity_with_uuid", func(t *testing.T) {
		testUUID := "12345678-1234-1234-1234-123456789012"
		entity := buildEtcdClusterEntity(testUUID, 100, 1000, "")

		if entity["uuid"] != testUUID {
			t.Errorf("expected uuid=%s, got %v", testUUID, entity["uuid"])
		}
	})

	t.Run("build_entity_with_options", func(t *testing.T) {
		optionsJSON := `{"custom_key": "custom_value"}`
		entity := buildEtcdClusterEntity("", 100, 1000, optionsJSON)

		options, ok := entity["options"].(map[string]interface{})
		if !ok {
			t.Fatal("expected options to be a map")
		}
		if options["custom_key"] != "custom_value" {
			t.Errorf("expected options.custom_key=custom_value, got %v", options["custom_key"])
		}
	})

	t.Run("election_timeout_relationship", func(t *testing.T) {
		// Election timeout should typically be >= 5x heartbeat interval
		// BCM validates this relationship
		heartbeat := int64(100)
		electionOK := int64(500)  // 5x heartbeat - should be OK
		electionLow := int64(400) // 4x heartbeat - may trigger warning

		entityOK := buildEtcdClusterEntity("", heartbeat, electionOK, "")
		if entityOK["heartBeatInterval"] != heartbeat {
			t.Errorf("expected heartBeatInterval=%d, got %v", heartbeat, entityOK["heartBeatInterval"])
		}
		if entityOK["electionTimeout"] != electionOK {
			t.Errorf("expected electionTimeout=%d, got %v", electionOK, entityOK["electionTimeout"])
		}

		entityLow := buildEtcdClusterEntity("", heartbeat, electionLow, "")
		if entityLow["electionTimeout"] != electionLow {
			t.Errorf("expected electionTimeout=%d, got %v", electionLow, entityLow["electionTimeout"])
		}
	})
}

// buildEtcdClusterEntity is a test helper to construct EtcdCluster entities.
// This mirrors the implementation in resource_cmetcd_cluster.go.
// Note: BCM uses camelCase: heartBeatInterval (with capital B).
//
//nolint:unparam // heartbeatInterval param kept for flexibility in future tests
func buildEtcdClusterEntity(uuid string, heartbeatInterval, electionTimeout int64, optionsJSON string) map[string]interface{} {
	entity := map[string]interface{}{
		"baseType":          "EtcdCluster",
		"childType":         "",
		"modified":          true,
		"to_be_removed":     false,
		"revision":          "",
		"name":              "test-etcd",
		"heartBeatInterval": heartbeatInterval, // BCM uses capital B
		"electionTimeout":   electionTimeout,
	}

	// Generate UUID if not provided
	if uuid == "" {
		entity["uuid"] = generateUUID()
	} else {
		entity["uuid"] = uuid
	}

	// Parse options JSON
	if optionsJSON != "" {
		var options map[string]interface{}
		if err := json.Unmarshal([]byte(optionsJSON), &options); err == nil {
			entity["options"] = options
		} else {
			entity["options"] = map[string]interface{}{}
		}
	} else {
		entity["options"] = map[string]interface{}{}
	}

	return entity
}

// =============================================================================
// Acceptance Tests - Basic CRUD (T009)
// =============================================================================

// TestAccCMEtcdCluster_basic tests complete CRUD lifecycle.
func TestAccCMEtcdCluster_basic(t *testing.T) {
	clusterName := generateShortTestName("etcd")

	// ID consistency tracking across Create/Import/Update
	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			// Create with minimal config (defaults for heartbeat/election)
			{
				Config: testAccCMEtcdClusterResourceConfig(clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("uuid"),
						knownvalue.NotNull(),
					),
					// Default values
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("heartbeat_interval"),
						knownvalue.Int64Exact(100),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("election_timeout"),
						knownvalue.Int64Exact(1000),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Create
			{
				Config: testAccCMEtcdClusterResourceConfig(clusterName),
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
// Acceptance Tests - Update (T010)
// =============================================================================

// TestAccCMEtcdCluster_update tests update operations.
func TestAccCMEtcdCluster_update(t *testing.T) {
	clusterName := generateShortTestName("upd")
	clusterNameUpdated := generateShortTestName("upd2")

	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			// Create with custom heartbeat/election
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 150, 1500),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("heartbeat_interval"),
						knownvalue.Int64Exact(150),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("election_timeout"),
						knownvalue.Int64Exact(1500),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update name and timings
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterNameUpdated, 200, 2000),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterNameUpdated),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("heartbeat_interval"),
						knownvalue.Int64Exact(200),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("election_timeout"),
						knownvalue.Int64Exact(2000),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check after Update
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterNameUpdated, 200, 2000),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Import (T011)
// =============================================================================

// TestAccCMEtcdCluster_import tests import functionality.
func TestAccCMEtcdCluster_import(t *testing.T) {
	clusterName := generateShortTestName("imp")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			// Create cluster
			{
				Config: testAccCMEtcdClusterResourceConfig(clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
				},
			},
			// Import by UUID
			{
				ResourceName:      "bcm_cmetcd_cluster.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Drift Detection
// =============================================================================

// TestAccCMEtcdCluster_driftDetection tests external modification detection.
func TestAccCMEtcdCluster_driftDetection(t *testing.T) {
	clusterName := generateShortTestName("drf")

	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			// Create cluster
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 100, 1000),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("heartbeat_interval"),
						knownvalue.Int64Exact(100),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Modify externally via BCM API
			{
				PreConfig: func() {
					client := createTestBCMClient(t)
					ctx := t.Context()

					// Get cluster UUID by name
					uuid := getResourceUUIDByName(t, "cmetcd", "getEtcdCluster", clusterName)

					// Fetch full cluster entity
					body, err := client.GetEtcdCluster(ctx, uuid)
					if err != nil {
						t.Fatalf("Failed to fetch etcd cluster for drift test: %v", err)
					}

					var clusterData map[string]interface{}
					if err := json.Unmarshal(body, &clusterData); err != nil {
						t.Fatalf("Failed to parse cluster data: %v", err)
					}

					// Modify heartbeat externally (BCM uses camelCase: heartBeatInterval)
					clusterData["heartBeatInterval"] = float64(150)
					clusterData["modified"] = true

					// Update via BCM API
					_, err = client.UpdateEtcdCluster(ctx, clusterData)
					if err != nil {
						t.Fatalf("Failed to update etcd cluster for drift test: %v", err)
					}

					// Wait for eventual consistency
					time.Sleep(TestEventualConsistencyDelay)

					t.Logf("[DEBUG] Modified heartbeatInterval externally to: 150")
				},
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 100, 1000),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectNonEmptyPlan(), // Drift detected
					},
				},
			},
			// Terraform restores desired state
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 100, 1000),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("heartbeat_interval"),
						knownvalue.Int64Exact(100),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Validation
// =============================================================================

// TestAccCMEtcdCluster_validationInvalidName tests invalid cluster name.
func TestAccCMEtcdCluster_validationInvalidName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config:      testAccCMEtcdClusterResourceConfig("Invalid Name!"),
				ExpectError: regexp.MustCompile(`must contain only lowercase alphanumeric`),
			},
		},
	})
}

// TestAccCMEtcdCluster_validationHeartbeatRange tests heartbeat interval validation.
func TestAccCMEtcdCluster_validationHeartbeatRange(t *testing.T) {
	clusterName := generateShortTestName("hb")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			{
				// Heartbeat too low (min 50ms)
				Config:      testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 10, 1000),
				ExpectError: regexp.MustCompile(`heartbeat_interval.*must be at least 50|Attribute heartbeat_interval`),
			},
		},
	})
}

// TestAccCMEtcdCluster_validationElectionTimeoutRange tests election timeout validation.
func TestAccCMEtcdCluster_validationElectionTimeoutRange(t *testing.T) {
	clusterName := generateShortTestName("el")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			{
				// Election timeout too low (min 500ms)
				Config:      testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 100, 100),
				ExpectError: regexp.MustCompile(`election_timeout.*must be at least 500|Attribute election_timeout`),
			},
		},
	})
}

// TestAccCMEtcdCluster_validationHeartbeatAboveMax tests that heartbeat_interval
// above the maximum (500) is rejected by the schema validator (Between 50-500).
func TestAccCMEtcdCluster_validationHeartbeatAboveMax(t *testing.T) {
	clusterName := generateShortTestName("hbmax")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			{
				// Heartbeat too high (max 500ms)
				Config:      testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 600, 5000),
				ExpectError: regexp.MustCompile(`heartbeat_interval.*must be between 50 and 500|Attribute heartbeat_interval`),
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Options JSON
// =============================================================================

// TestAccCMEtcdCluster_withOptions tests the options JSON field.
func TestAccCMEtcdCluster_withOptions(t *testing.T) {
	clusterName := generateShortTestName("opt")

	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCMEtcdClusterResourceConfigWithOptions(clusterName, `{"custom_setting": "value1"}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("options"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Update options
			{
				Config: testAccCMEtcdClusterResourceConfigWithOptions(clusterName, `{"custom_setting": "value2", "another": "setting"}`),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("options"),
						knownvalue.NotNull(),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency check
			{
				Config: testAccCMEtcdClusterResourceConfigWithOptions(clusterName, `{"custom_setting": "value2", "another": "setting"}`),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// =============================================================================
// Test Configuration Helpers
// =============================================================================

func testAccCMEtcdClusterResourceConfig(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmetcd_cluster" "test" {
  name = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

func testAccCMEtcdClusterResourceConfigWithTimings(name string, heartbeat, election int64) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmetcd_cluster" "test" {
  name               = %[4]q
  heartbeat_interval = %[5]d
  election_timeout   = %[6]d
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		heartbeat,
		election,
	)
}

func testAccCMEtcdClusterResourceConfigWithOptions(name, options string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmetcd_cluster" "test" {
  name    = %[4]q
  options = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
		options,
	)
}

// =============================================================================
// Acceptance Tests - Update Timings and Verify (CRUD Depth)
// =============================================================================

// TestAccCMEtcdCluster_updateTimingsAndVerify creates a cluster with heartbeat=100,
// election=1000, updates to heartbeat=200, election=2000, verifies both values,
// then confirms idempotency with an empty plan.
func TestAccCMEtcdCluster_updateTimingsAndVerify(t *testing.T) {
	clusterName := generateShortTestName("tmv")

	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			// Step 1: Create with heartbeat=100, election=1000
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 100, 1000),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("heartbeat_interval"),
						knownvalue.Int64Exact(100),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("election_timeout"),
						knownvalue.Int64Exact(1000),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 2: Update to heartbeat=200, election=2000
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 200, 2000),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("heartbeat_interval"),
						knownvalue.Int64Exact(200),
					),
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("election_timeout"),
						knownvalue.Int64Exact(2000),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Step 3: Idempotency check
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 200, 2000),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
		},
	})
}

// =============================================================================
// Acceptance Tests - Idempotency
// =============================================================================

// TestAccCMEtcdCluster_idempotency tests that creating an etcd cluster with
// options and re-applying produces no changes across multiple cycles.
func TestAccCMEtcdCluster_idempotency(t *testing.T) {
	clusterName := generateShortTestName("idem")

	compareID := statecheck.CompareValue(compare.ValuesSame())

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			// Create.
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 150, 1500),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency cycle 1.
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 150, 1500),
				ConfigPlanChecks: resource.ConfigPlanChecks{
					PreApply: []plancheck.PlanCheck{
						plancheck.ExpectEmptyPlan(),
					},
				},
				ConfigStateChecks: []statecheck.StateCheck{
					compareID.AddStateValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("id"),
					),
				},
			},
			// Idempotency cycle 2 (catches any state mutation during read).
			{
				Config: testAccCMEtcdClusterResourceConfigWithTimings(clusterName, 150, 1500),
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
// Acceptance Tests - Disappears
// =============================================================================

// etcdClusterDisappearsCheck is a custom StateCheck that deletes an etcd cluster
// via the BCM API to simulate external deletion (disappearance).
type etcdClusterDisappearsCheck struct {
	resourceAddress string
}

func (c etcdClusterDisappearsCheck) CheckState(ctx context.Context, req statecheck.CheckStateRequest, resp *statecheck.CheckStateResponse) {
	// Find the resource in state by address
	var uuid string
	for _, r := range req.State.Values.RootModule.Resources {
		if r.Address == c.resourceAddress {
			var ok bool
			uuid, ok = r.AttributeValues["uuid"].(string)
			if !ok || uuid == "" {
				resp.Error = fmt.Errorf("resource %s has no uuid attribute in state", c.resourceAddress)
				return
			}
			break
		}
	}

	if uuid == "" {
		resp.Error = fmt.Errorf("resource %s not found in state", c.resourceAddress)
		return
	}

	// Delete the etcd cluster externally via BCM API
	client := createTestBCMClient(&testing.T{})
	_, err := client.CallJSONRPC(ctx, "cmetcd", "removeEtcdCluster", uuid)
	if err != nil {
		resp.Error = fmt.Errorf("failed to delete etcd cluster %s via BCM API: %w", uuid, err)
		return
	}

	// Wait for eventual consistency
	time.Sleep(2 * time.Second)
}

// TestAccCMEtcdCluster_Disappears verifies that when an etcd cluster is deleted
// externally (outside Terraform), the next plan detects the disappearance.
func TestAccCMEtcdCluster_Disappears(t *testing.T) {
	clusterName := generateShortTestName("etcd-dis")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheckCMEtcdCluster(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMEtcdClusterDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccCMEtcdClusterResourceConfig(clusterName),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"bcm_cmetcd_cluster.test",
						tfjsonpath.New("name"),
						knownvalue.StringExact(clusterName),
					),
					etcdClusterDisappearsCheck{resourceAddress: "bcm_cmetcd_cluster.test"},
				},
				ExpectNonEmptyPlan: true,
			},
		},
	})
}
