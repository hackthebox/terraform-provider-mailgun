# List all configured alerts
data "mailgun_alerts" "all" {}

# Output the total number of alerts
output "total_alerts" {
  value = data.mailgun_alerts.all.total_count
}

# Output all alert event types being monitored
output "monitored_events" {
  value = [for alert in data.mailgun_alerts.all.alerts : alert.event_type]
}
