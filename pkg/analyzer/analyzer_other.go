//go:build !windows

package analyzer

func analyzeStartupProgramsPlatform() []StartupProgramInfo {
	// Startup program analysis is only available on Windows
	return nil
}
