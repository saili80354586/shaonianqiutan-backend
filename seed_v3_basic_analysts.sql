-- ============================================
-- 少年球探 - 分析师数据 v3.1 (修正表结构)
-- ============================================

INSERT INTO analysts (id, user_id, name, bio, specialty, experience, rating, review_count, status, created_at, updated_at) VALUES
(1, 30, '陈分析师', '陈分析师专注于进攻型球员的技术评估，拥有丰富的比赛分析经验。', '进攻分析', 5, 4.8, 12, 'active', datetime('now'), datetime('now')),
(2, 31, '林分析师', '林分析师擅长防守端球员的能力评估，专注于后卫和后腰位置。', '防守分析', 3, 4.6, 8, 'active', datetime('now'), datetime('now')),
(3, 32, '周分析师', '周分析师是守门员专项分析师，对门将的各项能力指标有深入研究。', '守门员专项', 2, 4.5, 5, 'active', datetime('now'), datetime('now')),
(4, 33, '吴分析师', '吴分析师擅长综合评估球员各项素质，预测球员发展潜力。', '综合评估', 6, 4.9, 15, 'active', datetime('now'), datetime('now'));

.print '分析师数据导入完成: ' || (SELECT COUNT(*) FROM analysts) || ' 条'
