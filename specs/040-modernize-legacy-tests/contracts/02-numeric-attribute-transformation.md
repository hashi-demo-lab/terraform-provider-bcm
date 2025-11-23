# Contract 02: Numeric Attribute Transformation

## Purpose

Transform legacy numeric attribute assertions from string-based `resource.TestCheckResourceAttr("attr", "9000")` to type-safe `statecheck.ExpectKnownValue()` with `knownvalue.Int64Exact(9000)` matcher.

---

## Before (Legacy Pattern)

```go
{
    Config: testAccCMNetNetworkConfigComplete(networkName),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000"),
        // Numeric value stored as STRING - loses type safety
    ),
},
```

---

## After (Modern Pattern)

```go
{
    Config: testAccCMNetNetworkConfigComplete(networkName),
    ConfigStateChecks: []statecheck.StateCheck{
        statecheck.ExpectKnownValue(
            "bcm_cmnet_network.test",
            tfjsonpath.New("mtu"),
            knownvalue.Int64Exact(9000),  // Type-safe integer - no quotes!
        ),
    },
},
```

---

## Key Changes

1. **Remove Quotes**: Legacy uses string `"9000"`, modern uses integer `9000`
2. **Type Matcher**: Use `knownvalue.Int64Exact(value)` for numeric attributes
3. **Type Safety**: Catches type mismatches at test time instead of runtime
4. **Integer Type**: Terraform framework uses `types.Int64`, so use `Int64Exact()` not `Float64Exact()`

---

## Numeric Type Decision Tree

```
Is attribute a numeric type?
├── YES → Try Int64Exact(value) first
│   ├── Test passes? → ✅ Use Int64Exact()
│   └── Type error? → Try Float64Exact(value)
│       ├── Test passes? → ✅ Use Float64Exact()
│       └── Still fails? → ❌ Check schema definition
└── NO → Use StringExact() or other appropriate matcher
```

---

## Required Imports

```go
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
```

---

## Examples from Target Files

### Example 1: MTU (resource_cmnet_network_test.go)

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("mtu"),
    knownvalue.Int64Exact(9000),  // No quotes - integer type
)
```

---

## List Count Attributes (Special Case)

### ❌ Wrong Approach: Using Int64Exact for List Counts

```go
// WRONG - Don't use Int64Exact for list.# attributes
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes").AtMapKey("#"),  // ❌ Wrong!
    knownvalue.Int64Exact(2),
)
```

### ✅ Correct Approach: Using ListSizeExact for List Counts

```go
// CORRECT - Use ListSizeExact for list size validation
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("worker_nodes"),
    knownvalue.ListSizeExact(2),  // ✅ Correct!
)
```

**Rationale**: `TestCheckResourceAttr("worker_nodes.#", "2")` checks the COUNT of list items. In modern patterns, use `ListSizeExact()` to validate list size, not `Int64Exact()`.

**See Also**: Contract 05 (List Attribute Transformation)

---

## Common Mistakes to Avoid

❌ **Keeping quotes around numeric values**:
```go
// WRONG - Numeric values must not be quoted
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("mtu"),
    knownvalue.Int64Exact("9000"),  // ❌ Won't compile - remove quotes!
)
```

❌ **Using StringExact for numeric attributes**:
```go
// WRONG - Loses type safety
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("mtu"),
    knownvalue.StringExact("9000"),  // ❌ Wrong matcher type!
)
```

❌ **Using Int64Exact for list counts** (see example above):
```go
// WRONG - Use ListSizeExact for lists
knownvalue.Int64Exact(2)  // ❌ for "worker_nodes.#"
```

---

## Verification

### Visual Inspection
- ✅ Numeric values have no quotes: `9000` not `"9000"`
- ✅ Uses `Int64Exact()` matcher (or `Float64Exact()` if needed)
- ✅ List counts transformed to `ListSizeExact()` not `Int64Exact()`

### Test Execution
```bash
# If test fails with type error like:
# "expected int64, got float64"
# Switch from Int64Exact to Float64Exact

# Example failure:
# Error: value type mismatch: expected types.Int64Type, got types.Float64Type

# Fix: Change Int64Exact(9000) → Float64Exact(9000.0)
```

---

## Known Numeric Attributes in Target Files

### resource_cmnet_network_test.go
- `mtu`: Network MTU (Maximum Transmission Unit)
  - Legacy: `"9000"` (string)
  - Modern: `9000` (int64)
  - Occurrences: 3 (lines 71, 104)

### resource_cmkube_cluster_test.go
- List counts (use ListSizeExact, NOT Int64Exact):
  - `master_nodes.#`: Lines 488 (legacy: `"1"`)
  - `worker_nodes.#`: Lines 374 (`"1"`), 392 (`"2"`), 410 (`"0"`), 489 (`"1"`)
  - `dns_servers.#`: Lines 690 (`"2"`), 724 (`"1"`), 882 (`"2"`)

**Total Numeric Transformations**: 9 occurrences (2 actual Int64, 7 list counts → ListSizeExact)

---

## Edge Case: Port Numbers, IDs, Timestamps

If you encounter attributes like:
- `port`: Network port (e.g., 8080) → `Int64Exact(8080)`
- `timeout`: Seconds (e.g., 300) → `Int64Exact(300)`
- `creation_time`: Unix timestamp → Use `NotNull()` (computed, variable value)
- `replica_count`: Integer count → `Int64Exact(3)`

**Rule**: If value is **known and fixed** in test config → `Int64Exact(value)`
**Rule**: If value is **computed or variable** → `NotNull()` (See Contract 04)

---

## Applicable To

- **File**: resource_cmnet_network_test.go (mtu attribute)
- **File**: resource_cmkube_cluster_test.go (list counts - use ListSizeExact)
- **Attributes**: All integer-type attributes with fixed test values
- **Test Functions**: 2 actual numeric assertions + 7 list counts

---

**Contract Status**: ✅ Complete - Ready for Implementation
