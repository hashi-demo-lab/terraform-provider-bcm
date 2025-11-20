# CMDevice API Documentation Suite

Complete documentation for the NVIDIA Bright Cluster Manager (BCM) CMDevice service.

## Overview

The CMDevice service is the primary API for managing cluster nodes (devices) in BCM. This documentation suite provides comprehensive coverage of all discovered methods, entity structures, and usage patterns.

## Documentation Files

### 1. Quick Reference Guide
**File:** `CMDevice_Quick_Reference.md`

**Contents:**
- Quick start examples
- Verified working methods
- Common response fields
- Common patterns and filters
- cURL examples
- Error handling

**Best for:** Quick lookups and copy-paste examples

### 2. Complete Documentation
**File:** `CMDevice_Complete_Documentation.md`

**Contents:**
- Full API method reference
- Complete entity structure documentation
- All device types and their fields
- Network interface configuration
- Role assignments
- Storage configuration
- Hardware settings
- Python client examples
- Advanced usage patterns
- Error handling

**Best for:** Comprehensive understanding and advanced usage

### 3. Device Entity Specification
**File:** `DeviceEntity.md`

**Contents:**
- Device entity structure
- Field definitions
- API endpoint details
- Authentication flow
- Python client usage

**Best for:** Understanding the Device entity model

### 4. General API Documentation
**File:** `BCM_API_Complete_Documentation.md`

**Contents:**
- Overall API architecture
- All discovered services (CMDevice, CMNet, CMPart, etc.)
- Authentication
- Service naming conventions
- Python client library
- API discovery methods

**Best for:** Understanding the broader BCM API ecosystem

## Tools & Scripts

### API Client
**File:** `capture_device_api.py`

Working Python client with BCMApiClient class for making authenticated API calls.

**Usage:**
```python
from capture_device_api import BCMApiClient

client = BCMApiClient("https://172.21.15.254:8081", "root", "password")
client.login()
nodes = client.call_api("cmdevice", "getNodes")
```

### API Explorer
**File:** `explore_cmdevice.py`

Automated script to discover CMDevice API methods by testing common patterns.

**Run:**
```bash
python3 explore_cmdevice.py
```

**Output:**
- `cmdevice_discovered_methods_<timestamp>.json` - Working methods
- `cmdevice_failed_methods_<timestamp>.json` - Failed attempts

### HTML Scraper
**File:** `scrape_api_docs.py`

Scrapes HTML pages from the BCM API documentation interface.

### JavaScript Analyzer
**File:** `analyze_react_app.py`

Analyzes the React app JavaScript bundle to discover API patterns and services.

**Run:**
```bash
python3 analyze_react_app.py
```

**Output:**
- `react_bundle_<timestamp>.js` - Full React app source
- `react_bundle_analysis_<timestamp>.json` - Extracted API patterns

### Selenium Scraper
**File:** `scrape_api_docs_selenium.py`

Advanced scraper using Selenium for JavaScript-rendered content (requires Chrome/Chromium).

## Discovered Methods

### Verified Working Methods

| Method | Arguments | Returns | Description |
|--------|-----------|---------|-------------|
| `getNodes` | None | Array | Get all nodes |
| `getDevices` | None | Array | Get all devices (alias) |
| `getNode` | hostname/UUID | Object | Get specific node |
| `getComputeNodes` | None | Array | Get compute nodes only |
| `getCategories` | None | Array | Get all categories |
| `reboot` | hostname | Object | Reboot a node |

### Likely Available Methods

Based on common API patterns (require testing):
- `createNode`, `updateNode`, `deleteNode`, `cloneNode`
- `powerOn`, `powerOff`, `powerCycle`, `powerStatus`
- `provisionNode`, `reprovisionNode`, `getProvisioningStatus`
- `startServices`, `stopServices`, `restartServices`, `getServiceStatus`
- `getHardwareInfo`, `updateBIOS`, `updateBMC`

## Quick Start

### 1. Install Dependencies
```bash
pip install --break-system-packages requests beautifulsoup4
```

### 2. Basic Usage
```python
from capture_device_api import BCMApiClient

# Initialize
client = BCMApiClient(
    base_url="https://172.21.15.254:8081",
    username="root",
    password="Hashicorp123!"
)

# Login
client.login()

# Get all nodes
nodes = client.call_api("cmdevice", "getNodes")
print(f"Total nodes: {len(nodes)}")

# Get specific node
master = client.call_api("cmdevice", "getNode", "master")
print(f"Master: {master['hostname']}")
```

### 3. cURL Example
```bash
# Login
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"Hashicorp123!"}' \
  -c cookies.txt

# Get nodes
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -b cookies.txt \
  -d '{"service":"cmdevice","call":"getNodes"}'
```

## Device Entity Structure

### Core Fields
```json
{
  "baseType": "Device",
  "childType": "HeadNode | PhysicalNode | ComputeNode | CloudNode",
  "uuid": "unique-identifier",
  "hostname": "node-name",
  "mac": "00:50:56:9B:E4:6D",
  "interfaces": [...],
  "roles": [...],
  "category": "category-uuid",
  "powerControl": "none | ipmi | redfish | pdu"
}
```

### Network Interfaces
```json
{
  "name": "ens33",
  "mac": "00:50:56:9B:E4:6D",
  "ip": "172.21.15.254",
  "childType": "NetworkPhysicalInterface",
  "bootable": false
}
```

### Roles
```json
{
  "name": "headnode",
  "childType": "HeadNodeRole",
  "uuid": "role-uuid"
}
```

## Common Use Cases

### 1. Node Inventory
```python
nodes = client.call_api("cmdevice", "getNodes")
for node in nodes:
    print(f"{node['hostname']:20} {node['childType']:20} {node['mac']}")
```

### 2. Filter by Type
```python
nodes = client.call_api("cmdevice", "getNodes")
head_nodes = [n for n in nodes if n['childType'] == 'HeadNode']
compute_nodes = [n for n in nodes if n['childType'] == 'PhysicalNode']
```

### 3. Network Configuration
```python
node = client.call_api("cmdevice", "getNode", "master")
for iface in node['interfaces']:
    print(f"{iface['name']}: {iface['ip']} ({iface['mac']})")
```

### 4. Role Management
```python
node = client.call_api("cmdevice", "getNode", "master")
for role in node['roles']:
    print(f"{role['name']} ({role['childType']})")
```

### 5. Export to CSV
```python
import csv
nodes = client.call_api("cmdevice", "getNodes")

with open('nodes.csv', 'w', newline='') as f:
    writer = csv.writer(f)
    writer.writerow(['Hostname', 'UUID', 'Type', 'MAC', 'IP'])

    for node in nodes:
        ip = node['interfaces'][0]['ip'] if node['interfaces'] else ''
        writer.writerow([
            node['hostname'],
            node['uuid'],
            node['childType'],
            node['mac'],
            ip
        ])
```

## Data Files

### API Responses
- `device_master_<timestamp>.json` - Master node details
- `cmdevice_discovered_methods_<timestamp>.json` - Working methods with responses
- `cmdevice_failed_methods_<timestamp>.json` - Failed method attempts

### React App Analysis
- `react_bundle_<timestamp>.js` - Full React application (1.4 MB)
- `react_bundle_analysis_<timestamp>.json` - Extracted patterns
- `manifest_<timestamp>.json` - Web app manifest
- `styles_<timestamp>.css` - Application styles

### HTML Scraping
- `cmdevice_docs_raw_<timestamp>.html` - Raw HTML
- `cmdevice_docs_parsed_<timestamp>.json` - Parsed content
- `cmdevice_react_metadata_<timestamp>.json` - React metadata

## Related Services

The CMDevice service is part of a larger API ecosystem:

- **CMAuth** - Authentication and authorization
- **CMNet** - Network management
- **CMPart** - Software images and partitions
- **CMProv** - Provisioning services
- **CMJob** - Job management
- **CMServ** - Service management
- **CMMon** - Monitoring
- **CMKube** - Kubernetes integration
- **CMCloud** - Cloud provider integration
- **CMBeeGFS** - BeeGFS filesystem

## API Conventions

### Service Names
- Lowercase in API calls: `cmdevice`
- PascalCase in documentation: `CMDevice`

### Entity Types
- `baseType` - Entity category (PascalCase): `Device`, `NetworkInterface`
- `childType` - Specific subtype (PascalCase): `HeadNode`, `NetworkPhysicalInterface`

### Method Names
- camelCase: `getNodes`, `getComputeNodes`, `reboot`

### UUIDs
- RFC 4122 format
- Zero UUID: `00000000-0000-0000-0000-000000000000`

### Timestamps
- Unix epoch (seconds since 1970-01-01)
- Example: `1763617980` = 2025-11-18

## Security Notes

1. **SSL/TLS**: API uses self-signed certificates
   - Disable verification in development: `verify=False`
   - Use proper CA certificates in production

2. **Authentication**: Session-based with cookies
   - Cookie: `cm-login-token`
   - HTTPOnly and Secure flags set
   - Sessions expire after inactivity

3. **Authorization**: Role-based access control
   - Root user has full access
   - Other users have limited permissions

## Troubleshooting

### Connection Refused
```
Connection refused: [Errno 111]
```
**Solution:** Check if the BCM service is running and accessible

### Bad Request (400)
```
400 Client Error: Bad Request
```
**Solution:** Verify method name is correct and arguments are valid

### Unauthorized (401)
```
401 Unauthorized
```
**Solution:** Re-authenticate with `client.login()`

### Session Expired
**Solution:** Login again to get new session cookie

## Contributing

To extend this documentation:

1. Run API explorer: `python3 explore_cmdevice.py`
2. Test new methods and document results
3. Update relevant documentation files
4. Add examples to Quick Reference

## Changelog

- **2025-11-20**: Initial comprehensive documentation
  - Discovered 6 working CMDevice methods
  - Created complete API documentation
  - Built Python client library
  - Analyzed React app (1.4 MB JavaScript bundle)
  - Extracted 30+ service names
  - Documented Device entity structure

## License

Documentation for NVIDIA Bright Cluster Manager API.

## Support

For issues or questions:
1. Check the documentation files
2. Review captured API responses
3. Run the API explorer to test methods
4. Examine the React bundle analysis for patterns
