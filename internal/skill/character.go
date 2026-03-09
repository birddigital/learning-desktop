// Package skill provides video game-style skill tree tracking.
// Progress is visible, measurable, and unlockable like RPG talent trees.
package skill

// ============================================================================
// CHARACTER & MANHOOD SKILL TREE
// ============================================================================
// For young men 13-17: Building character, integrity, responsibility
//
// Philosophy: You earn these through demonstrated action, not reading.
// You can't "complete" these skills. You demonstrate them repeatedly
// until they become who you are.
//
// Assessment: Real-world application + peer validation + mentor observation
// ============================================================================

// Character Tree: Building Better Men
var CharacterTree = TreeDefinition{
	Slug:        "character-manhood",
	Title:       "Character & Manhood",
	Description: "Build the foundation of who you are. Your word, your actions, your standards.",
	Icon:        "🦁",
	Category:    CategorySoftSkills,
	Nodes: []NodeDefinition{
		{
			Slug:         "integrity",
			Title:        "Integrity",
			Icon:         "🤝",
			Description:  "Your word is your bond. You do what you say, even when no one is watching.",
			Position:     NodePosition{Row: 0, Col: 0},
			MaxPoints:    200,
			RequiredScore: 0,
			RequiredNodes: []string{},
			Lessons:      []string{"integrity-what", "integrity-when-it-counts"},
			Projects:     []string{"integrity-30day", "integrity-peer-valid"},
		},
		{
			Slug:         "responsibility",
			Title:        "Ownership",
			Icon:         "💪",
			Description:  "You own your mistakes. You own your choices. No excuses.",
			Position:     NodePosition{Row: 1, Col: 0},
			MaxPoints:    200,
			RequiredScore: 50,
			RequiredNodes: []string{"integrity"},
			Lessons:      []string{"ownership-mindset", "ownership-accountability"},
			Projects:     []string{"ownership-audit", "ownership-repair"},
		},
		{
			Slug:         "discipline",
			Title:        "Self-Discipline",
			Icon:         "⚔️",
			Description:  "You do what needs to be done, even when you don't feel like it.",
			Position:     NodePosition{Row: 1, Col: 1},
			MaxPoints:    200,
			RequiredScore: 50,
			RequiredNodes: []string{"integrity"},
			Lessons:      []string{"discipline-foundation", "discipline-habits"},
			Projects:     []string{"discipline-21day", "discipline-routine"},
		},
		{
			Slug:         "respect",
			Title:        "Respect",
			Icon:         "👊",
			Description:  "Give it to earn it. Self-respect first. Then others.",
			Position:     NodePosition{Row: 2, Col: 0},
			MaxPoints:    150,
			RequiredScore: 60,
			RequiredNodes: []string{"responsibility"},
			Lessons:      []string{"respect-self", "respect-others", "respect-authority"},
			Projects:     []string{"respect-audit", "respect-actions"},
		},
		{
			Slug:         "resilience",
			Title:        "Resilience",
			Icon:         "🔥",
			Description:  "Fall down 7 times, stand up 8. Adapt and overcome.",
			Position:     NodePosition{Row: 2, Col: 1},
			MaxPoints:    200,
			RequiredScore: 60,
			RequiredNodes: []string{"discipline"},
			Lessons:      []string{"resilience-mindset", "resilience-failure"},
			Projects:     []string{"resilience-setback", "resilience-comeback"},
		},
		{
			Slug:         "leadership",
			Title:        "Leadership",
			Icon:         "👑",
			Description:  "Others follow because they want to, not because they have to.",
			Position:     NodePosition{Row: 3, Col: 0},
			MaxPoints:    250,
			RequiredScore: 75,
			RequiredNodes: []string{"respect", "resilience"},
			Lessons:      []string{"leadership-serve", "leadership-vision"},
			Projects:     []string{"leadership-project", "leadership-mentor"},
		},
		{
			Slug:         "purpose",
			Title:        "Purpose",
			Icon:         "🎯",
			Description:  "Know why you're here. What you're building. Who you're becoming.",
			Position:     NodePosition{Row: 3, Col: 1},
			MaxPoints:    200,
			RequiredScore: 75,
			RequiredNodes: []string{"discipline", "resilience"},
			Lessons:      []string{"purpose-discovery", "purpose-alignment"},
			Projects:     []string{"purpose-statement", "purpose-map"},
		},
		{
			Slug:         "brotherhood",
			Title:        "Brotherhood",
			Icon:         "🤝",
			Description:  "Build your circle. Lift others as you climb.",
			Position:     NodePosition{Row: 4, Col: 0},
			MaxPoints:    200,
			RequiredScore: 80,
			RequiredNodes: []string{"leadership", "respect"},
			Lessons:      []string{"brotherhood-trust", "brotherhood-loyalty"},
			Projects:     []string{"brotherhood-circle", "brotherhood-contribution"},
		},
		{
			Slug:         "legacy",
			Title:        "Legacy Thinking",
			Icon:         "🏛️",
			Description:  "Make decisions today that your future self will thank you for.",
			Position:     NodePosition{Row: 4, Col: 1},
			MaxPoints:    250,
			RequiredScore: 85,
			RequiredNodes: []string{"purpose", "brotherhood"},
			Lessons:      []string{"legacy-long-game", "legacy-impact"},
			Projects:     []string{"legacy-plan", "legacy-daily"},
		},
	},
}

// Assessment methods for character skills (real-world proof required)
const (
	// CharacterAssessmentPeer = "peer"       // Brothers validate your actions
	// CharacterAssessmentMentor = "mentor"   // Community leader observes
	// CharacterAssessmentProof = "proof"     // Documentary evidence
	// CharacterAssessmentReflection = "reflection" // Written self-assessment
)
