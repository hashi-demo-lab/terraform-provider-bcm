# Feature Specification: Category Dynamic Fields Schema Implementation

**Feature Branch**: `001-category-dynamic-fields`
**Created**: 2025-11-24
**Status**: Draft
**Input**: User description: "Implement proper schemas for 5 dynamic type fields in the bcm_cmdevice_category resource"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Static Route Configuration (Priority: P1)

As a cluster administrator, I need to configure custom static routes for nodes in a category so that traffic can be routed through specific gateways for different network segments (e.g., storage networks, management networks).

**Why this priority**: Static routes are fundamental network configuration that enable multi-network cluster architectures. This is commonly needed in HPC environments where compute nodes need separate routes for storage, management, and high-performance interconnects.

**Independent Test**: Can be fully tested by creating a category with static routes, verifying the routes are applied to nodes in that category, and updating/removing routes. Delivers immediate value for multi-network cluster configurations.

**Acceptance Scenarios**:

1. **Given** a new category, **When** I configure it with 2 static routes (one for storage network 192.168.1.0/24 via 10.0.0.1, one for backup network 192.168.2.0/24 via 10.0.0.2), **Then** both routes are persisted and returned on read operations
2. **Given** an existing category with static routes, **When** I update it to add a third route or modify an existing route's gateway, **Then** the changes are applied without affecting other routes
3. **Given** a category with static routes, **When** I import the resource using its UUID, **Then** all static routes are accurately imported into state
4. **Given** a category with static routes, **When** the routes are modified externally via BCM API, **Then** Terraform detects the drift and can restore the desired state

---

### User Story 2 - NFS Export Management (Priority: P1)

As a storage administrator, I need to configure NFS exports at the category level so that all nodes in the category automatically share filesystems with appropriate permissions and squash settings for cluster-wide storage access.

**Why this priority**: NFS exports are critical for HPC storage architecture. Categories with head node or storage roles commonly need to export /home, /shared, or scratch space to compute nodes. This enables centralized storage management.

**Independent Test**: Can be fully tested by creating a category with NFS exports, verifying export configurations persist through CRUD operations, and testing different permission combinations (read-only, read-write, root squash variations).

**Acceptance Scenarios**:

1. **Given** a new category, **When** I configure it with an NFS export for /home with read-write access and root squash enabled, **Then** the export configuration is saved and nodes in this category share /home accordingly
2. **Given** an existing category with exports, **When** I update export permissions from read-only to read-write or toggle async/sync mode, **Then** the changes apply to all nodes in the category
3. **Given** a category with exports, **When** I import the resource, **Then** all export paths, network references, and permission flags are preserved
4. **Given** a category with exports, **When** exports are modified externally (permission changes, path changes), **Then** Terraform detects drift and can remediate

---

### User Story 3 - Role Assignment (Priority: P2)

As a cluster architect, I need to assign roles (HeadNodeRole, StorageRole, ComputeRole, etc.) at the category level so that all nodes in a category inherit the same service roles and configurations, enabling role-based cluster organization.

**Why this priority**: Role assignment is important for cluster architecture but typically configured less frequently than network and storage settings. Roles define which services run on nodes (provisioning, monitoring, backup, etc.).

**Independent Test**: Can be fully tested by creating categories with different role combinations, verifying role inheritance by nodes, and testing role updates. Delivers value for role-based node organization.

**Acceptance Scenarios**:

1. **Given** a new category for head nodes, **When** I assign HeadNodeRole with add_services=true, **Then** the role is persisted and the UUID is computed after creation
2. **Given** an existing category, **When** I add multiple roles (StorageRole, BackupRole, MonitoringRole), **Then** all roles are applied and each has a unique UUID
3. **Given** a category with roles, **When** I import it, **Then** role names, types, UUIDs, and add_services flags are all preserved
4. **Given** a category with roles, **When** roles are modified externally (added/removed/reordered), **Then** Terraform detects the drift

---

### User Story 4 - GPU Configuration (Priority: P3)

As an AI/ML cluster administrator, I need to configure GPU settings at the category level so that compute nodes with GPUs inherit proper device IDs, models, and compute mode settings for GPU workloads.

**Why this priority**: GPU settings are specialized configuration for GPU-equipped nodes. Less commonly used than network/storage settings, but critical for AI/ML clusters. Lower priority as not all clusters have GPUs.

**Independent Test**: Can be fully tested by creating GPU-equipped categories, configuring multiple GPUs with different compute modes, and verifying settings persist through CRUD operations.

**Acceptance Scenarios**:

1. **Given** a new category for GPU nodes, **When** I configure 4 Tesla V100 GPUs with device IDs 0-3 and default compute mode, **Then** all GPU configurations are saved
2. **Given** an existing GPU category, **When** I update compute mode from default to exclusive for GPU 0, **Then** only that GPU's settings change
3. **Given** a GPU category, **When** I import it, **Then** all device IDs, models, and compute modes are preserved
4. **Given** a GPU category, **When** GPU settings are modified externally, **Then** Terraform detects the changes

---

### User Story 5 - Service Configuration (Priority: P3)

As a system administrator, I need to configure services at the category level so that all nodes in a category run the same set of system services with consistent configurations.

**Why this priority**: Service configuration is specialized and less commonly modified. Lower priority as the exact BCM API structure needs research before implementation. Can be implemented after higher priority fields are complete.

**Independent Test**: Can be fully tested once BCM API service structure is researched. Will follow same CRUD pattern as other fields.

**Acceptance Scenarios**:

1. **Given** the BCM API service structure is documented, **When** I configure services on a category, **Then** service configurations persist through CRUD operations
2. **Given** a category with services, **When** services are modified, **Then** changes apply to all nodes
3. **Given** a category with services, **When** I import the resource, **Then** all service configurations are preserved

---

### Edge Cases

- **Empty lists**: What happens when a field is set to an empty list vs. null? (Should preserve empty list in state, not convert to null)
- **Invalid CIDR notation**: How does system handle invalid destination values in static routes? (Terraform validators should catch before API call)
- **Non-existent network UUID**: What happens when fsexports reference a network UUID that doesn't exist? (BCM API validation should return error, Terraform should surface it)
- **Duplicate routes**: How does system handle multiple static routes with the same destination? (Allow - legitimate for different metrics/gateways)
- **Role UUID conflicts**: What happens when manually specifying a role UUID that already exists? (UUID is computed, not user-provided, so not applicable)
- **Unknown role types**: How does system validate role child_type values? (Use string validators with known role types, but allow any string for flexibility)
- **GPU device ID conflicts**: What happens with duplicate device IDs in gpu_settings? (Allow - BCM API should validate)
- **Maximum list sizes**: Are there BCM API limits on number of routes, exports, roles, services, GPUs? (No known limits, handle any reasonable size)

## Requirements *(mandatory)*

### Functional Requirements

- **FR-001**: System MUST define a proper ListNestedAttribute schema for static_routes field with Route nested object containing destination (string, required), gateway (string, required), and metric (int64, optional) attributes
- **FR-002**: System MUST define a proper ListNestedAttribute schema for fsexports field with FSExport nested object containing path (string, required), network (string, required), allow_write (bool, optional), root_squash (bool, optional), and async (bool, optional) attributes
- **FR-003**: System MUST define a proper ListNestedAttribute schema for roles field with Role nested object containing name (string, required), child_type (string, required), uuid (string, computed), and add_services (bool, optional) attributes
- **FR-004**: System MUST define a proper ListNestedAttribute schema for gpu_settings field with GPUSetting nested object containing device_id (string, required), model (string, optional), and compute_mode (string, optional) attributes
- **FR-005**: System MUST research BCM API service structure and define appropriate schema for services field
- **FR-006**: System MUST replace all types.DynamicNull() placeholders with proper list handling in Create/Read/Update methods
- **FR-007**: System MUST handle field name mapping between Terraform snake_case and BCM API camelCase for all new nested object attributes (e.g., allow_write → allowWrite, root_squash → rootSquash, add_services → addServices, device_id → deviceId, compute_mode → computeMode)
- **FR-008**: System MUST preserve empty lists in state (not convert to null) when fields are set to empty arrays
- **FR-009**: System MUST implement null-safe extraction helpers for nested object attributes when reading from BCM API responses
- **FR-010**: System MUST handle BCM API response structure where these fields may be null, empty array, or populated array
- **FR-011**: System MUST validate static route destination format using CIDR notation validation
- **FR-012**: System MUST validate static route gateway as valid IPv4 address format
- **FR-013**: System MUST support role child_type values including HeadNodeRole, StorageRole, BackupRole, MonitoringRole, ProvisioningRole, BootRole, ComputeRole (and allow others for flexibility)
- **FR-014**: System MUST include base_type in BCM API entity structure for fsexports (FSExport) and gpu_settings (GPUSetting) when serializing to API
- **FR-015**: System MUST NOT include base_type in Terraform schema (internal BCM implementation detail)

### Key Entities

- **StaticRoute**: Network routing configuration with destination CIDR, gateway IP, and optional metric
- **FSExport**: NFS filesystem export configuration with path, network UUID reference, and permission flags
- **Role**: Service role assignment with name, type classification, computed UUID, and service auto-add flag
- **GPUSetting**: GPU hardware configuration with device identifier, model name, and compute mode
- **Service**: System service configuration (structure to be determined from BCM API research)

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: All 5 dynamic type fields are replaced with properly typed schemas and pass type checking at compile time
- **SC-002**: Each field has comprehensive acceptance test coverage including Create, Read, Update, Delete, Import, Idempotency, and Drift Detection tests (minimum 7 test scenarios per field)
- **SC-003**: BCM API service structure is documented with example JSON responses in sampleRest directory
- **SC-004**: Category resource CRUD operations correctly serialize and deserialize all 5 fields between Terraform state and BCM API format without data loss
- **SC-005**: Empty lists are preserved in state (not converted to null) for all list fields
- **SC-006**: External modifications to these fields via BCM API are detected by Terraform drift detection within 5 seconds
- **SC-007**: Field name mapping between Terraform snake_case and BCM API camelCase works correctly in both directions (plan to API, API to state)
- **SC-008**: Provider documentation is auto-generated with examples for each field showing proper usage patterns
- **SC-009**: All nested object attributes support null values without causing runtime errors
- **SC-010**: Static route CIDR and IP validation catches invalid formats before API submission

## Assumptions *(optional)*

- BCM API maintains consistent structure for fsexports, roles, and gpu_settings as documented in sampleRest directory
- BCM API accepts null, empty array [], and populated arrays interchangeably for optional list fields
- Role UUIDs are assigned by BCM API and should be computed (read-only) in Terraform
- The services field structure can be determined through BCM API exploration similar to other fields
- No BCM API size limits exist for these lists beyond memory constraints
- Validation of network UUIDs for fsexports is handled by BCM API (Terraform doesn't need to pre-validate existence)
- Base type fields (baseType, childType in API) are BCM internal structure and should not be exposed in Terraform schema
- Users expect empty lists to remain empty lists (not null) in state to match their configuration intent

## Dependencies *(optional)*

- Issue #36 (bcm_cmdevice_roles data source) - Related to roles field, but not blocking. Roles can be configured via UUID reference even without data source
- BCM API cluster at https://172.21.15.254:8081 must be accessible for acceptance testing
- Test helpers in internal/provider/test_helpers.go for BCM client creation and resource verification
- Existing FSMount schema implementation as reference pattern for nested objects

## Related Issues

- GitHub Issue #49: Define proper schemas for dynamic type fields in bcm_cmdevice_category
- GitHub Issue #36: Implement bcm_cmdevice_roles data source (related to roles field)

## API Contract

### Static Routes (staticRoutes)

**BCM API Format:**
```json
{
  "destination": "192.168.1.0/24",
  "gateway": "10.0.0.1",
  "metric": 100
}
```

**Terraform Schema:**
```hcl
static_routes = [
  {
    destination = "192.168.1.0/24"
    gateway     = "10.0.0.1"
    metric      = 100
  }
]
```

### NFS Exports (fsexports)

**BCM API Format:**
```json
{
  "baseType": "FSExport",
  "path": "/home",
  "network": "network-uuid-1234",
  "allowWrite": true,
  "rootSquash": false,
  "async": true
}
```

**Terraform Schema:**
```hcl
fsexports = [
  {
    path        = "/home"
    network     = "network-uuid-1234"
    allow_write = true
    root_squash = false
    async       = true
  }
]
```

### Roles (roles)

**BCM API Format:**
```json
{
  "baseType": "Role",
  "childType": "HeadNodeRole",
  "name": "headnode",
  "uuid": "role-uuid-5678",
  "addServices": true
}
```

**Terraform Schema:**
```hcl
roles = [
  {
    name         = "headnode"
    child_type   = "HeadNodeRole"
    add_services = true
    # uuid is computed
  }
]
```

### GPU Settings (gpuSettings)

**BCM API Format:**
```json
{
  "baseType": "GPUSetting",
  "deviceId": "0",
  "model": "Tesla V100",
  "computeMode": "default"
}
```

**Terraform Schema:**
```hcl
gpu_settings = [
  {
    device_id    = "0"
    model        = "Tesla V100"
    compute_mode = "default"
  }
]
```

### Services (services)

**Status**: Structure needs research via BCM API exploration

**Action Required**: Query BCM API for categories with services configured to determine object structure
