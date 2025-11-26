# Example: Managing API Keys with Mailgun Provider
#
# This example demonstrates how to create API keys for Mailgun
# and optionally store them in HashiCorp Vault.

terraform {
  required_providers {
    mailgun = {
      source = "hackthebox/mailgun"
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

# Create the domain first (if not already exists)
resource "mailgun_domain" "main" {
  name        = var.domain_name
  spam_action = "disabled"
}

# Create a domain-specific sending key (recommended for production)
resource "mailgun_api_key" "sending" {
  role        = "sending"
  kind        = "domain"
  domain_name = mailgun_domain.main.name
  description = "Sending key for ${var.app_name}"
}

# Example: Create an admin key (use with caution!)
# resource "mailgun_api_key" "admin" {
#   role        = "admin"
#   description = "Admin key for infrastructure management"
# }

# Example: Create a key with expiration
# resource "mailgun_api_key" "temporary" {
#   role        = "sending"
#   kind        = "domain"
#   domain_name = mailgun_domain.main.name
#   description = "Temporary sending key"
#   expiration  = 86400 # 24 hours in seconds
# }

# Look up an existing API key by ID
# data "mailgun_api_key" "existing" {
#   id = "your-key-id-here"
# }

# List all API keys
data "mailgun_api_keys" "all" {}

# List domain-specific keys only
data "mailgun_api_keys" "domain_keys" {
  kind = "domain"
}

# ============================================================================
# Vault Integration (Uncomment to use)
# ============================================================================

# Store API key in Vault for applications to consume
# resource "vault_kv_secret_v2" "mailgun_api" {
#   mount = "secret"
#   name  = "${var.app_name}/mailgun"
#
#   data_json = jsonencode({
#     api_key     = mailgun_api_key.sending.secret
#     domain      = mailgun_domain.main.name
#     region      = var.mailgun_region
#     api_base    = var.mailgun_region == "EU" ? "https://api.eu.mailgun.org/v3" : "https://api.mailgun.org/v3"
#   })
# }
