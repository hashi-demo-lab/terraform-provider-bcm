# Implementation Plan: fsmounts Field Implementation

**Branch**: `084-fsmounts-implementation` | **Date**: 2025-11-26 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/084-fsmounts-implementation/spec.md`

## Summary

Implement the `fsmounts` field in `bcm_cmdevice_category` resource to serialize filesystem mount configurations to BCM API, parse API responses, and merge user config with BCM-computed UUIDs. Following the fsexports/roles pattern for nested list handling.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (state managed by Terraform)
**Testing**: TF_ACC=1 acceptance tests with BCM cluster
**Target Platform**: Terraform provider for BCM JSON-RPC API
**Project Type**: Single provider project
**Performance Goals**: N/A (standard provider operations)
**Constraints**: Must handle BCM API not persisting data (preservation pattern)
**Scale/Scope**: Single resource modification (bcm_cmdevice_category)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Gate | Status | Notes |
|------|--------|-------|
| TDD Required | ✅ PASS | Tests written first in RED phase |
| No Over-Engineering | ✅ PASS | Following existing fsexports/roles patterns |
| Parallel Execution | ✅ PASS | Tests run in parallel with other tests |
| Documentation Auto-Generated | ✅ PASS | make generate updates docs |

## Project Structure

### Documentation (this feature)

```text
specs/084-fsmounts-implementation/
├── spec.md              # Feature specification
├── plan.md              # This file
├── research.md          # Phase 0 research findings
├── data-model.md        # Data model documentation
├── quickstart.md        # Implementation quick start
├── contracts/           # API contracts
│   └── fsmounts-api.yaml
└── tasks.md             # Task breakdown (to be generated)
```

### Source Code (repository root)

```text
internal/provider/
├── resource_cmdevice_category.go      # Primary implementation file
├── resource_cmdevice_category_test.go # Acceptance tests
└── test_helpers.go                    # Shared test utilities
```

**Structure Decision**: Single file modification - all changes are in the existing resource_cmdevice_category.go file.

## Design Decisions

### D1: Merge Strategy for UUIDs

**Decision**: Use merge pattern (like roles) instead of simple preservation (like fsexports)

**Rationale**:
- FSMounts have computed `uuid` field that BCM assigns
- Simple preservation would discard BCM-assigned UUIDs
- Merge pattern preserves user config + populates computed values

**Match Key**: `device + mountpoint` combination (unique within category)

### D2: BCM API Field Mapping

**Decision**: Map Terraform snake_case to BCM API camelCase

| Terraform | BCM API |
|-----------|---------|
| mountpoint | path |
| filesystem | type |
| mountoptions | options |
| device | device |
| uuid | uuid |
| fsck | fsck |
| dump | dump |
| rdma | rdma |

**Rationale**: Matches existing field mapping patterns in codebase.

### D3: BCM Persistence Handling

**Decision**: Hybrid approach - merge if API returns data, preserve if not

**Rationale**:
- BCM may not persist fsmounts data long-term
- If API returns data: merge to populate UUIDs
- If API returns empty/null: preserve from plan/state

## Implementation Components

### C1: buildAPIEntity Serialization (Create/Update)

**Location**: `internal/provider/resource_cmdevice_category.go` ~line 2023
**Pattern**: Follow fsexports serialization (lines 1847-1871)

Key implementation:
- Check `!IsNull() && !IsUnknown()` before processing
- Use `ElementsAs()` to extract typed FSMountModel
- Build `[]map[string]interface{}` with baseType "FSMount"
- Map fields: mountpoint→path, filesystem→type, mountoptions→options

### C2: readCategory Parsing (Read)

**Location**: `internal/provider/resource_cmdevice_category.go` ~lines 2184-2196
**Pattern**: Follow fsexports parsing (lines 2198-2227)

Key implementation:
- Define fsMountObjectType with all attributes
- Check for `fsmounts` array in API response
- Use helper functions (getStringValue, getBoolValue) for null-safe extraction
- Map BCM fields to Terraform: path→mountpoint, type→filesystem, options→mountoptions

### C3: mergeFSMountsWithAPIResponse Function

**Location**: After mergeRolesWithAPIResponse (~line 2560)
**Pattern**: Follow mergeRolesWithAPIResponse pattern

Key implementation:
- Handle null/unknown cases
- Build lookup map by device+mountpoint
- Merge: preserve user config + populate UUID from API
- Generate UUID if not in API response (BCM didn't persist)

### C4: CRUD Function Modifications

**Create function** (~line 828, 1052):
- Save planFSMounts before readCategory
- Call mergeFSMountsWithAPIResponse after

**Read function** (~line 1077, 1195):
- Save originalFSMounts before readCategory
- Call mergeFSMountsWithAPIResponse after

**Update function** (~line 1321, 1390):
- Save planFSMounts before readCategory
- Call mergeFSMountsWithAPIResponse after

## Acceptance Test Plan

### T1: TestAccCMDeviceCategory_FSMountsBasic
- Create category with single fsmount
- Verify mount appears in state with UUID populated

### T2: TestAccCMDeviceCategory_FSMountsMultiple
- Create category with multiple fsmounts
- Verify all mounts appear in state

### T3: TestAccCMDeviceCategory_FSMountsUpdate
- Create category with one fsmount
- Add second fsmount
- Modify first fsmount options
- Remove second fsmount
- Verify each change is applied

### T4: TestAccCMDeviceCategory_FSMountsIdempotency
- Create category with fsmounts
- Run plan with no changes
- Verify empty plan (plancheck.ExpectEmptyPlan)

### T5: TestAccCMDeviceCategory_FSMountsImport
- Create category with fsmounts
- Import category
- Verify fsmounts appear in imported state

### T6: TestAccCMDeviceCategory_FSMountsDrift
- Create category with fsmounts
- Modify externally via BCM API
- Run plan, verify drift detected
- Apply, verify restored

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| BCM doesn't persist fsmounts | Medium | High | Preserve from plan/state if API returns empty |
| UUID generation conflicts | Low | Low | Use standard UUID generation function |
| Field mapping errors | Low | Medium | Follow documented mapping, add debug logging |

## Success Criteria Verification

| Criterion | Verification Method |
|-----------|---------------------|
| SC-001: Mounts in state | TestAccCMDeviceCategory_FSMountsBasic |
| SC-002: Idempotency | TestAccCMDeviceCategory_FSMountsIdempotency |
| SC-003: No regressions | All existing CMDeviceCategory tests pass |
| SC-004: New tests pass | All FSMounts tests pass |
| SC-005: NFS + local mounts | TestAccCMDeviceCategory_FSMountsMultiple |
