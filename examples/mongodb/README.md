# MongoDB Example 🍃

This example shows how to use GoTrust with MongoDB as the user store.

## Prerequisites

- MongoDB running locally or accessible via connection string
- Go 1.19+

## Setup

1. Start MongoDB (if not running):
```bash
# Using Docker
docker run -d -p 27017:27017 --name mongodb mongo:latest

# Or use your existing MongoDB installation
```

2. Set environment variables:
```bash
export JWT_SECRET="your-secret-key-at-least-32-characters"
export MONGO_URI="mongodb://localhost:27017"  # Optional, defaults to localhost
```

3. Install dependencies:
```bash
go mod init mongodb-example
go get github.com/mayurrawte/gotrust
go get go.mongodb.org/mongo-driver/mongo
go get github.com/labstack/echo/v4
```

4. Run the example:
```bash
go run main.go
```

## Testing the API

Create a user:
```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123",
    "name": "John Doe"
  }'
```

Sign in:
```bash
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

Access protected route:
```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_TOKEN_HERE"
```

## MongoDB Schema

The users collection will have documents like:
```json
{
  "_id": ObjectId("..."),
  "email": "john@example.com",
  "name": "John Doe",
  "avatar_url": "",
  "provider": "local",
  "password": "$2a$10$...",  // bcrypt hash
  "created_at": ISODate("2024-01-01T00:00:00Z"),
  "updated_at": ISODate("2024-01-01T00:00:00Z")
}
```

## Notes

- The email field has a unique index to prevent duplicates
- Passwords are hashed using bcrypt
- OAuth users won't have a password field
- The example uses in-memory session storage; for production, use Redis