package cli

import (
	"fmt"
	"strings"
	"time"

	"jobradar/internal/config"
	"jobradar/internal/model"
	"jobradar/internal/storage"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	historyLimit int
)

var historyCmd = &cobra.Command{
	Use:   "history",
	Short: "View notification history",
	Long:  `Display the history of job notifications that have been sent.`,
	RunE:  runHistory,
}

func init() {
	historyCmd.Flags().IntVarP(&historyLimit, "limit", "n", 20, "number of records to display")
	rootCmd.AddCommand(historyCmd)
}

func runHistory(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	store, err := storage.New(cfg.Storage.Database)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}
	defer store.Close()

	records, err := store.GetNotifyRecords(historyLimit)
	if err != nil {
		return fmt.Errorf("failed to get records: %w", err)
	}

	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("\nJobRadar - Notification History")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")

	if len(records) == 0 {
		fmt.Println("\nNo notification records found.")
		return nil
	}

	fmt.Println()

	// Print table header
	headerFmt := "%-20s  %-40s  %-10s  %-8s\n"
	rowFmt := "%-20s  %-40s  %-10s  %-8s\n"

	gray := color.New(color.FgHiBlack)
	gray.Printf(headerFmt, "Time", "Job Title", "Channel", "Status")
	gray.Println(strings.Repeat("─", 85))

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	for _, r := range records {
		timeStr := r.CreatedAt.Format("2006-01-02 15:04:05")
		title := truncateString(r.JobTitle, 38)
		channel := r.NotifyChannel

		var status string
		if r.Status == model.NotifyStatusSent {
			status = "✅ Sent"
		} else if r.Status == model.NotifyStatusFailed {
			status = "❌ Failed"
		} else {
			status = string(r.Status)
		}

		if r.Status == model.NotifyStatusSent {
			green.Printf(rowFmt, timeStr, title, channel, status)
		} else if r.Status == model.NotifyStatusFailed {
			red.Printf(rowFmt, timeStr, title, channel, status)
		} else {
			fmt.Printf(rowFmt, timeStr, title, channel, status)
		}
	}

	fmt.Println()
	fmt.Printf("Showing %d most recent notifications\n", len(records))

	return nil
}

func truncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-3] + "..."
}

// formatTimeAgo formats a time as a relative string
func formatTimeAgo(t time.Time) string {
	delta := time.Since(t)
	hours := delta.Hours()

	if hours < 1 {
		return fmt.Sprintf("%d min ago", int(delta.Minutes()))
	} else if hours < 24 {
		return fmt.Sprintf("%d hrs ago", int(hours))
	}
	return fmt.Sprintf("%d days ago", int(hours/24))
}
