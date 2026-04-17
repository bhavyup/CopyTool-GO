<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=0,1&height=200&text=CopyTool&fontSize=52&fontAlignY=35&desc=Turn%20messy%20codebases%20into%20clean%2C%20shareable%20context%20in%20seconds.&descAlignY=55&fontColor=ffffff" width="100%" />

</div>

<!-- readme-gen:start:badges -->
<p align="center">
  <img alt="Go" src="https://img.shields.io/badge/Go-1.26+-00ADD8?style=for-the-badge&logo=go&logoColor=white" />
  <img alt="Platform" src="https://img.shields.io/badge/Platform-Windows-0078D4?style=for-the-badge&logo=windows&logoColor=white" />
  <img alt="Mode" src="https://img.shields.io/badge/Modes-TUI%20%2B%20CLI-161B22?style=for-the-badge" />
  <img alt="Gitignore Aware" src="https://img.shields.io/badge/gitignore-aware-2EA043?style=for-the-badge" />
</p>

<p align="center">
  <img alt="CI" src="https://img.shields.io/github/actions/workflow/status/bhavyup/CopyTool-GO/ci.yml?branch=main&style=for-the-badge&label=CI" />
  <img alt="CodeQL" src="https://img.shields.io/github/actions/workflow/status/bhavyup/CopyTool-GO/codeql.yml?branch=main&style=for-the-badge&label=CodeQL" />
  <img alt="Release" src="https://img.shields.io/github/actions/workflow/status/bhavyup/CopyTool-GO/release.yml?style=for-the-badge&label=Release" />
</p>

<p align="center">
  <img alt="Stars" src="https://img.shields.io/github/stars/bhavyup/CopyTool-GO?style=for-the-badge&logo=github" />
  <img alt="Forks" src="https://img.shields.io/github/forks/bhavyup/CopyTool-GO?style=for-the-badge" />
  <img alt="Issues" src="https://img.shields.io/github/issues/bhavyup/CopyTool-GO?style=for-the-badge" />
  <img alt="Last Commit" src="https://img.shields.io/github/last-commit/bhavyup/CopyTool-GO?style=for-the-badge" />
</p>

<p align="center">
  <img src="https://skillicons.dev/icons?i=go,powershell,git,github&theme=dark" alt="Tech Stack" />
</p>
<!-- readme-gen:end:badges -->

CopyTool is a gitignore-aware context dumper that helps you curate exactly the files you want and export them as clean, fenced output for reviews, AI prompts, documentation, or handoff notes. Use the interactive TUI to browse and select with confidence, or run fast in direct CLI mode when you already know your file set.

<img src="https://capsule-render.vercel.app/api?type=rect&color=gradient&customColorList=0,1&height=1" width="100%" />

## Highlights

<table>
<tr>
<td width="50%" valign="top">

### 🧭 Dual Workflow
Use a full-screen TUI for visual selection and preview, or jump straight to scripted CLI export.

</td>
<td width="50%" valign="top">

### 🔍 Smart Selection
Respects root and nested .gitignore rules plus include/exclude extension and directory filters.

</td>
</tr>
<tr>
<td width="50%" valign="top">

### 📦 Share-Ready Output
Exports selected files into one markdown-fenced artifact with language-aware code fences.

</td>
<td width="50%" valign="top">

### 📋 Clipboard Friendly
Optionally mirrors exported content to clipboard for instant paste into chats, docs, or tickets.

</td>
</tr>
</table>

## Quick Start

```bash
git clone https://github.com/bhavyup/CopyTool-GO.git
cd CopyTool-GO
go run ./cmd/copytool
```

### Install Local Binary (Windows)

```powershell
./scripts/install-local.ps1
copytool -version
```

## Install From GitHub Release

Download the latest assets from:

- https://github.com/bhavyup/CopyTool-GO/releases/latest

Available release assets now include both:

- direct binaries, including Windows `.exe`
- packaged archives (`.zip` for Windows and `.tar.gz` for Linux/macOS)
- Windows installer binaries (`copytool-installer_<version>_windows_<arch>.exe`)

### Windows Installer (.exe)

If you do not have source code, use the installer executable from Releases:

```powershell
$version = "v0.1.2"
$asset = "copytool-installer_0.1.2_windows_amd64.exe"
$uri = "https://github.com/bhavyup/CopyTool-GO/releases/download/$version/$asset"

Invoke-WebRequest -Uri $uri -OutFile "copytool-installer.exe"
Unblock-File .\copytool-installer.exe
.\copytool-installer.exe -version latest
```

The installer:

- downloads the matching `copytool_<version>_windows_<arch>.exe`
- installs to `%LOCALAPPDATA%\Programs\copytool\bin\copytool.exe`
- adds the install bin directory to user `PATH`

### Windows (Direct .exe)

```powershell
$version = "v0.1.0"
$asset = "copytool_0.1.0_windows_amd64.exe"
$uri = "https://github.com/bhavyup/CopyTool-GO/releases/download/$version/$asset"

Invoke-WebRequest -Uri $uri -OutFile "copytool.exe"
Unblock-File .\copytool.exe
.\copytool.exe -version
```

### Linux/macOS (Direct Binary)

```bash
VERSION="v0.1.0"
ASSET="copytool_0.1.0_linux_amd64"
curl -L -o copytool "https://github.com/bhavyup/CopyTool-GO/releases/download/${VERSION}/${ASSET}"
chmod +x copytool
./copytool -version
```

### Using Archive Assets

- Windows archive: `copytool_<version>_windows_amd64.zip`
- Linux/macOS archive: `copytool_<version>_<os>_<arch>.tar.gz`

Extract the archive and run the binary inside.

### Usage Examples

Run interactive TUI:

```bash
copytool -root "D:\codes\myproject"
```

Direct CLI export with explicit paths:

```bash
copytool -cli -root "D:\codes\myproject" src README.md
```

Direct CLI export with filters and clipboard:

```bash
copytool -cli -include-ext ".go,.md" -exclude-dirs "dist,tests" -clipboard .
```

<img src="https://capsule-render.vercel.app/api?type=rect&color=gradient&customColorList=0,1&height=1" width="100%" />

<!-- readme-gen:start:architecture -->
## Architecture

```mermaid
graph TD
    A[🧑‍💻 User] --> B[🖥️ TUI Layer]
    A --> C[⌨️ CLI Mode]
    B --> D[🔍 Core Scan Engine]
    C --> D
    D --> E[🧹 Ignore + Filter Pipeline]
    E --> F[📂 Tree + Selection Model]
    F --> G[📦 Export Builder]
    G --> H[📄 tools/list.txt]
    G --> I[📝 output.txt or custom output path]
    G --> J[📋 Clipboard (optional)]
```
<!-- readme-gen:end:architecture -->

<!-- readme-gen:start:tree -->
## Project Structure

```text
📦 copytool-go
├── 📂 cmd/
│   └── 📂 copytool/
│       └── 📄 main.go              # CLI flags, mode selection, app boot
├── 📂 internal/
│   ├── 📂 cli/                      # direct export mode pipeline
│   ├── 📂 core/                     # scan, tree, filters, export, preview
│   ├── 📂 platform/                 # clipboard + native picker adapters
│   ├── 📂 theme/                    # shared visual tokens/styles
│   └── 📂 tui/                      # Bubble Tea UI model and views
├── 📂 scripts/
│   ├── 📄 build-release.ps1         # build + package zip
│   ├── 📄 install-local.ps1         # local install + PATH update
│   └── 📄 uninstall-local.ps1       # remove local install
├── 📂 tests/
│   └── 📄 ignore.txt                # ignore fixture file
├── 📂 tools/
│   ├── 📄 list.txt                  # selected file list
│   └── 📄 output.txt                # merged export output
└── 📄 go.mod
```
<!-- readme-gen:end:tree -->

## CLI Flags

```text
-root          project root
-output        output target (file, directory, ".", or empty default)
-tui           force TUI mode
-cli           run direct export mode
-clipboard     copy export result to clipboard
-include-ext   comma-separated extension allowlist
-exclude-ext   comma-separated extension denylist
-exclude-dirs  comma-separated directory denylist
-version       print build info
```

## Quality and Delivery

CopyTool now includes production-grade engineering gates and release automation:

- CI pipeline with formatting checks, vet, module hygiene, unit tests on Linux/macOS/Windows, race+coverage, vulnerability scanning, and build matrix validation
- Release pipeline that builds multi-platform artifacts, publishes checksums, and creates GitHub releases from semantic tags
- CodeQL static analysis and Dependabot automation for proactive security and dependency maintenance
- Contributor and maintainer runbooks for testing, quality gates, and release operations

Operational docs:

- [docs/QUALITY_GATES.md](docs/QUALITY_GATES.md)
- [docs/TESTING.md](docs/TESTING.md)
- [docs/RELEASE_PROCESS.md](docs/RELEASE_PROCESS.md)
- [SECURITY.md](SECURITY.md)

<!-- readme-gen:start:health -->
## Project Health

| Category | Status | Standard |
|:---------|:------:|:---------|
| Tests | ✅ | Core regression tests in `internal/core` with CI enforcement |
| CI/CD | ✅ | Multi-OS test matrix, race+coverage, vuln scan, build validation |
| Security | ✅ | CodeQL + `govulncheck` + private vulnerability reporting policy |
| Dependencies | ✅ | Weekly Dependabot updates for Go modules and GitHub Actions |
| Documentation | ✅ | Contributing, testing, quality-gate, release, and security runbooks |

> **Overall:** Production-ready baseline established. Next maturity target is expanding test coverage breadth beyond core package flows.
<!-- readme-gen:end:health -->

<img src="https://capsule-render.vercel.app/api?type=rect&color=gradient&customColorList=0,1&height=1" width="100%" />

## Contributing

We welcome contributions. See [CONTRIBUTING.md](CONTRIBUTING.md) for setup, workflow, and pull request guidance.

## License

This project is licensed under the MIT License. See [LICENSE](LICENSE).

<!-- readme-gen:start:footer -->
<div align="center">

<img src="https://capsule-render.vercel.app/api?type=waving&color=gradient&customColorList=0,1&height=100&section=footer" width="100%" />

Built with ❤️ by [Contributors](https://github.com/bhavyup/CopyTool-GO/graphs/contributors)

</div>
<!-- readme-gen:end:footer -->