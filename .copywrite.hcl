schema_version = 1

project {
  license = "MPL-2.0"

  copyright_holder = "Hack The Box"

  # (OPTIONAL) Represents the year that the project initially began
  # Default: <the year the repo was first created>
  # Pinned so the start year does not depend on git history. copywrite 0.25.x
  # derives it from GetRepoFirstCommitYear(), which returns 0 in CI's PR-merge
  # checkout, making `make generate` non-deterministic between local and CI.
  copyright_year = 2025

  # (OPTIONAL) A list of globs that should not have copyright or license headers .
  # Supports doublestar glob patterns for more flexibility in defining which
  # files or folders should be ignored
  # Default: []
  header_ignore = [
    "vendor/**",
    "**autogen**",
    ".idea/**",
    # examples used within documentation (prose)
    "examples/**",

    # GitHub issue template configuration
    ".github/ISSUE_TEMPLATE/*.yml",

    # golangci-lint tooling configuration
    ".golangci.yml",

    # GoReleaser tooling configuration
    ".goreleaser.yml"
  ]

  # (OPTIONAL) Links to an upstream repo for determining repo relationships
  # This is for special cases and should not normally be set.
  # Default: ""
  # upstream = "hashicorp/<REPONAME>"
}
