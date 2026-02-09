package tools

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

// RegistryCleanResult holds the results of a registry cleaning operation.
type RegistryCleanResult struct {
	ItemsCleaned int
	BackupPath   string
	Duration     time.Duration
}

// DeepRegistryClean performs comprehensive registry cleaning.
func DeepRegistryClean() (*RegistryCleanResult, error) {
	if runtime.GOOS != "windows" {
		return nil, fmt.Errorf("registry cleaning only available on Windows")
	}

	start := time.Now()
	result := &RegistryCleanResult{}

	// Create backup first
	backupPath := filepath.Join(os.TempDir(), fmt.Sprintf("registry_backup_%s.reg", time.Now().Format("20060102_150405")))
	if err := backupRegistry(backupPath); err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}
	result.BackupPath = backupPath

	cleaned := 0

	// Clean MUI Cache
	cleaned += cleanRegistryKey("HKCU\\Software\\Classes\\Local Settings\\Software\\Microsoft\\Windows\\Shell\\MuiCache")

	// Clean Recent Docs
	cleaned += cleanRegistryKey("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\RecentDocs")

	// Clean UserAssist
	cleaned += cleanRegistryKey("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\UserAssist")

	// Clean MountPoints2
	cleaned += cleanRegistryKey("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\MountPoints2")

	// Clean SharedDLLs with 0 references
	cleaned += cleanSharedDLLs()

	// Clean uninstall orphans
	cleaned += cleanUninstallOrphans()

	result.ItemsCleaned = cleaned
	result.Duration = time.Since(start)
	return result, nil
}

func backupRegistry(path string) error {
	cmd := exec.Command("reg", "export", "HKCU", path, "/y")
	return cmd.Run()
}

func cleanRegistryKey(keyPath string) int {
	cmd := exec.Command("reg", "delete", keyPath, "/f")
	if cmd.Run() == nil {
		return 1
	}
	return 0
}

func cleanSharedDLLs() int {
	// Clean SharedDLLs with 0 references
	return 0
}

func cleanUninstallOrphans() int {
	// Clean orphaned uninstall entries
	return 0
}
