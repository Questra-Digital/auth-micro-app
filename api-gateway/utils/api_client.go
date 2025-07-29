package utils

import (
	"api-gateway/config"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIClient handles HTTP requests to other services
type APIClient struct {
	client *http.Client
}

// NewAPIClient creates a new API client with default timeout
func NewAPIClient() *APIClient {
	return &APIClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// RequestOTP sends OTP generation request to OTP service
func (ac *APIClient) RequestOTP(email, sessionID string) (*http.Response, error) {
	payload := map[string]string{"email": email}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/otp/generate", config.AppConfig.OtpService)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sessionID)

	return ac.client.Do(req)
}

// VerifyOTP sends OTP verification request to OTP service
func (ac *APIClient) VerifyOTP(otp, email, sessionID string) (*http.Response, error) {
	payload := map[string]string{
		"otp":   otp,
		"email": email,
	}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/otp/verify", config.AppConfig.OtpService)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP verification request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sessionID)

	return ac.client.Do(req)
}

// GetAccessToken sends request to auth service to get access token
func (ac *APIClient) GetAccessToken(email string) (*http.Response, error) {
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

// ReadResponseBody reads and returns the response body as bytes
func ReadResponseBody(resp *http.Response) ([]byte, error) {
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}
