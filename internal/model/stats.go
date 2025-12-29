package model

import "time"

// RunStats represents statistics for a single run
type RunStats struct {
	StartedAt       time.Time  `json:"started_at"`
	FinishedAt      *time.Time `json:"finished_at,omitempty"`
	DurationSeconds float64    `json:"duration_seconds"`
	JobsFetched     int        `json:"jobs_fetched"`
	JobsMatched     int        `json:"jobs_matched"`
	JobsNotified    int        `json:"jobs_notified"`
	JobsSkipped     int        `json:"jobs_skipped"`
	ErrorMessage    string     `json:"error_message,omitempty"`
}

// NewRunStats creates a new RunStats instance with start time set
func NewRunStats() *RunStats {
	return &RunStats{
		StartedAt: time.Now(),
	}
}

// Finish marks the run as complete and calculates duration
func (s *RunStats) Finish() {
	now := time.Now()
	s.FinishedAt = &now
	s.DurationSeconds = now.Sub(s.StartedAt).Seconds()
}

// OverallStats represents aggregate statistics
type OverallStats struct {
	TotalRuns         int        `json:"total_runs"`
	TotalJobsFetched  int        `json:"total_jobs_fetched"`
	TotalJobsMatched  int        `json:"total_jobs_matched"`
	TotalJobsNotified int        `json:"total_jobs_notified"`
	LastRunAt         *time.Time `json:"last_run_at,omitempty"`
	LastMatchAt       *time.Time `json:"last_match_at,omitempty"`
}
