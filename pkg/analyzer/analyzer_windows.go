//go:build windows

package analyzer

import (
	"golang.org/x/sys/windows/registry"
)

var unnecessaryStartup = map[string]bool{
	"OneDrive":         true,
	"Skype":            true,
	"Spotify":          true,
	"Discord":          true,
	"Steam":            true,
	"EpicGamesLauncher": true,
	"AdobeUpdater":     true,
	"iTunes":           true,
	"iTunesHelper":     true,
	"CCleaner":         true,
}

func analyzeStartupProgramsPlatform() []StartupProgramInfo {
	var programs []StartupProgramInfo

	// Read from HKLM
	programs = append(programs, readStartupKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Run`)...)

	// Read from HKCU
	programs = append(programs, readStartupKey(registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Run`)...)

	return programs
}

func readStartupKey(root registry.Key, path string) []StartupProgramInfo {
	var programs []StartupProgramInfo

	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE)
	if err != nil {
		return programs
	}
	defer key.Close()

	names, err := key.ReadValueNames(-1)
	if err != nil {
		return programs
	}

	for _, name := range names {
		impact := "Low"
		recommended := true

		if unnecessaryStartup[name] {
			impact = "High"
			recommended = false
		}

		programs = append(programs, StartupProgramInfo{
			Name:        name,
			Impact:      impact,
			Recommended: recommended,
		})
	}
	return programs
}
