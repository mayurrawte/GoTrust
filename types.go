package gotrust

import "time"

// User represents a user in the system
type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name,omitempty"`
	AvatarURL string    `json:"avatar_url,omitempty"`
	Provider  string    `json:"provider,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// AuthResponse is returned after successful authentication
type AuthResponse struct {
	User        *User  `json:"user"`
	AccessToken string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	ExpiresIn   int64  `json:"expires_in"`
}

// SignUpRequest for email/password registration
type SignUpRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
	Name     string `json:"name,omitempty"`
}

// SignInRequest for email/password login
type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// OAuthProvider represents an OAuth provider
type OAuthProvider string

const (
	ProviderGoogle OAuthProvider = "google"
	ProviderGitHub OAuthProvider = "github"
	ProviderLocal  OAuthProvider = "local"
)

// OAuthUserInfo contains user information from OAuth providers
type OAuthUserInfo struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	AvatarURL string `json:"avatar_url"`
	Provider  string `json:"provider"`
}

// TokenClaims represents JWT token claims
type TokenClaims struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name,omitempty"`
	Provider string `json:"provider,omitempty"`
}

// SessionData represents session information
type SessionData struct {
	UserID    string    `json:"user_id"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// OAuthState represents OAuth state data
type OAuthState struct {
	State       string    `json:"state"`
	RedirectURI string    `json:"redirect_uri"`
	ExpiresAt   time.Time `json:"expires_at"`
}