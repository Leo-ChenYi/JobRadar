package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

// envVarRegex matches ${VAR_NAME} patterns
var envVarRegex = regexp.MustCompile(`\$\{([^}]+)\}`)

// Load reads and parses the configuration file
func Load() (*AppConfig, error) {
	cfg := DefaultConfig()

	if err := viper.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Expand environment variables in sensitive fields
	expandEnvVars(cfg)

	// Validate configuration
	if err := validate(cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

// expandEnvVars replaces ${VAR} patterns with environment variable values
func expandEnvVars(cfg *AppConfig) {
	// Telegram config
	cfg.Notifications.Telegram.BotToken = expandEnvVar(cfg.Notifications.Telegram.BotToken)
	cfg.Notifications.Telegram.ChatID = expandEnvVar(cfg.Notifications.Telegram.ChatID)

	// Email config
	cfg.Notifications.Email.Username = expandEnvVar(cfg.Notifications.Email.Username)
	cfg.Notifications.Email.Password = expandEnvVar(cfg.Notifications.Email.Password)
}

// expandEnvVar expands a single ${VAR} pattern
func expandEnvVar(s string) string {
	return envVarRegex.ReplaceAllStringFunc(s, func(match string) string {
		// Extract variable name from ${VAR_NAME}
		varName := match[2 : len(match)-1]
		if value := os.Getenv(varName); value != "" {
			return value
		}
		return match // Keep original if env var not found
	})
}

// validate checks if the configuration is valid
func validate(cfg *AppConfig) error {
	var errors []string

	// Check if at least one search is configured
	if len(cfg.Searches) == 0 {
		errors = append(errors, "at least one search configuration is required")
	}

	// Validate searches
	for i, search := range cfg.Searches {
		if search.Name == "" {
			errors = append(errors, fmt.Sprintf("search[%d]: name is required", i))
		}
		if len(search.Keywords) == 0 {
			errors = append(errors, fmt.Sprintf("search[%d]: at least one keyword is required", i))
		}
	}

	// Validate budget
	if cfg.Filters.Budget.Min < 0 {
		errors = append(errors, "filters.budget.min cannot be negative")
	}
	if cfg.Filters.Budget.Max < cfg.Filters.Budget.Min {
		errors = append(errors, "filters.budget.max must be >= min")
	}

	// Validate job type
	switch cfg.Filters.JobType {
	case JobTypeFixed, JobTypeHourly, JobTypeAll:
		// Valid
	default:
		errors = append(errors, fmt.Sprintf("invalid job_type: %s (must be fixed, hourly, or all)", cfg.Filters.JobType))
	}

	// Validate notifications - at least one should be enabled
	if !cfg.Notifications.Telegram.Enabled && !cfg.Notifications.Email.Enabled {
		errors = append(errors, "at least one notification channel must be enabled")
	}

	// Validate Telegram config if enabled
	if cfg.Notifications.Telegram.Enabled {
		if cfg.Notifications.Telegram.BotToken == "" || strings.HasPrefix(cfg.Notifications.Telegram.BotToken, "${") {
			errors = append(errors, "telegram.bot_token is required when telegram is enabled")
		}
		if cfg.Notifications.Telegram.ChatID == "" || strings.HasPrefix(cfg.Notifications.Telegram.ChatID, "${") {
			errors = append(errors, "telegram.chat_id is required when telegram is enabled")
		}
	}

	// Validate Email config if enabled
	if cfg.Notifications.Email.Enabled {
		if cfg.Notifications.Email.SMTPHost == "" {
			errors = append(errors, "email.smtp_host is required when email is enabled")
		}
		if cfg.Notifications.Email.To == "" {
			errors = append(errors, "email.to is required when email is enabled")
		}
	}

	// Validate schedule
	if cfg.Schedule.IntervalMinutes < 1 {
		errors = append(errors, "schedule.interval_minutes must be at least 1")
	}

	// Validate storage
	if cfg.Storage.Database == "" {
		errors = append(errors, "storage.database is required")
	}
	if cfg.Storage.RetentionDays < 1 {
		errors = append(errors, "storage.retention_days must be at least 1")
	}

	if len(errors) > 0 {
		return fmt.Errorf("configuration errors:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}

// ValidateOnly validates the configuration without returning it
func ValidateOnly() error {
	_, err := Load()
	return err
}
