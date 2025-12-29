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

// Fetch retrieves jobs for the given keywords
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
			log.Error().Int("status", resp.StatusCode).Str("keyword", keyword).Msg("RSS fetch failed")
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
