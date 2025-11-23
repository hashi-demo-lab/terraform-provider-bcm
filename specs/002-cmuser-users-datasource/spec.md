# Feature Specification: BCM CMUser Users Data Source

**Feature Branch**: `002-cmuser-users-datasource`
**Created**: 2025-11-23
**Status**: Draft
**Input**: User description: "Implement a new data source for querying BCM users via the CMUser service"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Query All Users (Priority: P1)

As a Terraform operator, I need to retrieve a list of all BCM users to understand who has access to the cluster and incorporate user information into infrastructure-as-code configurations.

**Why this priority**: This is the foundational capability that enables basic user discovery and inventory management. Without this, the data source provides no value.

**Independent Test**: Can be fully tested by configuring the data source with no filters and verifying it returns a list of users with expected attributes (username, email, role, etc.). Delivers immediate value for user inventory and audit purposes.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with multiple users, **When** I query the bcm_cmuser_users data source without filters, **Then** the data source returns all users with their complete attributes (id, uuid, username, email, role, full_name, enabled, last_login, creation_time)
2. **Given** a BCM cluster with no users, **When** I query the bcm_cmuser_users data source, **Then** the data source returns an empty list without errors
3. **Given** a BCM cluster, **When** I query the bcm_cmuser_users data source, **Then** each user object contains computed values for all mandatory fields (id, uuid, username)

---

### User Story 2 - Filter by Username Pattern (Priority: P2)

As a Terraform operator, I need to filter users by username pattern to selectively retrieve specific users or groups of users matching a naming convention.

**Why this priority**: Filtering by username is a common requirement for managing specific user groups or finding users matching organizational naming patterns (e.g., "admin*", "dev-*", "test-*").

**Independent Test**: Can be tested by creating users with known username patterns and verifying the data source returns only matching users. Delivers value for targeted user management and conditional resource configuration.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with users named "admin", "admin-backup", and "developer", **When** I filter by username_pattern "admin*", **Then** the data source returns only "admin" and "admin-backup"
2. **Given** a BCM cluster with mixed users, **When** I filter by username_pattern matching no users, **Then** the data source returns an empty list without errors
3. **Given** a BCM cluster, **When** I filter by exact username "admin", **Then** the data source returns only the user with username "admin"

---

### User Story 3 - Filter by Group ID (Priority: P3)

As a Terraform operator, I need to filter users by Unix group ID to identify users belonging to specific groups for access management and organizational reporting.

**Why this priority**: Group-based filtering enables organizational use cases, such as querying all users in a specific group, generating reports of group membership, or conditionally creating resources based on group assignments.

**Independent Test**: Can be tested by querying users with a specific group_id filter and verifying only users with that group ID are returned. Delivers value for group-based access management and organizational queries.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with users in different groups, **When** I filter by group_id "1000", **Then** the data source returns only users with groupID "1000"
2. **Given** a BCM cluster, **When** I filter by group_id that has no users, **Then** the data source returns an empty list without errors
3. **Given** a BCM cluster with multiple groups, **When** I filter by a specific group_id, **Then** all returned users have exactly that groupID

---

### User Story 4 - Query Unix User Attributes (Priority: P4)

As a Terraform operator, I need to access Unix-specific user attributes (home directory, login shell, SSH keys, shadow password fields) to manage user environments and understand account lifecycle.

**Why this priority**: Unix user attributes are essential for infrastructure management, SSH key distribution, shell environment configuration, and account expiration monitoring.

**Independent Test**: Can be tested by querying users and verifying Unix attributes are correctly populated (home_directory, login_shell, authorized_ssh_keys, shadow fields). Delivers value for Unix user management and account lifecycle tracking.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with Unix users, **When** I query the bcm_cmuser_users data source, **Then** each user includes home_directory, login_shell, and shadow password fields
2. **Given** a BCM cluster with users having SSH keys, **When** I query users, **Then** authorized_ssh_keys attribute contains the SSH public keys
3. **Given** a BCM cluster with expired accounts, **When** I query users, **Then** the account_active computed field correctly reflects account expiration status based on shadow_expire

---

### Edge Cases

- What happens when the CMUser API is unavailable or returns an error during user retrieval?
- How does the data source handle users with missing optional fields (email, common_name, notes)?
- What happens when multiple filters are applied simultaneously (e.g., username_pattern AND group_id AND user_id)?
- How does the system handle special characters or wildcards in username_pattern filters?
- What happens when the BCM API returns null or undefined values for user attributes?
- How does the data source behave when retrieving a very large number of users (performance)?
- How does account_active computation handle special values like shadowExpire=-1 (never expires)?
- What happens when shadow password fields are null or missing?
- How does the data source handle multi-line SSH keys in authorizedSshKeys field?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST retrieve user data from the BCM CMUser API using the getUsers method via JSON-RPC
- **FR-002**: System MUST return a list of user objects with all available Unix user attributes (id, uuid, username, user_id, group_id, email, common_name, surname, home_directory, login_shell, authorized_ssh_keys, shadow password fields)
- **FR-003**: System MUST support optional client-side filtering by username_pattern using pattern matching against the name field
- **FR-004**: System MUST support optional client-side filtering by group_id using exact string matching
- **FR-005**: System MUST support optional client-side filtering by user_id using exact string matching
- **FR-006**: System MUST allow multiple filters to be applied simultaneously with AND logic
- **FR-007**: System MUST handle null or missing values for optional user attributes gracefully using null-safe helper functions
- **FR-008**: System MUST set the data source ID to a deterministic value (e.g., "users" or timestamp-based)
- **FR-009**: System MUST return an empty list when no users match the filter criteria without raising errors
- **FR-010**: System MUST use standard Terraform Plugin Framework types (types.String, types.Bool, types.Int64) for all attributes
- **FR-011**: System MUST log API requests and responses using tflog for debugging purposes
- **FR-012**: System MUST perform all filtering operations client-side after retrieving the full user list from the API
- **FR-013**: System MUST follow the existing null-safe helper function pattern (getStringValue, getBoolValue, getInt64Value) for data extraction
- **FR-014**: System MUST compute account_active boolean attribute from shadowExpire field (active = shadowExpire == -1 OR shadowExpire > current_epoch_day)
- **FR-015**: System MUST map BCM API field name to Terraform attribute username for consistency with operator expectations

### Key Entities

- **User**: Represents a BCM Unix user account with identity and system attributes
  - Mandatory: id, uuid, username (mapped from name field)
  - Identity: user_id (Unix UID), group_id (Unix GID)
  - Profile: email, common_name, surname, notes, information
  - Unix Environment: home_directory, login_shell, authorized_ssh_keys
  - Shadow Password: shadow_expire, shadow_last_change, shadow_max, shadow_min, shadow_warning, shadow_inactive
  - Computed: account_active (derived from shadow_expire)
  - Relationships: None (standalone entity)

- **Filter**: Represents client-side filtering criteria for user queries
  - Attributes: username_pattern (string), group_id (string), user_id (string)
  - Behavior: All filters use AND logic when multiple are specified

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can retrieve all BCM users in a single data source query without errors
- **SC-002**: Data source correctly filters users by username pattern, returning only matching results
- **SC-003**: Data source correctly filters users by group_id, returning only users in the specified group
- **SC-004**: Data source correctly filters users by user_id, returning only the user with the specified Unix UID
- **SC-005**: Data source handles BCM API errors gracefully, providing clear error messages to operators
- **SC-006**: Data source performs client-side filtering without requiring multiple API calls
- **SC-007**: All Unix user attributes (including shadow password fields) are accurately mapped from BCM API response to Terraform state
- **SC-008**: Data source correctly computes account_active status based on shadow_expire field
- **SC-009**: Data source works correctly on any BCM cluster configuration without hardcoded assumptions
- **SC-010**: Acceptance tests pass with 100% success rate using modern terraform-plugin-testing patterns
- **SC-011**: Data source documentation is auto-generated and includes practical examples for Unix user management

## Assumptions

### Technical Assumptions

- **AS-001**: The BCM CMUser API endpoint exists and is accessible at the configured BCM endpoint
- **AS-002**: The getUsers API method returns a JSON array of user objects
- **AS-003**: User objects follow Unix user schema with at minimum uuid and name (username) fields
- **AS-004**: The BCM API does not support server-side filtering, requiring client-side filtering
- **AS-005**: Authentication is handled by the provider-level BCM client (cookie-based auth)
- **AS-006**: The CMUser service follows the same JSON-RPC pattern as other BCM services (CMDevice, CMPart, CMNet)

### Data Assumptions

- **AS-007**: Username patterns use standard wildcard syntax (e.g., "*" for zero or more characters)
- **AS-008**: Unix shadow password fields follow standard conventions (days since epoch, -1 for "never")
- **AS-009**: shadowExpire value of -1 means account never expires, positive values are days since Unix epoch (1970-01-01)
- **AS-010**: ID and groupID fields are string representations of Unix UID/GID values
- **AS-011**: Email addresses, notes, and information fields may be null or empty strings
- **AS-012**: authorizedSshKeys may contain multi-line strings with newline-separated public keys

### Testing Assumptions

- **AS-013**: Test environment has access to a BCM cluster with the CMUser API enabled
- **AS-014**: Tests can create and delete test users without affecting production data (or use read-only queries on existing users)
- **AS-015**: The test BCM cluster has at least one user account available for testing
- **AS-016**: Username patterns support basic wildcard matching (prefix, suffix, contains)

## API Contract

### BCM JSON-RPC Call

**Service**: `cmuser`
**Method**: `getUsers`
**Arguments**: None (returns all users, filtering done client-side)

**Request Example**:
```json
{
  "service": "cmuser",
  "call": "getUsers"
}
```

**Actual Response Format** (from API exploration):
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

**Note**: Fields omitted from Terraform schema (internal BCM use): `baseType`, `childType`, `modified`, `to_be_removed`, `revision`, `password`, `profile`, `homePage`, `createSshKey`, `disablePasswordSsh`, `homeDirOperation`, `writeSshProxyConfig`, `allowGPUWorkloadPowerProfiles`, `certSerialNumber`, `projectManager`, `extra_values`

### Field Mapping (BCM API → Terraform Schema)

| BCM API Field (camelCase) | Terraform Attribute (snake_case) | Type | Computed | Optional | Description |
|---------------------------|----------------------------------|------|----------|----------|-------------|
| uuid | uuid | types.String | Yes | No | Unique identifier for the user |
| uuid | id | types.String | Yes | No | Resource identifier (same as uuid) |
| name | username | types.String | Yes | No | User login name (BCM field is "name") |
| ID | user_id | types.String | Yes | No | Unix user ID (UID) |
| groupID | group_id | types.String | Yes | No | Unix group ID (GID) |
| email | email | types.String | Yes | No | User email address (may be empty) |
| commonName | common_name | types.String | Yes | No | User's common/display name |
| surname | surname | types.String | Yes | No | User's surname |
| homeDirectory | home_directory | types.String | Yes | No | Unix home directory path |
| loginShell | login_shell | types.String | Yes | No | Unix login shell |
| notes | notes | types.String | Yes | No | User notes |
| information | information | types.String | Yes | No | Additional user information |
| authorizedSshKeys | authorized_ssh_keys | types.String | Yes | No | SSH authorized public keys |
| shadowExpire | shadow_expire | types.Int64 | Yes | No | Account expiration (days since epoch, -1=never) |
| shadowLastChange | shadow_last_change | types.Int64 | Yes | No | Last password change (days since epoch) |
| shadowMax | shadow_max | types.Int64 | Yes | No | Maximum password age (days) |
| shadowMin | shadow_min | types.Int64 | Yes | No | Minimum password age (days) |
| shadowWarning | shadow_warning | types.Int64 | Yes | No | Password expiration warning (days) |
| shadowInactive | shadow_inactive | types.Int64 | Yes | No | Account inactive grace period (days) |
| N/A (computed) | account_active | types.Bool | Yes | No | Computed: account not expired (shadowExpire == -1 OR shadowExpire > current_epoch_day) |

### Filter Schema (Client-Side)

| Filter Attribute | Type | Optional | Description |
|------------------|------|----------|-------------|
| username_pattern | types.String | Yes | Pattern to match username field (supports wildcards like "admin*") |
| group_id | types.String | Yes | Exact Unix group ID to filter by |
| user_id | types.String | Yes | Exact Unix user ID to filter by |

## Implementation Guidelines

### Data Source Pattern

Follow the established pattern from `data_source_cmdevice_categories.go`:

1. Define data source struct with BCMClient field
2. Implement Metadata() to set TypeName: `bcm_cmuser_users`
3. Implement Schema() with filter attributes (optional) and users list (computed)
4. Implement Configure() to receive BCM client from provider
5. Implement Read() method:
   - Call `client.CallJSONRPC(ctx, "cmuser", "getUsers")`
   - Unmarshal response into `[]map[string]interface{}`
   - Apply client-side filters (username_pattern, group_id, user_id)
   - Map each user to UserDataModel using helper functions
   - Compute account_active from shadow_expire field
   - Set state with filtered users list
   - Generate deterministic ID

### Null-Safety Pattern

Use helper functions from `data_source_cmpart_softwareimages.go:399-431`:

```go
username := getStringValue(userData, "name")        // BCM field is "name", maps to "username"
userID := getStringValue(userData, "ID")            // BCM field is "ID", maps to "user_id"
groupID := getStringValue(userData, "groupID")
email := getStringValue(userData, "email")
commonName := getStringValue(userData, "commonName")
shadowExpire := getInt64Value(userData, "shadowExpire")
```

### Account Active Computation

```go
// Compute account_active from shadowExpire
accountActive := types.BoolValue(false)
if !shadowExpire.IsNull() {
    expireDays := shadowExpire.ValueInt64()
    if expireDays == -1 {
        // -1 means never expires
        accountActive = types.BoolValue(true)
    } else {
        // Compare to current day since Unix epoch
        currentEpochDay := time.Now().Unix() / 86400
        accountActive = types.BoolValue(expireDays > currentEpochDay)
    }
}
```

### Testing Pattern

Follow modern statecheck patterns from `data_source_cmdevice_categories_test.go`:

1. **Basic Test**: Verify data source retrieves users without errors, check Unix attributes populated
2. **Filter by Username Test**: Verify username_pattern filter works (query existing users)
3. **Filter by Group ID Test**: Verify group_id filter returns correct users
4. **Filter by User ID Test**: Verify user_id filter returns specific user
5. **Nested Attributes Test**: Verify all user attributes are correctly populated (shadow fields, SSH keys)
6. **Account Active Computation Test**: Verify account_active is correctly derived from shadow_expire

Use `statecheck.ExpectKnownValue()` with `knownvalue.StringExact()`, `knownvalue.Bool()`, `knownvalue.Int64Exact()`, `knownvalue.NotNull()`

### Environment Portability

- Do NOT hardcode expected user counts or specific usernames
- Generate unique test users using `generateUniqueTestName("test-user")`
- Create test resources in the test itself, don't assume existing users
- Use dynamic assertions (`knownvalue.NotNull()`) for cluster-dependent data

## Scope

### In Scope

- Data source implementation for bcm_cmuser_users
- Client-side filtering by username_pattern, group_id, and user_id
- Null-safe attribute mapping for all Unix user fields (including shadow password attributes)
- Computed attribute account_active derived from shadow_expire
- Comprehensive acceptance tests using modern statecheck patterns
- Auto-generated documentation with examples for Unix user management
- Integration with existing BCM provider authentication

### Out of Scope

- Server-side API filtering (BCM API does not support this)
- User resource CRUD operations (separate feature)
- Password management or user authentication
- Group or permission management
- User creation/modification/deletion capabilities
- Advanced filtering (regex patterns, date range filters, multi-value filters)
- Pagination or lazy loading for large user lists
- Caching of user data across Terraform runs

## Dependencies

### Provider Dependencies

- BCM provider client with authenticated session
- Standard BCM JSON-RPC API endpoint access
- CMUser service availability on target BCM cluster

### Implementation Dependencies

- Terraform Plugin Framework v1.16+
- terraform-plugin-testing v1.13.3+
- Existing helper functions (getStringValue, getBoolValue, getInt64Value)
- Test helper functions (createTestBCMClient, generateUniqueTestName)

### Documentation Dependencies

- tfplugindocs tool for documentation generation
- Example configurations in examples/data-sources/bcm_cmuser_users/

## Security Considerations

- User data may contain sensitive information (emails, Unix UIDs/GIDs, SSH keys, home directories)
- The data source does not expose passwords or credentials (password field is empty in API response)
- Authentication is handled at the provider level (existing BCM auth pattern)
- Data source operates in read-only mode (no user modifications)
- Filtering operations do not expose sensitive data in error messages
- All API communication uses the provider's configured endpoint and credentials
- SSH authorized keys are exposed as-is from the API for infrastructure management purposes

## Clarifications

### Session 2025-11-23

- Q: What is the actual structure of the BCM CMUser API response? → A: API returns Unix user objects with fields: name (not username), ID, groupID, commonName, surname, homeDirectory, loginShell, authorizedSshKeys, and shadow password fields (shadowExpire, shadowLastChange, shadowMax, shadowMin, shadowWarning, shadowInactive). Fields role, enabled, lastLogin, and creationTime do NOT exist in the API.
- Q: How should we handle the missing "role" and "enabled" fields from the original spec? → A: Remove role and enabled filters. Replace with group_id and user_id filters. Compute account_active from shadowExpire field (active = shadowExpire == -1 OR shadowExpire > current_epoch_day).
- Q: Should we expose all Unix user fields or filter them? → A: Expose comprehensive Unix user attributes for infrastructure management: user_id, group_id, home_directory, login_shell, authorized_ssh_keys, and all shadow password fields. Omit internal BCM fields (baseType, childType, modified, to_be_removed, revision).
- Q: How should the "name" field be mapped to Terraform attributes? → A: Map BCM API field "name" to Terraform attribute "username" for consistency with operator expectations.
- Q: How should we handle the separate commonName and surname fields? → A: Expose both common_name and surname as separate attributes; do not concatenate them (let users choose format).
- Q: What filtering capabilities should be provided? → A: username_pattern (wildcard matching), group_id (exact match), user_id (exact match). All filters use AND logic when multiple are specified.
