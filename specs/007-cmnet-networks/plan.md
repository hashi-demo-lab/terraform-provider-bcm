# Implementation Plan: BCM CMNet Networks Data Source

**Branch**: `007-cmnet-networks` | **Date**: 2025-11-21 | **Spec**: [/workspace/specs/007-cmnet-networks/spec.md](spec.md)
**Input**: Feature specification from `/workspace/specs/007-cmnet-networks/spec.md`

## Summary

Implement a read-only Terraform data source `bcm_cmnet_networks` that retrieves network configurations from the BCM CMNet service. The data source will call the BCM JSON-RPC API endpoint `{"service": "cmnet", "call": "getNetworks"}`, parse the network objects, support client-side filtering by name pattern and DHCP status, and follow TDD principles with acceptance tests written before implementation.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3
**Storage**: N/A (read-only data source)
**Testing**: Go testing framework with TF_ACC=1 acceptance tests
**Target Platform**: Linux (BCM cluster at 172.21.15.254:8081)
**Project Type**: Terraform Provider (data source component)
**Performance Goals**: Query all networks in <5 seconds for clusters with up to 100 networks
**Constraints**: Read-only operations, fail-fast error handling, no retry logic
**Scale/Scope**: Support BCM clusters with up to 100 networks, client-side filtering

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### TDD Principles
- ✅ **Tests First**: All acceptance tests will be written in Phase 1 (RED) before implementation
- ✅ **Minimal Implementation**: Phase 2 (GREEN) will implement only what's needed to pass tests
- ✅ **Refactor After Green**: Phase 3 (REFACTOR) will improve code quality while keeping tests green
- ✅ **No Complexity Violations**: This is a standard data source following established patterns

### Architecture Simplicity
- ✅ **Reuse Existing Patterns**: Follows patterns from `data_source_cmdevice_nodes.go` and `data_source_cmpart_softwareimages.go`
- ✅ **Reuse Helper Functions**: Uses existing `getStringValue()`, `getBoolValue()`, `getInt64Value()` helpers
- ✅ **Reuse BCMClient**: Uses existing `BCMClient.CallJSONRPC()` method
- ✅ **Client-Side Filtering**: Simple filtering logic, no new dependencies

### Testing Strategy
- ✅ **Acceptance Tests Only**: Data sources are inherently read-only, no complex unit test requirements
- ✅ **Live BCM Environment**: Tests use existing networks (no create/destroy needed)
- ✅ **4 Test Scenarios**: Basic read, name filter, DHCP filter, empty results
- ✅ **100% Coverage**: All read paths and filter combinations tested

**GATE RESULT**: ✅ PASS - No violations. Standard data source implementation.

## Project Structure

### Documentation (this feature)

```text
specs/007-cmnet-networks/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (API exploration)
├── data-model.md        # Phase 1 output (network entity schema)
├── quickstart.md        # Phase 1 output (developer quick start)
├── contracts/           # Phase 1 output (API request/response contracts)
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
internal/provider/
├── data_source_cmnet_networks.go       # Main implementation (Phase 2)
├── data_source_cmnet_networks_test.go  # Acceptance tests (Phase 1)
├── bcm_client.go                       # Existing client (reused)
├── data_source_cmpart_softwareimages.go # Helper functions (reused)
└── provider.go                         # Provider registration (modified)

examples/
└── data-sources/
    └── bcm_cmnet_networks/
        ├── data-source.tf              # Basic usage example (Phase 3)
        └── filtered.tf                 # Filtering examples (Phase 3)

docs/
└── data-sources/
    └── cmnet_networks.md               # Auto-generated docs (Phase 4)
```

**Structure Decision**: Standard Terraform provider structure with data source in `internal/provider/`. This follows the established pattern in the codebase where all data sources are co-located in the same directory for easy discovery and maintenance.

## Complexity Tracking

> **No violations to justify** - This implementation follows TDD constitution perfectly.

## Phase 0: Outline & Research

### Research Objectives

1. **Explore BCM CMNet API Response Structure**
   - Make live API call to `{"service": "cmnet", "call": "getNetworks"}`
   - Document actual JSON response structure
   - Identify all available fields and their data types
   - Confirm field naming conventions (camelCase vs snake_case)

2. **Analyze Existing Data Source Patterns**
   - Review `data_source_cmdevice_nodes.go` for nested object patterns
   - Review `data_source_cmpart_softwareimages.go` for helper function usage
   - Identify reusable patterns for filtering logic
   - Document schema definition best practices

3. **Verify Helper Function Locations**
   - Confirm `getStringValue()`, `getBoolValue()`, `getInt64Value()` are in `data_source_cmpart_softwareimages.go`
   - Verify these functions are accessible to other files in the same package
   - Document any considerations for code organization

### Research Deliverable: research.md

Document findings in `/workspace/specs/007-cmnet-networks/research.md` with:

- **API Response Sample**: Complete JSON response from `cmnet.getNetworks`
- **Field Mapping Table**: API field name → Terraform attribute name → data type
- **Pattern Analysis**: Which existing data source patterns apply
- **Helper Function Verification**: Confirm availability and usage patterns
- **Unknown Resolutions**: All "NEEDS CLARIFICATION" items from Technical Context resolved

## Phase 1: Design & Contracts

### Prerequisites
- `research.md` complete with actual API response structure

### Deliverables

#### 1. Data Model (`data-model.md`)

Extract entities and relationships:

**Network Entity**:
- **Identity**: `uuid` (unique ID), `name` (human-readable)
- **Network Config**: `network` (base IP), `netmask`, `gateway`
- **DNS**: `domain` (domain name)
- **DHCP**: `dhcp` (boolean flag)
- **Metadata**: `baseType`, `childType`, `creationTime`, `revision`, `revisionID`
- **State**: `modified`, `to_be_removed`

**Relationships**:
- Network → Nodes (via node interfaces, out of scope for this data source)

**Validation Rules**:
- None (read-only data source, no input validation needed)

#### 2. API Contracts (`contracts/`)

##### Request Contract (`contracts/cmnet-get-networks-request.json`)
```json
{
  "service": "cmnet",
  "call": "getNetworks"
}
```

##### Response Contract (`contracts/cmnet-get-networks-response.json`)
```json
[
  {
    "baseType": "Network",
    "childType": "Network",
    "uuid": "550e8400-e29b-41d4-a716-446655440000",
    "name": "management-net",
    "network": "192.168.1.0",
    "netmask": "255.255.255.0",
    "gateway": "192.168.1.1",
    "domain": "cluster.local",
    "dhcp": true,
    "modified": false,
    "to_be_removed": false,
    "creationTime": 1637856000,
    "revision": "1.0",
    "revisionID": 1
  }
]
```

#### 3. Quick Start Guide (`quickstart.md`)

Developer onboarding guide:
- Prerequisites (BCM cluster access, Go 1.24+)
- How to run API exploration scripts
- How to run acceptance tests
- TDD workflow for this feature
- Example Terraform configurations

#### 4. Agent Context Update

Run agent context update script:
```bash
/workspace/.specify/scripts/bash/update-agent-context.sh copilot
```

Add CMNet networks data source patterns to agent context:
- Network entity schema
- Filtering patterns (name_pattern, dhcp_enabled)
- Client-side filtering approach

## Phase 2: TDD Implementation (RED-GREEN-REFACTOR)

### Phase 2.1: RED - Write Failing Tests

**Objective**: Create 4 acceptance tests that define expected data source behavior, verify they fail.

**Test 1: Basic Read** (`TestAccCMNetNetworksDataSource_Basic`)
- Configuration: No filters, retrieve all networks
- Assertions:
  - `data.bcm_cmnet_networks.test.id` is set
  - `data.bcm_cmnet_networks.test.networks` is non-empty list
  - At least one network has `name` attribute set
  - At least one network has `uuid` attribute set
  - At least one network has `network` attribute set

**Test 2: Name Pattern Filter** (`TestAccCMNetNetworksDataSource_NameFilter`)
- Configuration: Filter by `name_pattern = "management"`
- Assertions:
  - All returned networks contain "management" in name (case-insensitive)
  - Networks not matching pattern are excluded

**Test 3: DHCP Filter** (`TestAccCMNetNetworksDataSource_DHCPFilter`)
- Configuration: Filter by `dhcp_enabled = true`
- Assertions:
  - All returned networks have `dhcp = true`
  - Networks with `dhcp = false` are excluded

**Test 4: Empty Results** (`TestAccCMNetNetworksDataSource_NoMatch`)
- Configuration: Filter by `name_pattern = "nonexistent-network-pattern-xyz"`
- Assertions:
  - No error returned
  - `networks` attribute is empty list

**Expected Result**: All tests FAIL with "data source not registered" or "unknown data source" errors.

**Files Created**:
- `/workspace/internal/provider/data_source_cmnet_networks_test.go`

**Verification**:
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks
```

### Phase 2.2: GREEN - Minimal Implementation

**Objective**: Write minimal code to make all tests pass.

**Step 1: Create Data Source File**

File: `/workspace/internal/provider/data_source_cmnet_networks.go`

Components:
1. **Struct Definitions**:
   - `CMNetNetworksDataSource` (implements `datasource.DataSource`, `datasource.DataSourceWithConfigure`)
   - `CMNetNetworksDataSourceModel` (root model: ID, Filter, Networks)
   - `NetworkFilterModel` (filter: name_pattern, dhcp_enabled)
   - `NetworkModel` (network attributes)

2. **Interface Methods**:
   - `Metadata()`: Returns `"bcm_cmnet_networks"`
   - `Schema()`: Defines all attributes and filter block
   - `Configure()`: Receives BCMClient from provider
   - `Read()`: Core logic (API call → parse → filter → state)

3. **Helper Functions**:
   - `mapAPIToNetwork()`: Converts API response map to NetworkModel
   - `matchesFilter()`: Client-side filtering logic
   - Reuse: `getStringValue()`, `getBoolValue()`, `getInt64Value()` from existing code

**Step 2: Register Data Source**

File: `/workspace/internal/provider/provider.go`

Add to `DataSources()` method:
```go
func (p *bcmProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        // ... existing data sources ...
        NewCMNetNetworksDataSource,
    }
}
```

**Step 3: Implement Core Read Logic**

```go
func (d *CMNetNetworksDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
    // 1. Get configuration
    var config CMNetNetworksDataSourceModel
    resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)

    // 2. Call BCM API
    body, err := d.client.CallJSONRPC(ctx, "cmnet", "getNetworks")
    if err != nil {
        resp.Diagnostics.AddError("Unable to Read BCM Networks", err.Error())
        return
    }

    // 3. Parse response
    var apiResponse []map[string]interface{}
    json.Unmarshal(body, &apiResponse)

    // 4. Map and filter
    state := CMNetNetworksDataSourceModel{
        ID: types.StringValue("placeholder"),
        Filter: config.Filter,
        Networks: []NetworkModel{},
    }

    for _, netData := range apiResponse {
        network := mapAPIToNetwork(netData)
        if matchesFilter(network, config.Filter) {
            state.Networks = append(state.Networks, network)
        }
    }

    // 5. Set state
    resp.Diagnostics.Append(resp.State.Set(ctx, &state)...)
}
```

**Expected Result**: All tests PASS.

**Verification**:
```bash
TF_ACC=1 BCM_ENDPOINT="https://172.21.15.254:8081" BCM_USERNAME="root" BCM_PASSWORD="Hashicorp123!" \
  go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks
```

### Phase 2.3: REFACTOR - Improve Code Quality

**Objective**: Enhance implementation while keeping all tests green.

**Improvements**:

1. **Enhanced Error Handling**:
   - Add detailed error messages with actionable guidance
   - Include HTTP status codes in authentication error messages
   - Add context about malformed JSON responses
   - Document retry behavior (none - fail fast)

2. **Comprehensive Logging**:
   - Add `tflog.Debug()` for API call initiation
   - Log total vs filtered network counts
   - Log filter criteria being applied

3. **Code Documentation**:
   - Add godoc comments for all exported types
   - Document filter matching behavior
   - Add inline comments for non-obvious logic

4. **Filtering Logic Refinement**:
   - Ensure case-insensitive name matching uses `strings.ToLower()`
   - Handle null filter values correctly (IsNull() checks)
   - Document AND behavior for multiple filters

5. **Schema Documentation**:
   - Complete MarkdownDescription for all attributes
   - Document filter behavior clearly
   - Add examples in descriptions

**Verification**:
```bash
# All tests still pass
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks

# Code quality checks pass
make lint
make fmt
```

## Phase 3: Examples & Documentation

### Example Configurations

**File**: `/workspace/examples/data-sources/bcm_cmnet_networks/data-source.tf`
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

output "network_count" {
  value = length(data.bcm_cmnet_networks.all.networks)
}
```

**File**: `/workspace/examples/data-sources/bcm_cmnet_networks/filtered.tf`
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

# Create map of network names to UUIDs
locals {
  network_map = {
    for net in data.bcm_cmnet_networks.dhcp_networks.networks :
    net.name => net.uuid
  }
}

output "management_network_uuid" {
  value = length(data.bcm_cmnet_networks.management.networks) > 0 ? data.bcm_cmnet_networks.management.networks[0].uuid : null
}

output "dhcp_network_map" {
  value = local.network_map
}
```

## Phase 4: Documentation Generation

### Auto-Generate Provider Documentation

```bash
# Generate documentation using tfplugindocs
make generate

# Or manually:
cd /workspace
go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate
```

**Expected Output**: `/workspace/docs/data-sources/cmnet_networks.md`

**Documentation Sections**:
1. **Description**: Purpose and use cases for the data source
2. **Example Usage**: Code examples from `examples/` directory
3. **Schema**: Complete attribute reference with types and descriptions
4. **Filter Options**: Detailed explanation of filtering behavior
5. **Notes**: Authentication requirements, performance considerations

### Verification

```bash
# Verify docs were generated
ls -la /workspace/docs/data-sources/cmnet_networks.md

# Verify no uncommitted changes (docs should be up to date)
git diff --exit-code docs/
```

## Testing Strategy

### Test Environment Setup

**Environment Variables**:
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

### Test Execution Plan

**Phase 1 (RED)**:
```bash
# Expected: All tests FAIL
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks
```

**Phase 2 (GREEN)**:
```bash
# Expected: All tests PASS
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks
```

**Phase 3 (REFACTOR)**:
```bash
# Expected: All tests still PASS after refactoring
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMNetNetworks

# Code quality
make lint
make fmt
```

**Full Test Suite**:
```bash
# Run all provider tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Check coverage
go test -v -cover ./internal/provider/
```

### Test Data Strategy

**Assumption**: BCM cluster at 172.21.15.254:8081 has at least one network configured.

**Read-Only Approach**:
- Tests DO NOT create or destroy networks
- Tests use existing networks in the BCM cluster
- Filtering tests assume some network names contain common patterns like "management", "compute", or "storage"
- Empty result test uses nonsensical pattern guaranteed not to match

**Test Independence**:
- Each test is independent (no shared state)
- Tests can run in parallel (`-parallel=4`)
- No cleanup needed (read-only operations)

## Implementation Checklist

### Phase 0: Research
- [ ] Run API exploration script for `cmnet.getNetworks`
- [ ] Document actual API response structure in `research.md`
- [ ] Create field mapping table (API → Terraform)
- [ ] Verify helper function availability
- [ ] Resolve all NEEDS CLARIFICATION items

### Phase 1: Design
- [ ] Create `data-model.md` with Network entity definition
- [ ] Create API request/response contracts in `contracts/`
- [ ] Create `quickstart.md` developer guide
- [ ] Run agent context update script

### Phase 2.1: RED
- [ ] Create `data_source_cmnet_networks_test.go`
- [ ] Write `TestAccCMNetNetworksDataSource_Basic`
- [ ] Write `TestAccCMNetNetworksDataSource_NameFilter`
- [ ] Write `TestAccCMNetNetworksDataSource_DHCPFilter`
- [ ] Write `TestAccCMNetNetworksDataSource_NoMatch`
- [ ] Verify all tests FAIL

### Phase 2.2: GREEN
- [ ] Create `data_source_cmnet_networks.go`
- [ ] Define data source structs (CMNetNetworksDataSource, models)
- [ ] Implement `Metadata()` method
- [ ] Implement `Schema()` method with all attributes
- [ ] Implement `Configure()` method
- [ ] Implement `Read()` method with API call and parsing
- [ ] Implement `mapAPIToNetwork()` helper
- [ ] Implement `matchesFilter()` helper
- [ ] Register data source in `provider.go`
- [ ] Verify all tests PASS

### Phase 2.3: REFACTOR
- [ ] Add comprehensive error handling
- [ ] Add debug logging with tflog
- [ ] Add godoc comments
- [ ] Enhance schema descriptions
- [ ] Optimize filtering logic
- [ ] Verify all tests still PASS
- [ ] Run `make lint` and `make fmt`

### Phase 3: Examples
- [ ] Create `examples/data-sources/bcm_cmnet_networks/data-source.tf`
- [ ] Create `examples/data-sources/bcm_cmnet_networks/filtered.tf`
- [ ] Test examples manually with `terraform plan`

### Phase 4: Documentation
- [ ] Run `make generate` to create docs
- [ ] Review generated `docs/data-sources/cmnet_networks.md`
- [ ] Verify documentation accuracy and completeness
- [ ] Commit generated documentation

### Final Verification
- [ ] All acceptance tests pass (100%)
- [ ] Code quality checks pass (lint, fmt)
- [ ] Documentation generated and committed
- [ ] Examples work with live BCM cluster
- [ ] Ready for PR submission

## Key Technical Decisions

### Decision 1: Client-Side Filtering
**Rationale**: BCM API `getNetworks` call does not support filtering parameters. Client-side filtering is the only option and follows the pattern used in `data_source_cmdevice_nodes.go`.

**Implementation**: Retrieve all networks, then filter in Go code before setting state.

### Decision 2: Helper Function Reuse
**Rationale**: Null-safe helper functions already exist in `data_source_cmpart_softwareimages.go` and are accessible to all files in the `provider` package.

**Implementation**: Import and use existing `getStringValue()`, `getBoolValue()`, `getInt64Value()` functions without modification.

### Decision 3: Read-Only Test Strategy
**Rationale**: Data sources are inherently read-only. Creating test networks would require implementing a network resource first, which is out of scope.

**Implementation**: Tests assume at least one network exists in the BCM cluster and use existing networks for validation.

### Decision 4: Fail-Fast Error Handling
**Rationale**: Following specification requirement for no retry logic, provide clear actionable error messages.

**Implementation**:
- Authentication failures: Include HTTP status code, suggest credential verification
- Network issues: Include connection error details, suggest endpoint verification
- JSON parsing errors: Include malformed response context
- No retries, no partial success

### Decision 5: Case-Insensitive Substring Matching
**Rationale**: User expectation for name filtering is typically flexible (e.g., "management" should match "Management-Net", "my-management", "MANAGEMENT").

**Implementation**: Use `strings.ToLower()` for both pattern and network name, then `strings.Contains()`.

## Success Criteria

- ✅ **SC-001**: Users can retrieve all networks from BCM cluster in under 5 seconds (for typical clusters with up to 100 networks)
- ✅ **SC-002**: Data source accurately reflects 100% of network attributes available in the BCM API response
- ✅ **SC-003**: Acceptance tests achieve 100% pass rate in CI/CD pipeline
- ✅ **SC-004**: Documentation is generated automatically and includes all attributes with clear descriptions
- ✅ **SC-005**: Data source follows the same code structure and patterns as existing data sources (verified through code review)
- ✅ **SC-006**: Filtering logic correctly processes all defined filter criteria with zero false positives or negatives
- ✅ **SC-007**: Error messages provide actionable guidance for authentication failures, network issues, and API errors

## Risk Mitigation

### Risk: BCM API response structure differs from specification
**Mitigation**: Phase 0 research includes live API call to verify actual response structure before implementation.

### Risk: No networks exist in test BCM cluster
**Mitigation**: Phase 0 research verifies at least one network exists. Document as prerequisite for acceptance tests.

### Risk: Helper functions unavailable or incompatible
**Mitigation**: Phase 0 research confirms helper function availability and usage patterns from existing data sources.

### Risk: Performance degradation with large network counts
**Mitigation**: Client-side filtering is O(n) where n = number of networks. For 100 networks, performance impact is negligible (<1ms filtering time).

## Stop Condition

This command (`/speckit.plan`) ends after Phase 1 planning artifacts are generated:
- ✅ `research.md` created
- ✅ `data-model.md` created
- ✅ `contracts/` directory created
- ✅ `quickstart.md` created
- ✅ Agent context updated

**Next Steps**: Run `/speckit.tasks` to generate `tasks.md`, then `/speckit.implement` to execute TDD workflow.

## Notes

- This plan assumes the BCM API endpoint `{"service": "cmnet", "call": "getNetworks"}` exists and returns a JSON array (based on API documentation patterns).
- Acceptance tests require a live BCM environment with at least one network configured.
- All file paths are absolute to avoid issues with agent thread working directory resets.
- The TDD workflow strictly follows RED-GREEN-REFACTOR with verification at each phase.
