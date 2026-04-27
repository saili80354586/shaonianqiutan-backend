-- 迁移脚本：俱乐部教练管理模块
-- 1. 删除 team_coaches 旧唯一索引 (team_id, user_id)
-- 2. 创建新唯一索引 (team_id, user_id, role) 支持教练兼任多队不同角色
-- 3. club_coaches 表由 AutoMigrate 自动创建

-- 必须先删除旧索引
DROP INDEX IF EXISTS idx_team_coach;

-- 创建新唯一索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_team_coach_role ON team_coaches(team_id, user_id, role);
