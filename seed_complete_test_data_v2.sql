-- ============================================================
-- 少年球探 - 完整测试数据补充脚本 V2
-- 为周报功能涉及的角色补充详细资料
-- ============================================================

-- ============================================================
-- 1. 完善8名测试球员在users表的详细信息
-- ============================================================
UPDATE users SET 
    nickname = '王小明',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=wangxiaoming',
    birth_date = '2014-03-15',
    age = 11,
    gender = '男',
    height = 145,
    weight = 38,
    foot = '右脚',
    position = '前锋',
    second_position = '边锋',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2020,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '蓝色',
    jersey_number = 9,
    father_height = 175,
    father_phone = '13900111101',
    father_edu = '本科',
    father_job = '工程师',
    father_athlete = 0,
    mother_height = 162,
    mother_phone = '13900111102',
    mother_edu = '本科',
    mother_job = '教师',
    mother_athlete = 0,
    updated_at = datetime('now')
WHERE id = 1001;

UPDATE users SET 
    nickname = '李小强',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=lixiaoqiang',
    birth_date = '2014-06-20',
    age = 11,
    gender = '男',
    height = 148,
    weight = 40,
    foot = '左脚',
    position = '中场',
    second_position = '前腰',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2019,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '蓝色',
    jersey_number = 8,
    father_height = 178,
    father_phone = '13900222201',
    father_edu = '硕士',
    father_job = '企业高管',
    father_athlete = 1,
    mother_height = 165,
    mother_phone = '13900222202',
    mother_edu = '硕士',
    mother_job = '医生',
    mother_athlete = 0,
    updated_at = datetime('now')
WHERE id = 1002;

UPDATE users SET 
    nickname = '张小刚',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhangxiaogang',
    birth_date = '2014-01-10',
    age = 11,
    gender = '男',
    height = 150,
    weight = 42,
    foot = '右脚',
    position = '后卫',
    second_position = '中后卫',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2020,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '蓝色',
    jersey_number = 4,
    father_height = 180,
    father_phone = '13900333301',
    father_edu = '本科',
    father_job = '销售经理',
    father_athlete = 0,
    mother_height = 163,
    mother_phone = '13900333302',
    mother_edu = '本科',
    mother_job = '会计师',
    mother_athlete = 0,
    updated_at = datetime('now')
WHERE id = 1003;

UPDATE users SET 
    nickname = '刘小军',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=liuxiaojun',
    birth_date = '2014-09-05',
    age = 11,
    gender = '男',
    height = 152,
    weight = 45,
    foot = '右脚',
    position = '门将',
    second_position = '',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2021,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '黄色',
    jersey_number = 1,
    father_height = 182,
    father_phone = '13900444401',
    father_edu = '本科',
    father_job = '警察',
    father_athlete = 1,
    mother_height = 168,
    mother_phone = '13900444402',
    mother_edu = '本科',
    mother_job = '律师',
    mother_athlete = 0,
    updated_at = datetime('now')
WHERE id = 1004;

UPDATE users SET 
    nickname = '陈小龙',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=chenxiaolong',
    birth_date = '2014-04-12',
    age = 11,
    gender = '男',
    height = 144,
    weight = 37,
    foot = '右脚',
    position = '前锋',
    second_position = '影锋',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2020,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '红色',
    jersey_number = 11,
    father_height = 176,
    father_phone = '13900555501',
    father_edu = '大专',
    father_job = '个体经营',
    father_athlete = 0,
    mother_height = 160,
    mother_phone = '13900555502',
    mother_edu = '大专',
    mother_job = '家庭主妇',
    mother_athlete = 0,
    updated_at = datetime('now')
WHERE id = 1005;

UPDATE users SET 
    nickname = '赵小虎',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhaoxiaohu',
    birth_date = '2014-07-28',
    age = 11,
    gender = '男',
    height = 146,
    weight = 39,
    foot = '双脚',
    position = '中场',
    second_position = '后腰',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2019,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '红色',
    jersey_number = 6,
    father_height = 174,
    father_phone = '13900666601',
    father_edu = '本科',
    father_job = 'IT工程师',
    father_athlete = 0,
    mother_height = 161,
    mother_phone = '13900666602',
    mother_edu = '本科',
    mother_job = '设计师',
    mother_athlete = 0,
    updated_at = datetime('now')
WHERE id = 1006;

UPDATE users SET 
    nickname = '马小军',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=maxiaojun',
    birth_date = '2012-11-15',
    age = 13,
    gender = '男',
    height = 158,
    weight = 48,
    foot = '左脚',
    position = '前锋',
    second_position = '左边锋',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2018,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '白色',
    jersey_number = 7,
    father_height = 179,
    father_phone = '13900777701',
    father_edu = '硕士',
    father_job = '大学教授',
    father_athlete = 1,
    mother_height = 166,
    mother_phone = '13900777702',
    mother_edu = '博士',
    mother_job = '研究员',
    mother_athlete = 1,
    updated_at = datetime('now')
WHERE id = 1007;

UPDATE users SET 
    nickname = '周小杰',
    avatar = 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhouxiaojie',
    birth_date = '2012-08-03',
    age = 13,
    gender = '男',
    height = 160,
    weight = 50,
    foot = '右脚',
    position = '后卫',
    second_position = '右后卫',
    country = '中国',
    province = '上海',
    city = '上海市',
    club = '上海绿地俱乐部',
    start_year = 2018,
    fa_registered = 1,
    association = '上海市足球协会',
    jersey_color = '白色',
    jersey_number = 2,
    father_height = 177,
    father_phone = '13900888801',
    father_edu = '本科',
    father_job = '建筑师',
    father_athlete = 0,
    mother_height = 164,
    mother_phone = '13900888802',
    mother_edu = '本科',
    mother_job = '银行职员',
    mother_athlete = 0,
    updated_at = datetime('now')
WHERE id = 1008;

-- ============================================================
-- 2. 在players表创建8名新球员的数据
-- ============================================================
INSERT OR REPLACE INTO players (id, user_id, name, nickname, province, city, district, position, age, birth_date, height, weight, foot, club, school, phone, avatar, video_url, status, create_time, update_time) VALUES
(101, 1001, '王小明', '小明', '上海', '上海市', '浦东新区', '前锋', 11, '2014-03-15', 145, 38, '右脚', '上海绿地俱乐部', '浦东实验小学', '13900111101', 'https://api.dicebear.com/7.x/avataaars/svg?seed=wangxiaoming', '', 1, datetime('now'), datetime('now')),
(102, 1002, '李小强', '小强', '上海', '上海市', '徐汇区', '中场', 11, '2014-06-20', 148, 40, '左脚', '上海绿地俱乐部', '徐汇实验小学', '13900222201', 'https://api.dicebear.com/7.x/avataaars/svg?seed=lixiaoqiang', '', 1, datetime('now'), datetime('now')),
(103, 1003, '张小刚', '小刚', '上海', '上海市', '黄浦区', '后卫', 11, '2014-01-10', 150, 42, '右脚', '上海绿地俱乐部', '黄浦实验小学', '13900333301', 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhangxiaogang', '', 1, datetime('now'), datetime('now')),
(104, 1004, '刘小军', '小军', '上海', '上海市', '静安区', '门将', 11, '2014-09-05', 152, 45, '右脚', '上海绿地俱乐部', '静安实验小学', '13900444401', 'https://api.dicebear.com/7.x/avataaars/svg?seed=liuxiaojun', '', 1, datetime('now'), datetime('now')),
(105, 1005, '陈小龙', '小龙', '上海', '上海市', '长宁区', '前锋', 11, '2014-04-12', 144, 37, '右脚', '上海绿地俱乐部', '长宁实验小学', '13900555501', 'https://api.dicebear.com/7.x/avataaars/svg?seed=chenxiaolong', '', 1, datetime('now'), datetime('now')),
(106, 1006, '赵小虎', '小虎', '上海', '上海市', '虹口区', '中场', 11, '2014-07-28', 146, 39, '双脚', '上海绿地俱乐部', '虹口实验小学', '13900666601', 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhaoxiaohu', '', 1, datetime('now'), datetime('now')),
(107, 1007, '马小军', '小军', '上海', '上海市', '杨浦区', '前锋', 13, '2012-11-15', 158, 48, '左脚', '上海绿地俱乐部', '杨浦实验中学', '13900777701', 'https://api.dicebear.com/7.x/avataaars/svg?seed=maxiaojun', '', 1, datetime('now'), datetime('now')),
(108, 1008, '周小杰', '小杰', '上海', '上海市', '普陀区', '后卫', 13, '2012-08-03', 160, 50, '右脚', '上海绿地俱乐部', '普陀实验中学', '13900888801', 'https://api.dicebear.com/7.x/avataaars/svg?seed=zhouxiaojie', '', 1, datetime('now'), datetime('now'));

-- ============================================================
-- 3. 完善教练666的详细信息
-- ============================================================
UPDATE coaches SET 
    license_type = '亚足联A级',
    license_number = 'AFC-A-2023-001',
    specialties = '["青训培养", "技术训练", "战术指导", "门将训练"]',
    bio = '20年青训经验，前职业球员，曾效力于上海申花队。退役后专注于青少年足球培训，培养出多名国青队球员。擅长技术细节打磨和比赛阅读能力培养。',
    coaching_years = 20,
    current_club = '上海绿地俱乐部',
    verified = 1,
    updated_at = datetime('now')
WHERE user_id = 666;

-- ============================================================
-- 4. 完善俱乐部777的详细信息
-- ============================================================
UPDATE clubs SET 
    logo = 'https://api.dicebear.com/7.x/identicon/svg?seed=shanghailvdi',
    description = '上海绿地青训俱乐部成立于2010年，是上海市领先的青少年足球培训机构。我们拥有专业的教练团队和完善的训练设施，致力于培养优秀的足球人才。俱乐部已获得上海市足协认证，是市级青训示范基地。',
    address = '上海市浦东新区世纪大道1000号绿地足球训练基地',
    contact_name = '张经理',
    contact_phone = '021-58888888',
    established_year = 2010,
    club_size = 'large',
    member_level = 'enterprise',
    free_test_quota = 50,
    updated_at = datetime('now')
WHERE user_id = 777;

-- ============================================================
-- 5. 创建俱乐部777的主页数据
-- ============================================================
INSERT OR REPLACE INTO club_homes (club_id, hero, about, contact, created_at, updated_at) 
VALUES (
    1,
    '{"title":"上海绿地青训俱乐部","subtitle":"专注青少年足球培训 · 成就未来足球之星","coverImage":"https://images.unsplash.com/photo-1574629810360-7efbbe195018?w=1200","stats":[{"label":"在训球员","value":"200+"},{"label":"专业教练","value":"15"},{"label":"成立年份","value":"2010"},{"label":"获得荣誉","value":"30+"}]}',
    '{"title":"关于我们","content":"上海绿地青训俱乐部成立于2010年，专注于8-16岁青少年足球培训。我们秉承\"技术为本、快乐足球\"的理念，通过科学的训练体系和专业的教练团队，帮助每一位学员实现足球梦想。","features":[{"icon":"Trophy","title":"专业认证","desc":"上海市足协认证青训机构"},{"icon":"Users","title":"精英教练","desc":"亚足联A级教练领衔"},{"icon":"Target","title":"科学训练","desc":"德国青训体系引进"},{"icon":"Home","title":"完善设施","desc":"标准11人制天然草球场"}],"images":["https://images.unsplash.com/photo-1517466787929-bc90951d0974?w=600","https://images.unsplash.com/photo-1560272564-c83b66b1ad12?w=600","https://images.unsplash.com/photo-1574629810360-7efbbe195018?w=600"]}',
    '{"address":"上海市浦东新区世纪大道1000号绿地足球训练基地","phone":"021-58888888","email":"contact@shanghailvdi.com","wechat":"lvdiqinxun","mapLocation":{"lat":31.2304,"lng":121.4737}}',
    datetime('now'),
    datetime('now')
);

-- ============================================================
-- 6. 创建/更新用户社交统计数据（使用正确的字段）
-- ============================================================
INSERT OR REPLACE INTO user_social_stats (user_id, likes_received, favorites_received, comments_received, followers_count, following_count, login_streak, last_login_date, updated_at) VALUES
(1001, 15, 8, 12, 8, 5, 3, datetime('now'), datetime('now')),
(1002, 10, 5, 8, 6, 4, 5, datetime('now'), datetime('now')),
(1003, 8, 4, 5, 4, 3, 2, datetime('now'), datetime('now')),
(1004, 9, 6, 6, 5, 4, 4, datetime('now'), datetime('now')),
(1005, 6, 3, 4, 3, 2, 1, datetime('now'), datetime('now')),
(1006, 5, 3, 3, 3, 3, 3, datetime('now'), datetime('now')),
(1007, 20, 12, 15, 10, 6, 7, datetime('now'), datetime('now')),
(1008, 12, 8, 10, 7, 5, 4, datetime('now'), datetime('now')),
(666, 50, 30, 25, 30, 20, 10, datetime('now'), datetime('now')),
(777, 100, 60, 50, 100, 10, 15, datetime('now'), datetime('now'));

-- ============================================================
-- 7. 创建球队主页数据（使用正确的字段）
-- ============================================================
INSERT OR REPLACE INTO team_homes (team_id, hero, about, honors, contact, created_at, updated_at) VALUES
(2, 
 '{"cover":"https://images.unsplash.com/photo-1574629810360-7efbbe195018?w=1200","title":"U12一队","subtitle":"团结拼搏 · 勇攀高峰"}',
 '{"content":"U12一队成立于2020年，是俱乐部重点培养梯队。球队现有球员20名，配备主教练1名、助理教练1名。球队技术风格以地面配合为主，注重个人技术能力和团队配合意识培养。","coach":"王教练","founded":"2020","players_count":20,"training_time":"每周二、四、六 16:00-18:00"}',
 '[{"title":"2025上海市U12联赛冠军","year":"2025","icon":"Trophy"},{"title":"2024绿地杯邀请赛冠军","year":"2024","icon":"Award"},{"title":"2024浦东新区U12联赛亚军","year":"2024","icon":"Medal"}]',
 '{"phone":"021-58888888","email":"u12@shanghailvdi.com"}',
 datetime('now'), datetime('now')
),
(3,
 '{"cover":"https://images.unsplash.com/photo-1551958219-acbc608c6377?w=1200","title":"U12二队","subtitle":"快乐足球 · 健康成长"}',
 '{"content":"U12二队成立于2021年，是俱乐部基础培养梯队。球队现有球员18名，注重基础技术训练和足球兴趣培养。","coach":"王教练","founded":"2021","players_count":18,"training_time":"每周三、五、日 16:00-18:00"}',
 '[{"title":"2025绿地杯U12组季军","year":"2025","icon":"Medal"},{"title":"2024秋季联赛优胜奖","year":"2024","icon":"Award"}]',
 '{"phone":"021-58888888","email":"u12b@shanghailvdi.com"}',
 datetime('now'), datetime('now')
),
(4,
 '{"cover":"https://images.unsplash.com/photo-1517466787929-bc90951d0974?w=1200","title":"U14精英队","subtitle":"精英培养 · 职业之路"}',
 '{"content":"U14精英队是俱乐部最高级别梯队，选拔优秀球员进行专业化培养。球队现有球员16名，目标是培养职业球员和国青队人才。","coach":"王教练","founded":"2019","players_count":16,"training_time":"每周一至周六 16:00-18:30"}',
 '[{"title":"2025全国U14联赛亚军","year":"2025","icon":"Trophy"},{"title":"2024上海市U14联赛冠军","year":"2024","icon":"Trophy"},{"title":"2024全国青少年足协杯四强","year":"2024","icon":"Award"}]',
 '{"phone":"021-58888888","email":"u14@shanghailvdi.com"}',
 datetime('now'), datetime('now')
);

-- ============================================================
-- 8. 创建训练记录（使用正确的字段）
-- ============================================================
INSERT OR REPLACE INTO training_notes (id, coach_id, player_id, title, content, category, tags, rating, is_public, view_count, created_at, updated_at) VALUES
(1, 1, 1001, '射门训练表现', '今天训练状态很好，射门准确率有提升，需要加强左脚训练。', '技术训练', '["射门", "左脚"]', 85, 0, 0, datetime('now', '-2 days'), datetime('now')),
(2, 1, 1002, '中场组织训练', '中场组织能力突出，传球视野开阔，防守意识需要加强。', '战术训练', '["传球", "视野"]', 88, 0, 0, datetime('now', '-2 days'), datetime('now')),
(3, 1, 1003, '防守位置训练', '防守位置感不错，出球需要更果断，身体对抗需要加强。', '防守训练', '["位置感", "对抗"]', 82, 0, 0, datetime('now', '-3 days'), datetime('now')),
(4, 1, 1007, '前锋射门训练', '射门技术出色，门前把握机会能力强，需要提高回防意识。', '进攻训练', '["射门", "把握机会"]', 90, 0, 0, datetime('now', '-1 days'), datetime('now'));

SELECT '测试数据补充完成V2！' as result;
