package model

import (
	"fmt"
	"time"
)

// JobType represents the type of job
type JobType string

const (
	JobTypeFixed  JobType = "fixed"
	JobTypeHourly JobType = "hourly"
)

// Job represents an Upwork job posting
type Job struct {
	ID          string `json:"id"`          // Upwork Job ID
	Title       string `json:"title"`       // Job title
	Description string `json:"description"` // Job description
	URL         string `json:"url"`         // Job URL

	// Budget information
	JobType       JobType  `json:"job_type"`
	BudgetMin     *float64 `json:"budget_min,omitempty"`
	BudgetMax     *float64 `json:"budget_max,omitempty"`
	HourlyRateMin *float64 `json:"hourly_rate_min,omitempty"`
	HourlyRateMax *float64 `json:"hourly_rate_max,omitempty"`

	// Competition information
	Proposals *int `json:"proposals,omitempty"`

	// Client information
	ClientCountry    string   `json:"client_country,omitempty"`
	ClientRating     *float64 `json:"client_rating,omitempty"`
	ClientTotalSpent *float64 `json:"client_total_spent,omitempty"`
	ClientTotalHires *int     `json:"client_total_hires,omitempty"`

	// Skill tags
	Skills []string `json:"skills"`

	// Time information
	PostedAt  time.Time `json:"posted_at"`
	FetchedAt time.Time `json:"fetched_at"`
}

// BudgetDisplay formats the budget for display
func (j *Job) BudgetDisplay() string {
	if j.JobType == JobTypeFixed {
		if j.BudgetMin != nil && j.BudgetMax != nil {
			if *j.BudgetMin == *j.BudgetMax {
				return fmt.Sprintf("$%.0f (Fixed)", *j.BudgetMax)
			}
			return fmt.Sprintf("$%.0f-$%.0f (Fixed)", *j.BudgetMin, *j.BudgetMax)
		} else if j.BudgetMax != nil {
			return fmt.Sprintf("$%.0f (Fixed)", *j.BudgetMax)
		} else if j.BudgetMin != nil {
			return fmt.Sprintf("$%.0f+ (Fixed)", *j.BudgetMin)
		}
		return "Budget not specified"
	}

	if j.HourlyRateMin != nil && j.HourlyRateMax != nil {
		return fmt.Sprintf("$%.0f-$%.0f/hr", *j.HourlyRateMin, *j.HourlyRateMax)
	} else if j.HourlyRateMin != nil {
		return fmt.Sprintf("$%.0f+/hr", *j.HourlyRateMin)
	}
	return "Hourly rate not specified"
}

// PostedAgo returns a human-readable time since posting
func (j *Job) PostedAgo() string {
	delta := time.Since(j.PostedAt)
	hours := delta.Hours()

	if hours < 1 {
		minutes := int(delta.Minutes())
		if minutes <= 1 {
			return "just now"
		}
		return fmt.Sprintf("%d minutes ago", minutes)
	} else if hours < 24 {
		h := int(hours)
		if h == 1 {
			return "1 hour ago"
		}
		return fmt.Sprintf("%d hours ago", h)
	}
	days := int(hours / 24)
	if days == 1 {
		return "1 day ago"
	}
	return fmt.Sprintf("%d days ago", days)
}

// MatchedJob represents a job that matched the search criteria
type MatchedJob struct {
	Job             *Job     `json:"job"`
	MatchedKeywords []string `json:"matched_keywords"`
	SearchName      string   `json:"search_name"`
	MatchScore      float64  `json:"match_score"`
}

// NewMatchedJob creates a new matched job instance
func NewMatchedJob(job *Job, keywords []string, searchName string) *MatchedJob {
	return &MatchedJob{
		Job:             job,
		MatchedKeywords: keywords,
		SearchName:      searchName,
		MatchScore:      1.0,
	}
}
