# Data Model: BCM CMPart Entity Info Data Source

**Feature**: `095-cmpart-entity-info`
**Date**: 2025-11-27

## Entity Definitions

### EntityInfo

Represents basic metadata about a BCM entity returned by the `getBasicEntityInformation` API.

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `name` | string | Yes | Human-readable entity name (mapped from API `resolveName`) |
| `type` | string | Yes | Entity type classification (e.g., "SoftwareImage", "Category") |
| `uuid` | string | Yes | Unique identifier in UUID format |

**Validation Rules**:
- `name`: May be empty string if API returns empty `resolveName`
- `type`: Must be non-empty (API always returns type)
- `uuid`: Must be valid UUID format (API always returns uuid)

**State Transitions**: N/A (read-only entity)

---

## Terraform Schema

### Data Source: `bcm_cmpart_entity_info`

#### Input Attributes

| Attribute | Type | Required | Default | Description |
|-----------|------|----------|---------|-------------|
| `type` | string | No | null | Filter by entity type (case-sensitive exact match) |
| `name_pattern` | string | No | null | Filter by name using glob pattern (case-insensitive) |

#### Computed Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Data source identifier for state management |
| `entities` | list(object) | List of matching entities |

#### Entity Object Schema

| Attribute | Type | Description |
|-----------|------|-------------|
| `name` | string | Entity name (from API `resolveName`) |
| `type` | string | Entity type classification |
| `uuid` | string | Entity unique identifier |

---

## Go Type Definitions

### Data Source Model

```go
// CMPartEntityInfoDataSourceModel describes the data source data model.
type CMPartEntityInfoDataSourceModel struct {
    ID          types.String      `tfsdk:"id"`
    Type        types.String      `tfsdk:"type"`
    NamePattern types.String      `tfsdk:"name_pattern"`
    Entities    []EntityInfoModel `tfsdk:"entities"`
}
```

### Entity Model

```go
// EntityInfoModel represents a BCM entity's basic information.
type EntityInfoModel struct {
    Name types.String `tfsdk:"name"`
    Type types.String `tfsdk:"type"`
    UUID types.String `tfsdk:"uuid"`
}
```

---

## API Response Mapping

### Request

```json
{
  "service": "cmpart",
  "call": "getBasicEntityInformation"
}
```

### Response Structure

```json
[
  {
    "resolveName": "default-image",
    "type": "SoftwareImage",
    "uuid": "8482c4e9-383c-43de-873f-8c54ee77ee74"
  },
  {
    "resolveName": "default",
    "type": "Category",
    "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  }
]
```

### Field Mapping

| API Field | Terraform Attribute | Transformation |
|-----------|---------------------|----------------|
| `resolveName` | `name` | Direct string copy |
| `type` | `type` | Direct string copy |
| `uuid` | `uuid` | Direct string copy |

---

## Filter Logic

### Type Filter

- **Match Type**: Case-sensitive exact match
- **Null Handling**: When null/unset, all entity types pass
- **Example**: `type = "SoftwareImage"` matches only entities where `type == "SoftwareImage"`

### Name Pattern Filter

- **Match Type**: Case-insensitive glob pattern
- **Wildcards**: `*` (any characters), `?` (single character)
- **Null Handling**: When null/unset, all entity names pass
- **Implementation**: Go `filepath.Match` with `strings.ToLower`
- **Example**: `name_pattern = "default*"` matches "default-image", "Default", "DEFAULT-NODE"

### Combined Filters

- **Logic**: AND (both conditions must match)
- **Example**: `type = "SoftwareImage"` AND `name_pattern = "default*"` matches only SoftwareImages with names starting with "default"

---

## Relationships

### No Direct Relationships

The `bcm_cmpart_entity_info` data source returns metadata only. It does not have direct relationships to other Terraform resources.

### Usage Pattern

Entities discovered via this data source can be referenced in other resources:

```hcl
# Discover software image UUID
data "bcm_cmpart_entity_info" "images" {
  type         = "SoftwareImage"
  name_pattern = "default*"
}

# Use discovered UUID in category resource
resource "bcm_cmdevice_category" "example" {
  name           = "my-category"
  software_image = data.bcm_cmpart_entity_info.images.entities[0].uuid
}
```

---

## Known Entity Types

The following entity types are commonly returned by the BCM API:

| Type | Description |
|------|-------------|
| `Category` | Node grouping categories |
| `SoftwareImage` | Operating system images |
| `Network` | Network configurations |
| `HeadNode` | Cluster head nodes |
| `PhysicalNode` | Physical compute nodes |
| `Partition` | Resource partitions |
| `Rack` | Physical rack definitions |
| `ConfigurationOverlay` | Configuration overlays |
| `FSPart` | Filesystem partitions |
| `Profile` | Node profiles |
| `Role` | Role definitions (various subtypes) |
| `KubeCluster` | Kubernetes clusters |
| `SlurmWlmCluster` | Slurm workload manager clusters |
| `MonitoringMeasurableMetric` | Monitoring metrics |
| `PrometheusQuery` | Prometheus query definitions |

Note: BCM supports 60+ entity types. This list shows common types; the API may return additional types not listed here.
