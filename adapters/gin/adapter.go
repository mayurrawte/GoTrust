package gin

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mayurrawte/gotrust"
)

// GinContext wraps gin.Context to implement gotrust.HTTPContext
type GinContext struct {
	*gin.Context
}

// Context returns the request context
func (g *GinContext) Context() context.Context {
	return g.Request.Context()
}

// GetHeader gets a request header
func (g *GinContext) GetHeader(key string) string {
	return g.GetHeader(key)
}

// GetQueryParam gets a query parameter
func (g *GinContext) GetQueryParam(key string) string {
	return g.Query(key)
}

// GetFormValue gets a form value
func (g *GinContext) GetFormValue(key string) string {
	return g.PostForm(key)
}

// Bind decodes request body
func (g *GinContext) Bind(dest interface{}) error {
	return g.ShouldBindJSON(dest)
}

// SetHeader sets a response header
func (g *GinContext) SetHeader(key, value string) {
	g.Header(key, value)
}

// SetStatus sets the response status code
func (g *GinContext) SetStatus(code int) {
	g.Status(code)
}

// JSON sends a JSON response
func (g *GinContext) JSON(code int, data interface{}) error {
	g.Context.JSON(code, data)
	return nil
}

// Redirect sends a redirect response
func (g *GinContext) Redirect(code int, url string) error {
	g.Context.Redirect(code, url)
	return nil
}

// String sends a text response
func (g *GinContext) String(code int, text string) error {
	g.Context.String(code, text)
	return nil
}

// GetCookie gets a cookie
func (g *GinContext) GetCookie(name string) (*http.Cookie, error) {
	value, err := g.Cookie(name)
	if err != nil {
		return nil, err
	}
	return &http.Cookie{Name: name, Value: value}, nil
}

// SetCookie sets a cookie
func (g *GinContext) SetCookie(cookie *http.Cookie) {
	g.Context.SetCookie(
		cookie.Name,
		cookie.Value,
		cookie.MaxAge,
		cookie.Path,
		cookie.Domain,
		cookie.Secure,
		cookie.HttpOnly,
	)
}

// WrapHandler converts a gotrust.HTTPHandler to gin.HandlerFunc
func WrapHandler(handler gotrust.HTTPHandler) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &GinContext{Context: c}
		if err := handler(ctx); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

// WrapMiddleware converts a gotrust.HTTPMiddleware to gin.HandlerFunc
func WrapMiddleware(middleware gotrust.HTTPMiddleware) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx := &GinContext{Context: c}
		
		nextHandler := func(httpCtx gotrust.HTTPContext) error {
			c.Next()
			return nil
		}
		
		wrappedNext := middleware(nextHandler)
		if err := wrappedNext(ctx); err != nil {
			c.Abort()
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		}
	}
}

// GinRouter wraps gin.RouterGroup to implement gotrust.Router
type GinRouter struct {
	group *gin.RouterGroup
}

// NewGinRouter creates a new Gin router wrapper
func NewGinRouter(group *gin.RouterGroup) *GinRouter {
	return &GinRouter{group: group}
}

// GET registers a GET route
func (r *GinRouter) GET(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	handlers := make([]gin.HandlerFunc, len(middleware)+1)
	for i, m := range middleware {
		handlers[i] = WrapMiddleware(m)
	}
	handlers[len(middleware)] = WrapHandler(handler)
	r.group.GET(path, handlers...)
}

// POST registers a POST route
func (r *GinRouter) POST(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	handlers := make([]gin.HandlerFunc, len(middleware)+1)
	for i, m := range middleware {
		handlers[i] = WrapMiddleware(m)
	}
	handlers[len(middleware)] = WrapHandler(handler)
	r.group.POST(path, handlers...)
}

// PUT registers a PUT route
func (r *GinRouter) PUT(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	handlers := make([]gin.HandlerFunc, len(middleware)+1)
	for i, m := range middleware {
		handlers[i] = WrapMiddleware(m)
	}
	handlers[len(middleware)] = WrapHandler(handler)
	r.group.PUT(path, handlers...)
}

// DELETE registers a DELETE route
func (r *GinRouter) DELETE(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	handlers := make([]gin.HandlerFunc, len(middleware)+1)
	for i, m := range middleware {
		handlers[i] = WrapMiddleware(m)
	}
	handlers[len(middleware)] = WrapHandler(handler)
	r.group.DELETE(path, handlers...)
}

// Group creates a new route group
func (r *GinRouter) Group(prefix string, middleware ...gotrust.HTTPMiddleware) gotrust.Router {
	ginMiddleware := make([]gin.HandlerFunc, len(middleware))
	for i, m := range middleware {
		ginMiddleware[i] = WrapMiddleware(m)
	}
	newGroup := r.group.Group(prefix, ginMiddleware...)
	return NewGinRouter(newGroup)
}

// RegisterRoutes registers all auth routes on a Gin engine
func RegisterRoutes(router *gin.Engine, basePath string, handlers *gotrust.GenericAuthHandlers) {
	auth := router.Group(basePath)
	r := NewGinRouter(auth)
	
	// Local auth
	r.POST("/signup", handlers.SignUpHandler)
	r.POST("/signin", handlers.SignInHandler)
	r.POST("/refresh", handlers.RefreshTokenHandler)
	r.POST("/logout", handlers.LogoutHandler, handlers.OptionalAuthMiddleware())
	r.GET("/user", handlers.GetUserHandler, handlers.AuthMiddleware())
	
	// OAuth
	r.GET("/google", handlers.OAuthHandler("google"))
	r.GET("/google/callback", handlers.OAuthCallbackHandler("google"))
	r.GET("/github", handlers.OAuthHandler("github"))
	r.GET("/github/callback", handlers.OAuthCallbackHandler("github"))
}