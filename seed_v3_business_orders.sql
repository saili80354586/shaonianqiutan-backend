-- ============================================
-- 少年球探 - 订单数据 v3.1 (修正表结构)
-- ============================================

-- 王小明 (2001) 的订单 - 2个
INSERT INTO orders (id, user_id, order_no, amount, status, analyst_id, created_at, updated_at) VALUES
(1, 2001, 'ORD20260401001', 299.00, 'completed', 30, datetime('now', '-30 days'), datetime('now', '-28 days')),
(2, 2001, 'ORD20260402001', 399.00, 'completed', 33, datetime('now', '-15 days'), datetime('now', '-13 days'));

-- 李小强 (2002) 的订单 - 1个
INSERT INTO orders (id, user_id, order_no, amount, status, analyst_id, created_at, updated_at) VALUES
(3, 2002, 'ORD20260403002', 299.00, 'completed', 30, datetime('now', '-20 days'), datetime('now', '-18 days'));

-- 张小刚 (2003) 的订单 - 1个
INSERT INTO orders (id, user_id, order_no, amount, status, analyst_id, created_at, updated_at) VALUES
(4, 2003, 'ORD20260404003', 299.00, 'completed', 31, datetime('now', '-25 days'), datetime('now', '-23 days'));

-- 陈小龙 (2005) 的订单 - 1个
INSERT INTO orders (id, user_id, order_no, amount, status, analyst_id, created_at, updated_at) VALUES
(5, 2005, 'ORD20260405005', 299.00, 'completed', 30, datetime('now', '-10 days'), datetime('now', '-8 days'));

.print '订单数据导入完成: ' || (SELECT COUNT(*) FROM orders) || ' 条'
