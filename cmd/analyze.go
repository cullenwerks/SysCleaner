package cmd

import (
	"syscleaner/pkg/analyzer"

	"github.com/spf13/cobra"
)

var analyzeCmd = &cobra.Command{
	Use:   "analyze",
	Short: "Analyze system performance and recommend optimizations",
	Long:  `Scan your system and generate a performance report with score, issues, and recommendations.`,
	Run: func(cmd *cobra.Command, args []string) {
		report := analyzer.AnalyzeSystem()
		analyzer.PrintReport(report)
	},
}

func init() {
	rootCmd.AddCommand(analyzeCmd)
}
