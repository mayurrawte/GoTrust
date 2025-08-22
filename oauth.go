package gotrust

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type OAuthManager struct {
	config        *Config
	sessionStore  SessionStore
	statePrefix   string
}

func NewOAuthManager(config *Config, sessionStore SessionStore) *OAuthManager {
	return &OAuthManager{
		config:       config,
		sessionStore: sessionStore,
		statePrefix:  "oauth:state",
	}
}

// GetAuthURL generates the OAuth authorization URL
func (o *OAuthManager) GetAuthURL(provider OAuthProvider, redirectURI string) (string, error) {
	state := generateRandomString(32)
	
	// Store state with redirect URI
	stateData := &OAuthState{
		State:       state,
		RedirectURI: redirectURI,
		ExpiresAt:   time.Now().Add(o.config.OAuthStateExpiration),
	}
	
	ctx := context.Background()
	stateKey := fmt.Sprintf("%s:%s", o.statePrefix, state)
	if err := o.sessionStore.Set(ctx, stateKey, stateData, o.config.OAuthStateExpiration); err != nil {
		return "", fmt.Errorf("failed to store oauth state: %w", err)
	}
	
	switch provider {
	case ProviderGoogle:
		return o.getGoogleAuthURL(state)
	case ProviderGitHub:
		return o.getGitHubAuthURL(state)
	default:
		return "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (o *OAuthManager) getGoogleAuthURL(state string) (string, error) {
	if o.config.GoogleClientID == "" {
		return "", fmt.Errorf("Google OAuth not configured")
	}
	
	params := url.Values{}
	params.Add("client_id", o.config.GoogleClientID)
	params.Add("redirect_uri", o.config.GoogleRedirectURI)
	params.Add("scope", strings.Join(o.config.GoogleScopes, " "))
	params.Add("response_type", "code")
	params.Add("state", state)
	params.Add("access_type", "offline")
	
	return "https://accounts.google.com/o/oauth2/auth?" + params.Encode(), nil
}

func (o *OAuthManager) getGitHubAuthURL(state string) (string, error) {
	if o.config.GitHubClientID == "" {
		return "", fmt.Errorf("GitHub OAuth not configured")
	}
	
	params := url.Values{}
	params.Add("client_id", o.config.GitHubClientID)
	params.Add("redirect_uri", o.config.GitHubRedirectURI)
	params.Add("scope", strings.Join(o.config.GitHubScopes, " "))
	params.Add("state", state)
	
	return "https://github.com/login/oauth/authorize?" + params.Encode(), nil
}

// ValidateCallback validates OAuth callback and returns user info
func (o *OAuthManager) ValidateCallback(provider OAuthProvider, state, code string) (*OAuthUserInfo, string, error) {
	// Validate state
	redirectURI, err := o.validateState(state)
	if err != nil {
		return nil, "", fmt.Errorf("invalid state: %w", err)
	}
	
	// Exchange code for token and get user info
	switch provider {
	case ProviderGoogle:
		userInfo, err := o.handleGoogleCallback(code)
		return userInfo, redirectURI, err
	case ProviderGitHub:
		userInfo, err := o.handleGitHubCallback(code)
		return userInfo, redirectURI, err
	default:
		return nil, "", fmt.Errorf("unsupported provider: %s", provider)
	}
}

func (o *OAuthManager) validateState(state string) (string, error) {
	ctx := context.Background()
	stateKey := fmt.Sprintf("%s:%s", o.statePrefix, state)
	
	var stateData OAuthState
	if err := o.sessionStore.Get(ctx, stateKey, &stateData); err != nil {
		return "", fmt.Errorf("state not found or expired")
	}
	
	// Delete used state
	o.sessionStore.Delete(ctx, stateKey)
	
	if time.Now().After(stateData.ExpiresAt) {
		return "", fmt.Errorf("state expired")
	}
	
	return stateData.RedirectURI, nil
}

func (o *OAuthManager) handleGoogleCallback(code string) (*OAuthUserInfo, error) {
	// Exchange code for token
	tokenURL := "https://oauth2.googleapis.com/token"
	data := url.Values{}
	data.Set("client_id", o.config.GoogleClientID)
	data.Set("client_secret", o.config.GoogleClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", o.config.GoogleRedirectURI)
	
	resp, err := http.Post(tokenURL, "application/x-www-form-urlencoded", strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}
	
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	
	// Get user info
	userInfoURL := "https://www.googleapis.com/oauth2/v2/userinfo"
	req, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	
	client := &http.Client{}
	userResp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userResp.Body.Close()
	
	if userResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", userResp.StatusCode)
	}
	
	var googleUser struct {
		ID      string `json:"id"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}
	
	if err := json.NewDecoder(userResp.Body).Decode(&googleUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}
	
	return &OAuthUserInfo{
		ID:        googleUser.ID,
		Email:     googleUser.Email,
		Name:      googleUser.Name,
		AvatarURL: googleUser.Picture,
		Provider:  string(ProviderGoogle),
	}, nil
}

func (o *OAuthManager) handleGitHubCallback(code string) (*OAuthUserInfo, error) {
	// Exchange code for token
	tokenURL := "https://github.com/login/oauth/access_token"
	data := url.Values{}
	data.Set("client_id", o.config.GitHubClientID)
	data.Set("client_secret", o.config.GitHubClientSecret)
	data.Set("code", code)
	
	req, err := http.NewRequest("POST", tokenURL, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to exchange code: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status: %d", resp.StatusCode)
	}
	
	var tokenResp struct {
		AccessToken string `json:"access_token"`
	}
	
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}
	
	// Get user info
	userInfoURL := "https://api.github.com/user"
	userReq, err := http.NewRequest("GET", userInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	
	userReq.Header.Set("Authorization", "Bearer "+tokenResp.AccessToken)
	userReq.Header.Set("Accept", "application/vnd.github.v3+json")
	
	userResp, err := client.Do(userReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get user info: %w", err)
	}
	defer userResp.Body.Close()
	
	if userResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed with status: %d", userResp.StatusCode)
	}
	
	var githubUser struct {
		ID        int64  `json:"id"`
		Login     string `json:"login"`
		Email     string `json:"email"`
		Name      string `json:"name"`
		AvatarURL string `json:"avatar_url"`
	}
	
	if err := json.NewDecoder(userResp.Body).Decode(&githubUser); err != nil {
		return nil, fmt.Errorf("failed to parse user info: %w", err)
	}
	
	// Get email if not public
	if githubUser.Email == "" {
		email, err := o.getGitHubEmail(tokenResp.AccessToken)
		if err == nil {
			githubUser.Email = email
		}
	}
	
	displayName := githubUser.Name
	if displayName == "" {
		displayName = githubUser.Login
	}
	
	return &OAuthUserInfo{
		ID:        fmt.Sprintf("%d", githubUser.ID),
		Email:     githubUser.Email,
		Name:      displayName,
		AvatarURL: githubUser.AvatarURL,
		Provider:  string(ProviderGitHub),
	}, nil
}

func (o *OAuthManager) getGitHubEmail(accessToken string) (string, error) {
	emailURL := "https://api.github.com/user/emails"
	
	req, err := http.NewRequest("GET", emailURL, nil)
	if err != nil {
		return "", err
	}
	
	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/vnd.github.v3+json")
	
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("email request failed")
	}
	
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	
	var emails []struct {
		Email    string `json:"email"`
		Primary  bool   `json:"primary"`
		Verified bool   `json:"verified"`
	}
	
	if err := json.Unmarshal(body, &emails); err != nil {
		return "", err
	}
	
	// Find primary verified email
	for _, email := range emails {
		if email.Primary && email.Verified {
			return email.Email, nil
		}
	}
	
	// Fallback to first verified email
	for _, email := range emails {
		if email.Verified {
			return email.Email, nil
		}
	}
	
	return "", fmt.Errorf("no verified email found")
}