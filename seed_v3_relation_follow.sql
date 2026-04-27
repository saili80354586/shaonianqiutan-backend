-- ============================================
-- 少年球探 - 用户关注关系 v3.0
-- ============================================

-- 球员之间互相关注
INSERT INTO follows (id, follower_id, following_id, created_at) VALUES
(1, 2001, 2002, datetime('now', '-10 days')),
(2, 2001, 2003, datetime('now', '-8 days')),
(3, 2002, 2001, datetime('now', '-9 days')),
(4, 2002, 2004, datetime('now', '-5 days')),
(5, 2003, 2001, datetime('now', '-7 days')),
(6, 2005, 2001, datetime('now', '-6 days')),
(7, 2006, 2002, datetime('now', '-4 days')),
(8, 2007, 2003, datetime('now', '-3 days')),
(9, 2008, 2004, datetime('now', '-2 days'));

-- 分析师/球探关注球员
INSERT INTO follows (id, follower_id, following_id, created_at) VALUES
(10, 30, 2001, datetime('now', '-15 days')),
(11, 30, 2002, datetime('now', '-12 days')),
(12, 31, 2003, datetime('now', '-10 days')),
(13, 24, 2001, datetime('now', '-20 days')),
(14, 24, 2002, datetime('now', '-18 days')),
(15, 25, 2005, datetime('now', '-10 days'));

.print '用户关注关系导入完成: ' || (SELECT COUNT(*) FROM follows) || ' 条'
