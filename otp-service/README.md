# OTP Service

A microservice for secure OTP (One-Time Password) generation and verification, with audit logging and rate limiting.

## Features
- Generate and verify OTP codes
- Redis-backed session storage
- PostgreSQL audit logging
- Rate limiting per IP
- Automatic cleanup of old OTP events

## Endpoints

| Endpoint         | Method | Description                |
|------------------|--------|----------------------------|
| `/otp/generate`  | POST   | Generate and send OTP      |
| `/otp/verify`    | POST   | Verify submitted OTP       |

## Example Usage

### Generate OTP
```bash
curl -X POST http://localhost:8081/otp/generate \
  -H "Content-Type: application/json" \
  -H "X-Session-Id: abcd1234" \
  -d '{"email":"user@example.com"}'
```

### Verify OTP
```bash
curl -X POST http://localhost:8081/otp/verify \
  -H "Content-Type: application/json" \
  -H "X-Session-Id: abcd1234" \
  -d '{"otp":"123456"}'
```

## Environment Variables

| Variable              | Example Value                | Description                                 |
|-----------------------|-----------------------------|---------------------------------------------|
| DB_HOST               | 172.17.0.1                  | PostgreSQL host                             |
| DB_PORT               | 5432                        | PostgreSQL port                             |
| DB_USER               | postgres                    | PostgreSQL user                             |
| DB_PASSWORD           | 12345678                    | PostgreSQL password                         |
| DB_NAME               | audit                       | Audit database name                         |
| DB_SSLMODE            | disable                     | PostgreSQL SSL mode                         |
| REDIS_HOST            | redis                       | Redis host                                  |
| REDIS_PORT            | 6379                        | Redis port                                  |
| REDIS_PASSWORD        | 12345678                    | Redis password                              |
| REDIS_DB              | 0                           | Redis DB index                              |
| APP_ENV               | development                 | Application environment                     |
| OTP_EVENT_TTL_DAYS    | 30                          | OTP event retention in days                 |
| OTP_SERVICE_PORT      | 8081                        | Service port                                |
| Email_Service_URL     | http://email-service:8082   | Email service endpoint                      |

## Running (Docker Compose)

This service is started automatically with `docker-compose up` from the project root.

---
