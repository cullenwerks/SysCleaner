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
