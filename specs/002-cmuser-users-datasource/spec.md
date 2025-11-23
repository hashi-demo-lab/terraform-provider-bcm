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

### User Story 3 - Filter by Role (Priority: P3)

As a Terraform operator, I need to filter users by role to identify users with specific permission levels for access management and compliance reporting.

**Why this priority**: Role filtering enables security and compliance use cases, such as auditing admin users, generating reports of privileged accounts, or conditionally creating resources based on user roles.

**Independent Test**: Can be tested by querying users with a specific role filter and verifying only users with that role are returned. Delivers value for security audits and role-based access management.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with admin and regular users, **When** I filter by role "admin", **Then** the data source returns only users with admin role
2. **Given** a BCM cluster, **When** I filter by role that has no users, **Then** the data source returns an empty list without errors
3. **Given** a BCM cluster with multiple role types, **When** I filter by a specific role, **Then** all returned users have exactly that role

---

### User Story 4 - Filter by Enabled Status (Priority: P4)

As a Terraform operator, I need to filter users by enabled/disabled status to identify active vs inactive accounts for license management and security compliance.

**Why this priority**: Filtering by enabled status helps with account lifecycle management, identifying dormant accounts, and ensuring only active users are counted for licensing or compliance purposes.

**Independent Test**: Can be tested by creating enabled and disabled users and verifying the filter returns the correct subset. Delivers value for account hygiene and compliance reporting.

**Acceptance Scenarios**:

1. **Given** a BCM cluster with enabled and disabled users, **When** I filter by enabled=true, **Then** the data source returns only enabled users
2. **Given** a BCM cluster, **When** I filter by enabled=false, **Then** the data source returns only disabled users
3. **Given** a BCM cluster with only enabled users, **When** I filter by enabled=false, **Then** the data source returns an empty list

---

### Edge Cases

- What happens when the CMUser API is unavailable or returns an error during user retrieval?
- How does the data source handle users with missing optional fields (email, full_name, last_login)?
- What happens when multiple filters are applied simultaneously (e.g., username_pattern AND role AND enabled)?
- How does the system handle special characters or wildcards in username_pattern filters?
- What happens when the BCM API returns null or undefined values for user attributes?
- How does the data source behave when retrieving a very large number of users (performance)?

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST retrieve user data from the BCM CMUser API using the getUsers method via JSON-RPC
- **FR-002**: System MUST return a list of user objects with all available attributes (id, uuid, username, email, role, full_name, enabled, last_login, creation_time)
- **FR-003**: System MUST support optional client-side filtering by username_pattern using pattern matching
- **FR-004**: System MUST support optional client-side filtering by role using exact string matching
- **FR-005**: System MUST support optional client-side filtering by enabled status using boolean comparison
- **FR-006**: System MUST allow multiple filters to be applied simultaneously with AND logic
- **FR-007**: System MUST handle null or missing values for optional user attributes gracefully using null-safe helper functions
- **FR-008**: System MUST set the data source ID to a deterministic value (e.g., "users" or timestamp-based)
- **FR-009**: System MUST return an empty list when no users match the filter criteria without raising errors
- **FR-010**: System MUST use standard Terraform Plugin Framework types (types.String, types.Bool, types.Int64) for all attributes
- **FR-011**: System MUST log API requests and responses using tflog for debugging purposes
- **FR-012**: System MUST perform all filtering operations client-side after retrieving the full user list from the API
- **FR-013**: System MUST follow the existing null-safe helper function pattern (getStringValue, getBoolValue, getInt64Value) for data extraction

### Key Entities

- **User**: Represents a BCM user account with identity and access attributes
  - Mandatory: id, uuid, username
  - Optional: email, role, full_name, enabled, last_login, creation_time
  - Relationships: None (standalone entity)

- **Filter**: Represents client-side filtering criteria for user queries
  - Attributes: username_pattern (string), role (string), enabled (boolean)
  - Behavior: All filters use AND logic when multiple are specified

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can retrieve all BCM users in a single data source query without errors
- **SC-002**: Data source correctly filters users by username pattern, returning only matching results
- **SC-003**: Data source correctly filters users by role, returning only users with the specified role
- **SC-004**: Data source correctly filters users by enabled status, returning only enabled or disabled users
- **SC-005**: Data source handles BCM API errors gracefully, providing clear error messages to operators
- **SC-006**: Data source performs client-side filtering without requiring multiple API calls
- **SC-007**: All user attributes are accurately mapped from BCM API response to Terraform state
- **SC-008**: Data source works correctly on any BCM cluster configuration without hardcoded assumptions
- **SC-009**: Acceptance tests pass with 100% success rate using modern terraform-plugin-testing patterns
- **SC-010**: Data source documentation is auto-generated and includes practical examples

## Assumptions

### Technical Assumptions

- **AS-001**: The BCM CMUser API endpoint exists and is accessible at the configured BCM endpoint
- **AS-002**: The getUsers API method returns a JSON array of user objects
- **AS-003**: User objects follow a consistent schema with at minimum uuid and username fields
- **AS-004**: The BCM API does not support server-side filtering, requiring client-side filtering
- **AS-005**: Authentication is handled by the provider-level BCM client (cookie-based auth)
- **AS-006**: The CMUser service follows the same JSON-RPC pattern as other BCM services (CMDevice, CMPart, CMNet)

### Data Assumptions

- **AS-007**: Username patterns use standard wildcard syntax (e.g., "*" for zero or more characters)
- **AS-008**: Role values are strings representing predefined BCM roles (e.g., "admin", "user", "operator")
- **AS-009**: The enabled field is a boolean indicating account active/inactive status
- **AS-010**: Last login and creation time are represented as Unix timestamps or ISO 8601 strings
- **AS-011**: Email addresses may be null or empty for users without configured email

### Testing Assumptions

- **AS-012**: Test environment has access to a BCM cluster with the CMUser API enabled
- **AS-013**: Tests can create and delete test users without affecting production data
- **AS-014**: The test BCM cluster has at least one user account available for testing
- **AS-015**: Username patterns support basic wildcard matching (prefix, suffix, contains)

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

**Expected Response Format**:
```json
[
  {
    "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
    "username": "admin",
    "email": "admin@example.com",
    "role": "admin",
    "fullName": "System Administrator",
    "enabled": true,
    "lastLogin": "2025-11-23T10:30:00Z",
    "creationTime": "2025-01-15T08:00:00Z"
  },
  {
    "uuid": "b2c3d4e5-f6a7-8901-bcde-f12345678901",
    "username": "developer",
    "email": "dev@example.com",
    "role": "user",
    "fullName": "Development User",
    "enabled": true,
    "lastLogin": "2025-11-22T15:45:00Z",
    "creationTime": "2025-03-10T09:30:00Z"
  }
]
```

### Field Mapping (BCM API → Terraform Schema)

| BCM API Field (camelCase) | Terraform Attribute (snake_case) | Type | Computed | Optional | Description |
|---------------------------|----------------------------------|------|----------|----------|-------------|
| uuid | uuid | types.String | Yes | No | Unique identifier for the user |
| uuid | id | types.String | Yes | No | Resource identifier (same as uuid) |
| username | username | types.String | Yes | No | User login name |
| email | email | types.String | Yes | No | User email address |
| role | role | types.String | Yes | No | User role/permission level |
| fullName | full_name | types.String | Yes | No | User's full display name |
| enabled | enabled | types.Bool | Yes | No | Whether the account is active |
| lastLogin | last_login | types.String | Yes | No | Timestamp of last login |
| creationTime | creation_time | types.String | Yes | No | Timestamp when account was created |

### Filter Schema (Client-Side)

| Filter Attribute | Type | Optional | Description |
|------------------|------|----------|-------------|
| username_pattern | types.String | Yes | Pattern to match usernames (supports wildcards) |
| role | types.String | Yes | Exact role name to filter by |
| enabled | types.Bool | Yes | Filter by enabled/disabled status |

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
   - Apply client-side filters (username_pattern, role, enabled)
   - Map each user to UserDataModel using helper functions
   - Set state with filtered users list
   - Generate deterministic ID

### Null-Safety Pattern

Use helper functions from `data_source_cmpart_softwareimages.go:399-431`:

```go
username := getStringValue(userData, "username")
email := getStringValue(userData, "email")
enabled := getBoolValue(userData, "enabled")
role := getStringValue(userData, "role")
```

### Testing Pattern

Follow modern statecheck patterns from `data_source_cmdevice_categories_test.go`:

1. **Basic Test**: Verify data source retrieves users without errors
2. **Filter by Username Test**: Create test user, verify username_pattern filter works
3. **Filter by Role Test**: Verify role filter returns correct users
4. **Filter by Enabled Test**: Verify enabled filter returns correct users
5. **Nested Attributes Test**: Verify all user attributes are correctly populated

Use `statecheck.ExpectKnownValue()` with `knownvalue.StringExact()`, `knownvalue.Bool()`, `knownvalue.NotNull()`

### Environment Portability

- Do NOT hardcode expected user counts or specific usernames
- Generate unique test users using `generateUniqueTestName("test-user")`
- Create test resources in the test itself, don't assume existing users
- Use dynamic assertions (`knownvalue.NotNull()`) for cluster-dependent data

## Scope

### In Scope

- Data source implementation for bcm_cmuser_users
- Client-side filtering by username_pattern, role, and enabled status
- Null-safe attribute mapping for all user fields
- Comprehensive acceptance tests using modern statecheck patterns
- Auto-generated documentation with examples
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

- User data may contain sensitive information (emails, roles, login times)
- The data source does not expose passwords or credentials
- Authentication is handled at the provider level (existing BCM auth pattern)
- Data source operates in read-only mode (no user modifications)
- Filtering operations do not expose sensitive data in error messages
- All API communication uses the provider's configured endpoint and credentials
