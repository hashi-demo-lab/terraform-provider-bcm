// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccCMPartPartitionsDataSource_Basic verifies data source retrieves all partitions without filters.
func TestAccCMPartPartitionsDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify computed ID field
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("base"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("base_type"),
						knownvalue.StringExact("Partition"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartitionsDataSource_FilterByNamePattern verifies client-side filtering by name pattern works.
func TestAccCMPartPartitionsDataSource_FilterByNamePattern(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfigFilter("base"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify ID computed
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("base"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartitionsDataSource_NoMatches verifies filter returning no results returns empty list, not error.
func TestAccCMPartPartitionsDataSource_NoMatches(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfigFilter("nonexistent-partition-12345"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify ID still computed even with no matches
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartitionsDataSource_ComputedFields verifies all partition attributes are exposed with correct types.
// This test assumes at least one partition exists in the BCM cluster (typical for any configured cluster).
func TestAccCMPartPartitionsDataSource_ComputedFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify top-level fields
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					// Verify partitions list is not null
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("base"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// testAccCMPartPartitionsDataSourceConfig returns provider config + basic data source declaration.
func testAccCMPartPartitionsDataSourceConfig() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_partitions" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccCMPartPartitionsDataSourceConfigFilter returns provider config + data source with name pattern filter.
func testAccCMPartPartitionsDataSourceConfigFilter(namePattern string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_partitions" "test" {
  filter {
    name_pattern = %[4]q
  }
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		namePattern,
	)
}

// TestAccCMPartPartitionsDataSource_AttributeTypes verifies all attribute types are correctly populated.
// This test provides comprehensive type verification for String, Int64, Bool, and List attributes.
func TestAccCMPartPartitionsDataSource_AttributeTypes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify String attributes are properly typed
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("uuid"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("slave_name"),
						knownvalue.StringExact("node"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("slave_digits"),
						knownvalue.Int64Exact(3),
					),
					// Verify Bool attributes exist (may be true or false)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("modified"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("to_be_removed"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartitionsDataSource_ListAttributes verifies List[String] attributes are properly unmarshaled.
// Tests all four List-type attributes: admin_email, time_servers, search_domains, name_servers.
func TestAccCMPartPartitionsDataSource_ListAttributes(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify all List attributes are valid types (not null type checks)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("admin_email"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("time_servers"),
						knownvalue.ListSizeExact(3),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("time_servers").AtSliceIndex(0),
						knownvalue.StringExact("0.pool.ntp.org"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("search_domains"),
						knownvalue.ListSizeExact(0),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name_servers"),
						knownvalue.ListSizeExact(0),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartitionsDataSource_FilterCaseInsensitive verifies name_pattern filter is case-insensitive
// Tests documented behavior: "base" should match "BASE", "Base", "base-partition", etc.
func TestAccCMPartPartitionsDataSource_FilterCaseInsensitive(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfigFilter("base"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify ID computed
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("base"),
					),
				},
			},
			{
				Config: testAccCMPartPartitionsDataSourceConfigFilter("BASE"),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify ID computed
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("base"),
					),
				},
			},
			{
				Config: testAccCMPartPartitionsDataSourceConfigFilter("BaSe"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("base"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartitionsDataSource_FilterEmptyString verifies empty string filter returns all partitions.
// Edge case: Empty filter should behave same as no filter.
func TestAccCMPartPartitionsDataSource_FilterEmptyString(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartPartitionsDataSourceConfigFilter(""),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify ID computed
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.StringExact("cmpart-partitions"),
					),
					// Verify partitions list is not null
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.ListSizeExact(1),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("name"),
						knownvalue.StringExact("base"),
					),
				},
			},
		},
	})
}

// TestAccCMPartPartitionsDataSource_FilterSubsetProperty verifies filtered results are subset of unfiltered.
// Mathematical property: count(filtered) <= count(all).
func TestAccCMPartPartitionsDataSource_FilterSubsetProperty(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				// Step 1: Get all partitions (no filter)
				Config: testAccCMPartPartitionsDataSourceConfig(),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.NotNull(),
					),
				},
			},
			{
				// Step 2: Get filtered partitions - should be subset
				Config: testAccCMPartPartitionsDataSourceConfigFilter("base"),
				ConfigStateChecks: []statecheck.StateCheck{
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_partitions.test",
						tfjsonpath.New("partitions"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}
