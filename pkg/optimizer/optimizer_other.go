//go:build !windows

package optimizer

func optimizeStartupPlatform() StartupResult { return StartupResult{} }
func setNetworkThrottling() error            { return nil }
func isSSDPresentNative() bool               { return false }
func enableTRIMNative() error                { return nil }
func createDefragTask() error                { return nil }
func setTCPOptimizationParam(name string, val uint32) error { return nil }

var tcpOptimizations []struct {
	desc string
	name string
	val  uint32
}
