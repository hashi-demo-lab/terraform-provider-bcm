# Data Model: BCM Network Resource

**Feature**: bcm_cmnet_network
**Date**: 2025-11-23
**Source**: Research findings from BCM CMNet API exploration

## Network Entity Schema

### Field Mappings (Terraform ↔ BCM API)

| Terraform Attribute | BCM API Field | Type | Category | Default | Validation |
|---------------------|---------------|------|----------|---------|------------|
| id | uuid | string | Computed | BCM-assigned | None |
| uuid | uuid | string | Computed | BCM-assigned | None |
| name | name | string | Required | N/A | Unique within cluster |
| subnet | N/A | string | Optional | null | CIDR notation regex |
| base_address | baseAddress | string | Computed | Parsed from subnet | IP address |
| netmask_bits | netmaskBits | int64 | Computed | Parsed from subnet | 0-32 |
| gateway | gateway | string | Optional | null | IP address |
| network_type | type | string | Computed | BCM-assigned | None |
| mtu | mtu | int64 | Optional | 1500 | 68-9000 |
| domain_name | domainName | string | Optional | "cluster.local" | Non-empty |
| dhcp_enabled | N/A | bool | Computed | Derived | None |
| dhcp_range_start | dynamicRangeStart | string | Optional | null | IP address |
| dhcp_range_end | dynamicRangeEnd | string | Optional | null | IP address |
| management | management | bool | Computed | BCM-assigned | None |
| bootable | bootable | bool | Computed | BCM-assigned | None |
| notes | notes | string | Optional | "" | None |
| base_type | baseType | string | Computed | "Network" | None |
| child_type | childType | string | Computed | "" | None |
| revision | revision | string | Computed | BCM-assigned | None |
| modified | modified | bool | Computed | BCM-assigned | None |
| to_be_removed | to_be_removed | bool | Computed | BCM-assigned | None |

### Attribute Categories

**Required (user must provide)**:
- `name` - Unique network identifier

**Optional (user can configure)**:
- `subnet` - Network CIDR (e.g., "10.0.1.0/24")
- `gateway` - Gateway IP address
- `mtu` - Maximum transmission unit (default: 1500)
- `domain_name` - DNS domain (default: "cluster.local")
- `dhcp_range_start` - DHCP pool start IP
- `dhcp_range_end` - DHCP pool end IP
- `notes` - User notes

**Computed (provider manages)**:
- `id` - Terraform resource ID (= uuid)
- `uuid` - BCM-assigned unique identifier
- `base_address` - Network base IP (from subnet)
- `netmask_bits` - CIDR mask bits (from subnet)
- `dhcp_enabled` - Derived from dhcp_range_start/end
- `network_type` - BCM-assigned network type
- `management` - BCM management network flag
- `bootable` - BCM bootable network flag
- `base_type`, `child_type`, `revision`, `modified`, `to_be_removed` - BCM entity metadata

## Computed Field Logic

### CIDR Parsing (subnet → base_address + netmask_bits)

**On Plan (Create/Update)**:
```go
if !data.Subnet.IsNull() && !data.Subnet.IsUnknown() {
    ip, ipnet, err := net.ParseCIDR(data.Subnet.ValueString())
    if err != nil {
        return nil, fmt.Errorf("invalid CIDR: %w", err)
    }
    entity["baseAddress"] = ip.String()
    maskBits, _ := ipnet.Mask.Size()
    entity["netmaskBits"] = maskBits
}
```

**On Read (API → State)**:
```go
if !baseAddr.IsNull() && !netmaskBits.IsNull() {
    subnet := fmt.Sprintf("%s/%d", baseAddr.ValueString(), netmaskBits.ValueInt64())
    data.Subnet = types.StringValue(subnet)
}
```

### DHCP Enabled Derivation

**Logic**:
```go
func isDHCPEnabled(rangeStart, rangeEnd string) bool {
    return rangeStart != "" && rangeStart != "0.0.0.0" &&
           rangeEnd != "" && rangeEnd != "0.0.0.0"
}
```

**Application**:
```go
dhcpEnabled := isDHCPEnabled(
    data.DHCPRangeStart.ValueString(),
    data.DHCPRangeEnd.ValueString(),
)
data.DHCPEnabled = types.BoolValue(dhcpEnabled)
```

### Domain Name Default

**Logic**:
```go
domainName := data.DomainName.ValueString()
if domainName == "" {
    domainName = "cluster.local"
}
entity["domainName"] = domainName
```

## BCM Entity Structure

### Create Operation Entity

```json
{
  "name": "compute-network",
  "baseAddress": "10.0.1.0",
  "netmaskBits": 24,
  "gateway": "10.0.1.1",
  "mtu": 1500,
  "domainName": "cluster.local",
  "dynamicRangeStart": "10.0.1.100",
  "dynamicRangeEnd": "10.0.1.200",
  "notes": "Created via Terraform",
  "baseType": "Network",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

**NOTE**: `uuid` must NOT be included for creates (BCM auto-generates)

### Update Operation Entity

```json
{
  "uuid": "21b20743-d055-42c6-b03c-583c0c061e2e",
  "revision": "type3",
  "name": "compute-network",
  "baseAddress": "10.0.1.0",
  "netmaskBits": 24,
  "gateway": "10.0.1.1",
  "mtu": 9000,
  "domainName": "cluster.local",
  "dynamicRangeStart": "10.0.1.100",
  "dynamicRangeEnd": "10.0.1.200",
  "notes": "Updated via Terraform",
  "baseType": "Network",
  "childType": "",
  "modified": true,
  "to_be_removed": false
}
```

**NOTE**: `uuid` and `revision` REQUIRED for updates

## State Management

### Terraform State Structure

```hcl
resource "bcm_cmnet_network" "example" {
  # User-configured
  name              = "compute-network"
  subnet            = "10.0.1.0/24"
  gateway           = "10.0.1.1"
  mtu               = 9000
  domain_name       = "cluster.local"
  dhcp_range_start  = "10.0.1.100"
  dhcp_range_end    = "10.0.1.200"
  notes             = "Created via Terraform"

  # Provider-managed (computed)
  id                = "21b20743-d055-42c6-b03c-583c0c061e2e"
  uuid              = "21b20743-d055-42c6-b03c-583c0c061e2e"
  base_address      = "10.0.1.0"
  netmask_bits      = 24
  dhcp_enabled      = true
  network_type      = "INTERNAL"
  management        = false
  bootable          = true
  base_type         = "Network"
  child_type        = ""
  revision          = "type3"
  modified          = false
  to_be_removed     = false
}
```

## Field Validation Rules

### Name
- **Type**: string
- **Required**: true
- **Validation**: Must be unique within BCM cluster
- **Pattern**: No special validation (BCM enforces uniqueness)

### Subnet
- **Type**: string
- **Required**: false
- **Validation**: CIDR notation regex
- **Pattern**: `^(\d{1,3}\.){3}\d{1,3}/\d{1,2}$`
- **Example**: "10.0.1.0/24", "192.168.0.0/16"

### Gateway
- **Type**: string
- **Required**: false
- **Validation**: Must be valid IP address within subnet
- **Pattern**: `^(\d{1,3}\.){3}\d{1,3}$`

### MTU
- **Type**: int64
- **Required**: false
- **Default**: 1500
- **Validation**: Range 68-9000 (standard jumbo frames)

### Domain Name
- **Type**: string
- **Required**: false (but BCM requires it - provider uses default)
- **Default**: "cluster.local"
- **Validation**: Non-empty string

### DHCP Range Start/End
- **Type**: string
- **Required**: false (both or neither)
- **Validation**: Must be valid IP addresses
- **Note**: Setting both enables DHCP (dhcp_enabled = true)

## Drift Detection Scenarios

### External MTU Change
```
1. Terraform creates network with mtu=1500
2. Admin changes MTU to 9000 via BCM UI
3. Terraform plan detects drift
4. Terraform apply restores mtu=1500
```

### External DHCP Configuration Change
```
1. Terraform creates network without DHCP (dhcp_enabled=false)
2. Admin enables DHCP via BCM UI (sets dynamicRangeStart/End)
3. Terraform plan detects drift (dhcp_enabled computed as true)
4. Terraform apply removes DHCP configuration
```

## Import Scenarios

### Import by UUID
```bash
terraform import bcm_cmnet_network.imported 21b20743-d055-42c6-b03c-583c0c061e2e
```

**Implementation**:
```go
func (r *CMNetNetworkResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
```

## Error Handling

### Create Errors
- **Name conflict**: 409 "Network with name 'X' already exists"
- **Invalid CIDR**: Client-side validation error from net.ParseCIDR
- **Missing domain_name**: Use default "cluster.local"

### Update Errors
- **Revision conflict**: 409 "Concurrent modification detected" (retry logic needed)
- **UUID not found**: 404 "Network not found" (trigger RemoveResource for drift)

### Delete Errors
- **Active assignments**: 409 "Network has active node assignments" (actionable error message)
- **UUID not found**: 404 ignored (already deleted)

---

**Ready for**: Phase 2 TDD implementation with complete data model
