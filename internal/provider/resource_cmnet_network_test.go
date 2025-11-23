// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccCMNetNetwork_Basic(t *testing.T) {
	networkName := generateUniqueTestName("test-network")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMNetNetworkDestroy,
		Steps: []resource.TestStep{
			// Create and Read testing
			{
				Config: testAccCMNetNetworkConfigBasic(networkName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
					resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "id"),
					resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "domain_name", "cluster.local"),
				),
			},
			// ImportState testing
			{
				Config:            testAccCMNetNetworkConfigBasic(networkName),
				ResourceName:      "bcm_cmnet_network.test",
				ImportState:       true,
				ImportStateVerify: true,
			},
		},
	})
}

func TestAccCMNetNetwork_Complete(t *testing.T) {
	networkName := generateUniqueTestName("test-network")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMNetNetworkDestroy,
		Steps: []resource.TestStep{
			// Create with all attributes
			{
				Config: testAccCMNetNetworkConfigComplete(networkName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "subnet", "192.168.100.0/24"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "gateway", "192.168.100.1"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "domain_name", "test.local"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_range_start", "192.168.100.100"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_range_end", "192.168.100.200"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "dhcp_enabled", "true"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "notes", "Test network"),
					resource.TestCheckResourceAttrSet("bcm_cmnet_network.test", "uuid"),
				),
			},
		},
	})
}

func TestAccCMNetNetwork_Update(t *testing.T) {
	networkName := generateUniqueTestName("test-network")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		CheckDestroy:             testAccCheckCMNetNetworkDestroy,
		Steps: []resource.TestStep{
			// Create with initial config
			{
				Config: testAccCMNetNetworkConfigBasic(networkName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
				),
			},
			// Update MTU and notes
			{
				Config: testAccCMNetNetworkConfigUpdate(networkName),
				Check: resource.ComposeAggregateTestCheckFunc(
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "name", networkName),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "mtu", "9000"),
					resource.TestCheckResourceAttr("bcm_cmnet_network.test", "notes", "Updated notes"),
				),
			},
		},
	})
}

func testAccCheckCMNetNetworkDestroy(s *terraform.State) error {
	client := createTestBCMClient(&testing.T{})
	ctx := context.Background()

	var errors []string
	resourceCount := 0

	for _, rs := range s.RootModule().Resources {
		if rs.Type != "bcm_cmnet_network" {
			continue
		}

		resourceCount++
		id := rs.Primary.ID

		// Verify network deleted with exponential backoff
		deleted := verifyResourceDeleted(
			ctx,
			client,
			"cmnet",
			"getNetwork",
			id,
			4, // retry count
		)

		if !deleted {
			errors = append(errors, fmt.Sprintf(
				"Network still exists after destroy. Type: %s, ID: %s, Retries: 4",
				rs.Type,
				id,
			))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("CheckDestroy failures:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

func testAccCMNetNetworkConfigBasic(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name        = %[4]q
  domain_name = "cluster.local"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

func testAccCMNetNetworkConfigComplete(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name             = %[4]q
  subnet           = "192.168.100.0/24"
  gateway          = "192.168.100.1"
  mtu              = 9000
  domain_name      = "test.local"
  dhcp_range_start = "192.168.100.100"
  dhcp_range_end   = "192.168.100.200"
  notes            = "Test network"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}

func testAccCMNetNetworkConfigUpdate(name string) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

resource "bcm_cmnet_network" "test" {
  name        = %[4]q
  domain_name = "cluster.local"
  mtu         = 9000
  notes       = "Updated notes"
}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		name,
	)
}
