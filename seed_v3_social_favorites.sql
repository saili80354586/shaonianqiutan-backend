-- ============================================
-- 少年球探 - 收藏数据 v3.1 (修正表结构)
-- ============================================

-- 分析师收藏球员/报告
INSERT INTO favorites (id, user_id, target_type, target_id, created_at) VALUES
(1, 30, 'player', 1, datetime('now', '-27 days')),
(2, 30, 'player', 2, datetime('now', '-17 days')),
(3, 33, 'player', 1, datetime('now', '-12 days')),
(4, 31, 'player', 3, datetime('now', '-22 days'));

-- 球探收藏球员
INSERT INTO favorites (id, user_id, target_type, target_id, created_at) VALUES
(5, 24, 'player', 1, datetime('now', '-28 days')),
(6, 24, 'player', 2, datetime('now', '-20 days')),
(7, 24, 'player', 3, datetime('now', '-15 days')),
(8, 25, 'player', 1, datetime('now', '-10 days')),
(9, 25, 'player', 5, datetime('now', '-8 days'));

-- 球员收藏内容
INSERT INTO favorites (id, user_id, target_type, target_id, created_at) VALUES
(10, 2002, 'player', 1, datetime('now', '-25 days')),
(11, 2001, 'player', 2, datetime('now', '-15 days'));

.print '收藏数据导入完成: ' || (SELECT COUNT(*) FROM favorites) || ' 条'
