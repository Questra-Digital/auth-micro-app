# Email Service

A microservice for sending OTP emails, with PostgreSQL auditing, Redis rate limiting, and RabbitMQ queue processing.

## Features
- SMTP email sending (Gmail or other)
- PostgreSQL audit logging
- Redis rate limiting
- RabbitMQ queue for reliable email delivery
- Automatic cleanup of old audit records

## Endpoints

| Endpoint     | Method | Description         |
|--------------|--------|---------------------|
| `/send-otp`  | POST   | Send OTP email      |

## Example Usage

### Send OTP Email
```bash
curl -X POST http://localhost:8082/send-otp \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","otp":"123456"}'
```

## Environment Variables

| Variable            | Example Value                        | Description                                 |
|---------------------|-------------------------------------|---------------------------------------------|
| PORT                | 8082                                | Service port                                |
| SMTP_HOST           | smtp.gmail.com                      | SMTP server host                            |
| SMTP_PORT           | 587                                 | SMTP server port                            |
| SMTP_USERNAME       | your-email@gmail.com                 | SMTP username                               |
| SMTP_PASSWORD       | your-app-password                    | SMTP password (App Password for Gmail)      |
| SMTP_FROM_NAME      | Questra Digital                      | Sender name                                 |
| REDIS_ADDR          | redis:6379                           | Redis address                               |
| REDIS_PASSWORD      | 12345678                             | Redis password                              |
| DB_HOST             | 172.17.0.1                           | PostgreSQL host                             |
| DB_PORT             | 5432                                 | PostgreSQL port                             |
| DB_USER             | postgres                             | PostgreSQL user                             |
| DB_PASSWORD         | 12345678                             | PostgreSQL password                         |
| DB_NAME             | audit                                | Audit database name                         |
| RATE_LIMIT_PER_SECOND| 10000                               | Requests per second per IP                  |
| RABBITMQ_URL        | amqp://admin:adminPassword@rabbitmq:5672/ | RabbitMQ connection URL               |
| RABBITMQ_QUEUE      | email_queue                          | RabbitMQ queue name                         |
| APP_MODE            | development                          | Application environment                     |

## Running (Docker Compose)

This service is started automatically with `docker-compose up` from the project root.

---
