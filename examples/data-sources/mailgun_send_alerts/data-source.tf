# List all send alerts
data "mailgun_send_alerts" "all" {}

# Output the total count
output "total_alerts" {
  value = data.mailgun_send_alerts.all.total_count
}

# Output all alert names
output "alert_names" {
  value = [for alert in data.mailgun_send_alerts.all.alerts : alert.name]
}

# Find alerts monitoring a specific metric
locals {
  bounce_alerts = [
    for alert in data.mailgun_send_alerts.all.alerts :
    alert if alert.metric == "hard_bounce_rate"
  ]
}

output "bounce_alert_names" {
  value = [for alert in local.bounce_alerts : alert.name]
}
