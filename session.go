package gotrust

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type SessionStore interface {
	Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error
	Get(ctx context.Context, key string, dest interface{}) error
	Delete(ctx context.Context, keys ...string) error
	Exists(ctx context.Context, keys ...string) (bool, error)
}

// RedisSessionStore uses Redis for session storage
type RedisSessionStore struct {
	client *redis.Client
}

func NewRedisSessionStore(redisURL string) (*RedisSessionStore, error) {
	if redisURL == "" {
		return nil, fmt.Errorf("redis URL is required")
	}
	
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis URL: %w", err)
	}
	
	opt.MaxRetries = 3
	opt.DialTimeout = 5 * time.Second
	opt.ReadTimeout = 3 * time.Second
	opt.WriteTimeout = 3 * time.Second
	
	client := redis.NewClient(opt)
	
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	
	if _, err := client.Ping(ctx).Result(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}
	
	return &RedisSessionStore{client: client}, nil
}

func (r *RedisSessionStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	return r.client.Set(ctx, key, data, expiration).Err()
}

func (r *RedisSessionStore) Get(ctx context.Context, key string, dest interface{}) error {
	data, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return fmt.Errorf("key not found")
	} else if err != nil {
		return err
	}
	
	return json.Unmarshal([]byte(data), dest)
}

func (r *RedisSessionStore) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

func (r *RedisSessionStore) Exists(ctx context.Context, keys ...string) (bool, error) {
	count, err := r.client.Exists(ctx, keys...).Result()
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *RedisSessionStore) Close() error {
	return r.client.Close()
}

// MemorySessionStore uses in-memory storage (for development/testing)
type MemorySessionStore struct {
	mu    sync.RWMutex
	store map[string]memoryItem
}

type memoryItem struct {
	value     []byte
	expiresAt time.Time
}

func NewMemorySessionStore() *MemorySessionStore {
	store := &MemorySessionStore{
		store: make(map[string]memoryItem),
	}
	
	// Start cleanup goroutine
	go store.cleanup()
	
	return store
}

func (m *MemorySessionStore) Set(ctx context.Context, key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal value: %w", err)
	}
	
	m.mu.Lock()
	defer m.mu.Unlock()
	
	m.store[key] = memoryItem{
		value:     data,
		expiresAt: time.Now().Add(expiration),
	}
	
	return nil
}

func (m *MemorySessionStore) Get(ctx context.Context, key string, dest interface{}) error {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	item, exists := m.store[key]
	if !exists {
		return fmt.Errorf("key not found")
	}
	
	if time.Now().After(item.expiresAt) {
		delete(m.store, key)
		return fmt.Errorf("key expired")
	}
	
	return json.Unmarshal(item.value, dest)
}

func (m *MemorySessionStore) Delete(ctx context.Context, keys ...string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	
	for _, key := range keys {
		delete(m.store, key)
	}
	
	return nil
}

func (m *MemorySessionStore) Exists(ctx context.Context, keys ...string) (bool, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	
	for _, key := range keys {
		if item, exists := m.store[key]; exists {
			if time.Now().After(item.expiresAt) {
				continue
			}
			return true, nil
		}
	}
	
	return false, nil
}

func (m *MemorySessionStore) cleanup() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()
	
	for range ticker.C {
		m.mu.Lock()
		now := time.Now()
		for key, item := range m.store {
			if now.After(item.expiresAt) {
				delete(m.store, key)
			}
		}
		m.mu.Unlock()
	}
}

// SessionManager handles session operations
type SessionManager struct {
	store SessionStore
	prefix string
}

func NewSessionManager(store SessionStore, prefix string) *SessionManager {
	if prefix == "" {
		prefix = "session"
	}
	return &SessionManager{
		store:  store,
		prefix: prefix,
	}
}

func (s *SessionManager) CreateSession(ctx context.Context, userID, email string, duration time.Duration) (string, error) {
	sessionID := generateRandomString(32)
	
	sessionData := &SessionData{
		UserID:    userID,
		Email:     email,
		CreatedAt: time.Now(),
		ExpiresAt: time.Now().Add(duration),
	}
	
	key := fmt.Sprintf("%s:%s", s.prefix, sessionID)
	if err := s.store.Set(ctx, key, sessionData, duration); err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	
	return sessionID, nil
}

func (s *SessionManager) GetSession(ctx context.Context, sessionID string) (*SessionData, error) {
	var sessionData SessionData
	
	key := fmt.Sprintf("%s:%s", s.prefix, sessionID)
	if err := s.store.Get(ctx, key, &sessionData); err != nil {
		return nil, fmt.Errorf("session not found: %w", err)
	}
	
	if time.Now().After(sessionData.ExpiresAt) {
		s.store.Delete(ctx, key)
		return nil, fmt.Errorf("session expired")
	}
	
	return &sessionData, nil
}

func (s *SessionManager) InvalidateSession(ctx context.Context, sessionID string) error {
	key := fmt.Sprintf("%s:%s", s.prefix, sessionID)
	return s.store.Delete(ctx, key)
}

func (s *SessionManager) InvalidateUserSessions(ctx context.Context, userID string) error {
	// This would require maintaining a user->sessions index
	// For now, individual session invalidation is supported
	log.Printf("Bulk session invalidation for user %s not implemented", userID)
	return nil
}

func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	randomBytes := make([]byte, length)
	
	if _, err := rand.Read(randomBytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	
	for i := range result {
		result[i] = charset[int(randomBytes[i])%len(charset)]
	}
	return string(result)
}