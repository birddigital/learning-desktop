-- ============================================================================
-- Learning Desktop Seed Data Rollback
-- Removes all seeded data
-- ============================================================================

-- Disable RLS for truncation
SET session_replication_role = 'replica';

-- Drop seeded data in reverse order
DELETE FROM schedule_blocks WHERE id LIKE '880e8400-e29b-41d4-a716-446655440%';
DELETE FROM milestones WHERE id LIKE '880e8400-e29b-41d4-a716-446655440%';
DELETE FROM goals WHERE id LIKE '880e8400-e29b-41d4-a716-446655440%';

DELETE FROM student_skills WHERE id LIKE '770e8400-e29b-41d4-a716-446655440%';

DELETE FROM skill_nodes WHERE id LIKE '660e8400-e29b-41d4-a716-44665544%' AND tree_id LIKE '660e8400-e29b-41d4-a716-44665544%';
DELETE FROM skill_trees WHERE id LIKE '660e8400-e29b-41d4-a716-44665544%';

DELETE FROM students WHERE id LIKE '550e8400-e29b-41d4-a716-446655440%';
DELETE FROM tenants WHERE id LIKE '550e8400-e29b-41d4-a716-446655440%';

-- Re-enable RLS
SET session_replication_role = 'origin';

SELECT 'Seed data removed successfully!' as status;
