-- ============================================
-- 少年球探 - 报告数据 v3.1 (修正表结构)
-- ============================================

-- 王小明 (2001) 的报告 - 2个
INSERT INTO reports (id, order_id, user_id, analyst_id, player_name, player_birth_date, player_position, player_province, player_city, content, status, created_at, updated_at) VALUES
(1, 1, 2001, 30, '王小明', '2014-03-15', '前锋', '上海', '上海市', '{"overall": 85, "shooting": 90, "dribbling": 88, "pace": 92, "positioning": 85, "summary": "王小明是一名极具潜力的进攻型球员，速度优势明显，射门欲望强烈。"}', 'completed', datetime('now', '-28 days'), datetime('now', '-28 days')),
(2, 2, 2001, 33, '王小明', '2014-03-15', '前锋', '上海', '上海市', '{"overall": 88, "technical": 85, "tactical": 82, "physical": 90, "mental": 86, "summary": "综合评估为优秀，建议加强战术理解能力的培养。"}', 'completed', datetime('now', '-13 days'), datetime('now', '-13 days'));

-- 李小强 (2002) 的报告 - 1个
INSERT INTO reports (id, order_id, user_id, analyst_id, player_name, player_birth_date, player_position, player_province, player_city, content, status, created_at, updated_at) VALUES
(3, 3, 2002, 30, '李小强', '2014-07-22', '中场', '上海', '上海市', '{"overall": 82, "passing": 88, "vision": 85, "control": 80, "stamina": 78, "summary": "李小强具备良好的中场调度能力，传球视野开阔。"}', 'completed', datetime('now', '-18 days'), datetime('now', '-18 days'));

-- 张小刚 (2003) 的报告 - 1个
INSERT INTO reports (id, order_id, user_id, analyst_id, player_name, player_birth_date, player_position, player_province, player_city, content, status, created_at, updated_at) VALUES
(4, 4, 2003, 31, '张小刚', '2014-05-10', '后卫', '上海', '上海市', '{"overall": 80, "tackling": 82, "heading": 85, "positioning": 78, "interception": 80, "summary": "张小刚身体素质出色，头球能力突出，位置感有待提高。"}', 'completed', datetime('now', '-23 days'), datetime('now', '-23 days'));

.print '报告数据导入完成: ' || (SELECT COUNT(*) FROM reports) || ' 条'
