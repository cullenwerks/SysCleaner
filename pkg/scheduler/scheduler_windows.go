//go:build windows

package scheduler

import "syscall"

func getSysProcAttr() *syscall.SysProcAttr {
	// NOTE: Do NOT set HideWindow: true — it triggers AV heuristics
	// (Trojan:Win32/Bearfoos.B!ml) because hidden child processes are
	// a common malware pattern. The commands we run are legitimate
	// system administration tools and don't need hidden windows.
	return &syscall.SysProcAttr{}
}
