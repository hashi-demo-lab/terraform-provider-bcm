# Quickstart: Category Test Coverage Enhancement

**Feature Branch**: `001-category-test-coverage`
**Date**: 2025-11-25

## Quick Reference

### Run Acceptance Tests

```bash
# Set environment variables
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Run all category tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategory"

# Run specific test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run "TestAccCMDeviceCategoryResource_InstallationModes"
```

### File Locations

| File | Purpose |
|------|---------|
| `/workspace/internal/provider/resource_cmdevice_category.go` | Resource implementation |
| `/workspace/internal/provider/resource_cmdevice_category_test.go` | Acceptance tests |
| `/workspace/internal/provider/test_helpers.go` | Shared test helpers |

---

## Test Development Workflow

### 1. Create Test Function

```go
func TestAccCMDeviceCategoryResource_FEATURE(t *testing.T) {
    categoryName := generateUniqueTestName("tftest-feature")

    // Cleanup leftover test resources
    testAccCMDeviceCategoryPreCheck(t, categoryName)

    // ID consistency tracking
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create
            // Step 2: Idempotency after Create
            // Step 3: Update
            // Step 4: Idempotency after Update
            // Step 5: Import (optional)
        },
    })
}
```

### 2. Create Config Helper

```go
func testAccCMDeviceCategoryResourceConfig_FEATURE(name string, param1 string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmdevice_categories" "all" {}
data "bcm_cmpart_softwareimages" "all" {}

locals {
  management_network_uuid = length(data.bcm_cmdevice_categories.all.categories) > 0 ? data.bcm_cmdevice_categories.all.categories[0].management_network_id : "00000000-0000-0000-0000-000000000000"
  software_image_uuid = length(data.bcm_cmpart_softwareimages.all.images) > 0 ? data.bcm_cmpart_softwareimages.all.images[0].uuid : "00000000-0000-0000-0000-000000000000"
}

resource "bcm_cmdevice_category" "test" {
  name               = %[4]q
  management_network = local.management_network_uuid

  # Feature-specific field
  feature_field      = %[5]q

  software_image_proxy = {
    parent_software_image = local.software_image_uuid
  }
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        param1,
    )
}
```

### 3. Add State Checks

```go
ConfigStateChecks: []statecheck.StateCheck{
    // Verify name
    statecheck.ExpectKnownValue(
        "bcm_cmdevice_category.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact(categoryName),
    ),
    // Verify feature field
    statecheck.ExpectKnownValue(
        "bcm_cmdevice_category.test",
        tfjsonpath.New("feature_field"),
        knownvalue.StringExact("expected_value"),
    ),
    // Track ID consistency
    compareID.AddStateValue(
        "bcm_cmdevice_category.test",
        tfjsonpath.New("id"),
    ),
},
```

### 4. Add Idempotency Check

```go
{
    Config: testAccCMDeviceCategoryResourceConfig_FEATURE(categoryName, "value"),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectEmptyPlan(),
        },
    },
},
```

### 5. Add Import Step

```go
{
    ResourceName:      "bcm_cmdevice_category.test",
    ImportState:       true,
    ImportStateVerify: true,
    ImportStateVerifyIgnore: []string{"force"},
},
```

---

## Modern Testing Patterns

### Required Imports

```go
import (
    "context"
    "encoding/json"
    "fmt"
    "os"
    "testing"
    "time"

    "github.com/hashicorp/terraform-plugin-testing/compare"
    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/terraform"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)
```

### State Check Reference

| Type | Pattern |
|------|---------|
| String exact | `knownvalue.StringExact("value")` |
| String regex | `knownvalue.StringRegexp(regexp.MustCompile("pattern"))` |
| Bool | `knownvalue.Bool(true)` |
| Int64 | `knownvalue.Int64Exact(100)` |
| Not null | `knownvalue.NotNull()` |
| List size | `knownvalue.ListSizeExact(2)` |

### Path Navigation

| Path | Pattern |
|------|---------|
| Top-level | `tfjsonpath.New("field_name")` |
| Nested object | `tfjsonpath.New("parent").AtMapKey("child")` |
| List element | `tfjsonpath.New("list").AtSliceIndex(0)` |
| Combined | `tfjsonpath.New("list").AtSliceIndex(0).AtMapKey("field")` |

---

## Test Helpers Reference

### generateUniqueTestName

Creates timestamp-based unique names to avoid conflicts:

```go
categoryName := generateUniqueTestName("tftest-feature")
// Result: "tftest-feature-1732537200"
```

### testAccCMDeviceCategoryPreCheck

Cleans up leftover test resources:

```go
testAccCMDeviceCategoryPreCheck(t, categoryName)
```

### createTestBCMClient

Creates authenticated BCM client for drift tests:

```go
client := createTestBCMClient(t)
```

### getResourceUUIDByName

Gets UUID from BCM API by name:

```go
uuid := getResourceUUIDByName(t, "cmdevice", "getCategory", categoryName)
```

### verifyResourceDeleted

Verifies resource deletion with exponential backoff:

```go
deleted := verifyResourceDeleted(ctx, client, "cmdevice", "getCategory", name, 5)
```

---

## Common Patterns

### Drift Detection Test

```go
{
    PreConfig: func() {
        client := createTestBCMClient(t)
        ctx := context.Background()
        uuid := getResourceUUIDByName(t, "cmdevice", "getCategory", categoryName)

        // Fetch and modify
        body, _ := client.CallJSONRPC(ctx, "cmdevice", "getCategory", categoryName)
        var categoryData map[string]interface{}
        json.Unmarshal(body, &categoryData)
        categoryData["field"] = "modified_value"

        // Update via API
        entity := map[string]interface{}{
            "baseType": "Category", "childType": "", "modified": true,
            "to_be_removed": false, "revision": "", "uuid": uuid,
        }
        for k, v := range categoryData { entity[k] = v }
        client.CallJSONRPC(ctx, "cmdevice", "updateCategory", entity)
        time.Sleep(2 * time.Second)
    },
    Config: testAccConfig(categoryName, "original_value"),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectNonEmptyPlan(),
        },
    },
},
```

### Validation Error Test

```go
{
    Config:      testAccConfigInvalid(categoryName, "invalid_value"),
    ExpectError: regexp.MustCompile(`error message pattern`),
},
```

---

## Checklist for New Tests

- [ ] Test function name follows pattern: `TestAccCMDeviceCategoryResource_FEATURE`
- [ ] Unique name generated with `generateUniqueTestName()`
- [ ] Pre-check cleanup with `testAccCMDeviceCategoryPreCheck()`
- [ ] ID consistency tracking with `statecheck.CompareValue()`
- [ ] Create step with state checks
- [ ] Idempotency check after create (`plancheck.ExpectEmptyPlan()`)
- [ ] Update step with state checks
- [ ] Idempotency check after update
- [ ] Import step (if applicable)
- [ ] CheckDestroy function specified
- [ ] Config helper uses environment variables for credentials
- [ ] Config helper looks up management_network and software_image UUIDs
