# Feature Specification: BCM Category List Fields Persistence Investigation

**Feature Branch**: `073-category-list-fields`
**Created**: 2025-11-27
**Status**: Draft
**Input**: User description: "Investigate BCM API category list fields persistence issue"
**GitHub Issue**: #73

## Executive Summary

BCM API accepts certain list fields during category creation/update but returns empty arrays on subsequent reads. This specification documents the investigation findings and proposes a resolution strategy.

### Affected Fields

| Terraform Field | BCM API Field | Type | Persistence Status |
|-----------------|---------------|------|-------------------|
| `static_routes` | `staticRoutes` | `[]StaticRoute` | Not persisted |
| `fsexports` | `fsexports` | `[]FSExport` | Not persisted |
| `roles` | `roles` | `[]Role` | Not persisted |
| `gpu_settings` | `gpuSettings` | `[]GPUSetting` | Not persisted |
| `services` | `services` | `[]Service` | Not persisted |

### Investigation Findings

**Evidence from Codebase**:

1. **Current Workaround in Provider** (`resource_cmdevice_category.go`):
   - Lines 1051-1054: "Restore plan values for optional list fields that BCM doesn't persist"
   - Lines 1257-1260: "BCM API doesn't persist these fields for categories - preserve user's configured values"
   - Lines 1509-1512: Similar pattern for Update operations
   - Lines 2736-2748: "BCM doesn't persist category roles" - generates UUIDs locally

2. **Test Comments Confirm Behavior** (`resource_cmdevice_category_test.go`):
   - Line 2590: "Note: BCM doesn't persist static_routes, fsexports, roles, gpu_settings, services"
   - Line 2785: Same note repeated
   - Line 3599: "BCM does not persist fsmounts after category creation"
   - Line 3696: "BCM may not persist fsexports after category creation"

3. **BCM API Schema Documentation** (`category_schema_documentation_20251126_195346.md`):
   - All affected fields are present in API response schema
   - All return type `list` with value `[]` (empty arrays)
   - BCM accepts these fields but does not store them

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Verify API Behavior (Priority: P1)

As a Terraform provider developer, I need to verify the actual BCM API behavior for category list fields so that I can determine whether this is a BCM limitation or a provider bug.

**Why this priority**: Without verified evidence of BCM API behavior, we cannot determine the correct fix. Current code assumes BCM doesn't persist these fields, but this assumption may be wrong.

**Independent Test**: Can be fully tested by creating a Python script that creates a category with list field values via BCM API directly (bypassing Terraform), then reading the category back to see what persists.

**Acceptance Scenarios**:

1. **Given** BCM API is accessible, **When** I create a category with `staticRoutes` populated via direct API call, **Then** I observe whether the routes persist on subsequent `getCategory` calls
2. **Given** BCM API is accessible, **When** I create a category with `roles` populated via direct API call, **Then** I observe whether the roles persist on subsequent `getCategory` calls
3. **Given** BCM API is accessible, **When** I update an existing category to add `fsexports` via direct API call, **Then** I observe whether the exports persist on subsequent `getCategory` calls

---

### User Story 2 - Document BCM Limitations (Priority: P2)

As a Terraform user, I need clear documentation about which category fields BCM does not persist so that I understand the expected behavior and can plan accordingly.

**Why this priority**: If BCM truly doesn't persist these fields, users need to understand this is a BCM limitation, not a provider bug.

**Independent Test**: Can be verified by checking documentation completeness in the `docs/resources/cmdevice_category.md` file.

**Acceptance Scenarios**:

1. **Given** BCM does not persist certain list fields, **When** a user reads the resource documentation, **Then** they see a clear warning about which fields are not persisted by BCM
2. **Given** documentation exists, **When** a user configures `static_routes` in their Terraform config, **Then** the provider explains that these values are preserved in Terraform state but not in BCM

---

### User Story 3 - Investigate Alternative APIs (Priority: P3)

As a provider developer, I need to investigate whether BCM provides alternative APIs for setting these list fields (e.g., separate `addRole` or `addStaticRoute` methods) so that we can properly persist user configurations.

**Why this priority**: Some BCM services have separate APIs for managing sub-resources. If such APIs exist for categories, we should use them.

**Independent Test**: Can be tested by API discovery using existing investigation scripts in `sampleRest/`.

**Acceptance Scenarios**:

1. **Given** BCM API exists, **When** I probe for methods like `addCategoryRole`, `addCategoryStaticRoute`, **Then** I document whether alternative APIs exist
2. **Given** alternative APIs may exist, **When** I analyze the BCM web UI behavior (via network inspector), **Then** I can see which API calls it makes to persist these settings

---

### Edge Cases

- What happens when a user imports a category that has list fields configured in BCM UI? (Answer: Fields will be empty in Terraform state)
- How does the provider handle drift detection when BCM returns empty arrays but user configured values? (Answer: Currently preserves plan values to avoid false drift)
- What if BCM version differences affect persistence behavior? (Answer: Document and test across BCM versions if possible)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: Provider MUST accurately reflect BCM API behavior for category list fields
- **FR-002**: Provider MUST not show false drift when BCM returns empty arrays for user-configured list fields
- **FR-003**: Provider documentation MUST clearly indicate which fields BCM does not persist
- **FR-004**: Provider MUST include investigation script to verify BCM API behavior for these fields
- **FR-005**: Provider MUST update `ImportStateVerifyIgnore` for fields BCM doesn't persist
- **FR-006**: Provider tests MUST validate that workarounds function correctly (idempotency checks pass)

### Key Entities

- **Category**: BCM device category containing configuration for nodes in that category. Key attributes: name, uuid, managementNetwork, softwareImageProxy
- **StaticRoute**: Network route with destination CIDR, gateway IP, and optional metric
- **FSExport**: NFS export configuration with path, network reference, and access options
- **Role**: Service role assignment with name, type (childType), and addServices flag
- **GPUSetting**: GPU configuration with deviceId, model, and computeMode
- **Service**: Service configuration (structure TBD - marked as POST-MVP in current code)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Investigation script successfully runs against BCM API and produces documented results
- **SC-002**: All existing acceptance tests for affected fields pass with proper idempotency
- **SC-003**: Provider documentation includes clear warnings about BCM persistence limitations for all 5 affected fields
- **SC-004**: Import tests correctly ignore non-persisted fields (no false verification failures)
- **SC-005**: Provider log output includes debug messages when preserving user-configured values over BCM empty responses

## Assumptions

1. **BCM API Version**: Testing against BCM version available at endpoint 172.21.15.254:8081
2. **Fields are Schema-Valid**: BCM schema includes these fields (confirmed in documentation), even if not persisted
3. **No Alternative APIs**: Initial assumption that no separate APIs exist for these sub-resources (to be verified)
4. **Current Workaround is Correct**: The "preserve plan values" approach is the right solution if BCM truly doesn't persist

## Dependencies

- BCM API access for verification testing
- Existing investigation scripts in `sampleRest/` directory
- Current provider implementation as reference

## Scope Boundaries

### In Scope

- Verify actual BCM API behavior through direct API testing
- Document findings with evidence
- Update provider documentation if needed
- Ensure tests properly handle non-persisted fields

### Out of Scope

- Implementing alternative API calls (unless discovered during investigation)
- Modifying BCM behavior (not possible - external system)
- Adding persistence for `services` field (marked POST-MVP in current schema)

## Investigation Plan

### Phase 1: API Verification

Create and run investigation script to test each affected field:

```
Test Workflow:
1. Create category with staticRoutes populated
2. Read category back - check if staticRoutes persisted
3. Update category to add more routes
4. Read again - verify persistence/non-persistence

Repeat for: roles, fsexports, gpuSettings, services
```

### Phase 2: Alternative API Discovery

Probe for methods:
- `addCategoryRole` / `removeCategoryRole`
- `addCategoryStaticRoute` / `removeCategoryStaticRoute`
- `setCategoryFSExports`
- etc.

### Phase 3: Documentation Update

Based on findings, update:
- `docs/resources/cmdevice_category.md` - Add BCM limitation notices
- `CLAUDE.md` - Add notes about category list field behavior
- Test file comments - Ensure accuracy of existing comments

### Phase 4: Code Validation

Verify current workarounds:
- Confirm idempotency tests pass
- Verify import tests have correct ignore lists
- Check debug logging is present
