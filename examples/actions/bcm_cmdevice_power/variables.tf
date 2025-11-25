# Variables for BCM CMDevice Power Action Example

variable "device_uuid" {
  type        = string
  description = "Target device UUID for power operations"
  default     = ""
}

variable "device_hostname" {
  type        = string
  description = "Target device hostname for power operations"
  default     = ""
}

variable "bcm_password" {
  type        = string
  description = "BCM password (use TF_VAR_bcm_password environment variable)"
  sensitive   = true
  default     = ""
}
