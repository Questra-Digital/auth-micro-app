package models

type RateLimitParams struct {
	Endpoint     string
	Method       string
	RateLimit    string
	RequestCount int
	Limit        int
	Window       string
	Blocked      bool
}