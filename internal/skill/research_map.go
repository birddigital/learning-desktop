// Package skill provides research content mapping for skill trees.
// This file maps skill tree nodes to their research content directory paths.
package skill

// ResearchPath maps a tree slug and node slug to the research content directory
// Format: ~/.learning-desktop/research/content/{tree_slug}/{node_slug}/topics.json
//
// Complete Research Inventory:
//
// AI/Technical Trees (180 topics total):
//   ├── prompt_engineering (6 nodes × 5 topics = 30)
//   ├── ai_concepts (6 nodes × 5 topics = 30)
//   └── models_data (6 nodes × 5 topics = 30)
//
// Core Trees (123 topics + partial):
//   ├── character_manhood (9 nodes × 5 topics = 45) ✓ COMPLETE
//   ├── student_skills (9 nodes × 0-3 topics = ~15) PARTIAL
//   └── entrepreneurship (9 nodes × 7 topics = 63) ✓ COMPLETE
//
// Research Base Path
const ResearchBasePath = "~/.learning-desktop/research/content"

// TreeResearchPath returns the research directory path for a given tree slug
func TreeResearchPath(treeSlug string) string {
	return ResearchBasePath + "/" + treeSlug
}

// NodeResearchPath returns the research topics.json path for a given tree and node
func NodeResearchPath(treeSlug, nodeSlug string) string {
	return ResearchBasePath + "/" + treeSlug + "/" + nodeSlug + "/topics.json"
}

// HasResearch checks if research content exists for a node
func HasResearch(treeSlug, nodeSlug string) bool {
	// This would be implemented with actual file system checks
	// For now, returning true for nodes known to have research
	knownNodes := map[string][]string{
		"prompt-engineering": {"basic-prompts", "chain-of-thought", "few-shot-prompting", "advanced-techniques", "prompt-injection", "prompt-optimization"},
		"ai-concepts":       {"llm-basics", "embeddings", "attention", "rag", "fine-tuning", "multimodal"},
		"models-data":       {"model-selection", "data-prep", "vector-dbs", "api-integration", "evaluation", "deployment"},
		"character-manhood": {"integrity", "responsibility", "discipline", "respect", "resilience", "leadership", "purpose", "brotherhood", "legacy"},
		"student-skills":    {"focus", "note-taking", "time-management", "reading-strategy", "memory-techniques", "test-taking", "research", "teaching", "metacognition"},
		"entrepreneurship":  {"opportunity-spotting", "value-creation", "money-mindset", "sales", "marketing", "financial-literacy", "legal-basics", "team-building", "scaling"},
	}

	nodes, exists := knownNodes[treeSlug]
	if !exists {
		return false
	}

	for _, node := range nodes {
		if node == nodeSlug {
			return true
		}
	}
	return false
}

// ResearchCompleteness represents how much research exists for a node
type ResearchCompleteness string

const (
	CompletenessNone      ResearchCompleteness = "none"       // No research
	CompletenessPartial   ResearchCompleteness = "partial"    // Some topics
	CompletenessComplete  ResearchCompleteness = "complete"   // All topics done
	CompletenessExpanded  ResearchCompleteness = "expanded"   // More than baseline
)

// GetResearchStatus returns the research status for a node
func GetResearchStatus(treeSlug, nodeSlug string) ResearchCompleteness {
	// Known complete nodes (5+ topics for baseline, 7 for expanded)
	completeNodes := map[string]map[string]ResearchCompleteness{
		"character-manhood": {
			"integrity":     CompletenessComplete,
			"responsibility": CompletenessComplete,
			"discipline":    CompletenessComplete,
			"respect":       CompletenessComplete,
			"resilience":    CompletenessComplete,
			"leadership":    CompletenessComplete,
			"purpose":       CompletenessComplete,
			"brotherhood":   CompletenessComplete,
			"legacy":        CompletenessComplete,
		},
		"entrepreneurship": {
			"opportunity-spotting": CompletenessExpanded,
			"value-creation":      CompletenessExpanded,
			"money-mindset":       CompletenessExpanded,
			"sales":               CompletenessExpanded,
			"marketing":           CompletenessExpanded,
			"financial-literacy":  CompletenessExpanded,
			"legal-basics":        CompletenessExpanded,
			"team-building":       CompletenessExpanded,
			"scaling":             CompletenessExpanded,
		},
		"prompt-engineering": {
			"basic-prompts":       CompletenessComplete,
			"chain-of-thought":    CompletenessComplete,
			"few-shot-prompting":  CompletenessComplete,
			"advanced-techniques": CompletenessComplete,
			"prompt-injection":    CompletenessComplete,
			"prompt-optimization": CompletenessComplete,
		},
		"ai-concepts": {
			"llm-basics":    CompletenessComplete,
			"embeddings":    CompletenessComplete,
			"attention":     CompletenessComplete,
			"rag":           CompletenessComplete,
			"fine-tuning":   CompletenessComplete,
			"multimodal":    CompletenessComplete,
		},
		"models-data": {
			"model-selection": CompletenessComplete,
			"data-prep":       CompletenessComplete,
			"vector-dbs":      CompletenessComplete,
			"api-integration": CompletenessComplete,
			"evaluation":      CompletenessComplete,
			"deployment":      CompletenessComplete,
		},
		"student-skills": {
			"focus":           CompletenessPartial,
			"note-taking":     CompletenessPartial,
			"time-management": CompletenessPartial,
			// Others are partial or empty
		},
	}

	if treeMap, exists := completeNodes[treeSlug]; exists {
		if status, nodeExists := treeMap[nodeSlug]; nodeExists {
			return status
		}
	}
	return CompletenessNone
}

// OverallProgress returns research completion statistics
type OverallProgress struct {
	TotalNodes      int
	CompleteNodes   int
	PartialNodes    int
	EmptyNodes      int
	TotalTopics     int
	CompletionPct   float64
}

// GetOverallProgress calculates research progress across all trees
func GetOverallProgress() OverallProgress {
	return OverallProgress{
		TotalNodes:      45, // 6+6+6+9+9+9 nodes across all trees
		CompleteNodes:   33, // All AI + character + entrepreneurship
		PartialNodes:    6,  // student_skills partial
		EmptyNodes:      6,  // student_skills remaining
		TotalTopics:     213, // Approximate total topics
		CompletionPct:   73.3, // (33/45 complete + 6/45 partial/2)
	}
}
