# BCM API Complete Documentation

## Overview

NVIDIA Bright Cluster Manager (BCM) provides a JSON-based RPC API for managing cluster infrastructure. The API is exposed through a React-based web interface and a JSON endpoint.

## API Architecture

### Base URL
```
https://172.21.15.254:8081
```

### Primary Endpoints

1. **JSON RPC Endpoint**: `/json` (POST)
   - Main API endpoint for all service calls
   - Requires session authentication

2. **Web UI**: `/api/*`
   - React-based documentation and interface
   - Requires JavaScript to render

## Authentication

### Login

```http
POST /json
Content-Type: application/json

{
  "service": "login",
  "username": "root",
  "password": "Hashicorp123!"
}
```

**Response:**
- Status: 200 OK
- Sets cookie: `cm-login-token=<token>`

### Authenticated Requests

All subsequent requests must include the session cookie:

```http
POST /json
Cookie: cm-login-token=<token>
Content-Type: application/json

{
  "service": "<service_name>",
  "call": "<method_name>",
  "arg": "<optional_argument>"
}
```

## Discovered Services

Based on JavaScript bundle analysis, the following services are available:

### Core Services

| Service | Description |
|---------|-------------|
| `CMAuth` | Authentication and authorization |
| `CMDevice` | Device/node management |
| `CMNet` | Network configuration |
| `CMPart` | Software images and partitions |
| `CMProv` | Provisioning services |
| `CMJob` | Job management |
| `CMServ` | Service management |
| `CMMon` | Monitoring |
| `CMKube` | Kubernetes integration |
| `CMCloud` | Cloud provider integration |
| `CMEtcd` | Etcd cluster management |
| `CMMain` | Main cluster configuration |
| `CMGui` | GUI settings |
| `CMCert` | Certificate management |
| `CMBeeGFS` | BeeGFS filesystem management |
| `CMProc` | Process management |

### Service Naming Convention

Services follow the pattern:
- `CM` prefix (Cluster Manager)
- Followed by category abbreviation
- Example: `CMDevice` = Cluster Manager Device service

## CMDevice Service

The Device service manages cluster nodes (servers, compute nodes, head nodes).

### Get Node

Retrieve device information by name or UUID:

```json
{
  "service": "cmdevice",
  "call": "getNode",
  "arg": "master"
}
```

**Response:** Complete Device entity object (see Device Entity Structure below)

### Device Entity Structure

```json
{
  "baseType": "Device",
  "childType": "HeadNode | ComputeNode | CloudNode",
  "uuid": "string (UUID)",
  "hostname": "string",
  "mac": "string (MAC address)",
  "creationTime": "integer (Unix timestamp)",
  "cmdaemonUrl": "string (URL)",
  "authenticationService": "string",
  "powerControl": "string",
  "provisioningTransport": "string",
  "interfaces": [NetworkInterface],
  "roles": [Role],
  "fsexports": [FSExport],
  "services": [Service],
  "staticRoutes": [Route],
  "gpuSettings": [GPUSettings],
  "biosSetup": BIOSSetup,
  "bmcSettings": BMCSettings
}
```

### Device Types (childType)

- `HeadNode` - Cluster management node
- `ComputeNode` - Compute/worker node
- `CloudNode` - Cloud-based node
- `StorageNode` - Storage node

## CMPart Service

Software image and partition management.

### Get Software Images

```json
{
  "service": "CMPart",
  "call": "getSoftwareImages"
}
```

**Response:** Array of SoftwareImage objects

### SoftwareImage Structure

```json
{
  "baseType": "SoftwareImage",
  "uuid": "string",
  "name": "string",
  "path": "string",
  "bootfspart": "string (UUID)",
  "fspart": "string (UUID)",
  "kernelVersion": "string",
  "kernelParameters": "string",
  "modules": [KernelModule]
}
```

## Entity Types

The BCM API uses a hierarchical entity system:

### Base Types

- `Entity` - Base for all entities
- `Device` - Physical/virtual machines
- `Role` - Service roles
- `Network` - Network definitions
- `NetworkInterface` - Network interfaces
- `SoftwareImage` - OS images
- `FSExport` - NFS exports
- `FSMount` - Filesystem mounts
- `KernelModule` - Kernel modules
- `Service` - System services

### Common Patterns

All entities share common fields:
- `baseType` - Entity category
- `childType` - Specific subtype
- `uuid` - Unique identifier
- `modified` - Modification flag
- `to_be_removed` - Deletion flag
- `revision` - Version info
- `extra_values` - Extension data

## Role Types

Roles define service assignments for devices:

| Role Type | Description |
|-----------|-------------|
| `HeadNodeRole` | Cluster management |
| `StorageRole` | NFS storage services |
| `BackupRole` | Backup services |
| `MonitoringRole` | Cluster monitoring |
| `ProvisioningRole` | Node provisioning |
| `BootRole` | Boot/PXE services |
| `BeeGFSClientRole` | BeeGFS client |
| `BeeGFSMetadataRole` | BeeGFS metadata server |
| `BeeGFSStorageRole` | BeeGFS storage server |

## Network Configuration

### Network Entity

```json
{
  "baseType": "Network",
  "uuid": "string",
  "name": "string",
  "network": "string (IP)",
  "netmask": "string",
  "gateway": "string",
  "domain": "string",
  "dhcp": "boolean"
}
```

### NetworkInterface Entity

```json
{
  "baseType": "NetworkInterface",
  "childType": "NetworkPhysicalInterface | NetworkBondInterface | NetworkBridgeInterface",
  "name": "string",
  "mac": "string",
  "ip": "string",
  "ipv6Ip": "string",
  "dhcp": "boolean",
  "network": "string (Network UUID)",
  "cardtype": "Ethernet | InfiniBand"
}
```

## BeeGFS Integration

BCM includes extensive BeeGFS parallel filesystem support:

### BeeGFS Services

- `BeeGFSManagementRole` - Management service
- `BeeGFSMetadataRole` - Metadata service
- `BeeGFSStorageRole` - Storage targets
- `BeeGFSClientRole` - Client mounts
- `BeeGFSHelperRole` - Helper daemon

### BeeGFS Configuration

Each role has associated configuration entities:
- Connection settings (TCP/UDP ports, RDMA)
- Log settings
- Performance tuning
- Data directories

## Python Client Library

### Basic Usage

```python
from capture_device_api import BCMApiClient

# Initialize client
client = BCMApiClient(
    base_url="https://172.21.15.254:8081",
    username="root",
    password="Hashicorp123!"
)

# Authenticate
client.login()

# Call API
result = client.call_api("cmdevice", "getNode", "master")

# Access data
print(result["hostname"])
print(result["interfaces"])
```

### Custom API Calls

```python
# Generic API call
def call_service(client, service, method, arg=None):
    payload = {
        "service": service,
        "call": method
    }
    if arg:
        payload["arg"] = arg

    response = client.session.post(
        f"{client.base_url}/json",
        json=payload
    )
    return response.json()

# Example: List all nodes
nodes = call_service(client, "cmdevice", "listNodes")
```

## API Discovery

### JavaScript Bundle Analysis

The React app bundle contains embedded API schema:
- Entity definitions with all fields
- Parameter types and constraints
- Enum values
- Default values
- Validation rules

File: `react_bundle_<timestamp>.js` (1.4 MB)

### Discovered Services List

Complete list extracted from JavaScript:
- 180+ entity types
- 30+ service endpoints
- 500+ API methods

See `react_bundle_analysis_<timestamp>.json` for full details.

## Error Handling

### Authentication Errors

- Missing cookie: Returns login page HTML
- Invalid credentials: HTTP 401 or error in JSON response
- Expired session: Re-authenticate

### API Errors

Errors are returned in the JSON response:

```json
{
  "error": "error message",
  "code": "error_code"
}
```

## Security Considerations

1. **SSL/TLS**: Uses self-signed certificates
   - Disable SSL verification in development
   - Use proper CA certificates in production

2. **Authentication**: Session-based
   - Cookies are HTTPOnly and Secure
   - Sessions expire after inactivity

3. **Authorization**: Role-based access control
   - Root user has full access
   - Other users have limited permissions

## Captured Files Reference

### Raw Data
- `device_master_<timestamp>.json` - Master node details
- `react_bundle_<timestamp>.js` - Full React application
- `cmdevice_docs_raw_<timestamp>.html` - Raw HTML page
- `manifest_<timestamp>.json` - Web app manifest

### Analysis
- `react_bundle_analysis_<timestamp>.json` - Extracted API patterns
- `cmdevice_react_metadata_<timestamp>.json` - React app metadata
- `cmdevice_docs_parsed_<timestamp>.json` - Parsed documentation

### Scripts
- `capture_device_api.py` - Working Python client
- `scrape_api_docs.py` - HTML scraper
- `analyze_react_app.py` - JavaScript analyzer
- `scrape_api_docs_selenium.py` - Selenium scraper (requires setup)

## Next Steps

To explore more of the API:

1. **List Methods**: Try different `call` values for each service
2. **Entity CRUD**: Look for create/read/update/delete patterns
3. **Search**: Try search endpoints for entity discovery
4. **Events**: Check for event/notification APIs
5. **Streaming**: Look for WebSocket or SSE endpoints

## Additional Resources

- React app runs at: `https://172.21.15.254:8081/api/`
- Main CSS: `/api/static/css/main.9e6e7576.css`
- Main JS: `/api/static/js/main.9875e9e5.js`
- Manifest: `/api/manifest.json`

## Notes

- All timestamps are Unix epoch (seconds since 1970-01-01)
- UUIDs follow RFC 4122 format
- The API is case-sensitive for service names
- Service names are typically lowercase when called
- Entity `baseType` and `childType` use PascalCase
