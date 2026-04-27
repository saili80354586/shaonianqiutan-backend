-- ============================================================
-- 少年球探 - 分析师测试账号数据
-- 生成时间: 2026-04-10
-- 账号规则: 138XXXXXXXX / 123456
-- ============================================================

-- 清理可能存在的旧数据
DELETE FROM analysts WHERE user_id BETWEEN 30 AND 33;
DELETE FROM users WHERE id BETWEEN 30 AND 33;

-- 1. 分析师 (4个) - ID: 30-33
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city, created_at, updated_at) VALUES
(30, '13800000030', '$2a$10$test', '陈分析师', '/images/avatars/analyst.png', 'analyst', 'active', '陈分析师', '上海', '上海市', datetime('now'), datetime('now')),
(31, '13800000031', '$2a$10$test', '林分析师', '/images/avatars/analyst.png', 'analyst', 'active', '林分析师', '北京', '北京市', datetime('now'), datetime('now')),
(32, '13800000032', '$2a$10$test', '周分析师', '/images/avatars/analyst.png', 'analyst', 'active', '周分析师', '广东', '广州市', datetime('now'), datetime('now')),
(33, '13800000033', '$2a$10$test', '吴分析师', '/images/avatars/analyst.png', 'analyst', 'active', '吴分析师', '四川', '成都市', datetime('now'), datetime('now'));

INSERT INTO analysts (user_id, name, bio, specialty, experience, profession, is_pro_player, has_case, case_detail, contact_phone, contact_email, rating, review_count, status, created_at, updated_at) VALUES
(30, '陈分析师', '10年足球分析经验，专注进攻战术和技术特点分析。曾为多家中超俱乐部提供球探报告。', '["进攻分析", "技术报告", "战术板分析"]', 10, '前职业球队技术分析师', 0, 1, '曾为上海绿地、北京国安等多家俱乐部提供球员分析报告，擅长技术动作拆解和进攻威胁评估。', '13800000030', 'chen_analyst@example.com', 4.8, 156, 'active', datetime('now'), datetime('now')),
(31, '林分析师', '8年青训分析经验，专注防守和体能评估。', '["防守分析", "体能报告", "青训评估"]', 8, '体育科学研究员', 0, 1, '专注青少年球员体能发展评估，为多家青训机构提供体能分析报告。', '13800000031', 'lin_analyst@example.com', 4.6, 98, 'active', datetime('now'), datetime('now')),
(32, '周分析师', '5年门将专项分析经验，为多家职业俱乐部评估门将。', '["门将专项", "位置感分析", "扑救技术"]', 5, '门将教练转型', 0, 1, '前职业门将，擅长门将位置感、扑救选位、脚下技术等全方位分析。', '13800000032', 'zhou_analyst@example.com', 4.7, 72, 'active', datetime('now'), datetime('now')),
(33, '林分析师', '12年综合评估经验，青训专家，擅长球员潜力评估和成长规划。', '["综合评估", "青训专家", "潜力预测"]', 12, '青训总监', 1, 1, '曾任中超俱乐部青训总监，累计评估青少年球员超过5000人次，擅长发现潜力新星。', '13800000033', 'wu_analyst@example.com', 4.9, 234, 'active', datetime('now'), datetime('now'));

SELECT '分析师测试账号插入完成!' as result;
