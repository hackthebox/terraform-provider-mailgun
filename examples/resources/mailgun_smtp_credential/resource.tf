# Create an SMTP credential for a domain
resource "mailgun_smtp_credential" "app" {
  domain   = "mail.example.com"
  login    = "app-mailer"
  password = var.smtp_password
}

# The full SMTP login will be: app-mailer@mail.example.com
output "smtp_full_login" {
  value = mailgun_smtp_credential.app.full_login
}

# Write-only password (recommended). Requires Terraform CLI >= 1.11.
# The secret is never written to Terraform state. Bump password_wo_version to rotate.
resource "random_password" "smtp" {
  length  = 24
  special = false
}

resource "mailgun_smtp_credential" "app_wo" {
  domain              = "mail.example.com"
  login               = "app-mailer-wo"
  password_wo         = random_password.smtp.result
  password_wo_version = 1
}
