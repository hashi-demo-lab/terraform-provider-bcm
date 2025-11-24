# Feature Specification: BCM Pre-flight Validation

**Feature Branch**: `001-bcm-preflight-validation`
**Created**: 2025-11-24
**Status**: Draft
**Input**: User description: "Implement pre-flight validation using BCM validate* API methods across all 5 resource types (software images, categories, devices, networks, kubernetes clusters). The BCM API provides validate* methods that perform server-side validation before CRUD operations."

## User Scenarios & Testing

### User Story 1 - Catch Invalid Field Values Before Create (Priority: P1)

A Terraform operator defines a new software image resource with an invalid SOL speed (baud rate). Instead of waiting for the create operation to fail after sending the request to BCM, the provider validates the configuration immediately and returns a clear error message identifying the invalid field.

**Why this priority**: This is the most critical user-facing benefit. Catching invalid values before API calls provides immediate feedback with specific field-level error messages, dramatically improving the user experience and reducing debugging time.

**Independent Test**: Can be fully tested by creating a resource with one invalid field value (e.g., SOL speed = 999999) and verifying that the provider returns a validation error before attempting the API call. Delivers immediate value by preventing malformed requests.

**Acceptance Scenarios**:

1. **Given** a Terraform configuration with invalid SOL speed (999999), **When** terraform apply is executed, **Then** the provider returns "Validation Error: SOLSpeed - Illegal value for: SOLSpeed" before calling the BCM API
2. **Given** a device resource with invalid hostname format ("invalid_hostname_123"), **When** terraform apply is executed, **Then** the provider returns "Validation Error: hostname - The hostname can only contain a-z, A-Z, 0-9 and dashes" before calling the BCM API
3. **Given** a category with missing required parentSoftwareImage field, **When** terraform apply is executed, **Then** the provider returns "Validation Error: parentSoftwareImage - Parent software image needs to be set" before calling the BCM API

---

### User Story 2 - Detect Duplicate Names Before Create (Priority: P2)

A Terraform operator attempts to create a resource with a name that already exists in BCM. The provider detects this duplicate name conflict during validation and returns an error, preventing the failed API call and providing clear guidance that the name is already in use.

**Why this priority**: Duplicate name detection is valuable for preventing common mistakes, but slightly less critical than field validation since name conflicts are typically caught during planning when the state is refreshed. Still provides significant value by catching issues early.

**Independent Test**: Can be fully tested by attempting to create a resource with a name matching an existing resource and verifying the provider returns "A [resource] with that name already exists" error. Delivers value by preventing duplicate name conflicts without requiring an API round-trip.

**Acceptance Scenarios**:

1. **Given** an existing category named "default", **When** terraform apply attempts to create another category named "default", **Then** the provider returns "Validation Error: name - A category with that name already exists"
2. **Given** an existing software image named "ubuntu-22.04", **When** terraform apply attempts to create another image named "ubuntu-22.04", **Then** the provider returns duplicate name validation error
3. **Given** an existing network named "management", **When** terraform apply attempts to create another network named "management", **Then** the provider returns duplicate name validation error

---

### User Story 3 - Receive Advisory Warnings for Suspicious Values (Priority: P3)

A Terraform operator defines a software image with a path that does not exist on the BCM server. The provider issues a warning but allows the operation to proceed since the path may be created later or may be a legitimate future reference.

**Why this priority**: Warnings are helpful for catching potential issues, but they don't block operations. This is lowest priority since warnings are informational and the operation can still succeed.

**Independent Test**: Can be fully tested by creating a software image with a non-existent path and verifying the provider issues a warning but completes the operation. Delivers value by alerting users to potential configuration issues without blocking valid workflows.

**Acceptance Scenarios**:

1. **Given** a software image with path "/cm/images/nonexistent.iso", **When** terraform apply is executed, **Then** the provider issues "Validation Warning: path - The software image path does not exist" but continues with create operation
2. **Given** a resource configuration that triggers a WARNING severity validation response, **When** terraform apply is executed, **Then** the provider displays the warning but does not halt execution
3. **Given** multiple validation issues with mixed ERROR and WARNING severity, **When** terraform apply is executed, **Then** the provider displays all warnings but only halts on ERROR severity issues

---

### User Story 4 - Validate Updates Before Modification (Priority: P2)

A Terraform operator updates an existing resource's configuration with invalid values. The provider validates the updated configuration against BCM's rules and rejects the change before sending the update request, preventing partial updates or corrupted state.

**Why this priority**: Update validation is equally important as create validation for data integrity, but ranked P2 because it shares the same implementation and test coverage as create operations. The value is identical but can be tested as part of the same validation framework.

**Independent Test**: Can be fully tested by updating a resource with invalid field values and verifying validation errors are returned before the update API call. Delivers value by preventing invalid updates from reaching the BCM API.

**Acceptance Scenarios**:

1. **Given** an existing software image resource, **When** terraform apply attempts to update SOL speed to invalid value 999999, **Then** the provider returns validation error before calling updateSoftwareImage API
2. **Given** an existing device resource, **When** terraform apply attempts to update hostname to invalid format, **Then** the provider returns validation error before calling updateDevice API
3. **Given** an existing category resource, **When** terraform apply attempts to update name to duplicate existing name, **Then** the provider returns duplicate name validation error before calling updateCategory API

---

### User Story 5 - Consistent Validation Across All Resource Types (Priority: P2)

A Terraform operator works with multiple BCM resource types (software images, categories, devices, networks, kubernetes clusters). All resources provide consistent validation behavior with the same error message format and severity handling, creating a predictable user experience.

**Why this priority**: Consistency improves overall user experience and maintainability, but is ranked P2 because it's a quality enhancement rather than a new functional capability. The value is in standardization across resources.

**Independent Test**: Can be fully tested by triggering validation errors on each of the 5 resource types and verifying identical error message structure, severity handling, and workflow. Delivers value through consistent, predictable behavior.

**Acceptance Scenarios**:

1. **Given** validation errors from software image, category, device, network, and kubernetes cluster resources, **When** errors are displayed, **Then** all follow format "Validation Error: [field] - [message]"
2. **Given** WARNING severity validations from any resource type, **When** terraform apply is executed, **Then** warnings are displayed but operation continues
3. **Given** ERROR severity validations from any resource type, **When** terraform apply is executed, **Then** operation halts with error diagnostics

---

### Edge Cases

- What happens when BCM API validation service is unavailable? System should return standard API error (timeout, connection refused) rather than treating it as validation failure.
- What happens when validation returns unknown severity levels? System should treat unknown severity as ERROR to be conservative.
- What happens when validation response cannot be parsed? System should return error indicating validation response parsing failure rather than silently continuing.
- What happens when CREATE operation receives "Zero UUID" error mixed with real errors? System should filter out expected "Zero UUID" errors while preserving real validation errors.
- What happens when validating kubernetes clusters using incorrect service name casing (CMKube instead of cmkube)? System should use correct lowercase "cmkube" service name as documented.
- What happens when multiple validation errors occur for the same field? System should display all validation messages for comprehensive feedback.

## Requirements

### Functional Requirements

- **FR-001**: System MUST provide a generic ValidateEntity() helper function in bcm_client.go that accepts service name, validation method, entity, and isCreate flag
- **FR-002**: System MUST call appropriate validate* method before all CREATE operations for software images, categories, devices, networks, and kubernetes clusters
- **FR-003**: System MUST call appropriate validate* method before all UPDATE operations for software images, categories, devices, networks, and kubernetes clusters
- **FR-004**: System MUST filter out "Zero UUID" validation errors when isCreate flag is true, since new entities have no UUID assigned yet
- **FR-005**: System MUST preserve and display all non-UUID validation errors and warnings from BCM API responses
- **FR-006**: System MUST distinguish between ERROR severity (halt operation) and WARNING severity (display but continue) in validation responses
- **FR-007**: System MUST use correct service name casing for each resource type (CMPart, CMDevice, CMNet, cmkube)
- **FR-008**: System MUST format validation errors as Terraform diagnostics with format "Validation Error: [field] - [message]"
- **FR-009**: System MUST format validation warnings as Terraform diagnostics with format "Validation Warning: [field] - [message]"
- **FR-010**: System MUST halt CREATE/UPDATE operation if any ERROR severity validation issues are detected
- **FR-011**: System MUST allow CREATE/UPDATE operation to proceed if only WARNING severity validation issues are detected
- **FR-012**: System MUST parse validation responses into structured ValidationError objects with Field, Message, ErrorCode, Severity, and EntityUUID fields
- **FR-013**: System MUST handle validation response parsing errors gracefully with descriptive error messages

### Key Entities

- **ValidationError**: Structured representation of BCM validation feedback containing Field (string, the attribute that failed validation), Message (string, human-readable error description), ErrorCode (string, BCM error code like BAD_VALUE or DUPLICATE_FIELD), Severity (string, either ERROR or WARNING), EntityUUID (string, reference to related entity if applicable)

- **Validation Response**: Array of validation objects returned by BCM validate* methods, either empty array for success or array of validation error objects with baseType "Validation"

## Success Criteria

### Measurable Outcomes

- **SC-001**: Terraform operators receive field-specific validation errors within 200ms for invalid configurations before any BCM API create/update calls are made
- **SC-002**: All 5 resource types (software images, categories, devices, networks, kubernetes clusters) provide consistent validation error messages following "Validation Error: [field] - [message]" format
- **SC-003**: Invalid field values (SOL speed, hostname format) are caught and rejected before API calls with specific field names in error messages
- **SC-004**: Duplicate name conflicts are detected during validation and reported with "already exists" messages before create operations attempt API calls
- **SC-005**: WARNING severity validations display advisory messages but allow operations to proceed, while ERROR severity validations halt operations
- **SC-006**: CREATE operations successfully filter out expected "Zero UUID" errors while preserving all real validation errors
- **SC-007**: Validation adds maximum 200ms overhead per CREATE/UPDATE operation (one additional API call)
- **SC-008**: Zero validation errors appear in provider logs related to incorrect service name casing (cmkube lowercase vs CamelCase)

## Assumptions

1. **BCM API Stability**: Validation API endpoints will continue to follow current response format (empty array for success, array of validation objects for errors/warnings)
2. **Service Name Consistency**: The cmkube service will continue to use lowercase naming while other services (CMPart, CMDevice, CMNet) remain CamelCase
3. **Zero UUID Pattern**: CREATE operations will continue to generate "Zero UUID" validation errors that need filtering
4. **Severity Levels**: BCM will continue to use only "ERROR" and "WARNING" severity levels in validation responses
5. **Validation Method Naming**: Validation methods will continue to follow validate[ResourceType] naming pattern across all services
6. **Performance Tolerance**: 50-200ms additional latency per operation for validation is acceptable trade-off for better error messages
7. **Network Reliability**: Validation API calls have similar reliability characteristics to create/update API calls
8. **Error Code Stability**: BCM error codes (BAD_VALUE, DUPLICATE_FIELD, NOT_NULL) will remain consistent
9. **Existing Resources**: Network resource (resource_cmnet_network.go) and all other resources follow standard CRUD pattern with buildAPIEntity() helper
10. **UUID Generation**: Provider will continue to generate UUIDs for new resources in buildAPIEntity(), making validation work identically for CREATE and UPDATE operations

## Dependencies

### External Dependencies

- **BCM API Services**: Requires availability of validate* methods across CMPart, CMDevice, CMNet, and cmkube services
- **BCM JSON-RPC Endpoint**: Validation calls use same authenticated JSON-RPC endpoint as other BCM operations
- **BCM API Version**: Assumes current BCM API version maintains validation method signatures and response formats

### Internal Dependencies

- **BCMClient**: Validation helper extends existing BCMClient in internal/provider/bcm_client.go
- **Resource Implementations**: Requires integration into Create() and Update() methods of 5 resource types
- **buildAPIEntity() Helper**: Validation relies on existing buildAPIEntity() helper to construct entities for validation
- **Terraform Plugin Framework**: Uses framework's diagnostics system (resp.Diagnostics.AddError/AddWarning) for error reporting
- **Test Helpers**: Uses existing test helper infrastructure (createTestBCMClient, generateUniqueTestName) for acceptance tests

### Blocked By

None - all required BCM API methods and provider infrastructure already exist

### Blocks

None - this is a quality-of-life enhancement that does not block other features

## Out of Scope

The following items are explicitly excluded from this feature:

1. **Client-Side Validation**: This feature only implements BCM server-side validation. Client-side validation rules (regex patterns, format validators) in Terraform schema remain unchanged
2. **Custom Validation Logic**: No new validation rules are added beyond what BCM API already provides
3. **Validation Caching**: Validation results are not cached; each operation performs fresh validation
4. **Partial Validation**: Either all fields are validated or none; no support for validating subset of fields
5. **Validation Ordering**: No control over order of validation checks; BCM determines validation sequence
6. **Async Validation**: All validation is synchronous; no support for background or deferred validation
7. **Validation Reports**: No aggregated validation reports across multiple resources
8. **Historical Validation**: No tracking or logging of validation history over time
9. **Validation Overrides**: No mechanism to skip or override validation failures (ERROR severity must be fixed)
10. **Cross-Resource Validation**: Each resource validates independently; no validation of relationships between resources

## Open Questions

None - all technical details have been validated through test scripts and analysis reports.

## Non-Functional Requirements

### Performance

- **NFR-001**: Validation should add no more than 200ms latency to CREATE/UPDATE operations
- **NFR-002**: Validation failures should return within same timeout window as BCM API calls (120 seconds default)
- **NFR-003**: Validation should not impact provider startup time or initialization

### Reliability

- **NFR-004**: Validation helper must handle malformed BCM responses gracefully without panics
- **NFR-005**: Validation failures must preserve original error context from BCM API
- **NFR-006**: Network failures during validation should be distinguishable from validation failures

### Maintainability

- **NFR-007**: Single ValidateEntity() helper function serves all 5 resource types with no duplication
- **NFR-008**: Service name and validation method name are parameters, allowing easy addition of new resources
- **NFR-009**: Validation response parsing uses null-safe getString() helper to prevent nil pointer errors
- **NFR-010**: Validation logic is covered by unit tests independent of acceptance tests

### Usability

- **NFR-011**: Validation error messages must include specific field names that failed validation
- **NFR-012**: Validation warnings must be visually distinct from errors in Terraform output
- **NFR-013**: Validation errors must provide actionable guidance on what needs to be fixed

### Security

- **NFR-014**: Validation requests use same authentication mechanism as other BCM API calls
- **NFR-015**: Validation error messages must not expose sensitive internal BCM system details
- **NFR-016**: Entity data sent for validation should be identical to data sent for create/update (no additional exposure)

## Implementation Notes

### Service Name Casing Reference

| Resource Type | Service Name | Validation Method | Provider File |
|---------------|--------------|-------------------|---------------|
| Software Images | CMPart | validateSoftwareImage | resource_cmpart_softwareimage.go |
| Categories | CMDevice | validateCategory | resource_cmdevice_category.go |
| Devices | CMDevice | validateDevice | resource_cmdevice_device.go |
| Networks | CMNet | validateNetwork | resource_cmnet_network.go |
| Kubernetes Clusters | cmkube | validateKubeCluster | resource_cmkube_cluster.go |

### Integration Points

**CREATE Operations:**
- resource_cmpart_softwareimage.go:288 (before addSoftwareImage)
- resource_cmdevice_category.go (before addCategory)
- resource_cmdevice_device.go (before addDevice)
- resource_cmnet_network.go (before network create)
- resource_cmkube_cluster.go:272 (before addKubeCluster)

**UPDATE Operations:**
- resource_cmpart_softwareimage.go:483 (before updateSoftwareImage)
- resource_cmdevice_category.go (before updateCategory)
- resource_cmdevice_device.go:783 (before updateDevice)
- resource_cmnet_network.go (before network update)
- resource_cmkube_cluster.go:569 (before updateKubeCluster)

### Zero UUID Filtering Logic

For CREATE operations only (isCreate=true):
- Filter validation errors where Field == "uuid" AND Message contains "Zero UUID"
- Log at DEBUG level: "Skipping expected Zero UUID validation error for new entity"
- Preserve all other validation errors and warnings
- For UPDATE operations (isCreate=false): No filtering applied

### Error Message Format Examples

**Field Validation Error:**
```
Validation Error: SOLSpeed
Illegal value for: SOLSpeed
```

**Duplicate Name Error:**
```
Validation Error: name
A category with that name already exists
```

**Path Warning:**
```
Validation Warning: path
The software image path does not exist
```

### Performance Trade-off Analysis

**Overhead**: +1 API call per CREATE/UPDATE operation (50-200ms)

**Benefits**:
- Immediate error feedback before failed operations
- Field-specific error messages
- Duplicate name detection
- Consistent validation across resources

**Conclusion**: Performance cost is acceptable given UX improvements

## Related Documentation

- **Analysis Report**: /workspace/ai_reports/bcm_validate_methods_analysis.md
- **Summary Report**: /workspace/ai_reports/bcm_validate_methods_summary.md
- **CREATE Validation Findings**: /workspace/ai_reports/create_operation_validation_findings.md
- **Test Scripts**: /workspace/sampleRest/test_validate_*.py
- **GitHub Issue**: https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/51
