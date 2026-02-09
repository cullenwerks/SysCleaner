package admin

import (
	"fmt"
	"os"
	"runtime"
)

// IsElevated checks whether the process is running with administrator privileges.
func IsElevated() bool {
	if runtime.GOOS != "windows" {
		return os.Geteuid() == 0
	}
	return isElevatedPlatform()
}

// RequireElevation returns an error if the process is not running as admin.
// Call this before any privileged operation so the user gets a clear message
// instead of silent failures that look suspicious to antivirus software.
func RequireElevation(operation string) error {
	if IsElevated() {
		return nil
	}
	return fmt.Errorf("%s requires administrator privileges.\nPlease right-click the executable and select \"Run as administrator\"", operation)
}
