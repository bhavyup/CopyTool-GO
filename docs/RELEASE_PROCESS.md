# Release Process

This repository publishes releases from Git tags through GitHub Actions.

## Trigger

Create and push a semantic version tag:

```bash
git tag v1.0.0
git push origin v1.0.0
```

The `release.yml` workflow will:

1. build binaries for supported OS/architecture targets
2. package archives
3. generate SHA-256 checksums
4. publish a GitHub Release with generated notes

## Version Metadata

Release builds embed:

- version tag
- commit short SHA
- UTC build timestamp

Values are surfaced via:

```bash
copytool -version
```

## Post-release Verification

1. Download each artifact from GitHub Releases
2. Verify checksums against `checksums.txt`
3. Smoke-test binary startup (`copytool -version`, `copytool -cli`)

## Rollback

If a release is faulty:

1. mark release as pre-release or draft if not yet consumed
2. publish a patched tag (for example `v1.0.1`)
3. document the issue in release notes
