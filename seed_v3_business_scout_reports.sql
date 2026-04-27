-- ============================================
-- 少年球探 - 球探报告数据 v3.1 (修正表结构)
-- ============================================

-- 赵球探的报告 (scout_id=1)
INSERT INTO scout_reports (id, scout_id, player_id, overall_rating, potential_rating, status, strengths, weaknesses, summary, recommendation, views, likes, published_at, created_at, updated_at) VALUES
(1, 1, 1, 92, '{"technical": 88, "physical": 90, "mental": 82, "potential": 92}', 'published', '身体素质出色，速度优势明显，进攻意识强烈', '战术理解能力有待提高', '身体素质出色，速度优势明显，进攻意识强烈，具有较高的发展潜力。建议加强战术训练。', '值得重点关注，建议持续跟踪', 45, 12, datetime('now', '-14 days'), datetime('now', '-15 days'), datetime('now', '-14 days')),
(2, 1, 2, 85, '{"passing": 88, "vision": 85, "control": 82, "leadership": 78}', 'published', '中场调度能力出色，传球视野开阔', '对抗能力需加强', '中场调度能力出色，传球视野开阔，具备一定的领导能力。建议加强身体对抗训练。', '具备中场指挥官潜质', 38, 10, datetime('now', '-11 days'), datetime('now', '-12 days'), datetime('now', '-11 days')),
(3, 1, 3, 82, '{"tackling": 82, "heading": 88, "positioning": 78, "physical": 85}', 'published', '身体强壮，头球能力突出', '位置感有待提高', '身体强壮，头球能力突出，位置感有待提高，但潜力不错。建议加强位置训练。', '潜力不错，值得关注', 30, 8, datetime('now', '-7 days'), datetime('now', '-8 days'), datetime('now', '-7 days'));

-- 陈球探的报告 (scout_id=2)
INSERT INTO scout_reports (id, scout_id, player_id, overall_rating, potential_rating, status, strengths, weaknesses, summary, recommendation, views, likes, published_at, created_at, updated_at) VALUES
(4, 2, 1, 90, '{"technical": 90, "pace": 92, "finishing": 88}', 'published', '速度出色，射门技术好', '需要加强团队配合', '经过多次观察，确认该球员具备优秀前锋潜质，建议持续关注。', '建议持续跟踪', 25, 6, datetime('now', '-4 days'), datetime('now', '-5 days'), datetime('now', '-4 days'));

.print '球探报告导入完成: ' || (SELECT COUNT(*) FROM scout_reports) || ' 条'
