//go:build gui

package views

import (
	"fmt"
	"os/exec"
	"strings"
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

	// Process whitelist section
	whitelistSection := createWhitelistSection()

	// Game profile selector
	gameProfileSection := createGameProfileSection()

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
		widget.NewLabelWithStyle("Process Whitelist", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("These processes will NOT be killed when Extreme Mode activates:"),
		whitelistSection,
		widget.NewSeparator(),
		widget.NewLabelWithStyle("Game Optimization Profile", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		gameProfileSection,
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
				"  - Close background apps (respecting whitelist)\n"+
				"  - Maximize game performance\n\n"+
				"You can only launch games from this window.\nContinue?",
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

func createWhitelistSection() fyne.CanvasObject {
	processesToKill := gaming.GetProcessesToKill()

	// Common processes users would want to protect
	commonWhitelist := map[string]bool{
		"Discord.exe":    true,
		"DiscordPTB.exe": true,
		"Spotify.exe":    true,
	}

	checks := make(map[string]*widget.Check)
	var grid []fyne.CanvasObject

	for _, proc := range processesToKill {
		p := proc
		check := widget.NewCheck(p, func(checked bool) {
			updateWhitelist(checks)
		})
		if commonWhitelist[p] {
			check.SetChecked(true)
		}
		checks[p] = check
		grid = append(grid, check)
	}

	// Initialize whitelist from defaults
	updateWhitelist(checks)

	return container.NewGridWithColumns(4, grid...)
}

func updateWhitelist(checks map[string]*widget.Check) {
	var whitelist []string
	for name, check := range checks {
		if check.Checked {
			whitelist = append(whitelist, name)
		}
	}
	gaming.ProcessWhitelist = whitelist
}

func createGameProfileSection() fyne.CanvasObject {
	profileLabel := widget.NewLabel("Select a game to auto-configure whitelist and service preservation:")
	infoLabel := widget.NewLabel("")
	infoLabel.Wrapping = fyne.TextWrapWord

	options := []string{"None (Manual)"}
	for _, g := range gaming.PredefinedGames {
		options = append(options, g.Name)
	}

	selector := widget.NewSelect(options, func(selected string) {
		profile := gaming.GetGameProfile(selected)
		if profile == nil {
			infoLabel.SetText("Manual configuration - set process whitelist above.")
			return
		}

		info := fmt.Sprintf("Game: %s\nCPU Priority: %s", profile.Name, profile.CPUPriority)
		if len(profile.PreserveProcesses) > 0 {
			info += fmt.Sprintf("\nAuto-whitelisted: %s", strings.Join(profile.PreserveProcesses, ", "))
		}
		if len(profile.PreserveServices) > 0 {
			info += fmt.Sprintf("\nPreserved services: %s", strings.Join(profile.PreserveServices, ", "))
		}
		if profile.Notes != "" {
			info += fmt.Sprintf("\nNotes: %s", profile.Notes)
		}
		infoLabel.SetText(info)

		// Add game's preserved processes to whitelist
		existing := make(map[string]bool)
		for _, p := range gaming.ProcessWhitelist {
			existing[p] = true
		}
		for _, p := range profile.PreserveProcesses {
			if !existing[p] {
				gaming.ProcessWhitelist = append(gaming.ProcessWhitelist, p)
			}
		}
	})
	selector.SetSelected("None (Manual)")

	return container.NewVBox(profileLabel, selector, infoLabel)
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
