# Implementation Plan: Deletion Order Validation

**Branch**: `002-deletion-order-validation` | **Date**: 2025-11-24 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/002-deletion-order-validation/spec.md`

## Summary

Implement deletion order validation to prevent BCM database corruption from orphaned cross-references. The feature addresses three critical areas: (1) cleanup scripts that delete resources in wrong order causing database corruption, (2) provider Delete methods that allow deletion of resources with active dependencies, and (3) test infrastructure CheckDestroy functions that fail to clean up resources properly. The solution involves fixing deletion order in cleanup scripts, adding pre-deletion dependency checks to provider resources, and enhancing test infrastructure with ordered cleanup helpers.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**:
- terraform-plugin-framework v1.16.1 (resource implementation)
- terraform-plugin-testing v1.13.3 (acceptance tests)
- BCM JSON-RPC API (cookie-based authentication)

**Storage**: BCM API backend (JSON-RPC over HTTPS)
**Testing**:
- Acceptance tests with TF_ACC=1
- Bash test scripts for cleanup validation
- Manual testing against live BCM cluster

**Target Platform**: Linux server (provider binary), BCM cluster 172.21.15.254:8081
**Project Type**: Terraform Provider (single Go project)
**Performance Goals**:
- Dependency checks must complete within 5 seconds for typical category (< 100 devices)
- CheckDestroy must complete within 30 seconds (per FR-011 from spec)
- Cleanup scripts must handle 1000+ resources within reasonable timeframe

**Constraints**:
- BCM API does not enforce referential integrity (allows orphaned references)
- BCM API lacks dedicated dependency check methods (must query and filter)
- Eventual consistency requires retry logic with exponential backoff
- Must maintain backward compatibility (no breaking changes to existing resources)

**Scale/Scope**:
- Support for 5 resource types: Software Images, Categories, Devices, Networks, Kubernetes Clusters
- 4 cleanup scripts to fix
- 2 provider resources to enhance (Category, SoftwareImage)
- 10+ acceptance tests to update/create
- Expected deployment: clusters with 10-1000 resources

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### TDD Compliance

**Rule**: "Tests first, always. RED-GREEN-REFACTOR in parallel batches."

**Status**: PASS - Feature design follows TDD workflow
- RED Phase: Write failing tests for dependency validation
- GREEN Phase: Implement minimal dependency check logic
- REFACTOR Phase: Extract shared helpers and improve error messages

**Evidence**:
- Spec includes comprehensive test strategy (Phase 1-3)
- Test approach defined before implementation
- Parallel execution possible (multiple resources, multiple scripts)

### Complexity Limits

**Rule**: "One new abstraction per feature. Maximum 3 components."

**Status**: PASS - Feature introduces minimal new abstractions
- Component 1: Dependency check helper functions (shared logic)
- Component 2: Error message builder (formatting utility)
- Component 3: Ordered CheckDestroy helper (test infrastructure)

Total: 3 components (at limit but justified)

### Implementation Guidelines

**Rule**: "Minimal viable implementation. Solve stated problem only."

**Status**: PASS - Scope is tightly focused
- IN SCOPE: Deletion order validation, dependency checks, error messages
- OUT OF SCOPE: Automatic dependency resolution, graphical visualization, cross-provider tracking

**Justification**: Feature solves specific problem (database corruption) without over-engineering.

## Project Structure

### Documentation (this feature)

```text
specs/002-deletion-order-validation/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (dependency check patterns)
├── data-model.md        # Phase 1 output (dependency graph model)
├── quickstart.md        # Phase 1 output (developer quick start)
├── contracts/           # Phase 1 output (API contract examples)
│   ├── check-devices-in-category.json
│   ├── check-categories-using-image.json
│   └── dependency-error-response.json
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
# Terraform Provider Project Structure
internal/provider/
├── bcm_client.go                      # JSON-RPC client (existing)
├── test_helpers.go                    # Test utilities (MODIFY - add dependency checks)
├── dependency_helpers.go              # NEW - Dependency check functions
├── error_messages.go                  # NEW - Error message formatting
├── resource_cmdevice_category.go      # MODIFY - Add dependency validation to Delete
├── resource_cmdevice_category_test.go # MODIFY - Add dependency validation tests
├── resource_cmpart_softwareimage.go   # MODIFY - Add dependency validation to Delete
├── resource_cmpart_softwareimage_test.go # MODIFY - Add dependency validation tests
└── [other resources remain unchanged]

scripts/
├── cleanup-basic-resources.sh         # MODIFY - Fix deletion order
├── cleanup-before-tests.sh            # MODIFY - Fix deletion order
├── cleanup-test-resources-auto.sh     # MODIFY - Fix deletion order
├── cleanup-test-resources-safe.sh     # MODIFY - Fix deletion order
├── test-cleanup-deletion-order.sh     # NEW - Validate deletion order
└── test-cleanup-dry-run.sh            # NEW - Test dry-run mode

# Project root files
.claudeignore                          # MODIFY - Add ai_reports/ if using reports
```

**Structure Decision**: This is a Terraform Provider project following the standard HashiCorp provider layout. Implementation files are added to `internal/provider/` directory with test files colocated. Cleanup scripts are in top-level `scripts/` directory. No new directories are needed - all changes integrate into existing structure.

## Complexity Tracking

No constitutional violations requiring justification. Feature stays within complexity limits:
- 3 components (at limit but necessary for separation of concerns)
- TDD workflow followed throughout
- Minimal implementation solving stated problem only

## Phase 0: Research & Discovery

**Goal**: Resolve all "NEEDS CLARIFICATION" items from Technical Context and determine optimal dependency check implementation patterns.

### Research Tasks

1. **BCM API Dependency Check Patterns**
   - Task: "Research how to efficiently check for dependent resources in BCM API"
   - Questions:
     - Does BCM API have a dedicated dependency check method?
     - What is the most efficient way to query devices in a category?
     - Can we query categories using a specific software image?
     - What is the performance impact of these queries on large installations?
   - Deliverable: Document BCM API methods for dependency checking with performance characteristics

2. **Force Parameter Behavior**
   - Task: "Research BCM API force parameter behavior in remove methods"
   - Questions:
     - Does `removeCategory(uuid, force=true)` bypass dependency checks in BCM?
     - What happens to dependent resources when force deletion is used?
     - Are there any BCM API side effects from force deletion?
   - Deliverable: Document force parameter behavior and implications

3. **Error Response Patterns**
   - Task: "Research BCM API error response formats for dependency violations"
   - Questions:
     - What error format does BCM return when deletion fails due to dependencies?
     - Are there specific error codes or keywords to detect?
     - Can we parse these errors to extract dependent resource information?
   - Deliverable: Document error response formats and parsing strategies

4. **Eventual Consistency Timing**
   - Task: "Determine optimal retry timing for deletion verification"
   - Questions:
     - How long does BCM take to propagate deletion operations?
     - What is the typical consistency window for getNodes/getCategories after removal?
     - What is an appropriate timeout for CheckDestroy verification?
   - Deliverable: Recommended retry schedule and timeout values

5. **Cleanup Script Best Practices**
   - Task: "Research bash script patterns for robust API-based cleanup"
   - Questions:
     - How should we implement dry-run mode in bash scripts?
     - What is the best way to handle partial deletion failures?
     - Should we implement rate limiting between deletion batches?
   - Deliverable: Bash script patterns and best practices

### Output Artifact

`research.md` containing:
- BCM API dependency check methods and performance characteristics
- Force parameter behavior documentation
- Error response parsing strategies
- Retry schedule and timeout recommendations
- Bash script implementation patterns

## Phase 1: Design & Contracts

**Prerequisites**: `research.md` complete

### Design Artifacts

#### 1. Data Model (`data-model.md`)

**Dependency Graph Entity**:
```
Resource Dependency Hierarchy:
- Software Images (Level 0 - no dependencies)
  └─ Categories (Level 1 - depend on Software Images)
      └─ Devices (Level 2 - depend on Categories)

Independent Resources:
- Networks (Level 0 - no dependencies from other resources)
- Kubernetes Clusters (Level 0 - no dependencies from other resources)

Deletion Order (reverse dependency order):
1. Devices (highest level - delete first)
2. Kubernetes Clusters (independent)
3. Networks (independent)
4. Categories (mid-level)
5. Software Images (lowest level - delete last)
```

**Dependency Check Result Model**:
```go
type DependencyCheckResult struct {
    HasDependencies bool
    DependentCount  int
    DependentType   string // "devices", "categories", etc.
    Identifiers     []string // Resource names or UUIDs
    ErrorMessage    string // Formatted user-facing message
}
```

#### 2. API Contracts (`contracts/`)

**Check Devices in Category** (`check-devices-in-category.json`):
```json
{
  "request": {
    "service": "cmdevice",
    "call": "getNodes",
    "args": []
  },
  "response_filtering": {
    "field": "category",
    "match_value": "<category-uuid>",
    "extract_fields": ["uuid", "hostname"]
  },
  "example_response": [
    {
      "uuid": "device-uuid-1",
      "hostname": "node01",
      "category": "category-uuid"
    }
  ]
}
```

**Check Categories Using Software Image** (`check-categories-using-image.json`):
```json
{
  "request": {
    "service": "cmdevice",
    "call": "getCategories",
    "args": []
  },
  "response_filtering": {
    "field": "softwareimage",
    "match_value": "<image-name>",
    "extract_fields": ["uuid", "name"]
  },
  "example_response": [
    {
      "uuid": "category-uuid-1",
      "name": "default",
      "softwareimage": "Rocky-8.10-NVIDIA-GPU-545"
    }
  ]
}
```

**Dependency Error Response** (`dependency-error-response.json`):
```json
{
  "error_title": "Category In Use - Cannot Delete",
  "error_detail": "Category 'test-category' cannot be deleted because it has 3 device(s) assigned.\n\nDependent devices:\n  - node01 (uuid: device-1)\n  - node02 (uuid: device-2)\n  - node03 (uuid: device-3)\n\nResolution options:\n  1. Reassign devices to another category before deleting\n  2. Delete the dependent devices first\n  3. Set 'force = true' to delete anyway (WARNING: will orphan device references)",
  "diagnostic_severity": "Error"
}
```

#### 3. Implementation Contracts (`quickstart.md`)

Developer quick start guide covering:
- How to add dependency checks to a new resource
- How to use dependency helper functions
- How to format error messages for users
- How to write drift detection tests with dependency validation
- How to update cleanup scripts with correct deletion order

### Agent Context Update

Run `.specify/scripts/bash/update-agent-context.sh copilot` to update:
- New dependency check patterns
- Error message formatting conventions
- Deletion order requirements
- Test infrastructure enhancements

## Phase 2: Implementation Planning

**Prerequisites**: Phase 1 design artifacts complete

### Component Breakdown

#### Component 1: Dependency Check Helpers (`dependency_helpers.go`)

**Purpose**: Shared dependency validation logic for provider resources

**Functions**:
```go
// CheckDevicesInCategory queries BCM for devices using the specified category
func CheckDevicesInCategory(ctx context.Context, client *BCMClient, categoryUUID string) (*DependencyCheckResult, error)

// CheckCategoriesUsingImage queries BCM for categories using the specified software image
func CheckCategoriesUsingImage(ctx context.Context, client *BCMClient, imageName string) (*DependencyCheckResult, error)

// HasDependencies is a convenience function that returns true if result has dependencies
func (r *DependencyCheckResult) HasDependencies() bool

// FormatErrorMessage formats dependency check results into user-facing error message
func (r *DependencyCheckResult) FormatErrorMessage(resourceType, resourceName string) string
```

**Design Decisions**:
- Use BCM's existing getNodes/getCategories methods with client-side filtering
- Return structured results (not just error strings) for flexibility
- Include resource identifiers in results for detailed error messages
- Handle BCM API errors gracefully (treat as warnings, not blockers)

#### Component 2: Error Message Builder (`error_messages.go`)

**Purpose**: Format user-friendly, actionable error messages for dependency violations

**Functions**:
```go
// BuildDependencyError creates formatted error message with resolution options
func BuildDependencyError(resourceType, resourceName, dependentType string, dependents []string) string

// BuildForceDeleteionWarning creates warning message for force deletion
func BuildForceDeleteionWarning(resourceType, resourceName string) string

// TruncateDependentList limits dependent list length for readability
func TruncateDependentList(dependents []string, maxShow int) (truncated []string, remaining int)
```

**Message Template**:
```
{ResourceType} In Use - Cannot Delete

{ResourceType} '{name}' cannot be deleted because it has {count} {dependentType}(s) assigned.

Dependent {dependentType}:
  - {name1} (uuid: {uuid1})
  - {name2} (uuid: {uuid2})
  ... (showing first 10 of {total})

Resolution options:
  1. {specific action for resource type}
  2. Delete the dependent {dependentType} first
  3. Set 'force = true' to delete anyway (WARNING: will orphan references)
```

#### Component 3: Ordered CheckDestroy Helper (`test_helpers.go` enhancement)

**Purpose**: Ensure test cleanup happens in correct dependency order

**Functions**:
```go
// TestAccCheckResourcesDestroyOrdered verifies resources are deleted in correct order
func TestAccCheckResourcesDestroyOrdered(t *testing.T) func(*terraform.State) error

// GroupResourcesByType extracts resources from state and groups by type
func GroupResourcesByType(s *terraform.State) map[string][]string

// VerifyResourcesDeleted checks multiple resources with exponential backoff
func VerifyResourcesDeleted(ctx context.Context, client *BCMClient, resources map[string][]string) error
```

**Deletion Order Logic**:
```go
// Delete in dependency order (highest to lowest)
deletionOrder := []string{
    "bcm_cmdevice_device",         // Level 2
    "bcm_cmkube_cluster",           // Independent
    "bcm_cmnet_network",            // Independent
    "bcm_cmdevice_category",        // Level 1
    "bcm_cmpart_softwareimage",     // Level 0
}
```

### File Modifications

**Modified Files**:
1. `internal/provider/resource_cmdevice_category.go` - Add dependency check to Delete method
2. `internal/provider/resource_cmpart_softwareimage.go` - Add dependency check to Delete method
3. `internal/provider/test_helpers.go` - Add ordered CheckDestroy helper
4. `scripts/cleanup-basic-resources.sh` - Fix deletion order
5. `scripts/cleanup-before-tests.sh` - Fix deletion order
6. `scripts/cleanup-test-resources-auto.sh` - Fix deletion order
7. `scripts/cleanup-test-resources-safe.sh` - Fix deletion order

**New Files**:
1. `internal/provider/dependency_helpers.go` - Dependency check functions
2. `internal/provider/error_messages.go` - Error formatting functions
3. `scripts/test-cleanup-deletion-order.sh` - Deletion order validation test
4. `scripts/test-cleanup-dry-run.sh` - Dry-run mode test

### TDD Implementation Workflow

#### RED Phase: Write Failing Tests

**Parallel Test Creation**:
```go
// Test 1: Category dependency validation
TestAccCMDeviceCategory_DeleteWithDependencies
  - Create category with device
  - Attempt to delete category
  - Expect error with device identifiers

// Test 2: Software image dependency validation
TestAccCMPartSoftwareImage_DeleteWithDependencies
  - Create image with category
  - Attempt to delete image
  - Expect error with category identifiers

// Test 3: Force deletion
TestAccCMDeviceCategory_DeleteWithForce
  - Create category with device
  - Delete with force=true
  - Verify category deleted, device remains

// Test 4: Force deletion warning
TestAccCMPartSoftwareImage_DeleteWithForce
  - Create image with category
  - Delete with force=true (check logs for warning)
  - Verify image deleted, category orphaned

// Test 5: Ordered CheckDestroy
TestAccMultipleResources_CheckDestroyOrder
  - Create image, category, device
  - Destroy all
  - Verify CheckDestroy deletes in correct order
```

**Run Tests** (expect failures):
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "DeleteWith"
```

#### GREEN Phase: Minimal Implementation

**Step 1: Implement Dependency Helpers**
```go
// dependency_helpers.go - minimal implementation
func CheckDevicesInCategory(ctx, client, categoryUUID) (*DependencyCheckResult, error) {
    body, err := client.CallJSONRPC(ctx, "cmdevice", "getNodes")
    // Parse and filter by category field
    // Return result with device identifiers
}
```

**Step 2: Implement Error Formatting**
```go
// error_messages.go - minimal implementation
func BuildDependencyError(resourceType, resourceName, dependentType, dependents) string {
    // Format basic error message with resource list
    // Include resolution options
}
```

**Step 3: Add Dependency Checks to Resources**
```go
// resource_cmdevice_category.go - Delete method enhancement
func (r *CMDeviceCategoryResource) Delete(ctx, req, resp) {
    // Get state
    var state CMDeviceCategoryResourceModel
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)

    // Check for dependencies (unless force=true)
    if !state.Force.ValueBool() {
        result, err := CheckDevicesInCategory(ctx, r.client, state.UUID.ValueString())
        if err != nil {
            resp.Diagnostics.AddWarning("Dependency check failed", err.Error())
        } else if result.HasDependencies() {
            resp.Diagnostics.AddError(
                "Category In Use - Cannot Delete",
                result.FormatErrorMessage("Category", state.Name.ValueString()),
            )
            return
        }
    }

    // Proceed with deletion
    _, err := r.client.CallJSONRPC(ctx, "cmdevice", "removeCategory", state.UUID.ValueString(), state.Force.ValueBool())
    // ... error handling
}
```

**Step 4: Implement Ordered CheckDestroy**
```go
// test_helpers.go - enhancement
func TestAccCheckResourcesDestroyOrdered(t *testing.T) func(*terraform.State) error {
    return func(s *terraform.State) error {
        client := createTestBCMClient(t)
        resources := GroupResourcesByType(s)
        return VerifyResourcesDeleted(context.Background(), client, resources)
    }
}
```

**Run Tests** (expect passes):
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "DeleteWith"
```

#### REFACTOR Phase: Improve Quality

**Improvements**:
1. Extract common error formatting patterns
2. Add comprehensive logging for dependency checks
3. Optimize BCM API queries (batch if possible)
4. Add detailed comments and documentation
5. Improve error message readability
6. Add performance monitoring for dependency checks

**Run Tests** (verify still passing):
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

### Cleanup Script Implementation

**Pattern for All Scripts**:
```bash
#!/usr/bin/env bash
# Correct deletion order: Devices -> Clusters -> Networks -> Categories -> Images

# Add dry-run mode support
DRY_RUN=${DRY_RUN:-false}

# Add health check function
check_bcm_health() {
    curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"cmgui","call":"getSystemStatus"}' > /dev/null
    if [ $? -ne 0 ]; then
        echo "ERROR: BCM health check failed"
        exit 1
    fi
}

# Delete in correct order
echo "Step 1: Deleting devices..."
delete_resources "devices" "cmdevice" "getNodes" "removeNodes" "hostname"
check_bcm_health

echo "Step 2: Deleting Kubernetes clusters..."
delete_resources "clusters" "cmkube" "getClusters" "removeClusters" "name"
check_bcm_health

echo "Step 3: Deleting networks..."
delete_resources "networks" "cmnet" "getNetworks" "removeNetworks" "name"
check_bcm_health

echo "Step 4: Deleting categories..."
delete_resources "categories" "cmdevice" "getCategories" "removeCategories" "name"
check_bcm_health

echo "Step 5: Deleting software images..."
delete_resources "images" "cmpart" "getSoftwareImages" "removeSoftwareImages" "name"
check_bcm_health
```

## Phase 3: Testing & Validation

### Test Coverage

**Acceptance Tests** (internal/provider/):
1. TestAccCMDeviceCategory_DeleteWithDependencies - Block deletion with devices
2. TestAccCMDeviceCategory_DeleteWithForce - Force delete with devices
3. TestAccCMPartSoftwareImage_DeleteWithDependencies - Block deletion with categories
4. TestAccCMPartSoftwareImage_DeleteWithForce - Force delete with categories
5. TestAccMultipleResources_CheckDestroyOrder - Verify ordered cleanup
6. TestAccCMDeviceCategory_DeleteNoDependencies - Allow deletion when no dependencies
7. TestAccCMPartSoftwareImage_DeleteNoDependencies - Allow deletion when no dependencies

**Cleanup Script Tests** (scripts/):
1. test-cleanup-deletion-order.sh - Verify correct deletion order
2. test-cleanup-dry-run.sh - Verify dry-run mode works

### Test Execution Strategy

**Unit Tests** (parallel):
```bash
go test -v -cover ./internal/provider/dependency_helpers_test.go
go test -v -cover ./internal/provider/error_messages_test.go
```

**Acceptance Tests** (parallel batches):
```bash
# Batch 1: Category tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory.*Delete"

# Batch 2: Software image tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMPartSoftwareImage.*Delete"

# Batch 3: Multi-resource tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "MultipleResources"
```

**Script Tests** (sequential):
```bash
# Test deletion order validation
BCM_ENDPOINT=... BCM_USERNAME=... BCM_PASSWORD=... ./scripts/test-cleanup-deletion-order.sh

# Test dry-run mode
DRY_RUN=1 ./scripts/cleanup-basic-resources.sh
```

### Success Criteria Validation

| Success Criteria | Validation Method | Target |
|-----------------|-------------------|--------|
| SC-001: Correct deletion order | test-cleanup-deletion-order.sh | 100% pass |
| SC-002: Zero corruption incidents | Manual testing + CI monitoring | 0 incidents |
| SC-003: Block orphaning deletions | Acceptance tests | 100% blocked |
| SC-004: Actionable error messages | Manual review + tests | 100% contain resolution steps |
| SC-005: CheckDestroy success rate | Test suite execution | 100% success |
| SC-006: Dependency check warnings | Test logs + manual testing | All failures produce warnings |
| SC-007: Documentation clarity | Peer review | Approved by team |
| SC-008: Dry-run support | test-cleanup-dry-run.sh | All scripts support |

## Risk Mitigation

### Risk 1: BCM API Performance Impact

**Risk**: Dependency checks (querying all devices/categories) may be slow on large clusters

**Mitigation**:
- Implement query caching for same-transaction checks
- Add timeout configuration (default 5s)
- Log query performance metrics
- Document performance expectations in quickstart.md
- Consider skip-check flag for emergency use

**Fallback**: If dependency check fails, log warning and allow user to proceed with force=true

### Risk 2: Eventual Consistency Issues

**Risk**: Resource deletion may not be immediately visible, causing false positives in CheckDestroy

**Mitigation**:
- Use exponential backoff (1s, 2s, 4s, 8s) for up to 15s total
- Document consistency expectations
- Add detailed logging for retry attempts
- Consider configurable retry count

**Fallback**: If resource still exists after retries, provide detailed error message with guidance

### Risk 3: Backward Compatibility

**Risk**: Adding dependency checks might break existing workflows that depend on force deletion

**Mitigation**:
- Make force parameter optional with default=false
- Add deprecation warnings in documentation
- Provide migration guide for existing configurations
- Test with real customer configurations before release

**Fallback**: Provide configuration flag to disable dependency checks globally if needed

### Risk 4: Error Message Localization

**Risk**: Error messages may contain BCM-specific terminology unfamiliar to users

**Mitigation**:
- Use consistent, clear language in error messages
- Include examples of resolution steps
- Add links to documentation
- Test messages with non-technical users

**Fallback**: Provide detailed troubleshooting guide in documentation

### Risk 5: Cleanup Script Reliability

**Risk**: Bash scripts may fail partway through, leaving some resources

**Mitigation**:
- Implement health checks between batches
- Add detailed error logging
- Support resume-from-failure mode
- Provide manual cleanup instructions

**Fallback**: Document manual cleanup procedures for worst-case scenarios

## Documentation Updates

### Files to Update

1. **CLAUDE.md** - Add deletion order requirements section
2. **examples/resources/bcm_cmdevice_category/resource.tf** - Add force parameter example
3. **examples/resources/bcm_cmpart_softwareimage/resource.tf** - Add force parameter example
4. **README.md** - Add dependency management section
5. **quickstart.md** (new) - Developer quick start for dependency checks

### Documentation Sections

**Dependency Graph Documentation**:
```markdown
## Resource Dependencies

BCM resources have the following dependency hierarchy:

- Software Images (no dependencies)
  └─ Categories (depend on Software Images)
      └─ Devices (depend on Categories)

Independent resources:
- Networks
- Kubernetes Clusters

### Deletion Order

Resources must be deleted in this order:
1. Devices
2. Kubernetes Clusters
3. Networks
4. Categories
5. Software Images

### Force Deletion

Set `force = true` to bypass dependency checks:

```hcl
resource "bcm_cmdevice_category" "example" {
  name               = "test"
  management_network = bcm_cmnet_network.mgmt.id
  force              = true  # Delete even if devices are assigned
}
```

**WARNING**: Force deletion may create orphaned references in the BCM database.
```

**Troubleshooting Guide**:
```markdown
## Troubleshooting Deletion Errors

### "Category In Use - Cannot Delete"

**Cause**: Category has devices assigned to it

**Resolution**:
1. Reassign devices to another category:
   ```hcl
   resource "bcm_cmdevice_device" "node01" {
     category = bcm_cmdevice_category.new_category.id
   }
   ```

2. Delete devices first:
   ```bash
   terraform destroy -target=bcm_cmdevice_device.node01
   terraform destroy -target=bcm_cmdevice_category.old_category
   ```

3. Force delete (use with caution):
   ```hcl
   resource "bcm_cmdevice_category" "old_category" {
     force = true
   }
   ```

### "Software Image In Use - Cannot Delete"

**Cause**: Categories are using this software image

**Resolution**: [Similar structure to above]
```

## Assumptions & Constraints

### Assumptions

1. **BCM API Behavior**: BCM API does not enforce referential integrity (confirmed through testing)
2. **Dependency Detection**: Client-side filtering of getNodes/getCategories is the only way to detect dependencies
3. **Force Parameter**: BCM API accepts force parameter in remove methods
4. **Resource Query Performance**: Querying all nodes/categories completes within 5 seconds for typical installations
5. **Test Resource Prefixes**: Test resources consistently use `citest-*` and `tftest-*` prefixes
6. **Eventual Consistency**: BCM API propagates deletions within 15 seconds maximum
7. **Error Message Format**: Users prefer structured error messages with resolution options

### Constraints

1. **No Breaking Changes**: Existing Terraform configurations must continue to work
2. **Performance**: Dependency checks must not add significant latency to destroy operations
3. **Backward Compatibility**: Force parameter must be optional (default false)
4. **Test Duration**: Acceptance tests must complete within 120 minutes total
5. **API Limitations**: Must work within BCM JSON-RPC API constraints (no batch operations)
6. **Platform Support**: Scripts must run on Linux (Bash 4.0+)

## Out of Scope

The following items are explicitly excluded from this feature:

1. **Automatic Dependency Resolution** - Feature does not automatically delete dependent resources
2. **Graphical Visualization** - No UI for viewing dependency graphs
3. **Cross-Provider Dependencies** - Only tracks dependencies within BCM resources
4. **Historical Audit Logging** - No persistent log of deletion attempts
5. **Undo/Rollback** - No ability to reverse deletions
6. **Extended Resource Types** - Only covers 5 core types (Images, Categories, Devices, Networks, Clusters)
7. **Real-Time Notifications** - No webhooks or event system for dependency changes
8. **Performance Optimization** - No caching or indexing for large deployments (>1000 resources)
9. **Circular Dependency Detection** - Assumes BCM API prevents circular references
10. **Multi-Cluster Support** - Dependency tracking limited to single BCM cluster

## Next Steps

After plan approval:

1. **Phase 0 Execution**: Run research tasks to resolve NEEDS CLARIFICATION items
2. **Phase 1 Execution**: Generate data-model.md, contracts/, quickstart.md
3. **Agent Context Update**: Run update-agent-context.sh to propagate design decisions
4. **Constitution Re-Check**: Verify design still complies with TDD rules
5. **Task Generation**: Run `/speckit.tasks` to create actionable task list
6. **Implementation**: Execute tasks using `/speckit.implement` command

## Approval Checklist

- [ ] Technical Context complete (all NEEDS CLARIFICATION resolved after Phase 0)
- [ ] Constitution Check passed (TDD compliance verified)
- [ ] Project Structure documented (file locations identified)
- [ ] Component Breakdown complete (3 components defined)
- [ ] TDD Workflow defined (RED-GREEN-REFACTOR phases)
- [ ] Risk Mitigation strategies documented
- [ ] Success Criteria validation methods defined
- [ ] Documentation plan complete
- [ ] Out of Scope items explicitly listed
