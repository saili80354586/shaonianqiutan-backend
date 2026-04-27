-- ============================================
-- 少年球探 - 俱乐部-球员关联 v3.2 (修正字段)
-- ============================================

-- 上海绿地俱乐部 (club_id=1) 的所有球员
INSERT INTO club_players (id, club_id, user_id, join_date, age_group, position, status, created_at) VALUES
(1, 1, 2001, datetime('now'), 'U12', '前锋', 'active', datetime('now')),
(2, 1, 2002, datetime('now'), 'U12', '中场', 'active', datetime('now')),
(3, 1, 2003, datetime('now'), 'U12', '后卫', 'active', datetime('now')),
(4, 1, 2004, datetime('now'), 'U12', '门将', 'active', datetime('now')),
(5, 1, 2005, datetime('now'), 'U12', '前锋', 'active', datetime('now')),
(6, 1, 2006, datetime('now'), 'U12', '中场', 'active', datetime('now')),
(7, 1, 2007, datetime('now'), 'U12', '后卫', 'active', datetime('now')),
(8, 1, 2008, datetime('now'), 'U12', '门将', 'active', datetime('now'));

.print '俱乐部-球员关联导入完成: ' || (SELECT COUNT(*) FROM club_players) || ' 条'
