# Contributing to GoTrust 🤝

Thanks for your interest in contributing! This project grew from solving real authentication problems, and contributions from developers facing similar challenges make it better for everyone.

## How Can I Contribute?

### Reporting Bugs

Before creating bug reports, please check existing issues as you might find out that you don't need to create one. When you are creating a bug report, please include as many details as possible:

* **Use a clear and descriptive title** for the issue to identify the problem.
* **Describe the exact steps which reproduce the problem** in as many details as possible.
* **Provide specific examples to demonstrate the steps**. Include links to files or GitHub projects, or copy/pasteable snippets.
* **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
* **Explain which behavior you expected to see instead and why.**
* **Include details about your configuration and environment:**
  * Which version of GoTrust are you using?
  * What's the name and version of the OS you're using?
  * Which version of Go are you using?

### Suggesting Enhancements

Enhancement suggestions are tracked as GitHub issues. When creating an enhancement suggestion, please include:

* **Use a clear and descriptive title** for the issue to identify the suggestion.
* **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
* **Provide specific examples to demonstrate the steps**.
* **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
* **Explain why this enhancement would be useful** to most GoTrust users.

### Pull Requests

1. Fork the repo and create your branch from `main`.
2. If you've added code that should be tested, add tests.
3. If you've changed APIs, update the documentation.
4. Ensure the test suite passes.
5. Make sure your code lints.
6. Issue that pull request!

## Development Setup

1. Fork and clone the repository:
```bash
git clone https://github.com/your-username/gotrust.git
cd gotrust
```

2. Install dependencies:
```bash
go mod download
```

3. Create a `.env` file for testing:
```bash
cat > .env <<EOF
JWT_SECRET=test-secret-key-for-development-only-32chars
REDIS_URL=redis://localhost:6379
EOF
```

4. Run tests:
```bash
go test ./...
```

5. Run tests with coverage:
```bash
go test -cover ./...
```

## Project Structure

```
gotrust/
├── auth.go           # Core authentication service
├── config.go         # Configuration management
├── handlers.go       # HTTP handlers
├── jwt.go           # JWT token management
├── middleware.go     # Authentication middleware
├── oauth.go         # OAuth providers integration
├── session.go       # Session management
├── types.go         # Type definitions
├── examples/        # Example implementations
│   ├── basic/       # Basic authentication example
│   ├── postgres/    # PostgreSQL integration example
│   └── redis/       # Redis session example
└── tests/           # Test files
```

## Coding Standards

### Go Code Style

* Follow the official [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
* Use `gofmt` to format your code
* Use `golint` and `go vet` to check for common issues
* Write clear, idiomatic Go code

### Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line

Example:
```
Add support for Microsoft OAuth provider

- Implement Microsoft OAuth flow
- Add configuration for Azure AD
- Update documentation with setup instructions

Fixes #123
```

### Testing

* Write unit tests for new functionality
* Ensure all tests pass before submitting PR
* Aim for at least 80% code coverage
* Include both positive and negative test cases

Example test:
```go
func TestJWTManager_GenerateToken(t *testing.T) {
    manager := NewJWTManager("secret", "issuer", time.Hour)
    
    claims := TokenClaims{
        UserID: "user123",
        Email:  "test@example.com",
    }
    
    token, err := manager.GenerateToken(claims)
    assert.NoError(t, err)
    assert.NotEmpty(t, token)
    
    // Validate the generated token
    validatedClaims, err := manager.ValidateToken(token)
    assert.NoError(t, err)
    assert.Equal(t, claims.UserID, validatedClaims.UserID)
}
```

### Documentation

* Add godoc comments to all exported functions, types, and packages
* Update README.md if you change functionality
* Include examples in your documentation
* Keep documentation clear and concise

Example:
```go
// NewAuthService creates a new authentication service with the provided configuration.
// It requires a UserStore implementation for user persistence and a SessionStore
// for session management. Returns an initialized AuthService ready to handle
// authentication requests.
//
// Example:
//   config := gotrust.NewConfig()
//   userStore := NewPostgresUserStore(db)
//   sessionStore := gotrust.NewMemorySessionStore()
//   authService := gotrust.NewAuthService(config, userStore, sessionStore)
func NewAuthService(config *Config, userStore UserStore, sessionStore SessionStore) *AuthService {
    // ...
}
```

## Release Process

1. Update CHANGELOG.md with release notes
2. Update version in relevant files
3. Create a git tag: `git tag -a v1.0.0 -m "Release version 1.0.0"`
4. Push the tag: `git push origin v1.0.0`
5. Create a GitHub release with the changelog

## Questions?

Feel free to open an issue with your question or reach out to the maintainers directly.

## License

By contributing, you agree that your contributions will be licensed under the MIT License.