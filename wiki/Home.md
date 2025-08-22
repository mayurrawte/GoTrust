# Welcome to GoTrust Wiki 🔐

GoTrust is a framework-agnostic authentication library for Go that handles the auth layer so you can focus on building your application.

## 📚 Documentation

### Getting Started
- [**Quick Start Guide**](Getting-Started) - Get up and running in 5 minutes
- [**Installation**](Installation) - Detailed installation instructions
- [**Configuration**](Configuration) - Environment variables and settings

### Core Concepts
- [**Architecture Overview**](Architecture) - How GoTrust works under the hood
- [**Framework Adapters**](Framework-Adapters) - Using GoTrust with Echo, Gin, Fiber, etc.
- [**Database Integration**](Database-Integration) - PostgreSQL, MongoDB, MySQL examples
- [**Session Management**](Session-Management) - Redis vs in-memory storage

### Guides
- [**Authentication Methods**](Authentication-Methods)
  - Email/Password
  - OAuth (Google, GitHub)
  - JWT Tokens
  - Refresh Tokens
- [**Middleware**](Middleware) - Protecting routes
- [**User Management**](User-Management) - User store implementation
- [**Security**](Security-Best-Practices) - Production security guidelines

### API Reference
- [**Core API**](API-Reference) - Complete API documentation
- [**HTTP Handlers**](HTTP-Handlers) - Handler reference
- [**Types**](Types) - Data structures and interfaces

### Advanced Topics
- [**Custom OAuth Providers**](Custom-OAuth-Providers) - Adding new OAuth providers
- [**Multi-tenancy**](Multi-Tenancy) - Building multi-tenant applications
- [**Microservices**](Microservices) - Using GoTrust in microservices
- [**Testing**](Testing) - Writing tests with GoTrust

### Resources
- [**Examples**](Examples) - Code examples for different scenarios
- [**Migration Guide**](Migration-Guide) - Migrating from other auth libraries
- [**Troubleshooting**](Troubleshooting) - Common issues and solutions
- [**FAQ**](FAQ) - Frequently asked questions

## 🚀 Quick Links

- [GitHub Repository](https://github.com/mayurrawte/GoTrust)
- [Report Issues](https://github.com/mayurrawte/GoTrust/issues)
- [Discussions](https://github.com/mayurrawte/GoTrust/discussions)
- [Releases](https://github.com/mayurrawte/GoTrust/releases)

## 💡 Why GoTrust?

### Framework Agnostic
Works with **any** Go web framework - Echo, Gin, Fiber, Chi, or standard net/http. No vendor lock-in.

### No Bloat
Modular design means you only import what you need. Using Echo? You won't get Gin dependencies.

### Production Ready
Built from real-world experience with security best practices baked in.

### Database Flexible
Works with any database through clean interfaces - PostgreSQL, MongoDB, MySQL, or even in-memory.

## 🎯 Use Cases

GoTrust is perfect for:
- **SaaS Applications** - Multi-tenant auth with team management
- **APIs** - JWT-based authentication for REST/GraphQL APIs  
- **Web Applications** - Session-based auth with OAuth support
- **Microservices** - Centralized auth service
- **MVPs** - Quick setup with production-ready auth

## 📊 Feature Comparison

| Feature | GoTrust | Auth0 | Firebase Auth | Supabase Auth |
|---------|---------|-------|---------------|---------------|
| Self-hosted | ✅ | ❌ | ❌ | ✅ |
| Framework agnostic | ✅ | ✅ | ❌ | ❌ |
| No vendor lock-in | ✅ | ❌ | ❌ | ❌ |
| Custom database | ✅ | ❌ | ❌ | ✅ |
| Free & Open Source | ✅ | ❌ | ❌ | ✅ |

## 🤝 Contributing

We welcome contributions! See our [Contributing Guide](https://github.com/mayurrawte/GoTrust/blob/main/CONTRIBUTING.md) for details.

## 📄 License

GoTrust is open source under the [MIT License](https://github.com/mayurrawte/GoTrust/blob/main/LICENSE).