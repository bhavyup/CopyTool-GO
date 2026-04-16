# Contributing to CopyTool

Thanks for helping improve CopyTool. This project uses strict, production-oriented quality gates to keep changes safe and maintainable.

## 1. Development Setup

```bash
git clone https://github.com/bhavyup/CopyTool-GO.git
cd CopyTool-GO
go run ./cmd/copytool
```

Optional local install on Windows:

```powershell
./scripts/install-local.ps1
copytool -version
```

## 2. Branching and Scope

- Base branch: `main`
- Keep each branch focused to one concern
- Avoid unrelated churn in pull requests

Suggested naming:

- `feat/<short-topic>`
- `fix/<short-topic>`
- `refactor/<short-topic>`
- `docs/<short-topic>`

## 3. Required Local Checks

Run these before opening a pull request:

```bash
gofmt -w ./...
go test ./...
go test -race ./...
go vet ./...
```

Optional local coverage view:

```bash
go test -covermode=atomic -coverprofile=coverage.out ./...
go tool cover -func=coverage.out
```

## 4. CI Gates

Pull requests are validated by automated workflows for:

1. formatting (`gofmt`)
2. vet checks
3. module hygiene (`go mod tidy` diff must be empty)
4. unit tests on Linux, macOS, and Windows
5. race detector + coverage
6. vulnerability scan (`govulncheck`)
7. build matrix verification
8. CodeQL security analysis

See [docs/QUALITY_GATES.md](docs/QUALITY_GATES.md) for details.

## 5. Testing Expectations

When changing behavior, add or update tests close to the changed logic.

Priority test areas:

- scan, ignore, and filter behavior
- selection and tree state transitions
- output target resolution and export output determinism

See [docs/TESTING.md](docs/TESTING.md).

## 6. Pull Request Checklist

- [ ] Summary explains problem and solution clearly
- [ ] Tests cover behavior changes
- [ ] All required local checks pass
- [ ] Docs/help text updated when user-visible behavior changes
- [ ] No secrets or credentials were added
- [ ] No unrelated files were modified

## 7. Release Impact

If your change affects output format, CLI flags, or packaging behavior:

- note it in PR description
- ensure backward compatibility expectations are explicit
- update docs as needed

Release workflow details: [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md)

## 8. Security Reporting

Do not report vulnerabilities in public issues.

- Read [SECURITY.md](SECURITY.md)
- Report privately via GitHub advisories:
	https://github.com/bhavyup/CopyTool-GO/security/advisories/new

## 9. Code of Conduct

Be respectful and collaborative in code review and issue discussions.