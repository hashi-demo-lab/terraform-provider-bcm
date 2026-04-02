# Test Coverage Report

Terraform Provider BCM — test coverage analysis based on
[HashiCorp provider testing patterns](https://developer.hashicorp.com/terraform/plugin/testing/testing-patterns).

## Summary

| Metric | Count |
|--------|-------|
| **Total tests** | 406 |
| **Acceptance tests** | 304 |
| **Unit tests** | 102 |
| **Test files** | 36 |
| **Resources tested** | 7/7 (100%) |
| **Data sources tested** | 9/9 (100%) |
| **Actions tested** | 1/1 (100%) |

## Resource Test Coverage

### Acceptance Tests by Resource

| Resource | Tests | Basic | Update | Import | Drift | Disappears | Validation | Idempotency |
|----------|------:|------:|-------:|-------:|------:|-----------:|-----------:|------------:|
| `bcm_cmdevice_device` | 76 | 2 | 12 | 6 | 8 | 1 | 9 | 9 |
| `bcm_cmdevice_category` | 39 | 7 | 4 | 2 | 3 | 1 | 5 | 1 |
| `bcm_cmkube_cluster` | 25 | 3 | 5 | 1 | 2 | 1 | 4 | 0 |
| `bcm_cmpart_softwareimage` | 21 | 2 | 4 | 0 | 4 | 1 | 1 | 1 |
| `bcm_cmuser_user` | 14 | 1 | 1 | 1 | 2 | 1 | 4 | 1 |
| `bcm_cmetcd_cluster` | 9 | 1 | 1 | 1 | 1 | 1 | 3 | 0 |
| `bcm_cmnet_network` | 8 | 2 | 2 | 1 | 1 | 1 | 2 | 0 |
| **Total** | **192** | **18** | **29** | **12** | **21** | **7** | **28** | **12** |

All 7 resources have complete scenario coverage across Basic, Update, Import,
Drift, Disappears, and Validation.

### Data Source Tests

| Data Source | Tests |
|-------------|------:|
| `bcm_cmpart_entity_info` | 11 |
| `bcm_cmpart_partitions` | 9 |
| `bcm_cmpart_softwareimages` | 7 |
| `bcm_cmkube_clusters` | 7 |
| `bcm_cmuser_users` | 6 |
| `bcm_cmdevice_roles` | 5 |
| `bcm_cmdevice_categories` | 4 |
| `bcm_cmdevice_nodes` | 4 |
| `bcm_cmnet_networks` | 4 |
| **Total** | **57** |

### Action Tests

| Action | Tests |
|--------|------:|
| `bcm_cmdevice_power` | 10 |

## Modern Pattern Adoption

Evaluated against [terraform-plugin-testing](https://github.com/hashicorp/terraform-plugin-testing)
best practices and the `provider-test-patterns` skill reference.

### Pattern Usage Across Resource Tests

| Pattern | Files Using | Total Uses | Status |
|---------|:----------:|:----------:|--------|
| `ConfigStateChecks` (modern assertions) | 11/14 | 320+ | Adopted |
| `ConfigPlanChecks` (plan assertions) | 11/14 | 146+ | Adopted |
| `CompareValue` (cross-step tracking) | 11/14 | 114+ | Adopted |
| `PreConfig` (drift simulation via API) | 10/14 | 24 | Adopted |
| Custom `statecheck.StateCheck` (disappears) | 7/7 | 8 | Complete |
| `CheckDestroy` | 7/7 resources | 100% | 100% |
| `ExpectError` (validation) | 7/7 | 75+ | Adopted |
| Schema validators | 6/7 resources | 79 | Adopted |
| Numbered format verbs (`%[1]q`) | 14/14 | — | Complete |
| `resource.Test` (serial) | 14/14 | — | All serial |
| `resource.ParallelTest` | 0/14 | 0 | Not used (shared BCM cluster) |

### Custom StateCheck Implementations

Every resource has a dedicated disappears check following the
`statecheck.StateCheck` interface pattern:

| Resource | Implementation |
|----------|---------------|
| `bcm_cmdevice_device` | `deviceDisappearsCheck` |
| `bcm_cmdevice_category` | `categoryDisappearsCheck` |
| `bcm_cmkube_cluster` | `kubeClusterDisappearsCheck` |
| `bcm_cmpart_softwareimage` | `softwareImageDisappearsCheck` |
| `bcm_cmuser_user` | `userDisappearsCheck` |
| `bcm_cmetcd_cluster` | `etcdClusterDisappearsCheck` |
| `bcm_cmnet_network` | `networkDisappearsCheck` |

### CheckDestroy Coverage

All validation-only tests (`ExpectError`) include `CheckDestroy` even though
no resources are created, for consistency.

| Resource | CheckDestroy | Coverage |
|----------|:-----------:|:--------:|
| `bcm_cmpart_softwareimage` | 21/21 | 100% |
| `bcm_cmuser_user` | 14/14 | 100% |
| `bcm_cmetcd_cluster` | 9/9 | 100% |
| `bcm_cmnet_network` | 8/8 | 100% |
| `bcm_cmdevice_category` | 39/39 | 100% |
| `bcm_cmdevice_device` | 68/76 | 89%* |
| `bcm_cmkube_cluster` | 23/25 | 92% |

*Device 89%: the 8 uncovered tests are `t.Skip()` documentation stubs with no
`resource.Test()` call — every test that creates a `TestCase` has `CheckDestroy`.

## Test Infrastructure

### Provider Factory

```go
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
    "bcm": providerserver.NewProtocol6WithError(New("test")()),
}
```

### Key Test Helpers

| Helper | Purpose |
|--------|---------|
| `createTestBCMClient(t)` | Authenticated BCM client for PreConfig drift simulation |
| `testAccCMDeviceDevicePreCheck(t, names)` | Pre-test cleanup of leftover resources |
| `verifyResourceDeleted(ctx, client, ...)` | Exponential backoff polling for eventual consistency |
| `generateShortTestName(prefix)` | RFC 1123 compliant names (<63 chars) |
| `generateUniqueMAC()` | Locally-administered MAC addresses |
| `getTestNetworkUUID(t, name)` | BCM network UUID lookup by name |

### Naming Conventions

- Test functions: `TestAcc{Resource}_{Scenario}` (e.g., `TestAccCMDeviceDevice_ImportAndRecategorize`)
- Config helpers: `testAcc{Resource}Config_{Variant}` with numbered format verbs
- Custom checks: `{resource}DisappearsCheck` implementing `statecheck.StateCheck`

### Pre-push Hooks

The following checks run before `git push` via pre-commit:

- `go build ./...` — compilation
- `go vet ./...` — static analysis
- `make install && make generate` + docs diff — stale doc detection

## Design Decisions

**Serial tests (`resource.Test`)**: All tests run sequentially because they
share a single BCM cluster. Parallel execution would cause resource name
collisions and API contention.

**Import verify ignore lists**: Device import tests use explicit
`ImportStateVerifyIgnore` lists because BCM returns computed defaults
(e.g., `boot_loader: "CATEGORY"`, `power_control: "none"`) that differ from
Terraform's null representation for unset Optional+Computed fields.

**Eventual consistency**: BCM operations are asynchronous. `CheckDestroy`
implementations use exponential backoff (1s, 2s, 4s, 8s) via
`verifyResourceDeleted()`. Drift tests use `TestEventualConsistencyDelay`
between API modifications and plan checks.

**Multi-phase CheckDestroy**: Device destroy verifies deletion in dependency
order — devices first, then categories, then software images — to handle
BCM's referential constraints.
