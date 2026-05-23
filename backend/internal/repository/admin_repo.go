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
