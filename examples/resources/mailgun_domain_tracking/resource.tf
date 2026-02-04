# Configure tracking for a domain
resource "mailgun_domain_tracking" "example" {
  domain = "mail.example.com"

  # Enable click and open tracking
  click_active = true
  open_active  = true

  # Enable unsubscribe with custom footer
  unsubscribe_active      = true
  unsubscribe_html_footer = "<p>Click <a href='%%unsubscribe_url%%'>here</a> to unsubscribe.</p>"
  unsubscribe_text_footer = "Click %%unsubscribe_url%% to unsubscribe."
}

# Minimal tracking configuration
resource "mailgun_domain_tracking" "minimal" {
  domain       = "notifications.example.com"
  click_active = true
  open_active  = true
}
