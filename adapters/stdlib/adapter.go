package stdlib

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/mayurrawte/gotrust"
)

// StdContext wraps http.Request and http.ResponseWriter to implement gotrust.HTTPContext
type StdContext struct {
	Request  *http.Request
	Response http.ResponseWriter
	values   map[string]interface{}
	status   int
}

// NewStdContext creates a new standard library context
func NewStdContext(w http.ResponseWriter, r *http.Request) *StdContext {
	return &StdContext{
		Request:  r,
		Response: w,
		values:   make(map[string]interface{}),
		status:   http.StatusOK,
	}
}

// Context returns the request context
func (c *StdContext) Context() context.Context {
	return c.Request.Context()
}

// GetHeader gets a request header
func (c *StdContext) GetHeader(key string) string {
	return c.Request.Header.Get(key)
}

// GetQueryParam gets a query parameter
func (c *StdContext) GetQueryParam(key string) string {
	return c.Request.URL.Query().Get(key)
}

// GetFormValue gets a form value
func (c *StdContext) GetFormValue(key string) string {
	return c.Request.FormValue(key)
}

// Bind decodes JSON request body
func (c *StdContext) Bind(dest interface{}) error {
	decoder := json.NewDecoder(c.Request.Body)
	return decoder.Decode(dest)
}

// SetHeader sets a response header
func (c *StdContext) SetHeader(key, value string) {
	c.Response.Header().Set(key, value)
}

// SetStatus sets the response status code
func (c *StdContext) SetStatus(code int) {
	c.status = code
}

// JSON sends a JSON response
func (c *StdContext) JSON(code int, data interface{}) error {
	c.Response.Header().Set("Content-Type", "application/json")
	c.Response.WriteHeader(code)
	encoder := json.NewEncoder(c.Response)
	return encoder.Encode(data)
}

// Redirect sends a redirect response
func (c *StdContext) Redirect(code int, url string) error {
	http.Redirect(c.Response, c.Request, url, code)
	return nil
}

// String sends a text response
func (c *StdContext) String(code int, text string) error {
	c.Response.Header().Set("Content-Type", "text/plain")
	c.Response.WriteHeader(code)
	_, err := c.Response.Write([]byte(text))
	return err
}

// GetCookie gets a cookie
func (c *StdContext) GetCookie(name string) (*http.Cookie, error) {
	return c.Request.Cookie(name)
}

// SetCookie sets a cookie
func (c *StdContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Response, cookie)
}

// Set sets a context value
func (c *StdContext) Set(key string, value interface{}) {
	c.values[key] = value
}

// Get gets a context value
func (c *StdContext) Get(key string) interface{} {
	return c.values[key]
}

// WrapHandler converts a gotrust.HTTPHandler to http.HandlerFunc
func WrapHandler(handler gotrust.HTTPHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := NewStdContext(w, r)
		if err := handler(ctx); err != nil {
			// Handle error (you might want to customize this)
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	}
}

// WrapMiddleware converts a gotrust.HTTPMiddleware to standard middleware
func WrapMiddleware(middleware gotrust.HTTPMiddleware) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := NewStdContext(w, r)
			
			nextHandler := func(httpCtx gotrust.HTTPContext) error {
				// Call the next handler
				next(w, r)
				return nil
			}
			
			wrappedNext := middleware(nextHandler)
			if err := wrappedNext(ctx); err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
			}
		}
	}
}

// Router implements gotrust.Router for standard net/http
type Router struct {
	mux        *http.ServeMux
	prefix     string
	middleware []gotrust.HTTPMiddleware
}

// NewRouter creates a new standard library router
func NewRouter(mux *http.ServeMux) *Router {
	return &Router{
		mux:    mux,
		prefix: "",
	}
}

// handle registers a route with middleware chain
func (r *Router) handle(method, path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	fullPath := r.prefix + path
	
	// Build middleware chain
	finalHandler := handler
	allMiddleware := append(r.middleware, middleware...)
	for i := len(allMiddleware) - 1; i >= 0; i-- {
		finalHandler = allMiddleware[i](finalHandler)
	}
	
	r.mux.HandleFunc(fullPath, func(w http.ResponseWriter, req *http.Request) {
		if req.Method != method {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		
		ctx := NewStdContext(w, req)
		if err := finalHandler(ctx); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
	})
}

// GET registers a GET route
func (r *Router) GET(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	r.handle("GET", path, handler, middleware...)
}

// POST registers a POST route
func (r *Router) POST(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	r.handle("POST", path, handler, middleware...)
}

// PUT registers a PUT route
func (r *Router) PUT(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	r.handle("PUT", path, handler, middleware...)
}

// DELETE registers a DELETE route
func (r *Router) DELETE(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	r.handle("DELETE", path, handler, middleware...)
}

// Group creates a new route group
func (r *Router) Group(prefix string, middleware ...gotrust.HTTPMiddleware) gotrust.Router {
	return &Router{
		mux:        r.mux,
		prefix:     r.prefix + prefix,
		middleware: append(r.middleware, middleware...),
	}
}

// RegisterRoutes registers all auth routes on a ServeMux
func RegisterRoutes(mux *http.ServeMux, basePath string, handlers *gotrust.GenericAuthHandlers) {
	router := &Router{
		mux:    mux,
		prefix: basePath,
	}
	
	// Local auth
	router.POST("/signup", handlers.SignUpHandler)
	router.POST("/signin", handlers.SignInHandler)
	router.POST("/refresh", handlers.RefreshTokenHandler)
	router.POST("/logout", handlers.LogoutHandler, handlers.OptionalAuthMiddleware())
	router.GET("/user", handlers.GetUserHandler, handlers.AuthMiddleware())
	
	// OAuth
	router.GET("/google", handlers.OAuthHandler("google"))
	router.GET("/google/callback", handlers.OAuthCallbackHandler("google"))
	router.GET("/github", handlers.OAuthHandler("github"))
	router.GET("/github/callback", handlers.OAuthCallbackHandler("github"))
}

// AuthMiddleware is a convenience function for using auth middleware with standard http
func AuthMiddleware(handlers *gotrust.GenericAuthHandlers) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := NewStdContext(w, r)
			
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ctx.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header is required",
				})
				return
			}
			
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				ctx.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Bearer token is required",
				})
				return
			}
			
			// Validate token using the auth service
			authMiddleware := handlers.AuthMiddleware()
			nextHandler := func(httpCtx gotrust.HTTPContext) error {
				next(w, r)
				return nil
			}
			
			if err := authMiddleware(nextHandler)(ctx); err != nil {
				// Error already handled by middleware
				return
			}
		}
	}
}