# Basic GoTrust Example 💡

This example demonstrates a basic implementation of GoTrust with in-memory storage.

## Features

- Email/password authentication
- JWT token generation
- Protected and public routes
- Optional authentication middleware
- OAuth integration ready (Google & GitHub)

## Setup

1. Set environment variables:
```bash
export JWT_SECRET="your-secret-key-at-least-32-characters-long"

# Optional: OAuth configuration
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
export GITHUB_CLIENT_ID="your-github-client-id"
export GITHUB_CLIENT_SECRET="your-github-client-secret"
```

2. Run the example:
```bash
go run main.go
```

3. The server will start on http://localhost:8080

## API Testing

### 1. Sign Up
```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

### 2. Sign In
```bash
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

Response:
```json
{
  "user": {
    "id": "abc123",
    "email": "user@example.com",
    "name": "John Doe"
  },
  "access_token": "eyJhbGciOiJ...",
  "refresh_token": "eyJhbGciOiJ...",
  "expires_in": 86400
}
```

### 3. Access Protected Route
```bash
# Use the access_token from signin response
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 4. Refresh Token
```bash
curl -X POST http://localhost:8080/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

### 5. Public Content (Optional Auth)
```bash
# Without authentication
curl -X GET http://localhost:8080/public/content

# With authentication (personalized)
curl -X GET http://localhost:8080/public/content \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

### 6. Logout
```bash
curl -X POST http://localhost:8080/auth/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## OAuth Testing

### Google OAuth
1. Visit http://localhost:8080/auth/google in your browser
2. Complete Google authentication
3. You'll be redirected back with authentication tokens

### GitHub OAuth
1. Visit http://localhost:8080/auth/github in your browser
2. Complete GitHub authentication
3. You'll be redirected back with authentication tokens

## Notes

- This example uses in-memory storage which resets on server restart
- For production, implement a proper database-backed UserStore
- Configure real OAuth credentials for testing OAuth flows
- The JWT secret should be stored securely in production