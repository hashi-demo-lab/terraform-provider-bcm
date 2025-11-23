# Specification Updates - 2025-11-23

## Overview
Updated spec.md to reflect the actual BCM CMUser API structure discovered during API exploration. All changes are based on real API response data from `sampleRest/cmuser-explore-output.json`.

## Major Changes

### 1. User Stories Updated
- **Story 3**: Changed from "Filter by Role" → "Filter by Group ID"
- **Story 4**: Changed from "Filter by Enabled Status" → "Query Unix User Attributes"

### 2. Field Mappings Corrected

**Added Fields** (Unix user attributes):
- `user_id` (Unix UID from API field "ID")
- `group_id` (Unix GID from API field "groupID")
- `common_name` (from "commonName")
- `surname` (from "surname")
- `home_directory` (from "homeDirectory")
- `login_shell` (from "loginShell")
- `authorized_ssh_keys` (from "authorizedSshKeys")
- `notes` (from "notes")
- `information` (from "information")
- `shadow_expire` (from "shadowExpire")
- `shadow_last_change` (from "shadowLastChange")
- `shadow_max` (from "shadowMax")
- `shadow_min` (from "shadowMin")
- `shadow_warning` (from "shadowWarning")
- `shadow_inactive` (from "shadowInactive")
- `account_active` (computed from shadowExpire)

**Removed Fields** (do not exist in API):
- `role` (no such field)
- `enabled` (no such field)
- `full_name` (replaced by common_name + surname)
- `last_login` (no such field)
- `creation_time` (no such field)

**Field Name Mapping Correction**:
- BCM API uses "name" field → maps to Terraform attribute "username"

### 3. Filter Schema Updated

**Updated Filters**:
- ✅ KEPT: `username_pattern` (filter by name field)
- ❌ REMOVED: `role` (field doesn't exist)
- ❌ REMOVED: `enabled` (field doesn't exist)
- ✅ ADDED: `group_id` (filter by groupID field)
- ✅ ADDED: `user_id` (filter by ID field)

### 4. Functional Requirements Updated
- **FR-002**: Updated to include all Unix user attributes
- **FR-004**: Changed from role filter to group_id filter
- **FR-005**: Changed from enabled filter to user_id filter
- **FR-014**: NEW - Compute account_active from shadowExpire
- **FR-015**: NEW - Map BCM "name" field to "username" attribute

### 5. Success Criteria Updated
- **SC-003**: Changed from role filtering to group_id filtering
- **SC-004**: Changed from enabled filtering to user_id filtering
- **SC-007**: Updated to include Unix/shadow attributes
- **SC-008**: NEW - Verify account_active computation
- **SC-011**: NEW - Documentation for Unix user management

### 6. Assumptions Updated
- **AS-003**: Clarified Unix user schema expected
- **AS-008**: Changed from role values to Unix shadow conventions
- **AS-009**: Changed from enabled field to shadowExpire semantics
- **AS-010**: Changed from lastLogin/creationTime to ID/groupID types
- **AS-011**: Updated to reflect actual nullable fields
- **AS-012**: NEW - authorizedSshKeys multi-line handling
- **AS-014**: Updated testing assumption for read-only queries
- **AS-016**: Renumbered (was AS-015)

### 7. Implementation Guidelines Enhanced

**Added Sections**:
- Account Active Computation pattern
- Field name mapping examples (name → username, ID → user_id)
- Shadow password field handling
- Unix user attribute testing

### 8. API Contract Updated
- Replaced hypothetical response with actual API response
- Added note about omitted internal BCM fields
- Updated field mapping table with 20 attributes (was 9)
- Updated filter schema

### 9. Scope Updated
- In Scope: Added shadow password attributes, account_active computation
- Updated filter list to reflect actual capabilities

### 10. Security Considerations Enhanced
- Added SSH keys exposure note
- Updated sensitive data types (UIDs/GIDs, home directories)
- Clarified password field is empty in API

### 11. Clarifications Section Added
Encoded all autonomous decisions from `clarifications.md` into spec.md:
- Actual API structure
- Missing field handling
- Unix field exposure strategy
- Field name mapping rationale
- Name field handling (commonName/surname)
- Filter capabilities

## Files Modified
- `/workspace/specs/002-cmuser-users-datasource/spec.md` - Comprehensive updates

## Files Referenced
- `/workspace/specs/002-cmuser-users-datasource/clarifications.md` - Autonomous decisions
- `/workspace/sampleRest/cmuser-explore-output.json` - Actual API response
- `/workspace/sampleRest/cmuser-explore-api.py` - API exploration script

## Verification
All changes are grounded in actual API exploration data. No assumptions made beyond what the API actually returns.

## Next Steps
Specification is now ready for `/speckit.plan` to generate implementation plan.
