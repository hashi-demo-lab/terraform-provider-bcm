# Research: Production-Ready Codebase Review Methodologies

**Feature**: Production-Ready Codebase Review
**Date**: 2025-11-22
**Phase**: Phase 0 - Research & Discovery

## Purpose

This document defines the methodologies, tools, and validation approaches for conducting a comprehensive production-readiness review of the Terraform Provider BCM codebase. The research addresses five key analysis areas identified in the feature specification.

---

## 1. Test Coverage Analysis Methodology

### Objective

Systematically identify missing CRUD, Import, Drift, and Idempotency tests across all resources and data sources.

### Research Question

How to systematically identify missing CRUD, Import, Drift, and Idempotency tests? What grep/glob patterns detect modern vs legacy testing patterns? How to parse test files to identify specific test step patterns?

### Decision: Pattern-Based Code Analysis

Use Glob for file discovery and Grep for pattern matching across all test files to identify specific test coverage gaps.

### Rationale

- **Comprehensive**: Glob pattern `resource_*.go` and `resource_*_test.go` ensures all resources are discovered
- **Precise**: Grep can identify specific test patterns like `ImportState: true`, `PreConfig`, `plancheck.ExpectEmptyPlan()`
- **Scalable**: Automated analysis scales to any number of resources without manual inspection
- **Objective**: Pattern matching eliminates subjective assessment of test quality

### Alternatives Considered

1. **Manual test file review** - Rejected: Time-consuming, error-prone, not repeatable
2. **AST parsing with go/ast** - Rejected: Overkill for pattern detection, requires complex parsing logic
3. **Test execution with coverage** - Rejected: Only shows code coverage, not test type coverage (CRUD/Import/Drift)

### Tools & Techniques

#### File Discovery (Glob)

```bash
# Find all resource implementation files
Glob: internal/provider/resource_*.go

# Find all resource test files
Glob: internal/provider/resource_*_test.go

# Find all data source implementation files
Glob: internal/provider/data_source_*.go

# Find all data source test files
Glob: internal/provider/data_source_*_test.go
```

#### Test Pattern Detection (Grep)

**CRUD Coverage Patterns:**

```bash
# Create test: Look for TestStep with Config and Create-like name
Grep: 'Config:\s+testAccResourceConfig.*Create'
Grep: 'TestStep.*{.*Config'  # Generic create step pattern

# Read test: Verify TestCheckResourceAttr or statecheck patterns
Grep: 'resource\.TestCheckResourceAttr'
Grep: 'statecheck\.ExpectKnownValue'

# Update test: Look for second Config with different value
Grep: 'Config:\s+testAccResourceConfig.*Update'

# Delete test: Verify CheckDestroy function exists
Grep: 'CheckDestroy:\s+testAccCheck.*Destroy'
```

**Import Test Pattern:**

```bash
# Import test step with ImportState: true
Grep: 'ImportState:\s+true'
Grep: 'ImportStateVerify:\s+true'
```

**Drift Test Pattern:**

```bash
# Drift test has PreConfig function that modifies resource externally
Grep: 'PreConfig:\s+func\(\)'
Grep: 'plancheck\.ExpectNonEmptyPlan'  # Verifies plan detects drift
```

**Idempotency Test Pattern:**

```bash
# Idempotency test verifies no plan after Create/Update
Grep: 'plancheck\.ExpectEmptyPlan'
```

**Modern vs Legacy Testing Patterns:**

```bash
# Modern patterns (terraform-plugin-testing v1.13.3+)
Grep: 'statecheck\.ExpectKnownValue'
Grep: 'knownvalue\.StringExact'
Grep: 'knownvalue\.Bool'
Grep: 'compareValue.*ValuesSame'
Grep: 'tfjsonpath\.New'

# Legacy patterns (should be migrated)
Grep: 'resource\.TestCheckResourceAttr'  # Count occurrences
Grep: 'resource\.TestCheckResourceAttrSet'
```

**Environment Portability Issues:**

```bash
# Hardcoded counts
Grep: 'TestCheckResourceAttr.*".*\.#",\s*"[0-9]+"'

# Hardcoded names
Grep: 'Config.*"default"'  # Literal strings in config
Grep: 'Config.*"test-'  # Non-unique test names

# Hardcoded UUIDs
Grep: '[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}'
```

#### Analysis Workflow

1. **Discover Files**: Use Glob to find all resource/data source files and their tests
2. **Pair Matching**: Match each implementation file with its test file (e.g., `resource_cmdevice_category.go` → `resource_cmdevice_category_test.go`)
3. **Pattern Analysis**: For each test file, run all Grep patterns to detect coverage
4. **Gap Identification**: Compare required patterns vs found patterns to identify gaps
5. **Report Generation**: Create matrix showing CRUD/Import/Drift/Idempotency status per resource

### Validation Approach

**Success Criteria:**
- 100% of resources analyzed (all `resource_*.go` files have corresponding test analysis)
- Gaps identified with specific file paths and line numbers
- Testing pattern classification (modern vs legacy) complete
- Environment portability issues flagged

**Validation Steps:**
1. Cross-reference discovered resources against provider.go Resources() registration
2. Spot-check 3 test files manually to verify Grep patterns are accurate
3. Validate that identified gaps match actual test file contents

---

## 2. BCM API Discovery Approach

### Objective

Query live BCM API to enumerate all available services and methods, then compare against provider implementation.

### Research Question

How to query BCM API to enumerate all available services and methods? What JSON-RPC introspection calls reveal API capabilities? How to compare discovered API methods against provider implementation?

### Decision: Multi-Source API Discovery with Live Validation

Combine authenticated BCM API exploration with cross-referencing against implementation files and sampleRest/ documentation.

### Rationale

- **Authoritative**: Live BCM API is source of truth for available functionality
- **Complete**: Systematic exploration discovers all services (CMDevice, CMPart, CMNet, CMProv, CMJob, CMServ, CMMon)
- **Validated**: Cross-reference against working examples in sampleRest/ ensures method signatures are correct
- **Actionable**: Direct comparison with provider.go Resources() shows gaps

### Alternatives Considered

1. **sampleRest/ docs only** - Rejected: May be incomplete or outdated, not authoritative
2. **Introspection API** - Rejected: BCM doesn't appear to have JSONRPC introspection endpoint
3. **Network traffic analysis** - Rejected: Complex, requires packet capture, less reliable

### Tools & Techniques

#### BCM Client Authentication

Use existing BCMClient with test credentials:

```go
// In Go test or analysis script
endpoint := os.Getenv("BCM_ENDPOINT")  // https://172.21.15.254:8081
username := os.Getenv("BCM_USERNAME")  // root
password := os.Getenv("BCM_PASSWORD")  // Hashicorp123!

client, err := NewBCMClient(ctx, endpoint, username, password, true, 30)
if err != nil {
    log.Fatalf("Authentication failed: %v", err)
}
```

#### Service Discovery Strategy

**Known BCM Services** (from codebase analysis):
- `CMDevice` - Device management (nodes, categories, etc.)
- `CMPart` - Software partitions (images, modules)
- `CMNet` - Network management
- `CMProv` - Provisioning
- `CMJob` - Job management
- `CMServ` - Service management
- `CMMon` - Monitoring

**Method Enumeration Approach:**

Since BCM doesn't have introspection, use systematic exploration:

1. **Review sampleRest/ scripts** - Extract all known methods from Python examples
2. **Query known list methods** - Start with methods like `getNodes`, `getSoftwareImages`, `getCategories`
3. **Analyze response formats** - Identify patterns for add/update/remove operations
4. **Cross-reference provider code** - Check internal/provider/*.go for additional BCM API calls

#### Method Discovery Patterns

```bash
# Extract BCM API calls from sampleRest/ Python scripts
Grep: 'call.*getSoftwareImages'
Grep: 'call.*addSoftwareImage'
Grep: 'call.*updateSoftwareImage'
Grep: 'call.*removeSoftwareImage'

# Extract BCM API calls from provider code
Grep: 'CallJSONRPC.*"CMDevice"'
Grep: 'CallJSONRPC.*"CMPart"'
Grep: 'CallJSONRPC.*"CMNet"'
```

#### Known BCM API Methods (from codebase)

**CMDevice Service:**
- `getNodes` - List all nodes (data source)
- `getCategories` - List all categories (data source)
- `getCategory(uuid)` - Get single category by UUID
- `addCategory(entity)` - Create category
- `updateCategory(entity, force)` - Update category
- `removeCategories([uuids], force)` - Delete categories
- `getDeviceDetails(name)` - Get device by name (with args parameter)
- Additional methods to discover: roles, power management, provisioning status

**CMPart Service:**
- `getSoftwareImages` - List all software images (data source)
- `getSoftwareImage(name)` - Get single image by name (direct lookup with args)
- `addSoftwareImage(entity, cloneFrom)` - Create/clone software image
- `updateSoftwareImage(entity)` - Update software image
- `removeSoftwareImages([names])` - Delete software images
- Additional methods to discover: modules, kernels, drivers

**CMNet Service:**
- `getNetworks` - List all networks (data source)
- Additional methods to discover: add/update/remove networks, VLANs, subnets

**Unknown Services to Explore:**
- CMProv - Provisioning workflows
- CMJob - Job management
- CMServ - Service configuration
- CMMon - Monitoring/metrics

#### Provider Implementation Analysis

```bash
# List all implemented resources
Read: internal/provider/provider.go
# Extract Resources() method, count bcm_* resources

# List all implemented data sources
Read: internal/provider/provider.go
# Extract DataSources() method, count bcm_* data sources
```

#### Gap Analysis Process

1. **Enumerate BCM API methods** - Create complete inventory per service
2. **Enumerate provider resources** - Extract from provider.go registration
3. **Cross-reference** - For each BCM method, check if provider has resource/data source
4. **Categorize gaps**:
   - **HIGH**: Core workflows (device management, provisioning, image management)
   - **MEDIUM**: Important features (monitoring, jobs, advanced networking)
   - **LOW**: Nice-to-have (GUI settings, minor utilities)

### Validation Approach

**Success Criteria:**
- All 7 BCM services enumerated with method lists
- Live API validation confirms methods work
- At least 10 high-value gaps identified
- Gaps categorized by business value

**Validation Steps:**
1. Test 3 sample API calls per service to confirm connectivity
2. Cross-reference discovered methods with sampleRest/ examples
3. Validate that identified gaps are truly missing from provider.go

---

## 3. Documentation Validation Strategy

### Objective

Execute scripts/test-examples.sh and perform root cause analysis on all failures.

### Research Question

How to execute scripts/test-examples.sh and capture detailed results? What constitutes comprehensive root cause analysis for failing examples? How to validate generated docs match source code examples?

### Decision: Automated Test Execution with Failure Forensics

Use test-examples.sh as primary validation tool, capture full output, and perform detailed failure analysis for each broken example.

### Rationale

- **Complete Coverage**: test-examples.sh discovers all examples in examples/ directory automatically
- **Real Validation**: Actually runs terraform init/validate/plan to catch real issues
- **Efficient**: Parallel execution for data sources (~10s for 10 examples)
- **Actionable**: Detailed error messages pinpoint exact fixes needed

### Alternatives Considered

1. **Manual terraform commands** - Rejected: Not repeatable, time-consuming
2. **Terraform test files** - Rejected: Not designed for example validation
3. **Static analysis only** - Rejected: Misses runtime issues like API compatibility

### Tools & Techniques

#### Test Execution Command

```bash
# Full test suite with verbose output
BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" \
BCM_PASSWORD="Hashicorp123!" \
SKIP_BUILD=true \
VERBOSE=true \
./scripts/test-examples.sh > test-examples-output.txt 2>&1

# Capture exit code
echo $? > test-examples-exit-code.txt
```

#### Output Analysis Patterns

**Pass/Fail Detection:**

```bash
# Count passing examples
Grep: '\[PASS\].*✓' -c

# Count failing examples
Grep: '\[FAIL\].*✗' -c

# Extract failed example names
Grep: '\[FAIL\].*✗\s+(.*)' output_mode=content
```

**Error Extraction:**

For each failing example, extract:
1. **Example name** - Directory path and file
2. **Failure phase** - terraform init, validate, plan, apply, destroy
3. **Error message** - First 10-20 lines of error output
4. **Context** - Which attribute or configuration caused failure

```bash
# Find failure phase
Grep: 'Failed at: (terraform init|terraform validate|terraform plan)'

# Extract error message
# Parse output between "[FAIL]" and next "[INFO]" or "[PASS]"
```

#### Root Cause Analysis Framework

For each failing example, determine:

1. **Configuration Issue** vs **Provider Bug**
   - Configuration: Missing required attribute, invalid value, syntax error
   - Provider Bug: Validation logic error, API client issue, schema mismatch

2. **Specific Fix Required**
   - Add missing attribute with example value
   - Fix attribute value to match validation
   - Update provider configuration
   - Fix provider schema/validation (file path + line number)

3. **Validation Approach**
   - Re-run test-examples.sh after fix
   - Verify terraform plan succeeds
   - Confirm example matches generated docs

#### Example Discovery

```bash
# Find all resource examples
Glob: examples/resources/**/resource.tf

# Find all data source examples
Glob: examples/data-sources/**/data-source.tf

# Cross-reference against registered resources
Read: internal/provider/provider.go
# Extract Resources() and DataSources() lists
# Compare against examples/ directory structure
# Identify missing examples
```

#### Documentation Sync Validation

```bash
# Check docs/ last modified time
ls -lt docs/ | head -20

# Check examples/ last modified time
ls -lt examples/resources/ examples/data-sources/ | head -20

# Check if make generate needed
git status docs/
# If docs/ shows uncommitted changes, needs regeneration
```

### Test-Examples.sh Implementation Details

Based on script analysis (scripts/test-examples.sh):

**Features:**
- **Phase 1**: Environment validation (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- **Phase 2**: Provider build (or skip with SKIP_BUILD=true)
- **Phase 3**: Example testing (parallel for data sources, sequential for resources)
- **Phase 4**: Cleanup (removes citest-* prefixed resources)

**Execution Strategy:**
- Data sources: Parallel (limit=4) - fast, read-only
- Resources: Sequential - creates real infrastructure, needs cleanup

**Test Phases per Example:**
1. `terraform init -backend=false`
2. `terraform validate`
3. `terraform plan`
4. `terraform apply` (for test-citest/ examples only)
5. `terraform destroy` (for test-citest/ examples only)

**Output Format:**
```
[INFO] [1/10] Testing bcm_cmdevice_categories/data-source.tf...
[PASS] ✓ bcm_cmdevice_categories/data-source.tf (2s)

[FAIL] ✗ bcm_cmpart_softwareimage/resource.tf (3s)
       Context: Phase 3 - Sequential test execution
       Failed at: terraform validate
       Error:
         Error: Insufficient kernel_parameters blocks
         on resource.tf line 12: Required attribute missing
```

### Validation Approach

**Success Criteria:**
- All examples tested (100% coverage of examples/ directory)
- Failing examples have root cause analysis with specific fix
- Missing examples identified (registered resources without examples)
- Documentation sync status determined

**Validation Steps:**
1. Execute test-examples.sh and capture full output
2. Parse output to extract pass/fail counts and example names
3. For each failure, perform manual inspection to confirm root cause
4. Cross-check missing examples list against provider.go registration

---

## 4. Code Consistency Analysis Patterns

### Objective

Analyze code patterns against HashiCorp best practices and identify deviations.

### Research Question

What HashiCorp best practices apply to this provider (Framework v1.16.1)? How to detect inconsistent error handling, schema patterns, and client usage? What terraform-provider-design skill queries validate compliance?

### Decision: Best Practice Queries + Pattern-Based Code Analysis

Use terraform-provider-design skill to query authoritative HashiCorp guidance, then apply Grep patterns to detect violations across all provider code.

### Rationale

- **Authoritative**: terraform-provider-design skill provides official HashiCorp best practices
- **Comprehensive**: Pattern analysis covers all resource/data source files
- **Objective**: Automated detection removes subjective judgment
- **Actionable**: Specific file paths and line numbers enable targeted fixes

### Alternatives Considered

1. **Manual code review** - Rejected: Time-consuming, subjective, inconsistent
2. **Static analysis tools (golangci-lint)** - Rejected: Doesn't check Terraform-specific patterns
3. **Provider scaffolding comparison** - Rejected: Scaffolds may not reflect latest best practices

### Tools & Techniques

#### HashiCorp Best Practice Queries

Use terraform-provider-design skill to query specific areas:

```plaintext
Query 1: Error Handling Best Practices
- How should provider errors be structured for terraform-plugin-framework v1.16.1?
- What diagnostic message format is recommended?
- How should API errors be wrapped and propagated?

Query 2: Schema Definition Best Practices
- What attributes should have MarkdownDescription?
- How should validators be structured?
- What's the pattern for optional vs required vs computed attributes?

Query 3: API Client Usage Best Practices
- How should HTTP clients be managed in providers?
- What's the pattern for authenticated session management?
- How should timeouts be handled?

Query 4: Async Operation Handling Best Practices
- What's the recommended retry/polling pattern?
- How should eventual consistency be handled?
- What timeout limits are appropriate?

Query 5: State Management Best Practices
- How should Unknown values be handled?
- What's the pattern for preserving plan values?
- How should null vs empty values be managed?
```

#### Code Pattern Analysis

**Error Handling Consistency:**

```bash
# Find parseErrorResponse() usage (should be universal)
Grep: 'parseErrorResponse'

# Find direct error handling (should use parseErrorResponse)
Grep: 'fmt\.Errorf.*API error'
Grep: 'return nil, err' -B 2  # Check if wrapped

# Find missing error diagnostics
Grep: 'resp\.Diagnostics\.AddError' -C 3
# Verify all have actionable messages
```

**Schema Description Completeness:**

```bash
# Find all schema attributes
Grep: 'Attributes:\s*map\[string\]schema\.Attribute'

# Find attributes without MarkdownDescription
Grep: 'schema\.StringAttribute\{' -A 10 | Grep: 'MarkdownDescription' -v
Grep: 'schema\.BoolAttribute\{' -A 10 | Grep: 'MarkdownDescription' -v
Grep: 'schema\.Int64Attribute\{' -A 10 | Grep: 'MarkdownDescription' -v
```

**BCM Client Usage Patterns:**

```bash
# Resource Read implementations (should use direct lookup)
Grep: 'func.*Resource.*Read' -A 50 | Grep: 'CallJSONRPC.*Args'

# Data source Read implementations (should use list methods)
Grep: 'func.*DataSource.*Read' -A 50 | Grep: 'CallJSONRPC'
```

**Async Operation Patterns:**

```bash
# Find time.Sleep usage (should use exponential backoff)
Grep: 'time\.Sleep'

# Find polling loops
Grep: 'for.*time\.Sleep' -C 5

# Verify exponential backoff pattern
Grep: 'waitTime.*\*=.*2'  # Doubling wait time
```

**State Management Issues:**

```bash
# Find Unknown value propagation
Grep: 'types\.StringUnknown'
Grep: '\.IsUnknown\(\).*true' -C 3

# Find plan value preservation patterns
Grep: 'state\..*=.*plan\.'
```

**Validator Consistency:**

```bash
# Find custom validators
Grep: 'stringvalidator\.'
Grep: 'int64validator\.'
Grep: 'boolvalidator\.'

# Identify inconsistent patterns
# E.g., some resources use URL validator, others don't
Grep: 'validators\.URLValidator' -l
```

#### Analysis Categories

Organize findings into categories (from plan.md):

1. **Error Handling** - parseErrorResponse usage, diagnostic messages
2. **Schema Patterns** - MarkdownDescription completeness, validator consistency
3. **Client Usage** - Direct lookup vs list+filter, args parameter usage
4. **Async Handling** - Polling patterns, retry logic, eventual consistency
5. **State Management** - Unknown values, plan preservation, null handling

### Validation Approach

**Success Criteria:**
- terraform-provider-design skill queried for 5 best practice areas
- At least 4 consistency categories analyzed
- All deviations flagged regardless of functionality
- Issues categorized by severity (Critical/Medium/Low)

**Validation Steps:**
1. Query terraform-provider-design skill and validate responses are authoritative
2. Run Grep patterns and manually verify 3 examples per category
3. Cross-check identified issues against working provider code (ensure patterns are accurate)

---

## 5. Remediation Planning Framework

### Objective

Synthesize all analysis findings into prioritized phases with measurable success criteria.

### Research Question

How to prioritize issues by production-readiness impact? What constitutes measurable success criteria for each phase? How to structure regression testing requirements?

### Decision: Impact-Based Prioritization with Quantifiable Metrics

Use 4-tier prioritization (Critical/High/Medium/Low) based on production-readiness impact, then group into phases with specific regression testing requirements.

### Rationale

- **Risk-Based**: Critical blockers addressed first (production deployment gates)
- **Measurable**: Each phase has quantifiable success criteria (e.g., "100% Import test coverage")
- **Validated**: Regression testing requirements ensure fixes don't break existing functionality
- **Actionable**: Clear phase dependencies and effort estimates enable roadmap planning

### Alternatives Considered

1. **Single priority list** - Rejected: No clear stopping points, difficult to plan sprints
2. **Fixed-scope phases** - Rejected: Doesn't adapt to actual findings
3. **Time-boxed phases** - Rejected: Quality over schedule per clarifications

### Tools & Techniques

#### Issue Aggregation

Combine findings from all 4 analysis reports:

```plaintext
Input Sources:
1. test-coverage-report.md - Missing tests, pattern migrations
2. api-gap-analysis.md - Missing resources/data sources
3. documentation-review.md - Failing examples, missing docs
4. code-consistency-report.md - Best practice violations
```

#### Prioritization Criteria

**Critical (Phase 0 - Production Blockers):**
- Missing Import tests (resource can't be imported into Terraform state)
- Missing Drift tests (Terraform can't detect external changes)
- Critical error handling gaps (silent failures, data loss risk)
- Failing examples that prevent documentation generation

**High (Phase 1 - Production Readiness):**
- Missing Idempotency tests (repeated applies cause changes)
- Missing high-value resources (top 3 from API gap analysis)
- All schema attributes missing descriptions (poor UX)
- Legacy test patterns (technical debt, harder maintenance)

**Medium (Phase 2 - Quality & Consistency):**
- BCM client usage inconsistencies (performance impact)
- Missing advanced examples (documentation completeness)
- Medium-value API gaps (nice-to-have features)
- Validator standardization

**Low (Phase 3 - Polish & Future):**
- Low-value API gaps
- Edge case test coverage
- Performance optimizations
- Enhanced diagnostics

#### Success Criteria Definition

Each phase must have:

1. **Quantifiable Metrics**
   - Example: "100% of resources have Import + Drift tests"
   - Example: "Zero usage of legacy TestCheckResourceAttr patterns"
   - Example: "All examples pass test-examples.sh validation"

2. **Verification Commands**
   - Example: `TF_ACC=1 go test -v ./internal/provider/`
   - Example: `./scripts/test-examples.sh`
   - Example: `make generate && git diff docs/`

3. **Regression Testing Requirements**
   - Full acceptance test suite must pass
   - Example validation must succeed
   - Documentation generation must complete without errors

#### Phase Structure Template

```markdown
## Phase N: [Name] ([Priority])

**Deliverables**:
1. [Specific fix with file paths]
2. [...]

**Success Criteria**:
- [Quantifiable metric 1]
- [Quantifiable metric 2]
- [...]

**Regression Testing**:
- [Test command 1]
- [Test command 2]
- [...]

**Estimated Effort**: [Days/weeks with justification]

**Dependencies**: [Other phases or prerequisites]
```

#### Effort Estimation

Use T-shirt sizing with reference points:

- **Small (1-2 days)**: Add missing test to single resource
- **Medium (3-5 days)**: Migrate all tests to modern patterns
- **Large (5-7 days)**: Implement 3 new high-value resources
- **X-Large (1-2 weeks)**: Complete API gap implementation

Reference: plan.md estimates 15-22 days total for full remediation

### Validation Approach

**Success Criteria:**
- All findings grouped into 3-5 distinct phases
- Each phase has quantifiable success criteria
- Effort estimates align with plan.md timeline (15-22 days total)
- Phase dependencies clearly specified

**Validation Steps:**
1. Verify all issues from 4 reports are included in remediation plan
2. Cross-check phase groupings with production-readiness impact
3. Validate that Phase 0 addresses all Critical issues
4. Confirm regression testing requirements are comprehensive

---

## Research Summary

### Key Decisions

1. **Test Coverage**: Pattern-based Grep analysis with CRUD/Import/Drift/Idempotency detection
2. **API Gap**: Multi-source discovery (live BCM API + sampleRest/ + provider code)
3. **Documentation**: Automated test-examples.sh execution with failure forensics
4. **Code Consistency**: terraform-provider-design skill queries + pattern analysis
5. **Remediation**: Impact-based 4-tier prioritization with quantifiable success criteria

### Methodologies Validated

All 5 research questions have concrete answers with tools, commands, and validation approaches documented. The methodologies are:
- ✅ Repeatable (commands/patterns can be re-run)
- ✅ Comprehensive (covers all resources and analysis areas)
- ✅ Objective (automated pattern detection)
- ✅ Actionable (produces specific file paths and line numbers)

### Ready for Phase 1

With these methodologies documented, the analysis phase (Phase 2-5) can proceed in parallel with clear approaches for each user story.

### Cross-References

- **Test Coverage Patterns**: See CLAUDE.md "Modern Testing Patterns" section for knownvalue matchers
- **BCM API Methods**: See sampleRest/CMDevice_Complete_Documentation.md for known methods
- **Test Infrastructure**: See scripts/test-examples.sh for validation workflow
- **Best Practices**: Query terraform-provider-design skill during Phase 5 analysis

---

## Appendices

### A. File Paths Summary

**Analysis Targets:**
- Resources: `/workspace/internal/provider/resource_*.go`
- Resource Tests: `/workspace/internal/provider/resource_*_test.go`
- Data Sources: `/workspace/internal/provider/data_source_*.go`
- Data Source Tests: `/workspace/internal/provider/data_source_*_test.go`
- Provider Registration: `/workspace/internal/provider/provider.go`
- Test Helpers: `/workspace/internal/provider/test_helpers.go`
- BCM Client: `/workspace/internal/provider/bcm_client.go`

**Validation Targets:**
- Examples: `/workspace/examples/{resources,data-sources}/*/`
- Documentation: `/workspace/docs/`
- Test Script: `/workspace/scripts/test-examples.sh`

**Output Artifacts:**
- Research: `/workspace/specs/001-production-review/research.md` (this file)
- Test Coverage Report: `/workspace/specs/001-production-review/test-coverage-report.md`
- API Gap Analysis: `/workspace/specs/001-production-review/api-gap-analysis.md`
- Documentation Review: `/workspace/specs/001-production-review/documentation-review.md`
- Code Consistency Report: `/workspace/specs/001-production-review/code-consistency-report.md`
- Remediation Plan: `/workspace/specs/001-production-review/remediation-plan.md`
- Quickstart Guide: `/workspace/specs/001-production-review/quickstart.md`

### B. Environment Variables

**Required for BCM API Discovery:**
```bash
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

**Required for Acceptance Tests:**
```bash
export TF_ACC=1
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

**Optional for Test Script:**
```bash
export SKIP_BUILD=true  # Skip provider rebuild
export VERBOSE=true     # Detailed output
export PARALLEL_LIMIT=4 # Data source parallelism
```

### C. Timeline Estimates

Based on tasks.md breakdown:

- **Phase 1: Research** - Complete (this document)
- **Phase 2: Test Coverage Analysis** - 3-4 hours
- **Phase 3: API Gap Analysis** - 2-3 hours (requires live BCM access)
- **Phase 4: Documentation Validation** - 1-2 hours
- **Phase 5: Code Consistency Review** - 3-4 hours
- **Phase 6: Remediation Planning** - 2-3 hours (synthesis)
- **Phase 7: Polish & Deliverables** - 1 hour

**Total Estimated Time**: 12-17 hours for complete analysis (aligns with plan.md estimate of 15-20 hours)

---

**Status**: Research complete - Ready to proceed with Phase 2-5 analysis workflows
