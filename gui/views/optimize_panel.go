//go:build gui

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"syscleaner/pkg/optimizer"
)

// NewOptimizePanel creates the optimization controls view.
func NewOptimizePanel() fyne.CanvasObject {
	resultText := widget.NewMultiLineEntry()
	resultText.SetPlaceHolder("Optimization results will appear here...")
	resultText.Disable()
	resultText.SetMinRowsVisible(10)

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Stop()
	progressBar.Hide()

	statusLabel := widget.NewLabel("Ready.")

	// Startup optimization
	startupBtn := widget.NewButton("Optimize Startup Programs", nil)
	networkBtn := widget.NewButton("Optimize Network", nil)
	diskBtn := widget.NewButton("Optimize Disk", nil)
	allBtn := widget.NewButton("Run All Optimizations", nil)

	allBtns := []*widget.Button{startupBtn, networkBtn, diskBtn, allBtn}
	disableAll := func() {
		for _, b := range allBtns {
			b.Disable()
		}
	}
	enableAll := func() {
		for _, b := range allBtns {
			b.Enable()
		}
	}

	startupBtn.OnTapped = func() {
		disableAll()
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Optimizing startup programs...")

		go func() {
			defer enableAll()
			result := optimizer.OptimizeStartup()
			progressBar.Stop()
			progressBar.Hide()
			statusLabel.SetText("Startup optimization complete.")

			text := fmt.Sprintf("Startup Optimization:\n  Programs disabled: %d\n", result.Disabled)
			for _, p := range result.Programs {
				status := "kept"
				if p.Disabled {
					status = "DISABLED"
				}
				text += fmt.Sprintf("  [%s] %s (%s)\n", status, p.Name, p.Impact)
			}
			resultText.SetText(text)
		}()
	}
	startupBtn.Importance = widget.HighImportance

	// Network optimization
	networkBtn.OnTapped = func() {
		disableAll()
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Optimizing network settings...")

		go func() {
			defer enableAll()
			result := optimizer.OptimizeNetwork()
			progressBar.Stop()
			progressBar.Hide()
			statusLabel.SetText("Network optimization complete.")

			text := fmt.Sprintf("Network Optimization:\n  Estimated latency reduction: %dms\n\n", result.LatencyReduction)
			for _, opt := range result.Optimizations {
				text += fmt.Sprintf("  - %s\n", opt)
			}
			resultText.SetText(text)
		}()
	}
	networkBtn.Importance = widget.HighImportance

	// Disk optimization
	diskBtn.OnTapped = func() {
		disableAll()
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Optimizing disk...")

		go func() {
			defer enableAll()
			result := optimizer.OptimizeDisk()
			progressBar.Stop()
			progressBar.Hide()
			statusLabel.SetText("Disk optimization complete.")

			diskType := "HDD"
			if result.IsSSD {
				diskType = "SSD"
			}
			text := fmt.Sprintf("Disk Optimization:\n  Disk type: %s\n", diskType)
			if result.Scheduled {
				if result.IsSSD {
					text += "  TRIM enabled for optimal SSD performance\n"
				} else {
					text += "  Weekly defragmentation scheduled\n"
				}
			}
			resultText.SetText(text)
		}()
	}
	diskBtn.Importance = widget.HighImportance

	// Run all
	allBtn.OnTapped = func() {
		disableAll()
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Running all optimizations...")

		go func() {
			defer enableAll()
			text := ""

			// Startup
			startupResult := optimizer.OptimizeStartup()
			text += fmt.Sprintf("Startup: %d programs disabled\n", startupResult.Disabled)

			// Network
			netResult := optimizer.OptimizeNetwork()
			text += fmt.Sprintf("Network: %dms latency reduction, %d optimizations\n",
				netResult.LatencyReduction, len(netResult.Optimizations))

			// Disk
			diskResult := optimizer.OptimizeDisk()
			diskType := "HDD"
			if diskResult.IsSSD {
				diskType = "SSD"
			}
			text += fmt.Sprintf("Disk: %s optimized\n", diskType)

			progressBar.Stop()
			progressBar.Hide()
			statusLabel.SetText("All optimizations complete!")
			resultText.SetText(text)
		}()
	}
	allBtn.Importance = widget.WarningImportance

	buttonGrid := container.NewGridWithColumns(3,
		startupBtn,
		networkBtn,
		diskBtn,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("System Optimization", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Select an optimization to run:"),
		buttonGrid,
		widget.NewSeparator(),
		allBtn,
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		resultText,
	)

	return container.NewScroll(container.NewPadded(content))
}
