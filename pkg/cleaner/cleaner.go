package cleaner

import (
	"fmt"
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
		r := cleanTempFiles(opts.DryRun)
		result.merge(r)
	}
	if opts.Browser {
		r := cleanBrowserCaches(opts.DryRun)
		result.merge(r)
	}
	if opts.Prefetch {
		r := cleanPrefetch(opts.DryRun)
		result.merge(r)
	}
	if opts.Thumbnails {
		r := cleanThumbnails(opts.DryRun)
		result.merge(r)
	}
	if opts.Logs {
		r := cleanLogFiles(opts.DryRun)
		result.merge(r)
	}
	if opts.Registry {
		r := cleanRegistry(opts.DryRun)
		result.merge(r)
	}

	result.Duration = time.Since(start)
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
	} else {
		dirs = append(dirs, os.TempDir())
		dirs = append(dirs, "/tmp")
	}
	return dedup(dirs)
}

func cleanBrowserCaches(dryRun bool) CleanResult {
	result := CleanResult{}
	localAppData := os.Getenv("LOCALAPPDATA")
	appData := os.Getenv("APPDATA")

	if localAppData == "" && appData == "" {
		// Not on Windows or env not set
		return result
	}

	cachePaths := []string{}
	if localAppData != "" {
		cachePaths = append(cachePaths,
			filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Cache"),
			filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Code Cache"),
			filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Cache"),
			filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Code Cache"),
			filepath.Join(localAppData, "Opera Software", "Opera Stable", "Cache"),
		)
	}
	if appData != "" {
		cachePaths = append(cachePaths,
			filepath.Join(appData, "Mozilla", "Firefox", "Profiles"),
		)
	}

	for _, p := range cachePaths {
		if strings.Contains(p, "Firefox") {
			// Firefox stores cache in profile subdirectories
			r := cleanFirefoxCache(p, dryRun)
			result.merge(r)
		} else {
			r := cleanDirectory(p, 0, dryRun)
			result.merge(r)
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
		if !entry.IsDir() && strings.HasPrefix(entry.Name(), "thumbcache_") {
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
