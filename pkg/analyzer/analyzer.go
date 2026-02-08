package analyzer

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

// Report holds the complete system analysis report.
type Report struct {
	System           SystemInfo
	PerformanceScore int
	Issues           []Issue
	Recommendations  []Recommendation
	ReclaimableSpace int64
	SpaceBreakdown   map[string]int64
	StartupPrograms  []StartupProgramInfo
}

// SystemInfo holds basic system information.
type SystemInfo struct {
	OS       string
	CPU      string
	CPUCores int
	TotalRAM uint64
	TotalDisk uint64
	FreeDisk  uint64
}

// Issue represents a detected performance issue.
type Issue struct {
	Description string
	Impact      string
	Severity    string // "High", "Medium", "Low"
}

// Recommendation represents an actionable optimization.
type Recommendation struct {
	Title        string
	ExpectedGain string
	Command      string
}

// StartupProgramInfo represents a startup program entry.
type StartupProgramInfo struct {
	Name        string
	Impact      string
	Recommended bool
}

// AnalyzeSystem performs a comprehensive system analysis.
func AnalyzeSystem() Report {
	report := Report{
		SpaceBreakdown: make(map[string]int64),
	}

	// Gather system info
	report.System = gatherSystemInfo()

	// Analyze disk space
	analyzeDiskSpace(&report)

	// Analyze startup programs
	report.StartupPrograms = analyzeStartupPrograms()

	// Detect issues
	detectIssues(&report)

	// Generate recommendations
	generateRecommendations(&report)

	// Calculate score
	report.PerformanceScore = calculateScore(&report)

	return report
}

func gatherSystemInfo() SystemInfo {
	info := SystemInfo{
		OS:       runtime.GOOS,
		CPUCores: runtime.NumCPU(),
	}

	if cpuInfo, err := cpu.Info(); err == nil && len(cpuInfo) > 0 {
		info.CPU = cpuInfo[0].ModelName
	}

	if vmem, err := mem.VirtualMemory(); err == nil {
		info.TotalRAM = vmem.Total
	}

	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	if usage, err := disk.Usage(diskPath); err == nil {
		info.TotalDisk = usage.Total
		info.FreeDisk = usage.Free
	}

	return info
}

func analyzeDiskSpace(report *Report) {
	tempSize := calculateTempSize()
	report.SpaceBreakdown["Temp Files"] = tempSize
	report.ReclaimableSpace += tempSize

	browserSize := calculateBrowserCacheSize()
	report.SpaceBreakdown["Browser Caches"] = browserSize
	report.ReclaimableSpace += browserSize

	logSize := calculateLogSize()
	report.SpaceBreakdown["Log Files"] = logSize
	report.ReclaimableSpace += logSize
}

func calculateTempSize() int64 {
	var total int64
	tempDirs := []string{os.TempDir()}
	if runtime.GOOS == "windows" {
		if t := os.Getenv("TEMP"); t != "" {
			tempDirs = append(tempDirs, t)
		}
		if winDir := os.Getenv("WINDIR"); winDir != "" {
			tempDirs = append(tempDirs, filepath.Join(winDir, "Temp"))
		}
	}
	seen := map[string]bool{}
	for _, dir := range tempDirs {
		if seen[dir] || dir == "" {
			continue
		}
		seen[dir] = true
		total += calculateDirectorySize(dir)
	}
	return total
}

func calculateBrowserCacheSize() int64 {
	var total int64
	localAppData := os.Getenv("LOCALAPPDATA")
	if localAppData == "" {
		return 0
	}
	cachePaths := []string{
		filepath.Join(localAppData, "Google", "Chrome", "User Data", "Default", "Cache"),
		filepath.Join(localAppData, "Microsoft", "Edge", "User Data", "Default", "Cache"),
		filepath.Join(localAppData, "Opera Software", "Opera Stable", "Cache"),
	}
	for _, p := range cachePaths {
		total += calculateDirectorySize(p)
	}
	return total
}

func calculateLogSize() int64 {
	var total int64
	if winDir := os.Getenv("WINDIR"); winDir != "" {
		total += calculateDirectorySize(filepath.Join(winDir, "Logs"))
	}
	if pd := os.Getenv("ProgramData"); pd != "" {
		total += calculateDirectorySize(filepath.Join(pd, "Logs"))
	}
	return total
}

func detectIssues(report *Report) {
	// Check RAM usage
	if vmem, err := mem.VirtualMemory(); err == nil {
		if vmem.UsedPercent > 85 {
			report.Issues = append(report.Issues, Issue{
				Description: fmt.Sprintf("High RAM usage: %.1f%%", vmem.UsedPercent),
				Impact:      "System may become slow and unresponsive",
				Severity:    "High",
			})
		}
	}

	// Check disk usage
	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	if usage, err := disk.Usage(diskPath); err == nil {
		if usage.UsedPercent > 90 {
			report.Issues = append(report.Issues, Issue{
				Description: fmt.Sprintf("Critically low disk space: %.1f%% used", usage.UsedPercent),
				Impact:      "System performance severely degraded, risk of crashes",
				Severity:    "High",
			})
		} else if usage.UsedPercent > 80 {
			report.Issues = append(report.Issues, Issue{
				Description: fmt.Sprintf("Low disk space: %.1f%% used", usage.UsedPercent),
				Impact:      "May cause slowdowns as disk fills up",
				Severity:    "Medium",
			})
		}
	}

	// Check reclaimable space
	if report.ReclaimableSpace > 1024*1024*1024 { // > 1GB
		report.Issues = append(report.Issues, Issue{
			Description: fmt.Sprintf("%s of reclaimable disk space found", formatBytes(report.ReclaimableSpace)),
			Impact:      "Wasted disk space from temp files and caches",
			Severity:    "Medium",
		})
	}

	// Check startup programs
	if len(report.StartupPrograms) > 10 {
		report.Issues = append(report.Issues, Issue{
			Description: fmt.Sprintf("Too many startup programs (%d)", len(report.StartupPrograms)),
			Impact:      "Slow boot times and high background resource usage",
			Severity:    "Medium",
		})
	}
}

func generateRecommendations(report *Report) {
	if report.ReclaimableSpace > 100*1024*1024 { // > 100MB
		report.Recommendations = append(report.Recommendations, Recommendation{
			Title:        "Clean temporary files and caches",
			ExpectedGain: fmt.Sprintf("Free %s of disk space", formatBytes(report.ReclaimableSpace)),
			Command:      "syscleaner clean --all",
		})
	}

	unncessaryStartup := 0
	for _, p := range report.StartupPrograms {
		if !p.Recommended {
			unncessaryStartup++
		}
	}
	if unncessaryStartup > 0 {
		report.Recommendations = append(report.Recommendations, Recommendation{
			Title:        "Optimize startup programs",
			ExpectedGain: fmt.Sprintf("Disable %d unnecessary startup programs", unncessaryStartup),
			Command:      "syscleaner optimize --startup",
		})
	}

	report.Recommendations = append(report.Recommendations, Recommendation{
		Title:        "Enable gaming mode for better game performance",
		ExpectedGain: "Reduced input lag, higher FPS, less background interference",
		Command:      "syscleaner gaming --enable",
	})
}

func calculateScore(report *Report) int {
	score := 100

	for _, issue := range report.Issues {
		switch issue.Severity {
		case "High":
			score -= 15
		case "Medium":
			score -= 10
		case "Low":
			score -= 5
		}
	}

	// Deduct for disk usage
	diskPath := "/"
	if runtime.GOOS == "windows" {
		diskPath = "C:\\"
	}
	if usage, err := disk.Usage(diskPath); err == nil {
		if usage.UsedPercent > 80 {
			score -= int(usage.UsedPercent-80) / 2
		}
	}

	// Deduct for too many startup programs
	if len(report.StartupPrograms) > 5 {
		score -= (len(report.StartupPrograms) - 5)
	}

	if score < 0 {
		score = 0
	}
	return score
}

func calculateDirectorySize(dir string) int64 {
	var total int64
	filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !d.IsDir() {
			if info, err := d.Info(); err == nil {
				total += info.Size()
			}
		}
		return nil
	})
	return total
}

// PrintReport outputs the analysis report.
func PrintReport(report Report) {
	fmt.Println("=== SysCleaner System Analysis ===")
	fmt.Println()

	// System Info
	fmt.Println("--- System Information ---")
	fmt.Printf("  OS:        %s\n", report.System.OS)
	fmt.Printf("  CPU:       %s\n", report.System.CPU)
	fmt.Printf("  Cores:     %d\n", report.System.CPUCores)
	fmt.Printf("  RAM:       %s\n", formatBytes(int64(report.System.TotalRAM)))
	fmt.Printf("  Disk:      %s total, %s free\n",
		formatBytes(int64(report.System.TotalDisk)),
		formatBytes(int64(report.System.FreeDisk)))
	fmt.Println()

	// Performance Score
	fmt.Println("--- Performance Score ---")
	fmt.Printf("  Score: %d/100  %s\n", report.PerformanceScore, printScoreBar(report.PerformanceScore))
	fmt.Printf("  %s\n", scoreLabel(report.PerformanceScore))
	fmt.Println()

	// Issues
	if len(report.Issues) > 0 {
		fmt.Println("--- Issues Found ---")
		for i, issue := range report.Issues {
			severity := strings.ToUpper(issue.Severity)
			fmt.Printf("  %d. [%s] %s\n", i+1, severity, issue.Description)
			fmt.Printf("     Impact: %s\n", issue.Impact)
		}
		fmt.Println()
	}

	// Reclaimable Space
	if report.ReclaimableSpace > 0 {
		fmt.Println("--- Reclaimable Disk Space ---")
		fmt.Printf("  Total: %s\n", formatBytes(report.ReclaimableSpace))
		for name, size := range report.SpaceBreakdown {
			if size > 0 {
				fmt.Printf("    %s: %s\n", name, formatBytes(size))
			}
		}
		fmt.Println()
	}

	// Startup Programs
	if len(report.StartupPrograms) > 0 {
		fmt.Println("--- Startup Programs ---")
		fmt.Printf("  Total: %d\n", len(report.StartupPrograms))
		for _, p := range report.StartupPrograms {
			rec := " "
			if !p.Recommended {
				rec = "*"
			}
			fmt.Printf("  %s %s (%s)\n", rec, p.Name, p.Impact)
		}
		fmt.Println("  (* = can be disabled for faster boot)")
		fmt.Println()
	}

	// Recommendations
	if len(report.Recommendations) > 0 {
		fmt.Println("--- Recommendations ---")
		for i, rec := range report.Recommendations {
			fmt.Printf("  %d. %s\n", i+1, rec.Title)
			fmt.Printf("     Expected: %s\n", rec.ExpectedGain)
			fmt.Printf("     Run:      %s\n", rec.Command)
		}
		fmt.Println()
	}
}

func printScoreBar(score int) string {
	filled := score / 5
	empty := 20 - filled
	bar := "[" + strings.Repeat("#", filled) + strings.Repeat("-", empty) + "]"
	return bar
}

func scoreLabel(score int) string {
	switch {
	case score >= 90:
		return "Excellent - Your system is well optimized"
	case score >= 70:
		return "Good - Minor optimizations available"
	case score >= 50:
		return "Fair - Several optimizations recommended"
	case score >= 30:
		return "Poor - Significant optimization needed"
	default:
		return "Critical - Immediate action recommended"
	}
}

func formatBytes(bytes int64) string {
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

// analyzeStartupPrograms reads startup entries - platform specific
func analyzeStartupPrograms() []StartupProgramInfo {
	// Time limit the analysis
	ch := make(chan []StartupProgramInfo, 1)
	go func() {
		ch <- analyzeStartupProgramsPlatform()
	}()
	select {
	case result := <-ch:
		return result
	case <-time.After(5 * time.Second):
		return nil
	}
}
