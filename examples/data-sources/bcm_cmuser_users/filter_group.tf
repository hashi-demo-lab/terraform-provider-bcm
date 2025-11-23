# Find all users in group 1000
data "bcm_cmuser_users" "group_1000" {
  group_id = "1000"
}

output "group_1000_members" {
  description = "Usernames in group 1000"
  value       = [for user in data.bcm_cmuser_users.group_1000.users : user.username]
}

output "group_1000_details" {
  description = "Detailed user information for group 1000"
  value = [for user in data.bcm_cmuser_users.group_1000.users : {
    username       = user.username
    user_id        = user.user_id
    home_directory = user.home_directory
    login_shell    = user.login_shell
  }]
}
