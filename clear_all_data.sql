-- ============================================
-- 少年球探 - 数据清理脚本 v1.1 (修正表名)
-- 执行前请务必备份数据库！
-- ============================================

.print '============================================'
.print '开始清理数据...'
.print '============================================'

-- 禁用外键检查
PRAGMA foreign_keys = OFF;

-- 1. 社交相关表
DELETE FROM notifications;
DELETE FROM comments;
DELETE FROM favorites;
DELETE FROM likes;
DELETE FROM user_social_achievements;
DELETE FROM social_achievements;
DELETE FROM user_social_stats;

-- 2. 关注关系表
DELETE FROM scout_follow_players;
DELETE FROM coach_follow_players;

-- 3. 业务数据表
DELETE FROM training_notes;
DELETE FROM scout_reports;
DELETE FROM physical_test_records;
DELETE FROM physical_test_reports;
DELETE FROM physical_test_template_customs;
DELETE FROM physical_test_activities;
DELETE FROM match_summaries;
DELETE FROM weekly_report_periods;
DELETE FROM weekly_reports;
DELETE FROM reports;
DELETE FROM club_orders;
DELETE FROM orders;

-- 4. 关联关系表
DELETE FROM club_players;
DELETE FROM team_coaches;
DELETE FROM team_players;
DELETE FROM team_invitations;
DELETE FROM scout_tasks;

-- 5. 主页配置表
DELETE FROM team_homes;
DELETE FROM club_homes;

-- 6. 基础数据表
DELETE FROM scouts;
DELETE FROM analyst_applications;
DELETE FROM analysts;
DELETE FROM coaches;
DELETE FROM teams;
DELETE FROM clubs;
DELETE FROM players;
DELETE FROM sms_codes;
DELETE FROM users;

-- 启用外键检查
PRAGMA foreign_keys = ON;

-- 输出清理结果
.print ''
.print '============================================'
.print '数据清理完成！'
.print '============================================'
.print ''
.print '清理结果统计:'
.print '  users: ' || (SELECT COUNT(*) FROM users) || ' 条'
.print '  players: ' || (SELECT COUNT(*) FROM players) || ' 条'
.print '  clubs: ' || (SELECT COUNT(*) FROM clubs) || ' 条'
.print '  teams: ' || (SELECT COUNT(*) FROM teams) || ' 条'
.print '  coaches: ' || (SELECT COUNT(*) FROM coaches) || ' 条'
.print '  orders: ' || (SELECT COUNT(*) FROM orders) || ' 条'
.print '  reports: ' || (SELECT COUNT(*) FROM reports) || ' 条'
.print '  weekly_reports: ' || (SELECT COUNT(*) FROM weekly_reports) || ' 条'
.print '  match_summaries: ' || (SELECT COUNT(*) FROM match_summaries) || ' 条'
.print '  physical_test_activities: ' || (SELECT COUNT(*) FROM physical_test_activities) || ' 条'
.print ''
.print '现在可以导入新数据了！'
