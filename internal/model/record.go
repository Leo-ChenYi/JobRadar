package model

import "time"

// NotifyStatus represents the status of a notification
type NotifyStatus string

const (
	NotifyStatusPending NotifyStatus = "pending"
	NotifyStatusSent    NotifyStatus = "sent"
	NotifyStatusFailed  NotifyStatus = "failed"
	NotifyStatusSkipped NotifyStatus = "skipped"
)

// NotifyRecord represents a notification record
type NotifyRecord struct {
	ID              int64        `json:"id" db:"id"`
	JobID           string       `json:"job_id" db:"job_id"`
	JobTitle        string       `json:"job_title" db:"job_title"`
	JobURL          string       `json:"job_url" db:"job_url"`
	SearchName      string       `json:"search_name" db:"search_name"`
	MatchedKeywords string       `json:"matched_keywords" db:"matched_keywords"`
	NotifyChannel   string       `json:"notify_channel" db:"notify_channel"`
	Status          NotifyStatus `json:"status" db:"status"`
	ErrorMessage    string       `json:"error_message,omitempty" db:"error_message"`
	CreatedAt       time.Time    `json:"created_at" db:"created_at"`
	SentAt          *time.Time   `json:"sent_at,omitempty" db:"sent_at"`
}

// JobSeen represents a record of a job that has been seen (for deduplication)
type JobSeen struct {
	JobID       string    `json:"job_id" db:"job_id"`
	JobTitle    string    `json:"job_title" db:"job_title"`
	JobURL      string    `json:"job_url" db:"job_url"`
	FirstSeenAt time.Time `json:"first_seen_at" db:"first_seen_at"`
	Notified    bool      `json:"notified" db:"notified"`
	CreatedAt   time.Time `json:"created_at" db:"created_at"`
}
