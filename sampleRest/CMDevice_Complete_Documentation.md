# CMDevice Service - Complete API Documentation

## Overview

The **CMDevice** service is the primary API for managing cluster devices (nodes) in NVIDIA Bright Cluster Manager (BCM). It provides comprehensive CRUD operations for physical nodes, head nodes, compute nodes, and cloud nodes.

## Service Information

- **Service Name**: `cmdevice` (lowercase in API calls)
- **Endpoint**: `POST https://172.21.15.254:8081/json`
- **Authentication**: Required (session cookie)
- **Content-Type**: `application/json`

## Discovered API Methods

### List & Query Operations

#### `getNodes`
Get all nodes in the cluster.

**Request:**
```json
{
  "service": "cmdevice",
  "call": "getNodes"
}
```

**Response:** Array of Device objects (all node types)

**Example Response:**
```json
[
  {
    "baseType": "Device",
    "childType": "PhysicalNode",
    "hostname": "node002",
    "uuid": "2870c0b0-6fda-4026-9b8f-28be4c372fee",
    "mac": "00:00:00:00:00:00",
    "interfaces": [...],
    "roles": [...],
    ...
  }
]
```

#### `getDevices`
Alias for `getNodes`. Returns all devices in the cluster.

**Request:**
```json
{
  "service": "cmdevice",
  "call": "getDevices"
}
```

**Response:** Array of Device objects (identical to getNodes)

#### `getComputeNodes`
Get only compute nodes.

**Request:**
```json
{
  "service": "cmdevice",
  "call": "getComputeNodes"
}
```

**Response:** Array of ComputeNode objects

**Usage:**
```python
compute_nodes = client.call_api("cmdevice", "getComputeNodes")
for node in compute_nodes:
    print(f"{node['hostname']}: {node['uuid']}")
```

#### `getCategories`
Get all node categories.

**Request:**
```json
{
  "service": "cmdevice",
  "call": "getCategories"
}
```

**Response:** Array of Category objects

**Example Response:**
```json
[
  {
    "baseType": "Category",
    "uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
    "name": "default",
    "description": "Default category"
  }
]
```

### Individual Node Operations

#### `getNode`
Get a specific node by hostname or UUID.

**Request:**
```json
{
  "service": "cmdevice",
  "call": "getNode",
  "arg": "master"
}
```

**Arguments:**
- Hostname (string): e.g., "master", "node001"
- UUID (string): e.g., "9f885869-a146-4cd6-af1f-f9b6c674a84c"

**Response:** Single Device object

**Python Example:**
```python
# Get by hostname
master_node = client.call_api("cmdevice", "getNode", "master")

# Get by UUID
node = client.call_api("cmdevice", "getNode", "9f885869-a146-4cd6-af1f-f9b6c674a84c")
```

### Power & Control Operations

#### `reboot`
Reboot a node.

**Request:**
```json
{
  "service": "cmdevice",
  "call": "reboot",
  "arg": "node001"
}
```

**Response:** Status object

**Note:** Additional power operations likely include:
- `powerOn`
- `powerOff`
- `powerCycle`
- `powerStatus`

## Device Entity Structure

### Base Device Object

All devices share these common fields:

| Field | Type | Description |
|-------|------|-------------|
| `baseType` | string | Always "Device" |
| `childType` | string | Specific device type |
| `uuid` | string | Unique identifier (RFC 4122) |
| `hostname` | string | Node hostname |
| `mac` | string | Primary MAC address |
| `creationTime` | integer | Unix timestamp |
| `modified` | boolean | Has unsaved changes |
| `to_be_removed` | boolean | Scheduled for deletion |

### Device Types (childType)

| Type | Description |
|------|-------------|
| `HeadNode` | Cluster management node |
| `PhysicalNode` | Physical compute node |
| `ComputeNode` | Compute worker node |
| `CloudNode` | Cloud-based node |
| `StorageNode` | Storage node |

### Network Configuration

#### interfaces
Array of NetworkInterface objects:

```json
{
  "baseType": "NetworkInterface",
  "childType": "NetworkPhysicalInterface",
  "name": "ens33",
  "mac": "00:50:56:9B:E4:6D",
  "ip": "172.21.15.254",
  "ipv6Ip": "::0",
  "dhcp": false,
  "network": "<network_uuid>",
  "cardtype": "Ethernet",
  "bootable": false,
  "startIf": "ALWAYS"
}
```

**Key Fields:**
- `name` - Interface name (ens33, eth0, etc.)
- `mac` - MAC address
- `ip` - IPv4 address
- `ipv6Ip` - IPv6 address
- `network` - UUID of associated Network object
- `cardtype` - "Ethernet", "InfiniBand", etc.
- `bootable` - Can PXE boot from this interface
- `startIf` - When to bring up interface

#### Network Interface Types

| childType | Description |
|-----------|-------------|
| `NetworkPhysicalInterface` | Physical NIC |
| `NetworkBondInterface` | Bonded interfaces |
| `NetworkBridgeInterface` | Bridge interfaces |
| `NetworkVlanInterface` | VLAN interfaces |

### Role Assignments

#### roles
Array of Role objects assigned to the device:

```json
{
  "baseType": "Role",
  "childType": "HeadNodeRole",
  "name": "headnode",
  "uuid": "<role_uuid>",
  "addServices": true
}
```

**Common Roles:**
- `HeadNodeRole` - Cluster management
- `StorageRole` - NFS storage
- `BackupRole` - Backup services
- `MonitoringRole` - Monitoring agent
- `ProvisioningRole` - Node provisioning
- `BootRole` - PXE boot services
- `ComputeRole` - Compute workload

### Storage Configuration

#### fsexports
Array of NFS export configurations:

```json
{
  "baseType": "FSExport",
  "path": "/home",
  "network": "<network_uuid>",
  "allowWrite": true,
  "rootSquash": false,
  "async": true
}
```

#### fsmounts
Array of filesystem mount configurations:

```json
{
  "baseType": "FSMount",
  "path": "/shared",
  "device": "nfs-server:/export",
  "type": "nfs",
  "options": "defaults"
}
```

### Hardware Configuration

#### gpuSettings
Array of GPU configurations:

```json
{
  "baseType": "GPUSetting",
  "deviceId": "0",
  "model": "Tesla V100",
  "computeMode": "default"
}
```

#### biosSetup
BIOS/UEFI configuration object (when available)

#### bmcSettings
BMC (IPMI/Redfish) configuration:

```json
{
  "baseType": "BMCSettings",
  "ip": "172.21.100.1",
  "username": "admin",
  "type": "IPMI"
}
```

### Provisioning & Boot

| Field | Type | Description |
|-------|------|-------------|
| `bootLoader` | string | Bootloader type (GRUB, PXELINUX, etc.) |
| `bootLoaderProtocol` | string | Boot protocol |
| `provisioningInterface` | string | Interface UUID for provisioning |
| `provisioningTransport` | string | RSYNCDAEMON, HTTP, etc. |
| `pxelabel` | string | PXE boot label |
| `installMode` | string | Installation mode |
| `nextBootInstallMode` | string | Next boot installation mode |

### Categorization

| Field | Type | Description |
|-------|------|-------------|
| `category` | string | Category UUID |
| `partition` | string | Partition UUID |
| `managementNetwork` | string | Management network UUID |

### Power Management

| Field | Type | Description |
|-------|------|-------------|
| `powerControl` | string | "none", "ipmi", "redfish", "pdu", etc. |
| `powerDistributionUnits` | array | Associated PDUs |
| `customPowerScript` | string | Custom power control script |
| `customPowerScriptArgument` | string | Script arguments |

### Monitoring

| Field | Type | Description |
|-------|------|-------------|
| `prometheusMetricForwarders` | array | Prometheus exporters |
| `customPingScript` | string | Custom ping/health check script |

### Additional Configuration

| Field | Type | Description |
|-------|------|-------------|
| `kernelParameters` | string | Kernel boot parameters |
| `kernelVersion` | string | Kernel version |
| `kernelOutputConsole` | string | Console output device |
| `modules` | array | Kernel modules to load |
| `ioScheduler` | string | I/O scheduler type |
| `disksetup` | string | Disk setup script |
| `finalize` | string | Finalization script |
| `initialize` | string | Initialization script |
| `notes` | string | Administrative notes |

## Python Client Examples

### Basic Usage

```python
from capture_device_api import BCMApiClient

# Initialize and login
client = BCMApiClient(
    base_url="https://172.21.15.254:8081",
    username="root",
    password="Hashicorp123!"
)
client.login()

# Get all nodes
nodes = client.call_api("cmdevice", "getNodes")
print(f"Total nodes: {len(nodes)}")

# Get specific node
master = client.call_api("cmdevice", "getNode", "master")
print(f"Master hostname: {master['hostname']}")
print(f"Master UUID: {master['uuid']}")

# Get compute nodes only
compute_nodes = client.call_api("cmdevice", "getComputeNodes")
for node in compute_nodes:
    print(f"{node['hostname']}: {node['childType']}")
```

### Querying Node Information

```python
# Get all nodes and filter by type
all_nodes = client.call_api("cmdevice", "getNodes")

head_nodes = [n for n in all_nodes if n['childType'] == 'HeadNode']
physical_nodes = [n for n in all_nodes if n['childType'] == 'PhysicalNode']
cloud_nodes = [n for n in all_nodes if n['childType'] == 'CloudNode']

print(f"Head nodes: {len(head_nodes)}")
print(f"Physical nodes: {len(physical_nodes)}")
print(f"Cloud nodes: {len(cloud_nodes)}")
```

### Working with Network Interfaces

```python
# Get node and display network configuration
node = client.call_api("cmdevice", "getNode", "node001")

print(f"Node: {node['hostname']}")
print(f"Interfaces:")
for iface in node['interfaces']:
    print(f"  {iface['name']}: {iface['ip']} ({iface['mac']})")
    print(f"    Type: {iface['childType']}")
    print(f"    Bootable: {iface['bootable']}")
```

### Working with Roles

```python
# Get node roles
node = client.call_api("cmdevice", "getNode", "master")

print(f"Roles for {node['hostname']}:")
for role in node['roles']:
    print(f"  {role['name']} ({role['childType']})")
```

### Filtering by Category

```python
# Get categories
categories = client.call_api("cmdevice", "getCategories")
print("Available categories:")
for cat in categories:
    print(f"  {cat['name']}: {cat['uuid']}")

# Get nodes and filter by category
nodes = client.call_api("cmdevice", "getNodes")
default_category_uuid = categories[0]['uuid']

default_nodes = [n for n in nodes if n['category'] == default_category_uuid]
print(f"Nodes in default category: {len(default_nodes)}")
```

## Error Handling

### Common Errors

#### 400 Bad Request
- Invalid method name
- Missing required parameter
- Malformed request

```python
try:
    result = client.call_api("cmdevice", "invalidMethod")
except requests.exceptions.HTTPError as e:
    if e.response.status_code == 400:
        print("Bad request - check method name and parameters")
```

#### 401 Unauthorized
- Session cookie expired
- Invalid credentials

```python
# Re-authenticate
client.login()
```

#### 404 Not Found
- Node doesn't exist
- Invalid UUID or hostname

```python
try:
    node = client.call_api("cmdevice", "getNode", "nonexistent")
except Exception as e:
    print("Node not found")
```

## Advanced Usage

### Bulk Operations

```python
# Get all nodes and extract specific information
nodes = client.call_api("cmdevice", "getNodes")

inventory = []
for node in nodes:
    inventory.append({
        'hostname': node['hostname'],
        'uuid': node['uuid'],
        'type': node['childType'],
        'mac': node['mac'],
        'ip': node['interfaces'][0]['ip'] if node['interfaces'] else None
    })

# Save to file
import json
with open('node_inventory.json', 'w') as f:
    json.dump(inventory, f, indent=2)
```

### Node Status Monitoring

```python
# Monitor nodes periodically
import time

def monitor_nodes(client, interval=60):
    """Monitor node status"""
    while True:
        nodes = client.call_api("cmdevice", "getNodes")

        print(f"\nNode Status at {time.strftime('%Y-%m-%d %H:%M:%S')}")
        print("="*60)

        for node in nodes:
            print(f"{node['hostname']:20} {node['childType']:20} {node['mac']}")

        time.sleep(interval)

# Run monitoring (Ctrl+C to stop)
# monitor_nodes(client)
```

### Exporting to CSV

```python
import csv

# Get nodes and export to CSV
nodes = client.call_api("cmdevice", "getNodes")

with open('nodes.csv', 'w', newline='') as f:
    writer = csv.writer(f)

    # Header
    writer.writerow(['Hostname', 'UUID', 'Type', 'MAC', 'IP', 'Roles'])

    # Data
    for node in nodes:
        ip = node['interfaces'][0]['ip'] if node['interfaces'] else ''
        roles = ', '.join([r['name'] for r in node['roles']])

        writer.writerow([
            node['hostname'],
            node['uuid'],
            node['childType'],
            node['mac'],
            ip,
            roles
        ])
```

## Likely Additional Methods

Based on common REST/RPC API patterns, these methods probably exist but require testing:

### Node Management
- `createNode` - Create new node
- `updateNode` - Update node configuration
- `deleteNode` - Remove node
- `cloneNode` - Clone node configuration

### Power Control
- `powerOn` - Power on node
- `powerOff` - Power off node
- `powerCycle` - Power cycle node
- `powerStatus` - Get power status

### Provisioning
- `provisionNode` - Start provisioning
- `reprovisionNode` - Re-provision node
- `getProvisioningStatus` - Get provisioning status

### Services
- `startServices` - Start services on node
- `stopServices` - Stop services
- `restartServices` - Restart services
- `getServiceStatus` - Get service status

### Hardware
- `getHardwareInfo` - Get hardware details
- `updateBIOS` - Update BIOS settings
- `updateBMC` - Update BMC settings

## Related Services

- **CMNet** - Network management
- **CMPart** - Software images and partitions
- **CMProv** - Provisioning services
- **CMMon** - Monitoring
- **CMJob** - Job management

## References

- Main documentation: `BCM_API_Complete_Documentation.md`
- Device Entity spec: `DeviceEntity.md`
- Python client: `capture_device_api.py`
- API explorer: `explore_cmdevice.py`
- Discovered methods: `cmdevice_discovered_methods_<timestamp>.json`

## Notes

1. All timestamps are Unix epoch (seconds since 1970)
2. UUIDs follow RFC 4122 format
3. Service name is lowercase (`cmdevice`) in API calls
4. Entity types use PascalCase (`HeadNode`, `PhysicalNode`)
5. Method names use camelCase (`getNodes`, `getComputeNodes`)
6. The API is session-based with cookie authentication
7. Response format is always JSON
8. The API is case-sensitive

## Changelog

- 2025-11-20: Initial documentation created from API exploration
- Methods discovered: `getNodes`, `getDevices`, `getNode`, `getComputeNodes`, `getCategories`, `reboot`
