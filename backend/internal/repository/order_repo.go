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
