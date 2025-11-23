# Contract 04: Computed Attribute Transformation

## Purpose

Transform legacy computed field existence checks from `resource.TestCheckResourceAttrSet()` to modern `statecheck.ExpectKnownValue()` with `knownvalue.NotNull()` matcher.

---

## Before (Legacy Pattern)

```go
{
    Config: testAccCMNetNetworkConfigBasic(networkName),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "id"),
        resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid"),
        // Only checks EXISTENCE - doesn't validate type or specific value
    ),
},
```

---

## After (Modern Pattern)

```go
{
    Config: testAccCMNetNetworkConfigBasic(networkName),
    ConfigStateChecks: []statecheck.StateCheck{
        statecheck.ExpectKnownValue(
            "bcm_cmnet_network.test",
            tfjsonpath.New("id"),
            knownvalue.NotNull(),  // Validates NOT NULL with type safety
        ),
        statecheck.ExpectKnownValue(
            "bcm_cmnet_network.test",
            tfjsonpath.New("uuid"),
            knownvalue.NotNull(),
        ),
    },
},
```

---

## Key Changes

1. **Function Change**: `TestCheckResourceAttrSet()` → `ExpectKnownValue()` with `NotNull()`
2. **Type Safety**: Modern pattern validates type exists (not just non-empty string)
3. **Null Handling**: Explicitly checks for null values, not just empty strings
4. **Better Semantics**: "NotNull" is clearer than "ResourceAttrSet"

---

## Required Imports

```go
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
```

---

## Examples from Target Files

### Example 1: Resource ID (resource_cmnet_network_test.go)

**Before**:
```go
resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "id")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("id"),
    knownvalue.NotNull(),
)
```

### Example 2: Resource UUID (resource_cmnet_network_test.go)

**Before**:
```go
resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("uuid"),
    knownvalue.NotNull(),
)
```

### Example 3: Data Source Computed Field (data_source_cmpart_partitions_test.go)

**Before**:
```go
resource.TestCheckResourceAttrSet("data.bcm_cmpart_partitions.test", "partitions.0.uuid")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "data.bcm_cmpart_partitions.test",
    tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("uuid"),
    knownvalue.NotNull(),
)
```

---

## When to Use NotNull()

Use `NotNull()` matcher for:

1. **Computed Attributes**: Values calculated by provider/backend
   - Examples: `id`, `uuid`, `creation_time`, `revision_id`

2. **API-Generated Values**: Values set by external system
   - Examples: Cluster IPs, auto-assigned ports, generated names

3. **Unknown Values in Data Sources**: Environment-dependent data
   - Examples: Partition attributes, network configurations from BCM cluster

4. **Optional Computed Fields**: May or may not be present
   - If testing presence, use `NotNull()`
   - If testing absence, use different matcher or skip check

---

## When NOT to Use NotNull()

❌ **Don't use for known values**:
```go
// WRONG - We know the value should be "test-network"
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("name"),
    knownvalue.NotNull(),  // ❌ Too weak - use StringExact(networkName)
)
```

❌ **Don't use for hardcoded test values**:
```go
// WRONG - We set mtu=9000 in config
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("mtu"),
    knownvalue.NotNull(),  // ❌ Too weak - use Int64Exact(9000)
)
```

---

## Nested Path Syntax

For nested attributes in lists/maps:

**List Index Access**:
```go
// Legacy: "partitions.0.uuid"
// Modern:
tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("uuid")
```

**Map Key Access**:
```go
// Legacy: "metadata.creation_time"
// Modern:
tfjsonpath.New("metadata").AtMapKey("creation_time")
```

**Deep Nesting**:
```go
// Legacy: "clusters.0.nodes.1.ip"
// Modern:
tfjsonpath.New("clusters").
    AtSliceIndex(0).
    AtMapKey("nodes").
    AtSliceIndex(1).
    AtMapKey("ip")
```

---

## Common Mistakes to Avoid

❌ **Using StringExact with computed values**:
```go
// WRONG - UUID is generated, we don't know the value
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("uuid"),
    knownvalue.StringExact("abc-123-def"),  // ❌ Hardcoded UUID!
)
```

❌ **Missing NotNull() matcher**:
```go
// WRONG - Must provide a matcher
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("uuid"),
    // ❌ Missing third argument - won't compile
)
```

❌ **Using NotNull() for lists when size matters**:
```go
// SUBOPTIMAL - If we know list size, be specific
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("master_nodes"),
    knownvalue.NotNull(),  // ⚠️ Weak - use ListSizeExact(1) instead
)
```

---

## Type Safety Benefit

### Without Type Safety (Legacy)
```go
// API bug: Returns empty string "" for uuid (should be null or value)
// Test PASSES (empty string is "set")
resource.TestCheckResourceAttrSet("resource", "uuid")  // ✅ False positive
```

### With Type Safety (Modern)
```go
// API bug: Returns empty string "" for uuid
// Test behavior depends on implementation:
// - Empty string "" is NOT null → test passes
// - Null value → test fails as expected
statecheck.ExpectKnownValue(
    "resource",
    tfjsonpath.New("uuid"),
    knownvalue.NotNull(),  // Validates actual null, not just "set"
)
```

---

## Verification

### Visual Inspection
- ✅ Uses `NotNull()` matcher for computed fields
- ✅ Uses `StringExact()` / `Int64Exact()` / `Bool()` for known values
- ✅ Nested paths use `.AtSliceIndex()` / `.AtMapKey()` syntax

### Semantic Verification
Ask: "Do I know this value at test time?"
- **YES** → Use specific matcher (`StringExact`, `Int64Exact`, `Bool`)
- **NO** → Use `NotNull()`

---

## Known Computed Attributes in Target Files

### resource_cmnet_network_test.go
- `id`: Resource identifier → `NotNull()`
- `uuid`: Resource UUID → `NotNull()`
**Occurrences**: 3 (lines 31, 32, 77)

### resource_cmkube_cluster_test.go
- `uuid`: Cluster UUID → `NotNull()` (lines 180)
- `id`: Cluster ID → `NotNull()` (lines 181)
- `storage_classes`: JSON-encoded storage → `NotNull()` (lines 759, 884)
- `addons`: JSON-encoded addons → `NotNull()` (lines 816, 845, 885)
- `ingress_controller`: JSON config → `NotNull()` (lines 817, 846, 886)
- `overlay_network`: Network UUID → `NotNull()` (line 887)
**Occurrences**: 12

### data_source_cmpart_partitions_test.go
- All `TestCheckResourceAttrSet` calls for computed fields
- `partitions.#`: List count → `NotNull()` (5 occurrences)
- `partitions.0.id`: Partition ID → `NotNull()`
- `partitions.0.uuid`: Partition UUID → `NotNull()`
- `partitions.0.name`: Partition name → `NotNull()`
- `partitions.0.base_type`: Base type → `NotNull()`
**Occurrences**: 10

**Total Computed Attribute Transformations**: 25 (3 P1 + 12 P2 + 10 P3)

---

## Decision Matrix: NotNull() vs Specific Matcher

| Attribute Characteristic | Matcher | Example |
|--------------------------|---------|---------|
| Set in test config, known value | Specific (`StringExact`, `Int64Exact`, `Bool`) | `name = "test-network"` |
| Computed by provider/backend | `NotNull()` | `uuid`, `id`, `creation_time` |
| Environment-dependent data | `NotNull()` | Data source results |
| API-generated value | `NotNull()` | Auto-assigned IPs, ports |
| JSON-encoded field | `NotNull()` | `addons`, `storage_classes` |

---

## Applicable To

- **File**: All 3 target files
- **Attributes**: All computed attributes (id, uuid, timestamps, JSON fields)
- **Test Functions**: 25 total computed attribute checks

---

**Contract Status**: ✅ Complete - Ready for Implementation
