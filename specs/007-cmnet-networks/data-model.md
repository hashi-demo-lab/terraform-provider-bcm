# Data Model: BCM CMNet Networks

**Feature**: bcm_cmnet_networks data source
**Date**: 2025-11-21
**Based On**: research.md API exploration findings

## Entities

### Network Entity

Represents a network configuration in the BCM cluster retrieved via `cmnet.getNetworks` API.

#### Identity Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `id` | String | Yes (Computed) | Unique network identifier (same as uuid) |
| `uuid` | String | Yes (Computed) | RFC 4122 UUID format network identifier |
| `name` | String | Yes (Computed) | Human-readable network name |

#### Network Configuration Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `base_address` | String | Yes (Computed) | Network base IP address (IPv4 format, e.g., "172.21.12.0") |
| `netmask_bits` | Int64 | Yes (Computed) | CIDR netmask bits (e.g., 22, 16, 0 for global network) |
| `gateway` | String | No (Computed) | Default gateway IP address ("0.0.0.0" indicates no gateway) |
| `network_type` | String | Yes (Computed) | Network type (INTERNAL, GLOBAL, etc.) |
| `mtu` | Int64 | Yes (Computed) | Maximum transmission unit (typically 1500) |

#### DNS Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `domain_name` | String | No (Computed) | DNS domain name for the network |
| `generate_dns_zone` | String | Yes (Computed) | DNS zone generation policy (BOTH, FORWARD, REVERSE, NONE) |
| `search_domain_index` | Int64 | No (Computed) | Search domain priority index |

#### DHCP Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `dhcp_enabled` | Bool | Yes (Computed) | Computed: true if dynamic_range_start and dynamic_range_end are not "0.0.0.0" |
| `dynamic_range_start` | String | No (Computed) | DHCP range start IP address |
| `dynamic_range_end` | String | No (Computed) | DHCP range end IP address |

**DHCP Detection Logic**:
```go
dhcpEnabled := network.DynamicRangeStart != "0.0.0.0" && network.DynamicRangeEnd != "0.0.0.0"
```

#### Network Flags

| Attribute | Type | Required | Description |
|---|---|---|---|
| `management` | Bool | Yes (Computed) | Is this a management network (used for cluster management traffic) |
| `bootable` | Bool | Yes (Computed) | Network supports PXE boot for node provisioning |
| `layer3` | Bool | Yes (Computed) | Layer 3 networking enabled |

#### IPv6 Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `ipv6_enabled` | Bool | Yes (Computed) | IPv6 protocol enabled flag |
| `ipv6_base_address` | String | No (Computed) | IPv6 network base address ("::" format, "::0" indicates not configured) |
| `ipv6_gateway` | String | No (Computed) | IPv6 gateway address |
| `ipv6_netmask_bits` | Int64 | No (Computed) | IPv6 CIDR netmask bits (0 indicates not configured) |

#### Metadata Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `base_type` | String | Yes (Computed) | Entity base type (always "Network") |
| `child_type` | String | No (Computed) | Network subtype (often empty string) |
| `revision` | String | No (Computed) | Revision identifier (may be empty) |
| `modified` | Bool | Yes (Computed) | Has unsaved changes pending |
| `to_be_removed` | Bool | Yes (Computed) | Scheduled for deletion flag |

#### Advanced Networking Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `layer3_route` | String | No (Computed) | Layer 3 routing mode (NONE, STATIC, etc.) |
| `gateway_metric` | Int64 | No (Computed) | Gateway routing metric for path selection |
| `allow_autosign` | String | Yes (Computed) | Autosign policy (AUTOMATIC, MANUAL, etc.) |

#### Cloud Integration Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `cloud_subnet_id` | String | No (Computed) | Cloud provider subnet identifier (AWS, Azure, etc.) |
| `ec2_availability_zone` | String | No (Computed) | AWS EC2 availability zone (empty for on-premises) |

#### Additional Attributes

| Attribute | Type | Required | Description |
|---|---|---|---|
| `notes` | String | No (Computed) | User-provided notes or description (often empty) |

## Relationships

### Network → Nodes (Out of Scope)

Networks are referenced by node interfaces. This relationship is managed in the node data source and resource implementations, not in the network data source.

**Future Integration**:
- `bcm_cmdevice_node` resource will reference network UUIDs for interface configuration
- Node interface configurations will use `network_uuid` to assign nodes to networks

## Filter Model

### NetworkFilter Entity

Client-side filtering criteria for network data source queries.

| Attribute | Type | Required | Description |
|---|---|---|---|
| `name_pattern` | String | No (Optional) | Case-insensitive substring match for network name |
| `dhcp_enabled` | Bool | No (Optional) | Filter by DHCP enabled status (exact match) |

**Filter Behavior**:
- Multiple filters use **AND** logic (all filters must match)
- `name_pattern` uses **case-insensitive substring** matching (`strings.Contains(strings.ToLower(name), strings.ToLower(pattern))`)
- `dhcp_enabled` uses **exact boolean** matching
- Omitting a filter attribute means "no filter" for that criterion (not "match false")

**Filter Examples**:
```hcl
# Case-insensitive substring match
filter {
  name_pattern = "management"  # Matches: "managementnet", "Management-Net", "MY-MANAGEMENT"
}

# DHCP enabled only
filter {
  dhcp_enabled = true  # Only networks with DHCP ranges configured
}

# Combined filters (AND logic)
filter {
  name_pattern = "internal"
  dhcp_enabled = true  # Networks with "internal" in name AND DHCP enabled
}
```

## Validation Rules

### Data Source Level

**No validation required** - This is a read-only data source. All data comes from the BCM API and is trusted.

### Field Level

**No field-level validation** - All attributes are computed from API response. The helper functions handle null/empty values gracefully.

### Filter Level

**No filter validation** - All filter values are optional. Invalid patterns will simply return empty results (not errors).

## State Management

### Data Source State

The data source state includes:

```hcl
{
  "id": "placeholder-id",  # Placeholder identifier for the data source instance
  "filter": {
    "name_pattern": "...",  # User-provided filter (if specified)
    "dhcp_enabled": true     # User-provided filter (if specified)
  },
  "networks": [
    {
      # Network entity attributes (all computed)
    }
  ]
}
```

### State Characteristics

- **ID**: Placeholder value (not used for lookups, data source doesn't have a real identity)
- **Filter**: Stored in state to track user intent (not used for refresh)
- **Networks**: List of network entities matching filter criteria at read time
- **Computed-Only**: All network attributes are computed (no user inputs besides filter)

## Schema Mapping Strategy

### API Field → Terraform Attribute Mapping

**Naming Convention**: API camelCase → Terraform snake_case

| API Field | Terraform Attribute | Transformation |
|---|---|---|
| `uuid` | `id`, `uuid` | Direct mapping |
| `name` | `name` | Direct mapping |
| `baseAddress` | `base_address` | Rename to snake_case |
| `netmaskBits` | `netmask_bits` | Rename to snake_case |
| `gateway` | `gateway` | Direct mapping |
| `domainName` | `domain_name` | Rename to snake_case |
| `dynamicRangeStart` | `dynamic_range_start` | Rename to snake_case |
| `dynamicRangeEnd` | `dynamic_range_end` | Rename to snake_case |
| N/A (computed) | `dhcp_enabled` | Computed from dynamic range |
| `management` | `management` | Direct mapping |
| `bootable` | `bootable` | Direct mapping |
| `type` | `network_type` | Rename to avoid reserved word |
| `mtu` | `mtu` | Direct mapping |
| `modified` | `modified` | Direct mapping |
| `to_be_removed` | `to_be_removed` | Direct mapping |
| `revision` | `revision` | Direct mapping |
| `baseType` | `base_type` | Rename to snake_case |
| `childType` | `child_type` | Rename to snake_case |
| `IPv6` | `ipv6_enabled` | Rename for clarity |
| `ipv6BaseAddress` | `ipv6_base_address` | Rename to snake_case |
| `ipv6Gateway` | `ipv6_gateway` | Rename to snake_case |
| `ipv6NetmaskBits` | `ipv6_netmask_bits` | Rename to snake_case |
| `allowAutosign` | `allow_autosign` | Rename to snake_case |
| `generateDNSZone` | `generate_dns_zone` | Rename to snake_case |
| `layer3` | `layer3` | Direct mapping |
| `layer3route` | `layer3_route` | Rename to snake_case |
| `gatewayMetric` | `gateway_metric` | Rename to snake_case |
| `searchDomainIndex` | `search_domain_index` | Rename to snake_case |
| `notes` | `notes` | Direct mapping |
| `cloudSubnetID` | `cloud_subnet_id` | Rename to snake_case |
| `EC2AvailabilityZone` | `ec2_availability_zone` | Rename to snake_case |

**Fields Excluded from Schema**:
- `extra_values`: Always null in API response
- `disableAutomaticExports`: Internal BCM setting, low user value
- `excludeFromSearchDomain`: Internal DNS setting, low user value
- `lockDownDhcpd`: Internal DHCP setting, low user value
- `layer3ecmp`: Advanced L3 feature, low usage
- `layer3splitStaticRoute`: Advanced L3 feature, low usage

## Null Handling Strategy

**Helper Function Behavior**:
- `getStringValue()`: Converts null → `types.StringNull()`, empty string → `types.StringValue("")`
- `getBoolValue()`: Converts null → `types.BoolNull()`, false → `types.BoolValue(false)`
- `getInt64Value()`: Converts null → `types.Int64Null()`, 0 → `types.Int64Value(0)`

**Special Cases**:
- `gateway: "0.0.0.0"` → Keep as string value (not converted to null)
- `revision: ""` → Keep as empty string (not null)
- `childType: ""` → Keep as empty string (not null)
- `ipv6BaseAddress: "::0"` → Keep as string value (indicates IPv6 not configured)

## Performance Considerations

**Expected Dataset Size**: 1-100 networks per BCM cluster

**Memory Footprint**:
- Small network object (~2KB per network)
- 100 networks = ~200KB in memory
- Negligible impact on provider performance

**API Call Frequency**:
- Single API call per data source read
- No pagination needed (all networks returned in one response)
- Typical response time: <1 second

**Client-Side Filtering Performance**:
- O(n) where n = number of networks
- String comparison operations are fast for small datasets
- Expected filtering time: <1ms for 100 networks

## Testing Implications

### Test Data Requirements

**Minimum Test Data** (from research.md):
- ✅ At least 3 networks available in BCM cluster
- ✅ Network with name containing "management" (managementnet)
- ✅ Network with name containing "internal" (internalnet)
- ✅ Network with DHCP enabled (managementnet, internalnet)
- ✅ Network with DHCP disabled (globalnet)

### Test Scenarios

1. **Basic Read**: Retrieve all 3 networks, verify attributes populated
2. **Name Filter**: Filter by "management", expect 1 match
3. **DHCP Filter**: Filter by dhcp_enabled=true, expect 2 matches
4. **No Match**: Filter by nonsensical pattern, expect empty list (no error)

### Test Independence

**Read-Only Strategy**:
- All tests use existing networks in BCM cluster
- No create/destroy operations needed
- Tests can run in parallel (no shared state)
- No cleanup required (data source is idempotent)

## Future Enhancements (Out of Scope)

1. **Advanced Filtering**: Regex patterns, IP range matching, CIDR notation
2. **Single Network Lookup**: Data source argument to retrieve one network by name/UUID
3. **Network Resource**: Create, update, delete network configurations
4. **Network Interface Mapping**: Include which nodes use each network
5. **IPv6 Full Support**: Dedicated IPv6 attributes and filtering
6. **Performance Optimization**: Caching, pagination for large deployments
