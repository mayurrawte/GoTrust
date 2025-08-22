# Security Policy

## Reporting Security Issues

Found a security vulnerability? Please report it responsibly.

### Please do NOT:
- Open a public GitHub issue
- Disclose the vulnerability publicly before it has been addressed

### Please DO:
- Email the details to: security@gotrust.dev (or create a private security advisory on GitHub)
- Include the following information:
  - Type of vulnerability (e.g., SQL injection, XSS, authentication bypass)
  - Full paths of source file(s) related to the vulnerability
  - Location of the affected source code (tag/branch/commit or direct URL)
  - Step-by-step instructions to reproduce the issue
  - Proof-of-concept or exploit code (if possible)
  - Impact of the issue, including how an attacker might exploit it

### What to Expect

- **Acknowledgment**: We will acknowledge receipt of your vulnerability report within 48 hours
- **Initial Assessment**: Within 7 days, we will provide an initial assessment of the vulnerability
- **Fix Timeline**: We aim to release fixes for critical vulnerabilities within 30 days
- **Disclosure**: We will coordinate with you on the disclosure timeline

## Security Best Practices for Users

When using GoTrust in your applications, follow these security best practices:

### 1. JWT Secret Management
- **Never hardcode JWT secrets** in your source code
- Use a **minimum of 32 characters** for JWT secrets
- Rotate JWT secrets periodically
- Use different secrets for different environments (dev, staging, production)

```bash
# Generate a secure JWT secret
openssl rand -base64 32
```

### 2. Environment Variables
- Store sensitive configuration in environment variables or secure vaults
- Never commit `.env` files to version control
- Use tools like HashiCorp Vault or AWS Secrets Manager in production

### 3. Password Requirements
Implement strong password policies:
```go
// Example password validation
func validatePassword(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters")
    }
    // Add more validation rules as needed
    return nil
}
```

### 4. Rate Limiting
Always implement rate limiting on authentication endpoints:
```go
// Example with Echo middleware
e.Use(middleware.RateLimiter(middleware.NewRateLimiterMemoryStore(
    10, // requests per second
)))
```

### 5. HTTPS Only
- Always use HTTPS in production
- Set secure cookie flags
- Implement HSTS headers

```go
// Secure cookie example
cookie := &http.Cookie{
    Name:     "session",
    Value:    sessionID,
    HttpOnly: true,
    Secure:   true, // HTTPS only
    SameSite: http.SameSiteStrictMode,
}
```

### 6. OAuth Security
- Validate OAuth redirect URIs
- Use state parameters to prevent CSRF
- Validate all tokens from OAuth providers

### 7. Session Security
- Implement session timeouts
- Invalidate sessions on logout
- Use secure session storage (Redis with AUTH)

### 8. Input Validation
Always validate and sanitize user input:
```go
// Email validation example
import "net/mail"

func validateEmail(email string) error {
    _, err := mail.ParseAddress(email)
    return err
}
```

### 9. Database Security
- Use parameterized queries to prevent SQL injection
- Encrypt sensitive data at rest
- Use separate database users with minimal privileges

### 10. Monitoring and Logging
- Log authentication attempts (success and failure)
- Monitor for suspicious patterns
- Implement account lockout mechanisms
- Never log sensitive information (passwords, tokens)

```go
// Safe logging example
log.Printf("Login attempt for user: %s, success: %v", email, success)
// Never: log.Printf("Login with password: %s", password)
```

## Security Checklist

Before deploying GoTrust to production:

- [ ] JWT secret is strong and securely stored
- [ ] All environment variables are properly configured
- [ ] HTTPS is enabled with valid certificates
- [ ] Rate limiting is implemented
- [ ] Input validation is in place
- [ ] Database queries are parameterized
- [ ] Logging excludes sensitive data
- [ ] Session management is properly configured
- [ ] OAuth redirect URIs are whitelisted
- [ ] Security headers are configured (CSP, HSTS, etc.)
- [ ] Dependencies are up to date
- [ ] Error messages don't leak sensitive information

## Known Security Considerations

### JWT Token Storage
- JWTs are stateless and cannot be invalidated server-side once issued
- Consider using refresh tokens with shorter access token lifetimes
- Store tokens securely on the client (httpOnly cookies preferred over localStorage)

### OAuth State Parameter
- The state parameter is crucial for preventing CSRF attacks
- GoTrust automatically generates and validates state parameters
- Ensure your session store is properly configured

### Password Storage
- GoTrust uses bcrypt with a configurable cost factor (default: 10)
- Consider increasing the cost factor for highly sensitive applications
- Passwords are never stored in plain text

## Updates and Patches

- Security patches are released as soon as possible after verification
- Subscribe to our security mailing list for updates
- Check the CHANGELOG.md for security-related updates
- Use tools like `go list -m -u all` to check for updates

## Contact

For security concerns, please contact:
- Email: security@gotrust.dev
- GitHub Security Advisories: [Create private advisory](https://github.com/mayurrawte/gotrust/security/advisories/new)

## Acknowledgments

We appreciate the security research community's efforts in helping keep GoTrust and its users safe. Responsible disclosure of vulnerabilities helps us ensure the security and privacy of all our users.