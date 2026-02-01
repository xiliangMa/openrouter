# API 文档

MassRouter SaaS平台提供完整的RESTful API接口，支持大模型统一接入、用户管理、计费统计等功能。

## 基础信息

### API端点
- **开发环境**: `http://localhost:8080/api/v1`
- **生产环境**: `https://api.massrouter.ai/api/v1`

### 认证方式
1. **API密钥认证**: 在请求头中添加 `Authorization: Bearer {api_key}`
2. **JWT令牌认证**: 用户登录后获取的访问令牌
3. **OAuth2**: 支持第三方OAuth2提供商登录

### 请求/响应格式
- **请求**: JSON格式，Content-Type: `application/json`
- **响应**: JSON格式，统一响应结构

### 统一响应格式
```json
{
  "success": true,
  "data": {},
  "message": "操作成功",
  "timestamp": "2024-01-01T00:00:00Z"
}
```

错误响应：
```json
{
  "success": false,
  "error": {
    "code": "ERR_001",
    "message": "错误描述",
    "details": {}
  },
  "timestamp": "2024-01-01T00:00:00Z"
}
```

## 认证API

### 用户注册
```http
POST /auth/register
```

**请求体**:
```json
{
  "email": "user@example.com",
  "username": "username",
  "password": "secure_password",
  "confirm_password": "secure_password"
}
```

**响应**:
```json
{
  "success": true,
  "data": {
    "user": {
      "id": "user_id",
      "email": "user@example.com",
      "username": "username"
    },
    "tokens": {
      "access_token": "jwt_access_token",
      "refresh_token": "jwt_refresh_token",
      "expires_in": 3600
    }
  }
}
```

### 用户登录
```http
POST /auth/login
```

**请求体**:
```json
{
  "email": "user@example.com",
  "password": "secure_password"
}
```

### OAuth2登录
```http
GET /auth/oauth/{provider}
```
支持的Provider: `github`, `google`, `wechat`, `feishu`, `twitter`, `facebook`, `xiaohongshu`

### 刷新令牌
```http
POST /auth/refresh
```

**请求头**:
```
Authorization: Bearer {refresh_token}
```

## 用户管理API

### 获取当前用户信息
```http
GET /users/me
```

### 更新用户信息
```http
PUT /users/me
```

### 获取用户API密钥列表
```http
GET /users/me/api-keys
```

### 创建API密钥
```http
POST /users/me/api-keys
```

**请求体**:
```json
{
  "name": "Production Key",
  "permissions": ["read", "write"],
  "expires_at": "2024-12-31T23:59:59Z"
}
```

### 删除API密钥
```http
DELETE /users/me/api-keys/{key_id}
```

## 模型API

### 获取模型列表
```http
GET /models
```

**查询参数**:
- `page`: 页码 (默认: 1)
- `limit`: 每页数量 (默认: 20)
- `category`: 模型分类
- `provider`: 供应商
- `search`: 搜索关键词
- `sort_by`: 排序字段 (popularity, price, name)
- `order`: 排序方向 (asc, desc)

### 获取模型详情
```http
GET /models/{model_id}
```

### 搜索模型
```http
GET /models/search
```

## 计费API

### 获取用户余额
```http
GET /billing/balance
```

### 获取消费记录
```http
GET /billing/transactions
```

**查询参数**:
- `start_date`: 开始日期
- `end_date`: 结束日期
- `type`: 类型 (payment, consumption, refund)

### 创建充值订单
```http
POST /billing/payments
```

**请求体**:
```json
{
  "amount": 100.00,
  "currency": "CNY",
  "payment_method": "alipay"
}
```

支持的支付方式: `alipay`, `wechat_pay`, `stripe`, `bank_transfer`

### 查询订单状态
```http
GET /billing/payments/{payment_id}
```

## 统计API

### 获取热门模型排行榜
```http
GET /statistics/leaderboard
```

**查询参数**:
- `period`: 时间周期 (day, week, month, year)
- `category`: 分类筛选
- `limit`: 返回数量 (默认: 10)

### 获取模型使用统计
```http
GET /statistics/models/{model_id}/usage
```

### 获取用户使用统计
```http
GET /statistics/users/me/usage
```

## 管理API (仅管理员)

### 获取所有用户列表
```http
GET /admin/users
```

### 管理用户状态
```http
PUT /admin/users/{user_id}/status
```

**请求体**:
```json
{
  "status": "active|inactive|suspended"
}
```

### 获取系统统计
```http
GET /admin/statistics/dashboard
```

### 更新模型价格
```http
PUT /admin/models/{model_id}/pricing
```

## 实时API (WebSocket)

### 连接WebSocket
```
ws://localhost:8080/ws
```

### 订阅主题
连接后发送订阅消息：
```json
{
  "action": "subscribe",
  "topic": "billing.updates"
}
```

可用主题:
- `billing.updates`: 计费更新
- `model.usage`: 模型使用统计
- `system.alerts`: 系统告警

## 错误码

### 通用错误码
- `ERR_001`: 请求参数验证失败
- `ERR_002`: 认证失败
- `ERR_003`: 权限不足
- `ERR_004`: 资源不存在
- `ERR_005`: 服务器内部错误

### 业务错误码
- `BILL_001`: 余额不足
- `BILL_002`: 支付失败
- `MODEL_001`: 模型不可用
- `APIKEY_001`: API密钥已过期
- `USER_001`: 用户已被禁用

## 速率限制

- **认证用户**: 1000次/小时
- **API密钥**: 根据密钥权限设置
- **匿名访问**: 100次/小时

## 版本控制

API版本通过URL路径控制，当前版本为v1。

## 更新日志

- **v1.0.0** (2024-01-01): 初始版本发布
