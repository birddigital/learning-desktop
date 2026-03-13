# Learning Desktop - Master Plan

**Vision**: An AI-powered learning platform that teaches people how to not fall behind in the AI onslaught—through voice-first, personalized tutoring with video game-style progression.

**Repository**: https://github.com/birddigital/learning-desktop
**Language**: Go 1.25+ (Golang First Policy)
**License**: MIT

---

## Table of Contents

1. [Product Vision](#product-vision)
2. [Target Audience](#target-audience)
3. [Core Value Proposition](#core-value-proposition)
4. [Technical Architecture](#technical-architecture)
5. [Feature Roadmap](#feature-roadmap)
6. [Content Strategy](#content-strategy)
7. [Business Model](#business-model)
8. [Success Metrics](#success-metrics)

---

## Product Vision

### The Problem

AI is advancing faster than most people can adapt. Professionals, students, and entrepreneurs feel:
- **Overwhelmed** by the pace of AI advancement
- **Left behind** as AI automates skills they spent years developing
- **Paralyzed** by too much information and no clear learning path
- **Unsupported** by traditional education that moves too slowly

### The Solution

Learning Desktop is a **Claude Desktop-style AI tutor** that:

1. **Meets students where they are**—voice-first interface for natural conversation
2. **Speaks their truth**—culturally relevant content that resonates with diverse backgrounds
3. **Builds real skills**—not just theory, but practical AI capabilities
4. **Tracks everything**—progress, engagement, speech patterns, understanding
5. **Never lets them quit**—insight engine that tells the hard truth when they're falling behind

### The "Why"

> "The AI revolution isn't coming—it's already here. The question isn't whether AI will transform your industry, but whether you'll be the one using it or being replaced by it."

Learning Desktop exists to ensure no one gets left behind.

---

## Target Audience

### Primary: Professionals at Risk (25-45)

- **Knowledge workers** facing AI automation (writers, designers, analysts)
- **Skilled trades** seeing AI encroachment (programmers, engineers)
- **Entrepreneurs** needing AI to stay competitive

**Pain Points**: Limited time, fear of obsolescence, practical need over academic interest

### Secondary: Students (18-25)

- **College students** preparing for AI-augmented careers
- **Self-learners** seeking alternative to traditional education

**Pain Points**: Limited budget, need clear career path, want community

### Tertiary: Organizations (B2B)

- **Companies** needing to upskill workforces
- **Educational institutions** seeking AI curriculum
- **Government agencies** preparing workforce transitions

---

## Core Value Proposition

### 1. Voice-First Learning

- Speak naturally, AI tutor responds conversationally
- Speech analysis tracks confidence, understanding, engagement
- Lower barrier to entry—no typing required for mobile users

### 2. Culturally Relevant Content

- **Character & Manhood** curriculum with African-American cultural context
- Miami-specific examples (Liberty City, Little Haiti, North Miami)
- Stoic philosophy foundations (Marcus Aurelius, Epictetus)

### 3. Video Game Progression

- Skill trees unlock as you progress
- XP, levels, achievements, badges
- Social proof through leaderboards and certificates

### 4. The Hard Truth

- Insight Engine analyzes actual progress vs goals
- No false encouragement—data-driven accountability
- Predictive trajectory: "At this pace, you'll miss your goal by 23 days"

### 5. Multi-Tenant Architecture

- Organizations can deploy branded versions
- Row-Level Security for data isolation
- Custom skill trees, themes, domains

---

## Technical Architecture

### Tech Stack (Golang First)

| Layer | Technology | Rationale |
|-------|-----------|-----------|
| **Backend** | Go 1.25+ | Performance, concurrency, deployment simplicity |
| **Frontend** | HTMX + vanilla JS | No build step, progressive enhancement |
| **Database** | PostgreSQL 16+ | RLS for multi-tenancy, JSONB for flexibility |
| **Vector DB** | ChromaDB | Semantic search for RAG lesson generation |
| **Real-time** | Server-Sent Events | Simple unidirectional push |
| **Voice** | Web Speech API + server fallback | Browser-native first, OpenAI Whisper backup |
| **AI** | Claude API (Anthropic) | Best reasoning for tutoring |
| **Hosting** | Self-hosted Docker | Cost control, data privacy |

### Architecture Diagram

```
┌─────────────────────────────────────────────────────────────────┐
│                         CLIENT LAYER                            │
├─────────────────────────────────────────────────────────────────┤
│  Browser (HTMX) ──► Voice (Web Speech API) ──► SSE (Real-time)  │
└────────────────────────────┬────────────────────────────────────┘
                             │
┌────────────────────────────▼────────────────────────────────────┐
│                      APPLICATION LAYER                          │
├─────────────────────────────────────────────────────────────────┤
│  ┌─────────────┐  ┌──────────────┐  ┌─────────────────────┐   │
│  │ Chat Handler│  │Voice Service │  │  Insight Engine     │   │
│  └──────┬──────┘  └──────┬───────┘  └──────────┬──────────┘   │
│         │                │                     │                │
│  ┌──────▼────────────────▼─────────────────────▼───────────┐   │
│  │              Skill Tree System                           │   │
│  │  (Prompt Engineering | AI Concepts | Character | ...)  │   │
│  └──────────────────────────┬──────────────────────────────┘   │
└─────────────────────────────┼──────────────────────────────────┘
                              │
┌─────────────────────────────▼──────────────────────────────────┐
│                        DATA LAYER                              │
├─────────────────────────────────────────────────────────────────┤
│  ┌──────────────┐  ┌───────────────┐  ┌─────────────────┐     │
│  │ PostgreSQL   │  │  ChromaDB     │  │  File Storage   │     │
│  │ (Multi-tenant│  │  (Vectors)    │  │  (Research)     │     │
│  │  + RLS)      │  │               │  │                 │     │
│  └──────────────┘  └───────────────┘  └─────────────────┘     │
└─────────────────────────────────────────────────────────────────┘
                              │
┌─────────────────────────────▼──────────────────────────────────┐
│                     EXTERNAL SERVICES                           │
├─────────────────────────────────────────────────────────────────┤
│  Claude API │ Whisper API │ OpenAI │ Stripe │ SendGrid │ SES   │
└─────────────────────────────────────────────────────────────────┘
```

### Key Components

| Component | File | Purpose |
|-----------|------|---------|
| **Models** | `internal/models/models.go` | All data structures (587 lines) |
| **Repository** | `internal/repository/repository.go` | Database access layer |
| **Insight Engine** | `internal/insight/engine.go` | Progress analysis |
| **Skill Trees** | `internal/skill/*.go` | Content structure |
| **Chat Handler** | `internal/learning/chat-handler.go` | In htmx-r dependency |
| **Voice Service** | In htmx-r | STT/TTS/VAD |
| **Migrations** | `migrations/*.sql` | Database schema |

---

## Feature Roadmap

### Phase 1: Foundation (COMPLETE ✅)

- [x] Data models and database schema
- [x] Multi-tenant architecture with RLS
- [x] Skill tree system (6 trees, 45 nodes, 233 topics)
- [x] Insight Engine for progress analysis
- [x] Voice integration framework (via htmx-r)
- [x] Chat component foundation

### Phase 2: AI Tutor Integration (IN PROGRESS ⚠️)

- [ ] Claude API integration for tutoring responses
- [ ] Conversation context management
- [ ] Lesson generation from RAG
- [ ] ChromaDB vectorization of research content
- [ ] Semantic search for relevant content

### Phase 3: Progress Tracking (PLANNED 📋)

- [ ] Student session persistence
- [ ] Progress event storage
- [ ] Checkpoint/quiz submission handling
- [ ] Certificate generation
- [ ] Goal milestone tracking

### Phase 4: Multi-Tenancy (PLANNED 📋)

- [ ] Tenant onboarding flow
- [ ] Subdomain/custom domain support
- [ ] Per-tenant branding/themes
- [ ] Usage-based billing infrastructure
- [ ] Admin dashboard for organizations

### Phase 5: Advanced Features (FUTURE 🔮)

- [ ] Mobile apps (React Native)
- [ ] Offline mode with sync
- [ ] Peer learning communities
- [ ] Live tutoring (human + AI)
- [ ] Integration with external platforms (Notion, Obsidian)

---

## Content Strategy

### Skill Trees (COMPLETE)

All 233 topics across 6 skill trees have been researched and documented:

| Tree | Nodes | Topics | Focus |
|------|-------|--------|-------|
| **Prompt Engineering** | 6 | 30 | Practical AI prompting |
| **AI Concepts** | 6 | 30 | How LLMs work under the hood |
| **Models & Data** | 6 | 30 | Implementation skills |
| **Character & Manhood** | 9 | 45 | Personal development, cultural relevance |
| **Student Skills** | 9 | 45 | Learning how to learn |
| **Entrepreneurship** | 9 | 63 | AI-augmented business building |

### Content Format

Each topic includes:
```json
{
  "id": "unique-id",
  "title": "Topic Title",
  "content": "Full educational content",
  "examples": ["Example 1", "Example 2", "Example 3"],
  "assessment": "How to demonstrate mastery",
  "cultural_notes": "Cultural context (Character tree)",
  "difficulty": "beginner|intermediate|advanced",
  "resources": [{"title": "...", "url": "..."}]
}
```

### Content Location

Research content lives at `~/.learning-desktop/research/content/` and must be:
1. Vectorized into ChromaDB for semantic search
2. Used by RAG pipeline for lesson generation
3. Cached for performance

---

## Business Model

### Pricing Tiers

| Tier | Price | Features | Target |
|------|-------|----------|--------|
| **Free** | $0 | 1 skill tree, basic chat, community support | Individual learners |
| **Pro** | $29/mo | All skill trees, AI tutor, insights, certificates | Serious learners |
| **Enterprise** | Custom | White-label, SSO, custom content, dedicated support | Organizations |

### Revenue Drivers

1. **Subscriptions** - Monthly recurring revenue
2. **Enterprise licenses** - Per-seat pricing for organizations
3. **Certificates** - Paid credential verification
4. **Content marketplace** - Premium skill trees from creators

### Go-to-Market

1. **Phase 1**: Direct to consumer via content marketing (AI education focus)
2. **Phase 2**: Partner with educational institutions
3. **Phase 3**: Enterprise sales to companies needing workforce upskilling

---

## Success Metrics

### North Star Metric

**"Learning Velocity"**: Average skill nodes completed per student per month

### Leading Indicators

| Metric | Target | Why |
|--------|--------|-----|
| DAU/WAU ratio | >40% | Daily engagement habit |
| Lesson completion rate | >70% | Content quality |
| Voice interaction % | >50% | Voice adoption |
| Insight accuracy | >80% | Student validates predictions |

### Lagging Indicators

| Metric | Target | Why |
|--------|--------|-----|
| MRR (Monthly Recurring Revenue) | $10K by month 6 | Business viability |
| Cohort retention (90-day) | >60% | Product value |
| NPS (Net Promoter Score) | >50 | Word-of-mouth |
| Certificate issuance | >100/month | Skill verification |

---

## Development Principles

### Golang First

All new code must be written in Go unless:
1. There's a significant performance advantage to another language
2. Go cannot accomplish the task (e.g., MLX training requires Python)
3. An existing ecosystem has no Go equivalent

### No Stub Implementations

Every feature must be fully functional. No TODOs that break the user experience.

### Voice as Primary Interface

Design for voice first, add text as secondary input. Voice should work on:
- Chrome/Edge (full support)
- Safari (TTS only, STT via fallback)
- Firefox (server-side fallback)

### Multi-Tenant from Day One

All data access must respect tenant isolation via Row-Level Security.

### Progressive Enhancement

Core functionality works without JavaScript. JavaScript enhances the experience.

---

## Contributing

See `CONTRIBUTING.md` (to be created) for:
- Code style guidelines
- Pull request process
- Testing requirements
- Documentation standards

---

## License

MIT License - see `LICENSE` file for details.

---

*Last Updated: 2026-03-13*
*Status: Foundation Complete, AI Integration In Progress*
