# SysCleaner Assets

This directory contains resources for building the SysCleaner GUI application.

## Icon Setup

### Creating the Icon

The application icon should be a flame/fire design in orange/red colors to match the SysCleaner brand.

**Option 1: Use an Icon Generator (Easiest)**

1. Go to https://www.favicon-generator.org/ or https://icoconvert.com/
2. Upload a PNG image (512x512 recommended):
   - Design: Flame/fire symbol
   - Colors: Orange (#FF5500) and red gradients
   - Background: Transparent or dark
3. Generate .ico file with multiple sizes (16x16, 32x32, 48x48, 256x256)
4. Save as `icon.ico` in this directory

**Option 2: Use GIMP (Free)**

1. Download GIMP: https://www.gimp.org/
2. Create 512x512 image with flame design
3. Export As → `icon.ico`
4. Select multiple sizes when prompted

**Option 3: Use Online AI Generator**

1. Go to https://www.bing.com/images/create or similar
2. Prompt: "Orange flame icon on transparent background, minimalist, gaming"
3. Download result
4. Convert to .ico using https://icoconvert.com/
5. Save to this directory

### Recommended Design

```
🔥 SysCleaner Icon Design Specs:
- Style: Modern, minimalist flame
- Primary Color: #FF5500 (Flame Orange)
- Secondary Color: #DC1E1E (Red for extreme mode)
- Background: Transparent or dark (#121212)
- Sizes: 16x16, 32x32, 48x48, 256x256
- Format: .ico (Windows icon format)
```

## Building with Icon

Once `icon.ico` is created in this directory:

### Windows (PowerShell)

```powershell
# 1. Install rsrc (resource compiler) - only needed once
go install github.com/akavel/rsrc@latest

# 2. Compile icon resource into .syso file
rsrc -ico assets/icon.ico -o rsrc_windows_amd64.syso

# 3. Build SysCleaner GUI
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe

# The .syso file will be automatically embedded
```

### Alternative: Using windres (if you have MinGW)

```powershell
# Compile resource file
windres -i assets/icon.rc -o rsrc.syso -O coff

# Build
go build -tags gui -ldflags="-s -w -H=windowsgui" -o SysCleaner.exe
```

## Quick Icon for Testing

If you don't have a custom icon yet, you can use a simple flame emoji or download a free icon from:
- https://www.flaticon.com/ (search "flame icon orange")
- https://iconarchive.com/ (search "fire icon")
- https://www.iconfinder.com/ (search "flame orange")

**Make sure to use icons with appropriate licenses (CC0, MIT, or commercial-friendly).**

## File Structure

```
assets/
├── README.md       (this file)
├── icon.ico        (application icon - you create this)
├── icon.rc         (resource definition file)
└── icon.png        (optional: source image for icon)
```

## Troubleshooting

**"rsrc: command not found"**
```powershell
go install github.com/akavel/rsrc@latest
# Make sure $GOPATH/bin is in your PATH
```

**Icon not showing in .exe**
- Make sure `icon.ico` exists in assets/ directory
- Rebuild the .syso file with rsrc
- Clean and rebuild: `go clean && go build ...`
- Check Windows Explorer: Right-click exe → Properties → Should show icon

**Build errors about .syso**
- Delete any old .syso files
- Regenerate with `rsrc -ico assets/icon.ico -o rsrc_windows_amd64.syso`
- Make sure .syso file is in the root directory (not assets/)
