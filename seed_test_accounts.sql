-- 登录页面测试账号 seed 数据
-- 所有账号使用手机号格式，密码统一为 test（后端特殊处理 $2a$10$test）
-- 执行前请确保数据库已初始化

-- 先删除可能存在的测试账号（避免重复插入错误）
DELETE FROM users WHERE phone IN (
  '13800000001', -- 管理员
  '13800000002', -- 分析师
  '13800000003', -- 俱乐部-上海申花
  '13800000004', -- 教练
  '13800138005', -- 球探
  '13800111001', '13800111002', '13800111003', '13800111004', '13800111005',
  '13800111006', '13800111007', '13800111008', '13800111009', '13800111010'
);

-- 删除关联数据（ coaches, clubs, analysts, scouts 表）
DELETE FROM coaches WHERE user_id IN (SELECT id FROM users WHERE phone LIKE '1380000000%' OR phone LIKE '1380013%' OR phone LIKE '13800111%');
DELETE FROM clubs WHERE user_id IN (SELECT id FROM users WHERE phone LIKE '1380000000%' OR phone LIKE '1380013%' OR phone LIKE '13800111%');
DELETE FROM analysts WHERE user_id IN (SELECT id FROM users WHERE phone LIKE '1380000000%' OR phone LIKE '1380013%' OR phone LIKE '13800111%');
DELETE FROM scouts WHERE user_id IN (SELECT id FROM users WHERE phone LIKE '1380000000%' OR phone LIKE '1380013%' OR phone LIKE '13800111%');

-- ==================== 1. 管理员账号 ====================
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (
  9001, 
  '13800000001', 
  '$2a$10$test', 
  '系统管理员', 
  '/images/avatars/admin.png', 
  'admin', 
  'active', 
  '管理员', 
  datetime('now'), 
  datetime('now')
);

-- ==================== 2. 分析师账号 ====================
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (
  9002, 
  '13800000002', 
  '$2a$10$test', 
  '专业分析师', 
  '/images/avatars/analyst.png', 
  'analyst', 
  'active', 
  '分析师', 
  datetime('now'), 
  datetime('now')
);

-- 分析师详情（修正列名以匹配实际表结构）
INSERT INTO analysts (user_id, name, bio, specialty, experience, rating, status, created_at, updated_at)
VALUES (
  9002,
  '分析师',
  '10年职业球探经验，专注于前锋球员分析',
  '前锋分析、技术评估、速度测试',
  10,
  4.9,
  'active',
  datetime('now'),
  datetime('now')
);

-- ==================== 3. 俱乐部账号 ====================
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (
  9003, 
  '13800000003', 
  '$2a$10$test', 
  '申花青训', 
  '/images/avatars/club.png', 
  'club', 
  'active', 
  '上海申花俱乐部', 
  datetime('now'), 
  datetime('now')
);

-- 俱乐部详情
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (
  9003,
  '上海申花俱乐部',
  '/images/clubs/shenhua.png',
  '上海申花青训基地，培养未来足球明星',
  '上海市浦东新区',
  '张经理',
  '13800000003',
  1993,
  'large',
  'enterprise',
  100,
  datetime('now'),
  datetime('now')
);

-- ==================== 4. 教练账号 ====================
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (
  9004, 
  '13800000004', 
  '$2a$10$test', 
  '王指导', 
  '/images/avatars/coach.png', 
  'coach', 
  'active', 
  '王教练', 
  datetime('now'), 
  datetime('now')
);

-- 教练详情（修正列名以匹配实际表结构）
INSERT INTO coaches (user_id, license_type, license_number, specialties, bio, coaching_years, current_club, verified, created_at, updated_at)
VALUES (
  9004,
  'A级',
  'A20240001',
  '["技术训练", "青少年培养", "战术指导"]',
  '15年青训经验，曾培养出多名职业球员',
  15,
  '上海申花俱乐部',
  1,
  datetime('now'),
  datetime('now')
);

-- ==================== 5. 球探账号 ====================
-- 先检查是否已存在
DELETE FROM users WHERE phone = '13800138005';
DELETE FROM scouts WHERE user_id = 9005;

INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (
  9005, 
  '13800138005', 
  '$2a$10$test', 
  '金牌球探', 
  '/images/avatars/scout.png', 
  'user', 
  'active', 
  '李球探', 
  datetime('now'), 
  datetime('now')
);

-- 球探详情（修正列名以匹配实际表结构）
INSERT INTO scouts (user_id, bio, specialties, scouting_regions, verified, total_discovered, total_reports, created_at, updated_at)
VALUES (
  9005,
  '资深球探，专注华东地区青少年球员挖掘',
  '["前锋", "中场"]',
  '["上海", "江苏", "浙江"]',
  1,
  50,
  120,
  datetime('now'),
  datetime('now')
);

-- ==================== 6. 各俱乐部账号 (13800111001-13800111010) ====================
-- 北京国安青训
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9101, '13800111001', '$2a$10$test', '国安青训', '/images/avatars/club.png', 'club', 'active', '北京国安青训', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9101, '北京国安青训', '/images/clubs/guoan.png', '北京国安足球俱乐部青训基地', '北京市朝阳区', '李经理', '13800111001', 1992, 'large', 'enterprise', 100, datetime('now'), datetime('now'));

-- 上海根宝足球基地
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9102, '13800111002', '$2a$10$test', '根宝基地', '/images/avatars/club.png', 'club', 'active', '上海根宝足球基地', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9102, '上海根宝足球基地', '/images/clubs/genbao.png', '徐根宝足球培训基地', '上海市崇明岛', '徐指导', '13800111002', 2000, 'large', 'enterprise', 100, datetime('now'), datetime('now'));

-- 广州恒大足校
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9103, '13800111003', '$2a$10$test', '恒大足校', '/images/avatars/club.png', 'club', 'active', '广州恒大足校', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9103, '广州恒大足校', '/images/clubs/hengda.png', '广州恒大足球学校', '广东省清远市', '王经理', '13800111003', 2012, 'large', 'enterprise', 100, datetime('now'), datetime('now'));

-- 山东泰山青训
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9104, '13800111004', '$2a$10$test', '泰山青训', '/images/avatars/club.png', 'club', 'active', '山东泰山青训', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9104, '山东泰山青训', '/images/clubs/taishan.png', '山东泰山足球俱乐部青训', '山东省济南市', '刘经理', '13800111004', 1993, 'large', 'enterprise', 100, datetime('now'), datetime('now'));

-- 江苏苏宁青训
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9105, '13800111005', '$2a$10$test', '苏宁青训', '/images/avatars/club.png', 'club', 'active', '江苏苏宁青训', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9105, '江苏苏宁青训', '/images/clubs/suning.png', '江苏苏宁足球俱乐部青训', '江苏省南京市', '张经理', '13800111005', 1994, 'large', 'professional', 50, datetime('now'), datetime('now'));

-- 成都蓉城青训
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9106, '13800111006', '$2a$10$test', '蓉城青训', '/images/avatars/club.png', 'club', 'active', '成都蓉城青训', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9106, '成都蓉城青训', '/images/clubs/rongcheng.png', '成都蓉城足球俱乐部青训', '四川省成都市', '陈经理', '13800111006', 2018, 'medium', 'professional', 50, datetime('now'), datetime('now'));

-- 武汉三镇青训
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9107, '13800111007', '$2a$10$test', '三镇青训', '/images/avatars/club.png', 'club', 'active', '武汉三镇青训', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9107, '武汉三镇青训', '/images/clubs/sanzhen.png', '武汉三镇足球俱乐部青训', '湖北省武汉市', '杨经理', '13800111007', 2013, 'medium', 'professional', 50, datetime('now'), datetime('now'));

-- 浙江绿城足校
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9108, '13800111008', '$2a$10$test', '绿城足校', '/images/avatars/club.png', 'club', 'active', '浙江绿城足校', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9108, '浙江绿城足校', '/images/clubs/greentown.png', '浙江绿城足球学校', '浙江省杭州市', '吴经理', '13800111008', 2004, 'large', 'enterprise', 100, datetime('now'), datetime('now'));

-- 河南嵩山青训
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9109, '13800111009', '$2a$10$test', '嵩山青训', '/images/avatars/club.png', 'club', 'active', '河南嵩山青训', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9109, '河南嵩山青训', '/images/clubs/songshan.png', '河南嵩山龙门青训', '河南省郑州市', '赵经理', '13800111009', 1994, 'medium', 'professional', 50, datetime('now'), datetime('now'));

-- 天津津门虎青训
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, created_at, updated_at)
VALUES (9110, '13800111010', '$2a$10$test', '津门虎青训', '/images/avatars/club.png', 'club', 'active', '天津津门虎青训', datetime('now'), datetime('now'));
INSERT INTO clubs (user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, free_test_quota, created_at, updated_at)
VALUES (9110, '天津津门虎青训', '/images/clubs/jinmenhu.png', '天津津门虎足球俱乐部青训', '天津市', '马经理', '13800111010', 1994, 'medium', 'professional', 50, datetime('now'), datetime('now'));

-- ==================== 完成 ====================
SELECT '测试账号插入完成' AS result;
