# Quickstart: BCM User Resource Development

**Feature Branch**: `068-cmuser-user-resource`
**Date**: 2025-11-26

## Overview

This guide provides quick-start instructions for developing, testing, and debugging the `bcm_cmuser_user` Terraform resource.

---

## Environment Setup

### Prerequisites

1. **Go 1.24.0+** installed
2. **Terraform 1.13.5+** installed
3. **BCM cluster access** with API credentials

### Environment Variables

```bash
# BCM API credentials (required for acceptance tests)
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Go environment (for doc generation)
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go

# Enable acceptance tests
export TF_ACC=1
```

### Verify Setup

```bash
# Check Go version
go version  # Should show go1.24.0 or later

# Check Terraform version
terraform version  # Should show 1.13.5 or later

# Test BCM connectivity
curl -k -X POST "${BCM_ENDPOINT}/json" \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"'${BCM_USERNAME}'","password":"'${BCM_PASSWORD}'"}'
# Should return: true
```

---

## Development Workflow

### TDD Cycle: RED-GREEN-REFACTOR

#### 1. RED Phase: Write Failing Test

```bash
# Create test file
touch internal/provider/resource_cmuser_user_test.go

# Run test (should fail - resource doesn't exist yet)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Basic
```

#### 2. GREEN Phase: Minimal Implementation

```bash
# Create resource file
touch internal/provider/resource_cmuser_user.go

# Implement minimal CRUD to pass test

# Run test (should pass)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Basic
```

#### 3. REFACTOR Phase: Production Quality

```bash
# Add validation, error handling, edge cases

# Run all tests (should still pass)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser

# Run linter
make lint

# Generate docs
make generate
```

---

## Running Tests

### Single Test

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Basic
```

### All Resource Tests

```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser
```

### With Debug Logging

```bash
TF_LOG=DEBUG TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Basic
```

### With Trace Logging (Maximum Verbosity)

```bash
TF_LOG=TRACE TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUser_Basic
```

---

## Example Terraform Configurations

### Basic User Creation

```hcl
provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "example" {
  username = "testuser"
  password = var.user_password
}
```

### User with Full Configuration

```hcl
resource "bcm_cmuser_user" "developer" {
  username       = "developer01"
  password       = var.developer_password
  full_name      = "Developer One"
  email          = "dev01@example.com"
  home_directory = "/home/developer01"
  shell          = "/bin/zsh"
  notes          = "DGX BasePOD developer account"

  authorized_ssh_keys = <<-EOT
    ssh-rsa AAAAB3Nza... user@example.com
    ssh-ed25519 AAAAC3Nza... user@laptop
  EOT

  shadow_max     = 90
  shadow_warning = 14
}
```

### Import Existing User

```bash
terraform import bcm_cmuser_user.existing cmsupport
```

---

## Debugging Tips

### API Response Inspection

```go
// Add to test file for debugging API responses
func TestDebugCMUserAPI(t *testing.T) {
    client := createTestBCMClient(t)
    ctx := context.Background()

    // Get existing user
    body, err := client.CallJSONRPC(ctx, "cmuser", "getUser", "cmsupport")
    if err != nil {
        t.Fatalf("API call failed: %v", err)
    }

    // Pretty print response
    var prettyJSON bytes.Buffer
    json.Indent(&prettyJSON, body, "", "  ")
    t.Logf("User data:\n%s", prettyJSON.String())
}
```

### Common Issues

#### 1. Password Not Persisted

**Symptom**: Password shows as changed on every plan

**Cause**: Password is write-only in BCM API

**Fix**: In Read(), preserve password from state instead of API response:
```go
// DON'T do this:
model.Password = getStringValue(userData, "password")

// DO this:
// Password is write-only - preserve from state
model.Password = state.Password
```

#### 2. UUID Mismatch After Import

**Symptom**: Import works but subsequent plan shows changes

**Cause**: Import by username vs UUID lookup

**Fix**: Use username for Read lookup, not UUID:
```go
func (r *CMUserUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
    resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("username"), req.ID)...)
}
```

#### 3. Drift Detection Not Working

**Symptom**: External changes not detected by terraform plan

**Cause**: Read() not properly mapping all API fields

**Fix**: Verify field mapping in readUser() helper:
```go
// Ensure all mutable fields are read from API
model.Shell = getStringValue(userData, "loginShell")  // NOT "shell"
model.HomeDirectory = getStringValue(userData, "homeDirectory")  // NOT "home_directory"
```

#### 4. Validation Errors on Create

**Symptom**: "Validation Error: field_name" during create

**Cause**: BCM server-side validation failure

**Fix**: Check validateUser response for specific requirements:
```go
validationErrors, err := r.client.ValidateEntity(ctx, "cmuser", "validateUser", entity, true)
for _, valErr := range validationErrors {
    t.Logf("Validation: field=%s, message=%s, severity=%s",
        valErr.Field, valErr.Message, valErr.Severity)
}
```

---

## File Locations

| File | Purpose |
|------|---------|
| `internal/provider/resource_cmuser_user.go` | Resource implementation |
| `internal/provider/resource_cmuser_user_test.go` | Acceptance tests |
| `internal/provider/data_source_cmuser_users.go` | Reference: existing data source |
| `internal/provider/test_helpers.go` | Shared test utilities |
| `internal/provider/provider.go` | Resource registration |
| `examples/resources/bcm_cmuser_user/resource.tf` | Example configurations |
| `docs/resources/cmuser_user.md` | Auto-generated documentation |

---

## Useful Commands

```bash
# Build provider
make build

# Install provider locally
make install

# Format code
make fmt

# Run linter
make lint

# Generate documentation
make generate

# Run all acceptance tests
make testacc

# Run specific test with coverage
go test -v -cover ./internal/provider/ -run TestAccCMUserUser
```

---

## References

- [Terraform Plugin Framework Documentation](https://developer.hashicorp.com/terraform/plugin/framework)
- [Terraform Plugin Testing Documentation](https://developer.hashicorp.com/terraform/plugin/testing)
- [Project TDD Guide](/workspace/AGENTS.md)
- [Provider Development Guide](/workspace/CLAUDE.md)
