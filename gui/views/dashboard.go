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

// NewDashboard creates the main dashboard view with animated score ring
func NewDashboard() fyne.CanvasObject {
	// System info section
	sysInfoLabel := widget.NewLabel("Loading system info...")
	sysInfoLabel.Wrapping = fyne.TextWrapWord

	// Animated performance score ring
	scoreRing := NewScoreRing()
	scoreDesc := widget.NewLabelWithStyle("Performance Score", fyne.TextAlignCenter, fyne.TextStyle{})

	// Status indicators with smooth animations
	cpuBar := widget.NewProgressBar()
	ramBar := widget.NewProgressBar()
	diskBar := widget.NewProgressBar()
	cpuLabel := widget.NewLabel("CPU: --")
	ramLabel := widget.NewLabel("RAM: --")
	diskLabel := widget.NewLabel("Disk: --")

	// Track previous values for smooth transitions
	var prevCPU, prevRAM, prevDisk float64

	// Gaming mode status with pulsing effect
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

	// Real-time update goroutine with smooth animations
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for range ticker.C {
			// CPU with smooth transition
			if cpuPercent, err := cpu.Percent(500*time.Millisecond, false); err == nil && len(cpuPercent) > 0 {
				targetCPU := cpuPercent[0] / 100.0
				// Smooth interpolation
				smoothCPU := prevCPU + (targetCPU-prevCPU)*0.3
				prevCPU = smoothCPU
				cpuBar.SetValue(smoothCPU)
				cpuLabel.SetText(fmt.Sprintf("CPU: %.1f%%", smoothCPU*100))
			}

			// RAM with smooth transition
			if vmem, err := mem.VirtualMemory(); err == nil {
				targetRAM := vmem.UsedPercent / 100.0
				smoothRAM := prevRAM + (targetRAM-prevRAM)*0.3
				prevRAM = smoothRAM
				ramBar.SetValue(smoothRAM)
				ramLabel.SetText(fmt.Sprintf("RAM: %.1f%% (%.1f / %.1f GB)",
					smoothRAM*100,
					float64(vmem.Used)/1024/1024/1024,
					float64(vmem.Total)/1024/1024/1024))
			}

			// Disk with smooth transition
			if usage, err := disk.Usage("/"); err == nil {
				targetDisk := usage.UsedPercent / 100.0
				smoothDisk := prevDisk + (targetDisk-prevDisk)*0.3
				prevDisk = smoothDisk
				diskBar.SetValue(smoothDisk)
				diskLabel.SetText(fmt.Sprintf("Disk: %.1f%% (%.1f / %.1f GB)",
					smoothDisk*100,
					float64(usage.Used)/1024/1024/1024,
					float64(usage.Total)/1024/1024/1024))
			}

			// Performance score with animation
			score := calculateDashboardScore(prevCPU, prevRAM, prevDisk)
			scoreRing.SetScore(score)

			// Mode status with styling
			if gaming.IsExtremeModeActive() {
				extremeStatus.SetText("⚡ EXTREME MODE ACTIVE ⚡")
				gamingStatus.SetText("🎮 Gaming Mode: ACTIVE")
			} else if gaming.IsEnabled() {
				gamingStatus.SetText("🎮 Gaming Mode: ACTIVE")
				extremeStatus.SetText("Extreme Mode: Inactive")
			} else {
				gamingStatus.SetText("Gaming Mode: Inactive")
				extremeStatus.SetText("Extreme Mode: Inactive")
			}
		}
	}()

	// Layout
	scoreSection := container.NewCenter(
		container.NewVBox(
			container.NewCenter(scoreRing),
			container.NewCenter(scoreDesc),
		),
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
	score -= cpuLoad * 30.0  // CPU weight 30%
	score -= ramLoad * 30.0  // RAM weight 30%
	score -= diskLoad * 20.0 // Disk weight 20%
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
