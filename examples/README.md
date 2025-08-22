# GoTrust Examples

This directory contains practical examples of using GoTrust with different databases and configurations.

## Available Examples 📚

### [basic/](./basic)
Simple in-memory implementation. Good for:
- Getting started quickly
- Testing and development
- Understanding the API

### [mongodb/](./mongodb)  
MongoDB integration example showing:
- Document-based user storage
- Unique email indexing
- Production-ready patterns

### Coming Soon
- PostgreSQL example
- MySQL example
- Redis sessions example
- Multi-tenant example
- Microservices example

## Running Examples

Each example has its own README with specific instructions. Generally:

1. Set required environment variables (especially `JWT_SECRET`)
2. Install dependencies with `go mod download`
3. Run with `go run main.go`

## Common Testing Commands

These work with all examples:

```bash
# Create account
curl -X POST http://localhost:8080/auth/signup \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Login
curl -X POST http://localhost:8080/auth/signin \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Use the returned token for protected routes
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/profile
```

## Contributing Examples

Have an interesting use case? Feel free to contribute an example. Make sure to:
- Keep it simple and focused
- Include a README with setup instructions
- Add practical comments in the code
- Test it works with the latest GoTrust version