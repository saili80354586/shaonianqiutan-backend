-- ============================================
-- 周报功能测试数据
-- 包含：球员账号、球队关联、周报数据、周报周期
-- ============================================

-- 1. 创建测试球员账号（8名球员，分配到不同球队）
-- 密码都是 '123456' 的 bcrypt 哈希
INSERT OR IGNORE INTO users (id, phone, password, nickname, name, role, status, gender, age, position, foot, created_at, updated_at) VALUES
-- U12一队球员 (4名)
(1001, '13800111001', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '王小明', '王小明', 'user', 'active', '男', 11, '前锋', '右脚', datetime('now'), datetime('now')),
(1002, '13800111002', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '李小强', '李小强', 'user', 'active', '男', 12, '中场', '左脚', datetime('now'), datetime('now')),
(1003, '13800111003', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '张小刚', '张小刚', 'user', 'active', '男', 11, '后卫', '右脚', datetime('now'), datetime('now')),
(1004, '13800111004', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '刘小军', '刘小军', 'user', 'active', '男', 12, '守门员', '右脚', datetime('now'), datetime('now')),
-- U12二队球员 (2名)
(1005, '13800111005', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '陈小龙', '陈小龙', 'user', 'active', '男', 11, '前锋', '左脚', datetime('now'), datetime('now')),
(1006, '13800111006', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '赵小虎', '赵小虎', 'user', 'active', '男', 12, '中场', '右脚', datetime('now'), datetime('now')),
-- U14精英队球员 (2名)
(1007, '13800111007', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '马小军', '马小军', 'user', 'active', '男', 13, '前锋', '右脚', datetime('now'), datetime('now')),
(1008, '13800111008', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjIXdB5YW8WfwcPx9DYRH0XmDnn0gS.', '周小杰', '周小杰', 'user', 'active', '男', 14, '后卫', '左脚', datetime('now'), datetime('now'));

-- 2. 创建球队-球员关联
INSERT OR IGNORE INTO team_players (team_id, user_id, player_id, status, joined_at, created_at, updated_at) VALUES
-- U12一队 (team_id=1)
(1, 1001, 1001, 'active', date('now'), datetime('now'), datetime('now')),
(1, 1002, 1002, 'active', date('now'), datetime('now'), datetime('now')),
(1, 1003, 1003, 'active', date('now'), datetime('now'), datetime('now')),
(1, 1004, 1004, 'active', date('now'), datetime('now'), datetime('now')),
-- U12二队 (team_id=2)
(2, 1005, 1005, 'active', date('now'), datetime('now'), datetime('now')),
(2, 1006, 1006, 'active', date('now'), datetime('now'), datetime('now')),
-- U14精英队 (team_id=3)
(3, 1007, 1007, 'active', date('now'), datetime('now'), datetime('now')),
(3, 1008, 1008, 'active', date('now'), datetime('now'), datetime('now'));

-- 3. 创建球队-教练关联（假设教练ID为666）
INSERT OR IGNORE INTO team_coaches (team_id, user_id, role, status, created_at, updated_at) VALUES
(1, 666, 'head_coach', 'active', datetime('now'), datetime('now')),
(2, 666, 'assistant', 'active', datetime('now'), datetime('now')),
(3, 666, 'head_coach', 'active', datetime('now'), datetime('now'));

-- 4. 创建周报周期（当前周和上周）
INSERT OR IGNORE INTO weekly_report_periods (id, team_id, week_start, week_end, deadline, status, total_players, submitted_count, pending_count, reviewed_count, created_at, updated_at) VALUES
-- 上周（已归档）
(1, 1, date('now', 'weekday 1', '-14 days'), date('now', 'weekday 0', '-7 days'), datetime('now', 'weekday 4', '-7 days'), 'archived', 4, 4, 0, 4, datetime('now'), datetime('now')),
(2, 2, date('now', 'weekday 1', '-14 days'), date('now', 'weekday 0', '-7 days'), datetime('now', 'weekday 4', '-7 days'), 'archived', 2, 2, 0, 2, datetime('now'), datetime('now')),
(3, 3, date('now', 'weekday 1', '-14 days'), date('now', 'weekday 0', '-7 days'), datetime('now', 'weekday 4', '-7 days'), 'archived', 2, 2, 0, 2, datetime('now'), datetime('now')),
-- 本周（进行中）
(4, 1, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'active', 4, 1, 2, 1, datetime('now'), datetime('now')),
(5, 2, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'active', 2, 0, 2, 0, datetime('now'), datetime('now')),
(6, 3, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'active', 2, 1, 0, 1, datetime('now'), datetime('now'));

-- 5. 创建周报数据（多种状态，形成完整测试场景）

-- ===== 上周数据（已归档） =====
-- 王小明 - 已通过
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, self_technique_rating, self_teamwork_rating,
  technical_performance, improvements, next_week_goals, notes,
  review_comment, review_rating, reviewed_at, created_at, updated_at) VALUES
(1, 1, 1001, 666, date('now', 'weekday 1', '-14 days'), date('now', 'weekday 0', '-7 days'), datetime('now', 'weekday 4', '-7 days'), 'submitted', 'approved',
 'full', 5, 5, 5, 5, 4, 4, '本周训练态度积极，技术动作有进步', '需要加强体能训练', '争取下周比赛中进球', '本周表现不错',
 '表现优秀，继续保持！', 5, datetime('now', '-5 days'), datetime('now', '-8 days'), datetime('now'));

-- 李小强 - 已通过
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, self_technique_rating, self_teamwork_rating,
  technical_performance, improvements, next_week_goals, notes,
  review_comment, review_rating, reviewed_at, created_at, updated_at) VALUES
(2, 1, 1002, 666, date('now', 'weekday 1', '-14 days'), date('now', 'weekday 0', '-7 days'), datetime('now', 'weekday 4', '-7 days'), 'submitted', 'approved',
 'full', 5, 4, 5, 4, 5, 5, '传球准确率提高，视野开阔', '防守意识需要加强', '提高拦截成功率', '',
 '中场组织得很好，技术全面。', 4, datetime('now', '-5 days'), datetime('now', '-8 days'), datetime('now'));

-- 张小刚 - 已通过
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, self_technique_rating, self_teamwork_rating,
  technical_performance, improvements, next_week_goals, notes,
  review_comment, review_rating, reviewed_at, created_at, updated_at) VALUES
(3, 1, 1003, 666, date('now', 'weekday 1', '-14 days'), date('now', 'weekday 0', '-7 days'), datetime('now', 'weekday 4', '-7 days'), 'submitted', 'approved',
 'partial', 3, 4, 4, 4, 4, 3, '防守稳健，但速度有待提高', '需要提高转身速度', '减少失误', '周三请假',
 '防守意识不错，继续加油。', 4, datetime('now', '-5 days'), datetime('now', '-8 days'), datetime('now'));

-- 刘小军 - 已通过
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, self_technique_rating, self_teamwork_rating,
  technical_performance, improvements, next_week_goals, notes,
  review_comment, review_rating, reviewed_at, created_at, updated_at) VALUES
(4, 1, 1004, 666, date('now', 'weekday 1', '-14 days'), date('now', 'weekday 0', '-7 days'), datetime('now', 'weekday 4', '-7 days'), 'submitted', 'approved',
 'full', 5, 5, 5, 5, 4, 4, '守门技术稳定，指挥防守有声', '出击时机把握', '提高高空球处理能力', '',
 '门将表现稳定，领导力强。', 5, datetime('now', '-5 days'), datetime('now', '-8 days'), datetime('now'));

-- ===== 本周数据（进行中） =====

-- U12一队本周数据
-- 王小明 - 本周已审核通过
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, self_technique_rating, self_teamwork_rating,
  technical_performance, improvements, next_week_goals, notes,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award,
  review_comment, review_rating, reviewed_at, submitted_at, created_at, updated_at) VALUES
(5, 1, 1001, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'submitted', 'approved',
 'full', 5, 5, 5, 5, 4, 4, '本周状态火热，训练中打进多球', '无', '周末比赛帽子戏法', '状态很好',
 5, 4, 4, 4, '射门技术出色，门前嗅觉灵敏', '多练习头球技术', '战术跑位意识强', '保持状态，准备周末比赛', 1,
 '本周表现非常出色，训练中展现出强烈的进球欲望！', 5, datetime('now', '-1 days'), datetime('now', '-3 days'), datetime('now', '-4 days'), datetime('now'));

-- 李小强 - 本周待审核
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, self_technique_rating, self_teamwork_rating,
  technical_performance, improvements, next_week_goals, notes,
  submitted_at, created_at, updated_at) VALUES
(6, 1, 1002, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'submitted', 'pending',
 'full', 5, 4, 5, 4, 5, 4, '传球配合默契，组织了多次进攻', '体能分配需要优化', '控制比赛节奏', '',
 datetime('now', '-2 days'), datetime('now', '-4 days'), datetime('now'));

-- 张小刚 - 本周草稿（未提交）
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  created_at, updated_at) VALUES
(7, 1, 1003, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'draft', 'pending',
 datetime('now', '-4 days'), datetime('now'));

-- 刘小军 - 本周草稿（未提交）
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  created_at, updated_at) VALUES
(8, 1, 1004, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'draft', 'pending',
 datetime('now', '-4 days'), datetime('now'));

-- U12二队本周数据（全部草稿，待填写）
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  created_at, updated_at) VALUES
(9, 2, 1005, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'draft', 'pending', datetime('now', '-4 days'), datetime('now')),
(10, 2, 1006, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'draft', 'pending', datetime('now', '-4 days'), datetime('now'));

-- U14精英队本周数据
-- 马小军 - 已审核通过
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, self_technique_rating, self_teamwork_rating,
  technical_performance, improvements, next_week_goals, notes,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award,
  review_comment, review_rating, reviewed_at, submitted_at, created_at, updated_at) VALUES
(11, 3, 1007, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'submitted', 'approved',
 'full', 5, 5, 5, 5, 5, 4, '作为队长带领球队取胜，打进2球1助攻', '无', '继续带领球队前进', '',
 5, 5, 4, 4, '射门力量和精准度都很出色', '可以多尝试远射', '战术理解能力强', '保持领袖气质', 1,
 '队长表现优异，是球队的灵魂人物！', 5, datetime('now', '-1 days'), datetime('now', '-3 days'), datetime('now', '-4 days'), datetime('now'));

-- 周小杰 - 草稿（部分填写）
INSERT OR IGNORE INTO weekly_reports (id, team_id, player_id, coach_id, week_start, week_end, deadline, submit_status, review_status,
  training_participation, training_count, physical_status, mental_status,
  self_attitude_rating, created_at, updated_at) VALUES
(12, 3, 1008, 666, date('now', 'weekday 1', '-7 days'), date('now', 'weekday 0'), datetime('now', 'weekday 4'), 'draft', 'pending',
 'partial', 3, 4, 4, 4, datetime('now', '-2 days'), datetime('now'));

-- 6. 创建一些通知数据（让通知中心有内容）
INSERT OR IGNORE INTO notifications (user_id, type, title, content, is_read, priority, created_at) VALUES
-- 给王小明的通知
(1001, 'weekly_report_created', '教练发起了本周周报', '王教练 发起了一篇新周报，请填写', 1, 3, datetime('now', '-4 days')),
(1001, 'weekly_report_approved', '周报已审核', '王教练 已审核你的周报，评分 5 星', 0, 3, datetime('now', '-1 days')),
-- 给李小强的通知
(1002, 'weekly_report_created', '教练发起了本周周报', '王教练 发起了一篇新周报，请填写', 1, 3, datetime('now', '-4 days')),
-- 给张小刚的通知
(1003, 'weekly_report_created', '教练发起了本周周报', '王教练 发起了一篇新周报，请填写', 0, 3, datetime('now', '-4 days')),
-- 给刘小军的通知
(1004, 'weekly_report_created', '教练发起了本周周报', '王教练 发起了一篇新周报，请填写', 1, 3, datetime('now', '-4 days')),
(1004, 'weekly_report_approved', '周报已审核', '王教练 已审核你的周报，评分 5 星', 0, 3, datetime('now', '-1 days'));
