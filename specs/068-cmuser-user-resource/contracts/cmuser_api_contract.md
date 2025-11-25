# CMUser API Contract

**Feature Branch**: `068-cmuser-user-resource`
**Date**: 2025-11-26
**Status**: Draft - Requires Verification

## Overview

This document defines the expected BCM API contract for the CMUser service, used by the `bcm_cmuser_user` resource.

---

## Service Information

- **Service Name**: `cmuser`
- **Endpoint**: `{BCM_ENDPOINT}/json`
- **Authentication**: Cookie-based (`cm-login-token`)
- **Request Format**: JSON-RPC

---

## API Methods

### getUser

Retrieves a single user by username.

**Request**:
```json
{
  "service": "cmuser",
  "call": "getUser",
  "args": ["username"]
}
```

**Response (Success)**:
```json
{
  "uuid": "c792c8d3-3a5a-5003-bf6e-5bed0e59706f",
  "name": "username",
  "ID": "1000",
  "groupID": "1000",
  "baseType": "User",
  "childType": "",
  "modified": false,
  "to_be_removed": false,
  "revision": "",
  "password": "",
  "email": "user@example.com",
  "commonName": "Full Name",
  "surname": "Surname",
  "homeDirectory": "/home/username",
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

**Response (Not Found)**:
```json
null
```
or
```json
{}
```

---

### getUsers

Retrieves all users.

**Request**:
```json
{
  "service": "cmuser",
  "call": "getUsers"
}
```

**Response**:
```json
[
  {
    "uuid": "...",
    "name": "cmsupport",
    ...
  },
  {
    "uuid": "...",
    "name": "root",
    ...
  }
]
```

---

### addUser

Creates a new user.

**Request**:
```json
{
  "service": "cmuser",
  "call": "addUser",
  "args": [
    {
      "baseType": "User",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "uuid": "",
      "name": "newuser",
      "password": "SecurePassword123!",
      "ID": "",
      "groupID": "",
      "homeDirectory": "/home/newuser",
      "loginShell": "/bin/bash",
      "commonName": "New User",
      "surname": "",
      "email": "",
      "notes": "",
      "authorizedSshKeys": ""
    },
    false
  ]
}
```

**Response (Success)**:
```json
"c792c8d3-3a5a-5003-bf6e-5bed0e59706f"
```
or
```json
{
  "uuid": "c792c8d3-3a5a-5003-bf6e-5bed0e59706f",
  ...
}
```

**Response (Duplicate)**:
```json
{
  "error": "User already exists",
  "code": "DUPLICATE_ENTRY"
}
```

---

### updateUser

Updates an existing user.

**Request**:
```json
{
  "service": "cmuser",
  "call": "updateUser",
  "args": [
    {
      "baseType": "User",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "uuid": "c792c8d3-3a5a-5003-bf6e-5bed0e59706f",
      "name": "existinguser",
      "loginShell": "/bin/zsh",
      "notes": "Updated notes"
    },
    false
  ]
}
```

**Response (Success)**:
```json
true
```
or
```json
{
  "uuid": "...",
  ...
}
```

---

### removeUser

Deletes a user by username.

**Request**:
```json
{
  "service": "cmuser",
  "call": "removeUser",
  "args": ["username"]
}
```

**Response (Success)**:
```json
true
```

**Response (Not Found)**:
```json
{
  "error": "User not found",
  "code": "NOT_FOUND"
}
```

---

### validateUser

Validates user entity before create/update.

**Request**:
```json
{
  "service": "cmuser",
  "call": "validateUser",
  "args": [
    {
      "baseType": "User",
      "name": "newuser",
      "password": "weak"
    },
    true
  ]
}
```

**Response (Validation Errors)**:
```json
[
  {
    "Field": "password",
    "Message": "Password does not meet complexity requirements",
    "ErrorCode": "WEAK_PASSWORD",
    "Severity": "ERROR",
    "EntityUUID": ""
  }
]
```

**Response (Valid)**:
```json
[]
```

---

## Field Mapping

### Terraform to BCM API

| Terraform Attribute | BCM API Field | Type | Direction |
|---------------------|---------------|------|-----------|
| username | name | string | read/write |
| password | password | string | write-only |
| uid | ID | string | read/write |
| gid | groupID | string | read/write |
| full_name | commonName | string | read/write |
| surname | surname | string | read/write |
| email | email | string | read/write |
| home_directory | homeDirectory | string | read/write |
| shell | loginShell | string | read/write |
| notes | notes | string | read/write |
| authorized_ssh_keys | authorizedSshKeys | string | read/write |
| shadow_expire | shadowExpire | int64 | read-only |
| shadow_last_change | shadowLastChange | int64 | read-only |
| shadow_max | shadowMax | int64 | read/write |
| shadow_min | shadowMin | int64 | read/write |
| shadow_warning | shadowWarning | int64 | read/write |
| shadow_inactive | shadowInactive | int64 | read/write |

---

## Verification Checklist

Run these verifications against live BCM API:

- [ ] `getUser("cmsupport")` returns user entity
- [ ] `getUser("nonexistent")` returns null or empty
- [ ] `addUser(entity, false)` creates user and returns UUID
- [ ] `updateUser(entity, false)` modifies user attributes
- [ ] `removeUser("testuser")` deletes user
- [ ] `validateUser(entity, true)` returns validation array
- [ ] Password field is always empty in getUser response
- [ ] UID/GID auto-assigned when not specified
- [ ] Duplicate username returns appropriate error

---

## Error Codes

| Code | Description | HTTP-like |
|------|-------------|-----------|
| DUPLICATE_ENTRY | Username already exists | 409 Conflict |
| NOT_FOUND | User not found | 404 Not Found |
| WEAK_PASSWORD | Password doesn't meet policy | 422 Unprocessable |
| INVALID_FORMAT | Field format invalid | 400 Bad Request |
| NOT_NULL | Required field missing | 400 Bad Request |

---

## Notes

1. **Password Handling**: BCM API never returns password values. The `password` field is always empty in responses. The provider must preserve password in Terraform state.

2. **UUID Generation**: For create operations, UUID can be:
   - Empty string (BCM generates)
   - Pre-generated UUID (client provides)

3. **Eventual Consistency**: After create/update, allow brief delay before read to ensure consistency.

4. **Force Parameter**: Second argument in add/update controls whether to proceed despite warnings.
