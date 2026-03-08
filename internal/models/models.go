// Package models provides data models for the Learning Desktop platform.
// Multi-tenant PostgreSQL with Row Level Security (RLS).
package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// TENANT
// ============================================================================

// Tenant represents an organization (company, school, etc)
type Tenant struct {
	ID           uuid.UUID          `db:"id" json:"id"`
	Name         string             `db:"name" json:"name"`
	Slug         string             `db:"slug" json:"slug"`
	Plan         TenantPlan         `db:"plan" json:"plan"`
	MaxStudents  int                `db:"max_students" json:"max_students"`
	MaxAICalls   int                `db:"max_ai_calls" json:"max_ai_calls"`
	BillingEmail string             `db:"billing_email" json:"billing_email,omitempty"`
	BillingPeriod BillingPeriod     `db:"billing_period" json:"billing_period"`
	CustomDomain string             `db:"custom_domain" json:"custom_domain,omitempty"`
	LogoURL      string             `db:"logo_url" json:"logo_url,omitempty"`
	ThemeConfig  json.RawMessage    `db:"theme_config" json:"theme_config,omitempty"`
	CreatedAt    time.Time          `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time          `db:"updated_at" json:"updated_at"`
}

// TenantPlan represents subscription tier
type TenantPlan string

const (
	PlanFree        TenantPlan = "free"
	PlanPro         TenantPlan = "pro"
	PlanEnterprise  TenantPlan = "enterprise"
)

// BillingPeriod represents billing cycle
type BillingPeriod string

const (
	BillingMonthly  BillingPeriod = "monthly"
	BillingAnnual   BillingPeriod = "annual"
)

// ============================================================================
// STUDENT
// ============================================================================

// Student represents a student user
type Student struct {
	ID             uuid.UUID   `db:"id" json:"id"`
	TenantID       uuid.UUID   `db:"tenant_id" json:"tenant_id"`
	ExternalID     string      `db:"external_id" json:"external_id,omitempty"`
	Name           string      `db:"name" json:"name"`
	Email          string      `db:"email" json:"email"`
	AvatarURL      string      `db:"avatar_url" json:"avatar_url,omitempty"`
	SkillLevel     SkillLevel  `db:"skill_level" json:"skill_level"`
	Interests      []string    `db:"interests" json:"interests"`
	Goals          []string    `db:"goals" json:"goals"`
	BackgroundText string      `db:"background_text" json:"background_text,omitempty"`
	Industry       string      `db:"industry" json:"industry,omitempty"`
	Role           string      `db:"role" json:"role,omitempty"`
	Status         StudentStatus `db:"status" json:"status"`
	LastLoginAt    *time.Time  `db:"last_login_at" json:"last_login_at,omitempty"`
	TotalMinutes   int         `db:"total_minutes" json:"total_minutes"`
	CreatedAt      time.Time   `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time   `db:"updated_at" json:"updated_at"`
}

// SkillLevel represents student's self-assessed skill level
type SkillLevel string

const (
	SkillBeginner      SkillLevel = "beginner"
	SkillIntermediate  SkillLevel = "intermediate"
	SkillAdvanced      SkillLevel = "advanced"
)

// StudentStatus represents account status
type StudentStatus string

const (
	StatusActive     StudentStatus = "active"
	StatusInactive   StudentStatus = "inactive"
	StatusSuspended  StudentStatus = "suspended"
)

// ============================================================================
// STUDENT SESSION
// ============================================================================

// StudentSession represents an active learning session
type StudentSession struct {
	ID             uuid.UUID      `db:"id" json:"id"`
	StudentID      uuid.UUID      `db:"student_id" json:"student_id"`
	TenantID       uuid.UUID      `db:"tenant_id" json:"tenant_id"`
	CurrentModule  *int           `db:"current_module" json:"current_module,omitempty"`
	CurrentLesson  *int           `db:"current_lesson" json:"current_lesson,omitempty"`
	CompletedIDs   []int          `db:"completed_ids" json:"completed_ids"`
	Status         SessionStatus  `db:"status" json:"status"`
	StartedAt      time.Time      `db:"started_at" json:"started_at"`
	LastActive     time.Time      `db:"last_active" json:"last_active"`
	CompletedAt    *time.Time     `db:"completed_at" json:"completed_at,omitempty"`
}

// SessionStatus represents session state
type SessionStatus string

const (
	SessionActive      SessionStatus = "active"
	SessionCompleted   SessionStatus = "completed"
	SessionAbandoned   SessionStatus = "abandoned"
)

// ============================================================================
// COURSE CONTENT
// ============================================================================

// Course represents a course definition
type Course struct {
	ID             uuid.UUID     `db:"id" json:"id"`
	Slug           string        `db:"slug" json:"slug"`
	Title          string        `db:"title" json:"title"`
	Description    string        `db:"description" json:"description,omitempty"`
	ThumbnailURL   string        `db:"thumbnail_url" json:"thumbnail_url,omitempty"`
	IsPublished    bool          `db:"is_published" json:"is_published"`
	RequiredPlan   TenantPlan    `db:"required_plan" json:"required_plan"`
	Difficulty     SkillLevel    `db:"difficulty" json:"difficulty"`
	DurationWeeks  int           `db:"duration_weeks" json:"duration_weeks"`
	TotalMinutes   int           `db:"total_minutes" json:"total_minutes"`
	CreatedAt      time.Time     `db:"created_at" json:"created_at"`
	UpdatedAt      time.Time     `db:"updated_at" json:"updated_at"`
	Modules        []CourseModule `db:"-" json:"modules,omitempty"` // Loaded separately
}

// CourseModule represents a section within a course
type CourseModule struct {
	ID             uuid.UUID  `db:"id" json:"id"`
	CourseID       uuid.UUID  `db:"course_id" json:"course_id"`
	OrderIndex     int        `db:"order_index" json:"order_index"`
	Title          string     `db:"title" json:"title"`
	Description    string     `db:"description" json:"description,omitempty"`
	DurationWeeks  int        `db:"duration_weeks" json:"duration_weeks"`
	CreatedAt      time.Time  `db:"created_at" json:"created_at"`
	Lessons        []CourseLesson `db:"-" json:"lessons,omitempty"` // Loaded separately
}

// CourseLesson represents a single lesson
type CourseLesson struct {
	ID                uuid.UUID       `db:"id" json:"id"`
	ModuleID          uuid.UUID       `db:"module_id" json:"module_id"`
	OrderIndex        int             `db:"order_index" json:"order_index"`
	Title             string          `db:"title" json:"title"`
	Description       string          `db:"description" json:"description,omitempty"`
	DurationMinutes   int             `db:"duration_minutes" json:"duration_minutes"`
	VideoID           string          `db:"video_id" json:"video_id,omitempty"`
	TextContent       json.RawMessage `db:"text_content" json:"text_content,omitempty"`
	InteractiveType   InteractiveType `db:"interactive_type" json:"interactive_type,omitempty"`
	InteractiveData   json.RawMessage `db:"interactive_data" json:"interactive_data,omitempty"`
	CheckpointType    CheckpointType  `db:"checkpoint_type" json:"checkpoint_type,omitempty"`
	CheckpointData    json.RawMessage `db:"checkpoint_data" json:"checkpoint_data,omitempty"`
	PassingScore      float64         `db:"passing_score" json:"passing_score"`
	Difficulty        SkillLevel      `db:"difficulty" json:"difficulty"`
	Concepts          []string        `db:"concepts" json:"concepts"`
	CreatedAt         time.Time       `db:"created_at" json:"created_at"`
}

// InteractiveType represents the type of interactive content
type InteractiveType string

const (
	Interactive3D    InteractiveType = "3d"
	InteractiveBoard InteractiveType = "board"
	InteractiveQuiz  InteractiveType = "quiz"
	InteractiveCode  InteractiveType = "code"
)

// CheckpointType represents the type of checkpoint assessment
type CheckpointType string

const (
	CheckpointQuiz     CheckpointType = "quiz"
	CheckpointExercise CheckpointType = "exercise"
	CheckpointProject  CheckpointType = "project"
)

// ============================================================================
// PROGRESS TRACKING
// ============================================================================

// ProgressEvent represents a learning progress event (event sourcing)
type ProgressEvent struct {
	ID               uuid.UUID       `db:"id" json:"id"`
	StudentID        uuid.UUID       `db:"student_id" json:"student_id"`
	TenantID         uuid.UUID       `db:"tenant_id" json:"tenant_id"`
	SessionID        *uuid.UUID      `db:"session_id" json:"session_id,omitempty"`
	EventType        ProgressEventType `db:"event_type" json:"event_type"`
	LessonID         *uuid.UUID      `db:"lesson_id" json:"lesson_id,omitempty"`
	ModuleID         *uuid.UUID      `db:"module_id" json:"module_id,omitempty"`
	CheckpointScore  *float64        `db:"checkpoint_score" json:"checkpoint_score,omitempty"`
	Data             json.RawMessage `db:"data" json:"data,omitempty"`
	Timestamp        time.Time       `db:"timestamp" json:"timestamp"`
	Version          int             `db:"version" json:"version"`
}

// ProgressEventType represents types of progress events
type ProgressEventType string

const (
	EventLessonStarted    ProgressEventType = "lesson_started"
	EventLessonCompleted  ProgressEventType = "lesson_completed"
	EventCheckpointPassed ProgressEventType = "checkpoint_passed"
	EventConceptLearned   ProgressEventType = "concept_learned"
)

// StudentProgress represents computed student progress (from materialized view)
type StudentProgress struct {
	StudentID           uuid.UUID  `db:"student_id" json:"student_id"`
	TenantID            uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	CourseID            uuid.UUID  `db:"course_id" json:"course_id"`
	TotalLessons        int        `db:"total_lessons" json:"total_lessons"`
	CompletedLessons    int        `db:"completed_lessons" json:"completed_lessons"`
	CurrentLessonIndex  int        `db:"current_lesson_index" json:"current_lesson_index"`
	MaxCheckpointIndex  int        `db:"max_checkpoint_index" json:"max_checkpoint_index"`
	TotalCheckpointScore float64   `db:"total_checkpoint_score" json:"total_checkpoint_score"`
	ConceptsLearned     int        `db:"concepts_learned" json:"concepts_learned"`
	StartedAt           *time.Time `db:"started_at" json:"started_at,omitempty"`
	LastActivityAt      *time.Time `db:"last_activity_at" json:"last_activity_at,omitempty"`
}

// CompletionPercent returns the percentage of course completed
func (p *StudentProgress) CompletionPercent() float64 {
	if p.TotalLessons == 0 {
		return 0
	}
	return float64(p.CompletedLessons) / float64(p.TotalLessons) * 100
}

// ============================================================================
// CHAT MESSAGES
// ============================================================================

// ChatMessage represents an AI tutor conversation message
type ChatMessage struct {
	ID           uuid.UUID       `db:"id" json:"id"`
	SessionID    uuid.UUID       `db:"session_id" json:"session_id"`
	StudentID    uuid.UUID       `db:"student_id" json:"student_id"`
	TenantID     uuid.UUID       `db:"tenant_id" json:"tenant_id"`
	Role         MessageRole     `db:"role" json:"role"`
	Content      string          `db:"content" json:"content"`
	MediaType    *InteractiveType `db:"media_type" json:"media_type,omitempty"`
	MediaSource  string          `db:"media_source" json:"media_source,omitempty"`
	MediaData    json.RawMessage `db:"media_data" json:"media_data,omitempty"`
	LessonID     *uuid.UUID      `db:"lesson_id" json:"lesson_id,omitempty"`
	Metadata     json.RawMessage `db:"metadata" json:"metadata,omitempty"`
	CreatedAt    time.Time       `db:"created_at" json:"created_at"`
}

// MessageRole represents the sender of a chat message
type MessageRole string

const (
	RoleUser      MessageRole = "user"
	RoleAssistant MessageRole = "assistant"
	RoleSystem    MessageRole = "system"
)

// ============================================================================
// AI USAGE TRACKING
// ============================================================================

// AIUsageEvent represents AI API usage for billing/limits
type AIUsageEvent struct {
	ID            uuid.UUID    `db:"id" json:"id"`
	TenantID      uuid.UUID    `db:"tenant_id" json:"tenant_id"`
	StudentID     *uuid.UUID   `db:"student_id" json:"student_id,omitempty"`
	Model         string       `db:"model" json:"model"` // e.g., "claude-3-5-sonnet"
	InputTokens   int          `db:"input_tokens" json:"input_tokens"`
	OutputTokens  int          `db:"output_tokens" json:"output_tokens"`
	TotalTokens   int          `db:"total_tokens" json:"total_tokens"`
	RequestType   string       `db:"request_type" json:"request_type"` // "chat" | "completion" | "embedding"
	SessionID     *uuid.UUID   `db:"session_id" json:"session_id,omitempty"`
	LessonID      *uuid.UUID   `db:"lesson_id" json:"lesson_id,omitempty"`
	CreatedAt     time.Time    `db:"created_at" json:"created_at"`
}

// MonthlyAIUsage represents aggregated monthly usage (from materialized view)
type MonthlyAIUsage struct {
	TenantID           uuid.UUID `db:"tenant_id" json:"tenant_id"`
	Month              time.Time `db:"month" json:"month"`
	Model              string    `db:"model" json:"model"`
	RequestCount       int       `db:"request_count" json:"request_count"`
	TotalInputTokens   int       `db:"total_input_tokens" json:"total_input_tokens"`
	TotalOutputTokens  int       `db:"total_output_tokens" json:"total_output_tokens"`
	TotalTokens        int       `db:"total_tokens" json:"total_tokens"`
}

// ============================================================================
// CHECKPOINT SUBMISSIONS
// ============================================================================

// CheckpointSubmission represents a student checkpoint/exercise submission
type CheckpointSubmission struct {
	ID              uuid.UUID       `db:"id" json:"id"`
	StudentID       uuid.UUID       `db:"student_id" json:"student_id"`
	TenantID        uuid.UUID       `db:"tenant_id" json:"tenant_id"`
	LessonID        uuid.UUID       `db:"lesson_id" json:"lesson_id"`
	CheckpointType  CheckpointType  `db:"checkpoint_type" json:"checkpoint_type"`
	Answers         json.RawMessage `db:"answers" json:"answers"`
	Score           float64         `db:"score" json:"score"`
	Passed          bool            `db:"passed" json:"passed"`
	Feedback        string          `db:"feedback" json:"feedback,omitempty"`
	FeedbackData    json.RawMessage `db:"feedback_data" json:"feedback_data,omitempty"`
	SubmittedAt     time.Time       `db:"submitted_at" json:"submitted_at"`
	GradedAt        *time.Time      `db:"graded_at" json:"graded_at,omitempty"`
}

// LessonBestScore represents best score per lesson (from materialized view)
type LessonBestScore struct {
	StudentID     uuid.UUID  `db:"student_id" json:"student_id"`
	TenantID      uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	LessonID      uuid.UUID  `db:"lesson_id" json:"lesson_id"`
	BestScore     float64    `db:"best_score" json:"best_score"`
	Attempts      int        `db:"attempts" json:"attempts"`
	LastAttempt   time.Time  `db:"last_attempt" json:"last_attempt"`
}

// ============================================================================
// CERTIFICATES
// ============================================================================

// Certificate represents a course completion certificate
type Certificate struct {
	ID                 uuid.UUID  `db:"id" json:"id"`
	StudentID          uuid.UUID  `db:"student_id" json:"student_id"`
	TenantID           uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	CourseID           uuid.UUID  `db:"course_id" json:"course_id"`
	CertificateNumber  string     `db:"certificate_number" json:"certificate_number"`
	VerificationHash   string     `db:"verification_hash" json:"verification_hash"`
	CompletedAt        time.Time  `db:"completed_at" json:"completed_at"`
	TotalMinutes       int        `db:"total_minutes" json:"total_minutes"`
	FinalScore         float64    `db:"final_score" json:"final_score"`
	PDFURL             string     `db:"pdf_url" json:"pdf_url,omitempty"`
	BadgeURL           string     `db:"badge_url" json:"badge_url,omitempty"`
	Revoked            bool       `db:"revoked" json:"revoked"`
	RevokedAt          *time.Time `db:"revoked_at" json:"revoked_at,omitempty"`
	RevokedReason      string     `db:"revoked_reason" json:"revoked_reason,omitempty"`
}

// ============================================================================
// SPEECH ANALYSIS (from voice service)
// ============================================================================

// StudentSpeechAnalysis contains insights from analyzing student speech
type StudentSpeechAnalysis struct {
	SessionID          uuid.UUID    `json:"session_id"`
	Timestamp          time.Time    `json:"timestamp"`
	Transcript         string       `json:"transcript"`
	Topics             []string     `json:"topics"`
	Sentiment          string       `json:"sentiment"` // "confident" | "uncertain" | "confused" | "engaged"
	Fillers            []FillerWord `json:"fillers"`
	Pace               SpeechPace   `json:"pace"`
	ConceptsMentioned  []string     `json:"concepts_mentioned"`
	QuestionsAsked     int          `json:"questions_asked"`
	UnderstandingScore float64      `json:"understanding_score"` // 0-1
}

// FillerWord represents a filler word detection
type FillerWord struct {
	Word     string        `json:"word"`
	Position time.Duration `json:"position"`
	Type     string        `json:"type"`
}

// SpeechPace describes speaking speed
type SpeechPace struct {
	WordsPerMinute int    `json:"words_per_minute"`
	Label          string `json:"label"` // "slow" | "normal" | "fast"
}

// ============================================================================
// GOALS & COMMITMENTS
// ============================================================================

// Goal represents a student's declared objective with a deadline
type Goal struct {
	ID          uuid.UUID  `db:"id" json:"id"`
	StudentID   uuid.UUID  `db:"student_id" json:"student_id"`
	TenantID    uuid.UUID  `db:"tenant_id" json:"tenant_id"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	TargetDate  time.Time  `db:"target_date" json:"target_date"`
	
	// AI planning output
	CourseID    *uuid.UUID `db:"course_id" json:"course_id,omitempty"`
	Confidence  float64    `db:"confidence" json:"confidence"` // AI's confidence in achievability (0-1)
	
	// Status tracking
	Status      GoalStatus `db:"status" json:"status"`
	Progress    float64    `db:"progress" json:"progress"` // 0-100
	
	// Timestamps
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	CompletedAt *time.Time `db:"completed_at" json:"completed_at,omitempty"`
	AbandonedAt *time.Time `db:"abandoned_at" json:"abandoned_at,omitempty"`
	
	// Relations
	Milestones  []Milestone `db:"-" json:"milestones,omitempty"`
	Insights    []InsightSnapshot `db:"-" json:"insights,omitempty"`
}

// GoalStatus represents the state of a goal
type GoalStatus string

const (
	GoalStatusActive     GoalStatus = "active"
	GoalStatusOnTrack    GoalStatus = "on_track"
	GoalStatusAtRisk     GoalStatus = "at_risk"
	GoalStatusBehind     GoalStatus = "behind"
	GoalStatusCompleted  GoalStatus = "completed"
	GoalStatusAbandoned  GoalStatus = "abandoned"
)

// Milestone represents a checkpoint on the path to a goal
type Milestone struct {
	ID             uuid.UUID       `db:"id" json:"id"`
	GoalID         uuid.UUID       `db:"goal_id" json:"goal_id"`
	OrderIndex     int             `db:"order_index" json:"order_index"`
	Title          string          `db:"title" json:"title"`
	Description    string          `db:"description" json:"description"`
	TargetDate     time.Time       `db:"target_date" json:"target_date"`
	Dependencies   []uuid.UUID     `db:"dependencies" json:"dependencies"`
	
	// Completion tracking
	LessonID       *uuid.UUID      `db:"lesson_id" json:"lesson_id,omitempty"`
	CompletedAt    *time.Time      `db:"completed_at" json:"completed_at,omitempty"`
	
	// Status
	Status         MilestoneStatus `db:"status" json:"status"`
	
	CreatedAt      time.Time       `db:"created_at" json:"created_at"`
}

// MilestoneStatus represents milestone completion state
type MilestoneStatus string

const (
	MilestoneStatusPending    MilestoneStatus = "pending"
	MilestoneStatusInProgress MilestoneStatus = "in_progress"
	MilestoneStatusCompleted  MilestoneStatus = "completed"
	MilestoneStatusSkipped    MilestoneStatus = "skipped"
)

// InsightSnapshot stores historical insight data for a goal
type InsightSnapshot struct {
	ID              uuid.UUID  `db:"id" json:"id"`
	GoalID          uuid.UUID  `db:"goal_id" json:"goal_id"`
	RecordedAt      time.Time `db:"recorded_at" json:"recorded_at"`
	
	// Snapshot data
	OnTrack         bool      `db:"on_track" json:"on_track"`
	ProgressPercent float64   `db:"progress_percent" json:"progress_percent"`
	ProjectedDate   *time.Time `db:"projected_date" json:"projected_date,omitempty"`
	MissedBy        *int       `db:"missed_by_days" json:"missed_by_days,omitempty"`
	
	// Metrics
	Velocity        float64   `db:"velocity" json:"velocity"` // progress per day
	RequiredVelocity float64   `db:"required_velocity" json:"required_velocity"`
	
	// Recommendation at time of snapshot
	Recommendation string    `db:"recommendation" json:"recommendation"`
	Urgency         string    `db:"urgency" json:"urgency"`
}

// ============================================================================
// SCHEDULING & COMMUNICATIONS
// ============================================================================

// ScheduleBlock represents a planned time slot for learning
type ScheduleBlock struct {
	ID          uuid.UUID     `db:"id" json:"id"`
	StudentID   uuid.UUID     `db:"student_id" json:"student_id"`
	GoalID      *uuid.UUID    `db:"goal_id" json:"goal_id,omitempty"`
	
	// When
	ScheduledAt time.Time     `db:"scheduled_at" json:"scheduled_at"`
	Duration    int           `db:"duration_minutes" json:"duration_minutes"`
	
	// What
	Title       string        `db:"title" json:"title"`
	Type        BlockType     `db:"type" json:"type"`
	LessonID    *uuid.UUID    `db:"lesson_id" json:"lesson_id,omitempty"`
	
	// Status
	Status      BlockStatus   `db:"status" json:"status"`
	CompletedAt *time.Time    `db:"completed_at" json:"completed_at,omitempty"`
	
	// Reminder settings
	Reminder    bool          `db:"reminder" json:"reminder"`
	RemindedAt  *time.Time    `db:"reminded_at" json:"reminded_at,omitempty"`
	
	CreatedAt   time.Time     `db:"created_at" json:"created_at"`
}

// BlockType represents the type of scheduled block
type BlockType string

const (
	BlockTypeLesson    BlockType = "lesson"
	BlockTypeQuiz      BlockType = "quiz"
	BlockTypeReview    BlockType = "review"
	BlockTypeProject   BlockType = "project"
	BlockTypeOther     BlockType = "other"
)

// BlockStatus represents whether a scheduled block was completed
type BlockStatus string

const (
	BlockStatusScheduled  BlockStatus = "scheduled"
	BlockStatusCompleted  BlockStatus = "completed"
	BlockStatusMissed     BlockStatus = "missed"
	BlockStatusSkipped    BlockStatus = "skipped"
)

// Communication represents an outbound message to a student
type Communication struct {
	ID           uuid.UUID      `db:"id" json:"id"`
	StudentID    uuid.UUID      `db:"student_id" json:"student_id"`
	GoalID       *uuid.UUID     `db:"goal_id" json:"goal_id,omitempty"`
	
	Type         CommType       `db:"type" json:"type"`
	Channel      CommChannel    `db:"channel" json:"channel"`
	Subject      string         `db:"subject" json:"subject"`
	Body         string         `db:"body" json:"body"`
	
	ScheduledAt  time.Time      `db:"scheduled_at" json:"scheduled_at"`
	SentAt       *time.Time     `db:"sent_at" json:"sent_at,omitempty"`
	DeliveredAt  *time.Time     `db:"delivered_at" json:"delivered_at,omitempty"`
	
	// Tracking
	Status       CommStatus     `db:"status" json:"status"`
	OpenedAt     *time.Time     `db:"opened_at" json:"opened_at,omitempty"`
	ClickedAt    *time.Time     `db:"clicked_at" json:"clicked_at,omitempty"`
	
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// CommType represents the type of communication
type CommType string

const (
	CommTypeReminder      CommType = "reminder"
	CommTypeEscalation    CommType = "escalation"
	CommTypeUpdate        CommType = "update"
	CommTypeCelebration   CommType = "celebration"
	CommTypeNudge         CommType = "nudge"
)

// CommChannel represents where the message is sent
type CommChannel string

const (
	CommChannelInApp    CommChannel = "in_app"
	CommChannelEmail    CommChannel = "email"
	CommChannelSMS      CommChannel = "sms"
	CommChannelPush     CommChannel = "push"
)

// CommStatus represents delivery status
type CommStatus string

const (
	CommStatusPending    CommStatus = "pending"
	CommStatusSent       CommStatus = "sent"
	CommStatusDelivered  CommStatus = "delivered"
	CommStatusOpened     CommStatus = "opened"
	CommStatusClicked    CommStatus = "clicked"
	CommStatusFailed     CommStatus = "failed"
)
