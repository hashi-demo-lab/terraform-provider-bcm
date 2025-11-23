# Clarifications for BCM CMUser Users Data Source

**Date**: 2025-11-23
**Status**: Autonomous Decisions Made

## API Research Results

### Confirmed API Method
- **Service**: `cmuser`
- **Method**: `getUsers` ✅ (CONFIRMED - returns user array)
- **Alternative methods tested**: All failed (listUsers, getUserList, getAllUsers)

### Actual BCM API Response Structure

The BCM CMUser API returns user objects with these **actual fields**:

```json
{
  "ID": "1000",
  "uuid": "c792c8d3-3a5a-5003-bf6e-5bed0e59706f",
  "name": "cmsupport",
  "email": "",
  "commonName": "cmsupport",
  "surname": "cmsupport",
  "groupID": "1000",
  "homeDirectory": "/home/cmsupport",
  "loginShell": "/bin/bash",
  "homePage": "",
  "information": "",
  "notes": "",
  "password": "",
  "profile": "",
  "baseType": "User",
  "childType": "",
  "modified": false,
  "to_be_removed": false,
  "revision": "",
  "shadowExpire": 24837,
  "shadowInactive": 0,
  "shadowLastChange": 20405,
  "shadowMax": 99999,
  "shadowMin": 0,
  "shadowWarning": 7,
  "authorizedSshKeys": "",
  "createSshKey": false,
  "disablePasswordSsh": false,
  "homeDirOperation": true,
  "writeSshProxyConfig": false,
  "allowGPUWorkloadPowerProfiles": false,
  "certSerialNumber": -1,
  "projectManager": null,
  "extra_values": null
}
```

## Autonomous Decisions Made

### Decision 1: Field Name Corrections
**Issue**: Original spec assumed `username`, `role`, `fullName`, `enabled`, `lastLogin`, `creationTime` fields
**Reality**: API uses `name`, `commonName`, `surname`, Unix shadow fields, no role/enabled/lastLogin/creationTime

**Decision**: Update field mappings to match actual API:
- `name` → `username` (Terraform attribute)
- `commonName` → `common_name` (additional attribute)
- `surname` → `surname` (additional attribute)
- `ID` → `user_id` (Unix user ID)
- `groupID` → `group_id` (Unix group ID)
- `shadowExpire` → `shadow_expire` (account expiration date)
- REMOVE: `role`, `enabled`, `lastLogin`, `creationTime` (not provided by API)

**Rationale**: Match reality of API, expose useful Unix user fields

### Decision 2: Filter Strategy Updates
**Issue**: Original spec included filters for `role` and `enabled` which don't exist in API
**Reality**: API provides `name`, `groupID`, `shadowExpire` for potential filtering

**Decision**: Update filter attributes to match available data:
- ✅ KEEP: `username_pattern` (filter by `name` field)
- ❌ REMOVE: `role` (field doesn't exist)
- ❌ REMOVE: `enabled` (field doesn't exist)
- ✅ ADD: `group_id` (filter by `groupID` field)
- ✅ ADD: `user_id` (filter by `ID` field)

**Rationale**: Provide useful filtering based on actual API fields (group-based queries are valuable for Unix user management)

### Decision 3: Account Status Determination
**Issue**: No direct `enabled` field
**Reality**: Unix shadow fields indicate account status

**Decision**: Add computed attribute `account_active` derived from `shadowExpire`:
- `account_active = (shadowExpire == -1 || shadowExpire > current_epoch_day)`
- This provides the "enabled" functionality users expect

**Rationale**: Unix `shadowExpire` field (days since epoch) determines if account is expired. Value of -1 means never expires.

### Decision 4: Full Name Handling
**Issue**: No single `fullName` field
**Reality**: Separate `commonName` and `surname` fields

**Decision**: Expose both `common_name` and `surname` as separate attributes
- Do NOT concatenate them (let users choose format)
- `commonName` often contains the display name already

**Rationale**: Different use cases need different name formats; preserve raw data

### Decision 5: Additional Useful Fields
**Issue**: API provides many Unix user management fields not in original spec
**Reality**: Shell, home directory, SSH keys, shadow password fields are valuable

**Decision**: Expose additional attributes for comprehensive user management:
- `home_directory` (string)
- `login_shell` (string)
- `authorized_ssh_keys` (string)
- `shadow_last_change` (int64) - last password change
- `shadow_max` (int64) - max password age
- `shadow_warning` (int64) - password expiration warning
- `notes` (string) - user notes
- `information` (string) - additional info

**Rationale**: These are standard Unix user attributes valuable for infrastructure management

### Decision 6: Omitted Fields
**Issue**: API returns internal BCM fields not useful for Terraform users
**Reality**: Fields like `baseType`, `childType`, `modified`, `to_be_removed`, `revision`, `password`, `profile`, `certSerialNumber`, `projectManager`, `extra_values`

**Decision**: OMIT these fields from Terraform schema
- Internal BCM fields: `baseType`, `childType`, `modified`, `to_be_removed`, `revision`
- Security sensitive: `password` (empty anyway)
- Rarely used: `homePage`, `profile`, `certSerialNumber`, `projectManager`, `extra_values`

**Rationale**: Keep schema focused on operationally useful attributes

## Updated API Contract

### Corrected Field Mapping (BCM API → Terraform Schema)

| BCM API Field (camelCase) | Terraform Attribute (snake_case) | Type | Description |
|---------------------------|----------------------------------|------|-------------|
| uuid | uuid | types.String | Unique identifier |
| uuid | id | types.String | Resource ID (same as uuid) |
| name | username | types.String | User login name |
| ID | user_id | types.String | Unix user ID |
| groupID | group_id | types.String | Unix group ID |
| email | email | types.String | Email address |
| commonName | common_name | types.String | Common/display name |
| surname | surname | types.String | Surname |
| homeDirectory | home_directory | types.String | Home directory path |
| loginShell | login_shell | types.String | Login shell |
| notes | notes | types.String | User notes |
| information | information | types.String | Additional information |
| authorizedSshKeys | authorized_ssh_keys | types.String | SSH authorized keys |
| shadowExpire | shadow_expire | types.Int64 | Account expiration (days since epoch, -1 = never) |
| shadowLastChange | shadow_last_change | types.Int64 | Last password change (days since epoch) |
| shadowMax | shadow_max | types.Int64 | Max password age (days) |
| shadowMin | shadow_min | types.Int64 | Min password age (days) |
| shadowWarning | shadow_warning | types.Int64 | Password expiration warning (days) |
| shadowInactive | shadow_inactive | types.Int64 | Inactive account grace period (days) |
| COMPUTED | account_active | types.Bool | Derived: account not expired |

### Updated Filter Schema

| Filter Attribute | Type | Description |
|------------------|------|-------------|
| username_pattern | types.String | Pattern to match username (supports wildcards) |
| group_id | types.String | Filter by Unix group ID (exact match) |
| user_id | types.String | Filter by Unix user ID (exact match) |

## Updated User Stories

### Updated Story 3 (was Filter by Role)
**New**: Filter by Group ID
- **Why**: Unix group-based filtering is valuable for managing user groups
- **Scenarios**: Query all users in group "1000", filter by specific group membership

### Removed Story 4 (was Filter by Enabled Status)
**Reason**: No direct `enabled` field in API
**Alternative**: Users can filter in Terraform using `account_active` computed field with `for` expressions

## Impact on Requirements

### Updated Functional Requirements
- **FR-002**: Updated to reflect actual available attributes
- **FR-004**: Changed from `role` filter to `group_id` filter
- **FR-005**: Removed (no enabled field)
- **FR-NEW**: Add computed `account_active` attribute based on `shadowExpire`

### Updated Assumptions
- **AS-008**: Removed (no role field)
- **AS-009**: Removed (no enabled field)
- **AS-010**: Removed (no lastLogin/creationTime fields)
- **AS-NEW**: Shadow password fields follow Unix conventions

## Testing Implications

### Test Updates Required
1. Remove tests for `role` and `enabled` filters
2. Add tests for `group_id` and `user_id` filters
3. Add test for `account_active` computed attribute
4. Verify all Unix user fields are correctly mapped
5. Test edge cases: empty email, expired accounts, special characters in SSH keys

### Environment Portability
- Generate unique test usernames
- Don't assume specific group IDs exist
- Handle varying shadow field values across BCM clusters

## Documentation Updates

### Example Updates Required
1. Basic example: Query all users, show Unix fields
2. Filter by username pattern
3. Filter by group_id
4. Use account_active for conditional logic
5. Access SSH keys and home directory info

## Summary of Changes

| Original Spec | Actual API | Decision |
|---------------|-----------|----------|
| `username` | `name` | Map `name` → `username` |
| `role` | NOT PRESENT | Remove role filter |
| `enabled` | NOT PRESENT | Compute `account_active` from `shadowExpire` |
| `fullName` | `commonName` + `surname` | Expose both separately |
| `lastLogin` | NOT PRESENT | Remove attribute |
| `creationTime` | NOT PRESENT | Remove attribute |
| N/A | Unix shadow fields | Add shadow password attributes |
| N/A | `homeDirectory`, `loginShell` | Add Unix user attributes |
| N/A | `authorizedSshKeys` | Add SSH key management |

## Next Steps

1. Update `spec.md` with corrected field mappings
2. Update user stories to reflect available filters
3. Remove references to non-existent fields (role, enabled, lastLogin, creationTime)
4. Add Unix user management use cases
5. Proceed to `/speckit.plan` with accurate API understanding
