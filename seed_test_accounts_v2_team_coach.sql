-- ============================================================
-- 少年球探 - 测试账号数据 Part 2b - 球队-教练关联
-- 生成时间: 2026-04-08
-- coaches表: id=5->user_id=20, id=6->user_id=21, id=7->user_id=22, id=8->user_id=23
-- ============================================================

-- 清理旧关联
DELETE FROM team_players;
DELETE FROM team_coaches;

-- 球队-球员关联 (player_id = players.id, user_id = users.id)
INSERT INTO team_players (team_id, user_id, player_id, jersey_number, position, status, joined_at, created_at, updated_at) VALUES
-- 上海绿地 U12一队 (ID=1)
(1, 2001, 201, '9', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2002, 202, '8', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2003, 203, '4', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2004, 204, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 上海绿地 U12二队 (ID=2)
(2, 2005, 205, '11', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(2, 2006, 206, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 北京国安 (ID=3)
(3, 2007, 207, '3', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(3, 2008, 208, '7', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 广州恒大 (ID=4)
(4, 2009, 209, '10', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(4, 2010, 210, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 山东泰山 (ID=5)
(5, 2011, 211, '8', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(5, 2012, 212, '9', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 江苏苏宁 (ID=6)
(6, 2013, 213, '4', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(6, 2014, 214, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 成都蓉城 (ID=7)
(7, 2015, 215, '10', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(7, 2016, 216, '3', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 武汉三镇 (ID=8)
(8, 2017, 217, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(8, 2018, 218, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 浙江绿城 (ID=9)
(9, 2019, 219, '11', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(9, 2020, 220, '2', '后卫', 'active', datetime('now'), datetime('now'), datetime('now'));

-- 球队-教练关联
-- coach_id 使用 coaches.id (5=王教练, 6=李教练, 7=张教练, 8=刘教练)
INSERT INTO team_coaches (team_id, user_id, coach_id, role, status, joined_at, created_at, updated_at) VALUES
-- 上海绿地 2支球队 -> 王教练 (coach_id=5, user_id=20)
(1, 20, 5, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(2, 20, 5, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 北京国安 + 广州恒大 -> 李教练 (coach_id=6, user_id=21)
(3, 21, 6, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(4, 21, 6, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 山东泰山 + 江苏苏宁 -> 张教练 (coach_id=7, user_id=22)
(5, 22, 7, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(6, 22, 7, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 成都蓉城 + 武汉三镇 + 浙江绿城 -> 刘教练 (coach_id=8, user_id=23)
(7, 23, 8, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(8, 23, 8, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(9, 23, 8, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now'));

SELECT '球队-球员-教练关联完成!' as result;
