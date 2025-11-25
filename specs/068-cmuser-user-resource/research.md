# Research: BCM User Resource (bcm_cmuser_user)

**Feature Branch**: `068-cmuser-user-resource`
**Date**: 2025-11-26
**Status**: NEEDS VERIFICATION

## Overview

This document captures research findings for the `bcm_cmuser_user` resource implementation. Research tasks are defined in `plan.md` Phase 0.

---

## R-001: CMUser API Methods for Resources

**Objective**: Verify existence and signatures of resource CRUD methods.

### Expected API Methods

Based on existing data source implementation and BCM API patterns:

| Method | Service | Arguments | Expected Response |
|--------|---------|-----------|-------------------|
| getUsers | cmuser | none | Array of user entities (VERIFIED - used by data source) |
| getUser | cmuser | username (string) | Single user entity or null |
| addUser | cmuser | entity (object), force (bool) | UUID string or entity object |
| updateUser | cmuser | entity (object), force (bool) | Success response or entity |
| removeUser | cmuser | username (string) | Success boolean |
| validateUser | cmuser | entity (object), isCreate (bool) | Validation array |

### Verification Status

**NEEDS VERIFICATION**: The following must be tested against live BCM API:

1. [ ] `getUser(username)` - Verify direct lookup works (vs list+filter)
2. [ ] `addUser(entity, force)` - Verify create method and response format
3. [ ] `updateUser(entity, force)` - Verify update method signature
4. [ ] `removeUser(username)` - Verify delete by username works
5. [ ] `validateUser(entity, isCreate)` - Verify validation method exists

### Research Notes

The existing `data_source_cmuser_users.go` uses:
```go
body, err := d.client.CallJSONRPC(ctx, "cmuser", "getUsers")
```

This confirms the `cmuser` service name and `getUsers` method. Based on BCM API patterns (CMPart, CMDevice), the resource methods should follow similar naming conventions.

---

## R-002: User Entity Structure

**Objective**: Confirm BCM entity structure for Create/Update operations.

### Known Entity Structure (from getUsers response)

```json
{
  "uuid": "c792c8d3-3a5a-5003-bf6e-5bed0e59706f",
  "name": "cmsupport",
  "ID": "1000",
  "groupID": "1000",
  "baseType": "User",
  "childType": "",
  "modified": false,
  "to_be_removed": false,
  "revision": "",
  "password": "",
  "email": "",
  "commonName": "cmsupport",
  "surname": "cmsupport",
  "homeDirectory": "/home/cmsupport",
  "loginShell": "/bin/bash",
  "notes": "",
  "information": "",
  "authorizedSshKeys": "",
  "shadowExpire": 24837,
  "shadowInactive": 0,
  "shadowLastChange": 20405,
  "shadowMax": 99999,
  "shadowMin": 0,
  "shadowWarning": 7,
  "homeDirOperation": true,
  "createSshKey": false,
  "disablePasswordSsh": false,
  "allowGPUWorkloadPowerProfiles": false,
  "writeSshProxyConfig": false
}
```

### Entity Structure for Create

Based on BCM patterns (CMPartSoftwareImage, CMDeviceCategory):

```json
{
  "baseType": "User",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "uuid": "",
  "name": "newuser",
  "password": "SecureP@ssw0rd!",
  "ID": "",
  "groupID": "",
  "homeDirectory": "/home/newuser",
  "loginShell": "/bin/bash",
  "commonName": "New User",
  "surname": "",
  "email": "",
  "notes": "",
  "authorizedSshKeys": ""
}
```

### Verification Status

**NEEDS VERIFICATION**:

1. [ ] Confirm minimum required fields for addUser
2. [ ] Verify UUID handling (empty for create, populated for update)
3. [ ] Confirm ID/groupID auto-assignment when empty
4. [ ] Test password field requirement on create

---

## R-003: Group Membership Handling

**Objective**: Determine how group memberships are managed.

### Hypothesis

Based on BCM API structure, groups may be handled in one of these ways:

1. **Embedded in user entity** (most likely)
   - `groups: ["wheel", "docker"]` as array field
   - Requires verification if this field exists

2. **Separate API calls** (less likely for BCM pattern)
   - `cmuser.addUserToGroup(username, groupname)`
   - `cmuser.removeUserFromGroup(username, groupname)`

3. **Primary group only** (possible limitation)
   - Only `groupID` field for primary group
   - Secondary groups managed outside user entity

### Verification Status

**NEEDS VERIFICATION**:

1. [ ] Check if `groups` field exists in user entity
2. [ ] Test creating user with group memberships
3. [ ] Verify group reference format (name vs ID)

### Research Notes

The existing data source model (`UserModel` in `data_source_cmuser_users.go`) does NOT include a `groups` field, suggesting:
- Groups may not be part of the user entity
- Groups might be a POST-MVP feature requiring separate API calls
- Need to verify BCM API documentation or test behavior

**Recommendation**: If groups are not in user entity, document as POST-MVP and exclude from initial implementation.

---

## R-004: UID/GID Auto-Assignment

**Objective**: Understand BCM behavior when UID/GID not specified.

### Expected Behavior

Based on standard Unix user management patterns:

1. **UID (ID field)**: If empty/omitted, BCM should auto-assign next available UID >= 1000
2. **GID (groupID field)**: If empty/omitted, BCM should:
   - Create a group with same name as username
   - Assign the new group's GID as primary group

### Verification Status

**NEEDS VERIFICATION**:

1. [ ] Create user without UID/GID
2. [ ] Verify auto-assigned values are in valid range
3. [ ] Confirm primary group creation behavior

---

## API Field Mapping Summary

### Terraform to BCM API Mapping

| Terraform (snake_case) | BCM API (camelCase) | Type | Notes |
|------------------------|---------------------|------|-------|
| username | name | string | Primary identifier |
| password | password | string | Write-only, never returned |
| uid | ID | string | Unix UID as string |
| gid | groupID | string | Primary GID as string |
| full_name | commonName | string | Display name |
| surname | surname | string | Last name |
| email | email | string | Email address |
| home_directory | homeDirectory | string | Home path |
| shell | loginShell | string | Login shell |
| notes | notes | string | User notes |
| authorized_ssh_keys | authorizedSshKeys | string | SSH keys (multi-line) |
| shadow_expire | shadowExpire | int64 | Account expiration |
| shadow_last_change | shadowLastChange | int64 | Password change date |
| shadow_max | shadowMax | int64 | Password max age |
| shadow_min | shadowMin | int64 | Password min age |
| shadow_warning | shadowWarning | int64 | Warning period |
| shadow_inactive | shadowInactive | int64 | Inactive period |

### Additional BCM Fields (Not in Terraform Schema)

These fields exist in BCM API but may not need Terraform exposure:

| BCM Field | Purpose | Include? |
|-----------|---------|----------|
| information | Additional info | Optional - similar to notes |
| homeDirOperation | Create home dir | POST-MVP |
| createSshKey | Generate SSH key | Out of scope |
| disablePasswordSsh | Disable password SSH | POST-MVP |
| allowGPUWorkloadPowerProfiles | GPU workload | POST-MVP |
| writeSshProxyConfig | SSH proxy config | POST-MVP |

---

## Research Execution Plan

### Phase 0 Research Tasks

To complete this research document:

1. **Create test script** to verify API methods:
```go
// test_cmuser_api.go - Manual API verification
func main() {
    client := createTestClient()

    // R-001: Verify methods exist
    testGetUser(client, "cmsupport")
    testAddUser(client, testEntity)
    testUpdateUser(client, modifiedEntity)
    testRemoveUser(client, "testuser")
    testValidateUser(client, testEntity)

    // R-002: Verify entity structure
    testMinimalCreate(client)

    // R-003: Verify group handling
    testGroupMembership(client)

    // R-004: Verify auto-assignment
    testUIDGIDAutoAssignment(client)
}
```

2. **Document findings** in this file with actual API responses
3. **Update plan.md** if research reveals deviations from assumptions

---

## Decisions Log

| Decision | Rationale | Date |
|----------|-----------|------|
| Use `getUser(username)` for Read | Efficient direct lookup, matches BCM API pattern | 2025-11-26 |
| Password write-only in state | BCM API returns empty string, must preserve from plan | 2025-11-26 |
| Groups POST-MVP if not in entity | Minimize scope until API behavior verified | 2025-11-26 |
| Import by username (not UUID) | More user-friendly, matches resource identifier | 2025-11-26 |

---

## Next Steps

1. Execute API verification tests (R-001 through R-004)
2. Update this document with actual API responses
3. Finalize field mappings based on verified behavior
4. Proceed to Phase 1 (Design & Contracts)

---

## References

- Existing data source: `/workspace/internal/provider/data_source_cmuser_users.go`
- BCM API patterns: `/workspace/internal/provider/bcm_client.go`
- Resource pattern: `/workspace/internal/provider/resource_cmpart_softwareimage.go`
- Feature spec: `/workspace/specs/068-cmuser-user-resource/spec.md`
