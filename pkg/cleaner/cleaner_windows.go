//go:build windows

package cleaner

import (
	"fmt"

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
