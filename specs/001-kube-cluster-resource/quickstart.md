# Quick Start: bcm_cmkube_cluster Resource Development

**Resource**: `bcm_cmkube_cluster`
**Feature Branch**: `001-kube-cluster-resource`
**Date**: 2025-11-22

## Overview

This guide helps developers quickly get started implementing and testing the BCM Kubernetes cluster resource using TDD (Test-Driven Development) workflow.

## Prerequisites

### BCM Environment
- BCM endpoint accessible at `https://172.21.15.254:8081`
- Valid credentials (username/password)
- Minimum 2 available nodes (1 master + 1 worker for comprehensive tests)

### Development Environment
- Go 1.24+
- terraform-plugin-framework v1.16.1
- terraform-plugin-testing v1.13.3
- make, golangci-lint, pre-commit hooks installed

## Environment Setup

### 1. Set Required Environment Variables

```bash
export TF_ACC=1                                  # Enable acceptance tests
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

### 2. Verify BCM Connectivity

```bash
# Test authentication with existing script
cd /workspace
BCM_ENDPOINT="$BCM_ENDPOINT" \
BCM_USERNAME="$BCM_USERNAME" \
BCM_PASSWORD="$BCM_PASSWORD" \
python3 sampleRest/cmnet-get-networks.py
```

Expected: Authentication successful, network list returned

## TDD Workflow (RED-GREEN-REFACTOR)

### Phase 2: RED - Write Failing Tests

**Goal**: Write comprehensive acceptance tests BEFORE implementation

```bash
# Create test file
vim internal/provider/resource_cmkube_cluster_test.go

# Run tests (they MUST fail - resource doesn't exist yet)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeCluster
```

**Expected Result**: All tests fail with "resource not found" or similar error

**Test Coverage Required**:
- Basic CRUD (Create, Read, Import, Update, Delete)
- Drift detection (external modifications detected)
- Worker node scaling (add/remove workers)
- Validation (invalid name, invalid version)
- Idempotency (no changes after apply)
- ID consistency (same ID across create/import/update)

### Phase 3: GREEN - Minimal Implementation

**Goal**: Write minimal code to make all tests PASS

```bash
# Create resource file
vim internal/provider/resource_cmkube_cluster.go

# Implement:
# 1. Resource struct and factory
# 2. Schema definition
# 3. Create method
# 4. Read method (with readCluster helper)
# 5. Update method
# 6. Delete method
# 7. ImportState method
# 8. buildClusterEntity helper

# Run tests again
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeCluster
```

**Expected Result**: All tests PASS

### Phase 4: REFACTOR - Improve Code Quality

**Goal**: Improve code while keeping tests green

```bash
# Run formatters and linters
make fmt
make lint

# Run pre-commit hooks
pre-commit run --all-files

# Re-run tests to ensure still passing
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeCluster
```

## Running Tests

### Run All Cluster Tests

```bash
TF_ACC=1 \
  BCM_ENDPOINT="https://172.21.15.254:8081" \
  BCM_USERNAME="root" \
  BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeCluster
```

### Run Specific Test

```bash
# Basic CRUD test only
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMKubeClusterResource_Basic

# Drift detection test only
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMKubeClusterResource_DriftDetection

# Worker nodes test only
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMKubeClusterResource_WorkerNodes

# Validation tests
TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMKubeClusterResource_Validation
```

### Enable Verbose Logging

```bash
# Maximum verbosity for debugging
TF_ACC=1 \
  TF_LOG=TRACE \
  TF_LOG_PATH=./terraform-test.log \
  go test -v -timeout 120m ./internal/provider/ -run TestAccCMKubeCluster
```

## Example Usage

### Basic Cluster (Minimal Config)

```hcl
provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true
}

resource "bcm_cmkube_cluster" "basic" {
  name         = "my-cluster"
  master_nodes = ["<master-node-uuid>"]
}

output "cluster_uuid" {
  value = bcm_cmkube_cluster.basic.uuid
}
```

### Cluster with Workers and Version

```hcl
resource "bcm_cmkube_cluster" "full" {
  name         = "production-cluster"
  master_nodes = ["<master-uuid-1>"]

  worker_nodes = [
    "<worker-uuid-1>",
    "<worker-uuid-2>",
  ]

  version            = "1.28.0"
  management_network = "<network-uuid>"
}
```

### Import Existing Cluster

```bash
terraform import bcm_cmkube_cluster.existing <cluster-uuid>
```

## File Structure

```
terraform-provider-bcm/
├── internal/provider/
│   ├── resource_cmkube_cluster.go           # ← Implementation
│   ├── resource_cmkube_cluster_test.go      # ← Acceptance tests
│   ├── bcm_client.go                         # Existing BCM API client
│   ├── test_helpers.go                       # Existing test utilities
│   └── provider.go                           # Register resource here
│
├── specs/001-kube-cluster-resource/
│   ├── spec.md                               # User requirements
│   ├── plan.md                               # Implementation plan
│   ├── tasks.md                              # Task breakdown
│   ├── research.md                           # Phase 0 API findings
│   ├── data-model.md                         # Schema and mappings
│   ├── quickstart.md                         # This file
│   └── contracts/                            # API contract examples
│       ├── create-cluster.json
│       ├── read-cluster.json
│       ├── update-cluster.json
│       └── delete-cluster.json
│
└── examples/resources/bcm_cmkube_cluster/
    ├── resource.tf                           # Basic example
    ├── import.sh                             # Import example
    └── advanced.tf                           # Advanced configuration
```

## Development Commands

```bash
# Format code
make fmt

# Run linter
make lint

# Run unit tests (no BCM required)
make test

# Run acceptance tests (requires BCM)
make testacc

# Generate documentation
make generate

# Build provider
make build

# Install provider locally
make install
```

## Common Patterns

### Helper Functions (Existing)

```go
// From internal/provider/test_helpers.go
createTestBCMClient(t)                    // Create authenticated BCM client
generateUniqueTestName("prefix")          // Generate timestamped unique name
verifyResourceDeleted(ctx, client, ...)   // Exponential backoff deletion check

// From internal/provider/data_source_cmpart_softwareimages.go
getStringValue(data, "fieldName")         // Null-safe string extraction
getBoolValue(data, "fieldName")           // Null-safe bool extraction
getInt64Value(data, "fieldName")          // Null-safe int64 extraction
```

### Modern Test Patterns

```go
// State checks (type-safe)
statecheck.ExpectKnownValue(
    "bcm_cmkube_cluster.test",
    tfjsonpath.New("name"),
    knownvalue.StringExact("test-cluster"),
)

// Plan checks (idempotency)
plancheck.ExpectEmptyPlan()  // No changes after apply

// ID consistency tracking
compareID := statecheck.CompareValue(compare.ValuesSame())
compareID.AddStateValue("bcm_cmkube_cluster.test", tfjsonpath.New("id"))
```

## Troubleshooting

### Tests Fail: "BCM_ENDPOINT not set"
**Solution**: Export environment variables before running tests

### Tests Fail: "Not enough nodes"
**Solution**: Check available nodes with `python3 sampleRest/cmdevice-get-nodes.py`

### Tests Timeout
**Solution**: Increase timeout or check BCM cluster performance
```bash
TF_ACC=1 go test -v -timeout 240m ./internal/provider/ -run TestAccCMKubeCluster
```

### Drift Test Fails
**Solution**: Ensure 2-second sleep after external API modifications for eventual consistency

### Import Test Fails
**Solution**: Verify ImportState uses `resource.ImportStatePassthroughID` for uuid attribute

## Next Steps

1. **Phase 2**: Write failing acceptance tests (TDD RED)
2. **Phase 3**: Implement minimal CRUD to pass tests (TDD GREEN)
3. **Phase 4**: Register resource, create examples, run full test suite
4. **Phase 5**: Generate documentation, run quality checks

## References

- **Existing Resources**: `resource_cmpart_softwareimage.go`, `resource_cmdevice_category.go`
- **BCM Client**: `internal/provider/bcm_client.go`
- **Test Patterns**: `resource_cmpart_softwareimage_test.go`
- **API Contracts**: `specs/001-kube-cluster-resource/contracts/`
- **Data Model**: `specs/001-kube-cluster-resource/data-model.md`
- **CLAUDE.md**: Project-level guidance
- **AGENTS.md**: TDD patterns and best practices

## Success Criteria

- ✅ All acceptance tests pass on first run
- ✅ No hardcoded UUIDs or values in tests
- ✅ Import functionality works correctly
- ✅ Drift detection accurately identifies changes
- ✅ Code passes fmt and lint checks
- ✅ Documentation auto-generates successfully
- ✅ Examples work in test environment

**Ready to start? Begin with Phase 2: Write failing tests!**
