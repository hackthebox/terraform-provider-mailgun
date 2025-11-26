terraform {
  required_providers {
    mailgun = {
      source = "registry.terraform.io/hackthebox/mailgun"
    }
  }
}

# Define variables for provider configuration
variable "api_key" {
  description = "Mailgun API key"
  type        = string
  sensitive   = true
}

variable "region" {
  description = "Mailgun region (US or EU)"
  type        = string
  default     = null
}

variable "endpoint" {
  description = "Mailgun API endpoint"
  type        = string
  default     = null
}

provider "mailgun" {
  api_key  = var.api_key
  region   = var.region
  endpoint = var.endpoint
}

# Retrieve domains from Mailgun
data "mailgun_domains" "this" {}

# Output the domains
output "domains" {
  value = data.mailgun_domains.this.items
}
