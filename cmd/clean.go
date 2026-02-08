package cmd

import (
	"fmt"

	"syscleaner/pkg/cleaner"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean system junk files and free disk space",
	Long:  `Remove temporary files, browser caches, log files, prefetch data, thumbnails, and registry junk.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := cleaner.CleanOptions{DryRun: dryRun}

		if all {
			opts.TempFiles = true
			opts.Browser = true
			opts.Registry = true
			opts.Logs = true
			opts.Prefetch = true
			opts.Thumbnails = true
		} else {
			opts.TempFiles, _ = cmd.Flags().GetBool("temp")
			opts.Browser, _ = cmd.Flags().GetBool("browser")
			opts.Registry, _ = cmd.Flags().GetBool("registry")
			opts.Logs, _ = cmd.Flags().GetBool("logs")
			opts.Prefetch, _ = cmd.Flags().GetBool("prefetch")
			opts.Thumbnails, _ = cmd.Flags().GetBool("thumbnails")
		}

		if !opts.TempFiles && !opts.Browser && !opts.Registry &&
			!opts.Logs && !opts.Prefetch && !opts.Thumbnails {
			fmt.Println("No cleaning targets specified. Use --all or specify targets (--temp, --browser, etc.)")
			return
		}

		if dryRun {
			fmt.Println("[DRY RUN] Scanning files without deleting...")
			fmt.Println()
		}

		fmt.Println("Starting system cleanup...")
		fmt.Println()

		if opts.TempFiles {
			fmt.Println("  Cleaning temporary files...")
		}
		if opts.Browser {
			fmt.Println("  Cleaning browser caches...")
		}
		if opts.Prefetch {
			fmt.Println("  Cleaning prefetch data...")
		}
		if opts.Thumbnails {
			fmt.Println("  Cleaning thumbnail caches...")
		}
		if opts.Logs {
			fmt.Println("  Cleaning log files...")
		}
		if opts.Registry {
			fmt.Println("  Cleaning registry junk...")
		}

		fmt.Println()

		result := cleaner.PerformClean(opts)

		fmt.Println("--- Cleanup Summary ---")
		if dryRun {
			fmt.Println("  Mode:          DRY RUN (no files deleted)")
		}
		fmt.Printf("  Files found:   %d\n", result.FilesDeleted)
		fmt.Printf("  Space freed:   %s\n", cleaner.FormatBytes(result.SpaceFreed))
		fmt.Printf("  Time taken:    %s\n", result.Duration.Round(1e6))
		if len(result.Errors) > 0 {
			fmt.Printf("  Errors:        %d\n", len(result.Errors))
		} else {
			fmt.Println("  Errors:        0")
		}
		fmt.Println()
		if dryRun {
			fmt.Println("Run without --dry-run to actually delete files.")
		} else {
			fmt.Println("Cleanup complete!")
		}
	},
}

func init() {
	cleanCmd.Flags().Bool("all", false, "Enable all cleaning types")
	cleanCmd.Flags().Bool("temp", false, "Clean temporary files")
	cleanCmd.Flags().Bool("browser", false, "Clean browser caches")
	cleanCmd.Flags().Bool("registry", false, "Clean registry junk")
	cleanCmd.Flags().Bool("logs", false, "Clean log files")
	cleanCmd.Flags().Bool("prefetch", false, "Clean prefetch data")
	cleanCmd.Flags().Bool("thumbnails", false, "Clean thumbnail caches")
	cleanCmd.Flags().Bool("dry-run", false, "Show what would be cleaned without deleting")
	rootCmd.AddCommand(cleanCmd)
}
