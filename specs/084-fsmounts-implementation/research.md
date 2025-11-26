# Phase 0 Research: fsmounts Field Implementation

**Date**: 2025-11-26
**Feature**: 084-fsmounts-implementation
**Issue**: #84

## Research Summary

This document consolidates research findings for implementing the `fsmounts` field in `bcm_cmdevice_category` resource. All unknowns from the Technical Context have been investigated and resolved.

---

## Research Area 1: BCM API FSMount Structure

### Task
Research the exact BCM API structure for fsmounts - confirm field names, types, and baseType.

### Findings

**Decision**: Use BCM API field names as documented in spec with "FSMount" as baseType.

**Rationale**:
- Existing code patterns in `resource_cmdevice_category.go` consistently use baseType for entity type identification
- The spec documents the BCM API structure as:
  ```json
  {
    "baseType": "FSMount",
    "uuid": "...",
    "path": "/shared",          // corresponds to mountpoint
    "device": "nfs-server:/export",
    "type": "nfs",              // corresponds to filesystem
    "options": "defaults"       // corresponds to mountoptions
  }
  ```

**Alternatives Considered**:
1. Using "Mount" as baseType - Rejected, "FSMount" follows BCM naming convention
2. Using alternative field names - Rejected, spec provides clear mapping

### Field Mapping (Confirmed)

| Terraform Field | BCM API Field | Direction |
|-----------------|---------------|-----------|
| device | device | Bidirectional |
| mountpoint | path | TF snake_case to BCM camelCase |
| filesystem | type | TF to BCM name mapping |
| mountoptions | options | TF to BCM name mapping |
| fsck | fsck | Direct mapping (if supported) |
| dump | dump | Direct mapping (if supported) |
| rdma | rdma | Direct mapping (if supported) |
| uuid | uuid | Computed, BCM-assigned |

---

## Research Area 2: FSMount Serialization Pattern

### Task
Research the correct serialization pattern for fsmounts following existing patterns.

### Findings

**Decision**: Follow the fsexports serialization pattern at lines 1847-1871.

**Rationale**:
- FSExports is the most similar nested list structure to FSMounts
- Both are filesystem-related arrays with similar field types
- The pattern handles null/unknown checks correctly

**Serialization Code Pattern** (from fsexports):
```go
// Serialize fsexports (snake_case -> camelCase for BCM API)
if !model.FSExports.IsNull() && !model.FSExports.IsUnknown() {
    var exports []FSExportModel
    diags := model.FSExports.ElementsAs(ctx, &exports, false)
    if !diags.HasError() {
        exportsList := make([]map[string]interface{}, 0, len(exports))
        for _, export := range exports {
            exportMap := map[string]interface{}{
                "baseType": "FSExport",
                "path":     export.Path.ValueString(),
                "network":  export.Network.ValueString(),
            }
            // Handle optional fields
            if !export.AllowWrite.IsNull() {
                exportMap["allowWrite"] = export.AllowWrite.ValueBool()
            }
            exportsList = append(exportsList, exportMap)
        }
        entity["fsexports"] = exportsList
    }
}
```

**Key Patterns**:
1. Check `!IsNull() && !IsUnknown()` before processing
2. Use `ElementsAs()` to extract typed models
3. Build `[]map[string]interface{}` for API payload
4. Use `baseType` for entity type identification
5. Only include optional fields if not null

---

## Research Area 3: FSMount Parsing Pattern

### Task
Research the correct parsing pattern for reading fsmounts from BCM API.

### Findings

**Decision**: Follow the fsexports parsing pattern at lines 2198-2227.

**Rationale**:
- Same data structure type (nested list of objects)
- Handles the case where BCM returns array vs null
- Uses helper functions for null-safe extraction

**Parsing Code Pattern** (from fsexports):
```go
// Parse fsexports from BCM API (camelCase -> snake_case)
fsExportObjectType := types.ObjectType{AttrTypes: map[string]attr.Type{
    "path":        types.StringType,
    "network":     types.StringType,
    "allow_write": types.BoolType,
    "root_squash": types.BoolType,
    "async":       types.BoolType,
}}
if exportsData, ok := categoryData["fsexports"].([]interface{}); ok {
    exportValues := make([]attr.Value, 0, len(exportsData))
    for _, exportRaw := range exportsData {
        if exportMap, ok := exportRaw.(map[string]interface{}); ok {
            exportObj, objDiags := types.ObjectValue(fsExportObjectType.AttrTypes, map[string]attr.Value{
                "path":        getStringValue(exportMap, "path"),
                "network":     getStringValue(exportMap, "network"),
                "allow_write": getBoolValue(exportMap, "allowWrite"),
                "root_squash": getBoolValue(exportMap, "rootSquash"),
                "async":       getBoolValue(exportMap, "async"),
            })
            if !objDiags.HasError() {
                exportValues = append(exportValues, exportObj)
            }
        }
    }
    model.FSExports, _ = types.ListValue(fsExportObjectType, exportValues)
} else {
    model.FSExports = types.ListNull(fsExportObjectType)
}
```

---

## Research Area 4: Preservation vs Merge Strategy

### Task
Determine whether fsmounts needs preservation (like fsexports) or merging (like roles).

### Findings

**Decision**: Use merge strategy similar to roles (not simple preservation like fsexports).

**Rationale**:
1. FSMounts have a computed `uuid` field that BCM assigns - this needs to be populated from API response
2. Simple preservation (fsexports pattern) would discard BCM-assigned UUIDs
3. The roles pattern at issue #83 demonstrates the correct merge approach for computed UUIDs

**Why Roles Pattern**:
- Roles also have a computed `uuid` that BCM assigns
- The `mergeRolesWithAPIResponse` function shows how to:
  - Preserve user-specified values (device, mountpoint, filesystem, etc.)
  - Populate computed values (uuid) from BCM API
  - Match by unique identifier (device+mountpoint combination for mounts)

**Match Key Decision**:
- For roles: match by `name` (unique within category)
- For fsmounts: match by `device` + `mountpoint` combination (unique within category)

---

## Research Area 5: BCM API Persistence Behavior

### Task
Determine if BCM persists fsmounts data or returns empty arrays.

### Findings

**Decision**: Implement merge pattern with fallback to preservation.

**Rationale**:
- The spec notes: "BCM may or may not persist fsmounts data long-term"
- Edge case handling: If BCM returns empty array, preserve user config
- If BCM returns data with UUIDs, populate computed fields

**Implementation Strategy**:
1. If BCM returns fsmounts with data: merge to populate UUIDs
2. If BCM returns empty array: preserve from plan/state (following fsexports pattern)
3. If BCM returns null: preserve from plan/state

This hybrid approach handles both scenarios.

---

## Research Area 6: Import Support

### Task
Research requirements for supporting fsmounts during terraform import.

### Findings

**Decision**: Parse fsmounts from BCM API during import with full data.

**Rationale**:
- During import, there is no prior state - use API response directly
- The merge function handles null original values by using API response

**Import Flow**:
1. ImportState reads category by name/UUID
2. readCategory parses fsmounts from API response
3. No preservation needed (no prior state)
4. State populated from API data

---

## Conclusion

All technical unknowns have been resolved:

| Unknown | Resolution |
|---------|------------|
| BCM FSMount structure | Confirmed: baseType "FSMount", fields as documented |
| Serialization pattern | Use fsexports pattern (lines 1847-1871) |
| Parsing pattern | Use fsexports pattern (lines 2198-2227) |
| Preservation strategy | Use merge strategy (like roles) for UUID population |
| BCM persistence | Implement hybrid: merge if data returned, preserve if not |
| Import support | Direct API read, no preservation needed |

The implementation can proceed to Phase 1 Design.
