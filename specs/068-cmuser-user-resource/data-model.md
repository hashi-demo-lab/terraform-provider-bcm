# Data Model: BCM User Resource (bcm_cmuser_user)

**Feature Branch**: `068-cmuser-user-resource`
**Date**: 2025-11-26
**Status**: Draft

## Overview

This document defines the data model for the `bcm_cmuser_user` Terraform resource, mapping BCM API entities to Terraform schema attributes.

---

## Entity Definitions

### Primary Entity: User

Represents a BCM Unix user account.

```
+----------------------------------+
|             User                 |
+----------------------------------+
| PK: uuid (computed)              |
| UK: username (required)          |
+----------------------------------+
| password (sensitive, write-only) |
| uid (optional, auto-assigned)    |
| gid (optional, auto-assigned)    |
| full_name (optional)             |
| surname (optional)               |
| email (optional)                 |
| home_directory (optional)        |
| shell (optional)                 |
| notes (optional)                 |
| authorized_ssh_keys (optional)   |
| shadow_max (optional)            |
| shadow_min (optional)            |
| shadow_warning (optional)        |
| shadow_inactive (optional)       |
| shadow_expire (computed)         |
| shadow_last_change (computed)    |
| account_active (computed)        |
| force (optional, not persisted)  |
+----------------------------------+
```

---

## Attribute Specifications

### Identity Attributes

| Attribute | Type | Required | Computed | Mutable | Description |
|-----------|------|----------|----------|---------|-------------|
| id | string | No | Yes | No | Resource identifier (same as UUID) |
| uuid | string | No | Yes | No | BCM-assigned unique identifier |
| username | string | Yes | No | No | Unique login name (1-32 chars, alphanumeric + underscore, starts with letter) |

### Authentication Attributes

| Attribute | Type | Required | Computed | Mutable | Sensitive | Description |
|-----------|------|----------|----------|---------|-----------|-------------|
| password | string | Yes (create) | No | Yes | Yes | User password (write-only, never returned by API) |

### Unix Identity Attributes

| Attribute | Type | Required | Computed | Mutable | Default | Description |
|-----------|------|----------|----------|---------|---------|-------------|
| uid | int64 | No | No | No | Auto-assigned | Unix user ID (stored as string "ID" in BCM API) |
| gid | int64 | No | No | No | Auto-assigned | Primary group ID (stored as string "groupID" in BCM API) |
| home_directory | string | No | No | Yes | /home/{username} | User home directory path |
| shell | string | No | No | Yes | /bin/bash | Login shell path |

### Profile Attributes

| Attribute | Type | Required | Computed | Mutable | Description |
|-----------|------|----------|----------|---------|-------------|
| full_name | string | No | No | Yes | Display name (maps to commonName) |
| surname | string | No | No | Yes | User's last name |
| email | string | No | No | Yes | Email address |
| notes | string | No | No | Yes | Administrative notes |
| authorized_ssh_keys | string | No | No | Yes | SSH public keys (multi-line) |

### Shadow Password Attributes

| Attribute | Type | Required | Computed | Mutable | Default | Description |
|-----------|------|----------|----------|---------|---------|-------------|
| shadow_max | int64 | No | No | Yes | 99999 | Maximum password age in days |
| shadow_min | int64 | No | No | Yes | 0 | Minimum password age in days |
| shadow_warning | int64 | No | No | Yes | 7 | Password expiration warning days |
| shadow_inactive | int64 | No | No | Yes | 0 | Days after expiration before inactive |
| shadow_expire | int64 | No | Yes | No | - | Account expiration (days since epoch) |
| shadow_last_change | int64 | No | Yes | No | - | Last password change (days since epoch) |

### Computed/Derived Attributes

| Attribute | Type | Derivation | Description |
|-----------|------|------------|-------------|
| account_active | bool | shadow_expire > current_epoch_day OR shadow_expire == -1 | Whether account is active |

### Operation Attributes

| Attribute | Type | Required | Persisted | Description |
|-----------|------|----------|-----------|-------------|
| force | bool | No | No | Force operation even with validation warnings |

---

## State Transitions

### Create Operation

**Required Inputs**:
- username (unique, validated format)
- password (meets BCM password policy)

**Optional Inputs**:
- All profile and shadow attributes

**BCM API Call**:
```
service: cmuser
call: addUser
args: [entity, force]
```

**State After Create**:
- uuid: BCM-assigned
- id: Same as uuid
- All provided attributes stored
- Computed attributes populated from API response

### Read Operation

**Lookup Strategy**:
- Use `getUser(username)` for efficient direct lookup
- Fallback to `getUsers()` + filter for import by UUID

**BCM API Call**:
```
service: cmuser
call: getUser
args: [username]
```

**Special Handling**:
- Password: Never read from API (always empty), preserve from state
- account_active: Computed from shadow_expire

### Update Operation

**Mutable Fields**:
- password (optional change)
- full_name, surname, email, notes
- home_directory, shell
- authorized_ssh_keys
- shadow_max, shadow_min, shadow_warning, shadow_inactive

**Immutable Fields** (require replacement):
- username
- uid
- gid

**BCM API Call**:
```
service: cmuser
call: updateUser
args: [entity, force]
```

### Delete Operation

**BCM API Call**:
```
service: cmuser
call: removeUser
args: [username]
```

**Idempotent Behavior**:
- If user already deleted externally, treat as success

### Import Operation

**Import Identifier**: username (string)

**Process**:
1. Accept username from `terraform import bcm_cmuser_user.x username`
2. Set username attribute in state
3. Trigger Read() to populate remaining attributes
4. Note: password will be null (cannot be recovered)

---

## Validation Rules

### Username Validation

```
Pattern: ^[a-zA-Z][a-zA-Z0-9_]{0,31}$
Min Length: 1
Max Length: 32
Start: Must begin with letter
Characters: Alphanumeric and underscore only
```

### Shell Validation

```
Pattern: ^/.*
Format: Absolute path starting with /
```

### Home Directory Validation

```
Pattern: ^/.*
Format: Absolute path starting with /
```

### UID/GID Validation

```
Range: 0 - 65535
Type: Positive integer
Special: 0-999 typically reserved for system accounts
```

---

## BCM API Mapping

### Request Entity Structure

```json
{
  "baseType": "User",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": "",
  "uuid": "",
  "name": "username",
  "password": "password",
  "ID": "1001",
  "groupID": "1001",
  "commonName": "Full Name",
  "surname": "Surname",
  "email": "user@example.com",
  "homeDirectory": "/home/username",
  "loginShell": "/bin/bash",
  "notes": "Notes here",
  "authorizedSshKeys": "ssh-rsa AAAA...",
  "shadowMax": 99999,
  "shadowMin": 0,
  "shadowWarning": 7,
  "shadowInactive": 0
}
```

### Response Entity Structure

```json
{
  "uuid": "c792c8d3-3a5a-5003-bf6e-5bed0e59706f",
  "name": "username",
  "ID": "1001",
  "groupID": "1001",
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
  "notes": "Notes here",
  "authorizedSshKeys": "ssh-rsa AAAA...",
  "shadowExpire": 24837,
  "shadowInactive": 0,
  "shadowLastChange": 20405,
  "shadowMax": 99999,
  "shadowMin": 0,
  "shadowWarning": 7
}
```

---

## Relationships

### User to Groups (POST-MVP)

```
User 1 ---< GroupMembership >--- N Group
```

**Notes**:
- Group membership handling requires API verification
- May be separate API calls or embedded in user entity
- Documented as POST-MVP until verified

### User to Kubernetes Cluster (External)

```
User 1 ---< KubeClusterUser >--- N KubeCluster
```

**Notes**:
- BCM user is prerequisite for `cm-kubernetes-setup --add-user`
- Managed outside this resource (separate bcm_cmkube_user resource)

---

## Terraform Schema

### Resource Schema Definition

```go
schema.Schema{
    MarkdownDescription: "Manages a BCM Unix user account.",

    Attributes: map[string]schema.Attribute{
        "id": schema.StringAttribute{
            Computed:            true,
            MarkdownDescription: "Resource identifier (same as UUID)",
            PlanModifiers: []planmodifier.String{
                stringplanmodifier.UseStateForUnknown(),
            },
        },
        "uuid": schema.StringAttribute{
            Computed:            true,
            MarkdownDescription: "BCM-assigned unique identifier",
            PlanModifiers: []planmodifier.String{
                stringplanmodifier.UseStateForUnknown(),
            },
        },
        "username": schema.StringAttribute{
            Required:            true,
            MarkdownDescription: "User login name (1-32 chars, starts with letter)",
            Validators: []validator.String{
                stringvalidator.LengthBetween(1, 32),
                stringvalidator.RegexMatches(
                    regexp.MustCompile(`^[a-zA-Z][a-zA-Z0-9_]*$`),
                    "must start with letter, contain only alphanumeric and underscore",
                ),
            },
            PlanModifiers: []planmodifier.String{
                stringplanmodifier.RequiresReplace(),
            },
        },
        "password": schema.StringAttribute{
            Required:            true,
            Sensitive:           true,
            MarkdownDescription: "User password (write-only, not returned by API)",
        },
        // ... additional attributes follow same pattern
    },
}
```

---

## References

- Feature Spec: `/workspace/specs/068-cmuser-user-resource/spec.md`
- Implementation Plan: `/workspace/specs/068-cmuser-user-resource/plan.md`
- Existing Data Source: `/workspace/internal/provider/data_source_cmuser_users.go`
