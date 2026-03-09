// Package skill provides video game-style skill tree tracking.
// Progress is visible, measurable, and unlockable like RPG talent trees.
package skill

// ============================================================================
// ENTREPRENEURSHIP SKILL TREE
// ============================================================================
// Business building: Create value, capture value, build wealth
//
// Philosophy: Entrepreneurship is a skill, not a trait. It can be learned.
// Test out by making real sales, solving real problems, creating real value.
//
// Assessment: Revenue generated + problems solved + customers served
// ============================================================================

// Entrepreneurship Tree
var EntrepreneurshipTree = TreeDefinition{
	Slug:        "entrepreneurship",
	Title:       "Entrepreneurship",
	Description: "Build businesses. Create value. Own your economic future.",
	Icon:        "💼",
	Category:    CategoryCareer,
	Nodes: []NodeDefinition{
		{
			Slug:         "opportunity-spotting",
			Title:        "See Opportunities",
			Icon:         "👁️",
			Description:  "Problems are opportunities in disguise. Train your eye.",
			Position:     NodePosition{Row: 0, Col: 0},
			MaxPoints:    150,
			RequiredScore: 0,
			RequiredNodes: []string{},
			Lessons:      []string{"opportunity-observation", "opportunity-problems", "opportunity-gaps"},
			Projects:     []string{"opportunity-log", "opportunity-validation"},
		},
		{
			Slug:         "value-creation",
			Title:        "Create Value",
			Icon:         "🔨",
			Description:  "Solve real problems. Build things people actually want.",
			Position:     NodePosition{Row: 1, Col: 0},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"opportunity-spotting"},
			Lessons:      []string{"value-problems", "value-solutions", "value-mvp"},
			Projects:     []string{"value-proto", "value-feedback"},
		},
		{
			Slug:         "money-mindset",
			Title:        "Money Mindset",
			Icon:         "💰",
			Description:  "Understand money. Make it work for you. Build wealth.",
			Position:     NodePosition{Row: 1, Col: 1},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"opportunity-spotting"},
			Lessons:      []string{"money-basics", "money-assets", "money-compounding"},
			Projects:     []string{"money-plan", "money-first-investment"},
		},
		{
			Slug:         "sales",
			Title:        "Sales",
			Icon:         "🤝",
			Description:  "Communication is currency. Persuade authentically. Close deals.",
			Position:     NodePosition{Row: 2, Col: 0},
			MaxPoints:    200,
			RequiredScore: 60,
			RequiredNodes: []string{"value-creation"},
			Lessons:      []string{"sales-foundation", "sales-communication", "sales-closing"},
			Projects:     []string{"sales-pitch", "sales-first-deal"},
		},
		{
			Slug:         "marketing",
			Title:        "Marketing",
			Icon:         "📢",
			Description:  "Tell your story. Reach your people. Build an audience.",
			Position:     NodePosition{Row: 2, Col: 1},
			MaxPoints:    150,
			RequiredScore: 60,
			RequiredNodes: []string{"value-creation"},
			Lessons:      []string{"marketing-positioning", "marketing-channels", "marketing-content"},
			Projects:     []string{"marketing-campaign", "marketing-measurement"},
		},
		{
			Slug:         "financial-literacy",
			Title:        "Financial Literacy",
			Icon:         "📊",
			Description:  "Read numbers. Make decisions based on data, not hope.",
			Position:     NodePosition{Row: 3, Col: 0},
			MaxPoints:    150,
			RequiredScore: 70,
			RequiredNodes: []string{"sales", "money-mindset"},
			Lessons:      []string{"finance-statements", "finance-cashflow", "finance-metrics"},
			Projects:     []string{"finance-spreadsheet", "finance-forecast"},
		},
		{
			Slug:         "legal-basics",
			Title:        "Legal Foundation",
			Icon:         "⚖️",
			Description:  "Protect yourself. Understand contracts. Entity structure.",
			Position:     NodePosition{Row: 3, Col: 1},
			MaxPoints:    150,
			RequiredScore: 70,
			RequiredNodes: []string{"marketing"},
			Lessons:      []string{"legal-entities", "legal-contracts", "legal-ip"},
			Projects:     []string{"legal-checklist", "legal-entity-setup"},
		},
		{
			Slug:         "team-building",
			Title:        "Build Teams",
			Icon:         "👥",
			Description:  "You can't do it alone. Hire right. Delegate well. Lead.",
			Position:     NodePosition{Row: 4, Col: 0},
			MaxPoints:    200,
			RequiredScore: 75,
			RequiredNodes: []string{"financial-literacy", "legal-basics"},
			Lessons:      []string{"team-hiring", "team-culture", "team-delegation"},
			Projects:     []string{"team-org-chart", "team-hiring-plan"},
		},
		{
			Slug:         "scaling",
			Title:        "Scale Up",
			Icon:         "📈",
			Description:  "Turn what works into systems. Grow without breaking.",
			Position:     NodePosition{Row: 4, Col: 1},
			MaxPoints:    200,
			RequiredScore: 80,
			RequiredNodes: []string{"team-building"},
			Lessons:      []string{"scaling-systems", "scaling-automation", "scaling-franchise"},
			Projects:     []string{"scaling-manual", "scaling-playbook"},
		},
	},
}

// Entrepreneurship is proven by results:
// - Made a sale? Prove it.
// - Solved a problem? Show the solution.
// - Built something? Demo it.
// - Created value? Customer testimonials.
//
// No participation trophies. No "I read about business."
// Either you created value or you didn't.
