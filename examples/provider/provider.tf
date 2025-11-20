terraform {
  required_version = ">= 1.5.0"

  required_providers {
    bcm = {
      source  = "hashicorp/bcm"
      version = "~> 0.1"
    }
  }
}

provider "bcm" {
  endpoint             = "https://172.21.15.254:8081"
  username             = "root"
  password             = "Hashicorp123!"
  insecure_skip_verify = true # Required for self-signed certificates
  timeout              = 30   # Optional: API timeout in seconds (default: 30)
}
