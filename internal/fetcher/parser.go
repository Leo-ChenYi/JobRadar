package fetcher

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"jobradar/internal/model"

	"github.com/mmcdole/gofeed"
)

var (
	// Regex patterns for parsing RSS description
	jobIDRegex      = regexp.MustCompile(`/jobs/(~\w+)`)
	budgetRegex     = regexp.MustCompile(`(?i)Budget[:\s]*\$?([\d,]+)(?:\s*-\s*\$?([\d,]+))?`)
	hourlyRegex     = regexp.MustCompile(`(?i)Hourly Range[:\s]*\$?([\d.]+)\s*-\s*\$?([\d.]+)`)
	skillsRegex     = regexp.MustCompile(`(?i)Skills[:\s]*([^<]+)`)
	countryRegex    = regexp.MustCompile(`(?i)Country[:\s]*([^<]+)`)
	proposalsRegex  = regexp.MustCompile(`(?i)Proposals[:\s]*(\d+)`)
	htmlTagRegex    = regexp.MustCompile(`<[^>]+>`)
	whitespaceRegex = regexp.MustCompile(`\s+`)
)

// ParseRSSItem converts an RSS item to a Job object
func ParseRSSItem(item *gofeed.Item) *model.Job {
	title := item.Title
	link := item.Link
	description := item.Description
	pubDate := item.PublishedParsed

	// Extract job ID from link
	jobID := extractJobID(link)

	// Parse job type and budget
	jobType, budgetMin, budgetMax, hourlyMin, hourlyMax := parseBudgetInfo(description)

	// Parse other fields
	skills := parseSkills(description)
	proposals := parseProposals(description)
	country := parseCountry(description)

	postedAt := time.Now()
	if pubDate != nil {
		postedAt = *pubDate
	}

	return &model.Job{
		ID:            jobID,
		Title:         cleanText(title),
		Description:   cleanDescription(description),
		URL:           link,
		JobType:       jobType,
		BudgetMin:     budgetMin,
		BudgetMax:     budgetMax,
		HourlyRateMin: hourlyMin,
		HourlyRateMax: hourlyMax,
		Proposals:     proposals,
		ClientCountry: country,
		Skills:        skills,
		PostedAt:      postedAt,
		FetchedAt:     time.Now(),
	}
}

// extractJobID extracts the job ID from the URL
func extractJobID(url string) string {
	matches := jobIDRegex.FindStringSubmatch(url)
	if len(matches) > 1 {
		return matches[1]
	}
	// Fallback: use URL hash
	return url
}

// parseBudgetInfo extracts budget/hourly rate information
func parseBudgetInfo(description string) (model.JobType, *float64, *float64, *float64, *float64) {
	// Check for hourly rate first
	hourlyMatches := hourlyRegex.FindStringSubmatch(description)
	if len(hourlyMatches) >= 3 {
		minRate := parseFloat(hourlyMatches[1])
		maxRate := parseFloat(hourlyMatches[2])
		return model.JobTypeHourly, nil, nil, &minRate, &maxRate
	}

	// Check for fixed budget
	budgetMatches := budgetRegex.FindStringSubmatch(description)
	if len(budgetMatches) >= 2 {
		minBudget := parseFloat(budgetMatches[1])
		maxBudget := minBudget
		if len(budgetMatches) > 2 && budgetMatches[2] != "" {
			maxBudget = parseFloat(budgetMatches[2])
		}
		return model.JobTypeFixed, &minBudget, &maxBudget, nil, nil
	}

	// Default to fixed if no budget info found
	return model.JobTypeFixed, nil, nil, nil, nil
}

// parseSkills extracts skill tags from description
func parseSkills(description string) []string {
	matches := skillsRegex.FindStringSubmatch(description)
	if len(matches) < 2 {
		return nil
	}

	skillsStr := strings.TrimSpace(matches[1])
	// Remove HTML entities and clean up
	skillsStr = strings.ReplaceAll(skillsStr, "&nbsp;", " ")
	skillsStr = htmlTagRegex.ReplaceAllString(skillsStr, "")

	skills := strings.Split(skillsStr, ",")

	result := make([]string, 0, len(skills))
	for _, s := range skills {
		if trimmed := strings.TrimSpace(s); trimmed != "" {
			result = append(result, trimmed)
		}
	}
	return result
}

// parseProposals extracts the number of proposals
func parseProposals(description string) *int {
	matches := proposalsRegex.FindStringSubmatch(description)
	if len(matches) < 2 {
		return nil
	}
	count, err := strconv.Atoi(matches[1])
	if err != nil {
		return nil
	}
	return &count
}

// parseCountry extracts the client's country
func parseCountry(description string) string {
	matches := countryRegex.FindStringSubmatch(description)
	if len(matches) < 2 {
		return ""
	}
	country := strings.TrimSpace(matches[1])
	country = htmlTagRegex.ReplaceAllString(country, "")
	return strings.TrimSpace(country)
}

// cleanDescription removes HTML tags and normalizes whitespace
func cleanDescription(description string) string {
	// Remove HTML tags
	clean := htmlTagRegex.ReplaceAllString(description, " ")
	// Replace HTML entities
	clean = strings.ReplaceAll(clean, "&nbsp;", " ")
	clean = strings.ReplaceAll(clean, "&amp;", "&")
	clean = strings.ReplaceAll(clean, "&lt;", "<")
	clean = strings.ReplaceAll(clean, "&gt;", ">")
	clean = strings.ReplaceAll(clean, "&quot;", "\"")
	// Normalize whitespace
	clean = whitespaceRegex.ReplaceAllString(clean, " ")
	return strings.TrimSpace(clean)
}

// cleanText cleans a simple text field
func cleanText(text string) string {
	text = strings.ReplaceAll(text, "&amp;", "&")
	text = strings.ReplaceAll(text, "&nbsp;", " ")
	return strings.TrimSpace(text)
}

// parseFloat parses a string to float64, handling commas
func parseFloat(s string) float64 {
	s = strings.ReplaceAll(s, ",", "")
	s = strings.TrimSpace(s)
	f, _ := strconv.ParseFloat(s, 64)
	return f
}
