package engine

import (
	"fmt"
	"strings"
	"time"

	"jobradar/internal/config"
	"jobradar/internal/fetcher"
	"jobradar/internal/filter"
	"jobradar/internal/model"
	"jobradar/internal/notifier"
	"jobradar/internal/scheduler"
	"jobradar/internal/storage"

	"github.com/rs/zerolog/log"
)

// Engine is the main JobRadar engine that coordinates all components
type Engine struct {
	config    *config.AppConfig
	storage   *storage.Storage
	fetcher   *fetcher.RSSFetcher
	filter    *filter.Filter
	notifiers []notifier.Notifier
	scheduler *scheduler.Scheduler
}

// New creates a new Engine instance
func New(cfg *config.AppConfig) (*Engine, error) {
	// Initialize storage
	store, err := storage.New(cfg.Storage.Database)
	if err != nil {
		return nil, fmt.Errorf("failed to init storage: %w", err)
	}

	// Initialize notifiers
	notifiers := make([]notifier.Notifier, 0)
	if cfg.Notifications.Telegram.Enabled {
		n := notifier.NewTelegram(cfg.Notifications.Telegram)
		notifiers = append(notifiers, n)
	}
	if cfg.Notifications.Email.Enabled {
		n := notifier.NewEmail(cfg.Notifications.Email)
		notifiers = append(notifiers, n)
	}

	return &Engine{
		config:    cfg,
		storage:   store,
		fetcher:   fetcher.NewRSSFetcher(),
		filter:    filter.New(cfg.Filters),
		notifiers: notifiers,
	}, nil
}

// Run executes a single check cycle
func (e *Engine) Run() (*model.RunStats, error) {
	stats := model.NewRunStats()

	log.Info().Msg("Fetching jobs from RSS feeds...")

	// 1. Fetch jobs - prefer direct RSS URLs over keyword search
	var allJobs []*model.Job
	var feedNames []string // Track which feeds the jobs came from

	// Method 1: Use direct RSS URLs (recommended)
	if len(e.config.RSSFeeds) > 0 {
		for _, feed := range e.config.RSSFeeds {
			jobs, err := e.fetcher.FetchFromURL(feed.URL)
			if err != nil {
				log.Error().Err(err).Str("feed", feed.Name).Msg("Failed to fetch RSS feed")
				continue
			}
			for _, job := range jobs {
				allJobs = append(allJobs, job)
				feedNames = append(feedNames, feed.Name)
			}
			log.Info().Str("feed", feed.Name).Int("count", len(jobs)).Msg("Fetched jobs from RSS feed")
		}
	}

	// Method 2: Fallback to keyword search (deprecated)
	if len(e.config.RSSFeeds) == 0 && len(e.config.Searches) > 0 {
		log.Warn().Msg("Using deprecated keyword search - public RSS is no longer supported by Upwork")
		for _, search := range e.config.Searches {
			jobs, err := e.fetcher.Fetch(search.Keywords)
			if err != nil {
				log.Error().Err(err).Str("search", search.Name).Msg("Failed to fetch")
				continue
			}
			for _, job := range jobs {
				allJobs = append(allJobs, job)
				feedNames = append(feedNames, search.Name)
			}
			log.Debug().Str("search", search.Name).Int("count", len(jobs)).Msg("Fetched jobs")
		}
	}

	stats.JobsFetched = len(allJobs)
	log.Info().Int("total", stats.JobsFetched).Msg("Total jobs fetched")

	// 2. Filter and match jobs
	log.Info().Msg("Filtering jobs...")
	var matchedJobs []*model.MatchedJob

	for i, job := range allJobs {
		// Get keywords to match against
		var keywords []string
		feedName := feedNames[i]

		// If using RSS feeds, use filter exclude keywords only (job already matches the feed criteria)
		if len(e.config.RSSFeeds) > 0 {
			// For RSS feeds, we consider all jobs as "matched" since they already come from a filtered feed
			// We just apply our additional filters
			keywords = []string{feedName} // Use feed name as the "matched keyword"
		} else {
			// For keyword search, find matching keywords
			for _, search := range e.config.Searches {
				if search.Name == feedName {
					keywords = search.Keywords
					break
				}
			}
		}

		matchedKeywords := e.filter.Match(job, keywords)
		if len(matchedKeywords) > 0 {
			matchedJobs = append(matchedJobs, model.NewMatchedJob(job, matchedKeywords, feedName))
		}
	}

	stats.JobsMatched = len(matchedJobs)
	log.Info().Int("matched", stats.JobsMatched).Msg("Jobs matched")

	// 3. Filter out already seen jobs
	var newJobs []*model.MatchedJob
	for _, matched := range matchedJobs {
		seen, err := e.storage.IsSeen(matched.Job.ID)
		if err != nil {
			log.Error().Err(err).Msg("Failed to check if seen")
			continue
		}
		if !seen {
			newJobs = append(newJobs, matched)
		} else {
			stats.JobsSkipped++
		}
	}

	log.Info().Int("new", len(newJobs)).Int("skipped", stats.JobsSkipped).Msg("Filtered seen jobs")

	// 4. Send notifications
	if len(newJobs) > 0 {
		log.Info().Int("count", len(newJobs)).Msg("Sending notifications...")

		for _, matched := range newJobs {
			if e.notify(matched) {
				stats.JobsNotified++
				e.storage.MarkSeen(matched.Job.ID, matched.Job.Title, matched.Job.URL)
			}
		}
	}

	stats.Finish()

	// Save run log
	if err := e.storage.SaveRunLog(stats); err != nil {
		log.Error().Err(err).Msg("Failed to save run log")
	}

	// Cleanup old records
	if err := e.storage.Cleanup(e.config.Storage.RetentionDays); err != nil {
		log.Error().Err(err).Msg("Failed to cleanup old records")
	}

	log.Info().Int("notified", stats.JobsNotified).Float64("duration", stats.DurationSeconds).Msg("Check completed")

	return stats, nil
}

// notify sends notifications to all enabled channels
func (e *Engine) notify(matched *model.MatchedJob) bool {
	success := false

	for _, n := range e.notifiers {
		if err := n.Send(matched); err != nil {
			log.Error().Err(err).Str("channel", n.Name()).Msg("Failed to send notification")

			e.storage.SaveNotifyRecord(&model.NotifyRecord{
				JobID:           matched.Job.ID,
				JobTitle:        matched.Job.Title,
				JobURL:          matched.Job.URL,
				SearchName:      matched.SearchName,
				MatchedKeywords: strings.Join(matched.MatchedKeywords, ","),
				NotifyChannel:   n.Name(),
				Status:          model.NotifyStatusFailed,
				ErrorMessage:    err.Error(),
				CreatedAt:       time.Now(),
			})
		} else {
			success = true
			now := time.Now()

			e.storage.SaveNotifyRecord(&model.NotifyRecord{
				JobID:           matched.Job.ID,
				JobTitle:        matched.Job.Title,
				JobURL:          matched.Job.URL,
				SearchName:      matched.SearchName,
				MatchedKeywords: strings.Join(matched.MatchedKeywords, ","),
				NotifyChannel:   n.Name(),
				Status:          model.NotifyStatusSent,
				CreatedAt:       now,
				SentAt:          &now,
			})

			log.Debug().Str("channel", n.Name()).Str("job", matched.Job.Title).Msg("Notification sent")
		}
	}

	return success
}

// StartScheduler starts the scheduled job monitoring
func (e *Engine) StartScheduler() {
	e.scheduler = scheduler.New(e.config.Schedule)
	e.scheduler.AddJob(func() {
		if _, err := e.Run(); err != nil {
			log.Error().Err(err).Msg("Scheduled check failed")
		}
	})
	e.scheduler.Start()
}

// StopScheduler stops the scheduled job monitoring
func (e *Engine) StopScheduler() {
	if e.scheduler != nil {
		e.scheduler.Stop()
	}
}

// GetStorage returns the storage instance
func (e *Engine) GetStorage() *storage.Storage {
	return e.storage
}

// GetNotifiers returns the notifiers
func (e *Engine) GetNotifiers() []notifier.Notifier {
	return e.notifiers
}

// Close closes the engine and releases resources
func (e *Engine) Close() error {
	return e.storage.Close()
}
