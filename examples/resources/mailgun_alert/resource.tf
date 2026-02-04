# Alert with email notifications
resource "mailgun_alert" "email_alert" {
  event_type = "hard_bounces"
  channel    = "email"
  emails     = ["alerts@example.com", "ops@example.com"]
}

# Alert with webhook notifications
resource "mailgun_alert" "webhook_alert" {
  event_type  = "complaints"
  channel     = "webhook"
  webhook_url = "https://example.com/mailgun-alerts"
}

# Alert with Slack notifications (requires Slack integration)
resource "mailgun_alert" "slack_alert" {
  event_type = "unsubscribes"
  channel    = "slack"
  slack_ids  = ["C01234ABCDE", "C56789FGHIJ"]
}

# Using available event types from data source
data "mailgun_alert_events" "available" {}

resource "mailgun_alert" "dynamic_alert" {
  # Use the first available event type
  event_type = data.mailgun_alert_events.available.events[0]
  channel    = "email"
  emails     = ["alerts@example.com"]
}
