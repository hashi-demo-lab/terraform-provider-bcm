# Find specific user by UID
data "bcm_cmuser_users" "user_1000" {
  user_id = "1000"
}

output "user_1000_details" {
  description = "User details for UID 1000"
  value = length(data.bcm_cmuser_users.user_1000.users) > 0 ? {
    username            = data.bcm_cmuser_users.user_1000.users[0].username
    email               = data.bcm_cmuser_users.user_1000.users[0].email
    home_directory      = data.bcm_cmuser_users.user_1000.users[0].home_directory
    login_shell         = data.bcm_cmuser_users.user_1000.users[0].login_shell
    authorized_ssh_keys = data.bcm_cmuser_users.user_1000.users[0].authorized_ssh_keys
    account_active      = data.bcm_cmuser_users.user_1000.users[0].account_active
  } : null
}
