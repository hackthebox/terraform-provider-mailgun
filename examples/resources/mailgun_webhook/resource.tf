# Configure a webhook to receive delivery notifications
resource "mailgun_webhook" "delivered" {
  domain       = "mail.example.com"
  webhook_type = "delivered"

  urls = [
    "https://example.com/webhooks/delivered"
  ]
}

# Configure a webhook with multiple URLs for redundancy
resource "mailgun_webhook" "bounces" {
  domain       = "mail.example.com"
  webhook_type = "permanent_fail"

  urls = [
    "https://primary.example.com/webhooks/bounces",
    "https://backup.example.com/webhooks/bounces"
  ]
}

# Configure webhooks for tracking email engagement
resource "mailgun_webhook" "opened" {
  domain       = "mail.example.com"
  webhook_type = "opened"

  urls = [
    "https://analytics.example.com/webhooks/email-opens"
  ]
}

resource "mailgun_webhook" "clicked" {
  domain       = "mail.example.com"
  webhook_type = "clicked"

  urls = [
    "https://analytics.example.com/webhooks/email-clicks"
  ]
}

# Configure a webhook for spam complaints
resource "mailgun_webhook" "complained" {
  domain       = "mail.example.com"
  webhook_type = "complained"

  urls = [
    "https://compliance.example.com/webhooks/complaints"
  ]
}
