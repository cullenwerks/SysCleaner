package cleaner

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

// CleanOptions specifies what to clean.
type CleanOptions struct {
	TempFiles  bool
	Browser    bool
	Registry   bool
	Logs       bool
	Prefetch   bool
	Thumbnails bool
	DryRun     bool
}

// CleanResult holds the result of a cleaning operation.
type CleanResult struct {
	FilesDeleted int64
	SpaceFreed   int64
	Duration     time.Duration
	Errors       []error
}

// PerformClean orchestrates all cleaning operations based on options.
func PerformClean(opts CleanOptions) CleanResult {
	start := time.Now()
	result := CleanResult{}

	if opts.TempFiles {
		log.Println("[SysCleaner] Cleaning temporary files...")
		r := cleanTempFiles(opts.DryRun)
		result.merge(r)
	}
	if opts.Browser {
		log.Println("[SysCleaner] Cleaning browser caches...")
		r := cleanBrowserCaches(opts.DryRun)
		result.merge(r)
	}
	if opts.Prefetch {
		log.Println("[SysCleaner] Cleaning prefetch data...")
		r := cleanPrefetch(opts.DryRun)
		result.merge(r)
	}
	if opts.Thumbnails {
		log.Println("[SysCleaner] Cleaning thumbnail caches...")
		r := cleanThumbnails(opts.DryRun)
		result.merge(r)
	}
	if opts.Logs {
		log.Println("[SysCleaner] Cleaning log files...")
		r := cleanLogFiles(opts.DryRun)
		result.merge(r)
	}
	if opts.Registry {
		log.Println("[SysCleaner] Cleaning registry junk...")
		r := cleanRegistry(opts.DryRun)
		result.merge(r)
	}

	result.Duration = time.Since(start)
	log.Printf("[SysCleaner] Cleanup complete: %d files, %s freed in %s",
		result.FilesDeleted, FormatBytes(result.SpaceFreed), result.Duration.Round(time.Millisecond))
	return result
}

func (r *CleanResult) merge(other CleanResult) {
	r.FilesDeleted += other.FilesDeleted
	r.SpaceFreed += other.SpaceFreed
	r.Errors = append(r.Errors, other.Errors...)
}

func cleanTempFiles(dryRun bool) CleanResult {
	result := CleanResult{}

	tempDirs := getTempDirs()
	for _, dir := range tempDirs {
		r := cleanDirectory(dir, 0, dryRun)
		result.merge(r)
	}
	return result
}

func getTempDirs() []string {
	dirs := []string{}
	if runtime.GOOS == "windows" {
		dirs = append(dirs, os.Getenv("TEMP"))
		dirs = append(dirs, os.Getenv("TMP"))
		winDir := os.Getenv("WINDIR")
		if winDir != "" {
			dirs = append(dirs, filepath.Join(winDir, "Temp"))
		}
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "Temp"))
		}

		// Windows Update cache
		if winDir != "" {
			dirs = append(dirs, filepath.Join(winDir, "SoftwareDistribution", "Download"))
		}

		// Windows Installer cache (orphaned patches)
		if winDir != "" {
			dirs = append(dirs, filepath.Join(winDir, "Installer", "$PatchCache$"))
		}

		// Crash dump files
		if localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "CrashDumps"))
		}
		if winDir != "" {
			dirs = append(dirs, filepath.Join(winDir, "Minidump"))
		}

		// Windows Error Reporting
		if localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "Microsoft", "Windows", "WER"))
		}
		programData := os.Getenv("ProgramData")
		if programData != "" {
			dirs = append(dirs, filepath.Join(programData, "Microsoft", "Windows", "WER"))
		}

		// DirectX shader cache
		if localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "D3DSCache"))
		}

		// Windows icon cache
		if localAppData != "" {
			dirs = append(dirs, filepath.Join(localAppData, "IconCache.db"))
		}

		// Font cache
		if winDir != "" {
			dirs = append(dirs, filepath.Join(winDir, "ServiceProfiles", "LocalService", "AppData", "Local", "FontCache"))
		}

	} else {
		dirs = append(dirs, os.TempDir())
		dirs = append(dirs, "/tmp")
		// Linux/macOS trash
		home, _ := os.UserHomeDir()
		if home != "" {
			dirs = append(dirs, filepath.Join(home, ".local", "share", "Trash", "files"))
			dirs = append(dirs, filepath.Join(home, ".cache"))
		}
	}
	return dedup(dirs)
}

func cleanBrowserCaches(dryRun bool) CleanResult {
	result := CleanResult{}
	localAppData := os.Getenv("LOCALAPPDATA")
	appData := os.Getenv("APPDATA")

	if runtime.GOOS != "windows" {
		// Linux/macOS browser caches
		home, _ := os.UserHomeDir()
		if home != "" {
			linuxPaths := []string{
				filepath.Join(home, ".cache", "google-chrome"),
				filepath.Join(home, ".cache", "chromium"),
				filepath.Join(home, ".cache", "mozilla", "firefox"),
				filepath.Join(home, ".cache", "BraveSoftware"),
				filepath.Join(home, ".cache", "vivaldi"),
				filepath.Join(home, "Library", "Caches", "Google", "Chrome"),       // macOS
				filepath.Join(home, "Library", "Caches", "com.brave.Browser"),       // macOS
				filepath.Join(home, "Library", "Caches", "com.vivaldi.Vivaldi"),     // macOS
				filepath.Join(home, "Library", "Caches", "com.operasoftware.Opera"), // macOS
			}
			for _, p := range linuxPaths {
				r := cleanDirectory(p, 0, dryRun)
				result.merge(r)
			}
		}
		return result
	}

	if localAppData == "" && appData == "" {
		return result
	}

	// Chromium-based browsers store profiles in "User Data" with numbered profiles.
	// We enumerate all profiles (Default, Profile 1, Profile 2, etc.) for each browser.
	type chromiumBrowser struct {
		name    string
		dataDir string
	}

	chromiumBrowsers := []chromiumBrowser{}
	if localAppData != "" {
		chromiumBrowsers = append(chromiumBrowsers,
			chromiumBrowser{"Chrome", filepath.Join(localAppData, "Google", "Chrome", "User Data")},
			chromiumBrowser{"Edge", filepath.Join(localAppData, "Microsoft", "Edge", "User Data")},
			chromiumBrowser{"Brave", filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data")},
			chromiumBrowser{"Vivaldi", filepath.Join(localAppData, "Vivaldi", "User Data")},
			chromiumBrowser{"Opera", filepath.Join(appData, "Opera Software", "Opera Stable")},
			chromiumBrowser{"Opera GX", filepath.Join(appData, "Opera Software", "Opera GX Stable")},
		)
	}

	// Clean all Chromium-based browsers
	for _, browser := range chromiumBrowsers {
		r := cleanChromiumProfiles(browser.name, browser.dataDir, dryRun)
		result.merge(r)
	}

	// Firefox (profile-based cache structure)
	if appData != "" {
		firefoxProfiles := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
		r := cleanFirefoxCache(firefoxProfiles, dryRun)
		result.merge(r)
	}

	// Discord cache
	if appData != "" {
		discordCacheDirs := []string{
			filepath.Join(appData, "discord", "Cache"),
			filepath.Join(appData, "discord", "Code Cache"),
			filepath.Join(appData, "discord", "GPUCache"),
		}
		for _, d := range discordCacheDirs {
			r := cleanDirectory(d, 0, dryRun)
			result.merge(r)
		}
	}

	// Spotify cache
	if localAppData != "" {
		r := cleanDirectory(filepath.Join(localAppData, "Spotify", "Storage"), 0, dryRun)
		result.merge(r)
	}

	// Steam web browser cache
	if localAppData != "" {
		r := cleanDirectory(filepath.Join(localAppData, "Steam", "htmlcache"), 0, dryRun)
		result.merge(r)
	}

	return result
}

// cleanChromiumProfiles enumerates Chromium profile directories and cleans cache folders.
func cleanChromiumProfiles(browserName, userDataDir string, dryRun bool) CleanResult {
	result := CleanResult{}
	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		return result
	}

	// Chromium cache subdirectories to clean
	cacheSubdirs := []string{
		"Cache", "Code Cache", "GPUCache", "Service Worker", "ShaderCache",
	}

	entries, err := os.ReadDir(userDataDir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		// Match "Default" or "Profile N" directories
		if name == "Default" || strings.HasPrefix(name, "Profile ") {
			for _, sub := range cacheSubdirs {
				cacheDir := filepath.Join(userDataDir, name, sub)
				r := cleanDirectory(cacheDir, 0, dryRun)
				result.merge(r)
			}
		}
	}

	return result
}

func cleanFirefoxCache(profilesDir string, dryRun bool) CleanResult {
	result := CleanResult{}
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if entry.IsDir() {
			cacheDir := filepath.Join(profilesDir, entry.Name(), "cache2")
			r := cleanDirectory(cacheDir, 0, dryRun)
			result.merge(r)
			// Also clean startupCache
			startupCache := filepath.Join(profilesDir, entry.Name(), "startupCache")
			r = cleanDirectory(startupCache, 0, dryRun)
			result.merge(r)
		}
	}
	return result
}

func cleanPrefetch(dryRun bool) CleanResult {
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return CleanResult{}
	}
	prefetchDir := filepath.Join(winDir, "Prefetch")
	return cleanDirectory(prefetchDir, 30*24*time.Hour, dryRun)
}

func cleanThumbnails(dryRun bool) CleanResult {
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return CleanResult{}
	}
	thumbDir := filepath.Join(localAppData, "Microsoft", "Windows", "Explorer")
	result := CleanResult{}

	entries, err := os.ReadDir(thumbDir)
	if err != nil {
		return result
	}
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasPrefix(entry.Name(), "thumbcache_") || strings.HasPrefix(entry.Name(), "iconcache_")) {
			fpath := filepath.Join(thumbDir, entry.Name())
			info, err := entry.Info()
			if err != nil {
				result.Errors = append(result.Errors, err)
				continue
			}
			if dryRun {
				result.FilesDeleted++
				result.SpaceFreed += info.Size()
			} else {
				if err := os.Remove(fpath); err != nil {
					result.Errors = append(result.Errors, err)
				} else {
					result.FilesDeleted++
					result.SpaceFreed += info.Size()
				}
			}
		}
	}
	return result
}

func cleanLogFiles(dryRun bool) CleanResult {
	result := CleanResult{}
	winDir := os.Getenv("WINDIR")
	programData := os.Getenv("ProgramData")

	logDirs := []string{}
	if winDir != "" {
		logDirs = append(logDirs, filepath.Join(winDir, "Logs"))
		logDirs = append(logDirs, filepath.Join(winDir, "Debug"))    // Windows debug logs
		logDirs = append(logDirs, filepath.Join(winDir, "Panther")) // Windows setup logs
	}
	if programData != "" {
		logDirs = append(logDirs, filepath.Join(programData, "Logs"))
	}

	for _, dir := range logDirs {
		r := cleanDirectory(dir, 30*24*time.Hour, dryRun)
		result.merge(r)
	}
	return result
}

func cleanRegistry(dryRun bool) CleanResult {
	// Registry cleaning is Windows-specific; implemented in cleaner_windows.go
	return cleanRegistryPlatform(dryRun)
}

// cleanDirectory removes files in a directory. If maxAge > 0, only files older than maxAge are removed.
func cleanDirectory(dir string, maxAge time.Duration, dryRun bool) CleanResult {
	result := CleanResult{}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return result
	}

	now := time.Now()
	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("walk error %s: %w", path, err))
			return nil // continue walking
		}
		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}

		if maxAge > 0 && now.Sub(info.ModTime()) < maxAge {
			return nil // skip recent files
		}

		if dryRun {
			result.FilesDeleted++
			result.SpaceFreed += info.Size()
		} else {
			if err := os.Remove(path); err != nil {
				result.Errors = append(result.Errors, fmt.Errorf("remove %s: %w", path, err))
			} else {
				result.FilesDeleted++
				result.SpaceFreed += info.Size()
			}
		}
		return nil
	})
	if err != nil {
		result.Errors = append(result.Errors, err)
	}
	return result
}

// FormatBytes formats a byte count into a human-readable string.
func FormatBytes(bytes int64) string {
	const (
		KB = 1024
		MB = KB * 1024
		GB = MB * 1024
	)
	switch {
	case bytes >= GB:
		return fmt.Sprintf("%.2f GB", float64(bytes)/float64(GB))
	case bytes >= MB:
		return fmt.Sprintf("%.2f MB", float64(bytes)/float64(MB))
	case bytes >= KB:
		return fmt.Sprintf("%.2f KB", float64(bytes)/float64(KB))
	default:
		return fmt.Sprintf("%d B", bytes)
	}
}

func dedup(ss []string) []string {
	seen := map[string]bool{}
	out := []string{}
	for _, s := range ss {
		if s == "" || seen[s] {
			continue
		}
		seen[s] = true
		out = append(out, s)
	}
	return out
}
