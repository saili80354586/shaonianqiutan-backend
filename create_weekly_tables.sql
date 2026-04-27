-- ============================================================
-- 周报表结构创建脚本
-- ============================================================

-- 删除旧表（如果存在）
DROP TABLE IF EXISTS weekly_reports;
DROP TABLE IF EXISTS weekly_report_periods;

-- 创建周报周期表
CREATE TABLE IF NOT EXISTS weekly_report_periods (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    deadline DATETIME,
    total_players INTEGER DEFAULT 0,
    submitted_count INTEGER DEFAULT 0,
    pending_count INTEGER DEFAULT 0,
    overdue_count INTEGER DEFAULT 0,
    reviewed_count INTEGER DEFAULT 0,
    status VARCHAR(20) DEFAULT 'active',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建周报表
CREATE TABLE IF NOT EXISTS weekly_reports (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    team_id INTEGER NOT NULL,
    player_id INTEGER NOT NULL,
    coach_id INTEGER NOT NULL,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    deadline DATETIME,
    
    -- 训练出勤情况
    training_count INTEGER DEFAULT 0,
    training_duration INTEGER DEFAULT 0,
    absence_count INTEGER DEFAULT 0,
    absence_reason TEXT,
    
    -- 训练内容反馈
    knowledge_summary TEXT,
    technical_content TEXT,
    tactical_content TEXT,
    physical_condition TEXT,
    match_performance TEXT,
    
    -- 自我评价 - 多维度评分
    self_attitude_rating INTEGER DEFAULT 0,
    self_technique_rating INTEGER DEFAULT 0,
    self_teamwork_rating INTEGER DEFAULT 0,
    improvements_detail TEXT,
    weaknesses TEXT,
    
    -- 身体状态反馈
    fatigue_level INTEGER DEFAULT 3,
    injuries TEXT,
    sleep_quality INTEGER DEFAULT 3,
    diet_condition TEXT,
    message_to_coach TEXT,
    attachments TEXT,
    
    -- 提交状态
    submit_status VARCHAR(20) DEFAULT 'draft',
    submitted_at DATETIME,
    
    -- 教练评价 - 多维度评分
    coach_attitude_rating INTEGER DEFAULT 0,
    coach_technique_rating INTEGER DEFAULT 0,
    coach_tactics_rating INTEGER DEFAULT 0,
    coach_knowledge_rating INTEGER DEFAULT 0,
    
    -- 教练评语
    review_status VARCHAR(20) DEFAULT 'pending',
    review_coach_id INTEGER,
    review_comment TEXT,
    strengths_acknowledgment TEXT,
    suggestions TEXT,
    knowledge_feedback TEXT,
    next_week_focus TEXT,
    recommend_award BOOLEAN DEFAULT 0,
    reviewed_at DATETIME,
    
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_weekly_reports_team ON weekly_reports(team_id);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_player ON weekly_reports(player_id);
CREATE INDEX IF NOT EXISTS idx_weekly_reports_week ON weekly_reports(week_start);
CREATE INDEX IF NOT EXISTS idx_weekly_report_periods_team ON weekly_report_periods(team_id);

SELECT '周报表结构创建完成！' as result;
