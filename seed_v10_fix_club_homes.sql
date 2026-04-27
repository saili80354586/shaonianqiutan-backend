-- ============================================
-- 少年球探 - 修复 club_homes 错误数据
-- 版本: v1.4
-- 日期: 2026-04-11
-- 说明: club_homes 表存在 club_id=14 的错误记录，但 clubs 表没有 ID=14
-- ============================================

-- 1. 查看当前 club_homes 数据
SELECT '修复前 club_homes 数据:' as info;
SELECT id, club_id, hero, created_at FROM club_homes;

-- 2. 删除 club_id 不存在于 clubs 表的记录
DELETE FROM club_homes WHERE club_id NOT IN (SELECT id FROM clubs);

-- 3. 确保只有 club_id=1 的记录（上海绿地俱乐部）
-- 如果有其他 club_id 的记录，也一并删除
DELETE FROM club_homes WHERE club_id != 1;

-- 4. 重新插入正确的 club_homes 记录（如果表被清空）
INSERT OR REPLACE INTO club_homes (id, club_id, hero, about, contact, created_at, updated_at)
VALUES (
    1,
    1,
    'https://images.unsplash.com/photo-1574629810360-7efbbe195018?w=1200',
    '上海绿地青训俱乐部成立于2015年，专注于青少年足球人才培养。俱乐部拥有专业教练团队和完善的训练体系，致力于发现和培养足球苗子。俱乐部优势：专业教练团队、科学训练体系、完善的比赛机会、升学通道对接。',
    '电话: 021-88888888 | 微信: shldqclub | 地址: 上海市浦东新区足球训练基地',
    '2026-04-10 16:56:54',
    '2026-04-10 16:56:54'
);

-- 5. 验证修复结果
SELECT '修复后 club_homes 数据:' as info;
SELECT id, club_id, hero, created_at FROM club_homes;

-- 6. 验证 clubs 表结构
SELECT 'clubs 表数据:' as info;
SELECT id, user_id, name FROM clubs;
