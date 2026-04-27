-- ============================================================
-- 少年球探 - 测试账号数据 Part 2 - 球队-球员关联
-- 生成时间: 2026-04-08
-- 使用已有的球队ID: 1-10
-- ============================================================

-- 清理旧的关联数据
DELETE FROM team_players;
DELETE FROM team_coaches;

-- 球队-球员关联 (使用已有的球队ID: 1-10)
INSERT INTO team_players (team_id, user_id, jersey_number, position, status, joined_at, created_at, updated_at) VALUES
-- 上海绿地 U12一队 (ID=1) - 4人
(1, 2001, '9', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2002, '8', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2003, '4', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2004, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 上海绿地 U12二队 (ID=2) - 2人
(2, 2005, '11', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(2, 2006, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 北京国安 (ID=3) - 2人
(3, 2007, '3', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(3, 2008, '7', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 广州恒大 (ID=4) - 2人
(4, 2009, '10', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(4, 2010, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 山东泰山 (ID=5) - 2人
(5, 2011, '8', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(5, 2012, '9', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 江苏苏宁 (ID=6) - 2人
(6, 2013, '4', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(6, 2014, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 成都蓉城 (ID=7) - 2人
(7, 2015, '10', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(7, 2016, '3', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 武汉三镇 (ID=8) - 2人
(8, 2017, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(8, 2018, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 浙江绿城 (ID=9) - 2人
(9, 2019, '11', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(9, 2020, '2', '后卫', 'active', datetime('now'), datetime('now'), datetime('now'));

-- 球队-教练关联
INSERT INTO team_coaches (team_id, user_id, role, status, joined_at, created_at, updated_at) VALUES
-- 上海绿地 2支球队 -> 王教练 (user_id=20)
(1, 20, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(2, 20, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 北京国安 + 广州恒大 -> 李教练 (user_id=21)
(3, 21, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(4, 21, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 山东泰山 + 江苏苏宁 -> 张教练 (user_id=22)
(5, 22, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(6, 22, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 成都蓉城 + 武汉三镇 + 浙江绿城 -> 刘教练 (user_id=23)
(7, 23, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(8, 23, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(9, 23, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now'));

SELECT '球队-球员关联完成!' as result;
