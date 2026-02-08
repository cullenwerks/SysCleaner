//go:build windows

package optimizer

import (
	"syscall"

	"golang.org/x/sys/windows/registry"
)

var unnecessaryStartup = []string{
	"OneDrive", "Skype", "Spotify", "Discord",
	"Steam", "EpicGamesLauncher", "AdobeUpdater",
	"iTunes", "iTunesHelper",
}

func getSysProcAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{HideWindow: true}
}

func optimizeStartupPlatform() StartupResult {
	result := StartupResult{}

	regPaths := []struct {
		root registry.Key
		path string
	}{
		{registry.LOCAL_MACHINE, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`},
		{registry.CURRENT_USER, `SOFTWARE\Microsoft\Windows\CurrentVersion\Run`},
	}

	for _, rp := range regPaths {
		key, err := registry.OpenKey(rp.root, rp.path, registry.QUERY_VALUE|registry.SET_VALUE)
		if err != nil {
			continue
		}

		names, err := key.ReadValueNames(-1)
		if err != nil {
			key.Close()
			continue
		}

		for _, name := range names {
			val, _, err := key.GetStringValue(name)
			if err != nil {
				continue
			}

			isUnnecessary := false
			for _, u := range unnecessaryStartup {
				if name == u {
					isUnnecessary = true
					break
				}
			}

			prog := StartupProgram{
				Name: name,
				Path: val,
			}

			if isUnnecessary {
				prog.Impact = "High"
				if err := key.DeleteValue(name); err == nil {
					prog.Disabled = true
					result.Disabled++
				}
			} else {
				prog.Impact = "Low"
			}
			result.Programs = append(result.Programs, prog)
		}
		key.Close()
	}

	return result
}

func optimizeRegistryPlatform() RegistryResult {
	result := RegistryResult{}

	// Clean MUI cache
	result.EntriesRemoved += cleanRegKey(registry.CURRENT_USER,
		`SOFTWARE\Classes\Local Settings\Software\Microsoft\Windows\Shell\MuiCache`)

	// Clean mount points
	result.EntriesRemoved += cleanRegKey(registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\MountPoints2`)

	// Clean recent docs
	result.EntriesRemoved += cleanRegKey(registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\RecentDocs`)

	// Clean UserAssist
	result.EntriesRemoved += cleanRegKey(registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\UserAssist`)

	// Clean SharedDLLs with 0 references
	result.EntriesRemoved += cleanSharedDLLs()

	return result
}

func cleanRegKey(root registry.Key, path string) int {
	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return 0
	}
	defer key.Close()

	names, err := key.ReadValueNames(-1)
	if err != nil {
		return 0
	}

	count := 0
	for _, name := range names {
		if err := key.DeleteValue(name); err == nil {
			count++
		}
	}
	return count
}

func cleanSharedDLLs() int {
	key, err := registry.OpenKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\SharedDLLs`,
		registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return 0
	}
	defer key.Close()

	names, err := key.ReadValueNames(-1)
	if err != nil {
		return 0
	}

	count := 0
	for _, name := range names {
		val, _, err := key.GetIntegerValue(name)
		if err != nil {
			continue
		}
		if val == 0 {
			if err := key.DeleteValue(name); err == nil {
				count++
			}
		}
	}
	return count
}

func setNetworkThrottling() error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE,
		`SOFTWARE\Microsoft\Windows NT\CurrentVersion\Multimedia\SystemProfile`,
		registry.SET_VALUE)
	if err != nil {
		return err
	}
	defer key.Close()
	return key.SetDWordValue("NetworkThrottlingIndex", 0xffffffff)
}
