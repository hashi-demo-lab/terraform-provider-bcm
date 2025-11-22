// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"encoding/json"
	"fmt"
	"os"
)

// Import the BCM client from the provider
// This script will be run from the repo root

// Known BCM services based on codebase analysis
var knownServices = []string{
	"cmdevice",
	"cmpart",
	"cmnet",
	"cmprov",
	"cmjob",
	"cmserv",
	"cmmon",
	"cmgui",
	"cmuser",
	"cmsw",
	"cmdaemon",
}

// Known method patterns - we'll try these for each service
var methodPatterns = []string{
	// List/Get patterns
	"getNodes",
	"getCategories",
	"getCategory",
	"getDevices",
	"getDevice",
	"getDeviceDetails",
	"getSoftwareImages",
	"getSoftwareImage",
	"getModules",
	"getModule",
	"getNetworks",
	"getNetwork",
	"getRoles",
	"getRole",
	"getPartitions",
	"getPartition",
	"getJobs",
	"getJob",
	"getProvisioningStatus",
	"getServices",
	"getService",
	"getMetrics",
	"getMonitoringData",
	"getUsers",
	"getUser",
	"getGroups",
	"getGroup",
	"getSettings",
	"getSetting",
	"getConfiguration",
	"getHealthStatus",

	// Add/Create patterns
	"addCategory",
	"addDevice",
	"addSoftwareImage",
	"addModule",
	"addNetwork",
	"addRole",
	"addPartition",
	"addJob",
	"addUser",
	"addGroup",
	"addService",

	// Update patterns
	"updateCategory",
	"updateDevice",
	"updateSoftwareImage",
	"updateModule",
	"updateNetwork",
	"updateRole",
	"updatePartition",
	"updateJob",
	"updateUser",
	"updateGroup",
	"updateService",
	"updateSettings",

	// Remove/Delete patterns
	"removeCategories",
	"removeCategory",
	"removeDevices",
	"removeDevice",
	"removeSoftwareImages",
	"removeSoftwareImage",
	"removeModules",
	"removeModule",
	"removeNetworks",
	"removeNetwork",
	"removeRoles",
	"removeRole",
	"removePartitions",
	"removePartition",
	"removeJobs",
	"removeJob",
	"removeUsers",
	"removeUser",
	"removeGroups",
	"removeGroup",

	// Action patterns
	"startProvisioning",
	"stopProvisioning",
	"cancelProvisioning",
	"powerOn",
	"powerOff",
	"reboot",
	"reset",
	"clone",
	"cloneImage",
	"assignRole",
	"unassignRole",
	"startService",
	"stopService",
	"restartService",
	"runJob",
	"cancelJob",
	"pauseJob",
	"resumeJob",
}

func main() {
	fmt.Println("=== BCM API Discovery Tool ===")
	fmt.Println()

	// Check for required environment variables
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		fmt.Println("ERROR: Required environment variables not set")
		fmt.Println("Please set: BCM_ENDPOINT, BCM_USERNAME, BCM_PASSWORD")
		fmt.Println()
		fmt.Println("Example:")
		fmt.Println("  export BCM_ENDPOINT='https://172.21.15.254:8081'")
		fmt.Println("  export BCM_USERNAME='root'")
		fmt.Println("  export BCM_PASSWORD='Hashicorp123!'")
		os.Exit(1)
	}

	fmt.Printf("BCM Endpoint: %s\n", endpoint)
	fmt.Printf("Username: %s\n", username)
	fmt.Println()

	// We'll output JSON for easier parsing
	// This will be a map of service -> discovered methods
	discoveredAPIs := make(map[string][]string)

	// Note: We can't actually import from internal/provider here without causing module issues
	// So we'll output the methodology and commands that should be run

	fmt.Println("## Discovery Methodology")
	fmt.Println()
	fmt.Println("Due to Go module constraints, this discovery will be performed using the sampleRest/ scripts")
	fmt.Println("and by analyzing existing provider code patterns.")
	fmt.Println()
	fmt.Println("### Services to Explore:")
	for _, service := range knownServices {
		fmt.Printf("- %s\n", service)
		discoveredAPIs[service] = []string{}
	}
	fmt.Println()

	fmt.Println("### Method Discovery Approach:")
	fmt.Println("1. Review all sampleRest/*.py scripts for BCM API calls")
	fmt.Println("2. Extract service/call patterns from provider code")
	fmt.Println("3. Cross-reference with BCM documentation")
	fmt.Println()

	// Output JSON structure for later parsing
	output, _ := json.MarshalIndent(discoveredAPIs, "", "  ")
	fmt.Println("### Initial API Structure (JSON):")
	fmt.Println(string(output))
}
