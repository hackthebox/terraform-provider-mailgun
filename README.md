# Terraform Provider for Mailgun

This Terraform provider allows you to manage [Mailgun](https://www.mailgun.com/) resources through Terraform: domains and their DNS, DKIM, IP and tracking configuration, SMTP credentials, domain sending keys, routes, webhooks, templates, mailing lists, IP allowlists and send alerts.

The examples below cover the most common resources. For the full, authoritative reference, see the [provider documentation on the Terraform Registry](https://registry.terraform.io/providers/hackthebox/mailgun/latest/docs).

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/downloads) >= 1.0
- [Go](https://golang.org/doc/install) at the version in the `go` directive of [`go.mod`](go.mod), for development only
- A Mailgun account, and an **account-wide API key** for that account

The provider authenticates with an account-wide key. A domain-scoped sending key cannot configure it, because most of what the provider manages (domains, subaccounts, the account-level IP allowlist) is not scoped to a single domain. Create the key in the Mailgun dashboard before running Terraform; the provider does not manage account-level keys, and could not, since it authenticates with the very credential such a resource would manage.

## Installation

Add the provider to your Terraform configuration:

```hcl
terraform {
  required_providers {
    mailgun = {
      source  = "hackthebox/mailgun"
      version = "~> 1.1"
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
| `api_key` | Your account-wide Mailgun API key. Can also be set via `MAILGUN_API_KEY` environment variable | No (if `MAILGUN_API_KEY` is set) | - |
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
  domain              = mailgun_domain.example.name
  login               = "app-sender"
  password_wo         = var.smtp_password
  password_wo_version = 1
}

# password_wo is never stored in state. Increment password_wo_version to rotate it.
# Requires Terraform >= 1.11; the older `password` argument is deprecated.

# The full SMTP login will be: app-sender@mail.example.com
output "smtp_login" {
  value = mailgun_smtp_credential.app.full_login
}
```

### `mailgun_domain_sending_key`

Manages a domain-scoped sending key.

```hcl
resource "mailgun_domain_sending_key" "sending" {
  domain      = mailgun_domain.example.name
  description = "Sending key for the app"
}

# Store the secret in Vault or another secrets manager
output "sending_key_secret" {
  value     = mailgun_domain_sending_key.sending.secret
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

### `mailgun_domain_sending_keys`

Query existing domain sending keys.

```hcl
data "mailgun_domain_sending_keys" "all" {
  domain = "mail.example.com"
}
```

## Complete Example with Vault Integration

A common use case is to create credentials and store them in HashiCorp Vault:

```hcl
terraform {
  required_providers {
    mailgun = {
      source  = "hackthebox/mailgun"
      version = "~> 1.1"
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
  domain              = mailgun_domain.app.name
  login               = "app-mailer"
  password_wo         = random_password.smtp.result
  password_wo_version = 1
}

resource "random_password" "smtp" {
  length  = 32
  special = false
}

# Create a domain-scoped sending key
resource "mailgun_domain_sending_key" "app" {
  domain      = mailgun_domain.app.name
  description = "MyApp sending key"
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
    sending_key   = mailgun_domain_sending_key.app.secret
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
