//go:build gui

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"syscleaner/pkg/tools"
)

// NewToolsPanel creates the advanced tools view.
func NewToolsPanel() fyne.CanvasObject {
	resultText := widget.NewMultiLineEntry()
	resultText.SetPlaceHolder("Tool results will appear here...")
	resultText.Disable()
	resultText.SetMinRowsVisible(12)

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Stop()
	progressBar.Hide()

	statusLabel := widget.NewLabel("Ready.")

	// Deep Registry Clean
	registryBtn := widget.NewButton("Deep Registry Clean", func() {
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Performing deep registry clean (backup first)...")

		go func() {
			result, err := tools.DeepRegistryClean()
			progressBar.Stop()
			progressBar.Hide()

			if err != nil {
				statusLabel.SetText("Registry clean failed.")
				resultText.SetText(fmt.Sprintf("Error: %v", err))
				return
			}

			statusLabel.SetText("Registry clean complete.")
			resultText.SetText(fmt.Sprintf("Deep Registry Clean:\n  Items cleaned: %d\n  Backup saved to: %s\n  Duration: %s",
				result.ItemsCleaned, result.BackupPath, result.Duration))
		}()
	})
	registryBtn.Importance = widget.HighImportance

	// System Image Scan
	scanBtn := widget.NewButton("Scan System Image (DISM)", func() {
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Scanning system image with DISM (this may take a while)...")

		go func() {
			result, err := tools.ScanSystemImage()
			progressBar.Stop()
			progressBar.Hide()

			if err != nil {
				statusLabel.SetText("System image scan failed.")
				resultText.SetText(fmt.Sprintf("Error: %v", err))
				return
			}

			health := "Healthy"
			if !result.Healthy {
				health = "Issues Found"
			}
			statusLabel.SetText("System image scan complete.")
			resultText.SetText(fmt.Sprintf("System Image Scan:\n  Status: %s\n  Issues found: %d",
				health, result.IssuesFound))
		}()
	})
	scanBtn.Importance = widget.HighImportance

	// System Image Repair
	repairBtn := widget.NewButton("Repair System Image (DISM + SFC)", func() {
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Repairing system image (this may take a long time)...")

		go func() {
			result, err := tools.RepairSystemImage()
			progressBar.Stop()
			progressBar.Hide()

			if err != nil {
				statusLabel.SetText("System image repair failed.")
				resultText.SetText(fmt.Sprintf("Error: %v", err))
				return
			}

			statusLabel.SetText("System image repair complete.")
			resultText.SetText(fmt.Sprintf("System Image Repair:\n  Issues repaired: %d\n  Reboot required: %v",
				result.IssuesRepaired, result.RebootRequired))
		}()
	})
	repairBtn.Importance = widget.WarningImportance

	// Debloat Windows
	debloatBtn := widget.NewButton("Remove Windows Bloatware", func() {
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Removing bloatware apps...")

		go func() {
			result, err := tools.DebloatWindows()
			progressBar.Stop()
			progressBar.Hide()

			if err != nil {
				statusLabel.SetText("Debloat failed.")
				resultText.SetText(fmt.Sprintf("Error: %v", err))
				return
			}

			statusLabel.SetText("Debloat complete.")
			text := fmt.Sprintf("Windows Debloat:\n  Apps removed: %d\n", result.AppsRemoved)
			if len(result.AppsFailed) > 0 {
				text += fmt.Sprintf("  Apps failed: %d\n", len(result.AppsFailed))
				for _, app := range result.AppsFailed {
					text += fmt.Sprintf("    - %s\n", app)
				}
			}
			resultText.SetText(text)
		}()
	})
	debloatBtn.Importance = widget.DangerImportance

	// Disable Telemetry
	telemetryBtn := widget.NewButton("Disable Telemetry & Tracking", func() {
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Disabling Windows telemetry...")

		go func() {
			result, err := tools.DisableTelemetry()
			progressBar.Stop()
			progressBar.Hide()

			if err != nil {
				statusLabel.SetText("Telemetry disabling failed.")
				resultText.SetText(fmt.Sprintf("Error: %v", err))
				return
			}

			statusLabel.SetText("Telemetry disabled.")
			resultText.SetText(fmt.Sprintf("Telemetry Disabled:\n  Services disabled: %d\n  Tasks disabled: %d\n  Registry entries changed: %d",
				result.ServicesDisabled, result.TasksDisabled, result.RegistryChanged))
		}()
	})
	telemetryBtn.Importance = widget.DangerImportance

	// Bloatware list preview
	listBtn := widget.NewButton("Preview Bloatware List", func() {
		apps := tools.GetBloatwareList()
		text := "Apps that will be removed:\n\n"
		for i, app := range apps {
			text += fmt.Sprintf("  %d. %s\n", i+1, app)
		}
		resultText.SetText(text)
		statusLabel.SetText(fmt.Sprintf("Found %d bloatware apps.", len(apps)))
	})

	// Layout sections
	registrySection := container.NewVBox(
		widget.NewLabelWithStyle("Registry Tools", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Deep clean registry with automatic backup."),
		registryBtn,
	)

	imageSection := container.NewVBox(
		widget.NewLabelWithStyle("System Image Tools", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Scan and repair Windows system image using DISM and SFC."),
		container.NewGridWithColumns(2, scanBtn, repairBtn),
	)

	debloatSection := container.NewVBox(
		widget.NewLabelWithStyle("Windows Debloat", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Remove pre-installed bloatware and unnecessary apps."),
		container.NewGridWithColumns(2, listBtn, debloatBtn),
	)

	telemetrySection := container.NewVBox(
		widget.NewLabelWithStyle("Privacy & Telemetry", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Disable Windows telemetry, tracking services, and scheduled tasks."),
		telemetryBtn,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("Advanced System Tools", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		registrySection,
		widget.NewSeparator(),
		imageSection,
		widget.NewSeparator(),
		debloatSection,
		widget.NewSeparator(),
		telemetrySection,
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		resultText,
	)

	return container.NewScroll(container.NewPadded(content))
}
