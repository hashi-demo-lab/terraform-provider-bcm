#!/bin/bash
# Import an existing BCM user into Terraform state
#
# Usage: terraform import bcm_cmuser_user.existing <username>
#
# Example:
#   terraform import bcm_cmuser_user.existing cmsupport
#
# Note: After importing, you must add the password to your Terraform
# configuration as it cannot be recovered from the BCM API.

terraform import bcm_cmuser_user.existing cmsupport
