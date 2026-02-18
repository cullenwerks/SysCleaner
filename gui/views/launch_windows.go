//go:build windows && gui

package views

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	shell32            = windows.NewLazySystemDLL("shell32.dll")
	procShellExecuteEx = shell32.NewProc("ShellExecuteExW")
)

// shellExecuteInfo mirrors the SHELLEXECUTEINFOW Win32 struct.
type shellExecuteInfo struct {
	cbSize         uint32
	fMask          uint32
	hwnd           uintptr
	lpVerb         *uint16
	lpFile         *uint16
	lpParameters   *uint16
	lpDirectory    *uint16
	nShow          int32
	hInstApp       uintptr
	lpIDList       uintptr
	lpClass        *uint16
	hkeyClass      uintptr
	dwHotKey       uint32
	hIconOrMonitor uintptr
	hProcess       uintptr
}

const (
	seeMaskNoAsync = 0x00000100
	swShow         = 5
)

// launchExeNative launches an executable using ShellExecuteExW.
// No CMD window, no inherited console, correct UAC handling.
func launchExeNative(exePath string) error {
	verbPtr, err := windows.UTF16PtrFromString("open")
	if err != nil {
		return err
	}
	filePtr, err := windows.UTF16PtrFromString(exePath)
	if err != nil {
		return err
	}

	info := shellExecuteInfo{
		fMask:  seeMaskNoAsync,
		lpVerb: verbPtr,
		lpFile: filePtr,
		nShow:  swShow,
	}
	info.cbSize = uint32(unsafe.Sizeof(info))

	ret, _, err := procShellExecuteEx.Call(uintptr(unsafe.Pointer(&info)))
	if ret == 0 {
		return fmt.Errorf("ShellExecuteEx failed: %w", err)
	}
	return nil
}
