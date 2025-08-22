package gotrust

import (
	"os"
	"time"
)

type Config struct {
	// JWT Configuration
	JWTSecret        string
	JWTExpiration    time.Duration
	JWTIssuer        string
	
	// OAuth Google Configuration
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURI  string
	GoogleScopes       []string
	
	// OAuth GitHub Configuration
	GitHubClientID     string
	GitHubClientSecret string
	GitHubRedirectURI  string
	GitHubScopes       []string
	
	// General OAuth Configuration
	OAuthStateExpiration time.Duration
	FrontendSuccessURL   string
	FrontendErrorURL     string
	
	// Redis Configuration (optional)
	RedisURL         string
	EnableRedisCache bool
	
	// Security Settings
	BCryptCost      int
	AllowSignup     bool
	RequireEmailVerification bool
}

func NewConfig() *Config {
	return &Config{
		JWTSecret:            getEnv("JWT_SECRET", ""),
		JWTExpiration:        24 * time.Hour,
		JWTIssuer:           getEnv("JWT_ISSUER", "gotrust"),
		
		GoogleClientID:       getEnv("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret:   getEnv("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURI:    getEnv("GOOGLE_REDIRECT_URI", "http://localhost:4000/auth/google/callback"),
		GoogleScopes:         []string{"email", "profile"},
		
		GitHubClientID:       getEnv("GITHUB_CLIENT_ID", ""),
		GitHubClientSecret:   getEnv("GITHUB_CLIENT_SECRET", ""),
		GitHubRedirectURI:    getEnv("GITHUB_REDIRECT_URI", "http://localhost:4000/auth/github/callback"),
		GitHubScopes:         []string{"user:email"},
		
		OAuthStateExpiration: 10 * time.Minute,
		FrontendSuccessURL:   getEnv("FRONTEND_SUCCESS_URL", "http://localhost:3000/auth/success"),
		FrontendErrorURL:     getEnv("FRONTEND_ERROR_URL", "http://localhost:3000/auth/error"),
		
		RedisURL:         getEnv("REDIS_URL", ""),
		EnableRedisCache: getEnv("ENABLE_REDIS_CACHE", "true") == "true",
		
		BCryptCost:               10,
		AllowSignup:              getEnv("ALLOW_SIGNUP", "true") == "true",
		RequireEmailVerification: getEnv("REQUIRE_EMAIL_VERIFICATION", "false") == "true",
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}