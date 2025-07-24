## Prerequisites

- **Go** (version 1.21+)
- **Docker** & **Docker Compose** (for containerization)
- A valid **SMTP account** (e.g., Gmail with App Password)

## Project Structure

```
auth-microservice/
├── Dockerfile
├── docker-compose.yml
├── .env            # Environment variables (not committed)
├── go.mod
├── go.sum
├── main.go
├── handlers/       # HTTP handlers for signup and verify
│   ├── signup.go
│   └── verify.go
├── utils/          # Shared utilities (email sender, store)
│   ├── sender.go
│   └── memory.go
└── README.md       # This file
```


### Configure Environment Variables

Create a `.env` file in the project root with the following:

```env
# SMTP settings
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your_email@gmail.com
SMTP_PASS=your_app_password

# Service port
PORT=8080
```

**Note**: For Gmail, enable 2-Step Verification and generate an App Password.

### Install Go Dependencies

```bash
go mod tidy
```

## Running Locally

1. **Load** `.env` and run the service:

```bash
go run main.go
```

2. The server will start on `http://localhost:8080`.


### Using Docker Compose

1. Ensure your `.env` file is in the project root.
2. Start services:

```bash
docker-compose up --build
```

3. To stop:

```bash
docker-compose down
```

## API Endpoints

| Endpoint | Method | Payload | Description |
|----------|--------|---------|-------------|
| `/signup` | POST | `{ "email": "user@example.com" }` | Generate and send verification code |
| `/verify` | POST | `{ "email": "user@example.com", "code": "123456" }` | Verify the code |

## Testing the Service

### Signup

```bash
curl -X POST http://localhost:8080/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com"}'
```

**Expected Response:**
```
Verification code sent
```

### Verify

```bash
curl -X POST http://localhost:8080/verify \
  -H "Content-Type: application/json" \
  -d '{"email": "you@example.com", "code": "123456"}'
```

**Expected Response:**
```
Email verification successful
```