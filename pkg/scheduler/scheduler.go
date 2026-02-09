package scheduler

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
)

// ScheduleConfig holds configuration for a scheduled weekly clean.
type ScheduleConfig struct {
	Enabled     bool
	DayOfWeek   string // e.g., "Sunday"
	Hour        int
	CleanPreset string // e.g., "system", "all", "browsers"
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

	// Build the action argument based on the clean preset.
	action := fmt.Sprintf("%s --headless --clean --%s", exePath, cfg.CleanPreset)

	// Format the start time as HH:00.
	startTime := fmt.Sprintf("%02d:00", cfg.Hour)

	// schtasks /create
	//   /tn  — task name
	//   /tr  — command to run
	//   /sc  — schedule type (weekly)
	//   /d   — day of week
	//   /st  — start time
	//   /f   — force overwrite if task already exists
	cmd := exec.Command("schtasks",
		"/create",
		"/tn", "SysCleanerWeeklyClean",
		"/tr", action,
		"/sc", "weekly",
		"/d", cfg.DayOfWeek,
		"/st", startTime,
		"/f",
	)
	cmd.SysProcAttr = getSysProcAttr()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to create scheduled task: %w\n%s", err, string(output))
	}

	return nil
}

// RemoveScheduledClean deletes the SysCleanerWeeklyClean scheduled task.
func RemoveScheduledClean() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("scheduled cleaning only available on Windows")
	}

	cmd := exec.Command("schtasks",
		"/delete",
		"/tn", "SysCleanerWeeklyClean",
		"/f",
	)
	cmd.SysProcAttr = getSysProcAttr()

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to remove scheduled task: %w\n%s", err, string(output))
	}

	return nil
}

// GetScheduledClean queries Windows Task Scheduler for the
// SysCleanerWeeklyClean task. It returns the parsed config if the task
// exists, or nil (with no error) if the task is not found.
func GetScheduledClean() (*ScheduleConfig, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("scheduled cleaning only available on Windows")
	}

	cmd := exec.Command("schtasks",
		"/query",
		"/tn", "SysCleanerWeeklyClean",
		"/fo", "LIST",
		"/v",
	)
	cmd.SysProcAttr = getSysProcAttr()

	output, err := cmd.CombinedOutput()
	if err != nil {
		// If the task does not exist schtasks returns an error.
		if strings.Contains(string(output), "ERROR") {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to query scheduled task: %w\n%s", err, string(output))
	}

	cfg := parseTaskOutput(string(output))
	return cfg, nil
}

// parseTaskOutput attempts to extract a ScheduleConfig from the verbose
// LIST output of schtasks /query.
func parseTaskOutput(output string) *ScheduleConfig {
	cfg := &ScheduleConfig{Enabled: true}

	for _, line := range strings.Split(output, "\n") {
		line = strings.TrimSpace(line)

		if strings.HasPrefix(line, "Days:") {
			cfg.DayOfWeek = strings.TrimSpace(strings.TrimPrefix(line, "Days:"))
		}

		if strings.HasPrefix(line, "Start Time:") {
			timeStr := strings.TrimSpace(strings.TrimPrefix(line, "Start Time:"))
			// Parse hour from HH:MM:SS or HH:MM format.
			var hour int
			fmt.Sscanf(timeStr, "%d:", &hour)
			cfg.Hour = hour
		}

		if strings.HasPrefix(line, "Task To Run:") {
			taskCmd := strings.TrimSpace(strings.TrimPrefix(line, "Task To Run:"))
			cfg.CleanPreset = parseCleanPreset(taskCmd)
		}
	}

	return cfg
}

// parseCleanPreset extracts the clean preset from the task command string.
// It looks for the last "--" prefixed flag which represents the preset
// (e.g., "--all", "--system", "--browsers").
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
