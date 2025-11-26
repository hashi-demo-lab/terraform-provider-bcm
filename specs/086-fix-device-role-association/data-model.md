# Data Model: Fix Device Role Association Bug

**Feature**: 086-fix-device-role-association
**Date**: 2025-11-26
**Status**: Complete

## Entity Overview

This feature modifies how roles are handled in the `bcm_cmdevice_device` resource. No new entities are introduced.

---

## Modified Entity: Device (bcm_cmdevice_device)

### Roles Attribute

**Current Schema**:
```go
"roles": schema.SetAttribute{
    Optional:    true,
    Computed:    true,
    ElementType: types.StringType,
    // Description mentions UUIDs
}
```

**Updated Schema** (no structural change, only documentation):
```go
"roles": schema.SetAttribute{
    Optional:    true,
    Computed:    true,
    ElementType: types.StringType,
    MarkdownDescription: "Set of role names assigned to this device. Roles define the device's function " +
        "in the cluster (e.g., \"backup\", \"provisioning\", \"boot\"). Use the `bcm_cmdevice_roles` " +
        "data source to discover available roles. Role names are case-sensitive. For backward " +
        "compatibility, role UUIDs are also accepted but role names are recommended for readability.",
}
```

### Input/Output Model

| Field | Input (Config) | Output (State) | API (BCM) |
|-------|----------------|----------------|-----------|
| roles | Set of strings (names or UUIDs) | Set of strings (names only) | Array of role objects |

### Input Validation Rules

1. **Empty String Check**: Role identifiers cannot be empty strings
2. **Uniqueness**: Duplicate role identifiers are deduplicated (Set behavior)
3. **Existence Check**: All role identifiers must exist in the BCM cluster
4. **Format Detection**: UUID format triggers UUID lookup; otherwise name lookup

### State Transitions

```
User Config           ->  Provider Logic          ->  Terraform State
-----------               --------------              ---------------
roles = ["backup"]    ->  Resolve "backup" to      -> roles = ["backup"]
                          full role object

roles = ["<uuid>"]    ->  Resolve UUID to          -> roles = ["backup"]
                          role object, extract
                          name for state

roles = []            ->  Empty role array          -> roles = []
                          (explicit removal)

roles = null          ->  Preserve existing         -> roles = [current]
(omitted)                 roles (no change)
```

---

## Data Flow

### Create/Update Flow

```
┌─────────────────┐
│ User Config     │
│ roles=["backup"]│
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Parse Config    │
│ Extract strings │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│ For each role identifier:       │
│ - isUUID(id)? → Lookup by UUID  │
│ - else       → Lookup by Name   │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Query BCM: getNodes             │
│ Build rolesByName & rolesByUUID │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Validate all identifiers exist  │
│ Return error if any missing     │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Build role objects array        │
│ (full BCM role structure)       │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Add to device entity            │
│ Call BCM addDevice/updateDevice │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────┐
│ Save State      │
│ roles=["backup"]│ (names only)
└─────────────────┘
```

### Read Flow

```
┌─────────────────┐
│ BCM API         │
│ getDevice(uuid) │
└────────┬────────┘
         │
         ▼
┌─────────────────────────────────┐
│ Parse response                  │
│ Extract roles array             │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────────────────────┐
│ For each role object:           │
│ Extract role["name"]            │
└────────┬────────────────────────┘
         │
         ▼
┌─────────────────┐
│ Return State    │
│ roles=["backup"]│
└─────────────────┘
```

---

## BCM API Role Object Structure

### Full Role Object (from getNodes)

```json
{
  "uuid": "12345678-1234-1234-1234-123456789abc",
  "name": "backup",
  "baseType": "Role",
  "childType": "BackupRole",
  "addServices": true,
  "revision": "",
  "modified": false,
  "to_be_removed": false
}
```

### Role Object for Assignment (sent to BCM)

```json
{
  "uuid": "12345678-1234-1234-1234-123456789abc",
  "name": "backup",
  "baseType": "Role",
  "childType": "BackupRole",
  "addServices": true
}
```

---

## Lookup Maps

### rolesByName Map

```go
type rolesByName map[string]map[string]interface{}

// Example:
{
    "backup": {
        "uuid": "12345678-...",
        "name": "backup",
        "baseType": "Role",
        "childType": "BackupRole",
        "addServices": true,
    },
    "provisioning": {
        "uuid": "87654321-...",
        "name": "provisioning",
        ...
    },
}
```

### rolesByUUID Map

```go
type rolesByUUID map[string]map[string]interface{}

// Example:
{
    "12345678-1234-1234-1234-123456789abc": {
        "uuid": "12345678-...",
        "name": "backup",
        ...
    },
    "87654321-4321-4321-4321-cba987654321": {
        "uuid": "87654321-...",
        "name": "provisioning",
        ...
    },
}
```

---

## Error Cases

### Invalid Role Name

**Input**: `roles = ["nonexistent-role"]`

**Error**:
```
Role 'nonexistent-role' does not exist in the BCM cluster.
Available roles: backup, boot, provisioning
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

### Invalid Role UUID

**Input**: `roles = ["99999999-9999-9999-9999-999999999999"]`

**Error**:
```
Role '99999999-9999-9999-9999-999999999999' does not exist in the BCM cluster.
Available roles: backup, boot, provisioning
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

### Empty String

**Input**: `roles = [""]`

**Error**:
```
Invalid role identifiers found: (empty string) - role identifiers must be non-empty strings
```

### Multiple Invalid Roles

**Input**: `roles = ["backup", "invalid1", "invalid2"]`

**Error**:
```
Roles not found in cluster: invalid1, invalid2
Available roles: backup, boot, provisioning
Use the `bcm_cmdevice_roles` data source to discover available roles.
```

---

## Backward Compatibility Matrix

| Scenario | Input | Behavior | State Output |
|----------|-------|----------|--------------|
| Name only | `["backup"]` | Lookup by name | `["backup"]` |
| UUID only | `["12345678-..."]` | Lookup by UUID | `["backup"]` |
| Mixed | `["backup", "12345678-..."]` | Both lookups | `["backup", "provisioning"]` |
| Legacy config | `[local.backup_role]` | UUID resolved | `["backup"]` |

---

## Testing Considerations

### State Comparison

Since state now stores names instead of UUIDs, existing configurations using UUIDs will see a "change" on first apply:

```diff
  roles = [
-   "12345678-1234-1234-1234-123456789abc",
+   "backup",
  ]
```

This is expected and acceptable. The roles themselves are unchanged in BCM.

### Import Behavior

When importing a device with roles:
1. Device is read from BCM API
2. Role objects are parsed
3. Role **names** are extracted and stored in state
4. User can then manage roles by name in their configuration
