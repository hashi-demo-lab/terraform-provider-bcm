# Terraform Provider Test Modernization Gap Analysis
**Analysis Date:** 2025-11-25
**Test Directory:** `internal/provider`

## Codebase Statistics
**Total Go Files:** 42 files | 25,334 lines
- Code: 19,409 lines (76%)
- Comments: 3,401 lines (13%)
- Blank: 2,524 lines (9%)

**By File Type:**
| Type | Files | Lines | Code Lines |
|------|------:|------:|-----------:|
| Test | 24 | 14,859 | 11,573 |
| Resource | 5 | 5,784 | 4,503 |
| Data Source | 8 | 3,269 | 2,603 |
| Other | 5 | 1,422 | 730 |
| **Total** | **42** | **25,334** | **19,409** |

**Test-to-Implementation Ratio:** 1.48:1 (11,573 test lines / 7,836 impl lines)

## Executive Summary
- **325** modern state checks (`statecheck.ExpectKnownValue`)
- **64** modern plan checks (`plancheck.Expect*`)
- **0** legacy check calls (needs cleanup)
- **6/6** acceptance test resources have drift detection tests
- **6/6** acceptance test resources have import tests
- **9/9** required fields have validation tests
- **71/99** optional fields are tested

**CRUD Coverage:**
- Create: 6/6
- Update: 6/6
- Delete: 6/6

**Cleanup Analysis:**
- **6/6** resources have robust cleanup verification
- **No cleanup issues detected** ✅

**Naming Uniqueness:**
- **All tests use unique name generation** ✅

**ID Consistency Tracking:**
- **5/6** resources use `CompareValue(ValuesSame())` for ID tracking
- **6/6** resources have ID consistency issues (7 total) ⚠️

**Average Quality Score:** 98/100

**2** mock/unit test files (1 error-only, import/drift N/A)

**Overall Grade: A**

## Resource Tests Analysis

### resource_cmpart_softwareimage_test.go
- **Test functions:** 20
- **Modern state checks:** 57
- **Modern plan checks:** 13
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 11
- **Validation tests:** 4
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Cleanup:** ✅ Robust
- **ID tracking:** ✅ Uses CompareValue (2/40 steps)
- **ID consistency issues:** 1 ⚠️
  - Partial ID tracking: 2/40 steps track ID
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmkube_cluster_mock_test.go
- **Test functions:** 4
- **Modern state checks:** 7
- **Modern plan checks:** 3
- **Test type:** Mock/Unit tests (import/drift N/A)
- **Idempotency checks:** 3
- **Validation tests:** 1
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 85/100
- **Cleanup:** ✅ Robust
- **ID tracking:** ✅ Uses CompareValue (8/9 steps)
- **Legacy checks:** None ✅
- **Status:** ✅ Mock/unit tests (error validation)

### resource_cmdevice_category_test.go
- **Test functions:** 18
- **Modern state checks:** 83
- **Modern plan checks:** 21
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 20
- **Validation tests:** 4
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Cleanup:** ✅ Robust
- **ID tracking:** ✅ Uses CompareValue (15/51 steps)
- **ID consistency issues:** 1 ⚠️
  - Partial ID tracking: 15/51 steps track ID
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_device_mock_test.go
- **Test functions:** 17
- **Modern state checks:** 0
- **Modern plan checks:** 0
- **Test type:** Error-only tests (uses httptest mock servers)
- **Idempotency checks:** 0
- **Validation tests:** 17
- **CRUD coverage:** Create, Update
- **Quality score:** 90/100
- **Legacy checks:** None ✅
- **Status:** ✅ Error path coverage (httptest mocks)

### resource_cmnet_network_test.go
- **Test functions:** 4
- **Modern state checks:** 20
- **Modern plan checks:** 5
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 4
- **Validation tests:** 0
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Cleanup:** ✅ Robust
- **ID tracking:** ✅ Uses CompareValue (2/11 steps)
- **ID consistency issues:** 1 ⚠️
  - Partial ID tracking: 2/11 steps track ID
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_device_idempotency_test.go
- **Test functions:** 5
- **Modern state checks:** 17
- **Modern plan checks:** 8
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 7
- **Validation tests:** 0
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 90/100
- **Cleanup:** ✅ Robust
- **ID tracking:** ❌ Missing CompareValue for ID consistency
- **ID consistency issues:** 2 ⚠️
  - Multiple test steps (14) without ID consistency tracking - add CompareValue(ValuesSame())
  - Uses 4 ExpectKnownValue for ID but no CompareValue consistency tracking
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmkube_cluster_test.go
- **Test functions:** 13
- **Modern state checks:** 41
- **Modern plan checks:** 10
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 9
- **Validation tests:** 2
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Cleanup:** ✅ Robust
- **ID tracking:** ✅ Uses CompareValue (2/30 steps)
- **ID consistency issues:** 1 ⚠️
  - Partial ID tracking: 2/30 steps track ID
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_device_test.go
- **Test functions:** 10
- **Modern state checks:** 14
- **Modern plan checks:** 4
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 2
- **Validation tests:** 4
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Cleanup:** ✅ Robust
- **ID tracking:** ✅ Uses CompareValue (3/15 steps)
- **ID consistency issues:** 1 ⚠️
  - Partial ID tracking: 3/15 steps track ID
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

## Data Source Tests Analysis

### data_source_cmpart_partitions_test.go
- **Test functions:** 9
- **Modern state checks:** 26
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### data_source_cmnet_networks_test.go
- **Test functions:** 4
- **Modern state checks:** 12
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### data_source_cmdevice_nodes_test.go
- **Test functions:** 4
- **Modern state checks:** 10
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### data_source_cmuser_users_test.go
- **Test functions:** 6
- **Modern state checks:** 15
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### data_source_cmkube_clusters_test.go
- **Test functions:** 7
- **Modern state checks:** 7
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### data_source_cmdevice_categories_test.go
- **Test functions:** 4
- **Modern state checks:** 5
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### data_source_cmdevice_roles_test.go
- **Test functions:** 5
- **Modern state checks:** 5
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### data_source_cmpart_softwareimages_test.go
- **Test functions:** 7
- **Modern state checks:** 6
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

## Gaps and Recommendations

### HIGH PRIORITY ⚠️

**ID Consistency Issues:**
These tests have inconsistent ID property usage:

- **`resource_cmdevice_device_idempotency_test.go`** (2 issues):
  - Multiple test steps (14) without ID consistency tracking - add CompareValue(ValuesSame())
  - Uses 4 ExpectKnownValue for ID but no CompareValue consistency tracking
- **`resource_cmpart_softwareimage_test.go`** (1 issues):
  - Partial ID tracking: 2/40 steps track ID
- **`resource_cmdevice_category_test.go`** (1 issues):
  - Partial ID tracking: 15/51 steps track ID
- **`resource_cmnet_network_test.go`** (1 issues):
  - Partial ID tracking: 2/11 steps track ID
- **`resource_cmkube_cluster_test.go`** (1 issues):
  - Partial ID tracking: 2/30 steps track ID
- **`resource_cmdevice_device_test.go`** (1 issues):
  - Partial ID tracking: 3/15 steps track ID


### MEDIUM PRIORITY 📋

**Untested Optional Fields:**
- **`cmdevice_category`**: io_scheduler, static_routes, fsmounts, fsexports, roles (and 23 more)

## Modern Testing Patterns Quick Reference

### Required Imports
```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/compare"
)
```

### Modern State Check
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "example_resource.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact("expected-value"),
    ),
}
```

### Idempotency Check
```go
ConfigPlanChecks: resource.ConfigPlanChecks{
    PreApply: []plancheck.PlanCheck{
        plancheck.ExpectEmptyPlan(),
    },
}
```

### ID Consistency Tracking
```go
// Initialize ID tracker at test start (before Steps)
compareID := statecheck.CompareValue(compare.ValuesSame())

// Step 1: Create - track ID
{
    Config: testAccResourceConfig(name),
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("example_resource.test", tfjsonpath.New("id")),
    },
}

// Step 2: Import - track ID to verify consistency
{
    ResourceName:      "example_resource.test",
    ImportState:       true,
    ImportStateVerify: true,
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("example_resource.test", tfjsonpath.New("id")),
    },
}

// Step 3: Update - track ID to ensure stability
{
    Config: testAccResourceConfig(name, "updated"),
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("example_resource.test", tfjsonpath.New("id")),
    },
}
// CompareValue(ValuesSame()) ensures ID remains identical across all steps
```

### Robust CheckDestroy Pattern
```go
func testAccCheckResourceDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "example_resource" {
            continue
        }

        // Verify resource deleted with retry logic
        deleted, err := verifyResourceDeleted(
            context.Background(),
            client,
            "Service",
            "getMethod",
            rs.Primary.ID,
            4, // retry count
        )

        if !deleted || err != nil {
            return fmt.Errorf("resource still exists after destroy: %s", rs.Primary.ID)
        }
    }

    return nil
}
```

### Cleanup-Friendly Naming
```go
// Use unique names for easy cleanup (recommended: tftest- prefix)
resourceName := generateUniqueTestName("tftest-resource")

// Alternative: citest prefix for CI/CD examples
resourceName := generateUniqueTestName("citest-resource")

// Manual timestamp-based naming
resourceName := fmt.Sprintf("tftest-%s-%d", "resource", time.Now().Unix())
```
