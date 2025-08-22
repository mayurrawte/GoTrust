# Frequently Asked Questions

## General

### What is GoTrust?

GoTrust is a framework-agnostic authentication library for Go that provides email/password auth, OAuth integration, JWT tokens, and session management without tying you to a specific web framework.

### Why not use Auth0/Firebase/Supabase?

GoTrust is self-hosted and gives you complete control:
- **No vendor lock-in** - Your auth, your rules
- **No usage limits** - Unlimited users, no pricing tiers
- **Data ownership** - User data stays in your database
- **Customizable** - Modify anything you need
- **Free forever** - Open source MIT license

### Which frameworks are supported?

Currently supported:
- Echo
- Gin
- Fiber (coming soon)
- Chi (coming soon)
- Standard net/http
- Any framework (via custom adapters)

### Which databases are supported?

Any database! GoTrust uses interfaces, so you can use:
- PostgreSQL
- MySQL/MariaDB
- MongoDB
- SQLite
- In-memory (for testing)
- Any custom storage

## Installation

### Why are adapters separate packages?

To avoid dependency bloat. If you use Echo, you shouldn't need Gin dependencies. Each adapter is isolated with its own go.mod file.

### How do I choose the right adapter?

Install only the adapter for your framework:
```bash
# For Echo
go get github.com/mayurrawte/gotrust/adapters/echo

# For Gin
go get github.com/mayurrawte/gotrust/adapters/gin

# For standard library
go get github.com/mayurrawte/gotrust/adapters/stdlib
```

### Can I use GoTrust without a framework?

Yes! Use the `stdlib` adapter with standard `net/http`:
```go
import stdAdapter "github.com/mayurrawte/gotrust/adapters/stdlib"
```

## Configuration

### What's the minimum JWT secret length?

32 characters minimum. Generate a secure one:
```bash
openssl rand -base64 32
```

### How do I configure OAuth providers?

Set environment variables:
```bash
export GOOGLE_CLIENT_ID="your-client-id"
export GOOGLE_CLIENT_SECRET="your-secret"
```

Or configure programmatically:
```go
config := gotrust.NewConfig()
config.GoogleClientID = "your-client-id"
```

### Can I use GoTrust without OAuth?

Yes! OAuth is optional. Just don't configure OAuth providers and those endpoints won't work.

### How do I change token expiration?

```go
config := gotrust.NewConfig()
config.JWTExpiration = 2 * time.Hour // Default is 24 hours
```

## Security

### How are passwords stored?

Passwords are hashed using bcrypt with a configurable cost factor (default 10):
```go
config.BCryptCost = 12 // Increase for more security
```

### Is GoTrust production-ready?

Yes, GoTrust implements security best practices:
- Bcrypt password hashing
- CSRF protection for OAuth
- Secure token generation
- SQL injection prevention
- XSS protection

### How do I implement rate limiting?

GoTrust doesn't include rate limiting, but you can add it:

```go
// With Echo
e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(10)))

// With Gin
router.Use(gin_rate_limit.RateLimiter(10, time.Second))
```

### Can I invalidate JWT tokens?

JWTs are stateless and can't be invalidated server-side. Options:
1. Use short expiration times with refresh tokens
2. Maintain a token blacklist in Redis
3. Use session-based auth instead of JWTs

## Common Issues

### "JWT_SECRET environment variable not set"

Set a secure JWT secret:
```bash
export JWT_SECRET="your-very-long-secret-key-at-least-32-chars"
```

### "User not found" on valid credentials

Check that:
1. User exists in database
2. Email is exact match (case-sensitive)
3. Password was hashed with same bcrypt settings

### "Invalid token" errors

Common causes:
- Missing "Bearer " prefix in Authorization header
- Token expired
- JWT secret changed
- Token from different environment

### OAuth redirect not working

1. Check redirect URI matches exactly
2. Verify OAuth app settings in Google/GitHub console
3. Ensure HTTPS in production
4. Check state parameter is being preserved

### "Too many connections" database error

Configure connection pooling:
```go
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
```

## Best Practices

### Should I use JWTs or sessions?

**JWTs** are good for:
- Stateless APIs
- Microservices
- Mobile apps
- Scaling horizontally

**Sessions** are good for:
- Traditional web apps
- Need to revoke access immediately
- Storing more user data

### How do I handle refresh tokens?

```go
// Store refresh token securely (httpOnly cookie recommended)
cookie := &http.Cookie{
    Name:     "refresh_token",
    Value:    response.RefreshToken,
    HttpOnly: true,
    Secure:   true, // HTTPS only
    SameSite: http.SameSiteStrictMode,
}
```

### How do I implement "Remember Me"?

Extend token expiration for remembered users:
```go
if rememberMe {
    config.JWTExpiration = 30 * 24 * time.Hour // 30 days
} else {
    config.JWTExpiration = 24 * time.Hour // 1 day
}
```

### How do I add custom claims to JWT?

Currently, you'd need to extend the JWT generation. We're working on making this easier in v2.

## Migration

### Migrating from Passport.js?

See our [Migration Guide](Migration-Guide) for step-by-step instructions.

### Can I use existing user tables?

Yes! Just map your existing schema in the UserStore implementation:
```go
func (s *Store) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
    // Map your schema to gotrust.User
    row := s.db.QueryRow("SELECT uuid, email_address, password_hash FROM accounts WHERE email_address = ?", email)
    // ...
}
```

### How do I migrate passwords from another system?

If using bcrypt, passwords work as-is. For other algorithms:
1. Store both old and new hashes temporarily
2. Verify with old algorithm, rehash with bcrypt
3. Update on successful login

## Performance

### How fast is GoTrust?

Typical performance:
- JWT validation: ~10μs
- Password hashing: ~50ms (bcrypt, intentionally slow)
- OAuth flow: Network-dependent
- Database queries: Depends on your implementation

### Can GoTrust handle high traffic?

Yes, GoTrust is stateless and scales horizontally. Tips:
- Use Redis for sessions
- Configure database connection pooling
- Cache user lookups if needed
- Use CDN for OAuth redirects

### Memory usage?

Minimal. GoTrust itself uses < 10MB. Memory usage depends on:
- Your framework choice
- Session storage (Redis vs memory)
- Connection pool size

## Troubleshooting

### Debug OAuth issues?

Enable debug logging:
```go
// Add logging to OAuth callbacks
fmt.Printf("OAuth state: %s\n", state)
fmt.Printf("OAuth code: %s\n", code)
```

### Test without HTTPS?

For development only:
```go
cookie.Secure = false // Allow HTTP in development
```

**Never disable secure cookies in production!**

## Future Features

### What's on the roadmap?

- Email verification
- Password reset
- Two-factor authentication (2FA)
- WebAuthn support
- More OAuth providers
- Admin dashboard

### Can I request a feature?

Yes! Open an issue on [GitHub](https://github.com/mayurrawte/GoTrust/issues) with your use case.

### How can I contribute?

See our [Contributing Guide](https://github.com/mayurrawte/GoTrust/blob/main/CONTRIBUTING.md). We welcome:
- Bug reports
- Feature requests
- Framework adapters
- Documentation improvements
- Code contributions

## Support

### Where can I get help?

1. Check this FAQ
2. Browse [Examples](https://github.com/mayurrawte/GoTrust/tree/main/examples)
3. Search [Issues](https://github.com/mayurrawte/GoTrust/issues)
4. Ask in [Discussions](https://github.com/mayurrawte/GoTrust/discussions)

### Is commercial support available?

Not currently, but feel free to reach out for consulting on large implementations.

### Found a security issue?

Please report security issues privately. See our [Security Policy](https://github.com/mayurrawte/GoTrust/blob/main/SECURITY.md).