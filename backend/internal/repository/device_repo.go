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
