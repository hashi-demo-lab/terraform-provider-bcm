# Query all BCM users
data "bcm_cmuser_users" "all" {
}

output "user_count" {
  description = "Total number of users"
  value       = length(data.bcm_cmuser_users.all.users)
}

output "usernames" {
  description = "List of all usernames"
  value       = [for user in data.bcm_cmuser_users.all.users : user.username]
}

output "active_users" {
  description = "List of users with active accounts"
  value = [for user in data.bcm_cmuser_users.all.users : {
    username       = user.username
    email          = user.email
    home_directory = user.home_directory
    login_shell    = user.login_shell
    account_active = user.account_active
  } if user.account_active]
}
