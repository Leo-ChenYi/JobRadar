package model

import (
	"testing"
	"time"
)

func TestJob_BudgetDisplay(t *testing.T) {
	tests := []struct {
		name string
		job  Job
		want string
	}{
		{
			name: "fixed budget range",
			job: Job{
				JobType:   JobTypeFixed,
				BudgetMin: floatPtr(100),
				BudgetMax: floatPtr(500),
			},
			want: "$100-$500 (Fixed)",
		},
		{
			name: "fixed budget single value",
			job: Job{
				JobType:   JobTypeFixed,
				BudgetMin: floatPtr(300),
				BudgetMax: floatPtr(300),
			},
			want: "$300 (Fixed)",
		},
		{
			name: "fixed budget max only",
			job: Job{
				JobType:   JobTypeFixed,
				BudgetMax: floatPtr(500),
			},
			want: "$500 (Fixed)",
		},
		{
			name: "fixed budget min only",
			job: Job{
				JobType:   JobTypeFixed,
				BudgetMin: floatPtr(200),
			},
			want: "$200+ (Fixed)",
		},
		{
			name: "fixed no budget",
			job: Job{
				JobType: JobTypeFixed,
			},
			want: "Budget not specified",
		},
		{
			name: "hourly rate range",
			job: Job{
				JobType:       JobTypeHourly,
				HourlyRateMin: floatPtr(25),
				HourlyRateMax: floatPtr(50),
			},
			want: "$25-$50/hr",
		},
		{
			name: "hourly rate min only",
			job: Job{
				JobType:       JobTypeHourly,
				HourlyRateMin: floatPtr(30),
			},
			want: "$30+/hr",
		},
		{
			name: "hourly no rate",
			job: Job{
				JobType: JobTypeHourly,
			},
			want: "Hourly rate not specified",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.job.BudgetDisplay()
			if got != tt.want {
				t.Errorf("BudgetDisplay() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestJob_PostedAgo(t *testing.T) {
	tests := []struct {
		name     string
		postedAt time.Time
		want     string
	}{
		{
			name:     "just now",
			postedAt: time.Now().Add(-30 * time.Second),
			want:     "just now",
		},
		{
			name:     "minutes ago",
			postedAt: time.Now().Add(-15 * time.Minute),
			want:     "15 minutes ago",
		},
		{
			name:     "1 hour ago",
			postedAt: time.Now().Add(-1 * time.Hour),
			want:     "1 hour ago",
		},
		{
			name:     "hours ago",
			postedAt: time.Now().Add(-5 * time.Hour),
			want:     "5 hours ago",
		},
		{
			name:     "1 day ago",
			postedAt: time.Now().Add(-25 * time.Hour),
			want:     "1 day ago",
		},
		{
			name:     "days ago",
			postedAt: time.Now().Add(-72 * time.Hour),
			want:     "3 days ago",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := Job{PostedAt: tt.postedAt}
			got := job.PostedAgo()
			if got != tt.want {
				t.Errorf("PostedAgo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMatchedJob(t *testing.T) {
	job := &Job{
		ID:    "test-123",
		Title: "Test Job",
	}
	keywords := []string{"golang", "api"}
	searchName := "Test Search"

	matched := NewMatchedJob(job, keywords, searchName)

	if matched.Job != job {
		t.Error("Job reference mismatch")
	}

	if len(matched.MatchedKeywords) != 2 {
		t.Errorf("MatchedKeywords count = %d, want 2", len(matched.MatchedKeywords))
	}

	if matched.SearchName != searchName {
		t.Errorf("SearchName = %v, want %v", matched.SearchName, searchName)
	}

	if matched.MatchScore != 1.0 {
		t.Errorf("MatchScore = %v, want 1.0", matched.MatchScore)
	}
}

func floatPtr(f float64) *float64 {
	return &f
}
