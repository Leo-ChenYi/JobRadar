// +build ignore

package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
)

const (
	authorizeURL = "https://www.upwork.com/ab/account-security/oauth2/authorize"
	tokenURL     = "https://www.upwork.com/api/v3/oauth2/token"
	redirectURI  = "http://localhost:8080/callback"
)

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
}

func main() {
	reader := bufio.NewReader(os.Stdin)

	fmt.Println("==============================================")
	fmt.Println("   Upwork OAuth 2.0 Helper for JobRadar")
	fmt.Println("==============================================")
	fmt.Println()

	// Get Client ID
	fmt.Print("Enter your Client ID (API Key): ")
	clientID, _ := reader.ReadString('\n')
	clientID = strings.TrimSpace(clientID)

	// Get Client Secret
	fmt.Print("Enter your Client Secret: ")
	clientSecret, _ := reader.ReadString('\n')
	clientSecret = strings.TrimSpace(clientSecret)

	// Build authorization URL
	authURL := fmt.Sprintf("%s?client_id=%s&response_type=code&redirect_uri=%s",
		authorizeURL,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURI))

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("Step 1: Open this URL in your browser:")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println(authURL)
	fmt.Println()
	fmt.Println("1. Login to Upwork if needed")
	fmt.Println("2. Click 'Authorize' to grant access")
	fmt.Println("3. You'll be redirected to localhost (may show error - that's OK)")
	fmt.Println("4. Copy the 'code' parameter from the URL")
	fmt.Println()

	// Get authorization code
	fmt.Print("Enter the authorization code from the URL: ")
	authCode, _ := reader.ReadString('\n')
	authCode = strings.TrimSpace(authCode)

	// Exchange code for token
	fmt.Println()
	fmt.Println("Exchanging code for access token...")

	token, err := exchangeCodeForToken(clientID, clientSecret, authCode)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("Success! Here are your tokens:")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Printf("Access Token:  %s\n", token.AccessToken)
	fmt.Printf("Refresh Token: %s\n", token.RefreshToken)
	fmt.Printf("Expires In:    %d seconds\n", token.ExpiresIn)
	fmt.Println()
	fmt.Println("==============================================")
	fmt.Println("Next Steps:")
	fmt.Println("==============================================")
	fmt.Println()
	fmt.Println("1. Set the environment variable:")
	fmt.Println()
	fmt.Printf("   export UPWORK_ACCESS_TOKEN=\"%s\"\n", token.AccessToken)
	fmt.Println()
	fmt.Println("2. Save your refresh token for later (to get new access tokens):")
	fmt.Println()
	fmt.Printf("   Refresh Token: %s\n", token.RefreshToken)
	fmt.Println()
	fmt.Println("3. Test the configuration:")
	fmt.Println()
	fmt.Println("   jobradar validate")
	fmt.Println("   jobradar check --verbose")
	fmt.Println()
}

func exchangeCodeForToken(clientID, clientSecret, code string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("client_id", clientID)
	data.Set("client_secret", clientSecret)
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)

	resp, err := http.PostForm(tokenURL, data)
	if err != nil {
		return nil, fmt.Errorf("failed to request token: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token request failed (%d): %s", resp.StatusCode, string(body))
	}

	var token TokenResponse
	if err := json.Unmarshal(body, &token); err != nil {
		return nil, fmt.Errorf("failed to parse response: %w", err)
	}

	return &token, nil
}

