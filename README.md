
# Terraform Provider for Mailgun

This Terraform provider allows you to manage [Mailgun](https://www.mailgun.com/) resources through Terraform. It provides the ability to create, read, update, and delete Mailgun domains, SMTP credentials, and API keys.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) >= 1.24 (for development)
- A Mailgun account with an API key

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    mailgun = {
      source  = "hackthebox/mailgun"
      version = "~> 0.2"
    }
  }
}
```

## Provider Configuration

The provider needs to be configured with your Mailgun API key. You can also optionally specify the region (US or EU) and a custom API endpoint if needed.

```hcl
provider "mailgun" {
  api_key  = var.mailgun_api_key  # Required (or set MAILGUN_API_KEY env var)
  region   = "US"                 # Optional: "US" (default) or "EU"
}
```

### Configuration Parameters

| Parameter | Description | Required | Default |
|-----------|-------------|----------|---------|
| `api_key` | Your Mailgun API key. Can also be set via `MAILGUN_API_KEY` environment variable | Yes | - |
| `region` | The Mailgun region (`US` or `EU`) | No | `US` |
| `endpoint` | Custom Mailgun API endpoint (overrides region) | No | - |

## Resources

### `mailgun_domain`

Manages a Mailgun domain.

```hcl
resource "mailgun_domain" "example" {
  name                          = "mail.example.com"
  spam_action                   = "tag"
  wildcard                      = false
  use_automatic_sender_security = true
  dkim_key_size                 = "2048"
  web_scheme                    = "https"
}
```

### `mailgun_smtp_credential`

Manages SMTP credentials for sending email via SMTP.

```hcl
resource "mailgun_smtp_credential" "app" {
  domain   = mailgun_domain.example.name
  login    = "app-sender"
  password = var.smtp_password
}

# The full SMTP login will be: app-sender@mail.example.com
output "smtp_login" {
  value = mailgun_smtp_credential.app.full_login
}
```

### `mailgun_api_key`

Manages Mailgun API keys for programmatic access.

```hcl
resource "mailgun_api_key" "sending" {
  role        = "sending"
  description = "API key for sending emails"
  domain_name = mailgun_domain.example.name
}

# Store the secret in Vault or another secrets manager
output "api_key_secret" {
  value     = mailgun_api_key.sending.secret
  sensitive = true
}
```

## Data Sources

### `mailgun_domains` / `mailgun_domain`

Query existing domains.

```hcl
# List all domains
data "mailgun_domains" "all" {}

# Get a specific domain
data "mailgun_domain" "example" {
  name = "mail.example.com"
}
```

### `mailgun_smtp_credentials` / `mailgun_smtp_credential`

Query existing SMTP credentials.

```hcl
# List all SMTP credentials for a domain
data "mailgun_smtp_credentials" "all" {
  domain = "mail.example.com"
}
```

### `mailgun_api_keys` / `mailgun_api_key`

Query existing API keys.

```hcl
# List all API keys
data "mailgun_api_keys" "all" {}
```

## Complete Example with Vault Integration

A common use case is to create credentials and store them in HashiCorp Vault:

```hcl
terraform {
  required_providers {
    mailgun = {
      source  = "hackthebox/mailgun"
      version = "~> 0.2"
    }
    vault = {
      source  = "hashicorp/vault"
      version = "~> 4.0"
    }
  }
}

provider "mailgun" {
  api_key = var.mailgun_api_key
  region  = "EU"
}

# Create domain
resource "mailgun_domain" "app" {
  name        = "mail.myapp.com"
  spam_action = "tag"
}

# Create SMTP credential
resource "mailgun_smtp_credential" "app" {
  domain   = mailgun_domain.app.name
  login    = "app-mailer"
  password = random_password.smtp.result
}

resource "random_password" "smtp" {
  length  = 32
  special = false
}

# Create API key for sending
resource "mailgun_api_key" "app" {
  role        = "sending"
  description = "MyApp sending key"
  domain_name = mailgun_domain.app.name
}

# Store credentials in Vault
resource "vault_kv_secret_v2" "mailgun" {
  mount = "secret"
  name  = "myapp/mailgun"

  data_json = jsonencode({
    smtp_host     = "smtp.eu.mailgun.org"
    smtp_port     = "587"
    smtp_username = mailgun_smtp_credential.app.full_login
    smtp_password = random_password.smtp.result
    api_key       = mailgun_api_key.app.secret
  })
}
```

## Development

### Building the Provider

```shell
git clone https://github.com/hackthebox/terraform-provider-mailgun.git
cd terraform-provider-mailgun
make build
```

### Running Tests

```shell
# Unit tests
make test

# Acceptance tests (requires MAILGUN_API_KEY)
export MAILGUN_API_KEY="your-api-key"
make testacc
```

### Local Installation

```shell
make install
```

## License

This provider is licensed under the [Mozilla Public License v2.0](LICENSE).
