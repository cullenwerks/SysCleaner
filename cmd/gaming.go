package cmd

import (
	"fmt"

	"syscleaner/pkg/cleaner"
	"syscleaner/pkg/gaming"

	"github.com/spf13/cobra"
)

var gamingCmd = &cobra.Command{
	Use:   "gaming",
	Short: "Gaming mode - optimize system for gaming performance",
	Long:  `Enable gaming mode to stop background services, boost game process priority, and optimize network settings.`,
	Run: func(cmd *cobra.Command, args []string) {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		showStatus, _ := cmd.Flags().GetBool("status")
		autoDetect, _ := cmd.Flags().GetBool("auto-detect")
		cpuBoost, _ := cmd.Flags().GetInt("cpu-boost")
		ramReserve, _ := cmd.Flags().GetInt("ram-reserve")

		if enable {
			fmt.Println("Enabling gaming mode...")
			fmt.Println()
			config := gaming.Config{
				AutoDetectGames: autoDetect,
				CPUBoost:        cpuBoost,
				RAMReserveGB:    ramReserve,
			}
			if err := gaming.Enable(config); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}
			fmt.Println("  Stopped background services")
			fmt.Println("  Set high performance power plan")
			fmt.Println("  Optimized network settings")
			if autoDetect {
				fmt.Println("  Game auto-detection enabled")
			}
			fmt.Println()
			fmt.Println("Gaming mode is now ACTIVE")
		} else if disable {
			fmt.Println("Disabling gaming mode...")
			fmt.Println()
			if err := gaming.Disable(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}
			fmt.Println("  Restarted background services")
			fmt.Println("  Restored balanced power plan")
			fmt.Println("  Restored process priorities")
			fmt.Println()
			fmt.Println("Gaming mode is now DISABLED")
		} else if showStatus {
			printGamingStatus()
		} else {
			printGamingStatus()
		}
	},
}

func printGamingStatus() {
	status := gaming.GetStatus()

	fmt.Println("--- Gaming Mode Status ---")
	fmt.Println()

	if status.Enabled {
		fmt.Println("  Status:     ENABLED")
	} else {
		fmt.Println("  Status:     DISABLED")
	}
	fmt.Println()

	fmt.Println("  System Resources:")
	fmt.Printf("    CPU Usage:  %.1f%%\n", status.CPUUsage)
	fmt.Printf("    RAM Usage:  %.1f%% (%s / %s)\n",
		status.RAMUsagePercent,
		cleaner.FormatBytes(int64(status.RAMUsed)),
		cleaner.FormatBytes(int64(status.RAMTotal)))
	fmt.Println()

	if len(status.ActiveGames) > 0 {
		fmt.Println("  Detected Games:")
		for _, g := range status.ActiveGames {
			fmt.Printf("    - %s (PID: %d, CPU: %.1f%%, RAM: %s)\n",
				g.Name, g.PID, g.CPUUsage,
				cleaner.FormatBytes(int64(g.RAMUsage)))
		}
	} else {
		fmt.Println("  Detected Games: None")
	}
	fmt.Println()

	if len(status.StoppedServices) > 0 {
		fmt.Printf("  Stopped Services: %d\n", len(status.StoppedServices))
		for _, s := range status.StoppedServices {
			fmt.Printf("    - %s\n", s)
		}
	}
}

func init() {
	gamingCmd.Flags().Bool("enable", false, "Enable gaming mode")
	gamingCmd.Flags().Bool("disable", false, "Disable gaming mode")
	gamingCmd.Flags().Bool("status", false, "Show gaming mode status")
	gamingCmd.Flags().Bool("auto-detect", true, "Auto-detect and boost game processes")
	gamingCmd.Flags().Int("cpu-boost", 80, "CPU boost percentage (0-100)")
	gamingCmd.Flags().Int("ram-reserve", 2, "GB of RAM to reserve for system")
	rootCmd.AddCommand(gamingCmd)
}
