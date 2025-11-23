# Implementation Plan: Modernize Legacy Testing Patterns

**Branch**: `040-modernize-legacy-tests` | **Date**: 2025-11-23 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/040-modernize-legacy-tests/spec.md`

## Summary

Modernize 3 test files (21 test functions total) from legacy `terraform-plugin-sdk` patterns to modern `terraform-plugin-testing` v1.13.3+ patterns. Convert 66 legacy assertions to type-safe state checks using `statecheck.ExpectKnownValue()`, add idempotency verification with `plancheck.ExpectEmptyPlan()`, and implement ID consistency tracking with `statecheck.CompareValue()` across all resource CRUD operations. This achieves 100% modern pattern adoption across the target test files.

**Technical Approach**: Sequential file-by-file transformation (P1 → P2 → P3) using established patterns from CLAUDE.md Modern Testing Patterns section. Each file modernization follows: analyze legacy patterns → transform assertions → add missing checks → verify with acceptance tests.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-testing v1.13.3, terraform-plugin-framework v1.16.1
**Storage**: N/A (test modernization only)
**Testing**: TF_ACC=1 go test -v -timeout 120m ./internal/provider/
**Target Platform**: Linux (BCM cluster integration)
**Project Type**: Terraform Provider (single codebase)
**Performance Goals**: Test execution time within 10% of baseline (<120m for full suite)
**Constraints**: Zero test regressions, 100% pass rate, BCM cluster availability required
**Scale/Scope**: 21 test functions, 66 legacy assertions, 3 test files

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**TDD Compliance Gates**:

1. **RED-GREEN-REFACTOR Cycle**: ✅ PASS - This is a refactoring task (tests already green, improving quality)
2. **Test-First Development**: ✅ PASS - Tests already exist and pass, we're enhancing them
3. **Acceptance Test Coverage**: ✅ PASS - All target files have comprehensive coverage, we're improving assertions
4. **No Implementation Changes**: ✅ PASS - Zero changes to resource/data source implementation code
5. **Backward Compatibility**: ✅ PASS - Test behavior unchanged, only assertion style modernized

**No violations** - Constitution gates passed.

## Project Structure

### Documentation (this feature)

```text
specs/040-modernize-legacy-tests/
├── plan.md              # This file (/speckit.plan output)
├── spec.md              # Feature specification
├── research.md          # Phase 0 output (patterns analysis)
├── data-model.md        # Phase 1 output (test structure mapping)
├── quickstart.md        # Phase 1 output (developer guide)
├── contracts/           # Phase 1 output (before/after examples)
└── tasks.md             # Phase 2 output (/speckit.tasks - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmnet_network_test.go           # P1: 3 functions, 18 legacy assertions
├── resource_cmkube_cluster_test.go          # P2: 10 functions, 38 legacy assertions
├── data_source_cmpart_partitions_test.go    # P3: 8 functions, 10 legacy assertions
└── test_helpers.go                          # Test utilities (no changes needed)

specs/040-modernize-legacy-tests/
└── [documentation artifacts from this plan]
```

**Structure Decision**: Standard Terraform Provider single-project structure. Test files are co-located with implementation in `internal/provider/`. Documentation artifacts in feature-specific specs directory follow Speckit conventions.

## Complexity Tracking

> No violations to track - all constitution gates passed.

---

## Phase 0: Outline & Research

### Objectives

1. Analyze current legacy patterns in each target file
2. Document modern pattern mappings from CLAUDE.md
3. Identify edge cases and special handling requirements
4. Validate test helper compatibility

### Research Tasks

#### Task 0.1: Legacy Pattern Inventory

**Target**: Catalog all legacy assertion patterns in use

**Method**: Grep analysis of target files

**Deliverable**: `research.md` section listing:
- All uses of `resource.TestCheckResourceAttr()`
- All uses of `resource.TestCheckResourceAttrSet()`
- Count by attribute type (string, int, bool, computed)
- Special cases (list attributes, nested paths)

**Expected Findings**:
- resource_cmnet_network_test.go: 18 legacy assertions (strings, ints, bools)
- resource_cmkube_cluster_test.go: 38 legacy assertions (mixed types, lists)
- data_source_cmpart_partitions_test.go: 10 legacy assertions (data source patterns)

#### Task 0.2: Modern Pattern Mapping

**Target**: Document transformation rules for each legacy pattern

**Method**: Extract patterns from CLAUDE.md Modern Testing Patterns section

**Deliverable**: `research.md` section with mapping table:

| Legacy Pattern | Modern Pattern | Type Matcher | Notes |
|----------------|----------------|--------------|-------|
| `TestCheckResourceAttr("resource", "name", "value")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("name"), knownvalue.StringExact("value"))` | StringExact | For string attributes |
| `TestCheckResourceAttr("resource", "mtu", "9000")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("mtu"), knownvalue.Int64Exact(9000))` | Int64Exact | For numeric attributes |
| `TestCheckResourceAttr("resource", "enabled", "true")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("enabled"), knownvalue.Bool(true))` | Bool | For boolean attributes |
| `TestCheckResourceAttrSet("resource", "uuid")` | `statecheck.ExpectKnownValue("resource", tfjsonpath.New("uuid"), knownvalue.NotNull())` | NotNull | For computed fields |

#### Task 0.3: Import Requirements Analysis

**Target**: Identify required import additions for each file

**Method**: Compare current imports vs modern pattern requirements

**Deliverable**: `research.md` section listing imports to add per file

**Expected Additions**:
```go
// Required for all files:
"github.com/hashicorp/terraform-plugin-testing/statecheck"
"github.com/hashicorp/terraform-plugin-testing/knownvalue"
"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
"github.com/hashicorp/terraform-plugin-testing/compare"  // For ID tracking

// Note: plancheck already imported in resource files
```

#### Task 0.4: Idempotency Pattern Analysis

**Target**: Identify where idempotency checks are missing

**Method**: Review test step sequences for Create/Update without follow-up idempotency steps

**Deliverable**: `research.md` section listing:
- Functions missing idempotency after Create
- Functions missing idempotency after Update
- Estimated new test steps to add (each idempotency check = 1 new step)

**Expected Findings**:
- resource_cmnet_network_test.go: TestAccCMNetNetwork_Complete, TestAccCMNetNetwork_Update need idempotency
- resource_cmkube_cluster_test.go: Already has comprehensive idempotency coverage
- data_source_cmpart_partitions_test.go: Data sources don't need idempotency (read-only)

#### Task 0.5: ID Tracking Pattern Analysis

**Target**: Identify resource tests needing ID consistency tracking

**Method**: Review resource tests for Create/Import/Update sequences

**Deliverable**: `research.md` section listing:
- Resource tests with Import steps (need ID tracking)
- Resource tests with Update steps (need ID tracking)
- Data source tests (don't need ID tracking - read-only)

**Expected Findings**:
- resource_cmnet_network_test.go: TestAccCMNetNetwork_Basic needs ID tracking (has Import)
- resource_cmkube_cluster_test.go: TestAccCMKubeClusterResource_Basic already has ID tracking
- data_source_cmpart_partitions_test.go: N/A (data sources)

### Research Outputs

**Artifact**: `specs/040-modernize-legacy-tests/research.md`

**Structure**:
1. Executive Summary
   - Total legacy assertions identified: 66
   - Transformation categories: 4 (string, int, bool, computed)
   - New test steps required: ~6 idempotency checks
   - Import additions: 3-4 per file

2. Legacy Pattern Inventory
   - File-by-file breakdown with line numbers
   - Pattern categorization

3. Modern Pattern Mappings
   - Comprehensive transformation table
   - Type matcher reference
   - Edge case handling

4. Import Requirements
   - Per-file import additions
   - Deduplication notes

5. Missing Check Identification
   - Idempotency gaps
   - ID tracking gaps
   - CheckDestroy verification

6. Edge Cases & Special Handling
   - List attribute assertions
   - Numeric type ambiguity (int vs float)
   - Environment portability concerns

---

## Phase 1: Design & Contracts

### Objectives

1. Generate data model of test structure (functions, steps, assertions)
2. Create before/after transformation contracts
3. Document migration sequence with dependencies
4. Update agent context for testing patterns

### Design Tasks

#### Task 1.1: Test Structure Data Model

**Target**: Create comprehensive map of test function structure

**Method**: Parse test files and extract test function metadata

**Deliverable**: `data-model.md` with schema:

```markdown
## Test File: resource_cmnet_network_test.go

### TestAccCMNetNetwork_Basic
- **Type**: Resource CRUD lifecycle
- **Current Steps**: 3 (Create, Idempotency, Import)
- **Legacy Assertions**: 4 (name, id, uuid, domain_name)
- **Modern Checks**: None
- **Missing**: None (already has idempotency)
- **Transformation**: Convert 4 legacy → 4 modern + add ID tracking

### TestAccCMNetNetwork_Complete
- **Type**: Resource full configuration
- **Current Steps**: 1 (Create only)
- **Legacy Assertions**: 10 (all attributes)
- **Modern Checks**: None
- **Missing**: Idempotency after Create
- **Transformation**: Convert 10 legacy → 10 modern + add idempotency step

[... similar for each function ...]
```

#### Task 1.2: Transformation Contracts

**Target**: Document exact before/after examples for each pattern type

**Method**: Extract representative examples from each file

**Deliverable**: `contracts/` directory with examples:

**File**: `contracts/01-string-attribute-transformation.md`
```markdown
# String Attribute Transformation

## Before (Legacy)
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "domain_name", "cluster.local"),
),
```

## After (Modern)
```go
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
},
```

## Key Changes
- Replace `Check:` with `ConfigStateChecks:`
- Use `tfjsonpath.New()` for path specification
- Use `knownvalue.StringExact()` for exact string matching
- Remove `resource.ComposeAggregateTestCheckFunc()` wrapper
```

**Additional Contract Files**:
- `contracts/02-numeric-attribute-transformation.md` (int64 handling)
- `contracts/03-boolean-attribute-transformation.md` (bool type safety)
- `contracts/04-computed-attribute-transformation.md` (NotNull pattern)
- `contracts/05-list-attribute-transformation.md` (ListSizeExact vs NotNull)
- `contracts/06-idempotency-addition.md` (new step pattern)
- `contracts/07-id-tracking-addition.md` (compareID pattern)

#### Task 1.3: CheckDestroy Verification

**Target**: Verify CheckDestroy functions match CLAUDE.md enhanced pattern

**Method**: Compare existing implementations against reference

**Deliverable**: `data-model.md` section documenting:
- Current CheckDestroy implementations
- Compliance with enhanced pattern
- Required modifications (if any)

**Expected Finding**:
- resource_cmnet_network_test.go: testAccCheckCMNetNetworkDestroy matches pattern ✅
- resource_cmkube_cluster_test.go: testAccCheckCMKubeClusterDestroy matches pattern ✅
- data_source_cmpart_partitions_test.go: No CheckDestroy (data source) N/A

#### Task 1.4: Migration Sequence Design

**Target**: Define order and dependencies for transformation

**Method**: Dependency analysis and risk assessment

**Deliverable**: `data-model.md` section with migration DAG:

```markdown
## Migration Sequence

### Phase 1 (P1): resource_cmnet_network_test.go
- **Rationale**: Smallest file, simplest patterns, minimal risk
- **Prerequisites**: None
- **Functions**: 3 total
  1. TestAccCMNetNetwork_Basic (has Import + idempotency) - Reference implementation
  2. TestAccCMNetNetwork_Complete (needs idempotency)
  3. TestAccCMNetNetwork_Update (needs idempotency)
- **Validation**: TF_ACC=1 go test -v -run TestAccCMNetNetwork
- **Estimated Time**: 30-45 minutes

### Phase 2 (P2): resource_cmkube_cluster_test.go
- **Rationale**: Most complex, already has some modern patterns, good reference
- **Prerequisites**: Phase 1 complete and validated
- **Functions**: 10 total (6 modern, 4 legacy)
  1. TestAccCMKubeClusterResource_Basic (modernize remaining legacy)
  2. TestAccCMKubeClusterResource_DriftDetection (modernize legacy)
  3. TestAccCMKubeClusterResource_WorkerNodes (modernize legacy)
  4. TestAccCMKubeClusterResource_CompleteConfiguration (modernize legacy)
  5. TestAccCMKubeClusterResource_P3AdvancedNetworking (already modern - no changes)
  6. TestAccCMKubeClusterResource_P3StorageAndLoadBalancer (already modern - no changes)
  7. TestAccCMKubeClusterResource_P3Addons (already modern - no changes)
  8. TestAccCMKubeClusterResource_P3FullStack (already modern - no changes)
  9. TestAccCMKubeClusterResource_ValidationInvalidName (no assertions - no changes)
  10. TestAccCMKubeClusterResource_ValidationInvalidVersion (no assertions - no changes)
- **Validation**: TF_ACC=1 go test -v -run TestAccCMKubeCluster
- **Estimated Time**: 60-90 minutes

### Phase 3 (P3): data_source_cmpart_partitions_test.go
- **Rationale**: Data source patterns, some functions already modern
- **Prerequisites**: Phases 1 & 2 complete and validated
- **Functions**: 8 total (mixed modern/legacy)
  1. TestAccCMPartPartitionsDataSource_Basic (already modern - no changes)
  2. TestAccCMPartPartitionsDataSource_FilterByNamePattern (already modern - no changes)
  3. TestAccCMPartPartitionsDataSource_NoMatches (already modern - no changes)
  4. TestAccCMPartPartitionsDataSource_ComputedFields (modernize legacy)
  5. TestAccCMPartPartitionsDataSource_AttributeTypes (modernize legacy)
  6. TestAccCMPartPartitionsDataSource_ListAttributes (modernize legacy)
  7. TestAccCMPartPartitionsDataSource_FilterCaseInsensitive (already modern - no changes)
  8. TestAccCMPartPartitionsDataSource_FilterEmptyString (modernize legacy)
  9. TestAccCMPartPartitionsDataSource_FilterSubsetProperty (modernize legacy)
- **Validation**: TF_ACC=1 go test -v -run TestAccCMPartPartitions
- **Estimated Time**: 45-60 minutes

### Phase 4 (Validation): Full Test Suite
- **Rationale**: Verify no regressions across entire test suite
- **Prerequisites**: All phases complete
- **Command**: TF_ACC=1 go test -v -timeout 120m ./internal/provider/
- **Success Criteria**: 100% pass rate, zero regressions
- **Estimated Time**: 120 minutes (full suite runtime)
```

#### Task 1.5: Quick Start Guide

**Target**: Create developer quick reference for applying patterns

**Deliverable**: `quickstart.md` with:

**Structure**:
1. Prerequisites (environment setup, BCM access)
2. Transformation Checklist
   - [ ] Add required imports
   - [ ] Transform string attributes
   - [ ] Transform numeric attributes
   - [ ] Transform boolean attributes
   - [ ] Transform computed attributes
   - [ ] Add idempotency checks
   - [ ] Add ID tracking (resources only)
   - [ ] Run acceptance tests
   - [ ] Verify zero regressions
3. Common Patterns Quick Reference
   - String: `knownvalue.StringExact("value")`
   - Integer: `knownvalue.Int64Exact(9000)`
   - Boolean: `knownvalue.Bool(true)`
   - Computed: `knownvalue.NotNull()`
   - List (known): `knownvalue.ListSizeExact(n)`
   - List (unknown): `knownvalue.NotNull()`
4. Testing Commands
   - Run specific test: `TF_ACC=1 go test -v -run TestName`
   - Run file tests: `TF_ACC=1 go test -v -run TestAccCMNetNetwork`
   - Run all tests: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/`
5. Troubleshooting
   - Type mismatch errors → check Int64Exact vs StringExact
   - Unknown value errors → ensure NotNull for computed fields
   - Import failures → verify compareID tracking added

#### Task 1.6: Agent Context Update

**Target**: Update agent-specific context files with testing modernization patterns

**Method**: Run `.specify/scripts/bash/update-agent-context.sh copilot`

**Deliverable**: Updated context files with:
- Modern testing pattern references
- File-specific transformation examples
- Common pitfall warnings

**Expected Updates**:
- Add "Testing Modernization" section to agent context
- Include pattern transformation quick reference
- Link to contracts/ examples

### Design Outputs

**Artifacts**:
1. `specs/040-modernize-legacy-tests/data-model.md` - Test structure mapping
2. `specs/040-modernize-legacy-tests/contracts/` - Before/after examples (7 files)
3. `specs/040-modernize-legacy-tests/quickstart.md` - Developer guide
4. Agent context updates - Testing pattern references

---

## Phase 2: Implementation Planning (Tasks)

**Note**: This phase is handled by the `/speckit.tasks` command, which generates `tasks.md` based on the plan and design artifacts from Phases 0-1.

**Expected Task Structure**:
1. **Setup Tasks**
   - Add required imports to P1 file
   - Add required imports to P2 file
   - Add required imports to P3 file

2. **P1 Implementation Tasks** (resource_cmnet_network_test.go)
   - Modernize TestAccCMNetNetwork_Basic
   - Modernize TestAccCMNetNetwork_Complete + add idempotency
   - Modernize TestAccCMNetNetwork_Update + add idempotency
   - Validate P1 with acceptance tests

3. **P2 Implementation Tasks** (resource_cmkube_cluster_test.go)
   - Modernize TestAccCMKubeClusterResource_Basic
   - Modernize TestAccCMKubeClusterResource_DriftDetection
   - Modernize TestAccCMKubeClusterResource_WorkerNodes
   - Modernize TestAccCMKubeClusterResource_CompleteConfiguration
   - Validate P2 with acceptance tests

4. **P3 Implementation Tasks** (data_source_cmpart_partitions_test.go)
   - Modernize TestAccCMPartPartitionsDataSource_ComputedFields
   - Modernize TestAccCMPartPartitionsDataSource_AttributeTypes
   - Modernize TestAccCMPartPartitionsDataSource_ListAttributes
   - Modernize remaining legacy functions
   - Validate P3 with acceptance tests

5. **Final Validation**
   - Run full test suite
   - Verify zero legacy patterns remain
   - Verify 100% pass rate
   - Update documentation

---

## Testing Strategy

### Unit Testing (N/A)

This feature does not involve unit testing - it's a test modernization effort.

### Integration Testing (Acceptance Tests)

**Strategy**: Validate each priority phase independently before proceeding

#### P1 Validation (resource_cmnet_network_test.go)

```bash
# After P1 transformations complete
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 30m ./internal/provider/ -run TestAccCMNetNetwork

# Expected: PASS for all 3 tests
# Expected: Zero regressions from baseline
```

**Success Criteria**:
- All 3 test functions pass
- No increase in test execution time (>10% threshold)
- Zero legacy `TestCheckResourceAttr` calls remain in file
- All modern patterns correctly implemented

#### P2 Validation (resource_cmkube_cluster_test.go)

```bash
# After P2 transformations complete
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 60m ./internal/provider/ -run TestAccCMKubeCluster

# Expected: PASS for all 10 tests
# Expected: Zero regressions from baseline
```

**Success Criteria**:
- All 10 test functions pass
- No increase in test execution time (>10% threshold)
- Zero legacy assertion calls remain in modernized functions
- ID tracking correctly implemented
- Idempotency checks pass

#### P3 Validation (data_source_cmpart_partitions_test.go)

```bash
# After P3 transformations complete
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 30m ./internal/provider/ -run TestAccCMPartPartitions

# Expected: PASS for all 8 tests
# Expected: Zero regressions from baseline
```

**Success Criteria**:
- All 8 test functions pass
- No increase in test execution time (>10% threshold)
- Zero legacy assertion calls remain in modernized functions
- Type-safe matchers correctly applied

#### Full Suite Validation

```bash
# After all phases complete
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/

# Expected: PASS for entire test suite
# Expected: Zero regressions across all tests
```

**Success Criteria** (SC-001 through SC-009):
- SC-001: Zero `resource.TestCheckResourceAttr()` calls in 3 target files ✅
- SC-002: Zero `resource.TestCheckResourceAttrSet()` calls in 3 target files ✅
- SC-003: All numeric attributes use `Int64Exact()` not string "9000" ✅
- SC-004: All boolean attributes use `Bool(true/false)` not string "true" ✅
- SC-005: All resource tests include idempotency verification ✅
- SC-006: All resource tests include ID consistency tracking ✅
- SC-007: 100% modern pattern adoption in all 3 files ✅
- SC-008: All acceptance tests pass without regressions ✅
- SC-009: Test execution time within 10% of baseline ✅

### Test Environment Requirements

**BCM Cluster**:
- Endpoint: `https://172.21.15.254:8081`
- Credentials: root / Hashicorp123!
- Available resources: networks, nodes, partitions
- Minimum viable configuration for all test scenarios

**Environment Variables**:
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
export TF_LOG=DEBUG  # Optional for debugging
```

---

## Architecture

### Test Modernization Architecture

```
┌─────────────────────────────────────────────────────────┐
│ Legacy Test Pattern (terraform-plugin-sdk style)        │
├─────────────────────────────────────────────────────────┤
│ Check: resource.ComposeAggregateTestCheckFunc(         │
│   resource.TestCheckResourceAttr("res", "attr", "val") │
│ )                                                       │
└─────────────────────────────────────────────────────────┘
                         │
                         │ Transformation
                         ▼
┌─────────────────────────────────────────────────────────┐
│ Modern Test Pattern (terraform-plugin-testing v1.13.3+) │
├─────────────────────────────────────────────────────────┤
│ ConfigStateChecks: []statecheck.StateCheck{            │
│   statecheck.ExpectKnownValue(                         │
│     "res",                                             │
│     tfjsonpath.New("attr"),                            │
│     knownvalue.StringExact("val"),                     │
│   ),                                                   │
│ }                                                      │
└─────────────────────────────────────────────────────────┘
```

### Type Matcher Decision Tree

```
Attribute Type?
├── String (name, path, notes)
│   └── knownvalue.StringExact("value")
│
├── Numeric (mtu, port, count)
│   ├── Try: knownvalue.Int64Exact(9000)
│   └── If fails: knownvalue.Float64Exact(9000.0)
│
├── Boolean (enabled, dhcp)
│   └── knownvalue.Bool(true/false)
│
├── Computed (uuid, id, timestamp)
│   └── knownvalue.NotNull()
│
└── List/Set
    ├── Known count (test-created resources)
    │   └── knownvalue.ListSizeExact(n)
    └── Unknown count (environment-dependent)
        └── knownvalue.NotNull()
```

### Idempotency Check Pattern

```
Test Step Sequence:

1. Create/Update Step
   ├── Config: testConfig(params)
   ├── Check: legacy assertions (being removed)
   └── ConfigStateChecks: modern assertions (being added)

2. Idempotency Step (NEW)
   ├── Config: same as step 1
   └── ConfigPlanChecks:
       └── PreApply: []plancheck.PlanCheck{
           plancheck.ExpectEmptyPlan(),
       }

Rationale: Verifies no spurious diffs after Create/Update
```

### ID Consistency Tracking Pattern

```
Test Function Structure:

func TestAccResource(t *testing.T) {
    // Initialize ID tracker ONCE at function start
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Create Step
            {
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("res", tfjsonpath.New("id")),
                    // First call captures baseline ID
                },
            },
            // Import Step
            {
                ImportState: true,
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("res", tfjsonpath.New("id")),
                    // Verifies imported ID matches created ID
                },
            },
            // Update Step
            {
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("res", tfjsonpath.New("id")),
                    // Verifies ID unchanged after update
                },
            },
        },
    })
}

Rationale: Ensures resource ID stability across CRUD operations
```

---

## Implementation Approach

### File-by-File Transformation Strategy

#### Priority 1: resource_cmnet_network_test.go

**Complexity**: Low
**Rationale**: Smallest file, straightforward patterns, good starter for establishing workflow

**Functions to Modernize**:

1. **TestAccCMNetNetwork_Basic** (lines 18-54)
   - **Current State**: Has idempotency check (line 37-44), has Import (line 46-51)
   - **Legacy Assertions**: 4 (lines 30-33)
     - `TestCheckResourceAttr("name", networkName)` → `StringExact(networkName)`
     - `TestCheckResourceAttrSet("id")` → `NotNull()`
     - `TestCheckResourceAttrSet("uuid")` → `NotNull()`
     - `TestCheckResourceAttr("domain_name", "cluster.local")` → `StringExact("cluster.local")`
   - **Missing**: ID tracking (compareID pattern)
   - **Transformation Steps**:
     1. Add imports (statecheck, knownvalue, tfjsonpath, compare)
     2. Replace Check block with ConfigStateChecks
     3. Add `compareID := statecheck.CompareValue(compare.ValuesSame())` at function start
     4. Add ID tracking to Create step (line 27-35)
     5. Add ID tracking to Import step (line 46-51)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMNetNetwork_Basic`

2. **TestAccCMNetNetwork_Complete** (lines 56-82)
   - **Current State**: Single Create step, no idempotency, no Import
   - **Legacy Assertions**: 10 (lines 68-78)
     - Strings: name, subnet, gateway, domain_name, dhcp_range_start, dhcp_range_end, notes
     - Numeric: mtu ("9000" → Int64Exact(9000))
     - Boolean: dhcp_enabled ("true" → Bool(true))
     - Computed: uuid
   - **Missing**: Idempotency check after Create
   - **Transformation Steps**:
     1. Replace Check block with ConfigStateChecks (10 assertions)
     2. Add idempotency step after Create:
        ```go
        {
            Config: testAccCMNetNetworkConfigComplete(networkName),
            ConfigPlanChecks: resource.ConfigPlanChecks{
                PreApply: []plancheck.PlanCheck{
                    plancheck.ExpectEmptyPlan(),
                },
            },
        },
        ```
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMNetNetwork_Complete`

3. **TestAccCMNetNetwork_Update** (lines 84-110)
   - **Current State**: Create → Update sequence, no idempotency checks
   - **Legacy Assertions**: 4 (lines 95-96 Create, 103-106 Update)
   - **Missing**: Idempotency after Create, idempotency after Update
   - **Transformation Steps**:
     1. Replace both Check blocks with ConfigStateChecks
     2. Add idempotency step after Create (line 98 insertion point)
     3. Add idempotency step after Update (line 107 insertion point)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMNetNetwork_Update`

**P1 Validation**: `TF_ACC=1 go test -v -run TestAccCMNetNetwork`

---

#### Priority 2: resource_cmkube_cluster_test.go

**Complexity**: High
**Rationale**: Largest file, mixed modern/legacy, comprehensive patterns already demonstrated

**Current State Analysis**:
- 6 functions already fully modern (P3 tests: lines 674-940)
- 2 validation tests with no assertions (lines 425-455)
- 4 functions needing modernization

**Functions to Modernize**:

1. **TestAccCMKubeClusterResource_Basic** (lines 160-254)
   - **Current State**: Already has modern ConfigStateChecks (lines 183-203), idempotency (205-213), ID tracking (199-202, 237-240)
   - **Legacy Assertions**: 3 (lines 179-182)
     - `TestCheckResourceAttr("name", clusterName)` → Already modern in ConfigStateChecks
     - `TestCheckResourceAttrSet("uuid")` → Already modern (line 189-193)
     - `TestCheckResourceAttrSet("id")` → Already modern (line 194-198)
   - **Issue**: Has BOTH Check block (legacy) AND ConfigStateChecks (modern) - redundant
   - **Transformation Steps**:
     1. Remove Check block entirely (lines 178-183)
     2. Keep ConfigStateChecks (already correct)
     3. Verify Update step Check block (line 228) - convert to ConfigStateChecks
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_Basic`

2. **TestAccCMKubeClusterResource_DriftDetection** (lines 257-352)
   - **Current State**: Has ConfigStateChecks (modern), but also Check blocks (legacy)
   - **Legacy Assertions**: 2 (line 270, line 340)
   - **Transformation Steps**:
     1. Remove Check block at line 270
     2. Remove Check block at line 340
     3. Keep ConfigStateChecks (already correct at lines 273-278, 343-348)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_DriftDetection`

3. **TestAccCMKubeClusterResource_WorkerNodes** (lines 355-422)
   - **Current State**: Has ConfigStateChecks, but also legacy Check blocks
   - **Legacy Assertions**: 3 (lines 374, 392, 410)
   - **Transformation Steps**:
     1. Remove Check block at line 374 (keep ConfigStateChecks 376-382)
     2. Remove Check block at line 392 (keep ConfigStateChecks 394-400)
     3. Remove Check block at line 410 (keep ConfigStateChecks 412-418)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_WorkerNodes`

4. **TestAccCMKubeClusterResource_CompleteConfiguration** (lines 462-564)
   - **Current State**: Has ConfigStateChecks, but also legacy Check blocks
   - **Legacy Assertions**: 5 (lines 486-492, 545-548)
   - **Transformation Steps**:
     1. Remove Check block at lines 486-492 (keep ConfigStateChecks 493-519)
     2. Remove Check block at lines 545-548 (keep ConfigStateChecks 549-561)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMKubeClusterResource_CompleteConfiguration`

**Note**: Functions already fully modern (NO CHANGES NEEDED):
- TestAccCMKubeClusterResource_P3AdvancedNetworking (lines 675-741)
- TestAccCMKubeClusterResource_P3StorageAndLoadBalancer (lines 744-799)
- TestAccCMKubeClusterResource_P3Addons (lines 802-864)
- TestAccCMKubeClusterResource_P3FullStack (lines 867-939)
- TestAccCMKubeClusterResource_ValidationInvalidName (lines 425-438) - no assertions
- TestAccCMKubeClusterResource_ValidationInvalidVersion (lines 441-455) - no assertions

**P2 Validation**: `TF_ACC=1 go test -v -run TestAccCMKubeCluster`

---

#### Priority 3: data_source_cmpart_partitions_test.go

**Complexity**: Medium
**Rationale**: Data source patterns, already has modern patterns in some functions

**Current State Analysis**:
- 4 functions already fully modern (Basic, FilterByNamePattern, NoMatches, FilterCaseInsensitive)
- 4 functions needing modernization (mixed legacy Check blocks)

**Functions to Modernize**:

1. **TestAccCMPartPartitionsDataSource_ComputedFields** (lines 85-114)
   - **Current State**: Has ConfigStateChecks (modern), but also legacy Check block
   - **Legacy Assertions**: 1 (line 95 - `TestCheckResourceAttrSet("partitions.#")`)
   - **Transformation Steps**:
     1. Remove Check block entirely (lines 93-96)
     2. Keep ConfigStateChecks (already correct at lines 98-109)
     3. Optionally add: `statecheck.ExpectKnownValue("partitions", knownvalue.NotNull())`
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_ComputedFields`

2. **TestAccCMPartPartitionsDataSource_AttributeTypes** (lines 158-215)
   - **Current State**: Has ConfigStateChecks (modern), but also legacy Check blocks
   - **Legacy Assertions**: 5 (lines 167-176)
   - **Transformation Steps**:
     1. Remove Check block entirely (lines 166-177)
     2. Keep ConfigStateChecks (already correct at lines 178-211)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_AttributeTypes`

3. **TestAccCMPartPartitionsDataSource_ListAttributes** (lines 218-257)
   - **Current State**: Has ConfigStateChecks (modern), but also legacy Check block
   - **Legacy Assertions**: 1 (line 228 - `TestCheckResourceAttrSet("partitions.#")`)
   - **Transformation Steps**:
     1. Remove Check block entirely (lines 226-229)
     2. Keep ConfigStateChecks (already correct at lines 230-253)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_ListAttributes`

4. **TestAccCMPartPartitionsDataSource_FilterEmptyString** (lines 326-354)
   - **Current State**: Has ConfigStateChecks (modern), but also legacy Check block
   - **Legacy Assertions**: 1 (line 335 - `TestCheckResourceAttrSet("partitions.#")`)
   - **Transformation Steps**:
     1. Remove Check block entirely (lines 333-336)
     2. Keep ConfigStateChecks (already correct at lines 337-350)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_FilterEmptyString`

5. **TestAccCMPartPartitionsDataSource_FilterSubsetProperty** (lines 357-406)
   - **Current State**: Has ConfigStateChecks (modern), but also legacy Check blocks
   - **Legacy Assertions**: 2 (line 368, line 388)
   - **Transformation Steps**:
     1. Remove Check block at lines 366-370
     2. Remove Check block at lines 386-390
     3. Keep ConfigStateChecks (already correct at lines 371-382, 391-403)
   - **Test**: `TF_ACC=1 go test -v -run TestAccCMPartPartitionsDataSource_FilterSubsetProperty`

**Note**: Functions already fully modern (NO CHANGES NEEDED):
- TestAccCMPartPartitionsDataSource_Basic (lines 18-38)
- TestAccCMPartPartitionsDataSource_FilterByNamePattern (lines 41-61)
- TestAccCMPartPartitionsDataSource_NoMatches (lines 64-82)
- TestAccCMPartPartitionsDataSource_FilterCaseInsensitive (lines 260-323)

**P3 Validation**: `TF_ACC=1 go test -v -run TestAccCMPartPartitions`

---

## Risk Mitigation

### Risk 1: Type Matcher Mismatches

**Risk**: Using wrong type matcher (e.g., StringExact for numeric) causes test failures

**Mitigation**:
- Start with empirical approach: try Int64Exact first for numerics
- If test fails with type error, switch to Float64Exact
- Document all numeric type decisions in research.md
- Cross-reference with schema definitions in resource implementation

**Fallback**: If type ambiguity persists, use NotNull() for computed fields

### Risk 2: Test Execution Time Regression

**Risk**: Modern patterns add overhead, increasing test execution time beyond 10% threshold

**Mitigation**:
- Baseline current test execution time before any changes
- Measure after each priority phase
- If >10% increase detected:
  - Review for inefficient assertions (nested loops in state checks)
  - Consider parallel test execution tuning
  - Document acceptable increase with justification

**Baseline Command**:
```bash
# Before any changes
TF_ACC=1 time go test -v -timeout 120m ./internal/provider/ -run "TestAccCMNetNetwork|TestAccCMKubeCluster|TestAccCMPartPartitions"
```

### Risk 3: BCM Cluster Unavailability

**Risk**: BCM cluster offline during transformation prevents validation

**Mitigation**:
- Verify cluster reachability before starting each phase
- Keep transformation work separate from validation
- Can complete transformation offline, defer validation until cluster available
- Document cluster requirements in quickstart.md

**Precheck Script**:
```bash
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"Hashicorp123!"}' \
  && echo "✅ BCM cluster reachable" || echo "❌ BCM cluster offline"
```

### Risk 4: Import Conflicts

**Risk**: Adding new imports conflicts with existing imports or causes Go module issues

**Mitigation**:
- Run `go mod tidy` after adding imports to each file
- Use Go's automatic import management: `goimports -w <file>`
- Verify no duplicate imports
- Test compilation before running acceptance tests

**Validation Command**:
```bash
go build ./internal/provider/ && echo "✅ Imports valid" || echo "❌ Import errors"
```

### Risk 5: ID Tracking State Pollution

**Risk**: compareID tracking across multiple test steps causes false positives if state polluted

**Mitigation**:
- Initialize compareID once per test function (not per step)
- Use unique test resource names (already done via generateUniqueTestName)
- Verify CheckDestroy properly cleans up resources
- Run tests in isolation first, then in full suite

**Validation**: Compare test results when run individually vs in full suite

### Risk 6: Incomplete Legacy Pattern Removal

**Risk**: Miss some legacy assertions during transformation, failing SC-001/SC-002 success criteria

**Mitigation**:
- After each file transformation, run grep verification:
  ```bash
  # Must return zero matches
  grep -n "TestCheckResourceAttr\|TestCheckResourceAttrSet" internal/provider/resource_cmnet_network_test.go
  ```
- Use IDE search to find all occurrences before declaring file complete
- Peer review checklist includes pattern verification

**Automated Check**:
```bash
# Add to validation script
LEGACY_COUNT=$(grep -c "TestCheckResourceAttr" internal/provider/{resource_cmnet_network_test.go,resource_cmkube_cluster_test.go,data_source_cmpart_partitions_test.go})
if [ "$LEGACY_COUNT" -eq 0 ]; then
  echo "✅ Zero legacy patterns remaining"
else
  echo "❌ Found $LEGACY_COUNT legacy patterns"
  exit 1
fi
```

---

## Deliverables

### Phase 0 Deliverables

1. ✅ `specs/040-modernize-legacy-tests/research.md`
   - Legacy pattern inventory (66 assertions cataloged)
   - Modern pattern mappings (transformation rules)
   - Import requirements (per-file additions)
   - Missing check identification (idempotency/ID tracking gaps)
   - Edge case documentation

### Phase 1 Deliverables

1. ✅ `specs/040-modernize-legacy-tests/data-model.md`
   - Test function structure mapping
   - CheckDestroy verification
   - Migration sequence design

2. ✅ `specs/040-modernize-legacy-tests/contracts/`
   - 01-string-attribute-transformation.md
   - 02-numeric-attribute-transformation.md
   - 03-boolean-attribute-transformation.md
   - 04-computed-attribute-transformation.md
   - 05-list-attribute-transformation.md
   - 06-idempotency-addition.md
   - 07-id-tracking-addition.md

3. ✅ `specs/040-modernize-legacy-tests/quickstart.md`
   - Prerequisites checklist
   - Transformation checklist
   - Pattern quick reference
   - Testing commands
   - Troubleshooting guide

4. ✅ Updated agent context files
   - Testing modernization patterns
   - Transformation examples
   - Common pitfalls

### Phase 2 Deliverables (via /speckit.tasks)

1. ✅ `specs/040-modernize-legacy-tests/tasks.md`
   - Dependency-ordered task list
   - Estimated time per task
   - Validation checkpoints

### Implementation Deliverables

1. ✅ Modernized test files (3 files)
   - `internal/provider/resource_cmnet_network_test.go`
   - `internal/provider/resource_cmkube_cluster_test.go`
   - `internal/provider/data_source_cmpart_partitions_test.go`

2. ✅ Validation reports
   - P1 validation results
   - P2 validation results
   - P3 validation results
   - Full suite validation results

3. ✅ Success criteria verification
   - SC-001 through SC-009 checklist completion
   - Zero legacy patterns remaining (grep verification)
   - 100% pass rate confirmation
   - Performance baseline comparison

---

## Timeline Estimate

| Phase | Duration | Dependencies | Deliverables |
|-------|----------|--------------|--------------|
| **Phase 0: Research** | 2-3 hours | None | research.md |
| **Phase 1: Design** | 3-4 hours | Phase 0 complete | data-model.md, contracts/, quickstart.md |
| **Phase 2: Task Planning** | 1 hour | Phase 1 complete | tasks.md (via /speckit.tasks) |
| **P1 Implementation** | 1-2 hours | Phase 2 complete | resource_cmnet_network_test.go modernized |
| **P1 Validation** | 30 min | P1 impl complete | P1 test results |
| **P2 Implementation** | 2-3 hours | P1 validated | resource_cmkube_cluster_test.go modernized |
| **P2 Validation** | 1 hour | P2 impl complete | P2 test results |
| **P3 Implementation** | 1-2 hours | P2 validated | data_source_cmpart_partitions_test.go modernized |
| **P3 Validation** | 30 min | P3 impl complete | P3 test results |
| **Full Suite Validation** | 2 hours | All phases complete | Final validation report |
| **Documentation** | 30 min | Validation complete | Updated CLAUDE.md if needed |
| **TOTAL** | **14-19 hours** | Sequential execution | All deliverables |

**Critical Path**: Phase 0 → Phase 1 → Phase 2 → P1 (impl+val) → P2 (impl+val) → P3 (impl+val) → Full validation

**Parallelization Opportunities**: None - sequential transformation ensures validation at each step before proceeding

---

## Success Metrics

### Quantitative Metrics

1. **Legacy Pattern Elimination**: 0 legacy assertions remaining (from baseline 66)
2. **Modern Pattern Adoption**: 100% coverage in 3 target files
3. **Test Pass Rate**: 100% (21 test functions passing)
4. **Performance**: Test execution time within 10% of baseline
5. **Coverage**: 21 test functions modernized (100% of target)

### Qualitative Metrics

1. **Type Safety**: All attributes use type-appropriate matchers
2. **Idempotency**: All resource tests verify no spurious diffs
3. **ID Consistency**: All resource tests track ID across CRUD operations
4. **Code Quality**: Tests follow HashiCorp best practices per CLAUDE.md
5. **Documentation**: Comprehensive guide for future test modernization

### Validation Checkpoints

- [ ] Phase 0: research.md complete with 66 assertions cataloged
- [ ] Phase 1: All design artifacts generated (data-model.md, contracts/, quickstart.md)
- [ ] P1 Complete: resource_cmnet_network_test.go 100% modern, all tests pass
- [ ] P2 Complete: resource_cmkube_cluster_test.go 100% modern, all tests pass
- [ ] P3 Complete: data_source_cmpart_partitions_test.go 100% modern, all tests pass
- [ ] Full Suite: All 21 functions pass, zero legacy patterns, <10% execution time increase
- [ ] SC-001 through SC-009: All success criteria met

---

## Notes

### Implementation Notes

1. **Import Organization**: Add new imports in alphabetical order per Go conventions
2. **Line Wrapping**: Keep ConfigStateChecks readable with one statecheck per logical group
3. **Comments**: Preserve existing test comments explaining BCM API limitations
4. **Test Helpers**: No changes needed to test_helpers.go - reuse existing functions
5. **CheckDestroy**: Already follows enhanced pattern - verify but don't modify

### Edge Cases

1. **Numeric Type Ambiguity**: Start with Int64Exact, fall back to Float64Exact if type errors
2. **Empty Lists**: Use NotNull() for environment-dependent counts, ListSizeExact(0) for known empty
3. **Computed Fields**: Always use NotNull(), never StringExact with hardcoded values
4. **Data Source ID**: data_source_cmpart_partitions_test.go uses "placeholder" ID - preserve this

### Dependencies on External Systems

1. **BCM Cluster**: Required for all acceptance test validation
2. **Go Toolchain**: Go 1.24.0+ for terraform-plugin-testing v1.13.3 compatibility
3. **Terraform**: terraform-plugin-framework v1.16.1 (already installed)
4. **Git**: For branch management and change tracking

### Future Considerations

1. **Template for Future Tests**: quickstart.md serves as template for new test development
2. **CI/CD Integration**: Consider adding legacy pattern detection to pre-commit hooks
3. **Documentation Updates**: Update CLAUDE.md if new patterns emerge during implementation
4. **Performance Baseline**: Establish benchmark for future test additions to prevent regression
