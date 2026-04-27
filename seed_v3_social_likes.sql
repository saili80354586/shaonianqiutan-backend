-- ============================================
-- 少年球探 - 点赞数据 v3.1 (修正表结构)
-- ============================================

-- 王小明的动态被点赞
INSERT INTO likes (id, user_id, target_type, target_id, created_at) VALUES
(1, 30, 'player', 2001, datetime('now', '-25 days')),
(2, 31, 'player', 2001, datetime('now', '-25 days')),
(3, 24, 'player', 2001, datetime('now', '-24 days')),
(4, 2002, 'player', 2001, datetime('now', '-23 days')),
(5, 2003, 'player', 2001, datetime('now', '-18 days')),
(6, 30, 'player', 2001, datetime('now', '-17 days')),
(7, 33, 'player', 2001, datetime('now', '-12 days')),
(8, 24, 'player', 2001, datetime('now', '-11 days')),
(9, 2005, 'player', 2001, datetime('now', '-8 days')),
(10, 30, 'report', 1, datetime('now', '-27 days'));

-- 李小强被点赞
INSERT INTO likes (id, user_id, target_type, target_id, created_at) VALUES
(11, 30, 'player', 2002, datetime('now', '-20 days')),
(12, 20, 'player', 2002, datetime('now', '-19 days')),
(13, 2001, 'player', 2002, datetime('now', '-15 days')),
(14, 24, 'player', 2002, datetime('now', '-14 days')),
(15, 2003, 'report', 3, datetime('now', '-17 days'));

-- 张小刚被点赞
INSERT INTO likes (id, user_id, target_type, target_id, created_at) VALUES
(16, 31, 'player', 2003, datetime('now', '-18 days')),
(17, 20, 'player', 2003, datetime('now', '-17 days')),
(18, 2002, 'report', 4, datetime('now', '-22 days'));

-- 球员互赞
INSERT INTO likes (id, user_id, target_type, target_id, created_at) VALUES
(19, 2001, 'player', 2002, datetime('now', '-19 days')),
(20, 2002, 'player', 2001, datetime('now', '-10 days')),
(21, 2001, 'player', 2003, datetime('now', '-16 days')),
(22, 2003, 'player', 2001, datetime('now', '-7 days')),
(23, 2005, 'player', 2006, datetime('now', '-6 days'));

.print '点赞数据导入完成: ' || (SELECT COUNT(*) FROM likes) || ' 条'
