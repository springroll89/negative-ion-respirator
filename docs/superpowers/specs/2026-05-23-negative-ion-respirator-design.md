# 负氧离子呼吸器 — 系统设计规格书

**日期**: 2026-05-23
**版本**: v1.0
**状态**: 已确认

---

## 1. 产品概述

共享型负氧离子吸入器 IoT 系统，支持扫码使用，接入"谦昱堂213"APP，实现用户身份识别、设备启动及后台远程控制。通过产生高浓度负氧离子并结合加热空气，为用户提供温润舒适的吸入体验。

### 技术选型

| 决策项 | 方案 |
|--------|------|
| MCU 平台 | ESP32 (ESP-IDF + FreeRTOS) |
| 通信方式 | Wi-Fi + 4G 双模 (SIMCom A7670E) |
| 通信协议 | MQTT (TLS 加密) |
| MQTT Broker | EMQX |
| 后台语言 | Go (Gin 框架) |
| 数据库(业务) | PostgreSQL |
| 数据库(时序) | TimescaleDB (PG 扩展) |
| 缓存 | Redis |
| 部署 | Docker Compose → 云服务器 |
| 管理后台前端 | Vue 3 + Element Plus |

---

## 2. 系统架构

### 2.1 三层架构

```
应用层:  "谦昱堂213" APP  |  管理后台 (Web)  |  运维监控 (Grafana)
                     ↕ HTTPS/REST
服务层:  Nginx → Go API → 设备管理服务
                 ↓ MQTT
              EMQX Broker
                 ↕ MQTT (TLS)
设备层:  呼吸器 #1 (ESP32 + Wi-Fi/4G) | 呼吸器 #2 | ...
```

### 2.2 数据流

```
用户扫码 → APP 验证 → Go API (鉴权/计费) → EMQX (MQTT下发) → ESP32 (启动/温控)
ESP32 状态上报 → MQTT → Go API → APP (状态展示) / 管理后台 (监控)
```

---

## 3. MQTT 协议设计

### 3.1 Topic 定义

| Topic | 方向 | QoS | 说明 |
|-------|------|-----|------|
| `device/{id}/cmd` | 云→设备 | 1 | 启动/停止/调温指令 |
| `device/{id}/status` | 设备→云 | 1 | 工作状态/温度/离子浓度 (5s) |
| `device/{id}/heartbeat` | 设备→云 | 0 | 心跳/在线状态 (30s) |
| `device/{id}/event` | 设备→云 | 1 | 异常告警/故障事件 |
| `device/{id}/ota` | 云→设备 | 2 | 固件 OTA 升级 |

### 3.2 消息格式 (JSON)

```json
// 云端→设备: cmd
{"cmd": "start", "tid": "uuid", "max_heat": 80, "target_out": 35}
{"cmd": "stop", "tid": "uuid"}
{"cmd": "config", "tid": "uuid", "max_heat": 75, "target_out": 32}

// 设备→云端: status (每5s)
{"status": "running", "heat_temp": 72, "out_temp": 34.5, "ion_ok": true, "uptime": 120}

// 设备→云端: heartbeat (每30s)
{"rssi": -45, "heap": 128456, "conn_type": "wifi"}

// 设备→云端: event (异常时)
{"event": "over_temp", "value": 85, "limit": 80, "action": "auto_shutdown"}
```

---

## 4. ESP32 固件设计

### 4.1 模块架构

```
main.c (入口 + 状态机)
components/
├── mqtt_client/     # MQTT 封装 (TLS + LWT + 断线重连)
├── heater_ctrl/     # PID 温控 (双传感器 PWM)
├── ion_gen/         # 负离子发生器驱动
├── led_indicator/   # LED 双指示灯
├── state_machine/   # 状态机 (IDLE→HEATING→RUNNING→DONE/ERR)
├── safety_watchdog/ # 硬件看门狗 + 过温保护
├── comm_4g/         # 4G 模块驱动 (AT 指令)
└── ota/             # OTA 固件升级
```

### 4.2 状态机

```
IDLE --(MQTT start)--> HEATING --(温度达标)--> RUNNING --(MQTT stop/超时)--> DONE --(自动)--> IDLE
任意状态 --(异常/过温)--> ERR --(恢复)--> IDLE
```

### 4.3 LED 指示灯

- **LED1 (电源)**: 通电常亮
- **LED2 (工作)**: 待机—灭, 预热—闪烁(1Hz), 运行—常亮, 故障—快闪(5Hz)

### 4.4 安全机制

- 硬件 WDT 看门狗
- 过温双重保护: 软件(80°C 限功率) + 硬件(85°C 断电)
- 负离子发生器异常检测与自动关闭
- NVS 断电状态保存
- MQTT LWT 遗嘱消息自动通知离线

---

## 5. Go 后台设计

### 5.1 分层架构

```
Transport:   Gin Router (REST API)
Handler:     device/user/admin/auth handler
Service:     Device/Order/MQTT/User/Monitor/Admin service
Repository:  DeviceRepo | OrderRepo | UserRepo | TelemetryRepo
Infra:       PostgreSQL | EMQX Client | Redis | JWT
```

### 5.2 核心 API

| Method | Path | 功能 | 调用方 |
|--------|------|------|--------|
| POST | `/api/v1/order/create` | 扫码创建订单 | APP |
| POST | `/api/v1/device/start` | 启动设备 | APP |
| POST | `/api/v1/device/stop` | 停止设备(结算) | APP |
| GET | `/api/v1/device/status/:id` | 查询设备状态 | APP/管理后台 |
| PUT | `/api/v1/admin/device/config` | 单设备温度配置 | 管理后台 |
| GET | `/api/v1/admin/devices` | 设备列表/批量管理 | 管理后台 |
| POST | `/api/v1/admin/batch/config` | 按地区/季节批量调参 | 管理后台 |
| GET | `/api/v1/admin/report/*` | 使用报表/营收 | 管理后台 |
| POST | `/api/v1/auth/login` | 管理后台登录 | 管理后台 |
| POST | `/api/v1/auth/refresh` | Token 刷新 | 管理后台 |

### 5.3 核心业务流程

**扫码使用**: 扫码 → 创建订单 → 验证用户 → 启动设备(MQTT) → 状态上报 → 停止设备 → 订单结算

**远程控温**: 管理员修改配置 → 校验权限/参数范围 → 写入DB → MQTT下发 → 设备更新PID → 确认回执

**批量调参**: 选择目标(地区/设备组) → 填写配置模板 → 后台遍历下发 → 记录批次任务 → 汇总结果

---

## 6. 数据库设计

### 6.1 核心表 (8张)

| 表 | 引擎 | 说明 | 预估量 |
|----|------|------|--------|
| `users` | PG | 用户账户、open_id、余额 | 万级 |
| `devices` | PG | 设备注册、序列号、在线状态 | 千级 |
| `device_config` | PG | 单设备温度参数 | 千级 |
| `region_config` | PG | 地区/季节默认模板 | 百级 |
| `orders` | PG | 使用订单、时长、费用 | 十万级 |
| `device_logs` | TimescaleDB | 遥测数据 (温度/离子/事件) | 千万级 |
| `batch_tasks` | PG | 批量操作任务记录 | 百级 |
| `admin_users` | PG | 后台管理员 | 十级 |

### 6.2 TimescaleDB 策略

```sql
SELECT create_hypertable('device_logs', 'timestamp', chunk_time_interval => INTERVAL '1 day');
SELECT add_compression_policy('device_logs', INTERVAL '7 days');
SELECT add_retention_policy('device_logs', INTERVAL '365 days');
```

---

## 7. 管理后台前端

### 7.1 页面结构

- **登录页**: 管理员账号密码登录 (JWT)
- **仪表盘**: 设备总数/在线数/今日订单/营收概览
- **设备管理**: 设备列表、状态监控、单设备配置
- **批量配置**: 按地区/季节批量下发温度参数
- **订单管理**: 订单列表、使用记录、退费处理
- **报表**: 使用趋势图、营收统计、设备利用率
- **系统设置**: 管理员账号管理、地区/季节配置模板

---

## 8. 项目工程结构

```
负氧离子呼吸器程序/
├── firmware/                    # ESP32 固件
│   ├── main/                    # 入口
│   └── components/              # 功能模块
├── backend/                     # Go 后台
│   ├── cmd/server/              # 入口
│   ├── internal/                # 内部包
│   ├── migrations/              # SQL 迁移
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── go.mod
├── web-admin/                   # 管理后台前端
├── docs/                        # 文档
│   ├── protocol.md              # 通信协议
│   └── superpowers/specs/       # 设计文档
└── 负氧离子呼吸器技术协议.doc
```

---

## 9. 开发阶段

| 阶段 | 内容 | 时间 |
|------|------|------|
| **Phase 1**: 基础设施 | 协议定义、Docker环境、Go/ESP32骨架、Migration | 3-5天 |
| **Phase 2**: 核心链路 | PID温控、负离子驱动、REST API、MQTT闭环、管理后台基础 | 5-7天 |
| **Phase 3**: 高级功能 | 4G驱动、OTA、批量调参、JWT认证、报表、安全加固 | 4-6天 |
| **Phase 4**: 联调上线 | 三方联调、压力测试、安全测试、部署文档 | 3-5天 |
| **总计** | | **15-23天** |

---

## 10. 与 APP 对接接口

APP 对接由客户方工程师负责，本系统需提供：

1. **扫码启动 API**: 接收设备 ID → 验证→ 创建订单 → 返回启动结果
2. **设备状态 API**: 实时查询设备运行状态 (温度/剩余时间等)
3. **停止/结算 API**: 结束使用 → 返回费用
4. **用户认证**: 接收 APP 端用户 token → 验证身份

通信协议需双方共同定义，本设计提供基础接口规范，后续配合 APP 端调整。
