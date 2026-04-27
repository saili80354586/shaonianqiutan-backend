-- ============================================
-- 少年球探 - 完善球员注册信息（家长/协会数据）
-- 版本: v1.4
-- 日期: 2026-04-11
-- 说明: 补充8个球员的家长信息、足协会员信息、第二位置
-- ============================================

-- 球员2001: 王小明
UPDATE users SET
    father_phone = '13800011001',
    mother_phone = '13800011002',
    father_height = 175,
    mother_height = 163,
    father_job = '工程师',
    mother_job = '教师',
    father_edu = '本科',
    mother_edu = '硕士',
    father_athlete = 0,
    mother_athlete = 0,
    association = '',
    fa_registered = 0,
    second_position = '影子前锋'
WHERE id = 2001;

-- 球员2002: 李小强
UPDATE users SET
    father_phone = '13800011003',
    mother_phone = '13800011004',
    father_height = 178,
    mother_height = 165,
    father_job = '医生',
    mother_job = '护士',
    father_edu = '硕士',
    mother_edu = '本科',
    father_athlete = 0,
    mother_athlete = 0,
    association = '',
    fa_registered = 0,
    second_position = '前腰'
WHERE id = 2002;

-- 球员2003: 张小刚
UPDATE users SET
    father_phone = '13800011005',
    mother_phone = '13800011006',
    father_height = 180,
    mother_height = 162,
    father_job = '企业家',
    mother_job = '会计',
    father_edu = '本科',
    mother_edu = '本科',
    father_athlete = 1,
    mother_athlete = 0,
    association = '上海市足球协会',
    fa_registered = 1,
    second_position = '边后卫'
WHERE id = 2003;

-- 球员2004: 刘小军
UPDATE users SET
    father_phone = '13800011007',
    mother_phone = '13800011008',
    father_height = 176,
    mother_height = 164,
    father_job = '公务员',
    mother_job = '银行职员',
    father_edu = '本科',
    mother_edu = '本科',
    father_athlete = 0,
    mother_athlete = 0,
    association = '',
    fa_registered = 0,
    second_position = '后卫'
WHERE id = 2004;

-- 球员2005: 陈小龙
UPDATE users SET
    father_phone = '13800011009',
    mother_phone = '13800011010',
    father_height = 179,
    mother_height = 166,
    father_job = '律师',
    mother_job = '设计师',
    father_edu = '硕士',
    mother_edu = '本科',
    father_athlete = 0,
    mother_athlete = 0,
    association = '上海市足球协会',
    fa_registered = 1,
    second_position = '边锋'
WHERE id = 2005;

-- 球员2006: 赵小虎
UPDATE users SET
    father_phone = '13800011011',
    mother_phone = '13800011012',
    father_height = 177,
    mother_height = 161,
    father_job = '销售经理',
    mother_job = '行政',
    father_edu = '本科',
    mother_edu = '大专',
    father_athlete = 0,
    mother_athlete = 0,
    association = '',
    fa_registered = 0,
    second_position = '边前卫'
WHERE id = 2006;

-- 球员2007: 孙小杰
UPDATE users SET
    father_phone = '13800011013',
    mother_phone = '13800011014',
    father_height = 182,
    mother_height = 167,
    father_job = '建筑师',
    mother_job = '医生',
    father_edu = '硕士',
    mother_edu = '硕士',
    father_athlete = 0,
    mother_athlete = 0,
    association = '',
    fa_registered = 0,
    second_position = '中后卫'
WHERE id = 2007;

-- 球员2008: 周小鹏
UPDATE users SET
    father_phone = '13800011015',
    mother_phone = '13800011016',
    father_height = 174,
    mother_height = 160,
    father_job = '厨师',
    mother_job = '服务员',
    father_edu = '高中',
    mother_edu = '高中',
    father_athlete = 0,
    mother_athlete = 0,
    association = '',
    fa_registered = 0,
    second_position = '后卫'
WHERE id = 2008;

-- 验证更新结果
SELECT '球员注册信息更新验证:' as info;
SELECT
    id,
    name,
    father_phone,
    mother_phone,
    father_height as "父高",
    mother_height as "母高",
    father_job as "父亲职业",
    mother_job as "母亲职业",
    association as "所属协会",
    fa_registered as "足协会员",
    second_position as "第二位置"
FROM users
WHERE id BETWEEN 2001 AND 2008;
