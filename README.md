# 少年球探后端 - Go重构版

基于 Gin + GORM + SQLite 重构的少年球探后端服务。

## 技术栈

- **Web框架:** [Gin](https://github.com/gin-gonic/gin)
- **ORM:** [GORM](https://gorm.io)
- **数据库:** SQLite
- **认证:** JWT (golang-jwt/jwt/v5)
- **密码加密:** golang.org/x/crypto/bcrypt

## 项目结构

```
.
├── config/         # 配置和数据库连接
├── controllers/    # 控制器层（处理HTTP请求）
├── middleware/     # 中间件（认证、CORS）
├── models/         # 数据模型和Repository层
├── routes/         # 路由定义
├── services/       # 业务逻辑层
├── utils/          # 工具函数
├── main.go         # 入口文件
└── go.mod
```

## API接口

保持与原Node.js项目兼容：

### 认证模块 `/api/auth`

- `POST /api/auth/send-code` - 发送短信验证码
  - Body: `{ "phone": "13800000000", "type": "register|reset" }`

- `POST /api/auth/register` - 用户注册
  - Body: `{ "phone": "13800000000", "code": "123456", "password": "password" }`

- `POST /api/auth/login` - 用户登录
  - Body: `{ "phone": "13800000000", "password": "password" }`

- `POST /api/auth/reset-password` - 重置密码
  - Body: `{ "phone": "13800000000", "code": "123456", "password": "newpassword" }`

- `GET /api/auth/me` - 获取当前用户信息 (需要认证)

- `PUT /api/auth/me` - 更新当前用户信息 (需要认证)

### 报告模块 `/api/report`

所有路由都需要认证

- `POST /api/report` - 创建报告 (分析师权限)
  - Body: `{ "order_id": 1, "player_name": "张三", ... }`

- `GET /api/report/:id` - 获取报告详情

- `GET /api/report/:id/download` - 下载PDF报告

- `GET /api/report/my` - 获取我的报告列表（作为买家）
  - Query: `page=1&pageSize=10`

- `GET /api/report/published` - 获取我发布的报告列表（作为分析师）
  - Query: `page=1&pageSize=10`

- `POST /api/report/:id/regenerate` - 重新生成PDF (分析师权限)

## 数据模型

### 用户表 (users)
存储用户信息，支持普通用户、分析师、管理员三种角色，包含完整的球员/分析师资料字段。

### 短信验证码表 (sms_codes)
存储短信验证码，支持注册和重置密码两种类型，自动过期清理。

### 报告表 (reports)
存储球探报告信息，包含球员基本信息、报告内容、PDF链接、状态等。

## 启动

```bash
# 复制环境变量文件
cp .env.example .env
# 编辑 .env 文件配置

# 编译
go build -o shaonianqiutan-backend .

# 运行
./shaonianqiutan-backend
```

## 开发模式

开发模式下 (`NODE_ENV=development`)，短信验证码会直接返回给前端，不需要实际发送短信，方便开发测试。

## 许可证

私有
