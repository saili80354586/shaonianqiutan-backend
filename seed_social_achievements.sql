-- 社交成就种子数据
-- 插入到 social_achievements 表

INSERT OR IGNORE INTO social_achievements (id, name, description, icon, category, condition, threshold) VALUES
-- 贡献类成就
('first_report', '初出茅庐', '发布第一份报告', 'trophy', 'contribution', '发布1份报告', 1),
('report_10', '小有名气', '发布10份报告', 'trophy', 'contribution', '发布10份报告', 10),
('report_100', '报告大师', '发布100份报告', 'trophy', 'contribution', '发布100份报告', 100),

-- 活跃类成就
('first_like', '赞不绝口', '收到第一个点赞', 'heart', 'engagement', '收到1个点赞', 1),
('likes_100', '点赞达人', '收到100个赞', 'heart', 'engagement', '收到100个赞', 100),
('first_comment', '议论纷纷', '收到第一条评论', 'message', 'engagement', '收到1条评论', 1),
('streak_7', '连续活跃', '连续登录7天', 'zap', 'engagement', '连续登录7天', 7),
('streak_30', '习惯养成', '连续登录30天', 'zap', 'engagement', '连续登录30天', 30),

-- 社交类成就
('first_follower', '收获粉丝', '获得第一个粉丝', 'users', 'social', '粉丝数=1', 1),
('followers_100', '小有名气', '获得100个粉丝', 'users', 'social', '粉丝数=100', 100),

-- 里程碑成就
('first_scout', '慧眼识珠', '球员被球探发掘', 'star', 'milestone', '球探报告提及', 1),
('player_joined', '星探伯乐', '发掘的球员加入俱乐部', 'award', 'milestone', '球员被俱乐部选中', 1);
