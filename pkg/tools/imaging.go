package tools

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"syscleaner/pkg/admin"
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

	if err := admin.RequireElevation("System Image Scan"); err != nil {
		return nil, err
	}

	result := &ImageScanResult{}

	log.Println("[SysCleaner] Running DISM /ScanHealth...")
	cmd := exec.Command("DISM.exe", "/Online", "/Cleanup-Image", "/ScanHealth")
	err := cmd.Run()

	result.Healthy = err == nil
	if err != nil {
		result.IssuesFound = 1
		log.Println("[SysCleaner] DISM scan found potential issues.")
	} else {
		log.Println("[SysCleaner] DISM scan complete — image is healthy.")
	}

	return result, nil
}

// RepairSystemImage uses DISM to repair the Windows image.
func RepairSystemImage() (*ImageScanResult, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("system image repair only available on Windows")
	}

	if err := admin.RequireElevation("System Image Repair"); err != nil {
		return nil, err
	}

	result := &ImageScanResult{}

	// Run DISM /Online /Cleanup-Image /RestoreHealth
	log.Println("[SysCleaner] Running DISM /RestoreHealth (this may take a while)...")
	cmd := exec.Command("DISM.exe", "/Online", "/Cleanup-Image", "/RestoreHealth")
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("DISM repair failed: %w", err)
	}

	// Run SFC /scannow after DISM
	log.Println("[SysCleaner] Running SFC /scannow...")
	sfcCmd := exec.Command("sfc.exe", "/scannow")
	sfcCmd.Run()

	result.IssuesRepaired = 1
	result.Healthy = true
	log.Println("[SysCleaner] System image repair complete.")
	return result, nil
}
