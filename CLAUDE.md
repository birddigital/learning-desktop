# Learning Desktop - Claude Agent Instructions

**Purpose**: Project-specific instructions for Claude (Claude Code, Claude AI, MCP agents) working on the Learning Desktop codebase.

---

## Agent Context

You are working on **Learning Desktop**, an AI-powered learning platform built in Go that teaches people how to adapt to the AI revolution through voice-first, personalized tutoring with video game-style progression.

**Repository**: https://github.com/birddigital/learning-desktop
**Working Directory**: `/Users/birddigital/sources/standalone-projects/learning-desktop`

---

## Core Principles

### 1. Golang First Policy

All new code MUST be written in Go unless:
- Another language is significantly more performant for the specific task
- Go cannot accomplish the task (e.g., MLX training requires Python)
- An existing ecosystem/library has no Go equivalent

**Examples of Go-first choices:**
- ✅ CLI tools: Go
- ✅ Servers/APIs: Go
- ✅ Scripts/automation: Go
- ✅ Build systems: Go
- ❌ ML/AI training: Python (MLX, PyTorch - necessary exception)

### 2. No Stub Implementations

Every feature must be fully functional. No `TODO` placeholders that break the user experience. If you can't implement it fully, don't implement it at all.

### 3. Voice as Primary Interface

Design for voice first. Text input is secondary. Ensure voice works on:
- Chrome/Edge (full Web Speech API support)
- Safari (TTS only, STT via server fallback)
- Firefox (server-side fallback only)

### 4. Multi-Tenant from Day One

All data access MUST respect tenant isolation. Use Row-Level Security (RLS) in PostgreSQL queries.

### 5. Progressive Enhancement

Core functionality works without JavaScript. JavaScript enhances the experience.

---

## Architecture

### Directory Structure

```
learning-desktop/
├── cmd/
│   └── server/
│       └── main.go              # Server entry point
├── internal/
│   ├── models/
│   │   └── models.go            # ALL data models (587 lines, DO NOT split unnecessarily)
│   ├── repository/
│   │   └── repository.go        # Database access layer
│   ├── insight/
│   │   └── engine.go            # Progress analysis and recommendations
│   ├── skill/
│   │   ├── tree.go              # Skill tree definitions
│   │   ├── research_map.go      # Content mapping
│   │   ├── character.go         # Character & Manhood tree
│   │   ├── entrepreneurship.go  # Entrepreneurship tree
│   │   ├── student.go           # Student Skills tree
│   │   └── examples.go          # Prompt Engineering, AI Concepts, Models & Data
│   └── learning/
│       └── (empty, chat in htmx-r)
├── migrations/                   # SQL migrations
├── docs/                         # Technical documentation
├── static/                       # Served from htmx-r
├── CLAUDE.md                     # THIS FILE
├── MEMORY.md                     # Knowledge capture
├── STATUS.md                     # Current state
└── MASTER_PLAN.md                # Overall vision
```

### Dependencies

- **htmx-r**: Local dependency at `../htmx-r` - provides voice service and chat components
- **PostgreSQL**: Data storage with Row-Level Security
- **ChromaDB**: Vector search for RAG (to be integrated)
- **Claude service**: AI tutoring (to be integrated)

---

## Critical Files

### Files You MUST Read Before Making Changes

| File | Purpose | Lines |
|------|---------|-------|
| `internal/models/models.go` | ALL data models—don't duplicate | 587 |
| `internal/insight/engine.go` | Progress analysis logic | 649 |
| `cmd/server/main.go` | Server setup, endpoints | 224 |
| `STATUS.md` | Current project state | - |
| `MASTER_PLAN.md` | Overall vision | - |

### Files You Should NEVER Modify Without Discussion

- Go module version in `go.mod` (discuss dependency changes first)
- Database migrations that have been deployed (create new ones)
- Content structure in `internal/skill/` (affects 233 topics)

---

## Coding Standards

### Go Style

1. Follow standard Go formatting (`gofmt`)
2. Use meaningful variable names—abbreviations are acceptable only for common ones (`ctx`, `req`, `resp`)
3. Package comments at top of every file
4. Exported functions have godoc comments
5. Error handling: Never ignore errors, use wrapped errors with context

### Example

```go
// Package ai provides Claude service integration for tutoring.
package ai

import (
    "context"
    "fmt"
)

// Tutor manages AI tutoring conversations.
type Tutor struct {
    clientKey string
    client    *http.Client
}

// GenerateResponse creates a tutoring response for the student's question.
func (t *Tutor) GenerateResponse(ctx context.Context, sessionID string, message string) (string, error) {
    // Implementation
    return response, nil
}
```

### Database Queries

- Use `sqlx` for named queries
- Always include tenant_id in WHERE clauses (multi-tenancy)
- Use transactions for multi-step operations

```go
// GetStudentProgress retrieves progress for a specific student within their tenant.
func (r *Repository) GetStudentProgress(ctx context.Context, tenantID, studentID uuid.UUID) (*models.StudentProgress, error) {
    var progress models.StudentProgress
    query := `
        SELECT * FROM student_progress
        WHERE tenant_id = :tenant_id AND student_id = :student_id`
    args := map[string]interface{}{
        "tenant_id":  tenantID,
        "student_id": studentID,
    }
    rows, err := r.db.NamedQueryContext(ctx, query, args)
    // ... handle error
    return &progress, nil
}
```

---

## Feature Implementation Guidelines

### Adding New Features

1. **Read STATUS.md first** - Understand current state
2. **Check existing models** - Don't duplicate structures
3. **Follow the pattern** - Look at similar existing code
4. **Update STATUS.md** - Mark progress when done
5. **Commit with conventional commits** - `feat:`, `fix:`, `refactor:`

### Adding New Skill Trees

1. Define in `internal/skill/<newtree>.go`
2. Add to `internal/skill/tree.go` registry
3. Create content in `~/.learning-desktop/research/content/<newtree>/`
4. Update `internal/skill/RESEARCH_MAP.md`

### Adding New Database Fields

1. Create new migration in `migrations/`
2. Update `internal/models/models.go`
3. Update repository queries
4. Update STATUS.md

---

## Common Tasks

### Running the Server

```bash
cd /Users/birddigital/sources/standalone-projects/learning-desktop
go run cmd/server/main.go
```

### Running Migrations

```bash
cd /Users/birddigital/sources/standalone-projects/learning-desktop
go run cmd/migrate/main.go up
```

### Adding a Dependency

```bash
go get github.com/package/name
go mod tidy
# Update go.mod commit message with "deps:" prefix
```

### Creating a New API Endpoint

1. Add handler function to `cmd/server/main.go` (or extract to `internal/handlers/`)
2. Register in `mux` in main()
3. Add godoc comment
4. Update `docs/api.md` (create if doesn't exist)

---

## AI Integration Guidelines

### When to Use Claude Service

- **Use for**: Tutoring responses, lesson explanations, personalized feedback
- **Don't use for**: Simple queries, data transformation (use code)
- **Always**: Cache responses when appropriate
- **Always**: Include conversation context for continuity

### RAG Pipeline Pattern

```go
// 1. Vectorize query
embedding := chromaDB.Embed(query)

// 2. Semantic search for relevant content
results := chromaDB.Search(embedding, topK=5)

// 3. Build prompt with context
prompt := fmt.Sprintf("Context:\n%s\n\nQuestion: %s", results, query)

// 4. Generate response
response := claude.Generate(prompt)
```

---

## Testing

**Current State**: No tests exist yet

**Priority**: Add tests before production deployment

**Framework**: Use standard `testing` package, consider `testify` for assertions

```go
func TestInsightEngine_Generate(t *testing.T) {
    engine := insight.DefaultEngine()
    student := &models.Student{ /* ... */ }
    // ... test implementation
}
```

---

## Git Workflow

### Commit Message Format

```
<type>: <description>

[optional body]

[optional footer]
```

Types: `feat`, `fix`, `refactor`, `docs`, `test`, `chore`

### Branch Protection

- `main` is protected
- Create feature branches for significant work
- PRs required for merging (when team grows)

---

## Troubleshooting

### Common Issues

| Issue | Solution |
|-------|----------|
| htmx-r not found | Ensure `../htmx-r` exists or push to GitHub |
| Database connection fails | Check DATABASE_URL in .env |
| Claude API errors | Verify Anthropic API key is configured |
| Voice not working | Check browser compatibility (Chrome/Edge preferred) |

---

## Resources

### Internal Documentation

- `MASTER_PLAN.md` - Overall vision and roadmap
- `STATUS.md` - Current state and blockers
- `MEMORY.md` - Knowledge capture and insights
- `docs/database-schema.md` - Database design
- `docs/voice-integration.md` - Voice capabilities

### External Documentation

- [Go Documentation](https://golang.org/doc/)
- [HTMX Documentation](https://htmx.org/)
- [Claude Service Reference](https://docs.anthropic.com/)
- [ChromaDB Documentation](https://docs.trychroma.com/)

---

## Before You Leave

1. **Update STATUS.md** - Mark what you completed
2. **Update MEMORY.md** - Add any insights discovered
3. **Commit your work** - With clear commit messages
4. **Note blockers** - Add to STATUS.md if you hit issues

---

*Last Updated: 2026-03-13*
*This file supplements global CLAUDE.md instructions*
