# Learning Desktop - Knowledge Memory

**Purpose**: Persistent knowledge capture for the Learning Desktop project. Insights, patterns, anti-patterns, and lessons learned are stored here.

---

## Session History

| Session | Date | Focus | Outcome |
|---------|------|-------|---------|
| nightshift-1773361596 | 2026-03-13 | Project documentation setup | Created MASTER_PLAN.md, STATUS.md, CLAUDE.md, MEMORY.md |

---

## ★ Insights Discovered

### Architecture Decisions

#### Why htmx-r Local Dependency

**Date**: 2026-03-08
**Context**: Learning Desktop depends on `github.com/birddigital/htmx-r` via local replace directive.

**Pattern**: Use local replace directive during active development of shared packages.

**Rationale**:
- Faster iteration—no need to push, tag, update versions
- Immediate access to changes across both projects
- Prevents "dependency hell" during early development

**Anti-Pattern**: Committing with local replace still in place when deploying.

**Migration Path**: Before production:
1. Push htmx-r to its own repo
2. Tag a stable release (v0.1.0)
3. Update learning-desktop go.mod to use tagged version
4. Remove replace directive

---

#### Why Insight Engine Exists

**Date**: 2026-03-08
**Context**: Students need accountability, not just encouragement.

**Pattern**: "The Hard Truth" engine that predicts if students will hit their goals.

**Problem Solved**: Most learning platforms only show progress bars. They don't tell students:
- "At this pace, you'll miss your goal by 23 days"
- "You need 2.5 hours/day instead of 1 hour/day"
- "You're decelerating—this is a risk factor"

**Implementation**:
- Velocity calculation: progress points per day
- Required velocity: based on goal deadline
- Trajectory prediction: when will they actually finish?
- Recommendations: specific actions with time investment

**Evidence**: Insight engine fully implemented at `internal/insight/engine.go` (649 lines)

**Reusability**: This pattern can apply to any goal-tracking system (fitness, financial, etc.)

---

### Content Structure

#### Skill Tree System

**Date**: 2026-03-08
**Context**: Need to organize 233 topics across 6 domains.

**Pattern**: Video game-style skill trees with unlock progression.

**Structure**:
```
SkillTree (6 total)
├── Tree (e.g., "Prompt Engineering")
│   ├── Node (e.g., "Basic Prompts")
│   │   ├── Topic (5 topics per node)
│   │   ├── Topic
│   │   └── Topic
│   └── Node ...
```

**Key Insight**: 5 topics per node is the "magic number" for:
- Focused deep dives
- Manageable completion (achievable in 1-2 sessions)
- Clear assessment boundaries

**Anti-Pattern**: Nodes with 10+ topics become overwhelming and have lower completion rates.

---

#### Cultural Context in Content

**Date**: 2026-03-08
**Context**: Character & Manhood curriculum targets specific demographic.

**Pattern**: Cultural scaffolding makes content resonate.

**Implementation**:
- African-American cultural references throughout
- Miami-specific examples (Liberty City, Little Haiti, North Miami)
- Stoic philosophy foundations (Marcus Aurelius, Epictetus)
- Practical assessments for "testing out"

**Evidence**: Character tree has 9 nodes × 5 topics = 45 topics, all culturally contextualized.

**Reusability**: For other demographics, maintain the pattern but change the cultural anchors.

---

### Voice-First Design

#### Web Speech API Preference

**Date**: 2026-03-08
**Context**: Voice is primary interface, but browser support varies.

**Pattern**: Browser-native first, server fallback second.

**Browser Support Matrix**:
| Browser | STT | TTS | Fallback |
|---------|-----|-----|----------|
| Chrome/Edge | ✅ Native | ✅ Native | None needed |
| Safari | ❌ | ✅ Native | Whisper API |
| Firefox | ❌ | ❌ | Whisper + TTS API |

**Key Insight**: 85%+ of users have Chrome/Edge. Server fallback for edge cases is acceptable.

**Cost Optimization**: Browser-native = $0 per call. Server fallback = $0.006/minute (Whisper).

---

### Data Models

#### Why Single 587-Line Models File

**Date**: 2026-03-13
**Context**: All data structures in one file vs. split across many files.

**Pattern**: Co-locate related models until splitting becomes necessary.

**Rationale**:
- Easy to see all relationships at once
- Fewer files to navigate during early development
- Clearer dependency graph
- Split when file becomes unwieldy (>1000 lines) or teams need parallel edits

**Anti-Pattern**: Premature splitting creates circular dependencies and confusion.

---

#### Event Sourcing for Progress

**Date**: 2026-03-08
**Context**: Need to track student progress over time.

**Pattern**: ProgressEvent log as source of truth, materialized views for queries.

**Benefits**:
- Complete audit trail
- Can recalculate insights with new algorithms
- Time-travel queries ("what did progress look like on date X?")
- Debugging capability

**Implementation**:
- `ProgressEvent` table stores all events
- `StudentProgress` materialized view for current state
- `LessonBestScore` materialized view for checkpoint data

---

### Anti-Patterns Discovered

#### TODO Comments in Production Code

**Date**: 2026-03-13
**Issue**: `cmd/server/main.go:131` has `TODO: Generate AI response`

**Why It's Bad**:
- Placeholder implementation (echo response) breaks user experience
- Easy to forget about
- No clear owner or timeline

**Solution**: Either:
1. Implement it fully before committing
2. Create a tracking issue with clear acceptance criteria
3. Use a feature flag to hide incomplete functionality

---

## Open Questions

### Question: RAG Implementation Approach

**Date**: 2026-03-13
**Context**: Need to generate lessons from 233 researched topics.

**Options**:
1. **Pre-generate all lessons** - Fast responses, less flexible, higher storage
2. **Generate on-demand with RAG** - Slower, more flexible, lower storage
3. **Hybrid** - Cache common lessons, RAG for edge cases

**Recommendation**: Start with option 2 (RAG on-demand) to validate approach, then add caching.

**Decision Needed**: 2026-03-15

---

### Question: Authentication Provider

**Date**: 2026-03-13
**Context**: Need student login for multi-tenant isolation.

**Options**:
1. **Auth0** - Fast setup, monthly cost
2. **Supabase Auth** - Open source, built-in RLS
3. **Custom OAuth** - Most control, most maintenance

**Recommendation**: Supabase Auth (built-in PostgreSQL + RLS alignment)

**Decision Needed**: 2026-03-20

---

## Patterns to Reuse

### Insight Generation Pattern

```go
// Pattern: Analyze historical data to predict future outcomes
func (e *Engine) Generate(student *Student, progress *Progress, events []Event) *Insight {
    // 1. Calculate confidence based on data quality
    confidence := e.calculateConfidence(events)

    // 2. Build current state snapshot
    status := e.buildStatusSnapshot(student, progress, events)

    // 3. Analyze trajectory (are we on track?)
    trajectory := e.analyzeTrajectory(goal, status, events)

    // 4. Generate actionable recommendation
    recommendation := e.generateRecommendation(trajectory, status)

    // 5. Calculate supporting metrics
    metrics := e.calculateMetrics(student, progress, events)

    return &Insight{
        Confidence:     confidence,
        CurrentStatus:  status,
        Trajectory:    trajectory,
        Recommendation: recommendation,
        Metrics:       metrics,
    }
}
```

**Applicable to**: Fitness tracking, financial goals, project management, sales pipelines

---

### Multi-Tenant Query Pattern

```go
// Pattern: Always include tenant_id in WHERE clause
func (r *Repository) GetStudent(ctx context.Context, tenantID, studentID uuid.UUID) (*Student, error) {
    query := `
        SELECT * FROM students
        WHERE tenant_id = $1 AND id = $2`
    //                    ^^^^^^^^^^^^ CRITICAL
    return r.db.GetContext(ctx, student, query, tenantID, studentID)
}
```

**Security Note**: Even with RLS, always filter by tenant_id in application code. Defense in depth.

---

## Experiments to Run

### Experiment: Voice vs Text Engagement

**Hypothesis**: Voice-first interface increases daily engagement by 40%+ compared to text-only.

**Method**: A/B test with new users, track DAU/WAU ratio.

**Success Criteria**: Voice users have DAU/WAU >40%, text users <30%.

**Timeline**: After AI integration is complete (2026-03-20)

---

### Experiment: Insight Accuracy

**Hypothesis**: Insight Engine predictions are accurate within ±7 days 80% of the time.

**Method**: Track predicted completion date vs. actual completion date for 100 students.

**Success Criteria**: 80% of predictions within ±7 days.

**Timeline**: After 100 students complete courses (2026-04-30)

---

## Glossary

| Term | Definition |
|------|------------|
| **DAU/WAU** | Daily Active Users / Weekly Active Users - engagement metric |
| **RLS** | Row-Level Security - PostgreSQL feature for multi-tenancy |
| **RAG** | Retrieval-Augmented Generation - AI responses with context retrieval |
| **SSE** | Server-Sent Events - unidirectional push from server to client |
| **STT/TTS** | Speech-to-Text / Text-to-Speech |
| **VAD** | Voice Activity Detection - knowing when user is speaking |

---

## References

### Internal
- `MASTER_PLAN.md` - Overall vision
- `STATUS.md` - Current state
- `CLAUDE.md` - Agent instructions

### External
- [HTMX Best Practices](https://htmx.org/docs/#best-practices)
- [Go Project Layout](https://github.com/golang-standards/project-layout)
- [Multi-Tenant SaaS](https://www.multitenant-saas.guide/)

---

*This MEMORY.md is the third persistent layer alongside CLAUDE.md and STATUS.md*
*Update whenever you learn something worth remembering*
