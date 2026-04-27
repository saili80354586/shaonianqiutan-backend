-- ============================================
-- 少年球探 - 补充周报数据至完整
-- 版本: v1.4
-- 日期: 2026-04-11
-- 说明: U12二队第二周期(period_id=4, team_id=2)缺少4条周报记录
-- ============================================

-- 1. 查看当前周报数据分布
SELECT '当前周报数据分布:' as info;
SELECT team_id, COUNT(*) as cnt FROM weekly_reports GROUP BY team_id;

-- 2. 查看 weekly_report_periods 确认周期信息
SELECT '周报周期信息:' as info;
SELECT id, team_id, week_start, week_end, status FROM weekly_report_periods;

-- 3. 补充 U12二队 (team_id=2) 第四周期 (period_id=4) 的周报数据
-- 球员2005: 陈小龙 - 前锋
INSERT INTO weekly_reports (
    team_id, player_id, coach_id, week_start, week_end,
    training_participation, physical_status, mental_status,
    technical_performance, improvements, next_week_goals,
    notes, review_status, review_coach_id, review_comment, review_rating,
    reviewed_at, created_at, updated_at, deadline, training_count,
    training_duration, absence_count, technical_content, tactical_content,
    physical_condition, match_performance, self_attitude_rating,
    self_technique_rating, self_teamwork_rating, improvements_detail,
    weaknesses, fatigue_level, injuries, sleep_quality, diet_condition,
    message_to_coach, attachments, submit_status, submitted_at,
    coach_attitude_rating, coach_technique_rating, coach_tactics_rating,
    coach_knowledge_rating, strengths_acknowledgment, suggestions,
    knowledge_feedback, next_week_focus, recommend_award
) VALUES (
    2, 2005, 21,
    '2026-04-03', '2026-04-09',
    'full', 5, 5,
    '进攻欲望强烈，射门准度高，本周进球3个',
    '体能分配需提升，下半场注意力下降',
    '加强体能训练，注意比赛节奏控制',
    '本周表现优异', 'pending', NULL, '', 0, NULL,
    '2026-04-10 17:21:40', '2026-04-10 17:21:40',
    '2026-04-10 17:21:40', 3, 90, 0, '进攻配合练习', '边路突破', '体能耐力训练', '比赛中打进3球', 5, 4, 5, '继续加强射门练习', '体能分配', 3, '无伤病', 3, '饮食正常', '希望增加射门训练时间', '', 'draft', NULL, 0, 0, 0, 0, '', '', '', '', 0
);

-- 球员2006: 赵小虎 - 中场
INSERT INTO weekly_reports (
    team_id, player_id, coach_id, week_start, week_end,
    training_participation, physical_status, mental_status,
    technical_performance, improvements, next_week_goals,
    notes, review_status, review_coach_id, review_comment, review_rating,
    reviewed_at, created_at, updated_at, deadline, training_count,
    training_duration, absence_count, technical_content, tactical_content,
    physical_condition, match_performance, self_attitude_rating,
    self_technique_rating, self_teamwork_rating, improvements_detail,
    weaknesses, fatigue_level, injuries, sleep_quality, diet_condition,
    message_to_coach, attachments, submit_status, submitted_at,
    coach_attitude_rating, coach_technique_rating, coach_tactics_rating,
    coach_knowledge_rating, strengths_acknowledgment, suggestions,
    knowledge_feedback, next_week_focus, recommend_award
) VALUES (
    2, 2006, 21,
    '2026-04-03', '2026-04-09',
    'full', 4, 4,
    '中场组织能力进步，长传准确率提升',
    '长传准确性还需加强，有时机的选择欠佳',
    '提升长传精准度，加强比赛阅读能力',
    '训练态度认真', 'pending', NULL, '', 0, NULL,
    '2026-04-10 17:21:40', '2026-04-10 17:21:40',
    '2026-04-10 17:21:40', 3, 90, 0, '中场组织练习', '战术跑位', '技术细节训练', '中场调度得当', 4, 4, 4, '减少无谓失误', '长传准确性', 2, '无伤病', 4, '饮食正常', '希望多进行实战对抗', '', 'draft', NULL, 0, 0, 0, 0, '', '', '', '', 0
);

-- 球员2007: 孙小杰 - 后卫
INSERT INTO weekly_reports (
    team_id, player_id, coach_id, week_start, week_end,
    training_participation, physical_status, mental_status,
    technical_performance, improvements, next_week_goals,
    notes, review_status, review_coach_id, review_comment, review_rating,
    reviewed_at, created_at, updated_at, deadline, training_count,
    training_duration, absence_count, technical_content, tactical_content,
    physical_condition, match_performance, self_attitude_rating,
    self_technique_rating, self_teamwork_rating, improvements_detail,
    weaknesses, fatigue_level, injuries, sleep_quality, diet_condition,
    message_to_coach, attachments, submit_status, submitted_at,
    coach_attitude_rating, coach_technique_rating, coach_tactics_rating,
    coach_knowledge_rating, strengths_acknowledgment, suggestions,
    knowledge_feedback, next_week_focus, recommend_award
) VALUES (
    2, 2007, 21,
    '2026-04-03', '2026-04-09',
    'partial', 4, 3,
    '防守意识增强，抢断果断，补位及时',
    '位置感还需提升，有时盯人失误',
    '加强位置感训练，提高防守专注度',
    '因病缺席1次训练', 'pending', NULL, '', 0, NULL,
    '2026-04-10 17:21:40', '2026-04-10 17:21:40',
    '2026-04-10 17:21:40', 2, 60, 1, '防守基础训练', '防守站位', '体能速度训练', '防守端表现稳定', 4, 3, 4, '加强位置意识', '盯人防守', 3, '感冒已康复', 4, '食欲一般', '希望加强个人防守练习', '', 'draft', NULL, 0, 0, 0, 0, '', '', '', '', 0
);

-- 球员2008: 周小鹏 - 门将
INSERT INTO weekly_reports (
    team_id, player_id, coach_id, week_start, week_end,
    training_participation, physical_status, mental_status,
    technical_performance, improvements, next_week_goals,
    notes, review_status, review_coach_id, review_comment, review_rating,
    reviewed_at, created_at, updated_at, deadline, training_count,
    training_duration, absence_count, technical_content, tactical_content,
    physical_condition, match_performance, self_attitude_rating,
    self_technique_rating, self_teamwork_rating, improvements_detail,
    weaknesses, fatigue_level, injuries, sleep_quality, diet_condition,
    message_to_coach, attachments, submit_status, submitted_at,
    coach_attitude_rating, coach_technique_rating, coach_tactics_rating,
    coach_knowledge_rating, strengths_acknowledgment, suggestions,
    knowledge_feedback, next_week_focus, recommend_award
) VALUES (
    2, 2008, 21,
    '2026-04-03', '2026-04-09',
    'full', 5, 5,
    '扑救反应迅速，门线技术稳定，出击果断',
    '出击时机需改进，有时过于保守',
    '提升出击时机判断，加强高空球处理',
    '本周训练状态极佳', 'pending', NULL, '', 0, NULL,
    '2026-04-10 17:21:40', '2026-04-10 17:21:40',
    '2026-04-10 17:21:40', 3, 90, 0, '门将专项训练', '门线技术', '反应速度训练', '扑出2个必进球', 5, 5, 4, '提高出击时机', '出击判断', 2, '无伤病', 5, '饮食正常', '希望进行更多实战扑救', '', 'draft', NULL, 0, 0, 0, 0, '', '', '', '', 0
);

-- 4. 验证更新结果
SELECT '补充后周报数据分布:' as info;
SELECT team_id, COUNT(*) as cnt FROM weekly_reports GROUP BY team_id;

SELECT '总周报数:' as info;
SELECT COUNT(*) as total FROM weekly_reports;

-- 5. 列出U12二队所有周报
SELECT 'U12二队所有周报:' as info;
SELECT
    wr.id,
    wr.player_id,
    u.name as player_name,
    wr.week_start,
    wr.week_end,
    wr.training_participation,
    wr.physical_status,
    wr.mental_status,
    wr.review_status
FROM weekly_reports wr
JOIN users u ON wr.player_id = u.id
WHERE wr.team_id = 2
ORDER BY wr.week_start DESC;
