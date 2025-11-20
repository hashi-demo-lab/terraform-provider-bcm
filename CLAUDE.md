# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Terraform Provider for Nvidia BCM (Bright Cluster Manager) built on the Terraform Plugin Framework v1.16.1. It follows TDD (Test-Driven Development) principles with parallel execution patterns for resources, data sources, and ephemeral resources.

## Development Commands

### Building and Installing

- `make install` - Build and install the provider (default: fmt, lint, install, generate)
- `make build` - Build provider without installing
- `go install` - Direct install to $GOPATH/bin

### Testing

**Unit Tests:**
- `make test` - Run unit tests (120s timeout, parallel=10)
- `go test -v -cover ./internal/provider/` - Run with coverage

**Acceptance Tests:**
- `make testacc` - Run all acceptance tests (requires TF_ACC=1, 120m timeout)
- `TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAcc<ResourceName>` - Run specific test

**Environment Variables for Acceptance Tests:**
```bash
export TF_ACC=1                                    # Enable acceptance tests
export BCM_ENDPOINT="https://172.21.15.254:8081"  # BCM API endpoint
export BCM_USERNAME="root"                         # BCM username
export BCM_PASSWORD="Hashicorp123!"                # BCM password
```

**Run Single Test:**
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ -run TestAccCMDeviceNodesDataSource_Basic
```

### Code Quality

- `make fmt` - Format Go code (`gofmt -s -w -e .`)
- `make lint` - Run golangci-lint
- `pre-commit install` - Install pre-commit hooks (first time)
- `pre-commit run --all-files` - Run all pre-commit hooks

### Documentation Generation

- `make generate` - Generate provider documentation
  - Runs `cd tools; go generate ./...`
  - Generates copyright headers via copywrite
  - Generates Terraform docs via tfplugindocs
  - Formats examples/ with terraform fmt

**Manual Documentation Generation:**
```bash
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
/workspace/.go/bin/tfplugindocs generate --provider-name bcm --tf-version 1.13.5
```

## Project Structure

```
terraform-provider-bcm/
├── internal/provider/
│   ├── provider.go                    # Provider configuration
│   ├── bcm_client.go                  # JSON-RPC API client
│   ├── data_source_*.go               # Data source implementations
│   ├── data_source_*_test.go          # Acceptance tests
│   └── helpers.go                     # Null-safe helper functions
├── examples/
│   ├── provider/                      # Provider config examples
│   ├── data-sources/                  # Data source examples
│   └── resources/                     # Resource examples
├── docs/                              # Generated docs (DON'T EDIT)
├── specs/                             # Speckit specifications
├── sampleRest/                        # API exploration scripts
├── tools/                             # Code generation tools
└── main.go                            # Provider entry point
```

## Architecture

### BCM API Client (`internal/provider/bcm_client.go`)

**Authentication:**
- Cookie-based authentication using `cm-login-token`
- Automatic cookie management via `http.Client` with cookiejar
- Login performed once during provider initialization

**JSON-RPC Pattern:**
```go
// All BCM API calls use this pattern:
{
  "service": "cmdevice",  // Service name
  "call": "getNodes"      // Method name
  // Note: "arg" parameter support deferred to post-POV
}
```

**Key Methods:**
- `NewBCMClient()` - Creates authenticated client
- `CallJSONRPC(ctx, service, method)` - Makes API calls
- `parseErrorResponse()` - Multi-layer error detection

### Data Source Pattern

All data sources follow this pattern:

1. **Schema Definition** - Define attributes with descriptions
2. **API Call** - Use `client.CallJSONRPC()` with hardcoded service/method
3. **JSON Unmarshaling** - Parse into `[]map[string]interface{}`
4. **Data Mapping** - Use helper functions (`getStringValue`, `getBoolValue`, etc.)
5. **Client-Side Filtering** - Filter results in Go, not API
6. **State Management** - Set computed attributes using terraform-plugin-framework types

**Helper Functions for Null-Safety:**
- `getStringValue(data, key)` - Returns `types.String` with null handling
- `getBoolValue(data, key)` - Returns `types.Bool` with null handling
- `getInt64Value(data, key)` - Returns `types.Int64` with null handling

### Test Configuration Pattern

Acceptance tests must include provider configuration:

```go
func testAccConfigFunction() string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_resource" "test" {
  # ...
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
    )
}
```

## TDD Workflow

This project follows **RED-GREEN-REFACTOR** cycles with parallel execution:

1. **RED**: Write failing acceptance tests first (`*_test.go` files)
2. **GREEN**: Write minimal implementation to pass tests
3. **REFACTOR**: Improve code quality while keeping tests green
4. **DOCUMENT**: Run `make generate` to update documentation

### Speckit Workflow (Feature Development)

For new features, use the Speckit workflow commands:

1. `/speckit.specify` - Create feature specification from natural language
2. `/speckit.clarify` - Ask clarification questions about underspecified areas
3. `/speckit.plan` - Generate implementation plan with design artifacts
4. `/speckit.tasks` - Generate actionable task list
5. `/speckit.analyze` - Analyze spec for TDD compliance
6. `/speckit.implement` - Execute all tasks to implement the feature

**Speckit Artifacts Location:** `/workspace/specs/<feature-name>/`
- `spec.md` - Feature requirements and API contracts
- `plan.md` - Implementation design
- `tasks.md` - Task breakdown
- `quickstart.md` - Developer quick start guide

## Key Dependencies

- **Framework**: `terraform-plugin-framework` (v1.16.1)
- **Testing**: `terraform-plugin-testing` (v1.13.3)
- **Documentation**: `terraform-plugin-docs` (via tools/tools.go)
- **Go Version**: 1.24.0
- **API**: BCM JSON-RPC (cookie-based auth)

## Primary Reference for TDD Patterns

See `./AGENTS.md` for comprehensive Terraform Provider TDD patterns, parallel execution strategies, and development best practices.

## Additional References

- **Speckit Constitution**: `.specify/memory/constitution.md` - TDD rules
- **Skills Available**:
  - `terraform-provider-design` - Provider design principles
  - `terraform-test` - Terraform test file patterns
  - `terraform-style-guide` - HashiCorp style guide
  - `terraform-stacks` - Terraform Stacks patterns
- **BCM API Documentation**: `sampleRest/CMDevice_Complete_Documentation.md`

## BCM-Specific Notes

### API Characteristics

- **Endpoint**: `https://172.21.15.254:8081/json`
- **Authentication**: POST to `/json` with `{"service":"login","username":"...","password":"..."}`
- **Session**: Cookie `cm-login-token` managed automatically
- **Services**: CMDevice (nodes), CMPart (software), CMNet (networks), etc.
- **Response Format**: JSON arrays of objects with polymorphic `childType` field

### Known Limitations (POV Scope)

- **No "arg" parameter support** - Only parameterless API calls (documented in `bcm_client.go:27-33`)
- **Client-side filtering only** - All filtering happens in Go after API response
- **Self-signed certs** - `insecure_skip_verify = true` required for local dev

### Common Development Patterns

**Adding a New Data Source:**

1. Create `internal/provider/data_source_<name>.go`
2. Define schema with `schema.Schema{}`
3. Implement `Read()` method using `client.CallJSONRPC()`
4. Use helper functions for null-safe field extraction
5. Register in `provider.go` `DataSources()` method
6. Create acceptance test in `internal/provider/data_source_<name>_test.go`
7. Add examples in `examples/data-sources/<name>/`
8. Run `make generate` to create documentation

**Test Pattern:**
- Test configs must be functions (not const strings) to inject provider config
- Use environment variables for credentials
- Include all CRUD steps: Create/Read, Import, Update/Read, Delete
- Verify both existence (`TestCheckResourceAttrSet`) and values (`TestCheckResourceAttr`)

## Important Notes

- **Parallel execution**: Use concurrent tool calls for independent operations
- **AskUserQuestion**: Use for clarification (especially during `/speckit.clarify`)
- **Pre-commit hooks**: Run before commits for code quality
- **Acceptance tests**: Set `TF_ACC=1` (creates real resources, may cost money)
- **Documentation**: Auto-generated in `docs/` - don't edit manually
- **Local testing**: BCM cluster at 172.21.15.254 for acceptance tests

## Troubleshooting

**Provider not found during doc generation:**
```bash
# Build provider for version expected by tfplugindocs
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bcm/0.0.1/linux_amd64
GOOS=linux GOARCH=amd64 go build -o ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bcm/0.0.1/linux_amd64/terraform-provider-bcm_v0.0.1
```

**Go module cache permissions:**
```bash
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
```

**Acceptance test authentication failures:**
- Verify BCM credentials (check `sampleRest/*.py` for working examples)
- Ensure BCM endpoint is reachable
- Check `cm-login-token` cookie is being set

## Updating This File

When you discover new implementation details, debugging insights, or architectural patterns:

- Update `CLAUDE.md` for project-level guidance
- Update `AGENTS.md` for TDD and parallel execution patterns
- Create component-specific `AGENTS.md` files in subdirectories for specialized guidance
