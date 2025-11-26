# BCM Category List Fields - Research Findings

**Status**: PENDING INVESTIGATION
**Last Updated**: 2025-11-27
**Investigator**: Speckit Plan Phase 0

## Executive Summary

This document will contain the findings from direct BCM API testing to verify which category list fields persist after create/update operations. The investigation will test 5 fields: `staticRoutes`, `fsexports`, `roles`, `gpuSettings`, and `services`.

**Current Status**: Waiting for investigation script execution.

## Test Environment

- **BCM Endpoint**: https://172.21.15.254:8081
- **BCM Version**: [To be captured during investigation]
- **Test Date**: [To be filled when script runs]
- **Test Script**: `sampleRest/investigate_category_list_fields.py`

## Field-by-Field Analysis

### static_routes

- **Terraform Attribute**: `static_routes`
- **BCM API Field**: `staticRoutes`
- **Expected Structure**: `[{destination, gateway, metric}]`
- **Persistence Status**: PENDING VERIFICATION
- **Evidence**: [To be captured]
- **Alternative API**: [To be tested]

**Provider Code Reference** (resource_cmdevice_category.go):
- Line 1054: `plan.StaticRoutes = planStaticRoutes` - preserves plan values

### fsexports

- **Terraform Attribute**: `fsexports`
- **BCM API Field**: `fsexports`
- **Expected Structure**: `[{path, network, allowWrite, async, rootSquash}]`
- **Persistence Status**: PENDING VERIFICATION
- **Evidence**: [To be captured]
- **Alternative API**: [To be tested]

**Provider Code Reference**:
- Test comment line 3696: "BCM may not persist fsexports after category creation"

### roles

- **Terraform Attribute**: `roles`
- **BCM API Field**: `roles`
- **Expected Structure**: `[{name, childType, addServices, uuid}]`
- **Persistence Status**: PENDING VERIFICATION
- **Evidence**: [To be captured]
- **Alternative API**: [To be tested]

**Provider Code Reference**:
- Line 2736: "Role not found in API response - BCM doesn't persist category roles"
- Line 2748: "Generated UUID for role (BCM doesn't persist category roles)"

### gpu_settings

- **Terraform Attribute**: `gpu_settings`
- **BCM API Field**: `gpuSettings`
- **Expected Structure**: `[{deviceId, model, computeMode}]`
- **Persistence Status**: PENDING VERIFICATION
- **Evidence**: [To be captured]
- **Alternative API**: [To be tested]

**Provider Code Reference**:
- Listed in ImportStateVerifyIgnore at lines 2601, 2797

### services

- **Terraform Attribute**: `services`
- **BCM API Field**: `services`
- **Expected Structure**: TBD (marked as POST-MVP in current schema)
- **Persistence Status**: PENDING VERIFICATION
- **Evidence**: [To be captured]
- **Alternative API**: [To be tested]

**Provider Code Reference**:
- Listed in ImportStateVerifyIgnore at lines 2602, 2798

---

## Existing Evidence from Codebase

### Provider Implementation Evidence

1. **Workaround Code** (resource_cmdevice_category.go:1051-1054):
   ```go
   // Restore plan values for optional list fields that BCM doesn't persist
   // BCM returns empty arrays for these fields, so we preserve what the user configured
   // This ensures Terraform state matches plan when BCM doesn't store these values
   plan.StaticRoutes = planStaticRoutes
   ```

2. **Role UUID Generation** (resource_cmdevice_category.go:2736-2748):
   - Provider generates UUIDs locally for roles since BCM doesn't return them
   - Debug logging confirms this behavior

3. **FSMount Handling** (resource_cmdevice_category.go:2929):
   ```go
   // Mount not found in API response - BCM doesn't persist category fsmounts
   ```

### Test Evidence

1. **ImportStateVerifyIgnore Lists**:
   - Multiple tests include these fields in ignore list
   - Lines 2596-2603, 2791-2798 explicitly list all 5 fields

2. **Test Comments**:
   - Line 2590: "Note: BCM doesn't persist static_routes, fsexports, roles, gpu_settings, services"
   - Line 3599: "BCM does not persist fsmounts after category creation"
   - Line 3772: "Note: BCM doesn't persist fsexports"
   - Line 3875: "Create category without roles (BCM doesn't persist roles)"

---

## Alternative API Discovery Results

**Status**: PENDING

| Method | Service | Status | Response |
|--------|---------|--------|----------|
| addCategoryRole | cmdevice | PENDING | - |
| removeCategoryRole | cmdevice | PENDING | - |
| setCategoryRoles | cmdevice | PENDING | - |
| addCategoryStaticRoute | cmdevice | PENDING | - |
| removeCategoryStaticRoute | cmdevice | PENDING | - |
| setCategoryStaticRoutes | cmdevice | PENDING | - |
| addCategoryFSExport | cmdevice | PENDING | - |
| setCategoryFSExports | cmdevice | PENDING | - |
| addCategoryGPUSetting | cmdevice | PENDING | - |
| setCategoryServices | cmdevice | PENDING | - |
| addNodeRole | cmdevice | PENDING | - |
| removeNodeRole | cmdevice | PENDING | - |
| getNodeRoles | cmdevice | PENDING | - |

---

## Preliminary Conclusions (Based on Codebase Analysis)

1. **High Confidence**: BCM does NOT persist these category list fields based on:
   - Explicit code comments in provider implementation
   - Workaround code that preserves plan values
   - Test comments confirming behavior
   - ImportStateVerifyIgnore configuration

2. **Medium Confidence**: No alternative APIs exist based on:
   - Existing `explore_roles_api.py` script found no working role methods
   - Provider implementation handles all cases locally

3. **Needs Verification**: Direct API testing will confirm or refute these assumptions

---

## Recommendations (Preliminary)

1. **Run Investigation Script**: Execute `investigate_category_list_fields.py` to capture direct evidence
2. **Document BCM Limitation**: Add warnings to resource documentation
3. **Validate Workarounds**: Ensure all idempotency tests pass
4. **Update Import Guidance**: Document that imported categories need re-apply for these fields

---

## Next Steps

1. Create investigation script: `/workspace/sampleRest/investigate_category_list_fields.py`
2. Execute script against BCM API
3. Capture JSON evidence in `/workspace/specs/073-category-list-fields/evidence/`
4. Update this document with actual findings
5. Finalize recommendations based on evidence
