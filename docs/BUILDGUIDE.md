# SysCleaner v2.0 - GUI Build Guide

Complete guide for building SysCleaner v2.0 with GUI, custom icon, and Windows integration.

**Supported architectures:** x64 (amd64) and ARM64 (Windows on ARM).

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Build](#quick-build)
- [Building for ARM64](#building-for-arm64)
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
   # For x64 systems: download the windows-amd64 installer
   # For ARM64 systems: download the windows-arm64 installer
   # Verify installation:
   go version
   # Should output: go version go1.21.x windows/amd64 (or windows/arm64)
   ```

2. **GCC Compiler (for CGO/Fyne)**
   ```powershell
   # --- For x64 builds ---
   # Option 1: TDM-GCC (easiest)
   # Download from: https://jmeubank.github.io/tdm-gcc/
   # Install to default location (C:\TDM-GCC-64)

   # Option 2: MSYS2 (more flexible)
   # Download from: https://www.msys2.org/
   # After install, run in MSYS2 terminal:
   pacman -S mingw-w64-x86_64-gcc

   # --- For ARM64 builds ---
   # Use LLVM/MinGW for ARM64 cross-compilation:
   # Download from: https://github.com/mstorsjo/llvm-mingw/releases
   # Or use MSYS2 on an ARM64 device:
   pacman -S mingw-w64-aarch64-gcc

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

- **OS**: Windows 10/11 (64-bit, x64 or ARM64)
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
rsrc -ico assets/icon.ico -arch amd64 -o rsrc_windows_amd64.syso

# 5. Build SysCleaner GUI (x64)
$env:GOARCH="amd64"
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-x64.exe

# 6. Run
.\SysCleaner-x64.exe
```

Or use the build script:
```powershell
.\build.ps1                  # x64 (default)
.\build.ps1 -Arch arm64     # ARM64
.\build.ps1 -Arch all       # Both architectures
```

---

## Building for ARM64

SysCleaner natively supports Windows on ARM (ARM64) devices such as Surface Pro X, Lenovo ThinkPad X13s, Samsung Galaxy Book Go, and other Snapdragon-powered Windows PCs.

### Native build on ARM64 device

If you're building directly on a Windows ARM64 device:

```powershell
# Go auto-detects the host architecture
git clone https://github.com/cullenwerks/SysCleaner.git
cd SysCleaner
go mod download
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-arm64.exe
```

### Cross-compile from x64

You can build ARM64 binaries from an x64 machine:

```powershell
# Set target architecture
$env:GOOS = "windows"
$env:GOARCH = "arm64"

# For CGO (Fyne GUI), you need an ARM64 cross-compiler:
# Option 1: Disable CGO (some Fyne features may be limited)
$env:CGO_ENABLED = "0"
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-arm64.exe

# Option 2: Use LLVM-MinGW cross-compiler (full CGO support)
# Download from: https://github.com/mstorsjo/llvm-mingw/releases
$env:CGO_ENABLED = "1"
$env:CC = "aarch64-w64-mingw32-gcc"
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-arm64.exe
```

### Using the build script

```powershell
# Build for ARM64
.\build.ps1 -Arch arm64

# Build both x64 and ARM64
.\build.ps1 -Arch all

# Build both + create release ZIPs
.\build-all.ps1 -Package
```

### Architecture compatibility notes

- All Windows APIs used by SysCleaner (NtSetSystemInformation, registry, service control) are fully available on ARM64
- The `golang.org/x/sys/windows` package handles architecture differences transparently
- Performance characteristics (RAM trimming, service management) are identical on both architectures
- Windows on ARM can run x64 binaries via emulation, but native ARM64 builds are faster and more power-efficient

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
# The -arch flag must match your target architecture

# For x64:
rsrc -ico assets/icon.ico -arch amd64 -o rsrc_windows_amd64.syso

# For ARM64:
rsrc -ico assets/icon.ico -arch arm64 -o rsrc_windows_arm64.syso

# Go will automatically include the matching .syso during build
```

### Step 5: Build the Executable

```powershell
# Build for x64 (default on most machines)
$env:GOARCH = "amd64"
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-x64.exe

# Build for ARM64
$env:GOARCH = "arm64"
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-arm64.exe

# Build completes in 30-60 seconds
# Output: ~15-20 MB per architecture
```

### Step 6: Test the Application

```powershell
# Launch SysCleaner (use the correct binary for your architecture)
.\SysCleaner-x64.exe    # On x64 systems
.\SysCleaner-arm64.exe  # On ARM64 systems

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

# For x64 builds:
# Download TDM-GCC: https://jmeubank.github.io/tdm-gcc/
# Or install MSYS2 and run: pacman -S mingw-w64-x86_64-gcc

# For ARM64 cross-compilation from x64:
# Download LLVM-MinGW: https://github.com/mstorsjo/llvm-mingw/releases
# Set: $env:CC = "aarch64-w64-mingw32-gcc"

# For native ARM64 builds:
# Install MSYS2 and run: pacman -S mingw-w64-aarch64-gcc

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
# TDM-GCC (x64): C:\TDM-GCC-64\bin
# MSYS2 (x64):   C:\msys64\mingw64\bin
# MSYS2 (ARM64): C:\msys64\clangarm64\bin

# For ARM64 cross-compilation, set the correct compiler:
$env:CC = "aarch64-w64-mingw32-gcc"

# Alternatively, use static build (no CGO):
$env:CGO_ENABLED=0
go build -tags gui -ldflags="-s -w" -o SysCleaner-x64.exe
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

### Creating Release Packages

Build and package for both architectures:

```powershell
# Automated: build both architectures and create release ZIPs
.\build-all.ps1 -Package

# This creates:
#   SysCleaner-v2.0-Windows-x64.zip
#   SysCleaner-v2.0-Windows-arm64.zip
```

Or manually:

```powershell
# 1. Build for each architecture
.\build.ps1 -Arch all

# 2. Create x64 release package
mkdir SysCleaner-v2.0-Windows-x64
copy SysCleaner-x64.exe SysCleaner-v2.0-Windows-x64\SysCleaner.exe
copy README.md SysCleaner-v2.0-Windows-x64\
copy LICENSE SysCleaner-v2.0-Windows-x64\
Compress-Archive -Path SysCleaner-v2.0-Windows-x64 -DestinationPath SysCleaner-v2.0-Windows-x64.zip

# 3. Create ARM64 release package
mkdir SysCleaner-v2.0-Windows-arm64
copy SysCleaner-arm64.exe SysCleaner-v2.0-Windows-arm64\SysCleaner.exe
copy README.md SysCleaner-v2.0-Windows-arm64\
copy LICENSE SysCleaner-v2.0-Windows-arm64\
Compress-Archive -Path SysCleaner-v2.0-Windows-arm64 -DestinationPath SysCleaner-v2.0-Windows-arm64.zip

# 4. Upload both ZIPs to GitHub Releases
```

### Code Signing (Optional)

For trusted distribution, sign the executable:

```powershell
# Requires code signing certificate
# Use signtool.exe from Windows SDK
signtool sign /f certificate.pfx /p password /t http://timestamp.digicert.com SysCleaner.exe
```

---

## Build Automation Scripts

The project includes two build scripts:

### `build.ps1` - Single or Multi-Architecture Build

```powershell
# Build for x64 (default)
.\build.ps1

# Build for ARM64
.\build.ps1 -Arch arm64

# Build for both architectures
.\build.ps1 -Arch all

# Debug build
.\build.ps1 -Debug

# Custom version
.\build.ps1 -Version "2.0.1"

# Build without icon
.\build.ps1 -WithIcon:$false
```

Parameters:
| Parameter | Default | Description |
|-----------|---------|-------------|
| `-Version` | `"2.0.0"` | Version string for output naming |
| `-Arch` | `"amd64"` | Target architecture: `amd64`, `arm64`, or `all` |
| `-WithIcon` | `$true` | Embed application icon |
| `-Debug` | `$false` | Include debug symbols and console window |

### `build-all.ps1` - Build & Package Both Architectures

```powershell
# Build both architectures
.\build-all.ps1

# Build both + create release ZIP packages
.\build-all.ps1 -Package

# Custom version
.\build-all.ps1 -Version "2.0.1"
```

---

## Summary

**Minimum steps for building SysCleaner:**

1. Install Go 1.21+
2. Install GCC (TDM-GCC for x64, LLVM-MinGW for ARM64 cross-compilation)
3. Clone repository
4. Run: `go mod download`
5. (Optional) Create icon and run: `rsrc -ico assets/icon.ico -arch amd64 -o rsrc_windows_amd64.syso`
6. Build:
   - **x64:** `$env:GOARCH="amd64"; go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-x64.exe`
   - **ARM64:** `$env:GOARCH="arm64"; go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner-arm64.exe`
   - **Both:** `.\build.ps1 -Arch all`
7. Launch: `.\SysCleaner-x64.exe` or `.\SysCleaner-arm64.exe`

**That's it! You now have a fully functional SysCleaner GUI application for your architecture.**

---

For questions or issues, see the main [README.md](../README.md) or open an issue on GitHub.
