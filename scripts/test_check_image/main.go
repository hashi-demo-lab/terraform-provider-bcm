// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/hashi-demo-lab/terraform-provider-bcm/internal/provider"
)

func main() {
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		log.Fatal("BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set")
	}

	client, err := provider.NewBCMClient(context.Background(), endpoint, username, password, true, 30)
	if err != nil {
		log.Fatalf("Failed to create BCM client: %v", err)
	}

	// Get specific software image
	imageName := "test-basic-image"
	body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImage", imageName)
	if err != nil {
		fmt.Printf("Image '%s' not found: %v\n", imageName, err)
	} else {
		var img map[string]interface{}
		if err := json.Unmarshal(body, &img); err != nil {
			fmt.Printf("Failed to parse response: %v\n", err)
		} else {
			fmt.Printf("Found image '%s':\n", imageName)
			fmt.Printf("  Path: %s\n", img["path"])
			fmt.Printf("  UUID: %s\n", img["uuid"])
			fmt.Printf("  Kernel Version: %s\n", img["kernelVersion"])
		}
	}
}
