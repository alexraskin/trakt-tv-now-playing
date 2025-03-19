package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alexraskin/trakt-tv-now-playing/models"
	"github.com/alexraskin/trakt-tv-now-playing/utils"
)

func makeRequest(reqBody string, timeout time.Duration) (*models.TraktTokenResponse, error) {

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", models.DeviceTokenURL, strings.NewReader(reqBody))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := http.DefaultClient
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusOK {
		var tokenResp models.TraktTokenResponse
		if err := json.Unmarshal(body, &tokenResp); err != nil {
			return nil, err
		}

		models.AccessToken = tokenResp.AccessToken
		models.RefreshToken = tokenResp.RefreshToken
		models.TokenExpiry = time.Now().Add(time.Duration(tokenResp.ExpiresIn) * time.Second)

		if err := utils.SaveToken(); err != nil {
			return nil, err
		}

		fmt.Printf("Token will expire at: %s\n", models.TokenExpiry.Format(time.RFC3339))
		return &tokenResp, nil
	}

	if resp.StatusCode == http.StatusBadRequest {
		var errorResp struct {
			Error string `json:"error"`
		}
		if err := json.Unmarshal(body, &errorResp); err != nil {
			return nil, fmt.Errorf("error parsing error response: %v", err)
		}
		if errorResp.Error != "" {
			return nil, fmt.Errorf("request failed: %s", errorResp.Error)
		}
	}

	return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
}

func PollForAuthorization(interval, expiresIn int) {
	maxAttempts := expiresIn / interval

	fmt.Printf("Starting authorization polling (will try for up to %d minutes)...\n", expiresIn/60)

	for attempt := range maxAttempts {
		time.Sleep(time.Duration(interval) * time.Second)

		reqBody := fmt.Sprintf("code=%s&client_id=%s&client_secret=%s",
			models.DeviceCode, models.Credentials.ClientID, models.Credentials.ClientSecret)

		fmt.Printf("Polling attempt %d/%d...\n", attempt+1, maxAttempts)
		tokenResp, err := makeRequest(reqBody, 10*time.Second)
		if err != nil {
			if strings.Contains(err.Error(), "authorization pending") {
				fmt.Println("Waiting for user to authorize...")
				continue
			}
			fmt.Printf("Error: %v\n", err)
			continue
		}

		if tokenResp != nil {
			fmt.Println("Successfully authorized with Trakt.tv")
			return
		}
	}

	fmt.Println("Authorization polling timed out")
}

func RefreshAccessToken() error {
	if models.RefreshToken == "" {
		return fmt.Errorf("no refresh token available")
	}

	reqBody := fmt.Sprintf("refresh_token=%s&client_id=%s&client_secret=%s&grant_type=refresh_token&redirect_uri=urn:ietf:wg:oauth:2.0:oob",
		models.RefreshToken, models.Credentials.ClientID, models.Credentials.ClientSecret)

	_, err := makeRequest(reqBody, 10*time.Second)
	if err != nil {
		return fmt.Errorf("failed to refresh token: %v", err)
	}

	fmt.Println("Successfully refreshed token!")
	return nil
}
