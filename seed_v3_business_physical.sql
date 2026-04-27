-- ============================================
-- 少年球探 - 体测数据 v3.1 (修正表结构)
-- ============================================

-- 体测活动1: U12一队体测 (已完成)
INSERT INTO physical_test_activities (id, club_id, name, description, start_date, end_date, location, status, created_by, created_at) VALUES
(1, 1, 'U12一队月度体能测试', '月度常规体能测试，包含速度、力量、灵敏度等项目', datetime('now', '-20 days'), datetime('now', '-20 days'), '浦东训练基地', 'completed', 10, datetime('now', '-21 days'));

-- U12一队体测记录
INSERT INTO physical_test_records (id, activity_id, player_id, club_id, test_date, height, weight, sprint50m, standing_long_jump, sit_and_reach, agility_ladder, created_at) VALUES
(1, 1, 2001, 1, datetime('now', '-20 days'), 148, 40, 8.2, 165, 8, 12.5, datetime('now', '-20 days')),
(2, 1, 2002, 1, datetime('now', '-20 days'), 145, 38, 7.9, 172, 10, 11.8, datetime('now', '-20 days')),
(3, 1, 2003, 1, datetime('now', '-20 days'), 152, 45, 8.0, 178, 7, 12.2, datetime('now', '-20 days')),
(4, 1, 2004, 1, datetime('now', '-20 days'), 150, 42, 8.5, 160, 9, 13.0, datetime('now', '-20 days'));

-- 体测活动2: U12一队体测 (进行中)
INSERT INTO physical_test_activities (id, club_id, name, description, start_date, location, status, created_by, created_at) VALUES
(2, 1, 'U12一队季度体能测试', '季度综合体能测试', datetime('now', '-2 days'), '浦东训练基地', 'in_progress', 10, datetime('now', '-5 days'));

INSERT INTO physical_test_records (id, activity_id, player_id, club_id, test_date, height, weight, created_at) VALUES
(5, 2, 2001, 1, datetime('now', '-2 days'), 148, 40, datetime('now', '-2 days')),
(6, 2, 2002, 1, datetime('now', '-2 days'), 145, 38, datetime('now', '-2 days'));

-- 体测活动3: U12二队体测 (已完成)
INSERT INTO physical_test_activities (id, club_id, name, description, start_date, end_date, location, status, created_by, created_at) VALUES
(3, 1, 'U12二队月度体能测试', 'U12二队首次正式体测', datetime('now', '-15 days'), datetime('now', '-15 days'), '浦东训练基地', 'completed', 10, datetime('now', '-16 days'));

INSERT INTO physical_test_records (id, activity_id, player_id, club_id, test_date, height, weight, sprint50m, standing_long_jump, sit_and_reach, agility_ladder, created_at) VALUES
(7, 3, 2005, 1, datetime('now', '-15 days'), 146, 39, 8.0, 162, 9, 12.0, datetime('now', '-15 days')),
(8, 3, 2006, 1, datetime('now', '-15 days'), 143, 37, 7.7, 158, 11, 11.2, datetime('now', '-15 days')),
(9, 3, 2007, 1, datetime('now', '-15 days'), 149, 43, 8.3, 170, 8, 12.8, datetime('now', '-15 days')),
(10, 3, 2008, 1, datetime('now', '-15 days'), 151, 41, 8.6, 155, 7, 13.5, datetime('now', '-15 days'));

.print '体测数据导入完成: ' || (SELECT COUNT(*) FROM physical_test_activities) || ' 个活动, ' || (SELECT COUNT(*) FROM physical_test_records) || ' 条记录'
