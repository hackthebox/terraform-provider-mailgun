# List all sending keys for a domain
data "mailgun_domain_sending_keys" "all" {
  domain = "mail.example.com"
}

output "sending_key_count" {
  value = data.mailgun_domain_sending_keys.all.total_count
}

output "sending_key_ids" {
  value = [for k in data.mailgun_domain_sending_keys.all.keys : k.id]
}
