package models

import "github.com/streadway/amqp"

type EmailJob struct {
	To       string `json:"to"`
	Subject  string `json:"subject"`
	HTMLBody string `json:"html_body"`
}

type OTPRequest struct {
	Email string `json:"email" binding:"required,email"`
	OTP   string `json:"otp" binding:"required,len=6"`
}

// RabbitMQ channel shared between producer/consumer
var EmailChannel *amqp.Channel
