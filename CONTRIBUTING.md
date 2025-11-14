# Contributing to Pantheon Metrics Prometheus Exporter

Thank you for your interest in contributing to this project! We welcome contributions of all kinds, including bug fixes, new features, documentation improvements, and more.

## How to Contribute

1. **Fork the repository** and create a new branch for your changes
2. **Make your changes** following our coding standards
3. **Write or update tests** as appropriate
4. **Ensure all tests pass** by running `go test -v ./...`
5. **Create a pull request** with a clear title and description

## Conventional Commits Requirement

This project requires all pull requests to follow the [Conventional Commits](https://www.conventionalcommits.org/) specification. This helps us maintain a clean and readable commit history, and enables automated changelog generation.

### PR Title Format

Your pull request title **must** follow this format:

```
<type>[optional scope]: <description>
```

#### Allowed Types

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

#### Optional Scopes

Scopes help categorize changes by area of the codebase:

- **metrics**: Metric collection and formatting
- **collector**: Prometheus collector implementation
- **refresh**: Periodic refresh and queue management
- **client**: Terminus API client interactions
- **auth**: Authentication and token management
- **config**: Configuration handling

#### Examples of Valid PR Titles

✅ Good:
- `feat: add support for custom metric labels`
- `fix: handle nil pointer in collector refresh`
- `docs: update prometheus configuration examples`
- `refactor(client): simplify authentication flow`
- `perf(refresh): optimize queue processing algorithm`
- `test: add coverage for multi-account scenarios`
- `chore: update dependencies to latest versions`
- `ci: add caching for Go modules in workflow`

❌ Bad:
- `Add feature` (missing type)
- `Feat: Add feature` (description should be lowercase)
- `feature: add support` (wrong type name)
- `fix stuff` (too vague)
- `WIP: working on metrics` (use draft PRs instead)

### Breaking Changes

If your change includes breaking changes (changes that require users to modify their configuration or usage), add `!` after the type/scope:

```
feat!: change default port from 8080 to 9090
refactor(client)!: remove deprecated fetchSiteInfo method
```

And include a `BREAKING CHANGE:` section in the PR description explaining the impact and migration path.

## Pull Request Process

1. **Update documentation** if you're adding or changing functionality
2. **Run linting** to catch code quality issues:
   ```bash
   golangci-lint run
   ```
3. **Run all tests** and ensure they pass:
   ```bash
   go test -v -race ./...
   go vet ./...
   ```
4. **Build the project** to ensure it compiles:
   ```bash
   go build -v -o pantheon-metrics-exporter
   ```
5. **Create your PR** with a conventional commit title
6. **Fill out the PR description** explaining:
   - What changes you made
   - Why you made them
   - How to test them
   - Any breaking changes or migration notes

### PR Title Validation

We use automated checks to validate PR titles. If your PR title doesn't follow the conventional commits format, the check will fail with a helpful error message. Simply edit your PR title to match the required format.

You can bypass this check by adding the `skip-conventional-commits` label, but please only use this for exceptional cases (like automated dependency PRs).

## Coding Standards

### Code Quality Tools

We use [golangci-lint](https://golangci-lint.run/) to maintain code quality. The linter runs automatically in CI, but you should run it locally before submitting PRs:

```bash
# Run all configured linters
golangci-lint run

# Run with auto-fix for some issues
golangci-lint run --fix
```

#### Pre-commit Hooks (Recommended)

We strongly recommend using pre-commit hooks to automatically check your code before committing:

1. Install [pre-commit](https://pre-commit.com/):
   ```bash
   # Using pip
   pip install pre-commit

   # Or using homebrew (macOS)
   brew install pre-commit
   ```

2. Install the git hooks:
   ```bash
   pre-commit install
   ```

The pre-commit hooks will automatically run:
- `go mod tidy` - Keep dependencies tidy
- `go fmt` - Format code
- `go vet` - Check for suspicious constructs
- `golangci-lint` - Run all configured linters
- `go test` - Run tests with race detector and coverage
- `go build` - Verify the project builds

This ensures your code meets quality standards before you commit, saving time in code review.

### Go Code Style

- Follow standard Go conventions and idioms
- Run `go fmt` before committing
- Use `go vet` to catch common issues
- Write clear, descriptive variable and function names
- Add comments for exported functions and complex logic

### Testing

- Write tests for new functionality
- Update tests when modifying existing code
- Aim for good test coverage
- Use table-driven tests where appropriate
- Test both success and error cases

### Error Handling

- Always handle errors appropriately
- Log errors with sufficient context
- Don't panic unless absolutely necessary
- Return errors rather than swallowing them

## Development Setup

### Prerequisites

- Go 1.23 or later
- Terminus CLI installed and configured
- A Pantheon machine token for testing

### Running Tests

```bash
# Run all tests
go test -v ./...

# Run tests with race detection
go test -v -race ./...

# Run tests with coverage
go test -v -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

### Testing Locally

```bash
# Build the exporter
go build -o pantheon-metrics-exporter

# Run with test token
export PANTHEON_MACHINE_TOKENS="your-test-token"
./pantheon-metrics-exporter -port=8080

# Check metrics endpoint
curl http://localhost:8080/metrics
```

## Reporting Bugs

When reporting bugs, please include:

- A clear description of the issue
- Steps to reproduce
- Expected behavior vs actual behavior
- Go version and OS
- Terminus CLI version
- Relevant logs or error messages

## Suggesting Features

We welcome feature suggestions! Please:

- Check if the feature has already been requested
- Explain the use case and benefits
- Consider if it fits the project's scope and goals

## Questions?

If you have questions about contributing, feel free to:

- Open an issue for discussion
- Ask in your pull request
- Review existing issues and PRs for examples

## License

By contributing to this project, you agree that your contributions will be licensed under the same license as the project.

---

Thank you for contributing! 🎉
