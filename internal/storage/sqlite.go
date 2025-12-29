package storage

import (
	"database/sql"
	"fmt"
	"time"

	"jobradar/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

// Storage handles SQLite database operations
type Storage struct {
	db *sql.DB
}

// New creates a new Storage instance and initializes the database
func New(dbPath string) (*Storage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	s := &Storage{db: db}
	if err := s.migrate(); err != nil {
		return nil, fmt.Errorf("failed to migrate database: %w", err)
	}

	return s, nil
}

// migrate creates the necessary tables
func (s *Storage) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS jobs_seen (
			job_id VARCHAR(100) PRIMARY KEY,
			job_title VARCHAR(500),
			job_url VARCHAR(1000),
			first_seen_at TIMESTAMP NOT NULL,
			notified BOOLEAN DEFAULT FALSE,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_jobs_seen_created ON jobs_seen(created_at)`,

		`CREATE TABLE IF NOT EXISTS notify_records (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			job_id VARCHAR(100) NOT NULL,
			job_title VARCHAR(500),
			job_url VARCHAR(1000),
			search_name VARCHAR(100),
			matched_keywords TEXT,
			notify_channel VARCHAR(50) NOT NULL,
			status VARCHAR(20) NOT NULL,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
			sent_at TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_notify_records_job ON notify_records(job_id)`,
		`CREATE INDEX IF NOT EXISTS idx_notify_records_created ON notify_records(created_at)`,
		`CREATE INDEX IF NOT EXISTS idx_notify_records_status ON notify_records(status)`,

		`CREATE TABLE IF NOT EXISTS run_logs (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			started_at TIMESTAMP NOT NULL,
			finished_at TIMESTAMP,
			jobs_fetched INT DEFAULT 0,
			jobs_matched INT DEFAULT 0,
			jobs_notified INT DEFAULT 0,
			jobs_skipped INT DEFAULT 0,
			error_message TEXT,
			created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)`,
		`CREATE INDEX IF NOT EXISTS idx_run_logs_started ON run_logs(started_at)`,
	}

	for _, query := range queries {
		if _, err := s.db.Exec(query); err != nil {
			return fmt.Errorf("failed to execute migration: %w", err)
		}
	}

	return nil
}

// IsSeen checks if a job has been seen before
func (s *Storage) IsSeen(jobID string) (bool, error) {
	var count int
	err := s.db.QueryRow(
		"SELECT COUNT(*) FROM jobs_seen WHERE job_id = ?",
		jobID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("failed to check if job seen: %w", err)
	}
	return count > 0, nil
}

// MarkSeen marks a job as seen
func (s *Storage) MarkSeen(jobID, title, url string) error {
	_, err := s.db.Exec(`
		INSERT OR REPLACE INTO jobs_seen 
		(job_id, job_title, job_url, first_seen_at, notified, created_at)
		VALUES (?, ?, ?, ?, TRUE, ?)
	`, jobID, title, url, time.Now(), time.Now())
	if err != nil {
		return fmt.Errorf("failed to mark job as seen: %w", err)
	}
	return nil
}

// SaveNotifyRecord saves a notification record
func (s *Storage) SaveNotifyRecord(record *model.NotifyRecord) error {
	_, err := s.db.Exec(`
		INSERT INTO notify_records 
		(job_id, job_title, job_url, search_name, matched_keywords, 
		 notify_channel, status, error_message, created_at, sent_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, record.JobID, record.JobTitle, record.JobURL, record.SearchName,
		record.MatchedKeywords, record.NotifyChannel, record.Status,
		record.ErrorMessage, record.CreatedAt, record.SentAt)
	if err != nil {
		return fmt.Errorf("failed to save notify record: %w", err)
	}
	return nil
}

// GetNotifyRecords retrieves notification records
func (s *Storage) GetNotifyRecords(limit int) ([]*model.NotifyRecord, error) {
	rows, err := s.db.Query(`
		SELECT id, job_id, job_title, job_url, search_name, matched_keywords,
		       notify_channel, status, error_message, created_at, sent_at
		FROM notify_records
		ORDER BY created_at DESC
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get notify records: %w", err)
	}
	defer rows.Close()

	var records []*model.NotifyRecord
	for rows.Next() {
		r := &model.NotifyRecord{}
		var sentAt sql.NullTime
		err := rows.Scan(
			&r.ID, &r.JobID, &r.JobTitle, &r.JobURL, &r.SearchName,
			&r.MatchedKeywords, &r.NotifyChannel, &r.Status,
			&r.ErrorMessage, &r.CreatedAt, &sentAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan notify record: %w", err)
		}
		if sentAt.Valid {
			r.SentAt = &sentAt.Time
		}
		records = append(records, r)
	}
	return records, nil
}

// SaveRunLog saves a run log entry
func (s *Storage) SaveRunLog(stats *model.RunStats) error {
	_, err := s.db.Exec(`
		INSERT INTO run_logs 
		(started_at, finished_at, jobs_fetched, jobs_matched, jobs_notified, jobs_skipped, error_message)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, stats.StartedAt, stats.FinishedAt, stats.JobsFetched, stats.JobsMatched,
		stats.JobsNotified, stats.JobsSkipped, stats.ErrorMessage)
	if err != nil {
		return fmt.Errorf("failed to save run log: %w", err)
	}
	return nil
}

// GetOverallStats retrieves aggregate statistics
func (s *Storage) GetOverallStats() (*model.OverallStats, error) {
	stats := &model.OverallStats{}

	// Get totals from run_logs
	err := s.db.QueryRow(`
		SELECT COUNT(*), COALESCE(SUM(jobs_fetched), 0), 
		       COALESCE(SUM(jobs_matched), 0), COALESCE(SUM(jobs_notified), 0)
		FROM run_logs
	`).Scan(&stats.TotalRuns, &stats.TotalJobsFetched,
		&stats.TotalJobsMatched, &stats.TotalJobsNotified)
	if err != nil {
		return nil, fmt.Errorf("failed to get overall stats: %w", err)
	}

	// Get last run time
	var lastRun sql.NullTime
	err = s.db.QueryRow(`
		SELECT MAX(started_at) FROM run_logs
	`).Scan(&lastRun)
	if err == nil && lastRun.Valid {
		stats.LastRunAt = &lastRun.Time
	}

	// Get last match time
	var lastMatch sql.NullTime
	err = s.db.QueryRow(`
		SELECT MAX(created_at) FROM notify_records WHERE status = 'sent'
	`).Scan(&lastMatch)
	if err == nil && lastMatch.Valid {
		stats.LastMatchAt = &lastMatch.Time
	}

	return stats, nil
}

// Cleanup removes records older than the specified retention period
func (s *Storage) Cleanup(retentionDays int) error {
	cutoff := time.Now().AddDate(0, 0, -retentionDays)

	_, err := s.db.Exec(
		"DELETE FROM jobs_seen WHERE created_at < ?",
		cutoff,
	)
	if err != nil {
		return fmt.Errorf("failed to cleanup jobs_seen: %w", err)
	}

	_, err = s.db.Exec(
		"DELETE FROM notify_records WHERE created_at < ?",
		cutoff,
	)
	if err != nil {
		return fmt.Errorf("failed to cleanup notify_records: %w", err)
	}

	_, err = s.db.Exec(
		"DELETE FROM run_logs WHERE created_at < ?",
		cutoff,
	)
	if err != nil {
		return fmt.Errorf("failed to cleanup run_logs: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *Storage) Close() error {
	return s.db.Close()
}
