package config

import "time"

const (
	OTPLength = 6
	OTPTTL    = 5 * time.Minute
	MaxAttempts = 3
	MaxResends = 3
	RateLimitPerMinute = "3-M"
)
