# API Gateway

The entry point for all client requests. Orchestrates signup, OTP, registration, login, and resource access flows by proxying to other microservices.

## Features
- Signup and OTP orchestration
- Session management via Redis
- User registration and login
- JWT token validation and refresh
- Resource access with scope-based authorization
- Rate limiting per IP
- Audit logging to PostgreSQL

## Endpoints

| Endpoint           | Method | Description                                 |
|--------------------|--------|---------------------------------------------|
| `/signup`          | POST   | Start signup and trigger OTP                |
| `/verify-otp`      | POST   | Verify OTP and get access token             |
| `/resources`       | GET    | Get all resources (requires read scope)     |
| `/resources/:id`   | GET    | Get specific resource (requires read scope) |
| `/resources`       | POST   | Create new resource (requires write scope)  |
| `/resources/:id`   | PUT    | Update resource (requires write scope)      |
| `/resources/:id`   | DELETE | Delete resource (requires write scope)      |

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

### Get Resources
```bash
curl -X GET http://localhost:8080/resources \
  --cookie "sessionId=abcd1234"
```

### Create Resource
```bash
curl -X POST http://localhost:8080/resources \
  -H "Content-Type: application/json" \
  -d '{"name":"New Resource","description":"A new resource"}' \
  --cookie "sessionId=abcd1234"
```

### Update Resource
```bash
curl -X PUT http://localhost:8080/resources/550e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -d '{"name":"Updated Resource","description":"Updated description"}' \
  --cookie "sessionId=abcd1234"
```

### Delete Resource
```bash
curl -X DELETE http://localhost:8080/resources/550e8400-e29b-41d4-a716-446655440001 \
  --cookie "sessionId=abcd1234"
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
| SESSION_TTL_HOURS           | 24                          | Session TTL in hours                        |
| APP_ENV                     | development                 | Application environment                     |
| Audit_TTL_Days              | 30                          | Audit log retention in days                 |
| Rate_Limit_Per_Minute       | 10000                       | Requests per minute per IP                  |
| OTP_SERVICE_URL             | http://otp-service:8081     | OTP service endpoint                        |
| AUTHORIZATION_SERVICE_URL   | http://auth-service:8083    | Auth service endpoint                       |
| RESOURCE_SERVICE_URL        | http://resource-service:8084| Resource service endpoint                   |
| API_GATEWAY_PORT            | 8080                        | Service port                                |

## Running (Docker Compose)

This service is started automatically with `docker-compose up` from the project root.

---
