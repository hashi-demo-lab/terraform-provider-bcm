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
//	true - Resource is deleted (not found or empty response)
//	false - Resource still exists after all retries
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
//	if deleted {
//	    t.Logf("✓ Resource deleted successfully")
//	} else {
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
			return true // Deleted - API returned error (likely 404 or resource not found)
		}

		// Check if response is empty
		if len(body) == 0 {
			return true // Deleted - empty response
		}

		// Parse response to check if data is empty object
		var data map[string]interface{}
		if json.Unmarshal(body, &data) == nil && len(data) == 0 {
			return true // Deleted - empty JSON object
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
func generateUniqueTestName(prefix string) string {
	timestamp := time.Now().Format("20060102-150405")
	nanos := time.Now().Nanosecond()
	return fmt.Sprintf("%s-%s-%d", prefix, timestamp, nanos)
}
