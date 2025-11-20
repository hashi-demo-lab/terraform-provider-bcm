# Quickstart: bcm_cmpart_softwareimage Resource

## Overview

This guide helps you quickly implement and test the `bcm_cmpart_softwareimage` Terraform resource for managing BCM software images.

## Prerequisites

- Go 1.24+ installed
- Access to BCM cluster at `172.21.15.254:8081`
- BCM credentials (username/password)
- Terraform 1.0+ installed
- Git repository cloned: `terraform-provider-bcm`

## Environment Setup

```bash
# Set BCM credentials for acceptance tests
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Set Go environment
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
```

## Phase 0: Research (Start Here)

### Task 1: Find Update/Delete API Methods

**Goal:** Discover the exact API method names for update and delete operations.

**Steps:**

1. Search BCM documentation:
   ```bash
   cd /workspace
   grep -r "updateSoftwareImage\|modifySoftwareImage\|removeSoftwareImage" sampleRest/
   ```

2. Test API methods manually (use Python or curl):
   ```python
   # Create test_api.py
   import requests
   import json

   session = requests.Session()
   session.verify = False

   # Login
   login_resp = session.post(
       "https://172.21.15.254:8081/json",
       json={"service": "login", "username": "root", "password": "Hashicorp123!"}
   )
   print(f"Login: {login_resp.json()}")

   # Test update methods
   test_methods = [
       "updateSoftwareImage",
       "modifySoftwareImage",
       "setSoftwareImage"
   ]

   for method in test_methods:
       resp = session.post(
           "https://172.21.15.254:8081/json",
           json={"service": "CMPart", "call": method, "args": []}
       )
       print(f"{method}: {resp.status_code} - {resp.text[:200]}")
   ```

3. Document findings in `research.md`:
   ```markdown
   # Research Findings

   ## Update Method
   - Method: `updateSoftwareImage` (or discovered name)
   - Parameters: [entity] or [uuid, entity]
   - Response: success boolean or updated entity

   ## Delete Method
   - Method: `removeSoftwareImage` (or discovered name)
   - Parameters: [uuid] or [name]
   - Response: success boolean
   ```

### Task 2: Test CRUD Lifecycle

**Goal:** Manually perform full CRUD cycle to understand API behavior.

**Steps:**

1. Create software image:
   ```python
   create_resp = session.post(
       "https://172.21.15.254:8081/json",
       json={
           "service": "CMPart",
           "call": "addSoftwareImage",
           "args": [
               {
                   "baseType": "SoftwareImage",
                   "childType": "",
                   "name": "test-research-image",
                   "path": "/cm/images/test-research-image",
                   "kernelVersion": "5.15.0-58-generic",
                   "kernelParameters": "quiet",
                   "modules": []
               },
               False  # force parameter
           ]
       }
   )
   print(f"Create: {create_resp.json()}")
   created_uuid = create_resp.json()  # Extract UUID
   ```

2. Read software image:
   ```python
   read_resp = session.post(
       "https://172.21.15.254:8081/json",
       json={"service": "CMPart", "call": "getSoftwareImages"}
   )
   images = read_resp.json()
   test_image = [img for img in images if img['uuid'] == created_uuid][0]
   print(f"Read: {test_image}")
   ```

3. Update software image:
   ```python
   # Try discovered update method
   test_image['kernelParameters'] = 'quiet splash'
   update_resp = session.post(
       "https://172.21.15.254:8081/json",
       json={
           "service": "CMPart",
           "call": "updateSoftwareImage",  # Use discovered method
           "args": [test_image]
       }
   )
   print(f"Update: {update_resp.json()}")
   ```

4. Delete software image:
   ```python
   # Try discovered delete method
   delete_resp = session.post(
       "https://172.21.15.254:8081/json",
       json={
           "service": "CMPart",
           "call": "removeSoftwareImage",  # Use discovered method
           "args": [created_uuid]
       }
   )
   print(f"Delete: {delete_resp.json()}")
   ```

5. Document results in `research.md`

## Phase 1: RED (Write Failing Tests)

### Task 1: Create Acceptance Test File

```bash
cd /workspace
touch internal/provider/resource_cmpart_softwareimage_test.go
```

**Copy test code from plan.md Phase 2 RED section** into the file.

### Task 2: Run Tests (Expect Failures)

```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic
```

**Expected output:**
```
=== RUN   TestAccCMPartSoftwareImageResource_Basic
--- FAIL: TestAccCMPartSoftwareImageResource_Basic (0.01s)
    Error: resource type not found: bcm_cmpart_softwareimage
```

This is correct - RED phase should fail.

## Phase 2: GREEN (Minimal Implementation)

### Task 1: Create Resource File

```bash
cd /workspace
touch internal/provider/resource_cmpart_softwareimage.go
```

**Copy minimal implementation from plan.md Phase 2 GREEN section** into the file.

### Task 2: Register Resource in Provider

Edit `/workspace/internal/provider/provider.go`:

```go
func (p *BCMProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        NewCMPartSoftwareImageResource, // ADD THIS LINE
    }
}
```

### Task 3: Run Tests (Expect Success)

```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic
```

**Expected output:**
```
=== RUN   TestAccCMPartSoftwareImageResource_Basic
--- PASS: TestAccCMPartSoftwareImageResource_Basic (15.23s)
PASS
```

Tests pass with hardcoded implementation.

## Phase 3: REFACTOR (Real API Integration)

### Task 1: Update Create Method

**Replace hardcoded Create method** with API call implementation from plan.md.

Key changes:
- Build API entity from plan
- Call `addSoftwareImage` with entity and force parameter
- Parse UUID from response
- Call Read to populate computed fields

### Task 2: Update Read Method

**Replace no-op Read method** with API implementation:
- Call `getSoftwareImages()`
- Filter results by UUID (client-side)
- Map API response to model using helper functions

### Task 3: Update Update Method

**Replace no-op Update method** with API implementation:
- Build updated entity from plan
- Call discovered update API method
- Read back to verify changes

### Task 4: Update Delete Method

**Replace no-op Delete method** with API implementation:
- Call discovered delete API method with UUID
- Verify success response

### Task 5: Run Full Test Suite

```bash
cd /workspace
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartSoftwareImageResource
```

**Expected output:**
```
=== RUN   TestAccCMPartSoftwareImageResource_Basic
--- PASS: TestAccCMPartSoftwareImageResource_Basic (25.43s)
=== RUN   TestAccCMPartSoftwareImageResource_FullConfig
--- PASS: TestAccCMPartSoftwareImageResource_FullConfig (30.12s)
=== RUN   TestAccCMPartSoftwareImageResource_UpdateModules
--- PASS: TestAccCMPartSoftwareImageResource_UpdateModules (35.67s)
PASS
```

## Phase 4: Documentation

### Task 1: Create Example Configuration

```bash
mkdir -p /workspace/examples/resources/bcm_cmpart_softwareimage
```

Create `/workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf`:

```hcl
resource "bcm_cmpart_softwareimage" "example" {
  name = "ubuntu-22.04-dpu"
  path = "/cm/images/ubuntu-22.04-dpu"

  kernel_version    = "5.15.0-58-generic"
  kernel_parameters = "quiet splash"

  enable_sol = true
  sol_speed  = "115200"

  modules = [
    {
      name       = "nvidia-drm"
      parameters = "modeset=1"
    },
    {
      name       = "e1000e"
      parameters = ""
    }
  ]

  notes = "Ubuntu 22.04 LTS for DPU nodes"
}
```

### Task 2: Generate Documentation

```bash
cd /workspace
make generate
```

This creates `/workspace/docs/resources/cmpart_softwareimage.md`.

### Task 3: Verify Documentation

```bash
cat docs/resources/cmpart_softwareimage.md
```

Ensure:
- Schema attributes documented
- Example included
- Import command shown

## Phase 5: Quality Checks

### Task 1: Run Linter

```bash
cd /workspace
make lint
```

Fix any reported issues.

### Task 2: Format Code

```bash
cd /workspace
make fmt
```

### Task 3: Run Pre-commit Hooks

```bash
cd /workspace
pre-commit run --all-files
```

### Task 4: Run Full Test Suite

```bash
cd /workspace
make test        # Unit tests
make testacc     # Acceptance tests (requires BCM access)
```

## Manual Smoke Test

### Task 1: Build Provider

```bash
cd /workspace
go install
```

### Task 2: Create Test Configuration

Create `/tmp/test-bcm-image.tf`:

```hcl
terraform {
  required_providers {
    bcm = {
      source = "hashicorp/bcm"
    }
  }
}

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true
}

resource "bcm_cmpart_softwareimage" "test" {
  name = "terraform-test-image"
  path = "/cm/images/terraform-test-image"

  kernel_version    = "5.15.0-58-generic"
  kernel_parameters = "quiet"

  notes = "Created by Terraform smoke test"
}

output "image_uuid" {
  value = bcm_cmpart_softwareimage.test.uuid
}
```

### Task 3: Apply Configuration

```bash
cd /tmp
terraform init
terraform plan
terraform apply -auto-approve
```

**Expected output:**
```
Apply complete! Resources: 1 added, 0 changed, 0 destroyed.

Outputs:
image_uuid = "uuid-of-created-image"
```

### Task 4: Verify in BCM

```bash
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -b "cm-login-token=<token>" \
  -d '{"service":"CMPart","call":"getSoftwareImages"}' | jq '.[] | select(.name=="terraform-test-image")'
```

### Task 5: Cleanup

```bash
cd /tmp
terraform destroy -auto-approve
```

## Troubleshooting

### Issue: Tests Fail with Authentication Error

**Solution:**
```bash
# Verify credentials
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"Hashicorp123!"}'

# Should return: true
```

### Issue: Tests Timeout

**Solution:**
```bash
# Increase timeout
TF_ACC=1 go test -v -timeout 240m ./internal/provider/ -run TestAccCMPartSoftwareImageResource
```

### Issue: Duplicate Name Error

**Solution:**
```bash
# Use unique names with timestamp
resource "bcm_cmpart_softwareimage" "test" {
  name = "test-image-${timestamp()}"
  # ...
}
```

### Issue: Module Not Found During Import

**Solution:**
```bash
# Ensure go.mod is up to date
go mod tidy
go mod download
```

## Key Files Reference

| File | Purpose |
|------|---------|
| `/workspace/internal/provider/resource_cmpart_softwareimage.go` | Resource implementation |
| `/workspace/internal/provider/resource_cmpart_softwareimage_test.go` | Acceptance tests |
| `/workspace/internal/provider/provider.go` | Resource registration |
| `/workspace/examples/resources/bcm_cmpart_softwareimage/resource.tf` | Example configuration |
| `/workspace/docs/resources/cmpart_softwareimage.md` | Generated documentation |
| `/workspace/specs/003-bcm-cmpart-softwareimage-resource/plan.md` | Implementation plan |
| `/workspace/specs/003-bcm-cmpart-softwareimage-resource/spec.md` | Feature specification |
| `/workspace/specs/003-bcm-cmpart-softwareimage-resource/research.md` | API research findings |

## Helper Commands

```bash
# Run specific test
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartSoftwareImageResource_Basic

# Run tests with verbose logging
TF_LOG=DEBUG TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartSoftwareImageResource

# Build provider locally
go build -o terraform-provider-bcm

# Install provider to local Terraform
go install

# Generate docs only
cd tools && go generate ./...

# Check for syntax errors
go vet ./...

# Format all Go files
gofmt -s -w -e .

# Run linter with auto-fix
golangci-lint run --fix
```

## Next Steps After Implementation

1. Test with different BCM versions
2. Add more comprehensive error handling
3. Implement retry logic for transient failures
4. Add unit tests for helper functions
5. Create integration tests with other resources
6. Document common usage patterns
7. Gather user feedback

## Resources

- Plan Document: `specs/003-bcm-cmpart-softwareimage-resource/plan.md`
- Specification: `specs/003-bcm-cmpart-softwareimage-resource/spec.md`
- BCM API Docs: `sampleRest/BCM_API_Complete_Documentation.md`
- Terraform Plugin Framework: https://developer.hashicorp.com/terraform/plugin/framework
- Provider Development Guide: `AGENTS.md`
- TDD Constitution: `.specify/memory/constitution.md`
