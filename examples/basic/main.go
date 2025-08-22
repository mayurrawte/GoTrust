package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/mayurrawte/gotrust"
)

// InMemoryUserStore implements UserStore interface for demonstration
type InMemoryUserStore struct {
	mu        sync.RWMutex
	users     map[string]*gotrust.User
	passwords map[string]string
}

func NewInMemoryUserStore() *InMemoryUserStore {
	return &InMemoryUserStore{
		users:     make(map[string]*gotrust.User),
		passwords: make(map[string]string),
	}
}

func (s *InMemoryUserStore) CreateUser(ctx context.Context, user *gotrust.User, hashedPassword string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if user already exists
	if _, exists := s.users[user.Email]; exists {
		return fmt.Errorf("user with email %s already exists", user.Email)
	}

	s.users[user.Email] = user
	if hashedPassword != "" {
		s.passwords[user.Email] = hashedPassword
	}
	
	log.Printf("User created: %s", user.Email)
	return nil
}

func (s *InMemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[email]
	if !exists {
		return nil, "", fmt.Errorf("user not found")
	}

	password := s.passwords[email]
	return user, password, nil
}

func (s *InMemoryUserStore) GetUserByID(ctx context.Context, userID string) (*gotrust.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	for _, user := range s.users {
		if user.ID == userID {
			return user, nil
		}
	}
	return nil, fmt.Errorf("user not found")
}

func (s *InMemoryUserStore) UpdateUser(ctx context.Context, user *gotrust.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.Email]; !exists {
		return fmt.Errorf("user not found")
	}

	s.users[user.Email] = user
	log.Printf("User updated: %s", user.Email)
	return nil
}

func (s *InMemoryUserStore) UserExists(ctx context.Context, email string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	_, exists := s.users[email]
	return exists, nil
}

func main() {
	// Create configuration
	config := gotrust.NewConfig()
	
	// Ensure JWT secret is set
	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}

	// Create stores
	userStore := NewInMemoryUserStore()
	sessionStore := gotrust.NewMemorySessionStore()

	// Create auth service
	authService := gotrust.NewAuthService(config, userStore, sessionStore)

	// Setup Echo server
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORS())

	// Register auth routes
	handlers := gotrust.NewAuthHandlers(authService, config)
	handlers.RegisterRoutes(e, "/auth")

	// Public routes
	e.GET("/", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"message": "Welcome to GoTrust Example API",
			"version": "1.0.0",
		})
	})

	// Protected routes
	api := e.Group("/api")
	api.Use(authService.AuthMiddleware())

	api.GET("/profile", func(c echo.Context) error {
		userID, _ := gotrust.GetUserFromContext(c)
		email := c.Get("user_email").(string)
		
		return c.JSON(200, map[string]interface{}{
			"user_id": userID,
			"email":   email,
			"message": "This is your protected profile",
		})
	})

	api.GET("/dashboard", func(c echo.Context) error {
		userID, _ := gotrust.GetUserFromContext(c)
		return c.JSON(200, map[string]interface{}{
			"user_id": userID,
			"data": map[string]interface{}{
				"stats": map[string]int{
					"views":    1234,
					"clicks":   567,
					"sessions": 89,
				},
			},
		})
	})

	// Optional auth routes (works for both authenticated and anonymous)
	public := e.Group("/public")
	public.Use(authService.OptionalAuthMiddleware())

	public.GET("/content", func(c echo.Context) error {
		userID, _ := c.Get("user_id").(string)
		
		response := map[string]interface{}{
			"content": "This is public content",
		}
		
		if userID != "" {
			response["personalized"] = true
			response["user_id"] = userID
			response["message"] = "Welcome back!"
		} else {
			response["personalized"] = false
			response["message"] = "Sign in for personalized content"
		}
		
		return c.JSON(200, response)
	})

	// Health check
	e.GET("/health", func(c echo.Context) error {
		return c.JSON(200, map[string]string{
			"status": "healthy",
		})
	})

	// Start server
	port := ":8080"
	log.Printf("Server starting on http://localhost%s", port)
	log.Println("Test with: curl -X POST http://localhost:8080/auth/signup -H 'Content-Type: application/json' -d '{\"email\":\"test@example.com\",\"password\":\"password123\"}'")

	if err := e.Start(port); err != nil {
		log.Fatal("Server failed to start:", err)
	}
}