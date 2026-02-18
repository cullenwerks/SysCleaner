//go:build windows

package optimizer

import (
	"fmt"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
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

func createDefragTaskCOM() error {
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		// RPC_E_CHANGED_MODE (0x80010106) means COM is already initialized
		// in a different mode — that's fine, proceed.
		if oleErr, ok := err.(*ole.OleError); !ok || oleErr.Code() != 0x80010106 {
			return fmt.Errorf("CoInitialize: %w", err)
		}
	}
	defer ole.CoUninitialize()

	taskServiceUnk, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return fmt.Errorf("create Schedule.Service: %w", err)
	}
	defer taskServiceUnk.Release()

	taskService, err := taskServiceUnk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("QueryInterface Schedule.Service: %w", err)
	}
	defer taskService.Release()

	if _, err := oleutil.CallMethod(taskService, "Connect"); err != nil {
		return fmt.Errorf("Connect: %w", err)
	}

	rootFolderRaw, err := oleutil.CallMethod(taskService, "GetFolder", `\`)
	if err != nil {
		return fmt.Errorf("GetFolder: %w", err)
	}
	rootFolder := rootFolderRaw.ToIDispatch()
	defer rootFolder.Release()

	taskDefRaw, err := oleutil.CallMethod(taskService, "NewTask", 0)
	if err != nil {
		return fmt.Errorf("NewTask: %w", err)
	}
	taskDef := taskDefRaw.ToIDispatch()
	defer taskDef.Release()

	regInfoRaw, err := oleutil.GetProperty(taskDef, "RegistrationInfo")
	if err != nil {
		return fmt.Errorf("get RegistrationInfo: %w", err)
	}
	regInfo := regInfoRaw.ToIDispatch()
	defer regInfo.Release()
	oleutil.PutProperty(regInfo, "Description", "SysCleaner weekly HDD defragmentation")
	oleutil.PutProperty(regInfo, "Author", "SysCleaner")

	settingsRaw, err := oleutil.GetProperty(taskDef, "Settings")
	if err != nil {
		return fmt.Errorf("get Settings: %w", err)
	}
	settings := settingsRaw.ToIDispatch()
	defer settings.Release()
	oleutil.PutProperty(settings, "StartWhenAvailable", true)
	oleutil.PutProperty(settings, "RunOnlyIfNetworkAvailable", false)

	triggersRaw, err := oleutil.GetProperty(taskDef, "Triggers")
	if err != nil {
		return fmt.Errorf("get Triggers: %w", err)
	}
	triggers := triggersRaw.ToIDispatch()
	defer triggers.Release()

	// TASK_TRIGGER_WEEKLY = 4
	triggerRaw, err := oleutil.CallMethod(triggers, "Create", 4)
	if err != nil {
		return fmt.Errorf("create weekly trigger: %w", err)
	}
	trigger := triggerRaw.ToIDispatch()
	defer trigger.Release()

	startTime := time.Now().Format("2006-01-02") + "T03:00:00"
	oleutil.PutProperty(trigger, "StartBoundary", startTime)
	oleutil.PutProperty(trigger, "DaysOfWeek", 1) // Sunday = 1
	oleutil.PutProperty(trigger, "WeeksInterval", 1)
	oleutil.PutProperty(trigger, "Enabled", true)

	actionsRaw, err := oleutil.GetProperty(taskDef, "Actions")
	if err != nil {
		return fmt.Errorf("get Actions: %w", err)
	}
	actions := actionsRaw.ToIDispatch()
	defer actions.Release()

	// TASK_ACTION_EXEC = 0
	actionRaw, err := oleutil.CallMethod(actions, "Create", 0)
	if err != nil {
		return fmt.Errorf("create action: %w", err)
	}
	action := actionRaw.ToIDispatch()
	defer action.Release()
	oleutil.PutProperty(action, "Path", `C:\Windows\System32\defrag.exe`)
	oleutil.PutProperty(action, "Arguments", "C: /O")

	// TASK_CREATE_OR_UPDATE = 6, TASK_LOGON_INTERACTIVE_TOKEN = 3
	_, err = oleutil.CallMethod(rootFolder, "RegisterTaskDefinition",
		"SysCleanerDefrag",
		taskDef,
		6,
		"",
		"",
		3,
		"",
	)
	if err != nil {
		return fmt.Errorf("RegisterTaskDefinition: %w", err)
	}

	return nil
}
