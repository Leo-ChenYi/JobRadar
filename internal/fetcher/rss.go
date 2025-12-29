package fetcher

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"jobradar/internal/model"

	"github.com/mmcdole/gofeed"
	"github.com/rs/zerolog/log"
)

const baseURL = "https://www.upwork.com/ab/feed/jobs/rss"

// RSSFetcher fetches jobs from Upwork RSS feeds
type RSSFetcher struct {
	client *http.Client
	parser *gofeed.Parser
}

// NewRSSFetcher creates a new RSS fetcher
func NewRSSFetcher() *RSSFetcher {
	return &RSSFetcher{
		client: &http.Client{Timeout: 30 * time.Second},
		parser: gofeed.NewParser(),
	}
}

// FetchFromURL retrieves jobs from a direct RSS URL (recommended method)
func (f *RSSFetcher) FetchFromURL(feedURL string) ([]*model.Job, error) {
	var jobs []*model.Job

	log.Debug().Str("url", feedURL).Msg("Fetching RSS feed from URL")

	resp, err := f.client.Get(feedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch RSS: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("RSS fetch failed with status %d", resp.StatusCode)
	}

	feed, err := f.parser.Parse(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse RSS: %w", err)
	}

	seen := make(map[string]bool)
	for _, item := range feed.Items {
		job := ParseRSSItem(item)

		// Deduplicate within this batch
		if !seen[job.ID] {
			seen[job.ID] = true
			jobs = append(jobs, job)
		}
	}

	log.Debug().Int("count", len(jobs)).Msg("Fetched jobs from URL")
	return jobs, nil
}

// Fetch retrieves jobs for the given keywords (DEPRECATED: Upwork no longer supports public RSS)
// Use FetchFromURL with authenticated RSS URLs instead
func (f *RSSFetcher) Fetch(keywords []string) ([]*model.Job, error) {
	var jobs []*model.Job
	seen := make(map[string]bool)

	for _, keyword := range keywords {
		feedURL := f.buildURL(keyword)
		log.Debug().Str("keyword", keyword).Str("url", feedURL).Msg("Fetching RSS feed")

		resp, err := f.client.Get(feedURL)
		if err != nil {
			log.Error().Err(err).Str("keyword", keyword).Msg("Failed to fetch RSS")
			continue
		}

		if resp.StatusCode != http.StatusOK {
			resp.Body.Close()
			log.Error().Int("status", resp.StatusCode).Str("keyword", keyword).Msg("RSS fetch failed - public RSS is deprecated, use rss_feeds config instead")
			continue
		}

		feed, err := f.parser.Parse(resp.Body)
		resp.Body.Close()
		if err != nil {
			log.Error().Err(err).Str("keyword", keyword).Msg("Failed to parse RSS")
			continue
		}

		for _, item := range feed.Items {
			job := ParseRSSItem(item)

			// Deduplicate within this batch
			if !seen[job.ID] {
				seen[job.ID] = true
				jobs = append(jobs, job)
			}
		}

		log.Debug().Str("keyword", keyword).Int("count", len(feed.Items)).Msg("Fetched jobs")
	}

	return jobs, nil
}

// buildURL constructs the Upwork RSS URL for a keyword
func (f *RSSFetcher) buildURL(keyword string) string {
	params := url.Values{}
	params.Set("q", keyword)
	params.Set("sort", "recency")

	return fmt.Sprintf("%s?%s", baseURL, params.Encode())
}
