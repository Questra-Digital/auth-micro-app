package queue

import (
	"context"
	"encoding/json"
	"email-service/internal/config"
	"email-service/internal/logger"
	"email-service/internal/mailer"
	"email-service/internal/models"
)

var consumerCtx context.Context
var consumerCancel context.CancelFunc

func StartEmailConsumer() error {
	consumerCtx, consumerCancel = context.WithCancel(context.Background())

	msgs, err := models.EmailChannel.Consume(
		config.AppConfig.RabbitMQQueue,
		"",    // consumer
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for {
			select {
			case msg, ok := <-msgs:
				if !ok {
					logger.Info("RabbitMQ channel closed")
					return
				}

				var job models.EmailJob
				if err := json.Unmarshal(msg.Body, &job); err != nil {
					logger.Error("Failed to parse email job: %v", err)
					continue
				}

				err := mailer.SendEmail(models.EmailJob{
					To:       job.To,
					Subject:  job.Subject,
					HTMLBody: job.HTMLBody,
				})
				if err != nil {
					logger.Error("Failed to send email: %v", err)
				} else {
					logger.SecureInfo("Email sent to: %s", job.To)
				}
			case <-consumerCtx.Done():
				logger.Info("Email consumer stopped")
				return
			}
		}
	}()

	logger.Info("Email consumer started")
	return nil
}

func StopEmailConsumer() {
	if consumerCancel != nil {
		consumerCancel()
		logger.Info("Email consumer shutdown signal sent")
	}
}
