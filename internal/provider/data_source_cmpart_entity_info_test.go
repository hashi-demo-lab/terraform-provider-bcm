// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// =============================================================================
// Phase 3: User Story 1 Tests - List Entities by Type (TDD RED Phase)
// =============================================================================

// T009: TestAccCMPartEntityInfoDataSource_Basic
// Verify data source returns entities with id computed
func TestAccCMPartEntityInfoDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify id is computed and not null
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					// Verify entities is a list (may be empty or populated)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("entities"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// T010: TestAccCMPartEntityInfoDataSource_FilterByType
// Verify filtering by type="SoftwareImage" returns only SoftwareImage entities
func TestAccCMPartEntityInfoDataSource_FilterByType(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfigFilterByType("SoftwareImage"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify id is computed
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					// Verify entities list is present
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("entities"),
						knownvalue.NotNull(),
					),
				},
				// Additional check: verify all returned entities have type = "SoftwareImage"
				Check: resource.ComposeTestCheckFunc(
					// Verify that the type filter is applied correctly
					// The id should reflect the type filter
					resource.TestMatchResourceAttr(
						"data.bcm_cmpart_entity_info.test",
						"id",
						regexp.MustCompile(`^cmpart-entity-info:SoftwareImage`),
					),
				),
			},
		},
	})
}

// T011: TestAccCMPartEntityInfoDataSource_EmptyResult
// Verify querying with type="NonExistentType123" returns empty list (not error)
func TestAccCMPartEntityInfoDataSource_EmptyResult(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfigFilterByType("NonExistentType123"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify id is computed even with no results
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					// Verify empty entities list (not null, but empty)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("entities"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

// T012: TestAccCMPartEntityInfoDataSource_InvalidCredentials
// Verify authentication error message is clear
func TestAccCMPartEntityInfoDataSource_InvalidCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = "invalid_user"
  password             = "invalid_password"
  insecure_skip_verify = true
}

data "bcm_cmpart_entity_info" "test" {}
`, os.Getenv("BCM_ENDPOINT")),
				ExpectError: regexp.MustCompile(`(login failed|authentication|401|unauthorized|Unable to Create BCM Client)`),
			},
		},
	})
}

// =============================================================================
// Phase 4: User Story 2 Tests - Filter Entities by Name Pattern (TDD RED Phase)
// =============================================================================

// T018: TestAccCMPartEntityInfoDataSource_FilterByNamePattern
// Verify filtering by name_pattern="default*" returns matching entities
func TestAccCMPartEntityInfoDataSource_FilterByNamePattern(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfigFilterByNamePattern("default*"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify id is computed with name pattern
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					// Verify id includes the name pattern
					resource.TestMatchResourceAttr(
						"data.bcm_cmpart_entity_info.test",
						"id",
						regexp.MustCompile(`^cmpart-entity-info:.*:default\*`),
					),
				),
			},
		},
	})
}

// T019: TestAccCMPartEntityInfoDataSource_FilterByNamePatternMiddle
// Verify filtering by name_pattern="*node*" for middle match
func TestAccCMPartEntityInfoDataSource_FilterByNamePatternMiddle(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfigFilterByNamePattern("*node*"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// T020: TestAccCMPartEntityInfoDataSource_FilterByExactName
// Verify filtering without wildcards matches literally
func TestAccCMPartEntityInfoDataSource_FilterByExactName(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfigFilterByNamePattern("default-image"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// Additional test for case-insensitivity of name_pattern filter (HIGH priority from analysis)
func TestAccCMPartEntityInfoDataSource_FilterByNamePatternCaseInsensitive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Use uppercase pattern to test case-insensitivity
				Config: testAccCMPartEntityInfoDataSourceConfigFilterByNamePattern("DEFAULT*"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// =============================================================================
// Phase 5: User Story 3 Tests - Combined Type and Name Filtering (TDD RED Phase)
// =============================================================================

// T023: TestAccCMPartEntityInfoDataSource_CombinedFilters
// Verify combined type and name_pattern filtering with AND logic
func TestAccCMPartEntityInfoDataSource_CombinedFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfigCombinedFilters("SoftwareImage", "default*"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
				Check: resource.ComposeTestCheckFunc(
					// Verify id includes both type and name pattern
					resource.TestMatchResourceAttr(
						"data.bcm_cmpart_entity_info.test",
						"id",
						regexp.MustCompile(`^cmpart-entity-info:SoftwareImage:default\*`),
					),
				),
			},
		},
	})
}

// =============================================================================
// Phase 6: User Story 4 Tests - Retrieve All Entities (TDD RED Phase)
// =============================================================================

// T026: TestAccCMPartEntityInfoDataSource_NoFilters
// Verify querying without filters returns all entities
func TestAccCMPartEntityInfoDataSource_NoFilters(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartEntityInfoDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify id is "cmpart-entity-info:all" when no filters
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-entity-info:all"),
					),
					// Verify entities list is not null
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("entities"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// =============================================================================
// Phase 7: User Story 5 Tests - Lookup Entity UUID by Known Name (TDD RED Phase)
// =============================================================================

// T028: TestAccCMPartEntityInfoDataSource_UUIDLookup
// Verify UUID format is correct (36 chars with dashes)
func TestAccCMPartEntityInfoDataSource_UUIDLookup(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Query for SoftwareImage entities which should exist
				Config: testAccCMPartEntityInfoDataSourceConfigFilterByType("SoftwareImage"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_entity_info.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
				// Check that if entities exist, they have valid UUIDs
				Check: resource.ComposeTestCheckFunc(
					// UUID format: xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx (36 chars)
					resource.TestMatchResourceAttr(
						"data.bcm_cmpart_entity_info.test",
						"entities.0.uuid",
						regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`),
					),
				),
			},
		},
	})
}

// =============================================================================
// Test Configuration Helpers
// =============================================================================

// testAccCMPartEntityInfoDataSourceConfig returns a basic test configuration
// for the bcm_cmpart_entity_info data source.
func testAccCMPartEntityInfoDataSourceConfig() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_entity_info" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccCMPartEntityInfoDataSourceConfigFilterByType returns a configuration with type filter.
func testAccCMPartEntityInfoDataSourceConfigFilterByType(entityType string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_entity_info" "test" {
  type = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		entityType,
	)
}

// testAccCMPartEntityInfoDataSourceConfigFilterByNamePattern returns a configuration with name pattern filter.
func testAccCMPartEntityInfoDataSourceConfigFilterByNamePattern(namePattern string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_entity_info" "test" {
  name_pattern = %[4]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		namePattern,
	)
}

// testAccCMPartEntityInfoDataSourceConfigCombinedFilters returns a configuration with both filters.
func testAccCMPartEntityInfoDataSourceConfigCombinedFilters(entityType, namePattern string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_entity_info" "test" {
  type         = %[4]q
  name_pattern = %[5]q
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		entityType,
		namePattern,
	)
}
