# Forward all emails to support@example.com to a webhook
resource "mailgun_route" "support_webhook" {
  expression  = "match_recipient('support@example.com')"
  description = "Forward support emails to webhook"
  priority    = 0

  actions = [
    "forward('https://example.com/webhooks/support')",
    "stop()"
  ]
}

# Store emails matching a pattern for later retrieval
resource "mailgun_route" "store_orders" {
  expression  = "match_recipient('orders@.*')"
  description = "Store order confirmation emails"
  priority    = 10

  actions = [
    "store(notify='https://example.com/webhooks/orders')"
  ]
}

# Forward emails to multiple recipients
resource "mailgun_route" "forward_team" {
  expression  = "match_header('subject', '.*urgent.*')"
  description = "Forward urgent emails to team"
  priority    = 5

  actions = [
    "forward('team-lead@example.com')",
    "forward('manager@example.com')"
  ]
}
