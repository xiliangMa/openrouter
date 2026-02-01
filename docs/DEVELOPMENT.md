# 开发指南

本文档提供MassRouter SaaS平台的本地开发环境配置和开发工作流说明。

## 环境要求

### 必需工具
- **Docker & Docker Compose**: 用于运行数据库和缓存服务
- **Go 1.22+**: 后端开发
- **Node.js 18+**: 前端开发
- **Git**: 版本控制

### 可选工具
- **golangci-lint**: Go代码质量检查
- **Air**: Go热重载开发工具
- **PostgreSQL客户端**: 如`psql`或DBeaver
- **Redis客户端**: 如`redis-cli`或RedisInsight

## 环境设置

### 1. 克隆仓库
```bash
git clone https://github.com/your-org/openrouter-saas.git
cd openrouter-saas
```

### 2. 配置环境变量
```bash
cp .env.example .env
```

编辑`.env`文件，配置必要的环境变量。开发环境可暂时使用默认值，但生产环境必须修改。

### 3. 启动基础设施
```bash
make start
```

此命令会启动：
- PostgreSQL数据库 (端口: 5432)
- Redis缓存 (端口: 6379)

### 4. 验证基础设施
```bash
# 检查PostgreSQL
docker-compose exec postgres pg_isready -U openrouter

# 检查Redis
docker-compose exec redis redis-cli ping
```

## 后端开发

### 项目结构
```
backend/
├── cmd/server/         # 应用入口
├── internal/           # 内部包（不对外暴露）
│   ├── config/        # 配置管理
│   ├── controller/    # HTTP控制器
│   ├── model/         # 数据模型
│   ├── repository/    # 数据访问层
│   ├── service/       # 业务逻辑层
│   └── middleware/    # HTTP中间件
├── pkg/               # 可复用包
├── api/               # API定义
├── migrations/        # 数据库迁移
└── tests/             # 测试文件
```

### 初始化Go模块
```bash
cd backend
go mod init github.com/your-org/openrouter/backend
go mod tidy
```

### 开发工作流

1. **启动开发服务器（热重载）**
   ```bash
   cd backend
   # 安装Air（如果未安装）
   go install github.com/cosmtrek/air@latest
   
   # 使用Air启动
   air
   ```

2. **运行测试**
   ```bash
   make backend-test
   ```

3. **代码质量检查**
   ```bash
   make backend-lint
   ```

4. **数据库迁移**
   ```bash
   # 创建新迁移
   go run ./cmd/migrate create add_users_table sql

   # 应用迁移
   make migrate-up

   # 回滚迁移
   make migrate-down
   ```

### API开发规范

1. **路由定义**：在`internal/controller`中按功能模块组织
2. **请求验证**：使用`github.com/go-playground/validator/v10`
3. **错误处理**：统一错误响应格式
4. **日志记录**：结构化日志，包含请求ID

## 前端开发

### 管理后台 (admin/)

#### 技术栈
- **框架**: Next.js 15 (App Router)
- **语言**: TypeScript
- **样式**: Tailwind CSS + shadcn/ui组件库
- **状态管理**: React Context + SWR
- **表单**: React Hook Form + Zod验证

#### 初始化
```bash
cd admin
npm install
```

#### 开发服务器
```bash
make admin-dev
# 或
cd admin && npm run dev
```

#### 构建
```bash
make admin-build
```

### 门户网站 (portal/)

初始化与开发流程与管理后台类似。

## 数据库开发

### 迁移管理
使用Goose进行数据库迁移管理：

1. **创建迁移**
   ```bash
   cd backend
   go run ./cmd/migrate create add_feature_name sql
   ```

2. **编辑迁移文件**
   - `up`函数：应用更改
   - `down`函数：回滚更改

3. **应用迁移**
   ```bash
   make migrate-up
   ```

### 数据模型规范
1. 所有表必须包含`created_at`和`updated_at`时间戳
2. 使用软删除（`deleted_at`字段）
3. 外键约束使用级联删除或设置为NULL

## 测试

### 后端测试
```bash
# 运行所有测试
make backend-test

# 运行特定包测试
cd backend
go test ./internal/service/...

# 运行测试并生成覆盖率报告
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### 前端测试
```bash
# 管理后台测试
cd admin
npm test

# 门户网站测试
cd portal
npm test
```

## 代码质量

### Go代码规范
1. 遵循[Go代码审查评论](https://github.com/golang/go/wiki/CodeReviewComments)
2. 使用`golangci-lint`进行静态检查
3. 提交前运行`make backend-lint`

### TypeScript规范
1. 使用ESLint和Prettier
2. 严格TypeScript模式
3. 组件使用函数式组件和Hooks

### Git工作流
采用Git Flow分支策略：
- `main`: 生产环境代码
- `develop`: 开发集成分支
- `feature/*`: 功能开发分支
- `release/*`: 发布准备分支
- `hotfix/*`: 紧急修复分支

## 常见问题

### 1. 数据库连接失败
确保PostgreSQL服务已启动：
```bash
docker-compose ps
docker-compose logs postgres
```

### 2. 迁移失败
检查迁移文件语法，可手动连接到数据库调试：
```bash
docker-compose exec postgres psql -U openrouter -d openrouter
```

### 3. 前端热重载不工作
检查Node.js版本和依赖：
```bash
node --version
npm ci  # 重新安装依赖
```

## 下一步

- [API文档](API.md) - API接口详细说明
- [部署指南](DEPLOYMENT.md) - 生产环境部署
- [架构设计](ARCHITECTURE.md) - 系统架构设计
