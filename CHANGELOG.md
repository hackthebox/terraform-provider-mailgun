## 1.0.0 (Unreleased)

BREAKING CHANGES:

* resource/mailgun_api_key: Removed in favor of `mailgun_domain_sending_key` ([#16](https://github.com/hackthebox/terraform-provider-mailgun/pull/16))

FEATURES:

* **New Resource:** `mailgun_domain_sending_key` ([#16](https://github.com/hackthebox/terraform-provider-mailgun/pull/16))
* **New Resource:** `mailgun_route` ([#13](https://github.com/hackthebox/terraform-provider-mailgun/pull/13))
* **New Resource:** `mailgun_webhook` ([#14](https://github.com/hackthebox/terraform-provider-mailgun/pull/14))
* **New Resource:** `mailgun_ip_allowlist` ([#17](https://github.com/hackthebox/terraform-provider-mailgun/pull/17))
* **New Resource:** `mailgun_template` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Resource:** `mailgun_template_version` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Resource:** `mailgun_mailing_list` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Resource:** `mailgun_mailing_list_member` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Resource:** `mailgun_domain_tracking` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Resource:** `mailgun_domain_dkim_key` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Resource:** `mailgun_domain_ip` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Resource:** `mailgun_alert` ([#32](https://github.com/hackthebox/terraform-provider-mailgun/pull/32))
* **New Data Source:** `mailgun_domain_sending_keys` ([#16](https://github.com/hackthebox/terraform-provider-mailgun/pull/16))
* **New Data Source:** `mailgun_routes` ([#13](https://github.com/hackthebox/terraform-provider-mailgun/pull/13))
* **New Data Source:** `mailgun_webhooks` ([#14](https://github.com/hackthebox/terraform-provider-mailgun/pull/14))
* **New Data Source:** `mailgun_ip_allowlist` ([#17](https://github.com/hackthebox/terraform-provider-mailgun/pull/17))
* **New Data Source:** `mailgun_templates` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Data Source:** `mailgun_template_versions` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Data Source:** `mailgun_mailing_lists` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Data Source:** `mailgun_mailing_list_members` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Data Source:** `mailgun_domain_tracking` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Data Source:** `mailgun_domain_dkim_keys` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Data Source:** `mailgun_domain_ips` ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* **New Data Source:** `mailgun_alert` ([#32](https://github.com/hackthebox/terraform-provider-mailgun/pull/32))
* **New Data Source:** `mailgun_alerts` ([#32](https://github.com/hackthebox/terraform-provider-mailgun/pull/32))
* **New Data Source:** `mailgun_subaccount` ([#32](https://github.com/hackthebox/terraform-provider-mailgun/pull/32))
* **New Data Source:** `mailgun_subaccounts` ([#32](https://github.com/hackthebox/terraform-provider-mailgun/pull/32))

ENHANCEMENTS:

* provider: Add schema versioning (Version: 0) to all 14 resources to enable future state migrations ([#33](https://github.com/hackthebox/terraform-provider-mailgun/pull/33))
* provider: Add go-changelog infrastructure for structured release notes ([#34](https://github.com/hackthebox/terraform-provider-mailgun/pull/34))
* provider: Upgraded mailgun-go SDK from v5.8.1 to v5.12.0 ([#30](https://github.com/hackthebox/terraform-provider-mailgun/pull/30))
* resource/mailgun_route: Added schema-level validation for actions and priority ([#15](https://github.com/hackthebox/terraform-provider-mailgun/pull/15))
* resource/mailgun_domain: Added schema-level validation for spam_action, dkim_key_size, and web_scheme ([#15](https://github.com/hackthebox/terraform-provider-mailgun/pull/15))

BUG FIXES:

* resource/mailgun_domain: Fixed incorrect handling of wildcard, spam_action, force_dkim_authority, and dkim_key_size attributes ([#19](https://github.com/hackthebox/terraform-provider-mailgun/pull/19))
* resource/mailgun_domain: Fixed resource import by using domain name as ID ([#19](https://github.com/hackthebox/terraform-provider-mailgun/pull/19))

## 0.5.0

FEATURES:

* **New Resource:** `mailgun_alert` - Manage send alerts for email metrics
* **New Data Source:** `mailgun_alert` - Get a send alert by name
* **New Data Source:** `mailgun_alerts` - List all send alerts
* **New Data Source:** `mailgun_subaccount` - Get a subaccount by ID
* **New Data Source:** `mailgun_subaccounts` - List all subaccounts

## 0.4.0

FEATURES:

* **New Resource:** `mailgun_template` - Manage email templates
* **New Resource:** `mailgun_template_version` - Manage template versions
* **New Resource:** `mailgun_mailing_list` - Manage mailing lists
* **New Resource:** `mailgun_mailing_list_member` - Manage mailing list members
* **New Resource:** `mailgun_domain_tracking` - Manage domain tracking settings
* **New Resource:** `mailgun_domain_dkim_key` - Manage DKIM keys for domains
* **New Resource:** `mailgun_domain_ip` - Manage IP associations for domains

## 0.3.0

FEATURES:

* **New Resource:** `mailgun_route` - Manage email routing rules with expressions and actions
* **New Resource:** `mailgun_webhook` - Manage webhook configurations
* **New Resource:** `mailgun_ip_allowlist` - Manage IP allowlist entries
* **New Resource:** `mailgun_domain_sending_key` - Manage domain-scoped sending API keys
* **New Data Source:** `mailgun_routes` - List all routes
* **New Data Source:** `mailgun_webhooks` - List webhooks for a domain
* **New Data Source:** `mailgun_ip_allowlist` - List IP allowlist entries
* **New Data Source:** `mailgun_domain_sending_keys` - List domain sending keys

BREAKING CHANGES:

* resource/mailgun_api_key: Removed in favor of `mailgun_domain_sending_key`

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
