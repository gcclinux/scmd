# Update version across all project files
# Usage: .\scripts\update-version.ps1 <new_version>
# Example: .\scripts\update-version.ps1 2.0.6

param(
    [Parameter(Mandatory = $true, Position = 0)]
    [string]$NewVersion
)

# Validate version format
if ($NewVersion -notmatch '^\d+\.\d+\.\d+$') {
    Write-Host "Error: Invalid version format '$NewVersion'. Expected format: X.Y.Z" -ForegroundColor Red
    exit 1
}

# Get current version from the release file
$CurrentVersion = (Get-Content -Path "release" -Raw).Trim()
Write-Host "Updating version: $CurrentVersion -> $NewVersion" -ForegroundColor Cyan

# 1. release file
Set-Content -Path "release" -Value $NewVersion -NoNewline
Add-Content -Path "release" -Value ""
Write-Host "Updated: release" -ForegroundColor Green

# 2. internal/updater/version.go
$versionGo = Get-Content -Path "internal/updater/version.go" -Raw
$versionGo = $versionGo -replace [regex]::Escape("const Release = `"$CurrentVersion`""), "const Release = `"$NewVersion`""
Set-Content -Path "internal/updater/version.go" -Value $versionGo -NoNewline
Write-Host "Updated: internal/updater/version.go" -ForegroundColor Green

# 3. docker/Dockerfile
$dockerfile = Get-Content -Path "docker/Dockerfile" -Raw
$dockerfile = $dockerfile -replace [regex]::Escape("app.version=`"$CurrentVersion`""), "app.version=`"$NewVersion`""
Set-Content -Path "docker/Dockerfile" -Value $dockerfile -NoNewline
Write-Host "Updated: docker/Dockerfile" -ForegroundColor Green

# 4. scripts/build.sh
$buildSh = Get-Content -Path "scripts/build.sh" -Raw
$buildSh = $buildSh -replace "(?m)^VERSION=`"$([regex]::Escape($CurrentVersion))`"", "VERSION=`"$NewVersion`""
Set-Content -Path "scripts/build.sh" -Value $buildSh -NoNewline
Write-Host "Updated: scripts/build.sh" -ForegroundColor Green

# 5. scripts/build.ps1
$buildPs1 = Get-Content -Path "scripts/build.ps1" -Raw
$buildPs1 = $buildPs1 -replace [regex]::Escape("`$VERSION = `"$CurrentVersion`""), "`$VERSION = `"$NewVersion`""
Set-Content -Path "scripts/build.ps1" -Value $buildPs1 -NoNewline
Write-Host "Updated: scripts/build.ps1" -ForegroundColor Green

Write-Host ""
Write-Host "Version updated to $NewVersion in all files." -ForegroundColor Green
