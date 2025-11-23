# BCM Test Resource Cleanup Scripts

## Overview

Two scripts for cleaning up orphaned test resources from BCM cluster:
- **Interactive cleanup** - Manual review and confirmation
- **Automated cleanup** - CI/CD integration (no prompts)

Both scripts clean up resources with `citest-*` or `tftest-*` prefixes.

## Resource Types Cleaned

| Resource Type | Service | Naming Field | Dependencies |
|---------------|---------|--------------|--------------|
| Software Images | CMPart | name | None (delete last) |
| Categories | CMDevice | name | Depends on images |
| Devices | CMDevice | hostname | Depends on categories |
| Kubernetes Clusters | CMKube | name | None |
| Networks | CMNet | name | None |

## Interactive Cleanup (Manual)

**Use for**: Manual cleanup, review before deletion

```bash
# Set credentials
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="your-password"

# Run interactive cleanup
./scripts/cleanup-basic-resources.sh
```

**Features**:
- Lists all matching resources
- Asks for confirmation before each resource type
- Safe for manual use

**Example output**:
```
Software Images to delete:
  - tftest-image-20251124-150405-123456 [uuid-1234...]
  - tftest-base-image-20251124-151234-789012 [uuid-5678...]

Delete these software images? (y/N)
```

## Automated Cleanup (CI/CD)

**Use for**: Automated cleanup in CI/CD pipelines

```bash
# Set credentials
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="your-password"

# Enable automated cleanup (safety check)
export AUTO_CLEANUP=1

# Run automated cleanup
./scripts/cleanup-test-resources-auto.sh
```

**Features**:
- No prompts or user interaction
- Deletes all matching resources automatically
- Proper dependency ordering (devices before categories, etc.)
- Requires `AUTO_CLEANUP=1` as safety check

**Example output**:
```
==========================================
Automated BCM Test Resource Cleanup
Prefixes: citest-*, tftest-*
==========================================

✓ Logged in successfully

→ Checking Devices...
  Found 3 Devices to delete:
    - tftest-device-20251124-150405 [uuid-1234]
    - tftest-compute-node-01-20251124-151000 [uuid-5678]
    - tftest-gpu-node-01-20251124-152000 [uuid-9012]
  ✓ Devices deleted

→ Checking Kubernetes Clusters...
  Found 1 Kubernetes Clusters to delete:
    - tftest-cluster-20251124-150500 [uuid-3456]
  ✓ Kubernetes Clusters deleted

...

==========================================
✓ Automated cleanup complete!
==========================================
```

## Naming Convention

| Prefix | Usage | Cleanup Method |
|--------|-------|----------------|
| `tftest-*` | Terraform acceptance tests | Both scripts |
| `citest-*` | CI/CD and example tests | Both scripts |
| Other | Production resources | **NOT touched** |

**Test naming pattern**:
```go
// In test files
imageName := generateUniqueTestName("tftest-image")
// Creates: tftest-image-20251124-150405-123456789

categoryName := generateUniqueTestName("tftest-category")
// Creates: tftest-category-20251124-150405-987654321
```

## When to Run Cleanup

**Manual cleanup needed when**:
- ✅ Tests interrupted with Ctrl+C
- ✅ Test runner killed or timed out
- ✅ Test failed mid-execution
- ✅ CheckDestroy never ran

**Automated cleanup integration**:
```yaml
# GitHub Actions example
- name: Cleanup test resources
  if: always()  # Run even if tests fail
  env:
    BCM_ENDPOINT: ${{ secrets.BCM_ENDPOINT }}
    BCM_USERNAME: ${{ secrets.BCM_USERNAME }}
    BCM_PASSWORD: ${{ secrets.BCM_PASSWORD }}
    AUTO_CLEANUP: 1
  run: ./scripts/cleanup-test-resources-auto.sh
```

## Safety Features

### Interactive Script
- ✅ Requires user confirmation for each resource type
- ✅ Shows all resources before deletion
- ✅ Can skip individual resource types

### Automated Script
- ✅ Requires `AUTO_CLEANUP=1` environment variable
- ✅ Only deletes resources with test prefixes
- ✅ Proper dependency ordering
- ✅ Verbose logging of all actions

## Troubleshooting

### Login fails
```bash
# Check credentials
curl -k -X POST "${BCM_ENDPOINT}/json" \
  -H "Content-Type: application/json" \
  -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}"
```

### No resources found but you know they exist
```bash
# Check resource naming
curl -k -s -b cookies.txt -X POST "${BCM_ENDPOINT}/json" \
  -H "Content-Type: application/json" \
  -d '{"service":"CMPart","call":"getSoftwareImages"}' | jq -r '.[].name'

# If names don't start with tftest- or citest-, they won't be cleaned up
```

### Automated script won't run
```bash
# Did you set AUTO_CLEANUP=1?
export AUTO_CLEANUP=1
./scripts/cleanup-test-resources-auto.sh

# Safety check prevents accidental execution
```

### Resources have dependencies
The automated script deletes in proper order:
1. Devices (depend on categories)
2. Kubernetes clusters (independent)
3. Networks (independent)
4. Categories (depend on software images)
5. Software images (no dependencies)

If manual deletion fails due to dependencies, follow this order.

## Resource Count Monitoring

```bash
# Check how many test resources exist
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="your-password"

# Login
COOKIE=$(mktemp)
curl -k -s -c "$COOKIE" -X POST "${BCM_ENDPOINT}/json" \
  -H "Content-Type: application/json" \
  -d "{\"service\":\"login\",\"username\":\"$BCM_USERNAME\",\"password\":\"$BCM_PASSWORD\"}" > /dev/null

# Count test resources
echo "Software Images:"
curl -k -s -b "$COOKIE" -X POST "${BCM_ENDPOINT}/json" \
  -d '{"service":"CMPart","call":"getSoftwareImages"}' | \
  jq '[.[] | select(.name | startswith("citest-") or startswith("tftest-"))] | length'

echo "Categories:"
curl -k -s -b "$COOKIE" -X POST "${BCM_ENDPOINT}/json" \
  -d '{"service":"CMDevice","call":"getCategories"}' | \
  jq '[.[] | select(.name | startswith("citest-") or startswith("tftest-"))] | length'

echo "Devices:"
curl -k -s -b "$COOKIE" -X POST "${BCM_ENDPOINT}/json" \
  -d '{"service":"CMDevice","call":"getNodes"}' | \
  jq '[.[] | select(.hostname | startswith("citest-") or startswith("tftest-"))] | length'

rm -f "$COOKIE"
```

## See Also

- Test naming convention: `/workspace/internal/provider/test_helpers.go`
- Test gap analysis: `/workspace/.claude/skills/terraform-provider-tests/`
- Example tests: `/workspace/examples/`
