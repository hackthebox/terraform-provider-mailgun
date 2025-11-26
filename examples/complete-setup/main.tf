terraform {
  required_providers {
    mailgun = {
      source = "hackthebox/mailgun"
    }
  }
}

provider "mailgun" {
  api_key = var.mailgun_api_key
  region  = "US" # or "EU"
}

# Create a domain
resource "mailgun_domain" "blog" {
  name        = var.domain_name
  spam_action = "disabled"
  wildcard    = false
}

# Create SMTP credentials
resource "mailgun_smtp_credential" "app" {
  domain_name = mailgun_domain.blog.name
  login       = "app@${var.domain_name}"
  password    = var.smtp_password
}

# Outputs
output "domain_name" {
  value       = mailgun_domain.blog.name
  description = "The created domain name"
}

output "smtp_username" {
  value       = mailgun_smtp_credential.app.login
  description = "SMTP username"
}

output "smtp_password" {
  value       = mailgun_smtp_credential.app.password
  sensitive   = true
  description = "SMTP password"
}

