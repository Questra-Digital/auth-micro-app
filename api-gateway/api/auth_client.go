package api

import (
	"api-gateway/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// AuthClient handles HTTP requests to Auth service
type AuthClient struct {
	client *http.Client
}

// NewAuthClient creates a new Auth client with default timeout
func NewAuthClient() *AuthClient {
	return &AuthClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// GetAccessToken sends request to auth service to get access token
func (ac *AuthClient) GetAccessToken(email string) (*http.Response, error) {
	payload := map[string]string{"email": email}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/getAccessToken", config.AppConfig.AuthorizationService)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create access token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return ac.client.Do(req)
}

// RefreshToken sends request to auth service to refresh access token
func (ac *AuthClient) RefreshToken(refreshToken, email string) (*http.Response, error) {
	payload := map[string]interface{}{
		"grant_type":    "refresh_token",
		"refresh_token": refreshToken,
		"email":         email,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/refreshToken", config.AppConfig.AuthorizationService)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create refresh token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	return ac.client.Do(req)
}

// RefreshAccessToken attempts to refresh the access token using the refresh token
func (ac *AuthClient) RefreshAccessToken(sessionData map[string]string, log interface{}) (string, error) {
	// Get refresh token from session data
	refreshToken, exists := sessionData["refreshTokenID"]
	if !exists || refreshToken == "" {
		return "", fmt.Errorf("no refresh token found in session")
	}

	// Get email from session data
	email, exists := sessionData["email"]
	if !exists || email == "" {
		return "", fmt.Errorf("no email found in session")
	}

	// Call auth service to refresh token
	resp, err := ac.RefreshToken(refreshToken, email)
	if err != nil {
		return "", fmt.Errorf("failed to call auth service: %w", err)
	}

	// Read response body
	respBody, err := ReadResponseBody(resp)
	if err != nil {
		return "", fmt.Errorf("failed to read response body: %w", err)
	}

	// Check if refresh was successful
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("auth service returned status %d: %s", resp.StatusCode, string(respBody))
	}

	// Parse response
	response, err := ParseAuthResponse(respBody)
	if err != nil {
		return "", fmt.Errorf("failed to parse auth response: %w", err)
	}

	// Extract new access token
	newAccessToken, ok := response["access_token"].(string)
	if !ok || newAccessToken == "" {
		return "", fmt.Errorf("no access token in response")
	}

	return newAccessToken, nil
}

// ParseAuthResponse parses the auth service response and extracts tokens
func ParseAuthResponse(respBody []byte) (map[string]interface{}, error) {
	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		return nil, fmt.Errorf("failed to parse auth response: %w", err)
	}
	return response, nil
}

// ExtractTokens extracts access and refresh tokens from auth response
func ExtractTokens(response map[string]interface{}) (string, string, error) {
	accessToken, ok1 := response["access_token"].(string)
	refreshToken, ok2 := response["refresh_token"].(string)

	if !ok1 || !ok2 {
		return "", "", fmt.Errorf("missing tokens in auth response")
	}

	return accessToken, refreshToken, nil
}

// ExtractTokensAndDuration extracts access token, refresh token, and refresh token duration
func ExtractTokensAndDuration(response map[string]interface{}) (string, string, int, error) {
	accessToken, refreshToken, err := ExtractTokens(response)
	if err != nil {
		return "", "", 0, err
	}

	// Extract refresh token duration (default to 7 days if not present)
	refreshTokenDuration := 7 // default value
	if duration, ok := response["refresh_token_duration_days"].(float64); ok {
		refreshTokenDuration = int(duration)
	}

	return accessToken, refreshToken, refreshTokenDuration, nil
}
