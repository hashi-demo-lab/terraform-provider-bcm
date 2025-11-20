# API Contract: BCM CMDevice Category

**Service**: cmdevice
**Endpoint**: `POST https://{bcm-host}:8081/json`
**Authentication**: Cookie-based session (`cm-login-token`)
**Date**: 2025-11-21

## Overview

This document defines the complete API contract for BCM CMDevice Category operations. The BCM API uses JSON-RPC over HTTPS with cookie-based authentication.

---

## Authentication

### Login (Required for all subsequent operations)

**Endpoint**: `POST https://{bcm-host}:8081/json`

**Request**:
```json
{
  "service": "login",
  "username": "root",
  "password": "your-password"
}
```

**Response** (200 OK):
```json
true
```

**Set-Cookie Header**:
```
cm-login-token=<session-token>; Path=/; HttpOnly
```

**Note**: All subsequent requests automatically include this cookie via the HTTP client's cookie jar.

---

## API Methods

### 1. getCategories

Lists all categories in the BCM cluster.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "getCategories"
}
```

**Response** (200 OK):
```json
[
  {
    "uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
    "baseType": "Category",
    "childType": "",
    "modified": false,
    "to_be_removed": false,
    "revision": "v1",
    "parent_uuid": null,
    "name": "default",
    "notes": "Default category",
    "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
    "bootLoader": "SYSLINUX",
    "kernelVersion": "5.15.0-58-generic",
    "kernelParameters": "quiet splash"
  },
  {
    "uuid": "1bf7e844-4126-5580-cged-df3e348b3920",
    "baseType": "Category",
    "childType": "",
    "name": "gpu-nodes",
    "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe"
  }
]
```

**Usage Pattern**: Data source implementation (list + filter client-side)

**Error Responses**:
- `401 Unauthorized`: Session cookie expired or invalid
- `500 Internal Server Error`: BCM internal error

---

### 2. getCategory

Retrieves a single category by name (efficient direct lookup).

**Request**:
```json
{
  "service": "cmdevice",
  "call": "getCategory",
  "args": ["category-name"]
}
```

**Arguments**:
- `args[0]` (string, required): Category name

**Response** (200 OK):
```json
{
  "uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
  "baseType": "Category",
  "childType": "",
  "modified": false,
  "to_be_removed": false,
  "revision": "v1",
  "parent_uuid": null,
  "name": "default",
  "notes": "Default category for all nodes",
  "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
  "softwareImageProxy": {
    "uuid": "9cf8f955-5237-6691-dhfe-gf4f459c4031",
    "baseType": "SoftwareImageProxy",
    "parentSoftwareImage": "7de7d822-2104-4358-aaba-bd33c4d2a698",
    "revisionID": 1
  },
  "bootLoader": "SYSLINUX",
  "bootLoaderProtocol": "HTTP",
  "kernelVersion": "5.15.0-58-generic",
  "kernelParameters": "quiet splash",
  "kernelOutputConsole": "tty0",
  "modules": [
    {
      "baseType": "KernelModule",
      "name": "nvidia-drm",
      "parameters": "modeset=1"
    }
  ],
  "disksetup": "<?xml version=\"1.0\"?><diskSetup>...</diskSetup>",
  "installMode": "AUTO",
  "newNodeInstallMode": "FULL",
  "installBootRecord": false,
  "defaultGateway": "192.168.1.1",
  "defaultGatewayMetric": 100,
  "nameServers": ["8.8.8.8", "8.8.4.4"],
  "searchDomains": ["example.com"],
  "timeServers": ["time.nist.gov"],
  "fsmounts": [
    {
      "uuid": "abc123",
      "baseType": "FSMount",
      "device": "nfs-server:/export/home",
      "mountpoint": "/home",
      "filesystem": "nfs",
      "mountoptions": "rsize=32768,wsize=32768",
      "fsck": "NONE",
      "dump": false,
      "rdma": false
    }
  ],
  "bmcSettings": {
    "uuid": "def456",
    "baseType": "BMCSettings",
    "userName": "admin",
    "password": "secret",
    "privilege": "ADMINISTRATOR",
    "firmwareManageMode": "AUTO"
  },
  "dataNode": false,
  "allowNetworkingRestart": true
}
```

**Usage Pattern**: Resource Read operation (efficient direct lookup)

**Error Responses**:
- `404 Not Found`: Category with specified name does not exist
- `401 Unauthorized`: Session cookie expired
- `500 Internal Server Error`: BCM internal error

---

### 3. addCategory

Creates a new category.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "addCategory",
  "args": [
    {
      "baseType": "Category",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "",
      "name": "new-category",
      "notes": "New category for testing",
      "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
      "bootLoader": "GRUB2",
      "kernelVersion": "5.15.0-58-generic",
      "kernelParameters": "quiet splash",
      "modules": [
        {
          "baseType": "KernelModule",
          "childType": "",
          "modified": true,
          "to_be_removed": false,
          "revision": "",
          "name": "nvidia-drm",
          "parameters": "modeset=1"
        }
      ],
      "installMode": "AUTO"
    },
    false
  ]
}
```

**Arguments**:
- `args[0]` (object, required): Category entity with all desired fields
  - Must include `baseType: "Category"`
  - Must include `name` (unique)
  - Must include `managementNetwork` (valid UUID)
  - All nested objects must include their `baseType`
- `args[1]` (boolean, optional): Force flag (default: false)
  - `false`: Fail on validation warnings
  - `true`: Override validation warnings, still fail on errors

**Response** (200 OK):
```json
"0ae6d733-3015-4479-bfab-ce2d237a2809"
```

**Response Type**: String containing the UUID of the newly created category

**Usage Pattern**: Resource Create operation

**Error Responses**:
- `400 Bad Request`: Invalid category entity (missing required fields, invalid references)
- `409 Conflict`: Category with same name already exists
- `422 Unprocessable Entity`: Validation failed (see validation error format below)
- `401 Unauthorized`: Session cookie expired
- `500 Internal Server Error`: BCM internal error

**Validation Error Format**:
```json
{
  "success": false,
  "validation": [
    {
      "field": "managementNetwork",
      "message": "Management network UUID not found"
    },
    {
      "field": "name",
      "message": "Category name already exists"
    }
  ]
}
```

---

### 4. updateCategory

Updates an existing category.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "updateCategory",
  "args": [
    {
      "uuid": "0ae6d733-3015-4479-bfab-ce2d237a2809",
      "baseType": "Category",
      "childType": "",
      "modified": true,
      "to_be_removed": false,
      "revision": "v1",
      "name": "updated-category",
      "notes": "Updated notes",
      "managementNetwork": "84d8d82b-3ae7-4433-a793-bb44d5c3b4fe",
      "kernelParameters": "quiet splash intel_iommu=on"
    },
    false
  ]
}
```

**Arguments**:
- `args[0]` (object, required): Complete category entity including UUID
  - Must include `uuid` of existing category
  - Must include `baseType: "Category"`
  - Should include all fields (full entity update)
- `args[1]` (boolean, optional): Force flag (default: false)
  - `false`: Fail if update affects assigned nodes
  - `true`: Apply update even with node assignments

**Response** (200 OK):
```json
true
```

**Response Type**: Boolean true on success, or updated entity object

**Usage Pattern**: Resource Update operation

**Error Responses**:
- `404 Not Found`: Category UUID not found
- `400 Bad Request`: Invalid category entity
- `422 Unprocessable Entity`: Validation failed or nodes assigned (without force)
- `401 Unauthorized`: Session cookie expired
- `500 Internal Server Error`: BCM internal error

---

### 5. validateCategory

Validates category configuration without persisting changes (pre-flight validation).

**Request**:
```json
{
  "service": "cmdevice",
  "call": "validateCategory",
  "args": [
    {
      "baseType": "Category",
      "name": "test-category",
      "managementNetwork": "invalid-uuid",
      "bootLoader": "INVALID_LOADER"
    }
  ]
}
```

**Arguments**:
- `args[0]` (object, required): Category entity to validate

**Response - Success** (200 OK):
```json
{
  "success": true,
  "validation": []
}
```

**Response - Validation Errors** (422 Unprocessable Entity):
```json
{
  "success": false,
  "validation": [
    {
      "field": "managementNetwork",
      "message": "Management network UUID format invalid",
      "severity": "error"
    },
    {
      "field": "bootLoader",
      "message": "Boot loader type not recognized",
      "severity": "error"
    }
  ]
}
```

**Response - Validation Warnings** (200 OK):
```json
{
  "success": true,
  "validation": [
    {
      "field": "kernelParameters",
      "message": "Experimental kernel parameter detected",
      "severity": "warning"
    }
  ]
}
```

**Usage Pattern**: Called before addCategory/updateCategory to provide early validation feedback

**Error Responses**:
- `400 Bad Request`: Malformed validation request
- `401 Unauthorized`: Session cookie expired
- `500 Internal Server Error`: BCM internal error

---

## Force Parameter Behavior Matrix

The `force` parameter controls how BCM handles warnings and conflicts during category operations. This table documents the exact behavior across different scenarios.

| Operation | Scenario | force=false | force=true | Notes |
|-----------|----------|-------------|------------|-------|
| **Create** | Validation warning (e.g., experimental kernel param) | ❌ Fails with diagnostic | ✅ Proceeds with category creation | Warnings can be overridden |
| **Create** | Validation error (e.g., invalid UUID format) | ❌ Fails with diagnostic | ❌ Fails with diagnostic | Hard errors cannot be overridden |
| **Create** | Duplicate category name | ❌ Fails with 409 Conflict | ❌ Fails with 409 Conflict | Name uniqueness always enforced |
| **Update** | Nodes currently provisioning | ❌ Fails with error | ✅ Proceeds with update | May affect node provisioning state |
| **Update** | Validation warning | ❌ Fails with diagnostic | ✅ Proceeds with update | Warnings can be overridden |
| **Update** | Invalid management_network UUID | ❌ Fails with diagnostic | ❌ Fails with diagnostic | Hard errors cannot be overridden |
| **Update** | Category does not exist | ❌ Fails with 404 | ❌ Fails with 404 | Force cannot create missing categories |
| **Delete** | Category has nodes assigned | ❌ Fails with error + node count | ✅ Proceeds with deletion | Node assignments unaffected, category removed |
| **Delete** | System/protected category | ❌ Fails with error | ❌ Fails with error | System categories cannot be deleted |
| **Delete** | Category does not exist | ❌ Fails with 404 | ❌ Fails with 404 | Already deleted |

### Force Parameter Guidelines

**When to use `force=true`**:
- Override validation warnings during create/update operations
- Delete categories with node assignments (nodes remain but lose category reference)
- Update categories while nodes are actively provisioning (use with caution)

**When `force=true` does NOT help**:
- Hard validation errors (invalid UUID format, malformed XML, etc.)
- Duplicate category name conflicts (use unique names)
- Missing/non-existent categories (404 errors)
- System-protected categories (cannot be deleted regardless)

**Best Practices**:
1. Always run without `force` first to see validation feedback
2. Review validation warnings before using `force=true`
3. Use `force=true` for intentional overrides, not to bypass all validation
4. Document why `force=true` is needed in Terraform configuration comments

---

### 6. removeCategory

Deletes a category.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "removeCategory",
  "args": [
    "0ae6d733-3015-4479-bfab-ce2d237a2809",
    false
  ]
}
```

**Arguments**:
- `args[0]` (string, required): UUID of category to delete
- `args[1]` (boolean, optional): Force flag (default: false)
  - `false`: Fail if category has assigned nodes
  - `true`: Delete category even if nodes are assigned

**Response** (200 OK):
```json
true
```

**Response Type**: Boolean true on success

**Usage Pattern**: Resource Delete operation

**Error Responses**:
- `404 Not Found`: Category UUID not found
- `409 Conflict`: Category has assigned nodes (without force flag)
- `422 Unprocessable Entity`: Category cannot be deleted (system category, etc.)
- `401 Unauthorized`: Session cookie expired
- `500 Internal Server Error`: BCM internal error

**Error Example (nodes assigned)**:
```json
{
  "error": "Category has 5 nodes assigned. Use force=true to override.",
  "code": "CATEGORY_IN_USE"
}
```

---

## Complete Entity Structure

### Category Entity (Full Specification)

```json
{
  "uuid": "string",
  "baseType": "Category",
  "childType": "",
  "modified": false,
  "to_be_removed": false,
  "revision": "string",
  "parent_uuid": "string|null",
  "name": "string",
  "notes": "string",
  "managementNetwork": "string",
  "softwareImageProxy": {
    "uuid": "string",
    "baseType": "SoftwareImageProxy",
    "childType": "",
    "parentSoftwareImage": "string",
    "revisionID": 0
  },
  "bootLoader": "string",
  "bootLoaderFile": "string",
  "bootLoaderProtocol": "string",
  "kernelVersion": "string",
  "kernelParameters": "string",
  "kernelOutputConsole": "string",
  "modules": [
    {
      "baseType": "KernelModule",
      "name": "string",
      "parameters": "string"
    }
  ],
  "disksetup": "string",
  "raidconf": "string",
  "installMode": "string",
  "newNodeInstallMode": "string",
  "installBootRecord": false,
  "ioScheduler": "string",
  "defaultGateway": "string",
  "defaultGatewayMetric": 0,
  "nameServers": ["string"],
  "searchDomains": ["string"],
  "timeServers": ["string"],
  "staticRoutes": [],
  "fsmounts": [
    {
      "uuid": "string",
      "baseType": "FSMount",
      "device": "string",
      "mountpoint": "string",
      "filesystem": "string",
      "mountoptions": "string",
      "fsck": "string",
      "dump": false,
      "rdma": false
    }
  ],
  "fsexports": [],
  "roles": [],
  "services": [],
  "gpuSettings": [],
  "bmcSettings": {
    "uuid": "string",
    "baseType": "BMCSettings",
    "userName": "string",
    "password": "string",
    "privilege": "string",
    "userID": 0,
    "firmwareManageMode": "string",
    "leakPolicy": "string",
    "leakReactionDelay": 0.0,
    "powerResetDelay": 0
  },
  "biosSetup": {},
  "dpuSettings": {},
  "proxySettings": {},
  "accessSettings": {},
  "seLinuxSettings": {},
  "timeZoneSettings": {},
  "ztpSettings": {},
  "initialize": "string",
  "finalize": "string",
  "excludeListFull": "string",
  "excludeListGrab": "string",
  "excludeListGrabnew": "string",
  "excludeListSync": "string",
  "excludeListUpdate": "string",
  "excludeListManipulateScript": "string",
  "authenticationService": "string",
  "allowNetworkingRestart": false,
  "dataNode": false,
  "nodeInstallerDisk": false,
  "versionConfigFiles": false,
  "interactiveUser": "string",
  "useExclusivelyFor": "string",
  "fips": "string"
}
```

---

## Enum Values

### bootLoader
- `SYSLINUX` (default)
- `GRUB`
- `GRUB2`
- `PXELINUX`
- `LOCALBOOT`

### bootLoaderProtocol
- `HTTP` (default)
- `TFTP`
- `NFS`

### installMode
- `AUTO` (default)
- `FULL`
- `MINIMAL`
- `CUSTOM`

### newNodeInstallMode
- `FULL` (default)
- `MINIMAL`
- `SKIP`

### authenticationService
- `AUTO` (default)
- `LDAP`
- `SSSD`
- `LOCAL`

### fips
- `YES`
- `NO`

### BMCSettings.privilege
- `USER`
- `OPERATOR`
- `ADMINISTRATOR`

### BMCSettings.firmwareManageMode
- `AUTO` (default)
- `MANUAL`
- `DISABLED`

---

## Error Handling Best Practices

### 1. Authentication Errors

**Error**: `401 Unauthorized`
**Cause**: Session cookie expired or invalid
**Solution**: Re-authenticate by calling login endpoint

### 2. Validation Errors

**Error**: `422 Unprocessable Entity`
**Cause**: Category configuration fails validation
**Solution**: Parse `validation` array and display errors to user

### 3. Conflict Errors

**Error**: `409 Conflict`
**Cause**: Duplicate category name or operation blocked by state
**Solution**: Prompt user to resolve conflict or use different name

### 4. Not Found Errors

**Error**: `404 Not Found`
**Cause**: Category UUID/name does not exist
**Solution**: Verify identifier or refresh resource list

### 5. Force Required Errors

**Error**: `422 Unprocessable Entity` with message about nodes
**Cause**: Operation requires force=true due to node assignments
**Solution**: Retry with force=true if user confirms

---

## Rate Limiting

BCM API does not implement explicit rate limiting, but best practices:
- Avoid rapid repeated calls (< 100ms between requests)
- Use exponential backoff for retry logic
- Cache category lists when possible

---

## Security Considerations

### 1. Sensitive Data
- BMCSettings.password is sensitive - never log or display in plain text
- Use HTTPS exclusively (self-signed certs require `insecure_skip_verify`)
- Session cookies have HttpOnly flag

### 2. Input Validation
- Always validate UUID format before API calls
- Sanitize user input for disksetup XML and scripts
- Validate IP address formats for gateways and nameservers

### 3. Authentication
- Session cookies expire after inactivity (default: 30 minutes)
- Re-authenticate automatically on 401 errors
- Store credentials securely (environment variables, not hardcoded)

---

## Performance Characteristics

| Operation | Typical Latency | Notes |
|-----------|----------------|-------|
| getCategories | < 1 second | Returns full list (dozens of items) |
| getCategory | < 500 ms | Direct lookup by name |
| addCategory | 1-3 seconds | Includes database transaction |
| updateCategory | 1-3 seconds | May trigger node updates |
| validateCategory | < 1 second | No persistence, fast validation |
| removeCategory | 1-2 seconds | Checks node assignments |

---

## Testing Recommendations

### Unit Tests
- Mock API responses for each method
- Test error handling for all error codes
- Validate request payload structure

### Integration Tests
- Test full CRUD cycle with real BCM instance
- Verify nested object persistence
- Test force parameter behavior
- Validate import by UUID workflow

### Acceptance Tests
- Use `TF_ACC=1` environment variable
- Set BCM credentials via environment variables
- Clean up resources in test teardown
- Use unique category names per test run

---

## Changelog

| Date | Change | Author |
|------|--------|--------|
| 2025-11-21 | Initial API contract documentation | BCM Provider Team |

---

## Related Documentation

- BCM API Reference: `sampleRest/CMDevice_Complete_Documentation.md`
- Terraform Provider: `/workspace/internal/provider/resource_cmdevice_category.go`
- Data Model: `/workspace/specs/001-cmdevice-category/data-model.md`
- Research: `/workspace/specs/001-cmdevice-category/research.md`
