# TERRAFORM PROVIDER TDD PARALLEL EXECUTION

🚨 CRITICAL: MANDATORY RULE: In Terraform Provider TDD workflows, ALL test and implementation cycles MUST be parallel where possible:

Use the terraform-provider-design skill for Terraform provider design principles and best practices

🔴 TDD-SPECIFIC CONCURRENT PATTERNS FOR TERRAFORM PROVIDERS:
Parallel Test Writing: Create multiple acceptance tests simultaneously for resources/data sources
Concurrent Implementation: Implement multiple CRUD operations in parallel
Batch Refactoring: Refactor multiple provider resources simultaneously after green phase
Parallel Test Suites: Run unit tests, acceptance tests, and integration tests concurrently
Simultaneous Documentation: Update provider docs and tests in parallel using tfplugindocs
⚡ TDD GOLDEN RULE: "RED-GREEN-REFACTOR IN PARALLEL BATCHES FOR TERRAFORM PROVIDERS"
✅ CORRECT TDD Pattern for Terraform Providers:

// RED PHASE - Write failing acceptance tests in parallel

[BatchTool - RED Phase]:

- Write("internal/provider/instance_resource_test.go", failingInstanceResourceTests)
- Write("internal/provider/network_resource_test.go", failingNetworkResourceTests)
- Write("internal/provider/user_data_source_test.go", failingUserDataSourceTests)
- Write("internal/provider/policy_resource_test.go", failingPolicyResourceTests)
- Bash("TF_ACC=1 go test -v -timeout 120m ./internal/provider/") // Verify all tests fail

// GREEN PHASE - Implement CRUD operations in parallel

[BatchTool - GREEN Phase]:

- Write("internal/provider/instance_resource.go", minimalInstanceResourceImplementation)
- Write("internal/provider/network_resource.go", minimalNetworkResourceImplementation)
- Write("internal/provider/user_data_source.go", minimalUserDataSourceImplementation)
- Write("internal/provider/policy_resource.go", minimalPolicyResourceImplementation)
- Bash("TF_ACC=1 go test -v -timeout 120m ./internal/provider/") // Verify all tests pass

// REFACTOR PHASE - Improve provider code in parallel

[BatchTool - REFACTOR Phase]:

- Edit("internal/provider/instance_resource.go", refactoredInstanceCode)
- Edit("internal/provider/network_resource.go", refactoredNetworkCode)
- Edit("internal/provider/user_data_source.go", refactoredUserCode)
- Edit("internal/provider/policy_resource.go", refactoredPolicyCode)
- Bash("TF_ACC=1 go test -v -timeout 120m ./internal/provider/") // Verify tests still pass
- Bash("go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate") // Generate docs

🎯 TERRAFORM PROVIDER PROJECT CONTEXT

Development Philosophy
🔴 Red: Write failing acceptance tests first (define resource/data source behavior)
🟢 Green: Write minimal CRUD code to pass tests (make it work)
🔄 Refactor: Improve provider code while keeping tests green (make it clean)
📊 Coverage: Maintain comprehensive acceptance test coverage
🧪 Testing Pyramid: Acceptance Tests > Unit Tests > Integration Tests

Terraform Provider Stack
Framework: Terraform Plugin Framework (recommended) or SDKv2
Testing: terraform-plugin-testing (acceptance tests with TF_ACC=1)
Language: Go (>= 1.24)
Documentation: terraform-plugin-docs (tfplugindocs)
State Management: terraform-plugin-framework types (StringValue, Int64Value, etc.)
API Client: Standard Go http.Client or vendor-specific SDK

🔧 TERRAFORM PROVIDER DEVELOPMENT PATTERNS

Provider Project Structure

// Standard Terraform Provider Framework structure
terraform-provider-bcm/
├── internal/
│ └── provider/
│ ├── provider.go // Provider definition
│ ├── provider_test.go // Provider tests
│ ├── resource_instance.go // Resource implementations
│ ├── resource_instance_test.go // Resource acceptance tests
│ ├── data_source_user.go // Data source implementations
│ └── data_source_user_test.go // Data source acceptance tests
├── examples/
│ ├── provider/ // Provider configuration examples
│ ├── resources/ // Resource examples
│ └── data-sources/ // Data source examples
├── docs/ // Generated documentation
├── main.go // Provider binary entry point
├── go.mod // Go module dependencies
├── go.sum // Go module checksums
├── .goreleaser.yml // Release configuration
└── GNUmakefile // Build and test commands

TDD Cycle Implementation

// Parallel TDD cycle execution for Terraform providers

[BatchTool - Complete TDD Cycle]:

// 1. RED: Write failing acceptance test

- Write("internal/provider/instance_resource_test.go", `
  package provider

import (
"fmt"
"testing"
"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccInstanceResource(t \*testing.T) {
resource.Test(t, resource.TestCase{
PreCheck: func() { testAccPreCheck(t) },
ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
Steps: []resource.TestStep{
// Create and Read testing
{
Config: testAccInstanceResourceConfig("test-instance"),
Check: resource.ComposeAggregateTestCheckFunc(
resource.TestCheckResourceAttr("bcm_instance.test", "name", "test-instance"),
resource.TestCheckResourceAttrSet("bcm_instance.test", "id"),
),
},
// ImportState testing
{
ResourceName: "bcm_instance.test",
ImportState: true,
ImportStateVerify: true,
},
// Update and Read testing
{
Config: testAccInstanceResourceConfig("test-instance-updated"),
Check: resource.ComposeAggregateTestCheckFunc(
resource.TestCheckResourceAttr("bcm_instance.test", "name", "test-instance-updated"),
),
},
// Delete testing automatically occurs in TestCase
},
})
}

func testAccInstanceResourceConfig(name string) string {
return fmt.Sprintf(\`
resource "bcm_instance" "test" {
name = %[1]q
}
\`, name)
}
`)

// 2. GREEN: Minimal resource implementation

- Write("internal/provider/instance_resource.go", `
  package provider

import (
"context"
"github.com/hashicorp/terraform-plugin-framework/path"
"github.com/hashicorp/terraform-plugin-framework/resource"
"github.com/hashicorp/terraform-plugin-framework/resource/schema"
"github.com/hashicorp/terraform-plugin-framework/types"
"github.com/hashicorp/terraform-plugin-log/tflog"
)

var \_ resource.Resource = \&InstanceResource{}  
var \_ resource.ResourceWithImportState = \&InstanceResource{}

type InstanceResource struct {
client \\\*http.Client
}

type InstanceResourceModel struct {
ID types.String \\`tfsdk:\"id\"\\`
Name types.String \\`tfsdk:\"name\"\\`
}

func (r \*InstanceResource) Schema(ctx context.Context, req resource.SchemaRequest, resp \*resource.SchemaResponse) {
resp.Schema = schema.Schema{
MarkdownDescription: "Instance resource",
Attributes: map[string]schema.Attribute{
"id": schema.StringAttribute{
Computed: true,
MarkdownDescription: "Instance identifier",
},
"name": schema.StringAttribute{
Required: true,
MarkdownDescription: "Instance name",
},
},
}
}

func (r \*InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp \*resource.CreateResponse) {
var data InstanceResourceModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
return
}

    // Minimal implementation - hardcoded ID
    data.ID = types.StringValue("instance-123")

    tflog.Trace(ctx, "created instance resource")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r \*InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp \*resource.ReadResponse) {
var data InstanceResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
return
}

    // Minimal read - state already has data
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r \*InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp \*resource.UpdateResponse) {
var data InstanceResourceModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
return
}

    tflog.Trace(ctx, "updated instance resource")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r \*InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp \*resource.DeleteResponse) {
var data InstanceResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
return
}

    // Minimal delete - nothing to do
    tflog.Trace(ctx, "deleted instance resource")

}

func (r \*InstanceResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp \*resource.ImportStateResponse) {
resource.ImportStatePassthroughID(ctx, path.Root("id"), req, resp)
}
`)

// 3. REFACTOR: Add full CRUD and API integration

- Edit("internal/provider/instance_resource.go", `
  func (r \*InstanceResource) Create(ctx context.Context, req resource.CreateRequest, resp \*resource.CreateResponse) {
  var data InstanceResourceModel
  resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
  if resp.Diagnostics.HasError() {
  return
  }
  // Make actual API call to create instance
  instance, err := r.client.CreateInstance(ctx, data.Name.ValueString())
  if err != nil {
  resp.Diagnostics.AddError(
  "Error Creating Instance",
  "Could not create instance, unexpected error: "+err.Error(),
  )
  return
  }

      data.ID = types.StringValue(instance.ID)
      tflog.Trace(ctx, "created instance resource")
      resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

  }

func (r \*InstanceResource) Read(ctx context.Context, req resource.ReadRequest, resp \*resource.ReadResponse) {
var data InstanceResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
return
}

    // Fetch instance from API
    instance, err := r.client.GetInstance(ctx, data.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Reading Instance",
            "Could not read instance ID "+data.ID.ValueString()+": "+err.Error(),
        )
        return
    }

    data.Name = types.StringValue(instance.Name)
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r \*InstanceResource) Update(ctx context.Context, req resource.UpdateRequest, resp \*resource.UpdateResponse) {
var data InstanceResourceModel
resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
return
}

    // Update instance via API
    err := r.client.UpdateInstance(ctx, data.ID.ValueString(), data.Name.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Updating Instance",
            "Could not update instance, unexpected error: "+err.Error(),
        )
        return
    }

    tflog.Trace(ctx, "updated instance resource")
    resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)

}

func (r \*InstanceResource) Delete(ctx context.Context, req resource.DeleteRequest, resp \*resource.DeleteResponse) {
var data InstanceResourceModel
resp.Diagnostics.Append(req.State.Get(ctx, &data)...)
if resp.Diagnostics.HasError() {
return
}

    // Delete instance via API
    err := r.client.DeleteInstance(ctx, data.ID.ValueString())
    if err != nil {
        resp.Diagnostics.AddError(
            "Error Deleting Instance",
            "Could not delete instance, unexpected error: "+err.Error(),
        )
        return
    }

    tflog.Trace(ctx, "deleted instance resource")

}
`)

// 4. Run acceptance tests to verify

- Bash("TF_ACC=1 go test -v -timeout 120m ./internal/provider/")

🧪 ACCEPTANCE TEST BEST PRACTICES

Key Acceptance Testing Patterns

// HashiCorp-recommended acceptance test patterns
acceptance_test_patterns: {
test_structure: {
create_and_read: "First test step verifies resource creation and read",
import_state: "Second step tests terraform import functionality",
update_and_read: "Third step verifies updates are applied correctly",
delete: "Automatically tested when TestCase completes"
},

test_helpers: {
config_functions: "Use fmt.Sprintf with %[1]q for safe string interpolation",
unique_names: "Generate unique resource names to avoid conflicts",
precheck: "Verify required environment variables in testAccPreCheck",
provider_factories: "Define testAccProtoV6ProviderFactories once, reuse everywhere"
},

state_checks: {
TestCheckResourceAttr: "Verify exact attribute values",
TestCheckResourceAttrSet: "Verify attribute exists with any value",
TestCheckResourceAttrPair: "Verify two resources have matching attributes",
ComposeAggregateTestCheckFunc: "Combine multiple checks, continue on failure"
},

import_state: {
required: "All resources MUST implement ImportState",
verify: "Set ImportStateVerify: true to compare imported vs actual state",
ignore: "Use ImportStateVerifyIgnore for computed-only attributes",
passthrough: "Use resource.ImportStatePassthroughID for simple imports"
}
}

Common Acceptance Test Patterns

// Provider test setup pattern
provider_test_setup: `
// internal/provider/provider_test.go
package provider

import (
"testing"
"github.com/hashicorp/terraform-plugin-framework/providerserver"
"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories are used to instantiate a provider during
// acceptance testing. The factory function will be invoked for every Terraform
// CLI command executed to create a provider server to which the CLI can
// reattach.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
"bcm": providerserver.NewProtocol6WithError(New("test")()),
}

func testAccPreCheck(t \*testing.T) {
// Check for required environment variables
// if v := os.Getenv("BCM_API_KEY"); v == "" {
// t.Fatal("BCM_API_KEY must be set for acceptance tests")
// }
}
`

// Complete CRUD test pattern
complete_crud_test: `
func TestAccInstanceResource(t \*testing.T) {
resource.Test(t, resource.TestCase{
PreCheck: func() { testAccPreCheck(t) },
ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
Steps: []resource.TestStep{
// Create and Read testing
{
Config: testAccInstanceResourceConfig("test"),
Check: resource.ComposeAggregateTestCheckFunc(
resource.TestCheckResourceAttr("bcm_instance.test", "name", "test"),
resource.TestCheckResourceAttrSet("bcm_instance.test", "id"),
resource.TestCheckResourceAttrSet("bcm_instance.test", "created_at"),
),
},
// ImportState testing
{
ResourceName: "bcm_instance.test",
ImportState: true,
ImportStateVerify: true,
// Ignore computed fields that can't be imported
ImportStateVerifyIgnore: []string{"last_updated"},
},
// Update and Read testing
{
Config: testAccInstanceResourceConfig("test-updated"),
Check: resource.ComposeAggregateTestCheckFunc(
resource.TestCheckResourceAttr("bcm_instance.test", "name", "test-updated"),
),
},
// Delete testing automatically occurs in TestCase
},
})
}

func testAccInstanceResourceConfig(name string) string {
return fmt.Sprintf(\`
resource "bcm_instance" "test" {
name = %[1]q
}
\`, name)
}
`

// Data source test pattern
data_source_test: `
func TestAccUserDataSource(t \*testing.T) {
resource.Test(t, resource.TestCase{
PreCheck: func() { testAccPreCheck(t) },
ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
Steps: []resource.TestStep{
// Read testing
{
Config: testAccUserDataSourceConfig,
Check: resource.ComposeAggregateTestCheckFunc(
resource.TestCheckResourceAttr("data.bcm_user.test", "username", "admin"),
resource.TestCheckResourceAttrSet("data.bcm_user.test", "id"),
),
},
},
})
}

const testAccUserDataSourceConfig = \`
data "bcm_user" "test" {
username = "admin"
}
\`
`

Enhanced CheckDestroy and PreCheck Patterns

// Idempotent CheckDestroy implementation with exponential backoff
idempotent_check_destroy: `
// CheckDestroy verifies resources are deleted after test completes
// Must be idempotent and handle eventual consistency
func testAccCheckCMPartSoftwareImageDestroy(s *terraform.State) error {
    client := createTestBCMClient(&testing.T{})
    ctx := context.Background()

    for _, rs := range s.RootModule().Resources {
        if rs.Type != "bcm_cmpart_softwareimage" {
            continue
        }

        imageName := rs.Primary.Attributes["name"]

        // Use exponential backoff verification (4 retries: 1s, 2s, 4s, 8s = 15s total)
        deleted, err := verifyResourceDeleted(ctx, client, "CMPart", "getSoftwareImage", imageName, 4)

        if err != nil {
            // API errors during verification are logged but not fatal
            // Resource may already be deleted or API temporarily unavailable
            fmt.Printf("[DEBUG] CheckDestroy verification error for %s: %v\n", imageName, err)
        }

        if !deleted {
            // Resource still exists after retries - this is the ONLY error condition
            return fmt.Errorf("Software image %s still exists after destroy", imageName)
        }
    }

    return nil
}

// verifyResourceDeleted is a shared helper with exponential backoff
// Located in test_helpers.go for reuse across all resource tests
func verifyResourceDeleted(ctx context.Context, client *BCMClient, service, method, identifier string, maxRetries int) (bool, error) {
    waitTime := 1 * time.Second

    for retry := 0; retry < maxRetries; retry++ {
        time.Sleep(waitTime)

        // Attempt to read resource
        body, err := client.CallJSONRPC(ctx, service, method, identifier)

        // Error response indicates resource not found (deleted)
        if err != nil {
            return true, nil
        }

        // Empty response indicates resource deleted
        if len(body) == 0 {
            return true, nil
        }

        // Empty JSON object indicates resource deleted
        var data map[string]interface{}
        if json.Unmarshal(body, &data) == nil && len(data) == 0 {
            return true, nil
        }

        // Resource still exists, exponential backoff
        waitTime *= 2
    }

    // Resource still exists after all retries
    return false, nil
}
`

// PreCheck with cleanup function for leftover test resources
precheck_with_cleanup: `
// testAccPreCheckCMPartSoftwareImage verifies environment and cleans up leftover resources
func testAccPreCheckCMPartSoftwareImage(t *testing.T) {
    // Standard provider credential checks
    testAccPreCheck(t)

    // Cleanup any leftover test resources from previous failed runs
    cleanupLeftoverSoftwareImages(t)
}

// cleanupLeftoverSoftwareImages removes test resources matching pattern
func cleanupLeftoverSoftwareImages(t *testing.T) {
    client := createTestBCMClient(t)
    ctx := context.Background()

    // Query all software images
    body, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImages")
    if err != nil {
        t.Logf("[WARN] Failed to query software images for cleanup: %v", err)
        return
    }

    var images []map[string]interface{}
    if err := json.Unmarshal(body, &images); err != nil {
        t.Logf("[WARN] Failed to parse software images: %v", err)
        return
    }

    // Find and remove test images (prefix "test-")
    var testImages []string
    for _, img := range images {
        if name, ok := img["name"].(string); ok {
            if strings.HasPrefix(name, "test-") {
                testImages = append(testImages, name)
            }
        }
    }

    if len(testImages) > 0 {
        t.Logf("[INFO] Cleaning up %d leftover test images", len(testImages))

        // BCM API expects array of names
        _, err := client.CallJSONRPC(ctx, "CMPart", "removeSoftwareImages", testImages, false)
        if err != nil {
            t.Logf("[WARN] Failed to remove test images: %v", err)
        }

        // Wait for cleanup to complete
        time.Sleep(2 * time.Second)
    }
}
`

// Shared test helper functions (test_helpers.go)
shared_test_helpers: `
// createTestBCMClient creates authenticated BCM client for test use
// Used by: CheckDestroy, PreCheck cleanup, drift detection PreConfig
func createTestBCMClient(t *testing.T) *BCMClient {
    endpoint := os.Getenv("BCM_ENDPOINT")
    username := os.Getenv("BCM_USERNAME")
    password := os.Getenv("BCM_PASSWORD")

    if endpoint == "" || username == "" || password == "" {
        t.Fatalf("BCM credentials not set (BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD)")
    }

    client, err := NewBCMClient(context.Background(), endpoint, username, password, true, 30)
    if err != nil {
        t.Fatalf("Failed to create BCM client: %v", err)
    }

    return client
}

// getResourceUUIDByName queries BCM API for resource UUID by name
// Used by: Drift detection tests to find UUID for external modifications
func getResourceUUIDByName(t *testing.T, service, method, name string) string {
    client := createTestBCMClient(t)
    ctx := context.Background()

    body, err := client.CallJSONRPC(ctx, service, method, name)
    if err != nil {
        t.Fatalf("Failed to query resource %s via %s.%s: %v", name, service, method, err)
    }

    var resourceData map[string]interface{}
    if err := json.Unmarshal(body, &resourceData); err != nil {
        t.Fatalf("Failed to parse resource response: %v", err)
    }

    uuid, ok := resourceData["uuid"].(string)
    if !ok || uuid == "" {
        t.Fatalf("Resource %s does not have a valid uuid field", name)
    }

    return uuid
}

// generateUniqueTestName creates unique resource names with timestamp
// Prevents conflicts in parallel tests or after cleanup failures
func generateUniqueTestName(prefix string) string {
    timestamp := time.Now().Format("20060102-150405")
    return prefix + "-" + timestamp
}
`

🐝 TDD SWARM ORCHESTRATION
TDD-Specialized Agent Roles for Terraform Providers
test_designer:
role: Acceptance Test Specification Designer
focus: [resource-behavior, crud-operations, state-management]
responsibilities:

- Design comprehensive acceptance test suites
- Define resource schemas and attributes
- Create test fixtures and configurations

  concurrent_tasks: [multiple-resource-tests, data-source-tests]

red_phase_agent:
role: Failing Acceptance Test Creator
focus: [acceptance-tests, terraform-configs, test-first-development]
responsibilities:

- Write failing acceptance tests for resources/data sources
- Define expected Terraform state and attributes
- Create test configurations with resource blocks

  concurrent_tasks: [multiple-failing-acceptance-tests, schema-definition]

green_phase_agent:
role: Minimal Implementation Creator
focus: [minimal-code, pass-tests, quick-implementation]
responsibilities:

- Write minimal code to pass tests
- Avoid over-engineering in green phase
- Focus on making tests pass quickly

  concurrent_tasks: [multiple-implementations, simple-solutions]

refactor_agent:
role: Code Quality Improver
focus: [clean-code, design-patterns, optimization]
responsibilities:

- Improve code quality while keeping tests green
- Apply design patterns and best practices
- Optimize performance and maintainability

  concurrent_tasks: [multiple-refactors, pattern-application]

coverage_analyst:
role: Test Coverage Monitor
focus: [coverage-analysis, gap-identification, quality-metrics]
responsibilities:

- Monitor test coverage metrics
- Identify untested code paths
- Ensure comprehensive test suites

  concurrent_tasks: [coverage-analysis, gap-reporting]

Provider Development Topology Recommendation

    # For Terraform Provider projects - use mesh topology for collaborative development
    claude-flow hive init --topology mesh --agents 5

    # Agent distribution for provider development
    # - 1 Test Designer (acceptance test specification)
    # - 1 Red Phase Agent (failing acceptance tests)
    # - 1 Green Phase Agent (minimal CRUD implementation)
    # - 1 Refactor Agent (code improvement and optimization)
    # - 1 Coverage Analyst (test coverage and quality assurance)

🧠 TERRAFORM PROVIDER MEMORY MANAGEMENT

Provider Context Storage

// Store provider-specific project context
provider_memory_patterns: {
"provider/testing-strategy": "Acceptance-first with full CRUD coverage",
"provider/test-framework": "terraform-plugin-testing with TF_ACC=1",
"provider/api-client": "Custom Go client with retry logic",
"provider/state-management": "terraform-plugin-framework types",
"provider/test-data-strategy": "Unique resource names with timestamp suffixes",
"provider/ci-integration": "Acceptance tests on every commit with API mocking"
}
Provider TDD Cycle Tracking

// Track provider TDD cycles and decisions
provider_cycles: {
"cycle_001_instance_resource": {
"red_phase": {
"tests_written": ["create_instance", "read_instance", "update_instance", "delete_instance", "import_instance"],
"expected_failures": 5,
"actual_failures": 5,
"status": "completed"
},
"green_phase": {
"implementation_approach": "minimal CRUD with hardcoded responses",
"tests_passing": 5,
"time_to_green": "30 minutes",
"status": "completed"
},
"refactor_phase": {
"improvements": ["added API client", "added error handling", "added state validation"],
"patterns_applied": ["resource interface", "diagnostic handling", "context propagation"],
"final_test_status": "all passing with real API",
"status": "completed"
}
}
}
🚀 TDD CI/CD PIPELINE
Terraform Provider CI/CD Pipeline (Parallel Execution)

Provider-focused CI/CD pipeline:

provider_pipeline:
quality_gates: - name: "Acceptance Test Pass Rate"
threshold: "100%"
action: "fail_build_if_below"

    - name: "golangci-lint"
      action: "fail_on_any_issue"

    - name: "Documentation Generation"
      action: "fail_if_outdated"

parallel_stages:
unit_tests: - "go test -v -cover ./..." - "go test -race ./..."

    acceptance_tests:
      - "TF_ACC=1 go test -v -timeout 120m ./internal/provider/"
      - "TF_ACC=1 go test -v -parallel=4 -timeout 120m ./..."

    static_analysis:
      - "golangci-lint run"
      - "go vet ./..."
      - "go mod verify"

    documentation:
      - "go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs generate"
      - "git diff --exit-code docs/" # Ensure docs are up to date

Terraform Provider Environment Setup

Development environment for provider development:

TF_ACC=0 # Disable acceptance tests for unit tests
TF_LOG=DEBUG # Enable detailed Terraform logging
TF_LOG_PATH=./terraform.log # Log file path
API_ENDPOINT=<http://localhost:8080> # Mock API endpoint for testing
API_KEY=test-key # Test API credentials

Acceptance testing environment:

TF_ACC=1 # Enable acceptance tests
TF_ACC_TERRAFORM_VERSION=1.5.0 # Terraform version for testing
TF_LOG=TRACE # Maximum verbosity for debugging
REAL_API_ENDPOINT=<https://api.example.com> # Real API endpoint
REAL_API_KEY=\${INTEGRATION_TEST_KEY} # Real credentials from CI secrets

📊 TERRAFORM PROVIDER MONITORING & METRICS

Provider Test Quality Metrics

// Provider quality tracking
provider_metrics: {
test_coverage: {
acceptance_tests: "All CRUD operations covered",
import_tests: "All resources importable",
data_sources: "All data sources tested",
edge_cases: "Error handling validated"
},

test_quality: {
acceptance_pass_rate: "100%",
unit_test_coverage: ">=80%",
test_execution_time: "<120m for full acceptance suite",
parallel_execution: "4-8 parallel tests"
},

tdd_adherence: {
red_green_cycles: "tracked per resource/data source",
test_first_percentage: ">=95%",
refactor_frequency: "after every green phase",
api_compatibility: "always maintained"
}
}

Provider Test Reporting

// Comprehensive provider test reporting
provider_reporting: {
coverage_report: "go test -cover output",
test_results: "Verbose test output with timing",
acceptance_logs: "TF_LOG=TRACE for debugging",
performance: "Test execution time per resource",
documentation: "tfplugindocs generated docs"
}
🔒 TERRAFORM PROVIDER SECURITY & TESTING
Security-First Testing for Providers
// Security testing in provider TDD cycles
provider_security_patterns: {
sensitive_data: "Mark sensitive attributes (passwords, tokens) appropriately",
input_validation: "Validate all resource attribute inputs in Schema",
api_credentials: "Test authentication failures before implementing API calls",
state_security: "Ensure sensitive data handling in state operations",
error_messages: "Avoid leaking sensitive info in error messages",
tls_verification: "Test API client TLS configuration"
}
Provider Security Test Examples

// Example security-focused provider TDD

[BatchTool - Provider Security TDD]:

- Write("internal/provider/api_key_resource_test.go", `
  package provider

import (
"testing"
"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccAPIKeyResource_Sensitive(t \*testing.T) {
resource.Test(t, resource.TestCase{
PreCheck: func() { testAccPreCheck(t) },
ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
Steps: []resource.TestStep{
{
Config: testAccAPIKeyResourceConfig("test-key"),
Check: resource.ComposeAggregateTestCheckFunc(
resource.TestCheckResourceAttr("bcm_api_key.test", "name", "test-key"),
// API key value should not be in state plainly
resource.TestCheckResourceAttrSet("bcm_api_key.test", "key_id"),
),
},
},
})
}
`)

// Implement with sensitive attribute handling

- Write("internal/provider/api_key_resource.go", `
  package provider

import (
"github.com/hashicorp/terraform-plugin-framework/resource/schema"
"github.com/hashicorp/terraform-plugin-framework/types"
)

func (r \*APIKeyResource) Schema(ctx context.Context, req resource.SchemaRequest, resp \*resource.SchemaResponse) {
resp.Schema = schema.Schema{
Attributes: map[string]schema.Attribute{
"name": schema.StringAttribute{
Required: true,
},
"key_value": schema.StringAttribute{
Computed: true,
Sensitive: true, // Mark as sensitive
},
},
}
}
`)
🧪 ADVANCED TERRAFORM PROVIDER TESTING TECHNIQUES

Table-Driven Tests

// Go's table-driven test pattern for providers

[BatchTool - Table-Driven Testing]:

- Write("internal/provider/validation_test.go", `
  package provider

import (
"testing"
)

func TestValidateResourceName(t \*testing.T) {
tests := []struct {
name string
input string
wantErr bool
}{
{"valid name", "my-resource-123", false},
{"empty name", "", true},
{"too long", string(make([]byte, 256)), true},
{"invalid chars", "my_resource!", true},
}

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := validateResourceName(tt.input)
            if (err != nil) != tt.wantErr {
                t.Errorf("validateResourceName() error = %v, wantErr %v", err, tt.wantErr)
            }
        })
    }

}
`)

Fuzz Testing Integration

// Go native fuzzing for input validation
fuzz_testing: {
framework: "go test -fuzz",
target_functions: [
"Schema validation",
"API request parsing",
"State value conversion",
"Attribute validators"
],
corpus_directory: "testdata/fuzz",
run_command: "go test -fuzz=FuzzValidation -fuzztime=30s"
}

Benchmark Testing Strategy

// Performance benchmarking for provider operations
benchmark_strategy: {
create_operations: "Benchmark resource creation with varying sizes",
read_operations: "Test state refresh performance",
list_operations: "Benchmark data source queries",
update_operations: "Measure update operation timing",
run_command: "go test -bench=. -benchmem ./..."
}

🎯 TERRAFORM PROVIDER TDD BEST PRACTICES

Provider Test Design Principles

// FIRST principles adapted for Terraform providers
provider_test_principles: {
Fast: "Unit tests run in milliseconds, acceptance tests < 2 hours",
Independent: "Each resource test is isolated with unique names",
Repeatable: "Tests work with any API endpoint (mock or real)",
Self_Validating: "Clear pass/fail with TestCheckResourceAttr",
Timely: "Acceptance tests written before CRUD implementation"
}

// Arrange-Act-Assert for provider tests
provider_test_structure: {
Arrange: "Define Terraform config with resource block",
Act: "Apply configuration through acceptance test framework",
Assert: "Verify resource attributes in state using TestCheck functions"
}

Common Provider TDD Anti-Patterns to Avoid

// Avoid these provider development mistakes
provider_antipatterns: {
"Skipping ImportState Tests": "Always test resource import functionality",
"Hardcoded Test Values": "Use unique resource names per test run",
"Incomplete CRUD": "Test all Create, Read, Update, Delete operations",
"Ignoring Error Cases": "Test API failures and invalid inputs",
"Missing Documentation": "Keep examples/ and docs/ in sync with code",
"Not Testing State Drift": "Verify Read correctly detects external changes",
"Brittle Acceptance Tests": "Don't depend on external state or ordering"
}

📚 Related Terraform Provider Resources

HashiCorp Provider Design Principles - Official design guidelines
Terraform Plugin Framework Docs - Framework API reference  
Provider Development Program - Partnership and verification
Terraform Registry - Publishing and distribution

### Common Anti-Patterns to Avoid

❌ Skipping ImportState tests
❌ Skipping Drift tests
❌ Using hardcoded test values
❌ Incomplete CRUD testing
❌ Ignoring error cases
❌ Missing or outdated documentation
❌ Not testing state drift
❌ Tests dependent on external state

Test infrastucture is available for you to automatically execute the tests.

### Recommendations

Testing a provider-defined function should ensure at least the following behaviors are covered:

Known values return the expected results.
For any list, map, object, and set parameters, null values for collection elements or object attributes. The AllowNullValue parameter setting does not affect Terraform sending these types of null values.
If any parameters enable AllowNullValue, null values for those arguments.
If any parameters enable AllowUnknownValues, unknown values for those arguments.
Any errors, such as argument validation errors.

## 🧪 TERRAFORM EXAMPLE TESTING INFRASTRUCTURE

### Overview

This provider includes automated testing infrastructure for all Terraform documentation examples in `/workspace/examples/`. The test script validates that all examples are syntactically correct, properly configured, and executable.

### Test Script Location

`/workspace/scripts/test-examples.sh` - Main test execution script with parallel/sequential execution, cleanup, and comprehensive reporting.

### Architecture

**Execution Strategy:**
- **Data Sources**: Parallel execution (parallel limit: 4 concurrent tests)
  - Read-only operations are safe to run concurrently
  - Significantly faster test execution (10s for 10 data sources)

- **Resources**: Sequential execution
  - State-modifying operations run one at a time
  - Prevents race conditions and resource conflicts
  - Ensures reliable cleanup (19s for 7 resources)

**Phases:**
1. **Phase 1: Environment Validation** - Verify BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD are set
2. **Phase 2: Provider Build** - Build latest terraform-provider-bcm binary (optional, set SKIP_BUILD=true to skip)
3. **Phase 3: Example Testing** - Execute terraform init → validate → plan for each example
4. **Phase 4: Cleanup** - Remove test resources via BCM API with exponential backoff retry

### Provider Environment Variable Support

**CRITICAL PATTERN**: The BCM provider reads authentication from environment variables, following Terraform best practices:

```hcl
# Provider schema (provider.go lines 49-61)
"endpoint": schema.StringAttribute{
    MarkdownDescription: "BCM JSON-RPC API endpoint. Can also be set via BCM_ENDPOINT environment variable.",
    Optional:            true,  # NOT Required - allows env var fallback
}
```

```go
// Provider configuration reads from env vars (provider.go lines 83-97)
endpoint := data.Endpoint.ValueString()
if endpoint == "" {
    endpoint = os.Getenv("BCM_ENDPOINT")  # Fallback to environment
}
```

**Documentation Example Pattern:**
```hcl
# Authentication via environment variables (recommended)
# export BCM_ENDPOINT="https://bcm.example.com:8081"
# export BCM_USERNAME="admin"
# export BCM_PASSWORD="your-password"
provider "bcm" {
  insecure_skip_verify = true  # Only for self-signed certificates
}
```

**Why This Matters:**
- ✅ Simplifies documentation examples (no variable declarations)
- ✅ Follows Terraform provider best practices
- ✅ Enables consistent testing without variable injection
- ✅ Prevents variable declaration conflicts
- ✅ Aligns with HashiCorp's AWS/Azure/GCP provider patterns

### Running Tests

**Full Test Suite:**
```bash
BCM_ENDPOINT="https://172.21.15.254:8081" \
BCM_USERNAME="root" \
BCM_PASSWORD="Hashicorp123!" \
SKIP_BUILD=true \
./scripts/test-examples.sh
```

**Data Sources Only:**
```bash
BCM_ENDPOINT="..." BCM_USERNAME="..." BCM_PASSWORD="..." \
./scripts/test-examples.sh --data-sources-only
```

**Resources Only:**
```bash
BCM_ENDPOINT="..." BCM_USERNAME="..." BCM_PASSWORD="..." \
./scripts/test-examples.sh --resources-only
```

**Verbose Mode (show full terraform output):**
```bash
BCM_ENDPOINT="..." BCM_USERNAME="..." BCM_PASSWORD="..." \
./scripts/test-examples.sh --verbose
```

**Skip Cleanup (debug failed tests):**
```bash
BCM_ENDPOINT="..." BCM_USERNAME="..." BCM_PASSWORD="..." \
./scripts/test-examples.sh --no-cleanup
```

### Example Discovery

The script automatically discovers all `.tf` files in:
- `/workspace/examples/data-sources/*/` - Data source examples
- `/workspace/examples/resources/*/` - Resource examples

**Discovery Pattern:**
- Finds individual `.tf` files (not directories)
- Tests each file in isolation to prevent conflicts
- Depth: 2 levels (category/example.tf)

**Why Per-File Testing:**
```
examples/data-sources/bcm_cmdevice_nodes/
├── data-source.tf          # Generic example
├── filter_by_type.tf       # Filtering pattern
├── filter_by_hostname.tf   # Alternative filter
└── dynamic_inventory.tf    # Advanced use case
```

Testing entire directories would cause:
- Duplicate resource name conflicts (multiple files defining `data "bcm_cmdevice_nodes" "all"`)
- Terraform would try to merge all `.tf` files in the directory

### Provider Injection

For examples that don't include provider blocks (most data source examples):

```hcl
# Auto-injected by test script (_provider.tf)
terraform {
  required_version = ">= 1.5.0"
  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

provider "bcm" {
  insecure_skip_verify = true
}
```

**Key Insight**: No hardcoded credentials! Provider reads from BCM_ENDPOINT/BCM_USERNAME/BCM_PASSWORD environment variables.

### Test Validation Stages

Each example goes through:

1. **terraform init -backend=false**
   - Initialize provider plugins
   - No backend configuration (examples are standalone)

2. **terraform validate**
   - Syntax validation
   - Schema validation
   - Reference validation

3. **terraform plan**
   - Provider authentication (reads from env vars)
   - API connectivity test
   - Resource/data source query execution
   - **ACTUAL EXECUTION** - not mocked!

### Cleanup Strategy

**Resource Naming Pattern:** All test resources include "citest" in the name for easy identification and cleanup.

**Cleanup Implementation:**
- Direct BCM API calls (not terraform destroy)
- Exponential backoff retry (1s, 2s, 4s, 8s = 15s total)
- Handles eventual consistency
- Parallel resource deletion where safe

**Cleanup Logic:**
```bash
# Query test resources
curl -k -b cookies.txt -X POST "$BCM_ENDPOINT/json" \
  -d '{"service":"CMPart","call":"getSoftwareImages"}'

# Filter by "citest" prefix
test_images=$(echo "$response" | jq -r '.[] | select(.name | startswith("citest")) | .name')

# Delete with retry
curl -k -b cookies.txt -X POST "$BCM_ENDPOINT/json" \
  -d '{"service":"CMPart","call":"removeSoftwareImages","args":[["citest-image1","citest-image2"],false]}'
```

### Expected Test Results

**Current Status (as of 2025-11-21):**
```
Examples tested: 17
  Passed: 17
  Failed: 0

Data sources: 10 examples (10s parallel execution)
Resources: 7 examples (19s sequential execution)
Total time: 33s
```

**Data Source Examples:**
- ✅ bcm_cmdevice_nodes (4 examples: data-source.tf, filter_by_type.tf, filter_by_hostname.tf, dynamic_inventory.tf)
- ✅ bcm_cmdevice_categories (3 examples: data-source.tf, data-source-filter.tf, data-source-workflow.tf)
- ✅ bcm_cmpart_softwareimages (1 example: data-source.tf)
- ✅ bcm_cmnet_networks (2 examples: data-source.tf, filtered.tf)

**Resource Examples:**
- ✅ bcm_cmpart_softwareimage (5 examples: resource.tf, resource-advanced.tf, edge-case-two-step-create.tf, edge-case-empty-modules.tf, edge-case-path-revision.tf)
- ✅ bcm_cmdevice_category (1 example: resource.tf)
- ✅ test-citest (1 example: resource.tf - CI testing validation)

### Troubleshooting

**Tests failing with "Missing BCM Endpoint" error:**
```bash
# Verify environment variables are set
echo $BCM_ENDPOINT
echo $BCM_USERNAME
echo $BCM_PASSWORD  # Will show *** if set

# Export if not set
export BCM_ENDPOINT="https://172.21.15.254:8081"
export BCM_USERNAME="root"
export BCM_PASSWORD="Hashicorp123!"
```

**Tests failing with "The provider hashicorp/bcm does not support data source X":**
```bash
# Rebuild provider to include latest data sources/resources
export GOMODCACHE=/workspace/.go/pkg/mod
export GOCACHE=/workspace/.go/cache
export GOPATH=/workspace/.go

go build -o terraform-provider-bcm_v0.1.0 .

# Install for both architectures
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bcm/0.1.0/linux_amd64
mkdir -p ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bcm/0.1.0/linux_arm64
cp terraform-provider-bcm_v0.1.0 ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bcm/0.1.0/linux_amd64/
cp terraform-provider-bcm_v0.1.0 ~/.terraform.d/plugins/registry.terraform.io/hashicorp/bcm/0.1.0/linux_arm64/
```

**Tests showing "SKIP" status:**
- This is EXPECTED for examples that require actual BCM API connectivity
- Script detects connectivity errors and marks as skipped (not failed)
- Tests still validate syntax, schema, and provider configuration

**Platform architecture mismatch:**
```bash
# Check which architecture terraform expects
terraform version

# Install provider for correct architecture
# linux_amd64 for x86_64
# linux_arm64 for arm64/aarch64
```

### Best Practices for Example Development

1. **Provider Configuration:**
   - Use environment variable pattern (no hardcoded credentials)
   - Include comment showing export commands
   - Only set `insecure_skip_verify = true` when needed

2. **Resource Naming:**
   - Include descriptive names
   - Use "citest" prefix for CI testing examples
   - Generate unique names with timestamps for parallel safety

3. **Documentation Comments:**
   - Explain use case at top of file
   - Document edge cases and BCM API quirks
   - Show production patterns vs. simple examples

4. **Testing Before Commit:**
   ```bash
   # Always test new examples before committing
   BCM_ENDPOINT="..." BCM_USERNAME="..." BCM_PASSWORD="..." \
   ./scripts/test-examples.sh
   ```

### CI/CD Integration

The test script is designed for CI/CD pipelines:

```yaml
# Example GitHub Actions workflow
- name: Test Terraform Examples
  env:
    BCM_ENDPOINT: ${{ secrets.BCM_ENDPOINT }}
    BCM_USERNAME: ${{ secrets.BCM_USERNAME }}
    BCM_PASSWORD: ${{ secrets.BCM_PASSWORD }}
    SKIP_BUILD: "false"  # Build provider in CI
  run: |
    ./scripts/test-examples.sh
```

**Exit Codes:**
- `0` - All tests passed
- `1` - One or more tests failed
- `2` - Environment validation failed
- `3` - Provider build failed

### Maintenance

When adding new resources/data sources:

1. Add implementation in `internal/provider/`
2. Register in `provider.go` DataSources() or Resources()
3. Create documentation examples in `examples/`
4. Run test script to validate
5. Commit passing examples

**Example Discovery is Automatic** - No need to update test script configuration!
