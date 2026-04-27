-- ============================================
-- 少年球探 - 评论数据 v3.1 (修正表结构)
-- ============================================

-- 球员被评论
INSERT INTO comments (id, user_id, target_type, target_id, content, likes_count, created_at) VALUES
(1, 20, 'player', 2001, '进球技术分析到位，继续保持！', 5, datetime('now', '-27 days')),
(2, 30, 'player', 2001, '感谢认可，期待你更出色的表现', 3, datetime('now', '-26 days')),
(3, 20, 'player', 2002, '中场调度能力评估准确', 4, datetime('now', '-17 days')),
(4, 24, 'player', 2001, '综合能力很强，前途无量', 6, datetime('now', '-12 days')),
(5, 31, 'player', 2003, '头球能力分析很专业', 4, datetime('now', '-22 days')),
(6, 2002, 'player', 2001, '太厉害了，我的偶像！', 8, datetime('now', '-24 days')),
(7, 2003, 'player', 2001, '向你学习！', 3, datetime('now', '-23 days')),
(8, 2001, 'player', 2002, '传球越来越准了', 5, datetime('now', '-19 days')),
(9, 2001, 'player', 2003, '防守越来越稳了', 4, datetime('now', '-17 days')),
(10, 2005, 'player', 2001, '前锋的跑位真飘逸', 6, datetime('now', '-10 days')),
(11, 2006, 'player', 2002, '中场大师！', 4, datetime('now', '-14 days')),
(12, 30, 'player', 2001, '继续保持这个状态', 3, datetime('now', '-8 days')),
(13, 24, 'player', 2001, '值得关注的前锋', 5, datetime('now', '-7 days'));

-- 报告评论
INSERT INTO comments (id, user_id, target_type, target_id, content, likes_count, created_at) VALUES
(14, 20, 'report', 1, '报告分析很到位！', 3, datetime('now', '-27 days')),
(15, 2002, 'report', 3, '学习了这篇报告很有收获', 4, datetime('now', '-17 days')),
(16, 30, 'report', 4, '专业的分析！', 2, datetime('now', '-22 days'));

.print '评论数据导入完成: ' || (SELECT COUNT(*) FROM comments) || ' 条'
