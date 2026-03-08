// Package repository provides database access for Learning Desktop models.
package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/birddigital/learning-desktop/internal/models"
	"github.com/jmoiron/sqlx"
	"github.com/google/uuid"
)

var (
	ErrNotFound      = errors.New("record not found")
	ErrDuplicate     = errors.New("duplicate record")
	ErrUnauthorized  = errors.New("unauthorized access")
)

// DB wraps sqlx.DB with tenant context
type DB struct {
	*sqlx.DB
}

// NewDB creates a new database connection
func NewDB(driver, dsn string) (*DB, error) {
	db, err := sqlx.Connect(driver, dsn)
	if err != nil {
		return nil, err
	}
	return &DB{DB: db}, nil
}

// SetTenantContext sets the tenant_id for RLS policies
func (db *DB) SetTenantContext(ctx context.Context, tenantID uuid.UUID) error {
	_, err := db.ExecContext(ctx, "SELECT set_tenant_context($1)", tenantID)
	return err
}

// SetStudentContext sets the student_id for RLS policies
func (db *DB) SetStudentContext(ctx context.Context, studentID uuid.UUID) error {
	_, err := db.ExecContext(ctx, "SELECT set_student_context($1)", studentID)
	return err
}

// ============================================================================
// TENANT REPOSITORY
// ============================================================================

// TenantRepository handles tenant operations
type TenantRepository struct {
	db *DB
}

func NewTenantRepository(db *DB) *TenantRepository {
	return &TenantRepository{db: db}
}

// Create inserts a new tenant
func (r *TenantRepository) Create(ctx context.Context, tenant *models.Tenant) error {
	query := `
		INSERT INTO tenants (id, name, slug, plan, max_students, max_ai_calls,
			billing_email, billing_period, custom_domain, logo_url, theme_config)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		tenant.ID, tenant.Name, tenant.Slug, tenant.Plan, tenant.MaxStudents,
		tenant.MaxAICalls, tenant.BillingEmail, tenant.BillingPeriod,
		tenant.CustomDomain, tenant.LogoURL, tenant.ThemeConfig,
	).Scan(&tenant.CreatedAt, &tenant.UpdatedAt)

	return err
}

// GetByID retrieves a tenant by ID
func (r *TenantRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tenant, error) {
	var tenant models.Tenant
	query := `
		SELECT id, name, slug, plan, max_students, max_ai_calls,
			billing_email, billing_period, custom_domain, logo_url, theme_config,
			created_at, updated_at
		FROM tenants WHERE id = $1`

	err := r.db.GetContext(ctx, &tenant, query, id)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// GetBySlug retrieves a tenant by slug
func (r *TenantRepository) GetBySlug(ctx context.Context, slug string) (*models.Tenant, error) {
	var tenant models.Tenant
	query := `
		SELECT id, name, slug, plan, max_students, max_ai_calls,
			billing_email, billing_period, custom_domain, logo_url, theme_config,
			created_at, updated_at
		FROM tenants WHERE slug = $1`

	err := r.db.GetContext(ctx, &tenant, query, slug)
	if err != nil {
		return nil, err
	}
	return &tenant, nil
}

// ============================================================================
// STUDENT REPOSITORY
// ============================================================================

// StudentRepository handles student operations
type StudentRepository struct {
	db *DB
}

func NewStudentRepository(db *DB) *StudentRepository {
	return &StudentRepository{db: db}
}

// Create inserts a new student
func (r *StudentRepository) Create(ctx context.Context, student *models.Student) error {
	query := `
		INSERT INTO students (id, tenant_id, external_id, name, email, avatar_url,
			skill_level, interests, goals, background_text, industry, role, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING created_at, updated_at`

	err := r.db.QueryRowContext(ctx, query,
		student.ID, student.TenantID, student.ExternalID, student.Name, student.Email,
		student.AvatarURL, student.SkillLevel, student.Interests, student.Goals,
		student.BackgroundText, student.Industry, student.Role, student.Status,
	).Scan(&student.CreatedAt, &student.UpdatedAt)

	return err
}

// GetByID retrieves a student by ID (requires student context set)
func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
	var student models.Student
	query := `
		SELECT id, tenant_id, external_id, name, email, avatar_url,
			skill_level, interests, goals, background_text, industry, role, status,
			last_login_at, total_minutes, created_at, updated_at
		FROM students WHERE id = $1`

	err := r.db.GetContext(ctx, &student, query, id)
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// GetByTenantEmail retrieves a student by tenant and email
func (r *StudentRepository) GetByTenantEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.Student, error) {
	var student models.Student
	query := `
		SELECT id, tenant_id, external_id, name, email, avatar_url,
			skill_level, interests, goals, background_text, industry, role, status,
			last_login_at, total_minutes, created_at, updated_at
		FROM students WHERE tenant_id = $1 AND email = $2`

	err := r.db.GetContext(ctx, &student, query, tenantID, email)
	if err != nil {
		return nil, err
	}
	return &student, nil
}

// UpdateLastLogin updates the last_login_at timestamp
func (r *StudentRepository) UpdateLastLogin(ctx context.Context, studentID uuid.UUID) error {
	query := `UPDATE students SET last_login_at = NOW() WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, studentID)
	return err
}

// AddTime adds learning time to a student's total
func (r *StudentRepository) AddTime(ctx context.Context, studentID uuid.UUID, minutes int) error {
	query := `UPDATE students SET total_minutes = total_minutes + $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, minutes, studentID)
	return err
}

// ============================================================================
// STUDENT SESSION REPOSITORY
// ============================================================================

// SessionRepository handles session operations
type SessionRepository struct {
	db *DB
}

func NewSessionRepository(db *DB) *SessionRepository {
	return &SessionRepository{db: db}
}

// Create inserts a new session
func (r *SessionRepository) Create(ctx context.Context, session *models.StudentSession) error {
	query := `
		INSERT INTO student_sessions (id, student_id, tenant_id, current_module,
			current_lesson, completed_ids, status)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING started_at, last_active`

	err := r.db.QueryRowContext(ctx, query,
		session.ID, session.StudentID, session.TenantID, session.CurrentModule,
		session.CurrentLesson, session.CompletedIDs, session.Status,
	).Scan(&session.StartedAt, &session.LastActive)

	return err
}

// GetActive retrieves the active session for a student
func (r *SessionRepository) GetActive(ctx context.Context, studentID uuid.UUID) (*models.StudentSession, error) {
	var session models.StudentSession
	query := `
		SELECT id, student_id, tenant_id, current_module, current_lesson,
			completed_ids, status, started_at, last_active, completed_at
		FROM student_sessions
		WHERE student_id = $1 AND status = 'active'
		ORDER BY last_active DESC
		LIMIT 1`

	err := r.db.GetContext(ctx, &session, query, studentID)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

// UpdateProgress updates session progress
func (r *SessionRepository) UpdateProgress(ctx context.Context, sessionID uuid.UUID, moduleID, lessonID *int) error {
	query := `
		UPDATE student_sessions
		SET current_module = $2, current_lesson = $3, last_active = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID, moduleID, lessonID)
	return err
}

// Complete marks a session as completed
func (r *SessionRepository) Complete(ctx context.Context, sessionID uuid.UUID) error {
	query := `
		UPDATE student_sessions
		SET status = 'completed', completed_at = NOW()
		WHERE id = $1`

	_, err := r.db.ExecContext(ctx, query, sessionID)
	return err
}

// ============================================================================
// COURSE REPOSITORY
// ============================================================================

// CourseRepository handles course operations
type CourseRepository struct {
	db *DB
}

func NewCourseRepository(db *DB) *CourseRepository {
	return &CourseRepository{db: db}
}

// GetPublished retrieves all published courses
func (r *CourseRepository) GetPublished(ctx context.Context) ([]models.Course, error) {
	var courses []models.Course
	query := `
		SELECT id, slug, title, description, thumbnail_url, is_published,
			required_plan, difficulty, duration_weeks, total_minutes, created_at, updated_at
		FROM courses WHERE is_published = true
		ORDER BY title`

	err := r.db.SelectContext(ctx, &courses, query)
	return courses, err
}

// GetBySlug retrieves a course by slug
func (r *CourseRepository) GetBySlug(ctx context.Context, slug string) (*models.Course, error) {
	var course models.Course
	query := `
		SELECT id, slug, title, description, thumbnail_url, is_published,
			required_plan, difficulty, duration_weeks, total_minutes, created_at, updated_at
		FROM courses WHERE slug = $1`

	err := r.db.GetContext(ctx, &course, query, slug)
	if err != nil {
		return nil, err
	}
	return &course, nil
}

// GetWithModules retrieves a course with its modules
func (r *CourseRepository) GetWithModules(ctx context.Context, slug string) (*models.Course, error) {
	course, err := r.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	// Load modules
	modules, err := r.getModulesForCourse(ctx, course.ID)
	if err != nil {
		return nil, err
	}
	course.Modules = modules

	return course, nil
}

func (r *CourseRepository) getModulesForCourse(ctx context.Context, courseID uuid.UUID) ([]models.CourseModule, error) {
	var modules []models.CourseModule
	query := `
		SELECT id, course_id, order_index, title, description, duration_weeks, created_at
		FROM course_modules WHERE course_id = $1
		ORDER BY order_index`

	err := r.db.SelectContext(ctx, &modules, query, courseID)
	return modules, err
}

// GetLessonByID retrieves a lesson by ID
func (r *CourseRepository) GetLessonByID(ctx context.Context, lessonID uuid.UUID) (*models.CourseLesson, error) {
	var lesson models.CourseLesson
	query := `
		SELECT id, module_id, order_index, title, description, duration_minutes,
			video_id, text_content, interactive_type, interactive_data,
			checkpoint_type, checkpoint_data, passing_score, difficulty, concepts, created_at
		FROM course_lessons WHERE id = $1`

	err := r.db.GetContext(ctx, &lesson, query, lessonID)
	if err != nil {
		return nil, err
	}
	return &lesson, nil
}

// ============================================================================
// PROGRESS REPOSITORY
// ============================================================================

// ProgressRepository handles progress tracking operations
type ProgressRepository struct {
	db *DB
}

func NewProgressRepository(db *DB) *ProgressRepository {
	return &ProgressRepository{db: db}
}

// RecordEvent records a progress event
func (r *ProgressRepository) RecordEvent(ctx context.Context, event *models.ProgressEvent) error {
	query := `
		INSERT INTO progress_events (id, student_id, tenant_id, session_id,
			event_type, lesson_id, module_id, checkpoint_score, data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING timestamp, version`

	err := r.db.QueryRowContext(ctx, query,
		event.ID, event.StudentID, event.TenantID, event.SessionID,
		event.EventType, event.LessonID, event.ModuleID, event.CheckpointScore, event.Data,
	).Scan(&event.Timestamp, &event.Version)

	if err != nil {
		return err
	}

	// Trigger materialized view refresh asynchronously
	go r.refreshStudentProgress(event.StudentID)

	return nil
}

// GetStudentProgress retrieves progress for a student in a course
func (r *ProgressRepository) GetStudentProgress(ctx context.Context, studentID, courseID uuid.UUID) (*models.StudentProgress, error) {
	var progress models.StudentProgress
	query := `
		SELECT student_id, tenant_id, course_id, total_lessons, completed_lessons,
			current_lesson_index, max_checkpoint_index, total_checkpoint_score,
			concepts_learned, started_at, last_activity_at
		FROM student_progress WHERE student_id = $1 AND course_id = $2`

	err := r.db.GetContext(ctx, &progress, query, studentID, courseID)
	if err != nil {
		return nil, err
	}
	return &progress, nil
}

// refreshStudentProgress refreshes the materialized view (non-blocking)
func (r *ProgressRepository) refreshStudentProgress(studentID uuid.UUID) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, _ = r.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY student_progress")
}

// ============================================================================
// CHAT REPOSITORY
// ============================================================================

// ChatRepository handles chat message operations
type ChatRepository struct {
	db *DB
}

func NewChatRepository(db *DB) *ChatRepository {
	return &ChatRepository{db: db}
}

// CreateMessage inserts a new chat message
func (r *ChatRepository) CreateMessage(ctx context.Context, msg *models.ChatMessage) error {
	query := `
		INSERT INTO chat_messages (id, session_id, student_id, tenant_id,
			role, content, media_type, media_source, media_data, lesson_id, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING created_at`

	err := r.db.QueryRowContext(ctx, query,
		msg.ID, msg.SessionID, msg.StudentID, msg.TenantID,
		msg.Role, msg.Content, msg.MediaType, msg.MediaSource, msg.MediaData,
		msg.LessonID, msg.Metadata,
	).Scan(&msg.CreatedAt)

	return err
}

// GetMessagesBySession retrieves messages for a session
func (r *ChatRepository) GetMessagesBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]models.ChatMessage, error) {
	var messages []models.ChatMessage
	query := `
		SELECT id, session_id, student_id, tenant_id, role, content,
			media_type, media_source, media_data, lesson_id, metadata, created_at
		FROM chat_messages
		WHERE session_id = $1
		ORDER BY created_at ASC
		LIMIT $2`

	err := r.db.SelectContext(ctx, &messages, query, sessionID, limit)
	return messages, err
}

// ============================================================================
// AI USAGE REPOSITORY
// ============================================================================

// AIUsageRepository handles AI usage tracking operations
type AIUsageRepository struct {
	db *DB
}

func NewAIUsageRepository(db *DB) *AIUsageRepository {
	return &AIUsageRepository{db: db}
}

// RecordUsage records an AI API usage event
func (r *AIUsageRepository) RecordUsage(ctx context.Context, usage *models.AIUsageEvent) error {
	query := `
		INSERT INTO ai_usage_events (id, tenant_id, student_id, model,
			input_tokens, output_tokens, request_type, session_id, lesson_id)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING created_at, total_tokens`

	return r.db.QueryRowContext(ctx, query,
		usage.ID, usage.TenantID, usage.StudentID, usage.Model,
		usage.InputTokens, usage.OutputTokens, usage.RequestType,
		usage.SessionID, usage.LessonID,
	).Scan(&usage.CreatedAt, &usage.TotalTokens)
}

// GetMonthlyUsage retrieves monthly usage for a tenant
func (r *AIUsageRepository) GetMonthlyUsage(ctx context.Context, tenantID uuid.UUID, year int, month time.Month) ([]models.MonthlyAIUsage, error) {
	var usage []models.MonthlyAIUsage
	query := `
		SELECT tenant_id, month, model, request_count,
			total_input_tokens, total_output_tokens, total_tokens
		FROM monthly_ai_usage
		WHERE tenant_id = $1
			AND EXTRACT(YEAR FROM month) = $2
			AND EXTRACT(MONTH FROM month) = $3
		ORDER BY month DESC, model`

	err := r.db.SelectContext(ctx, &usage, query, tenantID, year, month)
	return usage, err
}

// ============================================================================
// CHECKPOINT REPOSITORY
// ============================================================================

// CheckpointRepository handles checkpoint submission operations
type CheckpointRepository struct {
	db *DB
}

func NewCheckpointRepository(db *DB) *CheckpointRepository {
	return &CheckpointRepository{db: db}
}

// CreateSubmission records a checkpoint submission
func (r *CheckpointRepository) CreateSubmission(ctx context.Context, submission *models.CheckpointSubmission) error {
	query := `
		INSERT INTO checkpoint_submissions (id, student_id, tenant_id, lesson_id,
			checkpoint_type, answers, score, passed, feedback, feedback_data)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING submitted_at`

	err := r.db.QueryRowContext(ctx, query,
		submission.ID, submission.StudentID, submission.TenantID, submission.LessonID,
		submission.CheckpointType, submission.Answers, submission.Score,
		submission.Passed, submission.Feedback, submission.FeedbackData,
	).Scan(&submission.SubmittedAt)

	return err
}

// GetBestScores retrieves best scores for a student
func (r *CheckpointRepository) GetBestScores(ctx context.Context, studentID uuid.UUID) ([]models.LessonBestScore, error) {
	var scores []models.LessonBestScore
	query := `
		SELECT student_id, tenant_id, lesson_id, best_score, attempts, last_attempt
		FROM lesson_best_scores WHERE student_id = $1
		ORDER BY last_attempt DESC`

	err := r.db.SelectContext(ctx, &scores, query, studentID)
	return scores, err
}

// ============================================================================
// CERTIFICATE REPOSITORY
// ============================================================================

// CertificateRepository handles certificate operations
type CertificateRepository struct {
	db *DB
}

func NewCertificateRepository(db *DB) *CertificateRepository {
	return &CertificateRepository{db: db}
}

// Create creates a new certificate
func (r *CertificateRepository) Create(ctx context.Context, cert *models.Certificate) error {
	query := `
		INSERT INTO certificates (id, student_id, tenant_id, course_id,
			certificate_number, verification_hash, completed_at, total_minutes, final_score)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		RETURNING completed_at`

	return r.db.QueryRowContext(ctx, query,
		cert.ID, cert.StudentID, cert.TenantID, cert.CourseID,
		cert.CertificateNumber, cert.VerificationHash,
		cert.CompletedAt, cert.TotalMinutes, cert.FinalScore,
	).Scan(&cert.CompletedAt)
}

// GetByVerificationHash retrieves a certificate for verification
func (r *CertificateRepository) GetByVerificationHash(ctx context.Context, hash string) (*models.Certificate, error) {
	var cert models.Certificate
	query := `
		SELECT id, student_id, tenant_id, course_id, certificate_number,
			verification_hash, completed_at, total_minutes, final_score,
			pdf_url, badge_url, revoked, revoked_at, revoked_reason
		FROM certificates WHERE verification_hash = $1`

	err := r.db.GetContext(ctx, &cert, query, hash)
	if err != nil {
		return nil, err
	}
	return &cert, nil
}

// GetByStudent retrieves certificates for a student
func (r *CertificateRepository) GetByStudent(ctx context.Context, studentID uuid.UUID) ([]models.Certificate, error) {
	var certs []models.Certificate
	query := `
		SELECT id, student_id, tenant_id, course_id, certificate_number,
			verification_hash, completed_at, total_minutes, final_score,
			pdf_url, badge_url, revoked, revoked_at, revoked_reason
		FROM certificates WHERE student_id = $1 AND NOT revoked
		ORDER BY completed_at DESC`

	err := r.db.SelectContext(ctx, &certs, query, studentID)
	return certs, err
}
