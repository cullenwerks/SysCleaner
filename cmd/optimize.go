package cmd

import (
	"fmt"

	"syscleaner/pkg/optimizer"

	"github.com/spf13/cobra"
)

var optimizeCmd = &cobra.Command{
	Use:   "optimize",
	Short: "Optimize system performance",
	Long:  `Optimize startup programs, network settings, registry, and disk performance.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		startup, _ := cmd.Flags().GetBool("startup")
		network, _ := cmd.Flags().GetBool("network")
		reg, _ := cmd.Flags().GetBool("registry")
		disk, _ := cmd.Flags().GetBool("disk")

		if all {
			startup, network, reg, disk = true, true, true, true
		}

		if !startup && !network && !reg && !disk {
			fmt.Println("No optimization targets specified. Use --all or specify targets (--startup, --network, etc.)")
			return
		}

		fmt.Println("Starting system optimization...")
		fmt.Println()

		if startup {
			fmt.Println("--- Startup Optimization ---")
			result := optimizer.OptimizeStartup()
			optimizer.PrintStartupResult(result)
			fmt.Println()
		}

		if network {
			fmt.Println("--- Network Optimization ---")
			result := optimizer.OptimizeNetwork()
			optimizer.PrintNetworkResult(result)
			fmt.Println()
		}

		if reg {
			fmt.Println("--- Registry Optimization ---")
			result := optimizer.OptimizeRegistry()
			optimizer.PrintRegistryResult(result)
			fmt.Println()
		}

		if disk {
			fmt.Println("--- Disk Optimization ---")
			result := optimizer.OptimizeDisk()
			optimizer.PrintDiskResult(result)
			fmt.Println()
		}

		fmt.Println("Optimization complete!")
	},
}

func init() {
	optimizeCmd.Flags().Bool("all", false, "Run all optimizations")
	optimizeCmd.Flags().Bool("startup", false, "Optimize startup programs")
	optimizeCmd.Flags().Bool("network", false, "Optimize network settings")
	optimizeCmd.Flags().Bool("registry", false, "Clean registry")
	optimizeCmd.Flags().Bool("disk", false, "Optimize disk performance")
	rootCmd.AddCommand(optimizeCmd)
}
