# Contract 07: ID Consistency Tracking Addition

## Purpose

Add ID consistency tracking using `statecheck.CompareValue()` to verify resource IDs remain stable across Create, Import, and Update operations.

---

## Pattern: Complete ID Tracking Implementation

**Before** (No ID Tracking):
```go
func TestAccCMNetNetwork_Basic(t *testing.T) {
    networkName := generateUniqueTestName("test-network")

    resource.Test(t, resource.TestCase{
        // ... test case config ...
        Steps: []resource.TestStep{
            // Create
            {
                Config: testConfig(networkName),
                ConfigStateChecks: []statecheck.StateCheck{
                    // State checks but NO ID tracking
                    statecheck.ExpectKnownValue(...),
                },
            },
            // Import
            {
                ResourceName:      "bcm_cmnet_network.test",
                ImportState:       true,
                ImportStateVerify: true,
                // NO ID tracking in Import step
            },
        },
    })
}
```

**After** (With ID Tracking):
```go
func TestAccCMNetNetwork_Basic(t *testing.T) {
    networkName := generateUniqueTestName("test-network")

    // Initialize ID tracker ONCE at function start - NEW
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        // ... test case config ...
        Steps: []resource.TestStep{
            // Create
            {
                Config: testConfig(networkName),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(...),
                    // Add ID tracking - NEW
                    compareID.AddStateValue(
                        "bcm_cmnet_network.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Import
            {
                ResourceName:      "bcm_cmnet_network.test",
                ImportState:       true,
                ImportStateVerify: true,
                // Add ID tracking to Import - NEW
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue(
                        "bcm_cmnet_network.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
        },
    })
}
```

---

## Key Elements

1. **Single Initialization**: `compareID := statecheck.CompareValue(compare.ValuesSame())` ONCE at function start
2. **Add to ALL Operations**: Track ID in Create, Import, and Update steps
3. **Same Resource Path**: Use consistent resource address across all tracking calls
4. **Automatic Comparison**: Framework compares captured IDs, fails if any differ

---

## Required Imports

```go
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/compare"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
```

**Note**: resource_cmnet_network_test.go needs these imports added

---

## Example from Target File

### TestAccCMNetNetwork_Basic (resource_cmnet_network_test.go)

**Current State**: Lines 18-54
- ✅ Has Import step (line 46-51)
- ❌ NO ID tracking

**Required Changes**:

1. **Add compareID initialization** (insert after line 19):
```go
networkName := generateUniqueTestName("test-network")

// Initialize ID tracker for consistency verification across operations
compareID := statecheck.CompareValue(compare.ValuesSame())
```

2. **Add ID tracking to Create step** (add to ConfigStateChecks at line 29-34):
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(...),  // Existing checks
    statecheck.ExpectKnownValue(...),
    statecheck.ExpectKnownValue(...),
    statecheck.ExpectKnownValue(...),
    // Add ID tracking - NEW
    compareID.AddStateValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("id"),
    ),
},
```

3. **Add ConfigStateChecks to Import step** (modify lines 46-51):
```go
// ImportState testing
{
    Config:            testAccCMNetNetworkConfigBasic(networkName),
    ResourceName:      "bcm_cmnet_network.test",
    ImportState:       true,
    ImportStateVerify: true,
    // Add ID tracking to Import - NEW
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue(
            "bcm_cmnet_network.test",
            tfjsonpath.New("id"),
        ),
    },
},
```

---

## How ID Tracking Works

### Step 1: Create - Capture Baseline ID
```go
// First call to compareID.AddStateValue() captures the ID
{
    Config: testConfig(name),
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("resource", tfjsonpath.New("id")),
        // Stores ID internally as baseline
    },
},
```

### Step 2: Import - Verify ID Unchanged
```go
// Second call compares imported ID to baseline
{
    ImportState: true,
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("resource", tfjsonpath.New("id")),
        // Compares to baseline, fails if different
    },
},
```

### Step 3: Update - Verify ID Unchanged
```go
// Third call (if present) compares updated ID to baseline
{
    Config: testConfigUpdated(name),
    ConfigStateChecks: []statecheck.StateCheck{
        compareID.AddStateValue("resource", tfjsonpath.New("id")),
        // Compares to baseline, fails if different
    },
},
```

---

## When to Add ID Tracking

### Always Add When:
- ✅ Test has Import step (verify import preserves ID)
- ✅ Test has Update step AND Import step (verify both preserve ID)
- ✅ Resource is expected to have stable ID across operations

### Don't Add When:
- ❌ Test has no Import step (ID tracking less valuable without Import)
- ❌ Data source tests (data sources don't have import or stable IDs)
- ❌ Validation tests using ExpectError (no resource created)
- ❌ Resource intentionally recreates on update (ID expected to change)

---

## Why ID Tracking Matters

### Without ID Tracking
```go
// Resource has bug: Import creates new resource instead of importing existing
// Test PASSES (doesn't verify ID consistency)
Steps: []resource.TestStep{
    {Config: testConfig(name)},  // Creates resource with id="abc-123"
    {ImportState: true},         // Creates NEW resource with id="def-456" ❌ BUG!
}
```

### With ID Tracking
```go
// Resource has bug: Import creates new resource instead of importing existing
// Test FAILS (detects ID changed)
compareID := statecheck.CompareValue(compare.ValuesSame())

Steps: []resource.TestStep{
    {
        Config: testConfig(name),
        ConfigStateChecks: []statecheck.StateCheck{
            compareID.AddStateValue("resource", tfjsonpath.New("id")),  // Captures id="abc-123"
        },
    },
    {
        ImportState: true,
        ConfigStateChecks: []statecheck.StateCheck{
            compareID.AddStateValue("resource", tfjsonpath.New("id")),  // ❌ Fails: id="def-456" differs!
        },
    },
}
```

**Benefit**: Catches bugs where resource IDs unexpectedly change during Import or Update.

---

## Common Mistakes to Avoid

❌ **Multiple compareID initializations**:
```go
// WRONG - Initialize ONCE per test function
func TestAcc(t *testing.T) {
    compareID := statecheck.CompareValue(compare.ValuesSame())  // ✅ Good

    Steps: []resource.TestStep{
        {
            // ... step 1 ...
            compareID := statecheck.CompareValue(compare.ValuesSame())  // ❌ Don't reinitialize!
        },
    }
}
```

❌ **Tracking different attributes**:
```go
// WRONG - Must track SAME attribute across all steps
compareID := statecheck.CompareValue(compare.ValuesSame())

Steps: []resource.TestStep{
    {
        ConfigStateChecks: []statecheck.StateCheck{
            compareID.AddStateValue("resource", tfjsonpath.New("id")),  // Tracks "id"
        },
    },
    {
        ConfigStateChecks: []statecheck.StateCheck{
            compareID.AddStateValue("resource", tfjsonpath.New("uuid")),  // ❌ Tracks "uuid" instead!
        },
    },
}
```

❌ **Forgetting to add to Import step**:
```go
// INCOMPLETE - Import step missing ID tracking
Steps: []resource.TestStep{
    {
        Config: testConfig(name),
        ConfigStateChecks: []statecheck.StateCheck{
            compareID.AddStateValue("resource", tfjsonpath.New("id")),  // ✅ Create has it
        },
    },
    {
        ImportState: true,
        // ❌ Missing ConfigStateChecks with compareID!
    },
}
```

---

## ID Tracking in P2/P3 Files

**resource_cmkube_cluster_test.go**:
- ✅ TestAccCMKubeClusterResource_Basic already has complete ID tracking (lines 168, 199-202, 237-240)
- Other test functions don't have Import steps, so ID tracking not needed

**data_source_cmpart_partitions_test.go**:
- N/A - Data sources don't have Import or stable resource IDs

---

## Verification

### Visual Inspection
- ✅ `compareID` initialized once at function start
- ✅ `compareID.AddStateValue()` in Create step ConfigStateChecks
- ✅ `compareID.AddStateValue()` in Import step ConfigStateChecks
- ✅ Same resource address and attribute path in all tracking calls

### Test Output
```bash
# Successful ID tracking output:
=== RUN   TestAccCMNetNetwork_Basic
--- PASS: TestAccCMNetNetwork_Basic (15.23s)
    # Step 1: Create - ID captured: "net-abc123"
    # Step 2: Idempotency - ID same
    # Step 3: Import - ID same: "net-abc123"

# Failed ID tracking output (bug detected):
=== RUN   TestAccCMNetNetwork_Basic
--- FAIL: TestAccCMNetNetwork_Basic (10.45s)
    Error: ID mismatch detected
        Expected: "net-abc123"
        Got:      "net-def456"
```

---

## Target Files Needing ID Tracking

### resource_cmnet_network_test.go

**TestAccCMNetNetwork_Basic** (lines 18-54):
- ✅ Has Import step (line 46-51)
- ❌ NO ID tracking
- **Action**: Add compareID initialization + track in Create + track in Import

**TestAccCMNetNetwork_Complete** (lines 56-82):
- ❌ NO Import step
- **Action**: None (ID tracking not needed without Import)

**TestAccCMNetNetwork_Update** (lines 84-110):
- ❌ NO Import step
- **Action**: None (ID tracking not needed without Import)

**Total Additions**: 1 test function needs ID tracking (3 code changes)

---

## Alternative: Track UUID Instead of ID

Some resources may want to track UUID field instead:

```go
// Track UUID if it's the primary stable identifier
compareID := statecheck.CompareValue(compare.ValuesSame())

ConfigStateChecks: []statecheck.StateCheck{
    compareID.AddStateValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("uuid"),  // Track uuid instead of id
    ),
}
```

**Decision**: For BCM resources, track `id` field (primary identifier in Terraform state).

---

## Summary

| File | Function | Has Import? | ID Tracking Status | Action |
|------|----------|-------------|-------------------|--------|
| resource_cmnet_network_test.go | TestAccCMNetNetwork_Basic | ✅ Yes | ❌ Missing | ➕ Add tracking |
| resource_cmnet_network_test.go | TestAccCMNetNetwork_Complete | ❌ No | N/A | None |
| resource_cmnet_network_test.go | TestAccCMNetNetwork_Update | ❌ No | N/A | None |
| resource_cmkube_cluster_test.go | TestAccCMKubeClusterResource_Basic | ✅ Yes | ✅ Has tracking | None |
| data_source_cmpart_partitions_test.go | All functions | N/A | N/A | None |

---

**Contract Status**: ✅ Complete - Ready for Implementation
