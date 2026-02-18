package cleaner

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

// CleanOptions specifies what to clean with fine-grained control
type CleanOptions struct {
	// System categories
	WindowsTemp         bool
	UserTemp            bool
	WindowsUpdate       bool
	WindowsInstaller    bool
	Prefetch            bool
	CrashDumps          bool
	ErrorReports        bool
	ThumbnailCache      bool
	IconCache           bool
	FontCache           bool
	ShaderCache         bool
	DNSCache            bool
	WindowsLogs         bool
	EventLogs           bool
	DeliveryOptimization bool
	RecycleBin          bool

	// Application categories
	ChromeCache   bool
	FirefoxCache  bool
	EdgeCache     bool
	BraveCache    bool
	OperaCache    bool
	DiscordCache  bool
	SpotifyCache  bool
	SteamCache    bool
	TeamsCache    bool
	VSCodeCache   bool
	JavaCache     bool

	// Execution options
	DryRun   bool
	Progress ProgressFunc
}

// ProgressFunc is called to report progress during cleaning
type ProgressFunc func(category string, current, total int64)

// ErrorType categorizes cleaning errors for better user feedback
type ErrorType int

const (
	ErrorLocked           ErrorType = iota // File in use
	ErrorPermissionDenied                  // Access denied
	ErrorTimeout                           // Operation timed out
	ErrorNotFound                          // File not found
	ErrorOther                             // Other errors
)

// CleanError is a categorized error for cleaning operations
type CleanError struct {
	Path string
	Type ErrorType
	Err  error
}

func (e *CleanError) Error() string {
	return fmt.Sprintf("%s: %v", e.Path, e.Err)
}

// classifyError categorizes an OS error into a CleanError type
func classifyError(path string, err error) *CleanError {
	ce := &CleanError{Path: path, Err: err}
	errMsg := strings.ToLower(err.Error())
	switch {
	case os.IsPermission(err):
		ce.Type = ErrorPermissionDenied
	case os.IsNotExist(err):
		ce.Type = ErrorNotFound
	case strings.Contains(errMsg, "used by another process") ||
		strings.Contains(errMsg, "locked") ||
		strings.Contains(errMsg, "sharing violation"):
		ce.Type = ErrorLocked
	case strings.Contains(errMsg, "timeout"):
		ce.Type = ErrorTimeout
	default:
		ce.Type = ErrorOther
	}
	return ce
}

// CleanResult holds the result of a cleaning operation
type CleanResult struct {
	FilesDeleted    int64
	SkippedFiles    int64
	SpaceFreed      int64
	LockedFiles     int64
	PermissionFiles int64
	Duration        time.Duration
	Errors          []error
}

const (
	fileTimeout      = 2 * time.Second  // Per-file operation timeout
	dirTimeout       = 30 * time.Second // Per-directory timeout
	defaultOpTimeout = 5 * time.Minute  // Overall operation timeout
)

// cleanTask represents a single cleaning category to execute
type cleanTask struct {
	name string
	fn   func(CleanOptions) CleanResult
}

// maxCleanWorkers is the number of concurrent cleaning goroutines.
// Limited to 4 to avoid excessive disk I/O contention on spinning drives.
const maxCleanWorkers = 4

// PerformClean orchestrates all cleaning operations based on options.
// Independent categories run concurrently via a worker pool for faster execution.
func PerformClean(opts CleanOptions) CleanResult {
	start := time.Now()
	result := CleanResult{}

	ctx, cancel := context.WithTimeout(context.Background(), defaultOpTimeout)
	defer cancel()

	// Build list of enabled categories
	var tasks []cleanTask
	if opts.WindowsTemp {
		tasks = append(tasks, cleanTask{"Windows Temp", cleanWindowsTemp})
	}
	if opts.UserTemp {
		tasks = append(tasks, cleanTask{"User Temp", cleanUserTemp})
	}
	if opts.WindowsUpdate {
		tasks = append(tasks, cleanTask{"Windows Update Cache", cleanWindowsUpdate})
	}
	if opts.WindowsInstaller {
		tasks = append(tasks, cleanTask{"Windows Installer Cache", cleanWindowsInstaller})
	}
	if opts.Prefetch {
		tasks = append(tasks, cleanTask{"Prefetch", cleanPrefetch})
	}
	if opts.CrashDumps {
		tasks = append(tasks, cleanTask{"Crash Dumps", cleanCrashDumps})
	}
	if opts.ErrorReports {
		tasks = append(tasks, cleanTask{"Error Reports", cleanErrorReports})
	}
	if opts.ThumbnailCache {
		tasks = append(tasks, cleanTask{"Thumbnail Cache", cleanThumbnailCache})
	}
	if opts.IconCache {
		tasks = append(tasks, cleanTask{"Icon Cache", cleanIconCache})
	}
	if opts.FontCache {
		tasks = append(tasks, cleanTask{"Font Cache", cleanFontCache})
	}
	if opts.ShaderCache {
		tasks = append(tasks, cleanTask{"Shader Cache", cleanShaderCache})
	}
	if opts.DNSCache {
		tasks = append(tasks, cleanTask{"DNS Cache", cleanDNSCache})
	}
	if opts.WindowsLogs {
		tasks = append(tasks, cleanTask{"Windows Log Files", cleanWindowsLogs})
	}
	if opts.EventLogs {
		tasks = append(tasks, cleanTask{"Event Logs", cleanEventLogs})
	}
	if opts.DeliveryOptimization {
		tasks = append(tasks, cleanTask{"Delivery Optimization", cleanDeliveryOptimization})
	}
	if opts.RecycleBin {
		tasks = append(tasks, cleanTask{"Recycle Bin", cleanRecycleBin})
	}
	if opts.ChromeCache {
		tasks = append(tasks, cleanTask{"Chrome Cache", cleanChromeCache})
	}
	if opts.FirefoxCache {
		tasks = append(tasks, cleanTask{"Firefox Cache", cleanFirefoxCache})
	}
	if opts.EdgeCache {
		tasks = append(tasks, cleanTask{"Edge Cache", cleanEdgeCache})
	}
	if opts.BraveCache {
		tasks = append(tasks, cleanTask{"Brave Cache", cleanBraveCache})
	}
	if opts.OperaCache {
		tasks = append(tasks, cleanTask{"Opera Cache", cleanOperaCache})
	}
	if opts.DiscordCache {
		tasks = append(tasks, cleanTask{"Discord Cache", cleanDiscordCache})
	}
	if opts.SpotifyCache {
		tasks = append(tasks, cleanTask{"Spotify Cache", cleanSpotifyCache})
	}
	if opts.SteamCache {
		tasks = append(tasks, cleanTask{"Steam Cache", cleanSteamCache})
	}
	if opts.TeamsCache {
		tasks = append(tasks, cleanTask{"Teams Cache", cleanTeamsCache})
	}
	if opts.VSCodeCache {
		tasks = append(tasks, cleanTask{"VS Code Cache", cleanVSCodeCache})
	}
	if opts.JavaCache {
		tasks = append(tasks, cleanTask{"Java Cache", cleanJavaCache})
	}

	if len(tasks) == 0 {
		result.Duration = time.Since(start)
		return result
	}

	// Run categories concurrently via worker pool
	taskCh := make(chan cleanTask, len(tasks))
	resultCh := make(chan CleanResult, len(tasks))

	workers := maxCleanWorkers
	if len(tasks) < workers {
		workers = len(tasks)
	}

	var wg sync.WaitGroup
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for task := range taskCh {
				resultCh <- cleanCategory(ctx, task.name, task.fn, opts)
			}
		}()
	}

	for _, t := range tasks {
		taskCh <- t
	}
	close(taskCh)

	// Close results channel once all workers finish
	go func() {
		wg.Wait()
		close(resultCh)
	}()

	for r := range resultCh {
		result.merge(r)
	}

	result.Duration = time.Since(start)
	log.Printf("[SysCleaner] Cleanup complete: %d files deleted, %d skipped, %s freed in %s",
		result.FilesDeleted, result.SkippedFiles, FormatBytes(result.SpaceFreed), result.Duration.Round(time.Millisecond))
	return result
}

func (r *CleanResult) merge(other CleanResult) {
	r.FilesDeleted += other.FilesDeleted
	r.SkippedFiles += other.SkippedFiles
	r.SpaceFreed += other.SpaceFreed
	r.LockedFiles += other.LockedFiles
	r.PermissionFiles += other.PermissionFiles
	r.Errors = append(r.Errors, other.Errors...)
}

// cleanCategory runs a category cleaning function with timeout and progress reporting
func cleanCategory(ctx context.Context, category string, fn func(CleanOptions) CleanResult, opts CleanOptions) CleanResult {
	log.Printf("[SysCleaner] Cleaning %s...", category)

	if opts.Progress != nil {
		opts.Progress(category, 0, 100)
	}

	done := make(chan CleanResult, 1)
	go func() {
		done <- fn(opts)
	}()

	select {
	case result := <-done:
		if opts.Progress != nil {
			opts.Progress(category, 100, 100)
		}
		return result
	case <-ctx.Done():
		log.Printf("[SysCleaner] %s cleaning timed out", category)
		return CleanResult{Errors: []error{fmt.Errorf("%s cleaning timed out", category)}}
	}
}

// removeWithTimeout attempts to remove a file with a timeout
func removeWithTimeout(path string, timeout time.Duration) error {
	done := make(chan error, 1)
	go func() {
		done <- os.Remove(path)
	}()

	select {
	case err := <-done:
		return err
	case <-time.After(timeout):
		return fmt.Errorf("timeout removing %s", path)
	}
}

// cleanDirectory removes files in a directory with timeouts and proper error handling
func cleanDirectory(dir string, maxAge time.Duration, dryRun bool) CleanResult {
	result := CleanResult{}

	if _, err := os.Stat(dir); os.IsNotExist(err) {
		return result
	}

	ctx, cancel := context.WithTimeout(context.Background(), dirTimeout)
	defer cancel()

	done := make(chan CleanResult, 1)
	go func() {
		r := cleanDirectoryInternal(dir, maxAge, dryRun)
		done <- r
	}()

	select {
	case r := <-done:
		return r
	case <-ctx.Done():
		log.Printf("[SysCleaner] Directory cleanup timed out: %s", dir)
		result.Errors = append(result.Errors, fmt.Errorf("timeout cleaning %s", dir))
		return result
	}
}

func cleanDirectoryInternal(dir string, maxAge time.Duration, dryRun bool) CleanResult {
	result := CleanResult{}
	now := time.Now()

	err := filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			// Skip inaccessible directories gracefully
			if d != nil && d.IsDir() {
				log.Printf("[SysCleaner] Skipping inaccessible directory: %s", path)
				return filepath.SkipDir
			}
			result.Errors = append(result.Errors, fmt.Errorf("walk error %s: %w", path, err))
			return nil
		}

		if d.IsDir() {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			result.Errors = append(result.Errors, err)
			return nil
		}

		// Skip files newer than maxAge if specified
		if maxAge > 0 && now.Sub(info.ModTime()) < maxAge {
			return nil
		}

		if dryRun {
			result.FilesDeleted++
			result.SpaceFreed += info.Size()
		} else {
			if err := removeWithTimeout(path, fileTimeout); err != nil {
				ce := classifyError(path, err)
				switch ce.Type {
				case ErrorLocked, ErrorTimeout:
					result.SkippedFiles++
					result.LockedFiles++
				case ErrorPermissionDenied:
					result.SkippedFiles++
					result.PermissionFiles++
				default:
					result.Errors = append(result.Errors, ce)
				}
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

// System category cleaners
func cleanWindowsTemp(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(winDir, "Temp"), 0, opts.DryRun)
}

func cleanUserTemp(opts CleanOptions) CleanResult {
	result := CleanResult{}
	tempDirs := []string{os.Getenv("TEMP"), os.Getenv("TMP")}
	if runtime.GOOS == "windows" {
		localAppData := os.Getenv("LOCALAPPDATA")
		if localAppData != "" {
			tempDirs = append(tempDirs, filepath.Join(localAppData, "Temp"))
		}
	}
	for _, dir := range dedup(tempDirs) {
		if dir != "" {
			result.merge(cleanDirectory(dir, 0, opts.DryRun))
		}
	}
	return result
}

func cleanWindowsUpdate(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(winDir, "SoftwareDistribution", "Download"), 0, opts.DryRun)
}

func cleanWindowsInstaller(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(winDir, "Installer", "$PatchCache$"), 0, opts.DryRun)
}

func cleanPrefetch(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return CleanResult{}
	}
	// Only clean prefetch files older than 30 days
	return cleanDirectory(filepath.Join(winDir, "Prefetch"), 30*24*time.Hour, opts.DryRun)
}

func cleanCrashDumps(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	winDir := os.Getenv("WINDIR")

	dirs := []string{}
	if localAppData != "" {
		dirs = append(dirs, filepath.Join(localAppData, "CrashDumps"))
	}
	if winDir != "" {
		dirs = append(dirs, filepath.Join(winDir, "Minidump"))
		// Windows memory dump file
		memoryDump := filepath.Join(winDir, "MEMORY.DMP")
		if info, err := os.Stat(memoryDump); err == nil {
			if opts.DryRun {
				result.FilesDeleted++
				result.SpaceFreed += info.Size()
			} else {
				if err := removeWithTimeout(memoryDump, fileTimeout); err == nil {
					result.FilesDeleted++
					result.SpaceFreed += info.Size()
				}
			}
		}
	}

	for _, dir := range dirs {
		result.merge(cleanDirectory(dir, 0, opts.DryRun))
	}
	return result
}

func cleanErrorReports(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	programData := os.Getenv("ProgramData")

	dirs := []string{}
	if localAppData != "" {
		dirs = append(dirs, filepath.Join(localAppData, "Microsoft", "Windows", "WER"))
	}
	if programData != "" {
		dirs = append(dirs, filepath.Join(programData, "Microsoft", "Windows", "WER"))
	}

	for _, dir := range dirs {
		result.merge(cleanDirectory(dir, 0, opts.DryRun))
	}
	return result
}

func cleanThumbnailCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return result
	}

	thumbDir := filepath.Join(localAppData, "Microsoft", "Windows", "Explorer")
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
			if opts.DryRun {
				result.FilesDeleted++
				result.SpaceFreed += info.Size()
			} else {
				if err := removeWithTimeout(fpath, fileTimeout); err != nil {
					if strings.Contains(err.Error(), "timeout") {
						result.SkippedFiles++
					} else {
						result.Errors = append(result.Errors, err)
					}
				} else {
					result.FilesDeleted++
					result.SpaceFreed += info.Size()
				}
			}
		}
	}
	return result
}

func cleanIconCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return result
	}

	iconCacheFile := filepath.Join(localAppData, "IconCache.db")
	if info, err := os.Stat(iconCacheFile); err == nil {
		if opts.DryRun {
			result.FilesDeleted++
			result.SpaceFreed += info.Size()
		} else {
			if err := removeWithTimeout(iconCacheFile, fileTimeout); err == nil {
				result.FilesDeleted++
				result.SpaceFreed += info.Size()
			}
		}
	}
	return result
}

func cleanFontCache(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(winDir, "ServiceProfiles", "LocalService", "AppData", "Local", "FontCache"), 0, opts.DryRun)
}

func cleanShaderCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return result
	}

	shaderDirs := []string{
		filepath.Join(localAppData, "D3DSCache"),
		filepath.Join(localAppData, "NVIDIA", "DXCache"),
		filepath.Join(localAppData, "NVIDIA", "GLCache"),
		filepath.Join(localAppData, "AMD", "DxCache"),
	}

	for _, dir := range shaderDirs {
		result.merge(cleanDirectory(dir, 0, opts.DryRun))
	}
	return result
}

func cleanDNSCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" || opts.DryRun {
		return result
	}

	if err := flushDNSCacheNative(); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to flush DNS cache: %w", err))
	} else {
		log.Println("[SysCleaner] DNS cache flushed")
	}
	return result
}

func cleanWindowsLogs(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return result
	}

	logDirs := []string{
		filepath.Join(winDir, "Logs"),
		filepath.Join(winDir, "Debug"),
		filepath.Join(winDir, "Panther"),
	}

	for _, dir := range logDirs {
		result.merge(cleanDirectory(dir, 30*24*time.Hour, opts.DryRun))
	}
	return result
}

func cleanEventLogs(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" || opts.DryRun {
		return result
	}

	// NOTE: Security event log is intentionally excluded — clearing it is an
	// anti-forensics indicator that triggers AV heuristics.
	logs := []string{"System", "Application"}
	for _, logName := range logs {
		cmd := exec.Command("wevtutil", "cl", logName)
		if err := cmd.Run(); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to clear %s event log: %w", logName, err))
		} else {
			log.Printf("[SysCleaner] Cleared %s event log", logName)
		}
	}
	return result
}

func cleanDeliveryOptimization(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	winDir := os.Getenv("WINDIR")
	if winDir == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(winDir, "SoftwareDistribution", "DeliveryOptimization"), 0, opts.DryRun)
}

func cleanRecycleBin(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" || opts.DryRun {
		return result
	}

	// Use PowerShell to clear recycle bin for all drives
	cmd := exec.Command("powershell", "-Command", "Clear-RecycleBin -Force -ErrorAction SilentlyContinue")
	if err := cmd.Run(); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to clear recycle bin: %w", err))
	} else {
		log.Println("[SysCleaner] Recycle Bin cleared")
	}
	return result
}

// Application category cleaners
func cleanChromiumProfiles(userDataDir string, dryRun bool) CleanResult {
	result := CleanResult{}
	if _, err := os.Stat(userDataDir); os.IsNotExist(err) {
		return result
	}

	cacheSubdirs := []string{"Cache", "Code Cache", "GPUCache", "Service Worker", "ShaderCache"}

	entries, err := os.ReadDir(userDataDir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == "Default" || strings.HasPrefix(name, "Profile ") {
			for _, sub := range cacheSubdirs {
				cacheDir := filepath.Join(userDataDir, name, sub)
				result.merge(cleanDirectory(cacheDir, 0, dryRun))
			}
		}
	}
	return result
}

func cleanChromeCache(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return CleanResult{}
	}
	return cleanChromiumProfiles(filepath.Join(localAppData, "Google", "Chrome", "User Data"), opts.DryRun)
}

func cleanFirefoxCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return result
	}

	profilesDir := filepath.Join(appData, "Mozilla", "Firefox", "Profiles")
	entries, err := os.ReadDir(profilesDir)
	if err != nil {
		return result
	}

	for _, entry := range entries {
		if entry.IsDir() {
			result.merge(cleanDirectory(filepath.Join(profilesDir, entry.Name(), "cache2"), 0, opts.DryRun))
			result.merge(cleanDirectory(filepath.Join(profilesDir, entry.Name(), "startupCache"), 0, opts.DryRun))
		}
	}
	return result
}

func cleanEdgeCache(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return CleanResult{}
	}
	return cleanChromiumProfiles(filepath.Join(localAppData, "Microsoft", "Edge", "User Data"), opts.DryRun)
}

func cleanBraveCache(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return CleanResult{}
	}
	return cleanChromiumProfiles(filepath.Join(localAppData, "BraveSoftware", "Brave-Browser", "User Data"), opts.DryRun)
}

func cleanOperaCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return result
	}

	operaDirs := []string{
		filepath.Join(appData, "Opera Software", "Opera Stable"),
		filepath.Join(appData, "Opera Software", "Opera GX Stable"),
	}

	for _, dir := range operaDirs {
		result.merge(cleanChromiumProfiles(dir, opts.DryRun))
	}
	return result
}

func cleanDiscordCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return result
	}

	discordDirs := []string{
		filepath.Join(appData, "discord", "Cache"),
		filepath.Join(appData, "discord", "Code Cache"),
		filepath.Join(appData, "discord", "GPUCache"),
	}

	for _, dir := range discordDirs {
		result.merge(cleanDirectory(dir, 0, opts.DryRun))
	}
	return result
}

func cleanSpotifyCache(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(localAppData, "Spotify", "Storage"), 0, opts.DryRun)
}

func cleanSteamCache(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(localAppData, "Steam", "htmlcache"), 0, opts.DryRun)
}

func cleanTeamsCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return result
	}

	teamsDirs := []string{
		filepath.Join(appData, "Microsoft", "Teams", "Cache"),
		filepath.Join(appData, "Microsoft", "Teams", "blob_storage"),
		filepath.Join(appData, "Microsoft", "Teams", "GPUCache"),
	}

	for _, dir := range teamsDirs {
		result.merge(cleanDirectory(dir, 0, opts.DryRun))
	}
	return result
}

func cleanVSCodeCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" {
		return result
	}
	appData := os.Getenv("APPDATA")
	if appData == "" {
		return result
	}

	vscodeDirs := []string{
		filepath.Join(appData, "Code", "Cache"),
		filepath.Join(appData, "Code", "CachedData"),
		filepath.Join(appData, "Code", "CachedExtensions"),
	}

	for _, dir := range vscodeDirs {
		result.merge(cleanDirectory(dir, 0, opts.DryRun))
	}
	return result
}

func cleanJavaCache(opts CleanOptions) CleanResult {
	if runtime.GOOS != "windows" {
		return CleanResult{}
	}
	userProfile := os.Getenv("USERPROFILE")
	if userProfile == "" {
		return CleanResult{}
	}
	return cleanDirectory(filepath.Join(userProfile, "AppData", "LocalLow", "Sun", "Java", "Deployment", "cache"), 0, opts.DryRun)
}

// FormatBytes formats a byte count into a human-readable string
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
