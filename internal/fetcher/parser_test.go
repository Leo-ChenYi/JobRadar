package fetcher

import (
	"testing"
	"time"

	"jobradar/internal/model"

	"github.com/mmcdole/gofeed"
)

func TestExtractJobID(t *testing.T) {
	tests := []struct {
		name string
		url  string
		want string
	}{
		{
			name: "standard job URL",
			url:  "https://www.upwork.com/jobs/~01abc123def456",
			want: "~01abc123def456",
		},
		{
			name: "job URL with query params",
			url:  "https://www.upwork.com/jobs/~01xyz789?source=rss",
			want: "~01xyz789",
		},
		{
			name: "URL without job ID",
			url:  "https://www.upwork.com/other/page",
			want: "https://www.upwork.com/other/page",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := extractJobID(tt.url)
			if got != tt.want {
				t.Errorf("extractJobID() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseBudgetInfo(t *testing.T) {
	tests := []struct {
		name          string
		description   string
		wantJobType   model.JobType
		wantBudgetMin *float64
		wantBudgetMax *float64
		wantHourlyMin *float64
		wantHourlyMax *float64
	}{
		{
			name:          "fixed budget range",
			description:   "Budget: $100-$500",
			wantJobType:   model.JobTypeFixed,
			wantBudgetMin: floatPtr(100),
			wantBudgetMax: floatPtr(500),
		},
		{
			name:          "fixed budget single value",
			description:   "Budget: $300",
			wantJobType:   model.JobTypeFixed,
			wantBudgetMin: floatPtr(300),
			wantBudgetMax: floatPtr(300),
		},
		{
			name:          "hourly rate range",
			description:   "Hourly Range: $25-$50",
			wantJobType:   model.JobTypeHourly,
			wantHourlyMin: floatPtr(25),
			wantHourlyMax: floatPtr(50),
		},
		{
			name:          "budget with commas",
			description:   "Budget: $1,000-$5,000",
			wantJobType:   model.JobTypeFixed,
			wantBudgetMin: floatPtr(1000),
			wantBudgetMax: floatPtr(5000),
		},
		{
			name:        "no budget info",
			description: "Looking for a developer",
			wantJobType: model.JobTypeFixed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			jobType, budgetMin, budgetMax, hourlyMin, hourlyMax := parseBudgetInfo(tt.description)

			if jobType != tt.wantJobType {
				t.Errorf("jobType = %v, want %v", jobType, tt.wantJobType)
			}

			if !floatPtrEqual(budgetMin, tt.wantBudgetMin) {
				t.Errorf("budgetMin = %v, want %v", ptrValue(budgetMin), ptrValue(tt.wantBudgetMin))
			}

			if !floatPtrEqual(budgetMax, tt.wantBudgetMax) {
				t.Errorf("budgetMax = %v, want %v", ptrValue(budgetMax), ptrValue(tt.wantBudgetMax))
			}

			if !floatPtrEqual(hourlyMin, tt.wantHourlyMin) {
				t.Errorf("hourlyMin = %v, want %v", ptrValue(hourlyMin), ptrValue(tt.wantHourlyMin))
			}

			if !floatPtrEqual(hourlyMax, tt.wantHourlyMax) {
				t.Errorf("hourlyMax = %v, want %v", ptrValue(hourlyMax), ptrValue(tt.wantHourlyMax))
			}
		})
	}
}

func TestParseSkills(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        []string
	}{
		{
			name:        "multiple skills",
			description: "Skills: Golang, REST API, Microservices",
			want:        []string{"Golang", "REST API", "Microservices"},
		},
		{
			name:        "single skill",
			description: "Skills: Python",
			want:        []string{"Python"},
		},
		{
			name:        "skills with extra spaces",
			description: "Skills:   Go ,  Docker  ,  Kubernetes  ",
			want:        []string{"Go", "Docker", "Kubernetes"},
		},
		{
			name:        "no skills",
			description: "Looking for a developer",
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseSkills(tt.description)

			if len(got) != len(tt.want) {
				t.Errorf("parseSkills() returned %d skills, want %d", len(got), len(tt.want))
				return
			}

			for i, skill := range got {
				if skill != tt.want[i] {
					t.Errorf("parseSkills()[%d] = %v, want %v", i, skill, tt.want[i])
				}
			}
		})
	}
}

func TestParseProposals(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        *int
	}{
		{
			name:        "proposals present",
			description: "Proposals: 15",
			want:        intPtr(15),
		},
		{
			name:        "zero proposals",
			description: "Proposals: 0",
			want:        intPtr(0),
		},
		{
			name:        "no proposals info",
			description: "Looking for a developer",
			want:        nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseProposals(tt.description)

			if !intPtrEqual(got, tt.want) {
				t.Errorf("parseProposals() = %v, want %v", ptrValueInt(got), ptrValueInt(tt.want))
			}
		})
	}
}

func TestCleanDescription(t *testing.T) {
	tests := []struct {
		name        string
		description string
		want        string
	}{
		{
			name:        "remove HTML tags",
			description: "<p>Hello <b>World</b></p>",
			want:        "Hello World",
		},
		{
			name:        "replace HTML entities",
			description: "Tom &amp; Jerry &lt;3 &gt;",
			want:        "Tom & Jerry <3 >",
		},
		{
			name:        "normalize whitespace",
			description: "Hello    World   Test",
			want:        "Hello World Test",
		},
		{
			name:        "complex HTML",
			description: "<br/><b>Budget</b>: $500<br/><b>Skills</b>: Go, Python",
			want:        "Budget : $500 Skills : Go, Python",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := cleanDescription(tt.description)
			if got != tt.want {
				t.Errorf("cleanDescription() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestParseRSSItem(t *testing.T) {
	pubTime := time.Now().Add(-2 * time.Hour)

	item := &gofeed.Item{
		Title:           "Golang API Developer Needed",
		Link:            "https://www.upwork.com/jobs/~01abc123",
		Description:     "Budget: $500-$1000<br/>Skills: Golang, REST API<br/>Looking for experienced developer",
		PublishedParsed: &pubTime,
	}

	job := ParseRSSItem(item)

	if job.ID != "~01abc123" {
		t.Errorf("ID = %v, want ~01abc123", job.ID)
	}

	if job.Title != "Golang API Developer Needed" {
		t.Errorf("Title = %v, want 'Golang API Developer Needed'", job.Title)
	}

	if job.URL != "https://www.upwork.com/jobs/~01abc123" {
		t.Errorf("URL = %v, want 'https://www.upwork.com/jobs/~01abc123'", job.URL)
	}

	if job.JobType != model.JobTypeFixed {
		t.Errorf("JobType = %v, want fixed", job.JobType)
	}

	if job.BudgetMin == nil || *job.BudgetMin != 500 {
		t.Errorf("BudgetMin = %v, want 500", ptrValue(job.BudgetMin))
	}

	if job.BudgetMax == nil || *job.BudgetMax != 1000 {
		t.Errorf("BudgetMax = %v, want 1000", ptrValue(job.BudgetMax))
	}

	if len(job.Skills) != 2 {
		t.Errorf("Skills count = %d, want 2", len(job.Skills))
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}

func floatPtrEqual(a, b *float64) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func intPtrEqual(a, b *int) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	return *a == *b
}

func ptrValue(p *float64) string {
	if p == nil {
		return "nil"
	}
	return string(rune(*p))
}

func ptrValueInt(p *int) string {
	if p == nil {
		return "nil"
	}
	return string(rune(*p))
}
