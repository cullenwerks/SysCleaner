param(
    [string]$Version = "2.0.0",
    [ValidateSet("amd64", "arm64", "all")]
    [string]$Arch = "amd64",
    [switch]$WithIcon = $true,
    [switch]$Debug = $false
)

function Build-SysCleaner {
    param(
        [string]$TargetArch,
        [string]$Version,
        [bool]$WithIcon,
        [bool]$Debug
    )

    Write-Host ""
    Write-Host "=== Building SysCleaner v$Version ($TargetArch) ===" -ForegroundColor Cyan

    # Set environment for cross-compilation
    $env:GOOS = "windows"
    $env:GOARCH = $TargetArch

    # Determine output names
    $archSuffix = if ($TargetArch -eq "amd64") { "x64" } else { "arm64" }
    $sysoFile = "rsrc_windows_$TargetArch.syso"

    if ($Debug) {
        $exeName = "SysCleaner-v$Version-$archSuffix-debug.exe"
    } else {
        $exeName = "SysCleaner-$archSuffix.exe"
    }

    # Compile icon resource for this architecture
    if ($WithIcon -and (Test-Path "assets/icon.ico")) {
        Write-Host "  Compiling icon resource ($TargetArch)..." -ForegroundColor Yellow
        # Remove any existing .syso files to avoid conflicts
        Remove-Item -Path "*.syso" -ErrorAction SilentlyContinue
        rsrc -ico assets/icon.ico -arch $TargetArch -o $sysoFile
        if ($?) {
            Write-Host "  Icon compiled: $sysoFile" -ForegroundColor Green
        } else {
            Write-Host "  Icon compilation failed (continuing without icon)" -ForegroundColor Yellow
        }
    } elseif ($WithIcon) {
        Write-Host "  Icon file not found at assets/icon.ico" -ForegroundColor Yellow
        Write-Host "  Building without custom icon..." -ForegroundColor Yellow
    }

    # Build executable
    Write-Host "  Building executable..." -ForegroundColor Yellow

    if ($Debug) {
        go build -tags gui -o $exeName
    } else {
        go build -tags gui -ldflags="-s -w -H=windowsgui" -trimpath -o $exeName
    }

    # Clean up .syso file after build
    Remove-Item -Path $sysoFile -ErrorAction SilentlyContinue

    # Verify build
    if (Test-Path $exeName) {
        $size = (Get-Item $exeName).Length / 1MB
        Write-Host "  Build successful!" -ForegroundColor Green
        Write-Host "    Executable: $exeName" -ForegroundColor Cyan
        Write-Host "    Architecture: $TargetArch" -ForegroundColor Cyan
        Write-Host "    Size: $([math]::Round($size, 2)) MB" -ForegroundColor Cyan
        return $exeName
    } else {
        Write-Host "  Build failed for $TargetArch!" -ForegroundColor Red
        return $null
    }
}

# --- Main ---

Write-Host "Building SysCleaner v$Version..." -ForegroundColor Cyan

# Clean old builds
Write-Host "Cleaning old builds..." -ForegroundColor Yellow
Remove-Item -Path "SysCleaner*.exe" -ErrorAction SilentlyContinue
Remove-Item -Path "*.syso" -ErrorAction SilentlyContinue

# Download dependencies
Write-Host "Downloading dependencies..." -ForegroundColor Yellow
go mod download
go mod tidy

# Determine which architectures to build
if ($Arch -eq "all") {
    $architectures = @("amd64", "arm64")
} else {
    $architectures = @($Arch)
}

$builtFiles = @()

foreach ($targetArch in $architectures) {
    $result = Build-SysCleaner -TargetArch $targetArch -Version $Version -WithIcon $WithIcon -Debug $Debug
    if ($result) {
        $builtFiles += $result
    }
}

# Summary
Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
if ($builtFiles.Count -gt 0) {
    Write-Host "Build complete! Built $($builtFiles.Count) executable(s):" -ForegroundColor Green
    foreach ($file in $builtFiles) {
        Write-Host "  $file" -ForegroundColor Cyan
    }
    Write-Host ""
    Write-Host "Run with: .\<executable-name>" -ForegroundColor Green
} else {
    Write-Host "All builds failed!" -ForegroundColor Red
    exit 1
}
