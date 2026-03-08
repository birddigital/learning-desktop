-- ============================================================================
-- Learning Desktop Seed Data
-- Populates the database with example skill trees and demo data
-- ============================================================================

-- Disable RLS for seeding
SET session_replication_role = 'replica';

-- ============================================================================
-- DEMO TENANT
-- ============================================================================

INSERT INTO tenants (id, slug, name, domain, settings) VALUES
    ('550e8400-e29b-41d4-a716-446655440001', 'demo', 'Demo Organization', 'demo.learning.dev', '{
        "features": {
            "voice_enabled": true,
            "ai_assessment": true,
            "skill_trees": true,
            "accountability": true
        },
        "limits": {
            "max_students": 100,
            "max_ai_calls_per_month": 10000
        }
    }'::jsonb);

-- ============================================================================
-- DEMO STUDENT
-- ============================================================================

INSERT INTO students (id, tenant_id, name, email, level, goals, interests, settings) VALUES
    ('550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001',
     'Alex Learner', 'alex@demo.learning.dev', 'beginner',
     '["Learn AI fundamentals", "Master prompt engineering", "Build AI applications"]'::jsonb,
     '["AI", "Programming", "Data Science"]'::jsonb,
     '{
        "notifications": {
            "email": true,
            "reminders": true,
            "weekly_summary": true
        },
        "voice": {
            "enabled": true,
            "auto_transcribe": true
        }
     }'::jsonb);

-- ============================================================================
-- SKILL TREE: Prompt Engineering
-- ============================================================================

INSERT INTO skill_trees (id, tenant_id, slug, title, description, icon, category, required_level, total_nodes, max_points) VALUES
    ('660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     'prompt-engineering', 'Prompt Engineering', 'Master the art of communicating with AI effectively',
     '💬', 'technical', 'beginner', 6, 950);

-- Prompt Engineering Nodes
INSERT INTO skill_nodes (id, tree_id, tenant_id, slug, title, description, icon, position_row, position_col, max_points, required_score, required_nodes) VALUES
    -- Basic Prompts (Entry point)
    ('660e8400-e29b-41d4-a716-446655440102', '660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     'basic-prompts', 'Basic Prompt Writing', 'Write clear, effective prompts for common tasks',
     '📝', 0, 0, 100, 0, '[]'::jsonb),

    -- Chain of Thought (requires basic-prompts)
    ('660e8400-e29b-41d4-a716-446655440103', '660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     'chain-of-thought', 'Chain of Thought', 'Guide AI through step-by-step reasoning',
     '🔗', 1, 0, 150, 50, '["660e8400-e29b-41d4-a716-446655440102"]'::jsonb),

    -- Few-Shot Prompting (requires basic-prompts)
    ('660e8400-e29b-41d4-a716-446655440104', '660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     'few-shot-prompting', 'Few-Shot Prompting', 'Give examples to improve AI accuracy',
     '🎯', 1, 1, 150, 50, '["660e8400-e29b-41d4-a716-446655440102"]'::jsonb),

    -- Advanced Techniques (requires both chain-of-thought and few-shot)
    ('660e8400-e29b-41d4-a716-446655440105', '660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     'advanced-techniques', 'Advanced Techniques', 'Role-playing, constraint setting, output formatting',
     '🚀', 2, 0, 200, 70, '["660e8400-e29b-41d4-a716-446655440103", "660e8400-e29b-41d4-a716-446655440104"]'::jsonb),

    -- Security & Safety (requires few-shot)
    ('660e8400-e29b-41d4-a716-446655440106', '660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     'prompt-injection', 'Security & Safety', 'Understand and prevent prompt injection attacks',
     '🔒', 2, 1, 150, 60, '["660e8400-e29b-41d4-a716-446655440104"]'::jsonb),

    -- Prompt Optimization (requires advanced techniques)
    ('660e8400-e29b-41d4-a716-446655440107', '660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     'prompt-optimization', 'Prompt Optimization', 'Iterate and refine prompts for maximum effectiveness',
     '⚡', 3, 0, 200, 80, '["660e8400-e29b-41d4-a716-446655440105"]'::jsonb);

-- ============================================================================
-- SKILL TREE: AI Concepts
-- ============================================================================

INSERT INTO skill_trees (id, tenant_id, slug, title, description, icon, category, required_level, total_nodes, max_points) VALUES
    ('660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     'ai-concepts', 'AI Concepts & Logic', 'Understand how AI actually works under the hood',
     '🧠', 'technical', 'beginner', 6, 1000);

-- AI Concepts Nodes
INSERT INTO skill_nodes (id, tree_id, tenant_id, slug, title, description, icon, position_row, position_col, max_points, required_score, required_nodes) VALUES
    -- How LLMs Work (Entry point)
    ('660e8400-e29b-41d4-a716-446655440202', '660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     'llm-basics', 'How LLMs Work', 'Tokens, embeddings, transformer architecture',
     '🧠', 0, 0, 150, 0, '[]'::jsonb),

    -- Embeddings (requires llm-basics)
    ('660e8400-e29b-41d4-a716-446655440203', '660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     'embeddings', 'Embeddings & Vectors', 'How AI represents meaning as numbers',
     '📊', 1, 0, 150, 50, '["660e8400-e29b-41d4-a716-446655440202"]'::jsonb),

    -- Attention (requires llm-basics)
    ('660e8400-e29b-41d4-a716-446655440204', '660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     'attention', 'Attention Mechanism', 'The revolutionary technique behind transformers',
     '👁️', 1, 1, 150, 60, '["660e8400-e29b-41d4-a716-446655440202"]'::jsonb),

    -- RAG Systems (requires embeddings)
    ('660e8400-e29b-41d4-a716-446655440205', '660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     'rag', 'RAG Systems', 'Retrieval Augmented Generation for custom knowledge',
     '📚', 2, 0, 200, 70, '["660e8400-e29b-41d4-a716-446655440203"]'::jsonb),

    -- Fine-Tuning (requires llm-basics)
    ('660e8400-e29b-41d4-a716-446655440206', '660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     'fine-tuning', 'Fine-Tuning Models', 'Customize models for specific tasks',
     '🎛️', 2, 1, 200, 70, '["660e8400-e29b-41d4-a716-446655440202"]'::jsonb),

    -- Multimodal (requires both rag and fine-tuning)
    ('660e8400-e29b-41d4-a716-446655440207', '660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     'multimodal', 'Multimodal AI', 'AI that sees, hears, and speaks',
     '🖼️', 3, 0, 150, 75, '["660e8400-e29b-41d4-a716-446655440205", "660e8400-e29b-41d4-a716-446655440206"]'::jsonb);

-- ============================================================================
-- SKILL TREE: Models & Data
-- ============================================================================

INSERT INTO skill_trees (id, tenant_id, slug, title, description, icon, category, required_level, total_nodes, max_points) VALUES
    ('660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     'models-data', 'Models & Data', 'Work with AI models and data effectively',
     '🗄️', 'technical', 'beginner', 6, 900);

-- Models & Data Nodes
INSERT INTO skill_nodes (id, tree_id, tenant_id, slug, title, description, icon, position_row, position_col, max_points, required_score, required_nodes) VALUES
    -- Model Selection (Entry point)
    ('660e8400-e29b-41d4-a716-446655440302', '660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     'model-selection', 'Model Selection', 'Choose the right model for the job',
     '🎯', 0, 0, 100, 0, '[]'::jsonb),

    -- Data Preparation (requires model-selection)
    ('660e8400-e29b-41d4-a716-446655440303', '660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     'data-prep', 'Data Preparation', 'Clean and format data for AI consumption',
     '📊', 1, 0, 150, 50, '["660e8400-e29b-41d4-a716-446655440302"]'::jsonb),

    -- Vector Databases (requires model-selection)
    ('660e8400-e29b-41d4-a716-446655440304', '660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     'vector-dbs', 'Vector Databases', 'Store and search embeddings efficiently',
     '🗄️', 1, 1, 150, 50, '["660e8400-e29b-41d4-a716-446655440302"]'::jsonb),

    -- API Integration (requires data-prep)
    ('660e8400-e29b-41d4-a716-446655440305', '660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     'api-integration', 'API Integration', 'Connect to AI APIs in production',
     '🔌', 2, 0, 200, 70, '["660e8400-e29b-41d4-a716-446655440303"]'::jsonb),

    -- Model Evaluation (requires vector-dbs)
    ('660e8400-e29b-41d4-a716-446655440306', '660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     'evaluation', 'Model Evaluation', 'Measure model performance and accuracy',
     '📈', 2, 1, 150, 70, '["660e8400-e29b-41d4-a716-446655440304"]'::jsonb),

    -- Deployment (requires both api-integration and evaluation)
    ('660e8400-e29b-41d4-a716-446655440307', '660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     'deployment', 'Deployment & Scaling', 'Put AI models into production',
     '🚀', 3, 0, 200, 80, '["660e8400-e29b-41d4-a716-446655440305", "660e8400-e29b-41d4-a716-446655440306"]'::jsonb);

-- ============================================================================
-- DEMO STUDENT SKILLS (initialize with some progress)
-- ============================================================================

-- Demo student has started the first node in each tree
INSERT INTO student_skills (id, student_id, node_id, tree_id, tenant_id, score, level, points_earned, max_points, unlocked, unlocked_at) VALUES
    -- Prompt Engineering - Basic Prompts (partially complete)
    ('770e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440102',
     '660e8400-e29b-41d4-a716-446655440101', '550e8400-e29b-41d4-a716-446655440001',
     35.0, 'apprentice', 35, 100, true, NOW()),

    -- AI Concepts - LLM Basics (just started)
    ('770e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440202',
     '660e8400-e29b-41d4-a716-446655440201', '550e8400-e29b-41d4-a716-446655440001',
     15.0, 'novice', 15, 150, true, NOW()),

    -- Models & Data - Model Selection (not started but unlocked)
    ('770e8400-e29b-41d4-a716-446655440003', '550e8400-e29b-41d4-a716-446655440002', '660e8400-e29b-41d4-a716-446655440302',
     '660e8400-e29b-41d4-a716-446655440301', '550e8400-e29b-41d4-a716-446655440001',
     5.0, 'novice', 5, 100, true, NOW());

-- ============================================================================
-- DEMO GOAL (Accountability example)
-- ============================================================================

INSERT INTO goals (id, student_id, tenant_id, title, description, category, target_date, confidence, status) VALUES
    ('880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440002', '550e8400-e29b-41d4-a716-446655440001',
     'Get AI Literate in 90 Days', 'Complete all three AI skill trees and build a working AI application',
     'learning', NOW() + INTERVAL '90 days', 0.75, 'active');

-- Demo milestones
INSERT INTO milestones (id, goal_id, tenant_id, title, description, due_date, status) VALUES
    ('880e8400-e29b-41d4-a716-446655440101', '880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001',
     'Complete Prompt Engineering', 'Finish all 6 nodes in the Prompt Engineering tree',
     NOW() + INTERVAL '30 days', 'in_progress'),

    ('880e8400-e29b-41d4-a716-446655440102', '880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001',
     'Complete AI Concepts', 'Finish all 6 nodes in the AI Concepts tree',
     NOW() + INTERVAL '60 days', 'pending'),

    ('880e8400-e29b-41d4-a716-446655440103', '880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001',
     'Build AI Application', 'Create a working app using what you learned',
     NOW() + INTERVAL '90 days', 'pending');

-- Demo schedule blocks
INSERT INTO schedule_blocks (id, goal_id, tenant_id, title, scheduled_for, duration_minutes, is_completed) VALUES
    ('880e8400-e29b-41d4-a716-446655440201', '880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001',
     'Prompt Practice', NOW() + INTERVAL '1 day', 30, false),

    ('880e8400-e29b-41d4-a716-446655440202', '880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001',
     'Chain of Thought Exercises', NOW() + INTERVAL '2 days', 45, false),

    ('880e8400-e29b-41d4-a716-446655440203', '880e8400-e29b-41d4-a716-446655440001', '550e8400-e29b-41d4-a716-446655440001',
     'Few-Shot Workshop', NOW() + INTERVAL '4 days', 60, false);

-- Re-enable RLS
SET session_replication_role = 'origin';

-- ============================================================================
-- COMPLETE
-- ============================================================================
SELECT 'Seed data loaded successfully!' as status;
