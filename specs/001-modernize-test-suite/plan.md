# Implementation Plan: Modernize Terraform BCM Provider Test Suite

**Branch**: `001-modernize-test-suite` | **Date**: 2025-11-21 | **Spec**: [spec.md](./spec.md)

## Summary

Modernize the Terraform BCM provider test suite to achieve 90%+ quality score by implementing HashiCorp's modern testing patterns from terraform-plugin-testing v1.13.3. This includes replacing legacy string-based assertions with type-safe state checks, adding idempotency verification with plan checks, enhancing data source filter verification, removing environment-specific assumptions, and improving CheckDestroy error reporting. The implementation follows TDD RED-GREEN-REFACTOR cycles with parallel execution where possible.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**:
- terraform-plugin-framework v1.16.1
- terraform-plugin-testing v1.13.3 (modern patterns from v1.5.0+)
- terraform-plugin-go (tfprotov6)

**Storage**: BCM JSON-RPC API (cookie-based authentication)
**Testing**:
- Acceptance tests (TF_ACC=1, 120m timeout)
- Test helper functions (createTestBCMClient, verifyResourceDeleted, generateUniqueTestName)
- BCM cluster at BCM_ENDPOINT for live testing

**Target Platform**: Linux server (BCM cluster integration)
**Project Type**: Terraform provider (single project, internal/provider structure)
**Performance Goals**:
- Test execution time remains within 10% of current baseline
- 100% acceptance test pass rate maintained
- All tests pass on different BCM cluster configurations

**Constraints**:
- Tests must work on any BCM cluster (no hardcoded environment assumptions)
- Backward compatibility required (all existing test scenarios must pass)
- Exponential backoff for eventual consistency (within 30s for CheckDestroy)

**Scale/Scope**:
- 35 total tests across 6 test files
- 2 resources (software_image, category) with 18 tests
- 4 data sources (nodes, networks, categories, software_images) with 17 tests

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Test-First Development (TDD)

✅ **PASS** - This is test modernization, not new feature development. Existing tests already cover resource/data source behavior. We are enhancing test quality using modern patterns while maintaining 100% coverage.

### Simplicity and Incrementality

✅ **PASS** - Modernization follows incremental approach:
1. HIGH PRIORITY: Resource state checks + idempotency (core reliability)
2. HIGH PRIORITY: Data source filter verification (correctness)
3. MEDIUM PRIORITY: Validation tests + CheckDestroy improvements
4. LOW PRIORITY: Documentation and cleanup

### Parallel Execution Pattern

✅ **PASS** - TDD cycles will execute in parallel batches:
- Phase 1: Modernize multiple resource tests concurrently
- Phase 2: Enhance multiple data source tests concurrently
- Within each file: Add state checks + plan checks + filter verification in single RED-GREEN-REFACTOR cycle

### Constitution-Specific Gates

**No Additional Violations** - Test modernization follows existing architecture, no new patterns or abstractions introduced.

## Project Structure

### Documentation (this feature)

```text
specs/001-modernize-test-suite/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (modern testing patterns)
├── data-model.md        # Phase 1 output (test entities and patterns)
├── quickstart.md        # Phase 1 output (developer quick start)
├── contracts/           # Phase 1 output (test structure examples)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/provider/
├── test_helpers.go                         # Existing test utilities (NO CHANGES)
├── provider_test.go                        # Test setup (NO CHANGES)
│
├── resource_cmpart_softwareimage_test.go   # [MODIFY] Phase 1: 12 tests
│   # Current: Legacy TestCheckResourceAttr, no idempotency checks
│   # Target: Modern state checks, plan checks, ID tracking
│
├── resource_cmdevice_category_test.go      # [MODIFY] Phase 1: 6 tests
│   # Current: Legacy assertions, no idempotency checks
│   # Target: Modern state checks, plan checks, ID tracking
│
├── data_source_cmdevice_nodes_test.go      # [MODIFY] Phase 2: 4 tests
│   # Current: Basic existence checks, no filter verification
│   # Target: State checks + filter verification for node_type/hostname
│
├── data_source_cmnet_networks_test.go      # [MODIFY] Phase 2: 4 tests
│   # Current: Hardcoded counts ("networks.# = 3"), specific names
│   # Target: Dynamic assertions, filter verification for name/dhcp
│
├── data_source_cmpart_softwareimages_test.go # [MODIFY] Phase 2: 5 tests
│   # Current: Minimal verification, duplicate scenarios
│   # Target: State checks + filter verification
│
└── data_source_cmdevice_categories_test.go # [MODIFY] Phase 3: 4 tests
    # Current: Basic checks
    # Target: Enhanced state checks
```

**Structure Decision**: Standard Terraform provider structure with tests colocated in `internal/provider/`. No structural changes needed. Test helper functions already exist and work correctly. Focus is on enhancing test quality using modern terraform-plugin-testing patterns.

## Complexity Tracking

> **Not applicable** - No constitution violations to justify. This is test modernization following established TDD patterns.

---

# Phase 0: Research & Pattern Discovery

**Goal**: Research modern testing patterns in terraform-plugin-testing v1.13.3 and understand current test quality baseline.

**Prerequisites**: Access to HashiCorp documentation, existing test files, terraform-provider-design skill

**Output**: `research.md` with decisions on:
1. Modern state check patterns (ExpectKnownValue, CompareValue)
2. Plan check patterns (ExpectEmptyPlan for idempotency)
3. Known value matchers for BCM provider types (StringExact, Bool, Int64, ListExact)
4. Filter verification patterns (attribute path navigation with tfjsonpath)
5. Environment portability patterns (dynamic assertions vs hardcoded counts)
6. CheckDestroy enhancement patterns (detailed error messages)

## HashiCorp Testing Patterns Reference

**Source**: https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns

### Four Core Testing Pattern Categories

HashiCorp identifies four primary testing patterns for Terraform provider development:

1. **Built-in Patterns** - Framework-provided behaviors like automatic plan/apply/refresh/destroy cycles
2. **Basic Attribute Verification** - Testing resource creation and attribute persistence with correct values
3. **Configuration Updates** - Validating resource modifications across multiple test steps (superset of basic tests)
4. **Import Mode Testing** - Ensuring import operations produce equivalent state to creation

### Basic Attribute Verification Pattern

A foundational test should establish:
- Terraform can plan and apply a common resource configuration without error
- Expected attributes are saved to state with correct values
- Remote API/service values match state values
- Subsequent plans produce no differences (idempotency)

**Key Principle**: "Use a combination of the built-in `statecheck` implementations to verify attributes are saved to the state file correctly."

### Update Test Pattern

Update tests extend basic tests by adding additional `TestStep` instances with modified configurations. HashiCorp notes: "It's common for resources to just have the above update test, as it is a superset of the basic test."

This pattern demonstrates:
- Configuration changes are properly applied
- State reflects updated values
- No unintended differences remain after apply

### Import Mode Testing Pattern

Import testing uses convention-over-configuration:
- `ImportState: true` activates import mode
- Framework reuses prior configuration and state artifacts
- Generated import blocks combine with base configuration
- **Golden File Pattern**: "The statefile acts as a 'golden file' reference. In import mode, the planned values are expected to match the state values."

### Error and Plan Expectation Patterns

**ExpectNonEmptyPlan**: Allows tests to continue when configuration results in persistent plan differences (useful for demonstrating misconfiguration correction workflows).

**ExpectError**: Validates invalid configurations fail appropriately. "ExpectError expects a valid regular expression, and the error message must match in order to consider the error as expected."

### State Check Strategy

Tests should separate concerns:
- Create dedicated functions to verify remote resource existence
- Use custom state checks for scenario-specific validations
- Leverage built-in checks like `ExpectKnownValue` for state verification
- Framework provides type-specific matchers: Bool, Float32, Float64, Int32, Int64, List, Map, Number, Object, Set, String, Tuple, Null, and NotNull

### Regression Testing Best Practice

HashiCorp recommends a two-commit approach:
1. Introduce test reproducing the bug
2. Modify code to fix the issue

This allows independent verification that tests accurately capture problems before fixes are applied.

### Framework Features Summary

- **PreCheck Functions**: Run before test execution for environment validation
- **CheckDestroy Functions**: Verify proper cleanup after test completion
- **Built-in Behaviors**: Framework automatically validates that final plans produce no diffs unless explicitly permitted via `ExpectNonEmptyPlan`
- **Convention Over Configuration**: Emphasizes composable test components and separation of concerns (remote vs. local verification)

## Research Tasks

### Task 0.1: Analyze Current Test Quality Baseline

**Action**: Review all 6 test files and calculate current quality scores per file.

**Analysis Framework**:
```text
Quality Score = (Modern Patterns × 40%) + (Verification Completeness × 30%) +
                (Environment Portability × 20%) + (Best Practices × 10%)

Modern Patterns (40%):
  - Percentage of tests using statecheck.ExpectKnownValue()
  - Percentage of tests using plancheck.ExpectEmptyPlan()
  - Percentage of tests using statecheck.CompareValue() for ID tracking

Verification Completeness (30%):
  - Percentage of schema attributes verified with state checks
  - Filter verification present for data source tests

Environment Portability (20%):
  - No hardcoded resource counts (networks.# = "3")
  - No hardcoded resource names (networks.0.name = "managementnet")
  - Tests pass on different cluster configurations

Best Practices (10%):
  - ID consistency tracking across CRUD operations
  - Enhanced CheckDestroy with detailed error messages
  - Validation tests present
  - Unique resource names with generateUniqueTestName()
```

**Current Baseline** (from spec):
- resource_cmpart_softwareimage_test.go: 85% (12 tests)
- resource_cmdevice_category_test.go: 80% (6 tests)
- data_source_cmdevice_nodes_test.go: 40% (4 tests)
- data_source_cmnet_networks_test.go: 60% (4 tests)
- data_source_cmpart_softwareimages_test.go: 50% (5 tests)
- data_source_cmdevice_categories_test.go: 75% (4 tests)
- **Overall**: 69% average

**Expected Outcome**: Document current state and gap analysis showing what patterns are missing per file.

### Task 0.2: Research Modern State Check Patterns

**Action**: Review HashiCorp documentation and terraform-provider-design skill for state check best practices.

**HashiCorp Guidance**: "To verify attributes are saved to the state file correctly, use a combination of the built-in `statecheck` implementations." State checks verify both local (state file) and remote values using composable verification components.

**Key Patterns to Document**:

1. **ExpectKnownValue Pattern** - Type-safe attribute verification:
```go
import (
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_cmpart_softwareimage.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact(imageName),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmpart_softwareimage.test",
        tfjsonpath.New("enable_sol"),
        knownvalue.Bool(true),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmpart_softwareimage.test",
        tfjsonpath.New("uuid"),
        knownvalue.NotNull(),
    ),
}
```

2. **CompareValue Pattern** - ID consistency tracking:
```go
compareID := statecheck.CompareValue(compare.ValuesSame())

Steps: []resource.TestStep{
    {
        Config: config,
        ConfigStateChecks: []statecheck.StateCheck{
            compareID.AddStateValue("bcm_resource.test", tfjsonpath.New("id")),
        },
    },
    {
        ResourceName:      "bcm_resource.test",
        ImportState:       true,
        ImportStateVerify: true,
        ConfigStateChecks: []statecheck.StateCheck{
            compareID.AddStateValue("bcm_resource.test", tfjsonpath.New("id")),
        },
    },
}
```

3. **Known Value Matchers** - Type mappings for BCM provider:
```go
// String attributes
knownvalue.StringExact(value)    // name, path, notes, kernel_parameters

// Boolean attributes
knownvalue.Bool(value)            // enable_sol, dhcp_enabled, install_boot_record

// Numeric attributes
knownvalue.Int64Exact(value)      // sol_speed (115200)

// Computed fields
knownvalue.NotNull()              // uuid, id, creation_time

// Collections
knownvalue.ListExact([]knownvalue.Check{...})    // modules
knownvalue.ListSizeExact(n)                       // modules.# count verification
```

**Expected Outcome**: Comprehensive mapping of BCM provider attribute types to knownvalue matchers, with examples for each test file.

### Task 0.3: Research Plan Check Patterns

**Action**: Document idempotency verification patterns using plancheck.ExpectEmptyPlan().

**HashiCorp Guidance**: Basic attribute verification tests should confirm "subsequent plans produce no differences." The framework automatically validates final plans produce no diffs unless explicitly permitted via `ExpectNonEmptyPlan`. Idempotency checks ensure resources don't produce spurious diffs on re-apply.

**Key Patterns to Document**:

1. **Post-Create Idempotency**:
```go
Steps: []resource.TestStep{
    // Create resource
    {
        Config: config("initial"),
        Check: resource.ComposeAggregateTestCheckFunc(...),
    },
    // Verify no changes on re-apply
    {
        Config: config("initial"),
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectEmptyPlan(),
            },
        },
    },
}
```

2. **Post-Update Idempotency**:
```go
Steps: []resource.TestStep{
    // Update resource
    {
        Config: config("updated"),
        Check: resource.ComposeAggregateTestCheckFunc(...),
    },
    // Verify no changes on re-apply
    {
        Config: config("updated"),
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectEmptyPlan(),
            },
        },
    },
}
```

3. **Eventual Consistency Handling**:
```go
// For async operations like image cloning
{
    Config: config,
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
    // Note: Read operation includes polling for clone completion
    // Test should pass after Read resolves eventual consistency
}
```

**Expected Outcome**: Idempotency verification pattern applicable to all 18 resource tests (2 resources × ~9 tests each).

### Task 0.4: Research Filter Verification Patterns

**Action**: Document patterns for verifying data source filter correctness.

**HashiCorp Guidance**: Tests should separate concerns with "dedicated functions to verify remote resource existence" and "custom state checks for scenario-specific validations." For data sources with filters, verify that filtered results actually match the filter criteria, not just that results exist.

**Key Patterns to Document**:

1. **Single Element Verification**:
```go
// Verify first element matches filter criteria
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "data.bcm_cmdevice_nodes.test",
        tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("node_type"),
        knownvalue.StringExact("PhysicalNode"),
    ),
}
```

2. **Multiple Element Verification** (when result count known):
```go
// If test creates resources, can verify all elements
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "data.bcm_cmnet_networks.filtered",
        tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("dhcp_enabled"),
        knownvalue.Bool(true),
    ),
    statecheck.ExpectKnownValue(
        "data.bcm_cmnet_networks.filtered",
        tfjsonpath.New("networks").AtSliceIndex(1).AtMapKey("dhcp_enabled"),
        knownvalue.Bool(true),
    ),
}
```

3. **Pattern Matching** (for name filters):
```go
// Verify name contains pattern using StringRegexp
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "data.bcm_cmnet_networks.filtered",
        tfjsonpath.New("networks").AtSliceIndex(0).AtMapKey("name"),
        knownvalue.StringRegexp(regexp.MustCompile(".*management.*")),
    ),
}
```

**Expected Outcome**: Filter verification approach for each data source with filters (nodes: node_type/hostname, networks: name_pattern/dhcp_enabled, software_images: category).

### Task 0.5: Research Environment Portability Patterns

**Action**: Document patterns for removing environment-specific assumptions.

**Anti-Patterns to Remove**:
```go
// DON'T: Hardcode exact counts
resource.TestCheckResourceAttr("data.bcm_cmnet_networks.test", "networks.#", "3")

// DON'T: Assume specific resource names
resource.TestCheckResourceAttr("data.bcm_cmnet_networks.filtered", "networks.0.name", "managementnet")

// DON'T: Hardcode filter result counts
resource.TestCheckResourceAttr("data.bcm_cmnet_networks.dhcp", "networks.#", "2")
```

**Portable Patterns to Adopt**:
```go
// DO: Check for non-empty results
resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.#")

// DO: Use state check to verify list not empty
statecheck.ExpectKnownValue(
    "data.bcm_cmnet_networks.test",
    tfjsonpath.New("networks"),
    knownvalue.ListSizeExact(0).Not(), // Not empty
)

// DO: Create test resources with unique names
testNetworkName := generateUniqueTestName("test-network")

// DO: Verify filter logic works regardless of result count
// If filter matches 0 items: networks.# = "0" (valid)
// If filter matches 1+ items: verify first element matches criteria
```

**Expected Outcome**: Refactoring strategy for data_source_cmnet_networks_test.go (highest impact: 3 hardcoded values) and other data source tests.

### Task 0.6: Research CheckDestroy Enhancement Patterns

**Action**: Document enhanced error reporting patterns for CheckDestroy functions.

**Current Pattern** (basic):
```go
func testAccCheckResourceDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_resource" { continue }
        deleted, _ := verifyResourceDeleted(ctx, client, "Service", "getMethod", rs.Primary.ID, 4)
        if !deleted { return fmt.Errorf("resource still exists") }
    }
    return nil
}
```

**Enhanced Pattern** (detailed errors):
```go
func testAccCheckResourceDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    ctx := context.Background()
    var errors []string

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_resource" { continue }

        deleted, err := verifyResourceDeleted(ctx, client, "Service", "getMethod", rs.Primary.ID, 4)
        if !deleted {
            if err != nil {
                errors = append(errors, fmt.Sprintf(
                    "Resource %s (ID: %s) failed to delete: %v",
                    rs.Type, rs.Primary.ID, err,
                ))
            } else {
                errors = append(errors, fmt.Sprintf(
                    "Resource %s (ID: %s) still exists after %d retries (15s)",
                    rs.Type, rs.Primary.ID, 4,
                ))
            }
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failed:\n- %s", strings.Join(errors, "\n- "))
    }
    return nil
}
```

**Expected Outcome**: Enhanced CheckDestroy pattern applicable to 2 resource test files (software_image, category).

## Research Deliverable

`research.md` containing:
1. Current quality baseline analysis with gap identification per file
2. Modern state check pattern library with BCM provider type mappings
3. Idempotency verification pattern for resource tests
4. Filter verification approach for data source tests
5. Environment portability refactoring strategy
6. Enhanced CheckDestroy error reporting pattern
7. Import requirements and code snippets for each pattern

**Success Criteria**:
- All patterns documented with working code examples
- BCM provider attribute types mapped to knownvalue matchers
- Clear migration path from legacy to modern patterns
- Backward compatibility maintained (no test scenario removal)

---

# Phase 1: Design & Contracts

**Goal**: Generate detailed design artifacts for test modernization including data model, API contracts, and quickstart guide.

**Prerequisites**: `research.md` complete with modern testing patterns documented

**Output**:
- `data-model.md` - Test entities and pattern mappings
- `contracts/` - Example test structures for each pattern
- `quickstart.md` - Developer guide for applying modern patterns

## Phase 1 Tasks

### Task 1.1: Generate Data Model

**Action**: Create `data-model.md` mapping test entities to modern patterns.

**Entities to Document**:

1. **Test Step Configuration**:
```yaml
TestStep:
  - config: string (Terraform configuration)
  - check: TestCheckFunc (legacy assertions)
  - configStateChecks: []StateCheck (modern state verification)
  - configPlanChecks: ConfigPlanChecks (plan verification)
  - resourceName: string (for ImportState)
  - importState: bool
  - importStateVerify: bool
  - preConfig: func() (for drift detection)
```

2. **State Check Entity**:
```yaml
StateCheck:
  ExpectKnownValue:
    - resourceAddress: string ("bcm_cmpart_softwareimage.test")
    - attributePath: tfjsonpath (tfjsonpath.New("name"))
    - matcher: knownvalue.Check (knownvalue.StringExact("value"))

  CompareValue:
    - comparison: compare.ValueComparison (compare.ValuesSame())
    - stateValues: []StateValue (tracked across steps)
```

3. **Plan Check Entity**:
```yaml
PlanCheck:
  ExpectEmptyPlan:
    - phase: PreApply
    - usage: Idempotency verification

  ExpectNonEmptyPlan:
    - phase: PreApply
    - usage: Drift detection verification
```

4. **Known Value Matcher Entity**:
```yaml
KnownValueMatcher:
  StringExact:
    - bcmTypes: [name, path, notes, kernel_parameters]
  Bool:
    - bcmTypes: [enable_sol, dhcp_enabled, install_boot_record]
  Int64Exact:
    - bcmTypes: [sol_speed]
  NotNull:
    - bcmTypes: [uuid, id, creation_time]
  ListExact:
    - bcmTypes: [modules]
  ListSizeExact:
    - bcmTypes: [modules.#, networks.#, nodes.#]
```

5. **Filter Verification Entity**:
```yaml
FilterVerification:
  DataSource: bcm_cmdevice_nodes
  Filters:
    - node_type: string
      verification: ExpectKnownValue(nodes.0.node_type, StringExact(filterValue))
    - hostname_pattern: string
      verification: ExpectKnownValue(nodes.0.hostname, StringRegexp(pattern))

  DataSource: bcm_cmnet_networks
  Filters:
    - name_pattern: string
      verification: ExpectKnownValue(networks.0.name, StringRegexp(pattern))
    - dhcp_enabled: bool
      verification: ExpectKnownValue(networks.0.dhcp_enabled, Bool(filterValue))
```

**Expected Outcome**: Complete entity model showing how legacy patterns map to modern patterns, with BCM provider attribute type mappings.

### Task 1.2: Generate API Contracts

**Action**: Create `contracts/` directory with example test structures for each pattern.

**Contract Files to Create**:

1. **`contracts/resource_state_checks_example.go`**:
```go
// Example: Modern resource test with state checks
func TestAccExampleResource_StateChecks(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: testConfig("test"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Keep legacy checks for backward compatibility initially
                    resource.TestCheckResourceAttr("example.test", "name", "test"),
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    // Add modern state checks
                    statecheck.ExpectKnownValue(
                        "example.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact("test"),
                    ),
                    statecheck.ExpectKnownValue(
                        "example.test",
                        tfjsonpath.New("uuid"),
                        knownvalue.NotNull(),
                    ),
                },
            },
        },
    })
}
```

2. **`contracts/resource_idempotency_example.go`**:
```go
// Example: Idempotency verification with plan checks
func TestAccExampleResource_Idempotency(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Create
            {
                Config: testConfig("initial"),
                Check: resource.ComposeAggregateTestCheckFunc(...),
            },
            // Verify idempotency
            {
                Config: testConfig("initial"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
        },
    })
}
```

3. **`contracts/resource_id_tracking_example.go`**:
```go
// Example: ID consistency tracking across CRUD operations
func TestAccExampleResource_IDTracking(t *testing.T) {
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Create - capture ID
            {
                Config: testConfig("test"),
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("example.test", tfjsonpath.New("id")),
                },
            },
            // Import - verify ID matches
            {
                ResourceName:      "example.test",
                ImportState:       true,
                ImportStateVerify: true,
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("example.test", tfjsonpath.New("id")),
                },
            },
            // Update - verify ID unchanged
            {
                Config: testConfig("updated"),
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("example.test", tfjsonpath.New("id")),
                },
            },
        },
    })
}
```

4. **`contracts/data_source_filter_verification_example.go`**:
```go
// Example: Data source filter verification
func TestAccExampleDataSource_FilterVerification(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            {
                Config: testConfigWithFilter("node_type", "PhysicalNode"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Legacy: Just verify results exist
                    resource.TestCheckResourceAttrSet("data.example.test", "items.#"),
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    // Modern: Verify filter correctness
                    statecheck.ExpectKnownValue(
                        "data.example.test",
                        tfjsonpath.New("items").AtSliceIndex(0).AtMapKey("node_type"),
                        knownvalue.StringExact("PhysicalNode"),
                    ),
                },
            },
        },
    })
}
```

5. **`contracts/check_destroy_enhanced_example.go`**:
```go
// Example: Enhanced CheckDestroy with detailed error messages
func testAccCheckExampleDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    ctx := context.Background()
    var errors []string

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_example" { continue }

        deleted, err := verifyResourceDeleted(ctx, client, "Service", "getMethod", rs.Primary.ID, 4)
        if !deleted {
            if err != nil {
                errors = append(errors, fmt.Sprintf(
                    "Resource %s (ID: %s) failed to delete: %v",
                    rs.Type, rs.Primary.ID, err,
                ))
            } else {
                errors = append(errors, fmt.Sprintf(
                    "Resource %s (ID: %s) still exists after 4 retries (15s)",
                    rs.Type, rs.Primary.ID,
                ))
            }
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failed:\n- %s", strings.Join(errors, "\n- "))
    }
    return nil
}
```

**Expected Outcome**: 5 contract files demonstrating each modern pattern with complete, runnable examples.

### Task 1.3: Generate Quickstart Guide

**Action**: Create `quickstart.md` for developers applying modern patterns.

**Quickstart Sections**:

1. **Prerequisites**:
```markdown
- Go 1.24.0+
- terraform-plugin-testing v1.13.3
- BCM cluster access (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- Understanding of RED-GREEN-REFACTOR TDD cycle
```

2. **Import Requirements**:
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

3. **Quick Reference - BCM Attribute Type Mappings**:
```markdown
| BCM Attribute Type | Legacy Pattern | Modern Pattern |
|-------------------|----------------|----------------|
| String (name, path) | TestCheckResourceAttr | ExpectKnownValue + StringExact |
| Boolean (enable_sol) | TestCheckResourceAttr("true") | ExpectKnownValue + Bool(true) |
| Int64 (sol_speed) | TestCheckResourceAttr("115200") | ExpectKnownValue + Int64Exact(115200) |
| UUID/ID (computed) | TestCheckResourceAttrSet | ExpectKnownValue + NotNull() |
| List (modules) | TestCheckResourceAttr("modules.#", "2") | ExpectKnownValue + ListSizeExact(2) |
```

4. **Step-by-Step Modernization Workflow**:
```markdown
### Step 1: Add State Checks to Existing Test

1. Identify test step with legacy `Check` assertions
2. Add `ConfigStateChecks: []statecheck.StateCheck{}` to test step
3. For each `TestCheckResourceAttr`, add corresponding `ExpectKnownValue`
4. Run test to verify both legacy and modern checks pass
5. (Optional) Remove redundant legacy checks after verification

### Step 2: Add Idempotency Verification

1. After Create step, add new step with same config
2. Add `ConfigPlanChecks` with `ExpectEmptyPlan()`
3. After Update step, add new step with same config
4. Add `ConfigPlanChecks` with `ExpectEmptyPlan()`
5. Run test to verify no spurious diffs

### Step 3: Add ID Consistency Tracking

1. Before test steps, create: `compareID := statecheck.CompareValue(compare.ValuesSame())`
2. In Create step, add: `compareID.AddStateValue("resource.test", tfjsonpath.New("id"))`
3. In ImportState step, add: `compareID.AddStateValue("resource.test", tfjsonpath.New("id"))`
4. In Update step, add: `compareID.AddStateValue("resource.test", tfjsonpath.New("id"))`
5. Run test to verify ID remains consistent

### Step 4: Add Filter Verification (Data Sources)

1. Identify data source test with filter
2. Add `ConfigStateChecks` to test step
3. Add `ExpectKnownValue` for filtered attribute on first element
4. Use appropriate matcher (StringExact, Bool, StringRegexp)
5. Run test to verify filter correctness

### Step 5: Enhance CheckDestroy

1. Locate `testAccCheck<Resource>Destroy` function
2. Replace simple error return with error accumulation pattern
3. Add detailed error messages with resource type and ID
4. Run test to verify enhanced error reporting
```

5. **Running Modernized Tests**:
```bash
# Run all acceptance tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Run specific test file
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImage

# Run with detailed logging for debugging
TF_LOG=TRACE TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic
```

6. **Troubleshooting Common Issues**:
```markdown
**Issue**: Test fails with "invalid result object" error
**Cause**: Unknown value propagated to state
**Fix**: Ensure computed fields resolve to known values (null or actual value)

**Issue**: ExpectEmptyPlan fails unexpectedly
**Cause**: Computed field changes on every refresh (non-idempotent)
**Fix**: Investigate Read operation, ensure it preserves plan values correctly

**Issue**: Filter verification fails
**Cause**: Filter returns empty list (valid), but ExpectKnownValue tries to access [0]
**Fix**: Add check for list size > 0 before accessing elements, or use conditional checks
```

**Expected Outcome**: Complete quickstart guide enabling developers to apply modern patterns to any test file in ~30 minutes per test.

### Task 1.4: Update Agent Context

**Action**: Run `.specify/scripts/bash/update-agent-context.sh copilot` to update AI agent context files.

**Context Updates**:
- Add terraform-plugin-testing v1.13.3 modern patterns
- Add statecheck, plancheck, knownvalue package usage
- Add BCM provider attribute type mappings
- Add filter verification approach for data sources
- Preserve existing BCM API patterns and test helper documentation

**Expected Outcome**: Updated agent context file (copilot-specific) with modern testing pattern knowledge.

### Task 1.5: Re-evaluate Constitution Check

**Action**: Review Phase 1 design against constitution principles.

**Verification**:
- ✅ Test-First Development: Maintained (enhancing existing tests, not adding new features)
- ✅ Simplicity: No new abstractions, using standard terraform-plugin-testing patterns
- ✅ Parallel Execution: Design supports parallel RED-GREEN-REFACTOR cycles
- ✅ No New Violations: Design follows established HashiCorp patterns

**Expected Outcome**: Constitution check still PASS, proceed to Phase 2.

## Phase 1 Deliverables

1. `data-model.md` - Test entities and pattern mappings
2. `contracts/` - 5 example contract files demonstrating each pattern
3. `quickstart.md` - Developer guide with step-by-step workflow
4. Updated agent context file with modern testing patterns
5. Constitution check verification (re-run, should still pass)

**Success Criteria**:
- All Phase 0 research patterns translated to design artifacts
- Contract examples are complete and runnable
- Quickstart guide enables any developer to modernize tests
- Agent context updated with new patterns

---

# Phase 2: Planning (Generated by `/speckit.tasks`)

**Note**: This phase is executed by the `/speckit.tasks` command, NOT by `/speckit.plan`. The plan command stops here.

**Goal**: Generate actionable, dependency-ordered task list in `tasks.md` for implementation.

**Task Generation Approach**:

1. **Phase 1: HIGH PRIORITY - Resource State Checks + Idempotency** (2 files, 18 tests)
   - Task: Modernize resource_cmpart_softwareimage_test.go (12 tests)
   - Task: Modernize resource_cmdevice_category_test.go (6 tests)
   - TDD Cycle: RED (add failing modern checks) → GREEN (fix any issues) → REFACTOR (remove redundant legacy checks)

2. **Phase 2: HIGH PRIORITY - Data Source Filter Verification** (3 files, 13 tests)
   - Task: Fix data_source_cmnet_networks_test.go (remove hardcoded values, add filter verification)
   - Task: Enhance data_source_cmdevice_nodes_test.go (add filter verification)
   - Task: Enhance data_source_cmpart_softwareimages_test.go (add filter verification)
   - TDD Cycle: RED (add failing filter checks) → GREEN (verify filters work) → REFACTOR (cleanup)

3. **Phase 3: MEDIUM PRIORITY - Validation + CheckDestroy** (2 files)
   - Task: Add validation tests to resource_cmpart_softwareimage_test.go
   - Task: Enhance CheckDestroy in resource_cmpart_softwareimage_test.go
   - Task: Enhance CheckDestroy in resource_cmdevice_category_test.go
   - TDD Cycle: RED (add validation test, enhance CheckDestroy) → GREEN (verify) → REFACTOR

4. **Phase 4: LOW PRIORITY - Documentation + Cleanup** (1 file)
   - Task: Update data_source_cmdevice_categories_test.go (enhanced state checks)
   - Task: Update CLAUDE.md with modern testing patterns
   - Task: Final test suite quality verification (calculate scores, ensure 90%+ average)

**Parallel Execution Opportunities**:
- Phase 1: Both resource test files can be modernized in parallel
- Phase 2: All three data source files can be enhanced in parallel
- Phase 3: Both CheckDestroy enhancements can be done in parallel

**Success Criteria for Phase 2**:
- `tasks.md` generated with complete task breakdown
- Each task has clear acceptance criteria from spec
- Tasks ordered by priority and dependencies
- Parallel execution opportunities identified

---

# Implementation Notes

## TDD Workflow for Each Test File

```text
RED PHASE:
1. Add modern state checks (ConfigStateChecks with ExpectKnownValue)
2. Add plan checks for idempotency (ConfigPlanChecks with ExpectEmptyPlan)
3. Add ID tracking (CompareValue across test steps)
4. Add filter verification (for data sources)
5. Run tests - expect some may fail initially due to existing issues

GREEN PHASE:
1. Fix any test failures revealed by modern checks
2. Verify all tests pass with both legacy and modern assertions
3. Run full acceptance test suite to ensure no regressions

REFACTOR PHASE:
1. (Optional) Remove redundant legacy TestCheckResourceAttr calls
2. Add comments explaining modern pattern usage
3. Verify tests still pass after cleanup
4. Calculate updated quality score for file
```

## Quality Score Targets

| Test File | Current | Target | Key Improvements |
|-----------|---------|--------|------------------|
| resource_cmpart_softwareimage_test.go | 85% | 95%+ | State checks, idempotency, ID tracking |
| resource_cmdevice_category_test.go | 80% | 95%+ | State checks, idempotency, ID tracking |
| data_source_cmdevice_nodes_test.go | 40% | 90%+ | State checks, filter verification |
| data_source_cmnet_networks_test.go | 60% | 90%+ | Remove hardcoded values, filter verification |
| data_source_cmpart_softwareimages_test.go | 50% | 90%+ | State checks, filter verification |
| data_source_cmdevice_categories_test.go | 75% | 90%+ | Enhanced state checks |
| **Overall Average** | **69%** | **90%+** | Modern patterns across all tests |

## Backward Compatibility Strategy

- Keep legacy `TestCheckResourceAttr` assertions initially
- Run both legacy and modern checks side-by-side during migration
- Remove redundant legacy checks only after modern checks proven stable
- All existing test scenarios must continue passing (no coverage regression)

## Test Execution Time Budget

- Current baseline: Unknown (need to measure)
- Acceptable: Current + 10% (additional state/plan checks have minimal overhead)
- If exceeded: Review for unnecessary delays, optimize polling intervals

## Environment Portability Verification

**Before declaring success**, run tests against:
1. Original BCM cluster configuration
2. Modified cluster with different network counts
3. Modified cluster with different node types
4. Fresh cluster with minimal configuration

**All tests must pass** on all configurations, or skip gracefully with clear messages.

## Key Decisions

1. **Keep legacy checks initially**: Reduces risk during migration, allows gradual rollout
2. **Parallel execution for independent tests**: Resource tests (2 files) and data source tests (3 files) can be modernized concurrently
3. **Idempotency checks for all resource tests**: Every Create and Update must have follow-up ExpectEmptyPlan step
4. **Filter verification for all data source filter tests**: Every filter must verify results match criteria
5. **Enhanced CheckDestroy for all resources**: Detailed error messages improve debugging experience

---

# Branch and Artifact Locations

**Branch**: `001-modernize-test-suite`

**Design Artifacts** (generated by `/speckit.plan`):
- `/workspace/specs/001-modernize-test-suite/plan.md` (this file)
- `/workspace/specs/001-modernize-test-suite/research.md` (Phase 0 output)
- `/workspace/specs/001-modernize-test-suite/data-model.md` (Phase 1 output)
- `/workspace/specs/001-modernize-test-suite/quickstart.md` (Phase 1 output)
- `/workspace/specs/001-modernize-test-suite/contracts/` (Phase 1 output)

**Implementation Artifacts** (generated by `/speckit.tasks`):
- `/workspace/specs/001-modernize-test-suite/tasks.md` (Phase 2 output)

**Source Files to Modify** (during implementation):
- `/workspace/internal/provider/resource_cmpart_softwareimage_test.go`
- `/workspace/internal/provider/resource_cmdevice_category_test.go`
- `/workspace/internal/provider/data_source_cmdevice_nodes_test.go`
- `/workspace/internal/provider/data_source_cmnet_networks_test.go`
- `/workspace/internal/provider/data_source_cmpart_softwareimages_test.go`
- `/workspace/internal/provider/data_source_cmdevice_categories_test.go`

**No Changes Required**:
- `/workspace/internal/provider/test_helpers.go` (already optimal)
- `/workspace/internal/provider/provider_test.go` (test setup is fine)
- `/workspace/internal/provider/bcm_client.go` (API client works correctly)
- Resource/data source implementations (only test code changes)
