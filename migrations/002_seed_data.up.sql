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
-- SKILL TREE: Character & Manhood
-- ============================================================================

INSERT INTO skill_trees (id, tenant_id, slug, title, description, icon, category, required_level, total_nodes, max_points) VALUES
    ('660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'character-manhood', 'Character & Manhood', 'Build the foundation of who you are. Your word, your actions, your standards.',
     '🦁', 'soft_skills', 'beginner', 9, 1800);

-- Character & Manhood Nodes
INSERT INTO skill_nodes (id, tree_id, tenant_id, slug, title, description, icon, position_row, position_col, max_points, required_score, required_nodes) VALUES
    -- Integrity (Entry point)
    ('660e8400-e29b-41d4-a716-446655440402', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'integrity', 'Integrity', 'Your word is your bond. You do what you say, even when no one is watching.',
     '🤝', 0, 0, 200, 0, '[]'::jsonb),

    -- Ownership (requires integrity)
    ('660e8400-e29b-41d4-a716-446655440403', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'responsibility', 'Ownership', 'You own your mistakes. You own your choices. No excuses.',
     '💪', 1, 0, 200, 50, '["660e8400-e29b-41d4-a716-446655440402"]'::jsonb),

    -- Self-Discipline (requires integrity)
    ('660e8400-e29b-41d4-a716-446655440404', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'discipline', 'Self-Discipline', 'You do what needs to be done, even when you don''t feel like it.',
     '⚔️', 1, 1, 200, 50, '["660e8400-e29b-41d4-a716-446655440402"]'::jsonb),

    -- Respect (requires ownership)
    ('660e8400-e29b-41d4-a716-446655440405', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'respect', 'Respect', 'Give it to earn it. Self-respect first. Then others.',
     '👊', 2, 0, 150, 60, '["660e8400-e29b-41d4-a716-446655440403"]'::jsonb),

    -- Resilience (requires discipline)
    ('660e8400-e29b-41d4-a716-446655440406', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'resilience', 'Resilience', 'Fall down 7 times, stand up 8. Adapt and overcome.',
     '🔥', 2, 1, 200, 60, '["660e8400-e29b-41d4-a716-446655440404"]'::jsonb),

    -- Leadership (requires respect and resilience)
    ('660e8400-e29b-41d4-a716-446655440407', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'leadership', 'Leadership', 'Others follow because they want to, not because they have to.',
     '👑', 3, 0, 250, 75, '["660e8400-e29b-41d4-a716-446655440405", "660e8400-e29b-41d4-a716-446655440406"]'::jsonb),

    -- Purpose (requires discipline and resilience)
    ('660e8400-e29b-41d4-a716-446655440408', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'purpose', 'Purpose', 'Know why you''re here. What you''re building. Who you''re becoming.',
     '🎯', 3, 1, 200, 75, '["660e8400-e29b-41d4-a716-446655440404", "660e8400-e29b-41d4-a716-446655440406"]'::jsonb),

    -- Brotherhood (requires leadership)
    ('660e8400-e29b-41d4-a716-446655440409', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'brotherhood', 'Brotherhood', 'Build your circle. Lift others as you climb.',
     '🤝', 4, 0, 200, 80, '["660e8400-e29b-41d4-a716-446655440407"]'::jsonb),

    -- Legacy (requires purpose and brotherhood)
    ('660e8400-e29b-41d4-a716-446655440410', '660e8400-e29b-41d4-a716-446655440401', '550e8400-e29b-41d4-a716-446655440001',
     'legacy', 'Legacy Thinking', 'Make decisions today that your future self will thank you for.',
     '🏛️', 4, 1, 250, 85, '["660e8400-e29b-41d4-a716-446655440408", "660e8400-e29b-41d4-a716-446655440409"]'::jsonb);

-- ============================================================================
-- SKILL TREE: Student Skills
-- ============================================================================

INSERT INTO skill_trees (id, tenant_id, slug, title, description, icon, category, required_level, total_nodes, max_points) VALUES
    ('660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'student-skills', 'Student Skills', 'Master the art of learning. Study smarter, not harder.',
     '📚', 'soft_skills', 'beginner', 9, 1550);

-- Student Skills Nodes
INSERT INTO skill_nodes (id, tree_id, tenant_id, slug, title, description, icon, position_row, position_col, max_points, required_score, required_nodes) VALUES
    -- Deep Focus (Entry point)
    ('660e8400-e29b-41d4-a716-446655440502', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'focus', 'Deep Focus', 'Lock in. Eliminate distractions. Do deep work.',
     '🎯', 0, 0, 150, 0, '[]'::jsonb),

    -- Note Taking (requires focus)
    ('660e8400-e29b-41d4-a716-446655440503', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'note-taking', 'Note Taking', 'Capture what matters. Organize knowledge. Recall anything.',
     '📝', 1, 0, 150, 50, '["660e8400-e29b-41d4-a716-446655440502"]'::jsonb),

    -- Time Management (requires focus)
    ('660e8400-e29b-41d4-a716-446655440504', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'time-management', 'Time Management', 'Your time is your life. Spend it intentionally.',
     '⏰', 1, 1, 150, 50, '["660e8400-e29b-41d4-a716-446655440502"]'::jsonb),

    -- Strategic Reading (requires note-taking)
    ('660e8400-e29b-41d4-a716-446655440505', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'reading-strategy', 'Strategic Reading', 'Read faster. Remember more. Extract key insights.',
     '📖', 2, 0, 150, 60, '["660e8400-e29b-41d4-a716-446655440503"]'::jsonb),

    -- Memory Systems (requires note-taking)
    ('660e8400-e29b-41d4-a716-446655440506', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'memory-techniques', 'Memory Systems', 'Never forget what you learn. Build a memory palace.',
     '🧠', 2, 1, 150, 60, '["660e8400-e29b-41d4-a716-446655440503"]'::jsonb),

    -- Test Mastery (requires reading and memory)
    ('660e8400-e29b-41d4-a716-446655440507', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'test-taking', 'Test Mastery', 'Perform when it counts. Manage anxiety. Show what you know.',
     '✅', 3, 0, 150, 70, '["660e8400-e29b-41d4-a716-446655440505", "660e8400-e29b-41d4-a716-446655440506"]'::jsonb),

    -- Research Skills (requires reading)
    ('660e8400-e29b-41d4-a716-446655440508', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'research', 'Research Skills', 'Find truth. Verify sources. Build knowledge from evidence.',
     '🔍', 3, 1, 150, 70, '["660e8400-e29b-41d4-a716-446655440505"]'::jsonb),

    -- Teaching Others (requires test mastery)
    ('660e8400-e29b-41d4-a716-446655440509', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'teaching', 'Teaching Others', 'The best way to learn is to teach. Master by explaining.',
     '👨‍🏫', 4, 0, 200, 75, '["660e8400-e29b-41d4-a716-446655440507"]'::jsonb),

    -- Metacognition (requires teaching)
    ('660e8400-e29b-41d4-a716-446655440510', '660e8400-e29b-41d4-a716-446655440501', '550e8400-e29b-41d4-a716-446655440001',
     'metacognition', 'Metacognition', 'Think about your thinking. Know how you learn. Optimize.',
     '🔮', 4, 1, 200, 80, '["660e8400-e29b-41d4-a716-446655440509"]'::jsonb);

-- ============================================================================
-- SKILL TREE: Entrepreneurship
-- ============================================================================

INSERT INTO skill_trees (id, tenant_id, slug, title, description, icon, category, required_level, total_nodes, max_points) VALUES
    ('660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'entrepreneurship', 'Entrepreneurship', 'Build businesses. Create value. Own your economic future.',
     '💼', 'career', 'beginner', 9, 1600);

-- Entrepreneurship Nodes
INSERT INTO skill_nodes (id, tree_id, tenant_id, slug, title, description, icon, position_row, position_col, max_points, required_score, required_nodes) VALUES
    -- See Opportunities (Entry point)
    ('660e8400-e29b-41d4-a716-446655440602', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'opportunity-spotting', 'See Opportunities', 'Problems are opportunities in disguise. Train your eye.',
     '👁️', 0, 0, 150, 0, '[]'::jsonb),

    -- Create Value (requires opportunity)
    ('660e8400-e29b-41d4-a716-446655440603', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'value-creation', 'Create Value', 'Solve real problems. Build things people actually want.',
     '🔨', 1, 0, 150, 50, '["660e8400-e29b-41d4-a716-446655440602"]'::jsonb),

    -- Money Mindset (requires opportunity)
    ('660e8400-e29b-41d4-a716-446655440604', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'money-mindset', 'Money Mindset', 'Understand money. Make it work for you. Build wealth.',
     '💰', 1, 1, 150, 50, '["660e8400-e29b-41d4-a716-446655440602"]'::jsonb),

    -- Sales (requires value creation)
    ('660e8400-e29b-41d4-a716-446655440605', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'sales', 'Sales', 'Communication is currency. Persuade authentically. Close deals.',
     '🤝', 2, 0, 200, 60, '["660e8400-e29b-41d4-a716-446655440603"]'::jsonb),

    -- Marketing (requires value creation)
    ('660e8400-e29b-41d4-a716-446655440606', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'marketing', 'Marketing', 'Tell your story. Reach your people. Build an audience.',
     '📢', 2, 1, 150, 60, '["660e8400-e29b-41d4-a716-446655440603"]'::jsonb),

    -- Financial Literacy (requires sales and money)
    ('660e8400-e29b-41d4-a716-446655440607', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'financial-literacy', 'Financial Literacy', 'Read numbers. Make decisions based on data, not hope.',
     '📊', 3, 0, 150, 70, '["660e8400-e29b-41d4-a716-446655440605", "660e8400-e29b-41d4-a716-446655440604"]'::jsonb),

    -- Legal Basics (requires marketing)
    ('660e8400-e29b-41d4-a716-446655440608', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'legal-basics', 'Legal Foundation', 'Protect yourself. Understand contracts. Entity structure.',
     '⚖️', 3, 1, 150, 70, '["660e8400-e29b-41d4-a716-446655440606"]'::jsonb),

    -- Team Building (requires financial and legal)
    ('660e8400-e29b-41d4-a716-446655440609', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'team-building', 'Build Teams', 'You can''t do it alone. Hire right. Delegate well. Lead.',
     '👥', 4, 0, 200, 75, '["660e8400-e29b-41d4-a716-446655440607", "660e8400-e29b-41d4-a716-446655440608"]'::jsonb),

    -- Scaling (requires team)
    ('660e8400-e29b-41d4-a716-446655440610', '660e8400-e29b-41d4-a716-446655440601', '550e8400-e29b-41d4-a716-446655440001',
     'scaling', 'Scale Up', 'Turn what works into systems. Grow without breaking.',
     '📈', 4, 1, 200, 80, '["660e8400-e29b-41d4-a716-446655440609"]'::jsonb);

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
