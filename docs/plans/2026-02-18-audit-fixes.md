# Audit Fixes (2026-02-18) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all HIGH-impact audit findings across performance, dead code, and UX categories identified on 2026-02-18.

**Architecture:** All changes are surgical — no new packages, no new abstractions unless required. Fixes follow the existing GUI build-tag pattern (`//go:build gui`). Dead code is deleted, not refactored. Race conditions are fixed with `sync/atomic` or by confining mutable state to one goroutine. UI blocking is resolved by dispatching goroutines. Each task targets one concern and can be committed independently.

**Tech Stack:** Go 1.21, Fyne v2, gopsutil/v3, `sync/atomic`, standard library `context`/`time`.

---

### Task 1: Fix Button-Disabling (UX-01, UX-02, UX-03)

**Problem:** `analyzeBtn`, `cleanBtn`, `startupBtn`, `networkBtn`, `diskBtn`, `allBtn`, and `trimNowBtn` are never disabled before their goroutine runs, allowing concurrent double-clicks.

**Files:**
- Modify: `gui/views/clean_panel.go`
- Modify: `gui/views/optimize_panel.go`
- Modify: `gui/views/monitor_panel.go`

Read each file fully before editing. The pattern for each button: collect all buttons into a slice, disable all at the top of the goroutine, re-enable with `defer enableAll()`.

For `clean_panel.go`: create `analyzeBtn` and `cleanBtn` with `nil` OnTapped first, then assign. Build a shared `disableAll`/`enableAll` helper over `[]*widget.Button{analyzeBtn, cleanBtn}`. Each button's handler calls `disableAll()`, shows progress, launches a goroutine with `defer enableAll()`.

For `optimize_panel.go`: same pattern over `[]*widget.Button{startupBtn, networkBtn, diskBtn, allBtn}`.

For `monitor_panel.go`: change the existing `trimNowBtn` callback to:
```go
trimNowBtn := widget.NewButton("Trim RAM Now", nil)
trimNowBtn.OnTapped = func() {
    trimNowBtn.Disable()
    ramTrimStatusLabel.SetText("Trimming...")
    go func() {
        defer trimNowBtn.Enable()
        if err := sysmem.TrimNow(); err != nil {
            ramTrimStatusLabel.SetText(fmt.Sprintf("Trim failed: %v", err))
        } else {
            ramTrimStatusLabel.SetText(fmt.Sprintf("Last trim: %s", time.Now().Format("15:04:05")))
        }
    }()
}
```

Build: `go build -tags gui ./...`
Commit: `fix(ux): disable action buttons during operations to prevent concurrent runs`

---

### Task 2: Fix Log Mutex and Warn Spam in monitor_panel.go (PERF-01, PERF-02, UX-05)

**Problem:**
- `addLog` holds `logMu` while calling `logText.SetText(...)` — SetText dispatches to Fyne's render loop while the mutex is held, creating a latent deadlock surface.
- `current := logText.Text` reads a widget field from a background goroutine — unsafe.
- CPU > 90% and RAM > 90% warn entries are added every second they persist, flooding the log.

**Files:**
- Modify: `gui/views/monitor_panel.go`

**Fix addLog:** Maintain a `logBuffer string` variable (not `logText.Text`). Build the new string under the lock, release the lock, then call `logText.SetText` outside the lock:

```go
var logMu sync.Mutex
var logBuffer string

addLog := func(message string, isWarning bool) {
    timestamp := time.Now().Format("15:04:05")
    prefix := ""
    if isWarning {
        prefix = "⚠️ "
    }
    entry := fmt.Sprintf("[%s] %s%s\n", timestamp, prefix, message)

    logMu.Lock()
    logBuffer = entry + logBuffer
    if len(logBuffer) > 5000 {
        logBuffer = logBuffer[:5000]
    }
    snapshot := logBuffer
    logMu.Unlock()

    logText.SetText(snapshot)
}
```

**Fix Clear button:** Also update the clear button to clear `logBuffer` under lock, then call `SetText("")` outside:
```go
clearBtn := widget.NewButton("Clear Log", func() {
    logMu.Lock()
    logBuffer = ""
    logMu.Unlock()
    logText.SetText("")
})
```

**Fix warn spam:** Add cooldown tracking variables after the ticker is created:
```go
var lastCPUWarn, lastRAMWarn time.Time
const warnCooldown = 30 * time.Second
```

Then gate each warning:
```go
if cpuPercent[0] > 90 && time.Since(lastCPUWarn) > warnCooldown {
    lastCPUWarn = time.Now()
    addLog(fmt.Sprintf("HIGH CPU: %.1f%%", cpuPercent[0]), true)
}
// and for RAM:
if vmem.UsedPercent > 90 && time.Since(lastRAMWarn) > warnCooldown {
    lastRAMWarn = time.Now()
    addLog(fmt.Sprintf("HIGH RAM: %.1f%%", vmem.UsedPercent), true)
}
```

Build: `go build -tags gui ./...`
Commit: `fix(perf): move SetText outside logMu lock; add 30s cooldown on CPU/RAM warnings`

---

### Task 3: Deduplicate VirtualMemory Call and Fix Blocking CPU Poll (PERF-03, PERF-04)

**Problem:**
- `monitor_panel.go` calls `mem.VirtualMemory()` directly AND then calls `sysmem.GetCurrentStats()` which calls it again — two kernel calls per tick for the same data.
- `dashboard.go` calls `cpu.Percent(500ms, false)` — a 500ms blocking call on a 2s ticker. Use `cpu.Percent(0, false)` instead (reads accumulated kernel counters, non-blocking).

**Files:**
- Modify: `gui/views/monitor_panel.go`
- Modify: `gui/views/dashboard.go`

**Fix monitor_panel.go:** Remove the explicit `mem.VirtualMemory()` block. Derive the top RAM bar and label from `sysmem.GetCurrentStats()` instead:

```go
// Single call — covers both the RAM monitor section and the top summary bar
stats := sysmem.GetCurrentStats()

// Top-level RAM bar (was using mem.VirtualMemory directly)
ramProgress.SetValue(stats.UsedPercent / 100.0)
ramLabel.SetText(fmt.Sprintf("RAM: %.1f%% (%.1f GB / %.1f GB)",
    stats.UsedPercent, stats.UsedGB, stats.TotalGB))

if stats.UsedPercent > 90 && time.Since(lastRAMWarn) > warnCooldown {
    lastRAMWarn = time.Now()
    addLog(fmt.Sprintf("HIGH RAM: %.1f%%", stats.UsedPercent), true)
}

// RAM Monitor section
ramTotalLabel.SetText(fmt.Sprintf("Total: %.2f GB", stats.TotalGB))
ramUsedLabel.SetText(fmt.Sprintf("Used: %.2f GB (%.1f%%)", stats.UsedGB, stats.UsedPercent))
ramFreeLabel.SetText(fmt.Sprintf("Free: %.2f GB (%.1f%%)", stats.FreeGB, stats.FreePercent))
ramStandbyLabel.SetText(fmt.Sprintf("Standby: %.2f GB (%.1f%%)", stats.StandbyGB, stats.StandbyPercent))
ramTrimCountLabel.SetText(fmt.Sprintf("Trim Count: %d", stats.TrimCount))
ramUsedBar.SetValue(stats.UsedPercent / 100.0)
ramFreeBar.SetValue(stats.FreePercent / 100.0)
ramStandbyBar.SetValue(stats.StandbyPercent / 100.0)
if !stats.LastTrimTime.IsZero() {
    ramTrimStatusLabel.SetText(fmt.Sprintf("Last trim: %s ago",
        time.Since(stats.LastTrimTime).Round(time.Second)))
}
```

Remove the `"github.com/shirou/gopsutil/v3/mem"` import from monitor_panel.go if it's no longer used.

**Fix dashboard.go:** Change `cpu.Percent(500*time.Millisecond, false)` to `cpu.Percent(0, false)`.

Build: `go build -tags gui ./...`
Commit: `fix(perf): single VirtualMemory call per tick in monitor; non-blocking cpu.Percent in dashboard`

---

### Task 4: Fix Data Race in animated_score.go (PERF-06)

**Problem:** `ScoreRing` fields `animating`, `displayScore`, `score` are accessed from the caller goroutine (via `SetScore`) and the animation goroutine (via `animateTo`) with no synchronisation. The Go race detector flags this.

**Files:**
- Modify: `gui/views/animated_score.go`

Add `mu sync.Mutex` to the struct. Add `"sync"` to imports.

Protect all reads and writes to `animating`, `displayScore`, and `score` with `s.mu.Lock()`/`s.mu.Unlock()`. Critically, release the lock before calling `s.Refresh()` to avoid holding the mutex across Fyne render dispatches.

In `SetScore`:
```go
func (s *ScoreRing) SetScore(score int) {
    if score < 0 { score = 0 }
    if score > 100 { score = 100 }
    s.mu.Lock()
    s.score = score
    alreadyRunning := s.animating
    s.mu.Unlock()
    if !alreadyRunning {
        go s.animateTo(score)
    }
}
```

In `animateTo`, lock only around field access, unlock before `s.Refresh()`:
```go
func (s *ScoreRing) animateTo(target int) {
    s.mu.Lock()
    s.animating = true
    s.mu.Unlock()
    defer func() {
        s.mu.Lock()
        s.animating = false
        s.mu.Unlock()
    }()

    targetFloat := float64(target)
    for {
        s.mu.Lock()
        cur := s.displayScore
        s.mu.Unlock()
        if math.Abs(cur-targetFloat) < 0.5 { break }

        s.mu.Lock()
        if s.displayScore < targetFloat {
            s.displayScore = math.Min(s.displayScore+0.8, targetFloat)
        } else {
            s.displayScore = math.Max(s.displayScore-0.8, targetFloat)
        }
        s.mu.Unlock()

        s.Refresh()
        time.Sleep(15 * time.Millisecond)
    }
    s.mu.Lock()
    s.displayScore = targetFloat
    s.mu.Unlock()
    s.Refresh()
}
```

Add `"math"` to imports.

In the renderer's `Refresh()`, snapshot `displayScore` under lock before using it:
```go
func (r *scoreRingRenderer) Refresh() {
    r.score.mu.Lock()
    ds := r.score.displayScore
    scoreColor := r.score.getColorForScore() // reads displayScore — lock held
    r.score.mu.Unlock()
    // use ds and scoreColor below
    ...
}
```

Build: `go build -tags gui ./...`
Commit: `fix(race): protect ScoreRing fields with mutex to eliminate data race`

---

### Task 5: Fix gaming.GetStatus Mutex Contention and Extreme Mode UI Blocking (PERF-08, UX-09)

**Problem 1:** `gaming.GetStatus()` holds `mu` while calling `cpu.Percent(500ms, ...)`, blocking the mutex for 500ms.
**Problem 2:** `extreme_mode.go` calls `gaming.EnableExtremeMode()` / `gaming.DisableExtremeMode()` synchronously inside Fyne event callbacks, freezing the UI.

**Files:**
- Modify: `pkg/gaming/gaming.go`
- Modify: `gui/views/extreme_mode.go`

**Fix gaming.go:** In `GetStatus()`, snapshot protected state under the lock, release it, then perform blocking calls (cpu.Percent, mem.VirtualMemory, process list) outside the lock:

```go
func GetStatus() Status {
    mu.Lock()
    status := Status{
        Enabled:         gamingModeEnabled,
        StoppedServices: append([]string(nil), stoppedServices...),
    }
    mu.Unlock()
    // All blocking calls happen here, outside the lock:
    if cpuPercent, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(cpuPercent) > 0 {
        status.CPUUsage = cpuPercent[0]
    }
    // ... rest of the function unchanged
}
```

**Fix extreme_mode.go:** In `toggleExtremeMode()`, wrap both Enable and Disable paths in goroutines. Disable the toggle button before launching and re-enable with defer:

```go
func (p *extremeModePanel) toggleExtremeMode() {
    if p.isActive {
        p.toggleBtn.Disable()
        go func() {
            defer p.toggleBtn.Enable()
            if err := gaming.DisableExtremeMode(); err != nil {
                dialog.ShowError(err, p.window)
                return
            }
            p.isActive = false
            dialog.ShowInformation("Extreme Mode Disabled", "System restored to normal mode.", p.window)
            p.updateUI()
        }()
        return
    }
    dialog.ShowConfirm(
        "Activate Extreme Performance Mode?",
        "This will:\n\n"+
            "  - Stop Windows Explorer (no desktop/taskbar)\n"+
            "  - Stop all non-essential services\n"+
            "  - Close background apps (respecting whitelist)\n"+
            "  - Maximize game performance\n\n"+
            "You can only launch games from this window.\nContinue?",
        func(confirmed bool) {
            if confirmed {
                p.toggleBtn.Disable()
                go func() {
                    defer p.toggleBtn.Enable()
                    if err := gaming.EnableExtremeMode(); err != nil {
                        dialog.ShowError(err, p.window)
                        return
                    }
                    p.isActive = true
                    dialog.ShowInformation("Extreme Mode Activated", "System optimized for maximum performance!", p.window)
                    p.updateUI()
                }()
            }
        },
        p.window,
    )
}
```

Build: `go build -tags gui ./...`
Commit: `fix(perf): release gaming mutex before blocking cpu.Percent; dispatch extreme mode ops to goroutine`

---

### Task 6: Replace goroutine-per-file and Fix Memory Monitor Sleep (PERF-10, PERF-11)

**Problem 1:** `removeWithTimeout` in `pkg/cleaner/cleaner.go` spawns a goroutine per file removal. For thousands of files, this causes thousands of goroutines. Direct `os.Remove` is sufficient.
**Problem 2:** `StartContinuousMonitor` in `pkg/memory/memory_windows.go` uses `time.Sleep(2s)` inside a select loop, which can't be interrupted by the `monitorDone` channel.

**Files:**
- Modify: `pkg/cleaner/cleaner.go`
- Modify: `pkg/memory/memory_windows.go`

**Fix cleaner.go:** Delete the `removeWithTimeout` function entirely. Replace all call sites with `os.Remove(path)`. The error classification at each call site already handles the result; the timeout wrapper added no value for local filesystem operations.

**Fix memory_windows.go:** Replace `time.Sleep(2 * time.Second)` with a select:
```go
select {
case <-monitorDone:
    return
case <-time.After(2 * time.Second):
}
```

Build: `go build ./...` and `go test ./pkg/cleaner/... ./pkg/memory/...`
Commit: `fix(perf): direct os.Remove instead of goroutine-per-file; make monitor post-trim sleep cancellable`

---

### Task 7: Delete Orphaned Dead Code (DEAD-01, DEAD-02, DEAD-04, DEAD-05)

**Problem:** Multiple packages and symbols exist but are never imported or called anywhere.

**Files:**
- Delete: `pkg/monitor/` directory (entire package — never imported)
- Delete: `pkg/logger/` directory (entire package — never imported)
- Modify: `pkg/gaming/extreme.go` — delete `GetExtremeModeStats()` (no callers)
- Modify: `pkg/optimizer/optimizer.go` — delete `Results` type (never used)

**Step 1:** Confirm zero callers:
```bash
grep -r "syscleaner/pkg/monitor\|syscleaner/pkg/logger\|GetExtremeModeStats\|optimizer\.Results" . --include="*.go"
```
Only the definition files themselves should match.

**Step 2:** Delete:
```bash
rm -rf pkg/monitor pkg/logger
```

**Step 3:** Remove `GetExtremeModeStats` from `pkg/gaming/extreme.go` and `Results` from `pkg/optimizer/optimizer.go`.

Build: `go build -tags gui ./...` and `go test ./...`
Commit: `chore: delete orphaned pkg/monitor, pkg/logger, GetExtremeModeStats, and unused Results type`

---

### Task 8: Fix Config Extension and Optimize Panel Result Height (UX-16, UX-12)

**Problem 1:** `pkg/config/config.go` encodes config as JSON but saves it as `config.yaml`. Confusing for users who try to edit it manually.
**Problem 2:** `gui/views/optimize_panel.go` result text area has no `SetMinRowsVisible` call, collapsing to a tiny box.

**Files:**
- Modify: `pkg/config/config.go`
- Modify: `gui/views/optimize_panel.go`

**Fix config.go:** Change both occurrences of `"config.yaml"` to `"config.json"`. Existing `config.yaml` files on disk will be silently ignored (treated as missing, defaults loaded) — correct behaviour since the old file was JSON-encoded anyway.

**Fix optimize_panel.go:** After `resultText.Disable()`, add:
```go
resultText.SetMinRowsVisible(10)
```

Build: `go build -tags gui ./...` and `go test ./pkg/config/...`
Commit: `fix(ux): rename config file to .json; set min 10 rows on optimize result text`

---

## Summary

| # | Task | Findings Fixed | Risk |
|---|------|----------------|------|
| 1 | Disable buttons during operations | UX-01, UX-02, UX-03 | Low |
| 2 | Fix log mutex and warn spam | PERF-01, PERF-02, UX-05 | Low |
| 3 | Deduplicate VirtualMemory + CPU poll | PERF-03, PERF-04 | Medium |
| 4 | Fix ScoreRing data race | PERF-06 | Medium |
| 5 | Fix gaming mutex + extreme mode UI block | PERF-08, UX-09 | Medium |
| 6 | Replace goroutine-per-file + fix monitor sleep | PERF-10, PERF-11 | Low |
| 7 | Delete dead code | DEAD-01, DEAD-02, DEAD-04, DEAD-05 | Low |
| 8 | Config extension + resultText height | UX-16, UX-12 | Low |
