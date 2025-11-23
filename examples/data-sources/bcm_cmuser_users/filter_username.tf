# Find all users whose username starts with "cms"
data "bcm_cmuser_users" "cms_users" {
  username_pattern = "cms*"
}

output "cms_users" {
  description = "Users matching username pattern 'cms*'"
  value = [for user in data.bcm_cmuser_users.cms_users.users : {
    username       = user.username
    email          = user.email
    common_name    = user.common_name
    account_active = user.account_active
  }]
}
