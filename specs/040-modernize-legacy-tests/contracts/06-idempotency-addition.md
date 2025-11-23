# Contract 06: Idempotency Verification Addition

## Purpose

Add idempotency verification steps after Create and Update operations using `plancheck.ExpectEmptyPlan()` to ensure resources don't generate spurious diffs.

---

## Pattern: Add Idempotency Step After Create

**Before** (Missing Idempotency):
```go
Steps: []resource.TestStep{
    // Create resource
    {
        Config: testAccCMNetNetworkConfigComplete(networkName),
        ConfigStateChecks: []statecheck.StateCheck{
            statecheck.ExpectKnownValue(...),
        },
    },
    // Next step (Import, Update, or test end)
    {
        // ...
    },
},
```

**After** (With Idempotency):
```go
Steps: []resource.TestStep{
    // Create resource
    {
        Config: testAccCMNetNetworkConfigComplete(networkName),
        ConfigStateChecks: []statecheck.StateCheck{
            statecheck.ExpectKnownValue(...),
        },
    },
    // Idempotency check - NEW STEP
    {
        Config: testAccCMNetNetworkConfigComplete(networkName),  // Same config
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectEmptyPlan(),
            },
        },
    },
    // Next step (Import, Update, or test end)
    {
        // ...
    },
},
```

---

## Pattern: Add Idempotency Step After Update

**Before** (Missing Idempotency):
```go
Steps: []resource.TestStep{
    // ... earlier steps ...
    // Update resource
    {
        Config: testAccCMNetNetworkConfigUpdate(networkName),
        ConfigStateChecks: []statecheck.StateCheck{
            statecheck.ExpectKnownValue(...),
        },
    },
    // Test ends here - NO idempotency check
},
```

**After** (With Idempotency):
```go
Steps: []resource.TestStep{
    // ... earlier steps ...
    // Update resource
    {
        Config: testAccCMNetNetworkConfigUpdate(networkName),
        ConfigStateChecks: []statecheck.StateCheck{
            statecheck.ExpectKnownValue(...),
        },
    },
    // Idempotency check after Update - NEW STEP
    {
        Config: testAccCMNetNetworkConfigUpdate(networkName),  // Same config
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectEmptyPlan(),
            },
        },
    },
},
```

---

## Key Elements

1. **Same Config**: Idempotency step uses IDENTICAL config to previous step
2. **No State Checks**: Idempotency step has NO ConfigStateChecks (only ConfigPlanChecks)
3. **PreApply Check**: Uses `PreApply` not `PostApply` (checks plan before execution)
4. **ExpectEmptyPlan**: Verifies Terraform shows "No changes" when re-applying same config

---

## Required Imports

```go
"github.com/hashicorp/terraform-plugin-testing/helper/resource"
"github.com/hashicorp/terraform-plugin-testing/plancheck"
```

**Note**: `plancheck` import already present in `resource_cmnet_network_test.go` (line 14)

---

## Examples from Target Files

### Example 1: TestAccCMNetNetwork_Complete (resource_cmnet_network_test.go)

**Current State**: Lines 56-82 - Single Create step, no idempotency
**Required Change**: Add idempotency step after Create

**Insertion Point**: After line 79 (end of Create step), before closing `}`

**Code to Insert**:
```go
// Idempotency check after Create
{
    Config: testAccCMNetNetworkConfigComplete(networkName),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

---

### Example 2: TestAccCMNetNetwork_Update (resource_cmnet_network_test.go)

**Current State**: Lines 84-110 - Create + Update, no idempotency after either
**Required Changes**:
1. Add idempotency after Create (line 98 insertion)
2. Add idempotency after Update (line 107 insertion)

**Insertion Point 1** (After Create): After line 97, before Update step

**Code to Insert**:
```go
// Idempotency check after Create
{
    Config: testAccCMNetNetworkConfigBasic(networkName),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

**Insertion Point 2** (After Update): After line 106 (end of Update step)

**Code to Insert**:
```go
// Idempotency check after Update
{
    Config: testAccCMNetNetworkConfigUpdate(networkName),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

---

## When to Add Idempotency Checks

### Always Add After:
- ✅ Create steps (first step in test)
- ✅ Update steps (any config change)
- ✅ Any step that modifies resource state

### Don't Add After:
- ❌ Import steps (Import doesn't modify state)
- ❌ Steps with `PreConfig` (drift detection tests - intentional changes)
- ❌ Steps that already have `ExpectEmptyPlan` (already idempotent)
- ❌ Validation tests using `ExpectError`

---

## Why Idempotency Matters

### Without Idempotency Check
```go
// Resource has bug: Always marks "notes" as changed
// Test PASSES (only checks Create, not re-apply)
{
    Config: testConfig(name),
    ConfigStateChecks: [...],  // ✅ Passes - notes set correctly
}
```

### With Idempotency Check
```go
// Resource has bug: Always marks "notes" as changed
// Test FAILS (detects spurious diff on re-apply)
{
    Config: testConfig(name),
    ConfigStateChecks: [...],  // ✅ Passes
},
{
    Config: testConfig(name),  // Same config
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),  // ❌ Fails - notes shows diff!
        },
    },
},
```

**Benefit**: Catches bugs where resources generate unexpected diffs on refresh/re-apply.

---

## Common Mistakes to Avoid

❌ **Different config in idempotency step**:
```go
// WRONG - Config must be IDENTICAL to previous step
{
    Config: testConfig("name1"),
    ConfigStateChecks: [...],
},
{
    Config: testConfig("name2"),  // ❌ Different config!
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),  // Will fail - names differ
        },
    },
},
```

❌ **Using PostApply instead of PreApply**:
```go
// WRONG - Use PreApply to check plan before execution
ConfigPlanChecks: resource.ConfigPlanChecks{
    PostApply: []plancheck.PlanCheck{  // ❌ Wrong timing
        plancheck.ExpectEmptyPlan(),
    },
},
```

❌ **Including ConfigStateChecks in idempotency step**:
```go
// UNNECESSARY - Idempotency step only needs plan check
{
    Config: testConfig(name),
    ConfigStateChecks: []statecheck.StateCheck{  // ⚠️ Redundant
        statecheck.ExpectKnownValue(...),
    },
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

---

## Idempotency in P2/P3 Files

**resource_cmkube_cluster_test.go**:
- ✅ Already has comprehensive idempotency coverage
- All test functions have idempotency steps
- No changes needed

**data_source_cmpart_partitions_test.go**:
- N/A - Data sources are read-only, no Create/Update operations
- Idempotency not applicable

---

## Verification

### Visual Inspection
- ✅ Every Create step followed by idempotency step
- ✅ Every Update step followed by idempotency step
- ✅ Idempotency steps use identical config
- ✅ Idempotency steps use `ExpectEmptyPlan()`

### Test Output
```bash
# Successful idempotency check output:
=== RUN   TestAccCMNetNetwork_Complete
=== PAUSE TestAccCMNetNetwork_Complete
=== CONT  TestAccCMNetNetwork_Complete
--- PASS: TestAccCMNetNetwork_Complete (12.34s)
    # Step 1: Create shows changes
    # Step 2: Idempotency shows "No changes. Your infrastructure matches the configuration."
```

---

## Target Files Needing Idempotency

### resource_cmnet_network_test.go

**TestAccCMNetNetwork_Basic** (lines 18-54):
- ✅ Already has idempotency (lines 37-44)
- No changes needed

**TestAccCMNetNetwork_Complete** (lines 56-82):
- ❌ Missing idempotency after Create
- **Action**: Add 1 idempotency step

**TestAccCMNetNetwork_Update** (lines 84-110):
- ❌ Missing idempotency after Create (line 98)
- ❌ Missing idempotency after Update (line 107)
- **Action**: Add 2 idempotency steps

**Total Additions**: 3 idempotency steps

---

## Summary

| File | Function | Idempotency Gaps | Action |
|------|----------|------------------|--------|
| resource_cmnet_network_test.go | TestAccCMNetNetwork_Basic | None | ✅ Already has |
| resource_cmnet_network_test.go | TestAccCMNetNetwork_Complete | After Create | ➕ Add 1 step |
| resource_cmnet_network_test.go | TestAccCMNetNetwork_Update | After Create, After Update | ➕ Add 2 steps |
| resource_cmkube_cluster_test.go | All functions | None | ✅ Already complete |
| data_source_cmpart_partitions_test.go | All functions | N/A | N/A (data sources) |

---

**Contract Status**: ✅ Complete - Ready for Implementation
