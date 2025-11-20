# Feature Specification: BCM CMDevice Nodes Data Source

**Feature Branch**: `002-bcm-cmdevice-nodes-datasource`
**Date**: 2025-11-20
**Status**: Planning

## Overview

Implement a Terraform data source for querying BCM (Bright Cluster Manager) cluster nodes via the CMDevice API. This data source enables infrastructure automation workflows to discover and reference node configurations including network interfaces, roles, and hardware settings.

## Problem Statement

Terraform users managing BCM clusters need to:
- Query available compute nodes for dynamic infrastructure provisioning
- Filter nodes by type (PhysicalNode, HeadNode, ComputeNode, CloudNode)
- Access node network configuration for integration with other infrastructure
- Reference node roles and categories for policy-driven deployments
- Build dynamic inventories for configuration management tools

Currently, there is no Terraform provider capability to programmatically query BCM node inventory.

## Requirements

### Functional Requirements

1. **Node Discovery**
   - Query all nodes in the BCM cluster
   - Return comprehensive node metadata including UUID, hostname, MAC address
   - Support nested data structures (interfaces, roles)

2. **Filtering Capabilities**
   - Filter by node type (childType): PhysicalNode, HeadNode, ComputeNode, CloudNode
   - Filter by category UUID
   - Filter by hostname pattern (substring match)

3. **Network Configuration**
   - Expose network interfaces with IP addresses, MAC addresses, and network UUIDs
   - Include interface types (Physical, Bond, Bridge, VLAN)
   - Provide interface boot and startup configuration

4. **Role Information**
   - List assigned roles (HeadNodeRole, ComputeRole, StorageRole, etc.)
   - Include role configuration flags (addServices)

5. **Management Data**
   - Power control type (none, ipmi, redfish, pdu)
   - Provisioning transport method
   - Authentication service type
   - Category and partition associations

### Non-Functional Requirements

1. **Performance**
   - Single API call to retrieve all nodes (no N+1 queries - means no per-node follow-up API calls)
   - Client-side filtering to minimize BCM API load
   - Response caching handled by Terraform's built-in data source caching (automatic within single `terraform plan/apply` execution)

2. **Reliability**
   - Comprehensive error handling with actionable messages
   - Null-safe field extraction for optional API fields
   - Validation of API response structure

3. **Usability**
   - Clear attribute names following Terraform conventions
   - Detailed schema descriptions for documentation generation
   - Practical examples for common use cases

4. **Maintainability**
   - Follow existing provider patterns (see bcm_cmpart_softwareimages)
   - Use helper functions for type conversion
   - Comprehensive acceptance tests

## API Specification

### Endpoint

**URL**: `POST https://172.21.15.254:8081/json`

**Authentication**: Cookie-based session (cm-login-token)

### Request Format

```json
{
  "service": "cmdevice",
  "call": "getNodes"
}
```

### Response Format

Array of Device objects:

```json
[
  {
    "baseType": "Device",
    "childType": "PhysicalNode",
    "uuid": "2870c0b0-6fda-4026-9b8f-28be4c372fee",
    "hostname": "node002",
    "mac": "00:00:00:00:00:00",
    "creationTime": 1763617980,
    "interfaces": [
      {
        "baseType": "NetworkInterface",
        "childType": "NetworkPhysicalInterface",
        "name": "BOOTIF",
        "mac": "00:00:00:00:00:00",
        "ip": "172.21.15.254",
        "ipv6Ip": "::0",
        "dhcp": false,
        "network": "network-uuid",
        "cardtype": "Ethernet",
        "bootable": false
      }
    ],
    "roles": [
      {
        "baseType": "Role",
        "childType": "HeadNodeRole",
        "name": "headnode",
        "uuid": "role-uuid",
        "addServices": true
      }
    ],
    "category": "0ae6d733-3015-4479-bfab-ce2d237a2809",
    "partition": "partition-uuid",
    "powerControl": "none",
    "authenticationService": "CATEGORY",
    "provisioningTransport": "RSYNCDAEMON",
    "modified": false,
    "to_be_removed": false
  }
]
```

## Terraform Usage Examples

### Example 1: Query All Nodes

```hcl
data "bcm_cmdevice_nodes" "all" {}

output "all_node_hostnames" {
  value = [for node in data.bcm_cmdevice_nodes.all.nodes : node.hostname]
}
```

### Example 2: Filter by Node Type

```hcl
data "bcm_cmdevice_nodes" "compute_nodes" {
  filter {
    node_type = "PhysicalNode"
  }
}

output "compute_node_ips" {
  value = {
    for node in data.bcm_cmdevice_nodes.compute_nodes.nodes :
    node.hostname => node.interfaces[0].ip
  }
}
```

### Example 3: Filter by Category

```hcl
data "bcm_cmdevice_nodes" "category_nodes" {
  filter {
    category_uuid = "0ae6d733-3015-4479-bfab-ce2d237a2809"
  }
}

output "category_node_count" {
  value = length(data.bcm_cmdevice_nodes.category_nodes.nodes)
}
```

### Example 4: Dynamic Inventory

```hcl
data "bcm_cmdevice_nodes" "all" {}

locals {
  compute_inventory = {
    for node in data.bcm_cmdevice_nodes.all.nodes :
    node.hostname => {
      uuid         = node.uuid
      ip           = length(node.interfaces) > 0 ? node.interfaces[0].ip : null
      mac          = node.mac
      node_type    = node.child_type
      power_control = node.power_control
      roles        = [for role in node.roles : role.name]
    }
  }
}

output "compute_inventory" {
  value = local.compute_inventory
}
```

## Schema Design

### Data Source: bcm_cmdevice_nodes

**Root Attributes:**
- `id` (Computed, String) - Placeholder identifier
- `filter` (Optional, Block) - Filtering criteria
  - `node_type` (Optional, String) - Filter by childType
  - `category_uuid` (Optional, String) - Filter by category UUID
  - `hostname_pattern` (Optional, String) - Filter by hostname substring
- `nodes` (Computed, List of Object) - List of nodes

**Node Object Attributes:**

**Identity:**
- `id` (String) - Node UUID (same as uuid)
- `uuid` (String) - Unique node identifier
- `hostname` (String) - Node hostname
- `base_type` (String) - Always "Device"
- `child_type` (String) - Node type (PhysicalNode, HeadNode, etc.)

**Network:**
- `mac` (String) - Primary MAC address
- `interfaces` (List of Object) - Network interfaces
  - `name` (String) - Interface name (e.g., ens33)
  - `mac` (String) - Interface MAC address
  - `ip` (String) - IPv4 address
  - `ipv6_ip` (String) - IPv6 address
  - `dhcp` (Bool) - DHCP enabled
  - `network` (String) - Network UUID
  - `base_type` (String) - Always "NetworkInterface"
  - `child_type` (String) - Interface type
  - `cardtype` (String) - Card type (Ethernet, InfiniBand)
  - `bootable` (Bool) - PXE boot capable
  - `start_if` (String) - Startup condition

**Roles:**
- `roles` (List of Object) - Assigned roles
  - `uuid` (String) - Role UUID
  - `name` (String) - Role name
  - `base_type` (String) - Always "Role"
  - `child_type` (String) - Role type
  - `add_services` (Bool) - Auto-add services flag

**Management:**
- `category` (String) - Category UUID
- `partition` (String) - Partition UUID
- `power_control` (String) - Power control type
- `authentication_service` (String) - Auth service type
- `provisioning_transport` (String) - Provisioning method
- `creation_time` (Int64) - Unix timestamp

**State:**
- `modified` (Bool) - Has unsaved changes
- `to_be_removed` (Bool) - Scheduled for deletion

## Success Criteria

1. Data source successfully queries BCM CMDevice API
2. All node attributes correctly mapped to Terraform schema
3. Filtering works for node_type, category_uuid, and hostname_pattern
4. Nested interfaces and roles properly structured
5. Acceptance tests pass with real BCM API
6. Documentation generated via tfplugindocs
7. Example configurations validated

## Out of Scope

- Node creation/modification (CRUD operations)
- Power management operations
- Node provisioning triggers
- Detailed hardware configuration (BIOS, BMC, GPU)
- Filesystem exports/mounts
- Static routes and advanced networking

## Dependencies

- Existing BCM provider with authentication (bcm_client.go)
- Terraform Plugin Framework v1.16.1
- Access to BCM cluster for acceptance testing

## Testing Strategy

### Unit Tests
- Helper function validation (type conversions)
- Schema validation
- Model mapping tests

### Acceptance Tests
- Query all nodes
- Filter by node type
- Filter by category
- Filter by hostname pattern
- Verify nested structures (interfaces, roles)
- Error handling (auth failure, network error)

### Integration Tests
- Validate against real BCM cluster
- Verify data accuracy with Python API client
- Test with various node configurations

## Documentation

Auto-generated via tfplugindocs including:
- Data source reference
- Attribute descriptions
- Usage examples
- Filter patterns

## Rollout Plan

1. Phase 0: Research (verify API responses, document edge cases)
2. Phase 1: Design (complete schema, contracts, quickstart)
3. Phase 2: TDD Implementation (RED-GREEN-REFACTOR)
   - RED: Write failing acceptance tests
   - GREEN: Minimal implementation
   - REFACTOR: Production-ready code
4. Phase 3: Documentation and validation

## References

- BCM API Documentation: `/workspace/sampleRest/CMDevice_Complete_Documentation.md`
- Device Entity Spec: `/workspace/sampleRest/DeviceEntity.md`
- API Response Sample: `/workspace/sampleRest/cmdevice_discovered_methods_20251120_175345.json`
- Existing Data Source: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
