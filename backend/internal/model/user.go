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
