//go:build !windows

package gaming

import "fmt"

func stopServiceNative(name string) error {
	return fmt.Errorf("service control not available on this platform")
}

func startServiceNative(name string) error {
	return fmt.Errorf("service control not available on this platform")
}

func setProcessPriorityNative(pid uint32, priorityClass uint32) error {
	return fmt.Errorf("process priority not available on this platform")
}

func setVisualEffectsNative(enable bool) {
	// Visual effects control is only available on Windows
}
