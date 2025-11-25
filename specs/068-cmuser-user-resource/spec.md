# Feature Specification: BCM User Resource (bcm_cmuser_user)

**Feature Branch**: `068-cmuser-user-resource`
**Created**: 2025-11-26
**Status**: Draft
**GitHub Issue**: #68
**Input**: Implement `bcm_cmuser_user` resource for managing BCM Unix users via Terraform, enabling DGX BasePOD automation workflows that require user creation before Kubernetes cluster user setup.

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Create BCM User for Kubernetes Administration (Priority: P1)

As a DGX BasePOD administrator, I want to create a BCM user account via Terraform so that I can subsequently add that user to a Kubernetes cluster using `cm-kubernetes-setup --add-user`.

**Why this priority**: This is the primary use case driving this feature. The DGX BasePOD Deployment Guide explicitly requires creating a BCM user before adding them to Kubernetes. Without this resource, the full Kubernetes deployment workflow cannot be automated.

**Independent Test**: Can be fully tested by creating a user with username and password, then verifying the user exists in BCM via the data source `data.bcm_cmuser_users`.

**Acceptance Scenarios**:

1. **Given** no user exists with username "k8sadmin", **When** I apply a Terraform configuration with `bcm_cmuser_user.k8sadmin`, **Then** the user is created in BCM with the specified username and password.
2. **Given** a user "k8sadmin" was created by Terraform, **When** I run `terraform show`, **Then** the password field is NOT displayed (marked sensitive).
3. **Given** a user "k8sadmin" exists in BCM, **When** I run `terraform plan` with the same configuration, **Then** no changes are detected (idempotent).

---

### User Story 2 - Update User Attributes (Priority: P2)

As a BCM administrator, I want to modify user attributes (groups, shell, home directory) via Terraform so that I can manage user configurations as code and track changes in version control.

**Why this priority**: Updates are essential for ongoing user management but secondary to initial user creation.

**Independent Test**: Can be tested by creating a user, then modifying group membership and verifying the change is applied.

**Acceptance Scenarios**:

1. **Given** a user "testuser" exists with shell "/bin/bash", **When** I change the shell to "/bin/zsh" in the Terraform configuration and apply, **Then** the user's shell is updated in BCM.
2. **Given** a user "testuser" exists without group membership, **When** I add `groups = ["wheel", "docker"]` and apply, **Then** the user is added to those groups.
3. **Given** a user "testuser" has groups ["wheel", "docker"], **When** I remove "docker" from the groups list and apply, **Then** the user is removed from the "docker" group.

---

### User Story 3 - Import Existing User (Priority: P2)

As a BCM administrator with pre-existing users, I want to import existing BCM users into Terraform state so that I can manage them declaratively going forward.

**Why this priority**: Import functionality enables adoption of Terraform for existing BCM deployments without recreating users.

**Independent Test**: Can be tested by creating a user manually via BCM CLI/UI, then importing it into Terraform state.

**Acceptance Scenarios**:

1. **Given** a user "existinguser" exists in BCM (created outside Terraform), **When** I run `terraform import bcm_cmuser_user.existing existinguser`, **Then** the user is imported into Terraform state.
2. **Given** a user was imported, **When** I run `terraform plan`, **Then** no changes are detected if the configuration matches the imported state.
3. **Given** I try to import a non-existent user "nouser", **When** I run `terraform import`, **Then** an appropriate error message is returned.

---

### User Story 4 - Delete User (Priority: P2)

As a BCM administrator, I want to remove users via Terraform when they are no longer needed so that I can maintain proper access control and audit trails.

**Why this priority**: Deletion is part of complete lifecycle management but is less frequently used than creation and updates.

**Independent Test**: Can be tested by creating a user, then destroying it and verifying it no longer exists.

**Acceptance Scenarios**:

1. **Given** a user "tempuser" exists and is managed by Terraform, **When** I run `terraform destroy`, **Then** the user is removed from BCM.
2. **Given** a user "tempuser" was deleted, **When** I query `data.bcm_cmuser_users` with that username pattern, **Then** no matching users are returned.

---

### User Story 5 - Drift Detection (Priority: P3)

As a BCM administrator, I want Terraform to detect when user attributes are modified outside of Terraform so that I can reconcile configuration drift.

**Why this priority**: Drift detection ensures infrastructure-as-code integrity but is an advanced feature beyond basic CRUD.

**Independent Test**: Can be tested by modifying a user attribute directly via BCM API, then running terraform plan.

**Acceptance Scenarios**:

1. **Given** a user "driftuser" exists with shell "/bin/bash", **When** the shell is changed to "/bin/sh" via BCM API directly, **Then** `terraform plan` detects the drift and proposes to change shell back to "/bin/bash".
2. **Given** drift was detected, **When** I run `terraform apply`, **Then** the user's attributes are restored to the Terraform-defined state.

---

### Edge Cases

- What happens when a user is created with a username that already exists? The system should return an appropriate error from BCM API.
- What happens when trying to delete a user that is referenced by a Kubernetes cluster? BCM should handle this dependency validation.
- What happens when an invalid shell path is specified? BCM server-side validation should reject the configuration.
- What happens when groups list contains non-existent group names? BCM should validate and return an error.
- What happens when UID/GID conflicts with existing users? BCM should return an error if the UID/GID is already in use.
- What happens when home_directory path is invalid or inaccessible? BCM should validate the path format.
- What happens when password is empty or doesn't meet BCM password policy? BCM should validate and return appropriate error.

## Requirements *(mandatory)*

### Functional Requirements

#### CRUD Operations

- **FR-001**: System MUST create BCM users with at minimum a username and password via the CMUser API.
- **FR-002**: System MUST read user attributes from BCM using `cmuser.getUser(username)` for efficient direct lookup or `cmuser.getUsers()` with client-side filtering.
- **FR-003**: System MUST update mutable user attributes (groups, shell, home_directory, full_name, etc.) via the CMUser API.
- **FR-004**: System MUST delete users via the CMUser API when the Terraform resource is destroyed.
- **FR-005**: System MUST support importing existing BCM users into Terraform state using username as the import identifier.

#### Data Handling

- **FR-006**: System MUST mark the password attribute as sensitive to prevent display in logs, plan output, and state files.
- **FR-007**: System MUST NOT include password in Read responses (BCM API returns empty string for password field).
- **FR-008**: System MUST preserve password in state and only send to API during Create or when password changes.
- **FR-009**: System MUST handle optional attributes gracefully (null values should not be sent to API).

#### Validation

- **FR-010**: System MUST validate username format (alphanumeric with underscores, starting with letter, 1-32 characters).
- **FR-011**: System MUST validate shell path format (absolute path starting with /).
- **FR-012**: System MUST validate home_directory path format (absolute path starting with /).
- **FR-013**: System MUST validate UID/GID are positive integers within valid Unix range (0-65535).
- **FR-014**: System SHOULD leverage BCM server-side validation via `cmuser.validateUser` if available.

#### State Management

- **FR-015**: System MUST use UUID as the primary identifier in Terraform state (id = uuid).
- **FR-016**: System MUST detect drift for all mutable attributes (groups, shell, home_directory, full_name, notes).
- **FR-017**: System MUST NOT detect drift for password (API does not return password values).

### Key Entities

- **User**: Represents a BCM Unix user account with identity (username, uuid), authentication (password), Unix attributes (uid, gid, shell, home_directory), and metadata (full_name, notes, groups).
- **Group Membership**: Represents the association between a user and one or more Unix groups. Groups are referenced by name and managed as a list attribute on the user.

## BCM API Contract

### API Methods (Based on Existing Data Source Research)

| Operation | Service | Method | Arguments | Notes |
| --------- | ------- | ------ | --------- | ----- |
| List All  | cmuser  | getUsers | none | Returns array of user objects |
| Get One   | cmuser  | getUser | username | Direct lookup (may return null if not found) |
| Create    | cmuser  | addUser | entity | Creates user with entity payload |
| Update    | cmuser  | updateUser | entity | Updates user with modified entity |
| Delete    | cmuser  | removeUser | username | Removes user by username |
| Validate  | cmuser  | validateUser | entity | Pre-flight validation (if available) |

### User Entity Structure (from cmuser.getUsers response)

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

### Field Mapping (Terraform snake_case to BCM camelCase)

| Terraform Attribute | BCM API Field | Type | Required | Notes |
| ------------------- | ------------- | ---- | -------- | ----- |
| id | uuid | string | Computed | Same as UUID |
| uuid | uuid | string | Computed | BCM-assigned |
| username | name | string | Required | Unique identifier |
| password | password | string | Required | Sensitive, write-only |
| uid | ID | string | Optional | Unix UID (stored as string in BCM) |
| gid | groupID | string | Optional | Primary group ID |
| full_name | commonName | string | Optional | Display name |
| surname | surname | string | Optional | Last name |
| email | email | string | Optional | Email address |
| home_directory | homeDirectory | string | Optional | Default: /home/{username} |
| shell | loginShell | string | Optional | Default: /bin/bash |
| notes | notes | string | Optional | User notes |
| groups | (derived) | list(string) | Optional | Group memberships (requires separate handling) |
| authorized_ssh_keys | authorizedSshKeys | string | Optional | SSH public keys |
| shadow_expire | shadowExpire | int64 | Computed | Days since epoch |
| shadow_max | shadowMax | int64 | Optional | Password max age |
| shadow_min | shadowMin | int64 | Optional | Password min age |
| shadow_warning | shadowWarning | int64 | Optional | Warning days |
| shadow_inactive | shadowInactive | int64 | Optional | Inactive days |
| account_active | (computed) | bool | Computed | Derived from shadow_expire |
| creation_time | (derived) | int64 | Computed | Unix timestamp |

## Resource Schema

### Required Attributes

- **username** (string): Unique login name for the user. Must be 1-32 characters, alphanumeric plus underscores, starting with a letter.
- **password** (string, sensitive): User password. Required on create, write-only (not returned by API).

### Optional Attributes

- **uid** (int64): Unix user ID. If not specified, BCM auto-assigns next available UID.
- **gid** (int64): Primary group ID. If not specified, BCM creates a group with same name as username.
- **full_name** (string): User's full display name (maps to commonName).
- **surname** (string): User's last name.
- **email** (string): User's email address.
- **home_directory** (string): Home directory path. Default: `/home/{username}`.
- **shell** (string): Login shell. Default: `/bin/bash`.
- **notes** (string): Administrative notes about the user.
- **groups** (list of strings): Additional group memberships by name.
- **authorized_ssh_keys** (string): SSH authorized_keys content (multi-line).
- **shadow_max** (int64): Maximum password age in days. Default: 99999.
- **shadow_min** (int64): Minimum password age in days. Default: 0.
- **shadow_warning** (int64): Password expiration warning days. Default: 7.
- **shadow_inactive** (int64): Days after password expiration before account inactive. Default: 0.
- **force** (bool): Force operation even with validation warnings. Default: false.

### Computed Attributes

- **id** (string): Resource identifier, same as UUID.
- **uuid** (string): BCM-assigned unique identifier.
- **shadow_expire** (int64): Account expiration date in days since Unix epoch.
- **shadow_last_change** (int64): Last password change date in days since epoch.
- **account_active** (bool): Whether account is currently active (derived from shadow_expire).
- **creation_time** (int64): Unix timestamp when user was created.

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Operators can create a BCM user account with username and password in a single Terraform apply operation.
- **SC-002**: Password values are never displayed in plan output, logs, or state file content (marked as sensitive).
- **SC-003**: Existing BCM users can be imported into Terraform state using `terraform import` with the username.
- **SC-004**: Running `terraform plan` after successful `terraform apply` shows no changes (idempotent).
- **SC-005**: External modifications to user attributes (shell, groups, etc.) are detected by `terraform plan`.
- **SC-006**: All acceptance tests pass including Create, Read, Update, Delete, Import, and Drift Detection.
- **SC-007**: Provider documentation is auto-generated via `make generate` and includes complete examples.
- **SC-008**: Resource integrates with existing `data.bcm_cmuser_users` data source for validation and lookup.

## Assumptions

- BCM API methods `addUser`, `updateUser`, and `removeUser` exist and follow the same patterns as other BCM services (CMDevice, CMPart).
- Password is write-only in the API (set on create/update, but returned as empty string on read).
- Group memberships may require separate API calls or are managed as part of the user entity.
- BCM validates username uniqueness server-side and returns appropriate errors.
- The CMUser service follows the standard BCM entity structure with baseType, childType, uuid, modified, etc.
- UID and GID are stored as strings in the BCM API but represented as int64 in Terraform for validation.

## Out of Scope

- Kubernetes user management (`bcm_cmkube_user`) - tracked as separate feature.
- SSH key generation (createSshKey field) - users provide their own keys.
- Password policy enforcement beyond BCM server-side validation.
- Bulk user operations (creating multiple users in single resource).
- User quota management.
- Home directory creation/permissions management beyond BCM default behavior.

## Dependencies

- Existing `data.bcm_cmuser_users` data source implementation (`internal/provider/data_source_cmuser_users.go`).
- BCMClient with `CallJSONRPC` method supporting variadic args.
- BCM API access with authentication.

## Test Plan

### Acceptance Tests (Modern Patterns - terraform-plugin-testing v1.13.3+)

1. **TestAccCMUserUser_Basic**: Create user with minimal attributes (username, password), verify creation, destroy.
2. **TestAccCMUserUser_Complete**: Create user with all optional attributes, verify all values stored correctly.
3. **TestAccCMUserUser_Update**: Create user, update mutable fields (shell, groups, notes), verify changes.
4. **TestAccCMUserUser_Import**: Create user manually, import into Terraform, verify state matches.
5. **TestAccCMUserUser_Drift**: Create user, modify externally via API, verify drift detected.
6. **TestAccCMUserUser_DriftGroups**: Create user with groups, modify groups externally, verify drift detected.
7. **TestAccCMUserUser_Idempotent**: Create user, reapply same config, verify no changes detected.
8. **TestAccCMUserUser_InvalidUsername**: Attempt to create user with invalid username, verify error.
9. **TestAccCMUserUser_DuplicateUsername**: Attempt to create user with existing username, verify error.
10. **TestAccCMUserUser_PasswordSensitive**: Verify password not shown in plan output or logs.

### Test Configuration Pattern

```hcl
provider "bcm" {
  endpoint             = "%[1]s"
  username             = "%[2]s"
  password             = "%[3]s"
  insecure_skip_verify = true
}

resource "bcm_cmuser_user" "test" {
  username = "testuser-%[4]s"
  password = "TestP@ssw0rd123!"
  shell    = "/bin/bash"
  groups   = ["wheel"]
  notes    = "Created by Terraform acceptance test"
}
```

## Example Usage

### Basic User Creation

```hcl
resource "bcm_cmuser_user" "k8s_admin" {
  username = "k8sadmin"
  password = var.k8s_admin_password
}
```

### User with Full Configuration

```hcl
resource "bcm_cmuser_user" "developer" {
  username       = "developer01"
  password       = var.developer_password
  full_name      = "Developer One"
  email          = "dev01@example.com"
  home_directory = "/home/developer01"
  shell          = "/bin/zsh"
  groups         = ["wheel", "docker", "developers"]
  notes          = "DGX BasePOD developer account"

  authorized_ssh_keys = <<-EOT
    ssh-rsa AAAAB3Nza... user@example.com
    ssh-ed25519 AAAAC3Nza... user@laptop
  EOT
}
```

### Integration with Existing Resources

```hcl
# Create user for Kubernetes administration
resource "bcm_cmuser_user" "k8s_admin" {
  username = "k8sadmin"
  password = var.k8s_admin_password
  groups   = ["wheel", "docker"]
  shell    = "/bin/bash"
}

# Verify user creation via data source
data "bcm_cmuser_users" "verify" {
  depends_on = [bcm_cmuser_user.k8s_admin]

  username_pattern = "k8sadmin"
}

output "k8s_admin_uuid" {
  value = bcm_cmuser_user.k8s_admin.uuid
}

output "k8s_admin_verified" {
  value = length(data.bcm_cmuser_users.verify.users) > 0
}
```
