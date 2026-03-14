// Package auth provides OAuth provider integration for Learning Desktop.
package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"
)

// Provider represents an OAuth provider (Google, GitHub, etc.).
type Provider string

const (
	// ProviderGoogle is Google OAuth provider.
	ProviderGoogle Provider = "google"
	// ProviderGitHub is GitHub OAuth provider.
	ProviderGitHub Provider = "github"
	// ProviderMicrosoft is Microsoft OAuth provider.
	ProviderMicrosoft Provider = "microsoft"
	// ProviderApple is Apple Sign In provider.
	ProviderApple Provider = "apple"
)

// OAuthConfig holds OAuth provider configuration.
type OAuthConfig struct {
	Provider        Provider
	ClientID        string
	ClientSecret    string
	RedirectURL     string
	Scopes          []string
	AuthURL         string
	TokenURL        string
	UserInfoURL     string
}

// UserInfo represents user information from OAuth provider.
type UserInfo struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Picture       string `json:"picture"`
	EmailVerified bool   `json:"email_verified"`
}

// Manager handles OAuth provider operations.
type Manager struct {
	configs map[Provider]*OAuthConfig
}

// NewManager creates a new OAuth manager.
func NewManager() *Manager {
	return &Manager{
		configs: make(map[Provider]*OAuthConfig),
	}
}

// RegisterProvider registers an OAuth provider configuration.
func (m *Manager) RegisterProvider(cfg *OAuthConfig) {
	m.configs[cfg.Provider] = cfg
}

// GetAuthURL generates the authorization URL for OAuth flow.
func (m *Manager) GetAuthURL(provider Provider, state string) (string, error) {
	cfg, ok := m.configs[provider]
	if !ok {
		return "", fmt.Errorf("provider %s not configured", provider)
	}

	// Build authorization URL with query parameters
	authURL, _ := url.Parse(cfg.AuthURL)
	params := authURL.Query()
	params.Set("client_id", cfg.ClientID)
	params.Set("redirect_uri", cfg.RedirectURL)
	params.Set("response_type", "code")
	params.Set("state", state)
	params.Set("scope", joinScopes(cfg.Scopes))
	authURL.RawQuery = params.Encode()

	return authURL.String(), nil
}

// ExchangeCode exchanges the authorization code for an access token.
func (m *Manager) ExchangeCode(ctx context.Context, provider Provider, code string) (*TokenResponse, error) {
	cfg, ok := m.configs[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	// Build token request
	data := url.Values{}
	data.Set("client_id", cfg.ClientID)
	data.Set("client_secret", cfg.ClientSecret)
	data.Set("code", code)
	data.Set("grant_type", "authorization_code")
	data.Set("redirect_uri", cfg.RedirectURL)

	req, err := http.NewRequestWithContext(ctx, "POST", cfg.TokenURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.URL.RawQuery = data.Encode()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("exchange code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed: %s", resp.Status)
	}

	var tokenResp TokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &tokenResp, nil
}

// GetUserInfo fetches user information using the access token.
func (m *Manager) GetUserInfo(ctx context.Context, provider Provider, accessToken string) (*UserInfo, error) {
	cfg, ok := m.configs[provider]
	if !ok {
		return nil, fmt.Errorf("provider %s not configured", provider)
	}

	req, err := http.NewRequestWithContext(ctx, "GET", cfg.UserInfoURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetch user info: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user info request failed: %s", resp.Status)
	}

	var userInfo UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return &userInfo, nil
}

// TokenResponse represents OAuth token response.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int    `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Scope        string `json:"scope,omitempty"`
}

// StateGenerator generates and validates OAuth state parameters.
type StateGenerator struct {
	secret string
}

// NewStateGenerator creates a new state generator.
func NewStateGenerator(secret string) *StateGenerator {
	return &StateGenerator{secret: secret}
}

// Generate creates a new state parameter for CSRF protection.
func (sg *StateGenerator) Generate() (string, error) {
	// TODO: Implement proper state generation with signature
	return fmt.Sprintf("state_%d", time.Now().UnixNano()), nil
}

// Validate validates the state parameter.
func (sg *StateGenerator) Validate(state string) bool {
	// TODO: Implement proper state validation
	return len(state) > 0
}

// joinScopes joins OAuth scopes into a space-separated string.
func joinScopes(scopes []string) string {
	result := ""
	for i, scope := range scopes {
		if i > 0 {
			result += " "
		}
		result += scope
	}
	return result
}

// DefaultProviderConfigs returns default configurations for common providers.
func DefaultProviderConfigs(redirectBaseURL string) map[Provider]*OAuthConfig {
	return map[Provider]*OAuthConfig{
		ProviderGoogle: {
			Provider:    ProviderGoogle,
			RedirectURL: redirectBaseURL + "/auth/callback/google",
			Scopes:      []string{"openid", "email", "profile"},
			AuthURL:     "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL:    "https://oauth2.googleapis.com/token",
			UserInfoURL: "https://www.googleapis.com/oauth2/v2/userinfo",
		},
		ProviderGitHub: {
			Provider:    ProviderGitHub,
			RedirectURL: redirectBaseURL + "/auth/callback/github",
			Scopes:      []string{"user:email"},
			AuthURL:     "https://github.com/login/oauth/authorize",
			TokenURL:    "https://github.com/login/oauth/access_token",
			UserInfoURL: "https://api.github.com/user",
		},
		ProviderMicrosoft: {
			Provider:    ProviderMicrosoft,
			RedirectURL: redirectBaseURL + "/auth/callback/microsoft",
			Scopes:      []string{"openid", "email", "profile"},
			AuthURL:     "https://login.microsoftonline.com/common/oauth2/v2.0/authorize",
			TokenURL:    "https://login.microsoftonline.com/common/oauth2/v2.0/token",
			UserInfoURL: "https://graph.microsoft.com/v1.0/me",
		},
	}
}

// SessionInfo represents OAuth session information.
type SessionInfo struct {
	Provider    Provider `json:"provider"`
	State       string   `json:"state"`
	RedirectURL string   `json:"redirect_url"`
	CreatedAt   time.Time `json:"created_at"`
}

// CallbackHandler handles OAuth callback requests.
type CallbackHandler struct {
	manager      *Manager
	stateGen     *StateGenerator
	onSuccess    func(w http.ResponseWriter, r *http.Request, userInfo *UserInfo)
	onError      func(w http.ResponseWriter, r *http.Request, err error)
}

// NewCallbackHandler creates a new OAuth callback handler.
func NewCallbackHandler(manager *Manager, stateGen *StateGenerator) *CallbackHandler {
	return &CallbackHandler{
		manager:  manager,
		stateGen: stateGen,
	}
}

// ServeHTTP handles the OAuth callback.
func (ch *CallbackHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Extract code and state from query parameters
	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")
	errorParam := r.URL.Query().Get("error")

	if errorParam != "" {
		ch.handleError(w, r, fmt.Errorf("oauth error: %s", errorParam))
		return
	}

	if code == "" {
		ch.handleError(w, r, fmt.Errorf("missing authorization code"))
		return
	}

	if !ch.stateGen.Validate(state) {
		ch.handleError(w, r, fmt.Errorf("invalid state parameter"))
		return
	}

	// Exchange code for token
	// This is a placeholder - in production, you'd extract provider from URL path
	provider := ProviderGoogle // Default to Google for now
	tokenResp, err := ch.manager.ExchangeCode(r.Context(), provider, code)
	if err != nil {
		ch.handleError(w, r, fmt.Errorf("exchange code: %w", err))
		return
	}

	// Get user info
	userInfo, err := ch.manager.GetUserInfo(r.Context(), provider, tokenResp.AccessToken)
	if err != nil {
		ch.handleError(w, r, fmt.Errorf("get user info: %w", err))
		return
	}

	// Call success handler
	if ch.onSuccess != nil {
		ch.onSuccess(w, r, userInfo)
	}
}

func (ch *CallbackHandler) handleError(w http.ResponseWriter, r *http.Request, err error) {
	if ch.onError != nil {
		ch.onError(w, r, err)
	} else {
		http.Error(w, err.Error(), http.StatusBadRequest)
	}
}

// SetSuccessHandler sets the success callback handler.
func (ch *CallbackHandler) SetSuccessHandler(handler func(w http.ResponseWriter, r *http.Request, userInfo *UserInfo)) {
	ch.onSuccess = handler
}

// SetErrorHandler sets the error callback handler.
func (ch *CallbackHandler) SetErrorHandler(handler func(w http.ResponseWriter, r *http.Request, err error)) {
	ch.onError = handler
}
