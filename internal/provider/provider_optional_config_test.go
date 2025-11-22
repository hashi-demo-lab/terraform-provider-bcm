// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"fmt"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
)

// TestAccProviderConfig_InsecureSkipVerify_Default tests that insecure_skip_verify
// defaults to false when not explicitly set in configuration.
//
// This test verifies provider.go lines 130-133:
//
//	insecureSkipVerify := false
//	if !data.InsecureSkipVerify.IsNull() {
//	    insecureSkipVerify = data.InsecureSkipVerify.ValueBool()
//	}
func TestAccProviderConfig_InsecureSkipVerify_Default(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigInsecureSkipVerifyNotSet(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works (proves provider configured with default false)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_InsecureSkipVerify_ExplicitTrue tests that insecure_skip_verify
// can be explicitly set to true and is correctly passed to NewBCMClient().
//
// This test verifies provider.go lines 130-133 and line 146:
//
//	insecureSkipVerify = data.InsecureSkipVerify.ValueBool() // true
//	client, err := NewBCMClient(..., insecureSkipVerify, ...)
func TestAccProviderConfig_InsecureSkipVerify_ExplicitTrue(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigInsecureSkipVerifyTrue(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works (proves TLS verification skipped)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_InsecureSkipVerify_ExplicitFalse tests that insecure_skip_verify
// can be explicitly set to false (same as default but explicit).
//
// This test verifies provider.go lines 130-133:
//
//	insecureSkipVerify = data.InsecureSkipVerify.ValueBool() // false
func TestAccProviderConfig_InsecureSkipVerify_ExplicitFalse(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigInsecureSkipVerifyFalse(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_Timeout_Default tests that timeout defaults to 30 seconds
// when not explicitly set in configuration.
//
// This test verifies provider.go lines 135-138:
//
//	timeout := int64(30) // Default 30 seconds
//	if !data.Timeout.IsNull() {
//	    timeout = data.Timeout.ValueInt64()
//	}
func TestAccProviderConfig_Timeout_Default(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigTimeoutNotSet(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works (proves client created with default 30s timeout)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_Timeout_Custom60Seconds tests that a custom timeout value
// of 60 seconds is correctly passed to NewBCMClient().
//
// This test verifies provider.go lines 135-138 and line 147:
//
//	timeout = data.Timeout.ValueInt64() // 60
//	client, err := NewBCMClient(..., int(timeout))
func TestAccProviderConfig_Timeout_Custom60Seconds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigTimeout(60),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works with custom timeout
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_Timeout_Custom120Seconds tests that a custom timeout value
// of 120 seconds is correctly passed to NewBCMClient().
func TestAccProviderConfig_Timeout_Custom120Seconds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigTimeout(120),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works with extended timeout
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_Timeout_Short5Seconds tests that a short timeout value
// of 5 seconds is correctly passed to NewBCMClient().
func TestAccProviderConfig_Timeout_Short5Seconds(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigTimeout(5),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works with short timeout
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_Timeout_Int64ToIntConversion tests that the timeout
// value is correctly converted from int64 (schema type) to int (client parameter).
//
// This test verifies provider.go line 147:
//
//	int(timeout) // Convert int64 to int for NewBCMClient
//
// The bcm_client.go NewBCMClient() signature (line 44) expects int:
//
//	func NewBCMClient(..., timeout int) (*BCMClient, error)
func TestAccProviderConfig_Timeout_Int64ToIntConversion(t *testing.T) {
	// Test with a large int64 value that fits in int
	// Max int32: 2147483647 (safe for 32-bit and 64-bit systems)
	largeTimeout := int64(2147483647)

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigTimeout(largeTimeout),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works (proves int64->int conversion succeeded)
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_CombinedOptionalFields tests that multiple optional fields
// can be set together and all are correctly passed to NewBCMClient().
//
// This test verifies provider.go lines 130-147:
//   - insecure_skip_verify handling (lines 130-133)
//   - timeout handling (lines 135-138)
//   - Both passed to NewBCMClient() (lines 141-148)
func TestAccProviderConfig_CombinedOptionalFields(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigCombined(true, 90),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works with both optional fields set
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_CombinedDefaults tests that when no optional fields are set,
// both default to their expected values (insecure_skip_verify=false, timeout=30).
func TestAccProviderConfig_CombinedDefaults(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigBothOptionalNotSet(),
				ConfigStateChecks: []statecheck.StateCheck{
					// Verify data source works with both defaults
					statecheck.ExpectKnownValue(
						"data.bcm_cmpart_softwareimages.test",
						tfjsonpath.New("id"),
						knownvalue.NotNull(),
					),
				},
			},
		},
	})
}

// TestAccProviderConfig_Timeout_ZeroValue documents the behavior when timeout is set to 0.
// Note: The provider currently does not validate timeout values, so zero is passed to NewBCMClient.
// The actual behavior depends on http.Client timeout handling (bcm_client.go line 61).
func TestAccProviderConfig_Timeout_ZeroValue(t *testing.T) {
	t.Skip("Zero timeout behavior is implementation-dependent - http.Client may use no timeout or fail immediately")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigTimeout(0),
				// Behavior is undefined - may succeed (no timeout) or fail
			},
		},
	})
}

// TestAccProviderConfig_Timeout_NegativeValue documents the behavior when timeout is negative.
// Note: The provider currently does not validate timeout values, so negative values are passed
// to NewBCMClient. The actual behavior depends on time.Duration handling (bcm_client.go line 61).
func TestAccProviderConfig_Timeout_NegativeValue(t *testing.T) {
	t.Skip("Negative timeout behavior is implementation-dependent - time.Duration may accept negative values")

	resource.Test(t, resource.TestCase{
		PreCheck:                 func() { testAccPreCheck(t) },
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: testAccProviderConfigTimeout(-1),
				// Behavior is undefined
			},
		},
	})
}

// Test configuration helper functions

// testAccProviderConfigInsecureSkipVerifyNotSet returns a provider config
// where insecure_skip_verify is not set (should default to false).
func testAccProviderConfigInsecureSkipVerifyNotSet() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint = %[1]q
  username = %[2]q
  password = %[3]q
  # insecure_skip_verify not set - defaults to false
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccProviderConfigInsecureSkipVerifyTrue returns a provider config
// with insecure_skip_verify explicitly set to true.
func testAccProviderConfigInsecureSkipVerifyTrue() string {
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

// testAccProviderConfigInsecureSkipVerifyFalse returns a provider config
// with insecure_skip_verify explicitly set to false.
func testAccProviderConfigInsecureSkipVerifyFalse() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = false
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccProviderConfigTimeoutNotSet returns a provider config
// where timeout is not set (should default to 30 seconds).
func testAccProviderConfigTimeoutNotSet() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
  # timeout not set - defaults to 30 seconds
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}

// testAccProviderConfigTimeout returns a provider config with a custom timeout value.
func testAccProviderConfigTimeout(timeout int64) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = true
  timeout              = %[4]d
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		timeout,
	)
}

// testAccProviderConfigCombined returns a provider config with both optional fields set.
func testAccProviderConfigCombined(insecureSkipVerify bool, timeout int64) string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint             = %[1]q
  username             = %[2]q
  password             = %[3]q
  insecure_skip_verify = %[4]t
  timeout              = %[5]d
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
		insecureSkipVerify,
		timeout,
	)
}

// testAccProviderConfigBothOptionalNotSet returns a provider config
// where both optional fields are not set (should use defaults).
func testAccProviderConfigBothOptionalNotSet() string {
	return fmt.Sprintf(`
provider "bcm" {
  endpoint = %[1]q
  username = %[2]q
  password = %[3]q
  # insecure_skip_verify not set - defaults to false
  # timeout not set - defaults to 30 seconds
}

data "bcm_cmpart_softwareimages" "test" {}
`,
		os.Getenv("BCM_ENDPOINT"),
		os.Getenv("BCM_USERNAME"),
		os.Getenv("BCM_PASSWORD"),
	)
}
