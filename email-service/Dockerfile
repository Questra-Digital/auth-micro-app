#############################################
# 1) Build stage
#############################################
FROM golang:1.24.5-alpine AS builder

WORKDIR /app

# Copy go.mod / go.sum and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy the rest of the source
COPY . .

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o email-service .

#############################################
# 2) Final image
#############################################
FROM alpine:3.18

# Install ca‑certificates for TLS/SSL
RUN apk add --no-cache ca-certificates

WORKDIR /app

# Copy binary, templates, and .env file
COPY --from=builder /app/email-service .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/.env ./.env

# (Optional) hardcode port or omit — EXPOSE is documentation only
EXPOSE 8080

# Default command
CMD ["./email-service"]
