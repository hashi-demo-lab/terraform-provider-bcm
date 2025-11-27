# BCM Services and FSExports Entity Schema

Extracted from BCM API documentation JavaScript bundle.

## OSServiceConfig Entity

OS Service configuration for managing operating system services on nodes.

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes (unique) | - | Service name (max 64 chars) |
| `from` | string | No | - | Name of entity this service was created from (hidden) |
| `monitored` | bool | No | - | CMDaemon will periodically check if the service is running |
| `autostart` | bool | No | - | CMDaemon will restart a failed service |
| `runIf` | enum (RunCondition) | No | ALWAYS | Only run this service in the specified state |
| `managed` | bool | No | - | Manage config files from cmd (if any) |
| `sicknessCheckScript` | string | No | - | Script for sickness checking |
| `sicknessCheckScriptTimeout` | uint32 | No | 10 | Timeout after which the script is killed |
| `sicknessCheckInterval` | uint32 | No | 60 | Sickness checks interval (rounded up to 30s monitoring interval) |
| `scriptTimeout` | int32 | No | -1 | Service operation timeout |

### Hidden Fields (Internal Use)
- `addFromRole` (bool, default: true)
- `fromGenericRole` (bool, default: false)
- `ref_node_uuid` (uuid)
- `ref_role_uuid` (uuid)
- `ref_extra_uuid` (uuid)
- `internal` (bool, default: false)
- `serviceType` (uint32, default: 0)

### RunCondition Enum
- `ALWAYS` (default)
- Other values TBD

### BCM API Example

```json
{
  "baseType": "OSServiceConfig",
  "childType": "",
  "uuid": "...",
  "name": "nginx",
  "monitored": true,
  "autostart": true,
  "runIf": "ALWAYS",
  "managed": false,
  "sicknessCheckScript": "",
  "sicknessCheckScriptTimeout": 10,
  "sicknessCheckInterval": 60,
  "scriptTimeout": -1
}
```

### Terraform Schema Design

```hcl
services = [
  {
    name                        = "nginx"
    monitored                   = true
    autostart                   = true
    run_if                      = "ALWAYS"
    managed                     = false
    sickness_check_script       = "/usr/local/bin/check_nginx.sh"
    sickness_check_script_timeout = 10
    sickness_check_interval     = 60
    script_timeout              = 30
  }
]
```

---

## FSExport Entity

NFS filesystem export configuration for `/etc/exports`.

### Fields

| Field | Type | Required | Default | Description |
|-------|------|----------|---------|-------------|
| `name` | string | Yes (unique) | - | Export name (same as path for single exports) |
| `path` | string | No | - | Path to export |
| `network` | uuid | No | - | Network UUID the interface is connected to |
| `hosts` | string | No | - | Extra hosts-range allowed access (space separated) |
| `automatic` | bool | No (read-only) | false | Export was created automatically |
| `allowWrite` | bool | No | - | Allow writing |
| `async` | bool | No | true | Allow async NFS operations |
| `rootSquash` | bool | No | - | Map uid/gid 0 to anonymous |
| `allSquash` | bool | No | - | Map all uids/gids to anonymous |
| `anonUid` | uint32 | No | 65534 | Anonymous account user id |
| `anonGid` | uint32 | No | 65534 | Anonymous account group id |
| `extraOptions` | string | No | - | Extra options for export |
| `fsid` | uint32 | No | - | File system id for failover setup |
| `rdma` | bool | No | - | Enable NFS over RDMA |
| `disabled` | bool | No | - | Disable the export |
| `checkTree` | bool | No | - | Check tree |

### BCM API Example

```json
{
  "baseType": "FSExport",
  "childType": "",
  "uuid": "...",
  "name": "/export/home",
  "path": "/export/home",
  "network": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
  "hosts": "",
  "automatic": false,
  "allowWrite": true,
  "async": true,
  "rootSquash": true,
  "allSquash": false,
  "anonUid": 65534,
  "anonGid": 65534,
  "extraOptions": "",
  "fsid": 0,
  "rdma": false,
  "disabled": false,
  "checkTree": false
}
```

### Terraform Schema Design

```hcl
fsexports = [
  {
    name          = "/export/home"
    path          = "/export/home"
    network       = "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
    hosts         = "10.0.0.0/8"
    allow_write   = true
    async         = true
    root_squash   = true
    all_squash    = false
    anon_uid      = 65534
    anon_gid      = 65534
    extra_options = "no_subtree_check"
    rdma          = false
    disabled      = false
  }
]
```

---

## Notes

1. Both entities use BCM's standard entity structure with `baseType`/`childType`
2. UUID is BCM-assigned (computed)
3. `name` is the unique identifier for both entities
4. `FSExport.network` references a Network entity UUID
5. `OSServiceConfig` has many hidden fields for internal BCM use
