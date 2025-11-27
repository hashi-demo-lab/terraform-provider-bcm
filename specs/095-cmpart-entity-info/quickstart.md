# Quickstart: BCM CMPart Entity Info Data Source

**Feature**: `095-cmpart-entity-info`
**Date**: 2025-11-27

## Overview

The `bcm_cmpart_entity_info` data source retrieves entity metadata from BCM (Bright Cluster Manager) using the `cmpart.getBasicEntityInformation` API endpoint. Use this data source to discover BCM entities and obtain their UUIDs for reference in other Terraform resources.

## Basic Usage

### Retrieve All Entities

```hcl
data "bcm_cmpart_entity_info" "all" {}

output "total_entities" {
  value = length(data.bcm_cmpart_entity_info.all.entities)
}
```

### Filter by Entity Type

```hcl
# Get all software images
data "bcm_cmpart_entity_info" "images" {
  type = "SoftwareImage"
}

output "image_names" {
  value = [for e in data.bcm_cmpart_entity_info.images.entities : e.name]
}
```

### Filter by Name Pattern

```hcl
# Get entities with names starting with "default"
data "bcm_cmpart_entity_info" "defaults" {
  name_pattern = "default*"
}
```

### Combined Filters

```hcl
# Get software images with names containing "ubuntu"
data "bcm_cmpart_entity_info" "ubuntu_images" {
  type         = "SoftwareImage"
  name_pattern = "*ubuntu*"
}
```

## Common Use Cases

### 1. Lookup Entity UUID by Name

```hcl
data "bcm_cmpart_entity_info" "my_image" {
  type         = "SoftwareImage"
  name_pattern = "default-image"
}

# Use the UUID in another resource
resource "bcm_cmdevice_category" "example" {
  name           = "compute-nodes"
  software_image = data.bcm_cmpart_entity_info.my_image.entities[0].uuid
}
```

### 2. List All Categories

```hcl
data "bcm_cmpart_entity_info" "categories" {
  type = "Category"
}

output "category_list" {
  value = {
    for e in data.bcm_cmpart_entity_info.categories.entities :
    e.name => e.uuid
  }
}
```

### 3. Discovery - Find Entity Types

```hcl
data "bcm_cmpart_entity_info" "all" {}

# Group entities by type
output "entities_by_type" {
  value = {
    for type in distinct([for e in data.bcm_cmpart_entity_info.all.entities : e.type]) :
    type => [for e in data.bcm_cmpart_entity_info.all.entities : e.name if e.type == type]
  }
}
```

### 4. Validate Entity Exists

```hcl
data "bcm_cmpart_entity_info" "check" {
  type         = "Network"
  name_pattern = "management-net"
}

# Use count to conditionally create resources
resource "bcm_cmdevice_category" "conditional" {
  count = length(data.bcm_cmpart_entity_info.check.entities) > 0 ? 1 : 0
  name  = "depends-on-network"
  # ...
}
```

## Filter Reference

### Type Filter

- **Behavior**: Case-sensitive exact match
- **Valid values**: Any BCM entity type (e.g., `SoftwareImage`, `Category`, `Network`)
- **When omitted**: Returns all entity types

### Name Pattern Filter

- **Behavior**: Case-insensitive glob pattern matching
- **Wildcards**:
  - `*` - matches any sequence of characters (including empty)
  - `?` - matches exactly one character
- **When omitted**: Returns all entity names

### Filter Logic

When both filters are specified, they are combined with AND logic:
- Entity must match type exactly AND name pattern

## Output Reference

### Attributes

| Attribute | Type | Description |
|-----------|------|-------------|
| `id` | string | Data source identifier |
| `entities` | list | List of matching entities |
| `entities[].name` | string | Entity name |
| `entities[].type` | string | Entity type |
| `entities[].uuid` | string | Entity UUID |

## Troubleshooting

### Empty Results

If the `entities` list is empty:
1. Verify the `type` value is spelled correctly (case-sensitive)
2. Check the `name_pattern` syntax (glob patterns, not regex)
3. Use the basic query without filters to see what entities exist

### Performance

The API returns all entities in a single call (500+ entities typical). For better performance:
1. Always use `type` filter when you know the entity type
2. Consider caching results in local values if used multiple times

## Related Resources

- `bcm_cmpart_softwareimages` - Detailed software image information
- `bcm_cmdevice_categories` - Category details and management
- `bcm_cmnet_networks` - Network configuration details
