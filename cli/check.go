package cli

import (
	"fmt"

	"jobradar/internal/config"
	"jobradar/internal/engine"
	"jobradar/internal/model"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var checkCmd = &cobra.Command{
	Use:   "check",
	Short: "Check for new jobs and send notifications",
	Long:  `Immediately check for new jobs matching your criteria and send notifications.`,
	RunE:  runCheck,
}

func init() {
	rootCmd.AddCommand(checkCmd)
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Print header
	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("\nJobRadar v1.0.0")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	// Create engine
	eng, err := engine.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	// Execute check
	fmt.Println("🔍 Checking for new jobs...")
	fmt.Println()

	stats, err := eng.Run()
	if err != nil {
		return fmt.Errorf("check failed: %w", err)
	}

	// Print results
	printStats(stats)

	return nil
}

func printStats(stats *model.RunStats) {
	fmt.Println()
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	green := color.New(color.FgGreen)
	green.Printf("✅ Check completed in %.1fs\n", stats.DurationSeconds)

	fmt.Println()
	fmt.Println("📊 Summary:")
	fmt.Printf("   • Fetched: %d\n", stats.JobsFetched)
	fmt.Printf("   • Matched: %d\n", stats.JobsMatched)
	green.Printf("   • Notified: %d\n", stats.JobsNotified)
	fmt.Printf("   • Skipped: %d (already seen)\n", stats.JobsSkipped)
}
