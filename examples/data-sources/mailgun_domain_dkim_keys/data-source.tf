# List all DKIM keys for a domain
data "mailgun_domain_dkim_keys" "example" {
  domain = "mail.example.com"
}

# Output the DKIM keys
output "dkim_keys" {
  value = data.mailgun_domain_dkim_keys.example.keys
}
