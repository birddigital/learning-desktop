// Package skill provides video game-style skill tree tracking.
// Progress is visible, measurable, and unlockable like RPG talent trees.
package skill

// ============================================================================
// EXAMPLE SKILL TREES
// ============================================================================
//
// These are example skill tree definitions for the "Get Ahead of AI" course.
// In production, these would be stored in the database.
//
// Visualization:
//
//                    ┌─────────────────────────┐
//                    │     AI FUNDAMENTALS      │
//                    │    Master all 6 trees    │
//                    └───────────┬─────────────┘
//                                │
//      ┌─────────────────────────┼─────────────────────────┐
//      │                         │                         │
//  ┌───▼────┐              ┌────▼─────┐              ┌───▼────┐
//  │ PROMPT │              │  CONCEPTS│              │ MODELS │
//  │ENGINE  │              │   & LOGIC│              │&  DATA │
//  └───┬────┘              └────┬─────┘              └───┬────┘
//      │                         │                         │
//   ┌───▼────────┐       ┌─────▼───────┐       ┌─────────▼──┐
//   │Basic Prompt│       │  How LLMs  │       │  Vector   │
//   │  Writing   │       │   Work     │       │ Databases │
//   └────────────┘       └─────────────┘       └───────────┘
//
// Each node shows:
// - 🌱📚🎯⭐👑 Current proficiency level
// - Progress bar to next level
// - Lock icon if not yet unlocked
// - Prerequisite lines connecting nodes
//
// ============================================================================

// Example Tree: Prompt Engineering
const ExamplePromptTree = TreeDefinition{
	Slug:        "prompt-engineering",
	Title:       "Prompt Engineering",
	Description: "Master the art of communicating with AI effectively",
	Icon:        "💬",
	Category:    CategoryTechnical,
	Nodes: []NodeDefinition{
		{
			Slug:         "basic-prompts",
			Title:        "Basic Prompt Writing",
			Icon:         "📝",
			Description:  "Write clear, effective prompts for common tasks",
			Position:     NodePosition{Row: 0, Col: 0},
			MaxPoints:    100,
			RequiredScore: 0,
			Lessons:      []string{"prompt-basics", "prompt-structure"},
		},
		{
			Slug:         "chain-of-thought",
			Title:        "Chain of Thought",
			Icon:         "🔗",
			Description:  "Guide AI through step-by-step reasoning",
			Position:     NodePosition{Row: 1, Col: 0},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"basic-prompts"},
			Lessons:      []string{"cot-intro", "cot-practice"},
		},
		{
			Slug:         "few-shot-prompting",
			Title:        "Few-Shot Prompting",
			Icon:         "🎯",
			Description:  "Give examples to improve AI accuracy",
			Position:     NodePosition{Row: 1, Col: 1},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"basic-prompts"},
			Lessons:      []string{"few-shot-intro", "few-shot-patterns"},
		},
		{
			Slug:         "advanced-techniques",
			Title:        "Advanced Techniques",
			Icon:         "🚀",
			Description:  "Role-playing, constraint setting, output formatting",
			Position:     NodePosition{Row: 2, Col: 0},
			MaxPoints:    200,
			RequiredScore: 70,
			RequiredNodes: []string{"chain-of-thought", "few-shot-prompting"},
			Lessons:      []string{"advanced-roleplay", "advanced-constraints"},
		},
		{
			Slug:         "prompt-injection",
			Title:        "Security & Safety",
			Icon:         "🔒",
			Description:  "Understand and prevent prompt injection attacks",
			Position:     NodePosition{Row: 2, Col: 1},
			MaxPoints:    150,
			RequiredScore: 60,
			RequiredNodes: []string{"few-shot-prompting"},
			Lessons:      []string{"injection-basics", "injection-prevention"},
		},
		{
			Slug:         "prompt-optimization",
			Title:        "Prompt Optimization",
			Icon:         "⚡",
			Description:  "Iterate and refine prompts for maximum effectiveness",
			Position:     NodePosition{Row: 3, Col: 0},
			MaxPoints:    200,
			RequiredScore: 80,
			RequiredNodes: []string{"advanced-techniques"},
			Lessons:      []string{"optimization-techniques", "a-b-testing"},
		},
	},
}

// Example Tree: AI Concepts
const ExampleConceptsTree = TreeDefinition{
	Slug:        "ai-concepts",
	Title:       "AI Concepts & Logic",
	Description: "Understand how AI actually works under the hood",
	Icon:        "🧠",
	Category:    CategoryTechnical,
	Nodes: []NodeDefinition{
		{
			Slug:         "llm-basics",
			Title:        "How LLMs Work",
			Icon:         "🧠",
			Description:  "Tokens, embeddings, transformer architecture",
			Position:     NodePosition{Row: 0, Col: 0},
			MaxPoints:    150,
			RequiredScore: 0,
			Lessons:      []string{"llm-overview", "transformers-intro"},
		},
		{
			Slug:         "embeddings",
			Title:        "Embeddings & Vectors",
			Icon:         "📊",
			Description:  "How AI represents meaning as numbers",
			Position:     NodePosition{Row: 1, Col: 0},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"llm-basics"},
			Lessons:      []string{"embeddings-intro", "vector-similarity"},
		},
		{
			Slug:         "attention",
			Title:        "Attention Mechanism",
			Icon:         "👁️",
			Description:  "The revolutionary technique behind transformers",
			Position:     NodePosition{Row: 1, Col: 1},
			MaxPoints:    150,
			RequiredScore: 60,
			RequiredNodes: []string{"llm-basics"},
			Lessons:      []string{"attention-intro", "attention-visualization"},
		},
		{
			Slug:         "rag",
			Title:        "RAG Systems",
			Icon:         "📚",
			Description:  "Retrieval Augmented Generation for custom knowledge",
			Position:     NodePosition{Row: 2, Col: 0},
			MaxPoints:    200,
			RequiredScore: 70,
			RequiredNodes: []string{"embeddings"},
			Lessons:      []string{"rag-intro", "rag-implementations"},
		},
		{
			Slug:         "fine-tuning",
			Title:        "Fine-Tuning Models",
			Icon:         "🎛️",
			Description:  "Customize models for specific tasks",
			Position:     NodePosition{Row: 2, Col: 1},
			MaxPoints:    200,
			RequiredScore: 70,
			RequiredNodes: []string{"llm-basics"},
			Lessons:      []string{"finetuning-intro", "finetuning-practice"},
		},
		{
			Slug:         "multimodal",
			Title:        "Multimodal AI",
			Icon:         "🖼️",
			Description:  "AI that sees, hears, and speaks",
			Position:     NodePosition{Row: 3, Col: 0},
			MaxPoints:    150,
			RequiredScore: 75,
			RequiredNodes: []string{"rag", "fine-tuning"},
			Lessons:      []string{"multimodal-intro", "vision-models"},
		},
	},
}

// Example Tree: Models & Data
const ExampleModelsTree = TreeDefinition{
	Slug:        "models-data",
	Title:       "Models & Data",
	Description: "Work with AI models and data effectively",
	Icon:        "🗄️",
	Category:    CategoryTechnical,
	Nodes: []NodeDefinition{
		{
			Slug:         "model-selection",
			Title:        "Model Selection",
			Icon:         "🎯",
			Description:  "Choose the right model for the job",
			Position:     NodePosition{Row: 0, Col: 0},
			MaxPoints:    100,
			RequiredScore: 0,
			Lessons:      []string{"model-overview", "choosing-models"},
		},
		{
			Slug:         "data-prep",
			Title:        "Data Preparation",
			Icon:         "📊",
			Description:  "Clean and format data for AI consumption",
			Position:     NodePosition{Row: 1, Col: 0},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"model-selection"},
			Lessons:      []string{"data-cleaning", "data-formatting"},
		},
		{
			Slug:         "vector-dbs",
			Title:        "Vector Databases",
			Icon:         "🗄️",
			Description:  "Store and search embeddings efficiently",
			Position:     NodePosition{Row: 1, Col: 1},
			MaxPoints:    150,
			RequiredScore: 50,
			RequiredNodes: []string{"model-selection"},
			Lessons:      []string{"vectordb-intro", "vectordb-comparison"},
		},
		{
			Slug:         "api-integration",
			Title:        "API Integration",
			Icon:         "🔌",
			Description:  "Connect to AI APIs in production",
			Position:     NodePosition{Row: 2, Col: 0},
			MaxPoints:    200,
			RequiredScore: 70,
			RequiredNodes: []string{"data-prep"},
			Lessons:      []string{"api-design", "rate-limiting", "caching"},
		},
		{
			Slug:         "evaluation",
			Title:        "Model Evaluation",
			Icon:         "📈",
			Description:  "Measure model performance and accuracy",
			Position:     NodePosition{Row: 2, Col: 1},
			MaxPoints:    150,
			RequiredScore: 70,
			RequiredNodes: []string{"vector-dbs"},
			Lessons:      []string{"metrics-intro", "a-b-testing", "bias-detection"},
		},
		{
			Slug:         "deployment",
			Title:        "Deployment & Scaling",
			Icon:        "🚀",
			Description: "Put AI models into production",
			Position:     NodePosition{Row: 3, Col: 0},
			MaxPoints:    200,
			RequiredScore: 80,
			RequiredNodes: []string{"api-integration", "evaluation"},
			Lessons:      []string{"deployment-basics", "scaling-strategies"},
		},
	},
}

// ============================================================================
// DATA STRUCTURES FOR DEFINITIONS
// ============================================================================

// TreeDefinition defines a complete skill tree (used for seeding DB)
type TreeDefinition struct {
	Slug        string
	Title       string
	Description string
	Icon        string
	Category    SkillCategory
	Nodes       []NodeDefinition
}

// NodeDefinition defines a skill node
type NodeDefinition struct {
	Slug          string
	Title         string
	Icon          string
	Description   string
	Position      NodePosition
	MaxPoints     int
	RequiredScore float64
	RequiredNodes []string
	Lessons       []string // Lesson slugs
	Projects      []string // Project slugs
}

// All example trees for seeding
var ExampleTrees = []TreeDefinition{
	ExamplePromptTree,
	ExampleConceptsTree,
	ExampleModelsTree,
}

// ProficiencyDisplay returns the icon and label for a proficiency level
func ProficiencyDisplay(level Proficiency) (icon, label, color string) {
	_, icon, label, color = GetConfig(level)
	return
}
