-- ============================================================
-- 少年球探 - 球员个人主页虚拟数据
-- 生成时间: 2026-04-10
-- 执行: sqlite3 shaonianqiutan.db < seed_profile_data.sql
-- ============================================================

-- 更新球员账号 (ID: 1001-1020) - users表

UPDATE users SET name='王小明', nickname='小小前锋', avatar='/images/avatars/player_1.png', gender='男', birth_date='2014-03-15', age=12, height=155, weight=42, foot='右脚', position='前锋', second_position='左边锋', province='上海', city='上海市', club='上海绿地俱乐部', start_year=2019, fa_registered=1, association='上海市足球协会', jersey_color='绿色', jersey_number=10 WHERE id=1001;

UPDATE users SET name='李小强', nickname='中场指挥官', avatar='/images/avatars/player_2.png', gender='男', birth_date='2014-05-20', age=11, height=152, weight=40, foot='右脚', position='中场', second_position='后腰', province='上海', city='上海市', club='上海绿地俱乐部', start_year=2019, fa_registered=1, association='上海市足球协会', jersey_color='绿色', jersey_number=8 WHERE id=1002;

UPDATE users SET name='张小刚', nickname='铁闸', avatar='/images/avatars/player_3.png', gender='男', birth_date='2014-01-08', age=12, height=160, weight=45, foot='右脚', position='后卫', second_position='中后卫', province='上海', city='上海市', club='上海绿地俱乐部', start_year=2018, fa_registered=1, association='上海市足球协会', jersey_color='绿色', jersey_number=5 WHERE id=1003;

UPDATE users SET name='刘小军', nickname='门神', avatar='/images/avatars/player_4.png', gender='男', birth_date='2014-07-22', age=11, height=158, weight=48, foot='右脚', position='门将', province='上海', city='上海市', club='上海绿地俱乐部', start_year=2019, fa_registered=1, association='上海市足球协会', jersey_color='绿色', jersey_number=1 WHERE id=1004;

UPDATE users SET name='陈小龙', nickname='快马', avatar='/images/avatars/player_5.png', gender='男', birth_date='2014-09-10', age=11, height=150, weight=38, foot='左脚', position='前锋', second_position='右边锋', province='上海', city='上海市', club='上海绿地俱乐部', start_year=2020, fa_registered=0, association='上海市足球协会', jersey_color='绿色', jersey_number=11 WHERE id=1005;

UPDATE users SET name='赵小虎', nickname='中场发动机', avatar='/images/avatars/player_6.png', gender='男', birth_date='2014-11-05', age=11, height=148, weight=39, foot='右脚', position='中场', second_position='前腰', province='上海', city='上海市', club='上海绿地俱乐部', start_year=2020, fa_registered=0, association='上海市足球协会', jersey_color='绿色', jersey_number=7 WHERE id=1006;

UPDATE users SET name='孙小杰', nickname='铁卫', avatar='/images/avatars/player_7.png', gender='男', birth_date='2014-04-18', age=11, height=156, weight=44, foot='右脚', position='后卫', second_position='右后卫', province='北京', city='北京市', club='北京国安青训', start_year=2019, fa_registered=1, association='北京市足球协会', jersey_color='蓝色', jersey_number=4 WHERE id=1007;

UPDATE users SET name='周小鹏', nickname='中场大脑', avatar='/images/avatars/player_8.png', gender='男', birth_date='2014-06-25', age=11, height=153, weight=41, foot='右脚', position='中场', second_position='后腰', province='北京', city='北京市', club='北京国安青训', start_year=2019, fa_registered=1, association='北京市足球协会', jersey_color='蓝色', jersey_number=6 WHERE id=1008;

UPDATE users SET name='吴小峰', nickname='射门机器', avatar='/images/avatars/player_9.png', gender='男', birth_date='2014-02-14', age=12, height=154, weight=43, foot='右脚', position='前锋', second_position='中锋', province='广东', city='广州市', club='广州恒大足校', start_year=2018, fa_registered=1, association='广东省足球协会', jersey_color='红色', jersey_number=9 WHERE id=1009;

UPDATE users SET name='郑小磊', nickname='最后一道盾', avatar='/images/avatars/player_10.png', gender='男', birth_date='2014-08-30', age=11, height=162, weight=50, foot='右脚', position='门将', province='广东', city='广州市', club='广州恒大足校', start_year=2020, fa_registered=0, association='广东省足球协会', jersey_color='红色', jersey_number=1 WHERE id=1010;

UPDATE users SET name='王大雷', nickname='中场节拍器', avatar='/images/avatars/player_11.png', gender='男', birth_date='2014-03-08', age=12, height=151, weight=40, foot='右脚', position='中场', second_position='前腰', province='山东', city='济南市', club='山东泰山青训', start_year=2019, fa_registered=1, association='山东省足球协会', jersey_color='橙色', jersey_number=8 WHERE id=1011;

UPDATE users SET name='武磊', nickname='锋线杀手', avatar='/images/avatars/player_12.png', gender='男', birth_date='2014-10-12', age=11, height=149, weight=37, foot='右脚', position='前锋', second_position='左边锋', province='山东', city='济南市', club='山东泰山青训', start_year=2020, fa_registered=0, association='山东省足球协会', jersey_color='橙色', jersey_number=10 WHERE id=1012;

UPDATE users SET name='李明', nickname='后防中坚', avatar='/images/avatars/player_13.png', gender='男', birth_date='2014-05-05', age=11, height=157, weight=44, foot='右脚', position='后卫', second_position='中后卫', province='江苏', city='南京市', club='江苏苏宁青训', start_year=2019, fa_registered=1, association='江苏省足球协会', jersey_color='紫色', jersey_number=5 WHERE id=1013;

UPDATE users SET name='张伟', nickname='中场调度员', avatar='/images/avatars/player_14.png', gender='男', birth_date='2014-12-20', age=11, height=150, weight=39, foot='左脚', position='中场', second_position='边前卫', province='江苏', city='南京市', club='江苏苏宁青训', start_year=2020, fa_registered=0, association='江苏省足球协会', jersey_color='紫色', jersey_number=7 WHERE id=1014;

UPDATE users SET name='刘川', nickname='四川射手', avatar='/images/avatars/player_15.png', gender='男', birth_date='2014-07-15', age=11, height=152, weight=41, foot='右脚', position='前锋', second_position='影子前锋', province='四川', city='成都市', club='成都蓉城青训', start_year=2019, fa_registered=1, association='四川省足球协会', jersey_color='橙色', jersey_number=9 WHERE id=1015;

UPDATE users SET name='陈翔', nickname='蓉城铁卫', avatar='/images/avatars/player_16.png', gender='男', birth_date='2014-01-28', age=12, height=159, weight=46, foot='右脚', position='后卫', second_position='右后卫', province='四川', city='成都市', club='成都蓉城青训', start_year=2018, fa_registered=1, association='四川省足球协会', jersey_color='橙色', jersey_number=3 WHERE id=1016;

UPDATE users SET name='杨帆', nickname='武汉中场核心', avatar='/images/avatars/player_17.png', gender='男', birth_date='2014-09-03', age=11, height=154, weight=42, foot='右脚', position='中场', second_position='后腰', province='湖北', city='武汉市', club='武汉三镇青训', start_year=2019, fa_registered=1, association='湖北省足球协会', jersey_color='黄色', jersey_number=6 WHERE id=1017;

UPDATE users SET name='周涛', nickname='江城门神', avatar='/images/avatars/player_18.png', gender='男', birth_date='2014-04-12', age=11, height=160, weight=49, foot='右脚', position='门将', province='湖北', city='武汉市', club='武汉三镇青训', start_year=2020, fa_registered=0, association='湖北省足球协会', jersey_color='黄色', jersey_number=1 WHERE id=1018;

UPDATE users SET name='吴俊', nickname='浙江快马', avatar='/images/avatars/player_19.png', gender='男', birth_date='2014-11-25', age=11, height=148, weight=38, foot='左脚', position='前锋', second_position='边锋', province='浙江', city='杭州市', club='浙江绿城足校', start_year=2020, fa_registered=0, association='浙江省足球协会', jersey_color='绿色', jersey_number=11 WHERE id=1019;

UPDATE users SET name='郑鑫', nickname='绿城铁闸', avatar='/images/avatars/player_20.png', gender='男', birth_date='2014-06-08', age=11, height=158, weight=45, foot='右脚', position='后卫', second_position='左后卫', province='浙江', city='杭州市', club='浙江绿城足校', start_year=2019, fa_registered=1, association='浙江省足球协会', jersey_color='绿色', jersey_number=4 WHERE id=1020;

-- players表更新

UPDATE players SET name='王小明', nickname='小小前锋', province='上海', city='上海市', district='浦东新区', position='前锋', age=12, birth_date='2014-03-15', height=155, weight=42, foot='右脚', club='上海绿地俱乐部', school='上海市浦东新区第一小学', phone='13800011001' WHERE user_id=1001;

UPDATE players SET name='李小强', nickname='中场指挥官', province='上海', city='上海市', district='徐汇区', position='中场', age=11, birth_date='2014-05-20', height=152, weight=40, foot='右脚', club='上海绿地俱乐部', school='上海市徐汇区逸夫小学', phone='13800011002' WHERE user_id=1002;

UPDATE players SET name='张小刚', nickname='铁闸', province='上海', city='上海市', district='杨浦区', position='后卫', age=12, birth_date='2014-01-08', height=160, weight=45, foot='右脚', club='上海绿地俱乐部', school='上海市杨浦区打虎山路第一小学', phone='13800011003' WHERE user_id=1003;

UPDATE players SET name='刘小军', nickname='门神', province='上海', city='上海市', district='静安区', position='门将', age=11, birth_date='2014-07-22', height=158, weight=48, foot='右脚', club='上海绿地俱乐部', school='上海市静安区第一师范学校附属小学', phone='13800011004' WHERE user_id=1004;

UPDATE players SET name='陈小龙', nickname='快马', province='上海', city='上海市', district='闵行区', position='前锋', age=11, birth_date='2014-09-10', height=150, weight=38, foot='左脚', club='上海绿地俱乐部', school='上海市闵行区实验小学', phone='13800011005' WHERE user_id=1005;

UPDATE players SET name='赵小虎', nickname='中场发动机', province='上海', city='上海市', district='宝山区', position='中场', age=11, birth_date='2014-11-05', height=148, weight=39, foot='右脚', club='上海绿地俱乐部', school='上海市宝山区红星小学', phone='13800011006' WHERE user_id=1006;

UPDATE players SET name='孙小杰', nickname='铁卫', province='北京', city='北京市', district='朝阳区', position='后卫', age=11, birth_date='2014-04-18', height=156, weight=44, foot='右脚', club='北京国安青训', school='北京市朝阳区白家庄小学', phone='13800011007' WHERE user_id=1007;

UPDATE players SET name='周小鹏', nickname='中场大脑', province='北京', city='北京市', district='海淀区', position='中场', age=11, birth_date='2014-06-25', height=153, weight=41, foot='右脚', club='北京国安青训', school='北京市海淀区中关村第三小学', phone='13800011008' WHERE user_id=1008;

UPDATE players SET name='吴小峰', nickname='射门机器', province='广东', city='广州市', district='越秀区', position='前锋', age=12, birth_date='2014-02-14', height=154, weight=43, foot='右脚', club='广州恒大足校', school='广州市越秀区东风东路小学', phone='13800011009' WHERE user_id=1009;

UPDATE players SET name='郑小磊', nickname='最后一道盾', province='广东', city='广州市', district='天河区', position='门将', age=11, birth_date='2014-08-30', height=162, weight=50, foot='右脚', club='广州恒大足校', school='广州市天河区华南师范大学附属小学', phone='13800011010' WHERE user_id=1010;

UPDATE players SET name='王大雷', nickname='中场节拍器', province='山东', city='济南市', district='市中区', position='中场', age=12, birth_date='2014-03-08', height=151, weight=40, foot='右脚', club='山东泰山青训', school='济南市市中区胜利大街小学', phone='13800011011' WHERE user_id=1011;

UPDATE players SET name='武磊', nickname='锋线杀手', province='山东', city='济南市', district='历下区', position='前锋', age=11, birth_date='2014-10-12', height=149, weight=37, foot='右脚', club='山东泰山青训', school='济南市历下区解放路第一小学', phone='13800011012' WHERE user_id=1012;

UPDATE players SET name='李明', nickname='后防中坚', province='江苏', city='南京市', district='鼓楼区', position='后卫', age=11, birth_date='2014-05-05', height=157, weight=44, foot='右脚', club='江苏苏宁青训', school='南京市鼓楼区力学小学', phone='13800011013' WHERE user_id=1013;

UPDATE players SET name='张伟', nickname='中场调度员', province='江苏', city='南京市', district='玄武区', position='中场', age=11, birth_date='2014-12-20', height=150, weight=39, foot='左脚', club='江苏苏宁青训', school='南京市玄武区长江路小学', phone='13800011014' WHERE user_id=1014;

UPDATE players SET name='刘川', nickname='四川射手', province='四川', city='成都市', district='武侯区', position='前锋', age=11, birth_date='2014-07-15', height=152, weight=41, foot='右脚', club='成都蓉城青训', school='成都市武侯区四川大学附属实验小学', phone='13800011015' WHERE user_id=1015;

UPDATE players SET name='陈翔', nickname='蓉城铁卫', province='四川', city='成都市', district='锦江区', position='后卫', age=12, birth_date='2014-01-28', height=159, weight=46, foot='右脚', club='成都蓉城青训', school='成都市锦江区盐道街小学', phone='13800011016' WHERE user_id=1016;

UPDATE players SET name='杨帆', nickname='武汉中场核心', province='湖北', city='武汉市', district='江岸区', position='中场', age=11, birth_date='2014-09-03', height=154, weight=42, foot='右脚', club='武汉三镇青训', school='武汉市江岸区长春街小学', phone='13800011017' WHERE user_id=1017;

UPDATE players SET name='周涛', nickname='江城门神', province='湖北', city='武汉市', district='洪山区', position='门将', age=11, birth_date='2014-04-12', height=160, weight=49, foot='右脚', club='武汉三镇青训', school='武汉市洪山区第一小学', phone='13800011018' WHERE user_id=1018;

UPDATE players SET name='吴俊', nickname='浙江快马', province='浙江', city='杭州市', district='西湖区', position='前锋', age=11, birth_date='2014-11-25', height=148, weight=38, foot='左脚', club='浙江绿城足校', school='杭州市西湖区学军小学', phone='13800011019' WHERE user_id=1019;

UPDATE players SET name='郑鑫', nickname='绿城铁闸', province='浙江', city='杭州市', district='滨江区', position='后卫', age=11, birth_date='2014-06-08', height=158, weight=45, foot='右脚', club='浙江绿城足校', school='杭州市滨江区江南实验小学', phone='13800011020' WHERE user_id=1020;

SELECT '球员资料更新完成!' as result;
