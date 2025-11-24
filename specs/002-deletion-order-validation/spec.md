# Feature Specification: Deletion Order Validation

**Feature Branch**: `002-deletion-order-validation`
**Created**: 2025-11-24
**Status**: Draft
**Input**: Implement deletion order validation to prevent BCM database corruption

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Safe Cleanup Script Execution (Priority: P1)

As a developer running cleanup scripts, I need the scripts to delete resources in the correct dependency order so that I don't corrupt the BCM database with orphaned cross-references.

**Why this priority**: This is the highest priority because incorrect deletion order in cleanup scripts directly causes database corruption, test failures, and requires manual BCM database recovery. This affects all developers and CI/CD pipelines.

**Independent Test**: Can be fully tested by running each cleanup script against a BCM cluster with test resources and verifying: (1) resources are deleted in correct order, (2) no BCM errors occur, (3) no orphaned references remain in the database.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with test devices, categories, and software images, **When** I run `cleanup-basic-resources.sh`, **Then** resources are deleted in the order: Devices → Clusters → Networks → Categories → Software Images, and no database corruption occurs
2. **Given** a BCM cluster with interdependent resources, **When** I run any cleanup script with `DRY_RUN=1`, **Then** the script shows which resources would be deleted in what order without making any changes
3. **Given** a cleanup script execution that encounters a dependency error, **When** the script continues, **Then** it logs the error clearly and proceeds with remaining deletions in correct order
4. **Given** resources that cannot be deleted due to dependencies, **When** cleanup script runs, **Then** it reports exactly which resources are blocking deletion and suggests resolution steps

---

### User Story 2 - Provider Delete Method Protection (Priority: P2)

As a Terraform user deleting BCM resources, I need the provider to validate dependencies before deletion and provide clear error messages so that I understand why deletion failed and how to fix it.

**Why this priority**: This prevents users from accidentally corrupting their BCM database through normal Terraform operations. While less frequent than cleanup script issues, this affects production usage and user experience.

**Independent Test**: Can be fully tested by attempting to delete resources with dependencies via Terraform and verifying: (1) deletion is blocked when dependencies exist, (2) error message clearly explains the problem, (3) error message suggests actionable solutions.

**Acceptance Scenarios**:

1. **Given** a category with devices assigned, **When** I run `terraform destroy` on the category, **Then** Terraform blocks the deletion and shows an error message listing the dependent devices with resolution options
2. **Given** a software image referenced by categories, **When** I attempt to delete it via Terraform, **Then** deletion is blocked and the error message lists which categories depend on it
3. **Given** a category with devices, **When** I set `force = true` and run `terraform apply`, **Then** the category is deleted with a warning about potential orphaned references, and devices remain but lose their category reference
4. **Given** a resource with no dependencies, **When** I run `terraform destroy`, **Then** deletion succeeds immediately without dependency checks
5. **Given** a dependency check that fails due to BCM API error, **When** I attempt deletion, **Then** I receive a warning that dependency check failed but deletion can proceed with `force = true`

---

### User Story 3 - Test Infrastructure Reliability (Priority: P3)

As a developer writing acceptance tests, I need test CheckDestroy functions to clean up resources in the correct order so that test cleanup doesn't leave orphaned resources or fail due to dependency violations.

**Why this priority**: This improves test reliability and prevents test flakiness. While important for developer experience, it's less critical than preventing production database corruption.

**Independent Test**: Can be fully tested by running acceptance tests and verifying: (1) CheckDestroy functions delete resources in correct order, (2) all test resources are cleaned up, (3) no orphaned references remain after tests complete.

**Acceptance Scenarios**:

1. **Given** an acceptance test that creates devices, categories, and images, **When** the test completes, **Then** CheckDestroy deletes resources in order: Devices → Clusters → Networks → Categories → Images
2. **Given** a CheckDestroy function that encounters eventual consistency delays, **When** verifying deletion, **Then** it uses exponential backoff retries and provides detailed error messages if resources remain after max retries
3. **Given** a test that fails mid-execution, **When** CheckDestroy runs, **Then** it successfully cleans up all created resources despite the test failure
4. **Given** multiple test resources with complex dependencies, **When** CheckDestroy runs, **Then** it provides detailed logging showing which resources were deleted and in what order

---

### Edge Cases

- What happens when a resource is deleted externally (outside Terraform) before the provider attempts deletion?
- How does the system handle circular dependencies or unexpected dependency chains?
- What happens when BCM API returns inconsistent dependency information?
- How does the system handle partial deletion failures (some resources deleted, others failed)?
- What happens when a user sets `force = true` on multiple dependent resources simultaneously?
- How does the system handle very large dependency trees (e.g., 100+ devices in one category)?
- What happens when BCM API is slow or unresponsive during dependency checks?

## Requirements *(mandatory)*

### Functional Requirements

#### Cleanup Script Requirements

- **FR-001**: All cleanup scripts MUST delete resources in the dependency order: Devices → Kubernetes Clusters → Networks → Categories → Software Images
- **FR-002**: Cleanup scripts MUST support a `DRY_RUN` mode that shows what would be deleted without making changes
- **FR-003**: Cleanup scripts MUST provide detailed logging showing which resources are being deleted and in what order
- **FR-004**: Cleanup scripts MUST handle partial deletion failures gracefully and continue with remaining deletions
- **FR-005**: Cleanup scripts MUST validate BCM health between deletion batches and abort if BCM becomes unresponsive

#### Provider Delete Method Requirements

- **FR-006**: Provider Delete methods MUST perform dependency checks before deletion (unless `force = true`)
- **FR-007**: Provider Delete methods MUST provide actionable error messages when dependencies block deletion, including:
  - List of specific dependent resources (with names/IDs)
  - Clear explanation of why deletion failed
  - Resolution options (reassign/delete dependencies OR use `force = true`)
  - Warning about consequences of force deletion
- **FR-008**: Provider Delete methods MUST support a `force` parameter that bypasses dependency checks
- **FR-009**: Provider Delete methods with `force = true` MUST log warnings about potential orphaned references
- **FR-010**: Provider Delete methods MUST handle "already deleted" cases gracefully (idempotent deletion)

#### Test Infrastructure Requirements

- **FR-011**: Test CheckDestroy functions MUST delete resources in the correct dependency order
- **FR-012**: Test CheckDestroy functions MUST use exponential backoff for eventual consistency verification
- **FR-013**: Test CheckDestroy functions MUST provide detailed error messages showing which resources failed to delete and why
- **FR-014**: Test CheckDestroy functions MUST verify that all test-created resources are fully removed

### Key Entities

- **Dependency Graph**: Represents the relationships between BCM resources
  - Software Images (no dependencies)
  - Categories (depend on Software Images)
  - Devices (depend on Categories)
  - Networks (independent)
  - Kubernetes Clusters (independent)

- **Deletion Order**: The sequence in which resources must be deleted to avoid orphaned references
  - Order: Devices → Clusters → Networks → Categories → Software Images
  - Rationale: Delete dependent resources before their dependencies

- **Dependency Validation Result**: Information returned from dependency checks
  - Has dependencies (boolean)
  - List of dependent resource identifiers
  - Dependency type (devices, categories, etc.)
  - Actionable error message for user

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All cleanup scripts delete resources in correct dependency order without manual intervention
- **SC-002**: Zero database corruption incidents from cleanup script execution
- **SC-003**: Provider Delete methods block 100% of deletions that would create orphaned references (when `force = false`)
- **SC-004**: Error messages from blocked deletions contain specific resource identifiers and resolution steps in 100% of cases
- **SC-005**: Test CheckDestroy functions successfully clean up resources in correct order with 100% success rate
- **SC-006**: Dependency check failures provide warnings rather than blocking operations, allowing users to proceed with `force = true`
- **SC-007**: Documentation clearly explains dependency graph and deletion order requirements
- **SC-008**: All cleanup scripts support dry-run mode for safe validation before execution

## API Contracts

### BCM JSON-RPC Dependency Check Methods

The following BCM API methods will be used for dependency validation:

#### Check Devices in Category
```json
{
  "service": "cmdevice",
  "call": "getNodes",
  "args": []
}
```
Response filtering: Filter by `category` field matching category UUID

#### Check Categories Using Software Image
```json
{
  "service": "cmdevice",
  "call": "getCategories",
  "args": []
}
```
Response filtering: Filter by `softwareimage` field matching image name

#### Check Networks Referenced by Category
```json
{
  "service": "cmdevice",
  "call": "getCategory",
  "args": ["<category-uuid>"]
}
```
Response inspection: Check `networks` field for network UUIDs

#### Verify Resource Deletion
```json
{
  "service": "<service>",
  "call": "get<Resource>",
  "args": ["<resource-id>"]
}
```
Expected: Empty result or error indicating resource not found

### Cleanup Script Execution Flow

1. Login to BCM and establish session
2. For each resource type in deletion order:
   - Query BCM for resources matching test prefixes (citest-*, tftest-*)
   - Display resources to be deleted (with dry-run option)
   - Delete resources one at a time with rate limiting
   - Verify deletion success after each operation
   - Check BCM health after batch completion
3. Perform final BCM health check
4. Report summary of deletions

### Provider Delete Method Enhancement Flow

1. Extract resource UUID and `force` parameter from state
2. If `force = false`, perform dependency check:
   - For Categories: Query devices with matching category UUID
   - For Software Images: Query categories with matching image name
   - If dependencies exist, build error message with:
     - Count of dependent resources
     - List of dependent resource identifiers
     - Resolution options
   - Return error to block deletion
3. If `force = true` or no dependencies:
   - Log warning if force deletion (potential orphaned references)
   - Call BCM API removeResource method
   - Handle "already deleted" case gracefully
4. Return success or error

## TDD Test Strategy

### Phase 1: Cleanup Script Tests

**Test Files**:
- `scripts/test-cleanup-deletion-order.sh` - Validates deletion order
- `scripts/test-cleanup-dry-run.sh` - Validates dry-run mode

**Test Approach**:
1. Create test resources in BCM with known dependencies
2. Run cleanup script with logging
3. Verify deletion order by parsing log output
4. Verify no BCM API errors occurred
5. Verify no orphaned references remain in database

**Example Test**:
```bash
# Create test resources: image → category → device
# Run cleanup script
# Parse logs to verify order: device deleted before category, category before image
# Query BCM to verify all resources removed
```

### Phase 2: Provider Delete Method Tests

**Test Files**:
- `internal/provider/resource_cmdevice_category_test.go` - Add dependency validation tests
- `internal/provider/resource_cmpart_softwareimage_test.go` - Add dependency validation tests

**Test Approach** (RED-GREEN-REFACTOR):

**RED Phase - Write Failing Tests**:
```go
func TestAccCMDeviceCategory_DeleteWithDependencies(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Step 1: Create category and device
            {
                Config: testAccCategoryWithDevice(),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", "test-category"),
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", "test-device"),
                ),
            },
            // Step 2: Attempt to delete category with device still assigned
            // Expected: Deletion blocked with clear error message
            {
                Config: testAccCategoryOnly(), // Remove category from config
                ExpectError: regexp.MustCompile(
                    "Category In Use.*cannot be deleted.*has.*device.*assigned",
                ),
            },
        },
    })
}

func TestAccCMDeviceCategory_DeleteWithForce(t *testing.T) {
    resource.Test(t, resource.TestCase{
        Steps: []resource.TestStep{
            // Step 1: Create category and device
            {
                Config: testAccCategoryWithDevice(),
                Check: resource.ComposeAggregateTestCheckFunc(...),
            },
            // Step 2: Delete category with force=true
            // Expected: Category deleted, device remains
            {
                Config: testAccCategoryWithForce(), // force = true
                Check: resource.ComposeAggregateTestCheckFunc(
                    // Verify category deleted
                    // Verify device still exists
                ),
            },
        },
    })
}
```

**GREEN Phase - Minimal Implementation**:
```go
// In resource_cmdevice_category.go Delete method
func (r *CMDeviceCategoryResource) Delete(ctx, req, resp) {
    // Get state and force parameter
    var state CMDeviceCategoryResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    force := state.Force.ValueBool()

    // Dependency check (unless force=true)
    if !force {
        devices, err := r.getDevicesInCategory(ctx, state.UUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddWarning("Dependency check failed", err.Error())
        } else if len(devices) > 0 {
            resp.Diagnostics.AddError(
                "Category In Use - Cannot Delete",
                buildDependencyErrorMessage(state.Name.ValueString(), devices),
            )
            return
        }
    }

    // Proceed with deletion
    _, err := r.client.CallJSONRPC(ctx, "cmdevice", "removeCategory", state.UUID.ValueString(), force)
    // ... error handling
}

func (r *CMDeviceCategoryResource) getDevicesInCategory(ctx, categoryUUID) ([]string, error) {
    // Query BCM for devices with matching category
    body, err := r.client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    // Filter devices by category field
    // Return list of device identifiers
}
```

**REFACTOR Phase**:
- Extract shared dependency check logic
- Improve error message formatting
- Add comprehensive logging
- Optimize BCM API queries

### Phase 3: CheckDestroy Enhancement Tests

**Test Approach**:
1. Create helper function for ordered CheckDestroy
2. Verify it's used in all resource test files
3. Add logging to track deletion order
4. Test with multiple interdependent resources

**Example Implementation**:
```go
// In test_helpers.go
func testAccCheckResourcesDestroyOrdered(t *testing.T, s *terraform.State) error {
    client := createTestBCMClient(t)

    // Group resources by type
    devices := []string{}
    categories := []string{}
    images := []string{}

    for _, rs := range s.RootModule().Resources {
        switch rs.Type {
        case "bcm_cmdevice_device":
            devices = append(devices, rs.Primary.ID)
        case "bcm_cmdevice_category":
            categories = append(categories, rs.Primary.ID)
        case "bcm_cmpart_softwareimage":
            images = append(images, rs.Primary.ID)
        }
    }

    // Verify deletion in order: devices, categories, images
    if err := verifyResourcesDeleted(ctx, client, "cmdevice", "getNodes", devices); err != nil {
        return fmt.Errorf("Devices not deleted: %w", err)
    }
    if err := verifyResourcesDeleted(ctx, client, "cmdevice", "getCategories", categories); err != nil {
        return fmt.Errorf("Categories not deleted: %w", err)
    }
    if err := verifyResourcesDeleted(ctx, client, "cmpart", "getSoftwareImages", images); err != nil {
        return fmt.Errorf("Images not deleted: %w", err)
    }

    return nil
}
```

## Implementation Phases

### Phase 1: Fix Critical Cleanup Scripts (High Priority)

**Goal**: Prevent immediate database corruption from cleanup script usage

**Tasks**:
1. Fix `cleanup-basic-resources.sh` deletion order (devices before categories before images)
2. Add dry-run mode to all cleanup scripts
3. Add health checks between deletion batches
4. Standardize error handling across all scripts
5. Add detailed logging showing deletion order

**Validation**:
- Run all cleanup scripts against test BCM cluster
- Verify correct deletion order in logs
- Verify no BCM errors
- Verify no orphaned references remain

### Phase 2: Enhance Provider Delete Methods (High Priority)

**Goal**: Protect users from accidental database corruption via Terraform

**Tasks**:
1. Implement dependency check helper methods (getDevicesInCategory, getCategoriesUsingImage)
2. Add pre-deletion dependency validation to Category.Delete()
3. Add pre-deletion dependency validation to SoftwareImage.Delete()
4. Create actionable error message builder
5. Add comprehensive logging for delete operations
6. Document `force` parameter behavior

**Validation**:
- Write and run acceptance tests for dependency validation
- Write and run acceptance tests for force deletion
- Verify error messages are clear and actionable
- Test against live BCM cluster

### Phase 3: Improve Test Infrastructure (Medium Priority)

**Goal**: Ensure test cleanup doesn't leave orphaned resources

**Tasks**:
1. Create shared CheckDestroy helper with ordered deletion
2. Update all resource tests to use ordered CheckDestroy
3. Add detailed logging to CheckDestroy functions
4. Add exponential backoff for deletion verification
5. Create test utilities for dependency verification

**Validation**:
- Run full test suite
- Verify all tests clean up successfully
- Check BCM cluster for orphaned resources after test runs

### Phase 4: Documentation (Medium Priority)

**Goal**: Help users understand and follow dependency rules

**Tasks**:
1. Document dependency graph in provider README
2. Create troubleshooting guide for deletion failures
3. Add examples of safe resource configurations
4. Document cleanup script usage patterns
5. Explain `force` parameter implications and best practices

**Validation**:
- Review documentation with team
- Test examples against BCM cluster
- Verify troubleshooting guide covers common issues

## Assumptions

1. **BCM API Behavior**: Assumption that BCM API does not enforce referential integrity on deletion (based on observed behavior)
2. **Dependency Detection**: Assumption that querying related resources is the only way to detect dependencies (BCM API doesn't provide a dedicated dependency check method)
3. **Force Parameter**: Assumption that force=true on removeResource bypasses dependency checks in BCM
4. **Resource Query Methods**: Assumption that all BCM resources can be queried by specific fields (category, softwareimage, etc.)
5. **Test Prefixes**: Assumption that test resources consistently use `citest-*` and `tftest-*` prefixes
6. **Eventual Consistency**: Assumption that BCM API may require delays between operations for consistency
7. **Error Message Patterns**: Assumption that BCM API error messages contain consistent keywords for dependency violations ("in use", "assigned", "cannot be deleted")

## Out of Scope

- Automatic dependency resolution (e.g., automatically deleting dependent resources)
- Graphical visualization of dependency relationships
- Cross-provider dependency tracking (only BCM resources)
- Historical audit logging of deletion attempts
- Undo/rollback functionality for deletions
- Dependency validation for resources beyond the core five types (Images, Categories, Devices, Networks, Clusters)
- Real-time dependency change notifications
- Performance optimization for very large deployments (>1000 resources)
