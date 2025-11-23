# Contract 03: Boolean Attribute Transformation

## Purpose

Transform legacy boolean attribute assertions from string-based `resource.TestCheckResourceAttr("attr", "true")` to type-safe `statecheck.ExpectKnownValue()` with `knownvalue.Bool(true)` matcher.

---

## Before (Legacy Pattern)

```go
{
    Config: testAccCMNetNetworkConfigComplete(networkName),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_enabled", "true"),
        // Boolean value stored as STRING - loses type safety
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
            tfjsonpath.New("dhcp_enabled"),
            knownvalue.Bool(true),  // Type-safe boolean - no quotes!
        ),
    },
},
```

---

## Key Changes

1. **Remove Quotes**: Legacy uses string `"true"`, modern uses boolean `true`
2. **Type Matcher**: Use `knownvalue.Bool(value)` for boolean attributes
3. **Type Safety**: Catches if API returns string `"true"` instead of boolean `true`
4. **Boolean Values**: Use `true` or `false` (lowercase, no quotes)

---

## Required Imports

```go
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
```

---

## Examples from Target Files

### Example 1: DHCP Enabled (resource_cmnet_network_test.go)

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_enabled", "true")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("dhcp_enabled"),
    knownvalue.Bool(true),
)
```

### Example 2: Boolean False (Hypothetical)

**Before**:
```go
resource.TestCheckResourceAttr("bcm_resource.test", "enabled", "false")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_resource.test",
    tfjsonpath.New("enabled"),
    knownvalue.Bool(false),  // Use false for false values
)
```

---

## BCM Attribute Type Mapping

In `internal/provider/data_source_cmpart_partitions_test.go`, boolean attributes use `knownvalue.NotNull()` because we don't know if they're true or false (environment-dependent):

**Computed Boolean Attributes**:
```go
// For data sources with unknown boolean values
statecheck.ExpectKnownValue(
    "data.bcm_cmpart_partitions.test",
    tfjsonpath.New("partitions").AtSliceIndex(0).AtMapKey("modified"),
    knownvalue.NotNull(),  // Could be true or false
)
```

**Resource Boolean Attributes with Known Values**:
```go
// For resources with known test values
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("dhcp_enabled"),
    knownvalue.Bool(true),  // We know it's true from test config
)
```

---

## Common Mistakes to Avoid

❌ **Keeping quotes around boolean values**:
```go
// WRONG - Boolean values must not be quoted
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("dhcp_enabled"),
    knownvalue.Bool("true"),  // ❌ Won't compile - remove quotes!
)
```

❌ **Using StringExact for boolean attributes**:
```go
// WRONG - Loses type safety
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("dhcp_enabled"),
    knownvalue.StringExact("true"),  // ❌ Wrong matcher type!
)
```

❌ **Using uppercase TRUE/FALSE**:
```go
// WRONG - Go uses lowercase true/false
knownvalue.Bool(TRUE)  // ❌ Won't compile
knownvalue.Bool(False) // ❌ Won't compile
```

❌ **Using 1/0 instead of true/false**:
```go
// WRONG - Don't use numeric representation
knownvalue.Bool(1)  // ❌ Type error
```

---

## Type Safety Benefit

### Without Type Safety (Legacy)
```go
// API bug: Returns string "true" instead of boolean true
// Test PASSES (string comparison works)
resource.TestCheckResourceAttr("resource", "enabled", "true")  // ✅ False positive
```

### With Type Safety (Modern)
```go
// API bug: Returns string "true" instead of boolean true
// Test FAILS with clear error
statecheck.ExpectKnownValue(
    "resource",
    tfjsonpath.New("enabled"),
    knownvalue.Bool(true),  // ❌ Fails: "expected bool, got string"
)
```

**Benefit**: Catches schema/API type mismatches at test time!

---

## Verification

### Visual Inspection
- ✅ Boolean values have no quotes: `true` not `"true"`
- ✅ Uses `Bool()` matcher
- ✅ Lowercase: `true`/`false` not `TRUE`/`FALSE`

### Test Execution
```bash
# If test fails with type error like:
# "expected types.BoolType, got types.StringType"
# This indicates BCM API bug - returning string instead of bool

# Example failure:
# Error: value type mismatch: expected bool, got string "true"

# Resolution: Check resource schema - should use types.Bool not types.String
```

---

## Known Boolean Attributes in Target Files

### resource_cmnet_network_test.go
- `dhcp_enabled`: DHCP enabled flag
  - Legacy: `"true"` (string)
  - Modern: `true` (bool)
  - Occurrences: 1 (line 75)

### data_source_cmpart_partitions_test.go (Computed Booleans)
- `modified`: Partition modified flag → `NotNull()` (unknown value)
- `to_be_removed`: Deletion flag → `NotNull()` (unknown value)

**Total Boolean Transformations**: 1 known value + 2 computed (use NotNull)

---

## Decision Matrix: Bool(value) vs NotNull()

| Scenario | Matcher | Example |
|----------|---------|---------|
| Resource attribute with known test value | `Bool(true/false)` | `dhcp_enabled = true` in config |
| Data source computed boolean | `NotNull()` | `modified` field from API |
| Resource computed boolean | `NotNull()` | `is_default` calculated by backend |

---

## Applicable To

- **File**: resource_cmnet_network_test.go (dhcp_enabled)
- **Attributes**: All boolean-type attributes with known test values
- **Test Functions**: 1 boolean assertion in target files

---

**Contract Status**: ✅ Complete - Ready for Implementation
