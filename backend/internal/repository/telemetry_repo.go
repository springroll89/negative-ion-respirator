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
