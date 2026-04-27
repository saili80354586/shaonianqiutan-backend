-- ============================================
-- 少年球探 - 用户基础数据 v3.1
-- 包含 18 个测试账号，全部带头像
-- 密码: bcrypt加密后的 '123456'
-- ============================================

-- bcrypt hash for '123456': $2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK
-- 注意: 此hash为测试专用，生产环境请使用正确的加密方式

-- 管理员
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city) VALUES
(1, '13800000001', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '系统管理员', 'https://ui-avatars.com/api/?name=管理员&background=34495E&color=fff&size=200', 'admin', 'active', '系统管理员', '北京', '北京市');

-- 俱乐部管理员
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city, club) VALUES
(10, '13800000010', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '上海绿地管理员', 'https://ui-avatars.com/api/?name=管理员&background=2ECC71&color=fff&size=200', 'club', 'active', '上海绿地管理员', '上海', '上海市', '上海绿地青训俱乐部');

-- 教练 (2人)
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city) VALUES
(20, '13800000020', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '王教练', 'https://ui-avatars.com/api/?name=王教练&background=E74C3C&color=fff&size=200', 'coach', 'active', '王教练', '上海', '上海市'),
(21, '13800000021', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '李教练', 'https://ui-avatars.com/api/?name=李教练&background=3498DB&color=fff&size=200', 'coach', 'active', '李教练', '上海', '上海市');

-- 球探 (2人)
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city) VALUES
(24, '13800000024', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '赵球探', 'https://ui-avatars.com/api/?name=赵球探&background=11998E&color=fff&size=200', 'scout', 'active', '赵球探', '上海', '上海市'),
(25, '13800000025', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '陈球探', 'https://ui-avatars.com/api/?name=陈球探&background=38EF7D&color=fff&size=200', 'scout', 'active', '陈球探', '北京', '北京市');

-- 分析师 (4人)
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, province, city) VALUES
(30, '13800000030', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '陈分析师', 'https://ui-avatars.com/api/?name=陈分析师&background=667EEA&color=fff&size=200', 'analyst', 'active', '陈分析师', '上海', '上海市'),
(31, '13800000031', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '林分析师', 'https://ui-avatars.com/api/?name=林分析师&background=764BA2&color=fff&size=200', 'analyst', 'active', '林分析师', '上海', '上海市'),
(32, '13800000032', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '周分析师', 'https://ui-avatars.com/api/?name=周分析师&background=F093FB&color=fff&size=200', 'analyst', 'active', '周分析师', '北京', '北京市'),
(33, '13800000033', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '吴分析师', 'https://ui-avatars.com/api/?name=吴分析师&background=4FACFE&color=fff&size=200', 'analyst', 'active', '吴分析师', '广东', '广州市');

-- 球员 (8人) - U12一队
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, birth_date, age, gender, height, weight, foot, position, province, city, club, jersey_number) VALUES
(2001, '13800002001', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '王小明', 'https://ui-avatars.com/api/?name=王小明&background=FF6B6B&color=fff&size=200', 'user', 'active', '王小明', '2014-03-15', 12, '男', 148, 40, '右脚', '前锋', '上海', '上海市', '上海绿地青训俱乐部', 9),
(2002, '13800002002', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '李小强', 'https://ui-avatars.com/api/?name=李小强&background=4ECDC4&color=fff&size=200', 'user', 'active', '李小强', '2014-07-22', 12, '男', 145, 38, '右脚', '中场', '上海', '上海市', '上海绿地青训俱乐部', 8),
(2003, '13800002003', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '张小刚', 'https://ui-avatars.com/api/?name=张小刚&background=45B7D1&color=fff&size=200', 'user', 'active', '张小刚', '2014-05-10', 12, '男', 152, 45, '右脚', '后卫', '上海', '上海市', '上海绿地青训俱乐部', 4),
(2004, '13800002004', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '刘小军', 'https://ui-avatars.com/api/?name=刘小军&background=96CEB4&color=fff&size=200', 'user', 'active', '刘小军', '2014-01-08', 12, '男', 150, 42, '右脚', '门将', '上海', '上海市', '上海绿地青训俱乐部', 1);

-- 球员 (8人) - U12二队
INSERT INTO users (id, phone, password, nickname, avatar, role, status, name, birth_date, age, gender, height, weight, foot, position, province, city, club, jersey_number) VALUES
(2005, '13800002005', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '陈小龙', 'https://ui-avatars.com/api/?name=陈小龙&background=FFEAA7&color=333&size=200', 'user', 'active', '陈小龙', '2014-09-05', 12, '男', 146, 39, '左脚', '前锋', '上海', '上海市', '上海绿地青训俱乐部', 11),
(2006, '13800002006', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '赵小虎', 'https://ui-avatars.com/api/?name=赵小虎&background=DDA0DD&color=fff&size=200', 'user', 'active', '赵小虎', '2014-11-18', 12, '男', 143, 37, '右脚', '中场', '上海', '上海市', '上海绿地青训俱乐部', 6),
(2007, '13800002007', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '孙小杰', 'https://ui-avatars.com/api/?name=孙小杰&background=98D8C8&color=fff&size=200', 'user', 'active', '孙小杰', '2014-04-25', 12, '男', 149, 43, '右脚', '后卫', '上海', '上海市', '上海绿地青训俱乐部', 3),
(2008, '13800002008', '$2a$10$NspjCVgVBoffHUIXeR/oGO02QfSAazoNzNyNVpc1xlRqHnzhpyXuK', '周小鹏', 'https://ui-avatars.com/api/?name=周小鹏&background=F7DC6F&color=333&size=200', 'user', 'active', '周小鹏', '2014-02-14', 12, '男', 151, 41, '右脚', '门将', '上海', '上海市', '上海绿地青训俱乐部', 22);

.print '用户数据导入完成: ' || (SELECT COUNT(*) FROM users) || ' 条'
