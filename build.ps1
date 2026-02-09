param(
    [string]$Version = "2.0.0",
    [switch]$WithIcon = $true,
    [switch]$Debug = $false
)

Write-Host "Building SysCleaner v$Version..." -ForegroundColor Cyan

# Clean old builds
Write-Host "Cleaning old builds..." -ForegroundColor Yellow
Remove-Item -Path "*.exe" -ErrorAction SilentlyContinue
Remove-Item -Path "*.syso" -ErrorAction SilentlyContinue

# Download dependencies
Write-Host "Downloading dependencies..." -ForegroundColor Yellow
go mod download
go mod tidy

# Compile icon resource
if ($WithIcon -and (Test-Path "assets/icon.ico")) {
    Write-Host "Compiling icon resource..." -ForegroundColor Yellow
    rsrc -ico assets/icon.ico -o rsrc_windows_amd64.syso
    if ($?) {
        Write-Host "Icon compiled successfully" -ForegroundColor Green
    } else {
        Write-Host "Icon compilation failed (continuing without icon)" -ForegroundColor Yellow
    }
} elseif ($WithIcon) {
    Write-Host "Icon file not found at assets/icon.ico" -ForegroundColor Yellow
    Write-Host "  Building without custom icon..." -ForegroundColor Yellow
    Write-Host "  See assets/README.md for icon creation guide" -ForegroundColor Yellow
}

# Build executable
Write-Host "Building executable..." -ForegroundColor Yellow

if ($Debug) {
    go build -tags gui -o "SysCleaner-v$Version-debug.exe"
    $exeName = "SysCleaner-v$Version-debug.exe"
} else {
    go build -tags gui -ldflags="-s -w -H=windowsgui" -trimpath -o "SysCleaner.exe"
    $exeName = "SysCleaner.exe"
}

# Verify build
if (Test-Path $exeName) {
    $size = (Get-Item $exeName).Length / 1MB
    Write-Host ""
    Write-Host "Build successful!" -ForegroundColor Green
    Write-Host "  Executable: $exeName" -ForegroundColor Cyan
    Write-Host "  Size: $([math]::Round($size, 2)) MB" -ForegroundColor Cyan
    Write-Host "  Version: $Version" -ForegroundColor Cyan
    Write-Host ""
    Write-Host "Run with: .\$exeName" -ForegroundColor Green
} else {
    Write-Host ""
    Write-Host "Build failed!" -ForegroundColor Red
    Write-Host "  Check the error messages above" -ForegroundColor Red
    exit 1
}

# Clean up .syso file (optional - uncomment to remove after build)
# Remove-Item -Path "*.syso" -ErrorAction SilentlyContinue
