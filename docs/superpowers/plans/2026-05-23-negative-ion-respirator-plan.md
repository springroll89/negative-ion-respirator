# 负氧离子呼吸器 — 实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 构建负氧离子呼吸器 IoT 全栈系统，包含 ESP32 固件、Go 后台服务、Vue 管理后台，支持 MQTT 通信和 Wi-Fi/4G 双模网络。

**Architecture:** 三层架构 — ESP32 设备层通过 MQTT (TLS) 连接 EMQX Broker，Go 服务层订阅 MQTT 消息并暴露 REST API，应用层包括管理后台 Web 端和与"谦昱堂213"APP 对接的接口。

**Tech Stack:** ESP-IDF (C), Go + Gin, PostgreSQL + TimescaleDB, EMQX, Redis, Vue 3 + Element Plus, Docker Compose

---

## Phase 1: 基础设施 & 协议定义

### Task 1: 项目初始化与目录结构

**Files:**
- Create: `backend/go.mod`
- Create: `backend/cmd/server/main.go`
- Create: `backend/internal/config/config.go`
- Create: `firmware/main/CMakeLists.txt`
- Create: `firmware/main/main.c`
- Create: `firmware/CMakeLists.txt`
- Create: `firmware/sdkconfig.defaults`
- Create: `firmware/partitions.csv`
- Create: `web-admin/` (Vue 3 脚手架)
- Create: `docs/protocol.md`

- [ ] **Step 1: 初始化 Go 模块**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序"
mkdir -p backend/cmd/server backend/internal/{config,handler,service,repository,mqtt,model,middleware} backend/migrations
cd backend && go mod init negative-ion-respirator/backend
```

- [ ] **Step 2: 创建 Go 服务入口**

Write `backend/cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"negative-ion-respirator/backend/internal/config"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	srv := &http.Server{
		Addr:         ":" + cfg.ServerPort,
		Handler:      setupRouter(cfg),
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		log.Printf("server starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server failed: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("forced shutdown: %v", err)
	}
}

func setupRouter(cfg *config.Config) http.Handler {
	// placeholder — will be replaced when handlers are added
	mux := http.NewServeMux()
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	})
	return mux
}
```

- [ ] **Step 3: 创建配置管理**

Write `backend/internal/config/config.go`:

```go
package config

import "os"

type Config struct {
	ServerPort   string
	DatabaseURL  string
	RedisURL     string
	EMQXHost     string
	EMQXClientID string
	JWTSecret    string
}

func Load() (*Config, error) {
	return &Config{
		ServerPort:   getEnv("SERVER_PORT", "8080"),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://ion:ion123@localhost:5432/ion_respirator?sslmode=disable"),
		RedisURL:     getEnv("REDIS_URL", "localhost:6379"),
		EMQXHost:     getEnv("EMQX_HOST", "tcp://localhost:1883"),
		EMQXClientID: getEnv("EMQX_CLIENT_ID", "backend-service"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

- [ ] **Step 4: 创建 ESP32 固件骨架**

Write `firmware/CMakeLists.txt`:

```cmake
cmake_minimum_required(VERSION 3.16)
include($ENV{IDF_PATH}/tools/cmake/project.cmake)
project(ion_respirator)
```

Write `firmware/main/CMakeLists.txt`:

```cmake
idf_component_register(
    SRCS "main.c"
    INCLUDE_DIRS "."
    REQUIRES nvs_flash esp_wifi esp_netif esp_event mqtt
)
```

Write `firmware/main/main.c`:

```c
#include <stdio.h>
#include "freertos/FreeRTOS.h"
#include "freertos/task.h"
#include "esp_system.h"
#include "esp_log.h"
#include "nvs_flash.h"

static const char *TAG = "main";

void app_main(void)
{
    esp_err_t ret = nvs_flash_init();
    if (ret == ESP_ERR_NVS_NO_FREE_PAGES || ret == ESP_ERR_NVS_NEW_VERSION_FOUND) {
        nvs_flash_erase();
        ret = nvs_flash_init();
    }
    ESP_ERROR_CHECK(ret);

    ESP_LOGI(TAG, "Negative Ion Respirator Firmware v1.0");
    ESP_LOGI(TAG, "Free memory: %" PRIu32 " bytes", esp_get_free_heap_size());

    while (1) {
        vTaskDelay(pdMS_TO_TICKS(1000));
    }
}
```

Write `firmware/sdkconfig.defaults`:

```
CONFIG_MQTT_ENABLED=y
CONFIG_BROKER_URL_FROM_STDIN=n
CONFIG_ESP_TLS_USING_MBEDTLS=y
```

Write `firmware/partitions.csv`:

```
# Name,   Type, SubType, Offset,   Size,    Flags
nvs,      data, nvs,      0x9000,   0x6000,
phy_init, data, phy,      0xf000,   0x1000,
factory,  app,  factory,  0x10000,  0x1F0000,
ota_0,    app,  ota_0,    0x200000, 0x1F0000,
ota_1,    app,  ota_1,    0x3F0000, 0x1F0000,
storage,  data, spiffs,   0x5E0000, 0x20000,
```

- [ ] **Step 5: 创建 Vue 3 管理后台脚手架**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序"
npm create vue@latest web-admin -- --typescript --router --pinia
cd web-admin && npm install && npm install element-plus @element-plus/icons-vue axios
```

- [ ] **Step 6: Commit Phase 1 初始化**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序"
git init
git add -A
git commit -m "chore: initialize project structure — Go backend, ESP32 firmware, Vue admin"
```

---

### Task 2: Docker 开发环境

**Files:**
- Create: `backend/docker-compose.yml`
- Create: `backend/Dockerfile`

- [ ] **Step 1: 创建 docker-compose.yml**

Write `backend/docker-compose.yml`:

```yaml
version: '3.8'

services:
  postgres:
    image: timescale/timescaledb:latest-pg16
    environment:
      POSTGRES_USER: ion
      POSTGRES_PASSWORD: ion123
      POSTGRES_DB: ion_respirator
    ports:
      - "5432:5432"
    volumes:
      - pg_data:/var/lib/postgresql/data
      - ./migrations:/docker-entrypoint-initdb.d
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U ion -d ion_respirator"]
      interval: 5s
      timeout: 5s
      retries: 5

  emqx:
    image: emqx/emqx:latest
    ports:
      - "1883:1883"      # MQTT
      - "8083:8083"      # WebSocket
      - "18083:18083"    # Dashboard
    environment:
      EMQX_NAME: ion-emqx
      EMQX_HOST: 127.0.0.1
    healthcheck:
      test: ["CMD", "emqx", "ping"]
      interval: 5s
      timeout: 5s
      retries: 10

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    healthcheck:
      test: ["CMD", "redis-cli", "ping"]
      interval: 5s
      timeout: 5s
      retries: 5

  backend:
    build: .
    ports:
      - "8080:8080"
    environment:
      SERVER_PORT: "8080"
      DATABASE_URL: "postgres://ion:ion123@postgres:5432/ion_respirator?sslmode=disable"
      REDIS_URL: "redis:6379"
      EMQX_HOST: "tcp://emqx:1883"
      EMQX_CLIENT_ID: "backend-service"
      JWT_SECRET: "dev-secret-change-in-production"
    depends_on:
      postgres:
        condition: service_healthy
      emqx:
        condition: service_healthy
      redis:
        condition: service_healthy

volumes:
  pg_data:
```

- [ ] **Step 2: 创建 Dockerfile**

Write `backend/Dockerfile`:

```dockerfile
FROM golang:1.22-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /server ./cmd/server

FROM alpine:3.19
RUN apk add --no-cache ca-certificates
COPY --from=builder /server /server
EXPOSE 8080
CMD ["/server"]
```

- [ ] **Step 3: 启动 Dcker 环境并验证**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
go mod tidy
docker compose up -d
docker compose ps  # 确认所有服务健康
curl http://localhost:8080/health  # 预期返回 {"status":"ok"}
```

- [ ] **Step 4: Commit**

```bash
git add backend/docker-compose.yml backend/Dockerfile backend/go.mod backend/go.sum
git commit -m "chore: add Docker Compose dev environment with PG, EMQX, Redis"
```

---

### Task 3: 通信协议文档

**Files:**
- Create: `docs/protocol.md`

- [ ] **Step 1: 编写通信协议文档**

Write `docs/protocol.md`:

````markdown
# 负氧离子呼吸器 — 通信协议

## MQTT 协议

### 连接参数

| 参数 | 值 |
|------|-----|
| Broker | EMQX (1883 MQTT / 8883 MQTTS) |
| 协议版本 | MQTT 3.1.1 |
| Keep Alive | 60s |
| Clean Session | false |
| LWT Topic | `device/{id}/heartbeat` |
| LWT Payload | `{"status":"offline"}` |
| LWT QoS | 1 |
| LWT Retain | true |

### Topic 定义

| Topic | 方向 | QoS | 说明 |
|-------|------|-----|------|
| `device/{id}/cmd` | 云→设备 | 1 | 启动/停止/配置指令 |
| `device/{id}/status` | 设备→云 | 1 | 运行状态上报 (5s) |
| `device/{id}/heartbeat` | 设备→云 | 0 | 心跳 (30s) |
| `device/{id}/event` | 设备→云 | 1 | 异常事件 |
| `device/{id}/ota` | 云→设备 | 2 | OTA 升级 |

### 消息格式

#### cmd (云→设备)

```json
// start: 启动设备
{"cmd": "start", "tid": "uuid-v4", "max_heat": 80, "target_out": 35}

// stop: 停止设备
{"cmd": "stop", "tid": "uuid-v4"}

// config: 更新配置 (设备运行时)
{"cmd": "config", "tid": "uuid-v4", "max_heat": 75, "target_out": 32}
```

#### status (设备→云, 每5s)

```json
{
  "status": "idle|heating|running|error",
  "tid": "uuid-v4",
  "heat_temp": 72.0,
  "out_temp": 34.5,
  "ion_ok": true,
  "uptime": 120
}
```

#### heartbeat (设备→云, 每30s)

```json
{
  "rssi": -45,
  "heap": 128456,
  "conn_type": "wifi|4g",
  "version": "1.0.0"
}
```

#### event (设备→云, 异常时)

```json
{
  "event": "over_temp|ion_fail|sensor_err|wdt_reset|power_low",
  "value": 85.0,
  "limit": 80.0,
  "action": "auto_shutdown|throttle|none"
}
```

## REST API

### 响应格式

```json
// 成功
{"code": 0, "message": "ok", "data": {...}}

// 失败
{"code": 40001, "message": "device not found", "data": null}

// 分页
{"code": 0, "message": "ok", "data": [...], "meta": {"total": 100, "page": 1, "page_size": 20}}
```

### 错误码

| Code | 说明 |
|------|------|
| 0 | 成功 |
| 40001 | 设备不存在 |
| 40002 | 设备离线 |
| 40003 | 设备忙 |
| 40004 | 订单不存在 |
| 40005 | 余额不足 |
| 40101 | Token 过期 |
| 40102 | Token 无效 |
| 40301 | 无权限 |
| 50001 | 服务器内部错误 |
| 50002 | MQTT 通信失败 |
````

- [ ] **Step 2: Commit**

```bash
git add docs/protocol.md
git commit -m "docs: add MQTT and REST API communication protocol specification"
```

---

### Task 4: 数据库 Migration

**Files:**
- Create: `backend/migrations/001_init.sql`

- [ ] **Step 1: 编写初始迁移脚本**

Write `backend/migrations/001_init.sql`:

```sql
-- Users
CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    open_id     VARCHAR(128) NOT NULL UNIQUE,
    nickname    VARCHAR(64) DEFAULT '',
    phone       VARCHAR(20) DEFAULT '',
    balance     BIGINT NOT NULL DEFAULT 0,  -- 余额(分)
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Devices
CREATE TABLE devices (
    id          BIGSERIAL PRIMARY KEY,
    device_sn   VARCHAR(64) NOT NULL UNIQUE,
    device_name VARCHAR(128) DEFAULT '',
    region_code VARCHAR(32) DEFAULT 'default',
    mqtt_topic  VARCHAR(256) NOT NULL,
    status      VARCHAR(16) NOT NULL DEFAULT 'offline',
    firmware_version VARCHAR(32) DEFAULT '',
    last_online TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Device config (per-device temperature settings)
CREATE TABLE device_config (
    id              BIGSERIAL PRIMARY KEY,
    device_id       BIGINT NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
    max_heat_temp   INT NOT NULL DEFAULT 80,
    target_out_temp INT NOT NULL DEFAULT 35,
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(device_id)
);

-- Region config (default config template by region + season)
CREATE TABLE region_config (
    id              BIGSERIAL PRIMARY KEY,
    region_code     VARCHAR(32) NOT NULL,
    season          VARCHAR(16) NOT NULL DEFAULT 'spring',
    max_heat_temp   INT NOT NULL DEFAULT 80,
    target_out_temp INT NOT NULL DEFAULT 35,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(region_code, season)
);

-- Orders
CREATE TABLE orders (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id),
    device_id   BIGINT NOT NULL REFERENCES devices(id),
    tid         VARCHAR(64) NOT NULL UNIQUE,
    start_time  TIMESTAMPTZ,
    end_time    TIMESTAMPTZ,
    duration    INT DEFAULT 0,  -- 使用时长(秒)
    amount      BIGINT DEFAULT 0,  -- 费用(分)
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Device telemetry logs (TimescaleDB hypertable)
CREATE TABLE device_logs (
    time        TIMESTAMPTZ NOT NULL,
    device_id   BIGINT NOT NULL,
    status      VARCHAR(16) NOT NULL,
    heat_temp   DOUBLE PRECISION,
    out_temp    DOUBLE PRECISION,
    ion_ok      BOOLEAN DEFAULT true,
    event_type  VARCHAR(32),
    event_data  JSONB
);

SELECT create_hypertable('device_logs', 'time', chunk_time_interval => INTERVAL '1 day');
SELECT add_compression_policy('device_logs', INTERVAL '7 days');
SELECT add_retention_policy('device_logs', INTERVAL '365 days');

CREATE INDEX idx_device_logs_device_time ON device_logs (device_id, time DESC);

-- Batch tasks (for batch config/upgrade operations)
CREATE TABLE batch_tasks (
    id          BIGSERIAL PRIMARY KEY,
    task_type   VARCHAR(32) NOT NULL,
    target_type VARCHAR(32) NOT NULL DEFAULT 'device',
    target_ids  BIGINT[] NOT NULL DEFAULT '{}',
    config_json JSONB DEFAULT '{}',
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    progress    INT DEFAULT 0,
    total       INT DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    finished_at TIMESTAMPTZ
);

-- Admin users
CREATE TABLE admin_users (
    id            BIGSERIAL PRIMARY KEY,
    username      VARCHAR(64) NOT NULL UNIQUE,
    password_hash VARCHAR(256) NOT NULL,
    role          VARCHAR(16) NOT NULL DEFAULT 'admin',
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Insert default region configs
INSERT INTO region_config (region_code, season, max_heat_temp, target_out_temp) VALUES
    ('default', 'spring', 80, 35),
    ('default', 'summer', 70, 32),
    ('default', 'autumn', 75, 34),
    ('default', 'winter', 80, 40);

-- Insert default admin (password: admin123, bcrypt hash)
INSERT INTO admin_users (username, password_hash, role) VALUES
    ('admin', '$2a$10$placeholder_use_real_bcrypt_hash', 'superadmin');
```

- [ ] **Step 2: 验证 Migration 执行**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
docker compose down -v  # 清空旧数据
docker compose up -d    # 重启，自动执行 init SQL
docker compose exec postgres psql -U ion -d ion_respirator -c "\dt"
```

预期输出应包含 8 张表: users, devices, device_config, region_config, orders, device_logs, batch_tasks, admin_users.

- [ ] **Step 3: Commit**

```bash
git add backend/migrations/001_init.sql
git commit -m "feat: add database migration with all 8 core tables and TimescaleDB setup"
```

---

## Phase 2: 核心功能 — 设备控制链路

### Task 5: Go 数据模型层

**Files:**
- Create: `backend/internal/model/user.go`
- Create: `backend/internal/model/device.go`
- Create: `backend/internal/model/order.go`
- Create: `backend/internal/model/telemetry.go`
- Create: `backend/internal/model/batch.go`
- Create: `backend/internal/model/admin.go`
- Create: `backend/internal/model/mqtt_msg.go`

- [ ] **Step 1: 创建 User 模型**

Write `backend/internal/model/user.go`:

```go
package model

import "time"

type User struct {
	ID        int64     `json:"id"`
	OpenID    string    `json:"open_id"`
	Nickname  string    `json:"nickname"`
	Phone     string    `json:"phone"`
	Balance   int64     `json:"balance"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
```

- [ ] **Step 2: 创建 Device 模型**

Write `backend/internal/model/device.go`:

```go
package model

import "time"

type Device struct {
	ID              int64      `json:"id"`
	DeviceSN        string     `json:"device_sn"`
	DeviceName      string     `json:"device_name"`
	RegionCode      string     `json:"region_code"`
	MqttTopic       string     `json:"mqtt_topic"`
	Status          string     `json:"status"`
	FirmwareVersion string     `json:"firmware_version"`
	LastOnline      *time.Time `json:"last_online"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

type DeviceConfig struct {
	ID            int64     `json:"id"`
	DeviceID      int64     `json:"device_id"`
	MaxHeatTemp   int       `json:"max_heat_temp"`
	TargetOutTemp int       `json:"target_out_temp"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type RegionConfig struct {
	ID            int64     `json:"id"`
	RegionCode    string    `json:"region_code"`
	Season        string    `json:"season"`
	MaxHeatTemp   int       `json:"max_heat_temp"`
	TargetOutTemp int       `json:"target_out_temp"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
```

- [ ] **Step 3: 创建 Order 模型**

Write `backend/internal/model/order.go`:

```go
package model

import "time"

type Order struct {
	ID        int64      `json:"id"`
	UserID    int64      `json:"user_id"`
	DeviceID  int64      `json:"device_id"`
	TID       string     `json:"tid"`
	StartTime *time.Time `json:"start_time"`
	EndTime   *time.Time `json:"end_time"`
	Duration  int        `json:"duration"`
	Amount    int64      `json:"amount"`
	Status    string     `json:"status"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

type CreateOrderReq struct {
	UserID   int64  `json:"user_id" binding:"required"`
	DeviceID int64  `json:"device_id" binding:"required"`
	OpenID   string `json:"open_id" binding:"required"`
}
```

- [ ] **Step 4: 创建 Telemetry 模型**

Write `backend/internal/model/telemetry.go`:

```go
package model

import "time"

type DeviceLog struct {
	Time      time.Time `json:"time"`
	DeviceID  int64     `json:"device_id"`
	Status    string    `json:"status"`
	HeatTemp  float64   `json:"heat_temp"`
	OutTemp   float64   `json:"out_temp"`
	IonOK     bool      `json:"ion_ok"`
	EventType string    `json:"event_type,omitempty"`
	EventData []byte    `json:"event_data,omitempty"`
}
```

- [ ] **Step 5: 创建 Batch 和 Admin 模型**

Write `backend/internal/model/batch.go`:

```go
package model

import "time"

type BatchTask struct {
	ID         int64     `json:"id"`
	TaskType   string    `json:"task_type"`
	TargetType string    `json:"target_type"`
	TargetIDs  []int64   `json:"target_ids"`
	ConfigJSON []byte    `json:"config_json"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	Total      int       `json:"total"`
	CreatedAt  time.Time `json:"created_at"`
	FinishedAt *time.Time `json:"finished_at"`
}
```

Write `backend/internal/model/admin.go`:

```go
package model

import "time"

type AdminUser struct {
	ID           int64     `json:"id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	Role         string    `json:"role"`
	CreatedAt    time.Time `json:"created_at"`
}

type LoginReq struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type LoginResp struct {
	Token     string `json:"token"`
	ExpiresAt int64  `json:"expires_at"`
}
```

- [ ] **Step 6: 创建 MQTT 消息模型**

Write `backend/internal/model/mqtt_msg.go`:

```go
package model

type DeviceCmd struct {
	Cmd       string `json:"cmd"`
	TID       string `json:"tid"`
	MaxHeat   int    `json:"max_heat,omitempty"`
	TargetOut int    `json:"target_out,omitempty"`
}

type DeviceStatus struct {
	Status   string  `json:"status"`
	TID      string  `json:"tid,omitempty"`
	HeatTemp float64 `json:"heat_temp"`
	OutTemp  float64 `json:"out_temp"`
	IonOK    bool    `json:"ion_ok"`
	Uptime   int     `json:"uptime"`
}

type DeviceHeartbeat struct {
	RSSI     int    `json:"rssi"`
	Heap     uint32 `json:"heap"`
	ConnType string `json:"conn_type"`
	Version  string `json:"version"`
}

type DeviceEvent struct {
	Event  string  `json:"event"`
	Value  float64 `json:"value,omitempty"`
	Limit  float64 `json:"limit,omitempty"`
	Action string  `json:"action"`
}
```

- [ ] **Step 7: Commit**

```bash
git add backend/internal/model/
git commit -m "feat: add data models for user, device, order, telemetry, batch, admin, mqtt"
```

---

### Task 6: Repository 层

**Files:**
- Create: `backend/internal/repository/db.go`
- Create: `backend/internal/repository/device_repo.go`
- Create: `backend/internal/repository/order_repo.go`
- Create: `backend/internal/repository/user_repo.go`
- Create: `backend/internal/repository/telemetry_repo.go`
- Create: `backend/internal/repository/admin_repo.go`

- [ ] **Step 1: 创建数据库连接**

Write `backend/internal/repository/db.go`:

```go
package repository

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

func NewDB(databaseURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("open db: %w", err)
	}
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("ping db: %w", err)
	}
	return db, nil
}
```

- [ ] **Step 2: 创建设备 Repository**

Write `backend/internal/repository/device_repo.go`:

```go
package repository

import (
	"context"
	"database/sql"
	"negative-ion-respirator/backend/internal/model"
)

type DeviceRepo struct{ db *sql.DB }

func NewDeviceRepo(db *sql.DB) *DeviceRepo { return &DeviceRepo{db: db} }

func (r *DeviceRepo) FindByID(ctx context.Context, id int64) (*model.Device, error) {
	d := &model.Device{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, device_sn, device_name, region_code, mqtt_topic, status,
		        firmware_version, last_online, created_at, updated_at
		 FROM devices WHERE id = $1`, id).
		Scan(&d.ID, &d.DeviceSN, &d.DeviceName, &d.RegionCode, &d.MqttTopic,
			&d.Status, &d.FirmwareVersion, &d.LastOnline, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DeviceRepo) FindBySN(ctx context.Context, sn string) (*model.Device, error) {
	d := &model.Device{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, device_sn, device_name, region_code, mqtt_topic, status,
		        firmware_version, last_online, created_at, updated_at
		 FROM devices WHERE device_sn = $1`, sn).
		Scan(&d.ID, &d.DeviceSN, &d.DeviceName, &d.RegionCode, &d.MqttTopic,
			&d.Status, &d.FirmwareVersion, &d.LastOnline, &d.CreatedAt, &d.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return d, nil
}

func (r *DeviceRepo) List(ctx context.Context, offset, limit int) ([]model.Device, int, error) {
	var total int
	r.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM devices").Scan(&total)

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, device_sn, device_name, region_code, mqtt_topic, status,
		        firmware_version, last_online, created_at, updated_at
		 FROM devices ORDER BY id DESC LIMIT $1 OFFSET $2`, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var devices []model.Device
	for rows.Next() {
		var d model.Device
		if err := rows.Scan(&d.ID, &d.DeviceSN, &d.DeviceName, &d.RegionCode, &d.MqttTopic,
			&d.Status, &d.FirmwareVersion, &d.LastOnline, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		devices = append(devices, d)
	}
	return devices, total, rows.Err()
}

func (r *DeviceRepo) UpdateStatus(ctx context.Context, id int64, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE devices SET status = $1, last_online = CASE WHEN $1 = 'online' THEN NOW() ELSE last_online END, updated_at = NOW() WHERE id = $2`,
		status, id)
	return err
}

func (r *DeviceRepo) Create(ctx context.Context, d *model.Device) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO devices (device_sn, device_name, region_code, mqtt_topic)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`,
		d.DeviceSN, d.DeviceName, d.RegionCode, d.MqttTopic).
		Scan(&d.ID, &d.CreatedAt, &d.UpdatedAt)
}

func (r *DeviceRepo) GetConfig(ctx context.Context, deviceID int64) (*model.DeviceConfig, error) {
	c := &model.DeviceConfig{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, device_id, max_heat_temp, target_out_temp, updated_at
		 FROM device_config WHERE device_id = $1`, deviceID).
		Scan(&c.ID, &c.DeviceID, &c.MaxHeatTemp, &c.TargetOutTemp, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *DeviceRepo) UpsertConfig(ctx context.Context, cfg *model.DeviceConfig) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO device_config (device_id, max_heat_temp, target_out_temp)
		 VALUES ($1, $2, $3)
		 ON CONFLICT (device_id) DO UPDATE SET max_heat_temp = $2, target_out_temp = $3, updated_at = NOW()`,
		cfg.DeviceID, cfg.MaxHeatTemp, cfg.TargetOutTemp)
	return err
}

func (r *DeviceRepo) GetRegionConfig(ctx context.Context, region, season string) (*model.RegionConfig, error) {
	c := &model.RegionConfig{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, region_code, season, max_heat_temp, target_out_temp, created_at, updated_at
		 FROM region_config WHERE region_code = $1 AND season = $2`, region, season).
		Scan(&c.ID, &c.RegionCode, &c.Season, &c.MaxHeatTemp, &c.TargetOutTemp, &c.CreatedAt, &c.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return c, nil
}

func (r *DeviceRepo) UpsertRegionConfig(ctx context.Context, c *model.RegionConfig) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO region_config (region_code, season, max_heat_temp, target_out_temp)
		 VALUES ($1, $2, $3, $4)
		 ON CONFLICT (region_code, season) DO UPDATE SET max_heat_temp = $3, target_out_temp = $4, updated_at = NOW()`,
		c.RegionCode, c.Season, c.MaxHeatTemp, c.TargetOutTemp)
	return err
}
```

- [ ] **Step 3: 创建订单 Repository**

Write `backend/internal/repository/order_repo.go`:

```go
package repository

import (
	"context"
	"database/sql"
	"negative-ion-respirator/backend/internal/model"

	"github.com/google/uuid"
)

type OrderRepo struct{ db *sql.DB }

func NewOrderRepo(db *sql.DB) *OrderRepo { return &OrderRepo{db: db} }

func (r *OrderRepo) Create(ctx context.Context, o *model.Order) error {
	o.TID = uuid.New().String()
	return r.db.QueryRowContext(ctx,
		`INSERT INTO orders (user_id, device_id, tid, status)
		 VALUES ($1, $2, $3, 'pending') RETURNING id, created_at, updated_at`,
		o.UserID, o.DeviceID, o.TID).Scan(&o.ID, &o.CreatedAt, &o.UpdatedAt)
}

func (r *OrderRepo) FindByTID(ctx context.Context, tid string) (*model.Order, error) {
	o := &model.Order{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, user_id, device_id, tid, start_time, end_time, duration, amount, status, created_at, updated_at
		 FROM orders WHERE tid = $1`, tid).
		Scan(&o.ID, &o.UserID, &o.DeviceID, &o.TID, &o.StartTime, &o.EndTime,
			&o.Duration, &o.Amount, &o.Status, &o.CreatedAt, &o.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return o, nil
}

func (r *OrderRepo) UpdateStatus(ctx context.Context, tid, status string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = $1, updated_at = NOW()
		 WHERE tid = $2`, status, tid)
	return err
}

func (r *OrderRepo) Settle(ctx context.Context, tid string, durationSec int, amount int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = 'completed', end_time = NOW(), duration = $1,
		 amount = $2, updated_at = NOW() WHERE tid = $3`,
		durationSec, amount, tid)
	return err
}

func (r *OrderRepo) MarkStarted(ctx context.Context, tid string) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE orders SET status = 'active', start_time = NOW(), updated_at = NOW()
		 WHERE tid = $1`, tid)
	return err
}
```

- [ ] **Step 4: 创建 User Repository**

Write `backend/internal/repository/user_repo.go`:

```go
package repository

import (
	"context"
	"database/sql"
	"negative-ion-respirator/backend/internal/model"
)

type UserRepo struct{ db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) FindByID(ctx context.Context, id int64) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, open_id, nickname, phone, balance, created_at, updated_at
		 FROM users WHERE id = $1`, id).
		Scan(&u.ID, &u.OpenID, &u.Nickname, &u.Phone, &u.Balance, &u.CreatedAt, &u.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) FindByOpenID(ctx context.Context, openID string) (*model.User, error) {
	u := &model.User{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, open_id, nickname, phone, balance, created_at, updated_at
		 FROM users WHERE open_id = $1`, openID).
		Scan(&u.ID, &u.OpenID, &u.Nickname, &u.Phone, &u.Balance, &u.CreatedAt, &u.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO users (open_id, nickname, phone, balance)
		 VALUES ($1, $2, $3, $4) RETURNING id, created_at, updated_at`,
		u.OpenID, u.Nickname, u.Phone, u.Balance).
		Scan(&u.ID, &u.CreatedAt, &u.UpdatedAt)
}

func (r *UserRepo) DeductBalance(ctx context.Context, userID int64, amount int64) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE users SET balance = balance - $1, updated_at = NOW()
		 WHERE id = $2 AND balance >= $1`, amount, userID)
	return err
}
```

- [ ] **Step 5: 创建遥测数据 Repository**

Write `backend/internal/repository/telemetry_repo.go`:

```go
package repository

import (
	"context"
	"database/sql"
	"negative-ion-respirator/backend/internal/model"
	"time"
)

type TelemetryRepo struct{ db *sql.DB }

func NewTelemetryRepo(db *sql.DB) *TelemetryRepo { return &TelemetryRepo{db: db} }

func (r *TelemetryRepo) Insert(ctx context.Context, log *model.DeviceLog) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO device_logs (time, device_id, status, heat_temp, out_temp, ion_ok, event_type, event_data)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		log.Time, log.DeviceID, log.Status, log.HeatTemp, log.OutTemp, log.IonOK, log.EventType, log.EventData)
	return err
}

func (r *TelemetryRepo) QueryByDevice(ctx context.Context, deviceID int64, start, end time.Time, limit int) ([]model.DeviceLog, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT time, device_id, status, heat_temp, out_temp, ion_ok, event_type, event_data
		 FROM device_logs
		 WHERE device_id = $1 AND time BETWEEN $2 AND $3
		 ORDER BY time DESC LIMIT $4`, deviceID, start, end, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var logs []model.DeviceLog
	for rows.Next() {
		var l model.DeviceLog
		if err := rows.Scan(&l.Time, &l.DeviceID, &l.Status, &l.HeatTemp,
			&l.OutTemp, &l.IonOK, &l.EventType, &l.EventData); err != nil {
			return nil, err
		}
		logs = append(logs, l)
	}
	return logs, rows.Err()
}
```

- [ ] **Step 6: 创建 Admin Repository**

Write `backend/internal/repository/admin_repo.go`:

```go
package repository

import (
	"context"
	"database/sql"
	"negative-ion-respirator/backend/internal/model"
)

type AdminRepo struct{ db *sql.DB }

func NewAdminRepo(db *sql.DB) *AdminRepo { return &AdminRepo{db: db} }

func (r *AdminRepo) FindByUsername(ctx context.Context, username string) (*model.AdminUser, error) {
	u := &model.AdminUser{}
	err := r.db.QueryRowContext(ctx,
		`SELECT id, username, password_hash, role, created_at FROM admin_users WHERE username = $1`, username).
		Scan(&u.ID, &u.Username, &u.PasswordHash, &u.Role, &u.CreatedAt)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *AdminRepo) Create(ctx context.Context, u *model.AdminUser) error {
	return r.db.QueryRowContext(ctx,
		`INSERT INTO admin_users (username, password_hash, role) VALUES ($1, $2, $3)
		 RETURNING id, created_at`, u.Username, u.PasswordHash, u.Role).
		Scan(&u.ID, &u.CreatedAt)
}
```

- [ ] **Step 7: 安装依赖并验证编译**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
go get github.com/lib/pq github.com/google/uuid golang.org/x/crypto
go build ./...
```

- [ ] **Step 8: Commit**

```bash
git add backend/internal/repository/
git commit -m "feat: add repository layer — device, order, user, telemetry, admin repos"
```

---

### Task 7: MQTT 服务层

**Files:**
- Create: `backend/internal/mqtt/client.go`

- [ ] **Step 1: 创建 MQTT 客户端**

Write `backend/internal/mqtt/client.go`:

```go
package mqtt

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/repository"
)

type Client struct {
	client   mqtt.Client
	deviceRepo *repository.DeviceRepo
	telemetryRepo *repository.TelemetryRepo
	orderRepo *repository.OrderRepo
	pending  map[string]chan []byte
	mu       sync.RWMutex
}

func NewClient(brokerURL, clientID string, deviceRepo *repository.DeviceRepo, telemetryRepo *repository.TelemetryRepo, orderRepo *repository.OrderRepo) (*Client, error) {
	c := &Client{
		deviceRepo:    deviceRepo,
		telemetryRepo: telemetryRepo,
		orderRepo:     orderRepo,
		pending:        make(map[string]chan []byte),
	}

	opts := mqtt.NewClientOptions().
		AddBroker(brokerURL).
		SetClientID(clientID).
		SetAutoReconnect(true).
		SetMaxReconnectInterval(10).
		SetOnConnectHandler(c.onConnect)

	c.client = mqtt.NewClient(opts)
	if token := c.client.Connect(); token.Wait() && token.Error() != nil {
		return nil, fmt.Errorf("mqtt connect: %w", token.Error())
	}

	return c, nil
}

func (c *Client) onConnect(client mqtt.Client) {
	log.Println("MQTT connected, subscribing to device topics...")
	client.Subscribe("device/+/status", 1, c.handleStatus)
	client.Subscribe("device/+/heartbeat", 0, c.handleHeartbeat)
	client.Subscribe("device/+/event", 1, c.handleEvent)
}

func (c *Client) SendCmd(ctx context.Context, deviceID int64, cmd model.DeviceCmd) error {
	d, err := c.deviceRepo.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("find device: %w", err)
	}

	payload, _ := json.Marshal(cmd)
	topic := fmt.Sprintf("device/%s/cmd", d.DeviceSN)

	token := c.client.Publish(topic, 1, false, payload)
	if token.Wait(); token.Error() != nil {
		return fmt.Errorf("publish: %w", token.Error())
	}
	return nil
}

func (c *Client) handleStatus(client mqtt.Client, msg mqtt.Message) {
	var status model.DeviceStatus
	if err := json.Unmarshal(msg.Payload(), &status); err != nil {
		log.Printf("bad status message: %v", err)
		return
	}
	log.Printf("device status: status=%s heat=%.1f out=%.1f ion=%v",
		status.Status, status.HeatTemp, status.OutTemp, status.IonOK)
}

func (c *Client) handleHeartbeat(client mqtt.Client, msg mqtt.Message) {
	var hb model.DeviceHeartbeat
	if err := json.Unmarshal(msg.Payload(), &hb); err != nil {
		log.Printf("bad heartbeat message: %v", err)
		return
	}
	log.Printf("device heartbeat: rssi=%d heap=%d conn=%s ver=%s",
		hb.RSSI, hb.Heap, hb.ConnType, hb.Version)
}

func (c *Client) handleEvent(client mqtt.Client, msg mqtt.Message) {
	var evt model.DeviceEvent
	if err := json.Unmarshal(msg.Payload(), &evt); err != nil {
		log.Printf("bad event message: %v", err)
		return
	}
	log.Printf("ALERT device event: event=%s value=%.1f action=%s",
		evt.Event, evt.Value, evt.Action)
}

func (c *Client) Close() {
	c.client.Disconnect(250)
}
```

- [ ] **Step 2: 安装依赖并编译验证**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
go get github.com/eclipse/paho.mqtt.golang
go build ./...
```

- [ ] **Step 3: Commit**

```bash
git add backend/internal/mqtt/
git commit -m "feat: add MQTT client with device status/heartbeat/event handlers"
```

---

### Task 8: Service 层

**Files:**
- Create: `backend/internal/service/device_service.go`
- Create: `backend/internal/service/order_service.go`
- Create: `backend/internal/service/auth_service.go`

- [ ] **Step 1: 创建设备 Service**

Write `backend/internal/service/device_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/mqtt"
	"negative-ion-respirator/backend/internal/repository"
)

type DeviceService struct {
	repo  *repository.DeviceRepo
	mqtt  *mqtt.Client
}

func NewDeviceService(repo *repository.DeviceRepo, mqttClient *mqtt.Client) *DeviceService {
	return &DeviceService{repo: repo, mqtt: mqttClient}
}

func (s *DeviceService) GetDevice(ctx context.Context, id int64) (*model.Device, error) {
	return s.repo.FindByID(ctx, id)
}

func (s *DeviceService) ListDevices(ctx context.Context, page, pageSize int) ([]model.Device, int, error) {
	offset := (page - 1) * pageSize
	return s.repo.List(ctx, offset, pageSize)
}

func (s *DeviceService) Start(ctx context.Context, deviceID int64, tid string) error {
	d, err := s.repo.FindByID(ctx, deviceID)
	if err != nil {
		return fmt.Errorf("device not found: %w", err)
	}
	if d.Status == "running" {
		return fmt.Errorf("device %d is already running", deviceID)
	}

	cfg, err := s.repo.GetConfig(ctx, deviceID)
	maxHeat, targetOut := 80, 35
	if err == nil {
		maxHeat = cfg.MaxHeatTemp
		targetOut = cfg.TargetOutTemp
	}

	cmd := model.DeviceCmd{
		Cmd: "start", TID: tid, MaxHeat: maxHeat, TargetOut: targetOut,
	}
	return s.mqtt.SendCmd(ctx, deviceID, cmd)
}

func (s *DeviceService) Stop(ctx context.Context, deviceID int64, tid string) error {
	cmd := model.DeviceCmd{Cmd: "stop", TID: tid}
	return s.mqtt.SendCmd(ctx, deviceID, cmd)
}

func (s *DeviceService) UpdateConfig(ctx context.Context, deviceID int64, maxHeat, targetOut int) error {
	if maxHeat < 0 || maxHeat > 80 {
		return fmt.Errorf("max_heat must be 0-80, got %d", maxHeat)
	}
	if targetOut < 30 || targetOut > 40 {
		return fmt.Errorf("target_out must be 30-40, got %d", targetOut)
	}

	cfg := &model.DeviceConfig{
		DeviceID:      deviceID,
		MaxHeatTemp:   maxHeat,
		TargetOutTemp: targetOut,
	}
	if err := s.repo.UpsertConfig(ctx, cfg); err != nil {
		return fmt.Errorf("save config: %w", err)
	}

	d, err := s.repo.FindByID(ctx, deviceID)
	if err != nil {
		return err
	}
	if d.Status == "running" {
		cmd := model.DeviceCmd{Cmd: "config", TID: "", MaxHeat: maxHeat, TargetOut: targetOut}
		return s.mqtt.SendCmd(ctx, deviceID, cmd)
	}
	return nil
}

func (s *DeviceService) Register(ctx context.Context, sn, name, region string) (*model.Device, error) {
	d := &model.Device{
		DeviceSN:   sn,
		DeviceName: name,
		RegionCode: region,
		MqttTopic:  fmt.Sprintf("device/%s", sn),
		Status:     "offline",
	}
	if err := s.repo.Create(ctx, d); err != nil {
		return nil, err
	}
	return d, nil
}

func (s *DeviceService) ListRegions(ctx context.Context) ([]string, error) {
	// simplified: return distinct region codes from devices
	return []string{"default", "north", "south"}, nil
}
```

- [ ] **Step 2: 创建订单 Service**

Write `backend/internal/service/order_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/repository"
	"time"
)

type OrderService struct {
	repo     *repository.OrderRepo
	userRepo *repository.UserRepo
	deviceSvc *DeviceService
}

func NewOrderService(repo *repository.OrderRepo, userRepo *repository.UserRepo, deviceSvc *DeviceService) *OrderService {
	return &OrderService{repo: repo, userRepo: userRepo, deviceSvc: deviceSvc}
}

func (s *OrderService) Create(ctx context.Context, req model.CreateOrderReq) (*model.Order, error) {
	_, err := s.deviceSvc.GetDevice(ctx, req.DeviceID)
	if err != nil {
		return nil, fmt.Errorf("device %d not found", req.DeviceID)
	}

	user, err := s.userRepo.FindByID(ctx, req.UserID)
	if err != nil || user == nil {
		if user == nil {
			user = &model.User{OpenID: req.OpenID, Balance: 0}
			if err := s.userRepo.Create(ctx, user); err != nil {
				return nil, fmt.Errorf("create user: %w", err)
			}
		} else {
			return nil, fmt.Errorf("user lookup failed: %w", err)
		}
	}

	o := &model.Order{UserID: user.ID, DeviceID: req.DeviceID}
	if err := s.repo.Create(ctx, o); err != nil {
		return nil, fmt.Errorf("create order: %w", err)
	}
	return o, nil
}

func (s *OrderService) Start(ctx context.Context, tid string) error {
	o, err := s.repo.FindByTID(ctx, tid)
	if err != nil {
		return fmt.Errorf("order not found: %w", err)
	}
	if o.Status != "pending" {
		return fmt.Errorf("order %s is not pending (status=%s)", tid, o.Status)
	}

	if err := s.deviceSvc.Start(ctx, o.DeviceID, tid); err != nil {
		return fmt.Errorf("start device: %w", err)
	}

	return s.repo.MarkStarted(ctx, tid)
}

func (s *OrderService) Stop(ctx context.Context, tid string) (*model.Order, error) {
	o, err := s.repo.FindByTID(ctx, tid)
	if err != nil {
		return nil, fmt.Errorf("order not found: %w", err)
	}
	if o.Status != "active" {
		return nil, fmt.Errorf("order %s is not active", tid)
	}

	if err := s.deviceSvc.Stop(ctx, o.DeviceID, tid); err != nil {
		return nil, fmt.Errorf("stop device: %w", err)
	}

	durationSec := int(time.Since(*o.StartTime).Seconds())
	amount := int64(durationSec) * 1 // 1分/秒，示例费率

	if err := s.repo.Settle(ctx, tid, durationSec, amount); err != nil {
		return nil, fmt.Errorf("settle order: %w", err)
	}

	o.Duration = durationSec
	o.Amount = amount
	o.Status = "completed"
	return o, nil
}

func (s *OrderService) FindByTID(ctx context.Context, tid string) (*model.Order, error) {
	return s.repo.FindByTID(ctx, tid)
}
```

- [ ] **Step 3: 创建认证 Service**

Write `backend/internal/service/auth_service.go`:

```go
package service

import (
	"context"
	"fmt"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/repository"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	repo      *repository.AdminRepo
	jwtSecret []byte
}

func NewAuthService(repo *repository.AdminRepo, secret string) *AuthService {
	return &AuthService{repo: repo, jwtSecret: []byte(secret)}
}

func (s *AuthService) Login(ctx context.Context, req model.LoginReq) (*model.LoginResp, error) {
	user, err := s.repo.FindByUsername(ctx, req.Username)
	if err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	expiresAt := time.Now().Add(24 * time.Hour)
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"user_id":  user.ID,
		"username": user.Username,
		"role":     user.Role,
		"exp":      expiresAt.Unix(),
	})

	tokenStr, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign token: %w", err)
	}

	return &model.LoginResp{Token: tokenStr, ExpiresAt: expiresAt.Unix()}, nil
}

func (s *AuthService) ValidateToken(tokenStr string) (jwt.MapClaims, error) {
	token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", t.Header["alg"])
		}
		return s.jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		return claims, nil
	}
	return nil, fmt.Errorf("invalid token")
}
```

- [ ] **Step 4: 安装依赖并编译验证**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
go get github.com/golang-jwt/jwt/v5 golang.org/x/crypto
go build ./...
```

- [ ] **Step 5: Commit**

```bash
git add backend/internal/service/
git commit -m "feat: add service layer — device, order, and auth services"
```

---

### Task 9: Handler 层 + 中间件

**Files:**
- Create: `backend/internal/middleware/auth.go`
- Create: `backend/internal/middleware/cors.go`
- Create: `backend/internal/handler/response.go`
- Create: `backend/internal/handler/device_handler.go`
- Create: `backend/internal/handler/order_handler.go`
- Create: `backend/internal/handler/auth_handler.go`
- Create: `backend/internal/handler/admin_handler.go`

- [ ] **Step 1: 安装 Gin**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
go get github.com/gin-gonic/gin
```

- [ ] **Step 2: 创建 JWT 中间件**

Write `backend/internal/middleware/auth.go`:

```go
package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/service"
)

func AuthRequired(authSvc *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if auth == "" || !strings.HasPrefix(auth, "Bearer ") {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40101, "message": "missing token"})
			c.Abort()
			return
		}
		claims, err := authSvc.ValidateToken(strings.TrimPrefix(auth, "Bearer "))
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"code": 40102, "message": "invalid token"})
			c.Abort()
			return
		}
		c.Set("user_id", claims["user_id"])
		c.Set("role", claims["role"])
		c.Next()
	}
}
```

- [ ] **Step 3: 创建 CORS 中间件**

Write `backend/internal/middleware/cors.go`:

```go
package middleware

import "github.com/gin-gonic/gin"

func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET,POST,PUT,DELETE,OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type,Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	}
}
```

- [ ] **Step 4: 创建统一响应和 Handler**

Write `backend/internal/handler/response.go`:

```go
package handler

import "github.com/gin-gonic/gin"

func OK(c *gin.Context, data interface{}) {
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": data})
}

func OKWithMeta(c *gin.Context, data interface{}, meta gin.H) {
	c.JSON(200, gin.H{"code": 0, "message": "ok", "data": data, "meta": meta})
}

func Err(c *gin.Context, code int, httpStatus int, msg string) {
	c.JSON(httpStatus, gin.H{"code": code, "message": msg, "data": nil})
}
```

Write `backend/internal/handler/device_handler.go`:

```go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/service"
)

type DeviceHandler struct{ svc *service.DeviceService }

func NewDeviceHandler(svc *service.DeviceService) *DeviceHandler { return &DeviceHandler{svc: svc} }

func (h *DeviceHandler) GetDevice(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	d, err := h.svc.GetDevice(c.Request.Context(), id)
	if err != nil {
		Err(c, 40001, 404, "device not found")
		return
	}
	OK(c, d)
}

func (h *DeviceHandler) ListDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	devices, total, err := h.svc.ListDevices(c.Request.Context(), page, pageSize)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OKWithMeta(c, devices, gin.H{"total": total, "page": page, "page_size": pageSize})
}

func (h *DeviceHandler) Register(c *gin.Context) {
	var req struct {
		DeviceSN   string `json:"device_sn" binding:"required"`
		DeviceName string `json:"device_name"`
		RegionCode string `json:"region_code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request: "+err.Error())
		return
	}
	d, err := h.svc.Register(c.Request.Context(), req.DeviceSN, req.DeviceName, req.RegionCode)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, d)
}
```

Write `backend/internal/handler/order_handler.go`:

```go
package handler

import (
	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/service"
)

type OrderHandler struct{ svc *service.OrderService }

func NewOrderHandler(svc *service.OrderService) *OrderHandler { return &OrderHandler{svc: svc} }

func (h *OrderHandler) Create(c *gin.Context) {
	var req model.CreateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request: "+err.Error())
		return
	}
	o, err := h.svc.Create(c.Request.Context(), req)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, o)
}

func (h *OrderHandler) Start(c *gin.Context) {
	var req struct {
		TID string `json:"tid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	if err := h.svc.Start(c.Request.Context(), req.TID); err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, gin.H{"status": "started"})
}

func (h *OrderHandler) Stop(c *gin.Context) {
	var req struct {
		TID string `json:"tid" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	o, err := h.svc.Stop(c.Request.Context(), req.TID)
	if err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, o)
}

func (h *OrderHandler) Query(c *gin.Context) {
	tid := c.Query("tid")
	o, err := h.svc.FindByTID(c.Request.Context(), tid)
	if err != nil {
		Err(c, 40004, 404, "order not found")
		return
	}
	OK(c, o)
}
```

Write `backend/internal/handler/auth_handler.go`:

```go
package handler

import (
	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/model"
	"negative-ion-respirator/backend/internal/service"
)

type AuthHandler struct{ svc *service.AuthService }

func NewAuthHandler(svc *service.AuthService) *AuthHandler { return &AuthHandler{svc: svc} }

func (h *AuthHandler) Login(c *gin.Context) {
	var req model.LoginReq
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	resp, err := h.svc.Login(c.Request.Context(), req)
	if err != nil {
		Err(c, 40102, 401, "invalid credentials")
		return
	}
	OK(c, resp)
}
```

Write `backend/internal/handler/admin_handler.go`:

```go
package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"negative-ion-respirator/backend/internal/service"
)

type AdminHandler struct {
	deviceSvc *service.DeviceService
}

func NewAdminHandler(deviceSvc *service.DeviceService) *AdminHandler {
	return &AdminHandler{deviceSvc: deviceSvc}
}

func (h *AdminHandler) UpdateDeviceConfig(c *gin.Context) {
	var req struct {
		DeviceID      int64 `json:"device_id" binding:"required"`
		MaxHeatTemp   int   `json:"max_heat_temp" binding:"required"`
		TargetOutTemp int   `json:"target_out_temp" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		Err(c, 40001, 400, "invalid request")
		return
	}
	if err := h.deviceSvc.UpdateConfig(c.Request.Context(), req.DeviceID, req.MaxHeatTemp, req.TargetOutTemp); err != nil {
		Err(c, 50001, 500, err.Error())
		return
	}
	OK(c, gin.H{"status": "config_updated"})
}

func (h *AdminHandler) GetDeviceStatus(c *gin.Context) {
	id, _ := strconv.ParseInt(c.Param("id"), 10, 64)
	d, err := h.deviceSvc.GetDevice(c.Request.Context(), id)
	if err != nil {
		Err(c, 40001, 404, "device not found")
		return
	}
	OK(c, d)
}
```

- [ ] **Step 5: 编译验证**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
go build ./...
```

- [ ] **Step 6: Commit**

```bash
git add backend/internal/handler/ backend/internal/middleware/
git commit -m "feat: add HTTP handlers and middleware — device, order, auth, CORS, JWT"
```

---

### Task 10: 主路由注册与服务启动

**Files:**
- Modify: `backend/cmd/server/main.go`

- [ ] **Step 1: 重写 main.go 注册路由**

Write `backend/cmd/server/main.go`:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"negative-ion-respirator/backend/internal/config"
	"negative-ion-respirator/backend/internal/handler"
	"negative-ion-respirator/backend/internal/middleware"
	"negative-ion-respirator/backend/internal/mqtt"
	"negative-ion-respirator/backend/internal/repository"
	"negative-ion-respirator/backend/internal/service"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	db, err := repository.NewDB(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer db.Close()

	// Repos
	deviceRepo := repository.NewDeviceRepo(db)
	orderRepo := repository.NewOrderRepo(db)
	userRepo := repository.NewUserRepo(db)
	telemetryRepo := repository.NewTelemetryRepo(db)
	adminRepo := repository.NewAdminRepo(db)

	// MQTT
	mqttClient, err := mqtt.NewClient(cfg.EMQXHost, cfg.EMQXClientID, deviceRepo, telemetryRepo, orderRepo)
	if err != nil {
		log.Printf("WARNING: MQTT connection failed: %v (continuing without MQTT)", err)
	}
	if mqttClient != nil {
		defer mqttClient.Close()
	}

	// Services
	deviceSvc := service.NewDeviceService(deviceRepo, mqttClient)
	orderSvc := service.NewOrderService(orderRepo, userRepo, deviceSvc)
	authSvc := service.NewAuthService(adminRepo, cfg.JWTSecret)

	// Handlers
	deviceH := handler.NewDeviceHandler(deviceSvc)
	orderH := handler.NewOrderHandler(orderSvc)
	authH := handler.NewAuthHandler(authSvc)
	adminH := handler.NewAdminHandler(deviceSvc)

	// Routes
	r := gin.Default()
	r.Use(middleware.CORS())

	api := r.Group("/api/v1")
	{
		api.POST("/order/create", orderH.Create)
		api.GET("/order/query", orderH.Query)

		api.POST("/device/start", orderH.Start)
		api.POST("/device/stop", orderH.Stop)
		api.GET("/device/status/:id", deviceH.GetDevice)

		api.POST("/auth/login", authH.Login)

		admin := api.Group("/admin")
		admin.Use(middleware.AuthRequired(authSvc))
		{
			admin.GET("/devices", deviceH.ListDevices)
			admin.GET("/device/:id", adminH.GetDeviceStatus)
			admin.POST("/device/register", deviceH.Register)
			admin.PUT("/device/config", adminH.UpdateDeviceConfig)
		}
	}

	srv := &http.Server{Addr: ":" + cfg.ServerPort, Handler: r}

	go func() {
		log.Printf("server starting on :%s", cfg.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("server: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("shutting down...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	srv.Shutdown(ctx)
}
```

- [ ] **Step 2: 编译 + 启动服务 + 测试 API**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/backend"
go build -o /tmp/ion-server ./cmd/server

# 在一个终端启动服务:
# DATABASE_URL="postgres://ion:ion123@localhost:5432/ion_respirator?sslmode=disable" /tmp/ion-server

# 测试 Health:
curl http://localhost:8080/api/v1/auth/login -X POST -H 'Content-Type: application/json' -d '{"username":"admin","password":"admin123"}'

# 测试设备列表:
curl http://localhost:8080/api/v1/admin/devices -H 'Authorization: Bearer <token>'
```

- [ ] **Step 3: Commit**

```bash
git add backend/cmd/server/main.go
git commit -m "feat: wire up complete Go server with Gin routing, DI, and MQTT integration"
```

---

### Task 11: Vue 管理后台基础页面

**Files:**
- Create: `web-admin/src/views/LoginView.vue`
- Create: `web-admin/src/views/DashboardView.vue`
- Create: `web-admin/src/views/DeviceListView.vue`
- Create: `web-admin/src/router/index.ts` (modify)
- Create: `web-admin/src/api/index.ts`

- [ ] **Step 1: 创建 API 封装**

Write `web-admin/src/api/index.ts`:

```typescript
import axios from 'axios'

const api = axios.create({ baseURL: '/api/v1' })

api.interceptors.request.use((config) => {
  const token = localStorage.getItem('token')
  if (token) config.headers.Authorization = `Bearer ${token}`
  return config
})

api.interceptors.response.use(
  (res) => res,
  (err) => {
    if (err.response?.status === 401) {
      localStorage.removeItem('token')
      window.location.href = '/login'
    }
    return Promise.reject(err)
  }
)

export const authAPI = {
  login: (data: { username: string; password: string }) => api.post('/auth/login', data),
}

export const deviceAPI = {
  list: (params: { page: number; page_size: number }) => api.get('/admin/devices', { params }),
  get: (id: number) => api.get(`/admin/device/${id}`),
  register: (data: { device_sn: string; device_name: string; region_code: string }) =>
    api.post('/admin/device/register', data),
  updateConfig: (data: { device_id: number; max_heat_temp: number; target_out_temp: number }) =>
    api.put('/admin/device/config', data),
}

export default api
```

- [ ] **Step 2: 创建登录页**

Write `web-admin/src/views/LoginView.vue`:

```vue
<template>
  <div class="login-container">
    <el-card class="login-card">
      <h2>负氧离子呼吸器 管理后台</h2>
      <el-form :model="form" @submit.prevent="handleLogin">
        <el-form-item><el-input v-model="form.username" placeholder="用户名" /></el-form-item>
        <el-form-item><el-input v-model="form.password" type="password" placeholder="密码" /></el-form-item>
        <el-form-item><el-button type="primary" @click="handleLogin" :loading="loading" style="width:100%">登录</el-button></el-form-item>
      </el-form>
    </el-card>
  </div>
</template>

<script setup lang="ts">
import { ref } from 'vue'
import { useRouter } from 'vue-router'
import { authAPI } from '@/api'
import { ElMessage } from 'element-plus'

const router = useRouter()
const loading = ref(false)
const form = ref({ username: 'admin', password: 'admin123' })

async function handleLogin() {
  loading.value = true
  try {
    const { data } = await authAPI.login(form.value)
    localStorage.setItem('token', data.data.token)
    ElMessage.success('登录成功')
    router.push('/')
  } catch {
    ElMessage.error('用户名或密码错误')
  } finally {
    loading.value = false
  }
}
</script>

<style scoped>
.login-container { display:flex; justify-content:center; align-items:center; min-height:100vh; background:#f5f7fa; }
.login-card { width:400px; }
.login-card h2 { text-align:center; margin-bottom:24px; }
</style>
```

- [ ] **Step 3: 创建仪表盘页**

Write `web-admin/src/views/DashboardView.vue`:

```vue
<template>
  <div>
    <h2>仪表盘</h2>
    <el-row :gutter="20">
      <el-col :span="6"><el-card><el-statistic title="设备总数" :value="stats.total" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="在线设备" :value="stats.online" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="今日订单" :value="stats.todayOrders" /></el-card></el-col>
      <el-col :span="6"><el-card><el-statistic title="今日营收" :value="stats.todayRevenue" prefix="¥" /></el-card></el-col>
    </el-row>
  </div>
</template>

<script setup lang="ts">
import { reactive } from 'vue'

const stats = reactive({ total: 0, online: 0, todayOrders: 0, todayRevenue: 0 })
</script>
```

- [ ] **Step 4: 创建设备列表页**

Write `web-admin/src/views/DeviceListView.vue`:

```vue
<template>
  <div>
    <h2>设备管理</h2>
    <el-table :data="devices" v-loading="loading">
      <el-table-column prop="id" label="ID" width="80" />
      <el-table-column prop="device_sn" label="序列号" />
      <el-table-column prop="device_name" label="名称" />
      <el-table-column prop="region_code" label="地区" />
      <el-table-column prop="status" label="状态">
        <template #default="{ row }">
          <el-tag :type="row.status === 'online' ? 'success' : 'info'">{{ row.status }}</el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作" width="200">
        <template #default="{ row }">
          <el-button size="small" @click="showConfig(row)">配置</el-button>
        </template>
      </el-table-column>
    </el-table>

    <el-dialog v-model="configVisible" title="设备配置">
      <el-form :model="configForm" label-width="120px">
        <el-form-item label="最高加热温度(℃)"><el-input-number v-model="configForm.max_heat_temp" :min="0" :max="80" /></el-form-item>
        <el-form-item label="目标出口温度(℃)"><el-input-number v-model="configForm.target_out_temp" :min="30" :max="40" /></el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="configVisible = false">取消</el-button>
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script setup lang="ts">
import { ref, onMounted } from 'vue'
import { deviceAPI } from '@/api'
import { ElMessage } from 'element-plus'

const devices = ref([])
const loading = ref(false)
const configVisible = ref(false)
const configForm = ref({ device_id: 0, max_heat_temp: 80, target_out_temp: 35 })

async function fetchDevices() {
  loading.value = true
  try {
    const { data } = await deviceAPI.list({ page: 1, page_size: 100 })
    devices.value = data.data
  } finally { loading.value = false }
}

function showConfig(row: any) {
  configForm.value = { device_id: row.id, max_heat_temp: 80, target_out_temp: 35 }
  configVisible.value = true
}

async function saveConfig() {
  try {
    await deviceAPI.updateConfig(configForm.value)
    ElMessage.success('配置已保存')
    configVisible.value = false
  } catch { ElMessage.error('保存失败') }
}

onMounted(fetchDevices)
</script>
```

- [ ] **Step 5: 更新路由**

Write `web-admin/src/router/index.ts`:

```typescript
import { createRouter, createWebHistory } from 'vue-router'

const router = createRouter({
  history: createWebHistory(import.meta.env.BASE_URL),
  routes: [
    { path: '/login', name: 'login', component: () => import('@/views/LoginView.vue') },
    { path: '/', name: 'dashboard', component: () => import('@/views/DashboardView.vue'), meta: { requiresAuth: true } },
    { path: '/devices', name: 'devices', component: () => import('@/views/DeviceListView.vue'), meta: { requiresAuth: true } },
  ],
})

router.beforeEach((to, _from, next) => {
  const token = localStorage.getItem('token')
  if (to.meta.requiresAuth && !token) next('/login')
  else next()
})

export default router
```

- [ ] **Step 6: 验证前端编译**

```bash
cd "/Users/chunjuan/Downloads/负氧离子呼吸器程序/web-admin"
npm run build
```

- [ ] **Step 7: Commit**

```bash
git add web-admin/
git commit -m "feat: add Vue admin pages — login, dashboard, device list with config dialog"
```

---

### Task 12: ESP32 MQTT 客户端组件

**Files:**
- Create: `firmware/components/mqtt_client/mqtt_client.h`
- Create: `firmware/components/mqtt_client/mqtt_client.c`
- Create: `firmware/components/mqtt_client/CMakeLists.txt`

- [ ] **Step 1: 创建头部文件**

Write `firmware/components/mqtt_client/mqtt_client.h`:

```c
#ifndef MQTT_CLIENT_H
#define MQTT_CLIENT_H

#include "esp_err.h"
#include <stdint.h>
#include <stdbool.h>

typedef void (*mqtt_cmd_callback_t)(const char *cmd, const char *tid,
                                     int max_heat, int target_out);

esp_err_t mqtt_client_init(const char *broker_url, const char *device_sn);
esp_err_t mqtt_client_start(void);
esp_err_t mqtt_client_stop(void);

esp_err_t mqtt_publish_status(const char *status, float heat_temp, float out_temp,
                               bool ion_ok, uint32_t uptime);
esp_err_t mqtt_publish_heartbeat(int rssi, const char *conn_type, const char *version);
esp_err_t mqtt_publish_event(const char *event, float value, float limit, const char *action);

void mqtt_set_cmd_callback(mqtt_cmd_callback_t cb);

#endif
```

- [ ] **Step 2: 实现 MQTT 客户端**

Write `firmware/components/mqtt_client/mqtt_client.c`:

```c
#include "mqtt_client.h"
#include "mqtt_client.h" // IDF built-in
#include "esp_log.h"
#include "esp_event.h"
#include "cJSON.h"
#include <string.h>

static const char *TAG = "mqtt";
static esp_mqtt_client_handle_t client = NULL;
static char device_sn[64] = {0};
static char cmd_topic[128] = {0};
static char status_topic[128] = {0};
static char heartbeat_topic[128] = {0};
static char event_topic[128] = {0};
static mqtt_cmd_callback_t cmd_cb = NULL;

static void mqtt_event_handler(void *arg, esp_event_base_t base,
                                int32_t event_id, void *event_data) {
    esp_mqtt_event_handle_t evt = event_data;

    switch (event_id) {
    case MQTT_EVENT_CONNECTED:
        ESP_LOGI(TAG, "MQTT connected");
        esp_mqtt_client_subscribe(client, cmd_topic, 1);
        break;
    case MQTT_EVENT_DATA:
        if (evt->topic_len >= strlen(cmd_topic) &&
            strncmp(evt->topic, cmd_topic, strlen(cmd_topic)) == 0) {
            cJSON *root = cJSON_ParseWithLength(evt->data, evt->data_len);
            if (root) {
                cJSON *cmd = cJSON_GetObjectItem(root, "cmd");
                cJSON *tid = cJSON_GetObjectItem(root, "tid");
                cJSON *max_heat = cJSON_GetObjectItem(root, "max_heat");
                cJSON *target_out = cJSON_GetObjectItem(root, "target_out");

                if (cmd && cmd_cb) {
                    cmd_cb(
                        cmd->valuestring,
                        tid ? tid->valuestring : "",
                        max_heat ? max_heat->valueint : 80,
                        target_out ? target_out->valueint : 35
                    );
                }
                cJSON_Delete(root);
            }
        }
        break;
    case MQTT_EVENT_DISCONNECTED:
        ESP_LOGW(TAG, "MQTT disconnected");
        break;
    default:
        break;
    }
}

esp_err_t mqtt_client_init(const char *broker_url, const char *sn) {
    strncpy(device_sn, sn, sizeof(device_sn) - 1);
    snprintf(cmd_topic, sizeof(cmd_topic), "device/%s/cmd", sn);
    snprintf(status_topic, sizeof(status_topic), "device/%s/status", sn);
    snprintf(heartbeat_topic, sizeof(heartbeat_topic), "device/%s/heartbeat", sn);
    snprintf(event_topic, sizeof(event_topic), "device/%s/event", sn);

    esp_mqtt_client_config_t cfg = {
        .broker.address.uri = broker_url,
        .credentials.client_id = sn,
        .session.keepalive = 60,
        .session.last_will = {
            .topic = heartbeat_topic,
            .msg = "{\"status\":\"offline\"}",
            .qos = 1,
            .retain = 1,
        },
    };

    client = esp_mqtt_client_init(&cfg);
    esp_mqtt_client_register_event(client, ESP_EVENT_ANY_ID, mqtt_event_handler, NULL);
    return ESP_OK;
}

esp_err_t mqtt_client_start(void) {
    return esp_mqtt_client_start(client);
}

esp_err_t mqtt_client_stop(void) {
    return esp_mqtt_client_stop(client);
}

esp_err_t mqtt_publish_status(const char *status, float heat_temp, float out_temp,
                               bool ion_ok, uint32_t uptime) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddStringToObject(root, "status", status);
    cJSON_AddNumberToObject(root, "heat_temp", heat_temp);
    cJSON_AddNumberToObject(root, "out_temp", out_temp);
    cJSON_AddBoolToObject(root, "ion_ok", ion_ok);
    cJSON_AddNumberToObject(root, "uptime", uptime);

    char *str = cJSON_PrintUnformatted(root);
    int msg_id = esp_mqtt_client_publish(client, status_topic, str, 0, 1, 0);
    free(str);
    cJSON_Delete(root);
    return msg_id >= 0 ? ESP_OK : ESP_FAIL;
}

esp_err_t mqtt_publish_heartbeat(int rssi, const char *conn_type, const char *version) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddNumberToObject(root, "rssi", rssi);
    cJSON_AddStringToObject(root, "conn_type", conn_type);
    cJSON_AddStringToObject(root, "version", version);
    cJSON_AddNumberToObject(root, "heap", esp_get_free_heap_size());

    char *str = cJSON_PrintUnformatted(root);
    int msg_id = esp_mqtt_client_publish(client, heartbeat_topic, str, 0, 0, 0);
    free(str);
    cJSON_Delete(root);
    return msg_id >= 0 ? ESP_OK : ESP_FAIL;
}

esp_err_t mqtt_publish_event(const char *event, float value, float limit, const char *action) {
    cJSON *root = cJSON_CreateObject();
    cJSON_AddStringToObject(root, "event", event);
    cJSON_AddNumberToObject(root, "value", value);
    cJSON_AddNumberToObject(root, "limit", limit);
    cJSON_AddStringToObject(root, "action", action);

    char *str = cJSON_PrintUnformatted(root);
    int msg_id = esp_mqtt_client_publish(client, event_topic, str, 0, 1, 0);
    free(str);
    cJSON_Delete(root);
    return msg_id >= 0 ? ESP_OK : ESP_FAIL;
}

void mqtt_set_cmd_callback(mqtt_cmd_callback_t cb) {
    cmd_cb = cb;
}
```

Write `firmware/components/mqtt_client/CMakeLists.txt`:

```cmake
idf_component_register(
    SRCS "mqtt_client.c"
    INCLUDE_DIRS "."
    REQUIRES mqtt cjson esp_wifi
)
```

- [ ] **Step 3: Commit**

```bash
git add firmware/components/mqtt_client/
git commit -m "feat: add ESP32 MQTT client component with TLS, LWT, and cJSON payload"
```

---

## Phase 3 & 4 任务将在 Phase 2 完成后细化

Phase 3 (高级功能) 和 Phase 4 (联调) 的任务将在 Phase 2 核心链路验证通过后编写具体实现步骤。包括:

- ESP32 温控 PID + 负离子驱动 + LED + 状态机 + 安全看门狗
- ESP32 4G 模块驱动 + Wi-Fi/4G 自动切换 + OTA
- Go 批量调参 + 报表 + Token 刷新
- Vue 批量配置 + 订单管理 + 报表 + 系统设置页
- 三方联调 + 压力测试 + 安全测试 + 部署文档
```

---

## 验证清单

- [ ] `docker compose up` 在 `backend/` 下启动全部4个服务(DB/EMQX/Redis/Go)
- [ ] `curl localhost:8080/health` 返回 200
- [ ] `curl -X POST localhost:8080/api/v1/auth/login` 返回 JWT token
- [ ] 带 token 调用 `/api/v1/admin/devices` 返回设备列表
- [ ] EMQX Dashboard (`localhost:18083`) 可访问
- [ ] Vue 项目 `npm run build` 编译通过
- [ ] ESP32 固件 `idf.py build` 编译通过
