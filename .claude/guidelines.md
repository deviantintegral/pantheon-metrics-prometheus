# Claude Code Guidelines for Pantheon Metrics Prometheus Exporter

## Pull Request Title Requirements

**IMPORTANT:** This project REQUIRES all pull request titles to follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. When creating pull requests, you MUST use conventional commit format for the PR title.

### Required PR Title Format

```
<type>[optional scope]: <description>
```

### Allowed Types

- **feat**: A new feature
- **fix**: A bug fix
- **docs**: Documentation only changes
- **style**: Changes that don't affect code meaning (formatting, whitespace, etc.)
- **refactor**: Code changes that neither fix bugs nor add features
- **perf**: Performance improvements
- **test**: Adding or updating tests
- **chore**: Maintenance tasks, dependency updates
- **ci**: Changes to CI/CD configuration
- **build**: Changes to build system or dependencies

The full list of allowed types is in .github/workflows/pr-title-check.yml.

### Optional Scopes

Use scopes to categorize changes by area of the codebase:

- **deps**: Dependency updates
- **metrics**: Metric collection and formatting
- **collector**: Prometheus collector implementation
- **refresh**: Periodic refresh and queue management
- **client**: Terminus API client interactions
- **auth**: Authentication and token management
- **config**: Configuration handling

The full list of allowed scopes is in .github/workflows/pr-title-check.yml.

### Description Requirements

- MUST start with a lowercase letter
- MUST be concise and descriptive
- MUST describe what the change does, not how
- SHOULD complete the sentence: "This change will..."

### Examples of Valid PR Titles

✅ **Correct:**
- `feat: add support for custom metric labels`
- `fix: handle nil pointer in collector refresh`
- `docs: update prometheus configuration examples`
- `refactor(client): simplify authentication flow`
- `perf(refresh): optimize queue processing algorithm`
- `test: add coverage for multi-account scenarios`
- `chore(deps): update dependencies to latest versions`
- `ci: add caching for Go modules in workflow`

❌ **Incorrect:**
- `Add feature` (missing type)
- `feat: Add feature` (description starts with uppercase)
- `feature: add support` (invalid type - use 'feat')
- `fix stuff` (too vague)
- `WIP: working on metrics` (use draft PRs instead)
- `Update README` (missing type)

### Breaking Changes

If the change includes breaking changes, add `!` after the type/scope:

```
feat!: change default port from 8080 to 9090
refactor(client)!: remove deprecated fetchSiteInfo method
```

## Automated Validation

This repository has automated PR title validation using GitHub Actions. If your PR title doesn't follow conventional commits format:

1. The `pr-title-check` workflow will fail
2. A comment will be added to the PR with the error
3. You must update the PR title to match the required format

## Summary for Claude Code

When creating pull requests in this repository:

1. **ALWAYS** use conventional commit format for PR titles
2. **CHOOSE** the most appropriate type from the allowed list
3. **ADD** a scope if it helps categorize the change (optional but recommended)
4. **ENSURE** the description starts with a lowercase letter
5. **KEEP** the description concise but descriptive
6. **MARK** breaking changes with `!` if applicable

For detailed contributing guidelines, refer to [CONTRIBUTING.md](../CONTRIBUTING.md).
