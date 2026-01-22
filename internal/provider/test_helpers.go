// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"
)

const (
	// TestEventualConsistencyDelay is the delay to wait for BCM eventual consistency
	// in drift detection tests and cleanup operations. BCM API operations may take
	// time to propagate, so this delay allows the system to reach a consistent state.
	TestEventualConsistencyDelay = 2 * time.Second
)

// BCM API Field Name Mappings
//
// The BCM API uses camelCase field names, while Terraform schemas use snake_case.
// This mapping is critical for drift detection tests when modifying resources via BCM API.
//
// Common Patterns:
// - snake_case → camelCase: kernel_parameters → kernelParameters
// - Acronyms uppercase: enable_sol → enableSOL, sol_flow_control → solFlowControl
// - Preserved names: notes → notes (no transformation)
//
// bcm_cmpart_softwareimage Field Mappings:
//   Terraform Schema       BCM API Field
//   -----------------      ---------------
//   kernel_parameters   → kernelParameters
//   enable_sol          → enableSOL
//   sol_speed           → solSpeed
//   sol_flow_control    → solFlowControl
//   sol_port            → solPort
//   kernel_output_console → kernelOutputConsole
//   kernel_version      → kernelVersion
//   notes               → notes
//   path                → path
//   original_image      → originalImage
//   software_image_proxy → softwareImageProxy
//   modules             → modules
//
// bcm_cmdevice_category Field Mappings:
//   Terraform Schema           BCM API Field
//   -----------------          ---------------
//   kernel_parameters       → kernelParameters
//   notes                   → notes
//   install_boot_record     → installBootRecord
//   allow_networking_restart → allowNetworkingRestart
//   management_network      → managementNetwork
//   boot_loader             → bootLoader
//   software_image_proxy    → softwareImageProxy
//   bmc_settings            → bmcSettings
//   force                   → force (not persisted in BCM)
//
// bcm_cmkube_cluster Field Mappings:
//   Terraform Schema       BCM API Field
//   -----------------      ---------------
//   master_nodes        → masterNodes
//   worker_nodes        → workerNodes
//   management_network  → managementNetwork
//   overlay_network     → overlayNetwork
//   dns_servers         → dnsServers
//   version             → version
//   cni_plugin          → cniPlugin
//   storage_classes     → storageClasses (JSON-encoded)
//   load_balancer_mode  → loadBalancerMode
//   addons              → addons (JSON-encoded)
//   ingress_controller  → ingressController (JSON-encoded)
//   creation_time       → creationTime
//   revision_id         → revisionID
//   force               → force (not persisted in BCM)

// createTestBCMClient creates an authenticated BCM client for test use
//
// This helper function is used by:
// - Drift detection tests (PreConfig to modify resources externally)
// - CheckDestroy functions (to verify resource deletion)
// - PreCheck cleanup functions (to remove leftover test resources)
//
// Environment Variables (required):
//
//	BCM_ENDPOINT - BCM API endpoint (e.g., https://172.21.15.254:8081)
//	BCM_USERNAME - BCM username (e.g., root)
//	BCM_PASSWORD - BCM password
//
// Error Handling:
//
//	Calls t.Fatalf if credentials are missing or authentication fails
//
// Example Usage:
//
//	client := createTestBCMClient(t)
//	_, err := client.CallJSONRPC(ctx, "CMPart", "getSoftwareImage", imageName)
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

// verifyResourceDeleted polls BCM API to verify resource deletion with exponential backoff
//
// This function handles BCM's eventual consistency by retrying resource lookups with
// exponentially increasing wait times. It's designed to complete within the 30-second
// PreCheck requirement specified in FR-016.
//
// Parameters:
//
//	ctx - Context for API calls (can include timeout)
//	client - Authenticated BCM client
//	service - BCM service name (e.g., "CMPart", "cmdevice")
//	method - BCM method name (e.g., "getSoftwareImage", "getCategory")
//	identifier - Resource identifier (name or UUID)
//	maxRetries - Maximum retry attempts (e.g., 4 for 15s total: 1+2+4+8)
//
// Returns:
//
//	bool - true if resource is deleted (not found or empty response), false if still exists
//
// Retry Schedule (maxRetries=4):
//
//	Retry 0: Wait 1s, check (total: 1s)
//	Retry 1: Wait 2s, check (total: 3s)
//	Retry 2: Wait 4s, check (total: 7s)
//	Retry 3: Wait 8s, check (total: 15s)
//	Total: 15 seconds (within 30s requirement)
//
// Example Usage:
//
//	deleted := verifyResourceDeleted(ctx, client, "CMPart", "getSoftwareImage", imageName, 4)
//	if !deleted {
//	    t.Logf("⚠ Warning: Resource may still exist after retries")
//	}
func verifyResourceDeleted(ctx context.Context, client *BCMClient, service, method, identifier string, maxRetries int) bool {
	waitTime := 1 * time.Second

	for retry := 0; retry < maxRetries; retry++ {
		time.Sleep(waitTime)

		// Attempt to read resource
		body, err := client.CallJSONRPC(ctx, service, method, identifier)

		// Check if resource is gone (error response indicates not found)
		if err != nil {
			errStr := err.Error()
			// Only treat "not found" type errors as successful deletion
			// Common patterns: HTTP 404, "not found", "does not exist", empty result, null response
			if containsAny(errStr, []string{"404", "not found", "does not exist", "unknown", "no such", "null"}) {
				return true
			}
			// Other errors (network, auth, server errors) - resource status unknown, keep retrying
			waitTime *= 2
			continue
		}

		// Check if response is empty
		if len(body) == 0 {
			return true
		}

		// Parse response to check if data is empty object
		var data map[string]interface{}
		if json.Unmarshal(body, &data) == nil && len(data) == 0 {
			return true
		}

		// Resource still exists, wait longer for next retry
		waitTime *= 2 // Exponential backoff
	}

	// Resource still exists after all retries
	return false
}

// getResourceUUIDByName queries BCM API to get a resource's UUID by name
//
// This helper function extracts the common pattern used in drift detection tests
// where we need to find a resource's UUID before modifying it externally via the BCM API.
//
// Parameters:
//
//	t - Testing instance
//	service - BCM service name (e.g., "CMPart", "cmdevice")
//	method - BCM method to get resource by name (e.g., "getSoftwareImage", "getCategory")
//	name - Resource name to look up
//
// Returns:
//
//	UUID string of the resource
//
// Error Handling:
//
//	Calls t.Fatalf if API call fails or UUID cannot be extracted
//
// Example Usage:
//
//	uuid := getResourceUUIDByName(t, "CMPart", "getSoftwareImage", "test-image")
//	// Use uuid for BCM API update call
func getResourceUUIDByName(t *testing.T, service, method, name string) string {
	client := createTestBCMClient(t)
	ctx := context.Background()

	// Query BCM API with resource name
	body, err := client.CallJSONRPC(ctx, service, method, name)
	if err != nil {
		t.Fatalf("Failed to query resource %s via %s.%s: %v", name, service, method, err)
	}

	// Parse response to extract UUID
	var resourceData map[string]interface{}
	if err := json.Unmarshal(body, &resourceData); err != nil {
		t.Fatalf("Failed to parse resource response: %v", err)
	}

	// Extract UUID field
	uuid, ok := resourceData["uuid"].(string)
	if !ok || uuid == "" {
		t.Fatalf("Resource %s does not have a valid uuid field", name)
	}

	return uuid
}

// ============================================================================
// RFC 1123 DNS Label Requirements for Test Hostname Generation Functions
// ============================================================================
//
// RFC 1123 specifies strict requirements for DNS labels used as hostnames.
// Test hostname generation functions in this file MUST comply with these rules
// to ensure test resources are compatible with real-world Terraform provider
// use cases and BCM API hostname validation.
//
// RFC 1123 DNS Label Specification:
//
//   Length: 1-63 characters maximum
//   Characters: Lowercase letters (a-z), digits (0-9), and hyphens (-)
//   Start: MUST begin with alphanumeric character (a-z or 0-9)
//   End: MUST end with alphanumeric character (a-z or 0-9)
//   Pattern: ^[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$
//
// Why This Matters for Error Tests:
//
// Validation tests (e.g., TestAccCMDeviceDevice_ValidationInvalidHostname) must
// deliberately use INVALID hostnames to test error handling. Meanwhile, setup,
// cleanup, and prerequisite steps need VALID RFC 1123 hostnames.
//
// Invalid Test Cases (should trigger validation errors):
//   - "UPPERCASE" (uppercase not allowed, must be lowercase)
//   - "-leadinghyphen" (must start with alphanumeric)
//   - "trailing-" (must end with alphanumeric)
//   - "x" * 70 (exceeds 63 character maximum)
//   - "test_name" (underscores not allowed)
//   - "test.name" (dots not allowed in single labels)
//
// Valid Test Cases (RFC 1123 compliant):
//   - "tftest-device-20251124-150405-123456789-1234" (full unique name)
//   - "tftest-img-2511240714-a3f9" (short RFC 1123 compliant)
//   - "node1", "web-server", "app-test-01" (simple examples)
//
// Two Hostname Generation Functions:
//
// 1. generateUniqueTestName(prefix):
//    - Format: "prefix-YYYYMMDD-HHMMSS-nanoseconds-pid"
//    - Length: 44-60+ characters (may exceed 63 char limit)
//    - Purpose: General resource naming where length isn't a constraint
//    - Use for: Resource names that aren't hostnames (device IDs, image names, etc.)
//    - WARNING: May violate RFC 1123 63-char limit with long prefixes
//
// 2. generateShortTestName(prefix):
//    - Format: "tftest-{prefix}-{YYMMDDHHmm}-{4hex}"
//    - Length: Always 28-35 characters (always under 63 char limit)
//    - Purpose: RFC 1123 compliant hostname generation
//    - Use for: Actual hostnames and DNS-compatible test resource names
//    - Guarantees: Always valid RFC 1123 DNS labels
//
// Recommended Format Breakdown for generateShortTestName():
//
//   Component               Chars   Example      Purpose
//   ─────────────────       ─────   ─────────    ─────────────────────────
//   "tftest-" constant      7       tftest-      Marks name as test resource
//   User prefix             1-10    img, dev     Short descriptive label
//   "-" separator           1       -            Separator
//   Timestamp (YYMMDDHHmm)  10      2511240714   Minute precision timestamp
//   "-" separator           1       -            Separator
//   Hex suffix (4 chars)    4       a3f9         65536 combinations per minute
//   ─────────────────────────────────────────────────────────────────────
//   TOTAL:                  28-35   tftest-img-2511240714-a3f9
//
// Examples of Good vs Bad Test Hostname Names:
//
//   GOOD - RFC 1123 Compliant:
//   ✓ "tftest-device-20251124-150405-123456789-1234" (unique, timestamp-based)
//   ✓ "tftest-img-2511240714-a3f9" (short, hex-suffixed)
//   ✓ "tftest-err-cat-nf-2511240714-3f9a" (error test: category not found)
//   ✓ "node1", "web-server", "app-test-01" (human-readable)
//
//   BAD - RFC 1123 Violations:
//   ✗ "UPPERCASE-HOSTNAME" (uppercase not allowed)
//   ✗ "-leadinghyphen" (starts with hyphen)
//   ✗ "trailing-hyphen-" (ends with hyphen)
//   ✗ "double--hyphen" (consecutive hyphens poor practice)
//   ✗ "x" * 70 (exceeds 63 char maximum)
//   ✗ "test_name_with_underscores" (underscores not allowed)
//   ✗ "test.name.with.dots" (dots only allowed in FQDN, not single labels)
//
// Length Accounting Examples:
//
//   generateUniqueTestName("tftest-device"):
//   "tftest-device-20251124-150405-123456789-1234" = 45 chars (OK)
//   "tftest-very-long-prefix-error-scenario-20251124-150405-123456789-1234" = 70 chars (TOO LONG!)
//
//   generateShortTestName("img"):
//   "tftest-img-2511240714-a3f9" = 28 chars (always under 63)
//
//   generateShortTestName("err-cat-nf"):
//   "tftest-err-cat-nf-2511240714-3f9a" = 35 chars (always under 63)
//
// Prefix Abbreviation Recommendations:
//
//   For generateShortTestName(), use short prefixes (max 10 chars):
//   - "img" for software images
//   - "cat" for categories
//   - "net" for networks
//   - "dev" for devices
//   - "clust" for clusters
//   - "err-cat-nf" for error test: category not found
//   - "err-img-ex" for error test: image already exists
//   - "err-net-dupe" for error test: network duplicate

// generateUniqueTestName creates a unique test resource name with timestamp and nanosecond suffix
//
// This function ensures test resources have unique names to avoid conflicts when running
// tests in parallel or when previous test cleanup failed. It combines timestamp with
// nanoseconds to guarantee uniqueness even when multiple tests start in rapid succession.
//
// Parameters:
//
//	prefix - Resource name prefix (e.g., "test-image", "test-category")
//
// Returns:
//
//	Unique resource name with format: prefix-YYYYMMDD-HHMMSS-nanoseconds
//
// Example Usage:
//
//	name := generateUniqueTestName("test-image")
//	// Returns: "test-image-20250121-143052-123456789"
//
// generateUniqueTestName creates a unique test resource name with multiple uniqueness factors.
//
// Uniqueness strategy:
// - Date/time: 20060102-150405 (second precision)
// - Nanoseconds: 0-999999999 (nanosecond within current second)
// - Process ID: Helps with parallel test execution across processes
//
// This ensures uniqueness even when:
// - Multiple tests run in parallel (different PIDs)
// - Tests run in rapid succession (nanosecond precision)
// - Tests run across multiple days (date component)
//
// Example output: tftest-image-20251124-150405-123456789-12345
//
// Returns: "prefix-YYYYMMDD-HHMMSS-nanoseconds-pid"
//
// WARNING: This function may generate names longer than 63 characters, which violates
// RFC 1123 DNS label requirements. For hostnames and DNS-compatible names, use
// generateShortTestName() instead.
func generateUniqueTestName(prefix string) string {
	now := time.Now()
	timestamp := now.Format("20060102-150405")
	nanos := now.Nanosecond()
	pid := os.Getpid()
	return fmt.Sprintf("%s-%s-%d-%d", prefix, timestamp, nanos, pid)
}

// generateShortTestName creates a unique test resource name compliant with RFC 1123 DNS labels.
//
// RFC 1123 DNS Label Requirements:
// - Maximum length: 63 characters
// - Must start with alphanumeric character
// - Can contain alphanumeric characters and hyphens
// - Must end with alphanumeric character
//
// Uniqueness strategy:
// - Short prefix: Descriptive abbreviation (max 10 chars recommended)
// - Compact timestamp: YYMMDDHHmm (10 digits, minute precision)
// - Hex suffix: 4-character hex from nanoseconds (65536 combinations per minute)
//
// Format: "tftest-{prefix}-{YYMMDDHHmm}-{4hex}"
// Example: "tftest-img-2511240714-a3f9" (28 characters)
// Example: "tftest-err-cat-nf-2511240714-3f9a" (35 characters with longer prefix)
//
// Max length calculation:
// - "tftest-" prefix: 7 characters
// - User prefix: up to 10 characters (recommended)
// - Separators: 2 hyphens = 2 characters
// - Timestamp: 10 characters (YYMMDDHHmm)
// - Hex suffix: 4 characters
// - Total: 7 + 10 + 2 + 10 + 4 = 33 characters (well under 63 limit)
//
// Suggested prefix abbreviations:
// - "img" for software images
// - "cat" for categories
// - "net" for networks
// - "dev" for devices
// - "clust" for clusters
// - "err-cat-nf" for error test: category not found
// - "err-img-ex" for error test: image already exists
//
// This function replaces the previous implementation which could generate names
// exceeding 63 characters (e.g., "err-cat-nf-150405-123456789" = ~28 chars base,
// but "tftest-device-error-category-notfound-150405-123456789" = ~60+ chars).
//
// Parameters:
//
//	prefix - Short descriptive prefix (max 10 chars recommended, e.g., "img", "cat", "err-cat-nf")
//
// Returns:
//
//	RFC 1123 compliant unique name with format: tftest-prefix-YYMMDDHHmm-4hex
//
// Example Usage:
//
//	name := generateShortTestName("img")
//	// Returns: "tftest-img-2511240714-a3f9" (28 chars)
//
//	name := generateShortTestName("err-cat-nf")
//	// Returns: "tftest-err-cat-nf-2511240714-3f9a" (35 chars)
func generateShortTestName(prefix string) string {
	now := time.Now()
	// Compact timestamp: YYMMDDHHmm (minute precision)
	timestamp := now.Format("0601021504")
	// 4-character hex from nanoseconds (0x0000 to 0xFFFF range)
	// Use nanoseconds modulo 65536 to get 4-hex-digit value
	hexSuffix := fmt.Sprintf("%04x", now.Nanosecond()%65536)
	return fmt.Sprintf("tftest-%s-%s-%s", prefix, timestamp, hexSuffix)
}

// generateUniqueMAC creates a unique MAC address for testing
//
// This function generates a valid MAC address with a locally administered unicast prefix (02:00:00)
// followed by unique bytes derived from the current time. This ensures:
// - No conflicts with real hardware MAC addresses (locally administered bit set)
// - Uniqueness across parallel test runs (nanosecond-based suffix)
// - Valid MAC address format for BCM
//
// Returns:
//
//	Unique MAC address with format: 02:00:00:XX:YY:ZZ where XX:YY:ZZ are time-based
//
// Example Usage:
//
//	mac := generateUniqueMAC()
//	// Returns: "02:00:00:12:34:56"
func generateUniqueMAC() string {
	now := time.Now()
	// Use hour, minute, second, and millisecond to create unique MAC
	// This gives us 24*60*60*1000 = 86,400,000 possible combinations per day
	nanos := now.Nanosecond()
	hour := now.Hour()
	minute := now.Minute()
	second := now.Second()

	// Combine time components to create unique bytes
	// Use modulo to ensure values fit in single bytes (0-255)
	byte1 := uint8((hour*60 + minute) % 256)
	byte2 := uint8((second*1000 + nanos/1000000) % 256)
	byte3 := uint8((nanos / 1000) % 256)

	return fmt.Sprintf("02:00:00:%02X:%02X:%02X", byte1, byte2, byte3)
}

// =============================================================================
// CMEtcd Test Helpers
// =============================================================================

// bcm_cmetcd_cluster Field Mappings:
//
//	Terraform Schema       BCM API Field
//	-----------------      ---------------
//	heartbeat_interval  → heartbeatInterval
//	election_timeout    → electionTimeout
//	options             → options (JSON-encoded)
//	creation_time       → creationTime
//	revision_id         → revisionID

// getTestEtcdClusterUUID queries BCM for an existing EtcdCluster UUID by name.
// This helper is useful for tests that need to reference an existing etcd cluster.
//
// Parameters:
//
//	t - Testing instance
//	name - Etcd cluster name to look up
//
// Returns:
//
//	UUID string of the etcd cluster, or empty string if not found
//
//nolint:unused // test helper available for future tests
func getTestEtcdClusterUUID(t *testing.T, name string) string {
	client := createTestBCMClient(t)
	ctx := context.Background()

	// Query BCM API with cluster name
	body, err := client.CallJSONRPC(ctx, "cmetcd", "getEtcdCluster", name)
	if err != nil {
		// Cluster might not exist, return empty string
		return ""
	}

	// Parse response to extract UUID
	var clusterData map[string]interface{}
	if err := json.Unmarshal(body, &clusterData); err != nil {
		return ""
	}

	uuid, ok := clusterData["uuid"].(string)
	if !ok {
		return ""
	}

	return uuid
}

// createTestEtcdCluster creates an EtcdCluster via BCM API for test setup.
// This is useful for tests that need a pre-existing etcd cluster to reference.
//
// Parameters:
//
//	t - Testing instance
//	name - Etcd cluster name
//
// Returns:
//
//	UUID of the created etcd cluster
//
//nolint:unused // test helper available for future tests
func createTestEtcdCluster(t *testing.T, name string) string {
	client := createTestBCMClient(t)
	ctx := context.Background()

	// Generate UUID for the cluster
	clusterUUID := generateUUID()

	// Build the etcd cluster entity
	entity := map[string]interface{}{
		"baseType":          "EtcdCluster",
		"childType":         "",
		"uuid":              clusterUUID,
		"name":              name,
		"heartbeatInterval": 100,
		"electionTimeout":   1000,
		"options":           map[string]interface{}{},
		"modified":          true,
		"to_be_removed":     false,
		"revision":          "",
	}

	// Create via BCM API
	_, err := client.CallJSONRPC(ctx, "cmetcd", "addEtcdCluster", entity)
	if err != nil {
		t.Fatalf("Failed to create test etcd cluster: %v", err)
	}

	// Wait for eventual consistency
	time.Sleep(TestEventualConsistencyDelay)

	return clusterUUID
}

// deleteTestEtcdCluster removes an EtcdCluster via BCM API for test cleanup.
//
// Parameters:
//
//	t - Testing instance
//	uuid - Etcd cluster UUID to delete
//
//nolint:unused // test helper available for future tests
func deleteTestEtcdCluster(t *testing.T, uuid string) {
	if uuid == "" {
		return
	}

	client := createTestBCMClient(t)
	ctx := context.Background()

	_, err := client.CallJSONRPC(ctx, "cmetcd", "removeEtcdCluster", uuid)
	if err != nil {
		t.Logf("Warning: Failed to delete test etcd cluster %s: %v", uuid, err)
	}

	// Wait for eventual consistency
	time.Sleep(TestEventualConsistencyDelay)
}

// =============================================================================
// CMKube Aligned Test Helpers
// =============================================================================

// getTestNetworkUUID queries BCM for a network UUID by name.
// Used for tests that need network UUIDs for KubeCluster configuration.
//
// Parameters:
//
//	t - Testing instance
//	name - Network name to look up
//
// Returns:
//
//	UUID string of the network, or empty string if not found
//
//nolint:unused // test helper available for future tests
func getTestNetworkUUID(t *testing.T, name string) string {
	client := createTestBCMClient(t)
	ctx := context.Background()

	body, err := client.CallJSONRPC(ctx, "cmnet", "getNetwork", name)
	if err != nil {
		return ""
	}

	var networkData map[string]interface{}
	if err := json.Unmarshal(body, &networkData); err != nil {
		return ""
	}

	uuid, ok := networkData["uuid"].(string)
	if !ok {
		return ""
	}

	return uuid
}

// getFirstAvailableNetworkUUID returns the UUID of the first available network.
// Useful for tests that need any valid network reference.
//
//nolint:unused // test helper available for future tests
func getFirstAvailableNetworkUUID(t *testing.T) string {
	client := createTestBCMClient(t)
	ctx := context.Background()

	body, err := client.CallJSONRPC(ctx, "cmnet", "getNetworks")
	if err != nil {
		t.Skipf("Failed to get networks: %v", err)
		return ""
	}

	var networks []map[string]interface{}
	if err := json.Unmarshal(body, &networks); err != nil {
		t.Skipf("Failed to parse networks: %v", err)
		return ""
	}

	if len(networks) == 0 {
		t.Skip("No networks available for test")
		return ""
	}

	uuid, ok := networks[0]["uuid"].(string)
	if !ok {
		t.Skip("First network has no valid UUID")
		return ""
	}

	return uuid
}
