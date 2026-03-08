# Learning Desktop Database Migrations

This directory contains SQL migrations for the Learning Desktop database.

## Quick Start

```bash
# Build the migrate tool
go build -o bin/migrate ./cmd/migrate

# Run all pending migrations
./bin/migrate up

# Check migration status
./bin/migrate status

# Rollback last migration
./bin/migrate down

# Create a new migration
./bin/migrate create add_feature_table
```

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DATABASE_URL` | PostgreSQL connection string | `postgres://localhost/learning_desktop?sslmode=disable` |

## Migrations

| Version | Description | Up | Down |
|---------|-------------|-----|------|
| 001 | Initial schema (tables, RLS, functions) | ✓ | ✓ |
| 002 | Seed data (demo tenant, skill trees) | ✓ | ✓ |

## Schema Overview

### Core Tables

- `tenants` - Multi-tenant organizations
- `students` - Learners with profiles and goals
- `student_sessions` - Session tracking

### Learning Content

- `courses` - Course catalog
- `course_modules` - Course sections
- `course_lessons` - Individual lessons

### Progress Tracking

- `progress_events` - Event-sourced progress (source of truth)
- `student_progress` - Computed progress summaries
- `student_progress_mv` - Materialized view for fast queries

### Skill Trees

- `skill_trees` - Game-like skill categories
- `skill_nodes` - Individual skills with prerequisites
- `student_skills` - Student proficiency per node
- `assessments` - Skill evaluations (quiz, project, chat, voice)

### Accountability

- `goals` - Student goals with target dates
- `milestones` - Goal milestones
- `schedule_blocks` - Scheduled learning time
- `communications` - Reminders, escalations, updates

### AI & Chat

- `chat_messages` - Conversation history
- `ai_usage` - Token/cost tracking per month

## Row-Level Security (RLS)

All tenant-scoped tables have RLS enabled. The app must set context:

```sql
SELECT set_tenant_context('tenant-uuid');
SELECT set_student_context('student-uuid');
```

This is done automatically by the repository layer.

## Development

### Creating New Migrations

```bash
./bin/migrate create my_new_feature
# Creates: 003_my_new_feature.up.sql
#          003_my_new_feature.down.sql
```

### Migration File Format

**Up migration** (`001_name.up.sql`):
```sql
-- Migration 1: My Feature
-- 2024-01-01 12:00:00

CREATE TABLE my_feature (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    -- ...
);
```

**Down migration** (`001_name.down.sql`):
```sql
-- Rollback migration 1: My Feature

DROP TABLE IF EXISTS my_feature;
```

### Rules

1. **Always create both up and down** - Must be able to rollback
2. **Use IF EXISTS** - For safer rollbacks
3. **No data loss** - Down migration shouldn't delete user data
4. **Test locally first** - Run up/down before committing

## Production

### Backup First

```bash
pg_dump learning_desktop > backup_$(date +%Y%m%d).sql
```

### Run Migrations

```bash
# Check status first
./bin/migrate status

# Run pending migrations
./bin/migrate up

# Verify
./bin/migrate status
```

### Rollback if Needed

```bash
# Rollback last migration
./bin/migrate down

# Or rollback and re-apply (useful for fixes)
./bin/migrate redo
```

## Troubleshooting

### "Connection refused"

Make sure PostgreSQL is running:
```bash
brew services start postgresql
# or
docker start postgres
```

### "Migration already applied"

Check status:
```bash
./bin/migrate status
```

If stuck, manually update the version:
```sql
-- Check current version
SELECT * FROM schema_migrations;

-- Fix if needed
UPDATE schema_migrations SET version = N;
```

### "RLS policy violation"

Set tenant context:
```sql
SELECT set_tenant_context('your-tenant-id');
```

## Database Setup (First Time)

```bash
# Create database
createdb learning_desktop

# Run migrations
./bin/migrate up

# (Optional) Load seed data
./bin/migrate up  # Run again for seed migration
```
