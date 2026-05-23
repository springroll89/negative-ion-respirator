-- Users
CREATE TABLE users (
    id          BIGSERIAL PRIMARY KEY,
    open_id     VARCHAR(128) NOT NULL UNIQUE,
    nickname    VARCHAR(64) DEFAULT '',
    phone       VARCHAR(20) DEFAULT '',
    balance     BIGINT NOT NULL DEFAULT 0,
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
    duration    INT DEFAULT 0,
    amount      BIGINT DEFAULT 0,
    status      VARCHAR(16) NOT NULL DEFAULT 'pending',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Enable TimescaleDB extension
CREATE EXTENSION IF NOT EXISTS timescaledb;

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

SELECT create_hypertable('device_logs'::regclass, 'time'::name);
SELECT set_chunk_time_interval('device_logs'::regclass, INTERVAL '1 day');
ALTER TABLE device_logs SET (timescaledb.compress, timescaledb.compress_segmentby = 'device_id', timescaledb.compress_orderby = 'time DESC');
SELECT add_compression_policy('device_logs'::regclass, INTERVAL '7 days');
SELECT add_retention_policy('device_logs'::regclass, INTERVAL '365 days');

CREATE INDEX idx_device_logs_device_time ON device_logs (device_id, time DESC);

-- Batch tasks
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

-- Seed data: default region configs
INSERT INTO region_config (region_code, season, max_heat_temp, target_out_temp) VALUES
    ('default', 'spring', 80, 35),
    ('default', 'summer', 70, 32),
    ('default', 'autumn', 75, 34),
    ('default', 'winter', 80, 40);

-- Seed data: default admin (password: admin123)
-- Generate proper bcrypt hash: $2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy
INSERT INTO admin_users (username, password_hash, role) VALUES
    ('admin', '$2a$10$N9qo8uLOickgx2ZMRZoMyeIjZAgcfl7p92ldGxad68LJZdL17lhWy', 'superadmin');
