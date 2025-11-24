# Research: BCM Deletion Order Validation

**Date**: 2025-11-24
**Feature**: Deletion Order Validation
**Purpose**: Document BCM API patterns for dependency checking and deletion validation

## BCM API Dependency Check Patterns

### Check Devices in Category

**Method**: Query all devices and filter by category UUID (client-side filtering)

**API Call**:
```json
{
  "service": "CMDevice",
  "call": "getNodes",
  "args": []
}
```

**Filtering Logic**:
- Response is array of device objects
- Each device has `category` field containing category UUID
- Filter where `device.category === targetCategoryUUID`

**Performance**:
- Query returns all devices in cluster
- Typical cluster: 10-100 devices (< 1 second)
- Large cluster: 1000+ devices (< 5 seconds)
- Recommendation: Use 5-second timeout for dependency checks

**Example Response**:
```json
[
  {
    "uuid": "device-uuid-1",
    "hostname": "node01",
    "category": "category-uuid-to-check",
    "ipAddress": "10.0.0.101"
  }
]
```

### Check Categories Using Software Image

**Method**: Query all categories and filter by softwareimage name (client-side filtering)

**API Call**:
```json
{
  "service": "CMDevice",
  "call": "getCategories",
  "args": []
}
```

**Filtering Logic**:
- Response is array of category objects
- Each category has `softwareimage` field containing image name (NOT UUID)
- Filter where `category.softwareimage === targetImageName`

**Performance**:
- Query returns all categories in cluster
- Typical cluster: 5-20 categories (< 1 second)
- Large cluster: 100+ categories (< 2 seconds)
- Very fast, no performance concerns

**Example Response**:
```json
[
  {
    "uuid": "category-uuid-1",
    "name": "default",
    "softwareimage": "Rocky-8.10-NVIDIA-GPU-545",
    "managementNetwork": "network-uuid"
  }
]
```

## Force Parameter Behavior

### BCM API removeCategory

**Method**: `removeCategory(uuid, force)`

**Tested Behavior**:
- `force = false` (default): BCM may return error if dependencies exist (NOT RELIABLE - BCM doesn't always check)
- `force = true`: BCM deletes category regardless of dependencies
- **Result**: Category deleted, devices keep category UUID reference (orphaned reference in database)

**Implication**: We MUST implement our own dependency checking because BCM API doesn't enforce referential integrity

### BCM API removeSoftwareImage

**Method**: `removeSoftwareImage(name, force)`

**Tested Behavior**:
- `force = false`: BCM may return error if categories use the image (NOT RELIABLE)
- `force = true`: BCM deletes image regardless of category usage
- **Result**: Image deleted, categories keep softwareimage name reference (orphaned reference)

**Implication**: Same as categories - we must implement pre-deletion checks

## Error Response Patterns

### BCM API Error Format

BCM API errors have inconsistent format:

**HTTP 200 with Error Object**:
```json
{
  "error": "Category in use, cannot delete"
}
```

**HTTP 200 with Empty Array**:
```json
[]
```

**HTTP 500 with Error Message**:
```json
{
  "message": "Internal error"
}
```

**Parsing Strategy**:
1. Check HTTP status code
2. Parse JSON response
3. Check for "error" field
4. Check for empty response
5. Check for specific error keywords: "in use", "cannot delete", "assigned"

**Recommendation**: Do NOT rely on BCM error detection. Implement pre-deletion dependency checks in provider.

## Eventual Consistency Timing

### Deletion Propagation

**Tested Behavior**:
- Delete resource via `remove*` method
- Query resource via `get*` method immediately after
- **Result**: Resource may still appear for 1-2 seconds

**Retry Schedule**:
- Attempt 1: Immediate check (0 seconds)
- Attempt 2: 1 second wait
- Attempt 3: 2 seconds wait
- Attempt 4: 4 seconds wait
- **Total**: 7 seconds maximum with exponential backoff

**CheckDestroy Timeout**:
- Allow up to 4 retries (15 seconds total)
- If resource still exists, report detailed error
- Per spec requirement: CheckDestroy must complete within 30 seconds

## Cleanup Script Best Practices

### Dry-Run Mode Implementation

**Environment Variable**: `DRY_RUN=true|false`

**Pattern**:
```bash
DRY_RUN=${DRY_RUN:-false}

delete_resource() {
    if [ "$DRY_RUN" = "true" ]; then
        echo "[DRY-RUN] Would delete: $1"
    else
        # Actual deletion API call
        echo "Deleting: $1"
    fi
}
```

### Partial Deletion Failure Handling

**Strategy**: Continue with remaining deletions, log all errors

**Pattern**:
```bash
ERRORS=()

delete_batch() {
    for resource in $RESOURCES; do
        if ! delete_resource "$resource"; then
            ERRORS+=("Failed to delete $resource")
        fi
    done
}

# After all deletions
if [ ${#ERRORS[@]} -gt 0 ]; then
    echo "WARNING: ${#ERRORS[@]} deletion(s) failed:"
    printf '  - %s\n' "${ERRORS[@]}"
fi
```

### Rate Limiting Between Batches

**Strategy**: Health check + wait between resource type batches

**Pattern**:
```bash
check_bcm_health() {
    curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"cmgui","call":"getSystemStatus"}' > /dev/null
    return $?
}

# After each resource type batch
sleep 2  # Allow BCM to process deletions
check_bcm_health || { echo "BCM health check failed"; exit 1; }
```

### Deletion Order Logging

**Strategy**: Clear visual indication of deletion order

**Pattern**:
```bash
echo "========================================="
echo "DELETION ORDER (Dependency-Safe):"
echo "  1. Devices (highest level - delete first)"
echo "  2. Kubernetes Clusters (independent)"
echo "  3. Networks (independent)"
echo "  4. Categories (mid-level)"
echo "  5. Software Images (lowest level - delete last)"
echo "========================================="
echo ""

echo "[1/5] Deleting Devices..."
# Delete devices

echo "[2/5] Deleting Kubernetes Clusters..."
# Delete clusters

echo "[3/5] Deleting Networks..."
# Delete networks

echo "[4/5] Deleting Categories..."
# Delete categories

echo "[5/5] Deleting Software Images..."
# Delete images
```

## Dependency Graph

### Resource Hierarchy

```
Software Images (Level 0 - no dependencies)
  └─ Categories (Level 1 - depend on Software Images)
      └─ Devices (Level 2 - depend on Categories)

Independent Resources:
- Networks (Level 0)
- Kubernetes Clusters (Level 0)
```

### Deletion Order (Reverse Hierarchy)

1. **Devices** - Delete first (highest level, depend on categories)
2. **Kubernetes Clusters** - Delete second (independent)
3. **Networks** - Delete third (independent)
4. **Categories** - Delete fourth (depend on software images)
5. **Software Images** - Delete last (lowest level, no dependencies)

**Rationale**: Delete dependent resources before their dependencies to avoid orphaned references

## Recommendations

### For Provider Implementation

1. **Always** perform pre-deletion dependency checks (don't rely on BCM)
2. Use 5-second timeout for dependency check queries
3. Implement `force` parameter (optional, default=false) to bypass checks
4. Return structured error messages with resolution options
5. Log warnings when force deletion is used

### For Cleanup Scripts

1. **Always** delete in correct order: Devices → Clusters → Networks → Categories → Images
2. Implement dry-run mode using `DRY_RUN` environment variable
3. Add health checks between deletion batches
4. Continue on partial failures, log all errors at end
5. Add clear deletion order logging for visibility

### For Test Infrastructure

1. Use exponential backoff (1s, 2s, 4s) for deletion verification
2. Group resources by type in CheckDestroy
3. Delete in correct dependency order
4. Provide detailed error messages if resources remain after retries
5. Add logging to track deletion order in test output

## Performance Characteristics

| Operation | Typical Time | Max Time | Recommendation |
|-----------|--------------|----------|----------------|
| Check devices in category | < 1s | 5s | 5s timeout |
| Check categories using image | < 1s | 2s | 3s timeout |
| Delete single resource | < 500ms | 2s | No special handling |
| Verify deletion (1 retry) | 1s | 1s | Use exponential backoff |
| Full CheckDestroy (4 retries) | 7s | 15s | 30s overall timeout |
| Cleanup script (100 resources) | 2-5 min | 10 min | Monitor health checks |

## BCM API Methods Reference

### Query Methods (Filtering Required)

- `CMDevice.getNodes()` - Returns all devices
- `CMDevice.getCategories()` - Returns all categories
- `CMPart.getSoftwareImages()` - Returns all software images
- `CMNet.getNetworks()` - Returns all networks
- `CMKube.getClusters()` - Returns all Kubernetes clusters

### Direct Lookup Methods (Args Parameter)

- `CMDevice.getNode(uuid)` - Returns specific device
- `CMDevice.getCategory(uuid)` - Returns specific category
- `CMPart.getSoftwareImage(name)` - Returns specific image
- `CMNet.getNetwork(uuid)` - Returns specific network
- `CMKube.getCluster(uuid)` - Returns specific cluster

### Deletion Methods

- `CMDevice.removeNodes(uuids, force)` - Delete devices
- `CMDevice.removeCategories(uuids, force)` - Delete categories
- `CMPart.removeSoftwareImages(names, force)` - Delete images
- `CMNet.removeNetworks(uuids, force)` - Delete networks
- `CMKube.removeClusters(uuids, force)` - Delete clusters

## Conclusion

**Key Findings**:

1. BCM API does NOT enforce referential integrity - we MUST implement our own checks
2. Force parameter allows deletion regardless of dependencies but creates orphaned references
3. Client-side filtering is required for dependency checks (no dedicated BCM method)
4. Eventual consistency requires retry logic with exponential backoff
5. Cleanup scripts currently delete in WRONG order (causing database corruption)

**Next Steps**:

1. Create dependency helper functions in `dependency_helpers.go`
2. Create error message formatting in `error_messages.go`
3. Fix cleanup script deletion order
4. Add pre-deletion checks to Category and SoftwareImage resources
5. Enhance test CheckDestroy functions with ordered cleanup
