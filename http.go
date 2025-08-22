package gotrust

import (
	"context"
	"net/http"
)

// HTTPContext is a framework-agnostic interface for handling HTTP requests
type HTTPContext interface {
	// Request operations
	Context() context.Context
	Request() *http.Request
	GetHeader(key string) string
	GetQueryParam(key string) string
	GetFormValue(key string) string
	Bind(dest interface{}) error
	
	// Response operations
	SetHeader(key, value string)
	SetStatus(code int)
	JSON(code int, data interface{}) error
	Redirect(code int, url string) error
	String(code int, text string) error
	
	// Cookie operations
	GetCookie(name string) (*http.Cookie, error)
	SetCookie(cookie *http.Cookie)
	
	// Context values (for middleware)
	Set(key string, value interface{})
	Get(key string) interface{}
}

// HTTPHandler is a generic handler function
type HTTPHandler func(HTTPContext) error

// HTTPMiddleware is a generic middleware function
type HTTPMiddleware func(HTTPHandler) HTTPHandler

// Router interface for registering routes
type Router interface {
	GET(path string, handler HTTPHandler, middleware ...HTTPMiddleware)
	POST(path string, handler HTTPHandler, middleware ...HTTPMiddleware)
	PUT(path string, handler HTTPHandler, middleware ...HTTPMiddleware)
	DELETE(path string, handler HTTPHandler, middleware ...HTTPMiddleware)
	Group(prefix string, middleware ...HTTPMiddleware) Router
}

// Validator interface for request validation
type Validator interface {
	Validate(interface{}) error
}