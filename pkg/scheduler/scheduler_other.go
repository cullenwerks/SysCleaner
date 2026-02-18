//go:build !windows

package scheduler

import "fmt"

func createScheduledCleanNative(cfg ScheduleConfig, exePath string) error {
	return fmt.Errorf("scheduled cleaning not supported on this platform")
}

func removeScheduledCleanNative() error {
	return fmt.Errorf("scheduled cleaning not supported on this platform")
}

func getScheduledCleanNative() (*ScheduleConfig, error) {
	return nil, fmt.Errorf("scheduled cleaning not supported on this platform")
}
