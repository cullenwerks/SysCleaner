//go:build gui

package gui

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"

	"syscleaner/gui/views"
)

// modernTheme implements a sleek dark theme with flame-orange accents.
type modernTheme struct{}

func (m modernTheme) Color(name fyne.ThemeColorName, variant fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNameBackground:
		return color.RGBA{R: 18, G: 18, B: 18, A: 255}
	case theme.ColorNameButton:
		return color.RGBA{R: 45, G: 45, B: 48, A: 255}
	case theme.ColorNamePrimary:
		return color.RGBA{R: 255, G: 85, B: 0, A: 255}
	case theme.ColorNameHover:
		return color.RGBA{R: 255, G: 110, B: 30, A: 255}
	case theme.ColorNameForeground:
		return color.RGBA{R: 230, G: 230, B: 230, A: 255}
	case theme.ColorNameDisabled:
		return color.RGBA{R: 100, G: 100, B: 100, A: 255}
	case theme.ColorNameInputBackground:
		return color.RGBA{R: 30, G: 30, B: 33, A: 255}
	case theme.ColorNameSeparator:
		return color.RGBA{R: 55, G: 55, B: 58, A: 255}
	default:
		return theme.DefaultTheme().Color(name, variant)
	}
}

func (m modernTheme) Font(style fyne.TextStyle) fyne.Resource {
	return theme.DefaultTheme().Font(style)
}

func (m modernTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return theme.DefaultTheme().Icon(name)
}

func (m modernTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 10
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 22
	case theme.SizeNameSubHeadingText:
		return 17
	default:
		return theme.DefaultTheme().Size(name)
	}
}

// Run launches the GUI application.
func Run() {
	a := app.NewWithID("com.syscleaner.app")
	a.Settings().SetTheme(&modernTheme{})

	w := a.NewWindow("SysCleaner - Ultimate Performance")
	w.Resize(fyne.NewSize(1200, 800))
	w.CenterOnScreen()
	w.SetMaster()

	mainContainer := createMainInterface(w)
	w.SetContent(mainContainer)
	w.ShowAndRun()
}

func createMainInterface(w fyne.Window) fyne.CanvasObject {
	tabs := container.NewAppTabs(
		container.NewTabItem("Dashboard", views.NewDashboard()),
		container.NewTabItem("Extreme Mode", views.NewExtremeModePanel(w)),
		container.NewTabItem("Clean", views.NewCleanPanel()),
		container.NewTabItem("Optimize", views.NewOptimizePanel()),
		container.NewTabItem("Tools", views.NewToolsPanel()),
		container.NewTabItem("Monitor", views.NewMonitorPanel()),
	)

	tabs.SetTabLocation(container.TabLocationLeading)

	return tabs
}
