package optimizer

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
)

// Results holds overall optimization results.
type Results struct {
	StartupDisabled        int
	RegistryEntriesRemoved int
}

// StartupResult holds startup optimization results.
type StartupResult struct {
	Disabled int
	Programs []StartupProgram
}

// StartupProgram represents a startup entry.
type StartupProgram struct {
	Name     string
	Path     string
	Impact   string
	Disabled bool
}

// NetworkResult holds network optimization results.
type NetworkResult struct {
	LatencyReduction int
	Optimizations    []string
}

// RegistryResult holds registry optimization results.
type RegistryResult struct {
	EntriesRemoved int
	BackupPath     string
}

// DiskResult holds disk optimization results.
type DiskResult struct {
	IsSSD     bool
	Scheduled bool
}

// OptimizeStartup disables unnecessary startup programs.
func OptimizeStartup() StartupResult {
	return optimizeStartupPlatform()
}

// OptimizeNetwork optimizes network settings for low latency.
func OptimizeNetwork() NetworkResult {
	result := NetworkResult{}

	if runtime.GOOS != "windows" {
		result.Optimizations = append(result.Optimizations, "Network optimization is only available on Windows")
		return result
	}

	commands := []struct {
		args []string
		desc string
	}{
		{[]string{"netsh", "int", "tcp", "set", "global", "autotuninglevel=normal"}, "Set TCP auto-tuning to normal"},
		{[]string{"netsh", "int", "tcp", "set", "global", "chimney=enabled"}, "Enable TCP chimney offload"},
		{[]string{"netsh", "int", "tcp", "set", "global", "dca=enabled"}, "Enable direct cache access"},
		{[]string{"netsh", "int", "tcp", "set", "global", "netdma=enabled"}, "Enable NetDMA"},
		{[]string{"netsh", "int", "tcp", "set", "global", "rss=enabled"}, "Enable receive-side scaling"},
		{[]string{"netsh", "int", "tcp", "set", "heuristics", "disabled"}, "Disable TCP heuristics"},
	}

	for _, c := range commands {
		cmd := exec.Command(c.args[0], c.args[1:]...)
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = getSysProcAttr()
		}
		if err := cmd.Run(); err == nil {
			result.Optimizations = append(result.Optimizations, c.desc)
			result.LatencyReduction += 2
		}
	}

	// Disable network throttling via registry
	if err := setNetworkThrottling(); err == nil {
		result.Optimizations = append(result.Optimizations, "Disabled network throttling")
		result.LatencyReduction += 2
	}

	return result
}

// OptimizeRegistry cleans unnecessary registry entries.
func OptimizeRegistry() RegistryResult {
	return optimizeRegistryPlatform()
}

// OptimizeDisk optimizes disk performance.
func OptimizeDisk() DiskResult {
	result := DiskResult{}

	if runtime.GOOS != "windows" {
		return result
	}

	// Detect SSD
	out, err := exec.Command("powershell", "-Command",
		"Get-PhysicalDisk | Where-Object MediaType -eq 'SSD' | Measure-Object | Select-Object -ExpandProperty Count").Output()
	if err == nil {
		count := strings.TrimSpace(string(out))
		if count != "0" && count != "" {
			result.IsSSD = true
		}
	}

	if result.IsSSD {
		// Enable TRIM for SSD
		cmd := exec.Command("fsutil", "behavior", "set", "DisableDeleteNotify", "0")
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = getSysProcAttr()
		}
		if err := cmd.Run(); err == nil {
			result.Scheduled = true
		}
	} else {
		// Schedule weekly defrag for HDD
		cmd := exec.Command("schtasks", "/create", "/tn", "SysCleanerDefrag",
			"/sc", "weekly", "/d", "SUN", "/st", "03:00",
			"/tr", "defrag C: /O", "/f")
		if runtime.GOOS == "windows" {
			cmd.SysProcAttr = getSysProcAttr()
		}
		if err := cmd.Run(); err == nil {
			result.Scheduled = true
		}
	}

	return result
}

// PrintStartupResult displays startup optimization results.
func PrintStartupResult(result StartupResult) {
	fmt.Printf("  Startup programs disabled: %d\n", result.Disabled)
	for _, p := range result.Programs {
		status := "kept"
		if p.Disabled {
			status = "DISABLED"
		}
		fmt.Printf("    [%s] %s (%s)\n", status, p.Name, p.Impact)
	}
}

// PrintNetworkResult displays network optimization results.
func PrintNetworkResult(result NetworkResult) {
	fmt.Printf("  Estimated latency reduction: %dms\n", result.LatencyReduction)
	for _, opt := range result.Optimizations {
		fmt.Printf("    - %s\n", opt)
	}
}

// PrintRegistryResult displays registry optimization results.
func PrintRegistryResult(result RegistryResult) {
	fmt.Printf("  Registry entries cleaned: %d\n", result.EntriesRemoved)
}

// PrintDiskResult displays disk optimization results.
func PrintDiskResult(result DiskResult) {
	if result.IsSSD {
		fmt.Println("  Disk type: SSD")
		if result.Scheduled {
			fmt.Println("  TRIM enabled for optimal SSD performance")
		}
	} else {
		fmt.Println("  Disk type: HDD")
		if result.Scheduled {
			fmt.Println("  Weekly defragmentation scheduled (Sundays at 3:00 AM)")
		}
	}
}
