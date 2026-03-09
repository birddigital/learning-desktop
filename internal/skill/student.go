// Package skill provides video game-style skill tree tracking.
// Progress is visible, measurable, and unlockable like RPG talent trees.
package skill

// ============================================================================
// STUDENT SKILLS TREE
// ============================================================================
// Academic mastery: How to learn, study, and excel in any subject
//
// Philosophy: School subjects change. The skill of learning is forever.
// Test out of these by demonstrating real learning ability, not grades.
//
// Assessment: Practical application + peer teaching + real problem solving
// ============================================================================

// Student Skills Tree
var StudentTree = TreeDefinition{
	Slug:        "student-skills",
	Title:       "Student Skills",
	Description: "Master the art of learning. Study smarter, not harder.",
	Icon:        "📚",
	Category:    CategorySoftSkills,
	Nodes: []NodeDefinition{
		{
			Slug:         "focus",
			Title:        "Deep Focus",
			Icon:         "🎯",
			Description:  "Lock in. Eliminate distractions. Do deep work.",
			Position:     NodePosition{Row: 0, Col: 0},
			MaxPoints:    150,
			RequiredScore: 0,
			RequiredNodes: []string{},
			Lessons:      []string{"focus-basics", "focus-deep-work"},
			Projects:     []string{"focus-pomodoro", "focus-environment"},
		},
		{
			Slug:         "note-taking",
			Title:        "Note Taking",
			Icon:         "📝",
			Description:  "Capture what matters. Organize knowledge. Recall anything.",
			Position:     NodePosition{Row: 1, Col: 0},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"focus"},
			Lessons:      []string{"notes-cornell", "notes-mindmap", "notes-coding"},
			Projects:     []string{"notes-system", "notes-review"},
		},
		{
			Slug:         "time-management",
			Title:        "Time Management",
			Icon:         "⏰",
			Description:  "Your time is your life. Spend it intentionally.",
			Position:     NodePosition{Row: 1, Col: 1},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"focus"},
			Lessons:      []string{"time-prioritization", "time-blocking", "time-energy"},
			Projects:     []string{"time-calendar", "time-audit"},
		},
		{
			Slug:         "reading-strategy",
			Title:        "Strategic Reading",
			Icon:         "📖",
			Description:  "Read faster. Remember more. Extract key insights.",
			Position:     NodePosition{Row: 2, Col: 0},
			MaxPoints:    150,
			RequiredScore: 60,
			RequiredNodes: []string{"note-taking"},
			Lessons:      []string{"reading-speed", "reading-comprehension", "reading-retention"},
			Projects:     []string{"reading-log", "reading-synthesis"},
		},
		{
			Slug:         "memory-techniques",
			Title:        "Memory Systems",
			Icon:         "🧠",
			Description:  "Never forget what you learn. Build a memory palace.",
			Position:     NodePosition{Row: 2, Col: 1},
			MaxPoints:    150,
			RequiredScore: 60,
			RequiredNodes: []string{"note-taking"},
			Lessons:      []string{"memory-palace", "memory-spaced", "memory-associations"},
			Projects:     []string{"memory-demo", "memory-cards"},
		},
		{
			Slug:         "test-taking",
			Title:        "Test Mastery",
			Icon:         "✅",
			Description:  "Perform when it counts. Manage anxiety. Show what you know.",
			Position:     NodePosition{Row: 3, Col: 0},
			MaxPoints:    150,
			RequiredScore: 70,
			RequiredNodes: []string{"reading-strategy", "memory-techniques"},
			Lessons:      []string{"testing-strategy", "testing-anxiety", "testing-types"},
			Projects:     []string{"testing-plan", "testing-practice"},
		},
		{
			Slug:         "research",
			Title:        "Research Skills",
			Icon:         "🔍",
			Description:  "Find truth. Verify sources. Build knowledge from evidence.",
			Position:     NodePosition{Row: 3, Col: 1},
			MaxPoints:    150,
			RequiredScore: 70,
			RequiredNodes: []string{"reading-strategy"},
			Lessons:      []string{"research-sources", "research-verification", "research-synthesis"},
			Projects:     []string{"research-paper", "research-presentation"},
		},
		{
			Slug:         "teaching",
			Title:        "Teaching Others",
			Icon:         "👨‍🏫",
			Description:  "The best way to learn is to teach. Master by explaining.",
			Position:     NodePosition{Row: 4, Col: 0},
			MaxPoints:    200,
			RequiredScore: 75,
			RequiredNodes: []string{"test-taking", "research"},
			Lessons:      []string{"teaching-simplify", "teaching-examples", "teaching-feedback"},
			Projects:     []string{"teaching-session", "teaching-content"},
		},
		{
			Slug:         "metacognition",
			Title:        "Metacognition",
			Icon:         "🔮",
			Description:  "Think about your thinking. Know how you learn. Optimize.",
			Position:     NodePosition{Row: 4, Col: 1},
			MaxPoints:    200,
			RequiredScore: 80,
			RequiredNodes: []string{"teaching"},
			Lessons:      []string{"meta-awareness", "meta-strategies", "meta-adaptation"},
			Projects:     []string{"meta-journal", "meta-plan"},
		},
	},
}

// Student skills unlock by demonstrated performance, not seat time
// You prove you can focus by focusing. You prove you can teach by teaching.
