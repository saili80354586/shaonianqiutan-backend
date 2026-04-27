-- ============================================
-- 少年球探 - 周报数据 v3.1 (修正表结构)
-- ============================================

-- 上周周报 (已审核)
INSERT INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, training_count, training_participation, physical_status, mental_status, technical_performance, improvements, next_week_goals, review_status, review_comment, review_rating, reviewed_at, created_at, updated_at, deadline, submit_status) VALUES
(1, 1, 2001, 20, date('now', '-14 days'), date('now', '-8 days'), 3, 'full', 5, 5, '本周训练状态不错，射门练习有进步', '继续加强远射训练', '提升体能储备', 'reviewed', '进攻意识强，继续保持', 5, datetime('now', '-6 days'), datetime('now', '-7 days'), datetime('now', '-6 days'), datetime('now', '-1 days'), 'submitted'),
(2, 1, 2002, 20, date('now', '-14 days'), date('now', '-8 days'), 3, 'full', 4, 4, '传球准确率有所提高', '加强长传练习', '提升对抗能力', 'reviewed', '中场调度能力进步明显', 4, datetime('now', '-6 days'), datetime('now', '-7 days'), datetime('now', '-6 days'), datetime('now', '-1 days'), 'submitted'),
(3, 1, 2003, 20, date('now', '-14 days'), date('now', '-8 days'), 3, 'full', 4, 4, '头球争顶成功率提升', '加强地面防守', '提升位置感', 'reviewed', '防守位置感需要加强', 4, datetime('now', '-6 days'), datetime('now', '-7 days'), datetime('now', '-6 days'), datetime('now', '-1 days'), 'submitted'),
(4, 1, 2004, 20, date('now', '-14 days'), date('now', '-8 days'), 3, 'partial', 3, 3, '扑救反应有所提升', '加强出击练习', '提升判断力', 'reviewed', '出击时机判断需改进', 3, datetime('now', '-6 days'), datetime('now', '-7 days'), datetime('now', '-6 days'), datetime('now', '-1 days'), 'submitted');

-- 本周周报 (各种状态)
INSERT INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, training_count, training_participation, physical_status, mental_status, technical_performance, improvements, next_week_goals, review_status, review_comment, review_rating, reviewed_at, created_at, updated_at, deadline, submit_status) VALUES
-- 已审核
(5, 1, 2001, 20, date('now', '-7 days'), date('now', '-1 days'), 3, 'full', 5, 5, '本周比赛进球了，很开心', '保持状态', '继续进步', 'reviewed', '门前嗅觉敏锐，继续保持', 5, datetime('now', '-1 days'), datetime('now', '-2 days'), datetime('now', '-1 days'), datetime('now', '+1 days'), 'submitted'),
-- 待审核
(6, 1, 2002, 20, date('now', '-7 days'), date('now', '-1 days'), 3, 'full', 4, 4, '角球配合有进步', '加强定位球', '提升配合', 'pending', NULL, NULL, NULL, datetime('now', '-1 days'), NULL, datetime('now', '+1 days'), 'submitted'),
-- 草稿
(7, 1, 2003, 20, date('now', '-7 days'), date('now', '-1 days'), 3, 'partial', 3, 3, NULL, NULL, NULL, 'pending', NULL, NULL, NULL, datetime('now', '-3 days'), NULL, datetime('now', '+1 days'), 'draft'),
-- 草稿
(8, 1, 2004, 20, date('now', '-7 days'), date('now', '-1 days'), 3, 'full', 3, 3, NULL, NULL, NULL, 'pending', NULL, NULL, NULL, datetime('now', '-4 days'), NULL, datetime('now', '+1 days'), 'draft');

.print '周报数据导入完成: ' || (SELECT COUNT(*) FROM weekly_reports) || ' 条'
