# Terraform Provider TDD Patterns

## TDD Philosophy for Providers

**RED-GREEN-REFACTOR in Parallel Batches**

Terraform provider development follows strict TDD cycles with parallel execution:
- 🔴 **RED**: Write failing acceptance tests first
- 🟢 **GREEN**: Write minimal CRUD code to pass tests
- 🔄 **REFACTOR**: Improve code while keeping tests green

## Testing Pyramid

```
     Acceptance Tests (Most Important)
           /        \
          /          \
         /____________\
        /   Unit Tests  \
       /__________________\
      / Integration Tests  \
     /______________________\
```

## Test-First Development Workflow

### 1. Write Test First (RED Phase)

Before writing any resource code:

1. Define what the resource should do
2. Write acceptance test that describes the behavior
3. Run test and verify it fails
4. Understand WHY it fails

### 2. Make It Pass (GREEN Phase)

Implement minimum code to pass:

1. Start with hardcoded values if needed
2. Focus solely on making test green
3. Don't optimize or add features
4. Verify test passes

### 3. Improve Code (REFACTOR Phase)

With passing tests:

1. Add real API integration
2. Improve error handling
3. Add validation
4. Optimize performance
5. Keep tests green throughout

## Acceptance Test Structure

### Required Test Steps

Every resource MUST test:

1. **Create and Read** - First step verifies creation and read
2. **ImportState** - Second step tests `terraform import`
3. **Update and Read** - Third step verifies updates
4. **Delete** - Automatically tested when TestCase completes

### Test Pattern Template

```go
func TestAccInstanceResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read testing
            {
                Config: testAccInstanceResourceConfig("test"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_instance.test", "name", "test"),
                    resource.TestCheckResourceAttrSet("bcm_instance.test", "id"),
                    resource.TestCheckResourceAttrSet("bcm_instance.test", "created_at"),
                ),
            },
            // ImportState testing
            {
                ResourceName:      "bcm_instance.test",
                ImportState:       true,
                ImportStateVerify: true,
                ImportStateVerifyIgnore: []string{"last_updated"},
            },
            // Update and Read testing
            {
                Config: testAccInstanceResourceConfig("test-updated"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("bcm_instance.test", "name", "test-updated"),
                ),
            },
        },
    })
}

func testAccInstanceResourceConfig(name string) string {
    return fmt.Sprintf(`
resource "bcm_instance" "test" {
  name = %[1]q
}
`, name)
}
```

## Minimal Implementation Pattern (GREEN Phase)

Start with hardcoded values to pass tests quickly:

```go
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data InstanceResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Minimal implementation - hardcoded ID
    data.ID = types.StringValue("instance-123")

    tflog.Trace(ctx, "created instance resource")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

## Refactoring Pattern (REFACTOR Phase)

Add real API integration after tests pass:

```go
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data InstanceResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Make actual API call
    instance, err := r.client.CreateInstance(ctx, data.Name.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating Instance",
            "Could not create instance, unexpected error: "+err.Error(),
        )
        return
    }

    data.ID = types.StringValue(instance.ID)
    tflog.Trace(ctx, "created instance resource")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

## State Check Functions

### Common Check Patterns

```go
// Exact value match
resource.TestCheckResourceAttr("bcm_instance.test", "name", "expected-value")

// Attribute exists with any value
resource.TestCheckResourceAttrSet("bcm_instance.test", "id")

// Match two resources' attributes
resource.TestCheckResourceAttrPair(
    "bcm_instance.test", "vpc_id",
    "bcm_vpc.test", "id",
)

// Combine multiple checks (continue on failure)
resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr(...),
    resource.TestCheckResourceAttrSet(...),
)
```

## Provider Test Setup

```go
// internal/provider/provider_test.go
package provider

import (
    "testing"
    "github.com/hashicorp/terraform-plugin-framework/providerserver"
    "github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "bcm": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
    // Check for required environment variables
    // if v := os.Getenv("BCM_API_KEY"); v == "" {
    //     t.Fatal("BCM_API_KEY must be set for acceptance tests")
    // }
}
```

## Running Tests

### Unit Tests
```bash
go test -v ./...
go test -race ./...
```

### Acceptance Tests
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./...
```

### With Logging
```bash
TF_LOG=TRACE TF_ACC=1 go test -v ./internal/provider/
```

## Test Quality Standards

- **Coverage**: All CRUD operations tested
- **Import**: All resources must be importable
- **Edge Cases**: Error handling validated
- **Pass Rate**: 100% required
- **Execution Time**: <120m for full acceptance suite
- **Parallel Execution**: 4-8 parallel tests recommended

## Security Testing

Mark sensitive attributes appropriately:

```go
func (r *APIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        Attributes: map[string]schema.Attribute{
            "key_value": schema.StringAttribute{
                Computed:  true,
                Sensitive: true, // Prevents display in output
            },
        },
    }
}
```

## Data Source Testing

```go
func TestAccUserDataSource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccUserDataSourceConfig,
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("data.bcm_user.test", "username", "admin"),
                    resource.TestCheckResourceAttrSet("data.bcm_user.test", "id"),
                ),
            },
        },
    })
}

const testAccUserDataSourceConfig = `
data "bcm_user" "test" {
  username = "admin"
}
`
```

## Test Verification Best Practices

### What to Verify in Tests

**Basic Resource Test Should:**
1. Plan and apply configuration without error
2. Verify expected attributes saved to state
3. Verify values match remote API/service
4. Verify subsequent plan produces no diff

**Update Test Should:**
1. Apply initial configuration
2. Apply modified configuration
3. Verify updates reflected in state
4. Verify updates reflected in remote API

### CheckDestroy Function

Always implement CheckDestroy to verify cleanup:

```go
func testAccCheckExampleResourceDestroy(s *terraform.State) error {
    // Retrieve API client from provider
    // client := testAccProvider.Meta().(*APIClient)

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "example_resource" {
            continue
        }

        // Try to find the resource
        _, err := client.GetResource(rs.Primary.ID)
        if err == nil {
            return fmt.Errorf("Resource %s still exists", rs.Primary.ID)
        }

        // Verify it's a "not found" error
        if !isNotFoundError(err) {
            return err
        }
    }

    return nil
}
```

### Testing Edge Cases

```go
// Test error handling
func TestAccExampleResource_InvalidConfig(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config:      testAccExampleResourceConfig_Invalid(),
                ExpectError: regexp.MustCompile("invalid configuration"),
            },
        },
    })
}

// Test disappears (external deletion)
func TestAccExampleResource_Disappears(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            {
                Config: testAccExampleResourceConfig("test"),
                Check: resource.ComposeTestCheckFunc(
                    testAccCheckExampleResourceExists("example_resource.test"),
                    testAccCheckExampleResourceDisappears("example_resource.test"),
                ),
                ExpectNonEmptyPlan: true,
            },
        },
    })
}
```

## Common Anti-Patterns to Avoid

❌ **Skipping ImportState Tests** - Always test import functionality
❌ **Hardcoded Test Values** - Use unique resource names per test run
❌ **Incomplete CRUD** - Test all Create, Read, Update, Delete operations
❌ **Ignoring Error Cases** - Test API failures and invalid inputs
❌ **Missing Documentation** - Keep examples/ and docs/ in sync with code
❌ **Not Testing State Drift** - Verify Read correctly detects external changes
❌ **Brittle Tests** - Don't depend on external state or ordering
❌ **No CheckDestroy** - Always verify resources are destroyed
❌ **Skipping Edge Cases** - Test disappears, conflicts, invalid configs
❌ **Poor Test Isolation** - Each test should be fully independent
