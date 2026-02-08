//go:build windows

package cleaner

import (
	"golang.org/x/sys/windows/registry"
)

func cleanRegistryPlatform(dryRun bool) CleanResult {
	result := CleanResult{}

	// Clean MUI cache
	r := cleanRegistryKey(registry.CURRENT_USER,
		`SOFTWARE\Classes\Local Settings\Software\Microsoft\Windows\Shell\MuiCache`, dryRun)
	result.merge(r)

	// Clean recent docs
	r = cleanRegistryKey(registry.CURRENT_USER,
		`SOFTWARE\Microsoft\Windows\CurrentVersion\Explorer\RecentDocs`, dryRun)
	result.merge(r)

	return result
}

func cleanRegistryKey(root registry.Key, path string, dryRun bool) CleanResult {
	result := CleanResult{}

	key, err := registry.OpenKey(root, path, registry.QUERY_VALUE|registry.SET_VALUE)
	if err != nil {
		return result
	}
	defer key.Close()

	names, err := key.ReadValueNames(-1)
	if err != nil {
		result.Errors = append(result.Errors, err)
		return result
	}

	for _, name := range names {
		if dryRun {
			result.FilesDeleted++
		} else {
			if err := key.DeleteValue(name); err != nil {
				result.Errors = append(result.Errors, err)
			} else {
				result.FilesDeleted++
			}
		}
	}
	return result
}
