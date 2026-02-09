//go:build !windows

package priority

import "fmt"

// SetProcessPriority is not available on non-Windows platforms
func SetProcessPriority(processName string, cpuPriority, ioPriority, pagePriority int) error {
	return fmt.Errorf("CPU priority management is only available on Windows")
}

// RemoveProcessPriority is not available on non-Windows platforms
func RemoveProcessPriority(processName string) error {
	return fmt.Errorf("CPU priority management is only available on Windows")
}

// ListConfiguredPriorities is not available on non-Windows platforms
func ListConfiguredPriorities() ([]PriorityEntry, error) {
	return nil, fmt.Errorf("CPU priority management is only available on Windows")
}
