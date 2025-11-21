# Feature Specification: Comprehensive Test Review - Drift Detection and Destroy Testing

**Feature Branch**: `006-test-review`
**Created**: 2025-11-21
**Status**: Draft
**Input**: User description: "I want to review all my go tests for this terraform provider, key areas I want to focus on is drift and destroy testing"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Detect External Resource Modifications (Priority: P1)

As a Terraform user, when someone manually modifies a BCM resource outside of Terraform (e.g., via BCM UI or API), I need Terraform to detect this drift on the next `terraform plan` or `terraform refresh` so that I can reconcile the actual state with my desired configuration.

**Why this priority**: Drift detection is fundamental to Terraform's declarative model. Without reliable drift detection, users cannot trust that their Terraform state reflects reality, leading to configuration inconsistencies and potential infrastructure failures.

**Independent Test**: Can be fully tested by creating a resource via Terraform, manually modifying it via BCM API, running `terraform plan`, and verifying that Terraform detects the changes.

**Acceptance Scenarios**:

1. **Given** a software image exists in Terraform state with `kernel_parameters = "quiet splash"`, **When** the kernel parameters are changed to "quiet splash nomodeset" via BCM API, **Then** `terraform plan` shows a diff indicating the kernel_parameters changed from "quiet splash nomodeset" to "quiet splash"
2. **Given** a category exists in Terraform state with `notes = "Production"`, **When** the notes field is updated to "Staging" via BCM API, **Then** `terraform plan` detects the drift and proposes to restore notes to "Production"
3. **Given** a software image exists with 2 kernel modules configured, **When** a module is removed via BCM API, **Then** `terraform refresh` updates state to reflect 1 module and subsequent plan shows drift
4. **Given** a resource exists in Terraform state, **When** it is deleted entirely via BCM API, **Then** `terraform plan` detects the resource is missing and proposes to recreate it

---

### User Story 2 - Verify Complete Resource Cleanup (Priority: P1)

As a Terraform user, when I run `terraform destroy` or delete a resource from my configuration, I need confidence that all resources are completely removed from the BCM cluster with no orphaned resources remaining.

**Why this priority**: Incomplete cleanup leads to resource leaks, unexpected costs, and cluttered infrastructure. This is especially critical in test environments where resources are frequently created and destroyed.

**Independent Test**: Can be fully tested by creating resources via Terraform, destroying them, and verifying via BCM API that the resources no longer exist.

**Acceptance Scenarios**:

1. **Given** a software image resource exists in Terraform state, **When** `terraform destroy` is executed, **Then** the software image is completely removed from BCM and CheckDestroy verification passes
2. **Given** a category resource with associated nodes, **When** `terraform destroy` is executed with `force = true`, **Then** the category is removed despite associations and no orphaned data remains
3. **Given** multiple resources exist (software image + category), **When** `terraform destroy` is executed, **Then** resources are destroyed in correct dependency order and all resources are removed
4. **Given** a resource destroy operation fails mid-execution, **When** `terraform destroy` is retried, **Then** the operation is idempotent and completes successfully without errors

---

### User Story 3 - Handle Destroy Edge Cases (Priority: P2)

As a Terraform user, when resource deletion encounters edge cases (concurrent modifications, dependencies, timeouts), I need clear error messages and consistent behavior so that I can understand and resolve deletion issues.

**Why this priority**: Real-world deletion operations frequently encounter edge cases. Proper handling ensures users can safely clean up resources even in complex scenarios.

**Independent Test**: Can be fully tested by simulating edge case conditions (locked resources, concurrent operations) and verifying error handling and recovery.

**Acceptance Scenarios**:

1. **Given** a software image is being cloned (file operation in progress), **When** `terraform destroy` is executed, **Then** the destroy operation waits for the clone to complete or times out with a clear error message
2. **Given** a category has associated nodes, **When** `terraform destroy` is executed with `force = false`, **Then** a clear error indicates the category cannot be deleted due to node associations
3. **Given** a destroy operation is interrupted by network failure, **When** the operation is retried, **Then** the provider correctly handles partial deletion and completes the operation
4. **Given** a resource is already deleted (manual deletion after Terraform create), **When** `terraform destroy` is executed, **Then** the operation succeeds idempotently without error

---

### User Story 4 - Validate Drift Detection for All Attributes (Priority: P2)

As a Terraform user, I need drift detection to work for ALL resource attributes (not just a subset) so that my state accurately reflects all configuration aspects of my infrastructure.

**Why this priority**: Partial drift detection creates false confidence and hidden inconsistencies. Complete attribute coverage ensures comprehensive state accuracy.

**Independent Test**: Can be fully tested by systematically modifying each attribute type (strings, bools, lists, objects) via BCM API and verifying detection.

**Acceptance Scenarios**:

1. **Given** a resource with boolean attribute `enable_sol = false`, **When** changed to `true` via BCM API, **Then** drift is detected
2. **Given** a resource with list attribute `modules = [...]`, **When** modules are added/removed via BCM API, **Then** drift is detected with correct list comparison
3. **Given** a resource with nested object `software_image_proxy`, **When** nested fields are modified via BCM API, **Then** drift is detected at the nested field level
4. **Given** a resource with computed attribute `creation_time`, **When** the value changes, **Then** drift is NOT reported (computed values are expected to change)

---

### User Story 5 - Test Async Operation Drift Detection (Priority: P3)

As a Terraform user, when BCM performs asynchronous operations (e.g., image cloning) that modify resource state over time, I need drift detection to account for eventual consistency so that I don't see false positives during legitimate operations.

**Why this priority**: BCM operations like cloning can take time to complete, and certain fields reset after async operations. Tests should verify drift detection handles this correctly.

**Independent Test**: Can be tested by initiating an async operation and verifying drift detection behavior during and after operation completion.

**Acceptance Scenarios**:

1. **Given** a software image is being cloned with `original_image` set, **When** the clone completes and BCM resets `original_image` to all zeros, **Then** Terraform does NOT report drift on original_image (it's marked as ImportStateVerifyIgnore)
2. **Given** a resource with `file_operation_in_progress = true`, **When** refresh is called, **Then** Terraform correctly reads the transient state without false drift detection
3. **Given** an async operation modifies both config and computed fields, **When** `terraform plan` is run, **Then** only config field changes are reported as drift

---

### Edge Cases

- **Concurrent modifications**: What happens when Terraform and BCM API modify the same resource simultaneously?
- **Network timeouts during Read**: How does drift detection handle transient API failures during state refresh?
- **Malformed API responses**: How do tests handle unexpected response formats from BCM during drift checks?
- **Large list/object changes**: How efficiently does drift detection handle resources with 100+ modules or large nested objects?
- **Resource state transitions**: How is drift handled when a resource is in a transitional state (e.g., being provisioned)?
- **Partial failures**: What happens when CheckDestroy passes but resources still exist in BCM?
- **Orphaned dependencies**: How do destroy tests verify cleanup of implicitly created dependencies (e.g., filesystem partitions)?

## Requirements *(mandatory)*

### Functional Requirements

#### Drift Detection Requirements

- **FR-001**: Acceptance tests MUST verify that Read operation detects changes to ALL non-computed resource attributes when modified externally via BCM API
- **FR-002**: Acceptance tests MUST verify that Read operation correctly handles resource deletion (returns error or empty state) when resource is removed externally
- **FR-003**: Acceptance tests MUST verify that Read operation does NOT report drift for computed-only attributes
- **FR-004**: Tests MUST use BCM client directly to simulate external modifications between Terraform operations
- **FR-005**: Drift detection tests MUST cover all attribute types: strings, bools, ints, lists, nested objects, and dynamic types
- **FR-006**: Tests MUST verify drift detection for complex nested structures (e.g., `modules`, `software_image_proxy`, `bmc_settings`)
- **FR-007**: Tests MUST verify that Read operation handles eventual consistency for async operations (e.g., image cloning)
- **FR-008**: Tests MUST verify that ImportStateVerifyIgnore correctly excludes expected drift fields (e.g., `original_image`, `force`)

#### Destroy Testing Requirements

- **FR-009**: All resource tests MUST include a CheckDestroy function that verifies complete resource removal
- **FR-010**: CheckDestroy MUST verify resource deletion by attempting to read the resource via BCM API and confirming error or empty response
- **FR-011**: CheckDestroy MUST iterate over ALL resources of the type in Terraform state, not just a single resource
- **FR-012**: Tests MUST verify destroy operations handle dependencies correctly (e.g., can't delete category with associated nodes without force=true)
- **FR-013**: Tests MUST verify destroy operations are idempotent (running destroy twice succeeds without error)
- **FR-014**: Tests MUST verify destroy with special flags (e.g., `force = true` for categories)
- **FR-015**: Tests MUST include PreCheck cleanup to ensure clean test environment (delete leftover resources from previous runs)
- **FR-016**: PreCheck cleanup MUST verify deletion completion using exponential backoff retry logic for async operations

#### Test Quality Requirements

- **FR-017**: All acceptance tests MUST use unique resource names with timestamp suffixes to avoid collisions
- **FR-018**: Test configurations MUST be functions (not constants) to allow parameterization
- **FR-019**: All test configurations MUST include provider configuration with credentials from environment variables
- **FR-020**: Tests MUST use resource.ComposeAggregateTestCheckFunc to continue checking even if individual checks fail
- **FR-021**: Tests MUST verify both attribute existence (TestCheckResourceAttrSet) and specific values (TestCheckResourceAttr)
- **FR-022**: Negative tests MUST use ExpectError with appropriate regex patterns

#### Coverage Requirements

- **FR-023**: EVERY resource type MUST have drift detection tests covering attribute modifications
- **FR-024**: EVERY resource type MUST have CheckDestroy implementation and verification
- **FR-025**: EVERY resource type MUST have PreCheck cleanup implementation
- **FR-026**: Tests MUST cover drift detection for at least 80% of non-computed attributes per resource
- **FR-027**: Tests MUST cover destroy scenarios: basic deletion, force deletion (if applicable), deletion with dependencies
- **FR-028**: Data sources MUST have tests but CheckDestroy is NOT applicable (read-only operations)

### Key Entities *(test infrastructure)*

- **Test Resources**: bcm_cmpart_softwareimage, bcm_cmdevice_category
- **BCM API Client**: Used in tests to simulate external modifications and verify cleanup
- **Test Configuration Functions**: Parameterized HCL config generators for test scenarios
- **CheckDestroy Functions**: Per-resource verification functions that confirm complete deletion
- **PreCheck Functions**: Per-resource cleanup functions that prepare clean test environment
- **Test State**: Terraform state used to iterate resources during destroy verification

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: 100% of resource types have CheckDestroy implementations that verify complete resource removal via BCM API
- **SC-002**: 100% of resource types have PreCheck cleanup that removes leftover resources with verified deletion
- **SC-003**: All acceptance tests include at least one test step that simulates external modification and verifies drift detection
- **SC-004**: CheckDestroy functions verify deletion for ALL resources of a type in state, not just test resources
- **SC-005**: All destroy tests pass consistently without manual cleanup required between runs
- **SC-006**: Drift detection tests cover at least 80% of non-computed attributes for each resource type
- **SC-007**: All tests use unique resource names to support parallel test execution
- **SC-008**: PreCheck cleanup operations complete within 30 seconds per resource using exponential backoff
- **SC-009**: Test coverage includes at least 3 drift scenarios per resource (string attr, list attr, object attr)
- **SC-010**: Test coverage includes at least 3 destroy scenarios per resource (basic, with dependencies, idempotent retry)

## Out of Scope *(important clarifications)*

### Explicitly NOT Included

- **Performance testing**: Load testing or stress testing of Read/Destroy operations (focus is correctness, not performance)
- **Unit tests**: This spec covers acceptance tests only; unit tests are separate
- **Integration tests**: Testing Terraform operations against mocked BCM API (acceptance tests use real API)
- **Terraform state corruption**: Testing recovery from corrupted state files
- **Multi-provider scenarios**: Testing drift/destroy with resources from multiple providers
- **Terraform Cloud/Enterprise**: Testing remote backend or Terraform Cloud-specific behavior
- **Provider upgrade testing**: Testing state migration or compatibility between provider versions
- **Concurrent Terraform runs**: Testing race conditions when multiple Terraform processes run simultaneously
- **Network security**: Testing TLS, certificate validation, or authentication failures (covered by provider tests)
- **BCM API bugs**: Testing workarounds for BCM API issues (out of provider control)

### Deferred to Future Work

- **Automated test generation**: Tools to automatically generate drift tests from resource schema
- **Test coverage dashboard**: Visualization of drift/destroy test coverage per attribute
- **Chaos testing**: Randomly injecting failures during test execution
- **Property-based testing**: Using frameworks like gopter for exhaustive scenario testing

## Assumptions *(important context)*

### Technical Assumptions

- **BCM API availability**: Tests assume BCM cluster at 172.21.15.254:8081 is running and accessible
- **BCM API credentials**: Tests assume BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD environment variables are set
- **Test isolation**: Tests assume unique resource names prevent conflicts between parallel test runs
- **API stability**: Tests assume BCM API behavior is consistent and deterministic
- **Eventual consistency window**: Tests assume BCM operations complete within 30 seconds (exponential backoff with 5 retries)

### Test Environment Assumptions

- **Clean initial state**: PreCheck cleanup ensures tests start with clean environment
- **Resource uniqueness**: Generated timestamp-based names are sufficiently unique to avoid collisions
- **Terraform version**: Tests assume Terraform 1.5.0+ with protocol v6 support
- **Go version**: Tests assume Go 1.24+ with testing framework support
- **TF_ACC flag**: Tests only run when TF_ACC=1 is set (standard acceptance test convention)

### BCM-Specific Assumptions

- **Cookie-based auth**: BCM API uses cm-login-token cookie for authentication (already implemented in provider)
- **JSON-RPC protocol**: BCM API uses JSON-RPC format with service/call/args structure
- **Resource identification**: Resources are identified by UUID for API operations, name for user operations
- **Async operations**: Image cloning is async; other operations are synchronous
- **Field resets**: BCM resets certain fields (e.g., original_image) after async operations complete
- **Force deletion**: Categories support force deletion flag to override dependency checks

## Dependencies *(external requirements)*

### Required Components

- **BCM Test Cluster**: Running BCM instance at 172.21.15.254:8081 with administrative credentials
- **Test Credentials**: Environment variables BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD set to valid values
- **Terraform Plugin Framework**: v1.16.1 (already in go.mod)
- **Terraform Plugin Testing**: v1.13.3 (already in go.mod)
- **BCM Client Implementation**: internal/provider/bcm_client.go (already exists)
- **Existing Resources**: At least one software image (default-image) and network for test data sources

### Test Infrastructure Dependencies

- **Unique name generation**: generateUniqueTestName() function (already exists in resource_cmpart_softwareimage_test.go)
- **Provider factories**: testAccProtoV6ProviderFactories (already exists in provider_test.go)
- **PreCheck function**: testAccPreCheck() (already exists in provider_test.go)

## Test Implementation Patterns *(guidance)*

### Drift Detection Test Pattern

```go
// Example: TestAccResourceName_DriftDetection
func TestAccResourceName_DriftDetection(t *testing.T) {
    resourceName := generateUniqueTestName("test-drift")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccResourcePreCheck(t, resourceName) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckResourceDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create resource
            {
                Config: testAccResourceConfig_Basic(resourceName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_resource.test", "name", resourceName),
                    resource.TestCheckResourceAttr("bcm_resource.test", "attribute", "original-value"),
                ),
            },
            // Step 2: Manually modify via BCM API (simulating external change)
            {
                PreConfig: func() {
                    // Use BCM client to modify resource
                    client := createTestBCMClient(t)
                    err := client.CallJSONRPC(ctx, "service", "updateResource", uuid, newValue)
                    if err != nil {
                        t.Fatalf("Failed to simulate external modification: %v", err)
                    }
                },
                Config: testAccResourceConfig_Basic(resourceName), // Same config
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify drift detected by checking attribute reflects actual state
                    resource.TestCheckResourceAttr("bcm_resource.test", "attribute", "modified-value"),
                ),
                ExpectNonEmptyPlan: true, // Plan should show diff to restore original value
            },
            // Step 3: Apply to restore desired state
            {
                Config: testAccResourceConfig_Basic(resourceName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_resource.test", "attribute", "original-value"),
                ),
            },
        },
    })
}
```

### CheckDestroy Pattern

```go
// Example: testAccCheckResourceDestroy
func testAccCheckResourceDestroy(s *terraform.State) error {
    // Create BCM client
    endpoint := os.Getenv("BCM_ENDPOINT")
    username := os.Getenv("BCM_USERNAME")
    password := os.Getenv("BCM_PASSWORD")

    client, err := NewBCMClient(context.Background(), endpoint, username, password, true, 30)
    if err != nil {
        return fmt.Errorf("failed to create BCM client: %w", err)
    }

    // Check ALL resources in state
    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_resource_type" {
            continue
        }

        // Attempt to read resource
        name := rs.Primary.Attributes["name"]
        body, err := client.CallJSONRPC(context.Background(), "service", "getResource", name)

        // Resource should NOT exist
        if err == nil {
            var resourceData map[string]interface{}
            if json.Unmarshal(body, &resourceData) == nil && len(resourceData) > 0 {
                return fmt.Errorf("resource %s still exists after destroy", name)
            }
        }
        // If error or empty response, resource is deleted (expected)
    }

    return nil
}
```

### PreCheck Cleanup Pattern

```go
// Example: testAccResourcePreCheck
func testAccResourcePreCheck(t *testing.T, names ...string) {
    testAccPreCheck(t) // Call base precheck

    // Create BCM client
    endpoint := os.Getenv("BCM_ENDPOINT")
    username := os.Getenv("BCM_USERNAME")
    password := os.Getenv("BCM_PASSWORD")

    client, err := NewBCMClient(context.Background(), endpoint, username, password, true, 30)
    if err != nil {
        t.Logf("Failed to create BCM client for cleanup: %v", err)
        return
    }

    // Clean up leftover resources with retry logic
    for _, name := range names {
        body, err := client.CallJSONRPC(context.Background(), "service", "getResource", name)
        if err == nil {
            var resourceData map[string]interface{}
            if json.Unmarshal(body, &resourceData) == nil {
                if uuid, ok := resourceData["uuid"].(string); ok && uuid != "" {
                    // Resource exists, delete it
                    _, err := client.CallJSONRPC(context.Background(), "service", "removeResource", uuid, forceFlag)
                    if err != nil {
                        t.Logf("Failed to delete leftover resource %s: %v", name, err)
                        continue
                    }

                    // Verify deletion with exponential backoff
                    maxRetries := 5
                    waitTime := 1 * time.Second
                    deleted := false

                    for retry := 0; retry < maxRetries; retry++ {
                        time.Sleep(waitTime)

                        // Check if resource is gone
                        body, err := client.CallJSONRPC(context.Background(), "service", "getResource", name)
                        if err != nil || len(body) == 0 {
                            deleted = true
                            t.Logf("✓ Cleaned up leftover resource: %s", name)
                            break
                        }

                        var checkData map[string]interface{}
                        if json.Unmarshal(body, &checkData) == nil && len(checkData) == 0 {
                            deleted = true
                            t.Logf("✓ Cleaned up leftover resource: %s", name)
                            break
                        }

                        waitTime *= 2 // Exponential backoff
                    }

                    if !deleted {
                        t.Logf("⚠ Warning: Resource %s may not be fully deleted after %d retries", name, maxRetries)
                    }
                }
            }
        }
    }
}
```

## Risk Analysis *(important considerations)*

### High-Impact Risks

1. **Incomplete CheckDestroy verification**
   - **Risk**: CheckDestroy returns success but resources still exist in BCM
   - **Impact**: Resource leaks, test environment pollution, false confidence
   - **Mitigation**: Always verify via BCM API read, check for both error AND empty response

2. **Race conditions in async operations**
   - **Risk**: Drift tests check state before async operation completes
   - **Impact**: Flaky tests, false positives, intermittent failures
   - **Mitigation**: Use exponential backoff polling, verify operation completion before assertions

3. **Insufficient attribute coverage**
   - **Risk**: Drift detection tests only cover subset of attributes
   - **Impact**: Undetected drift in production, configuration inconsistencies
   - **Mitigation**: Systematic test generation for all non-computed attributes, coverage tracking

### Medium-Impact Risks

4. **Test environment cleanup failures**
   - **Risk**: PreCheck cleanup fails to remove leftover resources
   - **Impact**: Test failures due to name conflicts, unpredictable test behavior
   - **Mitigation**: Robust error handling, retry logic, fallback to manual cleanup instructions

5. **BCM API inconsistencies**
   - **Risk**: BCM API returns inconsistent data for same resource
   - **Impact**: False drift detection, test instability
   - **Mitigation**: Test against known-good BCM version, document API quirks

6. **Network timeouts during tests**
   - **Risk**: BCM API calls timeout during long-running operations
   - **Impact**: Test failures, incomplete verification
   - **Mitigation**: Configurable timeouts, retry logic, clear timeout error messages

### Low-Impact Risks

7. **Test name collisions**
   - **Risk**: Multiple test runs use same resource names
   - **Impact**: Test failures, incorrect results
   - **Mitigation**: Timestamp-based unique names, prefix with test function name

8. **Missing ImportStateVerifyIgnore**
   - **Risk**: Tests fail on expected computed/transient field differences
   - **Impact**: False negatives, time wasted investigating non-issues
   - **Mitigation**: Document all expected exceptions, centralize ignore lists

## Current Test Gaps (Analysis)

### Existing Coverage (Good)

✅ **bcm_cmpart_softwareimage**:
- Has CheckDestroy implementation with BCM API verification
- Has PreCheck cleanup with exponential backoff retry logic
- Has unique name generation
- Tests cover basic CRUD, ImportState, multiple update scenarios

✅ **bcm_cmdevice_category**:
- Has CheckDestroy implementation with BCM API verification
- Has PreCheck cleanup with deletion verification
- Tests cover basic CRUD, ImportState, force parameter

✅ **Data sources**:
- Tests cover basic read operations, empty responses, invalid credentials
- Correctly do NOT include CheckDestroy (read-only)

### Missing Coverage (Gaps to Address)

❌ **Drift detection tests**: NO drift detection tests exist for any resource
- No tests simulate external modifications via BCM API
- No tests verify Read correctly updates state after external changes
- No tests verify ExpectNonEmptyPlan after drift

❌ **Comprehensive attribute coverage in drift tests**:
- Need systematic testing of ALL attribute types per resource
- Need tests for nested objects (software_image_proxy, bmc_settings, modules)
- Need tests for list attributes (modules, name_servers, search_domains)

❌ **Resource deletion verification**: CheckDestroy exists but could be enhanced
- Could verify no orphaned dependencies remain (e.g., filesystem partitions)
- Could verify force deletion scenarios more thoroughly
- Could test retry after partial deletion

❌ **Destroy edge cases**: Limited coverage
- No tests for destroying resources with file operations in progress
- No tests for destroying resources with complex dependencies
- No tests for idempotent destroy (running twice)

❌ **Data source drift**: Not applicable but documentation needed
- Should document that data sources don't support drift detection
- Should verify data source reads always fetch fresh data

## Implementation Phases *(recommended approach)*

### Phase 1: Drift Detection Infrastructure (P1 - Week 1)

**Goal**: Add drift detection test framework and patterns

**Deliverables**:
1. Create helper function `createTestBCMClient(t)` for tests to use BCM client
2. Document drift detection test pattern with PreConfig example
3. Add drift detection test for bcm_cmpart_softwareimage (string attribute)
4. Add drift detection test for bcm_cmdevice_category (string attribute)
5. Verify ExpectNonEmptyPlan works correctly

**Validation**: 2 passing drift detection tests (one per resource)

### Phase 2: Comprehensive Drift Coverage (P1 - Week 1)

**Goal**: Achieve 80% attribute coverage for drift detection

**Deliverables**:
1. Add drift tests for all bcm_cmpart_softwareimage attributes:
   - String attributes: kernel_parameters, kernel_output_console, notes
   - Bool attributes: enable_sol, sol_flow_control
   - List attributes: modules (add/remove/modify)
2. Add drift tests for all bcm_cmdevice_category attributes:
   - String attributes: notes, kernel_parameters
   - Bool attributes: install_boot_record, allow_networking_restart
   - Nested object: software_image_proxy (modify parent_software_image)
3. Add drift test for resource deletion (external delete detection)

**Validation**: 10+ passing drift detection tests covering 80% of attributes

### Phase 3: Enhanced Destroy Testing (P2 - Week 2)

**Goal**: Improve destroy test coverage and robustness

**Deliverables**:
1. Enhance CheckDestroy functions:
   - Add verification for orphaned dependencies
   - Add logging for resources found/not found
   - Add timeout handling
2. Add destroy edge case tests:
   - Test destroy with resource in transitional state
   - Test idempotent destroy (run twice, verify success)
   - Test destroy with force flag combinations
3. Add comprehensive PreCheck cleanup tests:
   - Verify cleanup handles all resource types
   - Verify retry logic works correctly
   - Add metrics for cleanup duration

**Validation**: Enhanced CheckDestroy functions, 5+ new destroy edge case tests

### Phase 4: Documentation and Patterns (P3 - Week 2)

**Goal**: Document test patterns for future resource development

**Deliverables**:
1. Update CLAUDE.md with drift detection test patterns
2. Update AGENTS.md with CheckDestroy and PreCheck patterns
3. Create test coverage matrix showing drift/destroy coverage per resource
4. Document BCM-specific test considerations (async ops, eventual consistency)

**Validation**: Complete documentation, test coverage matrix

### Phase 5: Continuous Improvement (P3 - Ongoing)

**Goal**: Maintain and improve test quality over time

**Deliverables**:
1. Add test coverage CI checks (fail if coverage drops below 80%)
2. Add test performance monitoring (alert if tests take >5min)
3. Periodic review of test failures and flaky tests
4. Update tests when new BCM API patterns discovered

**Validation**: CI checks passing, test suite stable

## Acceptance Checklist *(definition of done)*

This feature is complete when:

- [ ] All 2 existing resource types have comprehensive drift detection tests (string, bool, list, object attributes)
- [ ] All 2 existing resource types have enhanced CheckDestroy with dependency verification
- [ ] All 2 existing resource types have robust PreCheck cleanup with retry logic
- [ ] Test coverage achieves 80% of non-computed attributes for drift detection
- [ ] At least 5 drift scenarios tested per resource (string, bool, list, object, deletion)
- [ ] At least 3 destroy scenarios tested per resource (basic, force, idempotent)
- [ ] All drift tests use ExpectNonEmptyPlan to verify plan shows changes
- [ ] All CheckDestroy functions verify via BCM API read, not just state check
- [ ] All PreCheck functions use exponential backoff for async operations
- [ ] Documentation updated in CLAUDE.md and AGENTS.md with test patterns
- [ ] Test coverage matrix created showing drift/destroy coverage
- [ ] All acceptance tests pass consistently without manual cleanup
- [ ] Code review completed with focus on test quality and coverage
