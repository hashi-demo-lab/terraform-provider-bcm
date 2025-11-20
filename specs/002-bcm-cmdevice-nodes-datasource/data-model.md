# Data Model: BCM CMDevice Nodes Data Source

**Feature**: bcm_cmdevice_nodes
**Phase**: 1 - Design & Contracts
**Date**: 2025-11-20

## Overview

This document defines the complete data model for the `bcm_cmdevice_nodes` Terraform data source. The model follows Terraform Plugin Framework conventions with strong typing, null-safe field handling, and nested attribute structures.

---

## Entity Relationship

```
CMDeviceNodesDataSource
└── CMDeviceNodesDataSourceModel
    ├── ID (string, computed)
    ├── Filter (FilterModel, optional)
    └── Nodes ([]NodeModel, computed)
        ├── NetworkInterfaces ([]NetworkInterfaceModel)
        └── Roles ([]RoleModel)
```

---

## Data Source Root Model

### CMDeviceNodesDataSourceModel

**Purpose**: Root state model for the data source

**Go Struct**:
```go
type CMDeviceNodesDataSourceModel struct {
    ID     types.String  `tfsdk:"id"`
    Filter *FilterModel  `tfsdk:"filter"`
    Nodes  []NodeModel   `tfsdk:"nodes"`
}
```

**Terraform Schema**:
```hcl
data "bcm_cmdevice_nodes" "example" {
  filter {
    node_type        = "PhysicalNode"
    category_uuid    = "uuid-string"
    hostname_pattern = "node"
  }
}
```

| Attribute | Type | Required | Computed | Description |
|-----------|------|----------|----------|-------------|
| id | string | No | Yes | Placeholder identifier (always "placeholder") |
| filter | FilterModel | No | No | Optional filtering criteria |
| nodes | []NodeModel | No | Yes | List of matching nodes |

**Validation Rules**:
- `nodes` is never null (empty array if no matches)
- `filter` is null if not specified
- `id` is always set to "placeholder" after successful read

---

## Filter Model

### FilterModel

**Purpose**: Client-side filtering configuration

**Go Struct**:
```go
type FilterModel struct {
    NodeType        types.String `tfsdk:"node_type"`
    CategoryUUID    types.String `tfsdk:"category_uuid"`
    HostnamePattern types.String `tfsdk:"hostname_pattern"`
}
```

**Terraform Block**:
```hcl
filter {
  node_type        = "PhysicalNode"           # Exact match on childType
  category_uuid    = "uuid-here"              # Exact match on category
  hostname_pattern = "compute"                # Substring match (case-insensitive)
}
```

| Attribute | Type | Optional | Description | Matching Logic |
|-----------|------|----------|-------------|----------------|
| node_type | string | Yes | Filter by childType | Exact match, case-sensitive |
| category_uuid | string | Yes | Filter by category UUID | Exact match |
| hostname_pattern | string | Yes | Filter by hostname | Substring match, case-insensitive |

**Validation Rules**:
- All attributes are optional (omitting all = no filtering)
- Multiple filters are AND-ed together
- Empty strings treated as null (no filter)

**Filter Logic**:
```go
func matchesFilter(node NodeModel, filter *FilterModel) bool {
    if filter == nil {
        return true
    }

    // node_type filter
    if !filter.NodeType.IsNull() && !filter.NodeType.IsUnknown() {
        if node.ChildType.ValueString() != filter.NodeType.ValueString() {
            return false
        }
    }

    // category_uuid filter
    if !filter.CategoryUUID.IsNull() && !filter.CategoryUUID.IsUnknown() {
        if node.Category.ValueString() != filter.CategoryUUID.ValueString() {
            return false
        }
    }

    // hostname_pattern filter (case-insensitive substring)
    if !filter.HostnamePattern.IsNull() && !filter.HostnamePattern.IsUnknown() {
        pattern := strings.ToLower(filter.HostnamePattern.ValueString())
        hostname := strings.ToLower(node.Hostname.ValueString())
        if !strings.Contains(hostname, pattern) {
            return false
        }
    }

    return true
}
```

---

## Node Model

### NodeModel

**Purpose**: Complete representation of a BCM cluster node

**Go Struct**:
```go
type NodeModel struct {
    // Identity
    ID                    types.String `tfsdk:"id"`
    UUID                  types.String `tfsdk:"uuid"`
    Hostname              types.String `tfsdk:"hostname"`
    BaseType              types.String `tfsdk:"base_type"`
    ChildType             types.String `tfsdk:"child_type"`
    MAC                   types.String `tfsdk:"mac"`
    CreationTime          types.Int64  `tfsdk:"creation_time"`

    // Network
    Interfaces            []NetworkInterfaceModel `tfsdk:"interfaces"`

    // Roles
    Roles                 []RoleModel `tfsdk:"roles"`

    // Categorization
    Category              types.String `tfsdk:"category"`
    Partition             types.String `tfsdk:"partition"`

    // Management
    PowerControl          types.String `tfsdk:"power_control"`
    AuthenticationService types.String `tfsdk:"authentication_service"`
    ProvisioningTransport types.String `tfsdk:"provisioning_transport"`

    // State
    Modified              types.Bool `tfsdk:"modified"`
    ToBeRemoved           types.Bool `tfsdk:"to_be_removed"`
}
```

**Terraform Access Pattern**:
```hcl
# Single node access
hostname = data.bcm_cmdevice_nodes.all.nodes[0].hostname
uuid     = data.bcm_cmdevice_nodes.all.nodes[0].uuid

# Iteration
node_list = [for node in data.bcm_cmdevice_nodes.all.nodes : node.hostname]

# Conditional
compute_nodes = [for node in data.bcm_cmdevice_nodes.all.nodes :
  node if node.child_type == "ComputeNode"
]
```

### Node Attributes

#### Identity Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| id | string | Yes | No | Node UUID (same as uuid) | uuid |
| uuid | string | Yes | No | Unique identifier | uuid |
| hostname | string | Yes | No | Node hostname | hostname |
| base_type | string | Yes | No | Always "Device" | baseType |
| child_type | string | Yes | No | Node type | childType |
| mac | string | Yes | No | Primary MAC address | mac |
| creation_time | int64 | Yes | No | Unix timestamp | creationTime |

**child_type Enumeration**:
- `PhysicalNode` - Physical compute node
- `HeadNode` - Cluster management node
- `ComputeNode` - Compute worker node
- `CloudNode` - Cloud-based node
- `StorageNode` - Storage node

#### Network Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| interfaces | []NetworkInterfaceModel | Yes | No | Network interfaces (see below) | interfaces |

**Note**: Empty array `[]` if node has no interfaces configured

#### Role Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| roles | []RoleModel | Yes | No | Assigned roles (see below) | roles |

**Note**: Empty array `[]` if node has no roles assigned

#### Categorization Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| category | string | Yes | Yes | Category UUID | category |
| partition | string | Yes | Yes | Partition UUID | partition |

#### Management Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| power_control | string | Yes | Yes | Power control type | powerControl |
| authentication_service | string | Yes | Yes | Auth service type | authenticationService |
| provisioning_transport | string | Yes | Yes | Provisioning method | provisioningTransport |

**power_control Enumeration**:
- `none` - No power control
- `ipmi` - IPMI control
- `redfish` - Redfish API
- `pdu` - PDU control

#### State Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| modified | bool | Yes | No | Has unsaved changes | modified |
| to_be_removed | bool | Yes | No | Scheduled for deletion | to_be_removed |

---

## Network Interface Model

### NetworkInterfaceModel

**Purpose**: Network interface configuration for a node

**Go Struct**:
```go
type NetworkInterfaceModel struct {
    Name      types.String `tfsdk:"name"`
    MAC       types.String `tfsdk:"mac"`
    IP        types.String `tfsdk:"ip"`
    IPv6IP    types.String `tfsdk:"ipv6_ip"`
    DHCP      types.Bool   `tfsdk:"dhcp"`
    Network   types.String `tfsdk:"network"`
    BaseType  types.String `tfsdk:"base_type"`
    ChildType types.String `tfsdk:"child_type"`
    CardType  types.String `tfsdk:"cardtype"`
    Bootable  types.Bool   `tfsdk:"bootable"`
    StartIf   types.String `tfsdk:"start_if"`
}
```

**Terraform Access**:
```hcl
# First interface of first node
primary_ip = data.bcm_cmdevice_nodes.all.nodes[0].interfaces[0].ip

# All IPs from all nodes
all_ips = flatten([
  for node in data.bcm_cmdevice_nodes.all.nodes : [
    for iface in node.interfaces : iface.ip if iface.ip != null
  ]
])
```

### Interface Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| name | string | Yes | No | Interface name (e.g., ens33) | name |
| mac | string | Yes | No | Interface MAC address | mac |
| ip | string | Yes | Yes | IPv4 address | ip |
| ipv6_ip | string | Yes | Yes | IPv6 address | ipv6Ip |
| dhcp | bool | Yes | No | DHCP enabled | dhcp |
| network | string | Yes | Yes | Network UUID reference | network |
| base_type | string | Yes | No | Always "NetworkInterface" | baseType |
| child_type | string | Yes | No | Interface type | childType |
| cardtype | string | Yes | Yes | Card type | cardtype |
| bootable | bool | Yes | No | PXE boot capable | bootable |
| start_if | string | Yes | Yes | Startup condition | startIf |

**child_type Enumeration**:
- `NetworkPhysicalInterface` - Physical NIC
- `NetworkBondInterface` - Bonded interfaces
- `NetworkBridgeInterface` - Bridge interfaces
- `NetworkVlanInterface` - VLAN interfaces

**cardtype Examples**:
- `Ethernet` - Standard Ethernet
- `InfiniBand` - InfiniBand HCA
- `Wireless` - Wireless interface

**start_if Values**:
- `ALWAYS` - Start on boot
- `NEVER` - Manual start only
- `HOTPLUG` - Start on hotplug

---

## Role Model

### RoleModel

**Purpose**: Service role assignment for a node

**Go Struct**:
```go
type RoleModel struct {
    UUID        types.String `tfsdk:"uuid"`
    Name        types.String `tfsdk:"name"`
    BaseType    types.String `tfsdk:"base_type"`
    ChildType   types.String `tfsdk:"child_type"`
    AddServices types.Bool   `tfsdk:"add_services"`
}
```

**Terraform Access**:
```hcl
# Role names for first node
role_names = [for role in data.bcm_cmdevice_nodes.all.nodes[0].roles : role.name]

# Nodes with specific role
head_nodes = [for node in data.bcm_cmdevice_nodes.all.nodes :
  node if contains([for role in node.roles : role.name], "headnode")
]
```

### Role Attributes

| Attribute | Type | Computed | Nullable | Description | API Field |
|-----------|------|----------|----------|-------------|-----------|
| uuid | string | Yes | No | Role UUID | uuid |
| name | string | Yes | No | Role name | name |
| base_type | string | Yes | No | Always "Role" | baseType |
| child_type | string | Yes | No | Role type | childType |
| add_services | bool | Yes | No | Auto-add related services | addServices |

**Common Role Types** (child_type):
- `HeadNodeRole` - Cluster management
- `ComputeRole` - Compute workload
- `StorageRole` - NFS storage
- `BackupRole` - Backup services
- `MonitoringRole` - Monitoring agent
- `ProvisioningRole` - Node provisioning
- `BootRole` - PXE boot services

---

## API to Model Mapping

### Mapping Function

**Purpose**: Convert API JSON response to Terraform models

```go
func mapAPIResponseToNode(apiData map[string]interface{}) NodeModel {
    model := NodeModel{}

    // Identity
    model.UUID = getStringValue(apiData, "uuid")
    model.ID = model.UUID // ID is same as UUID
    model.Hostname = getStringValue(apiData, "hostname")
    model.BaseType = getStringValue(apiData, "baseType")
    model.ChildType = getStringValue(apiData, "childType")
    model.MAC = getStringValue(apiData, "mac")
    model.CreationTime = getInt64Value(apiData, "creationTime")

    // Categorization
    model.Category = getStringValue(apiData, "category")
    model.Partition = getStringValue(apiData, "partition")

    // Management
    model.PowerControl = getStringValue(apiData, "powerControl")
    model.AuthenticationService = getStringValue(apiData, "authenticationService")
    model.ProvisioningTransport = getStringValue(apiData, "provisioningTransport")

    // State
    model.Modified = getBoolValue(apiData, "modified")
    model.ToBeRemoved = getBoolValue(apiData, "to_be_removed")

    // Network Interfaces
    model.Interfaces = mapInterfaces(apiData["interfaces"])

    // Roles
    model.Roles = mapRoles(apiData["roles"])

    return model
}

func mapInterfaces(data interface{}) []NetworkInterfaceModel {
    interfaceArray, ok := data.([]interface{})
    if !ok || interfaceArray == nil {
        return []NetworkInterfaceModel{}
    }

    models := make([]NetworkInterfaceModel, 0, len(interfaceArray))
    for _, iface := range interfaceArray {
        ifaceMap, ok := iface.(map[string]interface{})
        if !ok {
            continue
        }

        model := NetworkInterfaceModel{
            Name:      getStringValue(ifaceMap, "name"),
            MAC:       getStringValue(ifaceMap, "mac"),
            IP:        getStringValue(ifaceMap, "ip"),
            IPv6IP:    getStringValue(ifaceMap, "ipv6Ip"),
            DHCP:      getBoolValue(ifaceMap, "dhcp"),
            Network:   getStringValue(ifaceMap, "network"),
            BaseType:  getStringValue(ifaceMap, "baseType"),
            ChildType: getStringValue(ifaceMap, "childType"),
            CardType:  getStringValue(ifaceMap, "cardtype"),
            Bootable:  getBoolValue(ifaceMap, "bootable"),
            StartIf:   getStringValue(ifaceMap, "startIf"),
        }
        models = append(models, model)
    }

    return models
}

func mapRoles(data interface{}) []RoleModel {
    roleArray, ok := data.([]interface{})
    if !ok || roleArray == nil {
        return []RoleModel{}
    }

    models := make([]RoleModel, 0, len(roleArray))
    for _, role := range roleArray {
        roleMap, ok := role.(map[string]interface{})
        if !ok {
            continue
        }

        model := RoleModel{
            UUID:        getStringValue(roleMap, "uuid"),
            Name:        getStringValue(roleMap, "name"),
            BaseType:    getStringValue(roleMap, "baseType"),
            ChildType:   getStringValue(roleMap, "childType"),
            AddServices: getBoolValue(roleMap, "addServices"),
        }
        models = append(models, model)
    }

    return models
}
```

---

## Null Handling

### Null Value Strategy

**Principle**: Use Terraform Framework null types for optional/missing fields

| Scenario | API Value | Terraform Value | Rationale |
|----------|-----------|-----------------|-----------|
| Missing field | Field absent | types.StringNull() | Field not in response |
| Null value | `null` | types.StringNull() | Explicit null |
| Empty string | `""` | types.StringNull() | Normalize empty to null |
| Empty array | `[]` | `[]Model{}` | Empty slice, not null |
| Zero integer | `0` | types.Int64Value(0) | Valid value |
| False boolean | `false` | types.BoolValue(false) | Valid value |

### Helper Functions

```go
func getStringValue(data map[string]interface{}, key string) types.String {
    if val, ok := data[key]; ok && val != nil {
        if str, ok := val.(string); ok && str != "" {
            return types.StringValue(str)
        }
    }
    return types.StringNull()
}

func getBoolValue(data map[string]interface{}, key string) types.Bool {
    if val, ok := data[key]; ok && val != nil {
        if b, ok := val.(bool); ok {
            return types.BoolValue(b)
        }
    }
    return types.BoolNull()
}

func getInt64Value(data map[string]interface{}, key string) types.Int64 {
    if val, ok := data[key]; ok && val != nil {
        switch v := val.(type) {
        case float64:
            return types.Int64Value(int64(v))
        case int64:
            return types.Int64Value(v)
        case int:
            return types.Int64Value(int64(v))
        }
    }
    return types.Int64Null()
}
```

---

## Validation Rules

### Schema-Level Validation

1. **No Required User Inputs**: All user-facing attributes are optional (filter) or computed (nodes)
2. **Computed Attributes**: nodes, id always set by provider
3. **Optional Blocks**: filter block is optional (null if omitted)

### Runtime Validation

1. **UUID Format**: Not validated (trust API)
2. **MAC Address Format**: Not validated (trust API)
3. **IP Address Format**: Not validated (trust API)
4. **Hostname**: Any string accepted

**Rationale**: API is source of truth, provider is read-only data source

---

## Performance Considerations

### Memory Allocation

**Estimated Memory** (100 nodes):
- NodeModel: ~1KB per node
- NetworkInterfaceModel: ~200 bytes each (avg 3 per node)
- RoleModel: ~100 bytes each (avg 2 per node)
- **Total**: ~150KB for 100 nodes

**Acceptable**: Well within Terraform memory limits

### CPU Usage

**Filtering Algorithm**: O(n) linear scan

**Performance**:
- 100 nodes: <10ms
- 200 nodes: <20ms

**Acceptable**: Negligible compared to API call latency

---

## Testing Considerations

### Test Data Requirements

**Minimum Test Cluster**:
- 1 HeadNode with 2+ interfaces, 2+ roles
- 1 PhysicalNode with 1+ interface, 1+ role
- 1 Node with no interfaces (edge case)
- 1 Node with no roles (edge case)

### Test Scenarios

1. **All Nodes Query**: Verify complete data retrieval
2. **Filter by Type**: Verify correct filtering
3. **Filter by Category**: Verify UUID matching
4. **Filter by Hostname**: Verify substring matching
5. **Empty Cluster**: Verify empty array handling
6. **Nested Attributes**: Verify interfaces/roles structure

---

## Example Usage

### Complete Example

```hcl
data "bcm_cmdevice_nodes" "all" {}

# Extract compute inventory
locals {
  compute_inventory = {
    for node in data.bcm_cmdevice_nodes.all.nodes :
    node.hostname => {
      uuid          = node.uuid
      type          = node.child_type
      primary_ip    = length(node.interfaces) > 0 ? node.interfaces[0].ip : null
      primary_mac   = node.mac
      power_method  = node.power_control
      roles         = [for role in node.roles : role.name]
      category      = node.category
    }
  }
}

output "inventory" {
  value = local.compute_inventory
}

# Output example:
# inventory = {
#   "node002" = {
#     uuid         = "2870c0b0-6fda-4026-9b8f-28be4c372fee"
#     type         = "PhysicalNode"
#     primary_ip   = "172.21.15.254"
#     primary_mac  = "00:00:00:00:00:00"
#     power_method = "none"
#     roles        = ["compute"]
#     category     = "0ae6d733-3015-4479-bfab-ce2d237a2809"
#   }
# }
```

---

## References

- API Contract: `contracts/cmdevice_getNodes_contract.json`
- Research Findings: `research.md`
- Feature Spec: `spec.md`
- Terraform Plugin Framework: https://developer.hashicorp.com/terraform/plugin/framework
