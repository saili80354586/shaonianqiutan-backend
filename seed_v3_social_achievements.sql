-- ============================================
-- 少年球探 - 社交成就数据 v3.1 (修正表结构)
-- ============================================

-- 先插入成就定义
INSERT INTO social_achievements (id, name, description, icon, category, condition, threshold) VALUES
('social_like_50', '社交达人', '获得50次以上点赞', 'heart', 'social', 'likes_received >= 50', 50),
('social_comment_20', '评论活跃', '发表20条以上评论', 'message-circle', 'social', 'comments_count >= 20', 20),
('social_follower_100', '粉丝过百', '粉丝数超过100', 'users', 'social', 'followers_count >= 100', 100),
('social_like_30', '人气选手', '获得30次以上点赞', 'heart', 'social', 'likes_received >= 30', 30),
('social_new', '新星起步', '首次获得点赞', 'star', 'social', 'likes_received >= 1', 1),
('training_attend', '出勤之星', '月度训练出勤率100%', 'calendar', 'training', 'attendance_rate >= 100', 100),
('match_goal_10', '进球机器', '在正式比赛中攻入10粒进球', 'target', 'match', 'goals >= 10', 10),
('match_clean_sheet', '防守铁闸', '单赛季完成20次抢断', 'shield', 'match', 'tackles >= 20', 20),
('coaching_50', '桃李满天下', '培养球员超过50人', 'book-open', 'coaching', 'players_count >= 50', 50),
('coaching_champion', '冠军教练', '带领球队获得冠军', 'trophy', 'coaching', 'championships >= 1', 1),
('scout_discover_10', '伯乐识马', '发掘优秀球员10人', 'eye', 'scouting', 'discovered >= 10', 10),
('scout_report_20', '报告专家', '完成球探报告20份', 'file-text', 'scouting', 'reports >= 20', 20);

-- 用户成就
INSERT INTO user_social_achievements (user_id, achievement_id, achieved_at) VALUES
-- 王小明的成就
(2001, 'social_like_50', datetime('now', '-10 days')),
(2001, 'social_comment_20', datetime('now', '-15 days')),
(2001, 'social_follower_100', datetime('now', '-20 days')),
(2001, 'match_goal_10', datetime('now', '-30 days')),

-- 李小强的成就
(2002, 'social_like_30', datetime('now', '-8 days')),
(2002, 'training_attend', datetime('now', '-18 days')),

-- 张小刚的成就
(2003, 'social_new', datetime('now', '-5 days')),
(2003, 'match_clean_sheet', datetime('now', '-15 days')),

-- 教练成就
(20, 'coaching_50', datetime('now', '-30 days')),
(20, 'coaching_champion', datetime('now', '-45 days')),

-- 球探成就
(24, 'scout_discover_10', datetime('now', '-20 days')),
(24, 'scout_report_20', datetime('now', '-35 days'));

.print '社交成就数据导入完成: ' || (SELECT COUNT(*) FROM social_achievements) || ' 个成就, ' || (SELECT COUNT(*) FROM user_social_achievements) || ' 条用户成就'
