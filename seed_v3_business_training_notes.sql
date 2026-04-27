-- ============================================
-- 少年球探 - 训练笔记数据 v3.1 (修正表结构)
-- ============================================

-- 王教练的训练笔记 (coach_id=1 in coaches table)
INSERT INTO training_notes (id, coach_id, player_id, title, content, category, rating, is_public, created_at) VALUES
(1, 1, 2001, '前锋专项训练计划', '本周重点训练: 1.射门力量练习 2.跑位意识培养 3.对抗能力提升', 'training_plan', 5, 1, datetime('now', '-7 days')),
(2, 1, 2002, '中场技术提升建议', '李小强的传球视野很好，建议加强长传精准度和远射能力', 'feedback', 4, 1, datetime('now', '-5 days')),
(3, 1, 2003, '比赛表现记录', '张小刚在最近比赛中头球争顶成功率80%，需要继续保持并加强地面防守', 'match_note', 4, 1, datetime('now', '-3 days')),
(4, 1, 2001, '月度评估', '王小明4月份训练出勤率100%，态度认真，进取心强', 'assessment', 5, 1, datetime('now', '-1 days'));

-- 李教练的训练笔记 (coach_id=2 in coaches table)
INSERT INTO training_notes (id, coach_id, player_id, title, content, category, rating, is_public, created_at) VALUES
(5, 2, 2005, '前锋基础训练', '陈小龙左脚技术出色，本月重点培养射门信心', 'training_plan', 4, 1, datetime('now', '-6 days')),
(6, 2, 2006, '中场意识培养', '赵小虎需要加强无球跑动意识，扩展视野', 'feedback', 3, 1, datetime('now', '-4 days'));

.print '训练笔记导入完成: ' || (SELECT COUNT(*) FROM training_notes) || ' 条'
