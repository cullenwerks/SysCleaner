package gaming

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/process"

	"syscleaner/pkg/admin"
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
	if err := admin.RequireElevation("Gaming Mode"); err != nil {
		return err
	}

	mu.Lock()
	defer mu.Unlock()

	if gamingModeEnabled {
		return fmt.Errorf("gaming mode is already enabled")
	}

	if runtime.GOOS == "windows" {
		// Stop non-essential services
		log.Println("[SysCleaner] Stopping background services for gaming...")
		for _, svc := range servicesToStop {
			log.Printf("[SysCleaner] Stopping service: %s", svc)
			if err := stopService(svc); err == nil {
				stoppedServices = append(stoppedServices, svc)
			}
		}

		// Set high performance power plan
		log.Println("[SysCleaner] Setting high performance power plan...")
		runCmd("powercfg", "/setactive", "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c")

		// Optimize network
		log.Println("[SysCleaner] Optimizing network settings...")
		runCmd("netsh", "int", "tcp", "set", "global", "autotuninglevel=normal")
		runCmd("netsh", "int", "tcp", "set", "global", "chimney=enabled")
		runCmd("netsh", "int", "tcp", "set", "global", "dca=enabled")
	}

	gamingModeEnabled = true
	log.Println("[SysCleaner] Gaming mode enabled.")

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
		log.Println("[SysCleaner] Restoring background services...")
		for _, svc := range stoppedServices {
			log.Printf("[SysCleaner] Starting service: %s", svc)
			startService(svc)
		}
		stoppedServices = nil

		// Restore balanced power plan
		log.Println("[SysCleaner] Restoring balanced power plan...")
		runCmd("powercfg", "/setactive", "381b4222-f694-41f0-9685-ff5bb260df2e")
	}

	// Restore process priorities using native Windows API
	log.Println("[SysCleaner] Restoring process priorities...")
	for pid := range originalPriority {
		if runtime.GOOS == "windows" {
			// NORMAL_PRIORITY_CLASS = 0x20
			if err := setProcessPriorityNative(uint32(pid), 0x20); err != nil {
				log.Printf("[SysCleaner] Failed to restore priority for PID %d: %v", pid, err)
			}
		}
	}
	originalPriority = make(map[int32]int32)

	gamingModeEnabled = false
	log.Println("[SysCleaner] Gaming mode disabled.")
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
		name, _ := p.Name()
		log.Printf("[SysCleaner] Boosting priority for game process: %s (PID: %d)", name, p.Pid)
		// HIGH_PRIORITY_CLASS = 0x80 — use native API instead of wmic to avoid AV heuristics
		if err := setProcessPriorityNative(uint32(p.Pid), 0x80); err != nil {
			log.Printf("[SysCleaner] Failed to boost priority for %s: %v", name, err)
		}
	}
}

func stopService(name string) error {
	log.Printf("[SysCleaner] Requesting service stop: %s", name)
	// Use native SCM API instead of "net stop" to avoid spawning child
	// processes that trigger AV heuristics.
	return stopServiceNative(name)
}

func startService(name string) error {
	log.Printf("[SysCleaner] Requesting service start: %s", name)
	// Use native SCM API instead of "net start" to avoid spawning child
	// processes that trigger AV heuristics.
	return startServiceNative(name)
}

func runCmd(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	if runtime.GOOS == "windows" {
		cmd.SysProcAttr = getSysProcAttr()
	}
	return cmd.Run()
}
