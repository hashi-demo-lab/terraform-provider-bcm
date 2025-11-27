# Research: BCM CMPart Entity Info Data Source

**Feature**: `095-cmpart-entity-info`
**Date**: 2025-11-27

## Research Tasks

### RT-001: BCM API `getBasicEntityInformation` Response Structure

**Status**: Resolved

**Question**: What is the exact response structure from `cmpart.getBasicEntityInformation`?

**Finding**: Based on the BCM API documentation and sample REST calls in `/workspace/sampleRest/cmpart.http`, the API endpoint:

```json
{
  "service": "cmpart",
  "call": "getBasicEntityInformation"
}
```

Returns an array of entity objects. Each entity contains:

| Field | Type | Description |
|-------|------|-------------|
| `resolveName` | string | Human-readable entity name |
| `type` | string | Entity type classification (60+ types) |
| `uuid` | string | Unique identifier (UUID format) |

**Example Response** (inferred from API patterns):
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

**Decision**: Map `resolveName` to `name` in Terraform schema for clarity.

---

### RT-002: BCM Entity Types

**Status**: Resolved

**Question**: What entity types does BCM support?

**Finding**: Based on the spec and BCM API documentation, BCM supports 60+ entity types including:

**Core Infrastructure**:
- `Category` - Node groupings
- `SoftwareImage` - OS images for provisioning
- `Network` - Network configurations
- `Partition` - Resource partitions
- `Rack` - Physical rack definitions

**Devices**:
- `HeadNode` - Cluster head nodes
- `PhysicalNode` - Physical compute nodes
- `GpuNode` - GPU-equipped nodes

**Configuration**:
- `ConfigurationOverlay` - Configuration layers
- `FSPart` - Filesystem partitions
- `Profile` - Node profiles
- `Role` - Role definitions

**Workload Management**:
- `SlurmWlmCluster` - Slurm workload manager clusters
- `KubeCluster` - Kubernetes clusters

**Monitoring**:
- `MonitoringMeasurableMetric` - Metrics definitions
- `PrometheusQuery` - Prometheus queries

**Decision**: Type filter must be case-sensitive exact match since BCM API uses PascalCase consistently (e.g., `SoftwareImage`, not `softwareimage`).

---

### RT-003: Client-Side Filtering Best Practices

**Status**: Resolved

**Question**: How should client-side filtering be implemented for optimal performance and usability?

**Finding**: Analyzed existing data source implementations:

1. **`bcm_cmpart_softwareimages`** (`data_source_cmpart_softwareimages.go`):
   - Uses nested `filter` block with `name_pattern` (case-insensitive substring) and `category` (exact match)
   - Filters applied after full API response
   - AND logic for multiple filters

2. **`bcm_cmdevice_categories`** (`data_source_cmdevice_categories.go`):
   - Similar pattern with filter block

**Decision**: Use top-level filter attributes (not nested block) since only two filters exist. This is simpler for users:

```hcl
# Chosen approach - simpler
data "bcm_cmpart_entity_info" "images" {
  type = "SoftwareImage"
  name_pattern = "default*"
}

# Alternative (rejected) - more verbose
data "bcm_cmpart_entity_info" "images" {
  filter {
    type = "SoftwareImage"
    name_pattern = "default*"
  }
}
```

---

### RT-004: Glob Pattern Implementation

**Status**: Resolved

**Question**: How should glob patterns be implemented for the `name_pattern` filter?

**Finding**: Go stdlib provides `path/filepath.Match` which supports:
- `*` - matches any sequence of characters
- `?` - matches any single character
- `[abc]` - matches any character in set
- `[a-z]` - matches any character in range

**Limitations**:
- No `**` recursive matching (not needed for flat entity names)
- Pattern must match entire string, not substring

**Implementation**:
```go
import "path/filepath"

func matchesNamePattern(name, pattern string) bool {
    // Case-insensitive matching per FR-011
    matched, err := filepath.Match(
        strings.ToLower(pattern),
        strings.ToLower(name),
    )
    if err != nil {
        // Invalid pattern - treat as no match
        return false
    }
    return matched
}
```

**Decision**: Use `filepath.Match` with case normalization. Invalid patterns silently return no matches (defensive behavior).

---

### RT-005: Data Source ID Strategy

**Status**: Resolved

**Question**: How should the `id` attribute be generated for state management?

**Finding**: Terraform data sources require a unique `id` attribute. Existing patterns:

1. **Static placeholder**: `types.StringValue("placeholder")` (used in `bcm_cmpart_softwareimages`)
2. **Composite key**: Combination of filter values
3. **UUID generation**: Random UUID per read

**Decision**: Use composite key based on filter values for determinism:

```go
func generateDataSourceID(typeFilter, namePattern types.String) string {
    if typeFilter.IsNull() && namePattern.IsNull() {
        return "cmpart-entity-info:all"
    }

    typeVal := ""
    if !typeFilter.IsNull() {
        typeVal = typeFilter.ValueString()
    }

    patternVal := ""
    if !namePattern.IsNull() {
        patternVal = namePattern.ValueString()
    }

    return fmt.Sprintf("cmpart-entity-info:%s:%s", typeVal, patternVal)
}
```

This ensures:
- Consistent ID across terraform plan/apply cycles with same filters
- Different ID when filters change (triggers data source refresh)

---

### RT-006: Error Handling Patterns

**Status**: Resolved

**Question**: What error handling patterns should be used?

**Finding**: Analyzed `error_messages.go` and existing data sources:

1. **API call errors**: Include endpoint, service, call, and original error
2. **Parse errors**: Include response body (truncated) for debugging
3. **Empty results**: Return empty list, not error (matches spec FR-008)

**Error Message Template**:
```go
resp.Diagnostics.AddError(
    "Unable to Read BCM Entity Information",
    "An unexpected error occurred when calling the BCM API. "+
        "Check the endpoint, credentials, and network connectivity.\n\n"+
        "Error: "+err.Error(),
)
```

**Decision**: Follow established error message patterns from `error_messages.go`.

---

## Summary of Decisions

| Topic | Decision |
|-------|----------|
| API endpoint | `cmpart.getBasicEntityInformation` (no arguments) |
| Field mapping | `resolveName` -> `name`, `type` -> `type`, `uuid` -> `uuid` |
| Schema design | Top-level filter attributes (not nested block) |
| Type filter | Case-sensitive exact match |
| Name pattern | Case-insensitive glob using `filepath.Match` |
| ID generation | Composite key from filter values |
| Error handling | Follow `error_messages.go` patterns |
| Empty results | Return empty list (not error) |
