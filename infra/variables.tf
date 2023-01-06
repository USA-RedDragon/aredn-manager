variable "region" {
  default     = "us-east-2"
  description = "The AWS region to use for the infrastructure"
}

variable "instance-type" {
  default     = "t3a.micro"
  description = "The AWS instance type to use for the infrastructure"
}

variable "server-name" {
  default     = "KI5VMF-CLOUD-TUNNEL"
  description = "The name of the server in mesh status"
}

variable "disk-size" {
  default     = 8
  description = "The size of the disk in GB"
}

variable "configuration-json" {
  description = "The configuration JSON for the server"
  sensitive   = true
}

variable "cloudflare_api_token" {
  description = "The API token for Cloudflare"
  sensitive   = true
}

variable "domain" {
  default     = "mcswain.cloud"
  description = "The domain to use for the infrastructure"
}

variable "subdomain" {
  default     = "aredn-cloud-node"
  description = "The subdomain to use for the infrastructure"
}
