# SysCleaner - Complete Compilation Guide

This guide provides detailed instructions for compiling SysCleaner from source code into a standalone executable on Windows.

---

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Build (TL;DR)](#quick-build-tldr)
- [Detailed Build Steps](#detailed-build-steps)
- [Build Configurations](#build-configurations)
- [GUI Build (Fyne.io)](#gui-build-fyneio)
- [Cross-Compilation](#cross-compilation)
- [Optimization Flags](#optimization-flags)
- [Troubleshooting](#troubleshooting)
- [Verification](#verification)

---

## Prerequisites

### Required Software

1. **Go 1.21 or higher**
   - Download: https://go.dev/dl/
   - Install the MSI for Windows
   - Verify installation:
     ```powershell
     go version
     # Should output: go version go1.21.x windows/amd64 (or higher)
     ```

2. **Git** (optional, for cloning repository)
   - Download: https://git-scm.com/download/win
   - Or download ZIP from GitHub directly

3. **Windows 10/11**
   - Required OS for compilation and execution
   - PowerShell or Command Prompt

### System Requirements

- **OS**: Windows 10/11 (64-bit recommended)
- **RAM**: 2 GB minimum (4 GB recommended for compilation)
- **Disk Space**: ~500 MB for Go toolchain + ~50 MB for project

---

## Quick Build (TL;DR)

For experienced developers who just want to build quickly:

```powershell
# 1. Clone or extract the project
git clone https://github.com/YOUR_USERNAME/syscleaner.git
cd syscleaner

# 2. Download dependencies
go mod download

# 3. Build CLI executable (includes extreme mode, all tools)
go build -ldflags="-s -w" -o syscleaner.exe

# 4. Run
./syscleaner.exe --help

# 5. (Optional) Build GUI version with Fyne.io
go get fyne.io/fyne/v2@v2.4.3
go mod tidy
go build -tags gui -ldflags="-s -w -H=windowsgui" -o syscleaner-gui.exe
./syscleaner-gui.exe gui
```

---

## Detailed Build Steps

### Step 1: Obtain Source Code

**Option A: Clone with Git**
```powershell
# Open PowerShell or Command Prompt
git clone https://github.com/YOUR_USERNAME/syscleaner.git
cd syscleaner
```

**Option B: Download ZIP**
1. Go to: https://github.com/YOUR_USERNAME/syscleaner
2. Click **Code** → **Download ZIP**
3. Extract to desired location
4. Open PowerShell/CMD in extracted folder:
   ```powershell
   cd C:\path\to\Go-Optimizer-Windows-main
   ```

### Step 2: Verify Go Installation

```powershell
# Check Go is installed
go version

# Check Go environment
go env

# Verify GOPATH (should show a path)
go env GOPATH
```

**Expected output:**
```
go version go1.21.5 windows/amd64
```

### Step 3: Download Dependencies

```powershell
# Download all required modules
go mod download

# Optional: Verify and clean dependencies
go mod verify
go mod tidy
```

**This downloads (CLI dependencies):**
- `github.com/spf13/cobra` - CLI framework
- `github.com/shirou/gopsutil/v3` - System metrics
- `github.com/kardianos/service` - Windows service management
- `golang.org/x/sys` - System calls
- All transitive dependencies

**Additional GUI dependency (only needed with `-tags gui`):**
- `fyne.io/fyne/v2` - Cross-platform GUI toolkit

**Dependency verification:**
```powershell
# List all dependencies
go list -m all
```

### Step 4: Build the Executable

**Basic Build (Development)**
```powershell
# Simple build with default settings
go build -o syscleaner.exe

# Output: syscleaner.exe (~15-20 MB)
```

**Optimized Build (Production - Recommended)**
```powershell
# Build with optimization flags
go build -ldflags="-s -w" -o syscleaner.exe

# Explanation:
#   -ldflags="-s -w"  : Strip debug info and symbol table
#   -o syscleaner.exe : Output filename

# Output: syscleaner.exe (~8-12 MB)
```

**Ultra-Optimized Build**
```powershell
# Maximum optimization + UPX compression
go build -ldflags="-s -w -H=windowsgui" -trimpath -o syscleaner.exe

# Then compress with UPX (if installed)
upx --best --lzma syscleaner.exe

# Output: syscleaner.exe (~3-5 MB)
```

### Step 5: Verify Build

```powershell
# Check file was created
dir syscleaner.exe

# Test executable
./syscleaner.exe --version
./syscleaner.exe --help
```

**Expected output:**
```
syscleaner.exe
Size: ~10 MB
Modified: [current date/time]
```

---

## Build Configurations

### Development Build

**For testing and debugging:**
```powershell
go build -o syscleaner-dev.exe
```

**Characteristics:**
- ✅ Full debug symbols
- ✅ Faster compilation
- ✅ Better error messages
- ❌ Larger file size (~20 MB)
- ❌ Slower execution

**Use when:** Testing changes, debugging issues

---

### Release Build (Recommended)

**For distribution and production use:**
```powershell
go build -ldflags="-s -w" -trimpath -o syscleaner.exe
```

**Characteristics:**
- ✅ Optimized for size (~10 MB)
- ✅ Optimized for speed
- ✅ No debug symbols
- ✅ Clean binary paths
- ❌ Harder to debug

**Use when:** Distributing to users, final release

---

### Static Build

**Fully self-contained executable:**
```powershell
$env:CGO_ENABLED=0
go build -ldflags="-s -w -extldflags '-static'" -o syscleaner.exe
```

**Characteristics:**
- ✅ No external dependencies
- ✅ Works on any Windows 10/11
- ✅ Portable

**Use when:** Maximum compatibility needed

---

### No-Console Build

**For background/daemon use (no console window):**
```powershell
go build -ldflags="-s -w -H=windowsgui" -o syscleaner.exe
```

**Characteristics:**
- ✅ No console window appears
- ✅ Cleaner user experience
- ❌ No visible console output

**Use when:** Running as service/daemon

---

## GUI Build (Fyne.io)

SysCleaner includes an optional graphical interface built with [Fyne.io](https://fyne.io/). The GUI is gated behind the `gui` build tag so the CLI can be built without any GUI dependencies.

### Prerequisites for GUI Build

1. **Go 1.21+** (same as CLI)
2. **C compiler** (CGO is required by Fyne):
   - **Windows**: Install [MSYS2](https://www.msys2.org/) and run `pacman -S mingw-w64-x86_64-gcc`
   - **Linux**: `sudo apt install gcc libgl1-mesa-dev xorg-dev` (Debian/Ubuntu)
   - **macOS**: Install Xcode command line tools: `xcode-select --install`
3. **Fyne dependency**: Added to go.mod (see below)

### Step 1: Add Fyne Dependency

```powershell
# From the project root
go get fyne.io/fyne/v2@v2.4.3
go mod tidy
```

### Step 2: Build the GUI Binary

```powershell
# Development build (with console for debugging)
go build -tags gui -o syscleaner-gui.exe

# Release build (no console window)
go build -tags gui -ldflags="-s -w -H=windowsgui" -o syscleaner-gui.exe
```

### Step 3: Launch the GUI

```powershell
# Launch with the gui subcommand
./syscleaner-gui.exe gui
```

The GUI opens a tabbed interface with:
- **Dashboard** - Real-time CPU/RAM/disk metrics, performance score
- **Extreme Mode** - Toggle extreme performance, game launcher buttons
- **Clean** - Checkbox-based cleaning with preview and progress
- **Optimize** - Startup, network, registry, disk optimization
- **Tools** - Deep registry clean, DISM scan/repair, debloat, telemetry
- **Monitor** - Live resource graphs and system event log

### GUI Theme

The GUI uses a custom dark theme:
- Background: RGB(18, 18, 18)
- Accent color: RGB(255, 85, 0) (flame orange)
- Foreground: RGB(230, 230, 230)

### Single Binary Approach

You can ship one binary that supports both CLI and GUI:

```powershell
go build -tags gui -ldflags="-s -w" -o syscleaner.exe

# CLI usage (default)
./syscleaner.exe analyze
./syscleaner.exe extreme --enable

# GUI usage
./syscleaner.exe gui
```

### Without the GUI Tag

If you build without `-tags gui`, the `gui` subcommand prints a helpful message:

```
GUI mode is not available in this build.
Rebuild with: go build -tags gui
```

---

## Cross-Compilation

Build for different architectures from any OS:

### Windows 64-bit (AMD64) - Most Common
```powershell
$env:GOOS="windows"
$env:GOARCH="amd64"
go build -ldflags="-s -w" -o syscleaner-amd64.exe
```

### Windows 32-bit (386)
```powershell
$env:GOOS="windows"
$env:GOARCH="386"
go build -ldflags="-s -w" -o syscleaner-386.exe
```

### Windows ARM64
```powershell
$env:GOOS="windows"
$env:GOARCH="arm64"
go build -ldflags="-s -w" -o syscleaner-arm64.exe
```

### Build All Architectures
```powershell
# PowerShell script to build all versions
$architectures = @("amd64", "386", "arm64")

foreach ($arch in $architectures) {
    $env:GOOS="windows"
    $env:GOARCH=$arch
    $output = "syscleaner-windows-$arch.exe"
    
    Write-Host "Building for $arch..."
    go build -ldflags="-s -w" -o $output
    Write-Host "✓ Created: $output"
}
```

---

## Optimization Flags

### Linker Flags Explained

```powershell
-ldflags="-s -w -X 'main.Version=1.0.0' -H=windowsgui"
```

| Flag | Effect | Size Impact |
|------|--------|-------------|
| `-s` | Strip debug symbols | -30% |
| `-w` | Strip DWARF debug info | -20% |
| `-X 'main.Version=1.0.0'` | Set version variable | None |
| `-H=windowsgui` | Hide console window | None |
| `-extldflags '-static'` | Static linking | +10% |

### Compiler Flags

```powershell
-trimpath -gcflags="-l" -asmflags="-trimpath"
```

| Flag | Effect |
|------|--------|
| `-trimpath` | Remove absolute paths from binary |
| `-gcflags="-l"` | Disable inlining (smaller binary) |
| `-asmflags="-trimpath"` | Trim assembly paths |

### Build Tags

```powershell
# Build with GUI (Fyne.io)
go build -tags gui -ldflags="-s -w" -o syscleaner-gui.exe

# Build with specific features
go build -tags="release production" -o syscleaner.exe
```

| Tag | Effect |
|-----|--------|
| `gui` | Includes Fyne.io GUI (requires fyne.io/fyne/v2 in go.mod) |

---

## Build Scripts

### PowerShell Build Script

Create `build.ps1`:
```powershell
# build.ps1 - Automated build script
param(
    [string]$Version = "dev",
    [switch]$Release = $false
)

Write-Host "🔨 Building SysCleaner $Version..." -ForegroundColor Cyan

# Clean old builds
if (Test-Path syscleaner.exe) {
    Remove-Item syscleaner.exe
}

# Set build flags
$ldflags = "-s -w"
if ($Version -ne "dev") {
    $ldflags += " -X 'main.Version=$Version'"
}

# Build
if ($Release) {
    Write-Host "📦 Release build with optimizations..."
    go build -ldflags="$ldflags" -trimpath -o syscleaner.exe
} else {
    Write-Host "🔧 Development build..."
    go build -o syscleaner.exe
}

# Verify
if (Test-Path syscleaner.exe) {
    $size = (Get-Item syscleaner.exe).Length / 1MB
    Write-Host "✓ Build successful!" -ForegroundColor Green
    Write-Host "  Size: $([math]::Round($size, 2)) MB"
    
    # Test
    ./syscleaner.exe --version
} else {
    Write-Host "✗ Build failed!" -ForegroundColor Red
    exit 1
}
```

**Usage:**
```powershell
# Development build
./build.ps1

# Release build
./build.ps1 -Version "1.0.0" -Release

# Or with PowerShell execution policy
powershell -ExecutionPolicy Bypass -File build.ps1 -Release
```

### Batch Build Script

Create `build.bat`:
```batch
@echo off
echo Building SysCleaner...

REM Clean old builds
if exist syscleaner.exe del syscleaner.exe

REM Download dependencies
go mod download

REM Build
go build -ldflags="-s -w" -o syscleaner.exe

REM Verify
if exist syscleaner.exe (
    echo Build successful!
    syscleaner.exe --version
) else (
    echo Build failed!
    exit /b 1
)
```

**Usage:**
```cmd
build.bat
```

---

## Troubleshooting

### Issue: "go: command not found"

**Cause:** Go is not installed or not in PATH

**Solution:**
```powershell
# Download and install Go from: https://go.dev/dl/

# Verify PATH includes Go
$env:PATH -split ';' | Select-String "Go"

# If not found, add to PATH (PowerShell as Admin)
[Environment]::SetEnvironmentVariable(
    "Path",
    [Environment]::GetEnvironmentVariable("Path", "Machine") + ";C:\Go\bin",
    "Machine"
)
```

---

### Issue: "cannot find package"

**Cause:** Dependencies not downloaded

**Solution:**
```powershell
# Clear module cache
go clean -modcache

# Re-download dependencies
go mod download
go mod verify
go mod tidy

# Try build again
go build -o syscleaner.exe
```

---

### Issue: Build succeeds but exe doesn't run

**Cause:** Missing Visual C++ Redistributable or antivirus blocking

**Solution:**
```powershell
# 1. Check if file exists and has size
dir syscleaner.exe

# 2. Try to run
./syscleaner.exe --help

# 3. If blocked, add antivirus exception
# 4. Install VC++ Redistributable:
# https://aka.ms/vs/17/release/vc_redist.x64.exe

# 5. Build as static binary
$env:CGO_ENABLED=0
go build -ldflags="-s -w" -o syscleaner.exe
```

---

### Issue: "Access denied" or "Permission denied"

**Cause:** File is locked or need administrator privileges

**Solution:**
```powershell
# 1. Close running instances
taskkill /F /IM syscleaner.exe

# 2. Run PowerShell as Administrator
# Right-click PowerShell → "Run as Administrator"

# 3. Build in different directory
mkdir C:\temp\syscleaner
cd C:\temp\syscleaner
# Copy source and build here
```

---

### Issue: Large binary size (>20 MB)

**Cause:** Debug symbols not stripped

**Solution:**
```powershell
# Use optimization flags
go build -ldflags="-s -w" -trimpath -o syscleaner.exe

# Optional: Use UPX compression
# Download UPX: https://upx.github.io/
upx --best --lzma syscleaner.exe

# Expected sizes:
# Without flags: ~20 MB
# With -ldflags:  ~10 MB
# With UPX:       ~4-5 MB
```

---

### Issue: Build is slow

**Cause:** Full rebuild or slow disk

**Solution:**
```powershell
# Enable build cache
go env -w GOCACHE="C:\temp\go-cache"

# Parallel compilation
go build -p 8 -o syscleaner.exe

# Incremental builds (only changed files)
go build -i -o syscleaner.exe
```

---

### Issue: GUI build fails with "cannot find package fyne.io/fyne/v2"

**Cause:** Fyne dependency not added to go.mod

**Solution:**
```powershell
# Add the Fyne dependency
go get fyne.io/fyne/v2@v2.4.3
go mod tidy

# Then build with the gui tag
go build -tags gui -o syscleaner-gui.exe
```

---

### Issue: GUI build fails with CGO errors

**Cause:** Fyne requires a C compiler (CGO)

**Solution:**
```powershell
# Windows: Install MSYS2 and MinGW-w64
# From MSYS2 terminal:
pacman -S mingw-w64-x86_64-gcc

# Ensure MinGW bin is in PATH
$env:PATH += ";C:\msys64\mingw64\bin"

# Verify
gcc --version

# Then rebuild
go build -tags gui -o syscleaner-gui.exe
```

---

### Issue: "module requires Go 1.21"

**Cause:** Older Go version installed

**Solution:**
```powershell
# Check current version
go version

# Download latest from: https://go.dev/dl/
# Install and verify
go version
```

---

## Verification

### Verify Build Success

```powershell
# 1. File exists and has reasonable size
dir syscleaner.exe
# Expected: 8-15 MB

# 2. Execute help command
./syscleaner.exe --help
# Should show command list

# 3. Check version (if implemented)
./syscleaner.exe --version
# Should show version info

# 4. Test basic functionality
./syscleaner.exe analyze --help
```

### Verify Binary Properties

```powershell
# Check if 64-bit
dumpbin /headers syscleaner.exe | Select-String "machine"
# Should show: 8664 machine (x64)

# Alternative: Use Go tool
go version syscleaner.exe
# Should show: syscleaner.exe: go1.21.x
```

### Verify Dependencies

```powershell
# List imported packages
go list -f '{{ join .Imports "\n" }}' syscleaner

# List all dependencies
go mod graph

# Check for vulnerabilities
go list -m -json all | go run golang.org/x/vuln/cmd/govulncheck@latest -json -
```

### Security Scan

```powershell
# Windows Defender scan
# Start → Windows Security → Virus & threat protection
# → Scan options → Custom scan → Select syscleaner.exe

# Or via PowerShell (Admin)
Start-MpScan -ScanPath ".\syscleaner.exe" -ScanType CustomScan
```

---

## Post-Build Steps

### Create Portable Package

```powershell
# Create release directory
mkdir release
Copy-Item syscleaner.exe release/
Copy-Item README.md release/
Copy-Item LICENSE release/

# Create ZIP
Compress-Archive -Path release/* -DestinationPath syscleaner-v1.0.0-windows-amd64.zip
```

### Generate Checksums

```powershell
# SHA256 hash
$hash = Get-FileHash syscleaner.exe -Algorithm SHA256
$hash.Hash | Out-File syscleaner.exe.sha256

# Display
Write-Host "SHA256: $($hash.Hash)"
```

### Sign Binary (Optional)

```powershell
# Requires code signing certificate
signtool sign /f certificate.pfx /p password /tr http://timestamp.digicert.com /td SHA256 /fd SHA256 syscleaner.exe

# Verify signature
signtool verify /pa syscleaner.exe
```

---

## Advanced Build Techniques

### Embedded Resources

**Embed version info:**
```go
// In main.go or version.go
package main

var (
    Version   = "dev"
    BuildTime = "unknown"
    GitCommit = "unknown"
)
```

**Build with version:**
```powershell
$version = "1.0.0"
$buildTime = Get-Date -Format "yyyy-MM-dd HH:mm:ss"
$gitCommit = git rev-parse --short HEAD

go build -ldflags="-s -w -X 'main.Version=$version' -X 'main.BuildTime=$buildTime' -X 'main.GitCommit=$gitCommit'" -o syscleaner.exe
```

### Conditional Compilation

**Use build tags:**
```go
// +build windows,amd64

package main
// Windows AMD64 specific code
```

**Build with tags:**
```powershell
go build -tags "windows amd64 release" -o syscleaner.exe
```

---

## CI/CD Build Examples

### GitHub Actions

Create `.github/workflows/build.yml`:
```yaml
name: Build

on: [push, pull_request]

jobs:
  build:
    runs-on: windows-latest
    steps:
      - uses: actions/checkout@v3
      
      - name: Set up Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      
      - name: Build CLI
        run: |
          go mod download
          go build -ldflags="-s -w" -o syscleaner.exe

      - name: Build GUI
        run: |
          go get fyne.io/fyne/v2@v2.4.3
          go mod tidy
          go build -tags gui -ldflags="-s -w -H=windowsgui" -o syscleaner-gui.exe

      - name: Test
        run: go test ./...

      - name: Upload CLI artifact
        uses: actions/upload-artifact@v3
        with:
          name: syscleaner-cli-windows-amd64
          path: syscleaner.exe

      - name: Upload GUI artifact
        uses: actions/upload-artifact@v3
        with:
          name: syscleaner-gui-windows-amd64
          path: syscleaner-gui.exe
```

---

## Conclusion

You now have multiple ways to build SysCleaner:

- **Quick CLI build**: `go build -o syscleaner.exe`
- **Release CLI build**: `go build -ldflags="-s -w" -o syscleaner.exe`
- **GUI build**: `go build -tags gui -ldflags="-s -w -H=windowsgui" -o syscleaner-gui.exe`
- **Script build**: `./build.ps1 -Release`

For most users, the **release CLI build** is recommended. Add the GUI build if you want the graphical interface with dashboard, extreme mode panel, and real-time monitoring.

If you encounter issues not covered here, please:
- Check the [Troubleshooting](#troubleshooting) section
- Open an issue on GitHub
- Review Go documentation: https://go.dev/doc/

**Happy building! 🚀**
