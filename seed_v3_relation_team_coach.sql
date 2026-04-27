-- ============================================
-- 少年球探 - 球队-教练关联 v3.1 (修正表结构)
-- ============================================

-- 王教练: U12一队主教练 + U12二队助理教练
INSERT INTO team_coaches (id, team_id, user_id, coach_id, role, is_admin, status, joined_at, created_at) VALUES
(1, 1, 20, 1, 'head_coach', 1, 'active', datetime('now'), datetime('now')),
(3, 2, 20, 1, 'assistant', 0, 'active', datetime('now'), datetime('now'));

-- 李教练: U12二队主教练
INSERT INTO team_coaches (id, team_id, user_id, coach_id, role, is_admin, status, joined_at, created_at) VALUES
(2, 2, 21, 2, 'head_coach', 1, 'active', datetime('now'), datetime('now'));

.print '球队-教练关联导入完成: ' || (SELECT COUNT(*) FROM team_coaches) || ' 条'
