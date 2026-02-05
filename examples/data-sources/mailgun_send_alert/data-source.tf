# Fetch a specific send alert by name
data "mailgun_send_alert" "example" {
  name = "high-bounce-rate-alert"
}

# Use the alert data
output "alert_metric" {
  value = data.mailgun_send_alert.example.metric
}

output "alert_limit" {
  value = data.mailgun_send_alert.example.limit
}
