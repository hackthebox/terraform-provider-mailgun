# Look up an existing alert by ID
data "mailgun_alert" "example" {
  id = "12345678-1234-5678-1234-123456789abc"
}

# Use the alert information
output "alert_channel" {
  value = data.mailgun_alert.example.channel
}

output "alert_event_type" {
  value = data.mailgun_alert.example.event_type
}
