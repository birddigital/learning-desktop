// Package auth provides authentication and authorization middleware for Learning Desktop.
package auth

import (
	"context"
	"net/http"
	"strings"

	"github.com/birddigital/learning-desktop/internal/models"
	"github.com/google/uuid"
)

// Context keys for tenant and student IDs.
type contextKey string

const (
	// TenantIDKey is the context key for tenant ID.
	TenantIDKey contextKey = "tenant_id"
	// StudentIDKey is the context key for student ID.
	StudentIDKey contextKey = "student_id"
	// UserIDKey is the context key for user ID (authenticated user).
	UserIDKey contextKey = "user_id"
)

// Config holds authentication configuration.
type Config struct {
	// JWTSecret is the secret key for JWT signing.
	JWTSecret string
	// JWTIssuer is the issuer for JWT tokens.
	JWTIssuer string
	// RequireAuth determines if authentication is required.
	// If false, requests without auth headers proceed with default tenant/student.
	RequireAuth bool
	// DefaultTenantID is used when auth is not required and no tenant is provided.
	DefaultTenantID uuid.UUID
}

// Middleware creates an authentication middleware.
type Middleware struct {
	config *Config
}

// NewMiddleware creates a new authentication middleware.
func NewMiddleware(cfg *Config) *Middleware {
	if cfg == nil {
		cfg = &Config{}
	}
	return &Middleware{config: cfg}
}

// TenantContext extracts or sets tenant context from the request.
// It checks multiple sources in order:
// 1. JWT token (if present)
// 2. X-Tenant-ID header
// 3. Query parameter
// 4. Default tenant (if configured)
func (m *Middleware) TenantContext(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		// Try JWT first
		if authHeader := r.Header.Get("Authorization"); authHeader != "" {
			if strings.HasPrefix(authHeader, "Bearer ") {
				token := strings.TrimPrefix(authHeader, "Bearer ")
				if claims, err := m.parseJWT(token); err == nil {
					// Valid token - set tenant and student from claims
					ctx = m.setContextValues(ctx, claims)
				}
			}
		}

		// If no tenant from JWT, try header
		if tenantID := ctx.Value(TenantIDKey); tenantID == nil {
			if tenantHeader := r.Header.Get("X-Tenant-ID"); tenantHeader != "" {
				if id, err := uuid.Parse(tenantHeader); err == nil {
					ctx = context.WithValue(ctx, TenantIDKey, id)
				}
			}
		}

		// If still no tenant, try query parameter (for demo/testing)
		if tenantID := ctx.Value(TenantIDKey); tenantID == nil {
			if tenantQuery := r.URL.Query().Get("tenant_id"); tenantQuery != "" {
				if id, err := uuid.Parse(tenantQuery); err == nil {
					ctx = context.WithValue(ctx, TenantIDKey, id)
				}
			}
		}

		// Set default tenant if configured and no tenant found
		if tenantID := ctx.Value(TenantIDKey); tenantID == nil {
			if m.config.DefaultTenantID != uuid.Nil {
				ctx = context.WithValue(ctx, TenantIDKey, m.config.DefaultTenantID)
			}
		}

		// Set default student if none found (for demo purposes)
		if studentID := ctx.Value(StudentIDKey); studentID == nil {
			// In production, this would require authentication
			// For now, we'll let handlers create student IDs from cookies
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireAuth validates that the request has valid authentication.
// Returns 401 if authentication is required but not present.
func (m *Middleware) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			if m.config.RequireAuth {
				http.Error(w, "Authorization required", http.StatusUnauthorized)
				return
			}
			// Auth not required - proceed with default context
			next.ServeHTTP(w, r)
			return
		}

		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Invalid authorization header format", http.StatusUnauthorized)
			return
		}

		token := strings.TrimPrefix(authHeader, "Bearer ")
		claims, err := m.parseJWT(token)
		if err != nil {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Set context from valid claims
		ctx := m.setContextValues(r.Context(), claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// OptionalAuth attempts to authenticate but doesn't require it.
// Useful for endpoints that work for both authenticated and anonymous users.
func (m *Middleware) OptionalAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()

		authHeader := r.Header.Get("Authorization")
		if authHeader != "" && strings.HasPrefix(authHeader, "Bearer ") {
			token := strings.TrimPrefix(authHeader, "Bearer ")
			if claims, err := m.parseJWT(token); err == nil {
				ctx = m.setContextValues(ctx, claims)
			}
		}

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

// RequireRole checks that the authenticated user has the required role.
func (m *Middleware) RequireRole(role models.UserRole) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			userRole := r.Context().Value("user_role")
			if userRole == nil {
				http.Error(w, "Authentication required", http.StatusUnauthorized)
				return
			}

			if userRole.(models.UserRole) != role && userRole.(models.UserRole) != models.RoleAdmin {
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// JWTClaims represents JWT token claims.
type JWTClaims struct {
	TenantID  string `json:"tenant_id"`
	StudentID string `json:"student_id,omitempty"`
	UserID    string `json:"user_id,omitempty"`
	Role      string `json:"role,omitempty"`
	Issuer    string `json:"iss,omitempty"`
}

// parseJWT parses and validates a JWT token.
// This is a simplified implementation. In production, use a proper JWT library.
func (m *Middleware) parseJWT(token string) (*JWTClaims, error) {
	// TODO: Implement proper JWT validation using golang-jwt/jwt
	// For now, this is a placeholder that validates token format

	if token == "" {
		return nil, ErrInvalidToken
	}

	// Placeholder claims - in production, decode and validate JWT
	claims := &JWTClaims{
		Issuer: m.config.JWTIssuer,
	}

	return claims, nil
}

// setContextValues sets tenant and student IDs from claims into context.
func (m *Middleware) setContextValues(ctx context.Context, claims *JWTClaims) context.Context {
	if claims.TenantID != "" {
		if id, err := uuid.Parse(claims.TenantID); err == nil {
			ctx = context.WithValue(ctx, TenantIDKey, id)
		}
	}

	if claims.StudentID != "" {
		if id, err := uuid.Parse(claims.StudentID); err == nil {
			ctx = context.WithValue(ctx, StudentIDKey, id)
		}
	}

	if claims.UserID != "" {
		if id, err := uuid.Parse(claims.UserID); err == nil {
			ctx = context.WithValue(ctx, UserIDKey, id)
		}
	}

	if claims.Role != "" {
		ctx = context.WithValue(ctx, "user_role", models.UserRole(claims.Role))
	}

	return ctx
}

// Helper functions for extracting context values.

// GetTenantID retrieves the tenant ID from context.
func GetTenantID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(TenantIDKey).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// GetStudentID retrieves the student ID from context.
func GetStudentID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(StudentIDKey).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// GetUserID retrieves the user ID from context.
func GetUserID(ctx context.Context) (uuid.UUID, bool) {
	id, ok := ctx.Value(UserIDKey).(uuid.UUID)
	return id, ok && id != uuid.Nil
}

// MustGetTenantID retrieves the tenant ID or panics.
// Use only when you're certain the context has a tenant.
func MustGetTenantID(ctx context.Context) uuid.UUID {
	id, ok := GetTenantID(ctx)
	if !ok {
		panic("tenant ID not found in context")
	}
	return id
}

// MustGetStudentID retrieves the student ID or panics.
func MustGetStudentID(ctx context.Context) uuid.UUID {
	id, ok := GetStudentID(ctx)
	if !ok {
		panic("student ID not found in context")
	}
	return id
}

// Errors.
var (
	// ErrInvalidToken is returned when the token is invalid.
	ErrInvalidToken = &AuthError{Code: "invalid_token", Message: "Invalid authentication token"}
	// ErrExpiredToken is returned when the token has expired.
	ErrExpiredToken = &AuthError{Code: "expired_token", Message: "Authentication token has expired"}
	// ErrMissingTenant is returned when no tenant is found.
	ErrMissingTenant = &AuthError{Code: "missing_tenant", Message: "Tenant context not found"}
)

// AuthError represents an authentication error.
type AuthError struct {
	Code    string
	Message string
}

func (e *AuthError) Error() string {
	return e.Message
}
