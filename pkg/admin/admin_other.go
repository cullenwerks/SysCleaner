//go:build !windows

package admin

func isElevatedPlatform() bool {
	// On non-Windows, handled by os.Geteuid() in admin.go
	return false
}
