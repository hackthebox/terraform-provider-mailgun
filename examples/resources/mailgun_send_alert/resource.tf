# Basic send alert - Monitor hard bounce rate by domain
resource "mailgun_send_alert" "bounce_alert" {
  name       = "high-bounce-rate-alert"
  metric     = "hard_bounce_rate"
  comparator = ">"
  limit      = "0.05"
  dimension  = "domain"
}

# Send alert with email notifications
resource "mailgun_send_alert" "delivery_alert" {
  name           = "low-delivery-rate-alert"
  metric         = "delivered_rate"
  comparator     = "<"
  limit          = "0.95"
  dimension      = "domain"
  description    = "Alert when delivery rate drops below 95%"
  period         = "1d"
  alert_channels = ["email"]
}

# Send alert with multiple notification channels
resource "mailgun_send_alert" "complaint_alert" {
  name           = "complaint-rate-alert"
  metric         = "complained_rate"
  comparator     = ">="
  limit          = "0.001"
  dimension      = "ip"
  description    = "Alert when complaint rate reaches 0.1%"
  period         = "1h"
  alert_channels = ["email", "webhook"]
}

# Send alert with filters
resource "mailgun_send_alert" "filtered_alert" {
  name       = "specific-domain-bounce-alert"
  metric     = "hard_bounce_rate"
  comparator = ">"
  limit      = "0.10"
  dimension  = "domain"
  period     = "1d"

  filters {
    dimension = "domain"
    values    = ["mail.example.com", "notifications.example.com"]
  }
}
