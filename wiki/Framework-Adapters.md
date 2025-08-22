# Framework Adapters

GoTrust is framework-agnostic and provides adapters for popular Go web frameworks. Each adapter is a separate module, so you only import dependencies for your chosen framework.

## Supported Frameworks

| Framework | Adapter Package | Status |
|-----------|----------------|---------|
| [Echo](https://echo.labstack.com/) | `gotrust/adapters/echo` | ✅ Stable |
| [Gin](https://gin-gonic.com/) | `gotrust/adapters/gin` | ✅ Stable |
| [Fiber](https://gofiber.io/) | `gotrust/adapters/fiber` | 🚧 Coming Soon |
| [Chi](https://go-chi.io/) | `gotrust/adapters/chi` | 🚧 Coming Soon |
| net/http | `gotrust/adapters/stdlib` | ✅ Stable |
| [Gorilla/mux](https://github.com/gorilla/mux) | `gotrust/adapters/gorilla` | 🚧 Coming Soon |

## Echo Adapter

### Installation

```bash
go get github.com/mayurrawte/gotrust/adapters/echo
```

### Usage

```go
package main

import (
    "github.com/labstack/echo/v4"
    "github.com/mayurrawte/gotrust"
    echoAdapter "github.com/mayurrawte/gotrust/adapters/echo"
)

func main() {
    // Setup GoTrust
    config := gotrust.NewConfig()
    userStore := NewUserStore()
    authService := gotrust.NewAuthService(config, userStore, gotrust.NewMemorySessionStore())
    handlers := gotrust.NewGenericAuthHandlers(authService, config)
    
    // Setup Echo
    e := echo.New()
    
    // Register auth routes
    echoAdapter.RegisterRoutes(e, "/auth", handlers)
    
    // Protected routes
    api := e.Group("/api")
    api.Use(echoAdapter.WrapMiddleware(handlers.AuthMiddleware()))
    
    api.GET("/profile", func(c echo.Context) error {
        userID := c.Get("user_id").(string)
        return c.JSON(200, map[string]string{"user_id": userID})
    })
    
    e.Start(":8080")
}
```

### Middleware Usage

```go
// Required auth - returns 401 if not authenticated
api.Use(echoAdapter.WrapMiddleware(handlers.AuthMiddleware()))

// Optional auth - allows both authenticated and anonymous
public.Use(echoAdapter.WrapMiddleware(handlers.OptionalAuthMiddleware()))
```

## Gin Adapter

### Installation

```bash
go get github.com/mayurrawte/gotrust/adapters/gin
```

### Usage

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/mayurrawte/gotrust"
    ginAdapter "github.com/mayurrawte/gotrust/adapters/gin"
)

func main() {
    // Setup GoTrust
    config := gotrust.NewConfig()
    userStore := NewUserStore()
    authService := gotrust.NewAuthService(config, userStore, gotrust.NewMemorySessionStore())
    handlers := gotrust.NewGenericAuthHandlers(authService, config)
    
    // Setup Gin
    router := gin.Default()
    
    // Register auth routes
    ginAdapter.RegisterRoutes(router, "/auth", handlers)
    
    // Protected routes
    api := router.Group("/api")
    api.Use(ginAdapter.WrapMiddleware(handlers.AuthMiddleware()))
    
    api.GET("/profile", func(c *gin.Context) {
        userID := c.GetString("user_id")
        c.JSON(200, gin.H{"user_id": userID})
    })
    
    router.Run(":8080")
}
```

### Accessing User Context

```go
func profileHandler(c *gin.Context) {
    userID := c.GetString("user_id")
    email := c.GetString("user_email")
    name := c.GetString("user_name")
    
    c.JSON(200, gin.H{
        "user_id": userID,
        "email": email,
        "name": name,
    })
}
```

## Standard Library (net/http) Adapter

### Installation

```bash
go get github.com/mayurrawte/gotrust/adapters/stdlib
```

### Usage

```go
package main

import (
    "net/http"
    "github.com/mayurrawte/gotrust"
    stdAdapter "github.com/mayurrawte/gotrust/adapters/stdlib"
)

func main() {
    // Setup GoTrust
    config := gotrust.NewConfig()
    userStore := NewUserStore()
    authService := gotrust.NewAuthService(config, userStore, gotrust.NewMemorySessionStore())
    handlers := gotrust.NewGenericAuthHandlers(authService, config)
    
    // Setup routes
    mux := http.NewServeMux()
    
    // Register auth routes
    stdAdapter.RegisterRoutes(mux, "/auth", handlers)
    
    // Protected route
    mux.HandleFunc("/api/profile", stdAdapter.AuthMiddleware(handlers)(profileHandler))
    
    http.ListenAndServe(":8080", mux)
}

func profileHandler(w http.ResponseWriter, r *http.Request) {
    // User context is available via the adapter
    ctx := stdAdapter.NewStdContext(w, r)
    userID := ctx.Get("user_id").(string)
    
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "user_id": userID,
    })
}
```

### Using with Popular Routers

#### With Gorilla/mux

```go
import "github.com/gorilla/mux"

router := mux.NewRouter()
stdAdapter.RegisterRoutes(router, "/auth", handlers)

// Protected routes
api := router.PathPrefix("/api").Subrouter()
api.Use(func(next http.Handler) http.Handler {
    return stdAdapter.AuthMiddleware(handlers)(next.ServeHTTP)
})
```

#### With Chi

```go
import "github.com/go-chi/chi/v5"

router := chi.NewRouter()
stdAdapter.RegisterRoutes(router, "/auth", handlers)

// Protected routes
router.Route("/api", func(r chi.Router) {
    r.Use(func(next http.Handler) http.Handler {
        return stdAdapter.AuthMiddleware(handlers)(next.ServeHTTP)
    })
    r.Get("/profile", profileHandler)
})
```

## Creating Custom Adapters

If your framework isn't supported yet, you can create a custom adapter by implementing the `HTTPContext` interface:

### Step 1: Implement HTTPContext

```go
package myadapter

import (
    "context"
    "net/http"
    "github.com/mayurrawte/gotrust"
)

type MyFrameworkContext struct {
    // Your framework's context
    frameworkCtx *MyFrameworkCtx
    values map[string]interface{}
}

func (c *MyFrameworkContext) Context() context.Context {
    return c.frameworkCtx.Request().Context()
}

func (c *MyFrameworkContext) GetHeader(key string) string {
    return c.frameworkCtx.Header(key)
}

func (c *MyFrameworkContext) JSON(code int, data interface{}) error {
    return c.frameworkCtx.JSON(code, data)
}

// ... implement all HTTPContext methods
```

### Step 2: Create Wrapper Functions

```go
func WrapHandler(handler gotrust.HTTPHandler) MyFrameworkHandler {
    return func(ctx *MyFrameworkCtx) error {
        wrappedCtx := &MyFrameworkContext{
            frameworkCtx: ctx,
            values: make(map[string]interface{}),
        }
        return handler(wrappedCtx)
    }
}

func WrapMiddleware(middleware gotrust.HTTPMiddleware) MyFrameworkMiddleware {
    // Convert GoTrust middleware to your framework's middleware
}
```

### Step 3: Create Registration Function

```go
func RegisterRoutes(app *MyFramework, basePath string, handlers *gotrust.GenericAuthHandlers) {
    app.POST(basePath+"/signup", WrapHandler(handlers.SignUpHandler))
    app.POST(basePath+"/signin", WrapHandler(handlers.SignInHandler))
    app.POST(basePath+"/refresh", WrapHandler(handlers.RefreshTokenHandler))
    // ... register all routes
}
```

## Performance Considerations

### Adapter Overhead

Adapters add minimal overhead (typically < 1μs per request):
- Simple interface wrapping
- No reflection or heavy computation
- Direct method calls in most cases

### Benchmarks

```
BenchmarkEchoAdapter-8       5000000      215 ns/op
BenchmarkGinAdapter-8        5000000      198 ns/op  
BenchmarkStdlibAdapter-8     10000000     152 ns/op
```

## Choosing an Adapter

| Framework | Best For | Pros | Cons |
|-----------|----------|------|------|
| **Echo** | REST APIs | Fast, minimal, great middleware | Smaller ecosystem |
| **Gin** | High performance | Fastest, popular, lots of middleware | Less idiomatic |
| **Fiber** | Express.js developers | Familiar API, very fast | Different paradigm |
| **net/http** | Simplicity | No dependencies, standard library | More boilerplate |

## Troubleshooting

### "Cannot find adapter package"

Make sure to install the adapter separately:
```bash
go get github.com/mayurrawte/gotrust/adapters/yourframework
```

### "Type mismatch" errors

Ensure you're using the correct wrapper functions:
```go
// Wrong
api.Use(handlers.AuthMiddleware())

// Correct
api.Use(adapterPackage.WrapMiddleware(handlers.AuthMiddleware()))
```

### Context values not available

Check that middleware is properly wrapped and context values are being set:
```go
// In your handler
userID := ctx.Get("user_id").(string) // Should work after auth middleware
```

## Contributing an Adapter

Want to add support for another framework? See our [Adapter Development Guide](Adapter-Development) or check existing adapters as examples.