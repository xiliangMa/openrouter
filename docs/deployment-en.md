# MassRouter SaaS Platform - Deployment Guide

This guide covers deploying the MassRouter SaaS platform using Docker Compose for production-like environments.

## Prerequisites

- Docker 20.10+ and Docker Compose v2+
- Git
- At least 4GB RAM available
- Linux/macOS/Windows with WSL2

## Quick Start (Development)

1. **Clone the repository**
   ```bash
   git clone <repository-url>
   cd openrouter-opencode
   ```

2. **Configure environment variables**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Start services with Docker Compose**
   ```bash
   docker-compose up -d
   ```

4. **Initialize the database**
   ```bash
   # Run migrations and seed data
   docker exec openrouter-backend go run cmd/seed/main.go
   ```

5. **Access the services**
   - API: http://localhost:8080/api/v1
   - Admin Panel: http://localhost:3000
   - User Portal: http://localhost:3001
   - Default admin credentials: admin@test.com / password123

## Production Deployment

### 1. Environment Configuration

Create a production `.env` file:

```bash
# Database
DB_PASSWORD=strong_random_password
DATABASE_URL=postgres://openrouter:${DB_PASSWORD}@postgres:5432/openrouter?sslmode=require

# Redis
REDIS_PASSWORD=strong_random_password
REDIS_URL=redis://:${REDIS_PASSWORD}@redis:6379/0

# JWT Authentication
JWT_SECRET=very_strong_random_secret_key_at_least_32_chars
JWT_EXPIRY=24h
JWT_REFRESH_EXPIRY=168h

# AI Provider API Keys (REQUIRED for production)
OPENAI_API_KEY=sk-your-real-openai-key
ANTHROPIC_API_KEY=sk-your-real-anthropic-key
GOOGLE_API_KEY=your-google-api-key
META_API_KEY=your-meta-api-key
COHERE_API_KEY=your-cohere-api-key

# Server Configuration
PORT=8080
GIN_MODE=release
CORS_ALLOWED_ORIGINS=https://your-domain.com
RATE_LIMIT=100

# Frontend Configuration
NEXT_PUBLIC_API_URL=https://api.your-domain.com/api/v1
NEXT_PUBLIC_APP_URL=https://admin.your-domain.com
```

### 2. Docker Compose Production Setup

Create `docker-compose.prod.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: postgres:16-alpine
    container_name: openrouter-postgres-prod
    environment:
      POSTGRES_DB: openrouter
      POSTGRES_USER: openrouter
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data
      - ./backups:/backups
    command: >
      postgres
      -c max_connections=200
      -c shared_buffers=256MB
      -c effective_cache_size=1GB
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U openrouter"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - openrouter-network

  redis:
    image: redis:7-alpine
    container_name: openrouter-redis-prod
    command: >
      redis-server
      --requirepass ${REDIS_PASSWORD}
      --maxmemory 256mb
      --maxmemory-policy allkeys-lru
    volumes:
      - redis_data:/data
    healthcheck:
      test: ["CMD", "redis-cli", "-a", "${REDIS_PASSWORD}", "ping"]
      interval: 30s
      timeout: 10s
      retries: 3
    restart: unless-stopped
    networks:
      - openrouter-network

  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile.prod
    container_name: openrouter-backend-prod
    depends_on:
      postgres:
        condition: service_healthy
      redis:
        condition: service_healthy
    environment:
      DATABASE_URL: ${DATABASE_URL}
      REDIS_URL: ${REDIS_URL}
      JWT_SECRET: ${JWT_SECRET}
      GIN_MODE: ${GIN_MODE}
      PORT: ${PORT}
      CORS_ALLOWED_ORIGINS: ${CORS_ALLOWED_ORIGINS}
      RATE_LIMIT: ${RATE_LIMIT}
      # AI Provider API Keys
      OPENAI_API_KEY: ${OPENAI_API_KEY}
      ANTHROPIC_API_KEY: ${ANTHROPIC_API_KEY}
      GOOGLE_API_KEY: ${GOOGLE_API_KEY}
      META_API_KEY: ${META_API_KEY}
      COHERE_API_KEY: ${COHERE_API_KEY}
    ports:
      - "${PORT}:${PORT}"
    restart: unless-stopped
    networks:
      - openrouter-network
    healthcheck:
      test: ["CMD", "wget", "--no-verbose", "--tries=1", "--spider", "http://localhost:${PORT}/api/v1/health"]
      interval: 30s
      timeout: 10s
      retries: 3

  admin:
    build:
      context: ./admin
      dockerfile: Dockerfile.prod
    container_name: openrouter-admin-prod
    depends_on:
      - backend
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
      NEXT_PUBLIC_APP_URL: ${NEXT_PUBLIC_APP_URL}
      NODE_ENV: production
    ports:
      - "3000:3000"
    restart: unless-stopped
    networks:
      - openrouter-network

  portal:
    build:
      context: ./portal
      dockerfile: Dockerfile.prod
    container_name: openrouter-portal-prod
    depends_on:
      - backend
    environment:
      NEXT_PUBLIC_API_URL: ${NEXT_PUBLIC_API_URL}
      NEXT_PUBLIC_APP_URL: ${NEXT_PUBLIC_APP_URL}
      NODE_ENV: production
    ports:
      - "3001:3001"
    restart: unless-stopped
    networks:
      - openrouter-network

  nginx:
    image: nginx:alpine
    container_name: openrouter-nginx-prod
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf:ro
      - ./ssl:/etc/nginx/ssl:ro
      - ./logs/nginx:/var/log/nginx
    depends_on:
      - backend
      - admin
      - portal
    restart: unless-stopped
    networks:
      - openrouter-network

volumes:
  postgres_data:
    driver: local
  redis_data:
    driver: local

networks:
  openrouter-network:
    driver: bridge
```

### 3. Create Production Dockerfiles

**backend/Dockerfile.prod:**
```dockerfile
# Build stage
FROM golang:1.22-alpine AS builder

WORKDIR /app

# Copy dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy binary from builder
COPY --from=builder /app/server .

# Copy configuration files
COPY --from=builder /app/.env.prod .env

EXPOSE 8080

CMD ["./server"]
```

**admin/Dockerfile.prod and portal/Dockerfile.prod:**
```dockerfile
# Build stage
FROM node:18-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci --only=production

COPY . .
RUN npm run build

# Runtime stage
FROM node:18-alpine

WORKDIR /app

COPY --from=builder /app/package*.json ./
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/.next ./.next
COPY --from=builder /app/public ./public

EXPOSE 3000

CMD ["npm", "start"]
```

### 4. Nginx Configuration

Create `nginx.conf`:

```nginx
events {
    worker_connections 1024;
}

http {
    upstream backend {
        server backend:8080;
    }

    upstream admin {
        server admin:3000;
    }

    upstream portal {
        server portal:3001;
    }

    server {
        listen 80;
        server_name api.your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name api.your-domain.com;

        ssl_certificate /etc/nginx/ssl/api.your-domain.com.crt;
        ssl_certificate_key /etc/nginx/ssl/api.your-domain.com.key;
        ssl_protocols TLSv1.2 TLSv1.3;

        location / {
            proxy_pass http://backend;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
            
            # Rate limiting headers
            proxy_set_header X-RateLimit-Limit $upstream_http_x_ratelimit_limit;
            proxy_set_header X-RateLimit-Remaining $upstream_http_x_ratelimit_remaining;
            proxy_set_header X-RateLimit-Reset $upstream_http_x_ratelimit_reset;
        }
    }

    server {
        listen 80;
        server_name admin.your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name admin.your-domain.com;

        ssl_certificate /etc/nginx/ssl/admin.your-domain.com.crt;
        ssl_certificate_key /etc/nginx/ssl/admin.your-domain.com.key;
        ssl_protocols TLSv1.2 TLSv1.3;

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
        server_name app.your-domain.com;
        return 301 https://$server_name$request_uri;
    }

    server {
        listen 443 ssl http2;
        server_name app.your-domain.com;

        ssl_certificate /etc/nginx/ssl/app.your-domain.com.crt;
        ssl_certificate_key /etc/nginx/ssl/app.your-domain.com.key;
        ssl_protocols TLSv1.2 TLSv1.3;

        location / {
            proxy_pass http://portal;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
            proxy_set_header X-Forwarded-Proto $scheme;
        }
    }
}
```

### 5. SSL Certificates

Generate or obtain SSL certificates:

```bash
mkdir -p ssl
# Use Let's Encrypt or your CA to generate certificates
# Place certificates in ssl/ directory
```

### 6. Deployment Commands

```bash
# Start production stack
docker-compose -f docker-compose.prod.yml up -d

# Check logs
docker-compose -f docker-compose.prod.yml logs -f

# Stop production stack
docker-compose -f docker-compose.prod.yml down

# Backup database
docker exec openrouter-postgres-prod pg_dump -U openrouter openrouter > backup_$(date +%Y%m%d_%H%M%S).sql

# Restore database
cat backup_file.sql | docker exec -i openrouter-postgres-prod psql -U openrouter openrouter
```

### 7. Monitoring and Maintenance

**Health Checks:**
- API: `GET https://api.your-domain.com/api/v1/health`
- Database: `docker exec openrouter-postgres-prod pg_isready -U openrouter`
- Redis: `docker exec openrouter-redis-prod redis-cli -a ${REDIS_PASSWORD} ping`

**Logs:**
```bash
# View all logs
docker-compose -f docker-compose.prod.yml logs

# View specific service logs
docker-compose -f docker-compose.prod.yml logs backend

# Real-time logs
docker-compose -f docker-compose.prod.yml logs -f backend
```

**Updates:**
```bash
# Pull latest code
git pull origin main

# Rebuild and restart
docker-compose -f docker-compose.prod.yml down
docker-compose -f docker-compose.prod.yml build --no-cache
docker-compose -f docker-compose.prod.yml up -d

# Run migrations (if any)
docker exec openrouter-backend-prod go run cmd/migrate/main.go
```

## Kubernetes Deployment (Advanced)

For Kubernetes deployment, refer to the `kubernetes/` directory for:
- Deployment manifests
- Service definitions
- Ingress configuration
- Persistent volume claims
- ConfigMaps and Secrets

## Troubleshooting

### Common Issues

1. **Database connection errors**
   - Check PostgreSQL logs: `docker logs openrouter-postgres-prod`
   - Verify credentials in `.env` file
   - Ensure database is running: `docker ps | grep postgres`

2. **Redis connection errors**
   - Check Redis logs: `docker logs openrouter-redis-prod`
   - Verify password in `.env` matches Redis configuration

3. **API returning 503 errors**
   - Check if AI provider API keys are configured
   - Verify provider configuration in database
   - Check backend logs for detailed errors

4. **Rate limiting issues**
   - Check Redis is running and accessible
   - Verify rate limit configuration in `.env`
   - Check X-RateLimit headers in response

5. **Frontend not connecting to API**
   - Verify `NEXT_PUBLIC_API_URL` environment variable
   - Check CORS configuration in backend
   - Inspect browser console for network errors

### Performance Tuning

1. **Database Optimization**
   - Increase shared_buffers for PostgreSQL
   - Add appropriate indexes for frequently queried tables
   - Consider connection pooling with PgBouncer

2. **Redis Optimization**
   - Adjust maxmemory policy based on usage patterns
   - Monitor memory usage and adjust maxmemory accordingly
   - Consider Redis cluster for high availability

3. **Backend Optimization**
   - Adjust GOMAXPROCS for CPU utilization
   - Implement response caching for static data
   - Consider horizontal scaling with load balancer

## Security Considerations

1. **Secrets Management**
   - Use Docker secrets or external secret management (Hashicorp Vault, AWS Secrets Manager)
   - Never commit secrets to version control
   - Rotate secrets regularly

2. **Network Security**
   - Use internal Docker networks for inter-service communication
   - Implement firewall rules to restrict access
   - Use VPN for administrative access

3. **API Security**
   - Implement request signing for sensitive endpoints
   - Add IP whitelisting for admin endpoints
   - Regular security audits and penetration testing

4. **Data Protection**
   - Encrypt sensitive data at rest
   - Implement data retention policies
   - Regular backups with encryption

## Support

For issues and questions:
1. Check the logs: `docker-compose logs [service-name]`
2. Review this documentation
3. Check GitHub issues
4. Contact the development team