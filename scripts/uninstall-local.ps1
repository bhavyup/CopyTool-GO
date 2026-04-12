param(
    [string]$InstallRoot = (Join-Path $env:LOCALAPPDATA "Programs\copytool")
)

$ErrorActionPreference = "Stop"

$BinDir = Join-Path $InstallRoot "bin"
$currentPath = [Environment]::GetEnvironmentVariable("Path", "User")
$pathParts = @()

if ($currentPath) {
    $pathParts = $currentPath -split ';' | Where-Object { $_ -and $_.Trim() -ne '' }
}

$normalizedBin = $BinDir.TrimEnd('\')

$newPathParts = foreach ($part in $pathParts) {
    if ($part.TrimEnd('\') -ine $normalizedBin) {
        $part
    }
}

$newPath = $newPathParts -join ';'
[Environment]::SetEnvironmentVariable("Path", $newPath, "User")

if (Test-Path $InstallRoot) {
    Remove-Item $InstallRoot -Recurse -Force
}

Write-Host "Removed copytool install root:"
Write-Host "  $InstallRoot"
Write-Host ""
Write-Host "Removed bin directory from user PATH if it was present."
Write-Host "Open a NEW terminal window for PATH changes to take effect."