## 0.3.0 (Unreleased)

FEATURES:

* **New Resource:** `mailgun_route` - Manage email routing rules with expressions and actions
* **New Data Source:** `mailgun_routes` - List all routes

## 0.2.1

DOCUMENTATION:

* Add Example Usage sections to all resource and data source documentation
* Add Import sections with shell command examples for all resources
* Update documentation templates to support multiple import formats

## 0.2.0

FEATURES:

* **New Resource:** `mailgun_domain` - Manage Mailgun domains with full CRUD support
* **New Resource:** `mailgun_smtp_credential` - Manage SMTP credentials for domains
* **New Resource:** `mailgun_api_key` - Manage Mailgun API keys with role-based access

* **New Data Source:** `mailgun_domain` - Query a single domain by name
* **New Data Source:** `mailgun_domains` - List all domains with filtering options
* **New Data Source:** `mailgun_smtp_credential` - Query a single SMTP credential
* **New Data Source:** `mailgun_smtp_credentials` - List SMTP credentials for a domain
* **New Data Source:** `mailgun_api_key` - Query a single API key by ID
* **New Data Source:** `mailgun_api_keys` - List all API keys

ENHANCEMENTS:

* Support for both US and EU Mailgun regions via `region` provider configuration
* Custom endpoint support for advanced configurations
* Sensitive field handling for passwords and API key secrets
