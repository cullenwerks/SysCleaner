package gaming

import (
	"fmt"
	"os/exec"
	"runtime"
)

// ExtremeMode holds state for extreme performance mode.
type ExtremeMode struct {
	ShellStopped      bool
	AntiCheatServices []string
}

var (
	extremeModeActive bool
	extremeMode       ExtremeMode

	antiCheatServices = []string{
		"vgc",            // Riot Vanguard
		"vgk",            // Riot Vanguard Kernel
		"EasyAntiCheat",  // EasyAntiCheat
		"BEService",      // BattlEye
		"PnkBstrA",       // PunkBuster
		"PnkBstrB",       // PunkBuster
	}

	extremeServicesToStop = []string{
		"wuauserv",           // Windows Update
		"BITS",               // Background Transfer
		"DiagTrack",          // Telemetry
		"SysMain",            // Superfetch
		"WSearch",            // Windows Search
		"UsoSvc",             // Update Orchestrator
		"OneSyncSvc",         // OneDrive sync
		"TabletInputService", // Touch keyboard
		"WbioSrvc",           // Windows Biometric
		"DoSvc",              // Delivery Optimization
		"DusmSvc",            // Data Usage
		"MapsBroker",         // Downloaded Maps
		"RetailDemo",         // Retail Demo
		"Spooler",            // Print Spooler
	}
)

// EnableExtremeMode stops Windows Explorer and non-essential services.
func EnableExtremeMode() error {
	mu.Lock()
	defer mu.Unlock()

	if extremeModeActive {
		return fmt.Errorf("extreme mode already active")
	}

	if runtime.GOOS != "windows" {
		return fmt.Errorf("extreme mode only available on Windows")
	}

	// Enable regular gaming mode first if not already active
	if !gamingModeEnabled {
		mu.Unlock()
		if err := Enable(Config{AutoDetectGames: false, CPUBoost: 100, RAMReserveGB: 1}); err != nil {
			mu.Lock()
			return fmt.Errorf("failed to enable gaming mode: %w", err)
		}
		mu.Lock()
	}

	extremeMode = ExtremeMode{
		AntiCheatServices: antiCheatServices,
	}

	// Stop additional services for extreme mode
	for _, svc := range extremeServicesToStop {
		stopService(svc)
	}

	// Ensure anti-cheat services are running
	for _, svc := range antiCheatServices {
		startService(svc)
	}

	// Stop Windows Explorer (Desktop Experience)
	if err := stopWindowsExplorer(); err != nil {
		return fmt.Errorf("failed to stop explorer: %w", err)
	}
	extremeMode.ShellStopped = true

	// Set ultimate performance power plan
	runCmd("powercfg", "/setactive", "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c")

	// Disable visual effects for maximum performance
	disableVisualEffects()

	extremeModeActive = true
	return nil
}

// DisableExtremeMode restores Windows Explorer and services.
func DisableExtremeMode() error {
	mu.Lock()
	defer mu.Unlock()

	if !extremeModeActive {
		return fmt.Errorf("extreme mode not active")
	}

	// Restart Windows Explorer first
	if extremeMode.ShellStopped {
		if err := startWindowsExplorer(); err != nil {
			return fmt.Errorf("failed to restart explorer: %w", err)
		}
	}

	// Restore services
	for _, svc := range extremeServicesToStop {
		startService(svc)
	}

	// Re-enable visual effects
	enableVisualEffects()

	extremeModeActive = false

	// Disable regular gaming mode
	mu.Unlock()
	err := Disable()
	mu.Lock()
	return err
}

// IsExtremeModeActive returns extreme mode status.
func IsExtremeModeActive() bool {
	mu.Lock()
	defer mu.Unlock()
	return extremeModeActive
}

func stopWindowsExplorer() error {
	return runCmd("taskkill", "/F", "/IM", "explorer.exe")
}

func startWindowsExplorer() error {
	cmd := exec.Command("explorer.exe")
	return cmd.Start()
}

func disableVisualEffects() {
	runCmd("reg", "add", "HKCU\\Control Panel\\Desktop", "/v", "UserPreferencesMask",
		"/t", "REG_BINARY", "/d", "9012038010000000", "/f")
	runCmd("reg", "add", "HKCU\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize",
		"/v", "EnableTransparency", "/t", "REG_DWORD", "/d", "0", "/f")
}

func enableVisualEffects() {
	runCmd("reg", "add", "HKCU\\Control Panel\\Desktop", "/v", "UserPreferencesMask",
		"/t", "REG_BINARY", "/d", "9e3e078012000000", "/f")
	runCmd("reg", "add", "HKCU\\SOFTWARE\\Microsoft\\Windows\\CurrentVersion\\Themes\\Personalize",
		"/v", "EnableTransparency", "/t", "REG_DWORD", "/d", "1", "/f")
}
