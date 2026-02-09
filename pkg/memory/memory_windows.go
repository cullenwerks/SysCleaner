//go:build windows

package memory

import (
	"fmt"
	"log"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/shirou/gopsutil/v3/mem"
	"golang.org/x/sys/windows"
)

// Memory list commands for NtSetSystemInformation
const (
	SystemMemoryListInformation       = 80
	MemoryEmptyWorkingSets            = 0
	MemoryFlushModifiedList           = 1
	MemoryPurgeStandbyList            = 2
	MemoryPurgeLowPriorityStandbyList = 3
	MemoryCombinePageLists            = 4
)

const (
	SE_PROF_SINGLE_PROCESS_PRIVILEGE = "SeProfileSingleProcessPrivilege"
	SE_INCREASE_QUOTA_PRIVILEGE      = "SeIncreaseQuotaPrivilege"
)

var (
	ntdll                      = windows.NewLazySystemDLL("ntdll.dll")
	procNtSetSystemInformation = ntdll.NewProc("NtSetSystemInformation")
	psapi                      = windows.NewLazySystemDLL("psapi.dll")
	procEmptyWorkingSet        = psapi.NewProc("EmptyWorkingSet")

	monitorActive bool
	monitorDone   chan struct{}
	monitorMu     sync.Mutex

	// Configurable thresholds
	FreeMemoryThresholdPercent float64 = 15.0  // Trigger cleanup when free RAM drops below this %
	StandbyThresholdPercent    float64 = 40.0  // Only clear standby if it exceeds this % of total
	MinCleanInterval           = 30 * time.Second // Don't clean more often than this
	lastCleanTime              time.Time
	trimCountTotal             int64
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

// EnableSeProfileSingleProcessPrivilege enables the required privileges
// for memory list operations. Must be called once before any trim operations.
func EnableSeProfileSingleProcessPrivilege() error {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(),
		windows.TOKEN_ADJUST_PRIVILEGES|windows.TOKEN_QUERY, &token)
	if err != nil {
		return fmt.Errorf("failed to open process token: %w", err)
	}
	defer token.Close()

	// Enable SE_PROF_SINGLE_PROCESS_PRIVILEGE
	if err := enablePrivilege(token, SE_PROF_SINGLE_PROCESS_PRIVILEGE); err != nil {
		return fmt.Errorf("failed to enable SeProfileSingleProcessPrivilege: %w", err)
	}

	// Enable SE_INCREASE_QUOTA_PRIVILEGE
	if err := enablePrivilege(token, SE_INCREASE_QUOTA_PRIVILEGE); err != nil {
		return fmt.Errorf("failed to enable SeIncreaseQuotaPrivilege: %w", err)
	}

	return nil
}

func enablePrivilege(token windows.Token, privilegeName string) error {
	var luid windows.LUID
	privName, err := syscall.UTF16PtrFromString(privilegeName)
	if err != nil {
		return err
	}

	err = windows.LookupPrivilegeValue(nil, privName, &luid)
	if err != nil {
		return err
	}

	tp := windows.Tokenprivileges{
		PrivilegeCount: 1,
		Privileges: [1]windows.LUIDAndAttributes{
			{
				Luid:       luid,
				Attributes: windows.SE_PRIVILEGE_ENABLED,
			},
		},
	}

	err = windows.AdjustTokenPrivileges(token, false, &tp, 0, nil, nil)
	if err != nil {
		return err
	}

	return nil
}

// PurgeStandbyList clears the standby memory list (the big one for gaming).
// This is the equivalent of RAMMap's "Empty Standby List".
// Requires SeProfileSingleProcessPrivilege.
func PurgeStandbyList() error {
	cmd := int32(MemoryPurgeStandbyList)
	ret, _, err := procNtSetSystemInformation.Call(
		uintptr(SystemMemoryListInformation),
		uintptr(unsafe.Pointer(&cmd)),
		uintptr(unsafe.Sizeof(cmd)),
	)
	if ret != 0 {
		return fmt.Errorf("NtSetSystemInformation failed: %v (NTSTATUS: 0x%x)", err, ret)
	}
	return nil
}

// PurgeLowPriorityStandby clears only low-priority standby pages.
// This is gentler than PurgeStandbyList and less likely to cause stutter.
func PurgeLowPriorityStandby() error {
	cmd := int32(MemoryPurgeLowPriorityStandbyList)
	ret, _, err := procNtSetSystemInformation.Call(
		uintptr(SystemMemoryListInformation),
		uintptr(unsafe.Pointer(&cmd)),
		uintptr(unsafe.Sizeof(cmd)),
	)
	if ret != 0 {
		return fmt.Errorf("NtSetSystemInformation failed: %v (NTSTATUS: 0x%x)", err, ret)
	}
	return nil
}

// EmptyProcessWorkingSet trims the working set of a specific process.
// This is gentler than purging the standby list.
func EmptyProcessWorkingSet(pid uint32) error {
	handle, err := windows.OpenProcess(
		windows.PROCESS_QUERY_INFORMATION|windows.PROCESS_SET_QUOTA,
		false, pid)
	if err != nil {
		return err
	}
	defer windows.CloseHandle(handle)

	ret, _, err := procEmptyWorkingSet.Call(uintptr(handle))
	if ret == 0 {
		return fmt.Errorf("EmptyWorkingSet failed: %v", err)
	}
	return nil
}

// StartContinuousMonitor begins monitoring RAM and trimming standby memory
// when free RAM drops below the threshold. This is ONLY active during
// Extreme Gaming Mode.
//
// Strategy (to avoid performance drops):
//  1. Check free memory every 5 seconds
//  2. If free memory < 15% of total AND standby > 40% of total:
//     a. First attempt: PurgeLowPriorityStandby (gentle)
//     b. If still low after 10s: PurgeStandbyList (aggressive)
//  3. Never trim more often than every 30 seconds
//  4. Log every trim action with before/after stats
func StartContinuousMonitor(statsCallback func(MemoryStats)) {
	monitorMu.Lock()
	if monitorActive {
		monitorMu.Unlock()
		return
	}
	monitorDone = make(chan struct{})
	monitorActive = true
	monitorMu.Unlock()

	// Enable required privileges
	if err := EnableSeProfileSingleProcessPrivilege(); err != nil {
		log.Printf("[SysCleaner] Warning: Failed to enable memory privileges: %v", err)
		log.Println("[SysCleaner] RAM trimming may not work correctly. Run as Administrator.")
	}

	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-monitorDone:
				return
			case <-ticker.C:
				vmem, err := mem.VirtualMemory()
				if err != nil {
					continue
				}

				totalGB := float64(vmem.Total) / 1024 / 1024 / 1024
				usedGB := float64(vmem.Used) / 1024 / 1024 / 1024
				freeGB := float64(vmem.Available) / 1024 / 1024 / 1024
				// Standby ≈ Available - Free (approximation)
				standbyGB := freeGB - (float64(vmem.Free) / 1024 / 1024 / 1024)
				if standbyGB < 0 {
					standbyGB = 0
				}

				freePercent := (freeGB / totalGB) * 100
				standbyPercent := (standbyGB / totalGB) * 100

				stats := MemoryStats{
					TotalGB:        totalGB,
					UsedGB:         usedGB,
					FreeGB:         freeGB,
					StandbyGB:      standbyGB,
					UsedPercent:    vmem.UsedPercent,
					FreePercent:    freePercent,
					StandbyPercent: standbyPercent,
					LastTrimTime:   lastCleanTime,
					TrimCount:      trimCountTotal,
				}

				if statsCallback != nil {
					statsCallback(stats)
				}

				// Should we trim?
				if freePercent < FreeMemoryThresholdPercent &&
					standbyPercent > StandbyThresholdPercent &&
					time.Since(lastCleanTime) > MinCleanInterval {

					log.Printf("[SysCleaner] RAM Monitor: Free=%.1f%%, Standby=%.1f%% - Trimming...",
						freePercent, standbyPercent)

					// Try gentle first
					if err := PurgeLowPriorityStandby(); err != nil {
						log.Printf("[SysCleaner] Low-priority purge failed: %v", err)
					} else {
						log.Println("[SysCleaner] Low-priority standby trim completed")
					}

					lastCleanTime = time.Now()
					trimCountTotal++

					// Check if that was enough after a brief wait
					time.Sleep(2 * time.Second)
					if vmem2, err := mem.VirtualMemory(); err == nil {
						newFreePercent := (float64(vmem2.Available) / float64(vmem2.Total)) * 100
						if newFreePercent < FreeMemoryThresholdPercent {
							// Aggressive trim
							log.Println("[SysCleaner] Gentle trim insufficient, purging full standby list...")
							if err := PurgeStandbyList(); err != nil {
								log.Printf("[SysCleaner] Full standby purge failed: %v", err)
							} else {
								log.Println("[SysCleaner] Full standby trim completed")
							}
							trimCountTotal++
						}
					}
				}
			}
		}
	}()
}

// StopContinuousMonitor stops the RAM monitor.
func StopContinuousMonitor() {
	monitorMu.Lock()
	defer monitorMu.Unlock()
	if monitorActive && monitorDone != nil {
		close(monitorDone)
		monitorActive = false
	}
}

// TrimNow immediately trims standby memory.
func TrimNow() error {
	if err := EnableSeProfileSingleProcessPrivilege(); err != nil {
		return fmt.Errorf("failed to enable privileges: %w", err)
	}

	log.Println("[SysCleaner] Manual RAM trim requested...")
	if err := PurgeStandbyList(); err != nil {
		return fmt.Errorf("failed to purge standby list: %w", err)
	}

	lastCleanTime = time.Now()
	trimCountTotal++
	log.Println("[SysCleaner] Manual RAM trim completed")
	return nil
}

// GetCurrentStats returns current memory statistics.
func GetCurrentStats() MemoryStats {
	vmem, err := mem.VirtualMemory()
	if err != nil {
		return MemoryStats{}
	}

	totalGB := float64(vmem.Total) / 1024 / 1024 / 1024
	usedGB := float64(vmem.Used) / 1024 / 1024 / 1024
	freeGB := float64(vmem.Available) / 1024 / 1024 / 1024
	standbyGB := freeGB - (float64(vmem.Free) / 1024 / 1024 / 1024)
	if standbyGB < 0 {
		standbyGB = 0
	}

	freePercent := (freeGB / totalGB) * 100
	standbyPercent := (standbyGB / totalGB) * 100

	return MemoryStats{
		TotalGB:        totalGB,
		UsedGB:         usedGB,
		FreeGB:         freeGB,
		StandbyGB:      standbyGB,
		UsedPercent:    vmem.UsedPercent,
		FreePercent:    freePercent,
		StandbyPercent: standbyPercent,
		LastTrimTime:   lastCleanTime,
		TrimCount:      trimCountTotal,
	}
}
