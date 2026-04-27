-- 创建管理员账号
-- 默认账号: admin / admin123

-- 首先检查管理员是否已存在
-- 注意: 密码 admin123 使用 bcrypt 哈希后的值

INSERT INTO users (phone, password, nickname, role, status, created_at, updated_at)
VALUES ('admin', '$2a$10$Xb3e1Y7J5Q6zK8hJpY9QmOGjK3l0kN7qM5hJzY4vR8xL3n6wP0qQ', '管理员', 'admin', 'active', datetime('now'), datetime('now'))
WHERE NOT EXISTS (
    SELECT 1 FROM users WHERE phone = 'admin'
);

-- 如果上面的SQLite语法不支持,使用这条INSERT OR IGNORE语句:
-- INSERT OR IGNORE INTO users (phone, password, nickname, role, status, created_at, updated_at)
-- VALUES ('admin', '$2a$10$Xb3e1Y7J5Q6zK8hJpY9QmOGjK3l0kN7qM5hJzY4vR8xL3n6wP0qQ', '管理员', 'admin', 'active', datetime('now'), datetime('now'));

-- 注意: 如果您使用MySQL,请使用以下语法:
-- INSERT IGNORE INTO users (phone, password, nickname, role, status, created_at, updated_at)
-- VALUES ('admin', '$2a$10$Xb3e1Y7J5Q6zK8hJpY9QmOGjK3l0kN7qM5hJzY4vR8xL3n6wP0qQ', '管理员', 'admin', 'active', NOW(), NOW());
