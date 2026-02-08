package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "syscleaner",
	Short: "SysCleaner - Windows System Cleaner & Gaming Optimizer",
	Long: `SysCleaner is a free, open-source Windows system optimizer.

Features:
  - Deep system cleaning (temp files, browser caches, registry, logs)
  - Gaming mode (auto-detects games, boosts CPU/RAM priority)
  - Background daemon (auto-enables gaming mode when games launch)
  - System analyzer (performance scoring and recommendations)
  - System optimizer (startup, network, registry, disk optimizations)`,
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
