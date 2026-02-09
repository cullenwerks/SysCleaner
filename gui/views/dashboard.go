//go:build gui

package views

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/host"
	"github.com/shirou/gopsutil/v3/mem"

	"syscleaner/pkg/gaming"
)

// NewDashboard creates the main dashboard view.
func NewDashboard() fyne.CanvasObject {
	// System info section
	sysInfoLabel := widget.NewLabel("Loading system info...")
	sysInfoLabel.Wrapping = fyne.TextWrapWord

	// Performance score
	scoreLabel := widget.NewLabelWithStyle("--", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})
	scoreDesc := widget.NewLabelWithStyle("Performance Score", fyne.TextAlignCenter, fyne.TextStyle{})

	// Status indicators
	cpuBar := widget.NewProgressBar()
	ramBar := widget.NewProgressBar()
	diskBar := widget.NewProgressBar()
	cpuLabel := widget.NewLabel("CPU: --")
	ramLabel := widget.NewLabel("RAM: --")
	diskLabel := widget.NewLabel("Disk: --")

	// Gaming mode status
	gamingStatus := widget.NewLabel("Gaming Mode: Inactive")
	extremeStatus := widget.NewLabel("Extreme Mode: Inactive")

	// Load system info
	go func() {
		if info, err := host.Info(); err == nil {
			sysInfoLabel.SetText(fmt.Sprintf("OS: %s %s | Hostname: %s | Uptime: %s",
				info.Platform, info.PlatformVersion, info.Hostname,
				(time.Duration(info.Uptime) * time.Second).String()))
		}
	}()

	// Real-time update goroutine
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			// CPU
			if cpuPercent, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(cpuPercent) > 0 {
				cpuBar.SetValue(cpuPercent[0] / 100.0)
				cpuLabel.SetText(fmt.Sprintf("CPU: %.1f%%", cpuPercent[0]))
			}

			// RAM
			if vmem, err := mem.VirtualMemory(); err == nil {
				ramBar.SetValue(vmem.UsedPercent / 100.0)
				ramLabel.SetText(fmt.Sprintf("RAM: %.1f%% (%.1f / %.1f GB)",
					vmem.UsedPercent,
					float64(vmem.Used)/1024/1024/1024,
					float64(vmem.Total)/1024/1024/1024))
			}

			// Disk
			if usage, err := disk.Usage("/"); err == nil {
				diskBar.SetValue(usage.UsedPercent / 100.0)
				diskLabel.SetText(fmt.Sprintf("Disk: %.1f%% (%.1f / %.1f GB)",
					usage.UsedPercent,
					float64(usage.Used)/1024/1024/1024,
					float64(usage.Total)/1024/1024/1024))
			}

			// Performance score
			score := calculateDashboardScore(cpuBar.Value, ramBar.Value, diskBar.Value)
			scoreLabel.SetText(fmt.Sprintf("%d", score))

			// Mode status
			if gaming.IsExtremeModeActive() {
				extremeStatus.SetText("Extreme Mode: ACTIVE")
				gamingStatus.SetText("Gaming Mode: ACTIVE")
			} else if gaming.IsEnabled() {
				gamingStatus.SetText("Gaming Mode: ACTIVE")
				extremeStatus.SetText("Extreme Mode: Inactive")
			} else {
				gamingStatus.SetText("Gaming Mode: Inactive")
				extremeStatus.SetText("Extreme Mode: Inactive")
			}
		}
	}()

	// Layout
	scoreSection := container.NewVBox(
		container.NewCenter(scoreLabel),
		container.NewCenter(scoreDesc),
	)

	metricsSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("System Metrics", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		cpuLabel, cpuBar,
		ramLabel, ramBar,
		diskLabel, diskBar,
	)

	statusSection := container.NewVBox(
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Mode Status", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		gamingStatus,
		extremeStatus,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("SysCleaner Dashboard", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		sysInfoLabel,
		scoreSection,
		metricsSection,
		statusSection,
	)

	return container.NewScroll(container.NewPadded(content))
}

func calculateDashboardScore(cpuLoad, ramLoad, diskLoad float64) int {
	// Lower resource usage = higher score
	score := 100.0
	score -= cpuLoad * 30.0       // CPU weight 30%
	score -= ramLoad * 30.0       // RAM weight 30%
	score -= diskLoad * 20.0      // Disk weight 20%
	// Remaining 20% for mode bonuses
	if gaming.IsExtremeModeActive() {
		score += 10
	} else if gaming.IsEnabled() {
		score += 5
	}

	if score < 0 {
		score = 0
	}
	if score > 100 {
		score = 100
	}
	return int(score)
}
