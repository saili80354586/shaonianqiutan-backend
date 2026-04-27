-- ============================================
-- 少年球探 - 社交统计数据 v3.1 (修正表结构)
-- ============================================

-- 球员社交统计
INSERT INTO user_social_stats (user_id, likes_received, favorites_received, comments_received, followers_count, following_count, login_streak, last_login_date, updated_at) VALUES
(2001, 120, 15, 35, 25, 15, 7, datetime('now'), datetime('now')),
(2002, 80, 10, 22, 18, 12, 5, datetime('now'), datetime('now')),
(2003, 45, 8, 15, 12, 8, 3, datetime('now'), datetime('now')),
(2004, 30, 5, 10, 8, 5, 2, datetime('now'), datetime('now')),
(2005, 50, 6, 18, 10, 7, 4, datetime('now'), datetime('now')),
(2006, 25, 4, 8, 6, 4, 1, datetime('now'), datetime('now')),
(2007, 20, 3, 6, 5, 3, 1, datetime('now'), datetime('now')),
(2008, 18, 3, 5, 4, 3, 1, datetime('now'), datetime('now'));

-- 教练社交统计
INSERT INTO user_social_stats (user_id, likes_received, favorites_received, comments_received, followers_count, following_count, login_streak, last_login_date, updated_at) VALUES
(20, 150, 20, 45, 35, 20, 10, datetime('now'), datetime('now')),
(21, 80, 12, 25, 20, 12, 6, datetime('now'), datetime('now'));

-- 分析师/球探社交统计
INSERT INTO user_social_stats (user_id, likes_received, favorites_received, comments_received, followers_count, following_count, login_streak, last_login_date, updated_at) VALUES
(24, 200, 25, 60, 45, 30, 12, datetime('now'), datetime('now')),
(25, 120, 15, 35, 28, 18, 8, datetime('now'), datetime('now')),
(30, 100, 18, 30, 30, 15, 7, datetime('now'), datetime('now')),
(33, 75, 12, 20, 22, 10, 5, datetime('now'), datetime('now'));

.print '社交统计数据导入完成: ' || (SELECT COUNT(*) FROM user_social_stats) || ' 条'
