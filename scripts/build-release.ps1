param(
    [string]$Version = "dev",
    [string]$OutputDir = "dist"
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir
Set-Location $RepoRoot

$DistDir = Join-Path $RepoRoot $OutputDir
New-Item -ItemType Directory -Force -Path $DistDir | Out-Null

$Commit = "none"
try {
    $Commit = (& git -C $RepoRoot rev-parse --short HEAD 2>$null).Trim()
    if (-not $Commit) { $Commit = "none" }
} catch {
    $Commit = "none"
}

$BuildDate = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
$ExePath = Join-Path $DistDir "copytool.exe"

$ldflags = "-s -w -X main.version=$Version -X main.commit=$Commit -X main.buildDate=$BuildDate"

Write-Host "Building copytool..."
go build -trimpath -ldflags $ldflags -o $ExePath ./cmd/copytool

if (-not (Test-Path $ExePath)) {
    throw "Build did not produce $ExePath"
}

$ArtifactBase = "copytool-$Version-windows-amd64"
$StageDir = Join-Path $DistDir $ArtifactBase

if (Test-Path $StageDir) {
    Remove-Item $StageDir -Recurse -Force
}
New-Item -ItemType Directory -Force -Path $StageDir | Out-Null

Copy-Item $ExePath (Join-Path $StageDir "copytool.exe") -Force

$ReadmePath = Join-Path $RepoRoot "README.md"
if (Test-Path $ReadmePath) {
    Copy-Item $ReadmePath (Join-Path $StageDir "README.md") -Force
}

$LicensePath = Join-Path $RepoRoot "LICENSE"
if (Test-Path $LicensePath) {
    Copy-Item $LicensePath (Join-Path $StageDir "LICENSE") -Force
}

$ZipPath = Join-Path $DistDir "$ArtifactBase.zip"
if (Test-Path $ZipPath) {
    Remove-Item $ZipPath -Force
}

Compress-Archive -Path (Join-Path $StageDir "*") -DestinationPath $ZipPath -Force

Write-Host ""
Write-Host "Build complete."
Write-Host "Executable: $ExePath"
Write-Host "Package dir: $StageDir"
Write-Host "Zip package: $ZipPath"
Write-Host ""

& $ExePath -version