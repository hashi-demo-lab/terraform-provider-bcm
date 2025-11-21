# Feature Specification: Modernize Terraform BCM Provider Test Suite

**Feature Branch**: `001-modernize-test-suite`
**Created**: 2025-11-21
**Status**: Draft
**Input**: User description: "Create a comprehensive feature specification for modernizing and enhancing the Terraform BCM provider test suite to achieve 90%+ quality score by implementing HashiCorp's modern testing patterns from terraform-plugin-testing v1.5.0+."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Modern State Verification for Resources (Priority: P1)

As a Terraform provider developer, I need resource tests that use type-safe state verification with `statecheck.ExpectKnownValue()` and ID consistency tracking with `statecheck.CompareValue()`, so that I can catch type mismatches and state corruption issues that legacy `TestCheckResourceAttr()` calls would miss.

**Why this priority**: This is the foundation for reliable testing. Legacy string-based assertions miss type errors, null/unknown value issues, and ID consistency problems that cause real-world bugs in production Terraform configurations.

**Independent Test**: Can be fully tested by converting resource tests to use modern state checks and verifying they catch issues like type mismatches (e.g., boolean stored as string) and ID changes across operations.

**Acceptance Scenarios**:

1. **Given** a software image resource test, **When** converted to use `statecheck.ExpectKnownValue()` for name/path/UUID verification, **Then** the test passes and provides better type safety than legacy assertions
2. **Given** a category resource test with Create/Update/Import operations, **When** converted to use `statecheck.CompareValue()` to track ID consistency, **Then** the test detects if resource ID changes unexpectedly between operations
3. **Given** a resource test checking boolean attributes, **When** converted to use `knownvalue.Bool()` instead of string comparison, **Then** the test catches if BCM API returns "true" string instead of boolean true
4. **Given** a resource test checking list attributes, **When** converted to use `knownvalue.ListExact()`, **Then** the test verifies exact list contents including order and element types

---

### User Story 2 - Idempotency Verification for Resources (Priority: P1)

As a Terraform provider developer, I need resource tests that verify idempotency using `plancheck.ExpectEmptyPlan()` after Create and Update operations, so that I can ensure Terraform doesn't generate unnecessary change plans when re-applying the same configuration.

**Why this priority**: Idempotency is a core Terraform principle. Resources that aren't idempotent cause unnecessary API calls, confuse users with spurious diffs, and can trigger accidental destructive changes in production.

**Independent Test**: Can be fully tested by adding post-Create and post-Update test steps with `ExpectEmptyPlan()` checks and verifying they catch non-idempotent behavior like computed fields changing on every refresh.

**Acceptance Scenarios**:

1. **Given** a software image resource test with Create step, **When** a second test step re-applies the same config with `plancheck.ExpectEmptyPlan()`, **Then** Terraform shows no planned changes
2. **Given** a category resource test with Update step, **When** a follow-up step re-applies the updated config with `plancheck.ExpectEmptyPlan()`, **Then** Terraform shows no planned changes
3. **Given** a resource with computed fields that BCM API modifies, **When** idempotency test runs, **Then** the test fails and reveals which computed fields are causing drift
4. **Given** a resource with cloning operation (software image), **When** idempotency test runs after clone completes, **Then** the test verifies eventual consistency is handled correctly

---

### User Story 3 - Verified Data Source Filtering (Priority: P2)

As a Terraform provider developer, I need data source tests that verify filter correctness by checking that returned results actually match the filter criteria, so that I can ensure users get the data they expect when applying filters.

**Why this priority**: Current tests only verify that filters don't error and return some results, but don't verify the results are correct. This can hide bugs where filters are silently ignored or applied incorrectly.

**Independent Test**: Can be fully tested by enhancing data source filter tests to use `statecheck.ExpectKnownValue()` with pattern matching or conditional checks on returned attributes, ensuring filtered results match criteria.

**Acceptance Scenarios**:

1. **Given** a nodes data source filtered by `node_type = "PhysicalNode"`, **When** test verifies returned nodes using `statecheck.ExpectKnownValue()`, **Then** all returned nodes have `node_type` matching "PhysicalNode"
2. **Given** a networks data source filtered by `name_pattern = "management"`, **When** test verifies returned networks, **Then** all network names contain "management"
3. **Given** a networks data source filtered by `dhcp_enabled = true`, **When** test verifies returned networks, **Then** all networks have `dhcp_enabled = true`
4. **Given** a software images data source filtered by category, **When** test verifies returned images, **Then** all images belong to the specified category

---

### User Story 4 - Environment-Agnostic Test Data (Priority: P2)

As a Terraform provider developer, I need tests that work on any BCM cluster without hardcoded assumptions about cluster state (specific network names, exact counts, etc.), so that tests pass consistently in CI/CD, local dev, and different customer environments.

**Why this priority**: Hardcoded assumptions make tests brittle and environment-dependent. This blocks running tests in CI/CD, makes local development painful, and prevents customers from running acceptance tests in their own environments.

**Independent Test**: Can be fully tested by removing hardcoded values from network tests, running tests against different BCM clusters, and verifying they pass regardless of cluster-specific configuration.

**Acceptance Scenarios**:

1. **Given** network data source tests with hardcoded `networks.# = "3"`, **When** refactored to check `networks.# > 0`, **Then** tests pass on clusters with different network counts
2. **Given** network filter test expecting specific network name "managementnet", **When** refactored to create test network or use dynamic lookup, **Then** test passes on any cluster
3. **Given** DHCP filter test expecting exactly 2 networks, **When** refactored to verify filter logic without hardcoded counts, **Then** test passes regardless of cluster DHCP configuration
4. **Given** all data source tests, **When** run against fresh BCM cluster with minimal config, **Then** tests pass or skip gracefully with clear messages

---

### User Story 5 - Enhanced CheckDestroy Validation (Priority: P3)

As a Terraform provider developer, I need enhanced `CheckDestroy` functions that use modern state verification and provide detailed error messages about what resources failed to delete and why, so that I can quickly diagnose cleanup failures in acceptance tests.

**Why this priority**: Current CheckDestroy functions have basic error handling. Better error messages and verification help debug test failures faster, especially in CI/CD where you can't interactively inspect cluster state.

**Independent Test**: Can be fully tested by enhancing CheckDestroy functions to use verifyResourceDeleted helper with detailed logging, and verifying error messages clearly identify which resources and IDs failed cleanup.

**Acceptance Scenarios**:

1. **Given** a resource CheckDestroy function, **When** resource deletion fails due to BCM API error, **Then** error message includes resource type, ID, and BCM API error details
2. **Given** multiple test resources, **When** CheckDestroy runs and one fails to delete, **Then** error message lists all failed resources with their IDs and states
3. **Given** a resource with eventual consistency, **When** CheckDestroy uses exponential backoff verification, **Then** function waits appropriately before declaring deletion failed
4. **Given** CheckDestroy function execution, **When** all resources deleted successfully, **Then** function returns nil with no spurious warnings

---

### User Story 6 - Validation Error Testing (Priority: P3)

As a Terraform provider developer, I need tests that verify schema validation catches invalid input (invalid proxy URLs, invalid network names, out-of-range values) and returns helpful error messages, so that users get clear feedback when they misconfigure resources.

**Why this priority**: Good validation improves user experience by catching errors early with clear messages. Without validation tests, we can't ensure users get helpful feedback instead of cryptic BCM API errors.

**Independent Test**: Can be fully tested by adding ExpectError test steps for each schema validator, verifying error messages match expected patterns and guide users to fix configuration.

**Acceptance Scenarios**:

1. **Given** a software image resource with invalid `software_image_proxy` URL, **When** test applies config with validation, **Then** Terraform returns error matching "invalid URL format" pattern
2. **Given** a category resource with invalid `management_network` name, **When** test applies config with validation, **Then** Terraform returns error explaining network name requirements
3. **Given** a resource with numeric field outside valid range, **When** test applies config with validation, **Then** Terraform returns error specifying valid range
4. **Given** a resource with mutually exclusive fields both set, **When** test applies config with validation, **Then** Terraform returns error explaining the conflict

---

### Edge Cases

- **Eventual consistency**: Software image cloning operations are asynchronous - tests must poll with exponential backoff and verify Read handles resources in "cloning" state gracefully
- **BCM API null handling**: Some fields return null instead of empty string - modern state checks must distinguish between null, empty, and unknown values correctly
- **ID stability**: Tests must verify resource IDs remain consistent across Create/Read/Update/Import operations and detect if ID unexpectedly changes
- **Filter edge cases**: Data source filters with no matches must return empty list (not error), filters with special characters must handle escaping correctly
- **Concurrent test execution**: Tests using shared resources (like test categories) must use unique names to avoid conflicts when run in parallel
- **Test cleanup failures**: CheckDestroy must handle cases where resources already deleted (manual cleanup, previous test failure) without failing the test run

## Requirements *(mandatory)*

### Functional Requirements

#### Modern State Verification (P1)

- **FR-001**: Resource tests MUST use `statecheck.ExpectKnownValue()` with appropriate `knownvalue` matchers (StringExact, Bool, Int64, ListExact, etc.) instead of legacy `TestCheckResourceAttr()` for type-safe verification
- **FR-002**: Resource tests MUST use `statecheck.CompareValue()` to track ID consistency across Create, Read, Update, and Import operations
- **FR-003**: Data source tests MUST use `statecheck.ExpectKnownValue()` to verify attribute values match expected types and patterns
- **FR-004**: Tests MUST use `knownvalue.NotNull()` for required fields and distinguish between null, empty string, and unknown values
- **FR-005**: List and set attributes MUST use `knownvalue.ListExact()` or `knownvalue.SetExact()` to verify element count, types, and contents

#### Idempotency Verification (P1)

- **FR-006**: Resource tests MUST include test step after Create with same config and `plancheck.ExpectEmptyPlan()` to verify idempotent creation
- **FR-007**: Resource tests MUST include test step after Update with same config and `plancheck.ExpectEmptyPlan()` to verify idempotent updates
- **FR-008**: Tests MUST fail clearly when non-idempotent behavior detected, identifying which attributes cause spurious diffs
- **FR-009**: Idempotency tests MUST account for eventual consistency by including appropriate delays or retry logic before plan check

#### Data Source Filter Verification (P2)

- **FR-010**: Data source filter tests MUST verify returned results match filter criteria using `statecheck.ExpectKnownValue()` on relevant attributes
- **FR-011**: Node filter tests MUST verify `node_type` filter returns only nodes matching specified type
- **FR-012**: Network filter tests MUST verify `name_pattern` filter returns only networks with names matching pattern
- **FR-013**: Network filter tests MUST verify `dhcp_enabled` filter returns only networks with matching DHCP configuration
- **FR-014**: Software image filter tests MUST verify category filter returns only images in specified category

#### Environment Portability (P2)

- **FR-015**: Tests MUST NOT hardcode specific counts of BCM resources (networks, nodes, images) that vary by environment
- **FR-016**: Tests MUST NOT assume specific resource names exist (like "managementnet") unless test creates them
- **FR-017**: Tests requiring specific resources MUST create test resources with unique names using `generateUniqueTestName()` helper
- **FR-018**: Tests MUST use dynamic assertions (e.g., `networks.# > 0`) instead of exact counts when verifying data source results
- **FR-019**: Filter verification tests MUST work correctly whether filter matches 0, 1, or multiple resources

#### CheckDestroy Enhancement (P3)

- **FR-020**: CheckDestroy functions MUST use `verifyResourceDeleted()` helper with exponential backoff for eventual consistency
- **FR-021**: CheckDestroy functions MUST provide detailed error messages including resource type, ID, and failure reason
- **FR-022**: CheckDestroy functions MUST iterate all resources in state and report ALL deletion failures, not just first failure
- **FR-023**: CheckDestroy functions MUST handle gracefully when resource already deleted (return success, not error)

#### Validation Testing (P3)

- **FR-024**: Resource tests MUST include test cases for each schema validator using `ExpectError` to verify validation triggers
- **FR-025**: Validation error tests MUST verify error messages match expected patterns using regex matchers
- **FR-026**: Software image tests MUST verify `software_image_proxy` validator rejects invalid URLs
- **FR-027**: Category tests MUST verify `management_network` validator rejects invalid network names (if validator exists)

### Key Entities *(include if feature involves data)*

- **Test Step Configuration**: Each test step with config, state checks, plan checks, and pre/post functions
- **State Check**: Modern verification using statecheck.ExpectKnownValue() with resource address, attribute path (tfjsonpath), and expected value matcher
- **Plan Check**: Verification using plancheck.ExpectEmptyPlan() or plancheck.ExpectNonEmptyPlan() to validate Terraform plan behavior
- **Known Value Matcher**: Type-safe value matchers (StringExact, Bool, Int64, ListExact, SetExact, NotNull, etc.) from knownvalue package
- **Compare Value Tracker**: Cross-step value tracking using statecheck.CompareValue() to verify attribute stability (especially IDs)
- **Filter Verification**: State checks that verify data source results match applied filter criteria

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Test suite quality score increases from 69% to 90%+ when measured by framework pattern adoption, verification completeness, and environment portability
- **SC-002**: All resource tests (2 resources: software_image, category) include idempotency verification with `plancheck.ExpectEmptyPlan()` after Create and Update operations
- **SC-003**: All resource tests (2 resources) use `statecheck.ExpectKnownValue()` for type-safe verification of at least 80% of schema attributes
- **SC-004**: All resource tests (2 resources) use `statecheck.CompareValue()` to track ID consistency across all CRUD operations
- **SC-005**: All data source filter tests (3 data sources with filters: nodes, networks, software_images) verify filtered results match criteria using state checks
- **SC-006**: Zero tests contain hardcoded environment-specific values (network counts, specific names, etc.)
- **SC-007**: All tests pass on BCM clusters with different configurations (varying network counts, node types, image libraries)
- **SC-008**: All CheckDestroy functions use exponential backoff verification and provide detailed error messages for failures
- **SC-009**: Test execution time remains within 10% of current baseline (no significant performance regression from additional verification)
- **SC-010**: All existing test scenarios continue passing after modernization (no regressions in coverage)

## Out of Scope *(optional)*

- Adding new test scenarios beyond modernizing existing tests (e.g., new resource attributes, new filter types)
- Performance optimization of test execution speed (beyond avoiding regressions)
- Integration with external test reporting tools or dashboards
- Automated test quality scoring tools or CI/CD pipeline changes
- Refactoring resource/data source implementation code (only test code changes)
- Adding tests for resources/data sources that don't currently have tests
- Schema validator implementation (only testing existing validators)
- Documentation updates beyond code comments in test files

## Assumptions *(optional)*

- Terraform Plugin Framework v1.16.1 and terraform-plugin-testing v1.13.3 support all modern testing patterns (statecheck, plancheck, knownvalue)
- BCM API behavior remains consistent during test execution (no external changes to test resources)
- Test helper functions (createTestBCMClient, getResourceUUIDByName, verifyResourceDeleted, generateUniqueTestName) work correctly and need no modifications
- BCM clusters used for testing have at least one network, one node, and one software image (minimal viable cluster)
- Existing test scenarios adequately cover resource/data source functionality (no need to add new scenarios)
- Test environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD, TF_ACC) are properly configured
- Resource cleanup (CheckDestroy) is necessary to avoid polluting BCM cluster with test artifacts
- Eventual consistency delays for async operations (image cloning) are known and handled by existing code

## Dependencies *(optional)*

### Internal Dependencies

- Existing test helper functions in `internal/provider/test_helpers.go`:
  - `createTestBCMClient(t)` - Creates authenticated BCM client for API calls in PreConfig/CheckDestroy
  - `getResourceUUIDByName(t, service, method, name)` - Queries BCM API for resource UUID by name
  - `verifyResourceDeleted(ctx, client, service, method, id, retries)` - Exponential backoff deletion verification
  - `generateUniqueTestName(prefix)` - Creates timestamp-based unique test resource names

- Existing test configuration pattern:
  - Provider config injection via `fmt.Sprintf()` with environment variables
  - Test resource naming convention for cleanup identification
  - PreCheck function for environment validation

- BCM API contracts:
  - JSON-RPC service/method pattern for resource queries
  - Entity structure (baseType, childType, uuid, modified, etc.) for updates
  - Snake_case (Terraform) to camelCase (BCM API) field mapping
  - Eventual consistency behavior for async operations

### External Dependencies

- terraform-plugin-testing v1.13.3:
  - `github.com/hashicorp/terraform-plugin-testing/helper/resource` - Test framework
  - `github.com/hashicorp/terraform-plugin-testing/statecheck` - Modern state verification
  - `github.com/hashicorp/terraform-plugin-testing/plancheck` - Plan verification
  - `github.com/hashicorp/terraform-plugin-testing/knownvalue` - Type-safe value matchers
  - `github.com/hashicorp/terraform-plugin-testing/tfjsonpath` - Attribute path construction

- terraform-plugin-framework v1.16.1:
  - Schema types and validators referenced in tests
  - Resource/data source interfaces tested

- BCM Cluster:
  - Running BCM instance at BCM_ENDPOINT
  - Valid credentials (BCM_USERNAME, BCM_PASSWORD)
  - Minimal viable cluster state (at least one network/node/image)
  - Write permissions for creating/modifying/deleting test resources

### Version Requirements

- Go 1.24.0+
- Terraform Plugin Framework v1.16.1
- Terraform Plugin Testing v1.13.3 (includes modern testing patterns from v1.5.0+)
- BCM API version compatible with existing client implementation

## Technical Notes *(optional)*

### Modern Testing Pattern Migration

**Import Requirements**:
```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)
```

**State Check Pattern**:
```go
// Legacy (to be replaced)
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
    resource.TestCheckResourceAttrSet("bcm_cmpart_softwareimage.test", "uuid"),
),

// Modern (target pattern)
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_cmpart_softwareimage.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact(imageName),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmpart_softwareimage.test",
        tfjsonpath.New("uuid"),
        knownvalue.NotNull(),
    ),
},
```

**ID Consistency Tracking Pattern**:
```go
var compareID = statecheck.CompareValue(compare.ValuesSame())

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

**Idempotency Check Pattern**:
```go
Steps: []resource.TestStep{
    // Create resource
    {
        Config: config("initial"),
        Check: resource.ComposeAggregateTestCheckFunc(...),
    },
    // Verify idempotency
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

**Filter Verification Pattern**:
```go
// Test filtering by node_type = "PhysicalNode"
{
    Config: testConfig_FilterByType("PhysicalNode"),
    ConfigStateChecks: []statecheck.StateCheck{
        // Verify all returned nodes have correct type
        statecheck.ExpectKnownValue(
            "data.bcm_cmdevice_nodes.test",
            tfjsonpath.New("nodes").AtSliceIndex(0).AtMapKey("node_type"),
            knownvalue.StringExact("PhysicalNode"),
        ),
        // Can add checks for additional elements if list size known
    },
}
```

### Known Value Matchers Reference

Common matchers for BCM provider attributes:
- `knownvalue.StringExact(value)` - Exact string match (name, path, notes)
- `knownvalue.Bool(value)` - Boolean match (enable_sol, dhcp_enabled)
- `knownvalue.Int64Exact(value)` - Exact int64 match (sol_speed)
- `knownvalue.NotNull()` - Any non-null value (uuid, id, creation_time)
- `knownvalue.Null()` - Explicitly null value
- `knownvalue.ListExact([]knownvalue.Check{...})` - Exact list with element matchers
- `knownvalue.ListSizeExact(n)` - List with specific element count
- `knownvalue.SetExact([]knownvalue.Check{...})` - Exact set with element matchers
- `knownvalue.StringRegexp(regexp.MustCompile("pattern"))` - Regex match for patterns

### Environment Portability Guidelines

**Anti-patterns to avoid**:
```go
// DON'T: Hardcode exact counts
resource.TestCheckResourceAttr("data.bcm_cmnet_networks.test", "networks.#", "3"),

// DON'T: Assume specific resource names exist
resource.TestCheckResourceAttr("data.bcm_cmnet_networks.filtered", "networks.0.name", "managementnet"),

// DON'T: Expect specific DHCP network counts
resource.TestCheckResourceAttr("data.bcm_cmnet_networks.dhcp", "networks.#", "2"),
```

**Portable patterns**:
```go
// DO: Check for non-empty results
resource.TestCheckResourceAttrSet("data.bcm_cmnet_networks.test", "networks.#"),
// Or use state check:
statecheck.ExpectKnownValue(
    "data.bcm_cmnet_networks.test",
    tfjsonpath.New("networks"),
    knownvalue.ListSizeExact(0).Not(), // Not empty
),

// DO: Create test resources with unique names
testNetworkName := generateUniqueTestName("test-network")
// Then verify your created resource

// DO: Verify filter logic without hardcoded counts
// Filter should work correctly regardless of result count
```

### CheckDestroy Enhancement Pattern

**Current pattern**:
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

**Enhanced pattern**:
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
                errors = append(errors, fmt.Sprintf("Resource %s (ID: %s) failed to delete: %v", rs.Type, rs.Primary.ID, err))
            } else {
                errors = append(errors, fmt.Sprintf("Resource %s (ID: %s) still exists", rs.Type, rs.Primary.ID))
            }
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failed:\n- %s", strings.Join(errors, "\n- "))
    }
    return nil
}
```

### Files Requiring Modification

1. **internal/provider/resource_cmpart_softwareimage_test.go** (12 tests, 85% → 95%+ target):
   - Add state checks to all 12 tests
   - Add ID consistency tracking across Create/Import/Update
   - Add idempotency checks after Create (Basic, FullConfig, UpdateKernelConfig, etc.)
   - Add idempotency checks after Update
   - Enhance CheckDestroy with detailed error messages

2. **internal/provider/resource_cmdevice_category_test.go** (6 tests, 80% → 95%+ target):
   - Add state checks to all 6 tests
   - Add ID consistency tracking
   - Add idempotency checks after Create and Update
   - Enhance CheckDestroy with detailed error messages
   - Add validation tests for management_network if validator exists

3. **internal/provider/data_source_cmdevice_nodes_test.go** (4 tests, 40% → 90%+ target):
   - Add state checks for attribute verification (currently only checks existence)
   - Add filter verification for FilterByType test (verify node_type matches)
   - Add filter verification for FilterByHostname test (verify hostname pattern matches)
   - Remove hardcoded assumptions about node counts/names

4. **internal/provider/data_source_cmnet_networks_test.go** (4 tests, 60% → 90%+ target):
   - Remove hardcoded `networks.# = "3"` assertion
   - Remove hardcoded `networks.0.name = "managementnet"` assertion
   - Remove hardcoded `networks.# = "2"` for DHCP filter
   - Add filter verification for NameFilter test
   - Add filter verification for DHCPFilter test
   - Use dynamic assertions that work on any cluster

5. **internal/provider/data_source_cmpart_softwareimages_test.go** (5 tests, 50% → 90%+ target):
   - Add state checks for attribute verification
   - Add filtering tests with filter verification
   - Remove any duplicate test scenarios
   - Ensure filter tests verify results match criteria

6. **internal/provider/data_source_cmdevice_categories_test.go** (4 tests, 75% → 90%+ target):
   - Add state checks for better attribute verification
   - Add filter verification if filter tests exist
   - Ensure all attributes properly verified

### Test Quality Scoring Methodology

**Quality Score Calculation** (per test file):
- **Modern Patterns (40%)**: Percentage of tests using statecheck/plancheck/knownvalue
- **Verification Completeness (30%)**: Percentage of schema attributes verified with state checks
- **Environment Portability (20%)**: No hardcoded counts/names, tests pass on different clusters
- **Best Practices (10%)**: ID tracking, enhanced CheckDestroy, validation tests, good naming

**Target Distribution**:
- Resource tests: 95%+ (complete CRUD + idempotency + ID tracking)
- Data source tests: 90%+ (state checks + filter verification + portability)
- Overall suite: 90%+ average

### Backward Compatibility

- Keep legacy `TestCheckResourceAttr()` assertions initially while adding modern state checks
- Can run both legacy and modern checks side-by-side during migration
- Once modern checks proven stable, remove redundant legacy checks
- All existing test scenarios must continue passing (no regression in coverage)
