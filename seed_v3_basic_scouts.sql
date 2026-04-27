-- ============================================
-- 少年球探 - 球探数据 v3.1 (修正表结构)
-- ============================================

INSERT INTO scouts (id, user_id, scouting_experience, specialties, preferred_age_groups, scouting_regions, current_organization, bio, verified, total_discovered, total_reports, created_at, updated_at) VALUES
(1, 24, '5年', '进攻型球员,技术型中场,边锋', 'U10,U12,U14', '上海,江苏,浙江', '上海绿地球探部', '赵球探专注华东地区青少年足球人才发掘，尤其擅长发现进攻型天才球员。', 1, 18, 18, datetime('now'), datetime('now')),
(2, 25, '3年', '防守型球员,守门员,中后卫', 'U12,U14,U16', '北京,天津,河北', '北方球探网', '陈球探专注华北地区，善于发现防守型球员和守门员人才。', 1, 10, 10, datetime('now'), datetime('now'));

.print '球探数据导入完成: ' || (SELECT COUNT(*) FROM scouts) || ' 条'
