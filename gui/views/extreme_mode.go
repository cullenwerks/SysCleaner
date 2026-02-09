//go:build gui

package views

import (
	"fmt"
	"os/exec"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"

	"syscleaner/pkg/gaming"
)

type extremeModePanel struct {
	window      fyne.Window
	statusLabel *widget.Label
	toggleBtn   *widget.Button
	isActive    bool
}

// NewExtremeModePanel creates the extreme performance mode panel.
func NewExtremeModePanel(w fyne.Window) fyne.CanvasObject {
	panel := &extremeModePanel{
		window:      w,
		statusLabel: widget.NewLabelWithStyle("Extreme Mode: Inactive", fyne.TextAlignCenter, fyne.TextStyle{Bold: true}),
		isActive:    gaming.IsExtremeModeActive(),
	}

	panel.toggleBtn = widget.NewButton("ACTIVATE EXTREME PERFORMANCE MODE", func() {
		panel.toggleExtremeMode()
	})
	panel.toggleBtn.Importance = widget.HighImportance
	panel.updateUI()

	// Flame header
	flameLabel := widget.NewLabelWithStyle("EXTREME PERFORMANCE", fyne.TextAlignCenter, fyne.TextStyle{Bold: true})

	// Warning
	warningText := widget.NewLabel("WARNING: Extreme Mode stops Windows Explorer shell.\nOnly use when launching games from this panel.\nYour desktop and taskbar will be unavailable.")
	warningText.Wrapping = fyne.TextWrapWord
	warningText.Alignment = fyne.TextAlignCenter

	// Game launchers
	launcherSection := createGameLaunchers(w)

	// Status update goroutine
	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			active := gaming.IsExtremeModeActive()
			if active != panel.isActive {
				panel.isActive = active
				panel.updateUI()
			}
		}
	}()

	content := container.NewVBox(
		flameLabel,
		widget.NewSeparator(),
		container.NewCenter(panel.statusLabel),
		container.NewCenter(panel.toggleBtn),
		widget.NewSeparator(),
		warningText,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Game Launchers", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		launcherSection,
	)

	return container.NewScroll(container.NewPadded(content))
}

func (p *extremeModePanel) toggleExtremeMode() {
	if p.isActive {
		if err := gaming.DisableExtremeMode(); err != nil {
			dialog.ShowError(err, p.window)
			return
		}
		p.isActive = false
		dialog.ShowInformation("Extreme Mode Disabled", "System restored to normal mode.", p.window)
	} else {
		dialog.ShowConfirm(
			"Activate Extreme Performance Mode?",
			"This will:\n\n"+
				"  - Stop Windows Explorer (no desktop/taskbar)\n"+
				"  - Stop all non-essential services\n"+
				"  - Maximize game performance\n\n"+
				"You can only launch games from this window.\n\nContinue?",
			func(confirmed bool) {
				if confirmed {
					if err := gaming.EnableExtremeMode(); err != nil {
						dialog.ShowError(err, p.window)
						return
					}
					p.isActive = true
					dialog.ShowInformation("Extreme Mode Activated", "System optimized for maximum performance!", p.window)
					p.updateUI()
				}
			},
			p.window,
		)
		return // updateUI will be called in the callback
	}
	p.updateUI()
}

func (p *extremeModePanel) updateUI() {
	if p.isActive {
		p.statusLabel.SetText("EXTREME MODE ACTIVE")
		p.toggleBtn.SetText("Disable Extreme Mode & Restore System")
		p.toggleBtn.Importance = widget.DangerImportance
	} else {
		p.statusLabel.SetText("Extreme Mode: Inactive")
		p.toggleBtn.SetText("ACTIVATE EXTREME PERFORMANCE MODE")
		p.toggleBtn.Importance = widget.HighImportance
	}
}

func createGameLaunchers(w fyne.Window) fyne.CanvasObject {
	launchers := []struct {
		name string
		exe  string
	}{
		{"Steam", "C:\\Program Files (x86)\\Steam\\steam.exe"},
		{"Riot Client", "C:\\Riot Games\\Riot Client\\RiotClientServices.exe"},
		{"EA App", "C:\\Program Files\\Electronic Arts\\EA Desktop\\EA Desktop\\EADesktop.exe"},
		{"Epic Games", "C:\\Program Files (x86)\\Epic Games\\Launcher\\Portal\\Binaries\\Win64\\EpicGamesLauncher.exe"},
		{"Battle.net", "C:\\Program Files (x86)\\Battle.net\\Battle.net Launcher.exe"},
		{"GOG Galaxy", "C:\\Program Files (x86)\\GOG Galaxy\\GalaxyClient.exe"},
		{"Ubisoft Connect", "C:\\Program Files (x86)\\Ubisoft\\Ubisoft Game Launcher\\UbisoftConnect.exe"},
	}

	grid := container.NewGridWithColumns(2)

	for _, launcher := range launchers {
		name := launcher.name
		exe := launcher.exe

		btn := widget.NewButton(fmt.Sprintf("Launch %s", name), func() {
			cmd := exec.Command(exe)
			if err := cmd.Start(); err != nil {
				dialog.ShowError(fmt.Errorf("failed to launch %s: %v", name, err), w)
			} else {
				dialog.ShowInformation("Launched", fmt.Sprintf("%s started successfully!", name), w)
			}
		})
		btn.Importance = widget.HighImportance
		grid.Add(btn)
	}

	return grid
}
