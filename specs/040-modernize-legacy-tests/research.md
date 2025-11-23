# Research: Legacy Testing Pattern Analysis

**Feature**: Modernize Legacy Testing Patterns (Issue #40)
**Date**: 2025-11-23
**Analyst**: Claude Code

## Executive Summary

- **Total Legacy Assertions Identified**: 66 across 3 test files
  - resource_cmnet_network_test.go: 18 legacy assertions
  - resource_cmkube_cluster_test.go: 38 legacy assertions
  - data_source_cmpart_partitions_test.go: 10 legacy assertions
- **Transformation Categories**: 4 types (String, Int64, Bool, Computed)
- **New Test Steps Required**: ~6 idempotency checks
- **Import Additions**: 4 per file (statecheck, knownvalue, tfjsonpath, compare)

---

## 1. Legacy Pattern Inventory

### 1.1 resource_cmnet_network_test.go (18 assertions)

**TestAccCMNetNetwork_Basic** (lines 27-34) - 4 assertions:
- Line 30: `TestCheckResourceAttr("name", networkName)` → String
- Line 31: `TestCheckResourceAttrSet("id")` → Computed
- Line 32: `TestCheckResourceAttrSet("uuid")` → Computed
- Line 33: `TestCheckResourceAttr("domain_name", "cluster.local")` → String

**TestAccCMNetNetwork_Complete** (lines 67-77) - 10 assertions:
- Line 68: `TestCheckResourceAttr("name", networkName)` → String
- Line 69: `TestCheckResourceAttr("subnet", "192.168.100.0/24")` → String
- Line 70: `TestCheckResourceAttr("gateway", "192.168.100.1")` → String
- Line 71: `TestCheckResourceAttr("mtu", "9000")` → Int64
- Line 72: `TestCheckResourceAttr("domain_name", "test.local")` → String
- Line 73: `TestCheckResourceAttr("dhcp_range_start", "192.168.100.100")` → String
- Line 74: `TestCheckResourceAttr("dhcp_range_end", "192.168.100.200")` → String
- Line 75: `TestCheckResourceAttr("dhcp_enabled", "true")` → Bool
- Line 76: `TestCheckResourceAttr("notes", "Test network")` → String
- Line 77: `TestCheckResourceAttrSet("uuid")` → Computed

**TestAccCMNetNetwork_Update** (lines 95-96, 103-105) - 4 assertions:
- Line 96: `TestCheckResourceAttr("name", networkName)` → String (Create step)
- Line 103: `TestCheckResourceAttr("name", networkName)` → String (Update step)
- Line 104: `TestCheckResourceAttr("mtu", "9000")` → Int64
- Line 105: `TestCheckResourceAttr("notes", "Updated notes")` → String

**Summary**: 18 total assertions (14 String, 2 Int64, 1 Bool, 1 Computed + 2 implicit Computed from TestCheckResourceAttrSet)

---

### 1.2 resource_cmkube_cluster_test.go (38 assertions)

**IMPORTANT**: This file has MIXED patterns - some functions have both legacy Check blocks AND modern ConfigStateChecks. The legacy Check blocks are redundant and must be removed.

**TestAccCMKubeClusterResource_Basic** (lines 178-182, 228) - 4 assertions:
- Line 179: `TestCheckResourceAttr("name", clusterName)` → String
- Line 180: `TestCheckResourceAttrSet("uuid")` → Computed
- Line 181: `TestCheckResourceAttrSet("id")` → Computed
- Line 228: `TestCheckResourceAttr("name", clusterNameUpdated)` → String

**TestAccCMKubeClusterResource_DriftDetection** (lines 270, 340) - 2 assertions:
- Line 270: `TestCheckResourceAttr("version", "1.28.0")` → String
- Line 340: `TestCheckResourceAttr("version", "1.28.0")` → String

**TestAccCMKubeClusterResource_WorkerNodes** (lines 374, 392, 410) - 3 assertions:
- Line 374: `TestCheckResourceAttr("worker_nodes.#", "1")` → Int64/List count
- Line 392: `TestCheckResourceAttr("worker_nodes.#", "2")` → Int64/List count
- Line 410: `TestCheckResourceAttr("worker_nodes.#", "0")` → Int64/List count

**TestAccCMKubeClusterResource_CompleteConfiguration** (lines 486-492, 546-547) - 7 assertions:
- Line 487: `TestCheckResourceAttr("name", clusterName)` → String
- Line 488: `TestCheckResourceAttr("master_nodes.#", "1")` → Int64/List count
- Line 489: `TestCheckResourceAttr("worker_nodes.#", "1")` → Int64/List count
- Line 490: `TestCheckResourceAttr("management_network", managementNetworkUUID)` → String
- Line 491: `TestCheckResourceAttr("version", "1.28.0")` → String
- Line 546: `TestCheckResourceAttr("name", clusterNameUpdated)` → String
- Line 547: `TestCheckResourceAttr("version", "1.29.0")` → String

**TestAccCMKubeClusterResource_P3AdvancedNetworking** (lines 688-690, 723-724) - 5 assertions:
- Line 688: `TestCheckResourceAttr("name", clusterName)` → String
- Line 689: `TestCheckResourceAttr("cni_plugin", "calico")` → String
- Line 690: `TestCheckResourceAttr("dns_servers.#", "2")` → Int64/List count
- Line 723: `TestCheckResourceAttr("cni_plugin", "flannel")` → String
- Line 724: `TestCheckResourceAttr("dns_servers.#", "1")` → Int64/List count

**TestAccCMKubeClusterResource_P3StorageAndLoadBalancer** (lines 757-759, 787) - 4 assertions:
- Line 757: `TestCheckResourceAttr("name", clusterName)` → String
- Line 758: `TestCheckResourceAttr("load_balancer_mode", "metallb")` → String
- Line 759: `TestCheckResourceAttrSet("storage_classes")` → Computed
- Line 787: `TestCheckResourceAttr("load_balancer_mode", "haproxy")` → String

**TestAccCMKubeClusterResource_P3Addons** (lines 815-817, 845-846) - 5 assertions:
- Line 815: `TestCheckResourceAttr("name", clusterName)` → String
- Line 816: `TestCheckResourceAttrSet("addons")` → Computed
- Line 817: `TestCheckResourceAttrSet("ingress_controller")` → Computed
- Line 845: `TestCheckResourceAttrSet("addons")` → Computed
- Line 846: `TestCheckResourceAttrSet("ingress_controller")` → Computed

**TestAccCMKubeClusterResource_P3FullStack** (lines 880-887) - 8 assertions:
- Line 880: `TestCheckResourceAttr("name", clusterName)` → String
- Line 881: `TestCheckResourceAttr("cni_plugin", "calico")` → String
- Line 882: `TestCheckResourceAttr("dns_servers.#", "2")` → Int64/List count
- Line 883: `TestCheckResourceAttr("load_balancer_mode", "metallb")` → String
- Line 884: `TestCheckResourceAttrSet("storage_classes")` → Computed
- Line 885: `TestCheckResourceAttrSet("addons")` → Computed
- Line 886: `TestCheckResourceAttrSet("ingress_controller")` → Computed
- Line 887: `TestCheckResourceAttrSet("overlay_network")` → Computed

**Summary**: 38 total assertions (19 String, 7 Int64/List count, 12 Computed)

**CRITICAL FINDING**: Functions TestAccCMKubeClusterResource_Basic, DriftDetection, WorkerNodes, and CompleteConfiguration have BOTH legacy Check blocks AND modern ConfigStateChecks. The plan.md correctly identified these need the legacy blocks removed (not converted), since modern versions already exist.

---

### 1.3 data_source_cmpart_partitions_test.go (10 assertions)

**TestAccCMPartPartitionsDataSource_ComputedFields** (line 95) - 1 assertion:
- Line 95: `TestCheckResourceAttrSet("partitions.#")` → List/Computed count

**TestAccCMPartPartitionsDataSource_AttributeTypes** (lines 168-174) - 5 assertions:
- Line 168: `TestCheckResourceAttrSet("partitions.#")` → List/Computed count
- Line 170: `TestCheckResourceAttrSet("partitions.0.id")` → Computed
- Line 171: `TestCheckResourceAttrSet("partitions.0.uuid")` → Computed
- Line 172: `TestCheckResourceAttrSet("partitions.0.name")` → Computed
- Line 174: `TestCheckResourceAttrSet("partitions.0.base_type")` → Computed

**TestAccCMPartPartitionsDataSource_ListAttributes** (line 228) - 1 assertion:
- Line 228: `TestCheckResourceAttrSet("partitions.#")` → List/Computed count

**TestAccCMPartPartitionsDataSource_FilterEmptyString** (line 335) - 1 assertion:
- Line 335: `TestCheckResourceAttrSet("partitions.#")` → List/Computed count

**TestAccCMPartPartitionsDataSource_FilterSubsetProperty** (lines 368, 389) - 2 assertions:
- Line 368: `TestCheckResourceAttrSet("partitions.#")` → List/Computed count
- Line 389: `TestCheckResourceAttrSet("partitions.#")` → List/Computed count

**Summary**: 10 total assertions (all TestCheckResourceAttrSet - Computed/List attributes)

**CRITICAL FINDING**: All legacy assertions in this file use TestCheckResourceAttrSet for existence checks. All functions already have modern ConfigStateChecks with NotNull() patterns. The legacy Check blocks are redundant and should be removed (not converted).

---

## 2. Modern Pattern Mappings

### 2.1 Transformation Table

| Legacy Pattern | Modern Pattern | Type Matcher | Notes |
|----------------|----------------|--------------|-------|
| `TestCheckResourceAttr("resource", "name", "value")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("name"), knownvalue.StringExact("value"))` | StringExact | For string attributes |
| `TestCheckResourceAttr("resource", "mtu", "9000")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("mtu"), knownvalue.Int64Exact(9000))` | Int64Exact | For numeric attributes (no quotes) |
| `TestCheckResourceAttr("resource", "dhcp_enabled", "true")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("dhcp_enabled"), knownvalue.Bool(true))` | Bool | For boolean attributes (no quotes) |
| `TestCheckResourceAttrSet("resource", "uuid")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("uuid"), knownvalue.NotNull())` | NotNull | For computed fields |
| `TestCheckResourceAttr("resource", "list.#", "2")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("list"), knownvalue.ListSizeExact(2))` | ListSizeExact | For known list counts |
| `TestCheckResourceAttrSet("resource", "list.#")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("list"), knownvalue.NotNull())` | NotNull | For environment-dependent lists |

### 2.2 Key Differences

1. **Check vs ConfigStateChecks**: Legacy uses `Check:` field, modern uses `ConfigStateChecks:` field
2. **ComposeAggregateTestCheckFunc**: Legacy wraps checks in this function, modern uses slice of StateCheck
3. **Type Safety**: Modern patterns use typed matchers (Int64Exact, Bool), legacy uses string comparisons
4. **Path Specification**: Modern uses tfjsonpath.New() for attribute paths
5. **List Indexing**: Modern uses .AtSliceIndex(0).AtMapKey("field") for nested paths

### 2.3 Pattern Classification

**Pattern Type 1: Direct String Replacement** (19 occurrences)
- Simple string attributes (name, notes, domain_name, subnet, gateway, etc.)
- Transform: `TestCheckResourceAttr(res, attr, val)` → `ExpectKnownValue(res, tfjsonpath.New(attr), StringExact(val))`

**Pattern Type 2: Numeric Type Conversion** (9 occurrences)
- Numeric attributes stored as strings in legacy (mtu "9000", list counts "1", "2")
- Transform: Remove quotes, use Int64Exact(9000) or ListSizeExact(1)

**Pattern Type 3: Boolean Type Conversion** (1 occurrence)
- Boolean attributes stored as strings in legacy (dhcp_enabled "true")
- Transform: Remove quotes, use Bool(true)

**Pattern Type 4: Computed Field Existence** (15+ occurrences)
- TestCheckResourceAttrSet for existence checks
- Transform: Use NotNull() matcher

---

## 3. Import Requirements

### 3.1 Current Imports (Per File Analysis)

**resource_cmnet_network_test.go**:
```go
// Current imports:
"github.com/hashicorp/terraform-plugin-testing/helper/resource"
"github.com/hashicorp/terraform-plugin-testing/plancheck"
"github.com/hashicorp/terraform-plugin-testing/terraform"

// Need to add:
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
"github.com/hashicorp/terraform-plugin-testing/compare"
```

**resource_cmkube_cluster_test.go**:
```go
// Current imports (ALREADY HAS ALL MODERN IMPORTS):
"github.com/hashicorp/terraform-plugin-testing/compare"        // ✅ Present
"github.com/hashicorp/terraform-plugin-testing/helper/resource"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"     // ✅ Present
"github.com/hashicorp/terraform-plugin-testing/plancheck"
"github.com/hashicorp/terraform-plugin-testing/statecheck"     // ✅ Present
"github.com/hashicorp/terraform-plugin-testing/terraform"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"     // ✅ Present

// NO IMPORTS NEEDED - already complete!
```

**data_source_cmpart_partitions_test.go**:
```go
// Current imports:
"github.com/hashicorp/terraform-plugin-testing/helper/resource"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"     // ✅ Present
"github.com/hashicorp/terraform-plugin-testing/statecheck"     // ✅ Present
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"     // ✅ Present

// Need to add (for resources only, N/A for data sources):
// "github.com/hashicorp/terraform-plugin-testing/compare" - only needed for resource ID tracking
```

### 3.2 Import Summary

- **resource_cmnet_network_test.go**: Needs 4 imports added
- **resource_cmkube_cluster_test.go**: Needs 0 imports (already complete)
- **data_source_cmpart_partitions_test.go**: Needs 0 imports (data sources don't use compare)

---

## 4. Missing Check Identification

### 4.1 Idempotency Gaps

**resource_cmnet_network_test.go**:
- ✅ TestAccCMNetNetwork_Basic: Already has idempotency (lines 37-44)
- ❌ TestAccCMNetNetwork_Complete: Missing idempotency after Create
- ❌ TestAccCMNetNetwork_Update: Missing idempotency after Create (line 98) and after Update (line 107)

**resource_cmkube_cluster_test.go**:
- ✅ All test functions already have comprehensive idempotency coverage
- No gaps identified

**data_source_cmpart_partitions_test.go**:
- N/A - Data sources are read-only, don't need idempotency checks

**Summary**: 3 idempotency steps need to be added (2 in TestAccCMNetNetwork_Update, 1 in TestAccCMNetNetwork_Complete)

### 4.2 ID Tracking Gaps

**resource_cmnet_network_test.go**:
- ❌ TestAccCMNetNetwork_Basic: Has Import step but NO ID tracking (compareID pattern missing)
- TestAccCMNetNetwork_Complete: No Import step, no ID tracking needed
- TestAccCMNetNetwork_Update: No Import step, no ID tracking needed

**resource_cmkube_cluster_test.go**:
- ✅ TestAccCMKubeClusterResource_Basic: Already has ID tracking (lines 168, 199-202, 237-240)
- Other test functions don't have Import steps, no ID tracking needed

**data_source_cmpart_partitions_test.go**:
- N/A - Data sources don't have CRUD operations or Import

**Summary**: 1 ID tracking pattern needs to be added (TestAccCMNetNetwork_Basic)

### 4.3 CheckDestroy Verification

**resource_cmnet_network_test.go**:
- Function: testAccCheckCMNetNetworkDestroy (lines 112-151)
- ✅ Matches CLAUDE.md enhanced pattern with:
  - Detailed error messages
  - Resource count tracking
  - Exponential backoff via verifyResourceDeleted helper
  - Error aggregation with formatted output
- **Status**: COMPLIANT - no changes needed

**resource_cmkube_cluster_test.go**:
- Function: testAccCheckCMKubeClusterDestroy (lines 39-77)
- ✅ Matches CLAUDE.md enhanced pattern with:
  - Detailed error messages
  - Resource count tracking
  - Exponential backoff via verifyResourceDeleted helper
  - Error aggregation with formatted output
- **Status**: COMPLIANT - no changes needed

**data_source_cmpart_partitions_test.go**:
- No CheckDestroy function (data sources don't create resources)
- **Status**: N/A

---

## 5. Edge Cases & Special Handling

### 5.1 Numeric Type Ambiguity

**Issue**: Terraform framework uses int64 internally, but BCM API may return float64
**Solution**: Empirical testing approach
- Start with Int64Exact(9000) for all numeric attributes
- If type errors occur, switch to Float64Exact(9000.0)
- Document decisions in implementation tasks

**Affected Attributes**:
- `mtu`: Start with Int64Exact(9000)
- List counts (`worker_nodes.#`, `dns_servers.#`): Use ListSizeExact(n) not Int64Exact

### 5.2 List Attribute Handling

**Mixed Strategy Decision** (from spec.md clarifications):
- **Known counts**: Use ListSizeExact(n) for test-created resources
  - Example: `worker_nodes` in TestAccCMKubeClusterResource_WorkerNodes
- **Unknown counts**: Use NotNull() for environment-dependent data
  - Example: `partitions` in data source tests

**Transformation Rules**:
- `TestCheckResourceAttr("list.#", "2")` → `knownvalue.ListSizeExact(2)`
- `TestCheckResourceAttrSet("list.#")` → `knownvalue.NotNull()`

### 5.3 Environment Portability

**Anti-Patterns to Avoid**:
- ❌ Hardcoded partition names in data source tests
- ❌ Specific partition counts that vary by cluster
- ❌ Assumptions about BCM cluster state

**Correct Patterns**:
- ✅ Use NotNull() for data sources with unknown counts
- ✅ Use ListSizeExact(n) only for resources created in the test
- ✅ Use knownvalue.NotNull() for computed fields

### 5.4 Redundant Check Block Pattern (CRITICAL)

**Issue Discovered**: Many test functions in resource_cmkube_cluster_test.go and data_source_cmpart_partitions_test.go have BOTH:
- Legacy `Check: resource.ComposeAggregateTestCheckFunc(...)` blocks
- Modern `ConfigStateChecks: []statecheck.StateCheck{...}` blocks

**Impact**: The legacy Check blocks are redundant duplicates of the modern ConfigStateChecks

**Resolution Strategy**:
- **DO NOT CONVERT** the legacy Check blocks to modern syntax
- **REMOVE** the legacy Check blocks entirely
- **KEEP** the existing ConfigStateChecks (already correct)

**Affected Functions**:
- resource_cmkube_cluster_test.go: 4 functions (Basic, DriftDetection, WorkerNodes, CompleteConfiguration)
- data_source_cmpart_partitions_test.go: 5 functions (ComputedFields, AttributeTypes, ListAttributes, FilterEmptyString, FilterSubsetProperty)

### 5.5 P3 Test Functions

**Status**: Functions with "P3" prefix in resource_cmkube_cluster_test.go are fully modern but still have redundant legacy Check blocks
**Action**: Remove legacy Check blocks from P3 functions as well for 100% consistency

---

## 6. Transformation Workflow (Per Priority)

### 6.1 P1 File: resource_cmnet_network_test.go

**Complexity**: Low (straightforward conversions)
**Estimated Time**: 30-45 minutes

**Steps**:
1. Add 4 required imports
2. TestAccCMNetNetwork_Basic:
   - Convert 4 legacy assertions to ConfigStateChecks
   - Add compareID initialization and tracking
3. TestAccCMNetNetwork_Complete:
   - Convert 10 legacy assertions to ConfigStateChecks
   - Add idempotency step after Create
4. TestAccCMNetNetwork_Update:
   - Convert 4 legacy assertions (2 per step) to ConfigStateChecks
   - Add idempotency step after Create
   - Add idempotency step after Update

**Validation**: `TF_ACC=1 go test -v -run TestAccCMNetNetwork`

### 6.2 P2 File: resource_cmkube_cluster_test.go

**Complexity**: High (mixed patterns, remove redundant blocks)
**Estimated Time**: 60-90 minutes

**Steps**:
1. No import additions needed (already complete)
2. For each of 4 legacy functions:
   - **REMOVE** legacy Check block entirely
   - **KEEP** existing ConfigStateChecks
3. For each of 6 P3 functions:
   - **REMOVE** legacy Check block entirely
   - **KEEP** existing ConfigStateChecks
4. Validation tests have no assertions, no changes needed

**Validation**: `TF_ACC=1 go test -v -run TestAccCMKubeCluster`

### 6.3 P3 File: data_source_cmpart_partitions_test.go

**Complexity**: Medium (remove redundant blocks)
**Estimated Time**: 45-60 minutes

**Steps**:
1. No import additions needed (already complete)
2. For each of 5 legacy functions:
   - **REMOVE** legacy Check block entirely
   - **KEEP** existing ConfigStateChecks
3. Other 3 functions already fully modern, no changes needed

**Validation**: `TF_ACC=1 go test -v -run TestAccCMPartPartitions`

---

## 7. Baseline Performance Metrics

**Baseline Command**:
```bash
TF_ACC=1 time go test -v -timeout 120m ./internal/provider/ -run "TestAccCMNetNetwork|TestAccCMKubeCluster|TestAccCMPartPartitions"
```

**Note**: Baseline timing will be executed in task T007 to establish performance benchmark before any transformations.

**Success Threshold**: Post-modernization execution time must be within 10% of baseline.

---

## 8. Risk Assessment

### 8.1 High Risk Areas

1. **Numeric Type Mismatches**: Using Int64Exact when BCM returns float64
   - **Mitigation**: Start with Int64Exact, document failures, switch to Float64Exact if needed

2. **Test Execution Regressions**: Modern patterns adding overhead
   - **Mitigation**: Measure after each phase, compare to baseline

3. **Incomplete Legacy Removal**: Missing some legacy assertions during grep analysis
   - **Mitigation**: Run grep verification after each file transformation

### 8.2 Medium Risk Areas

1. **Import Conflicts**: Adding imports causing Go module issues
   - **Mitigation**: Run `go mod tidy` after each file modification

2. **BCM Cluster Availability**: Cluster offline during validation
   - **Mitigation**: Can complete transformations offline, defer validation

### 8.3 Low Risk Areas

1. **CheckDestroy Compliance**: Both resource CheckDestroy functions already compliant
2. **Test Helper Compatibility**: No modifications needed to test_helpers.go
3. **Modern Pattern Support**: terraform-plugin-testing v1.13.3 fully supports all patterns

---

## 9. Recommendations

### 9.1 Execution Order Rationale

1. **Start with P1** (resource_cmnet_network_test.go): Smallest file, straightforward conversions, establishes workflow
2. **Proceed to P2** (resource_cmkube_cluster_test.go): Most complex, but demonstrates removal pattern clearly
3. **Finish with P3** (data_source_cmpart_partitions_test.go): Data source patterns, reinforces removal approach

### 9.2 Critical Success Factors

1. **Verify after each function**: Run individual test after each function transformation
2. **Grep verification**: Confirm zero legacy patterns remain after each file
3. **Baseline comparison**: Track execution time at each phase gate
4. **Full suite validation**: Run complete test suite after all phases

### 9.3 Documentation Updates

If new patterns emerge during implementation:
- Update CLAUDE.md Modern Testing Patterns section
- Document numeric type decisions (Int64Exact vs Float64Exact)
- Add troubleshooting guide for common issues

---

## Appendix A: Legacy Assertion Count by Type

| File | String | Int64 | Bool | Computed | Total |
|------|--------|-------|------|----------|-------|
| resource_cmnet_network_test.go | 14 | 2 | 1 | 1+2 | 18 |
| resource_cmkube_cluster_test.go | 19 | 7 | 0 | 12 | 38 |
| data_source_cmpart_partitions_test.go | 0 | 0 | 0 | 10 | 10 |
| **TOTAL** | **33** | **9** | **1** | **23** | **66** |

## Appendix B: Function Modernization Strategy

| File | Total Functions | Fully Modern | Need Conversion | Need Removal |
|------|----------------|--------------|-----------------|--------------|
| resource_cmnet_network_test.go | 3 | 0 | 3 | 0 |
| resource_cmkube_cluster_test.go | 10 | 0 | 0 | 10 (remove legacy blocks) |
| data_source_cmpart_partitions_test.go | 8 | 3 | 0 | 5 (remove legacy blocks) |
| **TOTAL** | **21** | **3** | **3** | **15** |

**Key Insight**: Only 3 functions (all in resource_cmnet_network_test.go) need full conversion. The other 15 functions just need redundant legacy Check blocks removed!

---

**Research Complete**: Ready for Phase 2 (Design & Contracts)
