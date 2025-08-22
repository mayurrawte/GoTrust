package gotrust

import (
	"net/http"
	"net/url"

	"github.com/labstack/echo/v4"
)

// AuthHandlers provides HTTP handlers for authentication
type AuthHandlers struct {
	authService *AuthService
	config      *Config
}

// NewAuthHandlers creates new authentication handlers
func NewAuthHandlers(authService *AuthService, config *Config) *AuthHandlers {
	return &AuthHandlers{
		authService: authService,
		config:      config,
	}
}

// SignUpHandler handles user registration
func (h *AuthHandlers) SignUpHandler(c echo.Context) error {
	var req SignUpRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Sign up user
	response, err := h.authService.SignUp(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusCreated, response)
}

// SignInHandler handles user login
func (h *AuthHandlers) SignInHandler(c echo.Context) error {
	var req SignInRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Validate request
	if err := c.Validate(req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
	}

	// Sign in user
	response, err := h.authService.SignIn(c.Request().Context(), &req)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// RefreshTokenHandler handles token refresh
func (h *AuthHandlers) RefreshTokenHandler(c echo.Context) error {
	var req struct {
		RefreshToken string `json:"refresh_token" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request body",
		})
	}

	// Refresh token
	response, err := h.authService.RefreshToken(c.Request().Context(), req.RefreshToken)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": err.Error(),
		})
	}

	return c.JSON(http.StatusOK, response)
}

// LogoutHandler handles user logout
func (h *AuthHandlers) LogoutHandler(c echo.Context) error {
	// Get session ID from context
	sessionID, _ := GetSessionFromContext(c)

	// Logout
	if err := h.authService.Logout(c.Request().Context(), sessionID); err != nil {
		// Log error but return success
		c.Logger().Error("Failed to logout:", err)
	}

	return c.JSON(http.StatusOK, map[string]string{
		"message": "Successfully logged out",
	})
}

// GetUserHandler returns current user info
func (h *AuthHandlers) GetUserHandler(c echo.Context) error {
	userID, err := GetUserFromContext(c)
	if err != nil {
		return c.JSON(http.StatusUnauthorized, map[string]string{
			"error": "User not authenticated",
		})
	}

	email, _ := c.Get("user_email").(string)
	name, _ := c.Get("user_name").(string)
	provider, _ := c.Get("user_provider").(string)

	return c.JSON(http.StatusOK, map[string]interface{}{
		"user_id":  userID,
		"email":    email,
		"name":     name,
		"provider": provider,
	})
}

// OAuthHandler initiates OAuth flow
func (h *AuthHandlers) OAuthHandler(provider string) echo.HandlerFunc {
	return func(c echo.Context) error {
		var oauthProvider OAuthProvider
		switch provider {
		case "google":
			oauthProvider = ProviderGoogle
		case "github":
			oauthProvider = ProviderGitHub
		default:
			return c.JSON(http.StatusBadRequest, map[string]string{
				"error": "Unsupported provider",
			})
		}

		// Get redirect URI from query parameter
		redirectURI := c.QueryParam("redirect_uri")
		if redirectURI == "" {
			redirectURI = h.config.FrontendSuccessURL
		}

		// Get OAuth URL
		authURL, err := h.authService.GetOAuthURL(oauthProvider, redirectURI)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": err.Error(),
			})
		}

		// Redirect to OAuth provider
		return c.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

// OAuthCallbackHandler handles OAuth callback
func (h *AuthHandlers) OAuthCallbackHandler(provider string) echo.HandlerFunc {
	return func(c echo.Context) error {
		var oauthProvider OAuthProvider
		switch provider {
		case "google":
			oauthProvider = ProviderGoogle
		case "github":
			oauthProvider = ProviderGitHub
		default:
			return h.redirectWithError(c, "unsupported_provider")
		}

		// Get state and code
		state := c.QueryParam("state")
		code := c.QueryParam("code")

		if state == "" {
			return h.redirectWithError(c, "state_missing")
		}

		if code == "" {
			return h.redirectWithError(c, "code_missing")
		}

		// Handle OAuth callback
		response, err := h.authService.OAuthSignIn(c.Request().Context(), oauthProvider, state, code)
		if err != nil {
			return h.redirectWithError(c, err.Error())
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

		return c.Redirect(http.StatusTemporaryRedirect, callbackURL.String())
	}
}

// Helper method to redirect with error
func (h *AuthHandlers) redirectWithError(c echo.Context, errorMsg string) error {
	errorURL, _ := url.Parse(h.config.FrontendErrorURL)
	query := errorURL.Query()
	query.Set("error", errorMsg)
	errorURL.RawQuery = query.Encode()

	return c.Redirect(http.StatusTemporaryRedirect, errorURL.String())
}

// RegisterRoutes registers all auth routes
func (h *AuthHandlers) RegisterRoutes(e *echo.Echo, basePath string) {
	auth := e.Group(basePath)

	// Local auth
	auth.POST("/signup", h.SignUpHandler)
	auth.POST("/signin", h.SignInHandler)
	auth.POST("/refresh", h.RefreshTokenHandler)
	auth.POST("/logout", h.LogoutHandler, h.authService.OptionalAuthMiddleware())
	auth.GET("/user", h.GetUserHandler, h.authService.AuthMiddleware())

	// OAuth
	auth.GET("/google", h.OAuthHandler("google"))
	auth.GET("/google/callback", h.OAuthCallbackHandler("google"))
	auth.GET("/github", h.OAuthHandler("github"))
	auth.GET("/github/callback", h.OAuthCallbackHandler("github"))
}
