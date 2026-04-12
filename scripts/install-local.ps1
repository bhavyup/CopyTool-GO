param(
    [string]$Version = "dev",
    [string]$InstallRoot = (Join-Path $env:LOCALAPPDATA "Programs\copytool"),
    [switch]$SkipBuild
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$RepoRoot = Split-Path -Parent $ScriptDir

$BuildScript = Join-Path $ScriptDir "build-release.ps1"
$DistExe = Join-Path $RepoRoot "dist\copytool.exe"

if (-not $SkipBuild -or -not (Test-Path $DistExe)) {
    & $BuildScript -Version $Version
}

if (-not (Test-Path $DistExe)) {
    throw "Expected build output not found: $DistExe"
}

$BinDir = Join-Path $InstallRoot "bin"
New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

$InstalledExe = Join-Path $BinDir "copytool.exe"
Copy-Item $DistExe $InstalledExe -Force

$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathParts = @()

if ($currentPath) {
    $pathParts = $currentPath -split ';' | Where-Object { $_ -and $_.Trim() -ne '' }
}

$normalizedBin = $BinDir.TrimEnd('\')
$pathAlreadyContainsBin = $false

foreach ($part in $pathParts) {
    if ($part.TrimEnd('\') -ieq $normalizedBin) {
        $pathAlreadyContainsBin = $true
        break
    }
}

if (-not $pathAlreadyContainsBin) {
    $newPath = (($pathParts + $BinDir) | Select-Object -Unique) -join ';'
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $pathUpdated = $true
} else {
    $pathUpdated = $false
}

Write-Host ""
Write-Host "Installed copytool to:"
Write-Host "  $InstalledExe"
Write-Host ""

if ($pathUpdated) {
    Write-Host "Added to user PATH:"
    Write-Host "  $BinDir"
    Write-Host ""
    Write-Host "Open a NEW terminal window for PATH changes to take effect."
} else {
    Write-Host "User PATH already contains:"
    Write-Host "  $BinDir"
}

Write-Host ""
Write-Host "Installed version:"
& $InstalledExe -version

Write-Host ""
Write-Host "Next checks:"
Write-Host "  copytool -help"
Write-Host "  copytool -version"