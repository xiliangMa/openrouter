# MassRouter SaaS Platform

一站式大模型统一代理平台，提供60+供应商、300+大模型的统一接入、管理、计费和统计分析服务。

[![Go Version](https://img.shields.io/badge/Go-1.22+-00ADD8?style=flat&logo=go)](https://golang.org/)
[![Next.js](https://img.shields.io/badge/Next.js-15-000000?style=flat&logo=next.js)](https://nextjs.org/)
[![TypeScript](https://img.shields.io/badge/TypeScript-5.0+-3178C6?style=flat&logo=typescript)](https://www.typescriptlang.org/)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16-336791?style=flat&logo=postgresql)](https://www.postgresql.org/)
[![Redis](https://img.shields.io/badge/Redis-7-DC382D?style=flat&logo=redis)](https://redis.io/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

## ✨ 特性

### 🔌 统一模型接入
- 支持60+活跃供应商的300+大模型
- 覆盖OpenAI、Anthropic、Google、Claude等主流厂商
- 统一API接口，简化集成复杂度

### 💰 智能计费管理
- 按Token精确计费，实时扣费
- 支持多种支付方式：微信、支付宝、银行卡、信用卡
- 灵活的定价策略，支持免费和收费模型

### 📊 数据统计与排名
- 实时模型调用统计与性能监控
- 热门模型排行榜与市场份额分析
- 多维度数据分析与可视化报表

### 🔐 安全与权限管理
- 多因素认证与OAuth2登录（微信、飞书、GitHub等）
- 细粒度API密钥管理与权限控制
- 企业级安全审计与操作日志

### 🌐 现代化管理界面
- 响应式设计，支持桌面端和移动端
- 多语言支持（中文/英文）
- 直观的仪表板与数据可视化

## 🏗️ 技术栈

### 后端服务
- **语言**: Go 1.22+
- **框架**: Gin, Gorm
- **数据库**: PostgreSQL 16
- **缓存**: Redis 7
- **认证**: JWT, OAuth2

### 前端应用
- **管理后台**: Next.js 15, TypeScript, Tailwind CSS, shadcn/ui
- **门户网站**: Next.js 15, TypeScript, Tailwind CSS
- **样式**: Tailwind CSS, CSS Modules
- **状态管理**: React Context + SWR

### 基础设施
- **容器化**: Docker, Docker Compose
- **数据库迁移**: Goose
- **API文档**: Swagger/OpenAPI 3.0
- **代码质量**: golangci-lint, ESLint, Prettier

## 🚀 快速开始

### 前置要求
- Docker & Docker Compose
- Go 1.22+ (用于本地开发)
- Node.js 18+ (用于前端开发)

### 环境设置

1. **克隆仓库**
   ```bash
   git clone https://github.com/your-org/openrouter-saas.git
   cd openrouter-saas
   ```

2. **配置环境变量**
   ```bash
   cp .env.example .env
   # 编辑 .env 文件，配置必要的环境变量
   ```

3. **启动服务**
   ```bash
   make start
   ```

4. **应用数据库迁移**
   ```bash
   make migrate-up
   ```

5. **访问应用**
   - 管理后台: http://localhost:3000
   - 门户网站: http://localhost:3001
   - API文档: http://localhost:8080/swagger/index.html

### 开发命令

```bash
# 启动所有服务
make start

# 停止所有服务
make stop

# 查看日志
make logs

# 运行后端测试
make backend-test

# 启动管理后台开发服务器
make admin-dev

# 启动门户网站开发服务器
make portal-dev

# 代码质量检查
make lint
```

## 📁 项目结构

```
openrouter-saas/
├── backend/           # Go后端API服务
├── admin/            # 管理后台前端 (Next.js)
├── portal/           # 门户网站前端 (Next.js)
├── docs/             # 项目文档
├── docker-compose.yml # Docker编排配置
├── Makefile          # 开发命令
├── .env.example      # 环境变量模板
└── README.md         # 项目说明
```

## 📚 文档

- [开发指南](docs/DEVELOPMENT.md) - 本地开发环境配置
- [API文档](docs/API.md) - API接口详细说明
- [部署指南](docs/DEPLOYMENT.md) - 生产环境部署
- [架构设计](docs/ARCHITECTURE.md) - 系统架构设计文档

## 🔧 开发

详细开发指南请参考 [DEVELOPMENT.md](docs/DEVELOPMENT.md)。

## 🤝 贡献

欢迎提交Issue和Pull Request！请参考 [CONTRIBUTING.md](docs/CONTRIBUTING.md)。

## 📄 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 🔗 相关链接

- [MassRouter.ai](https://openrouter.ai) - 灵感来源
- [LobeHub](https://lobehub.com/zh) - 界面设计参考
- [DataLearner](https://www.datalearner.com) - 模型数据参考
