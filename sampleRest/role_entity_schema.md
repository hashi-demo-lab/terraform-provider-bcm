# BCM Role Entity Schema

Extracted from BCM API documentation JavaScript bundle.

## Entity Hierarchy

```
Entity
  └── Role (base) - name only
        ├── GenericRole - adds services, configuration, environments, exclude lists
        ├── BackupRole
        ├── BaseNginxRole
        ├── BeeGFSClientRole
        ├── BeeGFSHelperRole
        ├── BeeGFSManagementRole
        ├── BeeGFSMetadataRole
        ├── BeeGFSStorageRole
        ├── BootRole
        ├── CapiRole
        ├── CloudGatewayRole
        ├── DIGITSRole
        ├── DirectorRole
        ├── DnsRole
        ├── DockerHostRole
        ├── EtcdHostRole
        ├── FSPartRole
        ├── FailoverRole
        ├── FirewallRole
        ├── HeadNodeRole - disableAutomaticExports
        ├── JupyterHubRole
        ├── KubeletRole
        ├── LSFRole
        ├── MQTTRole
        ├── MonitoringRole
        ├── PRSClientRole
        ├── PRSServerRole
        ├── PbsProRole
        ├── ProvisioningRole
        ├── ScaleServerRole
        ├── SlurmAccountingRole
        ├── SlurmRole - wlmCluster
        ├── SnmpTrapRole
        ├── StorageRole - NFS settings
        ├── SubnetManagerRole
        └── WlmSubmitRole
```

## Role (Base Type)

- **parent**: Entity
- **plural**: Roles

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes (unique) | Role name |

## GenericRole (Most Common Child Type)

- **parent**: Role

### Fields

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Role name (inherited from Role) |
| `services` | array[string] | Services managed by this role |
| `configuration` | array[GenericRoleConfiguration] | Configuration files |
| `extraEnvironment` | array[GenericRoleEnvironment] | Environment variables |
| `excludeListSnippets` | array[ExcludeListSnippet] | Exclude list snippets |
| `dataNode` | bool | If enabled, node won't do FULL install without confirmation |

### BCM API Example

```json
{
  "baseType": "Role",
  "childType": "GenericRole",
  "uuid": "...",
  "name": "custom-role",
  "services": ["nginx", "custom-daemon"],
  "configuration": [],
  "extraEnvironment": [],
  "excludeListSnippets": [],
  "dataNode": false
}
```

---

## GenericRoleConfiguration (Sub-Entity)

Configuration file management for roles.

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes (unique) | - | Configuration name |
| `createDirectory` | bool | No | true | Create directory if doesn't exist |
| `filename` | string | No | - | File path (regex: `^/.*$`) |
| `mask` | uint32 | No | 644 | File permission mask |
| `userName` | string | No | - | User ownership |
| `groupName` | string | No | - | Group ownership |
| `disabled` | bool | No | false | Disabled flag |
| `serviceActionOnWrite` | enum | No | RESTART | Action on file change |
| `serviceStopOnFailure` | bool | No | true | Stop services on write failure |

### ServiceActionOnWrite Enum
- `RESTART` (default)
- Other values TBD

---

## GenericRoleEnvironment (Sub-Entity)

Environment variables for roles.

### Fields

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes (unique) | Variable name (regex: `^[a-zA-Z0-9_]+$`) |
| `value` | string | No | Variable value |
| `nodeEnvironment` | bool | No | Update node environment variables (default: false) |

---

## ExcludeListSnippet (Sub-Entity)

Exclude list configuration for image sync operations.

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes (unique) | - | Snippet name |
| `excludeList` | array[string] | No | - | Excluded paths |
| `disabled` | bool | No | false | Disabled flag |
| `noNewFiles` | bool | No | false | No new files mode |
| `modeSync` | bool | No | true | Include in sync mode |
| `modeFull` | bool | No | false | Include in full mode |
| `modeUpdate` | bool | No | true | Include in update mode |
| `modeGrab` | bool | No | false | Include in grab mode |
| `modeGrabNew` | bool | No | false | Include in grab new mode |

---

## Specialized Roles

### HeadNodeRole

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Role name |
| `failoverId` | uint64 | Failover ID (hidden) |
| `disableAutomaticExports` | bool | Disable automatic filesystem exports |

### StorageRole

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `name` | string | - | Role name |
| `nfsThreads` | uint32 | 0 | Number of NFS threads (0 = don't change) |
| `disableNFS1` | bool | - | Disable NFS v1 |
| `disableNFS2` | bool | - | Disable NFS v2 |
| `disableNFS3` | bool | - | Disable NFS v3 |
| `disableNFS4` | bool | - | Disable NFS v4 |
| `nfs4grace` | uint32 | 0 | NFS4 grace period |
| `statdPort` | uint32 | 0 | Stat daemon port |
| `statdOutgoingPort` | uint32 | 0 | Stat daemon outgoing port |
| `mountdPort` | uint32 | 0 | Mount daemon port |
| `rquotadPort` | uint32 | 0 | Rquota daemon port |

### SlurmRole

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Role name |
| `wlmCluster` | uuid | WLM cluster UUID |

---

## Current Terraform Implementation (CategoryRoleModel)

```go
type CategoryRoleModel struct {
    Name        types.String `tfsdk:"name"`         // Required
    ChildType   types.String `tfsdk:"child_type"`   // Required (e.g., "HeadNodeRole")
    UUID        types.String `tfsdk:"uuid"`         // Computed
    AddServices types.Bool   `tfsdk:"add_services"` // Optional (custom field)
}
```

### Current Terraform Schema

```hcl
roles = [
  {
    name        = "head"
    child_type  = "HeadNodeRole"
    uuid        = "..."  # Computed
    add_services = true  # Optional
  }
]
```

---

## Gap Analysis

### Missing Fields (GenericRole)

The current implementation is missing:
1. `services` - array of service names
2. `configuration` - GenericRoleConfiguration sub-entity list
3. `extraEnvironment` - GenericRoleEnvironment sub-entity list
4. `excludeListSnippets` - ExcludeListSnippet sub-entity list
5. `dataNode` - boolean flag

### Missing Specialized Role Fields

For specialized roles like StorageRole, HeadNodeRole, SlurmRole:
- Each has unique fields that aren't captured
- Would need polymorphic schema support or separate nested blocks

---

## Proposed Enhancement

### Option 1: Expand GenericRole Fields

Add GenericRole-specific fields to CategoryRoleModel:

```go
type CategoryRoleModel struct {
    // Base Role fields
    Name      types.String `tfsdk:"name"`
    ChildType types.String `tfsdk:"child_type"`
    UUID      types.String `tfsdk:"uuid"`

    // GenericRole fields
    Services              types.List `tfsdk:"services"`               // List of strings
    Configuration         types.List `tfsdk:"configuration"`          // List of GenericRoleConfiguration
    ExtraEnvironment      types.List `tfsdk:"extra_environment"`      // List of GenericRoleEnvironment
    ExcludeListSnippets   types.List `tfsdk:"exclude_list_snippets"`  // List of ExcludeListSnippet
    DataNode              types.Bool `tfsdk:"data_node"`

    // Specialized role fields (only used when child_type matches)
    // HeadNodeRole
    DisableAutomaticExports types.Bool `tfsdk:"disable_automatic_exports"`

    // StorageRole
    NfsThreads      types.Int64 `tfsdk:"nfs_threads"`
    DisableNFS1     types.Bool  `tfsdk:"disable_nfs1"`
    // ... etc
}
```

### Option 2: Separate Role Resources

Create separate Terraform resources for complex role management:
- `bcm_cmdevice_role` - CRUD for individual roles
- Keep `category.roles` as simple assignment list

---

## Notes

1. BCM API does NOT persist category roles - they're stored in Terraform state only
2. UUID is generated locally by the provider since BCM doesn't assign one
3. The `add_services` field in current implementation is custom (not in BCM API)
4. Role child types are polymorphic - different types have different fields
