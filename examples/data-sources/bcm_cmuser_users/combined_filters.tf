# Use multiple filters (AND logic)
# Find users whose username starts with "cms" AND are in group 1000
data "bcm_cmuser_users" "filtered" {
  username_pattern = "cms*"
  group_id         = "1000"
}

output "filtered_users" {
  description = "Users matching combined filters"
  value = [for user in data.bcm_cmuser_users.filtered.users : {
    username       = user.username
    user_id        = user.user_id
    group_id       = user.group_id
    email          = user.email
    account_active = user.account_active
    home_directory = user.home_directory
    login_shell    = user.login_shell
  }]
}
