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
3. upload both direct binaries and archive assets
4. generate SHA-256 checksums
5. publish a GitHub Release with generated notes

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

## End-user Install and Run

Release assets include:

1. Direct binaries:
	- `copytool_<version>_windows_amd64.exe`
	- `copytool_<version>_<os>_<arch>` for Linux/macOS
2. Windows installer binaries:
	- `copytool-installer_<version>_windows_amd64.exe`
	- `copytool-installer_<version>_windows_arm64.exe`
3. Archives:
	- `copytool_<version>_windows_amd64.zip`
	- `copytool_<version>_<os>_<arch>.tar.gz`

Example Windows installer usage (no source required):

```powershell
$version = "v0.1.2"
$asset = "copytool-installer_0.1.2_windows_amd64.exe"
$uri = "https://github.com/bhavyup/CopyTool-GO/releases/download/$version/$asset"
Invoke-WebRequest -Uri $uri -OutFile "copytool-installer.exe"
.\copytool-installer.exe -version latest
copytool -version
```

Silent/scripted mode:

```powershell
.\\copytool-installer.exe -version latest -quiet
```

Example Windows direct install:

```powershell
$version = "v0.1.0"
$asset = "copytool_0.1.0_windows_amd64.exe"
$uri = "https://github.com/bhavyup/CopyTool-GO/releases/download/$version/$asset"
Invoke-WebRequest -Uri $uri -OutFile "copytool.exe"
.\copytool.exe -version
```

Example Linux/macOS direct install:

```bash
VERSION="v0.1.0"
ASSET="copytool_0.1.0_linux_amd64"
curl -L -o copytool "https://github.com/bhavyup/CopyTool-GO/releases/download/${VERSION}/${ASSET}"
chmod +x copytool
./copytool -version
```

## Rollback

If a release is faulty:

1. mark release as pre-release or draft if not yet consumed
2. publish a patched tag (for example `v1.0.1`)
3. document the issue in release notes
