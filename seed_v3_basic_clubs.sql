-- ============================================
-- 少年球探 - 俱乐部数据 v3.1 (修正表结构)
-- ============================================

INSERT INTO clubs (id, user_id, name, logo, description, address, contact_name, contact_phone, established_year, club_size, member_level, created_at, updated_at) VALUES
(1, 10, '上海绿地青训俱乐部', 'https://ui-avatars.com/api/?name=上海绿地&background=2ECC71&color=fff&size=200', '上海绿地青训俱乐部成立于2015年，专注于青少年足球人才培养，拥有专业教练团队和完善的训练体系。', '上海市浦东新区足球训练基地', '王管理员', '021-88888888', 2015, '50-100人', 'pro', datetime('now'), datetime('now'));

.print '俱乐部数据导入完成: ' || (SELECT COUNT(*) FROM clubs) || ' 条'
