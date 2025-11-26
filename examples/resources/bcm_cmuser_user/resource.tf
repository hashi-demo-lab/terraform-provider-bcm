# Basic user creation with minimal required attributes
resource "bcm_cmuser_user" "basic" {
  username = "testuser01"
  password = var.user_password
}

# Full user configuration with all attributes
resource "bcm_cmuser_user" "developer" {
  username       = "developer01"
  password       = var.developer_password
  full_name      = "Developer One"
  surname        = "One"
  email          = "developer01@example.com"
  home_directory = "/home/developer01"
  shell          = "/bin/zsh"
  notes          = "DGX BasePOD developer account"

  authorized_ssh_keys = <<-EOT
    ssh-rsa AAAAB3NzaC1yc2EAAAA... developer01@workstation
    ssh-ed25519 AAAAC3NzaC1lZDI1NTE5AAAA... developer01@laptop
  EOT

  # Shadow password settings
  shadow_max     = 90 # Password expires after 90 days
  shadow_warning = 14 # Warn 14 days before expiration
}

# Service account with specific UID/GID
resource "bcm_cmuser_user" "service_account" {
  username       = "svc_backup"
  password       = var.service_password
  uid            = 2001
  gid            = 2001
  full_name      = "Backup Service Account"
  home_directory = "/var/lib/backup"
  shell          = "/sbin/nologin"
  notes          = "Service account for automated backups"
}

# Variables (define in terraform.tfvars or via environment)
variable "user_password" {
  description = "Password for the test user"
  type        = string
  sensitive   = true
  default     = "TestUser123!"
}

variable "developer_password" {
  description = "Password for the developer user"
  type        = string
  sensitive   = true
  default     = "DevUser123!"
}

variable "service_password" {
  description = "Password for the service account"
  type        = string
  sensitive   = true
  default     = "SvcAccount123!"
}
