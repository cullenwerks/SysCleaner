# Native Win32 API — No CMD Windows Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Replace every remaining `os/exec` child-process call with native Win32 API calls so no CMD/PowerShell windows ever flash on screen and Windows Defender heuristics are never triggered.

**Architecture:** All new native implementations live in `_windows.go` files and are called from platform-agnostic `.go` files via thin wrapper functions. Non-Windows builds use `_other.go` stubs that already exist. No child processes are spawned anywhere in the app after this change.

**Tech Stack:** `golang.org/x/sys/windows`, `golang.org/x/sys/windows/registry`, `github.com/go-ole/go-ole` + `github.com/yusufpapurcu/wmi` (all already in `go.mod`).

---

## Task 1: Cleaner — DNS cache flush via `DnsFlushResolverCache`

**Files:**
- Modify: `pkg/cleaner/cleaner.go` (function `cleanDNSCache`)
- Modify: `pkg/cleaner/cleaner_windows.go` (add native impl)

**Context:** `cleanDNSCache` currently runs `exec.Command("ipconfig", "/flushdns")`. Replace with a direct call to `dnsapi.dll!DnsFlushResolverCache`.

**Step 1: Add the native flush function to `cleaner_windows.go`**

The file currently only has a build tag and package declaration. Add:

```go
//go:build windows

package cleaner

import (
	"fmt"

	"golang.org/x/sys/windows"
)

var (
	dnsapi                    = windows.NewLazySystemDLL("dnsapi.dll")
	procDnsFlushResolverCache = dnsapi.NewProc("DnsFlushResolverCache")
)

func flushDNSCacheNative() error {
	ret, _, err := procDnsFlushResolverCache.Call()
	if ret == 0 {
		return fmt.Errorf("DnsFlushResolverCache failed: %w", err)
	}
	return nil
}
```

**Step 2: Update `cleanDNSCache` in `cleaner.go`**

Find `cleanDNSCache` (around line 625). Replace the body:

```go
func cleanDNSCache(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" || opts.DryRun {
		return result
	}

	if err := flushDNSCacheNative(); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to flush DNS cache: %w", err))
	} else {
		log.Println("[SysCleaner] DNS cache flushed")
	}
	return result
}
```

**Step 3: Remove `os/exec` from `cleaner.go` imports** — do NOT remove yet; there are two more exec calls in this file. Leave for Task 3.

**Step 4: Build check**

```
go build -tags windows ./pkg/cleaner/...
```
Expected: no errors.

**Step 5: Commit**

```bash
git add pkg/cleaner/cleaner.go pkg/cleaner/cleaner_windows.go
git commit -m "feat: flush DNS cache via DnsFlushResolverCache native API"
```

---

## Task 2: Cleaner — Event log clear via `EvtClearLog`

**Files:**
- Modify: `pkg/cleaner/cleaner.go` (function `cleanEventLogs`)
- Modify: `pkg/cleaner/cleaner_windows.go` (add native impl)

**Context:** `cleanEventLogs` currently loops over `["System","Application"]` and runs `wevtutil cl <logname>`. Replace with `wevtapi.dll!EvtClearLog`.

**Step 1: Add `clearEventLogNative` to `cleaner_windows.go`**

Append to the file (after the DNS vars/funcs):

```go
var (
	wevtapi        = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtClearLog = wevtapi.NewProc("EvtClearLog")
)

// clearEventLogNative clears a Windows event log channel by name.
// channelPath is e.g. "System" or "Application".
func clearEventLogNative(channelPath string) error {
	channel, err := windows.UTF16PtrFromString(channelPath)
	if err != nil {
		return fmt.Errorf("invalid channel name %q: %w", channelPath, err)
	}
	// EvtClearLog(Session, ChannelPath, TargetFilePath, Flags)
	// Session=0 means local, TargetFilePath=nil means discard cleared events.
	ret, _, err := procEvtClearLog.Call(
		0,
		uintptr(unsafe.Pointer(channel)),
		0,
		0,
	)
	if ret == 0 {
		return fmt.Errorf("EvtClearLog(%s) failed: %w", channelPath, err)
	}
	return nil
}
```

Add `"unsafe"` to the imports in `cleaner_windows.go`.

**Step 2: Update `cleanEventLogs` in `cleaner.go`**

Replace the body (around line 662):

```go
func cleanEventLogs(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" || opts.DryRun {
		return result
	}

	// Security event log is intentionally excluded — clearing it is an
	// anti-forensics indicator that triggers AV heuristics.
	for _, logName := range []string{"System", "Application"} {
		if err := clearEventLogNative(logName); err != nil {
			result.Errors = append(result.Errors, fmt.Errorf("failed to clear %s event log: %w", logName, err))
		} else {
			log.Printf("[SysCleaner] Cleared %s event log", logName)
		}
	}
	return result
}
```

**Step 3: Build check**

```
go build -tags windows ./pkg/cleaner/...
```
Expected: no errors.

**Step 4: Commit**

```bash
git add pkg/cleaner/cleaner.go pkg/cleaner/cleaner_windows.go
git commit -m "feat: clear event logs via EvtClearLog native API"
```

---

## Task 3: Cleaner — Recycle Bin via `SHEmptyRecycleBin`

**Files:**
- Modify: `pkg/cleaner/cleaner.go` (function `cleanRecycleBin`)
- Modify: `pkg/cleaner/cleaner_windows.go` (add native impl)

**Context:** `cleanRecycleBin` runs `powershell -Command Clear-RecycleBin ...`. Replace with `shell32.dll!SHEmptyRecycleBinW`.

**Step 1: Add `emptyRecycleBinNative` to `cleaner_windows.go`**

Append:

```go
var (
	shell32              = windows.NewLazySystemDLL("shell32.dll")
	procSHEmptyRecycleBin = shell32.NewProc("SHEmptyRecycleBinW")
)

const (
	sherbNoConfirmation = 0x00000001
	sherbNoProgressUI   = 0x00000002
	sherbNoSound        = 0x00000004
)

func emptyRecycleBinNative() error {
	// SHEmptyRecycleBinW(hwnd, pszRootPath, dwFlags)
	// hwnd=0, pszRootPath=nil clears all drives.
	ret, _, err := procSHEmptyRecycleBin.Call(
		0,
		0,
		uintptr(sherbNoConfirmation|sherbNoProgressUI|sherbNoSound),
	)
	// S_OK = 0, S_FALSE (0x1) = already empty — both are success
	if ret != 0 && ret != 0x1 {
		return fmt.Errorf("SHEmptyRecycleBin failed (HRESULT 0x%x): %w", ret, err)
	}
	return nil
}
```

**Step 2: Update `cleanRecycleBin` in `cleaner.go`**

Replace the body (around line 693):

```go
func cleanRecycleBin(opts CleanOptions) CleanResult {
	result := CleanResult{}
	if runtime.GOOS != "windows" || opts.DryRun {
		return result
	}

	if err := emptyRecycleBinNative(); err != nil {
		result.Errors = append(result.Errors, fmt.Errorf("failed to clear recycle bin: %w", err))
	} else {
		log.Println("[SysCleaner] Recycle Bin cleared")
	}
	return result
}
```

**Step 3: Remove `os/exec` import from `cleaner.go`**

Remove `"os/exec"` from the import block at the top of `cleaner.go`. All three exec calls are now gone.

**Step 4: Build check**

```
go build -tags windows ./pkg/cleaner/...
```
Expected: no errors, no unused import warnings.

**Step 5: Commit**

```bash
git add pkg/cleaner/cleaner.go pkg/cleaner/cleaner_windows.go
git commit -m "feat: empty recycle bin via SHEmptyRecycleBin native API, remove os/exec from cleaner"
```

---

## Task 4: Gaming — Power scheme via `PowerSetActiveScheme` + TCP via registry

**Files:**
- Create: `pkg/gaming/gaming_windows.go`
- Modify: `pkg/gaming/gaming.go`
- Modify: `pkg/gaming/extreme.go`

**Context:** `gaming.go` has a `runCmd` helper used for `powercfg /setactive` (2× in `Enable`/`Disable`) and `netsh int tcp set global` (3× in `Enable`). `extreme.go` calls `runCmd("powercfg", ...)` once more. All replaced with native APIs.

**Step 1: Create `pkg/gaming/gaming_windows.go`**

```go
//go:build windows

package gaming

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	powrprof              = windows.NewLazySystemDLL("powrprof.dll")
	procPowerSetActiveScheme = powrprof.NewProc("PowerSetActiveScheme")
)

// setPowerSchemeNative activates a Windows power scheme by GUID string.
// guidStr is in the form "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx".
func setPowerSchemeNative(guidStr string) error {
	guid, err := windows.GUIDFromString("{" + guidStr + "}")
	if err != nil {
		return fmt.Errorf("invalid power scheme GUID %q: %w", guidStr, err)
	}
	ret, _, err := procPowerSetActiveScheme.Call(
		0, // RootPowerKey — NULL means current user
		uintptr(unsafe.Pointer(&guid)),
	)
	if ret != 0 {
		return fmt.Errorf("PowerSetActiveScheme failed (0x%x): %w", ret, err)
	}
	return nil
}

const tcpParamsKey = `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`

// setTCPGamingParams writes the three TCP registry values used by gaming mode.
func setTCPGamingParams() error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, tcpParamsKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open TCP params key: %w", err)
	}
	defer key.Close()

	params := map[string]uint32{
		"TcpAutoTuningLevel": 0, // normal
		"EnableTCPChimney":   1,
		"EnableTCPDCA":       1,
	}
	for name, val := range params {
		if err := key.SetDWordValue(name, val); err != nil {
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
	}
	return nil
}

// startExplorerNative launches explorer.exe detached from our process
// so it has no inherited console and no CMD window appears.
func startExplorerNative() error {
	exePath, err := windows.UTF16PtrFromString(`C:\Windows\explorer.exe`)
	if err != nil {
		return err
	}

	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))

	const detachedProcess = 0x00000008
	err = syscall.CreateProcess(
		exePath,
		nil,
		nil,
		nil,
		false,
		detachedProcess,
		nil,
		nil,
		&si,
		&pi,
	)
	if err != nil {
		return fmt.Errorf("failed to start explorer.exe: %w", err)
	}
	syscall.CloseHandle(pi.Thread)
	syscall.CloseHandle(pi.Process)
	return nil
}
```

**Step 2: Update `gaming.go` — replace `runCmd` calls with native helpers**

In `Enable` (around line 97), replace:
```go
// Old:
runCmd("powercfg", "/setactive", "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c")
// ...
runCmd("netsh", "int", "tcp", "set", "global", "autotuninglevel=normal")
runCmd("netsh", "int", "tcp", "set", "global", "chimney=enabled")
runCmd("netsh", "int", "tcp", "set", "global", "dca=enabled")
```
With:
```go
if err := setPowerSchemeNative("8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"); err != nil {
    log.Printf("[SysCleaner] Failed to set high performance power plan: %v", err)
}
if err := setTCPGamingParams(); err != nil {
    log.Printf("[SysCleaner] Failed to set TCP gaming params: %v", err)
}
```

In `Disable` (around line 143), replace:
```go
// Old:
runCmd("powercfg", "/setactive", "381b4222-f694-41f0-9685-ff5bb260df2e")
```
With:
```go
if err := setPowerSchemeNative("381b4222-f694-41f0-9685-ff5bb260df2e"); err != nil {
    log.Printf("[SysCleaner] Failed to restore balanced power plan: %v", err)
}
```

**Step 3: Delete `runCmd` from `gaming.go`**

Remove the `runCmd` function (around line 293) and its `os/exec` import. Also remove the `getSysProcAttr()` call that referenced it.

**Step 4: Update `extreme.go` — replace powercfg runCmd + explorer functions**

In `EnableExtremeMode`, replace:
```go
// Old:
runCmd("powercfg", "/setactive", "8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c")
```
With:
```go
if err := setPowerSchemeNative("8c5e7fda-e8bf-4a96-9a85-a6e23a8c635c"); err != nil {
    log.Printf("[SysCleaner] Failed to set ultimate performance power plan: %v", err)
}
```

Replace `stopWindowsExplorer`:
```go
func stopWindowsExplorer() error {
	return terminateProcessByName("explorer.exe")
}
```

Replace `startWindowsExplorer`:
```go
func startWindowsExplorer() error {
	return startExplorerNative()
}
```

Remove `os/exec` import from `extreme.go`.

**Step 5: Delete procattr files** — they only contained `getSysProcAttr` which is no longer used anywhere.

Delete `pkg/gaming/procattr_windows.go` and `pkg/gaming/procattr_other.go`.

**Step 6: Build check**

```
go build -tags windows ./pkg/gaming/...
```
Expected: no errors.

**Step 7: Commit**

```bash
git add pkg/gaming/
git commit -m "feat: replace powercfg/netsh/taskkill/explorer exec calls with native Win32 APIs in gaming package"
```

---

## Task 5: Optimizer — Network optimization via TCP registry writes

**Files:**
- Modify: `pkg/optimizer/optimizer.go` (function `OptimizeNetwork`)
- Modify: `pkg/optimizer/optimizer_windows.go` (add native helpers)
- Modify: `pkg/optimizer/optimizer_other.go` (remove getSysProcAttr stub)

**Context:** `OptimizeNetwork` loops over 6 `netsh` commands. Replace with registry writes. All 6 settings have well-known registry equivalents in `HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`.

**Step 1: Add `setTCPOptimizationParam` to `optimizer_windows.go`**

Replace the entire content of `optimizer_windows.go` with:

```go
//go:build windows

package optimizer

import (
	"fmt"

	"golang.org/x/sys/windows/registry"
)

var unnecessaryStartup = []string{
	"OneDrive", "Skype", "Spotify", "Discord",
	"Steam", "EpicGamesLauncher", "AdobeUpdater",
	"iTunes", "iTunesHelper",
}

// getSysProcAttr is no longer needed — all exec.Command calls have been removed.

func optimizeStartupPlatform() StartupResult {
	// ... (keep existing implementation unchanged)
}

func setNetworkThrottling() error {
	// ... (keep existing implementation unchanged)
}

const tcpParamsKey = `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`

// tcpOptimizations maps human-readable description to registry key+value.
var tcpOptimizations = []struct {
	desc string
	name string
	val  uint32
}{
	{"Set TCP auto-tuning to normal", "TcpAutoTuningLevel", 0},
	{"Enable TCP chimney offload", "EnableTCPChimney", 1},
	{"Enable direct cache access", "EnableTCPDCA", 1},
	{"Enable NetDMA", "EnableTCPDMA", 1},
	{"Enable receive-side scaling", "EnableRSS", 1},
	{"Disable TCP heuristics", "TcpHeuristics", 0},
}

func setTCPOptimizationParam(name string, val uint32) error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, tcpParamsKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open TCP params key: %w", err)
	}
	defer key.Close()
	return key.SetDWordValue(name, val)
}

// isSSDPresentNative detects whether any SSD is present using WMI.
func isSSDPresentNative() bool {
	// Uses the wmi package which is already an indirect dep via gopsutil.
	// Import: github.com/yusufpapurcu/wmi
	type Win32DiskDrive struct {
		MediaType string
	}
	var drives []Win32DiskDrive
	if err := wmiQuery("SELECT MediaType FROM Win32_DiskDrive", &drives); err != nil {
		return false
	}
	for _, d := range drives {
		if d.MediaType == "SSD" || d.MediaType == "Solid State Drive" {
			return true
		}
	}
	return false
}

func enableTRIMNative() error {
	key, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\FileSystem`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("failed to open FileSystem key: %w", err)
	}
	defer key.Close()
	return key.SetDWordValue("DisableDeleteNotify", 0)
}
```

**Important:** The existing `optimizeStartupPlatform` and `setNetworkThrottling` function bodies must be kept intact — only add the new functions.

**Step 2: Add `wmiQuery` helper to `optimizer_windows.go`**

The `wmi` package provides a `Query` function. Add an import and thin wrapper:

```go
import (
	"fmt"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

func wmiQuery(query string, dst interface{}) error {
	return wmi.Query(query, dst)
}
```

**Step 3: Update `OptimizeNetwork` in `optimizer.go`**

Replace the `commands` loop with a call to the registry helpers:

```go
func OptimizeNetwork() NetworkResult {
	result := NetworkResult{}

	if runtime.GOOS != "windows" {
		result.Optimizations = append(result.Optimizations, "Network optimization is only available on Windows")
		return result
	}

	for _, opt := range tcpOptimizations {
		if err := setTCPOptimizationParam(opt.name, opt.val); err == nil {
			result.Optimizations = append(result.Optimizations, opt.desc)
			result.LatencyReduction += 2
		}
	}

	// Disable network throttling via registry (existing function)
	if err := setNetworkThrottling(); err == nil {
		result.Optimizations = append(result.Optimizations, "Disabled network throttling")
		result.LatencyReduction += 2
	}

	return result
}
```

**Step 4: Update `OptimizeDisk` in `optimizer.go`**

Replace the PowerShell SSD detection and fsutil/schtasks calls:

```go
func OptimizeDisk() DiskResult {
	result := DiskResult{}

	if runtime.GOOS != "windows" {
		return result
	}

	result.IsSSD = isSSDPresentNative()

	if result.IsSSD {
		if err := enableTRIMNative(); err == nil {
			result.Scheduled = true
		}
	} else {
		if err := createDefragTask(); err == nil {
			result.Scheduled = true
		}
	}

	return result
}
```

**Step 5: Add `createDefragTask` stub to `optimizer_windows.go`** (will be fully implemented in Task 6):

```go
func createDefragTask() error {
	return createDefragTaskCOM()
}
```

**Step 6: Remove `os/exec` import from `optimizer.go`**

**Step 7: Update `optimizer_other.go`** — remove the `getSysProcAttr` function and its `syscall` import. Add stubs for the new platform functions:

```go
//go:build !windows

package optimizer

func optimizeStartupPlatform() StartupResult { return StartupResult{} }
func setNetworkThrottling() error            { return nil }
func isSSDPresentNative() bool               { return false }
func enableTRIMNative() error                { return nil }
func createDefragTask() error                { return nil }
func setTCPOptimizationParam(name string, val uint32) error { return nil }
```

Also define `tcpOptimizations` as an empty slice in the other build so the loop compiles on non-Windows:

The cleanest approach: move `tcpOptimizations` declaration into `optimizer_windows.go` only, and add a corresponding empty declaration in `optimizer_other.go`:

```go
var tcpOptimizations []struct {
	desc string
	name string
	val  uint32
}
```

**Step 8: Build check**

```
go build -tags windows ./pkg/optimizer/...
```
Expected: no errors.

**Step 9: Commit**

```bash
git add pkg/optimizer/
git commit -m "feat: replace netsh/powershell/fsutil calls with native registry+WMI in optimizer"
```

---

## Task 6: Optimizer — Defrag task via Task Scheduler COM

**Files:**
- Modify: `pkg/optimizer/optimizer_windows.go` (add `createDefragTaskCOM`)

**Context:** The HDD path in `OptimizeDisk` needs to create a weekly scheduled task for `defrag C: /O`. Use the Windows Task Scheduler COM API via `go-ole`.

**Step 1: Add `createDefragTaskCOM` to `optimizer_windows.go`**

```go
import (
	"fmt"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

func createDefragTaskCOM() error {
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		return fmt.Errorf("CoInitialize: %w", err)
	}
	defer ole.CoUninitialize()

	taskServiceUnk, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return fmt.Errorf("create Schedule.Service: %w", err)
	}
	defer taskServiceUnk.Release()

	taskService, err := taskServiceUnk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("QueryInterface Schedule.Service: %w", err)
	}
	defer taskService.Release()

	// Connect to local task service
	if _, err := oleutil.CallMethod(taskService, "Connect"); err != nil {
		return fmt.Errorf("Connect: %w", err)
	}

	// Get root folder
	rootFolderRaw, err := oleutil.CallMethod(taskService, "GetFolder", `\`)
	if err != nil {
		return fmt.Errorf("GetFolder: %w", err)
	}
	rootFolder := rootFolderRaw.ToIDispatch()
	defer rootFolder.Release()

	// Create new task definition
	taskDefRaw, err := oleutil.CallMethod(taskService, "NewTask", 0)
	if err != nil {
		return fmt.Errorf("NewTask: %w", err)
	}
	taskDef := taskDefRaw.ToIDispatch()
	defer taskDef.Release()

	// Set registration info
	regInfoRaw, err := oleutil.GetProperty(taskDef, "RegistrationInfo")
	if err != nil {
		return fmt.Errorf("get RegistrationInfo: %w", err)
	}
	regInfo := regInfoRaw.ToIDispatch()
	defer regInfo.Release()
	oleutil.PutProperty(regInfo, "Description", "SysCleaner weekly HDD defragmentation")
	oleutil.PutProperty(regInfo, "Author", "SysCleaner")

	// Settings: run whether user is logged in or not, not stored password
	settingsRaw, err := oleutil.GetProperty(taskDef, "Settings")
	if err != nil {
		return fmt.Errorf("get Settings: %w", err)
	}
	settings := settingsRaw.ToIDispatch()
	defer settings.Release()
	oleutil.PutProperty(settings, "StartWhenAvailable", true)
	oleutil.PutProperty(settings, "RunOnlyIfNetworkAvailable", false)

	// Weekly trigger: every Sunday at 03:00
	triggersRaw, err := oleutil.GetProperty(taskDef, "Triggers")
	if err != nil {
		return fmt.Errorf("get Triggers: %w", err)
	}
	triggers := triggersRaw.ToIDispatch()
	defer triggers.Release()

	// TASK_TRIGGER_WEEKLY = 4
	triggerRaw, err := oleutil.CallMethod(triggers, "Create", 4)
	if err != nil {
		return fmt.Errorf("create weekly trigger: %w", err)
	}
	trigger := triggerRaw.ToIDispatch()
	defer trigger.Release()

	startTime := time.Now().Format("2006-01-02") + "T03:00:00"
	oleutil.PutProperty(trigger, "StartBoundary", startTime)
	oleutil.PutProperty(trigger, "DaysOfWeek", 1) // Sunday = 1
	oleutil.PutProperty(trigger, "WeeksInterval", 1)
	oleutil.PutProperty(trigger, "Enabled", true)

	// Action: run defrag
	actionsRaw, err := oleutil.GetProperty(taskDef, "Actions")
	if err != nil {
		return fmt.Errorf("get Actions: %w", err)
	}
	actions := actionsRaw.ToIDispatch()
	defer actions.Release()

	// TASK_ACTION_EXEC = 0
	actionRaw, err := oleutil.CallMethod(actions, "Create", 0)
	if err != nil {
		return fmt.Errorf("create action: %w", err)
	}
	action := actionRaw.ToIDispatch()
	defer action.Release()
	oleutil.PutProperty(action, "Path", `C:\Windows\System32\defrag.exe`)
	oleutil.PutProperty(action, "Arguments", "C: /O")

	// Register the task
	// TASK_CREATE_OR_UPDATE = 6, TASK_LOGON_INTERACTIVE_TOKEN = 3
	_, err = oleutil.CallMethod(rootFolder, "RegisterTaskDefinition",
		"SysCleanerDefrag",
		taskDef,
		6,    // TASK_CREATE_OR_UPDATE
		"",   // userId (empty = current user)
		"",   // password
		3,    // TASK_LOGON_INTERACTIVE_TOKEN
		"",   // sddl
	)
	if err != nil {
		return fmt.Errorf("RegisterTaskDefinition: %w", err)
	}

	return nil
}
```

**Step 2: Build check**

```
go build -tags windows ./pkg/optimizer/...
```
Expected: no errors.

**Step 3: Commit**

```bash
git add pkg/optimizer/optimizer_windows.go
git commit -m "feat: create defrag scheduled task via Task Scheduler COM API"
```

---

## Task 7: Scheduler — Replace all schtasks calls with Task Scheduler COM

**Files:**
- Modify: `pkg/scheduler/scheduler.go` (remove exec, delegate to platform funcs)
- Modify: `pkg/scheduler/scheduler_windows.go` (full COM implementation)
- Modify: `pkg/scheduler/scheduler_other.go` (update stub)

**Context:** `scheduler.go` has three public functions (`CreateScheduledClean`, `RemoveScheduledClean`, `GetScheduledClean`) each using `schtasks`. Replace with COM API. The public API and `ScheduleConfig` struct do not change.

**Step 1: Refactor `scheduler.go` to delegate platform calls**

Replace the file with:

```go
package scheduler

import (
	"fmt"
	"os"
	"runtime"
	"strings"
)

// ScheduleConfig holds configuration for a scheduled weekly clean.
type ScheduleConfig struct {
	Enabled     bool
	DayOfWeek   string
	Hour        int
	CleanPreset string
}

// CreateScheduledClean registers a weekly Windows scheduled task.
func CreateScheduledClean(cfg ScheduleConfig) error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("scheduled cleaning only available on Windows")
	}
	exePath, err := os.Executable()
	if err != nil {
		return fmt.Errorf("failed to determine executable path: %w", err)
	}
	return createScheduledCleanNative(cfg, exePath)
}

// RemoveScheduledClean deletes the SysCleanerWeeklyClean scheduled task.
func RemoveScheduledClean() error {
	if runtime.GOOS != "windows" {
		return fmt.Errorf("scheduled cleaning only available on Windows")
	}
	return removeScheduledCleanNative()
}

// GetScheduledClean queries the scheduler for SysCleanerWeeklyClean.
// Returns nil config (no error) if the task does not exist.
func GetScheduledClean() (*ScheduleConfig, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("scheduled cleaning only available on Windows")
	}
	return getScheduledCleanNative()
}

// parseCleanPreset extracts the clean preset flag from a task command string.
func parseCleanPreset(taskCmd string) string {
	parts := strings.Fields(taskCmd)
	for i := len(parts) - 1; i >= 0; i-- {
		p := parts[i]
		if strings.HasPrefix(p, "--") && p != "--headless" && p != "--clean" {
			return strings.TrimPrefix(p, "--")
		}
	}
	return "all"
}
```

**Step 2: Replace `scheduler_windows.go` with full COM implementation**

```go
//go:build windows

package scheduler

import (
	"fmt"
	"time"

	"github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const taskName = "SysCleanerWeeklyClean"

func withTaskService(fn func(svc *ole.IDispatch) error) error {
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		return fmt.Errorf("CoInitialize: %w", err)
	}
	defer ole.CoUninitialize()

	unk, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return fmt.Errorf("create Schedule.Service: %w", err)
	}
	defer unk.Release()

	svc, err := unk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("QueryInterface: %w", err)
	}
	defer svc.Release()

	if _, err := oleutil.CallMethod(svc, "Connect"); err != nil {
		return fmt.Errorf("Connect: %w", err)
	}

	return fn(svc)
}

func createScheduledCleanNative(cfg ScheduleConfig, exePath string) error {
	return withTaskService(func(svc *ole.IDispatch) error {
		rootFolderRaw, err := oleutil.CallMethod(svc, "GetFolder", `\`)
		if err != nil {
			return fmt.Errorf("GetFolder: %w", err)
		}
		rootFolder := rootFolderRaw.ToIDispatch()
		defer rootFolder.Release()

		taskDefRaw, err := oleutil.CallMethod(svc, "NewTask", 0)
		if err != nil {
			return fmt.Errorf("NewTask: %w", err)
		}
		taskDef := taskDefRaw.ToIDispatch()
		defer taskDef.Release()

		// Registration info
		regInfoRaw, _ := oleutil.GetProperty(taskDef, "RegistrationInfo")
		regInfo := regInfoRaw.ToIDispatch()
		defer regInfo.Release()
		oleutil.PutProperty(regInfo, "Description", "SysCleaner weekly scheduled clean")
		oleutil.PutProperty(regInfo, "Author", "SysCleaner")

		// Settings
		settingsRaw, _ := oleutil.GetProperty(taskDef, "Settings")
		settings := settingsRaw.ToIDispatch()
		defer settings.Release()
		oleutil.PutProperty(settings, "StartWhenAvailable", true)

		// Weekly trigger
		triggersRaw, _ := oleutil.GetProperty(taskDef, "Triggers")
		triggers := triggersRaw.ToIDispatch()
		defer triggers.Release()

		// TASK_TRIGGER_WEEKLY = 4
		triggerRaw, err := oleutil.CallMethod(triggers, "Create", 4)
		if err != nil {
			return fmt.Errorf("create trigger: %w", err)
		}
		trigger := triggerRaw.ToIDispatch()
		defer trigger.Release()

		startTime := time.Now().Format("2006-01-02") + fmt.Sprintf("T%02d:00:00", cfg.Hour)
		oleutil.PutProperty(trigger, "StartBoundary", startTime)
		// Map day name to bitmask: Sunday=1, Monday=2, Tuesday=4, ...
		oleutil.PutProperty(trigger, "DaysOfWeek", dayOfWeekBit(cfg.DayOfWeek))
		oleutil.PutProperty(trigger, "WeeksInterval", 1)
		oleutil.PutProperty(trigger, "Enabled", true)

		// Action
		actionsRaw, _ := oleutil.GetProperty(taskDef, "Actions")
		actions := actionsRaw.ToIDispatch()
		defer actions.Release()

		// TASK_ACTION_EXEC = 0
		actionRaw, err := oleutil.CallMethod(actions, "Create", 0)
		if err != nil {
			return fmt.Errorf("create action: %w", err)
		}
		action := actionRaw.ToIDispatch()
		defer action.Release()
		action_args := fmt.Sprintf("--headless --clean --%s", cfg.CleanPreset)
		oleutil.PutProperty(action, "Path", exePath)
		oleutil.PutProperty(action, "Arguments", action_args)

		// TASK_CREATE_OR_UPDATE = 6, TASK_LOGON_INTERACTIVE_TOKEN = 3
		_, err = oleutil.CallMethod(rootFolder, "RegisterTaskDefinition",
			taskName, taskDef, 6, "", "", 3, "")
		return err
	})
}

func removeScheduledCleanNative() error {
	return withTaskService(func(svc *ole.IDispatch) error {
		rootFolderRaw, err := oleutil.CallMethod(svc, "GetFolder", `\`)
		if err != nil {
			return fmt.Errorf("GetFolder: %w", err)
		}
		rootFolder := rootFolderRaw.ToIDispatch()
		defer rootFolder.Release()

		_, err = oleutil.CallMethod(rootFolder, "DeleteTask", taskName, 0)
		return err
	})
}

func getScheduledCleanNative() (*ScheduleConfig, error) {
	var result *ScheduleConfig
	err := withTaskService(func(svc *ole.IDispatch) error {
		rootFolderRaw, err := oleutil.CallMethod(svc, "GetFolder", `\`)
		if err != nil {
			return fmt.Errorf("GetFolder: %w", err)
		}
		rootFolder := rootFolderRaw.ToIDispatch()
		defer rootFolder.Release()

		taskRaw, err := oleutil.CallMethod(rootFolder, "GetTask", taskName)
		if err != nil {
			// Task not found — return nil config, no error
			result = nil
			return nil
		}
		task := taskRaw.ToIDispatch()
		defer task.Release()

		defRaw, err := oleutil.GetProperty(task, "Definition")
		if err != nil {
			return fmt.Errorf("get Definition: %w", err)
		}
		def := defRaw.ToIDispatch()
		defer def.Release()

		cfg := &ScheduleConfig{Enabled: true}

		// Parse first trigger
		triggersRaw, _ := oleutil.GetProperty(def, "Triggers")
		triggers := triggersRaw.ToIDispatch()
		defer triggers.Release()

		countRaw, _ := oleutil.GetProperty(triggers, "Count")
		if countRaw.Val > 0 {
			triggerRaw, _ := oleutil.CallMethod(triggers, "Item", 1)
			trigger := triggerRaw.ToIDispatch()
			defer trigger.Release()

			startRaw, _ := oleutil.GetProperty(trigger, "StartBoundary")
			startStr := startRaw.ToString()
			// Parse "2006-01-02T15:04:05" format
			if len(startStr) >= 13 {
				var h int
				fmt.Sscanf(startStr[11:13], "%d", &h)
				cfg.Hour = h
			}

			dowRaw, _ := oleutil.GetProperty(trigger, "DaysOfWeek")
			cfg.DayOfWeek = dayOfWeekName(int(dowRaw.Val))
		}

		// Parse first action
		actionsRaw, _ := oleutil.GetProperty(def, "Actions")
		actions := actionsRaw.ToIDispatch()
		defer actions.Release()

		actionCountRaw, _ := oleutil.GetProperty(actions, "Count")
		if actionCountRaw.Val > 0 {
			actionRaw, _ := oleutil.CallMethod(actions, "Item", 1)
			action := actionRaw.ToIDispatch()
			defer action.Release()
			argsRaw, _ := oleutil.GetProperty(action, "Arguments")
			cfg.CleanPreset = parseCleanPreset(argsRaw.ToString())
		}

		result = cfg
		return nil
	})
	return result, err
}

// dayOfWeekBit maps a day name to the Windows Task Scheduler bitmask.
// Sunday=1, Monday=2, Tuesday=4, Wednesday=8, Thursday=16, Friday=32, Saturday=64
func dayOfWeekBit(day string) int {
	days := map[string]int{
		"Sunday": 1, "Monday": 2, "Tuesday": 4,
		"Wednesday": 8, "Thursday": 16, "Friday": 32, "Saturday": 64,
	}
	if v, ok := days[day]; ok {
		return v
	}
	return 1 // Default Sunday
}

func dayOfWeekName(bit int) string {
	names := map[int]string{
		1: "Sunday", 2: "Monday", 4: "Tuesday",
		8: "Wednesday", 16: "Thursday", 32: "Friday", 64: "Saturday",
	}
	if n, ok := names[bit]; ok {
		return n
	}
	return "Sunday"
}
```

**Step 3: Update `scheduler_other.go`** — replace `getSysProcAttr` stub with platform function stubs:

```go
//go:build !windows

package scheduler

import "fmt"

func createScheduledCleanNative(cfg ScheduleConfig, exePath string) error {
	return fmt.Errorf("not supported on this platform")
}

func removeScheduledCleanNative() error {
	return fmt.Errorf("not supported on this platform")
}

func getScheduledCleanNative() (*ScheduleConfig, error) {
	return nil, fmt.Errorf("not supported on this platform")
}
```

**Step 4: Build check**

```
go build -tags windows ./pkg/scheduler/...
```
Expected: no errors.

**Step 5: Commit**

```bash
git add pkg/scheduler/
git commit -m "feat: replace schtasks exec calls with Task Scheduler COM API in scheduler package"
```

---

## Task 8: GUI — Game launchers via `ShellExecuteEx`

**Files:**
- Create: `gui/views/launch_windows.go`
- Create: `gui/views/launch_other.go`
- Modify: `gui/views/extreme_mode.go`

**Context:** `createGameLaunchers` in `extreme_mode.go` uses `exec.Command(exe).Start()` to launch game clients. Replace with `ShellExecuteExW` which is the idiomatic Windows API for launching applications — it respects UAC, handles paths with spaces, and does not inherit a console.

**Step 1: Create `gui/views/launch_windows.go`**

```go
//go:build windows && gui

package views

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shell32            = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteEx = shell32.NewProc("ShellExecuteExW")
)

type shellExecuteInfo struct {
	cbSize         uint32
	fMask          uint32
	hwnd           uintptr
	lpVerb         *uint16
	lpFile         *uint16
	lpParameters   *uint16
	lpDirectory    *uint16
	nShow          int32
	hInstApp       uintptr
	lpIDList       uintptr
	lpClass        *uint16
	hkeyClass      uintptr
	dwHotKey       uint32
	hIconOrMonitor uintptr
	hProcess       uintptr
}

const (
	seeMaskNoAsync = 0x00000100
	swShow         = 5
)

// launchExeNative launches an executable using ShellExecuteExW.
// This is the idiomatic Windows way to open applications: no inherited
// console, no CMD window, correct UAC handling.
func launchExeNative(exePath string) error {
	verbPtr, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	filePtr, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return err
	}

	info := shellExecuteInfo{
		fMask:  seeMaskNoAsync,
		lpVerb: verbPtr,
		lpFile: filePtr,
		nShow:  swShow,
	}
	info.cbSize = uint32(unsafe.Sizeof(info))

	ret, _, err := procShellExecuteEx.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return fmt.Errorf("ShellExecuteEx failed: %w", err)
	}
	return nil
}
```

**Step 2: Create `gui/views/launch_other.go`**

```go
//go:build !windows && gui

package views

import (
	"fmt"
	"os/exec"
)

func launchExeNative(exePath string) error {
	cmd := exec.Command(exePath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch %s: %w", exePath, err)
	}
	return nil
}
```

**Step 3: Update `extreme_mode.go` — replace exec.Command in `createGameLaunchers`**

Find the launcher button callback (around line 246):

```go
// Old:
cmd := exec.Command(exe)
if err := cmd.Start(); err != nil {
    dialog.ShowError(fmt.Errorf("failed to launch %s: %v", name, err), w)
} else {
    dialog.ShowInformation("Launched", fmt.Sprintf("%s started successfully!", name), w)
}

// New:
if err := launchExeNative(exe); err != nil {
    dialog.ShowError(fmt.Errorf("failed to launch %s: %v", name, err), w)
} else {
    dialog.ShowInformation("Launched", fmt.Sprintf("%s started successfully!", name), w)
}
```

Remove `"os/exec"` from `extreme_mode.go` imports.

**Step 4: Build check**

```
go build -tags "windows gui" ./gui/...
```
Expected: no errors.

**Step 5: Commit**

```bash
git add gui/views/
git commit -m "feat: launch game clients via ShellExecuteExW, remove os/exec from GUI"
```

---

## Task 9: Final cleanup and verification

**Files:**
- Verify: all `os/exec` and `getSysProcAttr` references gone
- Delete: `pkg/gaming/procattr_windows.go`, `pkg/gaming/procattr_other.go`
- Verify: `pkg/optimizer/optimizer_other.go` and `pkg/scheduler/scheduler_other.go` no longer reference `syscall`

**Step 1: Confirm no remaining exec.Command calls**

```bash
grep -r "exec\.Command" --include="*.go" .
```
Expected: **no output** (zero matches).

**Step 2: Confirm no getSysProcAttr references**

```bash
grep -r "getSysProcAttr" --include="*.go" .
```
Expected: **no output**.

**Step 3: Delete procattr files**

```bash
git rm pkg/gaming/procattr_windows.go pkg/gaming/procattr_other.go
```

**Step 4: Full build check (Windows target)**

```bash
GOOS=windows GOARCH=amd64 go build ./...
```
Expected: no errors.

**Step 5: Full build check (GUI + Windows)**

```bash
GOOS=windows GOARCH=amd64 go build -tags gui ./...
```
Expected: no errors.

**Step 6: Run existing tests**

```bash
go test ./...
```
Expected: all pass (tests are platform-agnostic and don't exercise exec.Command paths directly).

**Step 7: Final commit**

```bash
git add -A
git commit -m "chore: delete procattr helper files, confirm zero exec.Command calls remain"
```

---

## Summary of changes

| # | Package | Replaced | With |
|---|---|---|---|
| 1 | `pkg/cleaner` | `ipconfig /flushdns` | `DnsFlushResolverCache` (dnsapi.dll) |
| 2 | `pkg/cleaner` | `wevtutil cl` | `EvtClearLog` (wevtapi.dll) |
| 3 | `pkg/cleaner` | `powershell Clear-RecycleBin` | `SHEmptyRecycleBin` (shell32.dll) |
| 4 | `pkg/gaming` | `powercfg /setactive` | `PowerSetActiveScheme` (powrprof.dll) |
| 4 | `pkg/gaming` | `netsh int tcp set global` ×3 | Registry writes |
| 4 | `pkg/gaming` | `taskkill /F /IM explorer.exe` | `terminateProcessByName` (existing) |
| 4 | `pkg/gaming` | `exec.Command("explorer.exe")` | `CreateProcess` DETACHED |
| 5 | `pkg/optimizer` | `netsh int tcp set global` ×6 | Registry writes |
| 5 | `pkg/optimizer` | `powershell Get-PhysicalDisk` | WMI query via go-ole |
| 5 | `pkg/optimizer` | `fsutil behavior set ...` | Registry write |
| 6 | `pkg/optimizer` | `schtasks /create` (defrag) | Task Scheduler COM |
| 7 | `pkg/scheduler` | `schtasks /create/delete/query` | Task Scheduler COM |
| 8 | `gui/views` | `exec.Command(exe).Start()` | `ShellExecuteExW` |
