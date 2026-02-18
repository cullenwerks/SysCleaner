package scheduler

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// ScheduleConfig holds configuration for a scheduled weekly clean.
type ScheduleConfig struct {
	Enabled     bool
	DayOfWeek   string
	Hour        int
	CleanPreset string
}

// CreateScheduledClean registers a weekly Windows scheduled task that runs
// SysCleaner in headless mode with the given clean preset.
func CreateScheduledClean(cfg ScheduleConfig) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("scheduled cleaning only available on Windows")
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	return createScheduledCleanNative(cfg, exePath)
}

// RemoveScheduledClean deletes the SysCleanerWeeklyClean scheduled task.
func RemoveScheduledClean() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("scheduled cleaning only available on Windows")
	}
	return removeScheduledCleanNative()
}

// GetScheduledClean queries Windows Task Scheduler for the
// SysCleanerWeeklyClean task. Returns nil config (no error) if not found.
func GetScheduledClean() (*ScheduleConfig, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("scheduled cleaning only available on Windows")
	}
	return getScheduledCleanNative()
}

// parseCleanPreset extracts the clean preset from the task command string.
func parseCleanPreset(taskCmd string) string {
	parts := strings.Fields(taskCmd)
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if strings.HasPrefix(p, "--") && p != "--headless" && p != "--clean" {
			return strings.TrimPrefix(p, "--")
		}
	}
	return "all"
}
