# Developer Quickstart: BCM Software Image Resource TDD Refactoring

**Feature**: Complete TDD-Based Review and Refactoring of resource_cmpart_softwareimage
**Date**: 2025-11-21
**Purpose**: Quick reference for executing TDD cycles during refactoring

## Prerequisites

### BCM Test Environment

```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

**Verify BCM connectivity**:
```bash
curl -k $BCM_ENDPOINT/json \
  -X POST \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"'$BCM_USERNAME'","password":"'$BCM_PASSWORD'"}'
```

### Go Environment

```bash
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
```

## TDD Workflow for Each User Story

### RED Phase: Write Failing Test

```bash
# 1. Open test file
vim internal/provider/resource_cmpart_softwareimage_test.go

# 2. Add new test function following this pattern:
# func TestAccCMPartSoftwareImageResource_<Feature>(t *testing.T) {
#   resource.Test(t, resource.TestCase{
#     PreCheck: func() { testAccPreCheck(t) },
#     ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
#     CheckDestroy: testAccCheckCMPartSoftwareImageDestroy,
#     Steps: []resource.TestStep{
#       {
#         Config: testAccCMPartSoftwareImageResourceConfig_<feature>(),
#         Check: resource.ComposeAggregateTestCheckFunc(
#           resource.TestCheckResourceAttr("bcm_cmpart_softwareimage.test", "attribute", "value"),
#         ),
#       },
#     },
#   })
# }

# 3. Run the test to verify it fails
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_<Feature>

# Expected output: FAIL with clear error message
# Document the error message in research.md or commit message
```

### GREEN Phase: Minimal Implementation

```bash
# 1. Open implementation file
vim internal/provider/resource_cmpart_softwareimage.go

# 2. Add MINIMAL code to make test pass
# - No optimization
# - No extra features
# - Just enough to satisfy the test assertions

# 3. Run the test to verify it passes
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_<Feature>

# Expected output: PASS
# Commit with message: "GREEN: Minimal implementation for <Feature>"
```

### REFACTOR Phase: Improve Code Quality

```bash
# 1. Improve the implementation
vim internal/provider/resource_cmpart_softwareimage.go

# Refactoring opportunities:
# - Extract helper functions
# - Add error handling
# - Improve diagnostics messages
# - Add logging (tflog.Debug, tflog.Trace, tflog.Info)
# - Add comments explaining complex logic

# 2. Run test again to verify still passing
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_<Feature>

# Expected output: PASS (no regressions)
# Commit with message: "REFACTOR: Improve <specific aspect> for <Feature>"

# 3. Run full test suite to verify no breakage
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource"
```

## Running Individual Tests

### Test US1 (Create with Clone)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic
```

### Test US2 (Read and Import)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic  # Includes import step
```

### Test US3 (Update Kernel Config)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_UpdateKernelConfig
```

### Test US3 (Update Modules)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_UpdateModules
```

### Test US3 (Update SOL)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_UpdateSOL
```

### Test US4 (Delete)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic  # Includes delete via CheckDestroy
```

### Test US5 (Async Clone)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic  # Includes clone polling
```

### Test US6 (Validation - Missing Required)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_MissingRequired
```

### Test US6 (Validation - Invalid SOL Speed)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_InvalidSOLSpeed
```

### Test US6 (Validation - Invalid Path)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_InvalidPath
```

### Test US7 (Unknown Values)
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_FullConfig
```

## Running Full Test Suite

### All Software Image Tests
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource"
```

### Parallel Execution (4 tests at once)
```bash
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource"
```

### With Coverage Report
```bash
TF_ACC=1 go test -v -cover -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource"
```

### JSON Output for CI
```bash
TF_ACC=1 go test -v -json -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource" > test-results.json
```

## Documentation Generation

### Generate Provider Documentation
```bash
cd /workspace
make generate

# Verify documentation was generated
ls -la docs/resources/cmpart_softwareimage.md
```

### Manual Documentation Generation (if make fails)
```bash
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go

/workspace/.go/bin/tfplugindocs generate \
  --provider-name bcm \
  --tf-version 1.13.5
```

### Verify Documentation is Up-to-Date
```bash
make generate
git diff --exit-code docs/resources/cmpart_softwareimage.md

# Exit code 0 = no changes (documentation current)
# Exit code 1 = changes detected (documentation outdated)
```

## Code Quality Checks

### Format Code
```bash
cd /workspace
make fmt

# Or manually:
gofmt -s -w -e internal/provider/resource_cmpart_softwareimage.go
gofmt -s -w -e internal/provider/resource_cmpart_softwareimage_test.go
```

### Run Linter
```bash
cd /workspace
make lint

# Or manually:
golangci-lint run internal/provider/resource_cmpart_softwareimage.go
golangci-lint run internal/provider/resource_cmpart_softwareimage_test.go
```

### Run Pre-Commit Hooks
```bash
cd /workspace

# First time setup
pre-commit install

# Run all hooks
pre-commit run --all-files

# Run specific hook
pre-commit run golangci-lint --all-files
```

## Debugging Tips

### Enable Detailed Terraform Logging
```bash
export TF_LOG=DEBUG
export TF_LOG_PATH=./terraform.log

# Run test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic

# View logs
cat terraform.log | grep "cmpart_softwareimage"
```

### Enable Trace-Level Provider Logging
```bash
export TF_LOG=TRACE
export TF_LOG_PATH=./terraform-trace.log

# Run test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic

# View detailed API calls
cat terraform-trace.log | grep "CallJSONRPC"
```

### Debug Single Test Iteration
```bash
# Run test with verbose output and stop on first failure
TF_ACC=1 go test -v -timeout 120m -failfast ./internal/provider/ \
  -run TestAccCMPartSoftwareImageResource_Basic
```

### Measure Test Execution Time
```bash
time TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource"

# Or use Go's built-in timing:
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ \
  -run "TestAccCMPartSoftwareImageResource" 2>&1 | grep "PASS\|FAIL"
```

## Test Data Strategy

### Unique Resource Naming
All tests use timestamp-based naming to prevent collisions:
```go
testImageName := fmt.Sprintf("test-image-%d", time.Now().Unix())
```

### Provider Configuration Template
```hcl
provider "bcm" {
  endpoint             = "<BCM_ENDPOINT>"
  username             = "<BCM_USERNAME>"
  password             = "<BCM_PASSWORD>"
  insecure_skip_verify = true
}
```

### Default Image Lookup Pattern
Tests use data source to find default image (no hardcoded UUIDs):
```hcl
data "bcm_cmpart_softwareimages" "all" {}

locals {
  default_image_uuid = data.bcm_cmpart_softwareimages.all.images[0].uuid
}

resource "bcm_cmpart_softwareimage" "test" {
  name           = "test-image"
  path           = "/cm/images/test-image"
  original_image = local.default_image_uuid
}
```

## Common Patterns

### Two-Step Create Pattern (for kernel config)
```hcl
# Step 1: Create with basic config
resource "bcm_cmpart_softwareimage" "test" {
  name           = "test-image"
  path           = "/cm/images/test-image"
  original_image = data.bcm_cmpart_softwareimages.all.images[0].uuid
}

# Step 2: Update with kernel config (after clone completes)
resource "bcm_cmpart_softwareimage" "test" {
  name           = "test-image"
  path           = "/cm/images/test-image"
  original_image = data.bcm_cmpart_softwareimages.all.images[0].uuid

  kernel_version        = "5.15.0-custom"
  kernel_parameters     = "quiet splash"
  kernel_output_console = "ttyS0,115200"
}
```

### Module Configuration
```hcl
modules = [
  {
    name       = "nvidia-drm"
    parameters = "modeset=1"
  },
  {
    name       = "nvme"
    # parameters omitted → becomes empty string ""
  }
]
```

### SOL Configuration
```hcl
enable_sol        = true
sol_speed         = "115200"  # OneOf validator
sol_port          = "0"
sol_flow_control  = true
```

## Performance Benchmarks

### Expected Test Times
- Basic create/read/delete: <30 seconds
- Full config with clone: <60 seconds
- Update operations: <10 seconds each
- Validation tests: <2 seconds each
- Full test suite (8 tests): <5 minutes

### Clone Operation Timing
- Fast clones: 2-4 seconds
- Normal clones: 8-12 seconds
- Slow clones: 15-20 seconds
- Timeout threshold: 47 seconds (soft timeout, logs warning)

## Troubleshooting

### Test Fails with "invalid result object"
**Cause**: Unknown value propagated to state
**Solution**: Check Unknown resolution in Create/Read/Update operations
```go
// Bad: Unknown in state
state.OriginalImage = plan.OriginalImage  // May be Unknown

// Good: Resolve Unknown to known value
if !plan.OriginalImage.IsUnknown() {
    state.OriginalImage = plan.OriginalImage
} else {
    state.OriginalImage = types.StringNull()  // Known null
}
```

### Test Fails with Clone Timeout
**Cause**: Clone operation exceeded 47 seconds
**Solution**: Check BCM cluster storage performance
```bash
# Monitor clone operation
watch 'curl -k $BCM_ENDPOINT/json -X POST -H "Content-Type: application/json" \
  -d "{\"service\":\"CMPart\",\"call\":\"getSoftwareImage\",\"args\":[\"test-image\"]}" \
  | jq ".fileOperationInProgress"'
```

### Test Fails with "Kernel version not found"
**Cause**: Attempting to set kernel_version before clone completes
**Solution**: Use two-step pattern (create basic, then update with kernel config)

### Test Cleanup Fails
**Cause**: Images not deleted after test
**Solution**: Check CheckDestroy function and verify BCM connectivity
```bash
# Manually clean up test images
curl -k $BCM_ENDPOINT/json -X POST -H "Content-Type: application/json" \
  -d '{"service":"CMPart","call":"getSoftwareImages","args":[]}}' \
  | jq '.[] | select(.name | startswith("test-")) | .uuid'
```

## Next Steps After Quickstart

1. Review research.md for TDD audit findings
2. Review data-model.md for entity mapping
3. Review contracts/ for API specifications
4. Follow tasks.md for phase-by-phase implementation
5. Maintain RED-GREEN-REFACTOR discipline for all changes
6. Document TDD cycle decisions in commit messages

## Resources

- **Terraform Plugin Framework Docs**: https://developer.hashicorp.com/terraform/plugin/framework
- **terraform-plugin-testing**: https://pkg.go.dev/github.com/hashicorp/terraform-plugin-testing
- **BCM API Documentation**: `/workspace/sampleRest/CMDevice_Complete_Documentation.md`
- **terraform-provider-design Skill**: Available in Claude Code
- **AGENTS.md**: TDD patterns and parallel execution strategies
