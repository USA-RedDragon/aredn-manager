provider "cloudflare" {
  api_token = var.cloudflare_api_token
}

data "cloudflare_zone" "site-zone" {
  name = var.domain
}

resource "cloudflare_record" "record" {
  zone_id = data.cloudflare_zone.site-zone.id
  name    = var.subdomain
  value   = aws_eip.ip.public_ip
  type    = "A"
  proxied = false
}
