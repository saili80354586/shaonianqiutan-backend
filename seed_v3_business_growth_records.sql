-- ============================================
-- 少年球探 - 成长档案数据 v3.0
-- ============================================

-- 王小明的成长档案 (3条)
INSERT INTO growth_records (id, user_id, record_date, record_type, title, content, created_at) VALUES
(1, 2001, datetime('now', '-60 days'), 'milestone', '首次正式比赛进球', '在U12联赛首轮比赛中攻入职业生涯首球，帮助球队3:1获胜', datetime('now', '-60 days')),
(2, 2001, datetime('now', '-30 days'), 'achievement', '获得月度最佳球员', '凭借出色表现当选俱乐部U12组别4月最佳球员', datetime('now', '-30 days')),
(3, 2001, datetime('now', '-7 days'), 'training', '完成进阶射门课程', '完成了为期一个月的前锋射门进阶训练课程，掌握了多种射门技巧', datetime('now', '-7 days'));

-- 李小强的成长档案 (2条)
INSERT INTO growth_records (id, user_id, record_date, record_type, title, content, created_at) VALUES
(4, 2002, datetime('now', '-45 days'), 'training', '角球配合练习', '与前锋队友完成了20次角球配合练习，成功率提升至75%', datetime('now', '-45 days')),
(5, 2002, datetime('now', '-10 days'), 'match', '首次担任队长', '在俱乐部内部对抗赛中首次担任队长，展现出领袖气质', datetime('now', '-10 days'));

-- 张小刚的成长档案 (2条)
INSERT INTO growth_records (id, user_id, record_date, record_type, title, content, created_at) VALUES
(6, 2003, datetime('now', '-50 days'), 'milestone', '位置转型成功', '从边后卫转型为中后卫，表现稳定', datetime('now', '-50 days')),
(7, 2003, datetime('now', '-15 days'), 'achievement', '头球能力提升', '通过专项训练，头球争顶成功率从60%提升至80%', datetime('now', '-15 days'));

-- 刘小军的成长档案 (1条)
INSERT INTO growth_records (id, user_id, record_date, record_type, title, content, created_at) VALUES
(8, 2004, datetime('now', '-20 days'), 'training', '扑救反应训练', '完成了守门员专项反应训练，出击速度明显提升', datetime('now', '-20 days'));

-- U12二队球员成长档案 (各1条)
INSERT INTO growth_records (id, user_id, record_date, record_type, title, content, created_at) VALUES
(9, 2005, datetime('now', '-12 days'), 'match', '首粒正式比赛进球', '在U12联赛中攻入首球，帮助球队4:2获胜', datetime('now', '-12 days')),
(10, 2006, datetime('now', '-8 days'), 'training', '中场调度练习', '完成了中场调度专项训练，大局观有所提升', datetime('now', '-8 days')),
(11, 2007, datetime('now', '-5 days'), 'milestone', '首次零封对手', '在比赛中帮助球队完成首次零封', datetime('now', '-5 days')),
(12, 2008, datetime('now', '-3 days'), 'training', '扑点训练', '完成了守门员扑点训练，成功率提升至60%', datetime('now', '-3 days'));

.print '成长档案导入完成: ' || (SELECT COUNT(*) FROM growth_records) || ' 条'
