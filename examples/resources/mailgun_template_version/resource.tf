# Create a new version of an existing template
resource "mailgun_template" "welcome" {
  domain      = "mail.example.com"
  name        = "welcome-email"
  description = "Welcome email template"
  template    = "<h1>Welcome!</h1>"
}

resource "mailgun_template_version" "v2" {
  domain        = mailgun_template.welcome.domain
  template_name = mailgun_template.welcome.name
  tag           = "v2"
  template      = <<-EOF
    <html>
      <body>
        <h1>Welcome, {{name}}!</h1>
        <p>We're glad to have you here.</p>
        <p>Best regards,<br/>The {{company}} Team</p>
      </body>
    </html>
  EOF
  comment       = "Updated welcome email with improved formatting"
  active        = true
}

# Create an inactive version for testing
resource "mailgun_template_version" "draft" {
  domain        = "mail.example.com"
  template_name = "notification"
  tag           = "draft"
  template      = "<p>Draft version for testing</p>"
  comment       = "Draft version - do not use in production"
  active        = false
}
