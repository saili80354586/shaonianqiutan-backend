-- ============================================
-- 少年球探 - 比赛总结数据 v3.1 (修正表结构)
-- ============================================

-- 比赛1: 已完成 (教练点评已完成)
INSERT INTO match_summaries (id, team_id, match_name, match_date, opponent, our_score, opponent_score, match_result, player_summary, coach_summary, coach_id, status, created_at, updated_at) VALUES
(1, 1, '浦东联赛第三轮', date('now', '-10 days'), '浦东联队', 3, 1, 'win', '{"2001": {"self_review": "进攻配合流畅", "rating": 5}, "2002": {"self_review": "传球到位", "rating": 4}, "2003": {"self_review": "防守稳健", "rating": 4}}', '{"summary": "球员们表现出色，进攻欲望强烈，建议继续保持", "highlights": ["王小明梅开二度", "整体配合默契"]}', 20, 'completed', datetime('now', '-9 days'), datetime('now', '-8 days'));

-- 比赛2: 球员自评中
INSERT INTO match_summaries (id, team_id, match_name, match_date, opponent, our_score, opponent_score, match_result, player_summary, coach_summary, coach_id, status, created_at, updated_at) VALUES
(2, 1, '浦东联赛第四轮', date('now', '-3 days'), '虹口青训队', 2, 2, 'draw', '{"2001": {"self_review": "下半场有些松懈", "rating": 4}, "2002": {"self_review": "防守需要加强", "rating": 3}}', NULL, 20, 'in_progress', datetime('now', '-2 days'), NULL);

-- 比赛3: 待开始
INSERT INTO match_summaries (id, team_id, match_name, match_date, opponent, coach_id, status, created_at) VALUES
(3, 1, '浦东联赛第五轮', date('now', '+7 days'), '杨浦少年队', 20, 'pending', datetime('now'));

-- U12二队比赛
INSERT INTO match_summaries (id, team_id, match_name, match_date, opponent, our_score, opponent_score, match_result, player_summary, coach_summary, coach_id, status, created_at, updated_at) VALUES
(4, 2, '嘉定交流赛', date('now', '-5 days'), '嘉定U12队', 4, 2, 'win', '{"2005": {"self_review": "第一次首发进球", "rating": 5}, "2006": {"self_review": "中场控制不错", "rating": 4}, "2007": {"self_review": "防守到位", "rating": 4}}', '{"summary": "整体表现不错，进攻端有亮点", "highlights": ["陈小龙首球破门", "团队配合流畅"]}', 21, 'completed', datetime('now', '-4 days'), datetime('now', '-3 days'));

.print '比赛总结数据导入完成: ' || (SELECT COUNT(*) FROM match_summaries) || ' 条'
