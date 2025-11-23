# Contract 01: String Attribute Transformation

## Purpose

Transform legacy string attribute assertions from `resource.TestCheckResourceAttr()` to modern `statecheck.ExpectKnownValue()` with `knownvalue.StringExact()` matcher.

---

## Before (Legacy Pattern)

```go
{
    Config: testAccCMNetNetworkConfigBasic(networkName),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
        resource.TestCheckResourceAttr("bcm_cmnet_network.test", "domain_name", "cluster.local"),
        resource.TestCheckResourceAttr("bcm_cmnet_network.test", "notes", "Test network"),
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
            tfjsonpath.New("name"),
            knownvalue.StringExact(networkName),
        ),
        statecheck.ExpectKnownValue(
            "bcm_cmnet_network.test",
            tfjsonpath.New("domain_name"),
            knownvalue.StringExact("cluster.local"),
        ),
        statecheck.ExpectKnownValue(
            "bcm_cmnet_network.test",
            tfjsonpath.New("notes"),
            knownvalue.StringExact("Test network"),
        ),
    },
},
```

---

## Key Changes

1. **Field Replacement**: `Check:` → `ConfigStateChecks:`
2. **Type Change**: `resource.ComposeAggregateTestCheckFunc(...)` → `[]statecheck.StateCheck{...}`
3. **Assertion Format**: Three-argument format with explicit path and matcher
4. **Path Specification**: Use `tfjsonpath.New("attribute_name")` for attribute paths
5. **Type Matcher**: Use `knownvalue.StringExact(value)` for exact string matching
6. **Remove Wrapper**: No `ComposeAggregateTestCheckFunc()` wrapper needed

---

## Required Imports

```go
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
```

---

## Examples from Target Files

### Example 1: Variable String (resource_cmnet_network_test.go)

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName)
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("name"),
    knownvalue.StringExact(networkName),
)
```

### Example 2: Literal String (resource_cmnet_network_test.go)

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmnet_network.test", "subnet", "192.168.100.0/24")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("subnet"),
    knownvalue.StringExact("192.168.100.0/24"),
)
```

### Example 3: Kubernetes Version (resource_cmkube_cluster_test.go)

**Before**:
```go
resource.TestCheckResourceAttr("bcm_cmkube_cluster.test", "version", "1.28.0")
```

**After**:
```go
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("version"),
    knownvalue.StringExact("1.28.0"),
)
```

---

## Common Mistakes to Avoid

❌ **Keeping ComposeAggregateTestCheckFunc wrapper**:
```go
// WRONG - Don't use wrapper with ConfigStateChecks
ConfigStateChecks: []statecheck.StateCheck{
    resource.ComposeAggregateTestCheckFunc(  // ❌ Wrong!
        statecheck.ExpectKnownValue(...),
    ),
},
```

❌ **Using Check field instead of ConfigStateChecks**:
```go
// WRONG - Don't mix legacy and modern
Check: statecheck.ExpectKnownValue(...)  // ❌ Won't compile
```

❌ **Missing tfjsonpath.New() wrapper**:
```go
// WRONG - Path must be wrapped
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    "name",  // ❌ Wrong! Needs tfjsonpath.New()
    knownvalue.StringExact(networkName),
)
```

---

## Verification

### Visual Inspection
- ✅ `Check:` field removed
- ✅ `ConfigStateChecks:` field added
- ✅ All attribute names wrapped in `tfjsonpath.New()`
- ✅ All string values wrapped in `knownvalue.StringExact()`

### Grep Verification
```bash
# Should return 0 matches after transformation
grep -n "TestCheckResourceAttr" internal/provider/resource_cmnet_network_test.go | \
  grep -v "//" | wc -l
```

---

## Applicable To

- **File**: resource_cmnet_network_test.go (primary examples)
- **File**: resource_cmkube_cluster_test.go (version, cni_plugin, load_balancer_mode, management_network)
- **Attributes**: All string-type attributes (name, notes, domain_name, subnet, gateway, version, etc.)
- **Test Functions**: 14 string assertions in P1 file alone

---

**Contract Status**: ✅ Complete - Ready for Implementation
