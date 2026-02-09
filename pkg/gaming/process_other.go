//go:build !windows

package gaming

import "fmt"

func terminateProcessByName(name string) error {
	return fmt.Errorf("process termination not available on this platform")
}
