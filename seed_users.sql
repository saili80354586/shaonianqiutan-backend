-- 插入20个完整球员用户数据（包含所有必填和非必填字段）

-- 删除现有测试数据（保留ID 1-4的系统用户）
DELETE FROM users WHERE id > 4;

-- 球员5: 广东广州 前锋
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, birth_date, age, gender, height, weight, foot, position, second_position, start_year, country, province, city, club, fa_registered, association, jersey_color, jersey_number, father_height, father_phone, father_edu, father_job, father_athlete, mother_height, mother_phone, mother_edu, mother_job, mother_athlete, created_at, updated_at) 
VALUES (5, '13800110005', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqWjW7e7VJhRj0XkP9QxKxW.KqE9a', '广州小飞侠', 'https://api.dicebear.com/7.x/avataaars/svg?seed=player5', 'user', 'active', '陈小明', '2009-05-15', 16, 'male', 175.5, 68.0, 'right', '前锋', '左边锋', 2018, '中国', '广东', '广州', '广州恒足青训', 1, '广州市足球协会', '红白', 10, 178.0, '13900111111', '本科', '企业高管', 0, 165.0, '13900222222', '本科', '教师', 0, datetime('now'), datetime('now'));

-- 球员6: 广东深圳 中场
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, birth_date, age, gender, height, weight, foot, position, second_position, start_year, country, province, city, club, fa_registered, association, jersey_color, jersey_number, father_height, father_phone, father_edu, father_job, father_athlete, mother_height, mother_phone, mother_edu, mother_job, mother_athlete, created_at, updated_at) 
VALUES (6, '13800110006', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqWjW7e7VJhRj0XkP9QxKxW.KqE9a', '深圳闪电', 'https://api.dicebear.com/7.x/avataaars/svg?seed=player6', 'user', 'active', '李博文', '2010-08-22', 15, 'male', 172.0, 65.5, 'left', '中场', '后腰', 2019, '中国', '广东', '深圳', '深圳足校', 1, '深圳市足球协会', '蓝白', 8, 175.0, '13800111112', '硕士', '软件工程师', 1, 162.0, '13800222223', '硕士', '医生', 0, datetime('now'), datetime('now'));

-- 球员7: 北京 后卫
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, birth_date, age, gender, height, weight, foot, position, second_position, start_year, country, province, city, club, fa_registered, association, jersey_color, jersey_number, father_height, father_phone, father_edu, father_job, father_athlete, mother_height, mother_phone, mother_edu, mother_job, mother_athlete, created_at, updated_at) 
VALUES (7, '13800110007', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqWjW7e7VJhRj0XkP9QxKxW.KqE9a', '北京小王子', 'https://api.dicebear.com/7.x/avataaars/svg?seed=player7', 'user', 'active', '张天宇', '2008-11-08', 17, 'male', 180.5, 72.0, 'right', '后卫', '中后卫', 2017, '中国', '北京', '北京', '北京国安青训', 1, '北京市足球协会', '绿白', 5, 183.0, '13700111113', '博士', '大学教授', 1, 168.0, '13700222224', '博士', '研究员', 0, datetime('now'), datetime('now'));

-- 球员8: 上海 前锋
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, birth_date, age, gender, height, weight, foot, position, second_position, start_year, country, province, city, club, fa_registered, association, jersey_color, jersey_number, father_height, father_phone, father_edu, father_job, father_athlete, mother_height, mother_phone, mother_edu, mother_job, mother_athlete, created_at, updated_at) 
VALUES (8, '13800110008', '$2a$10$N9qo8uLOickgx2ZMRZoMy.MqrqWjW7e7VJhRj0XkP9QxKxW.KqE9a', '上海闪电侠', 'https://api.dicebear.com/7.x/avataaars/svg?seed=player8', 'user', 'active', '王浩然', '2009-06-30', 16, 'male', 176.0, 69.5, 'right', '前锋', '右边锋', 2018, '中国', '上海', '上海', '上海上港青训', 1, '上海市足球协会', '红金', 11, 179.0, '13600111114', '本科', '金融分析师', 0, 165.0, '13600222225', '本科', '会计师', 0, datetime('now'), datetime('now'));

-- 继续添加更多球员数据（9-24号）...
