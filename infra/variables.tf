variable "region" {
  default     = "us-east-2"
  description = "The AWS region to use for the infrastructure"
}

variable "instance-type" {
  default     = "t3a.small"
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

variable "wireguard_tap_address" {
  default     = "10.184.4.136"
  description = "The AREDN address to use for the WireGuard interface to tap into the mesh"
}

variable "wireguard_peer_publickey" {
  description = "The public key of the WireGuard peer"
  sensitive   = true
}

variable "wireguard_server_privatekey" {
  description = "The private key of the WireGuard server"
  sensitive   = true
}

variable "map-config-json" {
  sensitive   = true
  description = "The map configuration JSON for the server"
}

variable "server-gridsquare" {
  sensitive   = true
  description = "The grid square of the server"
}

variable "server-lon" {
  sensitive   = true
  description = "The longitude of the server"
}

variable "server-lat" {
  sensitive   = true
  description = "The latitude of the server"
}
