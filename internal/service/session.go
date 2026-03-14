// Package service provides business logic for Learning Desktop.
package service

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/birddigital/learning-desktop/internal/models"
	"github.com/birddigital/learning-desktop/internal/repository"
	"github.com/google/uuid"
)

// SessionService manages student chat sessions with persistent storage.
// Replaces the in-memory sessionStore that blocked horizontal scaling.
type SessionService struct {
	chatRepo    *repository.ChatRepository
	sessionRepo *repository.SessionRepository
	defaultTenantID uuid.UUID

	// Cache for active sessions (reduces database queries)
	cache      map[uuid.UUID]*models.StudentSession
	cacheMutex sync.RWMutex
	cacheTTL   time.Duration
}

// NewSessionService creates a new session service.
func NewSessionService(
	chatRepo *repository.ChatRepository,
	sessionRepo *repository.SessionRepository,
	defaultTenantID uuid.UUID,
) *SessionService {
	return &SessionService{
		chatRepo:       chatRepo,
		sessionRepo:    sessionRepo,
		defaultTenantID: defaultTenantID,
		cache:          make(map[uuid.UUID]*models.StudentSession),
		cacheTTL:       5 * time.Minute,
	}
}

// GetOrCreateSession retrieves an active session or creates a new one.
// Implements cache-aside pattern: check cache → check database → create if needed.
func (s *SessionService) GetOrCreateSession(ctx context.Context, studentID uuid.UUID) (*models.StudentSession, error) {
	// Check cache first
	s.cacheMutex.RLock()
	if session, exists := s.cache[studentID]; exists {
		// Check if session is still valid (not expired)
		if s.isSessionValid(session) {
			s.cacheMutex.RUnlock()
			return session, nil
		}
	}
	s.cacheMutex.RUnlock()

	// Cache miss - check database for active session
	session, err := s.sessionRepo.GetActiveByStudent(ctx, studentID)
	// "no rows in result set" is expected when there's no active session
	if err != nil && err != context.DeadlineExceeded && err != context.Canceled {
		// Only return error if it's not a "not found" error
		// sql.ErrNoRows would be the proper check, but we'll check the error string
		if err.Error() != "sql: no rows in result set" {
			return nil, fmt.Errorf("get active session: %w", err)
		}
	}

	// If no active session or expired, create new one
	if session == nil || !s.isSessionValid(session) {
		session = &models.StudentSession{
			ID:        uuid.New(),
			StudentID: studentID,
			TenantID:  s.defaultTenantID,
			Status:    models.SessionActive,
			StartedAt: time.Now(),
			LastActive: time.Now(),
		}
		if err := s.sessionRepo.Create(ctx, session); err != nil {
			return nil, fmt.Errorf("create session: %w", err)
		}
	}

	// Update cache
	s.cacheMutex.Lock()
	s.cache[studentID] = session
	s.cacheMutex.Unlock()

	return session, nil
}

// GetSessionByID retrieves a session by ID.
func (s *SessionService) GetSessionByID(ctx context.Context, sessionID uuid.UUID) (*models.StudentSession, error) {
	session, err := s.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}
	return session, nil
}

// GetRecentSessions returns recent sessions for a student.
func (s *SessionService) GetRecentSessions(ctx context.Context, studentID uuid.UUID, limit int) ([]*models.StudentSession, error) {
	sessions, err := s.sessionRepo.GetRecentByStudent(ctx, studentID, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent sessions: %w", err)
	}
	return sessions, nil
}

// AddMessage adds a message to a session and persists it.
func (s *SessionService) AddMessage(ctx context.Context, sessionID, studentID, tenantID uuid.UUID, role models.MessageRole, content string) (*models.ChatMessage, error) {
	// Validate role
	if role != models.RoleUser && role != models.RoleAssistant && role != models.RoleSystem {
		return nil, fmt.Errorf("invalid role: %s", role)
	}

	message := &models.ChatMessage{
		ID:        uuid.New(),
		SessionID: sessionID,
		StudentID: studentID,
		TenantID:  tenantID,
		Role:      role,
		Content:   content,
	}

	if err := s.chatRepo.CreateMessage(ctx, message); err != nil {
		return nil, fmt.Errorf("create message: %w", err)
	}

	// Update session last active time
	if err := s.sessionRepo.UpdateLastActive(ctx, sessionID); err != nil {
		// Log but don't fail - message was already saved
		fmt.Printf("Warning: failed to update session last active: %v\n", err)
	}

	return message, nil
}

// GetMessages retrieves messages for a session.
func (s *SessionService) GetMessages(ctx context.Context, sessionID uuid.UUID, limit int) ([]*models.ChatMessage, error) {
	messages, err := s.chatRepo.GetMessagesBySession(ctx, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}

	// Convert slice of values to slice of pointers
	result := make([]*models.ChatMessage, len(messages))
	for i := range messages {
		result[i] = &messages[i]
	}

	return result, nil
}

// EndSession marks a session as ended.
func (s *SessionService) EndSession(ctx context.Context, sessionID uuid.UUID) error {
	if err := s.sessionRepo.EndSession(ctx, sessionID); err != nil {
		return fmt.Errorf("end session: %w", err)
	}

	// Remove from cache
	s.cacheMutex.Lock()
	// Find and remove the session (need to iterate since cache is keyed by studentID)
	for studentID, session := range s.cache {
		if session.ID == sessionID {
			delete(s.cache, studentID)
			break
		}
	}
	s.cacheMutex.Unlock()

	return nil
}

// CleanupExpiredSessions removes expired sessions from cache.
// Should be called periodically (e.g., every minute).
func (s *SessionService) CleanupExpiredSessions() {
	s.cacheMutex.Lock()
	defer s.cacheMutex.Unlock()

	now := time.Now()
	for studentID, session := range s.cache {
		if now.Sub(session.LastActive) > s.cacheTTL {
			delete(s.cache, studentID)
		}
	}
}

// isSessionValid checks if a session is still valid (not too old).
func (s *SessionService) isSessionValid(session *models.StudentSession) bool {
	// Check if session was explicitly ended or completed
	if session.Status == models.SessionEnded || session.Status == models.SessionCompleted {
		return false
	}

	// Sessions older than 24 hours are considered expired
	if time.Since(session.LastActive) > 24*time.Hour {
		return false
	}

	return true
}
