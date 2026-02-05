# Changelog Process

This directory contains changelog entries for the Mailgun Terraform Provider. We use [go-changelog](https://github.com/hashicorp/go-changelog) to generate `CHANGELOG.md` from these entries.

## Adding a Changelog Entry

When submitting a pull request, create a file named `{PR-NUMBER}.txt` in this directory.

### File Format

```
```release-note:{TYPE}
{DESCRIPTION}
```
```

### Entry Types

| Type | When to Use |
|------|-------------|
| `new-resource` | Adding a new resource (use resource name only) |
| `new-data-source` | Adding a new data source (use data source name only) |
| `enhancement` | Improving existing functionality |
| `bug` | Fixing a bug |
| `breaking-change` | Changes that break backward compatibility |
| `deprecation` | Deprecating functionality |
| `note` | Important information for users |

### Examples

**New Resource** (`.changelog/123.txt`):
```
```release-note:new-resource
mailgun_domain_ip
```
```

**New Data Source** (`.changelog/124.txt`):
```
```release-note:new-data-source
mailgun_subaccounts
```
```

**Bug Fix** (`.changelog/125.txt`):
```
```release-note:bug
resource/mailgun_domain: Fixed incorrect handling of wildcard attribute
```
```

**Enhancement** (`.changelog/126.txt`):
```
```release-note:enhancement
resource/mailgun_webhook: Added support for multiple URLs
```
```

**Multiple Changes in One PR** (`.changelog/127.txt`):
```
```release-note:new-resource
mailgun_template
```

```release-note:new-data-source
mailgun_templates
```
```

## When Changelog Entries Are NOT Required

- Documentation-only changes
- Test-only changes
- Code refactoring without user-facing changes
- CI/CD workflow changes

For these PRs, add the `changelog-not-required` label.

## References

- [HashiCorp go-changelog](https://github.com/hashicorp/go-changelog)
- [AWS Provider Changelog Process](https://hashicorp.github.io/terraform-provider-aws/changelog-process/)
