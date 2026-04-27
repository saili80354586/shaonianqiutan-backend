-- ============================================================
-- 少年球探 - 完整测试账号数据
-- 生成时间: 2026-04-08
-- 账号规则:
--   管理员: admin / admin123456
--   其他: 138XXXXXXXX / 123456
-- ============================================================

-- 清理旧数据
DELETE FROM weekly_report_periods;
DELETE FROM weekly_reports;
DELETE FROM team_coaches;
DELETE FROM team_players;
DELETE FROM teams;
DELETE FROM players;
DELETE FROM scouts;
DELETE FROM coaches;
DELETE FROM clubs;
DELETE FROM users WHERE id BETWEEN 1 AND 30 OR phone LIKE '1380000%' OR phone = 'admin';

-- 1. 管理员
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (1, 'admin', '$2a$10$Xb3e1Y7J5Q6zK8hJpY9QmOGjK3l0kN7qM5hJzY4vR8xL3n6wP0qQ', '系统管理员', '/images/avatars/admin.png', 'admin', 'active', '系统管理员', datetime('now'), datetime('now'));

-- 2. 俱乐部 (10个) - ID: 10-19
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city, created_at, updated_at) VALUES
(10, '13800000010', '$2a$10$test', '绿地青训', '/images/avatars/club.png', 'club', 'active', '上海绿地俱乐部', '上海', '上海市', datetime('now'), datetime('now')),
(11, '13800000011', '$2a$10$test', '国安青训', '/images/avatars/club.png', 'club', 'active', '北京国安青训', '北京', '北京市', datetime('now'), datetime('now')),
(12, '13800000012', '$2a$10$test', '恒大足校', '/images/avatars/club.png', 'club', 'active', '广州恒大足校', '广东', '广州市', datetime('now'), datetime('now')),
(13, '13800000013', '$2a$10$test', '泰山青训', '/images/avatars/club.png', 'club', 'active', '山东泰山青训', '山东', '济南市', datetime('now'), datetime('now')),
(14, '13800000014', '$2a$10$test', '苏宁青训', '/images/avatars/club.png', 'club', 'active', '江苏苏宁青训', '江苏', '南京市', datetime('now'), datetime('now')),
(15, '13800000015', '$2a$10$test', '蓉城青训', '/images/avatars/club.png', 'club', 'active', '成都蓉城青训', '四川', '成都市', datetime('now'), datetime('now')),
(16, '13800000016', '$2a$10$test', '三镇青训', '/images/avatars/club.png', 'club', 'active', '武汉三镇青训', '湖北', '武汉市', datetime('now'), datetime('now')),
(17, '13800000017', '$2a$10$test', '绿城足校', '/images/avatars/club.png', 'club', 'active', '浙江绿城足校', '浙江', '杭州市', datetime('now'), datetime('now')),
(18, '13800000018', '$2a$10$test', '嵩山青训', '/images/avatars/club.png', 'club', 'active', '河南嵩山青训', '河南', '郑州市', datetime('now'), datetime('now')),
(19, '13800000019', '$2a$10$test', '津门虎青训', '/images/avatars/club.png', 'club', 'active', '天津津门虎青训', '天津', '天津市', datetime('now'), datetime('now'));

INSERT INTO clubs (user_id, name, description, address, province, city, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at) VALUES
(10, '上海绿地俱乐部', '上海绿地青训俱乐部成立于2010年，是上海市领先的青少年足球培训机构。', '上海市浦东新区世纪大道1000号', '上海', '上海市', '张经理', '13800000010', 2010, 'large', 'enterprise', 50, datetime('now'), datetime('now')),
(11, '北京国安青训', '北京国安足球俱乐部青训基地。', '北京市朝阳区', '北京', '北京市', '李经理', '13800000011', 2002, 'large', 'enterprise', 50, datetime('now'), datetime('now')),
(12, '广州恒大足校', '广州恒大足球学校，全国知名青训机构。', '广东省清远市', '广东', '广州市', '王经理', '13800000012', 2012, 'large', 'enterprise', 50, datetime('now'), datetime('now')),
(13, '山东泰山青训', '山东泰山足球俱乐部青训体系。', '山东省济南市', '山东', '济南市', '刘经理', '13800000013', 1993, 'large', 'professional', 30, datetime('now'), datetime('now')),
(14, '江苏苏宁青训', '江苏苏宁足球俱乐部青训基地。', '江苏省南京市', '江苏', '南京市', '陈经理', '13800000014', 1994, 'medium', 'professional', 30, datetime('now'), datetime('now')),
(15, '成都蓉城青训', '成都蓉城足球俱乐部青训。', '四川省成都市', '四川', '成都市', '赵经理', '13800000015', 2018, 'medium', 'professional', 30, datetime('now'), datetime('now')),
(16, '武汉三镇青训', '武汉三镇足球俱乐部青训。', '湖北省武汉市', '湖北', '武汉市', '杨经理', '13800000016', 2013, 'medium', 'professional', 30, datetime('now'), datetime('now')),
(17, '浙江绿城足校', '浙江绿城足球学校。', '浙江省杭州市', '浙江', '杭州市', '吴经理', '13800000017', 2004, 'large', 'professional', 30, datetime('now'), datetime('now')),
(18, '河南嵩山青训', '河南嵩山龙门足球俱乐部青训。', '河南省郑州市', '河南', '郑州市', '周经理', '13800000018', 1994, 'medium', 'professional', 30, datetime('now'), datetime('now')),
(19, '天津津门虎青训', '天津津门虎足球俱乐部青训。', '天津市', '天津', '天津市', '马经理', '13800000019', 1994, 'medium', 'professional', 30, datetime('now'), datetime('now'));

-- 3. 教练 (4个) - ID: 20-23
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city, created_at, updated_at) VALUES
(20, '13800000020', '$2a$10$test', '王教练', '/images/avatars/coach.png', 'coach', 'active', '王教练', '上海', '上海市', datetime('now'), datetime('now')),
(21, '13800000021', '$2a$10$test', '李教练', '/images/avatars/coach.png', 'coach', 'active', '李教练', '北京', '北京市', datetime('now'), datetime('now')),
(22, '13800000022', '$2a$10$test', '张教练', '/images/avatars/coach.png', 'coach', 'active', '张教练', '山东', '济南市', datetime('now'), datetime('now')),
(23, '13800000023', '$2a$10$test', '刘教练', '/images/avatars/coach.png', 'coach', 'active', '刘教练', '四川', '成都市', datetime('now'), datetime('now'));

INSERT INTO coaches (user_id, license_type, license_number, specialties, bio, coaching_years, current_club, verified, created_at, updated_at) VALUES
(20, 'A级', 'AFC-A-2023-001', '["技术训练", "青少年培养", "战术指导"]', '20年青训经验，前职业球员。', 20, '上海绿地俱乐部', 1, datetime('now'), datetime('now')),
(21, 'B级', 'AFC-B-2023-002', '["体能训练", "战术指导"]', '15年青训经验。', 15, '北京国安青训', 1, datetime('now'), datetime('now')),
(22, 'B级', 'AFC-B-2023-003', '["技术训练", "守门员训练"]', '12年青训经验。', 12, '山东泰山青训', 1, datetime('now'), datetime('now')),
(23, 'C级', 'AFC-C-2023-004', '["青少年培养", "心理辅导"]', '10年青训经验。', 10, '成都蓉城青训', 1, datetime('now'), datetime('now'));

-- 4. 球探 (4个) - ID: 24-27
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city, created_at, updated_at) VALUES
(24, '13800000024', '$2a$10$test', '赵球探', '/images/avatars/scout.png', 'scout', 'active', '赵球探', '上海', '上海市', datetime('now'), datetime('now')),
(25, '13800000025', '$2a$10$test', '陈球探', '/images/avatars/scout.png', 'scout', 'active', '陈球探', '北京', '北京市', datetime('now'), datetime('now')),
(26, '13800000026', '$2a$10$test', '周球探', '/images/avatars/scout.png', 'scout', 'active', '周球探', '广东', '广州市', datetime('now'), datetime('now')),
(27, '13800000027', '$2a$10$test', '吴球探', '/images/avatars/scout.png', 'scout', 'active', '吴球探', '四川', '成都市', datetime('now'), datetime('now'));

INSERT INTO scouts (user_id, scouting_experience, specialties, preferred_age_groups, scouting_regions, current_organization, bio, verified, total_discovered, total_reports, created_at, updated_at) VALUES
(24, '3-5年', '["前锋", "中场"]', '["U12", "U14"]', '["华东"]', '自由球探', '资深球探，专注华东地区。', 1, 50, 120, datetime('now'), datetime('now')),
(25, '1-3年', '["后卫", "中场"]', '["U10", "U12"]', '["华北"]', '北京国安球探部', '北京国安球探。', 1, 30, 80, datetime('now'), datetime('now')),
(26, '3-5年', '["前锋", "守门员"]', '["U12", "U14", "U16"]', '["华南"]', '广州恒大球探部', '广州恒大资深球探。', 1, 80, 200, datetime('now'), datetime('now')),
(27, '1-3年', '["中场", "后卫"]', '["U10", "U12"]', '["西南"]', '成都蓉城球探部', '成都蓉城球探。', 1, 20, 50, datetime('now'), datetime('now'));

-- 5. 球员 (8个) - ID: 101-108
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city, age, position, club, created_at, updated_at) VALUES
(101, '13800002001', '$2a$10$test', '王小明', '/images/avatars/player.png', 'user', 'active', '王小明', '上海', '上海市', 12, '前锋', '上海绿地俱乐部', datetime('now'), datetime('now')),
(102, '13800002002', '$2a$10$test', '李小强', '/images/avatars/player.png', 'user', 'active', '李小强', '北京', '北京市', 12, '中场', '北京国安青训', datetime('now'), datetime('now')),
(103, '13800002003', '$2a$10$test', '张小刚', '/images/avatars/player.png', 'user', 'active', '张小刚', '广东', '广州市', 12, '后卫', '广州恒大足校', datetime('now'), datetime('now')),
(104, '13800002004', '$2a$10$test', '刘小军', '/images/avatars/player.png', 'user', 'active', '刘小军', '山东', '济南市', 12, '守门员', '山东泰山青训', datetime('now'), datetime('now')),
(105, '13800002005', '$2a$10$test', '陈小龙', '/images/avatars/player.png', 'user', 'active', '陈小龙', '江苏', '南京市', 11, '前锋', '江苏苏宁青训', datetime('now'), datetime('now')),
(106, '13800002006', '$2a$10$test', '赵小虎', '/images/avatars/player.png', 'user', 'active', '赵小虎', '四川', '成都市', 11, '中场', '成都蓉城青训', datetime('now'), datetime('now')),
(107, '13800002007', '$2a$10$test', '孙小杰', '/images/avatars/player.png', 'user', 'active', '孙小杰', '浙江', '杭州市', 11, '后卫', '浙江绿城足校', datetime('now'), datetime('now')),
(108, '13800002008', '$2a$10$test', '周小鹏', '/images/avatars/player.png', 'user', 'active', '周小鹏', '湖北', '武汉市', 11, '前锋', '武汉三镇青训', datetime('now'), datetime('now'));

-- 6. 分析师 (4个) - ID: 30-33
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city, created_at, updated_at) VALUES
(30, '13800000030', '$2a$10$test', '李分析师', '/images/avatars/analyst.png', 'analyst', 'active', '李分析师', '上海', '上海市', datetime('now'), datetime('now')),
(31, '13800000031', '$2a$10$test', '王分析师', '/images/avatars/analyst.png', 'analyst', 'active', '王分析师', '北京', '北京市', datetime('now'), datetime('now')),
(32, '13800000032', '$2a$10$test', '张分析师', '/images/avatars/analyst.png', 'analyst', 'active', '张分析师', '广东', '广州市', datetime('now'), datetime('now')),
(33, '13800000033', '$2a$10$test', '赵分析师', '/images/avatars/analyst.png', 'analyst', 'active', '赵分析师', '山东', '济南市', datetime('now'), datetime('now'));

INSERT INTO analysts (user_id, name, specialty, experience, rating, review_count, bio, status, created_at, updated_at) VALUES
(30, '李分析师', '技术评估', 8, 4.8, 120, '专注青少年技术评估，8年经验。', 'active', datetime('now'), datetime('now')),
(31, '王分析师', '体能分析', 6, 4.6, 80, '专业体能分析师，擅长数据驱动分析。', 'active', datetime('now'), datetime('now')),
(32, '张分析师', '战术分析', 10, 4.9, 200, '前职业球队战术分析师，10年经验。', 'active', datetime('now'), datetime('now')),
(33, '赵分析师', '综合评估', 5, 4.5, 60, '多维度综合评估专家。', 'active', datetime('now'), datetime('now'));

SELECT '账号基础数据插入完成!' as result;
