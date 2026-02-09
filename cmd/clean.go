package cmd

import (
	"fmt"

	"syscleaner/pkg/cleaner"

	"github.com/spf13/cobra"
)

var cleanCmd = &cobra.Command{
	Use:   "clean",
	Short: "Clean system junk files and free disk space",
	Long: `Remove temporary files, browser caches, log files, prefetch data, and thumbnails.

You can select specific categories or use group flags like --all, --system, --browsers, --apps.`,
	Run: func(cmd *cobra.Command, args []string) {
		all, _ := cmd.Flags().GetBool("all")
		systemGroup, _ := cmd.Flags().GetBool("system")
		browsersGroup, _ := cmd.Flags().GetBool("browsers")
		appsGroup, _ := cmd.Flags().GetBool("apps")
		dryRun, _ := cmd.Flags().GetBool("dry-run")

		opts := cleaner.CleanOptions{DryRun: dryRun}

		// Group flags
		if all {
			systemGroup = true
			browsersGroup = true
			appsGroup = true
		}

		if systemGroup {
			opts.WindowsTemp = true
			opts.UserTemp = true
			opts.WindowsUpdate = true
			opts.WindowsInstaller = true
			opts.Prefetch = true
			opts.CrashDumps = true
			opts.ErrorReports = true
			opts.ThumbnailCache = true
			opts.IconCache = true
			opts.FontCache = true
			opts.ShaderCache = true
			opts.DNSCache = true
			opts.WindowsLogs = true
			opts.EventLogs = true
			opts.DeliveryOptimization = true
			opts.RecycleBin = true
		}

		if browsersGroup {
			opts.ChromeCache = true
			opts.FirefoxCache = true
			opts.EdgeCache = true
			opts.BraveCache = true
			opts.OperaCache = true
		}

		if appsGroup {
			opts.DiscordCache = true
			opts.SpotifyCache = true
			opts.SteamCache = true
			opts.TeamsCache = true
			opts.VSCodeCache = true
			opts.JavaCache = true
		}

		// Individual flags override groups
		if cmd.Flags().Changed("win-temp") {
			opts.WindowsTemp, _ = cmd.Flags().GetBool("win-temp")
		}
		if cmd.Flags().Changed("user-temp") {
			opts.UserTemp, _ = cmd.Flags().GetBool("user-temp")
		}
		if cmd.Flags().Changed("wupdate") {
			opts.WindowsUpdate, _ = cmd.Flags().GetBool("wupdate")
		}
		if cmd.Flags().Changed("installer") {
			opts.WindowsInstaller, _ = cmd.Flags().GetBool("installer")
		}
		if cmd.Flags().Changed("prefetch") {
			opts.Prefetch, _ = cmd.Flags().GetBool("prefetch")
		}
		if cmd.Flags().Changed("crashdumps") {
			opts.CrashDumps, _ = cmd.Flags().GetBool("crashdumps")
		}
		if cmd.Flags().Changed("wer") {
			opts.ErrorReports, _ = cmd.Flags().GetBool("wer")
		}
		if cmd.Flags().Changed("thumbcache") {
			opts.ThumbnailCache, _ = cmd.Flags().GetBool("thumbcache")
		}
		if cmd.Flags().Changed("iconcache") {
			opts.IconCache, _ = cmd.Flags().GetBool("iconcache")
		}
		if cmd.Flags().Changed("fontcache") {
			opts.FontCache, _ = cmd.Flags().GetBool("fontcache")
		}
		if cmd.Flags().Changed("shadercache") {
			opts.ShaderCache, _ = cmd.Flags().GetBool("shadercache")
		}
		if cmd.Flags().Changed("dnscache") {
			opts.DNSCache, _ = cmd.Flags().GetBool("dnscache")
		}
		if cmd.Flags().Changed("winlogs") {
			opts.WindowsLogs, _ = cmd.Flags().GetBool("winlogs")
		}
		if cmd.Flags().Changed("eventlogs") {
			opts.EventLogs, _ = cmd.Flags().GetBool("eventlogs")
		}
		if cmd.Flags().Changed("deliveryopt") {
			opts.DeliveryOptimization, _ = cmd.Flags().GetBool("deliveryopt")
		}
		if cmd.Flags().Changed("recyclebin") {
			opts.RecycleBin, _ = cmd.Flags().GetBool("recyclebin")
		}
		if cmd.Flags().Changed("chrome") {
			opts.ChromeCache, _ = cmd.Flags().GetBool("chrome")
		}
		if cmd.Flags().Changed("firefox") {
			opts.FirefoxCache, _ = cmd.Flags().GetBool("firefox")
		}
		if cmd.Flags().Changed("edge") {
			opts.EdgeCache, _ = cmd.Flags().GetBool("edge")
		}
		if cmd.Flags().Changed("brave") {
			opts.BraveCache, _ = cmd.Flags().GetBool("brave")
		}
		if cmd.Flags().Changed("opera") {
			opts.OperaCache, _ = cmd.Flags().GetBool("opera")
		}
		if cmd.Flags().Changed("discord") {
			opts.DiscordCache, _ = cmd.Flags().GetBool("discord")
		}
		if cmd.Flags().Changed("spotify") {
			opts.SpotifyCache, _ = cmd.Flags().GetBool("spotify")
		}
		if cmd.Flags().Changed("steam") {
			opts.SteamCache, _ = cmd.Flags().GetBool("steam")
		}
		if cmd.Flags().Changed("teams") {
			opts.TeamsCache, _ = cmd.Flags().GetBool("teams")
		}
		if cmd.Flags().Changed("vscode") {
			opts.VSCodeCache, _ = cmd.Flags().GetBool("vscode")
		}
		if cmd.Flags().Changed("java") {
			opts.JavaCache, _ = cmd.Flags().GetBool("java")
		}

		// Check if any category is selected
		hasSelection := opts.WindowsTemp || opts.UserTemp || opts.WindowsUpdate ||
			opts.WindowsInstaller || opts.Prefetch || opts.CrashDumps ||
			opts.ErrorReports || opts.ThumbnailCache || opts.IconCache ||
			opts.FontCache || opts.ShaderCache || opts.DNSCache ||
			opts.WindowsLogs || opts.EventLogs || opts.DeliveryOptimization ||
			opts.RecycleBin || opts.ChromeCache || opts.FirefoxCache ||
			opts.EdgeCache || opts.BraveCache || opts.OperaCache ||
			opts.DiscordCache || opts.SpotifyCache || opts.SteamCache ||
			opts.TeamsCache || opts.VSCodeCache || opts.JavaCache

		if !hasSelection {
			fmt.Println("No cleaning targets specified.")
			fmt.Println("\nGroup flags:")
			fmt.Println("  --all         : Clean everything")
			fmt.Println("  --system      : All system categories")
			fmt.Println("  --browsers    : All browser categories")
			fmt.Println("  --apps        : All application categories")
			fmt.Println("\nRun 'syscleaner clean --help' for a full list of categories.")
			return
		}

		if dryRun {
			fmt.Println("[DRY RUN] Scanning files without deleting...")
			fmt.Println()
		}

		fmt.Println("Starting system cleanup...")
		fmt.Println()

		result := cleaner.PerformClean(opts)

		fmt.Println("=== Cleanup Summary ===")
		if dryRun {
			fmt.Println("  Mode:          DRY RUN (no files deleted)")
		}
		fmt.Printf("  Files deleted: %d\n", result.FilesDeleted)
		fmt.Printf("  Files skipped: %d\n", result.SkippedFiles)
		fmt.Printf("  Space freed:   %s\n", cleaner.FormatBytes(result.SpaceFreed))
		fmt.Printf("  Time taken:    %s\n", result.Duration.Round(1e6))
		if result.LockedFiles > 0 {
			fmt.Printf("  Skipped (in use): %d\n", result.LockedFiles)
		}
		if result.PermissionFiles > 0 {
			fmt.Printf("  Permission errors: %d\n", result.PermissionFiles)
		}
		if len(result.Errors) > 0 {
			fmt.Printf("  Other errors:  %d\n", len(result.Errors))
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
	// Group flags
	cleanCmd.Flags().Bool("all", false, "Clean everything")
	cleanCmd.Flags().Bool("system", false, "All system categories")
	cleanCmd.Flags().Bool("browsers", false, "All browser categories")
	cleanCmd.Flags().Bool("apps", false, "All application categories")

	// System category flags
	cleanCmd.Flags().Bool("win-temp", false, "Windows Temp directory")
	cleanCmd.Flags().Bool("user-temp", false, "User Temp directories")
	cleanCmd.Flags().Bool("wupdate", false, "Windows Update cache")
	cleanCmd.Flags().Bool("installer", false, "Windows Installer cache")
	cleanCmd.Flags().Bool("prefetch", false, "Prefetch data (files older than 30 days)")
	cleanCmd.Flags().Bool("crashdumps", false, "Crash dump files")
	cleanCmd.Flags().Bool("wer", false, "Windows Error Reports")
	cleanCmd.Flags().Bool("thumbcache", false, "Thumbnail cache")
	cleanCmd.Flags().Bool("iconcache", false, "Icon cache")
	cleanCmd.Flags().Bool("fontcache", false, "Font cache")
	cleanCmd.Flags().Bool("shadercache", false, "DirectX shader cache")
	cleanCmd.Flags().Bool("dnscache", false, "DNS cache (flush)")
	cleanCmd.Flags().Bool("winlogs", false, "Windows log files")
	cleanCmd.Flags().Bool("eventlogs", false, "Windows Event Logs")
	cleanCmd.Flags().Bool("deliveryopt", false, "Delivery Optimization cache")
	cleanCmd.Flags().Bool("recyclebin", false, "Recycle Bin")

	// Application category flags
	cleanCmd.Flags().Bool("chrome", false, "Chrome cache")
	cleanCmd.Flags().Bool("firefox", false, "Firefox cache")
	cleanCmd.Flags().Bool("edge", false, "Edge cache")
	cleanCmd.Flags().Bool("brave", false, "Brave cache")
	cleanCmd.Flags().Bool("opera", false, "Opera cache")
	cleanCmd.Flags().Bool("discord", false, "Discord cache")
	cleanCmd.Flags().Bool("spotify", false, "Spotify cache")
	cleanCmd.Flags().Bool("steam", false, "Steam cache")
	cleanCmd.Flags().Bool("teams", false, "Teams cache")
	cleanCmd.Flags().Bool("vscode", false, "VS Code cache")
	cleanCmd.Flags().Bool("java", false, "Java cache")

	// Execution options
	cleanCmd.Flags().Bool("dry-run", false, "Show what would be cleaned without deleting")

	rootCmd.AddCommand(cleanCmd)
}
