package cmd

import (
	"fmt"
	"strings"

	"syscleaner/pkg/priority"

	"github.com/spf13/cobra"
)

var priorityCmd = &cobra.Command{
	Use:   "priority",
	Short: "Manage permanent CPU priority settings for processes",
	Long: `Set permanent CPU, I/O, and memory page priority for executables via Windows registry.
These settings persist across reboots and take effect the next time the process starts.

Examples:
  syscleaner priority --list
  syscleaner priority --set LeagueClient.exe --cpu high --io high --page normal
  syscleaner priority --remove LeagueClient.exe`,
	Run: func(cmd *cobra.Command, args []string) {
		listFlag, _ := cmd.Flags().GetBool("list")
		setProcess, _ := cmd.Flags().GetString("set")
		removeProcess, _ := cmd.Flags().GetString("remove")
		cpuPriority, _ := cmd.Flags().GetString("cpu")
		ioPriority, _ := cmd.Flags().GetString("io")
		pagePriority, _ := cmd.Flags().GetString("page")

		// List configured priorities
		if listFlag {
			entries, err := priority.ListConfiguredPriorities()
			if err != nil {
				fmt.Printf("Error listing priorities: %v\n", err)
				return
			}

			if len(entries) == 0 {
				fmt.Println("No processes have configured priority settings.")
				return
			}

			fmt.Println("Configured Process Priorities:")
			fmt.Println(strings.Repeat("=", 80))
			fmt.Printf("%-30s %-15s %-15s %-15s\n", "Process Name", "CPU Priority", "I/O Priority", "Page Priority")
			fmt.Println(strings.Repeat("-", 80))
			for _, entry := range entries {
				fmt.Printf("%-30s %-15s %-15s %-15s\n",
					entry.ProcessName,
					entry.CpuPriorityName,
					entry.IoPriorityName,
					entry.PagePriorityName)
			}
			return
		}

		// Remove priority settings
		if removeProcess != "" {
			if err := priority.RemoveProcessPriority(removeProcess); err != nil {
				fmt.Printf("Error removing priority for %s: %v\n", removeProcess, err)
				return
			}
			fmt.Printf("Successfully removed priority settings for %s\n", removeProcess)
			fmt.Println("Changes will take effect the next time the process starts.")
			return
		}

		// Set priority settings
		if setProcess != "" {
			if cpuPriority == "" {
				cpuPriority = "normal"
			}
			if ioPriority == "" {
				ioPriority = "normal"
			}
			if pagePriority == "" {
				pagePriority = "normal"
			}

			cpuVal := priority.ParseCpuPriorityName(cpuPriority)
			ioVal := priority.ParseIoPriorityName(ioPriority)
			pageVal := priority.ParsePagePriorityName(pagePriority)

			if err := priority.SetProcessPriority(setProcess, cpuVal, ioVal, pageVal); err != nil {
				fmt.Printf("Error setting priority for %s: %v\n", setProcess, err)
				return
			}

			fmt.Printf("Successfully set priority for %s:\n", setProcess)
			fmt.Printf("  CPU Priority:  %s\n", priority.GetCpuPriorityName(cpuVal))
			fmt.Printf("  I/O Priority:  %s\n", priority.GetIoPriorityName(ioVal))
			fmt.Printf("  Page Priority: %s\n", priority.GetPagePriorityName(pageVal))
			fmt.Println("\nChanges will take effect the next time the process starts.")
			return
		}

		// No action specified
		fmt.Println("No action specified. Use --list, --set, or --remove.")
		fmt.Println("Run 'syscleaner priority --help' for usage information.")
	},
}

func init() {
	priorityCmd.Flags().Bool("list", false, "List all configured process priorities")
	priorityCmd.Flags().String("set", "", "Set priority for a process (e.g., LeagueClient.exe)")
	priorityCmd.Flags().String("remove", "", "Remove priority settings for a process")
	priorityCmd.Flags().String("cpu", "normal", "CPU priority: idle, below-normal, normal, above-normal, high")
	priorityCmd.Flags().String("io", "normal", "I/O priority: very-low, low, normal, high")
	priorityCmd.Flags().String("page", "normal", "Page priority: idle, very-low, low, background, default, normal")
	rootCmd.AddCommand(priorityCmd)
}
