# BCM Power API Contract

**Service**: cmdevice
**Feature**: bcm_cmdevice_power Action
**Version**: 1.0.0

## Overview

This contract defines the BCM JSON-RPC API methods used by the `bcm_cmdevice_power` Terraform Action for device power control operations.

## Authentication

All API calls require prior authentication. The BCM client handles this via cookie-based authentication with `cm-login-token`.

## Endpoint

```
POST https://{bcm_endpoint}/json
Content-Type: application/json
Cookie: cm-login-token={session_token}
```

## Power Operation Methods

### powerOn

Power on a device via BMC/IPMI.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "powerOn",
  "args": ["{device_identifier}"]
}
```

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| device_identifier | string | Yes | Device hostname or UUID |

**Response (Success)**:
```json
true
```
or status object (to be verified in Phase 0).

**Response (Error)**:
```json
{
  "error": "Device not found",
  "code": 404
}
```

---

### powerOff

Power off a device via BMC/IPMI.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "powerOff",
  "args": ["{device_identifier}"]
}
```

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| device_identifier | string | Yes | Device hostname or UUID |

**Response**: Same as powerOn.

---

### reboot

Reboot a device (documented in BCM API).

**Request**:
```json
{
  "service": "cmdevice",
  "call": "reboot",
  "args": ["{device_identifier}"]
}
```

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| device_identifier | string | Yes | Device hostname or UUID |

**Response**: Same as powerOn.

---

### powerCycle

Power cycle a device (hard off/on).

**Request**:
```json
{
  "service": "cmdevice",
  "call": "powerCycle",
  "args": ["{device_identifier}"]
}
```

**Parameters**:
| Name | Type | Required | Description |
|------|------|----------|-------------|
| device_identifier | string | Yes | Device hostname or UUID |

**Response**: Same as powerOn.

---

## Status Query Method (for wait_for_completion)

### getNode (or powerStatus)

Query device to check power state.

**Request**:
```json
{
  "service": "cmdevice",
  "call": "getNode",
  "args": ["{device_identifier}"]
}
```

**Response**:
```json
{
  "baseType": "Device",
  "childType": "PhysicalNode",
  "hostname": "node001",
  "uuid": "2870c0b0-6fda-4026-9b8f-28be4c372fee",
  "powerControl": "ipmi",
  "powerStatus": "on|off|unknown"
}
```

**Note**: The `powerStatus` field existence needs Phase 0 verification.

---

## Error Responses

### Device Not Found

**HTTP Status**: 200 (BCM uses JSON error envelope)

**Response**:
```json
{
  "error": "Device 'unknown-node' not found",
  "code": "NOT_FOUND"
}
```

### BMC Unreachable

**Response**:
```json
{
  "error": "Failed to contact BMC for device 'node001': Connection timed out",
  "code": "BMC_ERROR"
}
```

### Authentication Error

**HTTP Status**: 401

**Response**:
```json
{
  "error": "Session expired",
  "code": 401
}
```

---

## Terraform Action to API Mapping

| Terraform power_action | BCM API Method |
|------------------------|----------------|
| `power_on` | `powerOn` |
| `power_off` | `powerOff` |
| `reboot` | `reboot` |
| `power_cycle` | `powerCycle` |

---

## Go Client Usage

```go
// Power on device
_, err := client.CallJSONRPC(ctx, "cmdevice", "powerOn", deviceID)

// Power off device
_, err := client.CallJSONRPC(ctx, "cmdevice", "powerOff", deviceID)

// Reboot device
_, err := client.CallJSONRPC(ctx, "cmdevice", "reboot", deviceID)

// Power cycle device
_, err := client.CallJSONRPC(ctx, "cmdevice", "powerCycle", deviceID)

// Query device for power status
body, err := client.CallJSONRPC(ctx, "cmdevice", "getNode", deviceID)
var device map[string]interface{}
json.Unmarshal(body, &device)
powerStatus := device["powerStatus"]
```

---

## Verification Checklist

Phase 0 verification results (from existing API exploration):

- [x] `reboot` method verified - Returns `{"operations": [], "success": true}` (from cmdevice_discovered_methods)
- [ ] `powerOn` method - Listed in BCM docs, failed with null arg (needs device_id)
- [ ] `powerOff` method - Listed in BCM docs, failed with null arg (needs device_id)
- [ ] `powerCycle` method - Listed in BCM docs, likely available
- [x] Power status query method - Use `getNode` and check device fields
- [x] Error response format documented

## Phase 0 Verification Notes (2025-11-26)

### Verified from cmdevice_discovered_methods_20251120_175345.json

The `reboot` method was tested and returned:
```json
{
  "operations": [],
  "success": true
}
```

This confirms the response format for successful power operations.

### Failed Methods Analysis

`powerOn`, `powerOff`, `powerStatus` failed in automated testing with null args.
These methods likely require a device identifier argument (hostname or UUID).

Based on BCM API patterns and the documented `reboot` behavior:
1. All power methods accept a single device identifier argument
2. Successful response includes `{"operations": [], "success": true}`
3. Device identifier can be hostname or UUID

### Implementation Decision

Proceed with implementation using the verified `reboot` pattern:
- Service: `cmdevice`
- Args: `[deviceID]` (string - hostname or UUID)
- Expected response: `{"operations": [], "success": true}`

Full verification against live BCM API is recommended during Phase 5 manual testing.
