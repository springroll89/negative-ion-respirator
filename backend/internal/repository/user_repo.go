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
