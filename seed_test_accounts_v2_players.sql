-- ============================================================
-- 少年球探 - 测试账号数据 Part 2 - 球员和关联
-- 生成时间: 2026-04-08
-- ============================================================

-- 5. 先在players表创建数据 (使用user_id作为关联)
INSERT INTO players (user_id, name, nickname, province, city, district, position, age, birth_date, height, weight, foot, club, school, phone, avatar, status, create_time, update_time) VALUES
(2001, '王小明', '小明', '上海', '上海市', '浦东新区', '前锋', 11, '2014-03-15', 145, 38, '右脚', '上海绿地俱乐部', '浦东实验小学', '13800002001', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2001', 1, datetime('now'), datetime('now')),
(2002, '李小强', '小强', '上海', '上海市', '徐汇区', '中场', 11, '2014-06-20', 148, 40, '左脚', '上海绿地俱乐部', '徐汇实验小学', '13800002002', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2002', 1, datetime('now'), datetime('now')),
(2003, '张小刚', '小刚', '上海', '上海市', '黄浦区', '后卫', 11, '2014-01-10', 150, 42, '右脚', '上海绿地俱乐部', '黄浦实验小学', '13800002003', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2003', 1, datetime('now'), datetime('now')),
(2004, '刘小军', '小军', '上海', '上海市', '静安区', '门将', 11, '2014-09-05', 152, 45, '右脚', '上海绿地俱乐部', '静安实验小学', '13800002004', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2004', 1, datetime('now'), datetime('now')),
(2005, '陈小龙', '小龙', '上海', '上海市', '长宁区', '前锋', 11, '2014-04-12', 144, 37, '右脚', '上海绿地俱乐部', '长宁实验小学', '13800002005', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2005', 1, datetime('now'), datetime('now')),
(2006, '赵小虎', '小虎', '上海', '上海市', '虹口区', '中场', 11, '2014-07-28', 146, 39, '双脚', '上海绿地俱乐部', '虹口实验小学', '13800002006', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2006', 1, datetime('now'), datetime('now')),
(2007, '孙小杰', '小杰', '北京', '北京市', '朝阳区', '后卫', 11, '2014-02-18', 147, 41, '右脚', '北京国安青训', '朝阳实验小学', '13800002007', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2007', 1, datetime('now'), datetime('now')),
(2008, '周小鹏', '小鹏', '北京', '北京市', '海淀区', '中场', 11, '2014-05-22', 149, 43, '左脚', '北京国安青训', '海淀实验小学', '13800002008', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2008', 1, datetime('now'), datetime('now')),
(2009, '吴小峰', '小峰', '广东', '广州市', '天河区', '前锋', 11, '2014-03-08', 146, 39, '右脚', '广州恒大足校', '天河实验小学', '13800002009', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2009', 1, datetime('now'), datetime('now')),
(2010, '郑小磊', '小磊', '广东', '广州市', '白云区', '门将', 11, '2014-08-30', 151, 44, '右脚', '广州恒大足校', '白云实验小学', '13800002010', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2010', 1, datetime('now'), datetime('now')),
(2011, '王大雷', '大雷', '山东', '济南市', '历下区', '中场', 11, '2014-04-15', 148, 41, '左脚', '山东泰山青训', '历下实验小学', '13800002011', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2011', 1, datetime('now'), datetime('now')),
(2012, '武磊', '武磊', '山东', '济南市', '市中区', '前锋', 11, '2014-07-22', 147, 40, '右脚', '山东泰山青训', '市中实验小学', '13800002012', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2012', 1, datetime('now'), datetime('now')),
(2013, '李明', '李明', '江苏', '南京市', '鼓楼区', '后卫', 11, '2014-01-05', 146, 38, '右脚', '江苏苏宁青训', '鼓楼实验小学', '13800002013', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2013', 1, datetime('now'), datetime('now')),
(2014, '张伟', '张伟', '江苏', '南京市', '玄武区', '中场', 11, '2014-06-18', 149, 42, '双脚', '江苏苏宁青训', '玄武实验小学', '13800002014', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2014', 1, datetime('now'), datetime('now')),
(2015, '刘川', '刘川', '四川', '成都市', '锦江区', '前锋', 11, '2014-03-25', 145, 39, '右脚', '成都蓉城青训', '锦江实验小学', '13800002015', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2015', 1, datetime('now'), datetime('now')),
(2016, '陈翔', '陈翔', '四川', '成都市', '青羊区', '后卫', 11, '2014-09-12', 150, 43, '左脚', '成都蓉城青训', '青羊实验小学', '13800002016', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2016', 1, datetime('now'), datetime('now')),
(2017, '杨帆', '杨帆', '湖北', '武汉市', '江汉区', '中场', 11, '2014-02-28', 147, 40, '右脚', '武汉三镇青训', '江汉实验小学', '13800002017', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2017', 1, datetime('now'), datetime('now')),
(2018, '周涛', '周涛', '湖北', '武汉市', '武昌区', '门将', 11, '2014-05-15', 148, 41, '右脚', '武汉三镇青训', '武昌实验小学', '13800002018', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2018', 1, datetime('now'), datetime('now')),
(2019, '吴俊', '吴俊', '浙江', '杭州市', '西湖区', '前锋', 11, '2014-04-20', 146, 39, '双脚', '浙江绿城足校', '西湖实验小学', '13800002019', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2019', 1, datetime('now'), datetime('now')),
(2020, '郑鑫', '郑鑫', '浙江', '杭州市', '滨江区', '后卫', 11, '2014-08-08', 149, 42, '右脚', '浙江绿城足校', '滨江实验小学', '13800002020', 'https://api.dicebear.com/7.x/avataaars/svg?seed=2020', 1, datetime('now'), datetime('now'));

-- 6. 球队-球员关联
-- 先清理旧数据
DELETE FROM team_players;
DELETE FROM team_coaches;

-- 球队-球员关联 (player_id = players.id, user_id = users.id)
INSERT INTO team_players (team_id, user_id, player_id, jersey_number, position, status, joined_at, created_at, updated_at) VALUES
-- 上海绿地 U12一队 (ID=1)
(1, 2001, 201, '9', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2002, 202, '8', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2003, 203, '4', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(1, 2004, 204, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 上海绿地 U12二队 (ID=2)
(2, 2005, 205, '11', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(2, 2006, 206, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 北京国安 (ID=3)
(3, 2007, 207, '3', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(3, 2008, 208, '7', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 广州恒大 (ID=4)
(4, 2009, 209, '10', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(4, 2010, 210, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 山东泰山 (ID=5)
(5, 2011, 211, '8', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(5, 2012, 212, '9', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 江苏苏宁 (ID=6)
(6, 2013, 213, '4', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
(6, 2014, 214, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 成都蓉城 (ID=7)
(7, 2015, 215, '10', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(7, 2016, 216, '3', '后卫', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 武汉三镇 (ID=8)
(8, 2017, 217, '6', '中场', 'active', datetime('now'), datetime('now'), datetime('now')),
(8, 2018, 218, '1', '门将', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 浙江绿城 (ID=9)
(9, 2019, 219, '11', '前锋', 'active', datetime('now'), datetime('now'), datetime('now')),
(9, 2020, 220, '2', '后卫', 'active', datetime('now'), datetime('now'), datetime('now'));

-- 7. 球队-教练关联
INSERT INTO team_coaches (team_id, user_id, role, status, joined_at, created_at, updated_at) VALUES
-- 上海绿地 2支球队 -> 王教练 (user_id=20)
(1, 20, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(2, 20, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 北京国安 + 广州恒大 -> 李教练 (user_id=21)
(3, 21, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(4, 21, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 山东泰山 + 江苏苏宁 -> 张教练 (user_id=22)
(5, 22, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(6, 22, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
-- 成都蓉城 + 武汉三镇 + 浙江绿城 -> 刘教练 (user_id=23)
(7, 23, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(8, 23, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now')),
(9, 23, 'head_coach', 'active', datetime('now'), datetime('now'), datetime('now'));

SELECT '球员和关联数据插入完成!' as result;
