# List all available alert event types
data "mailgun_alert_events" "available" {}

# Output the list of available event types
output "available_event_types" {
  description = "Event types that can be used when creating alerts"
  value       = data.mailgun_alert_events.available.events
}
