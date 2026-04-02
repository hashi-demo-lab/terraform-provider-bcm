// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package main

import (
	"context"
	"fmt"
	"os"

	provider "github.com/hashi-demo-lab/terraform-provider-bcm/internal/provider"
)

func main() {
	endpoint := os.Getenv("BCM_ENDPOINT")
	username := os.Getenv("BCM_USERNAME")
	password := os.Getenv("BCM_PASSWORD")

	if endpoint == "" || username == "" || password == "" {
		fmt.Fprintln(os.Stderr, "BCM_ENDPOINT, BCM_USERNAME, and BCM_PASSWORD must be set")
		os.Exit(1)
	}

	client, err := provider.NewBCMClient(context.Background(), endpoint, username, password, true, 30)
	if err != nil {
		fmt.Fprintf(os.Stderr, "client init failed: %v\n", err)
		os.Exit(1)
	}

	for _, call := range []struct {
		service string
		method  string
	}{
		{service: "cmpart", method: "getPartitions"},
		{service: "cmpart", method: "getBasicEntityInformation"},
		{service: "cmnet", method: "getNetworks"},
		{service: "cmdevice", method: "getNodes"},
	} {
		body, err := client.CallJSONRPC(context.Background(), call.service, call.method)
		fmt.Printf("== %s.%s ==\n", call.service, call.method)
		if err != nil {
			fmt.Printf("error: %v\n", err)
			continue
		}
		fmt.Println(string(body))
	}
}
