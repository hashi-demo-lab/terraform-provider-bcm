# TERRAFORM PROVIDER TDD GUIDE

## Core Principles

**TDD Golden Rule**: RED-GREEN-REFACTOR in parallel batches where possible.

- 🔴 **Red**: Write failing acceptance tests first
- 🟢 **Green**: Write minimal CRUD code to pass tests
- 🔄 **Refactor**: Improve code while keeping tests green

**Stack**: Terraform Plugin Framework v1.16+ | Go 1.24+ | terraform-plugin-testing

Use `terraform-provider-tests` skill for design principles and best practices.
if you need to geenrate a report and don't have a given path then use /workspace/ai_reports/

## Parallel Execution Pattern

```go
// RED: Multiple failing tests
- Write("internal/provider/resource_a_test.go", failingTest)
- Write("internal/provider/resource_b_test.go", failingTest)
- Bash("TF_ACC=1 go test -v ./internal/provider/")

// GREEN: Multiple minimal implementations
- Write("internal/provider/resource_a.go", minimalImpl)
- Write("internal/provider/resource_b.go", minimalImpl)
- Bash("TF_ACC=1 go test -v ./internal/provider/")

// REFACTOR: Improve quality
- Edit("internal/provider/resource_a.go", refactoredCode)
- Edit("internal/provider/resource_b.go", refactoredCode)
- Bash("TF_ACC=1 go test -v ./internal/provider/")
```

## Acceptance Test Structure

```go
func TestAccResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {Config: testConfig("initial"), Check: resource.ComposeAggregateTestCheckFunc(...)},
            {ResourceName: "bcm_resource.test", ImportState: true, ImportStateVerify: true},
            {Config: testConfig("updated"), Check: resource.ComposeAggregateTestCheckFunc(...)},
        },
    })
}
```

**Required Tests**: Create, Read, Update, Delete, Import, Drift detection

## Test Helpers Pattern

```go
// test_helpers.go - shared across all tests
func createTestBCMClient(t *testing.T) *BCMClient { /* ... */ }
func getResourceUUIDByName(t *testing.T, service, method, name string) string { /* ... */ }
func verifyResourceDeleted(ctx, client, service, method, id string, retries int) (bool, error) { /* ... */ }
func generateUniqueTestName(prefix string) string { /* timestamp-based unique names */ }
```

## CheckDestroy Pattern

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

## Drift Detection Test Pattern

```go
Steps: []resource.TestStep{
    {Config: config("initial"), Check: resource.TestCheckResourceAttr(...)},
    {
        PreConfig: func() {
            // Modify resource externally via API
            client := createTestBCMClient(t)
            uuid := getResourceUUIDByName(t, "Service", "getMethod", name)
            // Update via API, wait for consistency
        },
        Config: config("initial"),
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{plancheck.ExpectNonEmptyPlan()},
        },
    },
    {Config: config("initial"), Check: resource.TestCheckResourceAttr(...)},
}
```

## Anti-Patterns to Avoid

❌ Skipping ImportState/Drift tests
❌ Hardcoded test values (use unique names)
❌ Incomplete CRUD testing
❌ Tests dependent on external state
❌ Missing or outdated documentation

## Example Testing Infrastructure

**Location**: `/workspace/scripts/test-examples.sh`

**Execution Strategy**:

- Data sources: Parallel (4 concurrent, ~10s for 10 examples)
- Resources: Sequential (~19s for 7 examples)

**Run Tests**:

```bash
BCM_ENDPOINT="https://..." BCM_USERNAME="..." BCM_PASSWORD="..." \
SKIP_BUILD=true ./scripts/test-examples.sh

# Options: --data-sources-only, --resources-only, --verbose, --no-cleanup
```

**Test Phases**: Environment validation → Provider build → Example testing (init/validate/plan) → Cleanup

**Resource Naming**: Use "citest" prefix for automatic cleanup identification.

**Provider Pattern**: Examples use environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD) - no hardcoded credentials.

**Discovery**: Automatically finds all `.tf` files in `examples/{data-sources,resources}/*/`

## Environment Variables

```bash
# Development
TF_ACC=0                    # Disable acceptance tests
TF_LOG=DEBUG               # Enable logging

# Acceptance Testing
TF_ACC=1                    # Enable acceptance tests
BCM_ENDPOINT="https://..."  # API endpoint
BCM_USERNAME="root"         # Username
BCM_PASSWORD="..."          # Password
TF_LOG=TRACE               # Max verbosity
```

## Quality Metrics

- Acceptance test pass rate: 100%
- All CRUD operations tested
- Import functionality verified
- Drift detection validated
- Documentation auto-generated and in sync

## Test Recommendations

Provider-defined functions should test:

- Known values return expected results
- Null values for collection elements/object attributes
- `AllowNullValue` parameters handle nulls
- `AllowUnknownValues` parameters handle unknowns
- Validation errors handled correctly
