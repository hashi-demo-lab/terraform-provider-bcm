# Data Model: BCM Kubernetes Cluster Resource

**Resource**: `bcm_cmkube_cluster`
**Date**: 2025-11-22
**Based on**: research.md Phase 0 findings

## Overview

This document defines the data model for the Kubernetes cluster resource, including Terraform schema attributes, BCM API field mappings, and validation rules.

## Primary Entity: KubeCluster

### Terraform Resource Schema

```hcl
resource "bcm_cmkube_cluster" "example" {
  # Required attributes
  name         = "my-cluster"           # string
  master_nodes = ["node-uuid-1"]        # list(string)

  # Optional attributes
  worker_nodes       = ["node-uuid-2"]  # list(string)
  management_network = "network-uuid"   # string
  version            = "1.28.0"         # string (semver)
  force              = false            # bool

  # Computed attributes (read-only)
  id            = "cluster-uuid"        # string
  uuid          = "cluster-uuid"        # string
  creation_time = 1700000000000         # int64
  revision_id   = 42                    # int64
}
```

### Attribute Specifications

| Terraform Attribute | Type | BCM API Field | Required | Computed | Default | Description |
|---------------------|------|---------------|----------|----------|---------|-------------|
| **id** | string | (same as uuid) | No | Yes | - | Resource identifier for Terraform |
| **uuid** | string | uuid | No | Yes | - | BCM-assigned cluster UUID |
| **name** | string | name | Yes | No | - | Cluster name (alphanumeric, hyphens, underscores) |
| **master_nodes** | list(string) | masterNodes | Yes | No | - | List of master node UUIDs (min 1) |
| **worker_nodes** | list(string) | workerNodes | No | No | [] | List of worker node UUIDs |
| **management_network** | string | managementNetwork | No | No | null | Management network UUID |
| **version** | string | version | No | No | null | Kubernetes version (semver format) |
| **force** | bool | (args parameter) | No | No | false | Bypass validation warnings |
| **creation_time** | int64 | creationTime | No | Yes | - | Cluster creation Unix timestamp (ms) |
| **revision_id** | int64 | revisionID | No | Yes | - | BCM revision ID for optimistic locking |

### Field Name Mappings (snake_case → camelCase)

Terraform uses snake_case for readability, BCM API uses camelCase:

| Terraform | BCM API | Conversion Example |
|-----------|---------|-------------------|
| master_nodes | masterNodes | `entity["masterNodes"] = masterNodes` |
| worker_nodes | workerNodes | `entity["workerNodes"] = workerNodes` |
| management_network | managementNetwork | `entity["managementNetwork"] = managementNetwork` |
| creation_time | creationTime | `model.CreationTime = getInt64Value(data, "creationTime")` |
| revision_id | revisionID | `model.RevisionID = getInt64Value(data, "revisionID")` |

### BCM Entity Structure

Complete BCM KubeCluster entity for API calls:

```json
{
  "baseType": "KubeCluster",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "uuid": "550e8400-e29b-41d4-a716-446655440000",
  "name": "production-cluster",
  "masterNodes": [
    "650e8400-e29b-41d4-a716-446655440001"
  ],
  "workerNodes": [
    "750e8400-e29b-41d4-a716-446655440002",
    "750e8400-e29b-41d4-a716-446655440003"
  ],
  "managementNetwork": "850e8400-e29b-41d4-a716-446655440004",
  "version": "1.28.0",
  "creationTime": 1700000000000,
  "revisionID": 42
}
```

**Metadata Fields** (required for create/update):
- **baseType**: Always "KubeCluster"
- **childType**: Always "" (empty string)
- **modified**: true for create/update operations
- **to_be_removed**: false (true only in batch delete)
- **revision**: "" for create, current revision for update
- **uuid**: Omitted for create, required for update

## Validation Rules

### Schema-Level Validation

Enforced by Terraform during plan phase:

#### name Attribute
- **Validator**: `stringvalidator.RegexMatches(regexp.MustCompile(^[a-zA-Z0-9_-]+$))`
- **Valid Examples**: "my-cluster", "prod_k8s_01", "test-cluster-123"
- **Invalid Examples**: "my cluster" (space), "cluster!" (special char), "" (empty)
- **Error Message**: "must contain only alphanumeric characters, hyphens, and underscores"

#### version Attribute
- **Validator**: `stringvalidator.RegexMatches(regexp.MustCompile(^\d+\.\d+\.\d+$))`
- **Valid Examples**: "1.28.0", "1.29.1", "1.27.5"
- **Invalid Examples**: "1.28" (missing patch), "v1.28.0" (prefix), "latest" (not semver)
- **Error Message**: "must be valid semver format (e.g., '1.28.0')"

#### master_nodes Attribute
- **Type Validation**: Must be list of strings
- **Minimum Items**: 1 (enforced by BCM API, not schema validator)
- **Item Format**: Each item should be valid UUID v4 format
- **Error Location**: BCM API returns error if nodes don't exist

#### worker_nodes Attribute
- **Type Validation**: Must be list of strings
- **Minimum Items**: 0 (empty list allowed)
- **Item Format**: Each item should be valid UUID v4 format

### API-Level Validation

Enforced by BCM during create/update operations:

| Validation | Check | Error if Fails |
|------------|-------|----------------|
| Node existence | All node UUIDs must exist in BCM | "Node 'uuid' not found" |
| Network existence | management_network UUID must exist | "Network 'uuid' not found" |
| Duplicate name | Cluster name may need to be unique | "Cluster 'name' already exists" |
| Version support | Kubernetes version must be supported | "Version 'x.y.z' not supported" |
| Resource capacity | Enough resources for cluster | "Insufficient capacity" |

## State Transitions

Cluster provisioning is asynchronous. BCM tracks internal states:

```
[User Applies Config]
         ↓
    [addKubeCluster API Call]
         ↓
    [UUID Returned Immediately]  ← Terraform saves state here
         ↓
    [BCM Internal: CREATING]
         ↓
    [BCM Internal: PROVISIONING]  (5-30 minutes)
         ↓
    [BCM Internal: READY]
```

**Terraform Behavior**:
- Create operation completes when UUID is returned (not when cluster is READY)
- Read operations can be called during provisioning (returns current state)
- Update operations accepted during provisioning (may queue until READY)
- Delete operations work at any state

**State Fields** (if exposed by BCM API):
- **state**: CREATING | PROVISIONING | READY | ERROR | DELETING
- **stateMessage**: Human-readable status message

Note: State fields are NOT included in MVP schema. Terraform treats cluster as created once UUID is assigned.

## Type Conversions

### Terraform → BCM (Create/Update)

```go
// String fields (direct assignment)
entity["name"] = model.Name.ValueString()

// Optional string fields (null check)
if !model.ManagementNetwork.IsNull() {
    entity["managementNetwork"] = model.ManagementNetwork.ValueString()
}

// List fields (convert types.List to []string)
var masterNodes []string
diags.Append(model.MasterNodes.ElementsAs(ctx, &masterNodes, false)...)
entity["masterNodes"] = masterNodes

// Optional list fields
if !model.WorkerNodes.IsNull() {
    var workerNodes []string
    diags.Append(model.WorkerNodes.ElementsAs(ctx, &workerNodes, false)...)
    entity["workerNodes"] = workerNodes
} else {
    entity["workerNodes"] = []string{}  // Empty array, not null
}

// Boolean fields
entity_force_param := model.Force.ValueBool()  // Passed as separate arg
```

### BCM → Terraform (Read)

```go
// String fields (using helper)
model.Name = getStringValue(data, "name")
model.ManagementNetwork = getStringValue(data, "managementNetwork")
model.Version = getStringValue(data, "version")

// List fields (manual conversion)
if masterNodes, ok := data["masterNodes"].([]interface{}); ok {
    elements := make([]attr.Value, len(masterNodes))
    for i, node := range masterNodes {
        elements[i] = types.StringValue(node.(string))
    }
    model.MasterNodes, _ = types.ListValue(types.StringType, elements)
}

// Optional list fields (handle null)
if workerNodes, ok := data["workerNodes"].([]interface{}); ok && len(workerNodes) > 0 {
    elements := make([]attr.Value, len(workerNodes))
    for i, node := range workerNodes {
        elements[i] = types.StringValue(node.(string))
    }
    model.WorkerNodes, _ = types.ListValue(types.StringType, elements)
} else {
    model.WorkerNodes = types.ListNull(types.StringType)
}

// Int64 fields (using helper)
model.CreationTime = getInt64Value(data, "creationTime")
model.RevisionID = getInt64Value(data, "revisionID")

// UUID and ID (always same value)
model.UUID = types.StringValue(data["uuid"].(string))
model.ID = model.UUID
```

## Null vs Empty vs Omitted

BCM API behavior for optional fields:

| Terraform Value | BCM API Behavior | Example |
|-----------------|------------------|---------|
| Not set (null) | Field omitted from entity OR sent as null | `"version": null` or field absent |
| Empty string | Sent as empty string "" | `"version": ""` |
| Empty list | Sent as empty array [] | `"workerNodes": []` |

**MVP Decision**:
- Null string fields → omit from entity
- Null list fields → send as empty array []
- This matches existing resource patterns

## Helper Functions

### getStringValue (Null-Safe String Extraction)

```go
func getStringValue(data map[string]interface{}, key string) types.String {
    if val, ok := data[key]; ok && val != nil {
        if str, ok := val.(string); ok && str != "" {
            return types.StringValue(str)
        }
    }
    return types.StringNull()
}
```

### getInt64Value (Null-Safe Int64 Extraction)

```go
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

Note: These helpers are already defined in `data_source_cmpart_softwareimages.go` and can be reused.

## Schema Example

```go
func (r *CMKubeClusterResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Manages a BCM Kubernetes cluster.",

        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed: true,
                MarkdownDescription: "Cluster identifier (same as uuid)",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "uuid": schema.StringAttribute{
                Computed: true,
                MarkdownDescription: "BCM-assigned cluster UUID",
                PlanModifiers: []planmodifier.String{
                    stringplanmodifier.UseStateForUnknown(),
                },
            },
            "name": schema.StringAttribute{
                Required: true,
                MarkdownDescription: "Cluster name",
                Validators: []validator.String{
                    stringvalidator.RegexMatches(
                        regexp.MustCompile(`^[a-zA-Z0-9_-]+$`),
                        "must contain only alphanumeric characters, hyphens, and underscores",
                    ),
                },
            },
            // ... additional attributes
        },
    }
}
```

## Testing Considerations

### Test Fixtures

Tests use dynamic node discovery, not hardcoded UUIDs:

```go
func getTestMasterNodeUUID(t *testing.T) string {
    client := createTestBCMClient(t)
    nodes := queryAvailableNodes(client)
    if len(nodes) < 1 {
        t.Fatal("Need at least 1 node for cluster test")
    }
    return nodes[0]["uuid"].(string)
}
```

### State Assertions

Modern test patterns use type-safe assertions:

```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "bcm_cmkube_cluster.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact("my-cluster"),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmkube_cluster.test",
        tfjsonpath.New("master_nodes"),
        knownvalue.ListSizeExact(1),
    ),
    statecheck.ExpectKnownValue(
        "bcm_cmkube_cluster.test",
        tfjsonpath.New("uuid"),
        knownvalue.NotNull(),
    ),
}
```

## Summary

This data model provides:
- ✅ Complete Terraform schema definition
- ✅ BCM API field mappings (snake_case ↔ camelCase)
- ✅ Validation rules at schema and API levels
- ✅ Type conversion patterns for CRUD operations
- ✅ Null handling strategy
- ✅ Test fixture patterns for portability

The design follows established BCM resource patterns while maintaining Terraform best practices for resource implementation.
