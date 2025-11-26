# Example: Managing SMTP Credentials with Mailgun Provider
#
# This example demonstrates how to create SMTP credentials for a Mailgun domain
# and optionally store them in HashiCorp Vault.

terraform {
  required_providers {
    mailgun = {
      source = "hackthebox/mailgun"
    }
    random = {
      source  = "hashicorp/random"
      version = "~> 3.0"
    }
    # Uncomment to use Vault integration
    # vault = {
    #   source  = "hashicorp/vault"
    #   version = "~> 3.0"
    # }
  }
}

# Configure the Mailgun provider
provider "mailgun" {
  api_key = var.mailgun_api_key
  region  = var.mailgun_region # "US" or "EU"
}

# Generate a secure random password for SMTP
resource "random_password" "smtp" {
  length  = 32
  special = true
}

# Create the domain (if not already exists)
resource "mailgun_domain" "main" {
  name        = var.domain_name
  spam_action = "disabled"
}

# Create SMTP credential for the application
resource "mailgun_smtp_credential" "app" {
  domain   = mailgun_domain.main.name
  login    = var.smtp_login
  password = random_password.smtp.result
}

# Example: Look up an existing SMTP credential
data "mailgun_smtp_credential" "lookup" {
  domain = mailgun_domain.main.name
  login  = mailgun_smtp_credential.app.login
}

# Example: List all SMTP credentials for a domain
data "mailgun_smtp_credentials" "all" {
  domain = mailgun_domain.main.name
}

# ============================================================================
# Vault Integration (Uncomment to use)
# ============================================================================

# Store SMTP credentials in Vault for applications to consume
# resource "vault_kv_secret_v2" "smtp" {
#   mount = "secret"
#   name  = "${var.app_name}/smtp"
#
#   data_json = jsonencode({
#     host     = var.mailgun_region == "EU" ? "smtp.eu.mailgun.org" : "smtp.mailgun.org"
#     port     = "587"
#     username = mailgun_smtp_credential.app.full_login
#     password = random_password.smtp.result
#   })
# }
