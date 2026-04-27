-- ============================================
-- 少年球探 - 球员扩展数据 v3.0
-- 用于球探地图显示
-- ============================================

-- U12一队球员
INSERT INTO players (id, user_id, name, nickname, province, city, district, position, age, birth_date, height, weight, foot, club, school, phone, avatar, status) VALUES
(1, 2001, '王小明', '小明', '上海', '上海市', '浦东新区', '前锋', 12, '2014-03-15', 148, 40, '右脚', '上海绿地青训俱乐部', '浦东第一小学', '13800002001', 'https://ui-avatars.com/api/?name=王小明&background=FF6B6B&color=fff&size=200', 1),
(2, 2002, '李小强', '小强', '上海', '上海市', '浦东新区', '中场', 12, '2014-07-22', 145, 38, '右脚', '上海绿地青训俱乐部', '浦东第二小学', '13800002002', 'https://ui-avatars.com/api/?name=李小强&background=4ECDC4&color=fff&size=200', 1),
(3, 2003, '张小刚', '小刚', '上海', '上海市', '浦东新区', '后卫', 12, '2014-05-10', 152, 45, '右脚', '上海绿地青训俱乐部', '浦东第三小学', '13800002003', 'https://ui-avatars.com/api/?name=张小刚&background=45B7D1&color=fff&size=200', 1),
(4, 2004, '刘小军', '小军', '上海', '上海市', '浦东新区', '门将', 12, '2014-01-08', 150, 42, '右脚', '上海绿地青训俱乐部', '浦东第一小学', '13800002004', 'https://ui-avatars.com/api/?name=刘小军&background=96CEB4&color=fff&size=200', 1);

-- U12二队球员
INSERT INTO players (id, user_id, name, nickname, province, city, district, position, age, birth_date, height, weight, foot, club, school, phone, avatar, status) VALUES
(5, 2005, '陈小龙', '小龙', '上海', '上海市', '浦东新区', '前锋', 12, '2014-09-05', 146, 39, '左脚', '上海绿地青训俱乐部', '浦东第四小学', '13800002005', 'https://ui-avatars.com/api/?name=陈小龙&background=FFEAA7&color=333&size=200', 1),
(6, 2006, '赵小虎', '小虎', '上海', '上海市', '浦东新区', '中场', 12, '2014-11-18', 143, 37, '右脚', '上海绿地青训俱乐部', '浦东第五小学', '13800002006', 'https://ui-avatars.com/api/?name=赵小虎&background=DDA0DD&color=fff&size=200', 1),
(7, 2007, '孙小杰', '小杰', '上海', '上海市', '浦东新区', '后卫', 12, '2014-04-25', 149, 43, '右脚', '上海绿地青训俱乐部', '浦东第二小学', '13800002007', 'https://ui-avatars.com/api/?name=孙小杰&background=98D8C8&color=fff&size=200', 1),
(8, 2008, '周小鹏', '小鹏', '上海', '上海市', '浦东新区', '门将', 12, '2014-02-14', 151, 41, '右脚', '上海绿地青训俱乐部', '浦东第三小学', '13800002008', 'https://ui-avatars.com/api/?name=周小鹏&background=F7DC6F&color=333&size=200', 1);

.print '球员扩展数据导入完成: ' || (SELECT COUNT(*) FROM players) || ' 条'
