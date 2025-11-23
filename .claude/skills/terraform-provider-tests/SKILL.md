---
name: terraform-provider-tests
description: Analyze and improve Terraform provider test coverage using terraform-plugin-testing v1.13.3+ patterns. Use when (1) analyzing test coverage gaps, (2) adding missing tests (drift detection, import, idempotency), (3) converting legacy patterns to modern state checks, (4) tracking optional field coverage, or (5) verifying test quality. Supports automated coverage analysis and guided pattern improvements.
---

# Terraform Provider Tests

Analyze test coverage and improve Terraform provider acceptance tests using modern patterns from terraform-plugin-testing v1.13.3+.

## Quick Start

### Common Usage Scenarios

**Analyze test modernization gaps**:
```bash
python3 scripts/analyze_gap.py ./internal/provider/ --output tf_provider_tests_gap_$(date +%Y%m%d_%H%M%S).md
```

**Analyze a specific test file**:
"Analyze resource_example_test.go for coverage gaps"

**Verify compilation after changes**:
```bash
./scripts/verify_compilation.sh ./internal/provider/
```

### Recommended Naming Convention

All gap analysis reports should use the `tf_provider_tests_*` naming pattern:

| Report Type | Filename Pattern | Example |
|-------------|------------------|---------|
| Initial gap analysis | `tf_provider_tests_gap_YYYYMMDD_HHMMSS.md` | `tf_provider_tests_gap_20251123_225128.md` |
| Final analysis | `tf_provider_tests_final_YYYYMMDD_HHMMSS.md` | `tf_provider_tests_final_20251123_230145.md` |
| One-time analysis | `tf_provider_tests_gap.md` | `tf_provider_tests_gap.md` |

**Timestamp format**: `$(date +%Y%m%d_%H%M%S)` generates `YYYYMMDD_HHMMSS`

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
python3 scripts/analyze_gap.py <test_directory> [--output report.md]
```

**Example** (recommended naming pattern):
```bash
python3 scripts/analyze_gap.py ./internal/provider/ --output tf_provider_tests_gap_$(date +%Y%m%d_%H%M%S).md
```

**Simple filename** (for one-time analysis):
```bash
python3 scripts/analyze_gap.py ./internal/provider/ --output tf_provider_tests_gap.md
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
    resource.TestCheckResourceAttr("example_resource.test", "name", "expected"),
),
```

**After**:
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "example_resource.test",
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

### Parallel Test Execution (Recommended)

**Tool**: `scripts/run_tests_parallel.sh`

Run acceptance tests concurrently per file for faster execution.

**Usage**:
```bash
# Run all acceptance tests with 4 concurrent files
./scripts/run_tests_parallel.sh

# Run only resource tests with higher concurrency
./scripts/run_tests_parallel.sh --resources-only -c 8

# Run only data source tests
./scripts/run_tests_parallel.sh --data-sources-only

# Run tests matching specific pattern
./scripts/run_tests_parallel.sh -p "TestAccCMPartSoftwareImage"

# Run tests from specific file
./scripts/run_tests_parallel.sh -f resource_cmpart_softwareimage_test.go

# Verbose output with detailed test logs
./scripts/run_tests_parallel.sh --verbose
```

**Options**:
- `-d, --dir DIR` - Test directory (default: ./internal/provider)
- `-p, --pattern PATTERN` - Test pattern to match (default: TestAcc)
- `-c, --concurrency N` - Max concurrent test files (default: 4)
- `-t, --timeout DURATION` - Timeout per test file (default: 30m)
- `-f, --file FILE` - Run only tests from specific file
- `--resources-only` - Run only resource tests
- `--data-sources-only` - Run only data source tests
- `--verbose` - Show detailed test output
- `--no-color` - Disable colored output

**Benefits**:
- ⚡ Faster execution (4x-8x speedup with proper concurrency)
- 📊 Per-file progress tracking
- 🎯 Aggregated summary with pass/fail counts
- 🔍 Automatic failure highlighting

### Single Test (Sequential)
```bash
TF_ACC=1 go test -v -timeout 30m ./internal/provider/ -run "^TestAccResource_Specific$"
```

### Full Suite (Sequential)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

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
→ Match types correctly: String→StringExact, Bool→Bool, Int64→Int64Exact

**Duplicate validation**
→ Remove `Check` block, keep only `ConfigStateChecks`

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
- ✅ Complex operations (drift tests, API integration)
- ✅ Decision making (prioritization, test design)

This balance maintains control while automating tedious tasks.
