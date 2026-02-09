package cmd

import (
	"fmt"

	"syscleaner/pkg/gaming"

	"github.com/spf13/cobra"
)

var extremeCmd = &cobra.Command{
	Use:   "extreme",
	Short: "Extreme performance mode - maximum system optimization",
	Long: `Extreme performance mode stops Windows Explorer and all non-essential services
for maximum gaming performance. Anti-cheat services are preserved.

WARNING: This mode removes the desktop shell. Use the GUI launcher to start games.`,
	Run: func(cmd *cobra.Command, args []string) {
		enable, _ := cmd.Flags().GetBool("enable")
		disable, _ := cmd.Flags().GetBool("disable")
		showStatus, _ := cmd.Flags().GetBool("status")

		if enable {
			fmt.Println("Enabling extreme performance mode...")
			fmt.Println()
			fmt.Println("WARNING: This will stop Windows Explorer (no desktop/taskbar).")
			fmt.Println("Use 'syscleaner extreme --disable' to restore.")
			fmt.Println()

			if err := gaming.EnableExtremeMode(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}

			fmt.Println("  Stopped Windows Explorer")
			fmt.Println("  Stopped all non-essential services")
			fmt.Println("  Enabled anti-cheat services")
			fmt.Println("  Set ultimate performance power plan")
			fmt.Println("  Disabled visual effects")
			fmt.Println()
			fmt.Println("EXTREME PERFORMANCE MODE is now ACTIVE")
		} else if disable {
			fmt.Println("Disabling extreme performance mode...")
			fmt.Println()

			if err := gaming.DisableExtremeMode(); err != nil {
				fmt.Printf("  Error: %v\n", err)
				return
			}

			fmt.Println("  Restored Windows Explorer")
			fmt.Println("  Restored services")
			fmt.Println("  Re-enabled visual effects")
			fmt.Println("  Restored balanced power plan")
			fmt.Println()
			fmt.Println("System restored to normal mode.")
		} else if showStatus {
			printExtremeStatus()
		} else {
			printExtremeStatus()
		}
	},
}

func printExtremeStatus() {
	fmt.Println("--- Extreme Performance Mode Status ---")
	fmt.Println()

	if gaming.IsExtremeModeActive() {
		fmt.Println("  Status:  ACTIVE")
		fmt.Println()
		fmt.Println("  Windows Explorer: STOPPED")
		fmt.Println("  Visual Effects:   DISABLED")
		fmt.Println("  Power Plan:       Ultimate Performance")
	} else {
		fmt.Println("  Status:  INACTIVE")
	}
	fmt.Println()

	if gaming.IsEnabled() {
		fmt.Println("  Gaming Mode:      ACTIVE")
	} else {
		fmt.Println("  Gaming Mode:      INACTIVE")
	}
}

func init() {
	extremeCmd.Flags().Bool("enable", false, "Enable extreme performance mode")
	extremeCmd.Flags().Bool("disable", false, "Disable extreme performance mode")
	extremeCmd.Flags().Bool("status", false, "Show extreme mode status")
	rootCmd.AddCommand(extremeCmd)
}
