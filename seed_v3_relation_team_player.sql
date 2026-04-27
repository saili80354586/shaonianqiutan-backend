-- ============================================
-- 少年球探 - 球队-球员关联 v3.1 (修正表结构)
-- ============================================

-- U12一队球员 (4人)
INSERT INTO team_players (id, team_id, user_id, player_id, jersey_number, position, status, joined_at, created_at) VALUES
(1, 1, 2001, 1, '9', '前锋', 'active', datetime('now'), datetime('now')),
(2, 1, 2002, 2, '8', '中场', 'active', datetime('now'), datetime('now')),
(3, 1, 2003, 3, '4', '后卫', 'active', datetime('now'), datetime('now')),
(4, 1, 2004, 4, '1', '门将', 'active', datetime('now'), datetime('now'));

-- U12二队球员 (4人)
INSERT INTO team_players (id, team_id, user_id, player_id, jersey_number, position, status, joined_at, created_at) VALUES
(5, 2, 2005, 5, '11', '前锋', 'active', datetime('now'), datetime('now')),
(6, 2, 2006, 6, '6', '中场', 'active', datetime('now'), datetime('now')),
(7, 2, 2007, 7, '3', '后卫', 'active', datetime('now'), datetime('now')),
(8, 2, 2008, 8, '22', '门将', 'active', datetime('now'), datetime('now'));

.print '球队-球员关联导入完成: ' || (SELECT COUNT(*) FROM team_players) || ' 条'
