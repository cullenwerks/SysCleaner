//go:build gui

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"syscleaner/pkg/cleaner"
)

// NewCleanPanel creates the cleaning interface with granular category options.
func NewCleanPanel() fyne.CanvasObject {
	statusLabel := widget.NewLabel("Ready to clean.")
	statusLabel.Wrapping = fyne.TextWrapWord

	resultText := widget.NewMultiLineEntry()
	resultText.SetPlaceHolder("Cleaning results will appear here...")
	resultText.Disable()
	resultText.SetMinRowsVisible(10)

	progressBar := widget.NewProgressBarInfinite()
	progressBar.Stop()
	progressBar.Hide()

	// System categories
	winTempCheck := widget.NewCheck("Windows Temp", nil)
	winTempCheck.SetChecked(true)
	userTempCheck := widget.NewCheck("User Temp", nil)
	userTempCheck.SetChecked(true)
	prefetchCheck := widget.NewCheck("Prefetch (30+ days)", nil)
	prefetchCheck.SetChecked(true)
	crashDumpCheck := widget.NewCheck("Crash Dumps", nil)
	crashDumpCheck.SetChecked(true)
	errorReportsCheck := widget.NewCheck("Error Reports", nil)
	errorReportsCheck.SetChecked(true)
	thumbCacheCheck := widget.NewCheck("Thumbnail Cache", nil)
	thumbCacheCheck.SetChecked(true)
	iconCacheCheck := widget.NewCheck("Icon Cache", nil)
	shaderCacheCheck := widget.NewCheck("Shader Cache", nil)
	shaderCacheCheck.SetChecked(true)
	dnsCacheCheck := widget.NewCheck("DNS Cache", nil)
	winLogsCheck := widget.NewCheck("Windows Logs", nil)
	winLogsCheck.SetChecked(true)
	eventLogsCheck := widget.NewCheck("Event Logs", nil)
	deliveryOptCheck := widget.NewCheck("Delivery Optimization", nil)
	recycleBinCheck := widget.NewCheck("Recycle Bin", nil)
	winUpdateCheck := widget.NewCheck("Windows Update Cache", nil)
	winInstallerCheck := widget.NewCheck("Windows Installer Cache", nil)
	fontCacheCheck := widget.NewCheck("Font Cache", nil)

	// Browser categories
	chromeCheck := widget.NewCheck("Chrome", nil)
	chromeCheck.SetChecked(true)
	firefoxCheck := widget.NewCheck("Firefox", nil)
	firefoxCheck.SetChecked(true)
	edgeCheck := widget.NewCheck("Edge", nil)
	edgeCheck.SetChecked(true)
	braveCheck := widget.NewCheck("Brave", nil)
	braveCheck.SetChecked(true)
	operaCheck := widget.NewCheck("Opera", nil)
	operaCheck.SetChecked(true)

	// Application categories
	discordCheck := widget.NewCheck("Discord", nil)
	discordCheck.SetChecked(true)
	spotifyCheck := widget.NewCheck("Spotify", nil)
	spotifyCheck.SetChecked(true)
	steamCheck := widget.NewCheck("Steam", nil)
	steamCheck.SetChecked(true)
	teamsCheck := widget.NewCheck("Teams", nil)
	teamsCheck.SetChecked(true)
	vscodeCheck := widget.NewCheck("VS Code", nil)
	vscodeCheck.SetChecked(true)
	javaCheck := widget.NewCheck("Java", nil)
	javaCheck.SetChecked(true)

	systemChecks := []*widget.Check{
		winTempCheck, userTempCheck, prefetchCheck, crashDumpCheck,
		errorReportsCheck, thumbCacheCheck, iconCacheCheck, shaderCacheCheck,
		dnsCacheCheck, winLogsCheck, eventLogsCheck, deliveryOptCheck,
		recycleBinCheck, winUpdateCheck, winInstallerCheck, fontCacheCheck,
	}
	browserChecks := []*widget.Check{chromeCheck, firefoxCheck, edgeCheck, braveCheck, operaCheck}
	appChecks := []*widget.Check{discordCheck, spotifyCheck, steamCheck, teamsCheck, vscodeCheck, javaCheck}

	makeSelectAll := func(checks []*widget.Check, val bool) func() {
		return func() {
			for _, c := range checks {
				c.SetChecked(val)
			}
		}
	}

	// Build options from checkboxes
	buildOpts := func(dryRun bool) cleaner.CleanOptions {
		return cleaner.CleanOptions{
			WindowsTemp:          winTempCheck.Checked,
			UserTemp:             userTempCheck.Checked,
			Prefetch:             prefetchCheck.Checked,
			CrashDumps:           crashDumpCheck.Checked,
			ErrorReports:         errorReportsCheck.Checked,
			ThumbnailCache:       thumbCacheCheck.Checked,
			IconCache:            iconCacheCheck.Checked,
			ShaderCache:          shaderCacheCheck.Checked,
			DNSCache:             dnsCacheCheck.Checked,
			WindowsLogs:          winLogsCheck.Checked,
			EventLogs:            eventLogsCheck.Checked,
			DeliveryOptimization: deliveryOptCheck.Checked,
			RecycleBin:           recycleBinCheck.Checked,
			WindowsUpdate:        winUpdateCheck.Checked,
			WindowsInstaller:     winInstallerCheck.Checked,
			FontCache:            fontCacheCheck.Checked,
			ChromeCache:          chromeCheck.Checked,
			FirefoxCache:         firefoxCheck.Checked,
			EdgeCache:            edgeCheck.Checked,
			BraveCache:           braveCheck.Checked,
			OperaCache:           operaCheck.Checked,
			DiscordCache:         discordCheck.Checked,
			SpotifyCache:         spotifyCheck.Checked,
			SteamCache:           steamCheck.Checked,
			TeamsCache:           teamsCheck.Checked,
			VSCodeCache:          vscodeCheck.Checked,
			JavaCache:            javaCheck.Checked,
			DryRun:               dryRun,
		}
	}

	// Analyze button (preview / dry run)
	analyzeBtn := widget.NewButton("Analyze (Preview)", func() {
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Analyzing system for cleanable files...")

		go func() {
			opts := buildOpts(true)
			result := cleaner.PerformClean(opts)
			progressBar.Stop()
			progressBar.Hide()

			statusLabel.SetText("Analysis complete.")
			resultText.SetText(fmt.Sprintf(
				"Files found: %d\nSpace reclaimable: %s\nDuration: %s\n\nRun 'Clean Now' to remove these files.",
				result.FilesDeleted,
				cleaner.FormatBytes(result.SpaceFreed),
				result.Duration))
		}()
	})

	// Clean button
	cleanBtn := widget.NewButton("Clean Now", func() {
		progressBar.Show()
		progressBar.Start()
		statusLabel.SetText("Cleaning system...")

		go func() {
			opts := buildOpts(false)
			result := cleaner.PerformClean(opts)
			progressBar.Stop()
			progressBar.Hide()

			statusLabel.SetText("Cleaning complete!")
			text := fmt.Sprintf("Files removed: %d\nSpace freed: %s\nDuration: %s",
				result.FilesDeleted,
				cleaner.FormatBytes(result.SpaceFreed),
				result.Duration)
			if result.LockedFiles > 0 || result.PermissionFiles > 0 || len(result.Errors) > 0 {
				text += "\n"
				if result.LockedFiles > 0 {
					text += fmt.Sprintf("\nSkipped (in use): %d", result.LockedFiles)
				}
				if result.PermissionFiles > 0 {
					text += fmt.Sprintf("\nPermission errors: %d", result.PermissionFiles)
				}
				if len(result.Errors) > 0 {
					text += fmt.Sprintf("\nOther errors: %d", len(result.Errors))
				}
			}
			resultText.SetText(text)
		}()
	})
	cleanBtn.Importance = widget.HighImportance

	buttonRow := container.NewGridWithColumns(2, analyzeBtn, cleanBtn)

	// System section with select all/deselect all
	sysSelectAll := widget.NewButton("Select All", makeSelectAll(systemChecks, true))
	sysDeselectAll := widget.NewButton("Deselect All", makeSelectAll(systemChecks, false))
	systemHeader := container.NewHBox(
		widget.NewLabelWithStyle("System", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		sysSelectAll, sysDeselectAll,
	)
	systemGrid := container.NewGridWithColumns(4,
		winTempCheck, userTempCheck, prefetchCheck, crashDumpCheck,
		errorReportsCheck, thumbCacheCheck, iconCacheCheck, shaderCacheCheck,
		dnsCacheCheck, winLogsCheck, eventLogsCheck, deliveryOptCheck,
		recycleBinCheck, winUpdateCheck, winInstallerCheck, fontCacheCheck,
	)

	// Browser section
	browserSelectAll := widget.NewButton("Select All", makeSelectAll(browserChecks, true))
	browserDeselectAll := widget.NewButton("Deselect All", makeSelectAll(browserChecks, false))
	browserHeader := container.NewHBox(
		widget.NewLabelWithStyle("Browsers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		browserSelectAll, browserDeselectAll,
	)
	browserGrid := container.NewGridWithColumns(5,
		chromeCheck, firefoxCheck, edgeCheck, braveCheck, operaCheck,
	)

	// Apps section
	appSelectAll := widget.NewButton("Select All", makeSelectAll(appChecks, true))
	appDeselectAll := widget.NewButton("Deselect All", makeSelectAll(appChecks, false))
	appHeader := container.NewHBox(
		widget.NewLabelWithStyle("Applications", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		appSelectAll, appDeselectAll,
	)
	appGrid := container.NewGridWithColumns(3,
		discordCheck, spotifyCheck, steamCheck,
		teamsCheck, vscodeCheck, javaCheck,
	)

	content := container.NewVBox(
		widget.NewLabelWithStyle("System Cleaning", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		systemHeader,
		systemGrid,
		widget.NewSeparator(),
		browserHeader,
		browserGrid,
		widget.NewSeparator(),
		appHeader,
		appGrid,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		resultText,
	)

	return container.NewScroll(container.NewPadded(content))
}
