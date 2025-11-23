# Contract 05: List Attribute Transformation

## Purpose

Transform legacy list count assertions from `resource.TestCheckResourceAttr("list.#", "2")` to modern `statecheck.ExpectKnownValue()` with `knownvalue.ListSizeExact(2)` for known counts or `knownvalue.NotNull()` for environment-dependent counts.

---

## Before (Legacy Pattern)

**Known Count (Resource)**:
```go
{
    Config: testAccCMKubeClusterResourceConfigWithWorkers(clusterName, masterNodeUUID, []string{uuid1, uuid2}),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "worker_nodes.#", "2"),
        resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "dns_servers.#", "2"),
    ),
},
```

**Unknown Count (Data Source)**:
```go
{
    Config: testAccCMPartPartitionsDataSourceConfig(),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.#"),
    ),
},
```

---

## After (Modern Pattern)

**Known Count (Resource)**:
```go
{
    Config: testAccCMKubeClusterResourceConfigWithWorkers(clusterName, masterNodeUUID, []string{uuid1, uuid2}),
    ConfigStateChecks: []statecheck.StateCheck{
        statecheck.ExpectKnownValue(
            "bcm_cmkube_cluster.test",
            tfjsonpath.New("worker_nodes"),
            knownvalue.ListSizeExact(2),  // We know exact count from test config
        ),
        statecheck.ExpectKnownValue(
            "bcm_cmkube_cluster.test",
            tfjsonpath.New("dns_servers"),
            knownvalue.ListSizeExact(2),
        ),
    },
},
```

**Unknown Count (Data Source)**:
```go
{
    Config: testAccCMPartPartitionsDataSourceConfig(),
    ConfigStateChecks: []statecheck.StateCheck{
        statecheck.ExpectKnownValue(
            "data.bcm_cmpart_partitions.test",
            tfjsonpath.New("partitions"),
            knownvalue.NotNull(),  // Count varies by environment
        ),
    },
},
```

---

## Key Changes

1. **Path Difference**: Legacy uses `"list.#"` (with suffix), modern uses `"list"` (without suffix)
2. **Known Counts**: Use `ListSizeExact(n)` when test creates list with known size
3. **Unknown Counts**: Use `NotNull()` when list size varies by environment
4. **Type Safety**: Validates list type, not just string count comparison

---

## Decision Matrix: ListSizeExact vs NotNull

| Scenario | Matcher | Rationale |
|----------|---------|-----------|
| **Resource with test-created list** | `ListSizeExact(n)` | Test config defines exact list (`worker_nodes = [uuid1, uuid2]`) |
| **Data source listing environment resources** | `NotNull()` | BCM cluster may have varying number of partitions |
| **Empty list in test** | `ListSizeExact(0)` | Test explicitly sets empty list (`worker_nodes = []`) |
| **Optional list, may be null** | `NotNull()` | List may or may not be present |

---

## Required Imports

```go
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
```

---

## Examples from Target Files

### Example 1: Worker Nodes - Known Size (resource_cmkube_cluster_test.go)

**Test Context**: Creating cluster with exactly 2 worker nodes
**Config**: `worker_nodes = [uuid1, uuid2]`

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "worker_nodes.#", "2")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes"),  // No .# suffix
    knownvalue.ListSizeExact(2),
)
```

### Example 2: Empty List (resource_cmkube_cluster_test.go)

**Test Context**: Scaling down to zero workers
**Config**: `worker_nodes = []`

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "worker_nodes.#", "0")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes"),
    knownvalue.ListSizeExact(0),  // Explicit empty list check
)
```

### Example 3: Data Source List - Unknown Size (data_source_cmpart_partitions_test.go)

**Test Context**: Querying BCM cluster for partitions (count varies by environment)

**Before**:
```go
resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.#")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "data.bcm_cmpart_partitions.test",
    tfjsonpath.New("partitions"),  // No .# suffix
    knownvalue.NotNull(),  // Just verify list exists, don't check count
)
```

### Example 4: Fixed DNS Servers (resource_cmkube_cluster_test.go)

**Test Context**: Cluster with exactly 2 DNS servers
**Config**: `dns_servers = ["8.8.8.8", "8.8.4.4"]`

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "dns_servers.#", "2")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("dns_servers"),
    knownvalue.ListSizeExact(2),
)
```

---

## Common Mistakes to Avoid

❌ **Using .# suffix in path**:
```go
// WRONG - Don't use .# suffix in modern pattern
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes.#"),  // ❌ Remove .#
    knownvalue.ListSizeExact(2),
)
```

❌ **Using Int64Exact for list counts**:
```go
// WRONG - Use ListSizeExact, not Int64Exact
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes"),
    knownvalue.Int64Exact(2),  // ❌ Wrong matcher type!
)
```

❌ **Using ListSizeExact for environment-dependent lists**:
```go
// WRONG - Environment may have different partition count
statecheck.ExpectKnownValue(
    "data.bcm_cmpart_partitions.test",
    tfjsonpath.New("partitions"),
    knownvalue.ListSizeExact(3),  // ❌ Hardcodes cluster state!
)
```

❌ **Using NotNull when count is known**:
```go
// SUBOPTIMAL - We know there are exactly 2 workers in test config
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes"),
    knownvalue.NotNull(),  // ⚠️ Too weak - use ListSizeExact(2)
)
```

---

## List Item Validation (Advanced)

For validating **specific items** in a list:

**Check First Item Value**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("master_nodes").AtSliceIndex(0),
    knownvalue.StringExact(masterNodeUUID),
)
```

**Check List Contains Exact Values**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("dns_servers"),
    knownvalue.SetExact([]knownvalue.Check{
        knownvalue.StringExact("8.8.8.8"),
        knownvalue.StringExact("8.8.4.4"),
    }),
)
```

**Note**: For simplicity, current transformation focuses on list SIZE validation. Item-level validation is advanced use case.

---

## Nested List Attributes

For nested lists in data sources:

**Example**: Verify list attribute within partition object
```go
// partitions.0.admin_email is a List[String]
statecheck.ExpectKnownValue(
    "data.bcm_cmpart_partitions.test",
    tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("admin_email"),
    knownvalue.NotNull(),  // Unknown count of admin emails
)
```

---

## Verification

### Visual Inspection
- ✅ No `.#` suffix in `tfjsonpath.New()`
- ✅ Uses `ListSizeExact(n)` for known counts (resources)
- ✅ Uses `NotNull()` for unknown counts (data sources)
- ✅ List count as integer (`2` not `"2"`)

### Semantic Verification
Ask: "Does the test config define this list size?"
- **YES** → Use `ListSizeExact(n)`
- **NO** → Use `NotNull()`

---

## Known List Attributes in Target Files

### resource_cmkube_cluster_test.go (Known Counts)

**worker_nodes**:
- 1 worker: Lines 374 → `ListSizeExact(1)`
- 2 workers: Line 392 → `ListSizeExact(2)`
- 0 workers: Line 410 → `ListSizeExact(0)`

**master_nodes**:
- Line 488 → `ListSizeExact(1)`
- Line 489 → `ListSizeExact(1)`

**dns_servers**:
- Line 690 → `ListSizeExact(2)` (config: `["8.8.8.8", "8.8.4.4"]`)
- Line 724 → `ListSizeExact(1)` (config: `["1.1.1.1"]`)
- Line 882 → `ListSizeExact(2)`

**Total Resource List Transformations**: 9 occurrences (all ListSizeExact)

### data_source_cmpart_partitions_test.go (Unknown Counts)

**partitions**:
- Line 95 → `NotNull()` (environment-dependent count)
- Line 168 → `NotNull()`
- Line 228 → `NotNull()`
- Line 335 → `NotNull()`
- Line 368 → `NotNull()`
- Line 389 → `NotNull()`

**Nested list attributes** (within partitions.0):
- `admin_email` → `NotNull()` (unknown count)
- `time_servers` → `NotNull()`
- `search_domains` → `NotNull()`
- `name_servers` → `NotNull()`

**Total Data Source List Transformations**: 6 partitions.# + 4 nested (use NotNull)

---

## Environment Portability Principle

**From spec.md clarifications**:
> Mixed strategy: Use `ListSizeExact(n)` for test-created resources with known counts. Use `NotNull()` for environment-dependent data sources.

This ensures tests work on any BCM cluster configuration without hardcoding cluster-specific counts.

---

## Applicable To

- **File**: resource_cmkube_cluster_test.go (worker_nodes, master_nodes, dns_servers)
- **File**: data_source_cmpart_partitions_test.go (partitions, nested list attributes)
- **Attributes**: All list-type attributes
- **Test Functions**: 9 resource list checks + 10 data source list checks

---

**Contract Status**: ✅ Complete - Ready for Implementation
