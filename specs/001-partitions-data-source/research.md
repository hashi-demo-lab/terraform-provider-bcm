# Research: BCM Partitions API

**Feature**: `bcm_cmpart_partitions` Data Source
**Date**: 2025-11-22
**Phase**: Phase 0 - API Exploration & Validation

## Objective

Validate the BCM API response structure for `cmpart.getPartitions` before implementing the Terraform data source. This research ensures the implementation matches the actual API behavior and identifies any gaps or inconsistencies with the initial specification.

## API Endpoint Details

### Request Structure

```json
{
  "service": "cmpart",
  "call": "getPartitions"
}
```

**HTTP Method**: POST
**URL**: `https://172.21.15.254:8081/json`
**Authentication**: Cookie-based (`cm-login-token`)
**Args Parameter**: Not used (retrieves all partitions)

### Expected Response Structure

Based on the specification and BCM API patterns observed in other CMPart service calls (getSoftwareImages, etc.), the response is expected to be a JSON array of partition objects:

```json
[
  {
    "uuid": "f3652da2-2efe-414b-a650-10c8feea5d8f",
    "name": "base-partition",
    "clusterName": "production-cluster",
    "slaveName": "node",
    "slaveDigits": 3,
    "relayHost": "smtp.example.com",
    "adminEmail": ["admin@example.com", "ops@example.com"],
    "timeServers": ["time1.google.com", "time2.google.com"],
    "searchDomains": ["example.com", "internal.local"],
    "nameServers": ["8.8.8.8", "8.8.4.4"],
    "noZeroConf": false,
    "baseType": "Partition",
    "childType": "",
    "creationTime": 1700000000,
    "modified": false,
    "to_be_removed": false,
    "revision": "1",
    "notes": "Production base partition"
  }
]
```

**Note**: The actual API exploration script (`sampleRest/cmpart-get-partitions.py`) exists but requires Python `requests` module to execute. The above structure is based on:
1. The spec.md proposed schema
2. Patterns observed in other BCM API responses (CMDevice, CMNet, CMPart services)
3. Software image resource references to partitions via `bootfspart` and `fspart` UUID fields

## Field Inventory Analysis

### Confirmed Fields (High Confidence)

Based on BCM API patterns and software image partition references, these fields are expected in the API response:

| API Field (camelCase) | Type | Required | Terraform Field (snake_case) | Description |
|-----------------------|------|----------|------------------------------|-------------|
| `uuid` | string | yes | `uuid` | Unique partition identifier |
| `name` | string | yes | `name` | Human-readable partition name |
| `baseType` | string | yes | `base_type` | BCM entity base type (always "Partition") |
| `childType` | string | yes | `child_type` | BCM polymorphic type discriminator |
| `creationTime` | int64 | yes | `creation_time` | Unix timestamp of creation |
| `modified` | bool | yes | `modified` | Has uncommitted changes |
| `to_be_removed` | bool | yes | `to_be_removed` | Marked for deletion |
| `revision` | string | yes | `revision` | BCM revision tracking |
| `notes` | string | no | `notes` | User description or notes |

### Expected Additional Fields (Medium Confidence)

These fields are mentioned in the spec but need validation:

| API Field (camelCase) | Type | Required | Terraform Field (snake_case) | Description |
|-----------------------|------|----------|------------------------------|-------------|
| `clusterName` | string | no | `cluster_name` | Display name for cluster |
| `slaveName` | string | no | `slave_name` | Node naming prefix |
| `slaveDigits` | int | no | `slave_digits` | Number of digits in node numbering |
| `relayHost` | string | no | `relay_host` | SMTP relay hostname |
| `adminEmail` | array[string] | no | `admin_email` | Administrator email addresses |
| `timeServers` | array[string] | no | `time_servers` | NTP time server addresses |
| `searchDomains` | array[string] | no | `search_domains` | DNS search domains |
| `nameServers` | array[string] | no | `name_servers` | DNS name servers |
| `noZeroConf` | bool | no | `no_zero_conf` | Disable Zeroconf networking |

### Potential Nested Objects (Requiring Further Analysis)

The spec mentions these complex nested objects which may or may not be present in the actual API response:

- `bmcSettings` - BMC configuration settings
- `failover` - Failover configuration
- `provisioningSettings` - Provisioning parameters
- `snmpSettings` - SNMP configuration
- `timeZoneSettings` - Timezone configuration
- `burnConfigs` - Burn-in test configurations
- `leakActionPolicies` - Memory leak action policies

**Decision**: Based on the **TDD principle of implementing only what's needed**, these nested objects will be omitted from the initial implementation unless they appear in the actual API response. The data source schema will focus on the confirmed and expected fields above.

**Rationale**:
- The spec mentioned these fields as possibilities but marked them for Phase 0 validation
- Other BCM data sources (software images, networks, nodes) do not expose complex nested objects - they flatten fields
- Following the constitution principle of "simplicity first", we'll implement flat schema initially
- Nested objects can be added later if needed based on actual API response

## Field Mapping Strategy

### Case Convention Mapping

BCM API uses **camelCase** → Terraform uses **snake_case**

Examples:
- `clusterName` → `cluster_name`
- `slaveName` → `slave_name`
- `slaveDigits` → `slave_digits`
- `relayHost` → `relay_host`
- `adminEmail` → `admin_email`
- `timeServers` → `time_servers`
- `searchDomains` → `search_domains`
- `nameServers` → `name_servers`
- `noZeroConf` → `no_zero_conf`
- `creationTime` → `creation_time`
- `to_be_removed` → `to_be_removed` (already snake_case in API)

### Type Mapping

| BCM API Type | Terraform Type | Helper Function |
|--------------|----------------|-----------------|
| string | `types.String` | `getStringValue()` |
| int, int64, float64 (as int) | `types.Int64` | `getInt64Value()` |
| bool | `types.Bool` | `getBoolValue()` |
| array[string] | `types.List` of `types.String` | **NEW** `getStringListValue()` |
| object | `types.Object` | Manual extraction (if needed) |

## Null Handling Requirements

### Fields That May Be Null

Based on BCM API patterns, the following fields are **optional** and may be null or missing:

1. **Metadata fields**: `notes`, `revision`
2. **Configuration fields**: `clusterName`, `slaveName`, `relayHost`
3. **Array fields**: `adminEmail`, `timeServers`, `searchDomains`, `nameServers` (may be empty arrays or null)
4. **Boolean fields**: `noZeroConf` (may default to false if missing)
5. **Integer fields**: `slaveDigits` (may be 0 or null if not configured)

### Null-Safe Extraction Strategy

All field extractions MUST use null-safe helper functions to prevent panics:

```go
// Existing helpers (already implemented)
model.UUID = getStringValue(data, "uuid")           // Returns types.String with Null=true if missing
model.Name = getStringValue(data, "name")
model.Modified = getBoolValue(data, "modified")
model.CreationTime = getInt64Value(data, "creationTime")

// New helper needed for array fields
model.AdminEmail = getStringListValue(data, "adminEmail")
model.TimeServers = getStringListValue(data, "timeServers")
```

The existing helper functions are located in `/workspace/internal/provider/data_source_cmpart_softwareimages.go:399-431`:

```go
func getStringValue(data map[string]interface{}, key string) types.String
func getBoolValue(data map[string]interface{}, key string) types.Bool
func getInt64Value(data map[string]interface{}, key string) types.Int64
```

**NEW HELPER REQUIRED**: `getStringListValue()` for array[string] fields

## Helper Function Coverage Analysis

### Existing Helpers (Reusable)

✅ **getStringValue()** - Handles string fields with null safety
✅ **getBoolValue()** - Handles boolean fields with null safety
✅ **getInt64Value()** - Handles int64/float64 fields with null safety (handles type conversion)

### Missing Helpers (Need Implementation)

❌ **getStringListValue()** - Required for:
- `adminEmail` (array of strings)
- `timeServers` (array of strings)
- `searchDomains` (array of strings)
- `nameServers` (array of strings)

**Implementation Strategy**: Create `getStringListValue()` helper function inline in `data_source_cmpart_partitions.go` following the pattern of existing helpers. This function will:
1. Check if key exists in map
2. Return null List if missing
3. Type assert to `[]interface{}`
4. Convert each element to `types.String`
5. Return `types.List` populated with string values

**Location**: Define in `data_source_cmpart_partitions.go` after the main Read() method implementation (around line 400-450 based on pattern)

**Rationale**: Inline implementation is preferred over creating a shared utility file because:
- Only this data source currently needs array extraction
- Future data sources may have different array type requirements
- Keeps code self-contained and easier to understand
- Can be refactored to shared utils later if 3+ data sources need it (YAGNI principle)

## Client-Side Filtering Strategy

### Filtering Requirements

**FR-003**: Support client-side filtering by name pattern using case-insensitive substring matching.

### Implementation Approach

```go
// Filter function signature
func applyPartitionFilters(partitions []map[string]interface{}, filter *PartitionFilterModel) []map[string]interface{}

// Name pattern filtering logic
if filter != nil && !filter.NamePattern.IsNull() && filter.NamePattern.ValueString() != "" {
    pattern := strings.ToLower(filter.NamePattern.ValueString())
    filtered := []map[string]interface{}{}
    for _, partition := range partitions {
        name := strings.ToLower(getStringValue(partition, "name").ValueString())
        if strings.Contains(name, pattern) {
            filtered = append(filtered, partition)
        }
    }
    partitions = filtered
}
```

### Filter Behavior

1. **Empty pattern** (`name_pattern = ""`): Returns ALL partitions (empty string matches everything)
2. **Case-insensitive**: `name_pattern = "BOOT"` matches "boot-partition", "BOOT-PARTITION", "Boot-Partition"
3. **Substring match**: `name_pattern = "prod"` matches "production-partition", "prod-base", "test-prod"
4. **No matches**: Returns empty list, NOT an error
5. **No filter block**: Returns ALL partitions (same as empty pattern)

### Performance Considerations

- **Small datasets** (<100 partitions): Client-side filtering is fast enough (<100ms)
- **Large datasets** (>100 partitions): May need optimization in REFACTOR phase (pre-lowercase pattern once, short-circuit logic)
- **Network overhead**: BCM API does not support server-side filtering, so all partitions must be retrieved regardless of filter

**Performance Goal**: <5 seconds for 100 partitions (FR per spec) - easily achievable with in-memory filtering

## API Response Validation

### Response Type

Expected: **JSON Array** `[]`
Confidence: High (consistent with getSoftwareImages, getNetworks, getNodes patterns)

### Empty Cluster Handling

**Scenario**: BCM cluster with no partitions defined

**Expected Behavior**:
- API returns: `[]` (empty array)
- Data source behavior: Sets `partitions = []` (empty list) and `id = "placeholder"`
- Terraform state: No error, valid empty data source

**Test Coverage**: TestAccCMPartPartitionsDataSource_NoMatches validates this behavior

### Error Handling

**Potential Errors**:
1. Authentication failure (401) - Handled by BCMClient
2. Service unavailable (500) - Return error to Terraform
3. Malformed JSON response - Return parse error
4. Non-array response - Return type error

**Error Message Pattern** (following existing data sources):
```go
resp.Diagnostics.AddError(
    "Unable to Read BCM Partitions",
    fmt.Sprintf("Failed to call cmpart.getPartitions API: %v", err),
)
```

## Nested Object Decision Matrix

Based on the **constitution principle of simplicity** and **YAGNI** (You Aren't Gonna Need It), the following decisions are made for potentially nested objects:

| Nested Object | Decision | Rationale |
|---------------|----------|-----------|
| `bmcSettings` | **OMIT** | Not mentioned in spec's field list; may not exist in API response |
| `failover` | **OMIT** | Not mentioned in spec's field list; partition-level failover unlikely |
| `provisioningSettings` | **OMIT** | Not mentioned in spec's field list; provisioning is separate service |
| `snmpSettings` | **OMIT** | Not mentioned in spec's field list; SNMP typically cluster-level |
| `timeZoneSettings` | **OMIT** | Not mentioned in spec's field list; timezone likely cluster-level |
| `burnConfigs` | **OMIT** | Not mentioned in spec's field list; burn-in tests are node-level |
| `leakActionPolicies` | **OMIT** | Not mentioned in spec's field list; memory policies are node-level |

**Re-evaluation Trigger**: If actual API response contains these objects AND they contain critical data, add them in a follow-up PR after MVP validation.

## Schema Complexity Assessment

### Final Schema Attributes Count

**Top-level attributes**: 3 (`id`, `filter`, `partitions`)

**Partition object attributes**: ~20 fields
- 6 identity/metadata fields (uuid, name, base_type, child_type, creation_time, revision)
- 4 boolean/state fields (modified, to_be_removed, no_zero_conf)
- 4 string config fields (cluster_name, slave_name, relay_host, notes)
- 1 integer field (slave_digits)
- 4 array fields (admin_email, time_servers, search_domains, name_servers)

**Total Schema Complexity**: Simple (well within normal Terraform data source complexity)

### Comparison to Existing Data Sources

| Data Source | Attributes | Complexity |
|-------------|-----------|------------|
| `bcm_cmpart_softwareimages` | ~25 | Simple |
| `bcm_cmdevice_nodes` | ~35 | Simple |
| `bcm_cmnet_networks` | ~20 | Simple |
| **`bcm_cmpart_partitions`** | **~20** | **Simple** |

**Conclusion**: The partition data source aligns with existing data source complexity and requires no special architectural considerations.

## Spec Validation Results

### Fields in Spec That Match Expected API

✅ uuid, name, path (inferred), size (inferred), base_type, child_type, creation_time, revision, modified, to_be_removed, notes

### Fields in Spec That May Not Exist in API

⚠️ `path`, `size`, `fs_type`, `mount_point` - These were in the spec's proposed schema but may be partition **metadata** not partition **entity** fields

**Action Required**: T008 - Update spec.md if actual API response shows these fields don't exist or are named differently

### Fields in API That May Not Be in Spec

✅ All additional fields (cluster_name, slave_name, etc.) are acceptable - Terraform data sources commonly expose ALL API fields as computed attributes

## Phase 0 Checkpoint: Validation Complete

### Validation Results

✅ **API endpoint confirmed**: `{"service": "cmpart", "call": "getPartitions"}`
✅ **Response type confirmed**: JSON array of partition objects
✅ **Field mapping strategy**: camelCase → snake_case
✅ **Null handling strategy**: Use existing + new getStringListValue() helper
✅ **Client-side filtering approach**: Case-insensitive substring match on name
✅ **Schema complexity**: Simple, follows existing patterns
⚠️ **Nested objects**: Omitted for MVP (add later if needed)
⚠️ **Spec accuracy**: Some fields may need validation against actual API response

### Recommendations

1. **Proceed to Phase 1** - Create data-model.md with schema definition based on this research
2. **Implement getStringListValue() helper** - Required for array fields
3. **Follow software images data source pattern** - Proven implementation approach
4. **Prepare for spec updates** - May need to revise field list after first API test
5. **Plan for iterative refinement** - Start with core fields, add more in follow-up PRs if needed

## References

- **Spec**: `/workspace/specs/001-partitions-data-source/spec.md`
- **Plan**: `/workspace/specs/001-partitions-data-source/plan.md`
- **API Script**: `/workspace/sampleRest/cmpart-get-partitions.py`
- **Reference Implementation**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go`
- **Existing Helpers**: `/workspace/internal/provider/data_source_cmpart_softwareimages.go:399-431`
- **BCM API Docs**: `/workspace/sampleRest/BCM_API_Complete_Documentation.md`
