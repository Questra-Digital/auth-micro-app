# Email Service

A microservice for sending emails with PostgreSQL auditing, Redis rate limiting, and RabbitMQ queue processing.

## Features

- **SMTP Email Sending**: Gmail SMTP integration
- **PostgreSQL Auditing**: Complete email audit trail
- **Redis Rate Limiting**: Configurable request throttling
- **RabbitMQ Queue**: Reliable email processing
- **Docker Deployment**: Containerized with health checks

## Quick Start

### Prerequisites

- Docker & Docker Compose
- Gmail App Password (for SMTP)

### Setup

1. **Clone and configure:**
```bash
git clone <repository>
cd email-service
```

2. **Update `.env` file:**
```env
# Update with your Gmail credentials
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password
```

3. **Deploy:**
```bash
docker-compose up --build
```

### API Usage

**Send OTP Email:**
```bash
curl -X POST http://localhost:8080/send-otp \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "otp": "123456"
  }'
```

## Services

- **App**: `http://localhost:8080` - Email service API
- **PostgreSQL**: `localhost:5432` - Audit database
- **Redis**: `localhost:6379` - Rate limiting
- **RabbitMQ**: `localhost:5672` - Message queue
- **RabbitMQ Management**: `http://localhost:15672` - Queue monitoring

## Configuration

Key environment variables in `.env`:

```env
# SMTP (Gmail)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your-email@gmail.com
SMTP_PASSWORD=your-app-password

# Database
DB_HOST=postgres
DB_USER=admin
DB_PASSWORD=adminPassword
DB_NAME=email_audit

# Redis
REDIS_ADDR=redis:6379
REDIS_PASSWORD=12345678

# RabbitMQ
RABBITMQ_URL=amqp://admin:adminPassword@rabbitmq:5672/
RABBITMQ_QUEUE=email_queue
```

## Development

### Local Development

1. **Install Go 1.24+**
2. **Run services:**
```bash
docker-compose up postgres redis rabbitmq
```
3. **Run app:**
```bash
go run main.go
```

### Testing

```bash
# Test email sending
curl -X POST http://localhost:8080/send-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "123456"}'
```

## Monitoring

- **Email Audits**: Check PostgreSQL `email_audits` table
- **Queue Status**: RabbitMQ Management UI at `http://localhost:15672`
- **Rate Limiting**: Redis stores rate limit data
- **Logs**: `docker-compose logs app`

## Troubleshooting

### Common Issues

1. **SMTP Authentication Failed**
   - Use Gmail App Password, not regular password
   - Enable 2FA on Gmail account

2. **Database Connection Failed**
   - Check PostgreSQL is running: `docker-compose logs postgres`
   - Verify credentials in `.env`

3. **Rate Limiting Not Working**
   - Check Redis connection: `docker-compose logs redis`
   - Verify Redis password in `.env`

### Health Checks

```bash
# Check all services
docker-compose ps

# View logs
docker-compose logs app
docker-compose logs postgres
docker-compose logs redis
docker-compose logs rabbitmq
```

## Production Deployment

1. **Update `.env` with production values**
2. **Set up proper SSL certificates**
3. **Configure firewall rules**
4. **Set up monitoring and alerting**
5. **Use production-grade PostgreSQL/Redis/RabbitMQ**
