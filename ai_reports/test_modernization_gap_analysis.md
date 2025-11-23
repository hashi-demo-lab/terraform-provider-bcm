# Terraform Provider Test Modernization Gap Analysis
**Analysis Date:** 2025-11-23
**Test Directory:** `internal/provider`

## Executive Summary
- **256** modern state checks (`statecheck.ExpectKnownValue`)
- **42** modern plan checks (`plancheck.Expect*`)
- **0** legacy check calls (needs cleanup)
- **7/7** acceptance test resources have drift detection tests
- **7/7** acceptance test resources have import tests
- **11/11** required fields have validation tests
- **50/111** optional fields are tested

**CRUD Coverage:**
- Create: 7/7
- Update: 7/7
- Delete: 7/7

**Average Quality Score:** 100/100

**1** mock/unit test files (import/drift N/A)

**Overall Grade: A**

## Resource Tests Analysis

### resource_cmpart_softwareimage_test.go
- **Test functions:** 14
- **Modern state checks:** 38
- **Modern plan checks:** 8
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 6
- **Validation tests:** 4
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmpart_partition_test.go
- **Test functions:** 8
- **Modern state checks:** 23
- **Modern plan checks:** 4
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 3
- **Validation tests:** 2
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_category_test.go
- **Test functions:** 10
- **Modern state checks:** 28
- **Modern plan checks:** 5
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 4
- **Validation tests:** 4
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_device_mock_test.go
- **Test functions:** 17
- **Modern state checks:** 0
- **Modern plan checks:** 0
- **Test type:** Mock/Unit tests (import/drift N/A)
- **Idempotency checks:** 0
- **Validation tests:** 17
- **CRUD coverage:** Create, Update
- **Quality score:** 40/100
- **Legacy checks:** None ✅
- **Status:** ✅ Mock/unit tests (error validation)

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
- **Quality score:** 100/100
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmkube_cluster_test.go
- **Test functions:** 11
- **Modern state checks:** 35
- **Modern plan checks:** 8
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 7
- **Validation tests:** 2
- **CRUD coverage:** Create, Update, Delete
- **Quality score:** 100/100
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

### data_source_cmpart_softwareimages_test.go
- **Test functions:** 7
- **Modern state checks:** 6
- **Modern plan checks:** 0
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

## Gaps and Recommendations

### HIGH PRIORITY ⚠️

No high priority issues found! ✅

### MEDIUM PRIORITY 📋

**Untested Optional Fields:**
- **`cmdevice_category`**: revision_id, boot_loader_file, boot_loader_protocol, kernel_version, kernel_output_console (and 46 more)
- **`cmkube_cluster`**: force
- **`cmpart_partition`**: relay_host, no_zero_conf
- **`cmpart_softwareimage`**: kernel_output_console, sol_port, sol_flow_control, bootfspart, revision_id (and 2 more)

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
        "bcm_resource.test",
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
compareID := statecheck.CompareValue(compare.ValuesSame())

ConfigStateChecks: []statecheck.StateCheck{
    compareID.AddStateValue("bcm_resource.test", tfjsonpath.New("id")),
}
```
