-- ============================================================
-- 少年球探 - 测试账号数据 Part 3 - 主页虚拟数据
-- 生成时间: 2026-04-08
-- ============================================================

-- 清理旧数据
DELETE FROM club_homes;
DELETE FROM team_homes;
DELETE FROM user_social_stats;

-- 1. 俱乐部主页数据 (10个俱乐部)
INSERT INTO club_homes (club_id, hero, about, contact, created_at, updated_at) VALUES
(10,
'{"title":"上海绿地青训俱乐部","subtitle":"专注青少年足球培训 · 成就未来足球之星","coverImage":"https://images.unsplash.com/photo-1574629810360-7efbbe195018?w=1200","stats":[{"label":"在训球员","value":"200+"},{"label":"专业教练","value":"15"},{"label":"成立年份","value":"2010"},{"label":"获得荣誉","value":"30+"}]}',
'{"title":"关于我们","content":"上海绿地青训俱乐部成立于2010年，是上海市领先的青少年足球培训机构。我们拥有专业的教练团队和完善的训练设施，致力于培养优秀的足球人才。俱乐部已获得上海市足协认证，是市级青训示范基地。","features":[{"icon":"Trophy","title":"专业认证","desc":"上海市足协认证青训机构"},{"icon":"Users","title":"精英教练","desc":"亚足联A级教练领衔"},{"icon":"Target","title":"科学训练","desc":"德国青训体系引进"},{"icon":"Home","title":"完善设施","desc":"标准11人制天然草球场"}]}',
'{"address":"上海市浦东新区世纪大道1000号","phone":"021-58888888","email":"contact@shanghailvdi.com"}',
datetime('now'), datetime('now')),

(11,
'{"title":"北京国安青训","subtitle":"传承国安精神 · 培育足球英才","coverImage":"https://images.unsplash.com/photo-1431324155629-1a6deb1dec8d?w=1200","stats":[{"label":"在训球员","value":"180+"},{"label":"专业教练","value":"12"},{"label":"成立年份","value":"2002"},{"label":"获得荣誉","value":"25+"}]}',
'{"title":"关于我们","content":"北京国安足球俱乐部青训基地，专注青训20年。我们秉承国安精神，注重球员技术培养和品德教育。","features":[{"icon":"Trophy","title":"职业背景","desc":"北京国安俱乐部直属青训"},{"icon":"Users","title":"专业教练","desc":"B级以上教练团队"},{"icon":"Target","title":"技术优先","desc":"脚下技术为核心"},{"icon":"Home","title":"专业场地","desc":"俱乐部专业训练基地"}]}',
'{"address":"北京市朝阳区工体北路","phone":"010-65588888","email":"guoanacademy@126.com"}',
datetime('now'), datetime('now')),

(12,
'{"title":"广州恒大足校","subtitle":"冠军之路 · 从这里起步","coverImage":"https://images.unsplash.com/photo-1553778263-73a83bab9b0c?w=1200","stats":[{"label":"在训球员","value":"300+"},{"label":"专业教练","value":"25"},{"label":"成立年份","value":"2012"},{"label":"获得荣誉","value":"50+"}]}',
'{"title":"关于我们","content":"广州恒大足球学校是全国知名青训机构，采用西班牙青训体系，培养出多名国青队球员。","features":[{"icon":"Trophy","title":"冠军培养","desc":"多名国青队球员输送"},{"icon":"Users","title":"外教团队","desc":"西班牙外教亲自指导"},{"icon":"Target","title":"国际视野","desc":"定期海外交流学习"},{"icon":"Home","title":"顶级设施","desc":"欧式训练设施"}]}',
'{"address":"广东省清远市恒大足校","phone":"0763-6888888","email":"hengdaacademy@evergrande.com"}',
datetime('now'), datetime('now')),

(13,
'{"title":"山东泰山青训","subtitle":"泰山品质 · 足球梦想","coverImage":"https://images.unsplash.com/photo-1508098682722-e99c43a406d2?w=1200","stats":[{"label":"在训球员","value":"150+"},{"label":"专业教练","value":"10"},{"label":"成立年份","value":"1993"},{"label":"获得荣誉","value":"20+"}]}',
'{"title":"关于我们","content":"山东泰山足球俱乐部青训体系，拥有近30年青训经验，是山东省最专业的青训机构之一。","features":[{"icon":"Trophy","title":"省级示范","desc":"山东省青训示范机构"},{"icon":"Users","title":"专业团队","desc":"经验丰富的教练团队"},{"icon":"Target","title":"体能优势","desc":"科学体能训练体系"},{"icon":"Home","title":"标准场地","desc":"天然草和人工草场地"}]}',
'{"address":"山东省济南市奥体中心","phone":"0531-88888888","email":"taishanacademy@126.com"}',
datetime('now'), datetime('now')),

(14,
'{"title":"江苏苏宁青训","subtitle":"苏宁青训 · 未来之星","coverImage":"https://images.unsplash.com/photo-1518604666860-9ed391f76460?w=1200","stats":[{"label":"在训球员","value":"120+"},{"label":"专业教练","value":"8"},{"label":"成立年份","value":"1994"},{"label":"获得荣誉","value":"15+"}]}',
'{"title":"关于我们","content":"江苏苏宁足球俱乐部青训基地，致力于培养技术出众、品德兼优的足球人才。","features":[{"icon":"Trophy","title":"职业通道","desc":"对接职业俱乐部"},{"icon":"Users","title":"教练团队","desc":"职业球员转型教练"},{"icon":"Target","title":"技术细腻","desc":"技术细节重点培养"},{"icon":"Home","title":"完善设施","desc":"室内训练馆配置"}]}',
'{"address":"江苏省南京市奥体中心","phone":"025-88888888","email":"suningacademy@suning.com"}',
datetime('now'), datetime('now')),

(15,
'{"title":"成都蓉城青训","subtitle":"天府足球 · 快乐成长","coverImage":"https://images.unsplash.com/photo-1567880905821-9bb6d56a3f16?w=1200","stats":[{"label":"在训球员","value":"100+"},{"label":"专业教练","value":"7"},{"label":"成立年份","value":"2018"},{"label":"获得荣誉","value":"10+"}]}',
'{"title":"关于我们","content":"成都蓉城足球俱乐部青训，采用欧洲先进青训理念，让孩子在快乐中成长，在足球中进步。","features":[{"icon":"Trophy","title":"新锐俱乐部","desc":"中甲新军青训体系"},{"icon":"Users","title":"外教指导","desc":"外籍教练技术支持"},{"icon":"Target","title":"快乐足球","desc":"兴趣优先培养"},{"icon":"Home","title":"温暖气候","desc":"全年适宜训练"}]}',
'{"address":"四川省成都市凤凰山体育公园","phone":"028-88888888","email":"rongchengacademy@126.com"}',
datetime('now'), datetime('now')),

(16,
'{"title":"武汉三镇青训","subtitle":"三镇力量 · 足球希望","coverImage":"https://images.unsplash.com/photo-1568601127030-20a5cc52e478?w=1200","stats":[{"label":"在训球员","value":"90+"},{"label":"专业教练","value":"6"},{"label":"成立年份","value":"2013"},{"label":"获得荣誉","value":"8+"}]}',
'{"title":"关于我们","content":"武汉三镇足球俱乐部青训，致力于发掘和培养华中地区优秀足球人才。","features":[{"icon":"Trophy","title":"黑马俱乐部","desc":"中乙冠军升入中甲"},{"icon":"Users","title":"专业教练","desc":"职业级教练团队"},{"icon":"Target","title":"战术素养","desc":"战术理解为重点"},{"icon":"Home","title":"训练基地","desc":"专业足球训练基地"}]}',
'{"address":"湖北省武汉市塔子湖体育中心","phone":"027-88888888","email":"sanzhenacademy@126.com"}',
datetime('now'), datetime('now')),

(17,
'{"title":"浙江绿城足校","subtitle":"绿城青训 · 技术为本","coverImage":"https://images.unsplash.com/photo-1571731956672-f2b94d7dd0cb?w=1200","stats":[{"label":"在训球员","value":"130+"},{"label":"专业教练","value":"9"},{"label":"成立年份","value":"2004"},{"label":"获得荣誉","value":"18+"}]}',
'{"title":"关于我们","content":"浙江绿城足球学校是浙江省规模最大的青训机构之一，采用日本青训体系，注重技术培养。","features":[{"icon":"Trophy","title":"日本体系","desc":"日本教练技术指导"},{"icon":"Users","title":"精细训练","desc":"小班制精品训练"},{"icon":"Target","title":"技术为核","desc":"脚下技术为核心"},{"icon":"Home","title":"绿化环境","desc":"园林式训练环境"}]}',
'{"address":"浙江省杭州市绿城足球训练基地","phone":"0571-88888888","email":"greentownacademy@126.com"}',
datetime('now'), datetime('now')),

(18,
'{"title":"河南嵩山青训","subtitle":"嵩山精神 · 足球梦想","coverImage":"https://images.unsplash.com/photo-1540747913346-19e32dc3e97e?w=1200","stats":[{"label":"在训球员","value":"80+"},{"label":"专业教练","value":"5"},{"label":"成立年份","value":"1994"},{"label":"获得荣誉","value":"6+"}]}',
'{"title":"关于我们","content":"河南嵩山龙门足球俱乐部青训，致力于中原地区足球人才培养。","features":[{"icon":"Trophy","title":"中原代表","desc":"河南省顶级青训"},{"icon":"Users","title":"教练团队","desc":"本土+外聘教练"},{"icon":"Target","title":"体能技术","desc":"体能与技术并重"},{"icon":"Home","title":"专业场地","desc":"标准足球训练场"}]}',
'{"address":"河南省郑州市航海体育场","phone":"0371-88888888","email":"songshanacademy@126.com"}',
datetime('now'), datetime('now')),

(19,
'{"title":"天津津门虎青训","subtitle":"津门虎威 · 足球传承","coverImage":"https://images.unsplash.com/photo-1508009603885-50cf7c579365?w=1200","stats":[{"label":"在训球员","value":"85+"},{"label":"专业教练","value":"6"},{"label":"成立年份","value":"1994"},{"label":"获得荣誉","value":"7+"}]}',
'{"title":"关于我们","content":"天津津门虎足球俱乐部青训，传承天津足球文化，培养新一代足球人才。","features":[{"icon":"Trophy","title":"传统强队","desc":"天津足球代表"},{"icon":"Users","title":"教练团队","desc":"老中青结合团队"},{"icon":"Target","title":"青训传统","desc":"30年青训积累"},{"icon":"Home","title":"专业设施","desc":"专业训练基地"}]}',
'{"address":"天津市水滴体育场","phone":"022-88888888","email":"jinmenhuacademy@126.com"}',
datetime('now'), datetime('now'));

-- 2. 球队主页数据 (9支球队)
INSERT INTO team_homes (team_id, hero, about, honors, contact, created_at, updated_at) VALUES
(1,
'{"cover":"https://images.unsplash.com/photo-1574629810360-7efbbe195018?w=1200","title":"U12一队","subtitle":"团结拼搏 · 勇攀高峰"}',
'{"content":"U12一队成立于2020年，是俱乐部重点培养梯队。球队现有球员20名，配备主教练1名、助理教练1名。球队技术风格以地面配合为主，注重个人技术能力和团队配合意识培养。","coach":"王教练","founded":"2020","players_count":20,"training_time":"每周二、四、六 16:00-18:00"}',
'[{"title":"2025上海市U12联赛冠军","year":"2025","icon":"Trophy"},{"title":"2024绿地杯邀请赛冠军","year":"2024","icon":"Award"},{"title":"2024浦东新区U12联赛亚军","year":"2024","icon":"Medal"}]',
'{"phone":"021-58888888","email":"u12a@shanghailvdi.com"}',
datetime('now'), datetime('now')),

(2,
'{"cover":"https://images.unsplash.com/photo-1551958219-acbc608c6377?w=1200","title":"U12二队","subtitle":"快乐足球 · 健康成长"}',
'{"content":"U12二队成立于2021年，是俱乐部基础培养梯队。球队现有球员18名，注重基础技术训练和足球兴趣培养。","coach":"王教练","founded":"2021","players_count":18,"training_time":"每周三、五、日 16:00-18:00"}',
'[{"title":"2025绿地杯U12组季军","year":"2025","icon":"Medal"},{"title":"2024秋季联赛优胜奖","year":"2024","icon":"Award"}]',
'{"phone":"021-58888888","email":"u12b@shanghailvdi.com"}',
datetime('now'), datetime('now')),

(3,
'{"cover":"https://images.unsplash.com/photo-1431324155629-1a6deb1dec8d?w=1200","title":"U12队","subtitle":"国安精神 · 传承创新"}',
'{"content":"北京国安U12队，传承国安精神，注重技术培养和战术素养。球队现有球员18名。","coach":"李教练","founded":"2002","players_count":18,"training_time":"每周二、四、六 15:30-17:30"}',
'[{"title":"2025北京市U12联赛亚军","year":"2025","icon":"Medal"},{"title":"2024京津冀邀请赛冠军","year":"2024","icon":"Trophy"}]',
'{"phone":"010-65588888","email":"guoanutc@126.com"}',
datetime('now'), datetime('now')),

(4,
'{"cover":"https://images.unsplash.com/photo-1553778263-73a83bab9b0c?w=1200","title":"U12队","subtitle":"恒大标准 · 冠军之路"}',
'{"content":"广州恒大U12队，采用西班牙青训体系，培养技术型球员。球队现有球员22名。","coach":"李教练","founded":"2012","players_count":22,"training_time":"每周一至周六 16:00-18:00"}',
'[{"title":"2025全国U12联赛冠军","year":"2025","icon":"Trophy"},{"title":"2025广东省U12联赛冠军","year":"2025","icon":"Trophy"},{"title":"2024恒大杯邀请赛冠军","year":"2024","icon":"Award"}]',
'{"phone":"0763-6888888","email":"hengdautc@evergrande.com"}',
datetime('now'), datetime('now')),

(5,
'{"cover":"https://images.unsplash.com/photo-1508098682722-e99c43a406d2?w=1200","title":"U12队","subtitle":"泰山品质 · 体能为先"}',
'{"content":"山东泰山U12队，注重体能训练和战术执行。球队现有球员16名。","coach":"张教练","founded":"2005","players_count":16,"training_time":"每周二、四、六 15:00-17:00"}',
'[{"title":"2025山东省U12联赛冠军","year":"2025","icon":"Trophy"},{"title":"2024华东区邀请赛亚军","year":"2024","icon":"Medal"}]',
'{"phone":"0531-88888888","email":"taishanutc@126.com"}',
datetime('now'), datetime('now')),

(6,
'{"cover":"https://images.unsplash.com/photo-1518604666860-9ed391f76460?w=1200","title":"U12队","subtitle":"苏宁青训 · 技术细腻"}',
'{"content":"江苏苏宁U12队，强调技术细节和比赛阅读能力。球队现有球员15名。","coach":"张教练","founded":"2008","players_count":15,"training_time":"每周三、五、日 16:00-18:00"}',
'[{"title":"2025江苏省U12联赛亚军","year":"2025","icon":"Medal"},{"title":"2024长三角邀请赛冠军","year":"2024","icon":"Trophy"}]',
'{"phone":"025-88888888","email":"suningutc@suning.com"}',
datetime('now'), datetime('now')),

(7,
'{"cover":"https://images.unsplash.com/photo-1567880905821-9bb6d56a3f16?w=1200","title":"U12队","subtitle":"蓉城新星 · 快乐足球"}',
'{"content":"成都蓉城U12队，采用欧洲先进理念，让孩子在快乐中成长。球队现有球员14名。","coach":"刘教练","founded":"2018","players_count":14,"training_time":"每周二、四、六 16:00-18:00"}',
'[{"title":"2025西南区U12邀请赛冠军","year":"2025","icon":"Trophy"},{"title":"2024四川省U12联赛亚军","year":"2024","icon":"Medal"}]',
'{"phone":"028-88888888","email":"rongchengutc@126.com"}',
datetime('now'), datetime('now')),

(8,
'{"cover":"https://images.unsplash.com/photo-1568601127030-20a5cc52e478?w=1200","title":"U12队","subtitle":"三镇力量 · 青春风暴"}',
'{"content":"武汉三镇U12队，注重新生代球员培养。球队现有球员12名。","coach":"刘教练","founded":"2013","players_count":12,"training_time":"每周三、五、日 16:00-18:00"}',
'[{"title":"2025华中区U12邀请赛亚军","year":"2025","icon":"Medal"},{"title":"2024湖北省U12联赛季军","year":"2024","icon":"Medal"}]',
'{"phone":"027-88888888","email":"sanzhenutc@126.com"}',
datetime('now'), datetime('now')),

(9,
'{"cover":"https://images.unsplash.com/photo-1571731956672-f2b94d7dd0cb?w=1200","title":"U12队","subtitle":"绿城技术 · 日本风格"}',
'{"content":"浙江绿城U12队，采用日本青训体系，注重脚下技术和团队配合。球队现有球员15名。","coach":"刘教练","founded":"2004","players_count":15,"training_time":"每周二、四、六 15:30-17:30"}',
'[{"title":"2025浙江省U12联赛冠军","year":"2025","icon":"Trophy"},{"title":"2024长三角U12邀请赛冠军","year":"2024","icon":"Trophy"}]',
'{"phone":"0571-88888888","email":"greentownutc@126.com"}',
datetime('now'), datetime('now'));

-- 3. 用户社交统计数据
INSERT INTO user_social_stats (user_id, likes_received, favorites_received, comments_received, followers_count, following_count, login_streak, last_login_date, updated_at) VALUES
-- 球员社交数据
(2001, 45, 12, 28, 56, 23, 7, datetime('now'), datetime('now')),
(2002, 38, 8, 22, 45, 19, 5, datetime('now'), datetime('now')),
(2003, 25, 5, 15, 32, 15, 3, datetime('now'), datetime('now')),
(2004, 18, 3, 10, 28, 12, 4, datetime('now'), datetime('now')),
(2005, 30, 7, 18, 40, 16, 6, datetime('now'), datetime('now')),
(2006, 22, 4, 12, 35, 14, 2, datetime('now'), datetime('now')),
(2007, 15, 2, 8, 25, 10, 1, datetime('now'), datetime('now')),
(2008, 12, 2, 6, 22, 8, 3, datetime('now'), datetime('now')),
(2009, 28, 6, 16, 38, 17, 5, datetime('now'), datetime('now')),
(2010, 10, 1, 5, 20, 7, 2, datetime('now'), datetime('now')),
(2011, 20, 4, 11, 30, 13, 4, datetime('now'), datetime('now')),
(2012, 24, 5, 14, 33, 15, 3, datetime('now'), datetime('now')),
(2013, 14, 2, 7, 24, 9, 2, datetime('now'), datetime('now')),
(2014, 16, 3, 9, 26, 11, 3, datetime('now'), datetime('now')),
(2015, 19, 4, 10, 29, 12, 4, datetime('now'), datetime('now')),
(2016, 11, 1, 5, 21, 8, 1, datetime('now'), datetime('now')),
(2017, 13, 2, 6, 23, 9, 2, datetime('now'), datetime('now')),
(2018, 9, 1, 4, 18, 6, 1, datetime('now'), datetime('now')),
(2019, 21, 4, 12, 31, 14, 4, datetime('now'), datetime('now')),
(2020, 17, 3, 9, 27, 11, 3, datetime('now'), datetime('now')),
-- 教练社交数据
(20, 120, 35, 65, 156, 42, 15, datetime('now'), datetime('now')),
(21, 85, 25, 45, 120, 35, 10, datetime('now'), datetime('now')),
(22, 70, 20, 38, 95, 28, 8, datetime('now'), datetime('now')),
(23, 55, 15, 30, 78, 22, 6, datetime('now'), datetime('now')),
-- 俱乐部社交数据
(10, 250, 80, 120, 320, 45, 20, datetime('now'), datetime('now')),
(11, 180, 55, 90, 245, 38, 15, datetime('now'), datetime('now')),
(12, 220, 70, 105, 280, 42, 18, datetime('now'), datetime('now'));

SELECT '主页和社交数据插入完成!' as result;
