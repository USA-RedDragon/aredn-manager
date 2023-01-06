output "public-ip" {
  value = aws_eip.ip.public_ip
}

output "key" {
  value     = tls_private_key.key.private_key_pem
  sensitive = true
}
