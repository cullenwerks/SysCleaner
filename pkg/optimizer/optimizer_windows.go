//go:build windows

package optimizer

import (
	"fmt"

	"github.com/yusufpapurcu/wmi"
	"golang.org/x/sys/windows/registry"
)

var unnecessaryStartup = []string{
	"OneDrive", "Skype", "Spotify", "Discord",
	"Steam", "EpicGamesLauncher", "AdobeUpdater",
	"iTunes", "iTunesHelper",
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

const tcpParamsKey = `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`

var tcpOptimizations = []struct {
	desc string
	name string
	val  uint32
}{
	{"Set TCP auto-tuning to normal", "TcpAutoTuningLevel", 0},
	{"Enable TCP chimney offload", "EnableTCPChimney", 1},
	{"Enable direct cache access", "EnableTCPDCA", 1},
	{"Enable NetDMA", "EnableTCPDMA", 1},
	{"Enable receive-side scaling", "EnableRSS", 1},
	{"Disable TCP heuristics", "TcpHeuristics", 0},
}

func setTCPOptimizationParam(name string, val uint32) error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, tcpParamsKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open TCP params key: %w", err)
	}
	defer key.Close()
	return key.SetDWordValue(name, val)
}

type win32DiskDrive struct {
	MediaType string
}

func isSSDPresentNative() bool {
	var drives []win32DiskDrive
	if err := wmi.Query("SELECT MediaType FROM Win32_DiskDrive", &drives); err != nil {
		return false
	}
	for _, d := range drives {
		if d.MediaType == "SSD" || d.MediaType == "Solid State Drive" {
			return true
		}
	}
	return false
}

func enableTRIMNative() error {
	key, _, err := registry.CreateKey(
		registry.LOCAL_MACHINE,
		`SYSTEM\CurrentControlSet\Control\FileSystem`,
		registry.SET_VALUE,
	)
	if err != nil {
		return fmt.Errorf("failed to open FileSystem key: %w", err)
	}
	defer key.Close()
	return key.SetDWordValue("DisableDeleteNotify", 0)
}

func createDefragTask() error {
	return createDefragTaskCOM()
}

// createDefragTaskCOM is implemented in full in the next task.
// Stub here to keep the build green.
func createDefragTaskCOM() error {
	return nil
}
