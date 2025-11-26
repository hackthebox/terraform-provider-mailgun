terraform {
  required_providers {
    mailgun = {
      source = "hackthebox/mailgun"
    }
  }
}

provider "mailgun" {
  api_key = var.mailgun_api_key
  region  = var.mailgun_region
}

# Example 1: Create a basic domain
resource "mailgun_domain" "example" {
  name        = var.domain_name
  spam_action = "disabled"
  wildcard    = false
}

# Example 2: Create a domain with custom configuration
resource "mailgun_domain" "advanced" {
  name                          = var.advanced_domain_name
  spam_action                   = "tag"
  wildcard                      = true
  use_automatic_sender_security = true
  force_dkim_authority          = true
  dkim_key_size                 = "2048"
  web_scheme                    = "https"
}

# Example 3: Data source to lookup a SINGLE domain by name
data "mailgun_domain" "lookup" {
  name = mailgun_domain.example.name
}

# Example 4: Data source to list ALL domains in your account
data "mailgun_domains" "all" {
  # Optional: limit the number of domains returned
  limit = 50
}

# Example 5: Data source with filtering (search parameter)
data "mailgun_domains" "filtered" {
  limit  = 10
  search = "example"
}
