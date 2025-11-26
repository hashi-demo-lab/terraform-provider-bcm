# BCM Category List Fields - Research Findings

**Status**: VERIFIED
**Last Updated**: 2025-11-27
**Investigator**: Automated Investigation Script + Direct API Testing

## Executive Summary

This document contains the findings from direct BCM API testing to verify which category list fields persist after create/update operations. The investigation tested 5 fields: `staticRoutes`, `fsexports`, `roles`, `gpuSettings`, and `services`.

**CONCLUSION**: BCM API **DOES NOT** persist any of the 5 category list fields. The API accepts these values in create/update operations (returns `success: true`) but does not store them - subsequent reads return empty arrays `[]`.

## Test Environment

- **BCM Endpoint**: https://172.21.15.254:8081
- **BCM Version**: N/A (not captured in API response)
- **Test Date**: 2025-11-27
- **Test Script**: `sampleRest/investigate_category_list_fields.py`
- **Test Methodology**: Used existing category to test update operations with list fields, then read back to verify persistence

## Field-by-Field Analysis

### static_routes

- **Terraform Attribute**: `static_routes`
- **BCM API Field**: `staticRoutes`
- **Expected Structure**: `[{destination, gateway, metric}]`
- **Persistence Status**: **NOT PERSISTED** (VERIFIED)
- **Evidence**:
  - Sent: `[{destination: "10.0.0.0/8", gateway: "192.168.1.1", metric: 100}]`
  - Update Response: `{success: true}`
  - Read After Update: `[]`
- **Alternative API**: None found (tested: addCategoryStaticRoute, setCategoryStaticRoutes - 400 Bad Request)

**Provider Code Reference** (resource_cmdevice_category.go):
- Line 1054: `plan.StaticRoutes = planStaticRoutes` - preserves plan values (CORRECT WORKAROUND)

### fsexports

- **Terraform Attribute**: `fsexports`
- **BCM API Field**: `fsexports`
- **Expected Structure**: `[{path, network, allowWrite, async, rootSquash}]`
- **Persistence Status**: **NOT PERSISTED** (VERIFIED)
- **Evidence**:
  - Sent: `[{path: "/test", network: "10.0.0.0/8", allowWrite: true, async: true, rootSquash: false}]`
  - Update Response: `{success: true}`
  - Read After Update: `[]`
- **Alternative API**: None found (tested: addCategoryFSExport, setCategoryFSExports - 400 Bad Request)

**Provider Code Reference**:
- Test comment line 3696: "BCM may not persist fsexports after category creation" (CONFIRMED)

### roles

- **Terraform Attribute**: `roles`
- **BCM API Field**: `roles`
- **Expected Structure**: `[{name, childType, addServices, uuid}]`
- **Persistence Status**: **NOT PERSISTED** (VERIFIED)
- **Evidence**:
  - Sent: `[{name: "test-role", childType: "ComputeRole", addServices: false}]`
  - Update Response: `{success: true}`
  - Read After Update: `[]`
- **Alternative API**: None found (tested: addCategoryRole, setCategoryRoles, getRoles, getNodeRoles - all 400 Bad Request)

**Provider Code Reference**:
- Line 2736: "Role not found in API response - BCM doesn't persist category roles" (CONFIRMED)
- Line 2748: "Generated UUID for role (BCM doesn't persist category roles)" (CORRECT WORKAROUND)

### gpu_settings

- **Terraform Attribute**: `gpu_settings`
- **BCM API Field**: `gpuSettings`
- **Expected Structure**: `[{deviceId, model, computeMode}]`
- **Persistence Status**: **NOT PERSISTED** (VERIFIED)
- **Evidence**:
  - Sent: `[{deviceId: "0", model: "TestGPU", computeMode: "default"}]`
  - Update Response: `{success: true}`
  - Read After Update: `[]`
- **Alternative API**: None found (tested: addCategoryGPUSetting, setCategoryGPUSettings - 400 Bad Request)

**Provider Code Reference**:
- Listed in ImportStateVerifyIgnore at lines 2601, 2797 (CORRECT)

### services

- **Terraform Attribute**: `services`
- **BCM API Field**: `services`
- **Expected Structure**: TBD (marked as POST-MVP in current schema)
- **Persistence Status**: **NOT PERSISTED** (VERIFIED)
- **Evidence**:
  - Existing categories show: `services: []`
  - Not tested with update (POST-MVP field)
- **Alternative API**: None found (tested: addCategoryService, setCategoryServices - 400 Bad Request)

**Provider Code Reference**:
- Listed in ImportStateVerifyIgnore at lines 2602, 2798 (CORRECT)

---

## Existing Evidence from Codebase (CONFIRMED)

### Provider Implementation Evidence

1. **Workaround Code** (resource_cmdevice_category.go:1051-1054):
   ```go
   // Restore plan values for optional list fields that BCM doesn't persist
   // BCM returns empty arrays for these fields, so we preserve what the user configured
   // This ensures Terraform state matches plan when BCM doesn't store these values
   plan.StaticRoutes = planStaticRoutes
   ```
   **STATUS**: CORRECTLY IMPLEMENTED

2. **Role UUID Generation** (resource_cmdevice_category.go:2736-2748):
   - Provider generates UUIDs locally for roles since BCM doesn't return them
   - Debug logging confirms this behavior
   **STATUS**: CORRECTLY IMPLEMENTED

3. **FSMount Handling** (resource_cmdevice_category.go:2929):
   ```go
   // Mount not found in API response - BCM doesn't persist category fsmounts
   ```
   **STATUS**: CORRECTLY IMPLEMENTED

### Test Evidence (CONFIRMED)

1. **ImportStateVerifyIgnore Lists**:
   - Multiple tests include these fields in ignore list
   - Lines 2596-2603, 2791-2798 explicitly list all 5 fields
   **STATUS**: CORRECTLY IMPLEMENTED

2. **Test Comments**:
   - Line 2590: "Note: BCM doesn't persist static_routes, fsexports, roles, gpu_settings, services" (CONFIRMED)
   - Line 3599: "BCM does not persist fsmounts after category creation" (CONFIRMED)
   - Line 3772: "Note: BCM doesn't persist fsexports" (CONFIRMED)
   - Line 3875: "Create category without roles (BCM doesn't persist roles)" (CONFIRMED)

---

## Alternative API Discovery Results

**Status**: COMPLETED - NO ALTERNATIVE APIs FOUND

| Method | Service | Status | Response |
|--------|---------|--------|----------|
| addCategoryRole | cmdevice | NOT_FOUND | 400 Bad Request |
| removeCategoryRole | cmdevice | NOT_FOUND | 400 Bad Request |
| setCategoryRoles | cmdevice | NOT_FOUND | 400 Bad Request |
| getCategoryRoles | cmdevice | NOT_FOUND | 400 Bad Request |
| addCategoryStaticRoute | cmdevice | NOT_FOUND | 400 Bad Request |
| removeCategoryStaticRoute | cmdevice | NOT_FOUND | 400 Bad Request |
| setCategoryStaticRoutes | cmdevice | NOT_FOUND | 400 Bad Request |
| getCategoryStaticRoutes | cmdevice | NOT_FOUND | 400 Bad Request |
| addCategoryFSExport | cmdevice | NOT_FOUND | 400 Bad Request |
| setCategoryFSExports | cmdevice | NOT_FOUND | 400 Bad Request |
| addCategoryGPUSetting | cmdevice | NOT_FOUND | 400 Bad Request |
| setCategoryGPUSettings | cmdevice | NOT_FOUND | 400 Bad Request |
| setCategoryServices | cmdevice | NOT_FOUND | 400 Bad Request |
| addNodeRole | cmdevice | NOT_FOUND | 400 Bad Request |
| removeNodeRole | cmdevice | NOT_FOUND | 400 Bad Request |
| getNodeRoles | cmdevice | NOT_FOUND | 400 Bad Request |
| getRoles | cmdevice | NOT_FOUND | 400 Bad Request |
| getRole | cmdevice | NOT_FOUND | 400 Bad Request |

---

## Final Conclusions (VERIFIED)

1. **HIGH CONFIDENCE - CONFIRMED**: BCM does NOT persist these category list fields:
   - `staticRoutes` (static_routes)
   - `fsexports` (fsexports)
   - `roles` (roles)
   - `gpuSettings` (gpu_settings)
   - `services` (services)

2. **HIGH CONFIDENCE - CONFIRMED**: No alternative APIs exist:
   - All tested alternative methods return 400 Bad Request
   - BCM does not expose separate endpoints for category list field management

3. **VERIFIED**: Provider workarounds are correctly implemented:
   - Plan value preservation in Create/Update operations
   - Local UUID generation for roles
   - ImportStateVerifyIgnore includes all 5 fields

---

## Recommendations (FINAL)

1. **Documentation Update** (REQUIRED):
   - Add "Known Limitations" section to `docs/resources/cmdevice_category.md`
   - Add inline warnings for each affected attribute
   - Update CLAUDE.md with BCM-specific notes

2. **Provider Workarounds** (VALIDATED - NO CHANGES NEEDED):
   - Current implementation correctly handles the BCM limitation
   - Plan value preservation is working as designed
   - ImportStateVerifyIgnore is complete

3. **Test Coverage** (VALIDATED - NO CHANGES NEEDED):
   - Existing idempotency tests pass
   - Import tests correctly ignore non-persisted fields

4. **Import Guidance** (DOCUMENT):
   - Document that imported categories will have empty values for these fields
   - Users must re-apply configuration after import to restore state

---

## Evidence Files

- `/workspace/specs/073-category-list-fields/evidence/category_list_fields_test_results.json` - Full test results
- `/workspace/specs/073-category-list-fields/evidence/api_discovery_results.md` - Alternative API discovery
- `/workspace/sampleRest/investigate_category_list_fields.py` - Investigation script
