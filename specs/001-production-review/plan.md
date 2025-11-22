# Implementation Plan: Production-Ready Codebase Review

**Branch**: `001-production-review` | **Date**: 2025-11-22 | **Spec**: [spec.md](./spec.md)
**Input**: Feature specification from `/workspace/specs/001-production-review/spec.md`

## Summary

This feature implements a comprehensive production-readiness review of the Terraform Provider BCM codebase. The review systematically analyzes test coverage (CRUD, Import, Drift, Idempotency), BCM API coverage gaps, documentation and examples validation, code consistency against HashiCorp best practices, and produces a phased remediation plan. The goal is to identify all gaps preventing production deployment and create an actionable roadmap to address them. This is an analysis and planning feature that does not modify provider code - it produces diagnostic reports and remediation plans.

## Technical Context

**Language/Version**: Go 1.24.0
**Primary Dependencies**: terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3, BCM JSON-RPC API client
**Storage**: File-based reports in specs/001-production-review/ directory
**Testing**: Analysis scripts executed via Bash, validation via scripts/test-examples.sh
**Target Platform**: Linux development environment (per env context)
**Project Type**: Analysis and reporting tooling (generates markdown reports, not production code)
**Performance Goals**: Complete analysis in reasonable time (no strict limit per clarifications - thoroughness over speed)
**Constraints**: Must access live BCM cluster at https://172.21.15.254:8081 for API discovery, requires TF_ACC=1 environment for test execution analysis
**Scale/Scope**: Analyze ~10 resources, ~4 data sources, ~30 test files, all examples in examples/ directory

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

**TDD Compliance**: ✅ PASS - This is an analysis feature that examines TDD compliance of existing code. No new production code is being added that would require tests. The analysis scripts themselves will be verified through execution and output validation.

**Simplicity**: ✅ PASS - Analysis approach uses straightforward file parsing, grep/glob searches, and script execution. No complex abstractions or patterns needed for report generation.

**Test-First Development**: ✅ PASS - Not applicable for analysis tooling. The deliverables are reports (research.md, test-coverage-report.md, api-gap-analysis.md, etc.), not testable code.

**Parallel Execution**: ✅ PASS - Analysis tasks can be parallelized where appropriate (e.g., analyzing multiple resource files concurrently, running multiple grep searches simultaneously).

**Constitution Violations**: None identified. This is a meta-analysis feature that does not add complexity to the provider itself.

## Project Structure

### Documentation (this feature)

```text
specs/001-production-review/
├── plan.md                      # This file (/speckit.plan command output)
├── research.md                  # Phase 0 output - Analysis methodology research
├── test-coverage-report.md      # Phase 1 output - Comprehensive test coverage analysis
├── api-gap-analysis.md          # Phase 1 output - BCM API coverage gaps
├── documentation-review.md      # Phase 1 output - Examples and docs validation results
├── code-consistency-report.md   # Phase 1 output - Code patterns and consistency review
├── remediation-plan.md          # Phase 1 output - Phased improvement roadmap
└── quickstart.md                # Phase 1 output - How to use the analysis reports
```

### Source Code (repository root)

```text
internal/provider/
├── bcm_client.go                    # BCM JSON-RPC client (analysis target)
├── provider.go                      # Provider registration (analysis target)
├── resource_*.go                    # Resource implementations (analysis targets)
├── resource_*_test.go               # Resource tests (analysis targets)
├── data_source_*.go                 # Data source implementations (analysis targets)
├── data_source_*_test.go            # Data source tests (analysis targets)
└── test_helpers.go                  # Test infrastructure (analysis target)

examples/
├── provider/                        # Provider examples (validation targets)
├── resources/                       # Resource examples (validation targets)
└── data-sources/                    # Data source examples (validation targets)

scripts/
└── test-examples.sh                 # Example validation infrastructure (execution target)

docs/                                # Generated documentation (validation target)

specs/001-production-review/         # Analysis outputs (this feature)
```

**Structure Decision**: This is an analysis feature that examines the existing terraform-provider-bcm codebase structure. The feature does not add new source code to the provider - it produces analysis reports stored in the specs/001-production-review/ directory. All analysis targets exist in the current provider structure (internal/provider/, examples/, scripts/, docs/).

## Complexity Tracking

No constitution violations requiring justification.

## Phase 0: Research & Discovery

### Objective

Research and document the methodologies for comprehensive provider analysis, including automated analysis approaches, HashiCorp best practices discovery, and BCM API introspection techniques.

### Research Questions

1. **Test Coverage Analysis Methodology**
   - How to systematically identify missing CRUD, Import, Drift, and Idempotency tests?
   - What grep/glob patterns detect modern vs legacy testing patterns?
   - How to parse test files to identify specific test step patterns?

2. **BCM API Discovery Approach**
   - How to query BCM API to enumerate all available services and methods?
   - What JSON-RPC introspection calls reveal API capabilities?
   - How to compare discovered API methods against provider implementation?

3. **Documentation Validation Strategy**
   - How to execute scripts/test-examples.sh and capture detailed results?
   - What constitutes comprehensive root cause analysis for failing examples?
   - How to validate generated docs match source code examples?

4. **Code Consistency Analysis Patterns**
   - What HashiCorp best practices apply to this provider (Framework v1.16.1)?
   - How to detect inconsistent error handling, schema patterns, and client usage?
   - What terraform-provider-design skill queries validate compliance?

5. **Remediation Planning Framework**
   - How to prioritize issues by production-readiness impact?
   - What constitutes measurable success criteria for each phase?
   - How to structure regression testing requirements?

### Research Deliverable

**Output**: `/workspace/specs/001-production-review/research.md`

**Content Structure**:
- Decision: [Chosen analysis approach for each research question]
- Rationale: [Why this approach is optimal for production-readiness review]
- Alternatives Considered: [Other approaches evaluated and rejected]
- Tools & Techniques: [Specific commands, scripts, API calls to use]
- Validation Approach: [How to verify analysis accuracy]

### Research Tasks

1. Query BCM API documentation and test live endpoint for service enumeration patterns
2. Review terraform-plugin-testing v1.13.3 docs for modern testing pattern requirements
3. Examine existing test files to identify coverage analysis patterns (grep for TestStep, ImportState, PreConfig, plancheck, statecheck)
4. Research HashiCorp Provider Design Principles via terraform-provider-design skill
5. Analyze scripts/test-examples.sh to understand execution and output parsing
6. Document remediation planning frameworks from TDD and software quality literature

## Phase 1: Design & Implementation Planning

### Objective

Execute the analysis workflows defined in Phase 0 research and generate comprehensive diagnostic reports for each review area.

### Design Artifacts

#### 1. Test Coverage Analysis Report

**Output**: `/workspace/specs/001-production-review/test-coverage-report.md`

**Analysis Approach** (from research.md):
- Use Glob to find all `resource_*.go` and `resource_*_test.go` files
- Use Grep to search for specific test patterns in each test file:
  - CRUD coverage: Search for Create, Update, Delete test steps with Config checks
  - Import coverage: Search for `ImportState: true` patterns
  - Drift coverage: Search for `PreConfig` with external modification + `plancheck.ExpectNonEmptyPlan()`
  - Idempotency coverage: Search for `plancheck.ExpectEmptyPlan()` after Create/Update
  - Modern patterns: Search for `statecheck.ExpectKnownValue`, `knownvalue.StringExact`, `compareValue`
  - Legacy patterns: Count `resource.TestCheckResourceAttr` usage
  - Environment portability: Search for hardcoded counts, names, UUIDs in test configs

**Report Structure**:
```markdown
# Test Coverage Analysis Report

## Summary
- Total Resources: X
- Resources with Complete Coverage: X (XX%)
- Resources Missing Coverage: X (XX%)
- Total Data Sources: X
- Data Sources with Complete Coverage: X (XX%)

## Resource-by-Resource Analysis

### bcm_cmdevice_category
- CRUD Coverage: ✅ Create, ✅ Read, ✅ Update, ✅ Delete
- Import Test: ✅ Present (line 123)
- Drift Test: ✅ Present (TestAccCMDeviceCategory_DriftNotes, line 234)
- Idempotency Test: ❌ MISSING - No plancheck.ExpectEmptyPlan after Create/Update
- Modern Patterns: ⚠️ PARTIAL - Uses statecheck for some checks, legacy TestCheckResourceAttr for others
- Environment Portability: ✅ PASS - Uses generateUniqueTestName()
- **Gap Summary**: Missing idempotency verification, mixed testing patterns

[... repeat for each resource ...]

## Data Source Analysis

[Similar structure for data sources]

## Testing Pattern Modernization Gaps

- Legacy TestCheckResourceAttr usage: X instances across Y files
- Files needing modern pattern migration: [list with line numbers]

## Recommended Actions

1. [Specific file + test + gap to address]
2. [...]
```

#### 2. BCM API Gap Analysis

**Output**: `/workspace/specs/001-production-review/api-gap-analysis.md`

**Analysis Approach** (from research.md):
- Query live BCM API to discover all services (CMDevice, CMPart, CMNet, CMProv, CMJob, CMServ, CMMon, etc.)
- For each service, enumerate methods by pattern matching or introspection
- Parse provider.go Resources() and DataSources() methods to list implemented resources
- Cross-reference BCM API capabilities against provider implementation
- Categorize gaps by service and business value

**Report Structure**:
```markdown
# BCM API Gap Analysis

## Summary
- Total BCM Services Discovered: X
- Total BCM Methods: X
- Implemented Resources: X
- Implemented Data Sources: X
- High-Value Gaps: X
- Medium-Value Gaps: X
- Low-Value Gaps: X

## Service-by-Service Analysis

### CMDevice Service
**Methods Discovered**: getNodes, getCategories, addDevice, updateDevice, removeDevice, getDeviceDetails, ...

**Implementation Status**:
- ✅ bcm_cmdevice_device (resource) - uses addDevice, updateDevice, removeDevice, getDevice
- ✅ bcm_cmdevice_category (resource) - uses addCategory, updateCategory, removeCategory, getCategories
- ✅ bcm_cmdevice_nodes (data source) - uses getNodes
- ✅ bcm_cmdevice_categories (data source) - uses getCategories
- ❌ Missing: Device roles management (addRole, updateRole, removeRole)
- ❌ Missing: Device power management (powerOn, powerOff, reboot)

**Gap Priority**: HIGH - Device power management is core user workflow

[... repeat for each service ...]

## High-Value Missing Resources

1. **CMProv Provisioning Jobs** (service: cmprov, methods: getProvisioningStatus, startProvisioning, stopProvisioning)
   - Business Value: Users need to trigger and monitor provisioning workflows
   - Implementation Effort: Medium (async operation handling required)
   - API Coverage: Fully documented in live API

[... list with priority ranking ...]

## Recommended Implementation Order

1. Phase 1 (Critical): [High-value resource gaps]
2. Phase 2 (Important): [Medium-value resource gaps]
3. Phase 3 (Nice-to-have): [Low-value resource gaps]
```

#### 3. Documentation & Examples Validation

**Output**: `/workspace/specs/001-production-review/documentation-review.md`

**Analysis Approach** (from research.md):
- Execute `scripts/test-examples.sh` with full output capture
- For each example, record init/validate/plan status
- For failures, perform root cause analysis:
  - Parse error message for exact failure point
  - Determine if configuration issue (syntax, missing attributes) or provider bug
  - Document specific fix steps required
  - Define validation approach to confirm fix
- Cross-reference examples/ directory structure against provider.go registered resources

**Report Structure**:
```markdown
# Documentation & Examples Validation Report

## Summary
- Total Resource Examples: X
- Passing Examples: X (XX%)
- Failing Examples: X (XX%)
- Total Data Source Examples: X
- Passing Data Source Examples: X (XX%)
- Missing Examples: X

## Resource Examples Validation

### bcm_cmdevice_category
- Example Path: `examples/resources/bcm_cmdevice_category/resource.tf`
- terraform init: ✅ PASS
- terraform validate: ✅ PASS
- terraform plan: ✅ PASS
- Uses unique naming: ✅ Yes (citest- prefix)
- Uses env vars for auth: ✅ Yes

### bcm_cmpart_softwareimage
- Example Path: `examples/resources/bcm_cmpart_softwareimage/resource.tf`
- terraform init: ✅ PASS
- terraform validate: ❌ FAIL
- terraform plan: N/A (validate failed)
- **Failure Root Cause Analysis**:
  - Exact Error: "Error: Insufficient kernel_parameters blocks\n  on resource.tf line 12: Required attribute missing"
  - Underlying Cause: Example config missing required kernel_parameters attribute introduced in provider v0.1.0
  - Specific Fix: Add `kernel_parameters = "console=tty0 console=ttyS0,115200n8"` to example
  - Validation Approach: Re-run test-examples.sh after fix, verify plan succeeds
- Uses unique naming: ✅ Yes
- Uses env vars for auth: ✅ Yes

[... repeat for each example ...]

## Missing Examples

- bcm_cmdevice_device: No advanced example showing all optional attributes
- bcm_cmnet_networks: Only basic example, no filtering example

## Generated Documentation Sync

- docs/ directory last updated: [timestamp from git]
- Examples last modified: [timestamp from git]
- Sync Status: ✅ Current / ❌ Outdated (needs `make generate`)

## Recommended Actions

1. Fix failing example: bcm_cmpart_softwareimage/resource.tf (add kernel_parameters)
2. Add advanced examples: bcm_cmdevice_device (all attributes)
3. Run `make generate` to sync docs with examples
4. Re-run test-examples.sh to verify all fixes
```

#### 4. Code Consistency Analysis

**Output**: `/workspace/specs/001-production-review/code-consistency-report.md`

**Analysis Approach** (from research.md):
- Use terraform-provider-design skill to query HashiCorp best practices
- Analyze all resource/data source files for pattern consistency:
  - Error handling: Search for parseErrorResponse() usage
  - Schema patterns: Check all attributes have MarkdownDescription
  - BCM client usage: Verify resources use direct lookup (args), data sources use list+filter
  - Async handling: Identify async operations, verify exponential backoff polling
  - State management: Search for Unknown value propagation, plan value preservation
  - Validators: Check consistent validation approaches

**Report Structure**:
```markdown
# Code Consistency Analysis Report

## Summary
- Files Analyzed: X resources, Y data sources
- HashiCorp Best Practices Violations: X
- Consistency Issues Identified: X
- Critical Issues: X
- Medium Issues: X
- Low Issues: X

## Error Handling Consistency

**Standard Pattern** (from bcm_client.go parseErrorResponse()):
```go
func parseErrorResponse(body []byte) error {
    // Multi-layer error detection
    var errorResp ErrorResponse
    if err := json.Unmarshal(body, &errorResp); err == nil {
        if errorResp.Error != "" {
            return fmt.Errorf("BCM API error: %s", errorResp.Error)
        }
    }
    // ... additional layers
}
```

**Violations**:
- resource_cmdevice_device.go:234 - Direct error string comparison instead of parseErrorResponse()
- data_source_cmnet_networks.go:123 - No error handling for CallJSONRPC failure
- **HashiCorp Best Practice**: All API errors must be parsed consistently and return actionable diagnostic messages

[... similar sections for each consistency category ...]

## Schema Description Completeness

**Missing Descriptions**:
- resource_cmdevice_device.go:78 - Attribute "enable_sol" has no MarkdownDescription
- resource_cmpart_softwareimage.go:92 - Attribute "clone_from" has no MarkdownDescription
- **HashiCorp Best Practice**: All schema attributes must have clear, user-facing descriptions

## BCM Client Usage Patterns

**Resources (should use direct lookup with args)**:
- ✅ resource_cmpart_softwareimage.go - Uses CallJSONRPC("cmpart", "getSoftwareImage", name)
- ❌ resource_cmdevice_category.go - Uses CallJSONRPC("cmdevice", "getCategories") then filters (inefficient)
- **HashiCorp Best Practice**: Resource Read operations should use efficient direct lookups when API supports args parameter

**Data Sources (should use list methods)**:
- ✅ data_source_cmpart_softwareimages.go - Uses CallJSONRPC("cmpart", "getSoftwareImages") then client-side filter
- ✅ data_source_cmdevice_nodes.go - Uses CallJSONRPC("cmdevice", "getNodes") then client-side filter

[... continue for all consistency categories ...]

## Recommended Remediation

### Critical Issues (Must Fix for Production)
1. Add error handling to data_source_cmnet_networks.go:123 using parseErrorResponse()
2. Fix Unknown value propagation in resource_cmpart_softwareimage.go:345

### Medium Issues (Should Fix for Consistency)
1. Add schema descriptions for all missing MarkdownDescription fields
2. Migrate resource_cmdevice_category.go to use direct lookup pattern

### Low Issues (Nice-to-Have Improvements)
1. Standardize validator usage across resources
2. Add exponential backoff to resource_cmdevice_device.go async polling
```

#### 5. Phased Remediation Plan

**Output**: `/workspace/specs/001-production-review/remediation-plan.md`

**Structure**:
```markdown
# Production Readiness Remediation Plan

## Overview

This plan organizes all identified issues from test coverage, API gap, documentation, and code consistency analyses into prioritized phases with clear success criteria and regression testing requirements.

## Phase 0: Critical Blockers (Required for Production)

**Deliverables**:
1. Fix all missing Import tests (resource_cmdevice_device, resource_cmpart_softwareimage)
2. Fix all missing Drift tests (resource_cmdevice_device, resource_cmpart_softwareimage)
3. Fix critical error handling gaps (data_source_cmnet_networks.go, resource_cmdevice_device.go)
4. Fix failing examples that prevent documentation generation

**Success Criteria**:
- 100% of resources have Import + Drift tests
- Zero critical error handling violations
- scripts/test-examples.sh passes with 100% success rate
- All acceptance tests pass: `TF_ACC=1 go test -v ./internal/provider/`

**Regression Testing**:
- Run full acceptance test suite: `TF_ACC=1 go test -v -timeout 120m ./internal/provider/`
- Run example validation: `scripts/test-examples.sh`
- Validate generated docs: `make generate && git diff docs/`

**Estimated Effort**: 2-3 days

## Phase 1: High-Priority Gaps (Production Readiness)

**Deliverables**:
1. Add missing Idempotency tests for all resources
2. Migrate all tests to modern patterns (statecheck, plancheck, compareValue)
3. Add missing high-value resources from API gap analysis (top 3 resources)
4. Complete all missing schema descriptions

**Success Criteria**:
- 100% of resources have complete CRUD + Import + Drift + Idempotency coverage
- Zero usage of legacy TestCheckResourceAttr patterns
- Top 3 high-value API gaps implemented with full test coverage
- All schema attributes have MarkdownDescription

**Regression Testing**:
- Full acceptance test suite passes
- Example validation passes
- Documentation generation succeeds
- Live BCM cluster validation (manual testing of new resources)

**Estimated Effort**: 5-7 days

## Phase 2: Medium-Priority Improvements (Quality & Consistency)

**Deliverables**:
1. Fix all BCM client usage pattern inconsistencies
2. Add advanced examples for all resources
3. Implement medium-value API gaps from gap analysis
4. Standardize validator usage across resources

**Success Criteria**:
- All resources use efficient direct lookup patterns
- All resources have basic + advanced examples
- Medium-value API gaps implemented
- Consistent validation patterns across all resources

**Regression Testing**:
- Full acceptance test suite passes
- Example validation passes
- Code consistency analysis shows zero violations

**Estimated Effort**: 3-5 days

## Phase 3: Polish & Nice-to-Have (Future Enhancements)

**Deliverables**:
1. Implement low-value API gaps
2. Add comprehensive edge case tests
3. Performance optimization for test suite
4. Enhanced error messages and diagnostics

**Success Criteria**:
- 90%+ of BCM API methods exposed through provider
- Edge case test coverage for all resources
- Acceptance test suite execution time < 90 minutes
- All error messages provide actionable guidance

**Regression Testing**:
- Full acceptance test suite passes
- Performance benchmarks meet targets
- User testing feedback incorporated

**Estimated Effort**: 5-7 days

## Total Timeline Estimate

- Phase 0 (Critical): 2-3 days
- Phase 1 (High Priority): 5-7 days
- Phase 2 (Medium Priority): 3-5 days
- Phase 3 (Polish): 5-7 days
- **Total**: 15-22 days (3-4 weeks)

## Dependencies

- Phases must be completed sequentially (Phase 0 gates Phase 1, etc.)
- BCM cluster access required for all phases
- Acceptance test environment (TF_ACC=1, credentials) required for all phases
- terraform-provider-design skill access for HashiCorp best practices validation
```

#### 6. Developer Quickstart Guide

**Output**: `/workspace/specs/001-production-review/quickstart.md`

**Content**:
```markdown
# Production Review Analysis - Quickstart Guide

## Overview

This directory contains comprehensive production-readiness analysis reports for the Terraform Provider BCM. The analysis identifies all gaps preventing production deployment and provides a phased remediation plan.

## Generated Reports

1. **test-coverage-report.md** - Complete CRUD, Import, Drift, Idempotency test coverage analysis
2. **api-gap-analysis.md** - BCM API methods vs provider implementation coverage
3. **documentation-review.md** - Examples validation and documentation sync status
4. **code-consistency-report.md** - HashiCorp best practices compliance analysis
5. **remediation-plan.md** - Prioritized phases with success criteria and effort estimates

## How to Use These Reports

### For Maintainers

1. **Start with remediation-plan.md** to understand the overall roadmap
2. **Use test-coverage-report.md** to identify which tests need to be added
3. **Reference code-consistency-report.md** for specific file/line violations to fix
4. **Check documentation-review.md** for failing examples requiring fixes

### For Contributors

1. **Review api-gap-analysis.md** to find high-value features to implement
2. **Follow remediation-plan.md** phases to align contributions with priorities
3. **Use code-consistency-report.md** to understand expected patterns
4. **Validate changes against test-coverage-report.md** requirements

## Re-Running the Analysis

To regenerate these reports after code changes:

```bash
# Execute the analysis workflow
cd /workspace/specs/001-production-review

# Run test coverage analysis
# [Commands from research.md]

# Run API gap analysis
# [Commands from research.md]

# Run documentation validation
scripts/test-examples.sh

# Run code consistency analysis
# [Commands from research.md]
```

## Report Interpretation

### Test Coverage Report
- ✅ = Complete coverage
- ⚠️ = Partial coverage (needs improvement)
- ❌ = Missing coverage (required for production)

### API Gap Analysis
- **HIGH** = Critical user workflows (implement first)
- **MEDIUM** = Important features (implement second)
- **LOW** = Nice-to-have (implement later)

### Code Consistency Report
- **Critical** = Blocks production (fix immediately)
- **Medium** = Impacts maintainability (fix in Phase 1-2)
- **Low** = Minor improvements (fix opportunistically)

## Success Metrics

Production readiness achieved when:
- Test coverage report shows 100% CRUD + Import + Drift + Idempotency
- All examples in documentation-review.md pass validation
- Code consistency report shows zero Critical violations
- Remediation plan Phase 0 and Phase 1 complete

## Questions?

See research.md for detailed analysis methodologies and validation approaches.
```

### Implementation Strategy

**Execution Order**:
1. Research Phase (research.md) - Document analysis methodologies
2. Test Coverage Analysis (parallel with API Gap Analysis)
3. API Gap Analysis (parallel with Test Coverage Analysis)
4. Documentation Validation (parallel with Code Consistency Analysis)
5. Code Consistency Analysis (parallel with Documentation Validation)
6. Remediation Planning (synthesizes all prior analyses)
7. Quickstart Guide (final documentation)

**Validation Approach**:
- Each report must include specific file paths, line numbers, and actionable recommendations
- Cross-reference findings across reports for consistency
- Validate all recommendations against HashiCorp best practices via terraform-provider-design skill
- Test all example fixes by actually running scripts/test-examples.sh

## Phase 2: Task Generation

Phase 2 (task generation via `/speckit.tasks`) is NOT part of this command. After plan.md is complete, the user will run `/speckit.tasks` separately to generate tasks.md from this plan.

## Open Questions from Spec

The spec identified several edge cases to consider during implementation:

1. **BCM API unexpected structures** - Research phase will document error handling strategies
2. **Test suite BCM cluster variations** - Documentation review will identify environment portability issues
3. **Examples referencing non-existent resources** - Documentation review will flag these as failures
4. **BCM API version differences** - API gap analysis will note any version-specific methods
5. **Drift test concurrency** - Code consistency analysis will review test isolation patterns
6. **BCM API rate limiting** - Research phase will document timeout handling approaches
7. **Deep root cause analysis time** - Acknowledged per clarifications (thoroughness over speed)
8. **HashiCorp best practices volume** - terraform-provider-design skill queries will focus on high-impact areas
9. **BCM API method deprecation** - API gap analysis will flag deprecated methods discovered

## Dependencies

- Live BCM cluster access (https://172.21.15.254:8081) for API discovery
- Acceptance test environment (TF_ACC=1, BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)
- terraform-provider-design skill for HashiCorp best practices queries
- scripts/test-examples.sh for documentation validation
- Go 1.24.0 and all provider dependencies (terraform-plugin-framework v1.16.1, terraform-plugin-testing v1.13.3)

## Success Criteria

This plan is complete when:
- ✅ All Phase 0 research questions documented in research.md
- ✅ All Phase 1 analysis reports generated with specific findings
- ✅ Remediation plan structures all issues into prioritized phases
- ✅ Quickstart guide provides clear navigation of reports
- ✅ All reports include actionable file paths, line numbers, and fix steps
- ✅ Cross-report consistency validated (e.g., remediation plan references specific findings from other reports)
