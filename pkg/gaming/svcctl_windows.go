//go:build windows

package gaming

import (
	"fmt"
	"log"
	"time"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
	"golang.org/x/sys/windows/svc"
	"golang.org/x/sys/windows/svc/mgr"
)

// stopServiceNative stops a Windows service using the Service Control Manager API.
// This avoids spawning "net stop" child processes which trigger AV heuristics.
func stopServiceNative(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("failed to open service %s: %w", name, err)
	}
	defer s.Close()

	status, err := s.Control(svc.Stop)
	if err != nil {
		return fmt.Errorf("failed to stop service %s: %w", name, err)
	}

	// Wait for the service to actually stop (up to 10 seconds)
	deadline := time.Now().Add(10 * time.Second)
	for status.State != svc.Stopped && time.Now().Before(deadline) {
		time.Sleep(300 * time.Millisecond)
		status, err = s.Query()
		if err != nil {
			return fmt.Errorf("failed to query service %s: %w", name, err)
		}
	}

	if status.State != svc.Stopped {
		return fmt.Errorf("service %s did not stop within timeout", name)
	}
	return nil
}

// startServiceNative starts a Windows service using the Service Control Manager API.
func startServiceNative(name string) error {
	m, err := mgr.Connect()
	if err != nil {
		return fmt.Errorf("failed to connect to SCM: %w", err)
	}
	defer m.Disconnect()

	s, err := m.OpenService(name)
	if err != nil {
		return fmt.Errorf("failed to open service %s: %w", name, err)
	}
	defer s.Close()

	err = s.Start()
	if err != nil {
		return fmt.Errorf("failed to start service %s: %w", name, err)
	}
	return nil
}

// setProcessPriorityNative sets a process priority class using the Windows API.
// This replaces "wmic process where processid=X CALL setpriority Y" which
// triggers AV heuristics because WMIC-based process manipulation is a common
// malware pattern.
//
// Priority classes:
//
//	windows.IDLE_PRIORITY_CLASS         = 0x00000040
//	windows.BELOW_NORMAL_PRIORITY_CLASS = 0x00004000
//	windows.NORMAL_PRIORITY_CLASS       = 0x00000020
//	windows.ABOVE_NORMAL_PRIORITY_CLASS = 0x00008000
//	windows.HIGH_PRIORITY_CLASS         = 0x00000080
//	windows.REALTIME_PRIORITY_CLASS     = 0x00000100
func setProcessPriorityNative(pid uint32, priorityClass uint32) error {
	handle, err := windows.OpenProcess(
		windows.PROCESS_SET_INFORMATION|windows.PROCESS_QUERY_INFORMATION,
		false, pid)
	if err != nil {
		return fmt.Errorf("failed to open process %d: %w", pid, err)
	}
	defer windows.CloseHandle(handle)

	err = windows.SetPriorityClass(handle, priorityClass)
	if err != nil {
		return fmt.Errorf("failed to set priority for process %d: %w", pid, err)
	}
	return nil
}

// setVisualEffectsNative toggles Windows visual effects using native registry API
// instead of spawning reg.exe child processes.
func setVisualEffectsNative(enable bool) {
	// Set UserPreferencesMask in HKCU\Control Panel\Desktop
	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Control Panel\Desktop`, registry.SET_VALUE)
	if err != nil {
		log.Printf("[SysCleaner] Failed to open Desktop registry key: %v", err)
		return
	}

	var mask []byte
	if enable {
		// Default visual effects enabled
		mask = []byte{0x9e, 0x3e, 0x07, 0x80, 0x12, 0x00, 0x00, 0x00}
	} else {
		// Minimal visual effects for performance
		mask = []byte{0x90, 0x12, 0x03, 0x80, 0x10, 0x00, 0x00, 0x00}
	}
	if err := key.SetBinaryValue("UserPreferencesMask", mask); err != nil {
		log.Printf("[SysCleaner] Failed to set UserPreferencesMask: %v", err)
	}
	key.Close()

	// Set EnableTransparency in Themes\Personalize
	themeKey, _, err := registry.CreateKey(registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Themes\Personalize`,
		registry.SET_VALUE)
	if err != nil {
		log.Printf("[SysCleaner] Failed to open Themes registry key: %v", err)
		return
	}
	defer themeKey.Close()

	var transparencyVal uint32
	if enable {
		transparencyVal = 1
	}
	if err := themeKey.SetDWordValue("EnableTransparency", transparencyVal); err != nil {
		log.Printf("[SysCleaner] Failed to set EnableTransparency: %v", err)
	}
}
