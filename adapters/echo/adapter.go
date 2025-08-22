package echo

import (
	"context"
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/mayurrawte/gotrust"
)

// EchoContext wraps echo.Context to implement gotrust.HTTPContext
type EchoContext struct {
	echo.Context
}

// Context returns the request context
func (e *EchoContext) Context() context.Context {
	return e.Request().Context()
}

// GetHeader gets a request header
func (e *EchoContext) GetHeader(key string) string {
	return e.Request().Header.Get(key)
}

// GetQueryParam gets a query parameter
func (e *EchoContext) GetQueryParam(key string) string {
	return e.QueryParam(key)
}

// GetFormValue gets a form value
func (e *EchoContext) GetFormValue(key string) string {
	return e.FormValue(key)
}

// SetHeader sets a response header
func (e *EchoContext) SetHeader(key, value string) {
	e.Response().Header().Set(key, value)
}

// SetStatus sets the response status code
func (e *EchoContext) SetStatus(code int) {
	e.Response().Status = code
}

// GetCookie gets a cookie
func (e *EchoContext) GetCookie(name string) (*http.Cookie, error) {
	return e.Cookie(name)
}

// SetCookie sets a cookie
func (e *EchoContext) SetCookie(cookie *http.Cookie) {
	e.Context.SetCookie(cookie)
}

// WrapHandler converts a gotrust.HTTPHandler to echo.HandlerFunc
func WrapHandler(handler gotrust.HTTPHandler) echo.HandlerFunc {
	return func(c echo.Context) error {
		ctx := &EchoContext{Context: c}
		return handler(ctx)
	}
}

// WrapMiddleware converts a gotrust.HTTPMiddleware to echo.MiddlewareFunc
func WrapMiddleware(middleware gotrust.HTTPMiddleware) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			nextHandler := func(ctx gotrust.HTTPContext) error {
				// Extract echo context and call next
				if echoCtx, ok := ctx.(*EchoContext); ok {
					return next(echoCtx.Context)
				}
				return next(c)
			}
			
			wrappedNext := middleware(nextHandler)
			ctx := &EchoContext{Context: c}
			return wrappedNext(ctx)
		}
	}
}

// EchoRouter wraps echo.Group to implement gotrust.Router
type EchoRouter struct {
	group *echo.Group
}

// NewEchoRouter creates a new Echo router wrapper
func NewEchoRouter(group *echo.Group) *EchoRouter {
	return &EchoRouter{group: group}
}

// GET registers a GET route
func (r *EchoRouter) GET(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, m := range middleware {
		echoMiddleware[i] = WrapMiddleware(m)
	}
	r.group.GET(path, WrapHandler(handler), echoMiddleware...)
}

// POST registers a POST route
func (r *EchoRouter) POST(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, m := range middleware {
		echoMiddleware[i] = WrapMiddleware(m)
	}
	r.group.POST(path, WrapHandler(handler), echoMiddleware...)
}

// PUT registers a PUT route
func (r *EchoRouter) PUT(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, m := range middleware {
		echoMiddleware[i] = WrapMiddleware(m)
	}
	r.group.PUT(path, WrapHandler(handler), echoMiddleware...)
}

// DELETE registers a DELETE route
func (r *EchoRouter) DELETE(path string, handler gotrust.HTTPHandler, middleware ...gotrust.HTTPMiddleware) {
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, m := range middleware {
		echoMiddleware[i] = WrapMiddleware(m)
	}
	r.group.DELETE(path, WrapHandler(handler), echoMiddleware...)
}

// Group creates a new route group
func (r *EchoRouter) Group(prefix string, middleware ...gotrust.HTTPMiddleware) gotrust.Router {
	echoMiddleware := make([]echo.MiddlewareFunc, len(middleware))
	for i, m := range middleware {
		echoMiddleware[i] = WrapMiddleware(m)
	}
	newGroup := r.group.Group(prefix, echoMiddleware...)
	return NewEchoRouter(newGroup)
}

// RegisterRoutes registers all auth routes on an Echo instance
func RegisterRoutes(e *echo.Echo, basePath string, handlers *gotrust.GenericAuthHandlers) {
	auth := e.Group(basePath)
	router := NewEchoRouter(auth)
	
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