# Design: Replace exec.Command with Native Win32 APIs

**Date:** 2026-02-18
**Status:** Approved

## Problem

Several features in SysCleaner spawn child processes (`cmd.exe`, `powershell.exe`, system tools) via `os/exec`. This causes two user-visible issues:

1. **Flashing CMD/PowerShell windows** appear briefly on screen during operations.
2. **Windows Defender triggers** — rapid child process spawning is a known heuristic for `Trojan:Win32/Bearfoos.B!ml`. Using `HideWindow: true` makes this worse, not better.

Previous refactoring already replaced service control (`net stop/start`), process termination (`taskkill`), process priority (`wmic`), and visual effects (`reg.exe`) with native Win32 APIs. This design completes the job for all remaining `exec.Command` call sites.

## Goal

Replace every remaining `exec.Command` call with native Windows API calls (via `golang.org/x/sys/windows`, `golang.org/x/sys/windows/registry`, and `go-ole`). Zero child processes spawned during normal operation.

## No New Dependencies

All required packages are already in the module graph:
- `golang.org/x/sys/windows` — Win32 syscalls
- `golang.org/x/sys/windows/registry` — registry read/write
- `github.com/go-ole/go-ole` — COM automation (already indirect via gopsutil)

## Section 1 — Cleaner (`pkg/cleaner/`)

### `cleanDNSCache` → `DnsFlushResolverCache`
- **File:** `cleaner_windows.go`
- **API:** `dnsapi.dll!DnsFlushResolverCache()` — no arguments, returns BOOL
- Load via `windows.NewLazySystemDLL("dnsapi.dll").NewProc("DnsFlushResolverCache")`

### `cleanEventLogs` → `EvtClearLog`
- **File:** `cleaner_windows.go`
- **API:** `wevtapi.dll!EvtClearLog(session, channelPath, targetFilePath, flags)`
- Pass `0` for session (local machine), `nil` for targetFilePath (discard cleared events)
- Call once each for `"System"` and `"Application"`

### `cleanRecycleBin` → `SHEmptyRecycleBin`
- **File:** `cleaner_windows.go`
- **API:** `shell32.dll!SHEmptyRecycleBinW(hwnd, pszRootPath, dwFlags)`
- Pass `0` hwnd, `nil` path (all drives), flags = `SHERB_NOCONFIRMATION (0x1) | SHERB_NOPROGRESSUI (0x2) | SHERB_NOSOUND (0x4)`

## Section 2 — Gaming (`pkg/gaming/`)

### `runCmd("powercfg", "/setactive", GUID)` → `PowerSetActiveScheme`
- **File:** `gaming_windows.go` (new, follows existing pattern)
- **API:** `powrprof.dll!PowerSetActiveScheme(RootPowerKey, SchemeGuid)`
- Pass `0` for RootPowerKey; parse GUID string into `windows.GUID`
- New helper: `setPowerSchemeNative(guidStr string) error`
- Remove `runCmd` calls for powercfg in `gaming.go` and `extreme.go`

### `runCmd("netsh", "int", "tcp", "set", "global", ...)` → Registry writes
- **File:** `gaming_windows.go`
- **Registry path:** `HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`
- Mappings:
  - `autotuninglevel=normal` → `TcpAutoTuningLevel = 0` (DWORD)
  - `chimney=enabled` → `EnableTCPChimney = 1` (DWORD)
  - `dca=enabled` → `EnableTCPDCA = 1` (DWORD)
- New helper: `setTCPGamingParams() error`

### `stopWindowsExplorer` → `terminateProcessByName`
- **File:** `extreme.go`
- Already implemented in `process_windows.go` — just call `terminateProcessByName("explorer.exe")`
- Remove `taskkill` exec.Command usage

### `startWindowsExplorer` → `CreateProcess`
- **File:** `extreme.go` + `gaming_windows.go`
- **API:** `windows.CreateProcess` with `DETACHED_PROCESS` flag
- Starts explorer.exe detached from our process, no inherited console

## Section 3 — Optimizer (`pkg/optimizer/`)

### `OptimizeNetwork` — 6× `netsh` calls → Registry writes
- **File:** `optimizer_windows.go`
- **Registry path:** `HKLM\SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`
- Mappings:
  - `autotuninglevel=normal` → `TcpAutoTuningLevel = 0`
  - `chimney=enabled` → `EnableTCPChimney = 1`
  - `dca=enabled` → `EnableTCPDCA = 1`
  - `netdma=enabled` → `EnableTCPDMA = 1`
  - `rss=enabled` → `*RSS = 1` (per-adapter; also global `EnableRSS = 1`)
  - `heuristics disabled` → `DisableTaskOffload = 0` + `TcpHeuristics = 0`
- New helper: `setTCPOptimizationParam(name string, val uint32) error`
- Replace loop over `commands` slice; build equivalent results list from registry outcomes

### `OptimizeDisk` — `powershell Get-PhysicalDisk` → WMI via go-ole
- **File:** `optimizer_windows.go`
- **API:** WMI query `SELECT * FROM Win32_DiskDrive WHERE MediaType='SSD'`
- Use `go-ole` + `github.com/yusufpapurcu/wmi` (already an indirect dep)
- New helper: `isSSDPresentNative() bool`

### `OptimizeDisk` — `fsutil behavior set DisableDeleteNotify 0` → Registry write
- **File:** `optimizer_windows.go`
- **Registry:** `HKLM\SYSTEM\CurrentControlSet\Control\FileSystem\DisableDeleteNotify = 0` (DWORD)

### `OptimizeDisk` — `schtasks /create` (defrag HDD) → Task Scheduler COM
- **File:** `optimizer_windows.go`
- **API:** `ITaskService` COM via `go-ole`
- Create weekly task `SysCleanerDefrag` for `defrag.exe C: /O`

## Section 4 — Scheduler (`pkg/scheduler/`)

### All three `schtasks` calls → Task Scheduler COM API
- **File:** `scheduler_windows.go` (replace existing stub + schtasks calls)
- **API:** `ITaskService → ITaskFolder → IRegisteredTask` via `go-ole`
- `CreateScheduledClean` → `ITaskService.NewTask()`, configure weekly trigger + action, `ITaskFolder.RegisterTaskDefinition()`
- `RemoveScheduledClean` → `ITaskFolder.DeleteTask("SysCleanerWeeklyClean")`
- `GetScheduledClean` → `ITaskFolder.GetTask("SysCleanerWeeklyClean")`, parse trigger/action

## Section 5 — GUI (`gui/views/extreme_mode.go`)

### Game launcher buttons → `ShellExecuteEx`
- **File:** `gui/views/extreme_mode.go` (Windows-specific helper)
- **API:** `shell32.dll!ShellExecuteExW` with `SW_SHOW`
- Idiomatic Windows app launch: respects UAC, file associations, no inherited console
- Move to a `launchExeNative(path string) error` helper in a new `gui/views/launch_windows.go`

## Cleanup

- Remove `os/exec` import from every file that no longer needs it
- Delete `runCmd` helper in `gaming/gaming.go` once all call sites are replaced
- Delete `getSysProcAttr()` helpers in `gaming/procattr_windows.go`, `optimizer/optimizer_windows.go`, `scheduler/scheduler_windows.go` — no longer needed
- Keep `_other.go` stubs for non-Windows builds

## File Change Summary

| File | Action |
|---|---|
| `pkg/cleaner/cleaner.go` | Remove `os/exec` import; `cleanDNSCache`, `cleanEventLogs`, `cleanRecycleBin` call new platform funcs |
| `pkg/cleaner/cleaner_windows.go` | Add `flushDNSCacheNative`, `clearEventLogNative`, `emptyRecycleBinNative` |
| `pkg/gaming/gaming.go` | Remove `os/exec` import; remove `runCmd`; call native helpers |
| `pkg/gaming/extreme.go` | Remove `os/exec` import; replace `stopWindowsExplorer`/`startWindowsExplorer` |
| `pkg/gaming/gaming_windows.go` | New: `setPowerSchemeNative`, `setTCPGamingParams`, `startExplorerNative` |
| `pkg/gaming/procattr_windows.go` | Delete (no longer needed) |
| `pkg/gaming/procattr_other.go` | Delete (no longer needed) |
| `pkg/optimizer/optimizer.go` | Remove `os/exec` import; call native helpers |
| `pkg/optimizer/optimizer_windows.go` | Add `setTCPOptimizationParam`, `isSSDPresentNative`, TRIM registry write, defrag task via COM |
| `pkg/scheduler/scheduler.go` | Remove `os/exec` import; delegate to native platform funcs |
| `pkg/scheduler/scheduler_windows.go` | Replace `getSysProcAttr` with full COM Task Scheduler implementation |
| `gui/views/extreme_mode.go` | Remove `os/exec` import; call `launchExeNative` |
| `gui/views/launch_windows.go` | New: `launchExeNative` via `ShellExecuteExW` |
| `gui/views/launch_other.go` | New: stub for non-Windows builds |
