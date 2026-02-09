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
)

// NewMonitorPanel creates the real-time monitoring view.
func NewMonitorPanel() fyne.CanvasObject {
	cpuLabel := widget.NewLabel("CPU: --")
	ramLabel := widget.NewLabel("RAM: --")
	netLabel := widget.NewLabel("Network: --")

	cpuProgress := widget.NewProgressBar()
	ramProgress := widget.NewProgressBar()

	logText := widget.NewMultiLineEntry()
	logText.SetPlaceHolder("System events will appear here...")
	logText.Disable()
	logText.SetMinRowsVisible(15)

	var logMu sync.Mutex
	addLog := func(message string) {
		logMu.Lock()
		defer logMu.Unlock()
		timestamp := time.Now().Format("15:04:05")
		entry := fmt.Sprintf("[%s] %s\n", timestamp, message)
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

		for range ticker.C {
			// CPU usage
			if cpuPercent, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(cpuPercent) > 0 {
				cpuProgress.SetValue(cpuPercent[0] / 100.0)
				cpuLabel.SetText(fmt.Sprintf("CPU: %.1f%%", cpuPercent[0]))

				if cpuPercent[0] > 90 {
					addLog(fmt.Sprintf("HIGH CPU: %.1f%%", cpuPercent[0]))
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
					addLog(fmt.Sprintf("HIGH RAM: %.1f%%", vmem.UsedPercent))
				}
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
			if gaming.IsExtremeModeActive() {
				addLog("Extreme Mode Active - Maximum Performance")
			} else if gaming.IsEnabled() {
				addLog("Gaming Mode Active")
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
		widget.NewSeparator(),
		logsHeader,
		logText,
	)

	return container.NewScroll(container.NewPadded(content))
}
