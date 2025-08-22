package gotrust

import (
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret    []byte
	issuer    string
	expiresIn time.Duration
}

func NewJWTManager(secret string, issuer string, expiresIn time.Duration) *JWTManager {
	return &JWTManager{
		secret:    []byte(secret),
		issuer:    issuer,
		expiresIn: expiresIn,
	}
}

func (j *JWTManager) GenerateToken(claims TokenClaims) (string, error) {
	now := time.Now()
	
	jwtClaims := jwt.MapClaims{
		"user_id":  claims.UserID,
		"email":    claims.Email,
		"name":     claims.Name,
		"provider": claims.Provider,
		"iss":      j.issuer,
		"sub":      claims.UserID,
		"iat":      now.Unix(),
		"exp":      now.Add(j.expiresIn).Unix(),
		"nbf":      now.Unix(),
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwtClaims)
	return token.SignedString(j.secret)
}

func (j *JWTManager) ValidateToken(tokenString string) (*TokenClaims, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}
	
	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}
	
	userID, _ := claims["user_id"].(string)
	email, _ := claims["email"].(string)
	name, _ := claims["name"].(string)
	provider, _ := claims["provider"].(string)
	
	if userID == "" {
		return nil, fmt.Errorf("user_id not found in token")
	}
	
	return &TokenClaims{
		UserID:   userID,
		Email:    email,
		Name:     name,
		Provider: provider,
	}, nil
}

func (j *JWTManager) GenerateRefreshToken(userID string) (string, error) {
	now := time.Now()
	
	claims := jwt.MapClaims{
		"user_id": userID,
		"type":    "refresh",
		"iss":     j.issuer,
		"sub":     userID,
		"iat":     now.Unix(),
		"exp":     now.Add(30 * 24 * time.Hour).Unix(), // 30 days
	}
	
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(j.secret)
}

func (j *JWTManager) ValidateRefreshToken(tokenString string) (string, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secret, nil
	})
	
	if err != nil {
		return "", fmt.Errorf("failed to parse refresh token: %w", err)
	}
	
	if !token.Valid {
		return "", fmt.Errorf("invalid refresh token")
	}
	
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return "", fmt.Errorf("invalid refresh token claims")
	}
	
	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return "", fmt.Errorf("not a refresh token")
	}
	
	userID, _ := claims["user_id"].(string)
	if userID == "" {
		return "", fmt.Errorf("user_id not found in refresh token")
	}
	
	return userID, nil
}