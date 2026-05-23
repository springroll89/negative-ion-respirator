# 负氧离子呼吸器 — 部署运维手册

## 系统要求

- Linux (Ubuntu 22.04+ / CentOS 8+)
- Docker 24+ & Docker Compose v2
- 最低配置: 2 CPU, 4GB RAM, 20GB 磁盘
- 推荐配置: 4 CPU, 8GB RAM, 50GB SSD

## 快速部署

### 1. 环境准备

```bash
# 安装 Docker
curl -fsSL https://get.docker.com | bash
sudo usermod -aG docker $USER

# 克隆项目
git clone <repo-url> /opt/ion-respirator
cd /opt/ion-respirator/backend
```

### 2. 配置环境变量

```bash
cp .env.example .env
# 编辑 .env 文件，修改所有密码和密钥
vim .env
```

关键配置项:
- `PG_PASSWORD`: 数据库密码 (必须修改)
- `JWT_SECRET`: JWT签名密钥，建议: `openssl rand -hex 32`
- `DATABASE_URL`: 包含正确的用户名/密码

### 3. 准备 TLS 证书

```bash
mkdir -p certs
# 生产环境: 使用 Let's Encrypt
certbot certonly --standalone -d your-domain.com
cp /etc/letsencrypt/live/your-domain.com/fullchain.pem certs/
cp /etc/letsencrypt/live/your-domain.com/privkey.pem certs/

# 开发环境: 生成自签名证书
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout certs/privkey.pem -out certs/fullchain.pem \
  -subj "/CN=localhost"
```

### 4. 启动服务

```bash
# 生产环境
make prod

# 查看状态
make ps
make logs
```

## 服务端口

| 服务 | 端口 | 说明 |
|------|------|------|
| Nginx (HTTPS) | 443 | 统一入口 |
| Nginx (HTTP) | 80 | 自动跳转HTTPS |
| Go Backend | 8080 | API服务 (仅内网) |
| PostgreSQL | 5432 | 数据库 (仅内网) |
| EMQX MQTT | 1883 | MQTT连接 |
| EMQX Dashboard | 18083 | 管理界面 |
| Redis | 6379 | 缓存 (仅内网) |

## 运维操作

### 查看日志

```bash
make logs                    # 所有服务
docker compose logs -f backend  # 仅后端
```

### 重启服务

```bash
docker compose restart backend
```

### 数据库备份

```bash
docker compose exec postgres pg_dump -U ion ion_respirator > backup_$(date +%Y%m%d).sql
```

### 数据库恢复

```bash
docker compose exec -T postgres psql -U ion ion_respirator < backup.sql
```

### 版本升级

```bash
git pull
make deploy    # 重新构建并启动
```

### 健康检查

```bash
curl -k https://localhost/health
curl -k https://localhost/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin123"}'
```

## 监控

- EMQX Dashboard: `https://<host>/mqtt-dashboard/` (默认 admin/public)
- 容器资源: `docker stats`
- 日志: `docker compose logs -f`

## 故障排查

### 后端无法连接数据库
```bash
docker compose ps postgres    # 确认状态为 healthy
docker compose logs postgres  # 查看数据库日志
```

### MQTT 连接失败
```bash
docker compose ps emqx
telnet localhost 1883         # 测试端口连通性
```

### 设备离线
1. 检查 EMQX Dashboard 是否有设备连接
2. 检查设备 Wi-Fi/4G 信号
3. 查看后端日志中的 MQTT 连接状态

## 安全清单

- [ ] `.env` 中所有默认密码已修改
- [ ] JWT_SECRET 已设置为随机字符串
- [ ] TLS 证书已配置
- [ ] 防火墙只开放 80/443 端口
- [ ] PostgreSQL 不对外暴露 (bind 127.0.0.1)
- [ ] 定期备份数据库
- [ ] 管理员密码已修改 (默认 admin/admin123)
