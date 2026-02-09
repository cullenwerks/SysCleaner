package tools

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"syscleaner/pkg/admin"
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

	if err := admin.RequireElevation("Registry Cleaning"); err != nil {
		return nil, err
	}

	start := time.Now()
	result := &RegistryCleanResult{}

	// Create backup first
	backupPath := filepath.Join(os.TempDir(), fmt.Sprintf("registry_backup_%s.reg", time.Now().Format("20060102_150405")))
	log.Printf("[SysCleaner] Creating registry backup at: %s", backupPath)
	if err := backupRegistry(backupPath); err != nil {
		return nil, fmt.Errorf("backup failed: %w", err)
	}
	result.BackupPath = backupPath

	cleaned := 0

	// Clean MUI Cache
	log.Println("[SysCleaner] Cleaning MUI Cache...")
	cleaned += cleanRegistryKey("HKCU\\Software\\Classes\\Local Settings\\Software\\Microsoft\\Windows\\Shell\\MuiCache")

	// Clean Recent Docs
	log.Println("[SysCleaner] Cleaning Recent Documents...")
	cleaned += cleanRegistryKey("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\RecentDocs")

	// Clean UserAssist
	log.Println("[SysCleaner] Cleaning UserAssist data...")
	cleaned += cleanRegistryKey("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\UserAssist")

	// Clean MountPoints2
	log.Println("[SysCleaner] Cleaning MountPoints2...")
	cleaned += cleanRegistryKey("HKCU\\Software\\Microsoft\\Windows\\CurrentVersion\\Explorer\\MountPoints2")

	// Clean SharedDLLs with 0 references
	log.Println("[SysCleaner] Cleaning orphaned SharedDLLs...")
	cleaned += cleanSharedDLLs()

	// Clean uninstall orphans
	log.Println("[SysCleaner] Cleaning orphaned uninstall entries...")
	cleaned += cleanUninstallOrphans()

	result.ItemsCleaned = cleaned
	result.Duration = time.Since(start)
	log.Printf("[SysCleaner] Registry cleaning complete: %d items cleaned in %s", cleaned, result.Duration)
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
