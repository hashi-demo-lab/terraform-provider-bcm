# CMDevice API - Quick Reference Guide

## Quick Start

```python
from capture_device_api import BCMApiClient

client = BCMApiClient("https://172.21.15.254:8081", "root", "Hashicorp123!")
client.login()
```

## Verified Working Methods

### List All Nodes
```python
nodes = client.call_api("cmdevice", "getNodes")
# or
nodes = client.call_api("cmdevice", "getDevices")  # Alias
```

### Get Specific Node
```python
# By hostname
master = client.call_api("cmdevice", "getNode", "master")

# By UUID
node = client.call_api("cmdevice", "getNode", "uuid-here")
```

### Get Compute Nodes Only
```python
compute_nodes = client.call_api("cmdevice", "getComputeNodes")
```

### Get Categories
```python
categories = client.call_api("cmdevice", "getCategories")
```

### Reboot Node
```python
result = client.call_api("cmdevice", "reboot", "node001")
```

## Common Response Fields

### Node Object
```python
{
    "hostname": "node001",
    "uuid": "...",
    "childType": "PhysicalNode",  # or HeadNode, ComputeNode, CloudNode
    "mac": "00:50:56:9B:E4:6D",
    "interfaces": [...],
    "roles": [...],
    "category": "category-uuid",
    "powerControl": "none",  # or ipmi, redfish, pdu
    "creationTime": 1763617980
}
```

### Network Interface
```python
{
    "name": "ens33",
    "mac": "00:50:56:9B:E4:6D",
    "ip": "172.21.15.254",
    "ipv6Ip": "::0",
    "dhcp": false,
    "childType": "NetworkPhysicalInterface",
    "bootable": false
}
```

### Role Object
```python
{
    "name": "headnode",
    "childType": "HeadNodeRole",
    "uuid": "..."
}
```

## Common Patterns

### Filter Nodes by Type
```python
nodes = client.call_api("cmdevice", "getNodes")
head_nodes = [n for n in nodes if n['childType'] == 'HeadNode']
compute_nodes = [n for n in nodes if n['childType'] == 'PhysicalNode']
```

### Get Node IPs
```python
node = client.call_api("cmdevice", "getNode", "master")
for iface in node['interfaces']:
    print(f"{iface['name']}: {iface['ip']}")
```

### List Node Roles
```python
node = client.call_api("cmdevice", "getNode", "master")
for role in node['roles']:
    print(role['name'])
```

### Create Node Inventory
```python
nodes = client.call_api("cmdevice", "getNodes")
for node in nodes:
    print(f"{node['hostname']:20} {node['childType']:20} {node['mac']}")
```

## Node Types (childType)

- `HeadNode` - Management node
- `PhysicalNode` - Physical compute node
- `ComputeNode` - Compute worker
- `CloudNode` - Cloud-based node
- `StorageNode` - Storage node

## Common Role Types

- `HeadNodeRole` - Cluster management
- `StorageRole` - NFS storage
- `BackupRole` - Backup services
- `MonitoringRole` - Monitoring
- `ProvisioningRole` - Node provisioning
- `BootRole` - PXE boot services
- `ComputeRole` - Compute workload

## Network Interface Types

- `NetworkPhysicalInterface` - Physical NIC
- `NetworkBondInterface` - Bonded interfaces
- `NetworkBridgeInterface` - Bridge interfaces
- `NetworkVlanInterface` - VLAN interfaces

## Power Control Types

- `none` - No power control
- `ipmi` - IPMI (BMC)
- `redfish` - Redfish API
- `pdu` - Power Distribution Unit
- `custom` - Custom script

## Error Handling

```python
try:
    node = client.call_api("cmdevice", "getNode", "nonexistent")
except requests.exceptions.HTTPError as e:
    if e.response.status_code == 400:
        print("Bad request")
    elif e.response.status_code == 401:
        print("Unauthorized - relogin")
        client.login()
    elif e.response.status_code == 404:
        print("Not found")
```

## Export Examples

### To JSON
```python
import json
nodes = client.call_api("cmdevice", "getNodes")
with open('nodes.json', 'w') as f:
    json.dump(nodes, f, indent=2)
```

### To CSV
```python
import csv
nodes = client.call_api("cmdevice", "getNodes")
with open('nodes.csv', 'w', newline='') as f:
    writer = csv.DictWriter(f, ['hostname', 'uuid', 'type', 'mac'])
    writer.writeheader()
    for node in nodes:
        writer.writerow({
            'hostname': node['hostname'],
            'uuid': node['uuid'],
            'type': node['childType'],
            'mac': node['mac']
        })
```

## cURL Examples

### Login
```bash
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"Hashicorp123!"}' \
  -c cookies.txt
```

### Get Nodes
```bash
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"service":"cmdevice","call":"getNodes"}'
```

### Get Specific Node
```bash
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"service":"cmdevice","call":"getNode","arg":"master"}'
```

## API Request Format

```json
{
  "service": "cmdevice",
  "call": "methodName",
  "arg": "optional-argument"
}
```

## Common UUIDs

UUIDs are used throughout the API. Common zero UUID:
- `00000000-0000-0000-0000-000000000000` - Not assigned/default

## Time Formats

- `creationTime` - Unix timestamp (seconds since epoch)
- Example: `1763617980` = 2025-11-18

## Tips

1. Always login before making API calls
2. Session cookies expire after inactivity
3. Use `getNodes` for bulk operations
4. Use `getNode` for individual node details
5. Check `childType` to filter node types
6. Use `uuid` for unique identification
7. Use `hostname` for human-readable reference
8. The `modified` flag indicates unsaved changes

## Full Documentation

See `CMDevice_Complete_Documentation.md` for comprehensive details.
