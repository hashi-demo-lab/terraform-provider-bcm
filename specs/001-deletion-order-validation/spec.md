# Feature Specification: Deletion Order Validation for BCM Resources

**Feature Branch**: `001-deletion-order-validation`
**Created**: 2025-11-24
**Status**: Draft
**Input**: User description: "Implement proper deletion order validation to prevent BCM database corruption across cleanup scripts, provider Delete methods, and test infrastructure. The BCM API does not validate resource dependencies during deletion, which can cause database corruption with orphaned cross-references."

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Safe Cleanup Script Execution (Priority: P1)

A DevOps engineer runs cleanup scripts to remove test resources from a BCM cluster. The scripts delete resources in the correct dependency order (Devices → Clusters → Networks → Categories → Software Images) to prevent database corruption from orphaned references.

**Why this priority**: This is the most critical issue causing production database corruption. Cleanup scripts with incorrect deletion order are actively corrupting BCM databases during test cleanup and development workflows.

**Independent Test**: Run cleanup-basic-resources.sh against a BCM cluster with interdependent test resources. Verify all resources are deleted successfully without BCM API errors or database corruption warnings. Check BCM database integrity after cleanup completes.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with test resources (software images, categories, networks, devices, clusters), **When** the cleanup script executes, **Then** resources are deleted in the correct order (devices first, software images last) and all deletions succeed without database errors
2. **Given** a BCM cluster with dependencies (category references software image, device references category), **When** cleanup attempts to delete in wrong order, **Then** the script detects dependency violations and reorders deletion or provides clear error messages
3. **Given** a cleanup script running with verbose logging, **When** deletion order is enforced, **Then** logs clearly show the dependency-respecting deletion sequence with reasons for the order

---

### User Story 2 - Provider Delete Method Protection (Priority: P2)

A Terraform user runs `terraform destroy` to remove infrastructure managed by the BCM provider. The provider's Delete methods validate resource dependencies before attempting deletion, providing clear error messages when dependencies block deletion, and successfully delete resources when dependencies are satisfied.

**Why this priority**: This prevents users from encountering cryptic BCM API errors or database corruption when destroying Terraform-managed infrastructure. It's lower priority than P1 because it affects user-initiated operations (which can be retried) rather than automated cleanup.

**Independent Test**: Create Terraform configuration with interdependent resources (image → category → device). Run `terraform destroy`. Verify provider attempts deletion in correct order, provides actionable error messages for dependency violations, and successfully completes when dependencies are resolved.

**Acceptance Scenarios**:

1. **Given** a Terraform state with a category referencing a software image, **When** user runs terraform destroy on the image, **Then** provider detects the dependency and returns error message: "Cannot delete software image 'X': still referenced by categories: [Y]"
2. **Given** a Terraform state with device → category → image dependencies, **When** user runs terraform destroy, **Then** provider deletes resources in correct order (device, then category, then image) without manual intervention
3. **Given** a provider Delete method with force=true flag, **When** deletion is attempted, **Then** dependency validation is bypassed and deletion proceeds (for advanced cleanup scenarios)

---

### User Story 3 - Test Infrastructure Reliability (Priority: P3)

A provider developer runs acceptance tests with CheckDestroy functions that validate proper resource cleanup. The test infrastructure respects deletion order when verifying resource destruction, preventing false test failures and ensuring reliable test cleanup.

**Why this priority**: This improves developer experience and test reliability but doesn't directly prevent production issues. It's essential for maintaining test suite quality but lower priority than preventing actual database corruption.

**Independent Test**: Run acceptance test suite with CheckDestroy functions. Verify all tests pass, CheckDestroy validates deletion in correct order, and no orphaned resources remain after test completion.

**Acceptance Scenarios**:

1. **Given** an acceptance test creating interdependent resources, **When** CheckDestroy executes, **Then** resources are verified for deletion in dependency order (devices before images) and all checks pass
2. **Given** a CheckDestroy function with detailed logging, **When** verifying multi-resource cleanup, **Then** logs show the dependency-aware verification sequence
3. **Given** a test that creates circular dependencies (if possible), **When** CheckDestroy runs, **Then** clear error message indicates the circular dependency and suggests resolution

---

### Edge Cases

- **What happens when a user attempts to delete a software image referenced by 10+ categories?** System must efficiently query all dependencies and provide a complete list in the error message, not just the first violation found.
- **How does the system handle concurrent deletion attempts?** If multiple processes try to delete dependent resources simultaneously, the system must handle race conditions gracefully without corrupting the database.
- **What happens when BCM API is slow or unresponsive during dependency checks?** System must have reasonable timeouts and clear error messages, not hang indefinitely.
- **How does the system handle partially deleted dependency chains?** If deletion fails midway (network deleted but category deletion fails), system must provide clear recovery guidance.
- **What happens when force=true bypasses validation and corrupts the database?** Documentation must warn about force flag consequences and recommend backup/recovery procedures.
- **How does the system handle resources with no dependencies?** Independent resources (like Kubernetes clusters) should delete immediately without unnecessary dependency checks.

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Cleanup scripts MUST delete resources in dependency order: Devices → Kubernetes Clusters → Networks → Categories → Software Images
- **FR-002**: Provider Delete methods MUST query BCM API to check for dependent resources before attempting deletion (when force=false)
- **FR-003**: Provider Delete methods MUST return actionable error messages listing all blocking dependencies when deletion is prevented
- **FR-004**: Error messages MUST include resource names and types of dependencies, not just UUIDs
- **FR-005**: Provider Delete methods MUST support a force flag to bypass dependency validation for advanced cleanup scenarios
- **FR-006**: CheckDestroy functions MUST verify resource deletion in dependency order to prevent test false positives
- **FR-007**: System MUST provide helper functions for querying resource dependencies from BCM API
- **FR-008**: Documentation MUST include a dependency graph diagram showing all resource relationships
- **FR-009**: Documentation MUST provide deletion order best practices for each resource type
- **FR-010**: Cleanup scripts MUST log the deletion order and reason for the sequence
- **FR-011**: Provider Delete methods MUST handle eventual consistency (resource may appear to have dependencies briefly after dependent resource is deleted)
- **FR-012**: System MUST detect circular dependencies (if possible in BCM) and provide clear error messages

### Key Entities

- **Software Image (bcm_cmpart_softwareimage)**: Base dependency for categories and devices; must be deleted last; referenced by category.software_image_id field
- **Category (bcm_cmdevice_category)**: References software images and networks; referenced by devices; deletion blocked if devices exist with this category
- **Device (bcm_cmdevice_device)**: References categories; must be deleted first in the chain; no resources depend on devices
- **Network (bcm_cmnet_network)**: Referenced by categories via network configuration; deletion blocked if categories reference this network
- **Kubernetes Cluster (bcm_cmkube_cluster)**: Independent resource with no dependencies; can be deleted at any time

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All cleanup scripts execute without BCM database corruption warnings or API errors
- **SC-002**: 100% of cleanup script runs delete resources in the correct dependency order (verifiable via log analysis)
- **SC-003**: Provider Delete methods detect dependency violations before attempting deletion, preventing database corruption
- **SC-004**: Error messages for blocked deletions include complete list of blocking dependencies (no silent failures or incomplete information)
- **SC-005**: Acceptance test suite passes with 100% CheckDestroy success rate
- **SC-006**: Zero orphaned resource references in BCM database after cleanup operations complete
- **SC-007**: Documentation includes comprehensive dependency graph and deletion order guidance
- **SC-008**: Developers can successfully run cleanup operations without manual intervention or trial-and-error

## Assumptions

- **A-001**: BCM API provides methods to query resource dependencies (e.g., list all categories referencing a specific software image)
- **A-002**: BCM API returns dependency violation errors with sufficient detail to identify blocking resources
- **A-003**: The dependency graph is acyclic (no circular dependencies possible in BCM data model)
- **A-004**: Force flag usage is reserved for administrative cleanup scenarios with database backup available
- **A-005**: Cleanup scripts run with sufficient privileges to query and delete all resource types
- **A-006**: Test infrastructure has exclusive access to test resources (no concurrent modifications from other processes)

## Dependencies

- **D-001**: BCM API must expose methods to query reverse dependencies (e.g., "which categories reference this image?")
- **D-002**: Existing provider resource implementations (Delete methods to be enhanced)
- **D-003**: Test helper infrastructure (createTestBCMClient, verifyResourceDeleted functions)
- **D-004**: Cleanup script framework and logging infrastructure

## Scope

### In Scope

- Fix cleanup-basic-resources.sh deletion order
- Verify and fix cleanup-before-tests.sh deletion order
- Add dependency validation to all provider Delete methods
- Implement helper functions for querying resource dependencies
- Enhance error messages for dependency violations
- Update CheckDestroy functions to respect deletion order
- Create comprehensive dependency graph documentation
- Add deletion order logging to cleanup scripts

### Out of Scope

- Automatic dependency resolution (provider will not automatically reorder deletions in terraform destroy)
- Transitive dependency tracking beyond immediate relationships
- Database-level referential integrity enforcement (BCM API limitation)
- Historical dependency tracking or audit logging
- Dependency visualization tooling
- Cross-provider dependency management (e.g., dependencies with other Terraform providers)

## Open Questions

None at this time. All requirements are sufficiently specified based on the provided dependency graph and current implementation analysis.

## Related Work

- **Issue**: Database corruption from incorrect cleanup order (identified in cleanup-basic-resources.sh)
- **Files to modify**:
  - scripts/cleanup-basic-resources.sh (line order fix)
  - scripts/cleanup-before-tests.sh (verification needed)
  - internal/provider/resource_cmpart_softwareimage.go:511 (Delete method)
  - internal/provider/resource_cmdevice_category.go:985 (Delete method)
  - internal/provider/resource_cmdevice_device.go:868 (Delete method)
  - internal/provider/resource_cmnet_network.go:395 (Delete method)
  - internal/provider/resource_cmkube_cluster.go:640 (Delete method - may not need changes)
  - All *_test.go files with CheckDestroy functions
- **Documentation to create**:
  - Dependency graph diagram
  - Deletion order best practices
  - Force flag usage warnings
