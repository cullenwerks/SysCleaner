package optimizer

import (
	"fmt"
	"runtime"
)

// Results holds overall optimization results.
type Results struct {
	StartupDisabled int
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

	for _, opt := range tcpOptimizations {
		if err := setTCPOptimizationParam(opt.name, opt.val); err == nil {
			result.Optimizations = append(result.Optimizations, opt.desc)
			result.LatencyReduction += 2
		}
	}

	if err := setNetworkThrottling(); err == nil {
		result.Optimizations = append(result.Optimizations, "Disabled network throttling")
		result.LatencyReduction += 2
	}

	return result
}

// OptimizeDisk optimizes disk performance.
func OptimizeDisk() DiskResult {
	result := DiskResult{}

	if runtime.GOOS != "windows" {
		return result
	}

	result.IsSSD = isSSDPresentNative()

	if result.IsSSD {
		if err := enableTRIMNative(); err == nil {
			result.Scheduled = true
		}
	} else {
		if err := createDefragTask(); err == nil {
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
