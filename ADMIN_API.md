# 管理后台 API 文档

## 管理员登录

### POST /api/admin/login

**描述**: 管理员登录

**请求体**:
```json
{
  "username": "admin",
  "password": "admin123"
}
```

**响应**:
```json
{
  "success": true,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "admin": {
      "id": 1,
      "phone": "admin",
      "nickname": "管理员",
      "role": "admin",
      "status": "active"
    }
  }
}
```

---

## 数据统计

### GET /api/admin/statistics

**描述**: 获取核心数据统计

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "total_users": 100,
    "total_orders": 50,
    "total_reports": 48,
    "total_revenue": 50000.00,
    "today_new_users": 5,
    "today_orders": 3,
    "today_revenue": 3000.00,
    "pending_applications": 2,
    "pending_reports": 1
  }
}
```

### GET /api/admin/dashboard/growth?days=30

**描述**: 获取增长数据

**参数**:
- days: 天数 (默认: 30)

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": [
    {
      "date": "2026-03-01",
      "users": 10,
      "orders": 5,
      "revenue": 1000.00
    }
  ]
}
```

---

## 用户管理

### GET /api/admin/users?page=1&pageSize=10

**描述**: 获取用户列表

**参数**:
- page: 页码 (默认: 1)
- pageSize: 每页数量 (默认: 10)

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 1,
        "phone": "13800138000",
        "nickname": "张三",
        "role": "user",
        "status": "active",
        "created_at": "2026-03-01T10:00:00Z"
      }
    ],
    "total": 100,
    "page": 1,
    "pageSize": 10
  }
}
```

### PUT /api/admin/users/:id/status

**描述**: 更新用户状态

**请求体**:
```json
{
  "status": "active"
}
```

**状态值**:
- active: 正常
- inactive: 停用

**响应**:
```json
{
  "success": true,
  "message": "状态更新成功"
}
```

### DELETE /api/admin/users/:id

**描述**: 删除用户

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "message": "删除成功"
}
```

---

## 订单管理

### GET /api/admin/orders?page=1&pageSize=10&status=pending

**描述**: 获取订单列表

**参数**:
- page: 页码 (默认: 1)
- pageSize: 每页数量 (默认: 10)
- status: 订单状态 (可选)

**状态值**:
- pending: 待支付
- paid: 已支付
- assigned: 已分配
- processing: 处理中
- completed: 已完成
- cancelled: 已取消

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 1,
        "order_no": "ORD202603290001",
        "user": {
          "id": 1,
          "nickname": "张三"
        },
        "analyst": {
          "id": 2,
          "nickname": "分析师A"
        },
        "amount": 299.00,
        "status": "processing",
        "created_at": "2026-03-01T10:00:00Z"
      }
    ],
    "total": 50,
    "page": 1,
    "pageSize": 10
  }
}
```

### DELETE /api/admin/orders/:id

**描述**: 取消订单

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "message": "订单已取消"
}
```

---

## 分析师管理

### GET /api/admin/analysts?page=1&pageSize=10&status=active

**描述**: 获取分析师列表

**参数**:
- page: 页码 (默认: 1)
- pageSize: 每页数量 (默认: 10)
- status: 状态 (可选)

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 2,
        "phone": "13900139000",
        "nickname": "分析师A",
        "role": "analyst",
        "status": "active",
        "created_at": "2026-03-01T10:00:00Z"
      }
    ],
    "total": 10,
    "page": 1,
    "pageSize": 10
  }
}
```

### PUT /api/admin/analysts/:id/audit

**描述**: 审核分析师

**请求体**:
```json
{
  "status": "approved",
  "remark": "审核通过"
}
```

**状态值**:
- approved: 通过
- rejected: 拒绝

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "message": "审核完成"
}
```

### PUT /api/admin/analysts/:id/status

**描述**: 更新分析师状态

**请求体**:
```json
{
  "status": "active"
}
```

**状态值**:
- active: 正常
- inactive: 停用

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "message": "状态更新成功"
}
```

---

## 报告管理

### GET /api/admin/reports/pending?page=1&pageSize=10

**描述**: 获取待审核报告列表

**参数**:
- page: 页码 (默认: 1)
- pageSize: 每页数量 (默认: 10)

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "data": {
    "list": [
      {
        "id": 1,
        "order_id": 1,
        "analyst_id": 2,
        "status": "processing",
        "created_at": "2026-03-01T10:00:00Z"
      }
    ],
    "total": 5,
    "page": 1,
    "pageSize": 10
  }
}
```

### POST /api/admin/reports/:id/review

**描述**: 审核报告

**请求体**:
```json
{
  "status": "completed",
  "remark": "审核通过"
}
```

**状态值**:
- processing: 处理中
- completed: 已完成
- failed: 失败

**请求头**:
```
Authorization: Bearer <token>
```

**响应**:
```json
{
  "success": true,
  "message": "审核完成"
}
```

---

## 错误响应

所有API在失败时会返回以下格式:

```json
{
  "success": false,
  "message": "错误信息",
  "data": null
}
```

**常见错误码**:
- 400: 请求参数错误
- 401: 未授权/Token无效
- 403: 无权限访问
- 404: 资源不存在
- 500: 服务器内部错误

---

## 测试账号

### 管理员账号
- 用户名: `admin`
- 密码: `admin123`

### 注意事项

1. 所有需要认证的API都必须在请求头中携带JWT Token:
   ```
   Authorization: Bearer <token>
   ```

2. Token有效期为24小时,过期后需要重新登录

3. 管理员权限验证中间件尚未实现,当前仅验证登录状态

4. 分析师审核功能需要扩展分析师模型,当前返回成功但未实际处理

5. 测试数据脚本位于 `seed_admin.sql`
