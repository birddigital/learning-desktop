// Package skill provides video game-style skill tree tracking.
// Progress is visible, measurable, and unlockable like RPG talent trees.
package skill

import (
	"time"

	"github.com/google/uuid"
)

// ============================================================================
// SKILL TREE STRUCTURE
// ============================================================================

// Tree represents a complete skill tree (e.g., "AI Fundamentals")
type Tree struct {
	ID          uuid.UUID   `db:"id" json:"id"`
	Slug        string      `db:"slug" json:"slug"`
	Title       string      `db:"title" json:"title"`
	Description string      `db:"description" json:"description"`
	Icon        string      `db:"icon" json:"icon"`          // Emoji or icon name
	Category    SkillCategory `db:"category" json:"category"`

	// Requirements
	RequiredLevel SkillLevel `db:"required_level" json:"required_level"` // Suggested starting level

	// Progress tracking
	TotalNodes   int     `db:"total_nodes" json:"total_nodes"`
	MaxPoints     int     `db:"max_points" json:"max_points"`     // Total skill points available

	// Timestamps
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`

	// Relations
	Nodes        []Node   `db:"-" json:"nodes"`
}

// SkillCategory groups related trees
type SkillCategory string

const (
	CategoryTechnical      SkillCategory = "technical"
	CategorySoftSkills     SkillCategory = "soft_skills"
	CategoryDomain         SkillCategory = "domain"        // Industry-specific
	CategoryTools          SkillCategory = "tools"
	CategoryCareer         SkillCategory = "career"
)

// Node represents a single skill in the tree
type Node struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	TreeID       uuid.UUID     `db:"tree_id" json:"tree_id"`
	ParentID     *uuid.UUID    `db:"parent_id" json:"parent_id,omitempty"` // For hierarchy

	// Identity
	Slug         string        `db:"slug" json:"slug"`
	Title        string        `db:"title" json:"title"`
	Description  string        `db:"description" json:"description"`
	Icon         string        `db:"icon" json:"icon"`          // 🎯 ⚡ 🧠 📊

	// Position in tree (for visualization)
	Position     NodePosition  `db:"position" json:"position"`

	// Requirements
	RequiredScore float64      `db:"required_score" json:"required_score"` // 0-100 to unlock
	RequiredNodes []uuid.UUID   `db:"required_nodes" json:"required_nodes"`  // Prerequisites

	// Scoring
	MaxPoints    int           `db:"max_points" json:"max_points"`       // Points available
	Weight       float64       `db:"weight" json:"weight"`               // Importance multiplier

	// Learning resources
	Lessons      []uuid.UUID   `db:"-" json:"lessons,omitempty"`     // Associated lessons
	Projects     []uuid.UUID   `db:"-" json:"projects,omitempty"`   // Proof projects

	CreatedAt    time.Time     `db:"created_at" json:"created_at"`
}

// NodePosition defines visual layout
type NodePosition struct {
	Row  int     `json:"row"`   // Vertical tier (0=top)
	Col  int     `json:"col"`   // Horizontal position
	X    float64 `json:"x"`     // Precise X (optional)
	Y    float64 `json:"y"`     // Precise Y (optional)
}

// ============================================================================
// STUDENT SKILL STATE
// ============================================================================

// StudentSkill represents a student's proficiency in a specific skill node
type StudentSkill struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	StudentID    uuid.UUID     `db:"student_id" json:"student_id"`
	NodeID       uuid.UUID     `db:"node_id" json:"node_id"`
	TreeID       uuid.UUID     `db:"tree_id" json:"tree_id"`

	// Current state
	Score        float64      `db:"score" json:"score"`           // 0-100
	Level        Proficiency  `db:"level" json:"level"`           // novice → expert
	Points       int          `db:"points" json:"points"`         // Skill points earned
	MaxPoints    int          `db:"max_points" json:"max_points"` // Points available in node

	// Status
	Unlocked     bool         `db:"unlocked" json:"unlocked"`     // Can they access this?
	UnlockedAt   *time.Time   `db:"unlocked_at" json:"unlocked_at,omitempty"`
	Completed    bool         `db:"completed" json:"completed"`    // Mastered (100%)
	CompletedAt  *time.Time   `db:"completed_at" json:"completed_at,omitempty"`

	// Evidence tracking
	LastAssessedAt  *time.Time `db:"last_assessed_at" json:"last_assessed_at,omitempty"`
	AssessmentCount int       `db:"assessment_count" json:"assessment_count"` // How many times evaluated

	// Timestamps
	CreatedAt    time.Time    `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time    `db:"updated_at" json:"updated_at"`
}

// Proficiency levels (like game ranks)
type Proficiency string

const (
	ProficiencyUnknown    Proficiency = "unknown"     // 🔒 Locked
	ProficiencyNovice     Proficiency = "novice"      // 🌱 Just started
	ProficiencyApprentice Proficiency = "apprentice"  // 📚 Learning basics
	ProficiencyAdept      Proficiency = "adept"       // 🎯 Competent
	ProficiencyExpert     Proficiency = "expert"      // ⭐ Mastery
	ProficiencyMaster     Proficiency = "master"      // 👑 Can teach others
)

// ProfessionConfig defines thresholds for each level
var ProficiencyThresholds = map[Proficiency]struct {
	MinScore      float64
	Icon          string
	Label         string
	Color         string
}{
	ProficiencyUnknown:    {0, "🔒", "Locked", "#6B7280"},
	ProficiencyNovice:     {1, "🌱", "Novice", "#10B981"},
	ProficiencyApprentice: {25, "📚", "Apprentice", "#3B82F6"},
	ProficiencyAdept:      {50, "🎯", "Adept", "#8B5CF6"},
	ProficiencyExpert:     {75, "⭐", "Expert", "#F59E0B"},
	ProficiencyMaster:     {95, "👑", "Master", "#EF4444"},
}

// ============================================================================
// SKILL ASSESSMENT
// ============================================================================

// Assessment represents a skill evaluation
type Assessment struct {
	ID           uuid.UUID     `db:"id" json:"id"`
	StudentID    uuid.UUID     `db:"student_id" json:"student_id"`
	NodeID       uuid.UUID     `db:"node_id" json:"node_id"`
	TreeID       uuid.UUID     `db:"tree_id" json:"tree_id"`

	// Assessment data
	Method       AssessmentMethod `db:"method" json:"method"` // How they were assessed
	Score        float64         `db:"score" json:"score"`     // 0-100 for this assessment
	Weight       float64         `db:"weight" json:"weight"`     // Importance (0-1)

	// Evidence
	Responses    []Response      `db:"-" json:"responses,omitempty"`     // Quiz answers
	Submissions  []Submission    `db:"-" json:"submissions,omitempty"`  // Project work
	ChatAnalysis ChatAnalysis   `db:"-" json:"chat_analysis,omitempty"` // From conversation
	VoiceAnalysis VoiceAnalysis   `db:"-" json:"voice_analysis,omitempty"` // From speech

	// AI feedback
	Feedback     string         `db:"feedback" json:"feedback"`   // What to improve
	NextSteps    []string       `db:"next_steps" json:"next_steps"` // Recommended actions

	AssessedAt   time.Time      `db:"assessed_at" json:"assessed_at"`
	CreatedAt    time.Time      `db:"created_at" json:"created_at"`
}

// AssessmentMethod types
type AssessmentMethod string

const (
	AssessMethodQuiz        AssessmentMethod = "quiz"           // Multiple choice
	AssessMethodProject     AssessmentMethod = "project"        // Practical work
	AssessMethodCode        AssessmentMethod = "code"           // Code review
	AssessMethodChat        AssessmentMethod = "chat"           // Conversation analysis
	AssessMethodVoice       AssessmentMethod = "voice"          // Speech analysis
	AssessMethodSelf        AssessmentMethod = "self"           // Self-report
	AssessMethodPeer        AssessmentMethod = "peer"           // Peer review
	AssessMethodInterview   AssessmentMethod = "interview"      // Live assessment
)

// Response represents a quiz/assessment answer
type Response struct {
	ID           uuid.UUID `db:"id" json:"id"`
	AssessmentID uuid.UUID `db:"assessment_id" json:"assessment_id"`
	QuestionID   string    `db:"question_id" json:"question_id"`
	Correct      bool      `db:"correct" json:"correct"`
	Confidence   float64   `db:"confidence" json:"confidence"` // 0-1
	ResponseTime  int       `db:"response_time_ms" json:"response_time_ms"`
}

// Submission represents a project/work submission
type Submission struct {
	ID           uuid.UUID `db:"id" json:"id"`
	AssessmentID uuid.UUID `db:"assessment_id" json:"assessment_id"`
	Content      string    `db:"content" json:"content"`
	URL          string    `db:"url" json:"url,omitempty"`      // External link

	// Grading
	Score        float64  `db:"score" json:"score"`
	Feedback     string   `db:"feedback" json:"feedback"`
	GradedAt     *time.Time `db:"graded_at" json:"graded_at,omitempty"`
	GradedBy     *uuid.UUID `db:"graded_by" json:"graded_by,omitempty"` // AI or human
}

// ChatAnalysis represents insights from chat conversations
type ChatAnalysis struct {
	NumMessages      int       `json:"num_messages"`
	ConceptsUsed     []string  `json:"concepts_used"`       // Technical terms used correctly
	QuestionsAsked   int       `json:"questions_asked"`
	ClarityScore     float64   `json:"clarity_score"`       // How well they explained
	HelpingOthers    float64   `json:"helping_others"`      // Did they help peers
}

// VoiceAnalysis represents insights from voice/speech
type VoiceAnalysis struct {
	NumInteractions  int       `json:"num_interactions"`
	ConfidenceScore  float64   `json:"confidence_score"`   // Speaking confidence
	ClarityScore     float64   `json:"clarity_score"`       // Explanation clarity
	FillerRatio      float64   `json:"filler_ratio"`        // Um/uh per minute
	Pace             float64   `json:"pace"`                // Words per minute
}

// ============================================================================
// SKILL TREE VISUALIZATION DATA
// ============================================================================

// StudentTreeSummary represents a student's progress across a skill tree
type StudentTreeSummary struct {
	TreeID             uuid.UUID              `json:"tree_id"`
	TreeTitle          string                 `json:"tree_title"`
	TreeIcon           string                 `json:"tree_icon"`

	// Overall progress
	TotalNodes         int                    `json:"total_nodes"`
	UnlockedNodes      int                    `json:"unlocked_nodes"`
	CompletedNodes     int                    `json:"completed_nodes"`
	TotalPoints        int                    `json:"total_points"`
	MaxPoints          int                    `json:"max_points"`
	ProgressPercent    float64                `json:"progress_percent"`

	// Current level
	OverallLevel       Proficiency            `json:"overall_level"`

	// Nodes with state
	Nodes              []NodeSummary          `json:"nodes"`

	// Next unlocks
	UnlockableNodes     []NodeSummary          `json:"unlockable_nodes"` // Can unlock soon
	BlockedNodes        []NodeSummary          `json:"blocked_nodes"`    // Need prerequisites
}

// NodeSummary represents a skill node with student progress
type NodeSummary struct {
	NodeID              uuid.UUID    `json:"node_id"`
	Slug                string       `json:"slug"`
	Title               string       `json:"title"`
	Icon                string       `json:"icon"`
	Position            NodePosition `json:"position"`

	// Requirements
	RequiredScore       float64     `json:"required_score"`
	RequiredNodes       []uuid.UUID `json:"required_nodes"`
	RequirementsMet     bool        `json:"requirements_met"`

	// Student progress
	Score               float64     `json:"score"`          // 0-100
	Points              int         `json:"points"`
	MaxPoints           int         `json:"max_points"`
	Level               Proficiency  `json:"level"`

	// Status
	Unlocked            bool        `json:"unlocked"`
	Completed           bool        `json:"completed"`
	IsCurrent           bool        `json:"is_current"`     // Is this what they're working on?

	// Progress toward next level
	NextLevel           Proficiency `json:"next_level,omitempty"`
	NextLevelProgress   float64     `json:"next_level_progress,omitempty"` // 0-1
	PointsToNextLevel    int         `json:"points_to_next_level,omitempty"`
}

// ============================================================================
// SKILL DEPENDENCIES
// ============================================================================

// Dependency represents a prerequisite relationship between nodes
type Dependency struct {
	ID           uuid.UUID  `db:"id" json:"id"`
	TreeID       uuid.UUID  `db:"tree_id" json:"tree_id"`
	RequiresNode uuid.UUID  `db:"requires_node" json:"requires_node"` // Must complete this first
	UnlocksNode  uuid.UUID  `db:"unlocks_node" json:"unlocks_node"`    // To unlock this

	Type         DependencyType `db:"type" json:"type"`
	Option       bool          `db:"option" json:"option"` // true = OR gate, false = AND gate

	CreatedAt    time.Time     `db:"created_at" json:"created_at"`
}

// DependencyType represents the kind of prerequisite
type DependencyType string

const (
	DependencyHard    DependencyType = "hard"    // Must complete first
	DependencySoft    DependencyType = "soft"    // Recommended first
	DependencyParallel DependencyType = "parallel" // Can do simultaneously
)

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

// GetProficiencyForScore returns the proficiency level for a given score
func GetProficiencyForScore(score float64) Proficiency {
	for _, level := range []Proficiency{
		ProficiencyMaster,
		ProficiencyExpert,
		ProficiencyAdept,
		ProficiencyApprentice,
		ProficiencyNovice,
	} {
		if score >= ProficiencyThresholds[level].MinScore {
			return level
		}
	}
	return ProficiencyUnknown
}

// GetConfig returns the configuration for a proficiency level
func GetConfig(level Proficiency) (minScore float64, icon, label, color string) {
	cfg := ProficiencyThresholds[level]
	return cfg.MinScore, cfg.Icon, cfg.Label, cfg.Color
}

// PointsToNextLevel calculates points needed to reach the next proficiency
func PointsToNextLevel(currentScore float64, currentLevel Proficiency) int {
	// Find next level
	levels := []Proficiency{
		ProficiencyNovice,
		ProficiencyApprentice,
		ProficiencyAdept,
		ProficiencyExpert,
		ProficiencyMaster,
	}

	nextLevelIdx := -1
	for i, level := range levels {
		if level == currentLevel {
			nextLevelIdx = i + 1
			break
		}
	}

	if nextLevelIdx >= len(levels) {
		return 0 // Already at max
	}

	nextLevel := levels[nextLevelIdx]
	nextScore := ProficiencyThresholds[nextLevel].MinScore
	diff := nextScore - currentScore

	// Rough conversion: 1 point per 1% score
	return int(diff)
}

// IsUnlocked checks if a student meets requirements to access a node
func IsUnlocked(studentSkill *StudentSkill, allSkills []StudentSkill, dependencies []Dependency) bool {
	// Must have required score
	if studentSkill.Score < 1 {
		return false
	}

	// Check all dependencies
	for _, dep := range dependencies {
		// Find the prerequisite node skill
		var hasPrereq bool
		for _, skill := range allSkills {
			if skill.NodeID == dep.RequiresNode {
				// For hard dependencies, must be substantially complete
				if dep.Type == DependencyHard && skill.Score < 70 {
					return false
				}
				// For soft dependencies, just need some progress
				if dep.Type == DependencySoft && skill.Score < 25 {
					return false
				}
				hasPrereq = true
				break
			}
		}
		if !hasPrereq && !dep.Option {
			return false
		}
	}

	return true
}
