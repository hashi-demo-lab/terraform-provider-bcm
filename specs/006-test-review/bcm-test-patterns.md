# BCM-Specific Test Patterns and Considerations

**Feature**: Terraform BCM Provider Test Infrastructure
**Spec**: `specs/006-test-review/spec.md`
**Date**: 2025-01-21
**Author**: Claude Code (TDD Implementation)

## Overview

This document captures BCM API-specific patterns, quirks, and considerations discovered during acceptance test implementation. These patterns are critical for writing reliable, maintainable tests for the Terraform BCM provider.

---

## BCM API Characteristics

### JSON-RPC Protocol

The BCM API uses a custom JSON-RPC protocol with specific calling conventions:

```go
// Standard BCM API request structure
{
  "service": "CMPart",           // Service name (CMPart, cmdevice, CMNet, etc.)
  "call": "getSoftwareImage",    // Method name
  "args": ["image-name"]         // Optional arguments for parameterized calls
}
```

**Key Insights**:
- ✅ **Args Support**: BCM API supports variadic arguments for efficient direct lookups
- ✅ **Cookie-Based Auth**: Uses `cm-login-token` cookie, managed automatically by http.Client
- ⚠️ **Case Sensitivity**: Service names and method names are case-sensitive
- ⚠️ **Array vs String**: Some methods expect arrays even for single values (e.g., `removeSoftwareImages([name], false)`)

---

## Field Name Mapping (snake_case ↔ camelCase)

**Critical Pattern**: Terraform schemas use `snake_case`, but BCM API uses `camelCase`. This mapping is essential for drift detection tests when modifying resources via the BCM API.

### bcm_cmpart_softwareimage Mappings

| Terraform Schema (snake_case) | BCM API Field (camelCase) | Example |
|-------------------------------|---------------------------|---------|
| `kernel_parameters` | `kernelParameters` | `"quiet splash"` |
| `enable_sol` | `enableSOL` | `true` (acronym uppercase) |
| `sol_speed` | `solSpeed` | `115200` |
| `sol_flow_control` | `solFlowControl` | `"hardware"` |
| `sol_port` | `solPort` | `1` |
| `kernel_output_console` | `kernelOutputConsole` | `"ttyS0"` |
| `kernel_version` | `kernelVersion` | `"5.15.0"` |
| `notes` | `notes` | No transformation |
| `path` | `path` | No transformation |
| `original_image` | `originalImage` | `"base-image"` |
| `software_image_proxy` | `softwareImageProxy` | Object |
| `modules` | `modules` | Array |

### bcm_cmdevice_category Mappings

| Terraform Schema (snake_case) | BCM API Field (camelCase) | Example |
|-------------------------------|---------------------------|---------|
| `kernel_parameters` | `kernelParameters` | `"quiet splash"` |
| `notes` | `notes` | No transformation |
| `install_boot_record` | `installBootRecord` | `true` |
| `allow_networking_restart` | `allowNetworkingRestart` | `false` |
| `management_network` | `managementNetwork` | UUID |
| `boot_loader` | `bootLoader` | `"grub2"` |
| `software_image_proxy` | `softwareImageProxy` | Object |
| `bmc_settings` | `bmcSettings` | Object |
| `force` | `force` | **Not persisted in BCM** |

**Transformation Rules**:
1. **snake_case → camelCase**: `kernel_parameters` → `kernelParameters`
2. **Acronyms Uppercase**: `enable_sol` → `enableSOL`, `sol_flow_control` → `solFlowControl`
3. **No Transformation**: `notes` → `notes`, `path` → `path`
4. **Non-Persisted Fields**: `force` flag is write-only, not stored in BCM

---

## BCM Entity Structure for Updates

When updating resources via BCM API (e.g., in drift detection tests), you must wrap the resource data in a BCM entity structure:

```go
// Drift test PreConfig pattern: Modify resource externally via BCM API
PreConfig: func() {
    client := createTestBCMClient(t)
    ctx := context.Background()

    // 1. Get resource UUID by name
    uuid := getResourceUUIDByName(t, "CMPart", "getSoftwareImage", imageName)

    // 2. Fetch full resource data
    body, _ := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName)
    var resourceData map[string]interface{}
    json.Unmarshal(body, &resourceData)

    // 3. Modify field externally (CRITICAL: Use camelCase field name!)
    resourceData["kernelParameters"] = "quiet splash nomodeset"

    // 4. Wrap in BCM entity structure
    entity := map[string]interface{}{
        "baseType":      "SoftwareImage",  // Resource type
        "childType":     "",                // Usually empty
        "modified":      true,              // Mark as modified
        "to_be_removed": false,             // Not deleting
        "revision":      "",                // Usually empty
        "uuid":          uuid,              // Resource UUID
    }

    // 5. Copy resource data into entity (excluding uuid)
    for k, v := range resourceData {
        if k != "uuid" {
            entity[k] = v
        }
    }

    // 6. Update via BCM API
    client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, false)

    // 7. CRITICAL: Wait for eventual consistency
    time.Sleep(2 * time.Second)

    // 8. Optional: Verify modification (debug logging)
    t.Logf("[DEBUG] Modified kernelParameters externally to: %v", entity["kernelParameters"])
}
```

**Required Entity Fields**:
- `baseType`: Resource type (e.g., `"SoftwareImage"`, `"Category"`)
- `childType`: Subtype (usually empty string `""`)
- `modified`: Boolean flag (set to `true` for updates)
- `to_be_removed`: Boolean flag (set to `false` unless deleting)
- `revision`: Version string (usually empty string `""`)
- `uuid`: Resource UUID (MUST be included in entity structure)

---

## Eventual Consistency Handling

BCM API exhibits eventual consistency - changes may not be immediately visible.

### Recommended Wait Times

| Operation | Wait Time | Reason |
|-----------|-----------|--------|
| After external modification (drift tests) | 2 seconds | Allow BCM to propagate changes |
| After deletion | Exponential backoff (1s, 2s, 4s, 8s) | Handle async cleanup |
| After creation | No wait needed | Synchronous operation |
| After bulk delete (PreCheck cleanup) | 2 seconds | Allow cleanup to complete |

### Exponential Backoff Pattern

```go
// verifyResourceDeleted helper uses exponential backoff
func verifyResourceDeleted(ctx context.Context, client *BCMClient, service, method, identifier string, maxRetries int) (bool, error) {
    waitTime := 1 * time.Second

    for retry := 0; retry < maxRetries; retry++ {
        time.Sleep(waitTime)

        // Attempt to read resource
        body, err := client.CallJSONRPC(ctx, service, method, identifier)

        // Error = resource deleted
        if err != nil {
            return true, nil
        }

        // Empty response = resource deleted
        if len(body) == 0 {
            return true, nil
        }

        // Empty JSON object = resource deleted
        var data map[string]interface{}
        if json.Unmarshal(body, &data) == nil && len(data) == 0 {
            return true, nil
        }

        // Resource still exists, wait longer
        waitTime *= 2 // Exponential backoff: 1s → 2s → 4s → 8s
    }

    return false, nil // Resource still exists after all retries
}
```

**Retry Schedule (maxRetries=4)**:
- Retry 0: Wait 1s, check (total: 1s)
- Retry 1: Wait 2s, check (total: 3s)
- Retry 2: Wait 4s, check (total: 7s)
- Retry 3: Wait 8s, check (total: 15s)
- **Total**: 15 seconds (within 30s FR-016 requirement)

---

## BCM API Method Patterns

### List vs Get Methods

BCM API provides both list and get methods with different purposes:

| Pattern | Method | Args | Use Case | Example |
|---------|--------|------|----------|---------|
| **List (Plural)** | `getSoftwareImages()` | None | Get all resources, client-side filter | Data sources |
| **Get (Singular)** | `getSoftwareImage(name)` | Name/ID | Direct lookup by identifier | Resources (Read) |
| **Add** | `addSoftwareImage(entity)` | Entity | Create new resource | Resources (Create) |
| **Update** | `updateSoftwareImage(entity, force)` | Entity + flags | Modify existing resource | Resources (Update) |
| **Remove (Plural)** | `removeSoftwareImages(names, force)` | Array + flags | Delete resources | Resources (Delete) |

**Critical Insights**:
- ⚠️ **Remove takes arrays**: `removeSoftwareImages([name], false)` NOT `removeSoftwareImages(name, false)`
- ✅ **Get takes single arg**: `getSoftwareImage(name)` for direct lookup
- ⚠️ **Update takes entity**: Full entity structure required (not just changed fields)

---

## Error Handling Patterns

### CheckDestroy Idempotency

CheckDestroy functions MUST be idempotent - they should gracefully handle resources that are already deleted.

```go
func testAccCheckCMPartSoftwareImageDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    ctx := context.Background()

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmpart_softwareimage" {
            continue
        }

        imageName := rs.Primary.Attributes["name"]

        // Exponential backoff verification
        deleted, err := verifyResourceDeleted(ctx, client, "CMPart", "getSoftwareImage", imageName, 4)

        // CRITICAL: API errors are NOT fatal - resource may already be deleted
        if err != nil {
            fmt.Printf("[DEBUG] CheckDestroy verification error for %s: %v\n", imageName, err)
        }

        // ONLY fail if resource definitively still exists
        if !deleted {
            return fmt.Errorf("Software image %s still exists after destroy", imageName)
        }
    }

    return nil
}
```

**Key Principles**:
1. **API errors are warnings, not failures**: Resource may already be deleted
2. **Only fail if resource exists**: Confirmed existence is the error condition
3. **Use exponential backoff**: Handle eventual consistency
4. **Log debug info**: Help troubleshoot intermittent issues

---

## PreCheck Cleanup Patterns

Enhanced PreCheck functions clean up leftover resources from failed test runs:

```go
func testAccPreCheckCMPartSoftwareImage(t *testing.T) {
    // Standard provider credential checks
    testAccPreCheck(t)

    // Cleanup leftover test resources
    cleanupLeftoverSoftwareImages(t)
}

func cleanupLeftoverSoftwareImages(t *testing.T) {
    client := createTestBCMClient(t)
    ctx := context.Background()

    // 1. Query all resources
    body, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImages")
    if err != nil {
        t.Logf("[WARN] Failed to query software images for cleanup: %v", err)
        return // Don't fail test, just log warning
    }

    var images []map[string]interface{}
    if err := json.Unmarshal(body, &images); err != nil {
        t.Logf("[WARN] Failed to parse software images: %v", err)
        return
    }

    // 2. Find test resources (prefix matching)
    var testImages []string
    for _, img := range images {
        if name, ok := img["name"].(string); ok {
            if strings.HasPrefix(name, "test-") {
                testImages = append(testImages, name)
            }
        }
    }

    // 3. Bulk delete if found
    if len(testImages) > 0 {
        t.Logf("[INFO] Cleaning up %d leftover test images", len(testImages))

        // CRITICAL: Use array, not individual strings
        _, err := client.CallJSONRPC(ctx, "CMPart", "removeSoftwareImages", testImages, false)
        if err != nil {
            t.Logf("[WARN] Failed to remove test images: %v", err)
        }

        // Wait for cleanup to complete
        time.Sleep(2 * time.Second)
    }
}
```

**Best Practices**:
1. **Prefix test resources**: Use `test-` prefix for easy identification
2. **Bulk delete**: More efficient than individual deletes
3. **Don't fail on cleanup errors**: Log warnings, don't break tests
4. **Wait after cleanup**: Allow BCM to process deletions

---

## Drift Detection Test Patterns

### Three-Step Test Structure

Drift detection tests follow a specific three-step pattern:

```go
Steps: []resource.TestStep{
    // Step 1: Create resource with initial value
    {
        Config: testAccResourceConfig(name, "initial-value"),
        Check: resource.ComposeAggregateTestCheckFunc(
            resource.TestCheckResourceAttr("bcm_resource.test", "attr", "initial-value"),
        ),
    },
    // Step 2: Modify externally via BCM API, verify drift detected
    {
        PreConfig: func() {
            // Modify resource via BCM API (see entity structure above)
            // ...
        },
        Config: testAccResourceConfig(name, "initial-value"), // Same config!
        ConfigPlanChecks: resource.ConfigPlanChecks{
            PreApply: []plancheck.PlanCheck{
                plancheck.ExpectNonEmptyPlan(), // CRITICAL: Verify drift detected
            },
        },
    },
    // Step 3: Terraform restores desired state
    {
        Config: testAccResourceConfig(name, "initial-value"),
        Check: resource.ComposeAggregateTestCheckFunc(
            resource.TestCheckResourceAttr("bcm_resource.test", "attr", "initial-value"),
        ),
    },
}
```

**Common Mistakes**:
- ❌ **Wrong Pattern**: Using `Check` with `TestCheckResourceAttr` in Step 2
  - This checks **state values**, not plan changes
  - Test will fail because state still shows old value before apply
- ✅ **Correct Pattern**: Using `ConfigPlanChecks` with `ExpectNonEmptyPlan()`
  - This verifies **plan is not empty** (drift detected)
  - Terraform detects difference between state and actual resource

**Required Imports**:
```go
import (
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
)
```

---

## Test Resource Naming

### Unique Name Generation

Always use unique names to avoid conflicts between parallel tests or leftover resources:

```go
// Generate timestamp-based unique name
imageName := generateUniqueTestName("test-image")
// Returns: "test-image-20250121-143052"

// Use in test config
Config: testAccCMPartSoftwareImageResourceConfig_DriftKernel(
    imageName,  // Unique name
    imagePath,
    "quiet splash",
),
```

**Benefits**:
1. **Parallel test safety**: No conflicts between concurrent tests
2. **Cleanup resilience**: Failed tests leave identifiable resources
3. **Debug visibility**: Timestamps show when resource was created

---

## BCM-Specific Debug Tips

### Verify External Modifications

When drift tests fail, verify the external modification actually happened:

```go
PreConfig: func() {
    // ... modify resource ...

    // CRITICAL: Add debug logging
    t.Logf("[DEBUG] Modified kernelParameters externally to: %v", entity["kernelParameters"])

    // Optional: Verify modification by re-querying
    verifyBody, _ := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName)
    var verifyData map[string]interface{}
    json.Unmarshal(verifyBody, &verifyData)
    t.Logf("[DEBUG] BCM API shows kernelParameters: %v", verifyData["kernelParameters"])
}
```

### Check API Response Format

BCM API responses can have different formats:

```go
// Success: Non-empty JSON object
{"uuid": "...", "name": "...", "kernelParameters": "..."}

// Not found (deleted): Error response OR empty array OR empty object
// Error: API returns error (caught by CallJSONRPC)
// Empty: []
// Empty object: {}
```

---

## Common Pitfalls and Solutions

| Pitfall | Symptom | Solution |
|---------|---------|----------|
| **Using state values in drift Step 2** | Test expects modified value but gets original | Use `ConfigPlanChecks` with `ExpectNonEmptyPlan()` |
| **Wrong field name in BCM API call** | Drift not detected | Use camelCase mappings (see table above) |
| **Missing entity fields** | API error during update | Include all required fields (baseType, uuid, etc.) |
| **Array vs string confusion** | HTTP 400 type error | Use `[]string{name}` for remove methods |
| **No wait after modification** | Drift not detected | Add `time.Sleep(2 * time.Second)` |
| **CheckDestroy fails on cleanup** | Tests fail even though resources deleted | Make CheckDestroy idempotent (see pattern above) |
| **Missing management_network** | Category creation fails | Look up via data source before creating |

---

## Test Environment Configuration

### Required Environment Variables

```bash
export TF_ACC=1                                    # Enable acceptance tests
export BCM_ENDPOINT="https://172.21.15.254:8081"  # BCM API endpoint
export BCM_USERNAME="root"                         # BCM username
export BCM_PASSWORD="Hashicorp123!"                # BCM password
```

### Test Execution Commands

```bash
# Run all acceptance tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Run drift detection tests only
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Drift"

# Run destroy edge case tests only
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Destroy"

# Run single resource tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMPartSoftwareImage"
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "CMDeviceCategory"

# Verbose debug output
TF_LOG=TRACE TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "Drift"
```

---

## References

- **BCM API Documentation**: `sampleRest/CMDevice_Complete_Documentation.md`
- **Test Helpers**: `internal/provider/test_helpers.go`
- **Drift Detection Pattern**: `CLAUDE.md` (Drift Detection Test Pattern section)
- **Enhanced Test Patterns**: `AGENTS.md` (Enhanced CheckDestroy and PreCheck Patterns section)
- **Coverage Matrix**: `specs/006-test-review/test-coverage.md`

---

**Last Updated**: 2025-01-21
**Status**: MVP Complete (Phases 1-4)
**Next Steps**: Use these patterns when adding new resources or expanding test coverage
