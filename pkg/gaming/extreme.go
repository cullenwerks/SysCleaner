package gaming

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"syscleaner/pkg/admin"
	"syscleaner/pkg/memory"
)

// ExtremeMode holds state for extreme performance mode.
type ExtremeMode struct {
	ShellStopped      bool
	AntiCheatServices []string
	ClosedProcesses   []string
	ramMonitorActive  bool
}

var (
	extremeModeActive bool
	extremeMode       ExtremeMode

	// ProcessWhitelist contains process names that should never be killed.
	// Users can configure this to protect Discord, Steam, etc.
	ProcessWhitelist []string

	// Anti-cheat services that must NEVER be stopped
	antiCheatServices = []string{
		"vgc",           // Riot Vanguard
		"vgk",           // Riot Vanguard Kernel
		"EasyAntiCheat", // EasyAntiCheat
		"BEService",     // BattlEye
		"PnkBstrA",      // PunkBuster
		"PnkBstrB",      // PunkBuster
	}

	// Comprehensive list of non-essential services to stop in Extreme Gaming Mode
	// These have been verified safe to temporarily stop while gaming
	extremeServicesToStop = []string{
		// === Windows Update & Delivery ===
		"wuauserv",     // Windows Update
		"UsoSvc",       // Update Orchestrator
		"BITS",         // Background Intelligent Transfer
		"DoSvc",        // Delivery Optimization
		"WaaSMedicSvc", // Windows Update Medic

		// === Telemetry & Diagnostics ===
		"DiagTrack",        // Connected User Experiences and Telemetry
		"dmwappushservice", // WAP Push Message Routing (telemetry)
		"DPS",              // Diagnostic Policy Service
		"WdiServiceHost",   // Diagnostic Service Host
		"WdiSystemHost",    // Diagnostic System Host
		"PcaSvc",           // Program Compatibility Assistant
		"WerSvc",           // Windows Error Reporting

		// === Search & Indexing ===
		"WSearch", // Windows Search (indexer)

		// === Superfetch & Memory ===
		"SysMain", // Superfetch/SysMain (prefetch/cache manager)

		// === Sync & Cloud ===
		"OneSyncSvc", // Microsoft sync (Mail, Calendar, Contacts)

		// === Print & Fax ===
		"Spooler", // Print Spooler
		"Fax",     // Fax service

		// === Remote Access ===
		"RemoteRegistry", // Remote Registry
		"TermService",    // Remote Desktop Services
		"SessionEnv",     // Remote Desktop Configuration

		// === Biometrics & Security (non-essential) ===
		"WbioSrvc",  // Windows Biometric Service
		"MapsBroker", // Downloaded Maps Manager
		"lfsvc",      // Geolocation Service

		// === Phone & Mobile ===
		"PhoneSvc",  // Phone Service
		"SmsRouter", // Microsoft Windows SMS Router

		// === Touch & Tablet ===
		"TabletInputService", // Touch Keyboard and Handwriting Panel
		"WalletService",      // Wallet Service

		// === Retail & Demos ===
		"RetailDemo", // Retail Demo Service

		// === Miscellaneous ===
		"DusmSvc",              // Data Usage Service
		"wisvc",                // Windows Insider Service
		"icssvc",               // Windows Mobile Hotspot Service
		"WMPNetworkSvc",        // Windows Media Player Network Sharing
		"XblAuthManager",       // Xbox Live Auth Manager (disable if not using Xbox services)
		"XblGameSave",          // Xbox Live Game Save
		"XboxGipSvc",           // Xbox Accessory Management
		"XboxNetApiSvc",        // Xbox Live Networking Service
		"AJRouter",             // AllJoyn Router Service
		"ALG",                  // Application Layer Gateway
		"IKEEXT",               // IKE and AuthIP IPsec Keying Modules (if not using VPN)
		"iphlpsvc",             // IP Helper (IPv6 transition - safe if IPv4 only)
		"SharedAccess",         // Internet Connection Sharing
		"lmhosts",              // TCP/IP NetBIOS Helper
		"TrkWks",               // Distributed Link Tracking Client
		"WpcMonSvc",            // Parental Controls
		"SEMgrSvc",             // Payments and NFC/SE Manager
		"SCardSvr",             // Smart Card
		"ScDeviceEnum",         // Smart Card Device Enumeration
		"stisvc",               // Windows Image Acquisition (scanner/camera)
		"FrameServer",          // Windows Camera Frame Server
		"CDPSvc",               // Connected Devices Platform Service
		"CDPUserSvc",           // Connected Devices Platform User Service
		"WpnService",           // Windows Push Notifications System
		"WpnUserService",       // Windows Push Notifications User Service
		"BcastDVRUserService",  // GameDVR and Broadcast User Service (if not recording)
	}

	// Background applications to close in Extreme Mode
	processesToKill = []string{
		"OneDrive.exe",
		"Teams.exe",
		"ms-teams.exe",
		"Spotify.exe",
		"Discord.exe",
		"DiscordPTB.exe",
		"DiscordCanary.exe",
		"Skype.exe",
		"slack.exe",
		"Cortana.exe",
		"SearchUI.exe",
		"SearchApp.exe",
		"SearchHost.exe",
		"YourPhone.exe",
		"PhoneExperienceHost.exe",
		"CalculatorApp.exe",
		"Microsoft.Photos.exe",
		"Video.UI.exe",
		"HxTsr.exe",
		"HxCalendarAppImm.exe",
		"HxOutlook.exe",
		"GameBar.exe",
		"GameBarPresenceWriter.exe",
		// NOTE: SecurityHealthSystray.exe (Windows Defender tray) is intentionally
		// NOT included — killing security software UI triggers AV heuristics
		// and is a common malware pattern.
		"PeopleApp.exe",
		"msedge.exe",
		"MicrosoftEdgeUpdate.exe",
		"GoogleCrashHandler.exe",
		"GoogleCrashHandler64.exe",
		"jusched.exe",
		"AdobeARM.exe",
		"CCleaner64.exe",
		"iCloudServices.exe",
		"AppleMobileDeviceService.exe",
	}
)

// GetProcessesToKill returns the list of processes that would be terminated.
func GetProcessesToKill() []string {
	return processesToKill
}

// EnableExtremeMode stops Windows Explorer and non-essential services.
func EnableExtremeMode() error {
	if err := admin.RequireElevation("Extreme Performance Mode"); err != nil {
		return err
	}

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

	// Close non-essential background applications using native API
	log.Println("[SysCleaner] Closing background applications for extreme performance...")
	closedCount, closedApps := CloseBackgroundApps(ProcessWhitelist)
	extremeMode.ClosedProcesses = closedApps
	log.Printf("[SysCleaner] Closed %d background applications", closedCount)

	// Stop additional services for extreme mode.
	log.Println("[SysCleaner] Stopping non-essential services for extreme performance...")
	for i, svc := range extremeServicesToStop {
		stopService(svc)
		if i > 0 && i%3 == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Ensure anti-cheat services are running
	log.Println("[SysCleaner] Ensuring anti-cheat services are running...")
	for _, svc := range antiCheatServices {
		startService(svc)
	}

	// Stop Windows Explorer (Desktop Experience)
	// Uses taskkill for explorer.exe — native API is inappropriate for shell
	// processes that auto-restart via Windows Session Manager
	log.Println("[SysCleaner] Stopping Windows Explorer shell...")
	if err := stopWindowsExplorer(); err != nil {
		return fmt.Errorf("failed to stop explorer: %w", err)
	}
	extremeMode.ShellStopped = true

	// Set ultimate performance power plan
	// Uses powercfg — no native API equivalent exists
	runCmd("powercfg", "/setactive", "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c")

	// Disable visual effects for maximum performance using native registry API
	disableVisualEffects()

	// Start RAM monitoring for automatic standby trimming
	log.Println("[SysCleaner] Starting continuous RAM monitoring...")
	if err := memory.EnableSeProfileSingleProcessPrivilege(); err != nil {
		log.Printf("[SysCleaner] Warning: Failed to enable memory privileges: %v", err)
		log.Println("[SysCleaner] RAM trimming may not work correctly. Run as Administrator.")
	}
	memory.StartContinuousMonitor(nil)
	extremeMode.ramMonitorActive = true

	extremeModeActive = true
	log.Println("[SysCleaner] Extreme Mode ACTIVATED - Maximum performance enabled")
	return nil
}

// DisableExtremeMode restores Windows Explorer and services.
func DisableExtremeMode() error {
	mu.Lock()
	defer mu.Unlock()

	if !extremeModeActive {
		return fmt.Errorf("extreme mode not active")
	}

	// Stop RAM monitoring
	if extremeMode.ramMonitorActive {
		log.Println("[SysCleaner] Stopping RAM monitor...")
		memory.StopContinuousMonitor()
		extremeMode.ramMonitorActive = false
	}

	// Restart Windows Explorer first
	if extremeMode.ShellStopped {
		if err := startWindowsExplorer(); err != nil {
			return fmt.Errorf("failed to restart explorer: %w", err)
		}
	}

	// Restore services with pacing
	for i, svc := range extremeServicesToStop {
		startService(svc)
		if i > 0 && i%3 == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	// Re-enable visual effects
	enableVisualEffects()

	extremeModeActive = false

	// Disable regular gaming mode
	mu.Unlock()
	err := Disable()
	mu.Lock()

	log.Println("[SysCleaner] Extreme Mode DEACTIVATED - Normal operation restored")
	return err
}

// IsExtremeModeActive returns extreme mode status.
func IsExtremeModeActive() bool {
	mu.Lock()
	defer mu.Unlock()
	return extremeModeActive
}

// CloseBackgroundApps closes non-essential background applications.
// Returns the count of closed apps and a list of closed process names.
// Uses native TerminateProcess API instead of taskkill.exe to avoid
// triggering AV heuristics from rapid child process spawning.
// Processes in the whitelist are skipped.
func CloseBackgroundApps(whitelist []string) (int, []string) {
	closed := 0
	closedApps := []string{}

	whitelistMap := make(map[string]bool)
	for _, name := range whitelist {
		whitelistMap[strings.ToLower(name)] = true
	}

	for i, processName := range processesToKill {
		if whitelistMap[strings.ToLower(processName)] {
			log.Printf("[SysCleaner] Skipping whitelisted process: %s", processName)
			continue
		}

		if err := terminateProcessByName(processName); err == nil {
			closed++
			closedApps = append(closedApps, processName)
			log.Printf("[SysCleaner] Closed: %s", processName)
		}
		// Batch delay: 500ms every 3 processes to avoid AV heuristics
		if i > 0 && i%3 == 0 {
			time.Sleep(500 * time.Millisecond)
		}
	}

	return closed, closedApps
}

func stopWindowsExplorer() error {
	// Uses taskkill for explorer.exe — explorer is a shell process that
	// requires special handling; native TerminateProcess may not cleanly stop it
	return runCmd("taskkill", "/F", "/IM", "explorer.exe")
}

func startWindowsExplorer() error {
	cmd := exec.Command("explorer.exe")
	return cmd.Start()
}

// disableVisualEffects uses native registry API on Windows, falls back to reg.exe
func disableVisualEffects() {
	setVisualEffectsNative(false)
}

// enableVisualEffects uses native registry API on Windows, falls back to reg.exe
func enableVisualEffects() {
	setVisualEffectsNative(true)
}

// GetExtremeModeStats returns information about what extreme mode has done
func GetExtremeModeStats() (closedApps []string, ramTrimCount int64) {
	mu.Lock()
	defer mu.Unlock()

	if !extremeModeActive {
		return nil, 0
	}

	stats := memory.GetCurrentStats()
	return extremeMode.ClosedProcesses, stats.TrimCount
}
