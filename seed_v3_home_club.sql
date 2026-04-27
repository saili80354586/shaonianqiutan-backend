-- ============================================
-- 少年球探 - 俱乐部主页数据 v3.1 (修正表结构)
-- ============================================

INSERT INTO club_homes (id, club_id, hero, about, contact, created_at, updated_at) VALUES
(1, 1, 'https://images.unsplash.com/photo-1574629810360-7efbbe195018?w=1200', '上海绿地青训俱乐部成立于2015年，专注于青少年足球人才培养。俱乐部拥有专业教练团队和完善的训练体系，致力于发现和培养足球苗子。俱乐部优势：专业教练团队、科学训练体系、完善的比赛机会、升学通道对接。', '电话: 021-88888888 | 微信: shldqclub | 地址: 上海市浦东新区足球训练基地', datetime('now'), datetime('now'));

.print '俱乐部主页数据导入完成: ' || (SELECT COUNT(*) FROM club_homes) || ' 条'
