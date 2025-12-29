package cli

import (
	"fmt"

	"jobradar/internal/config"
	"jobradar/internal/engine"
	"jobradar/internal/notifier"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate configuration",
	Long:  `Validate the configuration file and check all settings.`,
	RunE:  runValidate,
}

var testNotifyCmd = &cobra.Command{
	Use:   "test-notify",
	Short: "Test notification channels",
	Long:  `Send a test notification to all enabled channels to verify configuration.`,
	RunE:  runTestNotify,
}

func init() {
	rootCmd.AddCommand(validateCmd)
	rootCmd.AddCommand(testNotifyCmd)
}

func runValidate(cmd *cobra.Command, args []string) error {
	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("\nJobRadar - Configuration Validation")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)
	yellow := color.New(color.FgYellow)

	// Try to load configuration
	cfg, err := config.Load()
	if err != nil {
		red.Println("❌ Configuration validation failed:")
		fmt.Printf("   %v\n", err)
		return nil
	}

	green.Println("✅ Configuration file is valid")
	fmt.Println()

	// Display configuration summary
	fmt.Println("📋 Configuration Summary:")
	fmt.Println()

	fmt.Printf("   Name: %s\n", cfg.Name)
	fmt.Printf("   Searches: %d configured\n", len(cfg.Searches))
	for _, s := range cfg.Searches {
		fmt.Printf("      • %s (%d keywords)\n", s.Name, len(s.Keywords))
	}
	fmt.Println()

	fmt.Println("   Filters:")
	fmt.Printf("      • Budget: $%d - $%d\n", cfg.Filters.Budget.Min, cfg.Filters.Budget.Max)
	fmt.Printf("      • Job Type: %s\n", cfg.Filters.JobType)
	fmt.Printf("      • Posted Within: %d hours\n", cfg.Filters.PostedWithinHours)
	if cfg.Filters.MaxProposals != nil {
		fmt.Printf("      • Max Proposals: %d\n", *cfg.Filters.MaxProposals)
	}
	fmt.Printf("      • Exclude Keywords: %d\n", len(cfg.Filters.ExcludeKeywords))
	fmt.Println()

	fmt.Println("   Notifications:")
	if cfg.Notifications.Telegram.Enabled {
		green.Println("      • Telegram: Enabled")
	} else {
		yellow.Println("      • Telegram: Disabled")
	}
	if cfg.Notifications.Email.Enabled {
		green.Println("      • Email: Enabled")
	} else {
		yellow.Println("      • Email: Disabled")
	}
	fmt.Println()

	fmt.Println("   Schedule:")
	fmt.Printf("      • Interval: %d minutes\n", cfg.Schedule.IntervalMinutes)
	if cfg.Schedule.QuietHours.Enabled {
		fmt.Printf("      • Quiet Hours: %s - %s (%s)\n",
			cfg.Schedule.QuietHours.Start,
			cfg.Schedule.QuietHours.End,
			cfg.Schedule.QuietHours.Timezone)
	}
	fmt.Println()

	fmt.Println("   Storage:")
	fmt.Printf("      • Database: %s\n", cfg.Storage.Database)
	fmt.Printf("      • Retention: %d days\n", cfg.Storage.RetentionDays)
	fmt.Println()

	return nil
}

func runTestNotify(cmd *cobra.Command, args []string) error {
	blue := color.New(color.FgBlue, color.Bold)
	blue.Println("\nJobRadar - Test Notifications")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Println()

	green := color.New(color.FgGreen)
	red := color.New(color.FgRed)

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create engine to get notifiers
	eng, err := engine.New(cfg)
	if err != nil {
		return fmt.Errorf("failed to create engine: %w", err)
	}
	defer eng.Close()

	notifiers := eng.GetNotifiers()
	if len(notifiers) == 0 {
		fmt.Println("⚠️  No notification channels are enabled.")
		return nil
	}

	fmt.Println("📤 Sending test notifications...")
	fmt.Println()

	for _, n := range notifiers {
		fmt.Printf("   Testing %s... ", n.Name())

		var testErr error
		switch tn := n.(type) {
		case *notifier.TelegramNotifier:
			testErr = tn.SendTest()
		case *notifier.EmailNotifier:
			testErr = tn.SendTest()
		default:
			testErr = fmt.Errorf("unknown notifier type")
		}

		if testErr != nil {
			red.Printf("❌ Failed: %v\n", testErr)
		} else {
			green.Println("✅ Success!")
		}
	}

	fmt.Println()
	fmt.Println("Test complete. Check your notification channels for test messages.")

	return nil
}
