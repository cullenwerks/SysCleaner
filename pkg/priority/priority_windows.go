//go:build windows

package priority

import (
	"fmt"
	"strings"

	"syscleaner/pkg/admin"

	"golang.org/x/sys/windows/registry"
)

const baseKeyPath = `SOFTWARE\Microsoft\Windows NT\CurrentVersion\Image File Execution Options`

// SetProcessPriority creates/updates the registry key for permanent priority
func SetProcessPriority(processName string, cpuPriority, ioPriority, pagePriority int) error {
	if err := admin.RequireElevation("CPU Priority Management"); err != nil {
		return err
	}

	// Ensure .exe extension
	if !strings.HasSuffix(strings.ToLower(processName), ".exe") {
		processName += ".exe"
	}

	// Validate
	validCpu := map[int]bool{1: true, 2: true, 3: true, 5: true, 6: true}
	if !validCpu[cpuPriority] {
		return fmt.Errorf("invalid CPU priority: %d (valid: 1,2,3,5,6 - no Realtime)", cpuPriority)
	}
	if ioPriority < 0 || ioPriority > 3 {
		return fmt.Errorf("invalid I/O priority: %d (valid: 0-3)", ioPriority)
	}
	if pagePriority < 0 || pagePriority > 5 {
		return fmt.Errorf("invalid page priority: %d (valid: 0-5)", pagePriority)
	}

	// Create process key
	processKeyPath := baseKeyPath + `\` + processName
	processKey, _, err := registry.CreateKey(registry.LOCAL_MACHINE, processKeyPath, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create process key: %w", err)
	}
	defer processKey.Close()

	// Create PerfOptions subkey
	perfKey, _, err := registry.CreateKey(registry.LOCAL_MACHINE, processKeyPath+`\PerfOptions`, registry.ALL_ACCESS)
	if err != nil {
		return fmt.Errorf("failed to create PerfOptions key: %w", err)
	}
	defer perfKey.Close()

	// Set values
	if err := perfKey.SetDWordValue("CpuPriorityClass", uint32(cpuPriority)); err != nil {
		return fmt.Errorf("failed to set CpuPriorityClass: %w", err)
	}
	if err := perfKey.SetDWordValue("IoPriority", uint32(ioPriority)); err != nil {
		return fmt.Errorf("failed to set IoPriority: %w", err)
	}
	if err := perfKey.SetDWordValue("PagePriority", uint32(pagePriority)); err != nil {
		return fmt.Errorf("failed to set PagePriority: %w", err)
	}

	return nil
}

// RemoveProcessPriority removes the PerfOptions key for a process
func RemoveProcessPriority(processName string) error {
	if err := admin.RequireElevation("CPU Priority Management"); err != nil {
		return err
	}

	// Ensure .exe extension
	if !strings.HasSuffix(strings.ToLower(processName), ".exe") {
		processName += ".exe"
	}

	processKeyPath := baseKeyPath + `\` + processName + `\PerfOptions`

	// Try to delete the PerfOptions key
	err := registry.DeleteKey(registry.LOCAL_MACHINE, processKeyPath)
	if err != nil {
		// If the key doesn't exist, that's fine
		if err == registry.ErrNotExist {
			return nil
		}
		return fmt.Errorf("failed to delete PerfOptions key: %w", err)
	}

	return nil
}

// ListConfiguredPriorities reads all processes that have PerfOptions set
func ListConfiguredPriorities() ([]PriorityEntry, error) {
	var entries []PriorityEntry

	// Open the Image File Execution Options key
	key, err := registry.OpenKey(registry.LOCAL_MACHINE, baseKeyPath, registry.ENUMERATE_SUB_KEYS)
	if err != nil {
		return nil, fmt.Errorf("failed to open base key: %w", err)
	}
	defer key.Close()

	// Enumerate all subkeys (process names)
	subkeys, err := key.ReadSubKeyNames(-1)
	if err != nil {
		return nil, fmt.Errorf("failed to read subkey names: %w", err)
	}

	for _, processName := range subkeys {
		// Check if this process has PerfOptions
		perfKeyPath := baseKeyPath + `\` + processName + `\PerfOptions`
		perfKey, err := registry.OpenKey(registry.LOCAL_MACHINE, perfKeyPath, registry.QUERY_VALUE)
		if err != nil {
			// No PerfOptions for this process, skip
			continue
		}

		// Read the priority values
		cpuPriority, _, err := perfKey.GetIntegerValue("CpuPriorityClass")
		if err != nil {
			cpuPriority = 2 // default to normal
		}

		ioPriority, _, err := perfKey.GetIntegerValue("IoPriority")
		if err != nil {
			ioPriority = 2 // default to normal
		}

		pagePriority, _, err := perfKey.GetIntegerValue("PagePriority")
		if err != nil {
			pagePriority = 5 // default to normal
		}

		perfKey.Close()

		entry := PriorityEntry{
			ProcessName:      processName,
			CpuPriority:      int(cpuPriority),
			IoPriority:       int(ioPriority),
			PagePriority:     int(pagePriority),
			CpuPriorityName:  GetCpuPriorityName(int(cpuPriority)),
			IoPriorityName:   GetIoPriorityName(int(ioPriority)),
			PagePriorityName: GetPagePriorityName(int(pagePriority)),
		}

		entries = append(entries, entry)
	}

	return entries, nil
}
