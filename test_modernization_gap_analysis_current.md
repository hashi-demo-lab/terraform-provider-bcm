# Terraform Provider Test Modernization Gap Analysis
**Analysis Date:** 2025-11-23
**Test Directory:** `internal/provider`

## Executive Summary
- **250** modern state checks (`statecheck.ExpectKnownValue`)
- **39** modern plan checks (`plancheck.Expect*`)
- **0** legacy check calls (needs cleanup)
- **6/8** resources have drift detection tests
- **6/8** resources have import tests

**Overall Grade: C**

## Resource Tests Analysis

### resource_cmpart_softwareimage_test.go
- **Test functions:** 13
- **Modern state checks:** 38
- **Modern plan checks:** 8
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 6
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmpart_partition_test.go
- **Test functions:** 7
- **Modern state checks:** 23
- **Modern plan checks:** 4
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 3
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_category_test.go
- **Test functions:** 10
- **Modern state checks:** 28
- **Modern plan checks:** 5
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 4
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmnet_network_test.go
- **Test functions:** 4
- **Modern state checks:** 20
- **Modern plan checks:** 5
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 4
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_device_idempotency_test.go
- **Test functions:** 3
- **Modern state checks:** 11
- **Modern plan checks:** 5
- **Has import test:** ❌
- **Has drift test:** ❌
- **Idempotency checks:** 5
- **Legacy checks:** None ✅
- **Status:** ⚠️ Missing import test

### resource_cmkube_cluster_test.go
- **Test functions:** 11
- **Modern state checks:** 35
- **Modern plan checks:** 8
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 7
- **Legacy checks:** None ✅
- **Status:** ✅ Fully modernized

### resource_cmdevice_device_errors_test.go
- **Test functions:** 17
- **Modern state checks:** 0
- **Modern plan checks:** 0
- **Has import test:** ❌
- **Has drift test:** ❌
- **Idempotency checks:** 0
- **Legacy checks:** None ✅
- **Status:** ⚠️ Missing import test

### resource_cmdevice_device_test.go
- **Test functions:** 10
- **Modern state checks:** 14
- **Modern plan checks:** 4
- **Has import test:** ✅
- **Has drift test:** ✅
- **Idempotency checks:** 2
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

**Missing Drift Detection Tests:**
- `resource_cmdevice_device_idempotency_test.go`
- `resource_cmdevice_device_errors_test.go`

### MEDIUM PRIORITY 📋

**Missing Import Tests:**
- `resource_cmdevice_device_idempotency_test.go`
- `resource_cmdevice_device_errors_test.go`

**Missing Idempotency Checks:**
- `resource_cmdevice_device_errors_test.go`

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
