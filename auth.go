package gotrust

import (
	"context"
	"fmt"
	"time"

	"golang.org/x/crypto/bcrypt"
)

// UserStore interface for user persistence
type UserStore interface {
	CreateUser(ctx context.Context, user *User, hashedPassword string) error
	GetUserByEmail(ctx context.Context, email string) (*User, string, error) // returns user and hashed password
	GetUserByID(ctx context.Context, userID string) (*User, error)
	UpdateUser(ctx context.Context, user *User) error
	UserExists(ctx context.Context, email string) (bool, error)
}

// AuthService handles authentication operations
type AuthService struct {
	config         *Config
	userStore      UserStore
	sessionManager *SessionManager
	jwtManager     *JWTManager
	oauthManager   *OAuthManager
}

// NewAuthService creates a new authentication service
func NewAuthService(config *Config, userStore UserStore, sessionStore SessionStore) *AuthService {
	return &AuthService{
		config:         config,
		userStore:      userStore,
		sessionManager: NewSessionManager(sessionStore, "session"),
		jwtManager:     NewJWTManager(config.JWTSecret, config.JWTIssuer, config.JWTExpiration),
		oauthManager:   NewOAuthManager(config, sessionStore),
	}
}

// SignUp registers a new user with email and password
func (a *AuthService) SignUp(ctx context.Context, req *SignUpRequest) (*AuthResponse, error) {
	if !a.config.AllowSignup {
		return nil, fmt.Errorf("signup is disabled")
	}
	
	// Check if user already exists
	exists, err := a.userStore.UserExists(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	
	if exists {
		return nil, fmt.Errorf("user already exists")
	}
	
	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), a.config.BCryptCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}
	
	// Create user
	user := &User{
		ID:        generateRandomString(16),
		Email:     req.Email,
		Name:      req.Name,
		Provider:  string(ProviderLocal),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	if err := a.userStore.CreateUser(ctx, user, string(hashedPassword)); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}
	
	// Generate tokens
	return a.generateAuthResponse(ctx, user)
}

// SignIn authenticates a user with email and password
func (a *AuthService) SignIn(ctx context.Context, req *SignInRequest) (*AuthResponse, error) {
	// Get user and password hash
	user, hashedPassword, err := a.userStore.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	
	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}
	
	// Generate tokens
	return a.generateAuthResponse(ctx, user)
}

// OAuthSignIn handles OAuth authentication
func (a *AuthService) OAuthSignIn(ctx context.Context, provider OAuthProvider, state, code string) (*AuthResponse, error) {
	// Validate OAuth callback
	oauthUser, _, err := a.oauthManager.ValidateCallback(provider, state, code)
	if err != nil {
		return nil, fmt.Errorf("oauth validation failed: %w", err)
	}
	
	if oauthUser.Email == "" {
		return nil, fmt.Errorf("email is required from OAuth provider")
	}
	
	// Check if user exists
	user, _, err := a.userStore.GetUserByEmail(ctx, oauthUser.Email)
	if err != nil {
		// Create new user from OAuth
		user = &User{
			ID:        fmt.Sprintf("%s_%s", provider, oauthUser.ID),
			Email:     oauthUser.Email,
			Name:      oauthUser.Name,
			AvatarURL: oauthUser.AvatarURL,
			Provider:  oauthUser.Provider,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		
		if err := a.userStore.CreateUser(ctx, user, ""); err != nil {
			return nil, fmt.Errorf("failed to create OAuth user: %w", err)
		}
	} else {
		// Update existing user
		user.Name = oauthUser.Name
		user.AvatarURL = oauthUser.AvatarURL
		user.UpdatedAt = time.Now()
		
		if err := a.userStore.UpdateUser(ctx, user); err != nil {
			// Log error but continue
			fmt.Printf("Failed to update user: %v\n", err)
		}
	}
	
	// Generate tokens
	return a.generateAuthResponse(ctx, user)
}

// RefreshToken generates new access token from refresh token
func (a *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*AuthResponse, error) {
	// Validate refresh token
	userID, err := a.jwtManager.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}
	
	// Get user
	user, err := a.userStore.GetUserByID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	
	// Generate new tokens
	return a.generateAuthResponse(ctx, user)
}

// ValidateToken validates an access token and returns claims
func (a *AuthService) ValidateToken(token string) (*TokenClaims, error) {
	return a.jwtManager.ValidateToken(token)
}

// GetOAuthURL generates OAuth authorization URL
func (a *AuthService) GetOAuthURL(provider OAuthProvider, redirectURI string) (string, error) {
	if redirectURI == "" {
		redirectURI = a.config.FrontendSuccessURL
	}
	return a.oauthManager.GetAuthURL(provider, redirectURI)
}

// Logout invalidates a session
func (a *AuthService) Logout(ctx context.Context, sessionID string) error {
	if sessionID != "" {
		return a.sessionManager.InvalidateSession(ctx, sessionID)
	}
	return nil
}

// LogoutAllSessions invalidates all sessions for a user
func (a *AuthService) LogoutAllSessions(ctx context.Context, userID string) error {
	return a.sessionManager.InvalidateUserSessions(ctx, userID)
}

// GetSession retrieves session data
func (a *AuthService) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	return a.sessionManager.GetSession(ctx, sessionID)
}

// Helper method to generate auth response with tokens
func (a *AuthService) generateAuthResponse(ctx context.Context, user *User) (*AuthResponse, error) {
	// Generate access token
	claims := TokenClaims{
		UserID:   user.ID,
		Email:    user.Email,
		Name:     user.Name,
		Provider: user.Provider,
	}
	
	accessToken, err := a.jwtManager.GenerateToken(claims)
	if err != nil {
		return nil, fmt.Errorf("failed to generate access token: %w", err)
	}
	
	// Generate refresh token
	refreshToken, err := a.jwtManager.GenerateRefreshToken(user.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to generate refresh token: %w", err)
	}
	
	// Create session
	_, err = a.sessionManager.CreateSession(ctx, user.ID, user.Email, a.config.JWTExpiration)
	if err != nil {
		// Log error but don't fail authentication
		fmt.Printf("Failed to create session: %v\n", err)
	}
	
	return &AuthResponse{
		User:         user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresIn:    int64(a.config.JWTExpiration.Seconds()),
	}, nil
}