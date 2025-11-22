# Data Model: BCM Partitions Data Source

**Feature**: `bcm_cmpart_partitions`
**Date**: 2025-11-22
**Phase**: Phase 1 - Schema & Data Structure Design

## Overview

This document defines the complete Terraform schema and Go data structures for the `bcm_cmpart_partitions` data source. The design follows terraform-plugin-framework patterns and matches the established structure from `data_source_cmpart_softwareimages.go`.

## Terraform Schema Definition

### Data Source: `bcm_cmpart_partitions`

```hcl
data "bcm_cmpart_partitions" "example" {
  # Optional client-side filtering
  filter {
    name_pattern = "base"  # Case-insensitive substring match
  }
}

# Access partition data
output "partition_uuids" {
  value = [for p in data.bcm_cmpart_partitions.example.partitions : p.uuid]
}
```

### Root Attributes

| Attribute | Type | Required | Computed | Description |
|-----------|------|----------|----------|-------------|
| `id` | string | no | yes | Placeholder identifier for Terraform state (value: "placeholder") |
| `filter` | block | no | no | Client-side filtering configuration |
| `partitions` | list of objects | no | yes | List of partition objects retrieved from BCM API |

### Filter Block Attributes

| Attribute | Type | Required | Computed | Description |
|-----------|------|----------|----------|-------------|
| `name_pattern` | string | no | no | Case-insensitive substring match for partition name filtering |

**Filter Behavior**:
- Empty string `""` or null: No filtering (returns all partitions)
- Substring match: `"boot"` matches "boot-partition", "my-boot-image", "BOOTFS"
- Case-insensitive: `"PROD"` matches "production", "Prod-Base", "test-prod"
- No matches: Returns empty list (not an error)

### Partition Object Attributes

Each partition in the `partitions` list has the following attributes:

#### Identity & Core Fields

| Attribute | Type | Computed | Description |
|-----------|------|----------|-------------|
| `id` | string | yes | Partition identifier (same as uuid) |
| `uuid` | string | yes | Partition UUID (primary identifier) |
| `name` | string | yes | Human-readable partition name |
| `base_type` | string | yes | BCM entity base type (typically "Partition") |
| `child_type` | string | yes | BCM polymorphic type discriminator |

#### Configuration Fields

| Attribute | Type | Computed | Description |
|-----------|------|----------|-------------|
| `cluster_name` | string | yes | Display name for the cluster |
| `slave_name` | string | yes | Node naming prefix (e.g., "node" for node001, node002) |
| `slave_digits` | int64 | yes | Number of digits in node numbering (e.g., 3 for node001) |
| `relay_host` | string | yes | SMTP relay hostname for email notifications |
| `no_zero_conf` | bool | yes | Disable Zeroconf networking |

#### Email & Network Configuration

| Attribute | Type | Computed | Description |
|-----------|------|----------|-------------|
| `admin_email` | list of string | yes | Administrator email addresses |
| `time_servers` | list of string | yes | NTP time server addresses |
| `search_domains` | list of string | yes | DNS search domains |
| `name_servers` | list of string | yes | DNS name server addresses |

#### Metadata & State Fields

| Attribute | Type | Computed | Description |
|-----------|------|----------|-------------|
| `creation_time` | int64 | yes | Unix timestamp of partition creation |
| `revision` | string | yes | BCM revision tracking field |
| `modified` | bool | yes | Whether partition has uncommitted modifications |
| `to_be_removed` | bool | yes | Whether partition is marked for deletion |
| `notes` | string | yes | User notes or description |

**Total Attributes**: 20 computed fields per partition object

## Go Data Structures

### Main Data Source Struct

```go
// CMPartPartitionsDataSource is the data source implementation.
type CMPartPartitionsDataSource struct {
	client *BCMClient
}
```

**Purpose**: Holds the BCM API client for making JSON-RPC calls to the cmpart service.

**Methods**:
- `Metadata(ctx, req, resp)` - Returns data source type name
- `Schema(ctx, req, resp)` - Defines Terraform schema
- `Configure(ctx, req, resp)` - Receives BCMClient from provider
- `Read(ctx, req, resp)` - Fetches partitions from API and applies filtering

### Data Source Model

```go
// CMPartPartitionsDataSourceModel describes the data source data model.
type CMPartPartitionsDataSourceModel struct {
	ID         types.String          `tfsdk:"id"`
	Filter     *PartitionFilterModel `tfsdk:"filter"`
	Partitions []PartitionModel      `tfsdk:"partitions"`
}
```

**Field Mapping**:
- `ID` - Placeholder identifier (always "placeholder")
- `Filter` - Optional client-side filtering configuration
- `Partitions` - List of partition objects from API

**Usage**:
```go
var data CMPartPartitionsDataSourceModel
resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
```

### Filter Model

```go
// PartitionFilterModel describes the filter block for client-side filtering.
type PartitionFilterModel struct {
	NamePattern types.String `tfsdk:"name_pattern"`
}
```

**Field Details**:
- `NamePattern` - Optional case-insensitive substring match for partition name

**Filter Logic**:
```go
func applyPartitionFilters(partitions []map[string]interface{}, filter *PartitionFilterModel) []map[string]interface{} {
	if filter == nil {
		return partitions
	}

	filtered := partitions

	// Name pattern filtering
	if !filter.NamePattern.IsNull() && filter.NamePattern.ValueString() != "" {
		pattern := strings.ToLower(filter.NamePattern.ValueString())
		var temp []map[string]interface{}
		for _, partition := range filtered {
			name := strings.ToLower(getStringValue(partition, "name").ValueString())
			if strings.Contains(name, pattern) {
				temp = append(temp, partition)
			}
		}
		filtered = temp
	}

	return filtered
}
```

### Partition Model

```go
// PartitionModel represents a BCM partition with all fields.
type PartitionModel struct {
	// Identity fields
	ID        types.String `tfsdk:"id"`
	UUID      types.String `tfsdk:"uuid"`
	Name      types.String `tfsdk:"name"`
	BaseType  types.String `tfsdk:"base_type"`
	ChildType types.String `tfsdk:"child_type"`

	// Configuration fields
	ClusterName types.String `tfsdk:"cluster_name"`
	SlaveName   types.String `tfsdk:"slave_name"`
	SlaveDigits types.Int64  `tfsdk:"slave_digits"`
	RelayHost   types.String `tfsdk:"relay_host"`
	NoZeroConf  types.Bool   `tfsdk:"no_zero_conf"`

	// Email and network configuration (arrays)
	AdminEmail    types.List `tfsdk:"admin_email"`    // List of types.String
	TimeServers   types.List `tfsdk:"time_servers"`   // List of types.String
	SearchDomains types.List `tfsdk:"search_domains"` // List of types.String
	NameServers   types.List `tfsdk:"name_servers"`   // List of types.String

	// Metadata and state fields
	CreationTime types.Int64  `tfsdk:"creation_time"`
	Revision     types.String `tfsdk:"revision"`
	Modified     types.Bool   `tfsdk:"modified"`
	ToBeRemoved  types.Bool   `tfsdk:"to_be_removed"`
	Notes        types.String `tfsdk:"notes"`
}
```

**Field Mapping from API (camelCase → snake_case)**:

| Go Struct Field | Type | API Field | Helper Function |
|-----------------|------|-----------|-----------------|
| `ID` | types.String | `uuid` | `getStringValue()` |
| `UUID` | types.String | `uuid` | `getStringValue()` |
| `Name` | types.String | `name` | `getStringValue()` |
| `BaseType` | types.String | `baseType` | `getStringValue()` |
| `ChildType` | types.String | `childType` | `getStringValue()` |
| `ClusterName` | types.String | `clusterName` | `getStringValue()` |
| `SlaveName` | types.String | `slaveName` | `getStringValue()` |
| `SlaveDigits` | types.Int64 | `slaveDigits` | `getInt64Value()` |
| `RelayHost` | types.String | `relayHost` | `getStringValue()` |
| `NoZeroConf` | types.Bool | `noZeroConf` | `getBoolValue()` |
| `AdminEmail` | types.List | `adminEmail` | **NEW** `getStringListValue()` |
| `TimeServers` | types.List | `timeServers` | **NEW** `getStringListValue()` |
| `SearchDomains` | types.List | `searchDomains` | **NEW** `getStringListValue()` |
| `NameServers` | types.List | `nameServers` | **NEW** `getStringListValue()` |
| `CreationTime` | types.Int64 | `creationTime` | `getInt64Value()` |
| `Revision` | types.String | `revision` | `getStringValue()` |
| `Modified` | types.Bool | `modified` | `getBoolValue()` |
| `ToBeRemoved` | types.Bool | `to_be_removed` | `getBoolValue()` |
| `Notes` | types.String | `notes` | `getStringValue()` |

## Helper Functions

### Existing Helpers (Reused)

Located in existing data sources (e.g., `data_source_cmpart_softwareimages.go:399-431`):

```go
// getStringValue extracts a string value from API response map with null safety
func getStringValue(data map[string]interface{}, key string) types.String {
	if val, ok := data[key]; ok && val != nil {
		if str, ok := val.(string); ok {
			return types.StringValue(str)
		}
	}
	return types.StringNull()
}

// getBoolValue extracts a boolean value from API response map with null safety
func getBoolValue(data map[string]interface{}, key string) types.Bool {
	if val, ok := data[key]; ok && val != nil {
		if b, ok := val.(bool); ok {
			return types.BoolValue(b)
		}
	}
	return types.BoolNull()
}

// getInt64Value extracts an int64 value from API response map with null safety
// Handles both int64 and float64 types from JSON unmarshaling
func getInt64Value(data map[string]interface{}, key string) types.Int64 {
	if val, ok := data[key]; ok && val != nil {
		// Try float64 first (JSON numbers unmarshal to float64)
		if f, ok := val.(float64); ok {
			return types.Int64Value(int64(f))
		}
		// Try int64 direct
		if i, ok := val.(int64); ok {
			return types.Int64Value(i)
		}
		// Try int
		if i, ok := val.(int); ok {
			return types.Int64Value(int64(i))
		}
	}
	return types.Int64Null()
}
```

### New Helper Function (Required)

**Implementation**: Add to `data_source_cmpart_partitions.go` after the Read() method

```go
// getStringListValue extracts a string array from API response map with null safety
// Returns a types.List of types.String elements, or null List if missing/invalid
func getStringListValue(ctx context.Context, data map[string]interface{}, key string) types.List {
	if val, ok := data[key]; ok && val != nil {
		if arr, ok := val.([]interface{}); ok {
			// Convert []interface{} to []attr.Value
			elements := make([]attr.Value, 0, len(arr))
			for _, item := range arr {
				if str, ok := item.(string); ok {
					elements = append(elements, types.StringValue(str))
				} else {
					// Skip non-string elements (defensive)
					tflog.Warn(ctx, fmt.Sprintf("Skipping non-string element in array field %s", key))
				}
			}

			// Create List with StringType element type
			listValue, diags := types.ListValue(types.StringType, elements)
			if diags.HasError() {
				tflog.Error(ctx, fmt.Sprintf("Failed to create List for field %s: %v", key, diags))
				return types.ListNull(types.StringType)
			}
			return listValue
		}
	}
	return types.ListNull(types.StringType)
}
```

**Usage Example**:
```go
partition.AdminEmail = getStringListValue(ctx, partitionData, "adminEmail")
partition.TimeServers = getStringListValue(ctx, partitionData, "timeServers")
partition.SearchDomains = getStringListValue(ctx, partitionData, "searchDomains")
partition.NameServers = getStringListValue(ctx, partitionData, "nameServers")
```

**Required Imports**:
```go
import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)
```

**Error Handling**: Returns null List if:
- Key doesn't exist in map
- Value is null
- Value is not an array
- List creation fails

## Schema Implementation Pattern

### Schema() Method Structure

```go
func (d *CMPartPartitionsDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Fetches partition information from BCM CMPart service. Partitions define filesystem partitions referenced by software images via bootfs_part and fs_part UUID fields.",

		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:            true,
				MarkdownDescription: "Placeholder identifier for this data source",
			},
			"partitions": schema.ListNestedAttribute{
				Computed:            true,
				MarkdownDescription: "List of partitions retrieved from BCM cluster",
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						// 20 partition attributes defined here
						"uuid": schema.StringAttribute{
							Computed:            true,
							MarkdownDescription: "Partition UUID (primary identifier)",
						},
						// ... (remaining attributes)
					},
				},
			},
		},
		Blocks: map[string]schema.Block{
			"filter": schema.SingleNestedBlock{
				MarkdownDescription: "Client-side filtering criteria for partitions",
				Attributes: map[string]schema.Attribute{
					"name_pattern": schema.StringAttribute{
						Optional:            true,
						MarkdownDescription: "Case-insensitive substring match for partition name (e.g., 'boot' matches 'boot-partition')",
					},
				},
			},
		},
	}
}
```

### Read() Method Structure

```go
func (d *CMPartPartitionsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	// Step 1: Read configuration
	var data CMPartPartitionsDataSourceModel
	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Step 2: Call BCM API
	body, err := d.client.CallJSONRPC(ctx, "cmpart", "getPartitions")
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read BCM Partitions",
			fmt.Sprintf("Failed to call cmpart.getPartitions API: %v", err),
		)
		return
	}

	// Step 3: Parse JSON response
	var partitions []map[string]interface{}
	if err := json.Unmarshal(body, &partitions); err != nil {
		resp.Diagnostics.AddError(
			"Unable to Parse BCM Partitions Response",
			fmt.Sprintf("Failed to unmarshal JSON response: %v", err),
		)
		return
	}

	// Step 4: Apply client-side filtering
	filteredPartitions := applyPartitionFilters(partitions, data.Filter)

	// Step 5: Map to Terraform state
	data.Partitions = make([]PartitionModel, 0, len(filteredPartitions))
	for _, partitionData := range filteredPartitions {
		partition := PartitionModel{
			ID:            getStringValue(partitionData, "uuid"),
			UUID:          getStringValue(partitionData, "uuid"),
			Name:          getStringValue(partitionData, "name"),
			// ... (map all 20 fields)
		}
		data.Partitions = append(data.Partitions, partition)
	}

	// Step 6: Set computed ID
	data.ID = types.StringValue("placeholder")

	// Step 7: Save state
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

## Data Flow Diagram

```
User HCL Config
      |
      v
[Filter Block] ----> applyPartitionFilters()
      |                      |
      v                      v
BCM API Call -------> Filtered Partitions
      |                      |
      v                      v
JSON Response -------> []map[string]interface{}
      |                      |
      v                      v
getXxxValue() -------> []PartitionModel
      |                      |
      v                      v
Terraform State <---- Set(ctx, &data)
```

## Testing Considerations

### Test Data Requirements

**Minimum Test Coverage**:
1. Basic retrieval (no filter) - verify ID computed and partitions list exists
2. Filter by name pattern - verify only matching partitions returned
3. Filter with no matches - verify empty list (not error)
4. Computed fields - verify all 20 attributes have correct types

**Environment Portability**:
- Tests MUST NOT assume specific partition counts
- Tests MUST NOT assume specific partition names
- Tests MUST work on any BCM cluster with at least 1 partition

### State Check Examples

```go
// Modern terraform-plugin-testing v1.13.3+ patterns
ConfigStateChecks: []statecheck.StateCheck{
	// Verify ID computed
	statecheck.ExpectKnownValue(
		"data.bcm_cmpart_partitions.test",
		tfjsonpath.New("id"),
		knownvalue.NotNull(),
	),
	// Note: Cannot verify partition count without hardcoding cluster state
}
```

## Implementation Checklist

- [ ] Define CMPartPartitionsDataSource struct
- [ ] Define CMPartPartitionsDataSourceModel struct
- [ ] Define PartitionFilterModel struct
- [ ] Define PartitionModel struct with 20 attributes
- [ ] Implement NewCMPartPartitionsDataSource() factory
- [ ] Implement Metadata() method
- [ ] Implement Schema() method with all 20 partition attributes
- [ ] Implement Configure() method
- [ ] Implement Read() method with 7 steps
- [ ] Implement applyPartitionFilters() function
- [ ] Implement getStringListValue() helper function
- [ ] Register data source in provider.go DataSources() method

## References

- **Research**: `/workspace/specs/001-partitions-data-source/research.md`
- **Spec**: `/workspace/specs/001-partitions-data-source/spec.md`
- **Plan**: `/workspace/specs/001-partitions-data-source/plan.md`
- **Reference Implementation**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
- **Terraform Plugin Framework**: https://developer.hashicorp.com/terraform/plugin/framework
