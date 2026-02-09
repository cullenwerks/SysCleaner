//go:build !windows

package memory

import (
	"fmt"
	"time"
)

// MemoryStats holds current memory status
type MemoryStats struct {
	TotalGB        float64
	UsedGB         float64
	FreeGB         float64
	StandbyGB      float64
	UsedPercent    float64
	FreePercent    float64
	StandbyPercent float64
	LastTrimTime   time.Time
	TrimCount      int64
}

// StartContinuousMonitor is not available on non-Windows platforms
func StartContinuousMonitor(statsCallback func(MemoryStats)) {
	// No-op on non-Windows
}

// StopContinuousMonitor is not available on non-Windows platforms
func StopContinuousMonitor() {
	// No-op on non-Windows
}

// TrimNow is not available on non-Windows platforms
func TrimNow() error {
	return fmt.Errorf("memory trimming is only available on Windows")
}

// GetCurrentStats returns empty stats on non-Windows platforms
func GetCurrentStats() MemoryStats {
	return MemoryStats{}
}

// EnableSeProfileSingleProcessPrivilege is not available on non-Windows
func EnableSeProfileSingleProcessPrivilege() error {
	return fmt.Errorf("privilege management is only available on Windows")
}

// PurgeStandbyList is not available on non-Windows
func PurgeStandbyList() error {
	return fmt.Errorf("memory operations are only available on Windows")
}

// PurgeLowPriorityStandby is not available on non-Windows
func PurgeLowPriorityStandby() error {
	return fmt.Errorf("memory operations are only available on Windows")
}
