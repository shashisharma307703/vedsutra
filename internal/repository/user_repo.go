package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shashisharma307703/vedantam/internal/domain/auth"
)

type userRepo struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &userRepo{db: db}
}

func (r *userRepo) Create(ctx context.Context, user *auth.User) (*auth.User, error) {
	query := `
		INSERT INTO users (tenant_id, email, username, password, is_active)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at, updated_at
	`
	err := r.db.QueryRowContext(ctx, query,
		user.TenantID, user.Email, user.Username, user.Password, user.IsActive,
	).Scan(&user.ID, &user.CreatedAt, &user.UpdatedAt)
	return user, err
}

func (r *userRepo) GetByID(ctx context.Context, id int64) (*auth.User, error) {
	query := `
		SELECT id, tenant_id, email, username, password, is_active, created_at, updated_at
		FROM users
		WHERE id = $1
	`
	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Username, &user.Password, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (*auth.User, error) {
	query := `
		SELECT id, tenant_id, email, username, password, is_active, created_at, updated_at
		FROM users
		WHERE email = $1
	`
	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Username, &user.Password, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepo) GetByUsername(ctx context.Context, username string) (*auth.User, error) {
	query := `
		SELECT id, tenant_id, email, username, password, is_active, created_at, updated_at
		FROM users
		WHERE username = $1
	`
	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, username).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Username, &user.Password, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepo) GetByEmailAndTenant(ctx context.Context, email string, tenantID int64) (*auth.User, error) {
	query := `
		SELECT id, tenant_id, email, username, password, is_active, created_at, updated_at
		FROM users
		WHERE email = $1 AND tenant_id = $2
	`
	user := &auth.User{}
	err := r.db.QueryRowContext(ctx, query, email, tenantID).Scan(
		&user.ID, &user.TenantID, &user.Email, &user.Username, &user.Password, &user.IsActive,
		&user.CreatedAt, &user.UpdatedAt,
	)
	return user, err
}

func (r *userRepo) Update(ctx context.Context, user *auth.User) error {
	query := `
		UPDATE users
		SET email = $1, username = $2, password = $3, is_active = $4, updated_at = $5
		WHERE id = $6
	`
	user.UpdatedAt = time.Now()
	_, err := r.db.ExecContext(ctx, query, user.Email, user.Username, user.Password, user.IsActive, user.UpdatedAt, user.ID)
	return err
}

func (r *userRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
