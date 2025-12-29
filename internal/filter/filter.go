package filter

import (
	"strings"
	"time"

	"jobradar/internal/config"
	"jobradar/internal/model"

	"github.com/rs/zerolog/log"
)

// Filter handles job filtering based on configuration
type Filter struct {
	config config.FilterConfig
}

// New creates a new Filter instance
func New(cfg config.FilterConfig) *Filter {
	return &Filter{config: cfg}
}

// Match checks if a job matches the filter criteria
// Returns matched keywords if the job passes all filters, nil otherwise
func (f *Filter) Match(job *model.Job, keywords []string) []string {
	// 1. Check exclude keywords
	if f.hasExcludeKeywords(job) {
		log.Debug().Str("job", job.ID).Msg("Excluded by keywords")
		return nil
	}

	// 2. Check budget
	if !f.checkBudget(job) {
		log.Debug().Str("job", job.ID).Msg("Excluded by budget")
		return nil
	}

	// 3. Check job type
	if !f.checkJobType(job) {
		log.Debug().Str("job", job.ID).Msg("Excluded by job type")
		return nil
	}

	// 4. Check proposals count
	if !f.checkProposals(job) {
		log.Debug().Str("job", job.ID).Msg("Excluded by proposals count")
		return nil
	}

	// 5. Check posted time
	if !f.checkPostedTime(job) {
		log.Debug().Str("job", job.ID).Msg("Excluded by posted time")
		return nil
	}

	// 6. Check keyword match
	matched := f.matchKeywords(job, keywords)
	if len(matched) == 0 {
		log.Debug().Str("job", job.ID).Msg("No keyword match")
		return nil
	}

	return matched
}

// hasExcludeKeywords checks if job contains any exclude keywords
func (f *Filter) hasExcludeKeywords(job *model.Job) bool {
	text := strings.ToLower(job.Title + " " + job.Description)

	for _, keyword := range f.config.ExcludeKeywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			log.Debug().Str("job", job.ID).Str("exclude", keyword).Msg("Found exclude keyword")
			return true
		}
	}

	return false
}

// checkBudget verifies the job budget is within range
func (f *Filter) checkBudget(job *model.Job) bool {
	var budget float64

	if job.JobType == model.JobTypeFixed {
		if job.BudgetMax != nil {
			budget = *job.BudgetMax
		} else if job.BudgetMin != nil {
			budget = *job.BudgetMin
		} else {
			// No budget info, allow by default
			return true
		}
	} else {
		// For hourly, check hourly rate
		if job.HourlyRateMax != nil {
			budget = *job.HourlyRateMax
		} else if job.HourlyRateMin != nil {
			budget = *job.HourlyRateMin
		} else {
			return true
		}
	}

	if budget < float64(f.config.Budget.Min) {
		return false
	}

	if f.config.Budget.Max > 0 && budget > float64(f.config.Budget.Max) {
		return false
	}

	return true
}

// checkJobType verifies the job type matches configuration
func (f *Filter) checkJobType(job *model.Job) bool {
	if f.config.JobType == config.JobTypeAll {
		return true
	}

	if f.config.JobType == config.JobTypeFixed && job.JobType == model.JobTypeFixed {
		return true
	}

	if f.config.JobType == config.JobTypeHourly && job.JobType == model.JobTypeHourly {
		return true
	}

	return false
}

// checkProposals verifies the proposal count is acceptable
func (f *Filter) checkProposals(job *model.Job) bool {
	if f.config.MaxProposals == nil {
		return true
	}

	if job.Proposals == nil {
		// No proposal info, allow by default
		return true
	}

	return *job.Proposals <= *f.config.MaxProposals
}

// checkPostedTime verifies the job was posted within the configured time window
func (f *Filter) checkPostedTime(job *model.Job) bool {
	if f.config.PostedWithinHours <= 0 {
		return true
	}

	cutoff := time.Now().Add(-time.Duration(f.config.PostedWithinHours) * time.Hour)
	return job.PostedAt.After(cutoff)
}

// matchKeywords finds matching keywords in job title and description
func (f *Filter) matchKeywords(job *model.Job, keywords []string) []string {
	text := strings.ToLower(job.Title + " " + job.Description)
	var matched []string

	for _, keyword := range keywords {
		if strings.Contains(text, strings.ToLower(keyword)) {
			matched = append(matched, keyword)
		}
	}

	return matched
}
