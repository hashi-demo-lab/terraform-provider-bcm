# Implementation Plan: BCM Partition Resource

**Branch**: `002-cmpart-partition` | **Date**: 2025-11-23 | **Spec**: [spec.md](/workspace/specs/002-cmpart-partition/spec.md)
**Input**: Feature specification from `/specs/002-cmpart-partition/spec.md`

## Summary

Implement the `bcm_cmpart_partition` Terraform resource to manage BCM cluster partitions (organizational units) via the CMPart JSON-RPC API. Partitions are logical groupings that define cluster-wide configuration including node naming conventions, network settings (DNS, NTP), and administrative contacts. The implementation follows TDD principles with full CRUD operations, drift detection, import support, and modern acceptance testing patterns using statecheck/plancheck from terraform-plugin-testing v1.13.3+.

**Technical Approach**: Leverage existing BCM client infrastructure (internal/provider/bcm_client.go) with JSON-RPC calls to cmpart service methods (addPartition, getPartition, updatePartition, removePartition). Use direct UUID-based lookup via args parameter for efficient Read operations. Handle list attributes (admin_email, time_servers, search_domains, name_servers) using terraform-plugin-framework types.List. Follow the proven CRUD pattern from resource_cmpart_softwareimage.go with proper field name mapping (snake_case → camelCase) and BCM entity structure wrapping.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**:
- terraform-plugin-framework v1.16.1 (resource schema, CRUD interfaces, types)
- terraform-plugin-testing v1.13.3 (acceptance tests with statecheck/plancheck)
- BCM JSON-RPC API (cmpart service)

**Storage**: BCM cluster state via JSON-RPC API (endpoint: https://172.21.15.254:8081/json)
**Testing**:
- Acceptance tests with TF_ACC=1 (terraform-plugin-testing framework)
- Modern patterns: statecheck.ExpectKnownValue, plancheck.ExpectEmptyPlan
- Test helpers: createTestBCMClient, generateUniqueTestName, verifyResourceDeleted

**Target Platform**: Linux server (Terraform provider binary)
**Project Type**: Terraform Provider Plugin (single binary with plugin framework)

**Performance Goals**:
- Create operation: <30 seconds (single API call + read-back)
- Read operation: <5 seconds (direct UUID lookup via getPartition(uuid))
- Update operation: <10 seconds (API update + read-back)
- Delete operation: <5 seconds (removePartition call)
- Drift detection: <5 seconds (Read operation detects external changes)

**Constraints**:
- All acceptance tests must pass with 100% success rate
- Import operation must achieve zero attribute mismatches
- Concurrent operations (multiple partitions) must not cause conflicts
- Test environment must be portable (no hardcoded assumptions about existing resources)
- Error messages must be actionable for users (surface BCM API errors clearly)
- State must never contain Unknown values (causes "invalid result object" errors)

**Scale/Scope**:
- Single resource implementation (bcm_cmpart_partition)
- 15 resource attributes (4 identity, 5 config, 4 list arrays, 2 metadata)
- 8 acceptance test scenarios (Create, Read, Update, Delete, Import, Drift, Idempotency, Validation)
- 4 API methods (addPartition, getPartition, updatePartition, removePartition)
- Environment portability: works on any BCM cluster without hardcoded dependencies

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**TDD Adherence**:
- ✅ RED-GREEN-REFACTOR cycle: Write failing acceptance tests before implementation
- ✅ Parallel execution: Acceptance tests can run independently with unique names
- ✅ Test coverage: All CRUD operations + Import + Drift detection
- ✅ Modern patterns: statecheck.ExpectKnownValue, plancheck.ExpectEmptyPlan

**Complexity Gates**:
- ✅ No new projects required (single provider codebase)
- ✅ No new design patterns required (follow resource_cmpart_softwareimage.go pattern)
- ✅ No additional frameworks required (existing terraform-plugin-framework)

**Gate Status**: ✅ PASSED - No violations. Implementation follows established patterns.

## Project Structure

### Documentation (this feature)

```text
specs/002-cmpart-partition/
├── spec.md              # Feature specification (input to /speckit.plan)
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (API method verification, field mapping research)
├── data-model.md        # Phase 1 output (PartitionResourceModel structure)
├── quickstart.md        # Phase 1 output (developer setup guide)
├── contracts/           # Phase 1 output (BCM API contracts for cmpart service)
│   └── cmpart-partition-api.json  # OpenAPI-style BCM API documentation
└── tasks.md             # Phase 2 output (/speckit.tasks command - NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/provider/
│   ├── provider.go                        # Provider definition (register new resource)
│   ├── bcm_client.go                      # JSON-RPC client (existing, reused)
│   ├── test_helpers.go                    # Test utilities (existing, reused)
│   ├── resource_cmpart_partition.go       # NEW: Resource implementation
│   ├── resource_cmpart_partition_test.go  # NEW: Acceptance tests
│   └── data_source_cmpart_partitions.go   # REFERENCE: Schema structure
│
├── examples/
│   └── resources/
│       └── bcm_cmpart_partition/          # NEW: Example Terraform configurations
│           └── resource.tf                # Usage examples
│
├── docs/
│   └── resources/
│       └── bcm_cmpart_partition.md        # GENERATED: tfplugindocs output
│
└── specs/002-cmpart-partition/            # Feature specification and design
    └── [documentation files listed above]
```

**Structure Decision**: Standard Terraform Provider Framework structure (single project). All resource implementations live in `internal/provider/` with corresponding acceptance tests in `*_test.go` files. Examples in `examples/resources/` are auto-processed by tfplugindocs to generate documentation in `docs/`. This structure follows HashiCorp conventions and matches existing resources in the provider.

## Complexity Tracking

> **Fill ONLY if Constitution Check has violations that must be justified**

No violations - table not needed.

---

## Phase 0: Outline & Research

### Objectives

Resolve all "NEEDS CLARIFICATION" items from Technical Context and research BCM API methods for partition management. Verify API method signatures, field name mappings (snake_case → camelCase), and entity structure requirements.

### Research Tasks

#### Task 0.1: Verify BCM CMPart API Methods
**Goal**: Confirm exact API method names and signatures for partition CRUD operations

**Investigation**:
- Query BCM API documentation or exploration scripts (sampleRest/ directory)
- Verify method names: `addPartition`, `getPartition`, `updatePartition`, `removePartition`
- Test args parameter support: `getPartition(uuid)` vs `getPartition()` + client-side filter
- Document response structure for each method

**Deliverable**: API method contracts in `contracts/cmpart-partition-api.json`

**Acceptance Criteria**:
- All 4 CRUD methods documented with request/response examples
- Args parameter usage confirmed for getPartition(uuid)
- Field name mappings verified (e.g., cluster_name → clusterName)

#### Task 0.2: Research Partition Field Mappings
**Goal**: Document Terraform snake_case → BCM API camelCase field name transformations

**Investigation**:
- Review data_source_cmpart_partitions.go schema (lines 54-82) for authoritative field list
- Test field name mappings with real BCM API calls
- Document special cases (e.g., no_zero_conf → noZeroConf)
- Verify list attribute handling (admin_email, time_servers, etc.)

**Deliverable**: Field mapping table in `research.md`

**Acceptance Criteria**:
- Complete mapping table for all 15 resource attributes
- List attribute serialization format documented (JSON array vs comma-separated)
- Edge cases documented (empty lists, null handling)

#### Task 0.3: Research BCM Entity Structure Requirements
**Goal**: Understand BCM API entity wrapper requirements for create/update operations

**Investigation**:
- Review resource_cmpart_softwareimage.go buildAPIEntity() pattern (reference implementation)
- Verify required entity fields: baseType, childType, modified, to_be_removed, revision, uuid
- Test partition-specific entity structure via BCM API
- Document partition entity baseType/childType values

**Deliverable**: Entity structure documentation in `research.md`

**Acceptance Criteria**:
- Entity wrapper structure documented with examples
- baseType confirmed as "Partition" (or actual value from API)
- childType polymorphism understood (if applicable to partitions)
- Required vs optional entity fields identified

#### Task 0.4: Research Concurrent Operation Safety
**Goal**: Verify partition operations are safe for parallel execution (multiple terraform resources)

**Investigation**:
- Review BCM API revision field usage for optimistic concurrency control
- Test creating multiple partitions simultaneously via API
- Document any API rate limits or locking mechanisms
- Verify unique constraint enforcement (partition names must be unique)

**Deliverable**: Concurrency safety documentation in `research.md`

**Acceptance Criteria**:
- Revision-based concurrency control understood
- Unique name constraint behavior documented
- Safe for parallel execution: YES/NO with justification

### Research Outputs

**File**: `specs/002-cmpart-partition/research.md`

**Structure**:
```markdown
# BCM Partition Resource Research

## Decision: API Method Selection
- CREATE: addPartition(entity) - creates new partition, returns UUID
- READ: getPartition(uuid) - direct lookup via args parameter
- UPDATE: updatePartition(entity, force=false) - updates existing partition
- DELETE: removePartition(uuid, force=false) - removes partition

**Rationale**: Direct UUID-based lookup for Read (efficient, follows resource pattern). Force parameter allows override of deletion safety checks.

**Alternatives considered**:
- getPartitions() + client-side filter (REJECTED: inefficient for resources)
- Name-based operations (REJECTED: UUID is stable identifier)

## Decision: Field Name Mapping
[Complete snake_case → camelCase mapping table]

## Decision: Entity Structure
[BCM entity wrapper format with required fields]

## Decision: Concurrency Safety
[Revision-based optimistic locking, safe for parallel execution]
```

---

## Phase 1: Design & Contracts

### Objectives

Generate data model (PartitionResourceModel), API contracts (OpenAPI-style BCM documentation), and quickstart guide. Update agent context with new BCM partition management technology.

### Design Artifacts

#### Artifact 1.1: Data Model
**File**: `specs/002-cmpart-partition/data-model.md`

**Content**:
```markdown
# Partition Resource Data Model

## PartitionResourceModel (Go Struct)

```go
type PartitionResourceModel struct {
    // Identity fields
    ID        types.String `tfsdk:"id"`         // Computed: same as UUID
    UUID      types.String `tfsdk:"uuid"`       // Computed: BCM-assigned unique identifier
    Name      types.String `tfsdk:"name"`       // Required: partition name (unique)
    BaseType  types.String `tfsdk:"base_type"`  // Computed: "Partition"
    ChildType types.String `tfsdk:"child_type"` // Computed: polymorphic type

    // Configuration fields
    ClusterName types.String `tfsdk:"cluster_name"` // Required: cluster display name
    SlaveName   types.String `tfsdk:"slave_name"`   // Optional: node naming prefix (default: "node")
    SlaveDigits types.Int64  `tfsdk:"slave_digits"` // Optional: node numbering digits (default: 3)
    RelayHost   types.String `tfsdk:"relay_host"`   // Optional: SMTP relay hostname
    NoZeroConf  types.Bool   `tfsdk:"no_zero_conf"` // Optional: disable Zeroconf (default: false)

    // Network configuration (list attributes)
    AdminEmail    types.List `tfsdk:"admin_email"`    // Optional: List[String] admin emails
    TimeServers   types.List `tfsdk:"time_servers"`   // Optional: List[String] NTP servers
    SearchDomains types.List `tfsdk:"search_domains"` // Optional: List[String] DNS search domains
    NameServers   types.List `tfsdk:"name_servers"`   // Optional: List[String] DNS resolvers

    // Metadata fields
    CreationTime types.Int64  `tfsdk:"creation_time"` // Computed: Unix timestamp
    Revision     types.String `tfsdk:"revision"`      // Computed: concurrency control version
    Modified     types.Bool   `tfsdk:"modified"`      // Computed: dirty flag
    ToBeRemoved  types.Bool   `tfsdk:"to_be_removed"` // Computed: deletion flag
    Notes        types.String `tfsdk:"notes"`         // Optional: description/notes
}
```

## Entity Relationships

**Partition** (1) → (N) **Nodes**: Partitions organize nodes into logical groups
**Partition** (1) → (1) **ClusterConfig**: Each partition has cluster-wide settings

## Validation Rules

- `name`: Length 1-255 characters, must be unique within BCM cluster
- `cluster_name`: Required, non-empty string
- `slave_digits`: Optional, 1-10 range (reasonable for node numbering)
- List attributes: Can be empty lists (valid configuration state)

## State Transitions

```
[NULL] --Create--> [EXISTS] --Update--> [EXISTS] --Delete--> [NULL]
                       |
                       └--Drift Detected--> [EXISTS with diff] --Reconcile--> [EXISTS]
```
```

#### Artifact 1.2: API Contracts
**File**: `specs/002-cmpart-partition/contracts/cmpart-partition-api.json`

**Content**: OpenAPI-style BCM JSON-RPC API documentation
```json
{
  "service": "cmpart",
  "methods": {
    "addPartition": {
      "description": "Create a new partition in BCM cluster",
      "request": {
        "service": "cmpart",
        "call": "addPartition",
        "args": [
          {
            "baseType": "Partition",
            "childType": "",
            "modified": true,
            "to_be_removed": false,
            "revision": "",
            "uuid": "",
            "name": "engineering",
            "clusterName": "hpc-prod",
            "slaveName": "node",
            "slaveDigits": 3,
            "relayHost": "",
            "noZeroConf": false,
            "adminEmail": ["admin@example.com"],
            "timeServers": ["ntp1.example.com"],
            "searchDomains": ["example.com"],
            "nameServers": ["8.8.8.8"],
            "notes": "Engineering partition"
          },
          false
        ]
      },
      "response": {
        "uuid": "550e8400-e29b-41d4-a716-446655440000",
        "name": "engineering",
        "clusterName": "hpc-prod",
        "...": "..."
      }
    },
    "getPartition": {
      "description": "Retrieve partition by UUID (direct lookup)",
      "request": {
        "service": "cmpart",
        "call": "getPartition",
        "args": ["550e8400-e29b-41d4-a716-446655440000"]
      },
      "response": {
        "uuid": "550e8400-e29b-41d4-a716-446655440000",
        "name": "engineering",
        "clusterName": "hpc-prod",
        "baseType": "Partition",
        "childType": "",
        "creationTime": 1705852800,
        "revision": "v1",
        "modified": false,
        "to_be_removed": false,
        "...": "..."
      }
    },
    "updatePartition": {
      "description": "Update existing partition configuration",
      "request": {
        "service": "cmpart",
        "call": "updatePartition",
        "args": [
          {
            "uuid": "550e8400-e29b-41d4-a716-446655440000",
            "baseType": "Partition",
            "childType": "",
            "modified": true,
            "to_be_removed": false,
            "revision": "v1",
            "name": "engineering",
            "clusterName": "hpc-prod-updated",
            "...": "..."
          },
          false
        ]
      },
      "response": "success or error"
    },
    "removePartition": {
      "description": "Delete partition from BCM cluster",
      "request": {
        "service": "cmpart",
        "call": "removePartition",
        "args": ["550e8400-e29b-41d4-a716-446655440000", false]
      },
      "response": "success or error"
    }
  },
  "field_mappings": {
    "cluster_name": "clusterName",
    "slave_name": "slaveName",
    "slave_digits": "slaveDigits",
    "relay_host": "relayHost",
    "no_zero_conf": "noZeroConf",
    "admin_email": "adminEmail",
    "time_servers": "timeServers",
    "search_domains": "searchDomains",
    "name_servers": "nameServers",
    "creation_time": "creationTime",
    "to_be_removed": "to_be_removed"
  }
}
```

#### Artifact 1.3: Quickstart Guide
**File**: `specs/002-cmpart-partition/quickstart.md`

**Content**:
```markdown
# BCM Partition Resource - Developer Quickstart

## Prerequisites

- Go 1.24.0+
- BCM cluster access (endpoint: https://172.21.15.254:8081)
- BCM credentials (username/password)
- Terraform CLI (for manual testing)

## Environment Setup

```bash
# Set BCM credentials for acceptance tests
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="your-password"

# Build and install provider
make install
```

## TDD Workflow

### RED Phase: Write Failing Tests

```bash
# Create test file first
vim internal/provider/resource_cmpart_partition_test.go

# Run tests (expect failures)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition
```

### GREEN Phase: Minimal Implementation

```bash
# Create resource file
vim internal/provider/resource_cmpart_partition.go

# Implement minimal CRUD (hardcoded responses OK)
# Run tests until passing
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition
```

### REFACTOR Phase: API Integration

```bash
# Add real BCM API calls
# Improve error handling
# Add validation

# Run all tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

## Manual Testing

```hcl
# examples/resources/bcm_cmpart_partition/resource.tf
terraform {
  required_providers {
    bcm = {
      source = "hashicorp.com/nvidia/bcm"
    }
  }
}

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "your-password"
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "example" {
  name         = "engineering"
  cluster_name = "hpc-prod"
  slave_name   = "compute"
  slave_digits = 4

  admin_email    = ["admin@example.com"]
  time_servers   = ["ntp1.example.com", "ntp2.example.com"]
  name_servers   = ["8.8.8.8", "8.8.4.4"]
  search_domains = ["example.com"]

  notes = "Engineering cluster partition"
}

output "partition_uuid" {
  value = bcm_cmpart_partition.example.uuid
}
```

```bash
# Test the example
cd examples/resources/bcm_cmpart_partition
terraform init
terraform plan
terraform apply
terraform import bcm_cmpart_partition.example <uuid>
terraform destroy
```

## File Structure

```
internal/provider/
├── resource_cmpart_partition.go       # Main implementation (500-800 lines)
│   ├── PartitionResourceModel         # Data model struct
│   ├── Schema()                       # Terraform schema definition
│   ├── Create()                       # Create partition via BCM API
│   ├── Read()                         # Read partition state (drift detection)
│   ├── Update()                       # Update partition configuration
│   ├── Delete()                       # Delete partition
│   ├── ImportState()                  # Import existing partition
│   ├── readPartition()                # Helper: fetch and map API response
│   └── buildAPIEntity()               # Helper: construct BCM entity structure
│
└── resource_cmpart_partition_test.go  # Acceptance tests (600-800 lines)
    ├── TestAccCMPartPartition_Basic                # Create, Read, Import
    ├── TestAccCMPartPartition_Update               # Update configuration
    ├── TestAccCMPartPartition_NetworkSettings      # List attributes
    ├── TestAccCMPartPartition_DriftDetection       # External modification
    ├── TestAccCMPartPartition_Idempotency          # No changes on re-apply
    ├── TestAccCMPartPartition_ValidationErrors     # Invalid inputs
    ├── testAccCheckPartitionDestroy()              # Verify deletion
    └── testAccPartitionConfig*()                   # Config helpers
```

## Reference Implementations

- **Schema Pattern**: `data_source_cmpart_partitions.go` (lines 54-180)
- **CRUD Pattern**: `resource_cmpart_softwareimage.go` (full resource lifecycle)
- **Test Pattern**: `resource_cmpart_softwareimage_test.go` (modern statecheck/plancheck)
- **API Client**: `bcm_client.go` (JSON-RPC with args parameter)
- **Test Helpers**: `test_helpers.go` (createTestBCMClient, generateUniqueTestName)

## Common Pitfalls

1. **Field Name Mapping**: Always map snake_case → camelCase (cluster_name → clusterName)
2. **Unknown Values**: Never propagate Unknown values to state (causes errors)
3. **Entity Structure**: Include all required BCM entity fields (baseType, childType, etc.)
4. **List Attributes**: Use types.List with proper ElementType (types.StringType)
5. **Import**: Use UUID as identifier, not name (UUID is stable)
6. **Drift Tests**: Wait 2 seconds after BCM API modification for consistency
7. **Test Names**: Always use generateUniqueTestName() for environment portability

## Documentation Generation

```bash
# After implementation
make generate

# Verify docs
cat docs/resources/bcm_cmpart_partition.md
```
```

#### Artifact 1.4: Agent Context Update
**Script**: `.specify/scripts/bash/update-agent-context.sh copilot`

**Execution**: Run script to update agent-specific context file with partition management knowledge

**Content to Add**:
```markdown
## BCM Partition Management (bcm_cmpart_partition)

### Resource Pattern
- Service: cmpart
- Methods: addPartition, getPartition, updatePartition, removePartition
- Identifier: UUID (stable, used for all operations)
- Read Strategy: Direct lookup via getPartition(uuid) with args parameter

### Field Mappings (Terraform → BCM API)
- cluster_name → clusterName
- slave_name → slaveName
- slave_digits → slaveDigits
- no_zero_conf → noZeroConf
- admin_email → adminEmail (List[String])
- time_servers → timeServers (List[String])
- search_domains → searchDomains (List[String])
- name_servers → nameServers (List[String])

### List Attribute Handling
- Use types.List with ElementType: types.StringType
- Serialize as JSON arrays for BCM API
- Empty lists are valid (no configuration state)

### Entity Structure
```go
entity := map[string]interface{}{
    "baseType":      "Partition",
    "childType":     "",
    "modified":      true,
    "to_be_removed": false,
    "revision":      partitionData["revision"],
    "uuid":          partitionData["uuid"],
    // ... partition fields in camelCase
}
```
```

### Design Outputs

**Files Generated**:
1. `specs/002-cmpart-partition/data-model.md` - PartitionResourceModel struct and validation
2. `specs/002-cmpart-partition/contracts/cmpart-partition-api.json` - BCM API documentation
3. `specs/002-cmpart-partition/quickstart.md` - Developer onboarding guide
4. Updated agent context file (Copilot/Cursor/etc.) with partition management knowledge

---

## Phase 2: Implementation Strategy (Post-Planning)

*Note: This section provides guidance for the /speckit.tasks command. The plan.md stops before task generation.*

### RED Phase: Acceptance Tests First

**Test Files**: `internal/provider/resource_cmpart_partition_test.go`

**Test Coverage** (8 test functions):
1. `TestAccCMPartPartition_Basic` - Create, Read, Import workflow
2. `TestAccCMPartPartition_Update` - Update cluster_name, notes, slave_name
3. `TestAccCMPartPartition_NetworkSettings` - List attributes (admin_email, time_servers, etc.)
4. `TestAccCMPartPartition_DriftDetection` - External modification via BCM API
5. `TestAccCMPartPartition_Idempotency` - ExpectEmptyPlan after Create/Update
6. `TestAccCMPartPartition_ValidationErrors` - Invalid name, empty cluster_name
7. `TestAccCMPartPartition_IDConsistency` - Track ID across Create/Import/Update
8. `testAccCheckPartitionDestroy` - Verify deletion with exponential backoff

**Modern Test Patterns**:
- Use `statecheck.ExpectKnownValue()` for type-safe state verification
- Use `plancheck.ExpectEmptyPlan()` for idempotency checks
- Use `statecheck.CompareValue()` for ID consistency tracking
- Environment portable: no hardcoded resource assumptions

### GREEN Phase: Minimal Implementation

**Implementation File**: `internal/provider/resource_cmpart_partition.go`

**Minimal CRUD**:
- Schema: Define all 15 attributes with proper types and descriptions
- Create: Hardcoded UUID, set plan values to state
- Read: Return existing state as-is
- Update: Set plan values to state
- Delete: No-op
- ImportState: resource.ImportStatePassthroughID

**Goal**: Make acceptance tests pass with minimal code (no API calls yet)

### REFACTOR Phase: API Integration

**Full CRUD Implementation**:

1. **Create**:
   - Build BCM entity with buildAPIEntity() helper
   - Call `r.client.CallJSONRPC(ctx, "cmpart", "addPartition", entity, false)`
   - Extract UUID from response
   - Call readPartition() to populate computed fields
   - Set state

2. **Read**:
   - Call `r.client.CallJSONRPC(ctx, "cmpart", "getPartition", state.UUID.ValueString())`
   - Parse response with proper field name mapping
   - Handle list attributes (unmarshal JSON arrays to types.List)
   - Set state (drift detection happens automatically)

3. **Update**:
   - Get current state UUID
   - Build updated entity with buildAPIEntity()
   - Call `r.client.CallJSONRPC(ctx, "cmpart", "updatePartition", entity, false)`
   - Call readPartition() to refresh state
   - Set state

4. **Delete**:
   - Call `r.client.CallJSONRPC(ctx, "cmpart", "removePartition", state.UUID.ValueString(), false)`
   - Handle deletion errors (partition has nodes, etc.)

5. **Helpers**:
   - `readPartition(ctx, model, diagnostics)` - Fetch and map API response
   - `buildAPIEntity(model) map[string]interface{}` - Construct BCM entity structure
   - Field mappers for list attributes (adminEmail, timeServers, etc.)

### Registration & Examples

**Provider Registration**: `internal/provider/provider.go`
```go
func (p *BCMProvider) Resources(ctx context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        // ... existing resources ...
        NewCMPartPartitionResource,  // ADD THIS
    }
}
```

**Example Configuration**: `examples/resources/bcm_cmpart_partition/resource.tf`
```hcl
resource "bcm_cmpart_partition" "engineering" {
  name         = "engineering"
  cluster_name = "HPC Production Cluster"
  slave_name   = "compute"
  slave_digits = 4

  admin_email    = ["admin@example.com", "ops@example.com"]
  time_servers   = ["ntp1.example.com", "ntp2.example.com"]
  name_servers   = ["8.8.8.8", "8.8.4.4"]
  search_domains = ["example.com", "corp.example.com"]

  relay_host  = "smtp.example.com"
  no_zero_conf = false

  notes = "Engineering team partition for GPU workloads"
}
```

### Documentation Generation

**Command**: `make generate`

**Generated Files**:
- `docs/resources/bcm_cmpart_partition.md` - Auto-generated from schema + examples
- Documentation includes all attributes, import syntax, and examples

---

## Architecture Design

### Component Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                    Terraform Core                               │
│  (applies configuration, tracks state, detects drift)           │
└───────────────────────────┬─────────────────────────────────────┘
                            │ Plugin Protocol (gRPC)
                            │
┌───────────────────────────▼─────────────────────────────────────┐
│              BCM Provider (terraform-provider-bcm)              │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  CMPartPartitionResource                                   │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │ │
│  │  │   Schema()   │  │   Create()   │  │     Read()      │  │ │
│  │  │  (defines    │  │  (addPart)   │  │  (getPart UUID) │  │ │
│  │  │  attributes) │  │              │  │  (drift detect) │  │ │
│  │  └──────────────┘  └──────────────┘  └─────────────────┘  │ │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────────┐  │ │
│  │  │   Update()   │  │   Delete()   │  │  ImportState()  │  │ │
│  │  │ (updatePart) │  │ (removePart) │  │  (UUID lookup)  │  │ │
│  │  └──────────────┘  └──────────────┘  └─────────────────┘  │ │
│  │                                                             │ │
│  │  ┌──────────────────────────────────────────────────────┐  │ │
│  │  │  Helper Functions                                    │  │ │
│  │  │  - readPartition(ctx, model, diags)                  │  │ │
│  │  │  - buildAPIEntity(model) → entity map                │  │ │
│  │  │  - mapListAttribute(tfList) → []string               │  │ │
│  │  └──────────────────────────────────────────────────────┘  │ │
│  └────────────────────────────┬───────────────────────────────┘ │
│                               │                                 │
│  ┌────────────────────────────▼───────────────────────────────┐ │
│  │  BCMClient (JSON-RPC client with cookie auth)             │ │
│  │  - CallJSONRPC(ctx, service, method, args...)             │ │
│  │  - Automatic cookie management (cm-login-token)           │ │
│  └────────────────────────────┬───────────────────────────────┘ │
└─────────────────────────────┬─┴─────────────────────────────────┘
                              │ HTTPS (JSON-RPC)
                              │
┌─────────────────────────────▼─────────────────────────────────────┐
│                    BCM Cluster API                                │
│  ┌──────────────────────────────────────────────────────────────┐ │
│  │  cmpart Service                                              │ │
│  │  - addPartition(entity, force) → uuid                        │ │
│  │  - getPartition(uuid) → partition data                       │ │
│  │  - updatePartition(entity, force) → success                  │ │
│  │  - removePartition(uuid, force) → success                    │ │
│  └──────────────────────────────────────────────────────────────┘ │
│                                                                    │
│  [Persistent Storage: Partition Configuration, Node Assignments]  │
└────────────────────────────────────────────────────────────────────┘
```

### Data Flow: Create Operation

```
1. User writes Terraform config:
   resource "bcm_cmpart_partition" "eng" {
     name         = "engineering"
     cluster_name = "hpc-prod"
   }

2. Terraform plan → calls Schema() to validate attributes

3. Terraform apply → calls Create(plan)
   ┌─────────────────────────────────────────────────────────┐
   │ Create(ctx, req, resp)                                  │
   │  1. Extract plan data → PartitionResourceModel          │
   │  2. buildAPIEntity(model) → BCM entity structure        │
   │     {                                                    │
   │       "baseType": "Partition",                          │
   │       "childType": "",                                  │
   │       "modified": true,                                 │
   │       "name": "engineering",                            │
   │       "clusterName": "hpc-prod",  // snake→camel        │
   │       ...                                               │
   │     }                                                    │
   │  3. CallJSONRPC("cmpart", "addPartition", entity, false)│
   │  4. Extract UUID from response                          │
   │  5. readPartition(ctx, model, diags) → fetch fresh data │
   │  6. Set state with UUID, computed fields                │
   └─────────────────────────────────────────────────────────┘

4. BCM API processes addPartition:
   - Validates partition name uniqueness
   - Creates partition record in database
   - Returns partition data with UUID

5. Terraform stores state with all attributes populated
```

### Data Flow: Read Operation (Drift Detection)

```
1. Terraform refresh or plan → calls Read(state)
   ┌─────────────────────────────────────────────────────────┐
   │ Read(ctx, req, resp)                                    │
   │  1. Extract state UUID                                  │
   │  2. CallJSONRPC("cmpart", "getPartition", uuid)         │
   │  3. Parse response (camelCase → snake_case)             │
   │     - clusterName → cluster_name                        │
   │     - adminEmail → admin_email (JSON array → types.List)│
   │  4. Compare fetched data vs state                       │
   │  5. If different → drift detected (plan shows changes)  │
   │  6. Set state with current BCM values                   │
   └─────────────────────────────────────────────────────────┘

2. BCM API returns current partition configuration

3. Terraform detects differences:
   - State: cluster_name = "hpc-prod"
   - BCM:   clusterName = "hpc-prod-modified"
   → Plan shows: cluster_name will be updated from "hpc-prod-modified" to "hpc-prod"
```

### Error Handling Strategy

```
┌─────────────────────────────────────────────────────────────────┐
│ Error Handling Layers                                           │
│                                                                  │
│  1. Schema Validation (pre-API call)                            │
│     - name length (1-255 chars)                                 │
│     - slave_digits range (1-10)                                 │
│     - cluster_name non-empty                                    │
│     → resp.Diagnostics.AddError() with user-friendly message    │
│                                                                  │
│  2. BCM API Errors (during CRUD operations)                     │
│     - Duplicate partition name (409 conflict)                   │
│     - Partition not found (404)                                 │
│     - Partition has nodes (400 cannot delete)                   │
│     - Invalid entity structure (400 validation error)           │
│     → Parse BCM error response, surface to user with context    │
│                                                                  │
│  3. Network/Connection Errors                                   │
│     - Connection timeout                                        │
│     - Authentication failure                                    │
│     - TLS verification failure                                  │
│     → Log technical details, show actionable message to user    │
│                                                                  │
│  4. State Consistency Errors                                    │
│     - Unknown values in state (NEVER propagate)                 │
│     - Revision conflicts (concurrent updates)                   │
│     - UUID mismatch on import                                   │
│     → Validate state values, provide recovery suggestions       │
└─────────────────────────────────────────────────────────────────┘
```

**Example Error Messages**:
```
✅ Good: "Partition name 'engineering' already exists. Choose a unique name or import the existing partition with: terraform import bcm_cmpart_partition.eng <uuid>"

❌ Bad: "Error: API returned 409"

✅ Good: "Cannot delete partition 'engineering' (uuid: 550e8400-...) because it has 15 active nodes assigned. Remove nodes first or use force=true parameter."

❌ Bad: "Error: removePartition failed"
```

---

## Testing Strategy

### Test Pyramid

```
                    ┌──────────────────┐
                    │  Acceptance Tests│  ← 8 test functions (TF_ACC=1)
                    │  (E2E with BCM)  │     Full CRUD + Drift + Import
                    └────────┬─────────┘
                             │
                ┌────────────▼────────────────┐
                │   Integration Tests         │  ← (Optional) Mock BCM API
                │   (API contract validation) │     Verify JSON-RPC formatting
                └────────────┬────────────────┘
                             │
            ┌────────────────▼────────────────────┐
            │      Unit Tests                     │  ← Helper function tests
            │  (buildAPIEntity, field mappers)    │     No API dependencies
            └─────────────────────────────────────┘
```

### Acceptance Test Scenarios

#### Test 1: Basic CRUD Workflow
```go
func TestAccCMPartPartition_Basic(t *testing.T) {
    partitionName := generateUniqueTestName("test-partition")
    compareID := statecheck.CompareValue(compare.ValuesSame())

    resource.Test(t, resource.TestCase{
        PreCheck: func() { testAccPreCheck(t) },
        ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
        CheckDestroy: testAccCheckPartitionDestroy,
        Steps: []resource.TestStep{
            // Step 1: Create and Read
            {
                Config: testAccPartitionConfigBasic(partitionName, "HPC Cluster"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("name"),
                        knownvalue.StringExact(partitionName),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("cluster_name"),
                        knownvalue.StringExact("HPC Cluster"),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("uuid"),
                        knownvalue.NotNull(),
                    ),
                    compareID.AddStateValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
            // Step 2: Idempotency check
            {
                Config: testAccPartitionConfigBasic(partitionName, "HPC Cluster"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectEmptyPlan(),
                    },
                },
            },
            // Step 3: Import
            {
                ResourceName:      "bcm_cmpart_partition.test",
                ImportState:       true,
                ImportStateVerify: true,
                ConfigStateChecks: []statecheck.StateCheck{
                    compareID.AddStateValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("id"),
                    ),
                },
            },
        },
    })
}
```

#### Test 2: Update Configuration
```go
func TestAccCMPartPartition_Update(t *testing.T) {
    partitionName := generateUniqueTestName("test-partition")

    resource.Test(t, resource.TestCase{
        // ... test case setup ...
        Steps: []resource.TestStep{
            // Create with initial values
            {
                Config: testAccPartitionConfigBasic(partitionName, "Cluster A"),
                // ... state checks ...
            },
            // Update cluster_name
            {
                Config: testAccPartitionConfigBasic(partitionName, "Cluster B"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("cluster_name"),
                        knownvalue.StringExact("Cluster B"),
                    ),
                },
            },
            // Idempotency after update
            {
                Config: testAccPartitionConfigBasic(partitionName, "Cluster B"),
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

#### Test 3: Network Settings (List Attributes)
```go
func TestAccCMPartPartition_NetworkSettings(t *testing.T) {
    partitionName := generateUniqueTestName("test-partition")

    resource.Test(t, resource.TestCase{
        // ... test case setup ...
        Steps: []resource.TestStep{
            {
                Config: testAccPartitionConfigNetworkSettings(
                    partitionName,
                    []string{"admin@example.com"},
                    []string{"ntp1.example.com", "ntp2.example.com"},
                ),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("admin_email"),
                        knownvalue.ListSizeExact(1),
                    ),
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("time_servers"),
                        knownvalue.ListSizeExact(2),
                    ),
                },
            },
        },
    })
}
```

#### Test 4: Drift Detection
```go
func TestAccCMPartPartition_DriftDetection(t *testing.T) {
    partitionName := generateUniqueTestName("test-partition")

    resource.Test(t, resource.TestCase{
        // ... test case setup ...
        Steps: []resource.TestStep{
            // Create with initial notes
            {
                Config: testAccPartitionConfigWithNotes(partitionName, "Initial notes"),
                // ... state checks ...
            },
            // Modify externally via BCM API
            {
                PreConfig: func() {
                    client := createTestBCMClient(t)
                    ctx := context.Background()

                    // Get partition UUID
                    uuid := getResourceUUIDByName(t, "cmpart", "getPartition", partitionName)

                    // Fetch full partition data
                    body, _ := client.CallJSONRPC(ctx, "cmpart", "getPartition", uuid)
                    var partitionData map[string]interface{}
                    json.Unmarshal(body, &partitionData)

                    // Modify notes field (snake_case → camelCase!)
                    partitionData["notes"] = "Modified externally"

                    // Build BCM entity structure
                    entity := map[string]interface{}{
                        "baseType":      "Partition",
                        "childType":     "",
                        "modified":      true,
                        "to_be_removed": false,
                        "revision":      partitionData["revision"],
                        "uuid":          uuid,
                    }
                    for k, v := range partitionData {
                        if k != "uuid" {
                            entity[k] = v
                        }
                    }

                    // Update via BCM API
                    client.CallJSONRPC(ctx, "cmpart", "updatePartition", entity, false)

                    // Wait for eventual consistency
                    time.Sleep(2 * time.Second)

                    t.Logf("[DEBUG] Modified notes externally to: %v", entity["notes"])
                },
                Config: testAccPartitionConfigWithNotes(partitionName, "Initial notes"),
                ConfigPlanChecks: resource.ConfigPlanChecks{
                    PreApply: []plancheck.PlanCheck{
                        plancheck.ExpectNonEmptyPlan(), // Drift detected!
                    },
                },
            },
            // Terraform restores desired state
            {
                Config: testAccPartitionConfigWithNotes(partitionName, "Initial notes"),
                ConfigStateChecks: []statecheck.StateCheck{
                    statecheck.ExpectKnownValue(
                        "bcm_cmpart_partition.test",
                        tfjsonpath.New("notes"),
                        knownvalue.StringExact("Initial notes"),
                    ),
                },
            },
        },
    })
}
```

#### Test 5: CheckDestroy
```go
func testAccCheckPartitionDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    var errors []string
    resourceCount := 0

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmpart_partition" {
            continue
        }

        resourceCount++
        uuid := rs.Primary.ID

        // Verify deletion with exponential backoff
        deleted := verifyResourceDeleted(
            context.Background(),
            client,
            "cmpart",
            "getPartition",
            uuid,
            4, // retry count (15s total)
        )

        if !deleted {
            errors = append(errors, fmt.Sprintf(
                "Partition still exists after destroy. UUID: %s",
                uuid,
            ))
        }
    }

    if len(errors) > 0 {
        return fmt.Errorf("CheckDestroy failures:\n  - %s", strings.Join(errors, "\n  - "))
    }

    return nil
}
```

### Test Helper Functions

```go
// Generate Terraform config for basic partition
func testAccPartitionConfigBasic(name, clusterName string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "test" {
  name         = %[4]q
  cluster_name = %[5]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        clusterName,
    )
}

// Generate Terraform config with network settings
func testAccPartitionConfigNetworkSettings(name string, adminEmails, timeServers []string) string {
    adminEmailsHCL := fmt.Sprintf("[%s]", strings.Join(quoteStrings(adminEmails), ", "))
    timeServersHCL := fmt.Sprintf("[%s]", strings.Join(quoteStrings(timeServers), ", "))

    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmpart_partition" "test" {
  name         = %[4]q
  cluster_name = "Test Cluster"

  admin_email  = %[5]s
  time_servers = %[6]s
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        name,
        adminEmailsHCL,
        timeServersHCL,
    )
}

// Helper: Quote string slice for HCL
func quoteStrings(strs []string) []string {
    quoted := make([]string, len(strs))
    for i, s := range strs {
        quoted[i] = fmt.Sprintf("%q", s)
    }
    return quoted
}
```

### Test Execution Strategy

**Local Development**:
```bash
# Run specific test during development
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition_Basic

# Run all partition tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition

# Run all provider tests (CI/CD)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

**Parallel Execution**:
```bash
# Run tests in parallel (4 concurrent)
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./internal/provider/
```

**Environment Portability**:
- No hardcoded partition names (use generateUniqueTestName)
- No assumptions about existing BCM resources
- Provider config included in every test
- Tests clean up after themselves (CheckDestroy verification)

---

## Risk Mitigation

### Risk 1: BCM API Method Names Unknown
**Severity**: HIGH
**Probability**: MEDIUM

**Impact**: Cannot implement CRUD operations without correct API method names

**Mitigation**:
- Phase 0 Task 0.1: Verify exact API method names via BCM exploration
- Reference sampleRest/ directory for working API examples
- Test API calls manually before implementing resource
- Document API contracts in contracts/cmpart-partition-api.json

**Contingency**: If API methods differ from assumptions, update plan.md with actual method names before implementation

---

### Risk 2: Field Name Mapping Errors
**Severity**: MEDIUM
**Probability**: MEDIUM

**Impact**: Drift detection fails, updates don't work correctly due to snake_case ↔ camelCase mismatches

**Mitigation**:
- Phase 0 Task 0.2: Document complete field mapping table
- Reference test_helpers.go field mapping documentation
- Add drift detection test to catch mapping errors early
- Test with real BCM API during Phase 0

**Contingency**: Update field mappings in buildAPIEntity() and readPartition() based on actual API responses

---

### Risk 3: List Attribute Serialization Issues
**Severity**: MEDIUM
**Probability**: LOW

**Impact**: admin_email, time_servers, etc. may not serialize correctly to/from BCM API

**Mitigation**:
- Phase 0 Task 0.2: Test list attribute API format (JSON array vs comma-separated)
- Reference data_source_cmpart_partitions.go for working list attribute handling
- Add dedicated acceptance test for network settings (Test 3)
- Handle empty lists explicitly (valid configuration state)

**Contingency**: Adjust list serialization logic in buildAPIEntity() based on BCM API requirements

---

### Risk 4: Partition Deletion Constraints
**Severity**: LOW
**Probability**: MEDIUM

**Impact**: BCM may reject partition deletion if nodes are assigned, causing test failures

**Mitigation**:
- Test deletion constraints during Phase 0
- Document deletion prerequisites in resource documentation
- Implement clear error messages: "Cannot delete partition with active nodes"
- Add acceptance test for deletion failure scenarios

**Contingency**: Add force parameter support if needed for test cleanup

---

### Risk 5: Concurrent Test Execution Conflicts
**Severity**: LOW
**Probability**: LOW

**Impact**: Parallel test runs may create partition name conflicts

**Mitigation**:
- Use generateUniqueTestName() for all test resources (timestamp + nanoseconds)
- Verify partition name uniqueness constraint during Phase 0
- Document test cleanup requirements in quickstart.md
- Implement robust CheckDestroy with exponential backoff

**Contingency**: Run tests sequentially if conflicts persist

---

### Risk 6: Test Environment Portability Issues
**Severity**: LOW
**Probability**: LOW

**Impact**: Tests fail on different BCM cluster configurations due to hardcoded assumptions

**Mitigation**:
- Avoid hardcoded resource assumptions (no "default" partition references)
- Create all test dependencies within test itself
- Use environment variables for BCM credentials (no hardcoded values)
- Document test requirements in CLAUDE.md

**Contingency**: Update tests to be more defensive (check resource existence before using)

---

## Success Metrics

### Implementation Success Criteria

| Metric | Target | Validation Method |
|--------|--------|-------------------|
| Acceptance Test Pass Rate | 100% | `TF_ACC=1 go test -v ./internal/provider/` |
| Test Coverage (CRUD) | All operations tested | 8 test functions covering Create/Read/Update/Delete/Import/Drift |
| Import Zero Mismatches | 0 attribute mismatches | ImportStateVerify: true in acceptance tests |
| Documentation Generated | Auto-generated docs complete | `make generate` produces docs/resources/bcm_cmpart_partition.md |
| Drift Detection | <5 seconds | Read operation detects external changes in single API call |
| Concurrent Operations | No conflicts | Multiple partitions managed in parallel without errors |
| Error Messages | Actionable | Users can understand and resolve errors without consulting code |
| Environment Portability | Works on any cluster | Tests pass on multiple BCM environments without modification |

### Performance Benchmarks

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Create | <30s | Time from terraform apply to state written |
| Read | <5s | Single getPartition(uuid) API call |
| Update | <10s | API call + read-back time |
| Delete | <5s | removePartition API call time |
| Import | <10s | UUID lookup + state population |
| Drift Detection | <5s | Read operation during plan |

### Quality Gates

**Before Merge**:
- ✅ All 8 acceptance tests passing
- ✅ Documentation generated and reviewed
- ✅ Examples tested manually
- ✅ Pre-commit hooks passing (lint, fmt)
- ✅ No hardcoded values in tests
- ✅ CheckDestroy verification working

**Post-Merge**:
- ✅ CI/CD pipeline passes
- ✅ No regressions in existing resources
- ✅ Documentation published to registry (if applicable)

---

## Appendix

### BCM API Field Mappings (Complete)

| Terraform Attribute (snake_case) | BCM API Field (camelCase) | Type | Notes |
|----------------------------------|---------------------------|------|-------|
| id | uuid | string | Computed, same as uuid |
| uuid | uuid | string | Computed, BCM-assigned identifier |
| name | name | string | Required, partition name |
| base_type | baseType | string | Computed, always "Partition" |
| child_type | childType | string | Computed, polymorphic type |
| cluster_name | clusterName | string | Required, cluster display name |
| slave_name | slaveName | string | Optional, node prefix (default: "node") |
| slave_digits | slaveDigits | int64 | Optional, node numbering digits (default: 3) |
| relay_host | relayHost | string | Optional, SMTP relay |
| no_zero_conf | noZeroConf | bool | Optional, disable Zeroconf (default: false) |
| admin_email | adminEmail | []string | Optional, admin contact emails |
| time_servers | timeServers | []string | Optional, NTP servers |
| search_domains | searchDomains | []string | Optional, DNS search domains |
| name_servers | nameServers | []string | Optional, DNS resolvers |
| creation_time | creationTime | int64 | Computed, Unix timestamp |
| revision | revision | string | Computed, concurrency version |
| modified | modified | bool | Computed, dirty flag |
| to_be_removed | to_be_removed | bool | Computed, deletion flag |
| notes | notes | string | Optional, description |

### Reference Files

**Schema Reference**: `internal/provider/data_source_cmpart_partitions.go` (lines 54-180)
**CRUD Pattern Reference**: `internal/provider/resource_cmpart_softwareimage.go`
**Test Pattern Reference**: `resource_cmpart_softwareimage_test.go`
**API Client Reference**: `internal/provider/bcm_client.go`
**Test Helpers Reference**: `internal/provider/test_helpers.go`

### Development Checklist

**Phase 0 (Research)**:
- [ ] Verify BCM API methods (addPartition, getPartition, updatePartition, removePartition)
- [ ] Test args parameter support for getPartition(uuid)
- [ ] Document field name mappings (complete table)
- [ ] Test list attribute serialization (admin_email, time_servers, etc.)
- [ ] Verify BCM entity structure requirements
- [ ] Test concurrency control (revision field)
- [ ] Create research.md with decisions and rationale

**Phase 1 (Design)**:
- [ ] Create data-model.md with PartitionResourceModel struct
- [ ] Create contracts/cmpart-partition-api.json with API documentation
- [ ] Create quickstart.md with developer setup guide
- [ ] Update agent context with partition management patterns
- [ ] Review design artifacts for consistency with spec.md

**Phase 2 (Tasks - via /speckit.tasks)**:
- [ ] Generate tasks.md with RED-GREEN-REFACTOR breakdown
- [ ] Task dependencies identified and ordered
- [ ] Acceptance criteria defined for each task
- [ ] Parallel execution opportunities identified

**Phase 3 (Implementation - via /speckit.implement)**:
- [ ] RED: Write 8 failing acceptance tests
- [ ] GREEN: Minimal implementation (hardcoded responses)
- [ ] REFACTOR: Full BCM API integration
- [ ] Register resource in provider.go
- [ ] Create examples/resources/bcm_cmpart_partition/resource.tf
- [ ] Run make generate for documentation
- [ ] All tests passing (TF_ACC=1 go test -v ./internal/provider/)

### Glossary

**Partition**: BCM cluster organizational unit that groups nodes and defines configuration
**UUID**: Universal Unique Identifier assigned by BCM (stable resource identifier)
**BCM Entity**: API wrapper structure with baseType, childType, modified, revision fields
**Drift Detection**: Terraform's ability to detect configuration changes made outside of Terraform
**Idempotency**: Property where applying the same configuration multiple times produces the same result
**Args Parameter**: BCM JSON-RPC feature enabling parameterized API calls (e.g., getPartition(uuid))
**Eventual Consistency**: BCM API pattern where changes may take time to propagate
**Revision**: BCM concurrency control field for optimistic locking
**statecheck**: Modern Terraform testing package for type-safe state verification
**plancheck**: Modern Terraform testing package for plan verification (idempotency, drift)
