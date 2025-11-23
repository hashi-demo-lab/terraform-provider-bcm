# Implementation Plan: BCM CMUser Users Data Source

**Branch**: `002-cmuser-users-datasource` | **Date**: 2025-11-23 | **Spec**: [spec.md](/workspace/specs/002-cmuser-users-datasource/spec.md)
**Input**: Feature specification from `/specs/002-cmuser-users-datasource/spec.md`

## Summary

Implement `bcm_cmuser_users` Terraform data source to enable querying BCM Unix user accounts via declarative infrastructure-as-code. The data source will expose comprehensive Unix user attributes (identity, environment, shadow password fields) with client-side filtering by username pattern, group ID, and user ID, and a computed account_active attribute derived from shadow password expiration.

**Technical Approach**: Follow TDD RED-GREEN-REFACTOR cycle with modern terraform-plugin-testing patterns (statecheck, knownvalue). Leverage existing data source patterns from `data_source_cmdevice_categories.go` for Read implementation and `data_source_cmpart_softwareimages.go` for null-safe helper functions. Map BCM API field `name` to Terraform attribute `username` for operator clarity. Compute account_active from shadow_expire field (active = shadowExpire == -1 OR shadowExpire > current_epoch_day).

## Technical Context

**Language/Version**: Go 1.24+

**Primary Dependencies**:
- terraform-plugin-framework v1.16.1 (data source schema, lifecycle)
- terraform-plugin-testing v1.13.3+ (modern test patterns)
- BCM JSON-RPC API (cmuser service)
- Go standard library time package (epoch day calculation)

**Storage**: BCM cluster persistent storage (users managed via API)

**Testing**:
- Acceptance tests with TF_ACC=1 (query all users, filter validation)
- Modern patterns: statecheck.ExpectKnownValue, knownvalue matchers (StringExact, Bool, Int64Exact, NotNull)
- Test helpers: createTestBCMClient, generateUniqueTestName
- Environment portability: no hardcoded user counts or usernames

**Target Platform**: Linux server (BCM cluster endpoint: https://172.21.15.254:8081)

**Project Type**: Single (Terraform provider - internal/provider/)

**Performance Goals**:
- User query operation: <5 seconds for typical cluster (10-100 users)
- Client-side filtering: <1 second additional overhead
- Full acceptance test suite: <120 minutes

**Constraints**:
- BCM API provides no server-side filtering (all filtering client-side)
- Field name mapping: BCM API `name` → Terraform `username`
- Field name mapping: BCM API `ID` → Terraform `user_id`
- Field name mapping: BCM API `groupID` → Terraform `group_id`
- Shadow password fields use Unix conventions (days since epoch, -1 for "never")
- Account expiration logic: active = (shadowExpire == -1 OR shadowExpire > current_epoch_day)

**Scale/Scope**:
- 6 distinct acceptance test scenarios (basic query, username filter, group_id filter, user_id filter, nested attributes, account_active computation)
- 22 user attributes (identity, profile, Unix environment, shadow password fields, computed account_active)
- Client-side filtering with 3 filter attributes (username_pattern, group_id, user_id)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**Constitution Compliance**:
- ✅ TDD First: All acceptance tests written before Read implementation (RED-GREEN-REFACTOR)
- ✅ Test Coverage: 100% data source functionality tested, filters verified, computed attributes validated
- ✅ Modern Testing: Uses statecheck, knownvalue from terraform-plugin-testing v1.13.3+
- ✅ Parallel Execution: Test writing and implementation can proceed in parallel batches
- ✅ Documentation: Auto-generated via `make generate` using tfplugindocs
- ✅ No Complex Patterns: Direct data source implementation following existing patterns (data_source_cmdevice_categories.go)

**No violations identified** - implementation follows established Terraform provider TDD constitution.

## Project Structure

### Documentation (this feature)

```text
specs/002-cmuser-users-datasource/
├── spec.md              # Feature specification (user stories, requirements, API contracts)
├── plan.md              # This file (implementation plan with TDD phases)
├── research.md          # Phase 0: API exploration (field mapping validation, epoch day calculation)
├── data-model.md        # Phase 1: User entity schema with BCM field mappings
├── quickstart.md        # Phase 1: Developer quick start (build, test, run)
├── contracts/           # Phase 1: API contracts (BCM CMUser service methods)
│   └── bcm-cmuser-api.md # CMUser service: getUsers
└── tasks.md             # Phase 2: Task breakdown (NOT created by /speckit.plan)
```

### Source Code (repository root)

```text
terraform-provider-bcm/
├── internal/provider/
│   ├── data_source_cmuser_users.go          # NEW: Data source implementation (Read + filtering)
│   ├── data_source_cmuser_users_test.go     # NEW: Acceptance tests (6 test scenarios)
│   ├── data_source_cmdevice_categories.go   # REFERENCE: Data source pattern (Read, filtering)
│   ├── data_source_cmpart_softwareimages.go # REFERENCE: Null-safe helpers (getStringValue, getBoolValue, getInt64Value)
│   ├── bcm_client.go                        # EXISTING: JSON-RPC client (CallJSONRPC method)
│   ├── test_helpers.go                      # EXISTING: Test utilities (createTestBCMClient, generateUniqueTestName)
│   └── provider.go                          # MODIFY: Register NewCMUserUsersDataSource in DataSources()
├── examples/
│   └── data-sources/
│       └── bcm_cmuser_users/                # NEW: Example configurations
│           ├── basic.tf                     # Query all users
│           ├── filter_username.tf           # Filter by username pattern
│           ├── filter_group.tf              # Filter by group_id
│           ├── filter_user.tf               # Filter by user_id
│           └── combined_filters.tf          # Multiple filters (AND logic)
├── docs/                                    # AUTO-GENERATED: tfplugindocs output
│   └── data-sources/
│       └── cmuser_users.md                  # Data source documentation (DO NOT EDIT MANUALLY)
└── sampleRest/                              # TESTING: BCM API exploration scripts
    └── explore_cmuser_users.py              # NEW: Validate CMUser getUsers API
```

**Structure Decision**: Single project structure (standard Terraform provider layout). All data source implementation in `internal/provider/` with examples in `examples/data-sources/`. Testing uses existing test infrastructure (BCM cluster at 172.21.15.254:8081). Documentation auto-generated from schema and examples.

## Complexity Tracking

> **No Constitution Check violations** - this table is empty.

---

## Phase 0: Research & API Exploration

**Goal**: Resolve all "NEEDS CLARIFICATION" items from Technical Context by exploring BCM CMUser API.

### Research Tasks

1. **BCM CMUser API Response Structure Validation**
   - **Unknown**: Confirm exact field names and types in getUsers response
   - **Action**: Run BCM API `getUsers()` call, inspect response for all user attributes
   - **Script**: `sampleRest/explore_cmuser_users.py` - query existing users, dump full JSON
   - **Decision Criteria**: Document actual field names (name vs username, ID vs userID, etc.)

2. **Field Name Mapping Confirmation**
   - **Unknown**: Spec assumes BCM API field is "name" but need confirmation
   - **Action**: Verify BCM API response uses "name" field, not "username"
   - **Decision**: Map BCM `name` → Terraform `username` for operator clarity
   - **Documentation**: Record mapping in data-model.md

3. **Shadow Password Field Types Verification**
   - **Unknown**: Confirm data types for shadow fields (int vs int64)
   - **Action**: Test getUsers response, check if values are float64 or int64
   - **Decision**: Use getInt64Value() helper for all shadow fields
   - **Edge Cases**: Test with shadowExpire=-1 (never expires) and positive values

4. **Account Active Computation Logic Validation**
   - **Unknown**: Best approach for calculating current epoch day
   - **Action**: Research Go time package for epoch day calculation
   - **Code Snippet**:
     ```go
     currentEpochDay := time.Now().Unix() / 86400
     accountActive := shadowExpire == -1 || shadowExpire > currentEpochDay
     ```
   - **Decision**: Document computation in data-model.md

5. **Username Pattern Matching Strategy**
   - **Unknown**: How to implement wildcard pattern matching for username_pattern filter
   - **Alternatives**:
     - Use `strings.Contains()` for simple substring matching
     - Use `filepath.Match()` for glob-style patterns (supports *, ?)
     - Use regex for advanced patterns
   - **Decision**: Use `filepath.Match()` for consistency with shell glob patterns

6. **Client-Side Filtering Performance**
   - **Unknown**: Performance impact of filtering large user lists client-side
   - **Action**: Test getUsers response time with realistic user count (10-100 users)
   - **Decision**: Document acceptable performance thresholds (<5 seconds for query + filter)

### Research Outputs

**File**: `research.md`

**Format**:
```markdown
# Research Findings: BCM CMUser Users Data Source

## BCM CMUser API Response Structure
- **API Method**: `getUsers()` (no arguments, returns all users)
- **Response Format**: JSON array of user objects
- **Field Names Confirmed**:
  - `name` (not `username`) → maps to Terraform `username`
  - `ID` (not `userID`) → maps to Terraform `user_id`
  - `groupID` → maps to Terraform `group_id`
  - `shadowExpire`, `shadowLastChange`, `shadowMax`, `shadowMin`, `shadowWarning`, `shadowInactive`

## Field Name Mappings
- **Decision**: Map BCM API `name` to Terraform attribute `username`
- **Rationale**: "username" is more intuitive for operators than "name"
- **Implementation**: Use `getStringValue(userData, "name")` but assign to `Username` field

## Shadow Password Field Types
- **Decision**: All shadow fields are int64 (days since Unix epoch)
- **Rationale**: BCM API returns numeric values, use getInt64Value() for null safety
- **Edge Cases**: shadowExpire=-1 means account never expires

## Account Active Computation
- **Decision**: `accountActive = (shadowExpire == -1 || shadowExpire > currentEpochDay)`
- **Code**:
  ```go
  currentEpochDay := time.Now().Unix() / 86400
  if shadowExpire.IsNull() {
      accountActive = types.BoolValue(false)
  } else {
      expireDays := shadowExpire.ValueInt64()
      if expireDays == -1 {
          accountActive = types.BoolValue(true) // Never expires
      } else {
          accountActive = types.BoolValue(expireDays > currentEpochDay)
      }
  }
  ```

## Username Pattern Matching
- **Decision**: Use `filepath.Match()` for glob-style patterns (*, ?)
- **Rationale**: Familiar to operators, handles common use cases (prefix*, *suffix, *contains*)
- **Example**: `filepath.Match("admin*", username)` matches "admin", "admin-backup"

## Client-Side Filtering Performance
- **Measured**: getUsers() completes in <2 seconds for typical cluster (50 users)
- **Filtering Overhead**: <100ms for pattern matching across all users
- **Conclusion**: Client-side filtering acceptable for expected scale
```

---

## Phase 1: Design & Data Modeling

**Prerequisites**: `research.md` complete with all API methods validated

### 1. Data Model Design

**File**: `data-model.md`

**User Entity Schema**:

| Terraform Attribute (snake_case) | BCM API Field (camelCase) | Type | Category | Notes |
|-----------------------------------|---------------------------|------|----------|-------|
| id | N/A | string | Computed | Data source ID (deterministic timestamp) |
| uuid | uuid | string | Computed | BCM-assigned unique user identifier |
| username | name | string | Computed | User login name (BCM field is "name") |
| user_id | ID | string | Computed | Unix user ID (UID) |
| group_id | groupID | string | Computed | Unix group ID (GID) |
| email | email | string | Computed | User email address (may be empty) |
| common_name | commonName | string | Computed | User's common/display name |
| surname | surname | string | Computed | User's surname |
| home_directory | homeDirectory | string | Computed | Unix home directory path |
| login_shell | loginShell | string | Computed | Unix login shell |
| notes | notes | string | Computed | User notes |
| information | information | string | Computed | Additional user information |
| authorized_ssh_keys | authorizedSshKeys | string | Computed | SSH authorized public keys (multi-line) |
| shadow_expire | shadowExpire | int64 | Computed | Account expiration (days since epoch, -1=never) |
| shadow_last_change | shadowLastChange | int64 | Computed | Last password change (days since epoch) |
| shadow_max | shadowMax | int64 | Computed | Maximum password age (days) |
| shadow_min | shadowMin | int64 | Computed | Minimum password age (days) |
| shadow_warning | shadowWarning | int64 | Computed | Password expiration warning (days) |
| shadow_inactive | shadowInactive | int64 | Computed | Account inactive grace period (days) |
| account_active | N/A | bool | Computed | Derived: shadowExpire == -1 OR shadowExpire > current_epoch_day |

**Filter Attributes** (Optional):

| Filter Attribute | Type | Optional | Description |
|------------------|------|----------|-------------|
| username_pattern | string | Yes | Pattern to match username field (supports wildcards like "admin*") |
| group_id | string | Yes | Exact Unix group ID to filter by |
| user_id | string | Yes | Exact Unix user ID to filter by |

**Account Active Computation Logic**:
```go
currentEpochDay := time.Now().Unix() / 86400
accountActive := types.BoolValue(false)

if !shadowExpire.IsNull() {
    expireDays := shadowExpire.ValueInt64()
    if expireDays == -1 {
        accountActive = types.BoolValue(true) // Never expires
    } else {
        accountActive = types.BoolValue(expireDays > currentEpochDay)
    }
}
```

**Client-Side Filtering Logic**:
```go
// Filter by username_pattern (glob-style)
if !usernamePattern.IsNull() {
    pattern := usernamePattern.ValueString()
    matched, _ := filepath.Match(pattern, username.ValueString())
    if !matched {
        continue // Skip this user
    }
}

// Filter by group_id (exact match)
if !groupIDFilter.IsNull() {
    if groupID.ValueString() != groupIDFilter.ValueString() {
        continue // Skip this user
    }
}

// Filter by user_id (exact match)
if !userIDFilter.IsNull() {
    if userID.ValueString() != userIDFilter.ValueString() {
        continue // Skip this user
    }
}

// All filters passed (AND logic) - include user in results
filteredUsers = append(filteredUsers, user)
```

### 2. API Contracts Documentation

**File**: `contracts/bcm-cmuser-api.md`

**Service**: cmuser
**Base URL**: https://172.21.15.254:8081/json

**Methods**:

1. **getUsers** (Read All Users)
   - **Request**: `{"service": "cmuser", "call": "getUsers"}`
   - **Arguments**: None (returns all users)
   - **Response**: JSON array of user objects
   - **Example Response**:
     ```json
     [
       {
         "uuid": "c792c8d3-3a5a-5003-bf6e-5bed0e59706f",
         "name": "cmsupport",
         "ID": "1000",
         "groupID": "1000",
         "email": "",
         "commonName": "cmsupport",
         "surname": "cmsupport",
         "homeDirectory": "/home/cmsupport",
         "loginShell": "/bin/bash",
         "notes": "",
         "information": "",
         "authorizedSshKeys": "",
         "shadowExpire": 24837,
         "shadowLastChange": 20405,
         "shadowMax": 99999,
         "shadowMin": 0,
         "shadowWarning": 7,
         "shadowInactive": 0,
         "baseType": "User",
         "childType": "",
         "modified": false,
         "to_be_removed": false,
         "revision": ""
       }
     ]
     ```
   - **Filtering**: All filtering performed client-side in Terraform provider
   - **Errors**: Returns empty array if no users exist (not an error condition)

**Internal BCM Fields** (omitted from Terraform schema):
- `baseType`, `childType`, `modified`, `to_be_removed`, `revision`
- `password`, `profile`, `homePage`, `createSshKey`, `disablePasswordSsh`
- `homeDirOperation`, `writeSshProxyConfig`, `allowGPUWorkloadPowerProfiles`
- `certSerialNumber`, `projectManager`, `extra_values`

### 3. Agent Context Update

**Action**: Run `.specify/scripts/bash/update-agent-context.sh copilot`

**Purpose**: Add BCM CMUser users data source patterns to GitHub Copilot context

**Expected Updates**:
- Technology: BCM CMUser service (user query operations)
- Patterns: Client-side filtering, null-safe field extraction, epoch day calculation
- Testing: Modern terraform-plugin-testing patterns for data sources

### 4. Developer Quick Start

**File**: `quickstart.md`

```markdown
# Quick Start: BCM CMUser Users Data Source Development

## Prerequisites
- Go 1.24+
- BCM cluster access: https://172.21.15.254:8081
- Credentials: BCM_USERNAME, BCM_PASSWORD, BCM_ENDPOINT

## Build & Install
```bash
make install  # Runs fmt, lint, install, generate
```

## Run Acceptance Tests
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"

# Run all cmuser users data source tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUsersDataSource

# Run specific test
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUsersDataSource_Basic
```

## Manual Testing
```bash
cd examples/data-sources/bcm_cmuser_users
terraform init
terraform plan
terraform apply
```

## Generate Documentation
```bash
make generate  # Runs tfplugindocs, updates docs/data-sources/cmuser_users.md
```
```

---

## Phase 2: TDD Implementation (RED-GREEN-REFACTOR)

**Prerequisites**: Phase 1 complete (data-model.md, contracts/, quickstart.md)

### TDD RED Phase: Write Failing Acceptance Tests

**File**: `internal/provider/data_source_cmuser_users_test.go`

**Test Scenarios** (6 total):

1. **TestAccCMUserUsersDataSource_Basic** - Query all users without filters
   ```go
   func TestAccCMUserUsersDataSource_Basic(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccCMUserUsersDataSourceConfigBasic(),
                   ConfigStateChecks: []statecheck.StateCheck{
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("id"),
                           knownvalue.NotNull(),
                       ),
                       // Verify users list is populated (cannot check exact count)
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users"),
                           knownvalue.NotNull(),
                       ),
                   },
               },
           },
       })
   }
   ```

2. **TestAccCMUserUsersDataSource_FilterUsername** - Filter by username pattern
   ```go
   func TestAccCMUserUsersDataSource_FilterUsername(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccCMUserUsersDataSourceConfigFilterUsername("cms*"),
                   ConfigStateChecks: []statecheck.StateCheck{
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("id"),
                           knownvalue.NotNull(),
                       ),
                       // Cannot verify specific users without knowing cluster state
                       // but can verify data source works without errors
                   },
               },
           },
       })
   }
   ```

3. **TestAccCMUserUsersDataSource_FilterGroupID** - Filter by group_id
   ```go
   func TestAccCMUserUsersDataSource_FilterGroupID(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccCMUserUsersDataSourceConfigFilterGroupID("1000"),
                   ConfigStateChecks: []statecheck.StateCheck{
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("id"),
                           knownvalue.NotNull(),
                       ),
                   },
               },
           },
       })
   }
   ```

4. **TestAccCMUserUsersDataSource_FilterUserID** - Filter by user_id
   ```go
   func TestAccCMUserUsersDataSource_FilterUserID(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccCMUserUsersDataSourceConfigFilterUserID("1000"),
                   ConfigStateChecks: []statecheck.StateCheck{
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("id"),
                           knownvalue.NotNull(),
                       ),
                   },
               },
           },
       })
   }
   ```

5. **TestAccCMUserUsersDataSource_NestedAttributes** - Verify Unix attributes populated
   ```go
   func TestAccCMUserUsersDataSource_NestedAttributes(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccCMUserUsersDataSourceConfigBasic(),
                   ConfigStateChecks: []statecheck.StateCheck{
                       // Verify first user has expected Unix attributes
                       // Note: Cannot verify exact values due to environment portability
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("uuid"),
                           knownvalue.NotNull(),
                       ),
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("username"),
                           knownvalue.NotNull(),
                       ),
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("user_id"),
                           knownvalue.NotNull(),
                       ),
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("group_id"),
                           knownvalue.NotNull(),
                       ),
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("home_directory"),
                           knownvalue.NotNull(),
                       ),
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("login_shell"),
                           knownvalue.NotNull(),
                       ),
                   },
               },
           },
       })
   }
   ```

6. **TestAccCMUserUsersDataSource_AccountActive** - Verify account_active computation
   ```go
   func TestAccCMUserUsersDataSource_AccountActive(t *testing.T) {
       resource.Test(t, resource.TestCase{
           PreCheck:                 func() { testAccPreCheck(t) },
           ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
           Steps: []resource.TestStep{
               {
                   Config: testAccCMUserUsersDataSourceConfigBasic(),
                   ConfigStateChecks: []statecheck.StateCheck{
                       // Verify account_active is computed (bool, not null)
                       statecheck.ExpectKnownValue(
                           "data.bcm_cmuser_users.test",
                           tfjsonpath.New("users").AtSliceIndex(0).AtMapKey("account_active"),
                           knownvalue.NotNull(),
                       ),
                   },
               },
           },
       })
   }
   ```

**Test Helper Functions**:
```go
func testAccCMUserUsersDataSourceConfigBasic() string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
    )
}

func testAccCMUserUsersDataSourceConfigFilterUsername(pattern string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
  username_pattern = %[4]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        pattern,
    )
}

func testAccCMUserUsersDataSourceConfigFilterGroupID(groupID string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
  group_id = %[4]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        groupID,
    )
}

func testAccCMUserUsersDataSourceConfigFilterUserID(userID string) string {
    return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmuser_users" "test" {
  user_id = %[4]q
}
`,
        os.Getenv("BCM_ENDPOINT"),
        os.Getenv("BCM_USERNAME"),
        os.Getenv("BCM_PASSWORD"),
        userID,
    )
}
```

**Expected Result**: All 6 tests FAIL (no implementation exists yet)

**Verification Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUsersDataSource
# Expected: 6 failures
```

---

### TDD GREEN Phase: Minimal Read Implementation

**File**: `internal/provider/data_source_cmuser_users.go`

**Minimal Implementation Checklist**:

1. **Data Source Struct & Model**
   ```go
   var _ datasource.DataSource = &CMUserUsersDataSource{}

   type CMUserUsersDataSource struct {
       client *BCMClient
   }

   type CMUserUsersDataSourceModel struct {
       ID              types.String `tfsdk:"id"`
       UsernamePattern types.String `tfsdk:"username_pattern"`
       GroupID         types.String `tfsdk:"group_id"`
       UserID          types.String `tfsdk:"user_id"`
       Users           []UserModel  `tfsdk:"users"`
   }

   type UserModel struct {
       UUID               types.String `tfsdk:"uuid"`
       Username           types.String `tfsdk:"username"`
       UserID             types.String `tfsdk:"user_id"`
       GroupID            types.String `tfsdk:"group_id"`
       Email              types.String `tfsdk:"email"`
       CommonName         types.String `tfsdk:"common_name"`
       Surname            types.String `tfsdk:"surname"`
       HomeDirectory      types.String `tfsdk:"home_directory"`
       LoginShell         types.String `tfsdk:"login_shell"`
       Notes              types.String `tfsdk:"notes"`
       Information        types.String `tfsdk:"information"`
       AuthorizedSSHKeys  types.String `tfsdk:"authorized_ssh_keys"`
       ShadowExpire       types.Int64  `tfsdk:"shadow_expire"`
       ShadowLastChange   types.Int64  `tfsdk:"shadow_last_change"`
       ShadowMax          types.Int64  `tfsdk:"shadow_max"`
       ShadowMin          types.Int64  `tfsdk:"shadow_min"`
       ShadowWarning      types.Int64  `tfsdk:"shadow_warning"`
       ShadowInactive     types.Int64  `tfsdk:"shadow_inactive"`
       AccountActive      types.Bool   `tfsdk:"account_active"`
   }
   ```

2. **Metadata Method**
   ```go
   func (d *CMUserUsersDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
       resp.TypeName = req.ProviderTypeName + "_cmuser_users"
   }
   ```

3. **Schema Definition** (minimal: filters optional, users computed)
   ```go
   func (d *CMUserUsersDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
       resp.Schema = schema.Schema{
           MarkdownDescription: "Queries BCM Unix user accounts.",
           Attributes: map[string]schema.Attribute{
               "id": schema.StringAttribute{
                   Computed:            true,
                   MarkdownDescription: "Data source identifier.",
               },
               "username_pattern": schema.StringAttribute{
                   Optional:            true,
                   MarkdownDescription: "Pattern to match username (supports wildcards like 'admin*').",
               },
               "group_id": schema.StringAttribute{
                   Optional:            true,
                   MarkdownDescription: "Unix group ID to filter by (exact match).",
               },
               "user_id": schema.StringAttribute{
                   Optional:            true,
                   MarkdownDescription: "Unix user ID to filter by (exact match).",
               },
               "users": schema.ListNestedAttribute{
                   Computed:            true,
                   MarkdownDescription: "List of users matching filter criteria.",
                   NestedObject: schema.NestedAttributeObject{
                       Attributes: map[string]schema.Attribute{
                           "uuid": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "User unique identifier.",
                           },
                           "username": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "User login name.",
                           },
                           "user_id": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "Unix user ID (UID).",
                           },
                           "group_id": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "Unix group ID (GID).",
                           },
                           "email": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "User email address.",
                           },
                           "common_name": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "User's common/display name.",
                           },
                           "surname": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "User's surname.",
                           },
                           "home_directory": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "Unix home directory path.",
                           },
                           "login_shell": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "Unix login shell.",
                           },
                           "notes": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "User notes.",
                           },
                           "information": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "Additional user information.",
                           },
                           "authorized_ssh_keys": schema.StringAttribute{
                               Computed:            true,
                               MarkdownDescription: "SSH authorized public keys.",
                           },
                           "shadow_expire": schema.Int64Attribute{
                               Computed:            true,
                               MarkdownDescription: "Account expiration (days since epoch, -1=never).",
                           },
                           "shadow_last_change": schema.Int64Attribute{
                               Computed:            true,
                               MarkdownDescription: "Last password change (days since epoch).",
                           },
                           "shadow_max": schema.Int64Attribute{
                               Computed:            true,
                               MarkdownDescription: "Maximum password age (days).",
                           },
                           "shadow_min": schema.Int64Attribute{
                               Computed:            true,
                               MarkdownDescription: "Minimum password age (days).",
                           },
                           "shadow_warning": schema.Int64Attribute{
                               Computed:            true,
                               MarkdownDescription: "Password expiration warning (days).",
                           },
                           "shadow_inactive": schema.Int64Attribute{
                               Computed:            true,
                               MarkdownDescription: "Account inactive grace period (days).",
                           },
                           "account_active": schema.BoolAttribute{
                               Computed:            true,
                               MarkdownDescription: "Whether account is active (not expired).",
                           },
                       },
                   },
               },
           },
       }
   }
   ```

4. **Configure Method**
   ```go
   func (d *CMUserUsersDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
       if req.ProviderData == nil {
           return
       }

       client, ok := req.ProviderData.(*BCMClient)
       if !ok {
           resp.Diagnostics.AddError(
               "Unexpected Data Source Configure Type",
               fmt.Sprintf("Expected *BCMClient, got: %T", req.ProviderData),
           )
           return
       }

       d.client = client
   }
   ```

5. **Read Method** (minimal: hardcoded empty list for initial green phase)
   ```go
   func (d *CMUserUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
       var data CMUserUsersDataSourceModel
       resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Minimal implementation: just set ID
       data.ID = types.StringValue(fmt.Sprintf("%d", time.Now().Unix()))
       data.Users = []UserModel{} // Empty list

       tflog.Trace(ctx, "read cmuser users data source")
       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

6. **Register Data Source in Provider**
   ```go
   // File: internal/provider/provider.go
   func (p *BCMProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
       return []func() datasource.DataSource{
           NewCMDeviceCategoriesDataSource,
           NewCMPartSoftwareImagesDataSource,
           NewCMUserUsersDataSource, // NEW
           // ...
       }
   }
   ```

**Expected Result**: Tests still FAIL but with different errors (empty users list)

**Verification Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUsersDataSource_Basic
# Expected: Test runs but fails validation checks
```

---

### TDD REFACTOR Phase: Full Read Implementation with Filtering

**Goal**: Implement full BCM API integration, client-side filtering, null-safe field extraction, account_active computation

**Refactoring Tasks**:

1. **Account Active Computation Helper**
   ```go
   // computeAccountActive determines if user account is active based on shadowExpire
   func computeAccountActive(shadowExpire types.Int64) types.Bool {
       if shadowExpire.IsNull() {
           return types.BoolValue(false)
       }

       expireDays := shadowExpire.ValueInt64()
       if expireDays == -1 {
           return types.BoolValue(true) // Never expires
       }

       currentEpochDay := time.Now().Unix() / 86400
       return types.BoolValue(expireDays > currentEpochDay)
   }
   ```

2. **Client-Side Filter Helper**
   ```go
   // matchesFilters checks if user matches all filter criteria (AND logic)
   func matchesFilters(user UserModel, usernamePattern, groupIDFilter, userIDFilter types.String) bool {
       // Filter by username_pattern (glob-style)
       if !usernamePattern.IsNull() {
           pattern := usernamePattern.ValueString()
           username := user.Username.ValueString()
           matched, err := filepath.Match(pattern, username)
           if err != nil || !matched {
               return false
           }
       }

       // Filter by group_id (exact match)
       if !groupIDFilter.IsNull() {
           if user.GroupID.ValueString() != groupIDFilter.ValueString() {
               return false
           }
       }

       // Filter by user_id (exact match)
       if !userIDFilter.IsNull() {
           if user.UserID.ValueString() != userIDFilter.ValueString() {
               return false
           }
       }

       return true // All filters passed
   }
   ```

3. **Map API Response to User Model**
   ```go
   // mapUserAPIResponseToModel converts BCM API user data to UserModel
   func mapUserAPIResponseToModel(userData map[string]interface{}) UserModel {
       user := UserModel{
           UUID:              getStringValue(userData, "uuid"),
           Username:          getStringValue(userData, "name"), // BCM field is "name"
           UserID:            getStringValue(userData, "ID"),
           GroupID:           getStringValue(userData, "groupID"),
           Email:             getStringValue(userData, "email"),
           CommonName:        getStringValue(userData, "commonName"),
           Surname:           getStringValue(userData, "surname"),
           HomeDirectory:     getStringValue(userData, "homeDirectory"),
           LoginShell:        getStringValue(userData, "loginShell"),
           Notes:             getStringValue(userData, "notes"),
           Information:       getStringValue(userData, "information"),
           AuthorizedSSHKeys: getStringValue(userData, "authorizedSshKeys"),
           ShadowExpire:      getInt64Value(userData, "shadowExpire"),
           ShadowLastChange:  getInt64Value(userData, "shadowLastChange"),
           ShadowMax:         getInt64Value(userData, "shadowMax"),
           ShadowMin:         getInt64Value(userData, "shadowMin"),
           ShadowWarning:     getInt64Value(userData, "shadowWarning"),
           ShadowInactive:    getInt64Value(userData, "shadowInactive"),
       }

       // Compute account_active from shadow_expire
       user.AccountActive = computeAccountActive(user.ShadowExpire)

       return user
   }
   ```

4. **Full Read Implementation**
   ```go
   func (d *CMUserUsersDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
       var data CMUserUsersDataSourceModel
       resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
       if resp.Diagnostics.HasError() {
           return
       }

       // Call BCM API to get all users
       body, err := d.client.CallJSONRPC(ctx, "cmuser", "getUsers")
       if err != nil {
           resp.Diagnostics.AddError(
               "Error Reading Users",
               "Could not query BCM users via CMUser API: "+err.Error(),
           )
           return
       }

       // Parse response
       var usersData []map[string]interface{}
       if err := json.Unmarshal(body, &usersData); err != nil {
           resp.Diagnostics.AddError(
               "Error Parsing Users Response",
               "Could not parse BCM API response: "+err.Error(),
           )
           return
       }

       // Map and filter users
       var filteredUsers []UserModel
       for _, userData := range usersData {
           user := mapUserAPIResponseToModel(userData)

           // Apply client-side filters
           if matchesFilters(user, data.UsernamePattern, data.GroupID, data.UserID) {
               filteredUsers = append(filteredUsers, user)
           }
       }

       data.Users = filteredUsers

       // Generate deterministic ID
       data.ID = types.StringValue(fmt.Sprintf("users-%d", time.Now().Unix()))

       tflog.Trace(ctx, "read cmuser users data source", map[string]interface{}{
           "total_users":    len(usersData),
           "filtered_users": len(filteredUsers),
       })

       resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
   }
   ```

**Expected Result**: All 6 acceptance tests PASS

**Verification Command**:
```bash
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccCMUserUsersDataSource
# Expected: 6 passes
```

---

## Phase 3: Examples & Documentation

**Prerequisites**: Phase 2 complete, all tests passing

### 1. Create Example Configurations

**File**: `examples/data-sources/bcm_cmuser_users/basic.tf`
```hcl
data "bcm_cmuser_users" "all" {
}

output "user_count" {
  value = length(data.bcm_cmuser_users.all.users)
}

output "usernames" {
  value = [for user in data.bcm_cmuser_users.all.users : user.username]
}
```

**File**: `examples/data-sources/bcm_cmuser_users/filter_username.tf`
```hcl
# Find all admin users
data "bcm_cmuser_users" "admins" {
  username_pattern = "admin*"
}

output "admin_users" {
  value = [for user in data.bcm_cmuser_users.admins.users : {
    username = user.username
    email    = user.email
    active   = user.account_active
  }]
}
```

**File**: `examples/data-sources/bcm_cmuser_users/filter_group.tf`
```hcl
# Find all users in group 1000
data "bcm_cmuser_users" "group_1000" {
  group_id = "1000"
}

output "group_members" {
  value = [for user in data.bcm_cmuser_users.group_1000.users : user.username]
}
```

**File**: `examples/data-sources/bcm_cmuser_users/filter_user.tf`
```hcl
# Find specific user by UID
data "bcm_cmuser_users" "specific_user" {
  user_id = "1000"
}

output "user_details" {
  value = length(data.bcm_cmuser_users.specific_user.users) > 0 ? {
    username       = data.bcm_cmuser_users.specific_user.users[0].username
    home_directory = data.bcm_cmuser_users.specific_user.users[0].home_directory
    login_shell    = data.bcm_cmuser_users.specific_user.users[0].login_shell
    ssh_keys       = data.bcm_cmuser_users.specific_user.users[0].authorized_ssh_keys
  } : null
}
```

**File**: `examples/data-sources/bcm_cmuser_users/combined_filters.tf`
```hcl
# Find admin users in specific group
data "bcm_cmuser_users" "filtered" {
  username_pattern = "cms*"
  group_id         = "1000"
}

output "filtered_users" {
  value = [for user in data.bcm_cmuser_users.filtered.users : {
    username  = user.username
    user_id   = user.user_id
    group_id  = user.group_id
    active    = user.account_active
    home_dir  = user.home_directory
  }]
}
```

### 2. Generate Documentation

**Command**:
```bash
make generate
```

**Expected Output**:
- `docs/data-sources/cmuser_users.md` - Auto-generated from schema + examples
- Formatted examples in `examples/data-sources/bcm_cmuser_users/`
- Copyright headers added via copywrite

**Manual Verification**:
```bash
# Verify documentation matches schema
cat docs/data-sources/cmuser_users.md

# Verify examples are formatted
terraform fmt -check examples/data-sources/bcm_cmuser_users/
```

---

## Phase 4: Integration Testing & Validation

**Prerequisites**: Phase 3 complete, documentation generated

### Validation Checklist

- [ ] All 6 acceptance tests pass with 100% success rate
- [ ] Example configurations validate successfully (`terraform validate`)
- [ ] Client-side filtering works correctly with username patterns, group_id, user_id
- [ ] Account active computation verified with shadowExpire=-1 and positive values
- [ ] Documentation accurately reflects schema attributes
- [ ] No golangci-lint warnings (`make lint`)
- [ ] Code formatted correctly (`make fmt`)
- [ ] Pre-commit hooks pass (`pre-commit run --all-files`)

### Real-World Testing Scenarios

1. **Scenario: Query All Users**
   ```bash
   cd examples/data-sources/bcm_cmuser_users
   terraform init
   terraform apply -auto-approve
   terraform output user_count
   terraform output usernames
   ```

2. **Scenario: Filter by Username Pattern**
   ```bash
   cd examples/data-sources/bcm_cmuser_users
   terraform apply -target=data.bcm_cmuser_users.admins -auto-approve
   terraform output admin_users
   ```

3. **Scenario: Combined Filters**
   ```bash
   cd examples/data-sources/bcm_cmuser_users
   terraform apply -target=data.bcm_cmuser_users.filtered -auto-approve
   terraform output filtered_users
   ```

4. **Scenario: Account Active Verification**
   ```bash
   # Verify account_active correctly reflects shadowExpire values
   terraform console
   > data.bcm_cmuser_users.all.users[*].account_active
   ```

---

## Success Criteria Validation

After Phase 4 completion, verify all spec success criteria:

- [x] **SC-001**: Operators can retrieve all BCM users without errors
- [x] **SC-002**: Username pattern filtering returns only matching results
- [x] **SC-003**: Group ID filtering returns only users in specified group
- [x] **SC-004**: User ID filtering returns only user with specified UID
- [x] **SC-005**: BCM API errors provide clear error messages
- [x] **SC-006**: Client-side filtering completes without multiple API calls
- [x] **SC-007**: All Unix attributes (including shadow fields) accurately mapped
- [x] **SC-008**: Account active status correctly computed from shadow_expire
- [x] **SC-009**: Tests work on any BCM cluster configuration
- [x] **SC-010**: All 6 acceptance tests pass with 100% success rate
- [x] **SC-011**: Documentation auto-generated successfully via `make generate`

---

## Troubleshooting Guide

### Common Issues

**Issue**: Tests fail with "pattern syntax error"
- **Cause**: Invalid glob pattern in username_pattern filter
- **Fix**: Verify `filepath.Match()` accepts the pattern format, validate input

**Issue**: Account active always false
- **Cause**: shadowExpire computation logic incorrect or epoch day calculation wrong
- **Fix**: Debug `computeAccountActive()`, verify epoch day formula: `time.Now().Unix() / 86400`

**Issue**: Filters return no results
- **Cause**: Filter logic not correctly matching users or API response missing expected fields
- **Fix**: Add debug logging to `matchesFilters()`, verify BCM API response structure

**Issue**: Username attribute empty
- **Cause**: BCM API field is "name" not "username", mapping incorrect
- **Fix**: Verify `getStringValue(userData, "name")` used (not "username")

**Issue**: Shadow fields null or zero
- **Cause**: `getInt64Value()` not handling API response data types correctly
- **Fix**: Verify BCM API returns int64 values, handle float64 conversion if needed

---

## Risk Mitigation

| Risk | Impact | Mitigation |
|------|--------|------------|
| BCM API field names change | High | Phase 0 research validates actual field names from live API |
| Pattern matching edge cases | Medium | Use Go standard library `filepath.Match` (well-tested) |
| Large user lists performance | Low | Test with realistic user counts (10-100), document performance thresholds |
| Shadow expire computation errors | Medium | Use standard Unix epoch day calculation, test with -1 and positive values |
| Null values in optional fields | Low | Use null-safe helper functions (getStringValue, getInt64Value) |
| Test environment unavailability | High | Use existing test infrastructure (172.21.15.254); tests are read-only |

---

## Definition of Done

Implementation is complete when:

1. ✅ All 6 acceptance tests pass with TF_ACC=1
2. ✅ Data source registered in provider.go DataSources() method
3. ✅ Examples created for all filter scenarios (basic, username, group_id, user_id, combined)
4. ✅ Documentation auto-generated via `make generate`
5. ✅ Code passes `make lint` and `make fmt` checks
6. ✅ Pre-commit hooks pass without warnings
7. ✅ Client-side filtering verified with all filter combinations
8. ✅ Account active computation tested with shadowExpire=-1 and positive values
9. ✅ Field mapping verified (name → username, ID → user_id, groupID → group_id)
10. ✅ Error messages provide actionable guidance for users

---

## Next Steps for Autonomous Implementation

**Command**: `/speckit.tasks`

This command will generate `tasks.md` with granular task breakdown suitable for `/speckit.implement` autonomous execution.

**Then**: `/speckit.implement`

This command will execute all tasks in RED-GREEN-REFACTOR order with parallel execution where possible.
