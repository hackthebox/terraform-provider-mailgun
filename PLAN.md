# Implementation Plan: SMTP Credentials and API Keys

## Overview

This plan outlines the implementation of two new resources for the Terraform Mailgun provider:
1. **SMTP Credentials** (`mailgun_smtp_credential`) - Manage SMTP authentication credentials per domain
2. **API Keys** (`mailgun_api_key`) - Manage API keys for sending and account access

Both resources follow the established Cloudflare provider pattern used in the `domains/` package.

---

## 1. SMTP Credentials Resource

### SDK Methods Available

| Operation | SDK Method | Notes |
|-----------|-----------|-------|
| Create | `CreateCredential(ctx, domain, login, password)` | Returns error only |
| Read | `ListCredentials(domain, opts)` | No single Get - must filter from list |
| Update | `ChangeCredentialPassword(ctx, domain, login, password)` | Only password can be changed |
| Delete | `DeleteCredential(ctx, domain, login)` | By domain + login |

### Credential Type (from SDK)

```go
type Credential struct {
    CreatedAt RFC2822Time `json:"created_at"`
    Login     string      `json:"login"`      // e.g., "postmaster" or "user@domain.com"
    Password  string      `json:"password"`   // Only returned in some contexts
}
```

### Terraform Resource Schema

```hcl
resource "mailgun_smtp_credential" "example" {
  # Required
  domain   = "example.com"          # Domain this credential belongs to
  login    = "notifications"        # Username part (will be login@domain)
  password = "secure-password"      # SMTP password (sensitive, write-only)
}
```

**Computed Attributes:**
- `id` - Composite ID: `{domain}/{login}`
- `created_at` - Timestamp when credential was created
- `full_login` - Full SMTP login: `{login}@{domain}`

### Resource Identifier

- **ID Format**: `{domain}/{login}` (e.g., `example.com/notifications`)
- **Import**: `terraform import mailgun_smtp_credential.example example.com/notifications`

### Implementation Notes

1. **No single Get endpoint**: Read operation must list all credentials and filter by login
2. **Password is write-only**: Password is not returned by API after creation, so we store it in state but cannot verify drift
3. **Login uniqueness**: Login must be unique per domain
4. **Full login format**: The actual SMTP username is `{login}@{domain}`

### File Structure

```
internal/provider/smtp_credentials/
├── resource.go              # CRUD implementation
├── resource_model.go        # SmtpCredentialModel struct
├── resource_schema.go       # Schema definition
├── data_source.go           # Single credential lookup
├── data_source_model.go     # Data source model
├── data_source_schema.go    # Data source schema
├── list_data_source.go      # List all credentials for a domain
├── list_data_source_model.go
├── list_data_source_schema.go
└── resource_test.go         # Tests
```

---

## 2. API Keys Resource

### SDK Methods Available

| Operation | SDK Method | Notes |
|-----------|-----------|-------|
| Create | `CreateAPIKey(ctx, role, opts)` | Returns full APIKey with secret |
| Read | `ListAPIKeys(ctx, opts)` | No single Get - must filter from list |
| Update | N/A | API keys are immutable after creation |
| Delete | `DeleteAPIKey(ctx, id)` | By key ID |

### CreateAPIKeyOptions

```go
type CreateAPIKeyOptions struct {
    Description string   // Human-readable description
    DomainName  string   // Required for "sending" role domain keys
    Email       string   // Email associated with key
    Expiration  uint64   // Key lifetime in seconds (0 = no expiration)
    Kind        string   // "domain", "user", or "web"
    UserID      string   // User ID to associate with key
    UserName    string   // Username to associate with key
}
```

### APIKey Type (from SDK)

```go
type APIKey struct {
    ID             string      `json:"id"`
    Description    string      `json:"description"`
    Kind           string      `json:"kind"`           // "domain", "user", "web"
    Role           string      `json:"role"`           // "admin", "sending", "developer", etc.
    CreatedAt      ISO8601Time `json:"created_at"`
    UpdatedAt      ISO8601Time `json:"updated_at"`
    DomainName     string      `json:"domain_name"`    // For domain-scoped keys
    Requestor      string      `json:"requestor"`
    UserName       string      `json:"user_name"`
    IsDisabled     bool        `json:"is_disabled"`
    ExpiresAt      ISO8601Time `json:"expires_at"`
    Secret         string      `json:"secret"`         // ONLY on create!
    DisabledReason string      `json:"disabled_reason"`
}
```

### Terraform Resource Schema

```hcl
resource "mailgun_api_key" "sending_key" {
  # Required
  role        = "sending"           # "admin", "sending", "developer", "basic", "support"

  # Optional
  description = "Production sending key"
  domain_name = "example.com"       # Required for "sending" role
  kind        = "domain"            # "domain", "user", "web" (default: "user")
  expiration  = 0                   # Seconds until expiration (0 = never)
}
```

**Computed Attributes:**
- `id` - API key ID from Mailgun
- `secret` - The API key secret (sensitive, only available after creation)
- `created_at` - Creation timestamp
- `updated_at` - Last update timestamp
- `is_disabled` - Whether the key is disabled
- `disabled_reason` - Reason if disabled
- `expires_at` - Expiration timestamp

### Key Types and Roles

| Kind | Description | Use Case |
|------|-------------|----------|
| `domain` | Domain-specific sending key | Production sending from specific domain |
| `user` | User-level key (default) | General API access |
| `web` | Web key (1-day max lifetime) | Temporary/web-based access |

| Role | Description | Permissions |
|------|-------------|-------------|
| `admin` | Full access | All operations |
| `sending` | Sending only | POST /messages, /messages.mime |
| `developer` | Development access | Most read/write operations |
| `basic` | Basic access | Limited read operations |
| `support` | Support access | Read-only support operations |

### Implementation Notes

1. **Secret only on create**: The `secret` field is ONLY returned when the key is created. After that, it cannot be retrieved. We must store it in state.
2. **Immutable after creation**: Role, kind, and most settings cannot be changed. Changes require destroy/recreate.
3. **Domain required for sending keys**: When `role = "sending"`, the `domain_name` is required.
4. **Plan modifier**: Most attributes should use `RequiresReplace()` since keys are immutable.

### File Structure

```
internal/provider/api_keys/
├── resource.go              # Create, Read, Delete (no Update)
├── resource_model.go        # ApiKeyModel struct
├── resource_schema.go       # Schema definition
├── data_source.go           # Single API key lookup by ID
├── data_source_model.go     # Data source model
├── data_source_schema.go    # Data source schema
├── list_data_source.go      # List all API keys
├── list_data_source_model.go
├── list_data_source_schema.go
└── resource_test.go         # Tests
```

---

## 3. Implementation Order

### Phase 1: SMTP Credentials (simpler, more commonly needed)
1. Create `internal/provider/smtp_credentials/` package
2. Implement resource model and schema
3. Implement CRUD operations
4. Implement single data source
5. Implement list data source
6. Register in provider.go
7. Add unit tests
8. Add examples
9. Commit and test

### Phase 2: API Keys
1. Create `internal/provider/api_keys/` package
2. Implement resource model and schema
3. Implement Create, Read, Delete (no Update)
4. Implement single data source
5. Implement list data source
6. Register in provider.go
7. Add unit tests
8. Add examples
9. Commit and test

---

## 4. Provider Registration

Update `internal/provider/provider.go`:

```go
func (p *mailgunProvider) DataSources(_ context.Context) []func() datasource.DataSource {
    return []func() datasource.DataSource{
        domains.NewDomainDataSource,
        domains.NewDomainsListDataSource,
        smtp_credentials.NewSmtpCredentialDataSource,      // NEW
        smtp_credentials.NewSmtpCredentialsListDataSource, // NEW
        api_keys.NewApiKeyDataSource,                      // NEW
        api_keys.NewApiKeysListDataSource,                 // NEW
    }
}

func (p *mailgunProvider) Resources(_ context.Context) []func() resource.Resource {
    return []func() resource.Resource{
        domains.NewDomainResource,
        smtp_credentials.NewSmtpCredentialResource, // NEW
        api_keys.NewApiKeyResource,                 // NEW
    }
}
```

---

## 5. Example Usage

### SMTP Credentials

```hcl
# Create a domain first
resource "mailgun_domain" "example" {
  name = "mail.example.com"
}

# Create SMTP credential for the domain
resource "mailgun_smtp_credential" "notifications" {
  domain   = mailgun_domain.example.name
  login    = "notifications"
  password = var.smtp_password
}

# Output the full login for use in applications
output "smtp_login" {
  value = mailgun_smtp_credential.notifications.full_login
  # e.g., "notifications@mail.example.com"
}
```

### API Keys

```hcl
# Create a domain-specific sending key
resource "mailgun_api_key" "production_sender" {
  role        = "sending"
  kind        = "domain"
  domain_name = mailgun_domain.example.name
  description = "Production sending key for notifications"
}

# Store the secret securely - it's only available after creation!
output "api_key_secret" {
  value     = mailgun_api_key.production_sender.secret
  sensitive = true
}

# Create an admin key for infrastructure management
resource "mailgun_api_key" "admin" {
  role        = "admin"
  description = "Terraform management key"
}
```

---

## 6. Testing Strategy

### Unit Tests
- Model attribute types validation
- Schema field presence and types
- Value conversion functions

### Acceptance Tests (require MAILGUN_API_KEY)
- Create SMTP credential, verify attributes
- Update SMTP credential password
- Import existing SMTP credential
- Delete SMTP credential
- Create API key, verify secret returned
- List API keys, verify filtering
- Delete API key

---

## 7. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Password drift undetectable | Medium | Document that password changes outside Terraform won't be detected |
| API key secret lost | High | Store in state; warn users to save externally |
| No single Get for credentials | Low | Filter from list; add caching if performance issue |
| API keys immutable | Low | Use RequiresReplace plan modifier |

---

## 8. Documentation

After implementation, generate docs with:
```bash
make docs
```

This will create:
- `docs/resources/smtp_credential.md`
- `docs/resources/api_key.md`
- `docs/data-sources/smtp_credential.md`
- `docs/data-sources/smtp_credentials.md`
- `docs/data-sources/api_key.md`
- `docs/data-sources/api_keys.md`

---

## Sources

- [Mailgun API Key Management](https://documentation.mailgun.com/docs/mailgun/user-manual/api-key-mgmt/rbac-mgmt)
- [API Key Roles](https://help.mailgun.com/hc/en-us/articles/26016288026907-API-Key-Roles)
- [Keys API Reference](https://documentation.mailgun.com/docs/mailgun/api-reference/send/mailgun/keys)
- [SMTP Credentials Help](https://help.mailgun.com/hc/en-us/articles/203380100-Where-can-I-find-my-API-keys-and-SMTP-credentials)
