# Learning Desktop - Database Schema

> Multi-tenant PostgreSQL schema with Row Level Security (RLS) for the "Get Ahead of AI" course platform.

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────┐
│                     Application Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐         │
│  │  Web Server │  │  AI Tutor   │  │  Media Svc  │         │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘         │
├─────────┼────────────────┼────────────────┼─────────────────┤
│         │                │                │                 │
│  ┌──────▼────────────────▼────────────────▼──────┐         │
│  │          PostgreSQL with RLS                   │         │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────┐    │         │
│  │  │  Tenant  │  │ Progress │  │  Course  │    │         │
│  │  │   Data   │  │  Events  │  │  Content │    │         │
│  │  └──────────┘  └──────────┘  └──────────┘    │         │
│  └────────────────────────────────────────────────┘         │
└─────────────────────────────────────────────────────────────┘
```

---

## Core Tables

### 1. Tenants

```sql
-- Tenant organizations (companies, schools, etc)
CREATE TABLE tenants (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name              TEXT NOT NULL,
    slug              TEXT UNIQUE NOT NULL,
    plan              TEXT NOT NULL DEFAULT 'free', -- 'free' | 'pro' | 'enterprise'
    max_students      INT NOT NULL DEFAULT 100,
    max_ai_calls      INT NOT NULL DEFAULT 10000, -- per month

    -- Billing
    billing_email     TEXT,
    billing_period    TEXT DEFAULT 'monthly', -- 'monthly' | 'annual'

    -- Configuration
    custom_domain     TEXT UNIQUE,
    logo_url          TEXT,
    theme_config      JSONB DEFAULT '{}',

    -- Timestamps
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT valid_plan CHECK (plan IN ('free', 'pro', 'enterprise'))
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_plan ON tenants(plan);

-- RLS: Only accessible by platform admins (not students)
ALTER TABLE tenants ENABLE ROW LEVEL SECURITY;
CREATE POLICY tenant_admin_only ON tenants
    USING (false); -- Application-layer enforcement
```

### 2. Students

```sql
-- Student users (belong to tenants)
CREATE TABLE students (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    external_id       TEXT, -- External auth ID (Auth0, Clerk, etc)

    -- Profile
    name              TEXT NOT NULL,
    email             TEXT NOT NULL,
    avatar_url        TEXT,

    -- AI Tutor Context
    skill_level       TEXT NOT NULL DEFAULT 'beginner', -- 'beginner' | 'intermediate' | 'advanced'
    interests         TEXT[] DEFAULT '{}',
    goals             TEXT[] DEFAULT '{}',

    -- Background info for personalization
    background_text   TEXT,
    industry          TEXT,
    role              TEXT,

    -- Status
    status            TEXT NOT NULL DEFAULT 'active', -- 'active' | 'inactive' | 'suspended'
    last_login_at     TIMESTAMPTZ,
    total_minutes     INT DEFAULT 0,

    -- Timestamps
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT valid_skill_level CHECK (skill_level IN ('beginner', 'intermediate', 'advanced')),
    CONSTRAINT valid_status CHECK (status IN ('active', 'inactive', 'suspended'))
);

CREATE INDEX idx_students_tenant ON students(tenant_id);
CREATE INDEX idx_students_external ON students(external_id);
CREATE INDEX idx_students_email ON students(email);
CREATE INDEX idx_students_status ON students(status);

-- RLS: Students can only see their own row; tenant admins see all
ALTER TABLE students ENABLE ROW LEVEL SECURITY;

CREATE POLICY students_select_self ON students
    FOR SELECT
    USING (id = current_setting('app.student_id', TRUE)::UUID);

CREATE POLICY students_select_tenant ON students
    FOR SELECT
    USING (tenant_id = current_setting('app.tenant_id', TRUE)::UUID)
    WITH CHECK (has_tenant_role(current_setting('app.tenant_id', TRUE)::UUID, 'admin'));
```

### 3. Student Sessions

```sql
-- Active chat sessions with AI tutor
CREATE TABLE student_sessions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id        UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Session state
    current_module    INT,
    current_lesson    INT,
    completed_ids     TEXT[] DEFAULT '{}',

    -- Status
    status            TEXT NOT NULL DEFAULT 'active', -- 'active' | 'completed' | 'abandoned'

    -- Timestamps
    started_at        TIMESTAMPTZ DEFAULT NOW(),
    last_active       TIMESTAMPTZ DEFAULT NOW(),
    completed_at      TIMESTAMPTZ,

    CONSTRAINT valid_session_status CHECK (status IN ('active', 'completed', 'abandoned'))
);

CREATE INDEX idx_sessions_student ON student_sessions(student_id);
CREATE INDEX idx_sessions_tenant ON student_sessions(tenant_id);
CREATE INDEX idx_sessions_active ON student_sessions(last_active)
    WHERE status = 'active';

-- RLS: Students see only their sessions
ALTER TABLE student_sessions ENABLE ROW LEVEL SECURITY;

CREATE POLICY sessions_select_own ON student_sessions
    FOR SELECT
    USING (student_id = current_setting('app.student_id', TRUE)::UUID);

CREATE POLICY sessions_insert_own ON student_sessions
    FOR INSERT
    WITH CHECK (student_id = current_setting('app.student_id', TRUE)::UUID);

CREATE POLICY sessions_update_own ON student_sessions
    FOR UPDATE
    USING (student_id = current_setting('app.student_id', TRUE)::UUID);
```

### 4. Course Content

```sql
-- Course definitions
CREATE TABLE courses (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    slug              TEXT UNIQUE NOT NULL,
    title             TEXT NOT NULL,
    description       TEXT,
    thumbnail_url     TEXT,

    -- Configuration
    is_published      BOOLEAN DEFAULT false,
    required_plan     TEXT DEFAULT 'free', -- Minimum plan to access

    -- Metadata
    difficulty        TEXT DEFAULT 'beginner',
    duration_weeks    INT DEFAULT 6,
    total_minutes     INT DEFAULT 600,

    -- Timestamps
    created_at        TIMESTAMPTZ DEFAULT NOW(),
    updated_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_courses_slug ON courses(slug);
CREATE INDEX idx_courses_published ON courses(is_published) WHERE is_published;

-- Course modules
CREATE TABLE course_modules (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    course_id         UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    order_index       INT NOT NULL,

    title             TEXT NOT NULL,
    description       TEXT,
    duration_weeks    INT DEFAULT 1,

    created_at        TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(course_id, order_index)
);

CREATE INDEX idx_modules_course ON course_modules(course_id);

-- Course lessons
CREATE TABLE course_lessons (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    module_id         UUID NOT NULL REFERENCES course_modules(id) ON DELETE CASCADE,
    order_index       INT NOT NULL,

    title             TEXT NOT NULL,
    description       TEXT,
    duration_minutes  INT DEFAULT 30,

    -- Content references
    video_id          TEXT, -- Remotion composition ID
    text_content      JSONB,
    interactive_type  TEXT, -- '3d' | 'board' | 'quiz' | 'code'
    interactive_data  JSONB,

    -- Checkpoint
    checkpoint_type   TEXT, -- 'quiz' | 'exercise' | 'project'
    checkpoint_data   JSONB,
    passing_score     FLOAT DEFAULT 0.8,

    -- Metadata
    difficulty        TEXT DEFAULT 'beginner',
    concepts          TEXT[] DEFAULT '{}',

    created_at        TIMESTAMPTZ DEFAULT NOW(),

    UNIQUE(module_id, order_index)
);

CREATE INDEX idx_lessons_module ON course_lessons(module_id);
```

### 5. Progress Tracking

```sql
-- Event-sourced progress events
CREATE TABLE progress_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id        UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    session_id        UUID REFERENCES student_sessions(id) ON DELETE SET NULL,

    -- Event type
    event_type        TEXT NOT NULL, -- 'lesson_started' | 'lesson_completed' | 'checkpoint_passed' | 'concept_learned'

    -- Event data
    lesson_id         UUID REFERENCES course_lessons(id),
    module_id         UUID REFERENCES course_modules(id),
    checkpoint_score  FLOAT,
    data              JSONB DEFAULT '{}',

    -- Timestamp & Versioning
    timestamp         TIMESTAMPTZ DEFAULT NOW(),
    version           INT DEFAULT 0
);

CREATE INDEX idx_progress_student ON progress_events(student_id, timestamp DESC);
CREATE INDEX idx_progress_tenant ON progress_events(tenant_id, timestamp DESC);
CREATE INDEX idx_progress_session ON progress_events(session_id);
CREATE INDEX idx_progress_type ON progress_events(event_type);

-- RLS: Students see only their events
ALTER TABLE progress_events ENABLE ROW LEVEL SECURITY;

CREATE POLICY progress_select_own ON progress_events
    FOR SELECT
    USING (student_id = current_setting('app.student_id', TRUE)::UUID);

-- Computed progress view (materialized for performance)
CREATE MATERIALIZED VIEW student_progress AS
SELECT
    s.id AS student_id,
    s.tenant_id,
    c.id AS course_id,
    COUNT(DISTINCT l.id) AS total_lessons,
    COUNT(DISTINCT CASE WHEN pe.event_type = 'lesson_completed' THEN l.id END) AS completed_lessons,
    MAX(l.order_index) AS current_lesson_index,
    MAX(COALESCE(cp.max_checkpoint_index, 0)) AS max_checkpoint_index,
    SUM(COALESCE(pe.checkpoint_score, 0)) AS total_checkpoint_score,
    COUNT(DISTINCT CASE WHEN pe.event_type = 'concept_learned' THEN pe.data->>'concept' END) AS concepts_learned,
    MIN(pe.timestamp) AS started_at,
    MAX(pe.timestamp) AS last_activity_at
FROM students s
CROSS JOIN courses c
LEFT JOIN course_modules cm ON cm.course_id = c.id
LEFT JOIN course_lessons l ON l.module_id = cm.id
LEFT JOIN progress_events pe ON pe.student_id = s.id
    AND pe.lesson_id = l.id
LEFT JOIN LATERAL (
    SELECT MAX(l2.order_index) AS max_checkpoint_index
    FROM course_lessons l2
    JOIN course_modules cm2 ON cm2.id = l2.module_id
    WHERE cm2.course_id = c.id
    AND EXISTS (
        SELECT 1 FROM progress_events pe2
        WHERE pe2.student_id = s.id
        AND pe2.lesson_id = l2.id
        AND pe2.event_type = 'checkpoint_passed'
    )
) cp ON true
GROUP BY s.id, s.tenant_id, c.id;

CREATE UNIQUE INDEX ON student_progress(student_id, course_id);
CREATE INDEX ON student_progress(tenant_id);

-- Refresh function for materialized view
CREATE OR REPLACE FUNCTION refresh_student_progress(student_uuid UUID)
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY student_progress;
END;
$$ LANGUAGE plpgsql;
```

### 6. Chat Messages

```sql
-- AI tutor conversation history
CREATE TABLE chat_messages (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id        UUID NOT NULL REFERENCES student_sessions(id) ON DELETE CASCADE,
    student_id        UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Message content
    role              TEXT NOT NULL, -- 'user' | 'assistant' | 'system'
    content           TEXT NOT NULL,

    -- Media injection
    media_type        TEXT, -- 'video' | '3d' | 'board' | 'code'
    media_source      TEXT,
    media_data        JSONB,

    -- Context for AI
    lesson_id         UUID REFERENCES course_lessons(id),
    metadata          JSONB DEFAULT '{}',

    -- Timestamp
    created_at        TIMESTAMPTZ DEFAULT NOW(),

    CONSTRAINT valid_role CHECK (role IN ('user', 'assistant', 'system'))
);

CREATE INDEX idx_chat_session ON chat_messages(session_id, created_at);
CREATE INDEX idx_chat_student ON chat_messages(student_id, created_at DESC);
CREATE INDEX idx_chat_tenant ON chat_messages(tenant_id);

-- RLS: Students see only their messages
ALTER TABLE chat_messages ENABLE ROW LEVEL SECURITY;

CREATE POLICY chat_select_own ON chat_messages
    FOR SELECT
    USING (student_id = current_setting('app.student_id', TRUE)::UUID);

CREATE POLICY chat_insert_own ON chat_messages
    FOR INSERT
    WITH CHECK (student_id = current_setting('app.student_id', TRUE)::UUID);
```

### 7. AI Usage Tracking

```sql
-- Track AI API usage per tenant (for billing/limits)
CREATE TABLE ai_usage_events (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    student_id        UUID REFERENCES students(id) ON DELETE SET NULL,

    -- Usage details
    model             TEXT NOT NULL, -- 'claude-3-5-sonnet' | 'claude-3-haiku'
    input_tokens      INT NOT NULL,
    output_tokens     INT NOT NULL,
    total_tokens      INT GENERATED ALWAYS AS (input_tokens + output_tokens) STORED,

    -- Request metadata
    request_type      TEXT NOT NULL, -- 'chat' | 'completion' | 'embedding'
    session_id        UUID,
    lesson_id         UUID,

    -- Timestamp
    created_at        TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_ai_usage_tenant ON ai_usage_events(tenant_id, created_at);
CREATE INDEX idx_ai_usage_monthly ON ai_usage_events(
    tenant_id,
    date_trunc('month', created_at)
);

-- Monthly usage aggregation (for billing)
CREATE MATERIALIZED VIEW monthly_ai_usage AS
SELECT
    tenant_id,
    date_trunc('month', created_at) AS month,
    model,
    COUNT(*) AS request_count,
    SUM(input_tokens) AS total_input_tokens,
    SUM(output_tokens) AS total_output_tokens,
    SUM(total_tokens) AS total_tokens
FROM ai_usage_events
GROUP BY tenant_id, date_trunc('month', created_at), model;

CREATE UNIQUE INDEX ON monthly_ai_usage(tenant_id, month, model);

CREATE OR REPLACE FUNCTION refresh_monthly_ai_usage()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY monthly_ai_usage;
END;
$$ LANGUAGE plpgsql;
```

### 8. Checkpoint Submissions

```sql
-- Student checkpoint/exercise submissions
CREATE TABLE checkpoint_submissions (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id        UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    lesson_id         UUID NOT NULL REFERENCES course_lessons(id),
    checkpoint_type   TEXT NOT NULL,

    -- Submission data
    answers           JSONB,
    score             FLOAT NOT NULL,
    passed            BOOLEAN NOT NULL,

    -- AI feedback
    feedback          TEXT,
    feedback_data     JSONB,

    -- Timestamps
    submitted_at      TIMESTAMPTZ DEFAULT NOW(),
    graded_at         TIMESTAMPTZ,

    CONSTRAINT valid_checkpoint_type CHECK (checkpoint_type IN ('quiz', 'exercise', 'project'))
);

CREATE INDEX idx_submissions_student ON checkpoint_submissions(student_id, submitted_at DESC);
CREATE INDEX idx_submissions_lesson ON checkpoint_submissions(lesson_id);

-- RLS: Students see only their submissions
ALTER TABLE checkpoint_submissions ENABLE ROW LEVEL SECURITY;

CREATE POLICY submissions_select_own ON checkpoint_submissions
    FOR SELECT
    USING (student_id = current_setting('app.student_id', TRUE)::UUID);

-- Best score per lesson (for grade book)
CREATE MATERIALIZED VIEW lesson_best_scores AS
SELECT
    student_id,
    tenant_id,
    lesson_id,
    MAX(score) AS best_score,
    COUNT(*) AS attempts,
    MAX(submitted_at) AS last_attempt
FROM checkpoint_submissions
GROUP BY student_id, tenant_id, lesson_id;

CREATE UNIQUE INDEX ON lesson_best_scores(student_id, lesson_id);
```

### 9. Certificates

```sql
-- Course completion certificates
CREATE TABLE certificates (
    id                UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    student_id        UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id         UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    course_id         UUID NOT NULL REFERENCES courses(id),

    -- Certificate details
    certificate_number TEXT UNIQUE NOT NULL,
    verification_hash TEXT UNIQUE NOT NULL,

    -- Completion data
    completed_at      TIMESTAMPTZ DEFAULT NOW(),
    total_minutes     INT NOT NULL,
    final_score       FLOAT NOT NULL,

    -- Certificate rendering
    pdf_url           TEXT,
    badge_url         TEXT,

    -- Verification
    revoked           BOOLEAN DEFAULT false,
    revoked_at        TIMESTAMPTZ,
    revoked_reason    TEXT
);

CREATE INDEX idx_certificates_student ON certificates(student_id);
CREATE INDEX idx_certificates_tenant ON certificates(tenant_id);
CREATE INDEX idx_certificates_verification ON certificates(verification_hash);

-- RLS: Students see only their certificates
ALTER TABLE certificates ENABLE ROW LEVEL SECURITY;

CREATE POLICY certificates_select_own ON certificates
    FOR SELECT
    USING (student_id = current_setting('app.student_id', TRUE)::UUID);
```

---

## Row Level Security (RLS) Setup

### Helper Functions

```sql
-- Check if user has tenant role
CREATE OR REPLACE FUNCTION has_tenant_role(tenant_uuid UUID, role_name TEXT)
RETURNS BOOLEAN AS $$
BEGIN
    -- Check application-managed role table
    -- This would be populated by your auth layer
    RETURN EXISTS (
        SELECT 1 FROM tenant_roles
        WHERE tenant_id = tenant_uuid
        AND user_id = current_setting('app.user_id', TRUE)::UUID
        AND role = role_name
    );
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Set tenant context (called at connection start)
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_uuid UUID)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.tenant_id', tenant_uuid::TEXT, TRUE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Set student context (called after authentication)
CREATE OR REPLACE FUNCTION set_student_context(student_uuid UUID)
RETURNS void AS $$
BEGIN
    PERFORM set_config('app.student_id', student_uuid::TEXT, TRUE);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;
```

---

## Migration Strategy

### Initial Migration

```sql
-- 001_initial_schema.up.sql
BEGIN;

-- Create all tables in order
-- (copy CREATE TABLE statements above)

COMMIT;
```

### Sample Data

```sql
-- Insert default course
INSERT INTO courses (slug, title, description, is_published) VALUES
('get-ahead-of-ai', 'Get Ahead of AI', 'Learn to not just survive but thrive in the AI era.', true);

-- Insert modules
INSERT INTO course_modules (course_id, order_index, title, description)
SELECT
    c.id,
    unnest(ARRAY[1,2,3,4,5,6]),
    unnest(ARRAY[
        'The Reality Check',
        'AI Literacy',
        'Your AI Workflow',
        'AI in Your Domain',
        'Staying Ahead',
        'The Human Advantage'
    ]),
    unnest(ARRAY[
        'Confront the AI revolution honestly',
        'Build foundational understanding of how AI works',
        'Integrate AI into daily work and learning',
        'Apply AI knowledge to your specific industry',
        'Build systems for continuous adaptation',
        'Lean into what makes humans irreplaceable'
    ])
FROM courses c
WHERE c.slug = 'get-ahead-of-ai';
```

---

## Performance Considerations

### Indexes Summary

| Table | Indexes | Purpose |
|-------|--------|---------|
| `tenants` | slug, plan | Admin queries, billing |
| `students` | tenant_id, external_id, email, status | Auth, RLS filtering |
| `student_sessions` | student_id, tenant_id, last_active | Active session lookup |
| `chat_messages` | session_id, student_id, tenant_id | Conversation history |
| `progress_events` | student_id, tenant_id, event_type | Event sourcing queries |
| `ai_usage_events` | tenant_id, monthly | Billing aggregation |

### Materialized Views

| View | Refresh Strategy | Usage |
|------|------------------|-------|
| `student_progress` | On progress event | Dashboard, course resume |
| `monthly_ai_usage` | Hourly cron | Billing reports |
| `lesson_best_scores` | On submission | Grade book |

---

## Data Retention Policy

```sql
-- Archive old chat messages (after 1 year)
CREATE OR REPLACE FUNCTION archive_old_chat_messages()
RETURNS void AS $$
BEGIN
    -- Move to archive table (separate schema/database)
    -- Implement based on your archival needs
END;
$$ LANGUAGE plpgsql;

-- Cleanup abandoned sessions (after 30 days)
CREATE OR REPLACE FUNCTION cleanup_abandoned_sessions()
RETURNS void AS $$
BEGIN
    UPDATE student_sessions
    SET status = 'abandoned'
    WHERE status = 'active'
    AND last_active < NOW() - INTERVAL '30 days';
END;
$$ LANGUAGE plpgsql;
```

---

## Go Structs

```go
// Student represents a student user
type Student struct {
    ID             uuid.UUID `db:"id" json:"id"`
    TenantID       uuid.UUID `db:"tenant_id" json:"tenant_id"`
    ExternalID     string    `db:"external_id" json:"external_id,omitempty"`
    Name           string    `db:"name" json:"name"`
    Email          string    `db:"email" json:"email"`
    AvatarURL      string    `db:"avatar_url" json:"avatar_url,omitempty"`
    SkillLevel     string    `db:"skill_level" json:"skill_level"`
    Interests      []string  `db:"interests" json:"interests"`
    Goals          []string  `db:"goals" json:"goals"`
    BackgroundText string    `db:"background_text" json:"background_text,omitempty"`
    Industry       string    `db:"industry" json:"industry,omitempty"`
    Role           string    `db:"role" json:"role,omitempty"`
    Status         string    `db:"status" json:"status"`
    LastLoginAt    time.Time `db:"last_login_at" json:"last_login_at"`
    TotalMinutes   int       `db:"total_minutes" json:"total_minutes"`
    CreatedAt      time.Time `db:"created_at" json:"created_at"`
    UpdatedAt      time.Time `db:"updated_at" json:"updated_at"`
}

// CourseProgress represents student progress in a course
type CourseProgress struct {
    StudentID           uuid.UUID `db:"student_id" json:"student_id"`
    TenantID            uuid.UUID `db:"tenant_id" json:"tenant_id"`
    CourseID            uuid.UUID `db:"course_id" json:"course_id"`
    TotalLessons        int       `db:"total_lessons" json:"total_lessons"`
    CompletedLessons    int       `db:"completed_lessons" json:"completed_lessons"`
    CurrentLessonIndex  int       `db:"current_lesson_index" json:"current_lesson_index"`
    ConceptsLearned     int       `db:"concepts_learned" json:"concepts_learned"`
    StartedAt           time.Time `db:"started_at" json:"started_at"`
    LastActivityAt      time.Time `db:"last_activity_at" json:"last_activity_at"`
    CompletionPercent   float64   `json:"completion_percent"`
}

// CheckpointSubmission represents a student submission
type CheckpointSubmission struct {
    ID             uuid.UUID `db:"id" json:"id"`
    StudentID      uuid.UUID `db:"student_id" json:"student_id"`
    TenantID       uuid.UUID `db:"tenant_id" json:"tenant_id"`
    LessonID       uuid.UUID `db:"lesson_id" json:"lesson_id"`
    CheckpointType string    `db:"checkpoint_type" json:"checkpoint_type"`
    Answers        json.RawMessage `db:"answers" json:"answers"`
    Score          float64   `db:"score" json:"score"`
    Passed         bool      `db:"passed" json:"passed"`
    Feedback       string    `db:"feedback" json:"feedback,omitempty"`
    SubmittedAt    time.Time `db:"submitted_at" json:"submitted_at"`
}
```

---

## Next Steps

1. **Media Injection Logic** → Rules engine for AI tutor media decisions
2. **htmx-r Integration** → Connect chat component to course backend
3. **API Handlers** → Go handlers for all database operations
