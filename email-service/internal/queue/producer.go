package queue

import (
	"encoding/json"
	"email-service/internal/config"
	"email-service/internal/logger"
	"email-service/internal/models"
	"github.com/streadway/amqp"
)

func InitRabbitMQ() error {
	logger.Info("Connecting to RabbitMQ at: %s", config.AppConfig.RabbitMQURL)
	
	conn, err := amqp.Dial(config.AppConfig.RabbitMQURL)
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ: %v", err)
		return err
	}

	ch, err := conn.Channel()
	if err != nil {
		logger.Error("Failed to open RabbitMQ channel: %v", err)
		return err
	}

	_, err = ch.QueueDeclare(
		config.AppConfig.RabbitMQQueue,
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		logger.Error("Failed to declare RabbitMQ queue: %v", err)
		return err
	}

	models.EmailChannel = ch
	logger.Info("RabbitMQ initialized successfully")
	return nil
}

func PublishEmailJob(job models.EmailJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return err
	}

	err = models.EmailChannel.Publish(
		"", // exchange
		config.AppConfig.RabbitMQQueue,
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		return err
	}

	logger.Info("Published email job to RabbitMQ")
	return nil
}
