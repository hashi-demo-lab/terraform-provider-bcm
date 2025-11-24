# Data Model: Deletion Order Validation

**Feature**: Deletion Order Validation
**Date**: 2025-11-24
**Purpose**: Define data structures and dependency relationships for BCM resource deletion validation

## Resource Dependency Hierarchy

### Dependency Graph

```
┌─────────────────────┐
│  Software Images    │  Level 0 (no dependencies)
│  (Level 0)          │
└──────────┬──────────┘
           │ depends on
           ▼
┌─────────────────────┐
│   Categories        │  Level 1 (depend on Software Images)
│   (Level 1)         │
└──────────┬──────────┘
           │ depends on
           ▼
┌─────────────────────┐
│    Devices          │  Level 2 (depend on Categories)
│    (Level 2)        │
└─────────────────────┘

Independent Resources (Level 0):
┌─────────────────────┐
│     Networks        │  No dependencies
└─────────────────────┘

┌─────────────────────┐
│ Kubernetes Clusters │  No dependencies
└─────────────────────┘
```

### Dependency Relationships

| Resource | Depends On | Referenced By | Dependency Field |
|----------|------------|---------------|------------------|
| Software Image | None | Categories | `category.softwareimage` |
| Category | Software Image | Devices | `device.category` |
| Device | Category | None | N/A |
| Network | None | Categories (optional) | `category.managementNetwork` |
| Kubernetes Cluster | None | None | N/A |

## Deletion Order

### Safe Deletion Sequence

Resources must be deleted in **reverse dependency order** (highest level first):

```
1. Devices          (Level 2) - Depend on Categories
   ↓
2. Kubernetes Clusters (Level 0) - Independent
   ↓
3. Networks         (Level 0) - Independent
   ↓
4. Categories       (Level 1) - Depend on Software Images
   ↓
5. Software Images  (Level 0) - Lowest level, no dependencies
```

**Rationale**: Deleting a resource that is depended upon by other resources creates orphaned references in the BCM database. Deletion order ensures dependent resources are removed before their dependencies.

## Data Structures

### DependencyCheckResult

**Purpose**: Encapsulate results of dependency validation

**Structure**:
```go
type DependencyCheckResult struct {
    // HasDependencies indicates if dependent resources were found
    HasDependencies bool

    // DependentCount is the number of dependent resources
    DependentCount int

    // DependentType is the resource type (e.g., "devices", "categories")
    DependentType string

    // Identifiers contains resource identifiers (names or UUIDs)
    Identifiers []ResourceIdentifier

    // ErrorMessage is formatted user-facing error message
    ErrorMessage string
}
```

**Methods**:
```go
// HasDependencies returns true if dependent resources exist
func (r *DependencyCheckResult) HasDependencies() bool

// FormatErrorMessage generates user-facing error message with resolution options
func (r *DependencyCheckResult) FormatErrorMessage(resourceType, resourceName string) string

// TruncateIdentifiers limits identifier list for readability (max 10)
func (r *DependencyCheckResult) TruncateIdentifiers(max int) []ResourceIdentifier
```

### ResourceIdentifier

**Purpose**: Uniquely identify dependent resources in error messages

**Structure**:
```go
type ResourceIdentifier struct {
    // UUID is the BCM resource UUID
    UUID string

    // Name is the human-readable resource name (hostname for devices, name for others)
    Name string

    // Type is the resource type (for categorization)
    Type string
}
```

**Example**:
```go
ResourceIdentifier{
    UUID: "abc123-def456-789",
    Name: "node01",
    Type: "device",
}
```

### DeletionOrder

**Purpose**: Define deletion order constants

**Structure**:
```go
// DeletionOrder defines the safe deletion sequence for BCM resources
var DeletionOrder = []ResourceType{
    ResourceTypeDevice,          // Level 2 - delete first
    ResourceTypeKubernetesCluster, // Level 0 - independent
    ResourceTypeNetwork,          // Level 0 - independent
    ResourceTypeCategory,         // Level 1 - delete before images
    ResourceTypeSoftwareImage,    // Level 0 - delete last
}

// ResourceType represents BCM resource types
type ResourceType string

const (
    ResourceTypeDevice            ResourceType = "bcm_cmdevice_device"
    ResourceTypeCategory          ResourceType = "bcm_cmdevice_category"
    ResourceTypeSoftwareImage     ResourceType = "bcm_cmpart_softwareimage"
    ResourceTypeNetwork           ResourceType = "bcm_cmnet_network"
    ResourceTypeKubernetesCluster ResourceType = "bcm_cmkube_cluster"
)
```

## Dependency Check Queries

### Check Devices in Category

**Query**:
```go
// Get all devices
body, err := client.CallJSONRPC(ctx, "CMDevice", "getNodes")

// Parse response
var devices []map[string]interface{}
json.Unmarshal(body, &devices)

// Filter by category UUID
var dependentDevices []ResourceIdentifier
for _, device := range devices {
    categoryUUID, ok := device["category"].(string)
    if ok && categoryUUID == targetCategoryUUID {
        dependentDevices = append(dependentDevices, ResourceIdentifier{
            UUID: device["uuid"].(string),
            Name: device["hostname"].(string),
            Type: "device",
        })
    }
}
```

**Result**:
```go
&DependencyCheckResult{
    HasDependencies: len(dependentDevices) > 0,
    DependentCount: len(dependentDevices),
    DependentType: "devices",
    Identifiers: dependentDevices,
}
```

### Check Categories Using Software Image

**Query**:
```go
// Get all categories
body, err := client.CallJSONRPC(ctx, "CMDevice", "getCategories")

// Parse response
var categories []map[string]interface{}
json.Unmarshal(body, &categories)

// Filter by software image name
var dependentCategories []ResourceIdentifier
for _, category := range categories {
    imageName, ok := category["softwareimage"].(string)
    if ok && imageName == targetImageName {
        dependentCategories = append(dependentCategories, ResourceIdentifier{
            UUID: category["uuid"].(string),
            Name: category["name"].(string),
            Type: "category",
        })
    }
}
```

**Result**:
```go
&DependencyCheckResult{
    HasDependencies: len(dependentCategories) > 0,
    DependentCount: len(dependentCategories),
    DependentType: "categories",
    Identifiers: dependentCategories,
}
```

## Error Message Templates

### Dependency Violation Error

**Template**:
```
{ResourceType} In Use - Cannot Delete

{ResourceType} '{resourceName}' cannot be deleted because it has {count} {dependentType}(s) assigned.

Dependent {dependentType}:
  - {name1} (uuid: {uuid1})
  - {name2} (uuid: {uuid2})
  - {name3} (uuid: {uuid3})
  ... (showing first 10 of {total})

Resolution options:
  1. {specific action for resource type}
  2. Delete the dependent {dependentType} first
  3. Set 'force = true' to delete anyway (WARNING: will orphan references)
```

**Example (Category with Devices)**:
```
Category In Use - Cannot Delete

Category 'default' cannot be deleted because it has 3 device(s) assigned.

Dependent devices:
  - node01 (uuid: abc-123)
  - node02 (uuid: def-456)
  - node03 (uuid: ghi-789)

Resolution options:
  1. Reassign devices to another category before deleting
  2. Delete the dependent devices first
  3. Set 'force = true' to delete anyway (WARNING: will orphan device references)
```

**Example (Software Image with Categories)**:
```
Software Image In Use - Cannot Delete

Software Image 'Rocky-8.10' cannot be deleted because it has 2 categor(ies) assigned.

Dependent categories:
  - default (uuid: abc-123)
  - compute (uuid: def-456)

Resolution options:
  1. Update categories to use a different software image before deleting
  2. Delete the dependent categories first (after removing their devices)
  3. Set 'force = true' to delete anyway (WARNING: will orphan category references)
```

### Force Deletion Warning

**Template**:
```
Force Deletion Warning

{ResourceType} '{resourceName}' is being deleted with force=true. This may create orphaned references in dependent resources.

Potential Impact:
  - {specific impact for resource type}

This operation cannot be undone.
```

**Example (Category Force Delete)**:
```
Force Deletion Warning

Category 'default' is being deleted with force=true. This may create orphaned references in dependent resources.

Potential Impact:
  - Devices assigned to this category will have invalid category references
  - Device provisioning may fail until devices are reassigned to a valid category

This operation cannot be undone.
```

## Validation Rules

### Pre-Deletion Validation

**Rule**: Before deleting a resource, check for dependent resources (unless force=true)

**Logic**:
```go
func ValidateDelete(resourceType ResourceType, resourceID string, force bool) error {
    if force {
        // Log warning and proceed
        LogForceDeleteWarning(resourceType, resourceID)
        return nil
    }

    // Perform dependency check
    result := CheckDependencies(resourceType, resourceID)

    if result.HasDependencies {
        return fmt.Errorf(result.ErrorMessage)
    }

    return nil
}
```

### Ordered Deletion Validation

**Rule**: When deleting multiple resources, verify deletion order follows dependency graph

**Logic**:
```go
func ValidateDeletionOrder(resources map[ResourceType][]string) error {
    for i, resourceType := range DeletionOrder {
        // Verify resources of this type exist
        if _, exists := resources[resourceType]; !exists {
            continue
        }

        // Verify no resources of lower levels remain
        for j := i + 1; j < len(DeletionOrder); j++ {
            lowerType := DeletionOrder[j]
            if _, exists := resources[lowerType]; exists {
                return fmt.Errorf(
                    "Deletion order violation: attempting to delete %s before %s",
                    resourceType,
                    lowerType,
                )
            }
        }
    }

    return nil
}
```

## Testing Considerations

### Test Data Structure

**Test Resources**:
```go
type TestResources struct {
    // Created in order (bottom-up)
    SoftwareImages []string
    Categories     []string
    Devices        []string
    Networks       []string
    Clusters       []string
}

// CreateTestResources creates interdependent test resources
func CreateTestResources(t *testing.T) *TestResources

// DeleteTestResources deletes in correct order (top-down)
func DeleteTestResources(t *testing.T, resources *TestResources)
```

### Dependency Verification

**Test Helper**:
```go
// VerifyNoDependencies checks that dependent resources don't exist
func VerifyNoDependencies(
    t *testing.T,
    resourceType ResourceType,
    resourceID string,
) error
```

## Constraints

### BCM API Constraints

1. No native dependency check methods (must query and filter)
2. No referential integrity enforcement
3. Force parameter allows orphaned references
4. Eventual consistency requires retry logic

### Performance Constraints

1. Dependency checks must complete within 5 seconds
2. CheckDestroy must complete within 30 seconds total
3. Cleanup scripts should handle 1000+ resources efficiently

### Backward Compatibility Constraints

1. Force parameter must be optional (default=false)
2. Existing configurations must continue to work
3. No breaking changes to resource schemas

## Summary

This data model defines:

1. **Dependency Hierarchy**: 3-level hierarchy (Images → Categories → Devices)
2. **Deletion Order**: 5-step deletion sequence (Devices → Clusters → Networks → Categories → Images)
3. **Data Structures**: DependencyCheckResult, ResourceIdentifier, DeletionOrder
4. **Query Patterns**: Client-side filtering for dependency detection
5. **Error Templates**: Structured, actionable error messages
6. **Validation Rules**: Pre-deletion and ordered deletion validation logic

These structures enable safe, dependency-aware deletion of BCM resources across cleanup scripts, provider resources, and test infrastructure.
