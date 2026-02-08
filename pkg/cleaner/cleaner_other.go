//go:build !windows

package cleaner

func cleanRegistryPlatform(dryRun bool) CleanResult {
	// Registry cleaning is only available on Windows
	return CleanResult{}
}
