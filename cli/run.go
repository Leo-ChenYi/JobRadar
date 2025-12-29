package cli

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"jobradar/internal/config"
	"jobradar/internal/engine"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Start scheduled job monitoring",
	Long: `Start the scheduler to periodically check for new jobs.
The check interval is configured in config.yaml.
Press Ctrl+C to stop.`,
	RunE: runScheduler,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runScheduler(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("\nJobRadar v1.0.0")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("⏰ Starting scheduler (every %d minutes)\n", cfg.Schedule.IntervalMinutes)

	if cfg.Schedule.QuietHours.Enabled {
		fmt.Printf("🌙 Quiet hours: %s - %s (%s)\n",
			cfg.Schedule.QuietHours.Start,
			cfg.Schedule.QuietHours.End,
			cfg.Schedule.QuietHours.Timezone)
	}

	fmt.Println()
	fmt.Println("Press Ctrl+C to stop")
	fmt.Println()

	eng, err := engine.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	// Run initial check
	fmt.Println("🔍 Running initial check...")
	if _, err := eng.Run(); err != nil {
		fmt.Printf("⚠️  Initial check failed: %v\n", err)
	}
	fmt.Println()

	// Start scheduler
	eng.StartScheduler()

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan

	fmt.Println()
	fmt.Println("🛑 Shutting down...")
	eng.StopScheduler()

	green := color.New(color.FgGreen)
	green.Println("✅ Scheduler stopped gracefully")

	return nil
}
