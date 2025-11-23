# Test Modernization Workflow

Step-by-step process for modernizing Terraform provider tests to terraform-plugin-testing v1.13.3+ patterns.

## Overview

Modernization follows a systematic approach:
1. **Analyze** - Identify gaps using automation
2. **Prioritize** - Focus on high-impact changes
3. **Apply** - Use guided patterns for code changes
4. **Verify** - Compile and validate changes
5. **Test** - Run acceptance tests to ensure correctness

## Phase 1: Gap Analysis (Automated)

**Tool**: `scripts/analyze_gap.py`

```bash
python scripts/analyze_gap.py ./internal/provider/ --output gap_analysis.md
```

**Output**:
- Legacy pattern count by file
- Missing tests (drift, import, idempotency)
- Modern pattern adoption statistics
- Prioritized recommendation list

**Review the report to understand**:
- Which files need the most work
- What patterns are missing
- Overall modernization progress

## Phase 2: Prioritization

**Focus on high-impact changes first**:

### Priority 1 (Critical)
- **Missing drift detection tests** - Required for production readiness
- **Missing import tests** - Essential for resource lifecycle
- **Heavy legacy usage** (>20 legacy calls per file)

### Priority 2 (Important)
- **Missing idempotency checks** - Validates stable state
- **Moderate legacy usage** (5-20 legacy calls per file)
- **Mixed patterns** (legacy + modern in same file)

### Priority 3 (Cleanup)
- **Light legacy usage** (<5 legacy calls per file)
- **Documentation improvements**
- **Test organization**

## Phase 3: Pattern Application (Guided)

For each file, apply patterns in this order:

### Step 1: Add Missing Tests

**If missing drift detection test**:
1. See `pattern_templates.md` → "Drift Detection Test"
2. Choose field(s) to test for drift
3. Implement three-step pattern: Create → Modify Externally → Verify Drift → Restore
4. Use BCM API helpers: `createTestBCMClient()`, `getResourceUUIDByName()`

**If missing import test**:
1. See `pattern_templates.md` → "Import Test Step"
2. Add after Create step
3. Use `ImportState: true` and `ImportStateVerify: true`
4. Track ID consistency with `compareID.AddStateValue()`

**If missing idempotency checks**:
1. See `pattern_templates.md` → "Idempotency Verification"
2. Add after Create step
3. Add after Update steps
4. Use `plancheck.ExpectEmptyPlan()` in `ConfigPlanChecks.PreApply`

### Step 2: Convert Legacy Checks to Modern State Checks

**Pattern**: Replace `Check: resource.ComposeAggregateTestCheckFunc()` with `ConfigStateChecks`

**Before**:
```go
{
    Config: testAccResourceConfig(name),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttr("bcm_resource.test", "name", "expected"),
        resource.TestCheckResourceAttr("bcm_resource.test", "enabled", "true"),
    ),
}
```

**After**:
```go
{
    Config: testAccResourceConfig(name),
    ConfigStateChecks: []statecheck.StateCheck{
        statecheck.ExpectKnownValue(
            "bcm_resource.test",
            tfjsonpath.New("name"),
            knownvalue.StringExact("expected"),
        ),
        statecheck.ExpectKnownValue(
            "bcm_resource.test",
            tfjsonpath.New("enabled"),
            knownvalue.Bool(true),
        ),
    },
}
```

**Type Mapping** (see `pattern_templates.md` for complete reference):
- String → `knownvalue.StringExact()`
- Bool → `knownvalue.Bool()`
- Int64 → `knownvalue.Int64Exact()`
- Computed (UUID, ID) → `knownvalue.NotNull()`
- List size → `knownvalue.ListSizeExact()`

### Step 3: Add Required Imports

Ensure file has modern imports:

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

Add additional imports if using:
- BCM API: `"encoding/json"`, `"time"`, `"context"`
- Drift tests: All BCM API imports

## Phase 4: Verification (Automated)

**Tool**: `scripts/verify_compilation.sh`

```bash
./scripts/verify_compilation.sh ./internal/provider/
```

**Validates**:
- Go syntax correctness
- No compilation errors
- Import completeness

**If compilation fails**:
1. Review error messages for missing imports
2. Check for unclosed blocks or syntax errors
3. Verify all `ConfigStateChecks` and `ConfigPlanChecks` are valid
4. Re-run after fixes

## Phase 5: Testing

### Compile Check (Fast)
```bash
# Quick syntax validation
go test -c ./internal/provider/ -o /tmp/provider_tests
```

### Single Test (Medium)
```bash
# Run specific modernized test
TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run "^TestAccResource_Specific$"
```

### Full Test Suite (Slow)
```bash
# Run all tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

**Expected results**:
- All tests pass ✅
- No unexpected plan diffs
- Resources properly created/updated/destroyed

## Phase 6: Documentation & Reporting

1. **Re-run gap analysis** to verify progress:
   ```bash
   python scripts/analyze_gap.py ./internal/provider/ --output final_analysis.md
   ```

2. **Compare before/after**:
   - Legacy check reduction
   - Modern pattern adoption increase
   - Test coverage improvements

3. **Update project documentation** if needed:
   - CLAUDE.md patterns section
   - Test helper documentation
   - BCM-specific quirks discovered

## Common Pitfalls

### 1. Forgetting Required Imports
**Symptom**: `undefined: statecheck` or `undefined: plancheck`
**Fix**: Add imports from Phase 3, Step 3

### 2. Wrong knownvalue Matcher
**Symptom**: Type mismatch errors in tests
**Fix**: Match BCM attribute types (see `pattern_templates.md`)

### 3. Missing ConfigStateChecks Block
**Symptom**: Empty test step with no validation
**Fix**: Add `ConfigStateChecks: []statecheck.StateCheck{...}`

### 4. Duplicate Validation (Legacy + Modern)
**Symptom**: Same assertion in both `Check` and `ConfigStateChecks`
**Fix**: Remove `Check` block entirely, keep only `ConfigStateChecks`

### 5. Incorrect BCM Field Mapping
**Symptom**: Drift test doesn't detect changes
**Fix**: Use snake_case → camelCase mapping (see `bcm_specifics.md`)

## Completion Criteria

A fully modernized test file has:
- ✅ Zero legacy `Check` blocks
- ✅ All state assertions use `statecheck.ExpectKnownValue()`
- ✅ Idempotency checks after Create and Update steps
- ✅ Import test with `ImportStateVerify`
- ✅ Drift detection test (resources only)
- ✅ ID consistency tracking with `CompareValue`
- ✅ All tests compile and pass
- ✅ Modern imports present

## Time Estimates

- **Simple file** (1-2 tests, <10 legacy checks): 15-30 minutes
- **Medium file** (3-5 tests, 10-30 legacy checks): 30-60 minutes
- **Complex file** (>5 tests, >30 legacy checks): 1-2 hours

**Total project** (7 resource files, 7 data source files): 4-8 hours spread across multiple sessions.
