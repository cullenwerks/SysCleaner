# SysCleaner v2.0 - GUI Build Guide

Complete guide for building SysCleaner v2.0 with GUI, custom icon, and Windows integration.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Build](#quick-build)
- [Detailed Build Steps](#detailed-build-steps)
- [Creating the Application Icon](#creating-the-application-icon)
- [Build Configurations](#build-configurations)
- [Optimization](#optimization)
- [Troubleshooting](#troubleshooting)
- [Distribution](#distribution)

---

## Prerequisites

### Required Software

1. **Go 1.21 or higher**
   ```powershell
   # Download from https://go.dev/dl/
   # Verify installation:
   go version
   # Should output: go version go1.21.x windows/amd64 or higher
   ```

2. **GCC Compiler (for CGO/Fyne)**
   ```powershell
   # Option 1: TDM-GCC (easiest)
   # Download from: https://jmeubank.github.io/tdm-gcc/
   # Install to default location (C:\TDM-GCC-64)

   # Option 2: MSYS2 (more flexible)
   # Download from: https://www.msys2.org/
   # After install, run in MSYS2 terminal:
   pacman -S mingw-w64-x86_64-gcc

   # Verify GCC:
   gcc --version
   ```

3. **Resource Compiler (for icon embedding)**
   ```powershell
   # Install rsrc tool:
   go install github.com/akavel/rsrc@latest

   # Verify (should be in %GOPATH%\bin or %USERPROFILE%\go\bin):
   rsrc -h
   ```

### System Requirements

- **OS**: Windows 10/11 (64-bit)
- **RAM**: 4 GB minimum (8 GB recommended)
- **Disk Space**: ~1 GB (Go toolchain + dependencies + build cache)

---

## Quick Build

For experienced developers:

```powershell
# 1. Clone repository
git clone https://github.com/cullenwerks/SysCleaner.git
cd SysCleaner

# 2. Install dependencies
go mod download

# 3. (Optional but recommended) Create icon
# See "Creating the Application Icon" section below
# Save as assets/icon.ico

# 4. Compile icon resource
rsrc -ico assets/icon.ico -o rsrc_windows_amd64.syso

# 5. Build SysCleaner GUI
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe

# 6. Run
.\SysCleaner.exe
```

---

## Detailed Build Steps

### Step 1: Obtain Source Code

```powershell
# Clone repository
git clone https://github.com/cullenwerks/SysCleaner.git
cd SysCleaner

# Or download ZIP from GitHub and extract
```

### Step 2: Install Dependencies

```powershell
# Download all Go module dependencies
go mod download

# Verify dependencies
go mod verify

# Clean up if needed
go mod tidy
```

### Step 3: Create Application Icon

See [Creating the Application Icon](#creating-the-application-icon) section below.

For quick testing, you can skip this step - the app will build without an icon.

### Step 4: Compile Icon Resource (Optional)

If you created `assets/icon.ico`:

```powershell
# Generate .syso file (Windows resource object)
rsrc -ico assets/icon.ico -o rsrc_windows_amd64.syso

# This creates rsrc_windows_amd64.syso in the root directory
# Go will automatically include it during build
```

### Step 5: Build the Executable

```powershell
# Build SysCleaner with all optimizations
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe

# Build completes in 30-60 seconds
# Output: SysCleaner.exe (~15-20 MB)
```

### Step 6: Test the Application

```powershell
# Launch SysCleaner
.\SysCleaner.exe

# GUI should open immediately
# No terminal window should appear
```

---

## Creating the Application Icon

SysCleaner uses a flame/fire icon in orange/red colors.

### Method 1: AI Generation (Recommended)

```
1. Go to https://www.bing.com/images/create (free AI image generator)
2. Prompt: "Minimalist orange flame icon on transparent background,
   gaming logo, simple geometric fire symbol, flat design"
3. Download the generated image (save as PNG)
4. Convert to .ico:
   - Go to https://icoconvert.com/
   - Upload PNG
   - Select sizes: 16x16, 32x32, 48x48, 256x256
   - Download icon.ico
5. Save to SysCleaner/assets/icon.ico
```

### Method 2: Use GIMP (Free Software)

```
1. Download GIMP: https://www.gimp.org/
2. Create new image: 512x512 pixels, transparent background
3. Design flame using gradient tools:
   - Foreground: #FF5500 (orange)
   - Background: #DC1E1E (red)
4. File → Export As → icon.ico
5. Select "Compressed (Windows BMP)" with multiple sizes
6. Save to SysCleaner/assets/icon.ico
```

### Method 3: Download Free Icon

```
Free icon sources (CC0 or commercial-friendly):
- https://www.flaticon.com/ (search "flame orange icon")
- https://www.iconfinder.com/ (filter by free license)
- https://iconarchive.com/ (check license)

Download as .ico or .png, convert if needed
Save to SysCleaner/assets/icon.ico
```

### Icon Specifications

```
Format: .ico (Windows icon format)
Sizes: 16x16, 32x32, 48x48, 256x256 (embedded in single .ico file)
Colors: Primary #FF5500, Secondary #DC1E1E
Background: Transparent
Style: Modern, minimalist flame/fire symbol
```

---

## Build Configurations

### Standard GUI Build (Recommended)

```powershell
# Full optimizations, no console window
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe

# Size: ~15-20 MB
# No terminal window
# Full GUI functionality
```

### Debug Build (For Development)

```powershell
# Includes debug symbols, console window for logs
go build -tags gui -o SysCleaner-debug.exe

# Size: ~30-40 MB
# Console window shows logs
# Useful for debugging
```

### Static Build (Maximum Compatibility)

```powershell
# Disable CGO for static linking
$env:CGO_ENABLED=0
go build -tags gui -ldflags="-s -w" -o SysCleaner-static.exe

# Size: ~16-22 MB
# No external DLL dependencies
# Works on systems without GCC runtime
# Note: Some Fyne features may be limited
```

### Ultra-Optimized Build

```powershell
# Maximum size reduction
go build -tags gui `
  -ldflags="-s -w -H=windowsgui" `
  -trimpath `
  -buildmode=exe `
  -o SysCleaner.exe

# Size: ~14-18 MB
# Smallest possible size
# Remove build paths for security
```

---

## Build Flags Explained

| Flag | Purpose | Effect |
|------|---------|--------|
| `-tags gui` | Enable GUI code | Required for Fyne.io GUI |
| `-ldflags="-s -w"` | Strip symbols | Reduces size by ~50% |
| `-ldflags="-H=windowsgui"` | No console | GUI-only, no terminal |
| `-trimpath` | Remove paths | Security, smaller binary |
| `-buildmode=exe` | Windows EXE | Explicit Windows executable |
| `-o SysCleaner.exe` | Output name | Name of final executable |

---

## Optimization

### Size Optimization

```powershell
# Smallest possible build:
go build -tags gui `
  -ldflags="-s -w -H=windowsgui" `
  -trimpath `
  -o SysCleaner.exe

# Then compress with UPX (optional):
# Download UPX: https://upx.github.io/
upx --best --lzma SysCleaner.exe
# Can reduce to ~6-8 MB (but slower startup)
```

### Performance Optimization

```powershell
# Enable all optimizations
$env:CGO_CFLAGS="-O3 -march=native"
$env:CGO_LDFLAGS="-O3"
go build -tags gui `
  -ldflags="-s -w -H=windowsgui" `
  -gcflags="all=-l -B" `
  -o SysCleaner.exe

# Faster execution
# Optimized for your CPU
# Larger binary size
```

---

## Troubleshooting

### "gcc: command not found"

```powershell
# Solution: Install GCC
# Download TDM-GCC: https://jmeubank.github.io/tdm-gcc/
# Or install MSYS2 and run: pacman -S mingw-w64-x86_64-gcc

# Verify after install:
gcc --version
```

### "rsrc: command not found"

```powershell
# Solution: Install rsrc tool
go install github.com/akavel/rsrc@latest

# Make sure Go bin directory is in PATH:
# Add to PATH: %USERPROFILE%\go\bin
# Or: C:\Users\YourName\go\bin
```

### "cannot find package fyne.io/fyne/v2"

```powershell
# Solution: Download dependencies
go mod download
go mod tidy

# If still failing:
go get fyne.io/fyne/v2@latest
```

### Icon not appearing in .exe

```powershell
# 1. Verify icon file exists
dir assets\icon.ico

# 2. Delete old .syso files
del *.syso

# 3. Regenerate .syso
rsrc -ico assets/icon.ico -o rsrc_windows_amd64.syso

# 4. Verify .syso was created
dir *.syso

# 5. Clean build cache and rebuild
go clean -cache
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe

# 6. Check icon in Windows Explorer:
# Right-click SysCleaner.exe → Properties → should show icon
```

### Build errors about CGO

```powershell
# For CGO issues, verify GCC is in PATH:
gcc --version

# If not found, add GCC to PATH:
# TDM-GCC: C:\TDM-GCC-64\bin
# MSYS2: C:\msys64\mingw64\bin

# Alternatively, use static build (no CGO):
$env:CGO_ENABLED=0
go build -tags gui -ldflags="-s -w" -o SysCleaner.exe
```

### "Permission denied" during build

```powershell
# Close running SysCleaner instances
taskkill /F /IM SysCleaner.exe

# Then rebuild
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe
```

### Large binary size (>30 MB)

```powershell
# Use optimization flags:
go build -tags gui `
  -ldflags="-s -w -H=windowsgui" `
  -trimpath `
  -o SysCleaner.exe

# Should be ~15-20 MB
# If still large, check:
go version  # Make sure using Go 1.21+
dir SysCleaner.exe  # Verify size
```

### Antivirus blocking build

```powershell
# Add exception in Windows Defender:
# 1. Windows Security → Virus & threat protection
# 2. Manage settings → Exclusions → Add exclusion
# 3. Add folder: Your SysCleaner project directory

# Or use static build (less likely to trigger):
$env:CGO_ENABLED=0
go build -tags gui -ldflags="-s -w" -o SysCleaner.exe
```

---

## Distribution

### Creating Release Package

```powershell
# 1. Build optimized executable
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe

# 2. Create distribution folder
mkdir SysCleaner-v2.0-Release
copy SysCleaner.exe SysCleaner-v2.0-Release\
copy README.md SysCleaner-v2.0-Release\
copy LICENSE SysCleaner-v2.0-Release\

# 3. Compress to ZIP
Compress-Archive -Path SysCleaner-v2.0-Release -DestinationPath SysCleaner-v2.0-Windows-amd64.zip

# 4. Upload to GitHub Releases
```

### Code Signing (Optional)

For trusted distribution, sign the executable:

```powershell
# Requires code signing certificate
# Use signtool.exe from Windows SDK
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com SysCleaner.exe
```

---

## Build Automation Script

Save as `build.ps1`:

```powershell
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
        Write-Host "✓ Icon compiled successfully" -ForegroundColor Green
    } else {
        Write-Host "✗ Icon compilation failed" -ForegroundColor Red
    }
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
    Write-Host "✓ Build successful!" -ForegroundColor Green
    Write-Host "  Executable: $exeName" -ForegroundColor Cyan
    Write-Host "  Size: $([math]::Round($size, 2)) MB" -ForegroundColor Cyan
    Write-Host "  Version: $Version" -ForegroundColor Cyan
} else {
    Write-Host "✗ Build failed!" -ForegroundColor Red
    exit 1
}

# Clean up .syso file (optional)
# Remove-Item -Path "*.syso" -ErrorAction SilentlyContinue

Write-Host "`nBuild complete! Run with: .\$exeName" -ForegroundColor Green
```

Usage:

```powershell
# Standard build
.\build.ps1

# Debug build
.\build.ps1 -Debug

# Custom version
.\build.ps1 -Version "2.0.1"

# Build without icon
.\build.ps1 -WithIcon:$false
```

---

## Summary

**Minimum steps for building SysCleaner:**

1. Install Go 1.21+
2. Install GCC (TDM-GCC recommended)
3. Clone repository
4. Run: `go mod download`
5. (Optional) Create icon and run: `rsrc -ico assets/icon.ico -o rsrc_windows_amd64.syso`
6. Run: `go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe`
7. Launch: `.\SysCleaner.exe`

**That's it! You now have a fully functional SysCleaner GUI application.**

---

For questions or issues, see the main [README.md](../README.md) or open an issue on GitHub.
