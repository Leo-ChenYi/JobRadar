package scheduler

import (
	"fmt"
	"time"

	"jobradar/internal/config"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
)

// Scheduler handles scheduled job checks
type Scheduler struct {
	config   config.ScheduleConfig
	cron     *cron.Cron
	location *time.Location
}

// New creates a new Scheduler instance
func New(cfg config.ScheduleConfig) *Scheduler {
	// Parse timezone
	loc := time.UTC
	if cfg.QuietHours.Timezone != "" {
		if l, err := time.LoadLocation(cfg.QuietHours.Timezone); err == nil {
			loc = l
		} else {
			log.Warn().Str("timezone", cfg.QuietHours.Timezone).Msg("Invalid timezone, using UTC")
		}
	}

	return &Scheduler{
		config:   cfg,
		cron:     cron.New(),
		location: loc,
	}
}

// AddJob adds a job to be executed at the configured interval
func (s *Scheduler) AddJob(fn func()) error {
	spec := fmt.Sprintf("@every %dm", s.config.IntervalMinutes)

	_, err := s.cron.AddFunc(spec, func() {
		// Check quiet hours
		if s.isQuietHours() {
			log.Debug().Msg("Skipping check during quiet hours")
			return
		}
		fn()
	})

	if err != nil {
		return fmt.Errorf("failed to add cron job: %w", err)
	}

	return nil
}

// Start starts the scheduler
func (s *Scheduler) Start() {
	log.Info().Int("interval", s.config.IntervalMinutes).Msg("Starting scheduler")
	s.cron.Start()
}

// Stop stops the scheduler
func (s *Scheduler) Stop() {
	log.Info().Msg("Stopping scheduler")
	ctx := s.cron.Stop()
	<-ctx.Done()
}

// isQuietHours checks if current time is within quiet hours
func (s *Scheduler) isQuietHours() bool {
	if !s.config.QuietHours.Enabled {
		return false
	}

	now := time.Now().In(s.location)
	currentMinutes := now.Hour()*60 + now.Minute()

	// Parse start time
	startHour, startMin, err := parseTime(s.config.QuietHours.Start)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid quiet hours start time")
		return false
	}
	startMinutes := startHour*60 + startMin

	// Parse end time
	endHour, endMin, err := parseTime(s.config.QuietHours.End)
	if err != nil {
		log.Warn().Err(err).Msg("Invalid quiet hours end time")
		return false
	}
	endMinutes := endHour*60 + endMin

	// Handle overnight quiet hours (e.g., 23:00 - 07:00)
	if startMinutes > endMinutes {
		// Quiet hours span midnight
		return currentMinutes >= startMinutes || currentMinutes < endMinutes
	}

	// Normal case (e.g., 01:00 - 06:00)
	return currentMinutes >= startMinutes && currentMinutes < endMinutes
}

// parseTime parses a time string in HH:MM format
func parseTime(timeStr string) (int, int, error) {
	var hour, min int
	_, err := fmt.Sscanf(timeStr, "%d:%d", &hour, &min)
	if err != nil {
		return 0, 0, fmt.Errorf("invalid time format: %s", timeStr)
	}
	if hour < 0 || hour > 23 || min < 0 || min > 59 {
		return 0, 0, fmt.Errorf("time out of range: %s", timeStr)
	}
	return hour, min, nil
}
