//go:build gui

package views

import (
	"fmt"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/mem"
	"github.com/shirou/gopsutil/v3/net"

	"syscleaner/pkg/gaming"
	sysmem "syscleaner/pkg/memory"
)

// NewMonitorPanel creates the real-time monitoring view with RAM monitoring
func NewMonitorPanel() fyne.CanvasObject {
	cpuLabel := widget.NewLabel("CPU: --")
	ramLabel := widget.NewLabel("RAM: --")
	netLabel := widget.NewLabel("Network: --")

	cpuProgress := widget.NewProgressBar()
	ramProgress := widget.NewProgressBar()

	// RAM Monitor Section
	ramTotalLabel := widget.NewLabel("Total: --")
	ramUsedLabel := widget.NewLabel("Used: --")
	ramFreeLabel := widget.NewLabel("Free: --")
	ramStandbyLabel := widget.NewLabel("Standby: --")
	ramTrimCountLabel := widget.NewLabel("Trim Count: 0")

	ramTrimStatusLabel := widget.NewLabel("Last Trim: Never")
	ramTrimStatusLabel.Wrapping = fyne.TextWrapWord

	ramUsedBar := widget.NewProgressBar()
	ramFreeBar := widget.NewProgressBar()
	ramStandbyBar := widget.NewProgressBar()

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

	logText := widget.NewMultiLineEntry()
	logText.SetPlaceHolder("System events will appear here...")
	logText.Disable()
	logText.SetMinRowsVisible(12)

	var logMu sync.Mutex
	addLog := func(message string, isWarning bool) {
		logMu.Lock()
		defer logMu.Unlock()
		timestamp := time.Now().Format("15:04:05")
		prefix := ""
		if isWarning {
			prefix = "⚠️ "
		}
		entry := fmt.Sprintf("[%s] %s%s\n", timestamp, prefix, message)
		current := logText.Text
		// Keep log manageable - trim after 5000 chars
		if len(current) > 5000 {
			current = current[:5000]
		}
		logText.SetText(entry + current)
	}

	// Track previous network counters for rate calculation
	var prevBytesRecv, prevBytesSent uint64
	var prevTime time.Time

	// Start monitoring
	go func() {
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		lastGameModeCheck := false
		lastExtremeModeCheck := false

		for range ticker.C {
			// CPU usage
			if cpuPercent, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(cpuPercent) > 0 {
				cpuProgress.SetValue(cpuPercent[0] / 100.0)
				cpuLabel.SetText(fmt.Sprintf("CPU: %.1f%%", cpuPercent[0]))

				if cpuPercent[0] > 90 {
					addLog(fmt.Sprintf("HIGH CPU: %.1f%%", cpuPercent[0]), true)
				}
			}

			// RAM usage
			if vmem, err := mem.VirtualMemory(); err == nil {
				ramProgress.SetValue(vmem.UsedPercent / 100.0)
				ramLabel.SetText(fmt.Sprintf("RAM: %.1f%% (%.1f GB / %.1f GB)",
					vmem.UsedPercent,
					float64(vmem.Used)/1024/1024/1024,
					float64(vmem.Total)/1024/1024/1024))

				if vmem.UsedPercent > 90 {
					addLog(fmt.Sprintf("HIGH RAM: %.1f%%", vmem.UsedPercent), true)
				}
			}

			// RAM Monitor Stats — always update
			stats := sysmem.GetCurrentStats()
			ramTotalLabel.SetText(fmt.Sprintf("Total: %.2f GB", stats.TotalGB))
			ramUsedLabel.SetText(fmt.Sprintf("Used: %.2f GB (%.1f%%)", stats.UsedGB, stats.UsedPercent))
			ramFreeLabel.SetText(fmt.Sprintf("Free: %.2f GB (%.1f%%)", stats.FreeGB, stats.FreePercent))
			ramStandbyLabel.SetText(fmt.Sprintf("Standby: %.2f GB (%.1f%%)", stats.StandbyGB, stats.StandbyPercent))
			ramTrimCountLabel.SetText(fmt.Sprintf("Trim Count: %d", stats.TrimCount))
			ramUsedBar.SetValue(stats.UsedPercent / 100.0)
			ramFreeBar.SetValue(stats.FreePercent / 100.0)
			ramStandbyBar.SetValue(stats.StandbyPercent / 100.0)
			if !stats.LastTrimTime.IsZero() {
				ramTrimStatusLabel.SetText(fmt.Sprintf("Last trim: %s ago", time.Since(stats.LastTrimTime).Round(time.Second)))
			}

			// Network usage (calculate rate)
			if netIO, err := net.IOCounters(false); err == nil && len(netIO) > 0 {
				now := time.Now()
				if !prevTime.IsZero() {
					elapsed := now.Sub(prevTime).Seconds()
					if elapsed > 0 {
						recvRate := float64(netIO[0].BytesRecv-prevBytesRecv) / elapsed / 1024 / 1024
						sentRate := float64(netIO[0].BytesSent-prevBytesSent) / elapsed / 1024 / 1024
						netLabel.SetText(fmt.Sprintf("Network: %.2f MB/s down | %.2f MB/s up", recvRate, sentRate))
					}
				}
				prevBytesRecv = netIO[0].BytesRecv
				prevBytesSent = netIO[0].BytesSent
				prevTime = now
			}

			// Log gaming mode status changes
			gameModeActive := gaming.IsEnabled()
			extremeModeActive := gaming.IsExtremeModeActive()

			if extremeModeActive && !lastExtremeModeCheck {
				addLog("🔥 Extreme Mode ACTIVATED - Maximum Performance", false)
				lastExtremeModeCheck = true
			} else if !extremeModeActive && lastExtremeModeCheck {
				addLog("Extreme Mode Deactivated", false)
				lastExtremeModeCheck = false
			}

			if gameModeActive && !lastGameModeCheck && !extremeModeActive {
				addLog("🎮 Gaming Mode Activated", false)
				lastGameModeCheck = true
			} else if !gameModeActive && lastGameModeCheck {
				addLog("Gaming Mode Deactivated", false)
				lastGameModeCheck = false
			}
		}
	}()

	// CPU section
	cpuSection := container.NewVBox(
		widget.NewLabelWithStyle("CPU Usage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		cpuLabel,
		cpuProgress,
	)

	// RAM section
	ramSection := container.NewVBox(
		widget.NewLabelWithStyle("Memory Usage", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		ramLabel,
		ramProgress,
	)

	// Network section
	netSection := container.NewVBox(
		widget.NewLabelWithStyle("Network", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		netLabel,
	)

	metrics := container.NewGridWithColumns(3,
		cpuSection,
		ramSection,
		netSection,
	)

	// RAM Monitor Section
	ramMonitorSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("RAM Monitor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		container.NewBorder(nil, nil, widget.NewLabel("Total RAM:"), nil, ramTotalLabel),
		ramUsedLabel,
		ramUsedBar,
		ramFreeLabel,
		ramFreeBar,
		ramStandbyLabel,
		ramStandbyBar,
		ramTrimCountLabel,
		trimNowBtn,
		ramTrimStatusLabel,
		widget.NewLabel("Auto-trim activates when free memory drops below 15% (Extreme Mode only)"),
	)

	// Clear log button
	clearBtn := widget.NewButton("Clear Log", func() {
		logMu.Lock()
		defer logMu.Unlock()
		logText.SetText("")
	})

	logsHeader := container.NewBorder(
		nil, nil, widget.NewLabelWithStyle("System Events", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}), clearBtn,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("Real-Time Monitor", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		metrics,
		ramMonitorSection,
		widget.NewSeparator(),
		logsHeader,
		logText,
	)

	return container.NewScroll(container.NewPadded(content))
}
