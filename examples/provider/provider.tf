terraform {
  required_version = ">= 1.5.0"

  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

# Production-ready provider configuration using environment variables
# Set these in your environment:
#   export BCM_ENDPOINT="https://bcm.example.com:8081"
#   export BCM_USERNAME="automation-user"
#   export BCM_PASSWORD="your-secure-password"
provider "bcm" {
  # Use environment variables for sensitive values (best practice)
  endpoint             = var.bcm_endpoint
  username             = var.bcm_username
  password             = var.bcm_password
  insecure_skip_verify = var.insecure_skip_verify
  timeout              = var.bcm_timeout
}

# Variables for provider configuration
variable "bcm_endpoint" {
  description = "BCM API endpoint URL"
  type        = string
  default     = "https://172.21.15.254:8081"
}

variable "bcm_username" {
  description = "BCM username for authentication"
  type        = string
  sensitive   = true
}

variable "bcm_password" {
  description = "BCM password for authentication"
  type        = string
  sensitive   = true
}

variable "insecure_skip_verify" {
  description = "Skip TLS certificate verification (only for self-signed certs in dev/test)"
  type        = bool
  default     = false
}

variable "bcm_timeout" {
  description = "API timeout in seconds"
  type        = number
  default     = 30
}
