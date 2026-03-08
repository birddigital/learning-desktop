-- ============================================================================
-- Learning Desktop Schema Rollback
-- Drops all tables, functions, and extensions in reverse order
-- ============================================================================

-- Drop triggers first
DROP TRIGGER IF EXISTS update_studentskills_updated_at ON student_skills;
DROP TRIGGER IF EXISTS update_skilltrees_updated_at ON skill_trees;
DROP TRIGGER IF EXISTS update_milestones_updated_at ON milestones;
DROP TRIGGER IF EXISTS update_goals_updated_at ON goals;
DROP TRIGGER IF EXISTS update_checkpoints_updated_at ON checkpoint_submissions;
DROP TRIGGER IF EXISTS update_progress_updated_at ON student_progress;
DROP TRIGGER IF EXISTS update_lessons_updated_at ON course_lessons;
DROP TRIGGER IF EXISTS update_modules_updated_at ON course_modules;
DROP TRIGGER IF EXISTS update_courses_updated_at ON courses;
DROP TRIGGER IF EXISTS update_students_updated_at ON students;
DROP TRIGGER IF EXISTS update_tenants_updated_at ON tenants;

DROP TRIGGER IF EXISTS trigger_refresh_progress_mv ON progress_events;

-- Drop functions
DROP FUNCTION IF EXISTS refresh_student_progress_mv();
DROP FUNCTION IF EXISTS update_updated_at();
DROP FUNCTION IF EXISTS current_student_id();
DROP FUNCTION IF EXISTS current_tenant_id();
DROP FUNCTION IF EXISTS set_student_context(UUID);
DROP FUNCTION IF EXISTS set_tenant_context(UUID);

-- Drop RLS policies
DROP POLICY IF EXISTS student_isolation_assessments ON assessments;
DROP POLICY IF EXISTS student_isolation_studentskills ON student_skills;
DROP POLICY IF EXISTS student_isolation_insights ON insight_snapshots;
DROP POLICY IF EXISTS student_isolation_goals ON goals;
DROP POLICY IF EXISTS student_isolation_certificates ON certificates;
DROP POLICY IF EXISTS student_isolation_checkpoints ON checkpoint_submissions;
DROP POLICY IF EXISTS student_isolation_chat ON chat_messages;
DROP POLICY IF EXISTS student_isolation_student_progress ON student_progress;
DROP POLICY IF EXISTS student_isolation_progress ON progress_events;
DROP POLICY IF EXISTS student_isolation_sessions ON student_sessions;

DROP POLICY IF EXISTS tenant_isolation_assessments ON assessments;
DROP POLICY IF EXISTS tenant_isolation_studentskills ON student_skills;
DROP POLICY IF EXISTS tenant_isolation_skillnodes ON skill_nodes;
DROP POLICY IF EXISTS tenant_isolation_skilltrees ON skill_trees;
DROP POLICY IF EXISTS tenant_isolation_insights ON insight_snapshots;
DROP POLICY IF EXISTS tenant_isolation_communications ON communications;
DROP POLICY IF EXISTS tenant_isolation_schedule ON schedule_blocks;
DROP POLICY IF EXISTS tenant_isolation_milestones ON milestones;
DROP POLICY IF EXISTS tenant_isolation_goals ON goals;
DROP POLICY IF EXISTS tenant_isolation_certificates ON certificates;
DROP POLICY IF EXISTS tenant_isolation_checkpoints ON checkpoint_submissions;
DROP POLICY IF EXISTS tenant_isolation_chat ON chat_messages;
DROP POLICY IF EXISTS tenant_isolation_progress ON student_progress;
DROP POLICY IF EXISTS tenant_isolation_events ON progress_events;
DROP POLICY IF EXISTS tenant_isolation_lessons ON course_lessons;
DROP POLICY IF EXISTS tenant_isolation_modules ON course_modules;
DROP POLICY IF EXISTS tenant_isolation_courses ON courses;
DROP POLICY IF EXISTS tenant_isolation_sessions ON student_sessions;
DROP POLICY IF EXISTS tenant_isolation ON students;

-- Disable RLS on all tables
ALTER TABLE assessments DISABLE ROW LEVEL SECURITY;
ALTER TABLE student_skills DISABLE ROW LEVEL SECURITY;
ALTER TABLE skill_nodes DISABLE ROW LEVEL SECURITY;
ALTER TABLE skill_trees DISABLE ROW LEVEL SECURITY;
ALTER TABLE insight_snapshots DISABLE ROW LEVEL SECURITY;
ALTER TABLE communications DISABLE ROW LEVEL SECURITY;
ALTER TABLE schedule_blocks DISABLE ROW LEVEL SECURITY;
ALTER TABLE milestones DISABLE ROW LEVEL SECURITY;
ALTER TABLE goals DISABLE ROW LEVEL SECURITY;
ALTER TABLE certificates DISABLE ROW LEVEL SECURITY;
ALTER TABLE checkpoint_submissions DISABLE ROW LEVEL SECURITY;
ALTER TABLE chat_messages DISABLE ROW LEVEL SECURITY;
ALTER TABLE student_progress DISABLE ROW LEVEL SECURITY;
ALTER TABLE progress_events DISABLE ROW LEVEL SECURITY;
ALTER TABLE course_lessons DISABLE ROW LEVEL SECURITY;
ALTER TABLE course_modules DISABLE ROW LEVEL SECURITY;
ALTER TABLE courses DISABLE ROW LEVEL SECURITY;
ALTER TABLE student_sessions DISABLE ROW LEVEL SECURITY;
ALTER TABLE students DISABLE ROW LEVEL SECURITY;

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS student_progress_mv;

-- Drop indexes (will be dropped with tables, but being explicit)
DROP INDEX IF EXISTS idx_assessments_method;
DROP INDEX IF EXISTS idx_assessments_tenant;
DROP INDEX IF EXISTS idx_assessments_node;
DROP INDEX IF EXISTS idx_assessments_student;

DROP INDEX IF EXISTS idx_studentskills_level;
DROP INDEX IF EXISTS idx_studentskills_tenant;
DROP INDEX IF EXISTS idx_studentskills_tree;
DROP INDEX IF EXISTS idx_studentskills_node;
DROP INDEX IF EXISTS idx_studentskills_student;

DROP INDEX IF EXISTS idx_skillnodes_tenant;
DROP INDEX IF EXISTS idx_skillnodes_tree;

DROP INDEX IF EXISTS idx_skilltrees_category;
DROP INDEX IF EXISTS idx_skilltrees_slug;
DROP INDEX IF EXISTS idx_skilltrees_tenant;

DROP INDEX IF EXISTS idx_insights_recorded;
DROP INDEX IF EXISTS idx_insights_tenant;
DROP INDEX IF EXISTS idx_insights_student;

DROP INDEX IF EXISTS idx_communications_status;
DROP INDEX IF EXISTS idx_communications_scheduled;
DROP INDEX IF EXISTS idx_communications_tenant;
DROP INDEX IF EXISTS idx_communications_goal;

DROP INDEX IF EXISTS idx_schedule_for;
DROP INDEX IF EXISTS idx_schedule_tenant;
DROP INDEX IF EXISTS idx_schedule_goal;

DROP INDEX IF EXISTS idx_milestones_due;
DROP INDEX IF EXISTS idx_milestones_status;
DROP INDEX IF EXISTS idx_milestones_tenant;
DROP INDEX IF EXISTS idx_milestones_goal;

DROP INDEX IF EXISTS idx_goals_target;
DROP INDEX IF EXISTS idx_goals_status;
DROP INDEX IF EXISTS idx_goals_tenant;
DROP INDEX IF EXISTS idx_goals_student;

DROP INDEX IF EXISTS idx_certificates_verification;
DROP INDEX IF EXISTS idx_certificates_course;
DROP INDEX IF EXISTS idx_certificates_tenant;
DROP INDEX IF EXISTS idx_certificates_student;

DROP INDEX IF EXISTS idx_checkpoints_status;
DROP INDEX IF EXISTS idx_checkpoints_lesson;
DROP INDEX IF EXISTS idx_checkpoints_tenant;
DROP INDEX IF EXISTS idx_checkpoints_student;

DROP INDEX IF EXISTS idx_chat_created;
DROP INDEX IF EXISTS idx_chat_session;
DROP INDEX IF EXISTS idx_chat_tenant;
DROP INDEX IF EXISTS idx_chat_student;

DROP INDEX IF EXISTS idx_usage_month;
DROP INDEX IF EXISTS idx_usage_student;
DROP INDEX IF EXISTS idx_usage_tenant;

DROP INDEX IF EXISTS idx_events_occurred;
DROP INDEX IF EXISTS idx_events_entity;
DROP INDEX IF EXISTS idx_events_type;
DROP INDEX IF EXISTS idx_events_tenant;
DROP INDEX IF EXISTS idx_events_student;

DROP INDEX IF EXISTS idx_lessons_tenant;
DROP INDEX IF EXISTS idx_lessons_course;
DROP INDEX IF EXISTS idx_lessons_module;

DROP INDEX IF EXISTS idx_modules_tenant;
DROP INDEX IF EXISTS idx_modules_course;

DROP INDEX IF EXISTS idx_courses_published;
DROP INDEX IF EXISTS idx_courses_category;
DROP INDEX IF EXISTS idx_courses_slug;
DROP INDEX IF EXISTS idx_courses_tenant;

DROP INDEX IF EXISTS idx_sessions_started;
DROP INDEX IF EXISTS idx_sessions_tenant;
DROP INDEX IF EXISTS idx_sessions_student;

DROP INDEX IF EXISTS idx_students_status;
DROP INDEX IF EXISTS idx_students_email;
DROP INDEX IF EXISTS idx_students_tenant;

DROP INDEX IF EXISTS idx_tenants_deleted;
DROP INDEX IF EXISTS idx_tenants_slug;

DROP INDEX IF EXISTS idx_progress_status;
DROP INDEX IF EXISTS idx_progress_tenant;
DROP INDEX IF EXISTS idx_progress_student;

-- Drop tables in reverse order of creation (due to foreign keys)
DROP TABLE IF EXISTS assessments;
DROP TABLE IF EXISTS student_skills;
DROP TABLE IF EXISTS skill_nodes;
DROP TABLE IF EXISTS skill_trees;
DROP TABLE IF EXISTS insight_snapshots;
DROP TABLE IF EXISTS communications;
DROP TABLE IF EXISTS schedule_blocks;
DROP TABLE IF EXISTS milestones;
DROP TABLE IF EXISTS goals;
DROP TABLE IF EXISTS certificates;
DROP TABLE IF EXISTS checkpoint_submissions;
DROP TABLE IF EXISTS ai_usage;
DROP TABLE IF EXISTS chat_messages;
DROP TABLE IF EXISTS student_progress;
DROP TABLE IF EXISTS progress_events;
DROP TABLE IF EXISTS course_lessons;
DROP TABLE IF EXISTS course_modules;
DROP TABLE IF EXISTS courses;
DROP TABLE IF EXISTS student_sessions;
DROP TABLE IF EXISTS students;
DROP TABLE IF EXISTS tenants;

-- Drop extensions
DROP EXTENSION IF EXISTS "pgcrypto";
DROP EXTENSION IF EXISTS "uuid-ossp";
