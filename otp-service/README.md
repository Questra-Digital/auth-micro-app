# OTP Service

A secure OTP (One-Time Password) service built with Go, Gin, Redis, and PostgreSQL.

## Features

- Generate and verify OTP codes
- Rate limiting with Redis
- Audit logging to PostgreSQL with automatic expiration
- Environment-based logging (development vs production)
- Docker support with optional external database deployment

## Setup

### Prerequisites
- Go 1.19+ (for local development)
- Docker and Docker Compose (for containerized deployment)
- PostgreSQL (optional - can use external database)
- Redis (optional - can use external Redis)

### Environment Variables
Create a `.env` file:

#### For Containerized Setup (PostgreSQL and Redis in containers)
```env
# PostgreSQL - using container
DB_HOST=postgres
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=12345678
DB_NAME=otp_audit
DB_SSLMODE=disable

# Redis - using container
REDIS_HOST=redis
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0

# App Environment
APP_ENV=development

# Data Retention (in days)
OTP_RETENTION_DAYS=30
RATE_LIMIT_RETENTION_DAYS=7
```

#### For External Database Setup
```env
# PostgreSQL - external database
DB_HOST=your-postgres-host
DB_PORT=5432
DB_USER=your-user
DB_PASSWORD=your-password
DB_NAME=otp_audit
DB_SSLMODE=disable

# Redis - external Redis
REDIS_HOST=your-redis-host
REDIS_PORT=6379
REDIS_PASSWORD=your-redis-password
REDIS_DB=0

# App Environment
APP_ENV=development

# Data Retention (in days)
OTP_RETENTION_DAYS=30
RATE_LIMIT_RETENTION_DAYS=7
```

## Deployment Options

### Option 1: Full Containerized Setup (Recommended for Development)
This setup includes PostgreSQL and Redis containers:

```bash
# Start all services (PostgreSQL, Redis, and OTP Service)
docker-compose up -d

# View logs
docker-compose logs -f otp-service

# Stop all services
docker-compose down
```

### Option 2: External Database Setup
For production or when using external databases, modify the docker-compose.yml:

1. **Remove PostgreSQL and Redis services** from docker-compose.yml
2. **Remove dependencies** from the otp-service section
3. **Update environment variables** to point to your external databases

```bash
# Start only the OTP service (requires external PostgreSQL and Redis)
docker-compose up -d otp-service
```

### Option 3: Local Development
```bash
# Run locally (requires PostgreSQL and Redis running)
go run main.go
```

## API Endpoints

The OTP service uses cookies for session management. Save cookies to a file for subsequent requests.

### 1. Generate OTP (POST /otp/generate)
```bash
curl -X POST http://localhost:8080/otp/generate \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -c cookies.txt
```

**Response:**
```json
{
  "success": true,
  "message": "OTP sent to test@example.com",
  "data": {
    "email": "test@example.com",
    "expires_at": "2024-01-15T10:30:00Z"
  }
}
```

### 2. Verify OTP (POST /otp/verify)
```bash
curl -X POST http://localhost:8080/otp/verify \
  -H "Content-Type: application/json" \
  -d '{"otp": "123456"}' \
  -b cookies.txt
```

**Response:**
```json
{
  "success": true,
  "message": "OTP verified successfully",
  "data": {
    "email": "test@example.com",
    "verified_at": "2024-01-15T10:25:00Z"
  }
}
```

### 3. Retry with Wrong OTP
```bash
curl -X POST http://localhost:8080/otp/verify \
  -H "Content-Type: application/json" \
  -d '{"otp": "000000"}' \
  -b cookies.txt
```

**Response:**
```json
{
  "success": false,
  "message": "Invalid OTP code",
  "error": "OTP_MISMATCH"
}
```

### 4. Generate Again (Resend OTP)
```bash
curl -X POST http://localhost:8080/otp/generate \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -b cookies.txt
```

## Data Retention

The service automatically expires audit records to prevent database bloat:

- **OTP Events**: 30 days (7 days in production)
- **Rate Limit Events**: 7 days (3 days in production)
- **Cleanup Job**: Runs every hour
- **Configurable**: Set `OTP_RETENTION_DAYS` and `RATE_LIMIT_RETENTION_DAYS` in `.env`

## Database Tables

- `otp_events`: OTP generation and verification logs
- `rate_limit_events`: Rate limiting activity logs

Both tables include `expires_at` field for automatic cleanup.