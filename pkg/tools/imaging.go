package tools

import (
	"fmt"
	"os/exec"
	"runtime"
)

// ImageScanResult holds the results of a system image scan/repair.
type ImageScanResult struct {
	Healthy        bool
	IssuesFound    int
	IssuesRepaired int
	RebootRequired bool
}

// ScanSystemImage uses DISM to check Windows image health.
func ScanSystemImage() (*ImageScanResult, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("system image scanning only available on Windows")
	}

	result := &ImageScanResult{}

	cmd := exec.Command("DISM.exe", "/Online", "/Cleanup-Image", "/ScanHealth")
	err := cmd.Run()

	result.Healthy = err == nil
	if err != nil {
		result.IssuesFound = 1
	}

	return result, nil
}

// RepairSystemImage uses DISM to repair the Windows image.
func RepairSystemImage() (*ImageScanResult, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("system image repair only available on Windows")
	}

	result := &ImageScanResult{}

	// Run DISM /Online /Cleanup-Image /RestoreHealth
	cmd := exec.Command("DISM.exe", "/Online", "/Cleanup-Image", "/RestoreHealth")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("DISM repair failed: %w", err)
	}

	// Run SFC /scannow after DISM
	sfcCmd := exec.Command("sfc.exe", "/scannow")
	sfcCmd.Run()

	result.IssuesRepaired = 1
	result.Healthy = true
	return result, nil
}
