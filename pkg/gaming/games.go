package gaming

import (
	"strings"
)

// GameProfile describes a known game along with its executables, preferred CPU
// priority, and any services or processes that should not be terminated while
// the game is running.
type GameProfile struct {
	Name              string
	Executables       []string
	CPUPriority       string
	PreserveServices  []string
	PreserveProcesses []string
	Notes             string
}

// PredefinedGames is the built-in list of supported game profiles.
var PredefinedGames = []GameProfile{
	{
		Name:              "League of Legends",
		Executables:       []string{"LeagueClient.exe", "League of Legends.exe"},
		CPUPriority:       "High",
		PreserveProcesses: []string{"Discord.exe"},
		Notes:             "Benefits most from RAM freeing",
	},
	{
		Name:             "Valorant",
		Executables:      []string{"VALORANT.exe", "VALORANT-Win64-Shipping.exe"},
		CPUPriority:      "High",
		PreserveServices: []string{"vgc", "vgk"},
		Notes:            "Vanguard anti-cheat is mandatory",
	},
	{
		Name:        "CS2",
		Executables: []string{"cs2.exe"},
		CPUPriority: "High",
		Notes:       "Benefits from I/O priority boost",
	},
	{
		Name:             "Fortnite",
		Executables:      []string{"FortniteClient-Win64-Shipping.exe"},
		CPUPriority:      "Above Normal",
		PreserveServices: []string{"EasyAntiCheat"},
		Notes:            "Don't over-boost or EAC complains",
	},
	{
		Name:             "Apex Legends",
		Executables:      []string{"r5apex.exe"},
		CPUPriority:      "High",
		PreserveServices: []string{"EasyAntiCheat"},
		Notes:            "Benefits from RAM freeing",
	},
}

// GetGameProfile returns a pointer to the GameProfile whose Name matches the
// given name (case-insensitive). It returns nil if no match is found.
func GetGameProfile(name string) *GameProfile {
	lower := strings.ToLower(name)
	for i := range PredefinedGames {
		if strings.ToLower(PredefinedGames[i].Name) == lower {
			return &PredefinedGames[i]
		}
	}
	return nil
}

// GetGameProfileByExe returns a pointer to the GameProfile that lists the
// given executable (case-insensitive) in its Executables slice. It returns nil
// if no match is found.
func GetGameProfileByExe(exe string) *GameProfile {
	lower := strings.ToLower(exe)
	for i := range PredefinedGames {
		for _, e := range PredefinedGames[i].Executables {
			if strings.ToLower(e) == lower {
				return &PredefinedGames[i]
			}
		}
	}
	return nil
}
