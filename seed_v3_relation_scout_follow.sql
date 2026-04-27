-- ============================================
-- 少年球探 - 球探关注球员 v3.1 (修正表结构)
-- ============================================

-- 赵球探关注球员 (scout_id=1)
INSERT INTO scout_follow_players (id, scout_id, user_id, notes, followed_at, created_at) VALUES
(1, 1, 2001, '进攻意识强，潜力大', datetime('now', '-10 days'), datetime('now', '-10 days')),
(2, 1, 2002, '技术全面，视野开阔', datetime('now', '-7 days'), datetime('now', '-7 days')),
(3, 1, 2003, '身体素质好，防空能力强', datetime('now', '-5 days'), datetime('now', '-5 days')),
(4, 1, 2005, '左脚技术出色', datetime('now', '-3 days'), datetime('now', '-3 days'));

-- 陈球探关注球员 (scout_id=2)
INSERT INTO scout_follow_players (id, scout_id, user_id, notes, followed_at, created_at) VALUES
(5, 2, 2001, '值得关注的前锋苗子', datetime('now', '-8 days'), datetime('now', '-8 days')),
(6, 2, 2006, '中场调度能力不错', datetime('now', '-2 days'), datetime('now', '-2 days'));

.print '球探关注球员导入完成: ' || (SELECT COUNT(*) FROM scout_follow_players) || ' 条'
