-- ============================================
-- 少年球探 - 球队主页数据 v3.1 (修正表结构)
-- ============================================

-- U12一队主页
INSERT INTO team_homes (id, team_id, hero, about, honors, contact, created_at, updated_at) VALUES
(1, 1, 'https://images.unsplash.com/photo-1551958219-acbc608c6377?w=1200', '上海绿地青训俱乐部U12一队是俱乐部精英梯队，球员经过严格选拔组成，具备较强的竞争力。球队注重技术训练和战术配合，培养球员的足球智慧。', '[{"year":"2024","title":"上海市U12联赛冠军"},{"year":"2023","title":"全国青少年足球邀请赛亚军"},{"year":"2023","title":"浦东新区杯冠军"}]', '主教练: 王教练 (13800000020)', datetime('now'), datetime('now'));

-- U12二队主页
INSERT INTO team_homes (id, team_id, hero, about, honors, contact, created_at, updated_at) VALUES
(2, 2, 'https://images.unsplash.com/photo-1579952363873-27f3bade9f55?w=1200', '上海绿地青训俱乐部U12二队是俱乐部发展梯队，注重基础训练和人才发掘。球队氛围活跃，鼓励球员发挥个人特点，为一队输送优秀人才。', '[{"year":"2024","title":"上海市U12发展组冠军"}]', '主教练: 李教练 (13800000021)', datetime('now'), datetime('now'));

.print '球队主页数据导入完成: ' || (SELECT COUNT(*) FROM team_homes) || ' 条'
