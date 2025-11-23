# Data Model: Test Function Structure Mapping

**Feature**: Modernize Legacy Testing Patterns (Issue #40)
**Date**: 2025-11-23

## Purpose

This document maps the structure of all 21 test functions across 3 target files, detailing current state, legacy assertions, missing checks, and transformation requirements.

---

## Test File 1: resource_cmnet_network_test.go

**File Stats**:
- Total Functions: 3
- Legacy Assertions: 18
- Functions Needing Full Conversion: 3
- Functions Needing Legacy Block Removal: 0

### TestAccCMNetNetwork_Basic

**Type**: Resource CRUD lifecycle with Import
**Location**: Lines 18-54
**Current Steps**: 3 (Create, Idempotency, Import)

**Current State**:
- ✅ Has idempotency check (lines 37-44)
- ✅ Has Import step (lines 46-51)
- ❌ Missing ID tracking (compareID pattern)

**Legacy Assertions** (4 total - lines 29-34):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),           // String
    resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "id"),                       // Computed
    resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid"),                     // Computed
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "domain_name", "cluster.local"), // String
),
```

**Modern Pattern Required**:
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("name"), knownvalue.StringExact(networkName)),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("id"), knownvalue.NotNull()),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("uuid"), knownvalue.NotNull()),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("domain_name"), knownvalue.StringExact("cluster.local")),
},
```

**Additional Changes Needed**:
1. Add `compareID := statecheck.CompareValue(compare.ValuesSame())` at function start
2. Add `compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id"))` to Create step ConfigStateChecks
3. Add `compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id"))` to Import step ConfigStateChecks

**Transformation Tasks**: T022-T027

---

### TestAccCMNetNetwork_Complete

**Type**: Resource full configuration validation
**Location**: Lines 56-82
**Current Steps**: 1 (Create only)

**Current State**:
- ❌ No idempotency check
- ❌ No Import step (not required for this test)
- ❌ No ID tracking (not required without Import)

**Legacy Assertions** (10 total - lines 67-78):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),                      // String
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "subnet", "192.168.100.0/24"),             // String
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "gateway", "192.168.100.1"),               // String
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000"),                            // Int64
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "domain_name", "test.local"),              // String
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_range_start", "192.168.100.100"),    // String
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_range_end", "192.168.100.200"),      // String
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_enabled", "true"),                   // Bool
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "notes", "Test network"),                  // String
    resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid"),                                // Computed
),
```

**Modern Pattern Required**:
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("name"), knownvalue.StringExact(networkName)),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("subnet"), knownvalue.StringExact("192.168.100.0/24")),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("gateway"), knownvalue.StringExact("192.168.100.1")),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("mtu"), knownvalue.Int64Exact(9000)),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("domain_name"), knownvalue.StringExact("test.local")),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("dhcp_range_start"), knownvalue.StringExact("192.168.100.100")),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("dhcp_range_end"), knownvalue.StringExact("192.168.100.200")),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("dhcp_enabled"), knownvalue.Bool(true)),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("notes"), knownvalue.StringExact("Test network")),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("uuid"), knownvalue.NotNull()),
},
```

**Additional Changes Needed**:
1. Add idempotency step after Create step:
```go
{
    Config: testAccCMNetNetworkConfigComplete(networkName),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

**Transformation Tasks**: T028-T030

---

### TestAccCMNetNetwork_Update

**Type**: Resource update validation
**Location**: Lines 84-110
**Current Steps**: 2 (Create, Update)

**Current State**:
- ❌ No idempotency check after Create
- ❌ No idempotency check after Update
- ❌ No Import step (not required for this test)

**Legacy Assertions**:
- **Create Step** (lines 95-96): 1 assertion
- **Update Step** (lines 103-106): 3 assertions

**Create Step Legacy**:
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName), // String
),
```

**Create Step Modern**:
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("name"), knownvalue.StringExact(networkName)),
},
```

**Update Step Legacy**:
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),           // String
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000"),                 // Int64
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "notes", "Updated notes"),      // String
),
```

**Update Step Modern**:
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("name"), knownvalue.StringExact(networkName)),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("mtu"), knownvalue.Int64Exact(9000)),
    statecheck.ExpectKnownValue("bcm_cmnet_network.test", tfjsonpath.New("notes"), knownvalue.StringExact("Updated notes")),
},
```

**Additional Changes Needed**:
1. Add idempotency step after Create (insert at line 98)
2. Add idempotency step after Update (insert at line 107)

**Transformation Tasks**: T031-T035

---

## Test File 2: resource_cmkube_cluster_test.go

**File Stats**:
- Total Functions: 10
- Legacy Assertions: 38
- Functions Needing Full Conversion: 0
- Functions Needing Legacy Block Removal: 10 (all have redundant Check blocks)

**IMPORTANT PATTERN**: All functions in this file have BOTH legacy Check blocks AND modern ConfigStateChecks. The modern ConfigStateChecks are already correct. We only need to **REMOVE** the redundant legacy Check blocks.

---

### TestAccCMKubeClusterResource_Basic

**Type**: Complete CRUD lifecycle with ID tracking
**Location**: Lines 160-254
**Current Steps**: 5 (Create, Idempotency, Import, Update, Idempotency)

**Current State**:
- ✅ Has idempotency checks (lines 205-213, 243-251)
- ✅ Has Import step with ImportStateVerifyIgnore (lines 214-223)
- ✅ Has ID tracking (lines 168, 199-202, 237-240)
- ✅ Has modern ConfigStateChecks (lines 183-203, 230-241)
- ❌ Has redundant legacy Check block (lines 178-182, 227-229)

**Legacy Block to Remove** (lines 178-182):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "name", clusterName),
    resource.TestCheckResourceAttrSet("bcm_cmkube_cluster.test", "uuid"),
    resource.TestCheckResourceAttrSet("bcm_cmkube_cluster.test", "id"),
),
```

**Legacy Block to Remove** (line 228):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "name", clusterNameUpdated),
),
```

**Action**: Delete lines 178-182 and line 228. Keep all ConfigStateChecks.

**Transformation Tasks**: T041-T043

---

### TestAccCMKubeClusterResource_DriftDetection

**Type**: External modification detection
**Location**: Lines 257-352
**Current Steps**: 3 (Create, Drift simulation with PreConfig, Restore)

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 272-278, 342-348)
- ❌ Has redundant legacy Check blocks (line 270, line 340)

**Legacy Blocks to Remove**:
- Line 270: `Check: resource.ComposeAggregateTestCheckFunc(...)`
- Line 340: `Check: resource.ComposeAggregateTestCheckFunc(...)`

**Action**: Delete both legacy Check blocks. Keep ConfigStateChecks and PreConfig drift logic.

**Transformation Tasks**: T044-T046

---

### TestAccCMKubeClusterResource_WorkerNodes

**Type**: Worker node scaling (add/remove)
**Location**: Lines 355-422
**Current Steps**: 3 (1 worker, 2 workers, 0 workers)

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 376-382, 394-400, 412-418)
- ❌ Has redundant legacy Check blocks (lines 374, 392, 410)

**Legacy Blocks to Remove**:
- Line 374: Worker count = 1
- Line 392: Worker count = 2
- Line 410: Worker count = 0

**Action**: Delete all three legacy Check blocks. Keep ConfigStateChecks.

**Transformation Tasks**: T047-T050

---

### TestAccCMKubeClusterResource_ValidationInvalidName

**Type**: Validation test (ExpectError)
**Location**: Lines 424-438
**Current Steps**: 1 (validation error expected)

**Current State**:
- ✅ No assertions needed (uses ExpectError)
- ✅ No legacy Check blocks

**Action**: No changes needed

---

### TestAccCMKubeClusterResource_ValidationInvalidVersion

**Type**: Validation test (ExpectError)
**Location**: Lines 441-455
**Current Steps**: 1 (validation error expected)

**Current State**:
- ✅ No assertions needed (uses ExpectError)
- ✅ No legacy Check blocks

**Action**: No changes needed

---

### TestAccCMKubeClusterResource_CompleteConfiguration

**Type**: Full optional fields validation
**Location**: Lines 462-564
**Current Steps**: 3 (Create, Idempotency, Update)

**Current State**:
- ✅ Has idempotency check (lines 521-535)
- ✅ Has modern ConfigStateChecks (lines 493-519, 549-561)
- ❌ Has redundant legacy Check blocks (lines 486-492, 545-548)

**Legacy Blocks to Remove**:
- Lines 486-492: Create step assertions (5 total)
- Lines 545-548: Update step assertions (2 total)

**Action**: Delete both legacy Check blocks. Keep ConfigStateChecks.

**Transformation Tasks**: T051-T053

---

### TestAccCMKubeClusterResource_P3AdvancedNetworking

**Type**: P3 networking features
**Location**: Lines 674-741
**Current Steps**: 3 (Create, Idempotency, Update)

**Current State**:
- ✅ Has idempotency check (lines 710-718)
- ✅ Has modern ConfigStateChecks (lines 692-708, 726-737)
- ❌ Has redundant legacy Check blocks (lines 687-691, 722-725)

**Legacy Blocks to Remove**:
- Lines 687-691: Create step
- Lines 722-725: Update step

**Action**: Delete both legacy Check blocks. Keep ConfigStateChecks.

**Note**: This is a P3 (priority 3) test but in P2 file, will be handled in Phase 4 (P2 Implementation).

---

### TestAccCMKubeClusterResource_P3StorageAndLoadBalancer

**Type**: P3 storage and LB features
**Location**: Lines 744-799
**Current Steps**: 3 (Create, Idempotency, Update)

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 761-772, 789-795)
- ❌ Has redundant legacy Check blocks (lines 756-760, 786-788)

**Legacy Blocks to Remove**:
- Lines 756-760: Create step (3 assertions)
- Lines 786-788: Update step (1 assertion)

**Action**: Delete both legacy Check blocks. Keep ConfigStateChecks.

---

### TestAccCMKubeClusterResource_P3Addons

**Type**: P3 cluster addons
**Location**: Lines 802-864
**Current Steps**: 3 (Create, Idempotency, Update)

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 819-830, 848-860)
- ❌ Has redundant legacy Check blocks (lines 814-818, 844-847)

**Legacy Blocks to Remove**:
- Lines 814-818: Create step (3 assertions)
- Lines 844-847: Update step (2 assertions)

**Action**: Delete both legacy Check blocks. Keep ConfigStateChecks.

---

### TestAccCMKubeClusterResource_P3FullStack

**Type**: P3 all fields together
**Location**: Lines 867-939
**Current Steps**: 2 (Create, Idempotency)

**Current State**:
- ✅ Has idempotency check (lines 928-936)
- ✅ Has modern ConfigStateChecks (lines 889-927)
- ❌ Has redundant legacy Check block (lines 879-888)

**Legacy Block to Remove**:
- Lines 879-888: Create step (8 assertions)

**Action**: Delete legacy Check block. Keep ConfigStateChecks.

---

## Test File 3: data_source_cmpart_partitions_test.go

**File Stats**:
- Total Functions: 8
- Legacy Assertions: 10
- Functions Needing Full Conversion: 0
- Functions Needing Legacy Block Removal: 5
- Functions Already Fully Modern: 3 (Basic, FilterByNamePattern, NoMatches)

---

### TestAccCMPartPartitionsDataSource_Basic

**Type**: Basic data source read
**Location**: Lines 18-38
**Current Steps**: 1

**Current State**:
- ✅ Fully modern (ConfigStateChecks only)
- ✅ No legacy Check blocks

**Action**: No changes needed

---

### TestAccCMPartPartitionsDataSource_FilterByNamePattern

**Type**: Name pattern filtering
**Location**: Lines 41-61
**Current Steps**: 1

**Current State**:
- ✅ Fully modern (ConfigStateChecks only)
- ✅ No legacy Check blocks

**Action**: No changes needed

---

### TestAccCMPartPartitionsDataSource_NoMatches

**Type**: Empty result validation
**Location**: Lines 64-82
**Current Steps**: 1

**Current State**:
- ✅ Fully modern (ConfigStateChecks only)
- ✅ No legacy Check blocks

**Action**: No changes needed

---

### TestAccCMPartPartitionsDataSource_ComputedFields

**Type**: Verify all partition attributes exposed
**Location**: Lines 85-114
**Current Steps**: 1

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 98-109)
- ❌ Has redundant legacy Check block (lines 93-96)

**Legacy Block to Remove** (lines 93-96):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.#"),
),
```

**Action**: Delete lines 93-96. Keep ConfigStateChecks.

**Transformation Tasks**: T059-T060

---

### TestAccCMPartPartitionsDataSource_AttributeTypes

**Type**: Comprehensive type verification
**Location**: Lines 158-215
**Current Steps**: 1

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 177-211)
- ❌ Has redundant legacy Check block (lines 166-177)

**Legacy Block to Remove** (lines 166-177):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.#"),
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.0.id"),
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.0.uuid"),
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.0.name"),
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.0.base_type"),
),
```

**Action**: Delete lines 166-177. Keep ConfigStateChecks.

**Transformation Tasks**: T061-T062

---

### TestAccCMPartPartitionsDataSource_ListAttributes

**Type**: List attribute validation
**Location**: Lines 218-257
**Current Steps**: 1

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 230-253)
- ❌ Has redundant legacy Check block (lines 226-229)

**Legacy Block to Remove** (lines 226-229):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.#"),
),
```

**Action**: Delete lines 226-229. Keep ConfigStateChecks.

**Transformation Tasks**: T063-T064

---

### TestAccCMPartPartitionsDataSource_FilterCaseInsensitive

**Type**: Case-insensitive filtering validation
**Location**: Lines 260-323
**Current Steps**: 3 (lowercase, uppercase, mixed case)

**Current State**:
- ✅ Fully modern (ConfigStateChecks only, lines 269-282, 287-300, 305-318)
- ✅ No legacy Check blocks

**Action**: No changes needed

---

### TestAccCMPartPartitionsDataSource_FilterEmptyString

**Type**: Empty filter edge case
**Location**: Lines 326-354
**Current Steps**: 1

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 337-350)
- ❌ Has redundant legacy Check block (lines 333-336)

**Legacy Block to Remove** (lines 333-336):
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.#"),
),
```

**Action**: Delete lines 333-336. Keep ConfigStateChecks.

**Transformation Tasks**: T065-T066

---

### TestAccCMPartPartitionsDataSource_FilterSubsetProperty

**Type**: Mathematical subset property validation
**Location**: Lines 357-406
**Current Steps**: 2 (all partitions, filtered partitions)

**Current State**:
- ✅ Has modern ConfigStateChecks (lines 370-382, 391-403)
- ❌ Has redundant legacy Check blocks (lines 366-370, 386-390)

**Legacy Blocks to Remove**:
- Lines 366-370: First step
- Lines 386-390: Second step

**Action**: Delete both legacy Check blocks. Keep ConfigStateChecks.

**Transformation Tasks**: T067-T069

---

## CheckDestroy Function Compliance

### resource_cmnet_network_test.go

**Function**: `testAccCheckCMNetNetworkDestroy`
**Location**: Lines 112-151
**Compliance**: ✅ PASS

**Matches Enhanced Pattern**:
- ✅ Uses verifyResourceDeleted helper with exponential backoff
- ✅ Aggregates errors with detailed messages
- ✅ Tracks resource count
- ✅ Returns formatted error with all failures

**Action**: No changes needed

---

### resource_cmkube_cluster_test.go

**Function**: `testAccCheckCMKubeClusterDestroy`
**Location**: Lines 39-77
**Compliance**: ✅ PASS

**Matches Enhanced Pattern**:
- ✅ Uses verifyResourceDeleted helper with exponential backoff
- ✅ Aggregates errors with detailed messages
- ✅ Tracks resource count
- ✅ Returns formatted error with all failures

**Action**: No changes needed

---

### data_source_cmpart_partitions_test.go

**Function**: N/A (data sources don't create resources)
**Compliance**: N/A

---

## Migration Sequence

### Phase 1: Research (Complete)
- ✅ research.md created
- ✅ 66 legacy assertions cataloged
- ✅ Transformation rules documented

### Phase 2: Design (Current)
- ✅ data-model.md (this document)
- ⏳ contracts/ directory with 7 before/after examples
- ⏳ quickstart.md developer guide
- ⏳ Agent context update

### Phase 3: P1 Implementation
- **File**: resource_cmnet_network_test.go
- **Order**: Basic → Complete → Update
- **Gate**: All 3 functions pass before P2

### Phase 4: P2 Implementation
- **File**: resource_cmkube_cluster_test.go
- **Order**: Basic → DriftDetection → WorkerNodes → CompleteConfiguration → P3 functions
- **Gate**: All 10 functions pass before P3

### Phase 5: P3 Implementation
- **File**: data_source_cmpart_partitions_test.go
- **Order**: ComputedFields → AttributeTypes → ListAttributes → FilterEmptyString → FilterSubsetProperty
- **Gate**: All 8 functions pass before final validation

### Phase 6: Final Validation
- Full test suite
- Success criteria verification
- Performance baseline comparison

---

## Summary Statistics

### By File
| File | Total Functions | Need Conversion | Need Removal | Already Modern |
|------|----------------|-----------------|--------------|----------------|
| resource_cmnet_network_test.go | 3 | 3 | 0 | 0 |
| resource_cmkube_cluster_test.go | 10 | 0 | 8 | 2 (validation) |
| data_source_cmpart_partitions_test.go | 8 | 0 | 5 | 3 |
| **TOTAL** | **21** | **3** | **13** | **5** |

### By Transformation Type
| Type | Count | Files |
|------|-------|-------|
| Full Conversion (legacy → modern) | 3 | P1 |
| Removal Only (delete redundant Check blocks) | 13 | P2 (8) + P3 (5) |
| No Changes (already modern) | 5 | P2 (2) + P3 (3) |

### Legacy Assertion Distribution
| Assertion Type | Count | Transformation |
|----------------|-------|----------------|
| String (TestCheckResourceAttr with string) | 33 | knownvalue.StringExact() |
| Int64 (TestCheckResourceAttr with numeric) | 9 | knownvalue.Int64Exact() or ListSizeExact() |
| Bool (TestCheckResourceAttr with "true"/"false") | 1 | knownvalue.Bool() |
| Computed (TestCheckResourceAttrSet) | 23 | knownvalue.NotNull() |
| **TOTAL** | **66** | - |

---

**Data Model Complete**: Ready for contracts/ creation
