---
name: terraform-provider-design
description: Comprehensive guide for designing and implementing Terraform providers using Test-Driven Development (TDD). Use when creating new Terraform providers, adding resources/data sources to existing providers, implementing TDD workflows (RED-GREEN-REFACTOR), designing provider architecture, or working in terraform-provider-* directories. Covers HashiCorp best practices, Plugin Framework patterns, acceptance testing, CRUD implementations, and parallel development workflows.
---

# Terraform Provider Design

## Overview

This skill guides Terraform provider development using Test-Driven Development (TDD) following HashiCorp's official best practices. It supports the complete RED-GREEN-REFACTOR cycle with parallel execution patterns for efficient provider development.

## When to Use This Skill

Use this skill when:

- Creating a new Terraform provider from scratch
- Adding resources or data sources to an existing provider
- Implementing TDD workflows for provider development
- Designing provider architecture and schemas
- Writing acceptance tests for provider resources
- Reviewing provider code for HashiCorp compliance

## Core TDD Workflow

### The RED-GREEN-REFACTOR Cycle

Provider development follows strict TDD phases executed in parallel where possible:

**🔴 RED Phase**: Write failing acceptance tests

- Define resource/data source behavior through tests
- Write Terraform configuration fixtures
- Verify tests fail for the right reasons

**🟢 GREEN Phase**: Write minimal CRUD code

- Implement simplest code to pass tests
- Use hardcoded values initially if needed
- Focus on making tests pass quickly

**🔄 REFACTOR Phase**: Improve implementation

- Add real API integration
- Enhance error handling
- Optimize code quality
- Keep all tests passing

### Parallel Execution Pattern

Execute multiple TDD cycles concurrently:

```bash
# RED PHASE - Write multiple failing tests in parallel
Write("internal/provider/instance_resource_test.go")
Write("internal/provider/network_resource_test.go")
Write("internal/provider/user_data_source_test.go")
Bash("TF_ACC=1 go test -v -timeout 120m ./internal/provider/")

# GREEN PHASE - Implement minimal CRUD in parallel
Write("internal/provider/instance_resource.go")
Write("internal/provider/network_resource.go")
Write("internal/provider/user_data_source.go")
Bash("TF_ACC=1 go test -v -timeout 120m ./internal/provider/")

# REFACTOR PHASE - Improve implementations in parallel
Edit("internal/provider/instance_resource.go")
Edit("internal/provider/network_resource.go")
Edit("internal/provider/user_data_source.go")
Bash("TF_ACC=1 go test -v -timeout 120m ./internal/provider/")
```

## Getting Started

### Project Initialization

Create a new provider project:

```bash
# Initialize Go module
go mod init github.com/yourusername/terraform-provider-{name}

# Add required dependencies
go get github.com/hashicorp/terraform-plugin-framework
go get github.com/hashicorp/terraform-plugin-go
go get github.com/hashicorp/terraform-plugin-log
go get github.com/hashicorp/terraform-plugin-testing

# Create directory structure
mkdir -p internal/provider examples/{provider,resources,data-sources} docs
```

### Provider Bootstrap Template

See `assets/provider_template.go` for a complete provider initialization example.

## Provider Structure

### Standard Directory Layout

```
terraform-provider-{name}/
├── internal/
│   └── provider/
│       ├── provider.go              # Provider definition
│       ├── provider_test.go         # Provider tests
│       ├── resource_*.go            # Resource implementations
│       ├── resource_*_test.go       # Resource acceptance tests
│       ├── data_source_*.go         # Data source implementations
│       └── data_source_*_test.go    # Data source acceptance tests
├── examples/
│   ├── provider/                    # Provider configuration examples
│   ├── resources/                   # Resource examples
│   └── data-sources/                # Data source examples
├── docs/                            # Generated documentation
├── main.go                          # Provider binary entry point
├── go.mod                           # Go module dependencies
├── CHANGELOG.md                     # Version history
├── .goreleaser.yml                  # Release automation
└── GNUmakefile                      # Build and test commands
```

## Design Principles

Follow HashiCorp's core design principles (see `references/hashicorp_best_practices.md` for details):

1. **Single API Focus**: One provider per API/service domain
2. **Single Object per Resource**: One API object per Terraform resource
3. **Schema Alignment**: Match underlying API unless it degrades UX
4. **Import Support**: All resources must support `terraform import`
5. **Version Continuity**: Maintain backward compatibility

## Resource Design Workflow

### Step 1: Define Resource Schema

Start with the schema that matches the underlying API:

```go
func (r *InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
    resp.Schema = schema.Schema{
        MarkdownDescription: "Instance resource",
        Attributes: map[string]schema.Attribute{
            "id": schema.StringAttribute{
                Computed:            true,
                MarkdownDescription: "Instance identifier",
            },
            "name": schema.StringAttribute{
                Required:            true,
                MarkdownDescription: "Instance name",
            },
            "tags": schema.MapAttribute{
                Optional:            true,
                ElementType:         types.StringType,
                MarkdownDescription: "Resource tags",
            },
        },
    }
}
```

### Step 2: Write Acceptance Test (RED)

Create failing test before implementation:

```go
func TestAccInstanceResource(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and Read
            {
                Config: testAccInstanceResourceConfig("test"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("example_instance.test", "name", "test"),
                    resource.TestCheckResourceAttrSet("example_instance.test", "id"),
                ),
            },
            // ImportState
            {
                ResourceName:      "example_instance.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
            // Update and Read
            {
                Config: testAccInstanceResourceConfig("test-updated"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("example_instance.test", "name", "test-updated"),
                ),
            },
        },
    })
}
```

### Step 3: Minimal Implementation (GREEN)

Implement simplest CRUD to pass tests:

```go
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data InstanceResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Minimal - hardcoded for now
    data.ID = types.StringValue("instance-123")

    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

### Step 4: Full Implementation (REFACTOR)

Add real API integration:

```go
func (r *InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
    var data InstanceResourceModel
    resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
    if resp.Diagnostics.HasError() {
        return
    }

    // Real API call
    instance, err := r.client.CreateInstance(ctx, data.Name.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Creating Instance",
            "Could not create instance: "+err.Error(),
        )
        return
    }

    data.ID = types.StringValue(instance.ID)
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}
```

## Naming Conventions

Follow HashiCorp naming standards:

**Resources**: Singular nouns with provider prefix

- Format: `{provider}_{resource_name}`
- Example: `aws_instance`, `postgresql_database`

**Data Sources**: Nouns (plural for collections)

- Format: `{provider}_{data_source_name}`
- Example: `aws_availability_zones`, `azurerm_subnet`

**Attributes**: Lowercase with underscores

- Single values: Singular nouns (`instance_type`)
- Collections: Plural nouns (`security_group_ids`)
- Booleans: Nouns describing what's enabled (`auto_scaling_enabled`)

**Functions**: Verbs without provider prefix

- Format: `{verb}_{object}`
- Example: `parse_rfc3339`, `encode_base64`

## Testing Requirements

### Acceptance Test Coverage

Every resource MUST test:

1. ✅ Create and Read operations
2. ✅ ImportState functionality
3. ✅ Update and Read operations
4. ✅ Delete operations (automatic)

### Testing Pattern Categories

HashiCorp recommends four core testing patterns (see [Testing Patterns](https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns)):

1. **Basic Attribute Verification** - Configuration applies, attributes persist with correct values
2. **Update Tests** - Resources correctly apply updates while maintaining state consistency
3. **Import Mode Testing** - Verify import behavior with `ImportState: true`
4. **Error and Plan Expectations** - Validate expected failures and plan outcomes

### State Checks

State checks validate resource attributes in Terraform state after `terraform apply`. They execute during Lifecycle (config) mode and provide comprehensive attribute verification.

**Key Characteristics**:

- Multiple state checks can run with aggregated error reporting
- Cleanup executes even when checks fail
- Primarily for validating computed attributes

**ExpectKnownValue - Verify Attribute Type and Value**:

```go
import (
    "github.com/hashicorp/terraform-plugin-testing/statecheck"
    "github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
    "github.com/hashicorp/terraform-plugin-testing/knownvalue"
)

// Basic attribute verification
StateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "example_instance.test",
        tfjsonpath.New("name"),
        knownvalue.StringExact("test-instance"),
    ),
    statecheck.ExpectKnownValue(
        "example_instance.test",
        tfjsonpath.New("enabled"),
        knownvalue.Bool(true),
    ),
}

// Nested attribute with map key access
StateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "example_instance.test",
        tfjsonpath.New("tags").AtMapKey("environment"),
        knownvalue.StringExact("production"),
    ),
}
```

**CompareValue - Track Attribute Changes Across Steps**:

```go
// Compare computed values across test steps
compareValuesSame := statecheck.CompareValue(compare.ValuesSame())

// First test step - capture initial value
{
    Config: testAccInstanceResourceConfig("test"),
    ConfigStateChecks: []statecheck.StateCheck{
        compareValuesSame.AddStateValue(
            "example_instance.test",
            tfjsonpath.New("computed_id"),
        ),
    },
}

// Second test step - verify value hasn't changed
{
    Config: testAccInstanceResourceConfig("test-updated"),
    ConfigStateChecks: []statecheck.StateCheck{
        compareValuesSame.AddStateValue(
            "example_instance.test",
            tfjsonpath.New("computed_id"),
        ),
    },
}
```

**Known Value Types**:

```go
// Boolean values
knownvalue.Bool(true)

// String values
knownvalue.StringExact("exact-match")

// Numeric values
knownvalue.Int64Exact(42)
knownvalue.Float64Exact(3.14)

// Null and existence checks
knownvalue.Null()
knownvalue.NotNull()

// Collections
knownvalue.ListExact([]knownvalue.Check{
    knownvalue.StringExact("item1"),
    knownvalue.StringExact("item2"),
})

knownvalue.MapExact(map[string]knownvalue.Check{
    "key1": knownvalue.StringExact("value1"),
    "key2": knownvalue.Int64Exact(100),
})

knownvalue.SetExact([]knownvalue.Check{
    knownvalue.StringExact("elem1"),
    knownvalue.StringExact("elem2"),
})
```

**Sensitive Value Verification (Terraform 1.4.6+)**:

```go
// Verify attribute is marked sensitive
StateChecks: []statecheck.StateCheck{
    statecheck.ExpectSensitiveValue(
        "example_instance.test",
        tfjsonpath.New("api_key"),
    ),
}
```

**State Check Best Practices**:

- Validate computed attributes, not user-provided configuration
- Use `ExpectKnownValue` for exact value matching
- Use `CompareValue` to track consistency across test steps
- Use `ExpectSensitiveValue` for sensitive attributes
- Leverage JSON paths for nested attribute access

### Plan Checks

Plan checks validate Terraform plan outcomes during test execution. They ensure configuration changes produce expected planning results.

**Key Characteristics**:

- Multiple checks per phase with aggregated error reporting
- Cleanup executes even when checks fail
- Available in Lifecycle (config) and Refresh test modes

**Built-in Plan Check Functions**:

```go
import "github.com/hashicorp/terraform-plugin-testing/plancheck"

// ExpectEmptyPlan - Asserts no operations for apply
// Use when: Verifying idempotency or that updates produce no changes
ConfigPlanChecks: resource.ConfigPlanChecks{
    PreApply: []plancheck.PlanCheck{
        plancheck.ExpectEmptyPlan(),
    },
}

// ExpectNonEmptyPlan - Asserts at least one operation for apply
// Use when: Confirming configuration changes trigger updates
ConfigPlanChecks: resource.ConfigPlanChecks{
    PreApply: []plancheck.PlanCheck{
        plancheck.ExpectNonEmptyPlan(),
    },
}
```

**Plan Check Example**:

```go
func TestAccInstanceResource_Idempotent(t *testing.T) {
    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create
            {
                Config: testAccInstanceResourceConfig("test"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("example_instance.test", "name", "test"),
                ),
            },
            // Verify idempotency - no plan changes
            {
                Config: testAccInstanceResourceConfig("test"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
        },
    })
}
```

**Plan Check Best Practices**:

- Use `ExpectEmptyPlan()` to verify idempotency after resource creation
- Use `ExpectNonEmptyPlan()` to confirm updates actually trigger changes
- Combine with drift detection tests to verify Read operations
- Available check types: General, Resource, Output, and Custom plan checks

### Comprehensive Testing Example

Combining state checks, plan checks, and traditional checks:

```go
func TestAccInstanceResource_Complete(t *testing.T) {
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        PreCheck:                 func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        Steps: []resource.TestStep{
            // Create and verify
            {
                Config: testAccInstanceResourceConfig("test"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("example_instance.test", "name", "test"),
                    resource.TestCheckResourceAttrSet("example_instance.test", "id"),
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "example_instance.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact("test"),
                    ),
                    compareID.AddStateValue(
                        "example_instance.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Verify idempotency
            {
                Config: testAccInstanceResourceConfig("test"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
            // Update and verify ID remains same
            {
                Config: testAccInstanceResourceConfig("test-updated"),
                Check: resource.ComposeAggregateTestCheckFunc(
                    resource.TestCheckResourceAttr("example_instance.test", "name", "test-updated"),
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "example_instance.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact("test-updated"),
                    ),
                    // Verify ID hasn't changed
                    compareID.AddStateValue(
                        "example_instance.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
        },
    })
}
```

### Regression Testing Pattern

When fixing bugs, follow this workflow:

1. **Create regression test** - Write failing test that reproduces the bug
2. **Commit test first** - Commit shows problem independently verified
3. **Implement fix** - Subsequent commits fix the issue
4. **Verify test passes** - Confirms fix resolves the problem

This allows independent verification of problem reproduction before evaluating solutions

### Running Tests

```bash
# Unit tests
go test -v ./...

# Acceptance tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Parallel execution
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./...

# With detailed logging
TF_LOG=TRACE TF_ACC=1 go test -v ./internal/provider/
```

### Quality Standards

- **Pass Rate**: 100% required
- **Coverage**: All CRUD operations
- **Import**: All resources importable
- **Execution Time**: <120m for full suite
- **Parallel Tests**: 4-8 concurrent recommended

## Sensitive Data Handling

### Recommended: Ephemeral Resources

For sensitive data like tokens or secrets:

```go
// Use ephemeral resources (Plugin Framework only)
// Data not persisted in state
```

### Alternative: Sensitive Flag

Mark sensitive attributes:

```go
"api_key": schema.StringAttribute{
    Required:  true,
    Sensitive: true, // Prevents display in output
}
```

**Note**: Sensitive flag does NOT encrypt state files. Use remote backends with encryption at rest.

## Documentation Generation

Generate provider documentation automatically:

```bash
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
```

Ensure `examples/` directory contains:

- `provider/provider.tf` - Provider configuration
- `resources/{resource_name}/resource.tf` - Resource examples
- `data-sources/{data_source_name}/data-source.tf` - Data source examples

## Versioning

Follow Semantic Versioning (`MAJOR.MINOR.PATCH`):

**MAJOR** - Breaking changes:

- Removing/renaming resources or attributes
- Changing authentication patterns
- Incompatible type changes

**MINOR** - New features:

- Adding resources or attributes
- Deprecation warnings
- Compatible type changes

**PATCH** - Bug fixes only

**Recommendation**: Major versions no more than once per year

## CI/CD and Tooling

### Continuous Integration

Set up automated testing with GitHub Actions:

```yaml
# .github/workflows/test.yml
name: Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v4
        with:
          go-version: "1.21"
      - run: go test -v -cover ./...
      - run: TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

### Code Quality and Release

See `references/ci_cd_patterns.md` for:

- golangci-lint configuration
- GoReleaser setup for provider releases
- Pre-commit hooks
- Documentation generation automation

## Testing Best Practices

### Test Environment Setup

Create isolated test environments:

```go
func testAccPreCheck(t *testing.T) {
    // Verify API credentials
    if v := os.Getenv("EXAMPLE_API_KEY"); v == "" {
        t.Fatal("EXAMPLE_API_KEY must be set for acceptance tests")
    }

    // Verify API endpoint
    if v := os.Getenv("EXAMPLE_API_ENDPOINT"); v == "" {
        t.Skip("EXAMPLE_API_ENDPOINT not set, skipping acceptance tests")
    }
}
```

### Test Safety

**CRITICAL**: Use separate accounts/namespaces for testing:

- Prevents accidental production infrastructure changes
- Allows safe parallel test execution
- Enables test cleanup without risk

### ID Attribute Requirement

For terraform-plugin-testing v1.5.0+, always add root-level `id` attribute:

```go
"id": schema.StringAttribute{
    Computed:            true,
    MarkdownDescription: "Resource identifier",
},
```

### Parallel Testing Configuration

```bash
# Run tests in parallel (recommended 4-8)
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./...

# Run specific test
TF_ACC=1 go test -v -run TestAccExampleResource ./internal/provider/
```

## Resources

### Templates (`assets/`)

- `provider_template.go` - Provider initialization and configuration template
- `main_template.go` - Provider binary entry point template
- `resource_template.go` - Complete resource implementation template
- `data_source_template.go` - Data source implementation template
- `acceptance_test_template.go` - Acceptance test template with all required steps

### References (`references/`)

- `hashicorp_best_practices.md` - Official HashiCorp design principles and standards
- `tdd_patterns.md` - Detailed TDD patterns, test structures, and anti-patterns
- `ci_cd_patterns.md` - CI/CD workflows, automation, tooling, and release management

## Quick Reference

### Provider Test Setup

```go
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "example": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t *testing.T) {
    // Verify required environment variables
}
```

### State Check Functions

```go
// Exact match
resource.TestCheckResourceAttr("example_instance.test", "name", "expected")

// Attribute exists
resource.TestCheckResourceAttrSet("example_instance.test", "id")

// Match between resources
resource.TestCheckResourceAttrPair("example_instance.test", "vpc_id", "example_vpc.test", "id")

// Combine checks
resource.ComposeAggregateTestCheckFunc(
    resource.TestCheckResourceAttr(...),
    resource.TestCheckResourceAttrSet(...),
)
```

### Common Anti-Patterns to Avoid

❌ Skipping ImportState tests
❌ Skipping Drift tests
❌ Skipping idempotency tests with plan checks
❌ Using hardcoded test values (use unique names per test run)
❌ Incomplete CRUD testing
❌ Ignoring error cases
❌ Missing or outdated documentation
❌ Not testing state drift
❌ Tests dependent on external state
❌ Not verifying plan outcomes with ExpectEmptyPlan/ExpectNonEmptyPlan
❌ Using TestCheckResourceAttr for computed values (use StateChecks instead)
❌ Not tracking attribute consistency across test steps (use CompareValue)
❌ Missing sensitive attribute verification (use ExpectSensitiveValue)

## Development Checklist

- [ ] Schema aligns with underlying API
- [ ] Resource supports import (ImportState tests)
- [ ] All CRUD operations tested
- [ ] State checks verify computed attributes (ExpectKnownValue)
- [ ] Attribute consistency tracked across steps (CompareValue)
- [ ] Idempotency verified with plan checks (ExpectEmptyPlan)
- [ ] Drift detection tests implemented
- [ ] Sensitive attributes verified (ExpectSensitiveValue)
- [ ] Error cases tested (ExpectError)
- [ ] Acceptance tests pass 100%
- [ ] Documentation generated
- [ ] CHANGELOG updated
- [ ] Examples provided
- [ ] Sensitive data handled properly
- [ ] Version follows semantic versioning
- [ ] Code follows HashiCorp conventions
