# BCM CMNet API Contracts

**Service**: cmnet
**Base URL**: https://172.21.15.254:8081/json
**Auth**: Cookie-based (cm-login-token)

## Methods

### addNetwork (Create)
**Request**: `{"service": "cmnet", "call": "addNetwork", "args": [entity]}`
**Response**: Created entity with UUID
**Errors**: 409 if name exists

### getNetwork (Read)
**Request**: `{"service": "cmnet", "call": "getNetwork", "args": [uuid]}`
**Response**: Network entity
**Errors**: 404 if not found

### updateNetwork (Update)
**Request**: `{"service": "cmnet", "call": "updateNetwork", "args": [entity]}`
**Response**: Updated entity
**Errors**: 409 revision conflict

### removeNetwork (Delete)
**Request**: `{"service": "cmnet", "call": "removeNetwork", "args": [uuid]}`
**Response**: Empty
**Errors**: 409 if active assignments

## Entity Structure

**Create** (no UUID):
```json
{
  "name": "network-name",
  "baseAddress": "10.0.1.0",
  "netmaskBits": 24,
  "gateway": "10.0.1.1",
  "mtu": 1500,
  "domainName": "cluster.local",
  "dynamicRangeStart": "10.0.1.100",
  "dynamicRangeEnd": "10.0.1.200",
  "notes": "Notes",
  "baseType": "Network",
  "childType": "",
  "modified": true,
  "to_be_removed": false,
  "revision": ""
}
```

**Update** (includes UUID + revision):
```json
{
  "uuid": "uuid-here",
  "revision": "revision-here",
  ... all other fields ...
}
```
