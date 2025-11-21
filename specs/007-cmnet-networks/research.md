# Phase 0 Research: BCM CMNet Networks API

**Date**: 2025-11-21
**API Endpoint**: `{"service": "cmnet", "call": "getNetworks"}`
**BCM Environment**: https://172.21.15.254:8081

## Research Summary

✅ **API Endpoint Verified**: The `cmnet.getNetworks` API call successfully returns network data
✅ **Response Format**: JSON array of network objects
✅ **Network Count**: 3 networks found in test BCM cluster
✅ **Helper Functions**: Confirmed available in `data_source_cmpart_softwareimages.go:401-431`
✅ **Test Prerequisites**: BCM cluster has sufficient networks for acceptance testing

## API Response Structure

### Complete JSON Response Sample

```json
[
  {
    "EC2AvailabilityZone": "",
    "IPv6": false,
    "allowAutosign": "AUTOMATIC",
    "baseAddress": "172.21.12.0",
    "baseType": "Network",
    "bootable": true,
    "childType": "",
    "cloudSubnetID": "",
    "disableAutomaticExports": false,
    "domainName": "hashicorp.local",
    "dynamicRangeEnd": "172.21.15.127",
    "dynamicRangeStart": "172.21.14.128",
    "excludeFromSearchDomain": false,
    "extra_values": null,
    "gateway": "172.21.12.1",
    "gatewayMetric": 0,
    "generateDNSZone": "BOTH",
    "ipv6BaseAddress": "::0",
    "ipv6Gateway": "::0",
    "ipv6NetmaskBits": 0,
    "layer3": false,
    "layer3ecmp": false,
    "layer3route": "NONE",
    "layer3splitStaticRoute": false,
    "lockDownDhcpd": false,
    "management": true,
    "modified": false,
    "mtu": 1500,
    "name": "managementnet",
    "netmaskBits": 22,
    "notes": "",
    "revision": "",
    "searchDomainIndex": 0,
    "to_be_removed": false,
    "type": "INTERNAL",
    "uuid": "21b20743-d055-42c6-b03c-583c0c061e2e"
  },
  {
    "EC2AvailabilityZone": "",
    "IPv6": false,
    "allowAutosign": "AUTOMATIC",
    "baseAddress": "0.0.0.0",
    "baseType": "Network",
    "bootable": false,
    "childType": "",
    "cloudSubnetID": "",
    "disableAutomaticExports": false,
    "domainName": "cm.cluster",
    "dynamicRangeEnd": "0.0.0.0",
    "dynamicRangeStart": "0.0.0.0",
    "excludeFromSearchDomain": false,
    "extra_values": null,
    "gateway": "0.0.0.0",
    "gatewayMetric": 0,
    "generateDNSZone": "BOTH",
    "ipv6BaseAddress": "::0",
    "ipv6Gateway": "::0",
    "ipv6NetmaskBits": 0,
    "layer3": false,
    "layer3ecmp": false,
    "layer3route": "NONE",
    "layer3splitStaticRoute": false,
    "lockDownDhcpd": false,
    "management": false,
    "modified": false,
    "mtu": 1500,
    "name": "globalnet",
    "netmaskBits": 0,
    "notes": "",
    "revision": "type3",
    "searchDomainIndex": 0,
    "to_be_removed": false,
    "type": "GLOBAL",
    "uuid": "51a3d9e5-855d-4499-9dd0-ba2a69080239"
  },
  {
    "EC2AvailabilityZone": "",
    "IPv6": false,
    "allowAutosign": "AUTOMATIC",
    "baseAddress": "10.141.0.0",
    "baseType": "Network",
    "bootable": true,
    "childType": "",
    "cloudSubnetID": "",
    "disableAutomaticExports": false,
    "domainName": "eth.cluster",
    "dynamicRangeEnd": "10.141.167.255",
    "dynamicRangeStart": "10.141.160.0",
    "excludeFromSearchDomain": false,
    "extra_values": null,
    "gateway": "0.0.0.0",
    "gatewayMetric": 0,
    "generateDNSZone": "BOTH",
    "ipv6BaseAddress": "::0",
    "ipv6Gateway": "::0",
    "ipv6NetmaskBits": 0,
    "layer3": false,
    "layer3ecmp": false,
    "layer3route": "NONE",
    "layer3splitStaticRoute": false,
    "lockDownDhcpd": false,
    "management": true,
    "modified": false,
    "mtu": 1500,
    "name": "internalnet",
    "netmaskBits": 16,
    "notes": "",
    "revision": "",
    "searchDomainIndex": 0,
    "to_be_removed": false,
    "type": "INTERNAL",
    "uuid": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  }
]
```

## Field Mapping Table

| API Field Name | Terraform Attribute | Data Type | Required | Notes |
|---|---|---|---|---|
| `uuid` | `id`, `uuid` | String | Yes | Unique network identifier (RFC 4122 UUID) |
| `name` | `name` | String | Yes | Human-readable network name |
| `baseType` | `base_type` | String | Yes | Always "Network" |
| `childType` | `child_type` | String | No | Network subtype (often empty string) |
| `baseAddress` | `base_address` | String | Yes | Network base IP (e.g., "172.21.12.0") |
| `netmaskBits` | `netmask_bits` | Int64 | Yes | CIDR netmask bits (e.g., 22, 16) |
| `gateway` | `gateway` | String | No | Gateway IP (may be "0.0.0.0" for no gateway) |
| `domainName` | `domain_name` | String | No | DNS domain name |
| `dynamicRangeStart` | `dynamic_range_start` | String | No | DHCP range start IP |
| `dynamicRangeEnd` | `dynamic_range_end` | String | No | DHCP range end IP |
| `management` | `management` | Bool | Yes | Is management network flag |
| `bootable` | `bootable` | Bool | Yes | Network supports PXE boot |
| `type` | `network_type` | String | Yes | Network type (INTERNAL, GLOBAL, etc.) |
| `mtu` | `mtu` | Int64 | Yes | Maximum transmission unit |
| `modified` | `modified` | Bool | Yes | Has unsaved changes |
| `to_be_removed` | `to_be_removed` | Bool | Yes | Scheduled for deletion |
| `revision` | `revision` | String | No | Revision string (may be empty) |
| `IPv6` | `ipv6_enabled` | Bool | Yes | IPv6 enabled flag |
| `ipv6BaseAddress` | `ipv6_base_address` | String | No | IPv6 network base address |
| `ipv6Gateway` | `ipv6_gateway` | String | No | IPv6 gateway address |
| `ipv6NetmaskBits` | `ipv6_netmask_bits` | Int64 | No | IPv6 CIDR netmask bits |
| `allowAutosign` | `allow_autosign` | String | Yes | Autosign policy (AUTOMATIC, etc.) |
| `generateDNSZone` | `generate_dns_zone` | String | Yes | DNS zone generation (BOTH, etc.) |
| `layer3` | `layer3` | Bool | Yes | Layer 3 networking enabled |
| `layer3route` | `layer3_route` | String | No | Layer 3 routing mode |
| `gatewayMetric` | `gateway_metric` | Int64 | No | Gateway routing metric |
| `searchDomainIndex` | `search_domain_index` | Int64 | No | Search domain priority |
| `notes` | `notes` | String | No | User notes (often empty) |
| `cloudSubnetID` | `cloud_subnet_id` | String | No | Cloud provider subnet ID |
| `EC2AvailabilityZone` | `ec2_availability_zone` | String | No | AWS EC2 availability zone |

## Key Findings

### 1. Response Format
- **Type**: JSON array
- **Count**: 3 networks in test environment
- **Structure**: Flat objects (no nested entities)

### 2. Field Naming Convention
- API uses **camelCase** (e.g., `baseAddress`, `netmaskBits`, `domainName`)
- Terraform attributes will use **snake_case** (e.g., `base_address`, `netmask_bits`, `domain_name`)

### 3. Network Types Observed
- `INTERNAL`: Standard internal networks with DHCP ranges
- `GLOBAL`: Special global network (baseAddress "0.0.0.0", netmaskBits 0)

### 4. DHCP Detection
**Important Discovery**: The API response does **NOT** include a direct "dhcp" boolean field as initially assumed in the specification.

DHCP enablement must be inferred from:
- Presence of non-zero DHCP range: `dynamicRangeStart != "0.0.0.0" && dynamicRangeEnd != "0.0.0.0"`

**Recommendation**: Add computed `dhcp_enabled` attribute in Terraform schema that evaluates:
```go
dhcpEnabled := network.DynamicRangeStart != "0.0.0.0" && network.DynamicRangeEnd != "0.0.0.0"
```

### 5. IPv6 Support
- All test networks have `IPv6: false`
- IPv6 fields present but unused (`ipv6BaseAddress: "::0"`)
- IPv6 attributes included in schema for future use

### 6. Special Cases
- Gateway `"0.0.0.0"` indicates no gateway
- Empty `childType` is common
- Empty `revision` string is valid
- `notes` field is often empty

### 7. Filter Test Data
Based on actual network names in BCM cluster:
- ✅ "management" pattern will match "managementnet"
- ✅ "internal" pattern will match "internalnet"
- ✅ "global" pattern will match "globalnet"
- ✅ DHCP filter can distinguish networks with DHCP ranges vs. without

## Helper Function Verification

Helper functions confirmed in `/workspace/internal/provider/data_source_cmpart_softwareimages.go`:

- **Line 401**: `func getStringValue(data map[string]interface{}, key string) types.String`
- **Line 410**: `func getBoolValue(data map[string]interface{}, key string) types.Bool`
- **Line 419**: `func getInt64Value(data map[string]interface{}, key string) types.Int64`

These functions are **package-private** (lowercase) and accessible to all files in the `provider` package.

**Usage Pattern**:
```go
name := getStringValue(networkData, "name")
management := getBoolValue(networkData, "management")
mtu := getInt64Value(networkData, "mtu")
```

## Pattern Analysis

### Existing Data Source Patterns

**From `data_source_cmdevice_nodes.go`**:
- List all entities, then client-side filter
- Use `CallJSONRPC(ctx, "cmdevice", "getNodes")`
- Map API response to Terraform models
- Set placeholder ID for data source

**From `data_source_cmpart_softwareimages.go`**:
- Use helper functions for null-safe field extraction
- Handle nested objects (modules array)
- Client-side filtering with substring matching

**Pattern to Follow**:
1. Call `client.CallJSONRPC(ctx, "cmnet", "getNetworks")`
2. Parse JSON array response
3. Map each network using helper functions
4. Apply client-side filters
5. Set state with filtered results

## Spec Update Required

**CRITICAL**: The specification assumed a `dhcp` boolean field in the API response. The actual API uses `dynamicRangeStart` and `dynamicRangeEnd` to indicate DHCP configuration.

**Recommendation**:
- Add computed `dhcp_enabled` attribute to Terraform schema
- Calculate from: `dynamicRangeStart != "0.0.0.0" && dynamicRangeEnd != "0.0.0.0"`
- Update filter logic to use this computed value
- Keep `dynamic_range_start` and `dynamic_range_end` as separate attributes

## Test Prerequisites

✅ **Prerequisite Met**: BCM cluster has 3 networks configured:
1. `managementnet` (INTERNAL, DHCP enabled, management network)
2. `globalnet` (GLOBAL, no DHCP, special network type)
3. `internalnet` (INTERNAL, DHCP enabled, management network)

**Test Strategy Validated**:
- Basic read test will return 3 networks
- Name filter "management" will match "managementnet"
- Name filter "internal" will match "internalnet"
- DHCP filter will differentiate networks based on dynamic range
- No-match filter can use pattern "nonexistent-xyz"

## Unknowns Resolved

All "NEEDS CLARIFICATION" items from the specification have been resolved through API exploration:

1. ✅ **Actual field names**: Documented in field mapping table
2. ✅ **Data types**: All fields typed (string, int, bool)
3. ✅ **DHCP representation**: Computed from `dynamicRangeStart`/`dynamicRangeEnd`
4. ✅ **Null handling**: Empty strings and "0.0.0.0" used for null values
5. ✅ **Network count**: 3 networks available for testing

## Next Steps

Phase 1 (Design Artifacts) can now proceed with:
- **data-model.md**: Complete network entity schema with all discovered fields
- **contracts/**: Update response contract with actual API structure
- **quickstart.md**: Developer guide with working API examples

## Appendix: Full API Field List

**All 36 fields returned by cmnet.getNetworks**:

1. EC2AvailabilityZone
2. IPv6
3. allowAutosign
4. baseAddress
5. baseType
6. bootable
7. childType
8. cloudSubnetID
9. disableAutomaticExports
10. domainName
11. dynamicRangeEnd
12. dynamicRangeStart
13. excludeFromSearchDomain
14. extra_values
15. gateway
16. gatewayMetric
17. generateDNSZone
18. ipv6BaseAddress
19. ipv6Gateway
20. ipv6NetmaskBits
21. layer3
22. layer3ecmp
23. layer3route
24. layer3splitStaticRoute
25. lockDownDhcpd
26. management
27. modified
28. mtu
29. name
30. netmaskBits
31. notes
32. revision
33. searchDomainIndex
34. to_be_removed
35. type
36. uuid

**Note**: Some fields are cloud-specific (EC2AvailabilityZone, cloudSubnetID) and may be empty for on-premises deployments.
