# Research Findings: BCM CMNet Network Resource

**Date**: 2025-11-23
**Feature**: bcm_cmnet_network resource
**Purpose**: Document API exploration results for Phase 1 design

## Executive Summary

This research validates the BCM CMNet API methods for network CRUD operations and confirms technical implementation approaches for the Terraform resource.

**Key Findings**:
- ✅ getNetwork(uuid) supports args parameter for efficient direct lookup
- ✅ CIDR parsing via Go net.ParseCIDR is validated and recommended
- ✅ BCM API structure follows standard entity pattern (baseType, childType, revision)
- ❌ VLAN support NOT found in BCM CMNet API (mark as OUT OF SCOPE)
- ⚠️  domainName is REQUIRED field for network creation

## 1. getNetwork Args Parameter Support

**Test**: Used existing network UUID to test direct lookup via args parameter

**Method**: `{"service": "cmnet", "call": "getNetwork", "args": [uuid]}`

**Result**: ✅ **SUPPORTED**

**Evidence**:
```json
Request: {
  "service": "cmnet",
  "call": "getNetwork",
  "args": ["21b20743-d055-42c6-b03c-583c0c061e2e"]
}

Response: {
  "name": "managementnet",
  "uuid": "21b20743-d055-42c6-b03c-583c0c061e2e",
  "baseAddress": "172.21.12.0",
  "netmaskBits": 22,
  ...
}
```

**Implementation Decision**:
- **Read operation**: Use `getNetwork(uuid)` with args parameter for efficient single-network lookup
- **Data source**: Use `getNetworks()` to list all networks, then filter client-side

**Code Pattern**:
```go
// Resource Read method
body, err := r.client.CallJSONRPC(ctx, "cmnet", "getNetwork", data.UUID.ValueString())
```

## 2. VLAN Field Mapping

**Test**: Inspected full network object structure for VLAN-related fields

**Method**: Retrieved existing network and searched all field names for "vlan" substring (case-insensitive)

**Result**: ❌ **NOT SUPPORTED**

**Evidence**:
- Full network object analyzed (42 fields inspected)
- No fields containing "vlan" found
- Fields checked: EC2AvailabilityZone, IPv6, allowAutosign, baseAddress, baseType, bootable, childType, cloudSubnetID, disableAutomaticExports, domainName, dynamicRangeEnd, dynamicRangeStart, excludeFromSearchDomain, extra_values, gateway, gatewayMetric, generateDNSZone, ipv6BaseAddress, ipv6Gateway, ipv6NetmaskBits, layer3, layer3ecmp, layer3route, layer3splitStaticRoute, lockDownDhcpd, management, modified, mtu, name, netmaskBits, notes, revision, searchDomainIndex, to_be_removed, type, uuid

**Implementation Decision**:
- **DO NOT** include VLAN attribute in resource schema
- **Mark as OUT OF SCOPE** in feature documentation
- **Remove** VLAN-related tasks from tasks.md (T028-T029, T092)

## 3. CIDR Parsing Strategy

**Test**: Validated Go standard library net.ParseCIDR function with various inputs

**Method**: Created Go test program with 6 test cases (valid and invalid CIDRs)

**Result**: ✅ **VALIDATED - Use net.ParseCIDR**

**Evidence**:
```go
// Test Cases
"10.0.1.0/24"    -> ✓ baseAddress: 10.0.1.0, netmaskBits: 24
"192.168.1.0/24" -> ✓ baseAddress: 192.168.1.0, netmaskBits: 24
"172.16.0.0/16"  -> ✓ baseAddress: 172.16.0.0, netmaskBits: 16
"10.0.0.0/8"     -> ✓ baseAddress: 10.0.0.0, netmaskBits: 8
"invalid-cidr"   -> ✗ Error: invalid CIDR address
"10.0.1.0/33"    -> ✗ Error: invalid CIDR address (mask too large)
```

**Implementation Decision**:
- **Use** `net.ParseCIDR` from Go standard library
- **Validation** automatic via ParseCIDR error handling
- **No** custom regex or third-party library needed

**Code Pattern**:
```go
// Parse CIDR (Plan -> API Entity)
func parseCIDR(cidr string) (baseAddress string, netmaskBits int, err error) {
    ip, ipnet, err := net.ParseCIDR(cidr)
    if err != nil {
        return "", 0, fmt.Errorf("invalid CIDR notation: %w", err)
    }
    baseAddress = ip.String()
    maskBits, _ := ipnet.Mask.Size()
    return baseAddress, maskBits, nil
}

// Reconstruct CIDR (API Response -> State)
func formatCIDR(baseAddress string, netmaskBits int64) string {
    return fmt.Sprintf("%s/%d", baseAddress, netmaskBits)
}
```

## 4. BCM CMNet API Methods

### API Service: cmnet

**Base URL**: `https://172.21.15.254:8081/json`

**Authentication**: Cookie-based (cm-login-token) - handled by BCMClient

### Method Signatures

#### 1. addNetwork (Create)

**Call**: `{"service": "cmnet", "call": "addNetwork", "args": [entity]}`

**Entity Structure**:
```json
{
  "name": "network-name",
  "baseAddress": "10.0.1.0",
  "netmaskBits": 24,
  "gateway": "10.0.1.1",
  "mtu": 1500,
  "domainName": "cluster.local",
  "dynamicRangeStart": "10.0.1.100",
  "dynamicRangeEnd": "10.0.1.200",
  "notes": "User notes",
  "baseType": "Network",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

**Required Fields**:
- `name` (string) - unique network name
- `domainName` (string) - DNS domain name
- `baseType` (string) - must be "Network"
- `childType` (string) - empty string for networks
- `modified` (boolean) - set to true for creates/updates
- `to_be_removed` (boolean) - set to false
- `revision` (string) - empty string for creates

**Optional Fields**:
- `baseAddress` (string) - network base IP
- `netmaskBits` (int) - CIDR mask bits
- `gateway` (string) - gateway IP address
- `mtu` (int) - MTU value (default: 1500)
- `dynamicRangeStart` (string) - DHCP pool start IP
- `dynamicRangeEnd` (string) - DHCP pool end IP
- `notes` (string) - user notes

**Response**: Created entity with UUID assigned

**Errors**:
- 409 if name already exists
- Validation errors if required fields missing

**NOTE**: UUID must NOT be included in create entity (BCM auto-generates)

#### 2. getNetwork (Read)

**Call**: `{"service": "cmnet", "call": "getNetwork", "args": [uuid]}`

**Parameters**:
- `uuid` (string) - network UUID

**Response**: Single network entity (full object with all fields)

**Errors**:
- 404 if UUID not found
- 400 if UUID is null/invalid

**Usage**:
- Resource Read operation (efficient direct lookup)
- Drift detection (compare API state with Terraform state)

#### 3. updateNetwork (Update)

**Call**: `{"service": "cmnet", "call": "updateNetwork", "args": [entity]}`

**Entity Requirements**:
- Must include `uuid` (from state)
- Must include `revision` (from Read response)
- Must set `modified = true`
- Include all fields (not just changed ones)

**Response**: Updated entity

**Errors**:
- 409 if revision conflict (concurrent modification)
- 404 if UUID not found

#### 4. removeNetwork (Delete)

**Call**: `{"service": "cmnet", "call": "removeNetwork", "args": [uuid]}`

**Parameters**:
- `uuid` (string) - network UUID to delete

**Response**: Empty on success

**Errors**:
- 409 if network has active assignments (nodes, etc.)
- 404 if UUID not found

**Force Parameter**: NOT required for basic deletion (removed from args)

## 5. BCM Entity Structure Pattern

**All BCM entities follow this pattern**:

```json
{
  "baseType": "Network",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "uuid": "auto-generated-on-create",
  ... entity-specific fields ...
}
```

**Rules**:
1. **Create**: Do NOT include `uuid` (BCM auto-generates)
2. **Update**: MUST include `uuid` and `revision` from Read response
3. **Delete**: Use `removeNetwork(uuid)` - no full entity required
4. **Read**: Use `getNetwork(uuid)` - returns full entity

## 6. Field Mappings (Terraform ↔ BCM API)

| Terraform Attribute (snake_case) | BCM API Field (camelCase) | Type   | Notes |
|-----------------------------------|---------------------------|--------|-------|
| id                                | uuid                      | string | Terraform convention |
| uuid                              | uuid                      | string | BCM primary key |
| name                              | name                      | string | Required, unique |
| subnet                            | N/A                       | string | User-facing CIDR |
| base_address                      | baseAddress               | string | Parsed from subnet |
| netmask_bits                      | netmaskBits               | int64  | Parsed from subnet |
| gateway                           | gateway                   | string | Optional IP |
| network_type                      | type                      | string | BCM-assigned |
| mtu                               | mtu                       | int64  | Default: 1500 |
| domain_name                       | domainName                | string | REQUIRED field |
| dhcp_enabled                      | N/A                       | bool   | Computed (derived) |
| dhcp_range_start                  | dynamicRangeStart         | string | Optional IP |
| dhcp_range_end                    | dynamicRangeEnd           | string | Optional IP |
| management                        | management                | bool   | BCM-assigned |
| bootable                          | bootable                  | bool   | BCM-assigned |
| notes                             | notes                     | string | Optional |
| base_type                         | baseType                  | string | Always "Network" |
| child_type                        | childType                 | string | Always "" |
| revision                          | revision                  | string | BCM versioning |
| modified                          | modified                  | bool   | BCM flag |
| to_be_removed                     | to_be_removed             | bool   | BCM flag |

## 7. DHCP Enabled Logic

**Derivation Rule**:
```
dhcp_enabled = (dhcp_range_start != null && dhcp_range_start != "0.0.0.0" &&
                dhcp_range_end != null && dhcp_range_end != "0.0.0.0")
```

**Implementation**:
```go
func isDHCPEnabled(rangeStart, rangeEnd string) bool {
    return rangeStart != "" && rangeStart != "0.0.0.0" &&
           rangeEnd != "" && rangeEnd != "0.0.0.0"
}
```

**BCM Behavior**:
- Networks without DHCP have `dynamicRangeStart: "0.0.0.0"` and `dynamicRangeEnd: "0.0.0.0"`
- Networks with DHCP have valid IP addresses in dynamic range fields
- `dhcp_enabled` is NOT a BCM API field - must be computed by provider

## 8. Domain Name Requirement

**Finding**: `domainName` is a REQUIRED field in BCM CMNet API

**Evidence**: Network creation without domainName results in validation error

**Implementation Decision**:
- **Schema**: Mark `domain_name` as Optional (user can choose to specify)
- **Default**: If not provided, use sensible default (e.g., "cluster.local")
- **Validation**: Must not be empty string

**Code Pattern**:
```go
// In buildNetworkAPIEntity
domainName := data.DomainName.ValueString()
if domainName == "" {
    domainName = "cluster.local" // Default value
}
entity["domainName"] = domainName
```

## 9. Recommendations for Implementation

### Schema Design

1. **Required Attributes**:
   - `name` (unique network name)

2. **Optional Attributes**:
   - `subnet` (CIDR notation, validates via stringvalidator.RegexMatches)
   - `gateway` (IP address within subnet)
   - `mtu` (default: 1500)
   - `domain_name` (default: "cluster.local" if not provided)
   - `dhcp_range_start` (DHCP pool start IP)
   - `dhcp_range_end` (DHCP pool end IP)
   - `notes` (user notes)

3. **Computed Attributes**:
   - `id` (same as uuid)
   - `uuid` (BCM-assigned)
   - `base_address` (parsed from subnet)
   - `netmask_bits` (parsed from subnet)
   - `dhcp_enabled` (derived from dhcp_range_start/end)
   - `network_type` (BCM-assigned)
   - `management` (BCM-assigned)
   - `bootable` (BCM-assigned)
   - `base_type`, `child_type`, `revision`, `modified`, `to_be_removed` (BCM entity fields)

4. **Excluded Attributes** (from research findings):
   - ❌ `vlan_id` (not supported by BCM API)
   - ❌ `force` (not needed for basic operations)

### CRUD Implementation

**Create**:
```go
entity := buildNetworkAPIEntity(ctx, plan)  // Excludes UUID
body, err := r.client.CallJSONRPC(ctx, "cmnet", "addNetwork", entity)
```

**Read**:
```go
body, err := r.client.CallJSONRPC(ctx, "cmnet", "getNetwork", state.UUID.ValueString())
if err != nil {
    resp.State.RemoveResource(ctx)  // Drift detection
    return
}
```

**Update**:
```go
entity := buildNetworkAPIEntity(ctx, plan)  // Includes UUID and revision
body, err := r.client.CallJSONRPC(ctx, "cmnet", "updateNetwork", entity)
```

**Delete**:
```go
_, err := r.client.CallJSONRPC(ctx, "cmnet", "removeNetwork", state.UUID.ValueString())
```

### Testing Approach

1. **Acceptance Tests**: Use existing test infrastructure (BCM cluster at 172.21.15.254:8081)
2. **Test Networks**: Create with unique names using generateUniqueTestName("test-network")
3. **Cleanup**: Use verifyResourceDeleted with exponential backoff
4. **Drift Detection**: Use PreConfig to modify network externally via BCM API

## 10. Open Questions Resolved

| Question | Answer | Impact |
|----------|--------|--------|
| Does getNetwork support args parameter? | ✅ YES | Use for efficient Read operations |
| Are VLAN fields available? | ❌ NO | Exclude from schema, mark OUT OF SCOPE |
| Can we use Go net.ParseCIDR? | ✅ YES | Use for CIDR parsing/validation |
| Is domainName required? | ✅ YES | Provide default if user doesn't specify |
| Do we need force parameter? | ⚠️ NO (for basic ops) | Omit from schema for MVP |

## 11. Files Generated

- `/workspace/sampleRest/test_network_getby_uuid.py` - Validates getNetwork args parameter support
- `/workspace/test_cidr_parsing.go` - Validates Go net.ParseCIDR functionality

## 12. Next Steps

**Phase 1 (Design)**:
1. Create `data-model.md` with field mappings table
2. Create `contracts/bcm-cmnet-api.md` with detailed API documentation
3. Create `quickstart.md` for developer onboarding
4. Update tasks.md to remove VLAN-related tasks (T028-T029, T092)

**Phase 2 (TDD RED)**:
5. Write all acceptance tests with modern terraform-plugin-testing patterns
6. Ensure tests FAIL before any implementation

**Phase 3-4 (TDD GREEN/REFACTOR)**:
7. Implement minimal resource skeleton
8. Implement full CRUD with BCM API integration
9. Implement CIDR parsing helpers
10. Implement DHCP derivation logic

---

**Research Complete**: ✅ All unknowns resolved, ready for Phase 1 design
