-- ============================================
-- 少年球探 - 教练数据 v3.1 (修正表结构)
-- ============================================

INSERT INTO coaches (id, user_id, license_type, specialties, bio, coaching_years, current_club, verified, created_at, updated_at) VALUES
(1, 20, 'A级', '进攻训练,心理辅导,战术分析', '王教练从事青少年足球训练8年，擅长进攻战术训练和球员心理辅导。', 8, '上海绿地青训俱乐部', 1, datetime('now'), datetime('now')),
(2, 21, 'B级', '防守训练,体能训练,守门员专项', '李教练专注于青少年足球体能和防守训练，对守门员培养有独到见解。', 5, '上海绿地青训俱乐部', 1, datetime('now'), datetime('now'));

.print '教练数据导入完成: ' || (SELECT COUNT(*) FROM coaches) || ' 条'
