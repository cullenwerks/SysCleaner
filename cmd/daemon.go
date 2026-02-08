package cmd

import (
	"fmt"
	"time"

	"syscleaner/pkg/daemon"

	"github.com/spf13/cobra"
)

var daemonCmd = &cobra.Command{
	Use:   "daemon",
	Short: "Manage the SysCleaner background service",
	Long:  `Install, start, stop, or check status of the SysCleaner background daemon that auto-enables gaming mode.`,
	Run: func(cmd *cobra.Command, args []string) {
		install, _ := cmd.Flags().GetBool("install")
		start, _ := cmd.Flags().GetBool("start")
		stop, _ := cmd.Flags().GetBool("stop")
		restart, _ := cmd.Flags().GetBool("restart")
		showStatus, _ := cmd.Flags().GetBool("status")
		uninstall, _ := cmd.Flags().GetBool("uninstall")

		switch {
		case install:
			fmt.Println("Installing SysCleaner service...")
			if err := daemon.Install(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				fmt.Println("  Make sure you are running as Administrator.")
				return
			}
			fmt.Println("  Service installed successfully.")
			fmt.Println("  Run 'syscleaner daemon --start' to start the service.")

		case uninstall:
			fmt.Println("Uninstalling SysCleaner service...")
			if err := daemon.Uninstall(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}
			fmt.Println("  Service uninstalled successfully.")

		case start:
			fmt.Println("Starting SysCleaner service...")
			if err := daemon.Start(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}
			fmt.Println("  Service started.")

		case stop:
			fmt.Println("Stopping SysCleaner service...")
			if err := daemon.Stop(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}
			fmt.Println("  Service stopped.")

		case restart:
			fmt.Println("Restarting SysCleaner service...")
			if err := daemon.Restart(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}
			fmt.Println("  Service restarted.")

		case showStatus:
			printDaemonStatus()

		default:
			// Run in foreground mode
			fmt.Println("Running SysCleaner daemon in foreground mode...")
			fmt.Println("Press Ctrl+C to stop.")
			fmt.Println()
			if err := daemon.RunForeground(); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		}
	},
}

func printDaemonStatus() {
	status := daemon.GetStatus()

	fmt.Println("--- SysCleaner Daemon Status ---")
	fmt.Println()

	if status.Running {
		fmt.Println("  Status:          Running")
		fmt.Printf("  PID:             %d\n", status.PID)
		fmt.Printf("  Uptime:          %s\n", status.Uptime.Round(time.Second))
	} else {
		fmt.Println("  Status:          Stopped")
	}

	fmt.Printf("  Auto-Gaming:     %v\n", status.AutoGamingEnabled)
	fmt.Printf("  Cleanup:         %s\n", status.CleanupSchedule)
	fmt.Printf("  CPU Threshold:   %.0f%%\n", status.CPUThreshold)
	fmt.Printf("  RAM Threshold:   %.0f%%\n", status.RAMThreshold)
	fmt.Println()

	if len(status.RecentActions) > 0 {
		fmt.Println("  Recent Actions:")
		for _, a := range status.RecentActions {
			fmt.Printf("    [%s] %s\n", a.Time, a.Description)
		}
	} else {
		fmt.Println("  Recent Actions:  None")
	}
}

func init() {
	daemonCmd.Flags().Bool("install", false, "Install as Windows service")
	daemonCmd.Flags().Bool("uninstall", false, "Uninstall the Windows service")
	daemonCmd.Flags().Bool("start", false, "Start the service")
	daemonCmd.Flags().Bool("stop", false, "Stop the service")
	daemonCmd.Flags().Bool("restart", false, "Restart the service")
	daemonCmd.Flags().Bool("status", false, "Show daemon status")
	daemonCmd.Flags().String("config", "", "Path to config file")
	rootCmd.AddCommand(daemonCmd)
}
