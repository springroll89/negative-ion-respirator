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
