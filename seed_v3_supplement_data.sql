-- ============================================
-- 少年球探 - 补充数据 v3.1 (修正表结构)
-- 执行时间: 2026-04-11
-- ============================================

-- 1. 补充 weekly_report_periods 数据（4条：2队 x 2周）
INSERT INTO weekly_report_periods (id, team_id, week_start, week_end, deadline, total_players, submitted_count, pending_count, reviewed_count, status, created_at, updated_at) VALUES
-- U12一队 上上周
(1, 1, date('now', '-14 days'), date('now', '-8 days'), datetime('now', '-1 days'), 4, 4, 0, 4, 'closed', datetime('now', '-7 days'), datetime('now', '-1 days')),
-- U12一队 上周
(2, 1, date('now', '-7 days'), date('now', '-1 days'), datetime('now', '+1 days'), 4, 2, 2, 1, 'active', datetime('now', '-6 days'), datetime('now')),
-- U12二队 上上周
(3, 2, date('now', '-14 days'), date('now', '-8 days'), datetime('now', '-1 days'), 4, 3, 1, 3, 'closed', datetime('now', '-7 days'), datetime('now', '-1 days')),
-- U12二队 上周
(4, 2, date('now', '-7 days'), date('now', '-1 days'), datetime('now', '+1 days'), 4, 1, 3, 0, 'active', datetime('now', '-6 days'), datetime('now'));

-- 2. 补充 weekly_reports 数据（为 U12二队 4名球员添加上周周报）
INSERT INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, training_count, training_participation, physical_status, mental_status, technical_performance, improvements, next_week_goals, review_status, review_comment, review_rating, reviewed_at, created_at, updated_at, deadline, submit_status) VALUES
-- 已提交待审核
(9, 2, 2005, 21, date('now', '-7 days'), date('now', '-1 days'), 2, 'full', 4, 4, '前锋跑位意识提升', '加强射门练习', '提升进球转化率', 'pending', NULL, NULL, NULL, datetime('now', '-2 days'), datetime('now', '-1 days'), datetime('now', '+1 days'), 'submitted'),
-- 草稿
(10, 2, 2006, 21, date('now', '-7 days'), date('now', '-1 days'), 2, 'partial', 3, 3, NULL, NULL, NULL, 'pending', NULL, NULL, NULL, datetime('now', '-3 days'), NULL, datetime('now', '+1 days'), 'draft'),
-- 草稿
(11, 2, 2007, 21, date('now', '-7 days'), date('now', '-1 days'), 2, 'full', 4, 3, NULL, NULL, NULL, 'pending', NULL, NULL, NULL, datetime('now', '-2 days'), NULL, datetime('now', '+1 days'), 'draft'),
-- 已提交
(12, 2, 2008, 21, date('now', '-7 days'), date('now', '-1 days'), 2, 'full', 3, 4, '扑救反应有所提升', '加强出击练习', '提升位置感', 'pending', NULL, NULL, NULL, datetime('now', '-1 days'), datetime('now', '-1 days'), datetime('now', '+1 days'), 'submitted');

-- 3. 补充 match_summaries 数据（第5条 - U12二队vs上海红队）
INSERT INTO match_summaries (id, team_id, match_name, match_date, opponent, our_score, opponent_score, match_result, status, player_ids, coach_id, created_at, updated_at) VALUES
(5, 2, '浦东联赛第六轮', datetime('now', '-5 days'), '上海红队', 2, 3, 'lose', 'completed', '[2005, 2006, 2007, 2008]', 21, datetime('now', '-5 days'), datetime('now', '-4 days'));

-- 4. 补充 physical_test_records 数据（activity_id=2 的球员已有记录，补充 activity_id=3）
INSERT INTO physical_test_records (activity_id, player_id, club_id, test_date, height, weight, bmi, sprint30m, sprint50m, sprint100m, agility_ladder, t_test, shuttle_run, standing_long_jump, vertical_jump, sit_and_reach, push_up, sit_up, plank, created_at, updated_at) VALUES
-- U12一队第二次体测（activity_id=2 但上面已有4条，再加2条凑够6人）
(2, 2001, 10, datetime('now', '-10 days'), 150, 42, 18.7, 5.9, 8.2, 15.0, 18.2, 11.0, 28.0, 180, 36, 8, 26, 32, 92, datetime('now', '-10 days'), datetime('now', '-10 days')),
(2, 2002, 10, datetime('now', '-10 days'), 148, 40, 18.3, 5.7, 8.0, 14.8, 17.8, 10.8, 27.5, 175, 34, 9, 28, 34, 90, datetime('now', '-10 days'), datetime('now', '-10 days'));

-- 5. 补充 user_social_stats 数据（缺失 ID 1 和 ID 10）
INSERT INTO user_social_stats (user_id, likes_received, favorites_received, comments_received, followers_count, following_count, login_streak, updated_at) VALUES
-- 管理员 (ID=1)
(1, 0, 0, 0, 5, 2, 10, datetime('now')),
-- 俱乐部管理员 (ID=10)
(10, 2, 1, 3, 12, 5, 15, datetime('now'));

-- 6. 导入成长档案数据（growth_records 表）
INSERT INTO growth_records (id, user_id, record_date, record_type, title, content, created_at) VALUES
-- 王小明的成长档案 (3条)
(1, 2001, datetime('now', '-60 days'), 'milestone', '首次正式比赛进球', '在U12联赛首轮比赛中攻入职业生涯首球，帮助球队3:1获胜', datetime('now', '-60 days')),
(2, 2001, datetime('now', '-30 days'), 'achievement', '获得月度最佳球员', '凭借出色表现当选俱乐部U12组别4月最佳球员', datetime('now', '-30 days')),
(3, 2001, datetime('now', '-7 days'), 'training', '完成进阶射门课程', '完成了为期一个月的前锋射门进阶训练课程，掌握了多种射门技巧', datetime('now', '-7 days')),
-- 李小强的成长档案 (2条)
(4, 2002, datetime('now', '-45 days'), 'training', '角球配合练习', '与前锋队友完成了20次角球配合练习，成功率提升至75%', datetime('now', '-45 days')),
(5, 2002, datetime('now', '-10 days'), 'match', '首次担任队长', '在俱乐部内部对抗赛中首次担任队长，展现出领袖气质', datetime('now', '-10 days')),
-- 张小刚的成长档案 (2条)
(6, 2003, datetime('now', '-50 days'), 'milestone', '位置转型成功', '从边后卫转型为中后卫，表现稳定', datetime('now', '-50 days')),
(7, 2003, datetime('now', '-15 days'), 'achievement', '头球能力提升', '通过专项训练，头球争顶成功率从60%提升至80%', datetime('now', '-15 days')),
-- 刘小军的成长档案 (1条)
(8, 2004, datetime('now', '-20 days'), 'training', '扑救反应训练', '完成了守门员专项反应训练，出击速度明显提升', datetime('now', '-20 days')),
-- U12二队球员成长档案 (各1条)
(9, 2005, datetime('now', '-12 days'), 'match', '首粒正式比赛进球', '在U12联赛中攻入首球，帮助球队4:2获胜', datetime('now', '-12 days')),
(10, 2006, datetime('now', '-8 days'), 'training', '中场调度练习', '完成了中场调度专项训练，大局观有所提升', datetime('now', '-8 days')),
(11, 2007, datetime('now', '-5 days'), 'milestone', '首次零封对手', '在比赛中帮助球队完成首次零封', datetime('now', '-5 days')),
(12, 2008, datetime('now', '-3 days'), 'training', '扑点训练', '完成了守门员扑点训练，成功率提升至60%', datetime('now', '-3 days'));

-- 7. 导入关注关系数据（follows 表）
INSERT INTO follows (id, follower_id, following_id, created_at) VALUES
-- 球员之间互相关注
(1, 2001, 2002, datetime('now', '-10 days')),
(2, 2001, 2003, datetime('now', '-8 days')),
(3, 2002, 2001, datetime('now', '-9 days')),
(4, 2002, 2004, datetime('now', '-5 days')),
(5, 2003, 2001, datetime('now', '-7 days')),
(6, 2005, 2001, datetime('now', '-6 days')),
(7, 2006, 2002, datetime('now', '-4 days')),
(8, 2007, 2003, datetime('now', '-3 days')),
(9, 2008, 2004, datetime('now', '-2 days')),
-- 分析师/球探关注球员
(10, 30, 2001, datetime('now', '-15 days')),
(11, 30, 2002, datetime('now', '-12 days')),
(12, 31, 2003, datetime('now', '-10 days')),
(13, 24, 2001, datetime('now', '-20 days')),
(14, 24, 2002, datetime('now', '-18 days')),
(15, 25, 2005, datetime('now', '-10 days'));

-- 输出统计
.print '=== 补充数据导入完成 ==='
.print 'weekly_report_periods: ' || (SELECT COUNT(*) FROM weekly_report_periods) || ' 条'
.print 'weekly_reports: ' || (SELECT COUNT(*) FROM weekly_reports) || ' 条'
.print 'match_summaries: ' || (SELECT COUNT(*) FROM match_summaries) || ' 条'
.print 'physical_test_records: ' || (SELECT COUNT(*) FROM physical_test_records) || ' 条'
.print 'user_social_stats: ' || (SELECT COUNT(*) FROM user_social_stats) || ' 条'
.print 'growth_records: ' || (SELECT COUNT(*) FROM growth_records) || ' 条'
.print 'follows: ' || (SELECT COUNT(*) FROM follows) || ' 条'