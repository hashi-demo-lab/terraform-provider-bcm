# Feature Specification: bcm_cmpart_softwareimage Resource

## Overview

This specification defines a Terraform resource for managing Nvidia BCM (Bright Cluster Manager) software images. Software images are operating system images used to provision DPU (Data Processing Unit) nodes with specific kernel configurations and kernel modules.

## User Stories

### US-1: Create Software Image
**As a** BCM cluster administrator
**I want to** create software images through Terraform
**So that** I can version-control and automate OS image provisioning

**Acceptance Criteria:**
- User can define software image with name and path
- User can specify kernel version and parameters
- User can configure Serial Over LAN settings
- User can define kernel modules list
- Terraform creates image via BCM API
- Image UUID is returned and stored in state

### US-2: Read Software Image
**As a** Terraform user
**I want to** Terraform to detect changes to software images
**So that** state drift is detected and reported

**Acceptance Criteria:**
- Terraform reads current image state from BCM
- Computed fields (UUID, creation time) are populated
- Changes made outside Terraform are detected
- Missing images trigger recreation

### US-3: Update Software Image
**As a** BCM administrator
**I want to** modify software image configuration
**So that** I can update kernel settings without recreating images

**Acceptance Criteria:**
- User can update kernel parameters
- User can add/remove kernel modules
- User can modify SOL settings
- Terraform applies changes via BCM API
- State reflects updated configuration

### US-4: Delete Software Image
**As a** BCM administrator
**I want to** remove software images when no longer needed
**So that** I can clean up unused images

**Acceptance Criteria:**
- `terraform destroy` removes image from BCM
- Image is fully deleted from backend
- State is cleared after successful deletion

### US-5: Import Existing Images
**As a** Terraform user
**I want to** import existing software images into Terraform state
**So that** I can manage pre-existing images with Terraform

**Acceptance Criteria:**
- User can run `terraform import bcm_cmpart_softwareimage.example <uuid>`
- Terraform fetches full image configuration
- Imported resource matches existing state
- Subsequent applies show no changes

## Functional Requirements

### FR-1: Resource Identity
- **FR-1.1:** Resource MUST have unique name within BCM cluster
- **FR-1.2:** Resource MUST have unique path within BCM cluster
- **FR-1.3:** Resource MUST be identified by server-generated UUID
- **FR-1.4:** UUID MUST be used for import operations

### FR-2: Kernel Configuration
- **FR-2.1:** Resource SHOULD support kernel version specification
- **FR-2.2:** Resource SHOULD support kernel boot parameters
- **FR-2.3:** Resource SHOULD support kernel output console configuration
- **FR-2.4:** Default kernel output console is "tty0"

### FR-3: Serial Over LAN (SOL)
- **FR-3.1:** Resource MUST support SOL enable/disable flag
- **FR-3.2:** Resource MUST support SOL port configuration
- **FR-3.3:** Resource MUST support SOL speed configuration (baud rate)
- **FR-3.4:** Resource MUST support SOL flow control flag
- **FR-3.5:** Default SOL settings: disabled, port "ttyS1", speed "115200", flow control enabled

### FR-4: Kernel Modules
- **FR-4.1:** Resource MUST support nested list of kernel modules
- **FR-4.2:** Each module MUST have name attribute
- **FR-4.3:** Each module MAY have parameters attribute
- **FR-4.4:** Modules list MAY be empty
- **FR-4.5:** Module order SHOULD be preserved in state

### FR-5: Filesystem Partitions
- **FR-5.1:** Resource MAY reference filesystem partition by UUID
- **FR-5.2:** Resource MAY reference boot partition by UUID
- **FR-5.3:** Partition references are optional

### FR-6: Metadata
- **FR-6.1:** Resource MAY have free-form notes field
- **FR-6.2:** Resource creation time MUST be computed (read-only)
- **FR-6.3:** Resource revision ID MUST be computed (read-only)
- **FR-6.4:** File operation status MUST be computed (read-only)

### FR-7: API Integration
- **FR-7.1:** Create operation MUST call `addSoftwareImage(entity, force)` API
- **FR-7.2:** Read operation MUST call `getSoftwareImages()` API and filter by UUID
- **FR-7.3:** Update operation MUST call discovered update API method (TBD in research)
- **FR-7.4:** Delete operation MUST call discovered delete API method (TBD in research)
- **FR-7.5:** Force parameter defaults to false, MAY be overridden by user

## Non-Functional Requirements

### NFR-1: Performance
- **NFR-1.1:** Resource creation SHOULD complete within 30 seconds
- **NFR-1.2:** Resource read operation SHOULD complete within 10 seconds
- **NFR-1.3:** Acceptance tests SHOULD complete within 120 minutes

### NFR-2: Reliability
- **NFR-2.1:** API errors MUST be reported with clear diagnostic messages
- **NFR-2.2:** Transient network failures SHOULD NOT leave inconsistent state
- **NFR-2.3:** Unique constraint violations MUST return user-friendly errors

### NFR-3: Security
- **NFR-3.1:** API credentials MUST NOT be logged
- **NFR-3.2:** Cookie-based authentication MUST be used for all API calls
- **NFR-3.3:** TLS certificate validation SHOULD be configurable

### NFR-4: Maintainability
- **NFR-4.1:** Code MUST follow Terraform Plugin Framework patterns
- **NFR-4.2:** Code MUST pass golangci-lint with no errors
- **NFR-4.3:** Code MUST include comprehensive acceptance tests
- **NFR-4.4:** Documentation MUST be auto-generated from schema

### NFR-5: Compatibility
- **NFR-5.1:** Resource MUST work with BCM API version used in test environment
- **NFR-5.2:** Resource MUST be compatible with Terraform 1.0+
- **NFR-5.3:** Resource MUST use Terraform Plugin Framework v1.16.1+

## API Contract

### Create Request
```json
POST /json
Cookie: cm-login-token=<token>
Content-Type: application/json

{
  "service": "CMPart",
  "call": "addSoftwareImage",
  "args": [
    {
      "baseType": "SoftwareImage",
      "childType": "",
      "name": "ubuntu-22.04-dpu",
      "path": "/cm/images/ubuntu-22.04-dpu",
      "kernelVersion": "5.15.0-58-generic",
      "kernelParameters": "quiet splash",
      "kernelOutputConsole": "tty0",
      "enableSOL": false,
      "SOLPort": "ttyS1",
      "SOLSpeed": "115200",
      "SOLFlowControl": true,
      "modules": [
        {
          "baseType": "KernelModule",
          "childType": "",
          "name": "nvidia-drm",
          "parameters": "modeset=1",
          "modified": true,
          "to_be_removed": false
        }
      ],
      "notes": "Ubuntu 22.04 for DPU nodes",
      "fspart": "00000000-0000-0000-0000-000000000000",
      "bootfspart": "00000000-0000-0000-0000-000000000000"
    },
    false
  ]
}
```

**Response:**
```json
"uuid-string"
```
OR
```json
{
  "uuid": "uuid-string",
  "name": "ubuntu-22.04-dpu",
  ...
}
```

### Read Request
```json
POST /json
Cookie: cm-login-token=<token>
Content-Type: application/json

{
  "service": "CMPart",
  "call": "getSoftwareImages"
}
```

**Response:**
```json
[
  {
    "uuid": "uuid-string",
    "baseType": "SoftwareImage",
    "childType": "",
    "name": "ubuntu-22.04-dpu",
    "path": "/cm/images/ubuntu-22.04-dpu",
    "kernelVersion": "5.15.0-58-generic",
    "kernelParameters": "quiet splash",
    "kernelOutputConsole": "tty0",
    "creationTime": 1700000000,
    "revisionID": 0,
    "fileOperationInProgress": false,
    "modules": [...],
    ...
  }
]
```

### Update Request ✅ CONFIRMED
```json
POST /json
Cookie: cm-login-token=<token>
Content-Type: application/json

{
  "service": "cmpart",
  "call": "updateSoftwareImage",
  "args": [softwareImage, force]
}
```

**Parameters**:
- `softwareImage` (object): Complete SoftwareImage entity with updated fields
- `force` (boolean): Force update flag (default: false)

### Delete Request ✅ CONFIRMED
```json
POST /json
Cookie: cm-login-token=<token>
Content-Type: application/json

{
  "service": "cmpart",
  "call": "removeSoftwareImage",
  "args": [uuid, removeData, removeAll, force]
}
```

**Parameters**:
- `uuid` (string): UUID of the SoftwareImage to remove
- `removeData` (boolean): Remove associated data files (default: false)
- `removeAll` (boolean): Remove all related entities (default: false)
- `force` (boolean): Force deletion flag (default: false)

## Data Model

### Resource Schema
```
bcm_cmpart_softwareimage {
  # Identity (Required)
  name: string (required, unique)
  path: string (required, unique, format: ^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$)

  # Identity (Computed)
  id: string (computed, immutable)
  uuid: string (computed, immutable)

  # Kernel Configuration (Optional)
  kernel_version: string (optional)
  kernel_parameters: string (optional)
  kernel_output_console: string (optional, default: "tty0")

  # Serial Over LAN (Optional)
  enable_sol: bool (optional, default: false)
  sol_port: string (optional, default: "ttyS1")
  sol_speed: enum (optional, default: "115200")
    values: ["115200", "57600", "38400", "19200", "9600", "4800", "2400", "1200"]
  sol_flow_control: bool (optional, default: true)

  # Filesystem Partitions (Optional)
  fspart: string (optional, uuid reference)
  bootfspart: string (optional, uuid reference)

  # Metadata (Optional/Computed)
  notes: string (optional)
  force: bool (optional, default: false)
  creation_time: int64 (computed)
  revision_id: int64 (computed)
  file_operation_in_progress: bool (computed)
  original_image: string (computed, uuid)
  parent_software_image: string (computed, uuid)

  # Nested Resources (Optional)
  modules: list(object) (optional, default: [])
    - name: string (required)
    - parameters: string (optional)
}
```

## Error Handling

### Error Scenarios

| Scenario | HTTP Status | Error Message | User Action |
|----------|-------------|---------------|-------------|
| Duplicate name | 400 | "Software image with name 'X' already exists" | Choose different name |
| Duplicate path | 400 | "Software image with path 'X' already exists" | Choose different path |
| Invalid path format | 400 | "Path must match format ^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$" | Fix path format |
| Missing required field | 400 | "Required field 'name' is missing" | Add required field |
| Invalid SOL speed | 400 | "SOL speed must be one of: 115200, 57600, ..." | Use valid enum value |
| Image not found (Read) | 404 | "Software image UUID not found" | Trigger recreation |
| Image not found (Update) | 404 | "Cannot update non-existent image" | Run terraform refresh |
| Image not found (Delete) | 404 | "Image already deleted" | No-op (success) |
| Authentication failure | 401 | "Login failed: invalid credentials" | Check BCM credentials |
| API unavailable | 503 | "BCM API unavailable: connection refused" | Check network/endpoint |

## Validation Rules

### Input Validation

1. **Name Validation:**
   - MUST NOT be empty
   - SHOULD be alphanumeric with hyphens/underscores
   - SHOULD be less than 255 characters

2. **Path Validation:**
   - MUST match regex: `^(/[-+_.a-zA-Z0-9]+)+/?(@\d+)?$`
   - MUST start with `/`
   - SHOULD start with `/cm/images/` (convention)

3. **SOL Speed Validation:**
   - MUST be one of: "115200", "57600", "38400", "19200", "9600", "4800", "2400", "1200"
   - Default: "115200"

4. **Module Validation:**
   - Module name MUST NOT be empty
   - Module parameters MAY be empty string

5. **UUID Validation:**
   - UUIDs MUST be valid UUID format (optional validation)
   - Empty string or null UUIDs are acceptable for optional fields

## Test Scenarios

### Positive Tests

1. **T1: Basic Create**
   - Input: name, path
   - Expected: Resource created with UUID, defaults applied

2. **T2: Full Create**
   - Input: All optional attributes
   - Expected: Resource created with all values set

3. **T3: Update Kernel Config**
   - Input: Change kernel_version and kernel_parameters
   - Expected: Image updated, changes reflected in state

4. **T4: Update Modules**
   - Input: Add/remove modules from list
   - Expected: Module list updated in BCM

5. **T5: Import Existing**
   - Input: UUID of existing image
   - Expected: Full configuration imported into state

6. **T6: Delete**
   - Input: Destroy command
   - Expected: Image removed from BCM

### Negative Tests

7. **T7: Duplicate Name**
   - Input: Name already exists
   - Expected: Error message, resource not created

8. **T8: Invalid Path**
   - Input: Path with invalid characters
   - Expected: Error message with format hint

9. **T9: Missing Required Field**
   - Input: Missing name or path
   - Expected: Validation error before API call

10. **T10: Invalid SOL Speed**
    - Input: SOL speed not in enum
    - Expected: Validation error with valid values

## Dependencies

### Internal Dependencies
- BCM API Client (`internal/provider/bcm_client.go`)
- Helper functions (`getStringValue`, `getBoolValue`, `getInt64Value`)
- Existing data source pattern (`data_source_cmpart_softwareimages.go`)

### External Dependencies
- Terraform Plugin Framework v1.16.1
- Terraform Plugin Testing v1.13.3
- BCM API endpoint (test environment)

### Data Dependencies
- BCM cluster with CMPart service enabled
- Valid authentication credentials
- Network connectivity to BCM endpoint

## Rollout Plan

### Phase 0: Research (Week 1)
- Discover update/delete API methods
- Test CRUD operations manually
- Document API behavior

### Phase 1: Implementation (Week 2)
- RED: Write acceptance tests
- GREEN: Minimal hardcoded implementation
- REFACTOR: Full API integration

### Phase 2: Testing (Week 3)
- Run acceptance tests against BCM
- Perform manual integration testing
- Fix bugs and edge cases

### Phase 3: Documentation (Week 3)
- Generate Terraform docs
- Write usage examples
- Update provider README

### Phase 4: Release (Week 4)
- Code review
- Pre-commit hooks validation
- Merge to main branch
- Tag release version

## Success Metrics

### Technical Metrics
- 100% acceptance test pass rate
- 0 golangci-lint errors
- < 120m acceptance test execution time
- Code coverage > 80% (unit tests)

### User Metrics
- Clear error messages for all failure modes
- Intuitive attribute naming
- Comprehensive documentation with examples
- ImportState functionality works on first try

## Open Questions

1. **Update API Method:** What is the exact method name and signature?
2. **Delete API Method:** What is the exact method name and parameter format?
3. **Force Parameter Semantics:** What does force=true do during creation?
4. **Immutable Fields:** Are name/path mutable or require ForceNew?
5. **Module UUID Generation:** Are module UUIDs client-side or server-side generated?
6. **Constraint Error Format:** What is the exact error response for duplicate name/path?

**Resolution:** All open questions will be resolved in Phase 0 Research.

## References

- BCM API Documentation: `/workspace/sampleRest/BCM_API_Complete_Documentation.md`
- SoftwareImage Entity: `/workspace/sampleRest/wip/resource_cmpart_softwareimage.md`
- Terraform Plugin Framework: https://developer.hashicorp.com/terraform/plugin/framework
- TDD Constitution: `/workspace/.specify/memory/constitution.md`
