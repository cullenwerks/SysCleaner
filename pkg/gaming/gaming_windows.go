//go:build windows

package gaming

import (
	"fmt"
	"syscall"
	"unsafe"

	"golang.org/x/sys/windows"
	"golang.org/x/sys/windows/registry"
)

var (
	powrprof                 = windows.NewLazySystemDLL("powrprof.dll")
	procPowerSetActiveScheme = powrprof.NewProc("PowerSetActiveScheme")
)

// setPowerSchemeNative activates a Windows power scheme by GUID string.
// guidStr is in the form "xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx" (without braces).
func setPowerSchemeNative(guidStr string) error {
	guid, err := windows.GUIDFromString("{" + guidStr + "}")
	if err != nil {
		return fmt.Errorf("invalid power scheme GUID %q: %w", guidStr, err)
	}
	ret, _, err := procPowerSetActiveScheme.Call(
		0,
		uintptr(unsafe.Pointer(&guid)),
	)
	if ret != 0 {
		return fmt.Errorf("PowerSetActiveScheme failed (0x%x): %w", ret, err)
	}
	return nil
}

const tcpParamsKey = `SYSTEM\CurrentControlSet\Services\Tcpip\Parameters`

// setTCPGamingParams writes the three TCP registry values used by gaming mode.
func setTCPGamingParams() error {
	key, _, err := registry.CreateKey(registry.LOCAL_MACHINE, tcpParamsKey, registry.SET_VALUE)
	if err != nil {
		return fmt.Errorf("failed to open TCP params key: %w", err)
	}
	defer key.Close()

	params := map[string]uint32{
		"TcpAutoTuningLevel": 0,
		"EnableTCPChimney":   1,
		"EnableTCPDCA":       1,
	}
	for name, val := range params {
		if err := key.SetDWordValue(name, val); err != nil {
			return fmt.Errorf("failed to set %s: %w", name, err)
		}
	}
	return nil
}

// startExplorerNative launches explorer.exe detached from our process
// so no CMD window appears and it doesn't inherit our console.
func startExplorerNative() error {
	exePath, err := windows.UTF16PtrFromString(`C:\Windows\explorer.exe`)
	if err != nil {
		return err
	}

	var si syscall.StartupInfo
	var pi syscall.ProcessInformation
	si.Cb = uint32(unsafe.Sizeof(si))

	const detachedProcess = 0x00000008
	err = syscall.CreateProcess(
		exePath,
		nil,
		nil,
		nil,
		false,
		detachedProcess,
		nil,
		nil,
		&si,
		&pi,
	)
	if err != nil {
		return fmt.Errorf("failed to start explorer.exe: %w", err)
	}
	syscall.CloseHandle(pi.Thread)
	syscall.CloseHandle(pi.Process)
	return nil
}
