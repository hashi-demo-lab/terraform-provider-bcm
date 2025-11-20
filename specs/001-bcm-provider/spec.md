# Feature Specification: Nvidia BCM Terraform Provider

**Feature Branch**: `001-bcm-provider`
**Created**: 2025-11-20
**Status**: Comprehensive Specification
**Input**: User description: "Create a comprehensive feature specification for a new Terraform provider for Nvidia BCM (BlueField Configuration Manager)"

## Clarifications

### Session 2025-11-20

- Q: The spec describes a GET-based search endpoint (/api/search?type=service&name=CMPart), but you originally mentioned a JSON-RPC style endpoint (/json) with POST body {"service":"CMPart","call":"getSoftwareImages"}. Which API pattern should the provider implement? → A: POST /json with JSON-RPC body (as originally mentioned)
- Q: How does the BCM JSON-RPC API authenticate requests to POST /json endpoint? → A: HTTP Basic Auth (username:password on every request)
- Q: The endpoint uses HTTPS with a private IP (172.21.15.254). What type of TLS certificate does the BCM server use? → A: Self-signed certificate
- Q: The SoftwareImage API response has 30+ fields including nested modules array. Which fields should the bcm_cmpart_softwareimages data source expose in its Terraform schema? → A: All fields including nested modules array
- Q: Can the getSoftwareImages API call accept parameters to filter results (e.g., by name or uuid), or does it always return all software images? → A: VERIFIED - getSoftwareImages returns all software images (no filtering parameters in request body, POV will return complete array)
- Q: What does the BCM JSON-RPC API return when an error occurs (e.g., invalid service, authentication failure, internal error)? → A: Unknown - need to test error scenarios
- Q: Authentication pattern correction - The BCM API does NOT use HTTP Basic Auth. What is the actual authentication mechanism? → A: Token-based cookie authentication: POST login request to /json with {"service":"login","username":"root","password":"Hashicorp123!"} returns token via Set-Cookie header. This token is then automatically included in subsequent requests via the Cookie header. The token never expires, so login once and reuse the HTTP client with cookie jar for all operations.

#### Verified API Test Results (2025-11-20)

**VERIFIED: Login API Call**
- Request: `POST https://172.21.15.254:8081/json` with body `{"service":"login","username":"root","password":"Hashicorp123!"}`
- Response: HTTP 200 OK with Content-Type: application/json and body `true` (JSON boolean, not string "true")
- Set-Cookie header: `cm-login-token=63rgukdSV8ZqeON6nGy7TCWDZcySl2sG;path=/;Secure;httponly`
- Cookie attributes: Secure, httponly, path=/
- Cookie name: `cm-login-token`
- CONFIRMED: Login returns boolean `true` on success (not JSON object)
- CONFIRMED: Token-based cookie authentication is correct (not HTTP Basic Auth)
- CONFIRMED: Cookie is automatically set via Set-Cookie header
- CONFIRMED: Cookie jar pattern will work (no manual cookie manipulation needed)

**VERIFIED: getSoftwareImages API Call**
- Request: `POST https://172.21.15.254:8081/json` with body `{"service":"CMPart","call":"getSoftwareImages"}`
- Required Cookie header: `cm-login-token=<token-from-login>`
- Response: HTTP 200 OK with 11KB JSON array of SoftwareImage objects
- Contains complete data including nested modules array (50+ kernel modules per image)
- Response structure matches the example payload provided earlier
- CONFIRMED: getSoftwareImages returns array of objects with full schema
- CONFIRMED: Nested modules array is present in response
- CONFIRMED: No filtering parameters needed - returns all images by default

**VERIFIED: cmdevice.getNode API Call (OUT OF SCOPE - Post-POV)**
- Request: `POST https://172.21.15.254:8081/json` with body `{"service":"cmdevice","call":"getNode","arg":"master"}`
- Required Cookie header: `cm-login-token=<token-from-login>` (same token as above)
- Purpose: Retrieve information about cluster nodes (e.g., master node)
- CONFIRMED: Same authentication token works across multiple BCM services (token is API-wide, not service-specific)
- CONFIRMED: New JSON-RPC pattern discovered: `"arg"` parameter for passing arguments to service calls
- STATUS: OUT OF SCOPE for POV - documented for future expansion only

**Key Implementation Confirmations:**
1. Login response format: Returns `true` (boolean), not JSON object
2. Cookie name: `cm-login-token` (use for debugging/logging)
3. Cookie attributes: Secure, httponly, path=/ (standard secure cookie)
4. Response size: ~11KB per image (plan for reasonable response sizes, no pagination needed for POV)
5. Modules array: Confirmed present with 50+ modules per image (test nested schema thoroughly)
6. **Token scope**: Authentication token is API-wide - works across all BCM services (CMPart, cmdevice, etc.)
7. **Parameter passing pattern**: API calls may include `"arg"` field for method arguments (observed in cmdevice.getNode)

## Provider Overview

### What is Nvidia BCM?

Nvidia Base Command Manager (BCM) is a comprehensive cluster management platform designed for HPC (High-Performance Computing) and AI infrastructure. BCM provides centralized management capabilities for:

- Host nodes and compute clusters
- Network topology and configuration
- Power management and firmware updates
- Workload scheduling and monitoring
- Resource provisioning and categorization

BCM exposes its functionality through multiple APIs:

- **JSON-RPC API** (port 8081): JSON-RPC service invocation endpoint at `/json`
- **Python API**: pythoncm package for programmatic cluster management
- **REST API** (historical): Legacy REST endpoints (not used in this provider)

This provider focuses exclusively on the **BCM JSON-RPC API** at `https://172.21.15.254:8081/json` using POST requests with JSON body containing `service` and `call` parameters.

**Authentication Pattern**: Token-based cookie authentication:
1. Initial login: POST to `/json` with `{"service":"login","username":"<user>","password":"<pass>"}`
2. BCM returns authentication token via `Set-Cookie` header
3. Subsequent API calls: Use HTTP client with cookie jar that automatically includes the token via `Cookie` header
4. Token lifetime: Never expires - login once during provider configuration and reuse the client for all operations

### Why a Terraform Provider?

Infrastructure teams managing BCM clusters need declarative, version-controlled infrastructure-as-code capabilities. A Terraform provider for BCM enables:

1. **Infrastructure as Code**: Define BCM cluster configurations in version-controlled Terraform files
2. **Automated Provisioning**: Programmatically provision and configure DPUs, networks, and resources
3. **State Management**: Track infrastructure state and detect configuration drift
4. **Multi-Cloud Integration**: Integrate BCM with other Terraform-managed infrastructure (AWS, Azure, GCP)
5. **CI/CD Integration**: Automate BCM infrastructure changes through GitOps workflows
6. **Consistency**: Apply consistent configurations across multiple BCM clusters
7. **Team Collaboration**: Review infrastructure changes through standard PR workflows

### API Suitability Assessment

**Analysis Date**: 2025-11-20
**Status**: VERIFIED with Live API Testing

After testing the BCM JSON-RPC API endpoint at `https://172.21.15.254:8081/json` with actual login and getSoftwareImages calls, the following assessment was confirmed:

**✅ BCM JSON-RPC API Endpoint Strengths:**

- **Service-Call Pattern**: JSON-RPC style invocation with `service` and `call` parameters in POST body
- **Structured Requests**: POST requests with JSON body containing service name and method call
- **JSON Response Format**: Structured JSON responses suitable for Terraform schema mapping
- **HTTP-based**: Standard HTTPS protocol with token-based cookie authentication
- **Service-Oriented Architecture**: Different services (CMPart, etc.) expose different method calls
- **Persistent Authentication**: Token-based session cookies that never expire (login once, reuse client)

**VERIFIED: JSON-RPC Authentication Flow**

```
# Step 1: Login (performed once during provider Configure())
POST https://172.21.15.254:8081/json
Content-Type: application/json

{"service":"login","username":"root","password":"Hashicorp123!"}

Response: HTTP 200 OK
Response Body: true
Response Headers:
Set-Cookie: cm-login-token=63rgukdSV8ZqeON6nGy7TCWDZcySl2sG;path=/;Secure;httponly

# Step 2: Subsequent API calls (automatic cookie inclusion via http.Client with cookie jar)
POST https://172.21.15.254:8081/json
Content-Type: application/json
Cookie: cm-login-token=63rgukdSV8ZqeON6nGy7TCWDZcySl2sG (automatically added by cookie jar)

{"service":"CMPart","call":"getSoftwareImages"}

Response: HTTP 200 OK
Response Body: [<array of SoftwareImage objects>] (~11KB per image)
```

**✅ JSON-RPC API Design for Terraform Provider:**

1. **Service-Call Foundation**: All data source queries use the JSON-RPC endpoint with service/call pattern
2. **Service-based Data Sources**: Each BCM service/call maps to a specific Terraform data source type
3. **Token-Based Authentication**: Login once during Configure() to obtain session cookie, reuse HTTP client with cookie jar for all subsequent operations
4. **Persistent Sessions**: Token never expires - no session refresh or re-authentication logic needed
5. **Extensible Design**: New services and calls can be added as additional data sources without provider architecture changes

**POV (Proof of Value) Scope:**

This initial POV implementation validates the JSON-RPC API integration pattern with minimal scope:

**POV Deliverables**:

- **Provider Authentication**: Token-based cookie authentication with login API call during Configure(), self-signed certificate support (insecure_skip_verify option)
- **Single Data Source**: `bcm_cmpart_softwareimages` - queries CMPart service getSoftwareImages call
- **Complete Schema**: Full SoftwareImage schema including nested modules array (30+ fields)
- **Acceptance Tests**: TDD-based acceptance test coverage for authentication and data source
- **Documentation**: Working examples and schema documentation

**POV Goals**:

1. Validate JSON-RPC API integration pattern (POST /json with {"service":"...","call":"..."} body)
2. Prove token-based cookie authentication works correctly (login API call + cookie jar pattern)
3. Validate self-signed certificate handling with insecure_skip_verify option
4. Demonstrate complex nested schema handling (modules array within SoftwareImage objects)
5. Establish TDD workflow for future data source additions
6. Create foundation for production provider expansion

**Future Expansion Beyond POV**:
After POV validation, the provider can be extended with:

- Additional CMPart data sources (other service calls)
- Data sources for other BCM services beyond CMPart
- Session error handling (if token invalidation scenarios are discovered during testing)
- Error handling refinements based on observed API error patterns
- Write operations (resources) if BCM JSON-RPC API supports state-changing calls

**Value Proposition**:
The BCM JSON-RPC provider enables infrastructure-as-code for BCM cluster management:

- **Declarative Configuration**: Define BCM queries in version-controlled Terraform files
- **State Visibility**: Query BCM software images and cluster state through Terraform
- **Multi-Cloud Coordination**: Use BCM data sources to inform cloud resource provisioning
- **GitOps Workflows**: Review infrastructure queries through standard PR processes
- **Extensible Architecture**: JSON-RPC pattern supports easy addition of new service calls

### Target Environment

- **BCM Instance**: http://172.21.15.254/
- **JSON-RPC API Endpoint**: https://172.21.15.254:8081/json (POST only)
- **Authentication**: Token-based cookie authentication
  - Login request: `{"service":"login","username":"root","password":"Hashicorp123!"}`
  - Token received via `Set-Cookie` header
  - Subsequent requests: Automatic cookie inclusion via http.Client cookie jar
  - Token lifetime: Never expires (no re-authentication needed)
- **TLS Certificate**: Self-signed (requires insecure_skip_verify option)
- **Test Credentials**: root:Hashicorp123!
- **API Version**: BCM 11 JSON-RPC API
- **Example Data Source Request Body**: `{"service":"CMPart","call":"getSoftwareImages"}`
- **Request Headers**:
  - `Content-Type: application/json`
  - `Cookie: <auth-token>` (automatically added by cookie jar after login)

## User Scenarios & Testing

### POV User Stories

### User Story 1 - Provider Authentication with Self-Signed Certificate (POV - Priority: P0)

As a platform engineer, I need to configure the BCM Terraform provider with token-based authentication credentials and handle the self-signed certificate so that I can execute JSON-RPC API calls.

**Why POV Critical**: Authentication is the foundational requirement. Without API connectivity to the JSON-RPC endpoint, the data source cannot function.

**API Endpoint**: `https://172.21.15.254:8081/json` (POST)

**Authentication Flow**:
1. Provider Configure() method performs login: POST `{"service":"login","username":"root","password":"Hashicorp123!"}`
2. BCM API returns authentication token via `Set-Cookie` header
3. Provider stores HTTP client with cookie jar containing the session token
4. All subsequent data source operations reuse this authenticated client (cookies automatically included)
5. No re-authentication needed (token never expires)

**Independent Test**: Configure provider block with credentials (root:Hashicorp123!) and `insecure_skip_verify = true`, execute a simple JSON-RPC call (getSoftwareImages) to verify authentication and connectivity.

**Acceptance Scenarios**:

1. **Given** valid BCM credentials (root:Hashicorp123!) and `insecure_skip_verify = true`, **When** provider is configured, **Then** provider successfully performs login API call, receives Set-Cookie response, and stores authenticated HTTP client with cookie jar
2. **Given** provider has authenticated HTTP client, **When** data source executes API call, **Then** provider automatically includes Cookie header with authentication token (no manual header manipulation needed)
3. **Given** invalid BCM credentials, **When** provider Configure() attempts login, **Then** provider returns clear authentication error with login failure message
4. **Given** unreachable BCM endpoint, **When** provider initialization is attempted, **Then** provider returns connection timeout error
5. **Given** self-signed certificate on BCM endpoint and `insecure_skip_verify = false` (or unset), **When** provider attempts connection, **Then** provider returns TLS certificate verification error
6. **Given** successful login, **When** multiple data source operations execute, **Then** provider reuses same authenticated HTTP client without re-authenticating (token persists)

**POV Implementation**: Token-based cookie authentication with login API call during Configure(), HTTP client with cookie jar for automatic cookie management, insecure_skip_verify option to bypass self-signed certificate validation. No session expiration handling needed (token never expires).

---

### User Story 2 - CMPart SoftwareImages Data Source (POV - Priority: P0)

As a platform engineer, I need to query BCM software images through a Terraform data source so that I can discover available OS images and their kernel module configurations.

**Why POV Critical**: This validates the complete JSON-RPC integration pattern, proving we can call BCM services, parse complex nested JSON responses, and expose them through Terraform schemas.

**API Endpoint**: `https://172.21.15.254:8081/json` (POST)
**Request Body**: `{"service":"CMPart","call":"getSoftwareImages"}`

**VERIFIED Response Schema** (confirmed via live API testing):

- HTTP 200 OK response with JSON array of SoftwareImage objects (~11KB per image)
- Each SoftwareImage object contains:
  - Top-level fields: uuid, name, path, kernelVersion, kernelParameters, SOL\* fields, revision fields, timestamps, flags
  - Nested modules array: Each module has uuid, name, parameters, baseType, childType, and metadata fields (50+ modules per image confirmed)

**Independent Test**: Define `data "bcm_cmpart_softwareimages" "all"` data source, run terraform plan/apply, verify all software images returned with complete schema including nested modules array.

**Acceptance Scenarios**:

1. **Given** software images exist in BCM, **When** I query `data "bcm_cmpart_softwareimages"`, **Then** Terraform returns array of SoftwareImage objects with all 30+ fields correctly mapped to Terraform schema
2. **Given** SoftwareImage objects contain nested modules arrays, **When** data source reads response, **Then** Terraform correctly parses modules as nested list attribute with KernelModule schema
3. **Given** API returns successful response with empty array, **When** I query the data source, **Then** Terraform returns empty images list (not an error)
4. **Given** API call fails with authentication error, **When** I query the data source, **Then** Terraform returns clear error message with details
5. **Given** API returns unexpected JSON structure, **When** data source parses response, **Then** Terraform logs full response at TRACE level and returns parse error with details

**POV Implementation**: Read-only data source for getSoftwareImages call. No filtering parameters in POV scope (future: may add name/uuid filters if API supports params). Complete schema mapping for all SoftwareImage fields plus nested modules array.

---

### Out of Scope (POV)

The following capabilities are **explicitly out of scope** for the POV:

- **Additional Data Sources**: Only `bcm_cmpart_softwareimages` implemented; other CMPart calls (e.g., getDevices, getHosts) deferred to post-POV
- **Other BCM Services**: Additional BCM service endpoints beyond CMPart (e.g., cmdevice.getNode) deferred to post-POV
- **API Argument Passing**: JSON-RPC calls with `"arg"` parameter pattern (observed in cmdevice.getNode) - POV only implements calls without arguments
- **Write Operations (Resources)**: No CREATE/UPDATE/DELETE operations in POV; read-only data source validation only
- **Filtering Parameters**: Data source returns all software images; name/uuid filtering deferred until API parameter support confirmed
- **Session Expiration Handling**: Token never expires per confirmed API behavior; no session refresh logic in POV scope
- **Advanced Error Handling**: POV implements defensive error handling checking both HTTP status and response body; refinements based on observed API error patterns deferred
- **Pagination**: API pagination support unknown; POV assumes single response contains all results
- **REST API Endpoints**: Only JSON-RPC API (`/json`) in scope; legacy REST endpoints (`/rest/v1/*`) out of scope
- **Configuration Management**: BCM cluster configuration managed through UI/CLI/Python API (not provider scope)

### Edge Cases

**Authentication & TLS:**

- What happens when login API call fails with invalid credentials? Provider Configure() should fail fast with clear "authentication failed" error message including login endpoint and credentials validation guidance
- What happens when `insecure_skip_verify = false` (or unset) with self-signed cert? Provider should return TLS certificate verification error with guidance to set `insecure_skip_verify = true` for POV
- What happens when credentials contain special characters in password? Provider must properly handle special characters in JSON request body (standard JSON escaping)
- What happens when cookie is not returned after successful login? Provider should detect missing Set-Cookie header and return error indicating authentication flow failure

**API Response Handling:**

- What happens when getSoftwareImages returns empty array? Data source should return empty images list (not an error) with DEBUG logging indicating zero images found
- What happens when BCM returns unexpected JSON structure (unknown fields or error format)? Provider should handle gracefully, log raw response at TRACE level, return parse error with response snippet in error message
- VERIFIED: What does successful login API call return? Returns boolean `true` in response body with HTTP 200 OK and Set-Cookie header containing `cm-login-token`
- VERIFIED: What does successful getSoftwareImages call return? Returns JSON array of SoftwareImage objects (~11KB per image) with HTTP 200 OK, includes nested modules array with 50+ modules per image
- What happens when API returns 200 OK with JSON error object instead of array? POV implements defensive check: if response is object with "error" field, treat as error; if array, treat as success (error format still needs testing with actual API errors)
- What happens when SoftwareImage object missing expected fields (e.g., no uuid or name)? Provider should mark those fields as null/empty in Terraform state, not fail entire parse
- VERIFIED: What happens when modules array is present? Confirmed modules array exists with 50+ kernel modules per image - provider must handle large nested arrays correctly

**Network & Performance:**

- What happens when network timeout occurs during JSON-RPC call? Provider should timeout after 30 seconds (configurable), return timeout error with endpoint and troubleshooting guidance
- What happens when BCM JSON-RPC endpoint is unavailable (connection refused)? Provider should return connection error with endpoint URL and suggestion to verify BCM service status
- What happens when API returns very large response (100+ images with 1000+ modules each)? POV accepts full response; pagination deferred to post-POV if performance issues observed

**Terraform State Management:**

- What happens on subsequent `terraform refresh` calls? Provider re-queries getSoftwareImages on every Read operation, updates state with latest data (no caching in POV)
- What happens when software image is deleted between terraform apply runs? Next refresh will show removed image no longer in state (expected behavior for data sources)

## Requirements

### Functional Requirements

#### Provider Configuration (POV)

- **FR-001**: Provider MUST support token-based authentication with username and password attributes for login API call
- **FR-002**: Provider MUST support configurable BCM JSON-RPC API endpoint URL (default: `https://172.21.15.254:8081/json`)
- **FR-003**: Provider MUST support `insecure_skip_verify` boolean flag to bypass TLS verification for self-signed certificates (required for POV environment)
- **FR-004** (VERIFIED): Provider Configure() method MUST perform login API call: POST `{"service":"login","username":"<user>","password":"<pass>"}` to `/json` endpoint. Response Content-Type MUST be application/json. Response body MUST be parseable as JSON boolean `true` (not string "true"). Set-Cookie header MUST contain `cm-login-token`
- **FR-005**: Provider MUST create HTTP client with cookie jar (http.CookieJar) to automatically manage session cookies
- **FR-006**: Provider MUST store authenticated HTTP client after successful login and reuse for all data source operations
- **FR-007** (VERIFIED): Provider MUST verify Set-Cookie header is present in login response with cookie name `cm-login-token`; if missing, return authentication error
- **FR-008**: Provider MUST include `Content-Type: application/json` header on all JSON-RPC POST requests
- **FR-009**: Provider MUST support configurable HTTP timeout (default: 30 seconds)
- **FR-010**: Provider authentication MUST execute during Configure() method (eager validation - fail fast on invalid credentials)

#### Data Source: bcm_cmpart_softwareimages (POV)

**Read Operation:**

- **FR-011**: Data source MUST implement Read operation using terraform-plugin-framework DataSource interface
- **FR-012** (VERIFIED): Data source MUST use authenticated HTTP client from provider configuration (with cookie jar containing `cm-login-token` session cookie)
- **FR-013**: Data source MUST construct JSON-RPC request body: `{"service":"CMPart","call":"getSoftwareImages"}`
- **FR-014**: Data source MUST POST request to configured JSON-RPC endpoint URL
- **FR-015** (VERIFIED): Data source MUST rely on cookie jar to automatically include Cookie header with `cm-login-token` authentication token (no manual header manipulation)
- **FR-016**: Data source MUST return empty images list (not error) when API returns empty array `[]`
- **FR-017**: Data source MUST re-query API on every terraform refresh (no client-side caching)

**Schema Definition:**

- **FR-018** (VERIFIED): Data source MUST expose computed attribute `images` (list of SoftwareImage objects) - confirmed API returns JSON array of objects
- **FR-019** (VERIFIED): SoftwareImage schema MUST include all top-level fields from API response (confirmed via live API testing):

  - `id` (computed) - Mapped from `uuid` field
  - `uuid` (string, computed)
  - `name` (string, computed)
  - `path` (string, computed)
  - `kernel_version` (string, computed) - Mapped from `kernelVersion`
  - `kernel_parameters` (string, computed) - Mapped from `kernelParameters`
  - `kernel_output_console` (string, computed) - Mapped from `kernelOutputConsole`
  - `bootfs_part` (string, computed) - Mapped from `bootfspart`
  - `fs_part` (string, computed) - Mapped from `fspart`
  - `enable_sol` (bool, computed) - Mapped from `enableSOL`
  - `sol_port` (string, computed) - Mapped from `SOLPort`
  - `sol_speed` (string, computed) - Mapped from `SOLSpeed`
  - `sol_flow_control` (bool, computed) - Mapped from `SOLFlowControl`
  - `base_type` (string, computed) - Mapped from `baseType`
  - `child_type` (string, computed) - Mapped from `childType`
  - `creation_time` (number, computed) - Mapped from `creationTime`
  - `file_operation_in_progress` (bool, computed) - Mapped from `fileOperationInProgress`
  - `modified` (bool, computed)
  - `notes` (string, computed)
  - `original_image` (string, computed) - Mapped from `originalImage`
  - `parent_software_image` (string, computed) - Mapped from `parentSoftwareImage`
  - `parent_uuid` (string, computed) - Mapped from `parent_uuid`
  - `revision` (string, computed)
  - `revision_id` (number, computed) - Mapped from `revisionID`
  - `to_be_removed` (bool, computed) - Mapped from `to_be_removed`
  - `modules` (list of KernelModule objects, computed) - Nested list attribute (VERIFIED: 50+ modules per image)

- **FR-020** (VERIFIED): KernelModule nested schema MUST include fields from modules array elements (confirmed via live API testing with 50+ modules per image):

  - `uuid` (string, computed)
  - `name` (string, computed)
  - `parameters` (string, computed)
  - `base_type` (string, computed) - Mapped from `baseType`
  - `child_type` (string, computed) - Mapped from `childType`
  - `revision` (string, computed)
  - `modified` (bool, computed)
  - `to_be_removed` (bool, computed) - Mapped from `to_be_removed`

- **FR-021**: Schema MUST use snake_case for Terraform attribute names (convert API camelCase fields)
- **FR-022**: Schema MUST mark all attributes as `Computed: true` (data source is read-only)
- **FR-023**: Schema MUST handle missing/null fields gracefully (use types.StringNull(), types.BoolNull(), etc.)
- **FR-024**: Modules attribute MUST be typed as `ListNestedAttribute` with KernelModule object type

#### Error Handling and Validation (POV)

- **FR-025**: Provider MUST return actionable error messages including:
  - HTTP status code (if available)
  - JSON-RPC endpoint URL
  - Request body sent (service/call values, excluding password in login requests)
  - Response body snippet (first 500 chars) for unexpected responses
- **FR-026**: Provider MUST log all JSON-RPC requests at TRACE level with full request body (excluding password field in login requests)
- **FR-027**: Provider MUST log all JSON-RPC responses at TRACE level with full response body
- **FR-028**: Provider MUST handle login authentication failures with clear error message indicating invalid credentials
- **FR-029**: Provider MUST validate Set-Cookie header presence in login response; if missing, return authentication error
- **FR-030**: Provider MUST handle connection timeouts with error including endpoint URL and timeout duration
- **FR-031**: Provider MUST handle TLS certificate errors with clear message suggesting `insecure_skip_verify = true` for POV
- **FR-032**: Provider MUST implement defensive error parsing:
  - Check HTTP status code first (non-2xx = error)
  - Check if response body is JSON object with error field (accept any of: "error", "message", "err", "errorMessage") - treat as API error
  - Check if response body is JSON array (treat as success)
  - Any other format = parse error
- **FR-033**: Provider MUST handle JSON parse errors with error message including raw response snippet
- **FR-034**: Provider MUST handle empty array response `[]` as success with zero images (not error)
- **FR-035**: Provider MUST handle connection refused errors with guidance to verify BCM service is running

#### API Integration (POV)

- **FR-036** (VERIFIED): Provider MUST use BCM JSON-RPC endpoint at `/json` (POST only) - confirmed working with live API
- **FR-037** (VERIFIED): Provider Configure() MUST perform login API call with request body: `{"service":"login","username":"<user>","password":"<pass>"}` and expect boolean `true` response with `cm-login-token` Set-Cookie header
- **FR-038** (VERIFIED): Provider MUST create HTTP client with cookie jar (net/http/cookiejar) before login to capture `cm-login-token` Set-Cookie response
- **FR-039** (VERIFIED): Provider MUST construct JSON-RPC request body with service and call fields for data source operations (e.g., `{"service":"CMPart","call":"getSoftwareImages"}`)
- **FR-040**: Provider MUST set HTTP header `Content-Type: application/json` on all requests
- **FR-041** (VERIFIED): Provider MUST NOT manually set Cookie or Authorization headers (cookie jar handles `cm-login-token` automatically after login)
- **FR-042**: Provider MUST use HTTP POST method (not GET) for all JSON-RPC calls
- **FR-043** (VERIFIED): Provider MUST parse JSON array response and map to SoftwareImage schema (~11KB per image with 50+ nested modules)
- **FR-044**: Provider MUST handle self-signed certificates when `insecure_skip_verify = true`
- **FR-045**: Provider MUST convert API camelCase fields to snake_case Terraform attributes
- **FR-046** (VERIFIED): Provider MUST map nested modules array to ListNestedAttribute in Terraform schema (confirmed 50+ modules per image)

#### Out of Scope (POV)

The following are **explicitly out of scope** for the POV:

- **FR-047**: Retry logic for transient failures (5xx errors) - POV fails fast; retry logic deferred
- **FR-048**: Rate limiting (429) handling - unknown if API uses rate limiting; deferred
- **FR-049**: Session token expiration handling - token never expires per confirmed API behavior; no refresh logic needed
- **FR-050**: Write operations (CREATE/UPDATE/DELETE) - POV is read-only data source
- **FR-051**: Filtering parameters - POV returns all images; filtering deferred pending API testing
- **FR-052**: Pagination - POV assumes single response; pagination deferred
- **FR-053**: Additional data sources beyond bcm_cmpart_softwareimages - deferred to post-POV
- **FR-054**: JSON-RPC calls with `"arg"` parameter - POV only implements argument-less calls (getSoftwareImages); argument passing pattern (observed in cmdevice.getNode) deferred to post-POV
- **FR-055**: cmdevice service integration - confirmed compatible with same authentication token; implementation deferred to post-POV

### Key Entities (POV)

#### SoftwareImage (POV Entity)

**Data Source**: `bcm_cmpart_softwareimages`
**API Call**: POST `/json` with body `{"service":"CMPart","call":"getSoftwareImages"}`
**Description**: Represents a BCM software image (operating system image) with kernel configuration and module information.

**Key Attributes**:

- `images` (computed, list of objects): Array of all software images in BCM system
- Each SoftwareImage object contains:
  - **Identity Fields**: `id`/`uuid`, `name`, `path`
  - **Kernel Configuration**: `kernel_version`, `kernel_parameters`, `kernel_output_console`
  - **Partitions**: `bootfs_part`, `fs_part`
  - **Serial Over LAN (SOL)**: `enable_sol`, `sol_port`, `sol_speed`, `sol_flow_control`
  - **Metadata**: `base_type`, `child_type`, `creation_time`, `revision`, `revision_id`
  - **Relationships**: `parent_software_image`, `parent_uuid`, `original_image`
  - **State Flags**: `modified`, `file_operation_in_progress`, `to_be_removed`
  - **Notes**: `notes` (free-form text field)
  - **Nested Modules**: `modules` (list of KernelModule objects)

**KernelModule (Nested Entity)**:

- **Description**: Represents a Linux kernel module configured for the software image
- **Key Attributes**: `uuid`, `name`, `parameters`, `base_type`, `child_type`, `revision`, `modified`, `to_be_removed`
- **Relationship**: Many modules per SoftwareImage (one-to-many)

**No Input Attributes**: POV data source has no configurable input attributes; it returns all software images. Filtering deferred to post-POV pending API testing.

#### Out of Scope Entities (POV)

The following are out of scope for POV:

- **Other CMPart Entities**: Additional CMPart service calls (getDevices, getHosts, etc.) deferred
- **Other BCM Services**: Services beyond CMPart (e.g., CMJob, CMEvent) deferred
- **Write Operations**: No resource entities (CREATE/UPDATE/DELETE) in POV
- **Configuration Entities**: BCM cluster configuration managed through UI/CLI/Python API (not provider scope)

## Success Criteria

### POV Success Criteria

The POV validates the JSON-RPC API integration pattern with minimal scope:

**Provider Configuration and Authentication (Critical):**

- **SC-001**: Platform engineers can configure BCM provider with credentials (username/password) and `insecure_skip_verify = true` within 2 minutes
- **SC-002** (VERIFIED): Provider Configure() successfully executes login API call to `https://172.21.15.254:8081/json` and receives boolean `true` response with `cm-login-token` authentication token via Set-Cookie header
- **SC-003** (VERIFIED): Provider stores HTTP client with cookie jar containing `cm-login-token` session cookie for reuse across all data source operations
- **SC-004**: Provider handles self-signed TLS certificates correctly when `insecure_skip_verify = true`
- **SC-005**: Provider returns actionable error messages for authentication failures including login endpoint and credentials validation guidance
- **SC-006**: Provider returns clear TLS certificate error when `insecure_skip_verify = false` with self-signed cert
- **SC-007** (VERIFIED): Provider reuses authenticated HTTP client across multiple data source read operations without re-authenticating (validates token persistence - token never expires)

**bcm_cmpart_softwareimages Data Source (Critical):**

- **SC-008** (VERIFIED): Data source successfully executes getSoftwareImages JSON-RPC call using authenticated HTTP client (with `cm-login-token` cookie jar) and receives JSON array response
- **SC-009** (VERIFIED): Data source does NOT manually set Cookie or Authorization headers (relies on cookie jar automatic `cm-login-token` inclusion)
- **SC-010** (VERIFIED): Data source correctly parses JSON array response (~11KB per image) and maps all 30+ SoftwareImage fields to Terraform schema
- **SC-011** (VERIFIED): Data source correctly parses nested modules array (50+ modules per image) and exposes as ListNestedAttribute
- **SC-012**: Data source returns empty images list (not error) when API returns empty array `[]`
- **SC-013**: Data source handles missing/null fields gracefully using types.StringNull(), types.BoolNull(), etc.
- **SC-014**: Data source converts API camelCase fields to snake_case Terraform attributes correctly
- **SC-015** (VERIFIED): Data source exposes all KernelModule fields for each module in modules array (confirmed 50+ modules per image with full schema)

**Testing and Quality (Critical):**

- **SC-016**: Acceptance tests achieve 100% pass rate for provider authentication and data source against test BCM instance
- **SC-017** (VERIFIED SCENARIOS): Acceptance tests cover:
  - Successful login API call returning boolean `true` and `cm-login-token` retrieval via Set-Cookie
  - Cookie jar automatic `cm-login-token` inclusion in subsequent requests
  - Invalid credentials error handling during login
  - TLS certificate error handling
  - Empty response handling
  - Complex nested schema (modules array with 50+ modules) validation
  - Multiple data source reads using same authenticated client (no re-authentication - token persists)
- **SC-018**: Unit tests achieve minimum 80% code coverage for provider client and data source implementation
- **SC-019**: Provider documentation includes working example for provider config and data source usage

**Error Handling (Important):**

- **SC-020**: Provider logs all JSON-RPC requests/responses at TRACE level for debugging (excluding password in login requests)
- **SC-021**: Provider implements defensive error parsing (HTTP status, error object detection, array detection)
- **SC-022**: Provider returns parse errors with raw response snippet for unexpected formats
- **SC-023**: Provider handles connection timeouts with clear error message and troubleshooting guidance
- **SC-024**: Provider validates Set-Cookie header presence after login; returns clear error if missing

**Documentation and Developer Experience (Important):**

- **SC-025**: Provider documentation includes:
  - Provider configuration example with username/password and insecure_skip_verify
  - Authentication flow explanation (login API call + cookie jar pattern)
  - Data source usage example
  - Complete schema reference (all 30+ fields documented)
  - Authentication troubleshooting guide
- **SC-026**: Terraform Registry documentation auto-generated via tfplugindocs
- **SC-027**: README includes POV scope clarification and post-POV expansion roadmap

**POV Exit Criteria:**

1. VERIFIED: Provider successfully authenticates to BCM JSON-RPC endpoint with token-based cookie authentication (login API call returns boolean `true` + `cm-login-token` cookie jar)
2. VERIFIED: Provider stores and reuses authenticated HTTP client across multiple operations (no re-authentication - token persists indefinitely)
3. VERIFIED: `bcm_cmpart_softwareimages` data source operational and tested against real BCM instance (confirmed working with live API)
4. All acceptance tests pass (100% pass rate)
5. VERIFIED: Complex nested schema (modules array with 50+ modules per image) correctly implemented and validated
6. Error handling covers login failures, TLS, parsing, and connection failures
7. Documentation complete for POV scope including authentication flow explanation with verified API response formats
8. Decision made: proceed to production provider expansion or iterate on POV based on findings

**Post-POV Success Criteria (Future):**

After POV validation, production expansion would target:

**Additional Data Sources:**

- **SC-024**: Additional CMPart data sources implemented (getDevices, getHosts, etc.)
- **SC-025**: Data sources for other BCM services beyond CMPart
- **SC-035**: cmdevice service data sources (e.g., `bcm_cmdevice_node` for cluster node information)

**Advanced Features:**

- **SC-028**: Session token refresh logic (if token expiration is discovered during extended testing)
- **SC-029**: Filtering parameters (name, uuid) if API supports them
- **SC-030**: Pagination for large result sets if API supports it
- **SC-031**: Retry logic with exponential backoff for transient failures
- **SC-036**: Support for JSON-RPC calls with `"arg"` parameter pattern (required for cmdevice.getNode and similar calls)

**Production Deployment:**

- **SC-032**: Provider passes HashiCorp Partner Provider verification (if pursuing Terraform Registry)
- **SC-033**: CI/CD pipeline runs acceptance tests on every commit
- **SC-034**: Provider released to Terraform Registry with semantic versioning

## Post-POV Expansion Roadmap

This section documents discovered API capabilities that are out of scope for the POV but should be implemented in future iterations.

### Discovered BCM Services (Post-POV)

**cmdevice Service:**

The `cmdevice` BCM service provides cluster node management capabilities. Initial testing confirmed compatibility with the POV authentication pattern.

**cmdevice.getNode API Call:**

- **Endpoint**: `POST https://172.21.15.254:8081/json`
- **Authentication**: Same `cm-login-token` cookie-based authentication (confirms token is API-wide across all services)
- **Request Format**: `{"service":"cmdevice","call":"getNode","arg":"master"}`
- **Purpose**: Retrieve information about cluster nodes (e.g., master node configuration and status)
- **Status**: OUT OF SCOPE for POV - documented for future expansion

**Key Observations:**

1. **API-Wide Authentication Token**: Confirmed that `cm-login-token` works across all BCM services (CMPart, cmdevice, etc.), not just service-specific tokens
2. **New JSON-RPC Pattern**: Introduces `"arg"` parameter for passing arguments to service calls (not seen in login or getSoftwareImages)
3. **Service Discovery**: This validates the extensibility of the JSON-RPC architecture - new services follow the same pattern

**Implementation Notes for Post-POV:**

- **Data Source Name**: `bcm_cmdevice_node` or `data "bcm_cmdevice" "node"`
- **Schema**: TBD - requires API response analysis (not captured in POV testing)
- **Argument Handling**: Provider client must support optional `"arg"` field in JSON-RPC request body
- **Pattern Generalization**: Consider abstracting JSON-RPC client to support both argument-less calls (getSoftwareImages) and parameterized calls (getNode)

**Future Data Sources (cmdevice service):**

Potential data sources for future implementation:

- `bcm_cmdevice_node` - Query cluster node information (master, worker nodes)
- `bcm_cmdevice_cluster` - Query cluster topology and configuration
- Additional cmdevice calls to be discovered through API documentation review

**Validation Status:**

- Authentication: VERIFIED - Same token works across services
- Request Pattern: VERIFIED - JSON-RPC with `"arg"` parameter
- Response Schema: NOT TESTED - deferred to post-POV
- Integration Complexity: LOW - Same authentication pattern, minor client generalization needed
