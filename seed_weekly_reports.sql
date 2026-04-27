-- ============================================================
-- 周报功能测试数据填充脚本
-- 为所有测试账号按角色填充周报数据
-- 最后更新: 2026-04-08
-- ============================================================

-- 清除旧数据（可选，如需重新填充取消注释）
-- DELETE FROM weekly_reports WHERE id > 0;
-- DELETE FROM weekly_report_periods WHERE id > 0;
-- DELETE FROM sqlite_sequence WHERE name IN ('weekly_reports', 'weekly_report_periods');

-- ============================================================
-- 1. 创建周报周期数据
-- ============================================================
-- 上周周期 (2026-03-30 ~ 2026-04-05) - 已归档
-- 本周周期 (2026-04-06 ~ 2026-04-13) - 进行中

INSERT OR REPLACE INTO weekly_report_periods (id, team_id, week_start, week_end, deadline, total_players, submitted_count, pending_count, overdue_count, reviewed_count, status, created_at, updated_at) VALUES
-- 上周周报周期 - 已归档
(1, 2, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59', 3, 3, 0, 0, 3, 'archived', datetime('now', '-7 days'), datetime('now')),
(2, 3, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59', 2, 2, 0, 0, 2, 'archived', datetime('now', '-7 days'), datetime('now')),
(3, 4, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59', 2, 2, 0, 0, 2, 'archived', datetime('now', '-7 days'), datetime('now')),
-- 本周周报周期 - 进行中
(4, 2, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59', 3, 2, 1, 0, 1, 'active', datetime('now'), datetime('now')),
(5, 3, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59', 2, 0, 2, 0, 0, 'active', datetime('now'), datetime('now')),
(6, 4, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59', 2, 1, 1, 0, 1, 'active', datetime('now'), datetime('now'));

-- ============================================================
-- 2. 创建上周周报数据 (2026-03-30 ~ 2026-04-05) - 全部已审核归档
-- ============================================================

-- U12一队 上周周报 (球员ID: 1001王小明, 1002李小强, 1003张小刚)
INSERT OR REPLACE INTO weekly_reports (
  id, team_id, player_id, coach_id, week_start, week_end, deadline,
  training_count, training_duration, absence_count, absence_reason,
  knowledge_summary, technical_content, tactical_content, physical_condition, match_performance,
  self_attitude_rating, self_technique_rating, self_teamwork_rating, improvements_detail, weaknesses,
  fatigue_level, injuries, sleep_quality, diet_condition, message_to_coach, attachments,
  submit_status, submitted_at,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  review_status, review_coach_id, review_comment, strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award, reviewed_at,
  created_at, updated_at
) VALUES
-- 王小明 - U12一队 - 上周已审核 (approved)
(1, 2, 1001, 666, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59',
 4, 360, 0, '',
 '本周主要学习了短传配合和跑位技巧，通过分组对抗加深了对战术的理解。',
 '重点练习了脚内侧传球和接球停球，提高了传球的准确性。',
 '学习了二过一配合和边路传中战术，在对抗中尝试运用。',
 '进行了折返跑和跳绳训练，体能有所提升。',
 '在周六的队内对抗赛中表现出色，打进1球并助攻1次。',
 5, 4, 5, '传球准确性有所提高，能够更好地观察队友位置。', '射门力量还需要加强，跑位时机把握不够准确。',
 2, '无伤病', 4, '饮食正常，每天保证充足的蛋白质摄入。', '感谢教练的悉心指导，我会继续努力！', '[]',
 'submitted', datetime('now', '-5 days'),
 5, 4, 4, 5, 'approved', 666, 
 '王小明本周表现非常出色，训练态度积极主动，技术动作规范，在对抗赛中展现了良好的战术意识。',
 '传球准确性明显提高，能够主动观察队友位置；对抗赛中敢于拿球突破，表现出良好的自信心。',
 '建议加强射门力量训练，多进行远射练习；跑位时机需要更精准，可以多观看职业比赛学习。',
 '对二过一配合的理解还不够深入，建议多进行无球跑动练习。',
 '下周重点训练射门力量和跑位时机把握。',
 1, datetime('now', '-4 days'),
 datetime('now', '-7 days'), datetime('now')),

-- 李小强 - U12一队 - 上周已审核 (approved)
(2, 2, 1002, 666, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59',
 4, 360, 0, '',
 '本周学习了中场组织和控球技巧，理解了作为中场核心的责任。',
 '练习了长传转移和一脚出球，控球稳定性有所提高。',
 '学习了中场接应和转身摆脱，在狭小空间内的处理球能力有进步。',
 '进行了核心力量训练和折返跑，体能储备充足。',
 '在对抗赛中表现稳定，多次成功拦截对方传球。',
 5, 4, 4, '控球稳定性提高，能够更好地保护球权。', '视野还需要拓宽，长传准确性有待提高。',
 3, '无伤病', 4, '饮食正常，注意补充碳水化合物。', '希望能有更多的对抗训练机会。', '[]',
 'submitted', datetime('now', '-5 days'),
 5, 4, 4, 4, 'approved', 666,
 '李小强本周表现稳定，作为中场球员展现了良好的控球能力和战术执行力。',
 '控球稳定性明显提高，能够在压力下保持冷静；拦截能力出色，预判准确。',
 '建议加强长传准确性训练，提高视野范围；转身速度可以更快一些。',
 '对中场组织角色的理解还需要加深，建议多观察优秀中场球员的比赛。',
 '下周重点训练长传准确性和视野拓展。',
 0, datetime('now', '-4 days'),
 datetime('now', '-7 days'), datetime('now')),

-- 张小刚 - U12一队 - 上周已审核 (approved)
(3, 2, 1003, 666, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59',
 3, 270, 1, '周三因学校活动请假一次。',
 '本周学习了防守站位和协防配合，理解了防守体系的重要性。',
 '练习了铲球和头球解围技术，防守动作更加规范。',
 '学习了人盯人防守和区域防守的转换，防守意识有所提高。',
 '因请假少了一次训练，但其他时间训练强度保持。',
 '在对抗赛中防守稳健，成功阻止了多次对方进攻。',
 4, 4, 4, '防守站位更加合理，协防意识增强。', '铲球时机把握还需要提高，有时过于急躁。',
 3, '无伤病', 3, '饮食正常。', '希望教练能多指导防守技巧。', '[]',
 'submitted', datetime('now', '-5 days'),
 4, 4, 4, 4, 'approved', 666,
 '张小刚本周防守表现良好，虽然请假一次但整体状态保持不错。',
 '防守站位更加合理，协防意识增强；对抗赛中表现稳健，不畏惧身体对抗。',
 '建议加强铲球时机判断，避免不必要的犯规；提高出球速度，减少后场风险。',
 '对防守体系的理解还需要深化，建议多观看防守集锦学习。',
 '下周重点训练铲球时机和出球速度。',
 0, datetime('now', '-4 days'),
 datetime('now', '-7 days'), datetime('now'));

-- U12二队 上周周报 (球员ID: 1005陈小龙, 1006赵小虎)
INSERT OR REPLACE INTO weekly_reports (
  id, team_id, player_id, coach_id, week_start, week_end, deadline,
  training_count, training_duration, absence_count, absence_reason,
  knowledge_summary, technical_content, tactical_content, physical_condition, match_performance,
  self_attitude_rating, self_technique_rating, self_teamwork_rating, improvements_detail, weaknesses,
  fatigue_level, injuries, sleep_quality, diet_condition, message_to_coach, attachments,
  submit_status, submitted_at,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  review_status, review_coach_id, review_comment, strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award, reviewed_at,
  created_at, updated_at
) VALUES
-- 陈小龙 - U12二队 - 上周已审核 (approved)
(4, 3, 1005, 666, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59',
 4, 360, 0, '',
 '本周学习了前锋跑位和射门技巧，理解了前锋的责任。',
 '重点练习了射门和头球攻门，射门准确性有所提高。',
 '学习了反越位跑位和前插时机，在对抗中有所尝试。',
 '进行了爆发力和速度训练，启动速度有提升。',
 '在对抗赛中打进2球，表现亮眼。',
 5, 5, 4, '射门准确性提高，跑位更加灵活。', '头球技术还需要加强，身体对抗稍显不足。',
 2, '无伤病', 4, '饮食正常，注意补充能量。', '感谢教练的指导！', '[]',
 'submitted', datetime('now', '-5 days'),
 5, 5, 4, 4, 'approved', 666,
 '陈小龙本周表现出色，作为前锋展现了良好的射门能力和跑位意识。',
 '射门准确性明显提高，跑位灵活；对抗赛中进球效率高，展现了良好的得分能力。',
 '建议加强头球技术训练，提高身体对抗能力；学习更多前锋跑位技巧。',
 '对反越位战术的理解还需要加深，建议多观看前锋集锦学习。',
 '下周重点训练头球技术和身体对抗。',
 1, datetime('now', '-4 days'),
 datetime('now', '-7 days'), datetime('now')),

-- 赵小虎 - U12二队 - 上周已审核 (approved)
(5, 3, 1006, 666, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59',
 4, 360, 0, '',
 '本周学习了中场调度和传球选择，理解了组织核心的作用。',
 '练习了长短传结合和一脚传球，传球质量有所提高。',
 '学习了中场控球和节奏控制，比赛掌控力增强。',
 '进行了耐力和灵活性训练，体能充沛。',
 '在对抗赛中助攻1次，组织了多次有效进攻。',
 5, 4, 5, '传球选择更加合理，能够控制比赛节奏。', '防守回追速度还需要提高，覆盖面可以更广泛。',
 3, '无伤病', 4, '饮食正常。', '希望教练能多指导中场组织技巧。', '[]',
 'submitted', datetime('now', '-5 days'),
 5, 4, 5, 4, 'approved', 666,
 '赵小虎本周表现稳定，作为中场组织者展现了良好的传球能力和团队意识。',
 '传球选择合理，能够控制比赛节奏；助攻能力强，团队配合意识好。',
 '建议加强防守回追速度，提高中场覆盖面积；增加远射训练。',
 '对比赛节奏的控制还需要更精准，建议多观察优秀中场球员。',
 '下周重点训练防守回追和远射能力。',
 0, datetime('now', '-4 days'),
 datetime('now', '-7 days'), datetime('now'));

-- U14精英 上周周报 (球员ID: 1007马小军, 1008周小杰)
INSERT OR REPLACE INTO weekly_reports (
  id, team_id, player_id, coach_id, week_start, week_end, deadline,
  training_count, training_duration, absence_count, absence_reason,
  knowledge_summary, technical_content, tactical_content, physical_condition, match_performance,
  self_attitude_rating, self_technique_rating, self_teamwork_rating, improvements_detail, weaknesses,
  fatigue_level, injuries, sleep_quality, diet_condition, message_to_coach, attachments,
  submit_status, submitted_at,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  review_status, review_coach_id, review_comment, strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award, reviewed_at,
  created_at, updated_at
) VALUES
-- 马小军 - U14精英 - 上周已审核 (approved)
(6, 4, 1007, 666, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59',
 5, 450, 0, '',
 '本周进行了高强度的战术训练，学习了高位逼抢和快速反击战术。',
 '技术训练重点为个人突破和变向运球，技术动作更加熟练。',
 '战术训练包括位置轮换和攻防转换，战术执行力提升。',
 '进行了高强度的体能储备训练，包括间歇跑和力量训练。',
 '在对抗赛中打进1球助攻1次，表现全面。',
 5, 5, 4, '个人突破能力提高，战术执行力增强。', '体能分配还需要优化，比赛后半段略显疲惫。',
 4, '轻微脚踝扭伤，已恢复', 3, '饮食注意营养搭配，补充蛋白质。', '精英队的训练强度很大，我会坚持！', '[]',
 'submitted', datetime('now', '-5 days'),
 5, 5, 5, 5, 'approved', 666,
 '马小军本周表现优秀，作为精英队球员展现了全面的技术和战术能力。',
 '个人突破能力突出，战术执行力强；对抗赛中表现全面，攻守俱佳。',
 '建议优化体能分配，提高比赛后半段表现；加强防守位置感训练。',
 '对高位逼抢战术的理解执行到位，建议继续保持。',
 '下周重点训练体能分配和防守位置感。',
 1, datetime('now', '-4 days'),
 datetime('now', '-7 days'), datetime('now')),

-- 周小杰 - U14精英 - 上周已审核 (approved)
(7, 4, 1008, 666, '2026-03-30', '2026-04-05', '2026-04-08 23:59:59',
 5, 450, 0, '',
 '本周学习了防守组织和协防体系，理解了防线指挥官的责任。',
 '技术训练重点为头球解围和长传发动进攻，技术运用更加合理。',
 '战术训练包括防线整体移动和造越位战术，防守组织能力提升。',
 '进行了力量训练和弹跳训练，对抗能力增强。',
 '在对抗赛中防守稳健，多次成功解围。',
 5, 4, 5, '防守组织能力提高，指挥防线更加自信。', '出球速度还需要提高，有时处理球过于犹豫。',
 3, '无伤病', 4, '饮食正常。', '希望能多学习防守组织技巧。', '[]',
 'submitted', datetime('now', '-5 days'),
 5, 4, 5, 4, 'approved', 666,
 '周小杰本周表现稳定，作为防线核心展现了良好的组织能力和防守意识。',
 '防守组织能力突出，指挥防线自信；对抗赛中表现稳健，是防线定海神针。',
 '建议提高出球速度，减少后场风险；增加向前传球训练，提升进攻参与度。',
 '对防线整体移动战术执行到位，建议继续保持。',
 '下周重点训练出球速度和向前传球能力。',
 0, datetime('now', '-4 days'),
 datetime('now', '-7 days'), datetime('now'));

-- ============================================================
-- 3. 创建本周周报数据 (2026-04-06 ~ 2026-04-13) - 混合状态
-- ============================================================

-- U12一队 本周周报 (球员ID: 1001王小明, 1002李小强, 1003张小刚)
INSERT OR REPLACE INTO weekly_reports (
  id, team_id, player_id, coach_id, week_start, week_end, deadline,
  training_count, training_duration, absence_count, absence_reason,
  knowledge_summary, technical_content, tactical_content, physical_condition, match_performance,
  self_attitude_rating, self_technique_rating, self_teamwork_rating, improvements_detail, weaknesses,
  fatigue_level, injuries, sleep_quality, diet_condition, message_to_coach, attachments,
  submit_status, submitted_at,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  review_status, review_coach_id, review_comment, strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award, reviewed_at,
  created_at, updated_at
) VALUES
-- 王小明 - U12一队 - 本周已审核 (approved)
(8, 2, 1001, 666, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 3, 270, 0, '',
 '本周继续巩固传球配合，学习了边路突破和下底传中战术。',
 '练习了边路运球和传中技术，传中准确性有所提高。',
 '学习了边路突破和倒三角传中，在训练中多次尝试。',
 '进行了爆发力训练和折返跑，体能状态良好。',
 '本周暂无比赛，训练表现积极。',
 5, 4, 4, '边路突破更加自信，传中技术有进步。', '逆足使用还需要加强，传中时机把握可以更好。',
 2, '无伤病', 4, '饮食正常，注意营养搭配。', '希望能有机会在实战中检验边路技术。', '[]',
 'submitted', datetime('now', '-2 days'),
 5, 4, 4, 4, 'approved', 666,
 '王小明本周继续保持良好状态，边路训练进步明显。',
 '边路突破自信，传中准确性提高；训练态度积极主动。',
 '建议加强逆足训练，提高左右脚均衡性；传中时机需要更精准。',
 '对边路战术的理解执行到位，建议继续保持。',
 '下周重点训练逆足技术和传中时机。',
 0, datetime('now', '-1 day'),
 datetime('now'), datetime('now')),

-- 李小强 - U12一队 - 本周待审核 (submitted)
(9, 2, 1002, 666, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 3, 270, 0, '',
 '本周学习了后腰防守和中场拦截，理解了后腰的防守职责。',
 '练习了正面拦截和侧身拦截技术，防守动作更加规范。',
 '学习了中场防守站位和协防配合，防守意识增强。',
 '进行了力量训练和耐力训练，体能储备充足。',
 '在训练对抗中多次成功拦截，表现稳健。',
 5, 4, 4, '防守拦截能力提高，中场控制力增强。', '出球速度还可以更快，有时过于求稳。',
 3, '无伤病', 4, '饮食正常。', '希望教练能指导后腰防守技巧。', '[]',
 'submitted', datetime('now', '-1 day'),
 0, 0, 0, 0, 'pending', 0, '', '', '', '', '', 0, NULL,
 datetime('now'), datetime('now')),

-- 张小刚 - U12一队 - 本周草稿 (draft)
(10, 2, 1003, 666, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 2, 180, 1, '周二因身体不适请假一次。',
 '本周因请假只参加了部分训练，主要复习了防守站位。',
 '练习了防守站位和盯人防守，但训练量不足。',
 '因病请假，战术学习有所欠缺。',
 '训练强度因请假有所降低。',
 '本周暂无比赛。',
 3, 3, 3, '请假影响了训练进度，需要补课。', '训练连续性还需要加强。',
 2, '感冒已好转', 3, '注意饮食，加强营养。', '希望尽快恢复状态。', '[]',
 'draft', NULL,
 0, 0, 0, 0, 'pending', 0, '', '', '', '', '', 0, NULL,
 datetime('now'), datetime('now'));

-- U12二队 本周周报 (球员ID: 1005陈小龙, 1006赵小虎) - 全部草稿
INSERT OR REPLACE INTO weekly_reports (
  id, team_id, player_id, coach_id, week_start, week_end, deadline,
  training_count, training_duration, absence_count, absence_reason,
  knowledge_summary, technical_content, tactical_content, physical_condition, match_performance,
  self_attitude_rating, self_technique_rating, self_teamwork_rating, improvements_detail, weaknesses,
  fatigue_level, injuries, sleep_quality, diet_condition, message_to_coach, attachments,
  submit_status, submitted_at,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  review_status, review_coach_id, review_comment, strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award, reviewed_at,
  created_at, updated_at
) VALUES
-- 陈小龙 - U12二队 - 本周草稿 (draft)
(11, 3, 1005, 666, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 0, 0, 0, '', '', '', '', '', '', 0, 0, 0, '', '', 0, '', 0, '', '', '[]', 'draft', NULL,
 0, 0, 0, 0, 'pending', 0, '', '', '', '', '', 0, NULL,
 datetime('now'), datetime('now')),

-- 赵小虎 - U12二队 - 本周草稿 (draft)
(12, 3, 1006, 666, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 0, 0, 0, '', '', '', '', '', '', 0, 0, 0, '', '', 0, '', 0, '', '', '[]', 'draft', NULL,
 0, 0, 0, 0, 'pending', 0, '', '', '', '', '', 0, NULL,
 datetime('now'), datetime('now'));

-- U14精英 本周周报 (球员ID: 1007马小军, 1008周小杰)
INSERT OR REPLACE INTO weekly_reports (
  id, team_id, player_id, coach_id, week_start, week_end, deadline,
  training_count, training_duration, absence_count, absence_reason,
  knowledge_summary, technical_content, tactical_content, physical_condition, match_performance,
  self_attitude_rating, self_technique_rating, self_teamwork_rating, improvements_detail, weaknesses,
  fatigue_level, injuries, sleep_quality, diet_condition, message_to_coach, attachments,
  submit_status, submitted_at,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  review_status, review_coach_id, review_comment, strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award, reviewed_at,
  created_at, updated_at
) VALUES
-- 马小军 - U14精英 - 本周已审核 (approved)
(13, 4, 1007, 666, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 4, 360, 0, '',
 '本周进行了专项射门训练，学习了不同角度的射门技巧。',
 '练习了正脚背射门和脚内侧推射，射门力量有所提高。',
 '学习了跑位接应和最后一传，进攻配合更加默契。',
 '进行了高强度间歇训练，体能储备良好。',
 '在队内对抗赛中打进2球，表现亮眼。',
 5, 5, 4, '射门力量和准确性提高，跑位更加灵活。', '头球攻门还需要加强，身体对抗可以更强硬。',
 3, '无伤病', 4, '饮食注意补充蛋白质。', '希望能继续保持状态！', '[]',
 'submitted', datetime('now', '-2 days'),
 5, 5, 4, 5, 'approved', 666,
 '马小军本周表现优异，射门训练进步明显，对抗赛中展现了出色的得分能力。',
 '射门力量和准确性都有提高，跑位灵活多变；对抗赛中进球效率高。',
 '建议加强头球攻门训练，提高身体对抗硬度；保持目前的训练状态。',
 '对射门时机的把握更加精准，建议继续保持。',
 '下周重点训练头球攻门和身体对抗。',
 1, datetime('now', '-1 day'),
 datetime('now'), datetime('now')),

-- 周小杰 - U14精英 - 本周草稿 (draft)
(14, 4, 1008, 666, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 3, 270, 0, '',
 '本周学习了防守预判和提前移动，理解了主动防守的重要性。',
 '练习了防守站位和预判拦截，但还需要更多实战检验。',
 '学习了防线整体压迫和造越位，战术执行有进步。',
 '进行了核心力量训练，身体对抗能力增强。',
 '在对抗赛中防守稳健，但还可以更主动。',
 4, 4, 4, '防守预判能力有所提高，出球更加果断。', '防守预判还可以更精准，有时启动稍慢。',
 3, '无伤病', 4, '饮食正常。', '希望教练能指导防守预判技巧。', '[]',
 'draft', NULL,
 0, 0, 0, 0, 'pending', 0, '', '', '', '', '', 0, NULL,
 datetime('now'), datetime('now'));

-- ============================================================
-- 4. 为老球员账号创建周报数据（部分球员）
-- ============================================================

-- 查找老球员ID并创建简单周报数据
-- 球员ID: 1009陈小明, 1010李博文, 1011张天宇, 1012王浩然

-- 陈小明 (广州小飞侠) - U12梯队 (team_id: 6)
INSERT OR REPLACE INTO weekly_reports (
  id, team_id, player_id, coach_id, week_start, week_end, deadline,
  training_count, training_duration, absence_count, absence_reason,
  knowledge_summary, technical_content, tactical_content, physical_condition, match_performance,
  self_attitude_rating, self_technique_rating, self_teamwork_rating, improvements_detail, weaknesses,
  fatigue_level, injuries, sleep_quality, diet_condition, message_to_coach, attachments,
  submit_status, submitted_at,
  coach_attitude_rating, coach_technique_rating, coach_tactics_rating, coach_knowledge_rating,
  review_status, review_coach_id, review_comment, strengths_acknowledgment, suggestions, knowledge_feedback, next_week_focus, recommend_award, reviewed_at,
  created_at, updated_at
) VALUES
-- 陈小明 - 本周已提交待审核
(15, 6, 1009, 667, '2026-04-06', '2026-04-12', '2026-04-15 23:59:59',
 3, 240, 0, '',
 '本周学习了基础控球和传球，作为新球员在适应球队节奏。',
 '练习了停球和短传，控球稳定性有进步。',
 '学习了基本的跑位和接应，配合意识在培养中。',
 '进行了基础体能训练，体能有所提升。',
 '本周参加了队内对抗，表现积极。',
 4, 3, 3, '控球能力有进步，适应球队节奏较快。', '传球准确性还需要提高，配合默契度需要时间。',
 2, '无伤病', 4, '饮食正常。', '希望尽快融入球队！', '[]',
 'submitted', datetime('now', '-1 day'),
 0, 0, 0, 0, 'pending', 0, '', '', '', '', '', 0, NULL,
 datetime('now'), datetime('now'));

-- ============================================================
-- 5. 数据汇总统计
-- ============================================================

-- 统计周报数据
SELECT '周报数据统计' as report;
SELECT 
  '本周周报' as period,
  COUNT(*) as total,
  SUM(CASE WHEN submit_status = 'draft' THEN 1 ELSE 0 END) as draft_count,
  SUM(CASE WHEN submit_status = 'submitted' THEN 1 ELSE 0 END) as submitted_count,
  SUM(CASE WHEN review_status = 'approved' THEN 1 ELSE 0 END) as approved_count
FROM weekly_reports 
WHERE week_start = '2026-04-06';

SELECT 
  '上周周报' as period,
  COUNT(*) as total,
  SUM(CASE WHEN submit_status = 'draft' THEN 1 ELSE 0 END) as draft_count,
  SUM(CASE WHEN submit_status = 'submitted' THEN 1 ELSE 0 END) as submitted_count,
  SUM(CASE WHEN review_status = 'approved' THEN 1 ELSE 0 END) as approved_count
FROM weekly_reports 
WHERE week_start = '2026-03-30';

SELECT '周报数据填充完成！' as result;
