# List all webhooks for a domain
data "mailgun_webhooks" "all" {
  domain = "mail.example.com"
}

output "webhook_types" {
  value = [for w in data.mailgun_webhooks.all.webhooks : w.webhook_type]
}

output "webhook_urls" {
  value = {
    for w in data.mailgun_webhooks.all.webhooks : w.webhook_type => w.urls
  }
}
