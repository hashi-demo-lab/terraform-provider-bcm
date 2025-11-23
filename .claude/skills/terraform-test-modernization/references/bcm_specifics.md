# BCM-Specific Testing Patterns

BCM (Bright Cluster Manager) has specific patterns and quirks that affect test modernization.

## Field Name Mapping (Critical!)

**BCM API uses camelCase, Terraform uses snake_case.**

When modifying resources via BCM API in drift detection tests, you MUST map field names:

| Terraform (snake_case) | BCM API (camelCase) |
|------------------------|---------------------|
| `kernel_parameters` | `kernelParameters` |
| `enable_sol` | `enableSol` |
| `sol_speed` | `solSpeed` |
| `install_boot_record` | `installBootRecord` |
| `dhcp_enabled` | `dhcpEnabled` |
| `original_image` | `originalImage` |
| `master_nodes` | `masterNodes` |
| `worker_nodes` | `workerNodes` |

**Example**:
```go
// ❌ WRONG - using snake_case
resourceData["kernel_parameters"] = "modified"

// ✅ CORRECT - using camelCase
resourceData["kernelParameters"] = "modified"
```

## BCM API Entity Structure

When updating resources via BCM API, you must wrap data in the correct entity structure:

```go
entity := map[string]interface{}{
    "baseType":      "ResourceType",  // e.g., "SoftwareImage", "Category", "Device"
    "childType":     "",               // Usually empty
    "modified":      true,             // Mark as modified
    "to_be_removed": false,            // Don't remove
    "revision":      "",               // Empty for updates
    "uuid":          uuid,             // Resource UUID
}

// Copy resource data fields (excluding uuid)
for k, v := range resourceData {
    if k != "uuid" {
        entity[k] = v
    }
}

// Update via API
client.CallJSONRPC(ctx, "service", "updateMethod", entity, false)
```

## Test Helpers

BCM provider includes test helpers in `internal/provider/test_helpers.go`:

### createTestBCMClient(t)
Creates authenticated BCM client for tests.

```go
client := createTestBCMClient(t)
```

### getResourceUUIDByName(t, service, method, name)
Queries BCM API for resource UUID by name.

```go
uuid := getResourceUUIDByName(t, "cmpart", "getSoftwareImage", imageName)
```

**Service/Method Mapping**:
| Resource Type | Service | Get Method |
|---------------|---------|------------|
| Software Image | `cmpart` | `getSoftwareImage` |
| Partition | `cmpart` | `getPartition` |
| Category | `cmdevice` | `getCategory` |
| Device | `cmdevice` | `getDevice` |
| Network | `cmnet` | `getNetwork` |
| Cluster | `cmkube` | `getCluster` |

### verifyResourceDeleted(ctx, client, service, method, id, retries)
Verifies resource deletion with exponential backoff.

```go
deleted, err := verifyResourceDeleted(
    ctx,
    client,
    "cmpart",
    "getSoftwareImage",
    imageID,
    4, // retry count
)
```

### generateUniqueTestName(prefix)
Generates unique timestamp-based test names.

```go
imageName := generateUniqueTestName("test-image")
```

## Eventual Consistency

BCM API has eventual consistency. After external modifications, **always wait**:

```go
// Update via BCM API
client.CallJSONRPC(ctx, "service", "updateMethod", entity, false)

// Wait for consistency
time.Sleep(2 * time.Second)
```

**Common wait times**:
- Simple field updates: 2 seconds
- Image cloning: 5-10 seconds with polling
- Cluster operations: 10-30 seconds with polling

## Provider Configuration in Tests

All test configs must include provider configuration:

```go
func testAccResourceConfig(name string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_resource" "test" {
  name = %[4]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
    )
}
```

**Environment Variables**:
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

## CheckDestroy Pattern

Enhanced CheckDestroy with BCM-specific exponential backoff:

```go
func testAccCheckCMResourceDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})

    var errors []string
    resourceCount := 0

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmresource" {
            continue
        }

        resourceCount++
        id := rs.Primary.ID

        // Verify deletion with exponential backoff
        deleted, err := verifyResourceDeleted(
            context.Background(),
            client,
            "service",
            "getMethod",
            id,
            4, // retry count
        )

        if err != nil {
            errors = append(errors, fmt.Sprintf(
                "Resource type: %s, ID: %s, Error: %v",
                rs.Type,
                id,
                err,
            ))
        }

        if !deleted {
            errors = append(errors, fmt.Sprintf(
                "Resource still exists after destroy. Type: %s, ID: %s",
                rs.Type,
                id,
            ))
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failures:\n  - %s",
            strings.Join(errors, "\n  - "))
    }

    return nil
}
```

## Common BCM API Patterns

### Listing Resources (Data Sources)
```go
// Use plural method without args
body, err := client.CallJSONRPC(ctx, "cmpart", "getSoftwareImages")
```

### Getting Single Resource (Resources - Read)
```go
// Use singular method with name/ID as arg
body, err := client.CallJSONRPC(ctx, "cmpart", "getSoftwareImage", imageName)
```

### Creating Resource
```go
entity := map[string]interface{}{
    "baseType": "SoftwareImage",
    "name":     imageName,
    "path":     "/path/to/image",
    // ... other fields
}

body, err := client.CallJSONRPC(ctx, "cmpart", "addSoftwareImage", entity)
```

### Updating Resource
```go
// Fetch existing
body, _ := client.CallJSONRPC(ctx, "cmpart", "getSoftwareImage", uuid)
var data map[string]interface{}
json.Unmarshal(body, &data)

// Modify fields
data["kernelParameters"] = "new-value"

// Build entity with metadata
entity := map[string]interface{}{
    "baseType":      "SoftwareImage",
    "childType":     "",
    "modified":      true,
    "to_be_removed": false,
    "revision":      "",
    "uuid":          uuid,
}
for k, v := range data {
    if k != "uuid" {
        entity[k] = v
    }
}

// Update
client.CallJSONRPC(ctx, "cmpart", "updateSoftwareImage", entity, false)
```

### Deleting Resource
```go
// Most resources
client.CallJSONRPC(ctx, "cmpart", "removeSoftwareImage", uuid)

// Resources with "force remove" option
client.CallJSONRPC(ctx, "cmpart", "removeSoftwareImage", uuid, true)
```

## BCM-Specific Test Patterns

### Testing Image Cloning (Async Operation)
```go
// Create with original_image triggers cloning
{
    Config: testAccSoftwareImageConfigWithClone(imageName, originalImage),
    ConfigStateChecks: []statecheck.StateCheck{
        statecheck.ExpectKnownValue(
            "bcm_cmpart_softwareimage.test",
            tfjsonpath.New("name"),
            knownvalue.StringExact(imageName),
        ),
        // After cloning completes, original_image is cleared by BCM
        // We preserve it in state, but it's NOT in BCM API response
    },
},
```

**Note**: `original_image` field is cleared by BCM API after cloning completes. The provider must preserve it in state.

### Testing Network MTU
```go
// MTU is int64 in Terraform, but may be float64 in BCM API
statecheck.ExpectKnownValue(
    "bcm_cmnet_network.test",
    tfjsonpath.New("mtu"),
    knownvalue.Int64Exact(1500),
),
```

### Testing Boolean Flags
```go
// enable_sol, dhcp_enabled are bools
statecheck.ExpectKnownValue(
    "bcm_resource.test",
    tfjsonpath.New("enable_sol"),
    knownvalue.Bool(true),
),
```

## Known BCM API Quirks

### 1. Fields Cleared After Operations
Some fields are cleared by BCM after certain operations:
- `original_image` - Cleared after cloning completes
- Solution: Preserve in Terraform state even if not in API response

### 2. Field Type Conversions
BCM API may return different types than expected:
- Int64 values may come as float64
- Helper functions handle conversion: `getInt64Value()`

### 3. Null vs Empty String
BCM API may return empty strings where Terraform expects null:
- Use helpers: `getStringValue()` handles null-safe extraction

### 4. UUID vs ID
- BCM resources have `uuid` field (internal)
- Terraform uses `id` field (computed)
- Mapping: `id = uuid` in most cases

### 5. Computed Fields
Some fields are computed by BCM and should be `Computed: true`:
- `uuid` - Always computed
- `creation_time` - Computed
- `id` - Computed (same as uuid)

## Example Drift Detection Test (BCM-Specific)

```go
func TestAccCMPartSoftwareImage_DriftKernelParameters(t *testing.T) {
    imageName := generateUniqueTestName("test-drift")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
        Steps: []resource.TestStep{
            {
                Config: testAccCMPartSoftwareImageConfigWithKernel(imageName, "initial params"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_softwareimage.test",
                        tfjsonpath.New("kernel_parameters"),
                        knownvalue.StringExact("initial params"),
                    ),
                },
            },
            {
                PreConfig: func() {
                    client := createTestBCMClient(t)
                    ctx := context.Background()

                    // Get UUID
                    uuid := getResourceUUIDByName(t, "cmpart", "getSoftwareImage", imageName)

                    // Fetch resource
                    body, _ := client.CallJSONRPC(ctx, "cmpart", "getSoftwareImage", uuid)
                    var imageData map[string]interface{}
                    json.Unmarshal(body, &imageData)

                    // Modify field (camelCase!)
                    imageData["kernelParameters"] = "modified externally"

                    // Build entity
                    entity := map[string]interface{}{
                        "baseType":      "SoftwareImage",
                        "childType":     "",
                        "modified":      true,
                        "to_be_removed": false,
                        "revision":      "",
                        "uuid":          uuid,
                    }
                    for k, v := range imageData {
                        if k != "uuid" {
                            entity[k] = v
                        }
                    }

                    // Update
                    client.CallJSONRPC(ctx, "cmpart", "updateSoftwareImage", entity, false)
                    time.Sleep(2 * time.Second)

                    t.Logf("[DEBUG] Modified kernelParameters to: %v", entity["kernelParameters"])
                },
                Config: testAccCMPartSoftwareImageConfigWithKernel(imageName, "initial params"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectNonEmptyPlan(),
                    },
                },
            },
            {
                Config: testAccCMPartSoftwareImageConfigWithKernel(imageName, "initial params"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_softwareimage.test",
                        tfjsonpath.New("kernel_parameters"),
                        knownvalue.StringExact("initial params"),
                    ),
                },
            },
        },
    })
}
```

**Key BCM-specific elements**:
1. ✅ Use `createTestBCMClient(t)` helper
2. ✅ Use `getResourceUUIDByName()` helper
3. ✅ Map field names: `kernel_parameters` → `kernelParameters`
4. ✅ Use correct service/method: `cmpart.getSoftwareImage`
5. ✅ Include full BCM entity structure
6. ✅ Wait 2 seconds for consistency
7. ✅ Use debug logging for troubleshooting
