# Resource Service

A microservice responsible for managing resources with basic CRUD operations and user context extraction.

## Features
- Basic CRUD operations for resources
- In-memory storage with pre-populated dummy data
- User context extraction from headers
- No authentication/authorization (handled by API Gateway)
- Simple and lightweight design

## Endpoints

| Endpoint            | Method | Description                |
|---------------------|--------|----------------------------|
| `/resources`        | GET    | Get all resources          |
| `/resources/:id`    | GET    | Get specific resource      |
| `/resources`        | POST   | Create new resource        |
| `/resources/:id`    | PUT    | Update existing resource   |
| `/resources/:id`    | DELETE | Delete resource            |

## Example Usage

### Get All Resources
```bash
curl -X GET http://localhost:8084/resources \
  -H "X-User-ID: user123" \
  -H "X-User-Email: user@example.com" \
  -H "X-User-Scopes: read,write"
```

### Get Specific Resource
```bash
curl -X GET http://localhost:8084/resources/550e8400-e29b-41d4-a716-446655440001 \
  -H "X-User-ID: user123" \
  -H "X-User-Email: user@example.com" \
  -H "X-User-Scopes: read,write"
```

### Create Resource
```bash
curl -X POST http://localhost:8084/resources \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -H "X-User-Email: user@example.com" \
  -H "X-User-Scopes: read,write" \
  -d '{"name":"New Resource","description":"A new resource"}'
```

### Update Resource
```bash
curl -X PUT http://localhost:8084/resources/550e8400-e29b-41d4-a716-446655440001 \
  -H "Content-Type: application/json" \
  -H "X-User-ID: user123" \
  -H "X-User-Email: user@example.com" \
  -H "X-User-Scopes: read,write" \
  -d '{"name":"Updated Resource","description":"Updated description"}'
```

### Delete Resource
```bash
curl -X DELETE http://localhost:8084/resources/550e8400-e29b-41d4-a716-446655440001 \
  -H "X-User-ID: user123" \
  -H "X-User-Email: user@example.com" \
  -H "X-User-Scopes: read,write"
```

## Environment Variables

| Variable                | Example Value         | Description                                 |
|-------------------------|----------------------|---------------------------------------------|
| APP_ENV                 | development          | Application environment                     |
| RESOURCE_SERVICE_PORT   | 8084                 | Service port                                |

## Running (Docker Compose)

This service is started automatically with `docker-compose up` from the project root.

--- 