# Getting Started with GoTrust

This guide will help you integrate GoTrust into your application in less than 5 minutes.

## Prerequisites

- Go 1.19 or higher
- A web framework (Echo, Gin, Fiber, or net/http)
- Basic understanding of Go

## Step 1: Installation

Install the core library and your framework adapter:

```bash
# Core library (always required)
go get github.com/mayurrawte/gotrust

# Choose your framework adapter:
go get github.com/mayurrawte/gotrust/adapters/echo   # For Echo
go get github.com/mayurrawte/gotrust/adapters/gin    # For Gin
go get github.com/mayurrawte/gotrust/adapters/stdlib # For net/http
```

## Step 2: Set Environment Variables

Create a `.env` file or set these environment variables:

```bash
# Required
export JWT_SECRET="your-secret-key-at-least-32-characters-long"

# Optional OAuth (if you want social login)
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
export GITHUB_CLIENT_ID="your-github-client-id"
export GITHUB_CLIENT_SECRET="your-github-client-secret"

# Optional Redis (for session storage)
export REDIS_URL="redis://localhost:6379"
```

## Step 3: Implement UserStore

GoTrust needs to know how to store users in your database. Create a simple implementation:

```go
type MyUserStore struct {
    db *sql.DB // or *mongo.Database, etc.
}

func (s *MyUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
    // Save user to your database
    _, err := s.db.ExecContext(ctx, 
        "INSERT INTO users (id, email, password) VALUES (?, ?, ?)",
        user.ID, user.Email, hashedPassword)
    return err
}

func (s *MyUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
    // Fetch user from your database
    var user gotrust.User
    var password string
    err := s.db.QueryRowContext(ctx,
        "SELECT id, email, password FROM users WHERE email = ?", email).
        Scan(&user.ID, &user.Email, &password)
    return &user, password, err
}

// ... implement other methods
```

See [Database Integration](Database-Integration) for complete examples.

## Step 4: Initialize GoTrust

```go
package main

import (
    "github.com/mayurrawte/gotrust"
    // Import your framework and adapter
)

func main() {
    // 1. Create configuration
    config := gotrust.NewConfig()
    
    // 2. Create your user store
    userStore := &MyUserStore{db: database}
    
    // 3. Create session store
    sessionStore := gotrust.NewMemorySessionStore()
    
    // 4. Initialize auth service
    authService := gotrust.NewAuthService(config, userStore, sessionStore)
    
    // 5. Create handlers
    handlers := gotrust.NewGenericAuthHandlers(authService, config)
    
    // Continue with framework setup...
}
```

## Step 5: Add to Your Framework

### Echo Example

```go
import (
    "github.com/labstack/echo/v4"
    echoAdapter "github.com/mayurrawte/gotrust/adapters/echo"
)

func main() {
    // ... GoTrust setup from Step 4
    
    e := echo.New()
    
    // Register auth routes
    echoAdapter.RegisterRoutes(e, "/auth", handlers)
    
    // Protect routes
    api := e.Group("/api")
    api.Use(echoAdapter.WrapMiddleware(handlers.AuthMiddleware()))
    
    api.GET("/profile", profileHandler)
    
    e.Start(":8080")
}
```

### Gin Example

```go
import (
    "github.com/gin-gonic/gin"
    ginAdapter "github.com/mayurrawte/gotrust/adapters/gin"
)

func main() {
    // ... GoTrust setup from Step 4
    
    router := gin.Default()
    
    // Register auth routes
    ginAdapter.RegisterRoutes(router, "/auth", handlers)
    
    // Protect routes
    api := router.Group("/api")
    api.Use(ginAdapter.WrapMiddleware(handlers.AuthMiddleware()))
    
    api.GET("/profile", profileHandler)
    
    router.Run(":8080")
}
```

### Standard net/http Example

```go
import (
    "net/http"
    stdAdapter "github.com/mayurrawte/gotrust/adapters/stdlib"
)

func main() {
    // ... GoTrust setup from Step 4
    
    mux := http.NewServeMux()
    
    // Register auth routes
    stdAdapter.RegisterRoutes(mux, "/auth", handlers)
    
    // Protect routes
    protectedHandler := stdAdapter.AuthMiddleware(handlers)(profileHandler)
    mux.HandleFunc("/api/profile", protectedHandler)
    
    http.ListenAndServe(":8080", mux)
}
```

## Step 6: Test Your Setup

### Create a User

```bash
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123",
    "name": "John Doe"
  }'
```

Response:
```json
{
  "user": {
    "id": "user_123",
    "email": "user@example.com",
    "name": "John Doe"
  },
  "access_token": "eyJhbGciOiJ...",
  "refresh_token": "eyJhbGciOiJ...",
  "expires_in": 86400
}
```

### Sign In

```bash
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{
    "email": "user@example.com",
    "password": "password123"
  }'
```

### Access Protected Route

```bash
curl -X GET http://localhost:8080/api/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## Available Auth Endpoints

After setup, these endpoints are automatically available:

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/auth/signup` | Register new user |
| POST | `/auth/signin` | Login with email/password |
| POST | `/auth/refresh` | Refresh access token |
| POST | `/auth/logout` | Logout (invalidate session) |
| GET | `/auth/user` | Get current user |
| GET | `/auth/google` | Google OAuth login |
| GET | `/auth/github` | GitHub OAuth login |

## Next Steps

- [Configure OAuth providers](OAuth-Setup) for social login
- [Add Redis](Session-Management) for production session storage
- [Implement password reset](Password-Reset)
- [Add email verification](Email-Verification)
- [Setup rate limiting](Rate-Limiting)

## Common Issues

### "JWT_SECRET not set"
Make sure to set the JWT_SECRET environment variable with at least 32 characters.

### "User store not implemented"
Ensure all UserStore interface methods are implemented, even if some return errors initially.

### "Token invalid"
Check that you're including "Bearer " prefix in the Authorization header.

## Need Help?

- Check the [FAQ](FAQ)
- Browse [Examples](https://github.com/mayurrawte/GoTrust/tree/main/examples)
- Open an [Issue](https://github.com/mayurrawte/GoTrust/issues)
- Join [Discussions](https://github.com/mayurrawte/GoTrust/discussions)