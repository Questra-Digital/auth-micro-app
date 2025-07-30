package api

import (
	"api-gateway/config"
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// OTPClient handles HTTP requests to OTP service
type OTPClient struct {
	client *http.Client
}

// NewOTPClient creates a new OTP client with default timeout
func NewOTPClient() *OTPClient {
	return &OTPClient{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
	}
}

// RequestOTP sends OTP generation request to OTP service
func (oc *OTPClient) RequestOTP(email, sessionID string) (*http.Response, error) {
	payload := map[string]string{"email": email}
	body, _ := json.Marshal(payload)

	url := fmt.Sprintf("%s/otp/generate", config.AppConfig.OtpService)
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create OTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Session-ID", sessionID)

	return oc.client.Do(req)
}

// VerifyOTP sends OTP verification request to OTP service
func (oc *OTPClient) VerifyOTP(otp, email, sessionID string) (*http.Response, error) {
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

	return oc.client.Do(req)
}
