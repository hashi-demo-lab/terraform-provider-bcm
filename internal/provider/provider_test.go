// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/echoprovider"
)

// testAccProtoV6ProviderFactories is used to instantiate a provider during acceptance testing.
// The factory function is called for each Terraform CLI command to create a provider
// server that the CLI can connect to and interact with.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"bcm": providerserver.NewProtocol6WithError(New("test")()),
}

// testAccProtoV6ProviderFactoriesWithEcho includes the echo provider alongside the BCM provider.
// It allows for testing assertions on data returned by an ephemeral resource during Open.
// The echoprovider is used to arrange tests by echoing ephemeral data into the Terraform state.
// This lets the data be referenced in test assertions with state checks.
var testAccProtoV6ProviderFactoriesWithEcho = map[string]func() (tfprotov6.ProviderServer, error){
	"bcm":  providerserver.NewProtocol6WithError(New("test")()),
	"echo": echoprovider.NewProviderServer(),
}

func testAccPreCheck(t *testing.T) {
	// Verify required environment variables for BCM acceptance tests
	// These are used to authenticate against a real BCM instance
	// Note: Unit tests do not require these variables

	if v := os.Getenv("BCM_ENDPOINT"); v == "" {
		t.Fatal("BCM_ENDPOINT must be set for acceptance tests")
	}
	if v := os.Getenv("BCM_USERNAME"); v == "" {
		t.Fatal("BCM_USERNAME must be set for acceptance tests")
	}
	if v := os.Getenv("BCM_PASSWORD"); v == "" {
		t.Fatal("BCM_PASSWORD must be set for acceptance tests")
	}
}
