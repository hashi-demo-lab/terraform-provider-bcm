#!/bin/bash
# Import an existing BCM category into Terraform state

# Import by UUID (recommended - UUIDs are stable identifiers)
terraform import bcm_cmdevice_category.example "0ae6d733-3015-4479-bfab-ce2d237a2809"

# After import, run terraform plan to verify no changes are detected
terraform plan

# Example workflow:
# 1. Create a category resource configuration in your .tf file
# 2. Run the import command with the category's UUID
# 3. Verify terraform plan shows no changes
# 4. Manage the category through Terraform going forward

# To find the UUID of a category:
# - Use BCM web UI: Navigate to CMDevice > Categories > Category Details
# - Or use BCM API: curl -k -X POST https://bcm-server:8081/json \
#   -H "Cookie: cm-login-token=<token>" \
#   -d '{"service":"cmdevice","call":"getCategories"}'
