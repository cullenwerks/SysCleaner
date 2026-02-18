//go:build windows

package scheduler

import (
	"fmt"
	"time"

	ole "github.com/go-ole/go-ole"
	"github.com/go-ole/go-ole/oleutil"
)

const taskName = "SysCleanerWeeklyClean"

func withTaskService(fn func(svc *ole.IDispatch) error) error {
	if err := ole.CoInitializeEx(0, ole.COINIT_MULTITHREADED); err != nil {
		if oleErr, ok := err.(*ole.OleError); !ok || oleErr.Code() != 0x80010106 {
			return fmt.Errorf("CoInitialize: %w", err)
		}
	}
	defer ole.CoUninitialize()

	unk, err := oleutil.CreateObject("Schedule.Service")
	if err != nil {
		return fmt.Errorf("create Schedule.Service: %w", err)
	}
	defer unk.Release()

	svc, err := unk.QueryInterface(ole.IID_IDispatch)
	if err != nil {
		return fmt.Errorf("QueryInterface: %w", err)
	}
	defer svc.Release()

	if _, err := oleutil.CallMethod(svc, "Connect"); err != nil {
		return fmt.Errorf("Connect: %w", err)
	}

	return fn(svc)
}

func createScheduledCleanNative(cfg ScheduleConfig, exePath string) error {
	return withTaskService(func(svc *ole.IDispatch) error {
		rootFolderRaw, err := oleutil.CallMethod(svc, "GetFolder", `\`)
		if err != nil {
			return fmt.Errorf("GetFolder: %w", err)
		}
		rootFolder := rootFolderRaw.ToIDispatch()
		defer rootFolder.Release()

		taskDefRaw, err := oleutil.CallMethod(svc, "NewTask", 0)
		if err != nil {
			return fmt.Errorf("NewTask: %w", err)
		}
		taskDef := taskDefRaw.ToIDispatch()
		defer taskDef.Release()

		regInfoRaw, _ := oleutil.GetProperty(taskDef, "RegistrationInfo")
		regInfo := regInfoRaw.ToIDispatch()
		defer regInfo.Release()
		oleutil.PutProperty(regInfo, "Description", "SysCleaner weekly scheduled clean")
		oleutil.PutProperty(regInfo, "Author", "SysCleaner")

		settingsRaw, _ := oleutil.GetProperty(taskDef, "Settings")
		settings := settingsRaw.ToIDispatch()
		defer settings.Release()
		oleutil.PutProperty(settings, "StartWhenAvailable", true)

		triggersRaw, _ := oleutil.GetProperty(taskDef, "Triggers")
		triggers := triggersRaw.ToIDispatch()
		defer triggers.Release()

		// TASK_TRIGGER_WEEKLY = 4
		triggerRaw, err := oleutil.CallMethod(triggers, "Create", 4)
		if err != nil {
			return fmt.Errorf("create trigger: %w", err)
		}
		trigger := triggerRaw.ToIDispatch()
		defer trigger.Release()

		startTime := time.Now().Format("2006-01-02") + fmt.Sprintf("T%02d:00:00", cfg.Hour)
		oleutil.PutProperty(trigger, "StartBoundary", startTime)
		oleutil.PutProperty(trigger, "DaysOfWeek", dayOfWeekBit(cfg.DayOfWeek))
		oleutil.PutProperty(trigger, "WeeksInterval", 1)
		oleutil.PutProperty(trigger, "Enabled", true)

		actionsRaw, _ := oleutil.GetProperty(taskDef, "Actions")
		actions := actionsRaw.ToIDispatch()
		defer actions.Release()

		// TASK_ACTION_EXEC = 0
		actionRaw, err := oleutil.CallMethod(actions, "Create", 0)
		if err != nil {
			return fmt.Errorf("create action: %w", err)
		}
		action := actionRaw.ToIDispatch()
		defer action.Release()
		oleutil.PutProperty(action, "Path", exePath)
		oleutil.PutProperty(action, "Arguments", fmt.Sprintf("--headless --clean --%s", cfg.CleanPreset))

		// TASK_CREATE_OR_UPDATE = 6, TASK_LOGON_INTERACTIVE_TOKEN = 3
		_, err = oleutil.CallMethod(rootFolder, "RegisterTaskDefinition",
			taskName, taskDef, 6, "", "", 3, "")
		return err
	})
}

func removeScheduledCleanNative() error {
	return withTaskService(func(svc *ole.IDispatch) error {
		rootFolderRaw, err := oleutil.CallMethod(svc, "GetFolder", `\`)
		if err != nil {
			return fmt.Errorf("GetFolder: %w", err)
		}
		rootFolder := rootFolderRaw.ToIDispatch()
		defer rootFolder.Release()

		_, err = oleutil.CallMethod(rootFolder, "DeleteTask", taskName, 0)
		return err
	})
}

func getScheduledCleanNative() (*ScheduleConfig, error) {
	var result *ScheduleConfig
	err := withTaskService(func(svc *ole.IDispatch) error {
		rootFolderRaw, err := oleutil.CallMethod(svc, "GetFolder", `\`)
		if err != nil {
			return fmt.Errorf("GetFolder: %w", err)
		}
		rootFolder := rootFolderRaw.ToIDispatch()
		defer rootFolder.Release()

		taskRaw, err := oleutil.CallMethod(rootFolder, "GetTask", taskName)
		if err != nil {
			// Task not found — return nil config, no error
			result = nil
			return nil
		}
		task := taskRaw.ToIDispatch()
		defer task.Release()

		defRaw, err := oleutil.GetProperty(task, "Definition")
		if err != nil {
			return fmt.Errorf("get Definition: %w", err)
		}
		def := defRaw.ToIDispatch()
		defer def.Release()

		cfg := &ScheduleConfig{Enabled: true}

		triggersRaw, _ := oleutil.GetProperty(def, "Triggers")
		triggers := triggersRaw.ToIDispatch()
		defer triggers.Release()

		countRaw, _ := oleutil.GetProperty(triggers, "Count")
		if countRaw.Val > 0 {
			triggerRaw, _ := oleutil.CallMethod(triggers, "Item", 1)
			trigger := triggerRaw.ToIDispatch()
			defer trigger.Release()

			startRaw, _ := oleutil.GetProperty(trigger, "StartBoundary")
			startStr := startRaw.ToString()
			if len(startStr) >= 13 {
				var h int
				fmt.Sscanf(startStr[11:13], "%d", &h)
				cfg.Hour = h
			}

			dowRaw, _ := oleutil.GetProperty(trigger, "DaysOfWeek")
			cfg.DayOfWeek = dayOfWeekName(int(dowRaw.Val))
		}

		actionsRaw, _ := oleutil.GetProperty(def, "Actions")
		actions := actionsRaw.ToIDispatch()
		defer actions.Release()

		actionCountRaw, _ := oleutil.GetProperty(actions, "Count")
		if actionCountRaw.Val > 0 {
			actionRaw, _ := oleutil.CallMethod(actions, "Item", 1)
			action := actionRaw.ToIDispatch()
			defer action.Release()
			argsRaw, _ := oleutil.GetProperty(action, "Arguments")
			cfg.CleanPreset = parseCleanPreset(argsRaw.ToString())
		}

		result = cfg
		return nil
	})
	return result, err
}

// dayOfWeekBit maps a day name to the Windows Task Scheduler bitmask.
// Sunday=1, Monday=2, Tuesday=4, Wednesday=8, Thursday=16, Friday=32, Saturday=64
func dayOfWeekBit(day string) int {
	days := map[string]int{
		"Sunday": 1, "Monday": 2, "Tuesday": 4,
		"Wednesday": 8, "Thursday": 16, "Friday": 32, "Saturday": 64,
	}
	if v, ok := days[day]; ok {
		return v
	}
	return 1
}

func dayOfWeekName(bit int) string {
	names := map[int]string{
		1: "Sunday", 2: "Monday", 4: "Tuesday",
		8: "Wednesday", 16: "Thursday", 32: "Friday", 64: "Saturday",
	}
	if n, ok := names[bit]; ok {
		return n
	}
	return "Sunday"
}
