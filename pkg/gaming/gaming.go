package gaming

import (
	"fmt"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"
)

// Config holds gaming mode configuration.
type Config struct {
	AutoDetectGames bool
	CPUBoost        int
	RAMReserveGB    int
}

// Status holds current gaming mode state.
type Status struct {
	Enabled          bool
	ActiveGames      []GameProcess
	CPUUsage         float64
	RAMUsagePercent  float64
	RAMUsed          uint64
	RAMTotal         uint64
	StoppedServices  []string
}

// GameProcess represents a detected game process.
type GameProcess struct {
	Name     string
	PID      int32
	CPUUsage float64
	RAMUsage uint64
}

var (
	gamingModeEnabled bool
	stoppedServices   []string
	originalPriority  = make(map[int32]int32)
	mu                sync.Mutex
	monitorDone       chan struct{}
)

var gameExecutables = []string{
	"LeagueClient.exe", "League of Legends.exe", "RiotClientServices.exe",
	"valorant.exe", "VALORANT-Win64-Shipping.exe",
	"csgo.exe", "cs2.exe",
	"FortniteClient-Win64-Shipping.exe",
	"ApexLegends.exe", "r5apex.exe",
	"Overwatch.exe",
	"steam.exe",
}

// servicesToStop lists Windows services to stop during gaming mode.
var servicesToStop = []string{
	"wuauserv",  // Windows Update
	"BITS",      // Background Intelligent Transfer
	"DiagTrack", // Diagnostics Tracking
	"SysMain",   // Superfetch
	"WSearch",   // Windows Search
}

// Enable activates gaming mode.
func Enable(config Config) error {
	mu.Lock()
	defer mu.Unlock()

	if gamingModeEnabled {
		return fmt.Errorf("gaming mode is already enabled")
	}

	if runtime.GOOS == "windows" {
		// Stop non-essential services
		for _, svc := range servicesToStop {
			if err := stopService(svc); err == nil {
				stoppedServices = append(stoppedServices, svc)
			}
		}

		// Set high performance power plan
		runCmd("powercfg", "/setactive", "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c")

		// Optimize network
		runCmd("netsh", "int", "tcp", "set", "global", "autotuninglevel=normal")
		runCmd("netsh", "int", "tcp", "set", "global", "chimney=enabled")
		runCmd("netsh", "int", "tcp", "set", "global", "dca=enabled")
	}

	gamingModeEnabled = true

	if config.AutoDetectGames {
		monitorDone = make(chan struct{})
		go monitorGameProcesses(monitorDone)
	}

	return nil
}

// Disable deactivates gaming mode and restores settings.
func Disable() error {
	mu.Lock()
	defer mu.Unlock()

	if !gamingModeEnabled {
		return fmt.Errorf("gaming mode is not enabled")
	}

	// Stop monitor goroutine
	if monitorDone != nil {
		close(monitorDone)
		monitorDone = nil
	}

	if runtime.GOOS == "windows" {
		// Restart stopped services
		for _, svc := range stoppedServices {
			startService(svc)
		}
		stoppedServices = nil

		// Restore balanced power plan
		runCmd("powercfg", "/setactive", "381b4222-f694-41f0-9685-ff5bb260df2e")
	}

	// Restore process priorities
	for pid := range originalPriority {
		if runtime.GOOS == "windows" {
			// Restore normal priority class (32 = normal)
			runCmd("wmic", "process", "where",
				fmt.Sprintf("processid=%d", pid),
				"CALL", "setpriority", "32")
		}
	}
	originalPriority = make(map[int32]int32)

	gamingModeEnabled = false
	return nil
}

// GetStatus returns current gaming mode status.
func GetStatus() Status {
	mu.Lock()
	defer mu.Unlock()

	status := Status{
		Enabled:         gamingModeEnabled,
		StoppedServices: stoppedServices,
	}

	// CPU usage
	if cpuPercent, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(cpuPercent) > 0 {
		status.CPUUsage = cpuPercent[0]
	}

	// RAM usage
	if vmem, err := mem.VirtualMemory(); err == nil {
		status.RAMUsagePercent = vmem.UsedPercent
		status.RAMUsed = vmem.Used
		status.RAMTotal = vmem.Total
	}

	// Detect game processes
	procs, err := process.Processes()
	if err == nil {
		for _, p := range procs {
			name, err := p.Name()
			if err != nil {
				continue
			}
			if isGameProcess(name) {
				gp := GameProcess{
					Name: name,
					PID:  p.Pid,
				}
				if cpuP, err := p.CPUPercent(); err == nil {
					gp.CPUUsage = cpuP
				}
				if memInfo, err := p.MemoryInfo(); err == nil && memInfo != nil {
					gp.RAMUsage = memInfo.RSS
				}
				status.ActiveGames = append(status.ActiveGames, gp)
			}
		}
	}

	return status
}

// IsEnabled returns whether gaming mode is active.
func IsEnabled() bool {
	mu.Lock()
	defer mu.Unlock()
	return gamingModeEnabled
}

func monitorGameProcesses(done chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			procs, err := process.Processes()
			if err != nil {
				continue
			}
			for _, p := range procs {
				name, err := p.Name()
				if err != nil {
					continue
				}
				if isGameProcess(name) {
					boostProcessPriority(p)
				}
			}
		}
	}
}

func isGameProcess(name string) bool {
	nameLower := strings.ToLower(name)
	for _, exe := range gameExecutables {
		if strings.ToLower(exe) == nameLower {
			return true
		}
	}
	return false
}

func boostProcessPriority(p *process.Process) {
	mu.Lock()
	defer mu.Unlock()

	if _, exists := originalPriority[p.Pid]; exists {
		return // already boosted
	}

	nice, err := p.Nice()
	if err != nil {
		nice = 0
	}
	originalPriority[p.Pid] = nice

	if runtime.GOOS == "windows" {
		// On Windows, use wmic to set high priority (128 = High)
		runCmd("wmic", "process", "where",
			fmt.Sprintf("processid=%d", p.Pid),
			"CALL", "setpriority", "128")
	}
}

func stopService(name string) error {
	return runCmd("sc.exe", "stop", name)
}

func startService(name string) error {
	return runCmd("sc.exe", "start", name)
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = getSysProcAttr()
	}
	return cmd.Run()
}
