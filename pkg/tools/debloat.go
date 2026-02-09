package tools

import (
	"fmt"
	"log"
	"os/exec"
	"runtime"

	"syscleaner/pkg/admin"
)

var bloatwareApps = []string{
	"Microsoft.BingWeather",
	"Microsoft.GetHelp",
	"Microsoft.Getstarted",
	"Microsoft.Microsoft3DViewer",
	"Microsoft.MicrosoftOfficeHub",
	"Microsoft.MicrosoftSolitaireCollection",
	"Microsoft.MixedReality.Portal",
	"Microsoft.Office.OneNote",
	"Microsoft.People",
	"Microsoft.SkypeApp",
	"Microsoft.Wallet",
	"Microsoft.WindowsFeedbackHub",
	"Microsoft.Xbox.TCUI",
	"Microsoft.XboxApp",
	"Microsoft.XboxGameOverlay",
	"Microsoft.XboxGamingOverlay",
	"Microsoft.XboxIdentityProvider",
	"Microsoft.XboxSpeechToTextOverlay",
	"Microsoft.YourPhone",
	"Microsoft.ZuneMusic",
	"Microsoft.ZuneVideo",
}

// DebloatResult holds the results of a debloat operation.
type DebloatResult struct {
	AppsRemoved int
	AppsFailed  []string
}

// GetBloatwareList returns the list of apps that will be removed.
func GetBloatwareList() []string {
	return bloatwareApps
}

// DebloatWindows removes bloatware apps.
func DebloatWindows() (*DebloatResult, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("debloating only available on Windows")
	}

	if err := admin.RequireElevation("Windows Debloat"); err != nil {
		return nil, err
	}

	result := &DebloatResult{}

	log.Println("[SysCleaner] Starting Windows debloat...")
	for _, app := range bloatwareApps {
		log.Printf("[SysCleaner] Removing: %s", app)
		cmd := exec.Command("powershell", "-Command",
			fmt.Sprintf("Get-AppxPackage *%s* | Remove-AppxPackage", app))

		if err := cmd.Run(); err != nil {
			result.AppsFailed = append(result.AppsFailed, app)
		} else {
			result.AppsRemoved++
		}
	}

	log.Printf("[SysCleaner] Debloat complete: %d apps removed, %d failed", result.AppsRemoved, len(result.AppsFailed))
	return result, nil
}
