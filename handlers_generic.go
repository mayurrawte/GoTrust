package gotrust

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

// GenericAuthHandlers provides framework-agnostic HTTP handlers for authentication
type GenericAuthHandlers struct {
	authService *AuthService
	config      *Config
}

// NewGenericAuthHandlers creates new framework-agnostic authentication handlers
func NewGenericAuthHandlers(authService *AuthService, config *Config) *GenericAuthHandlers {
	return &GenericAuthHandlers{
		authService: authService,
		config:      config,
	}
}

// SignUpHandler handles user registration
func (h *GenericAuthHandlers) SignUpHandler(ctx HTTPContext) error {
	var req SignUpRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}
	
	// Basic validation
	if req.Email == "" || req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}
	
	if len(req.Password) < 6 {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Password must be at least 6 characters",
		})
	}
	
	// Sign up user
	response, err := h.authService.SignUp(ctx.Context(), &req)
	if err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}
	
	return ctx.JSON(http.StatusCreated, response)
}

// SignInHandler handles user login
func (h *GenericAuthHandlers) SignInHandler(ctx HTTPContext) error {
	var req SignInRequest
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}
	
	// Basic validation
	if req.Email == "" || req.Password == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Email and password are required",
		})
	}
	
	// Sign in user
	response, err := h.authService.SignIn(ctx.Context(), &req)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}
	
	return ctx.JSON(http.StatusOK, response)
}

// RefreshTokenHandler handles token refresh
func (h *GenericAuthHandlers) RefreshTokenHandler(ctx HTTPContext) error {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}
	
	if err := ctx.Bind(&req); err != nil {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}
	
	if req.RefreshToken == "" {
		return ctx.JSON(http.StatusBadRequest, map[string]string{
			"error": "Refresh token is required",
		})
	}
	
	// Refresh token
	response, err := h.authService.RefreshToken(ctx.Context(), req.RefreshToken)
	if err != nil {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}
	
	return ctx.JSON(http.StatusOK, response)
}

// LogoutHandler handles user logout
func (h *GenericAuthHandlers) LogoutHandler(ctx HTTPContext) error {
	// Get session ID from context (set by middleware)
	sessionID, _ := ctx.Get("session_id").(string)
	
	// Logout
	if err := h.authService.Logout(ctx.Context(), sessionID); err != nil {
		// Log error but return success
		fmt.Printf("Failed to logout: %v\n", err)
	}
	
	return ctx.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

// GetUserHandler returns current user info
func (h *GenericAuthHandlers) GetUserHandler(ctx HTTPContext) error {
	userID, ok := ctx.Get("user_id").(string)
	if !ok {
		return ctx.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}
	
	email, _ := ctx.Get("user_email").(string)
	name, _ := ctx.Get("user_name").(string)
	provider, _ := ctx.Get("user_provider").(string)
	
	return ctx.JSON(http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"email":    email,
		"name":     name,
		"provider": provider,
	})
}

// OAuthHandler initiates OAuth flow
func (h *GenericAuthHandlers) OAuthHandler(provider string) HTTPHandler {
	return func(ctx HTTPContext) error {
		var oauthProvider OAuthProvider
		switch provider {
		case "google":
			oauthProvider = ProviderGoogle
		case "github":
			oauthProvider = ProviderGitHub
		default:
			return ctx.JSON(http.StatusBadRequest, map[string]string{
				"error": "Unsupported provider",
			})
		}
		
		// Get redirect URI from query parameter
		redirectURI := ctx.GetQueryParam("redirect_uri")
		if redirectURI == "" {
			redirectURI = h.config.FrontendSuccessURL
		}
		
		// Get OAuth URL
		authURL, err := h.authService.GetOAuthURL(oauthProvider, redirectURI)
		if err != nil {
			return ctx.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}
		
		// Redirect to OAuth provider
		return ctx.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

// OAuthCallbackHandler handles OAuth callback
func (h *GenericAuthHandlers) OAuthCallbackHandler(provider string) HTTPHandler {
	return func(ctx HTTPContext) error {
		var oauthProvider OAuthProvider
		switch provider {
		case "google":
			oauthProvider = ProviderGoogle
		case "github":
			oauthProvider = ProviderGitHub
		default:
			return h.redirectWithError(ctx, "unsupported_provider")
		}
		
		// Get state and code
		state := ctx.GetQueryParam("state")
		code := ctx.GetQueryParam("code")
		
		if state == "" {
			return h.redirectWithError(ctx, "state_missing")
		}
		
		if code == "" {
			return h.redirectWithError(ctx, "code_missing")
		}
		
		// Handle OAuth callback
		response, err := h.authService.OAuthSignIn(ctx.Context(), oauthProvider, state, code)
		if err != nil {
			return h.redirectWithError(ctx, err.Error())
		}
		
		// Get redirect URI from OAuth state
		redirectURI := h.config.FrontendSuccessURL
		
		// Build callback URL with auth data
		callbackURL, _ := url.Parse(redirectURI)
		query := callbackURL.Query()
		query.Set("token", response.AccessToken)
		query.Set("refresh_token", response.RefreshToken)
		query.Set("user_id", response.User.ID)
		query.Set("email", response.User.Email)
		query.Set("provider", provider)
		
		if response.User.Name != "" {
			query.Set("name", response.User.Name)
		}
		if response.User.AvatarURL != "" {
			query.Set("avatar_url", response.User.AvatarURL)
		}
		
		callbackURL.RawQuery = query.Encode()
		
		return ctx.Redirect(http.StatusTemporaryRedirect, callbackURL.String())
	}
}

// Helper method to redirect with error
func (h *GenericAuthHandlers) redirectWithError(ctx HTTPContext, errorMsg string) error {
	errorURL, _ := url.Parse(h.config.FrontendErrorURL)
	query := errorURL.Query()
	query.Set("error", errorMsg)
	errorURL.RawQuery = query.Encode()
	
	return ctx.Redirect(http.StatusTemporaryRedirect, errorURL.String())
}

// AuthMiddleware validates JWT tokens and sets user context
func (h *GenericAuthHandlers) AuthMiddleware() HTTPMiddleware {
	return func(next HTTPHandler) HTTPHandler {
		return func(ctx HTTPContext) error {
			authHeader := ctx.GetHeader("Authorization")
			if authHeader == "" {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Authorization header is required",
				})
			}
			
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Bearer token is required",
				})
			}
			
			// Validate token
			claims, err := h.authService.ValidateToken(tokenString)
			if err != nil {
				return ctx.JSON(http.StatusUnauthorized, map[string]string{
					"error": "Invalid token: " + err.Error(),
				})
			}
			
			// Set user context
			ctx.Set("user_id", claims.UserID)
			ctx.Set("user_email", claims.Email)
			ctx.Set("user_name", claims.Name)
			ctx.Set("user_provider", claims.Provider)
			ctx.Set("claims", claims)
			
			return next(ctx)
		}
	}
}

// OptionalAuthMiddleware allows both authenticated and unauthenticated requests
func (h *GenericAuthHandlers) OptionalAuthMiddleware() HTTPMiddleware {
	return func(next HTTPHandler) HTTPHandler {
		return func(ctx HTTPContext) error {
			authHeader := ctx.GetHeader("Authorization")
			
			// If no auth header, continue without authentication
			if authHeader == "" {
				return next(ctx)
			}
			
			// If auth header exists but is invalid format, continue without authentication
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			if tokenString == authHeader {
				return next(ctx)
			}
			
			// Try to validate token
			claims, err := h.authService.ValidateToken(tokenString)
			if err != nil {
				// Invalid token, continue without authentication
				return next(ctx)
			}
			
			// Set user context
			ctx.Set("user_id", claims.UserID)
			ctx.Set("user_email", claims.Email)
			ctx.Set("user_name", claims.Name)
			ctx.Set("user_provider", claims.Provider)
			ctx.Set("claims", claims)
			
			return next(ctx)
		}
	}
}

// GetUserFromContext extracts user ID from context
func GetUserFromContext(ctx HTTPContext) (string, error) {
	userID, ok := ctx.Get("user_id").(string)
	if !ok {
		return "", fmt.Errorf("user not authenticated")
	}
	return userID, nil
}