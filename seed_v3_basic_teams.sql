-- ============================================
-- 少年球探 - 球队数据 v3.1 (修正表结构)
-- ============================================

INSERT INTO teams (id, club_id, name, age_group, birth_year_start, birth_year_end, description, status, created_at, updated_at) VALUES
(1, 1, 'U12一队', 'U12', 2014, 2015, '上海绿地青训俱乐部U12年龄段精英队伍，曾获得多项青少年足球赛事奖项。', 'active', datetime('now'), datetime('now')),
(2, 1, 'U12二队', 'U12', 2014, 2015, '上海绿地青训俱乐部U12年龄段发展队伍，注重基础训练和人才梯队建设。', 'active', datetime('now'), datetime('now'));

.print '球队数据导入完成: ' || (SELECT COUNT(*) FROM teams) || ' 条'
