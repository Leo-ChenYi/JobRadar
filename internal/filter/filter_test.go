package filter

import (
	"testing"
	"time"

	"jobradar/internal/config"
	"jobradar/internal/model"
)

func TestFilter_Match_ExcludeKeywords(t *testing.T) {
	cfg := config.FilterConfig{
		Budget:          config.BudgetFilter{Min: 0, Max: 100000},
		JobType:         config.JobTypeAll,
		ExcludeKeywords: []string{"cheap", "lowest bid"},
	}
	f := New(cfg)

	tests := []struct {
		name     string
		job      *model.Job
		keywords []string
		want     bool
	}{
		{
			name: "job without exclude keywords should match",
			job: &model.Job{
				ID:          "1",
				Title:       "Golang API Developer",
				Description: "We need a skilled developer",
				PostedAt:    time.Now(),
			},
			keywords: []string{"golang", "api"},
			want:     true,
		},
		{
			name: "job with exclude keyword in title should not match",
			job: &model.Job{
				ID:          "2",
				Title:       "Cheap Golang Developer",
				Description: "Looking for developer",
				PostedAt:    time.Now(),
			},
			keywords: []string{"golang"},
			want:     false,
		},
		{
			name: "job with exclude keyword in description should not match",
			job: &model.Job{
				ID:          "3",
				Title:       "Golang Developer",
				Description: "Looking for lowest bid developer",
				PostedAt:    time.Now(),
			},
			keywords: []string{"golang"},
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Match(tt.job, tt.keywords)
			got := len(result) > 0
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Match_Budget(t *testing.T) {
	cfg := config.FilterConfig{
		Budget:  config.BudgetFilter{Min: 100, Max: 1000},
		JobType: config.JobTypeAll,
	}
	f := New(cfg)

	tests := []struct {
		name     string
		job      *model.Job
		keywords []string
		want     bool
	}{
		{
			name: "job within budget range should match",
			job: &model.Job{
				ID:          "1",
				Title:       "Golang Developer",
				Description: "Need golang developer",
				JobType:     model.JobTypeFixed,
				BudgetMin:   floatPtr(200),
				BudgetMax:   floatPtr(500),
				PostedAt:    time.Now(),
			},
			keywords: []string{"golang"},
			want:     true,
		},
		{
			name: "job below min budget should not match",
			job: &model.Job{
				ID:          "2",
				Title:       "Golang Developer",
				Description: "Need golang developer",
				JobType:     model.JobTypeFixed,
				BudgetMax:   floatPtr(50),
				PostedAt:    time.Now(),
			},
			keywords: []string{"golang"},
			want:     false,
		},
		{
			name: "job above max budget should not match",
			job: &model.Job{
				ID:          "3",
				Title:       "Golang Developer",
				Description: "Need golang developer",
				JobType:     model.JobTypeFixed,
				BudgetMin:   floatPtr(2000),
				PostedAt:    time.Now(),
			},
			keywords: []string{"golang"},
			want:     false,
		},
		{
			name: "job without budget info should match",
			job: &model.Job{
				ID:          "4",
				Title:       "Golang Developer",
				Description: "Need golang developer",
				JobType:     model.JobTypeFixed,
				PostedAt:    time.Now(),
			},
			keywords: []string{"golang"},
			want:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := f.Match(tt.job, tt.keywords)
			got := len(result) > 0
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Match_JobType(t *testing.T) {
	tests := []struct {
		name       string
		filterType config.JobType
		jobType    model.JobType
		want       bool
	}{
		{"all accepts fixed", config.JobTypeAll, model.JobTypeFixed, true},
		{"all accepts hourly", config.JobTypeAll, model.JobTypeHourly, true},
		{"fixed accepts fixed", config.JobTypeFixed, model.JobTypeFixed, true},
		{"fixed rejects hourly", config.JobTypeFixed, model.JobTypeHourly, false},
		{"hourly accepts hourly", config.JobTypeHourly, model.JobTypeHourly, true},
		{"hourly rejects fixed", config.JobTypeHourly, model.JobTypeFixed, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := config.FilterConfig{
				Budget:  config.BudgetFilter{Min: 0, Max: 100000},
				JobType: tt.filterType,
			}
			f := New(cfg)

			job := &model.Job{
				ID:          "1",
				Title:       "Golang Developer",
				Description: "Need golang developer",
				JobType:     tt.jobType,
				PostedAt:    time.Now(),
			}

			result := f.Match(job, []string{"golang"})
			got := len(result) > 0
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Match_Proposals(t *testing.T) {
	maxProposals := 10
	cfg := config.FilterConfig{
		Budget:       config.BudgetFilter{Min: 0, Max: 100000},
		JobType:      config.JobTypeAll,
		MaxProposals: &maxProposals,
	}
	f := New(cfg)

	tests := []struct {
		name      string
		proposals *int
		want      bool
	}{
		{"under max proposals should match", intPtr(5), true},
		{"at max proposals should match", intPtr(10), true},
		{"over max proposals should not match", intPtr(15), false},
		{"no proposals info should match", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &model.Job{
				ID:          "1",
				Title:       "Golang Developer",
				Description: "Need golang developer",
				JobType:     model.JobTypeFixed,
				Proposals:   tt.proposals,
				PostedAt:    time.Now(),
			}

			result := f.Match(job, []string{"golang"})
			got := len(result) > 0
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFilter_Match_Keywords(t *testing.T) {
	cfg := config.FilterConfig{
		Budget:  config.BudgetFilter{Min: 0, Max: 100000},
		JobType: config.JobTypeAll,
	}
	f := New(cfg)

	tests := []struct {
		name            string
		title           string
		description     string
		keywords        []string
		wantMatch       bool
		wantMatchedKeys []string
	}{
		{
			name:            "match in title",
			title:           "Golang API Developer",
			description:     "We need a developer",
			keywords:        []string{"golang", "python"},
			wantMatch:       true,
			wantMatchedKeys: []string{"golang"},
		},
		{
			name:            "match in description",
			title:           "Backend Developer",
			description:     "Experience with golang required",
			keywords:        []string{"golang", "python"},
			wantMatch:       true,
			wantMatchedKeys: []string{"golang"},
		},
		{
			name:            "match multiple keywords",
			title:           "Golang API Developer",
			description:     "Build REST API",
			keywords:        []string{"golang", "api", "python"},
			wantMatch:       true,
			wantMatchedKeys: []string{"golang", "api"},
		},
		{
			name:            "case insensitive match",
			title:           "GOLANG Developer",
			description:     "API development",
			keywords:        []string{"golang", "api"},
			wantMatch:       true,
			wantMatchedKeys: []string{"golang", "api"},
		},
		{
			name:            "no match",
			title:           "Python Developer",
			description:     "Django experience required",
			keywords:        []string{"rust", "java"},
			wantMatch:       false,
			wantMatchedKeys: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &model.Job{
				ID:          "1",
				Title:       tt.title,
				Description: tt.description,
				JobType:     model.JobTypeFixed,
				PostedAt:    time.Now(),
			}

			result := f.Match(job, tt.keywords)
			gotMatch := len(result) > 0

			if gotMatch != tt.wantMatch {
				t.Errorf("Match() matched = %v, want %v", gotMatch, tt.wantMatch)
			}

			if tt.wantMatch && len(result) != len(tt.wantMatchedKeys) {
				t.Errorf("Match() returned %d keywords, want %d", len(result), len(tt.wantMatchedKeys))
			}
		})
	}
}

func TestFilter_Match_PostedTime(t *testing.T) {
	cfg := config.FilterConfig{
		Budget:            config.BudgetFilter{Min: 0, Max: 100000},
		JobType:           config.JobTypeAll,
		PostedWithinHours: 24,
	}
	f := New(cfg)

	tests := []struct {
		name     string
		postedAt time.Time
		want     bool
	}{
		{"posted 1 hour ago should match", time.Now().Add(-1 * time.Hour), true},
		{"posted 12 hours ago should match", time.Now().Add(-12 * time.Hour), true},
		{"posted 23 hours ago should match", time.Now().Add(-23 * time.Hour), true},
		{"posted 25 hours ago should not match", time.Now().Add(-25 * time.Hour), false},
		{"posted 48 hours ago should not match", time.Now().Add(-48 * time.Hour), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			job := &model.Job{
				ID:          "1",
				Title:       "Golang Developer",
				Description: "Need golang developer",
				JobType:     model.JobTypeFixed,
				PostedAt:    tt.postedAt,
			}

			result := f.Match(job, []string{"golang"})
			got := len(result) > 0
			if got != tt.want {
				t.Errorf("Match() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
