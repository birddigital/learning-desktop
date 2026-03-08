-- ============================================================================
-- Learning Desktop Schema
-- Multi-tenant PostgreSQL with Row-Level Security
-- ============================================================================

-- Enable required extensions
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- ============================================================================
-- TENANT MANAGEMENT
-- ============================================================================

-- Tenants represent organizations (schools, families, companies)
CREATE TABLE tenants (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    slug            TEXT NOT NULL UNIQUE,
    name            TEXT NOT NULL,
    domain          TEXT,

    -- Settings
    settings        JSONB DEFAULT '{}'::jsonb,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Soft delete
    deleted_at      TIMESTAMPTZ
);

CREATE INDEX idx_tenants_slug ON tenants(slug);
CREATE INDEX idx_tenants_deleted ON tenants(deleted_at) WHERE deleted_at IS NULL;

-- ============================================================================
-- STUDENTS
-- ============================================================================

CREATE TABLE students (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identity
    name            TEXT NOT NULL,
    email           TEXT,
    avatar_url      TEXT,

    -- Learning profile
    level           TEXT NOT NULL DEFAULT 'beginner', -- beginner, intermediate, advanced
    goals           JSONB DEFAULT '[]'::jsonb,
    interests       JSONB DEFAULT '[]'::jsonb,

    -- Settings
    settings        JSONB DEFAULT '{}'::jsonb,

    -- Status
    status          TEXT NOT NULL DEFAULT 'active', -- active, inactive, archived
    last_active_at  TIMESTAMPTZ,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Ensure unique email per tenant
    UNIQUE(tenant_id, email)
);

CREATE INDEX idx_students_tenant ON students(tenant_id);
CREATE INDEX idx_students_email ON students(email);
CREATE INDEX idx_students_status ON students(status);

-- ============================================================================
-- STUDENT SESSIONS
-- ============================================================================

CREATE TABLE student_sessions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Session info
    started_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    ended_at        TIMESTAMPTZ,
    duration_seconds INT,

    -- Context
    entry_point     TEXT, -- 'chat', 'lesson', 'checkpoint'
    device_type     TEXT, -- 'desktop', 'mobile', 'tablet'

    -- Activity summary
    messages_sent   INT DEFAULT 0,
    lessons_completed INT DEFAULT 0,
    checkpoints_passed INT DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_sessions_student ON student_sessions(student_id);
CREATE INDEX idx_sessions_tenant ON student_sessions(tenant_id);
CREATE INDEX idx_sessions_started ON student_sessions(started_at DESC);

-- ============================================================================
-- COURSES
-- ============================================================================

CREATE TABLE courses (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identity
    slug            TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    icon            TEXT, -- emoji or icon name

    -- Course structure
    level           TEXT NOT NULL DEFAULT 'beginner',
    category        TEXT NOT NULL,
    tags            JSONB DEFAULT '[]'::jsonb,

    -- Progress tracking
    total_lessons   INT NOT NULL DEFAULT 0,
    max_points      INT NOT NULL DEFAULT 100,

    -- Enrollment
    is_published    BOOLEAN NOT NULL DEFAULT false,
    is_public       BOOLEAN NOT NULL DEFAULT false,
    enrollment_limit INT,

    -- Timestamps
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_courses_tenant ON courses(tenant_id);
CREATE INDEX idx_courses_slug ON courses(slug);
CREATE INDEX idx_courses_category ON courses(category);
CREATE INDEX idx_courses_published ON courses(is_published) WHERE is_published = true;

-- ============================================================================
-- COURSE MODULES
-- ============================================================================

CREATE TABLE course_modules (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    course_id       UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identity
    slug            TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,

    -- Order within course
    sort_order      INT NOT NULL DEFAULT 0,

    -- Progress
    total_lessons   INT NOT NULL DEFAULT 0,
    max_points      INT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(course_id, slug)
);

CREATE INDEX idx_modules_course ON course_modules(course_id);
CREATE INDEX idx_modules_tenant ON course_modules(tenant_id);

-- ============================================================================
-- COURSE LESSONS
-- ============================================================================

CREATE TABLE course_lessons (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    module_id       UUID NOT NULL REFERENCES course_modules(id) ON DELETE CASCADE,
    course_id       UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identity
    slug            TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    content         TEXT, -- markdown content

    -- Media
    video_url       TEXT,
    audio_url       TEXT,

    -- Order within module
    sort_order      INT NOT NULL DEFAULT 0,

    -- Difficulty
    difficulty      TEXT NOT NULL DEFAULT 'easy', -- easy, medium, hard
    estimated_minutes INT NOT NULL DEFAULT 30,

    -- Points
    max_points      INT NOT NULL DEFAULT 10,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(module_id, slug)
);

CREATE INDEX idx_lessons_module ON course_lessons(module_id);
CREATE INDEX idx_lessons_course ON course_lessons(course_id);
CREATE INDEX idx_lessons_tenant ON course_lessons(tenant_id);

-- ============================================================================
-- PROGRESS EVENTS (Event Sourcing)
-- ============================================================================

CREATE TABLE progress_events (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Event type
    event_type      TEXT NOT NULL, -- lesson_started, lesson_completed, checkpoint_passed, etc.
    entity_type     TEXT NOT NULL, -- lesson, module, course, skill_node
    entity_id       UUID NOT NULL,

    -- Score/Progress
    points_earned   INT NOT NULL DEFAULT 0,
    score           FLOAT NOT NULL DEFAULT 0, -- 0-100

    -- Metadata
    metadata        JSONB DEFAULT '{}'::jsonb,

    -- Timestamp
    occurred_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_events_student ON progress_events(student_id);
CREATE INDEX idx_events_tenant ON progress_events(tenant_id);
CREATE INDEX idx_events_type ON progress_events(event_type);
CREATE INDEX idx_events_entity ON progress_events(entity_type, entity_id);
CREATE INDEX idx_events_occurred ON progress_events(occurred_at DESC);

-- Materialized view for fast progress queries
CREATE MATERIALIZED VIEW student_progress_mv AS
SELECT
    student_id,
    tenant_id,
    entity_type,
    entity_id,
    COUNT(*) as event_count,
    SUM(points_earned) as total_points,
    MAX(score) as max_score,
    MAX(occurred_at) as last_activity,
    MAX(CASE WHEN event_type = 'lesson_completed' THEN occurred_at END) as completed_at
FROM progress_events
GROUP BY student_id, tenant_id, entity_type, entity_id;

CREATE UNIQUE INDEX ON student_progress_mv(student_id, entity_type, entity_id);
CREATE INDEX ON student_progress_mv(tenant_id);

-- ============================================================================
-- STUDENT PROGRESS (Computed Summary)
-- ============================================================================

CREATE TABLE student_progress (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Progress target
    entity_type     TEXT NOT NULL,
    entity_id       UUID NOT NULL,

    -- Computed progress
    score           FLOAT NOT NULL DEFAULT 0,
    points_earned   INT NOT NULL DEFAULT 0,
    max_points      INT NOT NULL DEFAULT 0,
    percent_complete FLOAT NOT NULL DEFAULT 0,

    -- Status
    status          TEXT NOT NULL DEFAULT 'not_started', -- not_started, in_progress, completed
    started_at      TIMESTAMPTZ,
    completed_at    TIMESTAMPTZ,

    -- Computed last activity
    last_activity_at TIMESTAMPTZ,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(student_id, entity_type, entity_id)
);

CREATE INDEX idx_progress_student ON student_progress(student_id);
CREATE INDEX idx_progress_tenant ON student_progress(tenant_id);
CREATE INDEX idx_progress_status ON student_progress(status);

-- ============================================================================
-- CHAT MESSAGES
-- ============================================================================

CREATE TABLE chat_messages (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    session_id      UUID REFERENCES student_sessions(id) ON DELETE SET NULL,

    -- Message
    role            TEXT NOT NULL, -- user, assistant, system
    content         TEXT NOT NULL,

    -- AI Metadata
    model           TEXT,
    tokens_used     INT,
    cost_cents      INT,

    -- Voice metadata (if applicable)
    is_voice        BOOLEAN NOT NULL DEFAULT false,
    voice_transcript TEXT, -- original speech text
    voice_audio_url TEXT,

    -- Context
    context_type    TEXT, -- lesson, checkpoint, general
    context_id      UUID,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_chat_student ON chat_messages(student_id);
CREATE INDEX idx_chat_tenant ON chat_messages(tenant_id);
CREATE INDEX idx_chat_session ON chat_messages(session_id);
CREATE INDEX idx_chat_created ON chat_messages(created_at DESC);

-- ============================================================================
-- AI USAGE TRACKING
-- ============================================================================

CREATE TABLE ai_usage (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,
    student_id      UUID REFERENCES students(id) ON DELETE SET NULL,

    -- Usage details
    model           TEXT NOT NULL,
    provider        TEXT NOT NULL, -- anthropic, openai, local

    -- Tokens
    prompt_tokens   INT NOT NULL DEFAULT 0,
    completion_tokens INT NOT NULL DEFAULT 0,
    total_tokens    INT NOT NULL DEFAULT 0,

    -- Cost
    cost_cents      INT NOT NULL DEFAULT 0,

    -- Request type
    request_type    TEXT NOT NULL, -- chat, assessment, insight, voice

    -- Period
    month           TEXT NOT NULL DEFAULT TO_CHAR(NOW(), 'YYYY-MM'),

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_usage_tenant ON ai_usage(tenant_id);
CREATE INDEX idx_usage_student ON ai_usage(student_id);
CREATE INDEX idx_usage_month ON ai_usage(month DESC);

-- ============================================================================
-- CHECKPOINT SUBMISSIONS
-- ============================================================================

CREATE TABLE checkpoint_submissions (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- What checkpoint
    lesson_id       UUID NOT NULL REFERENCES course_lessons(id) ON DELETE CASCADE,

    -- Submission
    content         TEXT NOT NULL,
    attachment_url  TEXT,

    -- Grading
    score           FLOAT, -- 0-100
    points_earned   INT,
    feedback        TEXT,
    graded_by       TEXT, -- ai, instructor
    graded_at       TIMESTAMPTZ,

    -- Status
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, passed, failed, needs_revision

    -- Attempts
    attempt_number  INT NOT NULL DEFAULT 1,
    max_attempts    INT NOT NULL DEFAULT 3,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_checkpoints_student ON checkpoint_submissions(student_id);
CREATE INDEX idx_checkpoints_tenant ON checkpoint_submissions(tenant_id);
CREATE INDEX idx_checkpoints_lesson ON checkpoint_submissions(lesson_id);
CREATE INDEX idx_checkpoints_status ON checkpoint_submissions(status);

-- ============================================================================
-- CERTIFICATES
-- ============================================================================

CREATE TABLE certificates (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Achievement
    course_id       UUID NOT NULL REFERENCES courses(id) ON DELETE CASCADE,
    title           TEXT NOT NULL,
    description     TEXT,

    -- Completion details
    final_score     FLOAT NOT NULL,
    completed_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Certificate
    certificate_url TEXT,
    verification_code TEXT UNIQUE NOT NULL DEFAULT encode(gen_random_bytes(16), 'hex'),

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_certificates_student ON certificates(student_id);
CREATE INDEX idx_certificates_tenant ON certificates(tenant_id);
CREATE INDEX idx_certificates_course ON certificates(course_id);
CREATE INDEX idx_certificates_verification ON certificates(verification_code);

-- ============================================================================
-- GOALS & ACCOUNTABILITY
-- ============================================================================

CREATE TABLE goals (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Goal
    title           TEXT NOT NULL,
    description     TEXT,
    category        TEXT NOT NULL, -- learning, fitness, career, personal

    -- Timeline
    target_date     TIMESTAMPTZ NOT NULL,

    -- AI Planning
    confidence      FLOAT NOT NULL DEFAULT 0.5, -- AI's confidence in achievability
    planned_by_ai   BOOLEAN NOT NULL DEFAULT true,

    -- Status
    status          TEXT NOT NULL DEFAULT 'planned', -- planned, active, on_track, at_risk, behind, complete, cancelled

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_goals_student ON goals(student_id);
CREATE INDEX idx_goals_tenant ON goals(tenant_id);
CREATE INDEX idx_goals_status ON goals(status);
CREATE INDEX idx_goals_target ON goals(target_date);

CREATE TABLE milestones (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    goal_id         UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    title           TEXT NOT NULL,
    description     TEXT,

    -- Timeline
    due_date        TIMESTAMPTZ NOT NULL,

    -- Dependencies
    dependencies    JSONB DEFAULT '[]'::jsonb, -- IDs of other milestones

    -- Status
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, in_progress, complete, skipped

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_milestones_goal ON milestones(goal_id);
CREATE INDEX idx_milestones_tenant ON milestones(tenant_id);
CREATE INDEX idx_milestones_status ON milestones(status);
CREATE INDEX idx_milestones_due ON milestones(due_date);

CREATE TABLE schedule_blocks (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    goal_id         UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Schedule
    title           TEXT NOT NULL,
    scheduled_for   TIMESTAMPTZ NOT NULL,
    duration_minutes INT NOT NULL,

    -- Completion
    is_completed    BOOLEAN NOT NULL DEFAULT false,
    completed_at    TIMESTAMPTZ,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_schedule_goal ON schedule_blocks(goal_id);
CREATE INDEX idx_schedule_tenant ON schedule_blocks(tenant_id);
CREATE INDEX idx_schedule_for ON schedule_blocks(scheduled_for);

CREATE TABLE communications (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    goal_id         UUID NOT NULL REFERENCES goals(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Communication
    type            TEXT NOT NULL, -- reminder, escalation, update, celebration
    channel         TEXT NOT NULL, -- email, sms, in_app
    recipient       TEXT NOT NULL, -- email, phone, or user_id

    -- Content
    subject         TEXT,
    body            TEXT NOT NULL,

    -- Scheduling
    scheduled_at    TIMESTAMPTZ NOT NULL,
    sent_at         TIMESTAMPTZ,

    -- Status
    status          TEXT NOT NULL DEFAULT 'pending', -- pending, sent, failed

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_communications_goal ON communications(goal_id);
CREATE INDEX idx_communications_tenant ON communications(tenant_id);
CREATE INDEX idx_communications_scheduled ON communications(scheduled_at);
CREATE INDEX idx_communications_status ON communications(status);

-- ============================================================================
-- INSIGHT SNAPSHOTS
-- ============================================================================

CREATE TABLE insight_snapshots (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Snapshot data
    on_track        BOOLEAN NOT NULL,
    velocity        FLOAT NOT NULL, -- points per day
    urgency         TEXT NOT NULL, -- critical, high, medium, low, none

    -- Projections
    eta_days        INT,
    projected_completion_date TIMESTAMPTZ,

    -- Analysis
    risk_factors    JSONB DEFAULT '[]'::jsonb,
    strengths       JSONB DEFAULT '[]'::jsonb,
    recommendations JSONB DEFAULT '[]'::jsonb,

    -- Period
    recorded_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_insights_student ON insight_snapshots(student_id);
CREATE INDEX idx_insights_tenant ON insight_snapshots(tenant_id);
CREATE INDEX idx_insights_recorded ON insight_snapshots(recorded_at DESC);

-- ============================================================================
-- SKILL TREES
-- ============================================================================

CREATE TABLE skill_trees (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identity
    slug            TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    icon            TEXT,

    -- Category
    category        TEXT NOT NULL, -- technical, soft_skills, domain, tools, career

    -- Requirements
    required_level  TEXT NOT NULL DEFAULT 'beginner',

    -- Progress tracking
    total_nodes     INT NOT NULL DEFAULT 0,
    max_points      INT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tenant_id, slug)
);

CREATE INDEX idx_skilltrees_tenant ON skill_trees(tenant_id);
CREATE INDEX idx_skilltrees_slug ON skill_trees(slug);
CREATE INDEX idx_skilltrees_category ON skill_trees(category);

CREATE TABLE skill_nodes (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    tree_id         UUID NOT NULL REFERENCES skill_trees(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Identity
    slug            TEXT NOT NULL,
    title           TEXT NOT NULL,
    description     TEXT,
    icon            TEXT,

    -- Position (for visualization)
    position_row    INT NOT NULL DEFAULT 0,
    position_col    INT NOT NULL DEFAULT 0,
    position_x      FLOAT,
    position_y      FLOAT,

    -- Requirements
    required_score  FLOAT NOT NULL DEFAULT 0,
    required_nodes  JSONB DEFAULT '[]'::jsonb, -- Array of node IDs

    -- Scoring
    max_points      INT NOT NULL DEFAULT 100,
    weight          FLOAT NOT NULL DEFAULT 1.0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(tree_id, slug)
);

CREATE INDEX idx_skillnodes_tree ON skill_nodes(tree_id);
CREATE INDEX idx_skillnodes_tenant ON skill_nodes(tenant_id);

-- ============================================================================
-- STUDENT SKILLS
-- ============================================================================

CREATE TABLE student_skills (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    node_id         UUID NOT NULL REFERENCES skill_nodes(id) ON DELETE CASCADE,
    tree_id         UUID NOT NULL REFERENCES skill_trees(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Current state
    score           FLOAT NOT NULL DEFAULT 0,
    level           TEXT NOT NULL DEFAULT 'unknown', -- unknown, novice, apprentice, adept, expert, master
    points_earned   INT NOT NULL DEFAULT 0,
    max_points      INT NOT NULL DEFAULT 0,

    -- Status
    unlocked        BOOLEAN NOT NULL DEFAULT false,
    unlocked_at     TIMESTAMPTZ,
    completed       BOOLEAN NOT NULL DEFAULT false,
    completed_at    TIMESTAMPTZ,

    -- Assessment tracking
    last_assessed_at TIMESTAMPTZ,
    assessment_count INT NOT NULL DEFAULT 0,

    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(student_id, node_id)
);

CREATE INDEX idx_studentskills_student ON student_skills(student_id);
CREATE INDEX idx_studentskills_node ON student_skills(node_id);
CREATE INDEX idx_studentskills_tree ON student_skills(tree_id);
CREATE INDEX idx_studentskills_tenant ON student_skills(tenant_id);
CREATE INDEX idx_studentskills_level ON student_skills(level);

-- ============================================================================
-- ASSESSMENTS
-- ============================================================================

CREATE TABLE assessments (
    id              UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    student_id      UUID NOT NULL REFERENCES students(id) ON DELETE CASCADE,
    node_id         UUID NOT NULL REFERENCES skill_nodes(id) ON DELETE CASCADE,
    tree_id         UUID NOT NULL REFERENCES skill_trees(id) ON DELETE CASCADE,
    tenant_id       UUID NOT NULL REFERENCES tenants(id) ON DELETE CASCADE,

    -- Assessment method
    method          TEXT NOT NULL, -- quiz, project, code, chat, voice, self, peer, interview

    -- Score
    score           FLOAT NOT NULL,
    weight          FLOAT NOT NULL DEFAULT 1.0,

    -- Evidence (stored as JSONB)
    responses       JSONB DEFAULT '[]'::jsonb,
    submissions     JSONB DEFAULT '[]'::jsonb,
    chat_analysis   JSONB,
    voice_analysis  JSONB,

    -- AI feedback
    feedback        TEXT,
    next_steps      JSONB DEFAULT '[]'::jsonb,

    -- Timestamps
    assessed_at     TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_assessments_student ON assessments(student_id);
CREATE INDEX idx_assessments_node ON assessments(node_id);
CREATE INDEX idx_assessments_tenant ON assessments(tenant_id);
CREATE INDEX idx_assessments_method ON assessments(method);

-- ============================================================================
-- ROW LEVEL SECURITY (RLS)
-- ============================================================================

-- Enable RLS on all tenant-scoped tables
ALTER TABLE students ENABLE ROW LEVEL SECURITY;
ALTER TABLE student_sessions ENABLE ROW LEVEL SECURITY;
ALTER TABLE courses ENABLE ROW LEVEL SECURITY;
ALTER TABLE course_modules ENABLE ROW LEVEL SECURITY;
ALTER TABLE course_lessons ENABLE ROW LEVEL SECURITY;
ALTER TABLE progress_events ENABLE ROW LEVEL SECURITY;
ALTER TABLE student_progress ENABLE ROW LEVEL SECURITY;
ALTER TABLE chat_messages ENABLE ROW LEVEL SECURITY;
ALTER TABLE checkpoint_submissions ENABLE ROW LEVEL SECURITY;
ALTER TABLE certificates ENABLE ROW LEVEL SECURITY;
ALTER TABLE goals ENABLE ROW LEVEL SECURITY;
ALTER TABLE milestones ENABLE ROW LEVEL SECURITY;
ALTER TABLE schedule_blocks ENABLE ROW LEVEL SECURITY;
ALTER TABLE communications ENABLE ROW LEVEL SECURITY;
ALTER TABLE insight_snapshots ENABLE ROW LEVEL SECURITY;
ALTER TABLE skill_trees ENABLE ROW LEVEL SECURITY;
ALTER TABLE skill_nodes ENABLE ROW LEVEL SECURITY;
ALTER TABLE student_skills ENABLE ROW LEVEL SECURITY;
ALTER TABLE assessments ENABLE ROW LEVEL SECURITY;

-- Function to set tenant context
CREATE OR REPLACE FUNCTION set_tenant_context(tenant_id UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_tenant_id', tenant_id::text, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Function to set student context (within tenant)
CREATE OR REPLACE FUNCTION set_student_context(student_id UUID)
RETURNS VOID AS $$
BEGIN
    PERFORM set_config('app.current_student_id', student_id::text, false);
END;
$$ LANGUAGE plpgsql SECURITY DEFINER;

-- Helper function to get current tenant_id
CREATE OR REPLACE FUNCTION current_tenant_id()
RETURNS UUID AS $$
    SELECT NULLIF(current_setting('app.current_tenant_id', true), '')::UUID;
$$ LANGUAGE sql STABLE;

-- Helper function to get current student_id
CREATE OR REPLACE FUNCTION current_student_id()
RETURNS UUID AS $$
    SELECT NULLIF(current_setting('app.current_student_id', true), '')::UUID;
$$ LANGUAGE sql STABLE;

-- RLS Policies: Students can only access their tenant's data
CREATE POLICY tenant_isolation ON students
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_sessions ON student_sessions
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_courses ON courses
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_modules ON course_modules
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_lessons ON course_lessons
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_events ON progress_events
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_progress ON student_progress
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_chat ON chat_messages
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_checkpoints ON checkpoint_submissions
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_certificates ON certificates
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_goals ON goals
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_milestones ON milestones
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_schedule ON schedule_blocks
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_communications ON communications
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_insights ON insight_snapshots
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_skilltrees ON skill_trees
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_skillnodes ON skill_nodes
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_studentskills ON student_skills
    USING (tenant_id = current_tenant_id());

CREATE POLICY tenant_isolation_assessments ON assessments
    USING (tenant_id = current_tenant_id());

-- Student-scoped policies (within tenant)
CREATE POLICY student_isolation ON student_sessions
    USING (student_id = current_student_id())
    WITH CHECK (student_id = current_student_id());

CREATE POLICY student_isolation_progress ON progress_events
    USING (student_id = current_student_id())
    WITH CHECK (student_id = current_student_id());

CREATE POLICY student_isolation_student_progress ON student_progress
    USING (student_id = current_student_id());

CREATE POLICY student_isolation_chat ON chat_messages
    USING (student_id = current_student_id())
    WITH CHECK (student_id = current_student_id());

CREATE POLICY student_isolation_checkpoints ON checkpoint_submissions
    USING (student_id = current_student_id())
    WITH CHECK (student_id = current_student_id());

CREATE POLICY student_isolation_certificates ON certificates
    USING (student_id = current_student_id());

CREATE POLICY student_isolation_goals ON goals
    USING (student_id = current_student_id())
    WITH CHECK (student_id = current_student_id());

CREATE POLICY student_isolation_insights ON insight_snapshots
    USING (student_id = current_student_id());

CREATE POLICY student_isolation_studentskills ON student_skills
    USING (student_id = current_student_id())
    WITH CHECK (student_id = current_student_id());

CREATE POLICY student_isolation_assessments ON assessments
    USING (student_id = current_student_id())
    WITH CHECK (student_id = current_student_id());

-- ============================================================================
-- FUNCTIONS AND TRIGGERS
-- ============================================================================

-- Function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_student_progress_mv()
RETURNS TRIGGER AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY student_progress_mv;
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

-- Trigger to refresh MV on progress events
CREATE TRIGGER trigger_refresh_progress_mv
    AFTER INSERT OR UPDATE ON progress_events
    FOR EACH STATEMENT
    EXECUTE FUNCTION refresh_student_progress_mv();

-- Function to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Add updated_at triggers to relevant tables
CREATE TRIGGER update_tenants_updated_at BEFORE UPDATE ON tenants
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_students_updated_at BEFORE UPDATE ON students
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_courses_updated_at BEFORE UPDATE ON courses
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_modules_updated_at BEFORE UPDATE ON course_modules
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_lessons_updated_at BEFORE UPDATE ON course_lessons
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_progress_updated_at BEFORE UPDATE ON student_progress
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_checkpoints_updated_at BEFORE UPDATE ON checkpoint_submissions
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_goals_updated_at BEFORE UPDATE ON goals
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_milestones_updated_at BEFORE UPDATE ON milestones
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_skilltrees_updated_at BEFORE UPDATE ON skill_trees
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

CREATE TRIGGER update_studentskills_updated_at BEFORE UPDATE ON student_skills
    FOR EACH ROW EXECUTE FUNCTION update_updated_at();

-- ============================================================================
-- COMPLETE
-- ============================================================================
