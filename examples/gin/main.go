package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/mayurrawte/gotrust"
	ginAdapter "github.com/mayurrawte/gotrust/adapters/gin"
)

// Simple in-memory user store for demo
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
	
	if _, exists := s.users[user.Email]; exists {
		return fmt.Errorf("user already exists")
	}
	
	s.users[user.Email] = user
	s.passwords[user.Email] = hashedPassword
	return nil
}

func (s *InMemoryUserStore) GetUserByEmail(ctx context.Context, email string) (*gotrust.User, string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	user, exists := s.users[email]
	if !exists {
		return nil, "", fmt.Errorf("user not found")
	}
	
	return user, s.passwords[email], nil
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
	
	s.users[user.Email] = user
	return nil
}

func (s *InMemoryUserStore) UserExists(ctx context.Context, email string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	_, exists := s.users[email]
	return exists, nil
}

func main() {
	// Setup GoTrust
	config := gotrust.NewConfig()
	if config.JWTSecret == "" {
		log.Fatal("JWT_SECRET environment variable is required")
	}
	
	userStore := NewInMemoryUserStore()
	sessionStore := gotrust.NewMemorySessionStore()
	authService := gotrust.NewAuthService(config, userStore, sessionStore)
	handlers := gotrust.NewGenericAuthHandlers(authService, config)
	
	// Setup Gin
	router := gin.Default()
	
	// Register auth routes
	ginAdapter.RegisterRoutes(router, "/auth", handlers)
	
	// Public route
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GoTrust with Gin Framework",
			"framework": "Gin",
		})
	})
	
	// Protected routes
	api := router.Group("/api")
	api.Use(ginAdapter.WrapMiddleware(handlers.AuthMiddleware()))
	
	api.GET("/profile", func(c *gin.Context) {
		userID, _ := c.Get("user_id")
		c.JSON(200, gin.H{
			"user_id": userID,
			"message": "Protected route accessed successfully",
		})
	})
	
	// Start server
	log.Println("Gin server starting on :8080")
	log.Println("Try: curl -X POST http://localhost:8080/auth/signup -H 'Content-Type: application/json' -d '{\"email\":\"test@example.com\",\"password\":\"password123\"}'")
	router.Run(":8080")
}