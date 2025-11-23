# Feature Specification: Modernize Legacy Testing Patterns (Issue #40)

**Feature Branch**: `040-modernize-legacy-tests`
**GitHub Issue**: #40
**Created**: 2025-11-23
**Status**: In Progress

## Summary

Modernize 3 test files using legacy `terraform-plugin-sdk` testing patterns to modern `terraform-plugin-testing` v1.13.3+ patterns with type-safe state checks, plan verification, and ID tracking. This achieves 100% modern pattern adoption across the test suite.

## Target Files

| Priority | File | Current State | Legacy Checks | Target |
|----------|------|---------------|---------------|--------|
| **P1** | `internal/provider/resource_cmnet_network_test.go` | 5% modern | 18 legacy | 100% modern |
| **P2** | `internal/provider/resource_cmkube_cluster_test.go` | 53% modern | 38 legacy | 100% modern |
| **P3** | `internal/provider/data_source_cmpart_partitions_test.go` | 73% modern | 10 legacy | 100% modern |

**Total**: 66 legacy assertions to modernize

## User Scenarios & Testing

### User Story 1 - Type-Safe Resource State Verification (P1)

**As a** Terraform provider developer
**I want** resource tests using type-safe state verification with `statecheck.ExpectKnownValue()`
**So that** I catch type mismatches at test time instead of runtime

**Acceptance Scenarios**:

1. **Given** `resource_cmnet_network_test.go` with string-based MTU check `"9000"`, **When** converted to `knownvalue.Int64Exact(9000)`, **Then** test validates integer type correctly
2. **Given** network resource test with boolean `dhcp_enabled` as string `"true"`, **When** converted to `knownvalue.Bool(true)`, **Then** test catches if API returns string instead of boolean
3. **Given** network resource test with computed `uuid` field using `TestCheckResourceAttrSet`, **When** converted to `knownvalue.NotNull()`, **Then** test verifies presence with better type safety

### User Story 2 - Idempotency Verification for Resources (P1)

**As a** Terraform provider developer
**I want** idempotency checks with `plancheck.ExpectEmptyPlan()` after Create and Update
**So that** I ensure resources don't generate spurious diffs

**Acceptance Scenarios**:

1. **Given** `resource_cmnet_network_test.go` Create step, **When** followed by step with same config and `ExpectEmptyPlan()`, **Then** Terraform shows no changes
2. **Given** `resource_cmkube_cluster_test.go` Update step, **When** followed by idempotency check, **Then** Terraform shows no changes
3. **Given** resource with computed fields, **When** idempotency test runs, **Then** test fails if fields cause drift

### User Story 3 - ID Consistency Tracking (P2)

**As a** Terraform provider developer
**I want** ID consistency tracking with `statecheck.CompareValue()` across CRUD operations
**So that** I detect if resource IDs unexpectedly change

**Acceptance Scenarios**:

1. **Given** `resource_cmnet_network_test.go` with Create/Import/Update steps, **When** ID tracking added, **Then** test verifies ID remains same across all operations
2. **Given** `resource_cmkube_cluster_test.go` resource operations, **When** ID changes unexpectedly, **Then** test fails with clear message showing ID mismatch

### User Story 4 - Data Source Type-Safe Verification (P3)

**As a** Terraform provider developer
**I want** data source tests using type-safe matchers for attributes
**So that** I verify data types match schema definitions

**Acceptance Scenarios**:

1. **Given** `data_source_cmpart_partitions_test.go` checking partition attributes, **When** converted to use `knownvalue.StringExact()` for names, **Then** test validates string types correctly
2. **Given** partitions data source with list attributes, **When** converted to use `knownvalue.ListSizeExact()`, **Then** test verifies list counts type-safely

## Requirements

### Functional Requirements

#### Modern State Verification (P1)

- **FR-001**: Replace all `resource.TestCheckResourceAttr()` calls with `statecheck.ExpectKnownValue()` + type-safe matchers
- **FR-002**: Use `knownvalue.StringExact()` for string attributes (name, path, notes, etc.)
- **FR-003**: Use `knownvalue.Int64Exact()` for integer attributes (mtu, port, sol_speed)
- **FR-004**: Use `knownvalue.Bool()` for boolean attributes (dhcp_enabled, enable_sol)
- **FR-005**: Replace `resource.TestCheckResourceAttrSet()` with `knownvalue.NotNull()`
- **FR-006**: Add required imports: `statecheck`, `knownvalue`, `tfjsonpath`, `plancheck`, `compare`

#### Idempotency Verification (P1)

- **FR-007**: Add test step after Create with same config and `plancheck.ExpectEmptyPlan()`
- **FR-008**: Add test step after Update with same config and `plancheck.ExpectEmptyPlan()`
- **FR-009**: Ensure all resource tests verify idempotent behavior

#### ID Consistency Tracking (P2)

- **FR-010**: Initialize `compareID := statecheck.CompareValue(compare.ValuesSame())` at test start
- **FR-011**: Add `compareID.AddStateValue()` to Create, Import, and Update steps
- **FR-012**: Verify ID remains consistent across all CRUD operations

#### Code Quality (P3)

- **FR-013**: Remove all legacy assertion patterns (zero remaining `TestCheckResourceAttr` calls)
- **FR-014**: Ensure all tests pass: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/`
- **FR-015**: No test regressions (all existing scenarios continue passing)

## Success Criteria

- **SC-001**: Zero remaining `resource.TestCheckResourceAttr()` calls in 3 target files (66 conversions)
- **SC-002**: Zero remaining `resource.TestCheckResourceAttrSet()` calls in 3 target files
- **SC-003**: All numeric attributes use `Int64Exact()` not string "9000"
- **SC-004**: All boolean attributes use `Bool(true/false)` not string "true"/"false"
- **SC-005**: All resource tests include idempotency verification after Create and Update
- **SC-006**: All resource tests include ID consistency tracking across operations
- **SC-007**: 100% modern pattern adoption in all 3 files
- **SC-008**: All acceptance tests pass without regressions
- **SC-009**: Test execution time within 10% of baseline (no performance degradation)

## Out of Scope

- Adding new test scenarios beyond modernizing existing tests
- Modifying resource/data source implementation code
- Refactoring test helper functions
- Performance optimization beyond avoiding regressions
- Adding tests for resources without current test coverage
- Schema validator implementation
- Documentation updates beyond code comments

## Assumptions

- terraform-plugin-testing v1.13.3 supports all modern patterns
- BCM API behavior remains stable during test execution
- Test helper functions work correctly and need no modifications
- BCM clusters have minimal viable configuration
- Test environment variables properly configured
- Existing test coverage adequate

## Dependencies

### Internal Dependencies

- Test helpers in `internal/provider/test_helpers.go`:
  - `createTestBCMClient(t)` - Authenticated BCM client
  - `getResourceUUIDByName(t, service, method, name)` - UUID lookup
  - `verifyResourceDeleted(ctx, client, service, method, id, retries)` - Deletion verification
  - `generateUniqueTestName(prefix)` - Unique test names

### External Dependencies

- terraform-plugin-testing v1.13.3:
  - `github.com/hashicorp/terraform-plugin-testing/helper/resource`
  - `github.com/hashicorp/terraform-plugin-testing/statecheck`
  - `github.com/hashicorp/terraform-plugin-testing/plancheck`
  - `github.com/hashicorp/terraform-plugin-testing/knownvalue`
  - `github.com/hashicorp/terraform-plugin-testing/tfjsonpath`
  - `github.com/hashicorp/terraform-plugin-testing/compare`

- BCM Cluster:
  - Running instance at BCM_ENDPOINT
  - Valid credentials
  - Write permissions for test resources

### Version Requirements

- Go 1.24.0+
- Terraform Plugin Framework v1.16.1
- Terraform Plugin Testing v1.13.3
- BCM API compatible with current client

## Technical Notes

### Migration Pattern for resource_cmnet_network_test.go (P1)

**Current Legacy Pattern**:
```go
Check: resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000"),
    resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_enabled", "true"),
    resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid"),
),
```

**Modern Pattern**:
```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact(networkName),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("mtu"),
        knownvalue.Int64Exact(9000),  // Type-safe integer
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("dhcp_enabled"),
        knownvalue.Bool(true),  // Type-safe boolean
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmnet_network.test",
        tfjsonpath.New("uuid"),
        knownvalue.NotNull(),  // Type-safe existence check
    ),
},
```

### Idempotency Pattern

```go
Steps: []resource.TestStep{
    // Create resource
    {
        Config: testAccCMNetNetworkConfig(networkName, 9000, true),
        ConfigStateChecks: []statecheck.StateCheck{
            // ... state checks ...
        },
    },
    // Verify idempotency after Create
    {
        Config: testAccCMNetNetworkConfig(networkName, 9000, true),
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectEmptyPlan(),
            },
        },
    },
}
```

### ID Consistency Tracking Pattern

```go
func TestAccCMNetNetwork_Complete(t *testing.T) {
    networkName := generateUniqueTestName("test-network")

    // Initialize ID tracker
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Create
            {
                Config: testConfig(networkName),
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id")),
                },
            },
            // Import
            {
                ResourceName:      "bcm_cmnet_network.test",
                ImportState:       true,
                ImportStateVerify: true,
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id")),
                },
            },
            // Update
            {
                Config: testConfig(networkName),
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue("bcm_cmnet_network.test", tfjsonpath.New("id")),
                },
            },
        },
    })
}
```

### Required Imports

Add to each modernized file:

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

### Type Matcher Reference

| BCM Attribute Type | Terraform Type | knownvalue Matcher | Example |
|-------------------|----------------|-------------------|---------|
| name, path, notes | types.String | StringExact("value") | Network name |
| mtu, port, sol_speed | types.Int64 | Int64Exact(9000) | MTU size |
| dhcp_enabled, enable_sol | types.Bool | Bool(true) | DHCP enabled |
| uuid, id | types.String (computed) | NotNull() | Resource UUID |
| networks, modules | types.List | ListSizeExact(n) | List count |

## Testing Strategy

### Test Execution Plan

1. **Phase 1 - Priority 1**: Modernize `resource_cmnet_network_test.go`
   - Convert 18 legacy assertions
   - Add idempotency checks
   - Add ID tracking
   - Run: `TF_ACC=1 go test -v -run TestAccCMNetNetwork ./internal/provider/`

2. **Phase 2 - Priority 2**: Modernize `resource_cmkube_cluster_test.go`
   - Convert 38 legacy assertions
   - Add idempotency checks
   - Add ID tracking
   - Run: `TF_ACC=1 go test -v -run TestAccCMKubeCluster ./internal/provider/`

3. **Phase 3 - Priority 3**: Modernize `data_source_cmpart_partitions_test.go`
   - Convert 10 legacy assertions
   - Add type-safe checks
   - Run: `TF_ACC=1 go test -v -run TestAccCMPartPartitions ./internal/provider/`

4. **Phase 4 - Full Validation**: Run all tests
   - Execute: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/`
   - Verify no regressions
   - Verify 100% pass rate

### Success Validation

```bash
# Run all affected tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMNetNetwork|TestAccCMKubeCluster|TestAccCMPartPartitions"

# Expected output: PASS for all tests
# Expected: Zero legacy patterns remaining
```

## Benefits

1. **Type Safety**: Catch type mismatches at test time instead of runtime
2. **Better Error Messages**: Clear indication of expected vs actual types/values
3. **Idempotency Verification**: Built-in checks prevent spurious diffs
4. **ID Consistency**: Track resource IDs across all operations
5. **HashiCorp Compliance**: Aligns with official testing best practices
6. **Maintainability**: Easier to understand and modify for future developers
7. **Quality Improvement**: Moves from 81% to 100% modern pattern adoption

## References

- GitHub Issue: #40
- Migration Guide: `/workspace/docs/TESTING_MODERNIZATION.md` (if exists)
- Project Standards: `/workspace/CLAUDE.md` (Modern Testing Patterns section)
- HashiCorp Docs: https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns
- terraform-plugin-testing: https://github.com/hashicorp/terraform-plugin-testing
