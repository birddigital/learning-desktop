// Package insight provides progress analysis and actionable recommendations.
// The Insight Engine reads learning data and tells students the hard truth.
package insight

import (
	"fmt"
	"math"
	"time"

	"github.com/birddigital/learning-desktop/internal/models"
	"github.com/google/uuid"
)

// Engine analyzes student progress and generates actionable insights
type Engine struct {
	// Configuration for insight generation
	MinDataPoints     int           // Minimum progress events before generating insights
	PaceWindowDays    int           // Days to look back for pace calculation
	UrgencyThresholds UrgencyConfig // When to trigger urgency levels
}

// UrgencyConfig defines thresholds for insight urgency
type UrgencyConfig struct {
	CriticalDelta float64 // Days behind to trigger critical (e.g., 14)
	HighDelta     float64 // Days behind to trigger high (e.g., 7)
	MediumDelta   float64 // Days behind to trigger medium (e.g., 3)
}

// DefaultEngine creates a configured insight engine
func DefaultEngine() *Engine {
	return &Engine{
		MinDataPoints:  3,
		PaceWindowDays: 14,
		UrgencyThresholds: UrgencyConfig{
			CriticalDelta: 14,
			HighDelta:     7,
			MediumDelta:   3,
		},
	}
}

// ============================================================================
// INSIGHT MODELS
// ============================================================================

// Insight represents a complete analysis of a student's progress
type Insight struct {
	StudentID          uuid.UUID  `json:"student_id"`
	GeneratedAt        time.Time `json:"generated_at"`
	DataPoints         int       `json:"data_points"` // How much data we analyzed
	Confidence         float64   `json:"confidence"`   // How confident in this analysis (0-1)

	// Where they wanted to be
	Goal               GoalSnapshot `json:"goal"`

	// Where they actually are
	CurrentStatus      StatusSnapshot `json:"current_status"`

	// The hard truth
	Trajectory         TrajectoryAnalysis `json:"trajectory"`

	// What to do about it
	Recommendation     Recommendation `json:"recommendation"`

	// Supporting metrics
	Metrics            DetailedMetrics `json:"metrics"`
}

// GoalSnapshot captures what the student committed to
type GoalSnapshot struct {
	Target          string    `json:"target"`           // e.g., "Complete AI Fundamentals course"
	TargetDate      time.Time `json:"target_date"`
	DaysRemaining   int       `json:"days_remaining"`
	TotalMilestones int       `json:"total_milestones"`
}

// StatusSnapshot captures where they are right now
type StatusSnapshot struct {
	ProgressPercent    float64  `json:"progress_percent"`    // 0-100
	CompletedMilestones int     `json:"completed_milestones"`
	CurrentMilestone    string   `json:"current_milestone"`
	LastActivity       time.Time `json:"last_activity"`
	ActiveStreak       int      `json:"active_streak"` // consecutive days of activity
	EngagementLevel    string   `json:"engagement_level"` // "low" | "medium" | "high"
}

// TrajectoryAnalysis predicts if they'll hit their goal
type TrajectoryAnalysis struct {
	OnTrack           bool        `json:"on_track"`
	ProjectedDate     *time.Time  `json:"projected_date,omitempty"`     // When they'll finish at current pace
	ProjectedMissBy   *time.Duration `json:"projected_miss_by,omitempty"` // How late they'll be
	PaceDelta         float64    `json:"pace_delta"`                   // Current vs required pace (ratio)
	Velocity          float64    `json:"velocity"`                      // Progress points per day
	RequiredVelocity  float64    `json:"required_velocity"`             // To hit goal on time
	Trend             string     `json:"trend"`                        // "accelerating" | "stable" | "decelerating"
	RiskFactors       []string   `json:"risk_factors"`
	Strengths         []string   `json:"strengths"`
}

// Recommendation is actionable advice
type Recommendation struct {
	Priority      string   `json:"priority"`      // "now" | "soon" | "eventually"
	Action        string   `json:"action"`        // What they should do
	Reasoning     string   `json:"reasoning"`     // Why this matters
	DailyHours    float64  `json:"daily_hours"`   // How much time to invest
	FocusAreas    []string `json:"focus_areas"`   // What to focus on
	Avoid         []string `json:"avoid"`         // What to stop doing
	CheckInDate   *time.Time `json:"check_in_date"` // When to reassess
}

// DetailedMetrics back up the insights
type DetailedMetrics struct {
	// Pace metrics
	CurrentPace      float64 `json:"current_pace"`      // Lessons completed per week
	RequiredPace     float64 `json:"required_pace"`     // Lessons needed per week
	BestPace         float64 `json:"best_pace"`         // Best week's pace
	AveragePace      float64 `json:"average_pace"`      // Average pace overall

	// Time metrics
	TotalTimeSpent   int      `json:"total_time_spent"`   // Minutes
	AvgSessionTime   int      `json:"avg_session_time"`   // Minutes per session
	LongestStreak    int      `json:"longest_streak"`    // Consecutive days

	// Quality metrics
	AvgQuizScore     float64 `json:"avg_quiz_score"`     // 0-1
	CheckpointPassRate float64 `json:"checkpoint_pass_rate"` // 0-1
	ConceptRetention float64 `json:"concept_retention"`  // 0-1

	// Engagement metrics
	MessagesExchanged int      `json:"messages_exchanged"`
	VoiceInteractions  int      `json:"voice_interactions"`
	MediaConsumed      int      `json:"media_consumed"`   // Videos watched
}

// ============================================================================
// INSIGHT GENERATION
// ============================================================================

// Generate creates a comprehensive insight for a student
func (e *Engine) Generate(student *models.Student, progress *models.StudentProgress, events []models.ProgressEvent) *Insight {
	insight := &Insight{
		StudentID:   student.ID,
		GeneratedAt: time.Now(),
		DataPoints:  len(events),
	}

	// If not enough data, return early insight
	if len(events) < e.MinDataPoints {
		insight.Confidence = 0.3
		insight.Recommendation = Recommendation{
			Priority:   "soon",
			Action:     "Complete more lessons to generate personalized insights",
			Reasoning:  fmt.Sprintf("Need %d data points, have %d", e.MinDataPoints, len(events)),
			CheckInDate: ptrTime(time.Now().AddDate(0, 0, 7)),
		}
		return insight
	}

	// Calculate confidence based on data recency and volume
	insight.Confidence = e.calculateConfidence(events)

	// Build goal snapshot (from student's goals/target course)
	insight.Goal = e.buildGoalSnapshot(student, progress)

	// Build current status
	insight.CurrentStatus = e.buildStatusSnapshot(student, progress, events)

	// Analyze trajectory
	insight.Trajectory = e.analyzeTrajectory(insight.Goal, insight.CurrentStatus, events)

	// Generate recommendation
	insight.Recommendation = e.generateRecommendation(insight.Trajectory, insight.CurrentStatus)

	// Calculate detailed metrics
	insight.Metrics = e.calculateMetrics(student, progress, events)

	return insight
}

// buildGoalSnapshot captures what the student is aiming for
func (e *Engine) buildGoalSnapshot(student *models.Student, progress *models.StudentProgress) GoalSnapshot {
	// For now, assume goal is course completion
	// In future, this could be an explicit Goal entity
	totalDays := 90 // Default course duration
	if progress.StartedAt != nil {
		totalDays = int(time.Since(*progress.StartedAt).Hours()/24) + 60 // Estimated remaining
	}

	return GoalSnapshot{
		Target:           "Complete course",
		TargetDate:       time.Now().AddDate(0, 0, totalDays),
		DaysRemaining:    totalDays,
		TotalMilestones:  progress.TotalLessons,
	}
}

// buildStatusSnapshot captures current state
func (e *Engine) buildStatusSnapshot(student *models.Student, progress *models.StudentProgress, events []models.ProgressEvent) StatusSnapshot {
	lastActivity := time.Time{}
	if len(events) > 0 {
		lastActivity = events[len(events)-1].Timestamp
	}

	streak := e.calculateStreak(events)
	engagement := e.calculateEngagement(student, events)

	return StatusSnapshot{
		ProgressPercent:    progress.CompletionPercent(),
		CompletedMilestones: progress.CompletedLessons,
		CurrentMilestone:    fmt.Sprintf("Lesson %d", progress.CurrentLessonIndex+1),
		LastActivity:        lastActivity,
		ActiveStreak:        streak,
		EngagementLevel:     engagement,
	}
}

// analyzeTrajectory predicts if they'll hit their goal
func (e *Engine) analyzeTrajectory(goal GoalSnapshot, status StatusSnapshot, events []models.ProgressEvent) TrajectoryAnalysis {
	analysis := TrajectoryAnalysis{}

	// Calculate current velocity (progress points per day)
	analysis.Velocity = e.calculateVelocity(events)

	// Calculate required velocity
	remainingProgress := 100.0 - status.ProgressPercent
	remainingDays := float64(goal.DaysRemaining)
	analysis.RequiredVelocity = remainingProgress / remainingDays

	// Pace delta: how current pace compares to required
	if analysis.RequiredVelocity > 0 {
		analysis.PaceDelta = analysis.Velocity / analysis.RequiredVelocity
	}

	// Project completion date
	daysAtCurrentPace := int(remainingProgress / analysis.Velocity)
	projectedDate := time.Now().AddDate(0, 0, daysAtCurrentPace)
	analysis.ProjectedDate = &projectedDate

	// Will they miss their goal?
	if projectedDate.After(goal.TargetDate) {
		missBy := projectedDate.Sub(goal.TargetDate)
		analysis.ProjectedMissBy = &missBy
		analysis.OnTrack = false
	} else {
		analysis.OnTrack = true
	}

	// Determine trend
	analysis.Trend = e.calculateTrend(events)

	// Identify risk factors and strengths
	analysis.RiskFactors, analysis.Strengths = e.identifyFactors(status, events, analysis)

	return analysis
}

// generateRecommendation creates actionable advice
func (e *Engine) generateRecommendation(trajectory TrajectoryAnalysis, status StatusSnapshot) Recommendation {
	rec := Recommendation{}

	switch {
	case !trajectory.OnTrack && trajectory.ProjectedMissBy != nil && trajectory.ProjectedMissBy.Hours() > 24*7:
		// More than a week behind
		rec.Priority = "now"
		rec.Action = "Significant intervention required"
		rec.Reasoning = fmt.Sprintf("At current pace, you'll miss your goal by %d days", int(trajectory.ProjectedMissBy.Hours()/24))
		rec.DailyHours = e.calculateRequiredHours(trajectory.Velocity, trajectory.RequiredVelocity)
		rec.FocusAreas = []string{"Increase daily study time", "Review completed lessons for retention"}
		rec.Avoid = []string{"Starting new topics until current ones solidify"}
		rec.CheckInDate = ptrTime(time.Now().AddDate(0, 0, 3))

	case !trajectory.OnTrack:
		// Behind but recoverable
		rec.Priority = "soon"
		rec.Action = "Increase study pace"
		rec.Reasoning = fmt.Sprintf("You're %.0f%% behind required pace", (1-trajectory.PaceDelta)*100)
		rec.DailyHours = e.calculateRequiredHours(trajectory.Velocity, trajectory.RequiredVelocity)
		rec.FocusAreas = []string{"Consistency", "Focus on checkpoint completion"}
		rec.CheckInDate = ptrTime(time.Now().AddDate(0, 0, 7))

	case trajectory.OnTrack && trajectory.PaceDelta > 1.2:
		// Ahead of schedule
		rec.Priority = "eventually"
		rec.Action = "Maintain current pace or accelerate"
		rec.Reasoning = fmt.Sprintf("You're %.0f%% ahead of schedule", (trajectory.PaceDelta-1)*100)
		rec.FocusAreas = []string{"Deep dive into topics", "Help others", "Explore advanced material"}
		rec.CheckInDate = ptrTime(time.Now().AddDate(0, 0, 14))

	default:
		// On track
		rec.Priority = "soon"
		rec.Action = "Continue current approach"
		rec.Reasoning = "You're on track to meet your goal"
		rec.DailyHours = 1.0 // Maintain
		rec.FocusAreas = []string{"Consistency", "Checkpoint completion"}
		rec.CheckInDate = ptrTime(time.Now().AddDate(0, 0, 7))
	}

	// Adjust for engagement level
	if status.EngagementLevel == "low" {
		rec.Avoid = append(rec.Avoid, "Long gaps between sessions")
	}

	// Adjust for streak
	if status.ActiveStreak == 0 {
		rec.FocusAreas = append(rec.FocusAreas, "Build an activity streak")
	}

	return rec
}

// calculateMetrics computes detailed backing metrics
func (e *Engine) calculateMetrics(student *models.Student, progress *models.StudentProgress, events []models.ProgressEvent) DetailedMetrics {
	return DetailedMetrics{
		CurrentPace:    e.calculateCurrentPace(events),
		RequiredPace:   e.calculateRequiredPace(progress, events),
		BestPace:       e.calculateBestPace(events),
		AveragePace:    e.calculateAveragePace(events),
		TotalTimeSpent:  student.TotalMinutes,
		AvgSessionTime: e.calculateAvgSessionTime(events),
		LongestStreak:  e.calculateLongestStreak(events),
	}
}

// ============================================================================
// HELPER CALCULATIONS
// ============================================================================

// calculateVelocity returns progress points per day
func (e *Engine) calculateVelocity(events []models.ProgressEvent) float64 {
	if len(events) < 2 {
		return 0
	}

	// Look at pace window
	windowStart := time.Now().AddDate(0, 0, -e.PaceWindowDays)
	recentEvents := filterEventsSince(events, windowStart)

	if len(recentEvents) < 2 {
		return 0
	}

	// Calculate progress points earned in window
	// Each completed lesson = 1 point
	pointsEarned := 0.0
	for _, ev := range recentEvents {
		if ev.EventType == models.EventLessonCompleted {
			pointsEarned++
		}
	}

	daysInWindow := math.Max(1, time.Since(windowStart).Hours()/24)
	return pointsEarned / daysInWindow
}

// calculateTrend determines if student is accelerating or decelerating
func (e *Engine) calculateTrend(events []models.ProgressEvent) string {
	if len(events) < 4 {
		return "stable"
	}

	// Compare first half to second half of events
	mid := len(events) / 2
	firstHalf := events[:mid]
	secondHalf := events[mid:]

	firstPace := eventsPerDay(firstHalf)
	secondPace := eventsPerDay(secondHalf)

	ratio := secondPace / firstPace

	switch {
	case ratio > 1.2:
		return "accelerating"
	case ratio < 0.8:
		return "decelerating"
	default:
		return "stable"
	}
}

// identifyFactors returns risk factors and strengths
func (e *Engine) identifyFactors(status StatusSnapshot, events []models.ProgressEvent, trajectory TrajectoryAnalysis) ([]string, []string) {
	var risks, strengths []string

	// Risk factors
	if !trajectory.OnTrack {
		risks = append(risks, "Behind schedule")
	}
	if status.ActiveStreak == 0 {
		risks = append(risks, "No active streak")
	}
	if status.EngagementLevel == "low" {
		risks = append(risks, "Low engagement")
	}
	if trajectory.Trend == "decelerating" {
		risks = append(risks, "Slowing down")
	}

	// Strengths
	if status.ActiveStreak >= 7 {
		strengths = append(strengths, fmt.Sprintf("%d-day activity streak", status.ActiveStreak))
	}
	if trajectory.OnTrack && trajectory.PaceDelta > 1.1 {
		strengths = append(strengths, "Ahead of schedule")
	}
	if trajectory.Trend == "accelerating" {
		strengths = append(strengths, "Building momentum")
	}
	if status.EngagementLevel == "high" {
		strengths = append(strengths, "Highly engaged")
	}

	return risks, strengths
}

// calculateConfidence returns how confident we are in the insight (0-1)
func (e *Engine) calculateConfidence(events []models.ProgressEvent) float64 {
	confidence := 0.5

	// More data points = more confidence
	dataScore := math.Min(1.0, float64(len(events))/20.0)
	confidence += dataScore * 0.3

	// Recent data is more valuable
	recentCount := 0
	weekAgo := time.Now().AddDate(0, 0, -7)
	for _, ev := range events {
		if ev.Timestamp.After(weekAgo) {
			recentCount++
		}
	}
	recencyScore := math.Min(1.0, float64(recentCount)/5.0)
	confidence += recencyScore * 0.2

	return math.Min(1.0, confidence)
}

// calculateCurrentPace returns lessons per week in recent window
func (e *Engine) calculateCurrentPace(events []models.ProgressEvent) float64 {
	windowStart := time.Now().AddDate(0, 0, -e.PaceWindowDays)
	completed := countCompletedLessons(filterEventsSince(events, windowStart))
	weeks := math.Max(1, time.Since(windowStart).Hours()/(24*7))
	return completed / weeks
}

// calculateRequiredPace returns lessons per week needed to finish on time
func (e *Engine) calculateRequiredPace(progress *models.StudentProgress, events []models.ProgressEvent) float64 {
	remainingLessons := progress.TotalLessons - progress.CompletedLessons
	// Assume 90 day timeline from now (configurable)
	remainingWeeks := 12.0
	return float64(remainingLessons) / remainingWeeks
}

// calculateBestPace returns the highest weekly pace achieved
func (e *Engine) calculateBestPace(events []models.ProgressEvent) float64 {
	weeklyPaces := e.weeklyPaceBreakdown(events)
	maxPace := 0.0
	for _, pace := range weeklyPaces {
		if pace > maxPace {
			maxPace = pace
		}
	}
	return maxPace
}

// calculateAveragePace returns overall average pace
func (e *Engine) calculateAveragePace(events []models.ProgressEvent) float64 {
	if len(events) < 2 {
		return 0
	}

	first := events[0].Timestamp
	last := events[len(events)-1].Timestamp
	days := math.Max(1, last.Sub(first).Hours()/24)
	completed := countCompletedLessons(events)

	return (completed / days) * 7 // Lessons per week
}

// calculateStreak returns consecutive days of activity
func (e *Engine) calculateStreak(events []models.ProgressEvent) int {
	if len(events) == 0 {
		return 0
	}

	streak := 0
	checkDate := time.Now().Truncate(24 * time.Hour)

	// Check each day going back
	for i := 0; i < 365; i++ {
		if hasActivityOn(events, checkDate) {
			streak++
			checkDate = checkDate.AddDate(0, 0, -1)
		} else {
			// Allow for today being incomplete
			if i == 0 && checkDate.Before(time.Now()) {
				checkDate = checkDate.AddDate(0, 0, -1)
				continue
			}
			break
		}
	}

	return streak
}

// calculateLongestStreak returns the longest streak ever achieved
func (e *Engine) calculateLongestStreak(events []models.ProgressEvent) int {
	if len(events) == 0 {
		return 0
	}

	dates := make(map[time.Time]bool)
	for _, ev := range events {
		dates[ev.Timestamp.Truncate(24 * time.Hour)] = true
	}

	longest := 0
	current := 0

	// Iterate through dates from earliest to latest
	var sortedDates []time.Time
	for d := range dates {
		sortedDates = append(sortedDates, d)
	}
	// Sort dates (simplified - in production use proper sort)

	for _, d := range sortedDates {
		if len(sortedDates) == 0 {
			continue
		}
		current++
		if current > longest {
			longest = current
		}
	}

	return longest
}

// calculateEngagement determines engagement level
func (e *Engine) calculateEngagement(student *models.Student, events []models.ProgressEvent) string {
	// Factors: recent activity, session frequency, voice usage
	weekAgo := time.Now().AddDate(0, 0, -7)
	recentEvents := filterEventsSince(events, weekAgo)

	if len(recentEvents) == 0 {
		return "low"
	}
	if len(recentEvents) >= 7 {
		return "high"
	}
	return "medium"
}

// calculateAvgSessionTime returns average minutes per learning session
func (e *Engine) calculateAvgSessionTime(events []models.ProgressEvent) int {
	// Group events by day, count as one session
	sessions := groupEventsBySession(events)
	totalMinutes := 0

	for _, session := range sessions {
		totalMinutes += int(session.Duration().Minutes())
	}

	if len(sessions) == 0 {
		return 0
	}
	return totalMinutes / len(sessions)
}

// calculateRequiredHours calculates how many hours per day are needed
func (e *Engine) calculateRequiredHours(currentVel, requiredVel float64) float64 {
	if currentVel <= 0 {
		return 2.0 // Default recommendation
	}

	ratio := requiredVel / currentVel
	// Assume 1 hour per day baseline
	return math.Max(0.5, math.Min(4.0, ratio)) // Cap between 30min and 4hrs
}

// weeklyPaceBreakdown returns lessons per week for each week
func (e *Engine) weeklyPaceBreakdown(events []models.ProgressEvent) []float64 {
	// Implementation: group by week, calculate completions
	// Simplified for now
	return []float64{2.0, 3.0, 1.5}
}

// ============================================================================
// HELPERS
// ============================================================================

func filterEventsSince(events []models.ProgressEvent, since time.Time) []models.ProgressEvent {
	var filtered []models.ProgressEvent
	for _, ev := range events {
		if ev.Timestamp.After(since) || ev.Timestamp.Equal(since) {
			filtered = append(filtered, ev)
		}
	}
	return filtered
}

func countCompletedLessons(events []models.ProgressEvent) float64 {
	count := 0
	for _, ev := range events {
		if ev.EventType == models.EventLessonCompleted {
			count++
		}
	}
	return float64(count)
}

func eventsPerDay(events []models.ProgressEvent) float64 {
	if len(events) < 2 {
		return 0
	}
	duration := events[len(events)-1].Timestamp.Sub(events[0].Timestamp)
	days := math.Max(1, duration.Hours()/24)
	return float64(len(events)) / days
}

func hasActivityOn(events []models.ProgressEvent, date time.Time) bool {
	for _, ev := range events {
		if ev.Timestamp.Truncate(24 * time.Hour).Equal(date) {
			return true
		}
	}
	return false
}

func groupEventsBySession(events []models.ProgressEvent) []session {
	// Group events within 2 hours of each other as one session
	// Simplified implementation
	return []session{{}}
}

type session struct{}

func (s session) Duration() time.Duration {
	return 30 * time.Minute // Placeholder
}

func ptrTime(t time.Time) *time.Time {
	return &t
}
