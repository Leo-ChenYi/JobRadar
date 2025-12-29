package cli

import (
	"fmt"

	"jobradar/internal/config"
	"jobradar/internal/storage"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var statsCmd = &cobra.Command{
	Use:   "stats",
	Short: "View statistics",
	Long:  `Display aggregate statistics about job monitoring and notifications.`,
	RunE:  runStats,
}

func init() {
	rootCmd.AddCommand(statsCmd)
}

func runStats(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := storage.New(cfg.Storage.Database)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	stats, err := store.GetOverallStats()
	if err != nil {
		return fmt.Errorf("failed to get stats: %w", err)
	}

	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("\nJobRadar - Statistics")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	cyan := color.New(color.FgCyan)
	green := color.New(color.FgGreen)

	fmt.Println("📊 Overall Statistics:")
	fmt.Println()

	cyan.Print("   Total Runs:        ")
	fmt.Printf("%d\n", stats.TotalRuns)

	cyan.Print("   Jobs Fetched:      ")
	fmt.Printf("%d\n", stats.TotalJobsFetched)

	cyan.Print("   Jobs Matched:      ")
	fmt.Printf("%d\n", stats.TotalJobsMatched)

	cyan.Print("   Jobs Notified:     ")
	green.Printf("%d\n", stats.TotalJobsNotified)

	fmt.Println()

	if stats.LastRunAt != nil {
		cyan.Print("   Last Run:          ")
		fmt.Printf("%s (%s)\n", stats.LastRunAt.Format("2006-01-02 15:04:05"), formatTimeAgo(*stats.LastRunAt))
	}

	if stats.LastMatchAt != nil {
		cyan.Print("   Last Match:        ")
		fmt.Printf("%s (%s)\n", stats.LastMatchAt.Format("2006-01-02 15:04:05"), formatTimeAgo(*stats.LastMatchAt))
	}

	fmt.Println()

	// Calculate match rate
	if stats.TotalJobsFetched > 0 {
		matchRate := float64(stats.TotalJobsMatched) / float64(stats.TotalJobsFetched) * 100
		cyan.Print("   Match Rate:        ")
		fmt.Printf("%.1f%%\n", matchRate)
	}

	// Calculate notification rate
	if stats.TotalJobsMatched > 0 {
		notifyRate := float64(stats.TotalJobsNotified) / float64(stats.TotalJobsMatched) * 100
		cyan.Print("   Notification Rate: ")
		fmt.Printf("%.1f%%\n", notifyRate)
	}

	fmt.Println()

	return nil
}
