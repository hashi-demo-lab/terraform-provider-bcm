---
name: terraform-test-modernization
description: Automate and guide modernization of Terraform provider tests to terraform-plugin-testing v1.13.3+ patterns. Use when (1) modernizing legacy test files with resource.TestCheckResourceAttr to statecheck.ExpectKnownValue, (2) adding missing tests (drift detection, import, idempotency), (3) analyzing test coverage gaps, (4) converting Check blocks to ConfigStateChecks, or (5) verifying test compilation after changes. Supports both automated gap analysis and guided pattern application with BCM-specific patterns.
---

# Terraform Test Modernization

Systematic workflow for modernizing Terraform provider acceptance tests to modern patterns from terraform-plugin-testing v1.13.3+.

## Quick Start

### Common Usage Scenarios

**Analyze test modernization gaps**:
```bash
python scripts/analyze_gap.py ./internal/provider/ --output gap_analysis.md
```

**Modernize a specific test file**:
"Modernize resource_cmpart_softwareimage_test.go to use modern patterns"

**Verify compilation after changes**:
```bash
./scripts/verify_compilation.sh ./internal/provider/
```

## Workflow Decision Tree

**Starting a modernization project?**
→ Begin with [Gap Analysis](#phase-1-gap-analysis-automated)

**Have gap analysis report?**
→ Proceed to [Pattern Application](#phase-3-pattern-application-guided)

**Made code changes?**
→ Run [Verification](#phase-4-verification-automated)

**All changes complete?**
→ Complete [Testing](#phase-5-testing)

## Phase 1: Gap Analysis (Automated)

**Tool**: `scripts/analyze_gap.py`

Automatically scan test files and generate comprehensive gap analysis report.

### Usage

```bash
python scripts/analyze_gap.py <test_directory> [--output report.md]
```

**Example**:
```bash
python scripts/analyze_gap.py ./internal/provider/ --output gap_analysis.md
```

### What It Detects

- ✅ Legacy `Check: resource.TestCheckResourceAttr()` patterns
- ✅ Missing drift detection tests
- ✅ Missing import tests
- ✅ Missing idempotency checks
- ✅ Modern pattern adoption statistics
- ✅ Prioritized recommendations

### Output

Markdown report with:
- Executive summary with overall grade (A/B/C)
- File-by-file analysis with status
- Prioritized recommendations (High/Medium/Low)
- Modern pattern quick reference

**Review the report to understand**:
- Which files need the most work
- What patterns are missing
- Overall modernization progress

## Phase 2: Prioritization

Focus on high-impact changes first:

### Priority 1 (Critical) ⚠️
- Missing drift detection tests
- Missing import tests
- Heavy legacy usage (>20 calls per file)

### Priority 2 (Important) 📋
- Missing idempotency checks
- Moderate legacy usage (5-20 calls)
- Mixed patterns (legacy + modern)

### Priority 3 (Cleanup) 📝
- Light legacy usage (<5 calls)
- Documentation improvements

## Phase 3: Pattern Application (Guided)

For each file, apply patterns in this order:

### Step 1: Add Missing Tests

**Missing drift detection?**
→ See [references/pattern_templates.md → "Drift Detection Test"](#references)

**Missing import test?**
→ See [references/pattern_templates.md → "Import Test Step"](#references)

**Missing idempotency checks?**
→ See [references/pattern_templates.md → "Idempotency Verification"](#references)

### Step 2: Convert Legacy to Modern

Replace `Check: resource.ComposeAggregateTestCheckFunc()` with `ConfigStateChecks`.

**Before**:
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_resource.test", "name", "expected"),
),
```

**After**:
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_resource.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact("expected"),
    ),
},
```

**Type mapping** → See [references/pattern_templates.md](#references)

### Step 3: Add Required Imports

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

For drift tests, also add: `"context"`, `"encoding/json"`, `"time"`

## Phase 4: Verification (Automated)

**Tool**: `scripts/verify_compilation.sh`

Quick compilation check without running tests.

### Usage

```bash
./scripts/verify_compilation.sh <test_directory>
```

**Example**:
```bash
./scripts/verify_compilation.sh ./internal/provider/
```

### What It Validates

- ✅ Go syntax correctness
- ✅ No compilation errors
- ✅ Import completeness
- ✅ Statistics (file count, test count)

**If compilation fails**:
1. Review error messages
2. Check for missing imports
3. Verify block closures
4. Re-run after fixes

## Phase 5: Testing

### Quick Compile Check
```bash
go test -c ./internal/provider/ -o /tmp/provider_tests
```

### Single Test
```bash
TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run "^TestAccResource_Specific$"
```

### Full Suite
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

## BCM-Specific Patterns

**Critical**: BCM uses camelCase, Terraform uses snake_case.

**Field mapping examples**:
- `kernel_parameters` → `kernelParameters`
- `enable_sol` → `enableSol`
- `dhcp_enabled` → `dhcpEnabled`

**See [references/bcm_specifics.md](#references) for**:
- Complete field name mapping table
- BCM API entity structure
- Test helpers usage
- Drift detection examples

## Completion Criteria

A fully modernized test file has:
- ✅ Zero legacy `Check` blocks
- ✅ All state assertions use `statecheck.ExpectKnownValue()`
- ✅ Idempotency checks after Create and Update
- ✅ Import test with `ImportStateVerify`
- ✅ Drift detection test (resources only)
- ✅ ID consistency tracking with `CompareValue`
- ✅ All tests compile and pass

## Common Pitfalls

**Missing imports**
→ Add all required imports from Phase 3, Step 3

**Wrong knownvalue matcher**
→ Match BCM types: String→StringExact, Bool→Bool, Int64→Int64Exact

**Duplicate validation**
→ Remove `Check` block, keep only `ConfigStateChecks`

**Wrong BCM field name**
→ Use camelCase in API calls, snake_case in Terraform

## References

This skill includes comprehensive reference documentation:

### references/workflow.md
Complete step-by-step modernization workflow with detailed guidance for each phase, common pitfalls, and completion criteria.

**When to read**: For detailed phase-by-phase instructions.

### references/pattern_templates.md
Ready-to-use code templates for all modern patterns:
- Legacy to modern conversion examples
- Idempotency verification
- Import test steps
- Drift detection tests (complete template)
- ID consistency tracking
- knownvalue type matchers
- Complete test examples

**When to read**: When applying specific patterns to code.

### references/bcm_specifics.md
BCM-specific testing patterns and quirks:
- Field name mapping (snake_case ↔ camelCase)
- BCM API entity structure
- Test helper functions
- Eventual consistency patterns
- Complete drift detection example

**When to read**: When working with BCM provider tests or encountering BCM-specific errors.

### references/hashicorp_official.md
Consolidated HashiCorp official documentation:
- TestCase and TestStep structure
- State checks (statecheck package)
- Plan checks (plancheck package)
- Value comparers (compare package)
- Import mode testing
- Official testing patterns

**When to read**: For authoritative information on terraform-plugin-testing features.

## Time Estimates

- **Simple file** (1-2 tests, <10 legacy checks): 15-30 minutes
- **Medium file** (3-5 tests, 10-30 legacy checks): 30-60 minutes
- **Complex file** (>5 tests, >30 legacy checks): 1-2 hours

**Total project** (7 resource + 7 data source files): 4-8 hours across multiple sessions.

## Hybrid Approach

This skill uses a **hybrid approach**:

**Automation** for:
- ✅ Gap analysis (find legacy patterns, missing tests)
- ✅ Compilation verification (syntax checking)

**Guided assistance** for:
- ✅ Code changes (apply patterns with context)
- ✅ Complex operations (drift tests, BCM API calls)
- ✅ Decision making (prioritization, field mapping)

This balance maintains control while automating tedious tasks.
