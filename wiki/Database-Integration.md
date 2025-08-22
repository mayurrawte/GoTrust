# Database Integration

GoTrust is database-agnostic. You can use any database by implementing the `UserStore` interface. This guide provides complete examples for popular databases.

## UserStore Interface

Every database integration must implement this interface:

```go
type UserStore interface {
    CreateUser(ctx context.Context, user *User, hashedPassword string) error
    GetUserByEmail(ctx context.Context, email string) (*User, string, error)
    GetUserByID(ctx context.Context, userID string) (*User, error)
    UpdateUser(ctx context.Context, user *User) error
    UserExists(ctx context.Context, email string) (bool, error)
}
```

## PostgreSQL Implementation

### Schema

```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    provider VARCHAR(50) DEFAULT 'local',
    password TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_provider ON users(provider);
```

### Implementation

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    "time"
    
    "github.com/mayurrawte/gotrust"
    _ "github.com/lib/pq"
)

type PostgresUserStore struct {
    db *sql.DB
}

func NewPostgresUserStore(connectionString string) (*PostgresUserStore, error) {
    db, err := sql.Open("postgres", connectionString)
    if err != nil {
        return nil, err
    }
    
    // Test connection
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    // Set connection pool settings
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(5)
    db.SetConnMaxLifetime(5 * time.Minute)
    
    return &PostgresUserStore{db: db}, nil
}

func (s *PostgresUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
    query := `
        INSERT INTO users (id, email, name, avatar_url, provider, password, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
        ON CONFLICT (email) DO NOTHING
    `
    
    result, err := s.db.ExecContext(ctx, query,
        user.ID, user.Email, user.Name, user.AvatarURL,
        user.Provider, hashedPassword, user.CreatedAt, user.UpdatedAt,
    )
    
    if err != nil {
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("user with email %s already exists", user.Email)
    }
    
    return nil
}

func (s *PostgresUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
    var user gotrust.User
    var password sql.NullString
    
    query := `
        SELECT id, email, name, avatar_url, provider, password, created_at, updated_at
        FROM users WHERE email = $1
    `
    
    err := s.db.QueryRowContext(ctx, query, email).Scan(
        &user.ID, &user.Email, &user.Name, &user.AvatarURL,
        &user.Provider, &password, &user.CreatedAt, &user.UpdatedAt,
    )
    
    if err == sql.ErrNoRows {
        return nil, "", fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, "", fmt.Errorf("database error: %w", err)
    }
    
    return &user, password.String, nil
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
    
    if err == sql.ErrNoRows {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }
    
    return &user, nil
}

func (s *PostgresUserStore) UpdateUser(ctx context.Context, user *gotrust.User) error {
    query := `
        UPDATE users 
        SET name = $2, avatar_url = $3, updated_at = $4
        WHERE id = $1
    `
    
    result, err := s.db.ExecContext(ctx, query,
        user.ID, user.Name, user.AvatarURL, time.Now(),
    )
    
    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }
    
    rowsAffected, _ := result.RowsAffected()
    if rowsAffected == 0 {
        return fmt.Errorf("user not found")
    }
    
    return nil
}

func (s *PostgresUserStore) UserExists(ctx context.Context, email string) (bool, error) {
    var exists bool
    query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
    
    err := s.db.QueryRowContext(ctx, query, email).Scan(&exists)
    if err != nil {
        return false, fmt.Errorf("database error: %w", err)
    }
    
    return exists, nil
}

func (s *PostgresUserStore) Close() error {
    return s.db.Close()
}
```

## MongoDB Implementation

### Schema Design

```javascript
// users collection
{
  "_id": ObjectId("..."),
  "email": "user@example.com",
  "name": "John Doe",
  "avatar_url": "https://...",
  "provider": "local",
  "password": "$2a$10$...", // bcrypt hash
  "created_at": ISODate("2024-01-01T00:00:00Z"),
  "updated_at": ISODate("2024-01-01T00:00:00Z")
}

// Indexes
db.users.createIndex({ "email": 1 }, { unique: true })
```

### Implementation

```go
package database

import (
    "context"
    "fmt"
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
    ID        primitive.ObjectID `bson:"_id,omitempty"`
    Email     string            `bson:"email"`
    Name      string            `bson:"name"`
    AvatarURL string            `bson:"avatar_url,omitempty"`
    Provider  string            `bson:"provider"`
    Password  string            `bson:"password,omitempty"`
    CreatedAt time.Time         `bson:"created_at"`
    UpdatedAt time.Time         `bson:"updated_at"`
}

func NewMongoUserStore(connectionString, dbName string) (*MongoUserStore, error) {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    client, err := mongo.Connect(ctx, options.Client().ApplyURI(connectionString))
    if err != nil {
        return nil, err
    }
    
    // Ping to verify connection
    if err := client.Ping(ctx, nil); err != nil {
        return nil, err
    }
    
    collection := client.Database(dbName).Collection("users")
    
    // Create unique index on email
    indexModel := mongo.IndexModel{
        Keys:    bson.D{{Key: "email", Value: 1}},
        Options: options.Index().SetUnique(true),
    }
    
    if _, err := collection.Indexes().CreateOne(ctx, indexModel); err != nil {
        return nil, fmt.Errorf("failed to create index: %w", err)
    }
    
    return &MongoUserStore{collection: collection}, nil
}

func (s *MongoUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
    doc := mongoUser{
        ID:        primitive.NewObjectID(),
        Email:     user.Email,
        Name:      user.Name,
        AvatarURL: user.AvatarURL,
        Provider:  user.Provider,
        Password:  hashedPassword,
        CreatedAt: user.CreatedAt,
        UpdatedAt: user.UpdatedAt,
    }
    
    result, err := s.collection.InsertOne(ctx, doc)
    if err != nil {
        if mongo.IsDuplicateKeyError(err) {
            return fmt.Errorf("user with email %s already exists", user.Email)
        }
        return fmt.Errorf("failed to create user: %w", err)
    }
    
    user.ID = result.InsertedID.(primitive.ObjectID).Hex()
    return nil
}

func (s *MongoUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
    var doc mongoUser
    
    err := s.collection.FindOne(ctx, bson.M{"email": email}).Decode(&doc)
    if err == mongo.ErrNoDocuments {
        return nil, "", fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, "", fmt.Errorf("database error: %w", err)
    }
    
    user := &gotrust.User{
        ID:        doc.ID.Hex(),
        Email:     doc.Email,
        Name:      doc.Name,
        AvatarURL: doc.AvatarURL,
        Provider:  doc.Provider,
        CreatedAt: doc.CreatedAt,
        UpdatedAt: doc.UpdatedAt,
    }
    
    return user, doc.Password, nil
}

func (s *MongoUserStore) GetUserByID(ctx context.Context, userID string) (*gotrust.User, error) {
    objectID, err := primitive.ObjectIDFromHex(userID)
    if err != nil {
        return nil, fmt.Errorf("invalid user ID format")
    }
    
    var doc mongoUser
    err = s.collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&doc)
    if err == mongo.ErrNoDocuments {
        return nil, fmt.Errorf("user not found")
    }
    if err != nil {
        return nil, fmt.Errorf("database error: %w", err)
    }
    
    return &gotrust.User{
        ID:        doc.ID.Hex(),
        Email:     doc.Email,
        Name:      doc.Name,
        AvatarURL: doc.AvatarURL,
        Provider:  doc.Provider,
        CreatedAt: doc.CreatedAt,
        UpdatedAt: doc.UpdatedAt,
    }, nil
}

func (s *MongoUserStore) UpdateUser(ctx context.Context, user *gotrust.User) error {
    objectID, err := primitive.ObjectIDFromHex(user.ID)
    if err != nil {
        return fmt.Errorf("invalid user ID format")
    }
    
    update := bson.M{
        "$set": bson.M{
            "name":       user.Name,
            "avatar_url": user.AvatarURL,
            "updated_at": time.Now(),
        },
    }
    
    result, err := s.collection.UpdateByID(ctx, objectID, update)
    if err != nil {
        return fmt.Errorf("failed to update user: %w", err)
    }
    
    if result.MatchedCount == 0 {
        return fmt.Errorf("user not found")
    }
    
    return nil
}

func (s *MongoUserStore) UserExists(ctx context.Context, email string) (bool, error) {
    count, err := s.collection.CountDocuments(ctx, bson.M{"email": email})
    if err != nil {
        return false, fmt.Errorf("database error: %w", err)
    }
    
    return count > 0, nil
}
```

## MySQL Implementation

### Schema

```sql
CREATE TABLE users (
    id VARCHAR(255) PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    avatar_url TEXT,
    provider VARCHAR(50) DEFAULT 'local',
    password TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,
    INDEX idx_email (email),
    INDEX idx_provider (provider)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_unicode_ci;
```

### Implementation

```go
package database

import (
    "context"
    "database/sql"
    "fmt"
    
    "github.com/mayurrawte/gotrust"
    _ "github.com/go-sql-driver/mysql"
)

type MySQLUserStore struct {
    db *sql.DB
}

func NewMySQLUserStore(dsn string) (*MySQLUserStore, error) {
    db, err := sql.Open("mysql", dsn)
    if err != nil {
        return nil, err
    }
    
    if err := db.Ping(); err != nil {
        return nil, err
    }
    
    return &MySQLUserStore{db: db}, nil
}

// Implementation similar to PostgreSQL with MySQL-specific syntax
// Main differences:
// - Use ? instead of $1 for placeholders
// - Use ON DUPLICATE KEY UPDATE for upsert
// - Different datetime handling
```

## SQLite Implementation

Perfect for development and small applications:

### Schema

```sql
CREATE TABLE users (
    id TEXT PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    name TEXT,
    avatar_url TEXT,
    provider TEXT DEFAULT 'local',
    password TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_users_email ON users(email);
```

## In-Memory Implementation

For testing and development:

```go
type InMemoryUserStore struct {
    mu        sync.RWMutex
    users     map[string]*gotrust.User
    passwords map[string]string
    emails    map[string]string // email -> userID mapping
}

func NewInMemoryUserStore() *InMemoryUserStore {
    return &InMemoryUserStore{
        users:     make(map[string]*gotrust.User),
        passwords: make(map[string]string),
        emails:    make(map[string]string),
    }
}

// Simple thread-safe implementation
// Perfect for testing and prototyping
```

## Best Practices

### 1. Connection Pooling

Always configure connection pools:

```go
// PostgreSQL/MySQL
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)

// MongoDB
clientOptions := options.Client().
    SetMaxPoolSize(100).
    SetMinPoolSize(10)
```

### 2. Context Handling

Always respect context cancellation:

```go
func (s *Store) GetUser(ctx context.Context, id string) (*User, error) {
    // Check context first
    if ctx.Err() != nil {
        return nil, ctx.Err()
    }
    
    // Use context in queries
    return s.queryWithContext(ctx, ...)
}
```

### 3. Error Handling

Distinguish between different error types:

```go
if err == sql.ErrNoRows || err == mongo.ErrNoDocuments {
    return nil, ErrUserNotFound
}
if isDuplicateKeyError(err) {
    return nil, ErrUserExists
}
return nil, fmt.Errorf("database error: %w", err)
```

### 4. Migrations

Use migration tools:

- **PostgreSQL/MySQL**: [golang-migrate](https://github.com/golang-migrate/migrate)
- **MongoDB**: Handle schema changes in application code

### 5. Indexes

Always create appropriate indexes:

```sql
-- Email lookup (most common)
CREATE INDEX idx_users_email ON users(email);

-- OAuth provider lookup
CREATE INDEX idx_users_provider ON users(provider);

-- Composite index for OAuth
CREATE INDEX idx_users_email_provider ON users(email, provider);
```

## Testing Your Implementation

```go
func TestUserStore(t *testing.T) {
    store := NewYourUserStore(...)
    
    // Test user creation
    user := &gotrust.User{
        ID:    "test_123",
        Email: "test@example.com",
    }
    
    err := store.CreateUser(context.Background(), user, "hashed_password")
    assert.NoError(t, err)
    
    // Test retrieval
    retrieved, password, err := store.GetUserByEmail(context.Background(), "test@example.com")
    assert.NoError(t, err)
    assert.Equal(t, user.ID, retrieved.ID)
    assert.Equal(t, "hashed_password", password)
    
    // Test duplicate prevention
    err = store.CreateUser(context.Background(), user, "password")
    assert.Error(t, err)
}
```

## Performance Tips

1. **Batch Operations**: For bulk operations, use transactions
2. **Caching**: Consider caching frequently accessed users
3. **Prepared Statements**: Use prepared statements for repeated queries
4. **Connection Reuse**: Don't create new connections for each operation

## Troubleshooting

### "Connection refused"
- Check database is running
- Verify connection string
- Check firewall/network settings

### "Too many connections"
- Increase connection pool size
- Check for connection leaks
- Use connection pooling

### "Duplicate key error"
- User already exists with that email
- Handle gracefully in your application

## Next Steps

- Implement [password reset](Password-Reset) functionality
- Add [email verification](Email-Verification)
- Setup [audit logging](Audit-Logging)
- Implement [soft deletes](Soft-Deletes) for users