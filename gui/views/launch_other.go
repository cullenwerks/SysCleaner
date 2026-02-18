//go:build !windows && gui

package views

import (
	"fmt"
	"os/exec"
)

func launchExeNative(exePath string) error {
	cmd := exec.Command(exePath)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch %s: %w", exePath, err)
	}
	return nil
}
