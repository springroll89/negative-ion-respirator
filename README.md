# 负氧离子呼吸器 IoT 系统

共享型负氧离子吸入器全栈 IoT 系统。支持扫码使用，接入「谦昱堂213」APP，实现用户身份识别、设备启动及后台远程控制。

## 系统架构

```
应用层:  "谦昱堂213" APP  |  管理后台 (Vue 3)
                     ↕ HTTPS/REST
服务层:  Nginx → Go API (Gin) → EMQX MQTT Broker
                     ↕ MQTT (TLS)
设备层:  ESP32 (Wi-Fi/4G) → 负离子发生器 + 陶瓷加热管 + PID温控
```

## 技术栈

| 层 | 技术 |
|----|------|
| 设备固件 | ESP32, ESP-IDF, FreeRTOS, cJSON |
| 后台服务 | Go, Gin, paho.mqtt.golang, golang-jwt |
| 数据库 | PostgreSQL + TimescaleDB |
| 消息队列 | EMQX (MQTT 3.1.1) |
| 缓存 | Redis |
| 管理后台 | Vue 3, TypeScript, Element Plus, Pinia |
| 部署 | Docker Compose, Nginx, systemd |

## 项目结构

```
负氧离子呼吸器程序/
├── firmware/                     # ESP32 固件
│   ├── main/main.c              # 入口 + 主循环
│   └── components/
│       ├── mqtt_client/         # MQTT客户端 (TLS + LWT)
│       ├── heater_ctrl/         # PID温控 (PWM)
│       ├── ion_gen/             # 负离子发生器驱动
│       ├── led_indicator/       # LED指示灯
│       ├── state_machine/       # 设备状态机
│       ├── safety_watchdog/     # 安全看门狗
│       ├── comm_4g/             # 4G模块 (AT指令)
│       └── ota/                # OTA固件升级
├── backend/                     # Go 后台服务
│   ├── cmd/server/main.go      # 服务入口
│   ├── internal/
│   │   ├── config/             # 配置管理
│   │   ├── model/              # 数据模型
│   │   ├── repository/         # 数据访问层
│   │   ├── service/            # 业务逻辑层
│   │   ├── handler/            # HTTP处理层
│   │   ├── mqtt/               # MQTT客户端
│   │   └── middleware/         # JWT/CORS/限流中间件
│   ├── migrations/             # 数据库迁移脚本
│   ├── scripts/                # 运维脚本
│   ├── Dockerfile
│   ├── docker-compose.yml      # 开发环境
│   ├── docker-compose.prod.yml # 生产环境
│   ├── nginx.conf              # 反向代理配置
│   └── Makefile
├── web-admin/                   # Vue 3 管理后台
│   └── src/
│       ├── views/              # 7个页面
│       ├── api/                # API 客户端
│       └── router/             # 路由 + 鉴权守卫
└── docs/
    ├── protocol.md             # MQTT + REST API 协议
    ├── deployment.md           # 部署运维手册
    └── superpowers/            # 设计文档
```

## 快速开始

### 开发环境

```bash
# 1. 启动基础设施
cd backend
cp .env.example .env
docker compose up -d

# 2. 启动后端 (或使用 docker compose 中的 backend 服务)
go run ./cmd/server

# 3. 启动前端
cd ../web-admin
npm install
npm run dev
```

### 生产部署

```bash
cd backend
cp .env.example .env
vim .env          # 修改所有默认密码
mkdir -p certs    # 放置 TLS 证书
make prod         # 启动全部6个服务
```

## API 端点

### 公共接口

| Method | Path | 说明 |
|--------|------|------|
| POST | `/api/v1/order/create` | 扫码创建订单 |
| POST | `/api/v1/device/start` | 启动设备 |
| POST | `/api/v1/device/stop` | 停止设备 |
| GET | `/api/v1/device/status/:id` | 查询设备状态 |
| GET | `/api/v1/order/query` | 查询订单 |
| POST | `/api/v1/auth/login` | 管理员登录 |
| POST | `/api/v1/auth/refresh` | 刷新Token |

### 管理接口 (需JWT认证)

| Method | Path | 说明 |
|--------|------|------|
| GET | `/api/v1/admin/devices` | 设备列表 |
| GET | `/api/v1/admin/device/:id` | 设备详情 |
| POST | `/api/v1/admin/device/register` | 注册设备 |
| PUT | `/api/v1/admin/device/config` | 单设备温度配置 |
| POST | `/api/v1/admin/batch/config` | 批量调参 |
| GET | `/api/v1/admin/batch/task/:id` | 批量任务进度 |
| GET | `/api/v1/admin/dashboard` | 仪表盘数据 |
| GET | `/api/v1/admin/report` | 使用报表 |

## MQTT 协议

| Topic | 方向 | QoS | 频率 |
|-------|------|-----|------|
| `device/{id}/cmd` | 云→设备 | 1 | 按需 |
| `device/{id}/status` | 设备→云 | 1 | 5s |
| `device/{id}/heartbeat` | 设备→云 | 0 | 30s |
| `device/{id}/event` | 设备→云 | 1 | 异常时 |
| `device/{id}/ota` | 云→设备 | 2 | 升级时 |

详见 [docs/protocol.md](docs/protocol.md)

## 管理后台页面

- 仪表盘 — 设备/订单/营收概览
- 设备管理 — 设备列表 + 温度配置
- 批量配置 — 按地区/季节批量调参
- 订单管理 — 订单列表/状态
- 数据报表 — 使用趋势/营收/设备利用率
- 系统设置 — 地区季节默认温度模板

## 运维

```bash
make help       # 查看所有命令
make ps         # 服务状态
make logs       # 查看日志
make test       # 运行测试
make deploy     # 生产部署
```

## 设备状态机

```
IDLE ──(扫码启动)──→ HEATING ──(温度达标)──→ RUNNING ──(使用结束)──→ DONE → IDLE
  ↑                     ↓                      ↓
  └────(故障恢复)─── ERROR ←──(过温/故障)──────┘
```

## License

Proprietary. All rights reserved.
