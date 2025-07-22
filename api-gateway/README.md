# API Gateway

The entry point for all client requests. Orchestrates signup, OTP, registration, and login flows by proxying to other microservices.

## Features
- Signup and OTP orchestration
- Session management via Redis
- User registration and login
- Rate limiting per IP
- Audit logging to PostgreSQL

## Endpoints

| Endpoint           | Method | Description                                 |
|--------------------|--------|---------------------------------------------|
| `/signup`          | POST   | Start signup and trigger OTP                |
| `/verify-otp`      | POST   | Verify OTP for session                      |
| `/registerUser`    | POST   | Register user after OTP verification        |
| `/login`           | POST   | Authenticate user and get JWT               |

## Example Usage

### Signup
```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'
```

### Verify OTP
```bash
curl -X POST http://localhost:8080/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"otp":"123456"}' --cookie "sessionId=abcd1234"
```

### Register User
```bash
curl -X POST http://localhost:8080/registerUser \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}' --cookie "sessionId=abcd1234"
```

### Login
```bash
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

## Environment Variables

| Variable                    | Example Value                | Description                                 |
|-----------------------------|-----------------------------|---------------------------------------------|
| DB_HOST                     | 172.17.0.1                  | PostgreSQL host                             |
| DB_PORT                     | 5432                        | PostgreSQL port                             |
| DB_USER                     | postgres                    | PostgreSQL user                             |
| DB_PASSWORD                 | 12345678                    | PostgreSQL password                         |
| DB_NAME                     | audit                       | Audit database name                         |
| DB_SSLMODE                  | disable                     | PostgreSQL SSL mode                         |
| REDIS_HOST                  | redis                       | Redis host                                  |
| REDIS_PORT                  | 6379                        | Redis port                                  |
| REDIS_PASSWORD              | 12345678                    | Redis password                              |
| REDIS_DB                    | 0                           | Redis DB index                              |
| APP_ENV                     | development                 | Application environment                     |
| Audit_TTL_Days              | 30                          | Audit log retention in days                 |
| Rate_Limit_Per_Minute       | 10000                       | Requests per minute per IP                  |
| OTP_SERVICE_URL             | http://otp-service:8081     | OTP service endpoint                        |
| AUTHORIZATION_SERVICE_URL   | http://auth-service:8083    | Auth service endpoint                       |
| API_GATEWAY_PORT            | 8080                        | Service port                                |

## Running (Docker Compose)

This service is started automatically with `docker-compose up` from the project root.

---
