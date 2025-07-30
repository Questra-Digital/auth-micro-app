# Auth Micro App - Test Flows

This document outlines all test scenarios for the auth-micro-app. All requests should be made to the **API Gateway** only.

## OTP Service Limits (from code analysis)
- **Max Resends**: 3 attempts
- **Max Verification Attempts**: 3 attempts
- **OTP TTL**: 5 minutes
- **Session TTL**: 15 minutes (API Gateway cookie)

---

## 1. Happy Path - Complete User Journey

### Flow 1.1: Successful Signup → OTP Verification → Resource Access
```bash
# 1. Signup
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -c cookies.txt

# 2. Verify OTP (use OTP from email/logs)
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "123456"}' \
  -b cookies.txt

# 3. Access Resources
curl -X GET http://localhost:8083/resources \
  -b cookies.txt

curl -X POST http://localhost:8083/resources \
  -H "Content-Type: application/json" \
  -d '{"name": "Test Resource", "description": "Test Description"}' \
  -b cookies.txt
```

---

## 2. OTP Flow Variations

### Flow 2.1: OTP Verification with Retries
```bash
# 1. Signup
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -c cookies.txt

# 2. First attempt (wrong OTP)
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "000000"}' \
  -b cookies.txt

# 3. Second attempt (wrong OTP)
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "000000"}' \
  -b cookies.txt

# 4. Third attempt (correct OTP) - Should succeed
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "123456"}' \
  -b cookies.txt

# 5. Fourth attempt - Should fail (max attempts exceeded)
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "000000"}' \
  -b cookies.txt
```

### Flow 2.2: OTP Resend Limits
```bash
# 1. Signup
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -c cookies.txt

# 2. First resend
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -b cookies.txt

# 3. Second resend
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -b cookies.txt

# 4. Third resend
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -b cookies.txt

# 5. Fourth resend - Should fail (max resends exceeded)
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -b cookies.txt
```

---

## 3. Resource API Access Flows

### Flow 3.1: Complete CRUD Operations
```bash
# 1. Signup and verify OTP first
# ... (use Flow 1.1 steps 1-2)

# 2. GET all resources
curl -X GET http://localhost:8083/resources \
  -b cookies.txt

# 3. GET specific resource
curl -X GET http://localhost:8083/resources/resource-1 \
  -b cookies.txt

# 4. POST new resource
curl -X POST http://localhost:8083/resources \
  -H "Content-Type: application/json" \
  -d '{"name": "New Resource", "description": "New Description"}' \
  -b cookies.txt

# 5. PUT update resource
curl -X PUT http://localhost:8083/resources/resource-1 \
  -H "Content-Type: application/json" \
  -d '{"name": "Updated Resource", "description": "Updated Description"}' \
  -b cookies.txt

# 6. DELETE resource
curl -X DELETE http://localhost:8083/resources/resource-1 \
  -b cookies.txt
```

### Flow 3.2: Resource Access with Different Scopes
```bash
# Test with different user scopes (requires JWT manipulation or different users)
# This tests the scope-based authorization in the API Gateway

# User with 'read' scope only
curl -X GET http://localhost:8083/resources \
  -b cookies.txt

# User with 'write' scope
curl -X POST http://localhost:8083/resources \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "description": "Test"}' \
  -b cookies.txt
```

---

## 4. Token Refresh Flow

### Flow 4.1: Automatic Token Refresh
```bash
# 1. Signup and verify OTP
# ... (use Flow 1.1 steps 1-2)

# 2. Access resource (should work normally)
curl -X GET http://localhost:8083/resources \
  -b cookies.txt

# 3. Wait for access token to expire (24 hours) or manipulate JWT
# 4. Try to access resource again - should auto-refresh and succeed
curl -X GET http://localhost:8083/resources \
  -b cookies.txt
```

### Flow 4.2: Expired Refresh Token
```bash
# 1. Signup and verify OTP
# 2. Wait for both access and refresh tokens to expire (7 days)
# 3. Try to access resource - should return 401 and clear session
curl -X GET http://localhost:8083/resources \
  -b cookies.txt
```

---

## 5. Error Scenarios

### Flow 5.1: Invalid Inputs
```bash
# Missing email
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{}' \
  -c cookies.txt

# Invalid email format
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "invalid-email"}' \
  -c cookies.txt

# Missing OTP
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -b cookies.txt

# Missing session cookie
curl -X GET http://localhost:8083/resources
```

### Flow 5.2: Session/Token Errors
```bash
# Expired session
# Wait 15 minutes after signup, then try to verify OTP
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "123456"}' \
  -b cookies.txt

# Invalid session ID
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "123456"}' \
  -H "Cookie: sessionId=invalid-session-id"

# Session from different IP (if implemented)
# Use sessionId from one IP on different IP
```

### Flow 5.3: Resource Not Found
```bash
# 1. Signup and verify OTP first
# 2. Try to access non-existent resource
curl -X GET http://localhost:8083/resources/non-existent-id \
  -b cookies.txt

curl -X PUT http://localhost:8083/resources/non-existent-id \
  -H "Content-Type: application/json" \
  -d '{"name": "Test", "description": "Test"}' \
  -b cookies.txt

curl -X DELETE http://localhost:8083/resources/non-existent-id \
  -b cookies.txt
```

---

## 6. Rate Limiting

### Flow 6.1: Rate Limit Testing
```bash
# Send multiple requests quickly to trigger rate limiting
for i in {1..10}; do
  curl -X POST http://localhost:8083/signup \
    -H "Content-Type: application/json" \
    -d "{\"email\": \"test$i@example.com\"}" \
    -c cookies$i.txt
done
```

---

## 7. Service Availability

### Flow 7.1: Downstream Service Failures
```bash
# 1. Stop OTP service and try signup
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -c cookies.txt

# 2. Stop Auth service and try OTP verification
curl -X POST http://localhost:8083/verify-otp \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com", "otp": "123456"}' \
  -b cookies.txt

# 3. Stop Resource service and try resource access
curl -X GET http://localhost:8083/resources \
  -b cookies.txt
```

---

## 8. Security Testing

### Flow 8.1: Session Hijacking Prevention
```bash
# 1. Signup from IP A
curl -X POST http://localhost:8083/signup \
  -H "Content-Type: application/json" \
  -d '{"email": "test@example.com"}' \
  -c cookies.txt

# 2. Try to use sessionId from IP A on IP B (if IP validation is implemented)
# This should fail
```

### Flow 8.2: JWT Tampering
```bash
# 1. Signup and verify OTP
# 2. Manually modify the JWT token in Redis or cookies
# 3. Try to access resources - should fail
```

---

## Expected Status Codes

| Scenario | Expected Status |
|----------|----------------|
| Successful signup | 200 |
| Successful OTP verification | 200 |
| Successful resource access | 200 |
| Missing email/OTP | 400 |
| Invalid email format | 400 |
| Invalid OTP | 401 |
| Expired session | 401 |
| Missing session | 401 |
| Max OTP attempts exceeded | 429 |
| Max OTP resends exceeded | 429 |
| Resource not found | 404 |
| Rate limit exceeded | 429 |
| Service unavailable | 502/503 |
| Unauthorized (no scope) | 403 |

---

## Notes

1. **Cookies**: Use `-c cookies.txt` to save cookies and `-b cookies.txt` to send cookies
2. **OTP**: In development mode, OTP is logged to console
3. **Session TTL**: 15 minutes for signup session, 24 hours for verified session
4. **Token TTL**: Access token 24 hours, refresh token 7 days
5. **Rate Limits**: 10,000 requests per minute per IP
6. **Dummy Data**: Resource service has 8 dummy resources pre-populated

---

## Quick Test Commands

```bash
# Quick happy path test
curl -X POST http://localhost:8083/signup -H "Content-Type: application/json" -d '{"email": "test@example.com"}' -c cookies.txt
curl -X POST http://localhost:8083/verify-otp -H "Content-Type: application/json" -d '{"email": "test@example.com", "otp": "123456"}' -b cookies.txt
curl -X GET http://localhost:8083/resources -b cookies.txt
``` 