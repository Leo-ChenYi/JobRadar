package fetcher

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"jobradar/internal/model"

	"github.com/rs/zerolog/log"
)

const upworkGraphQLURL = "https://api.upwork.com/graphql"

// UpworkAPIFetcher fetches jobs from Upwork GraphQL API
type UpworkAPIFetcher struct {
	client      *http.Client
	accessToken string
}

// NewUpworkAPIFetcher creates a new Upwork API fetcher
func NewUpworkAPIFetcher(accessToken string) *UpworkAPIFetcher {
	return &UpworkAPIFetcher{
		client:      &http.Client{Timeout: 30 * time.Second},
		accessToken: accessToken,
	}
}

// GraphQL request/response structures
type graphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type graphQLResponse struct {
	Data   *jobPostingsData `json:"data"`
	Errors []graphQLError   `json:"errors,omitempty"`
}

type graphQLError struct {
	Message string `json:"message"`
}

type jobPostingsData struct {
	MarketplaceJobPostings *marketplaceJobPostings `json:"marketplaceJobPostings"`
}

type marketplaceJobPostings struct {
	TotalCount int       `json:"totalCount"`
	Edges      []jobEdge `json:"edges"`
	PageInfo   *pageInfo `json:"pageInfo"`
}

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type jobEdge struct {
	Node   jobNode `json:"node"`
	Cursor string  `json:"cursor"`
}

type jobNode struct {
	ID              string      `json:"id"`
	Title           string      `json:"title"`
	Description     string      `json:"description"`
	CreatedDateTime string      `json:"createdDateTime"`
	Skills          []skillNode `json:"skills"`
	Budget          *budgetNode `json:"budget"`
	HourlyBudget    *hourlyNode `json:"hourlyBudget"`
	ContractType    string      `json:"contractType"`
	Client          *clientNode `json:"client"`
	Proposals       *int        `json:"totalApplicants"`
	CipherText      string      `json:"ciphertext"` // Used to construct job URL
}

type skillNode struct {
	Name string `json:"name"`
}

type budgetNode struct {
	Amount       float64 `json:"amount"`
	CurrencyCode string  `json:"currencyCode"`
}

type hourlyNode struct {
	Min float64 `json:"min"`
	Max float64 `json:"max"`
}

type clientNode struct {
	Location *locationNode `json:"location"`
}

type locationNode struct {
	Country string `json:"country"`
}

// buildQuery constructs the GraphQL query for job search
func buildQuery(searchTerm string, limit int) string {
	return fmt.Sprintf(`
query {
  marketplaceJobPostings(
    searchType: USER_JOBS_SEARCH
    searchExpression_eq: "%s"
    sortAttributes: { field: RECENCY }
    pagination: { first: %d }
  ) {
    totalCount
    edges {
      node {
        id
        ciphertext
        title
        description
        createdDateTime
        contractType
        totalApplicants
        skills {
          name
        }
        budget {
          amount
          currencyCode
        }
        hourlyBudget {
          min
          max
        }
        client {
          location {
            country
          }
        }
      }
    }
    pageInfo {
      hasNextPage
      endCursor
    }
  }
}
`, searchTerm, limit)
}

// FetchJobs retrieves jobs from Upwork API for the given search term
func (f *UpworkAPIFetcher) FetchJobs(searchTerm string, limit int) ([]*model.Job, error) {
	if limit <= 0 {
		limit = 50
	}

	query := buildQuery(searchTerm, limit)

	reqBody := graphQLRequest{
		Query: query,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequest("POST", upworkGraphQLURL, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+f.accessToken)

	log.Debug().Str("searchTerm", searchTerm).Msg("Fetching jobs from Upwork API")

	resp, err := f.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	var gqlResp graphQLResponse
	if err := json.Unmarshal(body, &gqlResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(gqlResp.Errors) > 0 {
		return nil, fmt.Errorf("GraphQL error: %s", gqlResp.Errors[0].Message)
	}

	if gqlResp.Data == nil || gqlResp.Data.MarketplaceJobPostings == nil {
		return nil, fmt.Errorf("no data in response")
	}

	// Convert to our Job model
	jobs := make([]*model.Job, 0, len(gqlResp.Data.MarketplaceJobPostings.Edges))
	for _, edge := range gqlResp.Data.MarketplaceJobPostings.Edges {
		job := convertToJob(edge.Node)
		jobs = append(jobs, job)
	}

	log.Info().
		Str("searchTerm", searchTerm).
		Int("count", len(jobs)).
		Int("totalAvailable", gqlResp.Data.MarketplaceJobPostings.TotalCount).
		Msg("Fetched jobs from Upwork API")

	return jobs, nil
}

// convertToJob converts an API response node to our Job model
func convertToJob(node jobNode) *model.Job {
	job := &model.Job{
		ID:          node.ID,
		Title:       node.Title,
		Description: node.Description,
		URL:         buildJobURL(node.CipherText),
		FetchedAt:   time.Now(),
	}

	// Parse created time
	if node.CreatedDateTime != "" {
		if t, err := time.Parse(time.RFC3339, node.CreatedDateTime); err == nil {
			job.PostedAt = t
		} else {
			job.PostedAt = time.Now()
		}
	}

	// Set job type and budget
	if node.ContractType == "HOURLY" || node.HourlyBudget != nil {
		job.JobType = model.JobTypeHourly
		if node.HourlyBudget != nil {
			job.HourlyRateMin = &node.HourlyBudget.Min
			job.HourlyRateMax = &node.HourlyBudget.Max
		}
	} else {
		job.JobType = model.JobTypeFixed
		if node.Budget != nil {
			job.BudgetMin = &node.Budget.Amount
			job.BudgetMax = &node.Budget.Amount
		}
	}

	// Set proposals
	job.Proposals = node.Proposals

	// Set skills
	if len(node.Skills) > 0 {
		job.Skills = make([]string, len(node.Skills))
		for i, skill := range node.Skills {
			job.Skills[i] = skill.Name
		}
	}

	// Set client country
	if node.Client != nil && node.Client.Location != nil {
		job.ClientCountry = node.Client.Location.Country
	}

	return job
}

// buildJobURL constructs the job URL from ciphertext
func buildJobURL(ciphertext string) string {
	if ciphertext == "" {
		return ""
	}
	return fmt.Sprintf("https://www.upwork.com/jobs/%s", ciphertext)
}
