# Basic template with Handlebars engine
resource "mailgun_template" "welcome" {
  domain      = "mail.example.com"
  name        = "welcome-email"
  description = "Welcome email template for new users"
  template    = <<-EOF
    <html>
      <body>
        <h1>Welcome, {{name}}!</h1>
        <p>Thank you for joining {{company}}.</p>
      </body>
    </html>
  EOF
}

# Template with Go templating engine
resource "mailgun_template" "notification" {
  domain      = "mail.example.com"
  name        = "notification"
  description = "System notification template"
  engine      = "go"
  template    = <<-EOF
    <html>
      <body>
        <h1>{{.Title}}</h1>
        <p>{{.Message}}</p>
      </body>
    </html>
  EOF
}

# Plain text template
resource "mailgun_template" "plain_text" {
  domain      = "mail.example.com"
  name        = "plain-notification"
  description = "Plain text notification"
  template    = "Hello {{name}}, you have a new message: {{message}}"
}
