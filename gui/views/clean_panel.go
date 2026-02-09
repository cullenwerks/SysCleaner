//go:build gui

package views

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"syscleaner/pkg/cleaner"
)

// NewCleanPanel creates the cleaning interface.
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

	// Clean categories
	tempCheck := widget.NewCheck("Temporary Files", nil)
	tempCheck.SetChecked(true)
	browserCheck := widget.NewCheck("Browser Caches", nil)
	browserCheck.SetChecked(true)
	prefetchCheck := widget.NewCheck("Prefetch Cache", nil)
	prefetchCheck.SetChecked(true)
	thumbCheck := widget.NewCheck("Thumbnail Cache", nil)
	thumbCheck.SetChecked(true)
	logCheck := widget.NewCheck("Log Files", nil)
	logCheck.SetChecked(true)
	registryCheck := widget.NewCheck("Registry Junk", nil)
	registryCheck.SetChecked(true)

	// Build options from checkboxes
	buildOpts := func(dryRun bool) cleaner.CleanOptions {
		return cleaner.CleanOptions{
			TempFiles:  tempCheck.Checked,
			Browser:    browserCheck.Checked,
			Prefetch:   prefetchCheck.Checked,
			Thumbnails: thumbCheck.Checked,
			Logs:       logCheck.Checked,
			Registry:   registryCheck.Checked,
			DryRun:     dryRun,
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
			if len(result.Errors) > 0 {
				text += fmt.Sprintf("\n\nErrors: %d (some files could not be removed)", len(result.Errors))
			}
			resultText.SetText(text)
		}()
	})
	cleanBtn.Importance = widget.HighImportance

	buttonRow := container.NewGridWithColumns(2, analyzeBtn, cleanBtn)

	content := container.NewVBox(
		widget.NewLabelWithStyle("System Cleaning", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewSeparator(),
		widget.NewLabel("Select categories to clean:"),
		tempCheck,
		browserCheck,
		prefetchCheck,
		thumbCheck,
		logCheck,
		registryCheck,
		widget.NewSeparator(),
		buttonRow,
		widget.NewSeparator(),
		statusLabel,
		progressBar,
		resultText,
	)

	return container.NewScroll(container.NewPadded(content))
}
