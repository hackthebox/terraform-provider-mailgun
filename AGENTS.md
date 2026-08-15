# AGENTS.md

Guidance for AI coding agents working in this repository. Human contributors should start with [CONTRIBUTING.md](CONTRIBUTING.md); the conventions below apply to them too. Provider usage and build commands are in [README.md](README.md).

## What this is

A Terraform provider for Mailgun built on the Terraform Plugin Framework and the `mailgun-go/v5` SDK, published to [registry.terraform.io/hackthebox/mailgun](https://registry.terraform.io/providers/hackthebox/mailgun/latest).

Everything is **hand-written**. The provider was migrated off OpenAPI-based generation and nothing is regenerated from a spec. Write schemas, models and CRUD by hand against the SDK.

**`internal/provider/provider.go` (`Resources` and `DataSources`) is the source of truth for what the provider registers.** Do not trust a list in a doc over it. README once advertised a `mailgun_api_key` resource that was never built.

## Layout

Each resource and data source lives in its own package under `internal/provider/`, with schema and model split per side:

```
internal/provider/<package>/
  resource.go              # CRUD
  resource_schema.go
  resource_model.go
  resource_test.go         # acceptance tests
  data_source.go           # singular data source (if any)
  list_data_source.go      # plural/list data source (if any)
  data_source_schema.go
  data_source_model.go
```

`ip_allowlist/` and `send_alerts/` predate this and use flat `schema.go`/`model.go`. Follow the split convention for new work. `domains/` is the reference implementation.

Two packages under `internal/provider/` are shared helpers rather than registered resources:

- `schema_validators/` — reusable `validator.String` implementations (`IPAddress`, `IPAddressOrCIDR`, `EmailAddress`). Use these instead of hand-rolling a validator.
- `test_helpers/` — acceptance-test scaffolding (precheck, provider factory, cleanup, randomised fixtures).

## Talking to Mailgun

`provider.go` builds one `*mailgun.Client`, applies the region, and installs `retry_transport.go` as its `RoundTripper` to retry HTTP 429 with `Retry-After` and exponential backoff. Every resource receives that client via `Configure`.

`ip_allowlist/client.go` and `send_alerts/api_client.go` call Mailgun directly because the SDK has no binding for those endpoints. Both build requests from `APIBase()`/`APIKey()` and dispatch through `HTTPClient()`, so they inherit the configured region, credentials and retry behaviour. **Any hand-rolled call must do the same**; a bare `http.Client` silently loses all three. Prefer an SDK method where one exists.

## Conventions that matter

**Detecting 404s.** Use `mgerr.IsNotFound(err)` (`internal/provider/mgerr`), not error-text matching. It wraps `mailgun.GetStatusFromErr(err) == http.StatusNotFound`, which resolves through `errors.As` against `*mailgun.UnexpectedResponseError` the same way the SDK's own `RateLimitedError` does, so a `%w`-wrapped error still matches. No package should contain `strings.Contains(err.Error(), "404")` or similar; that pattern has been removed everywhere except two genuine exceptions:

- `webhooks/resource.go` `Read` also treats `"returned no urls"` as absence, because `GetWebhook` signals a missing webhook as a 200 with an empty URL list, not a 404 (`webhook_type` is `stringvalidator.OneOf`-constrained, so a real 404 there is unreachable).
- `template_versions/resource.go` `Delete` keeps a dedicated `"deleting active version is not allowed"` check: that is a genuine business-rule response, not a 404 alias, and gets its own diagnostic instead of being swallowed.

**In-memory listing scans.** A handful of lookups (`ip_allowlist.GetIPAllowlistEntry`, `domain_dkim_key.findDomainKey`, `domain_ip`'s IP scan) have no per-item endpoint: they list (a single 200 response) and scan for a match. These return `(T, bool, error)`, never a synthetic 404. `found` distinguishes a genuine scan miss from a wire-level failure of the underlying list call; `err` is reserved for that failure and must never be treated as "not found" by the caller. Don't collapse the two into a single "not found" error the way `mailgun.GetStatusFromErr` would for a real 404.

**Hand-rolled clients producing a typed 404.** `ip_allowlist/client.go` and `send_alerts/api_client.go` build requests by hand, so they never produce a real `*mailgun.UnexpectedResponseError` and `mgerr.IsNotFound` would always report false for them. Where a caller needs to branch on status (today: `ip_allowlist`'s `Delete` and `send_alerts`'s `GetSendAlert`), wrap the error in `mgerr.StatusError(msg, status)` instead of a bare `fmt.Errorf`. Its `Error()` is exactly `msg` (never the SDK's deprecated, `%#v`-dumping `UnexpectedResponseError.String()`), while `mgerr.IsNotFound`/`mailgun.GetStatusFromErr` still resolve `status` through it. Don't add the wrapper to a block whose caller never branches on status; a plain `fmt.Errorf` is enough there.

**Missing resources.** When `Read` finds the resource gone, call `resp.State.RemoveResource(ctx)` so Terraform plans a recreate. Returning an error instead leaves users with a `plan` that fails until they hand-run `terraform state rm`.

**Deletes.** Only swallow an error from `Delete` when `mgerr.IsNotFound(err)` is true. Swallowing anything else drops the resource from state while it still exists in Mailgun.

**Secrets.** Mark passwords, keys and secrets `Sensitive: true` in both resource and data source schemas.

**State.** Use `types.String`/`Bool`/`Int64` and handle `IsNull()`/`IsUnknown()` explicitly. For Optional+Computed attributes that force replacement, set `UseStateForUnknown` or updates will plan a spurious destroy/recreate.

**Go version.** Resolve it from the `go` directive in `go.mod`, never from a hardcoded number in a doc. In CI use `actions/setup-go` with `go-version-file: 'go.mod'`. Stale hardcoded versions have caused real breakage here.

## Documentation

`docs/` is **generated** by `tfplugindocs`. Edit `templates/` and run `make generate`. Edits made directly to `docs/` are erased on the next generate, and CI fails on the resulting diff.

## Testing

Unit tests are mostly schema assertions. Real CRUD logic is covered only by acceptance tests, which need `TF_ACC=1` and `MAILGUN_API_KEY`, mutate a live Mailgun account, and cost money. Do not run them casually.

Some resources are account-level singletons (`mailgun_ip_allowlist`), so concurrent runs against a shared account contend with each other.

## Out of scope

Account-level and admin API key management. The provider authenticates with exactly that credential, so managing it creates a bootstrap paradox in which an apply can revoke the provider's own access. `mailgun_domain_sending_key` covers the domain-scoped subset and stays in scope.

## Before you commit

Run `make fmt` and `make lint`. Add a changelog entry per [.changelog/README.md](.changelog/README.md) unless the change is docs-, test-, refactor- or CI-only.
