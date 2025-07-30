# Auth Service

A microservice responsible for user authentication, JWT issuance, token refresh, and audit logging.

## Features
- User authentication and JWT token generation
- Access token refresh functionality
- PostgreSQL for user and audit data
- Redis for refresh token storage
- Audit logging for all authentication events
- Configurable rate limiting

## Endpoints

| Endpoint            | Method | Description                |
|---------------------|--------|----------------------------|
| `/getAccessToken`   | POST   | Get access token for user  |
| `/refreshToken`     | POST   | Refresh expired access token |

## Example Usage

### Get Access Token
```bash
curl -X POST http://localhost:8083/getAccessToken \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com"}'
```

### Refresh Token
```bash
curl -X POST http://localhost:8083/refreshToken \
  -H "Content-Type: application/json" \
  -d '{"grant_type":"refresh_token","refresh_token":"token123","email":"user@example.com"}'
```

## Environment Variables

| Variable                | Example Value         | Description                                 |
|-------------------------|----------------------|---------------------------------------------|
| JWT_SECRET              |   73853             | Secret for signing JWT tokens               |
| ACCESS_TOKEN_DURATION   | 1                   | Access token duration in hours              |
| REFRESH_TOKEN_DURATION  | 7                   | Refresh token duration in days              |
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
| APP_PORT                | 8083                 | Service port                                |

## Running (Docker Compose)

This service is started automatically with `docker-compose up` from the project root.

---
