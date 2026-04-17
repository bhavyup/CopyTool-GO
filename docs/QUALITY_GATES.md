# Quality Gates

This repository uses automated quality gates designed for production-grade maintainability.

## Required Gates (CI)

The CI workflow enforces the following on pull requests and pushes to `main`:

1. `gofmt` formatting check
2. `go vet ./...`
3. module hygiene check (`go mod tidy` with zero diff)
4. unit tests on Linux, macOS, and Windows
5. race detector and coverage run on Linux
6. minimum total coverage threshold (currently `10.0%`)
7. vulnerability scan with `govulncheck`
8. build verification across target OS/architecture matrix

## Security and Dependency Hygiene

- CodeQL analysis runs for pull requests, pushes to `main`, and weekly schedule.
- Dependabot tracks Go modules and GitHub Actions updates weekly.

## Branch Protection Recommendation

Enable branch protection on `main` with:

1. Require pull request before merging
2. Require status checks to pass (all CI jobs + CodeQL)
3. Require up-to-date branches before merge
4. Restrict force-push and branch deletion

## Local Pre-PR Checklist

Run this locally before opening a pull request:

```bash
go test ./...
go test -race ./...
go vet ./...
```

Then validate formatting:

```bash
gofmt -w ./...
```

## Release Integrity

Release workflow builds versioned archives for multiple platforms and publishes SHA-256 checksums.
