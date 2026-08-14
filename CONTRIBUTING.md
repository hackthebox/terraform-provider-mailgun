# Contributing

Thanks for contributing to the Mailgun Terraform provider.

Build, install and test commands are in [README.md](README.md). Codebase conventions (package layout, error handling, state management) are in [AGENTS.md](AGENTS.md); it is addressed to AI coding agents but applies equally to people.

Use the Go version in the `go` directive of `go.mod`. Do not trust a version number quoted in a doc.

## Tests

Unit tests (`make test`) are fast and hit no network. They cover schemas; most CRUD logic is not covered by them.

Acceptance tests (`make testacc`) **create real resources in a real Mailgun account and can cost money**. They need `TF_ACC=1` and `MAILGUN_API_KEY`. Run them deliberately, against a test account, never your production workspace. Some resources are account-level singletons, so parallel runs on a shared account interfere with each other.

See [TESTING.md](TESTING.md) for local end-to-end verification.

## Documentation

`docs/` is generated. Edit `templates/`, then run `make generate`. CI fails if the committed `docs/` differs from what the templates produce.

## Pull requests

- Branch off `main`, named for its issue where there is one, for example `fix/gh-102-typed-404-handling`.
- Use [conventional commits](https://www.conventionalcommits.org/): `feat|fix|docs|chore|test|ci(scope): description`.
- Every PR needs at least one review. Squash merge.
- Add a changelog entry named for your PR number, per [.changelog/README.md](.changelog/README.md). Docs-, test-, refactor- and CI-only changes are exempt; label those `changelog-not-required`.
- Run `make fmt` and `make lint` before pushing. CI additionally runs `govulncheck`, CodeQL, a docs-drift check, and the acceptance suite across a Terraform version matrix.

## Reporting issues

Open a GitHub issue with the provider version, Terraform version, a minimal configuration that reproduces the problem, and the output you got against the output you expected. Redact API keys, domains and other secrets from logs before pasting them.
