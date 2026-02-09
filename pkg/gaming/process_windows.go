//go:build windows

package gaming

import (
	"fmt"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

// terminateProcessByName finds and terminates a process by its executable name
// using native Windows APIs instead of spawning taskkill.exe child processes.
// This avoids triggering AV heuristics from rapid child process spawning.
func terminateProcessByName(name string) error {
	snapshot, err := windows.CreateToolhelp32Snapshot(windows.TH32CS_SNAPPROCESS, 0)
	if err != nil {
		return fmt.Errorf("failed to create process snapshot: %w", err)
	}
	defer windows.CloseHandle(snapshot)

	var entry windows.ProcessEntry32
	entry.Size = uint32(unsafe.Sizeof(entry))

	err = windows.Process32First(snapshot, &entry)
	if err != nil {
		return fmt.Errorf("failed to enumerate processes: %w", err)
	}

	nameLower := strings.ToLower(name)
	terminated := false

	for {
		exeName := windows.UTF16ToString(entry.ExeFile[:])
		if strings.ToLower(exeName) == nameLower {
			handle, err := windows.OpenProcess(windows.PROCESS_TERMINATE, false, entry.ProcessID)
			if err != nil {
				// Process may have exited or we lack permissions; skip
			} else {
				if err := windows.TerminateProcess(handle, 0); err == nil {
					terminated = true
				}
				windows.CloseHandle(handle)
			}
		}

		err = windows.Process32Next(snapshot, &entry)
		if err != nil {
			break
		}
	}

	if !terminated {
		return fmt.Errorf("process %s not found or could not be terminated", name)
	}
	return nil
}
