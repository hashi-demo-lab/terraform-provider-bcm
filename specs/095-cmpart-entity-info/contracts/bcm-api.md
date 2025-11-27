# BCM API Contract: getBasicEntityInformation

**Feature**: `095-cmpart-entity-info`
**Date**: 2025-11-27

## Endpoint

### Request

**Method**: POST
**URL**: `{BCM_ENDPOINT}/json`
**Content-Type**: `application/json`
**Authentication**: Cookie-based (`cm-login-token`)

```json
{
  "service": "cmpart",
  "call": "getBasicEntityInformation"
}
```

### Response

**Status**: 200 OK
**Content-Type**: `application/json`

#### Success Response

Array of entity objects:

```json
[
  {
    "resolveName": "default-image",
    "type": "SoftwareImage",
    "uuid": "8482c4e9-383c-43de-873f-8c54ee77ee74"
  },
  {
    "resolveName": "default",
    "type": "Category",
    "uuid": "a1b2c3d4-e5f6-7890-abcd-ef1234567890"
  },
  {
    "resolveName": "managementnet",
    "type": "Network",
    "uuid": "21b20743-d055-42c6-b03c-583c0c061e2e"
  }
]
```

#### Response Schema

```json
{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "type": "array",
  "items": {
    "type": "object",
    "required": ["resolveName", "type", "uuid"],
    "properties": {
      "resolveName": {
        "type": "string",
        "description": "Human-readable entity name"
      },
      "type": {
        "type": "string",
        "description": "Entity type classification"
      },
      "uuid": {
        "type": "string",
        "format": "uuid",
        "description": "Unique entity identifier"
      }
    }
  }
}
```

## Error Responses

### Authentication Error (401)

```json
{
  "error": "unauthorized",
  "code": 401
}
```

### Server Error (500)

```json
{
  "error": "internal server error",
  "message": "description of error"
}
```

## Field Reference

| Field | Type | Description | Example |
|-------|------|-------------|---------|
| `resolveName` | string | Display name for the entity. May be empty string. | `"default-image"` |
| `type` | string | Entity type classification. Always non-empty. | `"SoftwareImage"` |
| `uuid` | string | UUID v4 format identifier. Always non-empty. | `"8482c4e9-383c-43de-873f-8c54ee77ee74"` |

## Common Entity Types

| Type | Description |
|------|-------------|
| `Category` | Node grouping category |
| `SoftwareImage` | OS image for provisioning |
| `Network` | Network configuration |
| `HeadNode` | Cluster head node |
| `PhysicalNode` | Physical compute node |
| `Partition` | Resource partition |
| `Rack` | Physical rack |
| `ConfigurationOverlay` | Configuration overlay |
| `FSPart` | Filesystem partition |
| `Profile` | Node profile |
| `Role` | Role definition |
| `KubeCluster` | Kubernetes cluster |

## Usage Notes

1. **No Arguments Required**: The API returns all entities without any filtering parameters.

2. **Large Response**: Typical clusters return 500+ entities. Client-side filtering recommended.

3. **Case Sensitivity**: The `type` field is case-sensitive (PascalCase).

4. **Empty resolveName**: Some entities may have empty `resolveName` values.

## Sample cURL Request

```bash
curl -X POST https://172.21.15.254:8081/json \
  -H "Content-Type: application/json" \
  -H "Cookie: cm-login-token=YOUR_TOKEN" \
  -d '{"service":"cmpart","call":"getBasicEntityInformation"}' \
  -k
```

## Go Implementation Reference

```go
// Call BCM API
body, err := client.CallJSONRPC(ctx, "cmpart", "getBasicEntityInformation")
if err != nil {
    return fmt.Errorf("API call failed: %w", err)
}

// Parse response
var entities []map[string]interface{}
if err := json.Unmarshal(body, &entities); err != nil {
    return fmt.Errorf("failed to parse response: %w", err)
}

// Map to Terraform model
for _, e := range entities {
    model := EntityInfoModel{
        Name: getStringValue(e, "resolveName"),
        Type: getStringValue(e, "type"),
        UUID: getStringValue(e, "uuid"),
    }
    // Apply filters and add to result...
}
```
