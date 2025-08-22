package gotrust

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
)

// AuthMiddleware validates JWT tokens and sets user context
func (a *AuthService) AuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header is required",
				})
			}
			
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Bearer token is required",
				})
			}
			
			// Validate token
			claims, err := a.jwtManager.ValidateToken(tokenString)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token: " + err.Error(),
				})
			}
			
			// Set user context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_name", claims.Name)
			c.Set("user_provider", claims.Provider)
			c.Set("claims", claims)
			
			return next(c)
		}
	}
}

// OptionalAuthMiddleware allows both authenticated and unauthenticated requests
func (a *AuthService) OptionalAuthMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			authHeader := c.Request().Header.Get("Authorization")
			
			// If no auth header, continue without authentication
			if authHeader == "" {
				return next(c)
			}
			
			// If auth header exists but is invalid format, continue without authentication
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return next(c)
			}
			
			// Try to validate token
			claims, err := a.jwtManager.ValidateToken(tokenString)
			if err != nil {
				// Invalid token, continue without authentication
				return next(c)
			}
			
			// Set user context
			c.Set("user_id", claims.UserID)
			c.Set("user_email", claims.Email)
			c.Set("user_name", claims.Name)
			c.Set("user_provider", claims.Provider)
			c.Set("claims", claims)
			
			return next(c)
		}
	}
}

// SessionMiddleware validates session-based authentication
func (a *AuthService) SessionMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get session ID from cookie or header
			sessionID := ""
			
			// Try cookie first
			cookie, err := c.Cookie("session_id")
			if err == nil && cookie != nil {
				sessionID = cookie.Value
			}
			
			// Fallback to header
			if sessionID == "" {
				sessionID = c.Request().Header.Get("X-Session-ID")
			}
			
			if sessionID == "" {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Session ID is required",
				})
			}
			
			// Validate session
			sessionData, err := a.sessionManager.GetSession(c.Request().Context(), sessionID)
			if err != nil {
				return c.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid session: " + err.Error(),
				})
			}
			
			// Set user context
			c.Set("user_id", sessionData.UserID)
			c.Set("user_email", sessionData.Email)
			c.Set("session_id", sessionID)
			
			return next(c)
		}
	}
}

// GetUserFromContext extracts user ID from Echo context
func GetUserFromContext(c echo.Context) (string, error) {
	userID, ok := c.Get("user_id").(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "User not authenticated")
	}
	return userID, nil
}

// GetSessionFromContext extracts session ID from Echo context
func GetSessionFromContext(c echo.Context) (string, error) {
	sessionID, ok := c.Get("session_id").(string)
	if !ok {
		return "", echo.NewHTTPError(http.StatusUnauthorized, "Session not found")
	}
	return sessionID, nil
}