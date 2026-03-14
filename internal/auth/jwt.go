// Package auth provides JWT token management for Learning Desktop.
package auth

import (
	"fmt"
	"time"

	"github.com/birddigital/learning-desktop/internal/models"
	"github.com/google/uuid"
)

// TokenManager handles JWT token creation and validation.
type TokenManager struct {
	secret   string
	issuer   string
	expiresIn time.Duration
}

// NewTokenManager creates a new JWT token manager.
func NewTokenManager(secret, issuer string, expiresIn time.Duration) *TokenManager {
	if expiresIn == 0 {
		expiresIn = 24 * time.Hour // Default 24 hours
	}
	return &TokenManager{
		secret:    secret,
		issuer:    issuer,
		expiresIn: expiresIn,
	}
}

// Token represents a JWT token with metadata.
type Token struct {
	AccessToken string
	TokenType   string
	ExpiresIn   int64
	ExpiresAt   time.Time
}

// CreateTokenRequest contains information for creating a token.
type CreateTokenRequest struct {
	TenantID  uuid.UUID
	StudentID uuid.UUID
	UserID    uuid.UUID
	Role      models.UserRole
	Metadata  map[string]string
}

// CreateToken creates a new JWT token for the given subject.
// TODO: Implement using golang-jwt/jwt/v5 library
func (tm *TokenManager) CreateToken(req *CreateTokenRequest) (*Token, error) {
	if req.TenantID == uuid.Nil {
		return nil, fmt.Errorf("tenant_id is required")
	}

	// For now, return a placeholder token
	// In production, this would create a proper JWT signed with the secret
	expiresAt := time.Now().Add(tm.expiresIn)

	token := &Token{
		AccessToken: tm.createPlaceholderToken(req, expiresAt),
		TokenType:   "Bearer",
		ExpiresIn:   int64(tm.expiresIn.Seconds()),
		ExpiresAt:   expiresAt,
	}

	return token, nil
}

// CreateRefreshToken creates a refresh token for long-lived sessions.
func (tm *TokenManager) CreateRefreshToken(userID uuid.UUID) (string, error) {
	// Refresh tokens live longer (30 days)
	expiresIn := 30 * 24 * time.Hour
	expiresAt := time.Now().Add(expiresIn)

	// TODO: Implement proper refresh token with storage
	return fmt.Sprintf("refresh_%s_%d", userID.String(), expiresAt.Unix()), nil
}

// ValidateToken validates a JWT token and returns the claims.
func (tm *TokenManager) ValidateToken(tokenString string) (*JWTClaims, error) {
	if tokenString == "" {
		return nil, ErrInvalidToken
	}

	// TODO: Implement proper JWT validation
	// For now, return placeholder claims
	claims := &JWTClaims{
		Issuer: tm.issuer,
	}

	return claims, nil
}

// RefreshToken validates a refresh token and issues a new access token.
func (tm *TokenManager) RefreshToken(refreshToken string) (*Token, error) {
	// TODO: Implement proper refresh token validation and new token issuance
	return nil, fmt.Errorf("refresh token not yet implemented")
}

// createPlaceholderToken creates a placeholder token for development.
// TODO: Replace with proper JWT implementation using golang-jwt/jwt/v5
func (tm *TokenManager) createPlaceholderToken(req *CreateTokenRequest, expiresAt time.Time) string {
	// This is a placeholder that mimics JWT structure
	// In production, use github.com/golang-jwt/jwt/v5 to create real JWTs

	claims := &JWTClaims{
		TenantID:  req.TenantID.String(),
		StudentID: req.StudentID.String(),
		UserID:    req.UserID.String(),
		Role:      string(req.Role),
		Issuer:    tm.issuer,
	}

	// Format: "placeholder.jwt.token.{tenant_id}.{expires_at}"
	return fmt.Sprintf("placeholder.jwt.token.%s.%d", claims.TenantID, expiresAt.Unix())
}

// ParseTokenWithoutValidation parses a token without validating the signature.
// Useful for debugging and inspection.
func (tm *TokenManager) ParseTokenWithoutValidation(tokenString string) (*JWTClaims, error) {
	// TODO: Implement token parsing without validation
	return &JWTClaims{}, nil
}

// GetExpirationTime returns the expiration time for a token.
func (tm *TokenManager) GetExpirationTime() time.Time {
	return time.Now().Add(tm.expiresIn)
}

// IsExpired checks if a token has expired.
func (tm *TokenManager) IsExpired(expiresAt time.Time) bool {
	return time.Now().After(expiresAt)
}

// TokenMetadata contains metadata about a token.
type TokenMetadata struct {
	TenantID   uuid.UUID
	StudentID  uuid.UUID
	UserID     uuid.UUID
	Role       models.UserRole
	IssuedAt   time.Time
	ExpiresAt  time.Time
	Issuer     string
	Metadata   map[string]string
}

// ExtractMetadata extracts metadata from a token.
func (tm *TokenManager) ExtractMetadata(tokenString string) (*TokenMetadata, error) {
	claims, err := tm.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	metadata := &TokenMetadata{
		Issuer:   claims.Issuer,
		Metadata: make(map[string]string),
	}

	if claims.TenantID != "" {
		metadata.TenantID, _ = uuid.Parse(claims.TenantID)
	}
	if claims.StudentID != "" {
		metadata.StudentID, _ = uuid.Parse(claims.StudentID)
	}
	if claims.UserID != "" {
		metadata.UserID, _ = uuid.Parse(claims.UserID)
	}
	if claims.Role != "" {
		metadata.Role = models.UserRole(claims.Role)
	}

	return metadata, nil
}

// ClaimsFromRequest creates CreateTokenRequest from TokenMetadata.
func ClaimsFromRequest(metadata *TokenMetadata) *CreateTokenRequest {
	return &CreateTokenRequest{
		TenantID:  metadata.TenantID,
		StudentID: metadata.StudentID,
		UserID:    metadata.UserID,
		Role:      metadata.Role,
		Metadata:  metadata.Metadata,
	}
}
