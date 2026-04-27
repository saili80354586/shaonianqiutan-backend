-- ============================================
-- 少年球探 - 教练关注球员 v3.1 (修正表结构)
-- ============================================

-- 王教练关注球员 (coach_id=1)
INSERT INTO coach_follow_players (id, coach_id, user_id, notes, followed_at, created_at) VALUES
(1, 1, 2001, '进攻意识强，潜力大', datetime('now', '-5 days'), datetime('now', '-5 days')),
(2, 1, 2002, '技术全面，视野开阔', datetime('now', '-3 days'), datetime('now', '-3 days')),
(3, 1, 2003, '身体素质好，防空能力强', datetime('now', '-1 days'), datetime('now', '-1 days'));

-- 李教练关注球员 (coach_id=2)
INSERT INTO coach_follow_players (id, coach_id, user_id, notes, followed_at, created_at) VALUES
(4, 2, 2005, '左脚技术出色', datetime('now', '-2 days'), datetime('now', '-2 days')),
(5, 2, 2006, '中场调度能力不错', datetime('now', '-1 days'), datetime('now', '-1 days'));

.print '教练关注球员导入完成: ' || (SELECT COUNT(*) FROM coach_follow_players) || ' 条'
