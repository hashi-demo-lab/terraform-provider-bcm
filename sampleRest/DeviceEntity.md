# BCM Device Entity API Documentation

## Overview

The Device entity is the core resource type in the NVIDIA BCM (Bright Cluster Manager) API. It represents physical or virtual compute nodes in the cluster.

## API Endpoint

**Base URL:** `https://172.21.15.254:8081/json`

**Method:** `POST`

**Authentication:** Session-based with `cm-login-token` cookie

## Authentication

```json
POST /json

{
  "service": "login",
  "username": "root",
  "password": "Hashicorp123!"
}
```

**Response:**
- Status: 200 OK
- Sets cookie: `cm-login-token=<token>`

## Device Entity Structure

### Base Properties

| Field | Type | Description |
|-------|------|-------------|
| `baseType` | string | Entity base type (always "Device") |
| `childType` | string | Specific device type (e.g., "HeadNode", "ComputeNode") |
| `uuid` | string | Unique identifier for the device |
| `hostname` | string | Device hostname |
| `mac` | string | Primary MAC address |
| `creationTime` | integer | Unix timestamp of device creation |
| `modified` | boolean | Whether device has unsaved modifications |
| `to_be_removed` | boolean | Scheduled for deletion |

### Network Configuration

| Field | Type | Description |
|-------|------|-------------|
| `interfaces` | array | Network interfaces (see NetworkInterface schema) |
| `defaultGateway` | string | Default gateway IP address |
| `defaultGatewayMetric` | integer | Gateway metric/priority |
| `staticRoutes` | array | Static routing table entries |

### Management Settings

| Field | Type | Description |
|-------|------|-------------|
| `cmdaemonUrl` | string | Cluster manager daemon URL |
| `authenticationService` | string | Authentication service type |
| `powerControl` | string | Power control method (e.g., "none", "ipmi") |
| `provisioningTransport` | string | Provisioning method (e.g., "RSYNCDAEMON") |

### Storage & Filesystems

| Field | Type | Description |
|-------|------|-------------|
| `fsexports` | array | NFS exports (see FSExport schema) |
| `fsmounts` | array | Filesystem mounts |

### Role Assignments

| Field | Type | Description |
|-------|------|-------------|
| `roles` | array | Assigned roles (see Role schema) |

Common role types:
- `HeadNodeRole` - Cluster management node
- `StorageRole` - NFS storage services
- `BackupRole` - Backup services
- `MonitoringRole` - Cluster monitoring
- `ProvisioningRole` - Node provisioning
- `BootRole` - Boot/PXE services

### Hardware Configuration

| Field | Type | Description |
|-------|------|-------------|
| `biosSetup` | object | BIOS configuration settings |
| `bmcSettings` | object | Baseboard Management Controller settings |
| `gpuSettings` | array | GPU configuration |
| `serialNumber` | string | Hardware serial number |
| `partNumber` | string | Hardware part number |

## API Calls

### Get Device by Name

```json
POST /json
Cookie: cm-login-token=<token>

{
  "service": "cmdevice",
  "call": "getNode",
  "arg": "master"
}
```

**Response:** Complete Device object (JSON)

### Get Device by UUID

```json
POST /json
Cookie: cm-login-token=<token>

{
  "service": "cmdevice",
  "call": "getNode",
  "arg": "<uuid>"
}
```

## Example Device Response (Master Node)

```json
{
  "baseType": "Device",
  "childType": "HeadNode",
  "hostname": "bcm11-headnode",
  "uuid": "9f885869-a146-4cd6-af1f-f9b6c674a84c",
  "mac": "00:50:56:9B:E4:6D",
  "cmdaemonUrl": "https://172.21.15.254:8081",
  "authenticationService": "CATEGORY",
  "powerControl": "none",
  "provisioningTransport": "RSYNCDAEMON",
  "interfaces": [
    {
      "baseType": "NetworkInterface",
      "childType": "NetworkPhysicalInterface",
      "name": "ens33",
      "mac": "00:50:56:9B:E4:6D",
      "ip": "172.21.15.254",
      "ipv6Ip": "::0",
      "dhcp": false,
      "cardtype": "Ethernet"
    }
  ],
  "roles": [
    {
      "baseType": "Role",
      "childType": "HeadNodeRole",
      "name": "headnode"
    },
    {
      "baseType": "Role",
      "childType": "StorageRole",
      "name": "storage"
    }
  ]
}
```

## Nested Entity Types

### NetworkInterface

Physical or virtual network interface configuration.

**Key fields:**
- `name` - Interface name (e.g., "ens33")
- `mac` - MAC address
- `ip` - IPv4 address
- `ipv6Ip` - IPv6 address
- `dhcp` - DHCP enabled
- `cardtype` - Interface type ("Ethernet", "InfiniBand", etc.)
- `network` - Network UUID reference

### FSExport

NFS export configuration.

**Key fields:**
- `path` - Export path
- `network` - Network UUID for export
- `allowWrite` - Write permission
- `rootSquash` - Root squash enabled
- `async` - Async I/O enabled

### Role

Service role assignment for the device.

**Key fields:**
- `name` - Role name
- `baseType` - Always "Role"
- `childType` - Specific role type
- `addServices` - Auto-add related services

## Python Client Example

See `capture_device_api.py` for a complete working example.

```python
from capture_device_api import BCMApiClient

# Initialize client
client = BCMApiClient(
    base_url="https://172.21.15.254:8081",
    username="root",
    password="Hashicorp123!"
)

# Login
client.login()

# Get device
device = client.get_device("master")
print(device["hostname"])  # bcm11-headnode
```

## Captured Responses

Full API responses are saved in timestamped JSON files:
- `device_master_YYYYMMDD_HHMMSS.json` - Master node details

## Notes

- All timestamps are Unix epoch (seconds since 1970-01-01)
- UUIDs follow RFC 4122 format
- The API uses SSL with self-signed certificates (disable verification in clients)
- Session cookies expire after inactivity
- Many nested objects share the `baseType`/`childType` pattern for polymorphism
