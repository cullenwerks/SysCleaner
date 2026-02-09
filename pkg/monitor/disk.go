package monitor

import (
	"math"
	"os"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/disk"
)

// DiskStatus holds information about the current state of a disk drive.
type DiskStatus struct {
	DriveLetter string
	TotalGB     float64
	FreeGB      float64
	FreePercent float64
	Warning     bool
}

var (
	monitorDone   chan struct{}
	monitorActive bool
	monitorMu     sync.Mutex
)

// CheckDiskSpace checks free space on the system drive and returns a DiskStatus.
// It reads the SystemDrive environment variable, defaulting to "C:" if unset.
func CheckDiskSpace() DiskStatus {
	drive := os.Getenv("SystemDrive")
	if drive == "" {
		drive = "C:"
	}

	path := drive + "\\"

	usage, err := disk.Usage(path)
	if err != nil {
		return DiskStatus{
			DriveLetter: drive,
			Warning:     true,
		}
	}

	totalGB := float64(usage.Total) / (1024 * 1024 * 1024)
	freeGB := float64(usage.Free) / (1024 * 1024 * 1024)
	freePercent := 0.0
	if usage.Total > 0 {
		freePercent = (float64(usage.Free) / float64(usage.Total)) * 100.0
	}

	totalGB = math.Round(totalGB*100) / 100
	freeGB = math.Round(freeGB*100) / 100
	freePercent = math.Round(freePercent*100) / 100

	return DiskStatus{
		DriveLetter: drive,
		TotalGB:     totalGB,
		FreeGB:      freeGB,
		FreePercent: freePercent,
		Warning:     freePercent < 10.0,
	}
}

// StartDiskMonitor starts a background goroutine that checks disk space at the
// given interval and invokes the callback with the result each time. It is safe
// to call from multiple goroutines; duplicate starts are ignored.
func StartDiskMonitor(interval time.Duration, callback func(DiskStatus)) {
	monitorMu.Lock()
	defer monitorMu.Unlock()

	if monitorActive {
		return
	}

	monitorDone = make(chan struct{})
	monitorActive = true

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-monitorDone:
				return
			case <-ticker.C:
				status := CheckDiskSpace()
				callback(status)
			}
		}
	}()
}

// StopDiskMonitor stops the background disk monitoring goroutine. It is safe to
// call even if no monitor is running.
func StopDiskMonitor() {
	monitorMu.Lock()
	defer monitorMu.Unlock()

	if !monitorActive {
		return
	}

	close(monitorDone)
	monitorActive = false
}
