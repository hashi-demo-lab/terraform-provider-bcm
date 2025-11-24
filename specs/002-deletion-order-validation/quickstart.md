# Quick Start: Deletion Order Validation

**Feature**: Deletion Order Validation
**Audience**: Developers adding deletion validation to BCM resources
**Date**: 2025-11-24

## Overview

This guide shows how to implement deletion order validation for BCM resources using the dependency checking infrastructure. You'll learn how to:

1. Add dependency checks to resource Delete methods
2. Use dependency helper functions
3. Format error messages for users
4. Fix cleanup scripts with correct deletion order
5. Write tests for dependency validation

## Prerequisites

- Familiarity with Terraform Plugin Framework
- Understanding of BCM JSON-RPC API
- Knowledge of Go programming

## Dependency Graph Reference

```
Deletion Order (Safe Sequence):
1. Devices          → Depend on Categories
2. Kubernetes Clusters → Independent
3. Networks         → Independent
4. Categories       → Depend on Software Images
5. Software Images  → No dependencies
```

**Rule**: Always delete dependent resources before their dependencies.

## Adding Dependency Checks to Resources

### Step 1: Add Force Parameter to Schema

```go
// In resource schema (e.g., resource_cmdevice_category.go)
"force": schema.BoolAttribute{
    MarkdownDescription: "Force deletion even if resources depend on this category. " +
        "WARNING: Force deletion may create orphaned references in the BCM database.",
    Optional: true,
    Computed: true,
    Default:  booldefault.StaticBool(false),
},
```

### Step 2: Update Resource Model

```go
// In resource model struct
type CMDeviceCategoryResourceModel struct {
    // ... existing fields ...
    Force types.Bool `tfsdk:"force"`
}
```

### Step 3: Add Dependency Check to Delete Method

```go
func (r *CMDeviceCategoryResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
    var state CMDeviceCategoryResourceModel

    // Read state
    resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Check dependencies (unless force=true)
    if !state.Force.ValueBool() {
        result, err := CheckDevicesInCategory(ctx, r.client, state.UUID.ValueString())
        if err != nil {
            // Dependency check failed - log warning but allow force deletion
            resp.Diagnostics.AddWarning(
                "Dependency Check Failed",
                fmt.Sprintf(
                    "Unable to verify dependencies for category '%s': %s\n\n"+
                    "You can proceed with deletion by setting 'force = true'.",
                    state.Name.ValueString(),
                    err.Error(),
                ),
            )
            return
        }

        if result.HasDependencies {
            // Dependencies exist - block deletion
            resp.Diagnostics.AddError(
                "Category In Use - Cannot Delete",
                BuildDependencyError(
                    "Category",
                    state.Name.ValueString(),
                    "device",
                    result.Identifiers,
                ),
            )
            return
        }
    } else {
        // Force deletion - log warning
        tflog.Warn(ctx, "Force deleting category with potential dependencies", map[string]interface{}{
            "category_uuid": state.UUID.ValueString(),
            "category_name": state.Name.ValueString(),
        })
    }

    // Proceed with deletion
    _, err := r.client.CallJSONRPC(
        ctx,
        "CMDevice",
        "removeCategories",
        []string{state.UUID.ValueString()},
        state.Force.ValueBool(),
    )

    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting Category",
            fmt.Sprintf("BCM API error: %s", err.Error()),
        )
        return
    }

    tflog.Info(ctx, "Category deleted successfully", map[string]interface{}{
        "category_uuid": state.UUID.ValueString(),
        "category_name": state.Name.ValueString(),
        "force": state.Force.ValueBool(),
    })
}
```

## Using Dependency Helper Functions

### CheckDevicesInCategory

**Purpose**: Check if any devices are assigned to a category

**Usage**:
```go
import "github.com/hashicorp/terraform-provider-bcm/internal/provider"

result, err := provider.CheckDevicesInCategory(ctx, client, categoryUUID)
if err != nil {
    // API error - handle gracefully
    return err
}

if result.HasDependencies {
    // Dependencies exist - show error message
    errorMsg := provider.BuildDependencyError(
        "Category",
        categoryName,
        "device",
        result.Identifiers,
    )
    return fmt.Errorf(errorMsg)
}
```

### CheckCategoriesUsingImage

**Purpose**: Check if any categories are using a software image

**Usage**:
```go
result, err := provider.CheckCategoriesUsingImage(ctx, client, imageName)
if err != nil {
    // API error - handle gracefully
    return err
}

if result.HasDependencies {
    // Dependencies exist - show error message
    errorMsg := provider.BuildDependencyError(
        "Software Image",
        imageName,
        "category",
        result.Identifiers,
    )
    return fmt.Errorf(errorMsg)
}
```

## Error Message Formatting

### BuildDependencyError

**Purpose**: Create formatted, actionable error message

**Usage**:
```go
errorMsg := provider.BuildDependencyError(
    resourceType,   // "Category" or "Software Image"
    resourceName,   // "default"
    dependentType,  // "device" or "category"
    identifiers,    // []ResourceIdentifier from CheckResult
)
// Returns formatted multi-line error message
```

**Output Example**:
```
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

## Fixing Cleanup Scripts

### Correct Deletion Order

```bash
#!/usr/bin/env bash
# Cleanup script with correct deletion order

set -euo pipefail

# Support dry-run mode
DRY_RUN=${DRY_RUN:-false}

# Login and setup cookie file
# ... (authentication code)

echo "========================================="
echo "DELETION ORDER (Dependency-Safe):"
echo "  1. Devices (highest level)"
echo "  2. Kubernetes Clusters (independent)"
echo "  3. Networks (independent)"
echo "  4. Categories (mid-level)"
echo "  5. Software Images (lowest level)"
echo "========================================="
echo ""

# Step 1: Delete Devices
echo "[1/5] Deleting Devices..."
delete_resources "devices" "CMDevice" "getNodes" "removeNodes" "hostname"
check_bcm_health
sleep 2

# Step 2: Delete Kubernetes Clusters
echo "[2/5] Deleting Kubernetes Clusters..."
delete_resources "clusters" "CMKube" "getClusters" "removeClusters" "name"
check_bcm_health
sleep 2

# Step 3: Delete Networks
echo "[3/5] Deleting Networks..."
delete_resources "networks" "CMNet" "getNetworks" "removeNetworks" "name"
check_bcm_health
sleep 2

# Step 4: Delete Categories
echo "[4/5] Deleting Categories..."
delete_resources "categories" "CMDevice" "getCategories" "removeCategories" "name"
check_bcm_health
sleep 2

# Step 5: Delete Software Images
echo "[5/5] Deleting Software Images..."
delete_resources "images" "CMPart" "getSoftwareImages" "removeSoftwareImages" "name"
check_bcm_health

echo ""
echo "Cleanup complete!"
```

### Helper Functions

```bash
check_bcm_health() {
    curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d '{"service":"cmgui","call":"getSystemStatus"}' > /dev/null

    if [ $? -ne 0 ]; then
        echo "ERROR: BCM health check failed"
        exit 1
    fi
}

delete_resources() {
    local resource_type=$1
    local service=$2
    local get_method=$3
    local remove_method=$4
    local name_field=$5

    # Query resources
    RESOURCES=$(curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "{\"service\":\"$service\",\"call\":\"$get_method\"}" | \
        jq -r "[.[] | select(.$name_field | startswith(\"citest-\") or startswith(\"tftest-\")) | .uuid] | @json")

    if [ "$RESOURCES" = "[]" ]; then
        echo "  No $resource_type to delete"
        return
    fi

    # Dry-run mode
    if [ "$DRY_RUN" = "true" ]; then
        echo "  [DRY-RUN] Would delete $(echo $RESOURCES | jq -r 'length') $resource_type"
        return
    fi

    # Delete resources
    curl -k -s -b "$COOKIE_FILE" -X POST "${BCM_ENDPOINT}/json" \
        -H "Content-Type: application/json" \
        -d "{\"service\":\"$service\",\"call\":\"$remove_method\",\"args\":[$RESOURCES,false]}"

    echo "  Deleted $(echo $RESOURCES | jq -r 'length') $resource_type"
}
```

## Writing Tests for Dependency Validation

### Test Structure

```go
func TestAccCMDeviceCategory_DeleteWithDependencies(t *testing.T) {
    categoryName := generateUniqueTestName("test-category")
    deviceName := generateUniqueTestName("test-device")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create category and device
            {
                Config: testAccCategoryWithDevice(categoryName, deviceName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
                    resource.TestCheckResourceAttr("bcm_cmdevice_device.test", "hostname", deviceName),
                ),
            },
            // Step 2: Attempt to delete category with device assigned (should fail)
            {
                Config: testAccCategoryOnly(categoryName),  // Remove category from config
                ExpectError: regexp.MustCompile(
                    "Category In Use.*cannot be deleted.*has.*device.*assigned",
                ),
            },
        },
    })
}
```

### Test Force Deletion

```go
func TestAccCMDeviceCategory_DeleteWithForce(t *testing.T) {
    categoryName := generateUniqueTestName("test-category")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Step 1: Create category with device
            {
                Config: testAccCategoryWithDevice(categoryName, "test-device"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
                ),
            },
            // Step 2: Delete with force=true (should succeed)
            {
                Config: testAccCategoryWithForce(categoryName),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "force", "true"),
                ),
            },
        },
    })
}
```

### Test Config Helpers

```go
func testAccCategoryWithDevice(categoryName, deviceName string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name = "%[4]s-network"
}

resource "bcm_cmdevice_category" "test" {
  name                = %[4]q
  management_network  = bcm_cmnet_network.test.id
}

resource "bcm_cmdevice_device" "test" {
  hostname  = %[5]q
  category  = bcm_cmdevice_category.test.id
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        categoryName,
        deviceName,
    )
}
```

## Common Patterns

### Pattern 1: Dependency Check with Graceful Degradation

```go
// If dependency check fails, warn but allow force deletion
result, err := CheckDependencies(ctx, client, resourceID)
if err != nil {
    resp.Diagnostics.AddWarning(
        "Dependency Check Failed",
        fmt.Sprintf("API error: %s. Use 'force = true' to proceed.", err.Error()),
    )
    return
}
```

### Pattern 2: Force Deletion Logging

```go
if force {
    tflog.Warn(ctx, "Force deleting resource with potential dependencies", map[string]interface{}{
        "resource_type": "Category",
        "resource_uuid": uuid,
        "resource_name": name,
    })
}
```

### Pattern 3: Truncating Dependent Lists

```go
// Show max 10 dependents, truncate if more
identifiers := result.Identifiers
if len(identifiers) > 10 {
    identifiers = identifiers[:10]
    // Add truncation message to error
}
```

## Force Parameter Best Practices

### When to Use Force Deletion

✅ **Appropriate Use Cases**:
- Resource is corrupted and cannot be deleted normally
- Dependencies have been manually removed outside Terraform
- Emergency cleanup during incident response
- Testing and development (non-production)

❌ **Avoid Force Deletion**:
- Normal day-to-day operations
- Production environments
- Automated workflows (CI/CD)
- When dependency resolution is possible

### Documentation Example

```hcl
resource "bcm_cmdevice_category" "example" {
  name                = "test-category"
  management_network  = bcm_cmnet_network.mgmt.id

  # Force deletion even if devices are assigned
  # WARNING: This may create orphaned device references
  force = true
}
```

## Troubleshooting

### Problem: Dependency Check Returns False Positives

**Symptoms**: Dependency check reports dependencies that don't exist

**Solution**: BCM eventual consistency - add retry logic with exponential backoff

```go
result, err := CheckDependenciesWithRetry(ctx, client, resourceID, 3)
```

### Problem: Force Deletion Still Fails

**Symptoms**: Setting `force = true` doesn't bypass BCM error

**Solution**: BCM API may have other constraints. Check BCM API response for specific error:

```go
tflog.Debug(ctx, "BCM API response", map[string]interface{}{
    "response_body": string(body),
})
```

### Problem: Cleanup Script Hangs

**Symptoms**: Script doesn't complete, no error messages

**Solution**: Add timeout to curl commands and check BCM health:

```bash
curl --max-time 10 -k -s -b "$COOKIE_FILE" ...
check_bcm_health || exit 1
```

## Summary

**Key Takeaways**:

1. Always check dependencies before deletion (unless force=true)
2. Use helper functions for consistent dependency checking
3. Provide clear, actionable error messages
4. Log warnings for force deletions
5. Delete in correct order: Devices → Clusters → Networks → Categories → Images
6. Handle API errors gracefully (warn, don't block)
7. Write comprehensive tests for dependency validation
8. Document force parameter implications clearly

**Next Steps**:

1. Review existing resource Delete methods
2. Add dependency checks where appropriate
3. Update cleanup scripts with correct deletion order
4. Write acceptance tests for dependency validation
5. Update resource documentation with force parameter examples
