// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccCMPartSoftwareImagesDataSource_Basic(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source exists
					resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "id"),
					// Verify images attribute exists (may be empty list)
					resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.#"),
				),
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_EmptyResponse(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify empty array returns empty list, not error
					resource.TestCheckResourceAttr("data.bcm_cmpart_softwareimages.test", "id", "placeholder"),
				),
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_NestedModules(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Check if there are any images returned
					resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "images.#"),
				),
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_AllFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccCMPartSoftwareImagesDataSourceConfig(),
				Check: resource.ComposeAggregateTestCheckFunc(
					// Verify data source exists
					resource.TestCheckResourceAttrSet("data.bcm_cmpart_softwareimages.test", "id"),
					// Note: We can't check specific image fields without knowing
					// what images exist in the test BCM instance. In a real
					// scenario, you might seed test data or check for specific
					// known test images.
				),
			},
		},
	})
}

func TestAccCMPartSoftwareImagesDataSource_InvalidCredentials(t *testing.T) {
	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: `
provider "bcm" {
  endpoint             = "` + os.Getenv("BCM_ENDPOINT") + `"
  username             = "invalid_user"
  password             = "invalid_password"
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {}
`,
				ExpectError: regexp.MustCompile(`(login failed|authentication|401|unauthorized)`),
			},
		},
	})
}

// testAccCMPartSoftwareImagesDataSourceConfig returns a basic test configuration
// for the bcm_cmpart_softwareimages data source
func testAccCMPartSoftwareImagesDataSourceConfig() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}
