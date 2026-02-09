package tools

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"syscleaner/pkg/admin"
)

// TelemetryResult holds the results of telemetry disabling.
type TelemetryResult struct {
	ServicesDisabled int
	TasksDisabled    int
	RegistryChanged  int
}

// DisableTelemetry disables Windows telemetry and tracking.
func DisableTelemetry() (*TelemetryResult, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("telemetry disabling only available on Windows")
	}

	if err := admin.RequireElevation("Telemetry Management"); err != nil {
		return nil, err
	}

	result := &TelemetryResult{}

	// Disable telemetry services
	telemetryServices := []string{
		"DiagTrack",
		"dmwappushservice",
		"RetailDemo",
	}

	log.Println("[SysCleaner] Disabling telemetry services...")
	for _, svc := range telemetryServices {
		log.Printf("[SysCleaner] Disabling service: %s", svc)
		cmd := exec.Command("sc", "config", svc, "start=disabled")
		if cmd.Run() == nil {
			result.ServicesDisabled++
		}
		exec.Command("net", "stop", svc).Run()
	}

	// Disable telemetry scheduled tasks
	telemetryTasks := []string{
		"\\Microsoft\\Windows\\Application Experience\\Microsoft Compatibility Appraiser",
		"\\Microsoft\\Windows\\Application Experience\\ProgramDataUpdater",
		"\\Microsoft\\Windows\\Autochk\\Proxy",
		"\\Microsoft\\Windows\\Customer Experience Improvement Program\\Consolidator",
		"\\Microsoft\\Windows\\Customer Experience Improvement Program\\UsbCeip",
		"\\Microsoft\\Windows\\DiskDiagnostic\\Microsoft-Windows-DiskDiagnosticDataCollector",
	}

	log.Println("[SysCleaner] Disabling telemetry scheduled tasks...")
	for _, task := range telemetryTasks {
		log.Printf("[SysCleaner] Disabling task: %s", task)
		cmd := exec.Command("schtasks", "/Change", "/TN", task, "/Disable")
		if cmd.Run() == nil {
			result.TasksDisabled++
		}
	}

	// Registry changes to disable telemetry
	regEntries := []struct {
		key   string
		name  string
		value string
	}{
		{"HKLM\\SOFTWARE\\Policies\\Microsoft\\Windows\\DataCollection", "AllowTelemetry", "0"},
		{"HKLM\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Policies\\DataCollection", "AllowTelemetry", "0"},
	}

	log.Println("[SysCleaner] Applying telemetry registry settings...")
	for _, entry := range regEntries {
		log.Printf("[SysCleaner] Setting %s\\%s = %s", entry.key, entry.name, entry.value)
		cmd := exec.Command("reg", "add", entry.key, "/v", entry.name,
			"/t", "REG_DWORD", "/d", entry.value, "/f")
		if cmd.Run() == nil {
			result.RegistryChanged++
		}
	}

	log.Printf("[SysCleaner] Telemetry management complete: %d services, %d tasks, %d registry entries",
		result.ServicesDisabled, result.TasksDisabled, result.RegistryChanged)
	return result, nil
}

// GetTelemetryStatus returns a summary of what will be disabled.
func GetTelemetryStatus() string {
	return fmt.Sprintf("Will disable: 3 services, 6 scheduled tasks, 2 registry entries")
}
