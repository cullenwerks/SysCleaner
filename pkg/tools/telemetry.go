package tools

import (
	"fmt"
	"os/exec"
	"runtime"
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

	result := &TelemetryResult{}

	// Disable telemetry services
	telemetryServices := []string{
		"DiagTrack",
		"dmwappushservice",
		"RetailDemo",
	}

	for _, svc := range telemetryServices {
		cmd := exec.Command("sc", "config", svc, "start=disabled")
		if cmd.Run() == nil {
			result.ServicesDisabled++
		}
		exec.Command("sc", "stop", svc).Run()
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

	for _, task := range telemetryTasks {
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

	for _, entry := range regEntries {
		cmd := exec.Command("reg", "add", entry.key, "/v", entry.name,
			"/t", "REG_DWORD", "/d", entry.value, "/f")
		if cmd.Run() == nil {
			result.RegistryChanged++
		}
	}

	return result, nil
}

// GetTelemetryStatus returns a summary of what will be disabled.
func GetTelemetryStatus() string {
	return fmt.Sprintf("Will disable: 3 services, 6 scheduled tasks, 2 registry entries")
}
