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
