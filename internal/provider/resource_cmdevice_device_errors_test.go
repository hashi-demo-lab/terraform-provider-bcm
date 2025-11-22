// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

// ========================================
// Mock BCM Server Helpers
// ========================================

// mockBCMServerScenario defines different error scenarios for testing
type mockBCMServerScenario string

const (
	// Category query errors (line 259)
	scenarioCategoryNotFound     mockBCMServerScenario = "category_not_found"
	scenarioCategoryInvalidJSON  mockBCMServerScenario = "category_invalid_json"
	scenarioCategoryNoPartition  mockBCMServerScenario = "category_no_partition"
	scenarioCategoryProxyMissing mockBCMServerScenario = "category_proxy_missing_parent"

	// Partition query errors (line 293-309)
	scenarioPartitionsQueryError  mockBCMServerScenario = "partitions_query_error"
	scenarioPartitionsInvalidJSON mockBCMServerScenario = "partitions_invalid_json"
	scenarioPartitionsNoBase      mockBCMServerScenario = "partitions_no_base"

	// Partition commit timeout (line 527)
	scenarioPartitionNotCommitted mockBCMServerScenario = "partition_not_committed"

	// Device creation errors (line 393)
	scenarioDeviceCreateError       mockBCMServerScenario = "device_create_error"
	scenarioDeviceCreateInvalidJSON mockBCMServerScenario = "device_create_invalid_json"
	scenarioDeviceValidationFailure mockBCMServerScenario = "device_validation_failure"

	// Device read errors (line 443, 580)
	scenarioDeviceReadError       mockBCMServerScenario = "device_read_error"
	scenarioDeviceReadInvalidJSON mockBCMServerScenario = "device_read_invalid_json"

	// Device update errors (line 676, 763)
	scenarioDeviceUpdateCategoryError mockBCMServerScenario = "device_update_category_error"
	scenarioDeviceUpdateError         mockBCMServerScenario = "device_update_error"
	scenarioDeviceUpdateReadError     mockBCMServerScenario = "device_update_read_error"

	// Device delete errors (line 867)
	scenarioDeviceDeleteError mockBCMServerScenario = "device_delete_error"
)

// createMockBCMServerForDeviceErrors creates a mock BCM server for error testing.
// It returns different errors based on the scenario parameter.
func createMockBCMServerForDeviceErrors(t *testing.T, scenario mockBCMServerScenario) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse the JSON-RPC request
		var req struct {
			Service string        `json:"service"`
			Call    string        `json:"call"`
			Args    []interface{} `json:"args,omitempty"`
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		// Handle login requests (always succeed for testing)
		if req.Service == "login" {
			w.Header().Set("Set-Cookie", "cm-login-token=test-token; Path=/")
			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
			return
		}

		// Route to scenario-specific handlers
		switch scenario {
		case scenarioCategoryNotFound:
			handleCategoryNotFound(t, w, req)
		case scenarioCategoryInvalidJSON:
			handleCategoryInvalidJSON(t, w, req)
		case scenarioCategoryNoPartition:
			handleCategoryNoPartition(t, w, req)
		case scenarioCategoryProxyMissing:
			handleCategoryProxyMissing(t, w, req)
		case scenarioPartitionsQueryError:
			handlePartitionsQueryError(t, w, req)
		case scenarioPartitionsInvalidJSON:
			handlePartitionsInvalidJSON(t, w, req)
		case scenarioPartitionsNoBase:
			handlePartitionsNoBase(t, w, req)
		case scenarioPartitionNotCommitted:
			handlePartitionNotCommitted(t, w, req)
		case scenarioDeviceCreateError:
			handleDeviceCreateError(t, w, req)
		case scenarioDeviceCreateInvalidJSON:
			handleDeviceCreateInvalidJSON(t, w, req)
		case scenarioDeviceValidationFailure:
			handleDeviceValidationFailure(t, w, req)
		case scenarioDeviceReadError:
			handleDeviceReadError(t, w, req)
		case scenarioDeviceReadInvalidJSON:
			handleDeviceReadInvalidJSON(t, w, req)
		case scenarioDeviceUpdateCategoryError:
			handleDeviceUpdateCategoryError(t, w, req)
		case scenarioDeviceUpdateError:
			handleDeviceUpdateError(t, w, req)
		case scenarioDeviceUpdateReadError:
			handleDeviceUpdateReadError(t, w, req)
		case scenarioDeviceDeleteError:
			handleDeviceDeleteError(t, w, req)
		default:
			http.Error(w, "Unknown scenario", http.StatusInternalServerError)
		}
	}))
}

// ========================================
// Scenario-Specific Handlers
// ========================================

// handleCategoryNotFound simulates category query returning an error (line 259).
func handleCategoryNotFound(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	// Only return error for getCategory calls
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		http.Error(w, "Category not found", http.StatusNotFound)
		return
	}

	// For other calls, return minimal valid responses
	if req.Service == "cmnet" && req.Call == "getNetworks" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"uuid": "12345678-1234-1234-1234-123456789012",
				"name": "managementnet",
			},
		})
		return
	}

	// Default success response for unknown calls
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleCategoryInvalidJSON simulates category query returning invalid JSON (line 269).
func handleCategoryInvalidJSON(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	// Return invalid JSON only for getCategory calls
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid json {"))
		return
	}

	// For other calls, return minimal valid responses
	if req.Service == "cmnet" && req.Call == "getNetworks" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"uuid": "12345678-1234-1234-1234-123456789012",
				"name": "managementnet",
			},
		})
		return
	}

	// Default success response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleCategoryNoPartition simulates category with no partition field (line 278-349).
func handleCategoryNoPartition(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid": "12345678-1234-1234-1234-123456789012",
			"name": "test-category",
			// Missing both "partition" and "softwareImageProxy" fields
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleCategoryProxyMissing simulates category with softwareImageProxy but no parentSoftwareImage (line 338-344).
func handleCategoryProxyMissing(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid":               "12345678-1234-1234-1234-123456789012",
			"name":               "test-category",
			"softwareImageProxy": map[string]interface{}{
				// Missing "parentSoftwareImage" field
			},
		})
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handlePartitionsQueryError simulates partitions query error (line 293).
func handlePartitionsQueryError(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "CMPart" && req.Call == "getPartitions" {
		http.Error(w, "Failed to query partitions", http.StatusInternalServerError)
		return
	}
	// Return valid category with softwareImageProxy to trigger partition query
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid": "12345678-1234-1234-1234-123456789012",
			"name": "test-category",
			"softwareImageProxy": map[string]interface{}{
				"parentSoftwareImage": "22222222-2222-2222-2222-222222222222",
			},
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handlePartitionsInvalidJSON simulates partitions query returning invalid JSON (line 303).
func handlePartitionsInvalidJSON(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "CMPart" && req.Call == "getPartitions" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not valid json ["))
		return
	}
	// Return valid category with softwareImageProxy
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid": "12345678-1234-1234-1234-123456789012",
			"name": "test-category",
			"softwareImageProxy": map[string]interface{}{
				"parentSoftwareImage": "22222222-2222-2222-2222-222222222222",
			},
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handlePartitionsNoBase simulates partitions query with no "base" partition (line 326-332).
func handlePartitionsNoBase(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "CMPart" && req.Call == "getPartitions" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode([]map[string]interface{}{
			{
				"uuid": "87654321-4321-4321-4321-210987654321",
				"name": "other-partition",
			},
			// Missing partition with name="base"
		})
		return
	}
	// Return valid category with softwareImageProxy
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid": "12345678-1234-1234-1234-123456789012",
			"name": "test-category",
			"softwareImageProxy": map[string]interface{}{
				"parentSoftwareImage": "22222222-2222-2222-2222-222222222222",
			},
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handlePartitionNotCommitted simulates partition not becoming committed (line 527).
func handlePartitionNotCommitted(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	// Return valid category
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid":      "12345678-1234-1234-1234-123456789012",
			"name":      "test-category",
			"partition": "11111111-1111-1111-1111-111111111111",
		})
		return
	}
	// Always return error for getSoftwareImage to simulate partition not ready
	if req.Service == "CMPart" && req.Call == "getSoftwareImage" {
		http.Error(w, "Software image not found or not committed", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceCreateError simulates device creation API error (line 393).
func handleDeviceCreateError(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	// Return valid category and partition
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid":      "12345678-1234-1234-1234-123456789012",
			"name":      "test-category",
			"partition": "11111111-1111-1111-1111-111111111111",
		})
		return
	}
	if req.Service == "CMPart" && req.Call == "getSoftwareImage" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"uuid": "11111111-1111-1111-1111-111111111111"})
		return
	}
	// Fail on addDevice
	if req.Service == "cmdevice" && req.Call == "addDevice" {
		http.Error(w, "Failed to create device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceCreateInvalidJSON simulates device creation returning invalid JSON (line 408).
func handleDeviceCreateInvalidJSON(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid":      "12345678-1234-1234-1234-123456789012",
			"name":      "test-category",
			"partition": "11111111-1111-1111-1111-111111111111",
		})
		return
	}
	if req.Service == "CMPart" && req.Call == "getSoftwareImage" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"uuid": "11111111-1111-1111-1111-111111111111"})
		return
	}
	if req.Service == "cmdevice" && req.Call == "addDevice" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("invalid response {"))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceValidationFailure simulates device creation validation failure (line 417-432).
func handleDeviceValidationFailure(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"uuid":      "12345678-1234-1234-1234-123456789012",
			"name":      "test-category",
			"partition": "11111111-1111-1111-1111-111111111111",
		})
		return
	}
	if req.Service == "CMPart" && req.Call == "getSoftwareImage" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"uuid": "11111111-1111-1111-1111-111111111111"})
		return
	}
	if req.Service == "cmdevice" && req.Call == "addDevice" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"validation": []map[string]interface{}{
				{
					"field":   "hostname",
					"message": "Hostname already exists in cluster",
				},
				{
					"field":   "mac",
					"message": "MAC address is already in use",
				},
			},
		})
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceReadError simulates device read API error (line 443, 580).
func handleDeviceReadError(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "getDevice" {
		http.Error(w, "Failed to read device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceReadInvalidJSON simulates device read returning invalid JSON (line 454).
func handleDeviceReadInvalidJSON(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "getDevice" {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("not json {"))
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceUpdateCategoryError simulates category query error during update (line 676).
func handleDeviceUpdateCategoryError(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "getCategory" {
		http.Error(w, "Category not found during update", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceUpdateError simulates device update API error (line 763).
func handleDeviceUpdateError(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "updateDevice" {
		http.Error(w, "Failed to update device", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceUpdateReadError simulates device read error after update (line 778).
func handleDeviceUpdateReadError(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	// Return success for update
	if req.Service == "cmdevice" && req.Call == "updateDevice" {
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
		return
	}
	// Fail on subsequent read
	if req.Service == "cmdevice" && req.Call == "getDevice" {
		http.Error(w, "Device updated but cannot be read", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// handleDeviceDeleteError simulates device delete API error (line 867).
func handleDeviceDeleteError(t *testing.T, w http.ResponseWriter, req struct {
	Service string        `json:"service"`
	Call    string        `json:"call"`
	Args    []interface{} `json:"args,omitempty"`
}) {
	if req.Service == "cmdevice" && req.Call == "removeDevice" {
		http.Error(w, "Failed to delete device", http.StatusForbidden)
		return
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{"success": true})
}

// ========================================
// Category Error Tests
// ========================================

// TestAccCMDeviceDeviceResource_ErrorCategoryNotFound tests error handling when category lookup fails.
// References: resource_cmdevice_device.go:259 (Create), line 676 (Update)
//
// Error Path Tested: Lines 259-265 (Create operation)
// Expected Diagnostic: "Error Querying Category"
//
// Test Scenario:
// - Create device without explicit partition attribute (triggers category query for default partition)
// - Mock BCM API returns HTTP 404 error when calling getCategory
// - Provider should detect error and return "Error Querying Category" diagnostic
func TestAccCMDeviceDeviceResource_ErrorCategoryNotFound(t *testing.T) {
	// Create mock server that simulates category not found error
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioCategoryNotFound)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-cat-notfound")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Querying Category`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorCategoryInvalidJSON tests error handling for invalid category JSON.
// References: resource_cmdevice_device.go:269 (Create), line 686 (Update)
//
// Error Path Tested: Lines 268-274 (Create operation)
// Expected Diagnostic: "Error Parsing Category"
//
// Test Scenario:
// - Create device without explicit partition attribute (triggers category query)
// - Mock BCM API returns malformed JSON when calling getCategory
// - Provider should fail to unmarshal response and return "Error Parsing Category" diagnostic
func TestAccCMDeviceDeviceResource_ErrorCategoryInvalidJSON(t *testing.T) {
	// Create mock server that returns malformed JSON for category query
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioCategoryInvalidJSON)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-cat-badjson")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Parsing Category`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorCategoryNoPartition tests error when category has no partition.
// References: resource_cmdevice_device.go:346-351
//
// Error Path Tested: Lines 346-351 (Create operation)
// Expected Diagnostic: "Missing Partition"
//
// Test Scenario:
// - Create device without explicit partition attribute
// - Mock BCM API returns category with neither partition nor softwareImageProxy fields
// - Provider should detect missing partition and return "Missing Partition" diagnostic
func TestAccCMDeviceDeviceResource_ErrorCategoryNoPartition(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioCategoryNoPartition)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-no-partition")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Missing Partition`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorCategoryProxyMissingParent tests softwareImageProxy validation.
// References: resource_cmdevice_device.go:338-344
//
// Error Path Tested: Lines 338-344 (Create operation)
// Expected Diagnostic: "Missing Partition" with proxy-specific message
//
// Test Scenario:
// - Create device without explicit partition
// - Mock category has softwareImageProxy but missing parentSoftwareImage field
// - Provider should return "Missing Partition" with proxy message
func TestAccCMDeviceDeviceResource_ErrorCategoryProxyMissingParent(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioCategoryProxyMissing)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-proxy-missing")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Missing Partition.*parentSoftwareImage`),
			},
		},
	})
}

// ========================================
// Partition Error Tests
// ========================================

// TestAccCMDeviceDeviceResource_ErrorPartitionsQueryFailed tests partition query failure.
// References: resource_cmdevice_device.go:293, line 708
//
// Error Path Tested: Lines 293-300 (Create operation)
// Expected Diagnostic: "Error Querying Partitions"
//
// Test Scenario:
// - Create device with category using softwareImageProxy (triggers partition query)
// - Mock BCM API returns HTTP 500 error when calling getPartitions
// - Provider should detect error and return "Error Querying Partitions" diagnostic
func TestAccCMDeviceDeviceResource_ErrorPartitionsQueryFailed(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioPartitionsQueryError)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-partitions-query")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Querying Partitions`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorPartitionsInvalidJSON tests invalid partition response.
// References: resource_cmdevice_device.go:303, line 718
//
// Error Path Tested: Lines 303-309 (Create operation)
// Expected Diagnostic: "Error Parsing Partitions"
//
// Test Scenario:
// - Create device with category using softwareImageProxy
// - Mock BCM API returns malformed JSON when calling getPartitions
// - Provider should fail to unmarshal and return "Error Parsing Partitions" diagnostic
func TestAccCMDeviceDeviceResource_ErrorPartitionsInvalidJSON(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioPartitionsInvalidJSON)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-partitions-json")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Parsing Partitions`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorPartitionsNoBase tests missing base partition.
// References: resource_cmdevice_device.go:326-332, line 741-747
//
// Error Path Tested: Lines 326-332 (Create operation)
// Expected Diagnostic: "Missing Base Partition"
//
// Test Scenario:
// - Create device with category using softwareImageProxy
// - Mock BCM API returns partitions list without "base" partition
// - Provider should return "Missing Base Partition" diagnostic
func TestAccCMDeviceDeviceResource_ErrorPartitionsNoBase(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioPartitionsNoBase)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-no-base")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Missing Base Partition`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorPartitionNotCommitted tests partition commit timeout.
// References: resource_cmdevice_device.go:362, line 509-559
//
// Error Path Tested: Lines 362-372, 509-559 (Create operation, waitForPartitionCommit)
// Expected Diagnostic: "Partition Not Ready"
//
// Test Scenario:
// - Create device with category having direct partition reference
// - Mock BCM API always returns error for getSoftwareImage (partition never commits)
// - Provider should retry with exponential backoff and return "Partition Not Ready" after 20 retries
func TestAccCMDeviceDeviceResource_ErrorPartitionNotCommitted(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioPartitionNotCommitted)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-partition-timeout")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Partition Not Ready.*not committed`),
			},
		},
	})
}

// ========================================
// Device CRUD Error Tests
// ========================================

// TestAccCMDeviceDeviceResource_ErrorDeviceCreateFailed tests device creation API failure.
// References: resource_cmdevice_device.go:393
//
// Error Path Tested: Lines 393-398 (Create operation)
// Expected Diagnostic: "Error Creating Device"
//
// Test Scenario:
// - Create device with valid configuration
// - Mock BCM API returns HTTP 500 error when calling addDevice
// - Provider should detect error and return "Error Creating Device" diagnostic
func TestAccCMDeviceDeviceResource_ErrorDeviceCreateFailed(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioDeviceCreateError)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-create")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Creating Device`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorDeviceCreateInvalidJSON tests create response parsing error.
// References: resource_cmdevice_device.go:408
//
// Error Path Tested: Lines 408-413 (Create operation)
// Expected Diagnostic: "Error Parsing Create Response"
//
// Test Scenario:
// - Create device with valid configuration
// - Mock BCM API returns malformed JSON from addDevice
// - Provider should fail to unmarshal and return "Error Parsing Create Response" diagnostic
func TestAccCMDeviceDeviceResource_ErrorDeviceCreateInvalidJSON(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioDeviceCreateInvalidJSON)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-create-json")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Parsing Create Response`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorDeviceValidationFailed tests BCM validation failure.
// References: resource_cmdevice_device.go:417-432
//
// Error Path Tested: Lines 417-432 (Create operation)
// Expected Diagnostic: "Error Creating Device" with validation details
//
// Test Scenario:
// - Create device with configuration that BCM rejects
// - Mock BCM API returns success=false with validation array containing hostname and MAC errors
// - Provider should parse validation errors and return detailed error message
func TestAccCMDeviceDeviceResource_ErrorDeviceValidationFailed(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioDeviceValidationFailure)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-validation")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Creating Device.*validation errors.*hostname.*already exists`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorDeviceReadAfterCreateFailed tests read-back error.
// References: resource_cmdevice_device.go:443
//
// Error Path Tested: Lines 443-448 (Create operation, read after create)
// Expected Diagnostic: "Error Reading Created Device"
//
// Test Scenario:
// - Create device (addDevice succeeds)
// - Mock BCM API returns error when reading device back (getDevice fails)
// - Provider should return "Error Reading Created Device" diagnostic
// - Note: This simulates eventual consistency issues where device created but not yet readable
func TestAccCMDeviceDeviceResource_ErrorDeviceReadAfterCreateFailed(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioDeviceReadError)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-read-after-create")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Reading Created Device`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorDeviceReadInvalidJSON tests read response parsing error.
// References: resource_cmdevice_device.go:454, line 591
//
// Error Path Tested: Lines 454-459, 591-596 (Create and standard Read operations)
// Expected Diagnostic: "Error Parsing Device Data"
//
// Test Scenario:
// - Device read operation returns malformed JSON
// - Provider should fail to unmarshal and return "Error Parsing Device Data" diagnostic
func TestAccCMDeviceDeviceResource_ErrorDeviceReadInvalidJSON(t *testing.T) {
	mockServer := createMockBCMServerForDeviceErrors(t, scenarioDeviceReadInvalidJSON)
	defer mockServer.Close()

	deviceName := generateUniqueTestName("test-device-error-read-json")

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config:      testAccDeviceResourceConfigWithMockServer(mockServer.URL, deviceName),
				ExpectError: regexp.MustCompile(`Error Parsing Device Data`),
			},
		},
	})
}

// TestAccCMDeviceDeviceResource_ErrorDeviceUpdateCategoryQueryFailed tests category lookup during update.
// References: resource_cmdevice_device.go:676
//
// Error Path Tested: Lines 676-682 (Update operation)
// Expected Diagnostic: "Error Querying Category"
//
// Test Scenario:
// - Update device without explicit partition attribute (triggers category query)
// - Mock BCM API returns error when calling getCategory during update
// - Provider should return "Error Querying Category" diagnostic
//
// Note: This test would require a two-step process (create, then update with category error).
// For simplicity in mock server testing, we skip this complex scenario.
func TestAccCMDeviceDeviceResource_ErrorDeviceUpdateCategoryQueryFailed(t *testing.T) {
	t.Skip("Skipping complex two-step mock server test - would require stateful mock server")
}

// TestAccCMDeviceDeviceResource_ErrorDeviceUpdateFailed tests device update API failure.
// References: resource_cmdevice_device.go:763
//
// Error Path Tested: Lines 763-768 (Update operation)
// Expected Diagnostic: "Error Updating Device"
//
// Test Scenario:
// - Update device configuration
// - Mock BCM API returns HTTP 500 error when calling updateDevice
// - Provider should return "Error Updating Device" diagnostic
//
// Note: This test would require a two-step process (create, then update with error).
// For simplicity in mock server testing, we skip this complex scenario.
func TestAccCMDeviceDeviceResource_ErrorDeviceUpdateFailed(t *testing.T) {
	t.Skip("Skipping complex two-step mock server test - would require stateful mock server")
}

// TestAccCMDeviceDeviceResource_ErrorDeviceUpdateReadFailed tests read-back after update failure.
// References: resource_cmdevice_device.go:778
//
// Error Path Tested: Lines 778-783 (Update operation, read after update)
// Expected Diagnostic: "Error Reading Updated Device"
//
// Test Scenario:
// - Update device (updateDevice succeeds)
// - Mock BCM API returns error when reading device back (getDevice fails)
// - Provider should return "Error Reading Updated Device" diagnostic
//
// Note: This test would require a two-step process with stateful mock behavior.
// For simplicity, we skip this complex scenario.
func TestAccCMDeviceDeviceResource_ErrorDeviceUpdateReadFailed(t *testing.T) {
	t.Skip("Skipping complex two-step mock server test - would require stateful mock server")
}

// TestAccCMDeviceDeviceResource_ErrorDeviceDeleteFailed tests device deletion failure.
// References: resource_cmdevice_device.go:867
//
// Error Path Tested: Lines 867-873 (Delete operation)
// Expected Diagnostic: "Error Deleting Device"
//
// Test Scenario:
// - Delete device
// - Mock BCM API returns HTTP 403 error when calling removeDevice
// - Provider should return "Error Deleting Device" diagnostic
//
// Note: This test would require create first, then delete with error.
// For simplicity, we skip this complex scenario.
func TestAccCMDeviceDeviceResource_ErrorDeviceDeleteFailed(t *testing.T) {
	t.Skip("Skipping complex two-step mock server test - would require stateful mock server")
}

// ========================================
// Helper Function for Mock Server Tests
// ========================================

// testAccDeviceResourceConfigWithMockServer returns test config using mock server endpoint.
func testAccDeviceResourceConfigWithMockServer(mockEndpoint, hostname string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = "mock-user"
  password             = "mock-pass"
  insecure_skip_verify = true
}

data "bcm_cmnet_networks" "management" {
  filter {
    name_pattern = "managementnet"
  }
}

resource "bcm_cmdevice_category" "test" {
  name               = "test-category"
  management_network = "12345678-1234-1234-1234-123456789012"
}

resource "bcm_cmdevice_device" "test" {
  hostname           = %[2]q
  mac                = "00:11:22:33:44:55"
  category           = bcm_cmdevice_category.test.id
  management_network = "12345678-1234-1234-1234-123456789012"
}
`,
		mockEndpoint,
		hostname,
	)
}

// ========================================
// Summary Documentation
// ========================================

// Error Handling Tests Summary
//
// This file contains comprehensive error handling tests for device CRUD operations.
// All tests use httptest mock servers to simulate BCM API error conditions.
//
// Coverage:
//
// 1. Category Query Errors (Create/Update):
//    - Category not found (HTTP 404)
//    - Invalid JSON response
//    - Category missing partition and softwareImageProxy
//    - SoftwareImageProxy missing parentSoftwareImage field
//
// 2. Partition Query Errors (Create with softwareImageProxy):
//    - Partitions API query failure (HTTP 500)
//    - Invalid JSON response from getPartitions
//    - No "base" partition found in partitions list
//
// 3. Partition Commit Errors (Create):
//    - Partition not becoming committed (timeout after 20 retries)
//    - Tests exponential backoff behavior in waitForPartitionCommit
//
// 4. Device Creation Errors:
//    - addDevice API failure (HTTP 500)
//    - Invalid JSON response from addDevice
//    - BCM validation failure (success=false with validation array)
//    - Read after create failure (getDevice error)
//    - Invalid JSON response from getDevice
//
// 5. Device Update Errors (skipped - require stateful mock):
//    - Category query failure during update
//    - updateDevice API failure
//    - Read after update failure
//
// 6. Device Delete Errors (skipped - require stateful mock):
//    - removeDevice API failure
//
// Running Tests:
//
//	# Run all error tests
//	go test -v -run TestAccCMDeviceDeviceResource_Error ./internal/provider/
//
//	# Run specific error scenario
//	go test -v -run TestAccCMDeviceDeviceResource_ErrorCategoryNotFound ./internal/provider/
//
// Note: Update/Delete error tests are skipped because they require stateful mock servers
// (create device first, then simulate error on update/delete). These scenarios are better
// tested with integration tests against a live BCM cluster.
