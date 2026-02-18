//go:build windows

package cleaner

import (
	"fmt"
	"unsafe"

	"golang.org/x/sys/windows"
)

var (
	dnsapi                    = windows.NewLazySystemDLL("dnsapi.dll")
	procDnsFlushResolverCache = dnsapi.NewProc("DnsFlushResolverCache")
)

func flushDNSCacheNative() error {
	ret, _, err := procDnsFlushResolverCache.Call()
	if ret == 0 {
		return fmt.Errorf("DnsFlushResolverCache failed: %w", err)
	}
	return nil
}

var (
	wevtapi         = windows.NewLazySystemDLL("wevtapi.dll")
	procEvtClearLog = wevtapi.NewProc("EvtClearLog")
)

// clearEventLogNative clears a Windows event log channel by name.
// channelPath is e.g. "System" or "Application".
func clearEventLogNative(channelPath string) error {
	channel, err := windows.UTF16PtrFromString(channelPath)
	if err != nil {
		return fmt.Errorf("invalid channel name %q: %w", channelPath, err)
	}
	// EvtClearLog(Session, ChannelPath, TargetFilePath, Flags)
	// Session=0 means local, TargetFilePath=0 means discard cleared events.
	ret, _, err := procEvtClearLog.Call(
		0,
		uintptr(unsafe.Pointer(channel)),
		0,
		0,
	)
	if ret == 0 {
		return fmt.Errorf("EvtClearLog(%s) failed: %w", channelPath, err)
	}
	return nil
}

var (
	shell32               = windows.NewLazySystemDLL("shell32.dll")
	procSHEmptyRecycleBin = shell32.NewProc("SHEmptyRecycleBinW")
)

const (
	sherbNoConfirmation = 0x00000001
	sherbNoProgressUI   = 0x00000002
	sherbNoSound        = 0x00000004
)

func emptyRecycleBinNative() error {
	// SHEmptyRecycleBinW(hwnd, pszRootPath, dwFlags)
	// hwnd=0, pszRootPath=0 clears all drives, no dialog/sound.
	ret, _, err := procSHEmptyRecycleBin.Call(
		0,
		0,
		uintptr(sherbNoConfirmation|sherbNoProgressUI|sherbNoSound),
	)
	// S_OK=0 and S_FALSE=1 (already empty) are both success
	if ret != 0 && ret != 1 {
		return fmt.Errorf("SHEmptyRecycleBin failed (HRESULT 0x%x): %w", ret, err)
	}
	return nil
}
