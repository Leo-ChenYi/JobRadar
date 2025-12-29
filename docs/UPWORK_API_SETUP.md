# Upwork API Setup Guide

This guide walks you through setting up Upwork API access for JobRadar.

## Overview

Since Upwork discontinued RSS feeds in August 2024, JobRadar now uses the official Upwork GraphQL API to fetch job listings.

## Step 1: Apply for API Access

1. **Login to Upwork** with your account

2. **Go to Developer Portal**:
   - Visit: https://www.upwork.com/developer/keys/apply
   - Or navigate: Profile → Settings → Developer → API Keys

3. **Fill out the application**:
   - **Application Name**: JobRadar (or your preferred name)
   - **Application Description**: Personal tool for monitoring job postings
   - **Callback URL**: `http://localhost:8080/callback` (for OAuth flow)
   - **Permissions Required**: 
     - ✅ Read marketplace Job Postings - Public

4. **Submit and wait for approval** (usually 1-3 business days)

## Step 2: Get Your API Credentials

Once approved, you'll receive:
- **Client ID** (API Key)
- **Client Secret**

Save these securely - you'll need them for the OAuth flow.

## Step 3: Complete OAuth 2.0 Flow

Upwork uses OAuth 2.0 for authentication. You need to get an **Access Token**.

### Option A: Use the OAuth Helper Script

We provide a helper script to simplify the OAuth flow:

```bash
# Run the OAuth helper
go run scripts/oauth_helper.go

# Follow the prompts:
# 1. Enter your Client ID
# 2. Enter your Client Secret
# 3. Open the URL in browser and authorize
# 4. Copy the authorization code
# 5. Get your access token
```

### Option B: Manual OAuth Flow

1. **Authorization URL** - Open in browser:
   ```
   https://www.upwork.com/ab/account-security/oauth2/authorize?
     client_id=YOUR_CLIENT_ID&
     response_type=code&
     redirect_uri=http://localhost:8080/callback
   ```

2. **Authorize the application** in Upwork

3. **Get the authorization code** from the redirect URL:
   ```
   http://localhost:8080/callback?code=AUTHORIZATION_CODE
   ```

4. **Exchange code for access token**:
   ```bash
   curl -X POST https://www.upwork.com/api/v3/oauth2/token \
     -d "grant_type=authorization_code" \
     -d "client_id=YOUR_CLIENT_ID" \
     -d "client_secret=YOUR_CLIENT_SECRET" \
     -d "code=AUTHORIZATION_CODE" \
     -d "redirect_uri=http://localhost:8080/callback"
   ```

5. **Save the access token** from the response

## Step 4: Configure JobRadar

1. **Set environment variable**:
   ```bash
   export UPWORK_ACCESS_TOKEN="your_access_token_here"
   ```

2. **Or add to config.yaml directly** (not recommended for security):
   ```yaml
   upwork_api:
     enabled: true
     access_token: "your_access_token_here"
   ```

3. **Verify configuration**:
   ```bash
   jobradar validate
   ```

## Step 5: Test the API Connection

```bash
# Run a check to test the API
jobradar check --verbose
```

If successful, you should see job listings being fetched.

## Troubleshooting

### "401 Unauthorized" Error
- Your access token may have expired
- Re-run the OAuth flow to get a new token
- Upwork tokens typically expire after 2 weeks

### "403 Forbidden" Error
- Your API key may not have the required permissions
- Check that "Read marketplace Job Postings - Public" is enabled
- Contact Upwork support if needed

### "Rate Limit Exceeded" Error
- Upwork limits API calls (typically 100 calls/hour)
- Increase the `interval_minutes` in your config
- Reduce the number of searches or keywords

### No Jobs Returned
- Check your search keywords are valid
- Try broader search terms
- Verify the API is working with a simple query

## Token Refresh

Upwork access tokens expire. To refresh:

1. **If you have a refresh token**:
   ```bash
   curl -X POST https://www.upwork.com/api/v3/oauth2/token \
     -d "grant_type=refresh_token" \
     -d "client_id=YOUR_CLIENT_ID" \
     -d "client_secret=YOUR_CLIENT_SECRET" \
     -d "refresh_token=YOUR_REFRESH_TOKEN"
   ```

2. **If no refresh token**, re-run the full OAuth flow

## API Rate Limits

Upwork enforces rate limits:
- **100 requests per hour** (typical)
- **1000 requests per day** (typical)

JobRadar is designed to stay within these limits with default settings.

## GraphQL Query Reference

JobRadar uses this query to fetch jobs:

```graphql
query {
  marketplaceJobPostings(
    searchType: USER_JOBS_SEARCH
    searchExpression_eq: "golang"
    sortAttributes: { field: RECENCY }
    pagination: { first: 50 }
  ) {
    totalCount
    edges {
      node {
        id
        title
        description
        createdDateTime
        skills { name }
        budget { amount }
        hourlyBudget { min max }
        client { location { country } }
      }
    }
  }
}
```

You can customize this in `internal/fetcher/upwork_api.go`.

## Security Notes

- **Never commit your access token** to version control
- **Use environment variables** for sensitive data
- **Rotate tokens regularly** for security
- **Monitor API usage** in Upwork developer dashboard

## Resources

- [Upwork API Documentation](https://www.upwork.com/developer/documentation/graphql/api/docs/index.html)
- [OAuth 2.0 Guide](https://www.upwork.com/developer/documentation/graphql/api/docs/index.html#authentication)
- [GraphQL Reference](https://www.upwork.com/developer/documentation/graphql/api/docs/index.html)

