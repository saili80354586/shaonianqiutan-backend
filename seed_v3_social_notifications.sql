-- ============================================
-- 少年球探 - 通知数据 v3.1 (修正表结构)
-- ============================================

-- 王小明收到的通知 (7条)
INSERT INTO notifications (id, user_id, type, title, content, is_read, created_at) VALUES
(1, 2001, 'report', '报告已完成', '您购买的前锋专项能力评估报告已完成，请查看', 1, datetime('now', '-28 days')),
(2, 2001, 'report', '新报告生成', '综合能力评估报告已生成，快去看看吧', 1, datetime('now', '-13 days')),
(3, 2001, 'weekly', '周报审核通过', '教练已审核您的周报，给予5星评价', 1, datetime('now', '-6 days')),
(4, 2001, 'social', '新粉丝', '赵球探关注了您', 1, datetime('now', '-20 days')),
(5, 2001, 'social', '收到点赞', '您的动态收到10个赞', 1, datetime('now', '-11 days')),
(6, 2001, 'social', '收到评论', '李分析师评论了您的动态', 0, datetime('now', '-2 days')),
(7, 2001, 'match', '比赛通知', '下周将有一场与杨浦少年队的比赛，请做好准备', 0, datetime('now', '-1 days'));

-- 李小强收到的通知 (5条)
INSERT INTO notifications (id, user_id, type, title, content, is_read, created_at) VALUES
(8, 2002, 'report', '报告已完成', '您购买的中场技术分析报告已完成', 1, datetime('now', '-18 days')),
(9, 2002, 'weekly', '周报待审核', '请提交本周训练周报', 1, datetime('now', '-2 days')),
(10, 2002, 'social', '新粉丝', '王小明关注了您', 1, datetime('now', '-9 days')),
(11, 2002, 'social', '收到评论', '前锋王小明评论了您的动态', 0, datetime('now', '-3 days')),
(12, 2002, 'match', '比赛提醒', '周三有训练赛，请准时参加', 0, datetime('now', '-1 days'));

-- 张小刚收到的通知 (3条)
INSERT INTO notifications (id, user_id, type, title, content, is_read, created_at) VALUES
(13, 2003, 'report', '报告已完成', '您购买的后卫位置感评估报告已完成', 1, datetime('now', '-23 days')),
(14, 2003, 'weekly', '周报提醒', '本周周报还未提交，请尽快填写', 0, datetime('now', '-1 days')),
(15, 2003, 'social', '收到点赞', '您的动态收到5个赞', 0, datetime('now', '-5 days'));

-- 教练通知
INSERT INTO notifications (id, user_id, type, title, content, is_read, created_at) VALUES
(16, 20, 'weekly', '周报待审核', '有3份周报等待您审核', 1, datetime('now', '-2 days')),
(17, 20, 'match', '比赛总结', '与浦东联队的比赛总结已提交', 1, datetime('now', '-8 days')),
(18, 20, 'player', '球员关注', '您关注的王小明表现出色', 0, datetime('now', '-3 days'));

-- 分析师通知
INSERT INTO notifications (id, user_id, type, title, content, is_read, created_at) VALUES
(19, 30, 'order', '新订单', '您有一份新的报告订单，请及时处理', 1, datetime('now', '-30 days')),
(20, 30, 'order', '订单完成', '报告已完成并交付', 1, datetime('now', '-28 days')),
(21, 30, 'social', '收到评论', '王小明评论了您的报告', 0, datetime('now', '-2 days'));

.print '通知数据导入完成: ' || (SELECT COUNT(*) FROM notifications) || ' 条'
