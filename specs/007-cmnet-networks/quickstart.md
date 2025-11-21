# Quick Start: BCM CMNet Networks Data Source Development

**Feature**: `bcm_cmnet_networks` Terraform data source
**Branch**: `007-cmnet-networks`
**Status**: Implementation in progress

## Prerequisites

### Required Tools

- **Go**: Version 1.24.0 or later
- **Terraform**: Version 1.5.0 or later (for testing)
- **Python 3**: For API exploration scripts
- **Make**: For build automation
- **Git**: For version control

### Required Access

- **BCM Cluster**: https://172.21.15.254:8081
- **Credentials**: root / Hashicorp123!
- **Network Access**: Ability to reach BCM API endpoint

### Environment Setup

```bash
# Set BCM credentials for acceptance tests
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Set Go environment (if needed)
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go
```

## Project Structure

```
terraform-provider-bcm/
├── specs/007-cmnet-networks/          # Feature documentation
│   ├── spec.md                         # Feature specification
│   ├── plan.md                         # Implementation plan
│   ├── tasks.md                        # Task breakdown
│   ├── research.md                     # API exploration findings
│   ├── data-model.md                   # Data model documentation
│   ├── quickstart.md                   # This file
│   └── contracts/                      # API contracts
│       ├── cmnet-get-networks-request.json
│       └── cmnet-get-networks-response.json
├── internal/provider/
│   ├── data_source_cmnet_networks.go      # Implementation
│   ├── data_source_cmnet_networks_test.go # Acceptance tests
│   ├── bcm_client.go                      # Reused API client
│   └── data_source_cmpart_softwareimages.go # Helper functions
├── examples/data-sources/bcm_cmnet_networks/
│   ├── data-source.tf                  # Basic example
│   └── filtered.tf                     # Filtering examples
└── docs/data-sources/
    └── cmnet_networks.md               # Auto-generated docs
```

## API Exploration

### Run API Exploration Script

```bash
# Explore the BCM cmnet.getNetworks API
cd /workspace
BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" \
BCM_PASSWORD="Hashicorp123!" \
python3 sampleRest/cmnet-get-networks.py
```

**Expected Output**: JSON array with 3 network objects (managementnet, globalnet, internalnet)

### API Request Format

```json
{
  "service": "cmnet",
  "call": "getNetworks"
}
```

### API Response Format

See `/workspace/specs/007-cmnet-networks/contracts/cmnet-get-networks-response.json` for complete response structure.

**Key Fields**:
- `uuid`: Unique network identifier
- `name`: Network name (e.g., "managementnet")
- `baseAddress`: Network base IP (e.g., "172.21.12.0")
- `netmaskBits`: CIDR netmask (e.g., 22)
- `gateway`: Gateway IP
- `domainName`: DNS domain
- `dynamicRangeStart`/`dynamicRangeEnd`: DHCP range
- `management`: Is management network flag
- `type`: Network type (INTERNAL, GLOBAL)

## TDD Workflow

This feature follows strict **RED-GREEN-REFACTOR** TDD methodology:

### Phase 2.1: RED - Write Failing Tests

```bash
# Create acceptance test file
vim internal/provider/data_source_cmnet_networks_test.go

# Write tests (they will fail - data source doesn't exist yet)
# - TestAccCMNetNetworksDataSource_Basic
# - TestAccCMNetNetworksDataSource_NameFilter
# - TestAccCMNetNetworksDataSource_DHCPFilter
# - TestAccCMNetNetworksDataSource_NoMatch

# Run tests - verify they FAIL
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks

# Expected: All tests fail with "unknown data source bcm_cmnet_networks"
```

### Phase 2.2: GREEN - Minimal Implementation

```bash
# Create implementation file
vim internal/provider/data_source_cmnet_networks.go

# Implement minimal working code:
# 1. Define data source struct and models
# 2. Implement Metadata(), Schema(), Configure(), Read() methods
# 3. Create helper functions mapAPIToNetwork() and matchesFilter()
# 4. Register data source in provider.go

# Run tests - verify they PASS
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks

# Expected: All tests pass (100% pass rate)
```

### Phase 2.3: REFACTOR - Improve Code Quality

```bash
# Enhance implementation:
# 1. Add comprehensive error handling
# 2. Add debug logging with tflog
# 3. Add godoc comments
# 4. Enhance schema descriptions
# 5. Optimize filtering logic

# Run tests again - verify they STILL PASS
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks

# Run code quality checks
make lint
make fmt
```

## Running Tests

### Acceptance Tests (Full Suite)

```bash
# Run all cmnet_networks data source tests
TF_ACC=1 \
BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" \
BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks
```

### Single Test Execution

```bash
# Run specific test
TF_ACC=1 \
BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" \
BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworksDataSource_Basic
```

### All Provider Tests

```bash
# Verify no regressions in other data sources/resources
TF_ACC=1 \
BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" \
BCM_PASSWORD="Hashicorp123!" \
go test -v -timeout 120m ./internal/provider/
```

## Code Quality

### Format Code

```bash
make fmt
# Or manually:
gofmt -s -w -e .
```

### Lint Code

```bash
make lint
# Or manually:
golangci-lint run
```

### Run Unit Tests

```bash
make test
# Or manually:
go test -v -cover ./internal/provider/
```

## Documentation Generation

### Auto-Generate Provider Docs

```bash
# Generate documentation using tfplugindocs
make generate

# Or manually:
cd /workspace
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate

# Verify docs generated
ls -la docs/data-sources/cmnet_networks.md
```

### Verify Documentation

```bash
# Check for uncommitted doc changes
git diff docs/

# Review generated docs
cat docs/data-sources/cmnet_networks.md
```

## Example Terraform Configuration

### Basic Usage

```hcl
# Provider configuration
provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true
}

# Retrieve all networks
data "bcm_cmnet_networks" "all" {}

# Output network information
output "all_networks" {
  value = data.bcm_cmnet_networks.all.networks
}
```

### Filtered Usage

```hcl
# Find management network
data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "management"
  }
}

# Find DHCP-enabled networks
data "bcm_cmnet_networks" "dhcp_networks" {
  filter {
    dhcp_enabled = true
  }
}
```

## Helper Functions

The implementation reuses existing null-safe helper functions from `data_source_cmpart_softwareimages.go`:

```go
// Extract string value with null handling
name := getStringValue(networkData, "name")

// Extract boolean value with null handling
management := getBoolValue(networkData, "management")

// Extract int64 value with null handling
mtu := getInt64Value(networkData, "mtu")
```

**Location**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go:401-431`

## Common Issues & Solutions

### Issue: Tests fail with "authentication failed"

**Solution**: Verify BCM credentials are correct and endpoint is reachable:

```bash
# Test API access
curl -k -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -d '{"service":"login","username":"root","password":"Hashicorp123!"}'
```

### Issue: "unknown data source bcm_cmnet_networks"

**Solution**: Data source not registered in provider.go. Add to DataSources() method:

```go
func (p *bcmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        // ... existing data sources ...
        NewCMNetNetworksDataSource,  // Add this line
    }
}
```

### Issue: Tests hang or timeout

**Solution**: Check BCM cluster is online and responsive:

```bash
# Verify BCM API is responding
python3 sampleRest/cmnet-get-networks.py
```

### Issue: "no networks found"

**Solution**: BCM cluster must have at least one network configured. Verify:

```bash
# Run API exploration script
BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" \
BCM_PASSWORD="Hashicorp123!" \
python3 sampleRest/cmnet-get-networks.py

# Expected: JSON array with at least 1 network object
```

### Issue: Documentation not generated

**Solution**: Ensure examples exist before running make generate:

```bash
# Create examples directory
mkdir -p examples/data-sources/bcm_cmnet_networks

# Create example file
vim examples/data-sources/bcm_cmnet_networks/data-source.tf

# Run documentation generation
make generate
```

## Development Workflow

### Recommended Development Flow

1. **Research**: Run API exploration script, understand response structure
2. **Design**: Review data-model.md and contracts
3. **RED**: Write failing acceptance tests
4. **GREEN**: Implement minimal working code
5. **REFACTOR**: Enhance code quality
6. **Examples**: Create example configurations
7. **Documentation**: Generate provider docs
8. **Validation**: Run full test suite

### Git Workflow

```bash
# Create feature branch (if not already on it)
git checkout -b 007-cmnet-networks

# Stage changes incrementally
git add internal/provider/data_source_cmnet_networks_test.go
git commit -m "test: Add failing acceptance tests for bcm_cmnet_networks"

git add internal/provider/data_source_cmnet_networks.go
git add internal/provider/provider.go
git commit -m "feat: Implement bcm_cmnet_networks data source"

git add examples/data-sources/bcm_cmnet_networks/
git add docs/data-sources/cmnet_networks.md
git commit -m "docs: Add examples and generated docs for bcm_cmnet_networks"

# Push to remote
git push origin 007-cmnet-networks
```

## Testing Strategy

### Test Data Requirements

The acceptance tests assume the following networks exist in the BCM cluster:

1. **managementnet**: INTERNAL network with DHCP enabled
2. **internalnet**: INTERNAL network with DHCP enabled
3. **globalnet**: GLOBAL network with no DHCP

**Verification**:
```bash
python3 sampleRest/cmnet-get-networks.py | grep '"name"'
```

### Test Scenarios

1. **Basic Read**: Retrieve all networks without filters
2. **Name Filter**: Filter by name pattern "management"
3. **DHCP Filter**: Filter by dhcp_enabled=true
4. **Empty Results**: Filter by nonsensical pattern (no match)

### Read-Only Testing Approach

All tests are **read-only**:
- No network creation
- No network modification
- No network deletion
- Tests use existing networks in BCM cluster
- Tests can run in parallel

## Performance Expectations

- **API Call Time**: <1 second for getNetworks
- **Filtering Time**: <1ms for 100 networks
- **Total Read Time**: <5 seconds end-to-end
- **Test Execution**: <2 minutes for all 4 acceptance tests

## Key Files Reference

| File | Purpose |
|---|---|
| `spec.md` | Feature specification |
| `plan.md` | Implementation plan |
| `tasks.md` | Task breakdown |
| `research.md` | API exploration findings |
| `data-model.md` | Data model documentation |
| `contracts/cmnet-get-networks-request.json` | API request contract |
| `contracts/cmnet-get-networks-response.json` | API response contract |
| `internal/provider/data_source_cmnet_networks.go` | Implementation |
| `internal/provider/data_source_cmnet_networks_test.go` | Tests |
| `examples/data-sources/bcm_cmnet_networks/data-source.tf` | Example |
| `docs/data-sources/cmnet_networks.md` | Generated docs |

## Next Steps

After completing this feature:

1. Create pull request with all changes
2. Request code review
3. Address review feedback
4. Merge to main branch
5. Tag release version
6. Update changelog

## Additional Resources

- **BCM API Documentation**: `sampleRest/CMDevice_Complete_Documentation.md`
- **Provider Design Patterns**: `AGENTS.md`
- **TDD Constitution**: `.specify/memory/constitution.md`
- **Terraform Plugin Framework**: https://developer.hashicorp.com/terraform/plugin/framework
- **Acceptance Testing**: https://developer.hashicorp.com/terraform/plugin/testing/acceptance-tests
