# Drift Detection Quickstart Guide

**Feature**: Adding Drift Detection Tests to Terraform BCM Provider
**For**: Developers adding new resources or expanding test coverage
**Time**: ~15-30 minutes per test

## What is Drift Detection?

Drift detection tests verify that the provider's `Read` operation correctly detects when a resource has been modified externally (outside of Terraform). These tests ensure Terraform can:

1. **Detect Changes**: Identify when resource state differs from Terraform configuration
2. **Plan Correctly**: Show accurate diff in `terraform plan` output
3. **Restore State**: Apply changes to bring resource back to desired state

---

## Prerequisites

Before writing drift detection tests, ensure you have:

- ✅ Working basic CRUD test for the resource
- ✅ BCM test environment access (endpoint, credentials)
- ✅ Shared test helpers available (`test_helpers.go`)
- ✅ Knowledge of BCM API field name mappings (snake_case ↔ camelCase)

---

## Step 1: Choose an Attribute to Test

Select a **mutable** attribute that:
- Can be modified via BCM API
- Is persisted in BCM (not write-only like `force`)
- Represents real drift scenarios

**Good Choices**:
- `kernel_parameters` - Common configuration change
- `notes` - Metadata updates
- `boot_loader` - System configuration changes

**Bad Choices**:
- `uuid` - Immutable identifier
- `force` - Write-only flag, not persisted
- Computed-only fields - Not user-configurable

---

## Step 2: Find BCM API Field Name

**CRITICAL**: Terraform uses `snake_case`, BCM API uses `camelCase`!

Use the field mapping tables:
- See `/workspace/CLAUDE.md` (Drift Detection Test Pattern section)
- See `/workspace/specs/006-test-review/bcm-test-patterns.md` (Field Name Mapping section)
- See `/workspace/internal/provider/test_helpers.go:14-51` (BCM API Field Name Mappings comments)

**Example Mappings**:
- `kernel_parameters` → `kernelParameters`
- `enable_sol` → `enableSOL` (acronym uppercase!)
- `notes` → `notes` (no transformation)

---

## Step 3: Write the Three-Step Test

### Required Imports

```go
import (
    "context"
    "encoding/json"
    "testing"
    "time"

    "github.com/hashicorp/terraform-plugin-testing/helper/resource"
    "github.com/hashicorp/terraform-plugin-testing/plancheck"  // CRITICAL!
)
```

### Test Structure Template

```go
func TestAccCMPartSoftwareImage_DriftKernelParameters(t *testing.T) {
    imageName := generateUniqueTestName("test-image-drift")
    imagePath := "/cm/images/minimal.qcow2"

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMPartSoftwareImage(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMPartSoftwareImageDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create resource with initial value
            {
                Config: testAccCMPartSoftwareImageResourceConfig_DriftKernel(imageName, imagePath, "quiet splash"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "name", imageName),
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash"),
                ),
            },
            // Step 2: Modify resource externally via BCM API
            {
                PreConfig: func() {
                    client := createTestBCMClient(t)
                    ctx := context.Background()

                    // Get resource UUID by name
                    uuid := getResourceUUIDByName(t, "CMPart", "getSoftwareImage", imageName)

                    // Fetch full resource data
                    body, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName)
                    if err != nil {
                        t.Fatalf("Failed to get software image: %v", err)
                    }

                    var imageData map[string]interface{}
                    if err := json.Unmarshal(body, &imageData); err != nil {
                        t.Fatalf("Failed to parse image data: %v", err)
                    }

                    // CRITICAL: Modify field using camelCase name!
                    imageData["kernelParameters"] = "quiet splash nomodeset"

                    // Wrap in BCM entity structure
                    entity := map[string]interface{}{
                        "baseType":      "SoftwareImage",
                        "childType":     "",
                        "modified":      true,
                        "to_be_removed": false,
                        "revision":      "",
                        "uuid":          uuid,
                    }

                    // Copy all fields except uuid
                    for k, v := range imageData {
                        if k != "uuid" {
                            entity[k] = v
                        }
                    }

                    // Update via BCM API
                    _, err = client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, false)
                    if err != nil {
                        t.Fatalf("Failed to update software image: %v", err)
                    }

                    // CRITICAL: Wait for eventual consistency
                    time.Sleep(2 * time.Second)

                    // Debug logging (optional but helpful)
                    t.Logf("[DEBUG] Modified kernelParameters externally to: %v", entity["kernelParameters"])
                },
                Config: testAccCMPartSoftwareImageResourceConfig_DriftKernel(imageName, imagePath, "quiet splash"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectNonEmptyPlan(), // CRITICAL: Verify drift detected!
                    },
                },
            },
            // Step 3: Terraform restores desired state
            {
                Config: testAccCMPartSoftwareImageResourceConfig_DriftKernel(imageName, imagePath, "quiet splash"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "kernel_parameters", "quiet splash"),
                ),
            },
        },
    })
}

// Config helper function
func testAccCMPartSoftwareImageResourceConfig_DriftKernel(name, path, kernelParams string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name               = %[4]q
  path               = %[5]q
  kernel_parameters  = %[6]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        path,
        kernelParams,
    )
}
```

---

## Step 4: Common Mistakes and Fixes

### ❌ Mistake 1: Using Check Instead of ConfigPlanChecks

```go
// WRONG: This checks state values, not plan changes
{
    PreConfig: func() { /* modify externally */ },
    Config: testAccResourceConfig(name, "initial-value"),
    Check: resource.ComposeAggregateTestCheckFunc(
        resource.TestCheckResourceAttr("bcm_resource.test", "attr", "modified-value"), // WRONG!
    ),
}
```

```go
// CORRECT: This verifies plan is not empty (drift detected)
{
    PreConfig: func() { /* modify externally */ },
    Config: testAccResourceConfig(name, "initial-value"),
    ConfigPlanChecks: resource.ConfigPlanChecks{
        PreApply: []plancheck.PlanCheck{
            plancheck.ExpectNonEmptyPlan(), // CORRECT!
        },
    },
}
```

### ❌ Mistake 2: Wrong Field Name in BCM API

```go
// WRONG: Using Terraform snake_case name
imageData["kernel_parameters"] = "modified"  // BCM API won't recognize this!
```

```go
// CORRECT: Using BCM camelCase name
imageData["kernelParameters"] = "modified"  // BCM API recognizes this
```

### ❌ Mistake 3: Missing Entity Fields

```go
// WRONG: Incomplete entity structure
entity := map[string]interface{}{
    "uuid": uuid,
    "kernelParameters": "modified",
}
```

```go
// CORRECT: Complete entity structure
entity := map[string]interface{}{
    "baseType":      "SoftwareImage",  // Required!
    "childType":     "",                 // Required!
    "modified":      true,               // Required!
    "to_be_removed": false,              // Required!
    "revision":      "",                 // Required!
    "uuid":          uuid,               // Required!
}
// Then copy resource fields...
```

### ❌ Mistake 4: No Wait After Modification

```go
// WRONG: No wait for eventual consistency
client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, false)
// Test might run before BCM propagates changes!
```

```go
// CORRECT: Wait for BCM to propagate changes
client.CallJSONRPC(ctx, "CMPart", "updateSoftwareImage", entity, false)
time.Sleep(2 * time.Second)  // Allow BCM to process
```

---

## Step 5: Run the Test

### Set Environment Variables

```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

### Run Your New Test

```bash
# Run just your new drift test
go test -v -timeout 120m ./internal/provider/ -run "TestAccCMPartSoftwareImage_DriftKernelParameters"

# Expected output:
# === RUN   TestAccCMPartSoftwareImage_DriftKernelParameters
# [DEBUG] Modified kernelParameters externally to: quiet splash nomodeset
# --- PASS: TestAccCMPartSoftwareImage_DriftKernelParameters (44.26s)
# PASS
```

### Debug Failed Tests

If the test fails, check:

1. **Did external modification work?**
   - Look for `[DEBUG]` log line showing modified value
   - Add more logging to verify BCM API call succeeded

2. **Is field name correct?**
   - Check camelCase mapping table
   - Verify BCM API field name in `sampleRest/` examples

3. **Did we wait long enough?**
   - Try increasing sleep time to 5 seconds
   - Check if BCM requires longer propagation

4. **Is ConfigPlanChecks being used?**
   - Must use `plancheck.ExpectNonEmptyPlan()`
   - NOT `TestCheckResourceAttr` in Step 2

---

## Step 6: Test Multiple Attributes (Optional)

To test multiple attributes, create separate test functions:

```go
func TestAccCMPartSoftwareImage_DriftKernelParameters(t *testing.T) { /* ... */ }
func TestAccCMPartSoftwareImage_DriftNotes(t *testing.T) { /* ... */ }
func TestAccCMPartSoftwareImage_DriftEnableSOL(t *testing.T) { /* ... */ }
```

**OR** create a single parameterized test (advanced):

```go
func TestAccCMPartSoftwareImage_Drift(t *testing.T) {
    testCases := []struct {
        name          string
        attribute     string
        bcmField      string
        initialValue  interface{}
        modifiedValue interface{}
    }{
        {"KernelParameters", "kernel_parameters", "kernelParameters", "quiet", "quiet splash"},
        {"Notes", "notes", "notes", "Original note", "Modified note"},
        {"EnableSOL", "enable_sol", "enableSOL", true, false},
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            // Run drift test with tc.bcmField and tc.modifiedValue
        })
    }
}
```

---

## Quick Reference Checklist

Before submitting your drift detection test:

- [ ] Test name follows pattern: `TestAcc<Resource>_Drift<Attribute>`
- [ ] Three-step structure: Create → Modify External → Restore
- [ ] Step 2 uses `ConfigPlanChecks` with `ExpectNonEmptyPlan()`
- [ ] BCM API field name is camelCase (check mapping table!)
- [ ] Entity structure includes all required fields (baseType, uuid, etc.)
- [ ] 2-second sleep after external modification
- [ ] Debug logging added for troubleshooting
- [ ] Test passes when run individually
- [ ] Test passes when run with full suite

---

## Example: Complete Drift Test (Category Notes)

```go
func TestAccCMDeviceCategory_DriftNotes(t *testing.T) {
    categoryName := generateUniqueTestName("test-category-drift")

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheckCMDeviceCategory(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy:             testAccCheckCMDeviceCategoryDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create with initial notes
            {
                Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Original notes"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "name", categoryName),
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Original notes"),
                ),
            },
            // Step 2: Modify notes externally via BCM API
            {
                PreConfig: func() {
                    client := createTestBCMClient(t)
                    ctx := context.Background()

                    uuid := getResourceUUIDByName(t, "cmdevice", "getCategory", categoryName)

                    body, _ := client.CallJSONRPC(ctx, "cmdevice", "getCategory", categoryName)
                    var categoryData map[string]interface{}
                    json.Unmarshal(body, &categoryData)

                    // CRITICAL: notes field has same name in both APIs!
                    categoryData["notes"] = "Modified via BCM API"

                    entity := map[string]interface{}{
                        "baseType":      "Category",
                        "childType":     "",
                        "modified":      true,
                        "to_be_removed": false,
                        "revision":      "",
                        "uuid":          uuid,
                    }

                    for k, v := range categoryData {
                        if k != "uuid" {
                            entity[k] = v
                        }
                    }

                    client.CallJSONRPC(ctx, "cmdevice", "updateCategory", entity, false)
                    time.Sleep(2 * time.Second)

                    t.Logf("[DEBUG] Modified notes externally to: %v", entity["notes"])
                },
                Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Original notes"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectNonEmptyPlan(),
                    },
                },
            },
            // Step 3: Terraform restores original notes
            {
                Config: testAccCMDeviceCategoryResourceConfig_DriftNotes(categoryName, "Original notes"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_cmdevice_category.test", "notes", "Original notes"),
                ),
            },
        },
    })
}
```

---

## Resources

- **Detailed Patterns**: `/workspace/CLAUDE.md` (Drift Detection Test Pattern)
- **BCM Specifics**: `/workspace/specs/006-test-review/bcm-test-patterns.md`
- **Test Helpers**: `/workspace/internal/provider/test_helpers.go`
- **Field Mappings**: See comments in `test_helpers.go:14-51`
- **Example Tests**:
  - `resource_cmpart_softwareimage_test.go:359` (DriftKernelParameters)
  - `resource_cmdevice_category_test.go:355` (DriftNotes)

---

## Troubleshooting

### Test fails with "plan is empty"

**Cause**: Drift not detected, external modification didn't work
**Fix**:
1. Verify BCM API call succeeded (check for errors)
2. Check field name is correct camelCase
3. Increase sleep time to 5 seconds
4. Add debug logging to verify modified value

### Test fails with "invalid result object"

**Cause**: Read operation has bugs or returns Unknown values
**Fix**: Check resource's Read implementation for Unknown value propagation

### Test fails with "resource not found"

**Cause**: Resource was deleted or name is wrong
**Fix**: Verify `getResourceUUIDByName` succeeded, check resource name

### Test hangs or times out

**Cause**: BCM API call not completing
**Fix**: Add timeout to context, check BCM endpoint is reachable

---

**Happy Testing!** 🚀

If you encounter issues not covered here, check:
1. Test output logs for BCM API errors
2. Field mapping tables for correct camelCase names
3. Existing tests for working examples
4. BCM API documentation in `sampleRest/` directory
