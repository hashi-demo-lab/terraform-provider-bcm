# Tasks: BCM Partition Resource (bcm_cmpart_partition)

**Branch**: `002-cmpart-partition`
**Input**: Design documents from `/workspace/specs/002-cmpart-partition/`
**Prerequisites**: spec.md (user stories), plan.md (implementation strategy)

**TDD Workflow**: RED-GREEN-REFACTOR with modern testing patterns (statecheck, plancheck)

**Organization**: Tasks are grouped by implementation phase following TDD principles and terraform-provider-design patterns.

## Format: `- [ ] [ID] [P?] [Story?] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: User story this task belongs to (US1, US2, US3, US4)
- Include exact file paths in descriptions

---

## Phase 0: Research & API Verification

**Purpose**: Validate BCM API methods, field mappings, and implementation assumptions before design

**Dependencies**: None - can start immediately

**Duration Estimate**: 2-3 hours (one-time research)

### Research Tasks

- [ ] T001 [P] Verify BCM CMPart API methods for partition management using sampleRest/ exploration scripts
  - **Goal**: Confirm exact API method signatures: addPartition, getPartition, updatePartition, removePartition
  - **Actions**:
    - Test `CallJSONRPC(ctx, "cmpart", "getPartitions")` to list all partitions
    - Test `CallJSONRPC(ctx, "cmpart", "getPartition", uuid)` to verify args parameter support
    - Test `CallJSONRPC(ctx, "cmpart", "addPartition", entity, false)` with sample entity
    - Test `CallJSONRPC(ctx, "cmpart", "updatePartition", entity, false)` for updates
    - Test `CallJSONRPC(ctx, "cmpart", "removePartition", uuid, false)` for deletion
    - Document request/response examples for each method
  - **Verification**: Can execute all 4 CRUD operations successfully via BCM API
  - **Output**: Document findings in `specs/002-cmpart-partition/research.md` under "Decision: API Method Selection"

- [ ] T002 [P] Research partition field name mappings (Terraform snake_case → BCM API camelCase)
  - **Goal**: Create complete mapping table for all 15 partition attributes
  - **Actions**:
    - Review data_source_cmpart_partitions.go (lines 54-82) for authoritative field list
    - Test API responses to verify actual field names returned by BCM
    - Document special cases: no_zero_conf → noZeroConf, admin_email → adminEmail
    - Test list attribute serialization (JSON arrays vs comma-separated strings)
    - Verify empty list handling (null vs empty array vs omitted field)
  - **Verification**: Complete field mapping table with all 15 attributes documented
  - **Output**: Add mapping table to research.md under "Decision: Field Name Mapping"

- [ ] T003 [P] Research BCM entity structure requirements for partition operations
  - **Goal**: Understand BCM API entity wrapper format for create/update calls
  - **Actions**:
    - Review resource_cmpart_softwareimage.go buildAPIEntity() pattern as reference
    - Test partition entity structure: baseType, childType, modified, to_be_removed, revision, uuid
    - Verify baseType value for partitions (likely "Partition")
    - Test childType polymorphism (if applicable to partitions)
    - Document required vs optional entity wrapper fields
    - Test entity creation/update with minimal vs complete field sets
  - **Verification**: Can construct valid BCM partition entity for all CRUD operations
  - **Output**: Document entity structure in research.md under "Decision: Entity Structure"

- [ ] T004 [P] Research concurrent operation safety and partition constraints
  - **Goal**: Verify partition operations are safe for parallel Terraform execution
  - **Actions**:
    - Test creating multiple partitions simultaneously via BCM API
    - Verify revision field usage for optimistic concurrency control
    - Test partition name uniqueness constraint (what happens with duplicates?)
    - Test deletion constraints (can partition be deleted if nodes assigned?)
    - Document any API rate limits or locking mechanisms
    - Verify if force parameter affects deletion behavior
  - **Verification**: Document concurrency safety guarantees and constraints
  - **Output**: Add findings to research.md under "Decision: Concurrency Safety"

**Phase 0 Checkpoint**: All API behaviors documented, field mappings verified, ready for Phase 1 design

---

## Phase 1: Design Artifacts & Contracts

**Purpose**: Generate data model, API contracts, quickstart guide, and agent context updates

**Dependencies**: Phase 0 complete (research.md exists with API verification)

**Duration Estimate**: 2-3 hours

### Design Tasks

- [ ] T005 [P] Create data model specification in specs/002-cmpart-partition/data-model.md
  - **Goal**: Define PartitionResourceModel Go struct and validation rules
  - **Actions**:
    - Document complete PartitionResourceModel struct with all 15 attributes
    - Define Terraform types for each field (types.String, types.Int64, types.Bool, types.List)
    - Document required vs optional vs computed fields
    - Define validation rules (name length, cluster_name non-empty, slave_digits range)
    - Document list attribute element types (admin_email, time_servers, etc. are List[String])
    - Document entity relationships (Partition → Nodes, Partition → ClusterConfig)
    - Define state transitions (NULL → EXISTS → NULL, drift detection flow)
  - **Verification**: data-model.md includes complete struct definition and all validation rules
  - **Template**: Follow plan.md Phase 1 Artifact 1.1 structure

- [ ] T006 [P] Create API contracts documentation in specs/002-cmpart-partition/contracts/cmpart-partition-api.json
  - **Goal**: Document BCM JSON-RPC API contracts in OpenAPI-style format
  - **Actions**:
    - Document all 4 CRUD methods with request/response examples from Phase 0 research
    - Include complete field_mappings section (snake_case → camelCase for all 15 attributes)
    - Document entity wrapper structure with example values
    - Include error response examples (duplicate name, deletion failures, validation errors)
    - Document args parameter usage for getPartition(uuid) direct lookup
    - Add notes on force parameter behavior for create/update/delete operations
  - **Verification**: contracts/cmpart-partition-api.json is valid JSON and documents all API operations
  - **Template**: Follow plan.md Phase 1 Artifact 1.2 structure

- [ ] T007 [P] Create developer quickstart guide in specs/002-cmpart-partition/quickstart.md
  - **Goal**: Provide onboarding guide for developers implementing partition resource
  - **Actions**:
    - Document environment setup (Go 1.24+, BCM credentials, Terraform CLI)
    - Provide TDD workflow steps (RED: tests, GREEN: minimal impl, REFACTOR: API integration)
    - Include example Terraform configuration for manual testing
    - Document file structure (resource_cmpart_partition.go, resource_cmpart_partition_test.go)
    - List reference implementations (data_source_cmpart_partitions.go, resource_cmpart_softwareimage.go)
    - Document common pitfalls (field mappings, Unknown values, list attributes, drift tests)
    - Include commands for running tests, building provider, generating docs
  - **Verification**: quickstart.md provides complete developer onboarding path
  - **Template**: Follow plan.md Phase 1 Artifact 1.3 structure

- [ ] T008 Update project documentation with partition resource patterns in /workspace/CLAUDE.md
  - **Goal**: Add bcm_cmpart_partition resource knowledge to agent context
  - **Actions**:
    - Add "BCM Partition Management (bcm_cmpart_partition)" section to CLAUDE.md
    - Document resource pattern (service: cmpart, methods: add/get/update/remove)
    - Add field mapping examples (cluster_name → clusterName, etc.)
    - Document list attribute handling pattern (types.List with ElementType: types.StringType)
    - Include entity structure example for partition resources
    - Reference key files (resource implementation, tests, schema reference)
  - **Verification**: CLAUDE.md includes partition management section with all patterns documented
  - **Dependencies**: T005, T006, T007 complete (design artifacts available)

**Phase 1 Checkpoint**: All design artifacts created, ready for task generation and implementation

---

## Phase 2: Acceptance Tests (RED - Write Failing Tests)

**Purpose**: Write comprehensive acceptance tests covering all CRUD operations, drift detection, and modern testing patterns

**Dependencies**: Phase 1 complete (design artifacts exist)

**Duration Estimate**: 6-8 hours

**TDD Phase**: 🔴 RED - Write failing tests BEFORE implementation

### Test Infrastructure

- [ ] T009 Create acceptance test file internal/provider/resource_cmpart_partition_test.go with test setup
  - **Goal**: Initialize test file with required imports, provider factories, and PreCheck function
  - **Actions**:
    - Add copyright header and package declaration
    - Import required packages: terraform-plugin-testing (resource, statecheck, plancheck, knownvalue, tfjsonpath, compare)
    - Import context, encoding/json, fmt, os, testing, time, regexp, strings
    - Define testAccProtoV6ProviderFactories (reference from existing tests)
    - Implement testAccPreCheck(t) to verify BCM environment variables
    - Add file-level documentation explaining test coverage
  - **Verification**: Test file compiles, PreCheck validates environment correctly
  - **Reference**: resource_cmpart_softwareimage_test.go (lines 1-50)

### CRUD Test Suite

- [ ] T010 [US1] Write TestAccCMPartPartition_Basic test for Create, Read, and Import operations
  - **Goal**: Test basic partition creation, state verification, and import functionality
  - **Actions**:
    - Generate unique partition name with generateUniqueTestName("test-partition")
    - Initialize ID consistency tracker: compareID := statecheck.CompareValue(compare.ValuesSame())
    - **Step 1**: Create partition with name and cluster_name
      - Use testAccPartitionConfigBasic(name, clusterName) config helper
      - Add ConfigStateChecks with statecheck.ExpectKnownValue for:
        - name (knownvalue.StringExact)
        - cluster_name (knownvalue.StringExact)
        - uuid (knownvalue.NotNull)
        - id (knownvalue.NotNull)
      - Track ID with compareID.AddStateValue
    - **Step 2**: Idempotency check after Create
      - Same config as Step 1
      - Add ConfigPlanChecks with plancheck.ExpectEmptyPlan()
    - **Step 3**: Import by UUID
      - ResourceName: "bcm_cmpart_partition.test"
      - ImportState: true, ImportStateVerify: true
      - Track ID consistency with compareID.AddStateValue
  - **Verification**: Test runs, fails with "resource type not found" (expected - resource not implemented yet)
  - **Reference**: plan.md lines 924-982 for complete test structure

- [ ] T011 [US1] Write TestAccCMPartPartition_Update test for configuration updates
  - **Goal**: Verify in-place updates work correctly without recreation
  - **Actions**:
    - Generate unique partition name
    - **Step 1**: Create with initial cluster_name, slave_name, notes values
      - Verify initial values with statecheck.ExpectKnownValue
    - **Step 2**: Update cluster_name
      - Change cluster_name value, verify new value in state
    - **Step 3**: Idempotency after cluster_name update
      - plancheck.ExpectEmptyPlan()
    - **Step 4**: Update slave_name and slave_digits
      - Verify new values in state
    - **Step 5**: Idempotency after slave_name update
      - plancheck.ExpectEmptyPlan()
    - **Step 6**: Update notes field
      - Verify notes update in state
  - **Verification**: Test fails (resource not implemented), covers all updatable fields
  - **Reference**: plan.md lines 986-1019

- [ ] T012 [P] [US2] Write TestAccCMPartPartition_NetworkSettings test for list attributes
  - **Goal**: Test list attribute configuration (admin_email, time_servers, search_domains, name_servers)
  - **Actions**:
    - Generate unique partition name
    - **Step 1**: Create with network configuration lists
      - admin_email: ["admin@example.com", "ops@example.com"]
      - time_servers: ["ntp1.example.com", "ntp2.example.com"]
      - name_servers: ["8.8.8.8", "8.8.4.4"]
      - search_domains: ["example.com"]
      - Verify list sizes with knownvalue.ListSizeExact(2), ListSizeExact(1), etc.
    - **Step 2**: Update network lists (add/remove entries)
      - Modify admin_email list, verify new size
    - **Step 3**: Empty network lists
      - Set lists to empty [], verify ListSizeExact(0)
  - **Verification**: Test fails, covers all list attributes
  - **Reference**: plan.md lines 1024-1051

- [ ] T013 [P] [US3] Write TestAccCMPartPartition_DriftDetection test for external modifications
  - **Goal**: Verify Terraform detects and corrects configuration drift from external BCM API changes
  - **Actions**:
    - Generate unique partition name
    - **Step 1**: Create partition with initial notes value
      - Verify notes = "Initial notes"
    - **Step 2**: Modify partition externally via BCM API (drift simulation)
      - PreConfig function:
        - Create BCM client with createTestBCMClient(t)
        - Get partition UUID with getResourceUUIDByName(t, "cmpart", "getPartition", name)
        - Fetch full partition data via CallJSONRPC
        - Modify notes field to "Modified externally"
        - Build BCM entity structure (baseType, childType, modified, revision, uuid + all fields)
        - Call updatePartition via BCM API
        - Sleep 2 seconds for eventual consistency
      - ConfigPlanChecks: plancheck.ExpectNonEmptyPlan() (drift detected)
    - **Step 3**: Terraform restores desired state
      - Same config as Step 1
      - Verify notes restored to "Initial notes"
  - **Verification**: Test fails, demonstrates drift detection pattern
  - **Reference**: plan.md lines 1056-1127, test_helpers.go field mapping documentation

- [ ] T014 [P] [US4] Write TestAccCMPartPartition_SlaveNaming test for node naming configuration
  - **Goal**: Test slave_name and slave_digits configuration (node naming conventions)
  - **Actions**:
    - Generate unique partition name
    - **Step 1**: Create with custom slave naming
      - slave_name = "compute"
      - slave_digits = 4
      - Verify both values in state with knownvalue.StringExact and knownvalue.Int64Exact
    - **Step 2**: Update slave_digits
      - Change slave_digits from 4 to 3
      - Verify new value in state
    - **Step 3**: Update slave_name
      - Change slave_name from "compute" to "gpu"
      - Verify new value in state
  - **Verification**: Test fails, covers node naming configuration

- [ ] T015 [P] Write TestAccCMPartPartition_IDConsistency test for ID stability tracking
  - **Goal**: Verify ID remains stable across Create, Import, and Update operations
  - **Actions**:
    - Generate unique partition name
    - Initialize compareID := statecheck.CompareValue(compare.ValuesSame())
    - **Step 1**: Create partition, capture ID
      - compareID.AddStateValue("bcm_cmpart_partition.test", tfjsonpath.New("id"))
    - **Step 2**: Import partition, verify ID unchanged
      - compareID.AddStateValue (should match Step 1)
    - **Step 3**: Update partition, verify ID unchanged
      - compareID.AddStateValue (should still match)
  - **Verification**: Test fails, demonstrates ID consistency tracking pattern
  - **Reference**: Modern testing patterns from CLAUDE.md

### CheckDestroy Function

- [ ] T016 Write testAccCheckPartitionDestroy function with exponential backoff verification
  - **Goal**: Verify partition deletion in CheckDestroy with robust eventual consistency handling
  - **Actions**:
    - Create function signature: func testAccCheckPartitionDestroy(s *terraform.State) error
    - Initialize BCM client with createTestBCMClient(&testing.T{})
    - Initialize error slice and resource counter
    - Iterate through s.RootModule().Resources
    - Filter for rs.Type == "bcm_cmpart_partition"
    - For each partition:
      - Extract UUID from rs.Primary.ID
      - Call verifyResourceDeleted(ctx, client, "cmpart", "getPartition", uuid, 4)
      - Collect detailed error messages if still exists
    - Return aggregated error with all failures
    - Log successful deletion count
  - **Verification**: Function compiles, provides detailed error messages
  - **Reference**: plan.md lines 1132-1168, test_helpers.go verifyResourceDeleted pattern

### Test Config Helpers

- [ ] T017 [P] Write testAccPartitionConfigBasic helper function for basic partition config
  - **Goal**: Generate Terraform configuration for basic partition tests
  - **Actions**:
    - Function signature: func testAccPartitionConfigBasic(name, clusterName string) string
    - Return HCL string with:
      - Provider block (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD from env, insecure_skip_verify=true)
      - Resource "bcm_cmpart_partition" "test" with name and cluster_name
    - Use fmt.Sprintf with %[1]q...%[5]q for proper quoting
  - **Verification**: Generated config is valid HCL
  - **Reference**: plan.md lines 1175-1195

- [ ] T018 [P] Write testAccPartitionConfigWithNotes helper for drift detection tests
  - **Goal**: Generate config with notes field for drift testing
  - **Actions**:
    - Function signature: func testAccPartitionConfigWithNotes(name, notes string) string
    - Include provider block + resource with name, cluster_name, notes fields
  - **Verification**: Generated config includes notes field

- [ ] T019 [P] Write testAccPartitionConfigNetworkSettings helper for list attribute tests
  - **Goal**: Generate config with network list attributes
  - **Actions**:
    - Function signature: func testAccPartitionConfigNetworkSettings(name string, adminEmails, timeServers []string) string
    - Include helper function quoteStrings([]string) []string to quote list elements
    - Convert Go slices to HCL list syntax: ["item1", "item2"]
    - Include provider block + resource with admin_email and time_servers lists
  - **Verification**: Generated config correctly formats list attributes
  - **Reference**: plan.md lines 1198-1234

- [ ] T020 [P] Write testAccPartitionConfigComplete helper for full attribute coverage
  - **Goal**: Generate config with all optional fields populated
  - **Actions**:
    - Function signature accepts all partition attributes
    - Include all fields: name, cluster_name, slave_name, slave_digits, relay_host, no_zero_conf
    - Include all list attributes: admin_email, time_servers, search_domains, name_servers
    - Include notes field
  - **Verification**: Config exercises all schema attributes

**Phase 2 Checkpoint**: All acceptance tests written and failing (expected), ready for Phase 3 implementation

---

## Phase 3: Resource Implementation (GREEN - Make Tests Pass)

**Purpose**: Implement bcm_cmpart_partition resource with minimal CRUD operations to pass tests

**Dependencies**: Phase 2 complete (all tests written and failing)

**Duration Estimate**: 8-10 hours

**TDD Phase**: 🟢 GREEN - Write minimal code to pass tests

### Resource Structure

- [ ] T021 Create resource file internal/provider/resource_cmpart_partition.go with basic structure
  - **Goal**: Initialize resource file with schema definition and CRUD method stubs
  - **Actions**:
    - Add copyright header and package declaration
    - Import required packages: context, encoding/json, fmt, terraform-plugin-framework (resource, schema, types, diag, path, validator)
    - Define CMPartPartitionResource struct with client *BCMClient field
    - Define PartitionResourceModel struct with all 15 attributes (matching data-model.md)
    - Implement interface checks: var _ resource.Resource = &CMPartPartitionResource{}
    - Implement interface checks: var _ resource.ResourceWithImportState = &CMPartPartitionResource{}
    - Implement NewCMPartPartitionResource() resource.Resource factory function
    - Implement Metadata() method returning "bcm_cmpart_partition" type name
  - **Verification**: File compiles, resource structure defined
  - **Reference**: resource_cmpart_softwareimage.go lines 1-88

- [ ] T022 Implement Schema() method with complete partition attribute definitions
  - **Goal**: Define Terraform schema for all 15 partition attributes
  - **Actions**:
    - Define schema.Schema with MarkdownDescription
    - **Identity attributes** (all Computed):
      - id: StringAttribute, Computed, description "Resource identifier (same as UUID)"
      - uuid: StringAttribute, Computed, description "BCM-assigned unique identifier"
      - name: StringAttribute, Required, description "Partition name (unique within BCM cluster)"
      - base_type: StringAttribute, Computed, description "BCM entity base type (Partition)"
      - child_type: StringAttribute, Computed, description "BCM entity polymorphic child type"
    - **Configuration attributes**:
      - cluster_name: StringAttribute, Required, description "Cluster display name"
      - slave_name: StringAttribute, Optional, description "Node naming prefix (default: node)"
      - slave_digits: Int64Attribute, Optional, description "Node numbering digits (default: 3)"
      - relay_host: StringAttribute, Optional, description "SMTP relay hostname"
      - no_zero_conf: BoolAttribute, Optional, description "Disable Zeroconf (default: false)"
    - **List attributes** (all Optional, ElementType: types.StringType):
      - admin_email: ListAttribute, description "Admin contact email addresses"
      - time_servers: ListAttribute, description "NTP time servers"
      - search_domains: ListAttribute, description "DNS search domains"
      - name_servers: ListAttribute, description "DNS resolver servers"
    - **Metadata attributes** (all Computed):
      - creation_time: Int64Attribute, description "Unix timestamp of creation"
      - revision: StringAttribute, description "Concurrency control version"
      - modified: BoolAttribute, description "Dirty flag indicating unsaved changes"
      - to_be_removed: BoolAttribute, description "Deletion pending flag"
      - notes: StringAttribute, Optional, description "Partition description/notes"
  - **Verification**: Schema compiles, matches PartitionResourceModel struct exactly
  - **Reference**: data_source_cmpart_partitions.go lines 90-180 for schema patterns

- [ ] T023 Implement Configure() method to receive BCM client from provider
  - **Goal**: Configure resource with authenticated BCM client
  - **Actions**:
    - Extract client from req.ProviderData
    - Type assert to *BCMClient
    - Handle nil provider data (unconfigured state)
    - Add error diagnostic if type assertion fails
    - Store client in r.client field
  - **Verification**: Resource receives client correctly during provider initialization
  - **Reference**: resource_cmpart_softwareimage.go Configure() method

### CRUD Implementation (Minimal - Make Tests Pass)

- [ ] T024 [US1] Implement Create() method for partition creation
  - **Goal**: Create partition via BCM API and populate state
  - **Actions**:
    - Extract plan data into PartitionResourceModel
    - Build BCM entity with buildAPIEntity() helper (to be implemented in T029)
    - Call r.client.CallJSONRPC(ctx, "cmpart", "addPartition", entity, false)
    - Parse response to extract UUID
    - Call readPartition() helper (to be implemented in T028) to populate computed fields
    - Set ID = UUID in state
    - Set all state values from model
    - Add error diagnostics for API failures
  - **Verification**: TestAccCMPartPartition_Basic Step 1 passes (Create)
  - **Reference**: resource_cmpart_softwareimage.go Create() method (lines 200-300)

- [ ] T025 [US1] Implement Read() method for state refresh and drift detection
  - **Goal**: Fetch current partition state from BCM API
  - **Actions**:
    - Extract UUID from state
    - Call r.client.CallJSONRPC(ctx, "cmpart", "getPartition", state.UUID.ValueString())
    - Handle 404/not found errors (resource deleted externally - remove from state)
    - Parse response with readPartition() helper
    - Map BCM API fields (camelCase) to Terraform attributes (snake_case)
    - Handle list attributes: unmarshal JSON arrays to types.List
    - Set all state values from fetched data
    - Add error diagnostics for API failures
  - **Verification**: TestAccCMPartPartition_Basic passes completely (Create, Read, Import)
  - **Reference**: resource_cmpart_softwareimage.go Read() method

- [ ] T026 [US1] Implement Update() method for in-place partition updates
  - **Goal**: Update partition configuration without recreation
  - **Actions**:
    - Extract plan data into PartitionResourceModel
    - Preserve UUID from state (don't change)
    - Build updated BCM entity with buildAPIEntity()
    - Call r.client.CallJSONRPC(ctx, "cmpart", "updatePartition", entity, false)
    - Call readPartition() to refresh state with latest BCM data
    - Set all state values from model
    - Add error diagnostics for API failures
  - **Verification**: TestAccCMPartPartition_Update passes (all update scenarios)
  - **Reference**: resource_cmpart_softwareimage.go Update() method

- [ ] T027 [US1] Implement Delete() method for partition removal
  - **Goal**: Delete partition from BCM cluster
  - **Actions**:
    - Extract UUID from state
    - Call r.client.CallJSONRPC(ctx, "cmpart", "removePartition", state.UUID.ValueString(), false)
    - Handle deletion errors (partition has nodes assigned, etc.)
    - Provide actionable error message if deletion fails
    - No need to remove from state (Terraform handles this automatically)
  - **Verification**: testAccCheckPartitionDestroy passes (partition deleted)
  - **Reference**: resource_cmpart_softwareimage.go Delete() method

- [ ] T028 [US1] Implement ImportState() method for importing existing partitions
  - **Goal**: Import existing BCM partition into Terraform state by UUID
  - **Actions**:
    - Use resource.ImportStatePassthroughID to set ID from import identifier
    - Terraform will automatically call Read() to populate state
  - **Verification**: TestAccCMPartPartition_Basic Step 3 passes (Import)
  - **Reference**: resource_cmpart_softwareimage.go ImportState() implementation

### Helper Functions

- [ ] T029 [P] Implement readPartition() helper to fetch and map BCM API response to model
  - **Goal**: Centralize partition data fetching and field mapping logic
  - **Actions**:
    - Function signature: func (r *CMPartPartitionResource) readPartition(ctx context.Context, model *PartitionResourceModel, diags *diag.Diagnostics)
    - Call BCM API: r.client.CallJSONRPC(ctx, "cmpart", "getPartition", model.UUID.ValueString())
    - Parse JSON response into map[string]interface{}
    - Map identity fields: uuid, name, baseType, childType
    - Map config fields with camelCase conversion:
      - clusterName → cluster_name (types.String)
      - slaveName → slave_name (types.String)
      - slaveDigits → slave_digits (types.Int64)
      - relayHost → relay_host (types.String)
      - noZeroConf → no_zero_conf (types.Bool)
    - Map list attributes with JSON array unmarshaling:
      - adminEmail → admin_email (types.List)
      - timeServers → time_servers (types.List)
      - searchDomains → search_domains (types.List)
      - nameServers → name_servers (types.List)
    - Map metadata fields: creationTime, revision, modified, to_be_removed, notes
    - Handle null values gracefully (use types.StringNull(), types.ListNull(), etc.)
    - Add diagnostics for parsing errors
  - **Verification**: Read() and Create() methods work correctly with this helper
  - **Reference**: data_source_cmpart_partitions.go field mapping pattern

- [ ] T030 [P] Implement buildAPIEntity() helper to construct BCM entity from Terraform model
  - **Goal**: Build BCM API entity structure with proper field name mapping
  - **Actions**:
    - Function signature: func buildAPIEntity(model *PartitionResourceModel) map[string]interface{}
    - Create entity map with required BCM wrapper fields:
      - baseType: "Partition"
      - childType: "" (empty for base partitions)
      - modified: true (indicates changes)
      - to_be_removed: false
      - revision: model.Revision.ValueString() (preserve for optimistic locking)
      - uuid: model.UUID.ValueString() (for updates, empty for creates)
    - Add partition fields with snake_case → camelCase conversion:
      - name → name (no conversion)
      - cluster_name → clusterName
      - slave_name → slaveName
      - slave_digits → slaveDigits (convert types.Int64 to int)
      - relay_host → relayHost
      - no_zero_conf → noZeroConf
    - Convert list attributes to JSON arrays:
      - admin_email (types.List) → adminEmail ([]string)
      - time_servers → timeServers
      - search_domains → searchDomains
      - name_servers → nameServers
    - Add notes field (notes → notes)
    - Handle null/unknown values (omit from entity if null)
    - Return complete entity map
  - **Verification**: Create() and Update() methods work correctly with this helper
  - **Reference**: resource_cmpart_softwareimage.go buildAPIEntity pattern

- [ ] T031 [P] Implement mapListAttribute() helper to convert types.List to []string for BCM API
  - **Goal**: Convert Terraform list attributes to Go string slices for JSON serialization
  - **Actions**:
    - Function signature: func mapListAttribute(list types.List) []string
    - Handle null list (return nil or empty slice)
    - Handle empty list (return empty slice)
    - Extract list elements using list.Elements()
    - Convert each element from types.String to string value
    - Return []string suitable for JSON marshaling
  - **Verification**: List attributes serialize correctly in buildAPIEntity()
  - **Dependencies**: Used by T030

- [ ] T032 [P] Implement unmarshalListAttribute() helper to convert BCM API arrays to types.List
  - **Goal**: Convert JSON arrays from BCM API to Terraform types.List
  - **Actions**:
    - Function signature: func unmarshalListAttribute(data interface{}) types.List
    - Handle nil/null data (return types.ListNull)
    - Type assert to []interface{}
    - Convert each element to types.String
    - Build types.List with ElementType: types.StringType
    - Return populated types.List
  - **Verification**: List attributes deserialize correctly in readPartition()
  - **Dependencies**: Used by T029

### Provider Registration

- [ ] T033 Register bcm_cmpart_partition resource in internal/provider/provider.go
  - **Goal**: Make resource available to Terraform
  - **Actions**:
    - Open internal/provider/provider.go
    - Locate Resources() method
    - Add NewCMPartPartitionResource to returned slice
    - Ensure proper alphabetical ordering (after bcm_cmnet_network, before bcm_cmuser_user if it exists)
  - **Verification**: Provider compiles, terraform plan recognizes bcm_cmpart_partition resource type
  - **Reference**: provider.go Resources() method

**Phase 3 Checkpoint**: All acceptance tests passing (GREEN phase complete), ready for refactoring

---

## Phase 4: Refactoring & Validation (REFACTOR + Polish)

**Purpose**: Improve code quality, add validation, enhance error messages, verify test robustness

**Dependencies**: Phase 3 complete (all tests passing)

**Duration Estimate**: 3-4 hours

**TDD Phase**: 🔄 REFACTOR - Improve code while keeping tests green

### Code Quality Improvements

- [ ] T034 [P] Add schema validators for partition attributes
  - **Goal**: Add validation rules to prevent invalid configurations
  - **Actions**:
    - Import terraform-plugin-framework-validators/stringvalidator, int64validator
    - Add validators to name attribute:
      - stringvalidator.LengthBetween(1, 255) - prevent empty names
      - (Optional) stringvalidator.RegexMatches for valid partition name pattern
    - Add validator to cluster_name:
      - stringvalidator.LengthAtLeast(1) - prevent empty cluster names
    - Add validator to slave_digits:
      - int64validator.Between(1, 10) - reasonable range for node numbering
    - Consider adding validator for relay_host (URL format if needed)
  - **Verification**: Run tests, all still pass, invalid configs rejected with clear errors
  - **Reference**: resource_cmpart_softwareimage.go validator patterns

- [ ] T035 [P] Enhance error messages in CRUD operations
  - **Goal**: Provide actionable error messages for users
  - **Actions**:
    - In Create(): Add context about duplicate partition names
      - "Partition name 'X' already exists. Choose unique name or import existing partition."
    - In Delete(): Add context about deletion constraints
      - "Cannot delete partition 'X' (uuid: Y) - partition has N active nodes. Remove nodes first."
    - In Read(): Add context about missing resources
      - "Partition 'X' (uuid: Y) not found in BCM. Resource may have been deleted externally."
    - In Update(): Add context about conflicts
      - "Update failed for partition 'X' - revision conflict. Resource modified by another process."
    - Parse BCM API error responses to extract meaningful error details
  - **Verification**: Error messages are clear and actionable (manual testing)

- [ ] T036 [P] Add logging with tflog for debugging
  - **Goal**: Add structured logging for troubleshooting
  - **Actions**:
    - Import github.com/hashicorp/terraform-plugin-log/tflog
    - Add tflog.Debug in Create(): "Creating partition: name=%s, cluster_name=%s"
    - Add tflog.Debug in Read(): "Reading partition: uuid=%s"
    - Add tflog.Debug in Update(): "Updating partition: uuid=%s, changed_fields=%v"
    - Add tflog.Debug in Delete(): "Deleting partition: uuid=%s"
    - Add tflog.Trace for BCM API calls with request/response bodies
    - Add tflog.Warn for drift detection: "Partition drift detected: field=%s, terraform=%v, bcm=%v"
  - **Verification**: Run tests with TF_LOG=DEBUG, verify useful log output

- [ ] T037 Refactor common field mapping logic into reusable functions
  - **Goal**: Reduce code duplication in readPartition() and buildAPIEntity()
  - **Actions**:
    - Extract getStringValue(data, key) helper (null-safe string extraction)
    - Extract getBoolValue(data, key) helper (null-safe bool extraction)
    - Extract getInt64Value(data, key) helper (null-safe int64 extraction with float64 handling)
    - Use these helpers consistently in readPartition()
  - **Verification**: Tests still pass, code is more readable
  - **Reference**: data_source_cmpart_partitions.go lines 399-431 for helper function patterns

### Test Enhancements

- [ ] T038 [P] Add TestAccCMPartPartition_ValidationErrors test for schema validation
  - **Goal**: Verify schema validators reject invalid configurations
  - **Actions**:
    - Test empty partition name (should fail validation)
    - Test empty cluster_name (should fail validation)
    - Test invalid slave_digits value (e.g., 0 or 100) (should fail validation)
    - Use ExpectError with regexp.MustCompile to verify error messages
  - **Verification**: Test passes, validators working correctly
  - **Reference**: Modern testing patterns from CLAUDE.md

- [ ] T039 [P] Enhance testAccCheckPartitionDestroy with detailed logging
  - **Goal**: Improve CheckDestroy error messages for debugging
  - **Actions**:
    - Add t.Logf for each partition being verified
    - Include resource count in final success message
    - Format error messages with resource type, ID, and retry details
    - Add timing information (total verification time)
  - **Verification**: CheckDestroy provides clear debugging output

- [ ] T040 Verify all tests pass with parallel execution
  - **Goal**: Ensure tests are safe for parallel execution
  - **Actions**:
    - Run: TF_ACC=1 go test -v -parallel=4 -timeout 120m ./internal/provider/ -run TestAccCMPartPartition
    - Verify no test conflicts or race conditions
    - Verify all tests still pass with parallel execution
    - Check that generateUniqueTestName prevents name collisions
  - **Verification**: All tests pass with -parallel=4 flag

**Phase 4 Checkpoint**: Code refactored, validation added, all tests still passing

---

## Phase 5: Documentation & Examples

**Purpose**: Generate documentation, create examples, verify completeness

**Dependencies**: Phase 4 complete (implementation refactored and validated)

**Duration Estimate**: 2-3 hours

### Example Configuration

- [ ] T041 [P] Create example Terraform configuration in examples/resources/bcm_cmpart_partition/resource.tf
  - **Goal**: Provide working example for documentation and manual testing
  - **Actions**:
    - Create directory: examples/resources/bcm_cmpart_partition/
    - Create resource.tf with:
      - Terraform block requiring bcm provider
      - Provider block (use environment variables for credentials)
      - Resource "bcm_cmpart_partition" "engineering" with:
        - name = "engineering"
        - cluster_name = "HPC Production Cluster"
        - slave_name = "compute"
        - slave_digits = 4
        - admin_email = ["admin@example.com", "ops@example.com"]
        - time_servers = ["ntp1.example.com", "ntp2.example.com"]
        - name_servers = ["8.8.8.8", "8.8.4.4"]
        - search_domains = ["example.com", "corp.example.com"]
        - relay_host = "smtp.example.com"
        - no_zero_conf = false
        - notes = "Engineering team partition for GPU workloads"
      - Output blocks for uuid and id
  - **Verification**: Example is valid HCL, runs successfully with terraform apply
  - **Reference**: plan.md lines 706-723 for example structure

- [ ] T042 [P] Create minimal example in examples/resources/bcm_cmpart_partition/minimal.tf
  - **Goal**: Show minimal required configuration
  - **Actions**:
    - Create minimal.tf with only required fields:
      - name = "minimal-partition"
      - cluster_name = "Test Cluster"
    - Add comment explaining this is minimal config
  - **Verification**: Minimal example works correctly

- [ ] T043 [P] Create import example in examples/resources/bcm_cmpart_partition/import.sh
  - **Goal**: Document import procedure
  - **Actions**:
    - Create shell script demonstrating import command
    - Include steps: 1) Get UUID from BCM, 2) terraform import command, 3) Verify state
    - Add comments explaining each step
  - **Verification**: Import instructions are clear and accurate

### Documentation Generation

- [ ] T044 Generate provider documentation with make generate
  - **Goal**: Auto-generate docs/resources/bcm_cmpart_partition.md
  - **Actions**:
    - Run: make generate
    - Verify docs/resources/bcm_cmpart_partition.md is created
    - Review generated documentation for completeness
    - Verify all attributes are documented with descriptions
    - Verify examples are included in documentation
    - Check import syntax is documented
  - **Verification**: Generated documentation is complete and accurate
  - **Dependencies**: T041, T042, T043 (examples must exist)

- [ ] T045 [P] Update CHANGELOG.md with partition resource addition
  - **Goal**: Document new feature in changelog
  - **Actions**:
    - Add entry under "Unreleased" section (or next version)
    - Format: "**NEW RESOURCE**: `bcm_cmpart_partition` - Manage BCM cluster partitions"
    - Include brief description of capabilities
  - **Verification**: Changelog entry is clear and follows project format

### Final Validation

- [ ] T046 Run full acceptance test suite to verify no regressions
  - **Goal**: Ensure partition resource doesn't break existing functionality
  - **Actions**:
    - Run: TF_ACC=1 go test -v -timeout 120m ./internal/provider/
    - Verify all tests pass (not just partition tests)
    - Check test execution time is reasonable
    - Verify no unexpected errors or warnings
  - **Verification**: 100% test pass rate across entire provider

- [ ] T047 Run pre-commit hooks and code quality checks
  - **Goal**: Ensure code meets quality standards
  - **Actions**:
    - Run: make fmt (format code)
    - Run: make lint (golangci-lint)
    - Run: pre-commit run --all-files (if pre-commit configured)
    - Fix any issues identified
  - **Verification**: All quality checks pass

- [ ] T048 Manually test examples with real BCM cluster
  - **Goal**: Verify examples work in realistic scenario
  - **Actions**:
    - Set BCM environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
    - cd examples/resources/bcm_cmpart_partition
    - Run: terraform init
    - Run: terraform plan
    - Run: terraform apply (creates partition)
    - Verify partition in BCM UI/API
    - Run: terraform destroy (removes partition)
    - Verify deletion in BCM
  - **Verification**: Examples work end-to-end with real BCM cluster

**Phase 5 Checkpoint**: Documentation generated, examples tested, ready for merge

---

## Dependencies & Execution Order

### Phase Dependencies

1. **Phase 0 (Research)**: No dependencies - start immediately
2. **Phase 1 (Design)**: Requires Phase 0 complete (research.md exists)
3. **Phase 2 (Tests - RED)**: Requires Phase 1 complete (design artifacts exist)
4. **Phase 3 (Implementation - GREEN)**: Requires Phase 2 complete (tests written and failing)
5. **Phase 4 (Refactor)**: Requires Phase 3 complete (tests passing)
6. **Phase 5 (Documentation)**: Requires Phase 4 complete (implementation finalized)

### Task Dependencies Within Phases

**Phase 0 (Parallel)**:
- T001, T002, T003, T004 can all run in parallel (independent research)

**Phase 1 (Mostly Parallel)**:
- T005, T006, T007 can run in parallel
- T008 depends on T005, T006, T007 (needs design artifacts to reference)

**Phase 2 (Parallel Test Writing)**:
- T009 must complete first (test file structure)
- T010-T015 can run in parallel (different test functions)
- T016-T020 can run in parallel (test helpers)

**Phase 3 (Sequential Resource Implementation)**:
- T021 must complete first (resource structure)
- T022 must complete after T021 (schema requires structure)
- T023 can run in parallel with T022
- T024-T028 depend on T021-T023 (CRUD methods require schema + structure)
- T029-T032 can run in parallel (helper functions)
- T024-T028 use helpers from T029-T032 (may need iteration)
- T033 must be last (registration requires complete implementation)

**Phase 4 (Parallel Refactoring)**:
- T034-T039 can run in parallel (independent improvements)
- T040 must be last (final parallel execution verification)

**Phase 5 (Parallel Documentation)**:
- T041-T043 can run in parallel (examples)
- T044 depends on T041-T043 (needs examples for doc generation)
- T045 can run in parallel with others
- T046-T048 should run sequentially at end (final validation)

### Parallel Opportunities

**Maximum Parallelization Points**:
- Phase 0: 4 tasks in parallel (all research tasks)
- Phase 1: 3 tasks in parallel (T005, T006, T007)
- Phase 2: 6 tests + 5 helpers = 11 tasks in parallel after T009
- Phase 3: T029-T032 (4 helpers) in parallel
- Phase 4: T034-T039 (6 improvements) in parallel
- Phase 5: T041-T043 (3 examples) in parallel

---

## Implementation Strategy

### TDD Red-Green-Refactor Cycle

**RED (Phase 2)**:
1. Write all 8 acceptance tests (T010-T015, T038)
2. Write CheckDestroy function (T016)
3. Write test config helpers (T017-T020)
4. Run tests: `TF_ACC=1 go test -v ./internal/provider/ -run TestAccCMPartPartition`
5. **Verify all tests FAIL** (resource not implemented yet)

**GREEN (Phase 3)**:
1. Create resource structure (T021)
2. Define schema (T022)
3. Implement minimal CRUD (T024-T028)
4. Implement helper functions (T029-T032)
5. Register resource (T033)
6. Run tests repeatedly until ALL PASS
7. **Verify all tests PASS** (minimal implementation complete)

**REFACTOR (Phase 4)**:
1. Add validators (T034)
2. Enhance error messages (T035)
3. Add logging (T036)
4. Refactor common logic (T037)
5. Add validation tests (T038)
6. Enhance test debugging (T039)
7. Verify parallel execution (T040)
8. **Verify all tests STILL PASS** after refactoring

### MVP Delivery Strategy

**Minimum Viable Product** (User Story 1 Only):
- Complete Phase 0: Research (required for all work)
- Complete Phase 1: Design (required for all work)
- Complete Phase 2: All tests (TDD requires tests first)
- Complete Phase 3: Implementation (basic CRUD)
- **STOP and VALIDATE**: All tests passing?
- Complete Phase 5 (T041, T044, T048): Basic docs and examples
- **Deploy/Demo**: Basic partition management working

**Incremental Enhancement**:
- Already complete: US1 (basic CRUD) ✅
- Add: US2 validation (Phase 4 T034, T038) ✅
- Add: US3 drift detection (already tested in T013) ✅
- Add: US4 advanced naming (already tested in T014) ✅
- Complete Phase 4: Full refactoring
- Complete Phase 5: Complete documentation

### Testing Strategy

**Test Execution Order** (during development):
1. `TestAccCMPartPartition_Basic` - Core CRUD functionality
2. `TestAccCMPartPartition_Update` - In-place updates
3. `TestAccCMPartPartition_NetworkSettings` - List attributes
4. `TestAccCMPartPartition_SlaveNaming` - Node naming config
5. `TestAccCMPartPartition_IDConsistency` - ID stability
6. `TestAccCMPartPartition_DriftDetection` - External changes
7. `TestAccCMPartPartition_ValidationErrors` - Schema validation

**Run Individual Test** (fast iteration):
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition_Basic
```

**Run All Partition Tests**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition
```

**Run Full Test Suite** (before merge):
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/
```

---

## Success Metrics

### Implementation Completion Criteria

| Metric | Target | Validation Method |
|--------|--------|-------------------|
| Acceptance Test Pass Rate | 100% | All 8 tests passing in Phase 3 |
| Test Coverage | All CRUD operations | Create, Read, Update, Delete, Import, Drift tested |
| Import Verification | 0 attribute mismatches | ImportStateVerify: true passes |
| Documentation Generated | Complete | docs/resources/bcm_cmpart_partition.md exists and accurate |
| Code Quality | Passing | make fmt, make lint pass |
| Examples Working | Manual validation | Examples run successfully against BCM cluster |
| No Regressions | 100% | Full provider test suite passes |

### Test Execution Performance

| Operation | Target | Measurement |
|-----------|--------|-------------|
| Single Test | <2 minutes | TestAccCMPartPartition_Basic runtime |
| Full Test Suite | <15 minutes | All 8 partition tests |
| Parallel Execution | No failures | Tests pass with -parallel=4 |

### Phase Completion Checklist

**Phase 0 Complete**:
- [ ] research.md exists with all 4 research tasks documented
- [ ] API methods verified (addPartition, getPartition, updatePartition, removePartition)
- [ ] Field mappings documented (15 attributes)
- [ ] Entity structure understood

**Phase 1 Complete**:
- [ ] data-model.md exists with PartitionResourceModel struct
- [ ] contracts/cmpart-partition-api.json exists with API documentation
- [ ] quickstart.md exists with developer guide
- [ ] CLAUDE.md updated with partition patterns

**Phase 2 Complete**:
- [ ] resource_cmpart_partition_test.go exists with 8 test functions
- [ ] All tests written and FAILING (expected)
- [ ] CheckDestroy function implemented
- [ ] Test config helpers implemented

**Phase 3 Complete**:
- [ ] resource_cmpart_partition.go exists with full CRUD
- [ ] All tests PASSING
- [ ] Resource registered in provider.go
- [ ] Helper functions working correctly

**Phase 4 Complete**:
- [ ] Validators added to schema
- [ ] Error messages enhanced
- [ ] Logging added
- [ ] Code refactored, tests STILL PASSING
- [ ] Parallel execution verified

**Phase 5 Complete**:
- [ ] Examples created and tested
- [ ] Documentation generated
- [ ] Changelog updated
- [ ] Full test suite passing
- [ ] Ready for pull request

---

## Risk Mitigation

### Known Risks

1. **BCM API Method Names Unknown** (Mitigated by Phase 0 Task T001)
   - Verify exact method names before implementation
   - Document in research.md

2. **Field Name Mapping Errors** (Mitigated by Phase 0 Task T002)
   - Complete mapping table before implementation
   - Test with real API responses

3. **List Attribute Serialization** (Mitigated by Phase 3 Task T031-T032)
   - Reference working patterns from data_source_cmpart_partitions.go
   - Test list attributes thoroughly in T012

4. **Drift Detection Test Complexity** (Mitigated by Phase 2 Task T013)
   - Follow proven pattern from test_helpers.go
   - Wait 2 seconds for BCM eventual consistency

5. **Partition Deletion Constraints** (Mitigated by Phase 4 Task T035)
   - Add clear error messages for deletion failures
   - Document prerequisites in error text

---

## Notes

- **[P] marker**: Tasks marked [P] can run in parallel with other [P] tasks in same phase
- **[US#] marker**: Maps task to user story for traceability
- **File paths**: All paths are absolute from repository root (/workspace/)
- **TDD discipline**: Never implement before tests are written and failing
- **Commit strategy**: Commit after each phase completion (not individual tasks)
- **Testing**: Always run tests after changes to verify no regressions
- **Environment**: BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD must be set for acceptance tests
- **Reference files**: Always check reference implementations before writing new code

---

## Quick Reference Commands

```bash
# Run specific test during development
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition_Basic

# Run all partition tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMPartPartition

# Run full test suite (before merge)
TF_ACC=1 go test -v -timeout 120m ./internal/provider/

# Run tests in parallel
TF_ACC=1 go test -v -parallel=4 -timeout 120m ./internal/provider/

# Format code
make fmt

# Run linter
make lint

# Generate documentation
make generate

# Build and install provider
make install

# Run example
cd examples/resources/bcm_cmpart_partition && terraform init && terraform plan
```

---

**Total Tasks**: 48 tasks across 6 phases (0=Research, 1=Design, 2=Tests, 3=Implementation, 4=Refactor, 5=Documentation)

**Estimated Duration**: 20-25 hours total (sequential execution) | 12-15 hours (maximum parallelization)

**Critical Path**: Phase 0 → Phase 1 → Phase 2 → Phase 3 (T021-T033 sequential) → Phase 4 → Phase 5
