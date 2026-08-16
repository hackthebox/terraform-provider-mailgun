## 1.2.0 (August 16, 2026)

ENHANCEMENTS:
* resource/mailgun_ip_allowlist: Validate IP and CIDR syntax during planning; also validate related dedicated-IP, DKIM key-size, and mailing-list email fields. ([#104](https://github.com/hackthebox/terraform-provider-mailgun/pull/104))
* resource/mailgun_send_alert, data-source/mailgun_send_alert: Internal 404 handling now uses a typed error instead of a `(nil, nil)` result, matching every other resource. Read, ImportState and the data source were already correct, so this changes no observable behavior there except that Update's post-update read-back now surfaces a 404 as a diagnostic instead of silently mapping partial state. ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))

BUG FIXES:
* resource/mailgun_domain: Stop hard-erroring when a domain is deleted out-of-band. Read now removes the resource from state on a genuine 404 (retried up to ~8s to absorb Mailgun's eventual consistency) instead of leaving `plan`/`apply` permanently stuck; a 429, 500, or transport error still surfaces as a diagnostic with state left intact. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_domain_dkim_key: Stop hard-erroring when the parent domain is deleted out-of-band. Read now removes the resource from state on a genuine 404 from the domain-scoped key listing instead of leaving `plan`/`apply` permanently stuck. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_domain_ip: Stop hard-erroring when the parent domain is deleted out-of-band. Read now removes the resource from state on a genuine 404 from the domain-scoped IP listing instead of leaving `plan`/`apply` permanently stuck. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_domain_sending_key: Read no longer treats a failed API-key listing (500, rate limit, transport error) the same as the key being deleted. A listing failure during `terraform plan` now surfaces as a diagnostic with state left intact instead of silently dropping a live key from state. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_domain_tracking: Stop hard-erroring when the parent domain is deleted out-of-band. Read now removes the resource from state on a genuine 404 instead of leaving `plan`/`apply` permanently stuck. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_ip_allowlist: Read no longer mistakes a failed allowlist listing (a 500, a 429, or a transport error) for a deleted entry and drops a live resource from state. A missing entry is now determined by scanning a successful listing, so a lookup failure surfaces as a diagnostic instead. ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))
* resource/mailgun_ip_allowlist: Delete no longer silently ignores a real delete failure whose error message happens to contain "404" or "not found"; only a genuine 404 status is now treated as "already deleted". ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))
* resource/mailgun_mailing_list: Read no longer mistakes a 500 (or any error whose message happens to contain "404" or "not found") for a missing mailing list, which previously could drop a live resource from state. Detection is now a typed 404 check. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_mailing_list: Delete no longer silently ignores a delete failure whose error message happens to contain "404" or "not found"; only a genuine 404 is now treated as "already deleted". ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))
* resource/mailgun_mailing_list_member: Read no longer mistakes a 500 (or any error whose message happens to contain "404" or "not found") for a missing member, which previously could drop a live resource from state. Detection is now a typed 404 check. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_mailing_list_member: Delete no longer silently ignores a delete failure whose error message happens to contain "404" or "not found"; only a genuine 404 is now treated as "already deleted". ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))
* resource/mailgun_route: Stop hard-erroring when a route is deleted out-of-band. Read now removes the resource from state on a genuine 404 instead of leaving `plan`/`apply` permanently stuck. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_send_alert: Delete no longer hard-fails when the alert was already deleted out of band; a genuine 404 is now treated as "already deleted", matching every other resource's Delete guard. ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))
* resource/mailgun_smtp_credential: Read no longer treats a failed credential listing (500, rate limit, transport error) the same as the credential being deleted. A listing failure during `terraform plan` now surfaces as a diagnostic with state left intact instead of silently dropping a live credential from state. Separately, a genuine 404 from the domain-scoped listing (the parent domain was deleted out-of-band) now removes the resource from state instead of leaving `plan`/`apply` permanently stuck. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_template: Read no longer mistakes a 500 (or any error whose message happens to contain "404" or "not found") for a missing template, which previously could drop a live resource from state. Detection is now a typed 404 check. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_template: Delete no longer silently ignores a delete failure whose error message happens to contain "404" or "not found"; only a genuine 404 is now treated as "already deleted". ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))
* resource/mailgun_template_version: Read no longer mistakes a 500 (or any error whose message happens to contain "404" or "not found") for a missing template version, which previously could drop a live resource from state. Detection is now a typed 404 check. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_template_version: Delete no longer silently ignores a delete failure whose error message happens to contain "404" or "not found"; only a genuine 404 is now treated as "already deleted". The dedicated diagnostic for deleting an active version is unchanged. ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))
* resource/mailgun_webhook: Read no longer mistakes a 500 (or any error whose message happens to contain "404" or "not found") for a missing webhook, which previously could drop a live resource from state. Detection is now a typed 404 check, plus one narrow check for the SDK's own no-URLs sentinel on an empty webhook. ([#106](https://github.com/hackthebox/terraform-provider-mailgun/pull/106))
* resource/mailgun_webhook: Delete no longer silently ignores a delete failure whose error message happens to contain "404" or "not found"; only a genuine 404 is now treated as "already deleted". ([#108](https://github.com/hackthebox/terraform-provider-mailgun/pull/108))

## 1.1.1 (August 14, 2026)

NOTES:
* provider: Built with Go 1.25.13 (was 1.25.8) and `golang.org/x/text` v0.39.0, patching six vulnerabilities reachable from the provider's own call graph: GO-2026-6218, GO-2026-6090, GO-2026-5972, GO-2026-5970, GO-2026-5856 and GO-2026-5026. Releases up to and including v1.1.0 are affected. Building the provider from source now requires Go 1.25.13 or later. ([#97](https://github.com/hackthebox/terraform-provider-mailgun/pull/97))

ENHANCEMENTS:
* provider: Upgraded the Mailgun SDK (`mailgun-go/v5`) from v5.16.0 to v5.19.1. ([#94](https://github.com/hackthebox/terraform-provider-mailgun/pull/94))

## 1.1.0 (August 14, 2026)

DEPRECATIONS:
* resource/mailgun_smtp_credential: The `password` argument is deprecated in favor of `password_wo`/`password_wo_version` and will be removed in a future major release. ([#76](https://github.com/hackthebox/terraform-provider-mailgun/pull/76))

ENHANCEMENTS:
* provider: The `api_key` argument is now optional and falls back to the `MAILGUN_API_KEY` environment variable, matching the documented behaviour. Configuring `api_key` with a value that is unknown until apply is now an error rather than a silent fall back to the environment. ([#77](https://github.com/hackthebox/terraform-provider-mailgun/pull/77))
* resource/mailgun_smtp_credential: Add write-only `password_wo` and `password_wo_version` arguments. The secret is never stored in Terraform state; increment the version to rotate. Requires Terraform CLI >= 1.11. ([#76](https://github.com/hackthebox/terraform-provider-mailgun/pull/76))

BUG FIXES:
* provider: Retry requests that Mailgun rejects with HTTP 429, honouring `Retry-After` (capped at 30s) with exponential backoff. Previously a rate-limited response surfaced as a hard apply failure. ([#84](https://github.com/hackthebox/terraform-provider-mailgun/pull/84))
* resource/mailgun_domain_tracking: Stop overwriting configured attributes with the post-apply read. A read that disagrees with the configuration now surfaces as drift on the next refresh instead of failing the apply with "Provider produced inconsistent result after apply". ([#89](https://github.com/hackthebox/terraform-provider-mailgun/pull/89))
* resource/mailgun_smtp_credential: Stop stamping a client-side `created_at` when the credential listing lags a create. The value is now retried briefly and left null if still unavailable, so it can no longer disagree with the server. ([#88](https://github.com/hackthebox/terraform-provider-mailgun/pull/88))

## 1.0.6 (June 12, 2026)

BUG FIXES:
* resource/mailgun_domain: Fixed spurious domain replacement on in-place updates. The `spam_action`, `wildcard`, `force_dkim_authority`, and `dkim_key_size` attributes (Optional+Computed+RequiresReplace) lacked `UseStateForUnknown`, so any update to an existing domain planned them as "known after apply" and forced a full destroy/recreate. ([#61](https://github.com/hackthebox/terraform-provider-mailgun/pull/61))
* resource/mailgun_domain: Retry transient 404s when reading a domain to tolerate Mailgun's eventual consistency (a GET can 404 immediately after a successful create/update), instead of failing the read. ([#62](https://github.com/hackthebox/terraform-provider-mailgun/pull/62))

## 1.0.5 (June 1, 2026)

BUG FIXES:
* provider: Fixed inconsistent result on applying when `use_automatic_sender_security = true` ([#53](https://github.com/hackthebox/terraform-provider-mailgun/pull/53))

## 1.0.4 (April 6, 2026)

BUG FIXES:
* resource/mailgun_smtp_credential: Fixed imported SMTP credentials so they no longer require setting a synthetic password in configuration ([#51](https://github.com/hackthebox/terraform-provider-mailgun/pull/51))

## 1.0.3 (March 13, 2026)

ENHANCEMENTS:
* data-source/mailgun_domain: Improved `authentication_dns_records` attribute descriptions documenting DMARC-specific behavior ([#50](https://github.com/hackthebox/terraform-provider-mailgun/pull/50))
* resource/mailgun_domain: `authentication_dns_records.valid` now reflects actual DNS configuration status from Mailgun DMARC API ([#50](https://github.com/hackthebox/terraform-provider-mailgun/pull/50))

BUG FIXES:
* resource/mailgun_domain: Fixed `authentication_dns_records` to fetch DMARC records from correct Mailgun API endpoint (`GET /v1/dmarc/records/{domain}`) ([#50](https://github.com/hackthebox/terraform-provider-mailgun/pull/50))

## 1.0.2 (March 13, 2026)

ENHANCEMENTS:
* data-source/mailgun_domain: Added `authentication_dns_records` computed attribute exposing Mailgun-generated DMARC records ([#49](https://github.com/hackthebox/terraform-provider-mailgun/pull/49))
* resource/mailgun_domain: Added `authentication_dns_records` computed attribute exposing Mailgun-generated DMARC records ([#49](https://github.com/hackthebox/terraform-provider-mailgun/pull/49))

## 1.0.1 (February 6, 2026)

BUG FIXES:

* provider: Fixed EU region configuration by removing version suffix from API base URL

## 1.0.0 (February 5, 2026)

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
