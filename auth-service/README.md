# Auth Service

A microservice responsible for user registration, authentication, JWT issuance, and audit logging.

## Features
- User registration and login
- JWT token generation
- PostgreSQL for user and audit data
- Redis for caching and rate limiting
- Audit logging for all authentication events
- Configurable rate limiting

## Endpoints

| Endpoint            | Method | Description                |
|---------------------|--------|----------------------------|
| `/registerUser`     | POST   | Register a new user        |
| `/login`            | POST   | Authenticate and get token |

## Example Usage

### Register User
```bash
curl -X POST http://localhost:8083/registerUser \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

### Login
```bash
curl -X POST http://localhost:8083/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'
```

## Environment Variables

| Variable                | Example Value         | Description                                 |
|-------------------------|----------------------|---------------------------------------------|
| JWT_SECRET              | "73853"             | Secret for signing JWT tokens               |
| DB_HOST                 | 172.17.0.1           | PostgreSQL host                             |
| DB_PORT                 | 5432                 | PostgreSQL port                             |
| DB_USER                 | postgres             | PostgreSQL user                             |
| DB_PASSWORD             | 12345678             | PostgreSQL password                         |
| AUDIT_DB_NAME           | audit                | Audit database name                         |
| USER_DB_NAME            | users                | User database name                          |
| DB_SSLMODE              | disable              | PostgreSQL SSL mode                         |
| REDIS_HOST              | redis                | Redis host                                  |
| REDIS_PORT              | 6379                 | Redis port                                  |
| REDIS_PASSWORD          | 12345678             | Redis password                              |
| REDIS_DB                | 0                    | Redis DB index                              |
| APP_ENV                 | development          | Application environment                     |
| Audit_TTL_Days          | 30                   | Audit log retention in days                 |
| Rate_Limit_Per_Minute   | 10000                | Requests per minute per IP                  |
| API_GATEWAY_PORT        | 8083                 | Service port                                |

## Running (Docker Compose)

This service is started automatically with `docker-compose up` from the project root.

---
