# GoTrust 🔐

A pragmatic authentication library for Go applications. Built from real-world experience to handle the auth layer so you can focus on your business logic.

## Why GoTrust?

After building authentication for multiple production applications, I found myself writing the same auth code over and over. GoTrust extracts those battle-tested patterns into a reusable library that just works.

## Features ✨

- 🔑 Email/password authentication with bcrypt
- 🌐 OAuth providers (Google, GitHub) with easy extensibility  
- 🎫 JWT tokens with refresh token rotation
- 💾 Session management via Redis or in-memory storage
- 🗄️ Database-agnostic through interfaces
- ⚡ Echo framework middleware (more frameworks coming)
- 🛡️ Production-ready security defaults

## Installation

```bash
go get github.com/mayurrawte/gotrust
```

## Basic Usage

```go
// 1. Implement UserStore for your database
userStore := NewPostgresUserStore(db)  // or MongoDB, MySQL, etc.

// 2. Create auth service
authService := gotrust.NewAuthService(
    gotrust.NewConfig(),
    userStore,
    gotrust.NewMemorySessionStore(),
)

// 3. Add to your Echo app
e := echo.New()
handlers := gotrust.NewAuthHandlers(authService, config)
handlers.RegisterRoutes(e, "/auth")

// 4. Protect routes
api := e.Group("/api")
api.Use(authService.AuthMiddleware())
```

That's it! 🎉 Your app now has production-ready authentication.

## Full Setup Guide

### 1. Set Environment Variables 🔧

```bash
# JWT Configuration (Required)
export JWT_SECRET="your-secret-key-min-32-chars"
export JWT_ISSUER="your-app-name"

# Google OAuth (Optional)
export GOOGLE_CLIENT_ID="your-google-client-id"
export GOOGLE_CLIENT_SECRET="your-google-client-secret"
export GOOGLE_REDIRECT_URI="http://localhost:4000/auth/google/callback"

# GitHub OAuth (Optional)
export GITHUB_CLIENT_ID="your-github-client-id"
export GITHUB_CLIENT_SECRET="your-github-client-secret"
export GITHUB_REDIRECT_URI="http://localhost:4000/auth/github/callback"

# Redis (Optional - for session storage)
export REDIS_URL="redis://localhost:6379"

# Frontend URLs
export FRONTEND_SUCCESS_URL="http://localhost:3000/auth/success"
export FRONTEND_ERROR_URL="http://localhost:3000/auth/error"
```

### 2. Implement UserStore Interface

GoTrust works with any database. Just implement the `UserStore` interface:

```go
type UserStore interface {
    CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error
    GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error)
    GetUserByID(ctx context.Context, userID string) (*gotrust.User, error)
    UpdateUser(ctx context.Context, user *gotrust.User) error
    UserExists(ctx context.Context, email string) (bool, error)
}
```

#### Example: PostgreSQL Implementation 🐘

```go
package main

import (
    "context"
    "database/sql"
    "github.com/mayurrawte/gotrust"
    _ "github.com/lib/pq"
)

type PostgresUserStore struct {
    db *sql.DB
}

func NewPostgresUserStore(db *sql.DB) *PostgresUserStore {
    return &PostgresUserStore{db: db}
}

func (s *PostgresUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
    query := `
        INSERT INTO users (id, email, name, avatar_url, provider, password, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
    `
    _, err := s.db.ExecContext(ctx, query,
        user.ID, user.Email, user.Name, user.AvatarURL, 
        user.Provider, hashedPassword, user.CreatedAt, user.UpdatedAt,
    )
    return err
}

func (s *PostgresUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
    var user gotrust.User
    var hashedPassword string
    
    query := `
        SELECT id, email, name, avatar_url, provider, password, created_at, updated_at
        FROM users WHERE email = $1
    `
    err := s.db.QueryRowContext(ctx, query, email).Scan(
        &user.ID, &user.Email, &user.Name, &user.AvatarURL,
        &user.Provider, &hashedPassword, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        return nil, "", err
    }
    
    return &user, hashedPassword, nil
}

func (s *PostgresUserStore) GetUserByID(ctx context.Context, userID string) (*gotrust.User, error) {
    var user gotrust.User
    
    query := `
        SELECT id, email, name, avatar_url, provider, created_at, updated_at
        FROM users WHERE id = $1
    `
    err := s.db.QueryRowContext(ctx, query, userID).Scan(
        &user.ID, &user.Email, &user.Name, &user.AvatarURL,
        &user.Provider, &user.CreatedAt, &user.UpdatedAt,
    )
    if err != nil {
        return nil, err
    }
    
    return &user, nil
}

func (s *PostgresUserStore) UpdateUser(ctx context.Context, user *gotrust.User) error {
    query := `
        UPDATE users 
        SET name = $2, avatar_url = $3, updated_at = $4
        WHERE id = $1
    `
    _, err := s.db.ExecContext(ctx, query,
        user.ID, user.Name, user.AvatarURL, user.UpdatedAt,
    )
    return err
}

func (s *PostgresUserStore) UserExists(ctx context.Context, email string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
    err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
    return exists, err
}
```

#### Example: In-Memory Implementation (for testing)

```go
type InMemoryUserStore struct {
    mu       sync.RWMutex
    users    map[string]*gotrust.User
    passwords map[string]string
}

func NewInMemoryUserStore() *InMemoryUserStore {
    return &InMemoryUserStore{
        users:     make(map[string]*gotrust.User),
        passwords: make(map[string]string),
    }
}

func (s *InMemoryUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
    s.mu.Lock()
    defer s.mu.Unlock()
    
    s.users[user.Email] = user
    if hashedPassword != "" {
        s.passwords[user.Email] = hashedPassword
    }
    return nil
}

func (s *InMemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
    s.mu.RLock()
    defer s.mu.RUnlock()
    
    user, exists := s.users[email]
    if !exists {
        return nil, "", fmt.Errorf("user not found")
    }
    
    password := s.passwords[email]
    return user, password, nil
}

// ... implement other methods similarly
```

#### Example: MongoDB Implementation 🍃

```go
package main

import (
    "context"
    "time"
    
    "github.com/mayurrawte/gotrust"
    "go.mongodb.org/mongo-driver/bson"
    "go.mongodb.org/mongo-driver/bson/primitive"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
)

type MongoUserStore struct {
    collection *mongo.Collection
}

type mongoUser struct {
    ID           primitive.ObjectID `bson:"_id,omitempty"`
    Email        string            `bson:"email"`
    Name         string            `bson:"name"`
    AvatarURL    string            `bson:"avatar_url"`
    Provider     string            `bson:"provider"`
    Password     string            `bson:"password"`
    CreatedAt    time.Time         `bson:"created_at"`
    UpdatedAt    time.Time         `bson:"updated_at"`
}

func NewMongoUserStore(db *mongo.Database) (*MongoUserStore, error) {
    collection := db.Collection("users")
    
    // Create unique index on email
    _, err := collection.Indexes().CreateOne(context.Background(), mongo.IndexModel{
        Keys:    bson.D{{"email", 1}},
        Options: options.Index().SetUnique(true),
    })
    if err != nil {
        return nil, err
    }
    
    return &MongoUserStore{
        collection: collection,
    }, nil
}

func (s *MongoUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
    mongoDoc := mongoUser{
        ID:        primitive.NewObjectID(),
        Email:     user.Email,
        Name:      user.Name,
        AvatarURL: user.AvatarURL,
        Provider:  user.Provider,
        Password:  hashedPassword,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
    
    result, err := s.collection.InsertOne(ctx, mongoDoc)
    if err != nil {
        return err
    }
    
    // Update user ID with MongoDB ObjectID
    user.ID = result.InsertedID.(primitive.ObjectID).Hex()
    return nil
}

func (s *MongoUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
    var mongoDoc mongoUser
    
    err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&mongoDoc)
    if err != nil {
        if err == mongo.ErrNoDocuments {
            return nil, "", fmt.Errorf("user not found")
        }
        return nil, "", err
    }
    
    user := &gotrust.User{
        ID:        mongoDoc.ID.Hex(),
        Email:     mongoDoc.Email,
        Name:      mongoDoc.Name,
        AvatarURL: mongoDoc.AvatarURL,
        Provider:  mongoDoc.Provider,
        CreatedAt: mongoDoc.CreatedAt,
        UpdatedAt: mongoDoc.UpdatedAt,
    }
    
    return user, mongoDoc.Password, nil
}

func (s *MongoUserStore) GetUserByID(ctx context.Context, userID string) (*gotrust.User, error) {
    objectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, fmt.Errorf("invalid user ID")
    }
    
    var mongoDoc mongoUser
    err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&mongoDoc)
    if err != nil {
        return nil, err
    }
    
    return &gotrust.User{
        ID:        mongoDoc.ID.Hex(),
        Email:     mongoDoc.Email,
        Name:      mongoDoc.Name,
        AvatarURL: mongoDoc.AvatarURL,
        Provider:  mongoDoc.Provider,
        CreatedAt: mongoDoc.CreatedAt,
        UpdatedAt: mongoDoc.UpdatedAt,
    }, nil
}

func (s *MongoUserStore) UpdateUser(ctx context.Context, user *gotrust.User) error {
    objectID, err := primitive.ObjectIDFromHex(user.ID)
    if err != nil {
        return err
    }
    
    update := bson.M{
        "$set": bson.M{
            "name":       user.Name,
            "avatar_url": user.AvatarURL,
            "updated_at": time.Now(),
        },
    }
    
    _, err = s.collection.UpdateByID(ctx, objectID, update)
    return err
}

func (s *MongoUserStore) UserExists(ctx context.Context, email string) (bool, error) {
    count, err := s.collection.CountDocuments(ctx, bson.M{"email": email})
    if err != nil {
        return false, err
    }
    return count > 0, nil
}

// Usage with MongoDB
func main() {
    // Connect to MongoDB
    client, err := mongo.Connect(context.Background(), 
        options.Client().ApplyURI("mongodb://localhost:27017"))
    if err != nil {
        log.Fatal(err)
    }
    
    db := client.Database("myapp")
    userStore, err := NewMongoUserStore(db)
    if err != nil {
        log.Fatal(err)
    }
    
    // Rest of the setup is the same
    config := gotrust.NewConfig()
    sessionStore := gotrust.NewMemorySessionStore()
    authService := gotrust.NewAuthService(config, userStore, sessionStore)
    // ...
}
```

### 3. Initialize and Use GoTrust 🚀

```go
package main

import (
    "database/sql"
    "log"
    
    "github.com/labstack/echo/v4"
    "github.com/mayurrawte/gotrust"
    _ "github.com/lib/pq"
)

func main() {
    // 1. Setup database connection
    db, err := sql.Open("postgres", "postgresql://user:pass@localhost/dbname?sslmode=disable")
    if err != nil {
        log.Fatal(err)
    }
    defer db.Close()
    
    // Create users table if not exists
    createTableSQL := `
    CREATE TABLE IF NOT EXISTS users (
        id VARCHAR(255) PRIMARY KEY,
        email VARCHAR(255) UNIQUE NOT NULL,
        name VARCHAR(255),
        avatar_url TEXT,
        provider VARCHAR(50),
        password TEXT,
        created_at TIMESTAMP,
        updated_at TIMESTAMP
    );`
    db.Exec(createTableSQL)
    
    // 2. Create configuration
    config := gotrust.NewConfig()
    
    // 3. Create user store
    userStore := NewPostgresUserStore(db)
    
    // 4. Create session store (Redis or Memory)
    var sessionStore gotrust.SessionStore
    if config.RedisURL != "" {
        sessionStore, err = gotrust.NewRedisSessionStore(config.RedisURL)
        if err != nil {
            log.Printf("Redis not available, using memory store: %v", err)
            sessionStore = gotrust.NewMemorySessionStore()
        }
    } else {
        sessionStore = gotrust.NewMemorySessionStore()
    }
    
    // 5. Create auth service
    authService := gotrust.NewAuthService(config, userStore, sessionStore)
    
    // 6. Setup Echo server
    e := echo.New()
    
    // 7. Register auth routes
    handlers := gotrust.NewAuthHandlers(authService, config)
    handlers.RegisterRoutes(e, "/auth")
    
    // 8. Setup protected routes
    protected := e.Group("/api")
    protected.Use(authService.AuthMiddleware())
    
    protected.GET("/profile", func(c echo.Context) error {
        userID, _ := gotrust.GetUserFromContext(c)
        return c.JSON(200, map[string]string{
            "user_id": userID,
            "message": "This is a protected route",
        })
    })
    
    // 9. Setup optional auth routes (public + authenticated)
    public := e.Group("/public")
    public.Use(authService.OptionalAuthMiddleware())
    
    public.GET("/info", func(c echo.Context) error {
        userID, _ := c.Get("user_id").(string)
        if userID != "" {
            return c.JSON(200, map[string]string{
                "message": "Hello authenticated user",
                "user_id": userID,
            })
        }
        return c.JSON(200, map[string]string{
            "message": "Hello anonymous user",
        })
    })
    
    // 10. Start server
    log.Println("Server starting on :4000")
    e.Start(":4000")
}
```

## API Reference

### Authentication Endpoints

| Method | Endpoint | Description | Request Body |
|--------|----------|-------------|--------------|
| POST | `/auth/signup` | Register new user | `{"email": "...", "password": "...", "name": "..."}` |
| POST | `/auth/signin` | Login with email/password | `{"email": "...", "password": "..."}` |
| POST | `/auth/refresh` | Refresh access token | `{"refresh_token": "..."}` |
| POST | `/auth/logout` | Logout (invalidate session) | - |
| GET | `/auth/user` | Get current user info | - |

### OAuth Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/auth/google` | Initiate Google OAuth |
| GET | `/auth/google/callback` | Google OAuth callback |
| GET | `/auth/github` | Initiate GitHub OAuth |
| GET | `/auth/github/callback` | GitHub OAuth callback |

### Response Format

#### Successful Authentication
```json
{
    "user": {
        "id": "user_123",
        "email": "user@example.com",
        "name": "John Doe",
        "avatar_url": "https://...",
        "provider": "local"
    },
    "access_token": "eyJhbGciOiJ...",
    "refresh_token": "eyJhbGciOiJ...",
    "expires_in": 86400
}
```

#### Error Response
```json
{
    "error": "Invalid credentials"
}
```

## Middleware Options

### 1. Required Authentication
```go
// All routes in this group require authentication
protected := e.Group("/api")
protected.Use(authService.AuthMiddleware())

protected.GET("/secret", handler)
```

### 2. Optional Authentication
```go
// Routes work for both authenticated and anonymous users
public := e.Group("/public")
public.Use(authService.OptionalAuthMiddleware())

public.GET("/content", func(c echo.Context) error {
    userID, _ := c.Get("user_id").(string)
    if userID != "" {
        // User is authenticated
    } else {
        // User is anonymous
    }
})
```

### 3. Session-based Authentication
```go
// Use sessions instead of JWT tokens
sessionRoutes := e.Group("/session")
sessionRoutes.Use(authService.SessionMiddleware())
```

## Common Use Cases

### Custom User Data
```go
// Extend the User struct in your application
type AppUser struct {
    gotrust.User
    Role        string    `json:"role"`
    Permissions []string  `json:"permissions"`
    LastLoginAt time.Time `json:"last_login_at"`
}
```

### Rate Limiting
```go
// Add rate limiting to auth endpoints
import "github.com/labstack/echo/v4/middleware"

authGroup := e.Group("/auth")
authGroup.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))
handlers.RegisterRoutes(authGroup, "")
```

### Custom Claims in JWT
```go
// After successful authentication, add custom claims
func customAuthHandler(c echo.Context) error {
    // ... authenticate user
    
    // Add custom claims to context
    c.Set("user_role", "admin")
    c.Set("tenant_id", "tenant_123")
    
    // ... return response
}
```

## Security Best Practices 🔒

1. **Use strong JWT secrets**: At least 32 characters
2. **Enable HTTPS**: Always use TLS in production
3. **Set secure cookies**: Use `HttpOnly` and `Secure` flags
4. **Implement rate limiting**: Prevent brute force attacks
5. **Validate email addresses**: Implement email verification
6. **Use CSRF protection**: For web applications
7. **Regular token rotation**: Use refresh tokens
8. **Audit logging**: Log authentication events

## Configuration Options

| Environment Variable | Description | Default | Required |
|---------------------|-------------|---------|----------|
| `JWT_SECRET` | Secret key for JWT signing (min 32 chars) | - | ✅ |
| `JWT_ISSUER` | JWT issuer claim | `gotrust` | ❌ |
| `GOOGLE_CLIENT_ID` | Google OAuth client ID | - | ❌ |
| `GOOGLE_CLIENT_SECRET` | Google OAuth client secret | - | ❌ |
| `GITHUB_CLIENT_ID` | GitHub OAuth client ID | - | ❌ |
| `GITHUB_CLIENT_SECRET` | GitHub OAuth client secret | - | ❌ |
| `REDIS_URL` | Redis connection URL | - | ❌ |
| `ALLOW_SIGNUP` | Enable user registration | `true` | ❌ |
| `REQUIRE_EMAIL_VERIFICATION` | Require email verification | `false` | ❌ |
| `FRONTEND_SUCCESS_URL` | OAuth success redirect URL | `http://localhost:3000/auth/success` | ❌ |
| `FRONTEND_ERROR_URL` | OAuth error redirect URL | `http://localhost:3000/auth/error` | ❌ |

## Testing 🧪

```go
func TestAuthentication(t *testing.T) {
    // Use in-memory store for testing
    userStore := NewInMemoryUserStore()
    sessionStore := gotrust.NewMemorySessionStore()
    config := gotrust.NewConfig()
    
    authService := gotrust.NewAuthService(config, userStore, sessionStore)
    
    // Test signup
    ctx := context.Background()
    response, err := authService.SignUp(ctx, &gotrust.SignUpRequest{
        Email:    "test@example.com",
        Password: "password123",
        Name:     "Test User",
    })
    
    assert.NoError(t, err)
    assert.NotEmpty(t, response.AccessToken)
    
    // Test signin
    signInResp, err := authService.SignIn(ctx, &gotrust.SignInRequest{
        Email:    "test@example.com",
        Password: "password123",
    })
    
    assert.NoError(t, err)
    assert.Equal(t, "test@example.com", signInResp.User.Email)
}
```

## Contributing

Found a bug? Have a feature request? PRs are welcome. Check out [CONTRIBUTING.md](CONTRIBUTING.md) for guidelines.

## License

MIT - see [LICENSE](LICENSE)

## Support & Links

- [Report Issues](https://github.com/mayurrawte/GoTrust/issues)
- [Examples](https://github.com/mayurrawte/GoTrust/tree/main/examples)
- [Documentation](https://github.com/mayurrawte/GoTrust/wiki)

## Roadmap 🚀

- [ ] Email verification
- [ ] Password reset functionality
- [ ] Two-factor authentication (2FA)
- [ ] Additional OAuth providers (Microsoft, Twitter, etc.)
- [ ] WebAuthn support
- [ ] Account linking
- [ ] Audit logging
- [ ] Admin dashboard

---

Built for the Go community. Feel free to use, modify, and contribute.