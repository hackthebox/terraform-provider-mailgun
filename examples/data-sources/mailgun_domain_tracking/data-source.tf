# Get tracking settings for a domain
data "mailgun_domain_tracking" "example" {
  domain = "mail.example.com"
}

# Output tracking status
output "click_tracking_enabled" {
  value = data.mailgun_domain_tracking.example.click_active
}

output "open_tracking_enabled" {
  value = data.mailgun_domain_tracking.example.open_active
}
