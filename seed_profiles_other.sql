-- ============================================================
-- 分析师、教练、球探个人主页虚拟数据
-- 执行: sqlite3 shaonianqiutan.db < seed_profiles_other.sql
-- ============================================================

-- ============================================================
-- 一、更新分析师账号 (ID: 30-33)
-- ============================================================

UPDATE users SET name='陈分析师', nickname='进攻之眼', avatar='/images/avatars/analyst_1.png', gender='男', province='上海', city='上海市' WHERE id=30;

UPDATE analysts SET name='陈分析师', bio='10年足球分析经验，专注进攻战术和技术特点分析。曾为多家中超俱乐部提供球探报告，擅长技术动作拆解和进攻威胁评估。', specialty='["进攻分析", "技术报告", "战术板分析", "球员潜力评估"]', experience=10, profession='前职业球队技术分析师', is_pro_player=0, has_case=1, case_detail='曾为上海绿地、北京国安、广州恒大等多家俱乐部提供球员分析报告超过200份，重点关注年轻球员的技术成长潜力。', contact_phone='13800000030', contact_email='chen.analyst@shaonianqiutan.com', rating=4.8, review_count=156, status='active' WHERE user_id=30;

UPDATE users SET name='林分析师', nickname='防守大师', avatar='/images/avatars/analyst_2.png', gender='男', province='北京', city='北京市' WHERE id=31;

UPDATE analysts SET name='林分析师', bio='8年青训分析经验，专注防守和体能评估。体育科学背景，擅长利用数据分析球员的体能指标和防守位置感。', specialty='["防守分析", "体能报告", "青训评估", "数据建模"]', experience=8, profession='体育科学研究员', is_pro_player=0, has_case=1, case_detail='为多家青训机构提供体能分析报告，开发了一套青少年球员体能评估体系。', contact_phone='13800000031', contact_email='lin.analyst@shaonianqiutan.com', rating=4.6, review_count=98, status='active' WHERE user_id=31;

UPDATE users SET name='周分析师', nickname='门将专家', avatar='/images/avatars/analyst_3.png', gender='男', province='广东', city='广州市' WHERE id=32;

UPDATE analysts SET name='周分析师', bio='5年门将专项分析经验，为多家职业俱乐部评估门将。前职业门将，专注门将位置感和扑救技术分析。', specialty='["门将专项", "位置感分析", "扑救技术", "脚下技术"]', experience=5, profession='前职业门将/门将教练', is_pro_player=1, has_case=1, case_detail='曾任中超球队第二门将，为广州恒大、山东泰山等俱乐部评估年轻门将超过50名。', contact_phone='13800000032', contact_email='zhou.analyst@shaonianqiutan.com', rating=4.7, review_count=72, status='active' WHERE user_id=32;

UPDATE users SET name='吴分析师', nickname='青训教父', avatar='/images/avatars/analyst_4.png', gender='男', province='四川', city='成都市' WHERE id=33;

UPDATE analysts SET name='吴分析师', bio='12年综合评估经验，青训专家，擅长球员潜力评估和成长规划。曾任中超俱乐部青训总监。', specialty='["综合评估", "青训专家", "潜力预测", "成长规划", "战术分析"]', experience=12, profession='前青训总监/国字号教练', is_pro_player=1, has_case=1, case_detail='曾任中超俱乐部青训总监，累计评估青少年球员超过5000人次，培养出多名进入职业联赛的球员。', contact_phone='13800000033', contact_email='wu.analyst@shaonianqiutan.com', rating=4.9, review_count=234, status='active' WHERE user_id=33;

-- ============================================================
-- 二、更新教练账号 (ID: 20-23)
-- ============================================================

UPDATE users SET name='王教练', nickname='冠军教头', avatar='/images/avatars/coach_1.png', gender='男', province='上海', city='上海市' WHERE id=20;

UPDATE coaches SET license_type='A级', license_number='AFC-A-2023-0001', specialties='["技术训练", "青少年培养", "战术指导", "心理辅导"]', bio='20年青训经验，前职业球员。曾带领多支球队获得全国青少年足球联赛冠军，擅长发掘和培养年轻球员的技术天赋。', coaching_years=20, current_club='上海绿地俱乐部', verified=1 WHERE user_id=20;

UPDATE users SET name='李教练', nickname='战术大师', avatar='/images/avatars/coach_2.png', gender='男', province='北京', city='北京市' WHERE id=21;

UPDATE coaches SET license_type='B级', license_number='AFC-B-2023-0002', specialties='["体能训练", "战术指导", "位置训练"]', bio='15年青训经验，专注球员体能和战术素养提升。曾在北京国安各级青年队担任教练，培养出多名国字号球员。', coaching_years=15, current_club='北京国安青训', verified=1 WHERE user_id=21;

UPDATE users SET name='张教练', nickname='技术教父', avatar='/images/avatars/coach_3.png', gender='男', province='山东', city='济南市' WHERE id=22;

UPDATE coaches SET license_type='B级', license_number='AFC-B-2023-0003', specialties='["技术训练", "守门员训练", "个人技术"]', bio='12年青训经验，特别擅长技术动作拆解和守门员专项训练。培养出多名技术型球员和优秀守门员。', coaching_years=12, current_club='山东泰山青训', verified=1 WHERE user_id=22;

UPDATE users SET name='刘教练', nickname='心理专家', avatar='/images/avatars/coach_4.png', gender='男', province='四川', city='成都市' WHERE id=23;

UPDATE coaches SET license_type='C级', license_number='AFC-C-2023-0004', specialties='["青少年培养", "心理辅导", "团队建设"]', bio='10年青训经验，专注青少年心理发展和团队建设。擅长通过足球培养孩子的自信心和团队精神。', coaching_years=10, current_club='成都蓉城青训', verified=1 WHERE user_id=23;

-- ============================================================
-- 三、更新球探账号 (ID: 24-27)
-- ============================================================

UPDATE users SET name='赵球探', nickname='华东猎手', avatar='/images/avatars/scout_1.png', gender='男', province='上海', city='上海市' WHERE id=24;

UPDATE scouts SET scouting_experience='3-5年', specialties='["前锋", "中场", "边锋"]', preferred_age_groups='["U12", "U14", "U16"]', scouting_regions='["华东", "上海", "江苏", "浙江"]', current_organization='自由球探', bio='资深球探，专注华东地区青少年足球人才发掘。拥有丰富的球探网络资源，与多家俱乐部保持良好合作关系。', verified=1, total_discovered=50, total_reports=120, total_adopted=35 WHERE user_id=24;

UPDATE users SET name='陈球探', nickname='华北之眼', avatar='/images/avatars/scout_2.png', gender='男', province='北京', city='北京市' WHERE id=25;

UPDATE scouts SET scouting_experience='1-3年', specialties='["后卫", "中场", "守门员"]', preferred_age_groups='["U10", "U12", "U14"]', scouting_regions='["华北", "北京", "天津", "河北"]', current_organization='北京国安球探部', bio='北京国安球探部成员，专注华北地区青少年足球人才发掘。重点关注后卫和守门员位置的人才。', verified=1, total_discovered=30, total_reports=80, total_adopted=20 WHERE user_id=25;

UPDATE users SET name='周球探', nickname='华南鹰眼', avatar='/images/avatars/scout_3.png', gender='男', province='广东', city='广州市' WHERE id=26;

UPDATE scouts SET scouting_experience='3-5年', specialties='["前锋", "守门员", "中场"]', preferred_age_groups='["U12", "U14", "U16", "U18"]', scouting_regions='["华南", "广东", "广州", "深圳"]', current_organization='广州恒大球探部', bio='广州恒大资深球探，专注华南地区青少年足球人才发掘。特别擅长发现攻击型和守门员位置的人才。', verified=1, total_discovered=80, total_reports=200, total_adopted=60 WHERE user_id=26;

UPDATE users SET name='吴球探', nickname='西南之声', avatar='/images/avatars/scout_4.png', gender='男', province='四川', city='成都市' WHERE id=27;

UPDATE scouts SET scouting_experience='1-3年', specialties='["中场", "后卫", "边锋"]', preferred_age_groups='["U10", "U12", "U14"]', scouting_regions='["西南", "四川", "成都", "重庆"]', current_organization='成都蓉城球探部', bio='成都蓉城球探部成员，专注西南地区青少年足球人才发掘。重点关注技术型中场和边路球员。', verified=1, total_discovered=20, total_reports=50, total_adopted=12 WHERE user_id=27;

-- ============================================================
-- 四、更新俱乐部账号 (ID: 10-19) - clubs表
-- ============================================================

UPDATE clubs SET description='上海绿地青训俱乐部成立于2010年，是上海市领先的青少年足球培训机构。拥有专业教练团队30余人，学员超过500人，多次获得上海市和全国青少年足球比赛冠军。俱乐部秉承"快乐足球、健康成长"的理念，致力于培养全面发展的足球人才。', address='上海市浦东新区世纪大道1000号', contact_name='张明华', contact_phone='13800000010', established_year=2010, club_size='large', member_level='enterprise', free_test_quota=50 WHERE user_id=10;

UPDATE clubs SET description='北京国安足球俱乐部青训基地，隶属于北京国安足球俱乐部。拥有先进的训练设施和专业的教练团队，为北京乃至全国培养优秀足球人才。多名青训球员已进入一线队或转会至其他职业俱乐部。', address='北京市朝阳区工体北路8号', contact_name='李志伟', contact_phone='13800000011', established_year=2002, club_size='large', member_level='enterprise', free_test_quota=50 WHERE user_id=11;

UPDATE clubs SET description='广州恒大足球学校，全国知名青训机构，采用西班牙皇马青训体系。占地1000余亩，拥有世界级的训练设施。已向各级国字号球队输送多名球员，是中国足球青训的标杆。', address='广东省清远市恒大足球学校', contact_name='王强', contact_phone='13800000012', established_year=2012, club_size='large', member_level='enterprise', free_test_quota=50 WHERE user_id=12;

UPDATE clubs SET description='山东泰山足球俱乐部青训体系，山东省历史最悠久的职业足球俱乐部青训机构。依托俱乐部完善的后备人才培养体系，为山东足球和中国足球培养了大量人才。', address='山东省济南市经十路17688号', contact_name='刘建华', contact_phone='13800000013', established_year=1993, club_size='large', member_level='professional', free_test_quota=30 WHERE user_id=13;

UPDATE clubs SET description='江苏苏宁足球俱乐部青训基地，依托俱乐部资源建立的专业青训机构。注重技术培养和战术素养提升，在江苏省青少年足球赛事中表现优异。', address='江苏省南京市江宁区苏宁足球训练基地', contact_name='陈晓东', contact_phone='13800000014', established_year=1994, club_size='medium', member_level='professional', free_test_quota=30 WHERE user_id=14;

UPDATE clubs SET description='成都蓉城足球俱乐部青训，西部地区的优秀青训代表。拥有专业的外籍教练团队和先进的训练理念，专注于技术型球员的培养，在多项全国青少年赛事中取得佳绩。', address='四川省成都市双流区足球训练基地', contact_name='赵伟', contact_phone='13800000015', established_year=2018, club_size='medium', member_level='professional', free_test_quota=30 WHERE user_id=15;

UPDATE clubs SET description='武汉三镇足球俱乐部青训，近年来崛起的青训新锐。依托俱乐部金元支持，引入先进的训练体系，在青少年足球培养方面发展迅速。', address='湖北省武汉市汉阳区三镇足球中心', contact_name='杨海', contact_phone='13800000016', established_year=2013, club_size='medium', member_level='professional', free_test_quota=30 WHERE user_id=16;

UPDATE clubs SET description='浙江绿城足球学校，浙江省历史最悠久的职业俱乐部青训机构。注重青训体系建设，多年来为各级国家队和职业俱乐部输送了大量优秀球员。', address='浙江省杭州市余杭区绿城足球训练基地', contact_name='吴斌', contact_phone='13800000017', established_year=2004, club_size='large', member_level='professional', free_test_quota=30 WHERE user_id=17;

UPDATE clubs SET description='河南嵩山龙门足球俱乐部青训，中原地区重要的青训基地。依托俱乐部资源，建立完善的青少年培养体系，在华中地区青少年足球版图中占有重要地位。', address='河南省郑州市郑东新区嵩山足球中心', contact_name='周志刚', contact_phone='13800000018', established_year=1994, club_size='medium', member_level='professional', free_test_quota=30 WHERE user_id=18;

UPDATE clubs SET description='天津津门虎足球俱乐部青训，中国足球发源地天津的代表性青训机构。传承天津足球深厚底蕴，注重技术和意识培养，为天津足球培养接班人。', address='天津市滨海新区津门虎足球训练基地', contact_name='马晓磊', contact_phone='13800000019', established_year=1994, club_size='medium', member_level='professional', free_test_quota=30 WHERE user_id=19;

SELECT '分析师、教练、球探、俱乐部资料更新完成!' as result;
