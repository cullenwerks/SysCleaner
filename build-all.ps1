param(
    [string]$Version = "2.0.0",
    [switch]$WithIcon = $true,
    [switch]$Package = $false
)

Write-Host "Building SysCleaner v$Version for all architectures..." -ForegroundColor Cyan
Write-Host ""

# Build both architectures
.\build.ps1 -Version $Version -Arch all -WithIcon:$WithIcon

# Optionally create release packages
if ($Package) {
    Write-Host ""
    Write-Host "Creating release packages..." -ForegroundColor Cyan

    $archMap = @{
        "x64"   = "amd64"
        "arm64" = "arm64"
    }

    foreach ($entry in $archMap.GetEnumerator()) {
        $suffix = $entry.Key
        $goarch = $entry.Value
        $exeName = "SysCleaner-$suffix.exe"

        if (Test-Path $exeName) {
            $releaseDir = "SysCleaner-v$Version-Windows-$suffix"
            $zipName = "$releaseDir.zip"

            # Clean previous release artifacts
            Remove-Item -Path $releaseDir -Recurse -ErrorAction SilentlyContinue
            Remove-Item -Path $zipName -ErrorAction SilentlyContinue

            # Create release directory
            New-Item -ItemType Directory -Path $releaseDir | Out-Null
            Copy-Item $exeName "$releaseDir\SysCleaner.exe"
            Copy-Item README.md $releaseDir\
            Copy-Item LICENSE $releaseDir\

            # Create ZIP
            Compress-Archive -Path $releaseDir -DestinationPath $zipName
            Remove-Item -Path $releaseDir -Recurse

            $zipSize = (Get-Item $zipName).Length / 1MB
            Write-Host "  Created: $zipName ($([math]::Round($zipSize, 2)) MB)" -ForegroundColor Green
        }
    }

    Write-Host ""
    Write-Host "Release packages ready for upload to GitHub Releases." -ForegroundColor Green
}
