# CLAUDE.md

Terraform Provider for Nvidia BCM (Bright Cluster Manager) built on Plugin Framework v1.16.1.

## Quick Reference

```bash
# Build & Install
make install              # Build, lint, install, generate docs

# Testing
make test                 # Unit tests
TF_ACC=1 go test -v -timeout 120m ./internal/provider/ -run TestAccName

# Code Quality
make fmt && make lint     # Format and lint
pre-commit run --all-files

# Documentation
make generate             # Generate docs (don't edit docs/ manually)
```

**Test Environment:**
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

## Project Structure

```
internal/provider/
├── provider.go           # Provider config
├── bcm_client.go         # JSON-RPC client
├── resource_*.go         # Resources (CRUD)
├── data_source_*.go      # Data sources
└── *_test.go             # Tests
examples/                 # Terraform examples
docs/                     # Generated (don't edit)
specs/                    # Speckit specifications
```

## Architecture

### BCM API Pattern

```go
// JSON-RPC call pattern
client.CallJSONRPC(ctx, "cmdevice", "getNodes", args...)

// Data sources: list all, filter client-side
getSoftwareImages() → filter in Go

// Resources: direct lookup with args
getSoftwareImage(name) → single result
```

**Authentication:** Cookie-based (`cm-login-token`), auto-managed.

### Resource Pattern

1. **Create** - `addResource()`, handle async ops with polling
2. **Read** - Direct lookup with args parameter
3. **Update** - `updateResource()` with full entity + UUID
4. **Delete** - `removeResource()`
5. **Import** - `resource.ImportStatePassthroughID`

**Critical Rules:**
- NEVER propagate `Unknown` values to state
- Preserve plan values for fields BCM resets (e.g., `original_image`)
- Use exponential backoff for eventual consistency

### Validation

```go
// Pre-flight validation before CREATE/UPDATE
validationErrors, _ := r.client.ValidateEntity(ctx, "CMDevice", "validateCategory", entity, isCreate)
```

| Resource | Service | Method |
|----------|---------|--------|
| Software Images | CMPart | validateSoftwareImage |
| Categories | CMDevice | validateCategory |
| Devices | CMDevice | validateDevice |
| Networks | CMNet | validateNetwork |
| Kubernetes | **cmkube** (lowercase) | validateKubeCluster |

## BCM Quirks

### Non-Persisted Category Fields

BCM API accepts but doesn't store these fields:

| Field | Workaround |
|-------|------------|
| `static_routes` | Preserve plan values |
| `fsexports` | Preserve plan values |
| `roles` | Generate UUIDs locally |
| `gpu_settings` | Preserve plan values |
| `services` | Preserve plan values |

**Reference:** [Issue #73](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/73)

### disksetup XML

Root element must be `<diskSetup>` (camelCase). See [Issue #48](https://github.com/hashi-demo-lab/terraform-provider-bcm/issues/48).

### Field Mappings (snake_case → camelCase)

Key mappings for drift detection:
- `kernel_parameters` → `kernelParameters`
- `home_directory` → `homeDirectory`
- `login_shell` → `loginShell`
- `authorized_ssh_keys` → `authorizedSshKeys`

## TDD Workflow

See **AGENTS.md** for complete TDD patterns including:
- RED-GREEN-REFACTOR cycles
- Drift detection tests
- Modern testing patterns (statecheck, plancheck)
- CheckDestroy patterns

### Speckit Commands

```bash
/speckit.specify   # Create spec
/speckit.clarify   # Ask clarifications
/speckit.plan      # Generate plan
/speckit.tasks     # Generate tasks
/speckit.implement # Execute tasks
```

## Troubleshooting

**Unknown values error:** Resolve to actual value or `null`, never propagate `Unknown`.

**Read strategy:** Data sources list+filter, resources direct lookup with args.

**Auth failures:** Check `cm-login-token` cookie, use `insecure_skip_verify = true`.

## References

- **TDD Patterns:** `./AGENTS.md`
- **BCM API Docs:** `sampleRest/CMDevice_Complete_Documentation.md`
- **Skills:** `terraform-provider-tests`, `terraform-provider-design`
