# Changelog

All notable changes to this project will be documented here.

## [Unreleased]

### Added
- Initial release of GoTrust authentication library
- Email/password authentication with bcrypt hashing
- JWT token generation and validation
- OAuth 2.0 support for Google and GitHub
- Session management with Redis and in-memory storage
- Refresh token functionality
- Authentication middleware for Echo framework
- Database-agnostic design with UserStore interface
- Comprehensive documentation and examples
- Security best practices documentation
- Contributing guidelines

### Security
- Implemented secure password hashing with bcrypt
- Added CSRF protection for OAuth flows
- Secure session management with configurable expiration
- JWT token validation with configurable secrets

## [1.0.0] - 2024-01-XX (Planned)

### Added
- Core authentication functionality
- OAuth provider support (Google, GitHub)
- JWT and session-based authentication
- Middleware for Echo framework
- Documentation and examples

### Changed
- N/A (Initial release)

### Deprecated
- N/A (Initial release)

### Removed
- N/A (Initial release)

### Fixed
- N/A (Initial release)

### Security
- Initial security implementation following best practices

---

## Version History Guidelines

### Version Numbering
- **Major (X.0.0)**: Breaking API changes
- **Minor (0.X.0)**: New features, backwards compatible
- **Patch (0.0.X)**: Bug fixes, backwards compatible

### Categories
- **Added**: New features
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Fixed**: Bug fixes
- **Security**: Security fixes and improvements

### Example Future Entry
```markdown
## [1.1.0] - 2024-02-XX

### Added
- Microsoft OAuth provider support
- Two-factor authentication (2FA) with TOTP
- Email verification workflow
- Password reset functionality
- Account linking for multiple providers

### Changed
- Improved JWT token validation performance
- Updated default bcrypt cost to 12

### Fixed
- Fixed race condition in session cleanup
- Resolved OAuth state validation issue

### Security
- Added protection against timing attacks in password comparison
- Implemented account lockout after failed attempts
```