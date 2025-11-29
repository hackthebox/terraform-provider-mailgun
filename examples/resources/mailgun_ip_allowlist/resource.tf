# Allow a single IP address
resource "mailgun_ip_allowlist" "office" {
  address     = "203.0.113.50"
  description = "Office network"
}

# Allow a CIDR range
resource "mailgun_ip_allowlist" "vpn" {
  address     = "198.51.100.0/24"
  description = "VPN server range"
}

# Allow CI/CD runner IP
resource "mailgun_ip_allowlist" "github_actions" {
  address     = "192.0.2.100"
  description = "GitHub Actions runner"
}
