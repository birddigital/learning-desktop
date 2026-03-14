# Learning Desktop - Evolutionary Architecture Analysis

**Analyzed**: 2026-03-14
**Analyst**: Evolutionary Architect (Claude Sonnet 4.5)
**Project Status**: Foundation Complete, AI Integration In Progress

---

## Executive Summary

Learning Desktop demonstrates **strong architectural foundations** with clean separation of concerns, comprehensive data modeling, and proper multi-tenant design. However, critical bottlenecks exist in session management, service layer organization, and scalability readiness that will impede production deployment.

**Overall Assessment**: 🟡 **Good Foundation, Critical Refactoring Needed**

| Category | Score | Status |
|----------|-------|--------|
| Data Modeling | 9/10 | Excellent |
| Repository Layer | 8/10 | Strong |
| Multi-Tenant Design | 9/10 | Excellent |
| Service Layer | 4/10 | Needs Work |
| Session Management | 3/10 | Critical Bottleneck |
| Scalability | 4/10 | Not Ready |
| Testing | 0/10 | Gap |
| Documentation | 7/10 | Good |

**Key Finding**: The 586-line models file and 585-line repository file are **well-organized monoliths** that should NOT be split prematurely. The real issues are in the missing service layer and in-memory session management.

---

## 1. Current Architecture Assessment

### 1.1 Strengths

#### ✅ Excellent Data Modeling (587 lines)
**File**: `internal/models/models.go`

**Why It Works**:
- Single file provides **complete visibility** into all data structures
- Logical grouping with clear section headers
- Proper use of enums (`TenantPlan`, `SkillLevel`, etc.)
- JSON and DB tags on all fields
- Comprehensive coverage: Tenant, Student, Session, Progress, Chat, Certificates

**Evidence**:
```go
// Clean enum pattern
type TenantPlan string
const (
    PlanFree        TenantPlan = "free"
    PlanPro         TenantPlan = "pro"
    PlanEnterprise  TenantPlan = "enterprise"
)

// Proper relationships
type Goal struct {
    // ...
    Milestones  []Milestone      `db:"-" json:"milestones,omitempty"`
    Insights    []InsightSnapshot `db:"-" json:"insights,omitempty"`
}
```

**Recommendation**: **DO NOT SPLIT**. The current monolith is readable and maintainable.

#### ✅ Repository Layer Excellence (585 lines)
**File**: `internal/repository/repository.go`

**Why It Works**:
- Clean separation by domain (Tenant, Student, Session, Course, Progress, Chat, AI, Checkpoint, Certificate)
- Proper context usage throughout
- Named queries with sqlx
- Multi-tenant awareness (tenant_id in queries)
- Transaction support via `InTx`

**Evidence**:
```go
// Proper multi-tenant query
func (r *StudentRepository) GetByTenantEmail(ctx context.Context, tenantID uuid.UUID, email string) (*models.Student, error) {
    var student models.Student
    query := `
        SELECT id, tenant_id, external_id, name, email, avatar_url,
            skill_level, interests, goals, background_text, industry, role, status,
            last_login_at, total_minutes, created_at, updated_at
        FROM students WHERE tenant_id = $1 AND email = $2`
    // ...
}
```

**Recommendation**: Keep as-is. Consider extracting complex queries (materialized view refresh) to separate package.

#### ✅ Multi-Tenant Architecture
**Design**: Row-Level Security (RLS) with PostgreSQL

**Why It Works**:
- Tenant context set at session level: `SetTenantContext()`
- Student context for additional granularity: `SetStudentContext()`
- All repository queries include tenant filtering
- Materialized views for aggregations

**Evidence**: Database schema uses RLS policies properly.

**Recommendation**: Add integration tests to verify RLS enforcement.

#### ✅ AI Integration Layer
**File**: `internal/ai/tutor.go`

**Why It Works**:
- Clean abstraction over go-llm-providers
- Environment-based configuration
- System prompt customization
- Conversation context support
- Lesson-specific tutoring

**Evidence**:
```go
func NewLessonTutor(lessonTitle, lessonContent string) (*LessonTutor, error) {
    baseTutor, err := New()
    // ...
    systemPrompt := fmt.Sprintf(`You are tutoring a lesson on "%s"...`, lessonTitle)
    baseTutor.SetSystemPrompt(systemPrompt)
    return &LessonTutor{Tutor: baseTutor, ...}
}
```

**Recommendation**: Add response caching to reduce API costs.

### 1.2 Weaknesses

#### ❌ Critical: In-Memory Session Management
**File**: `cmd/server/main.go` (lines 24-42)

**Problem**:
```go
type sessionStore struct {
    mu       sync.RWMutex
    sessions map[string]*chatSession  // IN-MEMORY ONLY
}
```

**Impact**:
- **Cannot scale horizontally** - sessions tied to single server
- **Lost on restart** - all conversations wiped
- **Memory leak risk** - no session expiration/cleanup
- **No persistence** - chat history not saved to database

**Why This Matters**:
According to STATUS.md, ChatRepository exists in `internal/repository/repository.go` with `CreateMessage()` and `GetMessagesBySession()`, but **it's never used**. The server uses in-memory sessions instead.

**Evidence of Problem**:
```bash
# Lines of code comparison
internal/repository/repository.go:  585 lines (ChatRepository: lines 402-441)
cmd/server/main.go:                 371 lines (sessionStore: lines 24-42)

# ChatRepository exists but is IGNORED
```

**Recommendation**: **P0 - Immediate Fix Required**

#### ❌ Missing Service Layer
**Problem**: Business logic scattered across handlers and repositories.

**Current State**:
```
cmd/server/main.go (371 lines)
├── HTTP handlers (handleChat, handleClear, etc.)
├── Session management (in-memory)
├── AI tutor calls
└── No business logic layer
```

**What's Missing**:
```
internal/service/  # DOESN'T EXIST
├── chat_service.go     # Chat orchestration
├── session_service.go  # Session lifecycle
├── progress_service.go # Progress tracking
└── tenant_service.go   # Tenant operations
```

**Impact**:
- Handlers contain business logic (violates SRP)
- No reusability across HTTP/CLI/API
- Difficult to test business logic
- Code duplication when adding gRPC/WebSocket

**Recommendation**: **P0 - Create Service Layer**

#### ❌ Tight Coupling with htmx-r
**Problem**: Static files and components served from local dependency.

**Evidence**:
```go
// cmd/server/main.go:70
mux.Handle("/static/", http.StripPrefix("/static/",
    http.FileServer(http.Dir("../htmx-r/static"))))

// go.mod:22
replace github.com/birddigital/htmx-r => ../htmx-r
```

**Impact**:
- **Cannot deploy independently** - requires sibling directory
- **Breaks in production** - no `../htmx-r` on servers
- **Version coupling** - always uses latest local code
- **CI/CD complexity** - need to manage both repos

**Recommendation**: **P1 - Decouple htmx-r**

#### ❌ No Testing Infrastructure
**Problem**: Zero test files exist.

**Evidence**:
```bash
$ find . -name "*_test.go" | wc -l
0
```

**Impact**:
- **Cannot verify refactoring safety**
- **No regression protection**
- **Difficult to onboard new developers**
- **Dangerous multi-tenant code** (RLS untested)

**Recommendation**: **P1 - Add Test Suite**

#### ❌ Configuration Management
**Problem**: Hardcoded values and scattered env var handling.

**Evidence**:
```go
// cmd/server/main.go:47
port := flag.String("port", "3000", "Server port")

// internal/ai/tutor.go:28
apiKey := os.Getenv("ANTHROPIC_CREDENTIALS")

// internal/db/db.go:30-34
Host:         getEnv("DB_HOST", "localhost"),
Port:         getEnv("DB_PORT", "5432"),
```

**Impact**:
- No centralized config
- No validation
- No config file support
- Difficult to manage environments

**Recommendation**: **P2 - Implement Config Package**

---

## 2. Refactoring Recommendations

### 2.1 Critical Refactoring (P0 - Blockers)

#### 🔥 Session Management Overhaul

**Current State**: In-memory sessions (lines 24-42 in main.go)

**Target Architecture**:
```
internal/service/session_service.go
├── CreateSession(studentID, tenantID) -> SessionID
├── GetSession(sessionID) -> *Session
├── AddMessage(sessionID, message) -> error
├── GetMessages(sessionID, limit) -> []Message
└── CleanupExpiredSessions() -> error

Uses:
├── ChatRepository (persist messages)
├── SessionRepository (persist session state)
└── Redis (cache active sessions)
```

**Implementation Plan**:

1. **Create Session Service** (`internal/service/session_service.go`):
```go
package service

type SessionService struct {
    chatRepo    *repository.ChatRepository
    sessionRepo *repository.SessionRepository
    redis       *redis.Client // Optional, for caching
}

func (s *SessionService) GetOrCreateSession(ctx context.Context, studentID uuid.UUID) (*models.StudentSession, error) {
    // 1. Check Redis cache
    // 2. If miss, check database
    // 3. If no active session, create new
    // 4. Return session with messages
}

func (s *SessionService) AddMessage(ctx context.Context, sessionID uuid.UUID, msg *models.ChatMessage) error {
    // 1. Persist to database via ChatRepository
    // 2. Update cache
    return nil
}
```

2. **Update Handler** (`cmd/server/main.go`):
```go
// Remove sessionStore (lines 24-42)
// Add sessionService

var sessionService *service.SessionService

func handleChat(w http.ResponseWriter, r *http.Request) {
    // Parse request
    sessionID := getOrCreateSessionID(r)
    studentID := getStudentIDFromAuth(r) // New: Auth middleware

    // Use service instead of in-memory
    session, err := sessionService.GetOrCreateSession(r.Context(), studentID)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Add user message
    userMsg := &models.ChatMessage{
        SessionID: sessionID,
        StudentID: studentID,
        Role:      models.RoleUser,
        Content:   message,
    }
    if err := sessionService.AddMessage(r.Context(), sessionID, userMsg); err != nil {
        // Handle error
    }

    // Generate AI response (same as before)
    // ...

    // Add assistant message
    assistantMsg := &models.ChatMessage{
        SessionID: sessionID,
        StudentID: studentID,
        Role:      models.RoleAssistant,
        Content:   responseContent,
    }
    _ = sessionService.AddMessage(r.Context(), sessionID, assistantMsg)

    // Render response (same as before)
}
```

3. **Benefits**:
- ✅ Horizontal scaling (sessions in DB)
- ✅ Persistence (survives restarts)
- ✅ Memory efficiency (inactive sessions not in RAM)
- ✅ History tracking (all messages stored)
- ✅ Analytics possible (query chat history)

**Estimated Effort**: 1-2 days

---

### 2.2 High Priority Refactoring (P1 - Scalability)

#### 📦 Extract Service Layer

**Problem**: Business logic in handlers.

**Solution**: Create service package.

**Structure**:
```
internal/service/
├── chat.go          # Chat orchestration
├── session.go       # Session lifecycle
├── progress.go      # Progress tracking
├── lesson.go        # Lesson delivery
├── tenant.go        # Tenant operations
└── middleware.go    # Auth, context, logging
```

**Example**: `internal/service/chat.go`
```go
package service

type ChatService struct {
    tutor     *ai.Tutor
    session   *SessionService
    insight   *insight.Engine
}

// SendMessage handles the complete chat flow
func (s *ChatService) SendMessage(ctx context.Context, req *ChatRequest) (*ChatResponse, error) {
    // 1. Validate request
    // 2. Get/create session
    // 3. Store user message
    // 4. Generate AI response
    // 5. Store assistant message
    // 6. Trigger insight analysis (async)
    // 7. Record AI usage
    // 8. Return response
}
```

**Estimated Effort**: 3-4 days

#### 🔒 Decouple htmx-r

**Problem**: Local replace dependency breaks deployment.

**Solutions** (choose one):

**Option A: Publish htmx-r to GitHub**
```bash
# In htmx-r/
git tag v0.1.0
git push origin v0.1.0

# In learning-desktop/go.mod
replace github.com/birddigital/htmx-r => ../htmx-r
# Becomes:
require github.com/birddigital/htmx-r v0.1.0
```

**Option B: Embed Static Assets**
```go
//go:embed static
var staticFS embed.FS

mux.Handle("/static/", http.FileServer(http.FS(staticFS)))
```

**Option C: Serve from CDN**
```go
// Upload to S3 during build
// Serve from CloudFront
mux.Handle("/static/", http.StripPrefix("/static/",
    http.FileServer(http.Dir("/var/www/static"))))
```

**Recommendation**: **Option A** (publish htmx-r) for flexibility.

**Estimated Effort**: 0.5 days (if publishing) or 1 day (if embedding)

#### 🧪 Add Testing Infrastructure

**Problem**: Zero test coverage.

**Solution**: Create test framework.

**Structure**:
```
internal/
├── chat/
│   ├── chat.go
│   └── chat_test.go
├── service/
│   ├── chat.go
│   └── chat_test.go
└── repository/
    ├── repository.go
    └── repository_test.go
```

**Example Test**:
```go
func TestChatService_SendMessage(t *testing.T) {
    // Setup
    db := setupTestDB(t)
    repo := repository.NewChatRepository(db)
    tutor := mockTutor{response: "Test response"}
    svc := service.NewChatService(repo, &tutor, nil)

    // Execute
    resp, err := svc.SendMessage(context.Background(), &service.ChatRequest{
        StudentID: testStudentID,
        Message:   "Hello",
    })

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, "Test response", resp.Content)
    assert.NotEqual(t, uuid.Nil, resp.MessageID)
}
```

**Target Coverage**: 70%+

**Estimated Effort**: 3-5 days

---

### 2.3 Medium Priority Improvements (P2 - Quality)

#### ⚙️ Configuration Package

**Problem**: Scattered env var handling.

**Solution**: Centralized config with validation.

**File**: `internal/config/config.go`
```go
package config

type Config struct {
    Server    ServerConfig
    Database  DatabaseConfig
    AI        AIConfig
    Features  FeatureFlags
}

type ServerConfig struct {
    Port         int    `env:"PORT" envDefault:"3000"`
    Environment  string `env:"ENVIRONMENT" envDefault:"development"`
    LogLevel     string `env:"LOG_LEVEL" envDefault:"info"`
}

type DatabaseConfig struct {
    URL          string `env:"DATABASE_URL" envRequired:"true"`
    MaxOpenConns int    `env:"DB_MAX_OPEN" envDefault:"25"`
    MaxIdleConns int    `env:"DB_MAX_IDLE" envDefault:"5"`
}

type AIConfig struct {
    AnthropicKey string `env:"ANTHROPIC_CREDENTIALS" envRequired:"true"`
    Model        string `env:"CLAUDE_MODEL" envDefault:"claude-3-5-sonnet-20241022"`
}

type FeatureFlags struct {
    EnableVoice      bool `env:"ENABLE_VOICE" envDefault:"true"`
    EnableRAG        bool `env:"ENABLE_RAG" envDefault:"true"`
    EnableCerticates bool `env:"ENABLE_CERTIFICATES" envDefault:"false"`
}

func Load() (*Config, error) {
    cfg := &Config{}
    if err := env.Parse(cfg); err != nil {
        return nil, err
    }
    return cfg, nil
}
```

**Estimated Effort**: 1 day

#### 📊 Structured Logging

**Problem**: No structured logging.

**Solution**: Use `logrus` or `zap`.

**File**: `internal/log/log.go`
```go
package log

import "go.uber.org/zap"

var Logger *zap.Logger

func Init(level string) error {
    config := zap.NewProductionConfig()
    config.Level = zap.NewAtomicLevelAt(getLevel(level))
    logger, err := config.Build()
    if err != nil {
        return err
    }
    Logger = logger
    return nil
}

func WithRequestID(id string) *zap.Logger {
    return Logger.With(zap.String("request_id", id))
}
```

**Estimated Effort**: 1 day

#### 🔄 Connection Pooling

**Problem**: Single database connection in `db.go`.

**Current Code**:
```go
// internal/db/db.go:62
db, err := sqlx.Connect("pgx", cfg.DataSource())
```

**Improvement**:
```go
// Already in db.go:68-70
db.SetMaxOpenConns(cfg.MaxOpenConns)  // Already good
db.SetMaxIdleConns(cfg.MaxIdleConns)  // Already good
db.SetConnMaxLifetime(cfg.MaxLifetime) // Already good
```

**Status**: ✅ Already implemented correctly!

**Additional**: Add health check for pool exhaustion.
```go
func (db *DB) PoolStats() sqlx.DBStats {
    return db.DB.Stats()
}

// In health endpoint
stats := db.PoolStats()
if stats.WaitDuration > 1*time.Second {
    // Log warning: connection pool exhausted
}
```

**Estimated Effort**: 0.5 days

---

## 3. New Components Needed

### 3.1 Authentication Layer

**Missing**: No authentication exists.

**Required Components**:
```
internal/auth/
├── provider.go        # OAuth integration
├── jwt.go             # Token management
├── middleware.go      # Auth middleware
└── tenant.go          # Tenant resolution
```

**Implementation**: `internal/auth/middleware.go`
```go
func AuthMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // 1. Extract JWT from cookie/header
        // 2. Validate token
        // 3. Get student ID from token
        // 4. Set tenant context
        // 5. Call next
    })
}
```

**Estimated Effort**: 2-3 days

### 3.2 RAG Pipeline

**Missing**: ChromaDB integration exists but no pipeline.

**Required Components**:
```
internal/rag/
├── pipeline.go        # Orchestration
├── vectorizer.go      # Embedding generation
├── retriever.go       # ChromaDB queries
└── generator.go       # Lesson generation
```

**Implementation**: `internal/rag/pipeline.go`
```go
type RAGPipeline struct {
    chroma    *chroma.Client
    embedding *openai.EmbeddingClient
    tutor     *ai.Tutor
}

func (r *RAGPipeline) GenerateLesson(ctx context.Context, topic string) (string, error) {
    // 1. Vectorize topic
    // 2. Search ChromaDB for relevant content
    // 3. Build prompt with context
    // 4. Generate lesson
    // 5. Cache result
}
```

**Estimated Effort**: 2-3 days

### 3.3 Progress Tracking Service

**Missing**: Insight engine exists but no service integration.

**Required Components**:
```
internal/service/progress.go
├── RecordEvent()      # Store progress event
├── GetInsights()      # Trigger insight engine
├── UpdateGoal()       # Update goal progress
└── CheckMilestones()  # Check milestone completion
```

**Estimated Effort**: 2 days

### 3.4 Background Job Worker

**Missing**: No async job processing.

**Use Cases**:
- Refresh materialized views
- Generate certificates
- Send reminder emails
- Analyze speech (async)
- Generate insights

**Solution**: Use `tally` or `river` (Go job queues).

**Estimated Effort**: 2 days

---

## 4. Scalability Considerations

### 4.1 Session Management (Current Bottleneck)

**Problem**: In-memory sessions prevent horizontal scaling.

**Solution**: Database-backed sessions with Redis cache.

**Architecture**:
```
┌─────────────────────────────────────────────────────────┐
│                     Load Balancer                        │
└────────────┬──────────────────────┬─────────────────────┘
             │                      │
        ┌────▼────┐            ┌────▼────┐
        │ Server 1│            │ Server 2│
        └────┬────┘            └────┬────┘
             │                      │
             └──────────┬───────────┘
                        │
                ┌───────▼────────┐
                │  PostgreSQL    │
                │  (Sessions)    │
                └────────────────┘
                        │
                ┌───────▼────────┐
                │  Redis         │
                │  (Cache)       │
                └────────────────┘
```

**Benefits**:
- ✅ Horizontal scaling
- ✅ Session persistence
- ✅ Fast lookups (Redis cache)
- ✅ Analytics possible

### 4.2 Database Connection Pooling

**Current Status**: ✅ Already configured correctly.

**Configuration**:
```go
MaxOpenConns: 25
MaxIdleConns: 5
MaxLifetime:  5 * time.Minute
```

**Recommendations**:
- Monitor pool exhaustion in health checks
- Use PgBouncer for high-concurrency scenarios
- Consider read replicas for analytics queries

### 4.3 Caching Strategy

**Current Status**: ❌ No caching.

**Recommendation**: Multi-layer caching.

**Layers**:
```
1. Response Cache (AI responses)
   - Key: content hash
   - TTL: 1 hour
   - Storage: Redis

2. Session Cache (active sessions)
   - Key: session_id
   - TTL: 24 hours
   - Storage: Redis

3. Query Cache (frequently accessed data)
   - Key: query hash
   - TTL: 5 minutes
   - Storage: Redis

4. Static Asset Cache (CDN)
   - CSS/JS/images
   - TTL: 1 year
   - Storage: CloudFront/S3
```

**Implementation**: `internal/cache/redis.go`
```go
type Cache struct {
    client *redis.Client
}

func (c *Cache) GetOrSet(ctx context.Context, key string, ttl time.Duration, fn func() (interface{}, error)) (interface{}, error) {
    // 1. Try cache
    // 2. If miss, call fn()
    // 3. Set cache
    // 4. Return result
}
```

**Estimated Effort**: 2 days

### 4.4 Horizontal Readiness

**Current State**: ❌ Not horizontally scalable.

**Blockers**:
1. In-memory sessions
2. No shared cache
3. Static assets on local filesystem
4. No distributed tracing

**Path to Horizontal Scaling**:

**Phase 1**: Fix Session Management (P0)
- Move sessions to database
- Add Redis cache
- Effort: 2 days

**Phase 2**: Externalize State (P1)
- Move static assets to S3/CDN
- Use Redis for all caching
- Effort: 1 day

**Phase 3**: Add Observability (P1)
- Structured logging with request IDs
- Distributed tracing (OpenTelemetry)
- Metrics (Prometheus)
- Effort: 2 days

**Phase 4**: Load Balancing (P2)
- Add health check endpoint
- Configure load balancer
- Test failover
- Effort: 1 day

**Total Effort**: 6 days

---

## 5. Specific Action Items

### 5.1 File-by-File Recommendations

#### `cmd/server/main.go` (371 lines)

**Issues**:
- ❌ In-memory sessionStore (lines 24-42)
- ❌ No middleware (auth, logging, recovery)
- ❌ Handler business logic (lines 152-226)
- ❌ No error recovery
- ❌ Hardcoded timeouts (lines 85-87)

**Refactoring Plan**:

1. **Extract Session Management**:
```go
// Remove lines 24-42 (sessionStore)
// Remove lines 173-218 (session logic from handleChat)
// Add:
var sessionService *service.SessionService
```

2. **Add Middleware Chain**:
```go
func main() {
    // Setup
    cfg := config.Load()
    logger := log.Init(cfg.LogLevel)
    db := db.OpenFromEnv()

    // Middleware chain
    mux := http.NewServeMux()
    handler := middleware.Chain(
        mux,
        middleware.Recovery(),
        middleware.Logging(logger),
        middleware.RequestID(),
        middleware.Auth(cfg.JWTSecret),
    )

    server := &http.Server{
        Handler:      handler,
        ReadTimeout:  cfg.Server.ReadTimeout,  // From config
        WriteTimeout: cfg.Server.WriteTimeout, // From config
    }
}
```

3. **Extract Handlers**:
```go
// cmd/server/handlers/chat.go
type ChatHandler struct {
    chat  *service.ChatService
    tutor *ai.Tutor
}

func (h *ChatHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // Handle chat
}

// Register routes
mux.Handle("/api/chat", &ChatHandler{chat: chatService, tutor: tutor})
```

**Estimated Refactoring Effort**: 2-3 days

#### `internal/models/models.go` (587 lines)

**Issues**: None - this is well-organized.

**Recommendation**: **DO NOT SPLIT**

**Optional Enhancement**:
```go
// Add validation methods
func (s *Student) Validate() error {
    if s.Email == "" {
        return errors.New("email required")
    }
    if _, err := mail.ParseAddress(s.Email); err != nil {
        return fmt.Errorf("invalid email: %w", err)
    }
    return nil
}

// Add helper methods
func (p *StudentProgress) IsOnTrack() bool {
    return p.CompletionPercent() > 50 && time.Since(*p.LastActivityAt) < 7*24*time.Hour
}
```

**Estimated Effort**: 0.5 days

#### `internal/repository/repository.go` (585 lines)

**Issues**:
- ⚠️ Async materialized view refresh (line 369) - fire-and-forget
- ⚠️ No transaction usage examples
- ⚠️ No query logging

**Refactoring Plan**:

1. **Fix Materialized View Refresh**:
```go
// Current (line 369):
go r.refreshStudentProgress(event.StudentID)

// Problem: No error handling, no retry

// Solution:
func (r *ProgressRepository) refreshStudentProgress(studentID uuid.UUID) {
    ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer cancel()

    // Use background job queue instead
    jobs.Enqueue("refresh_progress", studentID)
}

// Or add retry:
func (r *ProgressRepository) refreshStudentProgress(studentID uuid.UUID) {
    const maxRetries = 3
    for i := 0; i < maxRetries; i++ {
        ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
        _, err := r.db.ExecContext(ctx, "REFRESH MATERIALIZED VIEW CONCURRENTLY student_progress")
        cancel()

        if err == nil {
            return
        }
        log.Printf("Refresh attempt %d failed: %v", i+1, err)
        time.Sleep(time.Second * time.Duration(i+1))
    }
    log.Printf("All refresh attempts failed for student %s", studentID)
}
```

2. **Add Transaction Example**:
```go
// Example usage in service layer
func (s *Service) CreateLesson(ctx context.Context, lesson *models.CourseLesson) error {
    return s.db.InTx(ctx, func(tx *sqlx.Tx) error {
        // 1. Create lesson
        if err := s.lessonRepo.CreateTx(ctx, tx, lesson); err != nil {
            return err
        }
        // 2. Update module
        if err := s.moduleRepo.AddLessonTx(ctx, tx, lesson.ModuleID, lesson.ID); err != nil {
            return err
        }
        return nil
    })
}
```

3. **Add Query Logging**:
```go
func (r *StudentRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Student, error) {
    start := time.Now()
    student, err := r.getByID(ctx, id)
    log.Printf("Query GetByID took %v", time.Since(start))
    return student, err
}
```

**Estimated Refactoring Effort**: 1 day

#### `internal/ai/tutor.go` (236 lines)

**Issues**:
- ⚠️ No response caching
- ⚠️ No rate limiting
- ⚠️ No streaming in handlers

**Refactoring Plan**:

1. **Add Caching**:
```go
type Tutor struct {
    client     *claude.Client
    model      string
    systemPrompt string
    cache      *cache.Cache  // Add cache
}

func (t *Tutor) Respond(ctx context.Context, question string) (string, error) {
    // Check cache
    cacheKey := fmt.Sprintf("tutor:%s:%x", t.model, sha256.Sum256([]byte(question)))
    if cached, err := t.cache.Get(ctx, cacheKey); err == nil {
        return cached, nil
    }

    // Generate response
    resp, err := t.respondUncached(ctx, question)
    if err != nil {
        return "", err
    }

    // Cache for 1 hour
    _ = t.cache.Set(ctx, cacheKey, resp, 1*time.Hour)
    return resp, nil
}
```

2. **Add Rate Limiting**:
```go
type Tutor struct {
    // ...
    rateLimiter *rate.Limiter
}

func New() (*Tutor, error) {
    // ...
    return &Tutor{
        // ...
        rateLimiter: rate.NewLimiter(rate.Limit(10), 20), // 10 req/sec, burst 20
    }
}

func (t *Tutor) Respond(ctx context.Context, question string) (string, error) {
    if err := t.rateLimiter.Wait(ctx); err != nil {
        return "", fmt.Errorf("rate limit: %w", err)
    }
    // ... rest of implementation
}
```

**Estimated Refactoring Effort**: 1 day

#### `internal/insight/engine.go` (649 lines)

**Issues**: None - well-designed analysis engine.

**Recommendation**: Keep as-is.

**Optional Enhancement**:
```go
// Add streaming insights
func (e *Engine) StreamInsights(ctx context.Context, studentID uuid.UUID) (<-chan *Insight, error) {
    ch := make(chan *Insight)
    go func() {
        defer close(ch)
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for {
            select {
            case <-ctx.Done():
                return
            case <-ticker.C:
                insight, err := e.Generate(ctx, studentID)
                if err != nil {
                    continue
                }
                ch <- insight
            }
        }
    }()
    return ch, nil
}
```

**Estimated Effort**: 1 day (optional)

### 5.2 Prioritized Task List

#### Phase 1: Critical Blockers (P0) - MUST DO

| Task | File | Effort | Impact |
|------|------|--------|--------|
| 1. Implement Session Service | `internal/service/session.go` | 1 day | Unblocks scaling |
| 2. Replace in-memory sessions | `cmd/server/main.go` | 1 day | Enables persistence |
| 3. Add ChatRepository usage | `cmd/server/main.go` | 0.5 days | Enables history |
| 4. Add Auth Middleware | `internal/auth/middleware.go` | 1 day | Security requirement |
| 5. Wire Database Connection | `cmd/server/main.go` | 0.5 days | Enable persistence |

**Total**: 4 days

#### Phase 2: Scalability (P1) - SHOULD DO

| Task | File | Effort | Impact |
|------|------|--------|--------|
| 6. Extract Service Layer | `internal/service/*.go` | 3 days | Maintainability |
| 7. Decouple htmx-r | `go.mod`, `cmd/server/main.go` | 1 day | Deployability |
| 8. Add Test Suite | `**/*_test.go` | 4 days | Confidence |
| 9. Add Caching Layer | `internal/cache/redis.go` | 2 days | Performance |
| 10. Add Structured Logging | `internal/log/log.go` | 1 day | Observability |

**Total**: 11 days

#### Phase 3: Quality of Life (P2) - NICE TO HAVE

| Task | File | Effort | Impact |
|------|------|--------|--------|
| 11. Implement Config Package | `internal/config/config.go` | 1 day | DX improvement |
| 12. Add Response Caching | `internal/ai/tutor.go` | 1 day | Cost reduction |
| 13. Add Background Jobs | `internal/jobs/worker.go` | 2 days | Async capability |
| 14. Implement RAG Pipeline | `internal/rag/pipeline.go` | 3 days | AI enhancement |
| 15. Add Metrics | `internal/metrics/prometheus.go` | 1 day | Monitoring |

**Total**: 8 days

**Grand Total**: 23 days (4.6 weeks)

---

## 6. Architecture Evolution Roadmap

### Current State (March 2026)

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Client                          │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────▼─────────────┐
        │  cmd/server/main.go      │
        │  ├─ HTTP Handlers        │
        │  ├─ In-Memory Sessions   │  ❌ BLOCKER
        │  └─ AI Tutor Calls       │
        └────────────┬─────────────┘
                     │
        ┌────────────▼─────────────────────────────────┐
        │  Repository Layer (585 lines)               │
        │  ├─ TenantRepository                       │
        │  ├─ StudentRepository                      │
        │  ├─ SessionRepository                      │  ⚠️ UNUSED
        │  ├─ ChatRepository                         │  ⚠️ UNUSED
        │  └─ ProgressRepository                     │
        └────────────┬─────────────────────────────────┘
                     │
        ┌────────────▼─────────────┐
        │  PostgreSQL              │
        │  (Row-Level Security)    │
        └──────────────────────────┘
```

### Target State (April 2026)

```
┌─────────────────────────────────────────────────────────┐
│                    HTTP Client                          │
└────────────────────┬────────────────────────────────────┘
                     │
        ┌────────────▼──────────────────────────────────┐
        │  Middleware Chain                             │
        │  ├─ Recovery                                  │
        │  ├─ Request ID                                │
        │  ├─ Logging                                   │
        │  └─ Auth/JWT                                  │
        └────────────┬──────────────────────────────────┘
                     │
        ┌────────────▼──────────────────────────────────┐
        │  HTTP Handlers (cmd/server/handlers/)        │
        │  ├─ ChatHandler                              │
        │  ├─ LessonHandler                            │
        │  └─ ProgressHandler                          │
        └────────────┬──────────────────────────────────┘
                     │
        ┌────────────▼──────────────────────────────────┐
        │  Service Layer (internal/service/)           │
        │  ├─ ChatService (orchestration)              │
        │  ├─ SessionService (lifecycle)               │
        │  ├─ ProgressService (tracking)               │
        │  └─ TenantService (operations)               │
        └────────────┬──────────────────────────────────┘
                     │
        ┌────────────▼──────────────────────────────────┐
        │  Repository Layer + Cache                    │
        │  ├─ Repositories (DB access)                 │
        │  └─ Redis Cache (fast lookups)               │
        └────────────┬──────────────────────────────────┘
                     │
        ┌────────────▼──────────────────────────────────┐
        │  External Services                            │
        │  ├─ PostgreSQL (data)                        │
        │  ├─ ChromaDB (vectors)                       │
        │  ├─ Claude API (AI)                          │
        │  └─ Redis (cache)                            │
        └───────────────────────────────────────────────┘
```

### Migration Path

**Week 1-2**: Fix Critical Blockers
- Implement Session Service
- Replace in-memory sessions
- Add database connection
- Add auth middleware

**Week 3-5**: Build Service Layer
- Extract business logic
- Implement caching
- Add structured logging
- Write tests

**Week 6-8**: Enhance Platform
- Implement RAG pipeline
- Add background jobs
- Decouple htmx-r
- Add monitoring

**Week 9+**: Production Hardening
- Load testing
- Security audit
- Performance tuning
- Documentation

---

## 7. Integration Improvements

### 7.1 htmx-r Coupling

**Current Problem**:
```go
// cmd/server/main.go:70
mux.Handle("/static/", http.StripPrefix("/static/",
    http.FileServer(http.Dir("../htmx-r/static"))))

// go.mod:22
replace github.com/birddigital/htmx-r => ../htmx-r
```

**Solution Options**:

**Option A: Publish htmx-r** (Recommended)
```bash
# In htmx-r/
git tag v0.1.0
git push origin v0.1.0
git push origin main

# In learning-desktop/
go get github.com/birddigital/htmx-r@v0.1.0
# Remove replace directive
```

**Benefits**:
- ✅ Version pinning
- ✅ Independent deployment
- ✅ CI/CD friendly
- ✅ Semantic versioning

**Option B: Vendor Assets**
```bash
# During build
cp -r ../htmx-r/static ./static
cp -r ../htmx-r/components ./components

# In code
mux.Handle("/static/", http.FileServer(http.Dir("./static")))
```

**Benefits**:
- ✅ No external dependency
- ✅ Customization possible
- ❌ Duplication
- ❌ Maintenance burden

**Option C: Module Proxy**
```go
// Use go-htmx-r or similar alternative
import "github.com/davidslab/htmx-r"
```

**Benefits**:
- ✅ No local dependency
- ❌ Loss of customizations
- ❌ Migration effort

**Recommendation**: **Option A** - Publish htmx-r to GitHub.

### 7.2 go-llm-providers Integration

**Current State**: ✅ Good abstraction
```go
// internal/ai/tutor.go:11-12
import (
    claude "github.com/birddigital/go-llm-providers/pkg/claude"
    "github.com/birddigital/go-llm-providers/pkg/providers"
)
```

**Enhancement**: Add fallback providers
```go
type Tutor struct {
    primary   *claude.Client
    fallback  *openai.Client  // Add OpenAI fallback
    cache     *cache.Cache
}

func (t *Tutor) Respond(ctx context.Context, question string) (string, error) {
    resp, err := t.primary.Complete(ctx, req)
    if err != nil {
        log.Printf("Primary provider failed: %v", err)
        return t.fallback.Complete(ctx, req)
    }
    return resp, nil
}
```

### 7.3 ChromaDB Integration

**Current State**: Client exists but unused.
```go
// internal/chroma/client.go exists
// But no RAG pipeline implemented
```

**Implementation**: `internal/rag/retriever.go`
```go
type Retriever struct {
    client *chroma.Client
    collection string
}

func (r *Retriever) Search(ctx context.Context, query string, topK int) ([]string, error) {
    // 1. Embed query
    embedding, err := r.embed(ctx, query)
    if err != nil {
        return nil, err
    }

    // 2. Search ChromaDB
    results, err := r.client.Query(
        ctx,
        r.collection,
        embedding,
        topK,
    )
    if err != nil {
        return nil, err
    }

    // 3. Return content
    return extractContent(results), nil
}
```

---

## 8. Code Examples

### 8.1 Proper Service Layer Implementation

**Before** (current `cmd/server/main.go`):
```go
func handleChat(w http.ResponseWriter, r *http.Request) {
    // Parse form
    message := r.FormValue("message")

    // Get session (in-memory)
    session := getOrCreateSession(sessionID)

    // Add user message
    session.Messages = append(session.Messages, userMsg)

    // Generate AI response
    resp, err := tutor.RespondWithConversation(ctx, session.Messages)

    // Add assistant message
    session.Messages = append(session.Messages, assistantMsg)

    // Render
    components.RenderMessageHTML(w, chatMessage)
}
```

**After** (proper service layer):
```go
// internal/service/chat.go
func (s *ChatService) SendMessage(ctx context.Context, req *SendMessageRequest) (*SendMessageResponse, error) {
    // 1. Validate
    if err := req.Validate(); err != nil {
        return nil, fmt.Errorf("validation: %w", err)
    }

    // 2. Get session
    session, err := s.session.GetOrCreateSession(ctx, req.StudentID)
    if err != nil {
        return nil, fmt.Errorf("get session: %w", err)
    }

    // 3. Store user message
    userMsg := &models.ChatMessage{
        SessionID: session.ID,
        StudentID: req.StudentID,
        TenantID:  req.TenantID,
        Role:      models.RoleUser,
        Content:   req.Message,
    }
    if err := s.chatRepo.CreateMessage(ctx, userMsg); err != nil {
        return nil, fmt.Errorf("store user message: %w", err)
    }

    // 4. Generate AI response
    messages, err := s.chatRepo.GetMessagesBySession(ctx, session.ID, 10)
    if err != nil {
        return nil, fmt.Errorf("get history: %w", err)
    }

    content, err := s.tutor.RespondWithConversation(ctx, toProviderMessages(messages))
    if err != nil {
        return nil, fmt.Errorf("AI response: %w", err)
    }

    // 5. Store assistant message
    assistantMsg := &models.ChatMessage{
        SessionID: session.ID,
        StudentID: req.StudentID,
        TenantID:  req.TenantID,
        Role:      models.RoleAssistant,
        Content:   content,
    }
    if err := s.chatRepo.CreateMessage(ctx, assistantMsg); err != nil {
        return nil, fmt.Errorf("store assistant message: %w", err)
    }

    // 6. Record AI usage (async)
    go s.recordAIUsage(context.Background(), req.StudentID, req.TenantID, content)

    // 7. Trigger insight analysis (async)
    go s.analyzeProgress(context.Background(), req.StudentID)

    return &SendMessageResponse{
        MessageID:    assistantMsg.ID,
        Content:      content,
        SessionID:    session.ID,
    }, nil
}
```

### 8.2 Proper Middleware Chain

**File**: `internal/middleware/chain.go`
```go
func Chain(handler http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        handler = middlewares[i](handler)
    }
    return handler
}

// Usage in main.go
mux := http.NewServeMux()

handler := middleware.Chain(
    mux,
    middleware.Recovery(),
    middleware.RequestID(),
    middleware.Logging(log.Logger),
    middleware.CORS(),
    middleware.Auth(cfg.JWTSecret),
    middleware.TenantContext(),
)

server := &http.Server{
    Handler: handler,
    // ...
}
```

### 8.3 Proper Error Handling

**File**: `internal/errors/errors.go`
```go
package errors

import (
    "errors"
    "fmt"
    "net/http"
)

// AppError represents an application error
type AppError struct {
    Code       int    `json:"code"`
    Message    string `json:"message"`
    Internal   error  `json:"-"`
    StackTrace string `json:"-"` // In development only
}

func (e *AppError) Error() string {
    return e.Message
}

func (e *AppError) Unwrap() error {
    return e.Internal
}

// Common errors
var (
    ErrNotFound     = &AppError{Code: http.StatusNotFound, Message: "resource not found"}
    ErrUnauthorized = &AppError{Code: http.StatusUnauthorized, Message: "unauthorized"}
    ErrValidation   = &AppError{Code: http.StatusBadRequest, Message: "validation failed"}
)

// New creates a new application error
func New(code int, message string, internal error) *AppError {
    return &AppError{
        Code:     code,
        Message:  message,
        Internal: internal,
    }
}

// Wrap wraps an error with context
func Wrap(err error, message string) *AppError {
    return &AppError{
        Code:     http.StatusInternalServerError,
        Message:  message,
        Internal: err,
    }
}

// HTTP handler
func WriteError(w http.ResponseWriter, err error) {
    var appErr *AppError
    if errors.As(err, &appErr) {
        w.Header().Set("Content-Type", "application/json")
        w.WriteHeader(appErr.Code)
        json.NewEncoder(w).Encode(map[string]interface{}{
            "error": appErr.Message,
        })
        return
    }

    // Unknown error
    w.WriteHeader(http.StatusInternalServerError)
    json.NewEncoder(w).Encode(map[string]interface{}{
        "error": "internal server error",
    })
}
```

---

## 9. Performance Optimizations

### 9.1 Database Query Optimization

**Current**: Individual queries in loops.

**Problem**: N+1 query problem.

**Solution**: Batch queries and joins.

**Before**:
```go
func (r *CourseRepository) GetWithModules(ctx context.Context, slug string) (*models.Course, error) {
    course, err := r.GetBySlug(ctx, slug)
    // ...
    for _, module := range course.Modules {
        lessons, _ := r.GetLessonsForModule(ctx, module.ID)  // N queries
        module.Lessons = lessons
    }
}
```

**After**:
```go
func (r *CourseRepository) GetWithModulesAndLessons(ctx context.Context, slug string) (*models.Course, error) {
    // Single query with JOIN
    query := `
        SELECT
            c.*,
            m.id as module_id,
            m.title as module_title,
            l.id as lesson_id,
            l.title as lesson_title
        FROM courses c
        LEFT JOIN course_modules m ON m.course_id = c.id
        LEFT JOIN course_lessons l ON l.module_id = m.id
        WHERE c.slug = $1
        ORDER BY m.order_index, l.order_index
    `
    // Parse nested structure
    return parseNestedResults(rows)
}
```

### 9.2 Response Streaming

**Current**: Blocking AI responses.

**Problem**: 25-second timeout feels slow.

**Solution**: Stream responses via SSE.

**Implementation**:
```go
func (h *ChatHandler) StreamChat(w http.ResponseWriter, r *http.Request) {
    flusher, ok := w.(http.Flusher)
    if !ok {
        http.Error(w, "Streaming not supported", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "text/event-stream")
    w.Header().Set("Cache-Control", "no-cache")
    w.Header().Set("Connection", "keep-alive")

    // Stream AI response
    stream, err := h.tutor.StreamRespond(ctx, message)
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }

    // Send chunks
    for chunk := range stream {
        fmt.Fprintf(w, "event: message\ndata: %s\n\n", chunk.Content)
        flusher.Flush()
    }

    fmt.Fprintf(w, "event: done\ndata: {}\n\n")
    flusher.Flush()
}
```

### 9.3 Connection Pool Tuning

**Current**: Default pool settings.

**Monitoring**:
```go
func (db *DB) PoolStats() sqlx.DBStats {
    return db.DB.Stats()
}

// In health endpoint
stats := db.PoolStats()
if stats.WaitCount > 1000 {
    log.Warn("Connection pool exhaustion detected")
    log.Warnf("WaitDuration: %v", stats.WaitDuration)
}
```

**Tuning**:
```go
// High-concurrency scenario
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(25)
db.SetConnMaxLifetime(1 * time.Hour)

// Low-concurrency scenario
db.SetMaxOpenConns(25)
db.SetMaxIdleConns(5)
db.SetConnMaxLifetime(5 * time.Minute)
```

---

## 10. Security Considerations

### 10.1 Multi-Tenant Isolation

**Current**: RLS policies defined but not enforced.

**Risk**: Tenant data leak.

**Solution**: Always set tenant context.

**Implementation**:
```go
func SetTenantContext(ctx context.Context, db *sqlx.DB, tenantID uuid.UUID) error {
    _, err := db.ExecContext(ctx, "SELECT set_tenant_context($1)", tenantID)
    return err
}

// In every request
func TenantMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        tenantID := getTenantIDFromToken(r)
        if err := SetTenantContext(r.Context(), db, tenantID); err != nil {
            http.Error(w, "Failed to set tenant context", http.StatusInternalServerError)
            return
        }
        next.ServeHTTP(w, r)
    })
}
```

### 10.2 Input Validation

**Current**: Minimal validation.

**Risk**: SQL injection, XSS, etc.

**Solution**: Validate all inputs.

**Implementation**:
```go
func (req *SendMessageRequest) Validate() error {
    if req.Message == "" {
        return errors.New("message required")
    }
    if len(req.Message) > 10000 {
        return errors.New("message too long")
    }
    if req.StudentID == uuid.Nil {
        return errors.New("student_id required")
    }
    if req.TenantID == uuid.Nil {
        return errors.New("tenant_id required")
    }
    return nil
}
```

### 10.3 Rate Limiting

**Current**: No rate limiting.

**Risk**: DoS, API cost explosion.

**Solution**: Per-tenant rate limiting.

**Implementation**:
```go
import "golang.org/x/time/rate"

type RateLimiter struct {
    tenants map[uuid.UUID]*rate.Limiter
    mu      sync.RWMutex
}

func (r *RateLimiter) Allow(tenantID uuid.UUID) bool {
    r.mu.RLock()
    limiter, exists := r.tenants[tenantID]
    r.mu.RUnlock()

    if !exists {
        r.mu.Lock()
        limiter = rate.NewLimiter(rate.Limit(10), 20) // 10 req/sec
        r.tenants[tenantID] = limiter
        r.mu.Unlock()
    }

    return limiter.Allow()
}
```

---

## Summary and Next Steps

### Critical Actions (This Week)

1. **Implement Session Service** - Replace in-memory sessions
2. **Wire Database Connection** - Enable persistence
3. **Add Auth Middleware** - Secure endpoints
4. **Decouple htmx-r** - Enable deployment

### High Priority (Next 2 Weeks)

5. **Extract Service Layer** - Improve maintainability
6. **Add Test Suite** - Ensure correctness
7. **Implement Caching** - Reduce costs
8. **Add Logging** - Enable debugging

### Medium Priority (Next Month)

9. **Config Package** - Simplify deployment
10. **RAG Pipeline** - Enhance AI
11. **Background Jobs** - Async capability
12. **Monitoring** - Production readiness

---

**Final Assessment**: Learning Desktop has a **solid foundation** with excellent data modeling and repository design. The primary blockers are in session management and service layer organization. Address these P0 issues, and the platform will be ready for production deployment.

**Estimated Time to Production**: 4-6 weeks (assuming 1 developer, full-time)

---

*Generated by Evolutionary Architect (Claude Sonnet 4.5)*
*Analysis Date: 2026-03-14*
*Project: Learning Desktop*
