# Feature Specification: Production-Ready Codebase Review

**Feature Branch**: `001-production-review`
**Created**: 2025-11-22
**Status**: Draft
**Input**: User description: "Create a comprehensive specification for a production-ready review of the Terraform Provider BCM codebase covering test coverage analysis, BCM API gap analysis, documentation and examples review, code consistency analysis, and phased remediation plan"

## Clarifications

### Session 2025-11-22

- Q: What minimum test coverage percentage should be considered acceptable for production readiness? → A: 100% CRUD + Import + Drift for all resources
- Q: For BCM API gap analysis, how should we prioritize undocumented API methods discovered at runtime that aren't in sampleRest/ documentation? → A: BCM API is the source of truth - query live endpoints for authoritative capabilities
- Q: When examples validation fails, what level of detail is required in the remediation plan for each failing example? → A: Root cause analysis + specific fix steps
- Q: For the code consistency analysis, should the review enforce HashiCorp best practices even if current working code uses different patterns? → A: Strict HashiCorp compliance required
- Q: What is the maximum acceptable execution time for the complete production review analysis (all 5 user stories)? → A: No strict limit - thoroughness over speed

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Test Coverage Audit (Priority: P1)

As a provider maintainer, I need comprehensive test coverage analysis to ensure all resources and data sources have complete acceptance test coverage (CRUD, Import, Drift, Idempotency) so that the provider is reliable and maintainable in production.

**Why this priority**: Test coverage is foundational to provider quality. Without comprehensive tests, we cannot guarantee the provider works correctly, prevents regressions, or handles drift detection properly. This directly impacts user trust and production reliability.

**Production Readiness Standard**: 100% of resources must have complete CRUD + Import + Drift + Idempotency test coverage. Zero tolerance for gaps in production-ready provider.

**Independent Test**: Can be fully tested by running the test coverage analysis against each resource/data source implementation file and comparing results against modern testing pattern requirements. Delivers a detailed test coverage report with specific gaps identified.

**Acceptance Scenarios**:

1. **Given** all resource implementations in `internal/provider/resource_*.go`, **When** I analyze their corresponding test files, **Then** each resource must have 100% coverage for Create, Read, Update, Delete, Import, Drift detection, and Idempotency verification with no exceptions
2. **Given** all data source implementations in `internal/provider/data_source_*.go`, **When** I analyze their corresponding test files, **Then** each data source must have tests for basic retrieval and filter functionality
3. **Given** existing test files, **When** I review testing patterns, **Then** all tests must use modern patterns (statecheck.ExpectKnownValue, plancheck, compareValue) instead of legacy string-based assertions
4. **Given** acceptance tests, **When** I review environment portability, **Then** tests must not contain hardcoded resource counts, names, or UUIDs that assume specific BCM cluster state
5. **Given** CheckDestroy functions, **When** I review implementation, **Then** they must use enhanced patterns with detailed error messages and exponential backoff verification

---

### User Story 2 - BCM API Gap Analysis (Priority: P1)

As a provider user, I need visibility into which BCM API capabilities are exposed through the provider and which are missing so that I can understand the provider's current scope and plan for future feature requests.

**Why this priority**: Understanding API coverage is critical for production readiness. Users need to know what BCM operations are supported and what gaps exist. This directly impacts whether the provider can meet user requirements.

**API Discovery Approach**: Live BCM API is the authoritative source of truth. Query BCM endpoints directly to discover all available services and methods. The `sampleRest/` documentation serves as reference material but the live API determines actual capabilities.

**Independent Test**: Can be fully tested by querying the BCM API to enumerate all available services/methods, comparing against implemented resources/data sources, and producing a gap analysis report. Delivers a concrete list of missing high-value resources.

**Acceptance Scenarios**:

1. **Given** live BCM API access, **When** I enumerate all available services by querying the API directly, **Then** I must have a complete list of API services with their methods from the authoritative source
2. **Given** provider implementation in `internal/provider/provider.go`, **When** I list all registered resources and data sources, **Then** I can identify which BCM services are currently implemented
3. **Given** BCM API capabilities from live queries and provider implementation, **When** I perform gap analysis, **Then** I must produce a prioritized list of missing resources/data sources with business value assessment based on actual API capabilities
4. **Given** identified gaps, **When** I categorize missing functionality, **Then** gaps must be grouped by service (Device, Network, Software, Provisioning, Job, Monitoring) with priority levels
5. **Given** `sampleRest/` documentation, **When** I cross-reference with live API results, **Then** documentation is used as supplementary reference material but live API is the authoritative source

---

### User Story 3 - Documentation & Examples Validation (Priority: P2)

As a provider user, I need comprehensive, working examples for all resources and data sources so that I can quickly understand how to use the provider and integrate it into my Terraform configurations.

**Why this priority**: Documentation quality directly impacts user adoption and satisfaction. Working examples reduce support burden and accelerate user onboarding. While critical for usability, it's lower priority than core functionality tests and API coverage.

**Failure Analysis Depth**: When examples fail validation, remediation plan must include root cause analysis (exact error, underlying cause - config issue vs provider bug), specific fix steps with code changes needed, and validation approach for each failing example.

**Independent Test**: Can be fully tested by running `scripts/test-examples.sh` against all examples in `examples/` directory and verifying each example passes init/validate/plan phases. Delivers a validation report showing which examples work and which need fixes.

**Acceptance Scenarios**:

1. **Given** all resource implementations, **When** I check `examples/resources/` directory, **Then** each resource must have at least one basic example and one advanced example (if applicable)
2. **Given** all data source implementations, **When** I check `examples/data-sources/` directory, **Then** each data source must have examples demonstrating basic retrieval and filtering
3. **Given** all examples in `examples/` directory, **When** I run `scripts/test-examples.sh`, **Then** every example must pass terraform init, validate, and plan without errors
4. **Given** resource examples, **When** I review example configurations, **Then** they must use unique test names (e.g., "citest-*" prefix) to support parallel testing and cleanup
5. **Given** provider examples in `examples/provider/`, **When** I review authentication patterns, **Then** they must use environment variables (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD) without hardcoded credentials
6. **Given** failing examples, **When** I document remediation needs, **Then** each failure must include root cause analysis, exact error details, specific code fix steps, and validation approach

---

### User Story 4 - Code Consistency Review (Priority: P2)

As a provider maintainer, I need consistent code patterns across all resources and data sources so that the codebase is maintainable, predictable, and follows HashiCorp best practices.

**Why this priority**: Code consistency improves maintainability and reduces cognitive load for contributors. While important for long-term health, it's lower priority than functional correctness and test coverage. This can be addressed incrementally.

**Compliance Standard**: Strict HashiCorp best practices compliance required. Flag any deviation from official HashiCorp/Terraform standards as consistency issues requiring remediation, even if current code works correctly. Use terraform-provider-design skill for authoritative best practices validation.

**Independent Test**: Can be fully tested by analyzing all resource/data source implementations against defined consistency criteria (error handling, schema patterns, BCM client usage, validation approaches). Delivers a consistency report with specific improvement recommendations.

**Acceptance Scenarios**:

1. **Given** all resource implementations, **When** I review error handling patterns, **Then** all resources must use consistent BCM API error parsing with multi-layer error detection per HashiCorp standards
2. **Given** all resource and data source schemas, **When** I review schema definitions, **Then** all attributes must have clear descriptions, appropriate validators, and correct computed/optional/required flags following HashiCorp conventions
3. **Given** all BCM API client usage, **When** I review CallJSONRPC patterns, **Then** resources must use direct lookup with args parameter (e.g., `getSoftwareImage(name)`) while data sources use list methods (e.g., `getSoftwareImages()`) with client-side filtering
4. **Given** all resource Read implementations, **When** I review eventual consistency handling, **Then** resources with async operations (cloning, provisioning) must use polling with exponential backoff per Terraform Plugin Framework best practices
5. **Given** all resource implementations, **When** I review state management, **Then** resources must never propagate Unknown values to state and must preserve plan values for fields that BCM API resets
6. **Given** any code pattern deviations from HashiCorp standards, **When** I document consistency issues, **Then** all deviations must be flagged for remediation regardless of whether current implementation functions correctly

---

### User Story 5 - Phased Remediation Plan (Priority: P3)

As a provider maintainer, I need a prioritized, phased remediation plan with clear success criteria so that I can systematically improve the provider to production-ready status with measurable progress.

**Why this priority**: A remediation plan organizes improvement work but doesn't directly deliver user value until implemented. It's the planning phase that enables the higher-priority work to be executed effectively.

**Independent Test**: Can be fully tested by validating that the remediation plan includes all identified issues grouped by priority with clear success criteria and regression testing requirements for each phase. Delivers a validated project plan.

**Acceptance Scenarios**:

1. **Given** all identified issues from test coverage, API gaps, documentation, and consistency reviews, **When** I create remediation plan, **Then** issues must be grouped into phases (Critical, High, Medium, Low priority)
2. **Given** each remediation phase, **When** I define success criteria, **Then** each phase must include specific deliverables with measurable outcomes
3. **Given** each remediation phase, **When** I define validation steps, **Then** phase completion must require full regression testing with `go test`, example validation with `test-examples.sh`, and BCM API validation against live cluster
4. **Given** remediation plan, **When** I estimate effort, **Then** each phase must have time estimates and resource requirements to guide implementation planning

---

### Edge Cases

- What happens when BCM API returns unexpected entity structures (missing fields, extra fields, polymorphic types)?
- How does the test suite handle BCM cluster state variations (different numbers of resources, different configurations)?
- What happens when examples reference resources that don't exist in a particular BCM cluster?
- How does the provider handle BCM API version differences or deprecated endpoints?
- What happens when drift detection tests run concurrently with other operations modifying the same resources?
- How do we handle BCM API rate limiting or timeout scenarios in acceptance tests?
- What happens when deep root cause analysis for example failures requires iterative debugging cycles that extend total review time significantly?
- How do we balance thoroughness in HashiCorp best practices validation against the volume of code to review?
- What happens when BCM API methods discovered via live queries are deprecated or become inaccessible in future BCM versions?

## Requirements *(mandatory)*

### Functional Requirements

#### Test Coverage Analysis

- **FR-001**: System MUST analyze all resource implementation files (`resource_*.go`) and identify which CRUD operations (Create, Read, Update, Delete) have corresponding acceptance tests
- **FR-002**: System MUST analyze all resource test files and verify presence of Import functionality tests using `ImportState: true` pattern
- **FR-003**: System MUST analyze all resource test files and verify presence of Drift detection tests using `PreConfig` modification and `plancheck.ExpectNonEmptyPlan()` pattern
- **FR-004**: System MUST analyze all resource test files and verify presence of Idempotency tests using `plancheck.ExpectEmptyPlan()` after Create and Update operations
- **FR-005**: System MUST identify usage of legacy testing patterns (resource.TestCheckResourceAttr for all checks) vs modern patterns (statecheck.ExpectKnownValue, plancheck, compareValue)
- **FR-006**: System MUST identify hardcoded test assumptions (resource counts, specific names/UUIDs, cluster-specific state) that violate environment portability
- **FR-007**: System MUST verify CheckDestroy functions use enhanced patterns with detailed error messages and exponential backoff verification via `verifyResourceDeleted()`
- **FR-008**: System MUST analyze data source test files and verify filter functionality is tested with appropriate type-safe checks

#### BCM API Gap Analysis

- **FR-009**: System MUST query live BCM API endpoint to enumerate all available services (CMDevice, CMPart, CMNet, CMProv, CMJob, CMServ, CMMon, CMKube, CMCloud, etc.) as the authoritative source of truth
- **FR-010**: System MUST query each BCM service to discover all available methods by introspection or systematic exploration of the API
- **FR-011**: System MUST compare live BCM API methods against implemented resources in `provider.go` Resources() method
- **FR-012**: System MUST compare live BCM API methods against implemented data sources in `provider.go` DataSources() method
- **FR-013**: System MUST categorize API gaps by service type and assess business value (high-value: core device/network/software management, medium-value: monitoring/jobs, low-value: GUI settings) based on actual API capabilities
- **FR-014**: System MUST identify BCM API methods that have partial implementation (e.g., only data source exists but no resource, or vice versa)
- **FR-015**: System MAY cross-reference `sampleRest/` documentation as supplementary reference material but live API is authoritative

#### Documentation & Examples Validation

- **FR-016**: System MUST verify each registered resource has at least one working example in `examples/resources/<resource_name>/` directory
- **FR-017**: System MUST verify each registered data source has at least one working example in `examples/data-sources/<datasource_name>/` directory
- **FR-018**: System MUST execute `scripts/test-examples.sh` and verify all examples pass terraform init, validate, and plan phases without errors
- **FR-019**: System MUST verify all resource examples use unique naming patterns (e.g., "citest-" prefix with timestamps) to support parallel testing
- **FR-020**: System MUST verify all examples use environment variables for authentication (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD) without hardcoded credentials
- **FR-021**: System MUST verify generated documentation in `docs/` directory is current and matches example configurations
- **FR-022**: System MUST perform root cause analysis for each failing example including exact error message, underlying cause determination (configuration issue vs provider bug), specific code changes needed to fix, and validation approach to confirm fix

#### Code Consistency Analysis

- **FR-023**: System MUST analyze error handling patterns across all resources and verify consistent use of `parseErrorResponse()` for BCM API errors per HashiCorp best practices
- **FR-024**: System MUST analyze schema definitions across all resources/data sources and verify all attributes have clear descriptions following HashiCorp conventions
- **FR-025**: System MUST analyze BCM client usage patterns and verify resources use direct lookup with args parameter while data sources use list methods with client-side filtering
- **FR-026**: System MUST identify resources with async operations (cloning, provisioning) and verify they implement polling with exponential backoff per Terraform Plugin Framework standards
- **FR-027**: System MUST verify resources never propagate Unknown values to state and preserve plan values for fields that BCM API resets (e.g., original_image after cloning)
- **FR-028**: System MUST identify schema validators and verify consistent validation approaches across similar attribute types following HashiCorp patterns
- **FR-029**: System MUST flag all deviations from HashiCorp/Terraform official standards as consistency issues requiring remediation, even if current implementation functions correctly

#### Phased Remediation Planning

- **FR-030**: System MUST group all identified issues into priority levels (Critical, High, Medium, Low) based on production-readiness impact
- **FR-031**: System MUST define clear success criteria for each remediation phase with measurable outcomes
- **FR-032**: System MUST specify regression testing requirements for each phase including `go test`, `test-examples.sh`, and BCM API validation
- **FR-033**: System MUST estimate effort and resource requirements for each remediation phase

### Key Entities *(include if feature involves data)*

- **Test Coverage Report**: Contains resource/data source name, CRUD operation coverage, Import test presence, Drift test presence, Idempotency test presence, testing pattern modernization status, environment portability issues
- **API Gap Record**: Contains BCM service name, API method name, implementation status (not implemented, data source only, resource only, fully implemented), business value priority
- **Example Validation Result**: Contains example file path, terraform init status, terraform validate status, terraform plan status, error messages if any, portability issues detected
- **Consistency Issue**: Contains file path, issue type (error handling, schema, client usage, async handling, state management, validation), severity level, specific code location, recommended fix
- **Remediation Phase**: Contains phase name, priority level, issue list, success criteria, regression test requirements, effort estimate, dependencies on other phases

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-001**: Test coverage analysis identifies exactly which CRUD operations, Import tests, Drift tests, and Idempotency tests are missing for each resource, with 100% coverage required for production readiness (no exceptions permitted)
- **SC-002**: 100% of existing resources and data sources have complete test coverage assessment documented in structured format (e.g., markdown tables with checkboxes), identifying all gaps that must be remediated
- **SC-003**: BCM API gap analysis produces a prioritized list of at least 10 high-value missing resources/data sources that could be implemented to improve provider completeness, based on live BCM API as the authoritative source of truth
- **SC-004**: All examples in `examples/` directory pass validation via `scripts/test-examples.sh` with 100% success rate, or specific failing examples are documented with root cause analysis, specific fix steps, and validation approach
- **SC-005**: Code consistency review identifies specific deviations from HashiCorp/Terraform best practices across at least 4 categories (error handling, schema, client usage, async handling) with concrete file locations and line numbers, flagging all deviations regardless of current functionality
- **SC-006**: Remediation plan groups all identified issues into 3-5 distinct phases with clear boundaries, dependencies, and regression testing requirements
- **SC-007**: Each remediation phase has quantifiable success criteria that can be verified through automated testing (go test pass rate, example validation pass rate, etc.)
- **SC-008**: Review process prioritizes thoroughness and accuracy over speed, taking necessary time to produce comprehensive, actionable artifacts (reports, task lists, prioritized backlog items) that enable immediate implementation work

## Assumptions *(mandatory)*

- BCM cluster at configured endpoint (https://172.21.15.254:8081) is available and accessible for API queries during analysis
- Acceptance test environment variables (TF_ACC=1, BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD) are configured correctly for running tests
- All current resource and data source implementations follow the JSON-RPC pattern documented in `internal/provider/bcm_client.go`
- BCM API documentation in `sampleRest/` directory is reasonably current and reflects actual API capabilities
- Test helpers in `internal/provider/test_helpers.go` are available for use in enhanced CheckDestroy patterns
- Modern testing patterns using terraform-plugin-testing v1.13.3+ are the target standard for all tests
- Examples testing infrastructure in `scripts/test-examples.sh` is functional and can be executed against current examples
- Provider follows Terraform Plugin Framework v1.16.1 conventions for resource/data source registration
- All code follows TDD principles with RED-GREEN-REFACTOR cycles as documented in `/workspace/AGENTS.md`
- terraform-provider-design skill is available for querying HashiCorp best practices during review

## Dependencies *(mandatory)*

- Access to BCM cluster for API enumeration and validation testing
- Terraform Plugin Framework v1.16.1 for understanding resource/data source patterns
- terraform-plugin-testing v1.13.3 for modern testing pattern requirements
- Go 1.24+ for running acceptance tests and analyzing code
- BCM API documentation files in `sampleRest/` directory
- Test execution infrastructure (`scripts/test-examples.sh`, environment variables)
- terraform-provider-design skill for best practices validation

## Out of Scope *(mandatory)*

- Implementation of fixes for identified issues (this spec only covers analysis and planning)
- Development of new resources or data sources for identified API gaps
- Modification of BCM API or cluster configuration
- Performance testing or load testing of the provider
- Security audit or penetration testing
- Multi-cluster or multi-tenant configuration testing
- Provider versioning strategy or release planning
- Terraform Cloud/Enterprise integration testing
- Cross-platform compatibility testing (Windows, macOS) beyond Linux development environment
- Migration guides for users upgrading from previous provider versions
- Internationalization or localization of error messages

## Open Questions *(optional)*

None at this time. The specification is complete with clear scope boundaries and testable acceptance criteria.

## Technical Constraints *(optional)*

- Analysis must be performed in Linux development environment (per env context)
- BCM API queries must handle self-signed certificates (insecure_skip_verify may be required)
- Test execution is limited by BCM cluster performance and API rate limits
- Acceptance tests create real resources in BCM cluster and may impact cluster state temporarily
- Example testing runs sequentially for resources (to avoid conflicts) but can run in parallel for data sources
- Some BCM API operations are asynchronous and may require polling with timeouts (e.g., image cloning up to 60 seconds)
- BCM API uses cookie-based authentication with session management that may timeout during long analysis runs

## Future Considerations *(optional)*

- Automated test coverage monitoring in CI/CD pipeline
- API gap analysis as part of regular provider maintenance cycle
- Example validation as pre-merge requirement for pull requests
- Code consistency enforcement via linting rules or pre-commit hooks
- Drift detection testing for all resources as standard practice
- Comprehensive integration test suite covering multi-resource workflows
- BCM API version compatibility matrix and deprecation tracking
- Provider capability matrix published for users to understand supported operations
