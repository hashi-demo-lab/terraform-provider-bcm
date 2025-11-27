# Implementation Plan: BCM CMPart Entity Info Data Source

**Branch**: `095-cmpart-entity-info` | **Date**: 2025-11-27 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/specs/095-cmpart-entity-info/spec.md`

## Summary

Implement a read-only data source `bcm_cmpart_entity_info` that queries the BCM API `cmpart.getBasicEntityInformation` endpoint to retrieve entity metadata (name, type, UUID) for discovery purposes. The data source supports optional client-side filtering by entity type (case-sensitive exact match) and name pattern (case-insensitive glob with `*` and `?` wildcards). This enables Terraform users to discover BCM entities and obtain UUIDs for cross-resource references.

## Technical Context

**Language/Version**: Go 1.24+
**Primary Dependencies**: Terraform Plugin Framework v1.16.1, terraform-plugin-testing
**Storage**: N/A (read-only data source, no state storage)
**Testing**: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/` (acceptance tests)
**Target Platform**: Linux server (BCM headnode), cross-platform Terraform provider
**Project Type**: Single project (Terraform provider)
**Performance Goals**: Query completion under 10 seconds for 500+ entities
**Constraints**: API returns all entities in single response (no server-side pagination), client-side filtering required
**Scale/Scope**: Support clusters with 500-1000+ entities across 60+ entity types

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

| Principle | Status | Notes |
|-----------|--------|-------|
| TDD Parallel Execution | PASS | Acceptance tests written first (RED), minimal implementation (GREEN), refactor |
| Read-only data source | PASS | No write operations, no state modification |
| Client-side filtering | PASS | Follows existing pattern from `bcm_cmpart_softwareimages` |
| Error handling | PASS | Use established error message patterns |
| Plugin Framework v1.16.1 | PASS | Uses framework datasource interface |

## Project Structure

### Documentation (this feature)

```text
specs/095-cmpart-entity-info/
├── plan.md              # This file (/speckit.plan command output)
├── research.md          # Phase 0 output (/speckit.plan command)
├── data-model.md        # Phase 1 output (/speckit.plan command)
├── quickstart.md        # Phase 1 output (/speckit.plan command)
├── contracts/           # Phase 1 output (/speckit.plan command)
└── tasks.md             # Phase 2 output (/speckit.tasks command)
```

### Source Code (repository root)

```text
internal/provider/
├── data_source_cmpart_entity_info.go       # Data source implementation
└── data_source_cmpart_entity_info_test.go  # Acceptance tests

examples/data-sources/bcm_cmpart_entity_info/
└── data-source.tf                          # Example configurations

docs/data-sources/                          # Generated documentation (make generate)
└── cmpart_entity_info.md                   # Auto-generated from schema
```

**Structure Decision**: Single project structure following existing provider conventions. New files integrate into existing `internal/provider/` directory alongside other data sources.

## Architecture Decisions

### AD-001: API Integration Pattern

**Decision**: Use `cmpart.getBasicEntityInformation` with no arguments (returns all entities), apply filtering client-side.

**Rationale**:
- BCM API does not support server-side filtering for this endpoint
- Consistent with existing data sources (`bcm_cmpart_softwareimages`, `bcm_cmdevice_categories`)
- Single API call minimizes network round-trips

**Alternatives Rejected**:
- Server-side filtering: Not supported by BCM API
- Multiple filtered calls: Would require N calls for N types, inefficient

### AD-002: Field Mapping Strategy

**Decision**: Map API `resolveName` field to Terraform `name` attribute for clarity.

**Rationale**:
- `resolveName` is BCM internal terminology
- `name` is intuitive for Terraform users
- Consistent with Terraform naming conventions

**Implementation**:
```go
// API response field -> Terraform attribute
"resolveName" -> "name"
"type"        -> "type"
"uuid"        -> "uuid"
```

### AD-003: Filter Implementation

**Decision**: Support two optional filters with AND logic:
- `type`: Case-sensitive exact match (matches API behavior)
- `name_pattern`: Case-insensitive glob pattern (`*`, `?` wildcards)

**Rationale**:
- Type filtering is most common use case (find all entities of a type)
- Name pattern filtering supports discovery when exact name unknown
- Case-insensitive name matching improves usability
- Glob patterns are familiar to Terraform users (simpler than regex)

**Implementation**: Use Go `filepath.Match` for glob pattern matching with case normalization.

### AD-004: Schema Design

**Decision**: Use top-level filter attributes (not nested block) for simplicity.

**Rationale**:
- Only two filter parameters
- Flat schema is simpler for users
- Consistent with simple filtering use case

**Schema**:
```hcl
data "bcm_cmpart_entity_info" "example" {
  type         = "SoftwareImage"    # Optional
  name_pattern = "default*"         # Optional
}
```

### AD-005: ID Attribute Strategy

**Decision**: Generate deterministic ID from filter parameters for state management.

**Rationale**:
- Data sources require unique `id` for Terraform state
- Hash of filter values ensures consistency across runs
- Format: `cmpart-entity-info:{type}:{name_pattern}` or `cmpart-entity-info:all` when no filters

## File Structure

### New Files to Create

| File | Purpose | Lines (est.) |
|------|---------|--------------|
| `internal/provider/data_source_cmpart_entity_info.go` | Data source implementation | 250-300 |
| `internal/provider/data_source_cmpart_entity_info_test.go` | Acceptance tests | 200-250 |
| `examples/data-sources/bcm_cmpart_entity_info/data-source.tf` | Usage examples | 60-80 |

### Existing Files to Modify

| File | Change | Lines (est.) |
|------|--------|--------------|
| `internal/provider/provider.go` | Add `NewCMPartEntityInfoDataSource` to `DataSources()` | +1 |

## Implementation Approach

### Phase 1: TDD RED Phase - Write Failing Tests

1. Create `data_source_cmpart_entity_info_test.go` with acceptance tests:
   - `TestAccCMPartEntityInfoDataSource_Basic` - No filters, returns all entities
   - `TestAccCMPartEntityInfoDataSource_FilterByType` - Type filter only
   - `TestAccCMPartEntityInfoDataSource_FilterByNamePattern` - Name pattern only
   - `TestAccCMPartEntityInfoDataSource_CombinedFilters` - Both filters (AND logic)
   - `TestAccCMPartEntityInfoDataSource_EmptyResult` - Non-matching filter
   - `TestAccCMPartEntityInfoDataSource_InvalidCredentials` - Error handling

2. All tests should fail initially (RED phase)

### Phase 2: TDD GREEN Phase - Minimal Implementation

1. Create `data_source_cmpart_entity_info.go`:
   - Define `CMPartEntityInfoDataSource` struct
   - Implement `datasource.DataSource` interface
   - Implement `datasource.DataSourceWithConfigure` interface
   - Define schema with `type`, `name_pattern`, `id`, and `entities` attributes
   - Implement `Read` method to call BCM API and apply filters

2. Register data source in `provider.go`

3. Tests should pass (GREEN phase)

### Phase 3: TDD REFACTOR Phase - Improve Quality

1. Extract filter matching to separate function for testability
2. Add comprehensive logging with tflog
3. Ensure error messages follow `error_messages.go` patterns
4. Add documentation comments
5. Create example configurations

### Implementation Details

#### Data Source Schema

```go
type CMPartEntityInfoDataSourceModel struct {
    ID          types.String          `tfsdk:"id"`
    Type        types.String          `tfsdk:"type"`
    NamePattern types.String          `tfsdk:"name_pattern"`
    Entities    []EntityInfoModel     `tfsdk:"entities"`
}

type EntityInfoModel struct {
    Name types.String `tfsdk:"name"`
    Type types.String `tfsdk:"type"`
    UUID types.String `tfsdk:"uuid"`
}
```

#### API Call Pattern

```go
// Call BCM API (no arguments)
body, err := d.client.CallJSONRPC(ctx, "cmpart", "getBasicEntityInformation")

// Parse response - array of objects with resolveName, type, uuid
var apiResponse []map[string]interface{}
json.Unmarshal(body, &apiResponse)
```

#### Filter Implementation

```go
func matchesEntityFilter(entity EntityInfoModel, typeFilter, namePattern types.String) bool {
    // Type filter: case-sensitive exact match
    if !typeFilter.IsNull() && !typeFilter.IsUnknown() {
        if entity.Type.ValueString() != typeFilter.ValueString() {
            return false
        }
    }

    // Name pattern: case-insensitive glob match
    if !namePattern.IsNull() && !namePattern.IsUnknown() {
        pattern := strings.ToLower(namePattern.ValueString())
        name := strings.ToLower(entity.Name.ValueString())
        matched, _ := filepath.Match(pattern, name)
        if !matched {
            return false
        }
    }

    return true
}
```

## Testing Strategy

### Acceptance Tests

| Test | Purpose | Filter Config |
|------|---------|---------------|
| Basic | Verify API call works, returns entities | None |
| FilterByType | Verify type filtering | `type = "SoftwareImage"` |
| FilterByNamePattern | Verify glob pattern matching | `name_pattern = "default*"` |
| CombinedFilters | Verify AND logic | Both filters |
| EmptyResult | Verify empty list handling | `type = "NonExistentType"` |
| InvalidCredentials | Verify error handling | Invalid provider config |

### Test Configuration Pattern

```go
func testAccCMPartEntityInfoDataSourceConfig() string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_entity_info" "test" {}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
    )
}
```

### State Checks (Modern terraform-plugin-testing patterns)

```go
ConfigStateChecks: []statecheck.StateCheck{
    statecheck.ExpectKnownValue(
        "data.bcm_cmpart_entity_info.test",
        tfjsonpath.New("id"),
        knownvalue.NotNull(),
    ),
}
```

## Dependencies on Existing Code

### Required Dependencies

| Component | Location | Usage |
|-----------|----------|-------|
| `BCMClient` | `bcm_client.go` | `CallJSONRPC(ctx, "cmpart", "getBasicEntityInformation")` |
| `getStringValue` | `data_source_cmpart_softwareimages.go` | Null-safe field extraction |
| `limitString` | `bcm_client.go` | Error message truncation |
| Test helpers | `test_helpers.go` | `testAccPreCheck`, `testAccProtoV6ProviderFactories` |

### Pattern References

| Pattern | Reference File | Adaptation |
|---------|----------------|------------|
| Data source structure | `data_source_cmpart_softwareimages.go` | Simpler schema (3 fields vs 20+) |
| Filter implementation | `data_source_cmpart_softwareimages.go` | Glob pattern instead of substring |
| Test structure | `data_source_cmpart_softwareimages_test.go` | Same state check patterns |
| Error handling | `error_messages.go` | Use established error message format |

## Risk Assessment

| Risk | Likelihood | Impact | Mitigation |
|------|------------|--------|------------|
| API response format changes | Low | Medium | Defensive parsing, log unknown fields |
| Large entity counts (1000+) | Medium | Low | Document performance implications |
| Glob pattern edge cases | Low | Low | Use stdlib `filepath.Match`, document limitations |

## Complexity Tracking

No constitution violations identified. Implementation follows established patterns from existing data sources.

## Acceptance Criteria Verification

| Spec Requirement | Implementation |
|------------------|----------------|
| FR-001: Query `cmpart.getBasicEntityInformation` | `CallJSONRPC(ctx, "cmpart", "getBasicEntityInformation")` |
| FR-002: Optional type filter (case-sensitive) | `type` attribute with exact string match |
| FR-003: Optional name_pattern filter (glob) | `name_pattern` with `filepath.Match` |
| FR-004: Return entities with name, type, uuid | `EntityInfoModel` struct |
| FR-005: Map resolveName to name | `getStringValue(apiData, "resolveName")` -> `Name` |
| FR-006: Client-side filtering | Filter loop after API call |
| FR-007: AND logic for combined filters | `matchesEntityFilter` checks both |
| FR-008: Return all entities when no filters | Filter functions return true for null filters |
| FR-009: Unique ID for state | Hash of filter values |
| FR-010: Graceful error handling | Error messages with endpoint and details |
| FR-011: Case-insensitive name pattern | `strings.ToLower` before `filepath.Match` |
