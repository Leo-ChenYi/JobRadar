package config

// JobType represents the type of job (fixed price or hourly)
type JobType string

const (
	JobTypeFixed  JobType = "fixed"
	JobTypeHourly JobType = "hourly"
	JobTypeAll    JobType = "all"
)

// SearchConfig represents a single search configuration
type SearchConfig struct {
	Name     string   `yaml:"name" mapstructure:"name"`
	Keywords []string `yaml:"keywords" mapstructure:"keywords"`
	Category string   `yaml:"category,omitempty" mapstructure:"category"`
}

// BudgetFilter represents budget range filter
type BudgetFilter struct {
	Min int `yaml:"min" mapstructure:"min"`
	Max int `yaml:"max" mapstructure:"max"`
}

// FilterConfig represents all filter conditions
type FilterConfig struct {
	Budget            BudgetFilter `yaml:"budget" mapstructure:"budget"`
	JobType           JobType      `yaml:"job_type" mapstructure:"job_type"`
	PostedWithinHours int          `yaml:"posted_within_hours" mapstructure:"posted_within_hours"`
	MaxProposals      *int         `yaml:"max_proposals,omitempty" mapstructure:"max_proposals"`
	ExcludeKeywords   []string     `yaml:"exclude_keywords" mapstructure:"exclude_keywords"`
}

// TelegramConfig represents Telegram notification settings
type TelegramConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	BotToken string `yaml:"bot_token" mapstructure:"bot_token"`
	ChatID   string `yaml:"chat_id" mapstructure:"chat_id"`
}

// EmailConfig represents email notification settings
type EmailConfig struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	SMTPHost string `yaml:"smtp_host" mapstructure:"smtp_host"`
	SMTPPort int    `yaml:"smtp_port" mapstructure:"smtp_port"`
	Username string `yaml:"username" mapstructure:"username"`
	Password string `yaml:"password" mapstructure:"password"`
	To       string `yaml:"to" mapstructure:"to"`
}

// NotificationConfig represents all notification channels
type NotificationConfig struct {
	Telegram TelegramConfig `yaml:"telegram" mapstructure:"telegram"`
	Email    EmailConfig    `yaml:"email" mapstructure:"email"`
}

// QuietHours represents the quiet period configuration
type QuietHours struct {
	Enabled  bool   `yaml:"enabled" mapstructure:"enabled"`
	Start    string `yaml:"start" mapstructure:"start"`
	End      string `yaml:"end" mapstructure:"end"`
	Timezone string `yaml:"timezone" mapstructure:"timezone"`
}

// ScheduleConfig represents scheduling settings
type ScheduleConfig struct {
	IntervalMinutes int        `yaml:"interval_minutes" mapstructure:"interval_minutes"`
	QuietHours      QuietHours `yaml:"quiet_hours" mapstructure:"quiet_hours"`
}

// StorageConfig represents storage settings
type StorageConfig struct {
	Database      string `yaml:"database" mapstructure:"database"`
	RetentionDays int    `yaml:"retention_days" mapstructure:"retention_days"`
}

// AppConfig represents the complete application configuration
type AppConfig struct {
	Name          string             `yaml:"name" mapstructure:"name"`
	Searches      []SearchConfig     `yaml:"searches" mapstructure:"searches"`
	Filters       FilterConfig       `yaml:"filters" mapstructure:"filters"`
	Notifications NotificationConfig `yaml:"notifications" mapstructure:"notifications"`
	Schedule      ScheduleConfig     `yaml:"schedule" mapstructure:"schedule"`
	Storage       StorageConfig      `yaml:"storage" mapstructure:"storage"`
}

// DefaultConfig returns a configuration with sensible defaults
func DefaultConfig() *AppConfig {
	maxProposals := 20
	return &AppConfig{
		Name: "JobRadar",
		Filters: FilterConfig{
			Budget:            BudgetFilter{Min: 0, Max: 100000},
			JobType:           JobTypeAll,
			PostedWithinHours: 24,
			MaxProposals:      &maxProposals,
			ExcludeKeywords:   []string{},
		},
		Schedule: ScheduleConfig{
			IntervalMinutes: 30,
			QuietHours: QuietHours{
				Enabled:  false,
				Timezone: "UTC",
			},
		},
		Storage: StorageConfig{
			Database:      "jobradar.db",
			RetentionDays: 7,
		},
	}
}
