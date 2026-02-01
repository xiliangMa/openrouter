# 部署指南

本文档提供MassRouter SaaS平台的生产环境部署说明。

## 部署架构

### 推荐架构
```
用户请求 → Cloudflare/CDN → 负载均衡器 → 应用服务器 → 数据库集群
```

### 组件说明
1. **CDN/代理层**: Cloudflare或类似CDN服务
2. **负载均衡器**: Nginx, HAProxy或云服务商负载均衡
3. **应用服务器**: 后端API服务、前端应用
4. **数据库层**: PostgreSQL主从集群
5. **缓存层**: Redis集群
6. **对象存储**: 用于静态资源存储
7. **监控告警**: Prometheus + Grafana + AlertManager

## 环境要求

### 服务器规格
| 组件 | 最低配置 | 推荐配置 | 说明 |
|------|----------|----------|------|
| 后端API | 2核4GB | 4核8GB+ | 根据并发量调整 |
| 前端应用 | 1核2GB | 2核4GB | 静态资源可CDN缓存 |
| PostgreSQL | 4核8GB | 8核16GB+ | SSD存储，根据数据量调整 |
| Redis | 1核2GB | 2核4GB+ | 内存根据缓存数据量调整 |

### 软件要求
- **操作系统**: Ubuntu 22.04 LTS 或 CentOS 8+
- **容器运行时**: Docker 24+ 或 containerd
- **编排工具**: Docker Compose 或 Kubernetes
- **数据库**: PostgreSQL 16+
- **缓存**: Redis 7+

## 单机部署 (Docker Compose)

适用于中小规模部署。

### 1. 准备服务器
```bash
# 更新系统
apt update && apt upgrade -y

# 安装Docker
curl -fsSL https://get.docker.com -o get-docker.sh
sh get-docker.sh

# 安装Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose
```

### 2. 部署应用
```bash
# 克隆代码
git clone https://github.com/your-org/openrouter-saas.git
cd openrouter-saas

# 配置环境变量
cp .env.example .env.production
# 编辑 .env.production，配置生产环境变量

# 构建镜像
docker-compose -f docker-compose.yml -f docker-compose.prod.yml build

# 启动服务
docker-compose -f docker-compose.yml -f docker-compose.prod.yml up -d
```

### 3. 应用数据库迁移
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml exec backend make migrate-up
```

### 4. 配置反向代理 (Nginx)
创建Nginx配置 `/etc/nginx/sites-available/openrouter`:
```nginx
upstream backend {
    server 127.0.0.1:8080;
}

upstream admin {
    server 127.0.0.1:3000;
}

upstream portal {
    server 127.0.0.1:3001;
}

server {
    listen 80;
    server_name api.openrouter.ai;
    
    location / {
        proxy_pass http://backend;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name admin.openrouter.ai;
    
    location / {
        proxy_pass http://admin;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}

server {
    listen 80;
    server_name openrouter.ai;
    
    location / {
        proxy_pass http://portal;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

启用配置:
```bash
ln -s /etc/nginx/sites-available/openrouter /etc/nginx/sites-enabled/
nginx -t
systemctl reload nginx
```

## 集群部署 (Kubernetes)

适用于大规模生产环境。

### 1. Kubernetes配置
创建命名空间:
```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: openrouter
```

### 2. PostgreSQL StatefulSet
```yaml
# k8s/postgres.yaml
apiVersion: apps/v1
kind: StatefulSet
metadata:
  name: postgres
  namespace: openrouter
spec:
  serviceName: postgres
  replicas: 1
  selector:
    matchLabels:
      app: postgres
  template:
    metadata:
      labels:
        app: postgres
    spec:
      containers:
      - name: postgres
        image: postgres:16-alpine
        env:
        - name: POSTGRES_DB
          value: openrouter
        - name: POSTGRES_USER
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: username
        - name: POSTGRES_PASSWORD
          valueFrom:
            secretKeyRef:
              name: postgres-secret
              key: password
        ports:
        - containerPort: 5432
        volumeMounts:
        - name: postgres-data
          mountPath: /var/lib/postgresql/data
        resources:
          requests:
            memory: "1Gi"
            cpu: "500m"
          limits:
            memory: "2Gi"
            cpu: "1"
  volumeClaimTemplates:
  - metadata:
      name: postgres-data
    spec:
      accessModes: [ "ReadWriteOnce" ]
      resources:
        requests:
          storage: 50Gi
```

### 3. 后端应用部署
```yaml
# k8s/backend.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend
  namespace: openrouter
spec:
  replicas: 3
  selector:
    matchLabels:
      app: backend
  template:
    metadata:
      labels:
        app: backend
    spec:
      containers:
      - name: backend
        image: your-registry/openrouter-backend:latest
        envFrom:
        - configMapRef:
            name: backend-config
        - secretRef:
            name: backend-secrets
        ports:
        - containerPort: 8080
        resources:
          requests:
            memory: "512Mi"
            cpu: "250m"
          limits:
            memory: "1Gi"
            cpu: "500m"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
---
apiVersion: v1
kind: Service
metadata:
  name: backend
  namespace: openrouter
spec:
  selector:
    app: backend
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

### 4. 前端应用部署
类似配置前端应用部署。

### 5. Ingress配置
```yaml
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: openrouter-ingress
  namespace: openrouter
  annotations:
    nginx.ingress.kubernetes.io/rewrite-target: /
spec:
  rules:
  - host: api.openrouter.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: backend
            port:
              number: 80
  - host: admin.openrouter.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: admin
            port:
              number: 80
  - host: openrouter.ai
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: portal
            port:
              number: 80
```

## 数据库管理

### 1. 备份策略
```bash
# 每日备份
0 2 * * * docker exec postgres pg_dump -U openrouter openrouter > /backup/openrouter_$(date +\%Y\%m\%d).sql

# 备份到云存储
rclone copy /backup/ s3:openrouter-backups/
```

### 2. 监控查询
```sql
-- 查询慢SQL
SELECT * FROM pg_stat_statements ORDER BY total_time DESC LIMIT 10;

-- 查看连接数
SELECT count(*) FROM pg_stat_activity;

-- 表大小统计
SELECT schemaname, tablename, pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) 
FROM pg_tables WHERE schemaname NOT IN ('pg_catalog', 'information_schema') 
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

## 安全配置

### 1. SSL/TLS证书
使用Let's Encrypt自动证书:
```bash
# 安装certbot
apt install certbot python3-certbot-nginx -y

# 获取证书
certbot --nginx -d openrouter.ai -d api.openrouter.ai -d admin.openrouter.ai
```

### 2. 防火墙配置
```bash
# 只开放必要端口
ufw allow 22/tcp    # SSH
ufw allow 80/tcp    # HTTP
ufw allow 443/tcp   # HTTPS
ufw enable
```

### 3. 数据库安全
- 修改默认端口
- 启用SSL连接
- 限制访问IP
- 定期更换密码

## 监控与告警

### 1. 应用监控
```yaml
# Prometheus配置
scrape_configs:
  - job_name: 'openrouter-backend'
    static_configs:
      - targets: ['backend:8080']
    metrics_path: '/metrics'
```

### 2. 关键指标
- API响应时间 (P50, P95, P99)
- 错误率 (< 1%)
- 数据库连接池使用率
- Redis缓存命中率
- 系统资源使用率

### 3. 告警规则
```yaml
groups:
- name: openrouter-alerts
  rules:
  - alert: HighErrorRate
    expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.05
    for: 2m
    labels:
      severity: critical
    annotations:
      summary: "High error rate on {{ $labels.instance }}"
      description: "Error rate is {{ $value }}"
```

## 维护与更新

### 1. 应用更新
```bash
# 拉取最新代码
git pull origin main

# 重新构建镜像
docker-compose build

# 滚动更新
docker-compose up -d
```

### 2. 数据库迁移
```bash
# 测试环境验证
docker-compose exec backend make migrate-up

# 生产环境执行
docker-compose -f docker-compose.yml -f docker-compose.prod.yml exec backend make migrate-up
```

### 3. 数据清理
```sql
-- 清理过期日志
DELETE FROM request_logs WHERE created_at < NOW() - INTERVAL '30 days';

-- 清理无效会话
DELETE FROM user_sessions WHERE expires_at < NOW();
```

## 故障排除

### 1. 服务不可用检查清单
1. 检查容器状态: `docker-compose ps`
2. 查看日志: `docker-compose logs [service]`
3. 检查网络连接: `telnet localhost 5432`
4. 验证数据库连接: `docker-compose exec postgres pg_isready`
5. 检查磁盘空间: `df -h`

### 2. 性能问题排查
1. 查看慢查询日志
2. 检查Redis内存使用
3. 监控系统资源
4. 分析API响应时间

## 备份与恢复

### 1. 全量备份
```bash
# 数据库备份
docker-compose exec postgres pg_dumpall -U openrouter > backup.sql

# 配置文件备份
tar czf config-backup.tar.gz .env* docker-compose*.yml
```

### 2. 灾难恢复
```bash
# 恢复数据库
docker-compose exec postgres psql -U openrouter < backup.sql

# 恢复应用
docker-compose up -d
```

## 后续步骤

1. **性能优化**: 根据监控数据调整资源配置
2. **安全审计**: 定期进行安全漏洞扫描
3. **容量规划**: 根据业务增长预测资源需求
4. **文档更新**: 记录部署经验和最佳实践
