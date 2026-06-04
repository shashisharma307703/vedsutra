package repository

import (
	"context"
	"database/sql"
	"time"

	"github.com/shashisharma307703/vedantam/internal/domain/auth"
)

// AuthSessionRepository handles refresh token sessions
type AuthSessionRepository interface {
	// Create saves a new refresh session
	Create(ctx context.Context, session *auth.RefreshSession) (*auth.RefreshSession, error)

	// GetByID retrieves a session by ID
	GetByID(ctx context.Context, id int64) (*auth.RefreshSession, error)

	// GetByUserID retrieves all active sessions for a user
	GetByUserID(ctx context.Context, userID int64) ([]*auth.RefreshSession, error)

	// GetByToken retrieves a session by refresh token hash
	GetByTokenHash(ctx context.Context, tokenHash string) (*auth.RefreshSession, error)

	// Update updates an existing session
	Update(ctx context.Context, session *auth.RefreshSession) error

	// Revoke marks a session as revoked
	Revoke(ctx context.Context, id int64) error

	// RevokeAllForUser revokes all sessions for a user
	RevokeAllForUser(ctx context.Context, userID int64) error

	// DeleteExpired deletes all expired sessions
	DeleteExpired(ctx context.Context) error
}

// AuthProviderRepository handles auth provider configurations
type AuthProviderRepository interface {
	// Create saves a new auth provider
	Create(ctx context.Context, provider *auth.AuthProvider) (*auth.AuthProvider, error)

	// GetByID retrieves a provider by ID
	GetByID(ctx context.Context, id int64) (*auth.AuthProvider, error)

	// GetByTenantAndType retrieves provider for a tenant by type
	GetByTenantAndType(ctx context.Context, tenantID int64, providerType auth.ProviderType) (*auth.AuthProvider, error)

	// GetActiveByTenant retrieves all active providers for a tenant
	GetActiveByTenant(ctx context.Context, tenantID int64) ([]*auth.AuthProvider, error)

	// Update updates an existing provider
	Update(ctx context.Context, provider *auth.AuthProvider) error

	// Delete deletes a provider
	Delete(ctx context.Context, id int64) error
}

// OIDCDiscoveryCacheRepository handles OIDC discovery document caching
type OIDCDiscoveryCacheRepository interface {
	// Create saves a new discovery document cache
	Create(ctx context.Context, cache *auth.OIDCDiscoveryCache) (*auth.OIDCDiscoveryCache, error)

	// GetByIssuer retrieves cached discovery document by issuer
	GetByIssuer(ctx context.Context, issuer string) (*auth.OIDCDiscoveryCache, error)

	// Update updates an existing cache
	Update(ctx context.Context, cache *auth.OIDCDiscoveryCache) error

	// Delete deletes a cache entry
	Delete(ctx context.Context, id int64) error

	// DeleteExpired deletes all expired cache entries
	DeleteExpired(ctx context.Context) error
}

// UserRepository handles user data
type UserRepository interface {
	// Create saves a new user
	Create(ctx context.Context, user *auth.User) (*auth.User, error)

	// GetByID retrieves a user by ID
	GetByID(ctx context.Context, id int64) (*auth.User, error)

	// GetByEmail retrieves a user by email
	GetByEmail(ctx context.Context, email string) (*auth.User, error)

	// GetByUsername retrieves a user by username
	GetByUsername(ctx context.Context, username string) (*auth.User, error)

	// GetByEmailAndTenant retrieves a user by email and tenant
	GetByEmailAndTenant(ctx context.Context, email string, tenantID int64) (*auth.User, error)

	// Update updates an existing user
	Update(ctx context.Context, user *auth.User) error

	// Delete deletes a user
	Delete(ctx context.Context, id int64) error
}

// Base implementations
type authSessionRepo struct {
	db *sql.DB
}

func NewAuthSessionRepository(db *sql.DB) AuthSessionRepository {
	return &authSessionRepo{db: db}
}

func (r *authSessionRepo) Create(ctx context.Context, session *auth.RefreshSession) (*auth.RefreshSession, error) {
	query := `
		INSERT INTO auth_sessions (user_id, tenant_id, refresh_token_hash, expires_at, ip_address, user_agent)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, created_at
	`
	err := r.db.QueryRowContext(ctx, query,
		session.UserID, session.TenantID, session.RefreshTokenHash, session.ExpiresAt,
		session.IPAddress, session.UserAgent,
	).Scan(&session.ID, &session.CreatedAt)
	return session, err
}

func (r *authSessionRepo) GetByID(ctx context.Context, id int64) (*auth.RefreshSession, error) {
	query := `
		SELECT id, user_id, tenant_id, refresh_token_hash, expires_at, created_at, last_used_at, ip_address, user_agent, is_revoked
		FROM auth_sessions
		WHERE id = $1
	`
	session := &auth.RefreshSession{}
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&session.ID, &session.UserID, &session.TenantID, &session.RefreshTokenHash, &session.ExpiresAt,
		&session.CreatedAt, &session.LastUsedAt, &session.IPAddress, &session.UserAgent, &session.IsRevoked,
	)
	return session, err
}

func (r *authSessionRepo) GetByUserID(ctx context.Context, userID int64) ([]*auth.RefreshSession, error) {
	query := `
		SELECT id, user_id, tenant_id, refresh_token_hash, expires_at, created_at, last_used_at, ip_address, user_agent, is_revoked
		FROM auth_sessions
		WHERE user_id = $1 AND is_revoked = false AND expires_at > NOW()
		ORDER BY created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []*auth.RefreshSession
	for rows.Next() {
		session := &auth.RefreshSession{}
		err := rows.Scan(
			&session.ID, &session.UserID, &session.TenantID, &session.RefreshTokenHash, &session.ExpiresAt,
			&session.CreatedAt, &session.LastUsedAt, &session.IPAddress, &session.UserAgent, &session.IsRevoked,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}
	return sessions, rows.Err()
}

func (r *authSessionRepo) GetByTokenHash(ctx context.Context, tokenHash string) (*auth.RefreshSession, error) {
	query := `
		SELECT id, user_id, tenant_id, refresh_token_hash, expires_at, created_at, last_used_at, ip_address, user_agent, is_revoked
		FROM auth_sessions
		WHERE refresh_token_hash = $1 AND is_revoked = false AND expires_at > NOW()
	`
	session := &auth.RefreshSession{}
	err := r.db.QueryRowContext(ctx, query, tokenHash).Scan(
		&session.ID, &session.UserID, &session.TenantID, &session.RefreshTokenHash, &session.ExpiresAt,
		&session.CreatedAt, &session.LastUsedAt, &session.IPAddress, &session.UserAgent, &session.IsRevoked,
	)
	return session, err
}

func (r *authSessionRepo) Update(ctx context.Context, session *auth.RefreshSession) error {
	query := `
		UPDATE auth_sessions
		SET last_used_at = $1, is_revoked = $2
		WHERE id = $3
	`
	_, err := r.db.ExecContext(ctx, query, time.Now(), session.IsRevoked, session.ID)
	return err
}

func (r *authSessionRepo) Revoke(ctx context.Context, id int64) error {
	query := `UPDATE auth_sessions SET is_revoked = true WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}

func (r *authSessionRepo) RevokeAllForUser(ctx context.Context, userID int64) error {
	query := `UPDATE auth_sessions SET is_revoked = true WHERE user_id = $1`
	_, err := r.db.ExecContext(ctx, query, userID)
	return err
}

func (r *authSessionRepo) DeleteExpired(ctx context.Context) error {
	query := `DELETE FROM auth_sessions WHERE expires_at < NOW()`
	_, err := r.db.ExecContext(ctx, query)
	return err
}
