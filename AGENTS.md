# Claude Development Notes

## Pre-commit Hooks

**IMPORTANT**: Always install and run pre-commit when starting a new session.

### Installation and Setup

```bash
# Install pre-commit hooks
pre-commit install

# Run pre-commit on all files
pre-commit run --all-files
```

### Why This Matters

- Pre-commit hooks ensure code quality and consistency
- They catch issues before committing
- Running them at the start of a session ensures any uncommitted changes are validated
- Fixes should be applied immediately before continuing work

### Git Commit Rules

**NEVER** use `git commit --no-verify`. This flag bypasses pre-commit hooks and must not be used under any circumstances. If pre-commit hooks fail, fix the underlying issues instead of skipping the checks.

### Workflow

1. Start new session
2. Install pre-commit: `pre-commit install`
3. Run on all files: `pre-commit run --all-files`
4. Fix any issues reported
5. Commit fixes if needed
6. Continue with development work
