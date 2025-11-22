// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

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

	// Get all software images
	body, err := client.CallJSONRPC(context.Background(), "CMPart", "getSoftwareImages")
	if err != nil {
		log.Fatalf("Failed to get software images: %v", err)
	}

	var images []map[string]interface{}
	if err := json.Unmarshal(body, &images); err != nil {
		log.Fatalf("Failed to parse response: %v", err)
	}

	fmt.Printf("Found %d software images\n", len(images))
	for _, img := range images {
		name, ok := img["name"].(string)
		if !ok {
			continue
		}
		if strings.HasPrefix(name, "test-") {
			uuid, ok := img["uuid"].(string)
			if !ok {
				fmt.Printf("Skipping %s: invalid UUID\n", name)
				continue
			}
			fmt.Printf("Deleting test image: %s (UUID: %s)\n", name, uuid)
			_, err := client.CallJSONRPC(context.Background(), "CMPart", "removeSoftwareImage", uuid, false, false, false)
			if err != nil {
				fmt.Printf("  Failed to delete: %v\n", err)
			} else {
				fmt.Printf("  Deleted successfully\n")
			}
		}
	}
}
