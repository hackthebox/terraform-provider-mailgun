# List all IPs assigned to a domain
data "mailgun_domain_ips" "example" {
  domain = "mail.example.com"
}

# Output the assigned IPs
output "domain_ips" {
  value = data.mailgun_domain_ips.example.ips
}
