output "smtp_host" {
  description = "SMTP server hostname"
  value       = var.mailgun_region == "EU" ? "smtp.eu.mailgun.org" : "smtp.mailgun.org"
}

output "smtp_port" {
  description = "SMTP server port"
  value       = "587"
}

output "smtp_username" {
  description = "SMTP username (full login)"
  value       = mailgun_smtp_credential.app.full_login
}

output "smtp_password" {
  description = "SMTP password"
  value       = random_password.smtp.result
  sensitive   = true
}

output "credential_created_at" {
  description = "When the credential was created"
  value       = mailgun_smtp_credential.app.created_at
}

output "all_credentials_count" {
  description = "Total number of SMTP credentials for this domain"
  value       = data.mailgun_smtp_credentials.all.total_count
}
